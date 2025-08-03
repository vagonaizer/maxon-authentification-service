package errors

const (
	CodeInternal           = "INTERNAL_ERROR"
	CodeValidation         = "VALIDATION_ERROR"
	CodeNotFound           = "NOT_FOUND"
	CodeAlreadyExists      = "ALREADY_EXISTS"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeForbidden          = "FORBIDDEN"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
	CodeTokenExpired       = "TOKEN_EXPIRED"
	CodeTokenInvalid       = "TOKEN_INVALID"
	CodeUserNotFound       = "USER_NOT_FOUND"
	CodeUserInactive       = "USER_INACTIVE"
	CodeUserNotVerified    = "USER_NOT_VERIFIED"
	CodeEmailExists        = "EMAIL_EXISTS"
	CodeUsernameExists     = "USERNAME_EXISTS"
	CodeWeakPassword       = "WEAK_PASSWORD"
	CodeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	CodeDatabaseError      = "DATABASE_ERROR"
	CodeCacheError         = "CACHE_ERROR"
	CodeExternalService    = "EXTERNAL_SERVICE_ERROR"
)
