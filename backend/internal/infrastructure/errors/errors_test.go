package errors

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ErrorsTestSuite defines the test suite for error handling
type ErrorsTestSuite struct {
	suite.Suite
}

// TestErrorsTestSuite runs the test suite
func TestErrorsTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorsTestSuite))
}

// TestNew tests creating a new application error
func (suite *ErrorsTestSuite) TestNew() {
	code := ErrCodeValidation
	message := "Invalid input provided"
	
	err := New(code, message)
	
	assert := assert.New(suite.T())
	assert.NotNil(err)
	assert.Equal(code, err.Code)
	assert.Equal(message, err.Message)
	assert.Empty(err.Details)
	assert.False(err.Timestamp.IsZero())
	assert.NotEmpty(err.Stack)
	assert.False(err.Retryable) // Validation errors are not retryable
	assert.Empty(err.RequestID)
	assert.Empty(err.Component)
	assert.Nil(err.Metadata)
	assert.Nil(err.Cause)
}

// TestNewWithDetails tests creating a new application error with details
func (suite *ErrorsTestSuite) TestNewWithDetails() {
	code := ErrCodeValidation
	message := "Validation failed"
	details := "Email field is required"
	
	err := NewWithDetails(code, message, details)
	
	assert := assert.New(suite.T())
	assert.NotNil(err)
	assert.Equal(code, err.Code)
	assert.Equal(message, err.Message)
	assert.Equal(details, err.Details)
	assert.False(err.Timestamp.IsZero())
	assert.NotEmpty(err.Stack)
	assert.False(err.Retryable)
}

// TestWrap tests wrapping an existing error
func (suite *ErrorsTestSuite) TestWrap() {
	originalErr := errors.New("database connection failed")
	code := ErrCodeDatabase
	message := "Unable to save record"
	
	err := Wrap(originalErr, code, message)
	
	assert := assert.New(suite.T())
	assert.NotNil(err)
	assert.Equal(code, err.Code)
	assert.Equal(message, err.Message)
	assert.Equal(originalErr.Error(), err.Details)
	assert.Equal(originalErr, err.Cause)
	assert.False(err.Timestamp.IsZero())
	assert.NotEmpty(err.Stack)
	assert.False(err.Retryable)
}

// TestWrapNilError tests wrapping a nil error
func (suite *ErrorsTestSuite) TestWrapNilError() {
	err := Wrap(nil, ErrCodeDatabase, "Test message")
	
	assert := assert.New(suite.T())
	assert.Nil(err)
}

// TestWrapAppError tests wrapping an existing AppError
func (suite *ErrorsTestSuite) TestWrapAppError() {
	innerErr := New(ErrCodeValidation, "Inner validation error")
	outerErr := Wrap(innerErr, ErrCodeInternal, "Outer wrapper error")
	
	assert := assert.New(suite.T())
	assert.NotNil(outerErr)
	assert.Equal(ErrCodeInternal, outerErr.Code)
	assert.Equal("Outer wrapper error", outerErr.Message)
	assert.Equal(innerErr, outerErr.Cause)
	
	// Test unwrapping
	unwrapped := outerErr.Unwrap()
	assert.Equal(innerErr, unwrapped)
}

// TestErrorInterface tests the Error() method
func (suite *ErrorsTestSuite) TestErrorInterface() {
	testCases := []struct {
		name     string
		err      *AppError
		expected string
	}{
		{
			name:     "error without details",
			err:      New(ErrCodeValidation, "Validation failed"),
			expected: "VALIDATION_ERROR: Validation failed",
		},
		{
			name:     "error with details",
			err:      NewWithDetails(ErrCodeValidation, "Validation failed", "Email is required"),
			expected: "VALIDATION_ERROR: Validation failed (Email is required)",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.err.Error())
		})
	}
}

