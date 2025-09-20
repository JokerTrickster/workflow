package infrastructure

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigValidator(t *testing.T) {
	validator := NewConfigValidator()
	assert.NotNil(t, validator)
}

func TestConfigValidator_ValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid configuration",
			config: &Config{
				App: AppConfig{
					Name:        "test-app",
					Version:     "1.0.0",
					Environment: "development",
				},
				Server: ServerConfig{
					Host: "localhost",
					Port: 8080,
				},
				Database: DatabaseConfig{
					Driver:       "sqlite",
					DSN:          ":memory:",
					MaxIdleConns: 5,
					MaxOpenConns: 10,
				},
				RabbitMQ: RabbitMQConfig{
					URL:         "amqp://localhost:5672",
					QueueName:   "test-queue",
					ConsumerTag: "test-consumer",
				},
				Claude: ClaudeConfig{
					APIKey:    "sk-ant-1234567890abcdefghijklmnopqrstuvwxyz1234567890",
					Model:     "claude-3-5-sonnet-20241022",
					MaxTokens: 1000,
					Timeout:   30,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
					Output: "stdout",
				},
			},
			expectError: false,
		},
		{
			name: "invalid app name",
			config: &Config{
				App: AppConfig{
					Name:        "",
					Version:     "1.0.0",
					Environment: "development",
				},
				Server:   ServerConfig{Host: "localhost", Port: 8080},
				Database: DatabaseConfig{Driver: "sqlite", DSN: ":memory:", MaxOpenConns: 10},
				RabbitMQ: RabbitMQConfig{URL: "amqp://localhost:5672", QueueName: "test", ConsumerTag: "test"},
				Claude:   ClaudeConfig{APIKey: "sk-ant-test", Model: "claude-3-5-sonnet-20241022", MaxTokens: 1000, Timeout: 30},
				Logging:  LoggingConfig{Level: "info", Format: "json", Output: "stdout"},
			},
			expectError: true,
			errorMsg:    "app config validation failed",
		},
		{
			name: "invalid port",
			config: &Config{
				App:      AppConfig{Name: "test", Version: "1.0.0", Environment: "development"},
				Server:   ServerConfig{Host: "localhost", Port: 70000},
				Database: DatabaseConfig{Driver: "sqlite", DSN: ":memory:", MaxOpenConns: 10},
				RabbitMQ: RabbitMQConfig{URL: "amqp://localhost:5672", QueueName: "test", ConsumerTag: "test"},
				Claude:   ClaudeConfig{APIKey: "sk-ant-test", Model: "claude-3-5-sonnet-20241022", MaxTokens: 1000, Timeout: 30},
				Logging:  LoggingConfig{Level: "info", Format: "json", Output: "stdout"},
			},
			expectError: true,
			errorMsg:    "server config validation failed",
		},
	}

	validator := NewConfigValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidator_ValidateApp(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name        string
		config      *AppConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid app config",
			config: &AppConfig{
				Name:        "test-app",
				Version:     "1.0.0",
				Environment: "development",
			},
			expectError: false,
		},
		{
			name: "empty app name",
			config: &AppConfig{
				Name:        "",
				Version:     "1.0.0",
				Environment: "development",
			},
			expectError: true,
			errorMsg:    "app name cannot be empty",
		},
		{
			name: "empty app version",
			config: &AppConfig{
				Name:        "test-app",
				Version:     "",
				Environment: "development",
			},
			expectError: true,
			errorMsg:    "app version cannot be empty",
		},
		{
			name: "invalid environment",
			config: &AppConfig{
				Name:        "test-app",
				Version:     "1.0.0",
				Environment: "invalid-env",
			},
			expectError: true,
			errorMsg:    "invalid environment",
		},
		{
			name: "valid production environment",
			config: &AppConfig{
				Name:        "test-app",
				Version:     "1.0.0",
				Environment: "production",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateApp(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidator_ValidateDatabase(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name        string
		config      *DatabaseConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid database config",
			config: &DatabaseConfig{
				Driver:       "sqlite",
				DSN:          ":memory:",
				MaxIdleConns: 5,
				MaxOpenConns: 10,
			},
			expectError: false,
		},
		{
			name: "unsupported driver",
			config: &DatabaseConfig{
				Driver:       "mysql",
				DSN:          "user:pass@tcp(localhost:3306)/db",
				MaxIdleConns: 5,
				MaxOpenConns: 10,
			},
			expectError: true,
			errorMsg:    "unsupported database driver",
		},
		{
			name: "invalid connection pool settings",
			config: &DatabaseConfig{
				Driver:       "sqlite",
				DSN:          ":memory:",
				MaxIdleConns: 15,
				MaxOpenConns: 10,
			},
			expectError: true,
			errorMsg:    "max idle connections (15) cannot exceed max open connections (10)",
		},
		{
			name: "negative max open connections",
			config: &DatabaseConfig{
				Driver:       "sqlite",
				DSN:          ":memory:",
				MaxIdleConns: 5,
				MaxOpenConns: -1,
			},
			expectError: true,
			errorMsg:    "max open connections must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateDatabase(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidator_ValidateRabbitMQ(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name        string
		config      *RabbitMQConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid rabbitmq config with amqp",
			config: &RabbitMQConfig{
				URL:         "amqp://localhost:5672",
				QueueName:   "test-queue",
				ConsumerTag: "test-consumer",
			},
			expectError: false,
		},
		{
			name: "valid rabbitmq config with amqps",
			config: &RabbitMQConfig{
				URL:         "amqps://secure.rabbitmq.com:5671",
				QueueName:   "secure-queue",
				ConsumerTag: "secure-consumer",
			},
			expectError: false,
		},
		{
			name: "invalid URL protocol",
			config: &RabbitMQConfig{
				URL:         "http://localhost:8080",
				QueueName:   "test-queue",
				ConsumerTag: "test-consumer",
			},
			expectError: true,
			errorMsg:    "rabbitmq URL must start with 'amqp://' or 'amqps://'",
		},
		{
			name: "empty queue name",
			config: &RabbitMQConfig{
				URL:         "amqp://localhost:5672",
				QueueName:   "",
				ConsumerTag: "test-consumer",
			},
			expectError: true,
			errorMsg:    "rabbitmq queue name cannot be empty",
		},
		{
			name: "queue name too long",
			config: &RabbitMQConfig{
				URL:         "amqp://localhost:5672",
				QueueName:   string(make([]byte, 256)),
				ConsumerTag: "test-consumer",
			},
			expectError: true,
			errorMsg:    "rabbitmq queue name cannot exceed 255 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateRabbitMQ(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidator_ValidateClaude(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name        string
		config      *ClaudeConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid claude config",
			config: &ClaudeConfig{
				APIKey:    "sk-ant-1234567890abcdefghijklmnopqrstuvwxyz1234567890",
				Model:     "claude-3-5-sonnet-20241022",
				MaxTokens: 1000,
				Timeout:   30,
			},
			expectError: false,
		},
		{
			name: "empty API key",
			config: &ClaudeConfig{
				APIKey:    "",
				Model:     "claude-3-5-sonnet-20241022",
				MaxTokens: 1000,
				Timeout:   30,
			},
			expectError: true,
			errorMsg:    "claude API key cannot be empty",
		},
		{
			name: "API key too short",
			config: &ClaudeConfig{
				APIKey:    "short",
				Model:     "claude-3-5-sonnet-20241022",
				MaxTokens: 1000,
				Timeout:   30,
			},
			expectError: true,
			errorMsg:    "claude API key appears to be invalid (too short)",
		},
		{
			name: "unknown model",
			config: &ClaudeConfig{
				APIKey:    "sk-ant-1234567890abcdefghijklmnopqrstuvwxyz1234567890",
				Model:     "claude-unknown-model",
				MaxTokens: 1000,
				Timeout:   30,
			},
			expectError: true,
			errorMsg:    "unknown claude model",
		},
		{
			name: "max tokens too high",
			config: &ClaudeConfig{
				APIKey:    "sk-ant-1234567890abcdefghijklmnopqrstuvwxyz1234567890",
				Model:     "claude-3-5-sonnet-20241022",
				MaxTokens: 300000,
				Timeout:   30,
			},
			expectError: true,
			errorMsg:    "claude max tokens cannot exceed 200000",
		},
		{
			name: "timeout too high",
			config: &ClaudeConfig{
				APIKey:    "sk-ant-1234567890abcdefghijklmnopqrstuvwxyz1234567890",
				Model:     "claude-3-5-sonnet-20241022",
				MaxTokens: 1000,
				Timeout:   400,
			},
			expectError: true,
			errorMsg:    "claude timeout cannot exceed 300 seconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateClaude(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidator_ValidateEnvironment(t *testing.T) {
	validator := NewConfigValidator()

	// Test with missing required environment variable
	t.Run("missing required env var", func(t *testing.T) {
		os.Unsetenv("CLAUDE_API_KEY")
		
		err := validator.ValidateEnvironment()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required environment variables")
		assert.Contains(t, err.Error(), "CLAUDE_API_KEY")
	})

	// Test with required environment variable set
	t.Run("required env var present", func(t *testing.T) {
		os.Setenv("CLAUDE_API_KEY", "sk-ant-test-key")
		defer os.Unsetenv("CLAUDE_API_KEY")
		
		err := validator.ValidateEnvironment()
		assert.NoError(t, err)
	})
}

func TestConfigValidator_ValidatePort(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name        string
		port        int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid port",
			port:        8080,
			expectError: false,
		},
		{
			name:        "port too low",
			port:        0,
			expectError: true,
			errorMsg:    "invalid port 0",
		},
		{
			name:        "port too high",
			port:        70000,
			expectError: true,
			errorMsg:    "invalid port 70000",
		},
		{
			name:        "reserved port range",
			port:        80,
			expectError: true,
			errorMsg:    "port 80 is in reserved range",
		},
		{
			name:        "valid high port",
			port:        3000,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePort(tt.port)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidator_ValidateConnectivity(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid connectivity config",
			config: &Config{
				RabbitMQ: RabbitMQConfig{
					URL: "amqp://localhost:5672",
				},
				Claude: ClaudeConfig{
					APIKey: "sk-ant-1234567890abcdefghijklmnopqrstuvwxyz1234567890",
				},
			},
			expectError: false,
		},
		{
			name: "invalid rabbitmq URL format",
			config: &Config{
				RabbitMQ: RabbitMQConfig{
					URL: "invalid-url-format",
				},
				Claude: ClaudeConfig{
					APIKey: "sk-ant-1234567890abcdefghijklmnopqrstuvwxyz1234567890",
				},
			},
			expectError: true,
			errorMsg:    "invalid RabbitMQ URL format",
		},
		{
			name: "invalid claude API key format",
			config: &Config{
				RabbitMQ: RabbitMQConfig{
					URL: "amqp://localhost:5672",
				},
				Claude: ClaudeConfig{
					APIKey: "invalid-key",
				},
			},
			expectError: true,
			errorMsg:    "claude API key format appears invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateConnectivity(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidator_GetConfigSummary(t *testing.T) {
	validator := NewConfigValidator()
	
	config := &Config{
		App: AppConfig{
			Name:        "test-app",
			Version:     "1.0.0",
			Environment: "development",
		},
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Driver:       "sqlite",
			MaxIdleConns: 5,
			MaxOpenConns: 10,
		},
		RabbitMQ: RabbitMQConfig{
			URL:         "amqp://localhost:5672",
			QueueName:   "test-queue",
			ConsumerTag: "test-consumer",
		},
		Claude: ClaudeConfig{
			APIKey:    "sk-ant-secret-key",
			Model:     "claude-3-5-sonnet-20241022",
			MaxTokens: 1000,
			Timeout:   30,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
	}

	summary := validator.GetConfigSummary(config)

	// Verify structure
	assert.Contains(t, summary, "app")
	assert.Contains(t, summary, "server")
	assert.Contains(t, summary, "database")
	assert.Contains(t, summary, "rabbitmq")
	assert.Contains(t, summary, "claude")
	assert.Contains(t, summary, "logging")

	// Verify sensitive data is not exposed
	claudeSection := summary["claude"].(map[string]interface{})
	assert.Equal(t, true, claudeSection["api_key_set"])
	assert.NotContains(t, summary, "sk-ant-secret-key")

	rabbitmqSection := summary["rabbitmq"].(map[string]interface{})
	assert.Equal(t, true, rabbitmqSection["url_set"])
	assert.NotContains(t, summary, "amqp://localhost:5672")

	// Verify non-sensitive data is included
	appSection := summary["app"].(map[string]interface{})
	assert.Equal(t, "test-app", appSection["name"])
	assert.Equal(t, "development", appSection["environment"])
}

func TestContains(t *testing.T) {
	testSlice := []string{"apple", "banana", "cherry"}

	tests := []struct {
		name     string
		slice    []string
		str      string
		expected bool
	}{
		{
			name:     "contains exact match",
			slice:    testSlice,
			str:      "banana",
			expected: true,
		},
		{
			name:     "contains case insensitive",
			slice:    testSlice,
			str:      "APPLE",
			expected: true,
		},
		{
			name:     "does not contain",
			slice:    testSlice,
			str:      "grape",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			str:      "anything",
			expected: false,
		},
		{
			name:     "empty string",
			slice:    testSlice,
			str:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.str)
			assert.Equal(t, tt.expected, result)
		})
	}
}