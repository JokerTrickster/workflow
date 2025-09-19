package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"local-backend-server/internal/domain/entities"
	"local-backend-server/internal/domain/repositories"
	"local-backend-server/internal/domain/services"
	"local-backend-server/internal/infrastructure/claude"
	"local-backend-server/internal/infrastructure/queue"
)

// WorkflowOrchestrator coordinates all services and manages workflow execution
type WorkflowOrchestrator struct {
	// Domain services
	requestService  *services.RequestService
	messageService  *services.MessageService
	sessionService  *services.SessionService

	// Infrastructure services
	claudeClient     *claude.Client
	contextManager   *claude.ContextManager
	templateManager  *claude.TemplateManager
	queueConsumer    *queue.Consumer

	// Repositories
	requestRepo      repositories.RequestRepository
	messageRepo      repositories.MessageRepository
	sessionRepo      repositories.SessionRepository
	processingRepo   repositories.ProcessingContextRepository

	// Configuration
	timeoutDuration  time.Duration
	maxRetries       int
}

// NewWorkflowOrchestrator creates a new workflow orchestrator
func NewWorkflowOrchestrator(
	requestService *services.RequestService,
	messageService *services.MessageService,
	sessionService *services.SessionService,
	claudeClient *claude.Client,
	contextManager *claude.ContextManager,
	templateManager *claude.TemplateManager,
	requestRepo repositories.RequestRepository,
	messageRepo repositories.MessageRepository,
	sessionRepo repositories.SessionRepository,
	processingRepo repositories.ProcessingContextRepository,
) *WorkflowOrchestrator {
	return &WorkflowOrchestrator{
		requestService:   requestService,
		messageService:   messageService,
		sessionService:   sessionService,
		claudeClient:     claudeClient,
		contextManager:   contextManager,
		templateManager:  templateManager,
		requestRepo:      requestRepo,
		messageRepo:      messageRepo,
		sessionRepo:      sessionRepo,
		processingRepo:   processingRepo,
		timeoutDuration:  30 * time.Minute,
		maxRetries:       3,
	}
}

// SetQueueConsumer sets the queue consumer after initialization
func (wo *WorkflowOrchestrator) SetQueueConsumer(consumer *queue.Consumer) {
	wo.queueConsumer = consumer
}

// ProcessWorkflowMessage implements the MessageProcessor interface
func (wo *WorkflowOrchestrator) ProcessWorkflowMessage(ctx context.Context, msg *queue.WorkflowMessage) error {
	log.Printf("Processing workflow message: type=%s, id=%s", msg.Type, msg.ID)

	switch msg.Type {
	case string(entities.MessageTypeWorkRequest):
		return wo.processWorkRequest(ctx, msg)
	case string(entities.MessageTypeCancel):
		return wo.processCancelRequest(ctx, msg)
	case string(entities.MessageTypeStatus):
		return wo.processStatusRequest(ctx, msg)
	default:
		return fmt.Errorf("unsupported message type: %s", msg.Type)
	}
}

// processWorkRequest handles work request messages
func (wo *WorkflowOrchestrator) processWorkRequest(ctx context.Context, msg *queue.WorkflowMessage) error {
	log.Printf("Processing work request: %s", msg.ID)

	// Extract request type and input from payload
	requestTypeStr, ok := msg.Payload["request_type"].(string)
	if !ok {
		return fmt.Errorf("invalid or missing request_type in payload")
	}

	input, ok := msg.Payload["input"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid or missing input in payload")
	}

	requestType := entities.RequestType(requestTypeStr)

	// Ensure session exists or create a new one
	session, err := wo.ensureSession(ctx, msg.SessionID, msg.Payload)
	if err != nil {
		return fmt.Errorf("failed to ensure session: %w", err)
	}

	// Create request entity
	request, err := wo.requestService.CreateRequest(ctx, session.ID, requestType, input)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Start processing in a separate goroutine
	go func() {
		processingCtx := context.Background()
		if err := wo.processRequestAsync(processingCtx, request); err != nil {
			log.Printf("Failed to process request %s: %v", request.ID, err)
		}
	}()

	return nil
}

// processCancelRequest handles cancel request messages
func (wo *WorkflowOrchestrator) processCancelRequest(ctx context.Context, msg *queue.WorkflowMessage) error {
	requestID, ok := msg.Payload["request_id"].(string)
	if !ok {
		return fmt.Errorf("invalid or missing request_id in payload")
	}

	log.Printf("Processing cancel request for: %s", requestID)

	return wo.requestService.CancelRequest(ctx, requestID)
}

