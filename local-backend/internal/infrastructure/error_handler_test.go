package infrastructure

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JokerTrickster/workflow/local-backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewErrorHandler(t *testing.T) {
	logger := NewLogger(&LoggingConfig{Level: "info", Format: "text", Output: "stdout"})
	handler := NewErrorHandler(logger)
	
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.logger)
}

func TestErrorHandler_HandleError(t *testing.T) {
	logger := NewLogger(&LoggingConfig{Level: "info", Format: "text", Output: "stdout"})
	handler := NewErrorHandler(logger)
	ctx := context.Background()

	t.Run("nil error", func(t *testing.T) {
		result := handler.HandleError(ctx, nil, "test-operation")
		assert.NoError(t, result)
	})

	t.Run("business error", func(t *testing.T) {
		businessErr := &domain.BusinessError{
			Code:    "INVALID_REQUEST",
			Message: "Request validation failed",
			Cause:   "Missing required field",
		}
		
		result := handler.HandleError(ctx, businessErr, "validate-request")
		assert.Error(t, result)
		assert.IsType(t, &domain.BusinessError{}, result)
		assert.Equal(t, businessErr, result)
	})

	t.Run("generic error", func(t *testing.T) {
		genericErr := errors.New("database connection failed")
		
		result := handler.HandleError(ctx, genericErr, "database-query")
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "system error in database-query")
		assert.Contains(t, result.Error(), "database connection failed")
	})

	t.Run("error with context values", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "req-123")
		ctx = context.WithValue(ctx, "user_id", "user-456")
		
		err := errors.New("test error")
		result := handler.HandleError(ctx, err, "test-operation")
		assert.Error(t, result)
	})

	t.Run("error with deadline context", func(t *testing.T) {
		deadline := time.Now().Add(5 * time.Second)
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()
		
		err := errors.New("test error")
		result := handler.HandleError(ctx, err, "test-operation")
		assert.Error(t, result)
	})
}

func TestErrorHandler_HandlePanic(t *testing.T) {
	logger := NewLogger(&LoggingConfig{Level: "info", Format: "text", Output: "stdout"})
	handler := NewErrorHandler(logger)

	t.Run("no panic", func(t *testing.T) {
		// This should not panic
		defer handler.HandlePanic("test-operation")
		// Normal execution
	})

	t.Run("with panic", func(t *testing.T) {
		defer func() {
			// The panic should be recovered by HandlePanic
			if r := recover(); r != nil {
				t.Errorf("Panic was not handled: %v", r)
			}
		}()

		func() {
			defer handler.HandlePanic("test-operation")
			panic("test panic")
		}()
	})
}

