package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JokerTrickster/workflow/local-backend/internal/domain"
	"github.com/JokerTrickster/workflow/local-backend/internal/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock Claude API server for testing
func createMockClaudeServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	
	// Mock the messages endpoint
	mux.HandleFunc("/v1/messages", func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and headers
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.NotEmpty(t, r.Header.Get("x-api-key"))
		assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))

		// Parse request body
		var requestBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		assert.NoError(t, err)

		// Verify required fields
		assert.Contains(t, requestBody, "model")
		assert.Contains(t, requestBody, "max_tokens")
		assert.Contains(t, requestBody, "messages")

		messages := requestBody["messages"].([]interface{})
		assert.NotEmpty(t, messages)

		// Extract the user message content for response simulation
		var userContent string
		for _, msg := range messages {
			msgMap := msg.(map[string]interface{})
			if msgMap["role"] == "user" {
				userContent = msgMap["content"].(string)
				break
			}
		}

		// Simulate different responses based on content
		var response map[string]interface{}
		
		if strings.Contains(userContent, "error") {
			// Simulate API error
			w.WriteHeader(http.StatusBadRequest)
			response = map[string]interface{}{
				"type":  "error",
				"error": map[string]interface{}{
					"type":    "invalid_request_error",
					"message": "Invalid request format",
				},
			}
		} else if strings.Contains(userContent, "timeout") {
			// Simulate timeout by delaying response
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusRequestTimeout)
			return
		} else if strings.Contains(userContent, "rate_limit") {
			// Simulate rate limiting
			w.WriteHeader(http.StatusTooManyRequests)
			response = map[string]interface{}{
				"type":  "error",
				"error": map[string]interface{}{
					"type":    "rate_limit_error",
					"message": "Rate limit exceeded",
				},
			}
		} else {
			// Simulate successful response
			w.WriteHeader(http.StatusOK)
			responseContent := generateMockResponse(userContent)
			response = map[string]interface{}{
				"id":      "msg_" + generateID(),
				"type":    "message",
				"role":    "assistant",
				"model":   requestBody["model"],
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": responseContent,
					},
				},
				"stop_reason": "end_turn",
				"usage": map[string]interface{}{
					"input_tokens":  50,
					"output_tokens": 100,
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	return httptest.NewServer(mux)
}

// generateMockResponse creates realistic responses based on input content
func generateMockResponse(userContent string) string {
	if strings.Contains(userContent, "func main") {
		return "This is a Go main function that serves as the entry point for the program. The code appears to be well-structured and follows Go conventions."
	}
	if strings.Contains(userContent, "analyze") {
		return "Based on my analysis, the code looks good. Here are some observations:\n\n1. The structure is clear\n2. The logic follows best practices\n3. No obvious issues detected"
	}
	if strings.Contains(userContent, "review") {
		return "Code review completed. The implementation is solid with good error handling and clean architecture patterns."
	}
	if strings.Contains(userContent, "optimize") {
		return "Here are some optimization suggestions:\n\n1. Consider using more efficient algorithms\n2. Add caching for frequently accessed data\n3. Implement connection pooling for database operations"
	}
	
	return "I've processed your request. The code analysis has been completed successfully."
}

// generateID creates a simple ID for mock responses
func generateID() string {
	return "test123456"
}

// setupClaudeService creates a Claude service configured to use the mock server
func setupClaudeService(t *testing.T, mockServer *httptest.Server) *infrastructure.ClaudeService {
	config := &infrastructure.ClaudeConfig{
		APIKey:    "sk-ant-test-key-1234567890",
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1000,
		Timeout:   30,
		BaseURL:   mockServer.URL, // Point to our mock server
	}

	logConfig := &infrastructure.LoggingConfig{
		Level:  "debug",
		Format: "text", 
		Output: "stdout",
	}
	logger := infrastructure.NewLogger(logConfig)

	service, err := infrastructure.NewClaudeService(config, logger)
	require.NoError(t, err)

	return service
}

