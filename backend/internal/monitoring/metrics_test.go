package monitoring

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "local-backend-server/internal/errors"
)

func TestErrorMetrics(t *testing.T) {
	metrics := NewErrorMetrics()

	t.Run("Initial state", func(t *testing.T) {
		stats := metrics.GetStats()
		assert.Equal(t, int64(0), stats.TotalErrors)
		assert.Equal(t, int64(0), stats.TotalRequests)
		assert.Equal(t, float64(0), stats.ErrorRate)
		assert.Empty(t, stats.RecentErrors)
	})

	t.Run("Record requests and errors", func(t *testing.T) {
		// Record some requests
		metrics.RecordRequest()
		metrics.RecordRequest()
		metrics.RecordRequest()

		// Record an error
		appErr := apperrors.NewDatabaseConnectionError(nil).
			WithRequestID("req-123").
			WithContext("operation", "test")
		metrics.RecordError(appErr)

		stats := metrics.GetStats()
		assert.Equal(t, int64(3), stats.TotalRequests)
		assert.Equal(t, int64(1), stats.TotalErrors)
		assert.InDelta(t, 0.333, stats.ErrorRate, 0.01)
		assert.Len(t, stats.RecentErrors, 1)

		// Check error categorization
		assert.Equal(t, int64(1), stats.ErrorsByCode[apperrors.ErrCodeDatabaseConnection])
		assert.Equal(t, int64(1), stats.ErrorsBySeverity[apperrors.SeverityCritical])
	})

	t.Run("Health status", func(t *testing.T) {
		health := metrics.GetHealthStatus()
		assert.Equal(t, "unhealthy", health.Status) // Critical error makes it unhealthy
		assert.Equal(t, int64(1), health.CriticalErrors)
		assert.NotNil(t, health.LastErrorTime)
	})
}

func TestHealthStatusDetermination(t *testing.T) {
	tests := []struct {
		name               string
		errorRate          float64
		criticalErrors     int64
		highErrors         int64
		expectedStatus     string
	}{
		{
			name:           "Healthy system",
			errorRate:      0.01,
			criticalErrors: 0,
			highErrors:     1,
			expectedStatus: "healthy",
		},
		{
			name:           "Degraded system - high error rate",
			errorRate:      0.08,
			criticalErrors: 0,
			highErrors:     5,
			expectedStatus: "degraded",
		},
		{
			name:           "Degraded system - many high errors",
			errorRate:      0.02,
			criticalErrors: 0,
			highErrors:     8, // Reduced to stay below unhealthy threshold
			expectedStatus: "degraded",
		},
		{
			name:           "Unhealthy system - critical errors",
			errorRate:      0.02,
			criticalErrors: 1,
			highErrors:     5,
			expectedStatus: "unhealthy",
		},
		{
			name:           "Unhealthy system - very high error rate",
			errorRate:      0.15,
			criticalErrors: 0,
			highErrors:     5,
			expectedStatus: "unhealthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := NewErrorMetrics()

			// Set up the scenario
			totalRequests := int64(100)
			totalErrors := int64(float64(totalRequests) * tt.errorRate)

			for i := int64(0); i < totalRequests; i++ {
				metrics.RecordRequest()
			}

			// Add critical errors
			for i := int64(0); i < tt.criticalErrors; i++ {
				criticalErr := apperrors.NewDatabaseConnectionError(nil)
				metrics.RecordError(criticalErr)
			}

			// Add high severity errors
			for i := int64(0); i < tt.highErrors; i++ {
				highErr := apperrors.NewDatabaseTimeoutError(nil)
				metrics.RecordError(highErr)
			}

			// Add remaining errors as medium/low severity
			remainingErrors := totalErrors - tt.criticalErrors - tt.highErrors
			for i := int64(0); i < remainingErrors; i++ {
				mediumErr := apperrors.NewValidationError("test error")
				metrics.RecordError(mediumErr)
			}

			health := metrics.GetHealthStatus()
			assert.Equal(t, tt.expectedStatus, health.Status)
		})
	}
}

func TestErrorEventTracking(t *testing.T) {
	metrics := NewErrorMetrics()

	// Create errors with different properties
	errors := []*apperrors.AppError{
		apperrors.NewDatabaseConnectionError(nil).WithRequestID("req-1"),
		apperrors.NewValidationError("Invalid input").WithRequestID("req-2"),
		apperrors.NewQueuePublishError(nil).WithRequestID("req-3"),
	}

	for _, err := range errors {
		metrics.RecordError(err)
	}

	stats := metrics.GetStats()
	assert.Len(t, stats.RecentErrors, 3)

	// Check that recent errors contain correct information
	recentError := stats.RecentErrors[0]
	assert.Equal(t, apperrors.ErrCodeDatabaseConnection, recentError.Code)
	assert.Equal(t, "req-1", recentError.RequestID)
	assert.Equal(t, apperrors.SeverityCritical, recentError.Severity)
}

