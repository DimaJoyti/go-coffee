package resilience

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go-coffee-ai-agents/internal/errors"
)

// ResilienceConfig configures all resilience patterns
type ResilienceConfig struct {
	Retry         map[string]RetryConfig         `yaml:"retry"`
	CircuitBreaker map[string]CircuitBreakerConfig `yaml:"circuit_breaker"`
	RateLimit     map[string]RateLimiterConfig   `yaml:"rate_limit"`
	Timeout       TimeoutConfig                  `yaml:"timeout"`
}

// DefaultResilienceConfig returns a default resilience configuration
func DefaultResilienceConfig() ResilienceConfig {
	return ResilienceConfig{
		Retry: map[string]RetryConfig{
			"default":      DefaultRetryConfig(),
			"fast":         PredefinedPolicies["fast"].Config,
			"standard":     PredefinedPolicies["standard"].Config,
			"slow":         PredefinedPolicies["slow"].Config,
			"external_api": PredefinedPolicies["external_api"].Config,
			"database":     PredefinedPolicies["database"].Config,
			"kafka":        PredefinedPolicies["kafka"].Config,
		},
		CircuitBreaker: map[string]CircuitBreakerConfig{
			"default":      DefaultCircuitBreakerConfig("default"),
			"ai_provider":  DefaultCircuitBreakerConfig("ai_provider"),
			"external_api": DefaultCircuitBreakerConfig("external_api"),
			"database":     DefaultCircuitBreakerConfig("database"),
			"kafka":        DefaultCircuitBreakerConfig("kafka"),
		},
		RateLimit: PredefinedRateLimiters,
		Timeout:   DefaultTimeoutConfig(),
	}
}

// ResilienceManager coordinates all resilience patterns
type ResilienceManager struct {
	config               ResilienceConfig
	retriers             map[string]*Retrier
	circuitBreakerManager *CircuitBreakerManager
	rateLimiterManager   *RateLimiterManager
	timeoutManager       *TimeoutManager
	logger               Logger
	mutex                sync.RWMutex
}

// NewResilienceManager creates a new resilience manager
func NewResilienceManager(config ResilienceConfig, logger Logger) *ResilienceManager {
	rm := &ResilienceManager{
		config:               config,
		retriers:             make(map[string]*Retrier),
		circuitBreakerManager: NewCircuitBreakerManager(logger),
		rateLimiterManager:   NewRateLimiterManager(logger),
		timeoutManager:       NewTimeoutManager(config.Timeout, logger),
		logger:               logger,
	}
	
	// Initialize retriers
	for name, retryConfig := range config.Retry {
		rm.retriers[name] = NewRetrier(retryConfig, logger)
	}
	
	return rm
}

// ExecuteWithResilience executes a function with all resilience patterns applied
func (rm *ResilienceManager) ExecuteWithResilience(ctx context.Context, operation string, fn func(context.Context) error) error {
	return rm.ExecuteWithResilienceConfig(ctx, operation, "default", fn)
}

// ExecuteWithResilienceConfig executes a function with specific resilience configuration
func (rm *ResilienceManager) ExecuteWithResilienceConfig(ctx context.Context, operation, configName string, fn func(context.Context) error) error {
	// Get rate limiter
	rateLimiter := rm.getRateLimiter(configName)
	
	// Check rate limit
	if !rateLimiter.Allow() {
		rm.logger.Warn("Request rate limited", "operation", operation, "config", configName)
		return errors.NewRateLimitError(operation, 0)
	}
	
	// Get circuit breaker
	circuitBreaker := rm.getCircuitBreaker(configName)
	
	// Execute with circuit breaker
	return circuitBreaker.Execute(ctx, func() error {
		// Get retrier
		retrier := rm.getRetrier(configName)
		
		// Execute with retry and timeout
		return retrier.Execute(ctx, operation, func() error {
			return rm.timeoutManager.WithTimeout(ctx, operation, fn)
		})
	})
}

