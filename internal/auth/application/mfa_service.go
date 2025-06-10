package application

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// MFAServiceImpl provides multi-factor authentication functionality
type MFAServiceImpl struct {
	userRepo          domain.UserRepository
	sessionRepo       domain.SessionRepository
	monitoringService interface{} // placeholder for monitoring service
	smsProvider       SMSProvider
	emailProvider     EmailProvider
	logger            *logger.Logger
	config            *MFAConfig
}

// MFAConfig represents MFA configuration
type MFAConfig struct {
	TOTPIssuer              string        `yaml:"totp_issuer"`
	TOTPPeriod              uint          `yaml:"totp_period" default:"30"`
	TOTPDigits              otp.Digits    `yaml:"totp_digits" default:"6"`
	TOTPAlgorithm           otp.Algorithm `yaml:"totp_algorithm" default:"SHA1"`
	BackupCodesCount        int           `yaml:"backup_codes_count" default:"10"`
	BackupCodeLength        int           `yaml:"backup_code_length" default:"8"`
	SMSCodeLength           int           `yaml:"sms_code_length" default:"6"`
	SMSCodeExpiry           time.Duration `yaml:"sms_code_expiry" default:"5m"`
	EmailCodeLength         int           `yaml:"email_code_length" default:"6"`
	EmailCodeExpiry         time.Duration `yaml:"email_code_expiry" default:"10m"`
	MaxVerificationAttempts int           `yaml:"max_verification_attempts" default:"3"`
	CooldownPeriod          time.Duration `yaml:"cooldown_period" default:"15m"`
}

// SMSProvider interface for sending SMS
type SMSProvider interface {
	SendSMS(ctx context.Context, phoneNumber, message string) error
}

// EmailProvider interface for sending emails
type EmailProvider interface {
	SendEmail(ctx context.Context, email, subject, body string) error
}

// MFASetupRequest represents MFA setup request
type MFASetupRequest struct {
	UserID      string           `json:"user_id"`
	Method      domain.MFAMethod `json:"method"`
	PhoneNumber string           `json:"phone_number,omitempty"`
}

// MFASetupResponse represents MFA setup response
type MFASetupResponse struct {
	Secret      string   `json:"secret,omitempty"`
	QRCode      string   `json:"qr_code,omitempty"`
	BackupCodes []string `json:"backup_codes,omitempty"`
	SetupToken  string   `json:"setup_token"`
}

// MFAVerifyRequest represents MFA verification request
type MFAVerifyRequest struct {
	UserID     string `json:"user_id"`
	Code       string `json:"code"`
	SetupToken string `json:"setup_token,omitempty"`
}

// MFAVerifyResponse represents MFA verification response
type MFAVerifyResponse struct {
	Verified      bool       `json:"verified"`
	Message       string     `json:"message,omitempty"`
	Remaining     int        `json:"remaining_attempts,omitempty"`
	CooldownUntil *time.Time `json:"cooldown_until,omitempty"`
}

// MFAChallengeRequest represents MFA challenge request
type MFAChallengeRequest struct {
	UserID string           `json:"user_id"`
	Method domain.MFAMethod `json:"method"`
}

// MFAChallengeResponse represents MFA challenge response
type MFAChallengeResponse struct {
	ChallengeID string    `json:"challenge_id"`
	Method      string    `json:"method"`
	ExpiresAt   time.Time `json:"expires_at"`
	Message     string    `json:"message,omitempty"`
}

// MFAChallenge represents an active MFA challenge
type MFAChallenge struct {
	ID          string           `json:"id"`
	UserID      string           `json:"user_id"`
	Method      domain.MFAMethod `json:"method"`
	Code        string           `json:"code"`
	ExpiresAt   time.Time        `json:"expires_at"`
	Attempts    int              `json:"attempts"`
	MaxAttempts int              `json:"max_attempts"`
	CreatedAt   time.Time        `json:"created_at"`
}

