package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"main/utils/db/mysql"
	"strings"
	"time"

	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

// TaskWorker handles RabbitMQ message consumption and task execution
type TaskWorker struct {
	connection   *amqp.Connection
	channel      *amqp.Channel
	queueName    string
	rabbitMQURL  string
	providerFactory *AIProviderFactory
	failureCount map[string]int       // Track failures per provider
	lastFailureTime map[string]time.Time // Track last failure time per provider
	maxRetries   int                  // Maximum retry attempts before skipping
	db           *gorm.DB             // Database connection for logging
}

// TaskMessage matches the structure from backend
type TaskMessage struct {
	RequestID      string `json:"request_id"`
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
	// Safely get database connection
	var db *gorm.DB
	if mysql.GormMysqlDB != nil {
		db = mysql.GormMysqlDB
		log.Printf("TaskWorker: Database connection available")
	} else {
		log.Printf("TaskWorker: Database connection not available, failure logging will be disabled")
	}

	return &TaskWorker{
		rabbitMQURL:     rabbitMQURL,
		queueName:       queueName,
		providerFactory: GlobalAIProviderFactory,
		failureCount:    make(map[string]int),
		lastFailureTime: make(map[string]time.Time),
		maxRetries:      5, // Skip after 5 consecutive failures
		db:              db,
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

		// Record failure to database before acknowledging
		w.recordTaskFailure(&taskMsg, err.Error())

		// Always acknowledge to remove from queue, don't requeue failed tasks
		msg.Ack(false)
		log.Printf("Task failed and removed from queue (no retry) for provider: %s", taskMsg.Provider)
		return
	}

	// Acknowledge successful processing
	msg.Ack(false)
	log.Printf("Task completed successfully for provider: %s", taskMsg.Provider)
}

