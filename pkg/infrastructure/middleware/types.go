package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"
)

// CORSConfig defines CORS configuration
type CORSConfig struct {
	AllowAllOrigins  bool
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowAllOrigins: false,
		AllowedOrigins:  []string{"http://localhost:3000", "http://localhost:8080"},
		AllowedMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:  []string{"Content-Type", "Authorization", "X-Request-ID", "X-API-Key"},
		ExposedHeaders:  []string{"X-Request-ID", "X-Total-Count"},
		AllowCredentials: true,
		MaxAge:          86400, // 24 hours
	}
}

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
	Enabled           bool
}

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		RequestsPerSecond: 10.0, // 10 requests per second
		BurstSize:         20,   // Allow bursts up to 20 requests
		Enabled:           true,
	}
}

// AuthConfig defines authentication configuration
type AuthConfig struct {
	Enabled       bool
	ExcludedPaths []string
	TokenHeader   string
	TokenPrefix   string
}

// DefaultAuthConfig returns default authentication configuration
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		Enabled: true,
		ExcludedPaths: []string{
			"/health",
			"/metrics",
			"/api/docs",
			"/api/v1/auth/login",
			"/api/v1/auth/register",
			"/static/",
		},
		TokenHeader: "Authorization",
		TokenPrefix: "Bearer ",
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response size
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// Helper functions

// getRequestID extracts request ID from context or header
func getRequestID(r *http.Request) string {
	if id := r.Context().Value("request_id"); id != nil {
		if requestID, ok := id.(string); ok {
			return requestID
		}
	}
	return r.Header.Get("X-Request-ID")
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// isOriginAllowed checks if origin is in allowed list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// isPathExcluded checks if path should be excluded from authentication
func isPathExcluded(path string, excludedPaths []string) bool {
	for _, excluded := range excludedPaths {
		if strings.HasPrefix(path, excluded) {
			return true
		}
	}
	return false
}

// extractToken extracts JWT token from Authorization header
func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	
	// Check for Bearer token
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	
	// Check for API key in X-API-Key header
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return apiKey
	}
	
	return ""
}

// respondWithAuthError sends authentication error response
func (m *Middleware) respondWithAuthError(w http.ResponseWriter, message, details string) {
	m.logger.WithFields(map[string]interface{}{
		"error":   message,
		"details": details,
	}).Warn("Authentication failed")
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{
		"error": "` + message + `",
		"message": "` + details + `",
		"timestamp": "` + time.Now().Format(time.RFC3339) + `"
	}`))
}

// MetricsConfig defines metrics collection configuration
type MetricsConfig struct {
	Enabled     bool
	ServiceName string
	Namespace   string
}

// DefaultMetricsConfig returns default metrics configuration
func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		Enabled:     true,
		ServiceName: "go-coffee",
		Namespace:   "http",
	}
}

// ValidationConfig defines request validation configuration
type ValidationConfig struct {
	MaxRequestSize int64
	Enabled        bool
}

// DefaultValidationConfig returns default validation configuration
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		MaxRequestSize: 1 << 20, // 1MB
		Enabled:        true,
	}
}

// CompressionConfig defines response compression configuration
type CompressionConfig struct {
	Enabled           bool
	MinSize           int
	CompressionLevel  int
	ExcludedMimeTypes []string
}

// DefaultCompressionConfig returns default compression configuration
func DefaultCompressionConfig() *CompressionConfig {
	return &CompressionConfig{
		Enabled:          true,
		MinSize:          1024, // 1KB
		CompressionLevel: 6,    // Default gzip level
		ExcludedMimeTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"video/mp4",
			"application/zip",
		},
	}
}

// MiddlewareConfig aggregates all middleware configurations
type MiddlewareConfig struct {
	CORS        *CORSConfig
	RateLimit   *RateLimitConfig
	Auth        *AuthConfig
	Metrics     *MetricsConfig
	Validation  *ValidationConfig
	Compression *CompressionConfig
}

// DefaultMiddlewareConfig returns default configuration for all middleware
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		CORS:        DefaultCORSConfig(),
		RateLimit:   DefaultRateLimitConfig(),
		Auth:        DefaultAuthConfig(),
		Metrics:     DefaultMetricsConfig(),
		Validation:  DefaultValidationConfig(),
		Compression: DefaultCompressionConfig(),
	}
}
