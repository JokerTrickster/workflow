package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ai-git-workbench/internal/domain/entities"
	"ai-git-workbench/internal/domain/valueobjects"
	"ai-git-workbench/tests/testutils"
)

func TestTask_Creation(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		title       string
		description string
		repository  string
		epic        string
		branch      string
		wantErr     bool
		errContains string
	}{
		{
			name:        "Valid task creation",
			userID:      "user123",
			title:       "Test Task",
			description: "A test task description",
			repository:  "owner/repo",
			epic:        "epic-1",
			branch:      "feature/test",
			wantErr:     false,
		},
		{
			name:        "Empty title should fail",
			userID:      "user123",
			title:       "",
			description: "A test task description",
			repository:  "owner/repo",
			epic:        "epic-1",
			branch:      "feature/test",
			wantErr:     true,
			errContains: "title cannot be empty",
		},
		{
			name:        "Title too long should fail",
			userID:      "user123",
			title:       string(make([]byte, 501)), // 501 characters
			description: "A test task description",
			repository:  "owner/repo",
			epic:        "epic-1",
			branch:      "feature/test",
			wantErr:     true,
			errContains: "title cannot exceed 500 characters",
		},
		{
			name:        "Description too long should fail",
			userID:      "user123",
			title:       "Test Task",
			description: string(make([]byte, 5001)), // 5001 characters
			repository:  "owner/repo",
			epic:        "epic-1",
			branch:      "feature/test",
			wantErr:     true,
			errContains: "description cannot exceed 5000 characters",
		},
		{
			name:        "Empty epic should fail",
			userID:      "user123",
			title:       "Test Task",
			description: "A test task description",
			repository:  "owner/repo",
			epic:        "",
			branch:      "feature/test",
			wantErr:     true,
			errContains: "epic cannot be empty",
		},
		{
			name:        "Invalid repository format should fail",
			userID:      "user123",
			title:       "Test Task",
			description: "A test task description",
			repository:  "invalid-repo-format",
			epic:        "epic-1",
			branch:      "feature/test",
			wantErr:     true,
			errContains: "repository path",
		},
		{
			name:        "Invalid branch name should fail",
			userID:      "user123",
			title:       "Test Task",
			description: "A test task description",
			repository:  "owner/repo",
			epic:        "epic-1",
			branch:      "invalid branch name",
			wantErr:     true,
			errContains: "branch name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := testutils.CreateTestTaskWithDetails(
				tt.userID, tt.title, tt.description, tt.repository, tt.epic, tt.branch,
			)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, task)
			} else {
				require.NoError(t, err)
				require.NotNil(t, task)
				
				assert.Equal(t, tt.title, task.Title())
				assert.Equal(t, tt.description, task.Description())
				assert.Equal(t, valueobjects.StatusPending, task.Status().Value())
				assert.Equal(t, 0, task.TokensUsed())
				assert.Equal(t, int64(1), task.Version())
				assert.NotZero(t, task.CreatedAt())
				assert.NotZero(t, task.UpdatedAt())
				assert.Nil(t, task.StartedAt())
				assert.Nil(t, task.CompletedAt())
				assert.False(t, task.IsActive())
				assert.False(t, task.IsCompleted())
				assert.True(t, task.CanBeModified())
			}
		})
	}
}

func TestTask_StatusTransitions(t *testing.T) {
	task, err := testutils.CreateTestTask("user123", "Test Task")
	require.NoError(t, err)

	tests := []struct {
		name           string
		initialStatus  string
		operation      func(*entities.Task) error
		expectedStatus string
		shouldFail     bool
	}{
		{
			name:           "Pending to Processing",
			initialStatus:  valueobjects.StatusPending,
			operation:      (*entities.Task).StartProcessing,
			expectedStatus: valueobjects.StatusProcessing,
			shouldFail:     false,
		},
		{
			name:          "Pending to Cancelled",
			initialStatus: valueobjects.StatusPending,
			operation:     (*entities.Task).Cancel,
			expectedStatus: valueobjects.StatusCancelled,
			shouldFail:    false,
		},
		{
			name:          "Pending to Completed should fail",
			initialStatus: valueobjects.StatusPending,
			operation:     (*entities.Task).Complete,
			shouldFail:    true,
		},
		{
			name:          "Pending to Failed should fail",
			initialStatus: valueobjects.StatusPending,
			operation:     (*entities.Task).Fail,
			shouldFail:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset task to pending status for each test
			freshTask, err := testutils.CreateTestTask("user123", "Test Task")
			require.NoError(t, err)

			initialVersion := freshTask.Version()
			err = tt.operation(freshTask)

			if tt.shouldFail {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid status transition")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, freshTask.Status().Value())
				assert.Greater(t, freshTask.Version(), initialVersion)
				assert.True(t, freshTask.UpdatedAt().After(freshTask.CreatedAt()))
			}
		})
	}
}

