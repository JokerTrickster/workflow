package errors

import (
	"fmt"

	domainErrors "ai-git-workbench/internal/domain/errors"
)

// ApplicationError represents an application layer error
type ApplicationError struct {
	Code       string
	Message    string
	Cause      error
	HTTPStatus int
}

func (e ApplicationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e ApplicationError) Unwrap() error {
	return e.Cause
}

// Application error codes
const (
	ErrCodeUnauthorized         = "UNAUTHORIZED"
	ErrCodeValidationFailed     = "VALIDATION_FAILED"
	ErrCodeOperationFailed      = "OPERATION_FAILED"
	ErrCodeResourceNotFound     = "RESOURCE_NOT_FOUND"
	ErrCodeConflict             = "CONFLICT"
	ErrCodeInternalServerError  = "INTERNAL_SERVER_ERROR"
	ErrCodeBadRequest           = "BAD_REQUEST"
	ErrCodeForbidden            = "FORBIDDEN"
	ErrCodeServiceUnavailable   = "SERVICE_UNAVAILABLE"
)

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) ApplicationError {
	return ApplicationError{
		Code:       ErrCodeUnauthorized,
		Message:    message,
		HTTPStatus: 401,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string, cause error) ApplicationError {
	return ApplicationError{
		Code:       ErrCodeValidationFailed,
		Message:    message,
		Cause:      cause,
		HTTPStatus: 400,
	}
}

// NewResourceNotFoundError creates a resource not found error
func NewResourceNotFoundError(resourceType, resourceID string) ApplicationError {
	return ApplicationError{
		Code:       ErrCodeResourceNotFound,
		Message:    fmt.Sprintf("%s with ID '%s' not found", resourceType, resourceID),
		HTTPStatus: 404,
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string, cause error) ApplicationError {
	return ApplicationError{
		Code:       ErrCodeConflict,
		Message:    message,
		Cause:      cause,
		HTTPStatus: 409,
	}
}

// NewOperationFailedError creates an operation failed error
func NewOperationFailedError(operation string, cause error) ApplicationError {
	return ApplicationError{
		Code:       ErrCodeOperationFailed,
		Message:    fmt.Sprintf("failed to %s", operation),
		Cause:      cause,
		HTTPStatus: 500,
	}
}

// NewInternalServerError creates an internal server error
func NewInternalServerError(message string, cause error) ApplicationError {
	return ApplicationError{
		Code:       ErrCodeInternalServerError,
		Message:    message,
		Cause:      cause,
		HTTPStatus: 500,
	}
}

// NewBadRequestError creates a bad request error
func NewBadRequestError(message string, cause error) ApplicationError {
	return ApplicationError{
		Code:       ErrCodeBadRequest,
		Message:    message,
		Cause:      cause,
		HTTPStatus: 400,
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) ApplicationError {
	return ApplicationError{
		Code:       ErrCodeForbidden,
		Message:    message,
		HTTPStatus: 403,
	}
}

// NewServiceUnavailableError creates a service unavailable error
func NewServiceUnavailableError(message string, cause error) ApplicationError {
	return ApplicationError{
		Code:       ErrCodeServiceUnavailable,
		Message:    message,
		Cause:      cause,
		HTTPStatus: 503,
	}
}

// TranslateDomainError translates domain errors to application errors
func TranslateDomainError(err error) ApplicationError {
	if err == nil {
		return ApplicationError{}
	}

	// Handle domain-specific errors
	if domainErr, ok := err.(domainErrors.TaskError); ok {
		switch domainErr.Code {
		case domainErrors.ErrCodeTaskNotFound:
			return NewResourceNotFoundError("Task", extractTaskIDFromMessage(domainErr.Message))
		case domainErrors.ErrCodeTaskAlreadyExists:
			return NewConflictError(domainErr.Message, domainErr.Cause)
		case domainErrors.ErrCodeInvalidTaskTitle, domainErrors.ErrCodeInvalidTaskDescription:
			return NewValidationError(domainErr.Message, domainErr.Cause)
		case domainErrors.ErrCodeInvalidStatusTransition:
			return NewBadRequestError(domainErr.Message, domainErr.Cause)
		case domainErrors.ErrCodeTaskValidationFailed:
			return NewValidationError(domainErr.Message, domainErr.Cause)
		case domainErrors.ErrCodeTaskAlreadyCompleted:
			return NewConflictError(domainErr.Message, domainErr.Cause)
		case domainErrors.ErrCodeTaskNotActive:
			return NewBadRequestError(domainErr.Message, domainErr.Cause)
		case domainErrors.ErrCodeConcurrentModification:
			return NewConflictError(domainErr.Message, domainErr.Cause)
		default:
			return NewInternalServerError("Unknown domain error", err)
		}
	}

	// Default to internal server error for unknown errors
	return NewInternalServerError("Internal server error", err)
}

// Helper function to extract task ID from error message
func extractTaskIDFromMessage(message string) string {
	// Simple extraction - in real implementation, you might use regex
	// For now, return a generic identifier
	return "unknown"
}

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	if appErr, ok := err.(ApplicationError); ok {
		return appErr.Code == ErrCodeResourceNotFound
	}
	return domainErrors.IsTaskNotFound(err)
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	if appErr, ok := err.(ApplicationError); ok {
		return appErr.Code == ErrCodeValidationFailed || appErr.Code == ErrCodeBadRequest
	}
	return domainErrors.IsValidationError(err)
}

// IsConflictError checks if the error is a conflict error
func IsConflictError(err error) bool {
	if appErr, ok := err.(ApplicationError); ok {
		return appErr.Code == ErrCodeConflict
	}
	return domainErrors.IsTaskAlreadyExists(err)
}

// IsUnauthorizedError checks if the error is an unauthorized error
func IsUnauthorizedError(err error) bool {
	if appErr, ok := err.(ApplicationError); ok {
		return appErr.Code == ErrCodeUnauthorized
	}
	return false
}

// IsForbiddenError checks if the error is a forbidden error
func IsForbiddenError(err error) bool {
	if appErr, ok := err.(ApplicationError); ok {
		return appErr.Code == ErrCodeForbidden
	}
	return false
}

// GetHTTPStatus returns the HTTP status code for an error
func GetHTTPStatus(err error) int {
	if appErr, ok := err.(ApplicationError); ok {
		return appErr.HTTPStatus
	}
	
	// Map domain errors to HTTP status codes
	if domainErrors.IsTaskNotFound(err) {
		return 404
	}
	if domainErrors.IsValidationError(err) {
		return 400
	}
	if domainErrors.IsTaskAlreadyExists(err) {
		return 409
	}
	
	// Default to 500
	return 500
}