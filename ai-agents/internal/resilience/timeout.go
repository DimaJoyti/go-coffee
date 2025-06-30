package resilience

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go-coffee-ai-agents/internal/errors"
)

// TimeoutConfig configures timeout behavior for different operations
type TimeoutConfig struct {
	Default     time.Duration            `yaml:"default"`
	Operations  map[string]time.Duration `yaml:"operations"`
	Services    map[string]time.Duration `yaml:"services"`
	Endpoints   map[string]time.Duration `yaml:"endpoints"`
}

// DefaultTimeoutConfig returns a default timeout configuration
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Default: 30 * time.Second,
		Operations: map[string]time.Duration{
			"database_query":    5 * time.Second,
			"database_write":    10 * time.Second,
			"kafka_publish":     3 * time.Second,
			"kafka_consume":     1 * time.Second,
			"ai_generation":     60 * time.Second,
			"external_api":      15 * time.Second,
			"file_upload":       120 * time.Second,
			"image_processing":  30 * time.Second,
		},
		Services: map[string]time.Duration{
			"clickup":       10 * time.Second,
			"slack":         5 * time.Second,
			"google_sheets": 15 * time.Second,
			"gemini":        45 * time.Second,
			"openai":        30 * time.Second,
			"ollama":        60 * time.Second,
		},
		Endpoints: map[string]time.Duration{
			"/health":           1 * time.Second,
			"/metrics":          2 * time.Second,
			"/api/beverages":    10 * time.Second,
			"/api/tasks":        5 * time.Second,
			"/api/notifications": 3 * time.Second,
		},
	}
}

// TimeoutManager manages timeouts for different operations
type TimeoutManager struct {
	config TimeoutConfig
	logger Logger
	mutex  sync.RWMutex
}

// NewTimeoutManager creates a new timeout manager
func NewTimeoutManager(config TimeoutConfig, logger Logger) *TimeoutManager {
	return &TimeoutManager{
		config: config,
		logger: logger,
	}
}

// WithTimeout executes a function with a timeout
func (tm *TimeoutManager) WithTimeout(ctx context.Context, operation string, fn func(context.Context) error) error {
	timeout := tm.getTimeout(operation)
	
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	tm.logger.Debug("Executing operation with timeout",
		"operation", operation,
		"timeout", timeout)
	
	done := make(chan error, 1)
	
	go func() {
		done <- fn(timeoutCtx)
	}()
	
	select {
	case err := <-done:
		if err != nil {
			tm.logger.Debug("Operation completed with error",
				"operation", operation,
				"error", err.Error())
		} else {
			tm.logger.Debug("Operation completed successfully",
				"operation", operation)
		}
		return err
		
	case <-timeoutCtx.Done():
		tm.logger.Warn("Operation timed out",
			"operation", operation,
			"timeout", timeout)
		return errors.NewTimeoutError(operation, timeout)
	}
}

// WithTimeoutAndResult executes a function with a timeout and returns a result
func (tm *TimeoutManager) WithTimeoutAndResult(ctx context.Context, operation string, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	timeout := tm.getTimeout(operation)
	
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	tm.logger.Debug("Executing operation with timeout and result",
		"operation", operation,
		"timeout", timeout)
	
	type resultError struct {
		result interface{}
		err    error
	}
	
	done := make(chan resultError, 1)
	
	go func() {
		res, err := fn(timeoutCtx)
		done <- resultError{result: res, err: err}
	}()
	
	select {
	case re := <-done:
		if re.err != nil {
			tm.logger.Debug("Operation completed with error",
				"operation", operation,
				"error", re.err.Error())
		} else {
			tm.logger.Debug("Operation completed successfully",
				"operation", operation)
		}
		return re.result, re.err
		
	case <-timeoutCtx.Done():
		tm.logger.Warn("Operation timed out",
			"operation", operation,
			"timeout", timeout)
		return nil, errors.NewTimeoutError(operation, timeout)
	}
}

// WithServiceTimeout executes a function with a service-specific timeout
func (tm *TimeoutManager) WithServiceTimeout(ctx context.Context, service string, fn func(context.Context) error) error {
	timeout := tm.getServiceTimeout(service)
	
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	tm.logger.Debug("Executing service operation with timeout",
		"service", service,
		"timeout", timeout)
	
	done := make(chan error, 1)
	
	go func() {
		done <- fn(timeoutCtx)
	}()
	
	select {
	case err := <-done:
		return err
	case <-timeoutCtx.Done():
		tm.logger.Warn("Service operation timed out",
			"service", service,
			"timeout", timeout)
		return errors.NewTimeoutError(fmt.Sprintf("service_%s", service), timeout)
	}
}

