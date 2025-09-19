package main

import (
	"log"

	"local-backend-server/internal/infrastructure/config"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting local-backend-server on %s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Println("Configuration loaded successfully")

	// TODO: Initialize database
	// TODO: Initialize RabbitMQ consumer  
	// TODO: Initialize Claude service
	// TODO: Start HTTP server
	
	log.Println("Server initialization complete")
}