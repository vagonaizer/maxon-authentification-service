package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/vagonaizer/authenitfication-service/internal/domain/services"
	"github.com/vagonaizer/authenitfication-service/internal/dto/request"
	"github.com/vagonaizer/authenitfication-service/internal/dto/response"
	"github.com/vagonaizer/authenitfication-service/pkg/errors"
	"github.com/vagonaizer/authenitfication-service/pkg/logger"
)

type AuthHandler struct {
	authService services.AuthService
	logger      *logger.Logger
}

func NewAuthHandler(authService services.AuthService, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req request.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request format",
			Code:    http.StatusBadRequest,
		})
	}

	if err := request.ValidateStruct(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}

	// Получаем IP адрес и User Agent из запроса
	ipAddress := c.RealIP()
	if ipAddress == "" {
		ipAddress = "127.0.0.1"
	}
	userAgent := c.Request().UserAgent()

	result, err := h.authService.Register(c.Request().Context(), &req, ipAddress, userAgent)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.JSON(appErr.StatusCode, response.ErrorResponse{
				Error:   appErr.Code,
				Message: appErr.Message,
				Code:    appErr.StatusCode,
				Details: appErr.Details,
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Internal server error",
			Code:    http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusCreated, result)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req request.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request format",
			Code:    http.StatusBadRequest,
		})
	}

	if err := request.ValidateStruct(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}

	// Получаем IP адрес и User Agent из запроса
	ipAddress := c.RealIP()
	if ipAddress == "" {
		ipAddress = "127.0.0.1"
	}
	userAgent := c.Request().UserAgent()

	result, err := h.authService.Login(c.Request().Context(), &req, ipAddress, userAgent)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.JSON(appErr.StatusCode, response.ErrorResponse{
				Error:   appErr.Code,
				Message: appErr.Message,
				Code:    appErr.StatusCode,
				Details: appErr.Details,
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Internal server error",
			Code:    http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req request.RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request format",
			Code:    http.StatusBadRequest,
		})
	}

	if err := request.ValidateStruct(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}

	result, err := h.authService.RefreshToken(c.Request().Context(), &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.JSON(appErr.StatusCode, response.ErrorResponse{
				Error:   appErr.Code,
				Message: appErr.Message,
				Code:    appErr.StatusCode,
				Details: appErr.Details,
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Internal server error",
			Code:    http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) Logout(c echo.Context) error {
	var req request.LogoutRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request format",
			Code:    http.StatusBadRequest,
		})
	}

	if err := request.ValidateStruct(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}

	err := h.authService.Logout(c.Request().Context(), &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.JSON(appErr.StatusCode, response.ErrorResponse{
				Error:   appErr.Code,
				Message: appErr.Message,
				Code:    appErr.StatusCode,
				Details: appErr.Details,
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Internal server error",
			Code:    http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusOK, response.SuccessResponse{
		Message: "Logged out successfully",
	})
}

func (h *AuthHandler) VerifyToken(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "MISSING_TOKEN",
			Message: "Authorization header is required",
			Code:    http.StatusUnauthorized,
		})
	}

	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	result, err := h.authService.VerifyToken(c.Request().Context(), token)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.JSON(appErr.StatusCode, response.ErrorResponse{
				Error:   appErr.Code,
				Message: appErr.Message,
				Code:    appErr.StatusCode,
				Details: appErr.Details,
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Internal server error",
			Code:    http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) ChangePassword(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req request.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request format",
			Code:    http.StatusBadRequest,
		})
	}

	req.UserID = userID

	if err := request.ValidateStruct(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}

	err := h.authService.ChangePassword(c.Request().Context(), &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.JSON(appErr.StatusCode, response.ErrorResponse{
				Error:   appErr.Code,
				Message: appErr.Message,
				Code:    appErr.StatusCode,
				Details: appErr.Details,
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Internal server error",
			Code:    http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusOK, response.SuccessResponse{
		Message: "Password changed successfully",
	})
}
