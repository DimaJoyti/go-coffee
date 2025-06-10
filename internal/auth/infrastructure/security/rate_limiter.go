package security

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/go-redis/redis/v8"
)

// RateLimiter defines the interface for rate limiting
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
	Reset(ctx context.Context, key string) error
	GetRemaining(ctx context.Context, key string, limit int, window time.Duration) (int, error)
}

// RedisRateLimiter implements rate limiting using Redis
type RedisRateLimiter struct {
	client *redis.Client
	logger *logger.Logger
	prefix string
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(client *redis.Client, logger *logger.Logger) RateLimiter {
	return &RedisRateLimiter{
		client: client,
		logger: logger,
		prefix: "auth:rate_limit:",
	}
}

// Allow checks if a request is allowed based on rate limiting rules
func (rl *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	fullKey := rl.getKey(key)

	// Use sliding window log algorithm
	now := time.Now()
	windowStart := now.Add(-window)

	// Remove expired entries and count current requests
	pipe := rl.client.TxPipeline()

	// Remove entries older than the window
	pipe.ZRemRangeByScore(ctx, fullKey, "0", strconv.FormatInt(windowStart.UnixNano(), 10))

	// Count current entries
	countCmd := pipe.ZCard(ctx, fullKey)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		rl.logger.ErrorWithFields("Failed to execute rate limit pipeline",
			logger.Error(err),
			logger.String("key", key))
		return false, fmt.Errorf("failed to execute rate limit pipeline: %w", err)
	}

	currentCount, err := countCmd.Result()
	if err != nil {
		rl.logger.ErrorWithFields("Failed to get current count",
			logger.Error(err),
			logger.String("key", key))
		return false, fmt.Errorf("failed to get current count: %w", err)
	}

	// Check if limit is exceeded
	if currentCount >= int64(limit) {
		rl.logger.WarnWithFields("Rate limit exceeded",
			logger.String("key", key),
			logger.String("current_count", fmt.Sprintf("%d", currentCount)),
			logger.Int("limit", limit))
		return false, nil
	}

	// Add current request
	score := now.UnixNano()
	err = rl.client.ZAdd(ctx, fullKey, &redis.Z{
		Score:  float64(score),
		Member: score, // Use timestamp as both score and member
	}).Err()

	if err != nil {
		rl.logger.ErrorWithFields("Failed to add request to rate limit",
			logger.Error(err),
			logger.String("key", key))
		return false, fmt.Errorf("failed to add request to rate limit: %w", err)
	}

	// Set expiry for the key
	rl.client.Expire(ctx, fullKey, window+time.Minute) // Add buffer to prevent premature expiry

	return true, nil
}

// Reset resets the rate limit for a key
func (rl *RedisRateLimiter) Reset(ctx context.Context, key string) error {
	fullKey := rl.getKey(key)

	err := rl.client.Del(ctx, fullKey).Err()
	if err != nil {
		rl.logger.ErrorWithFields("Failed to reset rate limit",
			logger.Error(err),
			logger.String("key", key))
		return fmt.Errorf("failed to reset rate limit: %w", err)
	}

	rl.logger.InfoWithFields("Rate limit reset", logger.String("key", key))
	return nil
}

// GetRemaining returns the number of remaining requests in the current window
func (rl *RedisRateLimiter) GetRemaining(ctx context.Context, key string, limit int, window time.Duration) (int, error) {
	fullKey := rl.getKey(key)

	// Remove expired entries and count current requests
	now := time.Now()
	windowStart := now.Add(-window)

	pipe := rl.client.TxPipeline()

	// Remove entries older than the window
	pipe.ZRemRangeByScore(ctx, fullKey, "0", strconv.FormatInt(windowStart.UnixNano(), 10))

	// Count current entries
	countCmd := pipe.ZCard(ctx, fullKey)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		rl.logger.ErrorWithFields("Failed to get remaining requests",
			logger.Error(err),
			logger.String("key", key))
		return 0, fmt.Errorf("failed to get remaining requests: %w", err)
	}

	currentCount, err := countCmd.Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get current count: %w", err)
	}

	remaining := limit - int(currentCount)
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// getKey generates the full Redis key for rate limiting
func (rl *RedisRateLimiter) getKey(key string) string {
	return fmt.Sprintf("%s%s", rl.prefix, key)
}

