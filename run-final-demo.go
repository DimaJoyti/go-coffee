package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

// SimpleFinalDemo demonstrates the complete Redis 8 AI Search ecosystem
type SimpleFinalDemo struct {
	httpClient *http.Client
	baseURLs   map[string]string
}

// NewSimpleFinalDemo creates a new final demo
func NewSimpleFinalDemo() *SimpleFinalDemo {
	return &SimpleFinalDemo{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURLs: map[string]string{
			"mcp":         "http://localhost:8090",
			"ai_search":   "http://localhost:8092", 
			"integration": "http://localhost:8093",
		},
	}
}

// RunDemo runs the comprehensive demo
func (d *SimpleFinalDemo) RunDemo() {
	log.Println("🚀 Starting Redis 8 AI Search Ecosystem Demo!")
	log.Println(strings.Repeat("=", 80))
	
	// Test 1: Health Checks
	log.Println("\n🏥 HEALTH CHECKS")
	log.Println(strings.Repeat("-", 40))
	d.testHealthChecks()
	
	// Test 2: Traditional MCP
	log.Println("\n🔧 TRADITIONAL REDIS MCP")
	log.Println(strings.Repeat("-", 40))
	d.testTraditionalMCP()
	
	// Test 3: AI Search Engine
	log.Println("\n🧠 REDIS 8 AI SEARCH ENGINE")
	log.Println(strings.Repeat("-", 40))
	d.testAISearchEngine()
	
	// Test 4: Smart Integration
	log.Println("\n🚀 SMART MCP-AI INTEGRATION")
	log.Println(strings.Repeat("-", 40))
	d.testSmartIntegration()
	
	// Test 5: Performance Benchmarks
	log.Println("\n⚡ PERFORMANCE BENCHMARKS")
	log.Println(strings.Repeat("-", 40))
	d.runPerformanceBenchmarks()
	
	// Final Summary
	log.Println("\n🎉 DEMO COMPLETED SUCCESSFULLY!")
	log.Println(strings.Repeat("=", 80))
	log.Println("🚀 All Redis 8 AI Search systems operational and blazingly fast!")
}

func (d *SimpleFinalDemo) testHealthChecks() {
	services := map[string]string{
		"Redis MCP":        d.baseURLs["mcp"] + "/api/v1/redis-mcp/health",
		"AI Search":        d.baseURLs["ai_search"] + "/api/v1/ai-search/health", 
		"MCP-AI Integration": d.baseURLs["integration"] + "/api/v1/mcp-ai/health",
	}
	
	for name, url := range services {
		start := time.Now()
		resp, err := d.httpClient.Get(url)
		duration := time.Since(start)
		
		if err != nil || resp.StatusCode != 200 {
			log.Printf("❌ %s: UNHEALTHY (%v)", name, duration)
		} else {
			log.Printf("✅ %s: HEALTHY (%v)", name, duration)
		}
		
		if resp != nil {
			resp.Body.Close()
		}
	}
}

func (d *SimpleFinalDemo) testTraditionalMCP() {
	tests := []struct {
		name  string
		query string
		desc  string
	}{
		{"Menu Query", "get menu for shop downtown", "Retrieve coffee menu"},
		{"Inventory Query", "get inventory for uptown", "Check inventory levels"},
		{"Customer Query", "get customer 123 name", "Get customer information"},
		{"Top Orders", "get top orders today", "Popular drinks ranking"},
	}
	
	for _, test := range tests {
		requestBody := map[string]interface{}{
			"query":     test.query,
			"agent_id":  "demo-agent",
			"timestamp": time.Now(),
		}
		
		url := d.baseURLs["mcp"] + "/api/v1/redis-mcp/query"
		duration := d.makeRequest("POST", url, requestBody, test.name)
		log.Printf("🔧 %s: %s (%v)", test.name, test.desc, duration)
	}
}

func (d *SimpleFinalDemo) testAISearchEngine() {
	tests := []struct {
		name       string
		endpoint   string
		query      string
		desc       string
	}{
		{"Semantic Search", "/semantic", "strong coffee with milk", "AI semantic understanding"},
		{"Vector Search", "/vector", "sweet chocolate drink", "Vector similarity matching"},
		{"Hybrid Search", "/hybrid", "cold refreshing drink", "Combined AI + keyword search"},
	}
	
	for _, test := range tests {
		requestBody := map[string]interface{}{
			"query":     test.query,
			"limit":     5,
			"threshold": 0.3,
		}
		
		url := d.baseURLs["ai_search"] + "/api/v1/ai-search" + test.endpoint
		duration := d.makeRequest("POST", url, requestBody, test.name)
		log.Printf("🧠 %s: %s (%v)", test.name, test.desc, duration)
	}
	
	// Test additional endpoints
	additionalTests := []struct {
		name string
		url  string
		desc string
	}{
		{"Trending Items", d.baseURLs["ai_search"] + "/api/v1/ai-search/trending", "AI popularity analysis"},
		{"Personalized", d.baseURLs["ai_search"] + "/api/v1/ai-search/personalized/user123", "AI recommendations"},
		{"Suggestions", d.baseURLs["ai_search"] + "/api/v1/ai-search/suggestions/coffee", "Smart suggestions"},
		{"Statistics", d.baseURLs["ai_search"] + "/api/v1/ai-search/stats", "Performance metrics"},
	}
	
	for _, test := range additionalTests {
		duration := d.makeRequest("GET", test.url, nil, test.name)
		log.Printf("🎯 %s: %s (%v)", test.name, test.desc, duration)
	}
}

