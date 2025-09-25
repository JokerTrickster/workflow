package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// TaskWorker handles RabbitMQ message consumption and task execution
type TaskWorker struct {
	connection   *amqp.Connection
	channel      *amqp.Channel
	queueName    string
	rabbitMQURL  string
	providerFactory *AIProviderFactory
}

// TaskMessage matches the structure from backend
type TaskMessage struct {
	Tasks          string `json:"tasks"`
	RepositoryName string `json:"repository_name"`
	WorkingDir     string `json:"working_dir,omitempty"`
	Interactive    bool   `json:"interactive,omitempty"`
	Cmd            string `json:"cmd,omitempty"`
	ContinueTask   bool   `json:"continue_task,omitempty"`
	Provider       string `json:"provider"`
}

// NewTaskWorker creates a new task worker instance
func NewTaskWorker(rabbitMQURL, queueName string) *TaskWorker {
	return &TaskWorker{
		rabbitMQURL:     rabbitMQURL,
		queueName:       queueName,
		providerFactory: GlobalAIProviderFactory,
	}
}

// Connect establishes connection to RabbitMQ
func (w *TaskWorker) Connect() error {
	conn, err := amqp.Dial(w.rabbitMQURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare queue to ensure it exists
	_, err = ch.QueueDeclare(
		w.queueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	w.connection = conn
	w.channel = ch

	return nil
}

// StartConsuming starts consuming messages from RabbitMQ queue
func (w *TaskWorker) StartConsuming(ctx context.Context) error {
	if w.channel == nil {
		return fmt.Errorf("worker not connected")
	}

	// Set QoS to process one message at a time
	err := w.channel.Qos(1, 0, false)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := w.channel.Consume(
		w.queueName, // queue
		"",          // consumer
		false,       // auto-ack (manual ack for reliability)
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("Task worker started, waiting for messages...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Task worker context cancelled, stopping...")
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				log.Println("Message channel closed")
				return fmt.Errorf("message channel closed")
			}
			w.handleMessage(ctx, msg)
		}
	}
}

// handleMessage processes individual RabbitMQ messages
func (w *TaskWorker) handleMessage(ctx context.Context, msg amqp.Delivery) {
	log.Printf("Received message: %s", string(msg.Body))

	// Parse message
	var taskMsg TaskMessage
	if err := json.Unmarshal(msg.Body, &taskMsg); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		msg.Nack(false, false) // Don't requeue malformed messages
		return
	}

	// Execute task
	err := w.executeTask(ctx, &taskMsg)
	if err != nil {
		log.Printf("Failed to execute task: %v", err)
		msg.Nack(false, true) // Requeue for retry
		return
	}

	// Acknowledge successful processing
	msg.Ack(false)
	log.Printf("Task completed successfully for provider: %s", taskMsg.Provider)
}

// executeTask processes the task using the appropriate AI provider
func (w *TaskWorker) executeTask(ctx context.Context, taskMsg *TaskMessage) error {
	log.Printf("Executing task with provider: %s", taskMsg.Provider)

	// Get the appropriate provider
	provider, exists := w.providerFactory.GetProvider(taskMsg.Provider)
	if !exists {
		return fmt.Errorf("unknown provider: %s", taskMsg.Provider)
	}

	// Check if provider is configured
	if !provider.IsConfigured() {
		return fmt.Errorf("provider %s is not properly configured", taskMsg.Provider)
	}

	// Convert TaskMessage to AITaskRequest
	request := &AITaskRequest{
		Tasks:          taskMsg.Tasks,
		RepositoryName: taskMsg.RepositoryName,
		WorkingDir:     taskMsg.WorkingDir,
		Interactive:    taskMsg.Interactive,
		Cmd:            taskMsg.Cmd,
		ContinueTask:   taskMsg.ContinueTask,
		Timeout:        30 * time.Minute, // Default timeout
		Options:        make(map[string]interface{}),
	}

	// Execute task with timeout context
	taskCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	response, err := provider.ExecuteTask(taskCtx, request)
	if err != nil {
		return fmt.Errorf("provider %s failed to execute task: %w", taskMsg.Provider, err)
	}

	// Log execution results
	log.Printf("Task execution completed:")
	log.Printf("  Provider: %s", taskMsg.Provider)
	log.Printf("  Success: %t", response.Success)
	log.Printf("  Execution Time: %v", response.ExecutionTime)
	log.Printf("  Tokens Used: %d", response.TokensUsed)

	if response.Error != "" {
		log.Printf("  Error: %s", response.Error)
	}

	if !response.Success {
		return fmt.Errorf("task execution failed: %s", response.Error)
	}

	return nil
}

// Close closes the RabbitMQ connections
func (w *TaskWorker) Close() error {
	if w.channel != nil {
		w.channel.Close()
	}
	if w.connection != nil {
		return w.connection.Close()
	}
	return nil
}

// GetAvailableProviders returns list of configured providers
func (w *TaskWorker) GetAvailableProviders() []string {
	return w.providerFactory.GetAvailableProviders()
}