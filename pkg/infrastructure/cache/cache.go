package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/config"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/redis"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Cache represents a cache interface
type Cache interface {
	// Basic operations
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// Batch operations
	GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error)
	SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
	DeleteMulti(ctx context.Context, keys []string) error

	// Advanced operations
	Increment(ctx context.Context, key string, delta int64) (int64, error)
	Decrement(ctx context.Context, key string, delta int64) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Pattern operations
	Keys(ctx context.Context, pattern string) ([]string, error)
	DeletePattern(ctx context.Context, pattern string) error

	// Cache management
	Clear(ctx context.Context) error
	Stats(ctx context.Context) (*Stats, error)

	// Connection management
	Ping(ctx context.Context) error
	Close() error
}

// Stats represents cache statistics
type Stats struct {
	Hits        int64 `json:"hits"`
	Misses      int64 `json:"misses"`
	Keys        int64 `json:"keys"`
	Memory      int64 `json:"memory"`
	Connections int   `json:"connections"`
}

// RedisCache implements Cache interface using Redis
type RedisCache struct {
	client redis.ClientInterface
	config *config.CacheConfig
	logger *logger.Logger
	stats  *Stats
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(redisClient redis.ClientInterface, cfg *config.CacheConfig, logger *logger.Logger) Cache {
	return &RedisCache{
		client: redisClient,
		config: cfg,
		logger: logger,
		stats:  &Stats{},
	}
}

// Get retrieves a value from cache
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key)
	if err != nil {
		c.stats.Misses++
		if err.Error() == "redis: nil" {
			return ErrCacheKeyNotFound
		}
		return fmt.Errorf("failed to get cache key %s: %w", key, err)
	}

	c.stats.Hits++

	if err := c.deserialize([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to deserialize cache value for key %s: %w", key, err)
	}

	return nil
}

// Set stores a value in cache
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}

	data, err := c.serialize(value)
	if err != nil {
		return fmt.Errorf("failed to serialize cache value for key %s: %w", key, err)
	}

	if err := c.client.Set(ctx, key, string(data), ttl); err != nil {
		return fmt.Errorf("failed to set cache key %s: %w", key, err)
	}

	return nil
}

// Delete removes a key from cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key); err != nil {
		return fmt.Errorf("failed to delete cache key %s: %w", key, err)
	}
	return nil
}

// Exists checks if a key exists in cache
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to check cache key existence %s: %w", key, err)
	}
	return count > 0, nil
}

// GetMulti retrieves multiple values from cache
func (c *RedisCache) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Use pipeline for better performance
	pipe := c.client.GetClient().Pipeline()
	cmds := make(map[string]interface{})

	for _, key := range keys {
		cmds[key] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err.Error() != "redis: nil" {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	for key := range cmds {
		// Get individual values instead of using pipeline results
		data, err := c.client.Get(ctx, key)
		if err != nil {
			if err.Error() == "redis: nil" {
				c.stats.Misses++
				continue
			}
			return nil, fmt.Errorf("failed to get cache key %s: %w", key, err)
		}

		c.stats.Hits++

		var value interface{}
		if err := c.deserialize([]byte(data), &value); err != nil {
			c.logger.WithError(err).WithField("key", key).Error("Failed to deserialize cache value")
			continue
		}

		result[key] = value
	}

	return result, nil
}

// SetMulti stores multiple values in cache
func (c *RedisCache) SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}

	pipe := c.client.Pipeline()

	for key, value := range items {
		data, err := c.serialize(value)
		if err != nil {
			return fmt.Errorf("failed to serialize cache value for key %s: %w", key, err)
		}

		pipe.Set(ctx, key, string(data), ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}

	return nil
}

// DeleteMulti removes multiple keys from cache
func (c *RedisCache) DeleteMulti(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	if err := c.client.Del(ctx, keys...); err != nil {
		return fmt.Errorf("failed to delete cache keys: %w", err)
	}

	return nil
}

// Increment increments a numeric value
func (c *RedisCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	result, err := c.client.GetClient().IncrBy(ctx, key, delta).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment cache key %s: %w", key, err)
	}
	return result, nil
}

