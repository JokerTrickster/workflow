package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// RequestValidationConfig holds validation configuration
type RequestValidationConfig struct {
	MaxBodySize    int64         // Maximum request body size in bytes
	RequiredFields []string      // Required fields for validation
	Timeout        time.Duration // Validation timeout
}

// DefaultValidationConfig returns default validation configuration
func DefaultValidationConfig() RequestValidationConfig {
	return RequestValidationConfig{
		MaxBodySize: 1024 * 1024, // 1MB
		Timeout:     5 * time.Second,
	}
}

// RequestValidationMiddleware validates incoming requests
func RequestValidationMiddleware(config RequestValidationConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Validate request size
			if err := validateRequestSize(c, config.MaxBodySize); err != nil {
				return err
			}

			// Validate content type for non-GET requests
			if err := validateContentType(c); err != nil {
				return err
			}

			// Validate query parameters
			if err := validateQueryParameters(c); err != nil {
				return err
			}

			return next(c)
		}
	}
}

// JSONValidationMiddleware validates JSON request bodies
func JSONValidationMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Only validate JSON for POST, PUT, PATCH requests
			method := c.Request().Method
			if method != "POST" && method != "PUT" && method != "PATCH" {
				return next(c)
			}

			// Check if content type is JSON
			contentType := c.Request().Header.Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				return echo.NewHTTPError(http.StatusBadRequest, 
					"Content-Type must be application/json for "+method+" requests")
			}

			// Validate JSON structure
			if err := validateJSONStructure(c); err != nil {
				return err
			}

			return next(c)
		}
	}
}

// ParameterValidationMiddleware validates path and query parameters
func ParameterValidationMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Validate path parameters
			if err := validatePathParameters(c); err != nil {
				return err
			}

			// Validate query parameters
			if err := validateQueryParameterTypes(c); err != nil {
				return err
			}

			return next(c)
		}
	}
}

// SecurityValidationMiddleware validates security-related aspects
func SecurityValidationMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Validate user agent
			if err := validateUserAgent(c); err != nil {
				return err
			}

			// Validate request origin
			if err := validateOrigin(c); err != nil {
				return err
			}

			// Check for malicious patterns
			if err := validateMaliciousPatterns(c); err != nil {
				return err
			}

			return next(c)
		}
	}
}

// Validation functions

// validateRequestSize checks if request body size is within limits
func validateRequestSize(c echo.Context, maxSize int64) error {
	if c.Request().ContentLength > maxSize {
		return echo.NewHTTPError(http.StatusRequestEntityTooLarge, 
			fmt.Sprintf("Request body too large. Maximum size: %d bytes", maxSize))
	}
	return nil
}

// validateContentType validates the content type header
func validateContentType(c echo.Context) error {
	method := c.Request().Method
	if method == "POST" || method == "PUT" || method == "PATCH" {
		contentType := c.Request().Header.Get("Content-Type")
		if contentType == "" {
			return echo.NewHTTPError(http.StatusBadRequest, 
				"Content-Type header is required for "+method+" requests")
		}
	}
	return nil
}

// validateJSONStructure validates JSON request body structure
func validateJSONStructure(c echo.Context) error {
	body := c.Request().Body
	if body == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Request body is required")
	}

	// Read and validate JSON
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to read request body")
	}

	// Restore body for handler
	c.Request().Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	// Validate JSON syntax
	var jsonData interface{}
	if err := json.Unmarshal(bodyBytes, &jsonData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, 
			"Invalid JSON format: "+err.Error())
	}

	// Validate JSON is an object (not array or primitive)
	if reflect.TypeOf(jsonData).Kind() != reflect.Map {
		return echo.NewHTTPError(http.StatusBadRequest, 
			"Request body must be a JSON object")
	}

	return nil
}

// validatePathParameters validates path parameters
func validatePathParameters(c echo.Context) error {
	// Validate task ID parameter if present
	if taskID := c.Param("id"); taskID != "" {
		if err := validateUUID(taskID, "id"); err != nil {
			return err
		}
	}

	// Validate user ID parameter if present
	if userID := c.Param("user_id"); userID != "" {
		if err := validateUUID(userID, "user_id"); err != nil {
			return err
		}
	}

	return nil
}

// validateQueryParameters validates query parameters
func validateQueryParameters(c echo.Context) error {
	// Validate user_id query parameter
	if userID := c.QueryParam("user_id"); userID != "" {
		if err := validateUUID(userID, "user_id"); err != nil {
			return err
		}
	}

	// Validate status parameter
	if status := c.QueryParam("status"); status != "" {
		validStatuses := []string{"pending", "processing", "completed", "failed", "cancelled"}
		if !contains(validStatuses, status) {
			return echo.NewHTTPError(http.StatusBadRequest, 
				"Invalid status value. Must be one of: "+strings.Join(validStatuses, ", "))
		}
	}

	// Validate pagination parameters
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err != nil || limit < 1 || limit > 100 {
			return echo.NewHTTPError(http.StatusBadRequest, 
				"Invalid limit value. Must be a number between 1 and 100")
		}
	}

	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err != nil || offset < 0 {
			return echo.NewHTTPError(http.StatusBadRequest, 
				"Invalid offset value. Must be a non-negative number")
		}
	}

	// Validate order_by parameter
	if orderBy := c.QueryParam("order_by"); orderBy != "" {
		validOrderFields := []string{"created_at", "updated_at", "title", "status", "tokens_used"}
		if !contains(validOrderFields, orderBy) {
			return echo.NewHTTPError(http.StatusBadRequest, 
				"Invalid order_by value. Must be one of: "+strings.Join(validOrderFields, ", "))
		}
	}

	// Validate order_direction parameter
	if orderDir := c.QueryParam("order_direction"); orderDir != "" {
		validDirections := []string{"ASC", "DESC", "asc", "desc"}
		if !contains(validDirections, orderDir) {
			return echo.NewHTTPError(http.StatusBadRequest, 
				"Invalid order_direction value. Must be ASC or DESC")
		}
	}

	return nil
}

