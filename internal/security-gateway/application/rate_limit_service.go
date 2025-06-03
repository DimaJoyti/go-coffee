package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/DimaJoyti/go-coffee/internal/security-gateway/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// RateLimitService provides rate limiting functionality
type RateLimitService struct {
	config      *RateLimitConfig
	redisClient *redis.Client
	logger      *logger.Logger
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool          `yaml:"enabled"`
	RequestsPerMinute int           `yaml:"requests_per_minute"`
	BurstSize         int           `yaml:"burst_size"`
	CleanupInterval   time.Duration `yaml:"cleanup_interval"`
	WindowSize        time.Duration `yaml:"window_size" default:"1m"`
	
	// Different limits for different types
	Limits map[string]RateLimit `yaml:"limits"`
}

// RateLimit represents a specific rate limit configuration
type RateLimit struct {
	RequestsPerMinute int           `yaml:"requests_per_minute"`
	BurstSize         int           `yaml:"burst_size"`
	WindowSize        time.Duration `yaml:"window_size"`
}

// RateLimitType represents different types of rate limits
type RateLimitType string

const (
	RateLimitTypeIP       RateLimitType = "ip"
	RateLimitTypeUser     RateLimitType = "user"
	RateLimitTypeEndpoint RateLimitType = "endpoint"
	RateLimitTypeGlobal   RateLimitType = "global"
)

// NewRateLimitService creates a new rate limit service
func NewRateLimitService(
	config *RateLimitConfig,
	redisClient *redis.Client,
	logger *logger.Logger,
) *RateLimitService {
	service := &RateLimitService{
		config:      config,
		redisClient: redisClient,
		logger:      logger,
	}

	// Set default limits if not configured
	if service.config.Limits == nil {
		service.config.Limits = make(map[string]RateLimit)
	}

	// Set default IP limit
	if _, exists := service.config.Limits["ip"]; !exists {
		service.config.Limits["ip"] = RateLimit{
			RequestsPerMinute: config.RequestsPerMinute,
			BurstSize:         config.BurstSize,
			WindowSize:        config.WindowSize,
		}
	}

	// Start cleanup routine
	go service.cleanupExpiredKeys()

	return service
}

// CheckRateLimit checks if a request is within rate limits
func (r *RateLimitService) CheckRateLimit(ctx context.Context, key string) (bool, *domain.RateLimitInfo, error) {
	if !r.config.Enabled {
		return true, &domain.RateLimitInfo{
			Limit:     -1,
			Remaining: -1,
			Reset:     time.Now().Add(time.Minute),
			Window:    time.Minute,
			Blocked:   false,
		}, nil
	}

	// Get rate limit configuration for this key type
	limitConfig := r.getRateLimitConfig(key)
	
	// Use sliding window rate limiting with Redis
	allowed, info, err := r.slidingWindowRateLimit(ctx, key, limitConfig)
	if err != nil {
		r.logger.WithError(err).Error("Rate limit check failed", map[string]any{
			"key": key,
		})
		// In case of error, allow the request but log it
		return true, &domain.RateLimitInfo{
			Limit:     limitConfig.RequestsPerMinute,
			Remaining: 0,
			Reset:     time.Now().Add(limitConfig.WindowSize),
			Window:    limitConfig.WindowSize,
			Blocked:   false,
		}, err
	}

	return allowed, info, nil
}

// CheckMultipleRateLimits checks multiple rate limits for a request
func (r *RateLimitService) CheckMultipleRateLimits(ctx context.Context, req *domain.SecurityRequest) (bool, []*domain.RateLimitInfo, error) {
	if !r.config.Enabled {
		return true, nil, nil
	}

	var infos []*domain.RateLimitInfo
	
	// Check IP-based rate limit
	ipAllowed, ipInfo, err := r.CheckRateLimit(ctx, fmt.Sprintf("ip:%s", req.IPAddress))
	if err != nil {
		return false, nil, err
	}
	infos = append(infos, ipInfo)
	
	if !ipAllowed {
		return false, infos, nil
	}

	// Check user-based rate limit if user is authenticated
	if req.UserID != "" {
		userAllowed, userInfo, err := r.CheckRateLimit(ctx, fmt.Sprintf("user:%s", req.UserID))
		if err != nil {
			return false, infos, err
		}
		infos = append(infos, userInfo)
		
		if !userAllowed {
			return false, infos, nil
		}
	}

	// Check endpoint-based rate limit
	endpoint := fmt.Sprintf("%s:%s", req.Method, req.URL)
	endpointAllowed, endpointInfo, err := r.CheckRateLimit(ctx, fmt.Sprintf("endpoint:%s", endpoint))
	if err != nil {
		return false, infos, err
	}
	infos = append(infos, endpointInfo)
	
	if !endpointAllowed {
		return false, infos, nil
	}

	// Check global rate limit
	globalAllowed, globalInfo, err := r.CheckRateLimit(ctx, "global")
	if err != nil {
		return false, infos, err
	}
	infos = append(infos, globalInfo)

	return globalAllowed, infos, nil
}

