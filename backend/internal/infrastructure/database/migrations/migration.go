package migrations

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"local-backend-server/internal/infrastructure/database/models"
)

// Migration represents a database migration
type Migration struct {
	ID          string    `gorm:"primaryKey;type:varchar(255)" json:"id"`
	Version     string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"version"`
	Description string    `gorm:"type:varchar(500);not null" json:"description"`
	AppliedAt   time.Time `gorm:"not null" json:"applied_at"`
}

// TableName specifies the table name for Migration model
func (Migration) TableName() string {
	return "migrations"
}

// MigrationFunc represents a migration function
type MigrationFunc func(*gorm.DB) error

// MigrationDefinition contains migration details
type MigrationDefinition struct {
	Version     string
	Description string
	Up          MigrationFunc
	Down        MigrationFunc
}

// Migrator handles database migrations
type Migrator struct {
	db *gorm.DB
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
}

// Initialize creates the migrations table if it doesn't exist
func (m *Migrator) Initialize() error {
	return m.db.AutoMigrate(&Migration{})
}

// RunMigrations executes all pending migrations
func (m *Migrator) RunMigrations() error {
	// Ensure migrations table exists
	if err := m.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize migrations table: %w", err)
	}

	migrations := GetAllMigrations()
	
	for _, migration := range migrations {
		applied, err := m.isMigrationApplied(migration.Version)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", migration.Version, err)
		}
		
		if !applied {
			log.Printf("Running migration: %s - %s", migration.Version, migration.Description)
			
			// Start transaction
			tx := m.db.Begin()
			
			// Run migration
			if err := migration.Up(tx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to run migration %s: %w", migration.Version, err)
			}
			
			// Record migration
			migrationRecord := &Migration{
				ID:          fmt.Sprintf("migration_%s_%d", migration.Version, time.Now().Unix()),
				Version:     migration.Version,
				Description: migration.Description,
				AppliedAt:   time.Now().UTC(),
			}
			
			if err := tx.Create(migrationRecord).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
			}
			
			// Commit transaction
			if err := tx.Commit().Error; err != nil {
				return fmt.Errorf("failed to commit migration %s: %w", migration.Version, err)
			}
			
			log.Printf("Migration completed: %s", migration.Version)
		}
	}
	
	log.Println("All migrations completed successfully")
	return nil
}

// RollbackMigration rolls back a specific migration
func (m *Migrator) RollbackMigration(version string) error {
	migration := GetMigrationByVersion(version)
	if migration == nil {
		return fmt.Errorf("migration not found: %s", version)
	}
	
	applied, err := m.isMigrationApplied(version)
	if err != nil {
		return fmt.Errorf("failed to check migration status for %s: %w", version, err)
	}
	
	if !applied {
		return fmt.Errorf("migration %s is not applied", version)
	}
	
	log.Printf("Rolling back migration: %s - %s", migration.Version, migration.Description)
	
	// Start transaction
	tx := m.db.Begin()
	
	// Run rollback
	if err := migration.Down(tx); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to rollback migration %s: %w", version, err)
	}
	
	// Remove migration record
	if err := tx.Where("version = ?", version).Delete(&Migration{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove migration record %s: %w", version, err)
	}
	
	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit rollback %s: %w", version, err)
	}
	
	log.Printf("Migration rolled back: %s", version)
	return nil
}

// GetAppliedMigrations returns all applied migrations
func (m *Migrator) GetAppliedMigrations() ([]Migration, error) {
	var migrations []Migration
	err := m.db.Order("applied_at ASC").Find(&migrations).Error
	return migrations, err
}

// isMigrationApplied checks if a migration is already applied
func (m *Migrator) isMigrationApplied(version string) (bool, error) {
	var count int64
	err := m.db.Model(&Migration{}).Where("version = ?", version).Count(&count).Error
	return count > 0, err
}

// GetAllMigrations returns all available migrations in order
func GetAllMigrations() []*MigrationDefinition {
	return []*MigrationDefinition{
		{
			Version:     "001_initial_schema",
			Description: "Create initial database schema",
			Up:          migration001Up,
			Down:        migration001Down,
		},
		{
			Version:     "002_add_indexes",
			Description: "Add database indexes for performance",
			Up:          migration002Up,
			Down:        migration002Down,
		},
	}
}

// GetMigrationByVersion returns a migration by version
func GetMigrationByVersion(version string) *MigrationDefinition {
	migrations := GetAllMigrations()
	for _, migration := range migrations {
		if migration.Version == version {
			return migration
		}
	}
	return nil
}

// Migration 001: Initial schema
func migration001Up(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Session{},
		&models.Message{},
		&models.Request{},
		&models.ProcessingContext{},
		&models.WorkflowHistory{},
	)
}

func migration001Down(db *gorm.DB) error {
	return db.Migrator().DropTable(
		&models.WorkflowHistory{},
		&models.ProcessingContext{},
		&models.Request{},
		&models.Message{},
		&models.Session{},
	)
}

// Migration 002: Add indexes
func migration002Up(db *gorm.DB) error {
	// Add custom indexes for better performance
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_status_expires ON sessions(status, expires_at)").Error; err != nil {
		return err
	}
	
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_messages_session_created ON messages(session_id, created_at)").Error; err != nil {
		return err
	}
	
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_requests_status_created ON requests(status, created_at)").Error; err != nil {
		return err
	}
	
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_requests_session_type ON requests(session_id, type)").Error; err != nil {
		return err
	}
	
	// Add workflow_histories indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_workflow_histories_request_id ON workflow_histories(request_id)").Error; err != nil {
		return err
	}
	
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_workflow_histories_status_created ON workflow_histories(status, created_at)").Error; err != nil {
		return err
	}
	
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_workflow_histories_repository ON workflow_histories(repository_name, created_at)").Error; err != nil {
		return err
	}
	
	return nil
}

func migration002Down(db *gorm.DB) error {
	// Drop custom indexes
	db.Exec("DROP INDEX IF EXISTS idx_sessions_status_expires")
	db.Exec("DROP INDEX IF EXISTS idx_messages_session_created")
	db.Exec("DROP INDEX IF EXISTS idx_requests_status_created")
	db.Exec("DROP INDEX IF EXISTS idx_requests_session_type")
	db.Exec("DROP INDEX IF EXISTS idx_workflow_histories_request_id")
	db.Exec("DROP INDEX IF EXISTS idx_workflow_histories_status_created")
	db.Exec("DROP INDEX IF EXISTS idx_workflow_histories_repository")
	
	return nil
}