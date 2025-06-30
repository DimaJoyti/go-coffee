package observability

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Context keys for request metadata
type contextKey string

const (
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
	traceIDKey   contextKey = "trace_id"
)

// TracingHelper provides utilities for distributed tracing
type TracingHelper struct {
	tracer trace.Tracer
}

// NewTracingHelper creates a new tracing helper
func NewTracingHelper(scope *InstrumentationScope) *TracingHelper {
	return &TracingHelper{
		tracer: scope.Tracer,
	}
}

// StartSpan starts a new span with common attributes
func (th *TracingHelper) StartSpan(ctx context.Context, operationName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	// Add common attributes
	defaultOpts := []trace.SpanStartOption{
		trace.WithAttributes(
			attribute.String("operation.name", operationName),
			attribute.String("service.name", "go-coffee-ai-agents"),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
	}

	// Merge with provided options
	allOpts := append(defaultOpts, opts...)

	return th.tracer.Start(ctx, operationName, allOpts...)
}

// StartHTTPSpan starts a span for HTTP operations
func (th *TracingHelper) StartHTTPSpan(ctx context.Context, method, url, userAgent string) (context.Context, trace.Span) {
	return th.StartSpan(ctx, fmt.Sprintf("HTTP %s", method),
		trace.WithAttributes(
			attribute.String("http.method", method),
			attribute.String("http.url", url),
			attribute.String("http.user_agent", userAgent),
		),
		trace.WithSpanKind(trace.SpanKindClient),
	)
}

// StartDatabaseSpan starts a span for database operations
func (th *TracingHelper) StartDatabaseSpan(ctx context.Context, operation, table string) (context.Context, trace.Span) {
	return th.StartSpan(ctx, fmt.Sprintf("DB %s %s", operation, table),
		trace.WithAttributes(
			attribute.String("db.operation", operation),
			attribute.String("db.table", table),
			attribute.String("db.system", "postgresql"),
		),
		trace.WithSpanKind(trace.SpanKindClient),
	)
}

// StartKafkaSpan starts a span for Kafka operations
func (th *TracingHelper) StartKafkaSpan(ctx context.Context, operation, topic string) (context.Context, trace.Span) {
	return th.StartSpan(ctx, fmt.Sprintf("Kafka %s %s", operation, topic),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.operation", operation),
			attribute.String("messaging.destination", topic),
		),
		trace.WithSpanKind(trace.SpanKindProducer),
	)
}

// StartAISpan starts a span for AI operations
func (th *TracingHelper) StartAISpan(ctx context.Context, operation, provider, model string) (context.Context, trace.Span) {
	return th.StartSpan(ctx, fmt.Sprintf("AI %s %s", provider, operation),
		trace.WithAttributes(
			attribute.String("ai.system", provider),
			attribute.String("ai.operation", operation),
			attribute.String("ai.model", model),
		),
		trace.WithSpanKind(trace.SpanKindClient),
	)
}

// StartExternalAPISpan starts a span for external API calls
func (th *TracingHelper) StartExternalAPISpan(ctx context.Context, service, operation string) (context.Context, trace.Span) {
	return th.StartSpan(ctx, fmt.Sprintf("API %s %s", service, operation),
		trace.WithAttributes(
			attribute.String("external.service", service),
			attribute.String("external.operation", operation),
		),
		trace.WithSpanKind(trace.SpanKindClient),
	)
}

// RecordError records an error in the span
func (th *TracingHelper) RecordError(span trace.Span, err error, message string) {
	if err != nil {
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error.message", message),
			attribute.String("error.type", fmt.Sprintf("%T", err)),
		))
		span.SetStatus(codes.Error, message)
	}
}

// RecordSuccess marks the span as successful
func (th *TracingHelper) RecordSuccess(span trace.Span, message string) {
	span.SetStatus(codes.Ok, message)
	span.SetAttributes(attribute.Bool("success", true))
}

// AddEvent adds an event to the span
func (th *TracingHelper) AddEvent(span trace.Span, name string, attrs ...attribute.KeyValue) {
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetAttributes sets attributes on the span
func (th *TracingHelper) SetAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	span.SetAttributes(attrs...)
}

// TraceFunction traces a function execution
func (th *TracingHelper) TraceFunction(ctx context.Context, fn func(context.Context) error) error {
	// Get caller function name
	pc, _, _, ok := runtime.Caller(1)
	var functionName string
	if ok {
		functionName = runtime.FuncForPC(pc).Name()
	} else {
		functionName = "unknown_function"
	}

	ctx, span := th.StartSpan(ctx, functionName)
	defer span.End()

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)

	// Add timing information
	span.SetAttributes(
		attribute.Int64("duration_ms", duration.Milliseconds()),
	)

	if err != nil {
		th.RecordError(span, err, "Function execution failed")
		return err
	}

	th.RecordSuccess(span, "Function executed successfully")
	return nil
}

// SpanFromContext extracts the span from context
func (th *TracingHelper) SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// ContextWithSpan returns a new context with the span
func (th *TracingHelper) ContextWithSpan(ctx context.Context, span trace.Span) context.Context {
	return trace.ContextWithSpan(ctx, span)
}

// GetTraceID returns the trace ID from the context
func (th *TracingHelper) GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID returns the span ID from the context
func (th *TracingHelper) GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// InjectTraceContext injects trace context into a map
func (th *TracingHelper) InjectTraceContext(ctx context.Context, carrier map[string]string) {
	// This would typically use otel.GetTextMapPropagator().Inject()
	// For now, we'll manually add trace information
	if traceID := th.GetTraceID(ctx); traceID != "" {
		carrier["trace-id"] = traceID
	}
	if spanID := th.GetSpanID(ctx); spanID != "" {
		carrier["span-id"] = spanID
	}
}