// processStatusRequest handles status request messages
func (wo *WorkflowOrchestrator) processStatusRequest(ctx context.Context, msg *queue.WorkflowMessage) error {
	requestID, ok := msg.Payload["request_id"].(string)
	if !ok {
		return fmt.Errorf("invalid or missing request_id in payload")
	}

	log.Printf("Processing status request for: %s", requestID)

	request, err := wo.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get request: %w", err)
	}

	if request == nil {
		return fmt.Errorf("request not found: %s", requestID)
	}

	// Log the current status (in a real implementation, this might send to a response queue)
	log.Printf("Request %s status: %s", requestID, request.Status)
	return nil
}

// processRequestAsync processes a request asynchronously
func (wo *WorkflowOrchestrator) processRequestAsync(ctx context.Context, request *entities.Request) error {
	log.Printf("Starting async processing for request: %s", request.ID)

	// Start the request
	if err := wo.requestService.StartRequest(ctx, request.ID); err != nil {
		return fmt.Errorf("failed to start request: %w", err)
	}

	// Create processing context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, wo.timeoutDuration)
	defer cancel()

	// Process the request with retries
	var lastErr error
	for attempt := 1; attempt <= wo.maxRetries; attempt++ {
		log.Printf("Processing attempt %d/%d for request: %s", attempt, wo.maxRetries, request.ID)

		result, err := wo.executeWorkflowWithRetry(timeoutCtx, request, attempt)
		if err == nil {
			// Success - complete the request
			if err := wo.requestService.CompleteRequest(timeoutCtx, request.ID, result); err != nil {
				log.Printf("Failed to mark request as completed: %v", err)
			}
			log.Printf("Successfully completed request: %s", request.ID)
			return nil
		}

		lastErr = err
		log.Printf("Attempt %d failed for request %s: %v", attempt, request.ID, err)

		// Wait before retry (except on last attempt)
		if attempt < wo.maxRetries {
			backoff := time.Duration(attempt) * time.Second
			time.Sleep(backoff)
		}
	}

	// All attempts failed
	if err := wo.requestService.FailRequest(timeoutCtx, request.ID, lastErr.Error()); err != nil {
		log.Printf("Failed to mark request as failed: %v", err)
	}

	return fmt.Errorf("request failed after %d attempts: %w", wo.maxRetries, lastErr)
}

// executeWorkflowWithRetry executes the actual workflow logic
func (wo *WorkflowOrchestrator) executeWorkflowWithRetry(ctx context.Context, request *entities.Request, attempt int) (map[string]interface{}, error) {
	// Get the latest request state
	currentRequest, err := wo.requestRepo.GetByID(ctx, request.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current request state: %w", err)
	}

	// Check if request was cancelled
	if currentRequest.Status == entities.RequestStatusCancelled {
		return nil, fmt.Errorf("request was cancelled")
	}

	// Create or get processing context
	processingContext, err := wo.contextManager.CreateContext(
		ctx, 
		request.ID, 
		request.SessionID, 
		wo.getSystemPromptForRequestType(request.Type),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create processing context: %w", err)
	}

	// Execute workflow based on request type
	switch request.Type {
	case entities.RequestTypeCodeReview:
		return wo.executeCodeReview(ctx, request, processingContext)
	case entities.RequestTypeIssueAnalysis:
		return wo.executeIssueAnalysis(ctx, request, processingContext)
	case entities.RequestTypeBugFix:
		return wo.executeBugFix(ctx, request, processingContext)
	case entities.RequestTypeFeature:
		return wo.executeFeatureImplementation(ctx, request, processingContext)
	default:
		return nil, fmt.Errorf("unsupported request type: %s", request.Type)
	}
}

// ensureSession ensures a session exists for the given session ID
func (wo *WorkflowOrchestrator) ensureSession(ctx context.Context, sessionID string, payload map[string]interface{}) (*entities.Session, error) {
	if sessionID == "" {
		// Create new session if no session ID provided
		userID := ""
		if uid, ok := payload["user_id"].(string); ok {
			userID = uid
		}
		return wo.sessionService.CreateSession(ctx, userID)
	}

	// Try to get existing session
	session, err := wo.sessionService.GetSession(ctx, sessionID)
	if err != nil {
		// Session doesn't exist or expired, create new one
		userID := ""
		if uid, ok := payload["user_id"].(string); ok {
			userID = uid
		}
		return wo.sessionService.CreateSession(ctx, userID)
	}

	return session, nil
}

