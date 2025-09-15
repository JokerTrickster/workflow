package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Simple standalone HTTP server to test our API patterns

// StandardResponse represents the standard API response format
type StandardResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// ErrorInfo represents error information in API responses
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// RequestIDKey is the context key for request ID
type RequestIDKey string

const RequestID RequestIDKey = "request_id"

func main() {
	log.Printf("🚀 Starting Test HTTP Server...")

	e := echo.New()
	e.HideBanner = true

	// Basic middleware
	e.Use(requestLoggingMiddleware())
	e.Use(corsMiddleware())
	e.Use(recoveryMiddleware())

	// Routes
	e.GET("/health", healthCheck)
	e.GET("/api/ping", ping)
	e.POST("/api/tasks", createTask)
	e.GET("/api/tasks/:id", getTask)

	port := "9090"
	log.Printf("🌐 Server starting on port %s", port)
	log.Printf("🏥 Health: http://localhost:%s/health", port)
	log.Printf("📋 API: http://localhost:%s/api/ping", port)

	if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Server error: %v", err)
	}
}

// Middleware functions

func requestLoggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			requestID := uuid.New().String()

			// Add request ID to context
			ctx := context.WithValue(c.Request().Context(), RequestID, requestID)
			ctx = context.WithValue(ctx, "timestamp", start.Format(time.RFC3339))
			c.SetRequest(c.Request().WithContext(ctx))

			// Add request ID to response header
			c.Response().Header().Set("X-Request-ID", requestID)

			// Log request
			log.Printf("🌐 [%s] %s %s", requestID, c.Request().Method, c.Request().URL.Path)

			err := next(c)

			// Log response
			duration := time.Since(start)
			status := c.Response().Status
			log.Printf("✅ [%s] %s %s - %d (%v)", requestID, c.Request().Method, c.Request().URL.Path, status, duration)

			return err
		}
	}
}

func corsMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Access-Control-Allow-Origin", "*")
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if c.Request().Method == "OPTIONS" {
				return c.NoContent(http.StatusNoContent)
			}

			return next(c)
		}
	}
}

func recoveryMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("🚨 PANIC: %v", r)
					respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred", "")
				}
			}()
			return next(c)
		}
	}
}

// Handler functions

func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "task-queue-api",
	})
}

func ping(c echo.Context) error {
	return respondSuccess(c, http.StatusOK, map[string]interface{}{
		"message": "pong",
		"version": "1.0.0",
	})
}

func createTask(c echo.Context) error {
	// Mock task creation
	var req map[string]interface{}
	if err := c.Bind(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON format", err.Error())
	}

	// Validate required fields
	if req["title"] == nil || req["user_id"] == nil {
		return respondError(c, http.StatusBadRequest, "MISSING_FIELDS", "title and user_id are required", "")
	}

	// Mock response
	response := map[string]interface{}{
		"task_id":    uuid.New().String(),
		"status":     "pending",
		"created_at": time.Now().Format(time.RFC3339),
		"title":      req["title"],
		"user_id":    req["user_id"],
	}

	return respondSuccess(c, http.StatusCreated, response)
}

func getTask(c echo.Context) error {
	taskID := c.Param("id")
	if taskID == "" {
		return respondError(c, http.StatusBadRequest, "MISSING_TASK_ID", "Task ID is required", "")
	}

	// Mock task response
	response := map[string]interface{}{
		"task_id":     taskID,
		"user_id":     "user-123",
		"title":       "Sample Task",
		"description": "This is a sample task",
		"status":      "pending",
		"repository":  "example/repo",
		"created_at":  time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		"updated_at":  time.Now().Format(time.RFC3339),
	}

	return respondSuccess(c, http.StatusOK, response)
}

// Helper functions

func respondSuccess(c echo.Context, status int, data interface{}) error {
	response := StandardResponse{
		Success:   true,
		Data:      data,
		Timestamp: getCurrentTimestamp(c),
	}
	return c.JSON(status, response)
}

func respondError(c echo.Context, status int, code, message, details string) error {
	response := StandardResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: getCurrentTimestamp(c),
	}
	return c.JSON(status, response)
}

func getCurrentTimestamp(c echo.Context) string {
	if timestamp := c.Request().Context().Value("timestamp"); timestamp != nil {
		return timestamp.(string)
	}
	return time.Now().Format(time.RFC3339)
}

func getRequestID(c echo.Context) string {
	if requestID := c.Request().Context().Value(RequestID); requestID != nil {
		return requestID.(string)
	}
	return "unknown"
}