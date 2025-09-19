package logging

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"local-backend-server/internal/infrastructure/config"
)

// LoggerTestSuite defines the test suite for Logger
type LoggerTestSuite struct {
	suite.Suite
	tempDir string
}

// SetupSuite runs before all tests in the suite
func (suite *LoggerTestSuite) SetupSuite() {
	// Create temporary directory for test logs
	tempDir, err := os.MkdirTemp("", "logger_test")
	suite.Require().NoError(err)
	suite.tempDir = tempDir
}

// TearDownSuite runs after all tests in the suite
func (suite *LoggerTestSuite) TearDownSuite() {
	// Clean up temporary directory
	os.RemoveAll(suite.tempDir)
}

// TestLoggerTestSuite runs the test suite
func TestLoggerTestSuite(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}

// TestNewLogger_ConsoleOutput tests creating logger with console output
func (suite *LoggerTestSuite) TestNewLogger_ConsoleOutput() {
	cfg := &config.LoggingConfig{
		Level:            "debug",
		Format:           "console",
		Output:           "stdout",
		EnableCaller:     true,
		EnableStacktrace: false,
	}
	
	logger, err := NewLogger(cfg)
	
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.NotNil(logger)
	assert.Equal(logrus.DebugLevel, logger.GetLevel())
	assert.True(logger.ReportCaller)
	assert.Equal(cfg, logger.config)
	
	// Test that it's using TextFormatter
	_, ok := logger.Formatter.(*logrus.TextFormatter)
	assert.True(ok)
}

// TestNewLogger_JSONOutput tests creating logger with JSON output
func (suite *LoggerTestSuite) TestNewLogger_JSONOutput() {
	cfg := &config.LoggingConfig{
		Level:            "info",
		Format:           "json",
		Output:           "stdout",
		EnableCaller:     false,
		EnableStacktrace: true,
	}
	
	logger, err := NewLogger(cfg)
	
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.NotNil(logger)
	assert.Equal(logrus.InfoLevel, logger.GetLevel())
	assert.False(logger.ReportCaller)
	
	// Test that it's using JSONFormatter
	jsonFormatter, ok := logger.Formatter.(*logrus.JSONFormatter)
	assert.True(ok)
	assert.Equal(time.RFC3339, jsonFormatter.TimestampFormat)
	assert.Equal("timestamp", jsonFormatter.FieldMap[logrus.FieldKeyTime])
	assert.Equal("level", jsonFormatter.FieldMap[logrus.FieldKeyLevel])
	assert.Equal("message", jsonFormatter.FieldMap[logrus.FieldKeyMsg])
	assert.Equal("caller", jsonFormatter.FieldMap[logrus.FieldKeyFunc])
}

// TestNewLogger_FileOutput tests creating logger with file output
func (suite *LoggerTestSuite) TestNewLogger_FileOutput() {
	logFile := filepath.Join(suite.tempDir, "test.log")
	cfg := &config.LoggingConfig{
		Level:      "warn",
		Format:     "json",
		Output:     "file",
		File:       logFile,
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
	}
	
	logger, err := NewLogger(cfg)
	
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.NotNil(logger)
	assert.Equal(logrus.WarnLevel, logger.GetLevel())
	
	// Test that log file is created when we write to it
	logger.Warn("Test log message")
	assert.FileExists(logFile)
	
	// Clean up
	logger.Close()
}

// TestNewLogger_InvalidLevel tests creating logger with invalid log level
func (suite *LoggerTestSuite) TestNewLogger_InvalidLevel() {
	cfg := &config.LoggingConfig{
		Level:  "invalid",
		Format: "console",
		Output: "stdout",
	}
	
	logger, err := NewLogger(cfg)
	
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Nil(logger)
	assert.Contains(err.Error(), "invalid log level")
}

// TestNewLogger_FileOutputDirectoryCreation tests automatic directory creation
func (suite *LoggerTestSuite) TestNewLogger_FileOutputDirectoryCreation() {
	logDir := filepath.Join(suite.tempDir, "nested", "logs")
	logFile := filepath.Join(logDir, "app.log")
	cfg := &config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "file",
		File:   logFile,
	}
	
	logger, err := NewLogger(cfg)
	
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.NotNil(logger)
	
	// Test that nested directory is created
	logger.Info("Test message")
	assert.DirExists(logDir)
	assert.FileExists(logFile)
	
	// Clean up
	logger.Close()
}

