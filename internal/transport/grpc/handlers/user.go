package handlers

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/vagonaizer/authenitfication-service/api/proto/generated"
	"github.com/vagonaizer/authenitfication-service/internal/domain/services"
	"github.com/vagonaizer/authenitfication-service/internal/dto/request"
	"github.com/vagonaizer/authenitfication-service/pkg/errors"
	"github.com/vagonaizer/authenitfication-service/pkg/logger"
)

type UserGRPCHandler struct {
	generated.UnimplementedUserServiceServer
	userService services.UserService
	logger      *logger.Logger
}

func NewUserGRPCHandler(userService services.UserService, logger *logger.Logger) *UserGRPCHandler {
	return &UserGRPCHandler{
		userService: userService,
		logger:      logger,
	}
}

func (h *UserGRPCHandler) GetProfile(ctx context.Context, req *generated.GetProfileRequest) (*generated.UserResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	result, err := h.userService.GetProfile(ctx, userID)
	if err != nil {
		return nil, h.handleError(err)
	}

	var lastLoginAt *timestamppb.Timestamp
	if result.LastLoginAt != nil {
		lastLoginAt = timestamppb.New(*result.LastLoginAt)
	}

	return &generated.UserResponse{
		Id:          result.ID.String(),
		Email:       result.Email,
		Username:    result.Username,
		FirstName:   h.stringPtrToString(result.FirstName),
		LastName:    h.stringPtrToString(result.LastName),
		IsActive:    result.IsActive,
		IsVerified:  result.IsVerified,
		LastLoginAt: lastLoginAt,
		CreatedAt:   timestamppb.New(result.CreatedAt),
		UpdatedAt:   timestamppb.New(result.UpdatedAt),
	}, nil
}

func (h *UserGRPCHandler) UpdateProfile(ctx context.Context, req *generated.UpdateProfileRequest) (*generated.UserResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	updateReq := &request.UpdateUserRequest{
		UserID: userID,
	}

	if req.FirstName != nil {
		updateReq.FirstName = req.FirstName
	}
	if req.LastName != nil {
		updateReq.LastName = req.LastName
	}
	if req.Username != nil {
		updateReq.Username = req.Username
	}

	result, err := h.userService.UpdateProfile(ctx, updateReq)
	if err != nil {
		return nil, h.handleError(err)
	}

	var lastLoginAt *timestamppb.Timestamp
	if result.LastLoginAt != nil {
		lastLoginAt = timestamppb.New(*result.LastLoginAt)
	}

	return &generated.UserResponse{
		Id:          result.ID.String(),
		Email:       result.Email,
		Username:    result.Username,
		FirstName:   h.stringPtrToString(result.FirstName),
		LastName:    h.stringPtrToString(result.LastName),
		IsActive:    result.IsActive,
		IsVerified:  result.IsVerified,
		LastLoginAt: lastLoginAt,
		CreatedAt:   timestamppb.New(result.CreatedAt),
		UpdatedAt:   timestamppb.New(result.UpdatedAt),
	}, nil
}

func (h *UserGRPCHandler) DeleteAccount(ctx context.Context, req *generated.DeleteAccountRequest) (*generated.DeleteAccountResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	err = h.userService.DeleteAccount(ctx, userID)
	if err != nil {
		return nil, h.handleError(err)
	}

	return &generated.DeleteAccountResponse{
		Message: "Account deleted successfully",
	}, nil
}

func (h *UserGRPCHandler) ListUsers(ctx context.Context, req *generated.ListUsersRequest) (*generated.UsersListResponse, error) {
	listReq := &request.ListUsersRequest{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
		Search:   req.Search,
		SortBy:   req.SortBy,
		SortDir:  req.SortDir,
	}

	result, err := h.userService.ListUsers(ctx, listReq)
	if err != nil {
		return nil, h.handleError(err)
	}

	users := make([]*generated.UserResponse, len(result.Users))
	for i, user := range result.Users {
		var lastLoginAt *timestamppb.Timestamp
		if user.LastLoginAt != nil {
			lastLoginAt = timestamppb.New(*user.LastLoginAt)
		}

		users[i] = &generated.UserResponse{
			Id:          user.ID.String(),
			Email:       user.Email,
			Username:    user.Username,
			FirstName:   h.stringPtrToString(user.FirstName),
			LastName:    h.stringPtrToString(user.LastName),
			IsActive:    user.IsActive,
			IsVerified:  user.IsVerified,
			LastLoginAt: lastLoginAt,
			CreatedAt:   timestamppb.New(user.CreatedAt),
			UpdatedAt:   timestamppb.New(user.UpdatedAt),
		}
	}

	return &generated.UsersListResponse{
		Users:      users,
		Total:      result.Total,
		Page:       int32(result.Page),
		PageSize:   int32(result.PageSize),
		TotalPages: int32(result.TotalPages),
	}, nil
}

