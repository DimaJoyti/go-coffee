package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Limiter manages rate limiting for AI providers
type Limiter struct {
	limiters map[string]*ProviderLimiter
	mutex    sync.RWMutex
	
	// Global limits
	globalRequestLimiter *rate.Limiter
	globalTokenLimiter   *rate.Limiter
	
	// Configuration
	config *Config
}

// ProviderLimiter manages rate limiting for a specific provider
type ProviderLimiter struct {
	Provider string
	
	// Request rate limiting
	RequestLimiter *rate.Limiter
	
	// Token rate limiting
	TokenLimiter *rate.Limiter
	
	// Concurrent request limiting
	ConcurrentRequests chan struct{}
	
	// Per-user rate limiting
	UserLimiters map[string]*UserLimiter
	UserMutex    sync.RWMutex
	
	// Configuration
	Config *ProviderConfig
	
	// Statistics
	Stats *LimiterStats
}

// UserLimiter manages rate limiting for a specific user
type UserLimiter struct {
	UserID         string
	RequestLimiter *rate.Limiter
	TokenLimiter   *rate.Limiter
	LastUsed       time.Time
}

// Config holds rate limiting configuration
type Config struct {
	// Global limits
	GlobalRequestsPerSecond float64       `yaml:"global_requests_per_second" json:"global_requests_per_second"`
	GlobalTokensPerSecond   float64       `yaml:"global_tokens_per_second" json:"global_tokens_per_second"`
	
	// Provider configurations
	Providers map[string]*ProviderConfig `yaml:"providers" json:"providers"`
	
	// User rate limiting
	EnableUserLimits        bool          `yaml:"enable_user_limits" json:"enable_user_limits"`
	UserRequestsPerSecond   float64       `yaml:"user_requests_per_second" json:"user_requests_per_second"`
	UserTokensPerSecond     float64       `yaml:"user_tokens_per_second" json:"user_tokens_per_second"`
	UserLimiterCleanup      time.Duration `yaml:"user_limiter_cleanup" json:"user_limiter_cleanup"`
	
	// Burst settings
	RequestBurst            int           `yaml:"request_burst" json:"request_burst"`
	TokenBurst              int           `yaml:"token_burst" json:"token_burst"`
}

// ProviderConfig holds provider-specific rate limiting configuration
type ProviderConfig struct {
	Provider               string        `yaml:"provider" json:"provider"`
	RequestsPerSecond      float64       `yaml:"requests_per_second" json:"requests_per_second"`
	TokensPerSecond        float64       `yaml:"tokens_per_second" json:"tokens_per_second"`
	MaxConcurrentRequests  int           `yaml:"max_concurrent_requests" json:"max_concurrent_requests"`
	RequestBurst           int           `yaml:"request_burst" json:"request_burst"`
	TokenBurst             int           `yaml:"token_burst" json:"token_burst"`
	
	// Time-based limits
	RequestsPerMinute      int           `yaml:"requests_per_minute" json:"requests_per_minute"`
	RequestsPerHour        int           `yaml:"requests_per_hour" json:"requests_per_hour"`
	RequestsPerDay         int           `yaml:"requests_per_day" json:"requests_per_day"`
	
	TokensPerMinute        int           `yaml:"tokens_per_minute" json:"tokens_per_minute"`
	TokensPerHour          int           `yaml:"tokens_per_hour" json:"tokens_per_hour"`
	TokensPerDay           int           `yaml:"tokens_per_day" json:"tokens_per_day"`
	
	// Priority settings
	Priority               int           `yaml:"priority" json:"priority"`
	
	// Enabled flag
	Enabled                bool          `yaml:"enabled" json:"enabled"`
}

// LimiterStats holds statistics for a rate limiter
type LimiterStats struct {
	TotalRequests      int64     `json:"total_requests"`
	AllowedRequests    int64     `json:"allowed_requests"`
	RejectedRequests   int64     `json:"rejected_requests"`
	TotalTokens        int64     `json:"total_tokens"`
	AllowedTokens      int64     `json:"allowed_tokens"`
	RejectedTokens     int64     `json:"rejected_tokens"`
	LastRequest        time.Time `json:"last_request"`
	LastRejection      time.Time `json:"last_rejection"`
	
	// Concurrent requests
	CurrentConcurrent  int       `json:"current_concurrent"`
	MaxConcurrent      int       `json:"max_concurrent"`
	
	// Rate limiting events
	RateLimitEvents    []RateLimitEvent `json:"rate_limit_events"`
}

