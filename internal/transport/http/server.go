package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/vagonaizer/authenitfication-service/internal/config"
	"github.com/vagonaizer/authenitfication-service/internal/transport/http/handlers"
	"github.com/vagonaizer/authenitfication-service/internal/transport/http/middleware"
	"github.com/vagonaizer/authenitfication-service/internal/transport/http/routes"
	"github.com/vagonaizer/authenitfication-service/pkg/logger"
)

type Server struct {
	echo          *echo.Echo
	server        *http.Server
	logger        *logger.Logger
	authHandler   *handlers.AuthHandler
	userHandler   *handlers.UserHandler
	healthHandler *handlers.HealthHandler
	authMW        *middleware.AuthMiddleware
}

func NewServer(
	cfg *config.Config,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	healthHandler *handlers.HealthHandler,
	authMW *middleware.AuthMiddleware,
	log *logger.Logger,
) *Server {
	e := echo.New()

	// Hide Echo banner
	e.HideBanner = true

	// Basic middleware
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.RequestID())

	// CORS middleware
	if cfg.Server.EnableCORS {
		e.Use(middleware.CORS())
	}

	// Rate limiting
	if cfg.Server.EnableRateLimit {
		e.Use(middleware.RateLimit(cfg.Server.RateLimitRPS))
	}

	// Logging middleware
	e.Use(middleware.Logging(log))

	// Request size limit
	e.Use(echomiddleware.BodyLimit(fmt.Sprintf("%d", cfg.Server.MaxRequestSize)))

	// Setup routes
	routes.SetupRoutes(e, authHandler, userHandler, healthHandler, authMW)

	server := &http.Server{
		Addr:         ":" + cfg.Server.HTTPPort,
		Handler:      e,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	return &Server{
		echo:          e,
		server:        server,
		logger:        log,
		authHandler:   authHandler,
		userHandler:   userHandler,
		healthHandler: healthHandler,
		authMW:        authMW,
	}
}

func (s *Server) Start() error {
	s.logger.Infof("HTTP server starting on %s", s.server.Addr)

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server")

	return s.server.Shutdown(ctx)
}

func (s *Server) Handler() http.Handler {
	return s.echo
}
