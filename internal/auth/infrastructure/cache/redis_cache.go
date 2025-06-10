package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/application"
	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// RedisCache implements the CacheService interface using Redis
type RedisCache struct {
	client *redis.Client
	logger *logger.Logger
	prefix string
}

// Config represents Redis cache configuration
type Config struct {
	Host     string        `yaml:"host"`
	Port     int           `yaml:"port"`
	Password string        `yaml:"password"`
	DB       int           `yaml:"db"`
	Prefix   string        `yaml:"prefix"`
	PoolSize int           `yaml:"pool_size"`
	Timeout  time.Duration `yaml:"timeout"`
}

// NewRedisCache creates a new Redis cache service
func NewRedisCache(config *Config, logger *logger.Logger) (application.CacheService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Connected to Redis successfully")

	return &RedisCache{
		client: client,
		logger: logger,
		prefix: config.Prefix,
	}, nil
}

// Set stores a value in the cache with expiration
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	fullKey := r.buildKey(key)

	// Serialize value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		r.logger.WithError(err).WithField("key", fullKey).Error("Failed to marshal cache value")
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	err = r.client.Set(ctx, fullKey, data, expiration).Err()
	if err != nil {
		r.logger.WithError(err).WithField("key", fullKey).Error("Failed to set cache value")
		return fmt.Errorf("failed to set cache value: %w", err)
	}

	r.logger.WithFields(map[string]interface{}{
		"key":        fullKey,
		"expiration": expiration,
	}).Debug("Cache value set successfully")

	return nil
}

// Get retrieves a value from the cache (interface method)
func (r *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	fullKey := r.buildKey(key)

	data, err := r.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, application.ErrCacheKeyNotFound
		}
		r.logger.WithError(err).WithField("key", fullKey).Error("Failed to get cache value")
		return nil, fmt.Errorf("failed to get cache value: %w", err)
	}

	// Deserialize JSON to interface{}
	var result interface{}
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		r.logger.WithError(err).WithField("key", fullKey).Error("Failed to unmarshal cache value")
		return nil, fmt.Errorf("failed to unmarshal cache value: %w", err)
	}

	r.logger.WithField("key", fullKey).Debug("Cache value retrieved successfully")
	return result, nil
}

// GetTyped retrieves a value from the cache into a specific type
func (r *RedisCache) GetTyped(ctx context.Context, key string, dest interface{}) error {
	fullKey := r.buildKey(key)

	data, err := r.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return application.ErrCacheKeyNotFound
		}
		r.logger.WithError(err).WithField("key", fullKey).Error("Failed to get cache value")
		return fmt.Errorf("failed to get cache value: %w", err)
	}

	// Deserialize JSON to destination
	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		r.logger.WithError(err).WithField("key", fullKey).Error("Failed to unmarshal cache value")
		return fmt.Errorf("failed to unmarshal cache value: %w", err)
	}

	r.logger.WithField("key", fullKey).Debug("Cache value retrieved successfully")
	return nil
}

// Delete removes a value from the cache
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := r.buildKey(key)

	err := r.client.Del(ctx, fullKey).Err()
	if err != nil {
		r.logger.WithError(err).WithField("key", fullKey).Error("Failed to delete cache value")
		return fmt.Errorf("failed to delete cache value: %w", err)
	}

	r.logger.WithField("key", fullKey).Debug("Cache value deleted successfully")
	return nil
}

// Exists checks if a key exists in the cache
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := r.buildKey(key)

	count, err := r.client.Exists(ctx, fullKey).Result()
	if err != nil {
		r.logger.WithError(err).WithField("key", fullKey).Error("Failed to check cache key existence")
		return false, fmt.Errorf("failed to check cache key existence: %w", err)
	}

	return count > 0, nil
}

// SetNX sets a value only if the key doesn't exist (atomic operation)
func (r *RedisCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	fullKey := r.buildKey(key)

	// Serialize value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		r.logger.WithError(err).WithField("key", fullKey).Error("Failed to marshal cache value")
		return false, fmt.Errorf("failed to marshal cache value: %w", err)
	}

	success, err := r.client.SetNX(ctx, fullKey, data, expiration).Result()
	if err != nil {
		r.logger.WithError(err).WithField("key", fullKey).Error("Failed to set cache value with NX")
		return false, fmt.Errorf("failed to set cache value with NX: %w", err)
	}

	if success {
		r.logger.WithFields(map[string]interface{}{
			"key":        fullKey,
			"expiration": expiration,
		}).Debug("Cache value set with NX successfully")
	}

	return success, nil
}

