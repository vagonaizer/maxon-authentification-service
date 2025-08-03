package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/vagonaizer/authenitfication-service/internal/dto/request"
	"github.com/vagonaizer/authenitfication-service/internal/dto/response"
)

type UserService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*response.UserResponse, error)
	UpdateProfile(ctx context.Context, req *request.UpdateUserRequest) (*response.UserResponse, error)
	DeleteAccount(ctx context.Context, userID uuid.UUID) error
	ListUsers(ctx context.Context, req *request.ListUsersRequest) (*response.UsersListResponse, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*response.UserResponse, error)
	ActivateUser(ctx context.Context, userID uuid.UUID) error
	DeactivateUser(ctx context.Context, userID uuid.UUID) error
	AssignRole(ctx context.Context, req *request.AssignRoleRequest) error
	RemoveRole(ctx context.Context, req *request.RemoveRoleRequest) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) (*response.UserRolesResponse, error)
}
