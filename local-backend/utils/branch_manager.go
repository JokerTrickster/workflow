package utils

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// BranchManager handles branch lifecycle and conflict prevention
type BranchManager struct {
	mutex           sync.RWMutex
	activeBranches  map[string]*BranchInfo // repository -> branch info
	repositoryLocks map[string]*sync.Mutex // repository -> lock
}

// BranchInfo tracks active branch information
type BranchInfo struct {
	Name          string
	TaskID        string
	CreatedAt     time.Time
	Repository    string
	WorkingDir    string
	IsActive      bool
}

var globalBranchManager = NewBranchManager()

// NewBranchManager creates a new branch manager
func NewBranchManager() *BranchManager {
	return &BranchManager{
		activeBranches:  make(map[string]*BranchInfo),
		repositoryLocks: make(map[string]*sync.Mutex),
	}
}

// GetGlobalBranchManager returns the global branch manager instance
func GetGlobalBranchManager() *BranchManager {
	return globalBranchManager
}

// CreateTaskBranch creates a unique branch for a task with conflict prevention
func (bm *BranchManager) CreateTaskBranch(ctx context.Context, taskMsg *TaskMessage) (*BranchInfo, error) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	// Get or create repository lock
	if _, exists := bm.repositoryLocks[taskMsg.RepositoryName]; !exists {
		bm.repositoryLocks[taskMsg.RepositoryName] = &sync.Mutex{}
	}
	repoLock := bm.repositoryLocks[taskMsg.RepositoryName]

	// Lock repository for branch creation
	repoLock.Lock()
	defer repoLock.Unlock()

	// Check if there's already an active branch for this repository
	if existingBranch, exists := bm.activeBranches[taskMsg.RepositoryName]; exists && existingBranch.IsActive {
		log.Printf("Repository %s already has active branch %s, waiting for completion...",
			taskMsg.RepositoryName, existingBranch.Name)

		// For now, we'll create a new branch anyway but with conflict prevention
		// In the future, we could implement queuing or branch merging
	}

	// Use requested branch name or generate unique one
	var branchName string
	if taskMsg.BranchName != "" {
		branchName = taskMsg.BranchName
		log.Printf("Using requested branch name: %s", branchName)
	} else {
		branchName = bm.generateUniqueBranchName(taskMsg)
		log.Printf("Generated branch name: %s", branchName)
	}

	// Create branch info
	branchInfo := &BranchInfo{
		Name:       branchName,
		TaskID:     bm.generateTaskID(taskMsg),
		CreatedAt:  time.Now(),
		Repository: taskMsg.RepositoryName,
		WorkingDir: taskMsg.WorkingDir,
		IsActive:   true,
	}

	// Create the actual Git branch using absolute path
	repoPath := GetRepositoryPath(taskMsg.RepositoryName)
	if err := bm.createGitBranch(ctx, repoPath, branchName); err != nil {
		return nil, fmt.Errorf("failed to create Git branch: %w", err)
	}

	// Register the branch
	bm.activeBranches[taskMsg.RepositoryName] = branchInfo

	log.Printf("Created task branch %s for repository %s", branchName, taskMsg.RepositoryName)
	return branchInfo, nil
}

// generateUniqueBranchName creates a unique branch name with conflict prevention
func (bm *BranchManager) generateUniqueBranchName(taskMsg *TaskMessage) string {
	// Create a hash of the task content for uniqueness
	hasher := sha256.New()
	hasher.Write([]byte(taskMsg.Tasks))
	hasher.Write([]byte(taskMsg.RepositoryName))
	hasher.Write([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))

	hash := fmt.Sprintf("%x", hasher.Sum(nil))[:8]

	// Sanitize task description for branch name
	sanitizedTask := bm.sanitizeForBranchName(taskMsg.Tasks)
	if len(sanitizedTask) > 30 {
		sanitizedTask = sanitizedTask[:30]
	}

	timestamp := time.Now().Format("0102-1504") // MMDD-HHMM

	return fmt.Sprintf("task/%s-%s-%s", sanitizedTask, timestamp, hash)
}

// sanitizeForBranchName removes invalid characters from branch names
func (bm *BranchManager) sanitizeForBranchName(input string) string {
	// Replace spaces and special characters with hyphens
	sanitized := strings.ReplaceAll(input, " ", "-")
	sanitized = strings.ReplaceAll(sanitized, "_", "-")
	sanitized = strings.ReplaceAll(sanitized, ".", "-")
	sanitized = strings.ReplaceAll(sanitized, "/", "-")
	sanitized = strings.ReplaceAll(sanitized, "\\", "-")
	sanitized = strings.ReplaceAll(sanitized, ":", "-")
	sanitized = strings.ReplaceAll(sanitized, "*", "-")
	sanitized = strings.ReplaceAll(sanitized, "?", "-")
	sanitized = strings.ReplaceAll(sanitized, "\"", "-")
	sanitized = strings.ReplaceAll(sanitized, "<", "-")
	sanitized = strings.ReplaceAll(sanitized, ">", "-")
	sanitized = strings.ReplaceAll(sanitized, "|", "-")

	// Remove consecutive hyphens
	for strings.Contains(sanitized, "--") {
		sanitized = strings.ReplaceAll(sanitized, "--", "-")
	}

	// Trim hyphens from start and end
	sanitized = strings.Trim(sanitized, "-")

	// Ensure it's not empty
	if sanitized == "" {
		sanitized = "task"
	}

	return strings.ToLower(sanitized)
}

