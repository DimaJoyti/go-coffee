package errors

import (
	"fmt"
	"runtime"
	"time"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	// Business logic errors
	ErrorTypeValidation    ErrorType = "validation"
	ErrorTypeBusiness      ErrorType = "business"
	ErrorTypeNotFound      ErrorType = "not_found"
	ErrorTypeConflict      ErrorType = "conflict"
	ErrorTypeUnauthorized  ErrorType = "unauthorized"
	ErrorTypeForbidden     ErrorType = "forbidden"
	
	// Infrastructure errors
	ErrorTypeDatabase      ErrorType = "database"
	ErrorTypeNetwork       ErrorType = "network"
	ErrorTypeTimeout       ErrorType = "timeout"
	ErrorTypeRateLimit     ErrorType = "rate_limit"
	ErrorTypeCircuitBreaker ErrorType = "circuit_breaker"
	
	// External service errors
	ErrorTypeExternalAPI   ErrorType = "external_api"
	ErrorTypeKafka         ErrorType = "kafka"
	ErrorTypeAI            ErrorType = "ai_provider"
	ErrorTypeTaskManager   ErrorType = "task_manager"
	ErrorTypeNotification  ErrorType = "notification"
	
	// System errors
	ErrorTypeInternal      ErrorType = "internal"
	ErrorTypeConfiguration ErrorType = "configuration"
	ErrorTypeSerialization ErrorType = "serialization"
	ErrorTypeUnknown       ErrorType = "unknown"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity string

const (
	SeverityLow      ErrorSeverity = "low"
	SeverityMedium   ErrorSeverity = "medium"
	SeverityHigh     ErrorSeverity = "high"
	SeverityCritical ErrorSeverity = "critical"
)

// AppError represents a structured application error
type AppError struct {
	ID          string                 `json:"id"`
	Type        ErrorType              `json:"type"`
	Severity    ErrorSeverity          `json:"severity"`
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Details     string                 `json:"details,omitempty"`
	Cause       error                  `json:"-"`
	Context     map[string]interface{} `json:"context,omitempty"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Service     string                 `json:"service"`
	Operation   string                 `json:"operation,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	Retryable   bool                   `json:"retryable"`
	RetryAfter  *time.Duration         `json:"retry_after,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s:%s] %s: %s", e.Type, e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s:%s] %s", e.Type, e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target
func (e *AppError) Is(target error) bool {
	if appErr, ok := target.(*AppError); ok {
		return e.Type == appErr.Type && e.Code == appErr.Code
	}
	return false
}

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithOperation sets the operation name
func (e *AppError) WithOperation(operation string) *AppError {
	e.Operation = operation
	return e
}

// WithUserID sets the user ID
func (e *AppError) WithUserID(userID string) *AppError {
	e.UserID = userID
	return e
}

// WithRequestID sets the request ID
func (e *AppError) WithRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

// WithTraceID sets the trace ID
func (e *AppError) WithTraceID(traceID string) *AppError {
	e.TraceID = traceID
	return e
}

// WithRetryAfter sets the retry delay
func (e *AppError) WithRetryAfter(delay time.Duration) *AppError {
	e.RetryAfter = &delay
	return e
}

// NewAppError creates a new application error
func NewAppError(errorType ErrorType, code, message string) *AppError {
	return &AppError{
		ID:        generateErrorID(),
		Type:      errorType,
		Severity:  determineSeverity(errorType),
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Service:   "ai-agents",
		Retryable: isRetryable(errorType),
	}
}

// NewAppErrorWithCause creates a new application error with an underlying cause
func NewAppErrorWithCause(errorType ErrorType, code, message string, cause error) *AppError {
	err := NewAppError(errorType, code, message)
	err.Cause = cause
	err.StackTrace = captureStackTrace()
	return err
}

// Wrap wraps an existing error with additional context
func Wrap(err error, errorType ErrorType, code, message string) *AppError {
	if appErr, ok := err.(*AppError); ok {
		// If it's already an AppError, create a new one that wraps it
		return NewAppErrorWithCause(errorType, code, message, appErr)
	}
	return NewAppErrorWithCause(errorType, code, message, err)
}

// Common error constructors

// NewValidationError creates a validation error
func NewValidationError(code, message string) *AppError {
	return NewAppError(ErrorTypeValidation, code, message)
}

// NewBusinessError creates a business logic error
func NewBusinessError(code, message string) *AppError {
	return NewAppError(ErrorTypeBusiness, code, message)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource, id string) *AppError {
	return NewAppError(ErrorTypeNotFound, "RESOURCE_NOT_FOUND", 
		fmt.Sprintf("%s with ID '%s' not found", resource, id)).
		WithContext("resource", resource).
		WithContext("resource_id", id)
}

// NewConflictError creates a conflict error
func NewConflictError(resource, message string) *AppError {
	return NewAppError(ErrorTypeConflict, "RESOURCE_CONFLICT", message).
		WithContext("resource", resource)
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
	return NewAppError(ErrorTypeUnauthorized, "UNAUTHORIZED", message)
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *AppError {
	return NewAppError(ErrorTypeForbidden, "FORBIDDEN", message)
}

// NewDatabaseError creates a database error
func NewDatabaseError(operation string, cause error) *AppError {
	return NewAppErrorWithCause(ErrorTypeDatabase, "DATABASE_ERROR", 
		fmt.Sprintf("Database operation failed: %s", operation), cause).
		WithOperation(operation)
}

// NewNetworkError creates a network error
func NewNetworkError(operation string, cause error) *AppError {
	return NewAppErrorWithCause(ErrorTypeNetwork, "NETWORK_ERROR", 
		fmt.Sprintf("Network operation failed: %s", operation), cause).
		WithOperation(operation)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation string, timeout time.Duration) *AppError {
	return NewAppError(ErrorTypeTimeout, "OPERATION_TIMEOUT", 
		fmt.Sprintf("Operation '%s' timed out after %v", operation, timeout)).
		WithOperation(operation).
		WithContext("timeout", timeout.String())
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(service string, retryAfter time.Duration) *AppError {
	return NewAppError(ErrorTypeRateLimit, "RATE_LIMIT_EXCEEDED", 
		fmt.Sprintf("Rate limit exceeded for service: %s", service)).
		WithContext("service", service).
		WithRetryAfter(retryAfter)
}

// NewCircuitBreakerError creates a circuit breaker error
func NewCircuitBreakerError(service string) *AppError {
	return NewAppError(ErrorTypeCircuitBreaker, "CIRCUIT_BREAKER_OPEN", 
		fmt.Sprintf("Circuit breaker is open for service: %s", service)).
		WithContext("service", service)
}

// NewExternalAPIError creates an external API error
func NewExternalAPIError(service string, statusCode int, cause error) *AppError {
	return NewAppErrorWithCause(ErrorTypeExternalAPI, "EXTERNAL_API_ERROR", 
		fmt.Sprintf("External API call failed: %s (status: %d)", service, statusCode), cause).
		WithContext("service", service).
		WithContext("status_code", statusCode)
}

// NewKafkaError creates a Kafka error
func NewKafkaError(operation string, cause error) *AppError {
	return NewAppErrorWithCause(ErrorTypeKafka, "KAFKA_ERROR", 
		fmt.Sprintf("Kafka operation failed: %s", operation), cause).
		WithOperation(operation)
}

// NewAIProviderError creates an AI provider error
func NewAIProviderError(provider string, cause error) *AppError {
	return NewAppErrorWithCause(ErrorTypeAI, "AI_PROVIDER_ERROR", 
		fmt.Sprintf("AI provider error: %s", provider), cause).
		WithContext("provider", provider)
}

// NewSerializationError creates a serialization error
func NewSerializationError(format string, cause error) *AppError {
	return NewAppErrorWithCause(ErrorTypeSerialization, "SERIALIZATION_ERROR", 
		fmt.Sprintf("Serialization failed for format: %s", format), cause).
		WithContext("format", format)
}

// NewInternalError creates an internal error
func NewInternalError(message string, cause error) *AppError {
	return NewAppErrorWithCause(ErrorTypeInternal, "INTERNAL_ERROR", message, cause)
}

// Helper functions

// generateErrorID generates a unique error ID
func generateErrorID() string {
	return fmt.Sprintf("err_%d", time.Now().UnixNano())
}

// determineSeverity determines the severity based on error type
func determineSeverity(errorType ErrorType) ErrorSeverity {
	switch errorType {
	case ErrorTypeValidation, ErrorTypeNotFound, ErrorTypeUnauthorized, ErrorTypeForbidden:
		return SeverityLow
	case ErrorTypeBusiness, ErrorTypeConflict, ErrorTypeRateLimit:
		return SeverityMedium
	case ErrorTypeDatabase, ErrorTypeNetwork, ErrorTypeExternalAPI, ErrorTypeKafka:
		return SeverityHigh
	case ErrorTypeInternal, ErrorTypeTimeout, ErrorTypeCircuitBreaker:
		return SeverityCritical
	default:
		return SeverityMedium
	}
}

// isRetryable determines if an error type is retryable
func isRetryable(errorType ErrorType) bool {
	switch errorType {
	case ErrorTypeNetwork, ErrorTypeTimeout, ErrorTypeRateLimit, ErrorTypeExternalAPI, ErrorTypeKafka:
		return true
	case ErrorTypeValidation, ErrorTypeNotFound, ErrorTypeUnauthorized, ErrorTypeForbidden, ErrorTypeConflict:
		return false
	default:
		return false
	}
}

// captureStackTrace captures the current stack trace
func captureStackTrace() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	
	var trace string
	for {
		frame, more := frames.Next()
		trace += fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}
	return trace
}
