package loadbalancer

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	// StateClosed - circuit breaker is closed, requests pass through
	StateClosed CircuitBreakerState = iota
	// StateOpen - circuit breaker is open, requests fail fast
	StateOpen
	// StateHalfOpen - circuit breaker is half-open, testing if service recovered
	StateHalfOpen
)

func (s CircuitBreakerState) String() string {
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

// CircuitBreakerConfig contains circuit breaker configuration
type CircuitBreakerConfig struct {
	// MaxRequests is the maximum number of requests allowed to pass through
	// when the circuit breaker is half-open
	MaxRequests uint32 `json:"max_requests"`
	
	// Interval is the cyclic period of the closed state for the circuit breaker
	// to clear the internal counts
	Interval time.Duration `json:"interval"`
	
	// Timeout is the period of the open state, after which the state becomes half-open
	Timeout time.Duration `json:"timeout"`
	
	// ReadyToTrip is called with a copy of counts whenever a request fails in the closed state
	// If ReadyToTrip returns true, the circuit breaker will be placed into the open state
	ReadyToTrip func(counts Counts) bool `json:"-"`
	
	// OnStateChange is called whenever the state of the circuit breaker changes
	OnStateChange func(name string, from CircuitBreakerState, to CircuitBreakerState) `json:"-"`
	
	// IsSuccessful is called with the error returned from a request
	// If IsSuccessful returns true, the request is considered successful
	IsSuccessful func(err error) bool `json:"-"`
}

// Counts holds the numbers of requests and their successes/failures
type Counts struct {
	Requests             uint32 `json:"requests"`
	TotalSuccesses       uint32 `json:"total_successes"`
	TotalFailures        uint32 `json:"total_failures"`
	ConsecutiveSuccesses uint32 `json:"consecutive_successes"`
	ConsecutiveFailures  uint32 `json:"consecutive_failures"`
}

// CircuitBreaker prevents cascading failures and provides fail-fast behavior
type CircuitBreaker struct {
	name          string
	maxRequests   uint32
	interval      time.Duration
	timeout       time.Duration
	readyToTrip   func(counts Counts) bool
	isSuccessful  func(err error) bool
	onStateChange func(name string, from CircuitBreakerState, to CircuitBreakerState)

	mutex      sync.Mutex
	state      CircuitBreakerState
	generation uint64
	counts     Counts
	expiry     time.Time
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, config CircuitBreakerConfig) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:        name,
		maxRequests: config.MaxRequests,
		interval:    config.Interval,
		timeout:     config.Timeout,
		readyToTrip: config.ReadyToTrip,
		isSuccessful: config.IsSuccessful,
		onStateChange: config.OnStateChange,
	}

	// Set default values
	if cb.maxRequests == 0 {
		cb.maxRequests = 1
	}
	if cb.interval <= 0 {
		cb.interval = 60 * time.Second
	}
	if cb.timeout <= 0 {
		cb.timeout = 60 * time.Second
	}
	if cb.readyToTrip == nil {
		cb.readyToTrip = defaultReadyToTrip
	}
	if cb.isSuccessful == nil {
		cb.isSuccessful = defaultIsSuccessful
	}

	cb.toNewGeneration(time.Now())
	return cb
}

// Execute runs the given request if the circuit breaker accepts it
func (cb *CircuitBreaker) Execute(fn func() error) error {
	generation, err := cb.beforeRequest()
	if err != nil {
		return err
	}

	defer func() {
		e := recover()
		if e != nil {
			cb.afterRequest(generation, false)
			panic(e)
		}
	}()

	result := fn()
	cb.afterRequest(generation, cb.isSuccessful(result))
	return result
}

// ExecuteWithContext runs the given request with context if the circuit breaker accepts it
func (cb *CircuitBreaker) ExecuteWithContext(ctx context.Context, fn func(context.Context) error) error {
	generation, err := cb.beforeRequest()
	if err != nil {
		return err
	}

	defer func() {
		e := recover()
		if e != nil {
			cb.afterRequest(generation, false)
			panic(e)
		}
	}()

	result := fn(ctx)
	cb.afterRequest(generation, cb.isSuccessful(result))
	return result
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, _ := cb.currentState(now)
	return state
}

// Counts returns a copy of the current counts
func (cb *CircuitBreaker) Counts() Counts {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	return cb.counts
}

// beforeRequest is called before a request
func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if state == StateOpen {
		return generation, ErrCircuitBreakerOpen
	} else if state == StateHalfOpen && cb.counts.Requests >= cb.maxRequests {
		return generation, ErrCircuitBreakerOpen
	}

	cb.counts.Requests++
	return generation, nil
}

