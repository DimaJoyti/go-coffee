package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// TestRequest represents a test request
type TestRequest struct {
	Name        string                 `json:"name"`
	Method      string                 `json:"method"`
	URL         string                 `json:"url"`
	Body        map[string]interface{} `json:"body,omitempty"`
	ExpectedStatus int                 `json:"expected_status"`
}

// TestResult represents a test result
type TestResult struct {
	Name       string        `json:"name"`
	Success    bool          `json:"success"`
	StatusCode int           `json:"status_code"`
	Duration   time.Duration `json:"duration"`
	Error      string        `json:"error,omitempty"`
	Response   interface{}   `json:"response,omitempty"`
}

func getSeparator() string {
	return strings.Repeat("=", 60)
}

func main() {
	fmt.Println("🚀 Testing Enhanced Bright Data Hub Integration")
	fmt.Println(getSeparator())
	
	baseURL := "http://localhost:8095"
	if len(baseURL) == 0 {
		fmt.Println("❌ Error: baseURL is not configured")
		return
	}
	
	// Test cases
	tests := []TestRequest{
		{
			Name:           "Health Check",
			Method:         "GET",
			URL:            baseURL + "/api/v1/bright-data/health",
			ExpectedStatus: 200,
		},
		{
			Name:           "Status Check",
			Method:         "GET",
			URL:            baseURL + "/api/v1/bright-data/status",
			ExpectedStatus: 200,
		},
		{
			Name:   "Search Engine - Google",
			Method: "POST",
			URL:    baseURL + "/api/v1/bright-data/search/engine",
			Body: map[string]interface{}{
				"query":  "coffee market trends 2025",
				"engine": "google",
			},
			ExpectedStatus: 200,
		},
		{
			Name:   "Scrape URL",
			Method: "POST",
			URL:    baseURL + "/api/v1/bright-data/search/scrape",
			Body: map[string]interface{}{
				"url": "https://example.com",
			},
			ExpectedStatus: 200,
		},
		{
			Name:   "Instagram Profile",
			Method: "POST",
			URL:    baseURL + "/api/v1/bright-data/social/instagram/profile",
			Body: map[string]interface{}{
				"url": "https://instagram.com/starbucks",
			},
			ExpectedStatus: 200,
		},
		{
			Name:   "Facebook Posts",
			Method: "POST",
			URL:    baseURL + "/api/v1/bright-data/social/facebook/posts",
			Body: map[string]interface{}{
				"url": "https://facebook.com/starbucks",
			},
			ExpectedStatus: 200,
		},
		{
			Name:   "Twitter Posts",
			Method: "POST",
			URL:    baseURL + "/api/v1/bright-data/social/twitter/posts",
			Body: map[string]interface{}{
				"url": "https://twitter.com/starbucks",
			},
			ExpectedStatus: 200,
		},
		{
			Name:   "LinkedIn Profile",
			Method: "POST",
			URL:    baseURL + "/api/v1/bright-data/social/linkedin/profile",
			Body: map[string]interface{}{
				"url": "https://linkedin.com/company/starbucks",
			},
			ExpectedStatus: 200,
		},
		{
			Name:   "Amazon Product",
			Method: "POST",
			URL:    baseURL + "/api/v1/bright-data/ecommerce/amazon/product",
			Body: map[string]interface{}{
				"url": "https://amazon.com/dp/B08N5WRWNW",
			},
			ExpectedStatus: 200,
		},
		{
			Name:   "Execute Function - Direct",
			Method: "POST",
			URL:    baseURL + "/api/v1/bright-data/execute",
			Body: map[string]interface{}{
				"function": "search_engine_Bright_Data",
				"params": map[string]interface{}{
					"query":  "coffee",
					"engine": "google",
				},
			},
			ExpectedStatus: 200,
		},
		{
			Name:           "Social Analytics",
			Method:         "GET",
			URL:            baseURL + "/api/v1/bright-data/social/analytics",
			ExpectedStatus: 200,
		},
		{
			Name:           "Trending Topics",
			Method:         "GET",
			URL:            baseURL + "/api/v1/bright-data/social/trending",
			ExpectedStatus: 200,
		},
	}

	// Run tests
	results := make([]TestResult, 0, len(tests))
	successCount := 0
	fmt.Println(getSeparator())
	for i, test := range tests {
		fmt.Printf("\n%d. Testing: %s\n", i+1, test.Name)
		fmt.Printf("   %s %s\n", test.Method, test.URL)
		
		result := runTest(test)
		results = append(results, result)
		if result.Success {
			fmt.Printf("   ✅ SUCCESS (%d) - %v\n", result.StatusCode, result.Duration)
			successCount++
		} else {
			fmt.Printf("   ❌ FAILED (%d) - %s\n", result.StatusCode, result.Error)
		}
		// Small delay between tests
		time.Sleep(100 * time.Millisecond)
	}

	// Summary
	fmt.Println(getSeparator())
	fmt.Printf("🎯 TEST SUMMARY\n")
	fmt.Printf("Total Tests: %d\n", len(tests))
	fmt.Printf("Passed: %d\n", successCount)
	fmt.Printf("Failed: %d\n", len(tests)-successCount)
	fmt.Printf("Success Rate: %.1f%%\n", float64(successCount)/float64(len(tests))*100)

	// Detailed results
	fmt.Println("\n📊 DETAILED RESULTS:")
	for _, result := range results {
		status := "✅"
		if !result.Success {
			status = "❌"
		}
		fmt.Printf("%s %s (%v)\n", status, result.Name, result.Duration)
		if result.Error != "" {
			fmt.Printf("   Error: %s\n", result.Error)
		}
	}

	// Performance analysis
	fmt.Println("\n⚡ PERFORMANCE ANALYSIS:")
	var totalDuration time.Duration
	var maxDuration time.Duration
	var minDuration time.Duration = time.Hour
	for _, result := range results {
		totalDuration += result.Duration
		if result.Duration > maxDuration {
			maxDuration = result.Duration
		}
		if result.Duration < minDuration {
			minDuration = result.Duration
		}
	}
	avgDuration := totalDuration / time.Duration(len(results))
	fmt.Printf("Average Response Time: %v\n", avgDuration)
	fmt.Printf("Fastest Response: %v\n", minDuration)
	fmt.Printf("Slowest Response: %v\n", maxDuration)

	if successCount == len(tests) {
		fmt.Println("\n🎉 ALL TESTS PASSED! Bright Data Hub is working perfectly!")
	} else {
		fmt.Printf("\n⚠️  %d tests failed. Please check the service configuration.\n", len(tests)-successCount)
	}
}

