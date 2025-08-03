package services

import (
	"context"

	"github.com/vagonaizer/authenitfication-service/internal/dto/request"
	"github.com/vagonaizer/authenitfication-service/internal/dto/response"
)

type AuthService interface {
	Register(ctx context.Context, req *request.RegisterRequest, ipAddress, userAgent string) (*response.AuthResponse, error)
	Login(ctx context.Context, req *request.LoginRequest, ipAddress, userAgent string) (*response.AuthResponse, error)
	RefreshToken(ctx context.Context, req *request.RefreshTokenRequest) (*response.TokenResponse, error)
	Logout(ctx context.Context, req *request.LogoutRequest) error
	LogoutAll(ctx context.Context, userID string) error
	VerifyToken(ctx context.Context, token string) (*response.TokenClaimsResponse, error)
	ChangePassword(ctx context.Context, req *request.ChangePasswordRequest) error
	ResetPassword(ctx context.Context, req *request.ResetPasswordRequest) error
	ConfirmResetPassword(ctx context.Context, req *request.ConfirmResetPasswordRequest) error
}
