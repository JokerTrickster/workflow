package performance

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"local-backend-server/internal/infrastructure/database/migrations"
	"local-backend-server/internal/infrastructure/database/models"
	"local-backend-server/internal/services"
)

// OptimizedPerformanceTestSuite for testing optimized components
type OptimizedPerformanceTestSuite struct {
	db              *gorm.DB
	optimizedService *services.OptimizedTaskHistoryService
}

// setupOptimizedTestSuite creates environment for optimized performance testing
func setupOptimizedTestSuite(t *testing.T) *OptimizedPerformanceTestSuite {
	// Setup in-memory database with optimizations
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		PrepareStmt:            true,
		SkipDefaultTransaction: true, // Disable for better performance in tests
	})
	require.NoError(t, err)

	// Run migrations
	migrator := migrations.NewMigrator(db)
	err = migrator.RunMigrations()
	require.NoError(t, err)

	// Apply test optimizations
	optimizations := []string{
		"PRAGMA journal_mode=MEMORY",   // Fastest for testing
		"PRAGMA synchronous=OFF",       // Fastest for testing (unsafe but OK for tests)
		"PRAGMA cache_size=10000",
		"PRAGMA temp_store=MEMORY",
		"PRAGMA mmap_size=134217728",
		"PRAGMA optimize",
	}

	for _, opt := range optimizations {
		err := db.Exec(opt).Error
		require.NoError(t, err)
	}

	// Create optimized service
	optimizedService := services.NewOptimizedTaskHistoryService(db)

	return &OptimizedPerformanceTestSuite{
		db:              db,
		optimizedService: optimizedService,
	}
}

// seedOptimizedData creates test data for optimized testing with simple insertion
func (suite *OptimizedPerformanceTestSuite) seedOptimizedData(t *testing.T, repoName string, count int) {
	baseTime := time.Now()

	// Simple batch creation without complex transactions
	var records []models.WorkflowHistory
	for i := 0; i < count; i++ {
		status := models.WorkflowStatusCompleted
		if i%10 == 0 {
			status = models.WorkflowStatusProcessing
		} else if i%15 == 0 {
			status = models.WorkflowStatusFailed
		}

		record := models.WorkflowHistory{
			RequestID:       fmt.Sprintf("opt-req-%s-%d", repoName, i),
			Status:          status,
			Tasks:           fmt.Sprintf("Optimized test task %d for %s", i, repoName),
			RepositoryName:  repoName,
			WorkingDir:      stringPtr(fmt.Sprintf("/opt/path/to/%s/task-%d", repoName, i)),
			ClaudeCmd:       stringPtr("claude code --optimized-test"),
			Interactive:     i%2 == 0,
			ContinueTask:    i%3 == 0,
			CreatedAt:       baseTime.Add(-time.Duration(i) * time.Minute),
			CompletedAt:     timePtr(baseTime.Add(-time.Duration(i-1) * time.Minute)),
			ProcessingTimeMs: int64Ptr(int64(25000 + i*500)), // Optimized processing times
			Result:          stringPtr(fmt.Sprintf("Optimized test result %d", i)),
		}

		if status == models.WorkflowStatusFailed {
			record.Error = stringPtr(fmt.Sprintf("Optimized test error %d", i))
			record.Result = nil
		}

		records = append(records, record)
	}

	// Insert all records at once with CreateInBatches
	err := suite.db.CreateInBatches(records, 100).Error
	require.NoError(t, err, "Failed to seed %d records for repo %s", count, repoName)
}

// TestOptimizedService_CachePerformance tests caching effectiveness
func TestOptimizedService_CachePerformance(t *testing.T) {
	suite := setupOptimizedTestSuite(t)

	// Seed test data
	suite.seedOptimizedData(t, "cache-test-repo", 500)

	params := &services.PaginationParams{Page: 1, Limit: 20}
	ctx := context.Background()

	// First request (cache miss)
	start1 := time.Now()
	response1, err := suite.optimizedService.OptimizedGetTaskHistory(ctx, "cache-test-repo", params)
	elapsed1 := time.Since(start1)
	require.NoError(t, err)
	require.NotNil(t, response1)
	assert.Len(t, response1.Data, 20)

	// Second request (cache hit)
	start2 := time.Now()
	response2, err := suite.optimizedService.OptimizedGetTaskHistory(ctx, "cache-test-repo", params)
	elapsed2 := time.Since(start2)
	require.NoError(t, err)
	require.NotNil(t, response2)

	// Cache hit should be significantly faster
	assert.Less(t, elapsed2, elapsed1/2, "Cache hit should be at least 2x faster than cache miss")
	assert.Equal(t, response1.Pagination.Total, response2.Pagination.Total)

	t.Logf("Cache performance:")
	t.Logf("  Cache miss: %v", elapsed1)
	t.Logf("  Cache hit: %v", elapsed2)
	t.Logf("  Speedup: %.2fx", float64(elapsed1)/float64(elapsed2))
}

