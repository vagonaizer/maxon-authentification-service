package services

import (
	"context"
	"time"

	"github.com/google/uuid"
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

type authService struct {
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
	accessExpiry, refreshExpiry time.Duration,
) *authService {
	return &authService{
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

func (s *authService) Register(ctx context.Context, req *request.RegisterRequest) (*response.AuthResponse, error) {
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

	defaultRole, err := s.roleRepo.GetByName(ctx, "user")
	if err != nil {
		s.logger.WithError(err).Warn("failed to get default role")
	} else {
		if err := s.roleRepo.AssignRoleToUser(ctx, user.ID, defaultRole.ID); err != nil {
			s.logger.WithError(err).Warn("failed to assign default role")
		}
	}

	userRoles, _ := s.roleRepo.GetUserRoles(ctx, user.ID)
	roleNames := make([]string, len(userRoles))
	for i, role := range userRoles {
		roleNames[i] = role.Name
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Username, roleNames, s.accessExpiry)
	if err != nil {
		s.logger.WithError(err).Error("failed to generate access token")
		return nil, errors.Internal("failed to generate tokens")
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, s.refreshExpiry)
	if err != nil {
		s.logger.WithError(err).Error("failed to generate refresh token")
		return nil, errors.Internal("failed to generate tokens")
	}

	session := &entities.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		IsActive:     true,
		ExpiresAt:    time.Now().Add(s.refreshExpiry),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

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

func (s *authService) Login(ctx context.Context, req *request.LoginRequest) (*response.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, utils.NormalizeEmail(req.Email))
	if err != nil {
		return nil, errors.InvalidCredentials()
	}

	if !user.IsActive {
		return nil, errors.UserInactive()
	}

	valid, err := s.passwordHasher.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil {
		s.logger.WithError(err).Error("failed to verify password")
		return nil, errors.Internal("authentication failed")
	}

	if !valid {
		return nil, errors.InvalidCredentials()
	}

	now := time.Now()
	user.LastLoginAt = &now
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.WithError(err).Warn("failed to update last login time")
	}

	userRoles, _ := s.roleRepo.GetUserRoles(ctx, user.ID)
	roleNames := make([]string, len(userRoles))
	for i, role := range userRoles {
		roleNames[i] = role.Name
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Username, roleNames, s.accessExpiry)
	if err != nil {
		s.logger.WithError(err).Error("failed to generate access token")
		return nil, errors.Internal("failed to generate tokens")
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, s.refreshExpiry)
	if err != nil {
		s.logger.WithError(err).Error("failed to generate refresh token")
		return nil, errors.Internal("failed to generate tokens")
	}

	session := &entities.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		IsActive:     true,
		ExpiresAt:    time.Now().Add(s.refreshExpiry),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	event := kafka.UserLoggedInEvent{
		BaseEvent: kafka.NewBaseEvent(kafka.TopicUserLoggedIn),
		UserID:    user.ID,
		Email:     user.Email,
	}

	if err := s.producer.PublishMessage(ctx, kafka.TopicUserLoggedIn, user.ID.String(), event); err != nil {
		s.logger.WithError(err).Warn("failed to publish user logged in event")
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

func (s *authService) RefreshToken(ctx context.Context, req *request.RefreshTokenRequest) (*response.TokenResponse, error) {
	claims, err := s.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, errors.TokenInvalid()
	}

	session, err := s.sessionRepo.GetByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, errors.TokenInvalid()
	}

	if !session.IsActive || time.Now().After(session.ExpiresAt) {
		return nil, errors.TokenExpired()
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, errors.UserNotFound()
	}

	if !user.IsActive {
		return nil, errors.UserInactive()
	}

	userRoles, _ := s.roleRepo.GetUserRoles(ctx, user.ID)
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

func (s *authService) Logout(ctx context.Context, req *request.LogoutRequest) error {
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

func (s *authService) LogoutAll(ctx context.Context, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return errors.Validation("invalid user ID")
	}

	if err := s.sessionRepo.DeleteByUserID(ctx, uid); err != nil {
		return err
	}

	return nil
}

func (s *authService) VerifyToken(ctx context.Context, token string) (*response.TokenClaimsResponse, error) {
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

func (s *authService) ChangePassword(ctx context.Context, req *request.ChangePasswordRequest) error {
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

func (s *authService) ResetPassword(ctx context.Context, req *request.ResetPasswordRequest) error {
	_, err := s.userRepo.GetByEmail(ctx, utils.NormalizeEmail(req.Email))
	if err != nil {
		return nil
	}

	return nil
}

func (s *authService) ConfirmResetPassword(ctx context.Context, req *request.ConfirmResetPasswordRequest) error {
	return nil
}
