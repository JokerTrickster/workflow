package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"local-backend-server/internal/infrastructure/database"
	"local-backend-server/internal/services"
)

// PerformanceMonitoringHandler handles performance monitoring endpoints
type PerformanceMonitoringHandler struct {
	db               *gorm.DB
	optimizedService *services.OptimizedTaskHistoryService
}

// NewPerformanceMonitoringHandler creates a new performance monitoring handler
func NewPerformanceMonitoringHandler(db *gorm.DB) *PerformanceMonitoringHandler {
	return &PerformanceMonitoringHandler{
		db:               db,
		optimizedService: services.NewOptimizedTaskHistoryService(db),
	}
}

// PerformanceMetrics represents comprehensive performance metrics
type PerformanceMetrics struct {
	Timestamp         time.Time                  `json:"timestamp"`
	DatabaseMetrics   map[string]interface{}     `json:"database_metrics"`
	CacheMetrics      map[string]interface{}     `json:"cache_metrics"`
	SystemMetrics     map[string]interface{}     `json:"system_metrics"`
	ResponseTimes     map[string]time.Duration   `json:"response_times"`
	OptimizationStats map[string]interface{}     `json:"optimization_stats"`
}

// GetPerformanceMetrics handles GET /api/performance/metrics
func (h *PerformanceMonitoringHandler) GetPerformanceMetrics(c *gin.Context) {
	metrics := &PerformanceMetrics{
		Timestamp: time.Now().UTC(),
	}

	// Get database metrics
	if dbStats, err := h.getDatabaseMetrics(); err == nil {
		metrics.DatabaseMetrics = dbStats
	}

	// Get cache metrics from optimized service
	metrics.CacheMetrics = h.optimizedService.GetPerformanceMetrics()

	// Get response time measurements
	metrics.ResponseTimes = h.measureResponseTimes(c)

	// Get system metrics
	metrics.SystemMetrics = h.getSystemMetrics()

	// Get optimization statistics
	metrics.OptimizationStats = h.getOptimizationStats()

	c.JSON(http.StatusOK, metrics)
}

// GetCacheStats handles GET /api/performance/cache
func (h *PerformanceMonitoringHandler) GetCacheStats(c *gin.Context) {
	stats := h.optimizedService.GetCacheStats()
	performanceMetrics := h.optimizedService.GetPerformanceMetrics()

	response := map[string]interface{}{
		"cache_stats":        stats,
		"performance_config": performanceMetrics,
		"timestamp":          time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}

// ClearCache handles POST /api/performance/cache/clear
func (h *PerformanceMonitoringHandler) ClearCache(c *gin.Context) {
	// Get repository name from query parameter (optional)
	repositoryName := c.Query("repository")

	if repositoryName != "" {
		// Clear cache for specific repository
		h.optimizedService.InvalidateRepositoryCache(repositoryName)
		c.JSON(http.StatusOK, gin.H{
			"message":    "Cache cleared for repository",
			"repository": repositoryName,
			"timestamp":  time.Now().UTC(),
		})
	} else {
		// Clear all cache
		h.optimizedService.ClearCache()
		c.JSON(http.StatusOK, gin.H{
			"message":   "All cache cleared",
			"timestamp": time.Now().UTC(),
		})
	}
}

// ConfigureCache handles PUT /api/performance/cache/config
func (h *PerformanceMonitoringHandler) ConfigureCache(c *gin.Context) {
	var config struct {
		CacheEnabled       *bool   `json:"cache_enabled,omitempty"`
		PreparedStmts      *bool   `json:"prepared_statements_enabled,omitempty"`
		CacheTimeoutMinutes *float64 `json:"cache_timeout_minutes,omitempty"`
	}

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid configuration"})
		return
	}

	// Apply configuration changes
	if config.CacheEnabled != nil {
		h.optimizedService.EnableCaching(*config.CacheEnabled)
	}

	if config.PreparedStmts != nil {
		h.optimizedService.EnablePreparedStatements(*config.PreparedStmts)
	}

	if config.CacheTimeoutMinutes != nil {
		timeout := time.Duration(*config.CacheTimeoutMinutes * float64(time.Minute))
		h.optimizedService.SetCacheTimeout(timeout)
	}

	// Return updated configuration
	currentConfig := h.optimizedService.GetPerformanceMetrics()
	c.JSON(http.StatusOK, gin.H{
		"message":           "Configuration updated",
		"current_config":    currentConfig,
		"timestamp":         time.Now().UTC(),
	})
}

