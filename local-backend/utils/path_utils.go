package utils

import (
	"os"
	"path/filepath"
)

// GetRepositoryBasePath returns the base path for repositories based on the current user
func GetRepositoryBasePath() string {
	// First try environment variable
	if basePath := os.Getenv("REPOSITORY_BASE_PATH"); basePath != "" {
		return basePath
	}

	// Use relative path from current project directory
	// Current path: /Users/luxrobo/project/workflow/local-backend
	// Target path: /Users/luxrobo/project/git-repository
	return "../../git-repository"
}

// GetRepositoryPath returns the full path for a specific repository
func GetRepositoryPath(repoName string) string {
	basePath := GetRepositoryBasePath()
	return filepath.Join(basePath, "JokerTrickster", repoName)
}

// GetJokerTricksterPath returns the path to the JokerTrickster directory
func GetJokerTricksterPath() string {
	basePath := GetRepositoryBasePath()
	return filepath.Join(basePath, "JokerTrickster")
}

// EnsureRepositoryPath creates the repository directory if it doesn't exist
func EnsureRepositoryPath(repoName string) error {
	repoPath := GetRepositoryPath(repoName)
	return os.MkdirAll(repoPath, 0755)
}

// EnsureJokerTricksterPath creates the JokerTrickster directory if it doesn't exist
func EnsureJokerTricksterPath() error {
	jokerPath := GetJokerTricksterPath()
	return os.MkdirAll(jokerPath, 0755)
}