package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// PrometheusMetrics provides Prometheus metrics collection
type PrometheusMetrics struct {
	registry  *prometheus.Registry
	container infrastructure.ContainerInterface
	logger    *logger.Logger
	config    *PrometheusConfig

	// HTTP metrics
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	httpRequestSize     *prometheus.HistogramVec
	httpResponseSize    *prometheus.HistogramVec

	// Infrastructure metrics
	databaseConnections *prometheus.GaugeVec
	redisConnections    *prometheus.GaugeVec
	cacheHitRatio       *prometheus.GaugeVec
	sessionCount        *prometheus.GaugeVec

	// System metrics
	goroutineCount prometheus.Gauge
	memoryUsage    prometheus.Gauge
	gcDuration     prometheus.Histogram

	// Business metrics
	ordersTotal   *prometheus.CounterVec
	orderDuration *prometheus.HistogramVec
	userSessions  *prometheus.GaugeVec
	errorRate     *prometheus.CounterVec

	// Health metrics
	healthCheckStatus   *prometheus.GaugeVec
	healthCheckDuration *prometheus.HistogramVec
}

// PrometheusConfig defines Prometheus configuration
type PrometheusConfig struct {
	Enabled     bool          `yaml:"enabled"`
	Namespace   string        `yaml:"namespace"`
	Subsystem   string        `yaml:"subsystem"`
	MetricsPath string        `yaml:"metrics_path"`
	Port        int           `yaml:"port"`
	Interval    time.Duration `yaml:"interval"`
}

// DefaultPrometheusConfig returns default Prometheus configuration
func DefaultPrometheusConfig() *PrometheusConfig {
	return &PrometheusConfig{
		Enabled:     true,
		Namespace:   "go_coffee",
		Subsystem:   "infrastructure",
		MetricsPath: "/metrics",
		Port:        9090,
		Interval:    15 * time.Second,
	}
}

// NewPrometheusMetrics creates a new Prometheus metrics collector
func NewPrometheusMetrics(container infrastructure.ContainerInterface, config *PrometheusConfig, logger *logger.Logger) *PrometheusMetrics {
	if config == nil {
		config = DefaultPrometheusConfig()
	}

	registry := prometheus.NewRegistry()

	pm := &PrometheusMetrics{
		registry:  registry,
		container: container,
		logger:    logger,
		config:    config,
	}

	pm.initializeMetrics()
	pm.registerMetrics()

	return pm
}

// initializeMetrics initializes all Prometheus metrics
func (pm *PrometheusMetrics) initializeMetrics() {
	// HTTP metrics
	pm.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: pm.config.Namespace,
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	pm.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: pm.config.Namespace,
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	pm.httpRequestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: pm.config.Namespace,
			Subsystem: "http",
			Name:      "request_size_bytes",
			Help:      "HTTP request size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	pm.httpResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: pm.config.Namespace,
			Subsystem: "http",
			Name:      "response_size_bytes",
			Help:      "HTTP response size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	// Infrastructure metrics
	pm.databaseConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: pm.config.Namespace,
			Subsystem: "database",
			Name:      "connections",
			Help:      "Number of database connections",
		},
		[]string{"state"},
	)

	pm.redisConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: pm.config.Namespace,
			Subsystem: "redis",
			Name:      "connections",
			Help:      "Number of Redis connections",
		},
		[]string{"state"},
	)

	pm.cacheHitRatio = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: pm.config.Namespace,
			Subsystem: "cache",
			Name:      "hit_ratio",
			Help:      "Cache hit ratio",
		},
		[]string{"cache_type"},
	)

	pm.sessionCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: pm.config.Namespace,
			Subsystem: "session",
			Name:      "active_sessions",
			Help:      "Number of active sessions",
		},
		[]string{"status"},
	)

	// System metrics
	pm.goroutineCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: pm.config.Namespace,
			Subsystem: "system",
			Name:      "goroutines",
			Help:      "Number of goroutines",
		},
	)

	pm.memoryUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: pm.config.Namespace,
			Subsystem: "system",
			Name:      "memory_usage_bytes",
			Help:      "Memory usage in bytes",
		},
	)

	pm.gcDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: pm.config.Namespace,
			Subsystem: "system",
			Name:      "gc_duration_seconds",
			Help:      "Garbage collection duration in seconds",
			Buckets:   prometheus.ExponentialBuckets(0.001, 2, 15),
		},
	)

	// Business metrics
	pm.ordersTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: pm.config.Namespace,
			Subsystem: "business",
			Name:      "orders_total",
			Help:      "Total number of orders",
		},
		[]string{"status", "type"},
	)

	pm.orderDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: pm.config.Namespace,
			Subsystem: "business",
			Name:      "order_duration_seconds",
			Help:      "Order processing duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"type"},
	)

	pm.userSessions = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: pm.config.Namespace,
			Subsystem: "business",
			Name:      "user_sessions",
			Help:      "Number of user sessions",
		},
		[]string{"role"},
	)

	pm.errorRate = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: pm.config.Namespace,
			Subsystem: "business",
			Name:      "errors_total",
			Help:      "Total number of errors",
		},
		[]string{"type", "component"},
	)

	// Health metrics
	pm.healthCheckStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: pm.config.Namespace,
			Subsystem: "health",
			Name:      "check_status",
			Help:      "Health check status (1=healthy, 0=unhealthy)",
		},
		[]string{"check_name"},
	)

	pm.healthCheckDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: pm.config.Namespace,
			Subsystem: "health",
			Name:      "check_duration_seconds",
			Help:      "Health check duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"check_name"},
	)
}

