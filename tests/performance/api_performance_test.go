package performance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"local-backend-server/internal/handlers"
	"local-backend-server/internal/infrastructure/database/models"
	"local-backend-server/internal/services"
)

// PerformanceTestSuite for API performance testing
type PerformanceTestSuite struct {
	db     *gorm.DB
	router *gin.Engine
}

// setupPerformanceTestSuite creates environment for performance testing
func setupPerformanceTestSuite(t *testing.T) *PerformanceTestSuite {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate models
	err = db.AutoMigrate(&models.WorkflowHistory{})
	require.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add request ID middleware
	router.Use(func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	})

	// Setup routes
	taskHistoryHandler := handlers.NewTaskHistoryHandler(db)
	api := router.Group("/api/v1")
	tasks := api.Group("/tasks")
	tasks.GET("/history/:repository_name", taskHistoryHandler.GetTaskHistory)

	return &PerformanceTestSuite{
		db:     db,
		router: router,
	}
}

// seedPerformanceData creates realistic test dataset
func (suite *PerformanceTestSuite) seedPerformanceData(t *testing.T, repoName string, count int) {
	baseTime := time.Now()
	batch := make([]models.WorkflowHistory, count)

	for i := 0; i < count; i++ {
		status := models.WorkflowStatusCompleted
		if i%10 == 0 {
			status = models.WorkflowStatusProcessing
		} else if i%15 == 0 {
			status = models.WorkflowStatusFailed
		}

		batch[i] = models.WorkflowHistory{
			RequestID:       fmt.Sprintf("perf-req-%s-%d", repoName, i),
			Status:          status,
			Tasks:           fmt.Sprintf("Performance test task %d for %s", i, repoName),
			RepositoryName:  repoName,
			WorkingDir:      stringPtr(fmt.Sprintf("/path/to/%s/task-%d", repoName, i)),
			ClaudeCmd:       stringPtr("claude code --performance-test"),
			Interactive:     i%2 == 0,
			ContinueTask:    i%3 == 0,
			CreatedAt:       baseTime.Add(-time.Duration(i) * time.Minute),
			CompletedAt:     timePtr(baseTime.Add(-time.Duration(i-1) * time.Minute)),
			ProcessingTimeMs: int64Ptr(int64(30000 + i*1000)), // Varied processing times
			Result:          stringPtr(fmt.Sprintf("Performance test result %d", i)),
		}

		// Add errors for failed tasks
		if status == models.WorkflowStatusFailed {
			batch[i].Error = stringPtr(fmt.Sprintf("Performance test error %d", i))
			batch[i].Result = nil
		}
	}

	// Batch insert for better performance
	err := suite.db.CreateInBatches(batch, 100).Error
	require.NoError(t, err)
}

// TestAPI_ResponseTimeRequirements validates <200ms requirement
func TestAPI_ResponseTimeRequirements(t *testing.T) {
	suite := setupPerformanceTestSuite(t)

	testCases := []struct {
		name      string
		dataSize  int
		repoName  string
		queryPath string
		maxTime   time.Duration
	}{
		{
			name:      "small dataset (50 records)",
			dataSize:  50,
			repoName:  "small-repo",
			queryPath: "/api/v1/tasks/history/small-repo?limit=20",
			maxTime:   200 * time.Millisecond,
		},
		{
			name:      "medium dataset (200 records)",
			dataSize:  200,
			repoName:  "medium-repo",
			queryPath: "/api/v1/tasks/history/medium-repo?limit=20",
			maxTime:   200 * time.Millisecond,
		},
		{
			name:      "large dataset (1000 records)",
			dataSize:  1000,
			repoName:  "large-repo",
			queryPath: "/api/v1/tasks/history/large-repo?limit=20",
			maxTime:   200 * time.Millisecond,
		},
		{
			name:      "pagination test (page 5)",
			dataSize:  200,
			repoName:  "pagination-repo",
			queryPath: "/api/v1/tasks/history/pagination-repo?page=5&limit=10",
			maxTime:   200 * time.Millisecond,
		},
		{
			name:      "large page size (50 items)",
			dataSize:  200,
			repoName:  "large-page-repo",
			queryPath: "/api/v1/tasks/history/large-page-repo?limit=50",
			maxTime:   200 * time.Millisecond,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Seed test data
			suite.seedPerformanceData(t, tc.repoName, tc.dataSize)

			// Perform multiple requests to get average performance
			const numRequests = 10
			var totalTime time.Duration
			var minTime = time.Hour
			var maxTime time.Duration

			for i := 0; i < numRequests; i++ {
				req, err := http.NewRequest("GET", tc.queryPath, nil)
				require.NoError(t, err)

				start := time.Now()
				recorder := httptest.NewRecorder()
				suite.router.ServeHTTP(recorder, req)
				elapsed := time.Since(start)

				assert.Equal(t, http.StatusOK, recorder.Code)

				totalTime += elapsed
				if elapsed < minTime {
					minTime = elapsed
				}
				if elapsed > maxTime {
					maxTime = elapsed
				}

				// Verify response structure
				var response services.TaskHistoryResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Greater(t, len(response.Data), 0)
			}

			avgTime := totalTime / numRequests

			// Performance assertions
			assert.Less(t, avgTime, tc.maxTime,
				"Average response time (%v) should be under %v", avgTime, tc.maxTime)
			assert.Less(t, maxTime, tc.maxTime*2,
				"Maximum response time (%v) should be under %v", maxTime, tc.maxTime*2)

			t.Logf("Performance metrics for %s:", tc.name)
			t.Logf("  Average: %v", avgTime)
			t.Logf("  Min: %v", minTime)
			t.Logf("  Max: %v", maxTime)
			t.Logf("  Data size: %d records", tc.dataSize)
		})
	}
}

