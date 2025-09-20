package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		msgType  MessageType
		payload  map[string]interface{}
		expected *Message
	}{
		{
			name:    "valid work request message",
			id:      "test-123",
			msgType: MessageTypeWorkRequest,
			payload: map[string]interface{}{
				"code": "func main() { println(\"hello\") }",
				"task": "analyze this code",
			},
			expected: &Message{
				ID:       "test-123",
				Type:     MessageTypeWorkRequest,
				Payload:  map[string]interface{}{"code": "func main() { println(\"hello\") }", "task": "analyze this code"},
				Metadata: map[string]string{},
			},
		},
		{
			name:    "valid cancellation message",
			id:      "cancel-456",
			msgType: MessageTypeCancellation,
			payload: map[string]interface{}{
				"request_id": "req-789",
			},
			expected: &Message{
				ID:       "cancel-456",
				Type:     MessageTypeCancellation,
				Payload:  map[string]interface{}{"request_id": "req-789"},
				Metadata: map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewMessage(tt.id, tt.msgType, tt.payload)
			
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.Payload, result.Payload)
			assert.NotNil(t, result.Metadata)
			assert.NotZero(t, result.ReceivedAt)
		})
	}
}

func TestMessage_SetAndGetContextID(t *testing.T) {
	message := NewMessage("test", MessageTypeWorkRequest, map[string]interface{}{})
	
	// Initially no context ID
	assert.Equal(t, "", message.GetContextID())
	
	// Set context ID
	contextID := "ctx-123"
	message.SetContextID(contextID)
	assert.Equal(t, contextID, message.GetContextID())
}

func TestMessage_IsWorkRequest(t *testing.T) {
	workMsg := NewMessage("test", MessageTypeWorkRequest, map[string]interface{}{})
	cancelMsg := NewMessage("test", MessageTypeCancellation, map[string]interface{}{})
	
	assert.True(t, workMsg.IsWorkRequest())
	assert.False(t, cancelMsg.IsWorkRequest())
}

func TestMessage_IsCancellation(t *testing.T) {
	workMsg := NewMessage("test", MessageTypeWorkRequest, map[string]interface{}{})
	cancelMsg := NewMessage("test", MessageTypeCancellation, map[string]interface{}{})
	
	assert.False(t, workMsg.IsCancellation())
	assert.True(t, cancelMsg.IsCancellation())
}

func TestNewRequest(t *testing.T) {
	id := "req-123"
	messageID := "msg-456"
	contextID := "ctx-789"
	requestData := `{"code": "test", "task": "analyze"}`
	
	request := NewRequest(id, messageID, contextID, requestData)
	
	assert.Equal(t, id, request.ID)
	assert.Equal(t, messageID, request.MessageID)
	assert.Equal(t, contextID, request.ContextID)
	assert.Equal(t, requestData, request.RequestData)
	assert.Equal(t, RequestStatusPending, request.Status)
	assert.Empty(t, request.Response)
	assert.Empty(t, request.ErrorMessage)
	assert.Nil(t, request.StartedAt)
	assert.Nil(t, request.CompletedAt)
	assert.NotZero(t, request.CreatedAt)
}

func TestRequest_Lifecycle(t *testing.T) {
	request := NewRequest("test", "msg", "ctx", "data")
	
	// Initially pending
	assert.Equal(t, RequestStatusPending, request.Status)
	assert.False(t, request.IsCompleted())
	assert.True(t, request.CanBeCancelled())
	
	// Start processing
	request.Start()
	assert.Equal(t, RequestStatusProcessing, request.Status)
	assert.NotNil(t, request.StartedAt)
	assert.False(t, request.IsCompleted())
	assert.True(t, request.CanBeCancelled())
	
	// Complete successfully
	response := "Analysis complete"
	request.Complete(response)
	assert.Equal(t, RequestStatusCompleted, request.Status)
	assert.Equal(t, response, request.Response)
	assert.NotNil(t, request.CompletedAt)
	assert.True(t, request.IsCompleted())
	assert.False(t, request.CanBeCancelled())
}

func TestRequest_FailureFlow(t *testing.T) {
	request := NewRequest("test", "msg", "ctx", "data")
	
	request.Start()
	
	errorMsg := "Processing failed"
	request.Fail(errorMsg)
	
	assert.Equal(t, RequestStatusFailed, request.Status)
	assert.Equal(t, errorMsg, request.ErrorMessage)
	assert.NotNil(t, request.CompletedAt)
	assert.True(t, request.IsCompleted())
	assert.False(t, request.CanBeCancelled())
}

func TestRequest_CancellationFlow(t *testing.T) {
	request := NewRequest("test", "msg", "ctx", "data")
	
	// Can cancel when pending
	assert.True(t, request.CanBeCancelled())
	request.Cancel()
	assert.Equal(t, RequestStatusCancelled, request.Status)
	assert.NotNil(t, request.CompletedAt)
	assert.True(t, request.IsCompleted())
	
	// Cannot cancel completed request
	newRequest := NewRequest("test2", "msg2", "ctx2", "data2")
	newRequest.Complete("done")
	assert.False(t, newRequest.CanBeCancelled())
}

func TestNewProcessingContext(t *testing.T) {
	id := "ctx-123"
	context := NewProcessingContext(id)
	
	assert.Equal(t, id, context.ID)
	assert.Empty(t, context.Messages)
	assert.NotNil(t, context.Metadata)
	assert.NotZero(t, context.CreatedAt)
	assert.NotZero(t, context.UpdatedAt)
	assert.NotZero(t, context.LastUsedAt)
}

func TestProcessingContext_AddMessages(t *testing.T) {
	context := NewProcessingContext("test")
	
	userMsg := "Hello, please analyze this code"
	assistantMsg := "I'll analyze that for you"
	
	context.AddUserMessage(userMsg)
	context.AddAssistantMessage(assistantMsg)
	
	messages := context.GetMessages()
	require.Len(t, messages, 2)
	
	assert.Equal(t, "user", messages[0].Role)
	assert.Equal(t, userMsg, messages[0].Content)
	assert.NotZero(t, messages[0].Timestamp)
	
	assert.Equal(t, "assistant", messages[1].Role)
	assert.Equal(t, assistantMsg, messages[1].Content)
	assert.NotZero(t, messages[1].Timestamp)
}

func TestProcessingContext_IsExpired(t *testing.T) {
	context := NewProcessingContext("test")
	
	// Fresh context should not be expired
	assert.False(t, context.IsExpired(1*time.Hour))
	
	// Manually set last used time to past
	context.LastUsedAt = time.Now().Add(-2 * time.Hour)
	assert.True(t, context.IsExpired(1*time.Hour))
	assert.False(t, context.IsExpired(3*time.Hour))
}

func TestProcessingContext_Metadata(t *testing.T) {
	context := NewProcessingContext("test")
	
	// Initially empty metadata
	value, exists := context.GetMetadata("key1")
	assert.False(t, exists)
	assert.Empty(t, value)
	
	// Set metadata
	context.SetMetadata("key1", "value1")
	context.SetMetadata("key2", "value2")
	
	value, exists = context.GetMetadata("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)
	
	value, exists = context.GetMetadata("key2")
	assert.True(t, exists)
	assert.Equal(t, "value2", value)
	
	value, exists = context.GetMetadata("nonexistent")
	assert.False(t, exists)
	assert.Empty(t, value)
}