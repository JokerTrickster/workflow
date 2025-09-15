package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ai-git-workbench/internal/domain/valueobjects"
)

func TestTaskID_Creation(t *testing.T) {
	t.Run("Generate unique task IDs", func(t *testing.T) {
		id1 := valueobjects.GenerateTaskID()
		id2 := valueobjects.GenerateTaskID()

		assert.NotEqual(t, id1.Value(), id2.Value(), "Generated IDs should be unique")
		assert.NotEmpty(t, id1.Value(), "ID should not be empty")
		assert.NotEmpty(t, id2.Value(), "ID should not be empty")
	})

	t.Run("Parse valid task ID", func(t *testing.T) {
		originalID := valueobjects.GenerateTaskID()
		
		parsedID, err := valueobjects.ParseTaskID(originalID.Value())
		require.NoError(t, err)
		
		assert.Equal(t, originalID.Value(), parsedID.Value())
	})

	t.Run("Parse invalid task ID", func(t *testing.T) {
		_, err := valueobjects.ParseTaskID("invalid-id")
		assert.Error(t, err, "Should reject invalid task ID format")
	})
}

func TestUserID_Creation(t *testing.T) {
	t.Run("Valid user ID creation", func(t *testing.T) {
		userID, err := valueobjects.NewUserID("user123")
		require.NoError(t, err)
		
		assert.Equal(t, "user123", userID.Value())
	})

	t.Run("Empty user ID should fail", func(t *testing.T) {
		_, err := valueobjects.NewUserID("")
		assert.Error(t, err, "Empty user ID should be rejected")
	})

	t.Run("User ID too long should fail", func(t *testing.T) {
		longID := string(make([]byte, 256)) // 256 characters
		_, err := valueobjects.NewUserID(longID)
		assert.Error(t, err, "Overly long user ID should be rejected")
	})
}

func TestTaskStatus_Values(t *testing.T) {
	t.Run("Valid status values", func(t *testing.T) {
		validStatuses := []string{
			valueobjects.StatusPending,
			valueobjects.StatusProcessing,
			valueobjects.StatusCompleted,
			valueobjects.StatusFailed,
			valueobjects.StatusCancelled,
		}

		for _, status := range validStatuses {
			taskStatus, err := valueobjects.NewTaskStatus(status)
			require.NoError(t, err, "Status %s should be valid", status)
			assert.Equal(t, status, taskStatus.Value())
		}
	})

	t.Run("Invalid status should fail", func(t *testing.T) {
		_, err := valueobjects.NewTaskStatus("invalid-status")
		assert.Error(t, err, "Invalid status should be rejected")
	})

	t.Run("Status transitions", func(t *testing.T) {
		pendingStatus, err := valueobjects.NewTaskStatus(valueobjects.StatusPending)
		require.NoError(t, err)

		processingStatus, err := valueobjects.NewTaskStatus(valueobjects.StatusProcessing)
		require.NoError(t, err)

		completedStatus, err := valueobjects.NewTaskStatus(valueobjects.StatusCompleted)
		require.NoError(t, err)

		// Valid transitions
		assert.True(t, pendingStatus.CanTransitionTo(processingStatus), "Pending can transition to Processing")
		assert.True(t, processingStatus.CanTransitionTo(completedStatus), "Processing can transition to Completed")

		// Invalid transitions  
		assert.False(t, completedStatus.CanTransitionTo(pendingStatus), "Completed cannot transition to Pending")
	})
}

func TestRepositoryPath_Validation(t *testing.T) {
	t.Run("Valid repository path", func(t *testing.T) {
		repo, err := valueobjects.NewRepositoryPath("owner/repo")
		require.NoError(t, err)
		
		assert.Equal(t, "owner/repo", repo.Value())
	})

	t.Run("Invalid repository path format", func(t *testing.T) {
		invalidPaths := []string{
			"",
			"invalid",
			"owner//repo",
			"owner/",
			"/repo",
		}

		for _, path := range invalidPaths {
			_, err := valueobjects.NewRepositoryPath(path)
			assert.Error(t, err, "Path %s should be invalid", path)
		}
	})
}

func TestBranchName_Validation(t *testing.T) {
	t.Run("Valid branch names", func(t *testing.T) {
		validBranches := []string{
			"main",
			"feature/test",
			"bugfix/issue-123",
			"release/v1.0.0",
		}

		for _, branch := range validBranches {
			branchName, err := valueobjects.NewBranchName(branch)
			require.NoError(t, err, "Branch %s should be valid", branch)
			assert.Equal(t, branch, branchName.Value())
		}
	})

	t.Run("Invalid branch names", func(t *testing.T) {
		invalidBranches := []string{
			"",
			"feature..test",
			"feature/",
			"feature/.test",
			string(make([]byte, 256)), // Too long
		}

		for _, branch := range invalidBranches {
			_, err := valueobjects.NewBranchName(branch)
			assert.Error(t, err, "Branch %s should be invalid", branch)
		}
	})
}