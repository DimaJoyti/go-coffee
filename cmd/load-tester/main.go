package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

type LoadTester struct {
	services []ServiceEndpoint
	results  *TestResults
	mu       sync.RWMutex
}

type ServiceEndpoint struct {
	Name    string            `json:"name"`
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Body    string            `json:"body,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type TestResults struct {
	TotalRequests  int64                     `json:"total_requests"`
	SuccessfulReqs int64                     `json:"successful_requests"`
	FailedReqs     int64                     `json:"failed_requests"`
	TotalDuration  time.Duration             `json:"total_duration"`
	AverageLatency time.Duration             `json:"average_latency"`
	MinLatency     time.Duration             `json:"min_latency"`
	MaxLatency     time.Duration             `json:"max_latency"`
	RequestsPerSec float64                   `json:"requests_per_second"`
	ServiceResults map[string]*ServiceResult `json:"service_results"`
	StartTime      time.Time                 `json:"start_time"`
	EndTime        time.Time                 `json:"end_time"`
}

type ServiceResult struct {
	Name           string        `json:"name"`
	Requests       int64         `json:"requests"`
	Successes      int64         `json:"successes"`
	Failures       int64         `json:"failures"`
	AverageLatency time.Duration `json:"average_latency"`
	MinLatency     time.Duration `json:"min_latency"`
	MaxLatency     time.Duration `json:"max_latency"`
	ErrorRate      float64       `json:"error_rate"`
}

func NewLoadTester() *LoadTester {
	return &LoadTester{
		services: []ServiceEndpoint{
			{Name: "api-gateway", URL: "http://localhost:8080/health", Method: "GET"},
			{Name: "redis-mcp-server", URL: "http://localhost:8108/api/v1/redis-mcp/health", Method: "GET"},
			{Name: "web-ui-mcp-server", URL: "http://localhost:3001/health", Method: "GET"},
			{Name: "web-ui-backend", URL: "http://localhost:8090/health", Method: "GET"},
			{Name: "ai-search", URL: "http://localhost:8092/api/v1/ai-search/health", Method: "GET"},
			{Name: "frontend", URL: "http://localhost:3002/", Method: "GET"},
			{Name: "performance-monitor", URL: "http://localhost:9999/health", Method: "GET"},
			{
				Name:    "ai-search-semantic",
				URL:     "http://localhost:8092/api/v1/ai-search/semantic",
				Method:  "POST",
				Body:    `{"query":"load test coffee","limit":3}`,
				Headers: map[string]string{"Content-Type": "application/json"},
			},
		},
		results: &TestResults{
			ServiceResults: make(map[string]*ServiceResult),
			MinLatency:     time.Hour, // Initialize with high value
		},
	}
}

func (lt *LoadTester) makeRequest(endpoint ServiceEndpoint) (time.Duration, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	start := time.Now()

	var req *http.Request
	var err error

	if endpoint.Body != "" {
		req, err = http.NewRequest(endpoint.Method, endpoint.URL,
			strings.NewReader(endpoint.Body))
	} else {
		req, err = http.NewRequest(endpoint.Method, endpoint.URL, nil)
	}

	if err != nil {
		return 0, err
	}

	// Add headers
	for key, value := range endpoint.Headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return duration, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return duration, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return duration, nil
}

func (lt *LoadTester) runLoadTest(ctx context.Context, concurrency int, duration time.Duration) {
	fmt.Printf("🚀 Starting load test with %d concurrent workers for %v\n", concurrency, duration)

	lt.results.StartTime = time.Now()

	// Initialize service results
	for _, service := range lt.services {
		lt.results.ServiceResults[service.Name] = &ServiceResult{
			Name:       service.Name,
			MinLatency: time.Hour,
		}
	}

	var wg sync.WaitGroup
	testCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	// Start workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			lt.worker(testCtx, workerID)
		}(i)
	}

	// Wait for all workers to complete
	wg.Wait()

	lt.results.EndTime = time.Now()
	lt.results.TotalDuration = lt.results.EndTime.Sub(lt.results.StartTime)

	// Calculate final statistics
	lt.calculateStats()

	fmt.Printf("✅ Load test completed!\n")
}

func (lt *LoadTester) worker(ctx context.Context, workerID int) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Test each service
			for _, service := range lt.services {
				select {
				case <-ctx.Done():
					return
				default:
					lt.testService(service)
				}
			}
		}
	}
}

func (lt *LoadTester) testService(service ServiceEndpoint) {
	duration, err := lt.makeRequest(service)

	atomic.AddInt64(&lt.results.TotalRequests, 1)

	lt.mu.Lock()
	serviceResult := lt.results.ServiceResults[service.Name]
	serviceResult.Requests++

	if err != nil {
		atomic.AddInt64(&lt.results.FailedReqs, 1)
		serviceResult.Failures++
	} else {
		atomic.AddInt64(&lt.results.SuccessfulReqs, 1)
		serviceResult.Successes++

		// Update latency stats
		if duration < lt.results.MinLatency {
			lt.results.MinLatency = duration
		}
		if duration > lt.results.MaxLatency {
			lt.results.MaxLatency = duration
		}

		if duration < serviceResult.MinLatency {
			serviceResult.MinLatency = duration
		}
		if duration > serviceResult.MaxLatency {
			serviceResult.MaxLatency = duration
		}
	}
	lt.mu.Unlock()
}

func (lt *LoadTester) calculateStats() {
	if lt.results.TotalRequests > 0 {
		lt.results.RequestsPerSec = float64(lt.results.TotalRequests) / lt.results.TotalDuration.Seconds()

		// Calculate average latency (simplified)
		if lt.results.SuccessfulReqs > 0 {
			lt.results.AverageLatency = (lt.results.MinLatency + lt.results.MaxLatency) / 2
		}
	}

	// Calculate service-specific stats
	for _, serviceResult := range lt.results.ServiceResults {
		if serviceResult.Requests > 0 {
			serviceResult.ErrorRate = float64(serviceResult.Failures) / float64(serviceResult.Requests) * 100
			if serviceResult.Successes > 0 {
				serviceResult.AverageLatency = (serviceResult.MinLatency + serviceResult.MaxLatency) / 2
			}
		}
	}
}

func (lt *LoadTester) printResults() {
	fmt.Println("\n📊 LOAD TEST RESULTS")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Printf("🕐 Duration: %v\n", lt.results.TotalDuration)
	fmt.Printf("📈 Total Requests: %d\n", lt.results.TotalRequests)
	fmt.Printf("✅ Successful: %d\n", lt.results.SuccessfulReqs)
	fmt.Printf("❌ Failed: %d\n", lt.results.FailedReqs)
	fmt.Printf("⚡ Requests/sec: %.2f\n", lt.results.RequestsPerSec)
	fmt.Printf("⏱️  Average Latency: %v\n", lt.results.AverageLatency)
	fmt.Printf("🚀 Min Latency: %v\n", lt.results.MinLatency)
	fmt.Printf("🐌 Max Latency: %v\n", lt.results.MaxLatency)

	fmt.Println("\n📋 SERVICE BREAKDOWN")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	for name, result := range lt.results.ServiceResults {
		fmt.Printf("🔧 %s:\n", name)
		fmt.Printf("   Requests: %d | Success: %d | Failed: %d\n",
			result.Requests, result.Successes, result.Failures)
		fmt.Printf("   Error Rate: %.2f%% | Avg Latency: %v\n",
			result.ErrorRate, result.AverageLatency)
		fmt.Println()
	}
}

func main() {
	fmt.Println("🚀 Go Coffee Load Tester")
	fmt.Println("📊 High-performance load testing for microservices")

	loadTester := NewLoadTester()

	// Setup HTTP server for results
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"service":   "load-tester",
			"timestamp": time.Now(),
		})
	})

	r.GET("/results", func(c *gin.Context) {
		c.JSON(200, loadTester.results)
	})

	r.POST("/test", func(c *gin.Context) {
		var req struct {
			Concurrency int    `json:"concurrency"`
			Duration    string `json:"duration"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		duration, err := time.ParseDuration(req.Duration)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid duration format"})
			return
		}

		go func() {
			ctx := context.Background()
			loadTester.runLoadTest(ctx, req.Concurrency, duration)
		}()

		c.JSON(200, gin.H{
			"message":     "Load test started",
			"concurrency": req.Concurrency,
			"duration":    req.Duration,
		})
	})

	// Start HTTP server in background
	go func() {
		port := os.Getenv("LOAD_TESTER_PORT")
		if port == "" {
			port = "8888"
		}

		fmt.Printf("🌐 Load Tester API running on http://localhost:%s\n", port)
		fmt.Printf("📊 Results: http://localhost:%s/results\n", port)
		fmt.Printf("🧪 Start test: POST http://localhost:%s/test\n", port)

		log.Fatal(http.ListenAndServe(":"+port, r))
	}()

	// Run default load test
	ctx := context.Background()

	// Quick test
	fmt.Println("🧪 Running quick performance test...")
	loadTester.runLoadTest(ctx, 5, 30*time.Second)
	loadTester.printResults()

	// Save results
	resultsFile := fmt.Sprintf("load-test-results-%s.json",
		time.Now().Format("20060102-150405"))

	if data, err := json.MarshalIndent(loadTester.results, "", "  "); err == nil {
		if err := os.WriteFile(resultsFile, data, 0644); err == nil {
			fmt.Printf("💾 Results saved to: %s\n", resultsFile)
		}
	}

	// Keep server running
	select {}
}
