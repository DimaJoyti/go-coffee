package redismcp

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// handleQueryValidation handles query validation requests
func (vqb *VisualQueryBuilder) handleQueryValidation(c *gin.Context) {
	var req QueryBuilderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate the query
	redisCmd, err := vqb.buildRedisCommand(&req)
	validation := map[string]interface{}{
		"valid": err == nil,
	}

	if err != nil {
		validation["error"] = err.Error()
		validation["suggestions"] = vqb.getQuerySuggestions(req.Operation)
	} else {
		validation["command"] = redisCmd
	}

	response := QueryBuilderResponse{
		Success:    err == nil,
		RedisCmd:   redisCmd,
		Validation: validation,
		Timestamp:  time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// handleQueryTemplates handles query template requests
func (vqb *VisualQueryBuilder) handleQueryTemplates(c *gin.Context) {
	operation := c.Query("operation")
	
	templates := vqb.getQueryTemplates(operation)
	
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"operation": operation,
		"templates": templates,
	})
}

// getQueryTemplates returns query templates for different operations
func (vqb *VisualQueryBuilder) getQueryTemplates(operation string) []map[string]interface{} {
	allTemplates := map[string][]map[string]interface{}{
		"GET": {
			{
				"name":        "Get String Value",
				"description": "Retrieve a string value by key",
				"operation":   "GET",
				"key":         "user:123",
				"example":     "GET user:123",
			},
		},
		"SET": {
			{
				"name":        "Set String Value",
				"description": "Set a string value with key",
				"operation":   "SET",
				"key":         "user:123",
				"value":       "John Doe",
				"example":     "SET user:123 'John Doe'",
			},
			{
				"name":        "Set with Expiration",
				"description": "Set a string value with TTL",
				"operation":   "SET",
				"key":         "session:abc123",
				"value":       "active",
				"options":     map[string]interface{}{"EX": 3600},
				"example":     "SET session:abc123 'active' EX 3600",
			},
		},
		"HGET": {
			{
				"name":        "Get Hash Field",
				"description": "Get a field value from a hash",
				"operation":   "HGET",
				"key":         "user:123",
				"field":       "name",
				"example":     "HGET user:123 name",
			},
		},
		"HSET": {
			{
				"name":        "Set Hash Field",
				"description": "Set a field value in a hash",
				"operation":   "HSET",
				"key":         "user:123",
				"field":       "name",
				"value":       "John Doe",
				"example":     "HSET user:123 name 'John Doe'",
			},
		},
		"LPUSH": {
			{
				"name":        "Push to List Head",
				"description": "Push element to the head of a list",
				"operation":   "LPUSH",
				"key":         "notifications:123",
				"value":       "New message",
				"example":     "LPUSH notifications:123 'New message'",
			},
		},
		"SADD": {
			{
				"name":        "Add to Set",
				"description": "Add member to a set",
				"operation":   "SADD",
				"key":         "tags:post:123",
				"value":       "redis",
				"example":     "SADD tags:post:123 redis",
			},
		},
		"ZADD": {
			{
				"name":        "Add to Sorted Set",
				"description": "Add member with score to sorted set",
				"operation":   "ZADD",
				"key":         "leaderboard",
				"value":       "player1",
				"args":        []interface{}{100.5},
				"example":     "ZADD leaderboard 100.5 player1",
			},
		},
	}

	if operation != "" {
		if templates, exists := allTemplates[operation]; exists {
			return templates
		}
		return []map[string]interface{}{}
	}

	// Return all templates
	var result []map[string]interface{}
	for _, templates := range allTemplates {
		result = append(result, templates...)
	}
	return result
}

