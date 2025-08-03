package kafka

import (
	"time"

	"github.com/google/uuid"
)

const (
	TopicUserRegistered  = "user.registered"
	TopicUserLoggedIn    = "user.logged_in"
	TopicUserLoggedOut   = "user.logged_out"
	TopicPasswordChanged = "user.password_changed"
	TopicUserActivated   = "user.activated"
	TopicUserDeactivated = "user.deactivated"
	TopicUserDeleted     = "user.deleted"
	TopicRoleAssigned    = "user.role_assigned"
	TopicRoleRemoved     = "user.role_removed"
)

type BaseEvent struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

type UserRegisteredEvent struct {
	BaseEvent
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	FirstName *string   `json:"first_name"`
	LastName  *string   `json:"last_name"`
}

type UserLoggedInEvent struct {
	BaseEvent
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
}

type UserLoggedOutEvent struct {
	BaseEvent
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	SessionID uuid.UUID `json:"session_id"`
}

type PasswordChangedEvent struct {
	BaseEvent
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

type UserActivatedEvent struct {
	BaseEvent
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

type UserDeactivatedEvent struct {
	BaseEvent
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

type UserDeletedEvent struct {
	BaseEvent
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

type RoleAssignedEvent struct {
	BaseEvent
	UserID   uuid.UUID `json:"user_id"`
	RoleID   uuid.UUID `json:"role_id"`
	RoleName string    `json:"role_name"`
}

type RoleRemovedEvent struct {
	BaseEvent
	UserID   uuid.UUID `json:"user_id"`
	RoleID   uuid.UUID `json:"role_id"`
	RoleName string    `json:"role_name"`
}

func NewBaseEvent(eventType string) BaseEvent {
	return BaseEvent{
		ID:        uuid.New(),
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Version:   "1.0",
	}
}
