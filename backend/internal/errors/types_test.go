package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAppError(t *testing.T) {
	cause := errors.New("underlying error")
	appErr := NewAppError(ErrCodeDatabaseConnection, "Database connection failed", cause)

	assert.Equal(t, ErrCodeDatabaseConnection, appErr.Code)
	assert.Equal(t, "Database connection failed", appErr.Message)
	assert.Equal(t, cause, appErr.Cause)
	assert.Equal(t, SeverityCritical, appErr.Severity)
	assert.Equal(t, 503, appErr.HTTPStatus)
	assert.True(t, appErr.Retryable)
	assert.NotZero(t, appErr.Timestamp)
}

func TestAppErrorWithContext(t *testing.T) {
	appErr := NewDatabaseConnectionError(nil).
		WithContext("host", "localhost").
		WithContext("port", 5432)

	assert.Equal(t, "localhost", appErr.Context["host"])
	assert.Equal(t, 5432, appErr.Context["port"])
}

func TestAppErrorWithRequestID(t *testing.T) {
	requestID := "req-123"
	appErr := NewValidationError("test error").WithRequestID(requestID)

	assert.Equal(t, requestID, appErr.RequestID)
}

func TestAppErrorWithDetails(t *testing.T) {
	details := "Additional error details"
	appErr := NewInternalError(nil).WithDetails(details)

	assert.Equal(t, details, appErr.Details)
}

func TestErrorUnwrap(t *testing.T) {
	cause := errors.New("root cause")
	appErr := NewDatabaseQueryError(cause)

	assert.Equal(t, cause, appErr.Unwrap())
}

func TestDatabaseErrorConstructors(t *testing.T) {
	cause := errors.New("test error")

	tests := []struct {
		name           string
		constructor    func(error) *AppError
		expectedCode   ErrorCode
		expectedStatus int
	}{
		{
			name:           "DatabaseConnectionError",
			constructor:    NewDatabaseConnectionError,
			expectedCode:   ErrCodeDatabaseConnection,
			expectedStatus: 503,
		},
		{
			name:           "DatabaseQueryError",
			constructor:    NewDatabaseQueryError,
			expectedCode:   ErrCodeDatabaseQuery,
			expectedStatus: 500,
		},
		{
			name:           "DatabaseTimeoutError",
			constructor:    NewDatabaseTimeoutError,
			expectedCode:   ErrCodeDatabaseTimeout,
			expectedStatus: 408,
		},
		{
			name:           "DatabaseConstraintError",
			constructor:    NewDatabaseConstraintError,
			expectedCode:   ErrCodeDatabaseConstraint,
			expectedStatus: 409,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := tt.constructor(cause)
			assert.Equal(t, tt.expectedCode, appErr.Code)
			assert.Equal(t, tt.expectedStatus, appErr.HTTPStatus)
			assert.Equal(t, cause, appErr.Cause)
		})
	}
}

func TestQueueErrorConstructors(t *testing.T) {
	cause := errors.New("queue error")

	tests := []struct {
		name           string
		constructor    func(error) *AppError
		expectedCode   ErrorCode
		expectedStatus int
	}{
		{
			name:           "QueueConnectionError",
			constructor:    NewQueueConnectionError,
			expectedCode:   ErrCodeQueueConnection,
			expectedStatus: 503,
		},
		{
			name:           "QueuePublishError",
			constructor:    NewQueuePublishError,
			expectedCode:   ErrCodeQueuePublish,
			expectedStatus: 500,
		},
		{
			name:           "QueueTimeoutError",
			constructor:    NewQueueTimeoutError,
			expectedCode:   ErrCodeQueueTimeout,
			expectedStatus: 408,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := tt.constructor(cause)
			assert.Equal(t, tt.expectedCode, appErr.Code)
			assert.Equal(t, tt.expectedStatus, appErr.HTTPStatus)
			assert.Equal(t, cause, appErr.Cause)
		})
	}
}

func TestValidationErrorConstructors(t *testing.T) {
	t.Run("ValidationError", func(t *testing.T) {
		message := "Validation failed"
		appErr := NewValidationError(message)
		assert.Equal(t, ErrCodeValidationFailed, appErr.Code)
		assert.Equal(t, message, appErr.Message)
		assert.Equal(t, 400, appErr.HTTPStatus)
		assert.False(t, appErr.Retryable)
	})

	t.Run("InvalidInputError", func(t *testing.T) {
		field := "email"
		value := "invalid-email"
		appErr := NewInvalidInputError(field, value)
		assert.Equal(t, ErrCodeInvalidInput, appErr.Code)
		assert.Contains(t, appErr.Message, field)
		assert.Equal(t, field, appErr.Context["field"])
		assert.Equal(t, value, appErr.Context["value"])
	})

	t.Run("MissingFieldError", func(t *testing.T) {
		field := "username"
		appErr := NewMissingFieldError(field)
		assert.Equal(t, ErrCodeMissingField, appErr.Code)
		assert.Contains(t, appErr.Message, field)
		assert.Equal(t, field, appErr.Context["field"])
	})
}

