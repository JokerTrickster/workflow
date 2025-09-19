package services

import (
	"context"
	"errors"
	"local-backend-server/internal/domain/entities"
	"local-backend-server/internal/domain/repositories"
	"time"
)

// RequestService handles business logic for requests
type RequestService struct {
	requestRepo repositories.RequestRepository
	sessionRepo repositories.SessionRepository
}

// NewRequestService creates a new RequestService
func NewRequestService(requestRepo repositories.RequestRepository, sessionRepo repositories.SessionRepository) *RequestService {
	return &RequestService{
		requestRepo: requestRepo,
		sessionRepo: sessionRepo,
	}
}

// CreateRequest creates a new request with validation
func (s *RequestService) CreateRequest(ctx context.Context, sessionID string, requestType entities.RequestType, input map[string]interface{}) (*entities.Request, error) {
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
	
	// Validate input
	if err := s.validateRequestInput(requestType, input); err != nil {
		return nil, err
	}
	
	// Create request
	request := entities.NewRequest(sessionID, requestType, input)
	
	// Validate request
	if !request.IsValid() {
		return nil, errors.New("invalid request data")
	}
	
	// Save request
	if err := s.requestRepo.Create(ctx, request); err != nil {
		return nil, err
	}
	
	return request, nil
}

// StartRequest marks a request as started
func (s *RequestService) StartRequest(ctx context.Context, requestID string) error {
	request, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}
	
	if request == nil {
		return errors.New("request not found")
	}
	
	if request.Status != entities.RequestStatusPending {
		return errors.New("request is not in pending status")
	}
	
	request.Start()
	
	return s.requestRepo.Update(ctx, request)
}

// CompleteRequest marks a request as completed
func (s *RequestService) CompleteRequest(ctx context.Context, requestID string, output map[string]interface{}) error {
	request, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}
	
	if request == nil {
		return errors.New("request not found")
	}
	
	if request.Status != entities.RequestStatusProcessing {
		return errors.New("request is not in processing status")
	}
	
	request.Complete(output)
	
	return s.requestRepo.Update(ctx, request)
}

// FailRequest marks a request as failed
func (s *RequestService) FailRequest(ctx context.Context, requestID string, errorMsg string) error {
	request, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}
	
	if request == nil {
		return errors.New("request not found")
	}
	
	if request.IsCompleted() {
		return errors.New("request is already completed")
	}
	
	request.Fail(errorMsg)
	
	return s.requestRepo.Update(ctx, request)
}

// CancelRequest marks a request as cancelled
func (s *RequestService) CancelRequest(ctx context.Context, requestID string) error {
	request, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}
	
	if request == nil {
		return errors.New("request not found")
	}
	
	if request.IsCompleted() {
		return errors.New("request is already completed")
	}
	
	request.Cancel()
	
	return s.requestRepo.Update(ctx, request)
}

// GetPendingRequests retrieves all pending requests
func (s *RequestService) GetPendingRequests(ctx context.Context) ([]*entities.Request, error) {
	return s.requestRepo.GetPendingRequests(ctx)
}

// CheckForTimeouts checks for requests that may have timed out
func (s *RequestService) CheckForTimeouts(ctx context.Context, timeout time.Duration) ([]*entities.Request, error) {
	return s.requestRepo.GetRequestsWithTimeout(ctx, timeout)
}

// TimeoutRequests marks requests as timed out
func (s *RequestService) TimeoutRequests(ctx context.Context, requests []*entities.Request) error {
	for _, request := range requests {
		request.Timeout()
		if err := s.requestRepo.Update(ctx, request); err != nil {
			return err
		}
	}
	return nil
}

// validateRequestInput validates request input based on type
func (s *RequestService) validateRequestInput(requestType entities.RequestType, input map[string]interface{}) error {
	if input == nil {
		return errors.New("request input cannot be nil")
	}
	
	switch requestType {
	case entities.RequestTypeCodeReview:
		if _, ok := input["code"]; !ok {
			return errors.New("code review request must include 'code' field")
		}
	case entities.RequestTypeIssueAnalysis:
		if _, ok := input["issue"]; !ok {
			return errors.New("issue analysis request must include 'issue' field")
		}
	case entities.RequestTypeBugFix:
		if _, ok := input["bug_description"]; !ok {
			return errors.New("bug fix request must include 'bug_description' field")
		}
	case entities.RequestTypeFeature:
		if _, ok := input["feature_description"]; !ok {
			return errors.New("feature request must include 'feature_description' field")
		}
	}
	
	return nil
}