// TestWithMethods tests the fluent methods for adding context
func (suite *ErrorsTestSuite) TestWithMethods() {
	err := New(ErrCodeInternal, "Internal error")
	
	// Test WithRequestID
	err = err.WithRequestID("req-123")
	assert.Equal(suite.T(), "req-123", err.RequestID)
	
	// Test WithComponent
	err = err.WithComponent("database")
	assert.Equal(suite.T(), "database", err.Component)
	
	// Test WithMetadata
	err = err.WithMetadata("operation", "create_user")
	err = err.WithMetadata("user_id", "user-456")
	
	assert.NotNil(suite.T(), err.Metadata)
	assert.Equal(suite.T(), "create_user", err.Metadata["operation"])
	assert.Equal(suite.T(), "user-456", err.Metadata["user_id"])
	
	// Test WithCause
	cause := errors.New("underlying cause")
	err = err.WithCause(cause)
	assert.Equal(suite.T(), cause, err.Cause)
}

// TestHTTPStatusCode tests HTTP status code mapping
func (suite *ErrorsTestSuite) TestHTTPStatusCode() {
	testCases := []struct {
		code     ErrorCode
		expected int
	}{
		{ErrCodeNotFound, http.StatusNotFound},
		{ErrCodeValidation, http.StatusBadRequest},
		{ErrCodeInvalidRequest, http.StatusBadRequest},
		{ErrCodeAuthentication, http.StatusUnauthorized},
		{ErrCodeAuthorization, http.StatusForbidden},
		{ErrCodeRateLimit, http.StatusTooManyRequests},
		{ErrCodeClaudeRateLimit, http.StatusTooManyRequests},
		{ErrCodeTimeout, http.StatusRequestTimeout},
		{ErrCodeDatabaseTimeout, http.StatusRequestTimeout},
		{ErrCodeQueueTimeout, http.StatusRequestTimeout},
		{ErrCodeClaudeTimeout, http.StatusRequestTimeout},
		{ErrCodeRequestTimeout, http.StatusRequestTimeout},
		{ErrCodeSessionExpired, http.StatusUnauthorized},
		{ErrCodeInternal, http.StatusInternalServerError},
		{ErrCodeDatabase, http.StatusInternalServerError},
		{ErrCodeQueue, http.StatusInternalServerError},
		{ErrCodeClaude, http.StatusInternalServerError},
	}
	
	for _, tc := range testCases {
		suite.T().Run(string(tc.code), func(t *testing.T) {
			err := New(tc.code, "test message")
			assert.Equal(t, tc.expected, err.HTTPStatusCode())
		})
	}
}

// TestHTTPStatusCodeWithCustomStatusCode tests custom status code override
func (suite *ErrorsTestSuite) TestHTTPStatusCodeWithCustomStatusCode() {
	err := New(ErrCodeInternal, "test message")
	err.StatusCode = http.StatusServiceUnavailable
	
	assert.Equal(suite.T(), http.StatusServiceUnavailable, err.HTTPStatusCode())
}

// TestIsRetryable tests retryable error detection
func (suite *ErrorsTestSuite) TestIsRetryable() {
	retryableCodes := []ErrorCode{
		ErrCodeTimeout,
		ErrCodeDatabaseTimeout,
		ErrCodeQueueTimeout,
		ErrCodeClaudeTimeout,
		ErrCodeQueueConnection,
		ErrCodeRateLimit,
		ErrCodeClaudeRateLimit,
	}
	
	nonRetryableCodes := []ErrorCode{
		ErrCodeInternal,
		ErrCodeNotFound,
		ErrCodeValidation,
		ErrCodeAuthentication,
		ErrCodeAuthorization,
		ErrCodeDatabase,
		ErrCodeQueue,
		ErrCodeClaude,
		ErrCodeInvalidRequest,
		ErrCodeSessionExpired,
		ErrCodeDuplicateKey,
	}
	
	// Test retryable codes
	for _, code := range retryableCodes {
		suite.T().Run("retryable_"+string(code), func(t *testing.T) {
			err := New(code, "test message")
			assert.True(t, err.Retryable)
			assert.True(t, IsRetryable(err))
		})
	}
	
	// Test non-retryable codes
	for _, code := range nonRetryableCodes {
		suite.T().Run("non_retryable_"+string(code), func(t *testing.T) {
			err := New(code, "test message")
			assert.False(t, err.Retryable)
			assert.False(t, IsRetryable(err))
		})
	}
}

