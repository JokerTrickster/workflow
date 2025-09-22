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
	"gorm.io/gorm/logger"

	"local-backend-server/internal/handlers"
	"local-backend-server/internal/infrastructure/database/migrations"
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
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger:               logger.Default.LogMode(logger.Silent), // Silent logging
		PrepareStmt:          true,
		SkipDefaultTransaction: true, // Disable default transactions
	})
	require.NoError(t, err)

	// Run proper migrations with indexes
	migrator := migrations.NewMigrator(db)
	err = migrator.RunMigrations()
	require.NoError(t, err)

	// Optimize SQLite for performance testing
	optimizeDatabase(t, db)

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

// optimizeDatabase applies performance optimizations for testing
func optimizeDatabase(t *testing.T, db *gorm.DB) {
	// SQLite performance optimizations
	queries := []string{
		"PRAGMA journal_mode=MEMORY",   // Use memory journal for speed
		"PRAGMA synchronous=OFF",       // Fastest writes (unsafe but OK for tests)
		"PRAGMA cache_size=10000",      // Larger cache
		"PRAGMA temp_store=MEMORY",     // Use memory for temp storage
		"PRAGMA mmap_size=134217728",   // 128MB memory mapping
	}

	for _, query := range queries {
		err := db.Exec(query).Error
		require.NoError(t, err, "Failed to execute optimization query: %s", query)
	}
}

// seedPerformanceData creates realistic test dataset
func (suite *PerformanceTestSuite) seedPerformanceData(t *testing.T, repoName string, count int) {
	baseTime := time.Now()

	// Insert records one by one to avoid transaction issues
	for i := 0; i < count; i++ {
		status := models.WorkflowStatusCompleted
		if i%10 == 0 {
			status = models.WorkflowStatusProcessing
		} else if i%15 == 0 {
			status = models.WorkflowStatusFailed
		}

		record := models.WorkflowHistory{
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
			record.Error = stringPtr(fmt.Sprintf("Performance test error %d", i))
			record.Result = nil
		}

		err := suite.db.Create(&record).Error
		require.NoError(t, err)
	}
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

			// Warm up the database cache
			req, _ := http.NewRequest("GET", tc.queryPath, nil)
			recorder := httptest.NewRecorder()
			suite.router.ServeHTTP(recorder, req)

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

// TestAPI_QueryPlanAnalysis performs database query analysis
func TestAPI_QueryPlanAnalysis(t *testing.T) {
	suite := setupPerformanceTestSuite(t)

	// Seed test data
	suite.seedPerformanceData(t, "analysis-repo", 1000)

	// Test query execution plan for repository filtering
	t.Run("repository_filter_query_plan", func(t *testing.T) {
		var results []map[string]interface{}

		// EXPLAIN QUERY PLAN for the main query
		query := `EXPLAIN QUERY PLAN
		         SELECT count(*) FROM workflow_histories
		         WHERE repository_name = ?`

		err := suite.db.Raw(query, "analysis-repo").Scan(&results).Error
		require.NoError(t, err)

		t.Logf("Count Query Plan:")
		for _, result := range results {
			t.Logf("  %v", result)
		}

		// Test pagination query plan
		results = []map[string]interface{}{}
		query = `EXPLAIN QUERY PLAN
		         SELECT * FROM workflow_histories
		         WHERE repository_name = ?
		         ORDER BY created_at DESC
		         LIMIT ? OFFSET ?`

		err = suite.db.Raw(query, "analysis-repo", 20, 0).Scan(&results).Error
		require.NoError(t, err)

		t.Logf("Pagination Query Plan:")
		for _, result := range results {
			t.Logf("  %v", result)
		}
	})
}

// Helper functions
func stringPtr(s string) *string    { return &s }
func timePtr(t time.Time) *time.Time { return &t }
func int64Ptr(i int64) *int64       { return &i }