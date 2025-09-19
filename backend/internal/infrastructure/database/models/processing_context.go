package models

import (
	"time"

	"gorm.io/gorm"
)

// ProcessingContext represents the processing context database model
type ProcessingContext struct {
	ID           string    `gorm:"primaryKey;type:varchar(255)" json:"id"`
	RequestID    string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"request_id"`
	SessionID    string    `gorm:"type:varchar(255);not null;index" json:"session_id"`
	SystemPrompt string    `gorm:"type:text" json:"system_prompt"`
	Metadata     JSON      `gorm:"type:json" json:"metadata"`
	TokenUsage   JSON      `gorm:"type:json" json:"token_usage"`
	CreatedAt    time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null" json:"updated_at"`
	
	// Relationships
	Request *Request `gorm:"foreignKey:RequestID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"request,omitempty"`
	Session *Session `gorm:"foreignKey:SessionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"session,omitempty"`
}

// TableName specifies the table name for ProcessingContext model
func (ProcessingContext) TableName() string {
	return "processing_contexts"
}

// BeforeCreate GORM hook called before creating a processing context
func (pc *ProcessingContext) BeforeCreate(tx *gorm.DB) error {
	if pc.CreatedAt.IsZero() {
		pc.CreatedAt = time.Now().UTC()
	}
	if pc.UpdatedAt.IsZero() {
		pc.UpdatedAt = time.Now().UTC()
	}
	return nil
}

// BeforeUpdate GORM hook called before updating a processing context
func (pc *ProcessingContext) BeforeUpdate(tx *gorm.DB) error {
	pc.UpdatedAt = time.Now().UTC()
	return nil
}