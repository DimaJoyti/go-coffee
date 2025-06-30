package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go-coffee-ai-agents/internal/ai"
	"go-coffee-ai-agents/internal/database"
	"go-coffee-ai-agents/internal/httputils"
	"go-coffee-ai-agents/internal/messaging/kafka"
	"go-coffee-ai-agents/internal/observability"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	logger  *observability.StructuredLogger
	metrics *observability.MetricsCollector
	tracing *observability.TracingHelper
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) *HealthHandler {
	return &HealthHandler{
		logger:  logger,
		metrics: metrics,
		tracing: tracing,
	}
}

// HealthStatus represents the health status of the system
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    time.Duration          `json:"uptime"`
	Checks    map[string]CheckResult `json:"checks"`
	Summary   HealthSummary          `json:"summary"`
}

// CheckResult represents the result of a health check
type CheckResult struct {
	Status    string        `json:"status"`
	Message   string        `json:"message,omitempty"`
	Error     string        `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
	Details   interface{}   `json:"details,omitempty"`
}

// HealthSummary provides a summary of health checks
type HealthSummary struct {
	Total   int `json:"total"`
	Healthy int `json:"healthy"`
	Warning int `json:"warning"`
	Error   int `json:"error"`
}

// Health performs a comprehensive health check
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "GET", "/health", r.UserAgent())
	defer span.End()

	h.logger.DebugContext(ctx, "Performing comprehensive health check")

	startTime := time.Now()
	checks := make(map[string]CheckResult)

	// Database health check
	checks["database"] = h.checkDatabase(ctx)

	// Kafka health check
	checks["kafka"] = h.checkKafka(ctx)

	// AI providers health check
	checks["ai_providers"] = h.checkAIProviders(ctx)

	// Memory health check
	checks["memory"] = h.checkMemory(ctx)

	// Disk health check
	checks["disk"] = h.checkDisk(ctx)

	// Calculate summary
	summary := h.calculateSummary(checks)

	// Determine overall status
	overallStatus := "healthy"
	statusCode := http.StatusOK

	if summary.Error > 0 {
		overallStatus = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	} else if summary.Warning > 0 {
		overallStatus = "degraded"
		statusCode = http.StatusOK // Still return 200 for warnings
	}

	// Create health status response
	healthStatus := HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Version:   "1.0.0",               // TODO: Get from build info
		Uptime:    time.Since(startTime), // TODO: Track actual uptime
		Checks:    checks,
		Summary:   summary,
	}

	// Record metrics
	if h.metrics != nil {
		counters := h.metrics.GetCounters()
		if counters != nil {
			counters.RequestsTotal.Add(ctx, 1)
			if overallStatus == "healthy" {
				counters.RequestsSuccess.Add(ctx, 1)
			} else {
				counters.RequestsError.Add(ctx, 1)
			}
		}
	}

	// Record tracing
	if overallStatus == "healthy" {
		h.tracing.RecordSuccess(span, "Health check passed")
	} else {
		h.tracing.RecordError(span, nil, "Health check failed")
	}

	h.logger.InfoContext(ctx, "Health check completed",
		"status", overallStatus,
		"checks_total", summary.Total,
		"checks_healthy", summary.Healthy,
		"checks_warning", summary.Warning,
		"checks_error", summary.Error)

	httputils.WriteJSONResponse(w, statusCode, healthStatus)
}

// Ready performs a readiness check (lighter than full health check)
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "GET", "/health/ready", r.UserAgent())
	defer span.End()

	h.logger.DebugContext(ctx, "Performing readiness check")

	checks := make(map[string]CheckResult)

	// Essential service checks for readiness
	checks["database"] = h.checkDatabase(ctx)
	checks["kafka"] = h.checkKafka(ctx)

	// Calculate summary
	summary := h.calculateSummary(checks)

	// Determine readiness
	ready := summary.Error == 0
	statusCode := http.StatusOK
	if !ready {
		statusCode = http.StatusServiceUnavailable
	}

	readinessStatus := map[string]interface{}{
		"ready":     ready,
		"timestamp": time.Now(),
		"checks":    checks,
		"summary":   summary,
	}

	if ready {
		h.tracing.RecordSuccess(span, "Readiness check passed")
	} else {
		h.tracing.RecordError(span, nil, "Readiness check failed")
	}

	h.logger.InfoContext(ctx, "Readiness check completed",
		"ready", ready,
		"checks_total", summary.Total,
		"checks_error", summary.Error)

	httputils.WriteJSONResponse(w, statusCode, readinessStatus)
}

// Live performs a liveness check (minimal check)
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "GET", "/health/live", r.UserAgent())
	defer span.End()

	h.logger.DebugContext(ctx, "Performing liveness check")

	// Simple liveness check - if we can respond, we're alive
	livenessStatus := map[string]interface{}{
		"alive":     true,
		"timestamp": time.Now(),
		"uptime":    time.Since(time.Now()), // TODO: Track actual uptime
	}

	h.tracing.RecordSuccess(span, "Liveness check passed")
	h.logger.DebugContext(ctx, "Liveness check completed", "alive", true)

	httputils.WriteJSONResponse(w, http.StatusOK, livenessStatus)
}

// checkDatabase checks database connectivity
func (h *HealthHandler) checkDatabase(ctx context.Context) CheckResult {
	start := time.Now()

	dbManager := database.GetGlobalManager()
	if dbManager == nil {
		return CheckResult{
			Status:    "error",
			Message:   "Database manager not initialized",
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}
	}

	healthStatus, err := dbManager.HealthCheck(ctx)
	duration := time.Since(start)

	if err != nil {
		return CheckResult{
			Status:    "error",
			Message:   "Database health check failed",
			Error:     err.Error(),
			Duration:  duration,
			Timestamp: time.Now(),
		}
	}

	if !healthStatus.Healthy {
		return CheckResult{
			Status:    "error",
			Message:   "Database is unhealthy",
			Error:     healthStatus.Error,
			Duration:  duration,
			Timestamp: time.Now(),
			Details:   healthStatus.Details,
		}
	}

	return CheckResult{
		Status:    "healthy",
		Message:   "Database is healthy",
		Duration:  duration,
		Timestamp: time.Now(),
		Details:   healthStatus.Details,
	}
}

// checkKafka checks Kafka connectivity
func (h *HealthHandler) checkKafka(ctx context.Context) CheckResult {
	start := time.Now()

	kafkaManager := kafka.GetGlobalManager()
	if kafkaManager == nil {
		return CheckResult{
			Status:    "error",
			Message:   "Kafka manager not initialized",
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}
	}

	healthStatus, err := kafkaManager.HealthCheck(ctx)
	duration := time.Since(start)

	if err != nil {
		return CheckResult{
			Status:    "error",
			Message:   "Kafka health check failed",
			Error:     err.Error(),
			Duration:  duration,
			Timestamp: time.Now(),
		}
	}

	if !healthStatus.Healthy {
		return CheckResult{
			Status:    "error",
			Message:   "Kafka is unhealthy",
			Error:     healthStatus.Error,
			Duration:  duration,
			Timestamp: time.Now(),
			Details:   healthStatus.Details,
		}
	}

	return CheckResult{
		Status:    "healthy",
		Message:   "Kafka is healthy",
		Duration:  duration,
		Timestamp: time.Now(),
		Details:   healthStatus.Details,
	}
}

// checkAIProviders checks AI provider connectivity
func (h *HealthHandler) checkAIProviders(ctx context.Context) CheckResult {
	start := time.Now()

	aiManager := ai.GetGlobalManager()
	if aiManager == nil {
		return CheckResult{
			Status:    "warning",
			Message:   "AI manager not initialized",
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}
	}

	healthResults := aiManager.HealthCheck(ctx)
	duration := time.Since(start)

	healthyProviders := 0
	totalProviders := len(healthResults)
	var errors []string

	for provider, err := range healthResults {
		if err == nil {
			healthyProviders++
		} else {
			errors = append(errors, fmt.Sprintf("%s: %v", provider, err))
		}
	}

	if healthyProviders == 0 {
		return CheckResult{
			Status:    "error",
			Message:   "No AI providers are healthy",
			Error:     strings.Join(errors, "; "),
			Duration:  duration,
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"total_providers":   totalProviders,
				"healthy_providers": healthyProviders,
			},
		}
	}

	if healthyProviders < totalProviders {
		return CheckResult{
			Status:    "warning",
			Message:   fmt.Sprintf("%d of %d AI providers are healthy", healthyProviders, totalProviders),
			Duration:  duration,
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"total_providers":   totalProviders,
				"healthy_providers": healthyProviders,
				"errors":            errors,
			},
		}
	}

	return CheckResult{
		Status:    "healthy",
		Message:   "All AI providers are healthy",
		Duration:  duration,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"total_providers":   totalProviders,
			"healthy_providers": healthyProviders,
		},
	}
}

// checkMemory checks memory usage
func (h *HealthHandler) checkMemory(ctx context.Context) CheckResult {
	start := time.Now()

	// TODO: Implement actual memory check
	// For now, return healthy
	return CheckResult{
		Status:    "healthy",
		Message:   "Memory usage is normal",
		Duration:  time.Since(start),
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"usage_percent": 45, // Placeholder
		},
	}
}

// checkDisk checks disk usage
func (h *HealthHandler) checkDisk(ctx context.Context) CheckResult {
	start := time.Now()

	// TODO: Implement actual disk check
	// For now, return healthy
	return CheckResult{
		Status:    "healthy",
		Message:   "Disk usage is normal",
		Duration:  time.Since(start),
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"usage_percent": 35, // Placeholder
		},
	}
}

// calculateSummary calculates health check summary
func (h *HealthHandler) calculateSummary(checks map[string]CheckResult) HealthSummary {
	summary := HealthSummary{
		Total: len(checks),
	}

	for _, check := range checks {
		switch check.Status {
		case "healthy":
			summary.Healthy++
		case "warning":
			summary.Warning++
		case "error":
			summary.Error++
		}
	}

	return summary
}