// TestLoggerWithFields tests structured logging with fields
func (suite *LoggerTestSuite) TestLoggerWithFields() {
	var buf bytes.Buffer
	cfg := &config.LoggingConfig{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	}
	
	logger, err := NewLogger(cfg)
	assert.NoError(suite.T(), err)
	
	// Redirect output to buffer for testing
	logger.SetOutput(&buf)
	
	// Test WithFields
	fields := Fields{
		"user_id":    "user-123",
		"session_id": "session-456",
		"operation":  "create_user",
	}
	
	logger.WithFields(fields).Info("User creation started")
	
	// Parse logged JSON
	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(suite.T(), err)
	
	assert := assert.New(suite.T())
	assert.Equal("info", logEntry["level"])
	assert.Equal("User creation started", logEntry["message"])
	assert.Equal("user-123", logEntry["user_id"])
	assert.Equal("session-456", logEntry["session_id"])
	assert.Equal("create_user", logEntry["operation"])
	assert.NotEmpty(logEntry["timestamp"])
}

// TestLoggerWithField tests single field logging
func (suite *LoggerTestSuite) TestLoggerWithField() {
	var buf bytes.Buffer
	cfg := &config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	
	logger, err := NewLogger(cfg)
	assert.NoError(suite.T(), err)
	logger.SetOutput(&buf)
	
	logger.WithField("request_id", "req-789").Info("Processing request")
	
	// Parse logged JSON
	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(suite.T(), err)
	
	assert := assert.New(suite.T())
	assert.Equal("info", logEntry["level"])
	assert.Equal("Processing request", logEntry["message"])
	assert.Equal("req-789", logEntry["request_id"])
}

// TestLoggerWithError tests error logging
func (suite *LoggerTestSuite) TestLoggerWithError() {
	var buf bytes.Buffer
	cfg := &config.LoggingConfig{
		Level:  "error",
		Format: "json",
		Output: "stdout",
	}
	
	logger, err := NewLogger(cfg)
	assert.NoError(suite.T(), err)
	logger.SetOutput(&buf)
	
	testErr := assert.AnError
	logger.WithError(testErr).Error("Operation failed")
	
	// Parse logged JSON
	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(suite.T(), err)
	
	assert := assert.New(suite.T())
	assert.Equal("error", logEntry["level"])
	assert.Equal("Operation failed", logEntry["message"])
	assert.Equal(testErr.Error(), logEntry["error"])
}

// TestLoggerConvenienceMethods tests convenience logging methods
func (suite *LoggerTestSuite) TestLoggerConvenienceMethods() {
	var buf bytes.Buffer
	cfg := &config.LoggingConfig{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	}
	
	logger, err := NewLogger(cfg)
	assert.NoError(suite.T(), err)
	logger.SetOutput(&buf)
	
	testCases := []struct {
		method   func(string) *logrus.Entry
		field    string
		value    string
		expected string
	}{
		{logger.WithRequestID, "request_id", "req-123", "req-123"},
		{logger.WithSessionID, "session_id", "session-456", "session-456"},
		{logger.WithComponent, "component", "user-service", "user-service"},
	}
	
	for _, tc := range testCases {
		buf.Reset()
		tc.method(tc.value).Info("Test message")
		
		var logEntry map[string]interface{}
		err = json.Unmarshal(buf.Bytes(), &logEntry)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), tc.expected, logEntry[tc.field])
	}
}

// TestLoggerWithDuration tests duration logging
func (suite *LoggerTestSuite) TestLoggerWithDuration() {
	var buf bytes.Buffer
	cfg := &config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	
	logger, err := NewLogger(cfg)
	assert.NoError(suite.T(), err)
	logger.SetOutput(&buf)
	
	duration := 250 * time.Millisecond
	logger.WithDuration(duration).Info("Operation completed")
	
	// Parse logged JSON
	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(suite.T(), err)
	
	assert := assert.New(suite.T())
	assert.Equal("info", logEntry["level"])
	assert.Equal("Operation completed", logEntry["message"])
	assert.Equal(float64(250), logEntry["duration_ms"])
}

