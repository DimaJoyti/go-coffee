package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Simple AI service for demo
type SimpleAIService struct {
	logger *logger.Logger
}

func NewSimpleAIService(logger *logger.Logger) *SimpleAIService {
	return &SimpleAIService{logger: logger}
}

func (s *SimpleAIService) GenerateText(ctx context.Context, prompt string) (string, error) {
	// Simple pattern matching for Redis MCP demo
	if contains(prompt, "get menu") {
		return `{
			"type": "read",
			"operation": "HGETALL",
			"key": "coffee:menu:downtown",
			"redis_cmd": ["HGETALL", "coffee:menu:downtown"],
			"confidence": 0.9
		}`, nil
	}

	if contains(prompt, "get inventory") {
		return `{
			"type": "read",
			"operation": "HGETALL",
			"key": "coffee:inventory:downtown",
			"redis_cmd": ["HGETALL", "coffee:inventory:downtown"],
			"confidence": 0.9
		}`, nil
	}

	if contains(prompt, "add") && contains(prompt, "ingredients") {
		return `{
			"type": "write",
			"operation": "SADD",
			"key": "ingredients:available",
			"value": "new_ingredient",
			"redis_cmd": ["SADD", "ingredients:available", "new_ingredient"],
			"confidence": 0.85
		}`, nil
	}

	if contains(prompt, "top") && contains(prompt, "orders") {
		return `{
			"type": "read",
			"operation": "ZREVRANGE",
			"key": "coffee:orders:today",
			"redis_cmd": ["ZREVRANGE", "coffee:orders:today", "0", "9", "WITHSCORES"],
			"confidence": 0.9
		}`, nil
	}

	// Default response
	return `{
		"type": "read",
		"operation": "GET",
		"key": "default:key",
		"redis_cmd": ["GET", "default:key"],
		"confidence": 0.5
	}`, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr ||
			 findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// MCPRequest represents a natural language request to Redis
type MCPRequest struct {
	Query     string                 `json:"query"`
	Context   map[string]interface{} `json:"context,omitempty"`
	AgentID   string                 `json:"agent_id"`
	Timestamp time.Time              `json:"timestamp"`
}

// MCPResponse represents the response from Redis MCP
type MCPResponse struct {
	Success   bool                   `json:"success"`
	Data      interface{}            `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Query     string                 `json:"executed_query,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// ParsedQuery represents a parsed natural language query
type ParsedQuery struct {
	Type       string                 `json:"type"`
	Operation  string                 `json:"operation"`
	Key        string                 `json:"key"`
	Value      interface{}            `json:"value,omitempty"`
	RedisCmd   []interface{}          `json:"redis_cmd"`
	Confidence float64                `json:"confidence"`
}

// SimpleMCPServer represents a simple Redis MCP server for demo
type SimpleMCPServer struct {
	redis     *redis.Client
	logger    *logger.Logger
	aiService *SimpleAIService
	router    *gin.Engine
}

func NewSimpleMCPServer(redisClient *redis.Client, logger *logger.Logger) *SimpleMCPServer {
	server := &SimpleMCPServer{
		redis:     redisClient,
		logger:    logger,
		aiService: NewSimpleAIService(logger),
		router:    gin.New(),
	}

	server.setupRoutes()
	return server
}

func (s *SimpleMCPServer) setupRoutes() {
	s.router.Use(gin.Recovery())
	s.router.Use(s.corsMiddleware())

	api := s.router.Group("/api/v1/redis-mcp")
	{
		api.POST("/query", s.handleQuery)
		api.GET("/health", s.handleHealth)
		api.GET("/stats", s.handleStats)
	}
}

func (s *SimpleMCPServer) handleQuery(c *gin.Context) {
	var req MCPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.respondError(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	req.Timestamp = time.Now()

	s.logger.Info("Processing MCP query", map[string]interface{}{
		"query":    req.Query,
		"agent_id": req.AgentID,
	})

	// Parse the natural language query
	aiResponse, err := s.aiService.GenerateText(c.Request.Context(), req.Query)
	if err != nil {
		s.respondError(c, http.StatusInternalServerError, "AI parsing failed", err)
		return
	}

	var parsedQuery ParsedQuery
	if err := json.Unmarshal([]byte(aiResponse), &parsedQuery); err != nil {
		s.respondError(c, http.StatusBadRequest, "Failed to parse AI response", err)
		return
	}

	// Execute the Redis operation
	result, err := s.executeRedisCommand(c.Request.Context(), &parsedQuery)
	if err != nil {
		s.respondError(c, http.StatusInternalServerError, "Failed to execute Redis command", err)
		return
	}

	// Prepare response
	response := MCPResponse{
		Success: true,
		Data:    result,
		Query:   fmt.Sprintf("%v", parsedQuery.RedisCmd),
		Metadata: map[string]interface{}{
			"query_type": parsedQuery.Type,
			"operation":  parsedQuery.Operation,
			"confidence": parsedQuery.Confidence,
			"agent_id":   req.AgentID,
		},
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

func (s *SimpleMCPServer) executeRedisCommand(ctx context.Context, query *ParsedQuery) (interface{}, error) {
	switch query.Operation {
	case "HGETALL":
		return s.redis.HGetAll(ctx, query.Key).Result()
	case "HGET":
		if len(query.RedisCmd) >= 3 {
			field := query.RedisCmd[2].(string)
			return s.redis.HGet(ctx, query.Key, field).Result()
		}
		return nil, fmt.Errorf("HGET requires field parameter")
	case "SADD":
		return s.redis.SAdd(ctx, query.Key, query.Value).Result()
	case "SMEMBERS":
		return s.redis.SMembers(ctx, query.Key).Result()
	case "ZREVRANGE":
		start := int64(0)
		stop := int64(9)
		withScores := false
		
		// Check for WITHSCORES
		for _, arg := range query.RedisCmd {
			if str, ok := arg.(string); ok && str == "WITHSCORES" {
				withScores = true
				break
			}
		}
		
		if withScores {
			return s.redis.ZRevRangeWithScores(ctx, query.Key, start, stop).Result()
		}
		return s.redis.ZRevRange(ctx, query.Key, start, stop).Result()
	case "GET":
		return s.redis.Get(ctx, query.Key).Result()
	case "SET":
		return s.redis.Set(ctx, query.Key, query.Value, 0).Result()
	default:
		return nil, fmt.Errorf("unsupported operation: %s", query.Operation)
	}
}

func (s *SimpleMCPServer) handleHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.redis.Ping(ctx).Result()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "Redis connection failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0-demo",
	})
}

func (s *SimpleMCPServer) handleStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"total_queries": 0,
		"success_rate":  100.0,
		"uptime":        time.Since(time.Now().Add(-time.Hour)),
	})
}

