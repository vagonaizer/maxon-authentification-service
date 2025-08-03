package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/vagonaizer/authenitfication-service/internal/config"
	"github.com/vagonaizer/authenitfication-service/internal/infrastructure/database/postgres"
	"github.com/vagonaizer/authenitfication-service/internal/infrastructure/database/redis"
	"github.com/vagonaizer/authenitfication-service/internal/infrastructure/messaging/kafka"
	grpcserver "github.com/vagonaizer/authenitfication-service/internal/transport/grpc"
	httpserver "github.com/vagonaizer/authenitfication-service/internal/transport/http"
	"github.com/vagonaizer/authenitfication-service/pkg/logger"
)

// Type aliases for Wire
type AccessTokenExpiry time.Duration
type RefreshTokenExpiry time.Duration

type App struct {
	cfg        *config.Config
	logger     *logger.Logger
	db         *postgres.DB
	redis      *redis.Client
	producer   *kafka.Producer
	httpServer *httpserver.Server
	grpcServer *grpcserver.Server
}

func NewApp(
	cfg *config.Config,
	logger *logger.Logger,
	db *postgres.DB,
	redis *redis.Client,
	producer *kafka.Producer,
	httpServer *httpserver.Server,
	grpcServer *grpcserver.Server,
) *App {
	return &App{
		cfg:        cfg,
		logger:     logger,
		db:         db,
		redis:      redis,
		producer:   producer,
		httpServer: httpServer,
		grpcServer: grpcServer,
	}
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
