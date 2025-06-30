package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go-coffee-ai-agents/internal/ai/providers"
)

// Cache provides intelligent caching for AI responses
type Cache struct {
	storage  Storage
	config   *Config
	stats    *Stats
	mutex    sync.RWMutex
	
	// Cache strategies
	strategies map[string]Strategy
}

// Storage interface for cache backends
type Storage interface {
	Get(ctx context.Context, key string) (*CacheEntry, error)
	Set(ctx context.Context, key string, entry *CacheEntry, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	Keys(ctx context.Context, pattern string) ([]string, error)
	Size(ctx context.Context) (int64, error)
}

// Strategy interface for cache strategies
type Strategy interface {
	ShouldCache(req *providers.GenerateRequest, resp *providers.GenerateResponse) bool
	GetTTL(req *providers.GenerateRequest, resp *providers.GenerateResponse) time.Duration
	GetKey(req *providers.GenerateRequest) string
}

// Config holds cache configuration
type Config struct {
	Enabled           bool          `yaml:"enabled" json:"enabled"`
	DefaultTTL        time.Duration `yaml:"default_ttl" json:"default_ttl"`
	MaxSize           int64         `yaml:"max_size" json:"max_size"`
	MaxEntrySize      int64         `yaml:"max_entry_size" json:"max_entry_size"`
	
	// Cache strategies
	Strategies        map[string]*StrategyConfig `yaml:"strategies" json:"strategies"`
	DefaultStrategy   string        `yaml:"default_strategy" json:"default_strategy"`
	
	// Cleanup settings
	CleanupInterval   time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
	MaxAge            time.Duration `yaml:"max_age" json:"max_age"`
	
	// Performance settings
	CompressionEnabled bool         `yaml:"compression_enabled" json:"compression_enabled"`
	CompressionLevel   int          `yaml:"compression_level" json:"compression_level"`
	
	// Security settings
	EncryptionEnabled  bool         `yaml:"encryption_enabled" json:"encryption_enabled"`
	EncryptionKey      string       `yaml:"encryption_key" json:"encryption_key"`
}

// StrategyConfig holds configuration for a cache strategy
type StrategyConfig struct {
	Type              string        `yaml:"type" json:"type"`
	TTL               time.Duration `yaml:"ttl" json:"ttl"`
	MaxTokens         int           `yaml:"max_tokens" json:"max_tokens"`
	MinTokens         int           `yaml:"min_tokens" json:"min_tokens"`
	CacheErrors       bool          `yaml:"cache_errors" json:"cache_errors"`
	CachePartial      bool          `yaml:"cache_partial" json:"cache_partial"`
	
	// Content-based settings
	CacheByContent    bool          `yaml:"cache_by_content" json:"cache_by_content"`
	ContentPatterns   []string      `yaml:"content_patterns" json:"content_patterns"`
	
	// Model-specific settings
	Models            []string      `yaml:"models" json:"models"`
	Providers         []string      `yaml:"providers" json:"providers"`
	
	// User-specific settings
	UserSpecific      bool          `yaml:"user_specific" json:"user_specific"`
	
	// Cost optimization
	MinCostToCache    float64       `yaml:"min_cost_to_cache" json:"min_cost_to_cache"`
}

// CacheEntry represents a cached response
type CacheEntry struct {
	Key           string                      `json:"key"`
	Request       *providers.GenerateRequest  `json:"request"`
	Response      *providers.GenerateResponse `json:"response"`
	
	// Metadata
	CreatedAt     time.Time                   `json:"created_at"`
	ExpiresAt     time.Time                   `json:"expires_at"`
	AccessCount   int64                       `json:"access_count"`
	LastAccessed  time.Time                   `json:"last_accessed"`
	
	// Cache metadata
	Strategy      string                      `json:"strategy"`
	Size          int64                       `json:"size"`
	Compressed    bool                        `json:"compressed"`
	Encrypted     bool                        `json:"encrypted"`
	
	// Cost savings
	OriginalCost  float64                     `json:"original_cost"`
	SavedCost     float64                     `json:"saved_cost"`
	
	// Quality metrics
	Confidence    float64                     `json:"confidence,omitempty"`
	Relevance     float64                     `json:"relevance,omitempty"`
}

// Stats holds cache statistics
type Stats struct {
	Hits              int64     `json:"hits"`
	Misses            int64     `json:"misses"`
	Sets              int64     `json:"sets"`
	Deletes           int64     `json:"deletes"`
	Evictions         int64     `json:"evictions"`
	
	// Size statistics
	CurrentSize       int64     `json:"current_size"`
	MaxSize           int64     `json:"max_size"`
	EntryCount        int64     `json:"entry_count"`
	
	// Performance statistics
	AverageHitTime    time.Duration `json:"average_hit_time"`
	AverageMissTime   time.Duration `json:"average_miss_time"`
	AverageSetTime    time.Duration `json:"average_set_time"`
	
	// Cost savings
	TotalSavedCost    float64   `json:"total_saved_cost"`
	TotalOriginalCost float64   `json:"total_original_cost"`
	
	// Time statistics
	LastHit           time.Time `json:"last_hit"`
	LastMiss          time.Time `json:"last_miss"`
	LastSet           time.Time `json:"last_set"`
	
	// Hit rate
	HitRate           float64   `json:"hit_rate"`
}

// NewCache creates a new cache instance
func NewCache(storage Storage, config *Config) *Cache {
	cache := &Cache{
		storage:    storage,
		config:     config,
		stats:      &Stats{MaxSize: config.MaxSize},
		strategies: make(map[string]Strategy),
	}
	
	// Initialize strategies
	cache.initializeStrategies()
	
	// Start cleanup goroutine
	if config.CleanupInterval > 0 {
		go cache.cleanup()
	}
	
	return cache
}

// Get retrieves a cached response
func (c *Cache) Get(ctx context.Context, req *providers.GenerateRequest) (*providers.GenerateResponse, bool) {
	if !c.config.Enabled {
		return nil, false
	}
	
	start := time.Now()
	defer func() {
		c.updateHitTime(time.Since(start))
	}()
	
	// Generate cache key
	key := c.generateKey(req)
	
	// Get from storage
	entry, err := c.storage.Get(ctx, key)
	if err != nil || entry == nil {
		c.recordMiss()
		return nil, false
	}
	
	// Check expiration
	if time.Now().After(entry.ExpiresAt) {
		c.recordMiss()
		// Async delete expired entry
		go c.storage.Delete(context.Background(), key)
		return nil, false
	}
	
	// Update access statistics
	entry.AccessCount++
	entry.LastAccessed = time.Now()
	
	// Update cache entry (async)
	go c.storage.Set(context.Background(), key, entry, time.Until(entry.ExpiresAt))
	
	// Record hit
	c.recordHit(entry)
	
	// Mark response as from cache
	response := entry.Response
	response.FromCache = true
	response.CacheKey = key
	
	return response, true
}

// Set stores a response in the cache
func (c *Cache) Set(ctx context.Context, req *providers.GenerateRequest, resp *providers.GenerateResponse) error {
	if !c.config.Enabled {
		return nil
	}
	
	start := time.Now()
	defer func() {
		c.updateSetTime(time.Since(start))
	}()
	
	// Check if response should be cached
	strategy := c.getStrategy(req)
	if !strategy.ShouldCache(req, resp) {
		return nil
	}
	
	// Generate cache key
	key := c.generateKey(req)
	
	// Get TTL from strategy
	ttl := strategy.GetTTL(req, resp)
	if ttl <= 0 {
		ttl = c.config.DefaultTTL
	}
	
	// Create cache entry
	entry := &CacheEntry{
		Key:          key,
		Request:      req,
		Response:     resp,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(ttl),
		AccessCount:  0,
		LastAccessed: time.Now(),
		Strategy:     c.config.DefaultStrategy,
	}
	
	// Calculate size
	entryData, _ := json.Marshal(entry)
	entry.Size = int64(len(entryData))
	
	// Check size limits
	if c.config.MaxEntrySize > 0 && entry.Size > c.config.MaxEntrySize {
		return fmt.Errorf("cache entry too large: %d bytes", entry.Size)
	}
	
	// Add cost information
	if resp.Cost != nil {
		entry.OriginalCost = resp.Cost.TotalCost
	}
	
	// Store in cache
	err := c.storage.Set(ctx, key, entry, ttl)
	if err != nil {
		return fmt.Errorf("failed to store cache entry: %w", err)
	}
	
	// Record set
	c.recordSet(entry)
	
	return nil
}

// Delete removes an entry from the cache
func (c *Cache) Delete(ctx context.Context, req *providers.GenerateRequest) error {
	if !c.config.Enabled {
		return nil
	}
	
	key := c.generateKey(req)
	err := c.storage.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete cache entry: %w", err)
	}
	
	c.recordDelete()
	return nil
}

