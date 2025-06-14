package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/config"
	"go.uber.org/zap"
)

// Manager provides a unified interface for caching operations
type Manager struct {
	advancedCache *AdvancedRedisCache
	config        *config.RedisConfig
	logger        *zap.Logger
}

// NewManager creates a new cache manager with advanced Redis features
func NewManager(cfg *config.RedisConfig, logger *zap.Logger) (*Manager, error) {
	// Convert existing config to advanced config
	advancedConfig := &AdvancedCacheConfig{
		Addrs:           []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		DialTimeout:     cfg.DialTimeout,
		MaxRetries:      cfg.MaxRetries,
		RetryDelay:      cfg.RetryDelay,
		DefaultTTL:      1 * time.Hour, // Default TTL since it's not in config
		MaxKeySize:      1024,          // 1KB max key size
		MaxValueSize:    10485760,      // 10MB max value size
		CompressionMin:  1024,          // Compress values larger than 1KB
		WarmupEnabled:   true,
		WarmupInterval:  5 * time.Minute,
		WarmupBatchSize: 100,
	}

	advancedCache, err := NewAdvancedRedisCache(advancedConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create advanced Redis cache: %w", err)
	}

	manager := &Manager{
		advancedCache: advancedCache,
		config:        cfg,
		logger:        logger,
	}

	// Set up cache warming strategies
	manager.setupWarmupStrategies()

	return manager, nil
}

// Set stores a value in cache
func (m *Manager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return m.advancedCache.Set(ctx, key, value, ttl)
}

// Get retrieves a value from cache
func (m *Manager) Get(ctx context.Context, key string, dest interface{}) error {
	return m.advancedCache.Get(ctx, key, dest)
}

// Delete removes a key from cache
func (m *Manager) Delete(ctx context.Context, key string) error {
	return m.advancedCache.Delete(ctx, key)
}

// MGet retrieves multiple values efficiently
func (m *Manager) MGet(ctx context.Context, keys []string) (map[string]interface{}, error) {
	return m.advancedCache.MGet(ctx, keys)
}

// MSet stores multiple values efficiently
func (m *Manager) MSet(ctx context.Context, pairs map[string]interface{}, ttl time.Duration) error {
	return m.advancedCache.MSet(ctx, pairs, ttl)
}

// Exists checks if keys exist in cache
func (m *Manager) Exists(ctx context.Context, keys ...string) (int64, error) {
	return m.advancedCache.Exists(ctx, keys...)
}

// TTL returns the time to live for a key
func (m *Manager) TTL(ctx context.Context, key string) (time.Duration, error) {
	return m.advancedCache.TTL(ctx, key)
}

// Expire sets a timeout on a key
func (m *Manager) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return m.advancedCache.Expire(ctx, key, ttl)
}

// FlushPattern removes all keys matching a pattern
func (m *Manager) FlushPattern(ctx context.Context, pattern string) error {
	return m.advancedCache.FlushPattern(ctx, pattern)
}

// GetMetrics returns current cache metrics
func (m *Manager) GetMetrics() CacheMetrics {
	return m.advancedCache.GetMetrics()
}

// Close closes the cache connection
func (m *Manager) Close() error {
	return m.advancedCache.Close()
}

// setupWarmupStrategies configures cache warming strategies
func (m *Manager) setupWarmupStrategies() {
	if m.advancedCache.warmupCache == nil {
		return
	}

	// Menu warmup strategy
	menuStrategy := &MenuWarmupStrategy{
		cache:  m.advancedCache,
		logger: m.logger,
	}
	m.advancedCache.warmupCache.AddStrategy("menu", menuStrategy)

	// Popular items warmup strategy
	popularItemsStrategy := &PopularItemsWarmupStrategy{
		cache:  m.advancedCache,
		logger: m.logger,
	}
	m.advancedCache.warmupCache.AddStrategy("popular_items", popularItemsStrategy)

	// User sessions warmup strategy
	userSessionsStrategy := &UserSessionsWarmupStrategy{
		cache:  m.advancedCache,
		logger: m.logger,
	}
	m.advancedCache.warmupCache.AddStrategy("user_sessions", userSessionsStrategy)
}

// MenuWarmupStrategy warms up menu data
type MenuWarmupStrategy struct {
	cache  *AdvancedRedisCache
	logger *zap.Logger
}

func (mws *MenuWarmupStrategy) Warmup(ctx context.Context) error {
	mws.logger.Debug("Starting menu warmup")

	// Simulate warming up menu data
	menuKeys := []string{
		"menu:coffee:espresso",
		"menu:coffee:latte",
		"menu:coffee:cappuccino",
		"menu:food:sandwich",
		"menu:food:pastry",
	}

	for _, key := range menuKeys {
		// Check if key exists, if not, warm it up
		exists, err := mws.cache.Exists(ctx, key)
		if err != nil {
			mws.logger.Error("Failed to check key existence", zap.String("key", key), zap.Error(err))
			continue
		}

		if exists == 0 {
			// Simulate fetching from database and caching
			menuData := map[string]interface{}{
				"name":        key,
				"price":       4.50,
				"available":   true,
				"description": "Delicious coffee item",
			}

			if err := mws.cache.Set(ctx, key, menuData, 1*time.Hour); err != nil {
				mws.logger.Error("Failed to warm up menu key", zap.String("key", key), zap.Error(err))
			} else {
				mws.logger.Debug("Warmed up menu key", zap.String("key", key))
			}
		}
	}

	return nil
}

