package resilience

import (
	"context"
	"sync"
	"time"

	"go-coffee-ai-agents/internal/errors"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	StateClosed   CircuitBreakerState = "closed"
	StateOpen     CircuitBreakerState = "open"
	StateHalfOpen CircuitBreakerState = "half_open"
)

// CircuitBreakerConfig configures circuit breaker behavior
type CircuitBreakerConfig struct {
	Name                string        `yaml:"name"`
	MaxFailures         int           `yaml:"max_failures"`
	ResetTimeout        time.Duration `yaml:"reset_timeout"`
	SuccessThreshold    int           `yaml:"success_threshold"`
	FailureThreshold    float64       `yaml:"failure_threshold"`
	MinRequestThreshold int           `yaml:"min_request_threshold"`
	SlidingWindowSize   int           `yaml:"sliding_window_size"`
	HalfOpenMaxCalls    int           `yaml:"half_open_max_calls"`
}

// DefaultCircuitBreakerConfig returns a default circuit breaker configuration
func DefaultCircuitBreakerConfig(name string) CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Name:                name,
		MaxFailures:         5,
		ResetTimeout:        60 * time.Second,
		SuccessThreshold:    3,
		FailureThreshold:    0.5, // 50% failure rate
		MinRequestThreshold: 10,
		SlidingWindowSize:   100,
		HalfOpenMaxCalls:    3,
	}
}

// CircuitBreakerMetrics tracks circuit breaker metrics
type CircuitBreakerMetrics struct {
	TotalRequests    int64
	SuccessfulCalls  int64
	FailedCalls      int64
	RejectedCalls    int64
	LastFailureTime  time.Time
	LastSuccessTime  time.Time
	StateChanges     int64
	CurrentState     CircuitBreakerState
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config      CircuitBreakerConfig
	state       CircuitBreakerState
	failures    int
	successes   int
	requests    int
	lastFailure time.Time
	nextAttempt time.Time
	mutex       sync.RWMutex
	metrics     CircuitBreakerMetrics
	logger      Logger
	
	// Sliding window for failure tracking
	window []bool // true for success, false for failure
	windowIndex int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig, logger Logger) *CircuitBreaker {
	return &CircuitBreaker{
		config:  config,
		state:   StateClosed,
		window:  make([]bool, config.SlidingWindowSize),
		metrics: CircuitBreakerMetrics{
			CurrentState: StateClosed,
		},
		logger: logger,
	}
}

// Execute executes a function through the circuit breaker
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	// Check if we can execute
	if !cb.canExecute() {
		cb.recordRejection()
		return errors.NewCircuitBreakerError(cb.config.Name)
	}
	
	// Execute the function
	err := fn()
	
	// Record the result
	if err != nil {
		cb.recordFailure()
		return err
	}
	
	cb.recordSuccess()
	return nil
}

// ExecuteWithResult executes a function through the circuit breaker and returns a result
func (cb *CircuitBreaker) ExecuteWithResult(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	// Check if we can execute
	if !cb.canExecute() {
		cb.recordRejection()
		return nil, errors.NewCircuitBreakerError(cb.config.Name)
	}
	
	// Execute the function
	res, err := fn()
	
	// Record the result
	if err != nil {
		cb.recordFailure()
		return nil, err
	}
	
	cb.recordSuccess()
	return res, nil
}

// canExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) canExecute() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		return time.Now().After(cb.nextAttempt)
	case StateHalfOpen:
		return cb.requests < cb.config.HalfOpenMaxCalls
	default:
		return false
	}
}

// recordSuccess records a successful execution
func (cb *CircuitBreaker) recordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.metrics.TotalRequests++
	cb.metrics.SuccessfulCalls++
	cb.metrics.LastSuccessTime = time.Now()
	
	// Update sliding window
	cb.window[cb.windowIndex] = true
	cb.windowIndex = (cb.windowIndex + 1) % cb.config.SlidingWindowSize
	
	switch cb.state {
	case StateClosed:
		// Reset failure count on success
		cb.failures = 0
		
	case StateHalfOpen:
		cb.successes++
		cb.requests++
		
		// Check if we should close the circuit
		if cb.successes >= cb.config.SuccessThreshold {
			cb.setState(StateClosed)
			cb.reset()
		}
		
	case StateOpen:
		// Transition to half-open on first success after timeout
		cb.setState(StateHalfOpen)
		cb.successes = 1
		cb.requests = 1
	}
	
	cb.logger.Debug("Circuit breaker recorded success",
		"name", cb.config.Name,
		"state", cb.state,
		"successes", cb.successes,
		"failures", cb.failures)
}

