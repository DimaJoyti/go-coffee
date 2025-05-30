package ratelimiter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RedisLimiter implements distributed rate limiting using Redis
type RedisLimiter struct {
	redis  *redis.Client
	logger *zap.Logger
	config *RedisLimiterConfig
}

// RedisLimiterConfig contains configuration for Redis rate limiter
type RedisLimiterConfig struct {
	KeyPrefix    string
	DefaultTTL   time.Duration
	CleanupInterval time.Duration
}

// LimitResult represents the result of a rate limit check
type LimitResult struct {
	Allowed      bool          `json:"allowed"`
	Limit        int64         `json:"limit"`
	Remaining    int64         `json:"remaining"`
	ResetTime    time.Time     `json:"reset_time"`
	RetryAfter   time.Duration `json:"retry_after,omitempty"`
	WindowStart  time.Time     `json:"window_start"`
	WindowEnd    time.Time     `json:"window_end"`
}

// RateLimitRule defines a rate limiting rule
type RateLimitRule struct {
	Key      string        `json:"key"`
	Limit    int64         `json:"limit"`
	Window   time.Duration `json:"window"`
	Burst    int64         `json:"burst,omitempty"`
	Algorithm string       `json:"algorithm"` // "sliding_window", "token_bucket", "fixed_window"
}

// NewRedisLimiter creates a new Redis-based rate limiter
func NewRedisLimiter(redisClient *redis.Client, logger *zap.Logger, config *RedisLimiterConfig) *RedisLimiter {
	if config == nil {
		config = &RedisLimiterConfig{
			KeyPrefix:       "rate_limit:",
			DefaultTTL:      time.Hour,
			CleanupInterval: 10 * time.Minute,
		}
	}

	limiter := &RedisLimiter{
		redis:  redisClient,
		logger: logger,
		config: config,
	}

	// Start cleanup goroutine
	go limiter.startCleanup()

	return limiter
}

// Allow checks if a request is allowed under the rate limit
func (rl *RedisLimiter) Allow(ctx context.Context, rule *RateLimitRule) (*LimitResult, error) {
	switch rule.Algorithm {
	case "sliding_window":
		return rl.slidingWindowAllow(ctx, rule)
	case "token_bucket":
		return rl.tokenBucketAllow(ctx, rule)
	case "fixed_window":
		return rl.fixedWindowAllow(ctx, rule)
	default:
		return rl.slidingWindowAllow(ctx, rule) // Default to sliding window
	}
}

// slidingWindowAllow implements sliding window rate limiting
func (rl *RedisLimiter) slidingWindowAllow(ctx context.Context, rule *RateLimitRule) (*LimitResult, error) {
	now := time.Now()
	windowStart := now.Add(-rule.Window)
	key := rl.getKey(rule.Key, "sliding")

	rl.logger.Debug("Checking sliding window rate limit",
		zap.String("key", key),
		zap.Int64("limit", rule.Limit),
		zap.Duration("window", rule.Window),
	)

	// Lua script for atomic sliding window check
	luaScript := `
		local key = KEYS[1]
		local window_start = tonumber(ARGV[1])
		local now = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		local ttl = tonumber(ARGV[4])

		-- Remove old entries
		redis.call('ZREMRANGEBYSCORE', key, 0, window_start)

		-- Count current entries
		local current = redis.call('ZCARD', key)

		if current < limit then
			-- Add current request
			redis.call('ZADD', key, now, now)
			redis.call('EXPIRE', key, ttl)
			return {1, current + 1, limit - current - 1}
		else
			return {0, current, 0}
		end
	`

	result, err := rl.redis.Eval(ctx, luaScript, []string{key}, 
		windowStart.UnixMilli(), now.UnixMilli(), rule.Limit, int64(rl.config.DefaultTTL.Seconds())).Result()
	
	if err != nil {
		return nil, fmt.Errorf("failed to execute sliding window script: %w", err)
	}

	resultSlice := result.([]interface{})
	allowed := resultSlice[0].(int64) == 1
	_ = resultSlice[1].(int64) // current count (not used)
	remaining := resultSlice[2].(int64)

	resetTime := now.Add(rule.Window)
	var retryAfter time.Duration
	if !allowed {
		retryAfter = rule.Window
	}

	return &LimitResult{
		Allowed:     allowed,
		Limit:       rule.Limit,
		Remaining:   remaining,
		ResetTime:   resetTime,
		RetryAfter:  retryAfter,
		WindowStart: windowStart,
		WindowEnd:   now,
	}, nil
}

