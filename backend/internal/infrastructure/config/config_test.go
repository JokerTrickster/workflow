package config

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ConfigTestSuite defines the test suite for Config
type ConfigTestSuite struct {
	suite.Suite
	originalEnv map[string]string
}

// SetupTest runs before each test
func (suite *ConfigTestSuite) SetupTest() {
	// Save original environment variables
	suite.originalEnv = make(map[string]string)
	envVars := []string{"ENV", "CLAUDE_API_KEY", "DB_USER", "DB_PASSWORD", "RABBITMQ_URL"}
	for _, envVar := range envVars {
		if val := os.Getenv(envVar); val != "" {
			suite.originalEnv[envVar] = val
		}
	}
	
	// Clear viper configuration
	viper.Reset()
}

// TearDownTest runs after each test
func (suite *ConfigTestSuite) TearDownTest() {
	// Restore original environment variables
	for _, envVar := range []string{"ENV", "CLAUDE_API_KEY", "DB_USER", "DB_PASSWORD", "RABBITMQ_URL"} {
		os.Unsetenv(envVar)
		if val, exists := suite.originalEnv[envVar]; exists {
			os.Setenv(envVar, val)
		}
	}
	
	// Clear viper configuration
	viper.Reset()
}

// TestConfigTestSuite runs the test suite
func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

// TestLoadConfig_Development tests loading configuration for development environment
func (suite *ConfigTestSuite) TestLoadConfig_Development() {
	// Set development environment
	os.Setenv("ENV", "development")
	
	// Load config
	cfg, err := LoadConfig()
	
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.NotNil(cfg)
	assert.Equal("development", cfg.Environment)
	assert.True(cfg.IsDevelopment())
	assert.False(cfg.IsProduction())
	assert.False(cfg.IsStaging())
	
	// Check default values
	assert.Equal(8080, cfg.Server.Port)
	assert.Equal("localhost", cfg.Server.Host)
	assert.Equal("sqlite", cfg.Database.Driver)
	assert.Equal("./data/workflow.db", cfg.Database.Path)
	assert.Equal("amqp://localhost:5672", cfg.RabbitMQ.URL)
	assert.Equal("workflow_queue", cfg.RabbitMQ.QueueName)
	assert.Equal("claude-3-sonnet-20240229", cfg.Claude.Model)
	assert.Equal(4096, cfg.Claude.MaxTokens)
	assert.Equal(0.7, cfg.Claude.Temperature)
	assert.Equal("debug", cfg.Logging.Level)
	assert.Equal("console", cfg.Logging.Format)
	assert.Equal("stdout", cfg.Logging.Output)
	assert.True(cfg.Logging.EnableCaller)
	assert.True(cfg.Logging.EnableStacktrace)
	assert.Equal(8081, cfg.Health.Port)
	assert.Equal(30*time.Second, cfg.Health.Interval)
	assert.Equal(30*time.Second, cfg.Timeout.Request)
}

// TestLoadConfig_Production tests loading configuration for production environment
func (suite *ConfigTestSuite) TestLoadConfig_Production() {
	// Set production environment and required API key
	os.Setenv("ENV", "production")
	os.Setenv("CLAUDE_API_KEY", "test-api-key")
	
	// Load config
	cfg, err := LoadConfig()
	
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.NotNil(cfg)
	assert.Equal("production", cfg.Environment)
	assert.True(cfg.IsProduction())
	assert.False(cfg.IsDevelopment())
	assert.False(cfg.IsStaging())
	
	// Check production-specific logging defaults
	assert.Equal("info", cfg.Logging.Level)
	assert.Equal("json", cfg.Logging.Format)
	assert.Equal("file", cfg.Logging.Output)
	assert.Equal("./logs/app.log", cfg.Logging.File)
	assert.False(cfg.Logging.EnableCaller)
	assert.False(cfg.Logging.EnableStacktrace)
	
	// Check that Claude API key is set
	assert.Equal("test-api-key", cfg.Claude.APIKey)
}

