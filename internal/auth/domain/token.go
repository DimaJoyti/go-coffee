package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// TokenType represents the type of token
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// TokenStatus represents the status of a token
type TokenStatus string

const (
	TokenStatusActive  TokenStatus = "active"
	TokenStatusRevoked TokenStatus = "revoked"
	TokenStatusExpired TokenStatus = "expired"
)

// Token represents a JWT token in the domain
type Token struct {
	AggregateRoot // Embed aggregate root for event functionality

	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	SessionID string            `json:"session_id"`
	Type      TokenType         `json:"type"`
	Status    TokenStatus       `json:"status"`
	Value     string            `json:"value"`
	ExpiresAt time.Time         `json:"expires_at"`
	IssuedAt  time.Time         `json:"issued_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// TokenClaims represents the claims in a JWT token
type TokenClaims struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Role      UserRole  `json:"role"`
	SessionID string    `json:"session_id"`
	TokenID   string    `json:"token_id"`
	Type      TokenType `json:"type"`
	IssuedAt  time.Time `json:"iat"`
	ExpiresAt time.Time `json:"exp"`
}

// Token validation errors
var (
	ErrTokenNotFound    = errors.New("token not found")
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenRevoked     = errors.New("token revoked")
	ErrTokenInvalid     = errors.New("token invalid")
	ErrTokenMalformed   = errors.New("token malformed")
	ErrTokenSignature   = errors.New("token signature invalid")
	ErrTokenClaims      = errors.New("token claims invalid")
	ErrTokenTypeInvalid = errors.New("token type invalid")
)

// NewToken creates a new token
func NewToken(userID, sessionID string, tokenType TokenType, value string, expiresAt time.Time) *Token {
	now := time.Now()
	token := &Token{
		ID:        uuid.New().String(),
		UserID:    userID,
		SessionID: sessionID,
		Type:      tokenType,
		Status:    TokenStatusActive,
		Value:     value,
		ExpiresAt: expiresAt,
		IssuedAt:  now,
		Metadata:  make(map[string]string),
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Generate token generated event
	event := NewDomainEvent(EventTypeTokenGenerated, token.ID, map[string]interface{}{
		"token_id":   token.ID,
		"user_id":    token.UserID,
		"session_id": token.SessionID,
		"type":       token.Type,
		"expires_at": token.ExpiresAt,
		"timestamp":  now,
	})
	token.AddEvent(*event)

	return token
}

// NewTokenClaims creates new token claims
func NewTokenClaims(userID, email string, role UserRole, sessionID, tokenID string, tokenType TokenType, expiresAt time.Time) *TokenClaims {
	now := time.Now()
	return &TokenClaims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		SessionID: sessionID,
		TokenID:   tokenID,
		Type:      tokenType,
		IssuedAt:  now,
		ExpiresAt: expiresAt,
	}
}

// IsValid checks if the token is valid and not expired
func (t *Token) IsValid() bool {
	now := time.Now()
	return t.Status == TokenStatusActive && now.Before(t.ExpiresAt)
}

// IsExpired checks if the token is expired
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsRevoked checks if the token is revoked
func (t *Token) IsRevoked() bool {
	return t.Status == TokenStatusRevoked
}

// Revoke revokes the token
func (t *Token) Revoke() {
	t.Status = TokenStatusRevoked
	t.UpdatedAt = time.Now()

	// Generate token revoked event
	event := NewDomainEvent(EventTypeTokenRevoked, t.ID, map[string]interface{}{
		"token_id":   t.ID,
		"user_id":    t.UserID,
		"session_id": t.SessionID,
		"type":       t.Type,
		"timestamp":  t.UpdatedAt,
	})
	t.AddEvent(*event)
}

// Expire expires the token
func (t *Token) Expire() {
	t.Status = TokenStatusExpired
	t.UpdatedAt = time.Now()

	// Generate token expired event
	event := NewDomainEvent(EventTypeTokenExpired, t.ID, map[string]interface{}{
		"token_id":   t.ID,
		"user_id":    t.UserID,
		"session_id": t.SessionID,
		"type":       t.Type,
		"timestamp":  t.UpdatedAt,
	})
	t.AddEvent(*event)
}

// Validate validates the token and returns appropriate error
func (t *Token) Validate() error {
	switch t.Status {
	case TokenStatusRevoked:
		return ErrTokenRevoked
	case TokenStatusExpired:
		return ErrTokenExpired
	}

	if t.IsExpired() {
		return ErrTokenExpired
	}

	return nil
}

// IsValidType checks if the token type is valid
func (t *Token) IsValidType(expectedType TokenType) bool {
	return t.Type == expectedType
}

// GetTimeUntilExpiry returns the duration until the token expires
func (t *Token) GetTimeUntilExpiry() time.Duration {
	return time.Until(t.ExpiresAt)
}

// IsValidClaims validates the token claims
func (tc *TokenClaims) IsValidClaims() error {
	if tc.UserID == "" {
		return ErrTokenClaims
	}
	if tc.Email == "" {
		return ErrTokenClaims
	}
	if tc.SessionID == "" {
		return ErrTokenClaims
	}
	if tc.TokenID == "" {
		return ErrTokenClaims
	}
	if tc.Type != TokenTypeAccess && tc.Type != TokenTypeRefresh {
		return ErrTokenTypeInvalid
	}
	if time.Now().After(tc.ExpiresAt) {
		return ErrTokenExpired
	}
	return nil
}

// IsExpiredClaims checks if the token claims are expired
func (tc *TokenClaims) IsExpiredClaims() bool {
	return time.Now().After(tc.ExpiresAt)
}

// GetTimeUntilExpiryClaims returns the duration until the token claims expire
func (tc *TokenClaims) GetTimeUntilExpiryClaims() time.Duration {
	return time.Until(tc.ExpiresAt)
}
