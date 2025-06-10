package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/monitoring"
)

// PerformanceMiddleware provides performance monitoring and metrics collection
func (m *Middleware) PerformanceMiddleware(prometheusMetrics *monitoring.PrometheusMetrics, config *PerformanceConfig) func(http.HandlerFunc) http.HandlerFunc {
	if config == nil {
		config = DefaultPerformanceConfig()
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create enhanced response writer to capture metrics
			wrapper := &performanceResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				size:           0,
			}

			// Add performance context
			ctx := context.WithValue(r.Context(), "performance_start", start)
			ctx = context.WithValue(ctx, "performance_config", config)

			// Process request
			next(wrapper, r.WithContext(ctx))

			// Calculate metrics
			duration := time.Since(start)
			requestSize := getRequestSize(r)
			responseSize := wrapper.size

			// Record metrics if Prometheus is available
			if prometheusMetrics != nil && config.EnableMetrics {
				prometheusMetrics.RecordHTTPRequest(
					r.Method,
					sanitizePath(r.URL.Path),
					wrapper.statusCode,
					duration,
					requestSize,
					responseSize,
				)
			}

			// Log performance metrics if enabled
			if config.EnableLogging {
				m.logPerformanceMetrics(r, wrapper.statusCode, duration, requestSize, responseSize)
			}

			// Check for slow requests
			if config.SlowRequestThreshold > 0 && duration > config.SlowRequestThreshold {
				m.logger.WithFields(map[string]interface{}{
					"method":        r.Method,
					"path":          r.URL.Path,
					"duration":      duration.String(),
					"duration_ms":   duration.Milliseconds(),
					"status":        wrapper.statusCode,
					"request_size":  requestSize,
					"response_size": responseSize,
					"user_agent":    r.UserAgent(),
					"remote_addr":   r.RemoteAddr,
				}).Warn("Slow request detected")
			}
		}
	}
}

// TracingMiddleware provides request tracing capabilities
func (m *Middleware) TracingMiddleware(config *TracingConfig) func(http.HandlerFunc) http.HandlerFunc {
	if config == nil {
		config = DefaultTracingConfig()
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !config.Enabled {
				next(w, r)
				return
			}

			// Generate trace ID if not present
			traceID := r.Header.Get("X-Trace-ID")
			if traceID == "" {
				traceID = generateTraceID()
			}

			// Generate span ID
			spanID := generateSpanID()

			// Add tracing headers to response
			w.Header().Set("X-Trace-ID", traceID)
			w.Header().Set("X-Span-ID", spanID)

			// Add tracing context
			ctx := context.WithValue(r.Context(), "trace_id", traceID)
			ctx = context.WithValue(ctx, "span_id", spanID)
			ctx = context.WithValue(ctx, "trace_start", time.Now())

			// Log trace start
			if config.EnableLogging {
				m.logger.WithFields(map[string]interface{}{
					"trace_id": traceID,
					"span_id":  spanID,
					"method":   r.Method,
					"path":     r.URL.Path,
				}).Debug("Request trace started")
			}

			// Process request
			next(w, r.WithContext(ctx))

			// Log trace completion
			if config.EnableLogging {
				duration := time.Since(ctx.Value("trace_start").(time.Time))
				m.logger.WithFields(map[string]interface{}{
					"trace_id":    traceID,
					"span_id":     spanID,
					"duration":    duration.String(),
					"duration_ms": duration.Milliseconds(),
				}).Debug("Request trace completed")
			}
		}
	}
}

// ProfilingMiddleware provides request profiling capabilities
func (m *Middleware) ProfilingMiddleware(config *ProfilingConfig) func(http.HandlerFunc) http.HandlerFunc {
	if config == nil {
		config = DefaultProfilingConfig()
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !config.Enabled {
				next(w, r)
				return
			}

			// Check if this request should be profiled
			if !shouldProfile(r, config) {
				next(w, r)
				return
			}

			start := time.Now()

			// Add profiling context
			ctx := context.WithValue(r.Context(), "profiling_enabled", true)
			ctx = context.WithValue(ctx, "profiling_start", start)

			// Process request with profiling
			next(w, r.WithContext(ctx))

			// Record profiling data
			duration := time.Since(start)
			if duration > config.MinDuration {
				m.recordProfilingData(r, duration, config)
			}
		}
	}
}

// Configuration types

// PerformanceConfig defines performance monitoring configuration
type PerformanceConfig struct {
	EnableMetrics        bool          `yaml:"enable_metrics"`
	EnableLogging        bool          `yaml:"enable_logging"`
	SlowRequestThreshold time.Duration `yaml:"slow_request_threshold"`
	DetailedLogging      bool          `yaml:"detailed_logging"`
}