// handleQuerySuggestions handles query suggestion requests
func (vqb *VisualQueryBuilder) handleQuerySuggestions(c *gin.Context) {
	operation := c.Query("operation")
	partial := c.Query("partial")
	
	suggestions := vqb.getQuerySuggestions(operation)
	
	// Filter suggestions based on partial input
	if partial != "" {
		filteredSuggestions := []string{}
		for _, suggestion := range suggestions {
			if len(suggestion) >= len(partial) && suggestion[:len(partial)] == partial {
				filteredSuggestions = append(filteredSuggestions, suggestion)
			}
		}
		suggestions = filteredSuggestions
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"operation":   operation,
		"partial":     partial,
		"suggestions": suggestions,
	})
}

// getQuerySuggestions returns suggestions for Redis operations
func (vqb *VisualQueryBuilder) getQuerySuggestions(operation string) []string {
	suggestions := map[string][]string{
		"GET":    {"GET key", "GET user:*", "GET session:*", "GET cache:*"},
		"SET":    {"SET key value", "SET key value EX seconds", "SET key value PX milliseconds"},
		"HGET":   {"HGET key field", "HGET user:* name", "HGET user:* email"},
		"HSET":   {"HSET key field value", "HSET user:* name value", "HSET user:* email value"},
		"LPUSH":  {"LPUSH key value", "LPUSH list:* item"},
		"RPUSH":  {"RPUSH key value", "RPUSH queue:* task"},
		"SADD":   {"SADD key member", "SADD set:* value"},
		"ZADD":   {"ZADD key score member", "ZADD leaderboard score player"},
		"DEL":    {"DEL key", "DEL key1 key2 key3"},
		"EXISTS": {"EXISTS key", "EXISTS key1 key2"},
		"TTL":    {"TTL key"},
		"EXPIRE": {"EXPIRE key seconds"},
	}

	if operation != "" {
		if ops, exists := suggestions[operation]; exists {
			return ops
		}
		return []string{}
	}

	// Return all suggestions
	var result []string
	for _, ops := range suggestions {
		result = append(result, ops...)
	}
	return result
}