// TestLoadConfig_Staging tests loading configuration for staging environment
func (suite *ConfigTestSuite) TestLoadConfig_Staging() {
	// Set staging environment
	os.Setenv("ENV", "staging")
	
	// Load config
	cfg, err := LoadConfig()
	
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.NotNil(cfg)
	assert.Equal("staging", cfg.Environment)
	assert.True(cfg.IsStaging())
	assert.False(cfg.IsDevelopment())
	assert.False(cfg.IsProduction())
	
	// Check staging-specific logging defaults
	assert.Equal("debug", cfg.Logging.Level)
	assert.Equal("json", cfg.Logging.Format)
	assert.Equal("stdout", cfg.Logging.Output)
	assert.True(cfg.Logging.EnableCaller)
	assert.False(cfg.Logging.EnableStacktrace)
}

// TestLoadConfig_DefaultEnvironment tests loading config with no environment set
func (suite *ConfigTestSuite) TestLoadConfig_DefaultEnvironment() {
	// Ensure ENV is not set
	os.Unsetenv("ENV")
	
	// Load config
	cfg, err := LoadConfig()
	
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.NotNil(cfg)
	assert.Equal("development", cfg.Environment) // Should default to development
	assert.True(cfg.IsDevelopment())
}

// TestLoadConfig_SensitiveEnvironmentVariables tests loading sensitive data from environment variables
func (suite *ConfigTestSuite) TestLoadConfig_SensitiveEnvironmentVariables() {
	// Set environment variables
	os.Setenv("ENV", "development")
	os.Setenv("CLAUDE_API_KEY", "secret-claude-key")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("RABBITMQ_URL", "amqp://testuser:testpass@localhost:5672")
	
	// Load config
	cfg, err := LoadConfig()
	
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.NotNil(cfg)
	assert.Equal("secret-claude-key", cfg.Claude.APIKey)
	assert.Equal("amqp://testuser:testpass@localhost:5672", cfg.RabbitMQ.URL)
}

// TestLoadConfig_ViperEnvironmentVariables tests loading config through viper environment variables
func (suite *ConfigTestSuite) TestLoadConfig_ViperEnvironmentVariables() {
	// Set environment variables with LOCAL_BACKEND prefix
	os.Setenv("ENV", "development")
	os.Setenv("LOCAL_BACKEND_SERVER_PORT", "9090")
	os.Setenv("LOCAL_BACKEND_SERVER_HOST", "0.0.0.0")
	os.Setenv("LOCAL_BACKEND_DATABASE_MAX_CONNECTIONS", "20")
	os.Setenv("LOCAL_BACKEND_CLAUDE_MAX_TOKENS", "8192")
	os.Setenv("LOCAL_BACKEND_CLAUDE_TEMPERATURE", "0.5")
	
	// Load config
	cfg, err := LoadConfig()
	
	assert := assert.New(suite.T())
	assert.NoError(err)
	assert.NotNil(cfg)
	assert.Equal(9090, cfg.Server.Port)
	assert.Equal("0.0.0.0", cfg.Server.Host)
	assert.Equal(20, cfg.Database.MaxConnections)
	assert.Equal(8192, cfg.Claude.MaxTokens)
	assert.Equal(0.5, cfg.Claude.Temperature)
	
	// Clean up
	os.Unsetenv("LOCAL_BACKEND_SERVER_PORT")
	os.Unsetenv("LOCAL_BACKEND_SERVER_HOST")
	os.Unsetenv("LOCAL_BACKEND_DATABASE_MAX_CONNECTIONS")
	os.Unsetenv("LOCAL_BACKEND_CLAUDE_MAX_TOKENS")
	os.Unsetenv("LOCAL_BACKEND_CLAUDE_TEMPERATURE")
}

