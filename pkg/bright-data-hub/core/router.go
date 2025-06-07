package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/bright-data-hub/config"
)

// RequestRouter handles intelligent routing of requests
type RequestRouter struct {
	config         *config.BrightDataHubConfig
	loadBalancer   *LoadBalancer
	circuitBreaker *CircuitBreaker
	mu             sync.RWMutex
}

// LoadBalancer manages load balancing across multiple endpoints
type LoadBalancer struct {
	endpoints []string
	current   int
	mu        sync.RWMutex
}

// NewRequestRouter creates a new request router
func NewRequestRouter(cfg *config.BrightDataHubConfig) *RequestRouter {
	return &RequestRouter{
		config:         cfg,
		loadBalancer:   NewLoadBalancer([]string{cfg.MCPServerURL}),
		circuitBreaker: NewCircuitBreaker(5, 30*time.Second),
	}
}

// RouteRequest routes a request through the appropriate path
func (r *RequestRouter) RouteRequest(ctx context.Context, client *MCPClient, method string, params interface{}, requestID string) (*MCPResponse, error) {
	// Select endpoint
	endpoint := r.loadBalancer.GetNext()
	
	// Check circuit breaker for this endpoint
	if !r.circuitBreaker.CanExecute() {
		return nil, fmt.Errorf("circuit breaker open for endpoint: %s", endpoint)
	}
	
	// Execute request
	response, err := client.ExecuteRequest(ctx, method, params, requestID)
	if err != nil {
		r.circuitBreaker.RecordFailure()
		return nil, err
	}
	
	r.circuitBreaker.RecordSuccess()
	return response, nil
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(endpoints []string) *LoadBalancer {
	return &LoadBalancer{
		endpoints: endpoints,
		current:   0,
	}
}

// GetNext returns the next endpoint in round-robin fashion
func (lb *LoadBalancer) GetNext() string {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	if len(lb.endpoints) == 0 {
		return ""
	}
	
	endpoint := lb.endpoints[lb.current]
	lb.current = (lb.current + 1) % len(lb.endpoints)
	return endpoint
}

// AddEndpoint adds a new endpoint to the load balancer
func (lb *LoadBalancer) AddEndpoint(endpoint string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	lb.endpoints = append(lb.endpoints, endpoint)
}

// RemoveEndpoint removes an endpoint from the load balancer
func (lb *LoadBalancer) RemoveEndpoint(endpoint string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	for i, ep := range lb.endpoints {
		if ep == endpoint {
			lb.endpoints = append(lb.endpoints[:i], lb.endpoints[i+1:]...)
			if lb.current >= len(lb.endpoints) {
				lb.current = 0
			}
			break
		}
	}
}
