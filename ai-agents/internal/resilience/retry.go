package resilience

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"go-coffee-ai-agents/internal/errors"
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxAttempts     int           `yaml:"max_attempts"`
	InitialDelay    time.Duration `yaml:"initial_delay"`
	MaxDelay        time.Duration `yaml:"max_delay"`
	BackoffFactor   float64       `yaml:"backoff_factor"`
	Jitter          bool          `yaml:"jitter"`
	RetryableErrors []string      `yaml:"retryable_errors"`
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
		RetryableErrors: []string{
			string(errors.ErrorTypeNetwork),
			string(errors.ErrorTypeTimeout),
			string(errors.ErrorTypeRateLimit),
			string(errors.ErrorTypeExternalAPI),
			string(errors.ErrorTypeKafka),
		},
	}
}

// RetryableFunc represents a function that can be retried
type RetryableFunc func() error

// RetryableFuncWithResult represents a function that returns a result and can be retried
type RetryableFuncWithResult func() (interface{}, error)

// Retrier handles retry logic with exponential backoff
type Retrier struct {
	config RetryConfig
	logger Logger
}

// Logger interface for the retrier
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, err error, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// NewRetrier creates a new retrier with the given configuration
func NewRetrier(config RetryConfig, logger Logger) *Retrier {
	return &Retrier{
		config: config,
		logger: logger,
	}
}

// Execute executes a function with retry logic
func (r *Retrier) Execute(ctx context.Context, operation string, fn RetryableFunc) error {
	var lastErr error
	
	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		// Check if context is cancelled
		if ctx.Err() != nil {
			return ctx.Err()
		}
		
		r.logger.Debug("Executing operation", 
			"operation", operation, 
			"attempt", attempt, 
			"max_attempts", r.config.MaxAttempts)
		
		err := fn()
		if err == nil {
			if attempt > 1 {
				r.logger.Info("Operation succeeded after retry", 
					"operation", operation, 
					"attempt", attempt)
			}
			return nil
		}
		
		lastErr = err
		
		// Check if error is retryable
		if !r.isRetryable(err) {
			r.logger.Debug("Error is not retryable", 
				"operation", operation, 
				"error", err.Error())
			return err
		}
		
		// Don't sleep after the last attempt
		if attempt == r.config.MaxAttempts {
			break
		}
		
		delay := r.calculateDelay(attempt)
		
		r.logger.Warn("Operation failed, retrying", 
			"operation", operation, 
			"attempt", attempt, 
			"error", err.Error(), 
			"retry_delay", delay)
		
		// Wait for the calculated delay or until context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}
	
	r.logger.Error("Operation failed after all retry attempts", lastErr, 
		"operation", operation, 
		"max_attempts", r.config.MaxAttempts)
	
	return fmt.Errorf("operation '%s' failed after %d attempts: %w", 
		operation, r.config.MaxAttempts, lastErr)
}

// ExecuteWithResult executes a function with retry logic and returns a result
func (r *Retrier) ExecuteWithResult(ctx context.Context, operation string, fn RetryableFuncWithResult) (interface{}, error) {
	var lastErr error
	
	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		// Check if context is cancelled
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		
		r.logger.Debug("Executing operation with result", 
			"operation", operation, 
			"attempt", attempt, 
			"max_attempts", r.config.MaxAttempts)
		
		res, err := fn()
		if err == nil {
			if attempt > 1 {
				r.logger.Info("Operation succeeded after retry", 
					"operation", operation, 
					"attempt", attempt)
			}
			return res, nil
		}
		
		lastErr = err
		
		// Check if error is retryable
		if !r.isRetryable(err) {
			r.logger.Debug("Error is not retryable", 
				"operation", operation, 
				"error", err.Error())
			return nil, err
		}
		
		// Don't sleep after the last attempt
		if attempt == r.config.MaxAttempts {
			break
		}
		
		delay := r.calculateDelay(attempt)
		
		r.logger.Warn("Operation failed, retrying", 
			"operation", operation, 
			"attempt", attempt, 
			"error", err.Error(), 
			"retry_delay", delay)
		
		// Wait for the calculated delay or until context is cancelled
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}
	
	r.logger.Error("Operation failed after all retry attempts", lastErr, 
		"operation", operation, 
		"max_attempts", r.config.MaxAttempts)
	
	return nil, fmt.Errorf("operation '%s' failed after %d attempts: %w", 
		operation, r.config.MaxAttempts, lastErr)
}

