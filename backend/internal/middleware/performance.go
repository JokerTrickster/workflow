package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// PerformanceMetrics holds performance tracking data
type PerformanceMetrics struct {
	TotalRequests   int64         `json:"total_requests"`
	AverageResponse time.Duration `json:"average_response"`
	P95Response     time.Duration `json:"p95_response"`
	P99Response     time.Duration `json:"p99_response"`
	MaxResponse     time.Duration `json:"max_response"`
	MinResponse     time.Duration `json:"min_response"`
	ErrorRate       float64       `json:"error_rate"`
	TotalErrors     int64         `json:"total_errors"`
	ResponseTimes   []time.Duration `json:"-"`
}

// PerformanceTracker tracks API performance metrics
type PerformanceTracker struct {
	mu      sync.RWMutex
	metrics map[string]*PerformanceMetrics
}

// NewPerformanceTracker creates a new performance tracker
func NewPerformanceTracker() *PerformanceTracker {
	return &PerformanceTracker{
		metrics: make(map[string]*PerformanceMetrics),
	}
}

// Global performance tracker instance
var globalTracker = NewPerformanceTracker()

// PerformanceMonitoring middleware tracks API performance metrics
func PerformanceMonitoring() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate response time
		elapsed := time.Since(start)

		// Get endpoint key
		endpoint := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())

		// Track metrics
		globalTracker.trackRequest(endpoint, elapsed, c.Writer.Status())

		// Add performance headers
		c.Header("X-Response-Time", elapsed.String())
		c.Header("X-Response-Time-Ms", fmt.Sprintf("%.2f", float64(elapsed.Nanoseconds())/1e6))
	}
}

// trackRequest records performance data for an endpoint
func (t *PerformanceTracker) trackRequest(endpoint string, responseTime time.Duration, statusCode int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.metrics[endpoint] == nil {
		t.metrics[endpoint] = &PerformanceMetrics{
			MinResponse:   responseTime,
			MaxResponse:   responseTime,
			ResponseTimes: make([]time.Duration, 0, 1000), // Pre-allocate space
		}
	}

	m := t.metrics[endpoint]
	m.TotalRequests++

	// Track response time
	m.ResponseTimes = append(m.ResponseTimes, responseTime)

	// Update min/max
	if responseTime < m.MinResponse {
		m.MinResponse = responseTime
	}
	if responseTime > m.MaxResponse {
		m.MaxResponse = responseTime
	}

	// Track errors (status codes >= 400)
	if statusCode >= 400 {
		m.TotalErrors++
	}

	// Calculate derived metrics
	m.calculateDerivedMetrics()

	// Limit response times history to prevent memory bloat
	if len(m.ResponseTimes) > 1000 {
		// Keep only the most recent 1000 entries
		copy(m.ResponseTimes, m.ResponseTimes[len(m.ResponseTimes)-1000:])
		m.ResponseTimes = m.ResponseTimes[:1000]
	}
}

// calculateDerivedMetrics computes average, percentiles, and error rate
func (m *PerformanceMetrics) calculateDerivedMetrics() {
	if len(m.ResponseTimes) == 0 {
		return
	}

	// Calculate average
	var total time.Duration
	for _, rt := range m.ResponseTimes {
		total += rt
	}
	m.AverageResponse = total / time.Duration(len(m.ResponseTimes))

	// Calculate percentiles (simplified for performance)
	times := make([]time.Duration, len(m.ResponseTimes))
	copy(times, m.ResponseTimes)

	// Simple percentile calculation for P95 and P99
	if len(times) >= 20 { // Need sufficient data for meaningful percentiles
		// Sort in ascending order (simple bubble sort for small datasets)
		for i := 0; i < len(times)-1; i++ {
			for j := 0; j < len(times)-i-1; j++ {
				if times[j] > times[j+1] {
					times[j], times[j+1] = times[j+1], times[j]
				}
			}
		}

		p95Index := int(float64(len(times)) * 0.95)
		p99Index := int(float64(len(times)) * 0.99)

		if p95Index < len(times) {
			m.P95Response = times[p95Index]
		}
		if p99Index < len(times) {
			m.P99Response = times[p99Index]
		}
	}

	// Calculate error rate
	if m.TotalRequests > 0 {
		m.ErrorRate = float64(m.TotalErrors) / float64(m.TotalRequests) * 100
	}
}

