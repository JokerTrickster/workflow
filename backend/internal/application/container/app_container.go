package container

import (
	"context"
	"fmt"
	"log"
	"os"

	"ai-git-workbench/internal/application/interfaces"
	"ai-git-workbench/internal/application/services"
	"ai-git-workbench/internal/application/usecases"
	"ai-git-workbench/internal/domain/repositories"
	domainServices "ai-git-workbench/internal/domain/services"
	"ai-git-workbench/internal/infrastructure/config"
	"ai-git-workbench/internal/infrastructure/database"
	"ai-git-workbench/internal/infrastructure/queue"
	mysqlRepo "ai-git-workbench/internal/infrastructure/repositories"
)

// ApplicationContainer holds all application layer dependencies
type ApplicationContainer struct {
	// Configuration and Infrastructure
	Config *config.Config
	DB     *database.DB

	// Repositories
	TaskRepository  repositories.TaskRepository
	QueueRepository repositories.QueueRepository

	// Domain Services
	TaskValidationService *domainServices.TaskValidationService
	TaskLifecycleService  *domainServices.TaskLifecycleService

	// Application Services
	AuthorizationService interfaces.TaskAuthorizationService
	EventService         interfaces.TaskEventService

	// Use Cases
	TaskUsecase interfaces.TaskUsecase
}

// NewApplicationContainer creates a new application container with all dependencies
func NewApplicationContainer() (*ApplicationContainer, error) {
	log.Printf("🏗️ Initializing application container...")

	// Load configuration
	cfg := config.Load()
	log.Printf("📋 Configuration loaded")

	// Initialize database connection
	db, err := database.NewMySQLConnection(&cfg.Database)
	if err != nil {
		log.Printf("❌ Failed to initialize database: %v", err)
		return nil, err
	}
	log.Printf("🗄️ Database connection established")

	// Initialize repositories
	taskRepo := mysqlRepo.NewMySQLTaskRepository(db)
	log.Printf("📦 Task repository initialized")

	// Initialize queue repository (RabbitMQ or Mock based on environment)
	var queueRepo repositories.QueueRepository
	var err error
	
	if os.Getenv("USE_RABBITMQ") == "true" {
		queueRepo, err = queue.NewRabbitMQRepository(&cfg.RabbitMQ)
		if err != nil {
			log.Printf("⚠️ Failed to initialize RabbitMQ, falling back to mock: %v", err)
			queueRepo = services.NewMockQueueRepository()
			log.Printf("📨 Mock queue repository initialized (fallback)")
		} else {
			log.Printf("📨 RabbitMQ queue repository initialized")
		}
	} else {
		queueRepo = services.NewMockQueueRepository()
		log.Printf("📨 Mock queue repository initialized (default)")
	}

	// Initialize domain services
	validationService := domainServices.NewTaskValidationService(taskRepo)
	lifecycleService := domainServices.NewTaskLifecycleService(taskRepo, queueRepo, validationService)
	log.Printf("🔍 Domain services initialized")

	// Initialize application services
	authService := services.NewAuthorizationService(taskRepo)
	eventService := services.NewEventService()
	log.Printf("🛡️ Application services initialized")

	// Initialize use cases
	taskUsecase := usecases.NewTaskUsecase(
		taskRepo,
		queueRepo,
		validationService,
		lifecycleService,
		authService,
		eventService,
	)
	log.Printf("🎯 Use cases initialized")

	container := &ApplicationContainer{
		Config:                cfg,
		DB:                    db,
		TaskRepository:        taskRepo,
		QueueRepository:       queueRepo,
		TaskValidationService: validationService,
		TaskLifecycleService:  lifecycleService,
		AuthorizationService:  authService,
		EventService:          eventService,
		TaskUsecase:           taskUsecase,
	}

	log.Printf("✅ Application container initialized successfully")
	return container, nil
}

// Close closes all resources
func (c *ApplicationContainer) Close() error {
	log.Printf("🧹 Closing application container resources...")

	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			log.Printf("❌ Error closing database: %v", err)
			return err
		}
	}

	// Close queue repository connection
	if c.QueueRepository != nil {
		// Check if it's the RabbitMQ implementation that has a Close method
		if closer, ok := c.QueueRepository.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				log.Printf("❌ Error closing queue repository: %v", err)
				return err
			}
		}
	}

	log.Printf("✅ Application container resources closed successfully")
	return nil
}

// GetTaskUsecase returns the task use case
func (c *ApplicationContainer) GetTaskUsecase() interfaces.TaskUsecase {
	return c.TaskUsecase
}

// GetAuthorizationService returns the authorization service
func (c *ApplicationContainer) GetAuthorizationService() interfaces.TaskAuthorizationService {
	return c.AuthorizationService
}

// GetEventService returns the event service
func (c *ApplicationContainer) GetEventService() interfaces.TaskEventService {
	return c.EventService
}

// GetTaskRepository returns the task repository
func (c *ApplicationContainer) GetTaskRepository() repositories.TaskRepository {
	return c.TaskRepository
}

// GetQueueRepository returns the queue repository
func (c *ApplicationContainer) GetQueueRepository() repositories.QueueRepository {
	return c.QueueRepository
}

// GetDomainServices returns the domain services
func (c *ApplicationContainer) GetDomainServices() (*domainServices.TaskValidationService, *domainServices.TaskLifecycleService) {
	return c.TaskValidationService, c.TaskLifecycleService
}

// Health check methods

// HealthCheck performs a health check on all components
func (c *ApplicationContainer) HealthCheck() error {
	log.Printf("🏥 Performing application health check...")

	// Check database connection
	if err := c.DB.Ping(); err != nil {
		log.Printf("❌ Database health check failed: %v", err)
		return err
	}

	// Check queue repository health
	if queueHealth, err := c.QueueRepository.GetQueueHealth(context.Background()); err != nil {
		log.Printf("❌ Queue repository health check failed: %v", err)
	} else if !queueHealth.IsHealthy {
		log.Printf("⚠️ Queue repository is unhealthy: %v", queueHealth.Issues)
	}

	log.Printf("✅ Application health check passed")
	return nil
}

// GetStatus returns the current status of all components
func (c *ApplicationContainer) GetStatus() map[string]string {
	status := make(map[string]string)

	// Database status
	if err := c.DB.Ping(); err != nil {
		status["database"] = "unhealthy: " + err.Error()
	} else {
		status["database"] = "healthy"
	}

	// Application services status
	status["task_usecase"] = "healthy"
	status["authorization_service"] = "healthy"
	status["event_service"] = "healthy"

	// Queue repository status
	if queueHealth, err := c.QueueRepository.GetQueueHealth(context.Background()); err != nil {
		status["queue_repository"] = "unhealthy: " + err.Error()
	} else if queueHealth.IsHealthy {
		status["queue_repository"] = "healthy"
	} else {
		status["queue_repository"] = "unhealthy: " + fmt.Sprintf("%v", queueHealth.Issues)
	}

	return status
}