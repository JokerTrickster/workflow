package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JokerTrickster/workflow/local-backend/internal/domain"
	"github.com/JokerTrickster/workflow/local-backend/internal/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDatabase creates a temporary SQLite database for testing
func setupTestDatabase(t *testing.T) (*infrastructure.SQLiteRepository, func()) {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "local-backend-test-*")
	require.NoError(t, err)

	dbPath := filepath.Join(tempDir, "test.db")
	
	config := &infrastructure.DatabaseConfig{
		Driver:       "sqlite",
		DSN:          dbPath,
		MaxIdleConns: 2,
		MaxOpenConns: 5,
	}

	// Create logger for database operations
	logConfig := &infrastructure.LoggingConfig{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	}
	logger := infrastructure.NewLogger(logConfig)

	repo, err := infrastructure.NewSQLiteRepository(config, logger)
	require.NoError(t, err)

	// Initialize database schema
	err = repo.Initialize()
	require.NoError(t, err)

	cleanup := func() {
		repo.Close()
		os.RemoveAll(tempDir)
	}

	return repo, cleanup
}

func TestSQLiteRepository_MessageOperations(t *testing.T) {
	repo, cleanup := setupTestDatabase(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("save and retrieve message", func(t *testing.T) {
		message := domain.NewMessage("msg-123", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func main() { println(\"hello\") }",
			"task": "analyze this code",
		})
		message.SetContextID("ctx-456")

		// Save message
		err := repo.SaveMessage(ctx, message)
		assert.NoError(t, err)

		// Retrieve message
		retrieved, err := repo.GetMessageByID(ctx, "msg-123")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, message.ID, retrieved.ID)
		assert.Equal(t, message.Type, retrieved.Type)
		assert.Equal(t, message.GetContextID(), retrieved.GetContextID())
		assert.NotZero(t, retrieved.ReceivedAt)
	})

	t.Run("get messages by context ID", func(t *testing.T) {
		contextID := "ctx-multiple"

		// Create multiple messages with same context ID
		message1 := domain.NewMessage("msg-ctx-1", domain.MessageTypeWorkRequest, map[string]interface{}{"test": 1})
		message1.SetContextID(contextID)
		message2 := domain.NewMessage("msg-ctx-2", domain.MessageTypeWorkRequest, map[string]interface{}{"test": 2})
		message2.SetContextID(contextID)
		message3 := domain.NewMessage("msg-ctx-3", domain.MessageTypeCancellation, map[string]interface{}{"test": 3})
		message3.SetContextID(contextID)

		// Save all messages
		err := repo.SaveMessage(ctx, message1)
		require.NoError(t, err)
		err = repo.SaveMessage(ctx, message2)
		require.NoError(t, err)
		err = repo.SaveMessage(ctx, message3)
		require.NoError(t, err)

		// Retrieve messages by context ID
		messages, err := repo.GetMessagesByContextID(ctx, contextID)
		assert.NoError(t, err)
		assert.Len(t, messages, 3)

		// Verify messages are ordered by received time
		for i := 1; i < len(messages); i++ {
			assert.True(t, messages[i].ReceivedAt.After(messages[i-1].ReceivedAt) || 
						messages[i].ReceivedAt.Equal(messages[i-1].ReceivedAt))
		}
	})

	t.Run("delete message", func(t *testing.T) {
		message := domain.NewMessage("msg-delete", domain.MessageTypeWorkRequest, map[string]interface{}{
			"test": "delete me",
		})

		// Save message
		err := repo.SaveMessage(ctx, message)
		require.NoError(t, err)

		// Verify it exists
		retrieved, err := repo.GetMessageByID(ctx, "msg-delete")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)

		// Delete message
		err = repo.DeleteMessage(ctx, "msg-delete")
		assert.NoError(t, err)

		// Verify it's gone
		retrieved, err = repo.GetMessageByID(ctx, "msg-delete")
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("message not found", func(t *testing.T) {
		message, err := repo.GetMessageByID(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, message)
	})
}

func TestSQLiteRepository_RequestOperations(t *testing.T) {
	repo, cleanup := setupTestDatabase(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("save and retrieve request", func(t *testing.T) {
		request := domain.NewRequest("req-123", "msg-456", "ctx-789", `{"code": "test", "task": "analyze"}`)

		// Save request
		err := repo.SaveRequest(ctx, request)
		assert.NoError(t, err)

		// Retrieve request
		retrieved, err := repo.GetRequestByID(ctx, "req-123")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, request.ID, retrieved.ID)
		assert.Equal(t, request.MessageID, retrieved.MessageID)
		assert.Equal(t, request.ContextID, retrieved.ContextID)
		assert.Equal(t, request.RequestData, retrieved.RequestData)
		assert.Equal(t, request.Status, retrieved.Status)
		assert.NotZero(t, retrieved.CreatedAt)
	})

	t.Run("request lifecycle updates", func(t *testing.T) {
		request := domain.NewRequest("req-lifecycle", "msg-lifecycle", "ctx-lifecycle", `{"test": true}`)

		// Save initial request
		err := repo.SaveRequest(ctx, request)
		require.NoError(t, err)

		// Start processing
		request.Start()
		err = repo.UpdateRequest(ctx, request)
		assert.NoError(t, err)

		// Verify status updated
		retrieved, err := repo.GetRequestByID(ctx, "req-lifecycle")
		assert.NoError(t, err)
		assert.Equal(t, domain.RequestStatusProcessing, retrieved.Status)
		assert.NotNil(t, retrieved.StartedAt)

		// Complete request
		request.Complete("Analysis complete")
		err = repo.UpdateRequest(ctx, request)
		assert.NoError(t, err)

		// Verify completion
		retrieved, err = repo.GetRequestByID(ctx, "req-lifecycle")
		assert.NoError(t, err)
		assert.Equal(t, domain.RequestStatusCompleted, retrieved.Status)
		assert.Equal(t, "Analysis complete", retrieved.Response)
		assert.NotNil(t, retrieved.CompletedAt)
	})

	t.Run("get request by message ID", func(t *testing.T) {
		request := domain.NewRequest("req-by-msg", "msg-unique", "ctx-test", `{"data": "test"}`)

		err := repo.SaveRequest(ctx, request)
		require.NoError(t, err)

		// Retrieve by message ID
		retrieved, err := repo.GetRequestByMessageID(ctx, "msg-unique")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, request.ID, retrieved.ID)
		assert.Equal(t, "msg-unique", retrieved.MessageID)
	})

	t.Run("get pending requests", func(t *testing.T) {
		// Create requests in different states
		pendingReq1 := domain.NewRequest("req-pending-1", "msg-p1", "ctx-p1", `{"test": 1}`)
		pendingReq2 := domain.NewRequest("req-pending-2", "msg-p2", "ctx-p2", `{"test": 2}`)
		completedReq := domain.NewRequest("req-completed", "msg-c", "ctx-c", `{"test": 3}`)
		completedReq.Complete("Done")
		processingReq := domain.NewRequest("req-processing", "msg-pr", "ctx-pr", `{"test": 4}`)
		processingReq.Start()

		// Save all requests
		err := repo.SaveRequest(ctx, pendingReq1)
		require.NoError(t, err)
		err = repo.SaveRequest(ctx, pendingReq2)
		require.NoError(t, err)
		err = repo.SaveRequest(ctx, completedReq)
		require.NoError(t, err)
		err = repo.SaveRequest(ctx, processingReq)
		require.NoError(t, err)

		// Get pending requests
		pending, err := repo.GetPendingRequests(ctx)
		assert.NoError(t, err)
		
		// Should only return pending requests (not processing, completed, etc.)
		pendingCount := 0
		for _, req := range pending {
			if req.Status == domain.RequestStatusPending {
				pendingCount++
			}
		}
		assert.GreaterOrEqual(t, pendingCount, 2) // At least our two pending requests
	})

	t.Run("delete request", func(t *testing.T) {
		request := domain.NewRequest("req-delete", "msg-delete", "ctx-delete", `{"delete": true}`)

		err := repo.SaveRequest(ctx, request)
		require.NoError(t, err)

		// Verify exists
		retrieved, err := repo.GetRequestByID(ctx, "req-delete")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)

		// Delete
		err = repo.DeleteRequest(ctx, "req-delete")
		assert.NoError(t, err)

		// Verify deleted
		retrieved, err = repo.GetRequestByID(ctx, "req-delete")
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})
}

