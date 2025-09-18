package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/JokerTrickster/workflow/local-backend/internal/domain"
)

// RequestServiceImpl implements the RequestService interface
type RequestServiceImpl struct {
	requestRepo  domain.RequestRepository
	contextRepo  domain.ProcessingContextRepository
	claudeService domain.ClaudeService
}

// NewRequestService creates a new request service
func NewRequestService(
	requestRepo domain.RequestRepository,
	contextRepo domain.ProcessingContextRepository,
	claudeService domain.ClaudeService,
) domain.RequestService {
	return &RequestServiceImpl{
		requestRepo:   requestRepo,
		contextRepo:   contextRepo,
		claudeService: claudeService,
	}
}

// CreateRequest creates a new request from a message
func (r *RequestServiceImpl) CreateRequest(ctx context.Context, message *domain.Message) (*domain.Request, error) {
	// Serialize request data
	requestData, err := json.Marshal(message.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request data: %w", err)
	}

	// Create request entity
	request := domain.NewRequest(
		message.ID,
		message.ID, // Using message ID as request ID for simplicity
		message.GetContextID(),
		string(requestData),
	)

	// Save to repository
	if err := r.requestRepo.Create(ctx, request); err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	log.Printf("Created request: %s (context: %s)", request.ID, request.ContextID)
	return request, nil
}

// ProcessRequest handles the full request processing lifecycle
func (r *RequestServiceImpl) ProcessRequest(ctx context.Context, request *domain.Request) error {
	log.Printf("Processing request: %s", request.ID)

	// Start processing
	request.Start()
	if err := r.requestRepo.Update(ctx, request); err != nil {
		return fmt.Errorf("failed to update request status: %w", err)
	}

	// Parse request data
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(request.RequestData), &payload); err != nil {
		request.Fail(fmt.Sprintf("Failed to parse request data: %v", err))
		r.requestRepo.Update(ctx, request)
		return fmt.Errorf("failed to parse request data: %w", err)
	}

	// Extract code and task from payload
	code, ok := payload["code"].(string)
	if !ok {
		request.Fail("Missing or invalid 'code' field in request")
		r.requestRepo.Update(ctx, request)
		return domain.NewInvalidRequestError("missing or invalid 'code' field")
	}

	task, ok := payload["task"].(string)
	if !ok {
		request.Fail("Missing or invalid 'task' field in request")
		r.requestRepo.Update(ctx, request)
		return domain.NewInvalidRequestError("missing or invalid 'task' field")
	}

	// Prepare analysis request
	analysisRequest := r.prepareAnalysisRequest(code, task, payload)

	// Send to Claude API
	response, err := r.claudeService.SendRequest(ctx, analysisRequest, request.ContextID)
	if err != nil {
		errorMsg := fmt.Sprintf("Claude API error: %v", err)
		request.Fail(errorMsg)
		r.requestRepo.Update(ctx, request)
		return fmt.Errorf("failed to process with Claude API: %w", err)
	}

	// Complete the request
	request.Complete(response)
	if err := r.requestRepo.Update(ctx, request); err != nil {
		log.Printf("Warning: failed to update completed request %s: %v", request.ID, err)
	}

	log.Printf("Successfully processed request: %s", request.ID)
	return nil
}

// CancelRequest cancels a pending or processing request
func (r *RequestServiceImpl) CancelRequest(ctx context.Context, requestID string) error {
	log.Printf("Cancelling request: %s", requestID)

	// Get the request
	request, err := r.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}

	// Check if request can be cancelled
	if !request.CanBeCancelled() {
		return domain.ErrRequestCannotBeCancelled
	}

	// Cancel the request
	request.Cancel()
	if err := r.requestRepo.Update(ctx, request); err != nil {
		return fmt.Errorf("failed to update cancelled request: %w", err)
	}

	log.Printf("Successfully cancelled request: %s", requestID)
	return nil
}

// GetRequestStatus returns the current status of a request
func (r *RequestServiceImpl) GetRequestStatus(ctx context.Context, requestID string) (domain.RequestStatus, error) {
	request, err := r.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return "", err
	}
	return request.Status, nil
}

// GetRequestsByContext returns all requests for a given context
func (r *RequestServiceImpl) GetRequestsByContext(ctx context.Context, contextID string) ([]*domain.Request, error) {
	return r.requestRepo.GetByContextID(ctx, contextID)
}

// prepareAnalysisRequest prepares the analysis request for Claude
func (r *RequestServiceImpl) prepareAnalysisRequest(code, task string, payload map[string]interface{}) string {
	// Build a comprehensive prompt for Claude
	prompt := fmt.Sprintf("Please analyze the following code and complete the requested task.\n\n**Task:** %s\n\n**Code to analyze:**\n```\n%s\n```\n\n**Instructions:**\n- Provide a detailed analysis of the code\n- Address the specific task mentioned above\n- Include any recommendations for improvements\n- Point out potential issues or bugs if found\n- Explain your reasoning clearly\n\nPlease provide your analysis:", task, code)

	// Add additional context if available
	if language, ok := payload["language"].(string); ok && language != "" {
		prompt = fmt.Sprintf("**Programming Language:** %s\n\n%s", language, prompt)
	}

	if framework, ok := payload["framework"].(string); ok && framework != "" {
		prompt = fmt.Sprintf("**Framework:** %s\n\n%s", framework, prompt)
	}

	if contextInfo, ok := payload["context"].(string); ok && contextInfo != "" {
		prompt = fmt.Sprintf("**Additional Context:** %s\n\n%s", contextInfo, prompt)
	}

	return prompt
}

// GetPendingRequests returns all pending requests for processing
func (r *RequestServiceImpl) GetPendingRequests(ctx context.Context) ([]*domain.Request, error) {
	return r.requestRepo.GetPendingRequests(ctx)
}

// GetProcessingRequests returns all currently processing requests
func (r *RequestServiceImpl) GetProcessingRequests(ctx context.Context) ([]*domain.Request, error) {
	return r.requestRepo.GetProcessingRequests(ctx)
}

// GetRequestStats returns statistics about requests
func (r *RequestServiceImpl) GetRequestStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count requests by status
	pendingCount, _ := r.requestRepo.CountByStatus(ctx, domain.RequestStatusPending)
	processingCount, _ := r.requestRepo.CountByStatus(ctx, domain.RequestStatusProcessing)
	completedCount, _ := r.requestRepo.CountByStatus(ctx, domain.RequestStatusCompleted)
	failedCount, _ := r.requestRepo.CountByStatus(ctx, domain.RequestStatusFailed)
	cancelledCount, _ := r.requestRepo.CountByStatus(ctx, domain.RequestStatusCancelled)

	stats["pending"] = pendingCount
	stats["processing"] = processingCount
	stats["completed"] = completedCount
	stats["failed"] = failedCount
	stats["cancelled"] = cancelledCount
	stats["total"] = pendingCount + processingCount + completedCount + failedCount + cancelledCount

	// Get recent requests (last 24 hours)
	since := time.Now().Add(-24 * time.Hour)
	recentRequests, _ := r.requestRepo.GetRequestsCreatedBetween(ctx, since, time.Now())
	stats["recent_24h"] = len(recentRequests)

	return stats, nil
}