// ExtractTraceContext extracts trace context from a map
func (th *TracingHelper) ExtractTraceContext(ctx context.Context, carrier map[string]string) context.Context {
	// This would typically use otel.GetTextMapPropagator().Extract()
	// For now, we'll return the original context
	return ctx
}

// BusinessOperationTracer provides tracing for business operations
type BusinessOperationTracer struct {
	helper *TracingHelper
}

// NewBusinessOperationTracer creates a new business operation tracer
func NewBusinessOperationTracer(helper *TracingHelper) *BusinessOperationTracer {
	return &BusinessOperationTracer{
		helper: helper,
	}
}

// TraceBeverageInvention traces beverage invention process
func (bot *BusinessOperationTracer) TraceBeverageInvention(ctx context.Context, ingredients []string, theme string, useAI bool, fn func(context.Context) error) error {
	ctx, span := bot.helper.StartSpan(ctx, "beverage_invention",
		trace.WithAttributes(
			attribute.StringSlice("ingredients", ingredients),
			attribute.String("theme", theme),
			attribute.Bool("use_ai", useAI),
			attribute.String("business.operation", "beverage_invention"),
		),
	)
	defer span.End()

	bot.helper.AddEvent(span, "invention_started",
		attribute.String("event.type", "business.start"))

	err := fn(ctx)

	if err != nil {
		bot.helper.RecordError(span, err, "Beverage invention failed")
		bot.helper.AddEvent(span, "invention_failed",
			attribute.String("event.type", "business.error"),
			attribute.String("error.message", err.Error()))
		return err
	}

	bot.helper.RecordSuccess(span, "Beverage invention completed")
	bot.helper.AddEvent(span, "invention_completed",
		attribute.String("event.type", "business.success"))

	return nil
}

// TraceTaskCreation traces task creation process
func (bot *BusinessOperationTracer) TraceTaskCreation(ctx context.Context, taskType, assignee string, fn func(context.Context) (string, error)) (string, error) {
	ctx, span := bot.helper.StartSpan(ctx, "task_creation",
		trace.WithAttributes(
			attribute.String("task.type", taskType),
			attribute.String("task.assignee", assignee),
			attribute.String("business.operation", "task_creation"),
		),
	)
	defer span.End()

	bot.helper.AddEvent(span, "task_creation_started",
		attribute.String("event.type", "business.start"))

	taskID, err := fn(ctx)

	if err != nil {
		bot.helper.RecordError(span, err, "Task creation failed")
		bot.helper.AddEvent(span, "task_creation_failed",
			attribute.String("event.type", "business.error"),
			attribute.String("error.message", err.Error()))
		return "", err
	}

	bot.helper.RecordSuccess(span, "Task creation completed")
	bot.helper.SetAttributes(span, attribute.String("task.id", taskID))
	bot.helper.AddEvent(span, "task_creation_completed",
		attribute.String("event.type", "business.success"),
		attribute.String("task.id", taskID))

	return taskID, nil
}

// TraceNotificationSending traces notification sending process
func (bot *BusinessOperationTracer) TraceNotificationSending(ctx context.Context, channel, recipient string, fn func(context.Context) error) error {
	ctx, span := bot.helper.StartSpan(ctx, "notification_sending",
		trace.WithAttributes(
			attribute.String("notification.channel", channel),
			attribute.String("notification.recipient", recipient),
			attribute.String("business.operation", "notification_sending"),
		),
	)
	defer span.End()

	bot.helper.AddEvent(span, "notification_sending_started",
		attribute.String("event.type", "business.start"))

	err := fn(ctx)

	if err != nil {
		bot.helper.RecordError(span, err, "Notification sending failed")
		bot.helper.AddEvent(span, "notification_sending_failed",
			attribute.String("event.type", "business.error"),
			attribute.String("error.message", err.Error()))
		return err
	}

	bot.helper.RecordSuccess(span, "Notification sent successfully")
	bot.helper.AddEvent(span, "notification_sending_completed",
		attribute.String("event.type", "business.success"))

	return nil
}

// Global tracing helper instance
var globalTracingHelper *TracingHelper

// InitGlobalTracing initializes the global tracing helper
func InitGlobalTracing(scope *InstrumentationScope) {
	globalTracingHelper = NewTracingHelper(scope)
}

// GetGlobalTracing returns the global tracing helper
func GetGlobalTracing() *TracingHelper {
	return globalTracingHelper
}

// Context utility functions

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetTraceID extracts trace ID from context
func GetTraceID(ctx context.Context) string {
	if traceID := ctx.Value(traceIDKey); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

// GetSpanFromContext extracts span from context (placeholder)
func GetSpanFromContext(ctx context.Context) trace.Span {
	// TODO: Implement actual span extraction from OpenTelemetry context
	return nil
}

// Attribute creates an attribute for tracing
func Attribute(key string, value interface{}) attribute.KeyValue {
	switch v := value.(type) {
	case string:
		return attribute.String(key, v)
	case int:
		return attribute.Int(key, v)
	case int64:
		return attribute.Int64(key, v)
	case float64:
		return attribute.Float64(key, v)
	case bool:
		return attribute.Bool(key, v)
	default:
		return attribute.String(key, fmt.Sprintf("%v", v))
	}
}
