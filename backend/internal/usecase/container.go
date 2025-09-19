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
	"local-backend-server/internal/infrastructure/errors"
	"local-backend-server/internal/infrastructure/logging"
	"local-backend-server/internal/infrastructure/queue"
)

// Container manages all application dependencies
type Container struct {
	config *config.Config

	// Infrastructure
	logger           *logging.Logger
	errorHandler     *errors.ErrorHandler
	recoveryManager  *errors.RecoveryManager
	dbConnection     *database.DB
	queueConn        *queue.Connection

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
	// Initialize logging first
	if err := c.initializeLogging(); err != nil {
		return fmt.Errorf("failed to initialize logging: %w", err)
	}

	// Initialize error handling
	if err := c.initializeErrorHandling(); err != nil {
		return fmt.Errorf("failed to initialize error handling: %w", err)
	}

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

// initializeLogging sets up structured logging
func (c *Container) initializeLogging() error {
	logger, err := logging.NewLogger(&c.config.Logging)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	
	c.logger = logger
	c.logger.WithComponent("container").Info("Logging system initialized")
	return nil
}

// initializeErrorHandling sets up error handling and recovery
func (c *Container) initializeErrorHandling() error {
	// Create error handler
	c.errorHandler = errors.NewErrorHandler(c.logger, c.config.IsDevelopment())
	
	// Create recovery manager with circuit breakers
	c.recoveryManager = errors.NewRecoveryManager(c.logger)
	
	// Add circuit breakers for external services
	c.recoveryManager.AddCircuitBreaker("database", 5, c.config.Timeout.Database*3)
	c.recoveryManager.AddCircuitBreaker("rabbitmq", 3, c.config.Timeout.Queue*2)
	c.recoveryManager.AddCircuitBreaker("claude", 5, c.config.Timeout.Claude*2)
	
	c.logger.WithComponent("container").Info("Error handling system initialized")
	return nil
}

// initializeInfrastructure sets up database and queue connections
func (c *Container) initializeInfrastructure(ctx context.Context) error {
	c.logger.WithComponent("container").Info("Initializing infrastructure components")

	// Initialize database connection with retry logic
	err := c.recoveryManager.ExecuteWithRetry(ctx, "database_init", func(ctx context.Context) error {
		conn, err := database.NewConnection(c.config)
		if err != nil {
			return errors.DatabaseError("connection", err)
		}
		c.dbConnection = conn
		return nil
	}, errors.DefaultRetryConfig())

	if err != nil {
		return fmt.Errorf("failed to create database connection: %w", err)
	}

	c.logger.WithComponent("database").Info("Database connection established")

	// Initialize queue connection with retry logic
	err = c.recoveryManager.ExecuteWithRetry(ctx, "rabbitmq_init", func(ctx context.Context) error {
		conn, err := queue.NewConnection(&c.config.RabbitMQ)
		if err != nil {
			return errors.QueueError("connection", err)
		}
		c.queueConn = conn
		return nil
	}, errors.DefaultRetryConfig())

	if err != nil {
		c.logger.WithComponent("rabbitmq").WithError(err).Warn("Failed to connect to RabbitMQ, will retry later")
		// Don't fail initialization for queue connection issues
		// The application can still serve health checks and handle some operations
	} else {
		c.logger.WithComponent("rabbitmq").Info("RabbitMQ connection established")
	}

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
		if c.config.Environment == "production" {
			return fmt.Errorf("failed to create Claude client: %w", err)
		}
		// In development, allow graceful degradation without Claude API
		c.logger.WithComponent("claude").WithError(err).Warn("Failed to create Claude client, will operate in degraded mode")
		c.claudeClient = nil
	} else {
		c.claudeClient = claudeClient
		c.logger.WithComponent("claude").Info("Claude client initialized successfully")
	}

	// Initialize template manager
	c.templateManager = claude.NewTemplateManager()

	// Initialize context manager
	// Only create if Claude client is available
	if c.claudeClient != nil {
		c.contextManager = claude.NewContextManager(
			c.claudeClient,
			c.processingRepo,
			c.messageRepo,
		)
		c.logger.WithComponent("claude").Info("Context manager initialized successfully")
	} else {
		c.contextManager = nil
		c.logger.WithComponent("claude").Info("Claude client not available, skipping context manager")
	}

	// Initialize queue consumer with nil processor for now
	// Only create consumer if queue connection is available
	if c.queueConn != nil {
		queueConsumer, err := queue.NewConsumer(&c.config.RabbitMQ, nil)
		if err != nil {
			c.logger.WithComponent("rabbitmq").WithError(err).Warn("Failed to create queue consumer, will retry later")
			// Don't fail initialization for queue consumer issues
		} else {
			c.queueConsumer = queueConsumer
			c.logger.WithComponent("rabbitmq").Info("Queue consumer created successfully")
		}
	} else {
		c.logger.WithComponent("rabbitmq").Info("Queue connection not available, skipping consumer creation")
	}

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
	// Only if consumer is available
	if c.queueConsumer != nil {
		c.queueConsumer.SetProcessor(c.workflowOrchestrator)
		c.logger.WithComponent("orchestrator").Info("Queue consumer processor set successfully")
	} else {
		c.logger.WithComponent("orchestrator").Info("Queue consumer not available, skipping processor setup")
	}

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

// GetLogger returns the logger instance
func (c *Container) GetLogger() *logging.Logger {
	return c.logger
}

// GetErrorHandler returns the error handler instance
func (c *Container) GetErrorHandler() *errors.ErrorHandler {
	return c.errorHandler
}

// GetRecoveryManager returns the recovery manager instance
func (c *Container) GetRecoveryManager() *errors.RecoveryManager {
	return c.recoveryManager
}

// Cleanup performs cleanup of all resources
func (c *Container) Cleanup() error {
	c.logger.WithComponent("container").Info("Starting cleanup of all resources")

	// Close queue connection
	if c.queueConn != nil {
		if err := c.queueConn.Close(); err != nil {
			c.logger.WithComponent("rabbitmq").WithError(err).Error("Failed to close queue connection")
			return fmt.Errorf("failed to close queue connection: %w", err)
		}
		c.logger.WithComponent("rabbitmq").Info("Queue connection closed")
	}

	// Close database connection
	if c.dbConnection != nil {
		if err := c.dbConnection.Close(); err != nil {
			c.logger.WithComponent("database").WithError(err).Error("Failed to close database connection")
			return fmt.Errorf("failed to close database connection: %w", err)
		}
		c.logger.WithComponent("database").Info("Database connection closed")
	}

	// Close logger last
	if c.logger != nil {
		if err := c.logger.Close(); err != nil {
			return fmt.Errorf("failed to close logger: %w", err)
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
	if c.queueConn == nil {
		health["components"].(map[string]interface{})["rabbitmq"] = map[string]interface{}{
			"status": "unavailable",
			"error":  "Queue connection not established",
		}
		health["status"] = "degraded"
	} else if err := c.queueConn.Health(); err != nil {
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