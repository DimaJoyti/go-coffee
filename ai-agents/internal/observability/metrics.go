package observability

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// MetricsCollector manages application metrics
type MetricsCollector struct {
	meter   metric.Meter
	metrics map[string]interface{}
	mutex   sync.RWMutex
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(scope *InstrumentationScope) *MetricsCollector {
	return &MetricsCollector{
		meter:   scope.Meter,
		metrics: make(map[string]interface{}),
	}
}

// Counter metrics
type CounterMetrics struct {
	RequestsTotal         metric.Int64Counter
	RequestsSuccess       metric.Int64Counter
	RequestsError         metric.Int64Counter
	BeveragesCreated      metric.Int64Counter
	TasksCreated          metric.Int64Counter
	NotificationsSent     metric.Int64Counter
	AIRequestsTotal       metric.Int64Counter
	AIRequestsSuccess     metric.Int64Counter
	AIRequestsError       metric.Int64Counter
	KafkaMessagesProduced metric.Int64Counter
	KafkaMessagesConsumed metric.Int64Counter
	KafkaMessagesTotal    metric.Int64Counter
	KafkaMessagesSuccess  metric.Int64Counter
	KafkaMessagesError    metric.Int64Counter
	DatabaseOperations    metric.Int64Counter
	CircuitBreakerTrips   metric.Int64Counter
	RetryAttempts         metric.Int64Counter
}

// Histogram metrics
type HistogramMetrics struct {
	RequestDuration       metric.Float64Histogram
	AIRequestDuration     metric.Float64Histogram
	DatabaseQueryDuration metric.Float64Histogram
	KafkaPublishDuration  metric.Float64Histogram
	KafkaConsumeDuration  metric.Float64Histogram
	TaskCreationDuration  metric.Float64Histogram
	BeverageGenDuration   metric.Float64Histogram
}

// Gauge metrics
type GaugeMetrics struct {
	ActiveConnections   metric.Int64UpDownCounter
	QueueSize           metric.Int64UpDownCounter
	CircuitBreakerState metric.Int64UpDownCounter
	RateLimitTokens     metric.Int64UpDownCounter
	MemoryUsage         metric.Int64UpDownCounter
	GoroutineCount      metric.Int64UpDownCounter
}

// InitializeMetrics initializes all application metrics
func (mc *MetricsCollector) InitializeMetrics() error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// Initialize counter metrics
	counters, err := mc.initializeCounters()
	if err != nil {
		return fmt.Errorf("failed to initialize counters: %w", err)
	}
	mc.metrics["counters"] = counters

	// Initialize histogram metrics
	histograms, err := mc.initializeHistograms()
	if err != nil {
		return fmt.Errorf("failed to initialize histograms: %w", err)
	}
	mc.metrics["histograms"] = histograms

	// Initialize gauge metrics
	gauges, err := mc.initializeGauges()
	if err != nil {
		return fmt.Errorf("failed to initialize gauges: %w", err)
	}
	mc.metrics["gauges"] = gauges

	return nil
}

