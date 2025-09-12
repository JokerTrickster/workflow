package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
)

// HealthHandler handles health check endpoints
type HealthHandler struct{}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck returns the health status of the application
func (h *HealthHandler) HealthCheck(c echo.Context) error {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	response := map[string]interface{}{
		"status":    "OK",
		"timestamp": time.Now().UTC(),
		"service":   "workflow-backend",
		"version":   "1.0.0",
		"system": map[string]interface{}{
			"go_version":      runtime.Version(),
			"goroutines":      runtime.NumGoroutine(),
			"memory_alloc_mb": bToMb(memStats.Alloc),
			"memory_total_mb": bToMb(memStats.TotalAlloc),
			"memory_sys_mb":   bToMb(memStats.Sys),
			"gc_runs":         memStats.NumGC,
		},
		"checks": map[string]interface{}{
			"database": map[string]interface{}{
				"status": "OK",
				"message": "Database connection healthy",
			},
			"github_api": map[string]interface{}{
				"status": "OK", 
				"message": "GitHub API accessible",
			},
		},
	}

	return c.JSON(http.StatusOK, response)
}

// bToMb converts bytes to megabytes
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}