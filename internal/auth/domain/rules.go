package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// Business Rules for Authentication Domain

// Password Rules
const (
	MinPasswordLength = 8
	MaxPasswordLength = 128
	MaxFailedAttempts = 5
	AccountLockDuration = 30 * time.Minute
	PasswordExpiryDays = 90
	MinPasswordComplexityScore = 3
)

// Session Rules
const (
	MaxConcurrentSessions = 5
	SessionIdleTimeout = 30 * time.Minute
	SessionAbsoluteTimeout = 24 * time.Hour
	RefreshTokenRotationThreshold = 7 * 24 * time.Hour
)

// Security Rules
const (
	MaxDevicesPerUser = 10
	SuspiciousActivityThreshold = 3
	RiskScoreThreshold = 0.8
	MFARequiredRiskScore = 0.5
)

// Email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Password Rules

// PasswordComplexityRule defines password complexity requirements
type PasswordComplexityRule struct {
	MinLength    int
	MaxLength    int
	RequireUpper bool
	RequireLower bool
	RequireDigit bool
	RequireSpecial bool
	ForbiddenPatterns []string
}

// DefaultPasswordComplexityRule returns the default password complexity rule
func DefaultPasswordComplexityRule() *PasswordComplexityRule {
	return &PasswordComplexityRule{
		MinLength:    MinPasswordLength,
		MaxLength:    MaxPasswordLength,
		RequireUpper: true,
		RequireLower: true,
		RequireDigit: true,
		RequireSpecial: true,
		ForbiddenPatterns: []string{
			"password", "123456", "qwerty", "admin", "user",
			"login", "welcome", "secret", "default",
		},
	}
}

