package utils

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetRepositoryBasePath returns the base path for repositories based on the current user
func GetRepositoryBasePath() string {
	// First try environment variable
	if basePath := os.Getenv("REPOSITORY_BASE_PATH"); basePath != "" {
		return basePath
	}

	// Fallback to dynamic path based on current user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If can't get home directory, fallback to current working directory
		if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
			// For macOS/Linux, try common paths
			if _, err := os.Stat("/Users"); err == nil {
				// Try to detect current user from common pattern
				if currentUser := os.Getenv("USER"); currentUser != "" {
					return filepath.Join("/Users", currentUser, "project", "git-repository")
				}
			}
		}
		return "/tmp/repositories" // Last resort fallback
	}

	return filepath.Join(homeDir, "project", "git-repository")
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