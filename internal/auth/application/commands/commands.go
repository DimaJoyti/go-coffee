package commands

import (
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
)

// Command represents a command in the CQRS pattern
type Command interface {
	CommandType() string
}

// CommandHandler represents a command handler
type CommandHandler[T Command] interface {
	Handle(cmd T) error
}

// User Commands

// RegisterUserCommand represents a command to register a new user
type RegisterUserCommand struct {
	Email       string            `json:"email" validate:"required,email"`
	Password    string            `json:"password" validate:"required,min=8"`
	FirstName   string            `json:"first_name" validate:"required"`
	LastName    string            `json:"last_name" validate:"required"`
	PhoneNumber string            `json:"phone_number,omitempty"`
	Role        domain.UserRole   `json:"role,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

func (c RegisterUserCommand) CommandType() string { return "RegisterUser" }

// LoginUserCommand represents a command to login a user
type LoginUserCommand struct {
	Email      string             `json:"email" validate:"required,email"`
	Password   string             `json:"password" validate:"required"`
	DeviceInfo *domain.DeviceInfo `json:"device_info,omitempty"`
	RememberMe bool               `json:"remember_me,omitempty"`
	IPAddress  string             `json:"ip_address,omitempty"`
	UserAgent  string             `json:"user_agent,omitempty"`
}

func (c LoginUserCommand) CommandType() string { return "LoginUser" }

// LogoutUserCommand represents a command to logout a user
type LogoutUserCommand struct {
	UserID    string `json:"user_id" validate:"required"`
	SessionID string `json:"session_id" validate:"required"`
	Reason    string `json:"reason,omitempty"`
}

func (c LogoutUserCommand) CommandType() string { return "LogoutUser" }

// ChangePasswordCommand represents a command to change user password
type ChangePasswordCommand struct {
	UserID      string `json:"user_id" validate:"required"`
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
	Forced      bool   `json:"forced,omitempty"`
}

func (c ChangePasswordCommand) CommandType() string { return "ChangePassword" }

// ResetPasswordCommand represents a command to reset user password
type ResetPasswordCommand struct {
	Email       string `json:"email" validate:"required,email"`
	ResetToken  string `json:"reset_token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

func (c ResetPasswordCommand) CommandType() string { return "ResetPassword" }

// UpdateUserProfileCommand represents a command to update user profile
type UpdateUserProfileCommand struct {
	UserID      string            `json:"user_id" validate:"required"`
	FirstName   string            `json:"first_name,omitempty"`
	LastName    string            `json:"last_name,omitempty"`
	PhoneNumber string            `json:"phone_number,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

func (c UpdateUserProfileCommand) CommandType() string { return "UpdateUserProfile" }

// DeactivateUserCommand represents a command to deactivate a user
type DeactivateUserCommand struct {
	UserID string `json:"user_id" validate:"required"`
	Reason string `json:"reason,omitempty"`
}

func (c DeactivateUserCommand) CommandType() string { return "DeactivateUser" }

// ReactivateUserCommand represents a command to reactivate a user
type ReactivateUserCommand struct {
	UserID string `json:"user_id" validate:"required"`
	Reason string `json:"reason,omitempty"`
}

func (c ReactivateUserCommand) CommandType() string { return "ReactivateUser" }

// Session Commands

// RefreshTokenCommand represents a command to refresh tokens
type RefreshTokenCommand struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
	DeviceInfo   *domain.DeviceInfo `json:"device_info,omitempty"`
}

func (c RefreshTokenCommand) CommandType() string { return "RefreshToken" }

// RevokeSessionCommand represents a command to revoke a session
type RevokeSessionCommand struct {
	UserID    string `json:"user_id" validate:"required"`
	SessionID string `json:"session_id" validate:"required"`
	Reason    string `json:"reason,omitempty"`
}

func (c RevokeSessionCommand) CommandType() string { return "RevokeSession" }

// RevokeAllUserSessionsCommand represents a command to revoke all user sessions
type RevokeAllUserSessionsCommand struct {
	UserID string `json:"user_id" validate:"required"`
	Reason string `json:"reason,omitempty"`
}

func (c RevokeAllUserSessionsCommand) CommandType() string { return "RevokeAllUserSessions" }

// MFA Commands

// EnableMFACommand represents a command to enable MFA
type EnableMFACommand struct {
	UserID      string           `json:"user_id" validate:"required"`
	Method      domain.MFAMethod `json:"method" validate:"required"`
	PhoneNumber string           `json:"phone_number,omitempty"`
}

func (c EnableMFACommand) CommandType() string { return "EnableMFA" }

// DisableMFACommand represents a command to disable MFA
type DisableMFACommand struct {
	UserID string `json:"user_id" validate:"required"`
	Code   string `json:"code" validate:"required"`
}

func (c DisableMFACommand) CommandType() string { return "DisableMFA" }

// VerifyMFACommand represents a command to verify MFA
type VerifyMFACommand struct {
	UserID    string `json:"user_id" validate:"required"`
	Code      string `json:"code" validate:"required"`
	SessionID string `json:"session_id" validate:"required"`
}

func (c VerifyMFACommand) CommandType() string { return "VerifyMFA" }

// GenerateBackupCodesCommand represents a command to generate backup codes
type GenerateBackupCodesCommand struct {
	UserID string `json:"user_id" validate:"required"`
}

func (c GenerateBackupCodesCommand) CommandType() string { return "GenerateBackupCodes" }

// UseBackupCodeCommand represents a command to use a backup code
type UseBackupCodeCommand struct {
	UserID string `json:"user_id" validate:"required"`
	Code   string `json:"code" validate:"required"`
}

func (c UseBackupCodeCommand) CommandType() string { return "UseBackupCode" }

// Security Commands

// LockUserAccountCommand represents a command to lock a user account
type LockUserAccountCommand struct {
	UserID    string        `json:"user_id" validate:"required"`
	Reason    string        `json:"reason" validate:"required"`
	Duration  time.Duration `json:"duration,omitempty"`
	AdminID   string        `json:"admin_id,omitempty"`
}

func (c LockUserAccountCommand) CommandType() string { return "LockUserAccount" }

// UnlockUserAccountCommand represents a command to unlock a user account
type UnlockUserAccountCommand struct {
	UserID  string `json:"user_id" validate:"required"`
	Reason  string `json:"reason,omitempty"`
	AdminID string `json:"admin_id,omitempty"`
}

func (c UnlockUserAccountCommand) CommandType() string { return "UnlockUserAccount" }

// UpdateRiskScoreCommand represents a command to update user risk score
type UpdateRiskScoreCommand struct {
	UserID   string   `json:"user_id" validate:"required"`
	NewScore float64  `json:"new_score" validate:"min=0,max=1"`
	Factors  []string `json:"factors,omitempty"`
}

func (c UpdateRiskScoreCommand) CommandType() string { return "UpdateRiskScore" }

// AddDeviceCommand represents a command to add a trusted device
type AddDeviceCommand struct {
	UserID      string `json:"user_id" validate:"required"`
	DeviceID    string `json:"device_id" validate:"required"`
	Fingerprint string `json:"fingerprint" validate:"required"`
	UserAgent   string `json:"user_agent" validate:"required"`
	IPAddress   string `json:"ip_address" validate:"required"`
	Location    string `json:"location,omitempty"`
}

func (c AddDeviceCommand) CommandType() string { return "AddDevice" }

// RemoveDeviceCommand represents a command to remove a trusted device
type RemoveDeviceCommand struct {
	UserID   string `json:"user_id" validate:"required"`
	DeviceID string `json:"device_id" validate:"required"`
}

func (c RemoveDeviceCommand) CommandType() string { return "RemoveDevice" }

// Verification Commands

// VerifyEmailCommand represents a command to verify email
type VerifyEmailCommand struct {
	UserID           string `json:"user_id" validate:"required"`
	VerificationCode string `json:"verification_code" validate:"required"`
}

func (c VerifyEmailCommand) CommandType() string { return "VerifyEmail" }

// VerifyPhoneCommand represents a command to verify phone
type VerifyPhoneCommand struct {
	UserID           string `json:"user_id" validate:"required"`
	VerificationCode string `json:"verification_code" validate:"required"`
}

func (c VerifyPhoneCommand) CommandType() string { return "VerifyPhone" }

// SendVerificationEmailCommand represents a command to send verification email
type SendVerificationEmailCommand struct {
	UserID string `json:"user_id" validate:"required"`
}

func (c SendVerificationEmailCommand) CommandType() string { return "SendVerificationEmail" }

// SendVerificationSMSCommand represents a command to send verification SMS
type SendVerificationSMSCommand struct {
	UserID string `json:"user_id" validate:"required"`
}

func (c SendVerificationSMSCommand) CommandType() string { return "SendVerificationSMS" }

// Admin Commands

// ChangeUserRoleCommand represents a command to change user role
type ChangeUserRoleCommand struct {
	UserID  string          `json:"user_id" validate:"required"`
	NewRole domain.UserRole `json:"new_role" validate:"required"`
	AdminID string          `json:"admin_id" validate:"required"`
	Reason  string          `json:"reason,omitempty"`
}

func (c ChangeUserRoleCommand) CommandType() string { return "ChangeUserRole" }

// DeleteUserCommand represents a command to delete a user
type DeleteUserCommand struct {
	UserID  string `json:"user_id" validate:"required"`
	AdminID string `json:"admin_id" validate:"required"`
	Reason  string `json:"reason,omitempty"`
}

func (c DeleteUserCommand) CommandType() string { return "DeleteUser" }

// Command Results

// CommandResult represents the result of a command execution
type CommandResult struct {
	Success   bool                   `json:"success"`
	Message   string                 `json:"message,omitempty"`
	Data      interface{}            `json:"data,omitempty"`
	Errors    []string               `json:"errors,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewCommandResult creates a new command result
func NewCommandResult(success bool, message string, data interface{}) *CommandResult {
	return &CommandResult{
		Success:   success,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// AddError adds an error to the command result
func (cr *CommandResult) AddError(err string) {
	if cr.Errors == nil {
		cr.Errors = make([]string, 0)
	}
	cr.Errors = append(cr.Errors, err)
	cr.Success = false
}

// AddMetadata adds metadata to the command result
func (cr *CommandResult) AddMetadata(key string, value interface{}) {
	if cr.Metadata == nil {
		cr.Metadata = make(map[string]interface{})
	}
	cr.Metadata[key] = value
}
