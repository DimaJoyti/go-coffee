package timeout

import (
	"context"
	"sync"
	"time"

	"go-coffee-ai-agents/internal/resilience/errors"
)

// TimeoutConfig holds timeout configuration for different operations
type TimeoutConfig struct {
	// Default timeout for operations
	Default time.Duration
	
	// Operation-specific timeouts
	Operations map[string]time.Duration
	
	// Component-specific timeouts
	Components map[string]time.Duration
	
	// Grace period for cleanup operations
	GracePeriod time.Duration
	
	// Whether to cascade timeouts to child operations
	CascadeTimeouts bool
}

// DefaultTimeoutConfig returns a default timeout configuration
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Default:     30 * time.Second,
		Operations:  make(map[string]time.Duration),
		Components:  make(map[string]time.Duration),
		GracePeriod: 5 * time.Second,
		CascadeTimeouts: true,
	}
}

// Manager manages timeouts for operations
type Manager struct {
	config TimeoutConfig
	mutex  sync.RWMutex
}

// NewManager creates a new timeout manager
func NewManager(config TimeoutConfig) *Manager {
	if config.Default <= 0 {
		config.Default = 30 * time.Second
	}
	if config.GracePeriod <= 0 {
		config.GracePeriod = 5 * time.Second
	}
	if config.Operations == nil {
		config.Operations = make(map[string]time.Duration)
	}
	if config.Components == nil {
		config.Components = make(map[string]time.Duration)
	}

	return &Manager{
		config: config,
	}
}

// GetTimeout returns the timeout for a specific operation and component
func (tm *Manager) GetTimeout(operation, component string) time.Duration {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	// Check for operation-specific timeout first
	if timeout, exists := tm.config.Operations[operation]; exists {
		return timeout
	}

	// Check for component-specific timeout
	if timeout, exists := tm.config.Components[component]; exists {
		return timeout
	}

	// Return default timeout
	return tm.config.Default
}

// SetOperationTimeout sets a timeout for a specific operation
func (tm *Manager) SetOperationTimeout(operation string, timeout time.Duration) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.config.Operations[operation] = timeout
}

// SetComponentTimeout sets a timeout for a specific component
func (tm *Manager) SetComponentTimeout(component string, timeout time.Duration) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.config.Components[component] = timeout
}

// WithTimeout creates a context with timeout for an operation
func (tm *Manager) WithTimeout(ctx context.Context, operation, component string) (context.Context, context.CancelFunc) {
	timeout := tm.GetTimeout(operation, component)
	return context.WithTimeout(ctx, timeout)
}

// WithDeadline creates a context with deadline for an operation
func (tm *Manager) WithDeadline(ctx context.Context, operation, component string, deadline time.Time) (context.Context, context.CancelFunc) {
	timeout := tm.GetTimeout(operation, component)
	calculatedDeadline := time.Now().Add(timeout)
	
	// Use the earlier deadline
	if deadline.Before(calculatedDeadline) {
		return context.WithDeadline(ctx, deadline)
	}
	return context.WithDeadline(ctx, calculatedDeadline)
}

// Execute executes a function with timeout management
func (tm *Manager) Execute(ctx context.Context, operation, component string, fn func(context.Context) error) error {
	timeoutCtx, cancel := tm.WithTimeout(ctx, operation, component)
	defer cancel()

	start := time.Now()
	done := make(chan error, 1)

	// Execute function in goroutine
	go func() {
		done <- fn(timeoutCtx)
	}()

	// Wait for completion or timeout
	select {
	case err := <-done:
		return err
	case <-timeoutCtx.Done():
		elapsed := time.Since(start)
		timeout := tm.GetTimeout(operation, component)
		
		if timeoutCtx.Err() == context.DeadlineExceeded {
			timeoutErr := errors.NewTimeoutError(timeout, elapsed)
			// Update the context with operation and component information
			ctx := timeoutErr.Context
			ctx.Operation = operation
			ctx.Component = component
			return timeoutErr.WithContext(ctx)
		}
		return timeoutCtx.Err()
	}
}

// ExecuteWithResult executes a function with timeout management and returns a result
func (tm *Manager) ExecuteWithResult(ctx context.Context, operation, component string, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	timeoutCtx, cancel := tm.WithTimeout(ctx, operation, component)
	defer cancel()

	start := time.Now()
	type result struct {
		value interface{}
		err   error
	}
	done := make(chan result, 1)

	// Execute function in goroutine
	go func() {
		value, err := fn(timeoutCtx)
		done <- result{value: value, err: err}
	}()

	// Wait for completion or timeout
	select {
	case res := <-done:
		return res.value, res.err
	case <-timeoutCtx.Done():
		elapsed := time.Since(start)
		timeout := tm.GetTimeout(operation, component)
		
		if timeoutCtx.Err() == context.DeadlineExceeded {
			timeoutErr := errors.NewTimeoutError(timeout, elapsed)
			// Update the context with operation and component information
			ctx := timeoutErr.Context
			ctx.Operation = operation
			ctx.Component = component
			return nil, timeoutErr.WithContext(ctx)
		}
		return nil, timeoutCtx.Err()
	}
}

