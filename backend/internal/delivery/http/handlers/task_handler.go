package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"ai-git-workbench/internal/application/dto"
	appErrors "ai-git-workbench/internal/application/errors"
	"ai-git-workbench/internal/application/interfaces"
)

// TaskHandler handles task-related HTTP endpoints using clean architecture
type TaskHandler struct {
	taskUsecase interfaces.TaskUsecase
}

// NewTaskHandler creates a new TaskHandler with task use case
func NewTaskHandler(taskUsecase interfaces.TaskUsecase) *TaskHandler {
	return &TaskHandler{
		taskUsecase: taskUsecase,
	}
}

// StandardResponse represents the standard API response format
type StandardResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// ErrorInfo represents error information in API responses
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// CreateTask handles POST /api/tasks - Create new task
func (h *TaskHandler) CreateTask(c echo.Context) error {
	ctx := c.Request().Context()

	// Parse request body
	var req dto.CreateTaskRequest
	if err := c.Bind(&req); err != nil {
		return h.respondError(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid JSON format", err.Error())
	}

	// Create task using use case
	response, err := h.taskUsecase.CreateTask(ctx, req)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	return h.respondSuccess(c, http.StatusCreated, response)
}

// GetTask handles GET /api/tasks/{id} - Get task by ID
func (h *TaskHandler) GetTask(c echo.Context) error {
	ctx := c.Request().Context()

	// Extract path parameters
	taskID := c.Param("id")
	if taskID == "" {
		return h.respondError(c, http.StatusBadRequest, "MISSING_TASK_ID", "Task ID is required", "")
	}

	// Extract user ID from query parameter or auth context
	// For now, using query parameter - in production, this would come from JWT token
	userID := c.QueryParam("user_id")
	if userID == "" {
		return h.respondError(c, http.StatusBadRequest, "MISSING_USER_ID", "User ID is required", "")
	}

	// Get task using use case
	response, err := h.taskUsecase.GetTask(ctx, taskID, userID)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	return h.respondSuccess(c, http.StatusOK, response)
}

// ListTasks handles GET /api/tasks - List tasks with filtering
func (h *TaskHandler) ListTasks(c echo.Context) error {
	ctx := c.Request().Context()

	// Parse query parameters
	req := dto.ListTasksRequest{}
	
	// User ID filter
	if userID := c.QueryParam("user_id"); userID != "" {
		req.UserID = &userID
	}

	// Status filter
	if status := c.QueryParam("status"); status != "" {
		req.Status = &status
	}

	// Repository filter
	if repository := c.QueryParam("repository"); repository != "" {
		req.Repository = &repository
	}

	// Epic filter
	if epic := c.QueryParam("epic"); epic != "" {
		req.Epic = &epic
	}

	// Branch filter
	if branch := c.QueryParam("branch"); branch != "" {
		req.Branch = &branch
	}

	// Title contains filter
	if titleContains := c.QueryParam("title_contains"); titleContains != "" {
		req.TitleContains = &titleContains
	}

	// Order by
	if orderBy := c.QueryParam("order_by"); orderBy != "" {
		req.OrderBy = orderBy
	}

	// Order direction
	if orderDirection := c.QueryParam("order_direction"); orderDirection != "" {
		req.OrderDirection = strings.ToUpper(orderDirection)
	}

	// Pagination
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}
	if req.Limit == 0 {
		req.Limit = 20 // Default limit
	}

	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	// List tasks using use case
	response, err := h.taskUsecase.ListTasks(ctx, req)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	return h.respondSuccess(c, http.StatusOK, response)
}

// UpdateTask handles PUT /api/tasks/{id} - Update task
func (h *TaskHandler) UpdateTask(c echo.Context) error {
	ctx := c.Request().Context()

	// Extract path parameters
	taskID := c.Param("id")
	if taskID == "" {
		return h.respondError(c, http.StatusBadRequest, "MISSING_TASK_ID", "Task ID is required", "")
	}

	// Extract user ID from query parameter or auth context
	userID := c.QueryParam("user_id")
	if userID == "" {
		return h.respondError(c, http.StatusBadRequest, "MISSING_USER_ID", "User ID is required", "")
	}

	// Parse request body
	var req dto.UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		return h.respondError(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid JSON format", err.Error())
	}

	// Update task using use case
	response, err := h.taskUsecase.UpdateTask(ctx, taskID, userID, req)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	return h.respondSuccess(c, http.StatusOK, response)
}

// DeleteTask handles DELETE /api/tasks/{id} - Cancel/delete task
func (h *TaskHandler) DeleteTask(c echo.Context) error {
	ctx := c.Request().Context()

	// Extract path parameters
	taskID := c.Param("id")
	if taskID == "" {
		return h.respondError(c, http.StatusBadRequest, "MISSING_TASK_ID", "Task ID is required", "")
	}

	// Extract user ID from query parameter or auth context
	userID := c.QueryParam("user_id")
	if userID == "" {
		return h.respondError(c, http.StatusBadRequest, "MISSING_USER_ID", "User ID is required", "")
	}

	// Parse optional request body for reason
	var req dto.TaskActionRequest
	c.Bind(&req) // Ignore errors as body is optional

	// Delete task using use case
	response, err := h.taskUsecase.DeleteTask(ctx, taskID, userID, req)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	return h.respondSuccess(c, http.StatusOK, response)
}