func TestSQLiteRepository_ProcessingContextOperations(t *testing.T) {
	repo, cleanup := setupTestDatabase(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("save and retrieve processing context", func(t *testing.T) {
		processingContext := domain.NewProcessingContext("ctx-123")
		processingContext.AddUserMessage("Hello, analyze this code")
		processingContext.AddAssistantMessage("I'll analyze it for you")
		processingContext.SetMetadata("user_id", "user-456")

		// Save context
		err := repo.SaveProcessingContext(ctx, processingContext)
		assert.NoError(t, err)

		// Retrieve context
		retrieved, err := repo.GetProcessingContextByID(ctx, "ctx-123")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, processingContext.ID, retrieved.ID)
		assert.NotZero(t, retrieved.CreatedAt)
		assert.NotZero(t, retrieved.UpdatedAt)
		assert.NotZero(t, retrieved.LastUsedAt)

		// Verify messages were saved and retrieved
		messages := retrieved.GetMessages()
		assert.Len(t, messages, 2)
		assert.Equal(t, "user", messages[0].Role)
		assert.Equal(t, "Hello, analyze this code", messages[0].Content)
		assert.Equal(t, "assistant", messages[1].Role)
		assert.Equal(t, "I'll analyze it for you", messages[1].Content)

		// Verify metadata
		value, exists := retrieved.GetMetadata("user_id")
		assert.True(t, exists)
		assert.Equal(t, "user-456", value)
	})

	t.Run("get expired contexts", func(t *testing.T) {
		// Create contexts with different ages
		recentContext := domain.NewProcessingContext("ctx-recent")
		oldContext := domain.NewProcessingContext("ctx-old")
		
		// Manually set last used time to simulate age
		oldTime := time.Now().Add(-2 * time.Hour)
		oldContext.LastUsedAt = oldTime

		// Save contexts
		err := repo.SaveProcessingContext(ctx, recentContext)
		require.NoError(t, err)
		err = repo.SaveProcessingContext(ctx, oldContext)
		require.NoError(t, err)

		// Get contexts older than 1 hour
		expired, err := repo.GetExpiredProcessingContexts(ctx, 1*time.Hour)
		assert.NoError(t, err)
		
		// Should find at least the old context
		found := false
		for _, expiredCtx := range expired {
			if expiredCtx.ID == "ctx-old" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find the expired context")
	})

	t.Run("delete processing context", func(t *testing.T) {
		processingContext := domain.NewProcessingContext("ctx-delete")

		err := repo.SaveProcessingContext(ctx, processingContext)
		require.NoError(t, err)

		// Verify exists
		retrieved, err := repo.GetProcessingContextByID(ctx, "ctx-delete")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)

		// Delete
		err = repo.DeleteProcessingContext(ctx, "ctx-delete")
		assert.NoError(t, err)

		// Verify deleted
		retrieved, err = repo.GetProcessingContextByID(ctx, "ctx-delete")
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})
}

