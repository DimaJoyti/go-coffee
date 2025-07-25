package performance_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/DimaJoyti/go-coffee/tests/testutils"
)

// LoadTestConfig defines configuration for load testing
type LoadTestConfig struct {
	BaseURL           string
	ConcurrentUsers   int
	TestDuration      time.Duration
	RampUpDuration    time.Duration
	RequestsPerSecond int
	Endpoints         []EndpointConfig
}

// EndpointConfig defines configuration for testing specific endpoints
type EndpointConfig struct {
	Path           string
	Method         string
	Weight         int // Relative frequency of requests
	RequiresAuth   bool
	PayloadFunc    func() any
	ValidateFunc   func(*http.Response) error
}

// LoadTestResults contains the results of a load test
type LoadTestResults struct {
	TotalRequests     int64
	SuccessfulRequests int64
	FailedRequests    int64
	AverageLatency    time.Duration
	P95Latency        time.Duration
	P99Latency        time.Duration
	MaxLatency        time.Duration
	MinLatency        time.Duration
	RequestsPerSecond float64
	ErrorRate         float64
	Throughput        float64
	EndpointResults   map[string]*EndpointResults
}

// EndpointResults contains results for a specific endpoint
type EndpointResults struct {
	Path              string
	TotalRequests     int64
	SuccessfulRequests int64
	FailedRequests    int64
	AverageLatency    time.Duration
	P95Latency        time.Duration
	P99Latency        time.Duration
	ErrorsByStatus    map[int]int64
}

// RequestResult represents the result of a single request
type RequestResult struct {
	Endpoint      string
	StatusCode    int
	Latency       time.Duration
	Error         error
	ResponseSize  int64
	Timestamp     time.Time
}