// TestLoggerSpecialLevels tests special log levels
func (suite *LoggerTestSuite) TestLoggerSpecialLevels() {
	var buf bytes.Buffer
	cfg := &config.LoggingConfig{
		Level:  "debug",
		Format: "console",
		Output: "stdout",
	}
	
	logger, err := NewLogger(cfg)
	assert.NoError(suite.T(), err)
	logger.SetOutput(&buf)
	
	// Test Notice (maps to Info)
	logger.Notice("This is a notice")
	assert.Contains(suite.T(), buf.String(), "This is a notice")
	
	buf.Reset()
	
	// Test Alert (maps to Error)
	logger.Alert("This is an alert")
	assert.Contains(suite.T(), buf.String(), "This is an alert")
	
	buf.Reset()
	
	// Test Critical (maps to Error)
	logger.Critical("This is critical")
	assert.Contains(suite.T(), buf.String(), "This is critical")
}

// TestLoggerContextualLoggers tests contextual logger creation
func (suite *LoggerTestSuite) TestLoggerContextualLoggers() {
	var buf bytes.Buffer
	cfg := &config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	
	logger, err := NewLogger(cfg)
	assert.NoError(suite.T(), err)
	logger.SetOutput(&buf)
	
	testCases := []struct {
		name     string
		logger   *logrus.Entry
		expected map[string]string
	}{
		{
			name:   "RequestLogger",
			logger: logger.RequestLogger("req-123", "POST", "/api/users"),
			expected: map[string]string{
				"request_id": "req-123",
				"method":     "POST",
				"path":       "/api/users",
				"component":  "http",
			},
		},
		{
			name:   "DatabaseLogger",
			logger: logger.DatabaseLogger("insert"),
			expected: map[string]string{
				"component": "database",
				"operation": "insert",
			},
		},
		{
			name:   "QueueLogger",
			logger: logger.QueueLogger("workflow_queue"),
			expected: map[string]string{
				"component":  "queue",
				"queue_name": "workflow_queue",
			},
		},
		{
			name:   "ClaudeLogger",
			logger: logger.ClaudeLogger("completion"),
			expected: map[string]string{
				"component": "claude",
				"operation": "completion",
			},
		},
		{
			name:   "HealthLogger",
			logger: logger.HealthLogger("database"),
			expected: map[string]string{
				"component": "health",
				"service":   "database",
			},
		},
		{
			name:   "MetricsLogger",
			logger: logger.MetricsLogger("request_count"),
			expected: map[string]string{
				"component": "metrics",
				"metric":    "request_count",
			},
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			buf.Reset()
			tc.logger.Info("Test message")
			
			var logEntry map[string]interface{}
			err = json.Unmarshal(buf.Bytes(), &logEntry)
			assert.NoError(t, err)
			
			for field, expectedValue := range tc.expected {
				assert.Equal(t, expectedValue, logEntry[field], "Field %s should match", field)
			}
		})
	}
}

// TestLoggerLevels tests different log levels
func (suite *LoggerTestSuite) TestLoggerLevels() {
	var buf bytes.Buffer
	cfg := &config.LoggingConfig{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	}
	
	logger, err := NewLogger(cfg)
	assert.NoError(suite.T(), err)
	logger.SetOutput(&buf)
	
	levels := []struct {
		level    string
		logFunc  func(args ...interface{})
		expected string
	}{
		{"debug", logger.Debug, "debug"},
		{"info", logger.Info, "info"},
		{"warn", logger.Warn, "warning"},
		{"error", logger.Error, "error"},
	}
	
	for _, level := range levels {
		suite.T().Run(level.level, func(t *testing.T) {
			buf.Reset()
			level.logFunc("Test message")
			
			var logEntry map[string]interface{}
			err = json.Unmarshal(buf.Bytes(), &logEntry)
			assert.NoError(t, err)
			assert.Equal(t, level.expected, logEntry["level"])
		})
	}
}

// TestLoggerLevelFiltering tests log level filtering
func (suite *LoggerTestSuite) TestLoggerLevelFiltering() {
	var buf bytes.Buffer
	cfg := &config.LoggingConfig{
		Level:  "warn", // Only warn and above should be logged
		Format: "console",
		Output: "stdout",
	}
	
	logger, err := NewLogger(cfg)
	assert.NoError(suite.T(), err)
	logger.SetOutput(&buf)
	
	// These should not be logged
	logger.Debug("Debug message")
	logger.Info("Info message")
	
	// These should be logged
	logger.Warn("Warning message")
	logger.Error("Error message")
	
	output := buf.String()
	assert := assert.New(suite.T())
	assert.NotContains(output, "Debug message")
	assert.NotContains(output, "Info message")
	assert.Contains(output, "Warning message")
	assert.Contains(output, "Error message")
}

