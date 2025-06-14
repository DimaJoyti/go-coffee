package concurrency

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int32

const (
	StateClosed CircuitBreakerState = iota
	StateHalfOpen
	StateOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateHalfOpen:
		return "HALF_OPEN"
	case StateOpen:
		return "OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker provides circuit breaker functionality
type CircuitBreaker struct {
	name   string
	logger *zap.Logger
	config *CircuitBreakerConfig

	// State management
	state          int32 // CircuitBreakerState
	stateChangedAt int64 // Unix timestamp

	// Counters
	requests  int64
	failures  int64
	successes int64
	timeouts  int64

	// Half-open state management
	halfOpenRequests  int64
	halfOpenSuccesses int64

	// Fallback function
	fallbackFunc func(context.Context, error) (interface{}, error)

	// Metrics
	metrics *CircuitBreakerMetrics

	mu sync.RWMutex
}

// CircuitBreakerConfig contains circuit breaker configuration
type CircuitBreakerConfig struct {
	FailureThreshold     int64         `json:"failure_threshold"`       // Number of failures to open circuit
	SuccessThreshold     int64         `json:"success_threshold"`       // Number of successes to close circuit
	TimeoutThreshold     time.Duration `json:"timeout_threshold"`       // Request timeout threshold
	OpenTimeout          time.Duration `json:"open_timeout"`            // Time to wait before trying half-open
	HalfOpenTimeout      time.Duration `json:"half_open_timeout"`       // Time to wait in half-open state
	HalfOpenMaxRequests  int64         `json:"half_open_max_requests"`  // Max requests in half-open state
	HalfOpenSuccessRatio float64       `json:"half_open_success_ratio"` // Success ratio to close circuit
	ResetTimeout         time.Duration `json:"reset_timeout"`           // Time to reset counters
	MonitoringInterval   time.Duration `json:"monitoring_interval"`     // Metrics collection interval
}

// CircuitBreakerMetrics tracks circuit breaker performance
type CircuitBreakerMetrics struct {
	State                CircuitBreakerState `json:"state"`
	TotalRequests        int64               `json:"total_requests"`
	TotalFailures        int64               `json:"total_failures"`
	TotalSuccesses       int64               `json:"total_successes"`
	TotalTimeouts        int64               `json:"total_timeouts"`
	FailureRate          float64             `json:"failure_rate"`
	SuccessRate          float64             `json:"success_rate"`
	TimeoutRate          float64             `json:"timeout_rate"`
	StateChangedAt       time.Time           `json:"state_changed_at"`
	TimeSinceStateChange time.Duration       `json:"time_since_state_change"`
	HalfOpenRequests     int64               `json:"half_open_requests"`
	HalfOpenSuccesses    int64               `json:"half_open_successes"`
}

// CircuitBreakerFunc represents a function that can be protected by circuit breaker
type CircuitBreakerFunc func(context.Context) (interface{}, error)

// FallbackFunc represents a fallback function
type FallbackFunc func(context.Context, error) (interface{}, error)

var (
	ErrCircuitBreakerOpen    = errors.New("circuit breaker is open")
	ErrCircuitBreakerTimeout = errors.New("circuit breaker timeout")
	ErrHalfOpenLimitExceeded = errors.New("half-open request limit exceeded")
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, config *CircuitBreakerConfig, logger *zap.Logger) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:           name,
		logger:         logger,
		config:         config,
		state:          int32(StateClosed),
		stateChangedAt: time.Now().Unix(),
		metrics:        &CircuitBreakerMetrics{},
	}

	// Start monitoring goroutine
	go cb.monitor()

	return cb
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn CircuitBreakerFunc) (interface{}, error) {
	// Check if circuit breaker allows the request
	if !cb.allowRequest() {
		cb.recordFailure()

		// Try fallback if available
		if cb.fallbackFunc != nil {
			cb.logger.Debug("Circuit breaker open, executing fallback",
				zap.String("circuit_breaker", cb.name))
			return cb.fallbackFunc(ctx, ErrCircuitBreakerOpen)
		}

		return nil, ErrCircuitBreakerOpen
	}

	// Execute the function with timeout
	return cb.executeWithTimeout(ctx, fn)
}