// Sliding window rate limiting implementation
func (r *RateLimitService) slidingWindowRateLimit(ctx context.Context, key string, config RateLimit) (bool, *domain.RateLimitInfo, error) {
	now := time.Now()
	windowStart := now.Add(-config.WindowSize)
	
	// Redis key for this rate limit
	redisKey := fmt.Sprintf("rate_limit:%s", key)
	
	// Use Redis pipeline for atomic operations
	pipe := r.redisClient.Pipeline()
	
	// Remove expired entries
	pipe.ZRemRangeByScore(ctx, redisKey, "0", strconv.FormatInt(windowStart.UnixNano(), 10))
	
	// Count current requests in window
	countCmd := pipe.ZCard(ctx, redisKey)
	
	// Add current request
	pipe.ZAdd(ctx, redisKey, &redis.Z{
		Score:  float64(now.UnixNano()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})
	
	// Set expiration
	pipe.Expire(ctx, redisKey, config.WindowSize*2)
	
	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, nil, fmt.Errorf("failed to execute rate limit pipeline: %w", err)
	}
	
	// Get current count
	currentCount := countCmd.Val()
	
	// Calculate remaining requests
	remaining := int64(config.RequestsPerMinute) - currentCount
	if remaining < 0 {
		remaining = 0
	}
	
	// Calculate reset time
	resetTime := now.Add(config.WindowSize)
	
	// Check if request is allowed
	allowed := currentCount <= int64(config.RequestsPerMinute)
	
	info := &domain.RateLimitInfo{
		Limit:     config.RequestsPerMinute,
		Remaining: int(remaining),
		Reset:     resetTime,
		Window:    config.WindowSize,
		Blocked:   !allowed,
	}
	
	if !allowed {
		r.logger.Warn("Rate limit exceeded", map[string]any{
			"key":           key,
			"current_count": currentCount,
			"limit":         config.RequestsPerMinute,
			"window":        config.WindowSize,
		})
	}
	
	return allowed, info, nil
}

// Token bucket rate limiting implementation (alternative)
func (r *RateLimitService) tokenBucketRateLimit(ctx context.Context, key string, config RateLimit) (bool, *domain.RateLimitInfo, error) {
	now := time.Now()
	redisKey := fmt.Sprintf("token_bucket:%s", key)
	
	// Lua script for atomic token bucket operations
	luaScript := `
		local key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local tokens = tonumber(ARGV[2])
		local interval = tonumber(ARGV[3])
		local now = tonumber(ARGV[4])
		
		local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
		local current_tokens = tonumber(bucket[1]) or capacity
		local last_refill = tonumber(bucket[2]) or now
		
		-- Calculate tokens to add based on time elapsed
		local elapsed = now - last_refill
		local tokens_to_add = math.floor(elapsed / interval * tokens)
		current_tokens = math.min(capacity, current_tokens + tokens_to_add)
		
		-- Check if request can be allowed
		local allowed = current_tokens >= 1
		if allowed then
			current_tokens = current_tokens - 1
		end
		
		-- Update bucket
		redis.call('HMSET', key, 'tokens', current_tokens, 'last_refill', now)
		redis.call('EXPIRE', key, interval * 2)
		
		return {allowed and 1 or 0, current_tokens}
	`
	
	result, err := r.redisClient.Eval(ctx, luaScript, []string{redisKey}, 
		config.BurstSize, 
		config.RequestsPerMinute, 
		config.WindowSize.Seconds(), 
		now.Unix()).Result()
	
	if err != nil {
		return false, nil, fmt.Errorf("failed to execute token bucket script: %w", err)
	}
	
	resultSlice := result.([]interface{})
	allowed := resultSlice[0].(int64) == 1
	remaining := int(resultSlice[1].(int64))
	
	info := &domain.RateLimitInfo{
		Limit:     config.RequestsPerMinute,
		Remaining: remaining,
		Reset:     now.Add(config.WindowSize),
		Window:    config.WindowSize,
		Blocked:   !allowed,
	}
	
	return allowed, info, nil
}

