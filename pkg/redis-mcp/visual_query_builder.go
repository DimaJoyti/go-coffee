package redismcp

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"

	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// VisualQueryBuilder handles visual query building and data exploration
type VisualQueryBuilder struct {
	redis    *redis.Client
	logger   *logger.Logger
	upgrader websocket.Upgrader
}

// DataExplorerRequest represents a request for data exploration
type DataExplorerRequest struct {
	DataType   string                 `json:"data_type"`   // "keys", "hash", "list", "set", "zset", "stream"
	Pattern    string                 `json:"pattern"`     // Key pattern for scanning
	Key        string                 `json:"key"`         // Specific key to explore
	Limit      int                    `json:"limit"`       // Limit results
	Offset     int                    `json:"offset"`      // Offset for pagination
	Filters    map[string]interface{} `json:"filters"`     // Additional filters
	SearchTerm string                 `json:"search_term"` // Search within data
}

// DataExplorerResponse represents the response from data exploration
type DataExplorerResponse struct {
	Success    bool                   `json:"success"`
	DataType   string                 `json:"data_type"`
	Data       interface{}            `json:"data"`
	Total      int                    `json:"total"`
	HasMore    bool                   `json:"has_more"`
	Metadata   map[string]interface{} `json:"metadata"`
	Timestamp  time.Time              `json:"timestamp"`
}

// QueryBuilderRequest represents a visual query builder request
type QueryBuilderRequest struct {
	Operation  string                 `json:"operation"`   // "GET", "SET", "HGET", etc.
	Key        string                 `json:"key"`         // Redis key
	Field      string                 `json:"field"`       // For hash operations
	Value      interface{}            `json:"value"`       // Value to set
	Args       []interface{}          `json:"args"`        // Additional arguments
	Options    map[string]interface{} `json:"options"`     // Query options
	Preview    bool                   `json:"preview"`     // Whether to preview only
}