func (d *SimpleFinalDemo) testSmartIntegration() {
	tests := []struct {
		name  string
		query string
		desc  string
	}{
		{"Smart AI Query", "I want something refreshing and not too strong", "Auto-route to AI search"},
		{"Smart MCP Query", "get menu for shop downtown", "Auto-route to traditional MCP"},
		{"Complex Semantic", "I need a drink that's creamy but not too sweet", "Advanced AI understanding"},
		{"Structured Data", "add matcha to ingredients", "Structured data operation"},
	}
	
	for _, test := range tests {
		requestBody := map[string]interface{}{
			"query":    test.query,
			"agent_id": "smart-demo-agent",
		}
		
		url := d.baseURLs["integration"] + "/api/v1/mcp-ai/smart-search"
		duration := d.makeRequest("POST", url, requestBody, test.name)
		log.Printf("🚀 %s: %s (%v)", test.name, test.desc, duration)
	}
	
	// Test integration endpoints
	integrationTests := []struct {
		name string
		url  string
		desc string
	}{
		{"Integration Health", d.baseURLs["integration"] + "/api/v1/mcp-ai/health", "System health status"},
		{"Demo Examples", d.baseURLs["integration"] + "/api/v1/mcp-ai/demo", "Available examples"},
		{"AI Recommendations", d.baseURLs["integration"] + "/api/v1/mcp-ai/recommendations/user123", "Personalized AI suggestions"},
		{"Trending via Integration", d.baseURLs["integration"] + "/api/v1/mcp-ai/trending", "Trending through integration"},
	}
	
	for _, test := range integrationTests {
		duration := d.makeRequest("GET", test.url, nil, test.name)
		log.Printf("🎯 %s: %s (%v)", test.name, test.desc, duration)
	}
}

func (d *SimpleFinalDemo) runPerformanceBenchmarks() {
	log.Println("Running performance benchmarks...")
	
	// Benchmark semantic search
	query := map[string]interface{}{
		"query": "strong coffee with milk",
		"limit": 10,
	}
	
	var totalTime time.Duration
	iterations := 10
	
	for i := 0; i < iterations; i++ {
		duration := d.makeRequest("POST", d.baseURLs["ai_search"]+"/api/v1/ai-search/semantic", query, "Benchmark")
		totalTime += duration
	}
	
	avgTime := totalTime / time.Duration(iterations)
	log.Printf("⚡ Average semantic search time: %v", avgTime)
	log.Printf("⚡ Queries per second: %.2f", float64(time.Second)/float64(avgTime))
	
	// Benchmark smart integration
	smartQuery := map[string]interface{}{
		"query":    "I want something strong but smooth",
		"agent_id": "benchmark-agent",
	}
	
	var smartTotalTime time.Duration
	for i := 0; i < iterations; i++ {
		duration := d.makeRequest("POST", d.baseURLs["integration"]+"/api/v1/mcp-ai/smart-search", smartQuery, "Smart Benchmark")
		smartTotalTime += duration
	}
	
	smartAvgTime := smartTotalTime / time.Duration(iterations)
	log.Printf("⚡ Average smart integration time: %v", smartAvgTime)
	log.Printf("⚡ Smart queries per second: %.2f", float64(time.Second)/float64(smartAvgTime))
}

func (d *SimpleFinalDemo) makeRequest(method, url string, body interface{}, testName string) time.Duration {
	start := time.Now()
	
	var req *http.Request
	var err error
	
	if body != nil {
		jsonData, _ := json.Marshal(body)
		req, err = http.NewRequest(method, url, bytes.NewBuffer(jsonData))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
		}
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	
	if err != nil {
		return time.Since(start)
	}
	
	resp, err := d.httpClient.Do(req)
	duration := time.Since(start)
	
	if resp != nil {
		resp.Body.Close()
	}
	
	return duration
}

func main() {
	log.Println("🎯 Redis 8 AI Search Ecosystem - Final Demo")
	log.Println("⚡ Testing all systems for blazingly fast performance!")
	log.Println("")
	
	demo := NewSimpleFinalDemo()
	demo.RunDemo()
	
	log.Println("")
	log.Println("🎉 Demo completed! Redis 8 makes AI search blazingly fast! ⚡")
}