func TestResourceErrorConstructors(t *testing.T) {
	t.Run("NotFoundError", func(t *testing.T) {
		resource := "user"
		id := "123"
		appErr := NewNotFoundError(resource, id)
		assert.Equal(t, ErrCodeNotFound, appErr.Code)
		assert.Contains(t, appErr.Message, resource)
		assert.Equal(t, resource, appErr.Context["resource"])
		assert.Equal(t, id, appErr.Context["id"])
		assert.Equal(t, 404, appErr.HTTPStatus)
	})

	t.Run("AlreadyExistsError", func(t *testing.T) {
		resource := "email"
		id := "test@example.com"
		appErr := NewAlreadyExistsError(resource, id)
		assert.Equal(t, ErrCodeAlreadyExists, appErr.Code)
		assert.Contains(t, appErr.Message, resource)
		assert.Equal(t, resource, appErr.Context["resource"])
		assert.Equal(t, id, appErr.Context["id"])
		assert.Equal(t, 409, appErr.HTTPStatus)
	})

	t.Run("ResourceConflictError", func(t *testing.T) {
		resource := "task"
		details := "Task is already running"
		appErr := NewResourceConflictError(resource, details)
		assert.Equal(t, ErrCodeResourceConflict, appErr.Code)
		assert.Contains(t, appErr.Message, resource)
		assert.Equal(t, resource, appErr.Context["resource"])
		assert.Equal(t, details, appErr.Details)
	})
}

func TestServiceErrorConstructors(t *testing.T) {
	t.Run("ServiceUnavailableError", func(t *testing.T) {
		service := "payment-service"
		cause := errors.New("connection refused")
		appErr := NewServiceUnavailableError(service, cause)
		assert.Equal(t, ErrCodeServiceUnavailable, appErr.Code)
		assert.Contains(t, appErr.Message, service)
		assert.Equal(t, service, appErr.Context["service"])
		assert.Equal(t, cause, appErr.Cause)
		assert.True(t, appErr.Retryable)
	})

	t.Run("InternalError", func(t *testing.T) {
		cause := errors.New("unexpected error")
		appErr := NewInternalError(cause)
		assert.Equal(t, ErrCodeInternalError, appErr.Code)
		assert.Equal(t, cause, appErr.Cause)
		assert.Equal(t, 500, appErr.HTTPStatus)
	})

	t.Run("ExternalServiceError", func(t *testing.T) {
		service := "github-api"
		cause := errors.New("API rate limit exceeded")
		appErr := NewExternalServiceError(service, cause)
		assert.Equal(t, ErrCodeExternalService, appErr.Code)
		assert.Contains(t, appErr.Message, service)
		assert.Equal(t, service, appErr.Context["service"])
		assert.Equal(t, cause, appErr.Cause)
	})
}

func TestAuthorizationErrorConstructors(t *testing.T) {
	t.Run("UnauthorizedError", func(t *testing.T) {
		reason := "Invalid credentials"
		appErr := NewUnauthorizedError(reason)
		assert.Equal(t, ErrCodeUnauthorized, appErr.Code)
		assert.Equal(t, reason, appErr.Details)
		assert.Equal(t, 401, appErr.HTTPStatus)
	})

	t.Run("ForbiddenError", func(t *testing.T) {
		reason := "Insufficient permissions"
		appErr := NewForbiddenError(reason)
		assert.Equal(t, ErrCodeForbidden, appErr.Code)
		assert.Equal(t, reason, appErr.Details)
		assert.Equal(t, 403, appErr.HTTPStatus)
	})

	t.Run("InvalidTokenError", func(t *testing.T) {
		cause := errors.New("token expired")
		appErr := NewInvalidTokenError(cause)
		assert.Equal(t, ErrCodeInvalidToken, appErr.Code)
		assert.Equal(t, cause, appErr.Cause)
		assert.Equal(t, 401, appErr.HTTPStatus)
	})
}

func TestGetErrorProperties(t *testing.T) {
	tests := []struct {
		code             ErrorCode
		expectedSeverity ErrorSeverity
		expectedStatus   int
		expectedRetryable bool
	}{
		{ErrCodeDatabaseConnection, SeverityCritical, 503, true},
		{ErrCodeDatabaseTimeout, SeverityHigh, 408, true},
		{ErrCodeValidationFailed, SeverityLow, 400, false},
		{ErrCodeNotFound, SeverityLow, 404, false},
		{ErrCodeInternalError, SeverityMedium, 500, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			severity, status, retryable := getErrorProperties(tt.code)
			assert.Equal(t, tt.expectedSeverity, severity)
			assert.Equal(t, tt.expectedStatus, status)
			assert.Equal(t, tt.expectedRetryable, retryable)
		})
	}
}

func TestAppErrorError(t *testing.T) {
	t.Run("WithDetails", func(t *testing.T) {
		appErr := NewValidationError("Invalid input").WithDetails("Field 'email' is required")
		expected := "VALIDATION_FAILED: Invalid input (Field 'email' is required)"
		assert.Equal(t, expected, appErr.Error())
	})

	t.Run("WithoutDetails", func(t *testing.T) {
		appErr := NewValidationError("Invalid input")
		expected := "VALIDATION_FAILED: Invalid input"
		assert.Equal(t, expected, appErr.Error())
	})
}

func TestAppErrorChaining(t *testing.T) {
	appErr := NewDatabaseConnectionError(nil).
		WithRequestID("req-123").
		WithContext("host", "localhost").
		WithContext("port", 5432).
		WithDetails("Connection timeout after 30 seconds")

	assert.Equal(t, "req-123", appErr.RequestID)
	assert.Equal(t, "localhost", appErr.Context["host"])
	assert.Equal(t, 5432, appErr.Context["port"])
	assert.Equal(t, "Connection timeout after 30 seconds", appErr.Details)
	assert.Equal(t, ErrCodeDatabaseConnection, appErr.Code)
}