// ExecuteWithGracefulShutdown executes a function with graceful shutdown support
func (tm *Manager) ExecuteWithGracefulShutdown(ctx context.Context, operation, component string, fn func(context.Context) error, cleanup func() error) error {
	timeoutCtx, cancel := tm.WithTimeout(ctx, operation, component)
	defer cancel()

	start := time.Now()
	done := make(chan error, 1)

	// Execute function in goroutine
	go func() {
		done <- fn(timeoutCtx)
	}()

	// Wait for completion or timeout
	select {
	case err := <-done:
		return err
	case <-timeoutCtx.Done():
		// Start graceful shutdown
		if cleanup != nil {
			cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), tm.config.GracePeriod)
			defer cleanupCancel()
			
			// Execute cleanup with timeout
			done := make(chan error, 1)
			go func() {
				done <- cleanup()
			}()
			
			select {
			case cleanupErr := <-done:
				if cleanupErr != nil {
					// Log cleanup error but don't override the timeout error
					// In a real implementation, you'd use a logger here
				}
			case <-cleanupCtx.Done():
				// Cleanup timed out, but continue with the original timeout error
				// In a real implementation, you'd log this timeout
			}
		}

		elapsed := time.Since(start)
		timeout := tm.GetTimeout(operation, component)
		
		if timeoutCtx.Err() == context.DeadlineExceeded {
			timeoutErr := errors.NewTimeoutError(timeout, elapsed)
			// Update the context with operation and component information
			ctx := timeoutErr.Context
			ctx.Operation = operation
			ctx.Component = component
			return timeoutErr.WithContext(ctx)
		}
		return timeoutCtx.Err()
	}
}

// TimeoutPolicy defines timeout behavior for different scenarios
type TimeoutPolicy struct {
	// Base timeout for the operation
	BaseTimeout time.Duration
	
	// Scaling factor based on load or other metrics
	ScalingFactor float64
	
	// Maximum allowed timeout
	MaxTimeout time.Duration
	
	// Minimum allowed timeout
	MinTimeout time.Duration
	
	// Whether to use adaptive timeouts based on historical performance
	Adaptive bool
	
	// Historical performance data for adaptive timeouts
	HistoricalData *PerformanceData
}

// PerformanceData tracks historical performance for adaptive timeouts
type PerformanceData struct {
	AverageLatency time.Duration
	P95Latency     time.Duration
	P99Latency     time.Duration
	SuccessRate    float64
	SampleCount    int64
	LastUpdated    time.Time
	mutex          sync.RWMutex
}

// NewPerformanceData creates new performance data tracker
func NewPerformanceData() *PerformanceData {
	return &PerformanceData{
		LastUpdated: time.Now(),
	}
}

// UpdateLatency updates the latency metrics
func (pd *PerformanceData) UpdateLatency(latency time.Duration, success bool) {
	pd.mutex.Lock()
	defer pd.mutex.Unlock()

	pd.SampleCount++
	
	// Simple moving average (in production, you'd use a more sophisticated approach)
	if pd.AverageLatency == 0 {
		pd.AverageLatency = latency
	} else {
		pd.AverageLatency = (pd.AverageLatency + latency) / 2
	}

	// Update percentiles (simplified - in production, use proper percentile tracking)
	if latency > pd.P95Latency {
		pd.P95Latency = latency
	}
	if latency > pd.P99Latency {
		pd.P99Latency = latency
	}

	// Update success rate
	if success {
		pd.SuccessRate = (pd.SuccessRate*float64(pd.SampleCount-1) + 1.0) / float64(pd.SampleCount)
	} else {
		pd.SuccessRate = (pd.SuccessRate * float64(pd.SampleCount-1)) / float64(pd.SampleCount)
	}

	pd.LastUpdated = time.Now()
}