// registerMetrics registers all metrics with the registry
func (pm *PrometheusMetrics) registerMetrics() {
	// HTTP metrics
	pm.registry.MustRegister(pm.httpRequestsTotal)
	pm.registry.MustRegister(pm.httpRequestDuration)
	pm.registry.MustRegister(pm.httpRequestSize)
	pm.registry.MustRegister(pm.httpResponseSize)

	// Infrastructure metrics
	pm.registry.MustRegister(pm.databaseConnections)
	pm.registry.MustRegister(pm.redisConnections)
	pm.registry.MustRegister(pm.cacheHitRatio)
	pm.registry.MustRegister(pm.sessionCount)

	// System metrics
	pm.registry.MustRegister(pm.goroutineCount)
	pm.registry.MustRegister(pm.memoryUsage)
	pm.registry.MustRegister(pm.gcDuration)

	// Business metrics
	pm.registry.MustRegister(pm.ordersTotal)
	pm.registry.MustRegister(pm.orderDuration)
	pm.registry.MustRegister(pm.userSessions)
	pm.registry.MustRegister(pm.errorRate)

	// Health metrics
	pm.registry.MustRegister(pm.healthCheckStatus)
	pm.registry.MustRegister(pm.healthCheckDuration)
}

// HTTP Metrics Recording

// RecordHTTPRequest records HTTP request metrics
func (pm *PrometheusMetrics) RecordHTTPRequest(method, path string, statusCode int, duration time.Duration, requestSize, responseSize int64) {
	status := strconv.Itoa(statusCode)

	pm.httpRequestsTotal.WithLabelValues(method, path, status).Inc()
	pm.httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
	pm.httpRequestSize.WithLabelValues(method, path).Observe(float64(requestSize))
	pm.httpResponseSize.WithLabelValues(method, path).Observe(float64(responseSize))
}

// Infrastructure Metrics Recording

// RecordDatabaseConnections records database connection metrics
func (pm *PrometheusMetrics) RecordDatabaseConnections(active, idle int) {
	pm.databaseConnections.WithLabelValues("active").Set(float64(active))
	pm.databaseConnections.WithLabelValues("idle").Set(float64(idle))
}

// RecordRedisConnections records Redis connection metrics
func (pm *PrometheusMetrics) RecordRedisConnections(active, idle int) {
	pm.redisConnections.WithLabelValues("active").Set(float64(active))
	pm.redisConnections.WithLabelValues("idle").Set(float64(idle))
}

// RecordCacheHitRatio records cache hit ratio
func (pm *PrometheusMetrics) RecordCacheHitRatio(cacheType string, ratio float64) {
	pm.cacheHitRatio.WithLabelValues(cacheType).Set(ratio)
}

// RecordSessionCount records session count
func (pm *PrometheusMetrics) RecordSessionCount(status string, count int) {
	pm.sessionCount.WithLabelValues(status).Set(float64(count))
}

