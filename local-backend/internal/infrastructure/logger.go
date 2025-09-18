package infrastructure

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Logger represents a structured logger
type Logger struct {
	level  LogLevel
	format string
}

// NewLogger creates a new logger with the specified configuration
func NewLogger(config *LoggingConfig) *Logger {
	level := parseLogLevel(config.Level)
	
	// Configure log output
	if config.Output == "file" {
		// For file output, would need to implement file rotation
		// For now, just use stdout
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(os.Stdout)
	}

	// Set log flags based on format
	if config.Format == "json" {
		log.SetFlags(0) // No standard prefix for JSON
	} else {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	return &Logger{
		level:  level,
		format: config.Format,
	}
}

// parseLogLevel converts string to LogLevel
func parseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warn", "warning":
		return LogLevelWarn
	case "error":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}

// Debug logs debug messages
func (l *Logger) Debug(msg string, fields ...interface{}) {
	if l.level <= LogLevelDebug {
		l.log("DEBUG", msg, fields...)
	}
}

// Info logs info messages
func (l *Logger) Info(msg string, fields ...interface{}) {
	if l.level <= LogLevelInfo {
		l.log("INFO", msg, fields...)
	}
}

// Warn logs warning messages
func (l *Logger) Warn(msg string, fields ...interface{}) {
	if l.level <= LogLevelWarn {
		l.log("WARN", msg, fields...)
	}
}

// Error logs error messages
func (l *Logger) Error(msg string, fields ...interface{}) {
	if l.level <= LogLevelError {
		l.log("ERROR", msg, fields...)
	}
}

// log formats and outputs the log message
func (l *Logger) log(level, msg string, fields ...interface{}) {
	if l.format == "json" {
		l.logJSON(level, msg, fields...)
	} else {
		l.logText(level, msg, fields...)
	}
}

// logJSON outputs JSON formatted log
func (l *Logger) logJSON(level, msg string, fields ...interface{}) {
	// Simple JSON formatting - in production would use proper JSON encoder
	fieldsStr := ""
	if len(fields) > 0 {
		fieldsStr = fmt.Sprintf(", \"fields\": %v", fields)
	}
	
	log.Printf(`{"level": "%s", "message": "%s", "timestamp": "%s"%s}`, 
		level, msg, "timestamp", fieldsStr)
}

// logText outputs text formatted log
func (l *Logger) logText(level, msg string, fields ...interface{}) {
	if len(fields) > 0 {
		log.Printf("[%s] %s - %v", level, msg, fields)
	} else {
		log.Printf("[%s] %s", level, msg)
	}
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	// In a real implementation, this would create a new logger instance
	// with the fields attached for all subsequent log calls
	return l
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() LogLevel {
	return l.level
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}