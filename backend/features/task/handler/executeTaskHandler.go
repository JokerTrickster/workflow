package handler

import (
	"net/http"

	_interface "main/features/task/model/interface"

	"github.com/labstack/echo/v4"
)

type ExecuteTaskHandler struct {
	executeTaskUseCase _interface.IExecuteTaskUseCase
}

func NewExecuteTaskHandler(e *echo.Echo, executeTaskUseCase _interface.IExecuteTaskUseCase) _interface.IExecuteTaskHandler {
	handler := &ExecuteTaskHandler{
		executeTaskUseCase: executeTaskUseCase,
	}

	// API 라우트 등록
	e.POST("/api/v1/tasks/:id/execute", handler.ExecuteTask)

	return handler
}

// ExecuteTask handles POST /api/v1/tasks/{id}/execute
// @Summary Execute a task
// @Description Execute a task by queueing it to RabbitMQ
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task Request ID"
// @Success 200 {object} response.ExecuteTaskResponse
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/tasks/{id}/execute [post]
func (h *ExecuteTaskHandler) ExecuteTask(c echo.Context) error {
	requestID := c.Param("id")
	if requestID == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Request ID is required",
		})
	}

	result, err := h.executeTaskUseCase.ExecuteTask(requestID)
	if err != nil {
		if err.Error() == "failed to get task: task not found with request_id: "+requestID {
			return c.JSON(http.StatusNotFound, map[string]any{
				"error":   "Task not found",
				"details": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error":   "Failed to execute task",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, result)
}
