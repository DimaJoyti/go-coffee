package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics represents the metrics for the service
type Metrics struct {
	registry             *prometheus.Registry
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	databaseQueryTotal   *prometheus.CounterVec
	databaseQueryDuration *prometheus.HistogramVec
	kafkaMessagesTotal   *prometheus.CounterVec
	kafkaMessageDuration *prometheus.HistogramVec
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	registry := prometheus.NewRegistry()

	// HTTP metrics
	httpRequestsTotal := promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration := promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Database metrics
	databaseQueryTotal := promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table", "status"},
	)

	databaseQueryDuration := promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// Kafka metrics
	kafkaMessagesTotal := promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_total",
			Help: "Total number of Kafka messages",
		},
		[]string{"topic", "event_type", "status"},
	)

	kafkaMessageDuration := promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kafka_message_duration_seconds",
			Help:    "Duration of Kafka message processing in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"topic", "event_type"},
	)

	return &Metrics{
		registry:             registry,
		httpRequestsTotal:    httpRequestsTotal,
		httpRequestDuration:  httpRequestDuration,
		databaseQueryTotal:   databaseQueryTotal,
		databaseQueryDuration: databaseQueryDuration,
		kafkaMessagesTotal:   kafkaMessagesTotal,
		kafkaMessageDuration: kafkaMessageDuration,
	}
}

// Handler returns an HTTP handler for the metrics endpoint
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// ObserveHTTPRequest observes an HTTP request
func (m *Metrics) ObserveHTTPRequest(method, path string, status int, duration time.Duration) {
	m.httpRequestsTotal.WithLabelValues(method, path, http.StatusText(status)).Inc()
	m.httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// ObserveDatabaseQuery observes a database query
func (m *Metrics) ObserveDatabaseQuery(operation, table string, err error, duration time.Duration) {
	status := "success"
	if err != nil {
		status = "error"
	}
	m.databaseQueryTotal.WithLabelValues(operation, table, status).Inc()
	m.databaseQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// ObserveKafkaMessage observes a Kafka message
func (m *Metrics) ObserveKafkaMessage(topic string, eventType string, err error, duration time.Duration) {
	status := "success"
	if err != nil {
		status = "error"
	}
	m.kafkaMessagesTotal.WithLabelValues(topic, eventType, status).Inc()
	m.kafkaMessageDuration.WithLabelValues(topic, eventType).Observe(duration.Seconds())
}

// HTTPMiddleware returns a middleware that observes HTTP requests
func (m *Metrics) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a response writer that captures the status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Call the next handler
		next.ServeHTTP(rw, r)
		
		// Observe the request
		duration := time.Since(start)
		m.ObserveHTTPRequest(r.Method, r.URL.Path, rw.statusCode, duration)
	})
}

// responseWriter is a wrapper around http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
