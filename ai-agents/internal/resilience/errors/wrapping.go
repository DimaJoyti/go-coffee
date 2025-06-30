package errors

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// ErrorBuilder provides a fluent interface for building resilience errors
type ErrorBuilder struct {
	err *ResilienceError
}

// NewError creates a new error builder
func NewError(code ErrorCode, message string) *ErrorBuilder {
	info, exists := GetErrorCodeInfo(code)
	if !exists {
		// Default values if code is not registered
		info = ErrorCodeInfo{
			Code:       code,
			Category:   CategoryInternal,
			Severity:   SeverityError,
			Recovery:   RecoveryNonRetryable,
			HTTPStatus: 500,
		}
	}

	return &ErrorBuilder{
		err: &ResilienceError{
			Code:     code,
			Message:  message,
			Category: info.Category,
			Severity: info.Severity,
			Recovery: info.Recovery,
			Context:  ErrorContext{Timestamp: time.Now()},
		},
	}
}

// WithCategory sets the error category
func (b *ErrorBuilder) WithCategory(category ErrorCategory) *ErrorBuilder {
	b.err.Category = category
	return b
}

// WithSeverity sets the error severity
func (b *ErrorBuilder) WithSeverity(severity ErrorSeverity) *ErrorBuilder {
	b.err.Severity = severity
	return b
}

// WithRecovery sets the error recovery type
func (b *ErrorBuilder) WithRecovery(recovery ErrorRecovery) *ErrorBuilder {
	b.err.Recovery = recovery
	return b
}

// WithCause sets the underlying cause
func (b *ErrorBuilder) WithCause(cause error) *ErrorBuilder {
	b.err.Cause = cause
	return b
}

// WithOperation sets the operation context
func (b *ErrorBuilder) WithOperation(operation string) *ErrorBuilder {
	b.err.Context.Operation = operation
	return b
}

// WithComponent sets the component context
func (b *ErrorBuilder) WithComponent(component string) *ErrorBuilder {
	b.err.Context.Component = component
	return b
}

// WithResource sets the resource context
func (b *ErrorBuilder) WithResource(resource, resourceID string) *ErrorBuilder {
	b.err.Context.Resource = resource
	b.err.Context.ResourceID = resourceID
	return b
}

// WithRequestContext sets the request context
func (b *ErrorBuilder) WithRequestContext(requestID, userID, traceID, spanID string) *ErrorBuilder {
	b.err.Context.RequestID = requestID
	b.err.Context.UserID = userID
	b.err.Context.TraceID = traceID
	b.err.Context.SpanID = spanID
	return b
}

// WithRetryContext sets the retry context
func (b *ErrorBuilder) WithRetryContext(attemptCount, maxAttempts int) *ErrorBuilder {
	b.err.Context.AttemptCount = attemptCount
	b.err.Context.MaxAttempts = maxAttempts
	return b
}

// WithDuration sets the operation duration
func (b *ErrorBuilder) WithDuration(duration time.Duration) *ErrorBuilder {
	b.err.Context.Duration = duration
	return b
}

// WithTimeout sets the operation timeout
func (b *ErrorBuilder) WithTimeout(timeout time.Duration) *ErrorBuilder {
	b.err.Context.Timeout = timeout
	return b
}

// WithDetail adds a detail to the error
func (b *ErrorBuilder) WithDetail(key string, value interface{}) *ErrorBuilder {
	if b.err.Details == nil {
		b.err.Details = make(map[string]interface{})
	}
	b.err.Details[key] = value
	return b
}

// WithMetadata adds metadata to the error context
func (b *ErrorBuilder) WithMetadata(key string, value interface{}) *ErrorBuilder {
	if b.err.Context.Metadata == nil {
		b.err.Context.Metadata = make(map[string]interface{})
	}
	b.err.Context.Metadata[key] = value
	return b
}

// WithSuggestion adds a suggestion for resolving the error
func (b *ErrorBuilder) WithSuggestion(suggestion string) *ErrorBuilder {
	b.err.Suggestions = append(b.err.Suggestions, suggestion)
	return b
}

// WithStackTrace adds stack trace information
func (b *ErrorBuilder) WithStackTrace() *ErrorBuilder {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(2, pcs[:])
	
	var frames []string
	for i := 0; i < n; i++ {
		pc := pcs[i]
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			file, line := fn.FileLine(pc)
			frames = append(frames, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
		}
	}
	
	return b.WithDetail("stack_trace", frames)
}

// Build returns the constructed error
func (b *ErrorBuilder) Build() *ResilienceError {
	return b.err
}

// Wrap wraps an existing error with resilience error information
func Wrap(err error, code ErrorCode, message string) *ResilienceError {
	if err == nil {
		return nil
	}

	// If it's already a ResilienceError, enhance it
	if resErr, ok := err.(*ResilienceError); ok {
		return NewError(code, message).
			WithCause(resErr.Cause).
			WithCategory(resErr.Category).
			WithSeverity(resErr.Severity).
			WithRecovery(resErr.Recovery).
			Build()
	}

	return NewError(code, message).
		WithCause(err).
		Build()
}

