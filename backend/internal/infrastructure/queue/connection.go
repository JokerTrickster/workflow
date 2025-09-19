package queue

import (
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"

	"local-backend-server/internal/infrastructure/config"
)

// Connection wraps the RabbitMQ connection
type Connection struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  *config.RabbitMQConfig
	done    chan bool
}

// NewConnection creates a new RabbitMQ connection
func NewConnection(cfg *config.RabbitMQConfig) (*Connection, error) {
	c := &Connection{
		config: cfg,
		done:   make(chan bool),
	}

	if err := c.connect(); err != nil {
		return nil, err
	}

	return c, nil
}

// connect establishes connection to RabbitMQ
func (c *Connection) connect() error {
	var err error

	log.Printf("Connecting to RabbitMQ: %s", c.config.URL)

	c.conn, err = amqp.Dial(c.config.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	c.channel, err = c.conn.Channel()
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Set QoS to process one message at a time
	if err := c.channel.Qos(1, 0, false); err != nil {
		c.close()
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Declare the queue
	_, err = c.channel.QueueDeclare(
		c.config.QueueName, // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		c.close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	log.Printf("Connected to RabbitMQ successfully, queue: %s", c.config.QueueName)

	// Handle connection close events
	go c.handleReconnect()

	return nil
}

// handleReconnect handles reconnection when connection is lost
func (c *Connection) handleReconnect() {
	notifyClose := make(chan *amqp.Error)
	c.conn.NotifyClose(notifyClose)

	select {
	case err := <-notifyClose:
		if err != nil {
			log.Printf("RabbitMQ connection lost: %v", err)
			c.reconnect()
		}
	case <-c.done:
		return
	}
}

// reconnect attempts to reconnect to RabbitMQ with exponential backoff
func (c *Connection) reconnect() {
	retryCount := 0
	maxRetries := c.config.MaxRetries

	for retryCount < maxRetries {
		retryCount++
		backoff := time.Duration(retryCount) * time.Second

		log.Printf("Attempting to reconnect to RabbitMQ (attempt %d/%d) in %v", 
			retryCount, maxRetries, backoff)

		time.Sleep(backoff)

		if err := c.connect(); err != nil {
			log.Printf("Reconnection attempt %d failed: %v", retryCount, err)
			continue
		}

		log.Println("Reconnected to RabbitMQ successfully")
		return
	}

	log.Printf("Failed to reconnect to RabbitMQ after %d attempts", maxRetries)
}

// GetChannel returns the RabbitMQ channel
func (c *Connection) GetChannel() *amqp.Channel {
	return c.channel
}

// IsConnected checks if the connection is active
func (c *Connection) IsConnected() bool {
	return c.conn != nil && !c.conn.IsClosed()
}

// Close closes the RabbitMQ connection
func (c *Connection) Close() error {
	close(c.done)
	return c.close()
}

// close internal method to close connection and channel
func (c *Connection) close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			log.Printf("Error closing channel: %v", err)
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
			return err
		}
	}

	return nil
}

// Health checks the connection health
func (c *Connection) Health() error {
	if !c.IsConnected() {
		return fmt.Errorf("RabbitMQ connection is not active")
	}

	// Try to declare a temporary queue to test the connection
	tempQueueName := fmt.Sprintf("health_check_%d", time.Now().Unix())
	_, err := c.channel.QueueDeclare(
		tempQueueName,
		false, // not durable
		true,  // auto delete
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		return fmt.Errorf("RabbitMQ health check failed: %w", err)
	}

	// Delete the temporary queue
	_, err = c.channel.QueueDelete(tempQueueName, false, false, false)
	if err != nil {
		log.Printf("Warning: failed to delete health check queue: %v", err)
	}

	return nil
}