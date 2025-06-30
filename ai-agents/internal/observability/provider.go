package observability

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// TelemetryProvider manages OpenTelemetry setup and lifecycle
type TelemetryProvider struct {
	config         TelemetryConfig
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	resource       *resource.Resource
	shutdownFuncs  []func(context.Context) error
}

// NewTelemetryProvider creates a new telemetry provider
func NewTelemetryProvider(config TelemetryConfig) (*TelemetryProvider, error) {
	provider := &TelemetryProvider{
		config:        config,
		shutdownFuncs: make([]func(context.Context) error, 0),
	}

	// Create resource
	res, err := provider.createResource()
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}
	provider.resource = res

	// Setup tracing
	if config.Tracing.Enabled {
		if err := provider.setupTracing(); err != nil {
			return nil, fmt.Errorf("failed to setup tracing: %w", err)
		}
	}

	// Setup metrics
	if config.Metrics.Enabled {
		if err := provider.setupMetrics(); err != nil {
			return nil, fmt.Errorf("failed to setup metrics: %w", err)
		}
	}

	// Setup propagation
	provider.setupPropagation()

	return provider, nil
}

// createResource creates an OpenTelemetry resource
func (tp *TelemetryProvider) createResource() (*resource.Resource, error) {
	attributes := []attribute.KeyValue{
		semconv.ServiceName(tp.config.ServiceName),
		semconv.ServiceVersion(tp.config.ServiceVersion),
		semconv.DeploymentEnvironment(tp.config.Environment),
		attribute.String("service.namespace", "go-coffee"),
	}

	// Add additional attributes from environment
	if hostname, err := os.Hostname(); err == nil {
		attributes = append(attributes, semconv.HostName(hostname))
	}

	if instanceID := os.Getenv("INSTANCE_ID"); instanceID != "" {
		attributes = append(attributes, semconv.ServiceInstanceID(instanceID))
	}

	if region := os.Getenv("AWS_REGION"); region != "" {
		attributes = append(attributes, semconv.CloudRegion(region))
	}

	if az := os.Getenv("AWS_AVAILABILITY_ZONE"); az != "" {
		attributes = append(attributes, semconv.CloudAvailabilityZone(az))
	}

	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			attributes...,
		),
	)
}

// setupTracing configures distributed tracing
func (tp *TelemetryProvider) setupTracing() error {
	var exporters []sdktrace.SpanExporter

	// Console exporter
	if tp.config.Exporters.Console.Enabled {
		exporter, err := stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			return fmt.Errorf("failed to create console exporter: %w", err)
		}
		exporters = append(exporters, exporter)
	}

	// Jaeger exporter
	if tp.config.Exporters.Jaeger.Enabled {
		exporter, err := jaeger.New(
			jaeger.WithCollectorEndpoint(
				jaeger.WithEndpoint(tp.config.Exporters.Jaeger.Endpoint),
			),
		)
		if err != nil {
			return fmt.Errorf("failed to create Jaeger exporter: %w", err)
		}
		exporters = append(exporters, exporter)
		tp.shutdownFuncs = append(tp.shutdownFuncs, exporter.Shutdown)
	}

	// OTLP exporter
	if tp.config.Exporters.OTLP.Enabled {
		ctx, cancel := context.WithTimeout(context.Background(), tp.config.Exporters.OTLP.Timeout)
		defer cancel()

		exporter, err := otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithEndpoint(tp.config.Exporters.OTLP.Endpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			return fmt.Errorf("failed to create OTLP exporter: %w", err)
		}
		exporters = append(exporters, exporter)
		tp.shutdownFuncs = append(tp.shutdownFuncs, exporter.Shutdown)
	}

	// Create span processors
	var spanProcessors []sdktrace.SpanProcessor
	for _, exporterInstance := range exporters {
		for _, processorConfig := range tp.config.Tracing.SpanProcessors {
			var processor sdktrace.SpanProcessor
			switch processorConfig.Type {
			case "batch":
				processor = sdktrace.NewBatchSpanProcessor(
					exporterInstance,
					sdktrace.WithBatchTimeout(processorConfig.Timeout),
					sdktrace.WithMaxExportBatchSize(processorConfig.BatchSize),
					sdktrace.WithMaxQueueSize(processorConfig.MaxQueue),
				)
			case "simple":
				processor = sdktrace.NewSimpleSpanProcessor(exporterInstance)
			default:
				processor = sdktrace.NewBatchSpanProcessor(exporterInstance)
			}
			spanProcessors = append(spanProcessors, processor)
		}
	}

	// Create tracer provider
	tp.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithResource(tp.resource),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(tp.config.Tracing.SamplingRate)),
		sdktrace.WithSpanLimits(sdktrace.SpanLimits{
			AttributeValueLengthLimit:   -1,
			AttributeCountLimit:         -1,
			EventCountLimit:             -1,
			LinkCountLimit:              -1,
			AttributePerEventCountLimit: -1,
			AttributePerLinkCountLimit:  -1,
		}),
	)

	// Add span processors
	for _, processor := range spanProcessors {
		tp.tracerProvider.RegisterSpanProcessor(processor)
	}

	// Set global tracer provider
	otel.SetTracerProvider(tp.tracerProvider)
	tp.shutdownFuncs = append(tp.shutdownFuncs, tp.tracerProvider.Shutdown)

	return nil
}

