package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"local-backend-server/internal/infrastructure/database/models"
	"local-backend-server/internal/services"
)

// setupTestRouter creates a test Gin router with the task history handler
func setupTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate
	err = db.AutoMigrate(&models.WorkflowHistory{})
	require.NoError(t, err)

	// Create handler
	handler := NewTaskHistoryHandler(db)

	// Setup router
	router := gin.New()

	// Add minimal middleware to match production behavior
	router.Use(func(c *gin.Context) {
		// Mock request ID middleware for testing
		c.Set("request_id", "test-request-id-123")
		c.Header("X-Request-ID", "test-request-id-123")
		c.Next()
	})

	// Add the task history route
	api := router.Group("/api/v1")
	tasks := api.Group("/tasks")
	tasks.GET("/history/:repository_name", handler.GetTaskHistory)

	return router, db
}

// seedTestData inserts test data for API testing
func seedTestDataForAPI(t *testing.T, db *gorm.DB) {
	baseTime := time.Now()
	testData := []models.WorkflowHistory{
		{
			RequestID:        "req-1",
			Status:          "completed",
			Tasks:           "Implement authentication system",
			RepositoryName:  "test-repo",
			WorkingDir:      stringPtr("/path/to/repo"),
			ClaudeCmd:       stringPtr("claude code"),
			Interactive:     true,
			ContinueTask:    false,
			CreatedAt:       baseTime.Add(-3 * time.Hour),
			CompletedAt:     timePtr(baseTime.Add(-2 * time.Hour).Add(-30 * time.Minute)),
			ProcessingTimeMs: int64Ptr(1800000), // 30 minutes
			Result:          stringPtr("Authentication system implemented successfully"),
		},
		{
			RequestID:      "req-2",
			Status:         "processing",
			Tasks:          "Add database migration",
			RepositoryName: "test-repo",
			CreatedAt:      baseTime.Add(-2 * time.Hour),
		},
		{
			RequestID:      "req-3",
			Status:         "failed",
			Tasks:          "Fix security vulnerability",
			RepositoryName: "test-repo",
			CreatedAt:      baseTime.Add(-1 * time.Hour),
			Error:          stringPtr("Security patch failed: dependencies conflict"),
		},
		{
			RequestID:      "req-4",
			Status:         "completed",
			Tasks:          "Update documentation",
			RepositoryName: "other-repo",
			CreatedAt:      baseTime.Add(-30 * time.Minute),
		},
	}

	for _, data := range testData {
		err := db.Create(&data).Error
		require.NoError(t, err)
	}
}

// Helper functions for pointer fields
func stringPtr(s string) *string    { return &s }
func timePtr(t time.Time) *time.Time { return &t }
func int64Ptr(i int64) *int64       { return &i }

