package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/cache"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/events"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/security"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// SessionManager handles real-time session management
type SessionManager struct {
	container  infrastructure.ContainerInterface
	cache      cache.Cache
	jwtService security.JWTService
	logger     *logger.Logger
	config     *SessionConfig
}

// SessionConfig represents session configuration
type SessionConfig struct {
	CookieName     string        `yaml:"cookie_name"`
	CookieDomain   string        `yaml:"cookie_domain"`
	CookiePath     string        `yaml:"cookie_path"`
	CookieSecure   bool          `yaml:"cookie_secure"`
	CookieHTTPOnly bool          `yaml:"cookie_http_only"`
	CookieSameSite http.SameSite `yaml:"cookie_same_site"`
	MaxAge         time.Duration `yaml:"max_age"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"`
}

// SessionInfo represents session information
type SessionInfo struct {
	SessionID    string                 `json:"session_id"`
	UserID       string                 `json:"user_id"`
	Email        string                 `json:"email"`
	Role         string                 `json:"role"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	CreatedAt    time.Time              `json:"created_at"`
	LastActivity time.Time              `json:"last_activity"`
	ExpiresAt    time.Time              `json:"expires_at"`
	IsActive     bool                   `json:"is_active"`
	DeviceInfo   map[string]interface{} `json:"device_info"`
	Location     map[string]interface{} `json:"location"`
}

// ActiveSession represents an active session for real-time tracking
type ActiveSession struct {
	SessionInfo
	TokenID       string    `json:"token_id"`
	RefreshToken  string    `json:"refresh_token,omitempty"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

// NewSessionManager creates a new session manager
func NewSessionManager(container infrastructure.ContainerInterface, logger *logger.Logger) *SessionManager {
	config := &SessionConfig{
		CookieName:     "auth_session",
		CookiePath:     "/",
		CookieSecure:   false, // Set to true in production with HTTPS
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteLaxMode,
		MaxAge:         24 * time.Hour,
		IdleTimeout:    30 * time.Minute,
	}

	return &SessionManager{
		container:  container,
		cache:      container.GetCache(),
		jwtService: container.GetJWTService(),
		logger:     logger,
		config:     config,
	}
}

// CreateSession creates a new session
func (sm *SessionManager) CreateSession(ctx context.Context, userID, email, role, tokenID string, r *http.Request) (*ActiveSession, error) {
	sessionID := fmt.Sprintf("session_%s_%d", userID, time.Now().UnixNano())

	now := time.Now()
	session := &ActiveSession{
		SessionInfo: SessionInfo{
			SessionID:    sessionID,
			UserID:       userID,
			Email:        email,
			Role:         role,
			IPAddress:    getClientIP(r),
			UserAgent:    r.UserAgent(),
			CreatedAt:    now,
			LastActivity: now,
			ExpiresAt:    now.Add(sm.config.MaxAge),
			IsActive:     true,
			DeviceInfo:   sm.extractDeviceInfo(r),
			Location:     sm.extractLocationInfo(r),
		},
		TokenID:       tokenID,
		LastHeartbeat: now,
	}

	// Store session in cache
	if err := sm.storeSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	// Add to user's active sessions
	if err := sm.addToUserSessions(ctx, userID, sessionID); err != nil {
		sm.logger.WithError(err).Error("Failed to add session to user sessions")
	}

	// Publish session created event
	sm.publishSessionEvent(ctx, "session.created", session)

	sm.logger.InfoWithFields("Session created",
		logger.String("session_id", sessionID),
		logger.String("user_id", userID),
		logger.String("ip_address", session.IPAddress))

	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*ActiveSession, error) {
	if sm.cache == nil {
		return nil, fmt.Errorf("cache not available")
	}

	key := sm.getSessionKey(sessionID)
	var session ActiveSession

	if err := sm.cache.Get(ctx, key, &session); err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		sm.InvalidateSession(ctx, sessionID)
		return nil, fmt.Errorf("session expired")
	}

	// Check idle timeout
	if time.Since(session.LastActivity) > sm.config.IdleTimeout {
		sm.InvalidateSession(ctx, sessionID)
		return nil, fmt.Errorf("session idle timeout")
	}

	return &session, nil
}

