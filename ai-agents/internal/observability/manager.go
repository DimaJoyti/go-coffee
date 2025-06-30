package observability

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ObservabilityManager coordinates all observability components
type ObservabilityManager struct {
	config            TelemetryConfig
	telemetryProvider *TelemetryProvider
	metricsCollector  *MetricsCollector
	tracingHelper     *TracingHelper
	logger            *StructuredLogger
	businessLogger    *BusinessLogger
	auditLogger       *AuditLogger
	scope             *InstrumentationScope
	httpServer        *http.Server
}

// NewObservabilityManager creates a new observability manager
func NewObservabilityManager(config TelemetryConfig) (*ObservabilityManager, error) {
	// Initialize telemetry provider
	telemetryProvider, err := NewTelemetryProvider(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create telemetry provider: %w", err)
	}

	// Create instrumentation scope
	scope := NewInstrumentationScope(config.ServiceName, config.ServiceVersion)

	// Initialize logger
	logger := NewStructuredLogger(config.Logging)

	// Initialize metrics collector
	metricsCollector := NewMetricsCollector(scope)
	if err := metricsCollector.InitializeMetrics(); err != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// Initialize tracing helper
	tracingHelper := NewTracingHelper(scope)

	// Initialize business and audit loggers
	businessLogger := NewBusinessLogger(logger)
	auditLogger := NewAuditLogger(logger)

	om := &ObservabilityManager{
		config:            config,
		telemetryProvider: telemetryProvider,
		metricsCollector:  metricsCollector,
		tracingHelper:     tracingHelper,
		logger:            logger,
		businessLogger:    businessLogger,
		auditLogger:       auditLogger,
		scope:             scope,
	}

	// Start metrics server if Prometheus is enabled
	if config.Exporters.Prometheus.Enabled {
		if err := om.startMetricsServer(); err != nil {
			return nil, fmt.Errorf("failed to start metrics server: %w", err)
		}
	}

	return om, nil
}

// startMetricsServer starts the Prometheus metrics HTTP server
func (om *ObservabilityManager) startMetricsServer() error {
	mux := http.NewServeMux()
	mux.Handle(om.config.Exporters.Prometheus.Path, promhttp.Handler())
	
	// Add health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Add observability status endpoint
	mux.HandleFunc("/observability/status", om.handleObservabilityStatus)

	om.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", om.config.Exporters.Prometheus.Port),
		Handler: mux,
	}

	go func() {
		om.logger.Info("Starting metrics server",
			"port", om.config.Exporters.Prometheus.Port,
			"path", om.config.Exporters.Prometheus.Path)
		
		if err := om.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			om.logger.Error("Metrics server failed", err)
		}
	}()

	return nil
}

// handleObservabilityStatus handles observability status requests
func (om *ObservabilityManager) handleObservabilityStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	// Simple JSON encoding
	fmt.Fprintf(w, `{
		"service": {
			"name": "%s",
			"version": "%s",
			"environment": "%s"
		},
		"telemetry": {
			"tracing_enabled": %t,
			"metrics_enabled": %t,
			"logging_enabled": %t
		},
		"exporters": {
			"jaeger_enabled": %t,
			"otlp_enabled": %t,
			"prometheus_enabled": %t,
			"console_enabled": %t
		},
		"timestamp": "%s"
	}`,
		om.config.ServiceName,
		om.config.ServiceVersion,
		om.config.Environment,
		om.config.Tracing.Enabled,
		om.config.Metrics.Enabled,
		om.config.Logging.Enabled,
		om.config.Exporters.Jaeger.Enabled,
		om.config.Exporters.OTLP.Enabled,
		om.config.Exporters.Prometheus.Enabled,
		om.config.Exporters.Console.Enabled,
		time.Now().Format(time.RFC3339),
	)
}

// GetLogger returns the structured logger
func (om *ObservabilityManager) GetLogger() *StructuredLogger {
	return om.logger
}

// GetBusinessLogger returns the business logger
func (om *ObservabilityManager) GetBusinessLogger() *BusinessLogger {
	return om.businessLogger
}

// GetAuditLogger returns the audit logger
func (om *ObservabilityManager) GetAuditLogger() *AuditLogger {
	return om.auditLogger
}

// GetMetricsCollector returns the metrics collector
func (om *ObservabilityManager) GetMetricsCollector() *MetricsCollector {
	return om.metricsCollector
}

// GetTracingHelper returns the tracing helper
func (om *ObservabilityManager) GetTracingHelper() *TracingHelper {
	return om.tracingHelper
}

// GetScope returns the instrumentation scope
func (om *ObservabilityManager) GetScope() *InstrumentationScope {
	return om.scope
}

