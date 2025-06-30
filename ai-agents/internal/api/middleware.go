package api

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/internal/httputils"
	"go-coffee-ai-agents/internal/observability"
)

// MiddlewareChain represents a chain of HTTP middleware
type MiddlewareChain struct {
	logger  *observability.StructuredLogger
	metrics *observability.MetricsCollector
	tracing *observability.TracingHelper
}

// NewMiddlewareChain creates a new middleware chain
func NewMiddlewareChain(
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) *MiddlewareChain {
	return &MiddlewareChain{
		logger:  logger,
		metrics: metrics,
		tracing: tracing,
	}
}

// Apply applies all middleware to the handler
func (mc *MiddlewareChain) Apply(handler http.Handler) http.Handler {
	// Apply middleware in reverse order (last applied = first executed)
	h := handler
	h = mc.Recovery(h)
	h = mc.Metrics(h)
	h = mc.Tracing(h)
	h = mc.Logging(h)
	h = mc.CORS(h)
	h = mc.RequestID(h)
	
	return h
}

// RequestID middleware adds a unique request ID to each request
func (mc *MiddlewareChain) RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)
		
		// Add request ID to context
		ctx := observability.WithRequestID(r.Context(), requestID)
		r = r.WithContext(ctx)
		
		next.ServeHTTP(w, r)
	})
}

// CORS middleware handles Cross-Origin Resource Sharing
func (mc *MiddlewareChain) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, X-Trace-ID")
		w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID, X-Trace-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// Logging middleware logs HTTP requests and responses
func (mc *MiddlewareChain) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		
		// Log request
		mc.logger.InfoContext(r.Context(), "HTTP request started",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"content_length", r.ContentLength)
		
		// Process request
		next.ServeHTTP(wrapped, r)
		
		// Calculate duration
		duration := time.Since(start)
		
		// Log response
		logLevel := "info"
		if wrapped.statusCode >= 400 {
			logLevel = "warn"
		}
		if wrapped.statusCode >= 500 {
			logLevel = "error"
		}
		
		if logLevel == "info" {
			mc.logger.InfoContext(r.Context(), "HTTP request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status_code", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"response_size", wrapped.bytesWritten)
		} else if logLevel == "warn" {
			mc.logger.WarnContext(r.Context(), "HTTP request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status_code", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"response_size", wrapped.bytesWritten)
		} else if logLevel == "error" {
			mc.logger.ErrorContext(r.Context(), "HTTP request completed", fmt.Errorf("HTTP error %d", wrapped.statusCode),
				"method", r.Method,
				"path", r.URL.Path,
				"status_code", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"response_size", wrapped.bytesWritten)
		}
	})
}

// Tracing middleware adds distributed tracing to HTTP requests
func (mc *MiddlewareChain) Tracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Start HTTP span
		ctx, span := mc.tracing.StartHTTPSpan(r.Context(), r.Method, r.URL.Path, r.UserAgent())
		defer span.End()
		
		// Add trace ID to response headers
		if traceID := observability.GetTraceID(ctx); traceID != "" {
			w.Header().Set("X-Trace-ID", traceID)
		}
		
		// Update request context
		r = r.WithContext(ctx)
		
		// Wrap response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		
		// Process request
		next.ServeHTTP(wrapped, r)
		
		// Record span attributes
		mc.tracing.SetAttributes(span,
			observability.Attribute("http.method", r.Method),
			observability.Attribute("http.url", r.URL.String()),
			observability.Attribute("http.status_code", wrapped.statusCode),
			observability.Attribute("http.user_agent", r.UserAgent()),
			observability.Attribute("http.remote_addr", r.RemoteAddr))
		
		// Record success or error
		if wrapped.statusCode >= 400 {
			mc.tracing.RecordError(span, fmt.Errorf("HTTP %d", wrapped.statusCode), "HTTP request failed")
		} else {
			mc.tracing.RecordSuccess(span, "HTTP request completed")
		}
	})
}