// UpdateSessionActivity updates the last activity time
func (sm *SessionManager) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.LastActivity = time.Now()
	session.LastHeartbeat = time.Now()

	return sm.storeSession(ctx, session)
}

// InvalidateSession invalidates a session
func (sm *SessionManager) InvalidateSession(ctx context.Context, sessionID string) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		// Session might already be invalid, that's okay
		return nil
	}

	// Mark as inactive
	session.IsActive = false

	// Store updated session (for audit purposes)
	if err := sm.storeSession(ctx, session); err != nil {
		sm.logger.WithError(err).Error("Failed to update session status")
	}

	// Remove from cache after a delay (for audit trail)
	go func() {
		time.Sleep(5 * time.Minute)
		key := sm.getSessionKey(sessionID)
		if sm.cache != nil {
			sm.cache.Delete(context.Background(), key)
		}
	}()

	// Remove from user's active sessions
	if err := sm.removeFromUserSessions(ctx, session.UserID, sessionID); err != nil {
		sm.logger.WithError(err).Error("Failed to remove session from user sessions")
	}

	// Revoke JWT token
	if sm.jwtService != nil && session.TokenID != "" {
		if err := sm.jwtService.RevokeToken(ctx, session.TokenID); err != nil {
			sm.logger.WithError(err).Error("Failed to revoke JWT token")
		}
	}

	// Publish session invalidated event
	sm.publishSessionEvent(ctx, "session.invalidated", session)

	sm.logger.InfoWithFields("Session invalidated",
		logger.String("session_id", sessionID),
		logger.String("user_id", session.UserID))

	return nil
}

// GetUserSessions returns all active sessions for a user
func (sm *SessionManager) GetUserSessions(ctx context.Context, userID string) ([]*ActiveSession, error) {
	if sm.cache == nil {
		return nil, fmt.Errorf("cache not available")
	}

	// Get session IDs for user
	key := sm.getUserSessionsKey(userID)
	var sessionIDs []string

	if err := sm.cache.Get(ctx, key, &sessionIDs); err != nil {
		return []*ActiveSession{}, nil // No sessions found
	}

	sessions := make([]*ActiveSession, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		session, err := sm.GetSession(ctx, sessionID)
		if err != nil {
			// Session might be expired, skip it
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// InvalidateUserSessions invalidates all sessions for a user
func (sm *SessionManager) InvalidateUserSessions(ctx context.Context, userID string, exceptSessionID string) error {
	sessions, err := sm.GetUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	for _, session := range sessions {
		if session.SessionID != exceptSessionID {
			if err := sm.InvalidateSession(ctx, session.SessionID); err != nil {
				sm.logger.WithError(err).Error("Failed to invalidate user session")
			}
		}
	}

	return nil
}

// SetSessionCookie sets the session cookie
func (sm *SessionManager) SetSessionCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     sm.config.CookieName,
		Value:    token,
		Path:     sm.config.CookiePath,
		Domain:   sm.config.CookieDomain,
		MaxAge:   int(sm.config.MaxAge.Seconds()),
		Secure:   sm.config.CookieSecure,
		HttpOnly: sm.config.CookieHTTPOnly,
		SameSite: sm.config.CookieSameSite,
	}

	http.SetCookie(w, cookie)
}

// ClearSessionCookie clears the session cookie
func (sm *SessionManager) ClearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     sm.config.CookieName,
		Value:    "",
		Path:     sm.config.CookiePath,
		Domain:   sm.config.CookieDomain,
		MaxAge:   -1,
		Secure:   sm.config.CookieSecure,
		HttpOnly: sm.config.CookieHTTPOnly,
		SameSite: sm.config.CookieSameSite,
	}

	http.SetCookie(w, cookie)
}