// TestOptimizedService_RepositoryExistsCache tests repository existence caching
func TestOptimizedService_RepositoryExistsCache(t *testing.T) {
	suite := setupOptimizedTestSuite(t)

	// Seed test data
	suite.seedOptimizedData(t, "exists-test-repo", 100)

	ctx := context.Background()

	// First check (cache miss)
	start1 := time.Now()
	exists1, err := suite.optimizedService.CheckRepositoryExistsOptimized(ctx, "exists-test-repo")
	elapsed1 := time.Since(start1)
	require.NoError(t, err)
	assert.True(t, exists1)

	// Second check (cache hit)
	start2 := time.Now()
	exists2, err := suite.optimizedService.CheckRepositoryExistsOptimized(ctx, "exists-test-repo")
	elapsed2 := time.Since(start2)
	require.NoError(t, err)
	assert.True(t, exists2)

	// Cache hit should be faster
	assert.Less(t, elapsed2, elapsed1, "Cache hit should be faster than cache miss")

	t.Logf("Repository exists cache performance:")
	t.Logf("  Cache miss: %v", elapsed1)
	t.Logf("  Cache hit: %v", elapsed2)
}

// TestOptimizedService_CacheInvalidation tests cache invalidation
func TestOptimizedService_CacheInvalidation(t *testing.T) {
	suite := setupOptimizedTestSuite(t)

	// Seed test data
	suite.seedOptimizedData(t, "invalidation-test-repo", 100)

	params := &services.PaginationParams{Page: 1, Limit: 20}
	ctx := context.Background()

	// Initial request to populate cache
	response1, err := suite.optimizedService.OptimizedGetTaskHistory(ctx, "invalidation-test-repo", params)
	require.NoError(t, err)
	initialTotal := response1.Pagination.Total

	// Verify cache is populated
	cacheStats := suite.optimizedService.GetCacheStats()
	assert.Greater(t, cacheStats["data_cache_entries"], 0)
	assert.Greater(t, cacheStats["count_cache_entries"], 0)

	// Invalidate cache for repository
	suite.optimizedService.InvalidateRepositoryCache("invalidation-test-repo")

	// Add new data (simulating real data change)
	suite.seedOptimizedData(t, "invalidation-test-repo", 50)

	// Request again - should see updated data
	response2, err := suite.optimizedService.OptimizedGetTaskHistory(ctx, "invalidation-test-repo", params)
	require.NoError(t, err)

	// Should have more total records
	assert.Greater(t, response2.Pagination.Total, initialTotal)

	t.Logf("Cache invalidation test:")
	t.Logf("  Initial total: %d", initialTotal)
	t.Logf("  After invalidation and new data: %d", response2.Pagination.Total)
}

// TestOptimizedService_CacheTimeout tests cache expiration
func TestOptimizedService_CacheTimeout(t *testing.T) {
	suite := setupOptimizedTestSuite(t)

	// Set short cache timeout for testing
	suite.optimizedService.SetCacheTimeout(100 * time.Millisecond)

	// Seed test data
	suite.seedOptimizedData(t, "timeout-test-repo", 100)

	params := &services.PaginationParams{Page: 1, Limit: 20}
	ctx := context.Background()

	// Initial request to populate cache
	start1 := time.Now()
	_, err := suite.optimizedService.OptimizedGetTaskHistory(ctx, "timeout-test-repo", params)
	elapsed1 := time.Since(start1)
	require.NoError(t, err)

	// Immediate second request (cache hit)
	start2 := time.Now()
	_, err = suite.optimizedService.OptimizedGetTaskHistory(ctx, "timeout-test-repo", params)
	elapsed2 := time.Since(start2)
	require.NoError(t, err)

	// Should be faster (cache hit)
	assert.Less(t, elapsed2, elapsed1)

	// Wait for cache timeout
	time.Sleep(150 * time.Millisecond)

	// Third request (cache miss due to timeout)
	start3 := time.Now()
	_, err = suite.optimizedService.OptimizedGetTaskHistory(ctx, "timeout-test-repo", params)
	elapsed3 := time.Since(start3)
	require.NoError(t, err)

	// Should be slower again (cache miss)
	assert.Greater(t, elapsed3, elapsed2)

	t.Logf("Cache timeout test:")
	t.Logf("  First request (cache miss): %v", elapsed1)
	t.Logf("  Second request (cache hit): %v", elapsed2)
	t.Logf("  Third request (cache miss after timeout): %v", elapsed3)
}

