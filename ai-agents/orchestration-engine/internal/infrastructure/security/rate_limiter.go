package security

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	requests map[string]*RequestTracker
	limit    int
	window   time.Duration
	mutex    sync.RWMutex
	logger   Logger
}

// RequestTracker tracks requests for a specific identifier
type RequestTracker struct {
	Count     int       `json:"count"`
	Window    time.Time `json:"window"`
	LastSeen  time.Time `json:"last_seen"`
	Blocked   bool      `json:"blocked"`
	BlockedAt time.Time `json:"blocked_at"`
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	RequestsPerWindow int           `json:"requests_per_window"`
	WindowDuration    time.Duration `json:"window_duration"`
	BlockDuration     time.Duration `json:"block_duration"`
	CleanupInterval   time.Duration `json:"cleanup_interval"`
	MaxTrackers       int           `json:"max_trackers"`
}

// RateLimitResult represents the result of a rate limit check
type RateLimitResult struct {
	Allowed       bool          `json:"allowed"`
	Remaining     int           `json:"remaining"`
	ResetTime     time.Time     `json:"reset_time"`
	RetryAfter    time.Duration `json:"retry_after"`
	Blocked       bool          `json:"blocked"`
	BlockedUntil  time.Time     `json:"blocked_until"`
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string]*RequestTracker),
		limit:    limit,
		window:   window,
	}
}

// NewAdvancedRateLimiter creates a new rate limiter with advanced configuration
func NewAdvancedRateLimiter(config *RateLimitConfig, logger Logger) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*RequestTracker),
		limit:    config.RequestsPerWindow,
		window:   config.WindowDuration,
		logger:   logger,
	}

	// Start cleanup routine
	go rl.cleanupRoutine(config.CleanupInterval, config.MaxTrackers)

	return rl
}

// Allow checks if a request is allowed for the given identifier
func (rl *RateLimiter) Allow(identifier string) bool {
	result := rl.Check(identifier)
	return result.Allowed
}

// Check performs a comprehensive rate limit check
func (rl *RateLimiter) Check(identifier string) *RateLimitResult {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	tracker, exists := rl.requests[identifier]

	if !exists {
		// First request for this identifier
		tracker = &RequestTracker{
			Count:    1,
			Window:   now.Add(rl.window),
			LastSeen: now,
		}
		rl.requests[identifier] = tracker

		return &RateLimitResult{
			Allowed:   true,
			Remaining: rl.limit - 1,
			ResetTime: tracker.Window,
		}
	}

	// Update last seen
	tracker.LastSeen = now

	// Check if currently blocked
	if tracker.Blocked && now.Before(tracker.BlockedAt.Add(time.Hour)) {
		return &RateLimitResult{
			Allowed:      false,
			Remaining:    0,
			ResetTime:    tracker.Window,
			Blocked:      true,
			BlockedUntil: tracker.BlockedAt.Add(time.Hour),
			RetryAfter:   tracker.BlockedAt.Add(time.Hour).Sub(now),
		}
	}

	// Reset window if expired
	if now.After(tracker.Window) {
		tracker.Count = 1
		tracker.Window = now.Add(rl.window)
		tracker.Blocked = false

		return &RateLimitResult{
			Allowed:   true,
			Remaining: rl.limit - 1,
			ResetTime: tracker.Window,
		}
	}

	// Check if limit exceeded
	if tracker.Count >= rl.limit {
		tracker.Blocked = true
		tracker.BlockedAt = now

		if rl.logger != nil {
			rl.logger.Warn("Rate limit exceeded", "identifier", identifier, "count", tracker.Count, "limit", rl.limit)
		}

		return &RateLimitResult{
			Allowed:      false,
			Remaining:    0,
			ResetTime:    tracker.Window,
			Blocked:      true,
			BlockedUntil: now.Add(time.Hour),
			RetryAfter:   time.Hour,
		}
	}

	// Increment count
	tracker.Count++

	return &RateLimitResult{
		Allowed:   true,
		Remaining: rl.limit - tracker.Count,
		ResetTime: tracker.Window,
	}
}

// Reset resets the rate limit for a specific identifier
func (rl *RateLimiter) Reset(identifier string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	delete(rl.requests, identifier)

	if rl.logger != nil {
		rl.logger.Debug("Rate limit reset", "identifier", identifier)
	}
}

