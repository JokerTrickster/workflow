package utils

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// ConfigManager handles configuration storage in database
type ConfigManager struct {
	db *sql.DB
}

// ConfigEntry represents a configuration entry
type ConfigEntry struct {
	ID          int    `json:"id"`
	ConfigKey   string `json:"config_key"`
	ConfigValue string `json:"config_value"`
	Description string `json:"description"`
	Category    string `json:"category"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// NewConfigManager creates a new config manager
func NewConfigManager(dsn string) (*ConfigManager, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	cm := &ConfigManager{db: db}
	
	// Create table if not exists
	if err := cm.createConfigTable(); err != nil {
		return nil, fmt.Errorf("failed to create config table: %w", err)
	}

	return cm, nil
}

// createConfigTable creates the configuration table
func (cm *ConfigManager) createConfigTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS app_configs (
		id INT AUTO_INCREMENT PRIMARY KEY,
		config_key VARCHAR(255) NOT NULL UNIQUE,
		config_value TEXT NOT NULL,
		description VARCHAR(500),
		category VARCHAR(100) DEFAULT 'general',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_category (category),
		INDEX idx_config_key (config_key)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	_, err := cm.db.Exec(query)
	return err
}

// SetConfig sets a configuration value
func (cm *ConfigManager) SetConfig(key, value, description, category string) error {
	query := `
	INSERT INTO app_configs (config_key, config_value, description, category)
	VALUES (?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE 
		config_value = VALUES(config_value),
		description = VALUES(description),
		category = VALUES(category),
		updated_at = CURRENT_TIMESTAMP
	`

	_, err := cm.db.Exec(query, key, value, description, category)
	if err != nil {
		return fmt.Errorf("failed to set config %s: %w", key, err)
	}

	log.Printf("Config set: %s = %s", key, value)
	return nil
}

// GetConfig gets a configuration value
func (cm *ConfigManager) GetConfig(key string) (string, error) {
	var value string
	query := "SELECT config_value FROM app_configs WHERE config_key = ?"
	
	err := cm.db.QueryRow(query, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("config key '%s' not found", key)
		}
		return "", fmt.Errorf("failed to get config %s: %w", key, err)
	}

	return value, nil
}

// GetConfigWithDefault gets a configuration value with default fallback
func (cm *ConfigManager) GetConfigWithDefault(key, defaultValue string) string {
	value, err := cm.GetConfig(key)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetConfigsByCategory gets all configurations by category
func (cm *ConfigManager) GetConfigsByCategory(category string) ([]ConfigEntry, error) {
	query := `
	SELECT id, config_key, config_value, description, category, created_at, updated_at
	FROM app_configs 
	WHERE category = ?
	ORDER BY config_key
	`

	rows, err := cm.db.Query(query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to query configs by category %s: %w", category, err)
	}
	defer rows.Close()

	var configs []ConfigEntry
	for rows.Next() {
		var config ConfigEntry
		err := rows.Scan(
			&config.ID,
			&config.ConfigKey,
			&config.ConfigValue,
			&config.Description,
			&config.Category,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan config row: %w", err)
		}
		configs = append(configs, config)
	}

	return configs, nil
}

// LoadRabbitMQConfigFromDB loads RabbitMQ configuration from database
func (cm *ConfigManager) LoadRabbitMQConfigFromDB() (*RabbitMQConfig, error) {
	configs, err := cm.GetConfigsByCategory("rabbitmq")
	if err != nil {
		return nil, err
	}

	// Create config map
	configMap := make(map[string]string)
	for _, config := range configs {
		configMap[config.ConfigKey] = config.ConfigValue
	}

	// Build RabbitMQ config
	config := &RabbitMQConfig{
		URL:       getConfigValue(configMap, "rabbitmq.url", "amqp://guest:guest@localhost:5672/"),
		QueueName: getConfigValue(configMap, "rabbitmq.queue_name", "workflow_queue"),
		Exchange:  getConfigValue(configMap, "rabbitmq.exchange", ""),
		Username:  getConfigValue(configMap, "rabbitmq.user", "guest"),
		Password:  getConfigValue(configMap, "rabbitmq.password", "guest"),
		Host:      getConfigValue(configMap, "rabbitmq.host", "localhost"),
		Port:      getConfigValue(configMap, "rabbitmq.port", "5672"),
	}

	return config, nil
}

// SaveRabbitMQConfigToDB saves RabbitMQ configuration to database
func (cm *ConfigManager) SaveRabbitMQConfigToDB(config *RabbitMQConfig) error {
	configs := []struct {
		key, value, description string
	}{
		{"rabbitmq.url", config.URL, "RabbitMQ connection URL"},
		{"rabbitmq.queue_name", config.QueueName, "RabbitMQ queue name"},
		{"rabbitmq.exchange", config.Exchange, "RabbitMQ exchange name"},
		{"rabbitmq.username", config.Username, "RabbitMQ username"},
		{"rabbitmq.password", config.Password, "RabbitMQ password"},
		{"rabbitmq.host", config.Host, "RabbitMQ host"},
		{"rabbitmq.port", config.Port, "RabbitMQ port"},
	}

	for _, cfg := range configs {
		if err := cm.SetConfig(cfg.key, cfg.value, cfg.description, "rabbitmq"); err != nil {
			return err
		}
	}

	log.Println("RabbitMQ configuration saved to database")
	return nil
}

// LoadConfigFromEnvAndSaveToDB loads config from .env and saves to database
func (cm *ConfigManager) LoadConfigFromEnvAndSaveToDB() error {
	log.Println("Loading configuration from .env and saving to database...")

	// RabbitMQ configuration
	rabbitMQConfig := LoadRabbitMQConfigFromEnv()
	if err := cm.SaveRabbitMQConfigToDB(rabbitMQConfig); err != nil {
		return fmt.Errorf("failed to save RabbitMQ config: %w", err)
	}

	// Database configuration
	dbConfigs := []struct {
		key, value, description string
	}{
		{"db.host", getEnvOrDefault("DB_HOST", "localhost"), "Database host"},
		{"db.port", getEnvOrDefault("DB_PORT", "3306"), "Database port"},
		{"db.name", getEnvOrDefault("DB_NAME", "dev_workflow"), "Database name"},
		{"db.username", getEnvOrDefault("DB_USERNAME", "root"), "Database username"},
		{"db.password", getEnvOrDefault("DB_PASSWORD", ""), "Database password"},
	}

	for _, cfg := range dbConfigs {
		if err := cm.SetConfig(cfg.key, cfg.value, cfg.description, "database"); err != nil {
			return err
		}
	}

	// Application configuration
	appConfigs := []struct {
		key, value, description string
	}{
		{"app.port", getEnvOrDefault("PORT", "8083"), "Application port"},
		{"app.host", getEnvOrDefault("SERVER_HOST", "localhost"), "Application host"},
		{"app.environment", getEnvOrDefault("ENVIRONMENT", "development"), "Application environment"},
		{"app.log_level", getEnvOrDefault("LOG_LEVEL", "debug"), "Log level"},
		{"app.working_dir", getEnvOrDefault("WORKING_DIR", "/tmp/claude-tasks"), "Working directory"},
	}

	for _, cfg := range appConfigs {
		if err := cm.SetConfig(cfg.key, cfg.value, cfg.description, "application"); err != nil {
			return err
		}
	}

	log.Println("✅ All configurations saved to database successfully!")
	return nil
}

// Close closes the database connection
func (cm *ConfigManager) Close() error {
	if cm.db != nil {
		return cm.db.Close()
	}
	return nil
}

// getConfigValue gets config value from map with default
func getConfigValue(configMap map[string]string, key, defaultValue string) string {
	if value, exists := configMap[key]; exists && value != "" {
		return value
	}
	return defaultValue
}

// CreateConfigManagerFromEnv creates config manager using environment variables
func CreateConfigManagerFromEnv() (*ConfigManager, error) {
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "3306")
	dbname := getEnvOrDefault("DB_NAME", "dev_workflow")
	username := getEnvOrDefault("DB_USERNAME", "root")
	password := getEnvOrDefault("DB_PASSWORD", "")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, host, port, dbname)

	return NewConfigManager(dsn)
}