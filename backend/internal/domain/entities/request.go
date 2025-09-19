package entities

import (
	"time"
)

// RequestStatus represents the status of a request
type RequestStatus string

const (
	RequestStatusPending    RequestStatus = "pending"
	RequestStatusProcessing RequestStatus = "processing"
	RequestStatusCompleted  RequestStatus = "completed"
	RequestStatusFailed     RequestStatus = "failed"
	RequestStatusCancelled  RequestStatus = "cancelled"
	RequestStatusTimeout    RequestStatus = "timeout"
)

// RequestType represents different types of requests
type RequestType string

const (
	RequestTypeCodeReview    RequestType = "code_review"
	RequestTypeIssueAnalysis RequestType = "issue_analysis"
	RequestTypeBugFix        RequestType = "bug_fix"
	RequestTypeFeature       RequestType = "feature"
)

// Request represents a workflow request entity
type Request struct {
	ID               string                 `json:"id"`
	SessionID        string                 `json:"session_id"`
	Type             RequestType            `json:"type"`
	Status           RequestStatus          `json:"status"`
	Input            map[string]interface{} `json:"input"`
	Output           map[string]interface{} `json:"output,omitempty"`
	Error            string                 `json:"error,omitempty"`
	ProcessingTimeMs int64                  `json:"processing_time_ms"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	StartedAt        *time.Time             `json:"started_at,omitempty"`
	CompletedAt      *time.Time             `json:"completed_at,omitempty"`
}

// NewRequest creates a new request entity
func NewRequest(sessionID string, requestType RequestType, input map[string]interface{}) *Request {
	now := time.Now()
	return &Request{
		ID:        generateID(),
		SessionID: sessionID,
		Type:      requestType,
		Status:    RequestStatusPending,
		Input:     input,
		Output:    make(map[string]interface{}),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Start marks the request as started
func (r *Request) Start() {
	now := time.Now()
	r.Status = RequestStatusProcessing
	r.StartedAt = &now
	r.UpdatedAt = now
}

// Complete marks the request as completed with output
func (r *Request) Complete(output map[string]interface{}) {
	now := time.Now()
	r.Status = RequestStatusCompleted
	r.Output = output
	r.CompletedAt = &now
	r.UpdatedAt = now
	
	if r.StartedAt != nil {
		r.ProcessingTimeMs = now.Sub(*r.StartedAt).Milliseconds()
	}
}

// Fail marks the request as failed with error
func (r *Request) Fail(errorMsg string) {
	now := time.Now()
	r.Status = RequestStatusFailed
	r.Error = errorMsg
	r.CompletedAt = &now
	r.UpdatedAt = now
	
	if r.StartedAt != nil {
		r.ProcessingTimeMs = now.Sub(*r.StartedAt).Milliseconds()
	}
}

// Cancel marks the request as cancelled
func (r *Request) Cancel() {
	now := time.Now()
	r.Status = RequestStatusCancelled
	r.CompletedAt = &now
	r.UpdatedAt = now
}

// Timeout marks the request as timed out
func (r *Request) Timeout() {
	now := time.Now()
	r.Status = RequestStatusTimeout
	r.Error = "Request processing timed out"
	r.CompletedAt = &now
	r.UpdatedAt = now
	
	if r.StartedAt != nil {
		r.ProcessingTimeMs = now.Sub(*r.StartedAt).Milliseconds()
	}
}

// IsCompleted checks if the request is in a completed state
func (r *Request) IsCompleted() bool {
	return r.Status == RequestStatusCompleted ||
		r.Status == RequestStatusFailed ||
		r.Status == RequestStatusCancelled ||
		r.Status == RequestStatusTimeout
}

// IsValid validates the request entity
func (r *Request) IsValid() bool {
	return r.ID != "" &&
		r.SessionID != "" &&
		r.Type != "" &&
		r.Status != "" &&
		r.Input != nil &&
		!r.CreatedAt.IsZero()
}