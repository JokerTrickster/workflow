package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"

	"local-backend-server/internal/domain/entities"
	"local-backend-server/internal/infrastructure/config"
)

// WorkflowMessage represents a message from the queue
type WorkflowMessage struct {
	Type      string                 `json:"type"`
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id,omitempty"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

// MessageProcessor defines the interface for processing messages
type MessageProcessor interface {
	ProcessWorkflowMessage(ctx context.Context, msg *WorkflowMessage) error
}

// Consumer handles RabbitMQ message consumption
type Consumer struct {
	conn      *Connection
	processor MessageProcessor
	config    *config.RabbitMQConfig
	done      chan bool
	consuming bool
}

// NewConsumer creates a new message consumer
func NewConsumer(cfg *config.RabbitMQConfig, processor MessageProcessor) (*Consumer, error) {
	conn, err := NewConnection(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	return &Consumer{
		conn:      conn,
		processor: processor,
		config:    cfg,
		done:      make(chan bool),
	}, nil
}

// SetProcessor sets the message processor
func (c *Consumer) SetProcessor(processor MessageProcessor) {
	c.processor = processor
}

// Start begins consuming messages from the queue
func (c *Consumer) Start(ctx context.Context) error {
	if c.consuming {
		return fmt.Errorf("consumer is already running")
	}

	c.consuming = true
	log.Printf("Starting message consumer for queue: %s", c.config.QueueName)

	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer context cancelled, stopping...")
			c.consuming = false
			return ctx.Err()
		case <-c.done:
			log.Println("Consumer stopped")
			c.consuming = false
			return nil
		default:
			if err := c.consume(ctx); err != nil {
				log.Printf("Consumer error: %v", err)
				// Wait before retrying
				time.Sleep(5 * time.Second)
			}
		}
	}
}

// consume handles the actual message consumption
func (c *Consumer) consume(ctx context.Context) error {
	if !c.conn.IsConnected() {
		return fmt.Errorf("RabbitMQ connection is not active")
	}

	channel := c.conn.GetChannel()
	if channel == nil {
		return fmt.Errorf("RabbitMQ channel is not available")
	}

	// Start consuming messages
	msgs, err := channel.Consume(
		c.config.QueueName, // queue
		"workflow-consumer", // consumer tag
		false,              // auto-ack (we'll ack manually)
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Println("Message consumer started, waiting for messages...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.done:
			return nil
		case msg, ok := <-msgs:
			if !ok {
				log.Println("Consumer channel closed, attempting to reconnect...")
				return fmt.Errorf("consumer channel closed")
			}

			if err := c.handleMessage(ctx, msg); err != nil {
				log.Printf("Failed to handle message: %v", err)
				// Reject message and requeue it
				msg.Nack(false, true)
			} else {
				// Acknowledge successful processing
				msg.Ack(false)
			}
		}
	}
}

// handleMessage processes a single message
func (c *Consumer) handleMessage(ctx context.Context, delivery amqp.Delivery) error {
	log.Printf("Received message: %s", string(delivery.Body))

	// Parse the message
	var workflowMsg WorkflowMessage
	if err := json.Unmarshal(delivery.Body, &workflowMsg); err != nil {
		return fmt.Errorf("failed to parse message JSON: %w", err)
	}

	// Validate message
	if err := c.validateMessage(&workflowMsg); err != nil {
		return fmt.Errorf("message validation failed: %w", err)
	}

	// Set timestamp if not provided
	if workflowMsg.Timestamp.IsZero() {
		workflowMsg.Timestamp = time.Now()
	}

	// Route message to appropriate processor
	return c.routeMessage(ctx, &workflowMsg)
}

// validateMessage validates the structure and content of a workflow message
func (c *Consumer) validateMessage(msg *WorkflowMessage) error {
	if msg.Type == "" {
		return fmt.Errorf("message type is required")
	}

	if msg.ID == "" {
		return fmt.Errorf("message ID is required")
	}

	if msg.Payload == nil {
		return fmt.Errorf("message payload is required")
	}

	// Validate message type
	switch msg.Type {
	case string(entities.MessageTypeWorkRequest):
		return c.validateWorkRequest(msg)
	case string(entities.MessageTypeCancel):
		return c.validateCancelRequest(msg)
	case string(entities.MessageTypeStatus):
		return c.validateStatusRequest(msg)
	default:
		return fmt.Errorf("unsupported message type: %s", msg.Type)
	}
}

// validateWorkRequest validates a work request message
func (c *Consumer) validateWorkRequest(msg *WorkflowMessage) error {
	if msg.SessionID == "" {
		return fmt.Errorf("session_id is required for work requests")
	}

	if _, ok := msg.Payload["request_type"]; !ok {
		return fmt.Errorf("request_type is required in work request payload")
	}

	if _, ok := msg.Payload["input"]; !ok {
		return fmt.Errorf("input is required in work request payload")
	}

	return nil
}

// validateCancelRequest validates a cancel request message
func (c *Consumer) validateCancelRequest(msg *WorkflowMessage) error {
	if _, ok := msg.Payload["request_id"]; !ok {
		return fmt.Errorf("request_id is required in cancel request payload")
	}

	return nil
}

// validateStatusRequest validates a status request message
func (c *Consumer) validateStatusRequest(msg *WorkflowMessage) error {
	if _, ok := msg.Payload["request_id"]; !ok {
		return fmt.Errorf("request_id is required in status request payload")
	}

	return nil
}

// routeMessage routes the message to the appropriate processor
func (c *Consumer) routeMessage(ctx context.Context, msg *WorkflowMessage) error {
	log.Printf("Routing message type: %s, ID: %s", msg.Type, msg.ID)

	// Process the message using the configured processor
	if err := c.processor.ProcessWorkflowMessage(ctx, msg); err != nil {
		return fmt.Errorf("failed to process message: %w", err)
	}

	log.Printf("Successfully processed message: %s", msg.ID)
	return nil
}

// Stop stops the consumer
func (c *Consumer) Stop() error {
	if !c.consuming {
		return nil
	}

	log.Println("Stopping message consumer...")
	close(c.done)

	// Wait for consumer to stop
	timeout := time.After(10 * time.Second)
	for c.consuming {
		select {
		case <-timeout:
			log.Println("Consumer stop timeout reached")
			break
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	return c.conn.Close()
}

// Health checks the consumer health
func (c *Consumer) Health() error {
	if !c.consuming {
		return fmt.Errorf("consumer is not running")
	}

	return c.conn.Health()
}

// GetQueueInfo returns information about the queue
func (c *Consumer) GetQueueInfo() (*QueueInfo, error) {
	if !c.conn.IsConnected() {
		return nil, fmt.Errorf("RabbitMQ connection is not active")
	}

	channel := c.conn.GetChannel()
	if channel == nil {
		return nil, fmt.Errorf("RabbitMQ channel is not available")
	}

	queue, err := channel.QueueInspect(c.config.QueueName)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect queue: %w", err)
	}

	return &QueueInfo{
		Name:      queue.Name,
		Messages:  queue.Messages,
		Consumers: queue.Consumers,
	}, nil
}

// QueueInfo holds information about a queue
type QueueInfo struct {
	Name      string `json:"name"`
	Messages  int    `json:"messages"`
	Consumers int    `json:"consumers"`
}