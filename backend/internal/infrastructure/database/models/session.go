package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Session represents the session database model
type Session struct {
	ID        string    `gorm:"primaryKey;type:varchar(255)" json:"id"`
	UserID    string    `gorm:"type:varchar(255);index" json:"user_id"`
	Status    string    `gorm:"type:varchar(50);not null" json:"status"`
	Metadata  JSON      `gorm:"type:json" json:"metadata"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
	ExpiresAt *time.Time `gorm:"index" json:"expires_at"`
	
	// Relationships
	Messages          []Message          `gorm:"foreignKey:SessionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"messages,omitempty"`
	Requests          []Request          `gorm:"foreignKey:SessionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"requests,omitempty"`
	ProcessingContexts []ProcessingContext `gorm:"foreignKey:SessionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"processing_contexts,omitempty"`
}

// TableName specifies the table name for Session model
func (Session) TableName() string {
	return "sessions"
}

// BeforeCreate GORM hook called before creating a session
func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now().UTC()
	}
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = time.Now().UTC()
	}
	return nil
}

// BeforeUpdate GORM hook called before updating a session
func (s *Session) BeforeUpdate(tx *gorm.DB) error {
	s.UpdatedAt = time.Now().UTC()
	return nil
}

// JSON type for handling JSON fields in GORM
type JSON map[string]interface{}

// Value implements the driver.Valuer interface for JSON
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSON
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, j)
}