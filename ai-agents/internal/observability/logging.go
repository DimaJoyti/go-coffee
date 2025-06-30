package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// LogLevel represents logging levels
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// StructuredLogger provides structured logging with OpenTelemetry integration
type StructuredLogger struct {
	logger *slog.Logger
	config LoggingConfig
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(config LoggingConfig) *StructuredLogger {
	var handler slog.Handler
	var output io.Writer = os.Stdout

	// Configure output
	if config.Format == "json" {
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{
			Level: parseLogLevel(config.Level),
			AddSource: true,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// Customize attribute formatting
				if a.Key == slog.TimeKey {
					return slog.Attr{
						Key:   "timestamp",
						Value: slog.StringValue(a.Value.Time().Format(time.RFC3339)),
					}
				}
				if a.Key == slog.LevelKey {
					return slog.Attr{
						Key:   "level",
						Value: slog.StringValue(strings.ToLower(a.Value.String())),
					}
				}
				if a.Key == slog.MessageKey {
					return slog.Attr{
						Key:   "message",
						Value: a.Value,
					}
				}
				return a
			},
		})
	} else {
		handler = slog.NewTextHandler(output, &slog.HandlerOptions{
			Level: parseLogLevel(config.Level),
			AddSource: true,
		})
	}

	return &StructuredLogger{
		logger: slog.New(handler),
		config: config,
	}
}

// parseLogLevel converts string to slog.Level
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Debug logs a debug message
func (sl *StructuredLogger) Debug(msg string, fields ...interface{}) {
	sl.logWithContext(context.Background(), slog.LevelDebug, msg, fields...)
}

// DebugContext logs a debug message with context
func (sl *StructuredLogger) DebugContext(ctx context.Context, msg string, fields ...interface{}) {
	sl.logWithContext(ctx, slog.LevelDebug, msg, fields...)
}

// Info logs an info message
func (sl *StructuredLogger) Info(msg string, fields ...interface{}) {
	sl.logWithContext(context.Background(), slog.LevelInfo, msg, fields...)
}

// InfoContext logs an info message with context
func (sl *StructuredLogger) InfoContext(ctx context.Context, msg string, fields ...interface{}) {
	sl.logWithContext(ctx, slog.LevelInfo, msg, fields...)
}

// Warn logs a warning message
func (sl *StructuredLogger) Warn(msg string, fields ...interface{}) {
	sl.logWithContext(context.Background(), slog.LevelWarn, msg, fields...)
}

// WarnContext logs a warning message with context
func (sl *StructuredLogger) WarnContext(ctx context.Context, msg string, fields ...interface{}) {
	sl.logWithContext(ctx, slog.LevelWarn, msg, fields...)
}

// Error logs an error message
func (sl *StructuredLogger) Error(msg string, err error, fields ...interface{}) {
	allFields := append([]interface{}{"error", err.Error()}, fields...)
	sl.logWithContext(context.Background(), slog.LevelError, msg, allFields...)
}

// ErrorContext logs an error message with context
func (sl *StructuredLogger) ErrorContext(ctx context.Context, msg string, err error, fields ...interface{}) {
	allFields := append([]interface{}{"error", err.Error()}, fields...)
	sl.logWithContext(ctx, slog.LevelError, msg, allFields...)
}

// logWithContext logs a message with context and trace information
func (sl *StructuredLogger) logWithContext(ctx context.Context, level slog.Level, msg string, fields ...interface{}) {
	attrs := sl.buildAttributes(ctx, fields...)
	sl.logger.LogAttrs(ctx, level, msg, attrs...)
}

