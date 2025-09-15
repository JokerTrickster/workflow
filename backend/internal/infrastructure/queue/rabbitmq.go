package queue

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"ai-git-workbench/internal/domain/entities"
	"ai-git-workbench/internal/domain/repositories"
	"ai-git-workbench/internal/domain/valueobjects"
	"ai-git-workbench/internal/infrastructure/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQRepository implements the QueueRepository interface using RabbitMQ
type RabbitMQRepository struct {
	config     *config.RabbitMQConfig
	connection *amqp.Connection
	channel    *amqp.Channel
	queue      amqp.Queue
	
	// Connection management
	mutex          sync.RWMutex
	connected      bool
	reconnectDelay time.Duration
	maxRetries     int
	
	// Statistics tracking
	statistics     *repositories.QueueStatistics
	statsMutex     sync.RWMutex
	
	// Shutdown management
	closeChan      chan struct{}
	reconnectChan  chan struct{}
}

// NewRabbitMQRepository creates a new RabbitMQ repository
func NewRabbitMQRepository(cfg *config.RabbitMQConfig) (repositories.QueueRepository, error) {
	if cfg == nil {
		return nil, fmt.Errorf("RabbitMQ configuration is required")
	}

	reconnectDelay, err := time.ParseDuration(cfg.ReconnectDelay)
	if err != nil {
		log.Printf("⚠️ Invalid reconnect delay '%s', using default 5s", cfg.ReconnectDelay)
		reconnectDelay = 5 * time.Second
	}

	repo := &RabbitMQRepository{
		config:         cfg,
		reconnectDelay: reconnectDelay,
		maxRetries:     cfg.MaxRetries,
		statistics: &repositories.QueueStatistics{
			TotalEnqueued:         0,
			TotalDequeued:         0,
			TotalProcessed:        0,
			TotalFailed:           0,
			CurrentQueueLength:    0,
			AverageProcessingTime: time.Minute,
			ThroughputPerMinute:   0,
			ErrorRate:             0,
			DeadLetterCount:       0,
			WorkersActive:         0,
			LastActivityAt:        time.Now(),
		},
		closeChan:     make(chan struct{}),
		reconnectChan: make(chan struct{}, 1),
	}

	// Initial connection
	if err := repo.connect(); err != nil {
		log.Printf("⚠️ Initial RabbitMQ connection failed: %v. Will retry in background.", err)
		// Start reconnection routine
		go repo.reconnectLoop()
	} else {
		log.Printf("✅ RabbitMQ repository initialized successfully")
	}

	return repo, nil
}

// connect establishes connection to RabbitMQ
func (r *RabbitMQRepository) connect() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Build connection URL
	url := fmt.Sprintf("amqp://%s:%s@%s:%s%s",
		r.config.Username,
		r.config.Password,
		r.config.Host,
		r.config.Port,
		r.config.VHost,
	)

	// Establish connection
	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare queue
	queue, err := ch.QueueDeclare(
		r.config.Queue,      // name
		r.config.Durable,    // durable
		r.config.AutoDelete, // delete when unused
		r.config.Exclusive,  // exclusive
		r.config.NoWait,     // no-wait
		nil,                 // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Clean up old connection if exists
	if r.connection != nil && !r.connection.IsClosed() {
		r.connection.Close()
	}

	r.connection = conn
	r.channel = ch
	r.queue = queue
	r.connected = true

	log.Printf("🔗 Connected to RabbitMQ - Queue: %s, Messages: %d", queue.Name, queue.Messages)

	// Set up connection close notification
	go r.handleConnectionClose()

	return nil
}

// handleConnectionClose monitors connection and triggers reconnection
func (r *RabbitMQRepository) handleConnectionClose() {
	closeChan := r.connection.NotifyClose(make(chan *amqp.Error))
	
	select {
	case err := <-closeChan:
		if err != nil {
			log.Printf("🔌 RabbitMQ connection lost: %v", err)
			r.mutex.Lock()
			r.connected = false
			r.mutex.Unlock()
			
			// Trigger reconnection
			select {
			case r.reconnectChan <- struct{}{}:
			default:
			}
		}
	case <-r.closeChan:
		return
	}
}

