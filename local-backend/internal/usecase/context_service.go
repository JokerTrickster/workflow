package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JokerTrickster/workflow/local-backend/internal/domain"
)

// ContextServiceImpl implements the ContextService interface
type ContextServiceImpl struct {
	contextRepo domain.ProcessingContextRepository
}

// NewContextService creates a new context service
func NewContextService(contextRepo domain.ProcessingContextRepository) domain.ContextService {
	return &ContextServiceImpl{
		contextRepo: contextRepo,
	}
}

// GetOrCreateContext retrieves existing context or creates new one
func (c *ContextServiceImpl) GetOrCreateContext(ctx context.Context, contextID string) (*domain.ProcessingContext, error) {
	if contextID == "" {
		// Generate a new context ID
		contextID = fmt.Sprintf("ctx_%d", time.Now().UnixNano())
	}

	// Try to get existing context
	processingContext, err := c.contextRepo.GetByID(ctx, contextID)
	if err == domain.ErrContextNotFound {
		// Create new context
		processingContext = domain.NewProcessingContext(contextID)
		if err := c.contextRepo.Create(ctx, processingContext); err != nil {
			return nil, fmt.Errorf("failed to create processing context: %w", err)
		}
		log.Printf("Created new processing context: %s", contextID)
	} else if err != nil {
		return nil, fmt.Errorf("failed to retrieve processing context: %w", err)
	} else {
		// Update last used time
		processingContext.LastUsedAt = time.Now()
		if err := c.contextRepo.Update(ctx, processingContext); err != nil {
			log.Printf("Warning: failed to update last used time for context %s: %v", contextID, err)
		}
	}

	return processingContext, nil
}

// UpdateContext updates context with new messages
func (c *ContextServiceImpl) UpdateContext(ctx context.Context, contextID string, userMessage, assistantMessage string) error {
	// Get the context
	processingContext, err := c.contextRepo.GetByID(ctx, contextID)
	if err != nil {
		return fmt.Errorf("failed to get context %s: %w", contextID, err)
	}

	// Add messages to context
	processingContext.AddUserMessage(userMessage)
	processingContext.AddAssistantMessage(assistantMessage)

	// Update in repository
	if err := c.contextRepo.Update(ctx, processingContext); err != nil {
		return fmt.Errorf("failed to update context %s: %w", contextID, err)
	}

	log.Printf("Updated context %s with new messages", contextID)
	return nil
}

// CleanupExpiredContexts removes old unused contexts
func (c *ContextServiceImpl) CleanupExpiredContexts(ctx context.Context) error {
	// Define max age for contexts (7 days)
	maxAge := int64((7 * 24 * time.Hour).Seconds())
	
	log.Printf("Cleaning up contexts older than %d seconds", maxAge)
	
	// Get expired contexts for logging
	expiredContexts, err := c.contextRepo.GetExpiredContexts(ctx, maxAge)
	if err != nil {
		return fmt.Errorf("failed to get expired contexts: %w", err)
	}

	if len(expiredContexts) > 0 {
		log.Printf("Found %d expired contexts to clean up", len(expiredContexts))
		
		// Clean up expired contexts
		if err := c.contextRepo.CleanupExpiredContexts(ctx, maxAge); err != nil {
			return fmt.Errorf("failed to cleanup expired contexts: %w", err)
		}
		
		log.Printf("Successfully cleaned up %d expired contexts", len(expiredContexts))
	} else {
		log.Println("No expired contexts found")
	}
	
	return nil
}

// GetContextStats returns statistics about processing contexts
func (c *ContextServiceImpl) GetContextStats(ctx context.Context) (map[string]interface{}, error) {
	stats, err := c.contextRepo.GetContextStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get context stats: %w", err)
	}
	
	return stats, nil
}

// GetActiveContexts returns contexts used within the specified duration
func (c *ContextServiceImpl) GetActiveContexts(ctx context.Context, since time.Duration) ([]*domain.ProcessingContext, error) {
	cutoffTime := time.Now().Add(-since)
	return c.contextRepo.GetContextsCreatedAfter(ctx, cutoffTime)
}

// GetContextByID retrieves a specific context by ID
func (c *ContextServiceImpl) GetContextByID(ctx context.Context, contextID string) (*domain.ProcessingContext, error) {
	return c.contextRepo.GetByID(ctx, contextID)
}

// DeleteContext removes a context
func (c *ContextServiceImpl) DeleteContext(ctx context.Context, contextID string) error {
	if err := c.contextRepo.Delete(ctx, contextID); err != nil {
		return fmt.Errorf("failed to delete context %s: %w", contextID, err)
	}
	
	log.Printf("Deleted context: %s", contextID)
	return nil
}