// WithEndpointTimeout executes a function with an endpoint-specific timeout
func (tm *TimeoutManager) WithEndpointTimeout(ctx context.Context, endpoint string, fn func(context.Context) error) error {
	timeout := tm.getEndpointTimeout(endpoint)
	
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	tm.logger.Debug("Executing endpoint operation with timeout",
		"endpoint", endpoint,
		"timeout", timeout)
	
	done := make(chan error, 1)
	
	go func() {
		done <- fn(timeoutCtx)
	}()
	
	select {
	case err := <-done:
		return err
	case <-timeoutCtx.Done():
		tm.logger.Warn("Endpoint operation timed out",
			"endpoint", endpoint,
			"timeout", timeout)
		return errors.NewTimeoutError(fmt.Sprintf("endpoint_%s", endpoint), timeout)
	}
}

// getTimeout returns the timeout for a specific operation
func (tm *TimeoutManager) getTimeout(operation string) time.Duration {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	if timeout, exists := tm.config.Operations[operation]; exists {
		return timeout
	}
	
	return tm.config.Default
}

// getServiceTimeout returns the timeout for a specific service
func (tm *TimeoutManager) getServiceTimeout(service string) time.Duration {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	if timeout, exists := tm.config.Services[service]; exists {
		return timeout
	}
	
	return tm.config.Default
}

// getEndpointTimeout returns the timeout for a specific endpoint
func (tm *TimeoutManager) getEndpointTimeout(endpoint string) time.Duration {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	if timeout, exists := tm.config.Endpoints[endpoint]; exists {
		return timeout
	}
	
	return tm.config.Default
}

// UpdateOperationTimeout updates the timeout for a specific operation
func (tm *TimeoutManager) UpdateOperationTimeout(operation string, timeout time.Duration) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	tm.config.Operations[operation] = timeout
	
	tm.logger.Info("Updated operation timeout",
		"operation", operation,
		"timeout", timeout)
}

// UpdateServiceTimeout updates the timeout for a specific service
func (tm *TimeoutManager) UpdateServiceTimeout(service string, timeout time.Duration) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	tm.config.Services[service] = timeout
	
	tm.logger.Info("Updated service timeout",
		"service", service,
		"timeout", timeout)
}

// GetTimeouts returns all configured timeouts
func (tm *TimeoutManager) GetTimeouts() TimeoutConfig {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	return tm.config
}

// TimeoutDecorator decorates functions with timeout behavior
type TimeoutDecorator struct {
	manager *TimeoutManager
}

// NewTimeoutDecorator creates a new timeout decorator
func NewTimeoutDecorator(manager *TimeoutManager) *TimeoutDecorator {
	return &TimeoutDecorator{
		manager: manager,
	}
}

// DecorateFunc decorates a function with timeout behavior
func (td *TimeoutDecorator) DecorateFunc(operation string, fn func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) error {
		return td.manager.WithTimeout(ctx, operation, fn)
	}
}

// DecorateFuncWithResult decorates a function with timeout behavior and result
func (td *TimeoutDecorator) DecorateFuncWithResult(operation string, fn func(context.Context) (interface{}, error)) func(context.Context) (interface{}, error) {
	return func(ctx context.Context) (interface{}, error) {
		return td.manager.WithTimeoutAndResult(ctx, operation, fn)
	}
}

// DecorateServiceFunc decorates a service function with timeout behavior
func (td *TimeoutDecorator) DecorateServiceFunc(service string, fn func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) error {
		return td.manager.WithServiceTimeout(ctx, service, fn)
	}
}

// Global timeout manager instance
var globalTimeoutManager *TimeoutManager
var timeoutManagerOnce sync.Once

// GetGlobalTimeoutManager returns the global timeout manager
func GetGlobalTimeoutManager(logger Logger) *TimeoutManager {
	timeoutManagerOnce.Do(func() {
		globalTimeoutManager = NewTimeoutManager(DefaultTimeoutConfig(), logger)
	})
	return globalTimeoutManager
}

// Convenience functions for common timeout operations

// WithDatabaseTimeout executes a database operation with timeout
func WithDatabaseTimeout(ctx context.Context, operation string, fn func(context.Context) error, logger Logger) error {
	tm := GetGlobalTimeoutManager(logger)
	return tm.WithTimeout(ctx, fmt.Sprintf("database_%s", operation), fn)
}

// WithKafkaTimeout executes a Kafka operation with timeout
func WithKafkaTimeout(ctx context.Context, operation string, fn func(context.Context) error, logger Logger) error {
	tm := GetGlobalTimeoutManager(logger)
	return tm.WithTimeout(ctx, fmt.Sprintf("kafka_%s", operation), fn)
}

// WithAITimeout executes an AI operation with timeout
func WithAITimeout(ctx context.Context, provider string, fn func(context.Context) error, logger Logger) error {
	tm := GetGlobalTimeoutManager(logger)
	return tm.WithServiceTimeout(ctx, provider, fn)
}

// WithExternalAPITimeout executes an external API call with timeout
func WithExternalAPITimeout(ctx context.Context, service string, fn func(context.Context) error, logger Logger) error {
	tm := GetGlobalTimeoutManager(logger)
	return tm.WithServiceTimeout(ctx, service, fn)
}
