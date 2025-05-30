package ratelimiter

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HTTPMiddleware provides rate limiting middleware for HTTP requests
type HTTPMiddleware struct {
	limiter *RedisLimiter
	logger  *zap.Logger
	config  *MiddlewareConfig
}

// MiddlewareConfig contains middleware configuration
type MiddlewareConfig struct {
	DefaultRule     *RateLimitRule
	PathRules       map[string]*RateLimitRule
	KeyGenerator    func(*gin.Context) string
	ErrorHandler    func(*gin.Context, *LimitResult)
	HeadersEnabled  bool
	SkipPaths       []string
	SkipSuccessful  bool
}

// NewHTTPMiddleware creates a new HTTP rate limiting middleware
func NewHTTPMiddleware(limiter *RedisLimiter, logger *zap.Logger, config *MiddlewareConfig) *HTTPMiddleware {
	if config == nil {
		config = &MiddlewareConfig{
			DefaultRule: &RateLimitRule{
				Limit:     100,
				Window:    time.Minute,
				Algorithm: "sliding_window",
			},
			HeadersEnabled: true,
		}
	}

	if config.KeyGenerator == nil {
		config.KeyGenerator = func(c *gin.Context) string {
			// Use client IP as default key
			return c.ClientIP()
		}
	}

	if config.ErrorHandler == nil {
		config.ErrorHandler = func(c *gin.Context, result *LimitResult) {
			c.Header("X-RateLimit-Limit", strconv.FormatInt(result.Limit, 10))
			c.Header("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))
			
			if result.RetryAfter > 0 {
				c.Header("Retry-After", strconv.FormatInt(int64(result.RetryAfter.Seconds()), 10))
			}

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests",
				"code":    "RATE_LIMIT_EXCEEDED",
				"limit":   result.Limit,
				"remaining": result.Remaining,
				"reset_time": result.ResetTime,
				"retry_after": result.RetryAfter.Seconds(),
			})
		}
	}

	return &HTTPMiddleware{
		limiter: limiter,
		logger:  logger,
		config:  config,
	}
}

// Middleware returns the Gin middleware function
func (m *HTTPMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if path is in skip list
		if m.shouldSkip(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Generate rate limit key
		key := m.config.KeyGenerator(c)

		// Get rate limit rule for this path
		rule := m.getRuleForPath(c.FullPath())
		rule.Key = key

		// Check rate limit
		result, err := m.limiter.Allow(c.Request.Context(), rule)
		if err != nil {
			m.logger.Error("Rate limit check failed", 
				zap.String("key", key), 
				zap.Error(err),
			)
			c.Next() // Continue on error
			return
		}

		// Add rate limit headers if enabled
		if m.config.HeadersEnabled {
			c.Header("X-RateLimit-Limit", strconv.FormatInt(result.Limit, 10))
			c.Header("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))
		}

		// Check if request is allowed
		if !result.Allowed {
			m.logger.Info("Rate limit exceeded", 
				zap.String("key", key),
				zap.String("path", c.FullPath()),
				zap.Int64("limit", result.Limit),
			)
			m.config.ErrorHandler(c, result)
			c.Abort()
			return
		}

		// Log successful rate limit check
		m.logger.Debug("Rate limit check passed", 
			zap.String("key", key),
			zap.Int64("remaining", result.Remaining),
		)

		c.Next()
	}
}

// MiddlewareWithRule returns middleware with a specific rule
func (m *HTTPMiddleware) MiddlewareWithRule(rule *RateLimitRule) gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.shouldSkip(c.Request.URL.Path) {
			c.Next()
			return
		}

		key := m.config.KeyGenerator(c)
		ruleWithKey := *rule // Copy rule
		ruleWithKey.Key = key

		result, err := m.limiter.Allow(c.Request.Context(), &ruleWithKey)
		if err != nil {
			m.logger.Error("Rate limit check failed", zap.Error(err))
			c.Next()
			return
		}

		if m.config.HeadersEnabled {
			c.Header("X-RateLimit-Limit", strconv.FormatInt(result.Limit, 10))
			c.Header("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))
		}

		if !result.Allowed {
			m.config.ErrorHandler(c, result)
			c.Abort()
			return
		}

		c.Next()
	}
}

// PerUserMiddleware creates middleware that limits per user
func (m *HTTPMiddleware) PerUserMiddleware(getUserID func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.shouldSkip(c.Request.URL.Path) {
			c.Next()
			return
		}

		userID := getUserID(c)
		if userID == "" {
			// No user ID, use IP-based limiting
			userID = c.ClientIP()
		}

		rule := m.getRuleForPath(c.FullPath())
		rule.Key = fmt.Sprintf("user:%s", userID)

		result, err := m.limiter.Allow(c.Request.Context(), rule)
		if err != nil {
			m.logger.Error("Rate limit check failed", zap.Error(err))
			c.Next()
			return
		}

		if m.config.HeadersEnabled {
			c.Header("X-RateLimit-Limit", strconv.FormatInt(result.Limit, 10))
			c.Header("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))
		}

		if !result.Allowed {
			m.logger.Info("User rate limit exceeded", 
				zap.String("user_id", userID),
				zap.String("path", c.FullPath()),
			)
			m.config.ErrorHandler(c, result)
			c.Abort()
			return
		}

		c.Next()
	}
}

