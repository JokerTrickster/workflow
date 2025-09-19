package services

import (
	"context"
	"errors"
	"local-backend-server/internal/domain/entities"
	"local-backend-server/internal/domain/repositories"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockRequestRepository is a mock implementation of RequestRepository
type MockRequestRepository struct {
	mock.Mock
}

func (m *MockRequestRepository) Create(ctx context.Context, request *entities.Request) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockRequestRepository) GetByID(ctx context.Context, id string) (*entities.Request, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.Request), args.Error(1)
}

func (m *MockRequestRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*entities.Request, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]*entities.Request), args.Error(1)
}

func (m *MockRequestRepository) GetByStatus(ctx context.Context, status entities.RequestStatus) ([]*entities.Request, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]*entities.Request), args.Error(1)
}

func (m *MockRequestRepository) GetPendingRequests(ctx context.Context) ([]*entities.Request, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entities.Request), args.Error(1)
}

func (m *MockRequestRepository) GetProcessingRequests(ctx context.Context) ([]*entities.Request, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entities.Request), args.Error(1)
}

func (m *MockRequestRepository) Update(ctx context.Context, request *entities.Request) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockRequestRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRequestRepository) GetByTypeAndStatus(ctx context.Context, requestType entities.RequestType, status entities.RequestStatus) ([]*entities.Request, error) {
	args := m.Called(ctx, requestType, status)
	return args.Get(0).([]*entities.Request), args.Error(1)
}

func (m *MockRequestRepository) GetRequestsCreatedAfter(ctx context.Context, after time.Time) ([]*entities.Request, error) {
	args := m.Called(ctx, after)
	return args.Get(0).([]*entities.Request), args.Error(1)
}

func (m *MockRequestRepository) GetRequestsWithTimeout(ctx context.Context, timeoutDuration time.Duration) ([]*entities.Request, error) {
	args := m.Called(ctx, timeoutDuration)
	return args.Get(0).([]*entities.Request), args.Error(1)
}

func (m *MockRequestRepository) CountByStatus(ctx context.Context, status entities.RequestStatus) (int, error) {
	args := m.Called(ctx, status)
	return args.Int(0), args.Error(1)
}

func (m *MockRequestRepository) GetRequestMetrics(ctx context.Context, from, to time.Time) (*repositories.RequestMetrics, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).(*repositories.RequestMetrics), args.Error(1)
}

// MockSessionRepository is a mock implementation of SessionRepository
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *entities.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByID(ctx context.Context, id string) (*entities.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Session), args.Error(1)
}

func (m *MockSessionRepository) GetByUserID(ctx context.Context, userID string) ([]*entities.Session, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*entities.Session), args.Error(1)
}

func (m *MockSessionRepository) GetActiveByUserID(ctx context.Context, userID string) ([]*entities.Session, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*entities.Session), args.Error(1)
}

func (m *MockSessionRepository) Update(ctx context.Context, session *entities.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByStatus(ctx context.Context, status entities.SessionStatus) ([]*entities.Session, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]*entities.Session), args.Error(1)
}

func (m *MockSessionRepository) GetExpiredSessions(ctx context.Context) ([]*entities.Session, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entities.Session), args.Error(1)
}

func (m *MockSessionRepository) CleanupExpiredSessions(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockSessionRepository) CountByStatus(ctx context.Context, status entities.SessionStatus) (int, error) {
	args := m.Called(ctx, status)
	return args.Int(0), args.Error(1)
}

// RequestServiceTestSuite defines the test suite for RequestService
type RequestServiceTestSuite struct {
	suite.Suite
	service           *RequestService
	mockRequestRepo   *MockRequestRepository
	mockSessionRepo   *MockSessionRepository
	ctx               context.Context
	activeSession     *entities.Session
	validCodeReviewInput map[string]interface{}
}

