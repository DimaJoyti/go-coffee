package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// TestOrder represents a test order for load testing
type TestOrder struct {
	CustomerName string `json:"customer_name"`
	CoffeeType   string `json:"coffee_type"`
}

// LoadTestResult holds the results of a load test
type LoadTestResult struct {
	TotalRequests     int
	SuccessfulRequests int
	FailedRequests    int
	AverageLatency    time.Duration
	MaxLatency        time.Duration
	MinLatency        time.Duration
	RequestsPerSecond float64
	StartTime         time.Time
	EndTime           time.Time
}

// RequestResult holds the result of a single request
type RequestResult struct {
	Success bool
	Latency time.Duration
	Error   error
}

func main() {
	// Test configuration
	serverURL := "http://localhost:8080"
	totalRequests := 1000
	concurrentWorkers := 50
	useAsyncMode := true

	log.Printf("🚀 Starting Load Test")
	log.Printf("📋 Configuration:")
	log.Printf("  Server URL: %s", serverURL)
	log.Printf("  Total Requests: %d", totalRequests)
	log.Printf("  Concurrent Workers: %d", concurrentWorkers)
	log.Printf("  Async Mode: %v", useAsyncMode)
	log.Println()

	// Test server health first
	if !testServerHealth(serverURL) {
		log.Fatal("❌ Server health check failed. Make sure the server is running.")
	}

	// Run baseline test (sync mode)
	log.Println("📊 Running Baseline Test (Sync Mode)...")
	syncResult := runLoadTest(serverURL, totalRequests/2, concurrentWorkers, false)
	printResults("Sync Mode", syncResult)

	// Wait a bit between tests
	time.Sleep(time.Second * 2)

	// Run optimized test (async mode)
	log.Println("📊 Running Optimized Test (Async Mode)...")
	asyncResult := runLoadTest(serverURL, totalRequests, concurrentWorkers, useAsyncMode)
	printResults("Async Mode", asyncResult)

	// Compare results
	compareResults(syncResult, asyncResult)

	// Test different concurrency levels
	log.Println("\n🔄 Testing Different Concurrency Levels...")
	testConcurrencyLevels(serverURL, useAsyncMode)

	// Test Redis batch operations if available
	log.Println("\n💾 Testing Cache Performance...")
	testCachePerformance(serverURL)

	log.Println("\n✅ Load Testing Complete!")
}

// testServerHealth checks if the server is responding
func testServerHealth(serverURL string) bool {
	resp, err := http.Get(serverURL + "/health")
	if err != nil {
		log.Printf("Health check error: %v", err)
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK
}

// runLoadTest executes a load test with the given parameters
func runLoadTest(serverURL string, requests, workers int, asyncMode bool) LoadTestResult {
	startTime := time.Now()
	
	// Create channels for work distribution and result collection
	workChan := make(chan int, requests)
	resultChan := make(chan RequestResult, requests)
	
	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(serverURL, asyncMode, workChan, resultChan, &wg)
	}
	
	// Send work to workers
	for i := 0; i < requests; i++ {
		workChan <- i
	}
	close(workChan)
	
	// Wait for all workers to complete
	wg.Wait()
	close(resultChan)
	
	endTime := time.Now()
	
	// Collect results
	var results []RequestResult
	for result := range resultChan {
		results = append(results, result)
	}
	
	return analyzeResults(results, startTime, endTime)
}

// worker performs HTTP requests
func worker(serverURL string, asyncMode bool, workChan <-chan int, resultChan chan<- RequestResult, wg *sync.WaitGroup) {
	defer wg.Done()
	
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	
	for requestID := range workChan {
		start := time.Now()
		
		// Create test order
		order := TestOrder{
			CustomerName: fmt.Sprintf("Customer-%d", requestID),
			CoffeeType:   "Espresso",
		}
		
		jsonData, _ := json.Marshal(order)
		
		// Build URL with async parameter if needed
		url := serverURL + "/order"
		if asyncMode {
			url += "?async=true"
		}
		
		// Make HTTP request
		resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
		latency := time.Since(start)
		
		result := RequestResult{
			Latency: latency,
		}
		
		if err != nil {
			result.Error = err
		} else {
			defer resp.Body.Close()
			result.Success = resp.StatusCode == http.StatusOK
			if !result.Success {
				body, _ := io.ReadAll(resp.Body)
				result.Error = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
			}
		}
		
		resultChan <- result
	}
}

// analyzeResults processes the test results and calculates metrics
func analyzeResults(results []RequestResult, startTime, endTime time.Time) LoadTestResult {
	totalRequests := len(results)
	successfulRequests := 0
	failedRequests := 0
	
	var totalLatency time.Duration
	var maxLatency time.Duration
	minLatency := time.Hour // Start with a large value
	
	for _, result := range results {
		if result.Success {
			successfulRequests++
		} else {
			failedRequests++
		}
		
		totalLatency += result.Latency
		
		if result.Latency > maxLatency {
			maxLatency = result.Latency
		}
		
		if result.Latency < minLatency {
			minLatency = result.Latency
		}
	}
	
	duration := endTime.Sub(startTime)
	requestsPerSecond := float64(totalRequests) / duration.Seconds()
	averageLatency := totalLatency / time.Duration(totalRequests)
	
	return LoadTestResult{
		TotalRequests:      totalRequests,
		SuccessfulRequests: successfulRequests,
		FailedRequests:     failedRequests,
		AverageLatency:     averageLatency,
		MaxLatency:         maxLatency,
		MinLatency:         minLatency,
		RequestsPerSecond:  requestsPerSecond,
		StartTime:          startTime,
		EndTime:            endTime,
	}
}

