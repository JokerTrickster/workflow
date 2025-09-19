package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/JokerTrickster/workflow/local-backend/internal/domain"
)

// ProcessingContextRepositoryImpl implements the ProcessingContextRepository interface
type ProcessingContextRepositoryImpl struct {
	db *gorm.DB
}

// NewProcessingContextRepository creates a new processing context repository
func NewProcessingContextRepository(db *gorm.DB) domain.ProcessingContextRepository {
	return &ProcessingContextRepositoryImpl{
		db: db,
	}
}

// Create saves a new processing context
func (r *ProcessingContextRepositoryImpl) Create(ctx context.Context, context *domain.ProcessingContext) error {
	if err := r.db.WithContext(ctx).Create(context).Error; err != nil {
		return fmt.Errorf("failed to create processing context: %w", err)
	}
	return nil
}

// GetByID retrieves a context by its ID
func (r *ProcessingContextRepositoryImpl) GetByID(ctx context.Context, id string) (*domain.ProcessingContext, error) {
	var processingContext domain.ProcessingContext
	err := r.db.WithContext(ctx).First(&processingContext, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrContextNotFound
		}
		return nil, fmt.Errorf("failed to get processing context by ID: %w", err)
	}
	return &processingContext, nil
}

// Update updates an existing context
func (r *ProcessingContextRepositoryImpl) Update(ctx context.Context, context *domain.ProcessingContext) error {
	// Update the UpdatedAt and LastUsedAt fields
	context.UpdatedAt = time.Now()
	context.LastUsedAt = time.Now()
	
	result := r.db.WithContext(ctx).Save(context)
	if result.Error != nil {
		return fmt.Errorf("failed to update processing context: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrContextNotFound
	}
	return nil
}

// Delete removes a context
func (r *ProcessingContextRepositoryImpl) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&domain.ProcessingContext{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete processing context: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrContextNotFound
	}
	return nil
}

// GetExpiredContexts retrieves contexts that haven't been used recently
func (r *ProcessingContextRepositoryImpl) GetExpiredContexts(ctx context.Context, maxAge int64) ([]*domain.ProcessingContext, error) {
	cutoffTime := time.Now().Add(-time.Duration(maxAge) * time.Second)
	
	var contexts []*domain.ProcessingContext
	err := r.db.WithContext(ctx).
		Where("last_used_at < ?", cutoffTime).
		Order("last_used_at ASC").
		Find(&contexts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get expired contexts: %w", err)
	}
	return contexts, nil
}

// CleanupExpiredContexts removes old contexts
func (r *ProcessingContextRepositoryImpl) CleanupExpiredContexts(ctx context.Context, maxAge int64) error {
	cutoffTime := time.Now().Add(-time.Duration(maxAge) * time.Second)
	
	result := r.db.WithContext(ctx).
		Where("last_used_at < ?", cutoffTime).
		Delete(&domain.ProcessingContext{})
	
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired contexts: %w", result.Error)
	}
	
	return nil
}

// Additional utility methods

// UpdateLastUsed updates the last used timestamp for a context
func (r *ProcessingContextRepositoryImpl) UpdateLastUsed(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.ProcessingContext{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_used_at": time.Now(),
			"updated_at":   time.Now(),
		})
	
	if result.Error != nil {
		return fmt.Errorf("failed to update last used time: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrContextNotFound
	}
	return nil
}

// GetContextsCreatedAfter retrieves contexts created after a specific time
func (r *ProcessingContextRepositoryImpl) GetContextsCreatedAfter(ctx context.Context, since time.Time) ([]*domain.ProcessingContext, error) {
	var contexts []*domain.ProcessingContext
	err := r.db.WithContext(ctx).
		Where("created_at > ?", since).
		Order("created_at ASC").
		Find(&contexts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get contexts created after %v: %w", since, err)
	}
	return contexts, nil
}

// CountActiveContexts counts contexts used within the specified duration
func (r *ProcessingContextRepositoryImpl) CountActiveContexts(ctx context.Context, since time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-since)
	
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.ProcessingContext{}).
		Where("last_used_at > ?", cutoffTime).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count active contexts: %w", err)
	}
	return count, nil
}

// GetAllContexts retrieves all contexts with pagination
func (r *ProcessingContextRepositoryImpl) GetAllContexts(ctx context.Context, limit, offset int) ([]*domain.ProcessingContext, error) {
	var contexts []*domain.ProcessingContext
	err := r.db.WithContext(ctx).
		Order("last_used_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&contexts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get all contexts: %w", err)
	}
	return contexts, nil
}

// GetContextStats returns statistics about contexts
func (r *ProcessingContextRepositoryImpl) GetContextStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total count
	var totalCount int64
	if err := r.db.WithContext(ctx).Model(&domain.ProcessingContext{}).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total context count: %w", err)
	}
	stats["total_contexts"] = totalCount
	
	// Active contexts (used in last 24 hours)
	activeCount, err := r.CountActiveContexts(ctx, 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to get active context count: %w", err)
	}
	stats["active_contexts_24h"] = activeCount
	
	// Expired contexts (not used in last 7 days)
	expiredContexts, err := r.GetExpiredContexts(ctx, int64((7 * 24 * time.Hour).Seconds()))
	if err != nil {
		return nil, fmt.Errorf("failed to get expired context count: %w", err)
	}
	stats["expired_contexts_7d"] = len(expiredContexts)
	
	return stats, nil
}