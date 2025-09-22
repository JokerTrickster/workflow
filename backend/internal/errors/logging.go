package errors

import (
	"context"
	"encoding/json"
	"log"
	"runtime"
	"time"
)

// LogLevel represents logging levels
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
	LogLevelFatal LogLevel = "FATAL"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Level     LogLevel               `json:"level"`
	Timestamp time.Time              `json:"timestamp"`
	Message   string                 `json:"message"`
	Error     *ErrorLogData          `json:"error,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	Source    SourceInfo             `json:"source"`
}

// ErrorLogData contains error-specific log information
type ErrorLogData struct {
	Code       ErrorCode                `json:"code"`
	Message    string                   `json:"message"`
	Details    string                   `json:"details,omitempty"`
	Severity   ErrorSeverity            `json:"severity"`
	Retryable  bool                     `json:"retryable"`
	Context    map[string]interface{}   `json:"context,omitempty"`
	StackTrace string                   `json:"stack_trace,omitempty"`
	Cause      string                   `json:"cause,omitempty"`
}

// SourceInfo contains information about where the log was generated
type SourceInfo struct {
	File     string `json:"file"`
	Function string `json:"function"`
	Line     int    `json:"line"`
}

// Logger provides structured logging capabilities
type Logger struct {
	minLevel  LogLevel
	requestID string
	traceID   string
	context   map[string]interface{}
}

// NewLogger creates a new structured logger
func NewLogger() *Logger {
	return &Logger{
		minLevel: LogLevelInfo,
		context:  make(map[string]interface{}),
	}
}

// WithLevel sets the minimum log level
func (l *Logger) WithLevel(level LogLevel) *Logger {
	l.minLevel = level
	return l
}

// WithRequestID adds request ID to all log entries
func (l *Logger) WithRequestID(requestID string) *Logger {
	l.requestID = requestID
	return l
}

// WithTraceID adds trace ID to all log entries
func (l *Logger) WithTraceID(traceID string) *Logger {
	l.traceID = traceID
	return l
}

// WithContext adds context that will be included in all log entries
func (l *Logger) WithContext(key string, value interface{}) *Logger {
	l.context[key] = value
	return l
}

// Error logs an error with full context
func (l *Logger) Error(err error, message string) {
	l.logError(LogLevelError, err, message, nil)
}

// ErrorWithContext logs an error with additional context
func (l *Logger) ErrorWithContext(err error, message string, context map[string]interface{}) {
	l.logError(LogLevelError, err, message, context)
}

// Fatal logs a fatal error and typically should be followed by os.Exit
func (l *Logger) Fatal(err error, message string) {
	l.logError(LogLevelFatal, err, message, nil)
}

// Warn logs a warning message
func (l *Logger) Warn(message string) {
	l.log(LogLevelWarn, message, nil, nil)
}

// WarnWithContext logs a warning with context
func (l *Logger) WarnWithContext(message string, context map[string]interface{}) {
	l.log(LogLevelWarn, message, context, nil)
}

// Info logs an informational message
func (l *Logger) Info(message string) {
	l.log(LogLevelInfo, message, nil, nil)
}

// InfoWithContext logs an info message with context
func (l *Logger) InfoWithContext(message string, context map[string]interface{}) {
	l.log(LogLevelInfo, message, context, nil)
}

// Debug logs a debug message
func (l *Logger) Debug(message string) {
	l.log(LogLevelDebug, message, nil, nil)
}

// DebugWithContext logs a debug message with context
func (l *Logger) DebugWithContext(message string, context map[string]interface{}) {
	l.log(LogLevelDebug, message, context, nil)
}

// logError logs an error with full error analysis
func (l *Logger) logError(level LogLevel, err error, message string, additionalContext map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	var errorData *ErrorLogData
	
	// Handle AppError specially
	if appErr, ok := err.(*AppError); ok {
		errorData = &ErrorLogData{
			Code:      appErr.Code,
			Message:   appErr.Message,
			Details:   appErr.Details,
			Severity:  appErr.Severity,
			Retryable: appErr.Retryable,
			Context:   appErr.Context,
		}
		
		if appErr.Cause != nil {
			errorData.Cause = appErr.Cause.Error()
		}
	} else if err != nil {
		// Handle regular errors
		errorData = &ErrorLogData{
			Message: err.Error(),
		}
	}

	// Add stack trace for errors
	if errorData != nil && level == LogLevelError {
		errorData.StackTrace = getStackTrace()
	}

	l.log(level, message, additionalContext, errorData)
}

// log is the main logging function
func (l *Logger) log(level LogLevel, message string, additionalContext map[string]interface{}, errorData *ErrorLogData) {
	if !l.shouldLog(level) {
		return
	}

	// Merge contexts
	context := make(map[string]interface{})
	for k, v := range l.context {
		context[k] = v
	}
	for k, v := range additionalContext {
		context[k] = v
	}

	entry := LogEntry{
		Level:     level,
		Timestamp: time.Now().UTC(),
		Message:   message,
		Error:     errorData,
		Context:   context,
		RequestID: l.requestID,
		TraceID:   l.traceID,
		Source:    getSourceInfo(),
	}

	// Convert to JSON and log
	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		log.Printf("%s: %s", level, message)
		return
	}

	log.Print(string(jsonBytes))
}

// shouldLog determines if a message should be logged based on level
func (l *Logger) shouldLog(level LogLevel) bool {
	levelOrder := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
		LogLevelFatal: 4,
	}

	return levelOrder[level] >= levelOrder[l.minLevel]
}

// getSourceInfo gets information about the caller
func getSourceInfo() SourceInfo {
	pc, file, line, ok := runtime.Caller(3) // Skip 3 frames to get actual caller
	if !ok {
		return SourceInfo{}
	}

	fn := runtime.FuncForPC(pc)
	functionName := ""
	if fn != nil {
		functionName = fn.Name()
	}

	return SourceInfo{
		File:     file,
		Function: functionName,
		Line:     line,
	}
}

// getStackTrace returns a formatted stack trace
func getStackTrace() string {
	stack := make([]byte, 4096)
	length := runtime.Stack(stack, false)
	return string(stack[:length])
}

// LogAppError is a convenience function to log an AppError
func LogAppError(appErr *AppError, message string) {
	logger := NewLogger()
	logger.Error(appErr, message)
}

// LogAppErrorWithContext logs an AppError with additional context
func LogAppErrorWithContext(appErr *AppError, message string, context map[string]interface{}) {
	logger := NewLogger()
	logger.ErrorWithContext(appErr, message, context)
}

// LogAppErrorForRequest logs an AppError for a specific request
func LogAppErrorForRequest(appErr *AppError, message string, requestID string) {
	logger := NewLogger().WithRequestID(requestID)
	logger.Error(appErr, message)
}

// ContextKey represents keys for context values
type ContextKey string

const (
	RequestIDKey ContextKey = "request_id"
	TraceIDKey   ContextKey = "trace_id"
)

// LoggerFromContext creates a logger from context values
func LoggerFromContext(ctx context.Context) *Logger {
	logger := NewLogger()
	
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		logger = logger.WithRequestID(requestID)
	}
	
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		logger = logger.WithTraceID(traceID)
	}
	
	return logger
}