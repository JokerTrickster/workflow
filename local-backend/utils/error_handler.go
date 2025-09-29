package utils

import (
	"context"
	"fmt"
	"log"
	"main/features/github/service"
	"main/utils/db/mysql"
	"os"
	"strings"
	"time"
)

// ErrorHandler manages comprehensive error handling and cleanup operations
type ErrorHandler struct {
	githubService *service.GitHubService
	githubToken   string
	branchManager *BranchManager
	db            interface{}
}

// NewErrorHandler creates a new error handler instance
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		githubService: service.NewGitHubService(),
		githubToken:   os.Getenv("GITHUB_TOKEN"),
		branchManager: GetGlobalBranchManager(),
		db:            mysql.GormMysqlDB,
	}
}

// ErrorType represents different types of errors that can occur
type ErrorType string

const (
	ErrorTypeGitHub      ErrorType = "github"
	ErrorTypeGit         ErrorType = "git"
	ErrorTypeProvider    ErrorType = "provider"
	ErrorTypeDatabase    ErrorType = "database"
	ErrorTypeNetwork     ErrorType = "network"
	ErrorTypeFileSystem  ErrorType = "filesystem"
	ErrorTypePermission  ErrorType = "permission"
	ErrorTypeTimeout     ErrorType = "timeout"
	ErrorTypeUnknown     ErrorType = "unknown"
)

// TaskError represents a comprehensive task execution error
type TaskError struct {
	Type        ErrorType
	Message     string
	OriginalErr error
	TaskMsg     *TaskMessage
	Context     map[string]interface{}
	Timestamp   time.Time
	Recoverable bool
}

// Error implements the error interface
func (e *TaskError) Error() string {
	return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.OriginalErr)
}

// HandleTaskError processes task errors with appropriate recovery strategies
func (eh *ErrorHandler) HandleTaskError(ctx context.Context, taskErr *TaskError) error {
	log.Printf("Handling task error: %s", taskErr.Error())

	// Log error details
	eh.logErrorDetails(taskErr)

	// Attempt recovery based on error type
	if taskErr.Recoverable {
		if err := eh.attemptRecovery(ctx, taskErr); err != nil {
			log.Printf("Recovery failed for task error: %v", err)
		}
	}

	// Perform cleanup operations
	if err := eh.performCleanup(ctx, taskErr); err != nil {
		log.Printf("Cleanup failed for task error: %v", err)
	}

	// Update issue status if applicable
	if err := eh.updateIssueStatus(ctx, taskErr); err != nil {
		log.Printf("Failed to update issue status: %v", err)
	}

	return nil
}

// ClassifyError determines the error type and recovery strategy
func (eh *ErrorHandler) ClassifyError(err error, taskMsg *TaskMessage) *TaskError {
	if err == nil {
		return nil
	}

	taskErr := &TaskError{
		OriginalErr: err,
		TaskMsg:     taskMsg,
		Timestamp:   time.Now(),
		Context:     make(map[string]interface{}),
	}

	errStr := strings.ToLower(err.Error())

	// Classify based on error message patterns
	switch {
	case strings.Contains(errStr, "github api") || strings.Contains(errStr, "401") || strings.Contains(errStr, "403"):
		taskErr.Type = ErrorTypeGitHub
		taskErr.Message = "GitHub API error"
		taskErr.Recoverable = strings.Contains(errStr, "rate limit") || strings.Contains(errStr, "timeout")

	case strings.Contains(errStr, "git") && (strings.Contains(errStr, "push") || strings.Contains(errStr, "pull") || strings.Contains(errStr, "clone")):
		taskErr.Type = ErrorTypeGit
		taskErr.Message = "Git operation error"
		taskErr.Recoverable = true

	case strings.Contains(errStr, "provider") || strings.Contains(errStr, "api key") || strings.Contains(errStr, "not configured"):
		taskErr.Type = ErrorTypeProvider
		taskErr.Message = "AI provider error"
		taskErr.Recoverable = false

	case strings.Contains(errStr, "database") || strings.Contains(errStr, "sql") || strings.Contains(errStr, "gorm"):
		taskErr.Type = ErrorTypeDatabase
		taskErr.Message = "Database operation error"
		taskErr.Recoverable = true

	case strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded"):
		taskErr.Type = ErrorTypeTimeout
		taskErr.Message = "Operation timeout"
		taskErr.Recoverable = true

	case strings.Contains(errStr, "network") || strings.Contains(errStr, "connection") || strings.Contains(errStr, "dial"):
		taskErr.Type = ErrorTypeNetwork
		taskErr.Message = "Network connectivity error"
		taskErr.Recoverable = true

	case strings.Contains(errStr, "permission") || strings.Contains(errStr, "access denied") || strings.Contains(errStr, "forbidden"):
		taskErr.Type = ErrorTypePermission
		taskErr.Message = "Permission error"
		taskErr.Recoverable = false

	case strings.Contains(errStr, "no such file") || strings.Contains(errStr, "directory") || strings.Contains(errStr, "path"):
		taskErr.Type = ErrorTypeFileSystem
		taskErr.Message = "File system error"
		taskErr.Recoverable = false

	default:
		taskErr.Type = ErrorTypeUnknown
		taskErr.Message = "Unknown error"
		taskErr.Recoverable = false
	}

	return taskErr
}

