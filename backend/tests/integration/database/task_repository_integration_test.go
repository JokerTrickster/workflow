package database_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ai-git-workbench/internal/domain/repositories"
	"ai-git-workbench/internal/domain/valueobjects"
	"ai-git-workbench/tests/testutils"
)

func TestTaskRepository_Integration_CRUD(t *testing.T) {
	env := testutils.SetupTestEnvironment(t)
	defer env.TearDown(t)

	ctx := context.Background()

	t.Run("Create and retrieve task", func(t *testing.T) {
		env.CleanDatabase(t)

		// Create test task
		task, err := testutils.CreateTestTask("user123", "Integration Test Task")
		require.NoError(t, err)

		// Create in repository
		err = env.TaskRepo.Create(ctx, task)
		require.NoError(t, err)

		// Retrieve the task
		retrieved, err := env.TaskRepo.GetByID(ctx, task.ID())
		require.NoError(t, err)

		// Verify task properties
		testutils.AssertTaskEqual(t, task, retrieved)
		assert.Equal(t, task.Version(), retrieved.Version())
		assert.Equal(t, task.TokensUsed(), retrieved.TokensUsed())
	})

	t.Run("Update task", func(t *testing.T) {
		env.CleanDatabase(t)

		// Create initial task
		task, err := testutils.CreateTestTask("user123", "Original Title")
		require.NoError(t, err)

		err = env.TaskRepo.Create(ctx, task)
		require.NoError(t, err)

		// Update task
		err = task.UpdateTitle("Updated Title")
		require.NoError(t, err)

		err = task.SetMetadata("priority", "high")
		require.NoError(t, err)

		// Update in repository
		err = env.TaskRepo.Update(ctx, task)
		require.NoError(t, err)

		// Retrieve and verify update
		updated, err := env.TaskRepo.GetByID(ctx, task.ID())
		require.NoError(t, err)

		assert.Equal(t, "Updated Title", updated.Title())
		value, exists := updated.GetMetadata("priority")
		assert.True(t, exists)
		assert.Equal(t, "high", value)
		assert.Greater(t, updated.Version(), int64(1))
	})

	t.Run("Delete task", func(t *testing.T) {
		env.CleanDatabase(t)

		// Create task
		task, err := testutils.CreateTestTask("user123", "Task to Delete")
		require.NoError(t, err)

		err = env.TaskRepo.Create(ctx, task)
		require.NoError(t, err)

		// Delete task
		err = env.TaskRepo.Delete(ctx, task.ID())
		require.NoError(t, err)

		// Verify deletion
		_, err = env.TaskRepo.GetByID(ctx, task.ID())
		assert.Error(t, err)
	})

	t.Run("Task with metadata", func(t *testing.T) {
		env.CleanDatabase(t)

		// Create task with metadata
		task, err := testutils.CreateTestTask("user123", "Task with Metadata")
		require.NoError(t, err)

		err = task.SetMetadata("epic", "test-epic")
		require.NoError(t, err)
		err = task.SetMetadata("priority", "high")
		require.NoError(t, err)
		err = task.SetMetadata("assignee", "developer-1")
		require.NoError(t, err)

		// Create in repository
		err = env.TaskRepo.Create(ctx, task)
		require.NoError(t, err)

		// Retrieve and verify metadata
		retrieved, err := env.TaskRepo.GetByID(ctx, task.ID())
		require.NoError(t, err)

		metadata := retrieved.Metadata()
		assert.Len(t, metadata, 3)
		assert.Equal(t, "test-epic", metadata["epic"])
		assert.Equal(t, "high", metadata["priority"])
		assert.Equal(t, "developer-1", metadata["assignee"])
	})
}

