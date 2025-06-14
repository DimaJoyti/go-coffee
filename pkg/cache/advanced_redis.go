package cache

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// AdvancedRedisCache provides high-performance caching with clustering support
type AdvancedRedisCache struct {
	client      redis.UniversalClient
	logger      *zap.Logger
	config      *AdvancedCacheConfig
	metrics     *CacheMetrics
	warmupCache *WarmupManager
	mu          sync.RWMutex
}

// AdvancedCacheConfig contains advanced caching configuration
type AdvancedCacheConfig struct {
	// Cluster settings
	Addrs        []string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int

	// Performance settings
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	DialTimeout  time.Duration
	MaxRetries   int
	RetryDelay   time.Duration

	// Cache settings
	DefaultTTL     time.Duration
	MaxKeySize     int
	MaxValueSize   int
	CompressionMin int

	// Warmup settings
	WarmupEnabled   bool
	WarmupInterval  time.Duration
	WarmupBatchSize int
}

// CacheMetrics tracks cache performance
type CacheMetrics struct {
	Hits        int64
	Misses      int64
	Sets        int64
	Deletes     int64
	Errors      int64
	TotalKeys   int64
	MemoryUsage int64
	AvgLatency  time.Duration
	HitRatio    float64
	mu          sync.RWMutex
}

// WarmupManager handles cache warming strategies
type WarmupManager struct {
	cache      *AdvancedRedisCache
	strategies map[string]WarmupStrategy
	scheduler  *time.Ticker
	mu         sync.RWMutex
}

// WarmupStrategy defines cache warming behavior
type WarmupStrategy interface {
	Warmup(ctx context.Context) error
	GetKeys() []string
	GetTTL() time.Duration
}

// CacheEntry represents a cached item with metadata
type CacheEntry struct {
	Value      interface{}   `json:"value"`
	CreatedAt  time.Time     `json:"created_at"`
	AccessedAt time.Time     `json:"accessed_at"`
	TTL        time.Duration `json:"ttl"`
	Compressed bool          `json:"compressed"`
}

// NewAdvancedRedisCache creates a new advanced Redis cache
func NewAdvancedRedisCache(config *AdvancedCacheConfig, logger *zap.Logger) (*AdvancedRedisCache, error) {
	// Create Redis client based on configuration
	var client redis.UniversalClient

	if len(config.Addrs) > 1 {
		// Use cluster client for multiple addresses
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        config.Addrs,
			Password:     config.Password,
			PoolSize:     config.PoolSize,
			MinIdleConns: config.MinIdleConns,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			DialTimeout:  config.DialTimeout,
			MaxRetries:   config.MaxRetries,
		})
	} else {
		// Use single client
		client = redis.NewClient(&redis.Options{
			Addr:         config.Addrs[0],
			Password:     config.Password,
			DB:           config.DB,
			PoolSize:     config.PoolSize,
			MinIdleConns: config.MinIdleConns,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			DialTimeout:  config.DialTimeout,
			MaxRetries:   config.MaxRetries,
		})
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	cache := &AdvancedRedisCache{
		client:  client,
		logger:  logger,
		config:  config,
		metrics: &CacheMetrics{},
	}

	// Initialize warmup manager if enabled
	if config.WarmupEnabled {
		cache.warmupCache = &WarmupManager{
			cache:      cache,
			strategies: make(map[string]WarmupStrategy),
			scheduler:  time.NewTicker(config.WarmupInterval),
		}
		go cache.warmupCache.start()
	}

	// Start metrics collection
	go cache.startMetricsCollection()

	return cache, nil
}

// Set stores a value in cache with advanced features
func (c *AdvancedRedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		c.updateLatency(time.Since(start))
	}()

	// Validate key and value size
	if len(key) > c.config.MaxKeySize {
		return fmt.Errorf("key size exceeds maximum: %d", c.config.MaxKeySize)
	}

	// Create cache entry
	entry := &CacheEntry{
		Value:     value,
		CreatedAt: time.Now(),
		TTL:       ttl,
	}

	// Serialize value
	data, err := json.Marshal(entry)
	if err != nil {
		c.recordError()
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	// Check value size and compress if needed
	if len(data) > c.config.CompressionMin {
		compressed, err := c.compress(data)
		if err == nil && len(compressed) < len(data) {
			data = compressed
			entry.Compressed = true
		}
	}

	if len(data) > c.config.MaxValueSize {
		return fmt.Errorf("value size exceeds maximum: %d", c.config.MaxValueSize)
	}

	// Set with TTL
	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}

	err = c.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		c.recordError()
		return fmt.Errorf("failed to set cache key: %w", err)
	}

	c.recordSet()
	return nil
}

