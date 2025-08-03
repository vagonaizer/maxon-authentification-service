package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vagonaizer/authenitfication-service/internal/domain/services"
	"github.com/vagonaizer/authenitfication-service/pkg/auth"
	"github.com/vagonaizer/authenitfication-service/pkg/errors"
	"github.com/vagonaizer/authenitfication-service/pkg/logger"
)

type tokenService struct {
	jwtManager *auth.JWTManager
	logger     *logger.Logger
}

func NewTokenService(jwtManager *auth.JWTManager, logger *logger.Logger) *tokenService {
	return &tokenService{
		jwtManager: jwtManager,
		logger:     logger,
	}
}

func (s *tokenService) GenerateAccessToken(ctx context.Context, userID uuid.UUID, roles []string) (string, error) {
	return s.jwtManager.GenerateAccessToken(userID, "", "", roles, 15*time.Minute)
}

func (s *tokenService) GenerateRefreshToken(ctx context.Context) (string, error) {
	return s.jwtManager.GenerateRefreshToken(uuid.New(), 24*time.Hour*7)
}

func (s *tokenService) ValidateAccessToken(ctx context.Context, token string) (*services.TokenClaims, error) {
	claims, err := s.jwtManager.ValidateAccessToken(token)
	if err != nil {
		return nil, errors.TokenInvalid()
	}

	return &services.TokenClaims{
		UserID:    claims.UserID,
		Email:     claims.Email,
		Username:  claims.Username,
		Roles:     claims.Roles,
		ExpiresAt: claims.ExpiresAt.Time,
		IssuedAt:  claims.IssuedAt.Time,
	}, nil
}

func (s *tokenService) ValidateRefreshToken(ctx context.Context, token string) (*services.TokenClaims, error) {
	claims, err := s.jwtManager.ValidateRefreshToken(token)
	if err != nil {
		return nil, errors.TokenInvalid()
	}

	return &services.TokenClaims{
		UserID:    claims.UserID,
		ExpiresAt: claims.ExpiresAt.Time,
		IssuedAt:  claims.IssuedAt.Time,
	}, nil
}

func (s *tokenService) RevokeToken(ctx context.Context, token string) error {
	return nil
}

func (s *tokenService) GetTokenExpiration(ctx context.Context, token string) (time.Time, error) {
	return s.jwtManager.GetTokenExpiration(token)
}
