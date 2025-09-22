package errors

import (
	"fmt"
	"net/http"
	"time"
)

// ErrorCode represents specific error types for categorization
type ErrorCode string

const (
	// Database errors
	ErrCodeDatabaseConnection ErrorCode = "DATABASE_CONNECTION"
	ErrCodeDatabaseQuery      ErrorCode = "DATABASE_QUERY"
	ErrCodeDatabaseTimeout    ErrorCode = "DATABASE_TIMEOUT"
	ErrCodeDatabaseConstraint ErrorCode = "DATABASE_CONSTRAINT"
	
	// Queue errors  
	ErrCodeQueueConnection    ErrorCode = "QUEUE_CONNECTION"
	ErrCodeQueuePublish       ErrorCode = "QUEUE_PUBLISH"
	ErrCodeQueueTimeout       ErrorCode = "QUEUE_TIMEOUT"
	
	// Validation errors
	ErrCodeValidationFailed   ErrorCode = "VALIDATION_FAILED"
	ErrCodeInvalidInput       ErrorCode = "INVALID_INPUT"
	ErrCodeMissingField       ErrorCode = "MISSING_FIELD"
	
	// Service errors
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrCodeExternalService    ErrorCode = "EXTERNAL_SERVICE"
	
	// Authorization errors
	ErrCodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden          ErrorCode = "FORBIDDEN"
	ErrCodeInvalidToken       ErrorCode = "INVALID_TOKEN"
	
	// Resource errors
	ErrCodeNotFound           ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists      ErrorCode = "ALREADY_EXISTS"
	ErrCodeResourceConflict   ErrorCode = "RESOURCE_CONFLICT"
)

// ErrorSeverity indicates the severity level of an error
type ErrorSeverity string

const (
	SeverityCritical ErrorSeverity = "CRITICAL"
	SeverityHigh     ErrorSeverity = "HIGH"
	SeverityMedium   ErrorSeverity = "MEDIUM"
	SeverityLow      ErrorSeverity = "LOW"
)