// DefaultPerformanceConfig returns default performance configuration
func DefaultPerformanceConfig() *PerformanceConfig {
	return &PerformanceConfig{
		EnableMetrics:        true,
		EnableLogging:        true,
		SlowRequestThreshold: 1 * time.Second,
		DetailedLogging:      false,
	}
}

// TracingConfig defines request tracing configuration
type TracingConfig struct {
	Enabled       bool    `yaml:"enabled"`
	EnableLogging bool    `yaml:"enable_logging"`
	SampleRate    float64 `yaml:"sample_rate"`
}

// DefaultTracingConfig returns default tracing configuration
func DefaultTracingConfig() *TracingConfig {
	return &TracingConfig{
		Enabled:       true,
		EnableLogging: false, // Set to true for debug mode
		SampleRate:    1.0,   // Sample all requests
	}
}

// ProfilingConfig defines request profiling configuration
type ProfilingConfig struct {
	Enabled     bool          `yaml:"enabled"`
	SampleRate  float64       `yaml:"sample_rate"`
	MinDuration time.Duration `yaml:"min_duration"`
	MaxSamples  int           `yaml:"max_samples"`
}

// DefaultProfilingConfig returns default profiling configuration
func DefaultProfilingConfig() *ProfilingConfig {
	return &ProfilingConfig{
		Enabled:     false, // Disabled by default for performance
		SampleRate:  0.01,  // 1% sampling
		MinDuration: 100 * time.Millisecond,
		MaxSamples:  1000,
	}
}

// Helper types and functions

// performanceResponseWriter wraps http.ResponseWriter to capture metrics
type performanceResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int64
}

func (prw *performanceResponseWriter) WriteHeader(code int) {
	prw.statusCode = code
	prw.ResponseWriter.WriteHeader(code)
}

func (prw *performanceResponseWriter) Write(b []byte) (int, error) {
	size, err := prw.ResponseWriter.Write(b)
	prw.size += int64(size)
	return size, err
}

// Helper functions

// getRequestSize estimates the request size
func getRequestSize(r *http.Request) int64 {
	size := int64(0)

	// Add URL length
	size += int64(len(r.URL.String()))

	// Add headers size
	for name, values := range r.Header {
		size += int64(len(name))
		for _, value := range values {
			size += int64(len(value))
		}
	}

	// Add content length if available
	if r.ContentLength > 0 {
		size += r.ContentLength
	}

	return size
}

// sanitizePath removes sensitive information from paths for metrics
func sanitizePath(path string) string {
	// In a real implementation, you would replace dynamic segments
	// For now, return the path as-is
	return path
}

// logPerformanceMetrics logs detailed performance metrics
func (m *Middleware) logPerformanceMetrics(r *http.Request, statusCode int, duration time.Duration, requestSize, responseSize int64) {
	m.logger.WithFields(map[string]interface{}{
		"method":        r.Method,
		"path":          r.URL.Path,
		"status":        statusCode,
		"duration":      duration.String(),
		"duration_ms":   duration.Milliseconds(),
		"request_size":  requestSize,
		"response_size": responseSize,
		"user_agent":    r.UserAgent(),
		"remote_addr":   r.RemoteAddr,
	}).Info("Request performance metrics")
}

// generateTraceID generates a unique trace ID
func generateTraceID() string {
	return generateRequestID() + "-trace"
}

// generateSpanID generates a unique span ID
func generateSpanID() string {
	return generateRequestID() + "-span"
}

// shouldProfile determines if a request should be profiled
func shouldProfile(r *http.Request, config *ProfilingConfig) bool {
	// Simple sampling based on sample rate
	// In a real implementation, you might use more sophisticated logic
	return true // For demo purposes, profile all requests when enabled
}

// recordProfilingData records profiling data for analysis
func (m *Middleware) recordProfilingData(r *http.Request, duration time.Duration, config *ProfilingConfig) {
	m.logger.WithFields(map[string]interface{}{
		"method":   r.Method,
		"path":     r.URL.Path,
		"duration": duration.String(),
		"profile":  true,
	}).Debug("Request profiling data recorded")
}

// Context helper functions

// GetTraceIDFromContext extracts trace ID from request context
func GetTraceIDFromContext(ctx context.Context) string {
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

// GetSpanIDFromContext extracts span ID from request context
func GetSpanIDFromContext(ctx context.Context) string {
	if spanID := ctx.Value("span_id"); spanID != nil {
		if id, ok := spanID.(string); ok {
			return id
		}
	}
	return ""
}
