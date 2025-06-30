package resilience

import (
	"context"
	"sync"
	"time"

	"go-coffee-ai-agents/internal/errors"
)

// RateLimiterType represents different rate limiting algorithms
type RateLimiterType string

const (
	TokenBucket   RateLimiterType = "token_bucket"
	SlidingWindow RateLimiterType = "sliding_window"
	FixedWindow   RateLimiterType = "fixed_window"
)

// RateLimiterConfig configures rate limiting behavior
type RateLimiterConfig struct {
	Type        RateLimiterType `yaml:"type"`
	Rate        int             `yaml:"rate"`          // requests per window
	Window      time.Duration   `yaml:"window"`        // time window
	BurstSize   int             `yaml:"burst_size"`    // max burst size for token bucket
	MaxWaitTime time.Duration   `yaml:"max_wait_time"` // max time to wait for token
}

// DefaultRateLimiterConfig returns a default rate limiter configuration
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		Type:        TokenBucket,
		Rate:        100,
		Window:      time.Minute,
		BurstSize:   10,
		MaxWaitTime: 5 * time.Second,
	}
}

// RateLimiter interface defines rate limiting operations
type RateLimiter interface {
	Allow() bool
	Wait(ctx context.Context) error
	Reserve() Reservation
	GetMetrics() RateLimiterMetrics
}

// Reservation represents a rate limiter reservation
type Reservation struct {
	OK        bool
	Delay     time.Duration
	TimeToAct time.Time
}

// RateLimiterMetrics tracks rate limiter metrics
type RateLimiterMetrics struct {
	TotalRequests   int64
	AllowedRequests int64
	RejectedRequests int64
	WaitingRequests int64
	CurrentTokens   int
	LastRefill      time.Time
}

// TokenBucketLimiter implements token bucket rate limiting
type TokenBucketLimiter struct {
	config       RateLimiterConfig
	tokens       int
	lastRefill   time.Time
	mutex        sync.Mutex
	metrics      RateLimiterMetrics
	logger       Logger
}

// NewTokenBucketLimiter creates a new token bucket rate limiter
func NewTokenBucketLimiter(config RateLimiterConfig, logger Logger) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		config:     config,
		tokens:     config.BurstSize,
		lastRefill: time.Now(),
		metrics: RateLimiterMetrics{
			CurrentTokens: config.BurstSize,
			LastRefill:    time.Now(),
		},
		logger: logger,
	}
}

// Allow checks if a request is allowed
func (tbl *TokenBucketLimiter) Allow() bool {
	tbl.mutex.Lock()
	defer tbl.mutex.Unlock()
	
	tbl.refillTokens()
	tbl.metrics.TotalRequests++
	
	if tbl.tokens > 0 {
		tbl.tokens--
		tbl.metrics.CurrentTokens = tbl.tokens
		tbl.metrics.AllowedRequests++
		return true
	}
	
	tbl.metrics.RejectedRequests++
	return false
}

// Wait waits for a token to become available
func (tbl *TokenBucketLimiter) Wait(ctx context.Context) error {
	reservation := tbl.Reserve()
	if !reservation.OK {
		return errors.NewRateLimitError("token_bucket", tbl.config.MaxWaitTime)
	}
	
	if reservation.Delay <= 0 {
		return nil
	}
	
	if reservation.Delay > tbl.config.MaxWaitTime {
		return errors.NewRateLimitError("token_bucket", reservation.Delay)
	}
	
	tbl.mutex.Lock()
	tbl.metrics.WaitingRequests++
	tbl.mutex.Unlock()
	
	timer := time.NewTimer(reservation.Delay)
	defer timer.Stop()
	
	select {
	case <-timer.C:
		tbl.mutex.Lock()
		tbl.metrics.WaitingRequests--
		tbl.mutex.Unlock()
		return nil
	case <-ctx.Done():
		tbl.mutex.Lock()
		tbl.metrics.WaitingRequests--
		tbl.mutex.Unlock()
		return ctx.Err()
	}
}

