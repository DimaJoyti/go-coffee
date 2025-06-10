package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/application/queries"
	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/internal/auth/infrastructure/cache"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	jwtService   JWTService
	sessionCache *cache.SessionCacheService
	queryBus     QueryBus
	logger       *logger.Logger
	config       *AuthConfig
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	TokenHeader    string        `yaml:"token_header"`
	TokenPrefix    string        `yaml:"token_prefix"`
	SkipPaths      []string      `yaml:"skip_paths"`
	SessionTimeout time.Duration `yaml:"session_timeout"`
	RequireHTTPS   bool          `yaml:"require_https"`
	CookieSecure   bool          `yaml:"cookie_secure"`
	CookieHTTPOnly bool          `yaml:"cookie_http_only"`
	CookieSameSite string        `yaml:"cookie_same_site"`
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(
	jwtService JWTService,
	sessionCache *cache.SessionCacheService,
	queryBus QueryBus,
	logger *logger.Logger,
	config *AuthConfig,
) *AuthMiddleware {
	if config == nil {
		config = &AuthConfig{
			TokenHeader:    "Authorization",
			TokenPrefix:    "Bearer",
			SessionTimeout: 30 * time.Minute,
			RequireHTTPS:   false, // Set to true in production
			CookieSecure:   false, // Set to true in production
			CookieHTTPOnly: true,
			CookieSameSite: "Strict",
		}
	}

	return &AuthMiddleware{
		jwtService:   jwtService,
		sessionCache: sessionCache,
		queryBus:     queryBus,
		logger:       logger,
		config:       config,
	}
}

