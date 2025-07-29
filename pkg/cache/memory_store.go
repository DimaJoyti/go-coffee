package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MemoryStoreItem represents an item in the memory store
type MemoryStoreItem struct {
	Value      interface{}
	Expiration time.Time
	LastAccess time.Time
	AccessCount uint64
}

// IsExpired checks if the item has expired
func (item *MemoryStoreItem) IsExpired() bool {
	return !item.Expiration.IsZero() && time.Now().After(item.Expiration)
}

// MemoryStore implements an in-memory cache with TTL and LRU eviction
type MemoryStore struct {
	mu          sync.RWMutex
	items       map[string]*MemoryStoreItem
	maxSize     int
	defaultTTL  time.Duration
	cleanupTick *time.Ticker
	stopCleanup chan struct{}
	stats       *MemoryStoreStats
}

// MemoryStoreStats tracks memory store performance
type MemoryStoreStats struct {
	mu           sync.RWMutex
	Hits         int64
	Misses       int64
	Evictions    int64
	Expirations  int64
	CurrentItems int64
}

// MemoryStoreConfig configures the memory store
type MemoryStoreConfig struct {
	MaxSize       int           // Maximum number of items
	DefaultTTL    time.Duration // Default TTL for items
	CleanupPeriod time.Duration // How often to clean up expired items
}

// NewMemoryStore creates a new in-memory store with TTL and LRU eviction
func NewMemoryStore(config MemoryStoreConfig) *MemoryStore {
	if config.MaxSize <= 0 {
		config.MaxSize = 1000
	}
	if config.DefaultTTL <= 0 {
		config.DefaultTTL = time.Hour
	}
	if config.CleanupPeriod <= 0 {
		config.CleanupPeriod = time.Minute * 5
	}

	store := &MemoryStore{
		items:       make(map[string]*MemoryStoreItem),
		maxSize:     config.MaxSize,
		defaultTTL:  config.DefaultTTL,
		cleanupTick: time.NewTicker(config.CleanupPeriod),
		stopCleanup: make(chan struct{}),
		stats:       &MemoryStoreStats{},
	}

	// Start cleanup goroutine
	go store.cleanup()

	return store
}

// Set stores a value with optional TTL
func (ms *MemoryStore) Set(key string, value interface{}, ttl time.Duration) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Use default TTL if not specified
	if ttl == 0 {
		ttl = ms.defaultTTL
	}

	expiration := time.Time{}
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	// Check if we need to evict items
	if len(ms.items) >= ms.maxSize {
		ms.evictLRU()
	}

	ms.items[key] = &MemoryStoreItem{
		Value:      value,
		Expiration: expiration,
		LastAccess: time.Now(),
		AccessCount: 1,
	}

	ms.stats.mu.Lock()
	ms.stats.CurrentItems = int64(len(ms.items))
	ms.stats.mu.Unlock()
}

// Get retrieves a value by key
func (ms *MemoryStore) Get(key string) (interface{}, bool) {
	ms.mu.RLock()
	item, exists := ms.items[key]
	ms.mu.RUnlock()

	if !exists {
		ms.stats.mu.Lock()
		ms.stats.Misses++
		ms.stats.mu.Unlock()
		return nil, false
	}

	// Check if expired
	if item.IsExpired() {
		ms.mu.Lock()
		delete(ms.items, key)
		ms.stats.mu.Lock()
		ms.stats.Expirations++
		ms.stats.Misses++
		ms.stats.CurrentItems = int64(len(ms.items))
		ms.stats.mu.Unlock()
		ms.mu.Unlock()
		return nil, false
	}

	// Update access information
	ms.mu.Lock()
	item.LastAccess = time.Now()
	item.AccessCount++
	ms.mu.Unlock()

	ms.stats.mu.Lock()
	ms.stats.Hits++
	ms.stats.mu.Unlock()

	return item.Value, true
}

// Delete removes a key
func (ms *MemoryStore) Delete(key string) bool {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	_, exists := ms.items[key]
	if exists {
		delete(ms.items, key)
		ms.stats.mu.Lock()
		ms.stats.CurrentItems = int64(len(ms.items))
		ms.stats.mu.Unlock()
	}

	return exists
}

// Clear removes all items
func (ms *MemoryStore) Clear() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.items = make(map[string]*MemoryStoreItem)
	ms.stats.mu.Lock()
	ms.stats.CurrentItems = 0
	ms.stats.mu.Unlock()
}

// Size returns the current number of items
func (ms *MemoryStore) Size() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return len(ms.items)
}

// Keys returns all keys (be careful with large stores)
func (ms *MemoryStore) Keys() []string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	keys := make([]string, 0, len(ms.items))
	for key := range ms.items {
		keys = append(keys, key)
	}
	return keys
}

// GetStats returns current statistics
func (ms *MemoryStore) GetStats() MemoryStoreStats {
	ms.stats.mu.RLock()
	defer ms.stats.mu.RUnlock()
	return *ms.stats
}

// evictLRU removes the least recently used item
func (ms *MemoryStore) evictLRU() {
	var oldestKey string
	var oldestTime time.Time
	var lowestAccess uint64 = ^uint64(0) // Max uint64

	// Find LRU item (combination of last access time and access count)
	for key, item := range ms.items {
		if item.IsExpired() {
			// Prefer expired items for eviction
			delete(ms.items, key)
			ms.stats.Expirations++
			return
		}

		// LRU logic: prioritize by last access time, then by access count
		if oldestTime.IsZero() || 
		   item.LastAccess.Before(oldestTime) ||
		   (item.LastAccess.Equal(oldestTime) && item.AccessCount < lowestAccess) {
			oldestKey = key
			oldestTime = item.LastAccess
			lowestAccess = item.AccessCount
		}
	}

	if oldestKey != "" {
		delete(ms.items, oldestKey)
		ms.stats.Evictions++
	}
}

