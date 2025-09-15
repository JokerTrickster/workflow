package queue

import (
	"context"
	"testing"
	"time"

	"ai-git-workbench/internal/domain/entities"
	"ai-git-workbench/internal/domain/valueobjects"
	"ai-git-workbench/internal/infrastructure/config"
)

// TestRabbitMQRepository_Config tests configuration parsing
func TestRabbitMQRepository_Config(t *testing.T) {
	cfg := &config.RabbitMQConfig{
		Host:         "localhost",
		Port:         "5672",
		Username:     "guest",
		Password:     "guest",
		VHost:        "/",
		Queue:        "test-queue",
		Exchange:     "",
		RoutingKey:   "test-queue",
		Durable:      true,
		AutoDelete:   false,
		Exclusive:    false,
		NoWait:       false,
		ReconnectDelay: "5s",
		MaxRetries:   3,
	}

	// This test just verifies the repository can be created with valid config
	repo, err := NewRabbitMQRepository(cfg)
	if err != nil {
		t.Fatalf("Failed to create RabbitMQ repository: %v", err)
	}

	if repo == nil {
		t.Fatal("Repository should not be nil")
	}

	// Clean up
	if closer, ok := repo.(interface{ Close() error }); ok {
		closer.Close()
	}
}

// TestRabbitMQRepository_InvalidConfig tests error handling for invalid config
func TestRabbitMQRepository_InvalidConfig(t *testing.T) {
	_, err := NewRabbitMQRepository(nil)
	if err == nil {
		t.Fatal("Should return error for nil config")
	}

	expectedError := "RabbitMQ configuration is required"
	if err.Error() != expectedError {
		t.Fatalf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// TestRabbitMQRepository_Operations tests basic operations
func TestRabbitMQRepository_Operations(t *testing.T) {
	cfg := &config.RabbitMQConfig{
		Host:         "localhost", // Use localhost for testing
		Port:         "5672",
		Username:     "guest",
		Password:     "guest",
		VHost:        "/",
		Queue:        "test-queue",
		Exchange:     "",
		RoutingKey:   "test-queue",
		Durable:      false,       // Use non-durable for testing
		AutoDelete:   true,        // Auto-delete for testing
		Exclusive:    false,
		NoWait:       false,
		ReconnectDelay: "1s",
		MaxRetries:   2,
	}

	repo, err := NewRabbitMQRepository(cfg)
	if err != nil {
		t.Skipf("Skipping test - RabbitMQ not available: %v", err)
	}
	defer func() {
		if closer, ok := repo.(interface{ Close() error }); ok {
			closer.Close()
		}
	}()

	ctx := context.Background()

	// Test health check
	health, err := repo.GetQueueHealth(ctx)
	if err != nil {
		t.Logf("Health check failed (RabbitMQ may not be available): %v", err)
	} else {
		t.Logf("Queue health: %+v", health)
	}

	// Test statistics
	stats, err := repo.GetQueueStatistics(ctx)
	if err != nil {
		t.Errorf("GetQueueStatistics failed: %v", err)
	} else {
		t.Logf("Queue statistics: %+v", stats)
	}

	// Test unsupported operations (should not error, but return appropriate responses)
	tasks, err := repo.GetTasksInProgress(ctx)
	if err != nil {
		t.Errorf("GetTasksInProgress should not error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks in progress, got %d", len(tasks))
	}

	// Test queue length (may fail if RabbitMQ not connected)
	_, err = repo.GetQueueLength(ctx)
	if err != nil {
		t.Logf("GetQueueLength failed (expected if RabbitMQ not available): %v", err)
	}

	// Test is empty (may fail if RabbitMQ not connected)
	_, err = repo.IsEmpty(ctx)
	if err != nil {
		t.Logf("IsEmpty failed (expected if RabbitMQ not available): %v", err)
	}
}

// TestRabbitMQRepository_EnqueueWithMockTask tests enqueueing with a mock task
func TestRabbitMQRepository_EnqueueWithMockTask(t *testing.T) {
	cfg := &config.RabbitMQConfig{
		Host:         "localhost",
		Port:         "5672",
		Username:     "guest",
		Password:     "guest",
		VHost:        "/",
		Queue:        "test-queue",
		Exchange:     "",
		RoutingKey:   "test-queue",
		Durable:      false,
		AutoDelete:   true,
		Exclusive:    false,
		NoWait:       false,
		ReconnectDelay: "1s",
		MaxRetries:   2,
	}

	repo, err := NewRabbitMQRepository(cfg)
	if err != nil {
		t.Skipf("Skipping test - RabbitMQ not available: %v", err)
	}
	defer func() {
		if closer, ok := repo.(interface{ Close() error }); ok {
			closer.Close()
		}
	}()

	// Create a mock task
	taskID := valueobjects.GenerateTaskID()
	userID, _ := valueobjects.NewUserID("test-user")
	repoPath, _ := valueobjects.NewRepositoryPath("test/repo")
	branchName, _ := valueobjects.NewBranchName("main")

	task, err := entities.NewTask(
		taskID,
		userID,
		"Test Task",
		"Test Description",
		repoPath,
		"test-epic",
		branchName,
	)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	ctx := context.Background()

	// Test enqueue
	err = repo.Enqueue(ctx, task)
	if err != nil {
		t.Logf("Enqueue failed (expected if RabbitMQ not available): %v", err)
	} else {
		t.Log("Successfully enqueued task")
	}
}

// TestMessageSerialization tests message serialization
func TestMessageSerialization(t *testing.T) {
	taskID := valueobjects.GenerateTaskID()
	userID, _ := valueobjects.NewUserID("test-user")
	repoPath, _ := valueobjects.NewRepositoryPath("test/repo")
	branchName, _ := valueobjects.NewBranchName("main")

	task, err := entities.NewTask(
		taskID,
		userID,
		"Test Task",
		"Test Description",
		repoPath,
		"test-epic",
		branchName,
	)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	// Test message creation
	message := NewTaskCreatedMessage(task)
	if message == nil {
		t.Fatal("Message should not be nil")
	}

	if message.Type != MessageTypeTaskCreated {
		t.Errorf("Expected message type %s, got %s", MessageTypeTaskCreated, message.Type)
	}

	if message.Payload.TaskID != task.ID().Value() {
		t.Errorf("Expected task ID %s, got %s", task.ID().Value(), message.Payload.TaskID)
	}

	// Test JSON serialization
	jsonData, err := message.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize message to JSON: %v", err)
	}

	if len(jsonData) == 0 {
		t.Fatal("JSON data should not be empty")
	}

	// Test JSON deserialization
	deserializedMessage, err := FromJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to deserialize message from JSON: %v", err)
	}

	if deserializedMessage.Type != message.Type {
		t.Errorf("Expected message type %s, got %s", message.Type, deserializedMessage.Type)
	}

	if deserializedMessage.Payload.TaskID != message.Payload.TaskID {
		t.Errorf("Expected task ID %s, got %s", message.Payload.TaskID, deserializedMessage.Payload.TaskID)
	}
}

