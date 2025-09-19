package values

import (
	"errors"
	"strings"
)

// SessionID represents a unique identifier for sessions
type SessionID struct {
	value string
}

// NewSessionID creates a new SessionID value object
func NewSessionID(value string) (*SessionID, error) {
	if err := validateSessionID(value); err != nil {
		return nil, err
	}
	
	return &SessionID{value: value}, nil
}

// Value returns the string value of the SessionID
func (s SessionID) Value() string {
	return s.value
}

// String implements the Stringer interface
func (s SessionID) String() string {
	return s.value
}

// Equals checks if two SessionIDs are equal
func (s SessionID) Equals(other SessionID) bool {
	return s.value == other.value
}

// validateSessionID validates the session ID format
func validateSessionID(value string) error {
	if value == "" {
		return errors.New("session ID cannot be empty")
	}
	
	if len(value) < 10 {
		return errors.New("session ID must be at least 10 characters")
	}
	
	if strings.Contains(value, " ") {
		return errors.New("session ID cannot contain spaces")
	}
	
	return nil
}