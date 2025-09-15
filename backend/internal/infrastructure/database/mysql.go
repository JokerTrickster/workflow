package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"ai-git-workbench/internal/infrastructure/config"
)

// DB holds the database connection and provides database operations
type DB struct {
	*sql.DB
}

// NewMySQLConnection creates a new MySQL database connection with enhanced configuration
func NewMySQLConnection(cfg *config.DatabaseConfig) (*DB, error) {
	// Build MySQL connection string with additional parameters
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local&multiStatements=true",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.Charset,
	)

	// Open database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Configure connection pool with optimized settings
	db.SetMaxOpenConns(25)                 // Maximum number of open connections
	db.SetMaxIdleConns(10)                 // Maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Connection lifetime
	db.SetConnMaxIdleTime(30 * time.Second) // Maximum idle time

	// Test the connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Printf("🔌 Connected to MySQL database: %s@%s:%s/%s", cfg.User, cfg.Host, cfg.Port, cfg.Name)

	// Initialize database (create tables if they don't exist)
	dbInstance := &DB{db}
	if err := dbInstance.Initialize(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error initializing database: %w", err)
	}

	return dbInstance, nil
}

// Initialize runs database migrations and creates tables
func (db *DB) Initialize() error {
	log.Println("🚀 Initializing database...")
	
	migrationPath := "migrations/001_create_tasks_table.sql"
	
	// Check if migration file exists
	if content, err := ioutil.ReadFile(migrationPath); err == nil {
		log.Printf("📄 Running migration: %s", migrationPath)
		
		// Execute migration
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("error executing migration %s: %w", migrationPath, err)
		}
		
		log.Printf("✅ Migration completed: %s", migrationPath)
	} else {
		// Fallback: create tables directly if migration file not found
		log.Println("⚠️ Migration file not found, creating tables directly...")
		if err := db.createTables(); err != nil {
			return fmt.Errorf("error creating tables: %w", err)
		}
	}
	
	log.Println("✅ Database initialization completed")
	return nil
}

// createTables creates the required tables directly (fallback)
func (db *DB) createTables() error {
	createTasksTable := `
	CREATE TABLE IF NOT EXISTS tasks (
		id VARCHAR(36) PRIMARY KEY,
		branch_name VARCHAR(255) NOT NULL,
		title VARCHAR(500) NOT NULL,
		content TEXT,
		repository VARCHAR(255) NOT NULL,
		user_id VARCHAR(255) NOT NULL,
		status ENUM('pending', 'processing', 'completed', 'failed', 'cancelled') DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		
		INDEX idx_user_status (user_id, status),
		INDEX idx_created_at (created_at),
		INDEX idx_repository (repository),
		INDEX idx_status (status)
	);`

	createMetadataTable := `
	CREATE TABLE IF NOT EXISTS task_metadata (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		task_id VARCHAR(36) NOT NULL,
		meta_key VARCHAR(255) NOT NULL,
		meta_value TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		UNIQUE KEY unique_task_key (task_id, meta_key),
		INDEX idx_task_id (task_id)
	);`

	// Execute table creation
	if _, err := db.Exec(createTasksTable); err != nil {
		return fmt.Errorf("error creating tasks table: %w", err)
	}

	if _, err := db.Exec(createMetadataTable); err != nil {
		return fmt.Errorf("error creating task_metadata table: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	log.Println("🔌 Closing database connection...")
	return db.DB.Close()
}

// Ping tests the database connection
func (db *DB) Ping() error {
	return db.DB.Ping()
}

// PingContext tests the database connection with context
func (db *DB) PingContext(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}

// GetStats returns database connection statistics
func (db *DB) GetStats() sql.DBStats {
	return db.DB.Stats()
}

// BeginTx starts a new transaction
func (db *DB) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return db.DB.BeginTx(ctx, nil)
}

// RunMigration executes a migration file
func (db *DB) RunMigration(migrationPath string) error {
	content, err := ioutil.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("error reading migration file %s: %w", migrationPath, err)
	}

	if _, err := db.Exec(string(content)); err != nil {
		return fmt.Errorf("error executing migration %s: %w", migrationPath, err)
	}

	log.Printf("✅ Migration executed successfully: %s", filepath.Base(migrationPath))
	return nil
}