// SetFallback sets a fallback function
func (cb *CircuitBreaker) SetFallback(fallback FallbackFunc) {
	cb.fallbackFunc = fallback
}

// allowRequest checks if a request should be allowed
func (cb *CircuitBreaker) allowRequest() bool {
	state := CircuitBreakerState(atomic.LoadInt32(&cb.state))
	now := time.Now()

	switch state {
	case StateClosed:
		return true

	case StateOpen:
		stateChangedAt := time.Unix(atomic.LoadInt64(&cb.stateChangedAt), 0)
		if now.Sub(stateChangedAt) >= cb.config.OpenTimeout {
			// Try to transition to half-open
			if cb.transitionToHalfOpen() {
				return true
			}
		}
		return false

	case StateHalfOpen:
		halfOpenRequests := atomic.LoadInt64(&cb.halfOpenRequests)
		if halfOpenRequests < cb.config.HalfOpenMaxRequests {
			return true
		}
		return false

	default:
		return false
	}
}

// executeWithTimeout executes function with timeout protection
func (cb *CircuitBreaker) executeWithTimeout(ctx context.Context, fn CircuitBreakerFunc) (interface{}, error) {
	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, cb.config.TimeoutThreshold)
	defer cancel()

	// Channel to receive result
	resultChan := make(chan struct {
		result interface{}
		err    error
	}, 1)

	// Execute function in goroutine
	go func() {
		result, err := fn(timeoutCtx)
		resultChan <- struct {
			result interface{}
			err    error
		}{result, err}
	}()

	// Wait for result or timeout
	select {
	case res := <-resultChan:
		if res.err != nil {
			cb.recordFailure()
			return res.result, res.err
		}
		cb.recordSuccess()
		return res.result, nil

	case <-timeoutCtx.Done():
		cb.recordTimeout()

		// Try fallback if available
		if cb.fallbackFunc != nil {
			cb.logger.Debug("Request timeout, executing fallback",
				zap.String("circuit_breaker", cb.name))
			return cb.fallbackFunc(ctx, ErrCircuitBreakerTimeout)
		}

		return nil, ErrCircuitBreakerTimeout
	}
}

// recordSuccess records a successful request
func (cb *CircuitBreaker) recordSuccess() {
	atomic.AddInt64(&cb.requests, 1)
	atomic.AddInt64(&cb.successes, 1)

	state := CircuitBreakerState(atomic.LoadInt32(&cb.state))

	if state == StateHalfOpen {
		halfOpenSuccesses := atomic.AddInt64(&cb.halfOpenSuccesses, 1)
		halfOpenRequests := atomic.AddInt64(&cb.halfOpenRequests, 1)

		// Check if we should close the circuit
		successRatio := float64(halfOpenSuccesses) / float64(halfOpenRequests)
		if halfOpenRequests >= cb.config.SuccessThreshold && successRatio >= cb.config.HalfOpenSuccessRatio {
			cb.transitionToClosed()
		}
	}
}

// recordFailure records a failed request
func (cb *CircuitBreaker) recordFailure() {
	atomic.AddInt64(&cb.requests, 1)
	atomic.AddInt64(&cb.failures, 1)

	state := CircuitBreakerState(atomic.LoadInt32(&cb.state))

	if state == StateClosed {
		failures := atomic.LoadInt64(&cb.failures)
		if failures >= cb.config.FailureThreshold {
			cb.transitionToOpen()
		}
	} else if state == StateHalfOpen {
		atomic.AddInt64(&cb.halfOpenRequests, 1)
		cb.transitionToOpen()
	}
}

