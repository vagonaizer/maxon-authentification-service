package interceptors

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/vagonaizer/authenitfication-service/pkg/auth"
	"github.com/vagonaizer/authenitfication-service/pkg/logger"
)

type AuthInterceptor struct {
	jwtManager *auth.JWTManager
	logger     *logger.Logger
}

func NewAuthInterceptor(jwtManager *auth.JWTManager, logger *logger.Logger) *AuthInterceptor {
	return &AuthInterceptor{
		jwtManager: jwtManager,
		logger:     logger,
	}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if i.isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		token, err := i.extractToken(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "missing or invalid token")
		}

		claims, err := i.jwtManager.ValidateAccessToken(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = i.setUserContext(ctx, claims)
		return handler(ctx, req)
	}
}

func (i *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if i.isPublicMethod(info.FullMethod) {
			return handler(srv, ss)
		}

		token, err := i.extractToken(ss.Context())
		if err != nil {
			return status.Error(codes.Unauthenticated, "missing or invalid token")
		}

		claims, err := i.jwtManager.ValidateAccessToken(token)
		if err != nil {
			return status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx := i.setUserContext(ss.Context(), claims)
		wrapped := &wrappedStream{ServerStream: ss, ctx: ctx}
		return handler(srv, wrapped)
	}
}

func (i *AuthInterceptor) extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization header")
	}

	authHeader := authHeaders[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", status.Error(codes.Unauthenticated, "invalid authorization header format")
	}

	return authHeader[7:], nil
}

func (i *AuthInterceptor) setUserContext(ctx context.Context, claims *auth.AccessTokenClaims) context.Context {
	ctx = context.WithValue(ctx, "user_id", claims.UserID.String())
	ctx = context.WithValue(ctx, "email", claims.Email)
	ctx = context.WithValue(ctx, "username", claims.Username)
	ctx = context.WithValue(ctx, "roles", claims.Roles)
	return ctx
}

func (i *AuthInterceptor) isPublicMethod(method string) bool {
	publicMethods := []string{
		"/auth.v1.AuthService/Register",
		"/auth.v1.AuthService/Login",
		"/auth.v1.AuthService/RefreshToken",
		"/auth.v1.AuthService/VerifyToken",
	}

	for _, publicMethod := range publicMethods {
		if method == publicMethod {
			return true
		}
	}
	return false
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}