// Increment atomically increments a numeric value
func (r *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	fullKey := r.buildKey(key)

	value, err := r.client.Incr(ctx, fullKey).Result()
	if err != nil {
		r.logger.WithError(err).WithField("key", fullKey).Error("Failed to increment cache value")
		return 0, fmt.Errorf("failed to increment cache value: %w", err)
	}

	r.logger.WithFields(map[string]interface{}{
		"key":   fullKey,
		"value": value,
	}).Debug("Cache value incremented successfully")

	return value, nil
}

// IncrementWithExpiry atomically increments a numeric value and sets expiration
func (r *RedisCache) IncrementWithExpiry(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	fullKey := r.buildKey(key)

	// Use pipeline for atomic operation
	pipe := r.client.Pipeline()
	incrCmd := pipe.Incr(ctx, fullKey)
	pipe.Expire(ctx, fullKey, expiration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		r.logger.WithError(err).WithField("key", fullKey).Error("Failed to increment cache value with expiry")
		return 0, fmt.Errorf("failed to increment cache value with expiry: %w", err)
	}

	value := incrCmd.Val()
	r.logger.WithFields(map[string]interface{}{
		"key":        fullKey,
		"value":      value,
		"expiration": expiration,
	}).Debug("Cache value incremented with expiry successfully")

	return value, nil
}

// GetTTL returns the time-to-live for a key
func (r *RedisCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := r.buildKey(key)

	ttl, err := r.client.TTL(ctx, fullKey).Result()
	if err != nil {
		r.logger.WithError(err).WithField("key", fullKey).Error("Failed to get cache key TTL")
		return 0, fmt.Errorf("failed to get cache key TTL: %w", err)
	}

	return ttl, nil
}

// SetExpiry sets expiration for an existing key
func (r *RedisCache) SetExpiry(ctx context.Context, key string, expiration time.Duration) error {
	fullKey := r.buildKey(key)

	success, err := r.client.Expire(ctx, fullKey, expiration).Result()
	if err != nil {
		r.logger.WithError(err).WithField("key", fullKey).Error("Failed to set cache key expiry")
		return fmt.Errorf("failed to set cache key expiry: %w", err)
	}

	if !success {
		return application.ErrCacheKeyNotFound
	}

	r.logger.WithFields(map[string]interface{}{
		"key":        fullKey,
		"expiration": expiration,
	}).Debug("Cache key expiry set successfully")

	return nil
}

// FlushPattern deletes all keys matching a pattern
func (r *RedisCache) FlushPattern(ctx context.Context, pattern string) error {
	fullPattern := r.buildKey(pattern)

	// Get all keys matching the pattern
	keys, err := r.client.Keys(ctx, fullPattern).Result()
	if err != nil {
		r.logger.WithError(err).WithField("pattern", fullPattern).Error("Failed to get keys by pattern")
		return fmt.Errorf("failed to get keys by pattern: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	// Delete all matching keys
	err = r.client.Del(ctx, keys...).Err()
	if err != nil {
		r.logger.WithError(err).WithField("pattern", fullPattern).Error("Failed to delete keys by pattern")
		return fmt.Errorf("failed to delete keys by pattern: %w", err)
	}

	r.logger.WithFields(map[string]interface{}{
		"pattern":      fullPattern,
		"keys_deleted": len(keys),
	}).Info("Cache keys deleted by pattern")

	return nil
}

// GetMultiple retrieves multiple values from the cache
func (r *RedisCache) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	// Build full keys
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.buildKey(key)
	}

	// Get all values
	values, err := r.client.MGet(ctx, fullKeys...).Result()
	if err != nil {
		r.logger.WithError(err).WithField("keys", fullKeys).Error("Failed to get multiple cache values")
		return nil, fmt.Errorf("failed to get multiple cache values: %w", err)
	}

	// Build result map
	result := make(map[string]interface{})
	for i, value := range values {
		if value != nil {
			var dest interface{}
			if err := json.Unmarshal([]byte(value.(string)), &dest); err != nil {
				r.logger.WithError(err).WithField("key", keys[i]).Error("Failed to unmarshal cache value")
				continue
			}
			result[keys[i]] = dest
		}
	}

	r.logger.WithFields(map[string]interface{}{
		"keys_requested": len(keys),
		"keys_found":     len(result),
	}).Debug("Multiple cache values retrieved")

	return result, nil
}