// Reserve reserves a token
func (tbl *TokenBucketLimiter) Reserve() Reservation {
	tbl.mutex.Lock()
	defer tbl.mutex.Unlock()
	
	tbl.refillTokens()
	
	if tbl.tokens > 0 {
		tbl.tokens--
		tbl.metrics.CurrentTokens = tbl.tokens
		return Reservation{
			OK:        true,
			Delay:     0,
			TimeToAct: time.Now(),
		}
	}
	
	// Calculate delay until next token is available
	tokensPerSecond := float64(tbl.config.Rate) / tbl.config.Window.Seconds()
	delay := time.Duration(float64(time.Second) / tokensPerSecond)
	
	return Reservation{
		OK:        true,
		Delay:     delay,
		TimeToAct: time.Now().Add(delay),
	}
}

// GetMetrics returns current metrics
func (tbl *TokenBucketLimiter) GetMetrics() RateLimiterMetrics {
	tbl.mutex.Lock()
	defer tbl.mutex.Unlock()
	return tbl.metrics
}

// refillTokens refills the token bucket based on elapsed time
func (tbl *TokenBucketLimiter) refillTokens() {
	now := time.Now()
	elapsed := now.Sub(tbl.lastRefill)
	
	if elapsed <= 0 {
		return
	}
	
	// Calculate tokens to add based on rate
	tokensPerSecond := float64(tbl.config.Rate) / tbl.config.Window.Seconds()
	tokensToAdd := int(tokensPerSecond * elapsed.Seconds())
	
	if tokensToAdd > 0 {
		tbl.tokens = min(tbl.tokens+tokensToAdd, tbl.config.BurstSize)
		tbl.lastRefill = now
		tbl.metrics.CurrentTokens = tbl.tokens
		tbl.metrics.LastRefill = now
	}
}

// SlidingWindowLimiter implements sliding window rate limiting
type SlidingWindowLimiter struct {
	config    RateLimiterConfig
	requests  []time.Time
	mutex     sync.Mutex
	metrics   RateLimiterMetrics
	logger    Logger
}

// NewSlidingWindowLimiter creates a new sliding window rate limiter
func NewSlidingWindowLimiter(config RateLimiterConfig, logger Logger) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		config:   config,
		requests: make([]time.Time, 0),
		logger:   logger,
	}
}

// Allow checks if a request is allowed
func (swl *SlidingWindowLimiter) Allow() bool {
	swl.mutex.Lock()
	defer swl.mutex.Unlock()
	
	now := time.Now()
	swl.cleanOldRequests(now)
	swl.metrics.TotalRequests++
	
	if len(swl.requests) < swl.config.Rate {
		swl.requests = append(swl.requests, now)
		swl.metrics.AllowedRequests++
		return true
	}
	
	swl.metrics.RejectedRequests++
	return false
}

// Wait waits for the sliding window to allow a request
func (swl *SlidingWindowLimiter) Wait(ctx context.Context) error {
	for {
		if swl.Allow() {
			return nil
		}
		
		swl.mutex.Lock()
		if len(swl.requests) == 0 {
			swl.mutex.Unlock()
			return nil
		}
		
		// Wait until the oldest request expires
		oldestRequest := swl.requests[0]
		waitTime := swl.config.Window - time.Since(oldestRequest)
		swl.mutex.Unlock()
		
		if waitTime > swl.config.MaxWaitTime {
			return errors.NewRateLimitError("sliding_window", waitTime)
		}
		
		if waitTime <= 0 {
			continue
		}
		
		swl.mutex.Lock()
		swl.metrics.WaitingRequests++
		swl.mutex.Unlock()
		
		timer := time.NewTimer(waitTime)
		defer timer.Stop()
		
		select {
		case <-timer.C:
			swl.mutex.Lock()
			swl.metrics.WaitingRequests--
			swl.mutex.Unlock()
			continue
		case <-ctx.Done():
			swl.mutex.Lock()
			swl.metrics.WaitingRequests--
			swl.mutex.Unlock()
			return ctx.Err()
		}
	}
}

// Reserve reserves a slot in the sliding window
func (swl *SlidingWindowLimiter) Reserve() Reservation {
	swl.mutex.Lock()
	defer swl.mutex.Unlock()
	
	now := time.Now()
	swl.cleanOldRequests(now)
	
	if len(swl.requests) < swl.config.Rate {
		return Reservation{
			OK:        true,
			Delay:     0,
			TimeToAct: now,
		}
	}
	
	// Calculate delay until oldest request expires
	oldestRequest := swl.requests[0]
	delay := swl.config.Window - time.Since(oldestRequest)
	
	return Reservation{
		OK:        true,
		Delay:     delay,
		TimeToAct: now.Add(delay),
	}
}

