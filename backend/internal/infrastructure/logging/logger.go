package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	"local-backend-server/internal/infrastructure/config"
)

// Logger wraps logrus with structured logging capabilities
type Logger struct {
	*logrus.Logger
	config *config.LoggingConfig
}

// Fields represents structured log fields
type Fields map[string]interface{}

// NewLogger creates a new structured logger based on configuration
func NewLogger(cfg *config.LoggingConfig) (*Logger, error) {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level %s: %w", cfg.Level, err)
	}
	logger.SetLevel(level)

	// Set output destination
	switch cfg.Output {
	case "file":
		if err := setupFileOutput(logger, cfg); err != nil {
			return nil, fmt.Errorf("failed to setup file output: %w", err)
		}
	case "stdout":
		logger.SetOutput(os.Stdout)
	default:
		logger.SetOutput(os.Stdout)
	}

	// Set formatter
	switch cfg.Format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "caller",
			},
		})
	case "console":
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
			ForceColors:     true,
		})
	default:
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	// Enable caller info if configured
	logger.SetReportCaller(cfg.EnableCaller)

	return &Logger{
		Logger: logger,
		config: cfg,
	}, nil
}

// setupFileOutput configures file-based logging with rotation
func setupFileOutput(logger *logrus.Logger, cfg *config.LoggingConfig) error {
	// Ensure log directory exists
	logDir := filepath.Dir(cfg.File)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Setup log rotation
	lumberjackLogger := &lumberjack.Logger{
		Filename:   cfg.File,
		MaxSize:    cfg.MaxSize,    // MB
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,     // days
		Compress:   true,
	}

	logger.SetOutput(lumberjackLogger)
	return nil
}

// WithFields adds structured fields to log entry
func (l *Logger) WithFields(fields Fields) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields(fields))
}

// WithField adds a single field to log entry
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithField(key, value)
}

// WithError adds error field to log entry
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

// WithRequestID adds request ID field to log entry
func (l *Logger) WithRequestID(requestID string) *logrus.Entry {
	return l.Logger.WithField("request_id", requestID)
}

// WithSessionID adds session ID field to log entry
func (l *Logger) WithSessionID(sessionID string) *logrus.Entry {
	return l.Logger.WithField("session_id", sessionID)
}

// WithComponent adds component field to log entry
func (l *Logger) WithComponent(component string) *logrus.Entry {
	return l.Logger.WithField("component", component)
}

// WithDuration adds duration field to log entry
func (l *Logger) WithDuration(duration time.Duration) *logrus.Entry {
	return l.Logger.WithField("duration_ms", duration.Milliseconds())
}

// Emergency logs at emergency level (system is unusable)
func (l *Logger) Emergency(args ...interface{}) {
	l.Logger.Fatal(args...)
}

// Alert logs at alert level (action must be taken immediately)
func (l *Logger) Alert(args ...interface{}) {
	l.Logger.Error(args...)
}

// Critical logs at critical level (critical conditions)
func (l *Logger) Critical(args ...interface{}) {
	l.Logger.Error(args...)
}

// Notice logs at notice level (normal but significant condition)
func (l *Logger) Notice(args ...interface{}) {
	l.Logger.Info(args...)
}

// RequestLogger creates a logger with request context
func (l *Logger) RequestLogger(requestID, method, path string) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"method":     method,
		"path":       path,
		"component":  "http",
	})
}

// DatabaseLogger creates a logger with database context
func (l *Logger) DatabaseLogger(operation string) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields{
		"component": "database",
		"operation": operation,
	})
}

// QueueLogger creates a logger with queue context
func (l *Logger) QueueLogger(queueName string) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields{
		"component":  "queue",
		"queue_name": queueName,
	})
}

// ClaudeLogger creates a logger with Claude API context
func (l *Logger) ClaudeLogger(operation string) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields{
		"component": "claude",
		"operation": operation,
	})
}

// HealthLogger creates a logger with health check context
func (l *Logger) HealthLogger(service string) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields{
		"component": "health",
		"service":   service,
	})
}

// MetricsLogger creates a logger with metrics context
func (l *Logger) MetricsLogger(metric string) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields{
		"component": "metrics",
		"metric":    metric,
	})
}

// Close closes the logger and flushes any remaining logs
func (l *Logger) Close() error {
	// If using file output with lumberjack, close it
	if l.config.Output == "file" {
		if closer, ok := l.Logger.Out.(*lumberjack.Logger); ok {
			return closer.Close()
		}
	}
	return nil
}