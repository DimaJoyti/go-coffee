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
	Email      string             `json:"email" validate:"required,email"`
	Password   string             `json:"password" validate:"required"`
	DeviceInfo *domain.DeviceInfo `json:"device_info,omitempty"`
	RememberMe bool               `json:"remember_me,omitempty"`
	IPAddress  string             `json:"ip_address,omitempty"`
	UserAgent  string             `json:"user_agent,omitempty"`
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
	LogoutAll bool   `json:"logout_all,omitempty"`
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
	Valid     bool       `json:"valid"`
	UserID    string     `json:"user_id,omitempty"`
	Role      string     `json:"role,omitempty"`
	SessionID string     `json:"session_id,omitempty"`
	User      *UserDTO   `json:"user,omitempty"`
	Claims    *ClaimsDTO `json:"claims,omitempty"`
	Message   string     `json:"message,omitempty"`
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
	UserID string `json:"user_id" validate:"required"`
}

// GetUserInfoResponse represents a get user info response
type GetUserInfoResponse struct {
	User *UserDTO `json:"user"`
}

// UserDTO represents a user data transfer object
type UserDTO struct {
	ID          string            `json:"id"`
	Email       string            `json:"email"`
	Role        string            `json:"role"`
	Status      string            `json:"status"`
	LastLoginAt *time.Time        `json:"last_login_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// SessionDTO represents a session data transfer object
type SessionDTO struct {
	ID         string             `json:"id"`
	UserID     string             `json:"user_id"`
	Status     string             `json:"status"`
	ExpiresAt  time.Time          `json:"expires_at"`
	DeviceInfo *domain.DeviceInfo `json:"device_info,omitempty"`
	IPAddress  string             `json:"ip_address,omitempty"`
	UserAgent  string             `json:"user_agent,omitempty"`
	CreatedAt  time.Time          `json:"created_at"`
	LastUsedAt *time.Time         `json:"last_used_at,omitempty"`
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

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	TokenType             string    `json:"token_type"`
}

// TokenClaims represents JWT token claims for application layer
type TokenClaims struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	SessionID string    `json:"session_id"`
	TokenID   string    `json:"token_id"`
	Type      string    `json:"type"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Issuer    string    `json:"issuer"`
	Audience  string    `json:"audience"`
}

// TokenMetadata represents token metadata
type TokenMetadata struct {
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Expired   bool      `json:"expired"`
}

// SecurityAnalysis represents the result of security analysis
type SecurityAnalysis struct {
	UserID    string    `json:"user_id"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	RiskScore float64   `json:"risk_score"`
	RiskLevel string    `json:"risk_level"`
	Factors   []string  `json:"factors"`
	Timestamp time.Time `json:"timestamp"`
}

// Additional DTOs for transport layer

// MFA DTOs

// EnableMFARequest represents a request to enable MFA
type EnableMFARequest struct {
	UserID      string           `json:"user_id" validate:"required"`
	Method      domain.MFAMethod `json:"method" validate:"required"`
	PhoneNumber string           `json:"phone_number,omitempty"`
}

// EnableMFAResponse represents a response to enable MFA
type EnableMFAResponse struct {
	Success     bool     `json:"success"`
	Message     string   `json:"message"`
	QRCode      string   `json:"qr_code,omitempty"`
	Secret      string   `json:"secret,omitempty"`
	BackupCodes []string `json:"backup_codes,omitempty"`
}

// DisableMFARequest represents a request to disable MFA
type DisableMFARequest struct {
	UserID   string `json:"user_id" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// DisableMFAResponse represents a response to disable MFA
type DisableMFAResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// VerifyMFARequest represents a request to verify MFA
type VerifyMFARequest struct {
	UserID string `json:"user_id" validate:"required"`
	Code   string `json:"code" validate:"required"`
}

// VerifyMFAResponse represents a response to MFA verification
type VerifyMFAResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GenerateBackupCodesRequest represents a request to generate backup codes
type GenerateBackupCodesRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// GenerateBackupCodesResponse represents a response with backup codes
type GenerateBackupCodesResponse struct {
	BackupCodes []string `json:"backup_codes"`
	Message     string   `json:"message"`
}

// GetBackupCodesRequest represents a request to get backup codes
type GetBackupCodesRequest struct {
	UserID string `json:"user_id" validate:"required"`
}

// GetBackupCodesResponse represents a response with backup codes count
type GetBackupCodesResponse struct {
	RemainingCodes int    `json:"remaining_codes"`
	Message        string `json:"message"`
}

// MFAStatusResponse represents MFA status information
type MFAStatusResponse struct {
	Enabled     bool             `json:"enabled"`
	Method      domain.MFAMethod `json:"method,omitempty"`
	BackupCodes int              `json:"backup_codes_remaining"`
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ForgotPasswordResponse represents a forgot password response
type ForgotPasswordResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ResetPasswordResponse represents a password reset response
type ResetPasswordResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// GetUserSessionsRequest represents a request to get user sessions
type GetUserSessionsRequest struct {
	UserID string `json:"user_id" validate:"required"`
}

// GetUserSessionsResponse represents a response with user sessions
type GetUserSessionsResponse struct {
	Sessions []*SessionDTO `json:"sessions"`
	Total    int           `json:"total"`
}

// RevokeSessionRequest represents a request to revoke a session
type RevokeSessionRequest struct {
	UserID    string `json:"user_id" validate:"required"`
	SessionID string `json:"session_id" validate:"required"`
}

// RevokeSessionResponse represents a response to session revocation
type RevokeSessionResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// GetSecurityEventsRequest represents a request to get security events
type GetSecurityEventsRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

// GetSecurityEventsResponse represents a response with security events
type GetSecurityEventsResponse struct {
	Events []*SecurityEventDTO `json:"events"`
	Total  int                 `json:"total"`
}

// GetTrustedDevicesRequest represents a request to get trusted devices
type GetTrustedDevicesRequest struct {
	UserID string `json:"user_id" validate:"required"`
}

// GetTrustedDevicesResponse represents a response with trusted devices
type GetTrustedDevicesResponse struct {
	Devices []*DeviceDTO `json:"devices"`
	Total   int          `json:"total"`
}

// TrustDeviceRequest represents a request to trust a device
type TrustDeviceRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	DeviceID string `json:"device_id" validate:"required"`
}

// TrustDeviceResponse represents a response to device trust
type TrustDeviceResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// RemoveDeviceRequest represents a request to remove a device
type RemoveDeviceRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	DeviceID string `json:"device_id" validate:"required"`
}

// RemoveDeviceResponse represents a response to device removal
type RemoveDeviceResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// DeviceDTO represents device information
type DeviceDTO struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Fingerprint string    `json:"fingerprint"`
	UserAgent   string    `json:"user_agent"`
	IPAddress   string    `json:"ip_address"`
	Location    string    `json:"location,omitempty"`
	Trusted     bool      `json:"trusted"`
	LastUsedAt  time.Time `json:"last_used_at"`
	CreatedAt   time.Time `json:"created_at"`
}
