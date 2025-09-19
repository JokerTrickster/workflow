package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"local-backend-server/internal/infrastructure/logging"
)

// ErrorResponse represents the JSON error response structure
type ErrorResponse struct {
	Error     ErrorInfo `json:"error"`
	RequestID string    `json:"request_id,omitempty"`
	Timestamp string    `json:"timestamp"`
}

// ErrorInfo contains the error details
type ErrorInfo struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

// ErrorHandler handles HTTP errors and converts them to appropriate responses
type ErrorHandler struct {
	logger      *logging.Logger
	development bool
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *logging.Logger, development bool) *ErrorHandler {
	return &ErrorHandler{
		logger:      logger,
		development: development,
	}
}

// HandleError handles an error and writes the appropriate HTTP response
func (h *ErrorHandler) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	// Extract request ID from context if available
	requestID := GetRequestID(r)

	// Convert to AppError if not already
	var appErr *AppError
	if !AsAppError(err, &appErr) {
		appErr = InternalError("An unexpected error occurred").
			WithRequestID(requestID).
			WithCause(err)
	}

	// Log the error
	h.logError(r, appErr)

	// Create error response
	response := ErrorResponse{
		Error: ErrorInfo{
			Code:    appErr.Code,
			Message: appErr.Message,
		},
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Include details in development mode or for specific error types
	if h.development || h.shouldIncludeDetails(appErr) {
		response.Error.Details = appErr.Details
	}

	// Set appropriate status code
	statusCode := appErr.HTTPStatusCode()

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.WithError(err).Error("Failed to encode error response")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// PanicRecoveryMiddleware recovers from panics and converts them to errors
func (h *ErrorHandler) PanicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				// Log the panic with stack trace
				requestID := GetRequestID(r)
				stack := string(debug.Stack())
				
				h.logger.WithFields(logging.Fields{
					"request_id": requestID,
					"method":     r.Method,
					"path":       r.URL.Path,
					"panic":      rvr,
					"stack":      stack,
				}).Error("Panic recovered")

				// Convert panic to error
				err := InternalError(fmt.Sprintf("Panic recovered: %v", rvr)).
					WithRequestID(requestID).
					WithComponent("panic_recovery").
					WithMetadata("stack", stack)

				h.HandleError(w, r, err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// RequestIDMiddleware adds request ID to context
func (h *ErrorHandler) RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := generateRequestID()
		r = r.WithContext(SetRequestID(r.Context(), requestID))
		
		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs HTTP requests and responses
func (h *ErrorHandler) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := GetRequestID(r)

		// Log request
		h.logger.WithFields(logging.Fields{
			"request_id":    requestID,
			"method":        r.Method,
			"path":          r.URL.Path,
			"query":         r.URL.RawQuery,
			"remote_addr":   r.RemoteAddr,
			"user_agent":    r.UserAgent(),
			"content_length": r.ContentLength,
		}).Info("HTTP request started")

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Log response
		duration := time.Since(start)
		h.logger.WithFields(logging.Fields{
			"request_id":   requestID,
			"method":       r.Method,
			"path":         r.URL.Path,
			"status_code":  wrapped.statusCode,
			"duration_ms":  duration.Milliseconds(),
			"response_size": wrapped.size,
		}).Info("HTTP request completed")
	})
}

// logError logs the error with appropriate level and context
func (h *ErrorHandler) logError(r *http.Request, appErr *AppError) {
	fields := logging.Fields{
		"error_code":  appErr.Code,
		"method":      r.Method,
		"path":        r.URL.Path,
		"request_id":  appErr.RequestID,
		"component":   appErr.Component,
		"retryable":   appErr.Retryable,
	}

	// Add metadata if present
	for k, v := range appErr.Metadata {
		fields[k] = v
	}

	entry := h.logger.WithFields(fields)

	// Include stack trace in development mode for internal errors
	if h.development && appErr.Code == ErrCodeInternal && appErr.Stack != "" {
		entry = entry.WithField("stack", appErr.Stack)
	}

	// Log with appropriate level based on error severity
	switch appErr.Code {
	case ErrCodeInternal:
		entry.WithError(appErr.Cause).Error(appErr.Message)
	case ErrCodeDatabase, ErrCodeQueue, ErrCodeClaude:
		entry.WithError(appErr.Cause).Warn(appErr.Message)
	case ErrCodeTimeout, ErrCodeDatabaseTimeout, ErrCodeQueueTimeout, ErrCodeClaudeTimeout:
		entry.WithError(appErr.Cause).Warn(appErr.Message)
	case ErrCodeValidation, ErrCodeNotFound:
		entry.Info(appErr.Message)
	default:
		entry.WithError(appErr.Cause).Error(appErr.Message)
	}
}

// shouldIncludeDetails determines if error details should be included in response
func (h *ErrorHandler) shouldIncludeDetails(appErr *AppError) bool {
	// Always include details for validation errors
	if appErr.Code == ErrCodeValidation {
		return true
	}

	// Include details for client errors (4xx)
	statusCode := appErr.HTTPStatusCode()
	return statusCode >= 400 && statusCode < 500
}

// responseWriter wraps http.ResponseWriter to capture response data
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}