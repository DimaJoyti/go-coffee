package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"go-coffee-ai-agents/orchestration-engine/internal/infrastructure/loadbalancer"
)

// OptimizedHTTPClient provides high-performance HTTP client with advanced features
type OptimizedHTTPClient struct {
	client       *http.Client
	loadBalancer *loadbalancer.LoadBalancer
	metrics      *HTTPMetrics
	logger       Logger
	config       *HTTPClientConfig
	mutex        sync.RWMutex
}

// HTTPClientConfig contains HTTP client configuration
type HTTPClientConfig struct {
	// Connection settings
	MaxIdleConns        int           `json:"max_idle_conns"`
	MaxIdleConnsPerHost int           `json:"max_idle_conns_per_host"`
	MaxConnsPerHost     int           `json:"max_conns_per_host"`
	IdleConnTimeout     time.Duration `json:"idle_conn_timeout"`
	
	// Timeout settings
	Timeout               time.Duration `json:"timeout"`
	DialTimeout           time.Duration `json:"dial_timeout"`
	KeepAlive             time.Duration `json:"keep_alive"`
	TLSHandshakeTimeout   time.Duration `json:"tls_handshake_timeout"`
	ResponseHeaderTimeout time.Duration `json:"response_header_timeout"`
	ExpectContinueTimeout time.Duration `json:"expect_continue_timeout"`
	
	// Retry settings
	MaxRetries    int           `json:"max_retries"`
	RetryDelay    time.Duration `json:"retry_delay"`
	RetryBackoff  float64       `json:"retry_backoff"`
	
	// Compression and optimization
	DisableCompression bool `json:"disable_compression"`
	DisableKeepAlives  bool `json:"disable_keep_alives"`
	
	// Security settings
	InsecureSkipVerify bool `json:"insecure_skip_verify"`
	
	// User agent
	UserAgent string `json:"user_agent"`
}

// HTTPMetrics tracks HTTP client performance metrics
type HTTPMetrics struct {
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	TotalResponseTime   time.Duration `json:"total_response_time"`
	ConnectionsCreated  int64         `json:"connections_created"`
	ConnectionsReused   int64         `json:"connections_reused"`
	DNSLookupTime       time.Duration `json:"dns_lookup_time"`
	ConnectTime         time.Duration `json:"connect_time"`
	TLSHandshakeTime    time.Duration `json:"tls_handshake_time"`
	LastUpdated         time.Time     `json:"last_updated"`
	mutex               sync.RWMutex
}

// Logger interface for logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
}

// NewOptimizedHTTPClient creates a new optimized HTTP client
func NewOptimizedHTTPClient(config *HTTPClientConfig, logger Logger) *OptimizedHTTPClient {
	if config == nil {
		config = DefaultHTTPClientConfig()
	}

	// Create custom transport with optimizations
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   config.DialTimeout,
			KeepAlive: config.KeepAlive,
		}).DialContext,
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       config.MaxConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ResponseHeaderTimeout: config.ResponseHeaderTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
		DisableCompression:    config.DisableCompression,
		DisableKeepAlives:     config.DisableKeepAlives,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	return &OptimizedHTTPClient{
		client:  client,
		metrics: &HTTPMetrics{LastUpdated: time.Now()},
		logger:  logger,
		config:  config,
	}
}

// DefaultHTTPClientConfig returns default HTTP client configuration
func DefaultHTTPClientConfig() *HTTPClientConfig {
	return &HTTPClientConfig{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		MaxConnsPerHost:       50,
		IdleConnTimeout:       90 * time.Second,
		Timeout:               30 * time.Second,
		DialTimeout:           10 * time.Second,
		KeepAlive:             30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxRetries:            3,
		RetryDelay:            1 * time.Second,
		RetryBackoff:          2.0,
		DisableCompression:    false,
		DisableKeepAlives:     false,
		InsecureSkipVerify:    false,
		UserAgent:             "OptimizedHTTPClient/1.0",
	}
}

// SetLoadBalancer sets the load balancer for the HTTP client
func (c *OptimizedHTTPClient) SetLoadBalancer(lb *loadbalancer.LoadBalancer) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.loadBalancer = lb
}

// Get performs an optimized HTTP GET request
func (c *OptimizedHTTPClient) Get(ctx context.Context, url string) (map[string]interface{}, error) {
	return c.doRequest(ctx, "GET", url, nil, nil)
}

