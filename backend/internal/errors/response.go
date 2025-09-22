package errors

import (
	"time"
)

// ErrorResponse represents the standardized error response format
type ErrorResponse struct {
	Error      ErrorDetail   `json:"error"`
	RequestID  string        `json:"request_id,omitempty"`
	Timestamp  time.Time     `json:"timestamp"`
	Path       string        `json:"path,omitempty"`
	Method     string        `json:"method,omitempty"`
	TraceID    string        `json:"trace_id,omitempty"`
}

// ErrorDetail contains the main error information
type ErrorDetail struct {
	Code       ErrorCode                `json:"code"`
	Message    string                   `json:"message"`
	Details    string                   `json:"details,omitempty"`
	Severity   ErrorSeverity            `json:"severity"`
	Retryable  bool                     `json:"retryable"`
	Context    map[string]interface{}   `json:"context,omitempty"`
	Validation []ValidationError        `json:"validation,omitempty"`
}

// ValidationError represents individual field validation errors
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
	Code    string      `json:"code,omitempty"`
}

// NewErrorResponse creates a standardized error response from an AppError
func NewErrorResponse(appErr *AppError) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDetail{
			Code:      appErr.Code,
			Message:   appErr.Message,
			Details:   appErr.Details,
			Severity:  appErr.Severity,
			Retryable: appErr.Retryable,
			Context:   appErr.Context,
		},
		RequestID: appErr.RequestID,
		Timestamp: appErr.Timestamp,
	}
}

// NewValidationErrorResponse creates an error response for validation failures
func NewValidationErrorResponse(validationErrors []ValidationError) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDetail{
			Code:       ErrCodeValidationFailed,
			Message:    "Request validation failed",
			Severity:   SeverityLow,
			Retryable:  false,
			Validation: validationErrors,
		},
		Timestamp: time.Now().UTC(),
	}
}

// WithRequestInfo adds request context to the error response
func (er *ErrorResponse) WithRequestInfo(requestID, path, method string) *ErrorResponse {
	er.RequestID = requestID
	er.Path = path
	er.Method = method
	return er
}

// WithTraceID adds trace ID for distributed tracing
func (er *ErrorResponse) WithTraceID(traceID string) *ErrorResponse {
	er.TraceID = traceID
	return er
}

// AddValidationError adds a validation error to the response
func (er *ErrorResponse) AddValidationError(field, message string, value interface{}) {
	if er.Error.Validation == nil {
		er.Error.Validation = make([]ValidationError, 0)
	}
	
	er.Error.Validation = append(er.Error.Validation, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}