// RateLimitEvent represents a rate limiting event
type RateLimitEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"` // request, token, concurrent
	UserID      string    `json:"user_id,omitempty"`
	Reason      string    `json:"reason"`
	RequestedTokens int   `json:"requested_tokens,omitempty"`
}

// RateLimitError represents a rate limiting error
type RateLimitError struct {
	Provider    string        `json:"provider"`
	Type        string        `json:"type"` // request, token, concurrent
	Limit       float64       `json:"limit"`
	Current     float64       `json:"current"`
	RetryAfter  time.Duration `json:"retry_after"`
	Message     string        `json:"message"`
}

// Error implements the error interface
func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded for %s: %s (limit: %.2f, current: %.2f, retry after: %v)",
		e.Provider, e.Message, e.Limit, e.Current, e.RetryAfter)
}

// NewLimiter creates a new rate limiter
func NewLimiter(config *Config) *Limiter {
	limiter := &Limiter{
		limiters: make(map[string]*ProviderLimiter),
		config:   config,
	}
	
	// Create global limiters
	if config.GlobalRequestsPerSecond > 0 {
		limiter.globalRequestLimiter = rate.NewLimiter(
			rate.Limit(config.GlobalRequestsPerSecond),
			config.RequestBurst,
		)
	}
	
	if config.GlobalTokensPerSecond > 0 {
		limiter.globalTokenLimiter = rate.NewLimiter(
			rate.Limit(config.GlobalTokensPerSecond),
			config.TokenBurst,
		)
	}
	
	// Create provider limiters
	for provider, providerConfig := range config.Providers {
		if providerConfig.Enabled {
			limiter.createProviderLimiter(provider, providerConfig)
		}
	}
	
	// Start cleanup goroutine for user limiters
	if config.EnableUserLimits && config.UserLimiterCleanup > 0 {
		go limiter.cleanupUserLimiters()
	}
	
	return limiter
}

// CheckRequest checks if a request is allowed
func (l *Limiter) CheckRequest(ctx context.Context, provider, userID string, estimatedTokens int) error {
	// Check global limits first
	if err := l.checkGlobalLimits(ctx, estimatedTokens); err != nil {
		return err
	}
	
	// Check provider limits
	return l.checkProviderLimits(ctx, provider, userID, estimatedTokens)
}

// ReserveRequest reserves capacity for a request
func (l *Limiter) ReserveRequest(ctx context.Context, provider, userID string, estimatedTokens int) (*Reservation, error) {
	// Check if request is allowed
	if err := l.CheckRequest(ctx, provider, userID, estimatedTokens); err != nil {
		return nil, err
	}
	
	// Create reservation
	reservation := &Reservation{
		Provider:        provider,
		UserID:          userID,
		EstimatedTokens: estimatedTokens,
		Timestamp:       time.Now(),
		limiter:         l,
	}
	
	// Reserve concurrent slot if applicable
	if err := l.reserveConcurrentSlot(provider); err != nil {
		return nil, err
	}
	
	reservation.ConcurrentSlotReserved = true
	
	return reservation, nil
}

// UpdateUsage updates usage after a request completes
func (l *Limiter) UpdateUsage(provider, userID string, actualTokens int) {
	l.mutex.RLock()
	providerLimiter, exists := l.limiters[provider]
	l.mutex.RUnlock()
	
	if !exists {
		return
	}
	
	// Update statistics
	providerLimiter.Stats.TotalRequests++
	providerLimiter.Stats.AllowedRequests++
	providerLimiter.Stats.TotalTokens += int64(actualTokens)
	providerLimiter.Stats.AllowedTokens += int64(actualTokens)
	providerLimiter.Stats.LastRequest = time.Now()
}

// GetStats returns statistics for a provider
func (l *Limiter) GetStats(provider string) *LimiterStats {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	if providerLimiter, exists := l.limiters[provider]; exists {
		return providerLimiter.Stats
	}
	
	return nil
}

// GetAllStats returns statistics for all providers
func (l *Limiter) GetAllStats() map[string]*LimiterStats {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	stats := make(map[string]*LimiterStats)
	for provider, limiter := range l.limiters {
		stats[provider] = limiter.Stats
	}
	
	return stats
}

