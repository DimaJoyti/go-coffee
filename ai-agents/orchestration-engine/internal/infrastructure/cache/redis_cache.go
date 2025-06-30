package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-coffee-ai-agents/orchestration-engine/internal/config"
)

// CacheInterface defines the caching operations
type CacheInterface interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Increment(ctx context.Context, key string) (int64, error)
	Decrement(ctx context.Context, key string) (int64, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	GetSet(ctx context.Context, key string, value interface{}) ([]byte, error)
	MGet(ctx context.Context, keys ...string) ([]interface{}, error)
	MSet(ctx context.Context, pairs ...interface{}) error
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	FlushAll(ctx context.Context) error
	Close() error
}

// RedisCache implements CacheInterface using Redis
type RedisCache struct {
	client RedisClient
	config *config.CacheConfig
	logger Logger
}

// RedisClient interface for Redis operations
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) (int64, error)
	Exists(ctx context.Context, keys ...string) (int64, error)
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	GetSet(ctx context.Context, key string, value interface{}) (string, error)
	MGet(ctx context.Context, keys ...string) ([]interface{}, error)
	MSet(ctx context.Context, values ...interface{}) error
	Expire(ctx context.Context, key string, expiration time.Duration) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	FlushAll(ctx context.Context) error
	Close() error
	Ping(ctx context.Context) error
}

// Logger interface for logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(client RedisClient, config *config.CacheConfig, logger Logger) *RedisCache {
	return &RedisCache{
		client: client,
		config: config,
		logger: logger,
	}
}

// Get retrieves a value from cache
func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	start := time.Now()
	defer func() {
		r.logger.Debug("Cache GET operation", "key", key, "duration", time.Since(start))
	}()

	value, err := r.client.Get(ctx, key)
	if err != nil {
		r.logger.Error("Failed to get from cache", err, "key", key)
		return nil, err
	}

	if value == "" {
		return nil, ErrCacheMiss
	}

	return []byte(value), nil
}

// Set stores a value in cache
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	start := time.Now()
	defer func() {
		r.logger.Debug("Cache SET operation", "key", key, "duration", time.Since(start))
	}()

	// Serialize value to JSON if it's not already a string or []byte
	var serializedValue interface{}
	switch v := value.(type) {
	case string:
		serializedValue = v
	case []byte:
		serializedValue = string(v)
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			r.logger.Error("Failed to serialize value for cache", err, "key", key)
			return fmt.Errorf("failed to serialize value: %w", err)
		}
		serializedValue = string(jsonData)
	}

	err := r.client.Set(ctx, key, serializedValue, expiration)
	if err != nil {
		r.logger.Error("Failed to set cache value", err, "key", key)
		return err
	}

	return nil
}

// Delete removes a value from cache
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	start := time.Now()
	defer func() {
		r.logger.Debug("Cache DELETE operation", "key", key, "duration", time.Since(start))
	}()

	_, err := r.client.Del(ctx, key)
	if err != nil {
		r.logger.Error("Failed to delete from cache", err, "key", key)
		return err
	}

	return nil
}

// Exists checks if a key exists in cache
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key)
	if err != nil {
		r.logger.Error("Failed to check cache existence", err, "key", key)
		return false, err
	}

	return count > 0, nil
}

// Increment increments a numeric value in cache
func (r *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	value, err := r.client.Incr(ctx, key)
	if err != nil {
		r.logger.Error("Failed to increment cache value", err, "key", key)
		return 0, err
	}

	return value, nil
}

// Decrement decrements a numeric value in cache
func (r *RedisCache) Decrement(ctx context.Context, key string) (int64, error) {
	value, err := r.client.Decr(ctx, key)
	if err != nil {
		r.logger.Error("Failed to decrement cache value", err, "key", key)
		return 0, err
	}

	return value, nil
}