// Block manually blocks an identifier for a specified duration
func (rl *RateLimiter) Block(identifier string, duration time.Duration) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	tracker, exists := rl.requests[identifier]

	if !exists {
		tracker = &RequestTracker{
			Count:    rl.limit,
			Window:   now.Add(rl.window),
			LastSeen: now,
		}
		rl.requests[identifier] = tracker
	}

	tracker.Blocked = true
	tracker.BlockedAt = now

	if rl.logger != nil {
		rl.logger.Info("Identifier manually blocked", "identifier", identifier, "duration", duration)
	}
}

// Unblock removes a block for a specific identifier
func (rl *RateLimiter) Unblock(identifier string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	tracker, exists := rl.requests[identifier]
	if exists {
		tracker.Blocked = false
		tracker.BlockedAt = time.Time{}

		if rl.logger != nil {
			rl.logger.Info("Identifier unblocked", "identifier", identifier)
		}
	}
}

// GetStats returns rate limiting statistics
func (rl *RateLimiter) GetStats() *RateLimitStats {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	stats := &RateLimitStats{
		TotalTrackers:   len(rl.requests),
		BlockedTrackers: 0,
		ActiveTrackers:  0,
		Timestamp:       time.Now(),
	}

	now := time.Now()
	for _, tracker := range rl.requests {
		if tracker.Blocked {
			stats.BlockedTrackers++
		}
		if now.Before(tracker.Window) {
			stats.ActiveTrackers++
		}
	}

	return stats
}

// cleanupRoutine periodically cleans up expired trackers
func (rl *RateLimiter) cleanupRoutine(interval time.Duration, maxTrackers int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup(maxTrackers)
	}
}

// cleanup removes expired and old trackers
func (rl *RateLimiter) cleanup(maxTrackers int) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	// Find expired trackers
	for identifier, tracker := range rl.requests {
		// Remove if window expired and not blocked, or if blocked time expired
		if (now.After(tracker.Window) && !tracker.Blocked) ||
			(tracker.Blocked && now.After(tracker.BlockedAt.Add(time.Hour))) {
			expiredKeys = append(expiredKeys, identifier)
		}
	}

	// Remove expired trackers
	for _, key := range expiredKeys {
		delete(rl.requests, key)
	}

	// If still too many trackers, remove oldest
	if len(rl.requests) > maxTrackers {
		oldestKeys := make([]string, 0)
		oldestTime := now

		for identifier, tracker := range rl.requests {
			if tracker.LastSeen.Before(oldestTime) {
				oldestTime = tracker.LastSeen
				oldestKeys = []string{identifier}
			} else if tracker.LastSeen.Equal(oldestTime) {
				oldestKeys = append(oldestKeys, identifier)
			}
		}

		// Remove oldest trackers until under limit
		removed := 0
		for _, key := range oldestKeys {
			if len(rl.requests)-removed <= maxTrackers {
				break
			}
			delete(rl.requests, key)
			removed++
		}
	}

	if rl.logger != nil && len(expiredKeys) > 0 {
		rl.logger.Debug("Cleaned up rate limit trackers", "expired", len(expiredKeys), "total", len(rl.requests))
	}
}

// RateLimitStats represents rate limiting statistics
type RateLimitStats struct {
	TotalTrackers   int       `json:"total_trackers"`
	BlockedTrackers int       `json:"blocked_trackers"`
	ActiveTrackers  int       `json:"active_trackers"`
	Timestamp       time.Time `json:"timestamp"`
}

// DistributedRateLimiter provides distributed rate limiting using Redis
type DistributedRateLimiter struct {
	cache  CacheInterface
	config *RateLimitConfig
	logger Logger
}

// CacheInterface defines cache operations for distributed rate limiting
type CacheInterface interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Increment(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
}

// NewDistributedRateLimiter creates a new distributed rate limiter
func NewDistributedRateLimiter(cache CacheInterface, config *RateLimitConfig, logger Logger) *DistributedRateLimiter {
	return &DistributedRateLimiter{
		cache:  cache,
		config: config,
		logger: logger,
	}
}

// Allow checks if a request is allowed using distributed cache
func (drl *DistributedRateLimiter) Allow(ctx context.Context, identifier string) bool {
	result := drl.Check(ctx, identifier)
	return result.Allowed
}

