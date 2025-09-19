package infrastructure

import (
	"fmt"
	"log"

	"gorm.io/gorm"

	"github.com/JokerTrickster/workflow/local-backend/internal/domain"
)

// MigrationManager handles database migrations
type MigrationManager struct {
	db *gorm.DB
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *gorm.DB) *MigrationManager {
	return &MigrationManager{db: db}
}

// RunMigrations executes all pending migrations
func (m *MigrationManager) RunMigrations() error {
	log.Println("Running database migrations...")

	// Run auto-migrations for all domain models
	if err := m.db.AutoMigrate(
		&domain.Request{},
		&domain.ProcessingContext{},
	); err != nil {
		return fmt.Errorf("failed to run auto migrations: %w", err)
	}

	// Run custom migrations
	if err := m.runCustomMigrations(); err != nil {
		return fmt.Errorf("failed to run custom migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// runCustomMigrations executes custom migration scripts
func (m *MigrationManager) runCustomMigrations() error {
	// Create indexes for better query performance
	if err := m.createIndexes(); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	// Add any custom constraints or triggers
	if err := m.addConstraints(); err != nil {
		return fmt.Errorf("failed to add constraints: %w", err)
	}

	return nil
}

// createIndexes creates database indexes for performance optimization
func (m *MigrationManager) createIndexes() error {
	// Index on requests table for common queries
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_requests_status ON requests(status)",
		"CREATE INDEX IF NOT EXISTS idx_requests_message_id ON requests(message_id)",
		"CREATE INDEX IF NOT EXISTS idx_requests_context_id ON requests(context_id)",
		"CREATE INDEX IF NOT EXISTS idx_requests_created_at ON requests(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_requests_status_created_at ON requests(status, created_at)",
		
		// Index on processing_contexts table
		"CREATE INDEX IF NOT EXISTS idx_contexts_last_used_at ON processing_contexts(last_used_at)",
		"CREATE INDEX IF NOT EXISTS idx_contexts_created_at ON processing_contexts(created_at)",
	}

	for _, indexSQL := range indexes {
		if err := m.db.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
			// Continue with other indexes even if one fails
		}
	}

	return nil
}

// addConstraints adds custom database constraints
func (m *MigrationManager) addConstraints() error {
	// Add check constraints for SQLite (if supported)
	constraints := []string{
		// Ensure status is valid
		"CREATE TRIGGER IF NOT EXISTS check_request_status " +
			"BEFORE INSERT ON requests " +
			"WHEN NEW.status NOT IN ('pending', 'processing', 'completed', 'failed', 'cancelled') " +
			"BEGIN " +
			"SELECT RAISE(ABORT, 'Invalid request status'); " +
			"END",

		"CREATE TRIGGER IF NOT EXISTS check_request_status_update " +
			"BEFORE UPDATE ON requests " +
			"WHEN NEW.status NOT IN ('pending', 'processing', 'completed', 'failed', 'cancelled') " +
			"BEGIN " +
			"SELECT RAISE(ABORT, 'Invalid request status'); " +
			"END",
	}

	for _, constraintSQL := range constraints {
		if err := m.db.Exec(constraintSQL).Error; err != nil {
			log.Printf("Warning: Failed to create constraint: %v", err)
			// Continue with other constraints even if one fails
		}
	}

	return nil
}

// DropAllTables drops all tables (for testing or reset purposes)
func (m *MigrationManager) DropAllTables() error {
	log.Println("Dropping all database tables...")

	tables := []interface{}{
		&domain.ProcessingContext{},
		&domain.Request{},
	}

	for _, table := range tables {
		if err := m.db.Migrator().DropTable(table); err != nil {
			return fmt.Errorf("failed to drop table %T: %w", table, err)
		}
	}

	log.Println("All tables dropped successfully")
	return nil
}

// GetMigrationInfo returns information about the current database schema
func (m *MigrationManager) GetMigrationInfo() map[string]interface{} {
	info := make(map[string]interface{})

	// Check which tables exist
	tables := []string{"requests", "processing_contexts"}
	existingTables := make([]string, 0)
	
	for _, table := range tables {
		if m.db.Migrator().HasTable(table) {
			existingTables = append(existingTables, table)
		}
	}
	
	info["existing_tables"] = existingTables

	// Get table counts
	var requestCount, contextCount int64
	m.db.Model(&domain.Request{}).Count(&requestCount)
	m.db.Model(&domain.ProcessingContext{}).Count(&contextCount)
	
	info["request_count"] = requestCount
	info["context_count"] = contextCount

	return info
}

// ValidateSchema validates that the database schema matches expectations
func (m *MigrationManager) ValidateSchema() error {
	// Check required tables exist
	requiredTables := []interface{}{
		&domain.Request{},
		&domain.ProcessingContext{},
	}

	for _, table := range requiredTables {
		if !m.db.Migrator().HasTable(table) {
			return fmt.Errorf("required table for %T does not exist", table)
		}
	}

	// Check required columns exist
	if !m.db.Migrator().HasColumn(&domain.Request{}, "id") ||
		!m.db.Migrator().HasColumn(&domain.Request{}, "status") ||
		!m.db.Migrator().HasColumn(&domain.Request{}, "message_id") {
		return fmt.Errorf("requests table is missing required columns")
	}

	if !m.db.Migrator().HasColumn(&domain.ProcessingContext{}, "id") ||
		!m.db.Migrator().HasColumn(&domain.ProcessingContext{}, "messages") ||
		!m.db.Migrator().HasColumn(&domain.ProcessingContext{}, "last_used_at") {
		return fmt.Errorf("processing_contexts table is missing required columns")
	}

	return nil
}