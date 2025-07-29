package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// ObservabilityManager manages monitoring and observability
type ObservabilityManager struct {
	logger          *zap.Logger
	metricsRegistry *prometheus.Registry
	healthChecker   *HealthChecker
	traceCollector  *TraceCollector
	alertManager    *AlertManager
	dashboards      map[string]*Dashboard
	running         bool
	httpServer      *http.Server
	mutex           sync.RWMutex
}

// HealthChecker monitors system health
type HealthChecker struct {
	logger    *zap.Logger
	checks    map[string]HealthCheck
	status    HealthStatus
	lastCheck time.Time
	interval  time.Duration
	mutex     sync.RWMutex
}

// HealthCheck represents a health check
type HealthCheck interface {
	Name() string
	Check(ctx context.Context) HealthResult
}

// HealthResult represents the result of a health check
type HealthResult struct {
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message"`
	Latency   time.Duration          `json:"latency"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// HealthStatus represents health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// TraceCollector collects distributed traces
type TraceCollector struct {
	logger *zap.Logger
	traces map[string]*Trace
	mutex  sync.RWMutex
}

// AlertManager manages alerts and notifications
type AlertManager struct {
	logger *zap.Logger
	alerts map[string]*Alert
	rules  []AlertRule
	mutex  sync.RWMutex
}

// Alert represents an alert
type Alert struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Severity    AlertSeverity     `json:"severity"`
	Status      AlertStatus       `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"starts_at"`
	EndsAt      *time.Time        `json:"ends_at,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// AlertRule represents an alert rule
type AlertRule struct {
	Name        string            `json:"name"`
	Expression  string            `json:"expression"`
	Duration    time.Duration     `json:"duration"`
	Severity    AlertSeverity     `json:"severity"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
}

// AlertSeverity represents alert severity
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertStatus represents alert status
type AlertStatus string

const (
	AlertStatusFiring   AlertStatus = "firing"
	AlertStatusResolved AlertStatus = "resolved"
)

// Trace represents a distributed trace
type Trace struct {
	ID        string            `json:"id"`
	Operation string            `json:"operation"`
	StartTime time.Time         `json:"start_time"`
	EndTime   *time.Time        `json:"end_time,omitempty"`
	Duration  time.Duration     `json:"duration"`
	Spans     []Span            `json:"spans"`
	Tags      map[string]string `json:"tags"`
	Status    TraceStatus       `json:"status"`
}

// Span represents a trace span
type Span struct {
	ID        string            `json:"id"`
	TraceID   string            `json:"trace_id"`
	ParentID  string            `json:"parent_id,omitempty"`
	Operation string            `json:"operation"`
	StartTime time.Time         `json:"start_time"`
	EndTime   *time.Time        `json:"end_time,omitempty"`
	Duration  time.Duration     `json:"duration"`
	Tags      map[string]string `json:"tags"`
	Logs      []SpanLog         `json:"logs"`
	Status    SpanStatus        `json:"status"`
}

// SpanLog represents a span log entry
type SpanLog struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields"`
}

// TraceStatus represents trace status
type TraceStatus string

const (
	TraceStatusSuccess TraceStatus = "success"
	TraceStatusError   TraceStatus = "error"
	TraceStatusTimeout TraceStatus = "timeout"
)

// SpanStatus represents span status
type SpanStatus string

const (
	SpanStatusSuccess SpanStatus = "success"
	SpanStatusError   SpanStatus = "error"
)

// Dashboard represents a monitoring dashboard
type Dashboard struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Panels      []DashboardPanel `json:"panels"`
	Tags        []string         `json:"tags"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// DashboardPanel represents a dashboard panel
type DashboardPanel struct {
	ID            string                 `json:"id"`
	Title         string                 `json:"title"`
	Type          PanelType              `json:"type"`
	Query         string                 `json:"query"`
	Visualization map[string]interface{} `json:"visualization"`
	Position      PanelPosition          `json:"position"`
}

// PanelType represents panel type
type PanelType string

const (
	PanelTypeGraph   PanelType = "graph"
	PanelTypeTable   PanelType = "table"
	PanelTypeStat    PanelType = "stat"
	PanelTypeGauge   PanelType = "gauge"
	PanelTypeHeatmap PanelType = "heatmap"
)

// PanelPosition represents panel position
type PanelPosition struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Prometheus metrics
var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "defi_requests_total",
			Help: "Total number of DeFi requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "defi_request_duration_seconds",
			Help:    "Duration of DeFi requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	arbitrageOpportunities = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "defi_arbitrage_opportunities",
			Help: "Number of active arbitrage opportunities",
		},
		[]string{"chain", "protocol"},
	)

	yieldFarmingAPY = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "defi_yield_farming_apy",
			Help: "Current yield farming APY",
		},
		[]string{"protocol", "pool"},
	)

	tradingBotPerformance = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "defi_trading_bot_performance",
			Help: "Trading bot performance metrics",
		},
		[]string{"bot_id", "metric"},
	)

	securityAlerts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "defi_security_alerts_total",
			Help: "Total number of security alerts",
		},
		[]string{"severity", "category"},
	)
)

// NewObservabilityManager creates a new observability manager
func NewObservabilityManager(logger *zap.Logger, port int) *ObservabilityManager {
	registry := prometheus.NewRegistry()

	// Register metrics
	registry.MustRegister(requestsTotal)
	registry.MustRegister(requestDuration)
	registry.MustRegister(arbitrageOpportunities)
	registry.MustRegister(yieldFarmingAPY)
	registry.MustRegister(tradingBotPerformance)
	registry.MustRegister(securityAlerts)

	return &ObservabilityManager{
		logger:          logger,
		metricsRegistry: registry,
		healthChecker:   NewHealthChecker(logger),
		traceCollector:  NewTraceCollector(logger),
		alertManager:    NewAlertManager(logger),
		dashboards:      make(map[string]*Dashboard),
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: nil,
		},
	}
}

// Start starts the observability manager
func (om *ObservabilityManager) Start(ctx context.Context) error {
	om.mutex.Lock()
	defer om.mutex.Unlock()

	if om.running {
		return fmt.Errorf("observability manager already running")
	}

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(om.metricsRegistry, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", om.handleHealth)
	mux.HandleFunc("/traces", om.handleTraces)
	mux.HandleFunc("/dashboards", om.handleDashboards)

	om.httpServer.Handler = mux

	// Start health checker
	if err := om.healthChecker.Start(ctx); err != nil {
		return fmt.Errorf("failed to start health checker: %w", err)
	}

	// Start HTTP server
	go func() {
		if err := om.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			om.logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	om.running = true
	om.logger.Info("Observability manager started", zap.String("addr", om.httpServer.Addr))
	return nil
}

// Stop stops the observability manager
func (om *ObservabilityManager) Stop(ctx context.Context) error {
	om.mutex.Lock()
	defer om.mutex.Unlock()

	if !om.running {
		return nil
	}

	// Stop HTTP server
	if err := om.httpServer.Shutdown(ctx); err != nil {
		om.logger.Error("Failed to shutdown HTTP server", zap.Error(err))
	}

	// Stop health checker
	om.healthChecker.Stop()

	om.running = false
	om.logger.Info("Observability manager stopped")
	return nil
}

// RecordRequest records a request metric
func (om *ObservabilityManager) RecordRequest(method, endpoint, status string, duration time.Duration) {
	requestsTotal.WithLabelValues(method, endpoint, status).Inc()
	requestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordArbitrageOpportunity records arbitrage opportunity metric
func (om *ObservabilityManager) RecordArbitrageOpportunity(chain, protocol string, count float64) {
	arbitrageOpportunities.WithLabelValues(chain, protocol).Set(count)
}

// RecordYieldFarmingAPY records yield farming APY metric
func (om *ObservabilityManager) RecordYieldFarmingAPY(protocol, pool string, apy decimal.Decimal) {
	apyFloat, _ := apy.Float64()
	yieldFarmingAPY.WithLabelValues(protocol, pool).Set(apyFloat)
}

// RecordTradingBotPerformance records trading bot performance
func (om *ObservabilityManager) RecordTradingBotPerformance(botID, metric string, value float64) {
	tradingBotPerformance.WithLabelValues(botID, metric).Set(value)
}

// RecordSecurityAlert records security alert
func (om *ObservabilityManager) RecordSecurityAlert(severity, category string) {
	securityAlerts.WithLabelValues(severity, category).Inc()
}

// handleHealth handles health check requests
func (om *ObservabilityManager) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := om.healthChecker.GetStatus()

	w.Header().Set("Content-Type", "application/json")

	if status.Status == HealthStatusHealthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	// Return health status as JSON
	fmt.Fprintf(w, `{"status":"%s","timestamp":"%s"}`,
		status.Status, status.Timestamp.Format(time.RFC3339))
}

// handleTraces handles trace requests
func (om *ObservabilityManager) handleTraces(w http.ResponseWriter, r *http.Request) {
	traces := om.traceCollector.GetTraces()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Return traces as JSON (simplified)
	fmt.Fprintf(w, `{"traces":%d}`, len(traces))
}

// handleDashboards handles dashboard requests
func (om *ObservabilityManager) handleDashboards(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Return dashboard list (simplified)
	fmt.Fprintf(w, `{"dashboards":%d}`, len(om.dashboards))
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(logger *zap.Logger) *HealthChecker {
	return &HealthChecker{
		logger:   logger,
		checks:   make(map[string]HealthCheck),
		status:   HealthStatusHealthy,
		interval: time.Minute,
	}
}

// Start starts the health checker
func (hc *HealthChecker) Start(ctx context.Context) error {
	go hc.runChecks(ctx)
	hc.logger.Info("Health checker started")
	return nil
}

// Stop stops the health checker
func (hc *HealthChecker) Stop() {
	hc.logger.Info("Health checker stopped")
}

// runChecks runs health checks periodically
func (hc *HealthChecker) runChecks(ctx context.Context) {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hc.performChecks(ctx)
		}
	}
}

// performChecks performs all health checks
func (hc *HealthChecker) performChecks(ctx context.Context) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	overallStatus := HealthStatusHealthy

	for name, check := range hc.checks {
		result := check.Check(ctx)

		hc.logger.Debug("Health check completed",
			zap.String("check", name),
			zap.String("status", string(result.Status)),
			zap.Duration("latency", result.Latency),
		)

		// Determine overall status
		if result.Status == HealthStatusUnhealthy {
			overallStatus = HealthStatusUnhealthy
		} else if result.Status == HealthStatusDegraded && overallStatus == HealthStatusHealthy {
			overallStatus = HealthStatusDegraded
		}
	}

	hc.status = overallStatus
	hc.lastCheck = time.Now()
}

// AddCheck adds a health check
func (hc *HealthChecker) AddCheck(check HealthCheck) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	hc.checks[check.Name()] = check
	hc.logger.Info("Health check added", zap.String("name", check.Name()))
}

// GetStatus returns current health status
func (hc *HealthChecker) GetStatus() HealthResult {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	return HealthResult{
		Status:    hc.status,
		Message:   fmt.Sprintf("Overall system status: %s", hc.status),
		Timestamp: hc.lastCheck,
	}
}

// NewTraceCollector creates a new trace collector
func NewTraceCollector(logger *zap.Logger) *TraceCollector {
	return &TraceCollector{
		logger: logger,
		traces: make(map[string]*Trace),
	}
}

// StartTrace starts a new trace
func (tc *TraceCollector) StartTrace(operation string) *Trace {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	trace := &Trace{
		ID:        fmt.Sprintf("trace_%d", time.Now().UnixNano()),
		Operation: operation,
		StartTime: time.Now(),
		Spans:     []Span{},
		Tags:      make(map[string]string),
		Status:    TraceStatusSuccess,
	}

	tc.traces[trace.ID] = trace
	return trace
}

// GetTraces returns all traces
func (tc *TraceCollector) GetTraces() map[string]*Trace {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()

	// Return a copy
	traces := make(map[string]*Trace)
	for k, v := range tc.traces {
		traces[k] = v
	}
	return traces
}

// NewAlertManager creates a new alert manager
func NewAlertManager(logger *zap.Logger) *AlertManager {
	return &AlertManager{
		logger: logger,
		alerts: make(map[string]*Alert),
		rules:  []AlertRule{},
	}
}

// AddRule adds an alert rule
func (am *AlertManager) AddRule(rule AlertRule) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	am.rules = append(am.rules, rule)
	am.logger.Info("Alert rule added", zap.String("name", rule.Name))
}

// TriggerAlert triggers an alert
func (am *AlertManager) TriggerAlert(name, description string, severity AlertSeverity, labels map[string]string) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	alert := &Alert{
		ID:          fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		Name:        name,
		Description: description,
		Severity:    severity,
		Status:      AlertStatusFiring,
		Labels:      labels,
		Annotations: make(map[string]string),
		StartsAt:    time.Now(),
		UpdatedAt:   time.Now(),
	}

	am.alerts[alert.ID] = alert
	am.logger.Warn("Alert triggered",
		zap.String("id", alert.ID),
		zap.String("name", alert.Name),
		zap.String("severity", string(alert.Severity)),
	)
}

// ResolveAlert resolves an alert
func (am *AlertManager) ResolveAlert(alertID string) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if alert, exists := am.alerts[alertID]; exists {
		now := time.Now()
		alert.Status = AlertStatusResolved
		alert.EndsAt = &now
		alert.UpdatedAt = now

		am.logger.Info("Alert resolved",
			zap.String("id", alertID),
			zap.String("name", alert.Name),
		)
	}
}

// GetActiveAlerts returns all active alerts
func (am *AlertManager) GetActiveAlerts() []*Alert {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var activeAlerts []*Alert
	for _, alert := range am.alerts {
		if alert.Status == AlertStatusFiring {
			activeAlerts = append(activeAlerts, alert)
		}
	}
	return activeAlerts
}