// Get retrieves a value from cache
func (c *AdvancedRedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	start := time.Now()
	defer func() {
		c.updateLatency(time.Since(start))
	}()

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			c.recordMiss()
			return ErrCacheMiss
		}
		c.recordError()
		return fmt.Errorf("failed to get cache key: %w", err)
	}

	// Deserialize entry
	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		// Try direct deserialization for backward compatibility
		if err := json.Unmarshal(data, dest); err != nil {
			c.recordError()
			return fmt.Errorf("failed to unmarshal cache entry: %w", err)
		}
		c.recordHit()
		return nil
	}

	// Decompress if needed
	if entry.Compressed {
		decompressed, err := c.decompress(data)
		if err != nil {
			c.recordError()
			return fmt.Errorf("failed to decompress cache entry: %w", err)
		}
		if err := json.Unmarshal(decompressed, &entry.Value); err != nil {
			c.recordError()
			return fmt.Errorf("failed to unmarshal decompressed entry: %w", err)
		}
	}

	// Update access time asynchronously
	go func() {
		c.updateAccessTime(key)
	}()

	// Copy value to destination
	valueData, err := json.Marshal(entry.Value)
	if err != nil {
		c.recordError()
		return fmt.Errorf("failed to marshal entry value: %w", err)
	}

	if err := json.Unmarshal(valueData, dest); err != nil {
		c.recordError()
		return fmt.Errorf("failed to unmarshal to destination: %w", err)
	}

	c.recordHit()
	return nil
}

// Delete removes a key from cache
func (c *AdvancedRedisCache) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		c.recordError()
		return fmt.Errorf("failed to delete cache key: %w", err)
	}

	c.recordDelete()
	return nil
}

// MGet retrieves multiple values efficiently
func (c *AdvancedRedisCache) MGet(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	start := time.Now()
	defer func() {
		c.updateLatency(time.Since(start))
	}()

	values, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		c.recordError()
		return nil, fmt.Errorf("failed to mget cache keys: %w", err)
	}

	result := make(map[string]interface{})
	for i, value := range values {
		if value != nil {
			var entry CacheEntry
			data := []byte(value.(string))

			if err := json.Unmarshal(data, &entry); err == nil {
				result[keys[i]] = entry.Value
				c.recordHit()
			} else {
				// Try direct deserialization
				var directValue interface{}
				if err := json.Unmarshal(data, &directValue); err == nil {
					result[keys[i]] = directValue
					c.recordHit()
				}
			}
		} else {
			c.recordMiss()
		}
	}

	return result, nil
}

// MSet stores multiple values efficiently
func (c *AdvancedRedisCache) MSet(ctx context.Context, pairs map[string]interface{}, ttl time.Duration) error {
	if len(pairs) == 0 {
		return nil
	}

	start := time.Now()
	defer func() {
		c.updateLatency(time.Since(start))
	}()

	// Prepare pipeline
	pipe := c.client.Pipeline()

	for key, value := range pairs {
		entry := &CacheEntry{
			Value:     value,
			CreatedAt: time.Now(),
			TTL:       ttl,
		}

		data, err := json.Marshal(entry)
		if err != nil {
			continue
		}

		if ttl == 0 {
			ttl = c.config.DefaultTTL
		}

		pipe.Set(ctx, key, data, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		c.recordError()
		return fmt.Errorf("failed to mset cache keys: %w", err)
	}

	c.metrics.mu.Lock()
	c.metrics.Sets += int64(len(pairs))
	c.metrics.mu.Unlock()

	return nil
}

// Exists checks if keys exist in cache
func (c *AdvancedRedisCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	count, err := c.client.Exists(ctx, keys...).Result()
	if err != nil {
		c.recordError()
		return 0, fmt.Errorf("failed to check key existence: %w", err)
	}
	return count, nil
}

// TTL returns the time to live for a key
func (c *AdvancedRedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		c.recordError()
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}
	return ttl, nil
}

// Expire sets a timeout on a key
func (c *AdvancedRedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	err := c.client.Expire(ctx, key, ttl).Err()
	if err != nil {
		c.recordError()
		return fmt.Errorf("failed to set expiration: %w", err)
	}
	return nil
}