func TestClaudeService_ProcessRequest(t *testing.T) {
	mockServer := createMockClaudeServer(t)
	defer mockServer.Close()

	claudeService := setupClaudeService(t, mockServer)
	ctx := context.Background()

	t.Run("successful code analysis request", func(t *testing.T) {
		message := domain.NewMessage("msg-analysis", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func main() { println(\"Hello, World!\") }",
			"task": "analyze this Go code",
		})

		response, err := claudeService.ProcessRequest(ctx, message)

		assert.NoError(t, err)
		assert.NotEmpty(t, response)
		assert.Contains(t, response, "Go main function")
		assert.Contains(t, response, "entry point")
	})

	t.Run("successful code review request", func(t *testing.T) {
		message := domain.NewMessage("msg-review", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func calculateSum(a, b int) int { return a + b }",
			"task": "review this code for best practices",
		})

		response, err := claudeService.ProcessRequest(ctx, message)

		assert.NoError(t, err)
		assert.NotEmpty(t, response)
		assert.Contains(t, response, "review")
		assert.Contains(t, response, "implementation")
	})

	t.Run("optimization request", func(t *testing.T) {
		message := domain.NewMessage("msg-optimize", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func processData() { /* some code */ }",
			"task": "optimize this function",
		})

		response, err := claudeService.ProcessRequest(ctx, message)

		assert.NoError(t, err)
		assert.NotEmpty(t, response)
		assert.Contains(t, response, "optimization")
		assert.Contains(t, response, "suggestions")
	})

	t.Run("API error handling", func(t *testing.T) {
		message := domain.NewMessage("msg-error", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "invalid code that triggers error",
			"task": "this should cause an error",
		})

		response, err := claudeService.ProcessRequest(ctx, message)

		assert.Error(t, err)
		assert.Empty(t, response)
		assert.Contains(t, err.Error(), "API request failed")
	})

	t.Run("rate limiting handling", func(t *testing.T) {
		message := domain.NewMessage("msg-rate-limit", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func test() { /* rate_limit trigger */ }",
			"task": "this should trigger rate limiting",
		})

		response, err := claudeService.ProcessRequest(ctx, message)

		assert.Error(t, err)
		assert.Empty(t, response)
		assert.Contains(t, err.Error(), "rate limit")
	})

	t.Run("context cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel() // Cancel immediately

		message := domain.NewMessage("msg-cancelled", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func test() {}",
			"task": "analyze",
		})

		response, err := claudeService.ProcessRequest(cancelCtx, message)

		assert.Error(t, err)
		assert.Empty(t, response)
		assert.Contains(t, err.Error(), "context")
	})

	t.Run("request timeout", func(t *testing.T) {
		timeoutCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
		defer cancel()

		message := domain.NewMessage("msg-timeout", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func test() { /* timeout trigger */ }",
			"task": "this should timeout",
		})

		response, err := claudeService.ProcessRequest(timeoutCtx, message)

		assert.Error(t, err)
		assert.Empty(t, response)
		// Error could be timeout or context deadline exceeded
		assert.True(t, strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "context deadline exceeded"))
	})
}

func TestClaudeService_ContextManagement(t *testing.T) {
	mockServer := createMockClaudeServer(t)
	defer mockServer.Close()

	claudeService := setupClaudeService(t, mockServer)
	ctx := context.Background()

	t.Run("conversation context handling", func(t *testing.T) {
		contextID := "ctx-conversation"

		// First message in conversation
		message1 := domain.NewMessage("msg-1", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func hello() { println(\"hello\") }",
			"task": "analyze this function",
		})
		message1.SetContextID(contextID)

		response1, err := claudeService.ProcessRequest(ctx, message1)
		assert.NoError(t, err)
		assert.NotEmpty(t, response1)

		// Second message in same conversation
		message2 := domain.NewMessage("msg-2", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func world() { println(\"world\") }",
			"task": "analyze this function too",
		})
		message2.SetContextID(contextID)

		response2, err := claudeService.ProcessRequest(ctx, message2)
		assert.NoError(t, err)
		assert.NotEmpty(t, response2)

		// Responses should both be successful
		assert.NotEqual(t, response1, response2)
	})

	t.Run("different contexts isolation", func(t *testing.T) {
		// Messages in different contexts should be handled independently
		message1 := domain.NewMessage("msg-ctx1", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func test1() {}",
			"task": "analyze",
		})
		message1.SetContextID("ctx-1")

		message2 := domain.NewMessage("msg-ctx2", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func test2() {}",
			"task": "analyze",
		})
		message2.SetContextID("ctx-2")

		response1, err1 := claudeService.ProcessRequest(ctx, message1)
		response2, err2 := claudeService.ProcessRequest(ctx, message2)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEmpty(t, response1)
		assert.NotEmpty(t, response2)
	})

	t.Run("message without context ID", func(t *testing.T) {
		message := domain.NewMessage("msg-no-ctx", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func standalone() {}",
			"task": "analyze",
		})
		// Deliberately not setting context ID

		response, err := claudeService.ProcessRequest(ctx, message)

		assert.NoError(t, err)
		assert.NotEmpty(t, response)
	})
}