// recordTimeout records a timeout
func (cb *CircuitBreaker) recordTimeout() {
	atomic.AddInt64(&cb.requests, 1)
	atomic.AddInt64(&cb.timeouts, 1)

	// Treat timeout as failure
	cb.recordFailure()
}

// transitionToOpen transitions circuit breaker to open state
func (cb *CircuitBreaker) transitionToOpen() {
	if atomic.CompareAndSwapInt32(&cb.state, int32(StateClosed), int32(StateOpen)) ||
		atomic.CompareAndSwapInt32(&cb.state, int32(StateHalfOpen), int32(StateOpen)) {

		atomic.StoreInt64(&cb.stateChangedAt, time.Now().Unix())
		atomic.StoreInt64(&cb.halfOpenRequests, 0)
		atomic.StoreInt64(&cb.halfOpenSuccesses, 0)

		cb.logger.Warn("Circuit breaker opened",
			zap.String("circuit_breaker", cb.name),
			zap.Int64("failures", atomic.LoadInt64(&cb.failures)),
			zap.Int64("threshold", cb.config.FailureThreshold))
	}
}

// transitionToHalfOpen transitions circuit breaker to half-open state
func (cb *CircuitBreaker) transitionToHalfOpen() bool {
	if atomic.CompareAndSwapInt32(&cb.state, int32(StateOpen), int32(StateHalfOpen)) {
		atomic.StoreInt64(&cb.stateChangedAt, time.Now().Unix())
		atomic.StoreInt64(&cb.halfOpenRequests, 0)
		atomic.StoreInt64(&cb.halfOpenSuccesses, 0)

		cb.logger.Info("Circuit breaker transitioned to half-open",
			zap.String("circuit_breaker", cb.name))
		return true
	}
	return false
}

// transitionToClosed transitions circuit breaker to closed state
func (cb *CircuitBreaker) transitionToClosed() {
	if atomic.CompareAndSwapInt32(&cb.state, int32(StateHalfOpen), int32(StateClosed)) {
		atomic.StoreInt64(&cb.stateChangedAt, time.Now().Unix())

		// Reset counters
		atomic.StoreInt64(&cb.requests, 0)
		atomic.StoreInt64(&cb.failures, 0)
		atomic.StoreInt64(&cb.successes, 0)
		atomic.StoreInt64(&cb.timeouts, 0)
		atomic.StoreInt64(&cb.halfOpenRequests, 0)
		atomic.StoreInt64(&cb.halfOpenSuccesses, 0)

		cb.logger.Info("Circuit breaker closed",
			zap.String("circuit_breaker", cb.name))
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	return CircuitBreakerState(atomic.LoadInt32(&cb.state))
}

// GetMetrics returns current circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() *CircuitBreakerMetrics {
	requests := atomic.LoadInt64(&cb.requests)
	failures := atomic.LoadInt64(&cb.failures)
	successes := atomic.LoadInt64(&cb.successes)
	timeouts := atomic.LoadInt64(&cb.timeouts)
	stateChangedAt := time.Unix(atomic.LoadInt64(&cb.stateChangedAt), 0)

	var failureRate, successRate, timeoutRate float64
	if requests > 0 {
		failureRate = float64(failures) / float64(requests)
		successRate = float64(successes) / float64(requests)
		timeoutRate = float64(timeouts) / float64(requests)
	}

	return &CircuitBreakerMetrics{
		State:                cb.GetState(),
		TotalRequests:        requests,
		TotalFailures:        failures,
		TotalSuccesses:       successes,
		TotalTimeouts:        timeouts,
		FailureRate:          failureRate,
		SuccessRate:          successRate,
		TimeoutRate:          timeoutRate,
		StateChangedAt:       stateChangedAt,
		TimeSinceStateChange: time.Since(stateChangedAt),
		HalfOpenRequests:     atomic.LoadInt64(&cb.halfOpenRequests),
		HalfOpenSuccesses:    atomic.LoadInt64(&cb.halfOpenSuccesses),
	}
}

