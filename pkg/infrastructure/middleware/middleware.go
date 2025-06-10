package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Middleware provides clean HTTP middleware for the infrastructure
type Middleware struct {
	container infrastructure.ContainerInterface
	logger    *logger.Logger
}

// NewMiddleware creates a new middleware instance
func NewMiddleware(container infrastructure.ContainerInterface, logger *logger.Logger) *Middleware {
	return &Middleware{
		container: container,
		logger:    logger,
	}
}

// Chain applies multiple middleware functions in order
func (m *Middleware) Chain(handler http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// LoggingMiddleware provides comprehensive request/response logging
func (m *Middleware) LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code and size
		wrapper := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Log incoming request
		m.logger.WithFields(map[string]interface{}{
			"method":       r.Method,
			"path":         r.URL.Path,
			"query":        r.URL.RawQuery,
			"remote_addr":  r.RemoteAddr,
			"user_agent":   r.UserAgent(),
			"content_type": r.Header.Get("Content-Type"),
			"request_id":   getRequestID(r),
		}).Info("Incoming request")

		// Process request
		next(wrapper, r)

		// Log completed request
		duration := time.Since(start)
		m.logger.WithFields(map[string]interface{}{
			"method":        r.Method,
			"path":          r.URL.Path,
			"status":        wrapper.statusCode,
			"duration":      duration.String(),
			"duration_ms":   duration.Milliseconds(),
			"response_size": wrapper.size,
			"request_id":    getRequestID(r),
		}).Info("Request completed")
	}
}

// RecoveryMiddleware handles panics gracefully
func (m *Middleware) RecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				m.logger.WithFields(map[string]interface{}{
					"panic":      err,
					"stack":      string(debug.Stack()),
					"method":     r.Method,
					"path":       r.URL.Path,
					"request_id": getRequestID(r),
				}).Error("Panic recovered")

				// Return error response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{
					"error": "Internal Server Error",
					"message": "An unexpected error occurred",
					"timestamp": "` + time.Now().Format(time.RFC3339) + `"
				}`))
			}
		}()

		next(w, r)
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func (m *Middleware) CORSMiddleware(config *CORSConfig) func(http.HandlerFunc) http.HandlerFunc {
	if config == nil {
		config = DefaultCORSConfig()
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Set CORS headers
			if config.AllowAllOrigins || isOriginAllowed(origin, config.AllowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if config.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next(w, r)
		}
	}
}

// RateLimitMiddleware provides rate limiting functionality
func (m *Middleware) RateLimitMiddleware(config *RateLimitConfig) func(http.HandlerFunc) http.HandlerFunc {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	// Create rate limiter
	limiter := rate.NewLimiter(rate.Limit(config.RequestsPerSecond), config.BurstSize)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Check rate limit
			if !limiter.Allow() {
				m.logger.WithFields(map[string]interface{}{
					"remote_addr": r.RemoteAddr,
					"path":        r.URL.Path,
					"method":      r.Method,
				}).Warn("Rate limit exceeded")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{
					"error": "Rate Limit Exceeded",
					"message": "Too many requests, please try again later",
					"timestamp": "` + time.Now().Format(time.RFC3339) + `"
				}`))
				return
			}

			next(w, r)
		}
	}
}

// AuthenticationMiddleware validates JWT tokens
func (m *Middleware) AuthenticationMiddleware(config *AuthConfig) func(http.HandlerFunc) http.HandlerFunc {
	if config == nil {
		config = DefaultAuthConfig()
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for excluded paths
			if isPathExcluded(r.URL.Path, config.ExcludedPaths) {
				next(w, r)
				return
			}

			// Extract token from header
			token := extractToken(r)
			if token == "" {
				m.respondWithAuthError(w, "Authentication required", "Missing or invalid authorization header")
				return
			}

			// Validate token using infrastructure JWT service
			jwtService := m.container.GetJWTService()
			if jwtService == nil {
				m.logger.Error("JWT service not available")
				m.respondWithAuthError(w, "Authentication service unavailable", "Internal authentication error")
				return
			}

			claims, err := jwtService.ValidateToken(r.Context(), token)
			if err != nil {
				m.logger.WithError(err).WithField("token", token[:10]+"...").Warn("Token validation failed")
				m.respondWithAuthError(w, "Invalid token", err.Error())
				return
			}

			// Add user context to request
			ctx := context.WithValue(r.Context(), "user_claims", claims)
			ctx = context.WithValue(ctx, "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "user_role", claims.Role)

			next(w, r.WithContext(ctx))
		}
	}
}

// SecurityHeadersMiddleware adds security headers
func (m *Middleware) SecurityHeadersMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next(w, r)
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func (m *Middleware) RequestIDMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Add to response header
		w.Header().Set("X-Request-ID", requestID)

		// Add to request context
		ctx := context.WithValue(r.Context(), "request_id", requestID)

		next(w, r.WithContext(ctx))
	}
}

// ValidationMiddleware validates request size and content
func (m *Middleware) ValidationMiddleware(config *ValidationConfig) func(http.HandlerFunc) http.HandlerFunc {
	if config == nil {
		config = DefaultValidationConfig()
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Limit request body size
			if config.MaxRequestSize > 0 {
				r.Body = http.MaxBytesReader(w, r.Body, config.MaxRequestSize)
			}

			next(w, r)
		}
	}
}

// MetricsMiddleware collects HTTP metrics
func (m *Middleware) MetricsMiddleware(config *MetricsConfig) func(http.HandlerFunc) http.HandlerFunc {
	if config == nil {
		config = DefaultMetricsConfig()
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create response writer wrapper
			wrapper := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Process request
			next(wrapper, r)

			// Record metrics
			duration := time.Since(start)

			// Log metrics (in a real implementation, this would send to a metrics system)
			m.logger.WithFields(map[string]interface{}{
				"service":     config.ServiceName,
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      wrapper.statusCode,
				"duration_ms": duration.Milliseconds(),
				"size":        wrapper.size,
			}).Debug("HTTP metrics")
		}
	}
}

// TimeoutMiddleware adds request timeout
func (m *Middleware) TimeoutMiddleware(timeout time.Duration) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)
			next(w, r)
		}
	}
}

// CacheMiddleware adds caching headers
func (m *Middleware) CacheMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set cache headers based on path
		if strings.HasPrefix(r.URL.Path, "/static/") {
			// Cache static files for 1 hour
			w.Header().Set("Cache-Control", "public, max-age=3600")
			w.Header().Set("ETag", fmt.Sprintf(`"%x"`, time.Now().Unix()))
		} else if strings.HasPrefix(r.URL.Path, "/api/") {
			// No cache for API endpoints
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}

		next(w, r)
	}
}
