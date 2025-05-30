package aisearch

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// Redis8AISearchEngine represents the blazingly fast AI search engine
type Redis8AISearchEngine struct {
	redis  *redis.Client
	router *gin.Engine
}

// SearchRequest represents an AI search request
type SearchRequest struct {
	Query       string                 `json:"query"`
	Type        string                 `json:"type,omitempty"`        // "semantic", "hybrid", "vector"
	Limit       int                    `json:"limit,omitempty"`       // Default 10
	Threshold   float64                `json:"threshold,omitempty"`   // Similarity threshold
	Context     map[string]interface{} `json:"context,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

// SearchResult represents a search result
type SearchResult struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Score       float64                `json:"score"`
	Metadata    map[string]interface{} `json:"metadata"`
	Embedding   []float64              `json:"embedding,omitempty"`
}

// SearchResponse represents the search response
type SearchResponse struct {
	Success     bool                   `json:"success"`
	Results     []SearchResult         `json:"results"`
	Total       int                    `json:"total"`
	QueryTime   time.Duration          `json:"query_time"`
	SearchType  string                 `json:"search_type"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
}

// CoffeeItem represents a coffee menu item with AI embeddings
type CoffeeItem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	Tags        []string  `json:"tags"`
	Ingredients []string  `json:"ingredients"`
	Embedding   []float64 `json:"embedding"`
	Popularity  float64   `json:"popularity"`
	Rating      float64   `json:"rating"`
}

// NewRedis8AISearchEngine creates a new Redis 8 AI search engine
func NewRedis8AISearchEngine(redisClient *redis.Client) *Redis8AISearchEngine {
	engine := &Redis8AISearchEngine{
		redis:  redisClient,
		router: gin.New(),
	}

	engine.setupRoutes()
	engine.initializeAIData()
	return engine
}

func (e *Redis8AISearchEngine) setupRoutes() {
	e.router.Use(gin.Recovery())
	e.router.Use(e.corsMiddleware())

	api := e.router.Group("/api/v1/ai-search")
	{
		api.POST("/semantic", e.handleSemanticSearch)
		api.POST("/vector", e.handleVectorSearch)
		api.POST("/hybrid", e.handleHybridSearch)
		api.GET("/suggestions/:query", e.handleSuggestions)
		api.GET("/trending", e.handleTrending)
		api.GET("/personalized/:user_id", e.handlePersonalized)
		api.GET("/health", e.handleHealth)
		api.GET("/stats", e.handleStats)
	}
}

// handleSemanticSearch handles semantic search using Redis 8 vector similarity
func (e *Redis8AISearchEngine) handleSemanticSearch(c *gin.Context) {
	startTime := time.Now()
	
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	log.Printf("🔍 Semantic search query: %s", req.Query)

	// Generate query embedding (mock implementation)
	queryEmbedding := e.generateQueryEmbedding(req.Query)

	// Perform blazingly fast vector search using Redis 8
	results, err := e.performVectorSearch(c.Request.Context(), queryEmbedding, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	// Generate AI-powered suggestions
	suggestions := e.generateSuggestions(req.Query)

	response := SearchResponse{
		Success:     true,
		Results:     results,
		Total:       len(results),
		QueryTime:   time.Since(startTime),
		SearchType:  "semantic",
		Suggestions: suggestions,
		Metadata: map[string]interface{}{
			"query_embedding_dim": len(queryEmbedding),
			"search_algorithm":    "redis8_vector_similarity",
			"performance":         "blazingly_fast",
		},
		Timestamp: time.Now(),
	}

	log.Printf("✅ Semantic search completed in %v", response.QueryTime)
	c.JSON(http.StatusOK, response)
}

// handleVectorSearch handles pure vector similarity search
func (e *Redis8AISearchEngine) handleVectorSearch(c *gin.Context) {
	startTime := time.Now()
	
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	log.Printf("🎯 Vector search query: %s", req.Query)

	// Generate high-dimensional embedding
	queryEmbedding := e.generateAdvancedEmbedding(req.Query)

	// Ultra-fast Redis 8 vector search
	results, err := e.performAdvancedVectorSearch(c.Request.Context(), queryEmbedding, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Vector search failed"})
		return
	}

	response := SearchResponse{
		Success:    true,
		Results:    results,
		Total:      len(results),
		QueryTime:  time.Since(startTime),
		SearchType: "vector",
		Metadata: map[string]interface{}{
			"embedding_model":     "coffee_ai_v2",
			"vector_dimensions":   384,
			"similarity_function": "cosine",
			"redis8_optimized":    true,
		},
		Timestamp: time.Now(),
	}

	log.Printf("⚡ Vector search completed in %v", response.QueryTime)
	c.JSON(http.StatusOK, response)
}

// handleHybridSearch combines semantic and traditional search
func (e *Redis8AISearchEngine) handleHybridSearch(c *gin.Context) {
	startTime := time.Now()
	
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	log.Printf("🔥 Hybrid search query: %s", req.Query)

	// Parallel execution of multiple search strategies
	semanticResults := make(chan []SearchResult, 1)
	keywordResults := make(chan []SearchResult, 1)

	// Semantic search
	go func() {
		embedding := e.generateQueryEmbedding(req.Query)
		results, _ := e.performVectorSearch(c.Request.Context(), embedding, req)
		semanticResults <- results
	}()

	// Keyword search
	go func() {
		results := e.performKeywordSearch(c.Request.Context(), req.Query, req)
		keywordResults <- results
	}()

	// Combine results with intelligent ranking
	semantic := <-semanticResults
	keyword := <-keywordResults
	
	combinedResults := e.combineAndRankResults(semantic, keyword, req.Query)

	response := SearchResponse{
		Success:    true,
		Results:    combinedResults,
		Total:      len(combinedResults),
		QueryTime:  time.Since(startTime),
		SearchType: "hybrid",
		Metadata: map[string]interface{}{
			"semantic_results": len(semantic),
			"keyword_results":  len(keyword),
			"fusion_algorithm": "reciprocal_rank_fusion",
			"redis8_powered":   true,
		},
		Timestamp: time.Now(),
	}

	log.Printf("🚀 Hybrid search completed in %v", response.QueryTime)
	c.JSON(http.StatusOK, response)
}

func (e *Redis8AISearchEngine) corsMiddleware() gin.HandlerFunc {
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

func (e *Redis8AISearchEngine) Start(port string) error {
	log.Printf("🚀 Starting Redis 8 AI Search Engine on port %s", port)
	return e.router.Run(":" + port)
}
