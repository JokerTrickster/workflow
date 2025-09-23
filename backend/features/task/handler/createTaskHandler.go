package handler

import (
	"net/http"

	_interface "main/features/task/model/interface"
	"main/features/task/model/request"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type CreateTaskHandler struct {
	createTaskUseCase _interface.ICreateTaskUseCase
	validator         *validator.Validate
}

func NewCreateTaskHandler(e *echo.Echo, createTaskUseCase _interface.ICreateTaskUseCase) _interface.ICreateTaskHandler {
	handler := &CreateTaskHandler{
		createTaskUseCase: createTaskUseCase,
		validator:         validator.New(),
	}

	// API 라우트 등록
	e.POST("/api/v1/tasks", handler.CreateTask)

	return handler
}

// CreateTask handles POST /api/v1/tasks
// @Summary Create a new task
// @Description Create a new workflow task
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body request.CreateTaskRequest true "Task creation request"
// @Success 201 {object} response.CreateTaskResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/tasks [post]
func (h *CreateTaskHandler) CreateTask(c echo.Context) error {
	var req request.CreateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	result, err := h.createTaskUseCase.CreateTask(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error":   "Failed to create task",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, result)
}

