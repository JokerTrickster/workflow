package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// SessionTestSuite defines the test suite for Session entity
type SessionTestSuite struct {
	suite.Suite
}

// TestSessionTestSuite runs the test suite
func TestSessionTestSuite(t *testing.T) {
	suite.Run(t, new(SessionTestSuite))
}

// TestNewSession tests the NewSession constructor
func (suite *SessionTestSuite) TestNewSession() {
	userID := "user-123"
	before := time.Now()
	session := NewSession(userID)
	after := time.Now()

	assert := assert.New(suite.T())
	assert.NotEmpty(session.ID)
	assert.Equal(userID, session.UserID)
	assert.Equal(SessionStatusActive, session.Status)
	assert.NotNil(session.Metadata)
	assert.Empty(session.Metadata)
	assert.True(session.CreatedAt.After(before) || session.CreatedAt.Equal(before))
	assert.True(session.CreatedAt.Before(after) || session.CreatedAt.Equal(after))
	assert.Equal(session.CreatedAt, session.UpdatedAt)
	assert.NotNil(session.ExpiresAt)
	
	// Check expiration is set to 24 hours from creation
	expectedExpiration := session.CreatedAt.Add(24 * time.Hour)
	assert.True(session.ExpiresAt.After(expectedExpiration.Add(-1*time.Second)))
	assert.True(session.ExpiresAt.Before(expectedExpiration.Add(1*time.Second)))
	
	assert.True(session.IsValid())
	assert.False(session.IsExpired())
}

// TestNewSessionWithEmptyUserID tests creating session with empty user ID
func (suite *SessionTestSuite) TestNewSessionWithEmptyUserID() {
	session := NewSession("")

	assert := assert.New(suite.T())
	assert.NotEmpty(session.ID)
	assert.Empty(session.UserID)
	assert.Equal(SessionStatusActive, session.Status)
	assert.True(session.IsValid()) // UserID is optional for validation
}

// TestSessionAddMetadata tests the AddMetadata method
func (suite *SessionTestSuite) TestSessionAddMetadata() {
	session := NewSession("user-123")
	initialUpdatedAt := session.UpdatedAt
	
	// Wait to ensure time difference
	time.Sleep(1 * time.Millisecond)
	
	// Add various metadata types
	session.AddMetadata("client_type", "web")
	session.AddMetadata("user_preferences", map[string]interface{}{"theme": "dark"})
	session.AddMetadata("session_count", 5)
	session.AddMetadata("feature_flags", []string{"feature_a", "feature_b"})

	assert := assert.New(suite.T())
	assert.True(session.UpdatedAt.After(initialUpdatedAt))
	
	// Check metadata values
	clientType, exists := session.Metadata["client_type"]
	assert.True(exists)
	assert.Equal("web", clientType)
	
	userPrefs, exists := session.Metadata["user_preferences"]
	assert.True(exists)
	expectedPrefs := map[string]interface{}{"theme": "dark"}
	assert.Equal(expectedPrefs, userPrefs)
	
	sessionCount, exists := session.Metadata["session_count"]
	assert.True(exists)
	assert.Equal(5, sessionCount)
	
	featureFlags, exists := session.Metadata["feature_flags"]
	assert.True(exists)
	expectedFlags := []string{"feature_a", "feature_b"}
	assert.Equal(expectedFlags, featureFlags)
}

