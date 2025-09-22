package services

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"gorm.io/gorm"

	"local-backend-server/internal/errors"
	"local-backend-server/internal/infrastructure/database/models"
)

// OptimizedTaskHistoryService provides enhanced performance for task history operations
type OptimizedTaskHistoryService struct {
	db               *gorm.DB
	cache            sync.Map
	countCache       sync.Map
	cacheTimeout     time.Duration
	enableCaching    bool
	enablePreparedStmts bool
}

// CacheEntry represents a cached query result
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// NewOptimizedTaskHistoryService creates a new optimized service
func NewOptimizedTaskHistoryService(db *gorm.DB) *OptimizedTaskHistoryService {
	return &OptimizedTaskHistoryService{
		db:               db,
		cacheTimeout:     5 * time.Minute, // 5 minute cache timeout
		enableCaching:    true,
		enablePreparedStmts: true,
	}
}

// cacheKey generates a cache key for the given parameters
func (s *OptimizedTaskHistoryService) cacheKey(repositoryName string, page, limit int) string {
	return fmt.Sprintf("task_history:%s:%d:%d", repositoryName, page, limit)
}

// countCacheKey generates a cache key for count queries
func (s *OptimizedTaskHistoryService) countCacheKey(repositoryName string) string {
	return fmt.Sprintf("task_history_count:%s", repositoryName)
}

// getCachedData retrieves data from cache if available and not expired
func (s *OptimizedTaskHistoryService) getCachedData(key string) (interface{}, bool) {
	if !s.enableCaching {
		return nil, false
	}

	if entry, ok := s.cache.Load(key); ok {
		cacheEntry := entry.(CacheEntry)
		if time.Now().Before(cacheEntry.ExpiresAt) {
			return cacheEntry.Data, true
		}
		// Remove expired entry
		s.cache.Delete(key)
	}
	return nil, false
}

// setCachedData stores data in cache with expiration
func (s *OptimizedTaskHistoryService) setCachedData(key string, data interface{}) {
	if !s.enableCaching {
		return
	}

	entry := CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(s.cacheTimeout),
	}
	s.cache.Store(key, entry)
}

// getCachedCount retrieves count from cache if available and not expired
func (s *OptimizedTaskHistoryService) getCachedCount(key string) (int64, bool) {
	if !s.enableCaching {
		return 0, false
	}

	if entry, ok := s.countCache.Load(key); ok {
		cacheEntry := entry.(CacheEntry)
		if time.Now().Before(cacheEntry.ExpiresAt) {
			return cacheEntry.Data.(int64), true
		}
		// Remove expired entry
		s.countCache.Delete(key)
	}
	return 0, false
}

// setCachedCount stores count in cache with expiration
func (s *OptimizedTaskHistoryService) setCachedCount(key string, count int64) {
	if !s.enableCaching {
		return
	}

	entry := CacheEntry{
		Data:      count,
		ExpiresAt: time.Now().Add(s.cacheTimeout),
	}
	s.countCache.Store(key, entry)
}

// OptimizedGetTaskHistory retrieves paginated task history with performance optimizations
func (s *OptimizedTaskHistoryService) OptimizedGetTaskHistory(ctx context.Context, repositoryName string, params *PaginationParams) (*TaskHistoryResponse, *errors.AppError) {
	// Validate repository name
	if err := ValidateRepositoryName(repositoryName); err != nil {
		return nil, err
	}

	// Check cache first
	cacheKey := s.cacheKey(repositoryName, params.Page, params.Limit)
	if cachedData, found := s.getCachedData(cacheKey); found {
		return cachedData.(*TaskHistoryResponse), nil
	}

	// Get total count with caching
	countKey := s.countCacheKey(repositoryName)
	var total int64
	var found bool

	if total, found = s.getCachedCount(countKey); !found {
		// Use optimized count query with proper context timeout
		countCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		// Create a new session to avoid transaction conflicts
		countDB := s.db.WithContext(countCtx).Session(&gorm.Session{NewDB: true})

		if err := countDB.
			Model(&models.WorkflowHistory{}).
			Where("repository_name = ?", repositoryName).
			Count(&total).Error; err != nil {
			return nil, errors.NewDatabaseQueryError(err).
				WithDetails("failed to count task history records")
		}

		// Cache the count
		s.setCachedCount(countKey, total)
	}

	// If no records found, return appropriate response
	if total == 0 {
		response := &TaskHistoryResponse{
			Data: []models.WorkflowHistory{},
			Pagination: PaginationMeta{
				Page:       params.Page,
				Limit:      params.Limit,
				Total:      0,
				TotalPages: 0,
			},
		}

		// Cache empty response briefly
		s.setCachedData(cacheKey, response)
		return response, nil
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))

	// Validate page doesn't exceed total pages
	if params.Page > totalPages {
		return nil, errors.NewValidationError("page exceeds available pages").
			WithDetails("requested page number exceeds total available pages")
	}

	// Query task history with optimizations
	var tasks []models.WorkflowHistory
	offset := (params.Page - 1) * params.Limit

	// Use optimized query with proper context timeout and separate session
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Create new session to avoid transaction conflicts
	queryDB := s.db.WithContext(queryCtx).Session(&gorm.Session{NewDB: true})

	// Build optimized query
	query := queryDB.
		Where("repository_name = ?", repositoryName).
		Order("created_at DESC").
		Offset(offset).
		Limit(params.Limit)

	// Enable prepared statements for better performance if configured
	if s.enablePreparedStmts {
		query = query.Session(&gorm.Session{PrepareStmt: true, NewDB: true})
	}

	if err := query.Find(&tasks).Error; err != nil {
		return nil, errors.NewDatabaseQueryError(err).
			WithDetails("failed to retrieve task history records")
	}

	response := &TaskHistoryResponse{
		Data: tasks,
		Pagination: PaginationMeta{
			Page:       params.Page,
			Limit:      params.Limit,
			Total:      int(total),
			TotalPages: totalPages,
		},
	}

	// Cache the response
	s.setCachedData(cacheKey, response)

	return response, nil
}

