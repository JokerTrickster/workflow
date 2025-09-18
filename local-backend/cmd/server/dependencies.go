package main

import (
	"fmt"
	"log"

	"github.com/JokerTrickster/workflow/local-backend/internal/infrastructure"
	"github.com/JokerTrickster/workflow/local-backend/internal/repository"
	"github.com/JokerTrickster/workflow/local-backend/internal/usecase"
)

// initializeServices initializes all services and returns the orchestrator
func initializeServices(config *infrastructure.Config) (*usecase.ServiceOrchestrator, func(), error) {
	// Initialize database
	db, err := infrastructure.NewDatabase(&config.Database)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Run migrations
	if err := db.AutoMigrate(); err != nil {
		return nil, nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database initialized and migrations completed")

	// Initialize repositories
	requestRepo := repository.NewRequestRepository(db.DB)
	contextRepo := repository.NewProcessingContextRepository(db.DB)

	// Initialize Claude service
	claudeService, err := infrastructure.NewSimplifiedClaudeService(&config.Claude, contextRepo)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize Claude service: %w", err)
	}

	// Validate Claude service configuration
	if err := claudeService.ValidateConfiguration(); err != nil {
		return nil, nil, fmt.Errorf("Claude service configuration invalid: %w", err)
	}

	log.Println("Claude service initialized")

	// Initialize RabbitMQ consumer
	rabbitMQConsumer, err := infrastructure.NewRabbitMQConsumer(&config.RabbitMQ)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize RabbitMQ consumer: %w", err)
	}

	log.Println("RabbitMQ consumer initialized")

	// Initialize use case services
	contextService := usecase.NewContextService(contextRepo)
	requestService := usecase.NewRequestService(requestRepo, contextRepo, claudeService)
	messageProcessor := usecase.NewMessageProcessor(requestService, contextService)

	// Initialize orchestrator
	orchestrator := usecase.NewServiceOrchestrator(
		messageProcessor,
		requestService,
		contextService,
		claudeService,
		rabbitMQConsumer,
	)

	// Cleanup function
	cleanup := func() {
		log.Println("Cleaning up resources...")
		
		if err := rabbitMQConsumer.StopConsuming(); err != nil {
			log.Printf("Error stopping RabbitMQ consumer: %v", err)
		}
		
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
		
		log.Println("Cleanup completed")
	}

	return orchestrator, cleanup, nil
}