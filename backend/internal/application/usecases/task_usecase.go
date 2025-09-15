package usecases

import (
	"context"
	"log"
	"time"

	"ai-git-workbench/internal/application/dto"
	appErrors "ai-git-workbench/internal/application/errors"
	"ai-git-workbench/internal/application/interfaces"
	"ai-git-workbench/internal/domain/entities"
	"ai-git-workbench/internal/domain/repositories"
	"ai-git-workbench/internal/domain/services"
	"ai-git-workbench/internal/domain/valueobjects"
)

// TaskUsecaseImpl implements the TaskUsecase interface
type TaskUsecaseImpl struct {
	taskRepo          repositories.TaskRepository
	queueRepo         repositories.QueueRepository
	validationService *services.TaskValidationService
	lifecycleService  *services.TaskLifecycleService
	authService       interfaces.TaskAuthorizationService
	eventService      interfaces.TaskEventService
}

// NewTaskUsecase creates a new task use case
func NewTaskUsecase(
	taskRepo repositories.TaskRepository,
	queueRepo repositories.QueueRepository,
	validationService *services.TaskValidationService,
	lifecycleService *services.TaskLifecycleService,
	authService interfaces.TaskAuthorizationService,
	eventService interfaces.TaskEventService,
) interfaces.TaskUsecase {
	return &TaskUsecaseImpl{
		taskRepo:          taskRepo,
		queueRepo:         queueRepo,
		validationService: validationService,
		lifecycleService:  lifecycleService,
		authService:       authService,
		eventService:      eventService,
	}
}

// CreateTask creates a new task with business validation and queuing
func (u *TaskUsecaseImpl) CreateTask(ctx context.Context, req dto.CreateTaskRequest) (*dto.CreateTaskResponse, error) {
	log.Printf("🎯 Creating task: %s for user: %s", req.Title, req.UserID)

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, appErrors.NewValidationError("invalid request", err)
	}

	// Convert to value objects
	userID, repository, branch, err := req.ToValueObjects()
	if err != nil {
		return nil, appErrors.NewValidationError("invalid value objects", err)
	}

	// Create task entity
	task, err := entities.CreateTask(
		userID,
		req.Title,
		req.Description,
		repository,
		req.Epic,
		branch,
	)
	if err != nil {
		return nil, appErrors.TranslateDomainError(err)
	}

	// Use domain service to create and enqueue
	if err := u.lifecycleService.CreateAndEnqueue(ctx, task); err != nil {
		log.Printf("❌ Failed to create and enqueue task %s: %v", task.ID().Value(), err)
		return nil, appErrors.TranslateDomainError(err)
	}

	// Publish event
	taskResponse := dto.ToGetTaskResponse(task)
	if err := u.eventService.PublishTaskCreated(ctx, &taskResponse); err != nil {
		log.Printf("⚠️ Failed to publish task created event: %v", err)
		// Don't fail the operation for event publishing errors
	}

	log.Printf("✅ Task created successfully: %s", task.ID().Value())
	return &dto.CreateTaskResponse{
		TaskID:    task.ID().Value(),
		Status:    task.Status().Value(),
		CreatedAt: task.CreatedAt().Format(time.RFC3339),
	}, nil
}

// GetTask retrieves a task by ID with authorization
func (u *TaskUsecaseImpl) GetTask(ctx context.Context, taskID string, userID string) (*dto.GetTaskResponse, error) {
	log.Printf("🔍 Getting task: %s for user: %s", taskID, userID)

	// Parse value objects
	taskIDVO, err := valueobjects.ParseTaskID(taskID)
	if err != nil {
		return nil, appErrors.NewValidationError("invalid task ID", err)
	}

	userIDVO, err := valueobjects.NewUserID(userID)
	if err != nil {
		return nil, appErrors.NewValidationError("invalid user ID", err)
	}

	// Check authorization
	if err := u.authService.CanAccessTask(ctx, userIDVO, taskIDVO); err != nil {
		return nil, err
	}

	// Retrieve task
	task, err := u.taskRepo.GetByID(ctx, taskIDVO)
	if err != nil {
		return nil, appErrors.TranslateDomainError(err)
	}

	response := dto.ToGetTaskResponse(task)
	log.Printf("✅ Task retrieved successfully: %s", taskID)
	return &response, nil
}

