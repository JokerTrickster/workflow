package services

import (
	"context"
	"log"
	"sync"
	"time"

	"ai-git-workbench/internal/domain/entities"
	"ai-git-workbench/internal/domain/repositories"
	"ai-git-workbench/internal/domain/valueobjects"
)

// MockQueueRepository implements QueueRepository interface for testing and development
type MockQueueRepository struct {
	queue         []*entities.Task
	deadLetter    []*repositories.DeadLetterTask
	workers       map[string][]*entities.Task
	statistics    *repositories.QueueStatistics
	mutex         sync.RWMutex
	enqueuedCount int64
	dequeuedCount int64
	processedCount int64
	failedCount   int64
}

// NewMockQueueRepository creates a new mock queue repository
func NewMockQueueRepository() repositories.QueueRepository {
	return &MockQueueRepository{
		queue:      make([]*entities.Task, 0),
		deadLetter: make([]*repositories.DeadLetterTask, 0),
		workers:    make(map[string][]*entities.Task),
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
	}
}

// Enqueue adds a task to the queue
func (r *MockQueueRepository) Enqueue(ctx context.Context, task *entities.Task) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.queue = append(r.queue, task)
	r.enqueuedCount++
	r.updateStatistics()

	log.Printf("📨 Task enqueued: %s (queue length: %d)", task.ID().Value(), len(r.queue))
	return nil
}

// Dequeue removes and returns a task from the queue
func (r *MockQueueRepository) Dequeue(ctx context.Context) (*entities.Task, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(r.queue) == 0 {
		return nil, nil // No tasks available
	}

	task := r.queue[0]
	r.queue = r.queue[1:]
	r.dequeuedCount++
	r.updateStatistics()

	log.Printf("📤 Task dequeued: %s (queue length: %d)", task.ID().Value(), len(r.queue))
	return task, nil
}

// DequeueWithTimeout dequeues with a timeout
func (r *MockQueueRepository) DequeueWithTimeout(ctx context.Context, timeout time.Duration) (*entities.Task, error) {
	// For mock implementation, just use regular dequeue
	return r.Dequeue(ctx)
}

// Peek returns the next task without removing it
func (r *MockQueueRepository) Peek(ctx context.Context) (*entities.Task, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if len(r.queue) == 0 {
		return nil, nil
	}

	return r.queue[0], nil
}

// PeekMultiple returns multiple tasks without removing them
func (r *MockQueueRepository) PeekMultiple(ctx context.Context, count int) ([]*entities.Task, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if count > len(r.queue) {
		count = len(r.queue)
	}

	tasks := make([]*entities.Task, count)
	copy(tasks, r.queue[:count])
	return tasks, nil
}

// GetQueueLength returns the current queue length
func (r *MockQueueRepository) GetQueueLength(ctx context.Context) (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.queue), nil
}

// IsEmpty checks if the queue is empty
func (r *MockQueueRepository) IsEmpty(ctx context.Context) (bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.queue) == 0, nil
}

// EnqueueWithPriority enqueues with priority (mock just uses regular enqueue)
func (r *MockQueueRepository) EnqueueWithPriority(ctx context.Context, task *entities.Task, priority int) error {
	// For mock, ignore priority and use regular enqueue
	return r.Enqueue(ctx, task)
}

// EnqueueBatch enqueues multiple tasks
func (r *MockQueueRepository) EnqueueBatch(ctx context.Context, tasks []*entities.Task) error {
	for _, task := range tasks {
		if err := r.Enqueue(ctx, task); err != nil {
			return err
		}
	}
	return nil
}