// reconnectLoop handles automatic reconnection
func (r *RabbitMQRepository) reconnectLoop() {
	for {
		select {
		case <-r.reconnectChan:
			r.attemptReconnect()
		case <-r.closeChan:
			return
		}
	}
}

// attemptReconnect tries to reconnect with exponential backoff
func (r *RabbitMQRepository) attemptReconnect() {
	retryCount := 0
	baseDelay := r.reconnectDelay

	for retryCount < r.maxRetries {
		select {
		case <-r.closeChan:
			return
		default:
		}

		log.Printf("🔄 Attempting to reconnect to RabbitMQ (attempt %d/%d)", retryCount+1, r.maxRetries)
		
		if err := r.connect(); err != nil {
			retryCount++
			delay := baseDelay * time.Duration(1<<retryCount) // Exponential backoff
			log.Printf("❌ Reconnection failed: %v. Retrying in %v", err, delay)
			time.Sleep(delay)
		} else {
			log.Printf("✅ Successfully reconnected to RabbitMQ")
			return
		}
	}

	log.Printf("💀 Max reconnection attempts (%d) reached. Operating in degraded mode.", r.maxRetries)
}

// isConnected checks if the connection is healthy
func (r *RabbitMQRepository) isConnected() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.connected && r.connection != nil && !r.connection.IsClosed()
}

// Enqueue adds a task to the queue
func (r *RabbitMQRepository) Enqueue(ctx context.Context, task *entities.Task) error {
	if !r.isConnected() {
		log.Printf("⚠️ RabbitMQ not connected, message will be lost: %s", task.ID().Value())
		return fmt.Errorf("RabbitMQ connection not available")
	}

	message := NewTaskCreatedMessage(task)
	body, err := message.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if r.channel == nil {
		return fmt.Errorf("RabbitMQ channel not available")
	}

	err = r.channel.PublishWithContext(
		ctx,
		r.config.Exchange,   // exchange
		r.config.RoutingKey, // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // Make message persistent
			Timestamp:    time.Now(),
			MessageId:    task.ID().Value(),
		},
	)

	if err != nil {
		r.updateFailureStats()
		return fmt.Errorf("failed to publish message: %w", err)
	}

	r.updateEnqueueStats()
	log.Printf("📨 Task enqueued to RabbitMQ: %s", task.ID().Value())
	return nil
}

// Note: For the comprehensive QueueRepository interface, many methods are not applicable 
// to a publish-only RabbitMQ implementation. These will return appropriate responses.

// Dequeue - Not applicable for publish-only queue
func (r *RabbitMQRepository) Dequeue(ctx context.Context) (*entities.Task, error) {
	return nil, fmt.Errorf("dequeue operation not supported in publish-only mode")
}

// DequeueWithTimeout - Not applicable for publish-only queue
func (r *RabbitMQRepository) DequeueWithTimeout(ctx context.Context, timeout time.Duration) (*entities.Task, error) {
	return nil, fmt.Errorf("dequeue operation not supported in publish-only mode")
}

// Peek - Not applicable for publish-only queue
func (r *RabbitMQRepository) Peek(ctx context.Context) (*entities.Task, error) {
	return nil, fmt.Errorf("peek operation not supported in publish-only mode")
}

// PeekMultiple - Not applicable for publish-only queue
func (r *RabbitMQRepository) PeekMultiple(ctx context.Context, count int) ([]*entities.Task, error) {
	return nil, fmt.Errorf("peek operation not supported in publish-only mode")
}