// printResults displays the test results
func printResults(testName string, result LoadTestResult) {
	fmt.Printf("\n📊 %s Results:\n", testName)
	fmt.Printf("  Total Requests: %d\n", result.TotalRequests)
	fmt.Printf("  Successful: %d (%.2f%%)\n", result.SuccessfulRequests, 
		float64(result.SuccessfulRequests)/float64(result.TotalRequests)*100)
	fmt.Printf("  Failed: %d (%.2f%%)\n", result.FailedRequests,
		float64(result.FailedRequests)/float64(result.TotalRequests)*100)
	fmt.Printf("  Requests/Second: %.2f\n", result.RequestsPerSecond)
	fmt.Printf("  Average Latency: %v\n", result.AverageLatency)
	fmt.Printf("  Min Latency: %v\n", result.MinLatency)
	fmt.Printf("  Max Latency: %v\n", result.MaxLatency)
	fmt.Printf("  Total Duration: %v\n", result.EndTime.Sub(result.StartTime))
}

// compareResults compares sync vs async results
func compareResults(syncResult, asyncResult LoadTestResult) {
	fmt.Printf("\n🔍 Performance Comparison:\n")
	
	rpsImprovement := ((asyncResult.RequestsPerSecond - syncResult.RequestsPerSecond) / syncResult.RequestsPerSecond) * 100
	latencyImprovement := ((syncResult.AverageLatency - asyncResult.AverageLatency).Nanoseconds() / syncResult.AverageLatency.Nanoseconds()) * 100
	
	fmt.Printf("  Requests/Second Improvement: %.2f%% (%s)\n", 
		rpsImprovement, 
		formatImprovement(rpsImprovement))
	
	fmt.Printf("  Average Latency Improvement: %.2f%% (%s)\n", 
		float64(latencyImprovement), 
		formatImprovement(float64(latencyImprovement)))
	
	successRateSync := float64(syncResult.SuccessfulRequests) / float64(syncResult.TotalRequests) * 100
	successRateAsync := float64(asyncResult.SuccessfulRequests) / float64(asyncResult.TotalRequests) * 100
	
	fmt.Printf("  Success Rate (Sync): %.2f%%\n", successRateSync)
	fmt.Printf("  Success Rate (Async): %.2f%%\n", successRateAsync)
}

// formatImprovement returns a formatted string with improvement indicator
func formatImprovement(improvement float64) string {
	if improvement > 0 {
		return "✅ Better"
	} else if improvement < 0 {
		return "❌ Worse"
	}
	return "➖ Same"
}

// testConcurrencyLevels tests different numbers of concurrent workers
func testConcurrencyLevels(serverURL string, asyncMode bool) {
	concurrencyLevels := []int{10, 25, 50, 100}
	requests := 500
	
	fmt.Printf("Testing concurrency levels with %d requests each:\n", requests)
	
	for _, concurrency := range concurrencyLevels {
		fmt.Printf("  Testing %d concurrent workers... ", concurrency)
		result := runLoadTest(serverURL, requests, concurrency, asyncMode)
		fmt.Printf("RPS: %.2f, Avg Latency: %v\n", result.RequestsPerSecond, result.AverageLatency)
	}
}

// testCachePerformance tests cache operations performance
func testCachePerformance(serverURL string) {
	fmt.Println("Testing cache performance by creating and retrieving orders...")
	
	// Create several orders first
	client := &http.Client{Timeout: time.Second * 5}
	orderIDs := make([]string, 10)
	
	for i := 0; i < 10; i++ {
		order := TestOrder{
			CustomerName: fmt.Sprintf("CacheTest-%d", i),
			CoffeeType:   "Latte",
		}
		
		jsonData, _ := json.Marshal(order)
		resp, err := client.Post(serverURL+"/order", "application/json", bytes.NewBuffer(jsonData))
		
		if err == nil && resp.StatusCode == http.StatusOK {
			// Extract order ID from response (simplified)
			var orderResp map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&orderResp)
			if order, ok := orderResp["order"].(map[string]interface{}); ok {
				if id, ok := order["id"].(string); ok {
					orderIDs[i] = id
				}
			}
			resp.Body.Close()
		}
	}
	
	// Test retrieval performance
	start := time.Now()
	successful := 0
	
	for _, orderID := range orderIDs {
		if orderID != "" {
			resp, err := client.Get(serverURL + "/order/" + orderID)
			if err == nil && resp.StatusCode == http.StatusOK {
				successful++
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
	}
	
	duration := time.Since(start)
	fmt.Printf("  Retrieved %d orders in %v (avg: %v per order)\n", 
		successful, duration, duration/time.Duration(successful))
}