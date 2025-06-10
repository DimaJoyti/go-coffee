package password

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"unicode"

	"github.com/DimaJoyti/go-coffee/internal/auth/application"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

// BcryptService implements the PasswordService interface using bcrypt
type BcryptService struct {
	config *Config
	logger *logger.Logger
}

// Config represents password service configuration
type Config struct {
	Cost                int      `yaml:"cost"`
	MinLength           int      `yaml:"min_length"`
	MaxLength           int      `yaml:"max_length"`
	RequireUppercase    bool     `yaml:"require_uppercase"`
	RequireLowercase    bool     `yaml:"require_lowercase"`
	RequireNumbers      bool     `yaml:"require_numbers"`
	RequireSpecialChars bool     `yaml:"require_special_chars"`
	SpecialChars        string   `yaml:"special_chars"`
	MaxRepeatingChars   int      `yaml:"max_repeating_chars"`
	ForbiddenPasswords  []string `yaml:"forbidden_passwords"`
}

// PasswordStrength represents password strength levels
type PasswordStrength int

const (
	PasswordStrengthWeak PasswordStrength = iota
	PasswordStrengthMedium
	PasswordStrengthStrong
	PasswordStrengthVeryStrong
)

// PasswordValidationResult represents password validation result
type PasswordValidationResult struct {
	Valid    bool             `json:"valid"`
	Strength PasswordStrength `json:"strength"`
	Score    int              `json:"score"`
	Errors   []string         `json:"errors"`
	Warnings []string         `json:"warnings"`
}

// NewBcryptService creates a new bcrypt password service
func NewBcryptService(config *Config, logger *logger.Logger) application.PasswordService {
	return &BcryptService{
		config: config,
		logger: logger,
	}
}

// HashPassword hashes a password using bcrypt
func (b *BcryptService) HashPassword(password string) (string, error) {
	// Validate password before hashing
	if result := b.validatePasswordInternal(password); !result.Valid {
		return "", fmt.Errorf("password validation failed: %v", result.Errors)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), b.config.Cost)
	if err != nil {
		b.logger.WithError(err).Error("Failed to hash password")
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	b.logger.Debug("Password hashed successfully")
	return string(hash), nil
}

// VerifyPassword verifies a password against its hash
func (b *BcryptService) VerifyPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if err != bcrypt.ErrMismatchedHashAndPassword {
			b.logger.WithError(err).Error("Failed to verify password")
		}
		return err
	}

	b.logger.Debug("Password verified successfully")
	return nil
}

// validatePasswordInternal validates a password against configured rules
func (b *BcryptService) validatePasswordInternal(password string) *PasswordValidationResult {
	result := &PasswordValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Check length
	if len(password) < b.config.MinLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Password must be at least %d characters long", b.config.MinLength))
	}

	if len(password) > b.config.MaxLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Password must be no more than %d characters long", b.config.MaxLength))
	}

	// Check character requirements
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		} else if unicode.IsLower(char) {
			hasLower = true
		} else if unicode.IsNumber(char) {
			hasNumber = true
		} else if strings.ContainsRune(b.config.SpecialChars, char) {
			hasSpecial = true
		}
	}

	if b.config.RequireUppercase && !hasUpper {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must contain at least one uppercase letter")
	}

	if b.config.RequireLowercase && !hasLower {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must contain at least one lowercase letter")
	}

	if b.config.RequireNumbers && !hasNumber {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must contain at least one number")
	}

	if b.config.RequireSpecialChars && !hasSpecial {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Password must contain at least one special character (%s)", b.config.SpecialChars))
	}

	// Check for repeating characters
	if b.config.MaxRepeatingChars > 0 {
		if hasRepeatingChars(password, b.config.MaxRepeatingChars) {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Password cannot have more than %d repeating characters", b.config.MaxRepeatingChars))
		}
	}

	// Check against forbidden passwords
	lowerPassword := strings.ToLower(password)
	for _, forbidden := range b.config.ForbiddenPasswords {
		if lowerPassword == strings.ToLower(forbidden) {
			result.Valid = false
			result.Errors = append(result.Errors, "Password is too common and not allowed")
			break
		}
	}

	// Check for common patterns
	if hasCommonPatterns(password) {
		result.Warnings = append(result.Warnings, "Password contains common patterns")
	}

	// Calculate strength and score
	result.Strength, result.Score = b.calculatePasswordStrength(password)

	return result
}