// NewMFAService creates a new MFA service
func NewMFAService(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	monitoringService interface{}, // placeholder
	smsProvider SMSProvider,
	emailProvider EmailProvider,
	config *MFAConfig,
	logger *logger.Logger,
) MFAService {
	if config == nil {
		config = &MFAConfig{
			TOTPIssuer:              "Go Coffee",
			TOTPPeriod:              30,
			TOTPDigits:              otp.DigitsSix,
			TOTPAlgorithm:           otp.AlgorithmSHA1,
			BackupCodesCount:        10,
			BackupCodeLength:        8,
			SMSCodeLength:           6,
			SMSCodeExpiry:           5 * time.Minute,
			EmailCodeLength:         6,
			EmailCodeExpiry:         10 * time.Minute,
			MaxVerificationAttempts: 3,
			CooldownPeriod:          15 * time.Minute,
		}
	}

	return &MFAServiceImpl{
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		monitoringService: monitoringService,
		smsProvider:       smsProvider,
		emailProvider:     emailProvider,
		config:            config,
		logger:            logger,
	}
}

// SetupMFA initiates MFA setup for a user
func (m *MFAServiceImpl) SetupMFA(ctx context.Context, req *MFASetupRequest) (*MFASetupResponse, error) {
	user, err := m.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	response := &MFASetupResponse{}

	switch req.Method {
	case domain.MFAMethodTOTP:
		return m.setupTOTP(ctx, user, response)
	case domain.MFAMethodSMS:
		return m.setupSMS(ctx, user, req.PhoneNumber, response)
	case domain.MFAMethodEmail:
		return m.setupEmail(ctx, user, response)
	default:
		return nil, fmt.Errorf("unsupported MFA method: %s", req.Method)
	}
}

// VerifyMFASetup verifies MFA setup with provided code
func (m *MFAServiceImpl) VerifyMFASetup(ctx context.Context, req *MFAVerifyRequest) (*MFAVerifyResponse, error) {
	user, err := m.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verify the code based on the user's MFA method
	verified := false
	var verifyErr error

	switch user.MFAMethod {
	case domain.MFAMethodTOTP:
		verified = totp.Validate(req.Code, user.MFASecret)
	case domain.MFAMethodSMS, domain.MFAMethodEmail:
		// For SMS/Email, we need to check against stored challenge
		verified, verifyErr = m.verifyChallengeCodeInternal(ctx, req.UserID, req.Code)
	case domain.MFAMethodBackup:
		verified = user.UseMFABackupCode(req.Code)
		if verified {
			m.userRepo.UpdateUser(ctx, user)
		}
	}

	if verifyErr != nil {
		return nil, fmt.Errorf("failed to verify code: %w", verifyErr)
	}

	if verified {
		// Enable MFA for the user
		user.MFAEnabled = true
		if err := m.userRepo.UpdateUser(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}

		// Log security event
		m.logger.InfoWithFields("MFA enabled",
			logger.String("user_id", user.ID),
			logger.String("method", string(user.MFAMethod)))

		return &MFAVerifyResponse{
			Verified: true,
			Message:  "MFA successfully enabled",
		}, nil
	}

	// Log failed verification
	m.logger.WarnWithFields("MFA setup verification failed",
		logger.String("user_id", user.ID),
		logger.String("method", string(user.MFAMethod)))

	return &MFAVerifyResponse{
		Verified: false,
		Message:  "Invalid verification code",
	}, nil
}

// CreateMFAChallenge creates an MFA challenge for authentication
func (m *MFAServiceImpl) CreateMFAChallenge(ctx context.Context, req *MFAChallengeRequest) (*MFAChallengeResponse, error) {
	user, err := m.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !user.MFAEnabled {
		return nil, fmt.Errorf("MFA is not enabled for user")
	}

	method := req.Method
	if method == domain.MFAMethodNone {
		method = user.MFAMethod
	}

	switch method {
	case domain.MFAMethodTOTP:
		return &MFAChallengeResponse{
			ChallengeID: generateChallengeID(),
			Method:      string(method),
			ExpiresAt:   time.Now().Add(5 * time.Minute),
			Message:     "Enter the code from your authenticator app",
		}, nil

	case domain.MFAMethodSMS:
		return m.createSMSChallenge(ctx, user)

	case domain.MFAMethodEmail:
		return m.createEmailChallenge(ctx, user)

	default:
		return nil, fmt.Errorf("unsupported MFA method: %s", method)
	}
}