// TestIsRetryableWithNonAppError tests IsRetryable with non-AppError
func (suite *ErrorsTestSuite) TestIsRetryableWithNonAppError() {
	err := errors.New("standard error")
	assert.False(suite.T(), IsRetryable(err))
}

// TestGetErrorCode tests error code extraction
func (suite *ErrorsTestSuite) TestGetErrorCode() {
	// Test with AppError
	appErr := New(ErrCodeValidation, "test message")
	assert.Equal(suite.T(), ErrCodeValidation, GetErrorCode(appErr))
	
	// Test with wrapped AppError
	wrappedErr := Wrap(appErr, ErrCodeInternal, "wrapper message")
	assert.Equal(suite.T(), ErrCodeInternal, GetErrorCode(wrappedErr))
	
	// Test with non-AppError
	standardErr := errors.New("standard error")
	assert.Equal(suite.T(), ErrCodeInternal, GetErrorCode(standardErr))
}

// TestPredefinedErrorConstructors tests predefined error constructor functions
func (suite *ErrorsTestSuite) TestPredefinedErrorConstructors() {
	assert := assert.New(suite.T())
	
	// Test InternalError
	internalErr := InternalError("Something went wrong")
	assert.Equal(ErrCodeInternal, internalErr.Code)
	assert.Equal("Something went wrong", internalErr.Message)
	
	// Test NotFoundError
	notFoundErr := NotFoundError("User")
	assert.Equal(ErrCodeNotFound, notFoundErr.Code)
	assert.Equal("User not found", notFoundErr.Message)
	
	// Test ValidationError
	validationErr := ValidationError("email", "is required")
	assert.Equal(ErrCodeValidation, validationErr.Code)
	assert.Equal("Validation failed", validationErr.Message)
	assert.Equal("email: is required", validationErr.Details)
	
	// Test DatabaseError
	dbErr := DatabaseError("insert", errors.New("connection failed"))
	assert.Equal(ErrCodeDatabase, dbErr.Code)
	assert.Equal("Database operation failed: insert", dbErr.Message)
	assert.Equal("connection failed", dbErr.Details)
	
	// Test QueueError
	queueErr := QueueError("publish", errors.New("queue full"))
	assert.Equal(ErrCodeQueue, queueErr.Code)
	assert.Equal("Queue operation failed: publish", queueErr.Message)
	assert.Equal("queue full", queueErr.Details)
	
	// Test ClaudeError
	claudeErr := ClaudeError("completion", errors.New("rate limited"))
	assert.Equal(ErrCodeClaude, claudeErr.Code)
	assert.Equal("Claude API operation failed: completion", claudeErr.Message)
	assert.Equal("rate limited", claudeErr.Details)
	
	// Test TimeoutError
	timeoutErr := TimeoutError("database query", 30*time.Second)
	assert.Equal(ErrCodeTimeout, timeoutErr.Code)
	assert.Equal("Operation timed out: database query", timeoutErr.Message)
	assert.Equal("Timeout after 30s", timeoutErr.Details)
	
	// Test AuthenticationError
	authErr := AuthenticationError("Invalid credentials")
	assert.Equal(ErrCodeAuthentication, authErr.Code)
	assert.Equal("Invalid credentials", authErr.Message)
	
	// Test AuthorizationError
	authzErr := AuthorizationError("admin panel")
	assert.Equal(ErrCodeAuthorization, authzErr.Code)
	assert.Equal("Access denied to admin panel", authzErr.Message)
}

// TestStackTrace tests stack trace capture
func (suite *ErrorsTestSuite) TestStackTrace() {
	err := New(ErrCodeInternal, "test error")
	
	assert := assert.New(suite.T())
	assert.NotEmpty(err.Stack)
	assert.Contains(err.Stack, "TestStackTrace") // Should contain current function name
}