// Decrement decrements a numeric value
func (c *RedisCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	result, err := c.client.GetClient().DecrBy(ctx, key, delta).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement cache key %s: %w", key, err)
	}
	return result, nil
}

// Expire sets expiration for a key
func (c *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := c.client.Expire(ctx, key, ttl); err != nil {
		return fmt.Errorf("failed to set expiration for cache key %s: %w", key, err)
	}
	return nil
}

// TTL returns time to live for a key
func (c *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.client.TTL(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for cache key %s: %w", key, err)
	}
	return ttl, nil
}

// Keys returns keys matching a pattern
func (c *RedisCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	keys, err := c.client.Keys(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to get keys with pattern %s: %w", pattern, err)
	}
	return keys, nil
}

// DeletePattern deletes keys matching a pattern
func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := c.client.Keys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to get keys with pattern %s: %w", pattern, err)
	}

	if len(keys) == 0 {
		return nil
	}

	if err := c.client.Del(ctx, keys...); err != nil {
		return fmt.Errorf("failed to delete keys with pattern %s: %w", pattern, err)
	}

	return nil
}

// Clear removes all keys from cache
func (c *RedisCache) Clear(ctx context.Context) error {
	if err := c.client.FlushDB(ctx); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	return nil
}

// Stats returns cache statistics
func (c *RedisCache) Stats(ctx context.Context) (*Stats, error) {
	poolStats := c.client.PoolStats()

	// Get key count using DBSIZE command
	keyCount, err := c.client.GetClient().DBSize(ctx).Result()
	if err != nil {
		c.logger.WithError(err).Error("Failed to get key count")
		keyCount = 0
	}

	return &Stats{
		Hits:        c.stats.Hits,
		Misses:      c.stats.Misses,
		Keys:        keyCount,
		Memory:      0, // Redis doesn't provide easy memory usage per database
		Connections: int(poolStats.TotalConns),
	}, nil
}

// Ping checks cache connectivity
func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx)
}

// Close closes the cache connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// serialize converts a value to bytes
func (c *RedisCache) serialize(value interface{}) ([]byte, error) {
	switch c.config.Serialization {
	case "json":
		return json.Marshal(value)
	default:
		return json.Marshal(value)
	}
}

// deserialize converts bytes to a value
func (c *RedisCache) deserialize(data []byte, dest interface{}) error {
	switch c.config.Serialization {
	case "json":
		return json.Unmarshal(data, dest)
	default:
		return json.Unmarshal(data, dest)
	}
}

// Cache errors
var (
	ErrCacheKeyNotFound = fmt.Errorf("cache key not found")
	ErrCacheTimeout     = fmt.Errorf("cache operation timeout")
	ErrCacheConnection  = fmt.Errorf("cache connection error")
)

// CacheManager manages multiple cache instances
type CacheManager struct {
	caches map[string]Cache
	logger *logger.Logger
}

// NewCacheManager creates a new cache manager
func NewCacheManager(logger *logger.Logger) *CacheManager {
	return &CacheManager{
		caches: make(map[string]Cache),
		logger: logger,
	}
}

// AddCache adds a cache instance
func (cm *CacheManager) AddCache(name string, cache Cache) {
	cm.caches[name] = cache
}

// GetCache returns a cache instance by name
func (cm *CacheManager) GetCache(name string) (Cache, bool) {
	cache, exists := cm.caches[name]
	return cache, exists
}

// GetDefaultCache returns the default cache instance
func (cm *CacheManager) GetDefaultCache() Cache {
	if cache, exists := cm.caches["default"]; exists {
		return cache
	}
	return nil
}

// CloseAll closes all cache connections
func (cm *CacheManager) CloseAll() error {
	var lastErr error
	for name, cache := range cm.caches {
		if err := cache.Close(); err != nil {
			cm.logger.WithError(err).WithField("cache", name).Error("Failed to close cache")
			lastErr = err
		}
	}
	return lastErr
}

// HealthCheck checks the health of all caches
func (cm *CacheManager) HealthCheck(ctx context.Context) map[string]error {
	results := make(map[string]error)
	for name, cache := range cm.caches {
		results[name] = cache.Ping(ctx)
	}
	return results
}