// checkGlobalLimits checks global rate limits
func (l *Limiter) checkGlobalLimits(ctx context.Context, estimatedTokens int) error {
	// Check global request limit
	if l.globalRequestLimiter != nil {
		if !l.globalRequestLimiter.Allow() {
			return &RateLimitError{
				Provider:   "global",
				Type:       "request",
				Limit:      float64(l.globalRequestLimiter.Limit()),
				Message:    "global request rate limit exceeded",
				RetryAfter: time.Second,
			}
		}
	}
	
	// Check global token limit
	if l.globalTokenLimiter != nil && estimatedTokens > 0 {
		if !l.globalTokenLimiter.AllowN(time.Now(), estimatedTokens) {
			return &RateLimitError{
				Provider:   "global",
				Type:       "token",
				Limit:      float64(l.globalTokenLimiter.Limit()),
				Current:    float64(estimatedTokens),
				Message:    "global token rate limit exceeded",
				RetryAfter: time.Second,
			}
		}
	}
	
	return nil
}

// checkProviderLimits checks provider-specific rate limits
func (l *Limiter) checkProviderLimits(ctx context.Context, provider, userID string, estimatedTokens int) error {
	l.mutex.RLock()
	providerLimiter, exists := l.limiters[provider]
	l.mutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("provider %s not configured for rate limiting", provider)
	}
	
	// Check provider request limit
	if providerLimiter.RequestLimiter != nil {
		if !providerLimiter.RequestLimiter.Allow() {
			providerLimiter.Stats.RejectedRequests++
			providerLimiter.Stats.LastRejection = time.Now()
			
			return &RateLimitError{
				Provider:   provider,
				Type:       "request",
				Limit:      float64(providerLimiter.RequestLimiter.Limit()),
				Message:    "provider request rate limit exceeded",
				RetryAfter: time.Second,
			}
		}
	}
	
	// Check provider token limit
	if providerLimiter.TokenLimiter != nil && estimatedTokens > 0 {
		if !providerLimiter.TokenLimiter.AllowN(time.Now(), estimatedTokens) {
			providerLimiter.Stats.RejectedTokens += int64(estimatedTokens)
			providerLimiter.Stats.LastRejection = time.Now()
			
			return &RateLimitError{
				Provider:   provider,
				Type:       "token",
				Limit:      float64(providerLimiter.TokenLimiter.Limit()),
				Current:    float64(estimatedTokens),
				Message:    "provider token rate limit exceeded",
				RetryAfter: time.Second,
			}
		}
	}
	
	// Check user limits if enabled
	if l.config.EnableUserLimits && userID != "" {
		if err := l.checkUserLimits(providerLimiter, userID, estimatedTokens); err != nil {
			return err
		}
	}
	
	return nil
}

// checkUserLimits checks user-specific rate limits
func (l *Limiter) checkUserLimits(providerLimiter *ProviderLimiter, userID string, estimatedTokens int) error {
	providerLimiter.UserMutex.Lock()
	defer providerLimiter.UserMutex.Unlock()
	
	// Get or create user limiter
	userLimiter, exists := providerLimiter.UserLimiters[userID]
	if !exists {
		userLimiter = &UserLimiter{
			UserID: userID,
			RequestLimiter: rate.NewLimiter(
				rate.Limit(l.config.UserRequestsPerSecond),
				l.config.RequestBurst,
			),
			TokenLimiter: rate.NewLimiter(
				rate.Limit(l.config.UserTokensPerSecond),
				l.config.TokenBurst,
			),
		}
		providerLimiter.UserLimiters[userID] = userLimiter
	}
	
	userLimiter.LastUsed = time.Now()
	
	// Check user request limit
	if !userLimiter.RequestLimiter.Allow() {
		return &RateLimitError{
			Provider:   providerLimiter.Provider,
			Type:       "user_request",
			Limit:      l.config.UserRequestsPerSecond,
			Message:    fmt.Sprintf("user %s request rate limit exceeded", userID),
			RetryAfter: time.Second,
		}
	}
	
	// Check user token limit
	if estimatedTokens > 0 && !userLimiter.TokenLimiter.AllowN(time.Now(), estimatedTokens) {
		return &RateLimitError{
			Provider:   providerLimiter.Provider,
			Type:       "user_token",
			Limit:      l.config.UserTokensPerSecond,
			Current:    float64(estimatedTokens),
			Message:    fmt.Sprintf("user %s token rate limit exceeded", userID),
			RetryAfter: time.Second,
		}
	}
	
	return nil
}

