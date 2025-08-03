package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/vagonaizer/authenitfication-service/internal/domain/entities"
	"github.com/vagonaizer/authenitfication-service/internal/domain/repositories"
	"github.com/vagonaizer/authenitfication-service/internal/dto/request"
	"github.com/vagonaizer/authenitfication-service/internal/dto/response"
	"github.com/vagonaizer/authenitfication-service/internal/infrastructure/messaging/kafka"
	"github.com/vagonaizer/authenitfication-service/pkg/auth"
	"github.com/vagonaizer/authenitfication-service/pkg/errors"
	"github.com/vagonaizer/authenitfication-service/pkg/logger"
	"github.com/vagonaizer/authenitfication-service/pkg/utils"
)

type AuthService struct {
	userRepo       repositories.UserRepository
	sessionRepo    repositories.SessionRepository
	roleRepo       repositories.RoleRepository
	passwordHasher *auth.PasswordHasher
	jwtManager     *auth.JWTManager
	producer       *kafka.Producer
	logger         *logger.Logger
	accessExpiry   time.Duration
	refreshExpiry  time.Duration
}

func NewAuthService(
	userRepo repositories.UserRepository,
	sessionRepo repositories.SessionRepository,
	roleRepo repositories.RoleRepository,
	passwordHasher *auth.PasswordHasher,
	jwtManager *auth.JWTManager,
	producer *kafka.Producer,
	logger *logger.Logger,
	accessExpiry time.Duration,
	refreshExpiry time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		roleRepo:       roleRepo,
		passwordHasher: passwordHasher,
		jwtManager:     jwtManager,
		producer:       producer,
		logger:         logger,
		accessExpiry:   accessExpiry,
		refreshExpiry:  refreshExpiry,
	}
}

