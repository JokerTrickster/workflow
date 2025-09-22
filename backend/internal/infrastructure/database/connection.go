package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"local-backend-server/internal/infrastructure/config"
	"local-backend-server/internal/infrastructure/database/migrations"
)

// DB holds the database connection
type DB struct {
	*gorm.DB
}

// NewConnection creates a new database connection with performance optimizations
func NewConnection(cfg *config.Config) (*DB, error) {
	// Ensure database directory exists
	dbPath := cfg.Database.Path
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Configure GORM logger
	logLevel := logger.Silent
	if cfg.Logging.Level == "debug" {
		logLevel = logger.Info
	}

	// Configure GORM with performance optimizations
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		// Performance optimizations
		PrepareStmt:                              true,  // Enable prepared statements
		DisableForeignKeyConstraintWhenMigrating: false, // Keep constraints for data integrity
		SkipDefaultTransaction:                   false, // Keep transactions for safety
	}

	// Open SQLite connection with performance pragmas
	dsn := buildOptimizedDSN(dbPath, cfg.Environment)
	db, err := gorm.Open(sqlite.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool for optimal performance
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Production-optimized connection pool settings
	maxOpenConns := cfg.Database.MaxConnections
	if maxOpenConns <= 0 {
		maxOpenConns = 10 // Default for SQLite
	}

	// SQLite is single-writer, so limit concurrent connections appropriately
	if maxOpenConns > 25 {
		maxOpenConns = 25 // SQLite optimal max
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxOpenConns / 2)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Apply SQLite performance optimizations
	if err := applyPerformanceOptimizations(db, cfg.Environment); err != nil {
		return nil, fmt.Errorf("failed to apply performance optimizations: %w", err)
	}

	// Run migrations
	migrator := migrations.NewMigrator(db)
	if err := migrator.RunMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Printf("Database connected successfully with optimizations: %s", dbPath)
	return &DB{db}, nil
}

// buildOptimizedDSN creates a DSN with performance-oriented SQLite pragmas
func buildOptimizedDSN(dbPath, environment string) string {
	if environment == "production" {
		// Production optimizations - safety with performance
		return fmt.Sprintf("%s?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=MEMORY&_mmap_size=134217728", dbPath)
	} else if environment == "development" {
		// Development optimizations - speed over safety
		return fmt.Sprintf("%s?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=MEMORY", dbPath)
	}
	// Default safe configuration
	return dbPath
}

// applyPerformanceOptimizations applies SQLite-specific performance settings
func applyPerformanceOptimizations(db *gorm.DB, environment string) error {
	optimizations := []string{
		"PRAGMA optimize",                    // Optimize query planner
		"PRAGMA analysis_limit=1000",        // Limit analysis for faster startup
		"PRAGMA case_sensitive_like=ON",     // Faster LIKE operations
	}

	if environment == "production" {
		// Production-specific optimizations
		productionOpts := []string{
			"PRAGMA busy_timeout=5000",       // 5 second busy timeout
			"PRAGMA wal_autocheckpoint=1000", // Checkpoint every 1000 pages
			"PRAGMA journal_size_limit=67108864", // 64MB journal limit
		}
		optimizations = append(optimizations, productionOpts...)
	}

	for _, pragma := range optimizations {
		if err := db.Exec(pragma).Error; err != nil {
			log.Printf("Warning: Failed to apply optimization '%s': %v", pragma, err)
			// Continue with other optimizations even if one fails
		}
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Health checks database connection health
func (db *DB) Health() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// OptimizeDatabase runs additional optimization commands
func (db *DB) OptimizeDatabase() error {
	optimizations := []string{
		"ANALYZE",           // Update query planner statistics
		"PRAGMA optimize",   // General optimization
		"VACUUM",           // Defragment database (use carefully in production)
	}

	for _, cmd := range optimizations {
		if err := db.Exec(cmd).Error; err != nil {
			log.Printf("Warning: Failed to run optimization '%s': %v", cmd, err)
		}
	}

	return nil
}

// GetDatabaseStats returns database performance statistics
func (db *DB) GetDatabaseStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get SQLite stats
	var pageCount, pageSize, freePages int64

	if err := db.Raw("PRAGMA page_count").Scan(&pageCount).Error; err == nil {
		stats["page_count"] = pageCount
	}

	if err := db.Raw("PRAGMA page_size").Scan(&pageSize).Error; err == nil {
		stats["page_size"] = pageSize
	}

	if err := db.Raw("PRAGMA freelist_count").Scan(&freePages).Error; err == nil {
		stats["free_pages"] = freePages
	}

	// Calculate database size
	if pageCount > 0 && pageSize > 0 {
		stats["database_size_bytes"] = pageCount * pageSize
		stats["free_space_bytes"] = freePages * pageSize
	}

	// Get connection pool stats
	sqlDB, err := db.DB.DB()
	if err == nil {
		poolStats := sqlDB.Stats()
		stats["open_connections"] = poolStats.OpenConnections
		stats["in_use_connections"] = poolStats.InUse
		stats["idle_connections"] = poolStats.Idle
		stats["wait_count"] = poolStats.WaitCount
		stats["wait_duration"] = poolStats.WaitDuration
		stats["max_idle_closed"] = poolStats.MaxIdleClosed
		stats["max_idle_time_closed"] = poolStats.MaxIdleTimeClosed
		stats["max_lifetime_closed"] = poolStats.MaxLifetimeClosed
	}

	return stats, nil
}