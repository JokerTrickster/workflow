package main

import (
	"context"
	"fmt"
	"log"
	"main/features"
	"main/utils"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	_ "main/docs"
	_ "main/utils" // Import utils package to trigger init() functions for AI providers

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title Local Backend API
// @version 1.0
// @description This is a local backend server API
// @host localhost:8081
// @BasePath /
func main() {
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, continuing with environment variables")
	}

	// Register AI providers after loading environment variables
	utils.RegisterClaudeProvider()
	utils.RegisterCodexProvider()
	utils.RegisterCursorProvider()

	// 서버 초기화
	if err := utils.InitServer(); err != nil {
		fmt.Printf("서버 초기화 에러 : %v", err.Error())
		return
	}

	// Initialize Task Worker for RabbitMQ consumption
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@localhost:5672/"
	}

	queueName := os.Getenv("RABBITMQ_QUEUE_NAME")
	if queueName == "" {
		queueName = "claude_tasks"
	}

	// Get max concurrent workers from environment (default 3)
	maxConcurrent := 3
	if concurrentStr := os.Getenv("MAX_CONCURRENT_TASKS"); concurrentStr != "" {
		if parsed, err := strconv.Atoi(concurrentStr); err == nil && parsed > 0 {
			maxConcurrent = parsed
		}
	}

	// Start Concurrent Task Worker
	wg.Add(1)
	go func() {
		defer wg.Done()
		startConcurrentTaskWorker(ctx, rabbitMQURL, queueName, maxConcurrent)
	}()

	// Setup Echo server
	e := echo.New()

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Handler 초기화
	if err := features.InitHandler(e); err != nil {
		fmt.Printf("handler 초기화 에러 : %v", err.Error())
		return
	}

	// Swagger 초기화
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Get server configuration from environment variables
	host := os.Getenv("SERVER_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8081"
	}

	// Start HTTP server in goroutine
	address := fmt.Sprintf("%s:%s", host, port)
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Server starting on %s", address)
		if err := e.Start(address); err != nil {
			log.Printf("Server stopped: %v", err)
		}
	}()

	// Wait for signal
	<-sigCh
	log.Println("Received shutdown signal")

	// Cancel context to stop all goroutines
	cancel()

	// Shutdown HTTP server
	if err := e.Shutdown(context.Background()); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	log.Println("Application shutdown complete")
}

// startConcurrentTaskWorker initializes and starts the concurrent RabbitMQ task worker
func startConcurrentTaskWorker(ctx context.Context, rabbitMQURL, queueName string, maxConcurrent int) {
	log.Printf("Starting Concurrent Task Worker with RabbitMQ URL: %s, Queue: %s, Max Concurrent: %d", rabbitMQURL, queueName, maxConcurrent)

	// Create concurrent task worker
	worker := utils.NewConcurrentTaskWorker(rabbitMQURL, queueName, maxConcurrent)

	// Connect to RabbitMQ
	if err := worker.Connect(); err != nil {
		log.Printf("Failed to connect to RabbitMQ: %v", err)
		log.Println("Concurrent Task Worker will not be available")
		return
	}
	defer worker.Close()

	log.Printf("Available AI providers: %v", worker.GetAvailableProviders())
	log.Printf("Provider failure counts: %v", worker.GetFailureCountsReport())

	// Reset failure counts on startup to allow fresh attempts
	worker.ResetAllFailureCounts()
	log.Println("Reset all provider failure counts on startup")

	// Start consuming messages concurrently
	if err := worker.StartConsuming(ctx); err != nil {
		if ctx.Err() != context.Canceled {
			log.Printf("Concurrent Task Worker error: %v", err)
		} else {
			log.Println("Concurrent Task Worker stopped by context cancellation")
		}
	}
}

// Legacy function kept for backward compatibility
func startTaskWorker(ctx context.Context, rabbitMQURL, queueName string) {
	// Default to 1 concurrent task for backward compatibility
	startConcurrentTaskWorker(ctx, rabbitMQURL, queueName, 1)
}
