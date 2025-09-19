package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// RequestTestSuite defines the test suite for Request entity
type RequestTestSuite struct {
	suite.Suite
	validInput map[string]interface{}
}

// SetupSuite runs before all tests in the suite
func (suite *RequestTestSuite) SetupSuite() {
	suite.validInput = map[string]interface{}{
		"code":        "function test() { return true; }",
		"description": "Test function for validation",
		"language":    "javascript",
	}
}

// TestRequestTestSuite runs the test suite
func TestRequestTestSuite(t *testing.T) {
	suite.Run(t, new(RequestTestSuite))
}

// TestNewRequest tests the NewRequest constructor
func (suite *RequestTestSuite) TestNewRequest() {
	sessionID := "session-123"
	requestType := RequestTypeCodeReview

	request := NewRequest(sessionID, requestType, suite.validInput)

	// Test basic properties
	assert := assert.New(suite.T())
	assert.NotEmpty(request.ID)
	assert.Equal(sessionID, request.SessionID)
	assert.Equal(requestType, request.Type)
	assert.Equal(RequestStatusPending, request.Status)
	assert.Equal(suite.validInput, request.Input)
	assert.NotNil(request.Output)
	assert.Empty(request.Error)
	assert.Zero(request.ProcessingTimeMs)
	assert.False(request.CreatedAt.IsZero())
	assert.False(request.UpdatedAt.IsZero())
	assert.Nil(request.StartedAt)
	assert.Nil(request.CompletedAt)
	assert.True(request.IsValid())
}

// TestRequestStart tests the Start method
func (suite *RequestTestSuite) TestRequestStart() {
	request := NewRequest("session-123", RequestTypeCodeReview, suite.validInput)
	initialCreatedAt := request.CreatedAt
	
	// Wait a small amount to ensure time difference
	time.Sleep(1 * time.Millisecond)
	
	request.Start()

	assert := assert.New(suite.T())
	assert.Equal(RequestStatusProcessing, request.Status)
	assert.NotNil(request.StartedAt)
	assert.True(request.UpdatedAt.After(initialCreatedAt))
	assert.False(request.IsCompleted())
}

// TestRequestComplete tests the Complete method
func (suite *RequestTestSuite) TestRequestComplete() {
	request := NewRequest("session-123", RequestTypeCodeReview, suite.validInput)
	request.Start()
	
	// Wait to ensure processing time calculation
	time.Sleep(5 * time.Millisecond)
	
	output := map[string]interface{}{
		"result":     "Code looks good",
		"score":      95,
		"suggestions": []string{"Add error handling", "Improve documentation"},
	}

	request.Complete(output)

	assert := assert.New(suite.T())
	assert.Equal(RequestStatusCompleted, request.Status)
	assert.Equal(output, request.Output)
	assert.Empty(request.Error)
	assert.NotNil(request.CompletedAt)
	assert.True(request.ProcessingTimeMs > 0)
	assert.True(request.IsCompleted())
}

// TestRequestFail tests the Fail method
func (suite *RequestTestSuite) TestRequestFail() {
	request := NewRequest("session-123", RequestTypeCodeReview, suite.validInput)
	request.Start()
	
	// Wait to ensure processing time calculation
	time.Sleep(5 * time.Millisecond)
	
	errorMsg := "Failed to process code: syntax error"
	request.Fail(errorMsg)

	assert := assert.New(suite.T())
	assert.Equal(RequestStatusFailed, request.Status)
	assert.Equal(errorMsg, request.Error)
	assert.NotNil(request.CompletedAt)
	assert.True(request.ProcessingTimeMs > 0)
	assert.True(request.IsCompleted())
}

// TestRequestCancel tests the Cancel method
func (suite *RequestTestSuite) TestRequestCancel() {
	request := NewRequest("session-123", RequestTypeCodeReview, suite.validInput)
	
	request.Cancel()

	assert := assert.New(suite.T())
	assert.Equal(RequestStatusCancelled, request.Status)
	assert.NotNil(request.CompletedAt)
	assert.Zero(request.ProcessingTimeMs) // No processing time if not started
	assert.True(request.IsCompleted())
}

// TestRequestTimeout tests the Timeout method
func (suite *RequestTestSuite) TestRequestTimeout() {
	request := NewRequest("session-123", RequestTypeCodeReview, suite.validInput)
	request.Start()
	
	// Wait to ensure processing time calculation
	time.Sleep(5 * time.Millisecond)
	
	request.Timeout()

	assert := assert.New(suite.T())
	assert.Equal(RequestStatusTimeout, request.Status)
	assert.Equal("Request processing timed out", request.Error)
	assert.NotNil(request.CompletedAt)
	assert.True(request.ProcessingTimeMs > 0)
	assert.True(request.IsCompleted())
}