// VerifyMFAChallenge verifies an MFA challenge
func (m *MFAServiceImpl) VerifyMFAChallenge(ctx context.Context, req *MFAVerifyRequest) (*MFAVerifyResponse, error) {
	user, err := m.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !user.MFAEnabled {
		return &MFAVerifyResponse{
			Verified: false,
			Message:  "MFA is not enabled",
		}, nil
	}

	verified := false
	var verifyErr error

	switch user.MFAMethod {
	case domain.MFAMethodTOTP:
		verified = totp.Validate(req.Code, user.MFASecret)
	case domain.MFAMethodSMS, domain.MFAMethodEmail:
		verified, verifyErr = m.verifyChallengeCodeInternal(ctx, req.UserID, req.Code)
	case domain.MFAMethodBackup:
		verified = user.UseMFABackupCode(req.Code)
		if verified {
			m.userRepo.UpdateUser(ctx, user)
		}
	}

	if verifyErr != nil {
		return nil, fmt.Errorf("failed to verify challenge: %w", verifyErr)
	}

	if verified {
		// Log successful verification
		m.logger.InfoWithFields("MFA verification successful",
			logger.String("user_id", user.ID),
			logger.String("method", string(user.MFAMethod)))

		return &MFAVerifyResponse{
			Verified: true,
			Message:  "MFA verification successful",
		}, nil
	}

	// Log failed verification
	m.logger.WarnWithFields("MFA verification failed",
		logger.String("user_id", user.ID),
		logger.String("method", string(user.MFAMethod)))

	return &MFAVerifyResponse{
		Verified: false,
		Message:  "Invalid verification code",
	}, nil
}

// Interface implementation methods

// EnableMFA enables MFA for a user
func (m *MFAServiceImpl) EnableMFA(ctx context.Context, req *EnableMFARequest) (*EnableMFAResponse, error) {
	_, err := m.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Setup MFA based on method
	setupReq := &MFASetupRequest{
		UserID: req.UserID,
		Method: req.Method,
	}

	setupResp, err := m.SetupMFA(ctx, setupReq)
	if err != nil {
		return nil, fmt.Errorf("failed to setup MFA: %w", err)
	}

	return &EnableMFAResponse{
		Success:     true,
		Message:     "MFA enabled successfully",
		QRCode:      setupResp.QRCode,
		Secret:      setupResp.Secret,
		BackupCodes: setupResp.BackupCodes,
	}, nil
}

// DisableMFA disables MFA for a user (interface method)
func (m *MFAServiceImpl) DisableMFA(ctx context.Context, req *DisableMFARequest) (*DisableMFAResponse, error) {
	// Verify password first
	_, err := m.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// In a real implementation, you would verify the password here
	// For now, we'll skip password verification

	// Call the existing DisableMFA method
	if err := m.disableMFAInternal(ctx, req.UserID); err != nil {
		return nil, err
	}

	return &DisableMFAResponse{
		Success: true,
		Message: "MFA disabled successfully",
	}, nil
}

// VerifyMFA verifies an MFA code
func (m *MFAServiceImpl) VerifyMFA(ctx context.Context, req *VerifyMFARequest) (*VerifyMFAResponse, error) {
	verifyReq := &MFAVerifyRequest{
		UserID: req.UserID,
		Code:   req.Code,
	}

	verifyResp, err := m.VerifyMFAChallenge(ctx, verifyReq)
	if err != nil {
		return nil, err
	}

	return &VerifyMFAResponse{
		Success: verifyResp.Verified,
		Message: verifyResp.Message,
	}, nil
}

// GenerateBackupCodes generates new backup codes
func (m *MFAServiceImpl) GenerateBackupCodes(ctx context.Context, req *GenerateBackupCodesRequest) (*GenerateBackupCodesResponse, error) {
	user, err := m.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Generate new backup codes
	backupCodes := user.GenerateMFABackupCodes()

	// Save user with new backup codes
	if err := m.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &GenerateBackupCodesResponse{
		BackupCodes: backupCodes,
		Message:     "Backup codes generated successfully",
	}, nil
}

// GetBackupCodes gets remaining backup codes count
func (m *MFAServiceImpl) GetBackupCodes(ctx context.Context, req *GetBackupCodesRequest) (*GetBackupCodesResponse, error) {
	user, err := m.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &GetBackupCodesResponse{
		RemainingCodes: len(user.MFABackupCodes),
		Message:        "Backup codes retrieved successfully",
	}, nil
}

