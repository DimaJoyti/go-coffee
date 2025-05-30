package circuitbreaker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// State represents the circuit breaker state
type State int

const (
	// StateClosed means the circuit breaker is closed (normal operation)
	StateClosed State = iota
	// StateOpen means the circuit breaker is open (failing fast)
	StateOpen
	// StateHalfOpen means the circuit breaker is half-open (testing)
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// Config contains circuit breaker configuration
type Config struct {
	Name                string        `json:"name"`
	MaxRequests         uint32        `json:"max_requests"`          // Max requests in half-open state
	Interval            time.Duration `json:"interval"`              // Interval to clear counters
	Timeout             time.Duration `json:"timeout"`               // Timeout to switch from open to half-open
	FailureThreshold    uint32        `json:"failure_threshold"`     // Failures needed to open circuit
	SuccessThreshold    uint32        `json:"success_threshold"`     // Successes needed to close circuit
	MinRequestThreshold uint32        `json:"min_request_threshold"` // Min requests before considering failure rate
}

// DefaultConfig returns a default circuit breaker configuration
func DefaultConfig(name string) *Config {
	return &Config{
		Name:                name,
		MaxRequests:         10,
		Interval:            60 * time.Second,
		Timeout:             30 * time.Second,
		FailureThreshold:    5,
		SuccessThreshold:    3,
		MinRequestThreshold: 10,
	}
}

// Counts holds the statistics for circuit breaker
type Counts struct {
	Requests             uint32 `json:"requests"`
	TotalSuccesses       uint32 `json:"total_successes"`
	TotalFailures        uint32 `json:"total_failures"`
	ConsecutiveSuccesses uint32 `json:"consecutive_successes"`
	ConsecutiveFailures  uint32 `json:"consecutive_failures"`
}

// FailureRate returns the failure rate
func (c *Counts) FailureRate() float64 {
	if c.Requests == 0 {
		return 0.0
	}
	return float64(c.TotalFailures) / float64(c.Requests)
}

// Reset resets all counters
func (c *Counts) Reset() {
	c.Requests = 0
	c.TotalSuccesses = 0
	c.TotalFailures = 0
	c.ConsecutiveSuccesses = 0
	c.ConsecutiveFailures = 0
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name         string
	config       *Config
	state        State
	counts       Counts
	stateChanged time.Time
	mutex        sync.RWMutex
	logger       *zap.Logger
	onStateChange func(name string, from State, to State)
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *Config, logger *zap.Logger) *CircuitBreaker {
	if config == nil {
		config = DefaultConfig("default")
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &CircuitBreaker{
		name:         config.Name,
		config:       config,
		state:        StateClosed,
		stateChanged: time.Now(),
		logger:       logger.With(zap.String("circuit_breaker", config.Name)),
	}
}

// SetOnStateChange sets a callback for state changes
func (cb *CircuitBreaker) SetOnStateChange(fn func(name string, from State, to State)) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.onStateChange = fn
}

// Execute executes the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	// Check if we can proceed
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	// Execute the function
	start := time.Now()
	err := fn()
	duration := time.Since(start)

	// Record the result
	cb.afterRequest(err, duration)

	return err
}

// ExecuteWithFallback executes the function with a fallback
func (cb *CircuitBreaker) ExecuteWithFallback(ctx context.Context, fn func() error, fallback func() error) error {
	err := cb.Execute(ctx, fn)
	if err != nil && cb.IsOpen() {
		cb.logger.Info("Circuit breaker is open, executing fallback")
		return fallback()
	}
	return err
}

// beforeRequest checks if the request can proceed
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, counts := cb.currentState(now)

	if state == StateOpen {
		return fmt.Errorf("circuit breaker '%s' is open", cb.name)
	}

	if state == StateHalfOpen && counts.Requests >= cb.config.MaxRequests {
		return fmt.Errorf("circuit breaker '%s' is half-open and max requests exceeded", cb.name)
	}

	cb.counts.Requests++
	return nil
}

