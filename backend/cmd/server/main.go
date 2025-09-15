package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"

	"ai-git-workbench/internal/application/container"
	"ai-git-workbench/internal/delivery/http/routes"
)

func main() {
	log.Printf("🚀 Starting Task Queue API Server...")

	// Initialize application container with all dependencies
	appContainer, err := container.NewApplicationContainer()
	if err != nil {
		log.Fatalf("❌ Failed to initialize application container: %v", err)
	}
	defer func() {
		log.Printf("🧹 Closing application container...")
		if err := appContainer.Close(); err != nil {
			log.Printf("❌ Error closing application container: %v", err)
		}
	}()

	// Perform initial health check
	if err := appContainer.HealthCheck(); err != nil {
		log.Fatalf("❌ Application health check failed: %v", err)
	}
	log.Printf("✅ Application container initialized and healthy")

	// Initialize Echo server
	e := echo.New()

	// Configure Echo settings
	e.HideBanner = true
	e.HidePort = true

	// Setup routes with middleware
	setupRoutes(e, appContainer)

	// Get server configuration
	port := getServerPort(appContainer)
	address := ":" + port

	// Start server in goroutine for graceful shutdown
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("🌐 Server starting on port %s", port)
		log.Printf("📋 API Base URL: http://localhost:%s/api", port)
		log.Printf("🏥 Health Check: http://localhost:%s/health", port)
		
		if err := e.Start(address); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Setup graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until shutdown signal or server error
	select {
	case sig := <-shutdown:
		log.Printf("🛑 Shutdown signal received: %v", sig)
	case err := <-serverErrors:
		log.Printf("❌ Server error: %v", err)
	}

	// Perform graceful shutdown
	performGracefulShutdown(e, appContainer)
}

// setupRoutes configures all routes and middleware
func setupRoutes(e *echo.Echo, appContainer *container.ApplicationContainer) {
	// Determine environment
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	// Setup routes based on environment
	switch environment {
	case "production":
		allowedOrigins := getProductionOrigins()
		routes.SetupProductionRoutes(e, appContainer, allowedOrigins)
		routes.SetupAPIRoutes(e, appContainer)
	case "development":
		routes.SetupAPIRoutes(e, appContainer)
		routes.SetupDevelopmentRoutes(e, appContainer)
	default:
		routes.SetupAPIRoutes(e, appContainer)
	}

	log.Printf("🛣️ Routes configured for %s environment", environment)
}

// getServerPort gets the server port from configuration
func getServerPort(appContainer *container.ApplicationContainer) string {
	// Try environment variable first
	if port := os.Getenv("PORT"); port != "" {
		return port
	}

	// Try configuration
	if appContainer.Config != nil && appContainer.Config.Server.Port != "" {
		return appContainer.Config.Server.Port
	}

	// Default port
	return "8080"
}

// getProductionOrigins returns allowed origins for production
func getProductionOrigins() []string {
	// Get from environment or configuration
	if origins := os.Getenv("ALLOWED_ORIGINS"); origins != "" {
		// Parse comma-separated origins
		return parseOrigins(origins)
	}

	// Default production origins
	return []string{
		"https://app.example.com",
		"https://admin.example.com",
	}
}

// parseOrigins parses comma-separated origins
func parseOrigins(origins string) []string {
	// Simple implementation - in production, use proper CSV parsing
	result := []string{}
	for _, origin := range splitString(origins, ",") {
		if trimmed := trimSpace(origin); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// performGracefulShutdown handles graceful server shutdown
func performGracefulShutdown(e *echo.Echo, appContainer *container.ApplicationContainer) {
	log.Printf("🔄 Starting graceful shutdown...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown Echo server
	if err := e.Shutdown(ctx); err != nil {
		log.Printf("❌ Error during server shutdown: %v", err)
	} else {
		log.Printf("✅ Server shutdown completed")
	}

	// Close application container (database connections, etc.)
	if err := appContainer.Close(); err != nil {
		log.Printf("❌ Error closing application resources: %v", err)
	} else {
		log.Printf("✅ Application resources closed")
	}

	log.Printf("✅ Graceful shutdown completed")
}

// Helper functions

// splitString splits a string by delimiter (simple implementation)
func splitString(s, delimiter string) []string {
	if s == "" {
		return []string{}
	}
	
	var result []string
	current := ""
	
	for i, char := range s {
		if string(char) == delimiter {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
		
		// Add last part
		if i == len(s)-1 && current != "" {
			result = append(result, current)
		}
	}
	
	return result
}

// trimSpace removes leading and trailing whitespace
func trimSpace(s string) string {
	start := 0
	end := len(s)
	
	// Find first non-space character
	for start < end && s[start] == ' ' {
		start++
	}
	
	// Find last non-space character
	for end > start && s[end-1] == ' ' {
		end--
	}
	
	return s[start:end]
}