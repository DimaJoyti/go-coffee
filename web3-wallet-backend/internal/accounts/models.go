package accounts

import (
	"time"
)

// Account represents a user account in the fintech platform
type Account struct {
	ID                string                  `json:"id" db:"id"`
	UserID            string                  `json:"user_id" db:"user_id"`
	Email             string                  `json:"email" db:"email"`
	Phone             string                  `json:"phone" db:"phone"`
	FirstName         string                  `json:"first_name" db:"first_name"`
	LastName          string                  `json:"last_name" db:"last_name"`
	DateOfBirth       *time.Time              `json:"date_of_birth" db:"date_of_birth"`
	Nationality       string                  `json:"nationality" db:"nationality"`
	Country           string                  `json:"country" db:"country"`
	State             string                  `json:"state" db:"state"`
	City              string                  `json:"city" db:"city"`
	Address           string                  `json:"address" db:"address"`
	PostalCode        string                  `json:"postal_code" db:"postal_code"`
	AccountType       AccountType             `json:"account_type" db:"account_type"`
	AccountStatus     AccountStatus           `json:"account_status" db:"account_status"`
	KYCStatus         KYCStatus               `json:"kyc_status" db:"kyc_status"`
	KYCLevel          KYCLevel                `json:"kyc_level" db:"kyc_level"`
	RiskScore         float64                 `json:"risk_score" db:"risk_score"`
	ComplianceFlags   []string                `json:"compliance_flags" db:"compliance_flags"`
	TwoFactorEnabled  bool                    `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorMethod   *TwoFactorMethod        `json:"two_factor_method" db:"two_factor_method"`
	LastLoginAt       *time.Time              `json:"last_login_at" db:"last_login_at"`
	LastLoginIP       string                  `json:"last_login_ip" db:"last_login_ip"`
	FailedLoginCount  int                     `json:"failed_login_count" db:"failed_login_count"`
	AccountLimits     AccountLimits           `json:"account_limits" db:"account_limits"`
	NotificationPrefs NotificationPreferences `json:"notification_preferences" db:"notification_preferences"`
	Metadata          map[string]interface{}  `json:"metadata" db:"metadata"`
	CreatedAt         time.Time               `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time               `json:"updated_at" db:"updated_at"`
	DeletedAt         *time.Time              `json:"deleted_at" db:"deleted_at"`
}

// AccountType represents the type of account
type AccountType string

const (
	AccountTypePersonal   AccountType = "personal"
	AccountTypeBusiness   AccountType = "business"
	AccountTypeEnterprise AccountType = "enterprise"
)

// AccountStatus represents the status of an account
type AccountStatus string

const (
	AccountStatusActive    AccountStatus = "active"
	AccountStatusInactive  AccountStatus = "inactive"
	AccountStatusSuspended AccountStatus = "suspended"
	AccountStatusClosed    AccountStatus = "closed"
	AccountStatusPending   AccountStatus = "pending"
)

// KYCStatus represents the KYC verification status
type KYCStatus string

const (
	KYCStatusNotStarted KYCStatus = "not_started"
	KYCStatusPending    KYCStatus = "pending"
	KYCStatusInReview   KYCStatus = "in_review"
	KYCStatusApproved   KYCStatus = "approved"
	KYCStatusRejected   KYCStatus = "rejected"
	KYCStatusExpired    KYCStatus = "expired"
)

// KYCLevel represents the level of KYC verification
type KYCLevel string

const (
	KYCLevelNone     KYCLevel = "none"
	KYCLevelBasic    KYCLevel = "basic"
	KYCLevelStandard KYCLevel = "standard"
	KYCLevelEnhanced KYCLevel = "enhanced"
)

// AccountLimits represents account transaction limits
type AccountLimits struct {
	DailyTransactionLimit   string `json:"daily_transaction_limit" db:"daily_transaction_limit"`
	MonthlyTransactionLimit string `json:"monthly_transaction_limit" db:"monthly_transaction_limit"`
	SingleTransactionLimit  string `json:"single_transaction_limit" db:"single_transaction_limit"`
	MaxWallets              int    `json:"max_wallets" db:"max_wallets"`
	MaxCards                int    `json:"max_cards" db:"max_cards"`
	WithdrawalLimit         string `json:"withdrawal_limit" db:"withdrawal_limit"`
	DepositLimit            string `json:"deposit_limit" db:"deposit_limit"`
}

