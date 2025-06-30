package services

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go-coffee-ai-agents/orchestration-engine/internal/common"
	"go-coffee-ai-agents/orchestration-engine/internal/config"
	"go-coffee-ai-agents/orchestration-engine/internal/infrastructure/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ObservabilityService provides comprehensive observability and monitoring
type ObservabilityService struct {
	telemetryManager       *observability.TelemetryManager
	kafkaInstrumentation   *observability.KafkaInstrumentation
	databaseInstrumentation *observability.DatabaseInstrumentation
	
	config                 *config.ObservabilityConfig
	logger                 common.Logger
	
	// Observability metrics
	observabilityMetrics   *ObservabilityMetrics
	
	// Health monitoring
	healthChecks           map[string]HealthCheck
	
	// Control
	mutex                  sync.RWMutex
	stopCh                 chan struct{}
}

// ObservabilityMetrics tracks observability-related metrics
type ObservabilityMetrics struct {
	TracesGenerated        int64     `json:"traces_generated"`
	SpansGenerated         int64     `json:"spans_generated"`
	MetricsCollected       int64     `json:"metrics_collected"`
	ErrorsRecorded         int64     `json:"errors_recorded"`
	HealthChecksPerformed  int64     `json:"health_checks_performed"`
	HealthCheckFailures    int64     `json:"health_check_failures"`
	InstrumentationErrors  int64     `json:"instrumentation_errors"`
	LastTraceGenerated     time.Time `json:"last_trace_generated"`
	LastMetricCollected    time.Time `json:"last_metric_collected"`
	LastHealthCheck        time.Time `json:"last_health_check"`
	SystemHealth           string    `json:"system_health"`
	LastUpdated            time.Time `json:"last_updated"`
}

// HealthCheck represents a health check function
type HealthCheck func(ctx context.Context) error

// ObservabilityReport represents a comprehensive observability report
type ObservabilityReport struct {
	Timestamp              time.Time                    `json:"timestamp"`
	SystemHealth           string                       `json:"system_health"`
	ObservabilityMetrics   *ObservabilityMetrics        `json:"observability_metrics"`
	ComponentHealth        map[string]string            `json:"component_health"`
	TracingStatus          *TracingStatus               `json:"tracing_status"`
	MetricsStatus          *MetricsStatus               `json:"metrics_status"`
	Recommendations        []string                     `json:"recommendations"`
}

// TracingStatus represents the status of distributed tracing
type TracingStatus struct {
	Enabled                bool      `json:"enabled"`
	SamplingRatio          float64   `json:"sampling_ratio"`
	TracesPerSecond        float64   `json:"traces_per_second"`
	AverageSpanDuration    time.Duration `json:"average_span_duration"`
	ErrorRate              float64   `json:"error_rate"`
	LastExport             time.Time `json:"last_export"`
}

// MetricsStatus represents the status of metrics collection
type MetricsStatus struct {
	Enabled                bool      `json:"enabled"`
	MetricsPerSecond       float64   `json:"metrics_per_second"`
	ActiveInstruments      int       `json:"active_instruments"`
	LastExport             time.Time `json:"last_export"`
}

// NewObservabilityService creates a new observability service
func NewObservabilityService(
	telemetryManager *observability.TelemetryManager,
	config *config.ObservabilityConfig,
	logger common.Logger,
) *ObservabilityService {
	
	os := &ObservabilityService{
		telemetryManager: telemetryManager,
		config:          config,
		logger:          logger,
		observabilityMetrics: &ObservabilityMetrics{
			LastUpdated: time.Now(),
		},
		healthChecks: make(map[string]HealthCheck),
		stopCh:       make(chan struct{}),
	}

	// Initialize instrumentation components
	os.kafkaInstrumentation = observability.NewKafkaInstrumentation(telemetryManager, logger)
	os.databaseInstrumentation = observability.NewDatabaseInstrumentation(telemetryManager, logger)

	// Register default health checks
	os.registerDefaultHealthChecks()

	return os
}

// Start starts the observability service
func (os *ObservabilityService) Start(ctx context.Context) error {
	os.logger.Info("Starting observability service")

	// Start monitoring routines
	go os.metricsCollectionLoop(ctx)
	go os.healthMonitoringLoop(ctx)

	os.logger.Info("Observability service started")
	return nil
}

