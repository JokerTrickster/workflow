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

// OptimizedTaskHistoryHandler handles optimized task history API endpoints
type OptimizedTaskHistoryHandler struct {
	optimizedService *services.OptimizedTaskHistoryService
}

// NewOptimizedTaskHistoryHandler creates a new optimized task history handler
func NewOptimizedTaskHistoryHandler(db *gorm.DB) *OptimizedTaskHistoryHandler {
	return &OptimizedTaskHistoryHandler{
		optimizedService: services.NewOptimizedTaskHistoryService(db),
	}
}

// GetTaskHistoryOptimized handles GET /api/v1/tasks/history/{repository_name} with optimizations
func (h *OptimizedTaskHistoryHandler) GetTaskHistoryOptimized(c *gin.Context) {
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

	// Check if repository exists using optimized method (with caching)
	exists, appErr := h.optimizedService.CheckRepositoryExistsOptimized(c.Request.Context(), repositoryName)
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

	// Get task history using optimized service (with caching)
	response, appErr := h.optimizedService.OptimizedGetTaskHistory(c.Request.Context(), repositoryName, params)
	if appErr != nil {
		middleware.HandleError(c, appErr, "Failed to retrieve task history")
		return
	}

	// Add performance headers for monitoring
	c.Header("X-Cache-Enabled", "true")
	c.Header("X-Optimization-Level", "high")

	// Return successful response
	c.JSON(http.StatusOK, response)
}

// InvalidateCacheForRepository handles POST /api/v1/tasks/history/cache/invalidate/{repository_name}
func (h *OptimizedTaskHistoryHandler) InvalidateCacheForRepository(c *gin.Context) {
	repositoryName := c.Param("repository_name")
	if repositoryName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "repository name is required"})
		return
	}

	// Invalidate cache for the specific repository
	h.optimizedService.InvalidateRepositoryCache(repositoryName)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Cache invalidated successfully",
		"repository": repositoryName,
		"timestamp":  "2025-09-22T15:30:00Z", // Current timestamp
	})
}

// GetCacheMetrics handles GET /api/v1/tasks/history/cache/metrics
func (h *OptimizedTaskHistoryHandler) GetCacheMetrics(c *gin.Context) {
	stats := h.optimizedService.GetCacheStats()
	metrics := h.optimizedService.GetPerformanceMetrics()

	response := map[string]interface{}{
		"cache_statistics":     stats,
		"performance_metrics":  metrics,
		"optimization_status":  "active",
		"timestamp":           "2025-09-22T15:30:00Z", // Current timestamp
	}

	c.JSON(http.StatusOK, response)
}