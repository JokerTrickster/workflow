package config

import (
	"fmt"
	"os"
)

type Config struct {
	Database DatabaseConfig
	RabbitMQ RabbitMQConfig
	Server   ServerConfig
}

type DatabaseConfig struct {
	DSN      string
	Host     string
	Port     string
	Name     string
	User     string
	Password string
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
	// Build MySQL DSN from individual components
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "dev_workflow")
	dbUser := getEnv("DB_USERNAME", "root")
	dbPassword := getEnv("DB_PASSWORD", "")

	// MySQL DSN format: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	mysqlDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	GlobalConfig = &Config{
		Database: DatabaseConfig{
			DSN:      mysqlDSN,
			Host:     dbHost,
			Port:     dbPort,
			Name:     dbName,
			User:     dbUser,
			Password: dbPassword,
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
