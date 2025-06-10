package session

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/cache"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/events"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// SessionStatus represents the status of a session
type SessionStatus string

const (
	SessionStatusActive   SessionStatus = "active"
	SessionStatusExpired  SessionStatus = "expired"
	SessionStatusRevoked  SessionStatus = "revoked"
	SessionStatusInactive SessionStatus = "inactive"
)

// Session represents a user session
type Session struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	Email        string                 `json:"email"`
	Role         string                 `json:"role"`
	Status       SessionStatus          `json:"status"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	DeviceInfo   *DeviceInfo            `json:"device_info,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	LastActivity time.Time              `json:"last_activity"`
	ExpiresAt    time.Time              `json:"expires_at"`
}

// IsExpired checks if the session is expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsIdleExpired checks if the session has been idle too long
func (s *Session) IsIdleExpired(idleTimeout time.Duration) bool {
	return time.Now().Sub(s.LastActivity) > idleTimeout
}

// IsValid checks if the session is valid and active
func (s *Session) IsValid() bool {
	return s.Status == SessionStatusActive && !s.IsExpired()
}

// Revoke marks the session as revoked
func (s *Session) Revoke() {
	s.Status = SessionStatusRevoked
	s.UpdatedAt = time.Now()
}

// UpdateActivity updates the last activity timestamp
func (s *Session) UpdateActivity() {
	s.LastActivity = time.Now()
	s.UpdatedAt = time.Now()
}

// DeviceInfo contains information about the device used for the session
type DeviceInfo struct {
	DeviceID   string `json:"device_id,omitempty"`
	DeviceType string `json:"device_type,omitempty"` // mobile, desktop, tablet
	OS         string `json:"os,omitempty"`
	Browser    string `json:"browser,omitempty"`
	AppVersion string `json:"app_version,omitempty"`
}

