package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
	
	"ai-git-workbench/internal/delivery/http/handlers"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(e *echo.Echo) {
	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	taskHandler := handlers.NewTaskHandler()
	repositoryHandler := handlers.NewRepositoryHandler()

	// API versioning group
	v1 := e.Group("/api/v1")

	// Health endpoints
	v1.GET("/health", healthHandler.HealthCheck)
	v1.GET("/ping", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "pong",
			"status":  "OK",
		})
	})

	// Task endpoints
	taskGroup := v1.Group("/tasks")
	{
		taskGroup.GET("", taskHandler.GetTasks)
		taskGroup.GET("/:id", taskHandler.GetTask)
		taskGroup.POST("", taskHandler.CreateTask)
		taskGroup.PUT("/:id", taskHandler.UpdateTask)
		taskGroup.DELETE("/:id", taskHandler.DeleteTask)
	}

	// Repository endpoints
	repoGroup := v1.Group("/repositories")
	{
		repoGroup.GET("", repositoryHandler.GetRepositories)
		repoGroup.GET("/:id", repositoryHandler.GetRepository)
		repoGroup.POST("", repositoryHandler.CreateRepository)
		repoGroup.PUT("/:id", repositoryHandler.UpdateRepository)
		repoGroup.DELETE("/:id", repositoryHandler.DeleteRepository)
	}

	// GitHub integration endpoints
	githubGroup := v1.Group("/github")
	{
		githubGroup.POST("/webhook", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{
				"message": "GitHub webhook received",
				"status":  "OK",
			})
		})
		githubGroup.GET("/repos", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"message": "GitHub repositories",
				"repos":   []string{},
			})
		})
	}

	// Workflow endpoints
	workflowGroup := v1.Group("/workflows")
	{
		workflowGroup.GET("", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"message":   "Workflows list",
				"workflows": []string{},
			})
		})
		workflowGroup.POST("", func(c echo.Context) error {
			return c.JSON(http.StatusCreated, map[string]string{
				"message": "Workflow created",
				"status":  "OK",
			})
		})
	}
}