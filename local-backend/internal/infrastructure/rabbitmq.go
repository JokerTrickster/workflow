package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"

	"github.com/JokerTrickster/workflow/local-backend/internal/domain"
)

// RabbitMQConsumer implements the QueueConsumer interface
type RabbitMQConsumer struct {
	config     *RabbitMQConfig
	connection *amqp.Connection
	channel    *amqp.Channel
	done       chan bool
	isRunning  bool
}

// NewRabbitMQConsumer creates a new RabbitMQ consumer
func NewRabbitMQConsumer(config *RabbitMQConfig) (*RabbitMQConsumer, error) {
	consumer := &RabbitMQConsumer{
		config: config,
		done:   make(chan bool),
	}

	if err := consumer.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return consumer, nil
}

// connect establishes connection to RabbitMQ
func (r *RabbitMQConsumer) connect() error {
	var err error
	
	// Establish connection
	r.connection, err = amqp.Dial(r.config.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create channel
	r.channel, err = r.connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare queue (idempotent)
	_, err = r.channel.QueueDeclare(
		r.config.QueueName, // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Set QoS to process one message at a time
	err = r.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	log.Printf("Connected to RabbitMQ: %s, Queue: %s", r.config.URL, r.config.QueueName)
	return nil
}

// StartConsuming begins consuming messages from the queue
func (r *RabbitMQConsumer) StartConsuming(ctx context.Context, handler domain.MessageHandler) error {
	if r.isRunning {
		return fmt.Errorf("consumer is already running")
	}

	// Register consumer
	messages, err := r.channel.Consume(
		r.config.QueueName,   // queue
		r.config.ConsumerTag, // consumer
		false,                // auto-ack (we'll manually ack)
		false,                // exclusive
		false,                // no-local
		false,                // no-wait
		nil,                  // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	r.isRunning = true
	log.Printf("Started consuming messages from queue: %s", r.config.QueueName)

	// Connection error handling
	closeChan := make(chan *amqp.Error)
	r.connection.NotifyClose(closeChan)

	// Process messages
	go func() {
		defer func() {
			r.isRunning = false
			log.Println("Message consumer stopped")
		}()

		for {
			select {
			case <-ctx.Done():
				log.Println("Context cancelled, stopping consumer")
				return

			case <-r.done:
				log.Println("Consumer shutdown signal received")
				return

			case err := <-closeChan:
				if err != nil {
					log.Printf("RabbitMQ connection closed: %v", err)
					// Attempt to reconnect
					if reconnectErr := r.reconnect(); reconnectErr != nil {
						log.Printf("Failed to reconnect: %v", reconnectErr)
						return
					}
					// Reset the close channel
					closeChan = make(chan *amqp.Error)
					r.connection.NotifyClose(closeChan)
				}

			case delivery, ok := <-messages:
				if !ok {
					log.Println("Messages channel closed")
					return
				}

				// Process message
				if err := r.processMessage(ctx, delivery, handler); err != nil {
					log.Printf("Failed to process message: %v", err)
					// Negative acknowledgment with requeue
					delivery.Nack(false, true)
				} else {
					// Positive acknowledgment
					delivery.Ack(false)
				}
			}
		}
	}()

	return nil
}

// processMessage handles a single message
func (r *RabbitMQConsumer) processMessage(ctx context.Context, delivery amqp.Delivery, handler domain.MessageHandler) error {
	log.Printf("Received message: %s", delivery.Body)

	// Parse message
	message, err := r.parseMessage(delivery.Body)
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// Add metadata from delivery
	message.Metadata["delivery_tag"] = fmt.Sprintf("%d", delivery.DeliveryTag)
	message.Metadata["exchange"] = delivery.Exchange
	message.Metadata["routing_key"] = delivery.RoutingKey
	message.Metadata["content_type"] = delivery.ContentType
	
	if delivery.Timestamp != (time.Time{}) {
		message.Metadata["timestamp"] = delivery.Timestamp.Format(time.RFC3339)
	}

	// Handle message
	return handler(ctx, message)
}

// parseMessage converts raw bytes to domain.Message
func (r *RabbitMQConsumer) parseMessage(body []byte) (*domain.Message, error) {
	var rawMessage struct {
		ID        string                 `json:"id"`
		Type      string                 `json:"type"`
		Payload   map[string]interface{} `json:"payload"`
		ContextID *string                `json:"context_id,omitempty"`
	}

	if err := json.Unmarshal(body, &rawMessage); err != nil {
		return nil, fmt.Errorf("invalid JSON message: %w", err)
	}

	// Validate required fields
	if rawMessage.ID == "" {
		return nil, fmt.Errorf("message ID is required")
	}
	if rawMessage.Type == "" {
		return nil, fmt.Errorf("message type is required")
	}
	if rawMessage.Payload == nil {
		rawMessage.Payload = make(map[string]interface{})
	}

	// Validate message type
	var msgType domain.MessageType
	switch rawMessage.Type {
	case string(domain.MessageTypeWorkRequest):
		msgType = domain.MessageTypeWorkRequest
	case string(domain.MessageTypeCancellation):
		msgType = domain.MessageTypeCancellation
	default:
		return nil, fmt.Errorf("unsupported message type: %s", rawMessage.Type)
	}

	// Create domain message
	message := domain.NewMessage(rawMessage.ID, msgType, rawMessage.Payload)
	if rawMessage.ContextID != nil {
		message.SetContextID(*rawMessage.ContextID)
	}

	return message, nil
}

// StopConsuming stops message consumption
func (r *RabbitMQConsumer) StopConsuming() error {
	if !r.isRunning {
		return nil
	}

	log.Println("Stopping RabbitMQ consumer...")
	
	// Cancel consumer
	if r.channel != nil {
		if err := r.channel.Cancel(r.config.ConsumerTag, false); err != nil {
			log.Printf("Failed to cancel consumer: %v", err)
		}
	}

	// Signal done
	select {
	case r.done <- true:
	default:
	}

	// Close channel and connection
	if r.channel != nil {
		r.channel.Close()
	}
	if r.connection != nil {
		r.connection.Close()
	}

	r.isRunning = false
	log.Println("RabbitMQ consumer stopped")
	return nil
}

// reconnect attempts to reconnect to RabbitMQ
func (r *RabbitMQConsumer) reconnect() error {
	log.Println("Attempting to reconnect to RabbitMQ...")
	
	// Close existing connections
	if r.channel != nil {
		r.channel.Close()
	}
	if r.connection != nil {
		r.connection.Close()
	}

	// Wait before reconnecting
	time.Sleep(5 * time.Second)

	// Attempt reconnection with retries
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		if err := r.connect(); err != nil {
			log.Printf("Reconnection attempt %d failed: %v", i+1, err)
			if i == maxRetries-1 {
				return fmt.Errorf("failed to reconnect after %d attempts", maxRetries)
			}
			time.Sleep(time.Duration(i+1) * 5 * time.Second)
			continue
		}
		
		log.Println("Successfully reconnected to RabbitMQ")
		return nil
	}

	return fmt.Errorf("unexpected error in reconnection loop")
}

// Health checks if the RabbitMQ connection is healthy
func (r *RabbitMQConsumer) Health() error {
	if r.connection == nil || r.connection.IsClosed() {
		return fmt.Errorf("RabbitMQ connection is closed")
	}
	if r.channel == nil {
		return fmt.Errorf("RabbitMQ channel is not available")
	}
	return nil
}

// GetStats returns RabbitMQ consumer statistics
func (r *RabbitMQConsumer) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	stats["is_running"] = r.isRunning
	stats["queue_name"] = r.config.QueueName
	stats["consumer_tag"] = r.config.ConsumerTag
	
	if r.connection != nil {
		stats["connection_closed"] = r.connection.IsClosed()
	} else {
		stats["connection_closed"] = true
	}
	
	if r.channel != nil {
		stats["channel_available"] = true
	} else {
		stats["channel_available"] = false
	}
	
	return stats
}