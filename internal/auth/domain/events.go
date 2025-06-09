package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a domain event in the auth service
type DomainEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	AggregateID string                 `json:"aggregate_id"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
	Version     string                 `json:"version"`
}

// NewDomainEvent creates a new domain event
func NewDomainEvent(eventType, aggregateID string, data map[string]interface{}) *DomainEvent {
	return &DomainEvent{
		ID:          uuid.New().String(),
		Type:        eventType,
		AggregateID: aggregateID,
		Data:        data,
		Timestamp:   time.Now(),
		Version:     "1.0",
	}
}

// ToJSON converts the event to JSON
func (e *DomainEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON creates an event from JSON
func FromJSON(data []byte) (*DomainEvent, error) {
	var event DomainEvent
	err := json.Unmarshal(data, &event)
	return &event, err
}

// EventHandler defines the interface for handling domain events
type EventHandler interface {
	Handle(event *DomainEvent) error
}

// EventHandlerFunc is a function type that implements EventHandler
type EventHandlerFunc func(event *DomainEvent) error

// Handle implements the EventHandler interface
func (f EventHandlerFunc) Handle(event *DomainEvent) error {
	return f(event)
}

// AggregateRoot provides event functionality for domain aggregates
type AggregateRoot struct {
	events []DomainEvent
}

// AddEvent adds a domain event to the aggregate
func (ar *AggregateRoot) AddEvent(event DomainEvent) {
	ar.events = append(ar.events, event)
}

// GetEvents returns all uncommitted events
func (ar *AggregateRoot) GetEvents() []DomainEvent {
	return ar.events
}

// ClearEvents clears all uncommitted events
func (ar *AggregateRoot) ClearEvents() {
	ar.events = nil
}

// Auth Domain Event Types
const (
	// User Events
	EventTypeUserRegistered      = "auth.user.registered"
	EventTypeUserLoggedIn        = "auth.user.logged_in"
	EventTypeUserLoggedOut       = "auth.user.logged_out"
	EventTypeUserPasswordChanged = "auth.user.password_changed"
	EventTypeUserLocked          = "auth.user.locked"
	EventTypeUserUnlocked        = "auth.user.unlocked"
	EventTypeUserStatusChanged   = "auth.user.status_changed"
	EventTypeUserRoleChanged     = "auth.user.role_changed"
	EventTypeUserDeleted         = "auth.user.deleted"

	// MFA Events
	EventTypeMFAEnabled         = "auth.mfa.enabled"
	EventTypeMFADisabled        = "auth.mfa.disabled"
	EventTypeMFABackupCodeUsed  = "auth.mfa.backup_code_used"
	EventTypeMFAMethodChanged   = "auth.mfa.method_changed"

	// Security Events
	EventTypeFailedLogin        = "auth.security.failed_login"
	EventTypeSuccessfulLogin    = "auth.security.successful_login"
	EventTypeSuspiciousActivity = "auth.security.suspicious_activity"
	EventTypeDeviceAdded        = "auth.security.device_added"
	EventTypeDeviceTrusted      = "auth.security.device_trusted"
	EventTypeRiskScoreUpdated   = "auth.security.risk_score_updated"

	// Session Events
	EventTypeSessionCreated   = "auth.session.created"
	EventTypeSessionExpired   = "auth.session.expired"
	EventTypeSessionRevoked   = "auth.session.revoked"
	EventTypeSessionRefreshed = "auth.session.refreshed"

	// Token Events
	EventTypeTokenGenerated = "auth.token.generated"
	EventTypeTokenValidated = "auth.token.validated"
	EventTypeTokenRevoked   = "auth.token.revoked"
	EventTypeTokenExpired   = "auth.token.expired"

	// Account Events
	EventTypeAccountVerified   = "auth.account.verified"
	EventTypeAccountSuspended  = "auth.account.suspended"
	EventTypeAccountReactivated = "auth.account.reactivated"
	EventTypeEmailVerified     = "auth.account.email_verified"
	EventTypePhoneVerified     = "auth.account.phone_verified"
)

// Event Data Structures

// UserRegisteredEventData represents data for user registration event
type UserRegisteredEventData struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Role      UserRole  `json:"role"`
	Timestamp time.Time `json:"timestamp"`
}

// UserLoggedInEventData represents data for user login event
type UserLoggedInEventData struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	IPAddress   string `json:"ip_address"`
	UserAgent   string `json:"user_agent"`
	DeviceID    string `json:"device_id,omitempty"`
	SessionID   string `json:"session_id"`
	MFAUsed     bool   `json:"mfa_used"`
	Timestamp   time.Time `json:"timestamp"`
}

// UserLoggedOutEventData represents data for user logout event
type UserLoggedOutEventData struct {
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	Reason    string    `json:"reason,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// PasswordChangedEventData represents data for password change event
type PasswordChangedEventData struct {
	UserID    string    `json:"user_id"`
	Forced    bool      `json:"forced"`
	Timestamp time.Time `json:"timestamp"`
}