// SetupTest runs before each test
func (suite *RequestServiceTestSuite) SetupTest() {
	suite.mockRequestRepo = &MockRequestRepository{}
	suite.mockSessionRepo = &MockSessionRepository{}
	suite.service = NewRequestService(suite.mockRequestRepo, suite.mockSessionRepo)
	suite.ctx = context.Background()
	
	// Create a valid active session for testing
	suite.activeSession = entities.NewSession("user-123")
	
	// Valid input for code review requests
	suite.validCodeReviewInput = map[string]interface{}{
		"code":        "function test() { return true; }",
		"language":    "javascript",
		"description": "Test function validation",
	}
}

// TestRequestServiceTestSuite runs the test suite
func TestRequestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RequestServiceTestSuite))
}

// TestNewRequestService tests the NewRequestService constructor
func (suite *RequestServiceTestSuite) TestNewRequestService() {
	service := NewRequestService(suite.mockRequestRepo, suite.mockSessionRepo)
	
	assert := assert.New(suite.T())
	assert.NotNil(service)
	assert.Equal(suite.mockRequestRepo, service.requestRepo)
	assert.Equal(suite.mockSessionRepo, service.sessionRepo)
}

// TestCreateRequest_Success tests successful request creation
func (suite *RequestServiceTestSuite) TestCreateRequest_Success() {
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, "session-123").Return(suite.activeSession, nil)
	suite.mockRequestRepo.On("Create", suite.ctx, mock.AnythingOfType("*entities.Request")).Return(nil)

	// Execute
	request, err := suite.service.CreateRequest(suite.ctx, "session-123", entities.RequestTypeCodeReview, suite.validCodeReviewInput)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.NotNil(request)
	assert.Equal("session-123", request.SessionID)
	assert.Equal(entities.RequestTypeCodeReview, request.Type)
	assert.Equal(entities.RequestStatusPending, request.Status)
	assert.Equal(suite.validCodeReviewInput, request.Input)
	assert.True(request.IsValid())
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
	suite.mockRequestRepo.AssertExpectations(suite.T())
}

// TestCreateRequest_SessionNotFound tests request creation with non-existent session
func (suite *RequestServiceTestSuite) TestCreateRequest_SessionNotFound() {
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, "invalid-session").Return(nil, errors.New("session not found"))

	// Execute
	request, err := suite.service.CreateRequest(suite.ctx, "invalid-session", entities.RequestTypeCodeReview, suite.validCodeReviewInput)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Nil(request)
	assert.Contains(err.Error(), "session not found")
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestCreateRequest_SessionNil tests request creation when session is nil
func (suite *RequestServiceTestSuite) TestCreateRequest_SessionNil() {
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, "null-session").Return(nil, nil)

	// Execute
	request, err := suite.service.CreateRequest(suite.ctx, "null-session", entities.RequestTypeCodeReview, suite.validCodeReviewInput)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Nil(request)
	assert.Equal("session not found", err.Error())
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestCreateRequest_SessionNotActive tests request creation with inactive session
func (suite *RequestServiceTestSuite) TestCreateRequest_SessionNotActive() {
	// Create inactive session
	inactiveSession := entities.NewSession("user-123")
	inactiveSession.Deactivate()
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, "inactive-session").Return(inactiveSession, nil)

	// Execute
	request, err := suite.service.CreateRequest(suite.ctx, "inactive-session", entities.RequestTypeCodeReview, suite.validCodeReviewInput)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Nil(request)
	assert.Equal("session is not active", err.Error())
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestCreateRequest_InvalidInput tests request creation with invalid input
func (suite *RequestServiceTestSuite) TestCreateRequest_InvalidInput() {
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, "session-123").Return(suite.activeSession, nil)

	// Test cases for invalid input
	testCases := []struct {
		name        string
		requestType entities.RequestType
		input       map[string]interface{}
		expectedErr string
	}{
		{
			name:        "nil input",
			requestType: entities.RequestTypeCodeReview,
			input:       nil,
			expectedErr: "request input cannot be nil",
		},
		{
			name:        "code review missing code",
			requestType: entities.RequestTypeCodeReview,
			input:       map[string]interface{}{"description": "test"},
			expectedErr: "code review request must include 'code' field",
		},
		{
			name:        "issue analysis missing issue",
			requestType: entities.RequestTypeIssueAnalysis,
			input:       map[string]interface{}{"description": "test"},
			expectedErr: "issue analysis request must include 'issue' field",
		},
		{
			name:        "bug fix missing bug_description",
			requestType: entities.RequestTypeBugFix,
			input:       map[string]interface{}{"description": "test"},
			expectedErr: "bug fix request must include 'bug_description' field",
		},
		{
			name:        "feature missing feature_description",
			requestType: entities.RequestTypeFeature,
			input:       map[string]interface{}{"description": "test"},
			expectedErr: "feature request must include 'feature_description' field",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Execute
			request, err := suite.service.CreateRequest(suite.ctx, "session-123", tc.requestType, tc.input)

			// Assert
			assert := assert.New(t)
			assert.Error(err)
			assert.Nil(request)
			assert.Contains(err.Error(), tc.expectedErr)
		})
	}
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestCreateRequest_RepositoryCreateError tests request creation with repository error
func (suite *RequestServiceTestSuite) TestCreateRequest_RepositoryCreateError() {
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, "session-123").Return(suite.activeSession, nil)
	suite.mockRequestRepo.On("Create", suite.ctx, mock.AnythingOfType("*entities.Request")).Return(errors.New("database error"))

	// Execute
	request, err := suite.service.CreateRequest(suite.ctx, "session-123", entities.RequestTypeCodeReview, suite.validCodeReviewInput)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Nil(request)
	assert.Contains(err.Error(), "database error")
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
	suite.mockRequestRepo.AssertExpectations(suite.T())
}