// Clear clears all cache entries
func (c *Cache) Clear(ctx context.Context) error {
	if !c.config.Enabled {
		return nil
	}
	
	err := c.storage.Clear(ctx)
	if err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	
	// Reset statistics
	c.mutex.Lock()
	c.stats = &Stats{MaxSize: c.config.MaxSize}
	c.mutex.Unlock()
	
	return nil
}

// GetStats returns cache statistics
func (c *Cache) GetStats() *Stats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	// Calculate hit rate
	total := c.stats.Hits + c.stats.Misses
	if total > 0 {
		c.stats.HitRate = float64(c.stats.Hits) / float64(total)
	}
	
	// Create a copy to prevent external modifications
	stats := *c.stats
	return &stats
}

// generateKey generates a cache key for a request
func (c *Cache) generateKey(req *providers.GenerateRequest) string {
	strategy := c.getStrategy(req)
	return strategy.GetKey(req)
}

// getStrategy returns the appropriate caching strategy for a request
func (c *Cache) getStrategy(req *providers.GenerateRequest) Strategy {
	// Try to find a specific strategy for the request
	for name, strategy := range c.strategies {
		if c.strategyMatches(name, req) {
			return strategy
		}
	}
	
	// Return default strategy
	if defaultStrategy, exists := c.strategies[c.config.DefaultStrategy]; exists {
		return defaultStrategy
	}
	
	// Fallback to basic strategy
	return &BasicStrategy{config: c.config}
}

