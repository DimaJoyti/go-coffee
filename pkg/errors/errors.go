package errors

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// ValidationError represents validation errors
	ValidationError ErrorType = "validation"
	// DatabaseError represents database errors
	DatabaseError ErrorType = "database"
	// NetworkError represents network errors
	NetworkError ErrorType = "network"
	// AuthenticationError represents authentication errors
	AuthenticationError ErrorType = "authentication"
	// AuthorizationError represents authorization errors
	AuthorizationError ErrorType = "authorization"
	// BusinessLogicError represents business logic errors
	BusinessLogicError ErrorType = "business_logic"
	// ExternalServiceError represents external service errors
	ExternalServiceError ErrorType = "external_service"
	// InternalError represents internal system errors
	InternalError ErrorType = "internal"
)

// AppError represents an enhanced application error with rich context
type AppError struct {
	// Err original error
	Err error `json:"-"`
	// Message error message
	Message string `json:"message"`
	// Code error code
	Code string `json:"code"`
	// Type error type
	Type ErrorType `json:"type"`
	// StatusCode HTTP status code
	StatusCode int `json:"status_code,omitempty"`
	// Stack call stack
	Stack string `json:"stack,omitempty"`
	// Context error context
	Context map[string]interface{} `json:"context,omitempty"`
	// Timestamp when error occurred
	Timestamp time.Time `json:"timestamp"`
	// Service that generated the error
	Service string `json:"service,omitempty"`
	// RequestID for tracing
	RequestID string `json:"request_id,omitempty"`
	// UserID if applicable
	UserID string `json:"user_id,omitempty"`
}

// Error returns the error message
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap returns the original error for error unwrapping
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value any) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// WithService sets the service name
func (e *AppError) WithService(service string) *AppError {
	e.Service = service
	return e
}

// WithRequestID sets the request ID for tracing
func (e *AppError) WithRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

// WithUserID sets the user ID
func (e *AppError) WithUserID(userID string) *AppError {
	e.UserID = userID
	return e
}

// WithCode sets the error code
func (e *AppError) WithCode(code string) *AppError {
	e.Code = code
	return e
}

// WithStatusCode sets the HTTP status code
func (e *AppError) WithStatusCode(statusCode int) *AppError {
	e.StatusCode = statusCode
	return e
}

// ToJSON returns the error as JSON
func (e *AppError) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// Is implements error comparison for Go 1.13+ error handling
func (e *AppError) Is(target error) bool {
	if t, ok := target.(*AppError); ok {
		return e.Type == t.Type && e.Code == t.Code
	}
	return false
}

// New створює нову помилку
func New(message string) *AppError {
	return &AppError{
		Message: message,
		Stack:   getStack(),
	}
}

// Wrap обгортає помилку
func Wrap(err error, message string) *AppError {
	if err == nil {
		return nil
	}

	// Якщо помилка вже є AppError, додаємо повідомлення
	if appErr, ok := err.(*AppError); ok {
		return &AppError{
			Err:        appErr.Err,
			Message:    fmt.Sprintf("%s: %s", message, appErr.Message),
			Code:       appErr.Code,
			StatusCode: appErr.StatusCode,
			Stack:      appErr.Stack,
			Context:    appErr.Context,
		}
	}

	return &AppError{
		Err:     err,
		Message: message,
		Stack:   getStack(),
	}
}

// getStack returns the call stack
func getStack() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var stack strings.Builder
	for {
		frame, more := frames.Next()
		if !strings.Contains(frame.File, "runtime/") && !strings.Contains(frame.File, "errors/errors.go") {
			stack.WriteString(fmt.Sprintf("%s:%d %s\n",
				filepath.Base(frame.File), frame.Line, frame.Function))
		}
		if !more {
			break
		}
	}
	return stack.String()
}

// IsTimeout checks if the error is a timeout error
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}

	// Check for timeout interfaces
	if t, ok := err.(interface{ Timeout() bool }); ok {
		return t.Timeout()
	}

	// Check error message for timeout indicators
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "deadline exceeded") ||
		strings.Contains(errMsg, "context deadline exceeded")
}

// IsRetryable checks if the error is retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Timeout errors are usually retryable
	if IsTimeout(err) {
		return true
	}

	// Check for temporary network errors
	if t, ok := err.(interface{ Temporary() bool }); ok {
		return t.Temporary()
	}

	return false
}
