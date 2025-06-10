package accounts

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/redis"
)

// Service defines the interface for account business logic
type Service interface {
	// Account management
	CreateAccount(ctx context.Context, req *CreateAccountRequest) (*Account, error)
	GetAccount(ctx context.Context, id string) (*Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)
	UpdateAccount(ctx context.Context, id string, req *UpdateAccountRequest) (*Account, error)
	DeleteAccount(ctx context.Context, id string) error
	ListAccounts(ctx context.Context, req *AccountListRequest) (*AccountListResponse, error)

	// Authentication
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	Logout(ctx context.Context, sessionToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
	ChangePassword(ctx context.Context, accountID string, req *ChangePasswordRequest) error
	ResetPassword(ctx context.Context, req *ResetPasswordRequest) error
	ConfirmPasswordReset(ctx context.Context, req *ConfirmPasswordResetRequest) error

	// Two-factor authentication
	EnableTwoFactor(ctx context.Context, accountID string, req *EnableTwoFactorRequest) error
	DisableTwoFactor(ctx context.Context, accountID string) error
	VerifyTwoFactor(ctx context.Context, accountID string, req *VerifyTwoFactorRequest) error

	// KYC management
	SubmitKYCDocument(ctx context.Context, accountID string, req *KYCSubmissionRequest) (*KYCDocument, error)
	GetKYCDocuments(ctx context.Context, accountID string) ([]KYCDocument, error)
	UpdateKYCStatus(ctx context.Context, accountID string, status KYCStatus, level KYCLevel) error

	// Security
	GetSecurityEvents(ctx context.Context, accountID string) ([]SecurityEvent, error)
	LogSecurityEvent(ctx context.Context, accountID string, eventType SecurityEventType, severity SecuritySeverity, description string, metadata map[string]interface{}) error

	// Session management
	ValidateSession(ctx context.Context, sessionToken string) (*Session, error)
	InvalidateSession(ctx context.Context, sessionToken string) error
	CleanupExpiredSessions(ctx context.Context) error
}

// AccountService implements the Service interface
type AccountService struct {
	repo   Repository
	config config.AccountsConfig
	logger *logger.Logger
	cache  redis.Client
}

// NewService creates a new account service
func NewService(repo Repository, cfg config.AccountsConfig, logger *logger.Logger, cache redis.Client) Service {
	return &AccountService{
		repo:   repo,
		config: cfg,
		logger: logger,
		cache:  cache,
	}
}

// CreateAccount creates a new account
func (s *AccountService) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*Account, error) {
	s.logger.Info("Creating new account", zap.String("email", req.Email))

	// Check if account already exists
	existingAccount, err := s.repo.GetAccountByEmail(ctx, req.Email)
	if err == nil && existingAccount != nil {
		return nil, fmt.Errorf("account with email %s already exists", req.Email)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create account
	account := &Account{
		ID:               uuid.New().String(),
		UserID:           uuid.New().String(),
		Email:            req.Email,
		Phone:            req.Phone,
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Country:          req.Country,
		AccountType:      req.AccountType,
		AccountStatus:    AccountStatusPending,
		KYCStatus:        KYCStatusNotStarted,
		KYCLevel:         KYCLevelNone,
		RiskScore:        0.0,
		ComplianceFlags:  []string{},
		TwoFactorEnabled: false,
		FailedLoginCount: 0,
		AccountLimits: AccountLimits{
			DailyTransactionLimit:   s.config.AccountLimits.DailyTransactionLimit,
			MonthlyTransactionLimit: s.config.AccountLimits.MonthlyTransactionLimit,
			MaxWallets:              s.config.AccountLimits.MaxWalletsPerUser,
			MaxCards:                s.config.AccountLimits.MaxCardsPerUser,
		},
		NotificationPrefs: NotificationPreferences{
			EmailEnabled:      s.config.NotificationSettings.EmailEnabled,
			SMSEnabled:        s.config.NotificationSettings.SMSEnabled,
			PushEnabled:       s.config.NotificationSettings.PushEnabled,
			SecurityAlerts:    s.config.NotificationSettings.SecurityAlerts,
			TransactionAlerts: s.config.NotificationSettings.TransactionAlerts,
		},
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store hashed password in metadata (in production, use separate table)
	account.Metadata["password_hash"] = string(hashedPassword)

	err = s.repo.CreateAccount(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	// Log security event
	s.LogSecurityEvent(ctx, account.ID, SecurityEventTypeLogin, SecuritySeverityLow, "Account created", nil)

	s.logger.Info("Account created successfully", zap.String("account_id", account.ID), zap.String("email", account.Email))
	return account, nil
}

// GetAccount retrieves an account by ID
func (s *AccountService) GetAccount(ctx context.Context, id string) (*Account, error) {
	account, err := s.repo.GetAccountByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Remove sensitive data
	account.Metadata = s.sanitizeMetadata(account.Metadata)

	return account, nil
}

// GetAccountByEmail retrieves an account by email
func (s *AccountService) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	account, err := s.repo.GetAccountByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Remove sensitive data
	account.Metadata = s.sanitizeMetadata(account.Metadata)

	return account, nil
}

// UpdateAccount updates an existing account
func (s *AccountService) UpdateAccount(ctx context.Context, id string, req *UpdateAccountRequest) (*Account, error) {
	s.logger.Info("Updating account", zap.String("account_id", id))

	account, err := s.repo.GetAccountByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Update fields if provided
	if req.FirstName != nil {
		account.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		account.LastName = *req.LastName
	}
	if req.Phone != nil {
		account.Phone = *req.Phone
	}
	if req.Country != nil {
		account.Country = *req.Country
	}
	if req.State != nil {
		account.State = *req.State
	}
	if req.City != nil {
		account.City = *req.City
	}
	if req.Address != nil {
		account.Address = *req.Address
	}
	if req.PostalCode != nil {
		account.PostalCode = *req.PostalCode
	}
	if req.NotificationPrefs != nil {
		account.NotificationPrefs = *req.NotificationPrefs
	}

	err = s.repo.UpdateAccount(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	// Remove sensitive data
	account.Metadata = s.sanitizeMetadata(account.Metadata)

	s.logger.Info("Account updated successfully", zap.String("account_id", id))
	return account, nil
}

// DeleteAccount deletes an account
func (s *AccountService) DeleteAccount(ctx context.Context, id string) error {
	s.logger.Info("Deleting account", zap.String("account_id", id))

	err := s.repo.DeleteAccount(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	s.logger.Info("Account deleted successfully", zap.String("account_id", id))
	return nil
}

// ListAccounts retrieves a list of accounts
func (s *AccountService) ListAccounts(ctx context.Context, req *AccountListRequest) (*AccountListResponse, error) {
	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	accounts, total, err := s.repo.ListAccounts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	// Remove sensitive data from all accounts
	for i := range accounts {
		accounts[i].Metadata = s.sanitizeMetadata(accounts[i].Metadata)
	}

	totalPages := (total + req.Limit - 1) / req.Limit

	return &AccountListResponse{
		Accounts:   accounts,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
	}, nil
}

// Login authenticates a user and creates a session
func (s *AccountService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	s.logger.Info("User login attempt", zap.String("email", req.Email))

	// Get account
	account, err := s.repo.GetAccountByEmail(ctx, req.Email)
	if err != nil {
		s.LogSecurityEvent(ctx, "", SecurityEventTypeFailedLogin, SecuritySeverityMedium, "Login attempt with invalid email", map[string]interface{}{"email": req.Email})
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check account status
	if account.AccountStatus != AccountStatusActive {
		s.LogSecurityEvent(ctx, account.ID, SecurityEventTypeFailedLogin, SecuritySeverityMedium, "Login attempt on inactive account", nil)
		return nil, fmt.Errorf("account is not active")
	}

	// Check failed login attempts
	if account.FailedLoginCount >= s.config.MaxLoginAttempts {
		s.LogSecurityEvent(ctx, account.ID, SecurityEventTypeAccountLocked, SecuritySeverityHigh, "Account locked due to too many failed login attempts", nil)
		return nil, fmt.Errorf("account is locked due to too many failed login attempts")
	}

	// Verify password
	passwordHash, ok := account.Metadata["password_hash"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid account data")
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		// Increment failed login count
		account.FailedLoginCount++
		s.repo.UpdateAccount(ctx, account)

		s.LogSecurityEvent(ctx, account.ID, SecurityEventTypeFailedLogin, SecuritySeverityMedium, "Invalid password", nil)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Reset failed login count on successful login
	account.FailedLoginCount = 0
	account.LastLoginAt = &[]time.Time{time.Now()}[0]
	s.repo.UpdateAccount(ctx, account)

	// Create session
	session := &Session{
		ID:           uuid.New().String(),
		AccountID:    account.ID,
		DeviceID:     req.DeviceID,
		SessionToken: s.generateToken(),
		RefreshToken: s.generateToken(),
		ExpiresAt:    time.Now().Add(24 * time.Hour), // 24 hours
		IsActive:     true,
		Metadata:     make(map[string]interface{}),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if req.RememberMe {
		session.ExpiresAt = time.Now().Add(30 * 24 * time.Hour) // 30 days
	}

	err = s.repo.CreateSession(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Log successful login
	s.LogSecurityEvent(ctx, account.ID, SecurityEventTypeLogin, SecuritySeverityLow, "Successful login", nil)

	// Remove sensitive data
	account.Metadata = s.sanitizeMetadata(account.Metadata)

	s.logger.Info("User logged in successfully", zap.String("account_id", account.ID), zap.String("email", account.Email))

	return &LoginResponse{
		Account:      account,
		AccessToken:  session.SessionToken,
		RefreshToken: session.RefreshToken,
		ExpiresIn:    int64(session.ExpiresAt.Sub(time.Now()).Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// Logout invalidates a session
func (s *AccountService) Logout(ctx context.Context, sessionToken string) error {
	session, err := s.repo.GetSessionByToken(ctx, sessionToken)
	if err != nil {
		return fmt.Errorf("invalid session")
	}

	err = s.repo.DeleteSession(ctx, session.ID)
	if err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	s.logger.Info("User logged out", zap.String("account_id", session.AccountID))
	return nil
}

// ValidateSession validates a session token
func (s *AccountService) ValidateSession(ctx context.Context, sessionToken string) (*Session, error) {
	session, err := s.repo.GetSessionByToken(ctx, sessionToken)
	if err != nil {
		return nil, fmt.Errorf("invalid session")
	}

	return session, nil
}

// LogSecurityEvent logs a security event
func (s *AccountService) LogSecurityEvent(ctx context.Context, accountID string, eventType SecurityEventType, severity SecuritySeverity, description string, metadata map[string]interface{}) error {
	event := &SecurityEvent{
		ID:          uuid.New().String(),
		AccountID:   accountID,
		EventType:   eventType,
		Severity:    severity,
		Description: description,
		Resolved:    false,
		Metadata:    metadata,
		CreatedAt:   time.Now(),
	}

	return s.repo.CreateSecurityEvent(ctx, event)
}

// GetSecurityEvents retrieves security events for an account
func (s *AccountService) GetSecurityEvents(ctx context.Context, accountID string) ([]SecurityEvent, error) {
	return s.repo.GetSecurityEvents(ctx, accountID, 50)
}

// Helper methods

// generateToken generates a random token
func (s *AccountService) generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// sanitizeMetadata removes sensitive data from metadata
func (s *AccountService) sanitizeMetadata(metadata map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})
	for k, v := range metadata {
		if k != "password_hash" {
			sanitized[k] = v
		}
	}
	return sanitized
}

// RefreshToken refreshes an access token using a refresh token
func (s *AccountService) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	s.logger.Info("Refreshing token", zap.String("refresh_token", refreshToken[:10]+"..."))

	// Find session by refresh token
	session, err := s.repo.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		s.LogSecurityEvent(ctx, "", SecurityEventTypeFailedLogin, SecuritySeverityMedium, "Invalid refresh token", nil)
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Check if session is still valid
	if !session.IsActive || session.ExpiresAt.Before(time.Now()) {
		s.LogSecurityEvent(ctx, session.AccountID, SecurityEventTypeFailedLogin, SecuritySeverityMedium, "Expired refresh token", nil)
		return nil, fmt.Errorf("refresh token expired")
	}

	// Get account
	account, err := s.repo.GetAccountByID(ctx, session.AccountID)
	if err != nil {
		return nil, fmt.Errorf("account not found")
	}

	// Generate new tokens
	newSessionToken := s.generateToken()
	newRefreshToken := s.generateToken()

	// Update session
	session.SessionToken = newSessionToken
	session.RefreshToken = newRefreshToken
	session.ExpiresAt = time.Now().Add(24 * time.Hour)
	session.UpdatedAt = time.Now()

	err = s.repo.UpdateSession(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	// Remove sensitive data
	account.Metadata = s.sanitizeMetadata(account.Metadata)

	s.logger.Info("Token refreshed successfully", zap.String("account_id", account.ID))

	return &LoginResponse{
		Account:      account,
		AccessToken:  newSessionToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(session.ExpiresAt.Sub(time.Now()).Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// ChangePassword changes user password
func (s *AccountService) ChangePassword(ctx context.Context, accountID string, req *ChangePasswordRequest) error {
	s.logger.Info("Changing password", zap.String("account_id", accountID))

	// Get account
	account, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("account not found")
	}

	// Verify current password
	currentPasswordHash, ok := account.Metadata["password_hash"].(string)
	if !ok {
		return fmt.Errorf("invalid account data")
	}

	err = bcrypt.CompareHashAndPassword([]byte(currentPasswordHash), []byte(req.CurrentPassword))
	if err != nil {
		s.LogSecurityEvent(ctx, accountID, SecurityEventTypeFailedLogin, SecuritySeverityMedium, "Invalid current password during password change", nil)
		return fmt.Errorf("invalid current password")
	}

	// Validate new password
	if err := s.validatePassword(req.NewPassword); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	account.Metadata["password_hash"] = string(hashedPassword)
	account.Metadata["password_changed_at"] = time.Now()
	account.UpdatedAt = time.Now()

	err = s.repo.UpdateAccount(ctx, account)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate all sessions except current one if provided
	if req.InvalidateOtherSessions {
		// TODO: Implement session invalidation
	}

	// Log security event
	s.LogSecurityEvent(ctx, accountID, SecurityEventTypePasswordChange, SecuritySeverityLow, "Password changed successfully", nil)

	s.logger.Info("Password changed successfully", zap.String("account_id", accountID))
	return nil
}

// ResetPassword initiates password reset process
func (s *AccountService) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	s.logger.Info("Password reset requested", zap.String("email", req.Email))

	// Get account by email
	account, err := s.repo.GetAccountByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal if email exists or not for security
		s.logger.Info("Password reset requested for non-existent email", zap.String("email", req.Email))
		return nil
	}

	// Generate reset token
	resetToken := s.generateToken()
	resetExpiry := time.Now().Add(1 * time.Hour) // 1 hour expiry

	// Store reset token in metadata
	account.Metadata["password_reset_token"] = resetToken
	account.Metadata["password_reset_expires"] = resetExpiry
	account.UpdatedAt = time.Now()

	err = s.repo.UpdateAccount(ctx, account)
	if err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	// Send reset email (placeholder - implement with actual email service)
	err = s.sendPasswordResetEmail(account.Email, resetToken)
	if err != nil {
		s.logger.Error("Failed to send password reset email", zap.Error(err), zap.String("email", account.Email))
		return fmt.Errorf("failed to send reset email")
	}

	// Log security event
	s.LogSecurityEvent(ctx, account.ID, SecurityEventTypePasswordReset, SecuritySeverityLow, "Password reset requested", nil)

	s.logger.Info("Password reset email sent", zap.String("account_id", account.ID))
	return nil
}

// ConfirmPasswordReset confirms password reset with token
func (s *AccountService) ConfirmPasswordReset(ctx context.Context, req *ConfirmPasswordResetRequest) error {
	s.logger.Info("Password reset confirmation", zap.String("token", req.Token[:10]+"..."))

	// Find account by reset token
	account, err := s.repo.GetAccountByResetToken(ctx, req.Token)
	if err != nil {
		s.LogSecurityEvent(ctx, "", SecurityEventTypeFailedPasswordReset, SecuritySeverityMedium, "Invalid password reset token", nil)
		return fmt.Errorf("invalid reset token")
	}

	// Check token expiry
	resetExpiry, ok := account.Metadata["password_reset_expires"].(time.Time)
	if !ok || resetExpiry.Before(time.Now()) {
		s.LogSecurityEvent(ctx, account.ID, SecurityEventTypeFailedPasswordReset, SecuritySeverityMedium, "Expired password reset token", nil)
		return fmt.Errorf("reset token expired")
	}

	// Validate new password
	if err := s.validatePassword(req.NewPassword); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password and clear reset token
	account.Metadata["password_hash"] = string(hashedPassword)
	account.Metadata["password_changed_at"] = time.Now()
	delete(account.Metadata, "password_reset_token")
	delete(account.Metadata, "password_reset_expires")
	account.UpdatedAt = time.Now()

	err = s.repo.UpdateAccount(ctx, account)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate all sessions
	err = s.repo.InvalidateAllSessions(ctx, account.ID)
	if err != nil {
		s.logger.Error("Failed to invalidate sessions after password reset", zap.Error(err))
	}

	// Log security event
	s.LogSecurityEvent(ctx, account.ID, SecurityEventTypePasswordReset, SecuritySeverityLow, "Password reset completed", nil)

	s.logger.Info("Password reset completed", zap.String("account_id", account.ID))
	return nil
}

func (s *AccountService) EnableTwoFactor(ctx context.Context, accountID string, req *EnableTwoFactorRequest) error {
	s.logger.Info("Enabling 2FA", zap.String("account_id", accountID), zap.String("method", string(req.Method)))

	// Get account
	account, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("account not found")
	}

	// Check if 2FA is already enabled
	if account.TwoFactorEnabled {
		return fmt.Errorf("two-factor authentication is already enabled")
	}

	// Generate secret for TOTP if method is totp
	var secret string
	if req.Method == TwoFactorMethodTOTP {
		secret = s.generateTOTPSecret()
		account.Metadata["totp_secret"] = secret
	}

	// Update account
	account.TwoFactorEnabled = true
	account.TwoFactorMethod = &req.Method
	account.UpdatedAt = time.Now()

	err = s.repo.UpdateAccount(ctx, account)
	if err != nil {
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}

	// Log security event
	s.LogSecurityEvent(ctx, accountID, SecurityEventTypeTwoFactorEnabled, SecuritySeverityLow, "Two-factor authentication enabled", map[string]interface{}{
		"method": req.Method,
	})

	s.logger.Info("2FA enabled successfully", zap.String("account_id", accountID))
	return nil
}

func (s *AccountService) DisableTwoFactor(ctx context.Context, accountID string) error {
	s.logger.Info("Disabling 2FA", zap.String("account_id", accountID))

	// Get account
	account, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("account not found")
	}

	// Check if 2FA is enabled
	if !account.TwoFactorEnabled {
		return fmt.Errorf("two-factor authentication is not enabled")
	}

	// Update account
	account.TwoFactorEnabled = false
	account.TwoFactorMethod = nil
	account.UpdatedAt = time.Now()

	// Remove 2FA secrets from metadata
	delete(account.Metadata, "totp_secret")

	err = s.repo.UpdateAccount(ctx, account)
	if err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	// Log security event
	s.LogSecurityEvent(ctx, accountID, SecurityEventTypeTwoFactorEnabled, SecuritySeverityMedium, "Two-factor authentication disabled", nil)

	s.logger.Info("2FA disabled successfully", zap.String("account_id", accountID))
	return nil
}

func (s *AccountService) VerifyTwoFactor(ctx context.Context, accountID string, req *VerifyTwoFactorRequest) error {
	s.logger.Info("Verifying 2FA", zap.String("account_id", accountID))

	// Get account
	account, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("account not found")
	}

	// Check if 2FA is enabled
	if !account.TwoFactorEnabled || account.TwoFactorMethod == nil {
		return fmt.Errorf("two-factor authentication is not enabled")
	}

	// Verify based on method
	switch *account.TwoFactorMethod {
	case TwoFactorMethodTOTP:
		return s.verifyTOTP(account, req.Code)
	case TwoFactorMethodSMS:
		return s.verifySMS(account, req.Code)
	case TwoFactorMethodEmail:
		return s.verifyEmail(account, req.Code)
	default:
		return fmt.Errorf("unsupported 2FA method")
	}
}

func (s *AccountService) SubmitKYCDocument(ctx context.Context, accountID string, req *KYCSubmissionRequest) (*KYCDocument, error) {
	s.logger.Info("Submitting KYC document", zap.String("account_id", accountID), zap.String("document_type", string(req.DocumentType)))

	// Create KYC document
	doc := &KYCDocument{
		ID:           uuid.New().String(),
		AccountID:    accountID,
		DocumentType: req.DocumentType,
		DocumentURL:  req.DocumentURL,
		Status:       DocumentStatusPending,
		UploadedAt:   time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	err := s.repo.CreateKYCDocument(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to create KYC document: %w", err)
	}

	// Log security event
	s.LogSecurityEvent(ctx, accountID, SecurityEventTypeDataAccess, SecuritySeverityLow, "KYC document submitted", map[string]interface{}{
		"document_type": req.DocumentType,
		"document_id":   doc.ID,
	})

	s.logger.Info("KYC document submitted successfully", zap.String("account_id", accountID), zap.String("document_id", doc.ID))
	return doc, nil
}

func (s *AccountService) GetKYCDocuments(ctx context.Context, accountID string) ([]KYCDocument, error) {
	return s.repo.GetKYCDocuments(ctx, accountID)
}

func (s *AccountService) UpdateKYCStatus(ctx context.Context, accountID string, status KYCStatus, level KYCLevel) error {
	s.logger.Info("Updating KYC status", zap.String("account_id", accountID), zap.String("status", string(status)), zap.String("level", string(level)))

	// Get account
	account, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("account not found")
	}

	// Update KYC status and level
	account.KYCStatus = status
	account.KYCLevel = level
	account.UpdatedAt = time.Now()

	err = s.repo.UpdateAccount(ctx, account)
	if err != nil {
		return fmt.Errorf("failed to update KYC status: %w", err)
	}

	// Log security event
	s.LogSecurityEvent(ctx, accountID, SecurityEventTypeDataAccess, SecuritySeverityLow, "KYC status updated", map[string]interface{}{
		"old_status": account.KYCStatus,
		"new_status": status,
		"old_level":  account.KYCLevel,
		"new_level":  level,
	})

	s.logger.Info("KYC status updated successfully", zap.String("account_id", accountID))
	return nil
}

func (s *AccountService) InvalidateSession(ctx context.Context, sessionToken string) error {
	return s.Logout(ctx, sessionToken)
}

func (s *AccountService) CleanupExpiredSessions(ctx context.Context) error {
	return s.repo.DeleteExpiredSessions(ctx)
}

// validatePassword validates password strength
func (s *AccountService) validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Add more password validation rules as needed
	// - uppercase letter
	// - lowercase letter
	// - number
	// - special character

	return nil
}

// sendPasswordResetEmail sends password reset email (placeholder)
func (s *AccountService) sendPasswordResetEmail(email, token string) error {
	// TODO: Implement actual email sending
	s.logger.Info("Password reset email would be sent", zap.String("email", email), zap.String("token", token[:10]+"..."))
	return nil
}

// generateTOTPSecret generates a TOTP secret
func (s *AccountService) generateTOTPSecret() string {
	// TODO: Implement actual TOTP secret generation
	return s.generateToken()
}

// verifyTOTP verifies a TOTP code
func (s *AccountService) verifyTOTP(account *Account, code string) error {
	// TODO: Implement actual TOTP verification
	s.logger.Info("TOTP verification would be performed", zap.String("account_id", account.ID))
	return nil
}

// verifySMS verifies an SMS code
func (s *AccountService) verifySMS(account *Account, code string) error {
	// TODO: Implement actual SMS verification
	s.logger.Info("SMS verification would be performed", zap.String("account_id", account.ID))
	return nil
}

// verifyEmail verifies an email code
func (s *AccountService) verifyEmail(account *Account, code string) error {
	// TODO: Implement actual email verification
	s.logger.Info("Email verification would be performed", zap.String("account_id", account.ID))
	return nil
}
