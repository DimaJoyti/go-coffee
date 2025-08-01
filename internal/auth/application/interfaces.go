package application

import (
	"context"
	"errors"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
)

// Cache errors
var (
	ErrCacheKeyNotFound = errors.New("cache key not found")
)

// AuthService defines the interface for authentication service
type AuthService interface {
	// Authentication operations
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	Logout(ctx context.Context, userID string, req *LogoutRequest) (*LogoutResponse, error)

	// Token operations
	RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error)
	ValidateToken(ctx context.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error)
	RevokeToken(ctx context.Context, tokenID string) error

	// User operations
	ChangePassword(ctx context.Context, userID string, req *ChangePasswordRequest) (*ChangePasswordResponse, error)
	GetUserInfo(ctx context.Context, req *GetUserInfoRequest) (*GetUserInfoResponse, error)

	// Session operations
	GetUserSessions(ctx context.Context, req *GetUserSessionsRequest) (*GetUserSessionsResponse, error)
	RevokeSession(ctx context.Context, req *RevokeSessionRequest) (*RevokeSessionResponse, error)
	RevokeAllUserSessions(ctx context.Context, userID string) error

	// Password reset operations
	ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) (*ForgotPasswordResponse, error)
	ResetPassword(ctx context.Context, req *ResetPasswordRequest) (*ResetPasswordResponse, error)

	// Security operations
	GetSecurityEvents(ctx context.Context, req *GetSecurityEventsRequest) (*GetSecurityEventsResponse, error)
	GetTrustedDevices(ctx context.Context, req *GetTrustedDevicesRequest) (*GetTrustedDevicesResponse, error)
	TrustDevice(ctx context.Context, req *TrustDeviceRequest) (*TrustDeviceResponse, error)
	RemoveDevice(ctx context.Context, req *RemoveDeviceRequest) (*RemoveDeviceResponse, error)
}

// MFAService defines the interface for multi-factor authentication operations
type MFAService interface {
	// MFA setup
	EnableMFA(ctx context.Context, req *EnableMFARequest) (*EnableMFAResponse, error)
	DisableMFA(ctx context.Context, req *DisableMFARequest) (*DisableMFAResponse, error)

	// MFA verification
	VerifyMFA(ctx context.Context, req *VerifyMFARequest) (*VerifyMFAResponse, error)

	// Backup codes
	GenerateBackupCodes(ctx context.Context, req *GenerateBackupCodesRequest) (*GenerateBackupCodesResponse, error)
	GetBackupCodes(ctx context.Context, req *GetBackupCodesRequest) (*GetBackupCodesResponse, error)
	UseBackupCode(ctx context.Context, userID, code string) error

	// MFA status
	GetMFAStatus(ctx context.Context, userID string) (*MFAStatusResponse, error)
	IsMFAEnabled(ctx context.Context, userID string) (bool, error)
}

// JWTService defines the interface for JWT token operations
type JWTService interface {
	// Token generation
	GenerateAccessToken(ctx context.Context, user *domain.User, sessionID string) (string, *domain.TokenClaims, error)
	GenerateRefreshToken(ctx context.Context, user *domain.User, sessionID string) (string, *domain.TokenClaims, error)
	GenerateTokenPair(ctx context.Context, user *domain.User, sessionID string) (accessToken, refreshToken string, accessClaims, refreshClaims *domain.TokenClaims, err error)

	// Token validation
	ValidateToken(ctx context.Context, tokenString string) (*domain.TokenClaims, error)
	ValidateAccessToken(ctx context.Context, tokenString string) (*domain.TokenClaims, error)
	ValidateRefreshToken(ctx context.Context, tokenString string) (*domain.TokenClaims, error)

	// Token parsing
	ParseToken(ctx context.Context, tokenString string) (*domain.TokenClaims, error)
	ParseTokenWithoutValidation(ctx context.Context, tokenString string) (*domain.TokenClaims, error)

	// Token utilities
	ExtractTokenFromHeader(authHeader string) (string, error)
	GetTokenExpiry(tokenType domain.TokenType) time.Duration
	IsTokenExpired(claims *domain.TokenClaims) bool
}

// PasswordService defines the interface for password operations
type PasswordService interface {
	// Password hashing
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) error

	// Password validation
	ValidatePassword(password string) error
	ValidatePasswordStrength(password string) error

	// Password policy
	GetPasswordPolicy() *PasswordPolicy
	CheckPasswordPolicy(password string) error
}

// PasswordPolicy represents password policy configuration
type PasswordPolicy struct {
	MinLength        int  `json:"min_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers"`
	RequireSymbols   bool `json:"require_symbols"`
	MaxLength        int  `json:"max_length"`
}