// TestOptimizedService_PreparedStatements tests prepared statement performance
func TestOptimizedService_PreparedStatements(t *testing.T) {
	suite := setupOptimizedTestSuite(t)

	// Seed test data
	suite.seedOptimizedData(t, "prepared-stmt-repo", 200)

	params := &services.PaginationParams{Page: 1, Limit: 20}
	ctx := context.Background()

	// Test with prepared statements enabled
	suite.optimizedService.EnablePreparedStatements(true)
	suite.optimizedService.ClearCache() // Clear cache to ensure fresh queries

	start1 := time.Now()
	for i := 0; i < 10; i++ {
		_, err := suite.optimizedService.OptimizedGetTaskHistory(ctx, "prepared-stmt-repo", params)
		require.NoError(t, err)
	}
	elapsed1 := time.Since(start1)

	// Test with prepared statements disabled
	suite.optimizedService.EnablePreparedStatements(false)
	suite.optimizedService.ClearCache() // Clear cache to ensure fresh queries

	start2 := time.Now()
	for i := 0; i < 10; i++ {
		_, err := suite.optimizedService.OptimizedGetTaskHistory(ctx, "prepared-stmt-repo", params)
		require.NoError(t, err)
	}
	elapsed2 := time.Since(start2)

	t.Logf("Prepared statements performance:")
	t.Logf("  With prepared statements (10 requests): %v", elapsed1)
	t.Logf("  Without prepared statements (10 requests): %v", elapsed2)
	t.Logf("  Performance ratio: %.2fx", float64(elapsed2)/float64(elapsed1))
}

// TestOptimizedService_LargeResultSet tests performance with large datasets
func TestOptimizedService_LargeResultSet(t *testing.T) {
	suite := setupOptimizedTestSuite(t)

	// Seed large test dataset
	suite.seedOptimizedData(t, "large-dataset-repo", 2000)

	testCases := []struct {
		name  string
		page  int
		limit int
	}{
		{"first_page_small", 1, 20},
		{"middle_page_small", 10, 20},
		{"last_page_small", 100, 20},
		{"first_page_large", 1, 100},
		{"middle_page_large", 10, 100},
	}

	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := &services.PaginationParams{Page: tc.page, Limit: tc.limit}

			start := time.Now()
			response, err := suite.optimizedService.OptimizedGetTaskHistory(ctx, "large-dataset-repo", params)
			elapsed := time.Since(start)

			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Len(t, response.Data, tc.limit)
			assert.Equal(t, 2000, response.Pagination.Total)

			// Performance assertion: should be under 10ms even for large datasets
			assert.Less(t, elapsed, 10*time.Millisecond, "Query should complete under 10ms")

			t.Logf("%s: %v (page %d, limit %d)", tc.name, elapsed, tc.page, tc.limit)
		})
	}
}

// TestOptimizedService_ConcurrentRequests tests concurrent access performance
func TestOptimizedService_ConcurrentRequests(t *testing.T) {
	suite := setupOptimizedTestSuite(t)

	// Seed test data
	suite.seedOptimizedData(t, "concurrent-repo", 1000)

	params := &services.PaginationParams{Page: 1, Limit: 20}
	ctx := context.Background()

	concurrencyLevels := []int{1, 5, 10, 20}

	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("concurrency_%d", concurrency), func(t *testing.T) {
			start := time.Now()

			// Run concurrent requests
			errChan := make(chan error, concurrency)
			for i := 0; i < concurrency; i++ {
				go func() {
					_, err := suite.optimizedService.OptimizedGetTaskHistory(ctx, "concurrent-repo", params)
					errChan <- err
				}()
			}

			// Collect results
			for i := 0; i < concurrency; i++ {
				err := <-errChan
				require.NoError(t, err)
			}

			elapsed := time.Since(start)
			throughput := float64(concurrency) / elapsed.Seconds()

			t.Logf("Concurrency %d: %v (%.2f req/sec)", concurrency, elapsed, throughput)

			// Performance assertion: should handle concurrent requests efficiently
			assert.Less(t, elapsed, 100*time.Millisecond, "Concurrent requests should complete quickly")
		})
	}
}

// TestOptimizedService_PerformanceMetrics tests metrics collection
func TestOptimizedService_PerformanceMetrics(t *testing.T) {
	suite := setupOptimizedTestSuite(t)

	// Seed test data
	suite.seedOptimizedData(t, "metrics-repo", 100)

	params := &services.PaginationParams{Page: 1, Limit: 20}
	ctx := context.Background()

	// Make some requests to populate cache
	_, err := suite.optimizedService.OptimizedGetTaskHistory(ctx, "metrics-repo", params)
	require.NoError(t, err)

	_, err = suite.optimizedService.CheckRepositoryExistsOptimized(ctx, "metrics-repo")
	require.NoError(t, err)

	// Get performance metrics
	metrics := suite.optimizedService.GetPerformanceMetrics()

	// Verify metrics structure
	assert.Contains(t, metrics, "cache_enabled")
	assert.Contains(t, metrics, "prepared_stmts_enabled")
	assert.Contains(t, metrics, "cache_timeout_minutes")
	assert.Contains(t, metrics, "data_cache_entries")
	assert.Contains(t, metrics, "count_cache_entries")

	// Verify values
	assert.True(t, metrics["cache_enabled"].(bool))
	assert.True(t, metrics["prepared_stmts_enabled"].(bool))
	assert.Greater(t, metrics["data_cache_entries"].(int), 0)
	assert.Greater(t, metrics["count_cache_entries"].(int), 0)

	t.Logf("Performance metrics: %+v", metrics)
}

// Helper functions
func stringPtr(s string) *string    { return &s }
func timePtr(t time.Time) *time.Time { return &t }
func int64Ptr(i int64) *int64       { return &i }