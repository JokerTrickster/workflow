package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

// RabbitMQConfig holds RabbitMQ connection configuration
type RabbitMQConfig struct {
	URL       string `json:"url"`
	QueueName string `json:"queue_name"`
	Exchange  string `json:"exchange"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Host      string `json:"host"`
	Port      string `json:"port"`
}

// RabbitMQConnection represents a RabbitMQ connection
type RabbitMQConnection struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	config   *RabbitMQConfig
	done     chan bool
}

// WorkflowMessage represents a message from the queue
type WorkflowMessage struct {
	Type      string                 `json:"type"`
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id,omitempty"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewRabbitMQConnection creates a new RabbitMQ connection
func NewRabbitMQConnection(config *RabbitMQConfig) (*RabbitMQConnection, error) {
	if config == nil {
		return nil, fmt.Errorf("RabbitMQ config is required")
	}

	conn := &RabbitMQConnection{
		config: config,
		done:   make(chan bool),
	}

	if err := conn.connect(); err != nil {
		return nil, err
	}

	return conn, nil
}

// connect establishes connection to RabbitMQ
func (r *RabbitMQConnection) connect() error {
	log.Printf("Connecting to RabbitMQ: %s", r.config.URL)

	var err error
	r.conn, err = amqp.Dial(r.config.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	r.channel, err = r.conn.Channel()
	if err != nil {
		r.conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare queue
	_, err = r.channel.QueueDeclare(
		r.config.QueueName, // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		r.channel.Close()
		r.conn.Close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	log.Printf("Connected to RabbitMQ successfully, queue: %s", r.config.QueueName)
	return nil
}

// ConsumeMessages starts consuming messages from the queue
func (r *RabbitMQConnection) ConsumeMessages(handler func(WorkflowMessage) error) error {
	msgs, err := r.channel.Consume(
		r.config.QueueName, // queue
		"",                 // consumer
		false,              // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("Starting to consume messages from queue: %s", r.config.QueueName)

	go func() {
		for d := range msgs {
			var msg WorkflowMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				d.Nack(false, false) // Reject message
				continue
			}

			log.Printf("Received message: type=%s, id=%s", msg.Type, msg.ID)

			if err := handler(msg); err != nil {
				log.Printf("Failed to handle message: %v", err)
				d.Nack(false, false) // Reject message
				continue
			}

			d.Ack(false) // Acknowledge message
		}
	}()

	// Wait for done signal
	<-r.done
	return nil
}

// PublishMessage publishes a message to the queue
func (r *RabbitMQConnection) PublishMessage(msg WorkflowMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = r.channel.Publish(
		"",                 // exchange
		r.config.QueueName, // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published message: type=%s, id=%s", msg.Type, msg.ID)
	return nil
}

// Close closes the RabbitMQ connection
func (r *RabbitMQConnection) Close() error {
	close(r.done)

	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			log.Printf("Error closing channel: %v", err)
		}
	}

	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}

	return nil
}

// LoadRabbitMQConfigFromEnv loads RabbitMQ configuration from environment variables
func LoadRabbitMQConfigFromEnv() *RabbitMQConfig {
	return &RabbitMQConfig{
		URL:       getEnvOrDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		QueueName: getEnvOrDefault("RABBITMQ_QUEUE_NAME", "workflow_queue"),
		Exchange:  getEnvOrDefault("RABBITMQ_EXCHANGE", ""),
		Username:  getEnvOrDefault("RABBITMQ_USERNAME", "guest"),
		Password:  getEnvOrDefault("RABBITMQ_PASSWORD", "guest"),
		Host:      getEnvOrDefault("RABBITMQ_HOST", "localhost"),
		Port:      getEnvOrDefault("RABBITMQ_PORT", "5672"),
	}
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestRabbitMQConnection tests the connection to RabbitMQ
func TestRabbitMQConnection(config *RabbitMQConfig) error {
	conn, err := NewRabbitMQConnection(config)
	if err != nil {
		return err
	}
	defer conn.Close()

	log.Println("RabbitMQ connection test successful!")
	return nil
}