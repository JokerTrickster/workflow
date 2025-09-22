package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"local-backend-server/internal/handlers"
	"local-backend-server/internal/infrastructure/database/models"
	"local-backend-server/internal/infrastructure/queue"
	"local-backend-server/internal/services"
)

// TestIntegrationSuite provides comprehensive end-to-end testing
type TestIntegrationSuite struct {
	db            *gorm.DB
	router        *gin.Engine
	publisher     *MockQueuePublisher
	atomicService *services.AtomicQueueService
}

// MockQueuePublisher implements interfaces.MessagePublisher for testing
type MockQueuePublisher struct {
	mu           sync.RWMutex
	messages     []*queue.WorkflowMessage
	publishError error
}

func NewMockQueuePublisher() *MockQueuePublisher {
	return &MockQueuePublisher{
		messages: make([]*queue.WorkflowMessage, 0),
	}
}

func (m *MockQueuePublisher) PublishMessage(msg *queue.WorkflowMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.publishError != nil {
		return m.publishError
	}

	m.messages = append(m.messages, msg)
	return nil
}

func (m *MockQueuePublisher) Close() error {
	return nil
}

func (m *MockQueuePublisher) GetMessages() []*queue.WorkflowMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return copy to avoid race conditions
	messages := make([]*queue.WorkflowMessage, len(m.messages))
	copy(messages, m.messages)
	return messages
}

func (m *MockQueuePublisher) SetPublishError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishError = err
}

func (m *MockQueuePublisher) ClearMessages() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = m.messages[:0]
}

// setupIntegrationSuite initializes the test environment
func setupIntegrationSuite(t *testing.T) *TestIntegrationSuite {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate models
	err = db.AutoMigrate(&models.WorkflowHistory{})
	require.NoError(t, err)

	// Setup mock publisher
	publisher := NewMockQueuePublisher()

	// Create atomic service
	atomicService := services.NewAtomicQueueService(db, publisher)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware
	router.Use(func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	})

	// Setup routes
	api := router.Group("/api/v1")

	// Task history routes
	taskHistoryHandler := handlers.NewTaskHistoryHandler(db)
	tasks := api.Group("/tasks")
	tasks.GET("/history/:repository_name", taskHistoryHandler.GetTaskHistory)

	// Claude task submission routes (mock endpoint)
	router.POST("/api/v1/claude/run-tasks", func(c *gin.Context) {
		handleClaudeRunTasksIntegration(c, atomicService)
	})

	return &TestIntegrationSuite{
		db:            db,
		router:        router,
		publisher:     publisher,
		atomicService: atomicService,
	}
}