func (s *AuthService) Register(ctx context.Context, req *request.RegisterRequest, ipAddress, userAgent string) (*response.AuthResponse, error) {
	if !utils.IsValidEmail(req.Email) {
		return nil, errors.Validation("invalid email format")
	}

	if !utils.IsValidUsername(req.Username) {
		return nil, errors.Validation("invalid username format")
	}

	if !utils.IsValidPassword(req.Password) {
		return nil, errors.WeakPassword()
	}

	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.EmailExists()
	}

	exists, err = s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.UsernameExists()
	}

	passwordHash, err := s.passwordHasher.HashPassword(req.Password)
	if err != nil {
		s.logger.WithError(err).Error("failed to hash password")
		return nil, errors.Internal("failed to process password")
	}

	user := &entities.User{
		ID:           uuid.New(),
		Email:        utils.NormalizeEmail(req.Email),
		Username:     utils.NormalizeUsername(req.Username),
		PasswordHash: passwordHash,
		FirstName:    &req.FirstName,
		LastName:     &req.LastName,
		IsActive:     true,
		IsVerified:   false,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Назначаем роль по умолчанию (игнорируем ошибки)
	defaultRole, err := s.roleRepo.GetByName(ctx, "user")
	if err != nil {
		s.logger.WithError(err).Warn("failed to get default role")
	} else {
		if err := s.roleRepo.AssignRoleToUser(ctx, user.ID, defaultRole.ID); err != nil {
			s.logger.WithError(err).Warn("failed to assign default role")
		}
	}

	// Получаем роли пользователя (с обработкой ошибок)
	userRoles, err := s.roleRepo.GetUserRoles(ctx, user.ID)
	if err != nil {
		s.logger.WithError(err).Warn("failed to get user roles, using empty roles")
		userRoles = []*entities.Role{}
	}

	roleNames := make([]string, len(userRoles))
	for i, role := range userRoles {
		roleNames[i] = role.Name
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Username, roleNames, s.accessExpiry)
	if err != nil {
		s.logger.WithError(err).Error("failed to generate access token")
		return nil, errors.Internal("failed to generate tokens")
	}

	// Генерируем короткий refresh token
	refreshToken, err := utils.GenerateSecureToken()
	if err != nil {
		s.logger.WithError(err).Error("failed to generate refresh token")
		return nil, errors.Internal("failed to generate tokens")
	}

	session := &entities.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		IsActive:     true,
		ExpiresAt:    time.Now().Add(s.refreshExpiry),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	// Публикуем событие (игнорируем ошибки)
	event := kafka.UserRegisteredEvent{
		BaseEvent: kafka.NewBaseEvent(kafka.TopicUserRegistered),
		UserID:    user.ID,
		Email:     user.Email,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	if err := s.producer.PublishMessage(ctx, kafka.TopicUserRegistered, user.ID.String(), event); err != nil {
		s.logger.WithError(err).Warn("failed to publish user registered event")
	}

	return &response.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessExpiry.Seconds()),
		User: &response.UserResponse{
			ID:          user.ID,
			Email:       user.Email,
			Username:    user.Username,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			IsActive:    user.IsActive,
			IsVerified:  user.IsVerified,
			LastLoginAt: user.LastLoginAt,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		},
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *request.LoginRequest, ipAddress, userAgent string) (*response.AuthResponse, error) {
	s.logger.WithFields(logger.Fields{
		"email": req.Email,
		"ip":    ipAddress,
	}).Info("login attempt started")

	// Шаг 1: Получение пользователя
	user, err := s.userRepo.GetByEmail(ctx, utils.NormalizeEmail(req.Email))
	if err != nil {
		s.logger.WithError(err).WithField("email", req.Email).Error("failed to get user by email")
		return nil, errors.InvalidCredentials()
	}
	s.logger.WithField("user_id", user.ID).Info("user found")

	// Шаг 2: Проверка активности пользователя
	if !user.IsActive {
		s.logger.WithField("user_id", user.ID).Warn("inactive user login attempt")
		return nil, errors.UserInactive()
	}

	// Шаг 3: Проверка пароля
	s.logger.WithField("user_id", user.ID).Info("verifying password")
	valid, err := s.passwordHasher.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", user.ID).Error("failed to verify password")
		return nil, errors.Internal("authentication failed")
	}

	if !valid {
		s.logger.WithField("user_id", user.ID).Warn("invalid password")
		return nil, errors.InvalidCredentials()
	}
	s.logger.WithField("user_id", user.ID).Info("password verified successfully")

	// Шаг 4: Обновление времени последнего входа
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.WithError(err).WithField("user_id", user.ID).Warn("failed to update last login time")
	}

	// Шаг 5: Получение ролей пользователя
	s.logger.WithField("user_id", user.ID).Info("getting user roles")
	userRoles, err := s.roleRepo.GetUserRoles(ctx, user.ID)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", user.ID).Error("failed to get user roles")
		return nil, errors.DatabaseError(fmt.Errorf("failed to retrieve user roles: %w", err))
	}

	roleNames := make([]string, len(userRoles))
	for i, role := range userRoles {
		roleNames[i] = role.Name
	}
	s.logger.WithFields(logger.Fields{
		"user_id": user.ID,
		"roles":   roleNames,
	}).Info("user roles retrieved")

	// Шаг 6: Генерация токенов
	s.logger.WithField("user_id", user.ID).Info("generating access token")
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Username, roleNames, s.accessExpiry)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", user.ID).Error("failed to generate access token")
		return nil, errors.Internal("failed to generate tokens")
	}

	s.logger.WithField("user_id", user.ID).Info("generating refresh token")
	// Генерируем короткий refresh token
	refreshToken, err := utils.GenerateSecureToken()
	if err != nil {
		s.logger.WithError(err).WithField("user_id", user.ID).Error("failed to generate refresh token")
		return nil, errors.Internal("failed to generate tokens")
	}

	// Шаг 7: Создание сессии
	s.logger.WithFields(logger.Fields{
		"user_id":              user.ID,
		"ip_address":           ipAddress,
		"user_agent":           userAgent,
		"refresh_token_length": len(refreshToken),
	}).Info("creating session")

	session := &entities.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		IsActive:     true,
		ExpiresAt:    time.Now().Add(s.refreshExpiry),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":              user.ID,
			"session_id":           session.ID,
			"ip_address":           ipAddress,
			"user_agent":           userAgent,
			"expires_at":           session.ExpiresAt,
			"refresh_token_length": len(refreshToken),
		}).Error("failed to create session")
		return nil, errors.DatabaseError(fmt.Errorf("failed to create session: %w", err))
	}

	s.logger.WithFields(logger.Fields{
		"user_id":    user.ID,
		"session_id": session.ID,
	}).Info("session created successfully")

	// Шаг 8: Публикация события (игнорируем ошибки)
	event := kafka.UserLoggedInEvent{
		BaseEvent: kafka.NewBaseEvent(kafka.TopicUserLoggedIn),
		UserID:    user.ID,
		Email:     user.Email,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	if err := s.producer.PublishMessage(ctx, kafka.TopicUserLoggedIn, user.ID.String(), event); err != nil {
		s.logger.WithError(err).Warn("failed to publish user logged in event")
	}

	s.logger.WithField("user_id", user.ID).Info("login completed successfully")

	return &response.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessExpiry.Seconds()),
		User: &response.UserResponse{
			ID:          user.ID,
			Email:       user.Email,
			Username:    user.Username,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			IsActive:    user.IsActive,
			IsVerified:  user.IsVerified,
			LastLoginAt: user.LastLoginAt,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		},
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *request.RefreshTokenRequest) (*response.TokenResponse, error) {
	// Для простых refresh токенов проверяем через базу данных
	session, err := s.sessionRepo.GetByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, errors.TokenInvalid()
	}

	if !session.IsActive || time.Now().After(session.ExpiresAt) {
		return nil, errors.TokenExpired()
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, errors.UserNotFound()
	}

	if !user.IsActive {
		return nil, errors.UserInactive()
	}

	// Получаем роли пользователя (с обработкой ошибок)
	userRoles, err := s.roleRepo.GetUserRoles(ctx, user.ID)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", user.ID).Warn("failed to get user roles, using empty roles")
		userRoles = []*entities.Role{}
	}

	roleNames := make([]string, len(userRoles))
	for i, role := range userRoles {
		roleNames[i] = role.Name
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Username, roleNames, s.accessExpiry)
	if err != nil {
		s.logger.WithError(err).Error("failed to generate access token")
		return nil, errors.Internal("failed to generate token")
	}

	return &response.TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.accessExpiry.Seconds()),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, req *request.LogoutRequest) error {
	session, err := s.sessionRepo.GetByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil
	}

	if err := s.sessionRepo.Delete(ctx, session.ID); err != nil {
		return err
	}

	event := kafka.UserLoggedOutEvent{
		BaseEvent: kafka.NewBaseEvent(kafka.TopicUserLoggedOut),
		UserID:    session.UserID,
		SessionID: session.ID,
	}

	if err := s.producer.PublishMessage(ctx, kafka.TopicUserLoggedOut, session.UserID.String(), event); err != nil {
		s.logger.WithError(err).Warn("failed to publish user logged out event")
	}

	return nil
}