// SessionConfig defines session configuration
type SessionConfig struct {
	CookieName      string        `yaml:"cookie_name"`
	CookiePath      string        `yaml:"cookie_path"`
	CookieDomain    string        `yaml:"cookie_domain"`
	CookieSecure    bool          `yaml:"cookie_secure"`
	CookieHTTPOnly  bool          `yaml:"cookie_http_only"`
	CookieSameSite  http.SameSite `yaml:"cookie_same_site"`
	MaxAge          time.Duration `yaml:"max_age"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
	EnableEvents    bool          `yaml:"enable_events"`
}

// DefaultSessionConfig returns default session configuration
func DefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		CookieName:      "session_id",
		CookiePath:      "/",
		CookieDomain:    "",
		CookieSecure:    false, // Set to true in production with HTTPS
		CookieHTTPOnly:  true,
		CookieSameSite:  http.SameSiteLaxMode,
		MaxAge:          24 * time.Hour,
		IdleTimeout:     30 * time.Minute,
		CleanupInterval: 1 * time.Hour,
		EnableEvents:    true,
	}
}

// Manager handles session management operations
type Manager struct {
	cache          cache.Cache
	eventPublisher events.EventPublisher
	logger         *logger.Logger
	config         *SessionConfig
	cleanupTicker  *time.Ticker
	cleanupStop    chan struct{}
}

// NewManager creates a new session manager
func NewManager(cache cache.Cache, eventPublisher events.EventPublisher, logger *logger.Logger, config *SessionConfig) *Manager {
	if config == nil {
		config = DefaultSessionConfig()
	}

	manager := &Manager{
		cache:          cache,
		eventPublisher: eventPublisher,
		logger:         logger,
		config:         config,
		cleanupStop:    make(chan struct{}),
	}

	// Start cleanup routine
	manager.startCleanupRoutine()

	return manager
}

// CreateSession creates a new session
func (m *Manager) CreateSession(ctx context.Context, userID, email, role string, r *http.Request) (*Session, error) {
	sessionID := uuid.New().String()
	now := time.Now()

	session := &Session{
		ID:           sessionID,
		UserID:       userID,
		Email:        email,
		Role:         role,
		Status:       SessionStatusActive,
		IPAddress:    getClientIP(r),
		UserAgent:    r.UserAgent(),
		DeviceInfo:   extractDeviceInfo(r),
		Metadata:     make(map[string]interface{}),
		CreatedAt:    now,
		UpdatedAt:    now,
		LastActivity: now,
		ExpiresAt:    now.Add(m.config.MaxAge),
	}

	// Store session in cache
	sessionKey := m.getSessionKey(sessionID)
	if err := m.cache.Set(ctx, sessionKey, session, m.config.MaxAge); err != nil {
		m.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to store session")
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	// Store user session mapping
	userSessionKey := m.getUserSessionKey(userID)
	if err := m.cache.Set(ctx, userSessionKey, sessionID, m.config.MaxAge); err != nil {
		m.logger.WithError(err).WithField("user_id", userID).Warn("Failed to store user session mapping")
	}

	// Publish session created event
	if m.config.EnableEvents && m.eventPublisher != nil {
		m.publishSessionEvent(ctx, "session.created", session)
	}

	m.logger.WithFields(map[string]interface{}{
		"session_id": sessionID,
		"user_id":    userID,
		"ip_address": session.IPAddress,
	}).Info("Session created")

	return session, nil
}

// GetSession retrieves a session by ID
func (m *Manager) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	sessionKey := m.getSessionKey(sessionID)

	var session Session
	if err := m.cache.Get(ctx, sessionKey, &session); err != nil {
		if err == cache.ErrCacheKeyNotFound {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Check if session is expired
	if session.IsExpired() {
		m.logger.WithField("session_id", sessionID).Debug("Session expired")
		return nil, ErrSessionExpired
	}

	// Check if session is idle too long
	if session.IsIdleExpired(m.config.IdleTimeout) {
		m.logger.WithField("session_id", sessionID).Debug("Session idle expired")
		return nil, ErrSessionExpired
	}

	return &session, nil
}

// UpdateSession updates session activity and metadata
func (m *Manager) UpdateSession(ctx context.Context, sessionID string, metadata map[string]interface{}) error {
	session, err := m.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	now := time.Now()
	session.LastActivity = now
	session.UpdatedAt = now

	// Update metadata if provided
	if metadata != nil {
		for key, value := range metadata {
			session.Metadata[key] = value
		}
	}

	// Store updated session
	sessionKey := m.getSessionKey(sessionID)
	if err := m.cache.Set(ctx, sessionKey, session, m.config.MaxAge); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// RevokeSession revokes a session
func (m *Manager) RevokeSession(ctx context.Context, sessionID string) error {
	session, err := m.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.Status = SessionStatusRevoked
	session.UpdatedAt = time.Now()

	// Store revoked session (with shorter TTL)
	sessionKey := m.getSessionKey(sessionID)
	if err := m.cache.Set(ctx, sessionKey, session, time.Hour); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	// Remove user session mapping
	userSessionKey := m.getUserSessionKey(session.UserID)
	m.cache.Delete(ctx, userSessionKey)

	// Publish session revoked event
	if m.config.EnableEvents && m.eventPublisher != nil {
		m.publishSessionEvent(ctx, "session.revoked", session)
	}

	m.logger.WithFields(map[string]interface{}{
		"session_id": sessionID,
		"user_id":    session.UserID,
	}).Info("Session revoked")

	return nil
}

// RevokeUserSessions revokes all sessions for a user
func (m *Manager) RevokeUserSessions(ctx context.Context, userID string) error {
	userSessionKey := m.getUserSessionKey(userID)

	var sessionID string
	if err := m.cache.Get(ctx, userSessionKey, &sessionID); err != nil {
		if err == cache.ErrCacheKeyNotFound {
			return nil // No active sessions
		}
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	// Revoke the session
	if err := m.RevokeSession(ctx, sessionID); err != nil {
		m.logger.WithError(err).WithField("session_id", sessionID).Warn("Failed to revoke session")
	}

	return nil
}

// ValidateSession validates a session and updates activity
func (m *Manager) ValidateSession(ctx context.Context, sessionID string) (*Session, error) {
	session, err := m.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.Status != SessionStatusActive {
		return nil, ErrSessionInactive
	}

	// Update last activity
	if err := m.UpdateSession(ctx, sessionID, nil); err != nil {
		m.logger.WithError(err).WithField("session_id", sessionID).Warn("Failed to update session activity")
	}

	return session, nil
}

// SetSessionCookie sets the session cookie in the response
func (m *Manager) SetSessionCookie(w http.ResponseWriter, sessionID string) {
	cookie := &http.Cookie{
		Name:     m.config.CookieName,
		Value:    sessionID,
		Path:     m.config.CookiePath,
		Domain:   m.config.CookieDomain,
		MaxAge:   int(m.config.MaxAge.Seconds()),
		Secure:   m.config.CookieSecure,
		HttpOnly: m.config.CookieHTTPOnly,
		SameSite: m.config.CookieSameSite,
	}
	http.SetCookie(w, cookie)
}

// ClearSessionCookie clears the session cookie
func (m *Manager) ClearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     m.config.CookieName,
		Value:    "",
		Path:     m.config.CookiePath,
		Domain:   m.config.CookieDomain,
		MaxAge:   -1,
		Secure:   m.config.CookieSecure,
		HttpOnly: m.config.CookieHTTPOnly,
		SameSite: m.config.CookieSameSite,
	}
	http.SetCookie(w, cookie)
}

// GetSessionFromRequest extracts session ID from request
func (m *Manager) GetSessionFromRequest(r *http.Request) string {
	// Try cookie first
	if cookie, err := r.Cookie(m.config.CookieName); err == nil {
		return cookie.Value
	}

	// Try Authorization header
	if auth := r.Header.Get("Authorization"); auth != "" {
		if len(auth) > 7 && auth[:7] == "Bearer " {
			return auth[7:]
		}
	}

	// Try X-Session-ID header
	return r.Header.Get("X-Session-ID")
}

// Shutdown stops the session manager
func (m *Manager) Shutdown() {
	if m.cleanupTicker != nil {
		m.cleanupTicker.Stop()
	}
	close(m.cleanupStop)
}

// Helper methods

// getSessionKey returns the cache key for a session
func (m *Manager) getSessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

// getUserSessionKey returns the cache key for user sessions
func (m *Manager) getUserSessionKey(userID string) string {
	return fmt.Sprintf("user_session:%s", userID)
}

// startCleanupRoutine starts the session cleanup routine
func (m *Manager) startCleanupRoutine() {
	if m.config.CleanupInterval <= 0 {
		return
	}

	m.cleanupTicker = time.NewTicker(m.config.CleanupInterval)
	go func() {
		for {
			select {
			case <-m.cleanupTicker.C:
				m.cleanupExpiredSessions()
			case <-m.cleanupStop:
				return
			}
		}
	}()
}

// cleanupExpiredSessions removes expired sessions (placeholder)
func (m *Manager) cleanupExpiredSessions() {
	// This would typically scan for expired sessions and remove them
	// For now, we rely on Redis TTL for automatic cleanup
	m.logger.Debug("Session cleanup routine executed")
}

// publishSessionEvent publishes a session event
func (m *Manager) publishSessionEvent(ctx context.Context, eventType string, session *Session) {
	event := &events.Event{
		ID:            fmt.Sprintf("session_%s_%d", session.ID, time.Now().UnixNano()),
		Type:          eventType,
		Source:        "session-manager",
		AggregateID:   session.UserID,
		AggregateType: "user",
		Version:       1,
		Data: map[string]interface{}{
			"session_id": session.ID,
			"user_id":    session.UserID,
			"email":      session.Email,
			"role":       session.Role,
			"ip_address": session.IPAddress,
		},
		Metadata: map[string]interface{}{
			"source":     "session-manager",
			"user_agent": session.UserAgent,
		},
		Timestamp: time.Now(),
	}

	if err := m.eventPublisher.Publish(ctx, event); err != nil {
		m.logger.WithError(err).WithField("event_type", eventType).Warn("Failed to publish session event")
	}
}

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// extractDeviceInfo extracts device information from request
func extractDeviceInfo(r *http.Request) *DeviceInfo {
	userAgent := r.UserAgent()
	if userAgent == "" {
		return nil
	}

	// Basic device info extraction (can be enhanced with a proper user agent parser)
	deviceInfo := &DeviceInfo{
		DeviceID:   r.Header.Get("X-Device-ID"),
		DeviceType: "unknown",
		OS:         "unknown",
		Browser:    "unknown",
		AppVersion: r.Header.Get("X-App-Version"),
	}

	// Simple user agent parsing (in production, use a proper library)
	if userAgent != "" {
		deviceInfo.Browser = userAgent
		// Add more sophisticated parsing here
	}

	return deviceInfo
}
