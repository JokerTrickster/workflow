package handler

import (
	"net/http"
	"strconv"

	_interface "main/features/task/model/interface"
	"main/features/task/model/request"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type ListTasksHandler struct {
	listTasksUseCase _interface.IListTasksUseCase
	validator        *validator.Validate
}

func NewListTasksHandler(e *echo.Echo, listTasksUseCase _interface.IListTasksUseCase) _interface.IListTasksHandler {
	handler := &ListTasksHandler{
		listTasksUseCase: listTasksUseCase,
		validator:        validator.New(),
	}

	// API 라우트 등록
	e.GET("/api/v1/tasks", handler.ListTasks)

	return handler
}

// ListTasks handles GET /api/v1/tasks
// @Summary List tasks
// @Description List tasks by repository with pagination
// @Tags tasks
// @Accept json
// @Produce json
// @Param repository_name query string true "Repository name"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Success 200 {object} response.ListTasksResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/tasks [get]
func (h *ListTasksHandler) ListTasks(c echo.Context) error {
	var req request.ListTasksRequest

	// Query 파라미터 바인딩
	req.RepositoryName = c.QueryParam("repository_name")
	if req.RepositoryName == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "repository_name query parameter is required",
		})
	}

	// 페이지 파라미터 처리
	if pageStr := c.QueryParam("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		} else {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"error": "invalid page parameter",
			})
		}
	} else {
		req.Page = 1
	}

	// 리미트 파라미터 처리
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		} else {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"error": "invalid limit parameter (must be 1-100)",
			})
		}
	} else {
		req.Limit = 20
	}

	if err := h.validator.Struct(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	result, err := h.listTasksUseCase.ListTasks(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error":   "Failed to list tasks",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, result)
}