func (s *SimpleMCPServer) respondError(c *gin.Context, status int, message string, err error) {
	response := MCPResponse{
		Success:   false,
		Error:     message,
		Timestamp: time.Now(),
	}

	if err != nil {
		response.Error = fmt.Sprintf("%s: %v", message, err)
		s.logger.Error("MCP Error", map[string]interface{}{
			"message": message,
			"error":   err.Error(),
		})
	}

	c.JSON(status, response)
}

func (s *SimpleMCPServer) corsMiddleware() gin.HandlerFunc {
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

func (s *SimpleMCPServer) Start(port string) error {
	s.logger.Info("Starting Simple Redis MCP Server", map[string]interface{}{
		"port": port,
	})
	return s.router.Run(":" + port)
}

func main() {
	log.Println("🚀 Starting Simple Redis MCP Demo Server...")

	// Initialize logger
	logger := logger.New("redis-mcp-demo")

	// Get Redis URL from environment
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8090"
	}

	// Initialize Redis client
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	redisClient := redis.NewClient(opt)

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	logger.Info("✅ Connected to Redis successfully", map[string]interface{}{
		"url": redisURL,
	})

	// Initialize sample data
	initializeSampleData(redisClient, logger)

	// Initialize Simple MCP server
	mcpServer := NewSimpleMCPServer(redisClient, logger)

	// Start server in a goroutine
	go func() {
		if err := mcpServer.Start(serverPort); err != nil {
			log.Fatalf("Failed to start Redis MCP server: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	logger.Info("🎯 Simple Redis MCP Demo Server is running", map[string]interface{}{
		"port": serverPort,
		"url":  fmt.Sprintf("http://localhost:%s", serverPort),
	})
	<-c

	logger.Info("🛑 Shutting down Simple Redis MCP Demo Server...")

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		logger.Error("Error closing Redis connection", map[string]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info("✅ Simple Redis MCP Demo Server stopped gracefully")
}

func initializeSampleData(client *redis.Client, logger *logger.Logger) {
	ctx := context.Background()

	logger.Info("🏪 Initializing sample coffee shop data...")

	// Sample coffee shop menu
	menu := map[string]string{
		"latte":      "4.50",
		"cappuccino": "4.00",
		"espresso":   "2.50",
		"americano":  "3.00",
		"macchiato":  "4.25",
	}

	for item, price := range menu {
		if err := client.HSet(ctx, "coffee:menu:downtown", item, price).Err(); err != nil {
			logger.Error("Failed to set menu item", map[string]interface{}{
				"error": err.Error(),
				"item":  item,
			})
		}
	}

	// Sample inventory
	inventory := map[string]string{
		"coffee_beans": "150",
		"milk":         "75",
		"sugar":        "50",
		"oat_milk":     "25",
	}

	for ingredient, quantity := range inventory {
		if err := client.HSet(ctx, "coffee:inventory:downtown", ingredient, quantity).Err(); err != nil {
			logger.Error("Failed to set inventory item", map[string]interface{}{
				"error":      err.Error(),
				"ingredient": ingredient,
			})
		}
	}

	// Sample ingredients set
	ingredients := []string{"coffee_beans", "milk", "sugar", "oat_milk", "vanilla_syrup"}
	for _, ingredient := range ingredients {
		if err := client.SAdd(ctx, "ingredients:available", ingredient).Err(); err != nil {
			logger.Error("Failed to add ingredient", map[string]interface{}{
				"error":      err.Error(),
				"ingredient": ingredient,
			})
		}
	}

	// Sample daily orders
	orders := map[string]float64{
		"latte":      150,
		"cappuccino": 120,
		"americano":  100,
		"espresso":   80,
	}

	for drink, count := range orders {
		if err := client.ZAdd(ctx, "coffee:orders:today", &redis.Z{
			Score:  count,
			Member: drink,
		}).Err(); err != nil {
			logger.Error("Failed to add order count", map[string]interface{}{
				"error": err.Error(),
				"drink": drink,
			})
		}
	}

	logger.Info("✅ Sample data initialized successfully!")
}
