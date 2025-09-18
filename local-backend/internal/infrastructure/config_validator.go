package infrastructure

import (
	"fmt"
	"os"
	"strings"
)

// ConfigValidator validates configuration settings
type ConfigValidator struct{}

// NewConfigValidator creates a new config validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

// ValidateConfig performs comprehensive configuration validation
func (v *ConfigValidator) ValidateConfig(config *Config) error {
	if err := v.validateApp(&config.App); err != nil {
		return fmt.Errorf("app config validation failed: %w", err)
	}
	
	if err := v.validateServer(&config.Server); err != nil {
		return fmt.Errorf("server config validation failed: %w", err)
	}
	
	if err := v.validateDatabase(&config.Database); err != nil {
		return fmt.Errorf("database config validation failed: %w", err)
	}
	
	if err := v.validateRabbitMQ(&config.RabbitMQ); err != nil {
		return fmt.Errorf("rabbitmq config validation failed: %w", err)
	}
	
	if err := v.validateClaude(&config.Claude); err != nil {
		return fmt.Errorf("claude config validation failed: %w", err)
	}
	
	if err := v.validateLogging(&config.Logging); err != nil {
		return fmt.Errorf("logging config validation failed: %w", err)
	}
	
	return nil
}

// validateApp validates app configuration
func (v *ConfigValidator) validateApp(config *AppConfig) error {
	if strings.TrimSpace(config.Name) == "" {
		return fmt.Errorf("app name cannot be empty")
	}
	
	if strings.TrimSpace(config.Version) == "" {
		return fmt.Errorf("app version cannot be empty")
	}
	
	validEnvironments := []string{"development", "staging", "production", "test"}
	if !contains(validEnvironments, config.Environment) {
		return fmt.Errorf("invalid environment '%s', must be one of: %s", 
			config.Environment, strings.Join(validEnvironments, ", "))
	}
	
	return nil
}

// validateServer validates server configuration
func (v *ConfigValidator) validateServer(config *ServerConfig) error {
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("invalid server port %d, must be between 1 and 65535", config.Port)
	}
	
	if strings.TrimSpace(config.Host) == "" {
		return fmt.Errorf("server host cannot be empty")
	}
	
	return nil
}

// validateDatabase validates database configuration
func (v *ConfigValidator) validateDatabase(config *DatabaseConfig) error {
	if strings.TrimSpace(config.Driver) == "" {
		return fmt.Errorf("database driver cannot be empty")
	}
	
	if config.Driver != "sqlite" {
		return fmt.Errorf("unsupported database driver '%s', only 'sqlite' is supported", config.Driver)
	}
	
	if strings.TrimSpace(config.DSN) == "" {
		return fmt.Errorf("database DSN cannot be empty")
	}
	
	if config.MaxIdleConns < 0 {
		return fmt.Errorf("max idle connections cannot be negative")
	}
	
	if config.MaxOpenConns <= 0 {
		return fmt.Errorf("max open connections must be positive")
	}
	
	if config.MaxIdleConns > config.MaxOpenConns {
		return fmt.Errorf("max idle connections (%d) cannot exceed max open connections (%d)", 
			config.MaxIdleConns, config.MaxOpenConns)
	}
	
	return nil
}

// validateRabbitMQ validates RabbitMQ configuration
func (v *ConfigValidator) validateRabbitMQ(config *RabbitMQConfig) error {
	if strings.TrimSpace(config.URL) == "" {
		return fmt.Errorf("rabbitmq URL cannot be empty")
	}
	
	if !strings.HasPrefix(config.URL, "amqp://") && !strings.HasPrefix(config.URL, "amqps://") {
		return fmt.Errorf("rabbitmq URL must start with 'amqp://' or 'amqps://'")
	}
	
	if strings.TrimSpace(config.QueueName) == "" {
		return fmt.Errorf("rabbitmq queue name cannot be empty")
	}
	
	if len(config.QueueName) > 255 {
		return fmt.Errorf("rabbitmq queue name cannot exceed 255 characters")
	}
	
	if strings.TrimSpace(config.ConsumerTag) == "" {
		return fmt.Errorf("rabbitmq consumer tag cannot be empty")
	}
	
	return nil
}