// Check performs a distributed rate limit check
func (drl *DistributedRateLimiter) Check(ctx context.Context, identifier string) *RateLimitResult {
	key := fmt.Sprintf("rate_limit:%s", identifier)
	blockKey := fmt.Sprintf("rate_limit_block:%s", identifier)

	// Check if blocked
	if _, err := drl.cache.Get(ctx, blockKey); err == nil {
		return &RateLimitResult{
			Allowed: false,
			Blocked: true,
		}
	}

	// Increment counter
	count, err := drl.cache.Increment(ctx, key)
	if err != nil {
		drl.logger.Error("Failed to increment rate limit counter", err)
		return &RateLimitResult{Allowed: true} // Fail open
	}

	// Set expiration on first request
	if count == 1 {
		if err := drl.cache.Expire(ctx, key, drl.config.WindowDuration); err != nil {
			drl.logger.Error("Failed to set rate limit expiration", err)
		}
	}

	// Check if limit exceeded
	if count > int64(drl.config.RequestsPerWindow) {
		// Block the identifier
		if err := drl.cache.Set(ctx, blockKey, "blocked", drl.config.BlockDuration); err != nil {
			drl.logger.Error("Failed to set rate limit block", err)
		}

		drl.logger.Warn("Distributed rate limit exceeded", "identifier", identifier, "count", count)

		return &RateLimitResult{
			Allowed:   false,
			Remaining: 0,
			Blocked:   true,
		}
	}

	return &RateLimitResult{
		Allowed:   true,
		Remaining: drl.config.RequestsPerWindow - int(count),
	}
}

// Reset resets the distributed rate limit for an identifier
func (drl *DistributedRateLimiter) Reset(ctx context.Context, identifier string) error {
	key := fmt.Sprintf("rate_limit:%s", identifier)
	blockKey := fmt.Sprintf("rate_limit_block:%s", identifier)

	// Delete both counter and block keys
	if err := drl.cache.Set(ctx, key, 0, time.Second); err != nil {
		return fmt.Errorf("failed to reset rate limit counter: %w", err)
	}

	if err := drl.cache.Set(ctx, blockKey, "", time.Second); err != nil {
		return fmt.Errorf("failed to reset rate limit block: %w", err)
	}

	drl.logger.Debug("Distributed rate limit reset", "identifier", identifier)
	return nil
}

// AdaptiveRateLimiter provides adaptive rate limiting based on system load
type AdaptiveRateLimiter struct {
	baseLimiter    *RateLimiter
	loadThreshold  float64
	adaptiveFactor float64
	logger         Logger
}

// NewAdaptiveRateLimiter creates a new adaptive rate limiter
func NewAdaptiveRateLimiter(baseLimit int, window time.Duration, loadThreshold, adaptiveFactor float64, logger Logger) *AdaptiveRateLimiter {
	return &AdaptiveRateLimiter{
		baseLimiter:    NewRateLimiter(baseLimit, window),
		loadThreshold:  loadThreshold,
		adaptiveFactor: adaptiveFactor,
		logger:         logger,
	}
}

// Allow checks if a request is allowed with adaptive limiting
func (arl *AdaptiveRateLimiter) Allow(identifier string, currentLoad float64) bool {
	// Adjust limit based on current system load
	adjustedLimit := arl.calculateAdjustedLimit(currentLoad)
	
	// Temporarily adjust the base limiter's limit
	originalLimit := arl.baseLimiter.limit
	arl.baseLimiter.limit = adjustedLimit
	
	result := arl.baseLimiter.Allow(identifier)
	
	// Restore original limit
	arl.baseLimiter.limit = originalLimit
	
	if !result && arl.logger != nil {
		arl.logger.Debug("Adaptive rate limit applied",
			"identifier", identifier,
			"current_load", currentLoad,
			"adjusted_limit", adjustedLimit,
			"original_limit", originalLimit)
	}
	
	return result
}

// calculateAdjustedLimit calculates the adjusted limit based on system load
func (arl *AdaptiveRateLimiter) calculateAdjustedLimit(currentLoad float64) int {
	if currentLoad <= arl.loadThreshold {
		return arl.baseLimiter.limit
	}
	
	// Reduce limit as load increases
	loadFactor := (currentLoad - arl.loadThreshold) / (1.0 - arl.loadThreshold)
	reduction := int(float64(arl.baseLimiter.limit) * arl.adaptiveFactor * loadFactor)
	adjustedLimit := arl.baseLimiter.limit - reduction
	
	// Ensure minimum limit of 1
	if adjustedLimit < 1 {
		adjustedLimit = 1
	}
	
	return adjustedLimit
}