// TestErrorChaining tests error chaining and unwrapping
func (suite *ErrorsTestSuite) TestErrorChaining() {
	// Create a chain of errors
	originalErr := errors.New("original error")
	level1Err := Wrap(originalErr, ErrCodeDatabase, "database error")
	level2Err := Wrap(level1Err, ErrCodeInternal, "internal error")
	
	assert := assert.New(suite.T())
	
	// Test the top-level error
	assert.Equal(ErrCodeInternal, level2Err.Code)
	assert.Equal("internal error", level2Err.Message)
	
	// Test unwrapping to level 1
	unwrapped1 := level2Err.Unwrap()
	assert.Equal(level1Err, unwrapped1)
	
	// Test unwrapping to original
	unwrapped2 := unwrapped1.(*AppError).Unwrap()
	assert.Equal(originalErr, unwrapped2)
	
	// Test errors.Is functionality
	assert.True(errors.Is(level2Err, level1Err))
	assert.True(errors.Is(level2Err, originalErr))
	
	// Test errors.As functionality
	var appErr *AppError
	assert.True(errors.As(level2Err, &appErr))
	assert.Equal(level2Err, appErr)
}

// TestErrorSerialization tests error field serialization
func (suite *ErrorsTestSuite) TestErrorSerialization() {
	err := New(ErrCodeValidation, "Validation failed").
		WithRequestID("req-123").
		WithComponent("user-service").
		WithMetadata("field", "email").
		WithMetadata("value", "invalid@")
	
	assert := assert.New(suite.T())
	assert.Equal(ErrCodeValidation, err.Code)
	assert.Equal("Validation failed", err.Message)
	assert.Equal("req-123", err.RequestID)
	assert.Equal("user-service", err.Component)
	assert.Equal("email", err.Metadata["field"])
	assert.Equal("invalid@", err.Metadata["value"])
	assert.False(err.Timestamp.IsZero())
	assert.NotEmpty(err.Stack)
}

// TestErrorCodes tests all defined error codes
func (suite *ErrorsTestSuite) TestErrorCodes() {
	errorCodes := []ErrorCode{
		ErrCodeInternal,
		ErrCodeNotFound,
		ErrCodeValidation,
		ErrCodeAuthentication,
		ErrCodeAuthorization,
		ErrCodeRateLimit,
		ErrCodeTimeout,
		ErrCodeDatabase,
		ErrCodeDatabaseTimeout,
		ErrCodeDuplicateKey,
		ErrCodeQueue,
		ErrCodeQueueTimeout,
		ErrCodeQueueConnection,
		ErrCodeClaude,
		ErrCodeClaudeTimeout,
		ErrCodeClaudeRateLimit,
		ErrCodeInvalidRequest,
		ErrCodeRequestTimeout,
		ErrCodeSessionExpired,
	}
	
	assert := assert.New(suite.T())
	
	// Test that all error codes are non-empty strings
	for _, code := range errorCodes {
		assert.NotEmpty(string(code))
		
		// Test that errors can be created with each code
		err := New(code, "test message")
		assert.Equal(code, err.Code)
		assert.NotZero(err.HTTPStatusCode())
	}
}

// TestMetadataHandling tests metadata handling edge cases
func (suite *ErrorsTestSuite) TestMetadataHandling() {
	err := New(ErrCodeInternal, "test error")
	
	assert := assert.New(suite.T())
	
	// Initially no metadata
	assert.Nil(err.Metadata)
	
	// Add first metadata item
	err.WithMetadata("key1", "value1")
	assert.NotNil(err.Metadata)
	assert.Equal("value1", err.Metadata["key1"])
	
	// Add more metadata
	err.WithMetadata("key2", 123)
	err.WithMetadata("key3", true)
	err.WithMetadata("key4", []string{"a", "b", "c"})
	err.WithMetadata("key5", map[string]string{"nested": "value"})
	
	assert.Len(err.Metadata, 5)
	assert.Equal("value1", err.Metadata["key1"])
	assert.Equal(123, err.Metadata["key2"])
	assert.Equal(true, err.Metadata["key3"])
	assert.Equal([]string{"a", "b", "c"}, err.Metadata["key4"])
	assert.Equal(map[string]string{"nested": "value"}, err.Metadata["key5"])
	
	// Test overwriting metadata
	err.WithMetadata("key1", "new_value")
	assert.Equal("new_value", err.Metadata["key1"])
}

