package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/bright-data-hub/config"
	"github.com/go-redis/redis/v8"
)

// AdvancedCache provides multi-level caching with Redis and in-memory fallback
type AdvancedCache struct {
	config      *config.BrightDataHubConfig
	redisClient *redis.Client
	localCache  *LocalCache
	enabled     bool
}

// LocalCache represents an in-memory cache with TTL support
type LocalCache struct {
	data    map[string]*CacheEntry
	maxSize int
	mu      sync.RWMutex
}

// CacheEntry represents a cached item with metadata
type CacheEntry struct {
	Value     interface{} `json:"value"`
	ExpiresAt time.Time   `json:"expires_at"`
	CreatedAt time.Time   `json:"created_at"`
	AccessCount int64     `json:"access_count"`
	LastAccess  time.Time `json:"last_access"`
}

// CacheStats provides cache performance metrics
type CacheStats struct {
	RedisHits      int64 `json:"redis_hits"`
	RedisMisses    int64 `json:"redis_misses"`
	LocalHits      int64 `json:"local_hits"`
	LocalMisses    int64 `json:"local_misses"`
	TotalRequests  int64 `json:"total_requests"`
	HitRatio       float64 `json:"hit_ratio"`
	LocalSize      int   `json:"local_size"`
	RedisConnected bool  `json:"redis_connected"`
}

// NewAdvancedCache creates a new advanced cache instance
func NewAdvancedCache(cfg *config.BrightDataHubConfig) (*AdvancedCache, error) {
	cache := &AdvancedCache{
		config:  cfg,
		enabled: true,
		localCache: &LocalCache{
			data:    make(map[string]*CacheEntry),
			maxSize: cfg.CacheMaxSize,
		},
	}
	
	// Initialize Redis client if URL is provided
	if cfg.RedisURL != "" {
		opts, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
		}
		
		cache.redisClient = redis.NewClient(opts)
		
		// Test Redis connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := cache.redisClient.Ping(ctx).Err(); err != nil {
			// Redis is not available, continue with local cache only
			cache.redisClient = nil
			fmt.Printf("Warning: Redis not available, using local cache only: %v\n", err)
		}
	}
	
	// Start cleanup goroutine for local cache
	go cache.startCleanupRoutine()
	
	return cache, nil
}

// Get retrieves a value from cache (Redis first, then local)
func (c *AdvancedCache) Get(key string) interface{} {
	if !c.enabled {
		return nil
	}
	
	// Try Redis first
	if c.redisClient != nil {
		if value := c.getFromRedis(key); value != nil {
			// Also store in local cache for faster access
			c.setInLocal(key, value, c.config.CacheTTL)
			return value
		}
	}
	
	// Try local cache
	return c.getFromLocal(key)
}

// Set stores a value in cache (both Redis and local)
func (c *AdvancedCache) Set(key string, value interface{}, ttl time.Duration) {
	if !c.enabled {
		return
	}
	
	// Store in Redis
	if c.redisClient != nil {
		c.setInRedis(key, value, ttl)
	}
	
	// Store in local cache
	c.setInLocal(key, value, ttl)
}

// Delete removes a value from cache
func (c *AdvancedCache) Delete(key string) {
	if !c.enabled {
		return
	}
	
	// Delete from Redis
	if c.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		c.redisClient.Del(ctx, key)
	}
	
	// Delete from local cache
	c.localCache.mu.Lock()
	delete(c.localCache.data, key)
	c.localCache.mu.Unlock()
}

// Clear removes all cached values
func (c *AdvancedCache) Clear() {
	if !c.enabled {
		return
	}
	
	// Clear Redis
	if c.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		c.redisClient.FlushDB(ctx)
	}
	
	// Clear local cache
	c.localCache.mu.Lock()
	c.localCache.data = make(map[string]*CacheEntry)
	c.localCache.mu.Unlock()
}

// GetStats returns cache performance statistics
func (c *AdvancedCache) GetStats() *CacheStats {
	stats := &CacheStats{
		LocalSize:      len(c.localCache.data),
		RedisConnected: c.redisClient != nil,
	}
	
	// Calculate hit ratio
	totalRequests := stats.RedisHits + stats.RedisMisses + stats.LocalHits + stats.LocalMisses
	if totalRequests > 0 {
		hits := stats.RedisHits + stats.LocalHits
		stats.HitRatio = float64(hits) / float64(totalRequests)
	}
	
	stats.TotalRequests = totalRequests
	
	return stats
}

