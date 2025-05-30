package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// MCPAIIntegration integrates Redis MCP with AI Search
type MCPAIIntegration struct {
	redis         *redis.Client
	aiSearchURL   string
	mcpServerURL  string
	router        *gin.Engine
	httpClient    *http.Client
}

// MCPAIRequest represents an enhanced MCP request with AI capabilities
type MCPAIRequest struct {
	Query       string                 `json:"query"`
	SearchType  string                 `json:"search_type,omitempty"`  // "semantic", "vector", "hybrid"
	UseAI       bool                   `json:"use_ai,omitempty"`       // Enable AI-powered search
	Context     map[string]interface{} `json:"context,omitempty"`
	AgentID     string                 `json:"agent_id"`
	UserID      string                 `json:"user_id,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// MCPAIResponse represents an enhanced MCP response with AI results
type MCPAIResponse struct {
	Success       bool                   `json:"success"`
	Data          interface{}            `json:"data,omitempty"`
	AIResults     []AISearchResult       `json:"ai_results,omitempty"`
	Suggestions   []string               `json:"suggestions,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Query         string                 `json:"executed_query,omitempty"`
	SearchType    string                 `json:"search_type,omitempty"`
	QueryTime     time.Duration          `json:"query_time,omitempty"`
	AIEnhanced    bool                   `json:"ai_enhanced"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}

// AISearchResult represents AI search result
type AISearchResult struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Score       float64                `json:"score"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewMCPAIIntegration creates a new MCP-AI integration
func NewMCPAIIntegration(redisClient *redis.Client, aiSearchURL, mcpServerURL string) *MCPAIIntegration {
	integration := &MCPAIIntegration{
		redis:        redisClient,
		aiSearchURL:  aiSearchURL,
		mcpServerURL: mcpServerURL,
		router:       gin.New(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	integration.setupRoutes()
	return integration
}

func (m *MCPAIIntegration) setupRoutes() {
	m.router.Use(gin.Recovery())
	m.router.Use(m.corsMiddleware())

	api := m.router.Group("/api/v1/mcp-ai")
	{
		api.POST("/query", m.handleEnhancedQuery)
		api.POST("/smart-search", m.handleSmartSearch)
		api.GET("/recommendations/:user_id", m.handleRecommendations)
		api.GET("/trending", m.handleTrending)
		api.GET("/health", m.handleHealth)
		api.GET("/demo", m.handleDemo)
	}
}

// handleEnhancedQuery handles enhanced MCP queries with AI capabilities
func (m *MCPAIIntegration) handleEnhancedQuery(c *gin.Context) {
	startTime := time.Now()
	
	var req MCPAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	req.Timestamp = time.Now()

	log.Printf("🤖 Processing enhanced MCP-AI query: %s (Agent: %s, AI: %v)", 
		req.Query, req.AgentID, req.UseAI)

	var response MCPAIResponse
	
	if req.UseAI {
		// Use AI-powered search
		aiResults, suggestions, err := m.performAISearch(c.Request.Context(), req)
		if err != nil {
			log.Printf("❌ AI search failed: %v", err)
			// Fallback to traditional MCP
			mcpResult, mcpErr := m.performTraditionalMCP(c.Request.Context(), req)
			if mcpErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Both AI and MCP search failed"})
				return
			}
			
			response = MCPAIResponse{
				Success:    true,
				Data:       mcpResult,
				AIEnhanced: false,
				SearchType: "mcp_fallback",
				QueryTime:  time.Since(startTime),
				Metadata: map[string]interface{}{
					"ai_search_failed": true,
					"fallback_used":    true,
				},
				Timestamp: time.Now(),
			}
		} else {
			// AI search successful
			response = MCPAIResponse{
				Success:     true,
				AIResults:   aiResults,
				Suggestions: suggestions,
				AIEnhanced:  true,
				SearchType:  req.SearchType,
				QueryTime:   time.Since(startTime),
				Metadata: map[string]interface{}{
					"ai_powered":       true,
					"results_count":    len(aiResults),
					"suggestions_count": len(suggestions),
				},
				Timestamp: time.Now(),
			}
		}
	} else {
		// Traditional MCP search
		mcpResult, err := m.performTraditionalMCP(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "MCP search failed"})
			return
		}
		
		response = MCPAIResponse{
			Success:    true,
			Data:       mcpResult,
			AIEnhanced: false,
			SearchType: "traditional_mcp",
			QueryTime:  time.Since(startTime),
			Metadata: map[string]interface{}{
				"mcp_powered": true,
			},
			Timestamp: time.Now(),
		}
	}

	log.Printf("✅ Enhanced MCP-AI query completed in %v", response.QueryTime)
	c.JSON(http.StatusOK, response)
}

