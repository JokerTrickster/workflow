package services

import (
	"context"

	"ai-git-workbench/internal/domain/entities"
	"ai-git-workbench/internal/domain/errors"
	"ai-git-workbench/internal/domain/repositories"
	"ai-git-workbench/internal/domain/valueobjects"
)

// TaskLifecycleService manages the lifecycle of tasks
type TaskLifecycleService struct {
	taskRepo       repositories.TaskRepository
	queueRepo      repositories.QueueRepository
	validationSvc  *TaskValidationService
}

// NewTaskLifecycleService creates a new TaskLifecycleService
func NewTaskLifecycleService(
	taskRepo repositories.TaskRepository,
	queueRepo repositories.QueueRepository,
	validationSvc *TaskValidationService,
) *TaskLifecycleService {
	return &TaskLifecycleService{
		taskRepo:      taskRepo,
		queueRepo:     queueRepo,
		validationSvc: validationSvc,
	}
}

// CreateAndEnqueue creates a new task and adds it to the queue
func (s *TaskLifecycleService) CreateAndEnqueue(ctx context.Context, task *entities.Task) error {
	// Validate task creation
	if err := s.validationSvc.ValidateTaskCreation(ctx, task); err != nil {
		return err
	}

	// Use transaction to ensure atomicity
	return s.taskRepo.WithTransaction(ctx, func(repo repositories.TaskRepository) error {
		// Save task to repository
		if err := repo.Save(ctx, task); err != nil {
			return errors.NewTaskValidationFailedError("failed to save task", err)
		}

		// Add to queue
		if err := s.queueRepo.Enqueue(ctx, task); err != nil {
			return errors.NewTaskValidationFailedError("failed to enqueue task", err)
		}

		return nil
	})
}

// StartProcessing transitions a task to processing status
func (s *TaskLifecycleService) StartProcessing(ctx context.Context, taskID valueobjects.TaskID, workerID string) error {
	// Get task with version for optimistic locking
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	// Validate status transition
	currentStatus := task.Status()
	processingStatus, _ := valueobjects.NewTaskStatus(valueobjects.StatusProcessing)
	
	if err := s.validationSvc.ValidateStatusTransition(ctx, taskID, currentStatus, processingStatus); err != nil {
		return err
	}

	// Use transaction for atomicity
	return s.taskRepo.WithTransaction(ctx, func(repo repositories.TaskRepository) error {
		// Start processing on the entity
		if err := task.StartProcessing(); err != nil {
			return err
		}

		// Update in repository
		if err := repo.Update(ctx, task); err != nil {
			return errors.NewTaskValidationFailedError("failed to update task status", err)
		}

		// Assign to worker in queue
		if err := s.queueRepo.AssignTaskToWorker(ctx, taskID, workerID); err != nil {
			return errors.NewTaskValidationFailedError("failed to assign task to worker", err)
		}

		return nil
	})
}

// CompleteTask marks a task as completed
func (s *TaskLifecycleService) CompleteTask(ctx context.Context, taskID valueobjects.TaskID, tokensUsed int) error {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	// Validate completion
	currentStatus := task.Status()
	completedStatus, _ := valueobjects.NewTaskStatus(valueobjects.StatusCompleted)
	
	if err := s.validationSvc.ValidateStatusTransition(ctx, taskID, currentStatus, completedStatus); err != nil {
		return err
	}

	return s.taskRepo.WithTransaction(ctx, func(repo repositories.TaskRepository) error {
		// Add tokens used
		if tokensUsed > 0 {
			if err := task.AddTokensUsed(tokensUsed); err != nil {
				return err
			}
		}

		// Complete the task
		if err := task.Complete(); err != nil {
			return err
		}

		// Update in repository
		if err := repo.Update(ctx, task); err != nil {
			return errors.NewTaskValidationFailedError("failed to update completed task", err)
		}

		// Release from worker
		if err := s.queueRepo.ReleaseTaskFromWorker(ctx, taskID); err != nil {
			return errors.NewTaskValidationFailedError("failed to release task from worker", err)
		}

		return nil
	})
}

// FailTask marks a task as failed
func (s *TaskLifecycleService) FailTask(ctx context.Context, taskID valueobjects.TaskID, reason string, tokensUsed int) error {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	// Validate failure transition
	currentStatus := task.Status()
	failedStatus, _ := valueobjects.NewTaskStatus(valueobjects.StatusFailed)
	
	if err := s.validationSvc.ValidateStatusTransition(ctx, taskID, currentStatus, failedStatus); err != nil {
		return err
	}

	return s.taskRepo.WithTransaction(ctx, func(repo repositories.TaskRepository) error {
		// Add tokens used if any
		if tokensUsed > 0 {
			if err := task.AddTokensUsed(tokensUsed); err != nil {
				return err
			}
		}

		// Set failure reason in metadata
		if reason != "" {
			if err := task.SetMetadata("failure_reason", reason); err != nil {
				return err
			}
		}

		// Fail the task
		if err := task.Fail(); err != nil {
			return err
		}

		// Update in repository
		if err := repo.Update(ctx, task); err != nil {
			return errors.NewTaskValidationFailedError("failed to update failed task", err)
		}

		// Move to dead letter queue
		if err := s.queueRepo.MoveToDeadLetter(ctx, task, reason); err != nil {
			return errors.NewTaskValidationFailedError("failed to move task to dead letter queue", err)
		}

		// Release from worker
		if err := s.queueRepo.ReleaseTaskFromWorker(ctx, taskID); err != nil {
			return errors.NewTaskValidationFailedError("failed to release task from worker", err)
		}

		return nil
	})
}

// GetTaskWithHealth retrieves a task with its health status
func (s *TaskLifecycleService) GetTaskWithHealth(ctx context.Context, taskID valueobjects.TaskID) (*entities.Task, TaskHealthStatus, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, TaskHealthStatus{}, err
	}

	health := s.validationSvc.GetTaskHealthStatus(task)
	return task, health, nil
}

// GetQueueStatistics returns comprehensive queue and task statistics
func (s *TaskLifecycleService) GetQueueStatistics(ctx context.Context) (*repositories.QueueStatistics, error) {
	return s.queueRepo.GetQueueStatistics(ctx)
}