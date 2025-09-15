package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// getCurrentTimestampRFC3339 returns current timestamp in RFC3339 format
func getCurrentTimestampRFC3339() string {
	return time.Now().Format(time.RFC3339)
}


// HealthCheckMiddleware provides a simple health check endpoint bypass
func HealthCheckMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip middleware for health check endpoints
			if c.Request().URL.Path == "/health" || c.Request().URL.Path == "/api/health" {
				return c.JSON(http.StatusOK, map[string]interface{}{
					"status":    "healthy",
					"timestamp": getCurrentTimestampRFC3339(),
					"service":   "task-queue-api",
				})
			}
			return next(c)
		}
	}
}

// TimeoutMiddleware provides request timeout handling
func TimeoutMiddleware(timeout time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Set request timeout in context
			ctx := c.Request().Context()
			if timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
				c.SetRequest(c.Request().WithContext(ctx))
			}
			
			return next(c)
		}
	}
}

// RequestSizeLimitMiddleware limits request body size
func RequestSizeLimitMiddleware(limit int64) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().ContentLength > limit {
				return echo.NewHTTPError(http.StatusRequestEntityTooLarge, 
					"Request body too large")
			}
			return next(c)
		}
	}
}

// CompressResponseMiddleware adds response compression
func CompressResponseMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Add compression headers if client supports it
			if acceptEncoding := c.Request().Header.Get("Accept-Encoding"); acceptEncoding != "" {
				c.Response().Header().Set("Vary", "Accept-Encoding")
			}
			return next(c)
		}
	}
}