package valueobjects

import (
	"fmt"
	"regexp"
	"strings"
)

// BranchName represents a valid Git branch name
type BranchName struct {
	value string
}

// Git branch name validation pattern
var branchNamePattern = regexp.MustCompile(`^[a-zA-Z0-9-_./]+$`)

// NewBranchName creates a new BranchName
func NewBranchName(value string) (BranchName, error) {
	if value == "" {
		return BranchName{}, fmt.Errorf("branch name cannot be empty")
	}

	value = strings.TrimSpace(value)
	if len(value) < 1 || len(value) > 255 {
		return BranchName{}, fmt.Errorf("branch name must be between 1 and 255 characters")
	}

	// Git branch name validation rules
	if strings.HasPrefix(value, "/") || strings.HasSuffix(value, "/") {
		return BranchName{}, fmt.Errorf("branch name cannot start or end with '/'")
	}

	if strings.Contains(value, "//") {
		return BranchName{}, fmt.Errorf("branch name cannot contain consecutive '//' characters")
	}

	if strings.HasPrefix(value, ".") || strings.HasSuffix(value, ".") {
		return BranchName{}, fmt.Errorf("branch name cannot start or end with '.'")
	}

	if !branchNamePattern.MatchString(value) {
		return BranchName{}, fmt.Errorf("branch name contains invalid characters. Only alphanumeric, hyphens, underscores, slashes, and dots are allowed")
	}

	return BranchName{value: value}, nil
}

// NewMainBranch creates a branch name for the main branch
func NewMainBranch() BranchName {
	return BranchName{value: "main"}
}

// Value returns the string value of the BranchName
func (b BranchName) Value() string {
	return b.value
}

// String implements the Stringer interface
func (b BranchName) String() string {
	return b.value
}

// Equals checks if two BranchNames are equal
func (b BranchName) Equals(other BranchName) bool {
	return b.value == other.value
}

// IsEmpty checks if the BranchName is empty
func (b BranchName) IsEmpty() bool {
	return b.value == ""
}

// IsMain checks if this is a main branch (main, master)
func (b BranchName) IsMain() bool {
	return b.value == "main" || b.value == "master"
}

// IsFeatureBranch checks if this is a feature branch
func (b BranchName) IsFeatureBranch() bool {
	return strings.HasPrefix(b.value, "feature/") || strings.HasPrefix(b.value, "feat/")
}

// IsBugfixBranch checks if this is a bugfix branch
func (b BranchName) IsBugfixBranch() bool {
	return strings.HasPrefix(b.value, "bugfix/") || strings.HasPrefix(b.value, "fix/")
}

// IsHotfixBranch checks if this is a hotfix branch
func (b BranchName) IsHotfixBranch() bool {
	return strings.HasPrefix(b.value, "hotfix/")
}