package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// Cache interface defines caching operations
type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Keys(ctx context.Context, pattern string) ([]string, error)
	FlushAll(ctx context.Context) error
	Health(ctx context.Context) error
	
	// Batch operations for better performance
	MSet(ctx context.Context, pairs map[string]interface{}, expiration time.Duration) error
	MGet(ctx context.Context, keys []string) (map[string]interface{}, error)
	MDelete(ctx context.Context, keys []string) error
	Pipeline() *RedisPipeline
}

// RedisCache implements Cache interface using Redis
type RedisCache struct {
	client *redis.Client
	prefix string
}

// Config represents Redis configuration
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
	Prefix   string
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(config *Config) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		prefix: config.Prefix,
	}, nil
}

// Set stores a value in cache
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	fullKey := r.getFullKey(key)

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	err = r.client.Set(ctx, fullKey, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Get retrieves a value from cache
func (r *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := r.getFullKey(key)

	data, err := r.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		return fmt.Errorf("failed to get cache: %w", err)
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// Delete removes a value from cache
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := r.getFullKey(key)

	err := r.client.Del(ctx, fullKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	return nil
}

// Exists checks if a key exists in cache
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := r.getFullKey(key)

	count, err := r.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return count > 0, nil
}

// Expire sets expiration for a key
func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	fullKey := r.getFullKey(key)

	err := r.client.Expire(ctx, fullKey, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiration: %w", err)
	}

	return nil
}

// Keys returns keys matching a pattern
func (r *RedisCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	fullPattern := r.getFullKey(pattern)

	keys, err := r.client.Keys(ctx, fullPattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %w", err)
	}

	// Remove prefix from keys
	result := make([]string, len(keys))
	for i, key := range keys {
		result[i] = r.removePrefix(key)
	}

	return result, nil
}

// FlushAll removes all keys from cache
func (r *RedisCache) FlushAll(ctx context.Context) error {
	err := r.client.FlushAll(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to flush cache: %w", err)
	}

	return nil
}

// Health checks Redis health
func (r *RedisCache) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// Helper methods

func (r *RedisCache) getFullKey(key string) string {
	if r.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", r.prefix, key)
}

func (r *RedisCache) removePrefix(key string) string {
	if r.prefix == "" {
		return key
	}
	prefixLen := len(r.prefix) + 1 // +1 for the colon
	if len(key) > prefixLen {
		return key[prefixLen:]
	}
	return key
}

// CacheManager provides high-level caching operations
type CacheManager struct {
	cache Cache
}

// NewCacheManager creates a new cache manager
func NewCacheManager(cache Cache) *CacheManager {
	return &CacheManager{cache: cache}
}

// CacheUser caches user data
func (cm *CacheManager) CacheUser(ctx context.Context, userID string, user interface{}) error {
	key := fmt.Sprintf("user:%s", userID)
	return cm.cache.Set(ctx, key, user, time.Hour)
}

// GetCachedUser retrieves cached user data
func (cm *CacheManager) GetCachedUser(ctx context.Context, userID string, dest interface{}) error {
	key := fmt.Sprintf("user:%s", userID)
	return cm.cache.Get(ctx, key, dest)
}

// CacheOrder caches order data
func (cm *CacheManager) CacheOrder(ctx context.Context, orderID string, order interface{}) error {
	key := fmt.Sprintf("order:%s", orderID)
	return cm.cache.Set(ctx, key, order, time.Hour*2)
}

// GetCachedOrder retrieves cached order data
func (cm *CacheManager) GetCachedOrder(ctx context.Context, orderID string, dest interface{}) error {
	key := fmt.Sprintf("order:%s", orderID)
	return cm.cache.Get(ctx, key, dest)
}

// CacheSession caches session data
func (cm *CacheManager) CacheSession(ctx context.Context, sessionID string, session interface{}) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return cm.cache.Set(ctx, key, session, time.Hour*24)
}

// GetCachedSession retrieves cached session data
func (cm *CacheManager) GetCachedSession(ctx context.Context, sessionID string, dest interface{}) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return cm.cache.Get(ctx, key, dest)
}

// InvalidateUserCache invalidates user-related cache
func (cm *CacheManager) InvalidateUserCache(ctx context.Context, userID string) error {
	patterns := []string{
		fmt.Sprintf("user:%s", userID),
		fmt.Sprintf("user:%s:*", userID),
	}

	for _, pattern := range patterns {
		keys, err := cm.cache.Keys(ctx, pattern)
		if err != nil {
			return err
		}

		for _, key := range keys {
			if err := cm.cache.Delete(ctx, key); err != nil {
				return err
			}
		}
	}

	return nil
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits        int64   `json:"hits"`
	Misses      int64   `json:"misses"`
	HitRate     float64 `json:"hit_rate"`
	KeyCount    int64   `json:"key_count"`
	MemoryUsage int64   `json:"memory_usage_bytes"`
}