func TestTaskRepository_Integration_Queries(t *testing.T) {
	env := testutils.SetupTestEnvironment(t)
	defer env.TearDown(t)

	ctx := context.Background()

	// Setup test data
	env.CleanDatabase(t)

	// Create multiple test tasks
	tasks := []struct {
		userID     string
		title      string
		repository string
		status     string
	}{
		{"user1", "Task 1", "owner/repo1", valueobjects.StatusPending},
		{"user1", "Task 2", "owner/repo1", valueobjects.StatusProcessing},
		{"user1", "Task 3", "owner/repo2", valueobjects.StatusCompleted},
		{"user2", "Task 4", "owner/repo1", valueobjects.StatusPending},
		{"user2", "Task 5", "owner/repo2", valueobjects.StatusFailed},
	}

	for _, taskData := range tasks {
		task, err := testutils.CreateTestTaskWithDetails(
			taskData.userID,
			taskData.title,
			"Test description",
			taskData.repository,
			"test-epic",
			"feature/test",
		)
		require.NoError(t, err)

		// Set status if not pending
		if taskData.status != valueobjects.StatusPending {
			if taskData.status == valueobjects.StatusProcessing {
				err = task.StartProcessing()
				require.NoError(t, err)
			} else if taskData.status == valueobjects.StatusCompleted {
				err = task.StartProcessing()
				require.NoError(t, err)
				err = task.Complete()
				require.NoError(t, err)
			} else if taskData.status == valueobjects.StatusFailed {
				err = task.StartProcessing()
				require.NoError(t, err)
				err = task.Fail()
				require.NoError(t, err)
			}
		}

		err = env.TaskRepo.Create(ctx, task)
		require.NoError(t, err)
	}

	t.Run("Get tasks by user ID", func(t *testing.T) {
		user1ID, err := valueobjects.NewUserID("user1")
		require.NoError(t, err)

		tasks, err := env.TaskRepo.GetByUserID(ctx, user1ID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, tasks, 3)

		// Verify all tasks belong to user1
		for _, task := range tasks {
			assert.Equal(t, "user1", task.UserID().Value())
		}
	})

	t.Run("Get tasks by status", func(t *testing.T) {
		pendingStatus, err := valueobjects.NewTaskStatus(valueobjects.StatusPending)
		require.NoError(t, err)

		tasks, err := env.TaskRepo.GetByStatus(ctx, pendingStatus, 10, 0)
		require.NoError(t, err)
		assert.Len(t, tasks, 2)

		// Verify all tasks have pending status
		for _, task := range tasks {
			assert.Equal(t, valueobjects.StatusPending, task.Status().Value())
		}
	})

	t.Run("Get all tasks with pagination", func(t *testing.T) {
		// Get first page
		tasks, err := env.TaskRepo.GetAll(ctx, 3, 0)
		require.NoError(t, err)
		assert.Len(t, tasks, 3)

		// Get second page
		tasks, err = env.TaskRepo.GetAll(ctx, 3, 3)
		require.NoError(t, err)
		assert.Len(t, tasks, 2)
	})
}

func TestTaskRepository_Integration_Statistics(t *testing.T) {
	env := testutils.SetupTestEnvironment(t)
	defer env.TearDown(t)

	ctx := context.Background()
	env.CleanDatabase(t)

	userID, err := valueobjects.NewUserID("stats-user")
	require.NoError(t, err)

	// Create tasks with different statuses
	taskData := []struct {
		title      string
		status     string
		tokens     int
	}{
		{"Completed Task 1", valueobjects.StatusCompleted, 100},
		{"Completed Task 2", valueobjects.StatusCompleted, 150},
		{"Failed Task 1", valueobjects.StatusFailed, 50},
		{"Pending Task 1", valueobjects.StatusPending, 0},
		{"Processing Task 1", valueobjects.StatusProcessing, 75},
	}

	for _, data := range taskData {
		task, err := testutils.CreateTestTask("stats-user", data.title)
		require.NoError(t, err)

		// Set status and tokens
		if data.status == valueobjects.StatusProcessing {
			err = task.StartProcessing()
			require.NoError(t, err)
		} else if data.status == valueobjects.StatusCompleted {
			err = task.StartProcessing()
			require.NoError(t, err)
			err = task.Complete()
			require.NoError(t, err)
		} else if data.status == valueobjects.StatusFailed {
			err = task.StartProcessing()
			require.NoError(t, err)
			err = task.Fail()
			require.NoError(t, err)
		}

		if data.tokens > 0 {
			err = task.AddTokensUsed(data.tokens)
			require.NoError(t, err)
		}

		err = env.TaskRepo.Create(ctx, task)
		require.NoError(t, err)
	}

	// Get statistics
	stats, err := env.TaskRepo.GetUserTaskStatistics(ctx, userID)
	require.NoError(t, err)

	assert.Equal(t, userID, stats.UserID)
	assert.Equal(t, 5, stats.TotalTasks)
	assert.Equal(t, 2, stats.CompletedTasks)
	assert.Equal(t, 1, stats.FailedTasks)
	assert.Equal(t, 1, stats.PendingTasks)
	assert.Equal(t, 1, stats.ProcessingTasks)
	assert.Equal(t, 375, stats.TotalTokensUsed) // 100 + 150 + 50 + 0 + 75
	assert.Equal(t, 75.0, stats.AverageTokensPerTask)
	assert.Equal(t, 0.4, stats.CompletionRate) // 2/5
}

