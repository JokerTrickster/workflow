package routes

import (
	"time"

	"github.com/labstack/echo/v4"

	"ai-git-workbench/internal/application/container"
	"ai-git-workbench/internal/delivery/http/handlers"
	"ai-git-workbench/internal/delivery/http/middleware"
)

// SetupAPIRoutes configures all API routes with proper middleware
func SetupAPIRoutes(e *echo.Echo, appContainer *container.ApplicationContainer) {
	// Initialize handlers
	taskHandler := handlers.NewTaskHandler(appContainer.GetTaskUsecase())
	healthHandler := handlers.NewHealthHandler(appContainer)

	// Global middleware
	setupGlobalMiddleware(e)

	// Health check routes (before API middleware)
	setupHealthRoutes(e, healthHandler)

	// API routes with middleware
	api := e.Group("/api")
	setupAPIMiddleware(api)

	// Setup route groups
	setupTaskRoutes(api, taskHandler)
	setupUserRoutes(api, taskHandler)
	setupQueueRoutes(api, taskHandler)
}

// setupGlobalMiddleware configures global middleware for all requests
func setupGlobalMiddleware(e *echo.Echo) {
	// Security headers (first)
	e.Use(middleware.SecurityHeadersMiddleware())

	// CORS (early in the chain)
	corsConfig := middleware.DefaultCORSConfig()
	e.Use(middleware.CORSMiddleware(corsConfig))

	// Request logging with ID generation
	e.Use(middleware.RequestLoggingMiddleware())

	// Recovery from panics
	e.Use(middleware.RecoveryMiddleware())

	// Request validation
	validationConfig := middleware.DefaultValidationConfig()
	e.Use(middleware.RequestValidationMiddleware(validationConfig))

	// Request size limiting
	e.Use(middleware.RequestSizeLimitMiddleware(10 * 1024 * 1024)) // 10MB limit

	// Request timeout
	e.Use(middleware.TimeoutMiddleware(30 * time.Second))

	// Response compression
	e.Use(middleware.CompressResponseMiddleware())

	// Rate limiting
	e.Use(middleware.RateLimitingMiddleware())
}

// setupAPIMiddleware configures middleware specific to API routes
func setupAPIMiddleware(api *echo.Group) {
	// JSON validation for API requests
	api.Use(middleware.JSONValidationMiddleware())

	// Parameter validation
	api.Use(middleware.ParameterValidationMiddleware())

	// Security validation
	api.Use(middleware.SecurityValidationMiddleware())

	// Error handling (last middleware)
	api.Use(middleware.ErrorHandlingMiddleware())

	// Validation error handling
	api.Use(middleware.ValidationErrorMiddleware())

	// Metrics logging
	api.Use(middleware.MetricsLoggingMiddleware())
}

// setupHealthRoutes configures health check routes
func setupHealthRoutes(e *echo.Echo, healthHandler *handlers.HealthHandler) {
	// Simple health check (no middleware)
	e.GET("/health", healthHandler.Health)
	e.GET("/live", healthHandler.Live)
	e.GET("/ready", healthHandler.Ready)

	// Detailed health check
	e.GET("/health/detailed", healthHandler.HealthDetailed)
	e.GET("/metrics", healthHandler.Metrics)
}

// setupTaskRoutes configures task-related routes
func setupTaskRoutes(api *echo.Group, taskHandler *handlers.TaskHandler) {
	tasks := api.Group("/tasks")

	// Task CRUD operations
	tasks.POST("", taskHandler.CreateTask)                    // POST /api/tasks
	tasks.GET("/:id", taskHandler.GetTask)                   // GET /api/tasks/{id}
	tasks.GET("", taskHandler.ListTasks)                     // GET /api/tasks
	tasks.PUT("/:id", taskHandler.UpdateTask)                // PUT /api/tasks/{id}
	tasks.DELETE("/:id", taskHandler.DeleteTask)             // DELETE /api/tasks/{id}

	// Task operations
	tasks.PUT("/:id/resume", taskHandler.ResumeTask)         // PUT /api/tasks/{id}/resume

	// Task monitoring
	tasks.GET("/:id/health", taskHandler.GetTaskHealth)     // GET /api/tasks/{id}/health
}

// setupUserRoutes configures user-related routes
func setupUserRoutes(api *echo.Group, taskHandler *handlers.TaskHandler) {
	users := api.Group("/users")

	// User statistics
	users.GET("/:id/stats", taskHandler.GetUserTaskStatistics) // GET /api/users/{id}/stats
}

// setupQueueRoutes configures queue-related routes
func setupQueueRoutes(api *echo.Group, taskHandler *handlers.TaskHandler) {
	queue := api.Group("/queue")

	// Queue statistics
	queue.GET("/stats", taskHandler.GetQueueStatistics) // GET /api/queue/stats
}

// SetupDevelopmentRoutes configures additional routes for development
func SetupDevelopmentRoutes(e *echo.Echo, appContainer *container.ApplicationContainer) {
	dev := e.Group("/dev")

	// Debug endpoints (only in development)
	dev.GET("/config", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"environment": "development",
			"debug":       true,
			"timestamp":   time.Now().Format(time.RFC3339),
		})
	})

	// Test error endpoint
	dev.GET("/error", func(c echo.Context) error {
		return echo.NewHTTPError(500, "Test error endpoint")
	})

	// Test panic endpoint
	dev.GET("/panic", func(c echo.Context) error {
		panic("Test panic endpoint")
	})
}

// SetupProductionRoutes configures routes for production environment
func SetupProductionRoutes(e *echo.Echo, appContainer *container.ApplicationContainer, allowedOrigins []string) {
	// Production CORS configuration
	corsConfig := middleware.ProductionCORSConfig(allowedOrigins)
	
	// Replace default CORS middleware with production config
	// Note: This should be called instead of setupGlobalMiddleware for production
	e.Use(middleware.SecurityHeadersMiddleware())
	e.Use(middleware.CORSMiddleware(corsConfig))
	e.Use(middleware.RequestLoggingMiddleware())
	e.Use(middleware.RecoveryMiddleware())
}