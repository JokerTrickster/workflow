package models

import (
	"time"

	"gorm.io/gorm"
)

// Message represents the message database model
type Message struct {
	ID        string    `gorm:"primaryKey;type:varchar(255)" json:"id"`
	SessionID string    `gorm:"type:varchar(255);not null;index" json:"session_id"`
	Type      string    `gorm:"type:varchar(50);not null" json:"type"`
	Role      string    `gorm:"type:varchar(50);not null" json:"role"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Metadata  JSON      `gorm:"type:json" json:"metadata"`
	CreatedAt time.Time `gorm:"not null;index" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
	
	// Relationships
	Session *Session `gorm:"foreignKey:SessionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"session,omitempty"`
}

// TableName specifies the table name for Message model
func (Message) TableName() string {
	return "messages"
}

// BeforeCreate GORM hook called before creating a message
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now().UTC()
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now().UTC()
	}
	return nil
}

// BeforeUpdate GORM hook called before updating a message
func (m *Message) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = time.Now().UTC()
	return nil
}