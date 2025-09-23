package config

import (
	"os"
)

type Config struct {
	Database DatabaseConfig
	RabbitMQ RabbitMQConfig
	Server   ServerConfig
}

type DatabaseConfig struct {
	DSN string
}

type RabbitMQConfig struct {
	URL       string
	QueueName string
}

type ServerConfig struct {
	Host string
	Port string
}

var GlobalConfig *Config

func InitConfig() error {
	GlobalConfig = &Config{
		Database: DatabaseConfig{
			DSN: getEnv("DATABASE_DSN", "./data/workflow.db"),
		},
		RabbitMQ: RabbitMQConfig{
			URL:       getEnv("RABBITMQ_URL", "amqp://localhost:5672"),
			QueueName: getEnv("RABBITMQ_QUEUE_NAME", "workflow_tasks"),
		},
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}