// Stop stops the observability service
func (os *ObservabilityService) Stop(ctx context.Context) error {
	os.logger.Info("Stopping observability service")
	
	close(os.stopCh)
	
	// Shutdown telemetry manager
	if os.telemetryManager != nil {
		if err := os.telemetryManager.Shutdown(ctx); err != nil {
			os.logger.Error("Failed to shutdown telemetry manager", err)
		}
	}
	
	os.logger.Info("Observability service stopped")
	return nil
}

// TraceWorkflow creates a trace for workflow execution
func (os *ObservabilityService) TraceWorkflow(ctx context.Context, workflowID, workflowType string) (context.Context, trace.Span) {
	spanName := fmt.Sprintf("workflow.execute %s", workflowType)
	ctx, span := os.telemetryManager.StartSpan(ctx, spanName,
		trace.WithAttributes(
			attribute.String("workflow.id", workflowID),
			attribute.String("workflow.type", workflowType),
			attribute.String("component", "orchestration"),
		),
	)

	// Update metrics
	os.mutex.Lock()
	os.observabilityMetrics.TracesGenerated++
	os.observabilityMetrics.SpansGenerated++
	os.observabilityMetrics.LastTraceGenerated = time.Now()
	os.observabilityMetrics.LastUpdated = time.Now()
	os.mutex.Unlock()

	return ctx, span
}

// TraceAgentCall creates a trace for agent calls
func (os *ObservabilityService) TraceAgentCall(ctx context.Context, agentType, operation string) (context.Context, trace.Span) {
	spanName := fmt.Sprintf("agent.call %s.%s", agentType, operation)
	ctx, span := os.telemetryManager.StartSpan(ctx, spanName,
		trace.WithAttributes(
			attribute.String("agent.type", agentType),
			attribute.String("agent.operation", operation),
			attribute.String("component", "agent"),
		),
	)

	// Update metrics
	os.mutex.Lock()
	os.observabilityMetrics.SpansGenerated++
	os.observabilityMetrics.LastUpdated = time.Now()
	os.mutex.Unlock()

	return ctx, span
}

// RecordWorkflowMetrics records workflow execution metrics
func (os *ObservabilityService) RecordWorkflowMetrics(ctx context.Context, workflowType, status string, duration time.Duration) {
	os.telemetryManager.RecordWorkflow(ctx, workflowType, status, duration)

	// Update internal metrics
	os.mutex.Lock()
	os.observabilityMetrics.MetricsCollected++
	os.observabilityMetrics.LastMetricCollected = time.Now()
	os.observabilityMetrics.LastUpdated = time.Now()
	os.mutex.Unlock()
}

// RecordAgentMetrics records agent call metrics
func (os *ObservabilityService) RecordAgentMetrics(ctx context.Context, agentType, operation, status string, duration time.Duration) {
	os.telemetryManager.RecordAgentCall(ctx, agentType, operation, status, duration)

	// Update internal metrics
	os.mutex.Lock()
	os.observabilityMetrics.MetricsCollected++
	os.observabilityMetrics.LastMetricCollected = time.Now()
	os.observabilityMetrics.LastUpdated = time.Now()
	os.mutex.Unlock()
}

// RecordError records an error with tracing and metrics
func (os *ObservabilityService) RecordError(ctx context.Context, err error, component string) {
	// Record error in span
	os.telemetryManager.RecordSpanError(ctx, err)

	// Record error metric
	os.telemetryManager.RecordError(ctx, err.Error(), component)

	// Update internal metrics
	os.mutex.Lock()
	os.observabilityMetrics.ErrorsRecorded++
	os.observabilityMetrics.LastUpdated = time.Now()
	os.mutex.Unlock()

	os.logger.Error("Error recorded in observability", err, "component", component)
}

// AddSpanAttributes adds attributes to the current span
func (os *ObservabilityService) AddSpanAttributes(ctx context.Context, attributes map[string]interface{}) {
	var otelAttributes []attribute.KeyValue
	for key, value := range attributes {
		switch v := value.(type) {
		case string:
			otelAttributes = append(otelAttributes, attribute.String(key, v))
		case int:
			otelAttributes = append(otelAttributes, attribute.Int(key, v))
		case int64:
			otelAttributes = append(otelAttributes, attribute.Int64(key, v))
		case float64:
			otelAttributes = append(otelAttributes, attribute.Float64(key, v))
		case bool:
			otelAttributes = append(otelAttributes, attribute.Bool(key, v))
		default:
			otelAttributes = append(otelAttributes, attribute.String(key, fmt.Sprintf("%v", v)))
		}
	}

	os.telemetryManager.AddSpanAttributes(ctx, otelAttributes...)
}

