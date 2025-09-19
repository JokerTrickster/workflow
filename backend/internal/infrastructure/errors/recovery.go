package errors

import (
	"context"
	"fmt"
	"math"
	"time"

	"local-backend-server/internal/infrastructure/logging"
)

// RetryConfig defines retry behavior configuration
type RetryConfig struct {
	MaxAttempts  int           `json:"max_attempts"`
	InitialDelay time.Duration `json:"initial_delay"`
	MaxDelay     time.Duration `json:"max_delay"`
	BackoffRate  float64       `json:"backoff_rate"`
	Jitter       bool          `json:"jitter"`
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		BackoffRate:  2.0,
		Jitter:       true,
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func(ctx context.Context) error

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	CircuitClosed CircuitBreakerState = iota
	CircuitOpen
	CircuitHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name           string
	maxFailures    int
	timeout        time.Duration
	failures       int
	lastFailTime   time.Time
	state          CircuitBreakerState
	logger         *logging.Logger
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, maxFailures int, timeout time.Duration, logger *logging.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		name:        name,
		maxFailures: maxFailures,
		timeout:     timeout,
		state:       CircuitClosed,
		logger:      logger,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn RetryableFunc) error {
	// Check if circuit is open
	if cb.state == CircuitOpen {
		if time.Since(cb.lastFailTime) < cb.timeout {
			return NewWithDetails(ErrCodeInternal, 
				fmt.Sprintf("Circuit breaker %s is open", cb.name),
				"Service is temporarily unavailable")
		}
		// Transition to half-open
		cb.state = CircuitHalfOpen
		cb.logger.WithComponent("circuit_breaker").
			WithField("circuit", cb.name).
			Info("Circuit breaker transitioning to half-open")
	}

	// Execute the function
	err := fn(ctx)
	
	// Handle result
	if err != nil {
		cb.recordFailure()
		return err
	}
	
	cb.recordSuccess()
	return nil
}

// recordFailure records a failure and potentially opens the circuit
func (cb *CircuitBreaker) recordFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()
	
	if cb.failures >= cb.maxFailures {
		if cb.state != CircuitOpen {
			cb.state = CircuitOpen
			cb.logger.WithComponent("circuit_breaker").
				WithField("circuit", cb.name).
				WithField("failures", cb.failures).
				Warn("Circuit breaker opened")
		}
	}
}

// recordSuccess records a success and potentially closes the circuit
func (cb *CircuitBreaker) recordSuccess() {
	cb.failures = 0
	
	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
		cb.logger.WithComponent("circuit_breaker").
			WithField("circuit", cb.name).
			Info("Circuit breaker closed")
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	return cb.state
}

// RecoveryManager manages error recovery strategies
type RecoveryManager struct {
	logger         *logging.Logger
	circuitBreakers map[string]*CircuitBreaker
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager(logger *logging.Logger) *RecoveryManager {
	return &RecoveryManager{
		logger:          logger,
		circuitBreakers: make(map[string]*CircuitBreaker),
	}
}

// AddCircuitBreaker adds a circuit breaker for a specific service
func (rm *RecoveryManager) AddCircuitBreaker(name string, maxFailures int, timeout time.Duration) {
	rm.circuitBreakers[name] = NewCircuitBreaker(name, maxFailures, timeout, rm.logger)
}

// ExecuteWithRetry executes a function with retry logic
func (rm *RecoveryManager) ExecuteWithRetry(ctx context.Context, name string, fn RetryableFunc, config RetryConfig) error {
	var lastErr error
	
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Execute with circuit breaker if available
		if cb, exists := rm.circuitBreakers[name]; exists {
			lastErr = cb.Execute(ctx, fn)
		} else {
			lastErr = fn(ctx)
		}
		
		// Success
		if lastErr == nil {
			if attempt > 1 {
				rm.logger.WithComponent("retry").
					WithField("operation", name).
					WithField("attempt", attempt).
					Info("Operation succeeded after retry")
			}
			return nil
		}
		
		// Check if error is retryable
		if !IsRetryable(lastErr) {
			rm.logger.WithComponent("retry").
				WithField("operation", name).
				WithError(lastErr).
				Debug("Operation failed with non-retryable error")
			return lastErr
		}
		
		// Don't sleep after the last attempt
		if attempt == config.MaxAttempts {
			break
		}
		
		// Calculate delay with exponential backoff
		delay := rm.calculateDelay(attempt, config)
		
		rm.logger.WithComponent("retry").
			WithField("operation", name).
			WithField("attempt", attempt).
			WithField("delay_ms", delay.Milliseconds()).
			WithError(lastErr).
			Warn("Operation failed, retrying")
		
		// Wait for the delay or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}
	
	// All attempts failed
	rm.logger.WithComponent("retry").
		WithField("operation", name).
		WithField("max_attempts", config.MaxAttempts).
		WithError(lastErr).
		Error("Operation failed after all retry attempts")
	
	return Wrap(lastErr, ErrCodeInternal, 
		fmt.Sprintf("Operation %s failed after %d attempts", name, config.MaxAttempts))
}

// calculateDelay calculates the delay for the next retry attempt
func (rm *RecoveryManager) calculateDelay(attempt int, config RetryConfig) time.Duration {
	// Calculate exponential backoff
	delay := time.Duration(float64(config.InitialDelay) * math.Pow(config.BackoffRate, float64(attempt-1)))
	
	// Cap at max delay
	if delay > config.MaxDelay {
		delay = config.MaxDelay
	}
	
	// Add jitter if enabled
	if config.Jitter {
		jitterPercent := (float64(time.Now().UnixNano()%100) / 100.0) * 0.2 - 0.1  // ±10% jitter
		jitter := time.Duration(float64(delay) * jitterPercent)
		delay += jitter
	}
	
	return delay
}

// ExecuteWithTimeout executes a function with timeout
func (rm *RecoveryManager) ExecuteWithTimeout(ctx context.Context, timeout time.Duration, fn RetryableFunc) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	done := make(chan error, 1)
	
	go func() {
		done <- fn(ctx)
	}()
	
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return TimeoutError("operation", timeout)
		}
		return ctx.Err()
	}
}

// ExecuteWithRecovery executes a function with full recovery (retry + circuit breaker + timeout)
func (rm *RecoveryManager) ExecuteWithRecovery(ctx context.Context, name string, fn RetryableFunc, timeout time.Duration, retryConfig RetryConfig) error {
	timeoutFn := func(ctx context.Context) error {
		return rm.ExecuteWithTimeout(ctx, timeout, fn)
	}
	
	return rm.ExecuteWithRetry(ctx, name, timeoutFn, retryConfig)
}

// GetCircuitBreakerState returns the state of a circuit breaker
func (rm *RecoveryManager) GetCircuitBreakerState(name string) (CircuitBreakerState, bool) {
	if cb, exists := rm.circuitBreakers[name]; exists {
		return cb.GetState(), true
	}
	return CircuitClosed, false
}

// GetCircuitBreakerStats returns statistics for all circuit breakers
func (rm *RecoveryManager) GetCircuitBreakerStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	for name, cb := range rm.circuitBreakers {
		state := "closed"
		switch cb.state {
		case CircuitOpen:
			state = "open"
		case CircuitHalfOpen:
			state = "half-open"
		}
		
		stats[name] = map[string]interface{}{
			"state":         state,
			"failures":      cb.failures,
			"max_failures":  cb.maxFailures,
			"last_fail_time": cb.lastFailTime,
			"timeout":       cb.timeout,
		}
	}
	
	return stats
}