// TestConfigValidation_Success tests successful configuration validation
func (suite *ConfigTestSuite) TestConfigValidation_Success() {
	// Create valid config
	cfg := &Config{
		Environment: "development",
		Server: ServerConfig{
			Port:            8080,
			Host:            "localhost",
			ReadTimeout:     15 * time.Second,
			WriteTimeout:    15 * time.Second,
			ShutdownTimeout: 30 * time.Second,
		},
		Database: DatabaseConfig{
			Driver:          "sqlite",
			Path:            "./test.db",
			MaxConnections:  10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 1 * time.Hour,
			RetryAttempts:   3,
			RetryDelay:      1 * time.Second,
		},
		RabbitMQ: RabbitMQConfig{
			URL:            "amqp://localhost:5672",
			QueueName:      "test_queue",
			MaxRetries:     3,
			RetryDelay:     5 * time.Second,
			ReconnectDelay: 10 * time.Second,
			PrefetchCount:  1,
			Durable:        true,
			AutoDelete:     false,
		},
		Claude: ClaudeConfig{
			APIKey:         "test-key",
			Model:          "claude-3-sonnet-20240229",
			MaxTokens:      4096,
			Temperature:    0.7,
			RequestTimeout: 120 * time.Second,
			MaxRetries:     3,
			RetryDelay:     2 * time.Second,
		},
		Logging: LoggingConfig{
			Level:            "debug",
			Format:           "console",
			Output:           "stdout",
			EnableCaller:     true,
			EnableStacktrace: true,
		},
		Health: HealthConfig{
			Port:     8081,
			Interval: 30 * time.Second,
			Timeout:  10 * time.Second,
		},
		Timeout: TimeoutConfig{
			Request:  30 * time.Second,
			Database: 5 * time.Second,
			Queue:    10 * time.Second,
			Claude:   120 * time.Second,
			Shutdown: 30 * time.Second,
		},
	}
	
	err := cfg.Validate()
	assert := assert.New(suite.T())
	assert.NoError(err)
}

// TestConfigValidation_InvalidServerPort tests validation with invalid server port
func (suite *ConfigTestSuite) TestConfigValidation_InvalidServerPort() {
	cfg := &Config{
		Server: ServerConfig{
			Port: 0, // Invalid port
			Host: "localhost",
		},
		Database: DatabaseConfig{
			Driver:         "sqlite",
			Path:           "./test.db",
			MaxConnections: 10,
		},
		RabbitMQ: RabbitMQConfig{
			URL:       "amqp://localhost:5672",
			QueueName: "test_queue",
		},
		Claude: ClaudeConfig{
			Model:       "claude-3-sonnet-20240229",
			MaxTokens:   4096,
			Temperature: 0.7,
		},
		Logging: LoggingConfig{
			Level:  "debug",
			Format: "console",
		},
		Health: HealthConfig{
			Port: 8081,
		},
		Timeout: TimeoutConfig{
			Request:  30 * time.Second,
			Database: 5 * time.Second,
		},
	}
	
	err := cfg.Validate()
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Contains(err.Error(), "server.port must be between 1 and 65535")
}

