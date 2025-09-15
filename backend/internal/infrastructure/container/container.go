package container

import (
	"log"

	"ai-git-workbench/internal/domain/repositories"
	"ai-git-workbench/internal/infrastructure/config"
	"ai-git-workbench/internal/infrastructure/database"
	mysqlRepo "ai-git-workbench/internal/infrastructure/repositories"
)

// Container holds all the application dependencies
type Container struct {
	Config         *config.Config
	DB             *database.DB
	TaskRepository repositories.TaskRepository
}

// NewContainer creates a new dependency injection container
func NewContainer() (*Container, error) {
	// Load configuration
	cfg := config.Load()
	log.Printf("📋 Configuration loaded")

	// Initialize database connection
	db, err := database.NewMySQLConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	// Initialize repositories
	taskRepo := mysqlRepo.NewMySQLTaskRepository(db)

	container := &Container{
		Config:         cfg,
		DB:             db,
		TaskRepository: taskRepo,
	}

	log.Printf("🏗️ Dependency container initialized successfully")
	return container, nil
}

// Close closes all resources
func (c *Container) Close() error {
	log.Printf("🧹 Closing container resources...")
	
	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			log.Printf("❌ Error closing database: %v", err)
			return err
		}
	}
	
	log.Printf("✅ Container resources closed successfully")
	return nil
}