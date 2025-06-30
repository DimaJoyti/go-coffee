package circuit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go-coffee-ai-agents/internal/resilience/errors"
)

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// String returns the string representation of the state
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Config holds circuit breaker configuration
type Config struct {
	// Name of the circuit breaker for identification
	Name string
	
	// Maximum number of failures before opening the circuit
	FailureThreshold int
	
	// Number of successful calls required to close the circuit from half-open
	SuccessThreshold int
	
	// Duration to wait before transitioning from open to half-open
	Timeout time.Duration
	
	// Maximum number of concurrent requests allowed in half-open state
	MaxConcurrentRequests int
	
	// Function to determine if an error should count as a failure
	IsFailure func(error) bool
	
	// Callback functions for state changes
	OnStateChange func(name string, from, to State)
	OnFailure     func(name string, err error)
	OnSuccess     func(name string, duration time.Duration)
}

// DefaultConfig returns a default circuit breaker configuration
func DefaultConfig(name string) Config {
	return Config{
		Name:                  name,
		FailureThreshold:      5,
		SuccessThreshold:      3,
		Timeout:               60 * time.Second,
		MaxConcurrentRequests: 10,
		IsFailure: func(err error) bool {
			// By default, all errors are considered failures
			return err != nil
		},
	}
}

// Metrics holds circuit breaker metrics
type Metrics struct {
	// Request counts
	TotalRequests    int64
	SuccessfulReqs   int64
	FailedRequests   int64
	
	// Consecutive counts
	ConsecutiveSuccesses int64
	ConsecutiveFailures  int64
	
	// Timing
	LastFailureTime time.Time
	LastSuccessTime time.Time
	
	// State transitions
	StateTransitions int64
	LastStateChange  time.Time
	
	// Performance
	AverageLatency   time.Duration
	MaxLatency       time.Duration
	MinLatency       time.Duration
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config  Config
	state   State
	metrics Metrics
	
	// Synchronization
	mutex sync.RWMutex
	
	// State management
	nextRetryTime time.Time
	
	// Half-open state management
	halfOpenRequests int64
	
	// Latency tracking
	latencySum   time.Duration
	latencyCount int64
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config Config) *CircuitBreaker {
	if config.FailureThreshold <= 0 {
		config.FailureThreshold = 5
	}
	if config.SuccessThreshold <= 0 {
		config.SuccessThreshold = 3
	}
	if config.Timeout <= 0 {
		config.Timeout = 60 * time.Second
	}
	if config.MaxConcurrentRequests <= 0 {
		config.MaxConcurrentRequests = 10
	}
	if config.IsFailure == nil {
		config.IsFailure = func(err error) bool { return err != nil }
	}

	return &CircuitBreaker{
		config:  config,
		state:   StateClosed,
		metrics: Metrics{},
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	// Check if we can execute the request
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	start := time.Now()
	err := fn()
	duration := time.Since(start)

	// Record the result
	cb.afterRequest(err, duration)

	return err
}

// ExecuteWithFallback executes a function with circuit breaker protection and fallback
func (cb *CircuitBreaker) ExecuteWithFallback(ctx context.Context, fn func() error, fallback func() error) error {
	err := cb.Execute(ctx, fn)
	
	// If circuit breaker is open or the function failed, try fallback
	if cb.IsOpen() || (err != nil && cb.config.IsFailure(err)) {
		if fallback != nil {
			return fallback()
		}
	}
	
	return err
}

// beforeRequest checks if the request can be executed
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()

	switch cb.state {
	case StateClosed:
		// Allow request in closed state
		return nil

	case StateOpen:
		// Check if we should transition to half-open
		if now.After(cb.nextRetryTime) {
			cb.setState(StateHalfOpen)
			cb.halfOpenRequests = 0
			return nil
		}
		
		// Circuit is still open
		return errors.NewCircuitBreakerError(
			cb.state.String(),
			int(cb.metrics.ConsecutiveFailures),
		).WithDetail("next_retry_time", cb.nextRetryTime)

	case StateHalfOpen:
		// Check if we've reached the maximum concurrent requests
		if cb.halfOpenRequests >= int64(cb.config.MaxConcurrentRequests) {
			return errors.NewCircuitBreakerError(
				cb.state.String(),
				int(cb.metrics.ConsecutiveFailures),
			).WithDetail("concurrent_requests", cb.halfOpenRequests)
		}
		
		cb.halfOpenRequests++
		return nil

	default:
		return errors.NewError(errors.CodeInternalError, "Unknown circuit breaker state").Build()
	}
}

// afterRequest records the result of a request
func (cb *CircuitBreaker) afterRequest(err error, duration time.Duration) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// Update metrics
	cb.metrics.TotalRequests++
	cb.updateLatencyMetrics(duration)

	isFailure := cb.config.IsFailure(err)

	if isFailure {
		cb.onFailure(err)
	} else {
		cb.onSuccess(duration)
	}

	// Handle state transitions based on current state
	switch cb.state {
	case StateClosed:
		if isFailure {
			if cb.metrics.ConsecutiveFailures >= int64(cb.config.FailureThreshold) {
				cb.setState(StateOpen)
				cb.nextRetryTime = time.Now().Add(cb.config.Timeout)
			}
		}

	case StateHalfOpen:
		cb.halfOpenRequests--
		
		if isFailure {
			// Any failure in half-open state opens the circuit
			cb.setState(StateOpen)
			cb.nextRetryTime = time.Now().Add(cb.config.Timeout)
		} else {
			// Check if we have enough successes to close the circuit
			if cb.metrics.ConsecutiveSuccesses >= int64(cb.config.SuccessThreshold) {
				cb.setState(StateClosed)
			}
		}

	case StateOpen:
		// Requests shouldn't reach here, but handle it gracefully
		if !isFailure {
			cb.setState(StateHalfOpen)
			cb.halfOpenRequests = 0
		}
	}
}

