package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"local-backend-server/internal/infrastructure/config"
	"local-backend-server/internal/infrastructure/logging"
	"local-backend-server/internal/usecase"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Create application context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize dependency injection container
	container := usecase.NewContainer(cfg)
	if err := container.Initialize(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize container: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := container.Cleanup(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to cleanup container: %v\n", err)
		}
	}()

	// Get initialized components
	logger := container.GetLogger()
	errorHandler := container.GetErrorHandler()

	logger.WithComponent("main").WithField("host", cfg.Server.Host).
		WithField("port", cfg.Server.Port).
		WithField("environment", cfg.Environment).
		Info("Local backend server starting")

	// Start workflow orchestrator consumer
	orchestrator := container.GetWorkflowOrchestrator()
	if orchestrator != nil {
		go func() {
			logger.WithComponent("orchestrator").Info("Starting workflow orchestrator consumer")
			if err := orchestrator.StartConsumer(ctx); err != nil {
				logger.WithComponent("orchestrator").WithError(err).Error("Workflow orchestrator consumer failed")
				cancel()
			}
		}()
	} else {
		logger.WithComponent("orchestrator").Warn("Workflow orchestrator not available, consumer not started")
	}

	// Create HTTP server with middleware and endpoints
	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("/health", createHealthHandler(container, logger))

	// Metrics endpoint
	mux.HandleFunc("/metrics", createMetricsHandler(container, logger))

	// Circuit breaker status endpoint
	mux.HandleFunc("/status/circuit-breakers", createCircuitBreakerHandler(container, logger))

	// Apply middleware
	var handler http.Handler = mux
	handler = errorHandler.LoggingMiddleware(handler)
	handler = errorHandler.RequestIDMiddleware(handler)
	handler = errorHandler.PanicRecoveryMiddleware(handler)

	// Create HTTP server with timeouts
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start HTTP server in goroutine
	go func() {
		logger.WithComponent("http").WithField("addr", server.Addr).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithComponent("http").WithError(err).Error("HTTP server failed")
			cancel()
		}
	}()

	logger.WithComponent("main").Info("Server initialization complete")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		logger.WithComponent("main").WithField("signal", sig.String()).Info("Received shutdown signal")
	case <-ctx.Done():
		logger.WithComponent("main").Info("Context cancelled, shutting down")
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	logger.WithComponent("main").Info("Starting graceful shutdown")

	// Stop workflow orchestrator
	if orchestrator != nil {
		if err := orchestrator.StopConsumer(); err != nil {
			logger.WithComponent("orchestrator").WithError(err).Error("Failed to stop workflow orchestrator")
		} else {
			logger.WithComponent("orchestrator").Info("Workflow orchestrator stopped")
		}
	}

	// Stop HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.WithComponent("http").WithError(err).Error("HTTP server shutdown failed")
	} else {
		logger.WithComponent("http").Info("HTTP server stopped")
	}

	logger.WithComponent("main").Info("Server shutdown complete")
}

// createHealthHandler creates the health check endpoint handler
func createHealthHandler(container *usecase.Container, logger *logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		health := container.HealthCheck(ctx)
		
		w.Header().Set("Content-Type", "application/json")
		
		if health["status"] == "healthy" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		
		// Simple JSON response with circuit breaker status
		recoveryManager := container.GetRecoveryManager()
		circuitStats := recoveryManager.GetCircuitBreakerStats()
		
		fmt.Fprintf(w, `{"status":"%s","timestamp":"%s","circuit_breakers":%d}`, 
			health["status"], time.Now().Format(time.RFC3339), len(circuitStats))
	}
}

// createMetricsHandler creates the metrics endpoint handler
func createMetricsHandler(container *usecase.Container, logger *logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		orchestrator := container.GetWorkflowOrchestrator()
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		if orchestrator != nil {
			metrics := orchestrator.GetMetrics(ctx)
			fmt.Fprintf(w, `{"total_requests":%v,"completed_requests":%v,"failed_requests":%v,"success_rate":%v}`,
				metrics["total_requests"], metrics["completed_requests"], 
				metrics["failed_requests"], metrics["success_rate"])
		} else {
			fmt.Fprintf(w, `{"error":"orchestrator not available"}`)
		}
	}
}

// createCircuitBreakerHandler creates the circuit breaker status endpoint
func createCircuitBreakerHandler(container *usecase.Container, logger *logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		recoveryManager := container.GetRecoveryManager()
		stats := recoveryManager.GetCircuitBreakerStats()
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		// Convert to JSON manually to avoid external dependencies
		var statsJSON string = "{"
		first := true
		for name, stat := range stats {
			if !first {
				statsJSON += ","
			}
			statsJSON += fmt.Sprintf(`"%s":%v`, name, stat)
			first = false
		}
		statsJSON += "}"
		
		fmt.Fprintf(w, `{"circuit_breakers":%s,"timestamp":"%s"}`, 
			statsJSON, time.Now().Format(time.RFC3339))
	}
}