func (mws *MenuWarmupStrategy) GetKeys() []string {
	return []string{
		"menu:coffee:*",
		"menu:food:*",
	}
}

func (mws *MenuWarmupStrategy) GetTTL() time.Duration {
	return 1 * time.Hour
}

// PopularItemsWarmupStrategy warms up popular items data
type PopularItemsWarmupStrategy struct {
	cache  *AdvancedRedisCache
	logger *zap.Logger
}

func (piws *PopularItemsWarmupStrategy) Warmup(ctx context.Context) error {
	piws.logger.Debug("Starting popular items warmup")

	popularKeys := []string{
		"popular:daily",
		"popular:weekly",
		"popular:monthly",
	}

	for _, key := range popularKeys {
		exists, err := piws.cache.Exists(ctx, key)
		if err != nil {
			continue
		}

		if exists == 0 {
			// Simulate popular items data
			popularData := map[string]interface{}{
				"items":      []string{"espresso", "latte", "cappuccino"},
				"updated_at": time.Now(),
			}

			if err := piws.cache.Set(ctx, key, popularData, 30*time.Minute); err != nil {
				piws.logger.Error("Failed to warm up popular items key", zap.String("key", key), zap.Error(err))
			}
		}
	}

	return nil
}

func (piws *PopularItemsWarmupStrategy) GetKeys() []string {
	return []string{
		"popular:*",
	}
}

func (piws *PopularItemsWarmupStrategy) GetTTL() time.Duration {
	return 30 * time.Minute
}

// UserSessionsWarmupStrategy warms up user session data
type UserSessionsWarmupStrategy struct {
	cache  *AdvancedRedisCache
	logger *zap.Logger
}

func (usws *UserSessionsWarmupStrategy) Warmup(ctx context.Context) error {
	usws.logger.Debug("Starting user sessions warmup")

	// This would typically fetch active user sessions from database
	// For demo purposes, we'll just ensure session infrastructure is ready
	sessionKeys := []string{
		"sessions:config",
		"sessions:cleanup_schedule",
	}

	for _, key := range sessionKeys {
		exists, err := usws.cache.Exists(ctx, key)
		if err != nil {
			continue
		}

		if exists == 0 {
			configData := map[string]interface{}{
				"timeout":          30 * time.Minute,
				"max_idle":         15 * time.Minute,
				"cleanup_interval": 5 * time.Minute,
			}

			if err := usws.cache.Set(ctx, key, configData, 1*time.Hour); err != nil {
				usws.logger.Error("Failed to warm up session key", zap.String("key", key), zap.Error(err))
			}
		}
	}

	return nil
}

func (usws *UserSessionsWarmupStrategy) GetKeys() []string {
	return []string{
		"sessions:*",
	}
}

func (usws *UserSessionsWarmupStrategy) GetTTL() time.Duration {
	return 1 * time.Hour
}

// CacheHelper provides utility functions for common caching patterns
type CacheHelper struct {
	manager *Manager
}

// NewCacheHelper creates a new cache helper
func (m *Manager) NewCacheHelper() *CacheHelper {
	return &CacheHelper{manager: m}
}

// GetOrSet gets a value from cache, or sets it if it doesn't exist
func (ch *CacheHelper) GetOrSet(ctx context.Context, key string, dest interface{}, ttl time.Duration, fetchFn func() (interface{}, error)) error {
	// Try to get from cache first
	err := ch.manager.Get(ctx, key, dest)
	if err == nil {
		return nil // Found in cache
	}

	if err != ErrCacheMiss {
		return fmt.Errorf("cache get error: %w", err)
	}

	// Not in cache, fetch the data
	value, err := fetchFn()
	if err != nil {
		return fmt.Errorf("fetch function error: %w", err)
	}

	// Store in cache
	if err := ch.manager.Set(ctx, key, value, ttl); err != nil {
		// Log error but don't fail the request
		ch.manager.logger.Error("Failed to set cache", zap.String("key", key), zap.Error(err))
	}

	// Copy value to destination
	return ch.manager.Get(ctx, key, dest)
}

// InvalidatePattern invalidates all keys matching a pattern
func (ch *CacheHelper) InvalidatePattern(ctx context.Context, pattern string) error {
	return ch.manager.FlushPattern(ctx, pattern)
}

// RefreshKey refreshes a key by deleting it (forcing a reload on next access)
func (ch *CacheHelper) RefreshKey(ctx context.Context, key string) error {
	return ch.manager.Delete(ctx, key)
}