// TestConfigValidation_EmptyRequiredFields tests validation with empty required fields
func (suite *ConfigTestSuite) TestConfigValidation_EmptyRequiredFields() {
	testCases := []struct {
		name     string
		config   *Config
		errorMsg string
	}{
		{
			name: "empty server host",
			config: &Config{
				Server: ServerConfig{Port: 8080, Host: ""},
				Database: DatabaseConfig{Driver: "sqlite", Path: "./test.db", MaxConnections: 10},
				RabbitMQ: RabbitMQConfig{URL: "amqp://localhost:5672", QueueName: "test"},
				Claude:   ClaudeConfig{Model: "claude-3-sonnet-20240229", MaxTokens: 4096, Temperature: 0.7},
				Logging:  LoggingConfig{Level: "debug", Format: "console"},
				Health:   HealthConfig{Port: 8081},
				Timeout:  TimeoutConfig{Request: 30 * time.Second, Database: 5 * time.Second},
			},
			errorMsg: "server.host cannot be empty",
		},
		{
			name: "empty database driver",
			config: &Config{
				Server:   ServerConfig{Port: 8080, Host: "localhost"},
				Database: DatabaseConfig{Driver: "", Path: "./test.db", MaxConnections: 10},
				RabbitMQ: RabbitMQConfig{URL: "amqp://localhost:5672", QueueName: "test"},
				Claude:   ClaudeConfig{Model: "claude-3-sonnet-20240229", MaxTokens: 4096, Temperature: 0.7},
				Logging:  LoggingConfig{Level: "debug", Format: "console"},
				Health:   HealthConfig{Port: 8081},
				Timeout:  TimeoutConfig{Request: 30 * time.Second, Database: 5 * time.Second},
			},
			errorMsg: "database.driver cannot be empty",
		},
		{
			name: "empty rabbitmq url",
			config: &Config{
				Server:   ServerConfig{Port: 8080, Host: "localhost"},
				Database: DatabaseConfig{Driver: "sqlite", Path: "./test.db", MaxConnections: 10},
				RabbitMQ: RabbitMQConfig{URL: "", QueueName: "test"},
				Claude:   ClaudeConfig{Model: "claude-3-sonnet-20240229", MaxTokens: 4096, Temperature: 0.7},
				Logging:  LoggingConfig{Level: "debug", Format: "console"},
				Health:   HealthConfig{Port: 8081},
				Timeout:  TimeoutConfig{Request: 30 * time.Second, Database: 5 * time.Second},
			},
			errorMsg: "rabbitmq.url cannot be empty",
		},
		{
			name: "empty claude model",
			config: &Config{
				Server:   ServerConfig{Port: 8080, Host: "localhost"},
				Database: DatabaseConfig{Driver: "sqlite", Path: "./test.db", MaxConnections: 10},
				RabbitMQ: RabbitMQConfig{URL: "amqp://localhost:5672", QueueName: "test"},
				Claude:   ClaudeConfig{Model: "", MaxTokens: 4096, Temperature: 0.7},
				Logging:  LoggingConfig{Level: "debug", Format: "console"},
				Health:   HealthConfig{Port: 8081},
				Timeout:  TimeoutConfig{Request: 30 * time.Second, Database: 5 * time.Second},
			},
			errorMsg: "claude.model cannot be empty",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			assert := assert.New(t)
			assert.Error(err)
			assert.Contains(err.Error(), tc.errorMsg)
		})
	}
}

// TestConfigValidation_InvalidValues tests validation with invalid values
func (suite *ConfigTestSuite) TestConfigValidation_InvalidValues() {
	testCases := []struct {
		name     string
		config   *Config
		errorMsg string
	}{
		{
			name: "invalid claude temperature",
			config: &Config{
				Environment: "development",
				Server:      ServerConfig{Port: 8080, Host: "localhost"},
				Database:    DatabaseConfig{Driver: "sqlite", Path: "./test.db", MaxConnections: 10},
				RabbitMQ:    RabbitMQConfig{URL: "amqp://localhost:5672", QueueName: "test"},
				Claude:      ClaudeConfig{Model: "claude-3-sonnet-20240229", MaxTokens: 4096, Temperature: 3.0}, // Invalid
				Logging:     LoggingConfig{Level: "debug", Format: "console"},
				Health:      HealthConfig{Port: 8081},
				Timeout:     TimeoutConfig{Request: 30 * time.Second, Database: 5 * time.Second},
			},
			errorMsg: "claude.temperature must be between 0 and 2",
		},
		{
			name: "invalid logging level",
			config: &Config{
				Environment: "development",
				Server:      ServerConfig{Port: 8080, Host: "localhost"},
				Database:    DatabaseConfig{Driver: "sqlite", Path: "./test.db", MaxConnections: 10},
				RabbitMQ:    RabbitMQConfig{URL: "amqp://localhost:5672", QueueName: "test"},
				Claude:      ClaudeConfig{Model: "claude-3-sonnet-20240229", MaxTokens: 4096, Temperature: 0.7},
				Logging:     LoggingConfig{Level: "invalid", Format: "console"}, // Invalid
				Health:      HealthConfig{Port: 8081},
				Timeout:     TimeoutConfig{Request: 30 * time.Second, Database: 5 * time.Second},
			},
			errorMsg: "logging.level must be one of:",
		},
		{
			name: "same health and server ports",
			config: &Config{
				Environment: "development",
				Server:      ServerConfig{Port: 8080, Host: "localhost"},
				Database:    DatabaseConfig{Driver: "sqlite", Path: "./test.db", MaxConnections: 10},
				RabbitMQ:    RabbitMQConfig{URL: "amqp://localhost:5672", QueueName: "test"},
				Claude:      ClaudeConfig{Model: "claude-3-sonnet-20240229", MaxTokens: 4096, Temperature: 0.7},
				Logging:     LoggingConfig{Level: "debug", Format: "console"},
				Health:      HealthConfig{Port: 8080}, // Same as server port
				Timeout:     TimeoutConfig{Request: 30 * time.Second, Database: 5 * time.Second},
			},
			errorMsg: "health.port cannot be the same as server.port",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			assert := assert.New(t)
			assert.Error(err)
			assert.Contains(err.Error(), tc.errorMsg)
		})
	}
}

