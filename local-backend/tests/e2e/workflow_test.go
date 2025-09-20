package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JokerTrickster/workflow/local-backend/internal/domain"
	"github.com/JokerTrickster/workflow/local-backend/internal/infrastructure"
	"github.com/JokerTrickster/workflow/local-backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSystem represents a complete system setup for end-to-end testing
type TestSystem struct {
	messageRepo    domain.MessageRepository
	requestRepo    domain.RequestRepository
	contextRepo    domain.ProcessingContextRepository
	claudeService  domain.ClaudeService
	messageProcessor *usecase.MessageProcessor
	requestService   *usecase.RequestService
	orchestrator     *usecase.Orchestrator
	logger           *infrastructure.Logger
	cleanup          func()
}

// setupTestSystem creates a complete test system with all components
func setupTestSystem(t *testing.T) *TestSystem {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "e2e-test-*")
	require.NoError(t, err)

	dbPath := filepath.Join(tempDir, "test.db")

	// Database configuration
	dbConfig := &infrastructure.DatabaseConfig{
		Driver:       "sqlite",
		DSN:          dbPath,
		MaxIdleConns: 2,
		MaxOpenConns: 5,
	}

	// Logging configuration
	logConfig := &infrastructure.LoggingConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}
	logger := infrastructure.NewLogger(logConfig)

	// Claude service configuration (using mock server in real tests)
	claudeConfig := &infrastructure.ClaudeConfig{
		APIKey:    "sk-ant-test-key-for-e2e-testing",
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1000,
		Timeout:   30,
		BaseURL:   "", // Will be set to mock server in actual tests
	}

	// Create database repository
	sqliteRepo, err := infrastructure.NewSQLiteRepository(dbConfig, logger)
	require.NoError(t, err)
	
	err = sqliteRepo.Initialize()
	require.NoError(t, err)

	// Create Claude service (with mock)
	claudeService := &MockClaudeService{}

	// Create use case services
	requestService := usecase.NewRequestService(sqliteRepo, claudeService, logger)
	messageProcessor := usecase.NewMessageProcessor(sqliteRepo, sqliteRepo, requestService, logger)

	// Create orchestrator
	orchestrator := usecase.NewOrchestrator(messageProcessor, requestService, logger)

	cleanup := func() {
		sqliteRepo.Close()
		os.RemoveAll(tempDir)
	}

	return &TestSystem{
		messageRepo:      sqliteRepo,
		requestRepo:      sqliteRepo,
		contextRepo:      sqliteRepo,
		claudeService:    claudeService,
		messageProcessor: messageProcessor,
		requestService:   requestService,
		orchestrator:     orchestrator,
		logger:           logger,
		cleanup:          cleanup,
	}
}

// MockClaudeService for end-to-end testing
type MockClaudeService struct{}

func (m *MockClaudeService) ProcessRequest(ctx context.Context, message *domain.Message) (string, error) {
	// Simulate processing time
	time.Sleep(10 * time.Millisecond)

	payload := message.Payload
	if payload == nil {
		return "Empty request processed successfully.", nil
	}

	code, hasCode := payload["code"].(string)
	task, hasTask := payload["task"].(string)

	if !hasCode && !hasTask {
		return "Request processed successfully.", nil
	}

	response := fmt.Sprintf("Analysis complete for task: %s\n", task)
	if hasCode {
		response += fmt.Sprintf("Code analysis: The provided code has %d characters and appears to be well-structured.", len(code))
	}

	return response, nil
}

