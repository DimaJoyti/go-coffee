package concurrency

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/cache"
	"go.uber.org/zap"
)

// RateLimiter provides advanced rate limiting functionality
type RateLimiter struct {
	logger      *zap.Logger
	cache       *cache.Manager
	config      *RateLimiterConfig
	localLimits map[string]*SlidingWindow
	mu          sync.RWMutex
}

// RateLimiterConfig contains rate limiter configuration
type RateLimiterConfig struct {
	Algorithm           string                    `json:"algorithm"`           // "sliding_window", "token_bucket", "fixed_window"
	DefaultLimit        int64                     `json:"default_limit"`       // requests per window
	DefaultWindow       time.Duration             `json:"default_window"`      // window duration
	BurstSize           int64                     `json:"burst_size"`          // burst allowance
	Distributed         bool                      `json:"distributed"`         // use Redis for distributed limiting
	CleanupInterval     time.Duration             `json:"cleanup_interval"`    // cleanup interval for local limits
	EndpointLimits      map[string]*EndpointLimit `json:"endpoint_limits"`     // per-endpoint limits
	UserLimits          map[string]*UserLimit     `json:"user_limits"`         // per-user limits
	IPLimits            map[string]*IPLimit       `json:"ip_limits"`           // per-IP limits
	HeadersEnabled      bool                      `json:"headers_enabled"`     // include rate limit headers
	BlockOnExceed       bool                      `json:"block_on_exceed"`     // block or just log
	RetryAfterEnabled   bool                      `json:"retry_after_enabled"` // include Retry-After header
}

// EndpointLimit defines rate limits for specific endpoints
type EndpointLimit struct {
	Path   string        `json:"path"`
	Limit  int64         `json:"limit"`
	Window time.Duration `json:"window"`
}

// UserLimit defines rate limits for specific users
type UserLimit struct {
	UserID string        `json:"user_id"`
	Limit  int64         `json:"limit"`
	Window time.Duration `json:"window"`
}

// IPLimit defines rate limits for specific IP addresses
type IPLimit struct {
	IPAddress string        `json:"ip_address"`
	Limit     int64         `json:"limit"`
	Window    time.Duration `json:"window"`
}

// SlidingWindow implements sliding window rate limiting
type SlidingWindow struct {
	limit      int64
	window     time.Duration
	requests   []time.Time
	mu         sync.Mutex
	lastCleanup time.Time
}

// RateLimitResult contains the result of a rate limit check
type RateLimitResult struct {
	Allowed       bool          `json:"allowed"`
	Limit         int64         `json:"limit"`
	Remaining     int64         `json:"remaining"`
	ResetTime     time.Time     `json:"reset_time"`
	RetryAfter    time.Duration `json:"retry_after"`
	WindowStart   time.Time     `json:"window_start"`
	WindowEnd     time.Time     `json:"window_end"`
}

// RateLimitKey represents a rate limit key
type RateLimitKey struct {
	Type       string // "global", "user", "ip", "endpoint"
	Identifier string // user ID, IP address, endpoint path
	Endpoint   string // endpoint path for user/IP limits
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *RateLimiterConfig, cacheManager *cache.Manager, logger *zap.Logger) *RateLimiter {
	rl := &RateLimiter{
		logger:      logger,
		cache:       cacheManager,
		config:      config,
		localLimits: make(map[string]*SlidingWindow),
	}

	// Start cleanup goroutine for local limits
	if !config.Distributed {
		go rl.cleanupLocalLimits()
	}

	return rl
}

// CheckLimit checks if a request is within rate limits
func (rl *RateLimiter) CheckLimit(ctx context.Context, key RateLimitKey) (*RateLimitResult, error) {
	// Determine the limit and window for this key
	limit, window := rl.getLimitAndWindow(key)

	if rl.config.Distributed && rl.cache != nil {
		return rl.checkDistributedLimit(ctx, key, limit, window)
	}

	return rl.checkLocalLimit(key, limit, window)
}