// strategyMatches checks if a strategy matches a request
func (c *Cache) strategyMatches(strategyName string, req *providers.GenerateRequest) bool {
	strategyConfig, exists := c.config.Strategies[strategyName]
	if !exists {
		return false
	}
	
	// Check provider match
	if len(strategyConfig.Providers) > 0 {
		found := false
		for _, provider := range strategyConfig.Providers {
			if provider == req.Provider {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	// Check model match
	if len(strategyConfig.Models) > 0 {
		found := false
		for _, model := range strategyConfig.Models {
			if model == req.Model {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	return true
}

// initializeStrategies initializes cache strategies
func (c *Cache) initializeStrategies() {
	// Create strategies based on configuration
	for name, strategyConfig := range c.config.Strategies {
		switch strategyConfig.Type {
		case "basic":
			c.strategies[name] = &BasicStrategy{
				config:         c.config,
				strategyConfig: strategyConfig,
			}
		case "content_based":
			c.strategies[name] = &ContentBasedStrategy{
				config:         c.config,
				strategyConfig: strategyConfig,
			}
		case "cost_optimized":
			c.strategies[name] = &CostOptimizedStrategy{
				config:         c.config,
				strategyConfig: strategyConfig,
			}
		default:
			// Default to basic strategy
			c.strategies[name] = &BasicStrategy{
				config:         c.config,
				strategyConfig: strategyConfig,
			}
		}
	}
	
	// Ensure default strategy exists
	if _, exists := c.strategies[c.config.DefaultStrategy]; !exists {
		c.strategies[c.config.DefaultStrategy] = &BasicStrategy{config: c.config}
	}
}

// cleanup periodically cleans up expired entries
func (c *Cache) cleanup() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		ctx := context.Background()
		
		// Get all keys
		keys, err := c.storage.Keys(ctx, "*")
		if err != nil {
			continue
		}
		
		// Check each entry for expiration
		for _, key := range keys {
			entry, err := c.storage.Get(ctx, key)
			if err != nil || entry == nil {
				continue
			}
			
			// Delete expired entries
			if time.Now().After(entry.ExpiresAt) {
				c.storage.Delete(ctx, key)
				c.recordEviction()
			}
			
			// Delete old entries
			if c.config.MaxAge > 0 && time.Since(entry.CreatedAt) > c.config.MaxAge {
				c.storage.Delete(ctx, key)
				c.recordEviction()
			}
		}
	}
}

// Record statistics methods
func (c *Cache) recordHit(entry *CacheEntry) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.stats.Hits++
	c.stats.LastHit = time.Now()
	
	// Add to saved cost
	if entry.OriginalCost > 0 {
		c.stats.TotalSavedCost += entry.OriginalCost
		entry.SavedCost += entry.OriginalCost
	}
}

func (c *Cache) recordMiss() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.stats.Misses++
	c.stats.LastMiss = time.Now()
}

func (c *Cache) recordSet(entry *CacheEntry) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.stats.Sets++
	c.stats.LastSet = time.Now()
	c.stats.CurrentSize += entry.Size
	c.stats.EntryCount++
	
	if entry.OriginalCost > 0 {
		c.stats.TotalOriginalCost += entry.OriginalCost
	}
}

func (c *Cache) recordDelete() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.stats.Deletes++
}

