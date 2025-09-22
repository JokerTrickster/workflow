package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"local-backend-server/internal/errors"
	"local-backend-server/internal/middleware"
	"local-backend-server/internal/services"
)

// TaskHistoryHandler handles task history API endpoints
type TaskHistoryHandler struct {
	taskHistoryService *services.TaskHistoryService
}

// NewTaskHistoryHandler creates a new TaskHistoryHandler
func NewTaskHistoryHandler(db *gorm.DB) *TaskHistoryHandler {
	return &TaskHistoryHandler{
		taskHistoryService: services.NewTaskHistoryService(db),
	}
}

// GetTaskHistory handles GET /api/tasks/history/{repository_name}
func (h *TaskHistoryHandler) GetTaskHistory(c *gin.Context) {
	// Extract repository name from path parameter
	repositoryName := c.Param("repository_name")
	if repositoryName == "" {
		appErr := errors.NewValidationError("repository name is required").
			WithDetails("repository_name path parameter cannot be empty")
		middleware.HandleError(c, appErr, "Missing repository name in path")
		return
	}

	// Parse pagination parameters with defaults
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		appErr := errors.NewValidationError("invalid page parameter").
			WithDetails("page parameter must be a valid integer")
		middleware.HandleError(c, appErr, "Invalid page parameter")
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		appErr := errors.NewValidationError("invalid limit parameter").
			WithDetails("limit parameter must be a valid integer")
		middleware.HandleError(c, appErr, "Invalid limit parameter")
		return
	}

	// Validate pagination parameters
	params, appErr := services.ValidatePaginationParams(page, limit)
	if appErr != nil {
		middleware.HandleError(c, appErr, "Invalid pagination parameters")
		return
	}

	// Check if repository exists (for 404 handling)
	exists, appErr := h.taskHistoryService.CheckRepositoryExists(c.Request.Context(), repositoryName)
	if appErr != nil {
		middleware.HandleError(c, appErr, "Failed to check repository existence")
		return
	}

	if !exists {
		appErr := errors.NewNotFoundError("repository", repositoryName).
			WithDetails("no task history found for repository: " + repositoryName)
		middleware.HandleError(c, appErr, "Repository not found")
		return
	}

	// Get task history
	response, appErr := h.taskHistoryService.GetTaskHistory(c.Request.Context(), repositoryName, params)
	if appErr != nil {
		middleware.HandleError(c, appErr, "Failed to retrieve task history")
		return
	}

	// Return successful response
	c.JSON(http.StatusOK, response)
}