// System Metrics Recording

// RecordSystemMetrics records system-level metrics
func (pm *PrometheusMetrics) RecordSystemMetrics() {
	pm.goroutineCount.Set(float64(getGoroutineCount()))
	pm.memoryUsage.Set(float64(getMemoryAlloc()))
}

// RecordGCDuration records garbage collection duration
func (pm *PrometheusMetrics) RecordGCDuration(duration time.Duration) {
	pm.gcDuration.Observe(duration.Seconds())
}

// Business Metrics Recording

// RecordOrder records order metrics
func (pm *PrometheusMetrics) RecordOrder(status, orderType string, duration time.Duration) {
	pm.ordersTotal.WithLabelValues(status, orderType).Inc()
	if duration > 0 {
		pm.orderDuration.WithLabelValues(orderType).Observe(duration.Seconds())
	}
}

// RecordUserSession records user session metrics
func (pm *PrometheusMetrics) RecordUserSession(role string, count int) {
	pm.userSessions.WithLabelValues(role).Set(float64(count))
}

// RecordError records error metrics
func (pm *PrometheusMetrics) RecordError(errorType, component string) {
	pm.errorRate.WithLabelValues(errorType, component).Inc()
}

// Health Metrics Recording

// RecordHealthCheck records health check metrics
func (pm *PrometheusMetrics) RecordHealthCheck(checkName string, healthy bool, duration time.Duration) {
	status := 0.0
	if healthy {
		status = 1.0
	}
	pm.healthCheckStatus.WithLabelValues(checkName).Set(status)
	pm.healthCheckDuration.WithLabelValues(checkName).Observe(duration.Seconds())
}

// Utility Methods

// GetRegistry returns the Prometheus registry
func (pm *PrometheusMetrics) GetRegistry() *prometheus.Registry {
	return pm.registry
}

// GetHandler returns the Prometheus HTTP handler
func (pm *PrometheusMetrics) GetHandler() http.Handler {
	return promhttp.HandlerFor(pm.registry, promhttp.HandlerOpts{})
}

// StartMetricsServer starts the Prometheus metrics server
func (pm *PrometheusMetrics) StartMetricsServer() error {
	if !pm.config.Enabled {
		pm.logger.Info("Prometheus metrics disabled")
		return nil
	}

	mux := http.NewServeMux()
	mux.Handle(pm.config.MetricsPath, pm.GetHandler())

	// Add health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	addr := fmt.Sprintf(":%d", pm.config.Port)
	pm.logger.WithField("address", addr).Info("Starting Prometheus metrics server")

	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			pm.logger.WithError(err).Error("Prometheus metrics server failed")
		}
	}()

	return nil
}

// CollectInfrastructureMetrics collects metrics from infrastructure components
func (pm *PrometheusMetrics) CollectInfrastructureMetrics(ctx context.Context) {
	// Collect system metrics
	pm.RecordSystemMetrics()

	// Collect database metrics if available
	if db := pm.container.GetDatabase(); db != nil {
		// In a real implementation, you would get actual connection stats
		pm.RecordDatabaseConnections(10, 5) // Mock values
	}

	// Collect Redis metrics if available
	if redis := pm.container.GetRedis(); redis != nil {
		// In a real implementation, you would get actual connection stats
		pm.RecordRedisConnections(5, 2) // Mock values
	}

	// Collect cache metrics if available
	if cache := pm.container.GetCache(); cache != nil {
		// In a real implementation, you would get actual hit ratio
		pm.RecordCacheHitRatio("redis", 0.85) // Mock value
	}

	// Collect session metrics if available
	if sessionManager := pm.container.GetSessionManager(); sessionManager != nil {
		// In a real implementation, you would get actual session counts
		pm.RecordSessionCount("active", 25) // Mock value
	}
}

// StartPeriodicCollection starts periodic metrics collection
func (pm *PrometheusMetrics) StartPeriodicCollection(ctx context.Context) {
	if !pm.config.Enabled {
		return
	}

	ticker := time.NewTicker(pm.config.Interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pm.CollectInfrastructureMetrics(ctx)
			}
		}
	}()

	pm.logger.WithField("interval", pm.config.Interval).Info("Started periodic metrics collection")
}
