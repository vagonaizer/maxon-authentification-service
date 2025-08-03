package services

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TokenService interface {
	GenerateAccessToken(ctx context.Context, userID uuid.UUID, roles []string) (string, error)
	GenerateRefreshToken(ctx context.Context) (string, error)
	ValidateAccessToken(ctx context.Context, token string) (*TokenClaims, error)
	ValidateRefreshToken(ctx context.Context, token string) (*TokenClaims, error)
	RevokeToken(ctx context.Context, token string) error
	GetTokenExpiration(ctx context.Context, token string) (time.Time, error)
}

type TokenClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Roles     []string  `json:"roles"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
}
