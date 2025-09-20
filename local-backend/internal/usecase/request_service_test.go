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
type MockRequestRepository struct {
	mock.Mock
}

func (m *MockRequestRepository) Save(ctx context.Context, request *domain.Request) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockRequestRepository) GetByID(ctx context.Context, id string) (*domain.Request, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Request), args.Error(1)
}

func (m *MockRequestRepository) GetByMessageID(ctx context.Context, messageID string) (*domain.Request, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Request), args.Error(1)
}

func (m *MockRequestRepository) GetPendingRequests(ctx context.Context) ([]*domain.Request, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Request), args.Error(1)
}

func (m *MockRequestRepository) Update(ctx context.Context, request *domain.Request) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockRequestRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockClaudeService struct {
	mock.Mock
}

func (m *MockClaudeService) ProcessRequest(ctx context.Context, message *domain.Message) (string, error) {
	args := m.Called(ctx, message)
	return args.String(0), args.Error(1)
}

func TestNewRequestService(t *testing.T) {
	mockRequestRepo := &MockRequestRepository{}
	mockClaudeService := &MockClaudeService{}
	mockLogger := &MockLogger{}

	service := NewRequestService(mockRequestRepo, mockClaudeService, mockLogger)

	assert.NotNil(t, service)
	assert.Equal(t, mockRequestRepo, service.requestRepo)
	assert.Equal(t, mockClaudeService, service.claudeService)
	assert.Equal(t, mockLogger, service.logger)
}

