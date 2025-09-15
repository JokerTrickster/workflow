package performance_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ai-git-workbench/internal/domain/valueobjects"
	"ai-git-workbench/tests/testutils"
)

func TestPerformance_ConcurrentTaskOperations(t *testing.T) {
	env := testutils.SetupTestEnvironment(t)
	defer env.TearDown(t)

	ctx := context.Background()
	env.CleanDatabase(t)

	config := testutils.DefaultBenchmarkConfig()

	t.Run("Concurrent task creation", func(t *testing.T) {
		var wg sync.WaitGroup
		var mu sync.Mutex
		errors := make([]error, 0)
		taskIDs := make([]string, 0)

		start := time.Now()

		for user := 0; user < config.ConcurrentUsers; user++ {
			wg.Add(1)
			go func(userID int) {
				defer wg.Done()
				
				userStr := fmt.Sprintf("load-user-%d", userID)
				for task := 0; task < config.TasksPerUser; task++ {
					// Create task
					testTask, err := testutils.CreateTestTask(
						userStr,
						fmt.Sprintf("Load Test Task %d-%d", userID, task),
					)
					
					if err != nil {
						mu.Lock()
						errors = append(errors, err)
						mu.Unlock()
						continue
					}

					// Store in repository
					err = env.TaskRepo.Create(ctx, testTask)
					if err != nil {
						mu.Lock()
						errors = append(errors, err)
						mu.Unlock()
						continue
					}

					mu.Lock()
					taskIDs = append(taskIDs, testTask.ID().Value())
					mu.Unlock()
				}
			}(user)
		}

		wg.Wait()
		duration := time.Since(start)

		// Verify no errors occurred
		require.Empty(t, errors, "No errors should occur during concurrent creation")

		// Verify all tasks were created
		expectedTasks := config.ConcurrentUsers * config.TasksPerUser
		assert.Len(t, taskIDs, expectedTasks)

		// Performance metrics
		totalOps := len(taskIDs)
		throughput := float64(totalOps) / duration.Seconds()

		t.Logf("Created %d tasks in %v", totalOps, duration)
		t.Logf("Throughput: %.2f tasks/second", throughput)
		t.Logf("Average response time: %v", duration/time.Duration(totalOps))

		// Performance assertions
		assert.Greater(t, throughput, float64(config.ExpectedThroughput), 
			fmt.Sprintf("Should achieve at least %d ops/sec", config.ExpectedThroughput))
		assert.Less(t, duration/time.Duration(totalOps), config.MaxResponseTime,
			fmt.Sprintf("Average response time should be less than %v", config.MaxResponseTime))
	})

	t.Run("Concurrent task queries", func(t *testing.T) {
		// First ensure we have tasks in the database
		userID, err := valueobjects.NewUserID("query-user")
		require.NoError(t, err)

		// Create test data
		const numTestTasks = 100
		for i := 0; i < numTestTasks; i++ {
			task, err := testutils.CreateTestTask("query-user", fmt.Sprintf("Query Test Task %d", i))
			require.NoError(t, err)

			err = env.TaskRepo.Create(ctx, task)
			require.NoError(t, err)
		}

		var wg sync.WaitGroup
		var mu sync.Mutex
		errors := make([]error, 0)
		totalQueries := 0

		start := time.Now()

		// Run concurrent queries
		for worker := 0; worker < config.ConcurrentUsers; worker++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				
				for i := 0; i < config.TasksPerUser; i++ {
					// Query tasks for user
					_, err := env.TaskRepo.GetByUserID(ctx, userID, 10, 0)
					if err != nil {
						mu.Lock()
						errors = append(errors, err)
						mu.Unlock()
						continue
					}

					mu.Lock()
					totalQueries++
					mu.Unlock()
				}
			}()
		}

		wg.Wait()
		duration := time.Since(start)

		// Verify no errors occurred
		require.Empty(t, errors, "No errors should occur during concurrent queries")

		// Performance metrics
		throughput := float64(totalQueries) / duration.Seconds()

		t.Logf("Executed %d queries in %v", totalQueries, duration)
		t.Logf("Query throughput: %.2f queries/second", throughput)
		t.Logf("Average query time: %v", duration/time.Duration(totalQueries))

		// Performance assertions
		assert.Greater(t, throughput, float64(config.ExpectedThroughput*2), 
			"Query throughput should be higher than creation throughput")
		assert.Less(t, duration/time.Duration(totalQueries), config.MaxResponseTime/2,
			"Query response time should be faster than creation")
	})

	t.Run("Mixed workload performance", func(t *testing.T) {
		var wg sync.WaitGroup
		var mu sync.Mutex
		errors := make([]error, 0)
		operations := make(map[string]int)

		start := time.Now()

		// Run mixed workload: 60% reads, 30% creates, 10% updates
		for worker := 0; worker < config.ConcurrentUsers; worker++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				
				userStr := fmt.Sprintf("mixed-user-%d", workerID)
				userID, err := valueobjects.NewUserID(userStr)
				if err != nil {
					mu.Lock()
					errors = append(errors, err)
					mu.Unlock()
					return
				}

				for op := 0; op < config.TasksPerUser; op++ {
					switch op % 10 {
					case 0, 1, 2: // 30% creates
						task, err := testutils.CreateTestTask(userStr, fmt.Sprintf("Mixed Task %d-%d", workerID, op))
						if err != nil {
							mu.Lock()
							errors = append(errors, err)
							mu.Unlock()
							continue
						}

						err = env.TaskRepo.Create(ctx, task)
						if err != nil {
							mu.Lock()
							errors = append(errors, err)
							mu.Unlock()
							continue
						}

						mu.Lock()
						operations["create"]++
						mu.Unlock()

					case 3: // 10% updates
						// Get existing tasks and update one
						tasks, err := env.TaskRepo.GetByUserID(ctx, userID, 1, 0)
						if err == nil && len(tasks) > 0 {
							err = tasks[0].SetMetadata("updated", "true")
							if err == nil {
								err = env.TaskRepo.Update(ctx, tasks[0])
							}
						}

						if err != nil {
							mu.Lock()
							errors = append(errors, err)
							mu.Unlock()
							continue
						}

						mu.Lock()
						operations["update"]++
						mu.Unlock()

					default: // 60% reads
						_, err := env.TaskRepo.GetByUserID(ctx, userID, 5, 0)
						if err != nil {
							mu.Lock()
							errors = append(errors, err)
							mu.Unlock()
							continue
						}

						mu.Lock()
						operations["read"]++
						mu.Unlock()
					}
				}
			}(worker)
		}

		wg.Wait()
		duration := time.Since(start)

		// Calculate total operations
		totalOps := 0
		for _, count := range operations {
			totalOps += count
		}

		// Verify reasonable error rate (< 5%)
		errorRate := float64(len(errors)) / float64(totalOps+len(errors))
		assert.Less(t, errorRate, 0.05, "Error rate should be less than 5%")

		// Performance metrics
		throughput := float64(totalOps) / duration.Seconds()

		t.Logf("Mixed workload results:")
		t.Logf("  Total operations: %d", totalOps)
		t.Logf("  Creates: %d, Reads: %d, Updates: %d", operations["create"], operations["read"], operations["update"])
		t.Logf("  Duration: %v", duration)
		t.Logf("  Throughput: %.2f ops/second", throughput)
		t.Logf("  Error rate: %.2f%%", errorRate*100)

		// Performance assertions
		assert.Greater(t, throughput, float64(config.ExpectedThroughput/2), 
			"Mixed workload should achieve reasonable throughput")
	})
}