// UserLockedEventData represents data for user lock event
type UserLockedEventData struct {
	UserID      string    `json:"user_id"`
	Reason      string    `json:"reason"`
	LockedUntil time.Time `json:"locked_until"`
	Timestamp   time.Time `json:"timestamp"`
}

// MFAEnabledEventData represents data for MFA enabled event
type MFAEnabledEventData struct {
	UserID    string    `json:"user_id"`
	Method    MFAMethod `json:"method"`
	Timestamp time.Time `json:"timestamp"`
}

// SecurityEventData represents data for security events
type SecurityEventData struct {
	UserID      string            `json:"user_id"`
	EventType   string            `json:"event_type"`
	Severity    string            `json:"severity"`
	Description string            `json:"description"`
	IPAddress   string            `json:"ip_address,omitempty"`
	UserAgent   string            `json:"user_agent,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
}

// SessionEventData represents data for session events
type SessionEventData struct {
	SessionID string    `json:"session_id"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// TokenEventData represents data for token events
type TokenEventData struct {
	TokenID   string    `json:"token_id"`
	UserID    string    `json:"user_id"`
	TokenType string    `json:"token_type"`
	Action    string    `json:"action"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// DeviceEventData represents data for device events
type DeviceEventData struct {
	UserID      string    `json:"user_id"`
	DeviceID    string    `json:"device_id"`
	Fingerprint string    `json:"fingerprint"`
	UserAgent   string    `json:"user_agent"`
	IPAddress   string    `json:"ip_address"`
	Location    string    `json:"location,omitempty"`
	Trusted     bool      `json:"trusted"`
	Timestamp   time.Time `json:"timestamp"`
}

// RiskScoreEventData represents data for risk score events
type RiskScoreEventData struct {
	UserID       string        `json:"user_id"`
	OldScore     float64       `json:"old_score"`
	NewScore     float64       `json:"new_score"`
	OldLevel     SecurityLevel `json:"old_level"`
	NewLevel     SecurityLevel `json:"new_level"`
	Factors      []string      `json:"factors,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
}

// Helper functions to create specific events

// CreateUserRegisteredEvent creates a user registered event
func CreateUserRegisteredEvent(userID, email string, role UserRole) *DomainEvent {
	data := UserRegisteredEventData{
		UserID:    userID,
		Email:     email,
		Role:      role,
		Timestamp: time.Now(),
	}

	return NewDomainEvent(EventTypeUserRegistered, userID, map[string]interface{}{
		"user_id":   data.UserID,
		"email":     data.Email,
		"role":      data.Role,
		"timestamp": data.Timestamp,
	})
}

// CreateUserLoggedInEvent creates a user logged in event
func CreateUserLoggedInEvent(userID, email, ipAddress, userAgent, sessionID string, mfaUsed bool) *DomainEvent {
	data := UserLoggedInEventData{
		UserID:    userID,
		Email:     email,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		SessionID: sessionID,
		MFAUsed:   mfaUsed,
		Timestamp: time.Now(),
	}

	return NewDomainEvent(EventTypeUserLoggedIn, userID, map[string]interface{}{
		"user_id":    data.UserID,
		"email":      data.Email,
		"ip_address": data.IPAddress,
		"user_agent": data.UserAgent,
		"session_id": data.SessionID,
		"mfa_used":   data.MFAUsed,
		"timestamp":  data.Timestamp,
	})
}

// CreatePasswordChangedEvent creates a password changed event
func CreatePasswordChangedEvent(userID string, forced bool) *DomainEvent {
	data := PasswordChangedEventData{
		UserID:    userID,
		Forced:    forced,
		Timestamp: time.Now(),
	}

	return NewDomainEvent(EventTypeUserPasswordChanged, userID, map[string]interface{}{
		"user_id":   data.UserID,
		"forced":    data.Forced,
		"timestamp": data.Timestamp,
	})
}

// CreateUserLockedEvent creates a user locked event
func CreateUserLockedEvent(userID, reason string, lockedUntil time.Time) *DomainEvent {
	data := UserLockedEventData{
		UserID:      userID,
		Reason:      reason,
		LockedUntil: lockedUntil,
		Timestamp:   time.Now(),
	}

	return NewDomainEvent(EventTypeUserLocked, userID, map[string]interface{}{
		"user_id":      data.UserID,
		"reason":       data.Reason,
		"locked_until": data.LockedUntil,
		"timestamp":    data.Timestamp,
	})
}

// CreateSecurityEvent creates a security event
func CreateSecurityEvent(userID, eventType, severity, description, ipAddress, userAgent string, metadata map[string]string) *DomainEvent {
	data := SecurityEventData{
		UserID:      userID,
		EventType:   eventType,
		Severity:    severity,
		Description: description,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Metadata:    metadata,
		Timestamp:   time.Now(),
	}

	return NewDomainEvent(EventTypeSuspiciousActivity, userID, map[string]interface{}{
		"user_id":     data.UserID,
		"event_type":  data.EventType,
		"severity":    data.Severity,
		"description": data.Description,
		"ip_address":  data.IPAddress,
		"user_agent":  data.UserAgent,
		"metadata":    data.Metadata,
		"timestamp":   data.Timestamp,
	})
}
