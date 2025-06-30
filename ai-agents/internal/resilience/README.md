# Enhanced Error Handling & Resilience

This package provides a comprehensive resilience framework for the Go Coffee AI agents, implementing enterprise-grade error handling, circuit breakers, retry logic, timeout management, recovery strategies, and health monitoring.

## Overview

The resilience framework implements multiple patterns to handle failures gracefully and improve system reliability:

1. **Custom Error Types**: Domain-specific errors with rich context and metadata
2. **Circuit Breaker Pattern**: Prevent cascading failures and enable fast failure
3. **Retry Logic**: Intelligent retry with exponential backoff and jitter
4. **Timeout Management**: Context-based timeout handling with adaptive policies
5. **Recovery Strategies**: Automatic recovery and fallback mechanisms
6. **Health Monitoring**: Proactive health checks and failure detection

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Error Types     │    │ Circuit Breaker │    │ Retry Logic     │
│ • Custom Errors │    │ • State Machine │    │ • Backoff       │
│ • Error Codes   │    │ • Metrics       │    │ • Jitter        │
│ • Context       │    │ • Thresholds    │    │ • Conditions    │
│ • Wrapping      │    └─────────────────┘    └─────────────────┘
└─────────────────┘              │                        │
         │                       ▼                        ▼
         ▼              ┌─────────────────┐    ┌─────────────────┐
┌─────────────────┐    │ Timeout Manager │    │ Recovery        │
│ Error Handling  │    │ • Context-based │    │ • Strategies    │
│ • Categorization│    │ • Adaptive      │    │ • Fallbacks     │
│ • Severity      │    │ • Policies      │    │ • Degradation   │
│ • Recovery Info │    └─────────────────┘    └─────────────────┘
└─────────────────┘              │                        │
         │                       ▼                        ▼
         └─────────────────────▶ ┌─────────────────┐ ◀────┘
                                │ Health Monitor  │
                                │ • Probes        │
                                │ • Metrics       │
                                │ • Alerts        │
                                └─────────────────┘
```

## Components

### 1. Error Types (`errors/`)

#### Custom Error Types (`types.go`)
```go
// ResilienceError is the base error type
type ResilienceError struct {
    Code        ErrorCode     `json:"code"`
    Message     string        `json:"message"`
    Category    ErrorCategory `json:"category"`
    Severity    ErrorSeverity `json:"severity"`
    Recovery    ErrorRecovery `json:"recovery"`
    Context     ErrorContext  `json:"context"`
    Cause       error         `json:"-"`
    Details     map[string]interface{} `json:"details,omitempty"`
    Suggestions []string      `json:"suggestions,omitempty"`
}

// Usage example
err := errors.NewValidationError("email", "invalid format", "user@invalid")
if err.IsRetryable() {
    // Handle retryable error
}
```

#### Error Codes (`codes.go`)
```go
// Comprehensive error code system
const (
    // Validation errors (1000-1999)
    CodeValidationFailed    ErrorCode = "E1001"
    CodeInvalidInput       ErrorCode = "E1002"
    
    // Network errors (5000-5999)
    CodeNetworkError       ErrorCode = "E5001"
    CodeConnectionTimeout  ErrorCode = "E5003"
    
    // AI service errors (8000-8999)
    CodeAIServiceError     ErrorCode = "E8001"
    CodeAIRateLimitExceeded ErrorCode = "E8005"
)

// Get error metadata
info, exists := errors.GetErrorCodeInfo(errors.CodeNetworkError)
if exists {
    fmt.Printf("HTTP Status: %d, Retryable: %v", info.HTTPStatus, info.Recovery == errors.RecoveryRetryable)
}
```

#### Error Wrapping (`wrapping.go`)
```go
// Fluent error building
err := errors.NewError(errors.CodeDatabaseError, "Connection failed").
    WithCause(originalErr).
    WithOperation("user_lookup").
    WithComponent("database").
    WithRetryContext(3, 5).
    WithSuggestion("Check database connectivity").
    Build()

// Context-aware wrapping
err = errors.WrapWithContext(ctx, originalErr, errors.CodeNetworkError, "API call failed")

// Type-specific wrapping
dbErr := errors.WrapDatabase(err, "SELECT users")
aiErr := errors.WrapAI(err, "openai", "gpt-4")
```

### 2. Circuit Breaker (`circuit/`)

#### Circuit Breaker Implementation (`breaker.go`)
```go
// Create circuit breaker
config := circuit.Config{
    Name:                  "ai-service",
    FailureThreshold:      5,
    SuccessThreshold:      3,
    Timeout:               60 * time.Second,
    MaxConcurrentRequests: 10,
    IsFailure: func(err error) bool {
        return err != nil && !errors.IsRetryable(err)
    },
    OnStateChange: func(name string, from, to circuit.State) {
        log.Printf("Circuit %s: %s -> %s", name, from, to)
    },
}

cb := circuit.NewCircuitBreaker(config)

// Execute with circuit breaker protection
err := cb.Execute(ctx, func() error {
    return callExternalService()
})

// Execute with fallback
err = cb.ExecuteWithFallback(ctx, func() error {
    return callPrimaryService()
}, func() error {
    return callFallbackService()
})

// Check circuit state
if cb.IsOpen() {
    log.Println("Circuit is open, using fallback")
}
```

### 3. Retry Logic (`retry/`)

#### Retry Policies (`policy.go`)
```go
// Create retry policy with exponential backoff
policy := retry.NewPolicy(retry.Config{
    MaxAttempts:     5,
    BackoffStrategy: retry.NewExponentialBackoff(1*time.Second, 30*time.Second, 2.0, 0.1),
    RetryCondition:  retry.DefaultRetryCondition,
    MaxDuration:     5 * time.Minute,
    AttemptTimeout:  30 * time.Second,
    OnRetry: func(attempt int, err error, delay time.Duration) {
        log.Printf("Retry attempt %d after %v: %v", attempt, delay, err)
    },
})

