package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	GitHub   GitHubConfig   `json:"github"`
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
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}