// executeTask processes the task using the appropriate AI provider
func (w *TaskWorker) executeTask(ctx context.Context, taskMsg *TaskMessage) error {
	log.Printf("Executing task with provider: %s", taskMsg.Provider)

	// Check if enough time has passed to reset failure count (10 minutes)
	if lastFailure, exists := w.lastFailureTime[taskMsg.Provider]; exists {
		if time.Since(lastFailure) > 10*time.Minute {
			log.Printf("Resetting failure count for provider %s due to time elapsed", taskMsg.Provider)
			w.failureCount[taskMsg.Provider] = 0
			delete(w.lastFailureTime, taskMsg.Provider)
		}
	}

	// Check failure count for this provider
	if w.failureCount[taskMsg.Provider] >= w.maxRetries {
		log.Printf("Provider %s has failed %d times consecutively, skipping task",
			taskMsg.Provider, w.failureCount[taskMsg.Provider])
		// Record this as a skipped task
		w.recordTaskFailure(taskMsg, fmt.Sprintf("provider %s skipped due to too many consecutive failures (%d)", taskMsg.Provider, w.failureCount[taskMsg.Provider]))
		return fmt.Errorf("provider %s skipped due to too many consecutive failures (%d)",
			taskMsg.Provider, w.failureCount[taskMsg.Provider])
	}

	// Get the appropriate provider
	provider, exists := w.providerFactory.GetProvider(taskMsg.Provider)
	if !exists {
		w.failureCount[taskMsg.Provider]++
		return fmt.Errorf("unknown provider: %s", taskMsg.Provider)
	}

	// Check if provider is configured
	if !provider.IsConfigured() {
		// Configuration errors should increment failure count more aggressively
		w.failureCount[taskMsg.Provider] += 2 // Double increment for config issues
		w.lastFailureTime[taskMsg.Provider] = time.Now()
		log.Printf("Provider %s configuration issue (failure count: %d)", taskMsg.Provider, w.failureCount[taskMsg.Provider])
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
		// Different error types get different failure count increments
		if strings.Contains(err.Error(), "not properly configured") ||
		   strings.Contains(err.Error(), "API key") ||
		   strings.Contains(err.Error(), "authentication") {
			// Configuration/auth errors: increment by 2
			w.failureCount[taskMsg.Provider] += 2
		} else if strings.Contains(err.Error(), "timeout") ||
		         strings.Contains(err.Error(), "network") ||
		         strings.Contains(err.Error(), "connection") {
			// Network/timeout errors: increment by 1 (might be temporary)
			w.failureCount[taskMsg.Provider]++
		} else {
			// Other errors: increment by 1
			w.failureCount[taskMsg.Provider]++
		}
		w.lastFailureTime[taskMsg.Provider] = time.Now()
		log.Printf("Provider %s execution error (failure count: %d): %v", taskMsg.Provider, w.failureCount[taskMsg.Provider], err)
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
		// Task failure: increment by 1 (might be task-specific)
		w.failureCount[taskMsg.Provider]++
		w.lastFailureTime[taskMsg.Provider] = time.Now()
		log.Printf("Provider %s task failed (failure count: %d): %s", taskMsg.Provider, w.failureCount[taskMsg.Provider], response.Error)
		return fmt.Errorf("task execution failed: %s", response.Error)
	}

	// Reset failure count on successful execution
	w.failureCount[taskMsg.Provider] = 0
	delete(w.lastFailureTime, taskMsg.Provider)
	log.Printf("Provider %s task succeeded, failure count reset", taskMsg.Provider)
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

// ResetFailureCount resets the failure count for a specific provider
func (w *TaskWorker) ResetFailureCount(provider string) {
	w.failureCount[provider] = 0
	log.Printf("Reset failure count for provider: %s", provider)
}

// ResetAllFailureCounts resets failure counts for all providers
func (w *TaskWorker) ResetAllFailureCounts() {
	for provider := range w.failureCount {
		w.failureCount[provider] = 0
	}
	log.Println("Reset all provider failure counts")
}

// GetFailureCount returns the current failure count for a provider
func (w *TaskWorker) GetFailureCount(provider string) int {
	return w.failureCount[provider]
}

// GetFailureCountsReport returns all failure counts for monitoring
func (w *TaskWorker) GetFailureCountsReport() map[string]int {
	report := make(map[string]int)
	for provider, count := range w.failureCount {
		report[provider] = count
	}
	return report
}

// recordTaskFailure records task failure to database
func (w *TaskWorker) recordTaskFailure(taskMsg *TaskMessage, errorMsg string) {
	if w.db == nil {
		log.Printf("Database connection not available, skipping failure record")
		return
	}

	// Additional safety check for GORM DB instance
	if w.db.Error != nil {
		log.Printf("Database connection has error, skipping failure record: %v", w.db.Error)
		return
	}

	// Generate unique request ID for this failed task
	requestID := mysql.PKIDGenerate()
	now := time.Now()
	processingTime := int64(0) // Minimal processing time for failed tasks

	// Create workflow history record
	history := mysql.WorkflowHistories{
		RequestID:        requestID,
		Status:          "failed",
		Tasks:           taskMsg.Tasks,
		RepositoryName:  taskMsg.RepositoryName,
		Provider:        taskMsg.Provider,
		Interactive:     taskMsg.Interactive,
		ContinueTask:    taskMsg.ContinueTask,
		CreatedAt:       now,
		CompletedAt:     &now,
		ProcessingTimeMs: &processingTime,
	}

	// Set optional fields if present
	if taskMsg.WorkingDir != "" {
		history.WorkingDir = &taskMsg.WorkingDir
	}
	if taskMsg.Cmd != "" {
		history.Cmd = &taskMsg.Cmd
	}
	if errorMsg != "" {
		history.Error = &errorMsg
	}

	// Insert to database
	if err := w.db.Create(&history).Error; err != nil {
		log.Printf("Failed to record task failure to database: %v", err)
	} else {
		log.Printf("Recorded task failure to database with request_id: %s", requestID)
	}
}