// ResumeTask handles PUT /api/tasks/{id}/resume - Resume failed/cancelled task
func (h *TaskHandler) ResumeTask(c echo.Context) error {
	ctx := c.Request().Context()

	// Extract path parameters
	taskID := c.Param("id")
	if taskID == "" {
		return h.respondError(c, http.StatusBadRequest, "MISSING_TASK_ID", "Task ID is required", "")
	}

	// Extract user ID from query parameter or auth context
	userID := c.QueryParam("user_id")
	if userID == "" {
		return h.respondError(c, http.StatusBadRequest, "MISSING_USER_ID", "User ID is required", "")
	}

	// Parse optional request body for reason
	var req dto.TaskActionRequest
	c.Bind(&req) // Ignore errors as body is optional

	// Resume task using use case
	response, err := h.taskUsecase.ResumeTask(ctx, taskID, userID, req)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	return h.respondSuccess(c, http.StatusOK, response)
}

// GetTaskHealth handles GET /api/tasks/{id}/health - Get task health
func (h *TaskHandler) GetTaskHealth(c echo.Context) error {
	ctx := c.Request().Context()

	// Extract path parameters
	taskID := c.Param("id")
	if taskID == "" {
		return h.respondError(c, http.StatusBadRequest, "MISSING_TASK_ID", "Task ID is required", "")
	}

	// Extract user ID from query parameter or auth context
	userID := c.QueryParam("user_id")
	if userID == "" {
		return h.respondError(c, http.StatusBadRequest, "MISSING_USER_ID", "User ID is required", "")
	}

	// Get task health using use case
	response, err := h.taskUsecase.GetTaskHealth(ctx, taskID, userID)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	return h.respondSuccess(c, http.StatusOK, response)
}

// GetUserTaskStatistics handles GET /api/users/{id}/stats - Get user task statistics
func (h *TaskHandler) GetUserTaskStatistics(c echo.Context) error {
	ctx := c.Request().Context()

	// Extract path parameters
	userID := c.Param("id")
	if userID == "" {
		return h.respondError(c, http.StatusBadRequest, "MISSING_USER_ID", "User ID is required", "")
	}

	// Get task statistics using use case
	response, err := h.taskUsecase.GetTaskStatistics(ctx, userID)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	return h.respondSuccess(c, http.StatusOK, response)
}

// GetQueueStatistics handles GET /api/queue/stats - Get queue statistics
func (h *TaskHandler) GetQueueStatistics(c echo.Context) error {
	ctx := c.Request().Context()

	// Get queue statistics using use case
	response, err := h.taskUsecase.GetQueueStatistics(ctx)
	if err != nil {
		return h.handleApplicationError(c, err)
	}

	return h.respondSuccess(c, http.StatusOK, response)
}

// Helper methods for response handling

// respondSuccess sends a successful response
func (h *TaskHandler) respondSuccess(c echo.Context, status int, data interface{}) error {
	response := StandardResponse{
		Success:   true,
		Data:      data,
		Timestamp: getCurrentTimestamp(),
	}
	return c.JSON(status, response)
}

// respondError sends an error response
func (h *TaskHandler) respondError(c echo.Context, status int, code, message, details string) error {
	response := StandardResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: getCurrentTimestamp(),
	}
	return c.JSON(status, response)
}

// handleApplicationError handles application errors and converts them to HTTP responses
func (h *TaskHandler) handleApplicationError(c echo.Context, err error) error {
	if appErr, ok := err.(appErrors.ApplicationError); ok {
		return h.respondError(c, appErr.HTTPStatus, appErr.Code, appErr.Message, getErrorDetails(appErr.Cause))
	}

	// Map other error types
	status := appErrors.GetHTTPStatus(err)
	if appErrors.IsNotFoundError(err) {
		return h.respondError(c, status, "RESOURCE_NOT_FOUND", err.Error(), "")
	}
	if appErrors.IsValidationError(err) {
		return h.respondError(c, status, "VALIDATION_ERROR", err.Error(), "")
	}
	if appErrors.IsConflictError(err) {
		return h.respondError(c, status, "CONFLICT", err.Error(), "")
	}
	if appErrors.IsUnauthorizedError(err) {
		return h.respondError(c, status, "UNAUTHORIZED", err.Error(), "")
	}
	if appErrors.IsForbiddenError(err) {
		return h.respondError(c, status, "FORBIDDEN", err.Error(), "")
	}

	// Default to internal server error
	return h.respondError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An internal error occurred", err.Error())
}

// Helper functions

// getCurrentTimestamp returns the current timestamp in RFC3339 format
func getCurrentTimestamp() string {
	if ctx := context.Background(); ctx != nil {
		if timestamp := ctx.Value("timestamp"); timestamp != nil {
			return timestamp.(string)
		}
	}
	return time.Now().Format(time.RFC3339)
}

// getErrorDetails extracts details from an error cause
func getErrorDetails(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}