// UpdateTask updates a task with validation and authorization
func (u *TaskUsecaseImpl) UpdateTask(ctx context.Context, taskID string, userID string, req dto.UpdateTaskRequest) (*dto.GetTaskResponse, error) {
	log.Printf("✏️ Updating task: %s for user: %s", taskID, userID)

	// Parse value objects
	taskIDVO, err := valueobjects.ParseTaskID(taskID)
	if err != nil {
		return nil, appErrors.NewValidationError("invalid task ID", err)
	}

	userIDVO, err := valueobjects.NewUserID(userID)
	if err != nil {
		return nil, appErrors.NewValidationError("invalid user ID", err)
	}

	// Check authorization
	if err := u.authService.CanModifyTask(ctx, userIDVO, taskIDVO); err != nil {
		return nil, err
	}

	// Get current task
	task, err := u.taskRepo.GetByID(ctx, taskIDVO)
	if err != nil {
		return nil, appErrors.TranslateDomainError(err)
	}

	// Validate version for optimistic locking
	if task.Version() != req.Version {
		return nil, appErrors.NewConflictError("task was modified by another process", nil)
	}

	// Apply updates
	if req.Title != nil {
		if err := task.UpdateTitle(*req.Title); err != nil {
			return nil, appErrors.TranslateDomainError(err)
		}
	}

	if req.Description != nil {
		if err := task.UpdateDescription(*req.Description); err != nil {
			return nil, appErrors.TranslateDomainError(err)
		}
	}

	// Use transaction for update
	err = u.taskRepo.WithTransaction(ctx, func(repo repositories.TaskRepository) error {
		// Validate update
		if err := u.validationService.ValidateTaskUpdate(ctx, task); err != nil {
			return err
		}

		// Save updated task
		return repo.Update(ctx, task)
	})

	if err != nil {
		return nil, appErrors.TranslateDomainError(err)
	}

	// Publish event
	taskResponse := dto.ToGetTaskResponse(task)
	if err := u.eventService.PublishTaskUpdated(ctx, &taskResponse); err != nil {
		log.Printf("⚠️ Failed to publish task updated event: %v", err)
	}

	log.Printf("✅ Task updated successfully: %s", taskID)
	return &taskResponse, nil
}

// DeleteTask cancels/deletes a task with authorization
func (u *TaskUsecaseImpl) DeleteTask(ctx context.Context, taskID string, userID string, req dto.TaskActionRequest) (*dto.TaskActionResponse, error) {
	log.Printf("🗑️ Deleting task: %s for user: %s", taskID, userID)

	// Parse value objects
	taskIDVO, err := valueobjects.ParseTaskID(taskID)
	if err != nil {
		return nil, appErrors.NewValidationError("invalid task ID", err)
	}

	userIDVO, err := valueobjects.NewUserID(userID)
	if err != nil {
		return nil, appErrors.NewValidationError("invalid user ID", err)
	}

	// Check authorization
	if err := u.authService.CanDeleteTask(ctx, userIDVO, taskIDVO); err != nil {
		return nil, err
	}

	// Get task
	task, err := u.taskRepo.GetByID(ctx, taskIDVO)
	if err != nil {
		return nil, appErrors.TranslateDomainError(err)
	}

	// Use transaction for deletion
	err = u.taskRepo.WithTransaction(ctx, func(repo repositories.TaskRepository) error {
		// Cancel the task (mark as cancelled)
		if err := task.Cancel(); err != nil {
			return err
		}

		// Update in repository
		if err := repo.Update(ctx, task); err != nil {
			return err
		}

		// Remove from queue if it's still there
		if err := u.queueRepo.RemoveTask(ctx, taskIDVO); err != nil {
			log.Printf("⚠️ Failed to remove task from queue: %v", err)
			// Don't fail the transaction for queue removal errors
		}

		return nil
	})

	if err != nil {
		return nil, appErrors.TranslateDomainError(err)
	}

	// Publish event
	if err := u.eventService.PublishTaskDeleted(ctx, taskID, userID); err != nil {
		log.Printf("⚠️ Failed to publish task deleted event: %v", err)
	}

	log.Printf("✅ Task deleted successfully: %s", taskID)
	return &dto.TaskActionResponse{
		TaskID:  taskID,
		Status:  task.Status().Value(),
		Message: "Task cancelled successfully",
	}, nil
}

