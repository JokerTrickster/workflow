package repositories

import (
	"context"

	"gorm.io/gorm"

	"local-backend-server/internal/domain/entities"
	"local-backend-server/internal/domain/repositories"
	"local-backend-server/internal/infrastructure/database/models"
)

// sessionRepository implements the SessionRepository interface
type sessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *gorm.DB) repositories.SessionRepository {
	return &sessionRepository{db: db}
}

// Create creates a new session
func (r *sessionRepository) Create(ctx context.Context, session *entities.Session) error {
	model := &models.Session{
		ID:        session.ID,
		UserID:    session.UserID,
		Status:    string(session.Status),
		Metadata:  models.JSON(session.Metadata),
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
		ExpiresAt: session.ExpiresAt,
	}
	
	return r.db.WithContext(ctx).Create(model).Error
}

// GetByID retrieves a session by its ID
func (r *sessionRepository) GetByID(ctx context.Context, id string) (*entities.Session, error) {
	var model models.Session
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return r.modelToEntity(&model), nil
}

// GetByUserID retrieves sessions for a user
func (r *sessionRepository) GetByUserID(ctx context.Context, userID string) ([]*entities.Session, error) {
	var models []models.Session
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	
	sessions := make([]*entities.Session, len(models))
	for i, model := range models {
		sessions[i] = r.modelToEntity(&model)
	}
	
	return sessions, nil
}

// GetActiveByUserID retrieves active sessions for a user
func (r *sessionRepository) GetActiveByUserID(ctx context.Context, userID string) ([]*entities.Session, error) {
	var models []models.Session
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, string(entities.SessionStatusActive)).
		Order("updated_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	
	sessions := make([]*entities.Session, len(models))
	for i, model := range models {
		sessions[i] = r.modelToEntity(&model)
	}
	
	return sessions, nil
}

// Update updates an existing session
func (r *sessionRepository) Update(ctx context.Context, session *entities.Session) error {
	model := &models.Session{
		ID:        session.ID,
		UserID:    session.UserID,
		Status:    string(session.Status),
		Metadata:  models.JSON(session.Metadata),
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
		ExpiresAt: session.ExpiresAt,
	}
	
	return r.db.WithContext(ctx).Save(model).Error
}

// Delete deletes a session by ID
func (r *sessionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Session{}, "id = ?", id).Error
}

// GetByStatus retrieves sessions by status
func (r *sessionRepository) GetByStatus(ctx context.Context, status entities.SessionStatus) ([]*entities.Session, error) {
	var models []models.Session
	err := r.db.WithContext(ctx).
		Where("status = ?", string(status)).
		Order("updated_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	
	sessions := make([]*entities.Session, len(models))
	for i, model := range models {
		sessions[i] = r.modelToEntity(&model)
	}
	
	return sessions, nil
}

// GetExpiredSessions retrieves sessions that have expired
func (r *sessionRepository) GetExpiredSessions(ctx context.Context) ([]*entities.Session, error) {
	var models []models.Session
	err := r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	
	sessions := make([]*entities.Session, len(models))
	for i, model := range models {
		sessions[i] = r.modelToEntity(&model)
	}
	
	return sessions, nil
}

// CleanupExpiredSessions removes expired sessions
func (r *sessionRepository) CleanupExpiredSessions(ctx context.Context) (int, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP").
		Delete(&models.Session{})
	
	return int(result.RowsAffected), result.Error
}

// CountByStatus returns the count of sessions by status
func (r *sessionRepository) CountByStatus(ctx context.Context, status entities.SessionStatus) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Session{}).
		Where("status = ?", string(status)).
		Count(&count).Error
	
	return int(count), err
}

// modelToEntity converts a database model to domain entity
func (r *sessionRepository) modelToEntity(model *models.Session) *entities.Session {
	return &entities.Session{
		ID:        model.ID,
		UserID:    model.UserID,
		Status:    entities.SessionStatus(model.Status),
		Metadata:  map[string]interface{}(model.Metadata),
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		ExpiresAt: model.ExpiresAt,
	}
}