// Post performs an optimized HTTP POST request
func (c *OptimizedHTTPClient) Post(ctx context.Context, url string, data interface{}) (map[string]interface{}, error) {
	return c.doRequest(ctx, "POST", url, data, nil)
}

// Put performs an optimized HTTP PUT request
func (c *OptimizedHTTPClient) Put(ctx context.Context, url string, data interface{}) (map[string]interface{}, error) {
	return c.doRequest(ctx, "PUT", url, data, nil)
}

// Delete performs an optimized HTTP DELETE request
func (c *OptimizedHTTPClient) Delete(ctx context.Context, url string) error {
	_, err := c.doRequest(ctx, "DELETE", url, nil, nil)
	return err
}

// doRequest performs the actual HTTP request with optimizations
func (c *OptimizedHTTPClient) doRequest(ctx context.Context, method, url string, data interface{}, headers map[string]string) (map[string]interface{}, error) {
	start := time.Now()
	
	// Update metrics
	defer func() {
		c.updateMetrics(time.Since(start), true)
	}()

	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			c.updateMetrics(time.Since(start), false)
			return nil, fmt.Errorf("failed to marshal request data: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		c.updateMetrics(time.Since(start), false)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.config.UserAgent)
	
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Add compression support
	if !c.config.DisableCompression {
		req.Header.Set("Accept-Encoding", "gzip, deflate")
	}

	// Perform request with retries
	var resp *http.Response
	var lastErr error
	
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate retry delay with exponential backoff
			backoffMultiplier := 1.0
			for i := 1; i < attempt; i++ {
				backoffMultiplier *= c.config.RetryBackoff
			}
			delay := time.Duration(float64(c.config.RetryDelay) * backoffMultiplier)
			
			c.logger.Debug("Retrying HTTP request",
				"attempt", attempt,
				"delay", delay,
				"url", url)
			
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		resp, lastErr = c.executeRequest(req)
		if lastErr == nil && resp.StatusCode < 500 {
			break // Success or client error (don't retry)
		}
		
		if resp != nil {
			resp.Body.Close()
		}
	}

	if lastErr != nil {
		c.updateMetrics(time.Since(start), false)
		return nil, fmt.Errorf("request failed after %d attempts: %w", c.config.MaxRetries+1, lastErr)
	}

	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.updateMetrics(time.Since(start), false)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode >= 400 {
		c.updateMetrics(time.Since(start), false)
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse JSON response
	var result map[string]interface{}
	if len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, &result); err != nil {
			// If JSON parsing fails, return raw response
			result = map[string]interface{}{
				"raw_response": string(responseBody),
				"status_code":  resp.StatusCode,
			}
		}
	} else {
		result = map[string]interface{}{
			"status_code": resp.StatusCode,
		}
	}

	c.logger.Debug("HTTP request completed",
		"method", method,
		"url", url,
		"status", resp.StatusCode,
		"duration", time.Since(start))

	return result, nil
}

// executeRequest executes a single HTTP request
func (c *OptimizedHTTPClient) executeRequest(req *http.Request) (*http.Response, error) {
	// Use load balancer if available
	c.mutex.RLock()
	lb := c.loadBalancer
	c.mutex.RUnlock()

	if lb != nil {
		endpoint, err := lb.SelectEndpoint()
		if err != nil {
			return nil, fmt.Errorf("failed to select endpoint: %w", err)
		}

		// Execute with circuit breaker protection
		var resp *http.Response
		var execErr error
		
		err = lb.ExecuteWithEndpoint(endpoint.ID, func() error {
			resp, execErr = c.client.Do(req)
			return execErr
		})
		
		if err != nil {
			return nil, err
		}
		
		return resp, execErr
	}

	// Direct execution without load balancer
	return c.client.Do(req)
}

// updateMetrics updates HTTP client metrics
func (c *OptimizedHTTPClient) updateMetrics(duration time.Duration, success bool) {
	c.metrics.mutex.Lock()
	defer c.metrics.mutex.Unlock()

	c.metrics.TotalRequests++
	c.metrics.TotalResponseTime += duration
	
	if success {
		c.metrics.SuccessfulRequests++
	} else {
		c.metrics.FailedRequests++
	}

	// Update average response time
	if c.metrics.TotalRequests > 0 {
		c.metrics.AverageResponseTime = c.metrics.TotalResponseTime / time.Duration(c.metrics.TotalRequests)
	}

	c.metrics.LastUpdated = time.Now()
}