// handleSmartSearch handles intelligent search that automatically chooses best method
func (m *MCPAIIntegration) handleSmartSearch(c *gin.Context) {
	startTime := time.Now()
	
	var req MCPAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	log.Printf("🧠 Smart search query: %s", req.Query)

	// Intelligent decision: when to use AI vs MCP
	useAI := m.shouldUseAI(req.Query)
	req.UseAI = useAI
	
	if useAI {
		req.SearchType = m.determineOptimalSearchType(req.Query)
	}

	// Execute search
	var response MCPAIResponse
	
	if useAI {
		aiResults, suggestions, err := m.performAISearch(c.Request.Context(), req)
		if err == nil {
			response = MCPAIResponse{
				Success:     true,
				AIResults:   aiResults,
				Suggestions: suggestions,
				AIEnhanced:  true,
				SearchType:  req.SearchType,
				QueryTime:   time.Since(startTime),
				Metadata: map[string]interface{}{
					"smart_decision": "ai_search",
					"reason":         "semantic_query_detected",
				},
				Timestamp: time.Now(),
			}
		} else {
			// Fallback to MCP
			mcpResult, _ := m.performTraditionalMCP(c.Request.Context(), req)
			response = MCPAIResponse{
				Success:    true,
				Data:       mcpResult,
				AIEnhanced: false,
				SearchType: "mcp_fallback",
				QueryTime:  time.Since(startTime),
				Metadata: map[string]interface{}{
					"smart_decision": "mcp_fallback",
					"reason":         "ai_search_failed",
				},
				Timestamp: time.Now(),
			}
		}
	} else {
		mcpResult, err := m.performTraditionalMCP(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Smart search failed"})
			return
		}
		
		response = MCPAIResponse{
			Success:    true,
			Data:       mcpResult,
			AIEnhanced: false,
			SearchType: "traditional_mcp",
			QueryTime:  time.Since(startTime),
			Metadata: map[string]interface{}{
				"smart_decision": "traditional_mcp",
				"reason":         "structured_query_detected",
			},
			Timestamp: time.Now(),
		}
	}

	log.Printf("🎯 Smart search completed in %v (Method: %s)", 
		response.QueryTime, response.Metadata["smart_decision"])
	c.JSON(http.StatusOK, response)
}

// handleRecommendations gets AI-powered recommendations
func (m *MCPAIIntegration) handleRecommendations(c *gin.Context) {
	userID := c.Param("user_id")
	
	// Get personalized recommendations from AI Search
	url := fmt.Sprintf("%s/api/v1/ai-search/personalized/%s", m.aiSearchURL, userID)
	resp, err := m.httpClient.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recommendations"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse recommendations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"recommendations": result["recommendations"],
		"user_id":         userID,
		"ai_powered":      true,
		"timestamp":       time.Now(),
	})
}

// handleTrending gets trending items
func (m *MCPAIIntegration) handleTrending(c *gin.Context) {
	url := fmt.Sprintf("%s/api/v1/ai-search/trending", m.aiSearchURL)
	resp, err := m.httpClient.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trending"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse trending"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"trending":   result["trending"],
		"ai_powered": true,
		"timestamp":  time.Now(),
	})
}

// performAISearch performs AI-powered search
func (m *MCPAIIntegration) performAISearch(ctx context.Context, req MCPAIRequest) ([]AISearchResult, []string, error) {
	searchType := req.SearchType
	if searchType == "" {
		searchType = "semantic" // Default to semantic search
	}

	// Prepare AI search request
	aiRequest := map[string]interface{}{
		"query":     req.Query,
		"limit":     10,
		"threshold": 0.3,
	}

	if req.UserID != "" {
		aiRequest["user_id"] = req.UserID
	}

	jsonData, err := json.Marshal(aiRequest)
	if err != nil {
		return nil, nil, err
	}

	// Call AI Search API
	url := fmt.Sprintf("%s/api/v1/ai-search/%s", m.aiSearchURL, searchType)
	resp, err := m.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	var aiResponse struct {
		Success     bool              `json:"success"`
		Results     []AISearchResult  `json:"results"`
		Suggestions []string          `json:"suggestions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&aiResponse); err != nil {
		return nil, nil, err
	}

	if !aiResponse.Success {
		return nil, nil, fmt.Errorf("AI search failed")
	}

	return aiResponse.Results, aiResponse.Suggestions, nil
}

// performTraditionalMCP performs traditional MCP search
func (m *MCPAIIntegration) performTraditionalMCP(ctx context.Context, req MCPAIRequest) (interface{}, error) {
	// Call traditional MCP server
	mcpRequest := map[string]interface{}{
		"query":     req.Query,
		"agent_id":  req.AgentID,
		"timestamp": req.Timestamp,
	}

	jsonData, err := json.Marshal(mcpRequest)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v1/redis-mcp/query", m.mcpServerURL)
	resp, err := m.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var mcpResponse struct {
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
		Error   string      `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&mcpResponse); err != nil {
		return nil, err
	}

	if !mcpResponse.Success {
		return nil, fmt.Errorf("MCP search failed: %s", mcpResponse.Error)
	}

	return mcpResponse.Data, nil
}

