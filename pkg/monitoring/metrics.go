package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics interface defines monitoring operations
type Metrics interface {
	IncrementCounter(name string, labels map[string]string)
	RecordHistogram(name string, value float64, labels map[string]string)
	SetGauge(name string, value float64, labels map[string]string)
	RecordDuration(name string, start time.Time, labels map[string]string)
}

// PrometheusMetrics implements Metrics using Prometheus
type PrometheusMetrics struct {
	registry *prometheus.Registry
	counters map[string]*prometheus.CounterVec
	histograms map[string]*prometheus.HistogramVec
	gauges   map[string]*prometheus.GaugeVec
}

// NewPrometheusMetrics creates a new Prometheus metrics instance
func NewPrometheusMetrics() *PrometheusMetrics {
	registry := prometheus.NewRegistry()
	
	pm := &PrometheusMetrics{
		registry:   registry,
		counters:   make(map[string]*prometheus.CounterVec),
		histograms: make(map[string]*prometheus.HistogramVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
	}

	// Register default metrics
	pm.registerDefaultMetrics()

	return pm
}

// registerDefaultMetrics registers common application metrics
func (pm *PrometheusMetrics) registerDefaultMetrics() {
	// HTTP request metrics
	pm.counters["http_requests_total"] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code", "service"},
	)

	pm.histograms["http_request_duration_seconds"] = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "service"},
	)

	// Database metrics
	pm.counters["database_queries_total"] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table", "status"},
	)

	pm.histograms["database_query_duration_seconds"] = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// Cache metrics
	pm.counters["cache_operations_total"] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"operation", "result"},
	)

	// Business metrics
	pm.counters["orders_total"] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_total",
			Help: "Total number of orders",
		},
		[]string{"status", "payment_method"},
	)

	pm.gauges["active_orders"] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_orders",
			Help: "Number of active orders",
		},
		[]string{"status"},
	)

	pm.histograms["order_value_dollars"] = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_value_dollars",
			Help:    "Order value in dollars",
			Buckets: []float64{1, 5, 10, 20, 50, 100, 200},
		},
		[]string{"payment_method"},
	)

	// Bitcoin/Crypto metrics
	pm.counters["bitcoin_transactions_total"] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bitcoin_transactions_total",
			Help: "Total number of Bitcoin transactions",
		},
		[]string{"type", "network", "status"},
	)

	pm.histograms["bitcoin_transaction_fee_satoshis"] = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bitcoin_transaction_fee_satoshis",
			Help:    "Bitcoin transaction fee in satoshis",
			Buckets: []float64{100, 500, 1000, 5000, 10000, 50000},
		},
		[]string{"network"},
	)

	// AI metrics
	pm.counters["ai_predictions_total"] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_predictions_total",
			Help: "Total number of AI predictions",
		},
		[]string{"model", "type"},
	)

	pm.histograms["ai_prediction_confidence"] = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_prediction_confidence",
			Help:    "AI prediction confidence score",
			Buckets: []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
		},
		[]string{"model", "type"},
	)

	// Register all metrics
	for _, counter := range pm.counters {
		pm.registry.MustRegister(counter)
	}
	for _, histogram := range pm.histograms {
		pm.registry.MustRegister(histogram)
	}
	for _, gauge := range pm.gauges {
		pm.registry.MustRegister(gauge)
	}
}

// IncrementCounter increments a counter metric
func (pm *PrometheusMetrics) IncrementCounter(name string, labels map[string]string) {
	if counter, exists := pm.counters[name]; exists {
		counter.With(labels).Inc()
	}
}

// RecordHistogram records a value in a histogram metric
func (pm *PrometheusMetrics) RecordHistogram(name string, value float64, labels map[string]string) {
	if histogram, exists := pm.histograms[name]; exists {
		histogram.With(labels).Observe(value)
	}
}

// SetGauge sets a gauge metric value
func (pm *PrometheusMetrics) SetGauge(name string, value float64, labels map[string]string) {
	if gauge, exists := pm.gauges[name]; exists {
		gauge.With(labels).Set(value)
	}
}

// RecordDuration records the duration since start time
func (pm *PrometheusMetrics) RecordDuration(name string, start time.Time, labels map[string]string) {
	duration := time.Since(start).Seconds()
	pm.RecordHistogram(name, duration, labels)
}

// Handler returns the Prometheus metrics HTTP handler
func (pm *PrometheusMetrics) Handler() http.Handler {
	return promhttp.HandlerFor(pm.registry, promhttp.HandlerOpts{})
}

