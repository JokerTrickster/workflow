package values

import (
	"errors"
	"strings"
)

const (
	MaxMessageContentLength = 10000
	MinMessageContentLength = 1
)

// MessageContent represents validated message content
type MessageContent struct {
	value string
}

// NewMessageContent creates a new MessageContent value object
func NewMessageContent(value string) (*MessageContent, error) {
	if err := validateMessageContent(value); err != nil {
		return nil, err
	}
	
	// Sanitize the content
	sanitized := sanitizeContent(value)
	
	return &MessageContent{value: sanitized}, nil
}

// Value returns the string value of the MessageContent
func (m MessageContent) Value() string {
	return m.value
}

// String implements the Stringer interface
func (m MessageContent) String() string {
	return m.value
}

// Length returns the length of the content
func (m MessageContent) Length() int {
	return len(m.value)
}

// Equals checks if two MessageContents are equal
func (m MessageContent) Equals(other MessageContent) bool {
	return m.value == other.value
}

// IsEmpty checks if the content is empty
func (m MessageContent) IsEmpty() bool {
	return strings.TrimSpace(m.value) == ""
}

// validateMessageContent validates the message content
func validateMessageContent(value string) error {
	if value == "" {
		return errors.New("message content cannot be empty")
	}
	
	if len(value) < MinMessageContentLength {
		return errors.New("message content is too short")
	}
	
	if len(value) > MaxMessageContentLength {
		return errors.New("message content is too long")
	}
	
	return nil
}

// sanitizeContent performs basic content sanitization
func sanitizeContent(content string) string {
	// Remove null bytes
	content = strings.ReplaceAll(content, "\x00", "")
	
	// Trim excessive whitespace
	content = strings.TrimSpace(content)
	
	return content
}