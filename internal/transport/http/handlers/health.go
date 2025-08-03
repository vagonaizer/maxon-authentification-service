package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/vagonaizer/authenitfication-service/internal/dto/response"
	"github.com/vagonaizer/authenitfication-service/internal/infrastructure/database/postgres"
	"github.com/vagonaizer/authenitfication-service/internal/infrastructure/database/redis"
	"github.com/vagonaizer/authenitfication-service/pkg/logger"
)

type HealthHandler struct {
	db     *postgres.DB
	redis  *redis.Client
	logger *logger.Logger
}

func NewHealthHandler(db *postgres.DB, redis *redis.Client, logger *logger.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

func (h *HealthHandler) Health(c echo.Context) error {
	services := make(map[string]string)

	if err := h.db.Health(); err != nil {
		services["database"] = "unhealthy"
		h.logger.WithError(err).Error("database health check failed")
	} else {
		services["database"] = "healthy"
	}

	if err := h.redis.Health(); err != nil {
		services["redis"] = "unhealthy"
		h.logger.WithError(err).Error("redis health check failed")
	} else {
		services["redis"] = "healthy"
	}

	status := "healthy"
	statusCode := http.StatusOK

	for _, serviceStatus := range services {
		if serviceStatus == "unhealthy" {
			status = "unhealthy"
			statusCode = http.StatusServiceUnavailable
			break
		}
	}

	healthResponse := response.HealthResponse{
		Status:    status,
		Timestamp: time.Now().Format(time.RFC3339),
		Services:  services,
	}

	return c.JSON(statusCode, healthResponse)
}

func (h *HealthHandler) Ready(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ready",
	})
}

func (h *HealthHandler) Live(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "alive",
	})
}