// ListTasks lists tasks with filtering and authorization
func (u *TaskUsecaseImpl) ListTasks(ctx context.Context, req dto.ListTasksRequest) (*dto.ListTasksResponse, error) {
	log.Printf("📋 Listing tasks with filter: %+v", req)

	// Set defaults
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	if req.OrderBy == "" {
		req.OrderBy = "created_at"
	}
	if req.OrderDirection == "" {
		req.OrderDirection = "DESC"
	}

	// Build query options
	queryOptions := repositories.QueryOptions{
		Filter:    buildTaskFilter(&req),
		OrderBy:   repositories.TaskOrderBy(req.OrderBy),
		Direction: repositories.TaskOrderDirection(req.OrderDirection),
		Limit:     req.Limit,
		Offset:    req.Offset,
	}

	// Execute query using extended repository interface if available
	var tasks []*entities.Task
	var total int
	var err error

	if extRepo, ok := u.taskRepo.(repositories.ExtendedTaskRepository); ok {
		tasks, total, err = extRepo.FindTasks(ctx, queryOptions)
	} else {
		// Fallback to basic repository methods
		tasks, err = u.executeBasicQuery(ctx, &req)
		if err == nil {
			total = len(tasks) // Approximate total
		}
	}

	if err != nil {
		return nil, appErrors.TranslateDomainError(err)
	}

	response := dto.ToListTasksResponse(tasks, total, req.Limit, req.Offset)
	log.Printf("✅ Listed %d tasks (total: %d)", len(tasks), total)
	return &response, nil
}