// FlushPattern removes all keys matching a pattern
func (c *AdvancedRedisCache) FlushPattern(ctx context.Context, pattern string) error {
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		c.recordError()
		return fmt.Errorf("failed to get keys by pattern: %w", err)
	}

	if len(keys) > 0 {
		err = c.client.Del(ctx, keys...).Err()
		if err != nil {
			c.recordError()
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}

	return nil
}

// GetMetrics returns current cache metrics
func (c *AdvancedRedisCache) GetMetrics() CacheMetrics {
	c.metrics.mu.RLock()
	defer c.metrics.mu.RUnlock()

	// Calculate hit ratio
	total := c.metrics.Hits + c.metrics.Misses
	if total > 0 {
		c.metrics.HitRatio = float64(c.metrics.Hits) / float64(total)
	}

	return *c.metrics
}

// Helper methods for metrics
func (c *AdvancedRedisCache) recordHit() {
	c.metrics.mu.Lock()
	c.metrics.Hits++
	c.metrics.mu.Unlock()
}

func (c *AdvancedRedisCache) recordMiss() {
	c.metrics.mu.Lock()
	c.metrics.Misses++
	c.metrics.mu.Unlock()
}

func (c *AdvancedRedisCache) recordSet() {
	c.metrics.mu.Lock()
	c.metrics.Sets++
	c.metrics.mu.Unlock()
}

func (c *AdvancedRedisCache) recordDelete() {
	c.metrics.mu.Lock()
	c.metrics.Deletes++
	c.metrics.mu.Unlock()
}

func (c *AdvancedRedisCache) recordError() {
	c.metrics.mu.Lock()
	c.metrics.Errors++
	c.metrics.mu.Unlock()
}

func (c *AdvancedRedisCache) updateLatency(duration time.Duration) {
	c.metrics.mu.Lock()
	if c.metrics.AvgLatency == 0 {
		c.metrics.AvgLatency = duration
	} else {
		c.metrics.AvgLatency = (c.metrics.AvgLatency + duration) / 2
	}
	c.metrics.mu.Unlock()
}

// compress compresses data using gzip
func (c *AdvancedRedisCache) compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	if _, err := writer.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write compressed data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// decompress decompresses gzip data
func (c *AdvancedRedisCache) decompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read decompressed data: %w", err)
	}

	return decompressed, nil
}

func (c *AdvancedRedisCache) updateAccessTime(key string) {
	// Update access time for LRU tracking
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c.client.Touch(ctx, key)
}

func (c *AdvancedRedisCache) startMetricsCollection() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		c.collectSystemMetrics()
	}
}

func (c *AdvancedRedisCache) collectSystemMetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get memory usage
	_, err := c.client.Info(ctx, "memory").Result()
	if err == nil {
		// Parse memory info and update metrics
		// Implementation depends on Redis info format
		// TODO: Parse Redis memory info and update c.metrics.MemoryUsage
	}

	// Get key count
	dbSize, err := c.client.DBSize(ctx).Result()
	if err == nil {
		c.metrics.mu.Lock()
		c.metrics.TotalKeys = dbSize
		c.metrics.mu.Unlock()
	}
}

// Close closes the Redis connection
func (c *AdvancedRedisCache) Close() error {
	if c.warmupCache != nil {
		c.warmupCache.Stop()
	}
	return c.client.Close()
}

// start starts the warmup manager
func (wm *WarmupManager) start() {
	for range wm.scheduler.C {
		wm.runWarmup()
	}
}

// Stop stops the warmup manager
func (wm *WarmupManager) Stop() {
	if wm.scheduler != nil {
		wm.scheduler.Stop()
	}
}

// runWarmup executes all registered warmup strategies
func (wm *WarmupManager) runWarmup() {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	for name, strategy := range wm.strategies {
		go func(strategyName string, s WarmupStrategy) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := s.Warmup(ctx); err != nil {
				// Log error but don't fail
				fmt.Printf("Warmup strategy %s failed: %v\n", strategyName, err)
			}
		}(name, strategy)
	}
}

// AddStrategy adds a warmup strategy
func (wm *WarmupManager) AddStrategy(name string, strategy WarmupStrategy) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.strategies[name] = strategy
}

// RemoveStrategy removes a warmup strategy
func (wm *WarmupManager) RemoveStrategy(name string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	delete(wm.strategies, name)
}

// Error definitions
var (
	ErrCacheMiss = fmt.Errorf("cache miss")
)