func (s *AuthService) LogoutAll(ctx context.Context, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return errors.Validation("invalid user ID")
	}

	if err := s.sessionRepo.DeleteByUserID(ctx, uid); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) VerifyToken(ctx context.Context, token string) (*response.TokenClaimsResponse, error) {
	claims, err := s.jwtManager.ValidateAccessToken(token)
	if err != nil {
		return nil, errors.TokenInvalid()
	}

	return &response.TokenClaimsResponse{
		UserID:    claims.UserID.String(),
		Email:     claims.Email,
		Username:  claims.Username,
		Roles:     claims.Roles,
		ExpiresAt: claims.ExpiresAt.Time,
		IssuedAt:  claims.IssuedAt.Time,
	}, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, req *request.ChangePasswordRequest) error {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return errors.Validation("invalid user ID")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	valid, err := s.passwordHasher.VerifyPassword(req.OldPassword, user.PasswordHash)
	if err != nil {
		s.logger.WithError(err).Error("failed to verify old password")
		return errors.Internal("password verification failed")
	}

	if !valid {
		return errors.InvalidCredentials()
	}

	if !utils.IsValidPassword(req.NewPassword) {
		return errors.WeakPassword()
	}

	newPasswordHash, err := s.passwordHasher.HashPassword(req.NewPassword)
	if err != nil {
		s.logger.WithError(err).Error("failed to hash new password")
		return errors.Internal("failed to process new password")
	}

	user.PasswordHash = newPasswordHash
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	if err := s.sessionRepo.DeleteByUserID(ctx, user.ID); err != nil {
		s.logger.WithError(err).Warn("failed to delete user sessions after password change")
	}

	event := kafka.PasswordChangedEvent{
		BaseEvent: kafka.NewBaseEvent(kafka.TopicPasswordChanged),
		UserID:    user.ID,
		Email:     user.Email,
	}

	if err := s.producer.PublishMessage(ctx, kafka.TopicPasswordChanged, user.ID.String(), event); err != nil {
		s.logger.WithError(err).Warn("failed to publish password changed event")
	}

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req *request.ResetPasswordRequest) error {
	_, err := s.userRepo.GetByEmail(ctx, utils.NormalizeEmail(req.Email))
	if err != nil {
		return nil
	}

	return nil
}

func (s *AuthService) ConfirmResetPassword(ctx context.Context, req *request.ConfirmResetPasswordRequest) error {
	return nil
}
