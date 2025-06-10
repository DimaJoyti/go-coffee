package http

import (
	"errors"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
)

// Request DTOs

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email       string            `json:"email" validate:"required,email"`
	Password    string            `json:"password" validate:"required,min=8"`
	FirstName   string            `json:"first_name" validate:"required"`
	LastName    string            `json:"last_name" validate:"required"`
	PhoneNumber string            `json:"phone_number,omitempty"`
	Role        string            `json:"role,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	RememberMe bool   `json:"remember_me,omitempty"`
}

// LogoutRequest represents a user logout request
type LogoutRequest struct {
	Reason string `json:"reason,omitempty"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ValidateTokenRequest represents a token validation request
type ValidateTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

// UpdateUserProfileRequest represents a user profile update request
type UpdateUserProfileRequest struct {
	FirstName   string            `json:"first_name,omitempty"`
	LastName    string            `json:"last_name,omitempty"`
	PhoneNumber string            `json:"phone_number,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// EnableMFARequest represents an MFA enable request
type EnableMFARequest struct {
	Method      string `json:"method" validate:"required"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

// DisableMFARequest represents an MFA disable request
type DisableMFARequest struct {
	Code string `json:"code" validate:"required"`
}

// VerifyMFARequest represents an MFA verification request
type VerifyMFARequest struct {
	Code string `json:"code" validate:"required"`
}

// Response DTOs

// RegisterResponse represents a user registration response
type RegisterResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// LoginResponse represents a user login response
type LoginResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// LogoutResponse represents a user logout response
type LogoutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ChangePasswordResponse represents a password change response
type ChangePasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// RefreshTokenResponse represents a token refresh response
type RefreshTokenResponse struct {
	Success      bool   `json:"success"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	Message      string `json:"message,omitempty"`
}

// ValidateTokenResponse represents a token validation response
type ValidateTokenResponse struct {
	Success bool        `json:"success"`
	Valid   bool        `json:"valid"`
	Claims  interface{} `json:"claims,omitempty"`
	Message string      `json:"message,omitempty"`
}

// UserProfileResponse represents a user profile response
type UserProfileResponse struct {
	Success bool        `json:"success"`
	User    interface{} `json:"user,omitempty"`
	Message string      `json:"message,omitempty"`
}

// UserSessionsResponse represents a user sessions response
type UserSessionsResponse struct {
	Success    bool        `json:"success"`
	Sessions   interface{} `json:"sessions,omitempty"`
	Pagination interface{} `json:"pagination,omitempty"`
	Message    string      `json:"message,omitempty"`
}

// MFAStatusResponse represents an MFA status response
type MFAStatusResponse struct {
	Success   bool   `json:"success"`
	Enabled   bool   `json:"enabled"`
	Method    string `json:"method,omitempty"`
	QRCode    string `json:"qr_code,omitempty"`
	Secret    string `json:"secret,omitempty"`
	Message   string `json:"message,omitempty"`
}

// BackupCodesResponse represents backup codes response
type BackupCodesResponse struct {
	Success bool     `json:"success"`
	Codes   []string `json:"codes,omitempty"`
	Message string   `json:"message,omitempty"`
}

// SecurityEventsResponse represents security events response
type SecurityEventsResponse struct {
	Success    bool        `json:"success"`
	Events     interface{} `json:"events,omitempty"`
	Pagination interface{} `json:"pagination,omitempty"`
	Message    string      `json:"message,omitempty"`
}

// TrustedDevicesResponse represents trusted devices response
type TrustedDevicesResponse struct {
	Success bool        `json:"success"`
	Devices interface{} `json:"devices,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// Validation methods

// validateRegisterRequest validates a registration request
func (h *CleanHandler) validateRegisterRequest(req *RegisterRequest) error {
	if req.Email == "" {
		return errors.New("email is required")
	}

	if req.Password == "" {
		return errors.New("password is required")
	}

	if len(req.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if req.FirstName == "" {
		return errors.New("first name is required")
	}

	if req.LastName == "" {
		return errors.New("last name is required")
	}

	// Validate email format
	if err := domain.ValidateEmail(req.Email); err != nil {
		return err
	}

	// Validate phone number if provided
	if req.PhoneNumber != "" {
		if err := domain.ValidatePhoneNumber(req.PhoneNumber); err != nil {
			return err
		}
	}

	// Validate role if provided
	if req.Role != "" {
		validRoles := []string{"user", "moderator", "admin"}
		roleValid := false
		for _, validRole := range validRoles {
			if strings.ToLower(req.Role) == validRole {
				roleValid = true
				break
			}
		}
		if !roleValid {
			return errors.New("invalid role")
		}
	}

	return nil
}

// validateLoginRequest validates a login request
func (h *CleanHandler) validateLoginRequest(req *LoginRequest) error {
	if req.Email == "" {
		return errors.New("email is required")
	}

	if req.Password == "" {
		return errors.New("password is required")
	}

	// Validate email format
	if err := domain.ValidateEmail(req.Email); err != nil {
		return err
	}

	return nil
}

// validateChangePasswordRequest validates a password change request
func (h *CleanHandler) validateChangePasswordRequest(req *ChangePasswordRequest) error {
	if req.OldPassword == "" {
		return errors.New("old password is required")
	}

	if req.NewPassword == "" {
		return errors.New("new password is required")
	}

	if len(req.NewPassword) < 8 {
		return errors.New("new password must be at least 8 characters long")
	}

	if req.OldPassword == req.NewPassword {
		return errors.New("new password must be different from old password")
	}

	return nil
}

// validateRefreshTokenRequest validates a token refresh request
func (h *CleanHandler) validateRefreshTokenRequest(req *RefreshTokenRequest) error {
	if req.RefreshToken == "" {
		return errors.New("refresh token is required")
	}

	return nil
}

// validateValidateTokenRequest validates a token validation request
func (h *CleanHandler) validateValidateTokenRequest(req *ValidateTokenRequest) error {
	if req.Token == "" {
		return errors.New("token is required")
	}

	return nil
}

// validateUpdateUserProfileRequest validates a user profile update request
func (h *CleanHandler) validateUpdateUserProfileRequest(req *UpdateUserProfileRequest) error {
	// Validate phone number if provided
	if req.PhoneNumber != "" {
		if err := domain.ValidatePhoneNumber(req.PhoneNumber); err != nil {
			return err
		}
	}

	return nil
}

// validateEnableMFARequest validates an MFA enable request
func (h *CleanHandler) validateEnableMFARequest(req *EnableMFARequest) error {
	if req.Method == "" {
		return errors.New("MFA method is required")
	}

	validMethods := []string{"totp", "sms", "email"}
	methodValid := false
	for _, validMethod := range validMethods {
		if strings.ToLower(req.Method) == validMethod {
			methodValid = true
			break
		}
	}

	if !methodValid {
		return errors.New("invalid MFA method")
	}

	// Validate phone number for SMS method
	if strings.ToLower(req.Method) == "sms" {
		if req.PhoneNumber == "" {
			return errors.New("phone number is required for SMS MFA")
		}
		if err := domain.ValidatePhoneNumber(req.PhoneNumber); err != nil {
			return err
		}
	}

	return nil
}

// validateDisableMFARequest validates an MFA disable request
func (h *CleanHandler) validateDisableMFARequest(req *DisableMFARequest) error {
	if req.Code == "" {
		return errors.New("MFA code is required")
	}

	return nil
}

// validateVerifyMFARequest validates an MFA verification request
func (h *CleanHandler) validateVerifyMFARequest(req *VerifyMFARequest) error {
	if req.Code == "" {
		return errors.New("MFA code is required")
	}

	return nil
}
