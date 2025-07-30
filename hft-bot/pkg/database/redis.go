package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/hft-bot/pkg/config"
)

// Logger interface for Redis operations
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
}

// RedisDB provides a mock Redis implementation for development
type RedisDB struct {
	config *config.RedisConfig
	logger Logger
	data   map[string]interface{} // In-memory storage for mock
	mu     sync.RWMutex           // Mutex for thread safety
}

// NewRedisDB creates a new mock Redis database connection
func NewRedisDB(cfg *config.RedisConfig, log Logger) (*RedisDB, error) {
	r := &RedisDB{
		config: cfg,
		logger: log,
		data:   make(map[string]interface{}),
	}

	log.Info("Connected to Mock Redis",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.Database,
	)

	return r, nil
}

// Close closes the Redis connection (mock - does nothing)
func (r *RedisDB) Close() error {
	r.logger.Info("Mock Redis connection closed")
	return nil
}

// HealthCheck performs a health check on Redis
func (r *RedisDB) HealthCheck(ctx context.Context) error {
	r.logger.Debug("Mock Redis health check - always healthy")
	return nil
}

// SetWithContext sets a key-value pair with expiration
func (r *RedisDB) SetWithContext(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	start := time.Now()
	
	r.mu.Lock()
	r.data[key] = value
	r.mu.Unlock()

	r.logger.Debug("Mock Redis SET operation",
		"key", key,
		"expiration", expiration,
		"duration", time.Since(start),
	)

	return nil
}

// GetWithContext gets a value by key
func (r *RedisDB) GetWithContext(ctx context.Context, key string) (string, error) {
	start := time.Now()
	
	r.mu.RLock()
	value, exists := r.data[key]
	r.mu.RUnlock()

	r.logger.Debug("Mock Redis GET operation",
		"key", key,
		"exists", exists,
		"duration", time.Since(start),
	)

	if !exists {
		return "", fmt.Errorf("key not found: %s", key)
	}

	if str, ok := value.(string); ok {
		return str, nil
	}

	return fmt.Sprintf("%v", value), nil
}

// DelWithContext deletes keys
func (r *RedisDB) DelWithContext(ctx context.Context, keys ...string) error {
	start := time.Now()
	
	r.mu.Lock()
	for _, key := range keys {
		delete(r.data, key)
	}
	r.mu.Unlock()

	r.logger.Debug("Mock Redis DEL operation",
		"keys", keys,
		"duration", time.Since(start),
	)

	return nil
}

// ExistsWithContext checks if keys exist
func (r *RedisDB) ExistsWithContext(ctx context.Context, keys ...string) (int64, error) {
	start := time.Now()
	
	r.mu.RLock()
	count := int64(0)
	for _, key := range keys {
		if _, exists := r.data[key]; exists {
			count++
		}
	}
	r.mu.RUnlock()

	r.logger.Debug("Mock Redis EXISTS operation",
		"keys", keys,
		"count", count,
		"duration", time.Since(start),
	)

	return count, nil
}

// HSetWithContext sets hash field
func (r *RedisDB) HSetWithContext(ctx context.Context, key string, values ...interface{}) error {
	start := time.Now()
	
	r.mu.Lock()
	hash, exists := r.data[key]
	if !exists {
		hash = make(map[string]interface{})
		r.data[key] = hash
	}
	
	if h, ok := hash.(map[string]interface{}); ok {
		for i := 0; i < len(values); i += 2 {
			if i+1 < len(values) {
				field := fmt.Sprintf("%v", values[i])
				value := values[i+1]
				h[field] = value
			}
		}
	}
	r.mu.Unlock()

	r.logger.Debug("Mock Redis HSET operation",
		"key", key,
		"duration", time.Since(start),
	)

	return nil
}

// HGetWithContext gets hash field
func (r *RedisDB) HGetWithContext(ctx context.Context, key, field string) (string, error) {
	start := time.Now()
	
	r.mu.RLock()
	hash, exists := r.data[key]
	r.mu.RUnlock()

	r.logger.Debug("Mock Redis HGET operation",
		"key", key,
		"field", field,
		"duration", time.Since(start),
	)

	if !exists {
		return "", fmt.Errorf("key not found: %s", key)
	}

	if h, ok := hash.(map[string]interface{}); ok {
		if value, fieldExists := h[field]; fieldExists {
			return fmt.Sprintf("%v", value), nil
		}
	}

	return "", fmt.Errorf("field not found: %s", field)
}

// PublishWithContext publishes a message to a channel
func (r *RedisDB) PublishWithContext(ctx context.Context, channel string, message interface{}) error {
	start := time.Now()

	r.logger.Debug("Mock Redis PUBLISH operation",
		"channel", channel,
		"duration", time.Since(start),
	)

	// Mock implementation - just log the publish
	r.logger.Info("Mock Redis message published",
		"channel", channel,
		"message", message,
	)

	return nil
}

// FlushAllWithContext flushes all databases
func (r *RedisDB) FlushAllWithContext(ctx context.Context) error {
	start := time.Now()
	
	r.mu.Lock()
	r.data = make(map[string]interface{})
	r.mu.Unlock()

	r.logger.Debug("Mock Redis FLUSHALL operation",
		"duration", time.Since(start),
	)

	r.logger.Info("Mock Redis databases flushed")
	return nil
}

// GetStats returns mock Redis statistics
func (r *RedisDB) GetStats() map[string]interface{} {
	r.mu.RLock()
	keyCount := len(r.data)
	r.mu.RUnlock()

	return map[string]interface{}{
		"connected":     true,
		"keys_count":    keyCount,
		"memory_usage":  "mock",
		"uptime":        "mock",
		"version":       "mock-redis-1.0.0",
		"host":          r.config.Host,
		"port":          r.config.Port,
		"database":      r.config.Database,
	}
}