// tokenBucketAllow implements token bucket rate limiting
func (rl *RedisLimiter) tokenBucketAllow(ctx context.Context, rule *RateLimitRule) (*LimitResult, error) {
	now := time.Now()
	key := rl.getKey(rule.Key, "bucket")
	
	burst := rule.Burst
	if burst == 0 {
		burst = rule.Limit
	}

	rl.logger.Debug("Checking token bucket rate limit",
		zap.String("key", key),
		zap.Int64("limit", rule.Limit),
		zap.Int64("burst", burst),
		zap.Duration("window", rule.Window),
	)

	// Lua script for atomic token bucket check
	luaScript := `
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local rate = tonumber(ARGV[2])
		local burst = tonumber(ARGV[3])
		local window_ms = tonumber(ARGV[4])
		local ttl = tonumber(ARGV[5])

		local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
		local tokens = tonumber(bucket[1]) or burst
		local last_refill = tonumber(bucket[2]) or now

		-- Calculate tokens to add
		local time_passed = now - last_refill
		local tokens_to_add = math.floor((time_passed / window_ms) * rate)
		tokens = math.min(burst, tokens + tokens_to_add)

		if tokens >= 1 then
			tokens = tokens - 1
			redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
			redis.call('EXPIRE', key, ttl)
			return {1, tokens, burst}
		else
			redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
			redis.call('EXPIRE', key, ttl)
			return {0, tokens, burst}
		end
	`

	result, err := rl.redis.Eval(ctx, luaScript, []string{key},
		now.UnixMilli(), rule.Limit, burst, rule.Window.Milliseconds(), int64(rl.config.DefaultTTL.Seconds())).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to execute token bucket script: %w", err)
	}

	resultSlice := result.([]interface{})
	allowed := resultSlice[0].(int64) == 1
	tokens := resultSlice[1].(int64)
	capacity := resultSlice[2].(int64)

	var retryAfter time.Duration
	if !allowed {
		// Calculate time until next token is available
		retryAfter = time.Duration(float64(rule.Window) / float64(rule.Limit))
	}

	return &LimitResult{
		Allowed:    allowed,
		Limit:      capacity,
		Remaining:  tokens,
		ResetTime:  now.Add(retryAfter),
		RetryAfter: retryAfter,
	}, nil
}

