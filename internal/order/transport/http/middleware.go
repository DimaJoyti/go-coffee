package http

import (
	"net/http"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Middleware represents HTTP middleware
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
		}).Info("Incoming request")

		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next(wrapper, r)

		duration := time.Since(start)
		m.logger.WithFields(map[string]interface{}{
			"method":   r.Method,
			"path":     r.URL.Path,
			"status":   wrapper.statusCode,
			"duration": duration.String(),
		}).Info("Request completed")
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
				}).Error("Panic recovered")
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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// AuthMiddleware validates authentication (placeholder)
func (m *Middleware) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health checks
		if r.URL.Path == "/health" {
			next(w, r)
			return
		}

		// In a real implementation, validate JWT token here
		// For now, just pass through
		next(w, r)
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
