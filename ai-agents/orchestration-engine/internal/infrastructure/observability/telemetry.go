package observability

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go-coffee-ai-agents/orchestration-engine/internal/common"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"
)

// TelemetryManager manages OpenTelemetry instrumentation
type TelemetryManager struct {
	tracer         trace.Tracer
	meter          metric.Meter
	traceProvider  *sdktrace.TracerProvider
	metricProvider *sdkmetric.MeterProvider
	config         *TelemetryConfig
	logger         common.Logger
	
	// Metrics
	requestCounter    metric.Int64Counter
	requestDuration   metric.Float64Histogram
	errorCounter      metric.Int64Counter
	workflowCounter   metric.Int64Counter
	agentCallCounter  metric.Int64Counter
	agentDuration     metric.Float64Histogram
	
	// Custom metrics
	customMetrics     map[string]interface{}
}

// TelemetryConfig contains telemetry configuration
type TelemetryConfig struct {
	ServiceName        string            `json:"service_name"`
	ServiceVersion     string            `json:"service_version"`
	Environment        string            `json:"environment"`
	
	// Tracing configuration
	TracingEnabled     bool              `json:"tracing_enabled"`
	JaegerEndpoint     string            `json:"jaeger_endpoint"`
	SamplingRatio      float64           `json:"sampling_ratio"`
	
	// Metrics configuration
	MetricsEnabled     bool              `json:"metrics_enabled"`
	PrometheusEndpoint string            `json:"prometheus_endpoint"`
	MetricsPort        int               `json:"metrics_port"`
	
	// Resource attributes
	ResourceAttributes map[string]string `json:"resource_attributes"`
	
	// Instrumentation configuration
	InstrumentHTTP     bool              `json:"instrument_http"`
	InstrumentGRPC     bool              `json:"instrument_grpc"`
	InstrumentDB       bool              `json:"instrument_db"`
	InstrumentKafka    bool              `json:"instrument_kafka"`
	
	// Advanced settings
	BatchTimeout       time.Duration     `json:"batch_timeout"`
	MaxBatchSize       int               `json:"max_batch_size"`
	MaxQueueSize       int               `json:"max_queue_size"`
}


// NewTelemetryManager creates a new telemetry manager
func NewTelemetryManager(config *TelemetryConfig, logger common.Logger) (*TelemetryManager, error) {
	if config == nil {
		config = DefaultTelemetryConfig()
	}

	tm := &TelemetryManager{
		config:        config,
		logger:        logger,
		customMetrics: make(map[string]interface{}),
	}

	// Initialize resource
	resource, err := tm.createResource()
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Initialize tracing
	if config.TracingEnabled {
		if err := tm.initializeTracing(resource); err != nil {
			return nil, fmt.Errorf("failed to initialize tracing: %w", err)
		}
	}

	// Initialize metrics
	if config.MetricsEnabled {
		if err := tm.initializeMetrics(resource); err != nil {
			return nil, fmt.Errorf("failed to initialize metrics: %w", err)
		}
	}

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Initialize standard metrics
	if err := tm.initializeStandardMetrics(); err != nil {
		return nil, fmt.Errorf("failed to initialize standard metrics: %w", err)
	}

	logger.Info("Telemetry manager initialized",
		"service", config.ServiceName,
		"version", config.ServiceVersion,
		"tracing_enabled", config.TracingEnabled,
		"metrics_enabled", config.MetricsEnabled)

	return tm, nil
}

// DefaultTelemetryConfig returns default telemetry configuration
func DefaultTelemetryConfig() *TelemetryConfig {
	return &TelemetryConfig{
		ServiceName:        "orchestration-engine",
		ServiceVersion:     "1.0.0",
		Environment:        "development",
		TracingEnabled:     true,
		JaegerEndpoint:     "http://localhost:14268/api/traces",
		SamplingRatio:      1.0,
		MetricsEnabled:     true,
		PrometheusEndpoint: "http://localhost:9090",
		MetricsPort:        8888,
		ResourceAttributes: map[string]string{
			"deployment.environment": "development",
			"service.namespace":      "ai-agents",
		},
		InstrumentHTTP:     true,
		InstrumentGRPC:     true,
		InstrumentDB:       true,
		InstrumentKafka:    true,
		BatchTimeout:       5 * time.Second,
		MaxBatchSize:       512,
		MaxQueueSize:       2048,
	}
}

// createResource creates an OpenTelemetry resource
func (tm *TelemetryManager) createResource() (*resource.Resource, error) {
	attributes := []attribute.KeyValue{
		semconv.ServiceName(tm.config.ServiceName),
		semconv.ServiceVersion(tm.config.ServiceVersion),
		semconv.DeploymentEnvironment(tm.config.Environment),
	}

	// Add custom resource attributes
	for key, value := range tm.config.ResourceAttributes {
		attributes = append(attributes, attribute.String(key, value))
	}

	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			attributes...,
		),
	)
}