// validateQueryParameterTypes validates query parameter types and formats
func validateQueryParameterTypes(c echo.Context) error {
	// Validate created_after and created_before date parameters
	if createdAfter := c.QueryParam("created_after"); createdAfter != "" {
		if _, err := time.Parse(time.RFC3339, createdAfter); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, 
				"Invalid created_after format. Must be RFC3339 timestamp")
		}
	}

	if createdBefore := c.QueryParam("created_before"); createdBefore != "" {
		if _, err := time.Parse(time.RFC3339, createdBefore); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, 
				"Invalid created_before format. Must be RFC3339 timestamp")
		}
	}

	return nil
}

// validateUserAgent validates user agent header
func validateUserAgent(c echo.Context) error {
	userAgent := c.Request().UserAgent()
	if userAgent == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "User-Agent header is required")
	}

	// Check for suspicious user agents
	suspiciousPatterns := []string{
		"<script>", "javascript:", "eval(", "alert(", "document.",
	}
	
	userAgentLower := strings.ToLower(userAgent)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(userAgentLower, pattern) {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid User-Agent header")
		}
	}

	return nil
}

// validateOrigin validates request origin
func validateOrigin(c echo.Context) error {
	origin := c.Request().Header.Get("Origin")
	if origin == "" {
		return nil // Origin is optional for non-CORS requests
	}

	// Basic origin validation
	if strings.Contains(origin, "<script>") || strings.Contains(origin, "javascript:") {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Origin header")
	}

	return nil
}

// validateMaliciousPatterns checks for common attack patterns
func validateMaliciousPatterns(c echo.Context) error {
	requestID := getRequestID(c)
	
	// Check URL for malicious patterns
	path := c.Request().URL.Path
	query := c.Request().URL.RawQuery
	
	maliciousPatterns := []string{
		"../", "..\\", "<script>", "javascript:", "eval(", "alert(", 
		"union select", "drop table", "insert into", "delete from",
		"<iframe>", "vbscript:", "data:", "onload=", "onerror=",
	}
	
	fullURL := path + "?" + query
	fullURLLower := strings.ToLower(fullURL)
	
	for _, pattern := range maliciousPatterns {
		if strings.Contains(fullURLLower, pattern) {
			requestContext := context.WithValue(c.Request().Context(), "security_violation", pattern)
			c.SetRequest(c.Request().WithContext(requestContext))
			
			// Log security violation
			logSecurityViolation(requestID, "malicious_pattern", pattern, c.RealIP(), c.Request().UserAgent())
			
			return echo.NewHTTPError(http.StatusBadRequest, "Request contains invalid characters")
		}
	}

	return nil
}

// Helper functions

// validateUUID validates if a string is a valid UUID
func validateUUID(uuid, fieldName string) error {
	// Simple UUID validation (basic format check)
	if len(uuid) != 36 {
		return echo.NewHTTPError(http.StatusBadRequest, 
			fmt.Sprintf("Invalid %s format. Must be a valid UUID", fieldName))
	}

	// Check UUID format with hyphens
	parts := strings.Split(uuid, "-")
	if len(parts) != 5 {
		return echo.NewHTTPError(http.StatusBadRequest, 
			fmt.Sprintf("Invalid %s format. Must be a valid UUID", fieldName))
	}

	// Validate each part length
	expectedLengths := []int{8, 4, 4, 4, 12}
	for i, part := range parts {
		if len(part) != expectedLengths[i] {
			return echo.NewHTTPError(http.StatusBadRequest, 
				fmt.Sprintf("Invalid %s format. Must be a valid UUID", fieldName))
		}
		
		// Check if part contains only hex characters
		for _, char := range part {
			if !isHexChar(char) {
				return echo.NewHTTPError(http.StatusBadRequest, 
					fmt.Sprintf("Invalid %s format. Must be a valid UUID", fieldName))
			}
		}
	}

	return nil
}

// isHexChar checks if a character is a valid hexadecimal character
func isHexChar(char rune) bool {
	return (char >= '0' && char <= '9') || 
		   (char >= 'a' && char <= 'f') || 
		   (char >= 'A' && char <= 'F')
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// logSecurityViolation logs security violations
func logSecurityViolation(requestID, violationType, details, remoteAddr, userAgent string) {
	securityLog := map[string]interface{}{
		"request_id":     requestID,
		"violation_type": violationType,
		"details":        details,
		"remote_addr":    remoteAddr,
		"user_agent":     userAgent,
		"timestamp":      time.Now().Format(time.RFC3339),
	}
	
	if securityJSON, err := json.Marshal(securityLog); err == nil {
		fmt.Printf("🚨 SECURITY: %s\n", string(securityJSON))
	}
}