// GetMetrics returns current metrics
func (swl *SlidingWindowLimiter) GetMetrics() RateLimiterMetrics {
	swl.mutex.Lock()
	defer swl.mutex.Unlock()
	return swl.metrics
}

// cleanOldRequests removes requests outside the sliding window
func (swl *SlidingWindowLimiter) cleanOldRequests(now time.Time) {
	cutoff := now.Add(-swl.config.Window)
	
	// Find the first request within the window
	i := 0
	for i < len(swl.requests) && swl.requests[i].Before(cutoff) {
		i++
	}
	
	// Remove old requests
	if i > 0 {
		swl.requests = swl.requests[i:]
	}
}

// RateLimiterManager manages multiple rate limiters
type RateLimiterManager struct {
	limiters map[string]RateLimiter
	mutex    sync.RWMutex
	logger   Logger
}

// NewRateLimiterManager creates a new rate limiter manager
func NewRateLimiterManager(logger Logger) *RateLimiterManager {
	return &RateLimiterManager{
		limiters: make(map[string]RateLimiter),
		logger:   logger,
	}
}

// GetOrCreate gets an existing rate limiter or creates a new one
func (rlm *RateLimiterManager) GetOrCreate(name string, config RateLimiterConfig) RateLimiter {
	rlm.mutex.Lock()
	defer rlm.mutex.Unlock()
	
	if limiter, exists := rlm.limiters[name]; exists {
		return limiter
	}
	
	var limiter RateLimiter
	switch config.Type {
	case TokenBucket:
		limiter = NewTokenBucketLimiter(config, rlm.logger)
	case SlidingWindow:
		limiter = NewSlidingWindowLimiter(config, rlm.logger)
	default:
		limiter = NewTokenBucketLimiter(config, rlm.logger)
	}
	
	rlm.limiters[name] = limiter
	rlm.logger.Info("Created new rate limiter", "name", name, "type", config.Type)
	
	return limiter
}

// Get gets an existing rate limiter
func (rlm *RateLimiterManager) Get(name string) (RateLimiter, bool) {
	rlm.mutex.RLock()
	defer rlm.mutex.RUnlock()
	
	limiter, exists := rlm.limiters[name]
	return limiter, exists
}

// GetMetrics returns metrics for all rate limiters
func (rlm *RateLimiterManager) GetMetrics() map[string]RateLimiterMetrics {
	rlm.mutex.RLock()
	defer rlm.mutex.RUnlock()
	
	result := make(map[string]RateLimiterMetrics)
	for name, limiter := range rlm.limiters {
		result[name] = limiter.GetMetrics()
	}
	
	return result
}

// Predefined rate limiter configurations
var PredefinedRateLimiters = map[string]RateLimiterConfig{
	"ai_provider": {
		Type:        TokenBucket,
		Rate:        60,
		Window:      time.Minute,
		BurstSize:   10,
		MaxWaitTime: 30 * time.Second,
	},
	"external_api": {
		Type:        TokenBucket,
		Rate:        100,
		Window:      time.Minute,
		BurstSize:   20,
		MaxWaitTime: 10 * time.Second,
	},
	"database": {
		Type:        SlidingWindow,
		Rate:        1000,
		Window:      time.Minute,
		MaxWaitTime: 1 * time.Second,
	},
	"kafka": {
		Type:        TokenBucket,
		Rate:        500,
		Window:      time.Minute,
		BurstSize:   50,
		MaxWaitTime: 5 * time.Second,
	},
}

// Global rate limiter manager instance
var globalRLManager *RateLimiterManager
var rlManagerOnce sync.Once

// GetGlobalRateLimiterManager returns the global rate limiter manager
func GetGlobalRateLimiterManager(logger Logger) *RateLimiterManager {
	rlManagerOnce.Do(func() {
		globalRLManager = NewRateLimiterManager(logger)
	})
	return globalRLManager
}

// Utility function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