// DequeueBatch dequeues multiple tasks
func (r *MockQueueRepository) DequeueBatch(ctx context.Context, count int) ([]*entities.Task, error) {
	tasks := make([]*entities.Task, 0, count)
	for i := 0; i < count; i++ {
		task, err := r.Dequeue(ctx)
		if err != nil {
			return tasks, err
		}
		if task == nil {
			break // No more tasks
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// Clear removes all tasks from the queue
func (r *MockQueueRepository) Clear(ctx context.Context) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.queue = make([]*entities.Task, 0)
	r.updateStatistics()
	log.Printf("🧹 Queue cleared")
	return nil
}

// RemoveTask removes a specific task from the queue
func (r *MockQueueRepository) RemoveTask(ctx context.Context, taskID valueobjects.TaskID) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i, task := range r.queue {
		if task.ID().Equals(taskID) {
			r.queue = append(r.queue[:i], r.queue[i+1:]...)
			r.updateStatistics()
			log.Printf("🗑️ Task removed from queue: %s", taskID.Value())
			return nil
		}
	}

	log.Printf("⚠️ Task not found in queue for removal: %s", taskID.Value())
	return nil // Not an error if task is not in queue
}

// GetFailedTasks returns tasks from dead letter queue
func (r *MockQueueRepository) GetFailedTasks(ctx context.Context, limit, offset int) ([]*entities.Task, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	start := offset
	if start >= len(r.deadLetter) {
		return []*entities.Task{}, nil
	}

	end := start + limit
	if end > len(r.deadLetter) {
		end = len(r.deadLetter)
	}

	tasks := make([]*entities.Task, end-start)
	for i := start; i < end; i++ {
		tasks[i-start] = r.deadLetter[i].Task
	}

	return tasks, nil
}

// MoveToDeadLetter moves a task to the dead letter queue
func (r *MockQueueRepository) MoveToDeadLetter(ctx context.Context, task *entities.Task, reason string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	deadLetterTask := &repositories.DeadLetterTask{
		Task:       task,
		Reason:     reason,
		FailedAt:   time.Now(),
		RetryCount: 0,
		LastError:  reason,
	}

	r.deadLetter = append(r.deadLetter, deadLetterTask)
	r.failedCount++
	r.updateStatistics()

	log.Printf("💀 Task moved to dead letter queue: %s (reason: %s)", task.ID().Value(), reason)
	return nil
}

// RetryFromDeadLetter moves a task from dead letter back to main queue
func (r *MockQueueRepository) RetryFromDeadLetter(ctx context.Context, taskID valueobjects.TaskID) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i, deadTask := range r.deadLetter {
		if deadTask.Task.ID().Equals(taskID) {
			// Move back to main queue
			r.queue = append(r.queue, deadTask.Task)
			// Remove from dead letter queue
			r.deadLetter = append(r.deadLetter[:i], r.deadLetter[i+1:]...)
			r.updateStatistics()
			log.Printf("🔄 Task retried from dead letter queue: %s", taskID.Value())
			return nil
		}
	}

	log.Printf("⚠️ Task not found in dead letter queue for retry: %s", taskID.Value())
	return nil
}

// GetQueueStatistics returns queue statistics
func (r *MockQueueRepository) GetQueueStatistics(ctx context.Context) (*repositories.QueueStatistics, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Update current statistics
	stats := *r.statistics // Copy
	stats.CurrentQueueLength = len(r.queue)
	stats.DeadLetterCount = len(r.deadLetter)
	stats.WorkersActive = len(r.workers)
	stats.LastActivityAt = time.Now()

	return &stats, nil
}

// GetTasksInProgress returns tasks currently being processed
func (r *MockQueueRepository) GetTasksInProgress(ctx context.Context) ([]*entities.Task, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var inProgress []*entities.Task
	for _, workerTasks := range r.workers {
		inProgress = append(inProgress, workerTasks...)
	}

	return inProgress, nil
}

// GetQueueHealth returns queue health status
func (r *MockQueueRepository) GetQueueHealth(ctx context.Context) (*repositories.QueueHealth, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	queueLength := len(r.queue)
	isHealthy := queueLength < 1000 // Arbitrary threshold for mock

	health := &repositories.QueueHealth{
		IsHealthy:           isHealthy,
		QueueLength:         queueLength,
		MaxQueueLength:      1000,
		ProcessingRate:      10.0, // Mock value
		ErrorRate:           0.05, // Mock value
		OldestTaskAge:       time.Minute * 5, // Mock value
		WorkersConnected:    len(r.workers),
		WorkersRequired:     3, // Mock value
		Issues:              []string{},
		LastHealthCheck:     time.Now(),
	}

	if !isHealthy {
		health.Issues = append(health.Issues, "queue_length_high")
	}

	return health, nil
}

