package infrastructure

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	Claude   ClaudeConfig   `mapstructure:"claude"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	Driver       string `mapstructure:"driver"`
	DSN          string `mapstructure:"dsn"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	LogLevel     string `mapstructure:"log_level"`
}

type RabbitMQConfig struct {
	URL         string `mapstructure:"url"`
	QueueName   string `mapstructure:"queue_name"`
	Exchange    string `mapstructure:"exchange"`
	RoutingKey  string `mapstructure:"routing_key"`
	ConsumerTag string `mapstructure:"consumer_tag"`
}

type ClaudeConfig struct {
	APIKey    string `mapstructure:"api_key"`
	Model     string `mapstructure:"model"`
	MaxTokens int    `mapstructure:"max_tokens"`
	Timeout   int    `mapstructure:"timeout"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig() (*Config, error) {
	// Set default config file path
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("../../configs")

	// Environment variable settings
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Environment variable mappings
	viper.BindEnv("database.dsn", "DATABASE_DSN")
	viper.BindEnv("rabbitmq.url", "RABBITMQ_URL")
	viper.BindEnv("rabbitmq.queue_name", "RABBITMQ_QUEUE")
	viper.BindEnv("claude.api_key", "CLAUDE_API_KEY")
	viper.BindEnv("server.host", "SERVER_HOST")
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("logging.level", "LOG_LEVEL")
	viper.BindEnv("logging.format", "LOG_FORMAT")
	viper.BindEnv("app.environment", "APP_ENV")

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// validateConfig validates required configuration values
func validateConfig(config *Config) error {
	if config.Database.DSN == "" {
		return fmt.Errorf("database DSN is required")
	}
	
	if config.RabbitMQ.URL == "" {
		return fmt.Errorf("RabbitMQ URL is required")
	}
	
	if config.RabbitMQ.QueueName == "" {
		return fmt.Errorf("RabbitMQ queue name is required")
	}

	// Claude API key is not required for basic startup, but will be needed for processing
	
	return nil
}