// GeneratePassword generates a secure random password
func (b *BcryptService) GeneratePassword(length int) (string, error) {
	if length < b.config.MinLength {
		length = b.config.MinLength
	}
	if length > b.config.MaxLength {
		length = b.config.MaxLength
	}

	// Character sets
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers := "0123456789"
	special := b.config.SpecialChars

	var charset string
	var password strings.Builder

	// Ensure at least one character from each required set
	if b.config.RequireLowercase {
		char, err := randomChar(lowercase)
		if err != nil {
			return "", err
		}
		password.WriteRune(char)
		charset += lowercase
	}

	if b.config.RequireUppercase {
		char, err := randomChar(uppercase)
		if err != nil {
			return "", err
		}
		password.WriteRune(char)
		charset += uppercase
	}

	if b.config.RequireNumbers {
		char, err := randomChar(numbers)
		if err != nil {
			return "", err
		}
		password.WriteRune(char)
		charset += numbers
	}

	if b.config.RequireSpecialChars {
		char, err := randomChar(special)
		if err != nil {
			return "", err
		}
		password.WriteRune(char)
		charset += special
	}

	// If no requirements, use all character sets
	if charset == "" {
		charset = lowercase + uppercase + numbers + special
	}

	// Fill remaining length with random characters
	for password.Len() < length {
		char, err := randomChar(charset)
		if err != nil {
			return "", err
		}
		password.WriteRune(char)
	}

	// Shuffle the password
	passwordBytes := []byte(password.String())
	for i := len(passwordBytes) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		passwordBytes[i], passwordBytes[j.Int64()] = passwordBytes[j.Int64()], passwordBytes[i]
	}

	generatedPassword := string(passwordBytes)

	b.logger.WithField("length", length).Info("Password generated successfully")
	return generatedPassword, nil
}

// IsPasswordCompromised checks if a password has been compromised (placeholder implementation)
func (b *BcryptService) IsPasswordCompromised(password string) (bool, error) {
	// In a real implementation, this would check against a database of compromised passwords
	// such as HaveIBeenPwned API or a local database

	// For now, just check against forbidden passwords
	lowerPassword := strings.ToLower(password)
	for _, forbidden := range b.config.ForbiddenPasswords {
		if lowerPassword == strings.ToLower(forbidden) {
			return true, nil
		}
	}

	return false, nil
}

// GetPasswordStrength returns the strength of a password
func (b *BcryptService) GetPasswordStrength(password string) PasswordStrength {
	strength, _ := b.calculatePasswordStrength(password)
	return strength
}

// ValidatePassword validates a password and returns an error if invalid (interface method)
func (b *BcryptService) ValidatePassword(password string) error {
	result := b.validatePasswordInternal(password)
	if !result.Valid {
		return fmt.Errorf("password validation failed: %v", result.Errors)
	}
	return nil
}

// ValidatePasswordStrength validates password strength and returns an error if weak
func (b *BcryptService) ValidatePasswordStrength(password string) error {
	strength, score := b.calculatePasswordStrength(password)
	if strength == PasswordStrengthWeak || score < 40 {
		return fmt.Errorf("password is too weak (score: %d)", score)
	}
	return nil
}

// CheckPasswordPolicy checks if password meets policy requirements
func (b *BcryptService) CheckPasswordPolicy(password string) error {
	result := b.validatePasswordInternal(password)
	if !result.Valid {
		return fmt.Errorf("password policy violation: %v", result.Errors)
	}
	return nil
}