// recordFailure records a failed execution
func (cb *CircuitBreaker) recordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.metrics.TotalRequests++
	cb.metrics.FailedCalls++
	cb.metrics.LastFailureTime = time.Now()
	cb.lastFailure = time.Now()
	
	// Update sliding window
	cb.window[cb.windowIndex] = false
	cb.windowIndex = (cb.windowIndex + 1) % cb.config.SlidingWindowSize
	
	switch cb.state {
	case StateClosed:
		cb.failures++
		
		// Check if we should open the circuit
		if cb.shouldOpen() {
			cb.setState(StateOpen)
			cb.nextAttempt = time.Now().Add(cb.config.ResetTimeout)
		}
		
	case StateHalfOpen:
		// Any failure in half-open state opens the circuit
		cb.setState(StateOpen)
		cb.nextAttempt = time.Now().Add(cb.config.ResetTimeout)
		cb.reset()
	}
	
	cb.logger.Warn("Circuit breaker recorded failure",
		"name", cb.config.Name,
		"state", cb.state,
		"successes", cb.successes,
		"failures", cb.failures)
}

// recordRejection records a rejected execution
func (cb *CircuitBreaker) recordRejection() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.metrics.RejectedCalls++
	
	cb.logger.Debug("Circuit breaker rejected call",
		"name", cb.config.Name,
		"state", cb.state)
}

// shouldOpen determines if the circuit should be opened
func (cb *CircuitBreaker) shouldOpen() bool {
	// Check simple failure count threshold
	if cb.failures >= cb.config.MaxFailures {
		return true
	}
	
	// Check failure rate in sliding window
	if cb.metrics.TotalRequests >= int64(cb.config.MinRequestThreshold) {
		failureRate := cb.calculateFailureRate()
		return failureRate >= cb.config.FailureThreshold
	}
	
	return false
}

// calculateFailureRate calculates the failure rate in the sliding window
func (cb *CircuitBreaker) calculateFailureRate() float64 {
	failures := 0
	total := 0
	
	for _, success := range cb.window {
		total++
		if !success {
			failures++
		}
	}
	
	if total == 0 {
		return 0
	}
	
	return float64(failures) / float64(total)
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(newState CircuitBreakerState) {
	if cb.state != newState {
		oldState := cb.state
		cb.state = newState
		cb.metrics.CurrentState = newState
		cb.metrics.StateChanges++
		
		cb.logger.Info("Circuit breaker state changed",
			"name", cb.config.Name,
			"old_state", oldState,
			"new_state", newState)
	}
}

// reset resets the circuit breaker counters
func (cb *CircuitBreaker) reset() {
	cb.failures = 0
	cb.successes = 0
	cb.requests = 0
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetMetrics returns the current metrics
func (cb *CircuitBreaker) GetMetrics() CircuitBreakerMetrics {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.metrics
}

// IsOpen returns true if the circuit breaker is open
func (cb *CircuitBreaker) IsOpen() bool {
	return cb.GetState() == StateOpen
}

// IsClosed returns true if the circuit breaker is closed
func (cb *CircuitBreaker) IsClosed() bool {
	return cb.GetState() == StateClosed
}

// IsHalfOpen returns true if the circuit breaker is half-open
func (cb *CircuitBreaker) IsHalfOpen() bool {
	return cb.GetState() == StateHalfOpen
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mutex    sync.RWMutex
	logger   Logger
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(logger Logger) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
		logger:   logger,
	}
}

// GetOrCreate gets an existing circuit breaker or creates a new one
func (cbm *CircuitBreakerManager) GetOrCreate(name string, config *CircuitBreakerConfig) *CircuitBreaker {
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()
	
	if cb, exists := cbm.breakers[name]; exists {
		return cb
	}
	
	var cbConfig CircuitBreakerConfig
	if config != nil {
		cbConfig = *config
	} else {
		cbConfig = DefaultCircuitBreakerConfig(name)
	}
	
	cb := NewCircuitBreaker(cbConfig, cbm.logger)
	cbm.breakers[name] = cb
	
	cbm.logger.Info("Created new circuit breaker", "name", name)
	
	return cb
}

// Get gets an existing circuit breaker
func (cbm *CircuitBreakerManager) Get(name string) (*CircuitBreaker, bool) {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()
	
	cb, exists := cbm.breakers[name]
	return cb, exists
}

// GetAll returns all circuit breakers
func (cbm *CircuitBreakerManager) GetAll() map[string]*CircuitBreaker {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()
	
	result := make(map[string]*CircuitBreaker)
	for name, cb := range cbm.breakers {
		result[name] = cb
	}
	
	return result
}

// GetMetrics returns metrics for all circuit breakers
func (cbm *CircuitBreakerManager) GetMetrics() map[string]CircuitBreakerMetrics {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()
	
	result := make(map[string]CircuitBreakerMetrics)
	for name, cb := range cbm.breakers {
		result[name] = cb.GetMetrics()
	}
	
	return result
}

// Global circuit breaker manager instance
var globalCBManager *CircuitBreakerManager
var cbManagerOnce sync.Once

// GetGlobalCircuitBreakerManager returns the global circuit breaker manager
func GetGlobalCircuitBreakerManager(logger Logger) *CircuitBreakerManager {
	cbManagerOnce.Do(func() {
		globalCBManager = NewCircuitBreakerManager(logger)
	})
	return globalCBManager
}
