package domain

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// UserStatus represents the status of a user account
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusLocked   UserStatus = "locked"
	UserStatusSuspended UserStatus = "suspended"
)

// UserRole represents the role of a user
type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

// User represents a user entity in the domain
type User struct {
	ID                string            `json:"id"`
	Email             string            `json:"email"`
	PasswordHash      string            `json:"-"` // Never serialize password hash
	Role              UserRole          `json:"role"`
	Status            UserStatus        `json:"status"`
	FailedLoginCount  int               `json:"failed_login_count"`
	LastLoginAt       *time.Time        `json:"last_login_at,omitempty"`
	LastFailedLoginAt *time.Time        `json:"last_failed_login_at,omitempty"`
	LockedUntil       *time.Time        `json:"locked_until,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// UserValidationErrors
var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrEmailRequired   = errors.New("email is required")
	ErrPasswordTooWeak = errors.New("password does not meet security requirements")
	ErrUserNotFound    = errors.New("user not found")
	ErrUserExists      = errors.New("user already exists")
	ErrUserLocked      = errors.New("user account is locked")
	ErrUserInactive    = errors.New("user account is inactive")
)

// Email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// NewUser creates a new user with validation
func NewUser(email, passwordHash string, role UserRole) (*User, error) {
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}

	if passwordHash == "" {
		return nil, errors.New("password hash is required")
	}

	now := time.Now()
	return &User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		Status:       UserStatusActive,
		Metadata:     make(map[string]string),
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return ErrEmailRequired
	}
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

// IsLocked checks if the user account is locked
func (u *User) IsLocked() bool {
	if u.Status == UserStatusLocked {
		return true
	}
	if u.LockedUntil != nil && time.Now().Before(*u.LockedUntil) {
		return true
	}
	return false
}

// IsActive checks if the user account is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive && !u.IsLocked()
}

// Lock locks the user account until the specified time
func (u *User) Lock(until time.Time) {
	u.Status = UserStatusLocked
	u.LockedUntil = &until
	u.UpdatedAt = time.Now()
}

// Unlock unlocks the user account
func (u *User) Unlock() {
	u.Status = UserStatusActive
	u.LockedUntil = nil
	u.FailedLoginCount = 0
	u.UpdatedAt = time.Now()
}

// IncrementFailedLogin increments the failed login count
func (u *User) IncrementFailedLogin() {
	u.FailedLoginCount++
	now := time.Now()
	u.LastFailedLoginAt = &now
	u.UpdatedAt = now
}

// ResetFailedLogin resets the failed login count
func (u *User) ResetFailedLogin() {
	u.FailedLoginCount = 0
	u.LastFailedLoginAt = nil
	u.UpdatedAt = time.Now()
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

// ChangePassword updates the password hash
func (u *User) ChangePassword(newPasswordHash string) {
	u.PasswordHash = newPasswordHash
	u.UpdatedAt = time.Now()
}


