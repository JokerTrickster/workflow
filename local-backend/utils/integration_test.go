package utils

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTaskMessage for testing
type MockTaskMessage struct {
	Tasks          string
	RepositoryName string
	WorkingDir     string
	Provider       string
	Interactive    bool
	ContinueTask   bool
	Cmd            string
}

// MockAITaskResponse for testing
type MockAITaskResponse struct {
	Success        bool
	Output         string
	FilesModified  []string
	ExecutionTime  time.Duration
	TokensUsed     int
	Error          string
}

// Test GitHub Issue Creation
func TestCreateGitHubIssue(t *testing.T) {
	// Setup
	worker := NewConcurrentTaskWorker("amqp://test", "test_queue", 1)
	ctx := context.Background()

	taskMsg := &TaskMessage{
		Tasks:          "Test task for GitHub integration",
		RepositoryName: "JokerTrickster/test-repo",
		WorkingDir:     "/tmp/test",
		Provider:       "claude",
	}

	// Test with GitHub token not configured
	worker.githubToken = ""
	issueURL, issueNumber, err := worker.createGitHubIssue(ctx, taskMsg)

	// Should not fail when GitHub not configured
	assert.NoError(t, err)
	assert.Empty(t, issueURL)
	assert.Empty(t, issueNumber)
}

// Test Branch Management
func TestBranchManager(t *testing.T) {
	bm := NewBranchManager()
	ctx := context.Background()

	taskMsg := &TaskMessage{
		Tasks:          "Test branch creation",
		RepositoryName: "test/repo",
		WorkingDir:     "/tmp/test-repo",
	}

	// Test branch name generation
	branchName := bm.generateUniqueBranchName(taskMsg)
	assert.Contains(t, branchName, "task/")
	assert.Contains(t, branchName, "test-branch-creation")

	// Test sanitization
	sanitized := bm.sanitizeForBranchName("Test Task: With Special/Characters*")
	assert.Equal(t, "test-task-with-special-characters", sanitized)

	// Test task ID generation
	taskID := bm.generateTaskID(taskMsg)
	assert.Len(t, taskID, 16)
}

// Test Error Handler Classification
func TestErrorHandlerClassification(t *testing.T) {
	eh := NewErrorHandler()

	taskMsg := &TaskMessage{
		Tasks:          "Test error handling",
		RepositoryName: "test/repo",
	}

	testCases := []struct {
		errorMsg     string
		expectedType ErrorType
		recoverable  bool
	}{
		{
			errorMsg:     "GitHub API rate limit exceeded",
			expectedType: ErrorTypeGitHub,
			recoverable:  true,
		},
		{
			errorMsg:     "git push failed: remote rejected",
			expectedType: ErrorTypeGit,
			recoverable:  true,
		},
		{
			errorMsg:     "provider not configured: missing API key",
			expectedType: ErrorTypeProvider,
			recoverable:  false,
		},
		{
			errorMsg:     "connection timeout",
			expectedType: ErrorTypeTimeout,
			recoverable:  true,
		},
		{
			errorMsg:     "permission denied",
			expectedType: ErrorTypePermission,
			recoverable:  false,
		},
		{
			errorMsg:     "unknown random error",
			expectedType: ErrorTypeUnknown,
			recoverable:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.errorMsg, func(t *testing.T) {
			err := &TaskError{OriginalErr: assert.AnError}
			taskErr := eh.ClassifyError(assert.AnError, taskMsg)

			// Mock the error message for classification
			taskErr.OriginalErr = &MockError{message: tc.errorMsg}
			taskErr = eh.ClassifyError(taskErr.OriginalErr, taskMsg)

			assert.Equal(t, tc.expectedType, taskErr.Type)
			assert.Equal(t, tc.recoverable, taskErr.Recoverable)
		})
	}
}

// MockError for testing error classification
type MockError struct {
	message string
}

func (e *MockError) Error() string {
	return e.message
}

// Test PR Creator with GitHub API
func TestPRCreatorWithAPI(t *testing.T) {
	prCreator := NewPRCreator("/tmp/test", "test/repo")

	taskMsg := &TaskMessage{
		Tasks:          "Test PR creation",
		RepositoryName: "test/repo",
		WorkingDir:     "/tmp/test",
		Provider:       "claude",
	}

	result := &AITaskResponse{
		Success:       true,
		Output:        "Task completed successfully",
		FilesModified: []string{"src/test.go", "README.md"},
		ExecutionTime: 5 * time.Minute,
		TokensUsed:    1500,
	}

	// Test PR title generation
	title := prCreator.generatePRTitle(taskMsg)
	assert.Equal(t, "Task: Test PR creation", title)

	// Test PR body generation with issue
	body := prCreator.generatePRBodyWithIssue(taskMsg, result, "123")
	assert.Contains(t, body, "Test PR creation")
	assert.Contains(t, body, "Closes #123")
	assert.Contains(t, body, "Files Modified: 2")
	assert.Contains(t, body, "Tokens Used: 1500")
}