// SecurityService defines the interface for security operations
type SecurityService interface {
	// Security event logging
	LogSecurityEvent(ctx context.Context, userID string, eventType domain.SecurityEventType, severity domain.SecuritySeverity, description string, metadata map[string]string) error
	GetSecurityEvents(ctx context.Context, userID string, limit int) ([]*SecurityEventDTO, error)

	// Rate limiting
	CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
	IncrementRateLimit(ctx context.Context, key string) error

	// IP blocking
	IsIPBlocked(ctx context.Context, ipAddress string) (bool, error)
	BlockIP(ctx context.Context, ipAddress, reason string) error
	UnblockIP(ctx context.Context, ipAddress string) error

	// Account security
	CheckAccountSecurity(ctx context.Context, userID string) error
	LockAccount(ctx context.Context, userID string, reason string) error
	UnlockAccount(ctx context.Context, userID string) error
	IsAccountLocked(ctx context.Context, userID string) (bool, error)

	// Failed login tracking
	TrackFailedLogin(ctx context.Context, email string) error
	ResetFailedLoginCount(ctx context.Context, email string) error
	RecordFailedLogin(ctx context.Context, userID, ipAddress string) error

	// MFA security
	RecordMFAFailure(ctx context.Context, userID string) error

	// Security analysis
	AnalyzeSuspiciousActivity(ctx context.Context, userID, ipAddress, userAgent string) (*SecurityAnalysis, error)
}

// CacheService defines the interface for caching operations
type CacheService interface {
	// Basic cache operations
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (interface{}, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// Cache with specific types
	SetString(ctx context.Context, key, value string, expiration time.Duration) error
	GetString(ctx context.Context, key string) (string, error)
	SetInt(ctx context.Context, key string, value int, expiration time.Duration) error
	GetInt(ctx context.Context, key string) (int, error)

	// Increment operations
	Increment(ctx context.Context, key string) (int64, error)
	IncrementWithExpiry(ctx context.Context, key string, expiration time.Duration) (int64, error)
	Decrement(ctx context.Context, key string) (int64, error)

	// Cache operations for auth
	SetUserSession(ctx context.Context, sessionID string, session *domain.Session, expiration time.Duration) error
	GetUserSession(ctx context.Context, sessionID string) (*domain.Session, error)
	DeleteUserSession(ctx context.Context, sessionID string) error

	// Blacklist operations
	AddToBlacklist(ctx context.Context, tokenID string, expiration time.Duration) error
	IsBlacklisted(ctx context.Context, tokenID string) (bool, error)
	RemoveFromBlacklist(ctx context.Context, tokenID string) error
}

// ValidationService defines the interface for validation operations
type ValidationService interface {
	// Request validation
	ValidateRegisterRequest(req *RegisterRequest) error
	ValidateLoginRequest(req *LoginRequest) error
	ValidateChangePasswordRequest(req *ChangePasswordRequest) error
	ValidateRefreshTokenRequest(req *RefreshTokenRequest) error

	// Data validation
	ValidateEmail(email string) error
	ValidateUserID(userID string) error
	ValidateSessionID(sessionID string) error
	ValidateTokenString(token string) error

	// Business rule validation
	ValidateUserRegistration(ctx context.Context, req *RegisterRequest) error
	ValidateUserLogin(ctx context.Context, user *domain.User, req *LoginRequest) error
	ValidatePasswordChange(ctx context.Context, user *domain.User, req *ChangePasswordRequest) error
}

// NotificationService defines the interface for notification operations
type NotificationService interface {
	// Security notifications
	SendSecurityAlert(ctx context.Context, userID string, eventType domain.SecurityEventType, details map[string]string) error
	SendLoginNotification(ctx context.Context, userID string, deviceInfo *domain.DeviceInfo, ipAddress string) error
	SendPasswordChangeNotification(ctx context.Context, userID string) error
	SendAccountLockNotification(ctx context.Context, userID string, reason string) error

	// Email notifications
	SendWelcomeEmail(ctx context.Context, userID string, email string) error
	SendPasswordResetEmail(ctx context.Context, userID string, email string, resetToken string) error
}

// AuditService defines the interface for audit operations
type AuditService interface {
	// Audit logging
	LogUserAction(ctx context.Context, userID string, action string, details map[string]interface{}) error
	LogSystemEvent(ctx context.Context, event string, details map[string]interface{}) error
	LogSecurityEvent(ctx context.Context, userID string, event string, severity string, details map[string]interface{}) error

	// Audit queries
	GetUserAuditLog(ctx context.Context, userID string, limit int) ([]interface{}, error)
	GetSystemAuditLog(ctx context.Context, limit int) ([]interface{}, error)
	GetSecurityAuditLog(ctx context.Context, limit int) ([]interface{}, error)
}
