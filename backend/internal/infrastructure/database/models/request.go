package models

import (
	"time"

	"gorm.io/gorm"
)

// Request represents the request database model
type Request struct {
	ID               string     `gorm:"primaryKey;type:varchar(255)" json:"id"`
	SessionID        string     `gorm:"type:varchar(255);not null;index" json:"session_id"`
	Type             string     `gorm:"type:varchar(50);not null" json:"type"`
	Status           string     `gorm:"type:varchar(50);not null;index" json:"status"`
	Input            JSON       `gorm:"type:json;not null" json:"input"`
	Output           JSON       `gorm:"type:json" json:"output"`
	Error            string     `gorm:"type:text" json:"error"`
	ProcessingTimeMs int64      `gorm:"default:0" json:"processing_time_ms"`
	CreatedAt        time.Time  `gorm:"not null;index" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"not null" json:"updated_at"`
	StartedAt        *time.Time `gorm:"index" json:"started_at"`
	CompletedAt      *time.Time `gorm:"index" json:"completed_at"`
	
	// Relationships
	Session           *Session           `gorm:"foreignKey:SessionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"session,omitempty"`
	ProcessingContext *ProcessingContext `gorm:"foreignKey:RequestID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"processing_context,omitempty"`
}

// TableName specifies the table name for Request model
func (Request) TableName() string {
	return "requests"
}

// BeforeCreate GORM hook called before creating a request
func (r *Request) BeforeCreate(tx *gorm.DB) error {
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}
	if r.UpdatedAt.IsZero() {
		r.UpdatedAt = time.Now().UTC()
	}
	return nil
}

// BeforeUpdate GORM hook called before updating a request
func (r *Request) BeforeUpdate(tx *gorm.DB) error {
	r.UpdatedAt = time.Now().UTC()
	return nil
}