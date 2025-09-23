package handler

import (
	"net/http"

	_interface "main/features/task/model/interface"

	"github.com/labstack/echo/v4"
)

type GetTaskStatusHandler struct {
	taskUseCase _interface.ITaskUseCase
}

func NewGetTaskStatusHandler(e *echo.Echo, taskUseCase _interface.ITaskUseCase) _interface.IGetTaskStatusHandler {
	handler := &GetTaskStatusHandler{
		taskUseCase: taskUseCase,
	}

	// API 라우트 등록
	e.GET("/api/v1/tasks/:id/status", handler.GetTaskStatus)

	return handler
}

// GetTaskStatus handles GET /api/v1/tasks/{id}/status
// @Summary Get task status
// @Description Get the current status of a task
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task Request ID"
// @Success 200 {object} response.TaskResponse
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/tasks/{id}/status [get]
func (h *GetTaskStatusHandler) GetTaskStatus(c echo.Context) error {
	requestID := c.Param("id")
	if requestID == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Request ID is required",
		})
	}

	result, err := h.taskUseCase.GetTaskStatus(requestID)
	if err != nil {
		if err.Error() == "failed to get task: task not found with request_id: "+requestID {
			return c.JSON(http.StatusNotFound, map[string]any{
				"error":   "Task not found",
				"details": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error":   "Failed to get task status",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, result)
}