// SetMultiple stores multiple values in the cache
func (r *RedisCache) SetMultiple(ctx context.Context, values map[string]interface{}, expiration time.Duration) error {
	if len(values) == 0 {
		return nil
	}

	// Use pipeline for efficiency
	pipe := r.client.Pipeline()

	for key, value := range values {
		fullKey := r.buildKey(key)

		// Serialize value to JSON
		data, err := json.Marshal(value)
		if err != nil {
			r.logger.WithError(err).WithField("key", fullKey).Error("Failed to marshal cache value")
			continue
		}

		pipe.Set(ctx, fullKey, data, expiration)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		r.logger.WithError(err).Error("Failed to set multiple cache values")
		return fmt.Errorf("failed to set multiple cache values: %w", err)
	}

	r.logger.WithFields(map[string]interface{}{
		"keys_set":   len(values),
		"expiration": expiration,
	}).Debug("Multiple cache values set successfully")

	return nil
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// Health checks the health of the Redis connection
func (r *RedisCache) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// GetStats returns Redis connection statistics
func (r *RedisCache) GetStats() map[string]interface{} {
	stats := r.client.PoolStats()
	return map[string]interface{}{
		"hits":        stats.Hits,
		"misses":      stats.Misses,
		"timeouts":    stats.Timeouts,
		"total_conns": stats.TotalConns,
		"idle_conns":  stats.IdleConns,
		"stale_conns": stats.StaleConns,
	}
}

// Interface methods for CacheService

// SetString sets a string value in the cache
func (r *RedisCache) SetString(ctx context.Context, key, value string, expiration time.Duration) error {
	fullKey := r.buildKey(key)
	err := r.client.Set(ctx, fullKey, value, expiration).Err()
	if err != nil {
		r.logger.WithError(err).WithField("key", fullKey).Error("Failed to set string cache value")
		return fmt.Errorf("failed to set string cache value: %w", err)
	}
	return nil
}

// GetString gets a string value from the cache
func (r *RedisCache) GetString(ctx context.Context, key string) (string, error) {
	fullKey := r.buildKey(key)
	value, err := r.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", application.ErrCacheKeyNotFound
		}
		r.logger.WithError(err).WithField("key", fullKey).Error("Failed to get string cache value")
		return "", fmt.Errorf("failed to get string cache value: %w", err)
	}
	return value, nil
}

// SetInt sets an integer value in the cache
func (r *RedisCache) SetInt(ctx context.Context, key string, value int, expiration time.Duration) error {
	fullKey := r.buildKey(key)
	err := r.client.Set(ctx, fullKey, value, expiration).Err()
	if err != nil {
		r.logger.WithError(err).WithField("key", fullKey).Error("Failed to set int cache value")
		return fmt.Errorf("failed to set int cache value: %w", err)
	}
	return nil
}

// GetInt gets an integer value from the cache
func (r *RedisCache) GetInt(ctx context.Context, key string) (int, error) {
	fullKey := r.buildKey(key)
	value, err := r.client.Get(ctx, fullKey).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, application.ErrCacheKeyNotFound
		}
		r.logger.WithError(err).WithField("key", fullKey).Error("Failed to get int cache value")
		return 0, fmt.Errorf("failed to get int cache value: %w", err)
	}
	return value, nil
}

// Decrement atomically decrements a numeric value
func (r *RedisCache) Decrement(ctx context.Context, key string) (int64, error) {
	fullKey := r.buildKey(key)
	value, err := r.client.Decr(ctx, fullKey).Result()
	if err != nil {
		r.logger.WithError(err).WithField("key", fullKey).Error("Failed to decrement cache value")
		return 0, fmt.Errorf("failed to decrement cache value: %w", err)
	}
	return value, nil
}

// SetUserSession stores a user session in the cache
func (r *RedisCache) SetUserSession(ctx context.Context, sessionID string, session *domain.Session, expiration time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.Set(ctx, key, session, expiration)
}

// GetUserSession retrieves a user session from the cache
func (r *RedisCache) GetUserSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	var session domain.Session
	err := r.GetTyped(ctx, key, &session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// DeleteUserSession removes a user session from the cache
func (r *RedisCache) DeleteUserSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.Delete(ctx, key)
}

// AddToBlacklist adds a token to the blacklist
func (r *RedisCache) AddToBlacklist(ctx context.Context, tokenID string, expiration time.Duration) error {
	key := fmt.Sprintf("blacklist:%s", tokenID)
	return r.SetString(ctx, key, "blacklisted", expiration)
}

// IsBlacklisted checks if a token is blacklisted
func (r *RedisCache) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", tokenID)
	exists, err := r.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// RemoveFromBlacklist removes a token from the blacklist
func (r *RedisCache) RemoveFromBlacklist(ctx context.Context, tokenID string) error {
	key := fmt.Sprintf("blacklist:%s", tokenID)
	return r.Delete(ctx, key)
}

// buildKey builds the full cache key with prefix
func (r *RedisCache) buildKey(key string) string {
	if r.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", r.prefix, key)
}

// DefaultConfig returns default Redis cache configuration
func DefaultConfig() *Config {
	return &Config{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
		Prefix:   "auth",
		PoolSize: 10,
		Timeout:  5 * time.Second,
	}
}
