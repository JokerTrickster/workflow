package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/streadway/amqp"

	"local-backend-server/internal/infrastructure/config"
)

// Publisher handles RabbitMQ message publishing
type Publisher struct {
	conn   *Connection
	config *config.RabbitMQConfig
}

// NewPublisher creates a new message publisher
func NewPublisher(cfg *config.RabbitMQConfig) (*Publisher, error) {
	conn, err := NewConnection(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	return &Publisher{
		conn:   conn,
		config: cfg,
	}, nil
}

// PublishMessage publishes a workflow message to the queue
func (p *Publisher) PublishMessage(msg *WorkflowMessage) error {
	if p.conn == nil {
		return fmt.Errorf("publisher connection is nil")
	}

	channel := p.conn.GetChannel()
	if channel == nil {
		return fmt.Errorf("publisher channel is nil")
	}

	// Set timestamp if not provided
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	// Marshal message to JSON
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Publish message
	err = channel.Publish(
		"",                     // exchange
		p.config.QueueName,     // routing key (queue name)
		false,                  // mandatory
		false,                  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Close closes the publisher connection
func (p *Publisher) Close() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

// PublishError represents a queue publishing error
type PublishError struct {
	Message string
}

func (e *PublishError) Error() string {
	return e.Message
}