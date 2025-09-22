package interfaces

import "local-backend-server/internal/infrastructure/queue"

// MessagePublisher defines the interface for publishing messages to a queue
type MessagePublisher interface {
	PublishMessage(msg *queue.WorkflowMessage) error
	Close() error
}