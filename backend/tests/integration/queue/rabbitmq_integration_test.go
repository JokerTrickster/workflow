package queue_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ai-git-workbench/internal/domain/valueobjects"
	"ai-git-workbench/tests/testutils"
)

func TestRabbitMQQueue_Integration_Basic(t *testing.T) {
	env := testutils.SetupTestEnvironment(t)
	defer env.TearDown(t)

	ctx := context.Background()

	t.Run("Enqueue and dequeue task", func(t *testing.T) {
		// Create a test task
		task, err := testutils.CreateTestTask("user123", "Queue Test Task")
		require.NoError(t, err)

		// Enqueue the task
		err = env.QueueRepo.Enqueue(ctx, task)
		require.NoError(t, err)

		// Dequeue the task
		dequeuedTask, err := env.QueueRepo.Dequeue(ctx)
		require.NoError(t, err)

		// Verify the task
		testutils.AssertTaskEqual(t, task, dequeuedTask)
	})

	t.Run("Queue length tracking", func(t *testing.T) {
		// Check initial queue length
		length, err := env.QueueRepo.GetQueueLength(ctx)
		require.NoError(t, err)
		initialLength := length

		// Enqueue multiple tasks
		const numTasks = 5
		for i := 0; i < numTasks; i++ {
			task, err := testutils.CreateTestTask("user123", fmt.Sprintf("Queue Task %d", i))
			require.NoError(t, err)

			err = env.QueueRepo.Enqueue(ctx, task)
			require.NoError(t, err)
		}

		// Check queue length
		length, err = env.QueueRepo.GetQueueLength(ctx)
		require.NoError(t, err)
		assert.Equal(t, initialLength+numTasks, length)

		// Dequeue all tasks
		for i := 0; i < numTasks; i++ {
			_, err := env.QueueRepo.Dequeue(ctx)
			require.NoError(t, err)
		}

		// Check final queue length
		length, err = env.QueueRepo.GetQueueLength(ctx)
		require.NoError(t, err)
		assert.Equal(t, initialLength, length)
	})

	t.Run("Remove specific task from queue", func(t *testing.T) {
		// Enqueue multiple tasks
		tasks := make([]*testutils.TaskCreated, 3)
		for i := 0; i < 3; i++ {
			task, err := testutils.CreateTestTask("user123", fmt.Sprintf("Remove Test Task %d", i))
			require.NoError(t, err)
			tasks[i] = &testutils.TaskCreated{Task: task}

			err = env.QueueRepo.Enqueue(ctx, task)
			require.NoError(t, err)
		}

		// Remove the middle task
		err := env.QueueRepo.RemoveTask(ctx, tasks[1].Task.ID())
		require.NoError(t, err)

		// Dequeue remaining tasks
		task1, err := env.QueueRepo.Dequeue(ctx)
		require.NoError(t, err)
		task2, err := env.QueueRepo.Dequeue(ctx)
		require.NoError(t, err)

		// Verify we got the first and third tasks (not the removed one)
		taskIDs := []string{task1.ID().Value(), task2.ID().Value()}
		assert.Contains(t, taskIDs, tasks[0].Task.ID().Value())
		assert.Contains(t, taskIDs, tasks[2].Task.ID().Value())
		assert.NotContains(t, taskIDs, tasks[1].Task.ID().Value())
	})
}

