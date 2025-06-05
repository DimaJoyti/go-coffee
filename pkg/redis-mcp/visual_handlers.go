package redismcp

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// handleKeyExploration handles key exploration requests
func (vqb *VisualQueryBuilder) handleKeyExploration(c *gin.Context) {
	pattern := c.DefaultQuery("pattern", "*")
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	req := DataExplorerRequest{
		DataType: "keys",
		Pattern:  pattern,
		Limit:    limit,
		Offset:   offset,
	}

	vqb.exploreKeys(c, c.Request.Context(), &req)
}

// handleKeyDetails handles detailed key information requests
func (vqb *VisualQueryBuilder) handleKeyDetails(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key parameter is required"})
		return
	}

	ctx := c.Request.Context()

	// Check if key exists
	exists := vqb.redis.Exists(ctx, key).Val()
	if exists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
		return
	}

	// Get key information
	keyType := vqb.redis.Type(ctx, key).Val()
	ttl := vqb.redis.TTL(ctx, key).Val()
	
	// Get memory usage if available
	var memoryUsage int64
	if result := vqb.redis.MemoryUsage(ctx, key); result.Err() == nil {
		memoryUsage = result.Val()
	}

	keyInfo := map[string]interface{}{
		"key":          key,
		"type":         keyType,
		"ttl":          ttl.Seconds(),
		"memory_usage": memoryUsage,
		"exists":       exists > 0,
	}

	// Get type-specific information
	switch keyType {
	case "string":
		length := vqb.redis.StrLen(ctx, key).Val()
		keyInfo["length"] = length
	case "list":
		length := vqb.redis.LLen(ctx, key).Val()
		keyInfo["length"] = length
	case "set":
		cardinality := vqb.redis.SCard(ctx, key).Val()
		keyInfo["cardinality"] = cardinality
	case "zset":
		cardinality := vqb.redis.ZCard(ctx, key).Val()
		keyInfo["cardinality"] = cardinality
	case "hash":
		length := vqb.redis.HLen(ctx, key).Val()
		keyInfo["field_count"] = length
	case "stream":
		length := vqb.redis.XLen(ctx, key).Val()
		keyInfo["length"] = length
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    keyInfo,
	})
}

// handleDataSearch handles data search requests
func (vqb *VisualQueryBuilder) handleDataSearch(c *gin.Context) {
	query := c.Query("q")
	dataType := c.DefaultQuery("type", "keys")
	limitStr := c.DefaultQuery("limit", "100")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	limit, _ := strconv.Atoi(limitStr)
	ctx := c.Request.Context()

	var results []interface{}

	switch dataType {
	case "keys":
		// Search for keys matching the query
		pattern := fmt.Sprintf("*%s*", query)
		keys := vqb.redis.Keys(ctx, pattern).Val()
		
		for i, key := range keys {
			if i >= limit {
				break
			}
			keyType := vqb.redis.Type(ctx, key).Val()
			results = append(results, map[string]interface{}{
				"key":  key,
				"type": keyType,
			})
		}
	case "values":
		// This would require a more complex implementation
		// For now, return a placeholder
		results = []interface{}{
			map[string]interface{}{
				"message": "Value search not implemented yet",
			},
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"query":   query,
		"type":    dataType,
		"results": results,
		"total":   len(results),
	})
}

