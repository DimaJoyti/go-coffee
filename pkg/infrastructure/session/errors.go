package session

import "errors"

// Session errors
var (
	ErrSessionNotFound  = errors.New("session not found")
	ErrSessionExpired   = errors.New("session expired")
	ErrSessionRevoked   = errors.New("session revoked")
	ErrSessionInactive  = errors.New("session inactive")
	ErrInvalidSession   = errors.New("invalid session")
	ErrSessionExists    = errors.New("session already exists")
	ErrInvalidUserID    = errors.New("invalid user ID")
	ErrInvalidSessionID = errors.New("invalid session ID")
)
