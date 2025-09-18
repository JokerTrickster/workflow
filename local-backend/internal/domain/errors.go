package domain

import (
	"errors"
	"fmt"
)

// Domain-specific errors
var (
	// ErrRequestNotFound indicates a request was not found
	ErrRequestNotFound = errors.New("request not found")
	
	// ErrContextNotFound indicates a processing context was not found
	ErrContextNotFound = errors.New("processing context not found")
	
	// ErrInvalidMessage indicates a message is invalid or malformed
	ErrInvalidMessage = errors.New("invalid message")
	
	// ErrRequestAlreadyProcessed indicates a request has already been processed
	ErrRequestAlreadyProcessed = errors.New("request already processed")
	
	// ErrRequestCannotBeCancelled indicates a request cannot be cancelled
	ErrRequestCannotBeCancelled = errors.New("request cannot be cancelled")
	
	// ErrClaudeAPIError indicates an error with the Claude API
	ErrClaudeAPIError = errors.New("claude api error")
	
	// ErrDatabaseError indicates a database operation error
	ErrDatabaseError = errors.New("database error")
	
	// ErrQueueError indicates a message queue error
	ErrQueueError = errors.New("queue error")
)

// BusinessError represents a domain business logic error
type BusinessError struct {
	Code    string
	Message string
	Cause   error
}

// Error implements the error interface
func (e *BusinessError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *BusinessError) Unwrap() error {
	return e.Cause
}

// NewBusinessError creates a new business error
func NewBusinessError(code, message string, cause error) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Error codes for business logic errors
const (
	ErrorCodeInvalidRequest     = "INVALID_REQUEST"
	ErrorCodeProcessingFailed   = "PROCESSING_FAILED"
	ErrorCodeContextExpired     = "CONTEXT_EXPIRED"
	ErrorCodeResourceNotFound   = "RESOURCE_NOT_FOUND"
	ErrorCodeUnauthorized       = "UNAUTHORIZED"
	ErrorCodeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	ErrorCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// Validation errors
func NewInvalidRequestError(message string) *BusinessError {
	return NewBusinessError(ErrorCodeInvalidRequest, message, ErrInvalidMessage)
}

func NewProcessingFailedError(message string, cause error) *BusinessError {
	return NewBusinessError(ErrorCodeProcessingFailed, message, cause)
}

func NewResourceNotFoundError(resource, id string) *BusinessError {
	return NewBusinessError(ErrorCodeResourceNotFound, 
		fmt.Sprintf("%s with id '%s' not found", resource, id), nil)
}

func NewServiceUnavailableError(service string, cause error) *BusinessError {
	return NewBusinessError(ErrorCodeServiceUnavailable,
		fmt.Sprintf("%s service is unavailable", service), cause)
}