func TestRecentErrorsLimit(t *testing.T) {
	metrics := NewErrorMetrics()
	metrics.maxRecentErrors = 5 // Set a small limit for testing

	// Add more errors than the limit
	for i := 0; i < 10; i++ {
		err := apperrors.NewValidationError("test error")
		metrics.RecordError(err)
	}

	stats := metrics.GetStats()
	assert.Len(t, stats.RecentErrors, 5) // Should be limited to maxRecentErrors
}

func TestTimeWindowTracking(t *testing.T) {
	metrics := NewErrorMetrics()

	// Create an error with a specific timestamp
	appErr := apperrors.NewDatabaseQueryError(nil)
	appErr.Timestamp = time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
	metrics.RecordError(appErr)

	timeWindows := metrics.GetErrorsByTimeWindow()
	expectedWindow := "2025-01-15T14"
	assert.Equal(t, int64(1), timeWindows[expectedWindow])
}

func TestMetricsReset(t *testing.T) {
	metrics := NewErrorMetrics()

	// Add some data
	metrics.RecordRequest()
	metrics.RecordError(apperrors.NewValidationError("test"))

	// Verify data exists
	stats := metrics.GetStats()
	assert.Equal(t, int64(1), stats.TotalRequests)
	assert.Equal(t, int64(1), stats.TotalErrors)

	// Reset metrics
	metrics.Reset()

	// Verify data is cleared
	stats = metrics.GetStats()
	assert.Equal(t, int64(0), stats.TotalRequests)
	assert.Equal(t, int64(0), stats.TotalErrors)
	assert.Empty(t, stats.RecentErrors)
}

func TestCleanupOldTimeWindows(t *testing.T) {
	metrics := NewErrorMetrics()

	// Add errors with old timestamps
	oldTime := time.Now().Add(-25 * time.Hour)
	recentTime := time.Now().Add(-1 * time.Hour)

	oldErr := apperrors.NewValidationError("old error")
	oldErr.Timestamp = oldTime
	metrics.RecordError(oldErr)

	recentErr := apperrors.NewValidationError("recent error")
	recentErr.Timestamp = recentTime
	metrics.RecordError(recentErr)

	// Should have 2 time windows
	timeWindows := metrics.GetErrorsByTimeWindow()
	assert.Len(t, timeWindows, 2)

	// Clean up old windows
	metrics.CleanupOldTimeWindows()

	// Should have only 1 time window (recent one)
	timeWindows = metrics.GetErrorsByTimeWindow()
	assert.Len(t, timeWindows, 1)
}

func TestConcurrentAccess(t *testing.T) {
	metrics := NewErrorMetrics()

	// Test concurrent access to metrics
	done := make(chan bool, 2)

	// Goroutine 1: Record requests and errors
	go func() {
		for i := 0; i < 100; i++ {
			metrics.RecordRequest()
			if i%10 == 0 {
				metrics.RecordError(apperrors.NewValidationError("test"))
			}
		}
		done <- true
	}()

	// Goroutine 2: Read stats
	go func() {
		for i := 0; i < 50; i++ {
			_ = metrics.GetStats()
			_ = metrics.GetHealthStatus()
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify final state is consistent
	stats := metrics.GetStats()
	assert.Equal(t, int64(100), stats.TotalRequests)
	assert.Equal(t, int64(10), stats.TotalErrors)
}

func TestGlobalMetrics(t *testing.T) {
	// Reset global metrics before test
	ResetGlobalMetrics()

	t.Run("Record global error", func(t *testing.T) {
		appErr := apperrors.NewDatabaseConnectionError(nil)
		RecordGlobalError(appErr)

		stats := GetGlobalStats()
		assert.Equal(t, int64(1), stats.TotalErrors)
	})

	t.Run("Record global request", func(t *testing.T) {
		RecordGlobalRequest()

		stats := GetGlobalStats()
		assert.Equal(t, int64(1), stats.TotalRequests)
	})

	t.Run("Get global health", func(t *testing.T) {
		health := GetGlobalHealthStatus()
		assert.Equal(t, "unhealthy", health.Status) // Due to critical error
	})

	// Clean up
	ResetGlobalMetrics()
}