// SetNX sets a value only if the key doesn't exist
func (r *RedisCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	// Serialize value
	var serializedValue interface{}
	switch v := value.(type) {
	case string:
		serializedValue = v
	case []byte:
		serializedValue = string(v)
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			return false, fmt.Errorf("failed to serialize value: %w", err)
		}
		serializedValue = string(jsonData)
	}

	success, err := r.client.SetNX(ctx, key, serializedValue, expiration)
	if err != nil {
		r.logger.Error("Failed to set cache value with NX", err, "key", key)
		return false, err
	}

	return success, nil
}

// GetSet atomically sets a value and returns the old value
func (r *RedisCache) GetSet(ctx context.Context, key string, value interface{}) ([]byte, error) {
	// Serialize value
	var serializedValue interface{}
	switch v := value.(type) {
	case string:
		serializedValue = v
	case []byte:
		serializedValue = string(v)
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize value: %w", err)
		}
		serializedValue = string(jsonData)
	}

	oldValue, err := r.client.GetSet(ctx, key, serializedValue)
	if err != nil {
		r.logger.Error("Failed to get-set cache value", err, "key", key)
		return nil, err
	}

	if oldValue == "" {
		return nil, ErrCacheMiss
	}

	return []byte(oldValue), nil
}

// MGet retrieves multiple values from cache
func (r *RedisCache) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	start := time.Now()
	defer func() {
		r.logger.Debug("Cache MGET operation", "keys_count", len(keys), "duration", time.Since(start))
	}()

	values, err := r.client.MGet(ctx, keys...)
	if err != nil {
		r.logger.Error("Failed to get multiple values from cache", err, "keys_count", len(keys))
		return nil, err
	}

	return values, nil
}

// MSet sets multiple key-value pairs
func (r *RedisCache) MSet(ctx context.Context, pairs ...interface{}) error {
	start := time.Now()
	defer func() {
		r.logger.Debug("Cache MSET operation", "pairs_count", len(pairs)/2, "duration", time.Since(start))
	}()

	// Serialize values in pairs
	serializedPairs := make([]interface{}, len(pairs))
	for i := 0; i < len(pairs); i += 2 {
		// Key (should be string)
		serializedPairs[i] = pairs[i]
		
		// Value (serialize if needed)
		value := pairs[i+1]
		switch v := value.(type) {
		case string:
			serializedPairs[i+1] = v
		case []byte:
			serializedPairs[i+1] = string(v)
		default:
			jsonData, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to serialize value at index %d: %w", i+1, err)
			}
			serializedPairs[i+1] = string(jsonData)
		}
	}

	err := r.client.MSet(ctx, serializedPairs...)
	if err != nil {
		r.logger.Error("Failed to set multiple cache values", err, "pairs_count", len(pairs)/2)
		return err
	}

	return nil
}

// Expire sets expiration time for a key
func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	_, err := r.client.Expire(ctx, key, expiration)
	if err != nil {
		r.logger.Error("Failed to set cache expiration", err, "key", key, "expiration", expiration)
		return err
	}

	return nil
}

// TTL returns the time to live for a key
func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key)
	if err != nil {
		r.logger.Error("Failed to get cache TTL", err, "key", key)
		return 0, err
	}

	return ttl, nil
}

// FlushAll removes all keys from cache
func (r *RedisCache) FlushAll(ctx context.Context) error {
	err := r.client.FlushAll(ctx)
	if err != nil {
		r.logger.Error("Failed to flush cache", err)
		return err
	}

	r.logger.Info("Cache flushed successfully")
	return nil
}

// Close closes the cache connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// Health checks cache health
func (r *RedisCache) Health(ctx context.Context) error {
	return r.client.Ping(ctx)
}

// Cache errors
var (
	ErrCacheMiss = fmt.Errorf("cache miss")
	ErrCacheConnectionFailed = fmt.Errorf("cache connection failed")
)

// CacheKeyBuilder helps build consistent cache keys
type CacheKeyBuilder struct {
	prefix string
}

