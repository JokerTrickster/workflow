package usecase

import (
	"context"
	"fmt"

	"local-backend-server/internal/domain/repositories"
	"local-backend-server/internal/domain/services"
	"local-backend-server/internal/infrastructure/claude"
	"local-backend-server/internal/infrastructure/config"
	"local-backend-server/internal/infrastructure/database"
	dbRepositories "local-backend-server/internal/infrastructure/database/repositories"
	"local-backend-server/internal/infrastructure/queue"
)

// Container manages all application dependencies
type Container struct {
	config *config.Config

	// Infrastructure
	dbConnection *database.DB
	queueConn    *queue.Connection

	// Repositories
	requestRepo      repositories.RequestRepository
	messageRepo      repositories.MessageRepository
	sessionRepo      repositories.SessionRepository
	processingRepo   repositories.ProcessingContextRepository

	// Domain Services
	requestService *services.RequestService
	messageService *services.MessageService
	sessionService *services.SessionService

	// Infrastructure Services
	claudeClient    *claude.Client
	contextManager  *claude.ContextManager
	templateManager *claude.TemplateManager
	queueConsumer   *queue.Consumer

	// Application Services
	workflowOrchestrator *WorkflowOrchestrator
}

// NewContainer creates a new dependency injection container
func NewContainer(cfg *config.Config) *Container {
	return &Container{
		config: cfg,
	}
}

// Initialize sets up all dependencies in the correct order
func (c *Container) Initialize(ctx context.Context) error {
	// Initialize infrastructure
	if err := c.initializeInfrastructure(ctx); err != nil {
		return fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	// Initialize repositories
	if err := c.initializeRepositories(); err != nil {
		return fmt.Errorf("failed to initialize repositories: %w", err)
	}

	// Initialize domain services
	if err := c.initializeDomainServices(); err != nil {
		return fmt.Errorf("failed to initialize domain services: %w", err)
	}

	// Initialize infrastructure services
	if err := c.initializeInfrastructureServices(); err != nil {
		return fmt.Errorf("failed to initialize infrastructure services: %w", err)
	}

	// Initialize application services
	if err := c.initializeApplicationServices(); err != nil {
		return fmt.Errorf("failed to initialize application services: %w", err)
	}

	return nil
}

// initializeInfrastructure sets up database and queue connections
func (c *Container) initializeInfrastructure(ctx context.Context) error {
	// Initialize database connection
	dbConn, err := database.NewConnection(c.config)
	if err != nil {
		return fmt.Errorf("failed to create database connection: %w", err)
	}

	c.dbConnection = dbConn

	// Initialize queue connection
	queueConn, err := queue.NewConnection(&c.config.RabbitMQ)
	if err != nil {
		return fmt.Errorf("failed to create queue connection: %w", err)
	}

	c.queueConn = queueConn

	return nil
}

// initializeRepositories creates repository instances
func (c *Container) initializeRepositories() error {
	c.requestRepo = dbRepositories.NewRequestRepository(c.dbConnection.DB)
	c.messageRepo = dbRepositories.NewMessageRepository(c.dbConnection.DB)
	c.sessionRepo = dbRepositories.NewSessionRepository(c.dbConnection.DB)
	// TODO: Fix ProcessingContextRepository compilation issue
	// c.processingRepo = dbRepositories.NewProcessingContextRepository(c.dbConnection.DB)
	c.processingRepo = nil

	return nil
}

// initializeDomainServices creates domain service instances
func (c *Container) initializeDomainServices() error {
	c.requestService = services.NewRequestService(c.requestRepo, c.sessionRepo)
	c.messageService = services.NewMessageService(c.messageRepo, c.sessionRepo)
	c.sessionService = services.NewSessionService(c.sessionRepo)

	return nil
}

// initializeInfrastructureServices creates infrastructure service instances
func (c *Container) initializeInfrastructureServices() error {
	// Initialize Claude client
	claudeClient, err := claude.NewClient(&c.config.Claude)
	if err != nil {
		return fmt.Errorf("failed to create Claude client: %w", err)
	}
	c.claudeClient = claudeClient

	// Initialize template manager
	c.templateManager = claude.NewTemplateManager()

	// Initialize context manager
	c.contextManager = claude.NewContextManager(
		c.claudeClient,
		c.processingRepo,
		c.messageRepo,
	)

	// Initialize queue consumer with nil processor for now
	queueConsumer, err := queue.NewConsumer(&c.config.RabbitMQ, nil)
	if err != nil {
		return fmt.Errorf("failed to create queue consumer: %w", err)
	}
	c.queueConsumer = queueConsumer

	return nil
}

// initializeApplicationServices creates application service instances
func (c *Container) initializeApplicationServices() error {
	// Create workflow orchestrator
	c.workflowOrchestrator = NewWorkflowOrchestrator(
		c.requestService,
		c.messageService,
		c.sessionService,
		c.claudeClient,
		c.contextManager,
		c.templateManager,
		c.requestRepo,
		c.messageRepo,
		c.sessionRepo,
		c.processingRepo,
	)

	// Set queue consumer reference in workflow orchestrator
	c.workflowOrchestrator.SetQueueConsumer(c.queueConsumer)

	// Set up queue consumer to use workflow orchestrator as message processor
	c.queueConsumer.SetProcessor(c.workflowOrchestrator)

	return nil
}

// GetWorkflowOrchestrator returns the workflow orchestrator instance
func (c *Container) GetWorkflowOrchestrator() *WorkflowOrchestrator {
	return c.workflowOrchestrator
}

// GetQueueConsumer returns the queue consumer instance
func (c *Container) GetQueueConsumer() *queue.Consumer {
	return c.queueConsumer
}

// Cleanup performs cleanup of all resources
func (c *Container) Cleanup() error {
	// Close queue connection
	if c.queueConn != nil {
		if err := c.queueConn.Close(); err != nil {
			return fmt.Errorf("failed to close queue connection: %w", err)
		}
	}

	// Close database connection
	if c.dbConnection != nil {
		if err := c.dbConnection.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}
	}

	return nil
}

// HealthCheck performs health check on all components
func (c *Container) HealthCheck(ctx context.Context) map[string]interface{} {
	health := map[string]interface{}{
		"status":     "healthy",
		"components": map[string]interface{}{},
	}

	// Check database
	if err := c.dbConnection.Health(); err != nil {
		health["components"].(map[string]interface{})["database"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		health["status"] = "degraded"
	} else {
		health["components"].(map[string]interface{})["database"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// Check queue connection
	if err := c.queueConn.Health(); err != nil {
		health["components"].(map[string]interface{})["rabbitmq"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		health["status"] = "degraded"
	} else {
		health["components"].(map[string]interface{})["rabbitmq"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// Check workflow orchestrator
	if c.workflowOrchestrator != nil {
		orchestratorHealth := c.workflowOrchestrator.HealthCheck(ctx)
		health["components"].(map[string]interface{})["workflow_orchestrator"] = orchestratorHealth
		if orchestratorHealth["status"] != "healthy" {
			health["status"] = "degraded"
		}
	}

	return health
}