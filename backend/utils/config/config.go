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
	QueueName string `json:"queue_name"`
	Exchange  string `json:"exchange"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Host      string `json:"host"`
	Port      string `json:"port"`
}

type ServerConfig struct {
	Host string
	Port string
}

var GlobalConfig *Config

func InitConfig() error {
	// Build MySQL DSN from individual components
	dbHost := getEnv("MYSQL_HOST", "localhost")
	dbPort := getEnv("MYSQL_PORT", "3306")
	dbName := getEnv("MYSQL_DATABASE", "dev_workflow")
	dbUser := getEnv("MYSQL_USER", "root")
	dbPassword := getEnv("MYSQL_PASSWORD", "")

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
			QueueName: getEnv("RABBITMQ_QUEUE_NAME", "workflow_queue"),
			Exchange:  getEnv("RABBITMQ_EXCHANGE", ""),
			Username:  getEnv("RABBITMQ_USER", "board"),
			Password:  getEnv("RABBITMQ_PASSWORD", "examplepassword"),
			Host:      getEnv("RABBITMQ_HOST", "13.203.37.93"),
			Port:      getEnv("RABBITMQ_PORT", "5672"),
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