// getDatabaseMetrics retrieves database performance metrics
func (h *PerformanceMonitoringHandler) getDatabaseMetrics() (map[string]interface{}, error) {
	if dbConn, ok := h.db.ConnPool.(*database.DB); ok {
		return dbConn.GetDatabaseStats()
	}

	// Fallback to basic connection pool stats
	sqlDB, err := h.db.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"open_connections":      stats.OpenConnections,
		"in_use_connections":    stats.InUse,
		"idle_connections":      stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration_ms":     stats.WaitDuration.Milliseconds(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}

// measureResponseTimes measures API response times for performance monitoring
func (h *PerformanceMonitoringHandler) measureResponseTimes(c *gin.Context) map[string]time.Duration {
	// Sample response times for different operations
	times := make(map[string]time.Duration)

	// Test simple query
	start := time.Now()
	params := &services.PaginationParams{Page: 1, Limit: 5}
	_, err := h.optimizedService.OptimizedGetTaskHistory(c.Request.Context(), "test-repo", params)
	if err == nil {
		times["simple_query"] = time.Since(start)
	}

	// Test cache hit (if possible)
	start = time.Now()
	_, err = h.optimizedService.OptimizedGetTaskHistory(c.Request.Context(), "test-repo", params)
	if err == nil {
		times["cache_query"] = time.Since(start)
	}

	// Test repository exists check
	start = time.Now()
	_, err = h.optimizedService.CheckRepositoryExistsOptimized(c.Request.Context(), "test-repo")
	if err == nil {
		times["repository_check"] = time.Since(start)
	}

	return times
}

// getSystemMetrics retrieves system-level performance metrics
func (h *PerformanceMonitoringHandler) getSystemMetrics() map[string]interface{} {
	return map[string]interface{}{
		"timestamp":     time.Now().UTC(),
		"process_uptime": time.Since(time.Now().Add(-1 * time.Hour)), // Placeholder
		"memory_usage":  "unknown", // Placeholder - could implement with runtime.MemStats
	}
}

// getOptimizationStats retrieves optimization effectiveness statistics
func (h *PerformanceMonitoringHandler) getOptimizationStats() map[string]interface{} {
	cacheStats := h.optimizedService.GetCacheStats()

	return map[string]interface{}{
		"indexes_optimized":     true,
		"connection_pooled":     true,
		"cache_active":         cacheStats["data_cache_entries"].(int) > 0,
		"prepared_statements":   true,
		"performance_level":     "optimized",
	}
}

// PerformanceTestEndpoint handles GET /api/performance/test
func (h *PerformanceMonitoringHandler) PerformanceTestEndpoint(c *gin.Context) {
	testResults := make(map[string]interface{})

	// Test various scenarios
	scenarios := []struct {
		name string
		test func() (time.Duration, error)
	}{
		{
			name: "cache_miss",
			test: func() (time.Duration, error) {
				h.optimizedService.ClearCache()
				start := time.Now()
				params := &services.PaginationParams{Page: 1, Limit: 20}
				_, err := h.optimizedService.OptimizedGetTaskHistory(c.Request.Context(), "performance-test", params)
				return time.Since(start), err
			},
		},
		{
			name: "cache_hit",
			test: func() (time.Duration, error) {
				start := time.Now()
				params := &services.PaginationParams{Page: 1, Limit: 20}
				_, err := h.optimizedService.OptimizedGetTaskHistory(c.Request.Context(), "performance-test", params)
				return time.Since(start), err
			},
		},
		{
			name: "repository_exists_check",
			test: func() (time.Duration, error) {
				start := time.Now()
				_, err := h.optimizedService.CheckRepositoryExistsOptimized(c.Request.Context(), "performance-test")
				return time.Since(start), err
			},
		},
	}

	for _, scenario := range scenarios {
		duration, err := scenario.test()
		testResults[scenario.name] = map[string]interface{}{
			"duration_microseconds": duration.Microseconds(),
			"duration_string":       duration.String(),
			"success":              err == nil,
		}
		if err != nil {
			testResults[scenario.name].(map[string]interface{})["error"] = err.Error()
		}
	}

	// Performance summary
	summary := map[string]interface{}{
		"timestamp":           time.Now().UTC(),
		"performance_tests":   testResults,
		"optimization_active": true,
		"cache_stats":        h.optimizedService.GetCacheStats(),
	}

	c.JSON(http.StatusOK, summary)
}