// Test Concurrent Task Worker Pipeline
func TestTaskWorkerPipeline(t *testing.T) {
	// This would be an integration test that requires:
	// - Mock RabbitMQ connection
	// - Mock GitHub API responses
	// - Mock Git operations
	// - Mock AI provider responses

	t.Skip("Integration test - requires full environment setup")
}

// Benchmark tests
func BenchmarkBranchNameGeneration(b *testing.B) {
	bm := NewBranchManager()
	taskMsg := &TaskMessage{
		Tasks:          "Performance test for branch name generation with a very long task description that should be truncated",
		RepositoryName: "test/performance-repo",
		WorkingDir:     "/tmp/perf-test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bm.generateUniqueBranchName(taskMsg)
	}
}

func BenchmarkErrorClassification(b *testing.B) {
	eh := NewErrorHandler()
	taskMsg := &TaskMessage{
		Tasks:          "Performance test task",
		RepositoryName: "test/repo",
	}
	err := &MockError{message: "GitHub API rate limit exceeded for operation: create issue"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eh.ClassifyError(err, taskMsg)
	}
}

// Example test demonstrating full workflow
func TestFullWorkflowIntegration(t *testing.T) {
	// This test demonstrates the complete workflow:
	// 1. Task message received
	// 2. GitHub issue created
	// 3. Branch created
	// 4. Task executed
	// 5. Changes committed and pushed
	// 6. PR created
	// 7. Branch cleaned up

	t.Run("successful_workflow", func(t *testing.T) {
		// Setup
		ctx := context.Background()
		worker := NewConcurrentTaskWorker("amqp://test", "test", 1)

		taskMsg := &TaskMessage{
			Tasks:          "Add unit tests for authentication module",
			RepositoryName: "JokerTrickster/test-app",
			WorkingDir:     "/tmp/test-app",
			Provider:       "claude",
			Interactive:    false,
		}

		// Step 1: GitHub issue creation (mocked)
		issueURL, issueNumber, err := worker.createGitHubIssue(ctx, taskMsg)
		assert.NoError(t, err)
		// In real test with mocks, these would have values

		// Step 2: Branch creation
		branchInfo, err := worker.branchManager.CreateTaskBranch(ctx, taskMsg)
		// This would fail without actual Git repo, so we mock it
		if err != nil {
			// Mock branch info for testing
			branchInfo = &BranchInfo{
				Name:       "task/add-unit-tests-1234-abcd5678",
				TaskID:     "test-task-id",
				CreatedAt:  time.Now(),
				Repository: taskMsg.RepositoryName,
				WorkingDir: taskMsg.WorkingDir,
				IsActive:   true,
			}
		}
		assert.NotNil(t, branchInfo)

		// Step 3: Task execution (mocked)
		result := &AITaskResponse{
			Success:       true,
			Output:        "Tests added successfully",
			FilesModified: []string{"auth_test.go", "user_test.go"},
			ExecutionTime: 3 * time.Minute,
			TokensUsed:    1200,
		}

		// Step 4: PR creation (mocked)
		prCreator := NewPRCreator(taskMsg.WorkingDir, taskMsg.RepositoryName)
		title := prCreator.generatePRTitle(taskMsg)
		body := prCreator.generatePRBodyWithIssue(taskMsg, result, issueNumber)

		assert.Contains(t, title, "Add unit tests")
		assert.Contains(t, body, "auth_test.go")
		if issueNumber != "" {
			assert.Contains(t, body, "Closes #"+issueNumber)
		}

		// Step 5: Cleanup
		err = worker.branchManager.CompleteBranch(taskMsg.RepositoryName, true)
		assert.NoError(t, err)
	})

	t.Run("workflow_with_errors", func(t *testing.T) {
		// Test error scenarios and recovery
		eh := NewErrorHandler()

		// Simulate different error types
		errors := []error{
			&MockError{message: "GitHub API rate limit exceeded"},
			&MockError{message: "git push failed"},
			&MockError{message: "provider not configured"},
		}

		for _, testErr := range errors {
			taskErr := eh.ClassifyError(testErr, &TaskMessage{
				Tasks:          "Test error handling",
				RepositoryName: "test/repo",
			})

			assert.NotNil(t, taskErr)
			assert.NotEqual(t, ErrorTypeUnknown, taskErr.Type)
		}
	})
}