// afterRequest is called after a request
func (cb *CircuitBreaker) afterRequest(before uint64, success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)
	if generation != before {
		return
	}

	if success {
		cb.onSuccess(state, now)
	} else {
		cb.onFailure(state, now)
	}
}

// onSuccess handles successful requests
func (cb *CircuitBreaker) onSuccess(state CircuitBreakerState, now time.Time) {
	cb.counts.TotalSuccesses++
	cb.counts.ConsecutiveSuccesses++
	cb.counts.ConsecutiveFailures = 0

	if state == StateHalfOpen && cb.counts.ConsecutiveSuccesses >= cb.maxRequests {
		cb.setState(StateClosed, now)
	}
}

// onFailure handles failed requests
func (cb *CircuitBreaker) onFailure(state CircuitBreakerState, now time.Time) {
	cb.counts.TotalFailures++
	cb.counts.ConsecutiveFailures++
	cb.counts.ConsecutiveSuccesses = 0

	if state == StateClosed {
		if cb.readyToTrip(cb.counts) {
			cb.setState(StateOpen, now)
		}
	} else if state == StateHalfOpen {
		cb.setState(StateOpen, now)
	}
}

// currentState returns the current state and generation
func (cb *CircuitBreaker) currentState(now time.Time) (CircuitBreakerState, uint64) {
	switch cb.state {
	case StateClosed:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.toNewGeneration(now)
		}
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
		}
	}
	return cb.state, cb.generation
}

// setState changes the state of the circuit breaker
func (cb *CircuitBreaker) setState(state CircuitBreakerState, now time.Time) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state

	cb.toNewGeneration(now)

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, prev, state)
	}
}

// toNewGeneration creates a new generation
func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
	cb.generation++
	cb.counts = Counts{}

	var zero time.Time
	switch cb.state {
	case StateClosed:
		if cb.interval == 0 {
			cb.expiry = zero
		} else {
			cb.expiry = now.Add(cb.interval)
		}
	case StateOpen:
		cb.expiry = now.Add(cb.timeout)
	default: // StateHalfOpen
		cb.expiry = zero
	}
}

// defaultReadyToTrip is the default ReadyToTrip function
func defaultReadyToTrip(counts Counts) bool {
	return counts.ConsecutiveFailures > 5
}

// defaultIsSuccessful is the default IsSuccessful function
func defaultIsSuccessful(err error) bool {
	return err == nil
}

// Circuit breaker errors
var (
	ErrCircuitBreakerOpen = fmt.Errorf("circuit breaker is open")
	ErrTooManyRequests    = fmt.Errorf("too many requests")
)

// LoadBalancer distributes requests across multiple endpoints with circuit breakers
type LoadBalancer struct {
	endpoints       []Endpoint
	strategy        LoadBalancingStrategy
	circuitBreakers map[string]*CircuitBreaker
	mutex           sync.RWMutex
	logger          Logger
}

// Endpoint represents a service endpoint
type Endpoint struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Weight   int    `json:"weight"`
	Healthy  bool   `json:"healthy"`
	LastSeen time.Time `json:"last_seen"`
}

// LoadBalancingStrategy defines load balancing strategies
type LoadBalancingStrategy int

const (
	// RoundRobin distributes requests evenly across endpoints
	RoundRobin LoadBalancingStrategy = iota
	// WeightedRoundRobin distributes requests based on endpoint weights
	WeightedRoundRobin
	// LeastConnections routes to endpoint with fewest active connections
	LeastConnections
	// Random selects endpoints randomly
	Random
)

// Logger interface for logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(strategy LoadBalancingStrategy, logger Logger) *LoadBalancer {
	return &LoadBalancer{
		endpoints:       make([]Endpoint, 0),
		strategy:        strategy,
		circuitBreakers: make(map[string]*CircuitBreaker),
		logger:          logger,
	}
}

// AddEndpoint adds an endpoint to the load balancer
func (lb *LoadBalancer) AddEndpoint(endpoint Endpoint) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	lb.endpoints = append(lb.endpoints, endpoint)
	
	// Create circuit breaker for the endpoint
	config := CircuitBreakerConfig{
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
		OnStateChange: func(name string, from, to CircuitBreakerState) {
			lb.logger.Info("Circuit breaker state changed",
				"endpoint", name,
				"from", from.String(),
				"to", to.String())
		},
	}
	
	lb.circuitBreakers[endpoint.ID] = NewCircuitBreaker(endpoint.ID, config)
	lb.logger.Info("Endpoint added to load balancer", "endpoint", endpoint.ID, "url", endpoint.URL)
}

// RemoveEndpoint removes an endpoint from the load balancer
func (lb *LoadBalancer) RemoveEndpoint(endpointID string) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	for i, endpoint := range lb.endpoints {
		if endpoint.ID == endpointID {
			lb.endpoints = append(lb.endpoints[:i], lb.endpoints[i+1:]...)
			delete(lb.circuitBreakers, endpointID)
			lb.logger.Info("Endpoint removed from load balancer", "endpoint", endpointID)
			break
		}
	}
}