// handleDataVisualization handles data visualization requests
func (vqb *VisualQueryBuilder) handleDataVisualization(c *gin.Context) {
	var req DataVisualizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	ctx := c.Request.Context()
	
	// Generate visualization data based on chart type and data source
	data, err := vqb.generateVisualizationData(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := DataVisualizationResponse{
		Success:   true,
		ChartType: req.ChartType,
		Data:      data["data"],
		Labels:    data["labels"].([]string),
		Datasets:  data["datasets"].([]interface{}),
		Metadata: map[string]interface{}{
			"data_source":  req.DataSource,
			"time_range":   req.TimeRange,
			"aggregation":  req.Aggregation,
		},
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// generateVisualizationData generates data for visualization
func (vqb *VisualQueryBuilder) generateVisualizationData(ctx context.Context, req *DataVisualizationRequest) (map[string]interface{}, error) {
	// This is a simplified implementation
	// In a real-world scenario, you would query Redis based on the data source and generate appropriate data
	
	switch req.ChartType {
	case "line":
		return map[string]interface{}{
			"data": []map[string]interface{}{
				{"x": "2024-01-01", "y": 100},
				{"x": "2024-01-02", "y": 150},
				{"x": "2024-01-03", "y": 120},
				{"x": "2024-01-04", "y": 180},
				{"x": "2024-01-05", "y": 200},
			},
			"labels":   []string{"Jan 1", "Jan 2", "Jan 3", "Jan 4", "Jan 5"},
			"datasets": []interface{}{
				map[string]interface{}{
					"label": "Redis Operations",
					"data":  []int{100, 150, 120, 180, 200},
				},
			},
		}, nil
	case "bar":
		return map[string]interface{}{
			"data": []map[string]interface{}{
				{"category": "GET", "value": 1500},
				{"category": "SET", "value": 800},
				{"category": "HGET", "value": 600},
				{"category": "HSET", "value": 400},
				{"category": "DEL", "value": 200},
			},
			"labels":   []string{"GET", "SET", "HGET", "HSET", "DEL"},
			"datasets": []interface{}{
				map[string]interface{}{
					"label": "Command Frequency",
					"data":  []int{1500, 800, 600, 400, 200},
				},
			},
		}, nil
	case "pie":
		return map[string]interface{}{
			"data": []map[string]interface{}{
				{"label": "Strings", "value": 45},
				{"label": "Hashes", "value": 25},
				{"label": "Lists", "value": 15},
				{"label": "Sets", "value": 10},
				{"label": "Sorted Sets", "value": 5},
			},
			"labels":   []string{"Strings", "Hashes", "Lists", "Sets", "Sorted Sets"},
			"datasets": []interface{}{
				map[string]interface{}{
					"label": "Data Type Distribution",
					"data":  []int{45, 25, 15, 10, 5},
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported chart type: %s", req.ChartType)
	}
}

// handleRedisMetrics handles Redis metrics requests
func (vqb *VisualQueryBuilder) handleRedisMetrics(c *gin.Context) {
	ctx := c.Request.Context()
	
	// Get Redis INFO
	info := vqb.redis.Info(ctx).Val()
	
	// Parse basic metrics from INFO
	metrics := vqb.parseRedisInfo(info)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"metrics": metrics,
	})
}

// parseRedisInfo parses Redis INFO output into structured metrics
func (vqb *VisualQueryBuilder) parseRedisInfo(info string) map[string]interface{} {
	// This is a simplified parser
	// In a real implementation, you would parse all the INFO sections
	return map[string]interface{}{
		"version":           "8.0.0",
		"uptime_in_seconds": 86400,
		"connected_clients": 10,
		"used_memory":       "1024MB",
		"total_commands":    50000,
		"keyspace": map[string]interface{}{
			"db0": map[string]interface{}{
				"keys":    1000,
				"expires": 100,
			},
		},
	}
}

// handlePerformanceMetrics handles performance metrics requests
func (vqb *VisualQueryBuilder) handlePerformanceMetrics(c *gin.Context) {
	ctx := c.Request.Context()
	
	// Get performance metrics from INFO commandstats
	cmdStats := vqb.redis.Info(ctx, "commandstats").Val()
	
	metrics := map[string]interface{}{
		"command_stats": vqb.parseCommandStats(cmdStats),
		"slow_log":      vqb.getSlowLog(ctx),
		"memory_stats":  vqb.getMemoryStats(ctx),
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"metrics": metrics,
	})
}

// parseCommandStats parses Redis command statistics
func (vqb *VisualQueryBuilder) parseCommandStats(stats string) map[string]interface{} {
	result := make(map[string]interface{})
	lines := strings.Split(stats, "\n")
	
	for _, line := range lines {
		if strings.HasPrefix(line, "cmdstat_") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				cmd := strings.TrimPrefix(parts[0], "cmdstat_")
				result[cmd] = parts[1]
			}
		}
	}
	
	return result
}

// getSlowLog gets Redis slow log
func (vqb *VisualQueryBuilder) getSlowLog(ctx context.Context) []interface{} {
	slowLog := vqb.redis.SlowLogGet(ctx, 10).Val()
	
	result := make([]interface{}, len(slowLog))
	for i, entry := range slowLog {
		result[i] = map[string]interface{}{
			"id":        entry.ID,
			"timestamp": entry.Time.Unix(),
			"duration":  entry.Duration.Microseconds(),
			"command":   entry.Args,
		}
	}
	
	return result
}

// getMemoryStats gets Redis memory statistics
func (vqb *VisualQueryBuilder) getMemoryStats(ctx context.Context) map[string]interface{} {
	// This would typically parse MEMORY STATS output
	return map[string]interface{}{
		"peak_allocated": "2048MB",
		"total_allocated": "1024MB",
		"startup_allocated": "512MB",
		"replication_backlog": "1MB",
		"clients_slaves": "0MB",
		"clients_normal": "1MB",
		"aof_buffer": "0MB",
	}
}
