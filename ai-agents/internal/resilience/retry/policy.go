package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"go-coffee-ai-agents/internal/resilience/errors"
)

// BackoffStrategy defines how to calculate retry delays
type BackoffStrategy interface {
	// NextDelay calculates the delay before the next retry attempt
	NextDelay(attempt int) time.Duration
	
	// Reset resets the backoff strategy state
	Reset()
	
	// String returns a string representation of the strategy
	String() string
}

// RetryCondition determines whether an error should trigger a retry
type RetryCondition func(error) bool

// Config holds retry policy configuration
type Config struct {
	// Maximum number of retry attempts (0 means no retries)
	MaxAttempts int
	
	// Backoff strategy for calculating delays
	BackoffStrategy BackoffStrategy
	
	// Condition to determine if an error should be retried
	RetryCondition RetryCondition
	
	// Maximum total time to spend on retries
	MaxDuration time.Duration
	
	// Context timeout for each individual attempt
	AttemptTimeout time.Duration
	
	// Callbacks
	OnRetry func(attempt int, err error, delay time.Duration)
	OnGiveUp func(attempt int, err error)
}

// DefaultConfig returns a default retry configuration
func DefaultConfig() Config {
	return Config{
		MaxAttempts:     3,
		BackoffStrategy: NewExponentialBackoff(1*time.Second, 30*time.Second, 2.0, 0.1),
		RetryCondition:  DefaultRetryCondition,
		MaxDuration:     5 * time.Minute,
		AttemptTimeout:  30 * time.Second,
	}
}

// DefaultRetryCondition is the default condition for retrying
func DefaultRetryCondition(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a resilience error
	if resErr, ok := err.(*errors.ResilienceError); ok {
		return resErr.IsRetryable()
	}

	// Check for context errors
	if err == context.Canceled {
		return false // Don't retry cancelled operations
	}
	if err == context.DeadlineExceeded {
		return true // Retry timeout errors
	}

	// Check for specific error types that are typically retryable
	return errors.IsRetryable(err)
}

// Policy implements retry logic with configurable backoff and conditions
type Policy struct {
	config Config
}

// NewPolicy creates a new retry policy
func NewPolicy(config Config) *Policy {
	if config.MaxAttempts < 0 {
		config.MaxAttempts = 0
	}
	if config.BackoffStrategy == nil {
		config.BackoffStrategy = NewExponentialBackoff(1*time.Second, 30*time.Second, 2.0, 0.1)
	}
	if config.RetryCondition == nil {
		config.RetryCondition = DefaultRetryCondition
	}
	if config.MaxDuration <= 0 {
		config.MaxDuration = 5 * time.Minute
	}

	return &Policy{config: config}
}

