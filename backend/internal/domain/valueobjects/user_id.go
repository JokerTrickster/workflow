package valueobjects

import (
	"fmt"
	"regexp"
	"strings"
)

// UserID represents a unique identifier for a user
type UserID struct {
	value string
}

var userIDPattern = regexp.MustCompile(`^[a-zA-Z0-9-_.@]+$`)

// NewUserID creates a new UserID
func NewUserID(value string) (UserID, error) {
	if value == "" {
		return UserID{}, fmt.Errorf("user ID cannot be empty")
	}

	value = strings.TrimSpace(value)
	if len(value) < 1 || len(value) > 255 {
		return UserID{}, fmt.Errorf("user ID must be between 1 and 255 characters")
	}

	if !userIDPattern.MatchString(value) {
		return UserID{}, fmt.Errorf("user ID can only contain alphanumeric characters, hyphens, underscores, dots, and @ symbols")
	}

	return UserID{value: value}, nil
}

// Value returns the string value of the UserID
func (u UserID) Value() string {
	return u.value
}

// String implements the Stringer interface
func (u UserID) String() string {
	return u.value
}

// Equals checks if two UserIDs are equal
func (u UserID) Equals(other UserID) bool {
	return u.value == other.value
}

// IsEmpty checks if the UserID is empty
func (u UserID) IsEmpty() bool {
	return u.value == ""
}