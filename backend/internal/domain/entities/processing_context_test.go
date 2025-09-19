package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ProcessingContextTestSuite defines the test suite for ProcessingContext entity
type ProcessingContextTestSuite struct {
	suite.Suite
}

// TestProcessingContextTestSuite runs the test suite
func TestProcessingContextTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessingContextTestSuite))
}

// TestNewProcessingContext tests the NewProcessingContext constructor
func (suite *ProcessingContextTestSuite) TestNewProcessingContext() {
	requestID := "request-123"
	sessionID := "session-456"
	before := time.Now()
	ctx := NewProcessingContext(requestID, sessionID)
	after := time.Now()

	assert := assert.New(suite.T())
	assert.NotEmpty(ctx.ID)
	assert.Equal(requestID, ctx.RequestID)
	assert.Equal(sessionID, ctx.SessionID)
	assert.NotNil(ctx.ConversationHistory)
	assert.Empty(ctx.ConversationHistory)
	assert.Empty(ctx.SystemPrompt)
	assert.NotNil(ctx.Metadata)
	assert.Empty(ctx.Metadata)
	assert.Nil(ctx.TokenUsage)
	assert.True(ctx.CreatedAt.After(before) || ctx.CreatedAt.Equal(before))
	assert.True(ctx.CreatedAt.Before(after) || ctx.CreatedAt.Equal(after))
	assert.Equal(ctx.CreatedAt, ctx.UpdatedAt)
	assert.True(ctx.IsValid())
	assert.Equal(0, ctx.GetMessageCount())
	assert.Nil(ctx.GetLatestMessage())
}

// TestProcessingContextAddMessage tests the AddMessage method
func (suite *ProcessingContextTestSuite) TestProcessingContextAddMessage() {
	ctx := NewProcessingContext("request-123", "session-456")
	initialUpdatedAt := ctx.UpdatedAt
	
	// Wait to ensure time difference
	time.Sleep(1 * time.Millisecond)
	
	// Create test messages
	message1 := NewMessage("session-456", MessageTypeWorkRequest, MessageRoleUser, "Hello, can you help me?")
	message2 := NewMessage("session-456", MessageTypeWorkRequest, MessageRoleAssistant, "Of course! How can I assist you?")
	message3 := NewMessage("session-456", MessageTypeStatus, MessageRoleSystem, "Processing request")

	// Add messages
	ctx.AddMessage(message1)
	ctx.AddMessage(message2)
	ctx.AddMessage(message3)

	assert := assert.New(suite.T())
	assert.True(ctx.UpdatedAt.After(initialUpdatedAt))
	assert.Equal(3, ctx.GetMessageCount())
	assert.Equal(3, len(ctx.ConversationHistory))
	
	// Check message order
	assert.Equal(message1, ctx.ConversationHistory[0])
	assert.Equal(message2, ctx.ConversationHistory[1])
	assert.Equal(message3, ctx.ConversationHistory[2])
	
	// Check latest message
	latestMessage := ctx.GetLatestMessage()
	assert.NotNil(latestMessage)
	assert.Equal(message3, latestMessage)
	assert.Equal("Processing request", latestMessage.Content)
}

// TestProcessingContextSetSystemPrompt tests the SetSystemPrompt method
func (suite *ProcessingContextTestSuite) TestProcessingContextSetSystemPrompt() {
	ctx := NewProcessingContext("request-123", "session-456")
	initialUpdatedAt := ctx.UpdatedAt
	
	// Wait to ensure time difference
	time.Sleep(1 * time.Millisecond)
	
	systemPrompt := "You are a helpful AI assistant specialized in code review and analysis."
	ctx.SetSystemPrompt(systemPrompt)

	assert := assert.New(suite.T())
	assert.Equal(systemPrompt, ctx.SystemPrompt)
	assert.True(ctx.UpdatedAt.After(initialUpdatedAt))
}

