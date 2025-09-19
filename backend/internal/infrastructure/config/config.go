package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Environment string         `mapstructure:"environment"`
	Server      ServerConfig   `mapstructure:"server"`
	Database    DatabaseConfig `mapstructure:"database"`
	RabbitMQ    RabbitMQConfig `mapstructure:"rabbitmq"`
	Claude      ClaudeConfig   `mapstructure:"claude"`
	Logging     LoggingConfig  `mapstructure:"logging"`
	Health      HealthConfig   `mapstructure:"health"`
	Timeout     TimeoutConfig  `mapstructure:"timeout"`
}

type ServerConfig struct {
	Port            int           `mapstructure:"port"`
	Host            string        `mapstructure:"host"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Driver           string        `mapstructure:"driver"`
	Path             string        `mapstructure:"path"`
	MaxConnections   int           `mapstructure:"max_connections"`
	MaxIdleConns     int           `mapstructure:"max_idle_connections"`
	ConnMaxLifetime  time.Duration `mapstructure:"connection_max_lifetime"`
	RetryAttempts    int           `mapstructure:"retry_attempts"`
	RetryDelay       time.Duration `mapstructure:"retry_delay"`
}

type RabbitMQConfig struct {
	URL             string        `mapstructure:"url"`
	QueueName       string        `mapstructure:"queue_name"`
	MaxRetries      int           `mapstructure:"max_retries"`
	RetryDelay      time.Duration `mapstructure:"retry_delay"`
	ReconnectDelay  time.Duration `mapstructure:"reconnect_delay"`
	PrefetchCount   int           `mapstructure:"prefetch_count"`
	Durable         bool          `mapstructure:"durable"`
	AutoDelete      bool          `mapstructure:"auto_delete"`
}

type ClaudeConfig struct {
	APIKey           string        `mapstructure:"api_key"`
	Model            string        `mapstructure:"model"`
	MaxTokens        int           `mapstructure:"max_tokens"`
	Temperature      float64       `mapstructure:"temperature"`
	RequestTimeout   time.Duration `mapstructure:"request_timeout"`
	MaxRetries       int           `mapstructure:"max_retries"`
	RetryDelay       time.Duration `mapstructure:"retry_delay"`
}

type LoggingConfig struct {
	Level          string `mapstructure:"level"`
	Format         string `mapstructure:"format"`
	Output         string `mapstructure:"output"`
	File           string `mapstructure:"file"`
	MaxSize        int    `mapstructure:"max_size"`
	MaxBackups     int    `mapstructure:"max_backups"`
	MaxAge         int    `mapstructure:"max_age"`
	EnableCaller   bool   `mapstructure:"enable_caller"`
	EnableStacktrace bool `mapstructure:"enable_stacktrace"`
}

type HealthConfig struct {
	Port     int           `mapstructure:"port"`
	Interval time.Duration `mapstructure:"interval"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

type TimeoutConfig struct {
	Request    time.Duration `mapstructure:"request"`
	Database   time.Duration `mapstructure:"database"`
	Queue      time.Duration `mapstructure:"queue"`
	Claude     time.Duration `mapstructure:"claude"`
	Shutdown   time.Duration `mapstructure:"shutdown"`
}

func LoadConfig() (*Config, error) {
	// Determine environment
	env := strings.ToLower(os.Getenv("ENV"))
	if env == "" {
		env = "development"
	}

	// Setup viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("../../configs")

	// Environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("LOCAL_BACKEND")

	// Set defaults based on environment
	setDefaults(env)

	// Load environment-specific configuration
	if err := loadEnvironmentConfig(env); err != nil {
		return nil, fmt.Errorf("failed to load environment config: %w", err)
	}

	// Read environment variables for sensitive data
	loadSensitiveEnvVars()

	// Read main config file
	if err := viper.ReadInConfig(); err != nil {
		// If config file not found, use environment variables and defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set environment in config
	cfg.Environment = env

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

func setDefaults(env string) {
	// Server defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.read_timeout", "15s")
	viper.SetDefault("server.write_timeout", "15s")
	viper.SetDefault("server.shutdown_timeout", "30s")

	// Database defaults
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.path", "./data/workflow.db")
	viper.SetDefault("database.max_connections", 10)
	viper.SetDefault("database.max_idle_connections", 5)
	viper.SetDefault("database.connection_max_lifetime", "1h")
	viper.SetDefault("database.retry_attempts", 3)
	viper.SetDefault("database.retry_delay", "1s")

	// RabbitMQ defaults
	viper.SetDefault("rabbitmq.url", "amqp://localhost:5672")
	viper.SetDefault("rabbitmq.queue_name", "workflow_queue")
	viper.SetDefault("rabbitmq.max_retries", 3)
	viper.SetDefault("rabbitmq.retry_delay", "5s")
	viper.SetDefault("rabbitmq.reconnect_delay", "10s")
	viper.SetDefault("rabbitmq.prefetch_count", 1)
	viper.SetDefault("rabbitmq.durable", true)
	viper.SetDefault("rabbitmq.auto_delete", false)

	// Claude defaults
	viper.SetDefault("claude.model", "claude-3-sonnet-20240229")
	viper.SetDefault("claude.max_tokens", 4096)
	viper.SetDefault("claude.temperature", 0.7)
	viper.SetDefault("claude.request_timeout", "120s")
	viper.SetDefault("claude.max_retries", 3)
	viper.SetDefault("claude.retry_delay", "2s")

	// Logging defaults
	setLoggingDefaults(env)

	// Health defaults
	viper.SetDefault("health.port", 8081)
	viper.SetDefault("health.interval", "30s")
	viper.SetDefault("health.timeout", "10s")

	// Timeout defaults
	viper.SetDefault("timeout.request", "30s")
	viper.SetDefault("timeout.database", "5s")
	viper.SetDefault("timeout.queue", "10s")
	viper.SetDefault("timeout.claude", "120s")
	viper.SetDefault("timeout.shutdown", "30s")
}

func setLoggingDefaults(env string) {
	switch env {
	case "production":
		viper.SetDefault("logging.level", "info")
		viper.SetDefault("logging.format", "json")
		viper.SetDefault("logging.output", "file")
		viper.SetDefault("logging.file", "./logs/app.log")
		viper.SetDefault("logging.enable_caller", false)
		viper.SetDefault("logging.enable_stacktrace", false)
	case "staging":
		viper.SetDefault("logging.level", "debug")
		viper.SetDefault("logging.format", "json")
		viper.SetDefault("logging.output", "stdout")
		viper.SetDefault("logging.enable_caller", true)
		viper.SetDefault("logging.enable_stacktrace", false)
	default: // development
		viper.SetDefault("logging.level", "debug")
		viper.SetDefault("logging.format", "console")
		viper.SetDefault("logging.output", "stdout")
		viper.SetDefault("logging.enable_caller", true)
		viper.SetDefault("logging.enable_stacktrace", true)
	}
	
	viper.SetDefault("logging.max_size", 100)    // 100MB
	viper.SetDefault("logging.max_backups", 3)
	viper.SetDefault("logging.max_age", 28)      // 28 days
}

func loadEnvironmentConfig(env string) error {
	// Try to load environment-specific config file
	envConfigName := fmt.Sprintf("config.%s", env)
	viper.SetConfigName(envConfigName)
	
	// If environment-specific config exists, merge it
	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
		// Environment-specific config not found, continue with defaults
	}
	
	return nil
}

func loadSensitiveEnvVars() {
	// Claude API Key
	if apiKey := os.Getenv("CLAUDE_API_KEY"); apiKey != "" {
		viper.Set("claude.api_key", apiKey)
	}

	// Database credentials (if using external DB in future)
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		viper.Set("database.user", dbUser)
	}
	if dbPass := os.Getenv("DB_PASSWORD"); dbPass != "" {
		viper.Set("database.password", dbPass)
	}

	// RabbitMQ credentials
	if rmqURL := os.Getenv("RABBITMQ_URL"); rmqURL != "" {
		viper.Set("rabbitmq.url", rmqURL)
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	var errors []string

	// Validate server configuration
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		errors = append(errors, "server.port must be between 1 and 65535")
	}
	if c.Server.Host == "" {
		errors = append(errors, "server.host cannot be empty")
	}

	// Validate database configuration
	if c.Database.Driver == "" {
		errors = append(errors, "database.driver cannot be empty")
	}
	if c.Database.Path == "" {
		errors = append(errors, "database.path cannot be empty")
	}
	if c.Database.MaxConnections <= 0 {
		errors = append(errors, "database.max_connections must be greater than 0")
	}

	// Validate RabbitMQ configuration
	if c.RabbitMQ.URL == "" {
		errors = append(errors, "rabbitmq.url cannot be empty")
	}
	if c.RabbitMQ.QueueName == "" {
		errors = append(errors, "rabbitmq.queue_name cannot be empty")
	}
	if c.RabbitMQ.MaxRetries < 0 {
		errors = append(errors, "rabbitmq.max_retries cannot be negative")
	}

	// Validate Claude configuration
	if c.Claude.APIKey == "" && c.Environment == "production" {
		errors = append(errors, "claude.api_key is required in production environment")
	}
	if c.Claude.Model == "" {
		errors = append(errors, "claude.model cannot be empty")
	}
	if c.Claude.MaxTokens <= 0 {
		errors = append(errors, "claude.max_tokens must be greater than 0")
	}
	if c.Claude.Temperature < 0 || c.Claude.Temperature > 2 {
		errors = append(errors, "claude.temperature must be between 0 and 2")
	}

	// Validate logging configuration
	validLogLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
	if !contains(validLogLevels, c.Logging.Level) {
		errors = append(errors, fmt.Sprintf("logging.level must be one of: %s", strings.Join(validLogLevels, ", ")))
	}
	validLogFormats := []string{"json", "console"}
	if !contains(validLogFormats, c.Logging.Format) {
		errors = append(errors, fmt.Sprintf("logging.format must be one of: %s", strings.Join(validLogFormats, ", ")))
	}

	// Validate health configuration
	if c.Health.Port <= 0 || c.Health.Port > 65535 {
		errors = append(errors, "health.port must be between 1 and 65535")
	}
	if c.Health.Port == c.Server.Port {
		errors = append(errors, "health.port cannot be the same as server.port")
	}

	// Validate timeout configuration
	if c.Timeout.Request <= 0 {
		errors = append(errors, "timeout.request must be greater than 0")
	}
	if c.Timeout.Database <= 0 {
		errors = append(errors, "timeout.database must be greater than 0")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation errors:\n- %s", strings.Join(errors, "\n- "))
	}

	return nil
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsStaging returns true if running in staging environment
func (c *Config) IsStaging() bool {
	return c.Environment == "staging"
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}