func (h *UserGRPCHandler) GetUserByID(ctx context.Context, req *generated.GetUserByIDRequest) (*generated.UserResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	result, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, h.handleError(err)
	}

	var lastLoginAt *timestamppb.Timestamp
	if result.LastLoginAt != nil {
		lastLoginAt = timestamppb.New(*result.LastLoginAt)
	}

	return &generated.UserResponse{
		Id:          result.ID.String(),
		Email:       result.Email,
		Username:    result.Username,
		FirstName:   h.stringPtrToString(result.FirstName),
		LastName:    h.stringPtrToString(result.LastName),
		IsActive:    result.IsActive,
		IsVerified:  result.IsVerified,
		LastLoginAt: lastLoginAt,
		CreatedAt:   timestamppb.New(result.CreatedAt),
		UpdatedAt:   timestamppb.New(result.UpdatedAt),
	}, nil
}

func (h *UserGRPCHandler) ActivateUser(ctx context.Context, req *generated.ActivateUserRequest) (*generated.ActivateUserResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	err = h.userService.ActivateUser(ctx, userID)
	if err != nil {
		return nil, h.handleError(err)
	}

	return &generated.ActivateUserResponse{
		Message: "User activated successfully",
	}, nil
}

func (h *UserGRPCHandler) DeactivateUser(ctx context.Context, req *generated.DeactivateUserRequest) (*generated.DeactivateUserResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	err = h.userService.DeactivateUser(ctx, userID)
	if err != nil {
		return nil, h.handleError(err)
	}

	return &generated.DeactivateUserResponse{
		Message: "User deactivated successfully",
	}, nil
}

func (h *UserGRPCHandler) AssignRole(ctx context.Context, req *generated.AssignRoleRequest) (*generated.AssignRoleResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid role ID format")
	}

	assignReq := &request.AssignRoleRequest{
		UserID: userID,
		RoleID: roleID,
	}

	err = h.userService.AssignRole(ctx, assignReq)
	if err != nil {
		return nil, h.handleError(err)
	}

	return &generated.AssignRoleResponse{
		Message: "Role assigned successfully",
	}, nil
}

func (h *UserGRPCHandler) RemoveRole(ctx context.Context, req *generated.RemoveRoleRequest) (*generated.RemoveRoleResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid role ID format")
	}

	removeReq := &request.RemoveRoleRequest{
		UserID: userID,
		RoleID: roleID,
	}

	err = h.userService.RemoveRole(ctx, removeReq)
	if err != nil {
		return nil, h.handleError(err)
	}

	return &generated.RemoveRoleResponse{
		Message: "Role removed successfully",
	}, nil
}

func (h *UserGRPCHandler) GetUserRoles(ctx context.Context, req *generated.GetUserRolesRequest) (*generated.UserRolesResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	result, err := h.userService.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, h.handleError(err)
	}

	roles := make([]*generated.Role, len(result.Roles))
	for i, role := range result.Roles {
		roles[i] = &generated.Role{
			Id:          role.ID.String(),
			Name:        role.Name,
			Description: h.stringPtrToString(role.Description),
			CreatedAt:   timestamppb.New(role.CreatedAt),
		}
	}

	return &generated.UserRolesResponse{
		UserId: result.UserID.String(),
		Roles:  roles,
	}, nil
}

func (h *UserGRPCHandler) handleError(err error) error {
	if appErr, ok := err.(*errors.AppError); ok {
		switch appErr.Code {
		case errors.CodeValidation:
			return status.Error(codes.InvalidArgument, appErr.Message)
		case errors.CodeNotFound:
			return status.Error(codes.NotFound, appErr.Message)
		case errors.CodeAlreadyExists:
			return status.Error(codes.AlreadyExists, appErr.Message)
		case errors.CodeUnauthorized:
			return status.Error(codes.Unauthenticated, appErr.Message)
		case errors.CodeForbidden:
			return status.Error(codes.PermissionDenied, appErr.Message)
		default:
			return status.Error(codes.Internal, appErr.Message)
		}
	}
	return status.Error(codes.Internal, "Internal server error")
}

func (h *UserGRPCHandler) stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