// TestProcessingContextAddMetadata tests the AddMetadata method
func (suite *ProcessingContextTestSuite) TestProcessingContextAddMetadata() {
	ctx := NewProcessingContext("request-123", "session-456")
	initialUpdatedAt := ctx.UpdatedAt
	
	// Wait to ensure time difference
	time.Sleep(1 * time.Millisecond)
	
	// Add various metadata types
	ctx.AddMetadata("model", "gpt-4")
	ctx.AddMetadata("temperature", 0.7)
	ctx.AddMetadata("max_tokens", 2048)
	ctx.AddMetadata("tools", []string{"code_analyzer", "file_reader"})
	ctx.AddMetadata("context_window", map[string]interface{}{"size": 8192, "used": 1024})

	assert := assert.New(suite.T())
	assert.True(ctx.UpdatedAt.After(initialUpdatedAt))
	
	// Check metadata values
	model, exists := ctx.Metadata["model"]
	assert.True(exists)
	assert.Equal("gpt-4", model)
	
	temperature, exists := ctx.Metadata["temperature"]
	assert.True(exists)
	assert.Equal(0.7, temperature)
	
	maxTokens, exists := ctx.Metadata["max_tokens"]
	assert.True(exists)
	assert.Equal(2048, maxTokens)
	
	tools, exists := ctx.Metadata["tools"]
	assert.True(exists)
	expectedTools := []string{"code_analyzer", "file_reader"}
	assert.Equal(expectedTools, tools)
	
	contextWindow, exists := ctx.Metadata["context_window"]
	assert.True(exists)
	expectedWindow := map[string]interface{}{"size": 8192, "used": 1024}
	assert.Equal(expectedWindow, contextWindow)
}