func TestTask_ProcessingFlow(t *testing.T) {
	task, err := testutils.CreateTestTask("user123", "Test Task")
	require.NoError(t, err)

	// Start processing
	err = task.StartProcessing()
	require.NoError(t, err)
	
	assert.Equal(t, valueobjects.StatusProcessing, task.Status().Value())
	assert.NotNil(t, task.StartedAt())
	assert.True(t, task.IsActive())
	assert.False(t, task.CanBeModified())

	// Complete the task
	err = task.Complete()
	require.NoError(t, err)
	
	assert.Equal(t, valueobjects.StatusCompleted, task.Status().Value())
	assert.NotNil(t, task.CompletedAt())
	assert.True(t, task.IsCompleted())
	assert.False(t, task.IsActive())
	assert.False(t, task.CanBeModified())

	// Verify duration calculation
	duration := task.GetDuration()
	require.NotNil(t, duration)
	assert.Greater(t, *duration, time.Duration(0))
}

func TestTask_FailureFlow(t *testing.T) {
	task, err := testutils.CreateTestTask("user123", "Test Task")
	require.NoError(t, err)

	// Start processing
	err = task.StartProcessing()
	require.NoError(t, err)

	// Fail the task
	err = task.Fail()
	require.NoError(t, err)
	
	assert.Equal(t, valueobjects.StatusFailed, task.Status().Value())
	assert.True(t, task.IsFailed())
	assert.False(t, task.IsActive())
	assert.False(t, task.CanBeModified())

	// Restart the task
	err = task.Restart()
	require.NoError(t, err)
	
	assert.Equal(t, valueobjects.StatusPending, task.Status().Value())
	assert.Nil(t, task.StartedAt())
	assert.Nil(t, task.CompletedAt())
	assert.False(t, task.IsActive())
	assert.True(t, task.CanBeModified())
}

func TestTask_CancellationFlow(t *testing.T) {
	task, err := testutils.CreateTestTask("user123", "Test Task")
	require.NoError(t, err)

	// Cancel the task
	err = task.Cancel()
	require.NoError(t, err)
	
	assert.Equal(t, valueobjects.StatusCancelled, task.Status().Value())
	assert.True(t, task.IsCancelled())
	assert.False(t, task.IsActive())
	assert.False(t, task.CanBeModified())

	// Restart the cancelled task
	err = task.Restart()
	require.NoError(t, err)
	
	assert.Equal(t, valueobjects.StatusPending, task.Status().Value())
	assert.True(t, task.CanBeModified())
}

