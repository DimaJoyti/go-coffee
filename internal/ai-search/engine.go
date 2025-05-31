package aisearch

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// Redis8AISearchEngine implements blazingly fast AI search using Redis 8
type Redis8AISearchEngine struct {
	client *redis.Client
	router *gin.Engine
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query       string            `json:"query" binding:"required"`
	Limit       int               `json:"limit,omitempty"`
	Offset      int               `json:"offset,omitempty"`
	Filters     map[string]string `json:"filters,omitempty"`
	UserID      string            `json:"user_id,omitempty"`
	Categories  []string          `json:"categories,omitempty"`
	MinScore    float64           `json:"min_score,omitempty"`
}

// SearchResult represents a search result
type SearchResult struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Price       float64           `json:"price,omitempty"`
	Score       float64           `json:"score"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Vector      []float64         `json:"vector,omitempty"`
}

// SearchResponse represents a search response
type SearchResponse struct {
	Results     []*SearchResult `json:"results"`
	Total       int             `json:"total"`
	QueryTime   string          `json:"query_time"`
	SearchType  string          `json:"search_type"`
	Suggestions []string        `json:"suggestions,omitempty"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status      string    `json:"status"`
	Service     string    `json:"service"`
	Version     string    `json:"version"`
	RedisStatus string    `json:"redis_status"`
	Timestamp   time.Time `json:"timestamp"`
	Uptime      string    `json:"uptime"`
}

// StatsResponse represents statistics response
type StatsResponse struct {
	TotalDocuments   int64             `json:"total_documents"`
	TotalSearches    int64             `json:"total_searches"`
	AverageQueryTime float64           `json:"average_query_time_ms"`
	PopularQueries   []string          `json:"popular_queries"`
	CategoryStats    map[string]int64  `json:"category_stats"`
	LastUpdated      time.Time         `json:"last_updated"`
}

var startTime = time.Now()

// NewRedis8AISearchEngine creates a new Redis 8 AI Search Engine
func NewRedis8AISearchEngine(client *redis.Client) *Redis8AISearchEngine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	engine := &Redis8AISearchEngine{
		client: client,
		router: router,
	}

	engine.setupRoutes()
	engine.initializeSearchData()

	return engine
}

// Start starts the AI search engine server
func (e *Redis8AISearchEngine) Start(port string) error {
	return e.router.Run(":" + port)
}

// setupRoutes sets up the API routes
func (e *Redis8AISearchEngine) setupRoutes() {
	api := e.router.Group("/api/v1/ai-search")
	{
		api.POST("/semantic", e.semanticSearch)
		api.POST("/vector", e.vectorSearch)
		api.POST("/hybrid", e.hybridSearch)
		api.GET("/suggestions/:query", e.getSuggestions)
		api.GET("/trending", e.getTrending)
		api.GET("/personalized/:user_id", e.getPersonalized)
		api.GET("/health", e.healthCheck)
		api.GET("/stats", e.getStats)
	}

	// Add CORS middleware
	e.router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
}