func TestErrorHandler_WrapWithRecovery(t *testing.T) {
	logger := NewLogger(&LoggingConfig{Level: "info", Format: "text", Output: "stdout"})
	handler := NewErrorHandler(logger)

	t.Run("successful execution", func(t *testing.T) {
		err := handler.WrapWithRecovery("test-op", func() error {
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("function returns error", func(t *testing.T) {
		expectedErr := errors.New("function error")
		
		err := handler.WrapWithRecovery("test-op", func() error {
			return expectedErr
		})
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("function panics", func(t *testing.T) {
		err := handler.WrapWithRecovery("test-op", func() error {
			panic("test panic")
		})
		
		// The function should complete without panicking
		// The panic is handled by HandlePanic, so no error is returned
		assert.NoError(t, err)
	})
}

func TestErrorHandler_RetryWithBackoff(t *testing.T) {
	logger := NewLogger(&LoggingConfig{Level: "info", Format: "text", Output: "stdout"})
	handler := NewErrorHandler(logger)

	t.Run("success on first attempt", func(t *testing.T) {
		ctx := context.Background()
		
		err := handler.RetryWithBackoff(ctx, "test-op", 3, func() error {
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("success after retries", func(t *testing.T) {
		ctx := context.Background()
		attempt := 0
		
		err := handler.RetryWithBackoff(ctx, "test-op", 3, func() error {
			attempt++
			if attempt < 3 {
				return errors.New("temporary failure")
			}
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 3, attempt)
	})

	t.Run("fail after max retries", func(t *testing.T) {
		ctx := context.Background()
		
		err := handler.RetryWithBackoff(ctx, "test-op", 2, func() error {
			return errors.New("persistent failure")
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "operation test-op failed after 2 attempts")
	})

	t.Run("business error no retry", func(t *testing.T) {
		ctx := context.Background()
		businessErr := &domain.BusinessError{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid input",
		}
		
		attempt := 0
		err := handler.RetryWithBackoff(ctx, "test-op", 3, func() error {
			attempt++
			return businessErr
		})
		
		assert.Error(t, err)
		assert.Equal(t, businessErr, err)
		assert.Equal(t, 1, attempt) // Should not retry business errors
	})

	t.Run("context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		err := handler.RetryWithBackoff(ctx, "test-op", 3, func() error {
			return errors.New("test error")
		})
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "operation cancelled")
	})

	t.Run("context cancelled during retry", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		
		err := handler.RetryWithBackoff(ctx, "test-op", 5, func() error {
			return errors.New("test error")
		})
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "operation cancelled")
	})
}

func TestErrorHandler_IsRetryable(t *testing.T) {
	logger := NewLogger(&LoggingConfig{Level: "info", Format: "text", Output: "stdout"})
	handler := NewErrorHandler(logger)

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name: "business error",
			err: &domain.BusinessError{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid input",
			},
			expected: false,
		},
		{
			name:     "context cancelled",
			err:      context.Canceled,
			expected: false,
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			expected: false,
		},
		{
			name:     "connection refused",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "timeout error",
			err:      errors.New("operation timeout"),
			expected: true,
		},
		{
			name:     "temporary failure",
			err:      errors.New("temporary failure in network"),
			expected: true,
		},
		{
			name:     "database locked",
			err:      errors.New("database is locked"),
			expected: true,
		},
		{
			name:     "generic error",
			err:      errors.New("some generic error"),
			expected: false,
		},
		{
			name:     "authentication error",
			err:      errors.New("authentication failed"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.IsRetryable(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorHandler_LogOperationStart(t *testing.T) {
	logger := NewLogger(&LoggingConfig{Level: "info", Format: "text", Output: "stdout"})
	handler := NewErrorHandler(logger)

	params := map[string]interface{}{
		"user_id":    "user-123",
		"request_id": "req-456",
	}

	// This should not panic or error
	handler.LogOperationStart("test-operation", params)
	handler.LogOperationStart("test-operation", nil)
}

func TestErrorHandler_LogOperationComplete(t *testing.T) {
	logger := NewLogger(&LoggingConfig{Level: "info", Format: "text", Output: "stdout"})
	handler := NewErrorHandler(logger)

	duration := 150 * time.Millisecond
	result := map[string]interface{}{
		"status": "success",
		"count":  42,
	}

	// This should not panic or error
	handler.LogOperationComplete("test-operation", duration, result)
	handler.LogOperationComplete("test-operation", duration, nil)
}

func TestErrorHandler_ExtractContextInfo(t *testing.T) {
	logger := NewLogger(&LoggingConfig{Level: "info", Format: "text", Output: "stdout"})
	handler := NewErrorHandler(logger)

	t.Run("empty context", func(t *testing.T) {
		ctx := context.Background()
		info := handler.extractContextInfo(ctx)
		
		assert.NotNil(t, info)
		// Should be empty or contain no specific keys
	})

	t.Run("context with deadline", func(t *testing.T) {
		deadline := time.Now().Add(5 * time.Second)
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()
		
		info := handler.extractContextInfo(ctx)
		
		assert.NotNil(t, info)
		assert.Contains(t, info, "deadline")
		assert.Contains(t, info, "time_remaining")
		
		timeRemaining := info["time_remaining"].(float64)
		assert.True(t, timeRemaining > 0 && timeRemaining <= 5)
	})

	t.Run("context with values", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "req-123")
		ctx = context.WithValue(ctx, "user_id", "user-456")
		
		info := handler.extractContextInfo(ctx)
		
		assert.NotNil(t, info)
		assert.Equal(t, "req-123", info["request_id"])
		assert.Equal(t, "user-456", info["user_id"])
	})

	t.Run("context with deadline and values", func(t *testing.T) {
		deadline := time.Now().Add(3 * time.Second)
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()
		
		ctx = context.WithValue(ctx, "request_id", "req-789")
		
		info := handler.extractContextInfo(ctx)
		
		assert.NotNil(t, info)
		assert.Contains(t, info, "deadline")
		assert.Contains(t, info, "time_remaining")
		assert.Equal(t, "req-789", info["request_id"])
	})
}