// ValidatePassword validates a password against complexity rules
func (r *PasswordComplexityRule) ValidatePassword(password string) error {
	if len(password) < r.MinLength {
		return fmt.Errorf("password must be at least %d characters long", r.MinLength)
	}

	if len(password) > r.MaxLength {
		return fmt.Errorf("password must be no more than %d characters long", r.MaxLength)
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if r.RequireUpper && !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	if r.RequireLower && !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	if r.RequireDigit && !hasDigit {
		return errors.New("password must contain at least one digit")
	}

	if r.RequireSpecial && !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	// Check for forbidden patterns
	passwordLower := strings.ToLower(password)
	for _, pattern := range r.ForbiddenPatterns {
		if strings.Contains(passwordLower, strings.ToLower(pattern)) {
			return fmt.Errorf("password contains forbidden pattern: %s", pattern)
		}
	}

	return nil
}

// CalculatePasswordComplexityScore calculates a complexity score (0-5)
func (r *PasswordComplexityRule) CalculatePasswordComplexityScore(password string) int {
	score := 0
	
	if len(password) >= r.MinLength {
		score++
	}
	
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	
	if hasUpper {
		score++
	}
	if hasLower {
		score++
	}
	if hasDigit {
		score++
	}
	if hasSpecial {
		score++
	}
	
	return score
}

// Account Lockout Rules

// AccountLockoutRule defines account lockout behavior
type AccountLockoutRule struct {
	MaxFailedAttempts int
	LockoutDuration   time.Duration
	ResetWindow       time.Duration
}

// DefaultAccountLockoutRule returns the default account lockout rule
func DefaultAccountLockoutRule() *AccountLockoutRule {
	return &AccountLockoutRule{
		MaxFailedAttempts: MaxFailedAttempts,
		LockoutDuration:   AccountLockDuration,
		ResetWindow:       time.Hour,
	}
}

// ShouldLockAccount determines if an account should be locked
func (r *AccountLockoutRule) ShouldLockAccount(failedAttempts int, lastFailedAt *time.Time) bool {
	if failedAttempts < r.MaxFailedAttempts {
		return false
	}

	// If last failed attempt was within the reset window, lock the account
	if lastFailedAt != nil && time.Since(*lastFailedAt) < r.ResetWindow {
		return true
	}

	return false
}

// GetLockoutDuration returns the lockout duration
func (r *AccountLockoutRule) GetLockoutDuration() time.Duration {
	return r.LockoutDuration
}

// Session Management Rules

// SessionRule defines session management rules
type SessionRule struct {
	MaxConcurrentSessions int
	IdleTimeout          time.Duration
	AbsoluteTimeout      time.Duration
	RequireMFAForExtended bool
}

// DefaultSessionRule returns the default session rule
func DefaultSessionRule() *SessionRule {
	return &SessionRule{
		MaxConcurrentSessions: MaxConcurrentSessions,
		IdleTimeout:          SessionIdleTimeout,
		AbsoluteTimeout:      SessionAbsoluteTimeout,
		RequireMFAForExtended: true,
	}
}

// CanCreateSession determines if a new session can be created
func (r *SessionRule) CanCreateSession(activeSessions int) bool {
	return activeSessions < r.MaxConcurrentSessions
}

// IsSessionExpired checks if a session is expired
func (r *SessionRule) IsSessionExpired(lastUsed, createdAt time.Time) bool {
	now := time.Now()
	
	// Check idle timeout
	if now.Sub(lastUsed) > r.IdleTimeout {
		return true
	}
	
	// Check absolute timeout
	if now.Sub(createdAt) > r.AbsoluteTimeout {
		return true
	}
	
	return false
}

// Security Rules

// SecurityRule defines security-related rules
type SecurityRule struct {
	MaxDevicesPerUser           int
	SuspiciousActivityThreshold int
	RiskScoreThreshold         float64
	MFARequiredRiskScore       float64
}

// DefaultSecurityRule returns the default security rule
func DefaultSecurityRule() *SecurityRule {
	return &SecurityRule{
		MaxDevicesPerUser:           MaxDevicesPerUser,
		SuspiciousActivityThreshold: SuspiciousActivityThreshold,
		RiskScoreThreshold:         RiskScoreThreshold,
		MFARequiredRiskScore:       MFARequiredRiskScore,
	}
}

// RequiresMFA determines if MFA is required based on risk score
func (r *SecurityRule) RequiresMFA(riskScore float64, securityLevel SecurityLevel) bool {
	if securityLevel == SecurityLevelHigh {
		return true
	}
	
	return riskScore >= r.MFARequiredRiskScore
}

// CanAddDevice determines if a new device can be added
func (r *SecurityRule) CanAddDevice(deviceCount int) bool {
	return deviceCount < r.MaxDevicesPerUser
}

// IsSuspiciousActivity determines if activity is suspicious
func (r *SecurityRule) IsSuspiciousActivity(failedAttempts int, riskScore float64) bool {
	return failedAttempts >= r.SuspiciousActivityThreshold || riskScore >= r.RiskScoreThreshold
}

// Validation Rules

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	
	if len(email) > 254 {
		return errors.New("email is too long")
	}
	
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	
	return nil
}

// ValidatePhoneNumber validates a phone number (basic validation)
func ValidatePhoneNumber(phone string) error {
	if phone == "" {
		return nil // Phone is optional
	}
	
	// Remove common formatting characters
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	cleaned = strings.ReplaceAll(cleaned, "+", "")
	
	if len(cleaned) < 10 || len(cleaned) > 15 {
		return errors.New("phone number must be between 10 and 15 digits")
	}
	
	// Check if all characters are digits
	for _, char := range cleaned {
		if !unicode.IsDigit(char) {
			return errors.New("phone number must contain only digits")
		}
	}
	
	return nil
}

// Business Rule Violations

// RuleViolation represents a business rule violation
type RuleViolation struct {
	Rule        string    `json:"rule"`
	Message     string    `json:"message"`
	Severity    string    `json:"severity"`
	Timestamp   time.Time `json:"timestamp"`
}

// NewRuleViolation creates a new rule violation
func NewRuleViolation(rule, message, severity string) *RuleViolation {
	return &RuleViolation{
		Rule:      rule,
		Message:   message,
		Severity:  severity,
		Timestamp: time.Now(),
	}
}
