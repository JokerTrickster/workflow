package services

import (
	"context"
	"errors"
	"local-backend-server/internal/domain/entities"
	"local-backend-server/internal/domain/repositories"
)

// MessageService handles business logic for messages
type MessageService struct {
	messageRepo repositories.MessageRepository
	sessionRepo repositories.SessionRepository
}

// NewMessageService creates a new MessageService
func NewMessageService(messageRepo repositories.MessageRepository, sessionRepo repositories.SessionRepository) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		sessionRepo: sessionRepo,
	}
}

// CreateMessage creates a new message with validation
func (s *MessageService) CreateMessage(ctx context.Context, sessionID string, messageType entities.MessageType, role entities.MessageRole, content string) (*entities.Message, error) {
	// Validate session exists and is active
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	
	if session == nil {
		return nil, errors.New("session not found")
	}
	
	if session.Status != entities.SessionStatusActive {
		return nil, errors.New("session is not active")
	}
	
	// Create message
	message := entities.NewMessage(sessionID, messageType, role, content)
	
	// Validate message
	if !message.IsValid() {
		return nil, errors.New("invalid message data")
	}
	
	// Save message
	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, err
	}
	
	return message, nil
}

// GetConversationHistory retrieves conversation history for a session
func (s *MessageService) GetConversationHistory(ctx context.Context, sessionID string, limit int) ([]*entities.Message, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}
	
	return s.messageRepo.GetBySessionIDWithPagination(ctx, sessionID, 0, limit)
}

// AddMetadataToMessage adds metadata to an existing message
func (s *MessageService) AddMetadataToMessage(ctx context.Context, messageID string, key string, value interface{}) error {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return err
	}
	
	if message == nil {
		return errors.New("message not found")
	}
	
	message.AddMetadata(key, value)
	
	return s.messageRepo.Update(ctx, message)
}

// GetMessagesByType retrieves messages of a specific type for a session
func (s *MessageService) GetMessagesByType(ctx context.Context, sessionID string, messageType entities.MessageType) ([]*entities.Message, error) {
	return s.messageRepo.GetByTypeAndSessionID(ctx, messageType, sessionID)
}

// ValidateMessageContent validates message content according to business rules
func (s *MessageService) ValidateMessageContent(content string) error {
	if content == "" {
		return errors.New("message content cannot be empty")
	}
	
	if len(content) > 10000 {
		return errors.New("message content is too long")
	}
	
	return nil
}