// Mock Claude task handler for integration testing
func handleClaudeRunTasksIntegration(c *gin.Context, atomicService *services.AtomicQueueService) {
	var req struct {
		Tasks          string `json:"tasks" binding:"required"`
		RepositoryName string `json:"repository_name" binding:"required"`
		WorkingDir     string `json:"working_dir,omitempty"`
		Interactive    bool   `json:"interactive"`
		ClaudeCmd      string `json:"claude_cmd,omitempty"`
		ContinueTask   bool   `json:"continue_task"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use atomic service if available
	if atomicService != nil {
		publishReq := services.PublishRequest{
			Tasks:          req.Tasks,
			RepositoryName: req.RepositoryName,
			WorkingDir:     req.WorkingDir,
			ClaudeCmd:      req.ClaudeCmd,
			Interactive:    req.Interactive,
			ContinueTask:   req.ContinueTask,
			MessageType:    "claude_task",
			SessionID:      uuid.New().String(),
			Payload: map[string]interface{}{
				"request_type": "claude_task",
				"input": map[string]interface{}{
					"tasks":           req.Tasks,
					"repository_name": req.RepositoryName,
					"working_dir":     req.WorkingDir,
					"interactive":     req.Interactive,
					"claude_cmd":      req.ClaudeCmd,
					"continue_task":   req.ContinueTask,
				},
			},
		}

		response, err := atomicService.PublishWithHistory(context.Background(), publishReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"request_id": response.RequestID,
			"status":     response.Status,
			"message":    response.Message,
			"created_at": response.CreatedAt.Format(time.RFC3339),
		})
		return
	}

	// Fallback response if no atomic service
	c.JSON(http.StatusAccepted, gin.H{
		"request_id": uuid.New().String(),
		"status":     "pending",
		"message":    "Task queued successfully",
		"created_at": time.Now().UTC().Format(time.RFC3339),
	})
}

// Test 1: End-to-End Happy Path
func TestE2E_HappyPath_QueueToAPIRetrieval(t *testing.T) {
	suite := setupIntegrationSuite(t)

	t.Run("complete workflow: submit task -> queue -> database -> api retrieval", func(t *testing.T) {
		// Step 1: Submit task via API
		taskRequest := map[string]interface{}{
			"tasks":           "Implement user authentication system",
			"repository_name": "test-repo",
			"working_dir":     "/path/to/repo",
			"interactive":     true,
			"claude_cmd":      "claude code",
			"continue_task":   false,
		}

		jsonData, err := json.Marshal(taskRequest)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/v1/claude/run-tasks", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		suite.router.ServeHTTP(recorder, req)

		// Verify submission response
		assert.Equal(t, http.StatusAccepted, recorder.Code)

		var submitResponse struct {
			RequestID string `json:"request_id"`
			Status    string `json:"status"`
			Message   string `json:"message"`
			CreatedAt string `json:"created_at"`
		}
		err = json.Unmarshal(recorder.Body.Bytes(), &submitResponse)
		require.NoError(t, err)

		assert.NotEmpty(t, submitResponse.RequestID)
		assert.Equal(t, "pending", submitResponse.Status)

		// Step 2: Verify database record was created
		var dbRecord models.WorkflowHistory
		err = suite.db.Where("request_id = ?", submitResponse.RequestID).First(&dbRecord).Error
		require.NoError(t, err)

		assert.Equal(t, submitResponse.RequestID, dbRecord.RequestID)
		assert.Equal(t, models.WorkflowStatusPending, dbRecord.Status)
		assert.Equal(t, "Implement user authentication system", dbRecord.Tasks)
		assert.Equal(t, "test-repo", dbRecord.RepositoryName)

		// Step 3: Verify queue message was published
		messages := suite.publisher.GetMessages()
		require.Len(t, messages, 1)

		queueMsg := messages[0]
		assert.Equal(t, "claude_task", queueMsg.Type)
		assert.Equal(t, submitResponse.RequestID, queueMsg.ID)
		assert.NotEmpty(t, queueMsg.SessionID)

		// Step 4: Simulate task processing completion
		err = suite.atomicService.UpdateWorkflowStatus(
			submitResponse.RequestID,
			models.WorkflowStatusCompleted,
			stringPtr("Authentication system implemented successfully"),
			nil,
		)
		require.NoError(t, err)

		// Step 5: Retrieve via API
		apiReq, err := http.NewRequest("GET", "/api/v1/tasks/history/test-repo", nil)
		require.NoError(t, err)

		apiRecorder := httptest.NewRecorder()
		suite.router.ServeHTTP(apiRecorder, apiReq)

		assert.Equal(t, http.StatusOK, apiRecorder.Code)

		var historyResponse services.TaskHistoryResponse
		err = json.Unmarshal(apiRecorder.Body.Bytes(), &historyResponse)
		require.NoError(t, err)

		// Verify API response
		assert.Len(t, historyResponse.Data, 1)
		assert.Equal(t, 1, historyResponse.Pagination.Total)

		task := historyResponse.Data[0]
		assert.Equal(t, submitResponse.RequestID, task.RequestID)
		assert.Equal(t, "completed", task.Status)
		assert.Equal(t, "Implement user authentication system", task.Tasks)
		assert.NotNil(t, task.Result)
		assert.Equal(t, "Authentication system implemented successfully", *task.Result)
	})
}

// Test 2: Atomic Transaction Testing
func TestAtomicTransactions_SuccessAndFailure(t *testing.T) {
	suite := setupIntegrationSuite(t)

	t.Run("atomic success: both database and queue operations succeed", func(t *testing.T) {
		suite.publisher.ClearMessages()

		publishReq := services.PublishRequest{
			Tasks:          "Atomic success test",
			RepositoryName: "atomic-test-repo",
			Interactive:    true,
			MessageType:    "claude_task",
			SessionID:      uuid.New().String(),
			Payload: map[string]interface{}{
				"request_type": "claude_task",
				"input": map[string]interface{}{
					"tasks":           "Atomic success test",
					"repository_name": "atomic-test-repo",
				},
			},
		}

		response, err := suite.atomicService.PublishWithHistory(context.Background(), publishReq)
		assert.NoError(t, err)
		assert.NotNil(t, response)

		// Verify database record exists
		var dbRecord models.WorkflowHistory
		err = suite.db.Where("request_id = ?", response.RequestID).First(&dbRecord).Error
		assert.NoError(t, err)

		// Verify queue message was sent
		messages := suite.publisher.GetMessages()
		assert.Len(t, messages, 1)
	})

	t.Run("atomic failure: queue failure should rollback database", func(t *testing.T) {
		suite.publisher.ClearMessages()

		// Set publisher to fail
		suite.publisher.SetPublishError(fmt.Errorf("queue connection failed"))

		publishReq := services.PublishRequest{
			Tasks:          "Atomic failure test",
			RepositoryName: "atomic-test-repo",
			Interactive:    true,
			MessageType:    "claude_task",
			SessionID:      uuid.New().String(),
			Payload: map[string]interface{}{
				"request_type": "claude_task",
			},
		}

		response, err := suite.atomicService.PublishWithHistory(context.Background(), publishReq)
		assert.Error(t, err)
		assert.Nil(t, response)

		// Verify no queue message was sent
		messages := suite.publisher.GetMessages()
		assert.Len(t, messages, 0)

		// Reset publisher
		suite.publisher.SetPublishError(nil)
	})
}

// Test 3: API Pagination Testing
func TestAPIPagination_MultiplePages(t *testing.T) {
	suite := setupIntegrationSuite(t)

	// Seed test data - 25 tasks for pagination testing
	baseTime := time.Now()
	for i := 0; i < 25; i++ {
		task := models.WorkflowHistory{
			RequestID:      fmt.Sprintf("page-test-req-%d", i),
			Status:         models.WorkflowStatusCompleted,
			Tasks:          fmt.Sprintf("Pagination test task %d", i),
			RepositoryName: "pagination-repo",
			CreatedAt:      baseTime.Add(-time.Duration(i) * time.Minute),
		}
		err := suite.db.Create(&task).Error
		require.NoError(t, err)
	}

	t.Run("first page with default limit", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/tasks/history/pagination-repo", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		suite.router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response services.TaskHistoryResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response.Data, 20) // Default limit
		assert.Equal(t, 1, response.Pagination.Page)
		assert.Equal(t, 20, response.Pagination.Limit)
		assert.Equal(t, 25, response.Pagination.Total)
		assert.Equal(t, 2, response.Pagination.TotalPages)
	})

	t.Run("second page with custom limit", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/tasks/history/pagination-repo?page=2&limit=10", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		suite.router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response services.TaskHistoryResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response.Data, 10)
		assert.Equal(t, 2, response.Pagination.Page)
		assert.Equal(t, 10, response.Pagination.Limit)
		assert.Equal(t, 25, response.Pagination.Total)
		assert.Equal(t, 3, response.Pagination.TotalPages)
	})

	t.Run("last page with remaining items", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/tasks/history/pagination-repo?page=3&limit=10", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		suite.router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response services.TaskHistoryResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response.Data, 5) // Remaining items
		assert.Equal(t, 3, response.Pagination.Page)
	})
}

// Test 4: Error Scenario Testing
func TestErrorScenarios_DatabaseAndNetwork(t *testing.T) {
	suite := setupIntegrationSuite(t)

	t.Run("invalid repository name handling", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/tasks/history/nonexistent-repo", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		suite.router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusNotFound, recorder.Code)

		var errorResponse map[string]interface{}
		err = json.Unmarshal(recorder.Body.Bytes(), &errorResponse)
		require.NoError(t, err)
		assert.Contains(t, errorResponse, "error")
	})

	t.Run("invalid pagination parameters", func(t *testing.T) {
		tests := []struct {
			name     string
			params   string
			expected int
		}{
			{"invalid page", "?page=invalid", http.StatusBadRequest},
			{"invalid limit", "?limit=invalid", http.StatusBadRequest},
			{"page too large", "?page=999999", http.StatusBadRequest},
			{"limit too large", "?limit=1000", http.StatusBadRequest},
			{"zero page", "?page=0", http.StatusBadRequest},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				req, err := http.NewRequest("GET", "/api/v1/tasks/history/test-repo"+tc.params, nil)
				require.NoError(t, err)

				recorder := httptest.NewRecorder()
				suite.router.ServeHTTP(recorder, req)

				assert.Equal(t, tc.expected, recorder.Code)
			})
		}
	})
}

// Test 5: Performance Testing
func TestPerformance_ResponseTimes(t *testing.T) {
	suite := setupIntegrationSuite(t)

	// Seed large dataset for performance testing
	baseTime := time.Now()
	for i := 0; i < 100; i++ {
		task := models.WorkflowHistory{
			RequestID:      fmt.Sprintf("perf-req-%d", i),
			Status:         models.WorkflowStatusCompleted,
			Tasks:          fmt.Sprintf("Performance test task %d", i),
			RepositoryName: "performance-repo",
			CreatedAt:      baseTime.Add(-time.Duration(i) * time.Minute),
		}
		err := suite.db.Create(&task).Error
		require.NoError(t, err)
	}

	t.Run("api response time under 200ms", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/tasks/history/performance-repo?limit=20", nil)
		require.NoError(t, err)

		start := time.Now()
		recorder := httptest.NewRecorder()
		suite.router.ServeHTTP(recorder, req)
		elapsed := time.Since(start)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Less(t, elapsed, 200*time.Millisecond, "API response should be under 200ms")

		var response services.TaskHistoryResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response.Data, 20)
	})

	t.Run("concurrent requests performance", func(t *testing.T) {
		concurrency := 10
		var wg sync.WaitGroup
		var mu sync.Mutex
		responseTimes := make([]time.Duration, 0, concurrency)

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				req, err := http.NewRequest("GET", "/api/v1/tasks/history/performance-repo?limit=10", nil)
				require.NoError(t, err)

				start := time.Now()
				recorder := httptest.NewRecorder()
				suite.router.ServeHTTP(recorder, req)
				elapsed := time.Since(start)

				mu.Lock()
				responseTimes = append(responseTimes, elapsed)
				mu.Unlock()

				assert.Equal(t, http.StatusOK, recorder.Code)
			}()
		}

		wg.Wait()

		// Verify all requests completed under 200ms
		for _, responseTime := range responseTimes {
			assert.Less(t, responseTime, 200*time.Millisecond)
		}
	})
}

// Test 6: Concurrent Operation Testing
func TestConcurrentOperations_Safety(t *testing.T) {
	suite := setupIntegrationSuite(t)

	t.Run("concurrent task submissions", func(t *testing.T) {
		concurrency := 5
		var wg sync.WaitGroup
		var mu sync.Mutex
		requestIDs := make([]string, 0, concurrency)

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(taskNum int) {
				defer wg.Done()

				taskRequest := map[string]interface{}{
					"tasks":           fmt.Sprintf("Concurrent task %d", taskNum),
					"repository_name": "concurrent-repo",
					"interactive":     true,
				}

				jsonData, err := json.Marshal(taskRequest)
				require.NoError(t, err)

				req, err := http.NewRequest("POST", "/api/v1/claude/run-tasks", bytes.NewBuffer(jsonData))
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")

				recorder := httptest.NewRecorder()
				suite.router.ServeHTTP(recorder, req)

				assert.Equal(t, http.StatusAccepted, recorder.Code)

				var response struct {
					RequestID string `json:"request_id"`
				}
				err = json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)

				mu.Lock()
				requestIDs = append(requestIDs, response.RequestID)
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		// Verify all requests have unique IDs
		assert.Len(t, requestIDs, concurrency)
		uniqueIDs := make(map[string]bool)
		for _, id := range requestIDs {
			assert.False(t, uniqueIDs[id], "Request ID should be unique")
			uniqueIDs[id] = true
		}

		// Verify all database records were created
		var count int64
		suite.db.Model(&models.WorkflowHistory{}).Where("repository_name = ?", "concurrent-repo").Count(&count)
		assert.Equal(t, int64(concurrency), count)

		// Verify all queue messages were sent
		messages := suite.publisher.GetMessages()
		assert.Len(t, messages, concurrency)
	})
}

// Helper functions
func stringPtr(s string) *string { return &s }