// PerAPIKeyMiddleware creates middleware that limits per API key
func (m *HTTPMiddleware) PerAPIKeyMiddleware(getAPIKey func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.shouldSkip(c.Request.URL.Path) {
			c.Next()
			return
		}

		apiKey := getAPIKey(c)
		if apiKey == "" {
			// No API key, use IP-based limiting
			apiKey = c.ClientIP()
		}

		rule := m.getRuleForPath(c.FullPath())
		rule.Key = fmt.Sprintf("api_key:%s", apiKey)

		result, err := m.limiter.Allow(c.Request.Context(), rule)
		if err != nil {
			m.logger.Error("Rate limit check failed", zap.Error(err))
			c.Next()
			return
		}

		if m.config.HeadersEnabled {
			c.Header("X-RateLimit-Limit", strconv.FormatInt(result.Limit, 10))
			c.Header("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))
		}

		if !result.Allowed {
			m.logger.Info("API key rate limit exceeded", 
				zap.String("api_key", apiKey),
				zap.String("path", c.FullPath()),
			)
			m.config.ErrorHandler(c, result)
			c.Abort()
			return
		}

		c.Next()
	}
}

// getRuleForPath returns the rate limit rule for a specific path
func (m *HTTPMiddleware) getRuleForPath(path string) *RateLimitRule {
	if m.config.PathRules != nil {
		if rule, exists := m.config.PathRules[path]; exists {
			return rule
		}
	}
	return m.config.DefaultRule
}

// shouldSkip checks if path should be skipped
func (m *HTTPMiddleware) shouldSkip(path string) bool {
	for _, skipPath := range m.config.SkipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// GetStats returns rate limiting statistics
func (m *HTTPMiddleware) GetStats() map[string]interface{} {
	// This would require tracking stats in the middleware
	// For now, return basic info
	return map[string]interface{}{
		"middleware_type": "rate_limiter",
		"default_rule":    m.config.DefaultRule,
		"path_rules":      len(m.config.PathRules),
		"headers_enabled": m.config.HeadersEnabled,
	}
}

// Adaptive Rate Limiting

// AdaptiveMiddleware provides adaptive rate limiting based on system load
type AdaptiveMiddleware struct {
	*HTTPMiddleware
	loadChecker func() float64 // Returns load factor (0.0 to 1.0)
	baseRule    *RateLimitRule
}

// NewAdaptiveMiddleware creates adaptive rate limiting middleware
func NewAdaptiveMiddleware(limiter *RedisLimiter, logger *zap.Logger, config *MiddlewareConfig, loadChecker func() float64) *AdaptiveMiddleware {
	httpMiddleware := NewHTTPMiddleware(limiter, logger, config)
	
	return &AdaptiveMiddleware{
		HTTPMiddleware: httpMiddleware,
		loadChecker:    loadChecker,
		baseRule:       config.DefaultRule,
	}
}

// AdaptiveMiddleware returns adaptive middleware function
func (am *AdaptiveMiddleware) AdaptiveMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if am.shouldSkip(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Get current system load
		load := am.loadChecker()
		
		// Adjust rate limit based on load
		rule := *am.baseRule // Copy base rule
		rule.Key = am.config.KeyGenerator(c)
		
		// Reduce limit as load increases
		if load > 0.8 {
			rule.Limit = int64(float64(rule.Limit) * 0.5) // 50% of normal limit
		} else if load > 0.6 {
			rule.Limit = int64(float64(rule.Limit) * 0.7) // 70% of normal limit
		} else if load > 0.4 {
			rule.Limit = int64(float64(rule.Limit) * 0.9) // 90% of normal limit
		}

		result, err := am.limiter.Allow(c.Request.Context(), &rule)
		if err != nil {
			am.logger.Error("Adaptive rate limit check failed", zap.Error(err))
			c.Next()
			return
		}

		if am.config.HeadersEnabled {
			c.Header("X-RateLimit-Limit", strconv.FormatInt(result.Limit, 10))
			c.Header("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))
			c.Header("X-RateLimit-Load", fmt.Sprintf("%.2f", load))
		}

		if !result.Allowed {
			am.logger.Info("Adaptive rate limit exceeded", 
				zap.String("key", rule.Key),
				zap.Float64("load", load),
				zap.Int64("adjusted_limit", rule.Limit),
			)
			am.config.ErrorHandler(c, result)
			c.Abort()
			return
		}

		c.Next()
	}
}
