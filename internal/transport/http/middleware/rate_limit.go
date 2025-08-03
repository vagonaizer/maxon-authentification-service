package middleware

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vagonaizer/authenitfication-service/internal/dto/response"
	"golang.org/x/time/rate"
)

func RateLimit(rps int) echo.MiddlewareFunc {
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(rps),
				Burst:     rps * 2,
				ExpiresIn: time.Hour,
			},
		),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return c.RealIP(), nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return c.JSON(http.StatusTooManyRequests, response.ErrorResponse{
				Error:   "RATE_LIMIT_EXCEEDED",
				Message: "Too many requests",
				Code:    http.StatusTooManyRequests,
			})
		},
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			return c.JSON(http.StatusTooManyRequests, response.ErrorResponse{
				Error:   "RATE_LIMIT_EXCEEDED",
				Message: "Rate limit exceeded",
				Code:    http.StatusTooManyRequests,
			})
		},
	})
}