// fixedWindowAllow implements fixed window rate limiting
func (rl *RedisLimiter) fixedWindowAllow(ctx context.Context, rule *RateLimitRule) (*LimitResult, error) {
	now := time.Now()
	windowStart := now.Truncate(rule.Window)
	windowEnd := windowStart.Add(rule.Window)
	key := rl.getKey(rule.Key, fmt.Sprintf("fixed:%d", windowStart.Unix()))

	rl.logger.Debug("Checking fixed window rate limit",
		zap.String("key", key),
		zap.Int64("limit", rule.Limit),
		zap.Time("window_start", windowStart),
		zap.Time("window_end", windowEnd),
	)

	// Lua script for atomic fixed window check
	luaScript := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local ttl = tonumber(ARGV[2])

		local current = redis.call('GET', key)
		if current == false then
			current = 0
		else
			current = tonumber(current)
		end

		if current < limit then
			local new_count = redis.call('INCR', key)
			redis.call('EXPIRE', key, ttl)
			return {1, new_count, limit - new_count}
		else
			return {0, current, 0}
		end
	`

	result, err := rl.redis.Eval(ctx, luaScript, []string{key},
		rule.Limit, int64(rule.Window.Seconds())).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to execute fixed window script: %w", err)
	}

	resultSlice := result.([]interface{})
	allowed := resultSlice[0].(int64) == 1
	_ = resultSlice[1].(int64) // current count (not used)
	remaining := resultSlice[2].(int64)

	var retryAfter time.Duration
	if !allowed {
		retryAfter = windowEnd.Sub(now)
	}

	return &LimitResult{
		Allowed:     allowed,
		Limit:       rule.Limit,
		Remaining:   remaining,
		ResetTime:   windowEnd,
		RetryAfter:  retryAfter,
		WindowStart: windowStart,
		WindowEnd:   windowEnd,
	}, nil
}

// AllowN checks if N requests are allowed
func (rl *RedisLimiter) AllowN(ctx context.Context, rule *RateLimitRule, n int64) (*LimitResult, error) {
	if n <= 0 {
		return &LimitResult{Allowed: true}, nil
	}

	// For simplicity, we'll check n times
	// In production, you might want to optimize this
	for i := int64(0); i < n; i++ {
		result, err := rl.Allow(ctx, rule)
		if err != nil {
			return nil, err
		}
		if !result.Allowed {
			return result, nil
		}
	}

	// Return the last result
	return rl.Allow(ctx, rule)
}

// Reset resets the rate limit for a key
func (rl *RedisLimiter) Reset(ctx context.Context, key string) error {
	pattern := rl.getKey(key, "*")
	keys, err := rl.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys for reset: %w", err)
	}

	if len(keys) > 0 {
		if err := rl.redis.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}

	rl.logger.Info("Rate limit reset", zap.String("key", key), zap.Int("deleted_keys", len(keys)))
	return nil
}

// GetStats returns statistics for a rate limit key
func (rl *RedisLimiter) GetStats(ctx context.Context, rule *RateLimitRule) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	switch rule.Algorithm {
	case "sliding_window":
		key := rl.getKey(rule.Key, "sliding")
		count, err := rl.redis.ZCard(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		stats["current_requests"] = count
		stats["algorithm"] = "sliding_window"

	case "token_bucket":
		key := rl.getKey(rule.Key, "bucket")
		bucket, err := rl.redis.HMGet(ctx, key, "tokens", "last_refill").Result()
		if err != nil {
			return nil, err
		}
		
		tokens := int64(0)
		lastRefill := int64(0)
		
		if bucket[0] != nil {
			tokens, _ = strconv.ParseInt(bucket[0].(string), 10, 64)
		}
		if bucket[1] != nil {
			lastRefill, _ = strconv.ParseInt(bucket[1].(string), 10, 64)
		}
		
		stats["tokens"] = tokens
		stats["last_refill"] = time.UnixMilli(lastRefill)
		stats["algorithm"] = "token_bucket"

	case "fixed_window":
		now := time.Now()
		windowStart := now.Truncate(rule.Window)
		key := rl.getKey(rule.Key, fmt.Sprintf("fixed:%d", windowStart.Unix()))
		
		count, err := rl.redis.Get(ctx, key).Int64()
		if err != nil && err != redis.Nil {
			return nil, err
		}
		
		stats["current_requests"] = count
		stats["window_start"] = windowStart
		stats["window_end"] = windowStart.Add(rule.Window)
		stats["algorithm"] = "fixed_window"
	}

	stats["limit"] = rule.Limit
	stats["window"] = rule.Window
	stats["key"] = rule.Key

	return stats, nil
}

// getKey generates a Redis key for rate limiting
func (rl *RedisLimiter) getKey(userKey, algorithm string) string {
	return fmt.Sprintf("%s%s:%s", rl.config.KeyPrefix, userKey, algorithm)
}

// startCleanup starts a background cleanup process
func (rl *RedisLimiter) startCleanup() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes expired rate limit entries
func (rl *RedisLimiter) cleanup() {
	ctx := context.Background()
	now := time.Now()
	cutoff := now.Add(-rl.config.DefaultTTL).UnixMilli()

	// Clean up sliding window entries
	pattern := rl.config.KeyPrefix + "*:sliding"
	keys, err := rl.redis.Keys(ctx, pattern).Result()
	if err != nil {
		rl.logger.Error("Failed to get sliding window keys for cleanup", zap.Error(err))
		return
	}

	for _, key := range keys {
		removed, err := rl.redis.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(cutoff, 10)).Result()
		if err != nil {
			rl.logger.Error("Failed to cleanup sliding window key", zap.String("key", key), zap.Error(err))
			continue
		}

		if removed > 0 {
			rl.logger.Debug("Cleaned up sliding window entries",
				zap.String("key", key),
				zap.Int64("removed", removed),
			)
		}
	}
}

// MultiLimiter manages multiple rate limiters
type MultiLimiter struct {
	limiters map[string]*RedisLimiter
	logger   *zap.Logger
}

// NewMultiLimiter creates a new multi-limiter
func NewMultiLimiter(logger *zap.Logger) *MultiLimiter {
	return &MultiLimiter{
		limiters: make(map[string]*RedisLimiter),
		logger:   logger,
	}
}

// AddLimiter adds a rate limiter
func (ml *MultiLimiter) AddLimiter(name string, limiter *RedisLimiter) {
	ml.limiters[name] = limiter
}

// CheckAll checks all rate limiters
func (ml *MultiLimiter) CheckAll(ctx context.Context, rules map[string]*RateLimitRule) (map[string]*LimitResult, error) {
	results := make(map[string]*LimitResult)

	for name, rule := range rules {
		if limiter, exists := ml.limiters[name]; exists {
			result, err := limiter.Allow(ctx, rule)
			if err != nil {
				return nil, fmt.Errorf("limiter %s failed: %w", name, err)
			}
			results[name] = result

			// If any limiter denies, return immediately
			if !result.Allowed {
				return results, nil
			}
		}
	}

	return results, nil
}