// monitor runs periodic monitoring and cleanup
func (cb *CircuitBreaker) monitor() {
	ticker := time.NewTicker(cb.config.MonitoringInterval)
	defer ticker.Stop()

	for range ticker.C {
		cb.performMaintenance()
	}
}

// performMaintenance performs periodic maintenance tasks
func (cb *CircuitBreaker) performMaintenance() {
	now := time.Now()
	stateChangedAt := time.Unix(atomic.LoadInt64(&cb.stateChangedAt), 0)

	// Reset counters if enough time has passed
	if now.Sub(stateChangedAt) >= cb.config.ResetTimeout {
		state := cb.GetState()
		if state == StateClosed {
			atomic.StoreInt64(&cb.requests, 0)
			atomic.StoreInt64(&cb.failures, 0)
			atomic.StoreInt64(&cb.successes, 0)
			atomic.StoreInt64(&cb.timeouts, 0)
		}
	}

	// Log current metrics
	metrics := cb.GetMetrics()
	cb.logger.Debug("Circuit breaker metrics",
		zap.String("circuit_breaker", cb.name),
		zap.String("state", metrics.State.String()),
		zap.Int64("requests", metrics.TotalRequests),
		zap.Int64("failures", metrics.TotalFailures),
		zap.Float64("failure_rate", metrics.FailureRate),
		zap.Duration("time_since_state_change", metrics.TimeSinceStateChange))
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	atomic.StoreInt32(&cb.state, int32(StateClosed))
	atomic.StoreInt64(&cb.stateChangedAt, time.Now().Unix())
	atomic.StoreInt64(&cb.requests, 0)
	atomic.StoreInt64(&cb.failures, 0)
	atomic.StoreInt64(&cb.successes, 0)
	atomic.StoreInt64(&cb.timeouts, 0)
	atomic.StoreInt64(&cb.halfOpenRequests, 0)
	atomic.StoreInt64(&cb.halfOpenSuccesses, 0)

	cb.logger.Info("Circuit breaker reset", zap.String("circuit_breaker", cb.name))
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex
	logger   *zap.Logger
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(logger *zap.Logger) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
		logger:   logger,
	}
}

// GetOrCreate gets an existing circuit breaker or creates a new one
func (cbm *CircuitBreakerManager) GetOrCreate(name string, config *CircuitBreakerConfig) *CircuitBreaker {
	cbm.mu.RLock()
	if cb, exists := cbm.breakers[name]; exists {
		cbm.mu.RUnlock()
		return cb
	}
	cbm.mu.RUnlock()

	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	// Double-check after acquiring write lock
	if cb, exists := cbm.breakers[name]; exists {
		return cb
	}

	cb := NewCircuitBreaker(name, config, cbm.logger)
	cbm.breakers[name] = cb

	cbm.logger.Info("Created new circuit breaker",
		zap.String("name", name),
		zap.Int64("failure_threshold", config.FailureThreshold))

	return cb
}

// Get gets an existing circuit breaker
func (cbm *CircuitBreakerManager) Get(name string) (*CircuitBreaker, bool) {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	cb, exists := cbm.breakers[name]
	return cb, exists
}

// GetAll returns all circuit breakers
func (cbm *CircuitBreakerManager) GetAll() map[string]*CircuitBreaker {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	result := make(map[string]*CircuitBreaker)
	for name, cb := range cbm.breakers {
		result[name] = cb
	}

	return result
}

// GetMetrics returns metrics for all circuit breakers
func (cbm *CircuitBreakerManager) GetMetrics() map[string]*CircuitBreakerMetrics {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	result := make(map[string]*CircuitBreakerMetrics)
	for name, cb := range cbm.breakers {
		result[name] = cb.GetMetrics()
	}

	return result
}