// GetPasswordPolicy returns the current password policy
func (b *BcryptService) GetPasswordPolicy() *application.PasswordPolicy {
	return &application.PasswordPolicy{
		MinLength:        b.config.MinLength,
		MaxLength:        b.config.MaxLength,
		RequireUppercase: b.config.RequireUppercase,
		RequireLowercase: b.config.RequireLowercase,
		RequireNumbers:   b.config.RequireNumbers,
		RequireSymbols:   b.config.RequireSpecialChars,
	}
}

// calculatePasswordStrength calculates password strength and score
func (b *BcryptService) calculatePasswordStrength(password string) (PasswordStrength, int) {
	score := 0

	// Length score
	if len(password) >= 8 {
		score += 10
	}
	if len(password) >= 12 {
		score += 10
	}
	if len(password) >= 16 {
		score += 10
	}

	// Character variety score
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[^A-Za-z0-9]`).MatchString(password)

	if hasUpper {
		score += 10
	}
	if hasLower {
		score += 10
	}
	if hasNumber {
		score += 10
	}
	if hasSpecial {
		score += 15
	}

	// Bonus for character variety
	variety := 0
	if hasUpper {
		variety++
	}
	if hasLower {
		variety++
	}
	if hasNumber {
		variety++
	}
	if hasSpecial {
		variety++
	}

	if variety >= 3 {
		score += 10
	}
	if variety == 4 {
		score += 10
	}

	// Penalty for common patterns
	if hasCommonPatterns(password) {
		score -= 20
	}

	// Penalty for repeating characters
	if hasRepeatingChars(password, 3) {
		score -= 15
	}

	// Determine strength based on score
	switch {
	case score >= 80:
		return PasswordStrengthVeryStrong, score
	case score >= 60:
		return PasswordStrengthStrong, score
	case score >= 40:
		return PasswordStrengthMedium, score
	default:
		return PasswordStrengthWeak, score
	}
}

// Helper functions

// randomChar returns a random character from the given charset
func randomChar(charset string) (rune, error) {
	if len(charset) == 0 {
		return 0, fmt.Errorf("charset cannot be empty")
	}

	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
	if err != nil {
		return 0, err
	}

	return rune(charset[index.Int64()]), nil
}

// hasRepeatingChars checks if password has more than maxRepeating consecutive identical characters
func hasRepeatingChars(password string, maxRepeating int) bool {
	if maxRepeating <= 0 {
		return false
	}

	count := 1
	for i := 1; i < len(password); i++ {
		if password[i] == password[i-1] {
			count++
			if count > maxRepeating {
				return true
			}
		} else {
			count = 1
		}
	}
	return false
}

// hasCommonPatterns checks for common password patterns
func hasCommonPatterns(password string) bool {
	lowerPassword := strings.ToLower(password)

	// Check for sequential characters
	sequences := []string{
		"abcdefghijklmnopqrstuvwxyz",
		"qwertyuiopasdfghjklzxcvbnm",
		"0123456789",
	}

	for _, seq := range sequences {
		for i := 0; i <= len(seq)-4; i++ {
			if strings.Contains(lowerPassword, seq[i:i+4]) {
				return true
			}
		}
	}

	// Check for common patterns
	commonPatterns := []string{
		"password", "123456", "qwerty", "admin", "login",
		"welcome", "monkey", "dragon", "master", "shadow",
	}

	for _, pattern := range commonPatterns {
		if strings.Contains(lowerPassword, pattern) {
			return true
		}
	}

	return false
}

// DefaultConfig returns default password service configuration
func DefaultConfig() *Config {
	return &Config{
		Cost:                12,
		MinLength:           8,
		MaxLength:           128,
		RequireUppercase:    true,
		RequireLowercase:    true,
		RequireNumbers:      true,
		RequireSpecialChars: true,
		SpecialChars:        "!@#$%^&*()_+-=[]{}|;:,.<>?",
		MaxRepeatingChars:   3,
		ForbiddenPasswords: []string{
			"password", "123456", "123456789", "qwerty", "abc123",
			"password123", "admin", "letmein", "welcome", "monkey",
		},
	}
}