// TestComplexErrorScenarios tests complex real-world error scenarios
func (suite *ErrorsTestSuite) TestComplexErrorScenarios() {
	assert := assert.New(suite.T())
	
	// Scenario 1: Database connection timeout during user creation
	dbConnErr := errors.New("connection timeout")
	userCreationErr := DatabaseError("create_user", dbConnErr).
		WithRequestID("req-456").
		WithComponent("user-service").
		WithMetadata("user_email", "test@example.com").
		WithMetadata("operation", "signup")
	
	assert.Equal(ErrCodeDatabase, userCreationErr.Code)
	assert.Contains(userCreationErr.Message, "create_user")
	assert.Equal("connection timeout", userCreationErr.Details)
	assert.Equal("req-456", userCreationErr.RequestID)
	assert.Equal("user-service", userCreationErr.Component)
	assert.Equal("test@example.com", userCreationErr.Metadata["user_email"])
	assert.Equal(dbConnErr, userCreationErr.Cause)
	assert.False(userCreationErr.Retryable)
	assert.Equal(http.StatusInternalServerError, userCreationErr.HTTPStatusCode())
	
	// Scenario 2: Claude API rate limit with retry logic
	claudeErr := ClaudeError("text_completion", errors.New("rate limit exceeded")).
		WithRequestID("req-789").
		WithComponent("claude-service").
		WithMetadata("model", "claude-3-sonnet").
		WithMetadata("retry_count", 2)
	
	// Wrap with timeout error (retryable)
	timeoutErr := Wrap(claudeErr, ErrCodeClaudeRateLimit, "Claude API rate limited")
	
	assert.Equal(ErrCodeClaudeRateLimit, timeoutErr.Code)
	assert.True(timeoutErr.Retryable)
	assert.Equal(http.StatusTooManyRequests, timeoutErr.HTTPStatusCode())
	assert.Equal(claudeErr, timeoutErr.Cause)
	
	// Scenario 3: Validation error with multiple field issues
	validationErr := ValidationError("user_data", "multiple validation errors").
		WithRequestID("req-101").
		WithComponent("validation-service").
		WithMetadata("errors", map[string]string{
			"email":    "invalid format",
			"password": "too short",
			"age":      "must be positive",
		})
	
	assert.Equal(ErrCodeValidation, validationErr.Code)
	assert.Equal(http.StatusBadRequest, validationErr.HTTPStatusCode())
	assert.False(validationErr.Retryable)
	assert.NotNil(validationErr.Metadata["errors"])
}

// BenchmarkErrorCreation benchmarks error creation
func BenchmarkErrorCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New(ErrCodeInternal, "test error")
	}
}

// BenchmarkErrorWithContext benchmarks error creation with context
func BenchmarkErrorWithContext(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New(ErrCodeInternal, "test error").
			WithRequestID("req-123").
			WithComponent("test-component").
			WithMetadata("key", "value")
	}
}

// BenchmarkErrorWrapping benchmarks error wrapping
func BenchmarkErrorWrapping(b *testing.B) {
	originalErr := errors.New("original error")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Wrap(originalErr, ErrCodeDatabase, "database error")
	}
}

// BenchmarkHTTPStatusCode benchmarks HTTP status code lookup
func BenchmarkHTTPStatusCode(b *testing.B) {
	err := New(ErrCodeValidation, "test error")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err.HTTPStatusCode()
	}
}