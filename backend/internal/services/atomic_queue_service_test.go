package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"local-backend-server/internal/infrastructure/database/models"
	"local-backend-server/internal/infrastructure/queue"
)

// MockPublisher is a mock queue publisher for testing
type MockPublisher struct {
	messages      []*queue.WorkflowMessage
	shouldFail    bool
	failAfterCall int
	callCount     int
}

func (m *MockPublisher) PublishMessage(msg *queue.WorkflowMessage) error {
	m.callCount++
	if m.shouldFail && m.callCount >= m.failAfterCall {
		return &queue.PublishError{Message: "mock publish error"}
	}
	m.messages = append(m.messages, msg)
	return nil
}

func (m *MockPublisher) Close() error {
	return nil
}

func (m *MockPublisher) GetMessages() []*queue.WorkflowMessage {
	return m.messages
}

func (m *MockPublisher) Reset() {
	m.messages = nil
	m.callCount = 0
	m.shouldFail = false
	m.failAfterCall = 0
}

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the test models
	err = db.AutoMigrate(&models.WorkflowHistory{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestAtomicQueueService_PublishWithHistory_Success(t *testing.T) {
	// Setup
	db, err := setupTestDB()
	require.NoError(t, err)

	mockPublisher := &MockPublisher{}
	service := NewAtomicQueueService(db, mockPublisher)

	// Test data
	req := PublishRequest{
		Tasks:          "Test task content",
		RepositoryName: "test-repo",
		WorkingDir:     "/test/dir",
		Interactive:    true,
		ClaudeCmd:      "claude test",
		ContinueTask:   false,
		MessageType:    "claude_task",
		SessionID:      "session-123",
		Payload: map[string]interface{}{
			"request_type": "claude_task",
			"input": map[string]interface{}{
				"tasks": "Test task content",
			},
		},
	}

	// Execute
	ctx := context.Background()
	response, err := service.PublishWithHistory(ctx, req)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, response)

	// Check response
	assert.NotEmpty(t, response.RequestID)
	assert.Equal(t, models.WorkflowStatusPending, response.Status)
	assert.Equal(t, "Task queued successfully", response.Message)
	assert.WithinDuration(t, time.Now(), response.CreatedAt, time.Second)

	// Check database record
	var history models.WorkflowHistory
	err = db.Where("request_id = ?", response.RequestID).First(&history).Error
	require.NoError(t, err)

	assert.Equal(t, response.RequestID, history.RequestID)
	assert.Equal(t, models.WorkflowStatusPending, history.Status)
	assert.Equal(t, "Test task content", history.Tasks)
	assert.Equal(t, "test-repo", history.RepositoryName)
	assert.Equal(t, "/test/dir", *history.WorkingDir)
	assert.Equal(t, "claude test", *history.ClaudeCmd)
	assert.True(t, history.Interactive)
	assert.False(t, history.ContinueTask)

	// Check queue message
	messages := mockPublisher.GetMessages()
	require.Len(t, messages, 1)

	msg := messages[0]
	assert.Equal(t, "claude_task", msg.Type)
	assert.Equal(t, response.RequestID, msg.ID)
	assert.Equal(t, "session-123", msg.SessionID)
	assert.NotNil(t, msg.Payload)
}

func TestAtomicQueueService_PublishWithHistory_QueueFailure(t *testing.T) {
	// Setup
	db, err := setupTestDB()
	require.NoError(t, err)

	mockPublisher := &MockPublisher{
		shouldFail:    true,
		failAfterCall: 1,
	}
	service := NewAtomicQueueService(db, mockPublisher)

	// Test data
	req := PublishRequest{
		Tasks:          "Test task content",
		RepositoryName: "test-repo",
		MessageType:    "claude_task",
		SessionID:      "session-123",
		Payload:        map[string]interface{}{"test": "data"},
	}

	// Execute
	ctx := context.Background()
	response, err := service.PublishWithHistory(ctx, req)

	// Verify
	require.Error(t, err)
	require.Nil(t, response)
	assert.Contains(t, err.Error(), "atomic operation failed")
	assert.Contains(t, err.Error(), "failed to publish message to queue")

	// Check that no database record was created (transaction rolled back)
	var count int64
	db.Model(&models.WorkflowHistory{}).Count(&count)
	assert.Equal(t, int64(0), count)

	// Check that no queue message was sent
	messages := mockPublisher.GetMessages()
	assert.Len(t, messages, 0)
}

func TestAtomicQueueService_PublishWithHistory_NilPublisher(t *testing.T) {
	// Setup
	db, err := setupTestDB()
	require.NoError(t, err)

	service := NewAtomicQueueService(db, nil)

	// Test data
	req := PublishRequest{
		Tasks:          "Test task content",
		RepositoryName: "test-repo",
		MessageType:    "claude_task",
		SessionID:      "session-123",
		Payload:        map[string]interface{}{"test": "data"},
	}

	// Execute
	ctx := context.Background()
	response, err := service.PublishWithHistory(ctx, req)

	// Verify - should succeed even with nil publisher
	require.NoError(t, err)
	require.NotNil(t, response)

	// Check database record was created
	var history models.WorkflowHistory
	err = db.Where("request_id = ?", response.RequestID).First(&history).Error
	require.NoError(t, err)
	assert.Equal(t, "Test task content", history.Tasks)
}

func TestAtomicQueueService_GetWorkflowHistory(t *testing.T) {
	// Setup
	db, err := setupTestDB()
	require.NoError(t, err)

	service := NewAtomicQueueService(db, nil)

	// Create test data
	history := &models.WorkflowHistory{
		RequestID:      "test-request-123",
		Status:         models.WorkflowStatusPending,
		Tasks:          "Test task",
		RepositoryName: "test-repo",
		CreatedAt:      time.Now(),
	}
	err = db.Create(history).Error
	require.NoError(t, err)

	// Execute
	result, err := service.GetWorkflowHistory("test-request-123")

	// Verify
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "test-request-123", result.RequestID)
	assert.Equal(t, "Test task", result.Tasks)
	assert.Equal(t, "test-repo", result.RepositoryName)
}

func TestAtomicQueueService_UpdateWorkflowStatus(t *testing.T) {
	// Setup
	db, err := setupTestDB()
	require.NoError(t, err)

	service := NewAtomicQueueService(db, nil)

	// Create test data
	history := &models.WorkflowHistory{
		RequestID:      "test-request-123",
		Status:         models.WorkflowStatusPending,
		Tasks:          "Test task",
		RepositoryName: "test-repo",
		CreatedAt:      time.Now(),
	}
	err = db.Create(history).Error
	require.NoError(t, err)

	// Execute
	result := "Task completed successfully"
	err = service.UpdateWorkflowStatus("test-request-123", models.WorkflowStatusCompleted, &result, nil)

	// Verify
	require.NoError(t, err)

	// Check updated record
	var updated models.WorkflowHistory
	err = db.Where("request_id = ?", "test-request-123").First(&updated).Error
	require.NoError(t, err)

	assert.Equal(t, models.WorkflowStatusCompleted, updated.Status)
	assert.Equal(t, "Task completed successfully", *updated.Result)
	assert.NotNil(t, updated.CompletedAt)
	assert.Nil(t, updated.Error)
}