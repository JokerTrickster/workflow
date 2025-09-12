package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// TaskHandler handles task-related endpoints
type TaskHandler struct{}

// NewTaskHandler creates a new TaskHandler
func NewTaskHandler() *TaskHandler {
	return &TaskHandler{}
}

// Task represents a task entity for API responses
type Task struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      string            `json:"status"`
	Repository  string            `json:"repository"`
	Epic        string            `json:"epic"`
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
	// Mock data for testing
	tasks := []Task{
		{
			ID:          "task-1",
			Title:       "Implement user authentication",
			Description: "Add JWT-based authentication to the API",
			Status:      "in_progress",
			Repository:  "workflow",
			Epic:        "authentication",
			Branch:      "feature/auth",
			CreatedAt:   "2024-01-15T10:00:00Z",
			UpdatedAt:   "2024-01-15T14:30:00Z",
			StartedAt:   "2024-01-15T11:00:00Z",
			TokensUsed:  1500,
		},
		{
			ID:          "task-2", 
			Title:       "Setup database migrations",
			Description: "Create initial database schema and migration system",
			Status:      "completed",
			Repository:  "workflow",
			Epic:        "infrastructure", 
			Branch:      "feature/db-setup",
			CreatedAt:   "2024-01-14T09:00:00Z",
			UpdatedAt:   "2024-01-15T16:00:00Z",
			StartedAt:   "2024-01-14T10:00:00Z",
			CompletedAt: "2024-01-15T16:00:00Z",
			TokensUsed:  2300,
		},
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"tasks": tasks,
		"total": len(tasks),
		"status": "success",
	})
}

// GetTask returns a single task by ID
func (h *TaskHandler) GetTask(c echo.Context) error {
	taskID := c.Param("id")
	
	// Mock task data
	task := Task{
		ID:          taskID,
		Title:       "Sample Task",
		Description: "This is a sample task for testing",
		Status:      "pending",
		Repository:  "workflow",
		Epic:        "development",
		CreatedAt:   "2024-01-15T10:00:00Z",
		UpdatedAt:   "2024-01-15T10:00:00Z",
		TokensUsed:  0,
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"task": task,
		"status": "success",
	})
}

// CreateTask creates a new task
func (h *TaskHandler) CreateTask(c echo.Context) error {
	var req Task
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Mock creation response
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Task created successfully",
		"task_id": "new-task-123",
		"status": "success",
	})
}

// UpdateTask updates an existing task
func (h *TaskHandler) UpdateTask(c echo.Context) error {
	taskID := c.Param("id")
	
	var req Task
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Task updated successfully",
		"task_id": taskID,
		"status": "success",
	})
}

// DeleteTask deletes a task
func (h *TaskHandler) DeleteTask(c echo.Context) error {
	taskID := c.Param("id")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Task deleted successfully", 
		"task_id": taskID,
		"status": "success",
	})
}