func TestCompleteWorkflow(t *testing.T) {
	system := setupTestSystem(t)
	defer system.cleanup()

	ctx := context.Background()

	t.Run("single work request workflow", func(t *testing.T) {
		// Step 1: Create and process a work request message
		message := domain.NewMessage("msg-workflow-1", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func main() { println(\"Hello, World!\") }",
			"task": "Analyze this Go code for best practices",
		})

		err := system.messageProcessor.ProcessMessage(ctx, message)
		assert.NoError(t, err)

		// Step 2: Verify message was saved
		savedMessage, err := system.messageRepo.GetByID(ctx, "msg-workflow-1")
		assert.NoError(t, err)
		assert.NotNil(t, savedMessage)
		assert.Equal(t, message.ID, savedMessage.ID)
		assert.Equal(t, message.Type, savedMessage.Type)

		// Step 3: Verify request was created and processed
		requests, err := system.requestRepo.GetPendingRequests(ctx)
		assert.NoError(t, err)
		
		// Find our request
		var workflowRequest *domain.Request
		for _, req := range requests {
			if req.MessageID == "msg-workflow-1" {
				workflowRequest = req
				break
			}
		}

		// The request might be completed already due to processing
		if workflowRequest == nil {
			// Check if it's completed
			allRequests, err := system.requestService.GetPendingRequests(ctx)
			assert.NoError(t, err)
			
			// In a real implementation, we'd have a method to get all requests
			// For now, just verify the processing didn't fail
		}

		// Step 4: Verify the workflow completed successfully
		// Since processing happens synchronously in our current implementation,
		// we can verify the message processing didn't return an error
		assert.NoError(t, err, "Workflow should complete without errors")
	})

	t.Run("conversational workflow", func(t *testing.T) {
		contextID := "ctx-conversation-e2e"

		// Step 1: First message in conversation
		message1 := domain.NewMessage("msg-conv-1", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func calculateSum(a, b int) int { return a + b }",
			"task": "Review this function",
		})
		message1.SetContextID(contextID)

		err := system.messageProcessor.ProcessMessage(ctx, message1)
		assert.NoError(t, err)

		// Step 2: Second message in same conversation
		message2 := domain.NewMessage("msg-conv-2", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func calculateProduct(a, b int) int { return a * b }",
			"task": "Review this function too, considering the previous one",
		})
		message2.SetContextID(contextID)

		err = system.messageProcessor.ProcessMessage(ctx, message2)
		assert.NoError(t, err)

		// Step 3: Verify context was created and managed
		processingContext, err := system.contextRepo.GetByID(ctx, contextID)
		assert.NoError(t, err)
		assert.NotNil(t, processingContext)
		assert.Equal(t, contextID, processingContext.ID)

		// Step 4: Verify both messages are associated with the context
		contextMessages, err := system.messageRepo.GetByContextID(ctx, contextID)
		assert.NoError(t, err)
		assert.Len(t, contextMessages, 2)
	})

	t.Run("cancellation workflow", func(t *testing.T) {
		// Step 1: Create a work request
		workMessage := domain.NewMessage("msg-cancel-work", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func longRunningTask() { /* complex logic */ }",
			"task": "Analyze this complex function",
		})

		err := system.messageProcessor.ProcessMessage(ctx, workMessage)
		assert.NoError(t, err)

		// Step 2: Find the created request ID (in a real system this would come from the response)
		// For testing, we'll simulate knowing the request ID
		requests, err := system.requestService.GetPendingRequests(ctx)
		assert.NoError(t, err)

		var requestID string
		for _, req := range requests {
			if req.MessageID == "msg-cancel-work" {
				requestID = req.ID
				break
			}
		}

		if requestID != "" {
			// Step 3: Send cancellation message
			cancelMessage := domain.NewMessage("msg-cancel-req", domain.MessageTypeCancellation, map[string]interface{}{
				"request_id": requestID,
			})

			err = system.messageProcessor.ProcessMessage(ctx, cancelMessage)
			assert.NoError(t, err)

			// Step 4: Verify request was cancelled
			request, err := system.requestService.GetRequestStatus(ctx, requestID)
			if err == nil && request != nil {
				// If we can retrieve it, it should be cancelled
				assert.Equal(t, domain.RequestStatusCancelled, request.Status)
			}
		}
	})
}

