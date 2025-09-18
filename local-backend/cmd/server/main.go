package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/JokerTrickster/workflow/local-backend/internal/infrastructure"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, continuing with environment variables")
	}

	// Initialize configuration
	config, err := infrastructure.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting %s v%s in %s mode", 
		config.App.Name, 
		config.App.Version, 
		config.App.Environment)

	// Initialize all services
	orchestrator, cleanup, err := initializeServices(config)
	if err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}
	defer cleanup()

	// Start the orchestrator
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := orchestrator.Start(ctx); err != nil {
		log.Fatalf("Failed to start services: %v", err)
	}

	log.Println("Local Backend Server started successfully")
	log.Println("Ready to process messages from RabbitMQ queue")

	// Wait for interrupt signal to gracefully shut down
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Local Backend Server...")
	
	// Cancel context to stop background tasks
	cancel()
	
	// Stop the orchestrator
	if err := orchestrator.Stop(); err != nil {
		log.Printf("Error stopping orchestrator: %v", err)
	}
	
	log.Println("Server stopped")
}