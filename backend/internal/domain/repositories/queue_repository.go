package repositories

import (
	"context"
	"time"

	"ai-git-workbench/internal/domain/entities"
	"ai-git-workbench/internal/domain/valueobjects"
)

// QueueRepository defines the interface for task queue operations
type QueueRepository interface {
	// Queue operations
	Enqueue(ctx context.Context, task *entities.Task) error
	Dequeue(ctx context.Context) (*entities.Task, error)
	DequeueWithTimeout(ctx context.Context, timeout time.Duration) (*entities.Task, error)
	
	// Peek operations (without removing from queue)
	Peek(ctx context.Context) (*entities.Task, error)
	PeekMultiple(ctx context.Context, count int) ([]*entities.Task, error)
	
	// Queue status
	GetQueueLength(ctx context.Context) (int, error)
	IsEmpty(ctx context.Context) (bool, error)
	
	// Priority queue operations
	EnqueueWithPriority(ctx context.Context, task *entities.Task, priority int) error
	
	// Batch operations
	EnqueueBatch(ctx context.Context, tasks []*entities.Task) error
	DequeueBatch(ctx context.Context, count int) ([]*entities.Task, error)
	
	// Queue management
	Clear(ctx context.Context) error
	RemoveTask(ctx context.Context, taskID valueobjects.TaskID) error
	
	// Dead letter queue operations
	GetFailedTasks(ctx context.Context, limit, offset int) ([]*entities.Task, error)
	MoveToDeadLetter(ctx context.Context, task *entities.Task, reason string) error
	RetryFromDeadLetter(ctx context.Context, taskID valueobjects.TaskID) error
	
	// Monitoring and metrics
	GetQueueStatistics(ctx context.Context) (*QueueStatistics, error)
	GetTasksInProgress(ctx context.Context) ([]*entities.Task, error)
	GetQueueHealth(ctx context.Context) (*QueueHealth, error)
	
	// Worker assignment
	AssignTaskToWorker(ctx context.Context, taskID valueobjects.TaskID, workerID string) error
	ReleaseTaskFromWorker(ctx context.Context, taskID valueobjects.TaskID) error
	GetWorkerTasks(ctx context.Context, workerID string) ([]*entities.Task, error)
	
	// Lease-based processing (for preventing duplicate processing)
	LeaseTask(ctx context.Context, taskID valueobjects.TaskID, workerID string, leaseDuration time.Duration) error
	RenewLease(ctx context.Context, taskID valueobjects.TaskID, workerID string, leaseDuration time.Duration) error
	ReleaseLease(ctx context.Context, taskID valueobjects.TaskID, workerID string) error
}

// QueueStatistics represents queue performance metrics
type QueueStatistics struct {
	TotalEnqueued       int64
	TotalDequeued       int64
	TotalProcessed      int64
	TotalFailed         int64
	CurrentQueueLength  int
	AverageProcessingTime time.Duration
	ThroughputPerMinute float64
	ErrorRate           float64
	DeadLetterCount     int
	WorkersActive       int
	LastActivityAt      time.Time
}

// QueueHealth represents the health status of the queue
type QueueHealth struct {
	IsHealthy           bool
	QueueLength         int
	MaxQueueLength      int
	ProcessingRate      float64 // tasks per minute
	ErrorRate           float64 // percentage
	OldestTaskAge       time.Duration
	WorkersConnected    int
	WorkersRequired     int
	Issues              []string
	LastHealthCheck     time.Time
}

// TaskLease represents a lease on a task for processing
type TaskLease struct {
	TaskID      valueobjects.TaskID
	WorkerID    string
	LeasedAt    time.Time
	ExpiresAt   time.Time
	RenewCount  int
}

// DeadLetterTask represents a task that failed processing
type DeadLetterTask struct {
	Task       *entities.Task
	Reason     string
	FailedAt   time.Time
	RetryCount int
	LastError  string
}

// QueueMessage represents a message in the queue with metadata
type QueueMessage struct {
	Task       *entities.Task
	EnqueuedAt time.Time
	Priority   int
	Attempts   int
	LastError  string
	WorkerID   string
	LeasedUntil *time.Time
}