package values

import (
	"errors"
	"strings"
)

// MessageID represents a unique identifier for messages
type MessageID struct {
	value string
}

// NewMessageID creates a new MessageID value object
func NewMessageID(value string) (*MessageID, error) {
	if err := validateMessageID(value); err != nil {
		return nil, err
	}
	
	return &MessageID{value: value}, nil
}

// Value returns the string value of the MessageID
func (m MessageID) Value() string {
	return m.value
}

// String implements the Stringer interface
func (m MessageID) String() string {
	return m.value
}

// Equals checks if two MessageIDs are equal
func (m MessageID) Equals(other MessageID) bool {
	return m.value == other.value
}

// validateMessageID validates the message ID format
func validateMessageID(value string) error {
	if value == "" {
		return errors.New("message ID cannot be empty")
	}
	
	if len(value) < 10 {
		return errors.New("message ID must be at least 10 characters")
	}
	
	if strings.Contains(value, " ") {
		return errors.New("message ID cannot contain spaces")
	}
	
	return nil
}