func TestClaudeService_EdgeCases(t *testing.T) {
	mockServer := createMockClaudeServer(t)
	defer mockServer.Close()

	claudeService := setupClaudeService(t, mockServer)
	ctx := context.Background()

	t.Run("empty message payload", func(t *testing.T) {
		message := domain.NewMessage("msg-empty", domain.MessageTypeWorkRequest, map[string]interface{}{})

		response, err := claudeService.ProcessRequest(ctx, message)

		// Should handle empty payload gracefully
		assert.NoError(t, err)
		assert.NotEmpty(t, response)
	})

	t.Run("very large code input", func(t *testing.T) {
		// Create a large code string
		largeCode := strings.Repeat("func test() { println(\"large code\") }\n", 1000)
		
		message := domain.NewMessage("msg-large", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": largeCode,
			"task": "analyze this large codebase",
		})

		response, err := claudeService.ProcessRequest(ctx, message)

		assert.NoError(t, err)
		assert.NotEmpty(t, response)
	})

	t.Run("special characters in code", func(t *testing.T) {
		specialCode := `func test() {
			// Test with special chars: 中文, العربية, русский
			s := "quotes: \"double\" and 'single'"
			r := ` + "`" + `raw string with backticks` + "`" + `
			println(s, r)
		}`

		message := domain.NewMessage("msg-special", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": specialCode,
			"task": "analyze code with special characters",
		})

		response, err := claudeService.ProcessRequest(ctx, message)

		assert.NoError(t, err)
		assert.NotEmpty(t, response)
	})

	t.Run("multiple concurrent requests", func(t *testing.T) {
		const numRequests = 5
		responses := make(chan string, numRequests)
		errors := make(chan error, numRequests)

		// Send multiple requests concurrently
		for i := 0; i < numRequests; i++ {
			go func(index int) {
				message := domain.NewMessage(
					fmt.Sprintf("msg-concurrent-%d", index),
					domain.MessageTypeWorkRequest,
					map[string]interface{}{
						"code": fmt.Sprintf("func test%d() {}", index),
						"task": "analyze",
					},
				)

				response, err := claudeService.ProcessRequest(ctx, message)
				responses <- response
				errors <- err
			}(i)
		}

		// Collect all responses
		for i := 0; i < numRequests; i++ {
			err := <-errors
			response := <-responses
			
			assert.NoError(t, err, "Request %d should succeed", i)
			assert.NotEmpty(t, response, "Response %d should not be empty", i)
		}
	})
}

func TestClaudeService_RequestValidation(t *testing.T) {
	mockServer := createMockClaudeServer(t)
	defer mockServer.Close()

	claudeService := setupClaudeService(t, mockServer)
	ctx := context.Background()

	t.Run("validate message type", func(t *testing.T) {
		// Test with cancellation message (should not be processed)
		message := domain.NewMessage("msg-cancel", domain.MessageTypeCancellation, map[string]interface{}{
			"request_id": "some-request",
		})

		response, err := claudeService.ProcessRequest(ctx, message)

		// Depending on implementation, this might return an error or empty response
		// Adjust assertion based on your actual implementation
		if err != nil {
			assert.Contains(t, err.Error(), "message type")
		} else {
			assert.NotEmpty(t, response)
		}
	})

	t.Run("validate required fields", func(t *testing.T) {
		// Message with minimal valid content
		message := domain.NewMessage("msg-minimal", domain.MessageTypeWorkRequest, map[string]interface{}{
			"task": "minimal analysis request",
		})

		response, err := claudeService.ProcessRequest(ctx, message)

		assert.NoError(t, err)
		assert.NotEmpty(t, response)
	})
}

func TestClaudeService_Configuration(t *testing.T) {
	mockServer := createMockClaudeServer(t)
	defer mockServer.Close()

	t.Run("invalid API key", func(t *testing.T) {
		config := &infrastructure.ClaudeConfig{
			APIKey:    "", // Empty API key
			Model:     "claude-3-5-sonnet-20241022",
			MaxTokens: 1000,
			Timeout:   30,
			BaseURL:   mockServer.URL,
		}

		logConfig := &infrastructure.LoggingConfig{
			Level: "info", Format: "text", Output: "stdout",
		}
		logger := infrastructure.NewLogger(logConfig)

		service, err := infrastructure.NewClaudeService(config, logger)

		// Should fail to create service with invalid config
		assert.Error(t, err)
		assert.Nil(t, service)
	})

	t.Run("invalid model", func(t *testing.T) {
		config := &infrastructure.ClaudeConfig{
			APIKey:    "sk-ant-test-key",
			Model:     "", // Empty model
			MaxTokens: 1000,
			Timeout:   30,
			BaseURL:   mockServer.URL,
		}

		logConfig := &infrastructure.LoggingConfig{
			Level: "info", Format: "text", Output: "stdout",
		}
		logger := infrastructure.NewLogger(logConfig)

		service, err := infrastructure.NewClaudeService(config, logger)

		assert.Error(t, err)
		assert.Nil(t, service)
	})

	t.Run("invalid max tokens", func(t *testing.T) {
		config := &infrastructure.ClaudeConfig{
			APIKey:    "sk-ant-test-key",
			Model:     "claude-3-5-sonnet-20241022",
			MaxTokens: 0, // Invalid max tokens
			Timeout:   30,
			BaseURL:   mockServer.URL,
		}

		logConfig := &infrastructure.LoggingConfig{
			Level: "info", Format: "text", Output: "stdout",
		}
		logger := infrastructure.NewLogger(logConfig)

		service, err := infrastructure.NewClaudeService(config, logger)

		assert.Error(t, err)
		assert.Nil(t, service)
	})
}