// setupMetrics configures metrics collection
func (tp *TelemetryProvider) setupMetrics() error {
	var readers []sdkmetric.Reader

	// Prometheus exporter
	if tp.config.Exporters.Prometheus.Enabled {
		exporter, err := prometheus.New()
		if err != nil {
			return fmt.Errorf("failed to create Prometheus exporter: %w", err)
		}
		readers = append(readers, exporter)
	}

	// Periodic reader for other exporters
	for _, readerConfig := range tp.config.Metrics.Readers {
		if readerConfig.Type == "periodic" {
			reader := sdkmetric.NewPeriodicReader(
				// Add exporters here based on configuration
				nil, // placeholder
				sdkmetric.WithInterval(readerConfig.Interval),
			)
			readers = append(readers, reader)
		}
	}

	// Create meter provider options
	meterOptions := []sdkmetric.Option{
		sdkmetric.WithResource(tp.resource),
	}

	for _, reader := range readers {
		meterOptions = append(meterOptions, sdkmetric.WithReader(reader))
	}

	tp.meterProvider = sdkmetric.NewMeterProvider(meterOptions...)

	// Set global meter provider
	otel.SetMeterProvider(tp.meterProvider)
	tp.shutdownFuncs = append(tp.shutdownFuncs, tp.meterProvider.Shutdown)

	return nil
}

// setupPropagation configures context propagation
func (tp *TelemetryProvider) setupPropagation() {
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)
}

// GetTracer returns a tracer for the given name
func (tp *TelemetryProvider) GetTracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return otel.Tracer(name, opts...)
}

// GetMeter returns a meter for the given name
func (tp *TelemetryProvider) GetMeter(name string, opts ...metric.MeterOption) metric.Meter {
	return otel.Meter(name, opts...)
}

// Shutdown gracefully shuts down the telemetry provider
func (tp *TelemetryProvider) Shutdown(ctx context.Context) error {
	var errors []error

	for _, shutdown := range tp.shutdownFuncs {
		if err := shutdown(ctx); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}

	return nil
}

// Global telemetry provider instance
var globalProvider *TelemetryProvider

// InitGlobalTelemetry initializes the global telemetry provider
func InitGlobalTelemetry(config TelemetryConfig) error {
	provider, err := NewTelemetryProvider(config)
	if err != nil {
		return err
	}

	globalProvider = provider
	log.Printf("Telemetry initialized for service: %s", config.ServiceName)
	return nil
}

// GetGlobalTracer returns a tracer from the global provider
func GetGlobalTracer(name string, opts ...trace.TracerOption) trace.Tracer {
	if globalProvider == nil {
		return otel.Tracer(name, opts...)
	}
	return globalProvider.GetTracer(name, opts...)
}

// GetGlobalMeter returns a meter from the global provider
func GetGlobalMeter(name string, opts ...metric.MeterOption) metric.Meter {
	if globalProvider == nil {
		return otel.Meter(name, opts...)
	}
	return globalProvider.GetMeter(name, opts...)
}

// ShutdownGlobalTelemetry shuts down the global telemetry provider
func ShutdownGlobalTelemetry(ctx context.Context) error {
	if globalProvider == nil {
		return nil
	}
	return globalProvider.Shutdown(ctx)
}

// InstrumentationScope represents an instrumentation scope
type InstrumentationScope struct {
	Name    string
	Version string
	Tracer  trace.Tracer
	Meter   metric.Meter
}

// NewInstrumentationScope creates a new instrumentation scope
func NewInstrumentationScope(name, version string) *InstrumentationScope {
	return &InstrumentationScope{
		Name:    name,
		Version: version,
		Tracer: GetGlobalTracer(name, trace.WithInstrumentationVersion(version)),
		Meter:  GetGlobalMeter(name, metric.WithInstrumentationVersion(version)),
	}
}

// StartSpan starts a new span
func (is *InstrumentationScope) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return is.Tracer.Start(ctx, name, opts...)
}

// RecordError records an error in the current span
func (is *InstrumentationScope) RecordError(span trace.Span, err error, opts ...trace.EventOption) {
	if err != nil {
		span.RecordError(err, opts...)
		span.SetStatus(codes.Error, err.Error())
	}
}

// SetSpanAttributes sets attributes on a span
func (is *InstrumentationScope) SetSpanAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	span.SetAttributes(attrs...)
}

// AddSpanEvent adds an event to a span
func (is *InstrumentationScope) AddSpanEvent(span trace.Span, name string, opts ...trace.EventOption) {
	span.AddEvent(name, opts...)
}