func TestTaskRepository_Integration_Concurrent(t *testing.T) {
	env := testutils.SetupTestEnvironment(t)
	defer env.TearDown(t)

	ctx := context.Background()
	env.CleanDatabase(t)

	// Test concurrent task creation
	t.Run("Concurrent task creation", func(t *testing.T) {
		const numTasks = 10
		done := make(chan error, numTasks)

		for i := 0; i < numTasks; i++ {
			go func(taskNum int) {
				task, err := testutils.CreateTestTask("concurrent-user", fmt.Sprintf("Concurrent Task %d", taskNum))
				if err != nil {
					done <- err
					return
				}

				err = env.TaskRepo.Create(ctx, task)
				done <- err
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numTasks; i++ {
			err := <-done
			require.NoError(t, err)
		}

		// Verify all tasks were created
		userID, err := valueobjects.NewUserID("concurrent-user")
		require.NoError(t, err)

		tasks, err := env.TaskRepo.GetByUserID(ctx, userID, 20, 0)
		require.NoError(t, err)
		assert.Len(t, tasks, numTasks)
	})

	// Test concurrent updates to the same task
	t.Run("Concurrent task updates", func(t *testing.T) {
		env.CleanDatabase(t)

		// Create a task
		task, err := testutils.CreateTestTask("update-user", "Task for Updates")
		require.NoError(t, err)

		err = env.TaskRepo.Create(ctx, task)
		require.NoError(t, err)

		const numUpdates = 5
		done := make(chan error, numUpdates)

		for i := 0; i < numUpdates; i++ {
			go func(updateNum int) {
				// Get fresh copy of task
				freshTask, err := env.TaskRepo.GetByID(ctx, task.ID())
				if err != nil {
					done <- err
					return
				}

				err = freshTask.SetMetadata(fmt.Sprintf("key-%d", updateNum), fmt.Sprintf("value-%d", updateNum))
				if err != nil {
					done <- err
					return
				}

				err = env.TaskRepo.Update(ctx, freshTask)
				done <- err
			}(i)
		}

		// Wait for all goroutines to complete
		successCount := 0
		for i := 0; i < numUpdates; i++ {
			err := <-done
			if err == nil {
				successCount++
			}
		}

		// At least one update should succeed
		assert.Greater(t, successCount, 0)

		// Verify final state
		finalTask, err := env.TaskRepo.GetByID(ctx, task.ID())
		require.NoError(t, err)
		assert.Greater(t, finalTask.Version(), task.Version())
	})
}

func TestTaskRepository_Integration_Transaction(t *testing.T) {
	env := testutils.SetupTestEnvironment(t)
	defer env.TearDown(t)

	ctx := context.Background()
	env.CleanDatabase(t)

	t.Run("Successful transaction", func(t *testing.T) {
		task, err := testutils.CreateTestTask("tx-user", "Transaction Test")
		require.NoError(t, err)

		err = env.TaskRepo.WithTransaction(ctx, func(repo repositories.TaskRepository) error {
			// Create task within transaction
			if err := repo.Create(ctx, task); err != nil {
				return err
			}

			// Update task within same transaction
			err := task.UpdateTitle("Updated in Transaction")
			if err != nil {
				return err
			}

			return repo.Update(ctx, task)
		})

		require.NoError(t, err)

		// Verify both operations succeeded
		retrieved, err := env.TaskRepo.GetByID(ctx, task.ID())
		require.NoError(t, err)
		assert.Equal(t, "Updated in Transaction", retrieved.Title())
	})

	t.Run("Failed transaction rollback", func(t *testing.T) {
		task, err := testutils.CreateTestTask("tx-user-2", "Rollback Test")
		require.NoError(t, err)

		err = env.TaskRepo.WithTransaction(ctx, func(repo repositories.TaskRepository) error {
			// Create task within transaction
			if err := repo.Create(ctx, task); err != nil {
				return err
			}

			// Force transaction to fail
			return errors.New("forced failure")
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "forced failure")

		// Verify task was not created due to rollback
		_, err = env.TaskRepo.GetByID(ctx, task.ID())
		assert.Error(t, err)
	})
}

func TestTaskRepository_Integration_Performance(t *testing.T) {
	env := testutils.SetupTestEnvironment(t)
	defer env.TearDown(t)

	ctx := context.Background()
	env.CleanDatabase(t)

	t.Run("Bulk operations performance", func(t *testing.T) {
		const numTasks = 100
		start := time.Now()

		// Create multiple tasks
		for i := 0; i < numTasks; i++ {
			task, err := testutils.CreateTestTask("perf-user", fmt.Sprintf("Performance Task %d", i))
			require.NoError(t, err)

			err = env.TaskRepo.Create(ctx, task)
			require.NoError(t, err)
		}

		createDuration := time.Since(start)
		t.Logf("Created %d tasks in %v (%.2f tasks/sec)", numTasks, createDuration, float64(numTasks)/createDuration.Seconds())

		// Query performance
		start = time.Now()
		userID, err := valueobjects.NewUserID("perf-user")
		require.NoError(t, err)

		tasks, err := env.TaskRepo.GetByUserID(ctx, userID, numTasks, 0)
		require.NoError(t, err)
		assert.Len(t, tasks, numTasks)

		queryDuration := time.Since(start)
		t.Logf("Queried %d tasks in %v", numTasks, queryDuration)

		// Performance assertions
		assert.Less(t, createDuration, 5*time.Second, "Task creation should be fast")
		assert.Less(t, queryDuration, 1*time.Second, "Task query should be fast")
	})
}

// Benchmark tests
func BenchmarkTaskRepository_Create(b *testing.B) {
	env := testutils.SetupTestEnvironment(b)
	defer env.TearDown(b)

	ctx := context.Background()
	env.CleanDatabase(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task, err := testutils.CreateTestTask("bench-user", fmt.Sprintf("Benchmark Task %d", i))
		if err != nil {
			b.Fatal(err)
		}

		err = env.TaskRepo.Create(ctx, task)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTaskRepository_GetByID(b *testing.B) {
	env := testutils.SetupTestEnvironment(b)
	defer env.TearDown(b)

	ctx := context.Background()
	env.CleanDatabase(b)

	// Create a test task
	task, err := testutils.CreateTestTask("bench-user", "Benchmark Task")
	if err != nil {
		b.Fatal(err)
	}

	err = env.TaskRepo.Create(ctx, task)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := env.TaskRepo.GetByID(ctx, task.ID())
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTaskRepository_Update(b *testing.B) {
	env := testutils.SetupTestEnvironment(b)
	defer env.TearDown(b)

	ctx := context.Background()
	env.CleanDatabase(b)

	// Create a test task
	task, err := testutils.CreateTestTask("bench-user", "Benchmark Task")
	if err != nil {
		b.Fatal(err)
	}

	err = env.TaskRepo.Create(ctx, task)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := task.SetMetadata(fmt.Sprintf("key-%d", i), fmt.Sprintf("value-%d", i))
		if err != nil {
			b.Fatal(err)
		}

		err = env.TaskRepo.Update(ctx, task)
		if err != nil {
			b.Fatal(err)
		}
	}
}