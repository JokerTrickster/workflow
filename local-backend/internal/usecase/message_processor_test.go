package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JokerTrickster/workflow/local-backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing
type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Save(ctx context.Context, message *domain.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageRepository) GetByID(ctx context.Context, id string) (*domain.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Message), args.Error(1)
}

func (m *MockMessageRepository) GetByContextID(ctx context.Context, contextID string) ([]*domain.Message, error) {
	args := m.Called(ctx, contextID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Message), args.Error(1)
}

func (m *MockMessageRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockContextRepository struct {
	mock.Mock
}

func (m *MockContextRepository) Save(ctx context.Context, context *domain.ProcessingContext) error {
	args := m.Called(ctx, context)
	return args.Error(0)
}

func (m *MockContextRepository) GetByID(ctx context.Context, id string) (*domain.ProcessingContext, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProcessingContext), args.Error(1)
}

func (m *MockContextRepository) GetExpired(ctx context.Context, maxAge time.Duration) ([]*domain.ProcessingContext, error) {
	args := m.Called(ctx, maxAge)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ProcessingContext), args.Error(1)
}

func (m *MockContextRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockRequestService struct {
	mock.Mock
}

func (m *MockRequestService) ProcessWorkRequest(ctx context.Context, message *domain.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockRequestService) ProcessCancellation(ctx context.Context, message *domain.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func TestNewMessageProcessor(t *testing.T) {
	mockMessageRepo := &MockMessageRepository{}
	mockContextRepo := &MockContextRepository{}
	mockRequestService := &MockRequestService{}
	mockLogger := &MockLogger{}

	processor := NewMessageProcessor(mockMessageRepo, mockContextRepo, mockRequestService, mockLogger)

	assert.NotNil(t, processor)
	assert.Equal(t, mockMessageRepo, processor.messageRepo)
	assert.Equal(t, mockContextRepo, processor.contextRepo)
	assert.Equal(t, mockRequestService, processor.requestService)
	assert.Equal(t, mockLogger, processor.logger)
}

func TestMessageProcessor_ProcessMessage_WorkRequest(t *testing.T) {
	mockMessageRepo := &MockMessageRepository{}
	mockContextRepo := &MockContextRepository{}
	mockRequestService := &MockRequestService{}
	mockLogger := &MockLogger{}

	processor := NewMessageProcessor(mockMessageRepo, mockContextRepo, mockRequestService, mockLogger)

	t.Run("successful work request processing", func(t *testing.T) {
		message := domain.NewMessage("msg-123", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func main() { println(\"hello\") }",
			"task": "analyze this code",
		})

		// Set up mock expectations
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockMessageRepo.On("Save", mock.Anything, message).Return(nil)
		mockRequestService.On("ProcessWorkRequest", mock.Anything, message).Return(nil)

		err := processor.ProcessMessage(context.Background(), message)

		assert.NoError(t, err)
		mockMessageRepo.AssertExpectations(t)
		mockRequestService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("work request with context ID", func(t *testing.T) {
		message := domain.NewMessage("msg-456", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func test() {}",
			"task": "review code",
		})
		contextID := "ctx-789"
		message.SetContextID(contextID)

		existingContext := domain.NewProcessingContext(contextID)

		// Set up mock expectations
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockContextRepo.On("GetByID", mock.Anything, contextID).Return(existingContext, nil)
		mockMessageRepo.On("Save", mock.Anything, message).Return(nil)
		mockRequestService.On("ProcessWorkRequest", mock.Anything, message).Return(nil)

		err := processor.ProcessMessage(context.Background(), message)

		assert.NoError(t, err)
		mockContextRepo.AssertExpectations(t)
		mockMessageRepo.AssertExpectations(t)
		mockRequestService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("message save failure", func(t *testing.T) {
		message := domain.NewMessage("msg-fail", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "test",
		})

		saveError := errors.New("database connection failed")

		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockMessageRepo.On("Save", mock.Anything, message).Return(saveError)

		err := processor.ProcessMessage(context.Background(), message)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save message")
		mockMessageRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("request processing failure", func(t *testing.T) {
		message := domain.NewMessage("msg-proc-fail", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "invalid code",
		})

		processingError := errors.New("Claude API error")

		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockMessageRepo.On("Save", mock.Anything, message).Return(nil)
		mockRequestService.On("ProcessWorkRequest", mock.Anything, message).Return(processingError)

		err := processor.ProcessMessage(context.Background(), message)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to process work request")
		mockMessageRepo.AssertExpectations(t)
		mockRequestService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestMessageProcessor_ProcessMessage_Cancellation(t *testing.T) {
	mockMessageRepo := &MockMessageRepository{}
	mockContextRepo := &MockContextRepository{}
	mockRequestService := &MockRequestService{}
	mockLogger := &MockLogger{}

	processor := NewMessageProcessor(mockMessageRepo, mockContextRepo, mockRequestService, mockLogger)

	t.Run("successful cancellation processing", func(t *testing.T) {
		message := domain.NewMessage("cancel-123", domain.MessageTypeCancellation, map[string]interface{}{
			"request_id": "req-456",
		})

		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockMessageRepo.On("Save", mock.Anything, message).Return(nil)
		mockRequestService.On("ProcessCancellation", mock.Anything, message).Return(nil)

		err := processor.ProcessMessage(context.Background(), message)

		assert.NoError(t, err)
		mockMessageRepo.AssertExpectations(t)
		mockRequestService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("cancellation processing failure", func(t *testing.T) {
		message := domain.NewMessage("cancel-fail", domain.MessageTypeCancellation, map[string]interface{}{
			"request_id": "nonexistent",
		})

		processingError := errors.New("request not found")

		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockMessageRepo.On("Save", mock.Anything, message).Return(nil)
		mockRequestService.On("ProcessCancellation", mock.Anything, message).Return(processingError)

		err := processor.ProcessMessage(context.Background(), message)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to process cancellation")
		mockMessageRepo.AssertExpectations(t)
		mockRequestService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestMessageProcessor_ProcessMessage_UnknownType(t *testing.T) {
	mockMessageRepo := &MockMessageRepository{}
	mockContextRepo := &MockContextRepository{}
	mockRequestService := &MockRequestService{}
	mockLogger := &MockLogger{}

	processor := NewMessageProcessor(mockMessageRepo, mockContextRepo, mockRequestService, mockLogger)

	message := domain.NewMessage("unknown-123", "unknown_type", map[string]interface{}{})

	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
	mockMessageRepo.On("Save", mock.Anything, message).Return(nil)

	err := processor.ProcessMessage(context.Background(), message)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown message type")
	mockMessageRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestMessageProcessor_EnsureContext(t *testing.T) {
	mockMessageRepo := &MockMessageRepository{}
	mockContextRepo := &MockContextRepository{}
	mockRequestService := &MockRequestService{}
	mockLogger := &MockLogger{}

	processor := NewMessageProcessor(mockMessageRepo, mockContextRepo, mockRequestService, mockLogger)

	t.Run("existing context", func(t *testing.T) {
		contextID := "existing-ctx"
		existingContext := domain.NewProcessingContext(contextID)

		mockContextRepo.On("GetByID", mock.Anything, contextID).Return(existingContext, nil)

		ctx, err := processor.ensureContext(context.Background(), contextID)

		assert.NoError(t, err)
		assert.Equal(t, existingContext, ctx)
		mockContextRepo.AssertExpectations(t)
	})

	t.Run("context not found - create new", func(t *testing.T) {
		contextID := "new-ctx"

		mockContextRepo.On("GetByID", mock.Anything, contextID).Return(nil, errors.New("context not found"))
		mockContextRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.ProcessingContext")).Return(nil)

		ctx, err := processor.ensureContext(context.Background(), contextID)

		assert.NoError(t, err)
		assert.NotNil(t, ctx)
		assert.Equal(t, contextID, ctx.ID)
		mockContextRepo.AssertExpectations(t)
	})

	t.Run("context save failure", func(t *testing.T) {
		contextID := "fail-ctx"
		saveError := errors.New("database error")

		mockContextRepo.On("GetByID", mock.Anything, contextID).Return(nil, errors.New("context not found"))
		mockContextRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.ProcessingContext")).Return(saveError)

		ctx, err := processor.ensureContext(context.Background(), contextID)

		assert.Error(t, err)
		assert.Nil(t, ctx)
		assert.Contains(t, err.Error(), "failed to save new context")
		mockContextRepo.AssertExpectations(t)
	})
}

func TestMessageProcessor_CleanupExpiredContexts(t *testing.T) {
	mockMessageRepo := &MockMessageRepository{}
	mockContextRepo := &MockContextRepository{}
	mockRequestService := &MockRequestService{}
	mockLogger := &MockLogger{}

	processor := NewMessageProcessor(mockMessageRepo, mockContextRepo, mockRequestService, mockLogger)

	t.Run("successful cleanup", func(t *testing.T) {
		expiredContexts := []*domain.ProcessingContext{
			domain.NewProcessingContext("expired-1"),
			domain.NewProcessingContext("expired-2"),
		}

		maxAge := 1 * time.Hour

		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockContextRepo.On("GetExpired", mock.Anything, maxAge).Return(expiredContexts, nil)
		mockContextRepo.On("Delete", mock.Anything, "expired-1").Return(nil)
		mockContextRepo.On("Delete", mock.Anything, "expired-2").Return(nil)

		err := processor.CleanupExpiredContexts(context.Background(), maxAge)

		assert.NoError(t, err)
		mockContextRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("no expired contexts", func(t *testing.T) {
		maxAge := 1 * time.Hour

		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockContextRepo.On("GetExpired", mock.Anything, maxAge).Return([]*domain.ProcessingContext{}, nil)

		err := processor.CleanupExpiredContexts(context.Background(), maxAge)

		assert.NoError(t, err)
		mockContextRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("cleanup failure", func(t *testing.T) {
		expiredContexts := []*domain.ProcessingContext{
			domain.NewProcessingContext("expired-1"),
		}

		maxAge := 1 * time.Hour
		deleteError := errors.New("deletion failed")

		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockContextRepo.On("GetExpired", mock.Anything, maxAge).Return(expiredContexts, nil)
		mockContextRepo.On("Delete", mock.Anything, "expired-1").Return(deleteError)

		err := processor.CleanupExpiredContexts(context.Background(), maxAge)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete expired context")
		mockContextRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("get expired contexts failure", func(t *testing.T) {
		maxAge := 1 * time.Hour
		getError := errors.New("database error")

		mockContextRepo.On("GetExpired", mock.Anything, maxAge).Return(nil, getError)

		err := processor.CleanupExpiredContexts(context.Background(), maxAge)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get expired contexts")
		mockContextRepo.AssertExpectations(t)
	})
}

func TestMessageProcessor_GetProcessingStats(t *testing.T) {
	mockMessageRepo := &MockMessageRepository{}
	mockContextRepo := &MockContextRepository{}
	mockRequestService := &MockRequestService{}
	mockLogger := &MockLogger{}

	processor := NewMessageProcessor(mockMessageRepo, mockContextRepo, mockRequestService, mockLogger)

	stats := processor.GetProcessingStats()

	assert.NotNil(t, stats)
	assert.Contains(t, stats, "processed_messages")
	assert.Contains(t, stats, "failed_messages")
	assert.Contains(t, stats, "processing_errors")
	assert.Equal(t, int64(0), stats["processed_messages"])
	assert.Equal(t, int64(0), stats["failed_messages"])
	assert.Equal(t, int64(0), stats["processing_errors"])
}

func TestMessageProcessor_IntegrationScenarios(t *testing.T) {
	mockMessageRepo := &MockMessageRepository{}
	mockContextRepo := &MockContextRepository{}
	mockRequestService := &MockRequestService{}
	mockLogger := &MockLogger{}

	processor := NewMessageProcessor(mockMessageRepo, mockContextRepo, mockRequestService, mockLogger)

	t.Run("complete workflow with context management", func(t *testing.T) {
		contextID := "workflow-ctx"
		
		// First message - creates new context
		message1 := domain.NewMessage("msg-1", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func hello() {}",
			"task": "analyze",
		})
		message1.SetContextID(contextID)

		// Second message - uses existing context  
		message2 := domain.NewMessage("msg-2", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func world() {}",
			"task": "review",
		})
		message2.SetContextID(contextID)

		// Set up expectations for message 1
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return().Times(2)
		mockContextRepo.On("GetByID", mock.Anything, contextID).Return(nil, errors.New("not found")).Once()
		mockContextRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.ProcessingContext")).Return(nil).Once()
		mockMessageRepo.On("Save", mock.Anything, message1).Return(nil).Once()
		mockRequestService.On("ProcessWorkRequest", mock.Anything, message1).Return(nil).Once()

		// Set up expectations for message 2
		existingContext := domain.NewProcessingContext(contextID)
		mockContextRepo.On("GetByID", mock.Anything, contextID).Return(existingContext, nil).Once()
		mockMessageRepo.On("Save", mock.Anything, message2).Return(nil).Once()
		mockRequestService.On("ProcessWorkRequest", mock.Anything, message2).Return(nil).Once()

		// Process both messages
		err1 := processor.ProcessMessage(context.Background(), message1)
		err2 := processor.ProcessMessage(context.Background(), message2)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		
		mockContextRepo.AssertExpectations(t)
		mockMessageRepo.AssertExpectations(t)
		mockRequestService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}