// Get rate limit configuration for a specific key
func (r *RateLimitService) getRateLimitConfig(key string) RateLimit {
	// Extract key type from key (e.g., "ip:192.168.1.1" -> "ip")
	keyType := "default"
	if colonIndex := len(key); colonIndex > 0 {
		for i, char := range key {
			if char == ':' {
				keyType = key[:i]
				break
			}
		}
	}
	
	// Get specific configuration or fall back to default
	if config, exists := r.config.Limits[keyType]; exists {
		return config
	}
	
	// Return default configuration
	return RateLimit{
		RequestsPerMinute: r.config.RequestsPerMinute,
		BurstSize:         r.config.BurstSize,
		WindowSize:        r.config.WindowSize,
	}
}

// Cleanup expired keys periodically
func (r *RateLimitService) cleanupExpiredKeys() {
	ticker := time.NewTicker(r.config.CleanupInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		ctx := context.Background()
		
		// Find all rate limit keys
		keys, err := r.redisClient.Keys(ctx, "rate_limit:*").Result()
		if err != nil {
			r.logger.WithError(err).Error("Failed to get rate limit keys for cleanup")
			continue
		}
		
		// Check each key and remove if expired
		for _, key := range keys {
			ttl, err := r.redisClient.TTL(ctx, key).Result()
			if err != nil {
				continue
			}
			
			// If TTL is -1 (no expiration) or very small, set a reasonable expiration
			if ttl == -1 || ttl < time.Minute {
				r.redisClient.Expire(ctx, key, r.config.WindowSize*2)
			}
		}
		
		r.logger.Debug("Rate limit cleanup completed", map[string]any{
			"keys_checked": len(keys),
		})
	}
}

// GetRateLimitStatus returns the current rate limit status for a key
func (r *RateLimitService) GetRateLimitStatus(ctx context.Context, key string) (*domain.RateLimitInfo, error) {
	if !r.config.Enabled {
		return &domain.RateLimitInfo{
			Limit:     -1,
			Remaining: -1,
			Reset:     time.Now().Add(time.Minute),
			Window:    time.Minute,
			Blocked:   false,
		}, nil
	}
	
	config := r.getRateLimitConfig(key)
	redisKey := fmt.Sprintf("rate_limit:%s", key)
	
	now := time.Now()
	windowStart := now.Add(-config.WindowSize)
	
	// Count current requests in window
	count, err := r.redisClient.ZCount(ctx, redisKey, 
		strconv.FormatInt(windowStart.UnixNano(), 10), 
		strconv.FormatInt(now.UnixNano(), 10)).Result()
	
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limit status: %w", err)
	}
	
	remaining := int64(config.RequestsPerMinute) - count
	if remaining < 0 {
		remaining = 0
	}
	
	return &domain.RateLimitInfo{
		Limit:     config.RequestsPerMinute,
		Remaining: int(remaining),
		Reset:     now.Add(config.WindowSize),
		Window:    config.WindowSize,
		Blocked:   count >= int64(config.RequestsPerMinute),
	}, nil
}

// ResetRateLimit resets the rate limit for a specific key
func (r *RateLimitService) ResetRateLimit(ctx context.Context, key string) error {
	redisKey := fmt.Sprintf("rate_limit:%s", key)
	
	err := r.redisClient.Del(ctx, redisKey).Err()
	if err != nil {
		return fmt.Errorf("failed to reset rate limit: %w", err)
	}
	
	r.logger.Info("Rate limit reset", map[string]any{
		"key": key,
	})
	
	return nil
}
