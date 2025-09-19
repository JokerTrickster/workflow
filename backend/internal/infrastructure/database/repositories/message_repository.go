package repositories

import (
	"context"

	"gorm.io/gorm"

	"local-backend-server/internal/domain/entities"
	"local-backend-server/internal/domain/repositories"
	"local-backend-server/internal/infrastructure/database/models"
)

// messageRepository implements the MessageRepository interface
type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *gorm.DB) repositories.MessageRepository {
	return &messageRepository{db: db}
}

// Create creates a new message
func (r *messageRepository) Create(ctx context.Context, message *entities.Message) error {
	model := &models.Message{
		ID:        message.ID,
		SessionID: message.SessionID,
		Type:      string(message.Type),
		Role:      string(message.Role),
		Content:   message.Content,
		Metadata:  models.JSON(message.Metadata),
		CreatedAt: message.CreatedAt,
		UpdatedAt: message.UpdatedAt,
	}
	
	return r.db.WithContext(ctx).Create(model).Error
}

// GetByID retrieves a message by its ID
func (r *messageRepository) GetByID(ctx context.Context, id string) (*entities.Message, error) {
	var model models.Message
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return r.modelToEntity(&model), nil
}

// GetBySessionID retrieves all messages for a session
func (r *messageRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*entities.Message, error) {
	var models []models.Message
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	
	messages := make([]*entities.Message, len(models))
	for i, model := range models {
		messages[i] = r.modelToEntity(&model)
	}
	
	return messages, nil
}

// GetBySessionIDWithPagination retrieves messages with pagination
func (r *messageRepository) GetBySessionIDWithPagination(ctx context.Context, sessionID string, offset, limit int) ([]*entities.Message, error) {
	var models []models.Message
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	
	messages := make([]*entities.Message, len(models))
	for i, model := range models {
		messages[i] = r.modelToEntity(&model)
	}
	
	return messages, nil
}

// Update updates an existing message
func (r *messageRepository) Update(ctx context.Context, message *entities.Message) error {
	model := &models.Message{
		ID:        message.ID,
		SessionID: message.SessionID,
		Type:      string(message.Type),
		Role:      string(message.Role),
		Content:   message.Content,
		Metadata:  models.JSON(message.Metadata),
		CreatedAt: message.CreatedAt,
		UpdatedAt: message.UpdatedAt,
	}
	
	return r.db.WithContext(ctx).Save(model).Error
}

// Delete deletes a message by ID
func (r *messageRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Message{}, "id = ?", id).Error
}

// GetByTypeAndSessionID retrieves messages by type and session
func (r *messageRepository) GetByTypeAndSessionID(ctx context.Context, messageType entities.MessageType, sessionID string) ([]*entities.Message, error) {
	var models []models.Message
	err := r.db.WithContext(ctx).
		Where("type = ? AND session_id = ?", string(messageType), sessionID).
		Order("created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	
	messages := make([]*entities.Message, len(models))
	for i, model := range models {
		messages[i] = r.modelToEntity(&model)
	}
	
	return messages, nil
}

// CountBySessionID returns the count of messages in a session
func (r *messageRepository) CountBySessionID(ctx context.Context, sessionID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("session_id = ?", sessionID).
		Count(&count).Error
	
	return int(count), err
}

// GetLatestBySessionID retrieves the most recent message in a session
func (r *messageRepository) GetLatestBySessionID(ctx context.Context, sessionID string) (*entities.Message, error) {
	var model models.Message
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at DESC").
		First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return r.modelToEntity(&model), nil
}

// modelToEntity converts a database model to domain entity
func (r *messageRepository) modelToEntity(model *models.Message) *entities.Message {
	return &entities.Message{
		ID:        model.ID,
		SessionID: model.SessionID,
		Type:      entities.MessageType(model.Type),
		Role:      entities.MessageRole(model.Role),
		Content:   model.Content,
		Metadata:  entities.Metadata(model.Metadata),
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}