// AppError represents a comprehensive application error
type AppError struct {
	Code        ErrorCode                `json:"code"`
	Message     string                   `json:"message"`
	Details     string                   `json:"details,omitempty"`
	RequestID   string                   `json:"request_id,omitempty"`
	Timestamp   time.Time                `json:"timestamp"`
	Severity    ErrorSeverity            `json:"severity"`
	Context     map[string]interface{}   `json:"context,omitempty"`
	Cause       error                    `json:"-"`
	HTTPStatus  int                      `json:"-"`
	Retryable   bool                     `json:"retryable"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause for error unwrapping
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithRequestID adds request ID to the error
func (e *AppError) WithRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

// WithDetails adds additional details to the error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string, cause error) *AppError {
	severity, httpStatus, retryable := getErrorProperties(code)
	
	return &AppError{
		Code:       code,
		Message:    message,
		Timestamp:  time.Now().UTC(),
		Severity:   severity,
		Cause:      cause,
		HTTPStatus: httpStatus,
		Retryable:  retryable,
		Context:    make(map[string]interface{}),
	}
}

// getErrorProperties returns default properties for error codes
func getErrorProperties(code ErrorCode) (ErrorSeverity, int, bool) {
	switch code {
	case ErrCodeDatabaseConnection, ErrCodeQueueConnection:
		return SeverityCritical, http.StatusServiceUnavailable, true
		
	case ErrCodeDatabaseTimeout, ErrCodeQueueTimeout:
		return SeverityHigh, http.StatusRequestTimeout, true
		
	case ErrCodeDatabaseQuery, ErrCodeQueuePublish:
		return SeverityMedium, http.StatusInternalServerError, false
		
	case ErrCodeValidationFailed, ErrCodeInvalidInput, ErrCodeMissingField:
		return SeverityLow, http.StatusBadRequest, false
		
	case ErrCodeUnauthorized, ErrCodeInvalidToken:
		return SeverityMedium, http.StatusUnauthorized, false
		
	case ErrCodeForbidden:
		return SeverityMedium, http.StatusForbidden, false
		
	case ErrCodeNotFound:
		return SeverityLow, http.StatusNotFound, false
		
	case ErrCodeAlreadyExists, ErrCodeResourceConflict:
		return SeverityLow, http.StatusConflict, false
		
	case ErrCodeServiceUnavailable, ErrCodeExternalService:
		return SeverityHigh, http.StatusServiceUnavailable, true
		
	case ErrCodeDatabaseConstraint:
		return SeverityMedium, http.StatusConflict, false
		
	default:
		return SeverityMedium, http.StatusInternalServerError, false
	}
}

// Database error constructors
func NewDatabaseConnectionError(err error) *AppError {
	return NewAppError(ErrCodeDatabaseConnection, "Database connection failed", err)
}

func NewDatabaseQueryError(err error) *AppError {
	return NewAppError(ErrCodeDatabaseQuery, "Database query failed", err)
}

func NewDatabaseTimeoutError(err error) *AppError {
	return NewAppError(ErrCodeDatabaseTimeout, "Database operation timed out", err)
}

func NewDatabaseConstraintError(err error) *AppError {
	return NewAppError(ErrCodeDatabaseConstraint, "Database constraint violation", err)
}

// Queue error constructors
func NewQueueConnectionError(err error) *AppError {
	return NewAppError(ErrCodeQueueConnection, "Queue connection failed", err)
}

func NewQueuePublishError(err error) *AppError {
	return NewAppError(ErrCodeQueuePublish, "Failed to publish message to queue", err)
}

func NewQueueTimeoutError(err error) *AppError {
	return NewAppError(ErrCodeQueueTimeout, "Queue operation timed out", err)
}

// Validation error constructors
func NewValidationError(message string) *AppError {
	return NewAppError(ErrCodeValidationFailed, message, nil)
}

func NewInvalidInputError(field string, value interface{}) *AppError {
	return NewAppError(ErrCodeInvalidInput, fmt.Sprintf("Invalid input for field '%s'", field), nil).
		WithContext("field", field).
		WithContext("value", value)
}

func NewMissingFieldError(field string) *AppError {
	return NewAppError(ErrCodeMissingField, fmt.Sprintf("Required field '%s' is missing", field), nil).
		WithContext("field", field)
}

// Service error constructors
func NewServiceUnavailableError(service string, err error) *AppError {
	return NewAppError(ErrCodeServiceUnavailable, fmt.Sprintf("Service '%s' is unavailable", service), err).
		WithContext("service", service)
}

func NewInternalError(err error) *AppError {
	return NewAppError(ErrCodeInternalError, "An internal error occurred", err)
}

func NewExternalServiceError(service string, err error) *AppError {
	return NewAppError(ErrCodeExternalService, fmt.Sprintf("External service '%s' error", service), err).
		WithContext("service", service)
}

// Resource error constructors
func NewNotFoundError(resource string, id string) *AppError {
	return NewAppError(ErrCodeNotFound, fmt.Sprintf("%s not found", resource), nil).
		WithContext("resource", resource).
		WithContext("id", id)
}

func NewAlreadyExistsError(resource string, id string) *AppError {
	return NewAppError(ErrCodeAlreadyExists, fmt.Sprintf("%s already exists", resource), nil).
		WithContext("resource", resource).
		WithContext("id", id)
}

func NewResourceConflictError(resource string, details string) *AppError {
	return NewAppError(ErrCodeResourceConflict, fmt.Sprintf("Conflict with %s", resource), nil).
		WithContext("resource", resource).
		WithDetails(details)
}

// Authorization error constructors
func NewUnauthorizedError(reason string) *AppError {
	return NewAppError(ErrCodeUnauthorized, "Unauthorized access", nil).
		WithDetails(reason)
}

func NewForbiddenError(reason string) *AppError {
	return NewAppError(ErrCodeForbidden, "Access forbidden", nil).
		WithDetails(reason)
}

func NewInvalidTokenError(err error) *AppError {
	return NewAppError(ErrCodeInvalidToken, "Invalid or expired token", err)
}