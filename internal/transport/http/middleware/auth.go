package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/vagonaizer/authenitfication-service/internal/dto/response"
	"github.com/vagonaizer/authenitfication-service/pkg/auth"
	"github.com/vagonaizer/authenitfication-service/pkg/logger"
)

type AuthMiddleware struct {
	jwtManager *auth.JWTManager
	logger     *logger.Logger
}

func NewAuthMiddleware(jwtManager *auth.JWTManager, logger *logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		logger:     logger,
	}
}

func (m *AuthMiddleware) RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, response.ErrorResponse{
					Error:   "MISSING_TOKEN",
					Message: "Authorization header is required",
					Code:    http.StatusUnauthorized,
				})
			}

			token, err := m.jwtManager.ExtractTokenFromHeader(authHeader)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, response.ErrorResponse{
					Error:   "INVALID_TOKEN_FORMAT",
					Message: "Invalid authorization header format",
					Code:    http.StatusUnauthorized,
				})
			}

			claims, err := m.jwtManager.ValidateAccessToken(token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, response.ErrorResponse{
					Error:   "INVALID_TOKEN",
					Message: "Invalid or expired token",
					Code:    http.StatusUnauthorized,
				})
			}

			c.Set("user_id", claims.UserID.String())
			c.Set("email", claims.Email)
			c.Set("username", claims.Username)
			c.Set("roles", claims.Roles)

			return next(c)
		}
	}
}

func (m *AuthMiddleware) RequireRole(requiredRole string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roles, ok := c.Get("roles").([]string)
			if !ok {
				return c.JSON(http.StatusForbidden, response.ErrorResponse{
					Error:   "INSUFFICIENT_PERMISSIONS",
					Message: "Insufficient permissions",
					Code:    http.StatusForbidden,
				})
			}

			hasRole := false
			for _, role := range roles {
				if role == requiredRole {
					hasRole = true
					break
				}
			}

			if !hasRole {
				return c.JSON(http.StatusForbidden, response.ErrorResponse{
					Error:   "INSUFFICIENT_PERMISSIONS",
					Message: "Insufficient permissions",
					Code:    http.StatusForbidden,
				})
			}

			return next(c)
		}
	}
}

func (m *AuthMiddleware) RequireAnyRole(requiredRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roles, ok := c.Get("roles").([]string)
			if !ok {
				return c.JSON(http.StatusForbidden, response.ErrorResponse{
					Error:   "INSUFFICIENT_PERMISSIONS",
					Message: "Insufficient permissions",
					Code:    http.StatusForbidden,
				})
			}

			hasRole := false
			for _, userRole := range roles {
				for _, requiredRole := range requiredRoles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				return c.JSON(http.StatusForbidden, response.ErrorResponse{
					Error:   "INSUFFICIENT_PERMISSIONS",
					Message: "Insufficient permissions",
					Code:    http.StatusForbidden,
				})
			}

			return next(c)
		}
	}
}

func (m *AuthMiddleware) OptionalAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return next(c)
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				return next(c)
			}

			token := authHeader[7:]
			claims, err := m.jwtManager.ValidateAccessToken(token)
			if err != nil {
				return next(c)
			}

			c.Set("user_id", claims.UserID.String())
			c.Set("email", claims.Email)
			c.Set("username", claims.Username)
			c.Set("roles", claims.Roles)

			return next(c)
		}
	}
}
