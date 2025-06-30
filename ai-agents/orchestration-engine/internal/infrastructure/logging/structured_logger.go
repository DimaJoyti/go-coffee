package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// StructuredLogger provides structured logging with correlation IDs and context
type StructuredLogger struct {
	output      io.Writer
	level       LogLevel
	serviceName string
	version     string
	environment string
	fields      map[string]interface{}
	mutex       sync.RWMutex
}

// LogLevel represents logging levels
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp     time.Time              `json:"timestamp"`
	Level         string                 `json:"level"`
	Message       string                 `json:"message"`
	ServiceName   string                 `json:"service_name"`
	Version       string                 `json:"version"`
	Environment   string                 `json:"environment"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	TraceID       string                 `json:"trace_id,omitempty"`
	SpanID        string                 `json:"span_id,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
	RequestID     string                 `json:"request_id,omitempty"`
	SessionID     string                 `json:"session_id,omitempty"`
	Component     string                 `json:"component,omitempty"`
	Operation     string                 `json:"operation,omitempty"`
	Duration      *time.Duration         `json:"duration,omitempty"`
	Error         *ErrorInfo             `json:"error,omitempty"`
	Fields        map[string]interface{} `json:"fields,omitempty"`
	Caller        *CallerInfo            `json:"caller,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
}

// ErrorInfo represents error information in logs
type ErrorInfo struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	StackTrace string `json:"stack_trace,omitempty"`
	Code       string `json:"code,omitempty"`
}

// CallerInfo represents caller information
type CallerInfo struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
}

// ContextKey represents context keys for logging
type ContextKey string

const (
	ContextKeyCorrelationID ContextKey = "correlation_id"
	ContextKeyTraceID       ContextKey = "trace_id"
	ContextKeySpanID        ContextKey = "span_id"
	ContextKeyUserID        ContextKey = "user_id"
	ContextKeyRequestID     ContextKey = "request_id"
	ContextKeySessionID     ContextKey = "session_id"
	ContextKeyComponent     ContextKey = "component"
	ContextKeyOperation     ContextKey = "operation"
)

// LoggerConfig contains logger configuration
type LoggerConfig struct {
	Level        LogLevel `json:"level"`
	ServiceName  string   `json:"service_name"`
	Version      string   `json:"version"`
	Environment  string   `json:"environment"`
	Output       string   `json:"output"` // stdout, stderr, file
	FilePath     string   `json:"file_path,omitempty"`
	EnableCaller bool     `json:"enable_caller"`
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(config *LoggerConfig) *StructuredLogger {
	if config == nil {
		config = DefaultLoggerConfig()
	}

	var output io.Writer = os.Stdout
	if config.Output == "stderr" {
		output = os.Stderr
	} else if config.Output == "file" && config.FilePath != "" {
		if file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
			output = file
		}
	}

	return &StructuredLogger{
		output:      output,
		level:       config.Level,
		serviceName: config.ServiceName,
		version:     config.Version,
		environment: config.Environment,
		fields:      make(map[string]interface{}),
	}
}

// DefaultLoggerConfig returns default logger configuration
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:        LogLevelInfo,
		ServiceName:  "orchestration-engine",
		Version:      "1.0.0",
		Environment:  "development",
		Output:       "stdout",
		EnableCaller: true,
	}
}

// WithContext creates a logger with context information
func (sl *StructuredLogger) WithContext(ctx context.Context) *StructuredLogger {
	logger := &StructuredLogger{
		output:      sl.output,
		level:       sl.level,
		serviceName: sl.serviceName,
		version:     sl.version,
		environment: sl.environment,
		fields:      make(map[string]interface{}),
	}

	// Copy existing fields
	sl.mutex.RLock()
	for k, v := range sl.fields {
		logger.fields[k] = v
	}
	sl.mutex.RUnlock()

	// Extract context values
	if correlationID := ctx.Value(ContextKeyCorrelationID); correlationID != nil {
		logger.fields["correlation_id"] = correlationID
	}
	if traceID := ctx.Value(ContextKeyTraceID); traceID != nil {
		logger.fields["trace_id"] = traceID
	}
	if spanID := ctx.Value(ContextKeySpanID); spanID != nil {
		logger.fields["span_id"] = spanID
	}
	if userID := ctx.Value(ContextKeyUserID); userID != nil {
		logger.fields["user_id"] = userID
	}
	if requestID := ctx.Value(ContextKeyRequestID); requestID != nil {
		logger.fields["request_id"] = requestID
	}
	if sessionID := ctx.Value(ContextKeySessionID); sessionID != nil {
		logger.fields["session_id"] = sessionID
	}
	if component := ctx.Value(ContextKeyComponent); component != nil {
		logger.fields["component"] = component
	}
	if operation := ctx.Value(ContextKeyOperation); operation != nil {
		logger.fields["operation"] = operation
	}

	return logger
}

// WithFields creates a logger with additional fields
func (sl *StructuredLogger) WithFields(fields map[string]interface{}) *StructuredLogger {
	logger := &StructuredLogger{
		output:      sl.output,
		level:       sl.level,
		serviceName: sl.serviceName,
		version:     sl.version,
		environment: sl.environment,
		fields:      make(map[string]interface{}),
	}

	// Copy existing fields
	sl.mutex.RLock()
	for k, v := range sl.fields {
		logger.fields[k] = v
	}
	sl.mutex.RUnlock()

	// Add new fields
	for k, v := range fields {
		logger.fields[k] = v
	}

	return logger
}

// WithField creates a logger with an additional field
func (sl *StructuredLogger) WithField(key string, value interface{}) *StructuredLogger {
	return sl.WithFields(map[string]interface{}{key: value})
}

// WithComponent creates a logger with component information
func (sl *StructuredLogger) WithComponent(component string) *StructuredLogger {
	return sl.WithField("component", component)
}

// WithOperation creates a logger with operation information
func (sl *StructuredLogger) WithOperation(operation string) *StructuredLogger {
	return sl.WithField("operation", operation)
}

// WithError creates a logger with error information
func (sl *StructuredLogger) WithError(err error) *StructuredLogger {
	if err == nil {
		return sl
	}

	errorInfo := &ErrorInfo{
		Type:    fmt.Sprintf("%T", err),
		Message: err.Error(),
	}

	return sl.WithField("error", errorInfo)
}

// Debug logs a debug message
func (sl *StructuredLogger) Debug(msg string, args ...interface{}) {
	if sl.level > LogLevelDebug {
		return
	}
	sl.log(LogLevelDebug, msg, args...)
}

// Info logs an info message
func (sl *StructuredLogger) Info(msg string, args ...interface{}) {
	if sl.level > LogLevelInfo {
		return
	}
	sl.log(LogLevelInfo, msg, args...)
}

// Warn logs a warning message
func (sl *StructuredLogger) Warn(msg string, args ...interface{}) {
	if sl.level > LogLevelWarn {
		return
	}
	sl.log(LogLevelWarn, msg, args...)
}

// Error logs an error message
func (sl *StructuredLogger) Error(msg string, err error, args ...interface{}) {
	if sl.level > LogLevelError {
		return
	}

	logger := sl
	if err != nil {
		logger = sl.WithError(err)
	}

	logger.log(LogLevelError, msg, args...)
}

// Fatal logs a fatal message and exits
func (sl *StructuredLogger) Fatal(msg string, args ...interface{}) {
	sl.log(LogLevelFatal, msg, args...)
	os.Exit(1)
}

// log performs the actual logging
func (sl *StructuredLogger) log(level LogLevel, msg string, args ...interface{}) {
	entry := &LogEntry{
		Timestamp:   time.Now().UTC(),
		Level:       sl.levelToString(level),
		Message:     fmt.Sprintf(msg, args...),
		ServiceName: sl.serviceName,
		Version:     sl.version,
		Environment: sl.environment,
		Fields:      make(map[string]interface{}),
	}

	// Copy fields
	sl.mutex.RLock()
	for k, v := range sl.fields {
		switch k {
		case "correlation_id":
			if s, ok := v.(string); ok {
				entry.CorrelationID = s
			}
		case "trace_id":
			if s, ok := v.(string); ok {
				entry.TraceID = s
			}
		case "span_id":
			if s, ok := v.(string); ok {
				entry.SpanID = s
			}
		case "user_id":
			if s, ok := v.(string); ok {
				entry.UserID = s
			}
		case "request_id":
			if s, ok := v.(string); ok {
				entry.RequestID = s
			}
		case "session_id":
			if s, ok := v.(string); ok {
				entry.SessionID = s
			}
		case "component":
			if s, ok := v.(string); ok {
				entry.Component = s
			}
		case "operation":
			if s, ok := v.(string); ok {
				entry.Operation = s
			}
		case "duration":
			if d, ok := v.(time.Duration); ok {
				entry.Duration = &d
			}
		case "error":
			if e, ok := v.(*ErrorInfo); ok {
				entry.Error = e
			}
		case "tags":
			if t, ok := v.([]string); ok {
				entry.Tags = t
			}
		default:
			entry.Fields[k] = v
		}
	}
	sl.mutex.RUnlock()

	// Add caller information
	if pc, file, line, ok := runtime.Caller(2); ok {
		entry.Caller = &CallerInfo{
			File:     sl.shortenFilePath(file),
			Line:     line,
			Function: sl.getFunctionName(pc),
		}
	}

	// Marshal and write
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
		return
	}

	sl.output.Write(append(data, '\n'))
}

// levelToString converts log level to string
func (sl *StructuredLogger) levelToString(level LogLevel) string {
	switch level {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// shortenFilePath shortens file path for readability
func (sl *StructuredLogger) shortenFilePath(file string) string {
	parts := strings.Split(file, "/")
	if len(parts) > 3 {
		return strings.Join(parts[len(parts)-3:], "/")
	}
	return file
}

// getFunctionName extracts function name from program counter
func (sl *StructuredLogger) getFunctionName(pc uintptr) string {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}

	name := fn.Name()
	if lastSlash := strings.LastIndex(name, "/"); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if lastDot := strings.LastIndex(name, "."); lastDot >= 0 {
		name = name[lastDot+1:]
	}

	return name
}

// Context helper functions

// WithCorrelationID adds correlation ID to context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	if correlationID == "" {
		correlationID = uuid.New().String()
	}
	return context.WithValue(ctx, ContextKeyCorrelationID, correlationID)
}

// WithTraceID adds trace ID to context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, ContextKeyTraceID, traceID)
}

// WithSpanID adds span ID to context
func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, ContextKeySpanID, spanID)
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, userID)
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		requestID = uuid.New().String()
	}
	return context.WithValue(ctx, ContextKeyRequestID, requestID)
}

// WithSessionID adds session ID to context
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, ContextKeySessionID, sessionID)
}

// WithComponent adds component to context
func WithComponent(ctx context.Context, component string) context.Context {
	return context.WithValue(ctx, ContextKeyComponent, component)
}

// WithOperation adds operation to context
func WithOperation(ctx context.Context, operation string) context.Context {
	return context.WithValue(ctx, ContextKeyOperation, operation)
}

// GetCorrelationID extracts correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if id := ctx.Value(ContextKeyCorrelationID); id != nil {
		if s, ok := id.(string); ok {
			return s
		}
	}
	return ""
}

// GetTraceID extracts trace ID from context
func GetTraceID(ctx context.Context) string {
	if id := ctx.Value(ContextKeyTraceID); id != nil {
		if s, ok := id.(string); ok {
			return s
		}
	}
	return ""
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if id := ctx.Value(ContextKeyRequestID); id != nil {
		if s, ok := id.(string); ok {
			return s
		}
	}
	return ""
}

// LoggerMiddleware creates HTTP middleware for request logging
func LoggerMiddleware(logger *StructuredLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Generate request ID if not present
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
				r.Header.Set("X-Request-ID", requestID)
			}

			// Add correlation ID
			correlationID := r.Header.Get("X-Correlation-ID")
			if correlationID == "" {
				correlationID = uuid.New().String()
			}

			// Create context with logging information
			ctx := r.Context()
			ctx = WithRequestID(ctx, requestID)
			ctx = WithCorrelationID(ctx, correlationID)
			ctx = WithComponent(ctx, "http_server")
			ctx = WithOperation(ctx, fmt.Sprintf("%s %s", r.Method, r.URL.Path))

			// Create request logger
			reqLogger := logger.WithContext(ctx).WithFields(map[string]interface{}{
				"method":     r.Method,
				"path":       r.URL.Path,
				"query":      r.URL.RawQuery,
				"user_agent": r.UserAgent(),
				"remote_ip":  r.RemoteAddr,
			})

			// Log request start
			reqLogger.Info("HTTP request started")

			// Create response writer wrapper
			wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

			// Process request
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// Log request completion
			duration := time.Since(start)
			reqLogger.WithFields(map[string]interface{}{
				"status_code":   wrapped.statusCode,
				"duration":      duration,
				"bytes_written": wrapped.bytesWritten,
			}).Info("HTTP request completed")
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and bytes written
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// TimedOperation logs the duration of an operation
func (sl *StructuredLogger) TimedOperation(ctx context.Context, operation string, fn func() error) error {
	start := time.Now()

	logger := sl.WithContext(ctx).WithOperation(operation)
	logger.Info("Operation started")

	err := fn()
	duration := time.Since(start)

	if err != nil {
		logger.WithError(err).WithField("duration", duration).Error("Operation failed", err)
	} else {
		logger.WithField("duration", duration).Info("Operation completed")
	}

	return err
}