// ResumeTask resumes a failed or cancelled task
func (u *TaskUsecaseImpl) ResumeTask(ctx context.Context, taskID string, userID string, req dto.TaskActionRequest) (*dto.TaskActionResponse, error) {
	log.Printf("▶️ Resuming task: %s for user: %s", taskID, userID)

	// Parse value objects
	taskIDVO, err := valueobjects.ParseTaskID(taskID)
	if err != nil {
		return nil, appErrors.NewValidationError("invalid task ID", err)
	}

	userIDVO, err := valueobjects.NewUserID(userID)
	if err != nil {
		return nil, appErrors.NewValidationError("invalid user ID", err)
	}

	// Check authorization
	if err := u.authService.CanModifyTask(ctx, userIDVO, taskIDVO); err != nil {
		return nil, err
	}

	// Get task
	task, err := u.taskRepo.GetByID(ctx, taskIDVO)
	if err != nil {
		return nil, appErrors.TranslateDomainError(err)
	}

	// Validate task can be resumed
	if !task.IsFailed() && !task.IsCancelled() {
		return nil, appErrors.NewBadRequestError(
			"task cannot be resumed: only failed or cancelled tasks can be resumed",
			nil,
		)
	}

	// Use transaction for resume
	err = u.taskRepo.WithTransaction(ctx, func(repo repositories.TaskRepository) error {
		// Restart the task (reset to pending)
		if err := task.Restart(); err != nil {
			return err
		}

		// Validate task restart
		if err := u.validationService.ValidateTaskUpdate(ctx, task); err != nil {
			return err
		}

		// Update in repository
		if err := repo.Update(ctx, task); err != nil {
			return err
		}

		// Re-enqueue the task
		if err := u.queueRepo.Enqueue(ctx, task); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, appErrors.TranslateDomainError(err)
	}

	// Publish event
	if err := u.eventService.PublishTaskResumed(ctx, taskID, req.Reason); err != nil {
		log.Printf("⚠️ Failed to publish task resumed event: %v", err)
	}

	log.Printf("✅ Task resumed successfully: %s", taskID)
	return &dto.TaskActionResponse{
		TaskID:  taskID,
		Status:  task.Status().Value(),
		Message: "Task resumed successfully",
	}, nil
}

// CancelTask cancels a task
func (u *TaskUsecaseImpl) CancelTask(ctx context.Context, taskID string, userID string, req dto.TaskActionRequest) (*dto.TaskActionResponse, error) {
	log.Printf("⏹️ Cancelling task: %s for user: %s", taskID, userID)

	// Parse value objects
	taskIDVO, err := valueobjects.ParseTaskID(taskID)
	if err != nil {
		return nil, appErrors.NewValidationError("invalid task ID", err)
	}

	userIDVO, err := valueobjects.NewUserID(userID)
	if err != nil {
		return nil, appErrors.NewValidationError("invalid user ID", err)
	}

	// Check authorization
	if err := u.authService.CanModifyTask(ctx, userIDVO, taskIDVO); err != nil {
		return nil, err
	}

	// Get task
	task, err := u.taskRepo.GetByID(ctx, taskIDVO)
	if err != nil {
		return nil, appErrors.TranslateDomainError(err)
	}

	// Validate task can be cancelled
	if task.IsCompleted() {
		return nil, appErrors.NewBadRequestError("cannot cancel completed task", nil)
	}

	// Use transaction for cancellation
	err = u.taskRepo.WithTransaction(ctx, func(repo repositories.TaskRepository) error {
		// Cancel the task
		if err := task.Cancel(); err != nil {
			return err
		}

		// Set cancellation reason if provided
		if req.Reason != "" {
			if err := task.SetMetadata("cancellation_reason", req.Reason); err != nil {
				return err
			}
		}

		// Update in repository
		if err := repo.Update(ctx, task); err != nil {
			return err
		}

		// Remove from queue
		if err := u.queueRepo.RemoveTask(ctx, taskIDVO); err != nil {
			log.Printf("⚠️ Failed to remove task from queue: %v", err)
			// Don't fail for queue removal errors
		}

		return nil
	})

	if err != nil {
		return nil, appErrors.TranslateDomainError(err)
	}

	// Publish event
	if err := u.eventService.PublishTaskStatusChanged(ctx, taskID, "pending", "cancelled"); err != nil {
		log.Printf("⚠️ Failed to publish task status changed event: %v", err)
	}

	log.Printf("✅ Task cancelled successfully: %s", taskID)
	return &dto.TaskActionResponse{
		TaskID:  taskID,
		Status:  task.Status().Value(),
		Message: "Task cancelled successfully",
	}, nil
}

// GetTaskHealth retrieves task health information
func (u *TaskUsecaseImpl) GetTaskHealth(ctx context.Context, taskID string, userID string) (*dto.TaskHealthResponse, error) {
	log.Printf("🏥 Getting task health: %s for user: %s", taskID, userID)

	// Parse value objects
	taskIDVO, err := valueobjects.ParseTaskID(taskID)
	if err != nil {
		return nil, appErrors.NewValidationError("invalid task ID", err)
	}

	userIDVO, err := valueobjects.NewUserID(userID)
	if err != nil {
		return nil, appErrors.NewValidationError("invalid user ID", err)
	}

	// Check authorization
	if err := u.authService.CanAccessTask(ctx, userIDVO, taskIDVO); err != nil {
		return nil, err
	}

	// Get task with health status
	task, health, err := u.lifecycleService.GetTaskWithHealth(ctx, taskIDVO)
	if err != nil {
		return nil, appErrors.TranslateDomainError(err)
	}

	return &dto.TaskHealthResponse{
		TaskID:  task.ID().Value(),
		Health:  string(health.Status),
		Message: health.Message,
		Issues:  health.Issues,
	}, nil
}

// GetTaskStatistics retrieves task statistics for a user
func (u *TaskUsecaseImpl) GetTaskStatistics(ctx context.Context, userID string) (*dto.TaskStatisticsResponse, error) {
	log.Printf("📊 Getting task statistics for user: %s", userID)

	// Parse user ID
	userIDVO, err := valueobjects.NewUserID(userID)
	if err != nil {
		return nil, appErrors.NewValidationError("invalid user ID", err)
	}

	// Check authorization
	if err := u.authService.CanViewStatistics(ctx, userIDVO, &userIDVO); err != nil {
		return nil, err
	}

	// Get statistics
	stats, err := u.taskRepo.GetUserTaskStatistics(ctx, userIDVO)
	if err != nil {
		return nil, appErrors.TranslateDomainError(err)
	}

	var lastActivityAt *string
	if stats.LastActivityAt != nil {
		s := stats.LastActivityAt.Format(time.RFC3339)
		lastActivityAt = &s
	}

	return &dto.TaskStatisticsResponse{
		UserID:               stats.UserID.Value(),
		TotalTasks:           stats.TotalTasks,
		CompletedTasks:       stats.CompletedTasks,
		FailedTasks:          stats.FailedTasks,
		PendingTasks:         stats.PendingTasks,
		ProcessingTasks:      stats.ProcessingTasks,
		CancelledTasks:       stats.CancelledTasks,
		TotalTokensUsed:      stats.TotalTokensUsed,
		AverageTokensPerTask: stats.AverageTokensPerTask,
		CompletionRate:       stats.CompletionRate,
		LastActivityAt:       lastActivityAt,
	}, nil
}

// GetQueueStatistics retrieves queue performance metrics
func (u *TaskUsecaseImpl) GetQueueStatistics(ctx context.Context) (*dto.QueueStatisticsResponse, error) {
	log.Printf("📈 Getting queue statistics")

	// Get statistics from lifecycle service
	stats, err := u.lifecycleService.GetQueueStatistics(ctx)
	if err != nil {
		return nil, appErrors.TranslateDomainError(err)
	}

	return &dto.QueueStatisticsResponse{
		TotalEnqueued:         stats.TotalEnqueued,
		TotalDequeued:         stats.TotalDequeued,
		TotalProcessed:        stats.TotalProcessed,
		TotalFailed:           stats.TotalFailed,
		CurrentQueueLength:    stats.CurrentQueueLength,
		AverageProcessingTime: stats.AverageProcessingTime.String(),
		ThroughputPerMinute:   stats.ThroughputPerMinute,
		ErrorRate:             stats.ErrorRate,
		DeadLetterCount:       stats.DeadLetterCount,
		WorkersActive:         stats.WorkersActive,
		LastActivityAt:        stats.LastActivityAt.Format(time.RFC3339),
	}, nil
}

// GetTasksRequiringAttention retrieves tasks that need administrative attention
func (u *TaskUsecaseImpl) GetTasksRequiringAttention(ctx context.Context) ([]dto.GetTaskResponse, error) {
	log.Printf("⚠️ Getting tasks requiring attention")

	// Get tasks requiring attention
	tasks, err := u.taskRepo.GetTasksRequiringAttention(ctx)
	if err != nil {
		return nil, appErrors.TranslateDomainError(err)
	}

	responses := make([]dto.GetTaskResponse, len(tasks))
	for i, task := range tasks {
		responses[i] = dto.ToGetTaskResponse(task)
	}

	log.Printf("📋 Found %d tasks requiring attention", len(responses))
	return responses, nil
}

// Helper methods

// buildTaskFilter builds a task filter from the list request
func buildTaskFilter(req *dto.ListTasksRequest) *repositories.TaskFilter {
	filter := &repositories.TaskFilter{}

	if req.UserID != nil {
		userID, err := valueobjects.NewUserID(*req.UserID)
		if err == nil {
			filter.UserID = &userID
		}
	}

	if req.Status != nil {
		status, err := valueobjects.NewTaskStatus(*req.Status)
		if err == nil {
			filter.Status = &status
		}
	}

	if req.Repository != nil {
		repo, err := valueobjects.NewRepositoryPath(*req.Repository)
		if err == nil {
			filter.Repository = &repo
		}
	}

	if req.Epic != nil {
		filter.Epic = req.Epic
	}

	if req.Branch != nil {
		branch, err := valueobjects.NewBranchName(*req.Branch)
		if err == nil {
			filter.Branch = &branch
		}
	}

	if req.CreatedAfter != nil {
		filter.CreatedAfter = req.CreatedAfter
	}

	if req.CreatedBefore != nil {
		filter.CreatedBefore = req.CreatedBefore
	}

	if req.TitleContains != nil {
		filter.TitleContains = req.TitleContains
	}

	return filter
}

// executeBasicQuery executes a basic query when extended repository is not available
func (u *TaskUsecaseImpl) executeBasicQuery(ctx context.Context, req *dto.ListTasksRequest) ([]*entities.Task, error) {
	// This is a simplified implementation for basic repository
	// In a real implementation, you would need to implement proper filtering

	if req.UserID != nil {
		userID, err := valueobjects.NewUserID(*req.UserID)
		if err != nil {
			return nil, err
		}
		return u.taskRepo.GetByUserID(ctx, userID, req.Limit, req.Offset)
	}

	if req.Status != nil {
		status, err := valueobjects.NewTaskStatus(*req.Status)
		if err != nil {
			return nil, err
		}
		return u.taskRepo.GetByStatus(ctx, status, req.Limit, req.Offset)
	}

	// Default to getting all tasks
	return u.taskRepo.GetAll(ctx, req.Limit, req.Offset)
}