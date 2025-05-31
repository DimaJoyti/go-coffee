package application

import (
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
)

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role,omitempty"`
}

// RegisterResponse represents a user registration response
type RegisterResponse struct {
	User         *UserDTO `json:"user"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int64    `json:"expires_in"`
	TokenType    string   `json:"token_type"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email      string                `json:"email" validate:"required,email"`
	Password   string                `json:"password" validate:"required"`
	DeviceInfo *domain.DeviceInfo    `json:"device_info,omitempty"`
	RememberMe bool                  `json:"remember_me,omitempty"`
}

// LoginResponse represents a user login response
type LoginResponse struct {
	User         *UserDTO `json:"user"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int64    `json:"expires_in"`
	TokenType    string   `json:"token_type"`
}

// LogoutRequest represents a user logout request
type LogoutRequest struct {
	SessionID string `json:"session_id,omitempty"`
	AllSessions bool `json:"all_sessions,omitempty"`
}

// LogoutResponse represents a user logout response
type LogoutResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse represents a token refresh response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// ValidateTokenRequest represents a token validation request
type ValidateTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

// ValidateTokenResponse represents a token validation response
type ValidateTokenResponse struct {
	Valid   bool     `json:"valid"`
	User    *UserDTO `json:"user,omitempty"`
	Claims  *ClaimsDTO `json:"claims,omitempty"`
	Message string   `json:"message,omitempty"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// ChangePasswordResponse represents a password change response
type ChangePasswordResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// GetUserInfoRequest represents a get user info request
type GetUserInfoRequest struct {
	UserID string `json:"user_id,omitempty"`
}

// GetUserInfoResponse represents a get user info response
type GetUserInfoResponse struct {
	User *UserDTO `json:"user"`
}

// UserDTO represents a user data transfer object
type UserDTO struct {
	ID        string                `json:"id"`
	Email     string                `json:"email"`
	Role      string                `json:"role"`
	Status    string                `json:"status"`
	LastLoginAt *time.Time          `json:"last_login_at,omitempty"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	Metadata  map[string]string     `json:"metadata,omitempty"`
}

// SessionDTO represents a session data transfer object
type SessionDTO struct {
	ID           string                `json:"id"`
	UserID       string                `json:"user_id"`
	Status       string                `json:"status"`
	ExpiresAt    time.Time             `json:"expires_at"`
	DeviceInfo   *domain.DeviceInfo    `json:"device_info,omitempty"`
	IPAddress    string                `json:"ip_address,omitempty"`
	UserAgent    string                `json:"user_agent,omitempty"`
	CreatedAt    time.Time             `json:"created_at"`
	LastUsedAt   *time.Time            `json:"last_used_at,omitempty"`
}

// ClaimsDTO represents token claims data transfer object
type ClaimsDTO struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	SessionID string    `json:"session_id"`
	TokenID   string    `json:"token_id"`
	Type      string    `json:"type"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// SecurityEventDTO represents a security event data transfer object
type SecurityEventDTO struct {
	ID          string            `json:"id"`
	UserID      string            `json:"user_id,omitempty"`
	Type        string            `json:"type"`
	Severity    string            `json:"severity"`
	Description string            `json:"description"`
	IPAddress   string            `json:"ip_address,omitempty"`
	UserAgent   string            `json:"user_agent,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Success bool        `json:"success"`
}

// Conversion functions

// ToUserDTO converts a domain User to UserDTO
func ToUserDTO(user *domain.User) *UserDTO {
	if user == nil {
		return nil
	}
	return &UserDTO{
		ID:          user.ID,
		Email:       user.Email,
		Role:        string(user.Role),
		Status:      string(user.Status),
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Metadata:    user.Metadata,
	}
}

// ToSessionDTO converts a domain Session to SessionDTO
func ToSessionDTO(session *domain.Session) *SessionDTO {
	if session == nil {
		return nil
	}
	return &SessionDTO{
		ID:         session.ID,
		UserID:     session.UserID,
		Status:     string(session.Status),
		ExpiresAt:  session.ExpiresAt,
		DeviceInfo: session.DeviceInfo,
		IPAddress:  session.IPAddress,
		UserAgent:  session.UserAgent,
		CreatedAt:  session.CreatedAt,
		LastUsedAt: session.LastUsedAt,
	}
}

// ToClaimsDTO converts domain TokenClaims to ClaimsDTO
func ToClaimsDTO(claims *domain.TokenClaims) *ClaimsDTO {
	if claims == nil {
		return nil
	}
	return &ClaimsDTO{
		UserID:    claims.UserID,
		Email:     claims.Email,
		Role:      string(claims.Role),
		SessionID: claims.SessionID,
		TokenID:   claims.TokenID,
		Type:      string(claims.Type),
		IssuedAt:  claims.IssuedAt,
		ExpiresAt: claims.ExpiresAt,
	}
}

// ToSecurityEventDTO converts domain SecurityEvent to SecurityEventDTO
func ToSecurityEventDTO(event *domain.SecurityEvent) *SecurityEventDTO {
	if event == nil {
		return nil
	}
	return &SecurityEventDTO{
		ID:          event.ID,
		UserID:      event.UserID,
		Type:        string(event.Type),
		Severity:    string(event.Severity),
		Description: event.Description,
		IPAddress:   event.IPAddress,
		UserAgent:   event.UserAgent,
		Metadata:    event.Metadata,
		CreatedAt:   event.CreatedAt,
	}
}