// Worker assignment methods

// AssignTaskToWorker assigns a task to a worker
func (r *MockQueueRepository) AssignTaskToWorker(ctx context.Context, taskID valueobjects.TaskID, workerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Find and remove task from main queue
	for i, task := range r.queue {
		if task.ID().Equals(taskID) {
			r.queue = append(r.queue[:i], r.queue[i+1:]...)
			
			// Assign to worker
			if r.workers[workerID] == nil {
				r.workers[workerID] = make([]*entities.Task, 0)
			}
			r.workers[workerID] = append(r.workers[workerID], task)
			
			log.Printf("👷 Task assigned to worker %s: %s", workerID, taskID.Value())
			return nil
		}
	}

	log.Printf("⚠️ Task not found for worker assignment: %s", taskID.Value())
	return nil
}

// ReleaseTaskFromWorker releases a task from a worker
func (r *MockQueueRepository) ReleaseTaskFromWorker(ctx context.Context, taskID valueobjects.TaskID) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for workerID, tasks := range r.workers {
		for i, task := range tasks {
			if task.ID().Equals(taskID) {
				// Remove from worker
				r.workers[workerID] = append(tasks[:i], tasks[i+1:]...)
				if len(r.workers[workerID]) == 0 {
					delete(r.workers, workerID)
				}
				r.processedCount++
				r.updateStatistics()
				log.Printf("🔓 Task released from worker %s: %s", workerID, taskID.Value())
				return nil
			}
		}
	}

	log.Printf("⚠️ Task not found for worker release: %s", taskID.Value())
	return nil
}

// GetWorkerTasks returns tasks assigned to a worker
func (r *MockQueueRepository) GetWorkerTasks(ctx context.Context, workerID string) ([]*entities.Task, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tasks := r.workers[workerID]
	if tasks == nil {
		return []*entities.Task{}, nil
	}

	// Return a copy
	result := make([]*entities.Task, len(tasks))
	copy(result, tasks)
	return result, nil
}

// Lease-based processing methods (simplified for mock)

// LeaseTask creates a lease on a task
func (r *MockQueueRepository) LeaseTask(ctx context.Context, taskID valueobjects.TaskID, workerID string, leaseDuration time.Duration) error {
	// For mock, just assign to worker
	return r.AssignTaskToWorker(ctx, taskID, workerID)
}

// RenewLease renews a task lease
func (r *MockQueueRepository) RenewLease(ctx context.Context, taskID valueobjects.TaskID, workerID string, leaseDuration time.Duration) error {
	// For mock, this is a no-op
	log.Printf("🔄 Lease renewed for task %s by worker %s", taskID.Value(), workerID)
	return nil
}

// ReleaseLease releases a task lease
func (r *MockQueueRepository) ReleaseLease(ctx context.Context, taskID valueobjects.TaskID, workerID string) error {
	return r.ReleaseTaskFromWorker(ctx, taskID)
}

// Helper method to update statistics
func (r *MockQueueRepository) updateStatistics() {
	r.statistics.TotalEnqueued = r.enqueuedCount
	r.statistics.TotalDequeued = r.dequeuedCount
	r.statistics.TotalProcessed = r.processedCount
	r.statistics.TotalFailed = r.failedCount
	r.statistics.CurrentQueueLength = len(r.queue)
	r.statistics.DeadLetterCount = len(r.deadLetter)
	r.statistics.WorkersActive = len(r.workers)
	r.statistics.LastActivityAt = time.Now()

	// Calculate derived metrics
	if r.statistics.TotalProcessed > 0 {
		r.statistics.ErrorRate = float64(r.statistics.TotalFailed) / float64(r.statistics.TotalProcessed)
	}
}