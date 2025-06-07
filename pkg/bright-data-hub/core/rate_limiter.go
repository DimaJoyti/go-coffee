package core

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	tokens   chan struct{}
	ticker   *time.Ticker
	capacity int
	rate     int
	mu       sync.Mutex
	closed   bool
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rps, burst int) *RateLimiter {
	rl := &RateLimiter{
		tokens:   make(chan struct{}, burst),
		capacity: burst,
		rate:     rps,
	}
	
	// Fill the bucket initially
	for i := 0; i < burst; i++ {
		rl.tokens <- struct{}{}
	}
	
	// Start the refill ticker
	if rps > 0 {
		interval := time.Second / time.Duration(rps)
		rl.ticker = time.NewTicker(interval)
		go rl.refill()
	}
	
	return rl
}

// Wait waits for a token to become available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TryAcquire tries to acquire a token without blocking
func (rl *RateLimiter) TryAcquire() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

// refill adds tokens to the bucket at the specified rate
func (rl *RateLimiter) refill() {
	for range rl.ticker.C {
		rl.mu.Lock()
		if rl.closed {
			rl.mu.Unlock()
			return
		}
		
		select {
		case rl.tokens <- struct{}{}:
		default:
			// Bucket is full
		}
		rl.mu.Unlock()
	}
}

// Close stops the rate limiter
func (rl *RateLimiter) Close() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	if !rl.closed {
		rl.closed = true
		if rl.ticker != nil {
			rl.ticker.Stop()
		}
		close(rl.tokens)
	}
}

// GetStats returns rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	return map[string]interface{}{
		"capacity":         rl.capacity,
		"rate":            rl.rate,
		"available_tokens": len(rl.tokens),
		"closed":          rl.closed,
	}
}