// NotificationPreferences represents user notification preferences
type NotificationPreferences struct {
	EmailEnabled      bool `json:"email_enabled" db:"email_enabled"`
	SMSEnabled        bool `json:"sms_enabled" db:"sms_enabled"`
	PushEnabled       bool `json:"push_enabled" db:"push_enabled"`
	SecurityAlerts    bool `json:"security_alerts" db:"security_alerts"`
	TransactionAlerts bool `json:"transaction_alerts" db:"transaction_alerts"`
	MarketingEmails   bool `json:"marketing_emails" db:"marketing_emails"`
	ProductUpdates    bool `json:"product_updates" db:"product_updates"`
	WeeklyReports     bool `json:"weekly_reports" db:"weekly_reports"`
	MonthlyStatements bool `json:"monthly_statements" db:"monthly_statements"`
}

// KYCDocument represents a KYC document
type KYCDocument struct {
	ID           string                 `json:"id" db:"id"`
	AccountID    string                 `json:"account_id" db:"account_id"`
	DocumentType DocumentType           `json:"document_type" db:"document_type"`
	DocumentURL  string                 `json:"document_url" db:"document_url"`
	Status       DocumentStatus         `json:"status" db:"status"`
	UploadedAt   time.Time              `json:"uploaded_at" db:"uploaded_at"`
	VerifiedAt   *time.Time             `json:"verified_at" db:"verified_at"`
	ExpiresAt    *time.Time             `json:"expires_at" db:"expires_at"`
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
}

// DocumentType represents the type of KYC document
type DocumentType string

const (
	DocumentTypePassport       DocumentType = "passport"
	DocumentTypeDriversLicense DocumentType = "drivers_license"
	DocumentTypeNationalID     DocumentType = "national_id"
	DocumentTypeUtilityBill    DocumentType = "utility_bill"
	DocumentTypeBankStatement  DocumentType = "bank_statement"
	DocumentTypeProofOfAddress DocumentType = "proof_of_address"
	DocumentTypeSelfie         DocumentType = "selfie"
)

// DocumentStatus represents the status of a KYC document
type DocumentStatus string

const (
	DocumentStatusPending  DocumentStatus = "pending"
	DocumentStatusApproved DocumentStatus = "approved"
	DocumentStatusRejected DocumentStatus = "rejected"
	DocumentStatusExpired  DocumentStatus = "expired"
)