// getLimitAndWindow returns the appropriate limit and window for a key
func (rl *RateLimiter) getLimitAndWindow(key RateLimitKey) (int64, time.Duration) {
	// Check endpoint-specific limits
	if rl.config.EndpointLimits != nil {
		for _, endpointLimit := range rl.config.EndpointLimits {
			if endpointLimit.Path == key.Endpoint {
				return endpointLimit.Limit, endpointLimit.Window
			}
		}
	}

	// Check user-specific limits
	if key.Type == "user" && rl.config.UserLimits != nil {
		for _, userLimit := range rl.config.UserLimits {
			if userLimit.UserID == key.Identifier {
				return userLimit.Limit, userLimit.Window
			}
		}
	}

	// Check IP-specific limits
	if key.Type == "ip" && rl.config.IPLimits != nil {
		for _, ipLimit := range rl.config.IPLimits {
			if ipLimit.IPAddress == key.Identifier {
				return ipLimit.Limit, ipLimit.Window
			}
		}
	}

	// Return default limits
	return rl.config.DefaultLimit, rl.config.DefaultWindow
}

// checkDistributedLimit checks rate limit using Redis
func (rl *RateLimiter) checkDistributedLimit(ctx context.Context, key RateLimitKey, limit int64, window time.Duration) (*RateLimitResult, error) {
	keyStr := rl.buildKeyString(key)
	now := time.Now()
	windowStart := now.Truncate(window)
	windowEnd := windowStart.Add(window)

	// Use Redis for distributed rate limiting
	cacheKey := fmt.Sprintf("rate_limit:%s:%d", keyStr, windowStart.Unix())

	// Get current count
	var currentCount int64
	err := rl.cache.Get(ctx, cacheKey, &currentCount)
	if err != nil {
		// Key doesn't exist, start with 0
		currentCount = 0
	}

	result := &RateLimitResult{
		Limit:       limit,
		WindowStart: windowStart,
		WindowEnd:   windowEnd,
		ResetTime:   windowEnd,
	}

	if currentCount >= limit {
		result.Allowed = false
		result.Remaining = 0
		result.RetryAfter = time.Until(windowEnd)
	} else {
		result.Allowed = true
		result.Remaining = limit - currentCount - 1

		// Increment counter
		newCount := currentCount + 1
		err = rl.cache.Set(ctx, cacheKey, newCount, window)
		if err != nil {
			rl.logger.Error("Failed to update rate limit counter", zap.Error(err))
		}
	}

	return result, nil
}

// checkLocalLimit checks rate limit using local sliding window
func (rl *RateLimiter) checkLocalLimit(key RateLimitKey, limit int64, window time.Duration) (*RateLimitResult, error) {
	keyStr := rl.buildKeyString(key)

	rl.mu.Lock()
	slidingWindow, exists := rl.localLimits[keyStr]
	if !exists {
		slidingWindow = &SlidingWindow{
			limit:   limit,
			window:  window,
			requests: make([]time.Time, 0),
		}
		rl.localLimits[keyStr] = slidingWindow
	}
	rl.mu.Unlock()

	return slidingWindow.checkLimit()
}

// checkLimit checks if a request is allowed in the sliding window
func (sw *SlidingWindow) checkLimit() (*RateLimitResult, error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-sw.window)

	// Clean up old requests
	sw.cleanupOldRequests(windowStart)

	result := &RateLimitResult{
		Limit:       sw.limit,
		WindowStart: windowStart,
		WindowEnd:   now,
		ResetTime:   now.Add(sw.window),
	}

	currentCount := int64(len(sw.requests))

	if currentCount >= sw.limit {
		result.Allowed = false
		result.Remaining = 0
		if len(sw.requests) > 0 {
			oldestRequest := sw.requests[0]
			result.RetryAfter = oldestRequest.Add(sw.window).Sub(now)
		}
	} else {
		result.Allowed = true
		result.Remaining = sw.limit - currentCount - 1
		sw.requests = append(sw.requests, now)
	}

	return result, nil
}

// cleanupOldRequests removes requests outside the current window
func (sw *SlidingWindow) cleanupOldRequests(windowStart time.Time) {
	validRequests := make([]time.Time, 0, len(sw.requests))
	for _, req := range sw.requests {
		if req.After(windowStart) {
			validRequests = append(validRequests, req)
		}
	}
	sw.requests = validRequests
	sw.lastCleanup = time.Now()
}

