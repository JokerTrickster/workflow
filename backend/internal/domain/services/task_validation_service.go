package services

import (
	"context"
	"time"

	"ai-git-workbench/internal/domain/entities"
	"ai-git-workbench/internal/domain/errors"
	"ai-git-workbench/internal/domain/repositories"
	"ai-git-workbench/internal/domain/valueobjects"
)

// TaskValidationService provides business rule validation for tasks
type TaskValidationService struct {
	taskRepo repositories.TaskRepository
}

// NewTaskValidationService creates a new TaskValidationService
func NewTaskValidationService(taskRepo repositories.TaskRepository) *TaskValidationService {
	return &TaskValidationService{
		taskRepo: taskRepo,
	}
}

// ValidateTaskCreation validates business rules for creating a new task
func (s *TaskValidationService) ValidateTaskCreation(ctx context.Context, task *entities.Task) error {
	// Check for duplicate task IDs
	existingTask, err := s.taskRepo.GetByID(ctx, task.ID())
	if err == nil && existingTask != nil {
		return errors.NewTaskAlreadyExistsError(task.ID().Value())
	}
	if err != nil && !errors.IsTaskNotFound(err) {
		return errors.NewTaskValidationFailedError("failed to check for existing task", err)
	}

	// Validate user task limits (business rule: max 10 active tasks per user)
	activeTasks, err := s.taskRepo.GetActiveTasks(ctx, task.UserID())
	if err != nil {
		return errors.NewTaskValidationFailedError("failed to check user active tasks", err)
	}

	if len(activeTasks) >= 10 {
		return errors.NewTaskValidationFailedError("user has reached maximum active task limit (10)", nil)
	}

	return nil
}

// ValidateTaskUpdate validates business rules for updating a task
func (s *TaskValidationService) ValidateTaskUpdate(ctx context.Context, task *entities.Task) error {
	// Verify task exists
	existingTask, err := s.taskRepo.GetByID(ctx, task.ID())
	if err != nil {
		if errors.IsTaskNotFound(err) {
			return errors.NewTaskNotFoundError(task.ID().Value())
		}
		return errors.NewTaskValidationFailedError("failed to retrieve existing task", err)
	}

	// Check if task can be modified
	if !existingTask.CanBeModified() {
		return errors.NewTaskAlreadyCompletedError(task.ID().Value())
	}

	// Validate version for optimistic locking
	if task.Version() != existingTask.Version() {
		return errors.NewConcurrentModificationError(task.ID().Value())
	}

	return nil
}

// ValidateStatusTransition validates if a status transition is allowed
func (s *TaskValidationService) ValidateStatusTransition(
	ctx context.Context,
	taskID valueobjects.TaskID,
	fromStatus valueobjects.TaskStatus,
	toStatus valueobjects.TaskStatus,
) error {
	// Check if transition is valid according to business rules
	if !fromStatus.CanTransitionTo(toStatus) {
		return errors.NewInvalidStatusTransitionError(fromStatus.Value(), toStatus.Value())
	}

	// Additional business rules for specific transitions
	switch toStatus.Value() {
	case valueobjects.StatusProcessing:
		return s.validateProcessingTransition(ctx, taskID)
	case valueobjects.StatusCompleted:
		return s.validateCompletionTransition(ctx, taskID)
	case valueobjects.StatusFailed:
		return s.validateFailureTransition(ctx, taskID)
	}

	return nil
}

// GetTaskHealthStatus returns the health status of a task
func (s *TaskValidationService) GetTaskHealthStatus(task *entities.Task) TaskHealthStatus {
	now := time.Now()
	
	// Check if task is long-running (more than 1 hour)
	if task.IsLongRunning(time.Hour) {
		return TaskHealthStatus{
			Status:  HealthWarning,
			Message: "Task has been running for more than 1 hour",
			Issues:  []string{"long_running"},
		}
	}

	// Check if task is stuck (more than 4 hours)
	if task.IsLongRunning(4 * time.Hour) {
		return TaskHealthStatus{
			Status:  HealthCritical,
			Message: "Task appears to be stuck (running for more than 4 hours)",
			Issues:  []string{"stuck", "long_running"},
		}
	}

	// Check if task is very old but not completed
	if task.Status().IsActive() && now.Sub(task.CreatedAt()) > 24*time.Hour {
		return TaskHealthStatus{
			Status:  HealthWarning,
			Message: "Task is more than 24 hours old but not completed",
			Issues:  []string{"stale"},
		}
	}

	return TaskHealthStatus{
		Status:  HealthOK,
		Message: "Task is healthy",
		Issues:  []string{},
	}
}

// Private validation methods

func (s *TaskValidationService) validateProcessingTransition(ctx context.Context, taskID valueobjects.TaskID) error {
	// Business rule: Check if user has too many processing tasks
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	processingStatus, _ := valueobjects.NewTaskStatus(valueobjects.StatusProcessing)
	processingTasks, err := s.taskRepo.GetByStatus(ctx, processingStatus, 10, 0)
	if err != nil {
		return errors.NewTaskValidationFailedError("failed to check processing tasks", err)
	}

	// Count processing tasks for this user
	userProcessingCount := 0
	for _, t := range processingTasks {
		if t.UserID().Equals(task.UserID()) {
			userProcessingCount++
		}
	}

	// Business rule: max 3 processing tasks per user
	if userProcessingCount >= 3 {
		return errors.NewTaskValidationFailedError("user has reached maximum processing task limit (3)", nil)
	}

	return nil
}

func (s *TaskValidationService) validateCompletionTransition(ctx context.Context, taskID valueobjects.TaskID) error {
	// Business rule: Task must have been processing to be completed
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	if !task.Status().IsProcessing() {
		return errors.NewInvalidStatusTransitionError(task.Status().Value(), valueobjects.StatusCompleted)
	}

	// Business rule: Task must have been started (have a started_at timestamp)
	if task.StartedAt() == nil {
		return errors.NewTaskValidationFailedError("task cannot be completed without being started", nil)
	}

	return nil
}

func (s *TaskValidationService) validateFailureTransition(ctx context.Context, taskID valueobjects.TaskID) error {
	// For now, any active task can be marked as failed
	// Additional business rules could be added here
	return nil
}

// TaskHealthStatus represents the health status of a task
type TaskHealthStatus struct {
	Status  HealthStatus
	Message string
	Issues  []string
}

// HealthStatus represents different health levels
type HealthStatus string

const (
	HealthOK       HealthStatus = "ok"
	HealthWarning  HealthStatus = "warning"
	HealthCritical HealthStatus = "critical"
)