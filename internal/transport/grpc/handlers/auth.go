package handlers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/vagonaizer/authenitfication-service/api/proto/generated"
	"github.com/vagonaizer/authenitfication-service/internal/domain/services"
	"github.com/vagonaizer/authenitfication-service/internal/dto/request"
	"github.com/vagonaizer/authenitfication-service/pkg/errors"
	"github.com/vagonaizer/authenitfication-service/pkg/logger"
)

type AuthGRPCHandler struct {
	generated.UnimplementedAuthServiceServer
	authService services.AuthService
	logger      *logger.Logger
}

func NewAuthGRPCHandler(authService services.AuthService, logger *logger.Logger) *AuthGRPCHandler {
	return &AuthGRPCHandler{
		authService: authService,
		logger:      logger,
	}
}

func (h *AuthGRPCHandler) Register(ctx context.Context, req *generated.RegisterRequest) (*generated.AuthResponse, error) {
	registerReq := &request.RegisterRequest{
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	// Для gRPC используем значения по умолчанию
	ipAddress := "127.0.0.1"
	userAgent := "gRPC-Client"

	result, err := h.authService.Register(ctx, registerReq, ipAddress, userAgent)
	if err != nil {
		return nil, h.handleError(err)
	}

	var lastLoginAt *timestamppb.Timestamp
	if result.User.LastLoginAt != nil {
		lastLoginAt = timestamppb.New(*result.User.LastLoginAt)
	}

	return &generated.AuthResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    result.TokenType,
		ExpiresIn:    result.ExpiresIn,
		User: &generated.User{
			Id:          result.User.ID.String(),
			Email:       result.User.Email,
			Username:    result.User.Username,
			FirstName:   h.stringPtrToString(result.User.FirstName),
			LastName:    h.stringPtrToString(result.User.LastName),
			IsActive:    result.User.IsActive,
			IsVerified:  result.User.IsVerified,
			LastLoginAt: lastLoginAt,
			CreatedAt:   timestamppb.New(result.User.CreatedAt),
			UpdatedAt:   timestamppb.New(result.User.UpdatedAt),
		},
	}, nil
}

func (h *AuthGRPCHandler) Login(ctx context.Context, req *generated.LoginRequest) (*generated.AuthResponse, error) {
	loginReq := &request.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	// Для gRPC используем значения по умолчанию
	ipAddress := "127.0.0.1"
	userAgent := "gRPC-Client"

	result, err := h.authService.Login(ctx, loginReq, ipAddress, userAgent)
	if err != nil {
		return nil, h.handleError(err)
	}

	var lastLoginAt *timestamppb.Timestamp
	if result.User.LastLoginAt != nil {
		lastLoginAt = timestamppb.New(*result.User.LastLoginAt)
	}

	return &generated.AuthResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    result.TokenType,
		ExpiresIn:    result.ExpiresIn,
		User: &generated.User{
			Id:          result.User.ID.String(),
			Email:       result.User.Email,
			Username:    result.User.Username,
			FirstName:   h.stringPtrToString(result.User.FirstName),
			LastName:    h.stringPtrToString(result.User.LastName),
			IsActive:    result.User.IsActive,
			IsVerified:  result.User.IsVerified,
			LastLoginAt: lastLoginAt,
			CreatedAt:   timestamppb.New(result.User.CreatedAt),
			UpdatedAt:   timestamppb.New(result.User.UpdatedAt),
		},
	}, nil
}

func (h *AuthGRPCHandler) RefreshToken(ctx context.Context, req *generated.RefreshTokenRequest) (*generated.TokenResponse, error) {
	refreshReq := &request.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	}

	result, err := h.authService.RefreshToken(ctx, refreshReq)
	if err != nil {
		return nil, h.handleError(err)
	}

	return &generated.TokenResponse{
		AccessToken: result.AccessToken,
		TokenType:   result.TokenType,
		ExpiresIn:   result.ExpiresIn,
	}, nil
}

func (h *AuthGRPCHandler) Logout(ctx context.Context, req *generated.LogoutRequest) (*generated.LogoutResponse, error) {
	logoutReq := &request.LogoutRequest{
		RefreshToken: req.RefreshToken,
	}

	err := h.authService.Logout(ctx, logoutReq)
	if err != nil {
		return nil, h.handleError(err)
	}

	return &generated.LogoutResponse{
		Message: "Logged out successfully",
	}, nil
}

func (h *AuthGRPCHandler) VerifyToken(ctx context.Context, req *generated.VerifyTokenRequest) (*generated.TokenClaimsResponse, error) {
	result, err := h.authService.VerifyToken(ctx, req.Token)
	if err != nil {
		return nil, h.handleError(err)
	}

	return &generated.TokenClaimsResponse{
		UserId:    result.UserID,
		Email:     result.Email,
		Username:  result.Username,
		Roles:     result.Roles,
		ExpiresAt: timestamppb.New(result.ExpiresAt),
		IssuedAt:  timestamppb.New(result.IssuedAt),
	}, nil
}

func (h *AuthGRPCHandler) ChangePassword(ctx context.Context, req *generated.ChangePasswordRequest) (*generated.ChangePasswordResponse, error) {
	changeReq := &request.ChangePasswordRequest{
		UserID:      req.UserId,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}

	err := h.authService.ChangePassword(ctx, changeReq)
	if err != nil {
		return nil, h.handleError(err)
	}

	return &generated.ChangePasswordResponse{
		Message: "Password changed successfully",
	}, nil
}

func (h *AuthGRPCHandler) handleError(err error) error {
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
		case errors.CodeInvalidCredentials:
			return status.Error(codes.Unauthenticated, appErr.Message)
		case errors.CodeTokenExpired:
			return status.Error(codes.Unauthenticated, appErr.Message)
		case errors.CodeTokenInvalid:
			return status.Error(codes.Unauthenticated, appErr.Message)
		default:
			return status.Error(codes.Internal, appErr.Message)
		}
	}
	return status.Error(codes.Internal, "Internal server error")
}

func (h *AuthGRPCHandler) stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
