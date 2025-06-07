package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/bright-data-hub/config"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// MCPClient represents an enhanced MCP client with connection pooling and advanced features
type MCPClient struct {
	config      *config.BrightDataHubConfig
	httpClient  *http.Client
	rateLimiter *RateLimiter
	cache       *AdvancedCache
	router      *RequestRouter
	metrics     *MetricsCollector
	logger      *logger.Logger
	
	// Connection pooling
	connectionPool *ConnectionPool
	
	// Circuit breaker
	circuitBreaker *CircuitBreaker
	
	// Request tracking
	activeRequests sync.Map
	requestCounter int64
	mu             sync.RWMutex
}

// MCPRequest represents a request to an MCP server
type MCPRequest struct {
	ID     string      `json:"id,omitempty"`
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

// MCPResponse represents a response from an MCP server
type MCPResponse struct {
	ID     string      `json:"id,omitempty"`
	Result interface{} `json:"result,omitempty"`
	Error  *MCPError   `json:"error,omitempty"`
}

// MCPError represents an error from an MCP server
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ConnectionPool manages HTTP connections to MCP servers
type ConnectionPool struct {
	maxConnections int
	connections    chan *http.Client
	mu             sync.RWMutex
}

// CircuitBreaker implements circuit breaker pattern for fault tolerance
type CircuitBreaker struct {
	maxFailures   int
	resetTimeout  time.Duration
	failures      int
	lastFailTime  time.Time
	state         CircuitState
	mu            sync.RWMutex
}

type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// NewMCPClient creates a new enhanced MCP client
func NewMCPClient(cfg *config.BrightDataHubConfig, log *logger.Logger) (*MCPClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	
	// Create HTTP client with optimized settings
	httpClient := &http.Client{
		Timeout: cfg.RequestTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	
	// Initialize rate limiter
	rateLimiter := NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)
	
	// Initialize cache
	cache, err := NewAdvancedCache(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}
	
	// Initialize connection pool
	connectionPool := NewConnectionPool(cfg.MaxConcurrent)
	
	// Initialize circuit breaker
	circuitBreaker := NewCircuitBreaker(5, 30*time.Second) // 5 failures, 30s reset
	
	// Initialize metrics collector
	metrics := NewMetricsCollector(cfg.MetricsEnabled)
	
	// Initialize request router
	router := NewRequestRouter(cfg)
	
	client := &MCPClient{
		config:         cfg,
		httpClient:     httpClient,
		rateLimiter:    rateLimiter,
		cache:          cache,
		router:         router,
		metrics:        metrics,
		logger:         log,
		connectionPool: connectionPool,
		circuitBreaker: circuitBreaker,
		activeRequests: sync.Map{},
	}
	
	return client, nil
}

// CallMCP makes an enhanced MCP call with all features
func (c *MCPClient) CallMCP(ctx context.Context, method string, params interface{}) (*MCPResponse, error) {
	// Generate request ID
	requestID := c.generateRequestID()
	
	// Start metrics tracking
	startTime := time.Now()
	c.metrics.IncrementRequests(method)
	
	defer func() {
		c.metrics.RecordLatency(method, time.Since(startTime))
	}()
	
	// Check circuit breaker
	if !c.circuitBreaker.CanExecute() {
		c.metrics.IncrementErrors(method, "circuit_breaker_open")
		return nil, fmt.Errorf("circuit breaker is open")
	}
	
	// Check cache first
	cacheKey := c.generateCacheKey(method, params)
	if cachedResponse := c.cache.Get(cacheKey); cachedResponse != nil {
		c.metrics.IncrementCacheHits(method)
		c.logger.Debug("Cache hit for method: %s", method)
		return cachedResponse.(*MCPResponse), nil
	}
	c.metrics.IncrementCacheMisses(method)
	
	// Apply rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		c.metrics.IncrementErrors(method, "rate_limit_exceeded")
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}
	
	// Route request through intelligent router
	response, err := c.router.RouteRequest(ctx, c, method, params, requestID)
	if err != nil {
		c.circuitBreaker.RecordFailure()
		c.metrics.IncrementErrors(method, "request_failed")
		return nil, err
	}
	
	// Record success
	c.circuitBreaker.RecordSuccess()
	c.metrics.IncrementSuccesses(method)
	
	// Cache successful response
	c.cache.Set(cacheKey, response, c.config.CacheTTL)
	
	return response, nil
}

// ExecuteRequest performs the actual HTTP request
func (c *MCPClient) ExecuteRequest(ctx context.Context, method string, params interface{}, requestID string) (*MCPResponse, error) {
	// Track active request
	c.activeRequests.Store(requestID, time.Now())
	defer c.activeRequests.Delete(requestID)
	
	// Create request
	request := MCPRequest{
		ID:     requestID,
		Method: method,
		Params: params,
	}
	
	// Marshal request
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.MCPServerURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Request-ID", requestID)
	
	// Execute request
	c.logger.Debug("Executing MCP request: %s (ID: %s)", method, requestID)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	// Parse response
	var mcpResp MCPResponse
	if err := json.Unmarshal(body, &mcpResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	// Check for MCP errors
	if mcpResp.Error != nil {
		return nil, fmt.Errorf("MCP error %d: %s", mcpResp.Error.Code, mcpResp.Error.Message)
	}
	
	return &mcpResp, nil
}

// GetActiveRequests returns information about currently active requests
func (c *MCPClient) GetActiveRequests() map[string]time.Time {
	result := make(map[string]time.Time)
	c.activeRequests.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(time.Time)
		return true
	})
	return result
}

// GetMetrics returns current metrics
func (c *MCPClient) GetMetrics() *Metrics {
	return c.metrics.GetMetrics()
}

// Close gracefully shuts down the client
func (c *MCPClient) Close() error {
	c.logger.Info("Shutting down MCP client")
	
	// Wait for active requests to complete (with timeout)
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			c.logger.Warn("Timeout waiting for active requests to complete")
			return nil
		case <-ticker.C:
			activeCount := 0
			c.activeRequests.Range(func(key, value interface{}) bool {
				activeCount++
				return true
			})
			if activeCount == 0 {
				c.logger.Info("All active requests completed")
				return nil
			}
		}
	}
}

// Helper methods
func (c *MCPClient) generateRequestID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestCounter++
	return fmt.Sprintf("req_%d_%d", time.Now().UnixNano(), c.requestCounter)
}

func (c *MCPClient) generateCacheKey(method string, params interface{}) string {
	paramsJSON, _ := json.Marshal(params)
	return fmt.Sprintf("%s:%x", method, paramsJSON)
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(maxConnections int) *ConnectionPool {
	pool := &ConnectionPool{
		maxConnections: maxConnections,
		connections:    make(chan *http.Client, maxConnections),
	}
	
	// Pre-populate the pool
	for i := 0; i < maxConnections; i++ {
		client := &http.Client{
			Timeout: 30 * time.Second,
		}
		pool.connections <- client
	}
	
	return pool
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CircuitClosed,
	}
}

// CanExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.failures = 0
	cb.state = CircuitClosed
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.failures++
	cb.lastFailTime = time.Now()
	
	if cb.failures >= cb.maxFailures {
		cb.state = CircuitOpen
	}
}