// cleanup periodically removes expired items
func (ms *MemoryStore) cleanup() {
	for {
		select {
		case <-ms.cleanupTick.C:
			ms.cleanupExpired()
		case <-ms.stopCleanup:
			return
		}
	}
}

// cleanupExpired removes all expired items
func (ms *MemoryStore) cleanupExpired() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	expiredKeys := make([]string, 0)
	for key, item := range ms.items {
		if item.IsExpired() {
			expiredKeys = append(expiredKeys, key)
		}
	}

	expiredCount := len(expiredKeys)
	for _, key := range expiredKeys {
		delete(ms.items, key)
	}

	if expiredCount > 0 {
		ms.stats.mu.Lock()
		ms.stats.Expirations += int64(expiredCount)
		ms.stats.CurrentItems = int64(len(ms.items))
		ms.stats.mu.Unlock()
	}
}

// Close stops the cleanup goroutine
func (ms *MemoryStore) Close() {
	close(ms.stopCleanup)
	ms.cleanupTick.Stop()
}

// HitRate returns the cache hit rate
func (ms *MemoryStore) HitRate() float64 {
	ms.stats.mu.RLock()
	defer ms.stats.mu.RUnlock()

	total := ms.stats.Hits + ms.stats.Misses
	if total == 0 {
		return 0.0
	}
	return float64(ms.stats.Hits) / float64(total)
}

// DefaultMemoryStoreConfig returns default configuration
func DefaultMemoryStoreConfig() MemoryStoreConfig {
	return MemoryStoreConfig{
		MaxSize:       10000,
		DefaultTTL:    time.Hour,
		CleanupPeriod: time.Minute * 5,
	}
}

// MemoryCacheAdapter adapts MemoryStore to Cache interface
type MemoryCacheAdapter struct {
	store *MemoryStore
}

// NewMemoryCacheAdapter creates a new adapter
func NewMemoryCacheAdapter(config MemoryStoreConfig) *MemoryCacheAdapter {
	return &MemoryCacheAdapter{
		store: NewMemoryStore(config),
	}
}

// Set implements Cache interface
func (m *MemoryCacheAdapter) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	m.store.Set(key, value, expiration)
	return nil
}

// Get implements Cache interface
func (m *MemoryCacheAdapter) Get(ctx context.Context, key string, dest interface{}) error {
	value, exists := m.store.Get(key)
	if !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	// Simple type assertion - in a real implementation, you might want more sophisticated type handling
	switch d := dest.(type) {
	case *interface{}:
		*d = value
	case *string:
		if str, ok := value.(string); ok {
			*d = str
		} else {
			return fmt.Errorf("value is not a string")
		}
	default:
		return fmt.Errorf("unsupported destination type")
	}

	return nil
}

// Delete implements Cache interface
func (m *MemoryCacheAdapter) Delete(ctx context.Context, key string) error {
	m.store.Delete(key)
	return nil
}

// Exists implements Cache interface
func (m *MemoryCacheAdapter) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := m.store.Get(key)
	return exists, nil
}

// Expire implements Cache interface (no-op for memory store as TTL is set on creation)
func (m *MemoryCacheAdapter) Expire(ctx context.Context, key string, expiration time.Duration) error {
	// For memory store, we would need to re-set the item with new TTL
	value, exists := m.store.Get(key)
	if !exists {
		return fmt.Errorf("key not found: %s", key)
	}
	m.store.Set(key, value, expiration)
	return nil
}

// Keys implements Cache interface
func (m *MemoryCacheAdapter) Keys(ctx context.Context, pattern string) ([]string, error) {
	// Simple implementation - doesn't support patterns
	return m.store.Keys(), nil
}

// FlushAll implements Cache interface
func (m *MemoryCacheAdapter) FlushAll(ctx context.Context) error {
	m.store.Clear()
	return nil
}

// Health implements Cache interface
func (m *MemoryCacheAdapter) Health(ctx context.Context) error {
	return nil // Memory store is always healthy
}

// MSet implements Cache interface
func (m *MemoryCacheAdapter) MSet(ctx context.Context, pairs map[string]interface{}, expiration time.Duration) error {
	for key, value := range pairs {
		m.store.Set(key, value, expiration)
	}
	return nil
}

// MGet implements Cache interface
func (m *MemoryCacheAdapter) MGet(ctx context.Context, keys []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, key := range keys {
		if value, exists := m.store.Get(key); exists {
			result[key] = value
		}
	}
	return result, nil
}

// MDelete implements Cache interface
func (m *MemoryCacheAdapter) MDelete(ctx context.Context, keys []string) error {
	for _, key := range keys {
		m.store.Delete(key)
	}
	return nil
}

// Pipeline implements Cache interface (returns nil as memory store doesn't need pipelines)
func (m *MemoryCacheAdapter) Pipeline() *RedisPipeline {
	return nil
}

// Close closes the memory store
func (m *MemoryCacheAdapter) Close() error {
	m.store.Close()
	return nil
}