// TestAPI_ConcurrentRequestPerformance tests performance under concurrent load
func TestAPI_ConcurrentRequestPerformance(t *testing.T) {
	suite := setupPerformanceTestSuite(t)

	// Seed test data
	suite.seedPerformanceData(t, "concurrent-repo", 500)

	concurrencyLevels := []int{1, 5, 10, 20}

	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("concurrency_%d", concurrency), func(t *testing.T) {
			var wg sync.WaitGroup
			var mu sync.Mutex
			responseTimes := make([]time.Duration, 0, concurrency)
			errors := make([]error, 0)

			start := time.Now()

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					req, err := http.NewRequest("GET", "/api/v1/tasks/history/concurrent-repo?limit=20", nil)
					if err != nil {
						mu.Lock()
						errors = append(errors, err)
						mu.Unlock()
						return
					}

					reqStart := time.Now()
					recorder := httptest.NewRecorder()
					suite.router.ServeHTTP(recorder, req)
					elapsed := time.Since(reqStart)

					mu.Lock()
					responseTimes = append(responseTimes, elapsed)
					mu.Unlock()

					if recorder.Code != http.StatusOK {
						mu.Lock()
						errors = append(errors, fmt.Errorf("HTTP %d", recorder.Code))
						mu.Unlock()
						return
					}

					// Verify response
					var response services.TaskHistoryResponse
					err = json.Unmarshal(recorder.Body.Bytes(), &response)
					if err != nil {
						mu.Lock()
						errors = append(errors, err)
						mu.Unlock()
					}
				}()
			}

			wg.Wait()
			totalTestTime := time.Since(start)

			// Verify no errors
			assert.Empty(t, errors, "No errors should occur during concurrent requests")

			// Verify all requests completed
			assert.Len(t, responseTimes, concurrency)

			// Calculate performance metrics
			var totalTime time.Duration
			var minTime = time.Hour
			var maxTime time.Duration

			for _, responseTime := range responseTimes {
				totalTime += responseTime
				if responseTime < minTime {
					minTime = responseTime
				}
				if responseTime > maxTime {
					maxTime = responseTime
				}
			}

			avgTime := totalTime / time.Duration(len(responseTimes))
			throughput := float64(concurrency) / totalTestTime.Seconds()

			// Performance assertions
			assert.Less(t, avgTime, 200*time.Millisecond,
				"Average response time should be under 200ms even with concurrency")
			assert.Less(t, maxTime, 500*time.Millisecond,
				"Max response time should be reasonable under load")

			t.Logf("Concurrency %d performance:", concurrency)
			t.Logf("  Total test time: %v", totalTestTime)
			t.Logf("  Average response time: %v", avgTime)
			t.Logf("  Min response time: %v", minTime)
			t.Logf("  Max response time: %v", maxTime)
			t.Logf("  Throughput: %.2f req/sec", throughput)
		})
	}
}

