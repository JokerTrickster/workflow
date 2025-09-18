package main

import (
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

	// TODO: Initialize database connection
	// TODO: Initialize RabbitMQ consumer  
	// TODO: Initialize Claude API client
	// TODO: Start message processing loop

	log.Println("Local Backend Server initialized successfully")
	log.Println("Ready to process messages from RabbitMQ queue")

	// Wait for interrupt signal to gracefully shut down
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Local Backend Server...")
	// TODO: Implement graceful shutdown
	log.Println("Server stopped")
}