// shouldUseAI determines whether to use AI search based on query characteristics
func (m *MCPAIIntegration) shouldUseAI(query string) bool {
	queryLower := strings.ToLower(query)

	// Use AI for semantic/descriptive queries
	semanticKeywords := []string{
		"strong", "mild", "sweet", "bitter", "smooth", "rich", "creamy",
		"refreshing", "hot", "cold", "best", "recommend", "like", "similar",
		"taste", "flavor", "mood", "feeling", "want", "need", "prefer",
	}

	for _, keyword := range semanticKeywords {
		if strings.Contains(queryLower, keyword) {
			return true
		}
	}

	// Use traditional MCP for structured queries
	structuredKeywords := []string{
		"get menu", "get inventory", "get customer", "get orders",
		"add", "update", "delete", "set", "create",
	}

	for _, keyword := range structuredKeywords {
		if strings.Contains(queryLower, keyword) {
			return false
		}
	}

	// Default to AI for ambiguous queries
	return true
}

// determineOptimalSearchType determines the best AI search type
func (m *MCPAIIntegration) determineOptimalSearchType(query string) string {
	queryLower := strings.ToLower(query)

	// Use vector search for very specific descriptions
	if strings.Contains(queryLower, "exactly like") || strings.Contains(queryLower, "similar to") {
		return "vector"
	}

	// Use hybrid search for complex queries
	if len(strings.Fields(query)) > 5 {
		return "hybrid"
	}

	// Default to semantic search
	return "semantic"
}

// handleHealth returns health status
func (m *MCPAIIntegration) handleHealth(c *gin.Context) {
	// Check Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.redis.Ping(ctx).Result()
	redisHealthy := err == nil

	// Check AI Search service
	aiHealthy := false
	if resp, err := m.httpClient.Get(fmt.Sprintf("%s/api/v1/ai-search/health", m.aiSearchURL)); err == nil {
		aiHealthy = resp.StatusCode == 200
		resp.Body.Close()
	}

	// Check MCP service
	mcpHealthy := false
	if resp, err := m.httpClient.Get(fmt.Sprintf("%s/api/v1/redis-mcp/health", m.mcpServerURL)); err == nil {
		mcpHealthy = resp.StatusCode == 200
		resp.Body.Close()
	}

	status := "healthy"
	if !redisHealthy || !aiHealthy || !mcpHealthy {
		status = "degraded"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      status,
		"redis":       redisHealthy,
		"ai_search":   aiHealthy,
		"mcp_server":  mcpHealthy,
		"integration": "operational",
		"timestamp":   time.Now(),
		"version":     "mcp-ai-v1.0.0",
	})
}

// handleDemo provides demo endpoints and examples
func (m *MCPAIIntegration) handleDemo(c *gin.Context) {
	examples := []map[string]interface{}{
		{
			"name":        "Semantic Coffee Search",
			"description": "Find coffee based on taste preferences",
			"endpoint":    "POST /api/v1/mcp-ai/query",
			"example": map[string]interface{}{
				"query":       "I want something strong but not too bitter",
				"use_ai":      true,
				"search_type": "semantic",
				"agent_id":    "demo-agent",
			},
		},
		{
			"name":        "Smart Search",
			"description": "Automatically chooses best search method",
			"endpoint":    "POST /api/v1/mcp-ai/smart-search",
			"example": map[string]interface{}{
				"query":    "refreshing cold drink for summer",
				"agent_id": "demo-agent",
			},
		},
		{
			"name":        "Traditional MCP",
			"description": "Structured data queries",
			"endpoint":    "POST /api/v1/mcp-ai/query",
			"example": map[string]interface{}{
				"query":    "get menu for shop downtown",
				"use_ai":   false,
				"agent_id": "demo-agent",
			},
		},
		{
			"name":        "Personalized Recommendations",
			"description": "AI-powered user recommendations",
			"endpoint":    "GET /api/v1/mcp-ai/recommendations/{user_id}",
			"example":     "GET /api/v1/mcp-ai/recommendations/user123",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"title":       "MCP-AI Integration Demo",
		"description": "Combines Redis MCP with AI Search for ultimate coffee experience",
		"features": []string{
			"Semantic Search",
			"Vector Similarity",
			"Hybrid Search",
			"Smart Query Routing",
			"Personalized Recommendations",
			"Traditional MCP Fallback",
		},
		"examples":  examples,
		"timestamp": time.Now(),
	})
}

func (m *MCPAIIntegration) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func (m *MCPAIIntegration) Start(port string) error {
	log.Printf("🚀 Starting MCP-AI Integration Server on port %s", port)
	return m.router.Run(":" + port)
}
