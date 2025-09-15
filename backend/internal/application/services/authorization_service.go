package services

import (
	"context"
	"log"

	"ai-git-workbench/internal/application/dto"
	appErrors "ai-git-workbench/internal/application/errors"
	"ai-git-workbench/internal/application/interfaces"
	"ai-git-workbench/internal/domain/repositories"
	"ai-git-workbench/internal/domain/valueobjects"
)

// AuthorizationService implements task authorization logic
type AuthorizationService struct {
	taskRepo repositories.TaskRepository
}

// NewAuthorizationService creates a new authorization service
func NewAuthorizationService(taskRepo repositories.TaskRepository) interfaces.TaskAuthorizationService {
	return &AuthorizationService{
		taskRepo: taskRepo,
	}
}

// CanAccessTask checks if the user can access the task
func (s *AuthorizationService) CanAccessTask(ctx context.Context, userID valueobjects.UserID, taskID valueobjects.TaskID) error {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return appErrors.TranslateDomainError(err)
	}

	// Users can only access their own tasks
	if !task.UserID().Equals(userID) {
		log.Printf("⚠️ User %s attempted to access task %s owned by %s", 
			userID.Value(), taskID.Value(), task.UserID().Value())
		return appErrors.NewForbiddenError("access denied: you can only access your own tasks")
	}

	return nil
}

// CanModifyTask checks if the user can modify the task
func (s *AuthorizationService) CanModifyTask(ctx context.Context, userID valueobjects.UserID, taskID valueobjects.TaskID) error {
	// First check if user can access the task
	if err := s.CanAccessTask(ctx, userID, taskID); err != nil {
		return err
	}

	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return appErrors.TranslateDomainError(err)
	}

	// Check if task can be modified (business rule)
	if !task.CanBeModified() {
		return appErrors.NewBadRequestError(
			"task cannot be modified: task is in a terminal state", 
			nil,
		)
	}

	return nil
}

// CanDeleteTask checks if the user can delete the task
func (s *AuthorizationService) CanDeleteTask(ctx context.Context, userID valueobjects.UserID, taskID valueobjects.TaskID) error {
	// First check if user can access the task
	if err := s.CanAccessTask(ctx, userID, taskID); err != nil {
		return err
	}

	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return appErrors.TranslateDomainError(err)
	}

	// Business rule: Cannot delete completed tasks (for audit purposes)
	if task.IsCompleted() {
		return appErrors.NewBadRequestError(
			"cannot delete completed task: completed tasks are retained for audit purposes", 
			nil,
		)
	}

	return nil
}

// CanListTasks checks if the user can list tasks with the given filter
func (s *AuthorizationService) CanListTasks(ctx context.Context, userID valueobjects.UserID, filter *dto.ListTasksRequest) error {
	// If no user filter is specified, default to the requesting user
	if filter.UserID == nil {
		userIDStr := userID.Value()
		filter.UserID = &userIDStr
		return nil
	}

	// If a specific user filter is provided, ensure it matches the requesting user
	// (In future implementations, admin users might be able to view other users' tasks)
	if *filter.UserID != userID.Value() {
		log.Printf("⚠️ User %s attempted to list tasks for user %s", 
			userID.Value(), *filter.UserID)
		return appErrors.NewForbiddenError("access denied: you can only list your own tasks")
	}

	return nil
}

// CanViewStatistics checks if the user can view task statistics
func (s *AuthorizationService) CanViewStatistics(ctx context.Context, userID valueobjects.UserID, targetUserID *valueobjects.UserID) error {
	// If no target user is specified, user is viewing their own statistics
	if targetUserID == nil {
		return nil
	}

	// Users can only view their own statistics
	// (In future implementations, admin users might be able to view other users' statistics)
	if !targetUserID.Equals(userID) {
		log.Printf("⚠️ User %s attempted to view statistics for user %s", 
			userID.Value(), targetUserID.Value())
		return appErrors.NewForbiddenError("access denied: you can only view your own statistics")
	}

	return nil
}

// Additional authorization methods for administrative operations

// CanAccessAdminFeatures checks if the user can access administrative features
func (s *AuthorizationService) CanAccessAdminFeatures(ctx context.Context, userID valueobjects.UserID) error {
	// For now, we'll implement a simple admin check
	// In a real implementation, this would check user roles/permissions
	
	// TODO: Implement proper role-based access control
	// For now, allow all users to access admin features for development
	return nil
}

// CanViewQueueStatistics checks if the user can view queue statistics
func (s *AuthorizationService) CanViewQueueStatistics(ctx context.Context, userID valueobjects.UserID) error {
	// Queue statistics are sensitive operational data
	return s.CanAccessAdminFeatures(ctx, userID)
}

// CanViewTasksRequiringAttention checks if the user can view tasks requiring attention
func (s *AuthorizationService) CanViewTasksRequiringAttention(ctx context.Context, userID valueobjects.UserID) error {
	// This is an administrative feature
	return s.CanAccessAdminFeatures(ctx, userID)
}

// ValidateTaskOwnership ensures the task belongs to the user
func (s *AuthorizationService) ValidateTaskOwnership(ctx context.Context, userID valueobjects.UserID, taskID valueobjects.TaskID) error {
	return s.CanAccessTask(ctx, userID, taskID)
}

// ValidateTaskModificationRights ensures the user can modify the task
func (s *AuthorizationService) ValidateTaskModificationRights(ctx context.Context, userID valueobjects.UserID, taskID valueobjects.TaskID) error {
	return s.CanModifyTask(ctx, userID, taskID)
}