// attemptRecovery tries to recover from recoverable errors
func (eh *ErrorHandler) attemptRecovery(ctx context.Context, taskErr *TaskError) error {
	log.Printf("Attempting recovery for error type: %s", taskErr.Type)

	switch taskErr.Type {
	case ErrorTypeGitHub:
		return eh.recoverFromGitHubError(ctx, taskErr)
	case ErrorTypeGit:
		return eh.recoverFromGitError(ctx, taskErr)
	case ErrorTypeNetwork:
		return eh.recoverFromNetworkError(ctx, taskErr)
	case ErrorTypeTimeout:
		return eh.recoverFromTimeoutError(ctx, taskErr)
	case ErrorTypeDatabase:
		return eh.recoverFromDatabaseError(ctx, taskErr)
	default:
		return fmt.Errorf("no recovery strategy for error type: %s", taskErr.Type)
	}
}

// recoverFromGitHubError handles GitHub-specific recovery
func (eh *ErrorHandler) recoverFromGitHubError(ctx context.Context, taskErr *TaskError) error {
	errStr := strings.ToLower(taskErr.OriginalErr.Error())

	if strings.Contains(errStr, "rate limit") {
		log.Printf("GitHub rate limit detected, implementing backoff strategy")
		// Wait for rate limit reset (GitHub rate limits reset every hour)
		time.Sleep(5 * time.Minute)
		return nil
	}

	if strings.Contains(errStr, "timeout") {
		log.Printf("GitHub API timeout detected, retrying with exponential backoff")
		time.Sleep(30 * time.Second)
		return nil
	}

	return fmt.Errorf("no recovery strategy for GitHub error: %s", taskErr.OriginalErr)
}

// recoverFromGitError handles Git operation recovery
func (eh *ErrorHandler) recoverFromGitError(ctx context.Context, taskErr *TaskError) error {
	if taskErr.TaskMsg == nil || taskErr.TaskMsg.WorkingDir == "" {
		return fmt.Errorf("insufficient context for Git recovery")
	}

	log.Printf("Attempting Git recovery in directory: %s", taskErr.TaskMsg.WorkingDir)

	// Try to reset to a clean state
	if err := eh.resetGitState(ctx, taskErr.TaskMsg.WorkingDir); err != nil {
		return fmt.Errorf("failed to reset Git state: %w", err)
	}

	return nil
}

// recoverFromNetworkError handles network connectivity issues
func (eh *ErrorHandler) recoverFromNetworkError(ctx context.Context, taskErr *TaskError) error {
	log.Printf("Implementing network recovery strategy")

	// Simple exponential backoff
	backoffDuration := 30 * time.Second
	log.Printf("Waiting %v for network recovery", backoffDuration)
	time.Sleep(backoffDuration)

	return nil
}

// recoverFromTimeoutError handles timeout scenarios
func (eh *ErrorHandler) recoverFromTimeoutError(ctx context.Context, taskErr *TaskError) error {
	log.Printf("Implementing timeout recovery strategy")

	// For timeouts, we typically just need to wait and retry
	time.Sleep(10 * time.Second)

	return nil
}

// recoverFromDatabaseError handles database connectivity issues
func (eh *ErrorHandler) recoverFromDatabaseError(ctx context.Context, taskErr *TaskError) error {
	log.Printf("Attempting database recovery")

	// Try to reconnect to database
	if mysql.GormMysqlDB != nil {
		db, err := mysql.GormMysqlDB.DB()
		if err == nil {
			if err := db.Ping(); err == nil {
				log.Printf("Database connection restored")
				return nil
			}
		}
	}

	// Wait before potential retry
	time.Sleep(5 * time.Second)
	return nil
}

