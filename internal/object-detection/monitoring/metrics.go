package monitoring

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricsInstance *Metrics
	metricsOnce     sync.Once
)

// Metrics holds all Prometheus metrics for the object detection service
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal     *prometheus.CounterVec
	HTTPRequestDuration   *prometheus.HistogramVec
	HTTPRequestsInFlight  prometheus.Gauge

	// Stream metrics
	ActiveStreams         prometheus.Gauge
	StreamsTotal          *prometheus.CounterVec
	StreamErrors          *prometheus.CounterVec
	StreamFramesProcessed *prometheus.CounterVec

	// Detection metrics
	DetectionRequestsTotal    *prometheus.CounterVec
	DetectionDuration         *prometheus.HistogramVec
	ObjectsDetectedTotal      *prometheus.CounterVec
	DetectionErrors           *prometheus.CounterVec
	ModelLoadDuration         prometheus.Histogram
	FrameProcessingDuration   *prometheus.HistogramVec

	// Tracking metrics
	ActiveTrackedObjects      prometheus.Gauge
	TrackingUpdatesTotal      *prometheus.CounterVec
	TrackingErrors            *prometheus.CounterVec

	// Alert metrics
	AlertsGeneratedTotal      *prometheus.CounterVec
	AlertsAcknowledgedTotal   *prometheus.CounterVec

	// WebSocket metrics
	WebSocketConnections      prometheus.Gauge
	WebSocketMessagesTotal    *prometheus.CounterVec

	// System metrics
	MemoryUsage               prometheus.Gauge
	CPUUsage                  prometheus.Gauge
	GoroutinesCount           prometheus.Gauge
}

// NewMetrics creates and registers all Prometheus metrics (singleton)
func NewMetrics() *Metrics {
	metricsOnce.Do(func() {
		metricsInstance = &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
		),

		// Stream metrics
		ActiveStreams: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "active_streams",
				Help: "Number of currently active video streams",
			},
		),
		StreamsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "streams_total",
				Help: "Total number of video streams",
			},
			[]string{"type", "status"},
		),
		StreamErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "stream_errors_total",
				Help: "Total number of stream errors",
			},
			[]string{"stream_id", "error_type"},
		),
		StreamFramesProcessed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "stream_frames_processed_total",
				Help: "Total number of frames processed per stream",
			},
			[]string{"stream_id"},
		),

		// Detection metrics
		DetectionRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "detection_requests_total",
				Help: "Total number of detection requests",
			},
			[]string{"stream_id", "model_type"},
		),
		DetectionDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "detection_duration_seconds",
				Help:    "Duration of object detection in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
			},
			[]string{"stream_id", "model_type"},
		),
		ObjectsDetectedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "objects_detected_total",
				Help: "Total number of objects detected",
			},
			[]string{"stream_id", "object_class"},
		),
		DetectionErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "detection_errors_total",
				Help: "Total number of detection errors",
			},
			[]string{"stream_id", "error_type"},
		),
		ModelLoadDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "model_load_duration_seconds",
				Help:    "Duration of model loading in seconds",
				Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0, 60.0},
			},
		),
		FrameProcessingDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "frame_processing_duration_seconds",
				Help:    "Duration of frame processing in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5},
			},
			[]string{"stream_id"},
		),

		// Tracking metrics
		ActiveTrackedObjects: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "active_tracked_objects",
				Help: "Number of currently tracked objects",
			},
		),
		TrackingUpdatesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tracking_updates_total",
				Help: "Total number of tracking updates",
			},
			[]string{"stream_id"},
		),
		TrackingErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tracking_errors_total",
				Help: "Total number of tracking errors",
			},
			[]string{"stream_id", "error_type"},
		),

		// Alert metrics
		AlertsGeneratedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "alerts_generated_total",
				Help: "Total number of alerts generated",
			},
			[]string{"stream_id", "alert_type", "severity"},
		),
		AlertsAcknowledgedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "alerts_acknowledged_total",
				Help: "Total number of alerts acknowledged",
			},
			[]string{"stream_id", "alert_type"},
		),

		// WebSocket metrics
		WebSocketConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "websocket_connections",
				Help: "Number of active WebSocket connections",
			},
		),
		WebSocketMessagesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "websocket_messages_total",
				Help: "Total number of WebSocket messages",
			},
			[]string{"type", "direction"},
		),

		// System metrics
		MemoryUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "memory_usage_bytes",
				Help: "Current memory usage in bytes",
			},
		),
		CPUUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "cpu_usage_percent",
				Help: "Current CPU usage percentage",
			},
		),
		GoroutinesCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "goroutines_count",
				Help: "Number of active goroutines",
			},
		),
		}
	})
	return metricsInstance
}
