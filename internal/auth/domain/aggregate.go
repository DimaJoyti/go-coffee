package domain

import (
	"errors"
	"time"
)

// UserAggregate represents the user aggregate root with enhanced business logic
type UserAggregate struct {
	*User
	passwordRule *PasswordComplexityRule
	lockoutRule  *AccountLockoutRule
	sessionRule  *SessionRule
	securityRule *SecurityRule
	violations   []*RuleViolation
}

// NewUserAggregate creates a new user aggregate with business rules
func NewUserAggregate(email, passwordHash string, role UserRole) (*UserAggregate, error) {
	user, err := NewUser(email, passwordHash, role)
	if err != nil {
		return nil, err
	}

	return &UserAggregate{
		User:         user,
		passwordRule: DefaultPasswordComplexityRule(),
		lockoutRule:  DefaultAccountLockoutRule(),
		sessionRule:  DefaultSessionRule(),
		securityRule: DefaultSecurityRule(),
		violations:   make([]*RuleViolation, 0),
	}, nil
}

// LoadUserAggregate loads an existing user into an aggregate
func LoadUserAggregate(user *User) *UserAggregate {
	return &UserAggregate{
		User:         user,
		passwordRule: DefaultPasswordComplexityRule(),
		lockoutRule:  DefaultAccountLockoutRule(),
		sessionRule:  DefaultSessionRule(),
		securityRule: DefaultSecurityRule(),
		violations:   make([]*RuleViolation, 0),
	}
}

// Business Logic Methods

// ChangePassword changes the user's password with validation
func (ua *UserAggregate) ChangePassword(newPassword string, forced bool) error {
	// Validate password complexity
	if err := ua.passwordRule.ValidatePassword(newPassword); err != nil {
		violation := NewRuleViolation("password_complexity", err.Error(), "error")
		ua.violations = append(ua.violations, violation)
		return err
	}

	// Check password complexity score
	score := ua.passwordRule.CalculatePasswordComplexityScore(newPassword)
	if score < MinPasswordComplexityScore {
		violation := NewRuleViolation("password_complexity", "password complexity score too low", "warning")
		ua.violations = append(ua.violations, violation)
		return errors.New("password does not meet complexity requirements")
	}

	// Update password
	ua.User.PasswordHash = newPassword // In real implementation, this would be hashed
	now := time.Now()
	ua.User.LastPasswordChange = &now
	ua.User.UpdatedAt = now

	// Generate password changed event
	event := CreatePasswordChangedEvent(ua.User.ID, forced)
	ua.User.AddEvent(*event)

	return nil
}

// AttemptLogin processes a login attempt with business rules
func (ua *UserAggregate) AttemptLogin(ipAddress, userAgent, sessionID string, mfaUsed bool) error {
	// Check if account is locked
	if ua.User.Status == UserStatusLocked {
		if ua.User.LockedUntil != nil && time.Now().Before(*ua.User.LockedUntil) {
			violation := NewRuleViolation("account_locked", "account is locked", "error")
			ua.violations = append(ua.violations, violation)
			return errors.New("account is locked")
		}
		// Auto-unlock if lock period has expired
		ua.User.Unlock()
	}

	// Check if account is active
	if ua.User.Status != UserStatusActive {
		violation := NewRuleViolation("account_status", "account is not active", "error")
		ua.violations = append(ua.violations, violation)
		return errors.New("account is not active")
	}

	// Check if MFA is required
	if ua.RequiresMFA() && !mfaUsed {
		violation := NewRuleViolation("mfa_required", "MFA is required for this account", "error")
		ua.violations = append(ua.violations, violation)
		return errors.New("MFA is required")
	}

	// Record successful login
	ua.User.RecordSuccessfulLogin(ipAddress, userAgent, sessionID, mfaUsed)

	return nil
}

// RecordFailedLogin records a failed login attempt with lockout logic
func (ua *UserAggregate) RecordFailedLogin(ipAddress, userAgent string) {
	ua.User.IncrementFailedLogin()

	// Check if account should be locked
	if ua.lockoutRule.ShouldLockAccount(ua.User.FailedLoginCount, ua.User.LastFailedLoginAt) {
		lockUntil := time.Now().Add(ua.lockoutRule.GetLockoutDuration())
		ua.User.Lock(lockUntil, "too many failed login attempts")

		violation := NewRuleViolation("account_lockout", "account locked due to failed login attempts", "critical")
		ua.violations = append(ua.violations, violation)
	}

	// Generate failed login event
	event := CreateSecurityEvent(
		ua.User.ID,
		"failed_login",
		"medium",
		"Failed login attempt",
		ipAddress,
		userAgent,
		map[string]string{
			"failed_count": string(rune(ua.User.FailedLoginCount)),
		},
	)
	ua.User.AddEvent(*event)
}

