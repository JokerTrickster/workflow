package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"local-backend-server/internal/infrastructure/config"
	"local-backend-server/internal/usecase"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting local-backend-server on %s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Println("Configuration loaded successfully")

	// Create application context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize dependency injection container
	container := usecase.NewContainer(cfg)
	if err := container.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}
	defer func() {
		if err := container.Cleanup(); err != nil {
			log.Printf("Failed to cleanup container: %v", err)
		}
	}()

	// Start workflow orchestrator consumer
	orchestrator := container.GetWorkflowOrchestrator()
	go func() {
		log.Println("Starting workflow orchestrator consumer...")
		if err := orchestrator.StartConsumer(ctx); err != nil {
			log.Printf("Workflow orchestrator consumer failed: %v", err)
			cancel()
		}
	}()

	// Create HTTP server with health endpoints
	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		health := container.HealthCheck(ctx)
		w.Header().Set("Content-Type", "application/json")
		
		if health["status"] == "healthy" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		
		// Simple JSON response without external dependencies
		fmt.Fprintf(w, `{"status":"%s","timestamp":"%s"}`, 
			health["status"], time.Now().Format(time.RFC3339))
	})

	// Metrics endpoint
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := orchestrator.GetMetrics(ctx)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		// Simple JSON response
		fmt.Fprintf(w, `{"total_requests":%v,"completed_requests":%v,"failed_requests":%v,"success_rate":%v}`,
			metrics["total_requests"], metrics["completed_requests"], 
			metrics["failed_requests"], metrics["success_rate"])
	})

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: mux,
	}

	// Start HTTP server in goroutine
	go func() {
		log.Printf("Starting HTTP server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server failed: %v", err)
			cancel()
		}
	}()

	log.Println("Server initialization complete")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down...", sig)
	case <-ctx.Done():
		log.Println("Context cancelled, shutting down...")
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop workflow orchestrator
	if err := orchestrator.StopConsumer(); err != nil {
		log.Printf("Failed to stop workflow orchestrator: %v", err)
	}

	// Stop HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown failed: %v", err)
	}

	log.Println("Server shutdown complete")
}