// MetricsMiddleware provides HTTP metrics middleware
func (pm *PrometheusMetrics) MetricsMiddleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

			next.ServeHTTP(wrapped, r)

			// Record metrics
			labels := map[string]string{
				"method":      r.Method,
				"endpoint":    r.URL.Path,
				"status_code": fmt.Sprintf("%d", wrapped.statusCode),
				"service":     serviceName,
			}

			pm.IncrementCounter("http_requests_total", labels)
			pm.RecordDuration("http_request_duration_seconds", start, map[string]string{
				"method":   r.Method,
				"endpoint": r.URL.Path,
				"service":  serviceName,
			})
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// BusinessMetrics provides business-specific metrics
type BusinessMetrics struct {
	metrics Metrics
}

// NewBusinessMetrics creates a new business metrics instance
func NewBusinessMetrics(metrics Metrics) *BusinessMetrics {
	return &BusinessMetrics{metrics: metrics}
}

// RecordOrder records order metrics
func (bm *BusinessMetrics) RecordOrder(status, paymentMethod string, value float64) {
	bm.metrics.IncrementCounter("orders_total", map[string]string{
		"status":         status,
		"payment_method": paymentMethod,
	})

	bm.metrics.RecordHistogram("order_value_dollars", value, map[string]string{
		"payment_method": paymentMethod,
	})
}

// UpdateActiveOrders updates active orders gauge
func (bm *BusinessMetrics) UpdateActiveOrders(status string, count float64) {
	bm.metrics.SetGauge("active_orders", count, map[string]string{
		"status": status,
	})
}

// RecordBitcoinTransaction records Bitcoin transaction metrics
func (bm *BusinessMetrics) RecordBitcoinTransaction(txType, network, status string, fee float64) {
	bm.metrics.IncrementCounter("bitcoin_transactions_total", map[string]string{
		"type":    txType,
		"network": network,
		"status":  status,
	})

	if fee > 0 {
		bm.metrics.RecordHistogram("bitcoin_transaction_fee_satoshis", fee, map[string]string{
			"network": network,
		})
	}
}

// RecordAIPrediction records AI prediction metrics
func (bm *BusinessMetrics) RecordAIPrediction(model, predictionType string, confidence float64) {
	bm.metrics.IncrementCounter("ai_predictions_total", map[string]string{
		"model": model,
		"type":  predictionType,
	})

	bm.metrics.RecordHistogram("ai_prediction_confidence", confidence, map[string]string{
		"model": model,
		"type":  predictionType,
	})
}

// HealthChecker provides health checking functionality
type HealthChecker struct {
	checks map[string]HealthCheck
}

// HealthCheck represents a health check function
type HealthCheck func(ctx context.Context) error

// HealthStatus represents the health status of a component
type HealthStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // healthy, unhealthy
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// OverallHealth represents the overall health of the system
type OverallHealth struct {
	Status string          `json:"status"`
	Checks []HealthStatus  `json:"checks"`
	Uptime string          `json:"uptime"`
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]HealthCheck),
	}
}

// AddCheck adds a health check
func (hc *HealthChecker) AddCheck(name string, check HealthCheck) {
	hc.checks[name] = check
}

// CheckHealth performs all health checks
func (hc *HealthChecker) CheckHealth(ctx context.Context) *OverallHealth {
	var checks []HealthStatus
	overallHealthy := true

	for name, check := range hc.checks {
		start := time.Now()
		err := check(ctx)
		latency := time.Since(start)

		status := HealthStatus{
			Name:    name,
			Latency: latency.String(),
		}

		if err != nil {
			status.Status = "unhealthy"
			status.Message = err.Error()
			overallHealthy = false
		} else {
			status.Status = "healthy"
		}

		checks = append(checks, status)
	}

	overallStatus := "healthy"
	if !overallHealthy {
		overallStatus = "unhealthy"
	}

	return &OverallHealth{
		Status: overallStatus,
		Checks: checks,
		Uptime: "running", // Would calculate actual uptime
	}
}

// Handler returns the health check HTTP handler
func (hc *HealthChecker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		health := hc.CheckHealth(ctx)

		w.Header().Set("Content-Type", "application/json")
		
		if health.Status == "healthy" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		// Would use json.NewEncoder(w).Encode(health) in real implementation
		fmt.Fprintf(w, `{"status":"%s","checks_count":%d}`, health.Status, len(health.Checks))
	}
}
