package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jwtauth "github.com/threatflux/libgo/internal/auth/jwt"
	userauth "github.com/threatflux/libgo/internal/auth/user"
	apierrors "github.com/threatflux/libgo/internal/errors"
	"github.com/threatflux/libgo/pkg/logger"
)

// Error types
var (
	ErrNotFound      = errors.New("resource not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrInternalError = errors.New("internal server error")
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// HandleError handles errors and returns appropriate HTTP responses
func HandleError(c *gin.Context, err error) {
	// Get the logger from context if available
	var log logger.Logger
	if loggerInstance, exists := c.Get("logger"); exists {
		if l, ok := loggerInstance.(logger.Logger); ok {
			log = l
		}
	}

	// If no context logger available, use default
	if log == nil {
		if l, exists := c.Get("defaultLogger"); exists {
			if defaultLogger, ok := l.(logger.Logger); ok {
				log = defaultLogger
			}
		}
	}

	// Fallback to a no-op logger if none available
	if log == nil {
		log = &noopLogger{}
	}

	// Determine status code and error code
	statusCode, errorCode := mapErrorToStatusAndCode(err)

	// Create sanitized message
	message := sanitizeErrorMessage(err, statusCode)

	// Log the error with context
	log.Error("API error",
		logger.String("path", c.Request.URL.Path),
		logger.String("method", c.Request.Method),
		logger.Int("status", statusCode),
		logger.String("code", errorCode),
		logger.Error(err))

	// Send response
	c.JSON(statusCode, ErrorResponse{
		Status:  statusCode,
		Code:    errorCode,
		Message: message,
	})
}

// mapErrorToStatusAndCode maps domain errors to HTTP status codes and error codes
func mapErrorToStatusAndCode(err error) (int, string) {
	// Check each error category
	if status, code := checkNotFoundErrors(err); status != 0 {
		return status, code
	}
	if status, code := checkBadRequestErrors(err); status != 0 {
		return status, code
	}
	if status, code := checkUnauthorizedErrors(err); status != 0 {
		return status, code
	}
	if status, code := checkForbiddenErrors(err); status != 0 {
		return status, code
	}
	if status, code := checkConflictErrors(err); status != 0 {
		return status, code
	}
	if errors.Is(err, ErrInternalError) {
		return http.StatusInternalServerError, "INTERNAL_SERVER_ERROR"
	}

	// Default case
	return http.StatusInternalServerError, "INTERNAL_SERVER_ERROR"
}

// checkNotFoundErrors checks for not found error types.
func checkNotFoundErrors(err error) (int, string) {
	notFoundErrors := []error{
		ErrNotFound,
		apierrors.ErrVMNotFound,
	}
	for _, target := range notFoundErrors {
		if errors.Is(err, target) {
			return http.StatusNotFound, "NOT_FOUND"
		}
	}
	return 0, ""
}

// checkBadRequestErrors checks for bad request error types.
func checkBadRequestErrors(err error) (int, string) {
	badRequestErrors := []error{
		ErrInvalidInput,
		apierrors.ErrInvalidCPUCount,
		apierrors.ErrInvalidMemorySize,
		apierrors.ErrInvalidDiskSize,
		apierrors.ErrInvalidDiskFormat,
		apierrors.ErrInvalidNetworkType,
		apierrors.ErrInvalidNetworkSource,
		apierrors.ErrVMInvalidState,
	}
	for _, target := range badRequestErrors {
		if errors.Is(err, target) {
			return http.StatusBadRequest, "INVALID_INPUT"
		}
	}
	return 0, ""
}

// checkUnauthorizedErrors checks for unauthorized error types.
func checkUnauthorizedErrors(err error) (int, string) {
	unauthorizedErrors := []error{
		ErrUnauthorized,
		apierrors.ErrInvalidCredentials,
		userauth.ErrInvalidCredentials,
		jwtauth.ErrTokenExpired,
		jwtauth.ErrInvalidToken,
	}
	for _, target := range unauthorizedErrors {
		if errors.Is(err, target) {
			return http.StatusUnauthorized, "UNAUTHORIZED"
		}
	}
	return 0, ""
}

// checkForbiddenErrors checks for forbidden error types.
func checkForbiddenErrors(err error) (int, string) {
	forbiddenErrors := []error{
		ErrForbidden,
		apierrors.ErrForbidden,
		apierrors.ErrUserInactive,
		userauth.ErrUserInactive,
	}
	for _, target := range forbiddenErrors {
		if errors.Is(err, target) {
			return http.StatusForbidden, "FORBIDDEN"
		}
	}
	return 0, ""
}

// checkConflictErrors checks for conflict error types.
func checkConflictErrors(err error) (int, string) {
	conflictErrors := []error{
		ErrAlreadyExists,
		apierrors.ErrVMAlreadyExists,
		apierrors.ErrDuplicateUsername,
		userauth.ErrDuplicateUsername,
	}
	for _, target := range conflictErrors {
		if errors.Is(err, target) {
			return http.StatusConflict, "RESOURCE_CONFLICT"
		}
	}
	return 0, ""
}

// sanitizeErrorMessage creates a user-friendly error message
func sanitizeErrorMessage(err error, statusCode int) string {
	// For internal server errors, don't expose details
	if statusCode == http.StatusInternalServerError {
		return "An internal server error occurred"
	}

	// For validation errors, include details
	if statusCode == http.StatusBadRequest && strings.Contains(err.Error(), "validation") {
		return err.Error()
	}

	// For other errors, use the error message but remove sensitive info
	message := err.Error()

	// Remove any potential sensitive information (paths, connections strings, etc.)
	// This is a simple approach and might need enhancement
	message = sanitizeSensitiveInfo(message)

	return message
}

// sanitizeSensitiveInfo removes sensitive information from error messages
func sanitizeSensitiveInfo(message string) string {
	// List of patterns to sanitize
	patterns := []string{
		"password",
		"token",
		"secret",
		"key",
		"connection string",
		"/var/",
		"/usr/",
		"/home/",
		"/tmp/",
		"C:\\",
	}

	// Check if message contains any sensitive pattern
	for _, pattern := range patterns {
		if strings.Contains(strings.ToLower(message), strings.ToLower(pattern)) {
			// If sensitive data detected, return a generic message
			return "An error occurred while processing your request"
		}
	}

	return message
}

// noopLogger provides a no-op implementation of logger.Logger
type noopLogger struct{}

func (l *noopLogger) Debug(msg string, fields ...logger.Field)        {}
func (l *noopLogger) Info(msg string, fields ...logger.Field)         {}
func (l *noopLogger) Warn(msg string, fields ...logger.Field)         {}
func (l *noopLogger) Error(msg string, fields ...logger.Field)        {}
func (l *noopLogger) Fatal(msg string, fields ...logger.Field)        {}
func (l *noopLogger) WithFields(fields ...logger.Field) logger.Logger { return l }
func (l *noopLogger) WithError(err error) logger.Logger               { return l }
func (l *noopLogger) Sync() error                                     { return nil }
