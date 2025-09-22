package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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
	"local-backend-server/internal/interfaces"
	"local-backend-server/internal/services"
)

// E2ETestSuite provides complete end-to-end testing environment
type E2ETestSuite struct {
	db            *gorm.DB
	router        *gin.Engine
	queueConsumer *MockQueueConsumer
	publisher     *MockPublisher
	atomicService *services.AtomicQueueService
}

// MockQueueConsumer simulates the local-backend queue consumer
type MockQueueConsumer struct {
	messages []*queue.WorkflowMessage
	db       *gorm.DB
}

func NewMockQueueConsumer(db *gorm.DB) *MockQueueConsumer {
	return &MockQueueConsumer{
		messages: make([]*queue.WorkflowMessage, 0),
		db:       db,
	}
}

func (m *MockQueueConsumer) ProcessMessage(msg *queue.WorkflowMessage) error {
	m.messages = append(m.messages, msg)

	// Simulate processing the task
	var workflow models.WorkflowHistory
	err := m.db.Where("request_id = ?", msg.ID).First(&workflow).Error
	if err != nil {
		return err
	}

	// Update status to processing
	workflow.Status = models.WorkflowStatusProcessing
	err = m.db.Save(&workflow).Error
	if err != nil {
		return err
	}

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// Complete the task
	workflow.Status = models.WorkflowStatusCompleted
	workflow.CompletedAt = timePtr(time.Now())
	workflow.ProcessingTimeMs = int64Ptr(100)
	workflow.Result = stringPtr("Task completed successfully by mock processor")

	return m.db.Save(&workflow).Error
}

func (m *MockQueueConsumer) GetProcessedMessages() []*queue.WorkflowMessage {
	return m.messages
}

// MockPublisher for queue publishing
type MockPublisher struct {
	messages     []*queue.WorkflowMessage
	consumer     *MockQueueConsumer
	publishError error
}

func NewMockPublisher(consumer *MockQueueConsumer) *MockPublisher {
	return &MockPublisher{
		messages: make([]*queue.WorkflowMessage, 0),
		consumer: consumer,
	}
}

func (m *MockPublisher) PublishMessage(msg *queue.WorkflowMessage) error {
	if m.publishError != nil {
		return m.publishError
	}

	m.messages = append(m.messages, msg)

	// Simulate immediate processing by consumer
	if m.consumer != nil {
		go func() {
			time.Sleep(50 * time.Millisecond) // Simulate queue delay
			m.consumer.ProcessMessage(msg)
		}()
	}

	return nil
}

func (m *MockPublisher) Close() error {
	return nil
}

func (m *MockPublisher) GetMessages() []*queue.WorkflowMessage {
	return m.messages
}

func (m *MockPublisher) SetPublishError(err error) {
	m.publishError = err
}

// setupE2ETestSuite creates complete testing environment
func setupE2ETestSuite(t *testing.T) *E2ETestSuite {
	// Setup database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.WorkflowHistory{})
	require.NoError(t, err)

	// Setup queue components
	consumer := NewMockQueueConsumer(db)
	publisher := NewMockPublisher(consumer)

	// Setup atomic service
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

	// Setup API routes
	api := router.Group("/api/v1")

	// Task history routes
	taskHistoryHandler := handlers.NewTaskHistoryHandler(db)
	tasks := api.Group("/tasks")
	tasks.GET("/history/:repository_name", taskHistoryHandler.GetTaskHistory)

	// Claude task submission
	router.POST("/api/v1/claude/run-tasks", func(c *gin.Context) {
		handleE2EClaudeRunTasks(c, atomicService)
	})

	return &E2ETestSuite{
		db:            db,
		router:        router,
		queueConsumer: consumer,
		publisher:     publisher,
		atomicService: atomicService,
	}
}