// TestSessionAddMetadataToNilMetadata tests adding metadata when Metadata is nil
func (suite *SessionTestSuite) TestSessionAddMetadataToNilMetadata() {
	session := &Session{
		ID:        "test-id",
		UserID:    "user-123",
		Status:    SessionStatusActive,
		Metadata:  nil, // Explicitly set to nil
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	session.AddMetadata("test_key", "test_value")

	assert := assert.New(suite.T())
	assert.NotNil(session.Metadata)
	
	value, exists := session.Metadata["test_key"]
	assert.True(exists)
	assert.Equal("test_value", value)
}

// TestSessionActivate tests the Activate method
func (suite *SessionTestSuite) TestSessionActivate() {
	session := NewSession("user-123")
	session.Status = SessionStatusInactive
	initialUpdatedAt := session.UpdatedAt
	
	// Wait to ensure time difference
	time.Sleep(1 * time.Millisecond)
	
	session.Activate()

	assert := assert.New(suite.T())
	assert.Equal(SessionStatusActive, session.Status)
	assert.True(session.UpdatedAt.After(initialUpdatedAt))
}

// TestSessionDeactivate tests the Deactivate method
func (suite *SessionTestSuite) TestSessionDeactivate() {
	session := NewSession("user-123")
	initialUpdatedAt := session.UpdatedAt
	
	// Wait to ensure time difference
	time.Sleep(1 * time.Millisecond)
	
	session.Deactivate()

	assert := assert.New(suite.T())
	assert.Equal(SessionStatusInactive, session.Status)
	assert.True(session.UpdatedAt.After(initialUpdatedAt))
}

// TestSessionExpire tests the Expire method
func (suite *SessionTestSuite) TestSessionExpire() {
	session := NewSession("user-123")
	initialUpdatedAt := session.UpdatedAt
	
	// Wait to ensure time difference
	time.Sleep(1 * time.Millisecond)
	
	session.Expire()

	assert := assert.New(suite.T())
	assert.Equal(SessionStatusExpired, session.Status)
	assert.True(session.UpdatedAt.After(initialUpdatedAt))
}

// TestSessionIsExpired tests the IsExpired method
func (suite *SessionTestSuite) TestSessionIsExpired() {
	assert := assert.New(suite.T())
	
	// Test session with future expiration
	futureSession := NewSession("user-123")
	assert.False(futureSession.IsExpired())
	
	// Test session with past expiration
	pastSession := NewSession("user-123")
	pastExpiration := time.Now().Add(-1 * time.Hour)
	pastSession.ExpiresAt = &pastExpiration
	assert.True(pastSession.IsExpired())
	
	// Test session with nil expiration
	nilExpirationSession := &Session{
		ID:        generateID(),
		UserID:    "user-123",
		Status:    SessionStatusActive,
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: nil,
	}
	assert.False(nilExpirationSession.IsExpired())
}

// TestSessionExtendExpiration tests the ExtendExpiration method
func (suite *SessionTestSuite) TestSessionExtendExpiration() {
	session := NewSession("user-123")
	initialExpiration := *session.ExpiresAt
	initialUpdatedAt := session.UpdatedAt
	
	// Wait to ensure time difference
	time.Sleep(1 * time.Millisecond)
	
	// Extend by 2 hours
	extension := 2 * time.Hour
	session.ExtendExpiration(extension)

	assert := assert.New(suite.T())
	assert.True(session.ExpiresAt.After(initialExpiration))
	assert.True(session.UpdatedAt.After(initialUpdatedAt))
	
	// Check the extension is approximately correct (within 1 second)
	expectedExpiration := time.Now().Add(extension)
	assert.True(session.ExpiresAt.After(expectedExpiration.Add(-1*time.Second)))
	assert.True(session.ExpiresAt.Before(expectedExpiration.Add(1*time.Second)))
}

// TestSessionIsValid tests the IsValid method
func (suite *SessionTestSuite) TestSessionIsValid() {
	assert := assert.New(suite.T())
	
	// Valid session
	validSession := NewSession("user-123")
	assert.True(validSession.IsValid())
	
	// Valid session without UserID (UserID is optional)
	validSessionNoUser := NewSession("")
	assert.True(validSessionNoUser.IsValid())
	
	// Invalid sessions
	invalidSessions := []*Session{
		{}, // Empty session
		{ID: "", UserID: "user-123", Status: SessionStatusActive, Metadata: make(map[string]interface{}), CreatedAt: time.Now()}, // Empty ID
		{ID: "session-123", UserID: "user-123", Status: "", Metadata: make(map[string]interface{}), CreatedAt: time.Now()}, // Empty Status
		{ID: "session-123", UserID: "user-123", Status: SessionStatusActive, Metadata: make(map[string]interface{}), CreatedAt: time.Time{}}, // Zero CreatedAt
	}
	
	for i, invalidSession := range invalidSessions {
		assert.False(invalidSession.IsValid(), "Invalid session %d should be invalid", i)
	}
}

// TestSessionStatuses tests all session statuses are valid
func (suite *SessionTestSuite) TestSessionStatuses() {
	assert := assert.New(suite.T())
	
	statuses := []SessionStatus{
		SessionStatusActive,
		SessionStatusInactive,
		SessionStatusExpired,
	}
	
	for _, status := range statuses {
		session := NewSession("user-123")
		session.Status = status
		assert.Equal(status, session.Status)
		assert.True(session.IsValid())
	}
}

// TestSessionMetadataTypes tests different metadata value types
func (suite *SessionTestSuite) TestSessionMetadataTypes() {
	session := NewSession("user-123")

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
		session.AddMetadata(key, expectedValue)
		actualValue, exists := session.Metadata[key]
		assert.True(exists, "Key %s should exist", key)
		assert.Equal(expectedValue, actualValue, "Value for key %s should match", key)
	}
}