// TestStartRequest_Success tests successful request starting
func (suite *RequestServiceTestSuite) TestStartRequest_Success() {
	// Create pending request
	pendingRequest := entities.NewRequest("session-123", entities.RequestTypeCodeReview, suite.validCodeReviewInput)
	
	// Setup mocks
	suite.mockRequestRepo.On("GetByID", suite.ctx, "request-123").Return(pendingRequest, nil)
	suite.mockRequestRepo.On("Update", suite.ctx, mock.AnythingOfType("*entities.Request")).Return(nil)

	// Execute
	err := suite.service.StartRequest(suite.ctx, "request-123")

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.Equal(entities.RequestStatusProcessing, pendingRequest.Status)
	assert.NotNil(pendingRequest.StartedAt)
	
	// Verify mock expectations
	suite.mockRequestRepo.AssertExpectations(suite.T())
}

// TestStartRequest_RequestNotFound tests starting non-existent request
func (suite *RequestServiceTestSuite) TestStartRequest_RequestNotFound() {
	// Setup mocks
	suite.mockRequestRepo.On("GetByID", suite.ctx, "invalid-request").Return(nil, errors.New("request not found"))

	// Execute
	err := suite.service.StartRequest(suite.ctx, "invalid-request")

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Contains(err.Error(), "request not found")
	
	// Verify mock expectations
	suite.mockRequestRepo.AssertExpectations(suite.T())
}

// TestStartRequest_RequestNil tests starting when request is nil
func (suite *RequestServiceTestSuite) TestStartRequest_RequestNil() {
	// Setup mocks
	suite.mockRequestRepo.On("GetByID", suite.ctx, "null-request").Return((*entities.Request)(nil), nil)

	// Execute
	err := suite.service.StartRequest(suite.ctx, "null-request")

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Equal("request not found", err.Error())
	
	// Verify mock expectations
	suite.mockRequestRepo.AssertExpectations(suite.T())
}

// TestStartRequest_RequestNotPending tests starting non-pending request
func (suite *RequestServiceTestSuite) TestStartRequest_RequestNotPending() {
	// Create processing request
	processingRequest := entities.NewRequest("session-123", entities.RequestTypeCodeReview, suite.validCodeReviewInput)
	processingRequest.Start()
	
	// Setup mocks
	suite.mockRequestRepo.On("GetByID", suite.ctx, "request-123").Return(processingRequest, nil)

	// Execute
	err := suite.service.StartRequest(suite.ctx, "request-123")

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Equal("request is not in pending status", err.Error())
	
	// Verify mock expectations
	suite.mockRequestRepo.AssertExpectations(suite.T())
}