// cleanupLocalLimits periodically cleans up unused local limits
func (rl *RateLimiter) cleanupLocalLimits() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, window := range rl.localLimits {
			// Remove windows that haven't been used recently
			if now.Sub(window.lastCleanup) > rl.config.CleanupInterval*2 {
				delete(rl.localLimits, key)
			}
		}
		rl.mu.Unlock()
	}
}

// buildKeyString builds a string key from RateLimitKey
func (rl *RateLimiter) buildKeyString(key RateLimitKey) string {
	return fmt.Sprintf("%s:%s:%s", key.Type, key.Identifier, key.Endpoint)
}

// HTTPMiddleware returns HTTP middleware for rate limiting
func (rl *RateLimiter) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Extract rate limit keys
			keys := rl.extractRateLimitKeys(r)

			// Check all applicable rate limits
			for _, key := range keys {
				result, err := rl.CheckLimit(ctx, key)
				if err != nil {
					rl.logger.Error("Rate limit check failed", zap.Error(err))
					continue
				}

				// Add rate limit headers
				if rl.config.HeadersEnabled {
					rl.addRateLimitHeaders(w, result)
				}

				// Block request if rate limit exceeded
				if !result.Allowed {
					if rl.config.BlockOnExceed {
						if rl.config.RetryAfterEnabled && result.RetryAfter > 0 {
							w.Header().Set("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))
						}
						http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
						return
					} else {
						rl.logger.Warn("Rate limit exceeded (not blocking)",
							zap.String("key", rl.buildKeyString(key)),
							zap.Int64("limit", result.Limit))
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractRateLimitKeys extracts rate limit keys from HTTP request
func (rl *RateLimiter) extractRateLimitKeys(r *http.Request) []RateLimitKey {
	var keys []RateLimitKey

	// Global rate limit
	keys = append(keys, RateLimitKey{
		Type:       "global",
		Identifier: "global",
		Endpoint:   r.URL.Path,
	})

	// IP-based rate limit
	clientIP := rl.getClientIP(r)
	if clientIP != "" {
		keys = append(keys, RateLimitKey{
			Type:       "ip",
			Identifier: clientIP,
			Endpoint:   r.URL.Path,
		})
	}

	// User-based rate limit (if user is authenticated)
	userID := rl.getUserID(r)
	if userID != "" {
		keys = append(keys, RateLimitKey{
			Type:       "user",
			Identifier: userID,
			Endpoint:   r.URL.Path,
		})
	}

	// Endpoint-specific rate limit
	keys = append(keys, RateLimitKey{
		Type:       "endpoint",
		Identifier: r.URL.Path,
		Endpoint:   r.URL.Path,
	})

	return keys
}

// getClientIP extracts client IP from request
func (rl *RateLimiter) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Use remote address
	return r.RemoteAddr
}

// getUserID extracts user ID from request (implement based on your auth system)
func (rl *RateLimiter) getUserID(r *http.Request) string {
	// This would typically extract user ID from JWT token or session
	// For demo purposes, we'll use a header
	return r.Header.Get("X-User-ID")
}

// addRateLimitHeaders adds rate limit headers to response
func (rl *RateLimiter) addRateLimitHeaders(w http.ResponseWriter, result *RateLimitResult) {
	w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(result.Limit, 10))
	w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))

	if !result.Allowed && result.RetryAfter > 0 {
		w.Header().Set("X-RateLimit-Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))
	}
}

// GetStats returns rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	stats := map[string]interface{}{
		"algorithm":           rl.config.Algorithm,
		"distributed":         rl.config.Distributed,
		"local_limits_count":  len(rl.localLimits),
		"default_limit":       rl.config.DefaultLimit,
		"default_window":      rl.config.DefaultWindow.String(),
		"endpoint_limits":     len(rl.config.EndpointLimits),
		"user_limits":         len(rl.config.UserLimits),
		"ip_limits":           len(rl.config.IPLimits),
	}

	// Add local window statistics
	if !rl.config.Distributed {
		windowStats := make(map[string]interface{})
		for key, window := range rl.localLimits {
			window.mu.Lock()
			windowStats[key] = map[string]interface{}{
				"current_requests": len(window.requests),
				"limit":           window.limit,
				"window_duration": window.window.String(),
				"last_cleanup":    window.lastCleanup,
			}
			window.mu.Unlock()
		}
		stats["windows"] = windowStats
	}

	return stats
}