// generateTaskID creates a unique task identifier
func (bm *BranchManager) generateTaskID(taskMsg *TaskMessage) string {
	hasher := sha256.New()
	hasher.Write([]byte(taskMsg.Tasks))
	hasher.Write([]byte(taskMsg.RepositoryName))
	hasher.Write([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return fmt.Sprintf("%x", hasher.Sum(nil))[:16]
}

// createGitBranch creates a new Git branch
func (bm *BranchManager) createGitBranch(ctx context.Context, workingDir, branchName string) error {
	// Ensure we're in a Git repository
	gitDir := filepath.Join(workingDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("not a Git repository: %s", workingDir)
	}

	// Fetch latest changes to ensure we're up to date
	if err := bm.runGitCommand(ctx, workingDir, "fetch", "origin"); err != nil {
		log.Printf("Warning: failed to fetch latest changes: %v", err)
		// Continue anyway, this is not critical
	}

	// Get the default branch name
	defaultBranch, err := bm.getDefaultBranch(ctx, workingDir)
	if err != nil {
		log.Printf("Warning: failed to get default branch, using 'main': %v", err)
		defaultBranch = "main"
	}

	// Checkout default branch first
	if err := bm.runGitCommand(ctx, workingDir, "checkout", defaultBranch); err != nil {
		// Try master if main doesn't exist
		if err := bm.runGitCommand(ctx, workingDir, "checkout", "master"); err != nil {
			return fmt.Errorf("failed to checkout default branch: %w", err)
		}
		defaultBranch = "master"
	}

	// Pull latest changes
	if err := bm.runGitCommand(ctx, workingDir, "pull", "origin", defaultBranch); err != nil {
		log.Printf("Warning: failed to pull latest changes: %v", err)
		// Continue anyway, this is not critical
	}

	// Create and checkout new branch
	if err := bm.runGitCommand(ctx, workingDir, "checkout", "-b", branchName); err != nil {
		return fmt.Errorf("failed to create branch %s: %w", branchName, err)
	}

	return nil
}

// getDefaultBranch determines the default branch of the repository
func (bm *BranchManager) getDefaultBranch(ctx context.Context, workingDir string) (string, error) {
	output, err := bm.runGitCommandWithOutput(ctx, workingDir, "remote", "show", "origin")
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "HEAD branch:") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return "", fmt.Errorf("could not determine default branch")
}

// CompleteBranch marks a branch as completed and performs cleanup
func (bm *BranchManager) CompleteBranch(repositoryName string, success bool) error {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	branchInfo, exists := bm.activeBranches[repositoryName]
	if !exists {
		log.Printf("No active branch found for repository %s", repositoryName)
		return nil
	}

	branchInfo.IsActive = false

	if success {
		log.Printf("Branch %s completed successfully for repository %s",
			branchInfo.Name, repositoryName)
	} else {
		log.Printf("Branch %s completed with errors for repository %s",
			branchInfo.Name, repositoryName)

		// For failed tasks, we might want to keep the branch for debugging
		// or perform additional cleanup
	}

	// Remove from active branches
	delete(bm.activeBranches, repositoryName)

	return nil
}

// CleanupOrphanedBranches removes old task branches that are no longer needed
func (bm *BranchManager) CleanupOrphanedBranches(ctx context.Context, workingDir string, olderThan time.Duration) error {
	// Get all local branches
	output, err := bm.runGitCommandWithOutput(ctx, workingDir, "branch", "--format=%(refname:short)")
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}

	branches := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, branch := range branches {
		branch = strings.TrimSpace(branch)
		if branch == "" || !strings.HasPrefix(branch, "task/") {
			continue
		}

		// Check if branch is older than threshold
		if bm.isBranchOlderThan(ctx, workingDir, branch, olderThan) {
			log.Printf("Cleaning up old task branch: %s", branch)

			// Switch to default branch before deleting
			defaultBranch, _ := bm.getDefaultBranch(ctx, workingDir)
			if defaultBranch == "" {
				defaultBranch = "main"
			}

			if err := bm.runGitCommand(ctx, workingDir, "checkout", defaultBranch); err != nil {
				log.Printf("Warning: failed to checkout %s before cleanup: %v", defaultBranch, err)
				continue
			}

			// Delete the branch
			if err := bm.runGitCommand(ctx, workingDir, "branch", "-D", branch); err != nil {
				log.Printf("Warning: failed to delete branch %s: %v", branch, err)
			}
		}
	}

	return nil
}

// isBranchOlderThan checks if a branch is older than the specified duration
func (bm *BranchManager) isBranchOlderThan(ctx context.Context, workingDir, branch string, duration time.Duration) bool {
	// Get the commit date of the branch
	output, err := bm.runGitCommandWithOutput(ctx, workingDir, "log", "-1", "--format=%ct", branch)
	if err != nil {
		return false
	}

	// Parse Unix timestamp
	timestamp := strings.TrimSpace(string(output))
	if timestamp == "" {
		return false
	}

	var unixTime int64
	if _, err := fmt.Sscanf(timestamp, "%d", &unixTime); err != nil {
		return false
	}

	branchTime := time.Unix(unixTime, 0)
	return time.Since(branchTime) > duration
}

// GetActiveBranches returns all currently active branches
func (bm *BranchManager) GetActiveBranches() map[string]*BranchInfo {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	result := make(map[string]*BranchInfo)
	for repo, branch := range bm.activeBranches {
		if branch.IsActive {
			result[repo] = branch
		}
	}

	return result
}

// runGitCommand executes a Git command in the specified directory
func (bm *BranchManager) runGitCommand(ctx context.Context, workingDir string, args ...string) error {
	_, err := bm.runGitCommandWithOutput(ctx, workingDir, args...)
	return err
}

// runGitCommandWithOutput executes a Git command and returns output
func (bm *BranchManager) runGitCommandWithOutput(ctx context.Context, workingDir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = workingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Git command failed: git %s\nOutput: %s\nError: %v",
			strings.Join(args, " "), string(output), err)
		return string(output), err
	}

	return string(output), nil
}