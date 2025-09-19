package errors

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

// ErrorCode represents specific error codes for the application
type ErrorCode string

const (
	// System errors
	ErrCodeInternal        ErrorCode = "INTERNAL_ERROR"
	ErrCodeNotFound        ErrorCode = "NOT_FOUND"
	ErrCodeValidation      ErrorCode = "VALIDATION_ERROR"
	ErrCodeAuthentication  ErrorCode = "AUTHENTICATION_ERROR"
	ErrCodeAuthorization   ErrorCode = "AUTHORIZATION_ERROR"
	ErrCodeRateLimit       ErrorCode = "RATE_LIMIT_ERROR"
	ErrCodeTimeout         ErrorCode = "TIMEOUT_ERROR"

	// Database errors
	ErrCodeDatabase        ErrorCode = "DATABASE_ERROR"
	ErrCodeDatabaseTimeout ErrorCode = "DATABASE_TIMEOUT"
	ErrCodeDuplicateKey    ErrorCode = "DUPLICATE_KEY_ERROR"

	// Queue errors
	ErrCodeQueue           ErrorCode = "QUEUE_ERROR"
	ErrCodeQueueTimeout    ErrorCode = "QUEUE_TIMEOUT"
	ErrCodeQueueConnection ErrorCode = "QUEUE_CONNECTION_ERROR"

	// Claude API errors
	ErrCodeClaude          ErrorCode = "CLAUDE_API_ERROR"
	ErrCodeClaudeTimeout   ErrorCode = "CLAUDE_TIMEOUT"
	ErrCodeClaudeRateLimit ErrorCode = "CLAUDE_RATE_LIMIT"

	// Business logic errors
	ErrCodeInvalidRequest  ErrorCode = "INVALID_REQUEST"
	ErrCodeRequestTimeout  ErrorCode = "REQUEST_TIMEOUT"
	ErrCodeSessionExpired  ErrorCode = "SESSION_EXPIRED"
)

// AppError represents an application error with context
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	StatusCode int                    `json:"-"`
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id,omitempty"`
	Component  string                 `json:"component,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Cause      error                  `json:"-"`
	Stack      string                 `json:"stack,omitempty"`
	Retryable  bool                   `json:"retryable"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithRequestID adds request ID to the error
func (e *AppError) WithRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

// WithComponent adds component information to the error
func (e *AppError) WithComponent(component string) *AppError {
	e.Component = component
	return e
}

// WithMetadata adds metadata to the error
func (e *AppError) WithMetadata(key string, value interface{}) *AppError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// WithCause sets the underlying cause
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// New creates a new application error
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Stack:     getStack(),
		Retryable: isRetryable(code),
	}
}

// NewWithDetails creates a new application error with details
func NewWithDetails(code ErrorCode, message, details string) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
		Stack:     getStack(),
		Retryable: isRetryable(code),
	}
}

// Wrap wraps an existing error with application error context
func Wrap(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}

	// If it's already an AppError, wrap it
	var appErr *AppError
	if errors.As(err, &appErr) {
		return &AppError{
			Code:      code,
			Message:   message,
			Timestamp: time.Now(),
			Cause:     appErr,
			Stack:     getStack(),
			Retryable: isRetryable(code),
		}
	}

	return &AppError{
		Code:      code,
		Message:   message,
		Details:   err.Error(),
		Timestamp: time.Now(),
		Cause:     err,
		Stack:     getStack(),
		Retryable: isRetryable(code),
	}
}

// getStack captures the current stack trace
func getStack() string {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return string(buf[:n])
		}
		buf = make([]byte, 2*len(buf))
	}
}

// isRetryable determines if an error is retryable
func isRetryable(code ErrorCode) bool {
	switch code {
	case ErrCodeTimeout, ErrCodeDatabaseTimeout, ErrCodeQueueTimeout, 
		 ErrCodeClaudeTimeout, ErrCodeQueueConnection, ErrCodeRateLimit,
		 ErrCodeClaudeRateLimit:
		return true
	default:
		return false
	}
}

// HTTPStatusCode returns the appropriate HTTP status code for the error
func (e *AppError) HTTPStatusCode() int {
	if e.StatusCode != 0 {
		return e.StatusCode
	}

	switch e.Code {
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeValidation, ErrCodeInvalidRequest:
		return http.StatusBadRequest
	case ErrCodeAuthentication:
		return http.StatusUnauthorized
	case ErrCodeAuthorization:
		return http.StatusForbidden
	case ErrCodeRateLimit, ErrCodeClaudeRateLimit:
		return http.StatusTooManyRequests
	case ErrCodeTimeout, ErrCodeDatabaseTimeout, ErrCodeQueueTimeout, 
		 ErrCodeClaudeTimeout, ErrCodeRequestTimeout:
		return http.StatusRequestTimeout
	case ErrCodeSessionExpired:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

// Predefined error constructors for common cases

// InternalError creates an internal server error
func InternalError(message string) *AppError {
	return New(ErrCodeInternal, message)
}

// NotFoundError creates a not found error
func NotFoundError(resource string) *AppError {
	return New(ErrCodeNotFound, fmt.Sprintf("%s not found", resource))
}

// ValidationError creates a validation error
func ValidationError(field, message string) *AppError {
	return NewWithDetails(ErrCodeValidation, "Validation failed", fmt.Sprintf("%s: %s", field, message))
}

// DatabaseError creates a database error
func DatabaseError(operation string, err error) *AppError {
	return Wrap(err, ErrCodeDatabase, fmt.Sprintf("Database operation failed: %s", operation))
}

// QueueError creates a queue error
func QueueError(operation string, err error) *AppError {
	return Wrap(err, ErrCodeQueue, fmt.Sprintf("Queue operation failed: %s", operation))
}

// ClaudeError creates a Claude API error
func ClaudeError(operation string, err error) *AppError {
	return Wrap(err, ErrCodeClaude, fmt.Sprintf("Claude API operation failed: %s", operation))
}

// TimeoutError creates a timeout error
func TimeoutError(operation string, duration time.Duration) *AppError {
	return NewWithDetails(ErrCodeTimeout, 
		fmt.Sprintf("Operation timed out: %s", operation),
		fmt.Sprintf("Timeout after %v", duration))
}

// AuthenticationError creates an authentication error
func AuthenticationError(message string) *AppError {
	return New(ErrCodeAuthentication, message)
}

// AuthorizationError creates an authorization error
func AuthorizationError(resource string) *AppError {
	return New(ErrCodeAuthorization, fmt.Sprintf("Access denied to %s", resource))
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Retryable
	}
	return false
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) ErrorCode {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ErrCodeInternal
}