// WrapWithContext wraps an error with context information
func WrapWithContext(ctx context.Context, err error, code ErrorCode, message string) *ResilienceError {
	if err == nil {
		return nil
	}

	builder := NewError(code, message).WithCause(err)

	// Extract context information
	if requestID := getRequestIDFromContext(ctx); requestID != "" {
		builder = builder.WithRequestContext(requestID, "", "", "")
	}

	if userID := getUserIDFromContext(ctx); userID != "" {
		builder.err.Context.UserID = userID
	}

	if traceID := getTraceIDFromContext(ctx); traceID != "" {
		builder.err.Context.TraceID = traceID
	}

	return builder.Build()
}

// WrapValidation creates a validation error from a generic error
func WrapValidation(err error, field, constraint string) *ValidationError {
	if err == nil {
		return nil
	}

	return &ValidationError{
		ResilienceError: NewError(CodeValidationFailed, fmt.Sprintf("Validation failed for field '%s': %s", field, constraint)).
			WithCause(err).
			Build(),
		Field:      field,
		Constraint: constraint,
	}
}

// WrapNetwork creates a network error from a generic error
func WrapNetwork(err error, host string) *NetworkError {
	if err == nil {
		return nil
	}

	return &NetworkError{
		ResilienceError: NewError(CodeNetworkError, fmt.Sprintf("Network error connecting to %s", host)).
			WithCause(err).
			Build(),
		Host: host,
	}
}

// WrapDatabase creates a database error from a generic error
func WrapDatabase(err error, operation string) *DatabaseError {
	if err == nil {
		return nil
	}

	return &DatabaseError{
		ResilienceError: NewError(CodeDatabaseError, fmt.Sprintf("Database error during %s operation", operation)).
			WithCause(err).
			Build(),
		Operation: operation,
	}
}

// WrapAI creates an AI error from a generic error
func WrapAI(err error, provider, model string) *AIError {
	if err == nil {
		return nil
	}

	return &AIError{
		ResilienceError: NewError(CodeAIServiceError, fmt.Sprintf("AI service error from provider %s", provider)).
			WithCause(err).
			Build(),
		Provider: provider,
		Model:    model,
	}
}

// WrapTimeout creates a timeout error from a generic error
func WrapTimeout(err error, timeout, elapsed time.Duration) *TimeoutError {
	if err == nil {
		return nil
	}

	return &TimeoutError{
		ResilienceError: NewError(CodeTimeout, fmt.Sprintf("Operation timed out after %v (timeout: %v)", elapsed, timeout)).
			WithCause(err).
			Build(),
		Timeout: timeout,
		Elapsed: elapsed,
	}
}

// Chain creates an error chain from multiple errors
func Chain(errors ...error) *ResilienceError {
	if len(errors) == 0 {
		return nil
	}

	// Filter out nil errors
	var validErrors []error
	for _, err := range errors {
		if err != nil {
			validErrors = append(validErrors, err)
		}
	}

	if len(validErrors) == 0 {
		return nil
	}

	if len(validErrors) == 1 {
		if resErr, ok := validErrors[0].(*ResilienceError); ok {
			return resErr
		}
		return Wrap(validErrors[0], CodeInternalError, "Error occurred")
	}

	// Create a chain error
	chainErr := NewError(CodeInternalError, fmt.Sprintf("Multiple errors occurred (%d errors)", len(validErrors))).
		WithDetail("error_count", len(validErrors)).
		Build()

	// Add all errors as details
	for i, err := range validErrors {
		chainErr = chainErr.WithDetail(fmt.Sprintf("error_%d", i+1), err.Error())
	}

	// Use the first error as the cause
	chainErr.Cause = validErrors[0]

	return chainErr
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	if resErr, ok := err.(*ResilienceError); ok {
		return resErr.IsRetryable()
	}

	// Check for specific error types that are typically retryable
	switch err {
	case context.DeadlineExceeded:
		return true
	case context.Canceled:
		return false
	default:
		// Check error message for common retryable patterns
		errMsg := err.Error()
		retryablePatterns := []string{
			"connection refused",
			"connection reset",
			"timeout",
			"temporary failure",
			"service unavailable",
			"rate limit",
		}

		for _, pattern := range retryablePatterns {
			if contains(errMsg, pattern) {
				return true
			}
		}
	}

	return false
}

// IsCritical checks if an error is critical
func IsCritical(err error) bool {
	if err == nil {
		return false
	}

	if resErr, ok := err.(*ResilienceError); ok {
		return resErr.IsSeverity(SeverityCritical)
	}

	return false
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) ErrorCode {
	if err == nil {
		return ""
	}

	if resErr, ok := err.(*ResilienceError); ok {
		return resErr.Code
	}

	return CodeInternalError
}

// GetErrorCategory extracts the error category from an error
func GetErrorCategory(err error) ErrorCategory {
	if err == nil {
		return ""
	}

	if resErr, ok := err.(*ResilienceError); ok {
		return resErr.Category
	}

	return CategoryInternal
}

// Helper functions for context extraction
func getRequestIDFromContext(ctx context.Context) string {
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

func getUserIDFromContext(ctx context.Context) string {
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

func getTraceIDFromContext(ctx context.Context) string {
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     (s[:len(substr)] == substr || 
		      s[len(s)-len(substr):] == substr ||
		      containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