// initializeTracing initializes distributed tracing
func (tm *TelemetryManager) initializeTracing(res *resource.Resource) error {
	// Create Jaeger exporter
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
		jaeger.WithEndpoint(tm.config.JaegerEndpoint),
	))
	if err != nil {
		return fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	// Create trace provider
	tm.traceProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(tm.config.BatchTimeout),
			sdktrace.WithMaxExportBatchSize(tm.config.MaxBatchSize),
			sdktrace.WithMaxQueueSize(tm.config.MaxQueueSize),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(tm.config.SamplingRatio)),
	)

	// Set global trace provider
	otel.SetTracerProvider(tm.traceProvider)

	// Create tracer
	tm.tracer = otel.Tracer(
		tm.config.ServiceName,
		trace.WithInstrumentationVersion(tm.config.ServiceVersion),
		trace.WithSchemaURL(semconv.SchemaURL),
	)

	return nil
}

// initializeMetrics initializes metrics collection
func (tm *TelemetryManager) initializeMetrics(res *resource.Resource) error {
	// Create Prometheus exporter
	exporter, err := prometheus.New()
	if err != nil {
		return fmt.Errorf("failed to create Prometheus exporter: %w", err)
	}

	// Create metric provider
	tm.metricProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(exporter),
	)

	// Set global metric provider
	otel.SetMeterProvider(tm.metricProvider)

	// Create meter
	tm.meter = otel.Meter(
		tm.config.ServiceName,
		metric.WithInstrumentationVersion(tm.config.ServiceVersion),
		metric.WithSchemaURL(semconv.SchemaURL),
	)

	return nil
}

// initializeStandardMetrics creates standard application metrics
func (tm *TelemetryManager) initializeStandardMetrics() error {
	if !tm.config.MetricsEnabled {
		return nil
	}

	var err error

	// Request metrics
	tm.requestCounter, err = tm.meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("failed to create request counter: %w", err)
	}

	tm.requestDuration, err = tm.meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("failed to create request duration histogram: %w", err)
	}

	// Error metrics
	tm.errorCounter, err = tm.meter.Int64Counter(
		"errors_total",
		metric.WithDescription("Total number of errors"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("failed to create error counter: %w", err)
	}

	// Workflow metrics
	tm.workflowCounter, err = tm.meter.Int64Counter(
		"workflows_total",
		metric.WithDescription("Total number of workflows executed"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("failed to create workflow counter: %w", err)
	}

	// Agent metrics
	tm.agentCallCounter, err = tm.meter.Int64Counter(
		"agent_calls_total",
		metric.WithDescription("Total number of agent calls"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("failed to create agent call counter: %w", err)
	}

	tm.agentDuration, err = tm.meter.Float64Histogram(
		"agent_call_duration_seconds",
		metric.WithDescription("Agent call duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("failed to create agent duration histogram: %w", err)
	}

	return nil
}

// StartSpan starts a new trace span
func (tm *TelemetryManager) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if !tm.config.TracingEnabled {
		return ctx, trace.SpanFromContext(ctx)
	}
	return tm.tracer.Start(ctx, name, opts...)
}

// RecordHTTPRequest records HTTP request metrics
func (tm *TelemetryManager) RecordHTTPRequest(ctx context.Context, method, path, status string, duration time.Duration) {
	if !tm.config.MetricsEnabled {
		return
	}

	attributes := []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.route", path),
		attribute.String("http.status_code", status),
	}

	tm.requestCounter.Add(ctx, 1, metric.WithAttributes(attributes...))
	tm.requestDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attributes...))
}

// RecordError records error metrics
func (tm *TelemetryManager) RecordError(ctx context.Context, errorType, component string) {
	if !tm.config.MetricsEnabled {
		return
	}

	attributes := []attribute.KeyValue{
		attribute.String("error.type", errorType),
		attribute.String("component", component),
	}

	tm.errorCounter.Add(ctx, 1, metric.WithAttributes(attributes...))
}

// RecordWorkflow records workflow execution metrics
func (tm *TelemetryManager) RecordWorkflow(ctx context.Context, workflowType, status string, duration time.Duration) {
	if !tm.config.MetricsEnabled {
		return
	}

	attributes := []attribute.KeyValue{
		attribute.String("workflow.type", workflowType),
		attribute.String("workflow.status", status),
	}

	tm.workflowCounter.Add(ctx, 1, metric.WithAttributes(attributes...))
}

// RecordAgentCall records agent call metrics
func (tm *TelemetryManager) RecordAgentCall(ctx context.Context, agentType, operation, status string, duration time.Duration) {
	if !tm.config.MetricsEnabled {
		return
	}

	attributes := []attribute.KeyValue{
		attribute.String("agent.type", agentType),
		attribute.String("agent.operation", operation),
		attribute.String("agent.status", status),
	}

	tm.agentCallCounter.Add(ctx, 1, metric.WithAttributes(attributes...))
	tm.agentDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attributes...))
}

// AddSpanAttributes adds attributes to the current span
func (tm *TelemetryManager) AddSpanAttributes(ctx context.Context, attributes ...attribute.KeyValue) {
	if !tm.config.TracingEnabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attributes...)
}