// onFailure handles a failed request
func (cb *CircuitBreaker) onFailure(err error) {
	cb.metrics.FailedRequests++
	cb.metrics.ConsecutiveFailures++
	cb.metrics.ConsecutiveSuccesses = 0
	cb.metrics.LastFailureTime = time.Now()

	if cb.config.OnFailure != nil {
		cb.config.OnFailure(cb.config.Name, err)
	}
}

// onSuccess handles a successful request
func (cb *CircuitBreaker) onSuccess(duration time.Duration) {
	cb.metrics.SuccessfulReqs++
	cb.metrics.ConsecutiveSuccesses++
	cb.metrics.ConsecutiveFailures = 0
	cb.metrics.LastSuccessTime = time.Now()

	if cb.config.OnSuccess != nil {
		cb.config.OnSuccess(cb.config.Name, duration)
	}
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(newState State) {
	if cb.state != newState {
		oldState := cb.state
		cb.state = newState
		cb.metrics.StateTransitions++
		cb.metrics.LastStateChange = time.Now()

		if cb.config.OnStateChange != nil {
			cb.config.OnStateChange(cb.config.Name, oldState, newState)
		}
	}
}

// updateLatencyMetrics updates latency tracking
func (cb *CircuitBreaker) updateLatencyMetrics(duration time.Duration) {
	cb.latencySum += duration
	cb.latencyCount++
	cb.metrics.AverageLatency = cb.latencySum / time.Duration(cb.latencyCount)

	if cb.metrics.MaxLatency == 0 || duration > cb.metrics.MaxLatency {
		cb.metrics.MaxLatency = duration
	}

	if cb.metrics.MinLatency == 0 || duration < cb.metrics.MinLatency {
		cb.metrics.MinLatency = duration
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() State {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// IsOpen returns true if the circuit breaker is open
func (cb *CircuitBreaker) IsOpen() bool {
	return cb.State() == StateOpen
}

// IsClosed returns true if the circuit breaker is closed
func (cb *CircuitBreaker) IsClosed() bool {
	return cb.State() == StateClosed
}

// IsHalfOpen returns true if the circuit breaker is half-open
func (cb *CircuitBreaker) IsHalfOpen() bool {
	return cb.State() == StateHalfOpen
}

// Metrics returns a copy of the current metrics
func (cb *CircuitBreaker) Metrics() Metrics {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.metrics
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.setState(StateClosed)
	cb.metrics.ConsecutiveFailures = 0
	cb.metrics.ConsecutiveSuccesses = 0
	cb.halfOpenRequests = 0
}

// ForceOpen forces the circuit breaker to open state
func (cb *CircuitBreaker) ForceOpen() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.setState(StateOpen)
	cb.nextRetryTime = time.Now().Add(cb.config.Timeout)
}

// String returns a string representation of the circuit breaker
func (cb *CircuitBreaker) String() string {
	metrics := cb.Metrics()
	return fmt.Sprintf("CircuitBreaker{name=%s, state=%s, failures=%d/%d, successes=%d/%d}",
		cb.config.Name,
		cb.State().String(),
		metrics.ConsecutiveFailures,
		cb.config.FailureThreshold,
		metrics.ConsecutiveSuccesses,
		cb.config.SuccessThreshold,
	)
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mutex    sync.RWMutex
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// GetOrCreate gets an existing circuit breaker or creates a new one
func (cbm *CircuitBreakerManager) GetOrCreate(name string, config Config) *CircuitBreaker {
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	if cb, exists := cbm.breakers[name]; exists {
		return cb
	}

	config.Name = name
	cb := NewCircuitBreaker(config)
	cbm.breakers[name] = cb
	return cb
}

// Get gets an existing circuit breaker
func (cbm *CircuitBreakerManager) Get(name string) (*CircuitBreaker, bool) {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	cb, exists := cbm.breakers[name]
	return cb, exists
}

// List returns all circuit breaker names
func (cbm *CircuitBreakerManager) List() []string {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	names := make([]string, 0, len(cbm.breakers))
	for name := range cbm.breakers {
		names = append(names, name)
	}
	return names
}

// Remove removes a circuit breaker
func (cbm *CircuitBreakerManager) Remove(name string) {
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	delete(cbm.breakers, name)
}

// ResetAll resets all circuit breakers
func (cbm *CircuitBreakerManager) ResetAll() {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	for _, cb := range cbm.breakers {
		cb.Reset()
	}
}

// Global circuit breaker manager
var globalManager = NewCircuitBreakerManager()

// GetGlobalManager returns the global circuit breaker manager
func GetGlobalManager() *CircuitBreakerManager {
	return globalManager
}

// GetOrCreateGlobal gets or creates a circuit breaker from the global manager
func GetOrCreateGlobal(name string, config Config) *CircuitBreaker {
	return globalManager.GetOrCreate(name, config)
}