// LoadTester manages load testing execution
type LoadTester struct {
	config    *LoadTestConfig
	client    *http.Client
	authToken string
	results   chan *RequestResult
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewLoadTester creates a new load tester instance
func NewLoadTester(config *LoadTestConfig) *LoadTester {
	ctx, cancel := context.WithTimeout(context.Background(), config.TestDuration+config.RampUpDuration)
	
	return &LoadTester{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		results: make(chan *RequestResult, 10000),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Run executes the load test
func (lt *LoadTester) Run() (*LoadTestResults, error) {
	l := logger.New("load-tester")
	l.Info("Starting load test - users: %d, duration: %v, url: %s", 
		lt.config.ConcurrentUsers, lt.config.TestDuration, lt.config.BaseURL)

	// Authenticate if needed
	if err := lt.authenticate(); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Start result collector
	resultsChan := make(chan *LoadTestResults, 1)
	go lt.collectResults(resultsChan)

	// Start load generation
	lt.generateLoad()

	// Wait for completion
	lt.wg.Wait()
	close(lt.results)

	// Get final results
	results := <-resultsChan
	
	l.Info("Load test completed - requests: %d, success: %.2f%%, latency: %v, rps: %.2f",
		results.TotalRequests, 
		(1-results.ErrorRate)*100, 
		results.AverageLatency, 
		results.RequestsPerSecond)

	return results, nil
}

// authenticate obtains an authentication token for protected endpoints
func (lt *LoadTester) authenticate() error {
	loginPayload := map[string]any{
		"email":    "loadtest@gocoffee.com",
		"password": "LoadTest123!",
	}

	body, _ := json.Marshal(loginPayload)
	req, err := http.NewRequest("POST", lt.config.BaseURL+"/api/v1/auth/login", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	
	resp, err := lt.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	var loginResp struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return err
	}

	lt.authToken = loginResp.AccessToken
	return nil
}

// generateLoad starts the load generation workers
func (lt *LoadTester) generateLoad() {
	// Calculate requests per user
	requestsPerUser := lt.config.RequestsPerSecond / lt.config.ConcurrentUsers
	if requestsPerUser == 0 {
		requestsPerUser = 1
	}

	// Start workers with ramp-up
	rampUpInterval := lt.config.RampUpDuration / time.Duration(lt.config.ConcurrentUsers)
	
	for i := 0; i < lt.config.ConcurrentUsers; i++ {
		lt.wg.Add(1)
		
		go func(workerID int) {
			defer lt.wg.Done()
			
			// Ramp-up delay
			time.Sleep(time.Duration(workerID) * rampUpInterval)
			
			// Request interval
			interval := time.Second / time.Duration(requestsPerUser)
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				select {
				case <-lt.ctx.Done():
					return
				case <-ticker.C:
					lt.makeRequest(workerID)
				}
			}
		}(i)
	}
}

// makeRequest executes a single request
func (lt *LoadTester) makeRequest(_ int) {
	// Select endpoint based on weight
	endpoint := lt.selectEndpoint()
	
	// Create request
	req, err := lt.createRequest(endpoint)
	if err != nil {
		lt.results <- &RequestResult{
			Endpoint:  endpoint.Path,
			Error:     err,
			Timestamp: time.Now(),
		}
		return
	}

	// Execute request
	start := time.Now()
	resp, err := lt.client.Do(req)
	latency := time.Since(start)

	result := &RequestResult{
		Endpoint:  endpoint.Path,
		Latency:   latency,
		Timestamp: time.Now(),
	}

	if err != nil {
		result.Error = err
	} else {
		result.StatusCode = resp.StatusCode
		result.ResponseSize = resp.ContentLength
		
		// Validate response if validator provided
		if endpoint.ValidateFunc != nil {
			if err := endpoint.ValidateFunc(resp); err != nil {
				result.Error = err
			}
		}
		
		resp.Body.Close()
	}

	lt.results <- result
}

// selectEndpoint selects an endpoint based on configured weights
func (lt *LoadTester) selectEndpoint() *EndpointConfig {
	totalWeight := 0
	for _, endpoint := range lt.config.Endpoints {
		totalWeight += endpoint.Weight
	}

	// Simple round-robin for now (could be improved with weighted selection)
	return &lt.config.Endpoints[time.Now().Nanosecond()%len(lt.config.Endpoints)]
}

// createRequest creates an HTTP request for the given endpoint
func (lt *LoadTester) createRequest(endpoint *EndpointConfig) (*http.Request, error) {
	var body []byte
	var err error

	// Generate payload if needed
	if endpoint.PayloadFunc != nil {
		payload := endpoint.PayloadFunc()
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}

	// Create request
	req, err := http.NewRequest(endpoint.Method, lt.config.BaseURL+endpoint.Path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Go-Coffee-LoadTest/1.0")

	// Add authentication if required
	if endpoint.RequiresAuth && lt.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+lt.authToken)
	}

	return req, nil
}

// collectResults processes and aggregates test results
func (lt *LoadTester) collectResults(resultsChan chan *LoadTestResults) {
	results := &LoadTestResults{
		EndpointResults: make(map[string]*EndpointResults),
	}

	var latencies []time.Duration
	endpointLatencies := make(map[string][]time.Duration)

	for result := range lt.results {
		results.TotalRequests++

		// Track endpoint-specific results
		if _, exists := results.EndpointResults[result.Endpoint]; !exists {
			results.EndpointResults[result.Endpoint] = &EndpointResults{
				Path:           result.Endpoint,
				ErrorsByStatus: make(map[int]int64),
			}
		}

		endpointResult := results.EndpointResults[result.Endpoint]
		endpointResult.TotalRequests++

		if result.Error != nil {
			results.FailedRequests++
			endpointResult.FailedRequests++
		} else {
			results.SuccessfulRequests++
			endpointResult.SuccessfulRequests++
			
			// Track latencies
			latencies = append(latencies, result.Latency)
			endpointLatencies[result.Endpoint] = append(endpointLatencies[result.Endpoint], result.Latency)
			
			// Track status codes
			endpointResult.ErrorsByStatus[result.StatusCode]++
		}
	}

	// Calculate aggregate metrics
	if len(latencies) > 0 {
		results.AverageLatency = calculateAverage(latencies)
		results.P95Latency = calculatePercentile(latencies, 95)
		results.P99Latency = calculatePercentile(latencies, 99)
		results.MaxLatency = calculateMax(latencies)
		results.MinLatency = calculateMin(latencies)
	}

	// Calculate endpoint-specific metrics
	for endpoint, endpointResult := range results.EndpointResults {
		if latencies, exists := endpointLatencies[endpoint]; exists && len(latencies) > 0 {
			endpointResult.AverageLatency = calculateAverage(latencies)
			endpointResult.P95Latency = calculatePercentile(latencies, 95)
			endpointResult.P99Latency = calculatePercentile(latencies, 99)
		}
	}

	// Calculate rates
	if results.TotalRequests > 0 {
		results.ErrorRate = float64(results.FailedRequests) / float64(results.TotalRequests)
		results.RequestsPerSecond = float64(results.TotalRequests) / lt.config.TestDuration.Seconds()
		results.Throughput = float64(results.SuccessfulRequests) / lt.config.TestDuration.Seconds()
	}

	resultsChan <- results
}

// Helper functions for statistical calculations
func calculateAverage(latencies []time.Duration) time.Duration {
	var total time.Duration
	for _, latency := range latencies {
		total += latency
	}
	return total / time.Duration(len(latencies))
}

func calculatePercentile(latencies []time.Duration, percentile int) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	
	// Simple percentile calculation (could be improved with proper sorting)
	index := (len(latencies) * percentile) / 100
	if index >= len(latencies) {
		index = len(latencies) - 1
	}
	return latencies[index]
}

func calculateMax(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	
	max := latencies[0]
	for _, latency := range latencies[1:] {
		if latency > max {
			max = latency
		}
	}
	return max
}

func calculateMin(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	
	min := latencies[0]
	for _, latency := range latencies[1:] {
		if latency < min {
			min = latency
		}
	}
	return min
}

// Test functions

