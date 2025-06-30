package monitoring

import (
	"context"
	"runtime"
	"sync"
	"time"
)

// PerformanceMonitor tracks system and application performance metrics
type PerformanceMonitor struct {
	metrics    *PerformanceMetrics
	collectors []MetricCollector
	logger     Logger
	mutex      sync.RWMutex
	stopCh     chan struct{}
	interval   time.Duration
}

// Logger interface for logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
}

// PerformanceMetrics represents comprehensive performance metrics
type PerformanceMetrics struct {
	// System metrics
	CPUUsage        float64   `json:"cpu_usage"`
	MemoryUsage     float64   `json:"memory_usage"`
	MemoryAllocated uint64    `json:"memory_allocated"`
	MemoryTotal     uint64    `json:"memory_total"`
	GoroutineCount  int       `json:"goroutine_count"`
	GCPauseTime     time.Duration `json:"gc_pause_time"`
	GCCount         uint32    `json:"gc_count"`
	
	// Application metrics
	RequestCount        int64         `json:"request_count"`
	RequestRate         float64       `json:"request_rate"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	ErrorRate           float64       `json:"error_rate"`
	ActiveConnections   int64         `json:"active_connections"`
	
	// Workflow metrics
	ActiveWorkflows     int64         `json:"active_workflows"`
	WorkflowThroughput  float64       `json:"workflow_throughput"`
	AvgWorkflowTime     time.Duration `json:"avg_workflow_time"`
	WorkflowErrorRate   float64       `json:"workflow_error_rate"`
	
	// Agent metrics
	AgentResponseTime   time.Duration `json:"agent_response_time"`
	AgentErrorRate      float64       `json:"agent_error_rate"`
	AgentUtilization    float64       `json:"agent_utilization"`
	
	// Cache metrics
	CacheHitRate        float64       `json:"cache_hit_rate"`
	CacheMissRate       float64       `json:"cache_miss_rate"`
	CacheLatency        time.Duration `json:"cache_latency"`
	
	// Database metrics
	DBConnectionCount   int           `json:"db_connection_count"`
	DBQueryTime         time.Duration `json:"db_query_time"`
	DBErrorRate         float64       `json:"db_error_rate"`
	
	// Network metrics
	NetworkBytesIn      uint64        `json:"network_bytes_in"`
	NetworkBytesOut     uint64        `json:"network_bytes_out"`
	NetworkLatency      time.Duration `json:"network_latency"`
	
	// Timestamps
	LastUpdated         time.Time     `json:"last_updated"`
	CollectionDuration  time.Duration `json:"collection_duration"`
}

// MetricCollector interface for collecting specific metrics
type MetricCollector interface {
	Collect(ctx context.Context) (map[string]interface{}, error)
	Name() string
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(logger Logger, interval time.Duration) *PerformanceMonitor {
	pm := &PerformanceMonitor{
		metrics:    &PerformanceMetrics{},
		collectors: make([]MetricCollector, 0),
		logger:     logger,
		stopCh:     make(chan struct{}),
		interval:   interval,
	}

	// Add default collectors
	pm.AddCollector(&SystemMetricsCollector{logger: logger})
	pm.AddCollector(&RuntimeMetricsCollector{logger: logger})
	pm.AddCollector(&GCMetricsCollector{logger: logger})

	return pm
}

// AddCollector adds a metric collector
func (pm *PerformanceMonitor) AddCollector(collector MetricCollector) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.collectors = append(pm.collectors, collector)
	pm.logger.Info("Metric collector added", "name", collector.Name())
}

// Start starts the performance monitoring
func (pm *PerformanceMonitor) Start(ctx context.Context) {
	pm.logger.Info("Starting performance monitor", "interval", pm.interval)
	
	ticker := time.NewTicker(pm.interval)
	defer ticker.Stop()

	// Collect initial metrics
	pm.collectMetrics(ctx)

	for {
		select {
		case <-ctx.Done():
			pm.logger.Info("Performance monitor stopped due to context cancellation")
			return
		case <-pm.stopCh:
			pm.logger.Info("Performance monitor stopped")
			return
		case <-ticker.C:
			pm.collectMetrics(ctx)
		}
	}
}

// Stop stops the performance monitoring
func (pm *PerformanceMonitor) Stop() {
	close(pm.stopCh)
}

// GetMetrics returns current performance metrics
func (pm *PerformanceMonitor) GetMetrics() *PerformanceMetrics {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	// Return a copy
	metricsCopy := *pm.metrics
	return &metricsCopy
}

// collectMetrics collects metrics from all collectors
func (pm *PerformanceMonitor) collectMetrics(ctx context.Context) {
	start := time.Now()
	
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Collect from all collectors
	allMetrics := make(map[string]interface{})
	
	for _, collector := range pm.collectors {
		metrics, err := collector.Collect(ctx)
		if err != nil {
			pm.logger.Error("Failed to collect metrics", err, "collector", collector.Name())
			continue
		}
		
		// Merge metrics
		for key, value := range metrics {
			allMetrics[key] = value
		}
	}

	// Update performance metrics
	pm.updateMetrics(allMetrics)
	pm.metrics.LastUpdated = time.Now()
	pm.metrics.CollectionDuration = time.Since(start)

	// Log performance warnings
	pm.checkPerformanceThresholds()
}

// updateMetrics updates the performance metrics structure
func (pm *PerformanceMonitor) updateMetrics(metrics map[string]interface{}) {
	// System metrics
	if val, ok := metrics["cpu_usage"].(float64); ok {
		pm.metrics.CPUUsage = val
	}
	if val, ok := metrics["memory_usage"].(float64); ok {
		pm.metrics.MemoryUsage = val
	}
	if val, ok := metrics["memory_allocated"].(uint64); ok {
		pm.metrics.MemoryAllocated = val
	}
	if val, ok := metrics["memory_total"].(uint64); ok {
		pm.metrics.MemoryTotal = val
	}
	if val, ok := metrics["goroutine_count"].(int); ok {
		pm.metrics.GoroutineCount = val
	}
	if val, ok := metrics["gc_pause_time"].(time.Duration); ok {
		pm.metrics.GCPauseTime = val
	}
	if val, ok := metrics["gc_count"].(uint32); ok {
		pm.metrics.GCCount = val
	}

	// Application metrics
	if val, ok := metrics["request_count"].(int64); ok {
		pm.metrics.RequestCount = val
	}
	if val, ok := metrics["request_rate"].(float64); ok {
		pm.metrics.RequestRate = val
	}
	if val, ok := metrics["average_response_time"].(time.Duration); ok {
		pm.metrics.AverageResponseTime = val
	}
	if val, ok := metrics["error_rate"].(float64); ok {
		pm.metrics.ErrorRate = val
	}
	if val, ok := metrics["active_connections"].(int64); ok {
		pm.metrics.ActiveConnections = val
	}

	// Workflow metrics
	if val, ok := metrics["active_workflows"].(int64); ok {
		pm.metrics.ActiveWorkflows = val
	}
	if val, ok := metrics["workflow_throughput"].(float64); ok {
		pm.metrics.WorkflowThroughput = val
	}
	if val, ok := metrics["avg_workflow_time"].(time.Duration); ok {
		pm.metrics.AvgWorkflowTime = val
	}
	if val, ok := metrics["workflow_error_rate"].(float64); ok {
		pm.metrics.WorkflowErrorRate = val
	}

	// Agent metrics
	if val, ok := metrics["agent_response_time"].(time.Duration); ok {
		pm.metrics.AgentResponseTime = val
	}
	if val, ok := metrics["agent_error_rate"].(float64); ok {
		pm.metrics.AgentErrorRate = val
	}
	if val, ok := metrics["agent_utilization"].(float64); ok {
		pm.metrics.AgentUtilization = val
	}

	// Cache metrics
	if val, ok := metrics["cache_hit_rate"].(float64); ok {
		pm.metrics.CacheHitRate = val
	}
	if val, ok := metrics["cache_miss_rate"].(float64); ok {
		pm.metrics.CacheMissRate = val
	}
	if val, ok := metrics["cache_latency"].(time.Duration); ok {
		pm.metrics.CacheLatency = val
	}

	// Database metrics
	if val, ok := metrics["db_connection_count"].(int); ok {
		pm.metrics.DBConnectionCount = val
	}
	if val, ok := metrics["db_query_time"].(time.Duration); ok {
		pm.metrics.DBQueryTime = val
	}
	if val, ok := metrics["db_error_rate"].(float64); ok {
		pm.metrics.DBErrorRate = val
	}

	// Network metrics
	if val, ok := metrics["network_bytes_in"].(uint64); ok {
		pm.metrics.NetworkBytesIn = val
	}
	if val, ok := metrics["network_bytes_out"].(uint64); ok {
		pm.metrics.NetworkBytesOut = val
	}
	if val, ok := metrics["network_latency"].(time.Duration); ok {
		pm.metrics.NetworkLatency = val
	}
}

// checkPerformanceThresholds checks for performance issues
func (pm *PerformanceMonitor) checkPerformanceThresholds() {
	// CPU usage warning
	if pm.metrics.CPUUsage > 80.0 {
		pm.logger.Warn("High CPU usage detected", "cpu_usage", pm.metrics.CPUUsage)
	}

	// Memory usage warning
	if pm.metrics.MemoryUsage > 85.0 {
		pm.logger.Warn("High memory usage detected", "memory_usage", pm.metrics.MemoryUsage)
	}

	// Goroutine count warning
	if pm.metrics.GoroutineCount > 10000 {
		pm.logger.Warn("High goroutine count detected", "goroutine_count", pm.metrics.GoroutineCount)
	}

	// Response time warning
	if pm.metrics.AverageResponseTime > 5*time.Second {
		pm.logger.Warn("High response time detected", "response_time", pm.metrics.AverageResponseTime)
	}

	// Error rate warning
	if pm.metrics.ErrorRate > 5.0 {
		pm.logger.Warn("High error rate detected", "error_rate", pm.metrics.ErrorRate)
	}

	// Cache hit rate warning
	if pm.metrics.CacheHitRate < 80.0 && pm.metrics.CacheHitRate > 0 {
		pm.logger.Warn("Low cache hit rate detected", "hit_rate", pm.metrics.CacheHitRate)
	}

	// Database query time warning
	if pm.metrics.DBQueryTime > 1*time.Second {
		pm.logger.Warn("High database query time detected", "query_time", pm.metrics.DBQueryTime)
	}
}

// SystemMetricsCollector collects system-level metrics
type SystemMetricsCollector struct {
	logger Logger
}

func (smc *SystemMetricsCollector) Name() string {
	return "system_metrics"
}

func (smc *SystemMetricsCollector) Collect(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Get memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics["memory_allocated"] = m.Alloc
	metrics["memory_total"] = m.TotalAlloc
	metrics["memory_usage"] = float64(m.Alloc) / float64(m.Sys) * 100

	return metrics, nil
}

// RuntimeMetricsCollector collects Go runtime metrics
type RuntimeMetricsCollector struct {
	logger Logger
}

func (rmc *RuntimeMetricsCollector) Name() string {
	return "runtime_metrics"
}

func (rmc *RuntimeMetricsCollector) Collect(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Goroutine count
	metrics["goroutine_count"] = runtime.NumGoroutine()

	// CPU count
	metrics["cpu_count"] = runtime.NumCPU()

	return metrics, nil
}

// GCMetricsCollector collects garbage collection metrics
type GCMetricsCollector struct {
	logger    Logger
	lastGCCount uint32
}

func (gmc *GCMetricsCollector) Name() string {
	return "gc_metrics"
}

func (gmc *GCMetricsCollector) Collect(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// GC metrics
	metrics["gc_count"] = m.NumGC
	
	// Calculate GC pause time
	if len(m.PauseNs) > 0 {
		// Get the most recent pause time
		recentPause := m.PauseNs[(m.NumGC+255)%256]
		metrics["gc_pause_time"] = time.Duration(recentPause)
	}

	// GC frequency (collections since last check)
	if gmc.lastGCCount > 0 {
		gcsSinceLastCheck := m.NumGC - gmc.lastGCCount
		metrics["gc_frequency"] = gcsSinceLastCheck
	}
	gmc.lastGCCount = m.NumGC

	return metrics, nil
}

// PerformanceAlert represents a performance alert
type PerformanceAlert struct {
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Message     string                 `json:"message"`
	Metrics     map[string]interface{} `json:"metrics"`
	Timestamp   time.Time              `json:"timestamp"`
	Threshold   interface{}            `json:"threshold"`
	CurrentValue interface{}           `json:"current_value"`
}

// AlertManager manages performance alerts
type AlertManager struct {
	alerts    []PerformanceAlert
	thresholds map[string]interface{}
	logger    Logger
	mutex     sync.RWMutex
}

// NewAlertManager creates a new alert manager
func NewAlertManager(logger Logger) *AlertManager {
	return &AlertManager{
		alerts: make([]PerformanceAlert, 0),
		thresholds: map[string]interface{}{
			"cpu_usage":           80.0,
			"memory_usage":        85.0,
			"goroutine_count":     10000,
			"response_time":       5 * time.Second,
			"error_rate":          5.0,
			"cache_hit_rate_min":  80.0,
			"db_query_time":       1 * time.Second,
		},
		logger: logger,
	}
}

// CheckMetrics checks metrics against thresholds and generates alerts
func (am *AlertManager) CheckMetrics(metrics *PerformanceMetrics) []PerformanceAlert {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	var newAlerts []PerformanceAlert

	// Check CPU usage
	if metrics.CPUUsage > am.thresholds["cpu_usage"].(float64) {
		alert := PerformanceAlert{
			Type:         "cpu_usage",
			Severity:     "warning",
			Message:      "High CPU usage detected",
			Timestamp:    time.Now(),
			Threshold:    am.thresholds["cpu_usage"],
			CurrentValue: metrics.CPUUsage,
		}
		newAlerts = append(newAlerts, alert)
	}

	// Check memory usage
	if metrics.MemoryUsage > am.thresholds["memory_usage"].(float64) {
		alert := PerformanceAlert{
			Type:         "memory_usage",
			Severity:     "warning",
			Message:      "High memory usage detected",
			Timestamp:    time.Now(),
			Threshold:    am.thresholds["memory_usage"],
			CurrentValue: metrics.MemoryUsage,
		}
		newAlerts = append(newAlerts, alert)
	}

	// Check goroutine count
	if metrics.GoroutineCount > am.thresholds["goroutine_count"].(int) {
		alert := PerformanceAlert{
			Type:         "goroutine_count",
			Severity:     "warning",
			Message:      "High goroutine count detected",
			Timestamp:    time.Now(),
			Threshold:    am.thresholds["goroutine_count"],
			CurrentValue: metrics.GoroutineCount,
		}
		newAlerts = append(newAlerts, alert)
	}

	// Check response time
	if metrics.AverageResponseTime > am.thresholds["response_time"].(time.Duration) {
		alert := PerformanceAlert{
			Type:         "response_time",
			Severity:     "warning",
			Message:      "High response time detected",
			Timestamp:    time.Now(),
			Threshold:    am.thresholds["response_time"],
			CurrentValue: metrics.AverageResponseTime,
		}
		newAlerts = append(newAlerts, alert)
	}

	// Add new alerts to the list
	am.alerts = append(am.alerts, newAlerts...)

	// Keep only recent alerts (last 100)
	if len(am.alerts) > 100 {
		am.alerts = am.alerts[len(am.alerts)-100:]
	}

	return newAlerts
}

// GetRecentAlerts returns recent alerts
func (am *AlertManager) GetRecentAlerts(limit int) []PerformanceAlert {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	if limit <= 0 || limit > len(am.alerts) {
		limit = len(am.alerts)
	}

	if limit == 0 {
		return []PerformanceAlert{}
	}

	// Return the most recent alerts
	start := len(am.alerts) - limit
	alerts := make([]PerformanceAlert, limit)
	copy(alerts, am.alerts[start:])
	
	return alerts
}
