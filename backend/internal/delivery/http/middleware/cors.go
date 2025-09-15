package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:3001", 
			"http://127.0.0.1:3000",
			"http://127.0.0.1:3001",
		},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodHead,
		},
		AllowedHeaders: []string{
			"Accept",
			"Accept-Language",
			"Content-Language",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
			"Cache-Control",
			"Pragma",
		},
		ExposedHeaders: []string{
			"X-Request-ID",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// ProductionCORSConfig returns production CORS configuration
func ProductionCORSConfig(allowedOrigins []string) CORSConfig {
	return CORSConfig{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
		},
		ExposedHeaders: []string{
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           3600, // 1 hour
	}
}

// CORSMiddleware returns a CORS middleware with the specified configuration
func CORSMiddleware(config CORSConfig) echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     config.AllowedOrigins,
		AllowMethods:     config.AllowedMethods,
		AllowHeaders:     config.AllowedHeaders,
		ExposeHeaders:    config.ExposedHeaders,
		AllowCredentials: config.AllowCredentials,
		MaxAge:           config.MaxAge,
	})
}

// SecurityHeadersMiddleware adds security headers to all responses
func SecurityHeadersMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Security headers
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "DENY")
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			c.Response().Header().Set("Content-Security-Policy", "default-src 'self'")
			
			// Remove server header
			c.Response().Header().Del("Server")
			
			return next(c)
		}
	}
}

// CustomCORSMiddleware provides more fine-grained CORS control
func CustomCORSMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			origin := c.Request().Header.Get("Origin")
			method := c.Request().Method

			// Handle preflight requests
			if method == "OPTIONS" {
				return handlePreflightRequest(c, origin)
			}

			// Handle actual requests
			if origin != "" {
				if isAllowedOrigin(origin) {
					c.Response().Header().Set("Access-Control-Allow-Origin", origin)
					c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
					c.Response().Header().Set("Access-Control-Expose-Headers", 
						"X-Request-ID, X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset")
				}
			}

			return next(c)
		}
	}
}

// handlePreflightRequest handles CORS preflight requests
func handlePreflightRequest(c echo.Context, origin string) error {
	if !isAllowedOrigin(origin) {
		return c.NoContent(http.StatusForbidden)
	}

	// Set CORS headers for preflight
	c.Response().Header().Set("Access-Control-Allow-Origin", origin)
	c.Response().Header().Set("Access-Control-Allow-Methods", 
		"GET, POST, PUT, DELETE, OPTIONS, HEAD")
	c.Response().Header().Set("Access-Control-Allow-Headers", 
		"Accept, Content-Type, Authorization, X-Requested-With, X-Request-ID")
	c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
	c.Response().Header().Set("Access-Control-Max-Age", "86400")

	return c.NoContent(http.StatusNoContent)
}

// isAllowedOrigin checks if an origin is allowed
func isAllowedOrigin(origin string) bool {
	allowedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"http://127.0.0.1:3000",
		"http://127.0.0.1:3001",
		"https://app.example.com",
		"https://admin.example.com",
	}

	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
		
		// Allow subdomain matching for production domains
		if strings.HasSuffix(allowed, ".example.com") && 
		   strings.HasSuffix(origin, ".example.com") {
			return true
		}
	}

	return false
}

// RateLimitingMiddleware provides basic rate limiting
func RateLimitingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Basic rate limiting headers
			c.Response().Header().Set("X-RateLimit-Limit", "1000")
			c.Response().Header().Set("X-RateLimit-Remaining", "999")
			c.Response().Header().Set("X-RateLimit-Reset", "3600")
			
			// TODO: Implement actual rate limiting logic
			// This would typically involve checking request count per IP/user
			// against a Redis cache or in-memory store
			
			return next(c)
		}
	}
}