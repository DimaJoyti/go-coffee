package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/cache"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// MetricsCollector collects and stores application metrics
type MetricsCollector struct {
	container infrastructure.ContainerInterface
	cache     cache.Cache
	logger    *logger.Logger
	config    *MetricsConfig
	metrics   map[string]*Metric
	mutex     sync.RWMutex
	stopChan  chan struct{}
	running   bool
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Enabled         bool          `yaml:"enabled"`
	CollectInterval time.Duration `yaml:"collect_interval"`
	RetentionPeriod time.Duration `yaml:"retention_period"`
	Namespace       string        `yaml:"namespace"`
	ServiceName     string        `yaml:"service_name"`
}

// Metric represents a single metric
type Metric struct {
	Name        string                 `json:"name"`
	Type        MetricType             `json:"type"`
	Value       float64                `json:"value"`
	Labels      map[string]string      `json:"labels"`
	Timestamp   time.Time              `json:"timestamp"`
	Description string                 `json:"description"`
	Unit        string                 `json:"unit"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// SystemMetrics represents system-level metrics
type SystemMetrics struct {
	CPUUsage         float64 `json:"cpu_usage"`
	MemoryUsage      float64 `json:"memory_usage"`
	MemoryTotal      uint64  `json:"memory_total"`
	MemoryUsed       uint64  `json:"memory_used"`
	GoroutineCount   int     `json:"goroutine_count"`
	GCPauseTotal     float64 `json:"gc_pause_total"`
	GCCount          uint32  `json:"gc_count"`
	HeapSize         uint64  `json:"heap_size"`
	HeapInUse        uint64  `json:"heap_in_use"`
	StackInUse       uint64  `json:"stack_in_use"`
	NextGC           uint64  `json:"next_gc"`
}

// ApplicationMetrics represents application-level metrics
type ApplicationMetrics struct {
	RequestCount       int64             `json:"request_count"`
	RequestDuration    float64           `json:"request_duration_avg"`
	ErrorCount         int64             `json:"error_count"`
	ErrorRate          float64           `json:"error_rate"`
	ActiveSessions     int64             `json:"active_sessions"`
	DatabaseConnections int              `json:"database_connections"`
	CacheHitRate       float64           `json:"cache_hit_rate"`
	EventsPublished    int64             `json:"events_published"`
	EventsProcessed    int64             `json:"events_processed"`
	CustomMetrics      map[string]float64 `json:"custom_metrics"`
}

// MetricsSnapshot represents a complete metrics snapshot
type MetricsSnapshot struct {
	Timestamp   time.Time           `json:"timestamp"`
	ServiceName string              `json:"service_name"`
	Version     string              `json:"version"`
	System      SystemMetrics       `json:"system"`
	Application ApplicationMetrics  `json:"application"`
	Health      map[string]string   `json:"health"`
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(container infrastructure.ContainerInterface, config *MetricsConfig, logger *logger.Logger) *MetricsCollector {
	if config == nil {
		config = &MetricsConfig{
			Enabled:         true,
			CollectInterval: 30 * time.Second,
			RetentionPeriod: 24 * time.Hour,
			Namespace:       "go_coffee",
			ServiceName:     "unknown",
		}
	}

	return &MetricsCollector{
		container: container,
		cache:     container.GetCache(),
		logger:    logger,
		config:    config,
		metrics:   make(map[string]*Metric),
		stopChan:  make(chan struct{}),
	}
}

// Start starts the metrics collection
func (mc *MetricsCollector) Start(ctx context.Context) error {
	if !mc.config.Enabled {
		mc.logger.Info("Metrics collection is disabled")
		return nil
	}

	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if mc.running {
		return fmt.Errorf("metrics collector is already running")
	}

	mc.running = true

	// Start collection goroutine
	go mc.collectLoop()

	mc.logger.InfoWithFields("Metrics collector started",
		logger.Duration("interval", mc.config.CollectInterval),
		logger.String("namespace", mc.config.Namespace))

	return nil
}

// Stop stops the metrics collection
func (mc *MetricsCollector) Stop(ctx context.Context) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if !mc.running {
		return nil
	}

	mc.running = false
	close(mc.stopChan)

	mc.logger.Info("Metrics collector stopped")
	return nil
}

// RecordMetric records a custom metric
func (mc *MetricsCollector) RecordMetric(name string, metricType MetricType, value float64, labels map[string]string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	metric := &Metric{
		Name:      name,
		Type:      metricType,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
	}

	mc.metrics[name] = metric

	// Store in cache for persistence
	if mc.cache != nil {
		key := fmt.Sprintf("metrics:%s:%s", mc.config.Namespace, name)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			mc.cache.Set(ctx, key, metric, mc.config.RetentionPeriod)
		}()
	}
}

// IncrementCounter increments a counter metric
func (mc *MetricsCollector) IncrementCounter(name string, labels map[string]string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if existing, exists := mc.metrics[name]; exists && existing.Type == MetricTypeCounter {
		existing.Value++
		existing.Timestamp = time.Now()
	} else {
		mc.metrics[name] = &Metric{
			Name:      name,
			Type:      MetricTypeCounter,
			Value:     1,
			Labels:    labels,
			Timestamp: time.Now(),
		}
	}
}

// SetGauge sets a gauge metric value
func (mc *MetricsCollector) SetGauge(name string, value float64, labels map[string]string) {
	mc.RecordMetric(name, MetricTypeGauge, value, labels)
}

// GetMetrics returns all current metrics
func (mc *MetricsCollector) GetMetrics() map[string]*Metric {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	result := make(map[string]*Metric)
	for k, v := range mc.metrics {
		result[k] = v
	}
	return result
}

// GetMetricsSnapshot returns a complete metrics snapshot
func (mc *MetricsCollector) GetMetricsSnapshot() *MetricsSnapshot {
	snapshot := &MetricsSnapshot{
		Timestamp:   time.Now(),
		ServiceName: mc.config.ServiceName,
		Version:     "1.0.0", // This could be injected
		System:      mc.collectSystemMetrics(),
		Application: mc.collectApplicationMetrics(),
		Health:      mc.collectHealthMetrics(),
	}

	return snapshot
}

// collectLoop runs the metrics collection loop
func (mc *MetricsCollector) collectLoop() {
	ticker := time.NewTicker(mc.config.CollectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mc.collectMetrics()
		case <-mc.stopChan:
			return
		}
	}
}

// collectMetrics collects all metrics
func (mc *MetricsCollector) collectMetrics() {
	// Collect system metrics
	systemMetrics := mc.collectSystemMetrics()
	mc.SetGauge("system_cpu_usage", systemMetrics.CPUUsage, nil)
	mc.SetGauge("system_memory_usage", systemMetrics.MemoryUsage, nil)
	mc.SetGauge("system_goroutines", float64(systemMetrics.GoroutineCount), nil)
	mc.SetGauge("system_heap_size", float64(systemMetrics.HeapSize), nil)

	// Collect application metrics
	appMetrics := mc.collectApplicationMetrics()
	mc.SetGauge("app_active_sessions", float64(appMetrics.ActiveSessions), nil)
	mc.SetGauge("app_cache_hit_rate", appMetrics.CacheHitRate, nil)
	mc.SetGauge("app_database_connections", float64(appMetrics.DatabaseConnections), nil)

	// Store snapshot in cache
	if mc.cache != nil {
		snapshot := mc.GetMetricsSnapshot()
		key := fmt.Sprintf("metrics_snapshot:%s:%d", mc.config.ServiceName, time.Now().Unix())
		
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			mc.cache.Set(ctx, key, snapshot, mc.config.RetentionPeriod)
		}()
	}
}

// collectSystemMetrics collects system-level metrics
func (mc *MetricsCollector) collectSystemMetrics() SystemMetrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return SystemMetrics{
		CPUUsage:       0, // Would need additional library for CPU usage
		MemoryUsage:    float64(memStats.Alloc) / float64(memStats.Sys) * 100,
		MemoryTotal:    memStats.Sys,
		MemoryUsed:     memStats.Alloc,
		GoroutineCount: runtime.NumGoroutine(),
		GCPauseTotal:   float64(memStats.PauseTotalNs) / 1e9,
		GCCount:        memStats.NumGC,
		HeapSize:       memStats.HeapSys,
		HeapInUse:      memStats.HeapInuse,
		StackInUse:     memStats.StackInuse,
		NextGC:         memStats.NextGC,
	}
}

// collectApplicationMetrics collects application-level metrics
func (mc *MetricsCollector) collectApplicationMetrics() ApplicationMetrics {
	metrics := ApplicationMetrics{
		CustomMetrics: make(map[string]float64),
	}

	// Get database connection stats
	if db := mc.container.GetDatabase(); db != nil {
		stats := db.Stats()
		metrics.DatabaseConnections = stats.OpenConnections
	}

	// Get cache stats
	if cache := mc.container.GetCache(); cache != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if cacheStats, err := cache.Stats(ctx); err == nil {
			if cacheStats.Hits+cacheStats.Misses > 0 {
				metrics.CacheHitRate = float64(cacheStats.Hits) / float64(cacheStats.Hits+cacheStats.Misses) * 100
			}
		}
	}

	// Get session count from cache
	if mc.cache != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		// Count active sessions (this is a simplified approach)
		if keys, err := mc.cache.Keys(ctx, "session:*"); err == nil {
			metrics.ActiveSessions = int64(len(keys))
		}
	}

	return metrics
}

// collectHealthMetrics collects health status metrics
func (mc *MetricsCollector) collectHealthMetrics() map[string]string {
	health := make(map[string]string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check infrastructure health
	if healthStatus, err := mc.container.HealthCheck(ctx); err == nil {
		health["overall"] = healthStatus.Overall
		
		// Add individual component health
		for name, err := range healthStatus.Database {
			if err == nil {
				health["database_"+name] = "healthy"
			} else {
				health["database_"+name] = "unhealthy"
			}
		}

		if healthStatus.Redis == nil {
			health["redis"] = "healthy"
		} else {
			health["redis"] = "unhealthy"
		}

		if healthStatus.Cache == nil {
			health["cache"] = "healthy"
		} else {
			health["cache"] = "unhealthy"
		}
	} else {
		health["overall"] = "unhealthy"
	}

	return health
}

// ExportPrometheusMetrics exports metrics in Prometheus format
func (mc *MetricsCollector) ExportPrometheusMetrics() string {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	var output string
	
	for name, metric := range mc.metrics {
		// Format metric name for Prometheus
		prometheusName := fmt.Sprintf("%s_%s", mc.config.Namespace, name)
		
		// Add help text
		output += fmt.Sprintf("# HELP %s %s\n", prometheusName, metric.Description)
		output += fmt.Sprintf("# TYPE %s %s\n", prometheusName, string(metric.Type))
		
		// Add labels
		labels := ""
		if len(metric.Labels) > 0 {
			labelPairs := make([]string, 0, len(metric.Labels))
			for k, v := range metric.Labels {
				labelPairs = append(labelPairs, fmt.Sprintf(`%s="%s"`, k, v))
			}
			labels = fmt.Sprintf("{%s}", fmt.Sprintf("%v", labelPairs))
		}
		
		// Add metric value
		output += fmt.Sprintf("%s%s %f %d\n", prometheusName, labels, metric.Value, metric.Timestamp.Unix()*1000)
	}

	return output
}

// ExportJSONMetrics exports metrics in JSON format
func (mc *MetricsCollector) ExportJSONMetrics() ([]byte, error) {
	snapshot := mc.GetMetricsSnapshot()
	return json.MarshalIndent(snapshot, "", "  ")
}

// GetMetricHistory returns historical metrics from cache
func (mc *MetricsCollector) GetMetricHistory(ctx context.Context, metricName string, duration time.Duration) ([]*Metric, error) {
	if mc.cache == nil {
		return nil, fmt.Errorf("cache not available")
	}

	// This is a simplified implementation
	// In a real system, you'd store time-series data more efficiently
	pattern := fmt.Sprintf("metrics:%s:%s:*", mc.config.Namespace, metricName)
	keys, err := mc.cache.Keys(ctx, pattern)
	if err != nil {
		return nil, err
	}

	metrics := make([]*Metric, 0, len(keys))
	for _, key := range keys {
		var metric Metric
		if err := mc.cache.Get(ctx, key, &metric); err == nil {
			// Filter by duration
			if time.Since(metric.Timestamp) <= duration {
				metrics = append(metrics, &metric)
			}
		}
	}

	return metrics, nil
}
