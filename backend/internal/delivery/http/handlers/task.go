package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"ai-git-workbench/internal/domain/entities"
	"ai-git-workbench/internal/domain/repositories"
)

// TaskHandler handles task-related endpoints
type TaskHandler struct {
	taskRepo repositories.TaskRepository
}

// NewTaskHandler creates a new TaskHandler
func NewTaskHandler(taskRepo repositories.TaskRepository) *TaskHandler {
	return &TaskHandler{
		taskRepo: taskRepo,
	}
}

// Task represents a task entity for API responses
type Task struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      string            `json:"status"`
	Repository  string            `json:"repository"`
	UserID      string            `json:"user_id"`
	Epic        string            `json:"epic,omitempty"`
	Branch      string            `json:"branch,omitempty"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
	StartedAt   string            `json:"started_at,omitempty"`
	CompletedAt string            `json:"completed_at,omitempty"`
	TokensUsed  int               `json:"tokens_used"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// GetTasks returns all tasks
func (h *TaskHandler) GetTasks(c echo.Context) error {
	ctx := c.Request().Context()
	
	// Parse query parameters for filtering
	userID := c.QueryParam("user_id")
	status := c.QueryParam("status")
	repository := c.QueryParam("repository")
	
	filter := repositories.TaskFilter{
		UserID:     userID,
		Repository: repository,
	}
	
	// Validate status if provided
	if status != "" {
		if !entities.IsValidStatus(status) {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid status value")
		}
		filter.Status = entities.TaskStatus(status)
	}
	
	// Get tasks from repository
	tasks, err := h.taskRepo.List(ctx, filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error retrieving tasks: "+err.Error())
	}
	
	// Convert to API response format
	apiTasks := make([]Task, len(tasks))
	for i, task := range tasks {
		apiTasks[i] = h.entityToAPI(task)
	}
	
	// Get total count
	total, err := h.taskRepo.Count(ctx, filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error counting tasks: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"tasks": apiTasks,
		"total": total,
		"status": "success",
	})
}

// GetTask returns a single task by ID
func (h *TaskHandler) GetTask(c echo.Context) error {
	ctx := c.Request().Context()
	taskID := c.Param("id")
	
	// Get task from repository
	task, err := h.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Task not found: "+err.Error())
	}
	
	// Convert to API response format
	apiTask := h.entityToAPI(task)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"task": apiTask,
		"status": "success",
	})
}

// CreateTask creates a new task
func (h *TaskHandler) CreateTask(c echo.Context) error {
	ctx := c.Request().Context()
	
	var req Task
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}
	
	// Validate required fields
	if req.Title == "" || req.Repository == "" || req.UserID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing required fields: title, repository, user_id")
	}
	
	// Convert API request to domain entity
	task := &entities.Task{
		ID:         uuid.New().String(),
		BranchName: req.Branch,
		Title:      req.Title,
		Content:    req.Description,
		Repository: req.Repository,
		UserID:     req.UserID,
		Status:     entities.TaskStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Metadata:   req.Metadata,
	}
	
	// Create task in repository
	if err := h.taskRepo.Create(ctx, task); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error creating task: "+err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Task created successfully",
		"task_id": task.ID,
		"status": "success",
	})
}

// UpdateTask updates an existing task
func (h *TaskHandler) UpdateTask(c echo.Context) error {
	ctx := c.Request().Context()
	taskID := c.Param("id")
	
	var req Task
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}
	
	// Get existing task first
	existingTask, err := h.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Task not found: "+err.Error())
	}
	
	// Update fields if provided
	if req.Title != "" {
		existingTask.Title = req.Title
	}
	if req.Description != "" {
		existingTask.Content = req.Description
	}
	if req.Branch != "" {
		existingTask.BranchName = req.Branch
	}
	if req.Repository != "" {
		existingTask.Repository = req.Repository
	}
	if req.Status != "" {
		if !entities.IsValidStatus(req.Status) {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid status value")
		}
		newStatus := entities.TaskStatus(req.Status)
		if !existingTask.CanTransitionTo(newStatus) {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid status transition")
		}
		existingTask.Status = newStatus
	}
	if req.Metadata != nil {
		existingTask.Metadata = req.Metadata
	}
	
	existingTask.UpdatedAt = time.Now()
	
	// Update task in repository
	if err := h.taskRepo.Update(ctx, existingTask); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error updating task: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Task updated successfully",
		"task_id": taskID,
		"status": "success",
	})
}

// DeleteTask deletes a task
func (h *TaskHandler) DeleteTask(c echo.Context) error {
	ctx := c.Request().Context()
	taskID := c.Param("id")
	
	// Delete task from repository
	if err := h.taskRepo.Delete(ctx, taskID); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Task not found: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Task deleted successfully", 
		"task_id": taskID,
		"status": "success",
	})
}

// Helper function to convert domain entity to API response
func (h *TaskHandler) entityToAPI(task *entities.Task) Task {
	return Task{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Content,
		Status:      string(task.Status),
		Repository:  task.Repository,
		UserID:      task.UserID,
		Epic:        "", // Epic is not part of the current domain model
		Branch:      task.BranchName,
		CreatedAt:   task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   task.UpdatedAt.Format(time.RFC3339),
		StartedAt:   "", // StartedAt is not part of the current domain model
		CompletedAt: "", // CompletedAt is not part of the current domain model
		TokensUsed:  0,  // TokensUsed is not part of the current domain model
		Metadata:    task.Metadata,
	}
}