// TestCompleteRequest_Success tests successful request completion
func (suite *RequestServiceTestSuite) TestCompleteRequest_Success() {
	// Create processing request
	processingRequest := entities.NewRequest("session-123", entities.RequestTypeCodeReview, suite.validCodeReviewInput)
	processingRequest.Start()
	
	output := map[string]interface{}{
		"result":     "Code review completed",
		"score":      95,
		"suggestions": []string{"Add error handling", "Improve documentation"},
	}
	
	// Setup mocks
	suite.mockRequestRepo.On("GetByID", suite.ctx, "request-123").Return(processingRequest, nil)
	suite.mockRequestRepo.On("Update", suite.ctx, mock.AnythingOfType("*entities.Request")).Return(nil)

	// Execute
	err := suite.service.CompleteRequest(suite.ctx, "request-123", output)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.Equal(entities.RequestStatusCompleted, processingRequest.Status)
	assert.Equal(output, processingRequest.Output)
	assert.NotNil(processingRequest.CompletedAt)
	assert.True(processingRequest.IsCompleted())
	
	// Verify mock expectations
	suite.mockRequestRepo.AssertExpectations(suite.T())
}

// TestCompleteRequest_RequestNotProcessing tests completing non-processing request
func (suite *RequestServiceTestSuite) TestCompleteRequest_RequestNotProcessing() {
	// Create pending request
	pendingRequest := entities.NewRequest("session-123", entities.RequestTypeCodeReview, suite.validCodeReviewInput)
	
	output := map[string]interface{}{"result": "test"}
	
	// Setup mocks
	suite.mockRequestRepo.On("GetByID", suite.ctx, "request-123").Return(pendingRequest, nil)

	// Execute
	err := suite.service.CompleteRequest(suite.ctx, "request-123", output)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Equal("request is not in processing status", err.Error())
	
	// Verify mock expectations
	suite.mockRequestRepo.AssertExpectations(suite.T())
}

// TestFailRequest_Success tests successful request failure
func (suite *RequestServiceTestSuite) TestFailRequest_Success() {
	// Create processing request
	processingRequest := entities.NewRequest("session-123", entities.RequestTypeCodeReview, suite.validCodeReviewInput)
	processingRequest.Start()
	
	errorMsg := "Processing failed due to invalid code syntax"
	
	// Setup mocks
	suite.mockRequestRepo.On("GetByID", suite.ctx, "request-123").Return(processingRequest, nil)
	suite.mockRequestRepo.On("Update", suite.ctx, mock.AnythingOfType("*entities.Request")).Return(nil)

	// Execute
	err := suite.service.FailRequest(suite.ctx, "request-123", errorMsg)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.Equal(entities.RequestStatusFailed, processingRequest.Status)
	assert.Equal(errorMsg, processingRequest.Error)
	assert.NotNil(processingRequest.CompletedAt)
	assert.True(processingRequest.IsCompleted())
	
	// Verify mock expectations
	suite.mockRequestRepo.AssertExpectations(suite.T())
}

// TestFailRequest_RequestAlreadyCompleted tests failing already completed request
func (suite *RequestServiceTestSuite) TestFailRequest_RequestAlreadyCompleted() {
	// Create completed request
	completedRequest := entities.NewRequest("session-123", entities.RequestTypeCodeReview, suite.validCodeReviewInput)
	completedRequest.Start()
	completedRequest.Complete(map[string]interface{}{"result": "done"})
	
	// Setup mocks
	suite.mockRequestRepo.On("GetByID", suite.ctx, "request-123").Return(completedRequest, nil)

	// Execute
	err := suite.service.FailRequest(suite.ctx, "request-123", "error message")

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Equal("request is already completed", err.Error())
	
	// Verify mock expectations
	suite.mockRequestRepo.AssertExpectations(suite.T())
}

