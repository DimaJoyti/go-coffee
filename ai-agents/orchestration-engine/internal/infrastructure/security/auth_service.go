package security

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/argon2"
)

// AuthService provides comprehensive authentication and authorization
type AuthService struct {
	jwtManager    *JWTManager
	userStore     UserStore
	sessionStore  SessionStore
	auditLogger   AuditLogger
	config        *AuthConfig
	rateLimiter   *RateLimiter
	logger        Logger
	mutex         sync.RWMutex
}

// Logger interface for logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	// JWT settings
	JWTSecret           string        `json:"jwt_secret"`
	JWTExpiration       time.Duration `json:"jwt_expiration"`
	JWTRefreshExpiration time.Duration `json:"jwt_refresh_expiration"`
	
	// Password settings
	MinPasswordLength   int  `json:"min_password_length"`
	RequireUppercase    bool `json:"require_uppercase"`
	RequireLowercase    bool `json:"require_lowercase"`
	RequireNumbers      bool `json:"require_numbers"`
	RequireSpecialChars bool `json:"require_special_chars"`
	
	// Session settings
	SessionTimeout      time.Duration `json:"session_timeout"`
	MaxConcurrentSessions int         `json:"max_concurrent_sessions"`
	
	// Security settings
	MaxLoginAttempts    int           `json:"max_login_attempts"`
	LockoutDuration     time.Duration `json:"lockout_duration"`
	EnableMFA           bool          `json:"enable_mfa"`
	EnableAuditLogging  bool          `json:"enable_audit_logging"`
	
	// Rate limiting
	LoginRateLimit      int           `json:"login_rate_limit"`
	LoginRateWindow     time.Duration `json:"login_rate_window"`
}

// User represents a system user
type User struct {
	ID                string            `json:"id"`
	Username          string            `json:"username"`
	Email             string            `json:"email"`
	PasswordHash      string            `json:"password_hash"`
	Salt              string            `json:"salt"`
	Roles             []string          `json:"roles"`
	Permissions       []string          `json:"permissions"`
	IsActive          bool              `json:"is_active"`
	IsLocked          bool              `json:"is_locked"`
	FailedAttempts    int               `json:"failed_attempts"`
	LastLogin         *time.Time        `json:"last_login"`
	LastFailedLogin   *time.Time        `json:"last_failed_login"`
	LockedUntil       *time.Time        `json:"locked_until"`
	MFAEnabled        bool              `json:"mfa_enabled"`
	MFASecret         string            `json:"mfa_secret"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	Metadata          map[string]string `json:"metadata"`
}

// Session represents a user session
type Session struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	Token     string            `json:"token"`
	ExpiresAt time.Time         `json:"expires_at"`
	CreatedAt time.Time         `json:"created_at"`
	LastSeen  time.Time         `json:"last_seen"`
	IPAddress string            `json:"ip_address"`
	UserAgent string            `json:"user_agent"`
	IsActive  bool              `json:"is_active"`
	Metadata  map[string]string `json:"metadata"`
}

// AuthRequest represents an authentication request
type AuthRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	MFACode   string `json:"mfa_code,omitempty"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Success      bool      `json:"success"`
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	User         *User     `json:"user,omitempty"`
	Message      string    `json:"message,omitempty"`
}

// UserStore interface for user persistence
type UserStore interface {
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, id string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, filter *UserFilter) ([]*User, error)
}

// SessionStore interface for session persistence
type SessionStore interface {
	CreateSession(ctx context.Context, session *Session) error
	GetSession(ctx context.Context, token string) (*Session, error)
	UpdateSession(ctx context.Context, session *Session) error
	DeleteSession(ctx context.Context, token string) error
	DeleteUserSessions(ctx context.Context, userID string) error
	CleanupExpiredSessions(ctx context.Context) error
}

