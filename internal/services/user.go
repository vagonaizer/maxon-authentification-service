package services

import (
	"context"
	"math"

	"github.com/google/uuid"
	"github.com/vagonaizer/authenitfication-service/internal/domain/repositories"
	"github.com/vagonaizer/authenitfication-service/internal/dto/request"
	"github.com/vagonaizer/authenitfication-service/internal/dto/response"
	"github.com/vagonaizer/authenitfication-service/internal/infrastructure/messaging/kafka"
	"github.com/vagonaizer/authenitfication-service/pkg/errors"
	"github.com/vagonaizer/authenitfication-service/pkg/logger"
	"github.com/vagonaizer/authenitfication-service/pkg/utils"
)

type userService struct {
	userRepo repositories.UserRepository
	roleRepo repositories.RoleRepository
	producer *kafka.Producer
	logger   *logger.Logger
}

func NewUserService(
	userRepo repositories.UserRepository,
	roleRepo repositories.RoleRepository,
	producer *kafka.Producer,
	logger *logger.Logger,
) *userService {
	return &userService{
		userRepo: userRepo,
		roleRepo: roleRepo,
		producer: producer,
		logger:   logger,
	}
}

func (s *userService) GetProfile(ctx context.Context, userID uuid.UUID) (*response.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &response.UserResponse{
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
	}, nil
}

func (s *userService) UpdateProfile(ctx context.Context, req *request.UpdateUserRequest) (*response.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	if req.FirstName != nil {
		user.FirstName = req.FirstName
	}

	if req.LastName != nil {
		user.LastName = req.LastName
	}

	if req.Username != nil {
		if !utils.IsValidUsername(*req.Username) {
			return nil, errors.Validation("invalid username format")
		}

		normalizedUsername := utils.NormalizeUsername(*req.Username)
		if normalizedUsername != user.Username {
			exists, err := s.userRepo.ExistsByUsername(ctx, normalizedUsername)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, errors.UsernameExists()
			}
			user.Username = normalizedUsername
		}
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return &response.UserResponse{
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
	}, nil
}

func (s *userService) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return err
	}

	user, _ := s.userRepo.GetByID(ctx, userID)
	if user != nil {
		event := kafka.UserDeletedEvent{
			BaseEvent: kafka.NewBaseEvent(kafka.TopicUserDeleted),
			UserID:    user.ID,
			Email:     user.Email,
		}

		if err := s.producer.PublishMessage(ctx, kafka.TopicUserDeleted, user.ID.String(), event); err != nil {
			s.logger.WithError(err).Warn("failed to publish user deleted event")
		}
	}

	return nil
}

func (s *userService) ListUsers(ctx context.Context, req *request.ListUsersRequest) (*response.UsersListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize
	users, err := s.userRepo.List(ctx, req.PageSize, offset)
	if err != nil {
		return nil, err
	}

	userResponses := make([]*response.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = &response.UserResponse{
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
		}
	}

	total := int64(len(users))
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	return &response.UsersListResponse{
		Users:      userResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *userService) GetUserByID(ctx context.Context, userID uuid.UUID) (*response.UserResponse, error) {
	return s.GetProfile(ctx, userID)
}

func (s *userService) ActivateUser(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.IsActive {
		return nil
	}

	user.IsActive = true
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	event := kafka.UserActivatedEvent{
		BaseEvent: kafka.NewBaseEvent(kafka.TopicUserActivated),
		UserID:    user.ID,
		Email:     user.Email,
	}

	if err := s.producer.PublishMessage(ctx, kafka.TopicUserActivated, user.ID.String(), event); err != nil {
		s.logger.WithError(err).Warn("failed to publish user activated event")
	}

	return nil
}

func (s *userService) DeactivateUser(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.IsActive {
		return nil
	}

	user.IsActive = false
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	event := kafka.UserDeactivatedEvent{
		BaseEvent: kafka.NewBaseEvent(kafka.TopicUserDeactivated),
		UserID:    user.ID,
		Email:     user.Email,
	}

	if err := s.producer.PublishMessage(ctx, kafka.TopicUserDeactivated, user.ID.String(), event); err != nil {
		s.logger.WithError(err).Warn("failed to publish user deactivated event")
	}

	return nil
}

func (s *userService) AssignRole(ctx context.Context, req *request.AssignRoleRequest) error {
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return err
	}

	role, err := s.roleRepo.GetByID(ctx, req.RoleID)
	if err != nil {
		return err
	}

	if err := s.roleRepo.AssignRoleToUser(ctx, req.UserID, req.RoleID); err != nil {
		return err
	}

	event := kafka.RoleAssignedEvent{
		BaseEvent: kafka.NewBaseEvent(kafka.TopicRoleAssigned),
		UserID:    user.ID,
		RoleID:    role.ID,
		RoleName:  role.Name,
	}

	if err := s.producer.PublishMessage(ctx, kafka.TopicRoleAssigned, user.ID.String(), event); err != nil {
		s.logger.WithError(err).Warn("failed to publish role assigned event")
	}

	return nil
}

func (s *userService) RemoveRole(ctx context.Context, req *request.RemoveRoleRequest) error {
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return err
	}

	role, err := s.roleRepo.GetByID(ctx, req.RoleID)
	if err != nil {
		return err
	}

	if err := s.roleRepo.RemoveRoleFromUser(ctx, req.UserID, req.RoleID); err != nil {
		return err
	}

	event := kafka.RoleRemovedEvent{
		BaseEvent: kafka.NewBaseEvent(kafka.TopicRoleRemoved),
		UserID:    user.ID,
		RoleID:    role.ID,
		RoleName:  role.Name,
	}

	if err := s.producer.PublishMessage(ctx, kafka.TopicRoleRemoved, user.ID.String(), event); err != nil {
		s.logger.WithError(err).Warn("failed to publish role removed event")
	}

	return nil
}

func (s *userService) GetUserRoles(ctx context.Context, userID uuid.UUID) (*response.UserRolesResponse, error) {
	roles, err := s.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	roleResponses := make([]*response.RoleResponse, len(roles))
	for i, role := range roles {
		roleResponses[i] = &response.RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			CreatedAt:   role.CreatedAt,
		}
	}

	return &response.UserRolesResponse{
		UserID: userID,
		Roles:  roleResponses,
	}, nil
}
