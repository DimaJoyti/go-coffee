package http

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/security"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/google/uuid"
)

// MiddlewareChain represents a chain of HTTP middleware
type MiddlewareChain struct {
	container  infrastructure.ContainerInterface
	logger     *logger.Logger
	jwtService security.JWTService
}

// NewMiddlewareChain creates a new middleware chain
func NewMiddlewareChain(container infrastructure.ContainerInterface, logger *logger.Logger) *MiddlewareChain {
	return &MiddlewareChain{
		container:  container,
		logger:     logger,
		jwtService: container.GetJWTService(),
	}
}

// RequestID middleware adds a unique request ID to each request
func (m *MiddlewareChain) RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add to response header
		w.Header().Set("X-Request-ID", requestID)

		// Add to context
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logging middleware logs HTTP requests
func (m *MiddlewareChain) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Log request
		duration := time.Since(start)
		requestID := r.Context().Value("request_id")

		m.logger.InfoWithFields("HTTP request",
			logger.String("method", r.Method),
			logger.String("path", r.URL.Path),
			logger.String("remote_addr", r.RemoteAddr),
			logger.String("user_agent", r.UserAgent()),
			logger.Int("status_code", wrapped.statusCode),
			logger.Duration("duration", duration),
			logger.Any("request_id", requestID),
		)
	})
}

// Recovery middleware recovers from panics
func (m *MiddlewareChain) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := r.Context().Value("request_id")

				m.logger.ErrorWithFields("Panic recovered",
					logger.Any("error", err),
					logger.String("method", r.Method),
					logger.String("path", r.URL.Path),
					logger.Any("request_id", requestID),
				)

				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// CORS middleware handles Cross-Origin Resource Sharing
func (m *MiddlewareChain) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders middleware adds security headers
func (m *MiddlewareChain) SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}

// RateLimit middleware implements rate limiting using Redis
func (m *MiddlewareChain) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cache := m.container.GetCache()
		if cache == nil {
			// If cache is not available, skip rate limiting
			next.ServeHTTP(w, r)
			return
		}

		// Get client IP
		clientIP := getClientIP(r)
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		// Check current count
		var count int
		if err := cache.Get(r.Context(), key, &count); err != nil {
			// Key doesn't exist, start counting
			count = 0
		}

		// Rate limit: 100 requests per minute
		limit := 100
		window := time.Minute

		if count >= limit {
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(window).Unix(), 10))

			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Increment counter
		count++
		if err := cache.Set(r.Context(), key, count, window); err != nil {
			m.logger.WithError(err).Error("Failed to set rate limit counter")
		}

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(limit-count))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(window).Unix(), 10))

		next.ServeHTTP(w, r)
	})
}

// Authentication middleware validates JWT tokens
func (m *MiddlewareChain) Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.jwtService == nil {
			http.Error(w, "Authentication service unavailable", http.StatusServiceUnavailable)
			return
		}

		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Validate token
		claims, err := m.jwtService.ValidateAccessToken(r.Context(), token)
		if err != nil {
			m.logger.WithError(err).Debug("Token validation failed")
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)
		ctx = context.WithValue(ctx, "user_role", claims.Role)
		ctx = context.WithValue(ctx, "token_id", claims.TokenID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Authorization middleware checks user roles
func (m *MiddlewareChain) Authorization(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value("user_role").(string)
			if !ok {
				http.Error(w, "User role not found in context", http.StatusForbidden)
				return
			}

			// Check if user has required role
			if !hasRequiredRole(userRole, requiredRole) {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ContentType middleware validates content type for POST/PUT requests
func (m *MiddlewareChain) ContentType(allowedTypes []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
				contentType := r.Header.Get("Content-Type")
				if contentType == "" {
					http.Error(w, "Content-Type header required", http.StatusBadRequest)
					return
				}

				// Check if content type is allowed
				allowed := false
				for _, allowedType := range allowedTypes {
					if strings.HasPrefix(contentType, allowedType) {
						allowed = true
						break
					}
				}

				if !allowed {
					http.Error(w, "Unsupported content type", http.StatusUnsupportedMediaType)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Metrics middleware collects request metrics
func (m *MiddlewareChain) Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		// Store metrics in cache for monitoring
		cache := m.container.GetCache()
		if cache != nil {
			metricsKey := fmt.Sprintf("metrics:requests:%s:%s", r.Method, r.URL.Path)
			metrics := map[string]interface{}{
				"count":       1,
				"duration_ms": duration.Milliseconds(),
				"status_code": wrapped.statusCode,
				"timestamp":   time.Now().Unix(),
			}

			// Store metrics (fire and forget)
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				cache.Set(ctx, metricsKey, metrics, time.Hour)
			}()
		}
	})
}

// Helper functions
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to remote address
	return r.RemoteAddr
}

func hasRequiredRole(userRole, requiredRole string) bool {
	// Simple role hierarchy: admin > user > guest
	roleHierarchy := map[string]int{
		"admin": 3,
		"user":  2,
		"guest": 1,
	}

	userLevel, userExists := roleHierarchy[userRole]
	requiredLevel, requiredExists := roleHierarchy[requiredRole]

	if !userExists || !requiredExists {
		return false
	}

	return userLevel >= requiredLevel
}