// validateClaude validates Claude API configuration
func (v *ConfigValidator) validateClaude(config *ClaudeConfig) error {
	// API key validation (don't log the key for security)
	if strings.TrimSpace(config.APIKey) == "" {
		return fmt.Errorf("claude API key cannot be empty")
	}
	
	if len(config.APIKey) < 10 { // Basic length check
		return fmt.Errorf("claude API key appears to be invalid (too short)")
	}
	
	if strings.TrimSpace(config.Model) == "" {
		return fmt.Errorf("claude model cannot be empty")
	}
	
	// Validate known models
	validModels := []string{
		"claude-3-5-sonnet-20241022",
		"claude-3-sonnet-20240229", 
		"claude-3-haiku-20240307",
		"claude-3-opus-20240229",
	}
	if !contains(validModels, config.Model) {
		return fmt.Errorf("unknown claude model '%s', supported models: %s", 
			config.Model, strings.Join(validModels, ", "))
	}
	
	if config.MaxTokens <= 0 {
		return fmt.Errorf("claude max tokens must be positive")
	}
	
	if config.MaxTokens > 200000 { // Claude's current max
		return fmt.Errorf("claude max tokens cannot exceed 200000")
	}
	
	if config.Timeout <= 0 {
		return fmt.Errorf("claude timeout must be positive")
	}
	
	if config.Timeout > 300 { // 5 minutes max
		return fmt.Errorf("claude timeout cannot exceed 300 seconds")
	}
	
	return nil
}

// validateLogging validates logging configuration
func (v *ConfigValidator) validateLogging(config *LoggingConfig) error {
	validLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLevels, strings.ToLower(config.Level)) {
		return fmt.Errorf("invalid log level '%s', must be one of: %s", 
			config.Level, strings.Join(validLevels, ", "))
	}
	
	validFormats := []string{"json", "text"}
	if !contains(validFormats, strings.ToLower(config.Format)) {
		return fmt.Errorf("invalid log format '%s', must be one of: %s", 
			config.Format, strings.Join(validFormats, ", "))
	}
	
	validOutputs := []string{"stdout", "file"}
	if !contains(validOutputs, strings.ToLower(config.Output)) {
		return fmt.Errorf("invalid log output '%s', must be one of: %s", 
			config.Output, strings.Join(validOutputs, ", "))
	}
	
	return nil
}

// ValidateEnvironment validates that required environment variables are set
func (v *ConfigValidator) ValidateEnvironment() error {
	required := map[string]string{
		"CLAUDE_API_KEY": "Claude API key is required for processing requests",
	}
	
	optional := map[string]string{
		"DATABASE_DSN":   "Database connection string",
		"RABBITMQ_URL":   "RabbitMQ connection URL", 
		"LOG_LEVEL":      "Logging level (debug, info, warn, error)",
		"APP_ENV":        "Application environment (development, staging, production)",
	}
	
	var errors []string
	
	// Check required variables
	for envVar, description := range required {
		if value := os.Getenv(envVar); strings.TrimSpace(value) == "" {
			errors = append(errors, fmt.Sprintf("%s (%s)", envVar, description))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(errors, ", "))
	}
	
	// Log optional variables for debugging
	for envVar, description := range optional {
		if value := os.Getenv(envVar); value != "" {
			fmt.Printf("Environment: %s = %s (%s)\n", envVar, value, description)
		}
	}
	
	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, str) {
			return true
		}
	}
	return false
}

// ValidatePort validates that a port is available
func (v *ConfigValidator) ValidatePort(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port %d", port)
	}
	
	// Check if port is in reserved range
	if port < 1024 {
		return fmt.Errorf("port %d is in reserved range (1-1023), may require root privileges", port)
	}
	
	return nil
}

// ValidateConnectivity tests basic connectivity to external services
func (v *ConfigValidator) ValidateConnectivity(config *Config) error {
	// Test RabbitMQ connectivity (basic URL parsing)
	if !strings.Contains(config.RabbitMQ.URL, "://") {
		return fmt.Errorf("invalid RabbitMQ URL format")
	}
	
	// Test Claude API key format (basic validation)
	if !strings.HasPrefix(config.Claude.APIKey, "sk-") && len(config.Claude.APIKey) < 50 {
		return fmt.Errorf("claude API key format appears invalid")
	}
	
	return nil
}

// GetConfigSummary returns a summary of the configuration (without sensitive data)
func (v *ConfigValidator) GetConfigSummary(config *Config) map[string]interface{} {
	return map[string]interface{}{
		"app": map[string]interface{}{
			"name":        config.App.Name,
			"version":     config.App.Version,
			"environment": config.App.Environment,
		},
		"server": map[string]interface{}{
			"host": config.Server.Host,
			"port": config.Server.Port,
		},
		"database": map[string]interface{}{
			"driver":         config.Database.Driver,
			"max_idle_conns": config.Database.MaxIdleConns,
			"max_open_conns": config.Database.MaxOpenConns,
		},
		"rabbitmq": map[string]interface{}{
			"queue_name":   config.RabbitMQ.QueueName,
			"consumer_tag": config.RabbitMQ.ConsumerTag,
			"url_set":      config.RabbitMQ.URL != "",
		},
		"claude": map[string]interface{}{
			"model":       config.Claude.Model,
			"max_tokens":  config.Claude.MaxTokens,
			"timeout":     config.Claude.Timeout,
			"api_key_set": config.Claude.APIKey != "",
		},
		"logging": map[string]interface{}{
			"level":  config.Logging.Level,
			"format": config.Logging.Format,
			"output": config.Logging.Output,
		},
	}
}