// TestAPI_MemoryUsagePerformance tests memory efficiency
func TestAPI_MemoryUsagePerformance(t *testing.T) {
	suite := setupPerformanceTestSuite(t)

	// Seed large dataset
	suite.seedPerformanceData(t, "memory-repo", 2000)

	// Test different page sizes to verify memory usage is bounded
	pageSizes := []int{10, 20, 50, 100}

	for _, pageSize := range pageSizes {
		t.Run(fmt.Sprintf("page_size_%d", pageSize), func(t *testing.T) {
			path := fmt.Sprintf("/api/v1/tasks/history/memory-repo?limit=%d", pageSize)

			req, err := http.NewRequest("GET", path, nil)
			require.NoError(t, err)

			start := time.Now()
			recorder := httptest.NewRecorder()
			suite.router.ServeHTTP(recorder, req)
			elapsed := time.Since(start)

			assert.Equal(t, http.StatusOK, recorder.Code)

			var response services.TaskHistoryResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)

			// Verify correct page size
			assert.Len(t, response.Data, pageSize)

			// Verify response time scales linearly with page size (not exponentially)
			expectedMaxTime := time.Duration(pageSize) * 2 * time.Millisecond // 2ms per record max
			assert.Less(t, elapsed, expectedMaxTime,
				"Response time should scale linearly with page size")

			t.Logf("Page size %d: %v (%.2fms per record)",
				pageSize, elapsed, float64(elapsed.Milliseconds())/float64(pageSize))
		})
	}
}

// TestAPI_DatabaseIndexPerformance validates index effectiveness
func TestAPI_DatabaseIndexPerformance(t *testing.T) {
	suite := setupPerformanceTestSuite(t)

	// Create multiple repositories with data
	repositories := []string{"repo1", "repo2", "repo3", "repo4", "repo5"}

	for _, repo := range repositories {
		suite.seedPerformanceData(t, repo, 200)
	}

	// Test filtering by repository (should use index)
	for _, repo := range repositories {
		t.Run(fmt.Sprintf("repository_%s", repo), func(t *testing.T) {
			req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/tasks/history/%s?limit=20", repo), nil)
			require.NoError(t, err)

			start := time.Now()
			recorder := httptest.NewRecorder()
			suite.router.ServeHTTP(recorder, req)
			elapsed := time.Since(start)

			assert.Equal(t, http.StatusOK, recorder.Code)
			assert.Less(t, elapsed, 200*time.Millisecond,
				"Repository filtering should be fast with proper indexing")

			var response services.TaskHistoryResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)

			// Verify all results are for the correct repository
			for _, task := range response.Data {
				assert.Equal(t, repo, task.RepositoryName)
			}

			// Verify correct total count
			assert.Equal(t, 200, response.Pagination.Total)
		})
	}
}

// TestAPI_StressTest performs stress testing with high load
func TestAPI_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	suite := setupPerformanceTestSuite(t)

	// Seed large dataset
	suite.seedPerformanceData(t, "stress-repo", 5000)

	t.Run("sustained_load_test", func(t *testing.T) {
		const (
			testDuration = 10 * time.Second
			concurrency  = 20
		)

		var wg sync.WaitGroup
		var mu sync.Mutex
		var requestCount int
		var errorCount int
		var totalResponseTime time.Duration

		stopChan := make(chan struct{})

		// Start background goroutines
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for {
					select {
					case <-stopChan:
						return
					default:
						req, err := http.NewRequest("GET", "/api/v1/tasks/history/stress-repo?limit=20", nil)
						if err != nil {
							mu.Lock()
							errorCount++
							mu.Unlock()
							continue
						}

						start := time.Now()
						recorder := httptest.NewRecorder()
						suite.router.ServeHTTP(recorder, req)
						elapsed := time.Since(start)

						mu.Lock()
						requestCount++
						totalResponseTime += elapsed
						if recorder.Code != http.StatusOK {
							errorCount++
						}
						mu.Unlock()
					}
				}
			}()
		}

		// Run for specified duration
		time.Sleep(testDuration)
		close(stopChan)
		wg.Wait()

		// Calculate metrics
		avgResponseTime := totalResponseTime / time.Duration(requestCount)
		throughput := float64(requestCount) / testDuration.Seconds()
		errorRate := float64(errorCount) / float64(requestCount) * 100

		// Performance assertions
		assert.Less(t, avgResponseTime, 200*time.Millisecond,
			"Average response time should remain under 200ms under sustained load")
		assert.Less(t, errorRate, 1.0,
			"Error rate should be under 1% during stress test")
		assert.Greater(t, throughput, 50.0,
			"Should maintain at least 50 requests/sec throughput")

		t.Logf("Stress test results (duration: %v, concurrency: %d):", testDuration, concurrency)
		t.Logf("  Total requests: %d", requestCount)
		t.Logf("  Average response time: %v", avgResponseTime)
		t.Logf("  Throughput: %.2f req/sec", throughput)
		t.Logf("  Error rate: %.2f%%", errorRate)
	})
}

// Helper functions
func stringPtr(s string) *string    { return &s }
func timePtr(t time.Time) *time.Time { return &t }
func int64Ptr(i int64) *int64       { return &i }