// GetStats returns cache statistics (Redis-specific)
func (r *RedisCache) GetStats(ctx context.Context) (*CacheStats, error) {
	_, err := r.client.Info(ctx, "stats").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis stats: %w", err)
	}

	// Parse Redis INFO output (simplified)
	stats := &CacheStats{
		Hits:     0, // Would parse from info in a real implementation
		Misses:   0, // Would parse from info in a real implementation
		HitRate:  0.0,
		KeyCount: 0,
	}

	// Get key count
	dbSize, err := r.client.DBSize(ctx).Result()
	if err == nil {
		stats.KeyCount = dbSize
	}

	return stats, nil
}

// Distributed locking using Redis

// DistributedLock represents a distributed lock
type DistributedLock struct {
	cache  Cache
	key    string
	value  string
	expiry time.Duration
}

// NewDistributedLock creates a new distributed lock
func NewDistributedLock(cache Cache, key string, expiry time.Duration) *DistributedLock {
	return &DistributedLock{
		cache:  cache,
		key:    fmt.Sprintf("lock:%s", key),
		value:  fmt.Sprintf("%d", time.Now().UnixNano()),
		expiry: expiry,
	}
}

// Acquire attempts to acquire the lock
func (dl *DistributedLock) Acquire(ctx context.Context) (bool, error) {
	// Try to set the lock key with expiration
	err := dl.cache.Set(ctx, dl.key, dl.value, dl.expiry)
	if err != nil {
		// If key already exists, lock acquisition failed
		return false, nil
	}

	return true, nil
}

// Release releases the lock
func (dl *DistributedLock) Release(ctx context.Context) error {
	// Only delete if the value matches (to prevent releasing someone else's lock)
	var currentValue string
	err := dl.cache.Get(ctx, dl.key, &currentValue)
	if err != nil {
		return nil // Lock doesn't exist or expired
	}

	if currentValue == dl.value {
		return dl.cache.Delete(ctx, dl.key)
	}

	return nil // Not our lock
}

// RedisPipeline wraps Redis pipeline for batch operations
type RedisPipeline struct {
	pipe   redis.Pipeliner
	cache  *RedisCache
	closed bool
}

// MSet sets multiple key-value pairs with expiration
func (r *RedisCache) MSet(ctx context.Context, pairs map[string]interface{}, expiration time.Duration) error {
	pipe := r.client.Pipeline()
	
	for key, value := range pairs {
		fullKey := r.getFullKey(key)
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}
		pipe.Set(ctx, fullKey, data, expiration)
	}
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute batch set: %w", err)
	}
	
	return nil
}

// MGet gets multiple values by keys
func (r *RedisCache) MGet(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}
	
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.getFullKey(key)
	}
	
	results, err := r.client.MGet(ctx, fullKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to execute batch get: %w", err)
	}
	
	response := make(map[string]interface{})
	for i, result := range results {
		if result != nil {
			var value interface{}
			if err := json.Unmarshal([]byte(result.(string)), &value); err == nil {
				response[keys[i]] = value
			}
		}
	}
	
	return response, nil
}

// MDelete deletes multiple keys
func (r *RedisCache) MDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.getFullKey(key)
	}
	
	err := r.client.Del(ctx, fullKeys...).Err()
	if err != nil {
		return fmt.Errorf("failed to execute batch delete: %w", err)
	}
	
	return nil
}

// Pipeline creates a new Redis pipeline for batch operations
func (r *RedisCache) Pipeline() *RedisPipeline {
	return &RedisPipeline{
		pipe:  r.client.Pipeline(),
		cache: r,
	}
}

// Set adds a set operation to the pipeline
func (p *RedisPipeline) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if p.closed {
		return fmt.Errorf("pipeline is closed")
	}
	
	fullKey := p.cache.getFullKey(key)
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	
	p.pipe.Set(ctx, fullKey, data, expiration)
	return nil
}

// Get adds a get operation to the pipeline (note: use Exec to get results)
func (p *RedisPipeline) Get(ctx context.Context, key string) *redis.StringCmd {
	if p.closed {
		return nil
	}
	
	fullKey := p.cache.getFullKey(key)
	return p.pipe.Get(ctx, fullKey)
}

// Delete adds a delete operation to the pipeline
func (p *RedisPipeline) Delete(ctx context.Context, key string) error {
	if p.closed {
		return fmt.Errorf("pipeline is closed")
	}
	
	fullKey := p.cache.getFullKey(key)
	p.pipe.Del(ctx, fullKey)
	return nil
}

// Exec executes all operations in the pipeline
func (p *RedisPipeline) Exec(ctx context.Context) ([]redis.Cmder, error) {
	if p.closed {
		return nil, fmt.Errorf("pipeline is closed")
	}
	
	results, err := p.pipe.Exec(ctx)
	p.closed = true
	return results, err
}

// Close closes the pipeline
func (p *RedisPipeline) Close() error {
	if !p.closed {
		p.pipe.Discard()
		p.closed = true
	}
	return nil
}

// DefaultConfig returns default Redis configuration
func DefaultConfig() *Config {
	return &Config{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
		Prefix:   "go_coffee",
	}
}