// Redis operations
func (c *AdvancedCache) getFromRedis(key string) interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	data, err := c.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err != redis.Nil {
			fmt.Printf("Redis get error: %v\n", err)
		}
		return nil
	}
	
	var entry CacheEntry
	if err := json.Unmarshal([]byte(data), &entry); err != nil {
		fmt.Printf("Redis unmarshal error: %v\n", err)
		return nil
	}
	
	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		// Delete expired entry
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			c.redisClient.Del(ctx, key)
		}()
		return nil
	}
	
	// Update access statistics
	entry.AccessCount++
	entry.LastAccess = time.Now()
	
	return entry.Value
}

func (c *AdvancedCache) setInRedis(key string, value interface{}, ttl time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	entry := &CacheEntry{
		Value:       value,
		ExpiresAt:   time.Now().Add(ttl),
		CreatedAt:   time.Now(),
		AccessCount: 0,
		LastAccess:  time.Now(),
	}
	
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("Redis marshal error: %v\n", err)
		return
	}
	
	if err := c.redisClient.Set(ctx, key, data, ttl).Err(); err != nil {
		fmt.Printf("Redis set error: %v\n", err)
	}
}

// Local cache operations
func (c *AdvancedCache) getFromLocal(key string) interface{} {
	c.localCache.mu.RLock()
	entry, exists := c.localCache.data[key]
	c.localCache.mu.RUnlock()
	
	if !exists {
		return nil
	}
	
	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		// Delete expired entry
		c.localCache.mu.Lock()
		delete(c.localCache.data, key)
		c.localCache.mu.Unlock()
		return nil
	}
	
	// Update access statistics
	c.localCache.mu.Lock()
	entry.AccessCount++
	entry.LastAccess = time.Now()
	c.localCache.mu.Unlock()
	
	return entry.Value
}

func (c *AdvancedCache) setInLocal(key string, value interface{}, ttl time.Duration) {
	c.localCache.mu.Lock()
	defer c.localCache.mu.Unlock()
	
	// Check if cache is full and evict if necessary
	if len(c.localCache.data) >= c.localCache.maxSize {
		c.evictLRU()
	}
	
	entry := &CacheEntry{
		Value:       value,
		ExpiresAt:   time.Now().Add(ttl),
		CreatedAt:   time.Now(),
		AccessCount: 0,
		LastAccess:  time.Now(),
	}
	
	c.localCache.data[key] = entry
}

// evictLRU removes the least recently used item from local cache
func (c *AdvancedCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time = time.Now()
	
	for key, entry := range c.localCache.data {
		if entry.LastAccess.Before(oldestTime) {
			oldestTime = entry.LastAccess
			oldestKey = key
		}
	}
	
	if oldestKey != "" {
		delete(c.localCache.data, oldestKey)
	}
}

// startCleanupRoutine starts a background routine to clean up expired entries
func (c *AdvancedCache) startCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		c.cleanupExpired()
	}
}

// cleanupExpired removes expired entries from local cache
func (c *AdvancedCache) cleanupExpired() {
	c.localCache.mu.Lock()
	defer c.localCache.mu.Unlock()
	
	now := time.Now()
	for key, entry := range c.localCache.data {
		if now.After(entry.ExpiresAt) {
			delete(c.localCache.data, key)
		}
	}
}

// GetOrFetch retrieves from cache or fetches using the provided function
func (c *AdvancedCache) GetOrFetch(key string, fetcher func() (interface{}, error), ttl time.Duration) (interface{}, error) {
	// Try to get from cache first
	if value := c.Get(key); value != nil {
		return value, nil
	}
	
	// Fetch the value
	value, err := fetcher()
	if err != nil {
		return nil, err
	}
	
	// Store in cache
	c.Set(key, value, ttl)
	
	return value, nil
}

// Close gracefully shuts down the cache
func (c *AdvancedCache) Close() error {
	if c.redisClient != nil {
		return c.redisClient.Close()
	}
	return nil
}