// GetMetrics returns performance metrics for all endpoints
func (t *PerformanceTracker) GetMetrics() map[string]*PerformanceMetrics {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return a copy to prevent race conditions
	result := make(map[string]*PerformanceMetrics)
	for endpoint, metrics := range t.metrics {
		// Create a copy of metrics without ResponseTimes slice
		result[endpoint] = &PerformanceMetrics{
			TotalRequests:   metrics.TotalRequests,
			AverageResponse: metrics.AverageResponse,
			P95Response:     metrics.P95Response,
			P99Response:     metrics.P99Response,
			MaxResponse:     metrics.MaxResponse,
			MinResponse:     metrics.MinResponse,
			ErrorRate:       metrics.ErrorRate,
			TotalErrors:     metrics.TotalErrors,
		}
	}

	return result
}

// GetTaskHistoryMetrics returns metrics specifically for task history endpoints
func (t *PerformanceTracker) GetTaskHistoryMetrics() map[string]*PerformanceMetrics {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make(map[string]*PerformanceMetrics)
	for endpoint, metrics := range t.metrics {
		// Filter for task history endpoints
		if endpoint == "GET /api/v1/tasks/history/:repository_name" {
			result[endpoint] = &PerformanceMetrics{
				TotalRequests:   metrics.TotalRequests,
				AverageResponse: metrics.AverageResponse,
				P95Response:     metrics.P95Response,
				P99Response:     metrics.P99Response,
				MaxResponse:     metrics.MaxResponse,
				MinResponse:     metrics.MinResponse,
				ErrorRate:       metrics.ErrorRate,
				TotalErrors:     metrics.TotalErrors,
			}
		}
	}

	return result
}

// Reset clears all performance metrics
func (t *PerformanceTracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.metrics = make(map[string]*PerformanceMetrics)
}

// GetGlobalTracker returns the global performance tracker instance
func GetGlobalTracker() *PerformanceTracker {
	return globalTracker
}

// PerformanceAlert checks if any metrics exceed thresholds
type PerformanceAlert struct {
	Endpoint    string        `json:"endpoint"`
	Metric      string        `json:"metric"`
	Value       interface{}   `json:"value"`
	Threshold   interface{}   `json:"threshold"`
	Message     string        `json:"message"`
	Timestamp   time.Time     `json:"timestamp"`
}

// CheckPerformanceAlerts returns alerts for metrics that exceed thresholds
func (t *PerformanceTracker) CheckPerformanceAlerts() []PerformanceAlert {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var alerts []PerformanceAlert
	now := time.Now()

	for endpoint, metrics := range t.metrics {
		// Check if this is a task history endpoint
		if endpoint == "GET /api/v1/tasks/history/:repository_name" {
			// Alert if average response time > 200ms
			if metrics.AverageResponse > 200*time.Millisecond {
				alerts = append(alerts, PerformanceAlert{
					Endpoint:  endpoint,
					Metric:    "average_response_time",
					Value:     metrics.AverageResponse,
					Threshold: 200 * time.Millisecond,
					Message:   "Average response time exceeds 200ms threshold",
					Timestamp: now,
				})
			}

			// Alert if P95 response time > 500ms
			if metrics.P95Response > 500*time.Millisecond {
				alerts = append(alerts, PerformanceAlert{
					Endpoint:  endpoint,
					Metric:    "p95_response_time",
					Value:     metrics.P95Response,
					Threshold: 500 * time.Millisecond,
					Message:   "P95 response time exceeds 500ms threshold",
					Timestamp: now,
				})
			}

			// Alert if error rate > 1%
			if metrics.ErrorRate > 1.0 {
				alerts = append(alerts, PerformanceAlert{
					Endpoint:  endpoint,
					Metric:    "error_rate",
					Value:     fmt.Sprintf("%.2f%%", metrics.ErrorRate),
					Threshold: "1.0%",
					Message:   "Error rate exceeds 1% threshold",
					Timestamp: now,
				})
			}
		}
	}

	return alerts
}

// RequestSizeLimit middleware to limit request size for performance
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.AbortWithStatus(413) // Request Entity Too Large
			return
		}
		c.Next()
	}
}

// ResponseTimeAlert middleware to set response time headers and alerts
func ResponseTimeAlert(threshold time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		elapsed := time.Since(start)
		if elapsed > threshold {
			c.Header("X-Performance-Alert", fmt.Sprintf("Response time %v exceeds threshold %v", elapsed, threshold))
		}
	}
}