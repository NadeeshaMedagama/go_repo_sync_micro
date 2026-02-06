package errors

import (
	"fmt"
)

// Error types for better error handling
type ErrorType string

const (
	ErrTypeValidation   ErrorType = "VALIDATION_ERROR"
	ErrTypeNotFound     ErrorType = "NOT_FOUND"
	ErrTypeUnauthorized ErrorType = "UNAUTHORIZED"
	ErrTypeRateLimit    ErrorType = "RATE_LIMIT"
	ErrTypeNetwork      ErrorType = "NETWORK_ERROR"
	ErrTypeExternal     ErrorType = "EXTERNAL_SERVICE_ERROR"
	ErrTypeInternal     ErrorType = "INTERNAL_ERROR"
	ErrTypeDatabase     ErrorType = "DATABASE_ERROR"
)

// AppError represents a custom application error
type AppError struct {
	Type    ErrorType
	Message string
	Err     error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(errType ErrorType, message string, err error) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}

// Validation creates a validation error
func Validation(message string) *AppError {
	return &AppError{
		Type:    ErrTypeValidation,
		Message: message,
	}
}

// NotFound creates a not found error
func NotFound(resource string) *AppError {
	return &AppError{
		Type:    ErrTypeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
	}
}

// Unauthorized creates an unauthorized error
func Unauthorized(message string) *AppError {
	return &AppError{
		Type:    ErrTypeUnauthorized,
		Message: message,
	}
}

// RateLimit creates a rate limit error
func RateLimit(message string) *AppError {
	return &AppError{
		Type:    ErrTypeRateLimit,
		Message: message,
	}
}

// Network creates a network error
func Network(message string, err error) *AppError {
	return &AppError{
		Type:    ErrTypeNetwork,
		Message: message,
		Err:     err,
	}
}

// External creates an external service error
func External(service, message string, err error) *AppError {
	return &AppError{
		Type:    ErrTypeExternal,
		Message: fmt.Sprintf("%s: %s", service, message),
		Err:     err,
	}
}

// Internal creates an internal error
func Internal(message string, err error) *AppError {
	return &AppError{
		Type:    ErrTypeInternal,
		Message: message,
		Err:     err,
	}
}

// Database creates a database error
func Database(message string, err error) *AppError {
	return &AppError{
		Type:    ErrTypeDatabase,
		Message: message,
		Err:     err,
	}
}
