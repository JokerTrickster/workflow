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
	"github.com/labstack/echo/v4/middleware"

	"ai-git-workbench/internal/delivery/http/handlers"
	"ai-git-workbench/internal/infrastructure/container"
)

func main() {
	// Initialize dependency container
	appContainer, err := container.NewContainer()
	if err != nil {
		log.Fatalf("❌ Failed to initialize container: %v", err)
	}
	defer func() {
		if err := appContainer.Close(); err != nil {
			log.Printf("❌ Error closing container: %v", err)
		}
	}()

	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Initialize handlers
	taskHandler := handlers.NewTaskHandler(appContainer.TaskRepository)
	healthHandler := handlers.NewHealthHandler()

	// Health check endpoint
	e.GET("/health", healthHandler.Health)

	// API routes group
	api := e.Group("/api/v1")
	{
		// Task routes
		tasks := api.Group("/tasks")
		{
			tasks.GET("", taskHandler.GetTasks)
			tasks.POST("", taskHandler.CreateTask)
			tasks.GET("/:id", taskHandler.GetTask)
			tasks.PUT("/:id", taskHandler.UpdateTask)
			tasks.DELETE("/:id", taskHandler.DeleteTask)
		}
	}

	// Start server
	port := appContainer.Config.Server.Port
	if port == "" {
		port = "8080"
	}

	// Graceful shutdown
	go func() {
		log.Printf("🚀 Server starting on port %s", port)
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown Echo server
	if err := e.Shutdown(ctx); err != nil {
		log.Printf("❌ Error during server shutdown: %v", err)
	}

	log.Println("✅ Server shutdown completed")
}