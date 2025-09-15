package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	GitHub   GitHubConfig   `json:"github"`
	RabbitMQ RabbitMQConfig `json:"rabbitmq"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string `json:"port"`
	Host string `json:"host"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Charset  string `json:"charset"`
}

// GitHubConfig holds GitHub configuration
type GitHubConfig struct {
	Token      string `json:"token"`
	WebhookURL string `json:"webhook_url"`
}

// RabbitMQConfig holds RabbitMQ configuration
type RabbitMQConfig struct {
	Host         string `json:"host"`
	Port         string `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	VHost        string `json:"vhost"`
	Queue        string `json:"queue"`
	Exchange     string `json:"exchange"`
	RoutingKey   string `json:"routing_key"`
	Durable      bool   `json:"durable"`
	AutoDelete   bool   `json:"auto_delete"`
	Exclusive    bool   `json:"exclusive"`
	NoWait       bool   `json:"no_wait"`
	ReconnectDelay string `json:"reconnect_delay"`
	MaxRetries   int    `json:"max_retries"`
}

// Load loads configuration from environment variables
func Load() *Config {
	// Try to load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "workflow"),
			Charset:  getEnv("DB_CHARSET", "utf8mb4"),
		},
		GitHub: GitHubConfig{
			Token:      getEnv("GITHUB_TOKEN", ""),
			WebhookURL: getEnv("GITHUB_WEBHOOK_URL", ""),
		},
		RabbitMQ: RabbitMQConfig{
			Host:         getEnv("RABBITMQ_HOST", "13.203.37.93"),
			Port:         getEnv("RABBITMQ_PORT", "5672"),
			Username:     getEnv("RABBITMQ_USERNAME", "guest"),
			Password:     getEnv("RABBITMQ_PASSWORD", "guest"),
			VHost:        getEnv("RABBITMQ_VHOST", "/"),
			Queue:        getEnv("RABBITMQ_QUEUE", "claude-tasks"),
			Exchange:     getEnv("RABBITMQ_EXCHANGE", ""),
			RoutingKey:   getEnv("RABBITMQ_ROUTING_KEY", "claude-tasks"),
			Durable:      getEnv("RABBITMQ_DURABLE", "true") == "true",
			AutoDelete:   getEnv("RABBITMQ_AUTO_DELETE", "false") == "true",
			Exclusive:    getEnv("RABBITMQ_EXCLUSIVE", "false") == "true",
			NoWait:       getEnv("RABBITMQ_NO_WAIT", "false") == "true",
			ReconnectDelay: getEnv("RABBITMQ_RECONNECT_DELAY", "5s"),
			MaxRetries:   parseInt(getEnv("RABBITMQ_MAX_RETRIES", "5")),
		},
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseInt converts string to int with fallback
func parseInt(value string) int {
	if result, err := strconv.Atoi(value); err == nil {
		return result
	}
	return 0
}