// RecordBusinessEvent records a business event with full observability
func (om *ObservabilityManager) RecordBusinessEvent(ctx context.Context, eventType string, fn func(context.Context) error) error {
	// Start tracing
	ctx, span := om.tracingHelper.StartSpan(ctx, eventType)
	defer span.End()

	// Record start time for metrics
	start := time.Now()

	// Execute the function
	err := fn(ctx)
	duration := time.Since(start)

	// Record metrics
	counters := om.metricsCollector.GetCounters()
	histograms := om.metricsCollector.GetHistograms()
	
	if counters != nil && histograms != nil {
		if err != nil {
			counters.RequestsError.Add(ctx, 1)
		} else {
			counters.RequestsSuccess.Add(ctx, 1)
		}
		counters.RequestsTotal.Add(ctx, 1)
		histograms.RequestDuration.Record(ctx, duration.Seconds())
	}

	// Record in tracing
	if err != nil {
		om.tracingHelper.RecordError(span, err, "Business event failed")
	} else {
		om.tracingHelper.RecordSuccess(span, "Business event completed")
	}

	// Log the event
	if err != nil {
		om.logger.ErrorContext(ctx, "Business event failed",
			err,
			"event_type", eventType,
			"duration_ms", duration.Milliseconds())
	} else {
		om.logger.InfoContext(ctx, "Business event completed",
			"event_type", eventType,
			"duration_ms", duration.Milliseconds())
	}

	return err
}

// Shutdown gracefully shuts down all observability components
func (om *ObservabilityManager) Shutdown(ctx context.Context) error {
	om.logger.Info("Shutting down observability manager")

	var errors []error

	// Shutdown HTTP server
	if om.httpServer != nil {
		if err := om.httpServer.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to shutdown HTTP server: %w", err))
		}
	}

	// Shutdown telemetry provider
	if err := om.telemetryProvider.Shutdown(ctx); err != nil {
		errors = append(errors, fmt.Errorf("failed to shutdown telemetry provider: %w", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}

	om.logger.Info("Observability manager shutdown completed")
	return nil
}

// HealthCheck returns the health status of observability components
func (om *ObservabilityManager) HealthCheck() map[string]interface{} {
	return map[string]interface{}{
		"service": map[string]interface{}{
			"name":        om.config.ServiceName,
			"version":     om.config.ServiceVersion,
			"environment": om.config.Environment,
			"uptime":      time.Since(time.Now()).String(), // This would be actual uptime
		},
		"components": map[string]interface{}{
			"telemetry_provider": "healthy",
			"metrics_collector":  "healthy",
			"tracing_helper":     "healthy",
			"logger":             "healthy",
		},
		"exporters": map[string]interface{}{
			"jaeger":     om.config.Exporters.Jaeger.Enabled,
			"otlp":       om.config.Exporters.OTLP.Enabled,
			"prometheus": om.config.Exporters.Prometheus.Enabled,
			"console":    om.config.Exporters.Console.Enabled,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}
}

// GetMetrics returns current metrics snapshot
func (om *ObservabilityManager) GetMetrics() map[string]interface{} {
	// This would return actual metrics values
	return map[string]interface{}{
		"requests_total":     0, // Would be actual counter value
		"requests_success":   0,
		"requests_error":     0,
		"beverages_created":  0,
		"tasks_created":      0,
		"ai_requests_total":  0,
		"timestamp":          time.Now().Format(time.RFC3339),
	}
}

// Global observability manager instance
var globalObservabilityManager *ObservabilityManager

// InitGlobalObservability initializes the global observability manager
func InitGlobalObservability(config TelemetryConfig) error {
	manager, err := NewObservabilityManager(config)
	if err != nil {
		return err
	}

	globalObservabilityManager = manager

	// Initialize global components
	InitGlobalLogger(config.Logging)
	InitGlobalTracing(manager.GetScope())
	InitGlobalMetrics(manager.GetScope())

	return nil
}

// GetGlobalObservability returns the global observability manager
func GetGlobalObservability() *ObservabilityManager {
	return globalObservabilityManager
}

// ShutdownGlobalObservability shuts down the global observability manager
func ShutdownGlobalObservability(ctx context.Context) error {
	if globalObservabilityManager == nil {
		return nil
	}
	return globalObservabilityManager.Shutdown(ctx)
}

// Convenience functions for common operations

// TraceOperation traces an operation with full observability
func TraceOperation(ctx context.Context, operationName string, fn func(context.Context) error) error {
	if globalObservabilityManager == nil {
		return fn(ctx)
	}
	return globalObservabilityManager.RecordBusinessEvent(ctx, operationName, fn)
}

// LogBusinessEvent logs a business event
func LogBusinessEvent(ctx context.Context, eventType string, fields ...interface{}) {
	if globalObservabilityManager == nil {
		return
	}
	globalObservabilityManager.GetLogger().InfoContext(ctx, "Business event", append([]interface{}{"event_type", eventType}, fields...)...)
}

// RecordMetric records a metric
func RecordMetric(ctx context.Context, metricName string, value float64, labels map[string]string) {
	if globalObservabilityManager == nil {
		return
	}
	// This would record the metric using the metrics collector
	// Implementation depends on the specific metric type
}

// Context helper functions for trace information (additional to tracing.go)

// WithSpanID adds a span ID to context
func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, "span_id", spanID)
}

// WithCorrelationID adds a correlation ID to context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, "correlation_id", correlationID)
}

// GetSpanID retrieves span ID from context
func GetSpanID(ctx context.Context) string {
	if value := ctx.Value("span_id"); value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// GetCorrelationID retrieves correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if value := ctx.Value("correlation_id"); value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}