// TestProcessingContextAddMetadataToNilMetadata tests adding metadata when Metadata is nil
func (suite *ProcessingContextTestSuite) TestProcessingContextAddMetadataToNilMetadata() {
	ctx := &ProcessingContext{
		ID:                  "test-id",
		RequestID:           "request-123",
		SessionID:           "session-456",
		ConversationHistory: make([]*Message, 0),
		Metadata:            nil, // Explicitly set to nil
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	ctx.AddMetadata("test_key", "test_value")

	assert := assert.New(suite.T())
	assert.NotNil(ctx.Metadata)
	
	value, exists := ctx.Metadata["test_key"]
	assert.True(exists)
	assert.Equal("test_value", value)
}

// TestProcessingContextUpdateTokenUsage tests the UpdateTokenUsage method
func (suite *ProcessingContextTestSuite) TestProcessingContextUpdateTokenUsage() {
	ctx := NewProcessingContext("request-123", "session-456")
	initialUpdatedAt := ctx.UpdatedAt
	
	// Wait to ensure time difference
	time.Sleep(1 * time.Millisecond)
	
	// First update
	ctx.UpdateTokenUsage(100, 50)

	assert := assert.New(suite.T())
	assert.True(ctx.UpdatedAt.After(initialUpdatedAt))
	assert.NotNil(ctx.TokenUsage)
	assert.Equal(100, ctx.TokenUsage.InputTokens)
	assert.Equal(50, ctx.TokenUsage.OutputTokens)
	assert.Equal(150, ctx.TokenUsage.TotalTokens)
	
	// Second update (should accumulate)
	previousUpdatedAt := ctx.UpdatedAt
	time.Sleep(1 * time.Millisecond)
	ctx.UpdateTokenUsage(200, 75)
	
	assert.True(ctx.UpdatedAt.After(previousUpdatedAt))
	assert.Equal(300, ctx.TokenUsage.InputTokens)  // 100 + 200
	assert.Equal(125, ctx.TokenUsage.OutputTokens) // 50 + 75
	assert.Equal(425, ctx.TokenUsage.TotalTokens)  // 300 + 125
}

// TestProcessingContextUpdateTokenUsageFromNil tests updating token usage when TokenUsage is nil
func (suite *ProcessingContextTestSuite) TestProcessingContextUpdateTokenUsageFromNil() {
	ctx := &ProcessingContext{
		ID:                  "test-id",
		RequestID:           "request-123",
		SessionID:           "session-456",
		ConversationHistory: make([]*Message, 0),
		Metadata:            make(map[string]interface{}),
		TokenUsage:          nil, // Explicitly set to nil
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	ctx.UpdateTokenUsage(50, 25)

	assert := assert.New(suite.T())
	assert.NotNil(ctx.TokenUsage)
	assert.Equal(50, ctx.TokenUsage.InputTokens)
	assert.Equal(25, ctx.TokenUsage.OutputTokens)
	assert.Equal(75, ctx.TokenUsage.TotalTokens)
}

// TestProcessingContextGetMessageCount tests the GetMessageCount method
func (suite *ProcessingContextTestSuite) TestProcessingContextGetMessageCount() {
	ctx := NewProcessingContext("request-123", "session-456")
	assert := assert.New(suite.T())
	
	// Initially empty
	assert.Equal(0, ctx.GetMessageCount())
	
	// Add messages and check count
	for i := 1; i <= 5; i++ {
		message := NewMessage("session-456", MessageTypeWorkRequest, MessageRoleUser, "Message "+string(rune(i+'0')))
		ctx.AddMessage(message)
		assert.Equal(i, ctx.GetMessageCount())
	}
}

// TestProcessingContextGetLatestMessage tests the GetLatestMessage method
func (suite *ProcessingContextTestSuite) TestProcessingContextGetLatestMessage() {
	ctx := NewProcessingContext("request-123", "session-456")
	assert := assert.New(suite.T())
	
	// Initially nil
	assert.Nil(ctx.GetLatestMessage())
	
	// Add messages and check latest
	message1 := NewMessage("session-456", MessageTypeWorkRequest, MessageRoleUser, "First message")
	message2 := NewMessage("session-456", MessageTypeWorkRequest, MessageRoleAssistant, "Second message")
	message3 := NewMessage("session-456", MessageTypeStatus, MessageRoleSystem, "Third message")
	
	ctx.AddMessage(message1)
	assert.Equal(message1, ctx.GetLatestMessage())
	
	ctx.AddMessage(message2)
	assert.Equal(message2, ctx.GetLatestMessage())
	
	ctx.AddMessage(message3)
	assert.Equal(message3, ctx.GetLatestMessage())
}

// TestProcessingContextIsValid tests the IsValid method
func (suite *ProcessingContextTestSuite) TestProcessingContextIsValid() {
	assert := assert.New(suite.T())
	
	// Valid context
	validContext := NewProcessingContext("request-123", "session-456")
	assert.True(validContext.IsValid())
	
	// Invalid contexts
	invalidContexts := []*ProcessingContext{
		{}, // Empty context
		{ID: "", RequestID: "request-123", SessionID: "session-456", ConversationHistory: make([]*Message, 0), Metadata: make(map[string]interface{}), CreatedAt: time.Now()}, // Empty ID
		{ID: "ctx-123", RequestID: "", SessionID: "session-456", ConversationHistory: make([]*Message, 0), Metadata: make(map[string]interface{}), CreatedAt: time.Now()}, // Empty RequestID
		{ID: "ctx-123", RequestID: "request-123", SessionID: "", ConversationHistory: make([]*Message, 0), Metadata: make(map[string]interface{}), CreatedAt: time.Now()}, // Empty SessionID
		{ID: "ctx-123", RequestID: "request-123", SessionID: "session-456", ConversationHistory: make([]*Message, 0), Metadata: make(map[string]interface{}), CreatedAt: time.Time{}}, // Zero CreatedAt
	}
	
	for i, invalidContext := range invalidContexts {
		assert.False(invalidContext.IsValid(), "Invalid context %d should be invalid", i)
	}
}

// TestTokenUsageStruct tests the TokenUsage struct directly
func (suite *ProcessingContextTestSuite) TestTokenUsageStruct() {
	assert := assert.New(suite.T())
	
	// Test TokenUsage creation
	tokenUsage := &TokenUsage{
		InputTokens:  100,
		OutputTokens: 50,
		TotalTokens:  150,
	}
	
	assert.Equal(100, tokenUsage.InputTokens)
	assert.Equal(50, tokenUsage.OutputTokens)
	assert.Equal(150, tokenUsage.TotalTokens)
}

// TestProcessingContextCompleteWorkflow tests a complete processing workflow
func (suite *ProcessingContextTestSuite) TestProcessingContextCompleteWorkflow() {
	assert := assert.New(suite.T())
	requestID := "workflow-request-123"
	sessionID := "workflow-session-456"
	
	// Create processing context
	ctx := NewProcessingContext(requestID, sessionID)
	assert.Equal(0, ctx.GetMessageCount())
	
	// Set system prompt
	systemPrompt := "You are a code review assistant. Analyze code for bugs, performance issues, and best practices."
	ctx.SetSystemPrompt(systemPrompt)
	assert.Equal(systemPrompt, ctx.SystemPrompt)
	
	// Add configuration metadata
	ctx.AddMetadata("model", "gpt-4")
	ctx.AddMetadata("temperature", 0.3)
	ctx.AddMetadata("review_type", "security_focused")
	
	// Add conversation messages
	userMessage := NewMessage(sessionID, MessageTypeWorkRequest, MessageRoleUser, "Please review this function: func authenticate(token string) bool { return token == \"admin\" }")
	ctx.AddMessage(userMessage)
	
	systemMessage := NewMessage(sessionID, MessageTypeStatus, MessageRoleSystem, "Analyzing code for security vulnerabilities")
	ctx.AddMessage(systemMessage)
	
	assistantMessage := NewMessage(sessionID, MessageTypeWorkRequest, MessageRoleAssistant, "SECURITY ISSUE: Hardcoded credential detected. The function uses a hardcoded token comparison.")
	ctx.AddMessage(assistantMessage)
	
	// Update token usage
	ctx.UpdateTokenUsage(250, 180)
	
	// Verify final state
	assert.Equal(3, ctx.GetMessageCount())
	assert.Equal(assistantMessage, ctx.GetLatestMessage())
	assert.Equal("SECURITY ISSUE: Hardcoded credential detected. The function uses a hardcoded token comparison.", ctx.GetLatestMessage().Content)
	assert.Equal(250, ctx.TokenUsage.InputTokens)
	assert.Equal(180, ctx.TokenUsage.OutputTokens)
	assert.Equal(430, ctx.TokenUsage.TotalTokens)
	assert.True(ctx.IsValid())
	
	// Verify metadata
	model, exists := ctx.Metadata["model"]
	assert.True(exists)
	assert.Equal("gpt-4", model)
	
	reviewType, exists := ctx.Metadata["review_type"]
	assert.True(exists)
	assert.Equal("security_focused", reviewType)
}

// TestProcessingContextTimestamps tests timestamp behavior
func (suite *ProcessingContextTestSuite) TestProcessingContextTimestamps() {
	before := time.Now()
	ctx := NewProcessingContext("request-123", "session-456")
	after := time.Now()

	assert := assert.New(suite.T())
	
	// CreatedAt should be between before and after
	assert.True(ctx.CreatedAt.After(before) || ctx.CreatedAt.Equal(before))
	assert.True(ctx.CreatedAt.Before(after) || ctx.CreatedAt.Equal(after))
	
	// UpdatedAt should initially equal CreatedAt
	assert.Equal(ctx.CreatedAt, ctx.UpdatedAt)
	
	// Adding message should update UpdatedAt
	originalUpdatedAt := ctx.UpdatedAt
	time.Sleep(1 * time.Millisecond)
	message := NewMessage("session-456", MessageTypeWorkRequest, MessageRoleUser, "Test message")
	ctx.AddMessage(message)
	assert.True(ctx.UpdatedAt.After(originalUpdatedAt))
	
	// Setting system prompt should update UpdatedAt
	originalUpdatedAt = ctx.UpdatedAt
	time.Sleep(1 * time.Millisecond)
	ctx.SetSystemPrompt("Test prompt")
	assert.True(ctx.UpdatedAt.After(originalUpdatedAt))
	
	// Adding metadata should update UpdatedAt
	originalUpdatedAt = ctx.UpdatedAt
	time.Sleep(1 * time.Millisecond)
	ctx.AddMetadata("test", "value")
	assert.True(ctx.UpdatedAt.After(originalUpdatedAt))
	
	// Updating token usage should update UpdatedAt
	originalUpdatedAt = ctx.UpdatedAt
	time.Sleep(1 * time.Millisecond)
	ctx.UpdateTokenUsage(10, 5)
	assert.True(ctx.UpdatedAt.After(originalUpdatedAt))
}

// BenchmarkNewProcessingContext benchmarks context creation
func BenchmarkNewProcessingContext(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewProcessingContext("request-123", "session-456")
	}
}

// BenchmarkProcessingContextAddMessage benchmarks message addition
func BenchmarkProcessingContextAddMessage(b *testing.B) {
	ctx := NewProcessingContext("request-123", "session-456")
	message := NewMessage("session-456", MessageTypeWorkRequest, MessageRoleUser, "Benchmark message")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.AddMessage(message)
	}
}

// BenchmarkProcessingContextAddMetadata benchmarks metadata addition
func BenchmarkProcessingContextAddMetadata(b *testing.B) {
	ctx := NewProcessingContext("request-123", "session-456")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.AddMetadata("key", "value")
	}
}

// BenchmarkProcessingContextUpdateTokenUsage benchmarks token usage updates
func BenchmarkProcessingContextUpdateTokenUsage(b *testing.B) {
	ctx := NewProcessingContext("request-123", "session-456")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.UpdateTokenUsage(10, 5)
	}
}

// BenchmarkProcessingContextGetLatestMessage benchmarks getting latest message
func BenchmarkProcessingContextGetLatestMessage(b *testing.B) {
	ctx := NewProcessingContext("request-123", "session-456")
	// Add some messages for realistic testing
	for i := 0; i < 10; i++ {
		message := NewMessage("session-456", MessageTypeWorkRequest, MessageRoleUser, "Message")
		ctx.AddMessage(message)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.GetLatestMessage()
	}
}