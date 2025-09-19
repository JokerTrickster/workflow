package services

import (
	"context"
	"errors"
	"local-backend-server/internal/domain/entities"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockMessageRepository is a mock implementation of MessageRepository
type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Create(ctx context.Context, message *entities.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageRepository) GetByID(ctx context.Context, id string) (*entities.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Message), args.Error(1)
}

func (m *MockMessageRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*entities.Message, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]*entities.Message), args.Error(1)
}

func (m *MockMessageRepository) GetBySessionIDWithPagination(ctx context.Context, sessionID string, offset, limit int) ([]*entities.Message, error) {
	args := m.Called(ctx, sessionID, offset, limit)
	return args.Get(0).([]*entities.Message), args.Error(1)
}

func (m *MockMessageRepository) Update(ctx context.Context, message *entities.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMessageRepository) GetByTypeAndSessionID(ctx context.Context, messageType entities.MessageType, sessionID string) ([]*entities.Message, error) {
	args := m.Called(ctx, messageType, sessionID)
	return args.Get(0).([]*entities.Message), args.Error(1)
}

func (m *MockMessageRepository) CountBySessionID(ctx context.Context, sessionID string) (int, error) {
	args := m.Called(ctx, sessionID)
	return args.Int(0), args.Error(1)
}

func (m *MockMessageRepository) GetLatestBySessionID(ctx context.Context, sessionID string) (*entities.Message, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Message), args.Error(1)
}

// MessageServiceTestSuite defines the test suite for MessageService
type MessageServiceTestSuite struct {
	suite.Suite
	service         *MessageService
	mockMessageRepo *MockMessageRepository
	mockSessionRepo *MockSessionRepository
	ctx             context.Context
	activeSession   *entities.Session
}

// SetupTest runs before each test
func (suite *MessageServiceTestSuite) SetupTest() {
	suite.mockMessageRepo = &MockMessageRepository{}
	suite.mockSessionRepo = &MockSessionRepository{}
	suite.service = NewMessageService(suite.mockMessageRepo, suite.mockSessionRepo)
	suite.ctx = context.Background()
	
	// Create a valid active session for testing
	suite.activeSession = entities.NewSession("user-123")
}

// TestMessageServiceTestSuite runs the test suite
func TestMessageServiceTestSuite(t *testing.T) {
	suite.Run(t, new(MessageServiceTestSuite))
}

// TestNewMessageService tests the NewMessageService constructor
func (suite *MessageServiceTestSuite) TestNewMessageService() {
	service := NewMessageService(suite.mockMessageRepo, suite.mockSessionRepo)
	
	assert := assert.New(suite.T())
	assert.NotNil(service)
	assert.Equal(suite.mockMessageRepo, service.messageRepo)
	assert.Equal(suite.mockSessionRepo, service.sessionRepo)
}

// TestCreateMessage_Success tests successful message creation
func (suite *MessageServiceTestSuite) TestCreateMessage_Success() {
	sessionID := "session-123"
	messageType := entities.MessageTypeWorkRequest
	role := entities.MessageRoleUser
	content := "Please review this code: function test() { return true; }"
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(suite.activeSession, nil)
	suite.mockMessageRepo.On("Create", suite.ctx, mock.AnythingOfType("*entities.Message")).Return(nil)

	// Execute
	message, err := suite.service.CreateMessage(suite.ctx, sessionID, messageType, role, content)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.NotNil(message)
	assert.Equal(sessionID, message.SessionID)
	assert.Equal(messageType, message.Type)
	assert.Equal(role, message.Role)
	assert.Equal(content, message.Content)
	assert.True(message.IsValid())
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
	suite.mockMessageRepo.AssertExpectations(suite.T())
}