func TestPerformance_QueueOperations(t *testing.T) {
	env := testutils.SetupTestEnvironment(t)
	defer env.TearDown(t)

	ctx := context.Background()
	config := testutils.DefaultBenchmarkConfig()

	t.Run("Queue throughput test", func(t *testing.T) {
		var wg sync.WaitGroup
		var mu sync.Mutex
		errors := make([]error, 0)
		enqueuedCount := 0

		start := time.Now()

		// Enqueue phase
		for worker := 0; worker < config.ConcurrentUsers; worker++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				
				for i := 0; i < config.TasksPerUser; i++ {
					task, err := testutils.CreateTestTask(
						fmt.Sprintf("queue-user-%d", workerID),
						fmt.Sprintf("Queue Task %d-%d", workerID, i),
					)
					if err != nil {
						mu.Lock()
						errors = append(errors, err)
						mu.Unlock()
						continue
					}

					err = env.QueueRepo.Enqueue(ctx, task)
					if err != nil {
						mu.Lock()
						errors = append(errors, err)
						mu.Unlock()
						continue
					}

					mu.Lock()
					enqueuedCount++
					mu.Unlock()
				}
			}(worker)
		}

		wg.Wait()
		enqueueTime := time.Since(start)

		// Verify enqueue operations
		require.Empty(t, errors, "No errors should occur during enqueue")
		
		enqueueThroughput := float64(enqueuedCount) / enqueueTime.Seconds()
		t.Logf("Enqueued %d tasks in %v (%.2f tasks/sec)", enqueuedCount, enqueueTime, enqueueThroughput)

		// Dequeue phase
		start = time.Now()
		dequeuedCount := 0
		errors = make([]error, 0)

		// Use fewer workers for dequeue to simulate processing
		for worker := 0; worker < config.ConcurrentUsers/2; worker++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				
				for {
					_, err := env.QueueRepo.Dequeue(ctx)
					if err != nil {
						// Queue might be empty
						return
					}

					mu.Lock()
					dequeuedCount++
					mu.Unlock()

					// Simulate processing time
					time.Sleep(time.Millisecond)
				}
			}()
		}

		// Let dequeue run for a limited time
		time.Sleep(5 * time.Second)
		dequeueTime := time.Since(start)

		dequeueThroughput := float64(dequeuedCount) / dequeueTime.Seconds()
		t.Logf("Dequeued %d tasks in %v (%.2f tasks/sec)", dequeuedCount, dequeueTime, dequeueThroughput)

		// Performance assertions
		assert.Greater(t, enqueueThroughput, float64(config.ExpectedThroughput), 
			"Enqueue throughput should meet requirements")
		assert.Greater(t, dequeueThroughput, float64(config.ExpectedThroughput/2), 
			"Dequeue throughput should be reasonable with processing delay")
	})
}