// GetSessionFromCookie gets session from cookie
func (sm *SessionManager) GetSessionFromCookie(r *http.Request) (*ActiveSession, error) {
	cookie, err := r.Cookie(sm.config.CookieName)
	if err != nil {
		return nil, fmt.Errorf("session cookie not found")
	}

	// Validate token and get claims
	claims, err := sm.jwtService.ValidateAccessToken(r.Context(), cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid session token: %w", err)
	}

	// Get session by token ID
	sessions, err := sm.GetUserSessions(r.Context(), claims.UserID)
	if err != nil {
		return nil, err
	}

	for _, session := range sessions {
		if session.TokenID == claims.TokenID {
			return session, nil
		}
	}

	return nil, fmt.Errorf("session not found")
}

// Helper methods
func (sm *SessionManager) storeSession(ctx context.Context, session *ActiveSession) error {
	if sm.cache == nil {
		return fmt.Errorf("cache not available")
	}

	key := sm.getSessionKey(session.SessionID)
	return sm.cache.Set(ctx, key, session, sm.config.MaxAge)
}

func (sm *SessionManager) addToUserSessions(ctx context.Context, userID, sessionID string) error {
	if sm.cache == nil {
		return fmt.Errorf("cache not available")
	}

	key := sm.getUserSessionsKey(userID)
	var sessionIDs []string

	// Get existing sessions
	sm.cache.Get(ctx, key, &sessionIDs)

	// Add new session
	sessionIDs = append(sessionIDs, sessionID)

	return sm.cache.Set(ctx, key, sessionIDs, sm.config.MaxAge)
}

func (sm *SessionManager) removeFromUserSessions(ctx context.Context, userID, sessionID string) error {
	if sm.cache == nil {
		return fmt.Errorf("cache not available")
	}

	key := sm.getUserSessionsKey(userID)
	var sessionIDs []string

	if err := sm.cache.Get(ctx, key, &sessionIDs); err != nil {
		return nil // No sessions to remove
	}

	// Remove session ID
	filtered := make([]string, 0, len(sessionIDs))
	for _, id := range sessionIDs {
		if id != sessionID {
			filtered = append(filtered, id)
		}
	}

	return sm.cache.Set(ctx, key, filtered, sm.config.MaxAge)
}

func (sm *SessionManager) getSessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

func (sm *SessionManager) getUserSessionsKey(userID string) string {
	return fmt.Sprintf("user_sessions:%s", userID)
}

func (sm *SessionManager) extractDeviceInfo(r *http.Request) map[string]interface{} {
	return map[string]interface{}{
		"user_agent": r.UserAgent(),
		"platform":   "web", // Could be enhanced with device detection
	}
}

func (sm *SessionManager) extractLocationInfo(r *http.Request) map[string]interface{} {
	// This could be enhanced with GeoIP lookup
	return map[string]interface{}{
		"ip_address": getClientIP(r),
	}
}

func (sm *SessionManager) publishSessionEvent(ctx context.Context, eventType string, session *ActiveSession) {
	eventPublisher := sm.container.GetEventPublisher()
	if eventPublisher == nil {
		return
	}

	eventData, _ := json.Marshal(session)
	var data map[string]interface{}
	json.Unmarshal(eventData, &data)

	event := &events.Event{
		ID:            fmt.Sprintf("session_event_%d", time.Now().UnixNano()),
		Type:          eventType,
		Source:        "auth-service",
		AggregateID:   session.SessionID,
		AggregateType: "session",
		Data:          data,
		Timestamp:     time.Now(),
	}

	// Publish asynchronously
	go func() {
		if err := eventPublisher.Publish(context.Background(), event); err != nil {
			sm.logger.WithError(err).Error("Failed to publish session event")
		}
	}()
}