// initializeCounters creates all counter metrics
func (mc *MetricsCollector) initializeCounters() (*CounterMetrics, error) {
	counters := &CounterMetrics{}

	var err error

	// Request metrics
	counters.RequestsTotal, err = mc.meter.Int64Counter(
		"requests_total",
		metric.WithDescription("Total number of requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	counters.RequestsSuccess, err = mc.meter.Int64Counter(
		"requests_success_total",
		metric.WithDescription("Total number of successful requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	counters.RequestsError, err = mc.meter.Int64Counter(
		"requests_error_total",
		metric.WithDescription("Total number of failed requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	// Business metrics
	counters.BeveragesCreated, err = mc.meter.Int64Counter(
		"beverages_created_total",
		metric.WithDescription("Total number of beverages created"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	counters.TasksCreated, err = mc.meter.Int64Counter(
		"tasks_created_total",
		metric.WithDescription("Total number of tasks created"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	counters.NotificationsSent, err = mc.meter.Int64Counter(
		"notifications_sent_total",
		metric.WithDescription("Total number of notifications sent"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	// AI metrics
	counters.AIRequestsTotal, err = mc.meter.Int64Counter(
		"ai_requests_total",
		metric.WithDescription("Total number of AI requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	counters.AIRequestsSuccess, err = mc.meter.Int64Counter(
		"ai_requests_success_total",
		metric.WithDescription("Total number of successful AI requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	counters.AIRequestsError, err = mc.meter.Int64Counter(
		"ai_requests_error_total",
		metric.WithDescription("Total number of failed AI requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	// Infrastructure metrics
	counters.KafkaMessagesProduced, err = mc.meter.Int64Counter(
		"kafka_messages_produced_total",
		metric.WithDescription("Total number of Kafka messages produced"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	counters.KafkaMessagesConsumed, err = mc.meter.Int64Counter(
		"kafka_messages_consumed_total",
		metric.WithDescription("Total number of Kafka messages consumed"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	counters.KafkaMessagesTotal, err = mc.meter.Int64Counter(
		"kafka_messages_total",
		metric.WithDescription("Total number of Kafka messages processed"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	counters.KafkaMessagesSuccess, err = mc.meter.Int64Counter(
		"kafka_messages_success_total",
		metric.WithDescription("Total number of successful Kafka message operations"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	counters.KafkaMessagesError, err = mc.meter.Int64Counter(
		"kafka_messages_error_total",
		metric.WithDescription("Total number of failed Kafka message operations"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	counters.DatabaseOperations, err = mc.meter.Int64Counter(
		"database_operations_total",
		metric.WithDescription("Total number of database operations"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	// Resilience metrics
	counters.CircuitBreakerTrips, err = mc.meter.Int64Counter(
		"circuit_breaker_trips_total",
		metric.WithDescription("Total number of circuit breaker trips"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	counters.RetryAttempts, err = mc.meter.Int64Counter(
		"retry_attempts_total",
		metric.WithDescription("Total number of retry attempts"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	return counters, nil
}

// initializeHistograms creates all histogram metrics
func (mc *MetricsCollector) initializeHistograms() (*HistogramMetrics, error) {
	histograms := &HistogramMetrics{}

	var err error

	// Request duration
	histograms.RequestDuration, err = mc.meter.Float64Histogram(
		"request_duration_seconds",
		metric.WithDescription("Request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	// AI request duration
	histograms.AIRequestDuration, err = mc.meter.Float64Histogram(
		"ai_request_duration_seconds",
		metric.WithDescription("AI request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	// Database query duration
	histograms.DatabaseQueryDuration, err = mc.meter.Float64Histogram(
		"database_query_duration_seconds",
		metric.WithDescription("Database query duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	// Kafka publish duration
	histograms.KafkaPublishDuration, err = mc.meter.Float64Histogram(
		"kafka_publish_duration_seconds",
		metric.WithDescription("Kafka publish duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	// Kafka consume duration
	histograms.KafkaConsumeDuration, err = mc.meter.Float64Histogram(
		"kafka_consume_duration_seconds",
		metric.WithDescription("Kafka consume duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	// Task creation duration
	histograms.TaskCreationDuration, err = mc.meter.Float64Histogram(
		"task_creation_duration_seconds",
		metric.WithDescription("Task creation duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	// Beverage generation duration
	histograms.BeverageGenDuration, err = mc.meter.Float64Histogram(
		"beverage_generation_duration_seconds",
		metric.WithDescription("Beverage generation duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	return histograms, nil
}

// initializeGauges creates all gauge metrics
func (mc *MetricsCollector) initializeGauges() (*GaugeMetrics, error) {
	gauges := &GaugeMetrics{}

	var err error

	// Active connections
	gauges.ActiveConnections, err = mc.meter.Int64UpDownCounter(
		"active_connections",
		metric.WithDescription("Number of active connections"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	// Queue size
	gauges.QueueSize, err = mc.meter.Int64UpDownCounter(
		"queue_size",
		metric.WithDescription("Current queue size"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	// Circuit breaker state
	gauges.CircuitBreakerState, err = mc.meter.Int64UpDownCounter(
		"circuit_breaker_state",
		metric.WithDescription("Circuit breaker state (0=closed, 1=open, 2=half-open)"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	// Rate limit tokens
	gauges.RateLimitTokens, err = mc.meter.Int64UpDownCounter(
		"rate_limit_tokens",
		metric.WithDescription("Current rate limit tokens"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	// Memory usage
	gauges.MemoryUsage, err = mc.meter.Int64UpDownCounter(
		"memory_usage_bytes",
		metric.WithDescription("Current memory usage in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, err
	}

	// Goroutine count
	gauges.GoroutineCount, err = mc.meter.Int64UpDownCounter(
		"goroutine_count",
		metric.WithDescription("Current number of goroutines"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	return gauges, nil
}

// GetCounters returns the counter metrics
func (mc *MetricsCollector) GetCounters() *CounterMetrics {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	if counters, ok := mc.metrics["counters"].(*CounterMetrics); ok {
		return counters
	}
	return nil
}

// GetHistograms returns the histogram metrics
func (mc *MetricsCollector) GetHistograms() *HistogramMetrics {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	if histograms, ok := mc.metrics["histograms"].(*HistogramMetrics); ok {
		return histograms
	}
	return nil
}

// GetGauges returns the gauge metrics
func (mc *MetricsCollector) GetGauges() *GaugeMetrics {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	if gauges, ok := mc.metrics["gauges"].(*GaugeMetrics); ok {
		return gauges
	}
	return nil
}

// RecordRequest records a request metric
func (mc *MetricsCollector) RecordRequest(ctx context.Context, method, endpoint string, duration time.Duration, success bool) {
	counters := mc.GetCounters()
	histograms := mc.GetHistograms()

	if counters == nil || histograms == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("method", method),
		attribute.String("endpoint", endpoint),
	}

	// Record total requests
	counters.RequestsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))

	// Record success/error
	if success {
		counters.RequestsSuccess.Add(ctx, 1, metric.WithAttributes(attrs...))
	} else {
		counters.RequestsError.Add(ctx, 1, metric.WithAttributes(attrs...))
	}

	// Record duration
	histograms.RequestDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

// RecordBeverageCreated records a beverage creation metric
func (mc *MetricsCollector) RecordBeverageCreated(ctx context.Context, theme string, aiUsed bool, duration time.Duration) {
	counters := mc.GetCounters()
	histograms := mc.GetHistograms()

	if counters == nil || histograms == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("theme", theme),
		attribute.Bool("ai_used", aiUsed),
	}

	counters.BeveragesCreated.Add(ctx, 1, metric.WithAttributes(attrs...))
	histograms.BeverageGenDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

// RecordAIRequest records an AI request metric
func (mc *MetricsCollector) RecordAIRequest(ctx context.Context, provider, operation string, duration time.Duration, success bool) {
	counters := mc.GetCounters()
	histograms := mc.GetHistograms()

	if counters == nil || histograms == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("provider", provider),
		attribute.String("operation", operation),
	}

	counters.AIRequestsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))

	if success {
		counters.AIRequestsSuccess.Add(ctx, 1, metric.WithAttributes(attrs...))
	} else {
		counters.AIRequestsError.Add(ctx, 1, metric.WithAttributes(attrs...))
	}

	histograms.AIRequestDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

// RecordKafkaMessage records a Kafka message metric
func (mc *MetricsCollector) RecordKafkaMessage(ctx context.Context, topic, operation string, duration time.Duration) {
	counters := mc.GetCounters()
	histograms := mc.GetHistograms()

	if counters == nil || histograms == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("topic", topic),
		attribute.String("operation", operation),
	}

	if operation == "produce" {
		counters.KafkaMessagesProduced.Add(ctx, 1, metric.WithAttributes(attrs...))
		histograms.KafkaPublishDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	} else if operation == "consume" {
		counters.KafkaMessagesConsumed.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}

// Global metrics collector instance
var globalMetricsCollector *MetricsCollector

// InitGlobalMetrics initializes the global metrics collector
func InitGlobalMetrics(scope *InstrumentationScope) error {
	globalMetricsCollector = NewMetricsCollector(scope)
	return globalMetricsCollector.InitializeMetrics()
}

// GetGlobalMetrics returns the global metrics collector
func GetGlobalMetrics() *MetricsCollector {
	return globalMetricsCollector
}
