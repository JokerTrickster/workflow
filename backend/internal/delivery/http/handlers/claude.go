package handlers

import (
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"local-backend-server/internal/domain/entities"
	"local-backend-server/internal/infrastructure/queue"
)

// ClaudeHandler handles Claude-related endpoints
type ClaudeHandler struct {
	validate  *validator.Validate
	publisher *queue.Publisher
}

// NewClaudeHandler creates a new ClaudeHandler
func NewClaudeHandler(publisher *queue.Publisher) *ClaudeHandler {
	return &ClaudeHandler{
		validate:  validator.New(),
		publisher: publisher,
	}
}

// ReqRunTasksClaude represents the request structure for running Claude tasks
// This matches the structure used by local-backend
type ReqRunTasksClaude struct {
	Tasks          string `json:"tasks" validate:"required"`           // 실행할 작업 내용
	RepositoryName string `json:"repository_name" validate:"required"` // 레포지토리 이름 (필수)
	WorkingDir     string `json:"working_dir,omitempty"`               // 작업 디렉토리 (옵션)
	Interactive    bool   `json:"interactive,omitempty"`               // 대화형 모드: 여러 작업을 순차 실행
	ClaudeCmd      string `json:"claude_cmd,omitempty"`                // Claude CLI 명령어 경로 (옵션)
	ContinueTask   bool   `json:"continue_task,omitempty"`             // 기존 작업 이어서 하기 (옵션)
}

// ClaudeTaskResponse represents the response structure for Claude task execution
type ClaudeTaskResponse struct {
	RequestID string    `json:"request_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// RunTasks handles the Claude task execution request
// POST /api/v1/claude/run-tasks
func (h *ClaudeHandler) RunTasks(c echo.Context) error {
	var req ReqRunTasksClaude
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate the request
	if err := h.validate.Struct(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Generate request ID
	requestID := "req-" + generateRequestID()
	sessionID := "session-" + generateRequestID()

	// Convert to internal request format for queue processing
	requestInput := map[string]interface{}{
		"tasks":           req.Tasks,
		"repository_name": req.RepositoryName,
		"working_dir":     req.WorkingDir,
		"interactive":     req.Interactive,
		"claude_cmd":      req.ClaudeCmd,
		"continue_task":   req.ContinueTask,
	}

	// Create workflow message for queue
	workflowMessage := &queue.WorkflowMessage{
		Type:      string(entities.MessageTypeClaudeTask),
		ID:        requestID,
		SessionID: sessionID,
		Payload: map[string]interface{}{
			"request_type": string(entities.RequestTypeClaudeTask),
			"input":        requestInput,
		},
		Timestamp: time.Now(),
	}

	// Publish message to queue if publisher is available
	if h.publisher != nil {
		if err := h.publisher.PublishMessage(workflowMessage); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to queue task: "+err.Error())
		}
	}

	// Return response
	response := ClaudeTaskResponse{
		RequestID: requestID,
		Status:    "pending",
		Message:   "Claude task has been queued for processing",
		CreatedAt: time.Now(),
	}

	return c.JSON(http.StatusAccepted, response)
}

// generateRequestID creates a simple request ID
// TODO: Replace with proper UUID generation
func generateRequestID() string {
	return time.Now().Format("20060102150405")
}