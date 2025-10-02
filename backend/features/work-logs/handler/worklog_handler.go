package handler

import (
	"log"
	"net/http"

	"main/features/work-logs/model/request"
	"main/features/work-logs/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type WorkLogHandler struct {
	useCase   *usecase.WorkLogUseCase
	validator *validator.Validate
}

func NewWorkLogHandler() *WorkLogHandler {
	return &WorkLogHandler{
		useCase:   usecase.NewWorkLogUseCase(),
		validator: validator.New(),
	}
}

// @Summary Get work logs
// @Description Get work logs for a repository within date range
// @Tags work-logs
// @Accept json
// @Produce json
// @Param repository query string true "Repository name"
// @Param startDate query string false "Start date (YYYY-MM-DD)"
// @Param endDate query string false "End date (YYYY-MM-DD)"
// @Success 200 {array} response.DailyWorkLog
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/work-logs [get]
func (h *WorkLogHandler) GetWorkLogs(c echo.Context) error {
	var req request.GetWorkLogsRequest

	// Bind query parameters
	if err := c.Bind(&req); err != nil {
		log.Printf("Failed to bind request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request parameters",
		})
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	// Get work logs
	workLogs, err := h.useCase.GetWorkLogs(&req)
	if err != nil {
		log.Printf("Failed to get work logs: %v", err)

		// Check if it's a validation error
		if err.Error() == "repository parameter is required" ||
		   err.Error() == "invalid repository name: contains illegal characters" ||
		   err.Error() == "invalid date format, expected YYYY-MM-DD" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Internal server error",
		})
	}

	return c.JSON(http.StatusOK, workLogs)
}

// @Summary Create work log entry
// @Description Create or append a work log entry to daily log file
// @Tags work-logs
// @Accept json
// @Produce json
// @Param request body request.CreateWorkLogEntryRequest true "Work log entry"
// @Success 200 {object} response.CreateWorkLogEntryResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/work-logs/entry [post]
func (h *WorkLogHandler) CreateWorkLogEntry(c echo.Context) error {
	var req request.CreateWorkLogEntryRequest

	// Bind request body
	if err := c.Bind(&req); err != nil {
		log.Printf("Failed to bind request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	// Validate entry
	if err := h.validator.Struct(&req.Entry); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Entry validation failed",
			"details": err.Error(),
		})
	}

	// Create work log entry
	result, err := h.useCase.CreateWorkLogEntry(&req)
	if err != nil {
		log.Printf("Failed to create work log entry: %v", err)

		// Check if it's a validation error
		if err.Error() == "repository parameter is required" ||
		   err.Error() == "invalid repository name: contains illegal characters" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Internal server error",
		})
	}

	return c.JSON(http.StatusOK, result)
}