// GetQueueLength returns the current queue length
func (r *RabbitMQRepository) GetQueueLength(ctx context.Context) (int, error) {
	if !r.isConnected() {
		return 0, fmt.Errorf("RabbitMQ connection not available")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if r.channel == nil {
		return 0, fmt.Errorf("RabbitMQ channel not available")
	}

	// Passive declare to get queue info without creating it
	queue, err := r.channel.QueueDeclarePassive(
		r.config.Queue, // name
		r.config.Durable,
		r.config.AutoDelete,
		r.config.Exclusive,
		r.config.NoWait,
		nil,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to get queue info: %w", err)
	}

	return queue.Messages, nil
}

// IsEmpty checks if the queue is empty
func (r *RabbitMQRepository) IsEmpty(ctx context.Context) (bool, error) {
	length, err := r.GetQueueLength(ctx)
	if err != nil {
		return true, err
	}
	return length == 0, nil
}

// EnqueueWithPriority - RabbitMQ basic queues don't support priority natively
func (r *RabbitMQRepository) EnqueueWithPriority(ctx context.Context, task *entities.Task, priority int) error {
	// For basic implementation, ignore priority and use regular enqueue
	log.Printf("⚠️ Priority queuing not implemented, using regular enqueue for task: %s", task.ID().Value())
	return r.Enqueue(ctx, task)
}

// EnqueueBatch enqueues multiple tasks
func (r *RabbitMQRepository) EnqueueBatch(ctx context.Context, tasks []*entities.Task) error {
	successCount := 0
	var lastError error

	for _, task := range tasks {
		if err := r.Enqueue(ctx, task); err != nil {
			lastError = err
			log.Printf("❌ Failed to enqueue task %s: %v", task.ID().Value(), err)
		} else {
			successCount++
		}
	}

	if successCount == 0 && lastError != nil {
		return fmt.Errorf("failed to enqueue any tasks: %w", lastError)
	}

	if successCount < len(tasks) {
		log.Printf("⚠️ Enqueued %d/%d tasks successfully", successCount, len(tasks))
	}

	return nil
}

// DequeueBatch - Not applicable for publish-only queue
func (r *RabbitMQRepository) DequeueBatch(ctx context.Context, count int) ([]*entities.Task, error) {
	return nil, fmt.Errorf("dequeue operation not supported in publish-only mode")
}

// Clear - Dangerous operation, not implemented for production safety
func (r *RabbitMQRepository) Clear(ctx context.Context) error {
	return fmt.Errorf("queue clear operation not supported for safety reasons")
}

// RemoveTask - Best effort message removal (not guaranteed in RabbitMQ)
func (r *RabbitMQRepository) RemoveTask(ctx context.Context, taskID valueobjects.TaskID) error {
	log.Printf("⚠️ Task removal not supported in RabbitMQ publish-only mode: %s", taskID.Value())
	return nil // Return nil to not break the application
}

// Dead letter queue operations - Simplified for basic implementation
func (r *RabbitMQRepository) GetFailedTasks(ctx context.Context, limit, offset int) ([]*entities.Task, error) {
	return []*entities.Task{}, nil
}

func (r *RabbitMQRepository) MoveToDeadLetter(ctx context.Context, task *entities.Task, reason string) error {
	log.Printf("💀 Task marked as failed (dead letter): %s - %s", task.ID().Value(), reason)
	r.updateFailureStats()
	return nil
}

func (r *RabbitMQRepository) RetryFromDeadLetter(ctx context.Context, taskID valueobjects.TaskID) error {
	return fmt.Errorf("retry from dead letter not implemented")
}

// GetQueueStatistics returns current queue statistics
func (r *RabbitMQRepository) GetQueueStatistics(ctx context.Context) (*repositories.QueueStatistics, error) {
	r.statsMutex.RLock()
	defer r.statsMutex.RUnlock()

	// Get current queue length
	currentLength := 0
	if length, err := r.GetQueueLength(ctx); err == nil {
		currentLength = length
	}

	// Create a copy of statistics
	stats := *r.statistics
	stats.CurrentQueueLength = currentLength
	stats.LastActivityAt = time.Now()

	return &stats, nil
}

// GetTasksInProgress - Not applicable for publish-only queue
func (r *RabbitMQRepository) GetTasksInProgress(ctx context.Context) ([]*entities.Task, error) {
	return []*entities.Task{}, nil
}

// GetQueueHealth returns queue health status
func (r *RabbitMQRepository) GetQueueHealth(ctx context.Context) (*repositories.QueueHealth, error) {
	isHealthy := r.isConnected()
	issues := []string{}

	if !isHealthy {
		issues = append(issues, "rabbitmq_connection_failed")
	}

	queueLength := 0
	if length, err := r.GetQueueLength(ctx); err == nil {
		queueLength = length
	} else if isHealthy {
		issues = append(issues, "queue_info_unavailable")
	}

	// Consider queue unhealthy if too many messages are backed up
	maxQueueLength := 10000
	if queueLength > maxQueueLength {
		isHealthy = false
		issues = append(issues, "queue_length_high")
	}

	return &repositories.QueueHealth{
		IsHealthy:        isHealthy,
		QueueLength:      queueLength,
		MaxQueueLength:   maxQueueLength,
		ProcessingRate:   0, // Not available in publish-only mode
		ErrorRate:        r.calculateErrorRate(),
		OldestTaskAge:    time.Duration(0), // Not available in publish-only mode
		WorkersConnected: 0,                // Not available in publish-only mode
		WorkersRequired:  1,                // Assume 1 worker needed
		Issues:           issues,
		LastHealthCheck:  time.Now(),
	}, nil
}

// Worker assignment operations - Not applicable for publish-only queue
func (r *RabbitMQRepository) AssignTaskToWorker(ctx context.Context, taskID valueobjects.TaskID, workerID string) error {
	return fmt.Errorf("worker assignment not supported in publish-only mode")
}

func (r *RabbitMQRepository) ReleaseTaskFromWorker(ctx context.Context, taskID valueobjects.TaskID) error {
	return fmt.Errorf("worker operations not supported in publish-only mode")
}

func (r *RabbitMQRepository) GetWorkerTasks(ctx context.Context, workerID string) ([]*entities.Task, error) {
	return []*entities.Task{}, nil
}

// Lease operations - Not applicable for publish-only queue
func (r *RabbitMQRepository) LeaseTask(ctx context.Context, taskID valueobjects.TaskID, workerID string, leaseDuration time.Duration) error {
	return fmt.Errorf("lease operations not supported in publish-only mode")
}

func (r *RabbitMQRepository) RenewLease(ctx context.Context, taskID valueobjects.TaskID, workerID string, leaseDuration time.Duration) error {
	return fmt.Errorf("lease operations not supported in publish-only mode")
}

func (r *RabbitMQRepository) ReleaseLease(ctx context.Context, taskID valueobjects.TaskID, workerID string) error {
	return fmt.Errorf("lease operations not supported in publish-only mode")
}

// Close closes the RabbitMQ connection
func (r *RabbitMQRepository) Close() error {
	log.Printf("🧹 Closing RabbitMQ repository...")
	
	// Signal shutdown
	close(r.closeChan)
	
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.connected = false
	
	if r.channel != nil {
		r.channel.Close()
		r.channel = nil
	}
	
	if r.connection != nil && !r.connection.IsClosed() {
		r.connection.Close()
		r.connection = nil
	}
	
	log.Printf("✅ RabbitMQ repository closed successfully")
	return nil
}

// Helper methods for statistics tracking

func (r *RabbitMQRepository) updateEnqueueStats() {
	r.statsMutex.Lock()
	defer r.statsMutex.Unlock()
	
	r.statistics.TotalEnqueued++
	r.statistics.LastActivityAt = time.Now()
	
	// Calculate throughput (simplified)
	if r.statistics.TotalEnqueued > 0 {
		duration := time.Since(r.statistics.LastActivityAt)
		if duration > 0 {
			r.statistics.ThroughputPerMinute = float64(r.statistics.TotalEnqueued) / duration.Minutes()
		}
	}
}

func (r *RabbitMQRepository) updateFailureStats() {
	r.statsMutex.Lock()
	defer r.statsMutex.Unlock()
	
	r.statistics.TotalFailed++
	r.statistics.LastActivityAt = time.Now()
}

func (r *RabbitMQRepository) calculateErrorRate() float64 {
	r.statsMutex.RLock()
	defer r.statsMutex.RUnlock()
	
	total := r.statistics.TotalEnqueued
	if total == 0 {
		return 0
	}
	
	return float64(r.statistics.TotalFailed) / float64(total)
}