// TestCancelRequest_Success tests successful request cancellation
func (suite *RequestServiceTestSuite) TestCancelRequest_Success() {
	// Create pending request
	pendingRequest := entities.NewRequest("session-123", entities.RequestTypeCodeReview, suite.validCodeReviewInput)
	
	// Setup mocks
	suite.mockRequestRepo.On("GetByID", suite.ctx, "request-123").Return(pendingRequest, nil)
	suite.mockRequestRepo.On("Update", suite.ctx, mock.AnythingOfType("*entities.Request")).Return(nil)

	// Execute
	err := suite.service.CancelRequest(suite.ctx, "request-123")

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.Equal(entities.RequestStatusCancelled, pendingRequest.Status)
	assert.NotNil(pendingRequest.CompletedAt)
	assert.True(pendingRequest.IsCompleted())
	
	// Verify mock expectations
	suite.mockRequestRepo.AssertExpectations(suite.T())
}

// TestGetPendingRequests_Success tests getting pending requests
func (suite *RequestServiceTestSuite) TestGetPendingRequests_Success() {
	// Create test data
	pendingRequests := []*entities.Request{
		entities.NewRequest("session-1", entities.RequestTypeCodeReview, suite.validCodeReviewInput),
		entities.NewRequest("session-2", entities.RequestTypeIssueAnalysis, map[string]interface{}{"issue": "test issue"}),
	}
	
	// Setup mocks
	suite.mockRequestRepo.On("GetPendingRequests", suite.ctx).Return(pendingRequests, nil)

	// Execute
	requests, err := suite.service.GetPendingRequests(suite.ctx)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.Equal(pendingRequests, requests)
	assert.Len(requests, 2)
	
	// Verify mock expectations
	suite.mockRequestRepo.AssertExpectations(suite.T())
}

// TestCheckForTimeouts_Success tests checking for timed out requests
func (suite *RequestServiceTestSuite) TestCheckForTimeouts_Success() {
	// Create test data
	timeoutDuration := 30 * time.Minute
	timedOutRequests := []*entities.Request{
		entities.NewRequest("session-1", entities.RequestTypeCodeReview, suite.validCodeReviewInput),
	}
	
	// Setup mocks
	suite.mockRequestRepo.On("GetRequestsWithTimeout", suite.ctx, timeoutDuration).Return(timedOutRequests, nil)

	// Execute
	requests, err := suite.service.CheckForTimeouts(suite.ctx, timeoutDuration)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.Equal(timedOutRequests, requests)
	assert.Len(requests, 1)
	
	// Verify mock expectations
	suite.mockRequestRepo.AssertExpectations(suite.T())
}

// TestTimeoutRequests_Success tests timing out multiple requests
func (suite *RequestServiceTestSuite) TestTimeoutRequests_Success() {
	// Create test data
	request1 := entities.NewRequest("session-1", entities.RequestTypeCodeReview, suite.validCodeReviewInput)
	request1.Start()
	request2 := entities.NewRequest("session-2", entities.RequestTypeIssueAnalysis, map[string]interface{}{"issue": "test"})
	request2.Start()
	
	requestsToTimeout := []*entities.Request{request1, request2}
	
	// Setup mocks
	suite.mockRequestRepo.On("Update", suite.ctx, request1).Return(nil)
	suite.mockRequestRepo.On("Update", suite.ctx, request2).Return(nil)

	// Execute
	err := suite.service.TimeoutRequests(suite.ctx, requestsToTimeout)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.Equal(entities.RequestStatusTimeout, request1.Status)
	assert.Equal(entities.RequestStatusTimeout, request2.Status)
	assert.Equal("Request processing timed out", request1.Error)
	assert.Equal("Request processing timed out", request2.Error)
	
	// Verify mock expectations
	suite.mockRequestRepo.AssertExpectations(suite.T())
}