// GetAdaptiveTimeout calculates an adaptive timeout based on performance data
func (pd *PerformanceData) GetAdaptiveTimeout(baseTimeout time.Duration) time.Duration {
	pd.mutex.RLock()
	defer pd.mutex.RUnlock()

	if pd.SampleCount < 10 {
		// Not enough data, use base timeout
		return baseTimeout
	}

	// Use P99 latency as base with some buffer
	adaptiveTimeout := time.Duration(float64(pd.P99Latency) * 1.5)

	// Adjust based on success rate
	if pd.SuccessRate < 0.95 {
		// Low success rate, increase timeout
		adaptiveTimeout = time.Duration(float64(adaptiveTimeout) * 1.2)
	}

	// Ensure it's within reasonable bounds
	minTimeout := baseTimeout / 2
	maxTimeout := baseTimeout * 3

	if adaptiveTimeout < minTimeout {
		adaptiveTimeout = minTimeout
	}
	if adaptiveTimeout > maxTimeout {
		adaptiveTimeout = maxTimeout
	}

	return adaptiveTimeout
}

// AdaptiveTimeoutManager manages adaptive timeouts based on performance
type AdaptiveTimeoutManager struct {
	*Manager
	policies map[string]*TimeoutPolicy
	mutex    sync.RWMutex
}

// NewAdaptiveTimeoutManager creates a new adaptive timeout manager
func NewAdaptiveTimeoutManager(config TimeoutConfig) *AdaptiveTimeoutManager {
	return &AdaptiveTimeoutManager{
		Manager:  NewManager(config),
		policies: make(map[string]*TimeoutPolicy),
	}
}

// SetPolicy sets a timeout policy for an operation
func (atm *AdaptiveTimeoutManager) SetPolicy(operation string, policy TimeoutPolicy) {
	atm.mutex.Lock()
	defer atm.mutex.Unlock()

	if policy.HistoricalData == nil {
		policy.HistoricalData = NewPerformanceData()
	}

	atm.policies[operation] = &policy
}

// GetAdaptiveTimeout returns an adaptive timeout for an operation
func (atm *AdaptiveTimeoutManager) GetAdaptiveTimeout(operation, component string) time.Duration {
	atm.mutex.RLock()
	policy, exists := atm.policies[operation]
	atm.mutex.RUnlock()

	if !exists || !policy.Adaptive {
		return atm.GetTimeout(operation, component)
	}

	baseTimeout := policy.BaseTimeout
	if baseTimeout == 0 {
		baseTimeout = atm.GetTimeout(operation, component)
	}

	adaptiveTimeout := policy.HistoricalData.GetAdaptiveTimeout(baseTimeout)

	// Apply scaling factor
	if policy.ScalingFactor > 0 {
		adaptiveTimeout = time.Duration(float64(adaptiveTimeout) * policy.ScalingFactor)
	}

	// Apply bounds
	if policy.MinTimeout > 0 && adaptiveTimeout < policy.MinTimeout {
		adaptiveTimeout = policy.MinTimeout
	}
	if policy.MaxTimeout > 0 && adaptiveTimeout > policy.MaxTimeout {
		adaptiveTimeout = policy.MaxTimeout
	}

	return adaptiveTimeout
}

// ExecuteWithAdaptiveTimeout executes a function with adaptive timeout
func (atm *AdaptiveTimeoutManager) ExecuteWithAdaptiveTimeout(ctx context.Context, operation, component string, fn func(context.Context) error) error {
	timeout := atm.GetAdaptiveTimeout(operation, component)
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	err := fn(timeoutCtx)
	elapsed := time.Since(start)

	// Update performance data
	atm.mutex.RLock()
	policy, exists := atm.policies[operation]
	atm.mutex.RUnlock()

	if exists && policy.Adaptive && policy.HistoricalData != nil {
		success := err == nil
		policy.HistoricalData.UpdateLatency(elapsed, success)
	}

	return err
}

// Global timeout manager
var globalManager *Manager

// InitGlobalManager initializes the global timeout manager
func InitGlobalManager(config TimeoutConfig) {
	globalManager = NewManager(config)
}

// GetGlobalManager returns the global timeout manager
func GetGlobalManager() *Manager {
	if globalManager == nil {
		globalManager = NewManager(DefaultTimeoutConfig())
	}
	return globalManager
}

// Convenience functions using the global manager

// WithTimeout creates a context with timeout using the global manager
func WithTimeout(ctx context.Context, operation, component string) (context.Context, context.CancelFunc) {
	return GetGlobalManager().WithTimeout(ctx, operation, component)
}

// Execute executes a function with timeout using the global manager
func Execute(ctx context.Context, operation, component string, fn func(context.Context) error) error {
	return GetGlobalManager().Execute(ctx, operation, component, fn)
}

// ExecuteWithResult executes a function with timeout and result using the global manager
func ExecuteWithResult(ctx context.Context, operation, component string, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	return GetGlobalManager().ExecuteWithResult(ctx, operation, component, fn)
}
