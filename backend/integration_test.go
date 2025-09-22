package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"local-backend-server/internal/infrastructure/database"
	"local-backend-server/internal/infrastructure/database/models"
	"local-backend-server/internal/infrastructure/queue"
	"local-backend-server/internal/interfaces"
	"local-backend-server/internal/services"
)

// MockPublisher for integration tests
type MockPublisher struct {
	messages []*Message
}

type Message struct {
	Type      string                 `json:"type"`
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

func (m *MockPublisher) PublishMessage(msg *queue.WorkflowMessage) error {
	// Convert to our Message type for testing
	message := &Message{
		Type:      msg.Type,
		ID:        msg.ID,
		SessionID: msg.SessionID,
		Payload:   msg.Payload,
		Timestamp: msg.Timestamp,
	}
	m.messages = append(m.messages, message)
	return nil
}

func (m *MockPublisher) Close() error {
	return nil
}

func (m *MockPublisher) GetMessages() []*Message {
	return m.messages
}

func TestIntegration_ClaudeRunTasks_WithAtomicService(t *testing.T) {
	// Skip if running in CI without database
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate
	err = testDB.AutoMigrate(&models.WorkflowHistory{})
	require.NoError(t, err)

	// Setup mock publisher
	mockPublisher := &MockPublisher{}

	// Setup test environment
	originalDbConnection := dbConnection
	originalAtomicService := atomicService
	defer func() {
		dbConnection = originalDbConnection
		atomicService = originalAtomicService
	}()

	// Set test globals
	dbConnection = &database.DB{DB: testDB}
	var publisher interfaces.MessagePublisher = mockPublisher
	atomicService = services.NewAtomicQueueService(testDB, publisher)

	// Setup Gin in test mode
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/claude/run-tasks", handleClaudeRunTasks)

	// Test request
	reqBody := ReqRunTasksClaude{
		Tasks:          "Implement user authentication",
		RepositoryName: "test-repo",
		WorkingDir:     "/test/dir",
		Interactive:    true,
		ClaudeCmd:      "claude test",
		ContinueTask:   false,
	}

	jsonData, err := json.Marshal(reqBody)
	require.NoError(t, err)

	// Execute request
	req, err := http.NewRequest("POST", "/api/v1/claude/run-tasks", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusAccepted, w.Code)

	var response ClaudeTaskResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.RequestID)
	assert.Equal(t, "pending", response.Status)
	assert.Equal(t, "Task queued successfully", response.Message)
	assert.NotEmpty(t, response.CreatedAt)

	// Verify database record
	var history models.WorkflowHistory
	err = testDB.Where("request_id = ?", response.RequestID).First(&history).Error
	require.NoError(t, err)

	assert.Equal(t, response.RequestID, history.RequestID)
	assert.Equal(t, models.WorkflowStatusPending, history.Status)
	assert.Equal(t, "Implement user authentication", history.Tasks)
	assert.Equal(t, "test-repo", history.RepositoryName)
	assert.Equal(t, "/test/dir", *history.WorkingDir)
	assert.Equal(t, "claude test", *history.ClaudeCmd)
	assert.True(t, history.Interactive)
	assert.False(t, history.ContinueTask)

	// Verify queue message
	messages := mockPublisher.GetMessages()
	require.Len(t, messages, 1)

	msg := messages[0]
	assert.Equal(t, "claude_task", msg.Type)
	assert.Equal(t, response.RequestID, msg.ID)
	assert.NotEmpty(t, msg.SessionID)
	assert.NotNil(t, msg.Payload)

	// Verify payload structure
	assert.Equal(t, "claude_task", msg.Payload["request_type"])
	input, ok := msg.Payload["input"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Implement user authentication", input["tasks"])
	assert.Equal(t, "test-repo", input["repository_name"])
}

func TestIntegration_ClaudeRunTasks_Fallback(t *testing.T) {
	// Test fallback when atomic service is not available
	originalAtomicService := atomicService
	defer func() {
		atomicService = originalAtomicService
	}()

	// Disable atomic service
	atomicService = nil

	// Setup Gin in test mode
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/claude/run-tasks", handleClaudeRunTasks)

	// Test request
	reqBody := ReqRunTasksClaude{
		Tasks:          "Test fallback",
		RepositoryName: "test-repo",
	}

	jsonData, err := json.Marshal(reqBody)
	require.NoError(t, err)

	// Execute request
	req, err := http.NewRequest("POST", "/api/v1/claude/run-tasks", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response - should still work with fallback
	assert.Equal(t, http.StatusAccepted, w.Code)

	var response ClaudeTaskResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.RequestID)
	assert.Equal(t, "pending", response.Status)
	assert.Equal(t, "Claude task has been queued for processing", response.Message)
}