// TestRequestIsCompleted tests the IsCompleted method
func (suite *RequestTestSuite) TestRequestIsCompleted() {
	assert := assert.New(suite.T())
	
	// Test non-completed statuses
	pendingRequest := NewRequest("session-123", RequestTypeCodeReview, suite.validInput)
	assert.False(pendingRequest.IsCompleted())
	
	processingRequest := NewRequest("session-123", RequestTypeCodeReview, suite.validInput)
	processingRequest.Start()
	assert.False(processingRequest.IsCompleted())
	
	// Test completed statuses
	completedRequest := NewRequest("session-123", RequestTypeCodeReview, suite.validInput)
	completedRequest.Complete(map[string]interface{}{"result": "done"})
	assert.True(completedRequest.IsCompleted())
	
	failedRequest := NewRequest("session-123", RequestTypeCodeReview, suite.validInput)
	failedRequest.Fail("error")
	assert.True(failedRequest.IsCompleted())
	
	cancelledRequest := NewRequest("session-123", RequestTypeCodeReview, suite.validInput)
	cancelledRequest.Cancel()
	assert.True(cancelledRequest.IsCompleted())
	
	timeoutRequest := NewRequest("session-123", RequestTypeCodeReview, suite.validInput)
	timeoutRequest.Timeout()
	assert.True(timeoutRequest.IsCompleted())
}

// TestRequestIsValid tests the IsValid method
func (suite *RequestTestSuite) TestRequestIsValid() {
	assert := assert.New(suite.T())
	
	// Valid request
	validRequest := NewRequest("session-123", RequestTypeCodeReview, suite.validInput)
	assert.True(validRequest.IsValid())
	
	// Invalid requests
	invalidRequests := []*Request{
		{}, // Empty request
		{ID: "", SessionID: "session-123", Type: RequestTypeCodeReview, Status: RequestStatusPending, Input: suite.validInput, CreatedAt: time.Now()}, // Empty ID
		{ID: "req-123", SessionID: "", Type: RequestTypeCodeReview, Status: RequestStatusPending, Input: suite.validInput, CreatedAt: time.Now()}, // Empty SessionID
		{ID: "req-123", SessionID: "session-123", Type: "", Status: RequestStatusPending, Input: suite.validInput, CreatedAt: time.Now()}, // Empty Type
		{ID: "req-123", SessionID: "session-123", Type: RequestTypeCodeReview, Status: "", Input: suite.validInput, CreatedAt: time.Now()}, // Empty Status
		{ID: "req-123", SessionID: "session-123", Type: RequestTypeCodeReview, Status: RequestStatusPending, Input: nil, CreatedAt: time.Now()}, // Nil Input
		{ID: "req-123", SessionID: "session-123", Type: RequestTypeCodeReview, Status: RequestStatusPending, Input: suite.validInput, CreatedAt: time.Time{}}, // Zero CreatedAt
	}
	
	for i, invalidRequest := range invalidRequests {
		assert.False(invalidRequest.IsValid(), "Invalid request %d should be invalid", i)
	}
}

// TestRequestTypes tests all request types are valid
func (suite *RequestTestSuite) TestRequestTypes() {
	assert := assert.New(suite.T())
	
	types := []RequestType{
		RequestTypeCodeReview,
		RequestTypeIssueAnalysis,
		RequestTypeBugFix,
		RequestTypeFeature,
	}
	
	for _, reqType := range types {
		request := NewRequest("session-123", reqType, suite.validInput)
		assert.Equal(reqType, request.Type)
		assert.True(request.IsValid())
	}
}

// TestRequestStatuses tests all request statuses are valid
func (suite *RequestTestSuite) TestRequestStatuses() {
	assert := assert.New(suite.T())
	
	statuses := []RequestStatus{
		RequestStatusPending,
		RequestStatusProcessing,
		RequestStatusCompleted,
		RequestStatusFailed,
		RequestStatusCancelled,
		RequestStatusTimeout,
	}
	
	for _, status := range statuses {
		request := NewRequest("session-123", RequestTypeCodeReview, suite.validInput)
		request.Status = status
		assert.Equal(status, request.Status)
	}
}

// TestRequestProcessingTime tests processing time calculation
func (suite *RequestTestSuite) TestRequestProcessingTime() {
	assert := assert.New(suite.T())
	
	request := NewRequest("session-123", RequestTypeCodeReview, suite.validInput)
	
	// Before starting - no processing time
	request.Complete(map[string]interface{}{"result": "done"})
	assert.Zero(request.ProcessingTimeMs)
	
	// After starting - should have processing time
	request2 := NewRequest("session-123", RequestTypeCodeReview, suite.validInput)
	request2.Start()
	time.Sleep(10 * time.Millisecond)
	request2.Complete(map[string]interface{}{"result": "done"})
	assert.True(request2.ProcessingTimeMs >= 10)
}

// TestGenerateID tests ID generation
func (suite *RequestTestSuite) TestGenerateID() {
	assert := assert.New(suite.T())
	
	// Generate multiple IDs
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateID()
		assert.NotEmpty(id)
		assert.NotContains(ids, id, "ID should be unique")
		ids[id] = true
	}
}

// BenchmarkNewRequest benchmarks request creation
func BenchmarkNewRequest(b *testing.B) {
	input := map[string]interface{}{
		"code": "function test() { return true; }",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewRequest("session-123", RequestTypeCodeReview, input)
	}
}

// BenchmarkRequestStateTransitions benchmarks state transitions
func BenchmarkRequestStateTransitions(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := NewRequest("session-123", RequestTypeCodeReview, map[string]interface{}{"code": "test"})
		request.Start()
		request.Complete(map[string]interface{}{"result": "done"})
	}
}