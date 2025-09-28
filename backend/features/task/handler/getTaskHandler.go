package handler

import (
	"net/http"

	_interface "main/features/task/model/interface"

	"github.com/labstack/echo/v4"
)

type GetTaskHandler struct {
	getTaskUseCase _interface.IGetTaskUseCase
}

func NewGetTaskHandler(e *echo.Echo, getTaskUseCase _interface.IGetTaskUseCase) _interface.IGetTaskHandler {
	handler := &GetTaskHandler{
		getTaskUseCase: getTaskUseCase,
	}

	// API 라우트 등록
	e.GET("/api/v1/tasks/:id", handler.GetTask)

	return handler
}

// GetTask handles GET /api/v1/tasks/{id}
// @Summary Get task details
// @Description Get the full details of a task
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task Request ID"
// @Success 200 {object} response.TaskResponse
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/tasks/{id} [get]
func (h *GetTaskHandler) GetTask(c echo.Context) error {
	requestID := c.Param("id")
	if requestID == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Request ID is required",
		})
	}

	result, err := h.getTaskUseCase.GetTask(requestID)
	if err != nil {
		if err.Error() == "failed to get task: task not found with request_id: "+requestID {
			return c.JSON(http.StatusNotFound, map[string]any{
				"error":   "Task not found",
				"details": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error":   "Failed to get task",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, result)
}