package domain

import (
	"context"
	"time"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// User CRUD operations
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, userID string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, userID string) error
	
	// User status operations
	LockUser(ctx context.Context, userID string, until time.Time) error
	UnlockUser(ctx context.Context, userID string) error
	
	// Failed login tracking
	IncrementFailedLogin(ctx context.Context, userID string) error
	ResetFailedLogin(ctx context.Context, userID string) error
	GetFailedLoginCount(ctx context.Context, email string) (int, error)
	SetFailedLoginCount(ctx context.Context, email string, count int, expiry time.Duration) error
	
	// User existence checks
	UserExists(ctx context.Context, email string) (bool, error)
	UserExistsByID(ctx context.Context, userID string) (bool, error)
}

// SessionRepository defines the interface for session data operations
type SessionRepository interface {
	// Session CRUD operations
	CreateSession(ctx context.Context, session *Session) error
	GetSessionByID(ctx context.Context, sessionID string) (*Session, error)
	GetSessionByAccessToken(ctx context.Context, accessToken string) (*Session, error)
	GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*Session, error)
	UpdateSession(ctx context.Context, session *Session) error
	DeleteSession(ctx context.Context, sessionID string) error
	
	// User sessions management
	GetUserSessions(ctx context.Context, userID string) ([]*Session, error)
	DeleteUserSessions(ctx context.Context, userID string) error
	DeleteExpiredSessions(ctx context.Context) error
	
	// Session status operations
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeUserSessions(ctx context.Context, userID string) error
	
	// Session validation
	IsSessionValid(ctx context.Context, sessionID string) (bool, error)
	UpdateSessionLastUsed(ctx context.Context, sessionID string) error
}

// TokenRepository defines the interface for token data operations
type TokenRepository interface {
	// Token CRUD operations
	CreateToken(ctx context.Context, token *Token) error
	GetTokenByID(ctx context.Context, tokenID string) (*Token, error)
	GetTokenByValue(ctx context.Context, tokenValue string) (*Token, error)
	UpdateToken(ctx context.Context, token *Token) error
	DeleteToken(ctx context.Context, tokenID string) error
	
	// Token validation and management
	RevokeToken(ctx context.Context, tokenID string) error
	RevokeTokensBySession(ctx context.Context, sessionID string) error
	RevokeTokensByUser(ctx context.Context, userID string) error
	DeleteExpiredTokens(ctx context.Context) error
	
	// Token existence and validation
	IsTokenRevoked(ctx context.Context, tokenID string) (bool, error)
	IsTokenValid(ctx context.Context, tokenID string) (bool, error)
	
	// Blacklist operations (for revoked tokens)
	AddToBlacklist(ctx context.Context, tokenID string, expiresAt time.Time) error
	IsBlacklisted(ctx context.Context, tokenID string) (bool, error)
	CleanupBlacklist(ctx context.Context) error
}

// SecurityEventRepository defines the interface for security event logging
type SecurityEventRepository interface {
	// Security event logging
	LogSecurityEvent(ctx context.Context, event *SecurityEvent) error
	GetSecurityEvents(ctx context.Context, userID string, limit int) ([]*SecurityEvent, error)
	GetSecurityEventsByType(ctx context.Context, eventType SecurityEventType, limit int) ([]*SecurityEvent, error)
	DeleteOldSecurityEvents(ctx context.Context, olderThan time.Time) error
}

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	ID          string              `json:"id"`
	UserID      string              `json:"user_id,omitempty"`
	Type        SecurityEventType   `json:"type"`
	Severity    SecuritySeverity    `json:"severity"`
	Description string              `json:"description"`
	IPAddress   string              `json:"ip_address,omitempty"`
	UserAgent   string              `json:"user_agent,omitempty"`
	Metadata    map[string]string   `json:"metadata,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
}

// SecurityEventType represents the type of security event
type SecurityEventType string

const (
	SecurityEventTypeLogin            SecurityEventType = "login"
	SecurityEventTypeLoginFailed      SecurityEventType = "login_failed"
	SecurityEventTypeLogout           SecurityEventType = "logout"
	SecurityEventTypePasswordChange   SecurityEventType = "password_change"
	SecurityEventTypeAccountLocked    SecurityEventType = "account_locked"
	SecurityEventTypeAccountUnlocked  SecurityEventType = "account_unlocked"
	SecurityEventTypeTokenRefresh     SecurityEventType = "token_refresh"
	SecurityEventTypeTokenRevoked     SecurityEventType = "token_revoked"
	SecurityEventTypeSessionCreated   SecurityEventType = "session_created"
	SecurityEventTypeSessionRevoked   SecurityEventType = "session_revoked"
	SecurityEventTypeSuspiciousActivity SecurityEventType = "suspicious_activity"
)

// SecuritySeverity represents the severity level of a security event
type SecuritySeverity string

const (
	SecuritySeverityLow      SecuritySeverity = "low"
	SecuritySeverityMedium   SecuritySeverity = "medium"
	SecuritySeverityHigh     SecuritySeverity = "high"
	SecuritySeverityCritical SecuritySeverity = "critical"
)

// NewSecurityEvent creates a new security event
func NewSecurityEvent(userID string, eventType SecurityEventType, severity SecuritySeverity, description string) *SecurityEvent {
	return &SecurityEvent{
		ID:          GenerateSecurityEventID(),
		UserID:      userID,
		Type:        eventType,
		Severity:    severity,
		Description: description,
		Metadata:    make(map[string]string),
		CreatedAt:   time.Now(),
	}
}

// GenerateSecurityEventID generates a unique security event ID
func GenerateSecurityEventID() string {
	return "event_" + time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
