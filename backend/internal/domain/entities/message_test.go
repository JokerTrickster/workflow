package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// MessageTestSuite defines the test suite for Message entity
type MessageTestSuite struct {
	suite.Suite
}

// TestMessageTestSuite runs the test suite
func TestMessageTestSuite(t *testing.T) {
	suite.Run(t, new(MessageTestSuite))
}

// TestNewMessage tests the NewMessage constructor
func (suite *MessageTestSuite) TestNewMessage() {
	sessionID := "session-123"
	messageType := MessageTypeWorkRequest
	role := MessageRoleUser
	content := "Please review this code"

	message := NewMessage(sessionID, messageType, role, content)

	assert := assert.New(suite.T())
	assert.NotEmpty(message.ID)
	assert.Equal(sessionID, message.SessionID)
	assert.Equal(messageType, message.Type)
	assert.Equal(role, message.Role)
	assert.Equal(content, message.Content)
	assert.NotNil(message.Metadata)
	assert.False(message.CreatedAt.IsZero())
	assert.False(message.UpdatedAt.IsZero())
	assert.True(message.IsValid())
}

// TestMessageAddMetadata tests the AddMetadata method
func (suite *MessageTestSuite) TestMessageAddMetadata() {
	message := NewMessage("session-123", MessageTypeWorkRequest, MessageRoleUser, "Test content")
	initialUpdatedAt := message.UpdatedAt
	
	// Wait to ensure time difference
	time.Sleep(1 * time.Millisecond)
	
	// Add metadata
	message.AddMetadata("priority", "high")
	message.AddMetadata("tokens", 150)
	message.AddMetadata("processing_time", 2.5)

	assert := assert.New(suite.T())
	assert.True(message.UpdatedAt.After(initialUpdatedAt))
	
	// Check metadata values
	priority, exists := message.GetMetadata("priority")
	assert.True(exists)
	assert.Equal("high", priority)
	
	tokens, exists := message.GetMetadata("tokens")
	assert.True(exists)
	assert.Equal(150, tokens)
	
	processingTime, exists := message.GetMetadata("processing_time")
	assert.True(exists)
	assert.Equal(2.5, processingTime)
}