func TestPerformance_MemoryUsage(t *testing.T) {
	env := testutils.SetupTestEnvironment(t)
	defer env.TearDown(t)

	ctx := context.Background()
	env.CleanDatabase(t)

	t.Run("Memory efficiency with large datasets", func(t *testing.T) {
		const numTasks = 1000
		
		// Create many tasks
		start := time.Now()
		for i := 0; i < numTasks; i++ {
			task, err := testutils.CreateTestTask("memory-user", fmt.Sprintf("Memory Test Task %d", i))
			require.NoError(t, err)

			// Add some metadata to increase memory usage
			err = task.SetMetadata("iteration", fmt.Sprintf("%d", i))
			require.NoError(t, err)
			err = task.SetMetadata("description", fmt.Sprintf("This is task number %d with some metadata", i))
			require.NoError(t, err)

			err = env.TaskRepo.Create(ctx, task)
			require.NoError(t, err)
		}

		createTime := time.Since(start)
		t.Logf("Created %d tasks in %v", numTasks, createTime)

		// Query all tasks
		userID, err := valueobjects.NewUserID("memory-user")
		require.NoError(t, err)

		start = time.Now()
		tasks, err := env.TaskRepo.GetByUserID(ctx, userID, numTasks, 0)
		require.NoError(t, err)
		queryTime := time.Since(start)

		assert.Len(t, tasks, numTasks)
		t.Logf("Queried %d tasks in %v", len(tasks), queryTime)

		// Performance assertions
		assert.Less(t, createTime, 30*time.Second, "Should create 1000 tasks within 30 seconds")
		assert.Less(t, queryTime, 5*time.Second, "Should query 1000 tasks within 5 seconds")
	})
}