// AddSpanEvent adds an event to the current span
func (os *ObservabilityService) AddSpanEvent(ctx context.Context, name string, attributes map[string]interface{}) {
	var otelAttributes []attribute.KeyValue
	for key, value := range attributes {
		switch v := value.(type) {
		case string:
			otelAttributes = append(otelAttributes, attribute.String(key, v))
		case int:
			otelAttributes = append(otelAttributes, attribute.Int(key, v))
		case int64:
			otelAttributes = append(otelAttributes, attribute.Int64(key, v))
		case float64:
			otelAttributes = append(otelAttributes, attribute.Float64(key, v))
		case bool:
			otelAttributes = append(otelAttributes, attribute.Bool(key, v))
		default:
			otelAttributes = append(otelAttributes, attribute.String(key, fmt.Sprintf("%v", v)))
		}
	}

	os.telemetryManager.AddSpanEvent(ctx, name, otelAttributes...)
}

// GetKafkaInstrumentation returns the Kafka instrumentation
func (os *ObservabilityService) GetKafkaInstrumentation() *observability.KafkaInstrumentation {
	return os.kafkaInstrumentation
}

// GetDatabaseInstrumentation returns the database instrumentation
func (os *ObservabilityService) GetDatabaseInstrumentation() *observability.DatabaseInstrumentation {
	return os.databaseInstrumentation
}

// GetHTTPMiddleware returns the HTTP instrumentation middleware
func (os *ObservabilityService) GetHTTPMiddleware() func(http.Handler) http.Handler {
	return os.telemetryManager.InstrumentationMiddleware
}

// RegisterHealthCheck registers a health check
func (os *ObservabilityService) RegisterHealthCheck(name string, check HealthCheck) {
	os.mutex.Lock()
	defer os.mutex.Unlock()
	
	os.healthChecks[name] = check
	os.logger.Info("Health check registered", "name", name)
}

// GetObservabilityReport generates a comprehensive observability report
func (os *ObservabilityService) GetObservabilityReport(ctx context.Context) (*ObservabilityReport, error) {
	os.mutex.RLock()
	metrics := *os.observabilityMetrics
	os.mutex.RUnlock()

	// Perform health checks
	componentHealth := os.performHealthChecks(ctx)

	// Calculate system health
	systemHealth := os.calculateSystemHealth(componentHealth)

	// Generate tracing status
	tracingStatus := &TracingStatus{
		Enabled:       os.telemetryManager != nil,
		SamplingRatio: 1.0, // This would come from config
		LastExport:    time.Now(),
	}

	// Generate metrics status
	metricsStatus := &MetricsStatus{
		Enabled:    os.telemetryManager != nil,
		LastExport: time.Now(),
	}

	// Generate recommendations
	recommendations := os.generateRecommendations(&metrics, componentHealth)

	report := &ObservabilityReport{
		Timestamp:            time.Now(),
		SystemHealth:         systemHealth,
		ObservabilityMetrics: &metrics,
		ComponentHealth:      componentHealth,
		TracingStatus:        tracingStatus,
		MetricsStatus:        metricsStatus,
		Recommendations:      recommendations,
	}

	return report, nil
}

// GetObservabilityMetrics returns current observability metrics
func (os *ObservabilityService) GetObservabilityMetrics() *ObservabilityMetrics {
	os.mutex.RLock()
	defer os.mutex.RUnlock()
	
	metricsCopy := *os.observabilityMetrics
	return &metricsCopy
}

// metricsCollectionLoop collects observability metrics
func (os *ObservabilityService) metricsCollectionLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-os.stopCh:
			return
		case <-ticker.C:
			os.collectMetrics(ctx)
		}
	}
}

// healthMonitoringLoop monitors system health
func (os *ObservabilityService) healthMonitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-os.stopCh:
			return
		case <-ticker.C:
			os.performHealthMonitoring(ctx)
		}
	}
}

// collectMetrics collects current observability metrics
func (os *ObservabilityService) collectMetrics(ctx context.Context) {
	os.mutex.Lock()
	defer os.mutex.Unlock()

	// Update system health
	os.observabilityMetrics.SystemHealth = os.calculateSystemHealth(nil)
	os.observabilityMetrics.LastUpdated = time.Now()
}