// Metrics middleware collects HTTP metrics
func (mc *MiddlewareChain) Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		
		// Process request
		next.ServeHTTP(wrapped, r)
		
		// Calculate duration
		duration := time.Since(start)
		
		// Record metrics
		if mc.metrics != nil {
			counters := mc.metrics.GetCounters()
			histograms := mc.metrics.GetHistograms()
			
			if counters != nil && histograms != nil {
				// Record request count
				counters.RequestsTotal.Add(r.Context(), 1)
				
				// Record success/error count
				if wrapped.statusCode >= 400 {
					counters.RequestsError.Add(r.Context(), 1)
				} else {
					counters.RequestsSuccess.Add(r.Context(), 1)
				}
				
				// Record duration
				histograms.RequestDuration.Record(r.Context(), duration.Seconds())
			}
		}
	})
}

// Recovery middleware recovers from panics and returns a 500 error
func (mc *MiddlewareChain) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				mc.logger.ErrorContext(r.Context(), "HTTP handler panic recovered",
					fmt.Errorf("panic: %v", err),
					"method", r.Method,
					"path", r.URL.Path,
					"stack", string(debug.Stack()))
				
				// Record error in tracing if available
				if span := observability.GetSpanFromContext(r.Context()); span != nil {
					mc.tracing.RecordError(span, fmt.Errorf("panic: %v", err), "HTTP handler panic")
				}
				
				// Return 500 error
				httputils.WriteErrorResponse(w, http.StatusInternalServerError, "internal_server_error", "Internal server error")
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

// Authentication middleware (placeholder for future implementation)
func (mc *MiddlewareChain) Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for health checks and public endpoints
		if isPublicEndpoint(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		
		// Extract authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			httputils.WriteErrorResponse(w, http.StatusUnauthorized, "missing_authorization", "Authorization header required")
			return
		}
		
		// Validate token (simplified - in production, use proper JWT validation)
		if !strings.HasPrefix(authHeader, "Bearer ") {
			httputils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid_authorization", "Bearer token required")
			return
		}
		
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			httputils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid_token", "Invalid token")
			return
		}
		
		// TODO: Implement proper token validation
		// For now, accept any non-empty token
		
		// Add user context (placeholder)
		ctx := observability.WithUserID(r.Context(), "user-from-token")
		r = r.WithContext(ctx)
		
		next.ServeHTTP(w, r)
	})
}

// RateLimit middleware implements rate limiting (placeholder for future implementation)
func (mc *MiddlewareChain) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement rate limiting
		// For now, just pass through
		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code and bytes written
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the number of bytes written
func (rw *responseWriter) Write(data []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(data)
	rw.bytesWritten += n
	return n, err
}

// isPublicEndpoint checks if an endpoint is public (doesn't require authentication)
func isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/health",
		"/health/ready",
		"/health/live",
		"/metrics",
		"/api/v1/docs",
		"/api/v1/openapi.json",
	}
	
	for _, publicPath := range publicPaths {
		if path == publicPath {
			return true
		}
	}
	
	return false
}

// ContentType middleware sets appropriate content type headers
func (mc *MiddlewareChain) ContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set default content type for API responses
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Content-Type", "application/json")
		}
		
		next.ServeHTTP(w, r)
	})
}

// Security middleware adds security headers
func (mc *MiddlewareChain) Security(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		next.ServeHTTP(w, r)
	})
}

// Compression middleware (placeholder for future implementation)
func (mc *MiddlewareChain) Compression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement gzip compression
		// For now, just pass through
		next.ServeHTTP(w, r)
	})
}

// Timeout middleware adds request timeout
func (mc *MiddlewareChain) Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			
			r = r.WithContext(ctx)
			
			done := make(chan struct{})
			go func() {
				defer close(done)
				next.ServeHTTP(w, r)
			}()
			
			select {
			case <-done:
				// Request completed normally
			case <-ctx.Done():
				// Request timed out
				if ctx.Err() == context.DeadlineExceeded {
					httputils.WriteErrorResponse(w, http.StatusRequestTimeout, "request_timeout", "Request timeout")
				}
			}
		})
	}
}