// NewCacheKeyBuilder creates a new cache key builder
func NewCacheKeyBuilder(prefix string) *CacheKeyBuilder {
	return &CacheKeyBuilder{prefix: prefix}
}

// WorkflowKey builds a cache key for workflow data
func (ckb *CacheKeyBuilder) WorkflowKey(workflowID string) string {
	return fmt.Sprintf("%s:workflow:%s", ckb.prefix, workflowID)
}

// ExecutionKey builds a cache key for execution data
func (ckb *CacheKeyBuilder) ExecutionKey(executionID string) string {
	return fmt.Sprintf("%s:execution:%s", ckb.prefix, executionID)
}

// AgentKey builds a cache key for agent data
func (ckb *CacheKeyBuilder) AgentKey(agentType string) string {
	return fmt.Sprintf("%s:agent:%s", ckb.prefix, agentType)
}

// MetricsKey builds a cache key for metrics data
func (ckb *CacheKeyBuilder) MetricsKey(metricType string) string {
	return fmt.Sprintf("%s:metrics:%s", ckb.prefix, metricType)
}

// AnalyticsKey builds a cache key for analytics data
func (ckb *CacheKeyBuilder) AnalyticsKey(analyticsType string) string {
	return fmt.Sprintf("%s:analytics:%s", ckb.prefix, analyticsType)
}

// SessionKey builds a cache key for session data
func (ckb *CacheKeyBuilder) SessionKey(sessionID string) string {
	return fmt.Sprintf("%s:session:%s", ckb.prefix, sessionID)
}

// LockKey builds a cache key for distributed locks
func (ckb *CacheKeyBuilder) LockKey(resource string) string {
	return fmt.Sprintf("%s:lock:%s", ckb.prefix, resource)
}

// RateLimitKey builds a cache key for rate limiting
func (ckb *CacheKeyBuilder) RateLimitKey(identifier string) string {
	return fmt.Sprintf("%s:ratelimit:%s", ckb.prefix, identifier)
}

// CacheStats represents cache statistics
type CacheStats struct {
	HitCount    int64   `json:"hit_count"`
	MissCount   int64   `json:"miss_count"`
	HitRate     float64 `json:"hit_rate"`
	TotalOps    int64   `json:"total_ops"`
	AvgLatency  time.Duration `json:"avg_latency"`
	LastUpdated time.Time `json:"last_updated"`
}

// CacheMonitor monitors cache performance
type CacheMonitor struct {
	stats  *CacheStats
	logger Logger
}

// NewCacheMonitor creates a new cache monitor
func NewCacheMonitor(logger Logger) *CacheMonitor {
	return &CacheMonitor{
		stats: &CacheStats{
			LastUpdated: time.Now(),
		},
		logger: logger,
	}
}

// RecordHit records a cache hit
func (cm *CacheMonitor) RecordHit(latency time.Duration) {
	cm.stats.HitCount++
	cm.stats.TotalOps++
	cm.updateStats(latency)
}

// RecordMiss records a cache miss
func (cm *CacheMonitor) RecordMiss(latency time.Duration) {
	cm.stats.MissCount++
	cm.stats.TotalOps++
	cm.updateStats(latency)
}

// updateStats updates cache statistics
func (cm *CacheMonitor) updateStats(latency time.Duration) {
	if cm.stats.TotalOps > 0 {
		cm.stats.HitRate = float64(cm.stats.HitCount) / float64(cm.stats.TotalOps) * 100
	}
	
	// Simple moving average for latency
	if cm.stats.AvgLatency == 0 {
		cm.stats.AvgLatency = latency
	} else {
		cm.stats.AvgLatency = (cm.stats.AvgLatency + latency) / 2
	}
	
	cm.stats.LastUpdated = time.Now()
}

// GetStats returns current cache statistics
func (cm *CacheMonitor) GetStats() *CacheStats {
	statsCopy := *cm.stats
	return &statsCopy
}