func TestRequestService_ProcessWorkRequest(t *testing.T) {
	mockRequestRepo := &MockRequestRepository{}
	mockClaudeService := &MockClaudeService{}
	mockLogger := &MockLogger{}

	service := NewRequestService(mockRequestRepo, mockClaudeService, mockLogger)

	t.Run("successful work request processing", func(t *testing.T) {
		message := domain.NewMessage("msg-123", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func main() { println(\"hello\") }",
			"task": "analyze this code",
		})

		expectedResponse := "The code is a simple Go program that prints 'hello' to the console."

		// Set up mock expectations
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockRequestRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Request")).Return(nil).Twice() // Once for create, once for update
		mockClaudeService.On("ProcessRequest", mock.Anything, message).Return(expectedResponse, nil)

		err := service.ProcessWorkRequest(context.Background(), message)

		assert.NoError(t, err)
		mockRequestRepo.AssertExpectations(t)
		mockClaudeService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("request creation failure", func(t *testing.T) {
		message := domain.NewMessage("msg-fail", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "test",
		})

		saveError := errors.New("database error")

		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockRequestRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Request")).Return(saveError)

		err := service.ProcessWorkRequest(context.Background(), message)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save request")
		mockRequestRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Claude service processing failure", func(t *testing.T) {
		message := domain.NewMessage("msg-claude-fail", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "invalid code",
		})

		claudeError := errors.New("Claude API error")

		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockRequestRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Request")).Return(nil).Twice() // Create and fail update
		mockClaudeService.On("ProcessRequest", mock.Anything, message).Return("", claudeError)

		err := service.ProcessWorkRequest(context.Background(), message)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to process request with Claude")
		mockRequestRepo.AssertExpectations(t)
		mockClaudeService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("request completion failure", func(t *testing.T) {
		message := domain.NewMessage("msg-complete-fail", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func test() {}",
		})

		expectedResponse := "Code analysis complete"
		updateError := errors.New("failed to update request")

		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockRequestRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Request")).Return(nil).Once()  // Create succeeds
		mockRequestRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Request")).Return(updateError).Once() // Update fails
		mockClaudeService.On("ProcessRequest", mock.Anything, message).Return(expectedResponse, nil)

		err := service.ProcessWorkRequest(context.Background(), message)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save completed request")
		mockRequestRepo.AssertExpectations(t)
		mockClaudeService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("context cancellation during processing", func(t *testing.T) {
		message := domain.NewMessage("msg-cancel", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func long() {}",
		})

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockRequestRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Request")).Return(nil)

		err := service.ProcessWorkRequest(ctx, message)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
		mockRequestRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestRequestService_ProcessCancellation(t *testing.T) {
	mockRequestRepo := &MockRequestRepository{}
	mockClaudeService := &MockClaudeService{}
	mockLogger := &MockLogger{}

	service := NewRequestService(mockRequestRepo, mockClaudeService, mockLogger)

	t.Run("successful cancellation", func(t *testing.T) {
		message := domain.NewMessage("cancel-123", domain.MessageTypeCancellation, map[string]interface{}{
			"request_id": "req-456",
		})

		existingRequest := domain.NewRequest("req-456", "msg-original", "ctx-123", "{\"code\": \"test\"}")
		existingRequest.Start() // Make it processing so it can be cancelled

		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockRequestRepo.On("GetByID", mock.Anything, "req-456").Return(existingRequest, nil)
		mockRequestRepo.On("Save", mock.Anything, existingRequest).Return(nil)

		err := service.ProcessCancellation(context.Background(), message)

		assert.NoError(t, err)
		assert.Equal(t, domain.RequestStatusCancelled, existingRequest.Status)
		mockRequestRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("invalid cancellation payload", func(t *testing.T) {
		message := domain.NewMessage("cancel-invalid", domain.MessageTypeCancellation, map[string]interface{}{
			// Missing request_id
		})

		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()

		err := service.ProcessCancellation(context.Background(), message)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid cancellation payload")
		mockLogger.AssertExpectations(t)
	})

	t.Run("request not found", func(t *testing.T) {
		message := domain.NewMessage("cancel-notfound", domain.MessageTypeCancellation, map[string]interface{}{
			"request_id": "nonexistent",
		})

		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockRequestRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, errors.New("request not found"))

		err := service.ProcessCancellation(context.Background(), message)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find request")
		mockRequestRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("request cannot be cancelled", func(t *testing.T) {
		message := domain.NewMessage("cancel-completed", domain.MessageTypeCancellation, map[string]interface{}{
			"request_id": "req-completed",
		})

		completedRequest := domain.NewRequest("req-completed", "msg-original", "ctx-123", "{\"code\": \"test\"}")
		completedRequest.Complete("Analysis complete") // Make it completed, so it cannot be cancelled

		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockRequestRepo.On("GetByID", mock.Anything, "req-completed").Return(completedRequest, nil)

		err := service.ProcessCancellation(context.Background(), message)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request cannot be cancelled")
		mockRequestRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("cancellation save failure", func(t *testing.T) {
		message := domain.NewMessage("cancel-save-fail", domain.MessageTypeCancellation, map[string]interface{}{
			"request_id": "req-save-fail",
		})

		existingRequest := domain.NewRequest("req-save-fail", "msg-original", "ctx-123", "{\"code\": \"test\"}")
		saveError := errors.New("database error")

		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockRequestRepo.On("GetByID", mock.Anything, "req-save-fail").Return(existingRequest, nil)
		mockRequestRepo.On("Save", mock.Anything, existingRequest).Return(saveError)

		err := service.ProcessCancellation(context.Background(), message)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save cancelled request")
		mockRequestRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestRequestService_GetRequestStatus(t *testing.T) {
	mockRequestRepo := &MockRequestRepository{}
	mockClaudeService := &MockClaudeService{}
	mockLogger := &MockLogger{}

	service := NewRequestService(mockRequestRepo, mockClaudeService, mockLogger)

	t.Run("get existing request status", func(t *testing.T) {
		request := domain.NewRequest("req-123", "msg-456", "ctx-789", "{\"code\": \"test\"}")
		request.Complete("Analysis complete")

		mockRequestRepo.On("GetByID", mock.Anything, "req-123").Return(request, nil)

		result, err := service.GetRequestStatus(context.Background(), "req-123")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.RequestStatusCompleted, result.Status)
		assert.Equal(t, "Analysis complete", result.Response)
		mockRequestRepo.AssertExpectations(t)
	})

	t.Run("request not found", func(t *testing.T) {
		mockRequestRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))

		result, err := service.GetRequestStatus(context.Background(), "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRequestRepo.AssertExpectations(t)
	})
}

func TestRequestService_GetPendingRequests(t *testing.T) {
	mockRequestRepo := &MockRequestRepository{}
	mockClaudeService := &MockClaudeService{}
	mockLogger := &MockLogger{}

	service := NewRequestService(mockRequestRepo, mockClaudeService, mockLogger)

	t.Run("get pending requests", func(t *testing.T) {
		pendingRequests := []*domain.Request{
			domain.NewRequest("req-1", "msg-1", "ctx-1", "{\"code\": \"test1\"}"),
			domain.NewRequest("req-2", "msg-2", "ctx-2", "{\"code\": \"test2\"}"),
		}

		mockRequestRepo.On("GetPendingRequests", mock.Anything).Return(pendingRequests, nil)

		result, err := service.GetPendingRequests(context.Background())

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, pendingRequests, result)
		mockRequestRepo.AssertExpectations(t)
	})

	t.Run("no pending requests", func(t *testing.T) {
		mockRequestRepo.On("GetPendingRequests", mock.Anything).Return([]*domain.Request{}, nil)

		result, err := service.GetPendingRequests(context.Background())

		assert.NoError(t, err)
		assert.Len(t, result, 0)
		mockRequestRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		repositoryError := errors.New("database error")
		mockRequestRepo.On("GetPendingRequests", mock.Anything).Return(nil, repositoryError)

		result, err := service.GetPendingRequests(context.Background())

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRequestRepo.AssertExpectations(t)
	})
}

func TestRequestService_GenerateRequestID(t *testing.T) {
	service := &RequestService{}

	id := service.generateRequestID()
	
	assert.NotEmpty(t, id)
	assert.True(t, len(id) > 10) // Should be a reasonable length
	
	// Generate another ID to ensure uniqueness
	id2 := service.generateRequestID()
	assert.NotEqual(t, id, id2)
}

func TestRequestService_ParseRequestID(t *testing.T) {
	service := &RequestService{}

	t.Run("valid request ID in payload", func(t *testing.T) {
		payload := map[string]interface{}{
			"request_id": "req-123",
		}

		requestID, err := service.parseRequestID(payload)

		assert.NoError(t, err)
		assert.Equal(t, "req-123", requestID)
	})

	t.Run("missing request_id", func(t *testing.T) {
		payload := map[string]interface{}{
			"other_field": "value",
		}

		requestID, err := service.parseRequestID(payload)

		assert.Error(t, err)
		assert.Empty(t, requestID)
		assert.Contains(t, err.Error(), "request_id not found")
	})

	t.Run("invalid request_id type", func(t *testing.T) {
		payload := map[string]interface{}{
			"request_id": 123, // Not a string
		}

		requestID, err := service.parseRequestID(payload)

		assert.Error(t, err)
		assert.Empty(t, requestID)
		assert.Contains(t, err.Error(), "request_id must be a string")
	})

	t.Run("empty request_id", func(t *testing.T) {
		payload := map[string]interface{}{
			"request_id": "",
		}

		requestID, err := service.parseRequestID(payload)

		assert.Error(t, err)
		assert.Empty(t, requestID)
		assert.Contains(t, err.Error(), "request_id cannot be empty")
	})
}

func TestRequestService_IntegrationScenarios(t *testing.T) {
	mockRequestRepo := &MockRequestRepository{}
	mockClaudeService := &MockClaudeService{}
	mockLogger := &MockLogger{}

	service := NewRequestService(mockRequestRepo, mockClaudeService, mockLogger)

	t.Run("complete request lifecycle", func(t *testing.T) {
		// Step 1: Process work request
		message := domain.NewMessage("msg-lifecycle", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func main() { println(\"test\") }",
			"task": "analyze",
		})

		claudeResponse := "This is a simple Go program"

		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return().Times(3)
		mockRequestRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Request")).Return(nil).Twice()
		mockClaudeService.On("ProcessRequest", mock.Anything, message).Return(claudeResponse, nil)

		err := service.ProcessWorkRequest(context.Background(), message)
		assert.NoError(t, err)

		// Step 2: Check status (simulated by capturing the request from Save calls)
		var savedRequest *domain.Request
		mockRequestRepo.On("GetByID", mock.Anything, mock.AnythingOfType("string")).Return(func(ctx context.Context, id string) *domain.Request {
			return savedRequest
		}, func(ctx context.Context, id string) error {
			if savedRequest == nil {
				return errors.New("not found")
			}
			return nil
		})

		// Verify all expectations were met
		mockRequestRepo.AssertExpectations(t)
		mockClaudeService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("request processing with cancellation race condition", func(t *testing.T) {
		// This test simulates a scenario where a cancellation comes in
		// while a request is being processed
		
		workMessage := domain.NewMessage("msg-race", domain.MessageTypeWorkRequest, map[string]interface{}{
			"code": "func slow() { time.Sleep(time.Hour) }",
			"task": "analyze",
		})

		cancelMessage := domain.NewMessage("cancel-race", domain.MessageTypeCancellation, map[string]interface{}{
			"request_id": "req-race",
		})

		// Set up work request processing
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return().Times(2)
		mockRequestRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Request")).Return(nil).Once()
		
		// Simulate Claude taking time and context getting cancelled
		mockClaudeService.On("ProcessRequest", mock.Anything, workMessage).Return("", context.Canceled)
		mockRequestRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Request")).Return(nil).Once() // For failure update

		// Set up cancellation processing  
		processingRequest := domain.NewRequest("req-race", "msg-race", "ctx-race", "{\"code\": \"test\"}")
		processingRequest.Start()
		mockRequestRepo.On("GetByID", mock.Anything, "req-race").Return(processingRequest, nil)
		mockRequestRepo.On("Save", mock.Anything, processingRequest).Return(nil)

		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return().Once() // For work request failure

		// Execute both operations
		err1 := service.ProcessWorkRequest(context.Background(), workMessage)
		err2 := service.ProcessCancellation(context.Background(), cancelMessage)

		// Work request should fail due to cancellation, cancellation should succeed
		assert.Error(t, err1)
		assert.NoError(t, err2)

		mockRequestRepo.AssertExpectations(t)
		mockClaudeService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}