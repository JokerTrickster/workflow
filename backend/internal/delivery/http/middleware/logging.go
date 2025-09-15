package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// RequestIDKey is the context key for request ID
type RequestIDKey string

const (
	RequestID RequestIDKey = "request_id"
	StartTime RequestIDKey = "start_time"
)

// RequestLoggingMiddleware provides structured logging for all HTTP requests
func RequestLoggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			requestID := uuid.New().String()

			// Add request ID and start time to context
			ctx := context.WithValue(c.Request().Context(), RequestID, requestID)
			ctx = context.WithValue(ctx, StartTime, start)
			ctx = context.WithValue(ctx, "timestamp", start.Format(time.RFC3339))
			c.SetRequest(c.Request().WithContext(ctx))

			// Add request ID to response headers
			c.Response().Header().Set("X-Request-ID", requestID)

			// Capture request details
			req := c.Request()
			method := req.Method
			path := req.URL.Path
			query := req.URL.RawQuery
			userAgent := req.UserAgent()
			remoteAddr := c.RealIP()

			// Capture request body for logging (only for non-GET requests)
			var requestBody string
			if method != "GET" && method != "HEAD" {
				if body, err := io.ReadAll(req.Body); err == nil {
					requestBody = string(body)
					// Restore the request body for the handler
					req.Body = io.NopCloser(bytes.NewBuffer(body))
				}
			}

			// Log request start
			log.Printf("🌐 [%s] %s %s - Remote: %s, User-Agent: %s, Query: %s", 
				requestID, method, path, remoteAddr, userAgent, query)

			if requestBody != "" && len(requestBody) < 1000 { // Only log small request bodies
				log.Printf("📥 [%s] Request Body: %s", requestID, requestBody)
			}

			// Execute the handler
			err := next(c)

			// Calculate response time
			duration := time.Since(start)
			status := c.Response().Status
			size := c.Response().Size

			// Log response
			if err != nil {
				log.Printf("❌ [%s] %s %s - Status: %d, Duration: %v, Size: %d bytes, Error: %v", 
					requestID, method, path, status, duration, size, err)
			} else {
				log.Printf("✅ [%s] %s %s - Status: %d, Duration: %v, Size: %d bytes", 
					requestID, method, path, status, duration, size)
			}

			// Log slow requests (> 1 second)
			if duration > time.Second {
				log.Printf("🐌 [%s] SLOW REQUEST - %s %s took %v", requestID, method, path, duration)
			}

			return err
		}
	}
}

// ResponseBodyLoggingMiddleware logs response bodies for debugging (use carefully in production)
func ResponseBodyLoggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Capture response body
			resBody := new(bytes.Buffer)
			mw := io.MultiWriter(c.Response().Writer, resBody)
			writer := &responseWriter{c.Response().Writer, mw}
			c.Response().Writer = writer

			err := next(c)

			// Log response body for debugging (only for errors or specific content types)
			if c.Response().Status >= 400 || c.Request().Header.Get("X-Debug-Response") == "true" {
				requestID := getRequestID(c)
				responseBody := resBody.String()
				
				if len(responseBody) < 2000 { // Only log reasonable-sized responses
					log.Printf("📤 [%s] Response Body: %s", requestID, responseBody)
				} else {
					log.Printf("📤 [%s] Response Body: <large response %d bytes>", requestID, len(responseBody))
				}
			}

			return err
		}
	}
}

// MetricsLoggingMiddleware logs metrics and performance data
func MetricsLoggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			
			err := next(c)
			
			duration := time.Since(start)
			requestID := getRequestID(c)
			
			// Log performance metrics
			metrics := map[string]interface{}{
				"request_id":       requestID,
				"method":          c.Request().Method,
				"path":            c.Request().URL.Path,
				"status":          c.Response().Status,
				"duration_ms":     duration.Milliseconds(),
				"response_size":   c.Response().Size,
				"remote_addr":     c.RealIP(),
				"user_agent":      c.Request().UserAgent(),
				"timestamp":       start.Format(time.RFC3339),
			}
			
			// Convert to JSON for structured logging
			if metricsJSON, err := json.Marshal(metrics); err == nil {
				log.Printf("📊 METRICS: %s", string(metricsJSON))
			}
			
			return err
		}
	}
}

// Helper types and functions

type responseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (r *responseWriter) Write(b []byte) (int, error) {
	return r.Writer.Write(b)
}

// getRequestID extracts request ID from context
func getRequestID(c echo.Context) string {
	if requestID := c.Request().Context().Value(RequestID); requestID != nil {
		return requestID.(string)
	}
	return "unknown"
}

// GetRequestIDFromContext extracts request ID from context (utility function)
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID := ctx.Value(RequestID); requestID != nil {
		return requestID.(string)
	}
	return "unknown"
}