// CheckRepositoryExistsOptimized checks if any tasks exist for the given repository with caching
func (s *OptimizedTaskHistoryService) CheckRepositoryExistsOptimized(ctx context.Context, repositoryName string) (bool, *errors.AppError) {
	countKey := s.countCacheKey(repositoryName)

	// Check cache first
	if count, found := s.getCachedCount(countKey); found {
		return count > 0, nil
	}

	// Query database with timeout and new session
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Create new session to avoid transaction conflicts
	queryDB := s.db.WithContext(queryCtx).Session(&gorm.Session{NewDB: true})

	var count int64
	if err := queryDB.
		Model(&models.WorkflowHistory{}).
		Where("repository_name = ?", repositoryName).
		Count(&count).Error; err != nil {
		return false, errors.NewDatabaseQueryError(err).
			WithDetails("failed to check repository existence")
	}

	// Cache the count
	s.setCachedCount(countKey, count)

	return count > 0, nil
}

// ClearCache clears all cached data
func (s *OptimizedTaskHistoryService) ClearCache() {
	s.cache.Range(func(key, value interface{}) bool {
		s.cache.Delete(key)
		return true
	})

	s.countCache.Range(func(key, value interface{}) bool {
		s.countCache.Delete(key)
		return true
	})
}

// InvalidateRepositoryCache invalidates cache for a specific repository
func (s *OptimizedTaskHistoryService) InvalidateRepositoryCache(repositoryName string) {
	// Clear count cache
	countKey := s.countCacheKey(repositoryName)
	s.countCache.Delete(countKey)

	// Clear all data cache entries for this repository
	s.cache.Range(func(key, value interface{}) bool {
		keyStr := key.(string)
		if len(keyStr) > len("task_history:") &&
		   keyStr[:len("task_history:")] == "task_history:" {
			// Extract repository name from cache key
			parts := keyStr[len("task_history:"):]
			if len(parts) > 0 {
				firstColon := 0
				for i, c := range parts {
					if c == ':' {
						firstColon = i
						break
					}
				}
				if firstColon > 0 {
					repoFromKey := parts[:firstColon]
					if repoFromKey == repositoryName {
						s.cache.Delete(key)
					}
				}
			}
		}
		return true
	})
}

// GetCacheStats returns cache statistics for monitoring
func (s *OptimizedTaskHistoryService) GetCacheStats() map[string]int {
	var dataCount, countCacheCount int

	s.cache.Range(func(key, value interface{}) bool {
		dataCount++
		return true
	})

	s.countCache.Range(func(key, value interface{}) bool {
		countCacheCount++
		return true
	})

	return map[string]int{
		"data_cache_entries":  dataCount,
		"count_cache_entries": countCacheCount,
	}
}

// SetCacheTimeout configures cache timeout duration
func (s *OptimizedTaskHistoryService) SetCacheTimeout(timeout time.Duration) {
	s.cacheTimeout = timeout
}

// EnableCaching enables or disables caching
func (s *OptimizedTaskHistoryService) EnableCaching(enabled bool) {
	s.enableCaching = enabled
	if !enabled {
		s.ClearCache()
	}
}

// EnablePreparedStatements enables or disables prepared statements
func (s *OptimizedTaskHistoryService) EnablePreparedStatements(enabled bool) {
	s.enablePreparedStmts = enabled
}

// GetPerformanceMetrics returns performance metrics for monitoring
func (s *OptimizedTaskHistoryService) GetPerformanceMetrics() map[string]interface{} {
	cacheStats := s.GetCacheStats()

	metrics := map[string]interface{}{
		"cache_enabled":         s.enableCaching,
		"prepared_stmts_enabled": s.enablePreparedStmts,
		"cache_timeout_minutes": s.cacheTimeout.Minutes(),
		"data_cache_entries":    cacheStats["data_cache_entries"],
		"count_cache_entries":   cacheStats["count_cache_entries"],
	}

	return metrics
}