// Execute executes a function with retry logic
func (p *Policy) Execute(ctx context.Context, fn func() error) error {
	if p.config.MaxAttempts == 0 {
		return fn()
	}

	var lastErr error
	startTime := time.Now()
	
	// Reset backoff strategy
	p.config.BackoffStrategy.Reset()

	for attempt := 1; attempt <= p.config.MaxAttempts+1; attempt++ {
		// Check if we've exceeded the maximum duration
		if time.Since(startTime) > p.config.MaxDuration {
			break
		}

		// Create context for this attempt with timeout if configured
		var err error
		if p.config.AttemptTimeout > 0 {
			attemptCtx, cancel := context.WithTimeout(ctx, p.config.AttemptTimeout)
			defer cancel()
			
			// Execute the function with timeout
			done := make(chan error, 1)
			go func() {
				done <- fn()
			}()
			
			select {
			case err = <-done:
				// Function completed
			case <-attemptCtx.Done():
				err = attemptCtx.Err()
			}
		} else {
			// Execute the function without timeout
			err = fn()
		}
		
		// If successful, return immediately
		if err == nil {
			return nil
		}

		lastErr = err

		// If this is the last attempt, don't retry
		if attempt > p.config.MaxAttempts {
			break
		}

		// Check if we should retry this error
		if !p.config.RetryCondition(err) {
			break
		}

		// Calculate delay for next attempt
		delay := p.config.BackoffStrategy.NextDelay(attempt - 1)

		// Call retry callback
		if p.config.OnRetry != nil {
			p.config.OnRetry(attempt, err, delay)
		}

		// Wait for the delay or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	// Call give up callback
	if p.config.OnGiveUp != nil {
		p.config.OnGiveUp(p.config.MaxAttempts, lastErr)
	}

	// Wrap the final error with retry context
	if resErr, ok := lastErr.(*errors.ResilienceError); ok {
		// Update the context with retry information
		ctx := resErr.Context
		ctx.AttemptCount = p.config.MaxAttempts
		ctx.MaxAttempts = p.config.MaxAttempts
		return resErr.WithContext(ctx)
	}

	// Create a new error with retry context
	return &errors.ResilienceError{
		Code:     errors.CodeInternalError,
		Message:  "Operation failed after retries",
		Category: errors.CategoryInternal,
		Severity: errors.SeverityError,
		Recovery: errors.RecoveryNonRetryable,
		Context: errors.ErrorContext{
			AttemptCount: p.config.MaxAttempts,
			MaxAttempts:  p.config.MaxAttempts,
			Timestamp:    time.Now(),
		},
		Cause: lastErr,
	}
}

// ExecuteWithResult executes a function that returns a result and error
func (p *Policy) ExecuteWithResult(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	var result interface{}
	err := p.Execute(ctx, func() error {
		var err error
		result, err = fn()
		return err
	})
	return result, err
}

// FixedBackoff implements a fixed delay backoff strategy
type FixedBackoff struct {
	delay time.Duration
}

// NewFixedBackoff creates a new fixed backoff strategy
func NewFixedBackoff(delay time.Duration) *FixedBackoff {
	return &FixedBackoff{delay: delay}
}

// NextDelay returns the fixed delay
func (fb *FixedBackoff) NextDelay(attempt int) time.Duration {
	return fb.delay
}

// Reset does nothing for fixed backoff
func (fb *FixedBackoff) Reset() {}

// String returns a string representation
func (fb *FixedBackoff) String() string {
	return fmt.Sprintf("FixedBackoff{delay=%v}", fb.delay)
}

// LinearBackoff implements a linear backoff strategy
type LinearBackoff struct {
	baseDelay time.Duration
	maxDelay  time.Duration
	increment time.Duration
}

// NewLinearBackoff creates a new linear backoff strategy
func NewLinearBackoff(baseDelay, maxDelay, increment time.Duration) *LinearBackoff {
	return &LinearBackoff{
		baseDelay: baseDelay,
		maxDelay:  maxDelay,
		increment: increment,
	}
}

// NextDelay calculates the linear delay
func (lb *LinearBackoff) NextDelay(attempt int) time.Duration {
	delay := lb.baseDelay + time.Duration(attempt)*lb.increment
	if delay > lb.maxDelay {
		delay = lb.maxDelay
	}
	return delay
}

// Reset does nothing for linear backoff
func (lb *LinearBackoff) Reset() {}

// String returns a string representation
func (lb *LinearBackoff) String() string {
	return fmt.Sprintf("LinearBackoff{base=%v, max=%v, increment=%v}", 
		lb.baseDelay, lb.maxDelay, lb.increment)
}

// ExponentialBackoff implements an exponential backoff strategy with jitter
type ExponentialBackoff struct {
	baseDelay  time.Duration
	maxDelay   time.Duration
	multiplier float64
	jitter     float64
	random     *rand.Rand
}

// NewExponentialBackoff creates a new exponential backoff strategy
func NewExponentialBackoff(baseDelay, maxDelay time.Duration, multiplier, jitter float64) *ExponentialBackoff {
	return &ExponentialBackoff{
		baseDelay:  baseDelay,
		maxDelay:   maxDelay,
		multiplier: multiplier,
		jitter:     jitter,
		random:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NextDelay calculates the exponential delay with jitter
func (eb *ExponentialBackoff) NextDelay(attempt int) time.Duration {
	// Calculate exponential delay
	delay := float64(eb.baseDelay) * math.Pow(eb.multiplier, float64(attempt))
	
	// Apply maximum delay limit
	if delay > float64(eb.maxDelay) {
		delay = float64(eb.maxDelay)
	}
	
	// Apply jitter
	if eb.jitter > 0 {
		jitterRange := delay * eb.jitter
		jitterOffset := (eb.random.Float64() - 0.5) * 2 * jitterRange
		delay += jitterOffset
		
		// Ensure delay is not negative
		if delay < 0 {
			delay = float64(eb.baseDelay)
		}
	}
	
	return time.Duration(delay)
}

// Reset resets the random seed
func (eb *ExponentialBackoff) Reset() {
	eb.random = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// String returns a string representation
func (eb *ExponentialBackoff) String() string {
	return fmt.Sprintf("ExponentialBackoff{base=%v, max=%v, multiplier=%.2f, jitter=%.2f}", 
		eb.baseDelay, eb.maxDelay, eb.multiplier, eb.jitter)
}

// DecorrelatedJitterBackoff implements AWS's decorrelated jitter backoff
type DecorrelatedJitterBackoff struct {
	baseDelay time.Duration
	maxDelay  time.Duration
	lastDelay time.Duration
	random    *rand.Rand
}

// NewDecorrelatedJitterBackoff creates a new decorrelated jitter backoff strategy
func NewDecorrelatedJitterBackoff(baseDelay, maxDelay time.Duration) *DecorrelatedJitterBackoff {
	return &DecorrelatedJitterBackoff{
		baseDelay: baseDelay,
		maxDelay:  maxDelay,
		lastDelay: baseDelay,
		random:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NextDelay calculates the decorrelated jitter delay
func (djb *DecorrelatedJitterBackoff) NextDelay(attempt int) time.Duration {
	// Decorrelated jitter: delay = random(base_delay, last_delay * 3)
	minDelay := float64(djb.baseDelay)
	maxDelayCandidate := float64(djb.lastDelay) * 3
	
	if maxDelayCandidate > float64(djb.maxDelay) {
		maxDelayCandidate = float64(djb.maxDelay)
	}
	
	if maxDelayCandidate < minDelay {
		maxDelayCandidate = minDelay
	}
	
	delay := minDelay + djb.random.Float64()*(maxDelayCandidate-minDelay)
	djb.lastDelay = time.Duration(delay)
	
	return djb.lastDelay
}

// Reset resets the backoff state
func (djb *DecorrelatedJitterBackoff) Reset() {
	djb.lastDelay = djb.baseDelay
	djb.random = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// String returns a string representation
func (djb *DecorrelatedJitterBackoff) String() string {
	return fmt.Sprintf("DecorrelatedJitterBackoff{base=%v, max=%v}", 
		djb.baseDelay, djb.maxDelay)
}

// Convenience functions for common retry patterns

// WithFixedBackoff creates a retry policy with fixed backoff
func WithFixedBackoff(maxAttempts int, delay time.Duration) *Policy {
	config := DefaultConfig()
	config.MaxAttempts = maxAttempts
	config.BackoffStrategy = NewFixedBackoff(delay)
	return NewPolicy(config)
}

// WithExponentialBackoff creates a retry policy with exponential backoff
func WithExponentialBackoff(maxAttempts int, baseDelay, maxDelay time.Duration) *Policy {
	config := DefaultConfig()
	config.MaxAttempts = maxAttempts
	config.BackoffStrategy = NewExponentialBackoff(baseDelay, maxDelay, 2.0, 0.1)
	return NewPolicy(config)
}

// WithLinearBackoff creates a retry policy with linear backoff
func WithLinearBackoff(maxAttempts int, baseDelay, increment, maxDelay time.Duration) *Policy {
	config := DefaultConfig()
	config.MaxAttempts = maxAttempts
	config.BackoffStrategy = NewLinearBackoff(baseDelay, maxDelay, increment)
	return NewPolicy(config)
}

// Do is a convenience function for simple retry with exponential backoff
func Do(ctx context.Context, maxAttempts int, fn func() error) error {
	policy := WithExponentialBackoff(maxAttempts, 1*time.Second, 30*time.Second)
	return policy.Execute(ctx, fn)
}

// DoWithResult is a convenience function for simple retry with result
func DoWithResult(ctx context.Context, maxAttempts int, fn func() (interface{}, error)) (interface{}, error) {
	policy := WithExponentialBackoff(maxAttempts, 1*time.Second, 30*time.Second)
	return policy.ExecuteWithResult(ctx, fn)
}