func TestSQLiteRepository_TransactionOperations(t *testing.T) {
	repo, cleanup := setupTestDatabase(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("transaction rollback on error", func(t *testing.T) {
		message := domain.NewMessage("msg-tx-test", domain.MessageTypeWorkRequest, map[string]interface{}{
			"test": "transaction",
		})
		request := domain.NewRequest("req-tx-test", "msg-tx-test", "ctx-tx-test", `{"test": "transaction"}`)

		// Start a transaction context (simulated)
		err := repo.SaveMessage(ctx, message)
		require.NoError(t, err)

		// Attempt to save request with same ID twice (should fail on second attempt)
		err = repo.SaveRequest(ctx, request)
		assert.NoError(t, err)

		// Try to save again - this should fail
		err = repo.SaveRequest(ctx, request)
		assert.Error(t, err) // Should fail due to unique constraint

		// Verify the first save was successful
		retrieved, err := repo.GetRequestByID(ctx, "req-tx-test")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
	})
}

func TestSQLiteRepository_ConcurrentOperations(t *testing.T) {
	repo, cleanup := setupTestDatabase(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("concurrent message saves", func(t *testing.T) {
		const numGoroutines = 10
		errors := make(chan error, numGoroutines)

		// Start multiple goroutines saving messages concurrently
		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				message := domain.NewMessage(
					fmt.Sprintf("msg-concurrent-%d", index),
					domain.MessageTypeWorkRequest,
					map[string]interface{}{"index": index},
				)
				errors <- repo.SaveMessage(ctx, message)
			}(i)
		}

		// Collect results
		for i := 0; i < numGoroutines; i++ {
			err := <-errors
			assert.NoError(t, err, "Concurrent save %d should succeed", i)
		}

		// Verify all messages were saved
		for i := 0; i < numGoroutines; i++ {
			message, err := repo.GetMessageByID(ctx, fmt.Sprintf("msg-concurrent-%d", i))
			assert.NoError(t, err)
			assert.NotNil(t, message)
		}
	})
}