// AddSpanEvent adds an event to the current span
func (tm *TelemetryManager) AddSpanEvent(ctx context.Context, name string, attributes ...attribute.KeyValue) {
	if !tm.config.TracingEnabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attributes...))
}

// RecordSpanError records an error in the current span
func (tm *TelemetryManager) RecordSpanError(ctx context.Context, err error) {
	if !tm.config.TracingEnabled || err == nil {
		return
	}

	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// CreateCustomCounter creates a custom counter metric
func (tm *TelemetryManager) CreateCustomCounter(name, description, unit string) (metric.Int64Counter, error) {
	if !tm.config.MetricsEnabled {
		return nil, fmt.Errorf("metrics not enabled")
	}

	counter, err := tm.meter.Int64Counter(
		name,
		metric.WithDescription(description),
		metric.WithUnit(unit),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create counter %s: %w", name, err)
	}

	tm.customMetrics[name] = counter
	return counter, nil
}

// CreateCustomHistogram creates a custom histogram metric
func (tm *TelemetryManager) CreateCustomHistogram(name, description, unit string) (metric.Float64Histogram, error) {
	if !tm.config.MetricsEnabled {
		return nil, fmt.Errorf("metrics not enabled")
	}

	histogram, err := tm.meter.Float64Histogram(
		name,
		metric.WithDescription(description),
		metric.WithUnit(unit),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create histogram %s: %w", name, err)
	}

	tm.customMetrics[name] = histogram
	return histogram, nil
}

// CreateCustomGauge creates a custom gauge metric
func (tm *TelemetryManager) CreateCustomGauge(name, description, unit string) (metric.Float64ObservableGauge, error) {
	if !tm.config.MetricsEnabled {
		return nil, fmt.Errorf("metrics not enabled")
	}

	gauge, err := tm.meter.Float64ObservableGauge(
		name,
		metric.WithDescription(description),
		metric.WithUnit(unit),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gauge %s: %w", name, err)
	}

	tm.customMetrics[name] = gauge
	return gauge, nil
}

// GetTracer returns the tracer instance
func (tm *TelemetryManager) GetTracer() trace.Tracer {
	return tm.tracer
}

// GetMeter returns the meter instance
func (tm *TelemetryManager) GetMeter() metric.Meter {
	return tm.meter
}

// Shutdown gracefully shuts down the telemetry manager
func (tm *TelemetryManager) Shutdown(ctx context.Context) error {
	var err error

	if tm.traceProvider != nil {
		if shutdownErr := tm.traceProvider.Shutdown(ctx); shutdownErr != nil {
			err = fmt.Errorf("failed to shutdown trace provider: %w", shutdownErr)
		}
	}

	if tm.metricProvider != nil {
		if shutdownErr := tm.metricProvider.Shutdown(ctx); shutdownErr != nil {
			if err != nil {
				err = fmt.Errorf("%w; failed to shutdown metric provider: %w", err, shutdownErr)
			} else {
				err = fmt.Errorf("failed to shutdown metric provider: %w", shutdownErr)
			}
		}
	}

	if err != nil {
		tm.logger.Error("Error during telemetry shutdown", err)
		return err
	}

	tm.logger.Info("Telemetry manager shutdown successfully")
	return nil
}

// InstrumentationMiddleware provides HTTP instrumentation middleware
func (tm *TelemetryManager) InstrumentationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !tm.config.TracingEnabled {
			next.ServeHTTP(w, r)
			return
		}

		// Extract trace context from headers
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		// Start span
		spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		ctx, span := tm.StartSpan(ctx, spanName,
			trace.WithAttributes(
				semconv.HTTPMethod(r.Method),
				semconv.HTTPRoute(r.URL.Path),
				semconv.HTTPScheme(r.URL.Scheme),
				attribute.String("http.host", r.Host),
				semconv.UserAgentOriginal(r.UserAgent()),
				semconv.ClientAddress(r.RemoteAddr),
			),
		)
		defer span.End()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		// Record start time
		start := time.Now()

		// Continue with request
		next.ServeHTTP(wrapped, r.WithContext(ctx))

		// Record metrics
		duration := time.Since(start)
		tm.RecordHTTPRequest(ctx, r.Method, r.URL.Path, fmt.Sprintf("%d", wrapped.statusCode), duration)

		// Update span with response information
		span.SetAttributes(
			semconv.HTTPStatusCode(wrapped.statusCode),
			semconv.HTTPResponseBodySize(int(wrapped.bytesWritten)),
		)

		// Set span status based on HTTP status code
		if wrapped.statusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", wrapped.statusCode))
		}
	})
}

// responseWriter wraps http.ResponseWriter to capture response information
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// HealthCheck performs a health check of the telemetry system
func (tm *TelemetryManager) HealthCheck(ctx context.Context) error {
	// Create a test span to verify tracing is working
	if tm.config.TracingEnabled {
		_, span := tm.StartSpan(ctx, "telemetry.health_check")
		span.SetAttributes(attribute.String("health_check", "telemetry"))
		span.End()
	}

	// Record a test metric to verify metrics are working
	if tm.config.MetricsEnabled {
		tm.RecordError(ctx, "health_check", "telemetry")
	}

	return nil
}
