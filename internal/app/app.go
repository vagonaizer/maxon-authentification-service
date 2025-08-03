package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/vagonaizer/authenitfication-service/internal/config"
	"github.com/vagonaizer/authenitfication-service/internal/infrastructure/database/postgres"
	postgresrepos "github.com/vagonaizer/authenitfication-service/internal/infrastructure/database/postgres/repositories"
	"github.com/vagonaizer/authenitfication-service/internal/infrastructure/database/redis"
	"github.com/vagonaizer/authenitfication-service/internal/infrastructure/messaging/kafka"
	"github.com/vagonaizer/authenitfication-service/internal/services"
	grpcserver "github.com/vagonaizer/authenitfication-service/internal/transport/grpc"
	grpchandlers "github.com/vagonaizer/authenitfication-service/internal/transport/grpc/handlers"
	grpcinterceptors "github.com/vagonaizer/authenitfication-service/internal/transport/grpc/interceptors"
	httpserver "github.com/vagonaizer/authenitfication-service/internal/transport/http"
	httphandlers "github.com/vagonaizer/authenitfication-service/internal/transport/http/handlers"
	httpmiddleware "github.com/vagonaizer/authenitfication-service/internal/transport/http/middleware"
	"github.com/vagonaizer/authenitfication-service/pkg/auth"
	"github.com/vagonaizer/authenitfication-service/pkg/logger"
)

type App struct {
	cfg        *config.Config
	logger     *logger.Logger
	db         *postgres.DB
	redis      *redis.Client
	producer   *kafka.Producer
	httpServer *httpserver.Server
	grpcServer *grpcserver.Server
}

func NewApp() (*App, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	log := logger.New(
		cfg.Logger.Level,
		cfg.Logger.Format,
		cfg.Logger.Output,
		cfg.Logger.MaxSize,
		cfg.Logger.MaxBackups,
		cfg.Logger.MaxAge,
		cfg.Logger.Compress,
	)

	// Initialize database
	db, err := postgres.NewConnection(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize Redis
	redisClient, err := redis.NewConnection(&cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	// Initialize Kafka producer
	producer := kafka.NewProducer(&cfg.Kafka, log)

	// Initialize repositories
	userRepo := postgresrepos.NewUserRepository(db)
	sessionRepo := postgresrepos.NewSessionRepository(db)
	roleRepo := postgresrepos.NewRoleRepository(db)

	// Initialize auth utilities
	passwordHasher := auth.NewPasswordHasher()
	jwtManager := auth.NewJWTManager(
		cfg.JWT.AccessTokenSecret,
		cfg.JWT.RefreshTokenSecret,
		cfg.JWT.Issuer,
		cfg.JWT.Audience,
	)

	// Initialize services
	authService := services.NewAuthService(
		userRepo,
		sessionRepo,
		roleRepo,
		passwordHasher,
		jwtManager,
		producer,
		log,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
	)
	userService := services.NewUserService(userRepo, roleRepo, producer, log)

	// Initialize HTTP handlers
	authHandler := httphandlers.NewAuthHandler(authService, log)
	userHandler := httphandlers.NewUserHandler(userService, log)
	healthHandler := httphandlers.NewHealthHandler(db, redisClient, log)
	authMiddleware := httpmiddleware.NewAuthMiddleware(jwtManager, log)

	// Initialize gRPC handlers
	authGRPCHandler := grpchandlers.NewAuthGRPCHandler(authService, log)
	userGRPCHandler := grpchandlers.NewUserGRPCHandler(userService, log)
	authInterceptor := grpcinterceptors.NewAuthInterceptor(jwtManager, log)
	loggingInterceptor := grpcinterceptors.NewLoggingInterceptor(log)

	// Initialize servers
	httpSrv := httpserver.NewServer(
		cfg,
		authHandler,
		userHandler,
		healthHandler,
		authMiddleware,
		log,
	)

	grpcSrv := grpcserver.NewServer(
		authGRPCHandler,
		userGRPCHandler,
		authInterceptor,
		loggingInterceptor,
		log,
	)

	return &App{
		cfg:        cfg,
		logger:     log,
		db:         db,
		redis:      redisClient,
		producer:   producer,
		httpServer: httpSrv,
		grpcServer: grpcSrv,
	}, nil
}

func (a *App) Run() error {
	a.logger.Info("starting application")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start servers
	var wg sync.WaitGroup

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.httpServer.Start(); err != nil {
			a.logger.WithError(err).Error("HTTP server error")
			cancel()
		}
	}()

	// Start gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.grpcServer.Start(":" + a.cfg.Server.GRPCPort); err != nil {
			a.logger.WithError(err).Error("gRPC server error")
			cancel()
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		a.logger.Infof("received signal: %v", sig)
	case <-ctx.Done():
		a.logger.Info("context cancelled")
	}

	// Graceful shutdown
	a.logger.Info("shutting down application")
	return a.shutdown()
}

func (a *App) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.Server.ShutdownTimeout)
	defer cancel()

	var wg sync.WaitGroup
	errChan := make(chan error, 3)

	// Shutdown HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.httpServer.Stop(ctx); err != nil {
			errChan <- fmt.Errorf("HTTP server shutdown error: %w", err)
		}
	}()

	// Shutdown gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.grpcServer.Stop()
	}()

	// Close connections
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.closeConnections(); err != nil {
			errChan <- fmt.Errorf("connections close error: %w", err)
		}
	}()

	// Wait for all shutdowns to complete
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		a.logger.Info("application shutdown completed")
	case <-ctx.Done():
		a.logger.Warn("shutdown timeout exceeded")
	}

	// Check for errors
	close(errChan)
	for err := range errChan {
		a.logger.WithError(err).Error("shutdown error")
	}

	return nil
}

func (a *App) closeConnections() error {
	var errors []error

	// Close Kafka producer
	if a.producer != nil {
		if err := a.producer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("kafka producer close error: %w", err))
		}
	}

	// Close Redis connection
	if a.redis != nil {
		if err := a.redis.Close(); err != nil {
			errors = append(errors, fmt.Errorf("redis close error: %w", err))
		}
	}

	// Close database connection
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			errors = append(errors, fmt.Errorf("database close error: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("multiple close errors: %v", errors)
	}

	return nil
}
