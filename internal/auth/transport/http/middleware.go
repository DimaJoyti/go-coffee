package http

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/application"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// AuthMiddleware provides authentication middleware for HTTP requests
type AuthMiddleware struct {
	authService application.AuthService
	logger      *logger.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService application.AuthService, logger *logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// Middleware returns the HTTP middleware function
func (am *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			am.respondWithError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		// Check Bearer token format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			am.respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			am.respondWithError(w, http.StatusUnauthorized, "Empty token")
			return
		}

		// Validate token
		validateReq := &application.ValidateTokenRequest{
			Token: token,
		}

		validateResp, err := am.authService.ValidateToken(r.Context(), validateReq)
		if err != nil {
			am.logger.WithError(err).WithField("token", token[:10]+"...").Error("Token validation failed")
			am.respondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		if !validateResp.Valid {
			am.respondWithError(w, http.StatusUnauthorized, "Token is not valid")
			return
		}

		// Add user information to request context
		ctx := context.WithValue(r.Context(), "user_id", validateResp.UserID)
		ctx = context.WithValue(ctx, "user_role", validateResp.Role)
		ctx = context.WithValue(ctx, "session_id", validateResp.SessionID)

		// Continue to next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// respondWithError sends an error response
func (am *AuthMiddleware) respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

// CORS middleware
func (h *Handler) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // In production, specify allowed origins
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Rate limiting middleware (simplified implementation)
func (h *Handler) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In a production environment, you would implement proper rate limiting
		// using Redis or an in-memory store with sliding window or token bucket algorithms

		// For now, we'll just pass through
		next.ServeHTTP(w, r)
	})
}

// Security headers middleware
func (h *Handler) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}

// Role-based access control middleware
func (h *Handler) requireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value("user_role").(string)
			if !ok {
				h.respondWithError(w, http.StatusUnauthorized, "User role not found")
				return
			}

			if userRole != role && userRole != "admin" { // Admin can access everything
				h.respondWithError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Admin only middleware
func (h *Handler) requireAdmin(next http.Handler) http.Handler {
	return h.requireRole("admin")(next)
}

// Request ID middleware
func (h *Handler) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		// Add request ID to context
		ctx := context.WithValue(r.Context(), "request_id", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// In production, use a proper UUID library
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// Device fingerprinting middleware
func (h *Handler) deviceFingerprintMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract device information
		userAgent := r.UserAgent()
		acceptLanguage := r.Header.Get("Accept-Language")
		acceptEncoding := r.Header.Get("Accept-Encoding")

		// Create device fingerprint
		fingerprint := createDeviceFingerprint(userAgent, acceptLanguage, acceptEncoding)

		// Add to context
		ctx := context.WithValue(r.Context(), "device_fingerprint", fingerprint)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// createDeviceFingerprint creates a device fingerprint from request headers
func createDeviceFingerprint(userAgent, acceptLanguage, acceptEncoding string) string {
	// In production, use a proper hashing algorithm
	data := userAgent + "|" + acceptLanguage + "|" + acceptEncoding
	return fmt.Sprintf("fp_%x", sha256.Sum256([]byte(data)))
}

// IP whitelist middleware
func (h *Handler) ipWhitelistMiddleware(allowedIPs []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := h.getClientIP(r)

			// Check if IP is in whitelist
			allowed := false
			for _, ip := range allowedIPs {
				if clientIP == ip {
					allowed = true
					break
				}
			}

			if !allowed {
				h.logger.WithField("ip", clientIP).Warn("IP not in whitelist")
				h.respondWithError(w, http.StatusForbidden, "Access denied")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Content type validation middleware
func (h *Handler) contentTypeMiddleware(allowedTypes []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip for GET requests
			if r.Method == "GET" || r.Method == "DELETE" {
				next.ServeHTTP(w, r)
				return
			}

			contentType := r.Header.Get("Content-Type")

			// Check if content type is allowed
			allowed := false
			for _, allowedType := range allowedTypes {
				if strings.Contains(contentType, allowedType) {
					allowed = true
					break
				}
			}

			if !allowed {
				h.respondWithError(w, http.StatusUnsupportedMediaType, "Unsupported content type")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