// EnableMFA enables multi-factor authentication
func (ua *UserAggregate) EnableMFA(method MFAMethod, secret string, backupCodes []string) error {
	if ua.User.MFAEnabled {
		return errors.New("MFA is already enabled")
	}

	ua.User.MFAEnabled = true
	ua.User.MFAMethod = method
	ua.User.MFASecret = secret
	ua.User.MFABackupCodes = backupCodes
	ua.User.UpdatedAt = time.Now()

	// Update security level if needed
	if ua.User.SecurityLevel == SecurityLevelLow {
		ua.User.SecurityLevel = SecurityLevelMedium
	}

	// Generate MFA enabled event
	event := NewDomainEvent(EventTypeMFAEnabled, ua.User.ID, map[string]interface{}{
		"user_id":   ua.User.ID,
		"method":    method,
		"timestamp": time.Now(),
	})
	ua.User.AddEvent(*event)

	return nil
}

// DisableMFA disables multi-factor authentication
func (ua *UserAggregate) DisableMFA() error {
	if !ua.User.MFAEnabled {
		return errors.New("MFA is not enabled")
	}

	ua.User.MFAEnabled = false
	ua.User.MFAMethod = ""
	ua.User.MFASecret = ""
	ua.User.MFABackupCodes = nil
	ua.User.UpdatedAt = time.Now()

	// Generate MFA disabled event
	event := NewDomainEvent(EventTypeMFADisabled, ua.User.ID, map[string]interface{}{
		"user_id":   ua.User.ID,
		"timestamp": time.Now(),
	})
	ua.User.AddEvent(*event)

	return nil
}

// UpdateRiskScore updates the user's risk score
func (ua *UserAggregate) UpdateRiskScore(newScore float64, factors []string) {
	oldScore := ua.User.RiskScore
	oldLevel := ua.User.SecurityLevel

	ua.User.RiskScore = newScore
	ua.User.UpdatedAt = time.Now()

	// Update security level based on risk score
	newLevel := ua.calculateSecurityLevel(newScore)
	if newLevel != oldLevel {
		ua.User.SecurityLevel = newLevel
	}

	// Generate risk score updated event
	event := NewDomainEvent(EventTypeRiskScoreUpdated, ua.User.ID, map[string]interface{}{
		"user_id":   ua.User.ID,
		"old_score": oldScore,
		"new_score": newScore,
		"old_level": oldLevel,
		"new_level": newLevel,
		"factors":   factors,
		"timestamp": time.Now(),
	})
	ua.User.AddEvent(*event)
}

// AddDevice adds a new device to the user's trusted devices
func (ua *UserAggregate) AddDevice(fingerprint DeviceFingerprint) error {
	// Check if device limit is reached
	if !ua.securityRule.CanAddDevice(len(ua.User.DeviceFingerprints)) {
		violation := NewRuleViolation("device_limit", "maximum number of devices reached", "warning")
		ua.violations = append(ua.violations, violation)
		return errors.New("maximum number of devices reached")
	}

	// Check if device already exists
	for _, device := range ua.User.DeviceFingerprints {
		if device.Fingerprint == fingerprint.Fingerprint {
			return errors.New("device already exists")
		}
	}

	ua.User.DeviceFingerprints = append(ua.User.DeviceFingerprints, fingerprint)
	ua.User.UpdatedAt = time.Now()

	// Generate device added event
	event := NewDomainEvent(EventTypeDeviceAdded, ua.User.ID, map[string]interface{}{
		"user_id":     ua.User.ID,
		"device_id":   fingerprint.ID,
		"fingerprint": fingerprint.Fingerprint,
		"user_agent":  fingerprint.UserAgent,
		"ip_address":  fingerprint.IPAddress,
		"timestamp":   time.Now(),
	})
	ua.User.AddEvent(*event)

	return nil
}

// VerifyEmail verifies the user's email address
func (ua *UserAggregate) VerifyEmail() {
	if ua.User.IsEmailVerified {
		return
	}

	ua.User.VerifyEmail()

	// Generate email verified event
	event := NewDomainEvent(EventTypeUserEmailVerified, ua.User.ID, map[string]interface{}{
		"user_id":   ua.User.ID,
		"email":     ua.User.Email,
		"timestamp": time.Now(),
	})
	ua.User.AddEvent(*event)
}

// VerifyPhone verifies the user's phone number
func (ua *UserAggregate) VerifyPhone() {
	if ua.User.IsPhoneVerified {
		return
	}

	ua.User.VerifyPhone()

	// Generate phone verified event
	event := NewDomainEvent(EventTypeUserPhoneVerified, ua.User.ID, map[string]interface{}{
		"user_id":      ua.User.ID,
		"phone_number": ua.User.PhoneNumber,
		"timestamp":    time.Now(),
	})
	ua.User.AddEvent(*event)
}

// Helper Methods

// calculateSecurityLevel calculates security level based on risk score
func (ua *UserAggregate) calculateSecurityLevel(riskScore float64) SecurityLevel {
	switch {
	case riskScore >= 0.8:
		return SecurityLevelHigh
	case riskScore >= 0.4:
		return SecurityLevelMedium
	default:
		return SecurityLevelLow
	}
}

// GetViolations returns all business rule violations
func (ua *UserAggregate) GetViolations() []*RuleViolation {
	return ua.violations
}

// ClearViolations clears all business rule violations
func (ua *UserAggregate) ClearViolations() {
	ua.violations = make([]*RuleViolation, 0)
}

// HasViolations returns true if there are any violations
func (ua *UserAggregate) HasViolations() bool {
	return len(ua.violations) > 0
}
