package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// SecurityMiddleware provides security-related middleware
type SecurityMiddleware struct {
	jwtService   JWTService
	sessionCache SessionCacheService
	rateLimiter  RateLimiter
	logger       *logger.Logger
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(
	jwtService JWTService,
	sessionCache SessionCacheService,
	rateLimiter RateLimiter,
	logger *logger.Logger,
) *SecurityMiddleware {
	return &SecurityMiddleware{
		jwtService:   jwtService,
		sessionCache: sessionCache,
		rateLimiter:  rateLimiter,
		logger:       logger,
	}
}

// AuthenticationMiddleware validates JWT tokens and loads user context
func (sm *SecurityMiddleware) AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sm.writeUnauthorizedResponse(w, "Missing authorization header")
			return
		}

		// Check Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			sm.writeUnauthorizedResponse(w, "Invalid authorization header format")
			return
		}

		token := parts[1]

		// Validate JWT token
		claims, err := sm.jwtService.ValidateToken(r.Context(), token)
		if err != nil {
			sm.logger.WarnWithFields("Invalid JWT token",
				logger.Error(err),
				logger.String("ip", r.RemoteAddr))
			sm.writeUnauthorizedResponse(w, "Invalid token")
			return
		}

		// Get session from cache
		session, err := sm.sessionCache.GetSessionByAccessToken(r.Context(), token)
		if err != nil {
			sm.logger.WarnWithFields("Session not found",
				logger.Error(err),
				logger.String("user_id", claims.UserID))
			sm.writeUnauthorizedResponse(w, "Session not found")
			return
		}

		// Check session status
		if session.Status != domain.SessionStatusActive {
			sm.writeUnauthorizedResponse(w, "Session is not active")
			return
		}

		// Check session expiry
		if time.Now().After(session.ExpiresAt) {
			sm.writeUnauthorizedResponse(w, "Session expired")
			return
		}

		// Update last used time
		session.LastUsedAt = &time.Time{}
		*session.LastUsedAt = time.Now()
		sm.sessionCache.UpdateSession(r.Context(), session)

		// Add user context to request
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "session_id", session.ID)
		ctx = context.WithValue(ctx, "user_role", claims.Role)
		ctx = context.WithValue(ctx, "claims", claims)
		ctx = context.WithValue(ctx, "session", session)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthorizationMiddleware checks user permissions
func (sm *SecurityMiddleware) AuthorizationMiddleware(requiredRole domain.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value("user_role").(domain.UserRole)
			if !ok {
				sm.writeForbiddenResponse(w, "User role not found in context")
				return
			}

			// Check if user has required role or higher
			if !sm.hasRequiredRole(userRole, requiredRole) {
				sm.logger.WarnWithFields("Insufficient permissions",
					logger.String("user_role", string(userRole)),
					logger.String("required_role", string(requiredRole)))
				sm.writeForbiddenResponse(w, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware applies rate limiting
func (sm *SecurityMiddleware) RateLimitMiddleware(limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use IP address as key
			key := sm.getRateLimitKey(r.RemoteAddr)

			allowed, err := sm.rateLimiter.Allow(r.Context(), key, limit, window)
			if err != nil {
				sm.logger.ErrorWithFields("Rate limiter error",
					logger.Error(err),
					logger.String("key", key))
				sm.writeInternalErrorResponse(w, "Rate limiter error")
				return
			}

			if !allowed {
				sm.logger.WarnWithFields("Rate limit exceeded",
					logger.String("ip", r.RemoteAddr),
					logger.String("path", r.URL.Path))
				sm.writeRateLimitResponse(w, "Rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersMiddleware adds security headers
func (sm *SecurityMiddleware) SecurityHeadersMiddleware(next http.Handler) http.Handler {
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

// CORSMiddleware handles CORS
func (sm *SecurityMiddleware) CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware logs HTTP requests
func (sm *SecurityMiddleware) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)

		sm.logger.InfoWithFields("HTTP request",
			logger.String("method", r.Method),
			logger.String("path", r.URL.Path),
			logger.String("remote_addr", r.RemoteAddr),
			logger.String("user_agent", r.UserAgent()),
			logger.Int("status_code", wrapper.statusCode),
			logger.Duration("duration", duration))
	})
}

// Helper methods

func (sm *SecurityMiddleware) hasRequiredRole(userRole, requiredRole domain.UserRole) bool {
	roleHierarchy := map[domain.UserRole]int{
		domain.UserRoleUser:      1,
		domain.UserRoleModerator: 2,
		domain.UserRoleAdmin:     3,
	}

	userLevel, userExists := roleHierarchy[userRole]
	requiredLevel, requiredExists := roleHierarchy[requiredRole]

	if !userExists || !requiredExists {
		return false
	}

	return userLevel >= requiredLevel
}

func (sm *SecurityMiddleware) getRateLimitKey(ip string) string {
	return fmt.Sprintf("rate_limit:%s", ip)
}

func (sm *SecurityMiddleware) writeUnauthorizedResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, message)))
}

func (sm *SecurityMiddleware) writeForbiddenResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, message)))
}

func (sm *SecurityMiddleware) writeRateLimitResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, message)))
}

func (sm *SecurityMiddleware) writeInternalErrorResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, message)))
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
