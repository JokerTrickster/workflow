package monitoring

import (
	"sync"
	"time"

	"local-backend-server/internal/errors"
)

// ErrorMetrics tracks error rates and patterns
type ErrorMetrics struct {
	mu                  sync.RWMutex
	errorCounts         map[errors.ErrorCode]int64
	severityCounts      map[errors.ErrorSeverity]int64
	httpStatusCounts    map[int]int64
	totalErrors         int64
	totalRequests       int64
	lastResetTime       time.Time
	errorsByTimeWindow  map[string]int64 // time window -> count
	recentErrors        []ErrorEvent
	maxRecentErrors     int
}

// ErrorEvent represents a single error occurrence
type ErrorEvent struct {
	Timestamp time.Time           `json:"timestamp"`
	Code      errors.ErrorCode    `json:"code"`
	Severity  errors.ErrorSeverity `json:"severity"`
	Message   string              `json:"message"`
	RequestID string              `json:"request_id,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// ErrorStats represents error statistics
type ErrorStats struct {
	TotalErrors        int64                              `json:"total_errors"`
	TotalRequests      int64                              `json:"total_requests"`
	ErrorRate          float64                            `json:"error_rate"`
	ErrorsByCode       map[errors.ErrorCode]int64         `json:"errors_by_code"`
	ErrorsBySeverity   map[errors.ErrorSeverity]int64     `json:"errors_by_severity"`
	ErrorsByHTTPStatus map[int]int64                      `json:"errors_by_http_status"`
	TimeWindow         string                             `json:"time_window"`
	RecentErrors       []ErrorEvent                       `json:"recent_errors"`
}

// HealthStatus represents the overall system health
type HealthStatus struct {
	Status         string     `json:"status"`         // "healthy", "degraded", "unhealthy"
	ErrorRate      float64    `json:"error_rate"`
	LastErrorTime  *time.Time `json:"last_error_time,omitempty"`
	CriticalErrors int64      `json:"critical_errors"`
	HighErrors     int64      `json:"high_errors"`
	Timestamp      time.Time  `json:"timestamp"`
}

// NewErrorMetrics creates a new error metrics tracker
func NewErrorMetrics() *ErrorMetrics {
	return &ErrorMetrics{
		errorCounts:        make(map[errors.ErrorCode]int64),
		severityCounts:     make(map[errors.ErrorSeverity]int64),
		httpStatusCounts:   make(map[int]int64),
		errorsByTimeWindow: make(map[string]int64),
		recentErrors:       make([]ErrorEvent, 0),
		maxRecentErrors:    100,
		lastResetTime:      time.Now(),
	}
}

// RecordError records an error occurrence
func (em *ErrorMetrics) RecordError(appErr *errors.AppError) {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Increment counters
	em.errorCounts[appErr.Code]++
	em.severityCounts[appErr.Severity]++
	em.httpStatusCounts[appErr.HTTPStatus]++
	em.totalErrors++

	// Record by time window (hour)
	timeWindow := appErr.Timestamp.Format("2006-01-02T15")
	em.errorsByTimeWindow[timeWindow]++

	// Add to recent errors
	errorEvent := ErrorEvent{
		Timestamp: appErr.Timestamp,
		Code:      appErr.Code,
		Severity:  appErr.Severity,
		Message:   appErr.Message,
		RequestID: appErr.RequestID,
		Context:   appErr.Context,
	}

	em.recentErrors = append(em.recentErrors, errorEvent)
	
	// Keep only the most recent errors
	if len(em.recentErrors) > em.maxRecentErrors {
		em.recentErrors = em.recentErrors[len(em.recentErrors)-em.maxRecentErrors:]
	}
}

// RecordRequest records a request (for calculating error rate)
func (em *ErrorMetrics) RecordRequest() {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.totalRequests++
}

// GetStats returns current error statistics
func (em *ErrorMetrics) GetStats() ErrorStats {
	em.mu.RLock()
	defer em.mu.RUnlock()

	errorRate := float64(0)
	if em.totalRequests > 0 {
		errorRate = float64(em.totalErrors) / float64(em.totalRequests)
	}

	// Copy maps to avoid race conditions
	errorsByCode := make(map[errors.ErrorCode]int64)
	for k, v := range em.errorCounts {
		errorsByCode[k] = v
	}

	errorsBySeverity := make(map[errors.ErrorSeverity]int64)
	for k, v := range em.severityCounts {
		errorsBySeverity[k] = v
	}

	errorsByHTTPStatus := make(map[int]int64)
	for k, v := range em.httpStatusCounts {
		errorsByHTTPStatus[k] = v
	}

	// Copy recent errors
	recentErrors := make([]ErrorEvent, len(em.recentErrors))
	copy(recentErrors, em.recentErrors)

	return ErrorStats{
		TotalErrors:        em.totalErrors,
		TotalRequests:      em.totalRequests,
		ErrorRate:          errorRate,
		ErrorsByCode:       errorsByCode,
		ErrorsBySeverity:   errorsBySeverity,
		ErrorsByHTTPStatus: errorsByHTTPStatus,
		TimeWindow:         em.lastResetTime.Format(time.RFC3339),
		RecentErrors:       recentErrors,
	}
}

// GetHealthStatus returns the current system health status
func (em *ErrorMetrics) GetHealthStatus() HealthStatus {
	em.mu.RLock()
	defer em.mu.RUnlock()

	stats := em.getStatsUnsafe()
	
	status := "healthy"
	var lastErrorTime *time.Time

	// Find the most recent error
	if len(em.recentErrors) > 0 {
		mostRecent := em.recentErrors[len(em.recentErrors)-1].Timestamp
		lastErrorTime = &mostRecent
	}

	// Determine health status based on error rate and severity
	criticalErrors := em.severityCounts[errors.SeverityCritical]
	highErrors := em.severityCounts[errors.SeverityHigh]

	if stats.ErrorRate > 0.1 || criticalErrors > 0 {
		status = "unhealthy"
	} else if stats.ErrorRate > 0.05 || highErrors > 10 {
		status = "degraded"
	}

	return HealthStatus{
		Status:         status,
		ErrorRate:      stats.ErrorRate,
		LastErrorTime:  lastErrorTime,
		CriticalErrors: criticalErrors,
		HighErrors:     highErrors,
		Timestamp:      time.Now().UTC(),
	}
}

// getStatsUnsafe returns stats without locking (for internal use)
func (em *ErrorMetrics) getStatsUnsafe() ErrorStats {
	errorRate := float64(0)
	if em.totalRequests > 0 {
		errorRate = float64(em.totalErrors) / float64(em.totalRequests)
	}

	return ErrorStats{
		TotalErrors:   em.totalErrors,
		TotalRequests: em.totalRequests,
		ErrorRate:     errorRate,
	}
}

// Reset resets the metrics (useful for testing or periodic resets)
func (em *ErrorMetrics) Reset() {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.errorCounts = make(map[errors.ErrorCode]int64)
	em.severityCounts = make(map[errors.ErrorSeverity]int64)
	em.httpStatusCounts = make(map[int]int64)
	em.errorsByTimeWindow = make(map[string]int64)
	em.recentErrors = make([]ErrorEvent, 0)
	em.totalErrors = 0
	em.totalRequests = 0
	em.lastResetTime = time.Now()
}

// GetErrorsByTimeWindow returns errors grouped by time window
func (em *ErrorMetrics) GetErrorsByTimeWindow() map[string]int64 {
	em.mu.RLock()
	defer em.mu.RUnlock()

	result := make(map[string]int64)
	for k, v := range em.errorsByTimeWindow {
		result[k] = v
	}
	return result
}

// CleanupOldTimeWindows removes old time window data (older than 24 hours)
func (em *ErrorMetrics) CleanupOldTimeWindows() {
	em.mu.Lock()
	defer em.mu.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour)
	cutoffWindow := cutoff.Format("2006-01-02T15")

	for window := range em.errorsByTimeWindow {
		if window < cutoffWindow {
			delete(em.errorsByTimeWindow, window)
		}
	}
}

// Global metrics instance
var globalMetrics = NewErrorMetrics()

// RecordGlobalError records an error in the global metrics
func RecordGlobalError(appErr *errors.AppError) {
	globalMetrics.RecordError(appErr)
}

// RecordGlobalRequest records a request in the global metrics
func RecordGlobalRequest() {
	globalMetrics.RecordRequest()
}

// GetGlobalStats returns global error statistics
func GetGlobalStats() ErrorStats {
	return globalMetrics.GetStats()
}

// GetGlobalHealthStatus returns global health status
func GetGlobalHealthStatus() HealthStatus {
	return globalMetrics.GetHealthStatus()
}

// ResetGlobalMetrics resets global metrics
func ResetGlobalMetrics() {
	globalMetrics.Reset()
}