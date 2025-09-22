package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"local-backend-server/internal/monitoring"
)

// HandleErrorMetrics returns error statistics
func HandleErrorMetrics(c *gin.Context) {
	stats := monitoring.GetGlobalStats()
	c.JSON(http.StatusOK, gin.H{
		"error_metrics": stats,
	})
}

// HandleHealthDetailed returns detailed health information
func HandleHealthDetailed(c *gin.Context) {
	healthStatus := monitoring.GetGlobalHealthStatus()
	stats := monitoring.GetGlobalStats()
	
	response := gin.H{
		"health": healthStatus,
		"metrics": gin.H{
			"total_requests":   stats.TotalRequests,
			"total_errors":     stats.TotalErrors,
			"error_rate":       stats.ErrorRate,
			"errors_by_code":   stats.ErrorsByCode,
			"errors_by_severity": stats.ErrorsBySeverity,
			"recent_errors":    stats.RecentErrors,
		},
		"time_windows": monitoring.GetGlobalStats().ErrorsByCode,
	}
	
	var httpStatus int
	switch healthStatus.Status {
	case "healthy":
		httpStatus = http.StatusOK
	case "degraded":
		httpStatus = http.StatusOK
	case "unhealthy":
		httpStatus = http.StatusServiceUnavailable
	default:
		httpStatus = http.StatusInternalServerError
	}
	
	c.JSON(httpStatus, response)
}

// HandleResetMetrics resets error metrics (useful for testing)
func HandleResetMetrics(c *gin.Context) {
	monitoring.ResetGlobalMetrics()
	c.JSON(http.StatusOK, gin.H{
		"message": "Error metrics have been reset",
	})
}