// performCleanup executes cleanup operations based on error context
func (eh *ErrorHandler) performCleanup(ctx context.Context, taskErr *TaskError) error {
	log.Printf("Performing cleanup operations for error type: %s", taskErr.Type)

	// Clean up branch if task failed
	if taskErr.TaskMsg != nil && taskErr.TaskMsg.RepositoryName != "" {
		if err := eh.branchManager.CompleteBranch(taskErr.TaskMsg.RepositoryName, false); err != nil {
			log.Printf("Failed to clean up branch: %v", err)
		}
	}

	// Clean up temporary files and resources
	if err := eh.cleanupTemporaryResources(ctx, taskErr); err != nil {
		log.Printf("Failed to clean up temporary resources: %v", err)
	}

	// Perform orphaned branch cleanup if needed
	if taskErr.Type == ErrorTypeGit && taskErr.TaskMsg != nil {
		if err := eh.performOrphanedBranchCleanup(ctx, taskErr.TaskMsg.WorkingDir); err != nil {
			log.Printf("Failed to clean up orphaned branches: %v", err)
		}
	}

	return nil
}

// updateIssueStatus updates GitHub issue status for failed tasks
func (eh *ErrorHandler) updateIssueStatus(ctx context.Context, taskErr *TaskError) error {
	if eh.githubService == nil || eh.githubToken == "" || taskErr.TaskMsg == nil {
		return nil
	}

	// Try to find and update associated GitHub issue
	// This would require tracking issue numbers in the task context
	log.Printf("Issue status update not implemented yet")

	return nil
}

// resetGitState resets Git repository to a clean state
func (eh *ErrorHandler) resetGitState(ctx context.Context, workingDir string) error {
	log.Printf("Resetting Git state in directory: %s", workingDir)

	// This is a simplified reset - in production, you might want more sophisticated logic
	// For now, we'll just ensure we're on the default branch
	if err := eh.branchManager.createGitBranch(ctx, workingDir, "main"); err != nil {
		// Try master if main doesn't work
		if err := eh.branchManager.createGitBranch(ctx, workingDir, "master"); err != nil {
			return fmt.Errorf("failed to reset to default branch: %w", err)
		}
	}

	return nil
}

// cleanupTemporaryResources removes temporary files and resources
func (eh *ErrorHandler) cleanupTemporaryResources(ctx context.Context, taskErr *TaskError) error {
	log.Printf("Cleaning up temporary resources")

	// Clean up any temporary directories or files created during task execution
	// This would be implemented based on specific temporary resource patterns

	return nil
}

// performOrphanedBranchCleanup removes old task branches
func (eh *ErrorHandler) performOrphanedBranchCleanup(ctx context.Context, workingDir string) error {
	if workingDir == "" {
		return nil
	}

	log.Printf("Performing orphaned branch cleanup in: %s", workingDir)

	// Clean up branches older than 24 hours
	return eh.branchManager.CleanupOrphanedBranches(ctx, workingDir, 24*time.Hour)
}

// logErrorDetails logs comprehensive error information
func (eh *ErrorHandler) logErrorDetails(taskErr *TaskError) {
	log.Printf("=== Task Error Details ===")
	log.Printf("Type: %s", taskErr.Type)
	log.Printf("Message: %s", taskErr.Message)
	log.Printf("Recoverable: %t", taskErr.Recoverable)
	log.Printf("Timestamp: %s", taskErr.Timestamp.Format(time.RFC3339))

	if taskErr.TaskMsg != nil {
		log.Printf("Task: %s", taskErr.TaskMsg.Tasks)
		log.Printf("Repository: %s", taskErr.TaskMsg.RepositoryName)
		log.Printf("Provider: %s", taskErr.TaskMsg.Provider)
	}

	if taskErr.OriginalErr != nil {
		log.Printf("Original Error: %v", taskErr.OriginalErr)
	}

	log.Printf("========================")
}

// GetGlobalErrorHandler returns a global error handler instance
var globalErrorHandler = NewErrorHandler()

func GetGlobalErrorHandler() *ErrorHandler {
	return globalErrorHandler
}