// semanticSearch performs semantic search using AI embeddings
func (e *Redis8AISearchEngine) semanticSearch(c *gin.Context) {
	start := time.Now()
	
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.MinScore == 0 {
		req.MinScore = 0.7
	}

	// Generate query embedding (mock implementation)
	queryVector := e.generateQueryEmbedding(req.Query)
	
	// Search using vector similarity
	results, err := e.performVectorSearch(queryVector, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Track search query
	e.trackSearchQuery(req.Query, "semantic")

	response := &SearchResponse{
		Results:     results,
		Total:       len(results),
		QueryTime:   fmt.Sprintf("%.2fms", float64(time.Since(start).Nanoseconds())/1e6),
		SearchType:  "semantic",
		Suggestions: e.generateSuggestions(req.Query),
	}

	c.JSON(http.StatusOK, response)
}

// vectorSearch performs pure vector similarity search
func (e *Redis8AISearchEngine) vectorSearch(c *gin.Context) {
	start := time.Now()
	
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	queryVector := e.generateQueryEmbedding(req.Query)
	results, err := e.performVectorSearch(queryVector, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	e.trackSearchQuery(req.Query, "vector")

	response := &SearchResponse{
		Results:    results,
		Total:      len(results),
		QueryTime:  fmt.Sprintf("%.2fms", float64(time.Since(start).Nanoseconds())/1e6),
		SearchType: "vector",
	}

	c.JSON(http.StatusOK, response)
}

// hybridSearch combines semantic and keyword search
func (e *Redis8AISearchEngine) hybridSearch(c *gin.Context) {
	start := time.Now()
	
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	// Perform both semantic and keyword search
	queryVector := e.generateQueryEmbedding(req.Query)
	semanticResults, _ := e.performVectorSearch(queryVector, req)
	keywordResults, _ := e.performKeywordSearch(req.Query, req)

	// Combine and rank results
	results := e.combineResults(semanticResults, keywordResults, req.Limit)

	e.trackSearchQuery(req.Query, "hybrid")

	response := &SearchResponse{
		Results:     results,
		Total:       len(results),
		QueryTime:   fmt.Sprintf("%.2fms", float64(time.Since(start).Nanoseconds())/1e6),
		SearchType:  "hybrid",
		Suggestions: e.generateSuggestions(req.Query),
	}

	c.JSON(http.StatusOK, response)
}

// getSuggestions returns search suggestions
func (e *Redis8AISearchEngine) getSuggestions(c *gin.Context) {
	query := c.Param("query")
	suggestions := e.generateSuggestions(query)
	
	c.JSON(http.StatusOK, gin.H{
		"query":       query,
		"suggestions": suggestions,
	})
}

// getTrending returns trending searches
func (e *Redis8AISearchEngine) getTrending(c *gin.Context) {
	trending := []string{
		"espresso",
		"cappuccino",
		"latte",
		"americano",
		"macchiato",
		"cold brew",
		"frappuccino",
		"mocha",
	}

	c.JSON(http.StatusOK, gin.H{
		"trending": trending,
		"updated":  time.Now(),
	})
}

// getPersonalized returns personalized recommendations
func (e *Redis8AISearchEngine) getPersonalized(c *gin.Context) {
	userID := c.Param("user_id")
	
	// Mock personalized recommendations
	recommendations := []*SearchResult{
		{
			ID:          "coffee_001",
			Title:       "Your Favorite Espresso",
			Description: "Based on your previous orders",
			Category:    "coffee",
			Price:       3.50,
			Score:       0.95,
		},
		{
			ID:          "coffee_002",
			Title:       "Recommended Latte",
			Description: "Similar to what you usually order",
			Category:    "coffee",
			Price:       4.25,
			Score:       0.88,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":         userID,
		"recommendations": recommendations,
		"generated_at":    time.Now(),
	})
}

// healthCheck returns service health status
func (e *Redis8AISearchEngine) healthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	redisStatus := "healthy"
	if _, err := e.client.Ping(ctx).Result(); err != nil {
		redisStatus = "unhealthy"
	}

	response := &HealthResponse{
		Status:      "healthy",
		Service:     "Redis 8 AI Search Engine",
		Version:     "1.0.0",
		RedisStatus: redisStatus,
		Timestamp:   time.Now(),
		Uptime:      time.Since(startTime).String(),
	}

	c.JSON(http.StatusOK, response)
}

// getStats returns search statistics
func (e *Redis8AISearchEngine) getStats(c *gin.Context) {
	ctx := context.Background()
	
	// Get stats from Redis
	totalDocs, _ := e.client.Get(ctx, "ai_search:stats:total_documents").Int64()
	totalSearches, _ := e.client.Get(ctx, "ai_search:stats:total_searches").Int64()
	
	response := &StatsResponse{
		TotalDocuments:   totalDocs,
		TotalSearches:    totalSearches,
		AverageQueryTime: 2.5, // Mock average
		PopularQueries:   []string{"coffee", "espresso", "latte", "cappuccino"},
		CategoryStats: map[string]int64{
			"coffee":    150,
			"tea":       75,
			"pastries":  50,
			"snacks":    25,
		},
		LastUpdated: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// Helper methods

// generateQueryEmbedding generates a mock embedding for the query
func (e *Redis8AISearchEngine) generateQueryEmbedding(query string) []float64 {
	// Mock embedding generation - in production, use actual AI models
	words := strings.Fields(strings.ToLower(query))
	embedding := make([]float64, 384) // Standard embedding size
	
	for i := range embedding {
		embedding[i] = math.Sin(float64(i) * 0.1 * float64(len(words)))
	}
	
	return embedding
}

// performVectorSearch performs vector similarity search
func (e *Redis8AISearchEngine) performVectorSearch(queryVector []float64, req SearchRequest) ([]*SearchResult, error) {
	// Mock vector search results - queryVector would be used in production
	_ = queryVector // Suppress unused parameter warning
	results := []*SearchResult{
		{
			ID:          "coffee_001",
			Title:       "Premium Espresso",
			Description: "Rich and bold espresso shot",
			Category:    "coffee",
			Price:       3.50,
			Score:       0.92,
		},
		{
			ID:          "coffee_002",
			Title:       "Cappuccino Deluxe",
			Description: "Creamy cappuccino with perfect foam",
			Category:    "coffee",
			Price:       4.25,
			Score:       0.88,
		},
		{
			ID:          "coffee_003",
			Title:       "Americano Classic",
			Description: "Traditional americano coffee",
			Category:    "coffee",
			Price:       3.00,
			Score:       0.85,
		},
	}

	// Filter by minimum score
	filtered := make([]*SearchResult, 0)
	for _, result := range results {
		if result.Score >= req.MinScore {
			filtered = append(filtered, result)
		}
	}

	// Apply limit
	if len(filtered) > req.Limit {
		filtered = filtered[:req.Limit]
	}

	return filtered, nil
}

// performKeywordSearch performs traditional keyword search
func (e *Redis8AISearchEngine) performKeywordSearch(query string, req SearchRequest) ([]*SearchResult, error) {
	// Mock keyword search - in production, use Redis FT.SEARCH
	_ = query // Suppress unused parameter warning
	_ = req   // Suppress unused parameter warning
	results := []*SearchResult{
		{
			ID:          "coffee_004",
			Title:       "Latte Supreme",
			Description: "Smooth latte with steamed milk",
			Category:    "coffee",
			Price:       4.50,
			Score:       0.80,
		},
	}

	return results, nil
}

// combineResults combines semantic and keyword search results
func (e *Redis8AISearchEngine) combineResults(semantic, keyword []*SearchResult, limit int) []*SearchResult {
	// Simple combination - in production, use sophisticated ranking
	combined := append(semantic, keyword...)
	
	if len(combined) > limit {
		combined = combined[:limit]
	}
	
	return combined
}

// generateSuggestions generates search suggestions
func (e *Redis8AISearchEngine) generateSuggestions(query string) []string {
	suggestions := []string{
		query + " with milk",
		query + " decaf",
		query + " large",
		"iced " + query,
		query + " beans",
	}
	
	return suggestions
}

// trackSearchQuery tracks search queries for analytics
func (e *Redis8AISearchEngine) trackSearchQuery(query, searchType string) {
	ctx := context.Background()
	
	// Increment total searches
	e.client.Incr(ctx, "ai_search:stats:total_searches")
	
	// Track query frequency
	e.client.ZIncrBy(ctx, "ai_search:popular_queries", 1, query)
	
	// Track search type
	e.client.Incr(ctx, fmt.Sprintf("ai_search:stats:search_type:%s", searchType))
}

// initializeSearchData initializes sample search data
func (e *Redis8AISearchEngine) initializeSearchData() {
	ctx := context.Background()
	
	// Set initial document count
	e.client.Set(ctx, "ai_search:stats:total_documents", 100, 0)
	
	log.Println("✅ Redis 8 AI Search Engine initialized with sample data")
}
