package request

import "github.com/google/uuid"

type UpdateUserRequest struct {
	UserID    uuid.UUID `json:"-"`
	FirstName *string   `json:"first_name" validate:"omitempty,max=100"`
	LastName  *string   `json:"last_name" validate:"omitempty,max=100"`
	Username  *string   `json:"username" validate:"omitempty,min=3,max=50"`
}

type ListUsersRequest struct {
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
	Search   string `json:"search" validate:"max=255"`
	SortBy   string `json:"sort_by" validate:"oneof=created_at updated_at email username"`
	SortDir  string `json:"sort_dir" validate:"oneof=asc desc"`
}

type AssignRoleRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}

type RemoveRoleRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}