func TestPerformance_StressTest(t *testing.T) {
	env := testutils.SetupTestEnvironment(t)
	defer env.TearDown(t)

	ctx := context.Background()
	env.CleanDatabase(t)

	// Use more aggressive configuration for stress test
	stressConfig := &testutils.BenchmarkConfig{
		ConcurrentUsers:    50,
		TasksPerUser:      20,
		MaxResponseTime:   500 * time.Millisecond,
		ExpectedThroughput: 50,
	}

	t.Run("High load stress test", func(t *testing.T) {
		var wg sync.WaitGroup
		var mu sync.Mutex
		errors := make([]error, 0)
		completedOps := 0

		start := time.Now()

		for worker := 0; worker < stressConfig.ConcurrentUsers; worker++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				
				userStr := fmt.Sprintf("stress-user-%d", workerID)
				
				for op := 0; op < stressConfig.TasksPerUser; op++ {
					// Create task
					task, err := testutils.CreateTestTask(userStr, fmt.Sprintf("Stress Task %d-%d", workerID, op))
					if err != nil {
						mu.Lock()
						errors = append(errors, err)
						mu.Unlock()
						continue
					}

					// Store in repository
					err = env.TaskRepo.Create(ctx, task)
					if err != nil {
						mu.Lock()
						errors = append(errors, err)
						mu.Unlock()
						continue
					}

					// Enqueue task
					err = env.QueueRepo.Enqueue(ctx, task)
					if err != nil {
						mu.Lock()
						errors = append(errors, err)
						mu.Unlock()
						continue
					}

					mu.Lock()
					completedOps++
					mu.Unlock()
				}
			}(worker)
		}

		wg.Wait()
		duration := time.Since(start)

		// Calculate metrics
		totalExpectedOps := stressConfig.ConcurrentUsers * stressConfig.TasksPerUser
		successRate := float64(completedOps) / float64(totalExpectedOps)
		throughput := float64(completedOps) / duration.Seconds()

		t.Logf("Stress test results:")
		t.Logf("  Expected operations: %d", totalExpectedOps)
		t.Logf("  Completed operations: %d", completedOps)
		t.Logf("  Success rate: %.2f%%", successRate*100)
		t.Logf("  Errors: %d", len(errors))
		t.Logf("  Duration: %v", duration)
		t.Logf("  Throughput: %.2f ops/second", throughput)

		// Stress test assertions (more lenient than normal tests)
		assert.Greater(t, successRate, 0.9, "Should maintain >90% success rate under stress")
		assert.Greater(t, throughput, float64(stressConfig.ExpectedThroughput), 
			"Should maintain minimum throughput under stress")
		assert.Less(t, float64(len(errors))/float64(totalExpectedOps), 0.1, 
			"Error rate should be less than 10% under stress")
	})
}

// Benchmark functions for go test -bench
func BenchmarkPerformance_TaskCreation(b *testing.B) {
	env := testutils.SetupTestEnvironment(b)
	defer env.TearDown(b)

	ctx := context.Background()
	env.CleanDatabase(b)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			task, err := testutils.CreateTestTask("bench-user", fmt.Sprintf("Benchmark Task %d", i))
			if err != nil {
				b.Fatal(err)
			}

			err = env.TaskRepo.Create(ctx, task)
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

func BenchmarkPerformance_TaskQuery(b *testing.B) {
	env := testutils.SetupTestEnvironment(b)
	defer env.TearDown(b)

	ctx := context.Background()
	env.CleanDatabase(b)

	// Setup test data
	userID, err := valueobjects.NewUserID("bench-user")
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < 100; i++ {
		task, err := testutils.CreateTestTask("bench-user", fmt.Sprintf("Query Benchmark Task %d", i))
		if err != nil {
			b.Fatal(err)
		}

		err = env.TaskRepo.Create(ctx, task)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := env.TaskRepo.GetByUserID(ctx, userID, 10, 0)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkPerformance_QueueOperations(b *testing.B) {
	env := testutils.SetupTestEnvironment(b)
	defer env.TearDown(b)

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			task, err := testutils.CreateTestTask("bench-user", fmt.Sprintf("Queue Benchmark Task %d", i))
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
			i++
		}
	})
}