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
	"github.com/DimaJoyti/go-coffee/pkg/security/monitoring"
)

// UserRepository provides access to user storage
type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUser(ctx context.Context, userID string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	DeleteUser(ctx context.Context, userID string) error
}

// SessionRepository provides access to session storage
type SessionRepository interface {
	CreateSession(ctx context.Context, session *domain.Session) error
	GetSession(ctx context.Context, sessionID string) (*domain.Session, error)
	UpdateSession(ctx context.Context, session *domain.Session) error
	DeleteSession(ctx context.Context, sessionID string) error
}

// MFAService provides multi-factor authentication functionality
type MFAService struct {
	userRepo          UserRepository
	sessionRepo       SessionRepository
	monitoringService *monitoring.SecurityMonitoringService
	smsProvider       SMSProvider
	emailProvider     EmailProvider
	logger            *logger.Logger
	config            *MFAConfig
}

// MFAConfig represents MFA configuration
type MFAConfig struct {
	TOTPIssuer           string        `yaml:"totp_issuer"`
	TOTPPeriod           uint          `yaml:"totp_period" default:"30"`
	TOTPDigits           otp.Digits    `yaml:"totp_digits" default:"6"`
	TOTPAlgorithm        otp.Algorithm `yaml:"totp_algorithm" default:"SHA1"`
	BackupCodesCount     int           `yaml:"backup_codes_count" default:"10"`
	BackupCodeLength     int           `yaml:"backup_code_length" default:"8"`
	SMSCodeLength        int           `yaml:"sms_code_length" default:"6"`
	SMSCodeExpiry        time.Duration `yaml:"sms_code_expiry" default:"5m"`
	EmailCodeLength      int           `yaml:"email_code_length" default:"6"`
	EmailCodeExpiry      time.Duration `yaml:"email_code_expiry" default:"10m"`
	MaxVerificationAttempts int        `yaml:"max_verification_attempts" default:"3"`
	CooldownPeriod       time.Duration `yaml:"cooldown_period" default:"15m"`
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
	UserID      string             `json:"user_id"`
	Method      domain.MFAMethod   `json:"method"`
	PhoneNumber string             `json:"phone_number,omitempty"`
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
	Verified    bool   `json:"verified"`
	Message     string `json:"message,omitempty"`
	Remaining   int    `json:"remaining_attempts,omitempty"`
	CooldownUntil *time.Time `json:"cooldown_until,omitempty"`
}

// MFAChallengeRequest represents MFA challenge request
type MFAChallengeRequest struct {
	UserID string             `json:"user_id"`
	Method domain.MFAMethod   `json:"method"`
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
	ID          string             `json:"id"`
	UserID      string             `json:"user_id"`
	Method      domain.MFAMethod   `json:"method"`
	Code        string             `json:"code"`
	ExpiresAt   time.Time          `json:"expires_at"`
	Attempts    int                `json:"attempts"`
	MaxAttempts int                `json:"max_attempts"`
	CreatedAt   time.Time          `json:"created_at"`
}