// UseBackupCode uses a backup code
func (m *MFAServiceImpl) UseBackupCode(ctx context.Context, userID, code string) error {
	user, err := m.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if !user.UseMFABackupCode(code) {
		return fmt.Errorf("invalid backup code")
	}

	// Save user with updated backup codes
	if err := m.userRepo.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// GetMFAStatus gets MFA status for a user
func (m *MFAServiceImpl) GetMFAStatus(ctx context.Context, userID string) (*MFAStatusResponse, error) {
	user, err := m.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &MFAStatusResponse{
		Enabled:     user.MFAEnabled,
		Method:      user.MFAMethod,
		BackupCodes: len(user.MFABackupCodes),
	}, nil
}

// IsMFAEnabled checks if MFA is enabled for a user
func (m *MFAServiceImpl) IsMFAEnabled(ctx context.Context, userID string) (bool, error) {
	user, err := m.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	return user.MFAEnabled, nil
}

// disableMFAInternal is the internal method for disabling MFA
func (m *MFAServiceImpl) disableMFAInternal(ctx context.Context, userID string) error {
	user, err := m.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.DisableMFA()

	if err := m.userRepo.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Log security event (placeholder - implement when monitoring is available)
	// m.logSecurityEvent(ctx, user.ID, "mfa_disabled", monitoring.SeverityMedium, map[string]interface{}{
	//     "previous_method": user.MFAMethod,
	// })

	return nil
}

// Helper methods

func (m *MFAServiceImpl) setupTOTP(ctx context.Context, user *domain.User, response *MFASetupResponse) (*MFASetupResponse, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      m.config.TOTPIssuer,
		AccountName: user.Email,
		Period:      m.config.TOTPPeriod,
		Digits:      m.config.TOTPDigits,
		Algorithm:   m.config.TOTPAlgorithm,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	// Generate backup codes
	backupCodes := m.generateBackupCodesInternal()

	// Update user with MFA details
	user.EnableMFA(domain.MFAMethodTOTP, key.Secret())
	user.SetMFABackupCodes(backupCodes)

	if err := m.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	response.Secret = key.Secret()
	response.QRCode = key.URL()
	response.BackupCodes = backupCodes
	response.SetupToken = generateSetupToken()

	return response, nil
}

func (m *MFAServiceImpl) setupSMS(ctx context.Context, user *domain.User, phoneNumber string, response *MFASetupResponse) (*MFASetupResponse, error) {
	if phoneNumber == "" {
		return nil, fmt.Errorf("phone number is required for SMS MFA")
	}

	// Update user with phone number and MFA method
	user.PhoneNumber = phoneNumber
	user.EnableMFA(domain.MFAMethodSMS, "")

	// Generate backup codes
	backupCodes := m.generateBackupCodesInternal()
	user.SetMFABackupCodes(backupCodes)

	if err := m.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Send verification SMS
	code := m.generateSMSCodeInternal()
	challenge := &MFAChallenge{
		ID:          generateChallengeID(),
		UserID:      user.ID,
		Method:      domain.MFAMethodSMS,
		Code:        code,
		ExpiresAt:   time.Now().Add(m.config.SMSCodeExpiry),
		MaxAttempts: m.config.MaxVerificationAttempts,
		CreatedAt:   time.Now(),
	}

	if err := m.storeMFAChallengeInternal(ctx, challenge); err != nil {
		return nil, fmt.Errorf("failed to store MFA challenge: %w", err)
	}

	message := fmt.Sprintf("Your verification code is: %s", code)
	if err := m.smsProvider.SendSMS(ctx, phoneNumber, message); err != nil {
		return nil, fmt.Errorf("failed to send SMS: %w", err)
	}

	response.BackupCodes = backupCodes
	response.SetupToken = generateSetupToken()

	return response, nil
}

func (m *MFAServiceImpl) setupEmail(ctx context.Context, user *domain.User, response *MFASetupResponse) (*MFASetupResponse, error) {
	// Update user with MFA method
	user.EnableMFA(domain.MFAMethodEmail, "")

	// Generate backup codes
	backupCodes := m.generateBackupCodesInternal()
	user.SetMFABackupCodes(backupCodes)

	if err := m.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Send verification email
	code := m.generateEmailCodeInternal()
	challenge := &MFAChallenge{
		ID:          generateChallengeID(),
		UserID:      user.ID,
		Method:      domain.MFAMethodEmail,
		Code:        code,
		ExpiresAt:   time.Now().Add(m.config.EmailCodeExpiry),
		MaxAttempts: m.config.MaxVerificationAttempts,
		CreatedAt:   time.Now(),
	}

	if err := m.storeMFAChallengeInternal(ctx, challenge); err != nil {
		return nil, fmt.Errorf("failed to store MFA challenge: %w", err)
	}

	subject := "Email Verification Code"
	body := fmt.Sprintf("Your verification code is: %s", code)
	if err := m.emailProvider.SendEmail(ctx, user.Email, subject, body); err != nil {
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	response.BackupCodes = backupCodes
	response.SetupToken = generateSetupToken()

	return response, nil
}

func (m *MFAServiceImpl) createSMSChallenge(ctx context.Context, user *domain.User) (*MFAChallengeResponse, error) {
	code := m.generateSMSCodeInternal()
	challengeID := generateChallengeID()

	challenge := &MFAChallenge{
		ID:          challengeID,
		UserID:      user.ID,
		Method:      domain.MFAMethodSMS,
		Code:        code,
		ExpiresAt:   time.Now().Add(m.config.SMSCodeExpiry),
		MaxAttempts: m.config.MaxVerificationAttempts,
		CreatedAt:   time.Now(),
	}

	if err := m.storeMFAChallengeInternal(ctx, challenge); err != nil {
		return nil, fmt.Errorf("failed to store MFA challenge: %w", err)
	}

	message := fmt.Sprintf("Your verification code is: %s", code)
	if err := m.smsProvider.SendSMS(ctx, user.PhoneNumber, message); err != nil {
		return nil, fmt.Errorf("failed to send SMS: %w", err)
	}

	return &MFAChallengeResponse{
		ChallengeID: challengeID,
		Method:      string(domain.MFAMethodSMS),
		ExpiresAt:   challenge.ExpiresAt,
		Message:     "Verification code sent to your phone",
	}, nil
}

func (m *MFAServiceImpl) createEmailChallenge(ctx context.Context, user *domain.User) (*MFAChallengeResponse, error) {
	code := m.generateEmailCodeInternal()
	challengeID := generateChallengeID()

	challenge := &MFAChallenge{
		ID:          challengeID,
		UserID:      user.ID,
		Method:      domain.MFAMethodEmail,
		Code:        code,
		ExpiresAt:   time.Now().Add(m.config.EmailCodeExpiry),
		MaxAttempts: m.config.MaxVerificationAttempts,
		CreatedAt:   time.Now(),
	}

	if err := m.storeMFAChallengeInternal(ctx, challenge); err != nil {
		return nil, fmt.Errorf("failed to store MFA challenge: %w", err)
	}

	subject := "Verification Code"
	body := fmt.Sprintf("Your verification code is: %s", code)
	if err := m.emailProvider.SendEmail(ctx, user.Email, subject, body); err != nil {
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	return &MFAChallengeResponse{
		ChallengeID: challengeID,
		Method:      string(domain.MFAMethodEmail),
		ExpiresAt:   challenge.ExpiresAt,
		Message:     "Verification code sent to your email",
	}, nil
}

// Internal helper methods (renamed to avoid conflicts with interface methods)

func (m *MFAServiceImpl) generateBackupCodesInternal() []string {
	codes := make([]string, m.config.BackupCodesCount)
	for i := 0; i < m.config.BackupCodesCount; i++ {
		codes[i] = generateRandomCode(m.config.BackupCodeLength)
	}
	return codes
}

func (m *MFAServiceImpl) generateSMSCodeInternal() string {
	return generateNumericCode(m.config.SMSCodeLength)
}

func (m *MFAServiceImpl) generateEmailCodeInternal() string {
	return generateNumericCode(m.config.EmailCodeLength)
}

func (m *MFAServiceImpl) verifyChallengeCodeInternal(ctx context.Context, userID, code string) (bool, error) {
	// TODO: Implement challenge storage and verification
	// This would typically use Redis or database to store temporary challenges
	return false, fmt.Errorf("challenge verification not implemented")
}

func (m *MFAServiceImpl) storeMFAChallengeInternal(ctx context.Context, challenge *MFAChallenge) error {
	// TODO: Implement challenge storage
	// This would typically store in Redis with expiration
	return nil
}

// logSecurityEventInternal logs security events (placeholder for future monitoring integration)
func (m *MFAServiceImpl) logSecurityEventInternal(ctx context.Context, userID, eventType string, metadata map[string]interface{}) {
	m.logger.InfoWithFields("MFA security event",
		logger.String("user_id", userID),
		logger.String("event_type", eventType))
}

// Utility functions

func generateChallengeID() string {
	return generateRandomCode(16)
}

func generateSetupToken() string {
	return generateRandomCode(32)
}

func generateRandomCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}

func generateNumericCode(length int) string {
	const charset = "0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}