// buildAttributes builds slog attributes from fields and context
func (sl *StructuredLogger) buildAttributes(ctx context.Context, fields ...interface{}) []slog.Attr {
	var attrs []slog.Attr

	// Add trace information if enabled
	if sl.config.IncludeTrace || sl.config.IncludeSpan {
		span := trace.SpanFromContext(ctx)
		if span.SpanContext().IsValid() {
			if sl.config.IncludeTrace {
				attrs = append(attrs, slog.String("trace_id", span.SpanContext().TraceID().String()))
			}
			if sl.config.IncludeSpan {
				attrs = append(attrs, slog.String("span_id", span.SpanContext().SpanID().String()))
			}
		}
	}

	// Add correlation ID if enabled
	if sl.config.CorrelationID {
		if correlationID := getCorrelationID(ctx); correlationID != "" {
			attrs = append(attrs, slog.String("correlation_id", correlationID))
		}
	}

	// Add caller information
	if pc, file, line, ok := runtime.Caller(3); ok {
		funcName := runtime.FuncForPC(pc).Name()
		attrs = append(attrs, slog.Group("source",
			slog.String("function", funcName),
			slog.String("file", file),
			slog.Int("line", line),
		))
	}

	// Process field pairs
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key := fmt.Sprintf("%v", fields[i])
			value := fields[i+1]
			attrs = append(attrs, slog.Any(key, value))
		}
	}

	return attrs
}

// WithFields returns a logger with additional fields
func (sl *StructuredLogger) WithFields(fields ...interface{}) *FieldLogger {
	return &FieldLogger{
		logger: sl,
		fields: fields,
	}
}

// WithContext returns a logger with context
func (sl *StructuredLogger) WithContext(ctx context.Context) *ContextLogger {
	return &ContextLogger{
		logger: sl,
		ctx:    ctx,
	}
}

// FieldLogger wraps StructuredLogger with additional fields
type FieldLogger struct {
	logger *StructuredLogger
	fields []interface{}
}

// Debug logs a debug message with additional fields
func (fl *FieldLogger) Debug(msg string, fields ...interface{}) {
	allFields := append(fl.fields, fields...)
	fl.logger.Debug(msg, allFields...)
}

// Info logs an info message with additional fields
func (fl *FieldLogger) Info(msg string, fields ...interface{}) {
	allFields := append(fl.fields, fields...)
	fl.logger.Info(msg, allFields...)
}

// Warn logs a warning message with additional fields
func (fl *FieldLogger) Warn(msg string, fields ...interface{}) {
	allFields := append(fl.fields, fields...)
	fl.logger.Warn(msg, allFields...)
}

// Error logs an error message with additional fields
func (fl *FieldLogger) Error(msg string, err error, fields ...interface{}) {
	allFields := append(fl.fields, fields...)
	fl.logger.Error(msg, err, allFields...)
}

// ContextLogger wraps StructuredLogger with context
type ContextLogger struct {
	logger *StructuredLogger
	ctx    context.Context
}

// Debug logs a debug message with context
func (cl *ContextLogger) Debug(msg string, fields ...interface{}) {
	cl.logger.DebugContext(cl.ctx, msg, fields...)
}

// Info logs an info message with context
func (cl *ContextLogger) Info(msg string, fields ...interface{}) {
	cl.logger.InfoContext(cl.ctx, msg, fields...)
}

// Warn logs a warning message with context
func (cl *ContextLogger) Warn(msg string, fields ...interface{}) {
	cl.logger.WarnContext(cl.ctx, msg, fields...)
}

// Error logs an error message with context
func (cl *ContextLogger) Error(msg string, err error, fields ...interface{}) {
	cl.logger.ErrorContext(cl.ctx, msg, err, fields...)
}

// BusinessLogger provides business-specific logging
type BusinessLogger struct {
	logger *StructuredLogger
}

// NewBusinessLogger creates a new business logger
func NewBusinessLogger(logger *StructuredLogger) *BusinessLogger {
	return &BusinessLogger{
		logger: logger,
	}
}

// LogBeverageCreated logs beverage creation
func (bl *BusinessLogger) LogBeverageCreated(ctx context.Context, beverageID, name, theme string, aiUsed bool, duration time.Duration) {
	bl.logger.InfoContext(ctx, "Beverage created",
		"event_type", "beverage_created",
		"beverage_id", beverageID,
		"beverage_name", name,
		"theme", theme,
		"ai_used", aiUsed,
		"duration_ms", duration.Milliseconds(),
	)
}