func (c *Cache) recordEviction() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.stats.Evictions++
}

func (c *Cache) updateHitTime(duration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	if c.stats.AverageHitTime == 0 {
		c.stats.AverageHitTime = duration
	} else {
		c.stats.AverageHitTime = (c.stats.AverageHitTime + duration) / 2
	}
}

func (c *Cache) updateSetTime(duration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	if c.stats.AverageSetTime == 0 {
		c.stats.AverageSetTime = duration
	} else {
		c.stats.AverageSetTime = (c.stats.AverageSetTime + duration) / 2
	}
}

// GenerateStandardKey generates a standard cache key for a request
func GenerateStandardKey(req *providers.GenerateRequest) string {
	// Create a hash of the request parameters that affect the response
	hasher := sha256.New()
	
	// Include prompt/messages
	if req.Prompt != "" {
		hasher.Write([]byte(req.Prompt))
	}
	
	if len(req.Messages) > 0 {
		for _, msg := range req.Messages {
			hasher.Write([]byte(msg.Role))
			hasher.Write([]byte(msg.Content))
		}
	}
	
	// Include model and provider
	hasher.Write([]byte(req.Model))
	hasher.Write([]byte(req.Provider))
	
	// Include generation parameters
	hasher.Write([]byte(fmt.Sprintf("%.2f", req.Temperature)))
	hasher.Write([]byte(fmt.Sprintf("%.2f", req.TopP)))
	hasher.Write([]byte(fmt.Sprintf("%d", req.MaxTokens)))
	
	// Include stop sequences
	for _, stop := range req.StopSequences {
		hasher.Write([]byte(stop))
	}
	
	// Generate hex hash
	hash := hex.EncodeToString(hasher.Sum(nil))
	
	// Return cache key with prefix
	return fmt.Sprintf("ai_cache:%s:%s:%s", req.Provider, req.Model, hash[:16])
}
