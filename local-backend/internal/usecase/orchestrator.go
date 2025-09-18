package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JokerTrickster/workflow/local-backend/internal/domain"
)

// ServiceOrchestrator coordinates all application services
type ServiceOrchestrator struct {
	messageProcessor domain.MessageProcessor
	requestService   domain.RequestService
	contextService   domain.ContextService
	claudeService    domain.ClaudeService
	queueConsumer    domain.QueueConsumer
}

// NewServiceOrchestrator creates a new service orchestrator
func NewServiceOrchestrator(
	messageProcessor domain.MessageProcessor,
	requestService domain.RequestService,
	contextService domain.ContextService,
	claudeService domain.ClaudeService,
	queueConsumer domain.QueueConsumer,
) *ServiceOrchestrator {
	return &ServiceOrchestrator{
		messageProcessor: messageProcessor,
		requestService:   requestService,
		contextService:   contextService,
		claudeService:    claudeService,
		queueConsumer:    queueConsumer,
	}
}

// Start starts all services and begins message processing
func (s *ServiceOrchestrator) Start(ctx context.Context) error {
	log.Println("Starting service orchestrator...")

	// Create message handler that uses our message processor
	messageHandler := func(ctx context.Context, message *domain.Message) error {
		return s.messageProcessor.ProcessMessage(ctx, message)
	}

	// Start consuming messages
	if err := s.queueConsumer.StartConsuming(ctx, messageHandler); err != nil {
		return fmt.Errorf("failed to start queue consumer: %w", err)
	}

	// Start background tasks
	go s.startBackgroundTasks(ctx)

	log.Println("Service orchestrator started successfully")
	return nil
}

// Stop stops all services gracefully
func (s *ServiceOrchestrator) Stop() error {
	log.Println("Stopping service orchestrator...")

	// Stop queue consumer
	if err := s.queueConsumer.StopConsuming(); err != nil {
		log.Printf("Error stopping queue consumer: %v", err)
	}

	log.Println("Service orchestrator stopped")
	return nil
}

// startBackgroundTasks starts background maintenance tasks
func (s *ServiceOrchestrator) startBackgroundTasks(ctx context.Context) {
	// Context cleanup task - runs every hour
	contextCleanupTicker := time.NewTicker(1 * time.Hour)
	defer contextCleanupTicker.Stop()

	// Health check task - runs every 5 minutes
	healthCheckTicker := time.NewTicker(5 * time.Minute)
	defer healthCheckTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Background tasks stopping due to context cancellation")
			return

		case <-contextCleanupTicker.C:
			s.performContextCleanup(ctx)

		case <-healthCheckTicker.C:
			s.performHealthChecks(ctx)
		}
	}
}

// performContextCleanup cleans up expired contexts
func (s *ServiceOrchestrator) performContextCleanup(ctx context.Context) {
	log.Println("Performing context cleanup...")
	
	if err := s.contextService.CleanupExpiredContexts(ctx); err != nil {
		log.Printf("Context cleanup failed: %v", err)
	} else {
		log.Println("Context cleanup completed successfully")
	}
}

// performHealthChecks performs health checks on services
func (s *ServiceOrchestrator) performHealthChecks(ctx context.Context) {
	log.Println("Performing health checks...")

	// Check Claude service health
	if claudeService, ok := s.claudeService.(interface{ Health(context.Context) error }); ok {
		if err := claudeService.Health(ctx); err != nil {
			log.Printf("Claude service health check failed: %v", err)
		} else {
			log.Println("Claude service health check passed")
		}
	}

	// Check queue consumer health
	if queueConsumer, ok := s.queueConsumer.(interface{ Health() error }); ok {
		if err := queueConsumer.Health(); err != nil {
			log.Printf("Queue consumer health check failed: %v", err)
		} else {
			log.Println("Queue consumer health check passed")
		}
	}
}

// GetStats returns comprehensive statistics from all services
func (s *ServiceOrchestrator) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get request statistics
	requestStats, err := s.requestService.GetRequestStats(ctx)
	if err != nil {
		log.Printf("Failed to get request stats: %v", err)
	} else {
		stats["requests"] = requestStats
	}

	// Get context statistics
	contextStats, err := s.contextService.GetContextStats(ctx)
	if err != nil {
		log.Printf("Failed to get context stats: %v", err)
	} else {
		stats["contexts"] = contextStats
	}

	// Get Claude service stats
	if claudeService, ok := s.claudeService.(interface{ GetStats() map[string]interface{} }); ok {
		stats["claude"] = claudeService.GetStats()
	}

	// Get queue consumer stats
	if queueConsumer, ok := s.queueConsumer.(interface{ GetStats() map[string]interface{} }); ok {
		stats["queue"] = queueConsumer.GetStats()
	}

	// Add timestamp
	stats["timestamp"] = time.Now().Unix()
	stats["uptime_seconds"] = time.Since(time.Now()).Seconds() // This would be tracked properly in real implementation

	return stats, nil
}

// ProcessPendingRequests manually processes any pending requests (for recovery scenarios)
func (s *ServiceOrchestrator) ProcessPendingRequests(ctx context.Context) error {
	log.Println("Processing pending requests...")

	pendingRequests, err := s.requestService.GetPendingRequests(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending requests: %w", err)
	}

	if len(pendingRequests) == 0 {
		log.Println("No pending requests found")
		return nil
	}

	log.Printf("Found %d pending requests to process", len(pendingRequests))

	for _, request := range pendingRequests {
		log.Printf("Processing pending request: %s", request.ID)
		
		if err := s.requestService.ProcessRequest(ctx, request); err != nil {
			log.Printf("Failed to process pending request %s: %v", request.ID, err)
			continue
		}
		
		log.Printf("Successfully processed pending request: %s", request.ID)
	}

	log.Println("Finished processing pending requests")
	return nil
}

// GetServiceHealth returns health status of all services
func (s *ServiceOrchestrator) GetServiceHealth(ctx context.Context) map[string]interface{} {
	health := make(map[string]interface{})

	// Check Claude service
	if claudeService, ok := s.claudeService.(interface{ Health(context.Context) error }); ok {
		err := claudeService.Health(ctx)
		health["claude"] = map[string]interface{}{
			"healthy": err == nil,
			"error":   fmt.Sprintf("%v", err),
		}
	}

	// Check queue consumer
	if queueConsumer, ok := s.queueConsumer.(interface{ Health() error }); ok {
		err := queueConsumer.Health()
		health["queue"] = map[string]interface{}{
			"healthy": err == nil,
			"error":   fmt.Sprintf("%v", err),
		}
	}

	health["timestamp"] = time.Now().Unix()
	return health
}