package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// MessageID represents a unique message identifier
type MessageID struct {
	value string
}

// NewMessageID creates a new MessageID with validation
func NewMessageID(value string) (*MessageID, error) {
	if strings.TrimSpace(value) == "" {
		return nil, fmt.Errorf("message ID cannot be empty")
	}
	if len(value) > 255 {
		return nil, fmt.Errorf("message ID cannot exceed 255 characters")
	}
	return &MessageID{value: value}, nil
}

// String returns the string representation
func (m *MessageID) String() string {
	return m.value
}

// Value returns the underlying value
func (m *MessageID) Value() string {
	return m.value
}

// Equals checks equality with another MessageID
func (m *MessageID) Equals(other *MessageID) bool {
	if other == nil {
		return false
	}
	return m.value == other.value
}

// RequestID represents a unique request identifier
type RequestID struct {
	value string
}

// NewRequestID creates a new RequestID with validation
func NewRequestID(value string) (*RequestID, error) {
	if strings.TrimSpace(value) == "" {
		return nil, fmt.Errorf("request ID cannot be empty")
	}
	if len(value) > 255 {
		return nil, fmt.Errorf("request ID cannot exceed 255 characters")
	}
	return &RequestID{value: value}, nil
}

// String returns the string representation
func (r *RequestID) String() string {
	return r.value
}

// Value returns the underlying value
func (r *RequestID) Value() string {
	return r.value
}

// Equals checks equality with another RequestID
func (r *RequestID) Equals(other *RequestID) bool {
	if other == nil {
		return false
	}
	return r.value == other.value
}

// ContextID represents a conversation context identifier
type ContextID struct {
	value string
}

// NewContextID creates a new ContextID with validation
func NewContextID(value string) (*ContextID, error) {
	if strings.TrimSpace(value) == "" {
		return nil, fmt.Errorf("context ID cannot be empty")
	}
	if len(value) > 255 {
		return nil, fmt.Errorf("context ID cannot exceed 255 characters")
	}
	return &ContextID{value: value}, nil
}

// String returns the string representation
func (c *ContextID) String() string {
	return c.value
}

// Value returns the underlying value
func (c *ContextID) Value() string {
	return c.value
}

// Equals checks equality with another ContextID
func (c *ContextID) Equals(other *ContextID) bool {
	if other == nil {
		return false
	}
	return c.value == other.value
}

// CodeAnalysisRequest represents a structured code analysis request
type CodeAnalysisRequest struct {
	Code        string            `json:"code"`
	Task        string            `json:"task"`
	Language    string            `json:"language,omitempty"`
	Framework   string            `json:"framework,omitempty"`
	Context     string            `json:"context,omitempty"`
	Preferences map[string]string `json:"preferences,omitempty"`
}

// NewCodeAnalysisRequest creates a new CodeAnalysisRequest with validation
func NewCodeAnalysisRequest(code, task string) (*CodeAnalysisRequest, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("code cannot be empty")
	}
	if strings.TrimSpace(task) == "" {
		return nil, fmt.Errorf("task cannot be empty")
	}
	if len(code) > 100000 { // 100KB limit
		return nil, fmt.Errorf("code exceeds maximum length of 100KB")
	}
	
	return &CodeAnalysisRequest{
		Code:        code,
		Task:        task,
		Preferences: make(map[string]string),
	}, nil
}

// SetLanguage sets the programming language
func (car *CodeAnalysisRequest) SetLanguage(language string) {
	car.Language = language
}

// SetFramework sets the framework information
func (car *CodeAnalysisRequest) SetFramework(framework string) {
	car.Framework = framework
}

// SetContext sets additional context information
func (car *CodeAnalysisRequest) SetContext(context string) {
	car.Context = context
}

// SetPreference sets a preference key-value pair
func (car *CodeAnalysisRequest) SetPreference(key, value string) {
	if car.Preferences == nil {
		car.Preferences = make(map[string]string)
	}
	car.Preferences[key] = value
}

// ToJSON serializes the request to JSON
func (car *CodeAnalysisRequest) ToJSON() (string, error) {
	data, err := json.Marshal(car)
	if err != nil {
		return "", fmt.Errorf("failed to serialize request: %w", err)
	}
	return string(data), nil
}

// FromJSON deserializes a request from JSON
func FromJSON(data string) (*CodeAnalysisRequest, error) {
	var request CodeAnalysisRequest
	if err := json.Unmarshal([]byte(data), &request); err != nil {
		return nil, fmt.Errorf("failed to deserialize request: %w", err)
	}
	
	// Validate deserialized data
	if strings.TrimSpace(request.Code) == "" {
		return nil, fmt.Errorf("code cannot be empty")
	}
	if strings.TrimSpace(request.Task) == "" {
		return nil, fmt.Errorf("task cannot be empty")
	}
	
	return &request, nil
}

// Duration represents a time duration with validation
type Duration struct {
	value time.Duration
}

// NewDuration creates a new Duration with validation
func NewDuration(value time.Duration) (*Duration, error) {
	if value < 0 {
		return nil, fmt.Errorf("duration cannot be negative")
	}
	if value > 24*time.Hour {
		return nil, fmt.Errorf("duration cannot exceed 24 hours")
	}
	return &Duration{value: value}, nil
}

// Value returns the underlying duration
func (d *Duration) Value() time.Duration {
	return d.value
}

// String returns string representation
func (d *Duration) String() string {
	return d.value.String()
}

// Seconds returns duration in seconds
func (d *Duration) Seconds() float64 {
	return d.value.Seconds()
}

// Minutes returns duration in minutes
func (d *Duration) Minutes() float64 {
	return d.value.Minutes()
}

// Hours returns duration in hours
func (d *Duration) Hours() float64 {
	return d.value.Hours()
}