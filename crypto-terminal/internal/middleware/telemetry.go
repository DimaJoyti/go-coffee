package middleware

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	// Trace and metric names
	tracerName = "crypto-terminal"
	meterName  = "crypto-terminal"
)

var (
	tracer trace.Tracer
	meter  metric.Meter

	// Metrics
	httpRequestsTotal    metric.Int64Counter
	httpRequestDuration  metric.Float64Histogram
	httpRequestsInFlight metric.Int64UpDownCounter
	wsConnectionsTotal   metric.Int64UpDownCounter
	wsMessagesTotal      metric.Int64Counter
)

// InitTelemetryMiddleware initializes telemetry middleware
func InitTelemetryMiddleware() error {
	tracer = otel.Tracer(tracerName)
	meter = otel.Meter(meterName)

	var err error

	// Initialize HTTP metrics
	httpRequestsTotal, err = meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		return err
	}

	httpRequestDuration, err = meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	httpRequestsInFlight, err = meter.Int64UpDownCounter(
		"http_requests_in_flight",
		metric.WithDescription("Number of HTTP requests currently being processed"),
	)
	if err != nil {
		return err
	}

	// Initialize WebSocket metrics
	wsConnectionsTotal, err = meter.Int64UpDownCounter(
		"websocket_connections_total",
		metric.WithDescription("Total number of active WebSocket connections"),
	)
	if err != nil {
		return err
	}

	wsMessagesTotal, err = meter.Int64Counter(
		"websocket_messages_total",
		metric.WithDescription("Total number of WebSocket messages"),
	)
	if err != nil {
		return err
	}

	return nil
}

// TracingMiddleware adds OpenTelemetry tracing to HTTP requests
func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start a new span
		ctx, span := tracer.Start(c.Request.Context(), c.Request.Method+" "+c.FullPath())
		defer span.End()

		// Add request ID for correlation
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			c.Header("X-Request-ID", requestID)
		}

		// Add span attributes
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.route", c.FullPath()),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("request.id", requestID),
		)

		// Add request ID to context for logging
		c.Set("request_id", requestID)
		c.Set("trace_id", span.SpanContext().TraceID().String())

		// Update context with span
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Add response attributes
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int("http.response_size", c.Writer.Size()),
		)

		// Set span status based on HTTP status code
		if c.Writer.Status() >= 400 {
			span.SetAttributes(attribute.String("error", "true"))
			if len(c.Errors) > 0 {
				span.SetAttributes(attribute.String("error.message", c.Errors.String()))
			}
		}
	}
}

// MetricsMiddleware adds Prometheus metrics to HTTP requests
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Increment in-flight requests
		httpRequestsInFlight.Add(c.Request.Context(), 1,
			metric.WithAttributes(
				attribute.String("method", c.Request.Method),
				attribute.String("route", c.FullPath()),
			),
		)

		// Process request
		c.Next()

		// Decrement in-flight requests
		httpRequestsInFlight.Add(c.Request.Context(), -1,
			metric.WithAttributes(
				attribute.String("method", c.Request.Method),
				attribute.String("route", c.FullPath()),
			),
		)

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		attrs := metric.WithAttributes(
			attribute.String("method", c.Request.Method),
			attribute.String("route", c.FullPath()),
			attribute.String("status", status),
		)

		httpRequestsTotal.Add(c.Request.Context(), 1, attrs)
		httpRequestDuration.Record(c.Request.Context(), duration, attrs)
	}
}

// LoggingMiddleware adds structured logging with correlation IDs
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Get correlation IDs
		requestID := c.GetString("request_id")
		traceID := c.GetString("trace_id")

		// Create logger with context
		logger := logrus.WithFields(logrus.Fields{
			"request_id": requestID,
			"trace_id":   traceID,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		})

		// Log request start
		logger.Info("Request started")

		// Process request
		c.Next()

		// Log request completion
		duration := time.Since(start)
		logger.WithFields(logrus.Fields{
			"status":   c.Writer.Status(),
			"duration": duration,
			"size":     c.Writer.Size(),
		}).Info("Request completed")

		// Log errors if any
		if len(c.Errors) > 0 {
			logger.WithField("errors", c.Errors.String()).Error("Request completed with errors")
		}
	}
}

// ErrorHandlingMiddleware provides centralized error handling
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			// Get correlation IDs for error tracking
			requestID := c.GetString("request_id")
			traceID := c.GetString("trace_id")

			// Log error with context
			logrus.WithFields(logrus.Fields{
				"request_id": requestID,
				"trace_id":   traceID,
				"error":      err.Error(),
				"type":       err.Type,
			}).Error("Request error")

			// Return appropriate error response
			switch err.Type {
			case gin.ErrorTypeBind:
				c.JSON(400, gin.H{
					"error":      "Invalid request format",
					"message":    "Please check your request format and try again",
					"request_id": requestID,
				})
			case gin.ErrorTypePublic:
				c.JSON(500, gin.H{
					"error":      "Internal server error",
					"message":    "An unexpected error occurred",
					"request_id": requestID,
				})
			default:
				c.JSON(500, gin.H{
					"error":      "Internal server error",
					"message":    "An unexpected error occurred",
					"request_id": requestID,
				})
			}
		}
	}
}

// RecoveryMiddleware provides panic recovery with telemetry
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// Get correlation IDs
		requestID := c.GetString("request_id")
		traceID := c.GetString("trace_id")

		// Log panic with context
		logrus.WithFields(logrus.Fields{
			"request_id": requestID,
			"trace_id":   traceID,
			"panic":      recovered,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
		}).Error("Request panic recovered")

		// Record panic metric
		httpRequestsTotal.Add(c.Request.Context(), 1,
			metric.WithAttributes(
				attribute.String("method", c.Request.Method),
				attribute.String("route", c.FullPath()),
				attribute.String("status", "500"),
				attribute.String("error", "panic"),
			),
		)

		// Return error response
		c.JSON(500, gin.H{
			"error":      "Internal server error",
			"message":    "An unexpected error occurred",
			"request_id": requestID,
		})
	})
}

// WebSocketMetrics provides metrics for WebSocket connections
type WebSocketMetrics struct {
	ctx context.Context
}

// NewWebSocketMetrics creates a new WebSocket metrics instance
func NewWebSocketMetrics(ctx context.Context) *WebSocketMetrics {
	return &WebSocketMetrics{ctx: ctx}
}

// ConnectionOpened records a new WebSocket connection
func (m *WebSocketMetrics) ConnectionOpened() {
	wsConnectionsTotal.Add(m.ctx, 1)
}

// ConnectionClosed records a closed WebSocket connection
func (m *WebSocketMetrics) ConnectionClosed() {
	wsConnectionsTotal.Add(m.ctx, -1)
}

// MessageSent records a sent WebSocket message
func (m *WebSocketMetrics) MessageSent(messageType string) {
	wsMessagesTotal.Add(m.ctx, 1,
		metric.WithAttributes(
			attribute.String("direction", "sent"),
			attribute.String("type", messageType),
		),
	)
}

// MessageReceived records a received WebSocket message
func (m *WebSocketMetrics) MessageReceived(messageType string) {
	wsMessagesTotal.Add(m.ctx, 1,
		metric.WithAttributes(
			attribute.String("direction", "received"),
			attribute.String("type", messageType),
		),
	)
}