// LogTaskCreated logs task creation
func (bl *BusinessLogger) LogTaskCreated(ctx context.Context, taskID, title, assignee string, priority string) {
	bl.logger.InfoContext(ctx, "Task created",
		"event_type", "task_created",
		"task_id", taskID,
		"task_title", title,
		"assignee", assignee,
		"priority", priority,
	)
}

// LogNotificationSent logs notification sending
func (bl *BusinessLogger) LogNotificationSent(ctx context.Context, channel, recipient string, success bool) {
	level := slog.LevelInfo
	message := "Notification sent successfully"
	
	if !success {
		level = slog.LevelWarn
		message = "Notification sending failed"
	}

	bl.logger.logWithContext(ctx, level, message,
		"event_type", "notification_sent",
		"channel", channel,
		"recipient", recipient,
		"success", success,
	)
}

// LogAIRequest logs AI provider requests
func (bl *BusinessLogger) LogAIRequest(ctx context.Context, provider, operation string, duration time.Duration, success bool, tokenCount int) {
	level := slog.LevelInfo
	message := "AI request completed"
	
	if !success {
		level = slog.LevelWarn
		message = "AI request failed"
	}

	bl.logger.logWithContext(ctx, level, message,
		"event_type", "ai_request",
		"provider", provider,
		"operation", operation,
		"duration_ms", duration.Milliseconds(),
		"success", success,
		"token_count", tokenCount,
	)
}

// AuditLogger provides audit logging
type AuditLogger struct {
	logger *StructuredLogger
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger *StructuredLogger) *AuditLogger {
	return &AuditLogger{
		logger: logger,
	}
}

// LogUserAction logs user actions for audit purposes
func (al *AuditLogger) LogUserAction(ctx context.Context, userID, action, resource string, metadata map[string]interface{}) {
	fields := []interface{}{
		"event_type", "user_action",
		"user_id", userID,
		"action", action,
		"resource", resource,
	}

	// Add metadata fields
	for key, value := range metadata {
		fields = append(fields, key, value)
	}

	al.logger.InfoContext(ctx, "User action performed", fields...)
}

// LogSystemEvent logs system events
func (al *AuditLogger) LogSystemEvent(ctx context.Context, event, component string, metadata map[string]interface{}) {
	fields := []interface{}{
		"event_type", "system_event",
		"event", event,
		"component", component,
	}

	// Add metadata fields
	for key, value := range metadata {
		fields = append(fields, key, value)
	}

	al.logger.InfoContext(ctx, "System event occurred", fields...)
}

// LogSecurityEvent logs security-related events
func (al *AuditLogger) LogSecurityEvent(ctx context.Context, event, userID, ipAddress string, severity string) {
	al.logger.WarnContext(ctx, "Security event detected",
		"event_type", "security_event",
		"event", event,
		"user_id", userID,
		"ip_address", ipAddress,
		"severity", severity,
	)
}

// Helper functions

// getCorrelationID extracts correlation ID from context
func getCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value("correlation_id").(string); ok {
		return correlationID
	}
	return ""
}

// LogJSON logs a JSON object
func LogJSON(logger *StructuredLogger, level LogLevel, msg string, obj interface{}) {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		logger.Error("Failed to marshal JSON for logging", err, "object", fmt.Sprintf("%+v", obj))
		return
	}

	switch level {
	case LevelDebug:
		logger.Debug(msg, "json_data", string(jsonData))
	case LevelInfo:
		logger.Info(msg, "json_data", string(jsonData))
	case LevelWarn:
		logger.Warn(msg, "json_data", string(jsonData))
	case LevelError:
		logger.Error(msg, fmt.Errorf("json log"), "json_data", string(jsonData))
	}
}

// Global logger instance
var globalLogger *StructuredLogger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(config LoggingConfig) {
	globalLogger = NewStructuredLogger(config)
}

// GetGlobalLogger returns the global logger
func GetGlobalLogger() *StructuredLogger {
	if globalLogger == nil {
		// Fallback to default logger
		globalLogger = NewStructuredLogger(LoggingConfig{
			Enabled:       true,
			Level:         "info",
			Format:        "json",
			IncludeTrace:  true,
			IncludeSpan:   true,
			CorrelationID: true,
		})
	}
	return globalLogger
}
