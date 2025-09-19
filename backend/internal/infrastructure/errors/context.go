package errors

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	requestIDKey contextKey = "request_id"
)

// SetRequestID sets the request ID in the context
func SetRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID gets the request ID from the context
func GetRequestID(r *http.Request) string {
	if r == nil {
		return ""
	}
	
	ctx := r.Context()
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	
	// Try to get from header as fallback
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		return requestID
	}
	
	return ""
}

// GetRequestIDFromContext gets the request ID from context directly
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// Generate random suffix
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to time-based ID if random generation fails
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405")))
	}
	
	return hex.EncodeToString(bytes)
}

// AsAppError checks if an error is an AppError and returns it
func AsAppError(err error, target **AppError) bool {
	if err == nil {
		return false
	}
	
	if appErr, ok := err.(*AppError); ok {
		*target = appErr
		return true
	}
	
	return false
}