// TestConfigValidation_ProductionRequirements tests production-specific validation
func (suite *ConfigTestSuite) TestConfigValidation_ProductionRequirements() {
	cfg := &Config{
		Environment: "production",
		Server:      ServerConfig{Port: 8080, Host: "localhost"},
		Database:    DatabaseConfig{Driver: "sqlite", Path: "./test.db", MaxConnections: 10},
		RabbitMQ:    RabbitMQConfig{URL: "amqp://localhost:5672", QueueName: "test"},
		Claude:      ClaudeConfig{APIKey: "", Model: "claude-3-sonnet-20240229", MaxTokens: 4096, Temperature: 0.7}, // Missing API key
		Logging:     LoggingConfig{Level: "debug", Format: "console"},
		Health:      HealthConfig{Port: 8081},
		Timeout:     TimeoutConfig{Request: 30 * time.Second, Database: 5 * time.Second},
	}
	
	err := cfg.Validate()
	assert := assert.New(suite.T())
	assert.Error(err)
	assert.Contains(err.Error(), "claude.api_key is required in production environment")
}

// TestConfigValidation_MultipleErrors tests validation with multiple errors
func (suite *ConfigTestSuite) TestConfigValidation_MultipleErrors() {
	cfg := &Config{
		Environment: "development",
		Server:      ServerConfig{Port: 0, Host: ""},           // Multiple server errors
		Database:    DatabaseConfig{Driver: "", Path: ""},     // Multiple database errors
		RabbitMQ:    RabbitMQConfig{URL: "", QueueName: ""},   // Multiple RabbitMQ errors
		Claude:      ClaudeConfig{Model: "", MaxTokens: 0},    // Multiple Claude errors
		Logging:     LoggingConfig{Level: "invalid", Format: "invalid"}, // Multiple logging errors
		Health:      HealthConfig{Port: 0},
		Timeout:     TimeoutConfig{Request: 0, Database: 0},
	}
	
	err := cfg.Validate()
	assert := assert.New(suite.T())
	assert.Error(err)
	
	// Check that multiple errors are reported
	errorString := err.Error()
	assert.Contains(errorString, "server.port must be between 1 and 65535")
	assert.Contains(errorString, "server.host cannot be empty")
	assert.Contains(errorString, "database.driver cannot be empty")
	assert.Contains(errorString, "database.path cannot be empty")
	assert.Contains(errorString, "rabbitmq.url cannot be empty")
	assert.Contains(errorString, "rabbitmq.queue_name cannot be empty")
	assert.Contains(errorString, "claude.model cannot be empty")
}

// TestSetDefaults tests the setDefaults function
func (suite *ConfigTestSuite) TestSetDefaults() {
	viper.Reset()
	
	// Test development defaults
	setDefaults("development")
	
	assert := assert.New(suite.T())
	assert.Equal(8080, viper.GetInt("server.port"))
	assert.Equal("localhost", viper.GetString("server.host"))
	assert.Equal("sqlite", viper.GetString("database.driver"))
	assert.Equal("./data/workflow.db", viper.GetString("database.path"))
	assert.Equal("debug", viper.GetString("logging.level"))
	assert.Equal("console", viper.GetString("logging.format"))
	assert.Equal("stdout", viper.GetString("logging.output"))
	assert.True(viper.GetBool("logging.enable_caller"))
	assert.True(viper.GetBool("logging.enable_stacktrace"))
}

