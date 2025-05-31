package redismcp

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// DataManager manages Redis data operations
type DataManager struct {
	redis  *redis.Client
	logger *logger.Logger
	stats  *OperationStats
	mutex  sync.RWMutex
}

// OperationStats tracks operation statistics
type OperationStats struct {
	TotalQueries    int64                    `json:"total_queries"`
	SuccessQueries  int64                    `json:"success_queries"`
	FailedQueries   int64                    `json:"failed_queries"`
	QueryTypes      map[string]int64         `json:"query_types"`
	AvgLatency      time.Duration            `json:"avg_latency"`
	LastUpdated     time.Time                `json:"last_updated"`
	OperationCounts map[string]int64         `json:"operation_counts"`
}

// DataSchema represents the Redis data schema
type DataSchema struct {
	Structures map[string]DataStructure `json:"structures"`
	Version    string                   `json:"version"`
	UpdatedAt  time.Time                `json:"updated_at"`
}

// DataStructure represents a Redis data structure
type DataStructure struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Fields      map[string]FieldInfo   `json:"fields,omitempty"`
	Examples    []string               `json:"examples"`
}

// FieldInfo represents field information
type FieldInfo struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// NewDataManager creates a new data manager
func NewDataManager(redisClient *redis.Client, logger *logger.Logger) *DataManager {
	return &DataManager{
		redis:  redisClient,
		logger: logger,
		stats: &OperationStats{
			QueryTypes:      make(map[string]int64),
			OperationCounts: make(map[string]int64),
			LastUpdated:     time.Now(),
		},
	}
}

// Execute executes a parsed query against Redis
func (dm *DataManager) Execute(ctx context.Context, query *ParsedQuery) (interface{}, error) {
	startTime := time.Now()
	
	dm.mutex.Lock()
	dm.stats.TotalQueries++
	dm.stats.QueryTypes[string(query.Type)]++
	dm.stats.OperationCounts[query.Operation]++
	dm.mutex.Unlock()

	defer func() {
		latency := time.Since(startTime)
		dm.updateLatency(latency)
	}()

	dm.logger.Info("Executing Redis operation",
		zap.String("operation", query.Operation),
		zap.String("key", query.Key),
		zap.Any("cmd", query.RedisCmd),
	)

	var result interface{}
	var err error

	switch query.Operation {
	case "HGETALL":
		result, err = dm.executeHGetAll(ctx, query)
	case "HGET":
		result, err = dm.executeHGet(ctx, query)
	case "HSET":
		result, err = dm.executeHSet(ctx, query)
	case "SADD":
		result, err = dm.executeSAdd(ctx, query)
	case "SMEMBERS":
		result, err = dm.executeSMembers(ctx, query)
	case "ZREVRANGE":
		result, err = dm.executeZRevRange(ctx, query)
	case "ZRANGE":
		result, err = dm.executeZRange(ctx, query)
	case "ZADD":
		result, err = dm.executeZAdd(ctx, query)
	case "LPUSH":
		result, err = dm.executeLPush(ctx, query)
	case "LRANGE":
		result, err = dm.executeLRange(ctx, query)
	case "XADD":
		result, err = dm.executeXAdd(ctx, query)
	case "XREAD":
		result, err = dm.executeXRead(ctx, query)
	case "SCAN":
		result, err = dm.executeScan(ctx, query)
	case "GET":
		result, err = dm.executeGet(ctx, query)
	case "SET":
		result, err = dm.executeSet(ctx, query)
	case "DEL":
		result, err = dm.executeDel(ctx, query)
	default:
		err = fmt.Errorf("unsupported operation: %s", query.Operation)
	}

	if err != nil {
		dm.mutex.Lock()
		dm.stats.FailedQueries++
		dm.mutex.Unlock()
		dm.logger.Error("Redis operation failed",
			zap.String("operation", query.Operation),
			zap.Error(err),
		)
		return nil, err
	}

	dm.mutex.Lock()
	dm.stats.SuccessQueries++
	dm.stats.LastUpdated = time.Now()
	dm.mutex.Unlock()

	return result, nil
}

// Hash operations
func (dm *DataManager) executeHGetAll(ctx context.Context, query *ParsedQuery) (interface{}, error) {
	return dm.redis.HGetAll(ctx, query.Key).Result()
}

func (dm *DataManager) executeHGet(ctx context.Context, query *ParsedQuery) (interface{}, error) {
	if len(query.RedisCmd) < 3 {
		return nil, fmt.Errorf("HGET requires key and field")
	}
	field := query.RedisCmd[2].(string)
	return dm.redis.HGet(ctx, query.Key, field).Result()
}