// Mock Claude task handler for E2E testing
func handleE2EClaudeRunTasks(c *gin.Context, atomicService *services.AtomicQueueService) {
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

	requestID := uuid.New().String()

	workflowData := services.WorkflowData{
		RequestID:      requestID,
		Tasks:          req.Tasks,
		RepositoryName: req.RepositoryName,
		WorkingDir:     req.WorkingDir,
		ClaudeCmd:      req.ClaudeCmd,
		Interactive:    req.Interactive,
		ContinueTask:   req.ContinueTask,
	}

	err := atomicService.CreateWorkflowAndQueue(context.Background(), workflowData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	response := struct {
		RequestID string `json:"request_id"`
		Status    string `json:"status"`
		Message   string `json:"message"`
		CreatedAt string `json:"created_at"`
	}{
		RequestID: requestID,
		Status:    "pending",
		Message:   "Task queued successfully",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusAccepted, response)
}

// TestE2E_CompleteWorkflow tests the entire system end-to-end
func TestE2E_CompleteWorkflow(t *testing.T) {
	suite := setupE2ETestSuite(t)

	t.Run("complete workflow: frontend submission -> queue -> processing -> api retrieval", func(t *testing.T) {
		// Step 1: Simulate frontend task submission
		taskRequest := map[string]interface{}{
			"tasks":           "Implement complete authentication system with JWT tokens",
			"repository_name": "e2e-test-repo",
			"working_dir":     "/path/to/e2e/repo",
			"interactive":     true,
			"claude_cmd":      "claude code --verbose",
			"continue_task":   false,
		}

		jsonData, err := json.Marshal(taskRequest)
		require.NoError(t, err)

		submitReq, err := http.NewRequest("POST", "/api/v1/claude/run-tasks", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		submitReq.Header.Set("Content-Type", "application/json")

		submitRecorder := httptest.NewRecorder()
		suite.router.ServeHTTP(submitRecorder, submitReq)

		// Verify submission response
		assert.Equal(t, http.StatusAccepted, submitRecorder.Code)

		var submitResponse struct {
			RequestID string `json:"request_id"`
			Status    string `json:"status"`
			Message   string `json:"message"`
			CreatedAt string `json:"created_at"`
		}
		err = json.Unmarshal(submitRecorder.Body.Bytes(), &submitResponse)
		require.NoError(t, err)

		assert.NotEmpty(t, submitResponse.RequestID)
		assert.Equal(t, "pending", submitResponse.Status)
		assert.Equal(t, "Task queued successfully", submitResponse.Message)

		// Step 2: Verify database record creation (atomic operation)
		var dbRecord models.WorkflowHistory
		err = suite.db.Where("request_id = ?", submitResponse.RequestID).First(&dbRecord).Error
		require.NoError(t, err)

		assert.Equal(t, submitResponse.RequestID, dbRecord.RequestID)
		assert.Equal(t, models.WorkflowStatusPending, dbRecord.Status)
		assert.Equal(t, "Implement complete authentication system with JWT tokens", dbRecord.Tasks)
		assert.Equal(t, "e2e-test-repo", dbRecord.RepositoryName)
		assert.Equal(t, "/path/to/e2e/repo", *dbRecord.WorkingDir)
		assert.Equal(t, "claude code --verbose", *dbRecord.ClaudeCmd)
		assert.True(t, dbRecord.Interactive)
		assert.False(t, dbRecord.ContinueTask)

		// Step 3: Verify queue message publication
		messages := suite.publisher.GetMessages()
		require.Len(t, messages, 1)

		queueMsg := messages[0]
		assert.Equal(t, "claude_task", queueMsg.Type)
		assert.Equal(t, submitResponse.RequestID, queueMsg.ID)
		assert.NotEmpty(t, queueMsg.SessionID)
		assert.NotNil(t, queueMsg.Payload)

		// Verify payload structure
		assert.Equal(t, "claude_task", queueMsg.Payload["request_type"])
		input, ok := queueMsg.Payload["input"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Implement complete authentication system with JWT tokens", input["tasks"])
		assert.Equal(t, "e2e-test-repo", input["repository_name"])

		// Step 4: Wait for queue processing to complete
		time.Sleep(300 * time.Millisecond)

		// Verify consumer processed the message
		processedMessages := suite.queueConsumer.GetProcessedMessages()
		require.Len(t, processedMessages, 1)
		assert.Equal(t, submitResponse.RequestID, processedMessages[0].ID)

		// Step 5: Verify database was updated after processing
		err = suite.db.Where("request_id = ?", submitResponse.RequestID).First(&dbRecord).Error
		require.NoError(t, err)

		assert.Equal(t, models.WorkflowStatusCompleted, dbRecord.Status)
		assert.NotNil(t, dbRecord.CompletedAt)
		assert.NotNil(t, dbRecord.ProcessingTimeMs)
		assert.NotNil(t, dbRecord.Result)
		assert.Equal(t, "Task completed successfully by mock processor", *dbRecord.Result)

		// Step 6: Test frontend API retrieval
		historyReq, err := http.NewRequest("GET", "/api/v1/tasks/history/e2e-test-repo", nil)
		require.NoError(t, err)

		historyRecorder := httptest.NewRecorder()
		suite.router.ServeHTTP(historyRecorder, historyReq)

		assert.Equal(t, http.StatusOK, historyRecorder.Code)

		var historyResponse services.TaskHistoryResponse
		err = json.Unmarshal(historyRecorder.Body.Bytes(), &historyResponse)
		require.NoError(t, err)

		// Verify API response structure
		assert.Len(t, historyResponse.Data, 1)
		assert.Equal(t, 1, historyResponse.Pagination.Total)
		assert.Equal(t, 1, historyResponse.Pagination.Page)
		assert.Equal(t, 1, historyResponse.Pagination.TotalPages)

		// Verify task data matches frontend expectations
		task := historyResponse.Data[0]
		assert.Equal(t, submitResponse.RequestID, task.RequestID)
		assert.Equal(t, "completed", task.Status)
		assert.Equal(t, "Implement complete authentication system with JWT tokens", task.Tasks)
		assert.Equal(t, "e2e-test-repo", task.RepositoryName)
		assert.Equal(t, "/path/to/e2e/repo", *task.WorkingDir)
		assert.Equal(t, "claude code --verbose", *task.ClaudeCmd)
		assert.True(t, task.Interactive)
		assert.False(t, task.ContinueTask)
		assert.NotNil(t, task.CompletedAt)
		assert.NotNil(t, task.ProcessingTimeMs)
		assert.NotNil(t, task.Result)
		assert.Equal(t, "Task completed successfully by mock processor", *task.Result)

		// Step 7: Verify processing time calculation
		createdAt, err := time.Parse(time.RFC3339, task.CreatedAt)
		require.NoError(t, err)

		completedAt, err := time.Parse(time.RFC3339, *task.CompletedAt)
		require.NoError(t, err)

		actualProcessingTime := completedAt.Sub(createdAt).Milliseconds()
		assert.GreaterOrEqual(t, actualProcessingTime, int64(100)) // At least our simulated processing time
		assert.LessOrEqual(t, actualProcessingTime, int64(1000))   // But reasonable upper bound
	})
}

// TestE2E_MultipleTasksWorkflow tests handling multiple concurrent tasks
func TestE2E_MultipleTasksWorkflow(t *testing.T) {
	suite := setupE2ETestSuite(t)

	t.Run("multiple tasks submitted and processed concurrently", func(t *testing.T) {
		numTasks := 3
		taskRequests := make([]map[string]interface{}, numTasks)
		submitResponses := make([]struct {
			RequestID string `json:"request_id"`
			Status    string `json:"status"`
		}, numTasks)

		// Submit multiple tasks
		for i := 0; i < numTasks; i++ {
			taskRequests[i] = map[string]interface{}{
				"tasks":           fmt.Sprintf("Multi-task workflow test %d", i),
				"repository_name": "multi-task-repo",
				"interactive":     i%2 == 0, // Alternate interactive mode
			}

			jsonData, err := json.Marshal(taskRequests[i])
			require.NoError(t, err)

			req, err := http.NewRequest("POST", "/api/v1/claude/run-tasks", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			suite.router.ServeHTTP(recorder, req)

			assert.Equal(t, http.StatusAccepted, recorder.Code)

			err = json.Unmarshal(recorder.Body.Bytes(), &submitResponses[i])
			require.NoError(t, err)
		}

		// Wait for all tasks to process
		time.Sleep(500 * time.Millisecond)

		// Verify all tasks were processed
		processedMessages := suite.queueConsumer.GetProcessedMessages()
		assert.Len(t, processedMessages, numTasks)

		// Retrieve all tasks via API
		req, err := http.NewRequest("GET", "/api/v1/tasks/history/multi-task-repo", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		suite.router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response services.TaskHistoryResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify all tasks are present and completed
		assert.Len(t, response.Data, numTasks)
		assert.Equal(t, numTasks, response.Pagination.Total)

		for i, task := range response.Data {
			assert.Equal(t, "completed", task.Status)
			assert.Contains(t, task.Tasks, "Multi-task workflow test")
			assert.Equal(t, "multi-task-repo", task.RepositoryName)
			assert.NotNil(t, task.Result)
		}

		// Verify tasks are sorted by created_at DESC (newest first)
		for i := 1; i < len(response.Data); i++ {
			prevTime, err := time.Parse(time.RFC3339, response.Data[i-1].CreatedAt)
			require.NoError(t, err)

			currTime, err := time.Parse(time.RFC3339, response.Data[i].CreatedAt)
			require.NoError(t, err)

			assert.True(t, prevTime.After(currTime) || prevTime.Equal(currTime),
				"Tasks should be sorted by created_at DESC")
		}
	})
}

// TestE2E_ErrorScenarios tests error handling throughout the workflow
func TestE2E_ErrorScenarios(t *testing.T) {
	suite := setupE2ETestSuite(t)

	t.Run("queue failure should prevent database record creation", func(t *testing.T) {
		// Setup publisher to fail
		suite.publisher.SetPublishError(fmt.Errorf("queue connection failed"))

		taskRequest := map[string]interface{}{
			"tasks":           "Task that should fail at queue",
			"repository_name": "error-test-repo",
		}

		jsonData, err := json.Marshal(taskRequest)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/v1/claude/run-tasks", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		suite.router.ServeHTTP(recorder, req)

		// Should return error
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)

		// Verify no database record was created (atomic rollback)
		var count int64
		suite.db.Model(&models.WorkflowHistory{}).Where("repository_name = ?", "error-test-repo").Count(&count)
		assert.Equal(t, int64(0), count)

		// Verify no messages were published
		messages := suite.publisher.GetMessages()
		assert.Len(t, messages, 0)

		// Reset publisher
		suite.publisher.SetPublishError(nil)
	})

	t.Run("repository not found returns proper error", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/tasks/history/nonexistent-repo", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		suite.router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusNotFound, recorder.Code)

		var errorResponse map[string]interface{}
		err = json.Unmarshal(recorder.Body.Bytes(), &errorResponse)
		require.NoError(t, err)

		assert.Contains(t, errorResponse, "error")
		assert.Contains(t, errorResponse, "request_id")
		assert.Contains(t, errorResponse, "timestamp")
	})
}

// TestE2E_PerformanceRequirements validates performance requirements
func TestE2E_PerformanceRequirements(t *testing.T) {
	suite := setupE2ETestSuite(t)

	t.Run("api response time meets requirements under load", func(t *testing.T) {
		// Create realistic dataset
		baseTime := time.Now()
		for i := 0; i < 50; i++ {
			task := models.WorkflowHistory{
				RequestID:        fmt.Sprintf("perf-req-%d", i),
				Status:          models.WorkflowStatusCompleted,
				Tasks:           fmt.Sprintf("Performance test task %d", i),
				RepositoryName:  "performance-repo",
				CreatedAt:       baseTime.Add(-time.Duration(i) * time.Minute),
				CompletedAt:     timePtr(baseTime.Add(-time.Duration(i-1) * time.Minute)),
				ProcessingTimeMs: int64Ptr(60000), // 1 minute
				Result:          stringPtr(fmt.Sprintf("Result for task %d", i)),
			}
			err := suite.db.Create(&task).Error
			require.NoError(t, err)
		}

		// Test API response time
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
		assert.Equal(t, 50, response.Pagination.Total)
	})
}

// Helper functions
func stringPtr(s string) *string    { return &s }
func timePtr(t time.Time) *time.Time { return &t }
func int64Ptr(i int64) *int64       { return &i }