// Session represents a user session
type Session struct {
	ID           string                 `json:"id" db:"id"`
	AccountID    string                 `json:"account_id" db:"account_id"`
	DeviceID     string                 `json:"device_id" db:"device_id"`
	IPAddress    string                 `json:"ip_address" db:"ip_address"`
	UserAgent    string                 `json:"user_agent" db:"user_agent"`
	Location     string                 `json:"location" db:"location"`
	SessionToken string                 `json:"session_token" db:"session_token"`
	RefreshToken string                 `json:"refresh_token" db:"refresh_token"`
	ExpiresAt    time.Time              `json:"expires_at" db:"expires_at"`
	IsActive     bool                   `json:"is_active" db:"is_active"`
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	ID          string                 `json:"id" db:"id"`
	AccountID   string                 `json:"account_id" db:"account_id"`
	EventType   SecurityEventType      `json:"event_type" db:"event_type"`
	Severity    SecuritySeverity       `json:"severity" db:"severity"`
	Description string                 `json:"description" db:"description"`
	IPAddress   string                 `json:"ip_address" db:"ip_address"`
	UserAgent   string                 `json:"user_agent" db:"user_agent"`
	Location    string                 `json:"location" db:"location"`
	Resolved    bool                   `json:"resolved" db:"resolved"`
	ResolvedAt  *time.Time             `json:"resolved_at" db:"resolved_at"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// SecurityEventType represents the type of security event
type SecurityEventType string

const (
	SecurityEventTypeLogin               SecurityEventType = "login"
	SecurityEventTypeFailedLogin         SecurityEventType = "failed_login"
	SecurityEventTypePasswordChange      SecurityEventType = "password_change"
	SecurityEventTypePasswordReset       SecurityEventType = "password_reset"
	SecurityEventTypeFailedPasswordReset SecurityEventType = "failed_password_reset"
	SecurityEventTypeTwoFactorEnabled    SecurityEventType = "two_factor_enabled"
	SecurityEventTypeSuspiciousActivity  SecurityEventType = "suspicious_activity"
	SecurityEventTypeAccountLocked       SecurityEventType = "account_locked"
	SecurityEventTypeDeviceAdded         SecurityEventType = "device_added"
	SecurityEventTypeLocationChange      SecurityEventType = "location_change"
	SecurityEventTypeDataAccess          SecurityEventType = "data_access"
)

// SecuritySeverity represents the severity of a security event
type SecuritySeverity string

const (
	SecuritySeverityLow      SecuritySeverity = "low"
	SecuritySeverityMedium   SecuritySeverity = "medium"
	SecuritySeverityHigh     SecuritySeverity = "high"
	SecuritySeverityCritical SecuritySeverity = "critical"
)

// CreateAccountRequest represents a request to create a new account
type CreateAccountRequest struct {
	Email       string      `json:"email" validate:"required,email"`
	Phone       string      `json:"phone" validate:"required"`
	FirstName   string      `json:"first_name" validate:"required"`
	LastName    string      `json:"last_name" validate:"required"`
	Password    string      `json:"password" validate:"required,min=8"`
	AccountType AccountType `json:"account_type" validate:"required"`
	Country     string      `json:"country" validate:"required"`
	AcceptTerms bool        `json:"accept_terms" validate:"required"`
}

// UpdateAccountRequest represents a request to update an account
type UpdateAccountRequest struct {
	FirstName         *string                  `json:"first_name,omitempty"`
	LastName          *string                  `json:"last_name,omitempty"`
	Phone             *string                  `json:"phone,omitempty"`
	Country           *string                  `json:"country,omitempty"`
	State             *string                  `json:"state,omitempty"`
	City              *string                  `json:"city,omitempty"`
	Address           *string                  `json:"address,omitempty"`
	PostalCode        *string                  `json:"postal_code,omitempty"`
	NotificationPrefs *NotificationPreferences `json:"notification_preferences,omitempty"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	DeviceID   string `json:"device_id"`
	RememberMe bool   `json:"remember_me"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Account      *Account `json:"account"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int64    `json:"expires_in"`
	TokenType    string   `json:"token_type"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword         string `json:"current_password" validate:"required"`
	NewPassword             string `json:"new_password" validate:"required,min=8"`
	InvalidateOtherSessions bool   `json:"invalidate_other_sessions"`
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ConfirmPasswordResetRequest represents a password reset confirmation
type ConfirmPasswordResetRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// TwoFactorMethod represents the two-factor authentication method
type TwoFactorMethod string

const (
	TwoFactorMethodSMS   TwoFactorMethod = "sms"
	TwoFactorMethodTOTP  TwoFactorMethod = "totp"
	TwoFactorMethodEmail TwoFactorMethod = "email"
)

// EnableTwoFactorRequest represents a request to enable 2FA
type EnableTwoFactorRequest struct {
	Method TwoFactorMethod `json:"method" validate:"required"`
	Phone  string          `json:"phone,omitempty"`
}

// VerifyTwoFactorRequest represents a request to verify 2FA
type VerifyTwoFactorRequest struct {
	Code string `json:"code" validate:"required"`
}

// KYCSubmissionRequest represents a KYC submission request
type KYCSubmissionRequest struct {
	DocumentType DocumentType `json:"document_type" validate:"required"`
	DocumentURL  string       `json:"document_url" validate:"required"`
}

// AccountListRequest represents a request to list accounts
type AccountListRequest struct {
	Page       int           `json:"page" validate:"min=1"`
	Limit      int           `json:"limit" validate:"min=1,max=100"`
	Status     AccountStatus `json:"status,omitempty"`
	KYCStatus  KYCStatus     `json:"kyc_status,omitempty"`
	Country    string        `json:"country,omitempty"`
	SearchTerm string        `json:"search_term,omitempty"`
}

// AccountListResponse represents a response to list accounts
type AccountListResponse struct {
	Accounts   []Account `json:"accounts"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	TotalPages int       `json:"total_pages"`
}
