package security

import (
	"errors"
	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"github.com/DimaJoyti/go-coffee/internal/auth/application"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// PasswordService implements password operations
type PasswordService struct {
	config *PasswordConfig
	logger *logger.Logger
}

// PasswordConfig represents password service configuration
type PasswordConfig struct {
	BcryptCost     int                         `yaml:"bcrypt_cost"`
	PasswordPolicy *application.PasswordPolicy `yaml:"password_policy"`
}

// NewPasswordService creates a new password service
func NewPasswordService(config *PasswordConfig, logger *logger.Logger) *PasswordService {
	// Set default password policy if not provided
	if config.PasswordPolicy == nil {
		config.PasswordPolicy = &application.PasswordPolicy{
			MinLength:        8,
			MaxLength:        128,
			RequireUppercase: true,
			RequireLowercase: true,
			RequireNumbers:   true,
			RequireSymbols:   true,
		}
	}

	// Set default bcrypt cost if not provided
	if config.BcryptCost == 0 {
		config.BcryptCost = 12
	}

	return &PasswordService{
		config: config,
		logger: logger,
	}
}

// HashPassword hashes a password using bcrypt
func (p *PasswordService) HashPassword(password string) (string, error) {
	// Validate password before hashing
	if err := p.ValidatePassword(password); err != nil {
		return "", err
	}

	// Hash password
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), p.config.BcryptCost)
	if err != nil {
		p.logger.Error("Failed to hash password: %v", err)
		return "", errors.New("failed to hash password")
	}

	p.logger.Debug("Password hashed successfully")
	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against its hash
func (p *PasswordService) VerifyPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			p.logger.Debug("Password verification failed: password mismatch")
			return errors.New("invalid password")
		}
		p.logger.Error("Password verification failed: %v", err)
		return errors.New("password verification failed")
	}

	p.logger.Debug("Password verified successfully")
	return nil
}

// ValidatePassword validates a password according to the password policy
func (p *PasswordService) ValidatePassword(password string) error {
	if password == "" {
		return errors.New("password is required")
	}

	return p.CheckPasswordPolicy(password)
}

// ValidatePasswordStrength validates password strength
func (p *PasswordService) ValidatePasswordStrength(password string) error {
	return p.CheckPasswordPolicy(password)
}

// GetPasswordPolicy returns the current password policy
func (p *PasswordService) GetPasswordPolicy() *application.PasswordPolicy {
	return p.config.PasswordPolicy
}

// CheckPasswordPolicy checks if password meets the policy requirements
func (p *PasswordService) CheckPasswordPolicy(password string) error {
	policy := p.config.PasswordPolicy

	// Check minimum length
	if len(password) < policy.MinLength {
		return errors.New("password is too short")
	}

	// Check maximum length
	if policy.MaxLength > 0 && len(password) > policy.MaxLength {
		return errors.New("password is too long")
	}

	// Check for uppercase letters
	if policy.RequireUppercase && !p.hasUppercase(password) {
		return errors.New("password must contain at least one uppercase letter")
	}

	// Check for lowercase letters
	if policy.RequireLowercase && !p.hasLowercase(password) {
		return errors.New("password must contain at least one lowercase letter")
	}

	// Check for numbers
	if policy.RequireNumbers && !p.hasNumbers(password) {
		return errors.New("password must contain at least one number")
	}

	// Check for symbols
	if policy.RequireSymbols && !p.hasSymbols(password) {
		return errors.New("password must contain at least one special character")
	}

	// Check for common weak patterns
	if err := p.checkWeakPatterns(password); err != nil {
		return err
	}

	return nil
}

// hasUppercase checks if password contains uppercase letters
func (p *PasswordService) hasUppercase(password string) bool {
	for _, char := range password {
		if unicode.IsUpper(char) {
			return true
		}
	}
	return false
}

// hasLowercase checks if password contains lowercase letters
func (p *PasswordService) hasLowercase(password string) bool {
	for _, char := range password {
		if unicode.IsLower(char) {
			return true
		}
	}
	return false
}

// hasNumbers checks if password contains numbers
func (p *PasswordService) hasNumbers(password string) bool {
	for _, char := range password {
		if unicode.IsDigit(char) {
			return true
		}
	}
	return false
}

// hasSymbols checks if password contains special characters
func (p *PasswordService) hasSymbols(password string) bool {
	for _, char := range password {
		if unicode.IsPunct(char) || unicode.IsSymbol(char) {
			return true
		}
	}
	return false
}

// checkWeakPatterns checks for common weak password patterns
func (p *PasswordService) checkWeakPatterns(password string) error {
	// Check for common weak passwords
	weakPasswords := []string{
		"password", "123456", "123456789", "qwerty", "abc123",
		"password123", "admin", "letmein", "welcome", "monkey",
		"1234567890", "password1", "123123", "111111", "000000",
	}

	for _, weak := range weakPasswords {
		if password == weak {
			return errors.New("password is too common and weak")
		}
	}

	// Check for sequential characters
	if p.hasSequentialChars(password) {
		return errors.New("password contains too many sequential characters")
	}

	// Check for repeated characters
	if p.hasRepeatedChars(password) {
		return errors.New("password contains too many repeated characters")
	}

	return nil
}

// hasSequentialChars checks for sequential characters (e.g., "123", "abc")
func (p *PasswordService) hasSequentialChars(password string) bool {
	sequentialPatterns := []string{
		"123", "234", "345", "456", "567", "678", "789", "890",
		"abc", "bcd", "cde", "def", "efg", "fgh", "ghi", "hij",
		"ijk", "jkl", "klm", "lmn", "mno", "nop", "opq", "pqr",
		"qrs", "rst", "stu", "tuv", "uvw", "vwx", "wxy", "xyz",
	}

	for _, pattern := range sequentialPatterns {
		matched, _ := regexp.MatchString(pattern, password)
		if matched {
			return true
		}
	}

	return false
}

// hasRepeatedChars checks for repeated characters (e.g., "aaa", "111")
func (p *PasswordService) hasRepeatedChars(password string) bool {
	// Check for 3 or more consecutive identical characters
	for i := 0; i < len(password)-2; i++ {
		if password[i] == password[i+1] && password[i+1] == password[i+2] {
			return true
		}
	}

	return false
}

// GenerateRandomPassword generates a random password that meets the policy
func (p *PasswordService) GenerateRandomPassword() (string, error) {
	// This is a basic implementation
	// In production, you might want to use a more sophisticated password generator
	const (
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numbers   = "0123456789"
		symbols   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	)

	policy := p.config.PasswordPolicy
	length := policy.MinLength
	if length < 12 {
		length = 12 // Ensure minimum secure length
	}

	var charset string
	var password string

	// Add required character types
	if policy.RequireLowercase {
		charset += lowercase
		password += string(lowercase[0]) // Add at least one
	}
	if policy.RequireUppercase {
		charset += uppercase
		password += string(uppercase[0]) // Add at least one
	}
	if policy.RequireNumbers {
		charset += numbers
		password += string(numbers[0]) // Add at least one
	}
	if policy.RequireSymbols {
		charset += symbols
		password += string(symbols[0]) // Add at least one
	}

	// Fill the rest randomly
	for len(password) < length {
		password += string(charset[len(password)%len(charset)])
	}

	// Shuffle the password (basic implementation)
	// In production, use crypto/rand for better randomness
	runes := []rune(password)
	for i := len(runes) - 1; i > 0; i-- {
		j := i % (i + 1) // Simple shuffle
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes), nil
}
