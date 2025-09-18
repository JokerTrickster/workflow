package infrastructure

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/JokerTrickster/workflow/local-backend/internal/domain"
)

// ErrorHandler provides centralized error handling and recovery
type ErrorHandler struct {
	logger *Logger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// HandleError handles errors with appropriate logging and recovery
func (e *ErrorHandler) HandleError(ctx context.Context, err error, operation string) error {
	if err == nil {
		return nil
	}

	// Extract context information
	contextInfo := e.extractContextInfo(ctx)
	
	// Log the error with context
	e.logger.Error(fmt.Sprintf("Error in %s: %v", operation, err), 
		"operation", operation,
		"context", contextInfo,
		"timestamp", time.Now().Unix())

	// Handle specific error types
	switch err := err.(type) {
	case *domain.BusinessError:
		return e.handleBusinessError(err)
	default:
		return e.handleGenericError(err, operation)
	}
}

// HandlePanic recovers from panics and logs them
func (e *ErrorHandler) HandlePanic(operation string) {
	if r := recover(); r != nil {
		stack := debug.Stack()
		e.logger.Error(fmt.Sprintf("Panic in %s: %v", operation, r),
			"operation", operation,
			"panic_value", r,
			"stack_trace", string(stack),
			"timestamp", time.Now().Unix())
		
		// In production, you might want to:
		// 1. Send alert to monitoring system
		// 2. Restart the service gracefully
		// 3. Log to external error tracking service
	}
}

// WrapWithRecovery wraps a function with panic recovery
func (e *ErrorHandler) WrapWithRecovery(operation string, fn func() error) error {
	defer e.HandlePanic(operation)
	return fn()
}

// handleBusinessError handles domain-specific business errors
func (e *ErrorHandler) handleBusinessError(err *domain.BusinessError) error {
	// Business errors are expected and should not be treated as system failures
	e.logger.Warn(fmt.Sprintf("Business error: %s", err.Message),
		"error_code", err.Code,
		"cause", err.Cause)
	
	return err // Return as-is for business logic to handle
}

// handleGenericError handles unexpected system errors
func (e *ErrorHandler) handleGenericError(err error, operation string) error {
	// System errors might require different handling
	e.logger.Error(fmt.Sprintf("System error in %s", operation),
		"error", err.Error(),
		"type", fmt.Sprintf("%T", err))
	
	// Wrap in a generic system error
	return fmt.Errorf("system error in %s: %w", operation, err)
}

// extractContextInfo extracts useful information from context
func (e *ErrorHandler) extractContextInfo(ctx context.Context) map[string]interface{} {
	info := make(map[string]interface{})
	
	// Extract deadline if set
	if deadline, ok := ctx.Deadline(); ok {
		info["deadline"] = deadline.Unix()
		info["time_remaining"] = time.Until(deadline).Seconds()
	}
	
	// Extract values from context if they exist
	// This would be extended based on what context values your app uses
	if val := ctx.Value("request_id"); val != nil {
		info["request_id"] = val
	}
	
	if val := ctx.Value("user_id"); val != nil {
		info["user_id"] = val
	}
	
	return info
}

// RetryWithBackoff retries an operation with exponential backoff
func (e *ErrorHandler) RetryWithBackoff(ctx context.Context, operation string, maxRetries int, fn func() error) error {
	var lastErr error
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Execute the function
		err := fn()
		if err == nil {
			if attempt > 0 {
				e.logger.Info(fmt.Sprintf("Operation %s succeeded after %d retries", operation, attempt))
			}
			return nil
		}
		
		lastErr = err
		
		// Don't retry on business errors
		if _, ok := err.(*domain.BusinessError); ok {
			return err
		}
		
		// Check if context is done
		if ctx.Err() != nil {
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		}
		
		// Calculate backoff delay
		delay := time.Duration(attempt+1) * time.Second
		
		e.logger.Warn(fmt.Sprintf("Operation %s failed (attempt %d/%d), retrying in %v", 
			operation, attempt+1, maxRetries, delay),
			"error", err.Error(),
			"attempt", attempt+1,
			"max_retries", maxRetries)
		
		// Wait before retry
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled during retry: %w", ctx.Err())
		case <-time.After(delay):
			// Continue to next attempt
		}
	}
	
	return fmt.Errorf("operation %s failed after %d attempts: %w", operation, maxRetries, lastErr)
}

// IsRetryable determines if an error is worth retrying
func (e *ErrorHandler) IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	
	// Don't retry business errors
	if _, ok := err.(*domain.BusinessError); ok {
		return false
	}
	
	// Don't retry context cancellation
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}
	
	// Add more specific retry logic based on error types
	errorStr := err.Error()
	
	// Network errors are usually retryable
	if strings.Contains(errorStr, "connection refused") ||
	   strings.Contains(errorStr, "timeout") ||
	   strings.Contains(errorStr, "temporary failure") {
		return true
	}
	
	// Database lock errors might be retryable
	if strings.Contains(errorStr, "database is locked") {
		return true
	}
	
	return false
}


// LogOperationStart logs the start of an operation
func (e *ErrorHandler) LogOperationStart(operation string, params map[string]interface{}) {
	e.logger.Info(fmt.Sprintf("Starting operation: %s", operation),
		"operation", operation,
		"params", params,
		"timestamp", time.Now().Unix())
}

// LogOperationComplete logs the completion of an operation
func (e *ErrorHandler) LogOperationComplete(operation string, duration time.Duration, result interface{}) {
	e.logger.Info(fmt.Sprintf("Completed operation: %s", operation),
		"operation", operation,
		"duration_ms", duration.Milliseconds(),
		"result", result,
		"timestamp", time.Now().Unix())
}