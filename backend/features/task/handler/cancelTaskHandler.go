package handler

import (
	"net/http"

	_interface "main/features/task/model/interface"

	"github.com/labstack/echo/v4"
)

type CancelTaskHandler struct {
	cancelTaskUseCase _interface.ICancelTaskUseCase
}

func NewCancelTaskHandler(e *echo.Echo, cancelTaskUseCase _interface.ICancelTaskUseCase) _interface.ICancelTaskHandler {
	handler := &CancelTaskHandler{
		cancelTaskUseCase: cancelTaskUseCase,
	}

	// API 라우트 등록
	e.DELETE("/api/v1/tasks/:id", handler.CancelTask)

	return handler
}

// CancelTask handles DELETE /api/v1/tasks/{id}
// @Summary Cancel a task
// @Description Cancel a task (soft delete)
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task Request ID"
// @Success 200 {object} response.CancelTaskResponse
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/tasks/{id} [delete]
func (h *CancelTaskHandler) CancelTask(c echo.Context) error {
	requestID := c.Param("id")
	if requestID == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Request ID is required",
		})
	}

	result, err := h.cancelTaskUseCase.CancelTask(requestID)
	if err != nil {
		if err.Error() == "failed to get task: task not found with request_id: "+requestID {
			return c.JSON(http.StatusNotFound, map[string]any{
				"error":   "Task not found",
				"details": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error":   "Failed to cancel task",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, result)
}

