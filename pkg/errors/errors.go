package errors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    map[string]string `json:"details,omitempty"`
	StatusCode int               `json:"-"`
	Err        error             `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

func Wrap(err error, code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

func WithDetails(err *AppError, details map[string]string) *AppError {
	err.Details = details
	return err
}

func Internal(message string) *AppError {
	return New(CodeInternal, message, http.StatusInternalServerError)
}

func InternalWrap(err error, message string) *AppError {
	return Wrap(err, CodeInternal, message, http.StatusInternalServerError)
}

func Validation(message string) *AppError {
	return New(CodeValidation, message, http.StatusBadRequest)
}

func ValidationWrap(err error, message string) *AppError {
	return Wrap(err, CodeValidation, message, http.StatusBadRequest)
}

func NotFound(message string) *AppError {
	return New(CodeNotFound, message, http.StatusNotFound)
}

func AlreadyExists(message string) *AppError {
	return New(CodeAlreadyExists, message, http.StatusConflict)
}

func Unauthorized(message string) *AppError {
	return New(CodeUnauthorized, message, http.StatusUnauthorized)
}

func Forbidden(message string) *AppError {
	return New(CodeForbidden, message, http.StatusForbidden)
}

func InvalidCredentials() *AppError {
	return New(CodeInvalidCredentials, "Invalid email or password", http.StatusUnauthorized)
}

func TokenExpired() *AppError {
	return New(CodeTokenExpired, "Token has expired", http.StatusUnauthorized)
}

func TokenInvalid() *AppError {
	return New(CodeTokenInvalid, "Invalid token", http.StatusUnauthorized)
}

func UserNotFound() *AppError {
	return New(CodeUserNotFound, "User not found", http.StatusNotFound)
}

func UserInactive() *AppError {
	return New(CodeUserInactive, "User account is inactive", http.StatusForbidden)
}

func UserNotVerified() *AppError {
	return New(CodeUserNotVerified, "User account is not verified", http.StatusForbidden)
}

func EmailExists() *AppError {
	return New(CodeEmailExists, "Email already exists", http.StatusConflict)
}

func UsernameExists() *AppError {
	return New(CodeUsernameExists, "Username already exists", http.StatusConflict)
}

func WeakPassword() *AppError {
	return New(CodeWeakPassword, "Password does not meet security requirements", http.StatusBadRequest)
}

func RateLimitExceeded() *AppError {
	return New(CodeRateLimitExceeded, "Rate limit exceeded", http.StatusTooManyRequests)
}

func DatabaseError(err error) *AppError {
	return Wrap(err, CodeDatabaseError, "Database operation failed", http.StatusInternalServerError)
}

func CacheError(err error) *AppError {
	return Wrap(err, CodeCacheError, "Cache operation failed", http.StatusInternalServerError)
}

func ExternalServiceError(err error, service string) *AppError {
	return Wrap(err, CodeExternalService, fmt.Sprintf("External service %s error", service), http.StatusServiceUnavailable)
}