// getSystemPromptForRequestType returns appropriate system prompt for request type
func (wo *WorkflowOrchestrator) getSystemPromptForRequestType(requestType entities.RequestType) string {
	switch requestType {
	case entities.RequestTypeCodeReview:
		return "You are an expert code reviewer. Provide comprehensive, constructive feedback on code quality, best practices, security, and performance."
	case entities.RequestTypeIssueAnalysis:
		return "You are a senior software engineer analyzing GitHub issues. Provide detailed technical analysis and implementation guidance."
	case entities.RequestTypeBugFix:
		return "You are a debugging expert. Analyze bugs systematically and provide robust, well-tested solutions."
	case entities.RequestTypeFeature:
		return "You are a software architect. Design scalable, maintainable solutions following best practices."
	default:
		return "You are an AI assistant helping with software development tasks. Provide helpful, accurate, and actionable responses."
	}
}

// StartConsumer starts the message consumer
func (wo *WorkflowOrchestrator) StartConsumer(ctx context.Context) error {
	if wo.queueConsumer == nil {
		log.Println("Queue consumer not available, starting in degraded mode")
		// Return nil to indicate successful start in degraded mode
		return nil
	}
	log.Println("Starting workflow orchestrator consumer")
	return wo.queueConsumer.Start(ctx)
}

// StopConsumer stops the message consumer
func (wo *WorkflowOrchestrator) StopConsumer() error {
	if wo.queueConsumer == nil {
		log.Println("Queue consumer not available, nothing to stop")
		return nil
	}
	log.Println("Stopping workflow orchestrator consumer")
	return wo.queueConsumer.Stop()
}