// reserveConcurrentSlot reserves a concurrent request slot
func (l *Limiter) reserveConcurrentSlot(provider string) error {
	l.mutex.RLock()
	providerLimiter, exists := l.limiters[provider]
	l.mutex.RUnlock()
	
	if !exists || providerLimiter.ConcurrentRequests == nil {
		return nil
	}
	
	select {
	case providerLimiter.ConcurrentRequests <- struct{}{}:
		providerLimiter.Stats.CurrentConcurrent++
		if providerLimiter.Stats.CurrentConcurrent > providerLimiter.Stats.MaxConcurrent {
			providerLimiter.Stats.MaxConcurrent = providerLimiter.Stats.CurrentConcurrent
		}
		return nil
	default:
		return &RateLimitError{
			Provider:   provider,
			Type:       "concurrent",
			Limit:      float64(providerLimiter.Config.MaxConcurrentRequests),
			Current:    float64(providerLimiter.Stats.CurrentConcurrent),
			Message:    "concurrent request limit exceeded",
			RetryAfter: time.Second,
		}
	}
}

// releaseConcurrentSlot releases a concurrent request slot
func (l *Limiter) releaseConcurrentSlot(provider string) {
	l.mutex.RLock()
	providerLimiter, exists := l.limiters[provider]
	l.mutex.RUnlock()
	
	if !exists || providerLimiter.ConcurrentRequests == nil {
		return
	}
	
	select {
	case <-providerLimiter.ConcurrentRequests:
		providerLimiter.Stats.CurrentConcurrent--
	default:
		// Channel is empty, nothing to release
	}
}

// createProviderLimiter creates a rate limiter for a provider
func (l *Limiter) createProviderLimiter(provider string, config *ProviderConfig) {
	providerLimiter := &ProviderLimiter{
		Provider:     provider,
		UserLimiters: make(map[string]*UserLimiter),
		Config:       config,
		Stats: &LimiterStats{
			RateLimitEvents: []RateLimitEvent{},
		},
	}
	
	// Create request limiter
	if config.RequestsPerSecond > 0 {
		providerLimiter.RequestLimiter = rate.NewLimiter(
			rate.Limit(config.RequestsPerSecond),
			config.RequestBurst,
		)
	}
	
	// Create token limiter
	if config.TokensPerSecond > 0 {
		providerLimiter.TokenLimiter = rate.NewLimiter(
			rate.Limit(config.TokensPerSecond),
			config.TokenBurst,
		)
	}
	
	// Create concurrent request limiter
	if config.MaxConcurrentRequests > 0 {
		providerLimiter.ConcurrentRequests = make(chan struct{}, config.MaxConcurrentRequests)
	}
	
	l.limiters[provider] = providerLimiter
}

// cleanupUserLimiters periodically cleans up unused user limiters
func (l *Limiter) cleanupUserLimiters() {
	ticker := time.NewTicker(l.config.UserLimiterCleanup)
	defer ticker.Stop()
	
	for range ticker.C {
		cutoff := time.Now().Add(-l.config.UserLimiterCleanup)
		
		l.mutex.RLock()
		for _, providerLimiter := range l.limiters {
			providerLimiter.UserMutex.Lock()
			for userID, userLimiter := range providerLimiter.UserLimiters {
				if userLimiter.LastUsed.Before(cutoff) {
					delete(providerLimiter.UserLimiters, userID)
				}
			}
			providerLimiter.UserMutex.Unlock()
		}
		l.mutex.RUnlock()
	}
}

// Reservation represents a rate limit reservation
type Reservation struct {
	Provider                string
	UserID                  string
	EstimatedTokens         int
	Timestamp               time.Time
	ConcurrentSlotReserved  bool
	limiter                 *Limiter
}

// Release releases the reservation
func (r *Reservation) Release() {
	if r.ConcurrentSlotReserved {
		r.limiter.releaseConcurrentSlot(r.Provider)
		r.ConcurrentSlotReserved = false
	}
}