// ExecuteWithResilienceAndResult executes a function with resilience patterns and returns a result
func (rm *ResilienceManager) ExecuteWithResilienceAndResult(ctx context.Context, operation, configName string, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	// Get rate limiter
	rateLimiter := rm.getRateLimiter(configName)
	
	// Check rate limit
	if !rateLimiter.Allow() {
		rm.logger.Warn("Request rate limited", "operation", operation, "config", configName)
		return nil, errors.NewRateLimitError(operation, 0)
	}
	
	// Get circuit breaker
	circuitBreaker := rm.getCircuitBreaker(configName)
	
	// Execute with circuit breaker
	return circuitBreaker.ExecuteWithResult(ctx, func() (interface{}, error) {
		// Get retrier
		retrier := rm.getRetrier(configName)
		
		// Execute with retry and timeout
		return retrier.ExecuteWithResult(ctx, operation, func() (interface{}, error) {
			return rm.timeoutManager.WithTimeoutAndResult(ctx, operation, fn)
		})
	})
}

// ExecuteWithCustomResilience executes a function with custom resilience settings
func (rm *ResilienceManager) ExecuteWithCustomResilience(ctx context.Context, settings ResilienceSettings, fn func(context.Context) error) error {
	// Apply rate limiting if configured
	if settings.RateLimit != nil {
		rateLimiter := rm.rateLimiterManager.GetOrCreate(settings.Name, *settings.RateLimit)
		if !rateLimiter.Allow() {
			return errors.NewRateLimitError(settings.Name, 0)
		}
	}
	
	// Apply circuit breaker if configured
	if settings.CircuitBreaker != nil {
		circuitBreaker := rm.circuitBreakerManager.GetOrCreate(settings.Name, settings.CircuitBreaker)
		return circuitBreaker.Execute(ctx, func() error {
			return rm.executeWithRetryAndTimeout(ctx, settings, fn)
		})
	}
	
	return rm.executeWithRetryAndTimeout(ctx, settings, fn)
}

// executeWithRetryAndTimeout executes a function with retry and timeout
func (rm *ResilienceManager) executeWithRetryAndTimeout(ctx context.Context, settings ResilienceSettings, fn func(context.Context) error) error {
	// Apply retry if configured
	if settings.Retry != nil {
		retrier := NewRetrier(*settings.Retry, rm.logger)
		return retrier.Execute(ctx, settings.Name, func() error {
			return rm.executeWithTimeout(ctx, settings, fn)
		})
	}
	
	return rm.executeWithTimeout(ctx, settings, fn)
}

// executeWithTimeout executes a function with timeout
func (rm *ResilienceManager) executeWithTimeout(ctx context.Context, settings ResilienceSettings, fn func(context.Context) error) error {
	if settings.Timeout != nil {
		timeoutCtx, cancel := context.WithTimeout(ctx, *settings.Timeout)
		defer cancel()
		return fn(timeoutCtx)
	}
	
	return fn(ctx)
}

// getRetrier gets a retrier by name
func (rm *ResilienceManager) getRetrier(name string) *Retrier {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	if retrier, exists := rm.retriers[name]; exists {
		return retrier
	}
	
	// Return default retrier if not found
	return rm.retriers["default"]
}

// getCircuitBreaker gets a circuit breaker by name
func (rm *ResilienceManager) getCircuitBreaker(name string) *CircuitBreaker {
	config, exists := rm.config.CircuitBreaker[name]
	if !exists {
		config = rm.config.CircuitBreaker["default"]
	}
	
	return rm.circuitBreakerManager.GetOrCreate(name, &config)
}

// getRateLimiter gets a rate limiter by name
func (rm *ResilienceManager) getRateLimiter(name string) RateLimiter {
	config, exists := rm.config.RateLimit[name]
	if !exists {
		config = PredefinedRateLimiters["external_api"] // Default
	}
	
	return rm.rateLimiterManager.GetOrCreate(name, config)
}

// ResilienceSettings defines custom resilience settings
type ResilienceSettings struct {
	Name           string                   `json:"name"`
	Retry          *RetryConfig             `json:"retry,omitempty"`
	CircuitBreaker *CircuitBreakerConfig    `json:"circuit_breaker,omitempty"`
	RateLimit      *RateLimiterConfig       `json:"rate_limit,omitempty"`
	Timeout        *time.Duration           `json:"timeout,omitempty"`
}