// RequireAuth middleware that requires authentication
func (am *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for certain paths
		if am.shouldSkipAuth(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Check HTTPS requirement
		if am.config.RequireHTTPS && r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
			am.writeErrorResponse(w, http.StatusUpgradeRequired, "HTTPS required")
			return
		}

		// Extract and validate token
		token, err := am.extractToken(r)
		if err != nil {
			am.logger.WarnWithFields("Token extraction failed",
				logger.Error(err),
				logger.String("path", r.URL.Path),
				logger.String("ip", r.RemoteAddr))
			am.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// Validate JWT token
		claims, err := am.jwtService.ValidateToken(r.Context(), token)
		if err != nil {
			am.logger.WarnWithFields("Token validation failed",
				logger.Error(err),
				logger.String("ip", r.RemoteAddr))
			am.writeErrorResponse(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Get session from cache
		session, err := am.sessionCache.GetSessionByAccessToken(r.Context(), token)
		if err != nil {
			am.logger.WarnWithFields("Session not found",
				logger.Error(err),
				logger.String("user_id", claims.UserID))
			am.writeErrorResponse(w, http.StatusUnauthorized, "Session not found")
			return
		}

		// Validate session
		if err := am.validateSession(session); err != nil {
			am.logger.WarnWithFields("Session validation failed",
				logger.Error(err),
				logger.String("session_id", session.ID))
			am.writeErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		// Get user information
		user, err := am.getUserInfo(r.Context(), claims.UserID)
		if err != nil {
			am.logger.ErrorWithFields("Failed to get user info",
				logger.Error(err),
				logger.String("user_id", claims.UserID))
			am.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user info")
			return
		}

		// Update session activity
		am.updateSessionActivity(r.Context(), session)

		// Add authentication context
		ctx := am.addAuthContext(r.Context(), claims, session, user)

		// Continue with authenticated request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole middleware that requires specific role
func (am *AuthMiddleware) RequireRole(role domain.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := am.getUserRoleFromContext(r.Context())
			if userRole == "" {
				am.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required")
				return
			}

			if !am.hasRequiredRole(domain.UserRole(userRole), role) {
				am.logger.WarnWithFields("Insufficient permissions",
					logger.String("user_role", userRole),
					logger.String("required_role", string(role)),
					logger.String("path", r.URL.Path))
				am.writeErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission middleware that requires specific permission
func (am *AuthMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := am.getUserIDFromContext(r.Context())
			if userID == "" {
				am.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required")
				return
			}

			// Check if user has the required permission
			hasPermission, err := am.checkUserPermission(r.Context(), userID, permission)
			if err != nil {
				am.logger.ErrorWithFields("Permission check failed",
					logger.Error(err),
					logger.String("user_id", userID),
					logger.String("permission", permission))
				am.writeErrorResponse(w, http.StatusInternalServerError, "Permission check failed")
				return
			}

			if !hasPermission {
				am.logger.WarnWithFields("Permission denied",
					logger.String("user_id", userID),
					logger.String("permission", permission),
					logger.String("path", r.URL.Path))
				am.writeErrorResponse(w, http.StatusForbidden, "Permission denied")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// OptionalAuth middleware that optionally authenticates
func (am *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to extract and validate token
		token, err := am.extractToken(r)
		if err != nil {
			// No token provided, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Validate JWT token
		claims, err := am.jwtService.ValidateToken(r.Context(), token)
		if err != nil {
			// Invalid token, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Get session from cache
		session, err := am.sessionCache.GetSessionByAccessToken(r.Context(), token)
		if err != nil {
			// Session not found, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Validate session
		if err := am.validateSession(session); err != nil {
			// Invalid session, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Get user information
		user, err := am.getUserInfo(r.Context(), claims.UserID)
		if err != nil {
			// Failed to get user info, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Update session activity
		am.updateSessionActivity(r.Context(), session)

		// Add authentication context
		ctx := am.addAuthContext(r.Context(), claims, session, user)

		// Continue with optionally authenticated request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper methods

func (am *AuthMiddleware) extractToken(r *http.Request) (string, error) {
	// Try Authorization header first
	authHeader := r.Header.Get(am.config.TokenHeader)
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == am.config.TokenPrefix {
			return parts[1], nil
		}
	}

	// Try cookie as fallback
	cookie, err := r.Cookie("access_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	return "", domain.ErrTokenNotFound
}

func (am *AuthMiddleware) validateSession(session *domain.Session) error {
	if session.Status != domain.SessionStatusActive {
		return domain.ErrSessionInactive
	}

	if time.Now().After(session.ExpiresAt) {
		return domain.ErrSessionExpired
	}

	return nil
}

func (am *AuthMiddleware) getUserInfo(ctx context.Context, userID string) (interface{}, error) {
	query := queries.GetUserByIDQuery{UserID: userID}
	result, err := am.queryBus.Handle(ctx, query)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (am *AuthMiddleware) updateSessionActivity(ctx context.Context, session *domain.Session) {
	now := time.Now()
	session.LastUsedAt = &now
	session.UpdatedAt = now

	// Update session in cache (fire and forget)
	go func() {
		if err := am.sessionCache.UpdateSession(context.Background(), session); err != nil {
			am.logger.ErrorWithFields("Failed to update session activity",
				logger.Error(err),
				logger.String("session_id", session.ID))
		}
	}()
}

func (am *AuthMiddleware) addAuthContext(ctx context.Context, claims *domain.TokenClaims, session *domain.Session, user interface{}) context.Context {
	ctx = context.WithValue(ctx, "user_id", claims.UserID)
	ctx = context.WithValue(ctx, "session_id", session.ID)
	ctx = context.WithValue(ctx, "user_role", claims.Role)
	ctx = context.WithValue(ctx, "user_email", claims.Email)
	ctx = context.WithValue(ctx, "claims", claims)
	ctx = context.WithValue(ctx, "session", session)
	ctx = context.WithValue(ctx, "user", user)
	return ctx
}

func (am *AuthMiddleware) shouldSkipAuth(path string) bool {
	for _, skipPath := range am.config.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

func (am *AuthMiddleware) hasRequiredRole(userRole, requiredRole domain.UserRole) bool {
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

func (am *AuthMiddleware) checkUserPermission(ctx context.Context, userID, permission string) (bool, error) {
	// This would typically check user permissions from database or cache
	// For now, return true as placeholder
	return true, nil
}

func (am *AuthMiddleware) getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

func (am *AuthMiddleware) getUserRoleFromContext(ctx context.Context) string {
	if role, ok := ctx.Value("user_role").(domain.UserRole); ok {
		return string(role)
	}
	return ""
}

func (am *AuthMiddleware) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error": "` + message + `"}`))
}