// TestCreateMessage_SessionNotFound tests message creation with non-existent session
func (suite *MessageServiceTestSuite) TestCreateMessage_SessionNotFound() {
	sessionID := "invalid-session"
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(nil, errors.New("session not found"))

	// Execute
	message, err := suite.service.CreateMessage(suite.ctx, sessionID, entities.MessageTypeWorkRequest, entities.MessageRoleUser, "test content")

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Nil(message)
	assert.Contains(err.Error(), "session not found")
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestCreateMessage_SessionNil tests message creation when session is nil
func (suite *MessageServiceTestSuite) TestCreateMessage_SessionNil() {
	sessionID := "null-session"
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(nil, nil)

	// Execute
	message, err := suite.service.CreateMessage(suite.ctx, sessionID, entities.MessageTypeWorkRequest, entities.MessageRoleUser, "test content")

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Nil(message)
	assert.Equal("session not found", err.Error())
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestCreateMessage_SessionNotActive tests message creation with inactive session
func (suite *MessageServiceTestSuite) TestCreateMessage_SessionNotActive() {
	sessionID := "inactive-session"
	
	// Create inactive session
	inactiveSession := entities.NewSession("user-123")
	inactiveSession.Deactivate()
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(inactiveSession, nil)

	// Execute
	message, err := suite.service.CreateMessage(suite.ctx, sessionID, entities.MessageTypeWorkRequest, entities.MessageRoleUser, "test content")

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Nil(message)
	assert.Equal("session is not active", err.Error())
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestCreateMessage_RepositoryCreateError tests message creation with repository error
func (suite *MessageServiceTestSuite) TestCreateMessage_RepositoryCreateError() {
	sessionID := "session-123"
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(suite.activeSession, nil)
	suite.mockMessageRepo.On("Create", suite.ctx, mock.AnythingOfType("*entities.Message")).Return(errors.New("database error"))

	// Execute
	message, err := suite.service.CreateMessage(suite.ctx, sessionID, entities.MessageTypeWorkRequest, entities.MessageRoleUser, "test content")

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Nil(message)
	assert.Contains(err.Error(), "database error")
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
	suite.mockMessageRepo.AssertExpectations(suite.T())
}

// TestGetConversationHistory_Success tests getting conversation history with default limit
func (suite *MessageServiceTestSuite) TestGetConversationHistory_Success() {
	sessionID := "session-123"
	
	// Create test messages
	messages := []*entities.Message{
		entities.NewMessage(sessionID, entities.MessageTypeWorkRequest, entities.MessageRoleUser, "Hello"),
		entities.NewMessage(sessionID, entities.MessageTypeWorkRequest, entities.MessageRoleAssistant, "Hi there!"),
		entities.NewMessage(sessionID, entities.MessageTypeStatus, entities.MessageRoleSystem, "Processing"),
	}
	
	// Setup mocks
	suite.mockMessageRepo.On("GetBySessionIDWithPagination", suite.ctx, sessionID, 0, 50).Return(messages, nil)

	// Execute
	result, err := suite.service.GetConversationHistory(suite.ctx, sessionID, 0)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.Equal(messages, result)
	assert.Len(result, 3)
	
	// Verify mock expectations
	suite.mockMessageRepo.AssertExpectations(suite.T())
}

// TestGetConversationHistory_WithCustomLimit tests getting conversation history with custom limit
func (suite *MessageServiceTestSuite) TestGetConversationHistory_WithCustomLimit() {
	sessionID := "session-123"
	limit := 10
	
	// Create test messages
	messages := []*entities.Message{
		entities.NewMessage(sessionID, entities.MessageTypeWorkRequest, entities.MessageRoleUser, "Hello"),
		entities.NewMessage(sessionID, entities.MessageTypeWorkRequest, entities.MessageRoleAssistant, "Hi there!"),
	}
	
	// Setup mocks
	suite.mockMessageRepo.On("GetBySessionIDWithPagination", suite.ctx, sessionID, 0, limit).Return(messages, nil)

	// Execute
	result, err := suite.service.GetConversationHistory(suite.ctx, sessionID, limit)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.Equal(messages, result)
	assert.Len(result, 2)
	
	// Verify mock expectations
	suite.mockMessageRepo.AssertExpectations(suite.T())
}

// TestGetConversationHistory_NegativeLimit tests getting conversation history with negative limit
func (suite *MessageServiceTestSuite) TestGetConversationHistory_NegativeLimit() {
	sessionID := "session-123"
	limit := -5
	
	// Create test messages
	messages := []*entities.Message{
		entities.NewMessage(sessionID, entities.MessageTypeWorkRequest, entities.MessageRoleUser, "Hello"),
	}
	
	// Setup mocks - should use default limit of 50
	suite.mockMessageRepo.On("GetBySessionIDWithPagination", suite.ctx, sessionID, 0, 50).Return(messages, nil)

	// Execute
	result, err := suite.service.GetConversationHistory(suite.ctx, sessionID, limit)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.Equal(messages, result)
	
	// Verify mock expectations
	suite.mockMessageRepo.AssertExpectations(suite.T())
}

// TestAddMetadataToMessage_Success tests adding metadata to an existing message
func (suite *MessageServiceTestSuite) TestAddMetadataToMessage_Success() {
	messageID := "message-123"
	key := "priority"
	value := "high"
	
	// Create test message
	message := entities.NewMessage("session-123", entities.MessageTypeWorkRequest, entities.MessageRoleUser, "Test message")
	
	// Setup mocks
	suite.mockMessageRepo.On("GetByID", suite.ctx, messageID).Return(message, nil)
	suite.mockMessageRepo.On("Update", suite.ctx, mock.AnythingOfType("*entities.Message")).Return(nil)

	// Execute
	err := suite.service.AddMetadataToMessage(suite.ctx, messageID, key, value)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	
	// Verify metadata was added
	actualValue, exists := message.GetMetadata(key)
	assert.True(exists)
	assert.Equal(value, actualValue)
	
	// Verify mock expectations
	suite.mockMessageRepo.AssertExpectations(suite.T())
}

// TestAddMetadataToMessage_MessageNotFound tests adding metadata to non-existent message
func (suite *MessageServiceTestSuite) TestAddMetadataToMessage_MessageNotFound() {
	messageID := "invalid-message"
	
	// Setup mocks
	suite.mockMessageRepo.On("GetByID", suite.ctx, messageID).Return(nil, errors.New("message not found"))

	// Execute
	err := suite.service.AddMetadataToMessage(suite.ctx, messageID, "key", "value")

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Contains(err.Error(), "message not found")
	
	// Verify mock expectations
	suite.mockMessageRepo.AssertExpectations(suite.T())
}

// TestAddMetadataToMessage_MessageNil tests adding metadata when message is nil
func (suite *MessageServiceTestSuite) TestAddMetadataToMessage_MessageNil() {
	messageID := "null-message"
	
	// Setup mocks
	suite.mockMessageRepo.On("GetByID", suite.ctx, messageID).Return(nil, nil)

	// Execute
	err := suite.service.AddMetadataToMessage(suite.ctx, messageID, "key", "value")

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Equal("message not found", err.Error())
	
	// Verify mock expectations
	suite.mockMessageRepo.AssertExpectations(suite.T())
}

// TestGetMessagesByType_Success tests getting messages by type
func (suite *MessageServiceTestSuite) TestGetMessagesByType_Success() {
	sessionID := "session-123"
	messageType := entities.MessageTypeWorkRequest
	
	// Create test messages
	messages := []*entities.Message{
		entities.NewMessage(sessionID, messageType, entities.MessageRoleUser, "Request 1"),
		entities.NewMessage(sessionID, messageType, entities.MessageRoleAssistant, "Response 1"),
	}
	
	// Setup mocks
	suite.mockMessageRepo.On("GetByTypeAndSessionID", suite.ctx, messageType, sessionID).Return(messages, nil)

	// Execute
	result, err := suite.service.GetMessagesByType(suite.ctx, sessionID, messageType)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.Equal(messages, result)
	assert.Len(result, 2)
	
	// Verify mock expectations
	suite.mockMessageRepo.AssertExpectations(suite.T())
}

// TestGetMessagesByType_RepositoryError tests getting messages by type with repository error
func (suite *MessageServiceTestSuite) TestGetMessagesByType_RepositoryError() {
	sessionID := "session-123"
	messageType := entities.MessageTypeWorkRequest
	
	// Setup mocks
	suite.mockMessageRepo.On("GetByTypeAndSessionID", suite.ctx, messageType, sessionID).Return([]*entities.Message(nil), errors.New("database error"))

	// Execute
	result, err := suite.service.GetMessagesByType(suite.ctx, sessionID, messageType)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Nil(result)
	assert.Contains(err.Error(), "database error")
	
	// Verify mock expectations
	suite.mockMessageRepo.AssertExpectations(suite.T())
}

// TestValidateMessageContent_Success tests valid message content
func (suite *MessageServiceTestSuite) TestValidateMessageContent_Success() {
	validContents := []string{
		"Hello, world!",
		"This is a valid message with proper length.",
		"Code review request: function test() { return true; }",
	}
	
	for _, content := range validContents {
		suite.T().Run("valid content: "+content[:min(len(content), 20)], func(t *testing.T) {
			err := suite.service.ValidateMessageContent(content)
			assert.NoError(t, err)
		})
	}
}

// TestValidateMessageContent_EmptyContent tests empty message content
func (suite *MessageServiceTestSuite) TestValidateMessageContent_EmptyContent() {
	err := suite.service.ValidateMessageContent("")
	
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Equal("message content cannot be empty", err.Error())
}

// TestValidateMessageContent_TooLong tests message content that is too long
func (suite *MessageServiceTestSuite) TestValidateMessageContent_TooLong() {
	// Create content longer than 10000 characters
	longContent := make([]byte, 10001)
	for i := range longContent {
		longContent[i] = 'a'
	}
	
	err := suite.service.ValidateMessageContent(string(longContent))
	
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Equal("message content is too long", err.Error())
}

// TestMessageService_CompleteWorkflow tests a complete message workflow
func (suite *MessageServiceTestSuite) TestMessageService_CompleteWorkflow() {
	sessionID := "workflow-session"
	
	// Create messages for conversation
	userMessage := entities.NewMessage(sessionID, entities.MessageTypeWorkRequest, entities.MessageRoleUser, "Please help me debug this code")
	systemMessage := entities.NewMessage(sessionID, entities.MessageTypeStatus, entities.MessageRoleSystem, "Processing request")
	assistantMessage := entities.NewMessage(sessionID, entities.MessageTypeWorkRequest, entities.MessageRoleAssistant, "I'll help you debug that code. Please share the code snippet.")
	
	allMessages := []*entities.Message{userMessage, systemMessage, assistantMessage}
	workRequestMessages := []*entities.Message{userMessage, assistantMessage}
	
	// Setup mocks for creating messages
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(suite.activeSession, nil).Times(3)
	suite.mockMessageRepo.On("Create", suite.ctx, mock.AnythingOfType("*entities.Message")).Return(nil).Times(3)
	
	// Setup mocks for retrieving conversation history
	suite.mockMessageRepo.On("GetBySessionIDWithPagination", suite.ctx, sessionID, 0, 50).Return(allMessages, nil)
	
	// Setup mocks for retrieving messages by type
	suite.mockMessageRepo.On("GetByTypeAndSessionID", suite.ctx, entities.MessageTypeWorkRequest, sessionID).Return(workRequestMessages, nil)
	
	// Setup mocks for adding metadata
	suite.mockMessageRepo.On("GetByID", suite.ctx, userMessage.ID).Return(userMessage, nil)
	suite.mockMessageRepo.On("Update", suite.ctx, mock.AnythingOfType("*entities.Message")).Return(nil)

	assert := assert.New(suite.T())
	
	// Create user message
	createdUserMsg, err := suite.service.CreateMessage(suite.ctx, sessionID, entities.MessageTypeWorkRequest, entities.MessageRoleUser, "Please help me debug this code")
	assert.NoError(err)
	assert.NotNil(createdUserMsg)
	
	// Create system message
	createdSystemMsg, err := suite.service.CreateMessage(suite.ctx, sessionID, entities.MessageTypeStatus, entities.MessageRoleSystem, "Processing request")
	assert.NoError(err)
	assert.NotNil(createdSystemMsg)
	
	// Create assistant message
	createdAssistantMsg, err := suite.service.CreateMessage(suite.ctx, sessionID, entities.MessageTypeWorkRequest, entities.MessageRoleAssistant, "I'll help you debug that code. Please share the code snippet.")
	assert.NoError(err)
	assert.NotNil(createdAssistantMsg)
	
	// Get conversation history
	history, err := suite.service.GetConversationHistory(suite.ctx, sessionID, 0)
	assert.NoError(err)
	assert.Len(history, 3)
	
	// Get work request messages only
	workRequests, err := suite.service.GetMessagesByType(suite.ctx, sessionID, entities.MessageTypeWorkRequest)
	assert.NoError(err)
	assert.Len(workRequests, 2)
	
	// Add metadata to user message
	err = suite.service.AddMetadataToMessage(suite.ctx, userMessage.ID, "urgency", "high")
	assert.NoError(err)
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
	suite.mockMessageRepo.AssertExpectations(suite.T())
}

// BenchmarkCreateMessage benchmarks message creation
func BenchmarkCreateMessage(b *testing.B) {
	mockMessageRepo := &MockMessageRepository{}
	mockSessionRepo := &MockSessionRepository{}
	service := NewMessageService(mockMessageRepo, mockSessionRepo)
	ctx := context.Background()
	
	activeSession := entities.NewSession("user-123")
	
	// Setup mocks
	mockSessionRepo.On("GetByID", ctx, "session-123").Return(activeSession, nil)
	mockMessageRepo.On("Create", ctx, mock.AnythingOfType("*entities.Message")).Return(nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CreateMessage(ctx, "session-123", entities.MessageTypeWorkRequest, entities.MessageRoleUser, "Benchmark message")
	}
}

// BenchmarkValidateMessageContent benchmarks message content validation
func BenchmarkValidateMessageContent(b *testing.B) {
	mockMessageRepo := &MockMessageRepository{}
	mockSessionRepo := &MockSessionRepository{}
	service := NewMessageService(mockMessageRepo, mockSessionRepo)
	
	content := "This is a sample message content for benchmarking validation performance."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ValidateMessageContent(content)
	}
}

// Helper function for min operation
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}