// GetMetrics returns current HTTP client metrics
func (c *OptimizedHTTPClient) GetMetrics() *HTTPMetrics {
	c.metrics.mutex.RLock()
	defer c.metrics.mutex.RUnlock()

	// Return a copy
	metricsCopy := *c.metrics
	return &metricsCopy
}

// Close closes the HTTP client and cleans up resources
func (c *OptimizedHTTPClient) Close() error {
	c.client.CloseIdleConnections()
	c.logger.Info("HTTP client closed and idle connections cleaned up")
	return nil
}

// Health checks the health of the HTTP client
func (c *OptimizedHTTPClient) Health(ctx context.Context) error {
	// Simple health check - try to create a request
	_, err := http.NewRequestWithContext(ctx, "GET", "http://example.com", nil)
	return err
}

// BatchRequest represents a batch HTTP request
type BatchRequest struct {
	ID      string                 `json:"id"`
	Method  string                 `json:"method"`
	URL     string                 `json:"url"`
	Data    interface{}            `json:"data"`
	Headers map[string]string      `json:"headers"`
}

// BatchResponse represents a batch HTTP response
type BatchResponse struct {
	ID       string                 `json:"id"`
	Response map[string]interface{} `json:"response"`
	Error    error                  `json:"error"`
	Duration time.Duration          `json:"duration"`
}

// BatchExecute executes multiple HTTP requests concurrently
func (c *OptimizedHTTPClient) BatchExecute(ctx context.Context, requests []BatchRequest) []BatchResponse {
	responses := make([]BatchResponse, len(requests))
	var wg sync.WaitGroup

	// Execute requests concurrently
	for i, req := range requests {
		wg.Add(1)
		go func(index int, request BatchRequest) {
			defer wg.Done()
			
			start := time.Now()
			response, err := c.doRequest(ctx, request.Method, request.URL, request.Data, request.Headers)
			duration := time.Since(start)

			responses[index] = BatchResponse{
				ID:       request.ID,
				Response: response,
				Error:    err,
				Duration: duration,
			}
		}(i, req)
	}

	wg.Wait()
	
	c.logger.Info("Batch HTTP requests completed",
		"total_requests", len(requests),
		"successful", countSuccessful(responses),
		"failed", countFailed(responses))

	return responses
}

// countSuccessful counts successful responses in a batch
func countSuccessful(responses []BatchResponse) int {
	count := 0
	for _, resp := range responses {
		if resp.Error == nil {
			count++
		}
	}
	return count
}

// countFailed counts failed responses in a batch
func countFailed(responses []BatchResponse) int {
	count := 0
	for _, resp := range responses {
		if resp.Error != nil {
			count++
		}
	}
	return count
}

// HTTPClientPool manages a pool of HTTP clients for different services
type HTTPClientPool struct {
	clients map[string]*OptimizedHTTPClient
	config  *HTTPClientConfig
	logger  Logger
	mutex   sync.RWMutex
}

// NewHTTPClientPool creates a new HTTP client pool
func NewHTTPClientPool(config *HTTPClientConfig, logger Logger) *HTTPClientPool {
	return &HTTPClientPool{
		clients: make(map[string]*OptimizedHTTPClient),
		config:  config,
		logger:  logger,
	}
}

// GetClient gets or creates an HTTP client for a service
func (p *HTTPClientPool) GetClient(serviceName string) *OptimizedHTTPClient {
	p.mutex.RLock()
	client, exists := p.clients[serviceName]
	p.mutex.RUnlock()

	if exists {
		return client
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Double-check after acquiring write lock
	if client, exists := p.clients[serviceName]; exists {
		return client
	}

	// Create new client
	client = NewOptimizedHTTPClient(p.config, p.logger)
	p.clients[serviceName] = client
	
	p.logger.Info("Created new HTTP client for service", "service", serviceName)
	return client
}

// CloseAll closes all HTTP clients in the pool
func (p *HTTPClientPool) CloseAll() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for serviceName, client := range p.clients {
		if err := client.Close(); err != nil {
			p.logger.Error("Failed to close HTTP client", err, "service", serviceName)
		}
	}

	p.clients = make(map[string]*OptimizedHTTPClient)
	p.logger.Info("All HTTP clients closed")
	return nil
}

// GetPoolStats returns statistics for all clients in the pool
func (p *HTTPClientPool) GetPoolStats() map[string]*HTTPMetrics {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	stats := make(map[string]*HTTPMetrics)
	for serviceName, client := range p.clients {
		stats[serviceName] = client.GetMetrics()
	}
	return stats
}
