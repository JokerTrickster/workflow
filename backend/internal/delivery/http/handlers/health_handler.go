package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"ai-git-workbench/internal/application/container"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	container *container.ApplicationContainer
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(container *container.ApplicationContainer) *HealthHandler {
	return &HealthHandler{
		container: container,
	}
}

// HealthResponse represents the health check response format
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Service   string            `json:"service"`
	Version   string            `json:"version"`
	Uptime    string            `json:"uptime"`
	Components map[string]string `json:"components,omitempty"`
}

var startTime = time.Now()

// Health handles GET /health - Basic health check
func (h *HealthHandler) Health(c echo.Context) error {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Service:   "task-queue-api",
		Version:   "1.0.0",
		Uptime:    time.Since(startTime).String(),
	}

	return c.JSON(http.StatusOK, response)
}

// HealthDetailed handles GET /health/detailed - Detailed health check with dependencies
func (h *HealthHandler) HealthDetailed(c echo.Context) error {
	var status string
	var httpStatus int
	components := make(map[string]string)

	// Check application container health if available
	if h.container != nil {
		if err := h.container.HealthCheck(); err != nil {
			status = "unhealthy"
			httpStatus = http.StatusServiceUnavailable
			components["container"] = "unhealthy: " + err.Error()
		} else {
			status = "healthy"
			httpStatus = http.StatusOK
			
			// Get detailed component status
			containerStatus := h.container.GetStatus()
			for k, v := range containerStatus {
				components[k] = v
			}
		}
	} else {
		status = "healthy"
		httpStatus = http.StatusOK
		components["container"] = "not_available"
	}

	response := HealthResponse{
		Status:     status,
		Timestamp:  time.Now().Format(time.RFC3339),
		Service:    "task-queue-api",
		Version:    "1.0.0",
		Uptime:     time.Since(startTime).String(),
		Components: components,
	}

	return c.JSON(httpStatus, response)
}

// Ready handles GET /ready - Readiness probe
func (h *HealthHandler) Ready(c echo.Context) error {
	// Check if service is ready to serve requests
	if h.container == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":    "not_ready",
			"message":   "container not initialized",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}

	// Perform readiness checks
	if err := h.container.HealthCheck(); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":    "not_ready",
			"message":   "dependencies not healthy",
			"error":     err.Error(),
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Live handles GET /live - Liveness probe
func (h *HealthHandler) Live(c echo.Context) error {
	// Simple liveness check - service is running
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(startTime).String(),
	})
}

// Metrics handles GET /metrics - Basic metrics endpoint
func (h *HealthHandler) Metrics(c echo.Context) error {
	metrics := map[string]interface{}{
		"uptime_seconds":     time.Since(startTime).Seconds(),
		"timestamp":          time.Now().Format(time.RFC3339),
		"service":            "task-queue-api",
		"version":            "1.0.0",
		"go_version":         "1.21",
	}

	// Add container metrics if available
	if h.container != nil {
		status := h.container.GetStatus()
		metrics["components"] = status
	}

	return c.JSON(http.StatusOK, metrics)
}