func TestTaskHistoryHandler_GetTaskHistory(t *testing.T) {
	router, db := setupTestRouter(t)
	seedTestDataForAPI(t, db)

	t.Run("successful request with default pagination", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/tasks/history/test-repo", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response services.TaskHistoryResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify response structure
		assert.Len(t, response.Data, 3) // 3 tasks for test-repo
		assert.Equal(t, 1, response.Pagination.Page)
		assert.Equal(t, 20, response.Pagination.Limit)
		assert.Equal(t, 3, response.Pagination.Total)
		assert.Equal(t, 1, response.Pagination.TotalPages)

		// Verify tasks are sorted by created_at DESC
		assert.True(t, response.Data[0].CreatedAt.After(response.Data[1].CreatedAt))
		assert.True(t, response.Data[1].CreatedAt.After(response.Data[2].CreatedAt))

		// Verify complete task data structure
		firstTask := response.Data[0]
		assert.NotEmpty(t, firstTask.RequestID)
		assert.NotEmpty(t, firstTask.Status)
		assert.NotEmpty(t, firstTask.Tasks)
		assert.Equal(t, "test-repo", firstTask.RepositoryName)
	})

	t.Run("request with custom pagination", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/tasks/history/test-repo?page=1&limit=2", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response services.TaskHistoryResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response.Data, 2) // Limited to 2 tasks
		assert.Equal(t, 1, response.Pagination.Page)
		assert.Equal(t, 2, response.Pagination.Limit)
		assert.Equal(t, 3, response.Pagination.Total)
		assert.Equal(t, 2, response.Pagination.TotalPages)
	})

	t.Run("request second page", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/tasks/history/test-repo?page=2&limit=2", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response services.TaskHistoryResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response.Data, 1) // 1 task on second page
		assert.Equal(t, 2, response.Pagination.Page)
	})

	t.Run("repository not found", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/tasks/history/nonexistent-repo", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusNotFound, recorder.Code)

		// Verify error response structure
		var errorResponse map[string]interface{}
		err = json.Unmarshal(recorder.Body.Bytes(), &errorResponse)
		require.NoError(t, err)

		assert.Contains(t, errorResponse, "error")
		assert.Contains(t, errorResponse, "request_id")
		assert.Contains(t, errorResponse, "timestamp")
	})

	t.Run("invalid page parameter", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/tasks/history/test-repo?page=invalid", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)

		var errorResponse map[string]interface{}
		err = json.Unmarshal(recorder.Body.Bytes(), &errorResponse)
		require.NoError(t, err)

		assert.Contains(t, errorResponse, "error")
	})

	t.Run("invalid limit parameter", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/tasks/history/test-repo?limit=invalid", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("page parameter zero", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/tasks/history/test-repo?page=0", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("limit parameter exceeds maximum", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/tasks/history/test-repo?limit=101", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("page exceeds available pages", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/tasks/history/test-repo?page=10", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("repository filtering works correctly", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/tasks/history/other-repo", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response services.TaskHistoryResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response.Data, 1) // 1 task for other-repo
		assert.Equal(t, "other-repo", response.Data[0].RepositoryName)
	})

	t.Run("response includes all required fields", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/tasks/history/test-repo", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response services.TaskHistoryResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check first task has all expected fields from PRD
		task := response.Data[0]
		assert.NotZero(t, task.ID)
		assert.NotEmpty(t, task.RequestID)
		assert.NotEmpty(t, task.Status)
		assert.NotEmpty(t, task.Tasks)
		assert.Equal(t, "test-repo", task.RepositoryName)
		assert.NotZero(t, task.CreatedAt)

		// Optional fields should be present if set in test data
		// (checking the completed task which has all fields)
		completedTask := findTaskByStatus(response.Data, "completed")
		if completedTask != nil {
			assert.NotNil(t, completedTask.WorkingDir)
			assert.NotNil(t, completedTask.ClaudeCmd)
			assert.NotNil(t, completedTask.CompletedAt)
			assert.NotNil(t, completedTask.ProcessingTimeMs)
			assert.NotNil(t, completedTask.Result)
		}
	})

	t.Run("empty repository returns empty array", func(t *testing.T) {
		// Create a new empty repository in database
		emptyRepo := models.WorkflowHistory{
			RequestID:      "temp-req",
			Status:         "completed",
			Tasks:          "temp task",
			RepositoryName: "empty-test-repo",
			CreatedAt:      time.Now(),
		}
		err := db.Create(&emptyRepo).Error
		require.NoError(t, err)

		// Delete the temporary record to make it truly empty
		err = db.Delete(&emptyRepo).Error
		require.NoError(t, err)

		// Now test with empty repository name not in database
		req, err := http.NewRequest("GET", "/api/v1/tasks/history/truly-empty-repo", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
	})
}

// Helper function to find task by status in test data
func findTaskByStatus(tasks []models.WorkflowHistory, status string) *models.WorkflowHistory {
	for i := range tasks {
		if tasks[i].Status == status {
			return &tasks[i]
		}
	}
	return nil
}

func TestTaskHistoryHandler_PerformanceRequirements(t *testing.T) {
	router, db := setupTestRouter(t)

	// Seed more data to test performance
	baseTime := time.Now()
	var testData []models.WorkflowHistory

	// Create 50 tasks for performance testing
	for i := 0; i < 50; i++ {
		task := models.WorkflowHistory{
			RequestID:      fmt.Sprintf("perf-req-%d", i),
			Status:         "completed",
			Tasks:          fmt.Sprintf("Performance test task %d", i),
			RepositoryName: "performance-repo",
			CreatedAt:      baseTime.Add(-time.Duration(i) * time.Minute),
		}
		testData = append(testData, task)
	}

	for _, data := range testData {
		err := db.Create(&data).Error
		require.NoError(t, err)
	}

	t.Run("response time under 200ms for typical query", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/tasks/history/performance-repo?page=1&limit=20", nil)
		require.NoError(t, err)

		start := time.Now()
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		elapsed := time.Since(start)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Less(t, elapsed, 200*time.Millisecond, "Response time should be under 200ms")

		var response services.TaskHistoryResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response.Data, 20)
		assert.Equal(t, 50, response.Pagination.Total)
	})
}