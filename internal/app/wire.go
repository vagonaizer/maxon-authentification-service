//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	"github.com/vagonaizer/authenitfication-service/internal/config"
	"github.com/vagonaizer/authenitfication-service/internal/domain/repositories"
	"github.com/vagonaizer/authenitfication-service/internal/domain/services"
	"github.com/vagonaizer/authenitfication-service/internal/infrastructure/database/postgres"
	postgresrepos "github.com/vagonaizer/authenitfication-service/internal/infrastructure/database/postgres/repositories"
	"github.com/vagonaizer/authenitfication-service/internal/infrastructure/database/redis"
	"github.com/vagonaizer/authenitfication-service/internal/infrastructure/messaging/kafka"
	grpcserver "github.com/vagonaizer/authenitfication-service/internal/interfaces/grpc"
	grpchandlers "github.com/vagonaizer/authenitfication-service/internal/interfaces/grpc/handlers"
	grpcinterceptors "github.com/vagonaizer/authenitfication-service/internal/interfaces/grpc/interceptors"
	httpserver "github.com/vagonaizer/authenitfication-service/internal/interfaces/http"
	httphandlers "github.com/vagonaizer/authenitfication-service/internal/interfaces/http/handlers"
	httpmiddleware "github.com/vagonaizer/authenitfication-service/internal/interfaces/http/middleware"
	appservices "github.com/vagonaizer/authenitfication-service/internal/services"
	"github.com/vagonaizer/authenitfication-service/pkg/auth"
	"github.com/vagonaizer/authenitfication-service/pkg/logger"
)

// Wire sets for dependency injection
var (
	// Infrastructure set
	InfrastructureSet = wire.NewSet(
		// Config
		config.Load,

		// Logger
		ProvideLogger,

		// Database
		postgres.NewConnection,

		// Redis
		redis.NewConnection,
		redis.NewCacheService,

		// Kafka
		kafka.NewProducer,

		// Auth utilities
		auth.NewPasswordHasher,
		ProvideJWTManager,
	)

	// Repository set
	RepositorySet = wire.NewSet(
		postgresrepos.NewUserRepository,
		postgresrepos.NewSessionRepository,
		postgresrepos.NewRoleRepository,

		// Bind interfaces
		wire.Bind(new(repositories.UserRepository), new(*postgresrepos.UserRepository)),
		wire.Bind(new(repositories.SessionRepository), new(*postgresrepos.SessionRepository)),
		wire.Bind(new(repositories.RoleRepository), new(*postgresrepos.RoleRepository)),
	)

	// Service set
	ServiceSet = wire.NewSet(
		appservices.NewTokenService,
		appservices.NewAuthService,
		appservices.NewUserService,
		appservices.NewNotificationService,

		// Bind interfaces
		wire.Bind(new(services.TokenService), new(*appservices.TokenService)),
		wire.Bind(new(services.AuthService), new(*appservices.AuthService)),
		wire.Bind(new(services.UserService), new(*appservices.UserService)),
	)

	// HTTP handler set
	HTTPHandlerSet = wire.NewSet(
		httphandlers.NewAuthHandler,
		httphandlers.NewUserHandler,
		httphandlers.NewHealthHandler,
		httpmiddleware.NewAuthMiddleware,
	)

	// gRPC handler set
	GRPCHandlerSet = wire.NewSet(
		grpchandlers.NewAuthGRPCHandler,
		grpchandlers.NewUserGRPCHandler,
		grpcinterceptors.NewAuthInterceptor,
		grpcinterceptors.NewLoggingInterceptor,
	)

	// Server set
	ServerSet = wire.NewSet(
		httpserver.NewServer,
		grpcserver.NewServer,
	)
)

// Provider functions
func ProvideLogger(cfg *config.Config) *logger.Logger {
	return logger.New(
		cfg.Logger.Level,
		cfg.Logger.Format,
		cfg.Logger.Output,
		cfg.Logger.MaxSize,
		cfg.Logger.MaxBackups,
		cfg.Logger.MaxAge,
		cfg.Logger.Compress,
	)
}

func ProvideJWTManager(cfg *config.Config) *auth.JWTManager {
	return auth.NewJWTManager(
		cfg.JWT.AccessTokenSecret,
		cfg.JWT.RefreshTokenSecret,
		cfg.JWT.Issuer,
		cfg.JWT.Audience,
	)
}

func ProvideAccessTokenExpiry(cfg *config.Config) AccessTokenExpiry {
	return AccessTokenExpiry(cfg.JWT.AccessTokenExpiry)
}

func ProvideRefreshTokenExpiry(cfg *config.Config) RefreshTokenExpiry {
	return RefreshTokenExpiry(cfg.JWT.RefreshTokenExpiry)
}

// Wire injector
func InitializeApp() (*App, error) {
	wire.Build(
		InfrastructureSet,
		RepositorySet,
		ServiceSet,
		HTTPHandlerSet,
		GRPCHandlerSet,
		ServerSet,
		ProvideAccessTokenExpiry,
		ProvideRefreshTokenExpiry,
		NewApp,
	)
	return &App{}, nil
}