// SelectEndpoint selects an endpoint based on the load balancing strategy
func (lb *LoadBalancer) SelectEndpoint() (*Endpoint, error) {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	healthyEndpoints := lb.getHealthyEndpoints()
	if len(healthyEndpoints) == 0 {
		return nil, fmt.Errorf("no healthy endpoints available")
	}

	switch lb.strategy {
	case RoundRobin:
		return lb.selectRoundRobin(healthyEndpoints), nil
	case WeightedRoundRobin:
		return lb.selectWeightedRoundRobin(healthyEndpoints), nil
	case LeastConnections:
		return lb.selectLeastConnections(healthyEndpoints), nil
	case Random:
		return lb.selectRandom(healthyEndpoints), nil
	default:
		return lb.selectRoundRobin(healthyEndpoints), nil
	}
}

// ExecuteWithEndpoint executes a function with circuit breaker protection
func (lb *LoadBalancer) ExecuteWithEndpoint(endpointID string, fn func() error) error {
	lb.mutex.RLock()
	cb, exists := lb.circuitBreakers[endpointID]
	lb.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("circuit breaker not found for endpoint: %s", endpointID)
	}

	return cb.Execute(fn)
}

// getHealthyEndpoints returns endpoints that are healthy and have closed circuit breakers
func (lb *LoadBalancer) getHealthyEndpoints() []Endpoint {
	var healthy []Endpoint
	
	for _, endpoint := range lb.endpoints {
		if !endpoint.Healthy {
			continue
		}
		
		cb, exists := lb.circuitBreakers[endpoint.ID]
		if !exists || cb.State() == StateOpen {
			continue
		}
		
		healthy = append(healthy, endpoint)
	}
	
	return healthy
}

// selectRoundRobin implements round-robin selection
func (lb *LoadBalancer) selectRoundRobin(endpoints []Endpoint) *Endpoint {
	// Simple round-robin implementation
	// In production, you'd maintain a counter
	if len(endpoints) > 0 {
		return &endpoints[0]
	}
	return nil
}

// selectWeightedRoundRobin implements weighted round-robin selection
func (lb *LoadBalancer) selectWeightedRoundRobin(endpoints []Endpoint) *Endpoint {
	totalWeight := 0
	for _, endpoint := range endpoints {
		totalWeight += endpoint.Weight
	}
	
	if totalWeight == 0 {
		return lb.selectRoundRobin(endpoints)
	}
	
	// Simple weighted selection (in production, use proper weighted round-robin)
	for _, endpoint := range endpoints {
		if endpoint.Weight > 0 {
			return &endpoint
		}
	}
	
	return nil
}

// selectLeastConnections implements least connections selection
func (lb *LoadBalancer) selectLeastConnections(endpoints []Endpoint) *Endpoint {
	// In production, you'd track active connections per endpoint
	// For now, return the first endpoint
	if len(endpoints) > 0 {
		return &endpoints[0]
	}
	return nil
}

// selectRandom implements random selection
func (lb *LoadBalancer) selectRandom(endpoints []Endpoint) *Endpoint {
	// Simple random selection
	if len(endpoints) > 0 {
		// In production, use proper random selection
		return &endpoints[0]
	}
	return nil
}

// UpdateEndpointHealth updates the health status of an endpoint
func (lb *LoadBalancer) UpdateEndpointHealth(endpointID string, healthy bool) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	for i, endpoint := range lb.endpoints {
		if endpoint.ID == endpointID {
			lb.endpoints[i].Healthy = healthy
			lb.endpoints[i].LastSeen = time.Now()
			lb.logger.Debug("Endpoint health updated",
				"endpoint", endpointID,
				"healthy", healthy)
			break
		}
	}
}

// GetEndpoints returns all endpoints
func (lb *LoadBalancer) GetEndpoints() []Endpoint {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	endpoints := make([]Endpoint, len(lb.endpoints))
	copy(endpoints, lb.endpoints)
	return endpoints
}

// GetCircuitBreakerStats returns circuit breaker statistics
func (lb *LoadBalancer) GetCircuitBreakerStats() map[string]CircuitBreakerStats {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	stats := make(map[string]CircuitBreakerStats)
	for id, cb := range lb.circuitBreakers {
		stats[id] = CircuitBreakerStats{
			State:  cb.State(),
			Counts: cb.Counts(),
		}
	}
	return stats
}

// CircuitBreakerStats represents circuit breaker statistics
type CircuitBreakerStats struct {
	State  CircuitBreakerState `json:"state"`
	Counts Counts              `json:"counts"`
}