// TestSetLoggingDefaults tests logging defaults for different environments
func (suite *ConfigTestSuite) TestSetLoggingDefaults() {
	testCases := []struct {
		env                   string
		expectedLevel         string
		expectedFormat        string
		expectedOutput        string
		expectedEnableCaller  bool
		expectedEnableStack   bool
	}{
		{
			env:                   "production",
			expectedLevel:         "info",
			expectedFormat:        "json",
			expectedOutput:        "file",
			expectedEnableCaller:  false,
			expectedEnableStack:   false,
		},
		{
			env:                   "staging",
			expectedLevel:         "debug",
			expectedFormat:        "json",
			expectedOutput:        "stdout",
			expectedEnableCaller:  true,
			expectedEnableStack:   false,
		},
		{
			env:                   "development",
			expectedLevel:         "debug",
			expectedFormat:        "console",
			expectedOutput:        "stdout",
			expectedEnableCaller:  true,
			expectedEnableStack:   true,
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.env, func(t *testing.T) {
			viper.Reset()
			setLoggingDefaults(tc.env)
			
			assert := assert.New(t)
			assert.Equal(tc.expectedLevel, viper.GetString("logging.level"))
			assert.Equal(tc.expectedFormat, viper.GetString("logging.format"))
			assert.Equal(tc.expectedOutput, viper.GetString("logging.output"))
			assert.Equal(tc.expectedEnableCaller, viper.GetBool("logging.enable_caller"))
			assert.Equal(tc.expectedEnableStack, viper.GetBool("logging.enable_stacktrace"))
		})
	}
}

// TestContains tests the contains helper function
func (suite *ConfigTestSuite) TestContains() {
	slice := []string{"a", "b", "c", "d"}
	
	assert := assert.New(suite.T())
	assert.True(contains(slice, "a"))
	assert.True(contains(slice, "c"))
	assert.False(contains(slice, "z"))
	assert.False(contains(slice, ""))
	
	// Test empty slice
	assert.False(contains([]string{}, "a"))
}

// TestEnvironmentHelpers tests environment helper methods
func (suite *ConfigTestSuite) TestEnvironmentHelpers() {
	testCases := []struct {
		env           string
		isDevelopment bool
		isProduction  bool
		isStaging     bool
	}{
		{"development", true, false, false},
		{"production", false, true, false},
		{"staging", false, false, true},
		{"test", false, false, false},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.env, func(t *testing.T) {
			cfg := &Config{Environment: tc.env}
			
			assert := assert.New(t)
			assert.Equal(tc.isDevelopment, cfg.IsDevelopment())
			assert.Equal(tc.isProduction, cfg.IsProduction())
			assert.Equal(tc.isStaging, cfg.IsStaging())
		})
	}
}

// BenchmarkLoadConfig benchmarks configuration loading
func BenchmarkLoadConfig(b *testing.B) {
	// Set development environment
	os.Setenv("ENV", "development")
	defer os.Unsetenv("ENV")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		viper.Reset()
		LoadConfig()
	}
}

// BenchmarkConfigValidation benchmarks configuration validation
func BenchmarkConfigValidation(b *testing.B) {
	cfg := &Config{
		Environment: "development",
		Server:      ServerConfig{Port: 8080, Host: "localhost"},
		Database:    DatabaseConfig{Driver: "sqlite", Path: "./test.db", MaxConnections: 10},
		RabbitMQ:    RabbitMQConfig{URL: "amqp://localhost:5672", QueueName: "test"},
		Claude:      ClaudeConfig{Model: "claude-3-sonnet-20240229", MaxTokens: 4096, Temperature: 0.7},
		Logging:     LoggingConfig{Level: "debug", Format: "console"},
		Health:      HealthConfig{Port: 8081},
		Timeout:     TimeoutConfig{Request: 30 * time.Second, Database: 5 * time.Second},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg.Validate()
	}
}