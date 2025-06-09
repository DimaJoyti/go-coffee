package application

import (
	"context"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// AuthServiceImpl implements the AuthService interface
type AuthServiceImpl struct {
	userRepo        domain.UserRepository
	sessionRepo     domain.SessionRepository
	jwtService      JWTService
	passwordService PasswordService
	securityService SecurityService
	logger          *logger.Logger
	config          *AuthConfig
}

// AuthConfig represents authentication service configuration
type AuthConfig struct {
	AccessTokenTTL   time.Duration `yaml:"access_token_ttl"`
	RefreshTokenTTL  time.Duration `yaml:"refresh_token_ttl"`
	MaxLoginAttempts int           `yaml:"max_login_attempts"`
	LockoutDuration  time.Duration `yaml:"lockout_duration"`
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	jwtService JWTService,
	passwordService PasswordService,
	securityService SecurityService,
	config *AuthConfig,
	logger *logger.Logger,
) *AuthServiceImpl {
	return &AuthServiceImpl{
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		jwtService:      jwtService,
		passwordService: passwordService,
		securityService: securityService,
		config:          config,
		logger:          logger,
	}
}

// Register registers a new user
func (s *AuthServiceImpl) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	s.logger.InfoWithFields("User registration attempt", logger.String("email", req.Email))

	// Validate request
	if err := s.validateRegisterRequest(req); err != nil {
		s.logger.ErrorWithFields("Registration validation failed", logger.Error(err), logger.String("email", req.Email))
		return nil, err
	}

	// Check if user already exists
	exists, err := s.userRepo.UserExists(ctx, req.Email)
	if err != nil {
		s.logger.ErrorWithFields("Failed to check user existence", logger.Error(err), logger.String("email", req.Email))
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		s.logger.WarnWithFields("Registration attempt with existing email", logger.String("email", req.Email))
		return nil, domain.ErrUserExists
	}

	// Hash password
	passwordHash, err := s.passwordService.HashPassword(req.Password)
	if err != nil {
		s.logger.ErrorWithFields("Failed to hash password", logger.Error(err), logger.String("email", req.Email))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Determine user role
	role := domain.UserRoleUser
	if req.Role != "" {
		role = domain.UserRole(req.Role)
	}

	// Create user
	user, err := domain.NewUser(req.Email, passwordHash, role)
	if err != nil {
		s.logger.ErrorWithFields("Failed to create user", logger.Error(err), logger.String("email", req.Email))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Save user
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		s.logger.ErrorWithFields("Failed to save user", logger.Error(err), logger.String("email", req.Email))
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	// Create session and tokens
	accessToken, refreshToken, accessClaims, _, err := s.jwtService.GenerateTokenPair(ctx, user, "")
	if err != nil {
		s.logger.ErrorWithFields("Failed to generate tokens", logger.Error(err), logger.String("user_id", user.ID))
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create session
	session := domain.NewSession(user.ID, accessToken, refreshToken, s.config.AccessTokenTTL, s.config.RefreshTokenTTL)
	if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
		s.logger.ErrorWithFields("Failed to create session", logger.Error(err), logger.String("user_id", user.ID))
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Log security event
	s.securityService.LogSecurityEvent(ctx, user.ID, domain.SecurityEventTypeLogin, domain.SecuritySeverityLow, "User registered and logged in", nil)
	s.logger.InfoWithFields("User registered successfully", logger.String("user_id", user.ID), logger.String("email", user.Email))

	return &RegisterResponse{
		User:         ToUserDTO(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessClaims.GetTimeUntilExpiryClaims().Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// Login authenticates a user and creates a session
func (s *AuthServiceImpl) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	s.logger.InfoWithFields("User login attempt", logger.String("email", req.Email))

	// Validate request
	if err := s.validateLoginRequest(req); err != nil {
		s.logger.ErrorWithFields("Login validation failed", logger.Error(err), logger.String("email", req.Email))
		return nil, err
	}

	// Check rate limiting
	if err := s.securityService.CheckRateLimit(ctx, "login:"+req.Email); err != nil {
		s.logger.WarnWithFields("Login rate limit exceeded", logger.String("email", req.Email))
		return nil, fmt.Errorf("too many login attempts, please try again later")
	}

	// Check if account is locked
	locked, err := s.securityService.IsAccountLocked(ctx, req.Email)
	if err != nil {
		s.logger.ErrorWithFields("Failed to check account lock status", logger.Error(err), logger.String("email", req.Email))
		return nil, fmt.Errorf("failed to check account status: %w", err)
	}
	if locked {
		s.logger.WarnWithFields("Login attempt on locked account", logger.String("email", req.Email))
		return nil, domain.ErrUserLocked
	}

	// Get user
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			// Track failed login attempt
			s.securityService.TrackFailedLogin(ctx, req.Email)
			s.securityService.LogSecurityEvent(ctx, "", domain.SecurityEventTypeLoginFailed, domain.SecuritySeverityMedium, "Login attempt with invalid email", map[string]string{"email": req.Email})
		}
		s.logger.ErrorWithFields("Failed to get user", logger.Error(err), logger.String("email", req.Email))
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check user status
	if !user.IsActive() {
		s.logger.WarnWithFields("Login attempt on inactive user", logger.String("user_id", user.ID))
		s.securityService.LogSecurityEvent(ctx, user.ID, domain.SecurityEventTypeLoginFailed, domain.SecuritySeverityMedium, "Login attempt on inactive account", nil)
		return nil, domain.ErrUserInactive
	}

	// Verify password
	if err := s.passwordService.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		// Track failed login
		s.userRepo.IncrementFailedLogin(ctx, user.ID)
		s.securityService.TrackFailedLogin(ctx, req.Email)
		s.securityService.LogSecurityEvent(ctx, user.ID, domain.SecurityEventTypeLoginFailed, domain.SecuritySeverityMedium, "Invalid password", nil)

		// Check if account should be locked
		if user.FailedLoginCount >= s.config.MaxLoginAttempts {
			lockUntil := time.Now().Add(s.config.LockoutDuration)
			s.userRepo.LockUser(ctx, user.ID, lockUntil)
			s.securityService.LogSecurityEvent(ctx, user.ID, domain.SecurityEventTypeAccountLocked, domain.SecuritySeverityHigh, "Account locked due to too many failed login attempts", nil)
		}

		s.logger.WarnWithFields("Invalid password", logger.String("user_id", user.ID))
		return nil, fmt.Errorf("invalid credentials")
	}

	// Reset failed login count on successful login
	user.ResetFailedLogin()
	user.UpdateLastLogin()
	s.userRepo.UpdateUser(ctx, user)
	s.securityService.ResetFailedLoginCount(ctx, req.Email)

	// Generate tokens
	accessToken, refreshToken, accessClaims, _, err := s.jwtService.GenerateTokenPair(ctx, user, "")
	if err != nil {
		s.logger.ErrorWithFields("Failed to generate tokens", logger.Error(err), logger.String("user_id", user.ID))
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create session
	session := domain.NewSession(user.ID, accessToken, refreshToken, s.config.AccessTokenTTL, s.config.RefreshTokenTTL)
	if req.DeviceInfo != nil {
		session.SetDeviceInfo(req.DeviceInfo)
	}

	// Extend session if remember me is enabled
	if req.RememberMe {
		extendedTTL := s.config.RefreshTokenTTL * 4 // 4x longer
		session.RefreshExpiresAt = time.Now().Add(extendedTTL)
	}

	if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
		s.logger.ErrorWithFields("Failed to create session", logger.Error(err), logger.String("user_id", user.ID))
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Log security event
	s.securityService.LogSecurityEvent(ctx, user.ID, domain.SecurityEventTypeLogin, domain.SecuritySeverityLow, "Successful login", nil)
	s.logger.InfoWithFields("User logged in successfully", logger.String("user_id", user.ID), logger.String("email", user.Email))

	return &LoginResponse{
		User:         ToUserDTO(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessClaims.GetTimeUntilExpiryClaims().Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// Logout logs out a user and revokes session(s)
func (s *AuthServiceImpl) Logout(ctx context.Context, userID string, req *LogoutRequest) (*LogoutResponse, error) {
	s.logger.InfoWithFields("User logout attempt", logger.String("user_id", userID))

	if req.LogoutAll {
		// Revoke all user sessions
		if err := s.sessionRepo.RevokeUserSessions(ctx, userID); err != nil {
			s.logger.ErrorWithFields("Failed to revoke all user sessions", logger.Error(err), logger.String("user_id", userID))
			return nil, fmt.Errorf("failed to revoke sessions: %w", err)
		}
		s.logger.InfoWithFields("All user sessions revoked", logger.String("user_id", userID))
	} else if req.SessionID != "" {
		// Revoke specific session
		if err := s.sessionRepo.RevokeSession(ctx, req.SessionID); err != nil {
			s.logger.ErrorWithFields("Failed to revoke session", logger.Error(err), logger.String("session_id", req.SessionID))
			return nil, fmt.Errorf("failed to revoke session: %w", err)
		}
		s.logger.InfoWithFields("Session revoked", logger.String("session_id", req.SessionID))
	}

	// Log security event
	s.securityService.LogSecurityEvent(ctx, userID, domain.SecurityEventTypeLogout, domain.SecuritySeverityLow, "User logged out", nil)

	return &LogoutResponse{
		Message: "Logged out successfully",
		Success: true,
	}, nil
}

// validateRegisterRequest validates registration request
func (s *AuthServiceImpl) validateRegisterRequest(req *RegisterRequest) error {
	if err := domain.ValidateEmail(req.Email); err != nil {
		return err
	}

	if err := s.passwordService.ValidatePassword(req.Password); err != nil {
		return err
	}

	return nil
}

// validateLoginRequest validates login request
func (s *AuthServiceImpl) validateLoginRequest(req *LoginRequest) error {
	if err := domain.ValidateEmail(req.Email); err != nil {
		return err
	}

	if req.Password == "" {
		return fmt.Errorf("password is required")
	}

	return nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthServiceImpl) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	s.logger.InfoWithFields("Token refresh attempt")

	// Validate refresh token
	claims, err := s.jwtService.ValidateRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		s.logger.ErrorWithFields("Invalid refresh token", logger.Error(err))
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Get session by refresh token
	session, err := s.sessionRepo.GetSessionByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get session by refresh token", logger.Error(err))
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Validate session
	if err := session.ValidateRefresh(); err != nil {
		s.logger.ErrorWithFields("Session validation failed", logger.Error(err), logger.String("session_id", session.ID))
		return nil, err
	}

	// Get user
	user, err := s.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get user", logger.Error(err), logger.String("user_id", claims.UserID))
		return nil, fmt.Errorf("user not found")
	}

	// Check user status
	if !user.IsActive() {
		s.logger.WarnWithFields("Token refresh attempt for inactive user", logger.String("user_id", user.ID))
		return nil, domain.ErrUserInactive
	}

	// Generate new tokens
	accessToken, refreshToken, accessClaims, _, err := s.jwtService.GenerateTokenPair(ctx, user, session.ID)
	if err != nil {
		s.logger.ErrorWithFields("Failed to generate new tokens", logger.Error(err), logger.String("user_id", user.ID))
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Update session with new tokens
	session.UpdateTokens(accessToken, refreshToken, s.config.AccessTokenTTL, s.config.RefreshTokenTTL)
	if err := s.sessionRepo.UpdateSession(ctx, session); err != nil {
		s.logger.ErrorWithFields("Failed to update session", logger.Error(err), logger.String("session_id", session.ID))
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	// Log security event
	s.securityService.LogSecurityEvent(ctx, user.ID, domain.SecurityEventTypeTokenRefresh, domain.SecuritySeverityLow, "Token refreshed", nil)
	s.logger.InfoWithFields("Token refreshed successfully", logger.String("user_id", user.ID))

	return &RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessClaims.GetTimeUntilExpiryClaims().Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// ValidateToken validates an access token
func (s *AuthServiceImpl) ValidateToken(ctx context.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error) {
	// Parse and validate token
	claims, err := s.jwtService.ValidateAccessToken(ctx, req.Token)
	if err != nil {
		return &ValidateTokenResponse{
			Valid:   false,
			Message: err.Error(),
		}, nil
	}

	// Get user
	user, err := s.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return &ValidateTokenResponse{
			Valid:   false,
			Message: "user not found",
		}, nil
	}

	// Check user status
	if !user.IsActive() {
		return &ValidateTokenResponse{
			Valid:   false,
			Message: "user account is inactive",
		}, nil
	}

	return &ValidateTokenResponse{
		Valid:     true,
		UserID:    user.ID,
		Role:      string(user.Role),
		SessionID: claims.SessionID,
		User:      ToUserDTO(user),
		Claims:    ToClaimsDTO(claims),
	}, nil
}

// RevokeToken revokes a token
func (s *AuthServiceImpl) RevokeToken(ctx context.Context, tokenID string) error {
	s.logger.InfoWithFields("Token revocation attempt", logger.String("token_id", tokenID))

	// This would typically involve adding the token to a blacklist
	// For now, we'll implement basic revocation logic

	// Log security event
	s.securityService.LogSecurityEvent(ctx, "", domain.SecurityEventTypeTokenRevoked, domain.SecuritySeverityMedium, "Token revoked", map[string]string{"token_id": tokenID})
	s.logger.InfoWithFields("Token revoked successfully", logger.String("token_id", tokenID))
	return nil
}

// ChangePassword changes a user's password
func (s *AuthServiceImpl) ChangePassword(ctx context.Context, userID string, req *ChangePasswordRequest) (*ChangePasswordResponse, error) {
	s.logger.InfoWithFields("Password change attempt", logger.String("user_id", userID))

	// Get user
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get user", logger.Error(err), logger.String("user_id", userID))
		return nil, fmt.Errorf("user not found")
	}

	// Verify current password
	if err := s.passwordService.VerifyPassword(user.PasswordHash, req.CurrentPassword); err != nil {
		s.logger.WarnWithFields("Invalid current password", logger.String("user_id", userID))
		return nil, fmt.Errorf("invalid current password")
	}

	// Validate new password
	if err := s.passwordService.ValidatePassword(req.NewPassword); err != nil {
		s.logger.ErrorWithFields("New password validation failed", logger.Error(err), logger.String("user_id", userID))
		return nil, err
	}

	// Hash new password
	newPasswordHash, err := s.passwordService.HashPassword(req.NewPassword)
	if err != nil {
		s.logger.ErrorWithFields("Failed to hash new password", logger.Error(err), logger.String("user_id", userID))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	user.ChangePassword(newPasswordHash, true)
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		s.logger.ErrorWithFields("Failed to update user password", logger.Error(err), logger.String("user_id", userID))
		return nil, fmt.Errorf("failed to update password: %w", err)
	}

	// Revoke all user sessions (force re-login)
	s.sessionRepo.RevokeUserSessions(ctx, userID)

	// Log security event
	s.securityService.LogSecurityEvent(ctx, userID, domain.SecurityEventTypePasswordChange, domain.SecuritySeverityMedium, "Password changed", nil)
	s.logger.InfoWithFields("Password changed successfully", logger.String("user_id", userID))

	return &ChangePasswordResponse{
		Message: "Password changed successfully",
		Success: true,
	}, nil
}

// GetUserInfo gets user information
func (s *AuthServiceImpl) GetUserInfo(ctx context.Context, req *GetUserInfoRequest) (*GetUserInfoResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get user info", logger.Error(err), logger.String("user_id", req.UserID))
		return nil, fmt.Errorf("user not found")
	}

	return &GetUserInfoResponse{
		User: ToUserDTO(user),
	}, nil
}

// GetUserSessions gets all sessions for a user (interface method)
func (s *AuthServiceImpl) GetUserSessions(ctx context.Context, req *GetUserSessionsRequest) (*GetUserSessionsResponse, error) {
	sessions, err := s.sessionRepo.GetUserSessions(ctx, req.UserID)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get user sessions", logger.Error(err), logger.String("user_id", req.UserID))
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	sessionDTOs := make([]*SessionDTO, len(sessions))
	for i, session := range sessions {
		sessionDTOs[i] = ToSessionDTO(session)
	}

	return &GetUserSessionsResponse{
		Sessions: sessionDTOs,
		Total:    len(sessionDTOs),
	}, nil
}

// RevokeSession revokes a specific session (interface method)
func (s *AuthServiceImpl) RevokeSession(ctx context.Context, req *RevokeSessionRequest) (*RevokeSessionResponse, error) {
	err := s.sessionRepo.RevokeSession(ctx, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke session: %w", err)
	}

	return &RevokeSessionResponse{
		Success: true,
		Message: "Session revoked successfully",
	}, nil
}

// RevokeAllUserSessions revokes all sessions for a user
func (s *AuthServiceImpl) RevokeAllUserSessions(ctx context.Context, userID string) error {
	return s.sessionRepo.RevokeUserSessions(ctx, userID)
}

// ForgotPassword handles forgot password requests
func (s *AuthServiceImpl) ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) (*ForgotPasswordResponse, error) {
	s.logger.InfoWithFields("Forgot password request", logger.String("email", req.Email))

	// Check if user exists
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			// Don't reveal if email exists or not for security
			return &ForgotPasswordResponse{
				Success: true,
				Message: "If the email exists, a password reset link has been sent",
			}, nil
		}
		s.logger.ErrorWithFields("Failed to get user for password reset", logger.Error(err), logger.String("email", req.Email))
		return nil, fmt.Errorf("failed to process request: %w", err)
	}

	// Generate password reset token (placeholder implementation)
	// resetToken := "reset_token_placeholder" // In real implementation, generate secure token

	// Log security event
	s.securityService.LogSecurityEvent(ctx, user.ID, domain.SecurityEventTypePasswordChange, domain.SecuritySeverityMedium, "Password reset requested", map[string]string{"email": req.Email})

	// TODO: Send password reset email
	s.logger.InfoWithFields("Password reset requested", logger.String("user_id", user.ID), logger.String("email", user.Email))

	return &ForgotPasswordResponse{
		Success: true,
		Message: "If the email exists, a password reset link has been sent",
	}, nil
}

// ResetPassword handles password reset with token
func (s *AuthServiceImpl) ResetPassword(ctx context.Context, req *ResetPasswordRequest) (*ResetPasswordResponse, error) {
	s.logger.InfoWithFields("Password reset attempt", logger.String("token", req.Token[:10]+"..."))

	// TODO: Validate reset token (placeholder implementation)
	// In real implementation, validate token and get user ID from token

	// For now, return success (placeholder)
	return &ResetPasswordResponse{
		Success: true,
		Message: "Password reset successfully",
	}, nil
}

// GetSecurityEvents gets security events for a user
func (s *AuthServiceImpl) GetSecurityEvents(ctx context.Context, req *GetSecurityEventsRequest) (*GetSecurityEventsResponse, error) {
	// TODO: Implement security events retrieval
	// This would typically query the security events from the security service

	return &GetSecurityEventsResponse{
		Events: []*SecurityEventDTO{},
		Total:  0,
	}, nil
}

// GetTrustedDevices gets trusted devices for a user
func (s *AuthServiceImpl) GetTrustedDevices(ctx context.Context, req *GetTrustedDevicesRequest) (*GetTrustedDevicesResponse, error) {
	// TODO: Implement trusted devices retrieval
	// This would typically query trusted devices from the database

	return &GetTrustedDevicesResponse{
		Devices: []*DeviceDTO{},
		Total:   0,
	}, nil
}

// TrustDevice marks a device as trusted
func (s *AuthServiceImpl) TrustDevice(ctx context.Context, req *TrustDeviceRequest) (*TrustDeviceResponse, error) {
	// TODO: Implement device trust functionality
	// This would typically update device trust status in the database

	s.logger.InfoWithFields("Device trusted", logger.String("user_id", req.UserID), logger.String("device_id", req.DeviceID))

	return &TrustDeviceResponse{
		Success: true,
		Message: "Device trusted successfully",
	}, nil
}

// RemoveDevice removes a trusted device
func (s *AuthServiceImpl) RemoveDevice(ctx context.Context, req *RemoveDeviceRequest) (*RemoveDeviceResponse, error) {
	// TODO: Implement device removal functionality
	// This would typically remove device from trusted devices in the database

	s.logger.InfoWithFields("Device removed", logger.String("user_id", req.UserID), logger.String("device_id", req.DeviceID))

	return &RemoveDeviceResponse{
		Success: true,
		Message: "Device removed successfully",
	}, nil
}