// TestMessageAddMetadataToNilMetadata tests adding metadata when Metadata is nil
func (suite *MessageTestSuite) TestMessageAddMetadataToNilMetadata() {
	message := &Message{
		ID:        "test-id",
		SessionID: "session-123",
		Type:      MessageTypeWorkRequest,
		Role:      MessageRoleUser,
		Content:   "Test content",
		Metadata:  nil, // Explicitly set to nil
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	message.AddMetadata("test_key", "test_value")

	assert := assert.New(suite.T())
	assert.NotNil(message.Metadata)
	
	value, exists := message.GetMetadata("test_key")
	assert.True(exists)
	assert.Equal("test_value", value)
}

// TestMessageGetMetadata tests the GetMetadata method
func (suite *MessageTestSuite) TestMessageGetMetadata() {
	message := NewMessage("session-123", MessageTypeWorkRequest, MessageRoleUser, "Test content")

	assert := assert.New(suite.T())
	
	// Test non-existent key
	value, exists := message.GetMetadata("non_existent")
	assert.False(exists)
	assert.Nil(value)
	
	// Add and retrieve metadata
	message.AddMetadata("test_key", "test_value")
	value, exists = message.GetMetadata("test_key")
	assert.True(exists)
	assert.Equal("test_value", value)
}

// TestMessageGetMetadataFromNilMetadata tests getting metadata when Metadata is nil
func (suite *MessageTestSuite) TestMessageGetMetadataFromNilMetadata() {
	message := &Message{
		ID:        "test-id",
		SessionID: "session-123",
		Type:      MessageTypeWorkRequest,
		Role:      MessageRoleUser,
		Content:   "Test content",
		Metadata:  nil, // Explicitly set to nil
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	value, exists := message.GetMetadata("any_key")

	assert := assert.New(suite.T())
	assert.False(exists)
	assert.Nil(value)
}

// TestMessageIsValid tests the IsValid method
func (suite *MessageTestSuite) TestMessageIsValid() {
	assert := assert.New(suite.T())
	
	// Valid message
	validMessage := NewMessage("session-123", MessageTypeWorkRequest, MessageRoleUser, "Test content")
	assert.True(validMessage.IsValid())
	
	// Invalid messages
	invalidMessages := []*Message{
		{}, // Empty message
		{ID: "", SessionID: "session-123", Type: MessageTypeWorkRequest, Role: MessageRoleUser, Content: "Test", CreatedAt: time.Now()}, // Empty ID
		{ID: "msg-123", SessionID: "", Type: MessageTypeWorkRequest, Role: MessageRoleUser, Content: "Test", CreatedAt: time.Now()}, // Empty SessionID
		{ID: "msg-123", SessionID: "session-123", Type: "", Role: MessageRoleUser, Content: "Test", CreatedAt: time.Now()}, // Empty Type
		{ID: "msg-123", SessionID: "session-123", Type: MessageTypeWorkRequest, Role: "", Content: "Test", CreatedAt: time.Now()}, // Empty Role
		{ID: "msg-123", SessionID: "session-123", Type: MessageTypeWorkRequest, Role: MessageRoleUser, Content: "", CreatedAt: time.Now()}, // Empty Content
		{ID: "msg-123", SessionID: "session-123", Type: MessageTypeWorkRequest, Role: MessageRoleUser, Content: "Test", CreatedAt: time.Time{}}, // Zero CreatedAt
	}
	
	for i, invalidMessage := range invalidMessages {
		assert.False(invalidMessage.IsValid(), "Invalid message %d should be invalid", i)
	}
}

// TestMessageTypes tests all message types are valid
func (suite *MessageTestSuite) TestMessageTypes() {
	assert := assert.New(suite.T())
	
	types := []MessageType{
		MessageTypeWorkRequest,
		MessageTypeCancel,
		MessageTypeStatus,
	}
	
	for _, msgType := range types {
		message := NewMessage("session-123", msgType, MessageRoleUser, "Test content")
		assert.Equal(msgType, message.Type)
		assert.True(message.IsValid())
	}
}

// TestMessageRoles tests all message roles are valid
func (suite *MessageTestSuite) TestMessageRoles() {
	assert := assert.New(suite.T())
	
	roles := []MessageRole{
		MessageRoleUser,
		MessageRoleAssistant,
		MessageRoleSystem,
	}
	
	for _, role := range roles {
		message := NewMessage("session-123", MessageTypeWorkRequest, role, "Test content")
		assert.Equal(role, message.Role)
		assert.True(message.IsValid())
	}
}

// TestMessageMetadataTypes tests different metadata value types
func (suite *MessageTestSuite) TestMessageMetadataTypes() {
	message := NewMessage("session-123", MessageTypeWorkRequest, MessageRoleUser, "Test content")

	assert := assert.New(suite.T())
	
	// Test different value types
	testCases := map[string]interface{}{
		"string":  "test string",
		"int":     42,
		"float":   3.14,
		"bool":    true,
		"slice":   []string{"a", "b", "c"},
		"map":     map[string]string{"key": "value"},
		"nil":     nil,
	}
	
	for key, expectedValue := range testCases {
		message.AddMetadata(key, expectedValue)
		actualValue, exists := message.GetMetadata(key)
		assert.True(exists, "Key %s should exist", key)
		assert.Equal(expectedValue, actualValue, "Value for key %s should match", key)
	}
}

// TestMessageTimestamps tests timestamp behavior
func (suite *MessageTestSuite) TestMessageTimestamps() {
	before := time.Now()
	message := NewMessage("session-123", MessageTypeWorkRequest, MessageRoleUser, "Test content")
	after := time.Now()

	assert := assert.New(suite.T())
	
	// CreatedAt should be between before and after
	assert.True(message.CreatedAt.After(before) || message.CreatedAt.Equal(before))
	assert.True(message.CreatedAt.Before(after) || message.CreatedAt.Equal(after))
	
	// UpdatedAt should initially equal CreatedAt
	assert.Equal(message.CreatedAt, message.UpdatedAt)
	
	// Adding metadata should update UpdatedAt
	originalUpdatedAt := message.UpdatedAt
	time.Sleep(1 * time.Millisecond)
	message.AddMetadata("test", "value")
	assert.True(message.UpdatedAt.After(originalUpdatedAt))
}

// TestMessageConversationFlow tests a typical conversation flow
func (suite *MessageTestSuite) TestMessageConversationFlow() {
	assert := assert.New(suite.T())
	sessionID := "conversation-session"
	
	// User sends initial request
	userMessage := NewMessage(sessionID, MessageTypeWorkRequest, MessageRoleUser, "Please review this code: function test() { return true; }")
	userMessage.AddMetadata("code_language", "javascript")
	userMessage.AddMetadata("request_type", "code_review")
	
	assert.True(userMessage.IsValid())
	assert.Equal(MessageRoleUser, userMessage.Role)
	assert.Equal(MessageTypeWorkRequest, userMessage.Type)
	
	codeLanguage, exists := userMessage.GetMetadata("code_language")
	assert.True(exists)
	assert.Equal("javascript", codeLanguage)
	
	// System processes request
	systemMessage := NewMessage(sessionID, MessageTypeStatus, MessageRoleSystem, "Processing code review request")
	systemMessage.AddMetadata("status", "processing")
	systemMessage.AddMetadata("estimated_completion", time.Now().Add(30*time.Second))
	
	assert.True(systemMessage.IsValid())
	assert.Equal(MessageRoleSystem, systemMessage.Role)
	assert.Equal(MessageTypeStatus, systemMessage.Type)
	
	// Assistant responds
	assistantMessage := NewMessage(sessionID, MessageTypeWorkRequest, MessageRoleAssistant, "Code review complete. The function looks good!")
	assistantMessage.AddMetadata("review_score", 85)
	assistantMessage.AddMetadata("suggestions", []string{"Add JSDoc comments", "Consider error handling"})
	assistantMessage.AddMetadata("completion_time", time.Now())
	
	assert.True(assistantMessage.IsValid())
	assert.Equal(MessageRoleAssistant, assistantMessage.Role)
	
	reviewScore, exists := assistantMessage.GetMetadata("review_score")
	assert.True(exists)
	assert.Equal(85, reviewScore)
	
	suggestions, exists := assistantMessage.GetMetadata("suggestions")
	assert.True(exists)
	expectedSuggestions := []string{"Add JSDoc comments", "Consider error handling"}
	assert.Equal(expectedSuggestions, suggestions)
}

// BenchmarkNewMessage benchmarks message creation
func BenchmarkNewMessage(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewMessage("session-123", MessageTypeWorkRequest, MessageRoleUser, "Test content")
	}
}

// BenchmarkMessageAddMetadata benchmarks metadata addition
func BenchmarkMessageAddMetadata(b *testing.B) {
	message := NewMessage("session-123", MessageTypeWorkRequest, MessageRoleUser, "Test content")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		message.AddMetadata("key", "value")
	}
}

// BenchmarkMessageGetMetadata benchmarks metadata retrieval
func BenchmarkMessageGetMetadata(b *testing.B) {
	message := NewMessage("session-123", MessageTypeWorkRequest, MessageRoleUser, "Test content")
	message.AddMetadata("test_key", "test_value")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		message.GetMetadata("test_key")
	}
}