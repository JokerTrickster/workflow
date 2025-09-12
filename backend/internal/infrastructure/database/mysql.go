package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"ai-git-workbench/internal/infrastructure/config"
)

// DB holds the database connection
type DB struct {
	*sql.DB
}

// NewMySQLConnection creates a new MySQL database connection
func NewMySQLConnection(cfg *config.DatabaseConfig) (*DB, error) {
	// Build MySQL connection string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
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

	// Configure connection pool
	db.SetMaxOpenConns(25)                 // Maximum number of open connections
	db.SetMaxIdleConns(25)                 // Maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Connection lifetime

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Printf("🔌 Connected to MySQL database: %s@%s:%s/%s", cfg.User, cfg.Host, cfg.Port, cfg.Name)

	return &DB{db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// Ping tests the database connection
func (db *DB) Ping() error {
	return db.DB.Ping()
}

// GetStats returns database connection statistics
func (db *DB) GetStats() sql.DBStats {
	return db.DB.Stats()
}