func (dm *DataManager) executeHSet(ctx context.Context, query *ParsedQuery) (interface{}, error) {
	if query.Fields == nil {
		return nil, fmt.Errorf("HSET requires fields")
	}
	
	args := make([]interface{}, 0, len(query.Fields)*2)
	for field, value := range query.Fields {
		args = append(args, field, value)
	}
	
	return dm.redis.HSet(ctx, query.Key, args...).Result()
}

// Set operations
func (dm *DataManager) executeSAdd(ctx context.Context, query *ParsedQuery) (interface{}, error) {
	return dm.redis.SAdd(ctx, query.Key, query.Value).Result()
}

func (dm *DataManager) executeSMembers(ctx context.Context, query *ParsedQuery) (interface{}, error) {
	return dm.redis.SMembers(ctx, query.Key).Result()
}

// Sorted Set operations
func (dm *DataManager) executeZRevRange(ctx context.Context, query *ParsedQuery) (interface{}, error) {
	start := int64(0)
	stop := int64(query.Limit - 1)
	if query.Limit == 0 {
		stop = -1
	}
	
	// Check if WITHSCORES is requested
	withScores := false
	for _, arg := range query.RedisCmd {
		if str, ok := arg.(string); ok && strings.ToUpper(str) == "WITHSCORES" {
			withScores = true
			break
		}
	}
	
	if withScores {
		return dm.redis.ZRevRangeWithScores(ctx, query.Key, start, stop).Result()
	}
	return dm.redis.ZRevRange(ctx, query.Key, start, stop).Result()
}

func (dm *DataManager) executeZRange(ctx context.Context, query *ParsedQuery) (interface{}, error) {
	start := int64(0)
	stop := int64(-1)
	
	// Check if WITHSCORES is requested
	withScores := false
	for _, arg := range query.RedisCmd {
		if str, ok := arg.(string); ok && strings.ToUpper(str) == "WITHSCORES" {
			withScores = true
			break
		}
	}
	
	if withScores {
		return dm.redis.ZRangeWithScores(ctx, query.Key, start, stop).Result()
	}
	return dm.redis.ZRange(ctx, query.Key, start, stop).Result()
}

func (dm *DataManager) executeZAdd(ctx context.Context, query *ParsedQuery) (interface{}, error) {
	if len(query.RedisCmd) < 4 {
		return nil, fmt.Errorf("ZADD requires key, score, and member")
	}
	
	score, err := strconv.ParseFloat(query.RedisCmd[2].(string), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid score: %v", err)
	}
	
	member := query.RedisCmd[3]
	return dm.redis.ZAdd(ctx, query.Key, &redis.Z{Score: score, Member: member}).Result()
}

// List operations
func (dm *DataManager) executeLPush(ctx context.Context, query *ParsedQuery) (interface{}, error) {
	return dm.redis.LPush(ctx, query.Key, query.Value).Result()
}

func (dm *DataManager) executeLRange(ctx context.Context, query *ParsedQuery) (interface{}, error) {
	start := int64(0)
	stop := int64(-1)
	
	if query.Offset > 0 {
		start = int64(query.Offset)
	}
	if query.Limit > 0 {
		stop = start + int64(query.Limit) - 1
	}
	
	return dm.redis.LRange(ctx, query.Key, start, stop).Result()
}

// Stream operations
func (dm *DataManager) executeXAdd(ctx context.Context, query *ParsedQuery) (interface{}, error) {
	if query.Fields == nil {
		return nil, fmt.Errorf("XADD requires fields")
	}
	
	args := &redis.XAddArgs{
		Stream: query.Key,
		ID:     "*",
		Values: query.Fields,
	}
	
	return dm.redis.XAdd(ctx, args).Result()
}

func (dm *DataManager) executeXRead(ctx context.Context, query *ParsedQuery) (interface{}, error) {
	args := &redis.XReadArgs{
		Streams: []string{query.Key, "$"},
		Count:   int64(query.Limit),
		Block:   time.Second,
	}
	
	return dm.redis.XRead(ctx, args).Result()
}

// Search operations
func (dm *DataManager) executeScan(ctx context.Context, query *ParsedQuery) (interface{}, error) {
	var cursor uint64
	var keys []string
	var err error
	
	pattern := query.Key
	if pattern == "" {
		pattern = "*"
	}
	
	for {
		var scanKeys []string
		scanKeys, cursor, err = dm.redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}
		
		keys = append(keys, scanKeys...)
		
		if cursor == 0 {
			break
		}
	}
	
	return keys, nil
}

