package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// corsMiddleware adds CORS headers to responses
func corsMiddleware(next http.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler, logger Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapper := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)

		logger.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapper.statusCode,
			"duration", duration.String(),
			"remote_addr", r.RemoteAddr,
			"user_agent", r.Header.Get("User-Agent"),
		)
	})
}

// responseWriterWrapper wraps http.ResponseWriter to capture status code
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// rateLimitMiddleware implements rate limiting
func rateLimitMiddleware(next http.Handler, rps, burst int) http.Handler {
	// Create a map to store rate limiters per IP
	limiters := make(map[string]*rate.Limiter)
	var mu sync.RWMutex

	// Cleanup old limiters periodically
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		
		for range ticker.C {
			mu.Lock()
			// In a real implementation, you would track last access time
			// and remove old limiters. For simplicity, we'll skip this.
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract client IP
		ip := getClientIP(r)

		// Get or create rate limiter for this IP
		mu.RLock()
		limiter, exists := limiters[ip]
		mu.RUnlock()

		if !exists {
			limiter = rate.NewLimiter(rate.Limit(rps), burst)
			mu.Lock()
			limiters[ip] = limiter
			mu.Unlock()
		}

		// Check if request is allowed
		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// compressionMiddleware adds gzip compression
func compressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Check if response should be compressed
		if !shouldCompress(r) {
			next.ServeHTTP(w, r)
			return
		}

		// Create gzip writer
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		gzipWriter := gzip.NewWriter(w)
		defer gzipWriter.Close()

		// Create wrapper that writes to gzip writer
		gzipResponseWriter := &gzipResponseWriter{
			ResponseWriter: w,
			Writer:         gzipWriter,
		}

		next.ServeHTTP(gzipResponseWriter, r)
	})
}

// gzipResponseWriter wraps http.ResponseWriter to write compressed responses
type gzipResponseWriter struct {
	http.ResponseWriter
	io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// authenticationMiddleware handles authentication (placeholder)
func authenticationMiddleware(next http.Handler, jwtSecret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for health checks and public endpoints
		if isPublicEndpoint(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Validate token (simplified implementation)
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if !isValidToken(token, jwtSecret) {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// securityHeadersMiddleware adds security headers
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}

// timeoutMiddleware adds request timeout
func timeoutMiddleware(next http.Handler, timeout time.Duration) http.Handler {
	return http.TimeoutHandler(next, timeout, "Request timeout")
}

// recoveryMiddleware recovers from panics
func recoveryMiddleware(next http.Handler, logger Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered", fmt.Errorf("%v", err),
					"method", r.Method,
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
				)

				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// metricsMiddleware collects HTTP metrics (placeholder)
func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapper := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)

		// In a real implementation, you would collect metrics here
		// For example, using Prometheus metrics
		_ = duration
		_ = wrapper.statusCode
	})
}

// Helper functions

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		if ips := strings.Split(xff, ","); len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	if ip := strings.Split(r.RemoteAddr, ":"); len(ip) > 0 {
		return ip[0]
	}

	return r.RemoteAddr
}

func shouldCompress(r *http.Request) bool {
	// Don't compress WebSocket upgrades
	if r.Header.Get("Upgrade") == "websocket" {
		return false
	}

	// Don't compress already compressed content
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "gzip") ||
		strings.Contains(contentType, "compress") ||
		strings.Contains(contentType, "deflate") {
		return false
	}

	// Compress text-based content
	compressibleTypes := []string{
		"text/",
		"application/json",
		"application/javascript",
		"application/xml",
		"application/rss+xml",
		"application/atom+xml",
		"image/svg+xml",
	}

	for _, compressibleType := range compressibleTypes {
		if strings.Contains(contentType, compressibleType) {
			return true
		}
	}

	// Default to compressing if no content type is set (likely JSON API)
	return contentType == ""
}

func isPublicEndpoint(path string) bool {
	publicEndpoints := []string{
		"/health",
		"/ready",
		"/metrics",
		"/status",
		"/api/v1/system/status",
	}

	for _, endpoint := range publicEndpoints {
		if path == endpoint {
			return true
		}
	}

	return false
}

func isValidToken(token, secret string) bool {
	// Simplified token validation
	// In a real implementation, you would validate JWT tokens properly
	return token != "" && secret != ""
}

// RequestIDMiddleware adds a unique request ID to each request
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate or extract request ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		// Add request ID to context (in a real implementation)
		// ctx := context.WithValue(r.Context(), "request_id", requestID)
		// r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func generateRequestID() string {
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}

// ContentTypeMiddleware ensures JSON content type for API endpoints
func contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set default content type for API endpoints
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Content-Type", "application/json")
		}

		next.ServeHTTP(w, r)
	})
}

// HealthCheckMiddleware provides a simple health check
func healthCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add health check headers
		w.Header().Set("X-Health-Check", "ok")
		w.Header().Set("X-Service", "orchestration-engine")
		w.Header().Set("X-Version", "1.0.0")

		next.ServeHTTP(w, r)
	})
}
