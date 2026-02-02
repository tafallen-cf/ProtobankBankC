package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Common error types
var (
	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenInvalid       = errors.New("invalid token")
	ErrUnauthorized       = errors.New("unauthorized")

	// User errors
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserInactive      = errors.New("user account is inactive")

	// Validation errors
	ErrInvalidInput      = errors.New("invalid input")
	ErrInvalidEmail      = errors.New("invalid email format")
	ErrWeakPassword      = errors.New("password does not meet strength requirements")
	ErrPasswordMismatch  = errors.New("passwords do not match")

	// Rate limiting
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// Internal errors
	ErrInternal         = errors.New("internal server error")
	ErrDatabaseError    = errors.New("database error")
	ErrCacheError       = errors.New("cache error")
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Err        error
	Message    string
	StatusCode int
	Internal   error // Internal error for logging (not exposed to client)
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "unknown error"
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(err error, message string, statusCode int) *AppError {
	return &AppError{
		Err:        err,
		Message:    message,
		StatusCode: statusCode,
	}
}

// NewBadRequest creates a 400 Bad Request error
func NewBadRequest(message string) *AppError {
	return &AppError{
		Err:        ErrInvalidInput,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// NewUnauthorized creates a 401 Unauthorized error
func NewUnauthorized(message string) *AppError {
	return &AppError{
		Err:        ErrUnauthorized,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// NewNotFound creates a 404 Not Found error
func NewNotFound(message string) *AppError {
	return &AppError{
		Err:        ErrUserNotFound,
		Message:    message,
		StatusCode: http.StatusNotFound,
	}
}

// NewForbidden creates a 403 Forbidden error
func NewForbidden(message string) *AppError {
	return &AppError{
		Err:        ErrUserInactive,
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// NewConflict creates a 409 Conflict error
func NewConflict(message string) *AppError {
	return &AppError{
		Err:        ErrUserAlreadyExists,
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

// NewTooManyRequests creates a 429 Too Many Requests error
func NewTooManyRequests(message string) *AppError {
	return &AppError{
		Err:        ErrRateLimitExceeded,
		Message:    message,
		StatusCode: http.StatusTooManyRequests,
	}
}

// NewInternalError creates a 500 Internal Server Error
func NewInternalError(err error, message string) *AppError {
	return &AppError{
		Err:        ErrInternal,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Internal:   err,
	}
}

// WrapError wraps an error with additional context
func WrapError(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError extracts AppError from an error chain
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

// GetStatusCode returns HTTP status code for an error
func GetStatusCode(err error) int {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}