// TestLoggerClose tests logger cleanup
func (suite *LoggerTestSuite) TestLoggerClose() {
	// Test closing stdout logger (should not error)
	cfg := &config.LoggingConfig{
		Level:  "info",
		Format: "console",
		Output: "stdout",
	}
	
	logger, err := NewLogger(cfg)
	assert.NoError(suite.T(), err)
	
	err = logger.Close()
	assert.NoError(suite.T(), err)
	
	// Test closing file logger
	logFile := filepath.Join(suite.tempDir, "close_test.log")
	fileCfg := &config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "file",
		File:   logFile,
	}
	
	fileLogger, err := NewLogger(fileCfg)
	assert.NoError(suite.T(), err)
	
	// Write something to ensure file is opened
	fileLogger.Info("Test message")
	
	err = fileLogger.Close()
	assert.NoError(suite.T(), err)
}

// TestLoggerFormatterDefaults tests default formatter behavior
func (suite *LoggerTestSuite) TestLoggerFormatterDefaults() {
	// Test default formatter for unknown format
	cfg := &config.LoggingConfig{
		Level:  "info",
		Format: "unknown",
		Output: "stdout",
	}
	
	logger, err := NewLogger(cfg)
	assert.NoError(suite.T(), err)
	
	// Should default to TextFormatter
	_, ok := logger.Formatter.(*logrus.TextFormatter)
	assert.True(suite.T(), ok)
	
	// Test default output for unknown output
	cfg.Output = "unknown"
	logger, err = NewLogger(cfg)
	assert.NoError(suite.T(), err)
	
	// Should default to stdout
	assert.Equal(suite.T(), os.Stdout, logger.Out)
}

// TestLoggerFieldTypes tests logging with different field types
func (suite *LoggerTestSuite) TestLoggerFieldTypes() {
	var buf bytes.Buffer
	cfg := &config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	
	logger, err := NewLogger(cfg)
	assert.NoError(suite.T(), err)
	logger.SetOutput(&buf)
	
	// Test various field types
	fields := Fields{
		"string_field":  "test_string",
		"int_field":     42,
		"float_field":   3.14,
		"bool_field":    true,
		"slice_field":   []string{"a", "b", "c"},
		"map_field":     map[string]string{"key": "value"},
		"nil_field":     nil,
		"time_field":    time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	
	logger.WithFields(fields).Info("Test with various field types")
	
	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(suite.T(), err)
	
	assert := assert.New(suite.T())
	assert.Equal("test_string", logEntry["string_field"])
	assert.Equal(float64(42), logEntry["int_field"]) // JSON numbers are float64
	assert.Equal(3.14, logEntry["float_field"])
	assert.Equal(true, logEntry["bool_field"])
	assert.Equal([]interface{}{"a", "b", "c"}, logEntry["slice_field"])
	assert.Equal(map[string]interface{}{"key": "value"}, logEntry["map_field"])
	assert.Nil(logEntry["nil_field"])
	assert.NotNil(logEntry["time_field"])
}

// BenchmarkLoggerCreation benchmarks logger creation
func BenchmarkLoggerCreation(b *testing.B) {
	cfg := &config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger, _ := NewLogger(cfg)
		logger.Close()
	}
}

// BenchmarkLoggerWithFields benchmarks structured logging
func BenchmarkLoggerWithFields(b *testing.B) {
	cfg := &config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	
	logger, _ := NewLogger(cfg)
	logger.SetOutput(&bytes.Buffer{}) // Discard output for benchmarking
	
	fields := Fields{
		"request_id": "req-123",
		"user_id":    "user-456",
		"operation":  "test_operation",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields(fields).Info("Benchmark message")
	}
}

// BenchmarkLoggerPlainLogging benchmarks plain logging
func BenchmarkLoggerPlainLogging(b *testing.B) {
	cfg := &config.LoggingConfig{
		Level:  "info",
		Format: "console",
		Output: "stdout",
	}
	
	logger, _ := NewLogger(cfg)
	logger.SetOutput(&bytes.Buffer{}) // Discard output for benchmarking
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark message")
	}
}