// TestTimeoutRequests_PartialFailure tests timing out requests with partial failure
func (suite *RequestServiceTestSuite) TestTimeoutRequests_PartialFailure() {
	// Create test data
	request1 := entities.NewRequest("session-1", entities.RequestTypeCodeReview, suite.validCodeReviewInput)
	request1.Start()
	request2 := entities.NewRequest("session-2", entities.RequestTypeIssueAnalysis, map[string]interface{}{"issue": "test"})
	request2.Start()
	
	requestsToTimeout := []*entities.Request{request1, request2}
	
	// Setup mocks - first succeeds, second fails
	suite.mockRequestRepo.On("Update", suite.ctx, request1).Return(nil)
	suite.mockRequestRepo.On("Update", suite.ctx, request2).Return(errors.New("database error"))

	// Execute
	err := suite.service.TimeoutRequests(suite.ctx, requestsToTimeout)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Contains(err.Error(), "database error")
	
	// Verify mock expectations
	suite.mockRequestRepo.AssertExpectations(suite.T())
}

// TestValidateRequestInput tests the validateRequestInput private method through CreateRequest
func (suite *RequestServiceTestSuite) TestValidateRequestInput() {
	// Setup mocks for session validation
	suite.mockSessionRepo.On("GetByID", suite.ctx, "session-123").Return(suite.activeSession, nil)

	testCases := []struct {
		name        string
		requestType entities.RequestType
		input       map[string]interface{}
		shouldPass  bool
	}{
		{
			name:        "valid code review",
			requestType: entities.RequestTypeCodeReview,
			input:       map[string]interface{}{"code": "function test() {}"},
			shouldPass:  true,
		},
		{
			name:        "valid issue analysis",
			requestType: entities.RequestTypeIssueAnalysis,
			input:       map[string]interface{}{"issue": "Bug description"},
			shouldPass:  true,
		},
		{
			name:        "valid bug fix",
			requestType: entities.RequestTypeBugFix,
			input:       map[string]interface{}{"bug_description": "Fix description"},
			shouldPass:  true,
		},
		{
			name:        "valid feature",
			requestType: entities.RequestTypeFeature,
			input:       map[string]interface{}{"feature_description": "Feature description"},
			shouldPass:  true,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			if tc.shouldPass {
				// Setup mock for successful creation
				suite.mockRequestRepo.On("Create", suite.ctx, mock.AnythingOfType("*entities.Request")).Return(nil).Once()
			}
			
			// Execute
			request, err := suite.service.CreateRequest(suite.ctx, "session-123", tc.requestType, tc.input)

			// Assert
			assert := assert.New(t)
			if tc.shouldPass {
				assert.NoError(err)
				assert.NotNil(request)
			} else {
				assert.Error(err)
				assert.Nil(request)
			}
		})
	}
}

// BenchmarkCreateRequest benchmarks request creation
func BenchmarkCreateRequest(b *testing.B) {
	mockRequestRepo := &MockRequestRepository{}
	mockSessionRepo := &MockSessionRepository{}
	service := NewRequestService(mockRequestRepo, mockSessionRepo)
	ctx := context.Background()
	
	activeSession := entities.NewSession("user-123")
	input := map[string]interface{}{"code": "function test() { return true; }"}
	
	// Setup mocks
	mockSessionRepo.On("GetByID", ctx, "session-123").Return(activeSession, nil)
	mockRequestRepo.On("Create", ctx, mock.AnythingOfType("*entities.Request")).Return(nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CreateRequest(ctx, "session-123", entities.RequestTypeCodeReview, input)
	}
}

// BenchmarkStartRequest benchmarks request starting
func BenchmarkStartRequest(b *testing.B) {
	mockRequestRepo := &MockRequestRepository{}
	mockSessionRepo := &MockSessionRepository{}
	service := NewRequestService(mockRequestRepo, mockSessionRepo)
	ctx := context.Background()
	
	pendingRequest := entities.NewRequest("session-123", entities.RequestTypeCodeReview, map[string]interface{}{"code": "test"})
	
	// Setup mocks
	mockRequestRepo.On("GetByID", ctx, "request-123").Return(pendingRequest, nil)
	mockRequestRepo.On("Update", ctx, mock.AnythingOfType("*entities.Request")).Return(nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset request status for each iteration
		pendingRequest.Status = entities.RequestStatusPending
		service.StartRequest(ctx, "request-123")
	}
}