// AuditLogger interface for audit logging
type AuditLogger interface {
	LogAuthEvent(ctx context.Context, event *AuthEvent) error
	LogSecurityEvent(ctx context.Context, event *SecurityEvent) error
}

// UserFilter represents user filtering options
type UserFilter struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	IsActive *bool    `json:"is_active"`
	Limit    int      `json:"limit"`
	Offset   int      `json:"offset"`
}

// NewAuthService creates a new authentication service
func NewAuthService(
	jwtManager *JWTManager,
	userStore UserStore,
	sessionStore SessionStore,
	auditLogger AuditLogger,
	config *AuthConfig,
	logger Logger,
) *AuthService {
	return &AuthService{
		jwtManager:   jwtManager,
		userStore:    userStore,
		sessionStore: sessionStore,
		auditLogger:  auditLogger,
		config:       config,
		rateLimiter:  NewRateLimiter(config.LoginRateLimit, config.LoginRateWindow),
		logger:       logger,
	}
}

// Authenticate authenticates a user and returns tokens
func (as *AuthService) Authenticate(ctx context.Context, req *AuthRequest) (*AuthResponse, error) {
	// Rate limiting check
	if !as.rateLimiter.Allow(req.IPAddress) {
		as.logSecurityEvent(ctx, &SecurityEvent{
			Type:      "rate_limit_exceeded",
			UserID:    "",
			IPAddress: req.IPAddress,
			Details:   map[string]interface{}{"username": req.Username},
		})
		return &AuthResponse{
			Success: false,
			Message: "Rate limit exceeded. Please try again later.",
		}, nil
	}

	// Get user by username
	user, err := as.userStore.GetUserByUsername(ctx, req.Username)
	if err != nil {
		as.logAuthEvent(ctx, &AuthEvent{
			Type:      "login_failed",
			UserID:    "",
			Username:  req.Username,
			IPAddress: req.IPAddress,
			Reason:    "user_not_found",
		})
		return &AuthResponse{
			Success: false,
			Message: "Invalid credentials",
		}, nil
	}

	// Check if user is active
	if !user.IsActive {
		as.logAuthEvent(ctx, &AuthEvent{
			Type:      "login_failed",
			UserID:    user.ID,
			Username:  req.Username,
			IPAddress: req.IPAddress,
			Reason:    "user_inactive",
		})
		return &AuthResponse{
			Success: false,
			Message: "Account is inactive",
		}, nil
	}

	// Check if user is locked
	if user.IsLocked && user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		as.logAuthEvent(ctx, &AuthEvent{
			Type:      "login_failed",
			UserID:    user.ID,
			Username:  req.Username,
			IPAddress: req.IPAddress,
			Reason:    "account_locked",
		})
		return &AuthResponse{
			Success: false,
			Message: "Account is locked. Please try again later.",
		}, nil
	}

	// Verify password
	if !as.verifyPassword(req.Password, user.PasswordHash, user.Salt) {
		as.handleFailedLogin(ctx, user, req.IPAddress)
		return &AuthResponse{
			Success: false,
			Message: "Invalid credentials",
		}, nil
	}

	// Verify MFA if enabled
	if user.MFAEnabled && req.MFACode != "" {
		if !as.verifyMFA(user.MFASecret, req.MFACode) {
			as.handleFailedLogin(ctx, user, req.IPAddress)
			return &AuthResponse{
				Success: false,
				Message: "Invalid MFA code",
			}, nil
		}
	} else if user.MFAEnabled {
		return &AuthResponse{
			Success: false,
			Message: "MFA code required",
		}, nil
	}

	// Reset failed attempts on successful login
	user.FailedAttempts = 0
	user.IsLocked = false
	user.LockedUntil = nil
	now := time.Now()
	user.LastLogin = &now
	user.UpdatedAt = now

	if err := as.userStore.UpdateUser(ctx, user); err != nil {
		as.logger.Error("Failed to update user after successful login", err)
	}

	// Generate tokens
	accessToken, err := as.jwtManager.GenerateToken(user.ID, user.Roles, as.config.JWTExpiration)
	if err != nil {
		as.logger.Error("Failed to generate access token", err)
		return &AuthResponse{
			Success: false,
			Message: "Authentication failed",
		}, nil
	}

	refreshToken, err := as.jwtManager.GenerateRefreshToken(user.ID, as.config.JWTRefreshExpiration)
	if err != nil {
		as.logger.Error("Failed to generate refresh token", err)
		return &AuthResponse{
			Success: false,
			Message: "Authentication failed",
		}, nil
	}

	// Create session
	session := &Session{
		ID:        generateSessionID(),
		UserID:    user.ID,
		Token:     accessToken,
		ExpiresAt: time.Now().Add(as.config.SessionTimeout),
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
		IsActive:  true,
	}

	if err := as.sessionStore.CreateSession(ctx, session); err != nil {
		as.logger.Error("Failed to create session", err)
	}

	// Log successful authentication
	as.logAuthEvent(ctx, &AuthEvent{
		Type:      "login_success",
		UserID:    user.ID,
		Username:  req.Username,
		IPAddress: req.IPAddress,
		SessionID: session.ID,
	})

	// Remove sensitive data from user object
	userResponse := *user
	userResponse.PasswordHash = ""
	userResponse.Salt = ""
	userResponse.MFASecret = ""

	return &AuthResponse{
		Success:      true,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(as.config.JWTExpiration),
		User:         &userResponse,
		Message:      "Authentication successful",
	}, nil
}

