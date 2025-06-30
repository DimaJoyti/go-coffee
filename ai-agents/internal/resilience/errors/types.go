package errors

import (
	"fmt"
	"time"
)

// ErrorCategory represents the category of an error
type ErrorCategory string

const (
	// Core error categories
	CategoryValidation    ErrorCategory = "validation"
	CategoryAuthentication ErrorCategory = "authentication"
	CategoryAuthorization ErrorCategory = "authorization"
	CategoryNotFound      ErrorCategory = "not_found"
	CategoryConflict      ErrorCategory = "conflict"
	CategoryRateLimit     ErrorCategory = "rate_limit"
	
	// Infrastructure error categories
	CategoryNetwork       ErrorCategory = "network"
	CategoryDatabase      ErrorCategory = "database"
	CategoryMessaging     ErrorCategory = "messaging"
	CategoryStorage       ErrorCategory = "storage"
	CategoryCache         ErrorCategory = "cache"
	
	// Service error categories
	CategoryAI            ErrorCategory = "ai"
	CategoryExternal      ErrorCategory = "external"
	CategoryInternal      ErrorCategory = "internal"
	CategoryTimeout       ErrorCategory = "timeout"
	CategoryCircuitBreaker ErrorCategory = "circuit_breaker"
	
	// Business logic error categories
	CategoryBusiness      ErrorCategory = "business"
	CategoryWorkflow      ErrorCategory = "workflow"
	CategoryResource      ErrorCategory = "resource"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity string

const (
	SeverityInfo     ErrorSeverity = "info"
	SeverityWarning  ErrorSeverity = "warning"
	SeverityError    ErrorSeverity = "error"
	SeverityCritical ErrorSeverity = "critical"
)

// ErrorRecovery indicates whether an error is recoverable
type ErrorRecovery string

const (
	RecoveryRetryable    ErrorRecovery = "retryable"
	RecoveryNonRetryable ErrorRecovery = "non_retryable"
	RecoveryFallback     ErrorRecovery = "fallback"
	RecoveryCircuitBreak ErrorRecovery = "circuit_break"
)

// ErrorContext provides rich context about an error
type ErrorContext struct {
	// Operation context
	Operation   string            `json:"operation"`
	Component   string            `json:"component"`
	Resource    string            `json:"resource,omitempty"`
	ResourceID  string            `json:"resource_id,omitempty"`
	
	// Request context
	RequestID   string            `json:"request_id,omitempty"`
	UserID      string            `json:"user_id,omitempty"`
	TraceID     string            `json:"trace_id,omitempty"`
	SpanID      string            `json:"span_id,omitempty"`
	
	// Error metadata
	Timestamp   time.Time         `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	
	// Retry context
	AttemptCount int              `json:"attempt_count,omitempty"`
	MaxAttempts  int              `json:"max_attempts,omitempty"`
	
	// Performance context
	Duration    time.Duration     `json:"duration,omitempty"`
	Timeout     time.Duration     `json:"timeout,omitempty"`
}

// ResilienceError is the base error type for all resilience-aware errors
type ResilienceError struct {
	// Core error information
	Code        ErrorCode     `json:"code"`
	Message     string        `json:"message"`
	Category    ErrorCategory `json:"category"`
	Severity    ErrorSeverity `json:"severity"`
	Recovery    ErrorRecovery `json:"recovery"`
	
	// Error context
	Context     ErrorContext  `json:"context"`
	
	// Error chain
	Cause       error         `json:"-"`
	
	// Additional details
	Details     map[string]interface{} `json:"details,omitempty"`
	
	// Suggestions for resolution
	Suggestions []string      `json:"suggestions,omitempty"`
}

// Error implements the error interface
func (e *ResilienceError) Error() string {
	if e.Context.Operation != "" {
		return fmt.Sprintf("[%s] %s: %s (operation: %s)", e.Code, e.Category, e.Message, e.Context.Operation)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Category, e.Message)
}

// Unwrap returns the underlying cause error
func (e *ResilienceError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target error
func (e *ResilienceError) Is(target error) bool {
	if t, ok := target.(*ResilienceError); ok {
		return e.Code == t.Code
	}
	return false
}

// IsCategory checks if the error belongs to a specific category
func (e *ResilienceError) IsCategory(category ErrorCategory) bool {
	return e.Category == category
}

// IsSeverity checks if the error has a specific severity
func (e *ResilienceError) IsSeverity(severity ErrorSeverity) bool {
	return e.Severity == severity
}

// IsRecoverable checks if the error is recoverable
func (e *ResilienceError) IsRecoverable() bool {
	return e.Recovery == RecoveryRetryable || e.Recovery == RecoveryFallback
}

// IsRetryable checks if the error should trigger a retry
func (e *ResilienceError) IsRetryable() bool {
	return e.Recovery == RecoveryRetryable
}

// ShouldCircuitBreak checks if the error should trigger circuit breaking
func (e *ResilienceError) ShouldCircuitBreak() bool {
	return e.Recovery == RecoveryCircuitBreak || e.Severity == SeverityCritical
}

// WithContext adds context to the error
func (e *ResilienceError) WithContext(ctx ErrorContext) *ResilienceError {
	newErr := *e
	newErr.Context = ctx
	return &newErr
}

// WithCause adds a cause to the error
func (e *ResilienceError) WithCause(cause error) *ResilienceError {
	newErr := *e
	newErr.Cause = cause
	return &newErr
}

// WithDetail adds a detail to the error
func (e *ResilienceError) WithDetail(key string, value interface{}) *ResilienceError {
	newErr := *e
	if newErr.Details == nil {
		newErr.Details = make(map[string]interface{})
	}
	newErr.Details[key] = value
	return &newErr
}

// WithSuggestion adds a suggestion to the error
func (e *ResilienceError) WithSuggestion(suggestion string) *ResilienceError {
	newErr := *e
	newErr.Suggestions = append(newErr.Suggestions, suggestion)
	return &newErr
}

// ValidationError represents a validation error
type ValidationError struct {
	*ResilienceError
	Field       string      `json:"field"`
	Value       interface{} `json:"value,omitempty"`
	Constraint  string      `json:"constraint"`
}

// NewValidationError creates a new validation error
func NewValidationError(field, constraint string, value interface{}) *ValidationError {
	return &ValidationError{
		ResilienceError: &ResilienceError{
			Code:     CodeValidationFailed,
			Message:  fmt.Sprintf("Validation failed for field '%s': %s", field, constraint),
			Category: CategoryValidation,
			Severity: SeverityWarning,
			Recovery: RecoveryNonRetryable,
			Context:  ErrorContext{Timestamp: time.Now()},
		},
		Field:      field,
		Value:      value,
		Constraint: constraint,
	}
}

// NetworkError represents a network-related error
type NetworkError struct {
	*ResilienceError
	Host        string        `json:"host"`
	Port        int           `json:"port,omitempty"`
	Protocol    string        `json:"protocol,omitempty"`
	StatusCode  int           `json:"status_code,omitempty"`
	Latency     time.Duration `json:"latency,omitempty"`
}

// NewNetworkError creates a new network error
func NewNetworkError(host string, cause error) *NetworkError {
	return &NetworkError{
		ResilienceError: &ResilienceError{
			Code:     CodeNetworkError,
			Message:  fmt.Sprintf("Network error connecting to %s", host),
			Category: CategoryNetwork,
			Severity: SeverityError,
			Recovery: RecoveryRetryable,
			Context:  ErrorContext{Timestamp: time.Now()},
			Cause:    cause,
		},
		Host: host,
	}
}

// DatabaseError represents a database-related error
type DatabaseError struct {
	*ResilienceError
	Query       string        `json:"query,omitempty"`
	Table       string        `json:"table,omitempty"`
	Operation   string        `json:"operation"`
	Duration    time.Duration `json:"duration,omitempty"`
}

// NewDatabaseError creates a new database error
func NewDatabaseError(operation string, cause error) *DatabaseError {
	return &DatabaseError{
		ResilienceError: &ResilienceError{
			Code:     CodeDatabaseError,
			Message:  fmt.Sprintf("Database error during %s operation", operation),
			Category: CategoryDatabase,
			Severity: SeverityError,
			Recovery: RecoveryRetryable,
			Context:  ErrorContext{Timestamp: time.Now()},
			Cause:    cause,
		},
		Operation: operation,
	}
}

// AIError represents an AI service error
type AIError struct {
	*ResilienceError
	Provider    string `json:"provider"`
	Model       string `json:"model,omitempty"`
	TokensUsed  int    `json:"tokens_used,omitempty"`
	Cost        float64 `json:"cost,omitempty"`
}

// NewAIError creates a new AI error
func NewAIError(provider, model string, cause error) *AIError {
	return &AIError{
		ResilienceError: &ResilienceError{
			Code:     CodeAIServiceError,
			Message:  fmt.Sprintf("AI service error from provider %s", provider),
			Category: CategoryAI,
			Severity: SeverityError,
			Recovery: RecoveryRetryable,
			Context:  ErrorContext{Timestamp: time.Now()},
			Cause:    cause,
		},
		Provider: provider,
		Model:    model,
	}
}

// TimeoutError represents a timeout error
type TimeoutError struct {
	*ResilienceError
	Timeout     time.Duration `json:"timeout"`
	Elapsed     time.Duration `json:"elapsed"`
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(timeout, elapsed time.Duration) *TimeoutError {
	return &TimeoutError{
		ResilienceError: &ResilienceError{
			Code:     CodeTimeout,
			Message:  fmt.Sprintf("Operation timed out after %v (timeout: %v)", elapsed, timeout),
			Category: CategoryTimeout,
			Severity: SeverityError,
			Recovery: RecoveryRetryable,
			Context:  ErrorContext{Timestamp: time.Now()},
		},
		Timeout: timeout,
		Elapsed: elapsed,
	}
}

// CircuitBreakerError represents a circuit breaker error
type CircuitBreakerError struct {
	*ResilienceError
	State           string        `json:"state"`
	FailureCount    int           `json:"failure_count"`
	LastFailureTime time.Time     `json:"last_failure_time"`
	NextRetryTime   time.Time     `json:"next_retry_time"`
}

// NewCircuitBreakerError creates a new circuit breaker error
func NewCircuitBreakerError(state string, failureCount int) *CircuitBreakerError {
	return &CircuitBreakerError{
		ResilienceError: &ResilienceError{
			Code:     CodeCircuitBreakerOpen,
			Message:  fmt.Sprintf("Circuit breaker is %s (failures: %d)", state, failureCount),
			Category: CategoryCircuitBreaker,
			Severity: SeverityWarning,
			Recovery: RecoveryCircuitBreak,
			Context:  ErrorContext{Timestamp: time.Now()},
		},
		State:        state,
		FailureCount: failureCount,
	}
}

// BusinessError represents a business logic error
type BusinessError struct {
	*ResilienceError
	BusinessRule string      `json:"business_rule"`
	Entity       string      `json:"entity,omitempty"`
	EntityID     string      `json:"entity_id,omitempty"`
}

// NewBusinessError creates a new business error
func NewBusinessError(rule, message string) *BusinessError {
	return &BusinessError{
		ResilienceError: &ResilienceError{
			Code:     CodeBusinessRuleViolation,
			Message:  message,
			Category: CategoryBusiness,
			Severity: SeverityWarning,
			Recovery: RecoveryNonRetryable,
			Context:  ErrorContext{Timestamp: time.Now()},
		},
		BusinessRule: rule,
	}
}
