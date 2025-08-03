package handlers

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/vagonaizer/authenitfication-service/internal/domain/services"
	"github.com/vagonaizer/authenitfication-service/internal/dto/request"
	"github.com/vagonaizer/authenitfication-service/internal/dto/response"
	"github.com/vagonaizer/authenitfication-service/pkg/errors"
	"github.com/vagonaizer/authenitfication-service/pkg/logger"
)

type UserHandler struct {
	userService services.UserService
	logger      *logger.Logger
}

func NewUserHandler(userService services.UserService, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

func (h *UserHandler) GetProfile(c echo.Context) error {
	userIDStr := c.Get("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "INVALID_USER_ID",
			Message: "Invalid user ID format",
			Code:    http.StatusBadRequest,
		})
	}

	result, err := h.userService.GetProfile(c.Request().Context(), userID)
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

func (h *UserHandler) UpdateProfile(c echo.Context) error {
	userIDStr := c.Get("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "INVALID_USER_ID",
			Message: "Invalid user ID format",
			Code:    http.StatusBadRequest,
		})
	}

	var req request.UpdateUserRequest
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

	result, err := h.userService.UpdateProfile(c.Request().Context(), &req)
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

func (h *UserHandler) DeleteAccount(c echo.Context) error {
	userIDStr := c.Get("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "INVALID_USER_ID",
			Message: "Invalid user ID format",
			Code:    http.StatusBadRequest,
		})
	}

	err = h.userService.DeleteAccount(c.Request().Context(), userID)
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
		Message: "Account deleted successfully",
	})
}

func (h *UserHandler) ListUsers(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	search := c.QueryParam("search")
	sortBy := c.QueryParam("sort_by")
	sortDir := c.QueryParam("sort_dir")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortDir == "" {
		sortDir = "desc"
	}

	req := &request.ListUsersRequest{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
		SortBy:   sortBy,
		SortDir:  sortDir,
	}

	if err := request.ValidateStruct(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}

	result, err := h.userService.ListUsers(c.Request().Context(), req)
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

func (h *UserHandler) GetUserByID(c echo.Context) error {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "INVALID_USER_ID",
			Message: "Invalid user ID format",
			Code:    http.StatusBadRequest,
		})
	}

	result, err := h.userService.GetUserByID(c.Request().Context(), userID)
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

func (h *UserHandler) AssignRole(c echo.Context) error {
	var req request.AssignRoleRequest
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

	err := h.userService.AssignRole(c.Request().Context(), &req)
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
		Message: "Role assigned successfully",
	})
}

func (h *UserHandler) RemoveRole(c echo.Context) error {
	var req request.RemoveRoleRequest
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

	err := h.userService.RemoveRole(c.Request().Context(), &req)
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
		Message: "Role removed successfully",
	})
}

func (h *UserHandler) GetUserRoles(c echo.Context) error {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "INVALID_USER_ID",
			Message: "Invalid user ID format",
			Code:    http.StatusBadRequest,
		})
	}

	result, err := h.userService.GetUserRoles(c.Request().Context(), userID)
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
