package redismcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/internal/ai"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
)

// MCPServer represents the Redis MCP server
type MCPServer struct {
	redis       *redis.Client
	logger      *logger.Logger
	aiService   *ai.Service
	queryParser *QueryParser
	dataManager *DataManager
	router      *gin.Engine
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

// QueryType represents different types of Redis operations
type QueryType string

const (
	QueryTypeRead   QueryType = "read"
	QueryTypeWrite  QueryType = "write"
	QueryTypeDelete QueryType = "delete"
	QueryTypeStream QueryType = "stream"
	QueryTypeSearch QueryType = "search"
)

// ParsedQuery represents a parsed natural language query
type ParsedQuery struct {
	Type        QueryType              `json:"type"`
	Operation   string                 `json:"operation"`
	Key         string                 `json:"key"`
	Value       interface{}            `json:"value,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Offset      int                    `json:"offset,omitempty"`
	RedisCmd    []interface{}          `json:"redis_cmd"`
	Confidence  float64                `json:"confidence"`
}

// NewMCPServer creates a new Redis MCP server
func NewMCPServer(redisClient *redis.Client, aiService *ai.Service, logger *logger.Logger) *MCPServer {
	server := &MCPServer{
		redis:     redisClient,
		logger:    logger,
		aiService: aiService,
		router:    gin.New(),
	}

	// Initialize components
	server.queryParser = NewQueryParser(aiService, logger)
	server.dataManager = NewDataManager(redisClient, logger)

	// Setup routes
	server.setupRoutes()

	return server
}

// setupRoutes configures the HTTP routes for the MCP server
func (s *MCPServer) setupRoutes() {
	s.router.Use(gin.Recovery())
	s.router.Use(s.loggingMiddleware())
	s.router.Use(s.corsMiddleware())

	api := s.router.Group("/api/v1/redis-mcp")
	{
		api.POST("/query", s.handleQuery)
		api.GET("/health", s.handleHealth)
		api.GET("/stats", s.handleStats)
		api.POST("/batch", s.handleBatchQuery)
		api.GET("/schema", s.handleSchema)
	}
}

// handleQuery processes natural language queries to Redis
func (s *MCPServer) handleQuery(c *gin.Context) {
	var req MCPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.respondError(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Add timestamp
	req.Timestamp = time.Now()

	// Log the request
	s.logger.Info("Processing MCP query",
		zap.String("query", req.Query),
		zap.String("agent_id", req.AgentID),
	)

	// Parse the natural language query
	parsedQuery, err := s.queryParser.Parse(c.Request.Context(), req.Query, req.Context)
	if err != nil {
		s.respondError(c, http.StatusBadRequest, "Failed to parse query", err)
		return
	}

	// Validate confidence threshold
	if parsedQuery.Confidence < 0.7 {
		s.respondError(c, http.StatusBadRequest, 
			fmt.Sprintf("Query confidence too low: %.2f", parsedQuery.Confidence), nil)
		return
	}

	// Execute the Redis operation
	result, err := s.dataManager.Execute(c.Request.Context(), parsedQuery)
	if err != nil {
		s.respondError(c, http.StatusInternalServerError, "Failed to execute query", err)
		return
	}

	// Prepare response
	response := MCPResponse{
		Success:   true,
		Data:      result,
		Query:     fmt.Sprintf("%v", parsedQuery.RedisCmd),
		Metadata: map[string]interface{}{
			"query_type":   parsedQuery.Type,
			"operation":    parsedQuery.Operation,
			"confidence":   parsedQuery.Confidence,
			"agent_id":     req.AgentID,
		},
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// handleBatchQuery processes multiple queries in a single request
func (s *MCPServer) handleBatchQuery(c *gin.Context) {
	var requests []MCPRequest
	if err := c.ShouldBindJSON(&requests); err != nil {
		s.respondError(c, http.StatusBadRequest, "Invalid batch request format", err)
		return
	}

	responses := make([]MCPResponse, len(requests))
	
	for i, req := range requests {
		req.Timestamp = time.Now()
		
		// Parse query
		parsedQuery, err := s.queryParser.Parse(c.Request.Context(), req.Query, req.Context)
		if err != nil {
			responses[i] = MCPResponse{
				Success:   false,
				Error:     fmt.Sprintf("Parse error: %v", err),
				Timestamp: time.Now(),
			}
			continue
		}

		// Execute query
		result, err := s.dataManager.Execute(c.Request.Context(), parsedQuery)
		if err != nil {
			responses[i] = MCPResponse{
				Success:   false,
				Error:     fmt.Sprintf("Execution error: %v", err),
				Timestamp: time.Now(),
			}
			continue
		}

		responses[i] = MCPResponse{
			Success:   true,
			Data:      result,
			Query:     fmt.Sprintf("%v", parsedQuery.RedisCmd),
			Timestamp: time.Now(),
		}
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"responses": responses,
		"total":     len(responses),
		"success":   countSuccessful(responses),
		"failed":    len(responses) - countSuccessful(responses),
	})
}

// handleHealth returns the health status of the MCP server
func (s *MCPServer) handleHealth(c *gin.Context) {
	// Check Redis connection
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
		"version":   "1.0.0",
	})
}

// handleStats returns usage statistics
func (s *MCPServer) handleStats(c *gin.Context) {
	stats := s.dataManager.GetStats()
	c.JSON(http.StatusOK, stats)
}

// handleSchema returns the data schema information
func (s *MCPServer) handleSchema(c *gin.Context) {
	schema := s.dataManager.GetSchema()
	c.JSON(http.StatusOK, schema)
}

// respondError sends an error response
func (s *MCPServer) respondError(c *gin.Context, status int, message string, err error) {
	response := MCPResponse{
		Success:   false,
		Error:     message,
		Timestamp: time.Now(),
	}

	if err != nil {
		response.Error = fmt.Sprintf("%s: %v", message, err)
		s.logger.Error("MCP Error", zap.String("message", message), zap.Error(err))
	}

	c.JSON(status, response)
}

// loggingMiddleware logs HTTP requests
func (s *MCPServer) loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// corsMiddleware handles CORS
func (s *MCPServer) corsMiddleware() gin.HandlerFunc {
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

// countSuccessful counts successful responses in batch
func countSuccessful(responses []MCPResponse) int {
	count := 0
	for _, resp := range responses {
		if resp.Success {
			count++
		}
	}
	return count
}

// Start starts the MCP server
func (s *MCPServer) Start(port string) error {
	s.logger.Info("Starting Redis MCP Server", zap.String("port", port))
	return s.router.Run(":" + port)
}

// Stop gracefully stops the MCP server
func (s *MCPServer) Stop() {
	s.logger.Info("Stopping Redis MCP Server")
	// Cleanup resources
}
