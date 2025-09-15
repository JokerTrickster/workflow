package main

import (
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"ai-git-workbench/internal/delivery/http/handlers"
	"ai-git-workbench/internal/delivery/http/middleware"
)

func main() {
	log.Printf("🚀 Starting Task Queue API Server (Simple Mode)...")

	// Initialize Echo server
	e := echo.New()

	// Configure Echo settings
	e.HideBanner = true
	e.HidePort = true

	// Setup middleware
	setupSimpleMiddleware(e)

	// Setup basic routes
	setupSimpleRoutes(e)

	// Start server
	port := "8080"
	log.Printf("🌐 Server starting on port %s", port)
	log.Printf("🏥 Health Check: http://localhost:%s/health", port)
	
	if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Server error: %v", err)
	}
}

// setupSimpleMiddleware configures basic middleware
func setupSimpleMiddleware(e *echo.Echo) {
	// Security headers
	e.Use(middleware.SecurityHeadersMiddleware())

	// CORS
	corsConfig := middleware.DefaultCORSConfig()
	e.Use(middleware.CORSMiddleware(corsConfig))

	// Request logging
	e.Use(middleware.RequestLoggingMiddleware())

	// Recovery
	e.Use(middleware.RecoveryMiddleware())

	// Basic validation
	validationConfig := middleware.DefaultValidationConfig()
	e.Use(middleware.RequestValidationMiddleware(validationConfig))
}

// setupSimpleRoutes configures basic routes for testing
func setupSimpleRoutes(e *echo.Echo) {
	// Simple health handler without container dependency
	healthHandler := &handlers.SimpleHealthHandler{}

	// Health routes
	e.GET("/health", healthHandler.Health)
	e.GET("/live", healthHandler.Live)
	e.GET("/ready", healthHandler.Ready)

	// API group
	api := e.Group("/api")
	api.GET("/ping", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":   "pong",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "task-queue-api",
		})
	})
}