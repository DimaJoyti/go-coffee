package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

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
	Type       string        `json:"type"`
	Operation  string        `json:"operation"`
	Key        string        `json:"key"`
	Value      interface{}   `json:"value,omitempty"`
	RedisCmd   []interface{} `json:"redis_cmd"`
	Confidence float64       `json:"confidence"`
}

// SimpleMCPServer represents a simple Redis MCP server for demo
type SimpleMCPServer struct {
	redis  *redis.Client
	router *gin.Engine
}

func NewSimpleMCPServer(redisClient *redis.Client) *SimpleMCPServer {
	server := &SimpleMCPServer{
		redis:  redisClient,
		router: gin.New(),
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

	log.Printf("🔍 Processing MCP query: %s (Agent: %s)", req.Query, req.AgentID)

	// Parse the natural language query using simple pattern matching
	parsedQuery := s.parseQuery(req.Query)

	// Execute the Redis operation
	result, err := s.executeRedisCommand(c.Request.Context(), parsedQuery)
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

	log.Printf("✅ Query executed successfully: %s", parsedQuery.Operation)
	c.JSON(http.StatusOK, response)
}

func (s *SimpleMCPServer) parseQuery(query string) *ParsedQuery {
	query = strings.ToLower(query)

	// Pattern matching for different query types
	if strings.Contains(query, "get menu") {
		shopID := "downtown"
		if strings.Contains(query, "uptown") {
			shopID = "uptown"
		} else if strings.Contains(query, "westside") {
			shopID = "westside"
		}
		
		return &ParsedQuery{
			Type:       "read",
			Operation:  "HGETALL",
			Key:        fmt.Sprintf("coffee:menu:%s", shopID),
			RedisCmd:   []interface{}{"HGETALL", fmt.Sprintf("coffee:menu:%s", shopID)},
			Confidence: 0.9,
		}
	}

	if strings.Contains(query, "get inventory") {
		shopID := "downtown"
		if strings.Contains(query, "uptown") {
			shopID = "uptown"
		} else if strings.Contains(query, "westside") {
			shopID = "westside"
		}
		
		return &ParsedQuery{
			Type:       "read",
			Operation:  "HGETALL",
			Key:        fmt.Sprintf("coffee:inventory:%s", shopID),
			RedisCmd:   []interface{}{"HGETALL", fmt.Sprintf("coffee:inventory:%s", shopID)},
			Confidence: 0.9,
		}
	}

	if strings.Contains(query, "add") && strings.Contains(query, "ingredient") {
		ingredient := "new_ingredient"
		// Try to extract ingredient name
		words := strings.Fields(query)
		for i, word := range words {
			if word == "add" && i+1 < len(words) {
				ingredient = words[i+1]
				break
			}
		}
		
		return &ParsedQuery{
			Type:       "write",
			Operation:  "SADD",
			Key:        "ingredients:available",
			Value:      ingredient,
			RedisCmd:   []interface{}{"SADD", "ingredients:available", ingredient},
			Confidence: 0.85,
		}
	}

	if strings.Contains(query, "top") && strings.Contains(query, "orders") {
		return &ParsedQuery{
			Type:       "read",
			Operation:  "ZREVRANGE",
			Key:        "coffee:orders:today",
			RedisCmd:   []interface{}{"ZREVRANGE", "coffee:orders:today", "0", "9", "WITHSCORES"},
			Confidence: 0.9,
		}
	}

	if strings.Contains(query, "get customer") {
		customerID := "123"
		field := "name"
		
		// Try to extract customer ID and field
		words := strings.Fields(query)
		for i, word := range words {
			if word == "customer" && i+1 < len(words) {
				customerID = words[i+1]
				if i+2 < len(words) {
					field = words[i+2]
				}
				break
			}
		}
		
		return &ParsedQuery{
			Type:       "read",
			Operation:  "HGET",
			Key:        fmt.Sprintf("customer:%s", customerID),
			RedisCmd:   []interface{}{"HGET", fmt.Sprintf("customer:%s", customerID), field},
			Confidence: 0.8,
		}
	}

	if strings.Contains(query, "search") {
		searchTerm := "coffee"
		words := strings.Fields(query)
		for i, word := range words {
			if word == "search" && i+1 < len(words) {
				searchTerm = words[i+1]
				break
			}
		}
		
		return &ParsedQuery{
			Type:       "search",
			Operation:  "SCAN",
			Key:        fmt.Sprintf("*%s*", searchTerm),
			RedisCmd:   []interface{}{"SCAN", "0", "MATCH", fmt.Sprintf("*%s*", searchTerm), "COUNT", "10"},
			Confidence: 0.7,
		}
	}

	// Default fallback
	return &ParsedQuery{
		Type:       "read",
		Operation:  "KEYS",
		Key:        "*",
		RedisCmd:   []interface{}{"KEYS", "*"},
		Confidence: 0.5,
	}
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
	case "SCAN":
		pattern := query.Key
		var cursor uint64
		var keys []string
		var err error
		
		for {
			var scanKeys []string
			scanKeys, cursor, err = s.redis.Scan(ctx, cursor, pattern, 10).Result()
			if err != nil {
				return nil, err
			}
			
			keys = append(keys, scanKeys...)
			
			if cursor == 0 {
				break
			}
		}
		
		return keys, nil
	case "KEYS":
		return s.redis.Keys(ctx, query.Key).Result()
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
		"version":   "1.0.0-simple-demo",
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
		log.Printf("❌ MCP Error: %s - %v", message, err)
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
	log.Printf("🌐 Starting Simple Redis MCP Server on port %s", port)
	return s.router.Run(":" + port)
}

// StartSimpleMCPServer starts the simple Redis MCP demo server
func StartSimpleMCPServer() {
	log.Println("🚀 Starting Simple Redis MCP Demo Server...")

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

	log.Printf("✅ Connected to Redis successfully")

	// Initialize sample data
	initializeSampleData(redisClient)

	// Initialize Simple MCP server
	mcpServer := NewSimpleMCPServer(redisClient)

	// Start server in a goroutine
	go func() {
		if err := mcpServer.Start(serverPort); err != nil {
			log.Fatalf("Failed to start Redis MCP server: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	log.Printf("🎯 Simple Redis MCP Demo Server is running on http://localhost:%s", serverPort)
	log.Println("📋 Try these example queries:")
	log.Println("   - get menu for shop downtown")
	log.Println("   - get inventory for uptown")
	log.Println("   - add oat_milk to ingredients")
	log.Println("   - get top orders today")
	log.Println("   - get customer 123 name")
	log.Println("   - search coffee")
	<-c

	log.Println("🛑 Shutting down Simple Redis MCP Demo Server...")

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	}

	log.Println("✅ Simple Redis MCP Demo Server stopped gracefully")
}

func initializeSampleData(client *redis.Client) {
	ctx := context.Background()

	log.Println("🏪 Initializing sample coffee shop data...")

	// Sample coffee shop menus
	shops := map[string]map[string]string{
		"downtown": {
			"latte":      "4.50",
			"cappuccino": "4.00",
			"espresso":   "2.50",
			"americano":  "3.00",
			"macchiato":  "4.25",
		},
		"uptown": {
			"latte":      "4.75",
			"cappuccino": "4.25",
			"espresso":   "2.75",
			"americano":  "3.25",
			"mocha":      "5.00",
		},
		"westside": {
			"latte":      "4.25",
			"cappuccino": "3.75",
			"espresso":   "2.25",
			"americano":  "2.75",
			"flat_white": "4.50",
		},
	}

	// Set coffee shop menus
	for shopID, menu := range shops {
		menuKey := "coffee:menu:" + shopID
		for item, price := range menu {
			client.HSet(ctx, menuKey, item, price)
		}
		log.Printf("✅ Menu set for shop %s", shopID)
	}

	// Sample inventory data
	inventories := map[string]map[string]string{
		"downtown": {
			"coffee_beans": "150",
			"milk":         "75",
			"sugar":        "50",
			"oat_milk":     "25",
			"almond_milk":  "20",
		},
		"uptown": {
			"coffee_beans": "200",
			"milk":         "100",
			"sugar":        "60",
			"oat_milk":     "30",
			"coconut_milk": "15",
		},
		"westside": {
			"coffee_beans": "120",
			"milk":         "60",
			"sugar":        "40",
			"oat_milk":     "20",
			"soy_milk":     "25",
		},
	}

	// Set inventory data
	for shopID, inventory := range inventories {
		inventoryKey := "coffee:inventory:" + shopID
		for ingredient, quantity := range inventory {
			client.HSet(ctx, inventoryKey, ingredient, quantity)
		}
		log.Printf("✅ Inventory set for shop %s", shopID)
	}

	// Sample available ingredients set
	ingredients := []string{
		"coffee_beans", "milk", "sugar", "oat_milk", "almond_milk",
		"coconut_milk", "soy_milk", "vanilla_syrup", "caramel_syrup",
		"chocolate_syrup", "whipped_cream", "cinnamon", "nutmeg",
	}

	for _, ingredient := range ingredients {
		client.SAdd(ctx, "ingredients:available", ingredient)
	}
	log.Println("✅ Available ingredients set")

	// Sample daily orders (sorted set)
	orders := map[string]float64{
		"latte":      150,
		"cappuccino": 120,
		"americano":  100,
		"espresso":   80,
		"macchiato":  60,
		"mocha":      45,
		"flat_white": 30,
	}

	for drink, count := range orders {
		client.ZAdd(ctx, "coffee:orders:today", &redis.Z{
			Score:  count,
			Member: drink,
		})
	}
	log.Println("✅ Daily orders data set")

	// Sample customer data
	customers := map[string]map[string]string{
		"customer:123": {
			"name":           "John Doe",
			"email":          "john@example.com",
			"favorite_drink": "latte",
			"loyalty_points": "150",
			"visits":         "25",
		},
		"customer:456": {
			"name":           "Jane Smith",
			"email":          "jane@example.com",
			"favorite_drink": "cappuccino",
			"loyalty_points": "200",
			"visits":         "30",
		},
		"customer:789": {
			"name":           "Bob Johnson",
			"email":          "bob@example.com",
			"favorite_drink": "americano",
			"loyalty_points": "75",
			"visits":         "12",
		},
	}

	for customerKey, data := range customers {
		for field, value := range data {
			client.HSet(ctx, customerKey, field, value)
		}
		log.Printf("✅ Customer data set for %s", customerKey)
	}

	log.Println("🎉 Sample data initialization completed successfully!")
}