// TestMessageTypeValidation tests message type validation
func TestMessageTypeValidation(t *testing.T) {
	validTypes := []string{
		string(MessageTypeTaskCreated),
		string(MessageTypeTaskUpdated),
		string(MessageTypeTaskResumed),
	}

	for _, msgType := range validTypes {
		if !IsValidMessageType(msgType) {
			t.Errorf("Message type '%s' should be valid", msgType)
		}
	}

	invalidTypes := []string{
		"invalid_type",
		"",
		"task_deleted",
	}

	for _, msgType := range invalidTypes {
		if IsValidMessageType(msgType) {
			t.Errorf("Message type '%s' should be invalid", msgType)
		}
	}
}

// TestMessageRetryIncrement tests retry count increment
func TestMessageRetryIncrement(t *testing.T) {
	taskID := valueobjects.GenerateTaskID()
	userID, _ := valueobjects.NewUserID("test-user")
	repoPath, _ := valueobjects.NewRepositoryPath("test/repo")
	branchName, _ := valueobjects.NewBranchName("main")

	task, err := entities.NewTask(
		taskID,
		userID,
		"Test Task",
		"Test Description",
		repoPath,
		"test-epic",
		branchName,
	)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	message := NewTaskCreatedMessage(task)
	originalTimestamp := message.Timestamp

	if message.RetryCount != 0 {
		t.Errorf("Initial retry count should be 0, got %d", message.RetryCount)
	}

	// Wait a bit to ensure timestamp changes
	time.Sleep(time.Millisecond * 10)

	message.IncrementRetryCount()

	if message.RetryCount != 1 {
		t.Errorf("Retry count should be 1 after increment, got %d", message.RetryCount)
	}

	if !message.Timestamp.After(originalTimestamp) {
		t.Error("Timestamp should be updated after retry increment")
	}
}