// isRetryable checks if an error is retryable
func (r *Retrier) isRetryable(err error) bool {
	// Check if it's an AppError with retryable flag
	if appErr, ok := err.(*errors.AppError); ok {
		return appErr.Retryable
	}
	
	// Check against configured retryable error types
	for _, retryableType := range r.config.RetryableErrors {
		if appErr, ok := err.(*errors.AppError); ok {
			if string(appErr.Type) == retryableType {
				return true
			}
		}
	}
	
	return false
}

// calculateDelay calculates the delay for the next retry attempt
func (r *Retrier) calculateDelay(attempt int) time.Duration {
	// Calculate exponential backoff
	delay := float64(r.config.InitialDelay) * math.Pow(r.config.BackoffFactor, float64(attempt-1))
	
	// Apply maximum delay limit
	if delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}
	
	// Add jitter if enabled
	if r.config.Jitter {
		jitter := rand.Float64() * 0.1 * delay // 10% jitter
		delay += jitter
	}
	
	return time.Duration(delay)
}

// RetryPolicy defines different retry policies for different operations
type RetryPolicy struct {
	Name   string
	Config RetryConfig
}

// PredefinedPolicies contains common retry policies
var PredefinedPolicies = map[string]RetryPolicy{
	"fast": {
		Name: "fast",
		Config: RetryConfig{
			MaxAttempts:   3,
			InitialDelay:  50 * time.Millisecond,
			MaxDelay:      1 * time.Second,
			BackoffFactor: 1.5,
			Jitter:        true,
		},
	},
	"standard": {
		Name: "standard",
		Config: RetryConfig{
			MaxAttempts:   5,
			InitialDelay:  100 * time.Millisecond,
			MaxDelay:      10 * time.Second,
			BackoffFactor: 2.0,
			Jitter:        true,
		},
	},
	"slow": {
		Name: "slow",
		Config: RetryConfig{
			MaxAttempts:   7,
			InitialDelay:  500 * time.Millisecond,
			MaxDelay:      60 * time.Second,
			BackoffFactor: 2.5,
			Jitter:        true,
		},
	},
	"external_api": {
		Name: "external_api",
		Config: RetryConfig{
			MaxAttempts:   4,
			InitialDelay:  200 * time.Millisecond,
			MaxDelay:      30 * time.Second,
			BackoffFactor: 2.0,
			Jitter:        true,
			RetryableErrors: []string{
				string(errors.ErrorTypeNetwork),
				string(errors.ErrorTypeTimeout),
				string(errors.ErrorTypeRateLimit),
				string(errors.ErrorTypeExternalAPI),
			},
		},
	},
	"database": {
		Name: "database",
		Config: RetryConfig{
			MaxAttempts:   3,
			InitialDelay:  100 * time.Millisecond,
			MaxDelay:      5 * time.Second,
			BackoffFactor: 2.0,
			Jitter:        false, // No jitter for database operations
			RetryableErrors: []string{
				string(errors.ErrorTypeDatabase),
				string(errors.ErrorTypeNetwork),
				string(errors.ErrorTypeTimeout),
			},
		},
	},
	"kafka": {
		Name: "kafka",
		Config: RetryConfig{
			MaxAttempts:   5,
			InitialDelay:  100 * time.Millisecond,
			MaxDelay:      15 * time.Second,
			BackoffFactor: 2.0,
			Jitter:        true,
			RetryableErrors: []string{
				string(errors.ErrorTypeKafka),
				string(errors.ErrorTypeNetwork),
				string(errors.ErrorTypeTimeout),
			},
		},
	},
}

// GetRetrier returns a retrier for a specific policy
func GetRetrier(policyName string, logger Logger) *Retrier {
	policy, exists := PredefinedPolicies[policyName]
	if !exists {
		policy = PredefinedPolicies["standard"]
	}
	
	return NewRetrier(policy.Config, logger)
}

// RetryWithPolicy executes a function with a specific retry policy
func RetryWithPolicy(ctx context.Context, policyName, operation string, fn RetryableFunc, logger Logger) error {
	retrier := GetRetrier(policyName, logger)
	return retrier.Execute(ctx, operation, fn)
}

// RetryWithPolicyAndResult executes a function with a specific retry policy and returns a result
func RetryWithPolicyAndResult(ctx context.Context, policyName, operation string, fn RetryableFuncWithResult, logger Logger) (interface{}, error) {
	retrier := GetRetrier(policyName, logger)
	return retrier.ExecuteWithResult(ctx, operation, fn)
}
