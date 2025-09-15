package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/labstack/echo/v4"

	appErrors "ai-git-workbench/internal/application/errors"
)

// ErrorResponse represents the standard error response format
type ErrorResponse struct {
	Success   bool       `json:"success"`
	Error     *ErrorInfo `json:"error"`
	Timestamp string     `json:"timestamp"`
}

// ErrorInfo represents error information in responses
type ErrorInfo struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// ErrorHandlingMiddleware provides centralized error handling for the application
func ErrorHandlingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				return handleError(c, err)
			}
			return nil
		}
	}
}

// RecoveryMiddleware provides panic recovery with structured error response
func RecoveryMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					requestID := getRequestID(c)
					
					// Log the panic with stack trace
					log.Printf("🚨 [%s] PANIC RECOVERED: %v", requestID, r)
					log.Printf("🚨 [%s] Stack trace:\n%s", requestID, string(debug.Stack()))
					
					// Return structured error response
					response := ErrorResponse{
						Success: false,
						Error: &ErrorInfo{
							Code:      "INTERNAL_SERVER_ERROR",
							Message:   "An unexpected error occurred",
							Details:   "Internal server error - please contact support",
							RequestID: requestID,
						},
						Timestamp: getCurrentTimestamp(c),
					}
					
					if !c.Response().Committed {
						c.JSON(http.StatusInternalServerError, response)
					}
				}
			}()
			
			return next(c)
		}
	}
}

// handleError centralizes error handling logic
func handleError(c echo.Context, err error) error {
	requestID := getRequestID(c)
	
	// Log the error
	log.Printf("❌ [%s] Error: %v", requestID, err)
	
	// Don't handle errors that are already HTTP errors from Echo
	if he, ok := err.(*echo.HTTPError); ok {
		// Convert Echo HTTP error to our standard format
		code := getErrorCodeFromStatus(he.Code)
		message := getErrorMessage(he.Message)
		
		response := ErrorResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:      code,
				Message:   message,
				RequestID: requestID,
			},
			Timestamp: getCurrentTimestamp(c),
		}
		
		return c.JSON(he.Code, response)
	}
	
	// Handle application errors
	if appErr, ok := err.(appErrors.ApplicationError); ok {
		response := ErrorResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:      appErr.Code,
				Message:   appErr.Message,
				Details:   getErrorDetails(appErr.Cause),
				RequestID: requestID,
			},
			Timestamp: getCurrentTimestamp(c),
		}
		
		return c.JSON(appErr.HTTPStatus, response)
	}
	
	// Handle domain errors through application error translation
	if translatedErr := appErrors.TranslateDomainError(err); translatedErr.Code != "" {
		response := ErrorResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:      translatedErr.Code,
				Message:   translatedErr.Message,
				Details:   getErrorDetails(translatedErr.Cause),
				RequestID: requestID,
			},
			Timestamp: getCurrentTimestamp(c),
		}
		
		return c.JSON(translatedErr.HTTPStatus, response)
	}
	
	// Default error handling
	log.Printf("🚨 [%s] Unhandled error type: %T - %v", requestID, err, err)
	
	response := ErrorResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:      "INTERNAL_SERVER_ERROR",
			Message:   "An internal error occurred",
			Details:   "Please contact support if this persists",
			RequestID: requestID,
		},
		Timestamp: getCurrentTimestamp(c),
	}
	
	return c.JSON(http.StatusInternalServerError, response)
}

// ValidationErrorMiddleware handles validation errors specifically
func ValidationErrorMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			
			if err != nil && appErrors.IsValidationError(err) {
				requestID := getRequestID(c)
				
				response := ErrorResponse{
					Success: false,
					Error: &ErrorInfo{
						Code:      "VALIDATION_ERROR",
						Message:   "Request validation failed",
						Details:   err.Error(),
						RequestID: requestID,
					},
					Timestamp: getCurrentTimestamp(c),
				}
				
				return c.JSON(http.StatusBadRequest, response)
			}
			
			return err
		}
	}
}

// Helper functions

// getErrorCodeFromStatus maps HTTP status codes to error codes
func getErrorCodeFromStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusMethodNotAllowed:
		return "METHOD_NOT_ALLOWED"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusUnprocessableEntity:
		return "UNPROCESSABLE_ENTITY"
	case http.StatusTooManyRequests:
		return "TOO_MANY_REQUESTS"
	case http.StatusInternalServerError:
		return "INTERNAL_SERVER_ERROR"
	case http.StatusBadGateway:
		return "BAD_GATEWAY"
	case http.StatusServiceUnavailable:
		return "SERVICE_UNAVAILABLE"
	case http.StatusGatewayTimeout:
		return "GATEWAY_TIMEOUT"
	default:
		return "UNKNOWN_ERROR"
	}
}

// getErrorMessage extracts error message from various error types
func getErrorMessage(message interface{}) string {
	switch msg := message.(type) {
	case string:
		return msg
	case error:
		return msg.Error()
	default:
		return "An error occurred"
	}
}

// getErrorDetails extracts details from an error cause
func getErrorDetails(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// getCurrentTimestamp returns current timestamp from context or generates new one
func getCurrentTimestamp(c echo.Context) string {
	if timestamp := c.Request().Context().Value("timestamp"); timestamp != nil {
		return timestamp.(string)
	}
	return getCurrentTimestampDefault()
}

// getCurrentTimestampDefault generates current timestamp in RFC3339 format
func getCurrentTimestampDefault() string {
	return getCurrentTimestampRFC3339()
}