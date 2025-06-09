package http

import (
	"net/http"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Middleware represents HTTP middleware for DeFi service
type Middleware struct {
	logger *logger.Logger
}

// NewMiddleware creates a new middleware instance
func NewMiddleware(logger *logger.Logger) *Middleware {
	return &Middleware{
		logger: logger,
	}
}

// LoggingMiddleware logs all requests
func (m *Middleware) LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		m.logger.WithFields(map[string]interface{}{
			"method":      r.Method,
			"path":        r.URL.Path,
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
		}).Info("Incoming DeFi request")
		
		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next(wrapper, r)
		
		duration := time.Since(start)
		m.logger.WithFields(map[string]interface{}{
			"method":   r.Method,
			"path":     r.URL.Path,
			"status":   wrapper.statusCode,
			"duration": duration.String(),
		}).Info("DeFi request completed")
	}
}

// RecoveryMiddleware recovers from panics
func (m *Middleware) RecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				m.logger.WithFields(map[string]interface{}{
					"error": err,
					"path":  r.URL.Path,
				}).Error("Panic recovered in DeFi service")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next(w, r)
	}
}

// CORSMiddleware adds CORS headers
func (m *Middleware) CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next(w, r)
	}
}

// RateLimitMiddleware implements basic rate limiting (placeholder)
func (m *Middleware) RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting for health checks
		if r.URL.Path == "/health" {
			next(w, r)
			return
		}
		
		// In a real implementation, implement rate limiting here
		// For now, just pass through
		next(w, r)
	}
}

// AuthMiddleware validates authentication for DeFi operations
func (m *Middleware) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health checks and public endpoints
		if r.URL.Path == "/health" || 
		   r.URL.Path == "/api/v1/tokens/price" ||
		   r.URL.Path == "/api/v1/pools" ||
		   r.URL.Path == "/api/v1/arbitrage/opportunities" ||
		   r.URL.Path == "/api/v1/yield/opportunities" {
			next(w, r)
			return
		}
		
		// Check for API key in header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// Check Authorization header
			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}
		}
		
		// In a real implementation, validate the API key or JWT token here
		// For now, just pass through if any auth header is present
		next(w, r)
	}
}

// SecurityHeadersMiddleware adds security headers
func (m *Middleware) SecurityHeadersMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		
		next(w, r)
	}
}

// MetricsMiddleware adds metrics collection (placeholder)
func (m *Middleware) MetricsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next(wrapper, r)
		
		duration := time.Since(start)
		
		// In a real implementation, collect metrics here
		m.logger.WithFields(map[string]interface{}{
			"endpoint": r.URL.Path,
			"method":   r.Method,
			"status":   wrapper.statusCode,
			"duration": duration.Milliseconds(),
		}).Debug("DeFi API metrics")
	}
}

// Chain applies multiple middleware functions
func (m *Middleware) Chain(handler http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
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