// TokenBucketRateLimiter implements token bucket algorithm
type TokenBucketRateLimiter struct {
	client *redis.Client
	logger *logger.Logger
	prefix string
}

// NewTokenBucketRateLimiter creates a new token bucket rate limiter
func NewTokenBucketRateLimiter(client *redis.Client, logger *logger.Logger) RateLimiter {
	return &TokenBucketRateLimiter{
		client: client,
		logger: logger,
		prefix: "auth:token_bucket:",
	}
}

// Allow checks if a request is allowed using token bucket algorithm
func (tb *TokenBucketRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	fullKey := tb.getKey(key)

	// Lua script for atomic token bucket operation
	luaScript := `
		local key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local refill_rate = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		
		local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
		local tokens = tonumber(bucket[1]) or capacity
		local last_refill = tonumber(bucket[2]) or now
		
		-- Calculate tokens to add based on time elapsed
		local elapsed = now - last_refill
		local tokens_to_add = math.floor(elapsed * refill_rate)
		tokens = math.min(capacity, tokens + tokens_to_add)
		
		if tokens >= 1 then
			tokens = tokens - 1
			redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
			redis.call('EXPIRE', key, 3600) -- 1 hour expiry
			return 1
		else
			redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
			redis.call('EXPIRE', key, 3600)
			return 0
		end
	`

	// Calculate refill rate (tokens per nanosecond)
	refillRate := float64(limit) / float64(window.Nanoseconds())

	result, err := tb.client.Eval(ctx, luaScript, []string{fullKey}, limit, refillRate, time.Now().UnixNano()).Result()
	if err != nil {
		tb.logger.ErrorWithFields("Failed to execute token bucket script",
			logger.Error(err),
			logger.String("key", key))
		return false, fmt.Errorf("failed to execute token bucket script: %w", err)
	}

	allowed := result.(int64) == 1

	if !allowed {
		tb.logger.WarnWithFields("Token bucket rate limit exceeded",
			logger.String("key", key),
			logger.Int("limit", limit))
	}

	return allowed, nil
}

// Reset resets the token bucket for a key
func (tb *TokenBucketRateLimiter) Reset(ctx context.Context, key string) error {
	fullKey := tb.getKey(key)

	err := tb.client.Del(ctx, fullKey).Err()
	if err != nil {
		tb.logger.ErrorWithFields("Failed to reset token bucket",
			logger.Error(err),
			logger.String("key", key))
		return fmt.Errorf("failed to reset token bucket: %w", err)
	}

	tb.logger.InfoWithFields("Token bucket reset", logger.String("key", key))
	return nil
}

// GetRemaining returns the number of remaining tokens
func (tb *TokenBucketRateLimiter) GetRemaining(ctx context.Context, key string, limit int, window time.Duration) (int, error) {
	fullKey := tb.getKey(key)

	bucket, err := tb.client.HMGet(ctx, fullKey, "tokens", "last_refill").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get bucket info: %w", err)
	}

	tokens := limit // Default to full capacity
	if bucket[0] != nil {
		if t, err := strconv.Atoi(bucket[0].(string)); err == nil {
			tokens = t
		}
	}

	// Calculate additional tokens based on time elapsed
	if bucket[1] != nil {
		if lastRefill, err := strconv.ParseInt(bucket[1].(string), 10, 64); err == nil {
			now := time.Now().UnixNano()
			elapsed := now - lastRefill
			refillRate := float64(limit) / float64(window.Nanoseconds())
			tokensToAdd := int(float64(elapsed) * refillRate)
			tokens = min(limit, tokens+tokensToAdd)
		}
	}

	return tokens, nil
}

// getKey generates the full Redis key for token bucket
func (tb *TokenBucketRateLimiter) getKey(key string) string {
	return fmt.Sprintf("%s%s", tb.prefix, key)
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