// NewMFAService creates a new MFA service
func NewMFAService(
	userRepo UserRepository,
	sessionRepo SessionRepository,
	monitoringService *monitoring.SecurityMonitoringService,
	smsProvider SMSProvider,
	emailProvider EmailProvider,
	config *MFAConfig,
	logger *logger.Logger,
) *MFAService {
	return &MFAService{
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
func (m *MFAService) SetupMFA(ctx context.Context, req *MFASetupRequest) (*MFASetupResponse, error) {
	user, err := m.userRepo.GetUser(ctx, req.UserID)
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
func (m *MFAService) VerifyMFASetup(ctx context.Context, req *MFAVerifyRequest) (*MFAVerifyResponse, error) {
	user, err := m.userRepo.GetUser(ctx, req.UserID)
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
		verified, verifyErr = m.verifyChallengeCode(ctx, req.UserID, req.Code)
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
		m.logSecurityEvent(ctx, user.ID, "mfa_enabled", monitoring.SeverityInfo, map[string]interface{}{
			"method": user.MFAMethod,
		})

		return &MFAVerifyResponse{
			Verified: true,
			Message:  "MFA successfully enabled",
		}, nil
	}

	// Log failed verification
	m.logSecurityEvent(ctx, user.ID, "mfa_setup_verification_failed", monitoring.SeverityMedium, map[string]interface{}{
		"method": user.MFAMethod,
	})

	return &MFAVerifyResponse{
		Verified: false,
		Message:  "Invalid verification code",
	}, nil
}

// CreateMFAChallenge creates an MFA challenge for authentication
func (m *MFAService) CreateMFAChallenge(ctx context.Context, req *MFAChallengeRequest) (*MFAChallengeResponse, error) {
	user, err := m.userRepo.GetUser(ctx, req.UserID)
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
func (m *MFAService) VerifyMFAChallenge(ctx context.Context, req *MFAVerifyRequest) (*MFAVerifyResponse, error) {
	user, err := m.userRepo.GetUser(ctx, req.UserID)
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
		verified, verifyErr = m.verifyChallengeCode(ctx, req.UserID, req.Code)
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
		m.logSecurityEvent(ctx, user.ID, "mfa_verification_success", monitoring.SeverityInfo, map[string]interface{}{
			"method": user.MFAMethod,
		})

		return &MFAVerifyResponse{
			Verified: true,
			Message:  "MFA verification successful",
		}, nil
	}

	// Log failed verification
	m.logSecurityEvent(ctx, user.ID, "mfa_verification_failed", monitoring.SeverityMedium, map[string]interface{}{
		"method": user.MFAMethod,
	})

	return &MFAVerifyResponse{
		Verified: false,
		Message:  "Invalid verification code",
	}, nil
}

// DisableMFA disables MFA for a user
func (m *MFAService) DisableMFA(ctx context.Context, userID string) error {
	user, err := m.userRepo.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.DisableMFA()

	if err := m.userRepo.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Log security event
	m.logSecurityEvent(ctx, user.ID, "mfa_disabled", monitoring.SeverityMedium, map[string]interface{}{
		"previous_method": user.MFAMethod,
	})

	return nil
}

// Helper methods

func (m *MFAService) setupTOTP(ctx context.Context, user *domain.User, response *MFASetupResponse) (*MFASetupResponse, error) {
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
	backupCodes := m.generateBackupCodes()

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

func (m *MFAService) setupSMS(ctx context.Context, user *domain.User, phoneNumber string, response *MFASetupResponse) (*MFASetupResponse, error) {
	if phoneNumber == "" {
		return nil, fmt.Errorf("phone number is required for SMS MFA")
	}

	// Update user with phone number and MFA method
	user.PhoneNumber = phoneNumber
	user.EnableMFA(domain.MFAMethodSMS, "")

	// Generate backup codes
	backupCodes := m.generateBackupCodes()
	user.SetMFABackupCodes(backupCodes)

	if err := m.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Send verification SMS
	code := m.generateSMSCode()
	challenge := &MFAChallenge{
		ID:          generateChallengeID(),
		UserID:      user.ID,
		Method:      domain.MFAMethodSMS,
		Code:        code,
		ExpiresAt:   time.Now().Add(m.config.SMSCodeExpiry),
		MaxAttempts: m.config.MaxVerificationAttempts,
		CreatedAt:   time.Now(),
	}

	if err := m.storeMFAChallenge(ctx, challenge); err != nil {
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

func (m *MFAService) setupEmail(ctx context.Context, user *domain.User, response *MFASetupResponse) (*MFASetupResponse, error) {
	// Update user with MFA method
	user.EnableMFA(domain.MFAMethodEmail, "")

	// Generate backup codes
	backupCodes := m.generateBackupCodes()
	user.SetMFABackupCodes(backupCodes)

	if err := m.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Send verification email
	code := m.generateEmailCode()
	challenge := &MFAChallenge{
		ID:          generateChallengeID(),
		UserID:      user.ID,
		Method:      domain.MFAMethodEmail,
		Code:        code,
		ExpiresAt:   time.Now().Add(m.config.EmailCodeExpiry),
		MaxAttempts: m.config.MaxVerificationAttempts,
		CreatedAt:   time.Now(),
	}

	if err := m.storeMFAChallenge(ctx, challenge); err != nil {
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

func (m *MFAService) createSMSChallenge(ctx context.Context, user *domain.User) (*MFAChallengeResponse, error) {
	code := m.generateSMSCode()
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

	if err := m.storeMFAChallenge(ctx, challenge); err != nil {
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

func (m *MFAService) createEmailChallenge(ctx context.Context, user *domain.User) (*MFAChallengeResponse, error) {
	code := m.generateEmailCode()
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

	if err := m.storeMFAChallenge(ctx, challenge); err != nil {
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

func (m *MFAService) generateBackupCodes() []string {
	codes := make([]string, m.config.BackupCodesCount)
	for i := 0; i < m.config.BackupCodesCount; i++ {
		codes[i] = generateRandomCode(m.config.BackupCodeLength)
	}
	return codes
}

func (m *MFAService) generateSMSCode() string {
	return generateNumericCode(m.config.SMSCodeLength)
}

func (m *MFAService) generateEmailCode() string {
	return generateNumericCode(m.config.EmailCodeLength)
}

func (m *MFAService) verifyChallengeCode(ctx context.Context, userID, code string) (bool, error) {
	// TODO: Implement challenge storage and verification
	// This would typically use Redis or database to store temporary challenges
	return false, fmt.Errorf("challenge verification not implemented")
}

func (m *MFAService) storeMFAChallenge(ctx context.Context, challenge *MFAChallenge) error {
	// TODO: Implement challenge storage
	// This would typically store in Redis with expiration
	return nil
}

func (m *MFAService) logSecurityEvent(ctx context.Context, userID, eventType string, severity monitoring.SecuritySeverity, metadata map[string]interface{}) {
	event := &monitoring.SecurityEvent{
		EventType:   monitoring.SecurityEventType(eventType),
		Severity:    severity,
		Source:      "mfa-service",
		UserID:      userID,
		Description: eventType,
		Metadata:    metadata,
	}

	m.monitoringService.LogSecurityEvent(ctx, event)
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
