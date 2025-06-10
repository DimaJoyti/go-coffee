package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/session"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// SessionMiddleware provides session management middleware
func (m *Middleware) SessionMiddleware(sessionManager *session.Manager, config *SessionConfig) func(http.HandlerFunc) http.HandlerFunc {
	if config == nil {
		config = DefaultSessionConfig()
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Extract session ID from request
			sessionID := sessionManager.GetSessionFromRequest(r)
			
			if sessionID != "" {
				// Validate existing session
				sess, err := sessionManager.ValidateSession(ctx, sessionID)
				if err != nil {
					if err == session.ErrSessionExpired || err == session.ErrSessionNotFound {
						// Clear invalid session cookie
						sessionManager.ClearSessionCookie(w)
						m.logger.WithFields(map[string]interface{}{
							"session_id": sessionID,
							"error":      err.Error(),
						}).Debug("Session validation failed")
					} else {
						m.logger.WithError(err).WithField("session_id", sessionID).Warn("Session validation error")
					}
				} else {
					// Add session to request context
					ctx = context.WithValue(ctx, "session", sess)
					ctx = context.WithValue(ctx, "session_id", sess.ID)
					ctx = context.WithValue(ctx, "user_id", sess.UserID)
					ctx = context.WithValue(ctx, "user_email", sess.Email)
					ctx = context.WithValue(ctx, "user_role", sess.Role)
					
					m.logger.WithFields(map[string]interface{}{
						"session_id": sess.ID,
						"user_id":    sess.UserID,
						"user_role":  sess.Role,
					}).Debug("Session validated successfully")
				}
			}

			// Continue with updated context
			next(w, r.WithContext(ctx))
		}
	}
}

// RequireSessionMiddleware ensures a valid session exists
func (m *Middleware) RequireSessionMiddleware(sessionManager *session.Manager, config *SessionConfig) func(http.HandlerFunc) http.HandlerFunc {
	if config == nil {
		config = DefaultSessionConfig()
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Check if session exists in context (should be set by SessionMiddleware)
			sess := GetSessionFromContext(r.Context())
			if sess == nil {
				m.respondWithAuthError(w, "Session required", "Valid session is required to access this resource")
				return
			}

			// Check if session is still valid
			if !sess.IsValid() {
				sessionManager.ClearSessionCookie(w)
				m.respondWithAuthError(w, "Session expired", "Your session has expired, please log in again")
				return
			}

			next(w, r)
		}
	}
}

// RequireRoleMiddleware ensures the user has the required role
func (m *Middleware) RequireRoleMiddleware(requiredRoles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			sess := GetSessionFromContext(r.Context())
			if sess == nil {
				m.respondWithAuthError(w, "Session required", "Valid session is required")
				return
			}

			// Check if user has any of the required roles
			hasRole := false
			for _, role := range requiredRoles {
				if sess.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				m.logger.WithFields(map[string]interface{}{
					"user_id":        sess.UserID,
					"user_role":      sess.Role,
					"required_roles": requiredRoles,
				}).Warn("Access denied - insufficient role")
				
				m.respondWithAuthError(w, "Access denied", "Insufficient permissions to access this resource")
				return
			}

			next(w, r)
		}
	}
}

// SessionConfig defines session middleware configuration
type SessionConfig struct {
	Enabled         bool          `yaml:"enabled"`
	RequireSession  bool          `yaml:"require_session"`
	ExcludedPaths   []string      `yaml:"excluded_paths"`
	SessionTimeout  time.Duration `yaml:"session_timeout"`
	RefreshInterval time.Duration `yaml:"refresh_interval"`
}

// DefaultSessionConfig returns default session middleware configuration
func DefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		Enabled:        true,
		RequireSession: false, // Set to true to require sessions for all routes
		ExcludedPaths: []string{
			"/health",
			"/metrics",
			"/api/docs",
			"/api/v1/auth/login",
			"/api/v1/auth/register",
			"/static/",
		},
		SessionTimeout:  30 * time.Minute,
		RefreshInterval: 5 * time.Minute,
	}
}

// Helper functions for session context management

// GetSessionFromContext extracts session from request context
func GetSessionFromContext(ctx context.Context) *session.Session {
	if sess := ctx.Value("session"); sess != nil {
		if session, ok := sess.(*session.Session); ok {
			return session
		}
	}
	return nil
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(ctx context.Context) string {
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// GetUserEmailFromContext extracts user email from request context
func GetUserEmailFromContext(ctx context.Context) string {
	if email := ctx.Value("user_email"); email != nil {
		if e, ok := email.(string); ok {
			return e
		}
	}
	return ""
}

// GetUserRoleFromContext extracts user role from request context
func GetUserRoleFromContext(ctx context.Context) string {
	if role := ctx.Value("user_role"); role != nil {
		if r, ok := role.(string); ok {
			return r
		}
	}
	return ""
}

// GetSessionIDFromContext extracts session ID from request context
func GetSessionIDFromContext(ctx context.Context) string {
	if sessionID := ctx.Value("session_id"); sessionID != nil {
		if id, ok := sessionID.(string); ok {
			return id
		}
	}
	return ""
}

// SessionLoginHelper helps create sessions during login
type SessionLoginHelper struct {
	sessionManager *session.Manager
	logger         *logger.Logger
}

// NewSessionLoginHelper creates a new session login helper
func NewSessionLoginHelper(sessionManager *session.Manager, logger *logger.Logger) *SessionLoginHelper {
	return &SessionLoginHelper{
		sessionManager: sessionManager,
		logger:         logger,
	}
}

// CreateSessionForUser creates a new session for a user and sets the cookie
func (h *SessionLoginHelper) CreateSessionForUser(ctx context.Context, w http.ResponseWriter, r *http.Request, userID, email, role string) (*session.Session, error) {
	// Create new session
	sess, err := h.sessionManager.CreateSession(ctx, userID, email, role, r)
	if err != nil {
		h.logger.WithError(err).WithFields(map[string]interface{}{
			"user_id": userID,
			"email":   email,
		}).Error("Failed to create session")
		return nil, err
	}

	// Set session cookie
	h.sessionManager.SetSessionCookie(w, sess.ID)

	h.logger.WithFields(map[string]interface{}{
		"session_id": sess.ID,
		"user_id":    userID,
		"email":      email,
		"role":       role,
	}).Info("Session created for user")

	return sess, nil
}

// LogoutUser revokes the user's session and clears the cookie
func (h *SessionLoginHelper) LogoutUser(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	sessionID := h.sessionManager.GetSessionFromRequest(r)
	if sessionID == "" {
		return nil // No session to logout
	}

	// Revoke session
	if err := h.sessionManager.RevokeSession(ctx, sessionID); err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Warn("Failed to revoke session")
	}

	// Clear session cookie
	h.sessionManager.ClearSessionCookie(w)

	h.logger.WithField("session_id", sessionID).Info("User logged out")
	return nil
}

// LogoutAllUserSessions revokes all sessions for a user
func (h *SessionLoginHelper) LogoutAllUserSessions(ctx context.Context, userID string) error {
	if err := h.sessionManager.RevokeUserSessions(ctx, userID); err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to revoke all user sessions")
		return err
	}

	h.logger.WithField("user_id", userID).Info("All user sessions revoked")
	return nil
}
