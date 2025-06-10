package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
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
	AggregateRoot // Embed aggregate root for event functionality

	ID               string            `json:"id"`
	UserID           string            `json:"user_id"`
	AccessToken      string            `json:"access_token"`
	RefreshToken     string            `json:"refresh_token"`
	TokenHash        string            `json:"token_hash,omitempty"`
	RefreshTokenHash string            `json:"refresh_token_hash,omitempty"`
	Status           SessionStatus     `json:"status"`
	IsActive         bool              `json:"is_active"`
	ExpiresAt        time.Time         `json:"expires_at"`
	RefreshExpiresAt time.Time         `json:"refresh_expires_at"`
	DeviceInfo       *DeviceInfo       `json:"device_info,omitempty"`
	IPAddress        string            `json:"ip_address,omitempty"`
	UserAgent        string            `json:"user_agent,omitempty"`
	LastUsedAt       *time.Time        `json:"last_used_at,omitempty"`
	LastActivity     *time.Time        `json:"last_activity,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// DeviceInfo contains information about the device used for the session
type DeviceInfo struct {
	DeviceID   string `json:"device_id,omitempty"`
	DeviceType string `json:"device_type,omitempty"` // mobile, desktop, tablet, etc.
	OS         string `json:"os,omitempty"`
	Browser    string `json:"browser,omitempty"`
	AppVersion string `json:"app_version,omitempty"`
}

// Session validation errors
var (
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionExpired      = errors.New("session expired")
	ErrSessionRevoked      = errors.New("session revoked")
	ErrSessionInactive     = errors.New("session inactive")
	ErrInvalidToken        = errors.New("invalid token")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
)

// NewSession creates a new session
func NewSession(userID, accessToken, refreshToken string, accessTTL, refreshTTL time.Duration) *Session {
	now := time.Now()
	session := &Session{
		ID:               uuid.New().String(),
		UserID:           userID,
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		Status:           SessionStatusActive,
		IsActive:         true,
		ExpiresAt:        now.Add(accessTTL),
		RefreshExpiresAt: now.Add(refreshTTL),
		Metadata:         make(map[string]string),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Generate session created event
	event := NewDomainEvent(EventTypeSessionCreated, session.ID, map[string]interface{}{
		"session_id": session.ID,
		"user_id":    session.UserID,
		"expires_at": session.ExpiresAt,
		"timestamp":  now,
	})
	session.AddEvent(*event)

	return session
}

// IsValid checks if the session is valid and not expired
func (s *Session) IsValid() bool {
	now := time.Now()
	return s.Status == SessionStatusActive && now.Before(s.ExpiresAt)
}

// IsRefreshValid checks if the refresh token is valid and not expired
func (s *Session) IsRefreshValid() bool {
	now := time.Now()
	return s.Status == SessionStatusActive && now.Before(s.RefreshExpiresAt)
}

// IsExpired checks if the session is expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsRefreshExpired checks if the refresh token is expired
func (s *Session) IsRefreshExpired() bool {
	return time.Now().After(s.RefreshExpiresAt)
}

// Revoke revokes the session
func (s *Session) Revoke() {
	s.Status = SessionStatusRevoked
	s.UpdatedAt = time.Now()

	// Generate session revoked event
	event := NewDomainEvent(EventTypeSessionRevoked, s.ID, map[string]interface{}{
		"session_id": s.ID,
		"user_id":    s.UserID,
		"timestamp":  s.UpdatedAt,
	})
	s.AddEvent(*event)
}

// Expire expires the session
func (s *Session) Expire() {
	s.Status = SessionStatusExpired
	s.UpdatedAt = time.Now()

	// Generate session expired event
	event := NewDomainEvent(EventTypeSessionExpired, s.ID, map[string]interface{}{
		"session_id": s.ID,
		"user_id":    s.UserID,
		"timestamp":  s.UpdatedAt,
	})
	s.AddEvent(*event)
}

// UpdateTokens updates the access and refresh tokens
func (s *Session) UpdateTokens(accessToken, refreshToken string, accessTTL, refreshTTL time.Duration) {
	now := time.Now()
	s.AccessToken = accessToken
	s.RefreshToken = refreshToken
	s.ExpiresAt = now.Add(accessTTL)
	s.RefreshExpiresAt = now.Add(refreshTTL)
	s.UpdatedAt = now

	// Generate session refreshed event
	event := NewDomainEvent(EventTypeSessionRefreshed, s.ID, map[string]interface{}{
		"session_id": s.ID,
		"user_id":    s.UserID,
		"expires_at": s.ExpiresAt,
		"timestamp":  now,
	})
	s.AddEvent(*event)
}

// UpdateLastUsed updates the last used timestamp
func (s *Session) UpdateLastUsed() {
	now := time.Now()
	s.LastUsedAt = &now
	s.UpdatedAt = now
}

// SetDeviceInfo sets the device information
func (s *Session) SetDeviceInfo(deviceInfo *DeviceInfo) {
	s.DeviceInfo = deviceInfo
	s.UpdatedAt = time.Now()
}

// SetIPAddress sets the IP address
func (s *Session) SetIPAddress(ipAddress string) {
	s.IPAddress = ipAddress
	s.UpdatedAt = time.Now()
}

// SetUserAgent sets the user agent
func (s *Session) SetUserAgent(userAgent string) {
	s.UserAgent = userAgent
	s.UpdatedAt = time.Now()
}

// Validate validates the session and returns appropriate error
func (s *Session) Validate() error {
	switch s.Status {
	case SessionStatusRevoked:
		return ErrSessionRevoked
	case SessionStatusExpired:
		return ErrSessionExpired
	case SessionStatusInactive:
		return ErrSessionInactive
	}

	if s.IsExpired() {
		return ErrSessionExpired
	}

	return nil
}

// ValidateRefresh validates the refresh token and returns appropriate error
func (s *Session) ValidateRefresh() error {
	if err := s.Validate(); err != nil {
		return err
	}

	if s.IsRefreshExpired() {
		return ErrRefreshTokenExpired
	}

	return nil
}