// QueryBuilderResponse represents the response from query builder
type QueryBuilderResponse struct {
	Success     bool                   `json:"success"`
	RedisCmd    string                 `json:"redis_cmd"`
	Result      interface{}            `json:"result,omitempty"`
	Preview     string                 `json:"preview,omitempty"`
	Validation  map[string]interface{} `json:"validation"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// DataVisualizationRequest represents a request for data visualization
type DataVisualizationRequest struct {
	ChartType   string                 `json:"chart_type"`   // "line", "bar", "pie", "scatter"
	DataSource  string                 `json:"data_source"`  // Redis key or pattern
	TimeRange   string                 `json:"time_range"`   // "1h", "24h", "7d", "30d"
	Aggregation string                 `json:"aggregation"`  // "sum", "avg", "count", "max", "min"
	GroupBy     string                 `json:"group_by"`     // Field to group by
	Filters     map[string]interface{} `json:"filters"`      // Data filters
}

// DataVisualizationResponse represents the response for data visualization
type DataVisualizationResponse struct {
	Success   bool                   `json:"success"`
	ChartType string                 `json:"chart_type"`
	Data      interface{}            `json:"data"`
	Labels    []string               `json:"labels"`
	Datasets  []interface{}          `json:"datasets"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewVisualQueryBuilder creates a new visual query builder
func NewVisualQueryBuilder(redisClient *redis.Client, logger *logger.Logger) *VisualQueryBuilder {
	return &VisualQueryBuilder{
		redis:  redisClient,
		logger: logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
	}
}

// SetupVisualRoutes adds visual query routes to the router
func (vqb *VisualQueryBuilder) SetupVisualRoutes(router *gin.RouterGroup) {
	visual := router.Group("/visual")
	{
		// Data exploration endpoints
		visual.POST("/explore", vqb.handleDataExploration)
		visual.GET("/keys", vqb.handleKeyExploration)
		visual.GET("/key/:key", vqb.handleKeyDetails)
		visual.GET("/search", vqb.handleDataSearch)
		
		// Query builder endpoints
		visual.POST("/query/build", vqb.handleQueryBuilder)
		visual.POST("/query/validate", vqb.handleQueryValidation)
		visual.GET("/query/templates", vqb.handleQueryTemplates)
		visual.GET("/query/suggestions", vqb.handleQuerySuggestions)
		
		// Data visualization endpoints
		visual.POST("/visualize", vqb.handleDataVisualization)
		visual.GET("/metrics", vqb.handleRedisMetrics)
		visual.GET("/performance", vqb.handlePerformanceMetrics)
		
		// Real-time data streaming
		visual.GET("/stream", vqb.handleWebSocketConnection)
		visual.POST("/stream/subscribe", vqb.handleStreamSubscription)
	}
}

// handleDataExploration handles data exploration requests
func (vqb *VisualQueryBuilder) handleDataExploration(c *gin.Context) {
	var req DataExplorerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	ctx := c.Request.Context()
	
	switch req.DataType {
	case "keys":
		vqb.exploreKeys(c, ctx, &req)
	case "hash":
		vqb.exploreHash(c, ctx, &req)
	case "list":
		vqb.exploreList(c, ctx, &req)
	case "set":
		vqb.exploreSet(c, ctx, &req)
	case "zset":
		vqb.exploreZSet(c, ctx, &req)
	case "stream":
		vqb.exploreStream(c, ctx, &req)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported data type"})
	}
}

// exploreKeys explores Redis keys based on pattern
func (vqb *VisualQueryBuilder) exploreKeys(c *gin.Context, ctx context.Context, req *DataExplorerRequest) {
	pattern := req.Pattern
	if pattern == "" {
		pattern = "*"
	}

	// Use SCAN for better performance
	var keys []string
	var cursor uint64
	limit := req.Limit
	if limit == 0 {
		limit = 100
	}

	for len(keys) < limit {
		result := vqb.redis.Scan(ctx, cursor, pattern, int64(limit-len(keys)))
		if result.Err() != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Err().Error()})
			return
		}

		scanKeys, newCursor := result.Val()
		keys = append(keys, scanKeys...)
		cursor = newCursor

		if cursor == 0 {
			break
		}
	}

	// Get key types and TTL information
	keyDetails := make([]map[string]interface{}, 0, len(keys))
	for _, key := range keys {
		keyType := vqb.redis.Type(ctx, key).Val()
		ttl := vqb.redis.TTL(ctx, key).Val()
		
		keyDetails = append(keyDetails, map[string]interface{}{
			"key":  key,
			"type": keyType,
			"ttl":  ttl.Seconds(),
		})
	}

	response := DataExplorerResponse{
		Success:   true,
		DataType:  "keys",
		Data:      keyDetails,
		Total:     len(keyDetails),
		HasMore:   cursor != 0,
		Metadata: map[string]interface{}{
			"pattern": pattern,
			"cursor":  cursor,
		},
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// exploreHash explores Redis hash data
func (vqb *VisualQueryBuilder) exploreHash(c *gin.Context, ctx context.Context, req *DataExplorerRequest) {
	if req.Key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is required for hash exploration"})
		return
	}

	// Check if key exists and is a hash
	keyType := vqb.redis.Type(ctx, req.Key).Val()
	if keyType != "hash" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is not a hash"})
		return
	}

	// Get all hash fields and values
	hashData := vqb.redis.HGetAll(ctx, req.Key).Val()
	
	// Apply search filter if provided
	if req.SearchTerm != "" {
		filteredData := make(map[string]string)
		searchLower := strings.ToLower(req.SearchTerm)
		
		for field, value := range hashData {
			if strings.Contains(strings.ToLower(field), searchLower) ||
			   strings.Contains(strings.ToLower(value), searchLower) {
				filteredData[field] = value
			}
		}
		hashData = filteredData
	}

	response := DataExplorerResponse{
		Success:   true,
		DataType:  "hash",
		Data:      hashData,
		Total:     len(hashData),
		HasMore:   false,
		Metadata: map[string]interface{}{
			"key":         req.Key,
			"field_count": len(hashData),
		},
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// exploreList explores Redis list data
func (vqb *VisualQueryBuilder) exploreList(c *gin.Context, ctx context.Context, req *DataExplorerRequest) {
	if req.Key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is required for list exploration"})
		return
	}

	keyType := vqb.redis.Type(ctx, req.Key).Val()
	if keyType != "list" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is not a list"})
		return
	}

	// Get list length
	length := vqb.redis.LLen(ctx, req.Key).Val()

	// Calculate range for pagination
	start := int64(req.Offset)
	end := start + int64(req.Limit) - 1
	if req.Limit == 0 {
		end = -1 // Get all elements
	}

	// Get list elements
	elements := vqb.redis.LRange(ctx, req.Key, start, end).Val()

	response := DataExplorerResponse{
		Success:   true,
		DataType:  "list",
		Data:      elements,
		Total:     int(length),
		HasMore:   end < length-1 && req.Limit > 0,
		Metadata: map[string]interface{}{
			"key":    req.Key,
			"length": length,
			"start":  start,
			"end":    end,
		},
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// exploreSet explores Redis set data
func (vqb *VisualQueryBuilder) exploreSet(c *gin.Context, ctx context.Context, req *DataExplorerRequest) {
	if req.Key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is required for set exploration"})
		return
	}

	keyType := vqb.redis.Type(ctx, req.Key).Val()
	if keyType != "set" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is not a set"})
		return
	}

	// Get all set members
	members := vqb.redis.SMembers(ctx, req.Key).Val()

	// Apply search filter if provided
	if req.SearchTerm != "" {
		filteredMembers := make([]string, 0)
		searchLower := strings.ToLower(req.SearchTerm)

		for _, member := range members {
			if strings.Contains(strings.ToLower(member), searchLower) {
				filteredMembers = append(filteredMembers, member)
			}
		}
		members = filteredMembers
	}

	response := DataExplorerResponse{
		Success:   true,
		DataType:  "set",
		Data:      members,
		Total:     len(members),
		HasMore:   false,
		Metadata: map[string]interface{}{
			"key":           req.Key,
			"member_count":  len(members),
		},
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// exploreZSet explores Redis sorted set data
func (vqb *VisualQueryBuilder) exploreZSet(c *gin.Context, ctx context.Context, req *DataExplorerRequest) {
	if req.Key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is required for sorted set exploration"})
		return
	}

	keyType := vqb.redis.Type(ctx, req.Key).Val()
	if keyType != "zset" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is not a sorted set"})
		return
	}

	// Get sorted set length
	length := vqb.redis.ZCard(ctx, req.Key).Val()

	// Calculate range for pagination
	start := int64(req.Offset)
	end := start + int64(req.Limit) - 1
	if req.Limit == 0 {
		end = -1 // Get all elements
	}

	// Get sorted set elements with scores
	elements := vqb.redis.ZRangeWithScores(ctx, req.Key, start, end).Val()

	// Convert to more readable format
	data := make([]map[string]interface{}, len(elements))
	for i, elem := range elements {
		data[i] = map[string]interface{}{
			"member": elem.Member,
			"score":  elem.Score,
		}
	}

	response := DataExplorerResponse{
		Success:   true,
		DataType:  "zset",
		Data:      data,
		Total:     int(length),
		HasMore:   end < length-1 && req.Limit > 0,
		Metadata: map[string]interface{}{
			"key":    req.Key,
			"length": length,
			"start":  start,
			"end":    end,
		},
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// exploreStream explores Redis stream data
func (vqb *VisualQueryBuilder) exploreStream(c *gin.Context, ctx context.Context, req *DataExplorerRequest) {
	if req.Key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is required for stream exploration"})
		return
	}

	keyType := vqb.redis.Type(ctx, req.Key).Val()
	if keyType != "stream" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is not a stream"})
		return
	}

	// Get stream length
	length := vqb.redis.XLen(ctx, req.Key).Val()

	// Get stream entries
	count := int64(req.Limit)
	if count == 0 {
		count = 100 // Default limit
	}

	entries := vqb.redis.XRevRange(ctx, req.Key, "+", "-").Val()
	if len(entries) > int(count) {
		entries = entries[:count]
	}

	// Convert to more readable format
	data := make([]map[string]interface{}, len(entries))
	for i, entry := range entries {
		data[i] = map[string]interface{}{
			"id":     entry.ID,
			"values": entry.Values,
		}
	}

	response := DataExplorerResponse{
		Success:   true,
		DataType:  "stream",
		Data:      data,
		Total:     int(length),
		HasMore:   int64(len(entries)) < length,
		Metadata: map[string]interface{}{
			"key":    req.Key,
			"length": length,
		},
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}