// GetMetrics returns comprehensive resilience metrics
func (rm *ResilienceManager) GetMetrics() ResilienceMetrics {
	return ResilienceMetrics{
		CircuitBreakers: rm.circuitBreakerManager.GetMetrics(),
		RateLimiters:    rm.rateLimiterManager.GetMetrics(),
		Timeouts:        rm.timeoutManager.GetTimeouts(),
	}
}

// ResilienceMetrics contains metrics for all resilience patterns
type ResilienceMetrics struct {
	CircuitBreakers map[string]CircuitBreakerMetrics `json:"circuit_breakers"`
	RateLimiters    map[string]RateLimiterMetrics    `json:"rate_limiters"`
	Timeouts        TimeoutConfig                    `json:"timeouts"`
}

// HealthCheck performs a health check on all resilience components
func (rm *ResilienceManager) HealthCheck() map[string]interface{} {
	health := make(map[string]interface{})
	
	// Circuit breaker health
	cbHealth := make(map[string]string)
	for name, cb := range rm.circuitBreakerManager.GetAll() {
		switch cb.GetState() {
		case StateClosed:
			cbHealth[name] = "healthy"
		case StateHalfOpen:
			cbHealth[name] = "recovering"
		case StateOpen:
			cbHealth[name] = "unhealthy"
		}
	}
	health["circuit_breakers"] = cbHealth
	
	// Rate limiter health
	rlHealth := make(map[string]interface{})
	for name, metrics := range rm.rateLimiterManager.GetMetrics() {
		rlHealth[name] = map[string]interface{}{
			"total_requests":    metrics.TotalRequests,
			"allowed_requests":  metrics.AllowedRequests,
			"rejected_requests": metrics.RejectedRequests,
			"current_tokens":    metrics.CurrentTokens,
		}
	}
	health["rate_limiters"] = rlHealth
	
	return health
}

// Global resilience manager instance
var globalResilienceManager *ResilienceManager
var resilienceManagerOnce sync.Once

// GetGlobalResilienceManager returns the global resilience manager
func GetGlobalResilienceManager(logger Logger) *ResilienceManager {
	resilienceManagerOnce.Do(func() {
		globalResilienceManager = NewResilienceManager(DefaultResilienceConfig(), logger)
	})
	return globalResilienceManager
}

// Convenience functions for common operations

// ExecuteWithStandardResilience executes a function with standard resilience patterns
func ExecuteWithStandardResilience(ctx context.Context, operation string, fn func(context.Context) error, logger Logger) error {
	rm := GetGlobalResilienceManager(logger)
	return rm.ExecuteWithResilience(ctx, operation, fn)
}

// ExecuteAIOperation executes an AI operation with appropriate resilience
func ExecuteAIOperation(ctx context.Context, provider string, fn func(context.Context) error, logger Logger) error {
	rm := GetGlobalResilienceManager(logger)
	return rm.ExecuteWithResilienceConfig(ctx, fmt.Sprintf("ai_%s", provider), "ai_provider", fn)
}

// ExecuteExternalAPICall executes an external API call with appropriate resilience
func ExecuteExternalAPICall(ctx context.Context, service string, fn func(context.Context) error, logger Logger) error {
	rm := GetGlobalResilienceManager(logger)
	return rm.ExecuteWithResilienceConfig(ctx, fmt.Sprintf("api_%s", service), "external_api", fn)
}

// ExecuteDatabaseOperation executes a database operation with appropriate resilience
func ExecuteDatabaseOperation(ctx context.Context, operation string, fn func(context.Context) error, logger Logger) error {
	rm := GetGlobalResilienceManager(logger)
	return rm.ExecuteWithResilienceConfig(ctx, fmt.Sprintf("db_%s", operation), "database", fn)
}

// ExecuteKafkaOperation executes a Kafka operation with appropriate resilience
func ExecuteKafkaOperation(ctx context.Context, operation string, fn func(context.Context) error, logger Logger) error {
	rm := GetGlobalResilienceManager(logger)
	return rm.ExecuteWithResilienceConfig(ctx, fmt.Sprintf("kafka_%s", operation), "kafka", fn)
}
