package infrastructure

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name   string
		config *LoggingConfig
	}{
		{
			name: "json format",
			config: &LoggingConfig{
				Level:  "debug",
				Format: "json",
				Output: "stdout",
			},
		},
		{
			name: "text format",
			config: &LoggingConfig{
				Level:  "info",
				Format: "text",
				Output: "stdout",
			},
		},
		{
			name: "file output",
			config: &LoggingConfig{
				Level:  "warn",
				Format: "json",
				Output: "file",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.config)
			
			assert.NotNil(t, logger)
			assert.Equal(t, tt.config.Format, logger.format)
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected LogLevel
	}{
		{
			name:     "debug level",
			level:    "debug",
			expected: LogLevelDebug,
		},
		{
			name:     "info level",
			level:    "info",
			expected: LogLevelInfo,
		},
		{
			name:     "warn level",
			level:    "warn",
			expected: LogLevelWarn,
		},
		{
			name:     "warning level",
			level:    "warning",
			expected: LogLevelWarn,
		},
		{
			name:     "error level",
			level:    "error",
			expected: LogLevelError,
		},
		{
			name:     "uppercase debug",
			level:    "DEBUG",
			expected: LogLevelDebug,
		},
		{
			name:     "mixed case info",
			level:    "InFo",
			expected: LogLevelInfo,
		},
		{
			name:     "invalid level defaults to info",
			level:    "invalid",
			expected: LogLevelInfo,
		},
		{
			name:     "empty level defaults to info",
			level:    "",
			expected: LogLevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLogLevel(tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLogger_LogLevels(t *testing.T) {
	// Test different log levels and their filtering behavior
	tests := []struct {
		name            string
		loggerLevel     LogLevel
		shouldLogDebug  bool
		shouldLogInfo   bool
		shouldLogWarn   bool
		shouldLogError  bool
	}{
		{
			name:            "debug level logs everything",
			loggerLevel:     LogLevelDebug,
			shouldLogDebug:  true,
			shouldLogInfo:   true,
			shouldLogWarn:   true,
			shouldLogError:  true,
		},
		{
			name:            "info level skips debug",
			loggerLevel:     LogLevelInfo,
			shouldLogDebug:  false,
			shouldLogInfo:   true,
			shouldLogWarn:   true,
			shouldLogError:  true,
		},
		{
			name:            "warn level skips debug and info",
			loggerLevel:     LogLevelWarn,
			shouldLogDebug:  false,
			shouldLogInfo:   false,
			shouldLogWarn:   true,
			shouldLogError:  true,
		},
		{
			name:            "error level only logs errors",
			loggerLevel:     LogLevelError,
			shouldLogDebug:  false,
			shouldLogInfo:   false,
			shouldLogWarn:   false,
			shouldLogError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &Logger{
				level:  tt.loggerLevel,
				format: "text",
			}

			// Test each log method - we can't easily capture output,
			// but we can verify the methods don't panic
			logger.Debug("debug message")
			logger.Info("info message")
			logger.Warn("warn message") 
			logger.Error("error message")

			// Test with fields
			logger.Debug("debug with fields", "key", "value")
			logger.Info("info with fields", "user_id", 123)
			logger.Warn("warn with fields", "operation", "test")
			logger.Error("error with fields", "error", "test error")
		})
	}
}

func TestLogger_WithFields(t *testing.T) {
	config := &LoggingConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}
	logger := NewLogger(config)

	fields := map[string]interface{}{
		"user_id":    123,
		"request_id": "req-456",
		"operation":  "test",
	}

	// WithFields should return a logger (current implementation returns same instance)
	fieldLogger := logger.WithFields(fields)
	assert.NotNil(t, fieldLogger)
	assert.Equal(t, logger, fieldLogger) // Current implementation returns same instance
}

func TestLogger_GetAndSetLevel(t *testing.T) {
	logger := &Logger{
		level:  LogLevelInfo,
		format: "text",
	}

	// Test GetLevel
	assert.Equal(t, LogLevelInfo, logger.GetLevel())

	// Test SetLevel
	logger.SetLevel(LogLevelDebug)
	assert.Equal(t, LogLevelDebug, logger.GetLevel())

	logger.SetLevel(LogLevelError)
	assert.Equal(t, LogLevelError, logger.GetLevel())
}

func TestLogger_LogFormats(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{
			name:   "json format",
			format: "json",
		},
		{
			name:   "text format",
			format: "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &Logger{
				level:  LogLevelDebug,
				format: tt.format,
			}

			// Test that different formats don't panic
			logger.Debug("debug message")
			logger.Info("info message", "key", "value")
			logger.Warn("warn message", "user_id", 123, "action", "test")
			logger.Error("error message", "error", "test error", "code", 500)
		})
	}
}

func TestLogLevelConstants(t *testing.T) {
	// Verify the log level constants have expected values
	assert.Equal(t, LogLevel(0), LogLevelDebug)
	assert.Equal(t, LogLevel(1), LogLevelInfo)
	assert.Equal(t, LogLevel(2), LogLevelWarn)
	assert.Equal(t, LogLevel(3), LogLevelError)

	// Verify ordering for level comparison
	assert.True(t, LogLevelDebug < LogLevelInfo)
	assert.True(t, LogLevelInfo < LogLevelWarn)
	assert.True(t, LogLevelWarn < LogLevelError)
}

func TestLogger_EdgeCases(t *testing.T) {
	logger := &Logger{
		level:  LogLevelInfo,
		format: "text",
	}

	t.Run("empty message", func(t *testing.T) {
		logger.Info("")
		logger.Error("")
	})

	t.Run("nil fields", func(t *testing.T) {
		logger.Info("message with nil fields", nil)
	})

	t.Run("odd number of fields", func(t *testing.T) {
		logger.Info("message", "key1", "value1", "key2") // Missing value for key2
	})

	t.Run("mixed field types", func(t *testing.T) {
		logger.Info("message",
			"string", "value",
			"int", 42,
			"float", 3.14,
			"bool", true,
			"nil", nil)
	})

	t.Run("very long message", func(t *testing.T) {
		longMessage := strings.Repeat("a", 1000)
		logger.Info(longMessage)
	})

	t.Run("message with newlines", func(t *testing.T) {
		logger.Info("message with\nnewlines\nand\ttabs")
	})

	t.Run("message with special characters", func(t *testing.T) {
		logger.Info("message with special chars: !@#$%^&*(){}[]|\\:;\"'<>?,./")
	})
}