// afterRequest records the result of a request
func (cb *CircuitBreaker) afterRequest(err error, duration time.Duration) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, _ := cb.currentState(now)

	if err != nil {
		cb.onFailure(state, now)
	} else {
		cb.onSuccess(state, now)
	}

	cb.logger.Debug("Request completed",
		zap.Error(err),
		zap.Duration("duration", duration),
		zap.String("state", state.String()),
		zap.Uint32("requests", cb.counts.Requests),
		zap.Uint32("failures", cb.counts.TotalFailures),
	)
}

// onSuccess handles successful requests
func (cb *CircuitBreaker) onSuccess(state State, now time.Time) {
	cb.counts.TotalSuccesses++
	cb.counts.ConsecutiveFailures = 0
	cb.counts.ConsecutiveSuccesses++

	if state == StateHalfOpen && cb.counts.ConsecutiveSuccesses >= cb.config.SuccessThreshold {
		cb.setState(StateClosed, now)
	}
}

// onFailure handles failed requests
func (cb *CircuitBreaker) onFailure(state State, now time.Time) {
	cb.counts.TotalFailures++
	cb.counts.ConsecutiveSuccesses = 0
	cb.counts.ConsecutiveFailures++

	if state == StateClosed {
		if cb.shouldOpen() {
			cb.setState(StateOpen, now)
		}
	} else if state == StateHalfOpen {
		cb.setState(StateOpen, now)
	}
}

// shouldOpen determines if the circuit should open
func (cb *CircuitBreaker) shouldOpen() bool {
	// Need minimum number of requests
	if cb.counts.Requests < cb.config.MinRequestThreshold {
		return false
	}

	// Check failure threshold
	return cb.counts.ConsecutiveFailures >= cb.config.FailureThreshold
}

// currentState returns the current state and counts
func (cb *CircuitBreaker) currentState(now time.Time) (State, Counts) {
	switch cb.state {
	case StateClosed:
		if cb.shouldClearCounts(now) {
			cb.counts.Reset()
		}
	case StateOpen:
		if now.Sub(cb.stateChanged) >= cb.config.Timeout {
			cb.setState(StateHalfOpen, now)
		}
	}

	return cb.state, cb.counts
}

// shouldClearCounts determines if counts should be cleared
func (cb *CircuitBreaker) shouldClearCounts(now time.Time) bool {
	return now.Sub(cb.stateChanged) >= cb.config.Interval
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(state State, now time.Time) {
	if cb.state == state {
		return
	}

	prevState := cb.state
	cb.state = state
	cb.stateChanged = now

	if state == StateClosed || state == StateHalfOpen {
		cb.counts.Reset()
	}

	cb.logger.Info("Circuit breaker state changed",
		zap.String("from", prevState.String()),
		zap.String("to", state.String()),
		zap.Time("changed_at", now),
	)

	if cb.onStateChange != nil {
		go cb.onStateChange(cb.name, prevState, state)
	}
}

// State returns the current state
func (cb *CircuitBreaker) State() State {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	state, _ := cb.currentState(time.Now())
	return state
}

// Counts returns the current counts
func (cb *CircuitBreaker) Counts() Counts {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	_, counts := cb.currentState(time.Now())
	return counts
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

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	cb.setState(StateClosed, now)
	cb.logger.Info("Circuit breaker manually reset")
}

// ForceOpen manually forces the circuit breaker to open state
func (cb *CircuitBreaker) ForceOpen() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	cb.setState(StateOpen, now)
	cb.logger.Info("Circuit breaker manually opened")
}

// Stats returns detailed statistics
func (cb *CircuitBreaker) Stats() map[string]interface{} {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	state, counts := cb.currentState(time.Now())

	return map[string]interface{}{
		"name":                   cb.name,
		"state":                  state.String(),
		"state_changed":          cb.stateChanged,
		"requests":               counts.Requests,
		"total_successes":        counts.TotalSuccesses,
		"total_failures":         counts.TotalFailures,
		"consecutive_successes":  counts.ConsecutiveSuccesses,
		"consecutive_failures":   counts.ConsecutiveFailures,
		"failure_rate":           counts.FailureRate(),
		"config":                 cb.config,
	}
}

// Common errors
var (
	ErrCircuitBreakerOpen     = errors.New("circuit breaker is open")
	ErrCircuitBreakerHalfOpen = errors.New("circuit breaker is half-open and max requests exceeded")
	ErrTimeout                = errors.New("request timeout")
)