func TestTask_UpdateOperations(t *testing.T) {
	task, err := testutils.CreateTestTask("user123", "Original Title")
	require.NoError(t, err)

	originalVersion := task.Version()

	t.Run("Update title", func(t *testing.T) {
		err := task.UpdateTitle("Updated Title")
		require.NoError(t, err)
		
		assert.Equal(t, "Updated Title", task.Title())
		assert.Greater(t, task.Version(), originalVersion)
	})

	t.Run("Update title with invalid input", func(t *testing.T) {
		err := task.UpdateTitle("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "title cannot be empty")
	})

	t.Run("Update description", func(t *testing.T) {
		err := task.UpdateDescription("Updated description")
		require.NoError(t, err)
		
		assert.Equal(t, "Updated description", task.Description())
	})

	t.Run("Add tokens used", func(t *testing.T) {
		initialTokens := task.TokensUsed()
		err := task.AddTokensUsed(100)
		require.NoError(t, err)
		
		assert.Equal(t, initialTokens+100, task.TokensUsed())
	})

	t.Run("Add negative tokens should fail", func(t *testing.T) {
		err := task.AddTokensUsed(-50)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tokens used cannot be negative")
	})
}

func TestTask_MetadataOperations(t *testing.T) {
	task, err := testutils.CreateTestTask("user123", "Test Task")
	require.NoError(t, err)

	t.Run("Set metadata", func(t *testing.T) {
		err := task.SetMetadata("priority", "high")
		require.NoError(t, err)
		
		value, exists := task.GetMetadata("priority")
		assert.True(t, exists)
		assert.Equal(t, "high", value)
	})

	t.Run("Set empty key should fail", func(t *testing.T) {
		err := task.SetMetadata("", "value")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metadata key cannot be empty")
	})

	t.Run("Get non-existent metadata", func(t *testing.T) {
		value, exists := task.GetMetadata("nonexistent")
		assert.False(t, exists)
		assert.Empty(t, value)
	})

	t.Run("Remove metadata", func(t *testing.T) {
		err := task.SetMetadata("temp", "value")
		require.NoError(t, err)
		
		task.RemoveMetadata("temp")
		value, exists := task.GetMetadata("temp")
		assert.False(t, exists)
		assert.Empty(t, value)
	})

	t.Run("Get all metadata", func(t *testing.T) {
		err := task.SetMetadata("key1", "value1")
		require.NoError(t, err)
		err = task.SetMetadata("key2", "value2")
		require.NoError(t, err)
		
		metadata := task.Metadata()
		assert.Len(t, metadata, 3) // Including "priority" from earlier test
		assert.Equal(t, "value1", metadata["key1"])
		assert.Equal(t, "value2", metadata["key2"])
		assert.Equal(t, "high", metadata["priority"])
	})
}

func TestTask_BusinessRuleQueries(t *testing.T) {
	task, err := testutils.CreateTestTask("user123", "Test Task")
	require.NoError(t, err)

	t.Run("Initial state", func(t *testing.T) {
		assert.False(t, task.IsActive())
		assert.False(t, task.IsCompleted())
		assert.False(t, task.IsFailed())
		assert.False(t, task.IsCancelled())
		assert.True(t, task.CanBeModified())
		assert.Nil(t, task.GetDuration())
	})

	t.Run("Processing state", func(t *testing.T) {
		err := task.StartProcessing()
		require.NoError(t, err)
		
		assert.True(t, task.IsActive())
		assert.False(t, task.IsCompleted())
		assert.False(t, task.IsFailed())
		assert.False(t, task.IsCancelled())
		assert.False(t, task.CanBeModified())
		assert.NotNil(t, task.GetDuration())
	})

	t.Run("Long running check", func(t *testing.T) {
		// Task should not be long running immediately
		assert.False(t, task.IsLongRunning(1*time.Hour))
		
		// For testing purposes, we'll check with a very short threshold
		time.Sleep(1 * time.Millisecond)
		assert.True(t, task.IsLongRunning(1*time.Nanosecond))
	})

	t.Run("Completed state", func(t *testing.T) {
		err := task.Complete()
		require.NoError(t, err)
		
		assert.False(t, task.IsActive())
		assert.True(t, task.IsCompleted())
		assert.False(t, task.IsFailed())
		assert.False(t, task.IsCancelled())
		assert.False(t, task.CanBeModified())
		assert.False(t, task.IsLongRunning(1*time.Hour)) // Completed tasks are not long running
	})
}

func TestTask_ReconstructionFromPersistence(t *testing.T) {
	// Create original task
	original, err := testutils.CreateTestTask("user123", "Test Task")
	require.NoError(t, err)

	// Simulate some operations
	err = original.StartProcessing()
	require.NoError(t, err)
	err = original.SetMetadata("key", "value")
	require.NoError(t, err)

	// Reconstruct from persistence
	reconstructed := entities.ReconstructTask(
		original.ID(),
		original.UserID(),
		original.Title(),
		original.Description(),
		original.Status(),
		original.Repository(),
		original.Epic(),
		original.Branch(),
		original.TokensUsed(),
		original.CreatedAt(),
		original.UpdatedAt(),
		original.StartedAt(),
		original.CompletedAt(),
		original.Metadata(),
		original.Version(),
	)

	// Verify reconstruction
	testutils.AssertTaskEqual(t, original, reconstructed)
	assert.Equal(t, original.Version(), reconstructed.Version())
	assert.Equal(t, original.TokensUsed(), reconstructed.TokensUsed())
	assert.Equal(t, original.Metadata(), reconstructed.Metadata())
	assert.Equal(t, original.StartedAt(), reconstructed.StartedAt())
	assert.Equal(t, original.CompletedAt(), reconstructed.CompletedAt())
}

func TestTask_ConcurrentModification(t *testing.T) {
	task, err := testutils.CreateTestTask("user123", "Test Task")
	require.NoError(t, err)

	// Simulate concurrent modifications by checking version increments
	initialVersion := task.Version()

	err = task.UpdateTitle("Title 1")
	require.NoError(t, err)
	assert.Equal(t, initialVersion+1, task.Version())

	err = task.SetMetadata("key", "value")
	require.NoError(t, err)
	assert.Equal(t, initialVersion+2, task.Version())

	err = task.StartProcessing()
	require.NoError(t, err)
	assert.Equal(t, initialVersion+3, task.Version())
}

// Benchmark tests
func BenchmarkTask_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := testutils.CreateTestTask("user123", "Benchmark Task")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTask_StatusTransition(b *testing.B) {
	task, err := testutils.CreateTestTask("user123", "Benchmark Task")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create new task for each iteration to avoid status conflicts
		freshTask, err := testutils.CreateTestTask("user123", "Benchmark Task")
		if err != nil {
			b.Fatal(err)
		}
		
		err = freshTask.StartProcessing()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTask_MetadataOperations(b *testing.B) {
	task, err := testutils.CreateTestTask("user123", "Benchmark Task")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := task.SetMetadata("benchmark", "value")
		if err != nil {
			b.Fatal(err)
		}
		
		_, exists := task.GetMetadata("benchmark")
		if !exists {
			b.Fatal("metadata should exist")
		}
	}
}