// TestSessionLifecycle tests a complete session lifecycle
func (suite *SessionTestSuite) TestSessionLifecycle() {
	assert := assert.New(suite.T())
	userID := "lifecycle-user"
	
	// Create new session
	session := NewSession(userID)
	assert.Equal(SessionStatusActive, session.Status)
	assert.False(session.IsExpired())
	
	// Add some metadata
	session.AddMetadata("login_method", "oauth")
	session.AddMetadata("ip_address", "192.168.1.1")
	
	// Verify metadata
	loginMethod, exists := session.Metadata["login_method"]
	assert.True(exists)
	assert.Equal("oauth", loginMethod)
	
	// Deactivate session
	session.Deactivate()
	assert.Equal(SessionStatusInactive, session.Status)
	
	// Reactivate session
	session.Activate()
	assert.Equal(SessionStatusActive, session.Status)
	
	// Extend session
	originalExpiration := *session.ExpiresAt
	session.ExtendExpiration(1 * time.Hour)
	assert.True(session.ExpiresAt.After(originalExpiration))
	
	// Expire session
	session.Expire()
	assert.Equal(SessionStatusExpired, session.Status)
	
	// Session should still be valid (expired but valid structure)
	assert.True(session.IsValid())
}

// TestSessionTimestamps tests timestamp behavior
func (suite *SessionTestSuite) TestSessionTimestamps() {
	before := time.Now()
	session := NewSession("user-123")
	after := time.Now()

	assert := assert.New(suite.T())
	
	// CreatedAt should be between before and after
	assert.True(session.CreatedAt.After(before) || session.CreatedAt.Equal(before))
	assert.True(session.CreatedAt.Before(after) || session.CreatedAt.Equal(after))
	
	// UpdatedAt should initially equal CreatedAt
	assert.Equal(session.CreatedAt, session.UpdatedAt)
	
	// Adding metadata should update UpdatedAt
	originalUpdatedAt := session.UpdatedAt
	time.Sleep(1 * time.Millisecond)
	session.AddMetadata("test", "value")
	assert.True(session.UpdatedAt.After(originalUpdatedAt))
	
	// Status changes should update UpdatedAt
	originalUpdatedAt = session.UpdatedAt
	time.Sleep(1 * time.Millisecond)
	session.Deactivate()
	assert.True(session.UpdatedAt.After(originalUpdatedAt))
}

// BenchmarkNewSession benchmarks session creation
func BenchmarkNewSession(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewSession("user-123")
	}
}

// BenchmarkSessionAddMetadata benchmarks metadata addition
func BenchmarkSessionAddMetadata(b *testing.B) {
	session := NewSession("user-123")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session.AddMetadata("key", "value")
	}
}

// BenchmarkSessionStatusChanges benchmarks status transitions
func BenchmarkSessionStatusChanges(b *testing.B) {
	session := NewSession("user-123")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session.Deactivate()
		session.Activate()
	}
}

// BenchmarkSessionIsExpired benchmarks expiration checking
func BenchmarkSessionIsExpired(b *testing.B) {
	session := NewSession("user-123")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session.IsExpired()
	}
}