func runTest(test TestRequest) TestResult {
	start := time.Now()
	var body io.Reader
	if test.Body != nil {
		jsonBody, err := json.Marshal(test.Body)
		if err != nil {
			return TestResult{
				Name:    test.Name,
				Success: false,
				Error:   fmt.Sprintf("Failed to marshal request body: %v", err),
				Duration: time.Since(start),
			}
		}
		body = bytes.NewBuffer(jsonBody)
	}
	
	req, err := http.NewRequest(test.Method, test.URL, body)
	if err != nil {
		return TestResult{
			Name:    test.Name,
			Success: false,
			Error:   fmt.Sprintf("Failed to create request: %v", err),
			Duration: time.Since(start),
		}
	}
	
	if test.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return TestResult{
			Name:       test.Name,
			Success:    false,
			StatusCode: 0,
			Error:      fmt.Sprintf("Request failed: %v", err),
			Duration:   time.Since(start),
		}
	}
	defer resp.Body.Close()
	
	duration := time.Since(start)
	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return TestResult{
			Name:       test.Name,
			Success:    false,
			StatusCode: resp.StatusCode,
			Error:      fmt.Sprintf("Failed to read response: %v", err),
			Duration:   duration,
		}
	}
	
	var responseData interface{}
	if len(respBody) > 0 {
		json.Unmarshal(respBody, &responseData)
	}
	
	success := resp.StatusCode == test.ExpectedStatus
	errorMsg := ""
	if !success {
		errorMsg = fmt.Sprintf("Expected status %d, got %d", test.ExpectedStatus, resp.StatusCode)
		if len(respBody) > 0 && len(respBody) < 500 {
			errorMsg += fmt.Sprintf(" - Response: %s", string(respBody))
		}
	}
	
	return TestResult{
		Name:       test.Name,
		Success:    success,
		StatusCode: resp.StatusCode,
		Error:      errorMsg,
		Duration:   duration,
		Response:   responseData,
	}
}