// handleQueryBuilder handles visual query building requests
func (vqb *VisualQueryBuilder) handleQueryBuilder(c *gin.Context) {
	var req QueryBuilderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Build Redis command based on operation
	redisCmd, err := vqb.buildRedisCommand(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := QueryBuilderResponse{
		Success:    true,
		RedisCmd:   redisCmd,
		Validation: map[string]interface{}{"valid": true},
		Timestamp:  time.Now(),
	}

	// If not preview mode, execute the command
	if !req.Preview {
		result, err := vqb.executeRedisCommand(c.Request.Context(), &req)
		if err != nil {
			response.Success = false
			response.Validation = map[string]interface{}{
				"valid": false,
				"error": err.Error(),
			}
		} else {
			response.Result = result
		}
	} else {
		response.Preview = fmt.Sprintf("Command: %s", redisCmd)
	}

	c.JSON(http.StatusOK, response)
}

// buildRedisCommand builds a Redis command string from the request
func (vqb *VisualQueryBuilder) buildRedisCommand(req *QueryBuilderRequest) (string, error) {
	switch strings.ToUpper(req.Operation) {
	case "GET":
		if req.Key == "" {
			return "", fmt.Errorf("key is required for GET operation")
		}
		return fmt.Sprintf("GET %s", req.Key), nil
	case "SET":
		if req.Key == "" || req.Value == nil {
			return "", fmt.Errorf("key and value are required for SET operation")
		}
		return fmt.Sprintf("SET %s %v", req.Key, req.Value), nil
	case "HGET":
		if req.Key == "" || req.Field == "" {
			return "", fmt.Errorf("key and field are required for HGET operation")
		}
		return fmt.Sprintf("HGET %s %s", req.Key, req.Field), nil
	case "HSET":
		if req.Key == "" || req.Field == "" || req.Value == nil {
			return "", fmt.Errorf("key, field, and value are required for HSET operation")
		}
		return fmt.Sprintf("HSET %s %s %v", req.Key, req.Field, req.Value), nil
	case "LPUSH":
		if req.Key == "" || req.Value == nil {
			return "", fmt.Errorf("key and value are required for LPUSH operation")
		}
		return fmt.Sprintf("LPUSH %s %v", req.Key, req.Value), nil
	case "RPUSH":
		if req.Key == "" || req.Value == nil {
			return "", fmt.Errorf("key and value are required for RPUSH operation")
		}
		return fmt.Sprintf("RPUSH %s %v", req.Key, req.Value), nil
	case "SADD":
		if req.Key == "" || req.Value == nil {
			return "", fmt.Errorf("key and value are required for SADD operation")
		}
		return fmt.Sprintf("SADD %s %v", req.Key, req.Value), nil
	case "ZADD":
		if req.Key == "" || req.Value == nil || len(req.Args) == 0 {
			return "", fmt.Errorf("key, score, and member are required for ZADD operation")
		}
		return fmt.Sprintf("ZADD %s %v %v", req.Key, req.Args[0], req.Value), nil
	case "DEL":
		if req.Key == "" {
			return "", fmt.Errorf("key is required for DEL operation")
		}
		return fmt.Sprintf("DEL %s", req.Key), nil
	case "EXISTS":
		if req.Key == "" {
			return "", fmt.Errorf("key is required for EXISTS operation")
		}
		return fmt.Sprintf("EXISTS %s", req.Key), nil
	case "TTL":
		if req.Key == "" {
			return "", fmt.Errorf("key is required for TTL operation")
		}
		return fmt.Sprintf("TTL %s", req.Key), nil
	case "EXPIRE":
		if req.Key == "" || len(req.Args) == 0 {
			return "", fmt.Errorf("key and expiration time are required for EXPIRE operation")
		}
		return fmt.Sprintf("EXPIRE %s %v", req.Key, req.Args[0]), nil
	default:
		return "", fmt.Errorf("unsupported operation: %s", req.Operation)
	}
}

// executeRedisCommand executes a Redis command based on the request
func (vqb *VisualQueryBuilder) executeRedisCommand(ctx context.Context, req *QueryBuilderRequest) (interface{}, error) {
	switch strings.ToUpper(req.Operation) {
	case "GET":
		return vqb.redis.Get(ctx, req.Key).Result()
	case "SET":
		return vqb.redis.Set(ctx, req.Key, req.Value, 0).Result()
	case "HGET":
		return vqb.redis.HGet(ctx, req.Key, req.Field).Result()
	case "HSET":
		return vqb.redis.HSet(ctx, req.Key, req.Field, req.Value).Result()
	case "LPUSH":
		return vqb.redis.LPush(ctx, req.Key, req.Value).Result()
	case "RPUSH":
		return vqb.redis.RPush(ctx, req.Key, req.Value).Result()
	case "SADD":
		return vqb.redis.SAdd(ctx, req.Key, req.Value).Result()
	case "ZADD":
		if len(req.Args) == 0 {
			return nil, fmt.Errorf("score is required for ZADD")
		}
		score, ok := req.Args[0].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid score type")
		}
		return vqb.redis.ZAdd(ctx, req.Key, &redis.Z{Score: score, Member: req.Value}).Result()
	case "DEL":
		return vqb.redis.Del(ctx, req.Key).Result()
	case "EXISTS":
		return vqb.redis.Exists(ctx, req.Key).Result()
	case "TTL":
		result := vqb.redis.TTL(ctx, req.Key).Val()
		return result.Seconds(), nil
	case "EXPIRE":
		if len(req.Args) == 0 {
			return nil, fmt.Errorf("expiration time is required")
		}
		expiration, ok := req.Args[0].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid expiration time type")
		}
		return vqb.redis.Expire(ctx, req.Key, time.Duration(expiration)*time.Second).Result()
	default:
		return nil, fmt.Errorf("unsupported operation: %s", req.Operation)
	}
}