// Execute with retry
err := policy.Execute(ctx, func() error {
    return callUnreliableService()
})

// Execute with result
result, err := policy.ExecuteWithResult(ctx, func() (interface{}, error) {
    return fetchDataFromAPI()
})
```

#### Backoff Strategies
```go
// Fixed backoff
fixed := retry.NewFixedBackoff(5 * time.Second)

// Linear backoff
linear := retry.NewLinearBackoff(1*time.Second, 30*time.Second, 2*time.Second)

// Exponential backoff with jitter
exponential := retry.NewExponentialBackoff(1*time.Second, 30*time.Second, 2.0, 0.1)

// Decorrelated jitter (AWS style)
decorrelated := retry.NewDecorrelatedJitterBackoff(1*time.Second, 30*time.Second)

// Convenience functions
err := retry.Do(ctx, 3, func() error {
    return callService()
})

result, err := retry.DoWithResult(ctx, 5, func() (interface{}, error) {
    return fetchData()
})
```

### 4. Timeout Management (`timeout/`)

#### Timeout Manager (`manager.go`)
```go
// Initialize timeout manager
config := timeout.TimeoutConfig{
    Default: 30 * time.Second,
    Operations: map[string]time.Duration{
        "ai_generation": 2 * time.Minute,
        "database_query": 10 * time.Second,
    },
    Components: map[string]time.Duration{
        "openai": 90 * time.Second,
        "database": 15 * time.Second,
    },
    GracePeriod: 5 * time.Second,
}

tm := timeout.NewManager(config)

// Execute with timeout
err := tm.Execute(ctx, "ai_generation", "openai", func(ctx context.Context) error {
    return generateWithAI(ctx)
})

// Execute with result
result, err := tm.ExecuteWithResult(ctx, "database_query", "postgres", func(ctx context.Context) (interface{}, error) {
    return queryDatabase(ctx)
})
```

### 5. Recovery Strategies (`recovery/`)

#### Recovery Manager (`strategies.go`)
```go
// Create recovery manager
config := recovery.RecoveryConfig{
    MaxAttempts:   3,
    RecoveryDelay: 5 * time.Second,
    AutoRecovery:  true,
    OnRecoverySuccess: func(strategy string, attempt int) {
        log.Printf("Recovery successful with %s on attempt %d", strategy, attempt)
    },
}

rm := recovery.NewRecoveryManager(config)

// Register recovery strategies
rm.RegisterStrategy(recovery.NewRestartStrategy("service-restart", func(ctx context.Context) error {
    return restartService(ctx)
}, func(err error) bool {
    return errors.GetErrorCategory(err) == errors.CategoryInternal
}))

// Attempt recovery
err := rm.Recover(ctx, originalError)
```

### 6. Health Monitoring (`health/`)

#### Health Checker (`checker.go`)
```go
// Create health checker
config := health.HealthConfig{
    CheckInterval:  30 * time.Second,
    DefaultTimeout: 10 * time.Second,
    Parallel:       true,
    MaxConcurrent:  10,
    CacheResults:   true,
    CacheTTL:       60 * time.Second,
}

hc := health.NewHealthChecker(config)

// Register health checks
hc.RegisterCheck(health.NewDatabaseHealthCheck("postgres", func(ctx context.Context) error {
    return db.PingContext(ctx)
}, 5*time.Second))

// Start periodic health checking
go hc.Start(ctx)

// Check all health
results := hc.CheckAll(ctx)
for name, result := range results {
    fmt.Printf("%s: %s (%v)\n", name, result.Status, result.Duration)
}

// Get overall status
status := hc.GetOverallStatus()
fmt.Printf("Overall health: %s\n", status)
```

## Usage Examples

### Basic Error Handling
```go
package main

import (
    "context"
    "fmt"
    
    "go-coffee-ai-agents/internal/resilience/errors"
    "go-coffee-ai-agents/internal/resilience/retry"
    "go-coffee-ai-agents/internal/resilience/circuit"
)

func main() {
    ctx := context.Background()
    
    // Create circuit breaker
    cb := circuit.GetOrCreateGlobal("api-service", circuit.DefaultConfig("api-service"))
    
    // Create retry policy
    retryPolicy := retry.WithExponentialBackoff(3, 1*time.Second, 30*time.Second)
    
    // Execute with resilience
    err := retryPolicy.Execute(ctx, func() error {
        return cb.Execute(ctx, func() error {
            return callExternalAPI()
        })
    })
    
    if err != nil {
        if resErr, ok := err.(*errors.ResilienceError); ok {
            fmt.Printf("Error: %s (Code: %s, Category: %s)\n", 
                resErr.Message, resErr.Code, resErr.Category)
            
            if resErr.IsRetryable() {
                fmt.Println("Error is retryable")
            }
            
            for _, suggestion := range resErr.Suggestions {
                fmt.Printf("Suggestion: %s\n", suggestion)
            }
        }
    }
}

func callExternalAPI() error {
    // Simulate API call that might fail
    return errors.NewNetworkError("api.example.com", fmt.Errorf("connection refused"))
}
```

This Enhanced Error Handling & Resilience framework provides a robust foundation for building reliable, fault-tolerant systems that can gracefully handle failures and maintain service availability even under adverse conditions.