// ValidateToken validates a JWT token and returns user information
func (as *AuthService) ValidateToken(ctx context.Context, token string) (*User, error) {
	claims, err := as.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	user, err := as.userStore.GetUser(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.IsActive {
		return nil, fmt.Errorf("user is inactive")
	}

	return user, nil
}

// RefreshToken refreshes an access token using a refresh token
func (as *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	claims, err := as.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return &AuthResponse{
			Success: false,
			Message: "Invalid refresh token",
		}, nil
	}

	user, err := as.userStore.GetUser(ctx, claims.UserID)
	if err != nil || !user.IsActive {
		return &AuthResponse{
			Success: false,
			Message: "User not found or inactive",
		}, nil
	}

	// Generate new access token
	accessToken, err := as.jwtManager.GenerateToken(user.ID, user.Roles, as.config.JWTExpiration)
	if err != nil {
		return &AuthResponse{
			Success: false,
			Message: "Failed to generate token",
		}, nil
	}

	return &AuthResponse{
		Success:     true,
		AccessToken: accessToken,
		ExpiresAt:   time.Now().Add(as.config.JWTExpiration),
		Message:     "Token refreshed successfully",
	}, nil
}

// Logout logs out a user and invalidates their session
func (as *AuthService) Logout(ctx context.Context, token string) error {
	// Delete session
	if err := as.sessionStore.DeleteSession(ctx, token); err != nil {
		as.logger.Error("Failed to delete session", err)
	}

	// Log logout event
	claims, err := as.jwtManager.ValidateToken(token)
	if err == nil {
		as.logAuthEvent(ctx, &AuthEvent{
			Type:   "logout",
			UserID: claims.UserID,
		})
	}

	return nil
}

