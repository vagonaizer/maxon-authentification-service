package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/vagonaizer/authenitfication-service/internal/transport/http/handlers"
	"github.com/vagonaizer/authenitfication-service/internal/transport/http/middleware"
)

func SetupRoutes(
	e *echo.Echo,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	healthHandler *handlers.HealthHandler,
	authMiddleware *middleware.AuthMiddleware,
) {
	// Health check routes
	e.GET("/health", healthHandler.Health)
	e.GET("/ready", healthHandler.Ready)
	e.GET("/live", healthHandler.Live)

	// API v1 routes
	v1 := e.Group("/api/v1")

	// Auth routes (public)
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", authHandler.Logout)
		auth.GET("/verify", authHandler.VerifyToken)
	}

	// Protected auth routes
	authProtected := v1.Group("/auth", authMiddleware.RequireAuth())
	{
		authProtected.POST("/change-password", authHandler.ChangePassword)
	}

	// User routes (protected)
	users := v1.Group("/users", authMiddleware.RequireAuth())
	{
		users.GET("/profile", userHandler.GetProfile)
		users.PUT("/profile", userHandler.UpdateProfile)
		users.DELETE("/profile", userHandler.DeleteAccount)
		users.GET("/:id", userHandler.GetUserByID)
		users.GET("/:id/roles", userHandler.GetUserRoles)
	}

	// Admin routes (require admin role)
	admin := v1.Group("/admin", authMiddleware.RequireAuth(), authMiddleware.RequireRole("admin"))
	{
		admin.GET("/users", userHandler.ListUsers)
		//admin.POST("/users/:id/activate", userHandler.ActivateUser)
		//admin.POST("/users/:id/deactivate", userHandler.DeactivateUser)
		admin.POST("/users/roles/assign", userHandler.AssignRole)
		admin.DELETE("/users/roles/remove", userHandler.RemoveRole)
	}
}