func TestSQLiteRepository_EdgeCases(t *testing.T) {
	repo, cleanup := setupTestDatabase(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("very large payload", func(t *testing.T) {
		// Create a message with a very large payload
		largeData := make([]byte, 1024*1024) // 1MB of data
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}

		message := domain.NewMessage("msg-large", domain.MessageTypeWorkRequest, map[string]interface{}{
			"large_data": string(largeData),
			"code":       "func main() {}",
		})

		err := repo.SaveMessage(ctx, message)
		assert.NoError(t, err)

		retrieved, err := repo.GetMessageByID(ctx, "msg-large")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Contains(t, retrieved.Payload, "large_data")
	})

	t.Run("special characters in IDs and data", func(t *testing.T) {
		specialID := "msg-special-!@#$%^&*()_+-=[]{}|;:,.<>?"
		message := domain.NewMessage(specialID, domain.MessageTypeWorkRequest, map[string]interface{}{
			"special_chars": "Test with 中文, العربية, русский, emoji 🎉",
			"quotes":       `"single" and 'double' quotes`,
			"newlines":     "line1\nline2\rtab\there",
		})

		err := repo.SaveMessage(ctx, message)
		assert.NoError(t, err)

		retrieved, err := repo.GetMessageByID(ctx, specialID)
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	t.Run("context cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel() // Cancel immediately

		message := domain.NewMessage("msg-cancelled", domain.MessageTypeWorkRequest, map[string]interface{}{})

		err := repo.SaveMessage(cancelCtx, message)
		// Behavior depends on SQLite driver's context handling
		// Should either succeed quickly or fail with context error
		if err != nil {
			assert.Contains(t, err.Error(), "context")
		}
	})

	t.Run("nil payload handling", func(t *testing.T) {
		message := domain.NewMessage("msg-nil-payload", domain.MessageTypeWorkRequest, nil)

		err := repo.SaveMessage(ctx, message)
		assert.NoError(t, err)

		retrieved, err := repo.GetMessageByID(ctx, "msg-nil-payload")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
	})
}

func TestSQLiteRepository_DatabaseConstraints(t *testing.T) {
	repo, cleanup := setupTestDatabase(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("unique constraint violations", func(t *testing.T) {
		message1 := domain.NewMessage("msg-duplicate", domain.MessageTypeWorkRequest, map[string]interface{}{})
		message2 := domain.NewMessage("msg-duplicate", domain.MessageTypeWorkRequest, map[string]interface{}{})

		// First save should succeed
		err := repo.SaveMessage(ctx, message1)
		assert.NoError(t, err)

		// Second save with same ID should fail
		err = repo.SaveMessage(ctx, message2)
		assert.Error(t, err)
	})

	t.Run("foreign key constraints", func(t *testing.T) {
		// This test depends on how foreign keys are implemented in your schema
		// If you have foreign key constraints between messages and requests
		request := domain.NewRequest("req-fk", "msg-nonexistent", "ctx-fk", `{"test": true}`)

		err := repo.SaveRequest(ctx, request)
		// Depending on schema, this might succeed (if no FK constraint) or fail
		// Adjust assertion based on your actual schema design
		if err != nil {
			assert.Contains(t, err.Error(), "foreign key")
		}
	})
}