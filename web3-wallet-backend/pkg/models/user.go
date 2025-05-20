package models

import (
	"time"
)

// User represents a user of the system
type User struct {
	ID             string    `json:"id" db:"id"`
	Email          string    `json:"email" db:"email"`
	PasswordHash   string    `json:"-" db:"password_hash"`
	FirstName      string    `json:"first_name" db:"first_name"`
	LastName       string    `json:"last_name" db:"last_name"`
	Phone          string    `json:"phone" db:"phone"`
	EmailVerified  bool      `json:"email_verified" db:"email_verified"`
	PhoneVerified  bool      `json:"phone_verified" db:"phone_verified"`
	TwoFactorEnabled bool    `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorSecret string   `json:"-" db:"two_factor_secret"`
	Role           UserRole  `json:"role" db:"role"`
	Status         UserStatus `json:"status" db:"status"`
	LastLoginAt    time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// UserRole represents the role of a user
type UserRole string

const (
	// UserRoleUser represents a regular user
	UserRoleUser UserRole = "user"
	// UserRoleAdmin represents an administrator
	UserRoleAdmin UserRole = "admin"
)

// UserStatus represents the status of a user
type UserStatus string

const (
	// UserStatusActive represents an active user
	UserStatusActive UserStatus = "active"
	// UserStatusInactive represents an inactive user
	UserStatusInactive UserStatus = "inactive"
	// UserStatusSuspended represents a suspended user
	UserStatusSuspended UserStatus = "suspended"
)

// RegisterRequest represents a request to register a new user
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Phone     string `json:"phone"`
}

// RegisterResponse represents a response to a register request
type RegisterResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

// LoginRequest represents a request to login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	TwoFactorCode string `json:"two_factor_code"`
}

// LoginResponse represents a response to a login request
type LoginResponse struct {
	User         User   `json:"user"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokenRequest represents a request to refresh a token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse represents a response to a refresh token request
type RefreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// GetUserRequest represents a request to get a user
type GetUserRequest struct {
	ID string `json:"id" validate:"required"`
}

// GetUserResponse represents a response to a get user request
type GetUserResponse struct {
	User User `json:"user"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	ID        string `json:"id" validate:"required"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

// UpdateUserResponse represents a response to an update user request
type UpdateUserResponse struct {
	User User `json:"user"`
}

// ChangePasswordRequest represents a request to change a password
type ChangePasswordRequest struct {
	ID          string `json:"id" validate:"required"`
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ChangePasswordResponse represents a response to a change password request
type ChangePasswordResponse struct {
	Success bool `json:"success"`
}

// ForgotPasswordRequest represents a request to initiate a password reset
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ForgotPasswordResponse represents a response to a forgot password request
type ForgotPasswordResponse struct {
	Success bool `json:"success"`
}

// ResetPasswordRequest represents a request to reset a password
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ResetPasswordResponse represents a response to a reset password request
type ResetPasswordResponse struct {
	Success bool `json:"success"`
}

// VerifyEmailRequest represents a request to verify an email
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// VerifyEmailResponse represents a response to a verify email request
type VerifyEmailResponse struct {
	Success bool `json:"success"`
}

// EnableTwoFactorRequest represents a request to enable two-factor authentication
type EnableTwoFactorRequest struct {
	ID   string `json:"id" validate:"required"`
	Code string `json:"code" validate:"required"`
}

// EnableTwoFactorResponse represents a response to an enable two-factor authentication request
type EnableTwoFactorResponse struct {
	Secret     string `json:"secret"`
	QRCodeURL  string `json:"qr_code_url"`
	RecoveryCodes []string `json:"recovery_codes"`
}

// DisableTwoFactorRequest represents a request to disable two-factor authentication
type DisableTwoFactorRequest struct {
	ID       string `json:"id" validate:"required"`
	Password string `json:"password" validate:"required"`
	Code     string `json:"code" validate:"required"`
}

// DisableTwoFactorResponse represents a response to a disable two-factor authentication request
type DisableTwoFactorResponse struct {
	Success bool `json:"success"`
}