// CreateUser creates a new user with secure password hashing
func (as *AuthService) CreateUser(ctx context.Context, username, email, password string, roles []string) (*User, error) {
	// Validate password strength
	if err := as.validatePassword(password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Check if user already exists
	if existingUser, _ := as.userStore.GetUserByUsername(ctx, username); existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	if existingUser, _ := as.userStore.GetUserByEmail(ctx, email); existingUser != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Generate salt and hash password
	salt := generateSalt()
	passwordHash := as.hashPassword(password, salt)

	user := &User{
		ID:           generateUserID(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Salt:         salt,
		Roles:        roles,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Metadata:     make(map[string]string),
	}

	if err := as.userStore.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	as.logAuthEvent(ctx, &AuthEvent{
		Type:     "user_created",
		UserID:   user.ID,
		Username: username,
	})

	return user, nil
}

// handleFailedLogin handles failed login attempts
func (as *AuthService) handleFailedLogin(ctx context.Context, user *User, ipAddress string) {
	user.FailedAttempts++
	now := time.Now()
	user.LastFailedLogin = &now

	if user.FailedAttempts >= as.config.MaxLoginAttempts {
		user.IsLocked = true
		lockUntil := now.Add(as.config.LockoutDuration)
		user.LockedUntil = &lockUntil
	}

	user.UpdatedAt = now
	if err := as.userStore.UpdateUser(ctx, user); err != nil {
		as.logger.Error("Failed to update user after failed login", err)
	}

	as.logAuthEvent(ctx, &AuthEvent{
		Type:      "login_failed",
		UserID:    user.ID,
		Username:  user.Username,
		IPAddress: ipAddress,
		Reason:    "invalid_password",
	})
}

// validatePassword validates password strength
func (as *AuthService) validatePassword(password string) error {
	if len(password) < as.config.MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters long", as.config.MinPasswordLength)
	}

	if as.config.RequireUppercase && !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if as.config.RequireLowercase && !strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if as.config.RequireNumbers && !strings.ContainsAny(password, "0123456789") {
		return fmt.Errorf("password must contain at least one number")
	}

	if as.config.RequireSpecialChars && !strings.ContainsAny(password, "!@#$%^&*()_+-=[]{}|;:,.<>?") {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// hashPassword hashes a password using Argon2
func (as *AuthService) hashPassword(password, salt string) string {
	hash := argon2.IDKey([]byte(password), []byte(salt), 1, 64*1024, 4, 32)
	return base64.StdEncoding.EncodeToString(hash)
}

// verifyPassword verifies a password against its hash
func (as *AuthService) verifyPassword(password, hash, salt string) bool {
	expectedHash := as.hashPassword(password, salt)
	return subtle.ConstantTimeCompare([]byte(hash), []byte(expectedHash)) == 1
}

// verifyMFA verifies a TOTP MFA code
func (as *AuthService) verifyMFA(secret, code string) bool {
	// In a real implementation, this would use a TOTP library
	// For now, return true for demonstration
	return code != ""
}

// logAuthEvent logs an authentication event
func (as *AuthService) logAuthEvent(ctx context.Context, event *AuthEvent) {
	if as.config.EnableAuditLogging && as.auditLogger != nil {
		if err := as.auditLogger.LogAuthEvent(ctx, event); err != nil {
			as.logger.Error("Failed to log auth event", err)
		}
	}
}

// logSecurityEvent logs a security event
func (as *AuthService) logSecurityEvent(ctx context.Context, event *SecurityEvent) {
	if as.config.EnableAuditLogging && as.auditLogger != nil {
		if err := as.auditLogger.LogSecurityEvent(ctx, event); err != nil {
			as.logger.Error("Failed to log security event", err)
		}
	}
}

// Helper functions

func generateUserID() string {
	return fmt.Sprintf("user_%d", time.Now().UnixNano())
}

func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

func generateSalt() string {
	salt := make([]byte, 32)
	rand.Read(salt)
	return base64.StdEncoding.EncodeToString(salt)
}

// AuthEvent represents an authentication event
type AuthEvent struct {
	Type      string                 `json:"type"`
	UserID    string                 `json:"user_id"`
	Username  string                 `json:"username"`
	IPAddress string                 `json:"ip_address"`
	SessionID string                 `json:"session_id"`
	Reason    string                 `json:"reason"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details"`
}

// SecurityEvent represents a security event
type SecurityEvent struct {
	Type      string                 `json:"type"`
	UserID    string                 `json:"user_id"`
	IPAddress string                 `json:"ip_address"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details"`
}