// performHealthMonitoring performs health monitoring
func (os *ObservabilityService) performHealthMonitoring(ctx context.Context) {
	componentHealth := os.performHealthChecks(ctx)
	
	os.mutex.Lock()
	os.observabilityMetrics.HealthChecksPerformed += int64(len(componentHealth))
	
	// Count failures
	failures := 0
	for _, health := range componentHealth {
		if health != "healthy" {
			failures++
		}
	}
	os.observabilityMetrics.HealthCheckFailures += int64(failures)
	os.observabilityMetrics.LastHealthCheck = time.Now()
	os.observabilityMetrics.LastUpdated = time.Now()
	os.mutex.Unlock()

	if failures > 0 {
		os.logger.Warn("Health check failures detected", "failures", failures, "total", len(componentHealth))
	}
}

// performHealthChecks performs all registered health checks
func (os *ObservabilityService) performHealthChecks(ctx context.Context) map[string]string {
	os.mutex.RLock()
	checks := make(map[string]HealthCheck)
	for name, check := range os.healthChecks {
		checks[name] = check
	}
	os.mutex.RUnlock()

	results := make(map[string]string)
	for name, check := range checks {
		if err := check(ctx); err != nil {
			results[name] = "unhealthy"
			os.logger.Error("Health check failed", err, "component", name)
		} else {
			results[name] = "healthy"
		}
	}

	return results
}

// registerDefaultHealthChecks registers default health checks
func (os *ObservabilityService) registerDefaultHealthChecks() {
	// Telemetry health check
	os.RegisterHealthCheck("telemetry", func(ctx context.Context) error {
		if os.telemetryManager == nil {
			return fmt.Errorf("telemetry manager not initialized")
		}
		return os.telemetryManager.HealthCheck(ctx)
	})

	// Observability service health check
	os.RegisterHealthCheck("observability_service", func(ctx context.Context) error {
		// Check if metrics are being collected
		if time.Since(os.observabilityMetrics.LastUpdated) > 5*time.Minute {
			return fmt.Errorf("metrics not updated recently")
		}
		return nil
	})
}

// calculateSystemHealth calculates overall system health
func (os *ObservabilityService) calculateSystemHealth(componentHealth map[string]string) string {
	if componentHealth == nil {
		return "unknown"
	}

	healthyCount := 0
	totalCount := len(componentHealth)

	for _, health := range componentHealth {
		if health == "healthy" {
			healthyCount++
		}
	}

	if totalCount == 0 {
		return "unknown"
	}

	healthPercentage := float64(healthyCount) / float64(totalCount) * 100

	if healthPercentage >= 95 {
		return "excellent"
	} else if healthPercentage >= 80 {
		return "good"
	} else if healthPercentage >= 60 {
		return "fair"
	} else if healthPercentage >= 40 {
		return "poor"
	} else {
		return "critical"
	}
}

// generateRecommendations generates observability recommendations
func (os *ObservabilityService) generateRecommendations(metrics *ObservabilityMetrics, componentHealth map[string]string) []string {
	var recommendations []string

	// Check trace generation
	if time.Since(metrics.LastTraceGenerated) > time.Hour {
		recommendations = append(recommendations, "No traces generated recently. Verify tracing is enabled and working correctly.")
	}

	// Check metric collection
	if time.Since(metrics.LastMetricCollected) > time.Hour {
		recommendations = append(recommendations, "No metrics collected recently. Verify metrics collection is enabled and working correctly.")
	}

	// Check error rate
	if metrics.ErrorsRecorded > 100 {
		recommendations = append(recommendations, "High error rate detected. Review error logs and implement error reduction strategies.")
	}

	// Check health check failures
	if metrics.HealthCheckFailures > 10 {
		recommendations = append(recommendations, "Multiple health check failures detected. Review system components and dependencies.")
	}

	// Check component health
	unhealthyComponents := 0
	for _, health := range componentHealth {
		if health != "healthy" {
			unhealthyComponents++
		}
	}

	if unhealthyComponents > 0 {
		recommendations = append(recommendations, fmt.Sprintf("%d components are unhealthy. Review component status and resolve issues.", unhealthyComponents))
	}

	return recommendations
}

// Health checks the health of the observability service
func (os *ObservabilityService) Health(ctx context.Context) error {
	// Check if telemetry manager is healthy
	if os.telemetryManager != nil {
		if err := os.telemetryManager.HealthCheck(ctx); err != nil {
			return fmt.Errorf("telemetry manager health check failed: %w", err)
		}
	}

	// Check if metrics are being updated
	if time.Since(os.observabilityMetrics.LastUpdated) > 10*time.Minute {
		return fmt.Errorf("observability metrics not updated recently")
	}

	return nil
}
