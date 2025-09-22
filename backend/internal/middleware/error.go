package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"local-backend-server/internal/errors"
	"local-backend-server/internal/monitoring"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		
		// Add to context for downstream services
		ctx := context.WithValue(c.Request.Context(), errors.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)
		
		c.Next()
	}
}

// ErrorHandlingMiddleware provides comprehensive error handling
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// Record request for metrics
		monitoring.RecordGlobalRequest()
		
		// Handle panic as critical error
		var appErr *errors.AppError
		if err, ok := recovered.(error); ok {
			appErr = errors.NewInternalError(err).
				WithRequestID(getRequestID(c)).
				WithContext("panic", true)
		} else {
			appErr = errors.NewInternalError(fmt.Errorf("panic: %v", recovered)).
				WithRequestID(getRequestID(c)).
				WithContext("panic", true)
		}
		
		// Log the error
		logger := errors.NewLogger().WithRequestID(getRequestID(c))
		logger.Fatal(appErr, "Panic occurred during request processing")
		
		// Record error metrics
		monitoring.RecordGlobalError(appErr)
		
		// Send error response
		sendErrorResponse(c, appErr)
	})
}

// ErrorHandler handles errors that occur during request processing
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record request for metrics
		monitoring.RecordGlobalRequest()
		
		c.Next()
		
		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			var appErr *errors.AppError
			
			// Check if it's already an AppError
			if appError, ok := err.Err.(*errors.AppError); ok {
				appErr = appError
			} else {
				// Convert to AppError
				appErr = errors.NewInternalError(err.Err)
			}
			
			// Add request context
			appErr = appErr.WithRequestID(getRequestID(c))
			
			// Log the error
			logger := errors.NewLogger().WithRequestID(getRequestID(c))
			logger.Error(appErr, "Request processing error")
			
			// Record error metrics
			monitoring.RecordGlobalError(appErr)
			
			// Send error response if not already sent
			if !c.Writer.Written() {
				sendErrorResponse(c, appErr)
			}
		}
	}
}

// ValidationErrorHandler creates validation errors from gin binding errors
func ValidationErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// Check for validation errors
		for _, err := range c.Errors {
			if err.Type == gin.ErrorTypeBind {
				// Create validation error response
				validationErrors := []errors.ValidationError{
					{
						Field:   "request_body",
						Message: err.Error(),
						Code:    "BINDING_ERROR",
					},
				}
				
				errResponse := errors.NewValidationErrorResponse(validationErrors).
					WithRequestInfo(getRequestID(c), c.Request.URL.Path, c.Request.Method)
				
				// Create AppError for logging and metrics
				appErr := errors.NewValidationError("Request validation failed").
					WithRequestID(getRequestID(c)).
					WithDetails(err.Error())
				
				// Log validation error
				logger := errors.NewLogger().WithRequestID(getRequestID(c))
				logger.WarnWithContext("Validation error", map[string]interface{}{
					"error": err.Error(),
					"type":  "validation",
				})
				
				// Record error metrics
				monitoring.RecordGlobalError(appErr)
				
				// Send validation error response
				c.JSON(http.StatusBadRequest, errResponse)
				c.Abort()
				return
			}
		}
	}
}

// TimeoutMiddleware adds request timeout handling
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		
		c.Request = c.Request.WithContext(ctx)
		
		done := make(chan struct{})
		go func() {
			c.Next()
			close(done)
		}()
		
		select {
		case <-done:
			// Request completed normally
		case <-ctx.Done():
			// Request timed out
			appErr := errors.NewServiceUnavailableError("request", ctx.Err()).
				WithRequestID(getRequestID(c)).
				WithDetails(fmt.Sprintf("Request timed out after %v", timeout))
			
			// Log timeout error
			logger := errors.NewLogger().WithRequestID(getRequestID(c))
			logger.Error(appErr, "Request timeout")
			
			// Record error metrics
			monitoring.RecordGlobalError(appErr)
			
			// Send timeout response
			sendErrorResponse(c, appErr)
			c.Abort()
		}
	}
}

// HealthCheckMiddleware provides health check with error metrics
func HealthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			healthStatus := monitoring.GetGlobalHealthStatus()
			
			var httpStatus int
			switch healthStatus.Status {
			case "healthy":
				httpStatus = http.StatusOK
			case "degraded":
				httpStatus = http.StatusOK // Still OK but with warnings
			case "unhealthy":
				httpStatus = http.StatusServiceUnavailable
			default:
				httpStatus = http.StatusInternalServerError
			}
			
			c.JSON(httpStatus, gin.H{
				"status":    healthStatus.Status,
				"timestamp": healthStatus.Timestamp,
				"metrics": gin.H{
					"error_rate":      healthStatus.ErrorRate,
					"critical_errors": healthStatus.CriticalErrors,
					"high_errors":     healthStatus.HighErrors,
					"last_error_time": healthStatus.LastErrorTime,
				},
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// sendErrorResponse sends a standardized error response
func sendErrorResponse(c *gin.Context, appErr *errors.AppError) {
	errorResponse := errors.NewErrorResponse(appErr).
		WithRequestInfo(getRequestID(c), c.Request.URL.Path, c.Request.Method)
	
	c.JSON(appErr.HTTPStatus, errorResponse)
}

// getRequestID extracts request ID from gin context
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// AbortWithError aborts the request with an AppError
func AbortWithError(c *gin.Context, appErr *errors.AppError) {
	appErr = appErr.WithRequestID(getRequestID(c))
	
	// Log the error
	logger := errors.NewLogger().WithRequestID(getRequestID(c))
	logger.Error(appErr, "Request aborted with error")
	
	// Record error metrics
	monitoring.RecordGlobalError(appErr)
	
	// Send error response
	sendErrorResponse(c, appErr)
	c.Abort()
}

// HandleError handles an error in a gin handler
func HandleError(c *gin.Context, err error, message string) {
	var appErr *errors.AppError
	
	// Check if it's already an AppError
	if appError, ok := err.(*errors.AppError); ok {
		appErr = appError
	} else {
		// Convert to AppError
		appErr = errors.NewInternalError(err)
	}
	
	appErr = appErr.WithRequestID(getRequestID(c))
	
	// Log the error
	logger := errors.NewLogger().WithRequestID(getRequestID(c))
	logger.Error(appErr, message)
	
	// Record error metrics
	monitoring.RecordGlobalError(appErr)
	
	// Send error response
	sendErrorResponse(c, appErr)
}