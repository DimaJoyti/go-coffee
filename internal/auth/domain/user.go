package domain

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// UserStatus represents the status of a user account
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusLocked    UserStatus = "locked"
	UserStatusSuspended UserStatus = "suspended"
)

// UserRole represents the role of a user
type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

// MFAMethod represents the multi-factor authentication method
type MFAMethod string

const (
	MFAMethodNone   MFAMethod = "none"
	MFAMethodTOTP   MFAMethod = "totp"
	MFAMethodSMS    MFAMethod = "sms"
	MFAMethodEmail  MFAMethod = "email"
	MFAMethodBackup MFAMethod = "backup"
)

// SecurityLevel represents the security level of a user
type SecurityLevel string

const (
	SecurityLevelLow    SecurityLevel = "low"
	SecurityLevelMedium SecurityLevel = "medium"
	SecurityLevelHigh   SecurityLevel = "high"
)

// DeviceFingerprint represents a device fingerprint
type DeviceFingerprint struct {
	ID          string    `json:"id"`
	Fingerprint string    `json:"fingerprint"`
	UserAgent   string    `json:"user_agent"`
	IPAddress   string    `json:"ip_address"`
	Location    string    `json:"location,omitempty"`
	Trusted     bool      `json:"trusted"`
	LastUsed    time.Time `json:"last_used"`
	CreatedAt   time.Time `json:"created_at"`
}

// User represents a user entity in the domain
type User struct {
	AggregateRoot // Embed aggregate root for event functionality

	ID                string            `json:"id"`
	Email             string            `json:"email"`
	PasswordHash      string            `json:"-"` // Never serialize password hash
	Role              UserRole          `json:"role"`
	Status            UserStatus        `json:"status"`
	FailedLoginCount  int               `json:"failed_login_count"`
	LastLoginAt       *time.Time        `json:"last_login_at,omitempty"`
	LastFailedLoginAt *time.Time        `json:"last_failed_login_at,omitempty"`
	LockedUntil       *time.Time        `json:"locked_until,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`

	// MFA fields
	MFAEnabled      bool      `json:"mfa_enabled"`
	MFASecret       string    `json:"-"` // Never serialize MFA secret
	MFABackupCodes  []string  `json:"-"` // Never serialize backup codes
	MFAMethod       MFAMethod `json:"mfa_method"`
	PhoneNumber     string    `json:"phone_number,omitempty"`
	IsPhoneVerified bool      `json:"is_phone_verified"`
	IsEmailVerified bool      `json:"is_email_verified"`

	// Security fields
	SecurityLevel      SecurityLevel       `json:"security_level"`
	RiskScore          float64             `json:"risk_score"`
	LastPasswordChange *time.Time          `json:"last_password_change,omitempty"`
	DeviceFingerprints []DeviceFingerprint `json:"device_fingerprints,omitempty"`
}

// UserValidationErrors
var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrEmailRequired   = errors.New("email is required")
	ErrPasswordTooWeak = errors.New("password does not meet security requirements")
	ErrUserNotFound    = errors.New("user not found")
	ErrUserExists      = errors.New("user already exists")
	ErrUserLocked      = errors.New("user account is locked")
	ErrUserInactive    = errors.New("user account is inactive")
)

// Email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// NewUser creates a new user with validation
func NewUser(email, passwordHash string, role UserRole) (*User, error) {
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}

	if passwordHash == "" {
		return nil, errors.New("password hash is required")
	}

	now := time.Now()
	user := &User{
		ID:            uuid.New().String(),
		Email:         email,
		PasswordHash:  passwordHash,
		Role:          role,
		Status:        UserStatusActive,
		Metadata:      make(map[string]string),
		CreatedAt:     now,
		UpdatedAt:     now,
		SecurityLevel: SecurityLevelLow,
		RiskScore:     0.0,
	}

	// Generate user registered event
	event := CreateUserRegisteredEvent(user.ID, user.Email, user.Role)
	user.AddEvent(*event)

	return user, nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return ErrEmailRequired
	}
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

// IsLocked checks if the user account is locked
func (u *User) IsLocked() bool {
	if u.Status == UserStatusLocked {
		return true
	}
	if u.LockedUntil != nil && time.Now().Before(*u.LockedUntil) {
		return true
	}
	return false
}

// IsActive checks if the user account is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive && !u.IsLocked()
}

// Lock locks the user account until the specified time
func (u *User) Lock(until time.Time, reason string) {
	u.Status = UserStatusLocked
	u.LockedUntil = &until
	u.UpdatedAt = time.Now()

	// Generate user locked event
	event := CreateUserLockedEvent(u.ID, reason, until)
	u.AddEvent(*event)
}

