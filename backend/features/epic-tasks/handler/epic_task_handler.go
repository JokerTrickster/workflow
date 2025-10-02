package handler

import (
	"log"
	"net/http"
	"time"

	"main/features/epic-tasks/model/request"
	"main/features/epic-tasks/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type EpicTaskHandler struct {
	useCase   *usecase.EpicTaskUseCase
	validator *validator.Validate
}

func NewEpicTaskHandler() *EpicTaskHandler {
	return &EpicTaskHandler{
		useCase:   usecase.NewEpicTaskUseCase(),
		validator: validator.New(),
	}
}

// @Summary Get all epic tasks
// @Description Get all epic tasks for a repository with frontmatter
// @Tags epic-tasks
// @Accept json
// @Produce json
// @Param repository query string false "Repository name (default: workflow)"
// @Success 200 {object} response.EpicTasksListResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/epics/tasks [get]
func (h *EpicTaskHandler) GetAllTasks(c echo.Context) error {
	repository := c.QueryParam("repository")
	if repository == "" {
		repository = "workflow"
	}

	result, err := h.useCase.GetAllTasks(repository)
	if err != nil {
		log.Printf("Failed to get epic tasks: %v", err)

		if err.Error() == "invalid repository name: contains illegal characters" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to load tasks",
		})
	}

	// Add cache-busting headers (matching frontend implementation)
	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0, private, no-transform")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", "0")
	c.Response().Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))

	return c.JSON(http.StatusOK, result.Tasks)
}

// @Summary Get specific epic task
// @Description Get a specific epic task by ID
// @Tags epic-tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param repository query string false "Repository name (default: workflow)"
// @Success 200 {object} response.EpicTaskResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/epics/tasks/{id} [get]
func (h *EpicTaskHandler) GetTask(c echo.Context) error {
	taskID := c.Param("id")
	if taskID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Task ID is required",
		})
	}

	repository := c.QueryParam("repository")
	if repository == "" {
		repository = "workflow"
	}

	result, err := h.useCase.GetTask(repository, taskID)
	if err != nil {
		log.Printf("Failed to get epic task %s: %v", taskID, err)

		if err.Error() == "task not found" {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": "Task not found",
			})
		}

		if err.Error() == "invalid repository name: contains illegal characters" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to get task",
		})
	}

	return c.JSON(http.StatusOK, result.Task)
}

// @Summary Create new epic task
// @Description Create a new epic task
// @Tags epic-tasks
// @Accept json
// @Produce json
// @Param request body request.CreateEpicTaskRequest true "Epic task data"
// @Success 201 {object} response.EpicTaskResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/epics/tasks [post]
func (h *EpicTaskHandler) CreateTask(c echo.Context) error {
	var req request.CreateEpicTaskRequest

	// Bind request body
	if err := c.Bind(&req); err != nil {
		log.Printf("Failed to bind request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// Use default repository if not provided
	if req.Repository == "" {
		req.Repository = "workflow"
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	// Validate metadata
	if req.Task.Metadata.ID == "" || req.Task.Metadata.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Task ID and title are required",
		})
	}

	result, err := h.useCase.CreateTask(&req)
	if err != nil {
		log.Printf("Failed to create epic task: %v", err)

		if err.Error() == "task file already exists" {
			return c.JSON(http.StatusConflict, map[string]interface{}{
				"error": "Task file already exists",
			})
		}

		if err.Error() == "invalid repository name: contains illegal characters" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to create task",
		})
	}

	return c.JSON(http.StatusCreated, result.Task)
}

// @Summary Update epic task
// @Description Update an existing epic task
// @Tags epic-tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param request body request.UpdateEpicTaskRequest true "Epic task data"
// @Success 200 {object} response.EpicTaskResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/epics/tasks/{id} [put]
func (h *EpicTaskHandler) UpdateTask(c echo.Context) error {
	taskID := c.Param("id")
	if taskID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Task ID is required",
		})
	}

	var req request.UpdateEpicTaskRequest

	// Bind request body
	if err := c.Bind(&req); err != nil {
		log.Printf("Failed to bind request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// Use default repository if not provided
	if req.Repository == "" {
		req.Repository = "workflow"
	}

	// Ensure task ID matches
	if req.Task.Metadata.ID != taskID {
		req.Task.Metadata.ID = taskID
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	result, err := h.useCase.UpdateTask(&req)
	if err != nil {
		log.Printf("Failed to update epic task %s: %v", taskID, err)

		if err.Error() == "task not found" {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": "Task not found",
			})
		}

		if err.Error() == "invalid repository name: contains illegal characters" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to update task",
		})
	}

	return c.JSON(http.StatusOK, result.Task)
}

// @Summary Delete epic task
// @Description Delete an epic task
// @Tags epic-tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param repository query string false "Repository name (default: workflow)"
// @Success 200 {object} response.DeleteEpicTaskResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/epics/tasks/{id} [delete]
func (h *EpicTaskHandler) DeleteTask(c echo.Context) error {
	taskID := c.Param("id")
	if taskID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Task ID is required",
		})
	}

	repository := c.QueryParam("repository")
	if repository == "" {
		repository = "workflow"
	}

	result, err := h.useCase.DeleteTask(repository, taskID)
	if err != nil {
		log.Printf("Failed to delete epic task %s: %v", taskID, err)

		if err.Error() == "task not found" {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": "Task not found",
			})
		}

		if err.Error() == "invalid repository name: contains illegal characters" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to delete task",
		})
	}

	return c.JSON(http.StatusOK, result)
}
