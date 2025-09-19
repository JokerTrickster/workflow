package config

import (
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	Claude   ClaudeConfig   `mapstructure:"claude"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Health   HealthConfig   `mapstructure:"health"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type DatabaseConfig struct {
	Driver         string `mapstructure:"driver"`
	Path           string `mapstructure:"path"`
	MaxConnections int    `mapstructure:"max_connections"`
}

type RabbitMQConfig struct {
	URL       string `mapstructure:"url"`
	QueueName string `mapstructure:"queue_name"`
	MaxRetries int   `mapstructure:"max_retries"`
}

type ClaudeConfig struct {
	APIKey    string `mapstructure:"api_key"`
	Model     string `mapstructure:"model"`
	MaxTokens int    `mapstructure:"max_tokens"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type HealthConfig struct {
	Port int `mapstructure:"port"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("../../configs")

	// Environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("LOCAL_BACKEND")

	// Set defaults
	setDefaults()

	// Read environment variables for sensitive data
	if apiKey := os.Getenv("CLAUDE_API_KEY"); apiKey != "" {
		viper.Set("claude.api_key", apiKey)
	}

	if err := viper.ReadInConfig(); err != nil {
		// If config file not found, use environment variables and defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults() {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.path", "./data/workflow.db")
	viper.SetDefault("database.max_connections", 10)
	viper.SetDefault("rabbitmq.url", "amqp://localhost:5672")
	viper.SetDefault("rabbitmq.queue_name", "workflow_queue")
	viper.SetDefault("rabbitmq.max_retries", 3)
	viper.SetDefault("claude.model", "claude-3-sonnet-20240229")
	viper.SetDefault("claude.max_tokens", 4096)
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("health.port", 8081)
}