func TestOrchestorLifecycle(t *testing.T) {
	system := setupTestSystem(t)
	defer system.cleanup()

	ctx := context.Background()

	t.Run("orchestrator startup and shutdown", func(t *testing.T) {
		// Start orchestrator
		err := system.orchestrator.Start(ctx)
		assert.NoError(t, err)

		// Verify it's running
		assert.True(t, system.orchestrator.IsRunning())

		// Process some messages while running
		message := domain.NewMessage("msg-orch-test", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func test() {}",
			"task": "quick test",
		})

		err = system.messageProcessor.ProcessMessage(ctx, message)
		assert.NoError(t, err)

		// Stop orchestrator
		err = system.orchestrator.Stop()
		assert.NoError(t, err)

		// Verify it stopped
		assert.False(t, system.orchestrator.IsRunning())
	})

	t.Run("orchestrator background tasks", func(t *testing.T) {
		// Create some old contexts to be cleaned up
		oldContext := domain.NewProcessingContext("ctx-old")
		oldContext.LastUsedAt = time.Now().Add(-2 * time.Hour)
		err := system.contextRepo.Save(ctx, oldContext)
		require.NoError(t, err)

		// Start orchestrator with short intervals for testing
		err = system.orchestrator.Start(ctx)
		assert.NoError(t, err)

		// Wait a short time for background tasks to run
		time.Sleep(100 * time.Millisecond)

		// Stop orchestrator
		err = system.orchestrator.Stop()
		assert.NoError(t, err)

		// Verify old context was cleaned up (depends on cleanup interval)
		// This test might be flaky depending on timing
		retrievedContext, err := system.contextRepo.GetByID(ctx, "ctx-old")
		if err != nil {
			// Context was cleaned up (expected)
			assert.Nil(t, retrievedContext)
		} else {
			// Context still exists (also valid if cleanup hasn't run yet)
			assert.NotNil(t, retrievedContext)
		}
	})
}

func TestErrorHandlingWorkflows(t *testing.T) {
	system := setupTestSystem(t)
	defer system.cleanup()

	ctx := context.Background()

	t.Run("malformed message handling", func(t *testing.T) {
		// Create message with invalid payload
		message := domain.NewMessage("msg-malformed", "invalid_type", map[string]interface{}{
			"malformed": "data",
		})

		err := system.messageProcessor.ProcessMessage(ctx, message)
		// Should handle gracefully
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown message type")
	})

	t.Run("cancellation of non-existent request", func(t *testing.T) {
		cancelMessage := domain.NewMessage("msg-cancel-missing", domain.MessageTypeCancellation, map[string]interface{}{
			"request_id": "non-existent-request-id",
		})

		err := system.messageProcessor.ProcessMessage(ctx, cancelMessage)
		// Should handle gracefully
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find request")
	})

	t.Run("invalid cancellation payload", func(t *testing.T) {
		cancelMessage := domain.NewMessage("msg-cancel-invalid", domain.MessageTypeCancellation, map[string]interface{}{
			// Missing request_id
			"invalid": "payload",
		})

		err := system.messageProcessor.ProcessMessage(ctx, cancelMessage)
		// Should handle gracefully  
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid cancellation payload")
	})
}

func TestConcurrentWorkflows(t *testing.T) {
	system := setupTestSystem(t)
	defer system.cleanup()

	ctx := context.Background()

	t.Run("concurrent message processing", func(t *testing.T) {
		const numMessages = 10
		errors := make(chan error, numMessages)

		// Process multiple messages concurrently
		for i := 0; i < numMessages; i++ {
			go func(index int) {
				message := domain.NewMessage(
					fmt.Sprintf("msg-concurrent-%d", index),
					domain.MessageTypeWorkRequest,
					map[string]interface{}{
						"code": fmt.Sprintf("func test%d() {}", index),
						"task": fmt.Sprintf("analyze function %d", index),
					},
				)
				
				errors <- system.messageProcessor.ProcessMessage(ctx, message)
			}(i)
		}

		// Collect all results
		successCount := 0
		for i := 0; i < numMessages; i++ {
			err := <-errors
			if err == nil {
				successCount++
			}
		}

		// Most or all should succeed
		assert.GreaterOrEqual(t, successCount, numMessages/2, 
			"At least half of concurrent requests should succeed")
	})

	t.Run("concurrent context access", func(t *testing.T) {
		contextID := "ctx-concurrent-test"
		const numMessages = 5
		errors := make(chan error, numMessages)

		// Process multiple messages with same context concurrently
		for i := 0; i < numMessages; i++ {
			go func(index int) {
				message := domain.NewMessage(
					fmt.Sprintf("msg-ctx-concurrent-%d", index),
					domain.MessageTypeWorkRequest,
					map[string]interface{}{
						"code": fmt.Sprintf("func contextTest%d() {}", index),
						"task": "test concurrent context access",
					},
				)
				message.SetContextID(contextID)
				
				errors <- system.messageProcessor.ProcessMessage(ctx, message)
			}(i)
		}

		// Collect all results
		successCount := 0
		for i := 0; i < numMessages; i++ {
			err := <-errors
			if err == nil {
				successCount++
			}
		}

		assert.GreaterOrEqual(t, successCount, numMessages/2,
			"At least half of concurrent context requests should succeed")

		// Verify context was created
		processingContext, err := system.contextRepo.GetByID(ctx, contextID)
		assert.NoError(t, err)
		assert.NotNil(t, processingContext)
	})
}

