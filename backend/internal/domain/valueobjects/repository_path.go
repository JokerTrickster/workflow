package valueobjects

import (
	"fmt"
	"regexp"
	"strings"
)

// RepositoryPath represents a valid repository path
type RepositoryPath struct {
	value string
}

// GitHub repository path pattern: owner/repo
var repoPathPattern = regexp.MustCompile(`^[a-zA-Z0-9-_.]+/[a-zA-Z0-9-_.]+$`)

// NewRepositoryPath creates a new RepositoryPath
func NewRepositoryPath(value string) (RepositoryPath, error) {
	if value == "" {
		return RepositoryPath{}, fmt.Errorf("repository path cannot be empty")
	}

	value = strings.TrimSpace(value)
	if len(value) < 3 || len(value) > 500 {
		return RepositoryPath{}, fmt.Errorf("repository path must be between 3 and 500 characters")
	}

	if !repoPathPattern.MatchString(value) {
		return RepositoryPath{}, fmt.Errorf("repository path must be in format 'owner/repository' and contain only alphanumeric characters, hyphens, underscores, and dots")
	}

	return RepositoryPath{value: value}, nil
}

// Value returns the string value of the RepositoryPath
func (r RepositoryPath) Value() string {
	return r.value
}

// String implements the Stringer interface
func (r RepositoryPath) String() string {
	return r.value
}

// Equals checks if two RepositoryPaths are equal
func (r RepositoryPath) Equals(other RepositoryPath) bool {
	return r.value == other.value
}

// Owner returns the owner part of the repository path
func (r RepositoryPath) Owner() string {
	parts := strings.Split(r.value, "/")
	if len(parts) >= 1 {
		return parts[0]
	}
	return ""
}

// Repository returns the repository name part of the path
func (r RepositoryPath) Repository() string {
	parts := strings.Split(r.value, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// IsEmpty checks if the RepositoryPath is empty
func (r RepositoryPath) IsEmpty() bool {
	return r.value == ""
}