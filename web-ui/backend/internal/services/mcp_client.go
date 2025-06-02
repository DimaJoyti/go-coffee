package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// RateLimiter implements a simple rate limiter
type RateLimiter struct {
	tokens   chan struct{}
	interval time.Duration
	mu       sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		tokens:   make(chan struct{}, limit),
		interval: interval,
	}

	// Fill the bucket initially
	for i := 0; i < limit; i++ {
		rl.tokens <- struct{}{}
	}

	// Refill tokens periodically
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			select {
			case rl.tokens <- struct{}{}:
			default:
				// Bucket is full
			}
		}
	}()

	return rl
}

// Wait waits for a token to become available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// MCPClient handles communication with MCP servers
type MCPClient struct {
	client      *http.Client
	baseURL     string
	rateLimiter *RateLimiter
	cache       map[string]cacheEntry
	cacheMu     sync.RWMutex
	cacheTTL    time.Duration
}

type cacheEntry struct {
	data      interface{}
	timestamp time.Time
}

// MCPRequest represents a request to an MCP server
type MCPRequest struct {
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

// MCPResponse represents a response from an MCP server
type MCPResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  *MCPError   `json:"error,omitempty"`
}

// MCPError represents an error from an MCP server
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewMCPClient creates a new MCP client with rate limiting and caching
func NewMCPClient() *MCPClient {
	// Get rate limit from environment or use default
	rateLimit := 10
	if rateLimitStr := os.Getenv("BRIGHT_DATA_RATE_LIMIT"); rateLimitStr != "" {
		if limit, err := strconv.Atoi(rateLimitStr); err == nil {
			rateLimit = limit
		}
	}

	// Get cache TTL from environment or use default
	cacheTTL := 5 * time.Minute
	if cacheTTLStr := os.Getenv("BRIGHT_DATA_CACHE_TTL"); cacheTTLStr != "" {
		if ttl, err := strconv.Atoi(cacheTTLStr); err == nil {
			cacheTTL = time.Duration(ttl) * time.Second
		}
	}

	return &MCPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:     os.Getenv("MCP_SERVER_URL"),
		rateLimiter: NewRateLimiter(rateLimit, time.Minute),
		cache:       make(map[string]cacheEntry),
		cacheTTL:    cacheTTL,
	}
}

// CallMCP makes a call to an MCP server with caching and rate limiting
func (c *MCPClient) CallMCP(method string, params interface{}) (*MCPResponse, error) {
	return c.CallMCPWithContext(context.Background(), method, params)
}

// CallMCPWithContext makes a call to an MCP server with context, caching and rate limiting
func (c *MCPClient) CallMCPWithContext(ctx context.Context, method string, params interface{}) (*MCPResponse, error) {
	// Create cache key
	cacheKey := c.createCacheKey(method, params)

	// Check cache first
	if cachedResp := c.getFromCache(cacheKey); cachedResp != nil {
		log.Printf("Cache hit for method: %s", method)
		return cachedResp, nil
	}

	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	request := MCPRequest{
		Method: method,
		Params: params,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("Making MCP call: %s", method)
	resp, err := c.client.Post(c.baseURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var mcpResp MCPResponse
	if err := json.Unmarshal(body, &mcpResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if mcpResp.Error != nil {
		return nil, fmt.Errorf("MCP error %d: %s", mcpResp.Error.Code, mcpResp.Error.Message)
	}

	// Cache successful response
	c.setCache(cacheKey, &mcpResp)

	return &mcpResp, nil
}

// BrightDataMCPService integrates with Bright Data through MCP
type BrightDataMCPService struct {
	mcpClient *MCPClient
}

// NewBrightDataMCPService creates a new Bright Data MCP service
func NewBrightDataMCPService() *BrightDataMCPService {
	return &BrightDataMCPService{
		mcpClient: NewMCPClient(),
	}
}

// ScrapeURL scrapes a URL using Bright Data MCP
func (s *BrightDataMCPService) ScrapeURL(url string) (*ScrapingResponse, error) {
	params := map[string]interface{}{
		"url": url,
	}

	resp, err := s.mcpClient.CallMCP("scrape_as_markdown_Bright_Data", params)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape URL: %w", err)
	}

	return &ScrapingResponse{
		Success: true,
		Data:    resp.Result,
	}, nil
}

// SearchEngine performs a search using Bright Data MCP
func (s *BrightDataMCPService) SearchEngine(query string, engine string) (*ScrapingResponse, error) {
	params := map[string]interface{}{
		"query":  query,
		"engine": engine, // google, bing, yandex
	}

	resp, err := s.mcpClient.CallMCP("search_engine_Bright_Data", params)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	return &ScrapingResponse{
		Success: true,
		Data:    resp.Result,
	}, nil
}

// GetSessionStats gets session statistics from Bright Data MCP
func (s *BrightDataMCPService) GetSessionStats() (*ScrapingResponse, error) {
	resp, err := s.mcpClient.CallMCP("session_stats_Bright_Data", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get session stats: %w", err)
	}

	return &ScrapingResponse{
		Success: true,
		Data:    resp.Result,
	}, nil
}

// Cache helper methods
func (c *MCPClient) createCacheKey(method string, params interface{}) string {
	paramsJSON, _ := json.Marshal(params)
	return fmt.Sprintf("%s:%s", method, string(paramsJSON))
}

func (c *MCPClient) getFromCache(key string) *MCPResponse {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil
	}

	// Check if cache entry is expired
	if time.Since(entry.timestamp) > c.cacheTTL {
		// Remove expired entry
		delete(c.cache, key)
		return nil
	}

	if resp, ok := entry.data.(*MCPResponse); ok {
		return resp
	}

	return nil
}

func (c *MCPClient) setCache(key string, response *MCPResponse) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	c.cache[key] = cacheEntry{
		data:      response,
		timestamp: time.Now(),
	}
}

// ClearCache clears all cached entries
func (c *MCPClient) ClearCache() {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	c.cache = make(map[string]cacheEntry)
}