// TestAPIGatewayLoad tests the API Gateway under load
func TestAPIGatewayLoad(t *testing.T) {
	config := &LoadTestConfig{
		BaseURL:           testutils.GetTestBaseURL(),
		ConcurrentUsers:   50,
		TestDuration:      2 * time.Minute,
		RampUpDuration:    30 * time.Second,
		RequestsPerSecond: 100,
		Endpoints: []EndpointConfig{
			{
				Path:   "/health",
				Method: "GET",
				Weight: 10,
				ValidateFunc: func(resp *http.Response) error {
					if resp.StatusCode != http.StatusOK {
						return fmt.Errorf("expected 200, got %d", resp.StatusCode)
					}
					return nil
				},
			},
			{
				Path:   "/api/v1/menu",
				Method: "GET",
				Weight: 30,
				ValidateFunc: func(resp *http.Response) error {
					if resp.StatusCode != http.StatusOK {
						return fmt.Errorf("expected 200, got %d", resp.StatusCode)
					}
					return nil
				},
			},
			{
				Path:         "/api/v1/orders",
				Method:       "GET",
				Weight:       20,
				RequiresAuth: true,
				ValidateFunc: func(resp *http.Response) error {
					if resp.StatusCode != http.StatusOK {
						return fmt.Errorf("expected 200, got %d", resp.StatusCode)
					}
					return nil
				},
			},
		},
	}

	tester := NewLoadTester(config)
	results, err := tester.Run()
	
	require.NoError(t, err)
	require.NotNil(t, results)

	// Performance assertions
	assert.Less(t, results.ErrorRate, 0.01, "Error rate should be less than 1%")
	assert.Less(t, results.P95Latency, 500*time.Millisecond, "P95 latency should be less than 500ms")
	assert.Less(t, results.P99Latency, 1*time.Second, "P99 latency should be less than 1s")
	assert.Greater(t, results.RequestsPerSecond, 80.0, "Should handle at least 80 RPS")

	// Log detailed results
	t.Logf("Load Test Results:")
	t.Logf("  Total Requests: %d", results.TotalRequests)
	t.Logf("  Success Rate: %.2f%%", (1-results.ErrorRate)*100)
	t.Logf("  Average Latency: %v", results.AverageLatency)
	t.Logf("  P95 Latency: %v", results.P95Latency)
	t.Logf("  P99 Latency: %v", results.P99Latency)
	t.Logf("  Requests/Second: %.2f", results.RequestsPerSecond)
	t.Logf("  Throughput: %.2f", results.Throughput)
}

// TestOrderServiceLoad tests the Order Service under load
func TestOrderServiceLoad(t *testing.T) {
	config := &LoadTestConfig{
		BaseURL:           testutils.GetTestBaseURL(),
		ConcurrentUsers:   30,
		TestDuration:      90 * time.Second,
		RampUpDuration:    20 * time.Second,
		RequestsPerSecond: 60,
		Endpoints: []EndpointConfig{
			{
				Path:         "/api/v1/orders",
				Method:       "POST",
				Weight:       50,
				RequiresAuth: true,
				PayloadFunc: func() any {
					return map[string]any{
						"items": []map[string]any{
							{
								"menu_item_id": "menu_123",
								"quantity":     1,
								"customizations": []string{"oat milk"},
							},
						},
						"payment_method_id": "pm_test_123",
					}
				},
				ValidateFunc: func(resp *http.Response) error {
					if resp.StatusCode != http.StatusCreated {
						return fmt.Errorf("expected 201, got %d", resp.StatusCode)
					}
					return nil
				},
			},
			{
				Path:         "/api/v1/orders",
				Method:       "GET",
				Weight:       30,
				RequiresAuth: true,
				ValidateFunc: func(resp *http.Response) error {
					if resp.StatusCode != http.StatusOK {
						return fmt.Errorf("expected 200, got %d", resp.StatusCode)
					}
					return nil
				},
			},
		},
	}

	tester := NewLoadTester(config)
	results, err := tester.Run()
	
	require.NoError(t, err)
	require.NotNil(t, results)

	// Performance assertions for order service
	assert.Less(t, results.ErrorRate, 0.02, "Error rate should be less than 2%")
	assert.Less(t, results.P95Latency, 800*time.Millisecond, "P95 latency should be less than 800ms")
	assert.Greater(t, results.RequestsPerSecond, 50.0, "Should handle at least 50 RPS")

	// Verify order creation performance
	orderCreateResults := results.EndpointResults["/api/v1/orders"]
	if orderCreateResults != nil {
		assert.Less(t, orderCreateResults.P95Latency, 1*time.Second, "Order creation P95 should be less than 1s")
	}
}

// BenchmarkSingleRequest benchmarks a single API request
func BenchmarkSingleRequest(b *testing.B) {
	client := &http.Client{Timeout: 10 * time.Second}
	baseURL := testutils.GetTestBaseURL()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resp, err := client.Get(baseURL + "/health")
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

// BenchmarkConcurrentRequests benchmarks concurrent API requests
func BenchmarkConcurrentRequests(b *testing.B) {
	client := &http.Client{Timeout: 10 * time.Second}
	baseURL := testutils.GetTestBaseURL()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(baseURL + "/health")
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
		}
	})
}