// Unlock unlocks the user account
func (u *User) Unlock() {
	u.Status = UserStatusActive
	u.LockedUntil = nil
	u.FailedLoginCount = 0
	u.UpdatedAt = time.Now()
}

// IncrementFailedLogin increments the failed login count
func (u *User) IncrementFailedLogin() {
	u.FailedLoginCount++
	now := time.Now()
	u.LastFailedLoginAt = &now
	u.UpdatedAt = now
}

// ResetFailedLogin resets the failed login count
func (u *User) ResetFailedLogin() {
	u.FailedLoginCount = 0
	u.LastFailedLoginAt = nil
	u.UpdatedAt = time.Now()
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

// RecordSuccessfulLogin records a successful login with event generation
func (u *User) RecordSuccessfulLogin(ipAddress, userAgent, sessionID string, mfaUsed bool) {
	u.UpdateLastLogin()
	u.ResetFailedLogin()

	// Generate successful login event
	event := CreateUserLoggedInEvent(u.ID, u.Email, ipAddress, userAgent, sessionID, mfaUsed)
	u.AddEvent(*event)
}

// ChangePassword updates the password hash
func (u *User) ChangePassword(newPasswordHash string, forced bool) {
	u.PasswordHash = newPasswordHash
	now := time.Now()
	u.LastPasswordChange = &now
	u.UpdatedAt = now

	// Generate password changed event
	event := CreatePasswordChangedEvent(u.ID, forced)
	u.AddEvent(*event)
}

// EnableMFA enables multi-factor authentication
func (u *User) EnableMFA(method MFAMethod, secret string) {
	u.MFAEnabled = true
	u.MFAMethod = method
	u.MFASecret = secret
	u.UpdatedAt = time.Now()
}

// DisableMFA disables multi-factor authentication
func (u *User) DisableMFA() {
	u.MFAEnabled = false
	u.MFAMethod = MFAMethodNone
	u.MFASecret = ""
	u.MFABackupCodes = nil
	u.UpdatedAt = time.Now()
}

// SetMFABackupCodes sets the MFA backup codes
func (u *User) SetMFABackupCodes(codes []string) {
	u.MFABackupCodes = codes
	u.UpdatedAt = time.Now()
}

// UseMFABackupCode uses one of the MFA backup codes
func (u *User) UseMFABackupCode(code string) bool {
	for i, backupCode := range u.MFABackupCodes {
		if backupCode == code {
			// Remove the used backup code
			u.MFABackupCodes = append(u.MFABackupCodes[:i], u.MFABackupCodes[i+1:]...)
			u.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// AddDeviceFingerprint adds a new device fingerprint
func (u *User) AddDeviceFingerprint(fingerprint, userAgent, ipAddress, location string) {
	deviceFingerprint := DeviceFingerprint{
		ID:          uuid.New().String(),
		Fingerprint: fingerprint,
		UserAgent:   userAgent,
		IPAddress:   ipAddress,
		Location:    location,
		Trusted:     false,
		LastUsed:    time.Now(),
		CreatedAt:   time.Now(),
	}

	u.DeviceFingerprints = append(u.DeviceFingerprints, deviceFingerprint)
	u.UpdatedAt = time.Now()
}

// TrustDevice marks a device as trusted
func (u *User) TrustDevice(deviceID string) bool {
	for i := range u.DeviceFingerprints {
		if u.DeviceFingerprints[i].ID == deviceID {
			u.DeviceFingerprints[i].Trusted = true
			u.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// IsDeviceTrusted checks if a device is trusted
func (u *User) IsDeviceTrusted(fingerprint string) bool {
	for _, device := range u.DeviceFingerprints {
		if device.Fingerprint == fingerprint && device.Trusted {
			return true
		}
	}
	return false
}

// UpdateRiskScore updates the user's risk score
func (u *User) UpdateRiskScore(score float64) {
	u.RiskScore = score

	// Update security level based on risk score
	switch {
	case score >= 0.8:
		u.SecurityLevel = SecurityLevelHigh
	case score >= 0.5:
		u.SecurityLevel = SecurityLevelMedium
	default:
		u.SecurityLevel = SecurityLevelLow
	}

	u.UpdatedAt = time.Now()
}

// RequiresMFA checks if the user requires MFA
func (u *User) RequiresMFA() bool {
	return u.MFAEnabled || u.SecurityLevel == SecurityLevelHigh || u.RiskScore >= 0.7
}

// VerifyPhone marks the phone number as verified
func (u *User) VerifyPhone() {
	u.IsPhoneVerified = true
	u.UpdatedAt = time.Now()
}

// VerifyEmail marks the email as verified
func (u *User) VerifyEmail() {
	u.IsEmailVerified = true
	u.UpdatedAt = time.Now()
}
