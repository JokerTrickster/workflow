package services

import (
	"context"
	"errors"
	"local-backend-server/internal/domain/entities"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// SessionServiceTestSuite defines the test suite for SessionService
type SessionServiceTestSuite struct {
	suite.Suite
	service         *SessionService
	mockSessionRepo *MockSessionRepository
	ctx             context.Context
}

// SetupTest runs before each test
func (suite *SessionServiceTestSuite) SetupTest() {
	suite.mockSessionRepo = &MockSessionRepository{}
	suite.service = NewSessionService(suite.mockSessionRepo)
	suite.ctx = context.Background()
}

// TestSessionServiceTestSuite runs the test suite
func TestSessionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SessionServiceTestSuite))
}

// TestNewSessionService tests the NewSessionService constructor
func (suite *SessionServiceTestSuite) TestNewSessionService() {
	service := NewSessionService(suite.mockSessionRepo)
	
	assert := assert.New(suite.T())
	assert.NotNil(service)
	assert.Equal(suite.mockSessionRepo, service.sessionRepo)
}

// TestCreateSession_Success tests successful session creation
func (suite *SessionServiceTestSuite) TestCreateSession_Success() {
	userID := "user-123"
	
	// Setup mocks
	suite.mockSessionRepo.On("Create", suite.ctx, mock.AnythingOfType("*entities.Session")).Return(nil)

	// Execute
	session, err := suite.service.CreateSession(suite.ctx, userID)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.NotNil(session)
	assert.Equal(userID, session.UserID)
	assert.Equal(entities.SessionStatusActive, session.Status)
	assert.NotNil(session.ExpiresAt)
	assert.True(session.IsValid())
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestCreateSession_WithEmptyUserID tests session creation with empty user ID
func (suite *SessionServiceTestSuite) TestCreateSession_WithEmptyUserID() {
	userID := ""
	
	// Setup mocks
	suite.mockSessionRepo.On("Create", suite.ctx, mock.AnythingOfType("*entities.Session")).Return(nil)

	// Execute
	session, err := suite.service.CreateSession(suite.ctx, userID)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.NotNil(session)
	assert.Equal(userID, session.UserID)
	assert.Equal(entities.SessionStatusActive, session.Status)
	assert.True(session.IsValid()) // UserID is optional for validation
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestCreateSession_RepositoryError tests session creation with repository error
func (suite *SessionServiceTestSuite) TestCreateSession_RepositoryError() {
	userID := "user-123"
	
	// Setup mocks
	suite.mockSessionRepo.On("Create", suite.ctx, mock.AnythingOfType("*entities.Session")).Return(errors.New("database error"))

	// Execute
	session, err := suite.service.CreateSession(suite.ctx, userID)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Nil(session)
	assert.Contains(err.Error(), "database error")
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestGetSession_Success tests successful session retrieval
func (suite *SessionServiceTestSuite) TestGetSession_Success() {
	sessionID := "session-123"
	
	// Create test session
	testSession := entities.NewSession("user-123")
	testSession.ID = sessionID
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(testSession, nil)

	// Execute
	session, err := suite.service.GetSession(suite.ctx, sessionID)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.NotNil(session)
	assert.Equal(sessionID, session.ID)
	assert.Equal(entities.SessionStatusActive, session.Status)
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestGetSession_NotFound tests session retrieval with non-existent session
func (suite *SessionServiceTestSuite) TestGetSession_NotFound() {
	sessionID := "invalid-session"
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(nil, errors.New("session not found"))

	// Execute
	session, err := suite.service.GetSession(suite.ctx, sessionID)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Nil(session)
	assert.Contains(err.Error(), "session not found")
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestGetSession_SessionNil tests session retrieval when session is nil
func (suite *SessionServiceTestSuite) TestGetSession_SessionNil() {
	sessionID := "null-session"
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(nil, nil)

	// Execute
	session, err := suite.service.GetSession(suite.ctx, sessionID)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Nil(session)
	assert.Equal("session not found", err.Error())
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestGetSession_ExpiredSession tests session retrieval with expired session
func (suite *SessionServiceTestSuite) TestGetSession_ExpiredSession() {
	sessionID := "expired-session"
	
	// Create expired session
	expiredSession := entities.NewSession("user-123")
	expiredSession.ID = sessionID
	pastTime := time.Now().Add(-1 * time.Hour)
	expiredSession.ExpiresAt = &pastTime
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(expiredSession, nil)
	suite.mockSessionRepo.On("Update", suite.ctx, mock.AnythingOfType("*entities.Session")).Return(nil)

	// Execute
	session, err := suite.service.GetSession(suite.ctx, sessionID)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Nil(session)
	assert.Equal("session has expired", err.Error())
	assert.Equal(entities.SessionStatusExpired, expiredSession.Status)
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestGetActiveSession_Success tests successful active session retrieval
func (suite *SessionServiceTestSuite) TestGetActiveSession_Success() {
	userID := "user-123"
	
	// Create test sessions with different update times
	session1 := entities.NewSession(userID)
	session1.ID = "session-1"
	session1.UpdatedAt = time.Now().Add(-1 * time.Hour)
	
	session2 := entities.NewSession(userID)
	session2.ID = "session-2"
	session2.UpdatedAt = time.Now().Add(-30 * time.Minute) // More recent
	
	activeSessions := []*entities.Session{session1, session2}
	
	// Setup mocks
	suite.mockSessionRepo.On("GetActiveByUserID", suite.ctx, userID).Return(activeSessions, nil)

	// Execute
	session, err := suite.service.GetActiveSession(suite.ctx, userID)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.NotNil(session)
	assert.Equal("session-2", session.ID) // Should return the most recent one
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestGetActiveSession_NoActiveSession tests active session retrieval with no active sessions
func (suite *SessionServiceTestSuite) TestGetActiveSession_NoActiveSession() {
	userID := "user-123"
	
	// Setup mocks
	suite.mockSessionRepo.On("GetActiveByUserID", suite.ctx, userID).Return([]*entities.Session{}, nil)

	// Execute
	session, err := suite.service.GetActiveSession(suite.ctx, userID)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Nil(session)
	assert.Equal("no active session found", err.Error())
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestGetActiveSession_AllExpired tests active session retrieval with all sessions expired
func (suite *SessionServiceTestSuite) TestGetActiveSession_AllExpired() {
	userID := "user-123"
	
	// Create expired sessions
	expiredSession1 := entities.NewSession(userID)
	pastTime1 := time.Now().Add(-2 * time.Hour)
	expiredSession1.ExpiresAt = &pastTime1
	
	expiredSession2 := entities.NewSession(userID)
	pastTime2 := time.Now().Add(-1 * time.Hour)
	expiredSession2.ExpiresAt = &pastTime2
	
	expiredSessions := []*entities.Session{expiredSession1, expiredSession2}
	
	// Setup mocks
	suite.mockSessionRepo.On("GetActiveByUserID", suite.ctx, userID).Return(expiredSessions, nil)

	// Execute
	session, err := suite.service.GetActiveSession(suite.ctx, userID)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Nil(session)
	assert.Equal("no active session found", err.Error())
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestExtendSession_Success tests successful session extension
func (suite *SessionServiceTestSuite) TestExtendSession_Success() {
	sessionID := "session-123"
	extension := 2 * time.Hour
	
	// Create test session
	testSession := entities.NewSession("user-123")
	testSession.ID = sessionID
	originalExpiration := *testSession.ExpiresAt
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(testSession, nil)
	suite.mockSessionRepo.On("Update", suite.ctx, mock.AnythingOfType("*entities.Session")).Return(nil)

	// Execute
	err := suite.service.ExtendSession(suite.ctx, sessionID, extension)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.True(testSession.ExpiresAt.After(originalExpiration))
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestExtendSession_SessionNotFound tests extending non-existent session
func (suite *SessionServiceTestSuite) TestExtendSession_SessionNotFound() {
	sessionID := "invalid-session"
	extension := 2 * time.Hour
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(nil, errors.New("session not found"))

	// Execute
	err := suite.service.ExtendSession(suite.ctx, sessionID, extension)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Contains(err.Error(), "session not found")
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestExtendSession_SessionNil tests extending when session is nil
func (suite *SessionServiceTestSuite) TestExtendSession_SessionNil() {
	sessionID := "null-session"
	extension := 2 * time.Hour
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(nil, nil)

	// Execute
	err := suite.service.ExtendSession(suite.ctx, sessionID, extension)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Equal("session not found", err.Error())
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestExtendSession_SessionNotActive tests extending inactive session
func (suite *SessionServiceTestSuite) TestExtendSession_SessionNotActive() {
	sessionID := "inactive-session"
	extension := 2 * time.Hour
	
	// Create inactive session
	inactiveSession := entities.NewSession("user-123")
	inactiveSession.ID = sessionID
	inactiveSession.Deactivate()
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(inactiveSession, nil)

	// Execute
	err := suite.service.ExtendSession(suite.ctx, sessionID, extension)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Equal("session is not active", err.Error())
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestDeactivateSession_Success tests successful session deactivation
func (suite *SessionServiceTestSuite) TestDeactivateSession_Success() {
	sessionID := "session-123"
	
	// Create test session
	testSession := entities.NewSession("user-123")
	testSession.ID = sessionID
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(testSession, nil)
	suite.mockSessionRepo.On("Update", suite.ctx, mock.AnythingOfType("*entities.Session")).Return(nil)

	// Execute
	err := suite.service.DeactivateSession(suite.ctx, sessionID)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.Equal(entities.SessionStatusInactive, testSession.Status)
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestCleanupExpiredSessions_Success tests successful cleanup of expired sessions
func (suite *SessionServiceTestSuite) TestCleanupExpiredSessions_Success() {
	cleanedCount := 5
	
	// Setup mocks
	suite.mockSessionRepo.On("CleanupExpiredSessions", suite.ctx).Return(cleanedCount, nil)

	// Execute
	count, err := suite.service.CleanupExpiredSessions(suite.ctx)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.Equal(cleanedCount, count)
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestAddSessionMetadata_Success tests adding metadata to a session
func (suite *SessionServiceTestSuite) TestAddSessionMetadata_Success() {
	sessionID := "session-123"
	key := "client_type"
	value := "mobile"
	
	// Create test session
	testSession := entities.NewSession("user-123")
	testSession.ID = sessionID
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(testSession, nil)
	suite.mockSessionRepo.On("Update", suite.ctx, mock.AnythingOfType("*entities.Session")).Return(nil)

	// Execute
	err := suite.service.AddSessionMetadata(suite.ctx, sessionID, key, value)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	
	// Verify metadata was added
	actualValue, exists := testSession.Metadata[key]
	assert.True(exists)
	assert.Equal(value, actualValue)
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestAddSessionMetadata_SessionNotFound tests adding metadata to non-existent session
func (suite *SessionServiceTestSuite) TestAddSessionMetadata_SessionNotFound() {
	sessionID := "invalid-session"
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(nil, errors.New("session not found"))

	// Execute
	err := suite.service.AddSessionMetadata(suite.ctx, sessionID, "key", "value")

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Contains(err.Error(), "session not found")
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestValidateSessionAccess_Success tests successful session access validation
func (suite *SessionServiceTestSuite) TestValidateSessionAccess_Success() {
	sessionID := "session-123"
	userID := "user-123"
	
	// Create test session
	testSession := entities.NewSession(userID)
	testSession.ID = sessionID
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(testSession, nil)

	// Execute
	err := suite.service.ValidateSessionAccess(suite.ctx, sessionID, userID)

	// Assert
	assert := assert.New(suite.T())
	assert.NoError(err)
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestValidateSessionAccess_SessionNotFound tests access validation with non-existent session
func (suite *SessionServiceTestSuite) TestValidateSessionAccess_SessionNotFound() {
	sessionID := "invalid-session"
	userID := "user-123"
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(nil, errors.New("session not found"))

	// Execute
	err := suite.service.ValidateSessionAccess(suite.ctx, sessionID, userID)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Contains(err.Error(), "session not found")
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestValidateSessionAccess_SessionNil tests access validation when session is nil
func (suite *SessionServiceTestSuite) TestValidateSessionAccess_SessionNil() {
	sessionID := "null-session"
	userID := "user-123"
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(nil, nil)

	// Execute
	err := suite.service.ValidateSessionAccess(suite.ctx, sessionID, userID)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Equal("session not found", err.Error())
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestValidateSessionAccess_WrongUser tests access validation with wrong user
func (suite *SessionServiceTestSuite) TestValidateSessionAccess_WrongUser() {
	sessionID := "session-123"
	sessionUserID := "user-123"
	requestingUserID := "user-456"
	
	// Create test session
	testSession := entities.NewSession(sessionUserID)
	testSession.ID = sessionID
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(testSession, nil)

	// Execute
	err := suite.service.ValidateSessionAccess(suite.ctx, sessionID, requestingUserID)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Equal("access denied: session belongs to different user", err.Error())
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestValidateSessionAccess_ExpiredSession tests access validation with expired session
func (suite *SessionServiceTestSuite) TestValidateSessionAccess_ExpiredSession() {
	sessionID := "session-123"
	userID := "user-123"
	
	// Create expired session
	expiredSession := entities.NewSession(userID)
	expiredSession.ID = sessionID
	pastTime := time.Now().Add(-1 * time.Hour)
	expiredSession.ExpiresAt = &pastTime
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(expiredSession, nil)

	// Execute
	err := suite.service.ValidateSessionAccess(suite.ctx, sessionID, userID)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Equal("session has expired", err.Error())
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestValidateSessionAccess_InactiveSession tests access validation with inactive session
func (suite *SessionServiceTestSuite) TestValidateSessionAccess_InactiveSession() {
	sessionID := "session-123"
	userID := "user-123"
	
	// Create inactive session
	inactiveSession := entities.NewSession(userID)
	inactiveSession.ID = sessionID
	inactiveSession.Deactivate()
	
	// Setup mocks
	suite.mockSessionRepo.On("GetByID", suite.ctx, sessionID).Return(inactiveSession, nil)

	// Execute
	err := suite.service.ValidateSessionAccess(suite.ctx, sessionID, userID)

	// Assert
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Equal("session is not active", err.Error())
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// TestSessionService_CompleteWorkflow tests a complete session workflow
func (suite *SessionServiceTestSuite) TestSessionService_CompleteWorkflow() {
	userID := "workflow-user"
	
	// Create initial session
	createdSession := entities.NewSession(userID)
	createdSession.ID = "workflow-session"
	
	// Setup mocks for creating session
	suite.mockSessionRepo.On("Create", suite.ctx, mock.AnythingOfType("*entities.Session")).Return(nil)
	
	// Setup mocks for getting session
	suite.mockSessionRepo.On("GetByID", suite.ctx, createdSession.ID).Return(createdSession, nil).Times(3)
	
	// Setup mocks for adding metadata and extending session
	suite.mockSessionRepo.On("Update", suite.ctx, mock.AnythingOfType("*entities.Session")).Return(nil).Times(2)
	
	// Setup mocks for getting active session
	activeSessions := []*entities.Session{createdSession}
	suite.mockSessionRepo.On("GetActiveByUserID", suite.ctx, userID).Return(activeSessions, nil)
	
	// Setup mocks for validating access
	suite.mockSessionRepo.On("GetByID", suite.ctx, createdSession.ID).Return(createdSession, nil).Once()

	assert := assert.New(suite.T())
	
	// Create session
	session, err := suite.service.CreateSession(suite.ctx, userID)
	assert.NoError(err)
	assert.NotNil(session)
	assert.Equal(userID, session.UserID)
	
	// Get session
	retrievedSession, err := suite.service.GetSession(suite.ctx, createdSession.ID)
	assert.NoError(err)
	assert.Equal(createdSession.ID, retrievedSession.ID)
	
	// Add metadata
	err = suite.service.AddSessionMetadata(suite.ctx, createdSession.ID, "platform", "web")
	assert.NoError(err)
	
	// Extend session
	err = suite.service.ExtendSession(suite.ctx, createdSession.ID, 2*time.Hour)
	assert.NoError(err)
	
	// Get active session
	activeSession, err := suite.service.GetActiveSession(suite.ctx, userID)
	assert.NoError(err)
	assert.Equal(createdSession.ID, activeSession.ID)
	
	// Validate access
	err = suite.service.ValidateSessionAccess(suite.ctx, createdSession.ID, userID)
	assert.NoError(err)
	
	// Verify mock expectations
	suite.mockSessionRepo.AssertExpectations(suite.T())
}

// BenchmarkCreateSession benchmarks session creation
func BenchmarkCreateSession(b *testing.B) {
	mockSessionRepo := &MockSessionRepository{}
	service := NewSessionService(mockSessionRepo)
	ctx := context.Background()
	
	// Setup mocks
	mockSessionRepo.On("Create", ctx, mock.AnythingOfType("*entities.Session")).Return(nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CreateSession(ctx, "user-123")
	}
}

// BenchmarkValidateSessionAccess benchmarks session access validation
func BenchmarkValidateSessionAccess(b *testing.B) {
	mockSessionRepo := &MockSessionRepository{}
	service := NewSessionService(mockSessionRepo)
	ctx := context.Background()
	
	testSession := entities.NewSession("user-123")
	
	// Setup mocks
	mockSessionRepo.On("GetByID", ctx, "session-123").Return(testSession, nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ValidateSessionAccess(ctx, "session-123", "user-123")
	}
}