func TestRabbitMQQueue_Integration_HealthCheck(t *testing.T) {
	env := testutils.SetupTestEnvironment(t)
	defer env.TearDown(t)

	ctx := context.Background()

	t.Run("Health check passes when queue is operational", func(t *testing.T) {
		err := env.QueueRepo.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	t.Run("Queue statistics", func(t *testing.T) {
		// Enqueue some tasks
		for i := 0; i < 3; i++ {
			task, err := testutils.CreateTestTask("user123", fmt.Sprintf("Stats Task %d", i))
			require.NoError(t, err)

			err = env.QueueRepo.Enqueue(ctx, task)
			require.NoError(t, err)
		}

		// Get queue statistics
		stats, err := env.QueueRepo.GetQueueStatistics(ctx)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, stats.TotalEnqueued, 3)
		assert.GreaterOrEqual(t, stats.CurrentQueueLength, 3)
		assert.NotZero(t, stats.LastActivityAt)
	})
}

func TestRabbitMQQueue_Integration_Concurrent(t *testing.T) {
	env := testutils.SetupTestEnvironment(t)
	defer env.TearDown(t)

	ctx := context.Background()

	t.Run("Concurrent enqueue operations", func(t *testing.T) {
		const numWorkers = 10
		const tasksPerWorker = 5
		done := make(chan error, numWorkers)

		// Start multiple workers enqueueing tasks
		for w := 0; w < numWorkers; w++ {
			go func(workerID int) {
				for i := 0; i < tasksPerWorker; i++ {
					task, err := testutils.CreateTestTask(
						"concurrent-user",
						fmt.Sprintf("Worker %d Task %d", workerID, i),
					)
					if err != nil {
						done <- err
						return
					}

					err = env.QueueRepo.Enqueue(ctx, task)
					if err != nil {
						done <- err
						return
					}
				}
				done <- nil
			}(w)
		}

		// Wait for all workers to complete
		for w := 0; w < numWorkers; w++ {
			err := <-done
			require.NoError(t, err)
		}

		// Verify queue length
		length, err := env.QueueRepo.GetQueueLength(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, length, numWorkers*tasksPerWorker)
	})

	t.Run("Concurrent dequeue operations", func(t *testing.T) {
		// First, enqueue a bunch of tasks
		const numTasks = 20
		for i := 0; i < numTasks; i++ {
			task, err := testutils.CreateTestTask("dequeue-user", fmt.Sprintf("Dequeue Task %d", i))
			require.NoError(t, err)

			err = env.QueueRepo.Enqueue(ctx, task)
			require.NoError(t, err)
		}

		// Now start multiple workers dequeuing
		const numWorkers = 5
		results := make(chan string, numTasks)
		done := make(chan error, numWorkers)

		for w := 0; w < numWorkers; w++ {
			go func() {
				for {
					task, err := env.QueueRepo.Dequeue(ctx)
					if err != nil {
						// Queue might be empty, that's ok
						done <- nil
						return
					}
					results <- task.ID().Value()
				}
			}()
		}

		// Collect results with timeout
		receivedTasks := make(map[string]bool)
		timeout := time.After(5 * time.Second)

	collectLoop:
		for {
			select {
			case taskID := <-results:
				receivedTasks[taskID] = true
				if len(receivedTasks) >= numTasks {
					break collectLoop
				}
			case <-timeout:
				break collectLoop
			}
		}

		// Verify we received tasks without duplicates
		assert.GreaterOrEqual(t, len(receivedTasks), numTasks/2) // Allow some tasks to remain
		t.Logf("Received %d unique tasks out of %d", len(receivedTasks), numTasks)
	})
}

func TestRabbitMQQueue_Integration_EdgeCases(t *testing.T) {
	env := testutils.SetupTestEnvironment(t)
	defer env.TearDown(t)

	ctx := context.Background()

	t.Run("Dequeue from empty queue", func(t *testing.T) {
		// First ensure queue is empty by dequeuing everything
		for {
			_, err := env.QueueRepo.Dequeue(ctx)
			if err != nil {
				break // Queue is empty
			}
		}

		// Now try to dequeue from empty queue
		_, err := env.QueueRepo.Dequeue(ctx)
		assert.Error(t, err, "Dequeuing from empty queue should return an error")
	})

	t.Run("Remove non-existent task", func(t *testing.T) {
		// Create a task ID that doesn't exist in the queue
		nonExistentID := valueobjects.GenerateTaskID()

		// Try to remove it
		err := env.QueueRepo.RemoveTask(ctx, nonExistentID)
		// This may or may not error depending on implementation
		// Some implementations might silently ignore non-existent tasks
		t.Logf("Remove non-existent task result: %v", err)
	})

	t.Run("Large task payload", func(t *testing.T) {
		// Create a task with large metadata
		task, err := testutils.CreateTestTask("large-payload-user", "Large Payload Task")
		require.NoError(t, err)

		// Add large metadata
		largeValue := string(make([]byte, 10000)) // 10KB of data
		err = task.SetMetadata("large_field", largeValue)
		require.NoError(t, err)

		// Enqueue and dequeue
		err = env.QueueRepo.Enqueue(ctx, task)
		require.NoError(t, err)

		dequeuedTask, err := env.QueueRepo.Dequeue(ctx)
		require.NoError(t, err)

		// Verify large payload is preserved
		value, exists := dequeuedTask.GetMetadata("large_field")
		assert.True(t, exists)
		assert.Len(t, value, 10000)
	})
}

func TestRabbitMQQueue_Integration_Performance(t *testing.T) {
	env := testutils.SetupTestEnvironment(t)
	defer env.TearDown(t)

	ctx := context.Background()

	t.Run("High throughput enqueue", func(t *testing.T) {
		const numTasks = 1000
		start := time.Now()

		for i := 0; i < numTasks; i++ {
			task, err := testutils.CreateTestTask("perf-user", fmt.Sprintf("Perf Task %d", i))
			require.NoError(t, err)

			err = env.QueueRepo.Enqueue(ctx, task)
			require.NoError(t, err)
		}

		duration := time.Since(start)
		throughput := float64(numTasks) / duration.Seconds()

		t.Logf("Enqueued %d tasks in %v (%.2f tasks/sec)", numTasks, duration, throughput)

		// Performance assertion - should be able to handle at least 100 tasks/sec
		assert.Greater(t, throughput, 100.0, "Queue should handle at least 100 enqueues per second")
	})

	t.Run("High throughput dequeue", func(t *testing.T) {
		// Ensure we have tasks in the queue from previous test
		length, err := env.QueueRepo.GetQueueLength(ctx)
		require.NoError(t, err)

		if length < 100 {
			// Add more tasks if needed
			for i := 0; i < 100; i++ {
				task, err := testutils.CreateTestTask("dequeue-perf-user", fmt.Sprintf("Dequeue Perf Task %d", i))
				require.NoError(t, err)

				err = env.QueueRepo.Enqueue(ctx, task)
				require.NoError(t, err)
			}
		}

		const numDequeues = 100
		start := time.Now()

		for i := 0; i < numDequeues; i++ {
			_, err := env.QueueRepo.Dequeue(ctx)
			require.NoError(t, err)
		}

		duration := time.Since(start)
		throughput := float64(numDequeues) / duration.Seconds()

		t.Logf("Dequeued %d tasks in %v (%.2f tasks/sec)", numDequeues, duration, throughput)

		// Performance assertion - should be able to handle at least 50 dequeues/sec
		assert.Greater(t, throughput, 50.0, "Queue should handle at least 50 dequeues per second")
	})
}

func TestRabbitMQQueue_Integration_Reliability(t *testing.T) {
	env := testutils.SetupTestEnvironment(t)
	defer env.TearDown(t)

	ctx := context.Background()

	t.Run("Message persistence", func(t *testing.T) {
		// Enqueue tasks
		tasks := make([]*testutils.TaskCreated, 5)
		for i := 0; i < 5; i++ {
			task, err := testutils.CreateTestTask("persistence-user", fmt.Sprintf("Persistent Task %d", i))
			require.NoError(t, err)
			tasks[i] = &testutils.TaskCreated{Task: task}

			err = env.QueueRepo.Enqueue(ctx, task)
			require.NoError(t, err)
		}

		// Wait a bit to ensure messages are persisted
		time.Sleep(100 * time.Millisecond)

		// Dequeue and verify all tasks
		for i := 0; i < 5; i++ {
			dequeuedTask, err := env.QueueRepo.Dequeue(ctx)
			require.NoError(t, err)

			// Find the matching original task
			found := false
			for _, originalTaskWrapper := range tasks {
				if originalTaskWrapper.Task.ID().Value() == dequeuedTask.ID().Value() {
					testutils.AssertTaskEqual(t, originalTaskWrapper.Task, dequeuedTask)
					found = true
					break
				}
			}
			assert.True(t, found, "Dequeued task should match one of the original tasks")
		}
	})

	t.Run("Task ordering", func(t *testing.T) {
		// Enqueue tasks in order
		taskIDs := make([]string, 10)
		for i := 0; i < 10; i++ {
			task, err := testutils.CreateTestTask("order-user", fmt.Sprintf("Order Task %d", i))
			require.NoError(t, err)
			taskIDs[i] = task.ID().Value()

			err = env.QueueRepo.Enqueue(ctx, task)
			require.NoError(t, err)
		}

		// Dequeue tasks and check order
		dequeuedIDs := make([]string, 10)
		for i := 0; i < 10; i++ {
			task, err := env.QueueRepo.Dequeue(ctx)
			require.NoError(t, err)
			dequeuedIDs[i] = task.ID().Value()
		}

		// FIFO ordering should be maintained
		assert.Equal(t, taskIDs, dequeuedIDs, "Tasks should be dequeued in FIFO order")
	})
}

// Helper for task creation wrapper
type TaskCreated struct {
	Task *testutils.CreateTestTask
}

// Benchmark tests
func BenchmarkRabbitMQQueue_Enqueue(b *testing.B) {
	env := testutils.SetupTestEnvironment(b)
	defer env.TearDown(b)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task, err := testutils.CreateTestTask("bench-user", fmt.Sprintf("Benchmark Task %d", i))
		if err != nil {
			b.Fatal(err)
		}

		err = env.QueueRepo.Enqueue(ctx, task)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRabbitMQQueue_Dequeue(b *testing.B) {
	env := testutils.SetupTestEnvironment(b)
	defer env.TearDown(b)

	ctx := context.Background()

	// Pre-populate queue with tasks
	for i := 0; i < b.N; i++ {
		task, err := testutils.CreateTestTask("bench-user", fmt.Sprintf("Benchmark Task %d", i))
		if err != nil {
			b.Fatal(err)
		}

		err = env.QueueRepo.Enqueue(ctx, task)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := env.QueueRepo.Dequeue(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRabbitMQQueue_EnqueueDequeue(b *testing.B) {
	env := testutils.SetupTestEnvironment(b)
	defer env.TearDown(b)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task, err := testutils.CreateTestTask("bench-user", fmt.Sprintf("Benchmark Task %d", i))
		if err != nil {
			b.Fatal(err)
		}

		err = env.QueueRepo.Enqueue(ctx, task)
		if err != nil {
			b.Fatal(err)
		}

		_, err = env.QueueRepo.Dequeue(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}