// String operations
func (dm *DataManager) executeGet(ctx context.Context, query *ParsedQuery) (interface{}, error) {
	return dm.redis.Get(ctx, query.Key).Result()
}

func (dm *DataManager) executeSet(ctx context.Context, query *ParsedQuery) (interface{}, error) {
	return dm.redis.Set(ctx, query.Key, query.Value, 0).Result()
}

func (dm *DataManager) executeDel(ctx context.Context, query *ParsedQuery) (interface{}, error) {
	return dm.redis.Del(ctx, query.Key).Result()
}

// GetStats returns operation statistics
func (dm *DataManager) GetStats() *OperationStats {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()
	
	// Create a copy to avoid race conditions
	stats := &OperationStats{
		TotalQueries:    dm.stats.TotalQueries,
		SuccessQueries:  dm.stats.SuccessQueries,
		FailedQueries:   dm.stats.FailedQueries,
		QueryTypes:      make(map[string]int64),
		AvgLatency:      dm.stats.AvgLatency,
		LastUpdated:     dm.stats.LastUpdated,
		OperationCounts: make(map[string]int64),
	}
	
	for k, v := range dm.stats.QueryTypes {
		stats.QueryTypes[k] = v
	}
	for k, v := range dm.stats.OperationCounts {
		stats.OperationCounts[k] = v
	}
	
	return stats
}

// GetSchema returns the data schema
func (dm *DataManager) GetSchema() *DataSchema {
	return &DataSchema{
		Structures: map[string]DataStructure{
			"coffee:menu:{shop_id}": {
				Type:        "hash",
				Description: "Coffee menu items for a specific shop",
				Fields: map[string]FieldInfo{
					"latte":     {Type: "string", Description: "Latte price", Required: false},
					"espresso":  {Type: "string", Description: "Espresso price", Required: false},
					"cappuccino": {Type: "string", Description: "Cappuccino price", Required: false},
				},
				Examples: []string{"HGETALL coffee:menu:downtown", "HGET coffee:menu:downtown latte"},
			},
			"coffee:inventory:{shop_id}": {
				Type:        "hash",
				Description: "Ingredient inventory for a specific shop",
				Fields: map[string]FieldInfo{
					"coffee_beans": {Type: "number", Description: "Coffee beans quantity in kg", Required: false},
					"milk":         {Type: "number", Description: "Milk quantity in liters", Required: false},
					"sugar":        {Type: "number", Description: "Sugar quantity in kg", Required: false},
				},
				Examples: []string{"HGETALL coffee:inventory:downtown", "HGET coffee:inventory:downtown milk"},
			},
			"coffee:orders:{date}": {
				Type:        "sorted_set",
				Description: "Daily coffee orders ranked by popularity",
				Examples:    []string{"ZREVRANGE coffee:orders:today 0 9 WITHSCORES"},
			},
			"customer:{id}": {
				Type:        "hash",
				Description: "Customer profile data",
				Fields: map[string]FieldInfo{
					"name":            {Type: "string", Description: "Customer name", Required: true},
					"email":           {Type: "string", Description: "Customer email", Required: true},
					"favorite_drink":  {Type: "string", Description: "Favorite coffee drink", Required: false},
					"loyalty_points":  {Type: "number", Description: "Loyalty points balance", Required: false},
				},
				Examples: []string{"HGETALL customer:123", "HGET customer:123 favorite_drink"},
			},
			"ingredients:available": {
				Type:        "set",
				Description: "Set of available ingredients",
				Examples:    []string{"SMEMBERS ingredients:available", "SADD ingredients:available oat_milk"},
			},
			"feedback:{shop_id}": {
				Type:        "stream",
				Description: "Customer feedback stream for a specific shop",
				Examples:    []string{"XREAD STREAMS feedback:downtown $", "XADD feedback:downtown * rating 5 comment great_coffee"},
			},
		},
		Version:   "1.0.0",
		UpdatedAt: time.Now(),
	}
}

// updateLatency updates the average latency
func (dm *DataManager) updateLatency(latency time.Duration) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	
	if dm.stats.TotalQueries == 1 {
		dm.stats.AvgLatency = latency
	} else {
		// Calculate running average
		total := dm.stats.AvgLatency * time.Duration(dm.stats.TotalQueries-1)
		dm.stats.AvgLatency = (total + latency) / time.Duration(dm.stats.TotalQueries)
	}
}