// HealthCheck performs health check on all components
func (wo *WorkflowOrchestrator) HealthCheck(ctx context.Context) map[string]interface{} {
	health := map[string]interface{}{
		"status": "healthy",
		"timestamp": time.Now(),
		"components": map[string]interface{}{},
	}

	// Check Claude API
	if err := wo.claudeClient.Health(ctx); err != nil {
		health["components"].(map[string]interface{})["claude_api"] = map[string]interface{}{
			"status": "unhealthy",
			"error": err.Error(),
		}
		health["status"] = "degraded"
	} else {
		health["components"].(map[string]interface{})["claude_api"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// Check message consumer
	if wo.queueConsumer == nil {
		health["components"].(map[string]interface{})["message_consumer"] = map[string]interface{}{
			"status": "unavailable",
			"error": "Queue consumer not initialized",
		}
		health["status"] = "degraded"
	} else if err := wo.queueConsumer.Health(); err != nil {
		health["components"].(map[string]interface{})["message_consumer"] = map[string]interface{}{
			"status": "unhealthy",
			"error": err.Error(),
		}
		health["status"] = "degraded"
	} else {
		health["components"].(map[string]interface{})["message_consumer"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	return health
}

// GetMetrics returns orchestrator metrics
func (wo *WorkflowOrchestrator) GetMetrics(ctx context.Context) map[string]interface{} {
	now := time.Now()
	dayAgo := now.Add(-24 * time.Hour)

	metrics, err := wo.requestRepo.GetRequestMetrics(ctx, dayAgo, now)
	if err != nil {
		log.Printf("Failed to get request metrics: %v", err)
		return map[string]interface{}{
			"error": "failed to retrieve metrics",
		}
	}

	return map[string]interface{}{
		"period": "24h",
		"total_requests": metrics.TotalRequests,
		"completed_requests": metrics.CompletedRequests,
		"failed_requests": metrics.FailedRequests,
		"success_rate": metrics.SuccessRate,
		"average_processing_time_ms": metrics.AverageTimeMs,
		"timestamp": now,
	}
}

// executeCodeReview executes code review workflow
func (wo *WorkflowOrchestrator) executeCodeReview(ctx context.Context, request *entities.Request, processingContext *entities.ProcessingContext) (map[string]interface{}, error) {
	log.Printf("Executing code review for request: %s", request.ID)

	// Extract code and context from input
	code, ok := request.Input["code"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid or missing code in request input")
	}

	context := ""
	if ctx, ok := request.Input["context"].(string); ok {
		context = ctx
	}

	// Use Claude API for code review
	response, err := wo.claudeClient.CreateCodeReview(ctx, code, context)
	if err != nil {
		return nil, fmt.Errorf("failed to create code review: %w", err)
	}

	// Create result message
	resultMessage := entities.NewMessage(
		request.SessionID,
		entities.MessageTypeWorkRequest,
		entities.MessageRoleAssistant,
		response.Content,
	)

	// Save result message
	if err := wo.messageRepo.Create(ctx, resultMessage); err != nil {
		log.Printf("Warning: failed to save code review result message: %v", err)
	}

	// Update context with response
	if err := wo.contextManager.AddMessage(ctx, processingContext, resultMessage); err != nil {
		log.Printf("Warning: failed to add message to context: %v", err)
	}

	return map[string]interface{}{
		"type": "code_review",
		"result": response.Content,
		"token_usage": response.TokenUsage,
		"message_id": resultMessage.ID,
	}, nil
}

// executeIssueAnalysis executes issue analysis workflow
func (wo *WorkflowOrchestrator) executeIssueAnalysis(ctx context.Context, request *entities.Request, processingContext *entities.ProcessingContext) (map[string]interface{}, error) {
	log.Printf("Executing issue analysis for request: %s", request.ID)

	// Extract issue content and context from input
	issue, ok := request.Input["issue"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid or missing issue in request input")
	}

	context := ""
	if ctx, ok := request.Input["context"].(string); ok {
		context = ctx
	}

	// Use Claude API for issue analysis
	response, err := wo.claudeClient.AnalyzeIssue(ctx, issue, context)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze issue: %w", err)
	}

	// Create result message
	resultMessage := entities.NewMessage(
		request.SessionID,
		entities.MessageTypeWorkRequest,
		entities.MessageRoleAssistant,
		response.Content,
	)

	// Save result message
	if err := wo.messageRepo.Create(ctx, resultMessage); err != nil {
		log.Printf("Warning: failed to save issue analysis result message: %v", err)
	}

	// Update context with response
	if err := wo.contextManager.AddMessage(ctx, processingContext, resultMessage); err != nil {
		log.Printf("Warning: failed to add message to context: %v", err)
	}

	return map[string]interface{}{
		"type": "issue_analysis",
		"result": response.Content,
		"token_usage": response.TokenUsage,
		"message_id": resultMessage.ID,
	}, nil
}

// executeBugFix executes bug fix workflow
func (wo *WorkflowOrchestrator) executeBugFix(ctx context.Context, request *entities.Request, processingContext *entities.ProcessingContext) (map[string]interface{}, error) {
	log.Printf("Executing bug fix for request: %s", request.ID)

	// Extract bug details from input
	bugDescription, ok := request.Input["bug_description"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid or missing bug_description in request input")
	}

	// Build bug fix prompt using template
	variables := map[string]interface{}{
		"bug_description": bugDescription,
		"environment": request.Input["environment"],
		"steps_to_reproduce": request.Input["steps_to_reproduce"],
		"expected_behavior": request.Input["expected_behavior"],
		"actual_behavior": request.Input["actual_behavior"],
		"code": request.Input["code"],
		"language": request.Input["language"],
		"error_logs": request.Input["error_logs"],
	}

	systemPrompt, userPrompt, err := wo.templateManager.RenderTemplate("bug_fix", variables)
	if err != nil {
		return nil, fmt.Errorf("failed to render bug fix template: %w", err)
	}

	// Process with context manager
	response, err := wo.contextManager.ProcessWithContext(ctx, processingContext, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to process bug fix with context: %w", err)
	}

	return map[string]interface{}{
		"type": "bug_fix",
		"result": response.Content,
		"token_usage": response.TokenUsage,
		"system_prompt": systemPrompt,
	}, nil
}

// executeFeatureImplementation executes feature implementation workflow
func (wo *WorkflowOrchestrator) executeFeatureImplementation(ctx context.Context, request *entities.Request, processingContext *entities.ProcessingContext) (map[string]interface{}, error) {
	log.Printf("Executing feature implementation for request: %s", request.ID)

	// Extract feature details from input
	featureDescription, ok := request.Input["feature_description"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid or missing feature_description in request input")
	}

	// Build feature implementation prompt using template
	variables := map[string]interface{}{
		"feature_description": featureDescription,
		"requirements": request.Input["requirements"],
		"system_context": request.Input["system_context"],
		"constraints": request.Input["constraints"],
		"target_users": request.Input["target_users"],
	}

	systemPrompt, userPrompt, err := wo.templateManager.RenderTemplate("feature_implementation", variables)
	if err != nil {
		return nil, fmt.Errorf("failed to render feature implementation template: %w", err)
	}

	// Process with context manager
	response, err := wo.contextManager.ProcessWithContext(ctx, processingContext, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to process feature implementation with context: %w", err)
	}

	return map[string]interface{}{
		"type": "feature_implementation",
		"result": response.Content,
		"token_usage": response.TokenUsage,
		"system_prompt": systemPrompt,
	}, nil
}