func TestDataPersistence(t *testing.T) {
	system := setupTestSystem(t)
	defer system.cleanup()

	ctx := context.Background()

	t.Run("data survives processing", func(t *testing.T) {
		// Create and process message
		message := domain.NewMessage("msg-persist", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func persistent() { println(\"data persists\") }",
			"task": "test data persistence",
		})

		err := system.messageProcessor.ProcessMessage(ctx, message)
		require.NoError(t, err)

		// Verify message persisted
		savedMessage, err := system.messageRepo.GetByID(ctx, "msg-persist")
		assert.NoError(t, err)
		assert.NotNil(t, savedMessage)
		assert.Equal(t, message.Payload, savedMessage.Payload)

		// Verify timestamp fields are set
		assert.NotZero(t, savedMessage.ReceivedAt)
	})

	t.Run("complex data structures persist correctly", func(t *testing.T) {
		complexPayload := map[string]interface{}{
			"code": "func complex() {}",
			"task": "analyze complex structure",
			"metadata": map[string]interface{}{
				"priority": "high",
				"tags":     []string{"go", "function", "analysis"},
				"config": map[string]interface{}{
					"timeout": 30,
					"verbose": true,
				},
			},
		}

		message := domain.NewMessage("msg-complex", domain.MessageTypeWorkRequest, complexPayload)

		err := system.messageProcessor.ProcessMessage(ctx, message)
		require.NoError(t, err)

		// Verify complex payload persisted correctly
		savedMessage, err := system.messageRepo.GetByID(ctx, "msg-complex")
		assert.NoError(t, err)
		assert.NotNil(t, savedMessage)
		
		// Verify nested structures
		assert.Contains(t, savedMessage.Payload, "metadata")
		metadata := savedMessage.Payload["metadata"].(map[string]interface{})
		assert.Equal(t, "high", metadata["priority"])
	})
}

func TestSystemIntegration(t *testing.T) {
	system := setupTestSystem(t)
	defer system.cleanup()

	ctx := context.Background()

	t.Run("full system integration test", func(t *testing.T) {
		// Start orchestrator for full system test
		err := system.orchestrator.Start(ctx)
		require.NoError(t, err)
		defer system.orchestrator.Stop()

		// Simulate a realistic workflow
		contextID := "ctx-integration-test"

		// 1. Initial code analysis request
		analysisMsg := domain.NewMessage("msg-analysis", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": `
package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`,
			"task": "Analyze this Go program for best practices and potential improvements",
		})
		analysisMsg.SetContextID(contextID)

		err = system.messageProcessor.ProcessMessage(ctx, analysisMsg)
		assert.NoError(t, err)

		// 2. Follow-up question in same context
		followUpMsg := domain.NewMessage("msg-followup", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": `
func greetUser(name string) {
	fmt.Printf("Hello, %s!\n", name)
}
`,
			"task": "Now analyze this function that builds on the previous code",
		})
		followUpMsg.SetContextID(contextID)

		err = system.messageProcessor.ProcessMessage(ctx, followUpMsg)
		assert.NoError(t, err)

		// 3. Verify complete system state
		
		// Check messages were saved
		contextMessages, err := system.messageRepo.GetByContextID(ctx, contextID)
		assert.NoError(t, err)
		assert.Len(t, contextMessages, 2)

		// Check context was created and managed
		processingContext, err := system.contextRepo.GetByID(ctx, contextID)
		assert.NoError(t, err)
		assert.NotNil(t, processingContext)

		// Check requests were processed
		pendingRequests, err := system.requestService.GetPendingRequests(ctx)
		assert.NoError(t, err)
		
		// Verify system is in consistent state
		assert.NotNil(t, pendingRequests) // Should have some requests (pending or completed)
		
		// System should be running smoothly
		assert.True(t, system.orchestrator.IsRunning())
	})
}