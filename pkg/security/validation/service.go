package validation

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// ValidationService provides input validation and sanitization
type ValidationService struct {
	config *Config
	logger *logger.Logger
}

// Config represents validation configuration
type Config struct {
	MaxInputLength    int      `yaml:"max_input_length" env:"MAX_INPUT_LENGTH" default:"10000"`
	AllowedFileTypes  []string `yaml:"allowed_file_types" env:"ALLOWED_FILE_TYPES"`
	BlockedPatterns   []string `yaml:"blocked_patterns" env:"BLOCKED_PATTERNS"`
	AllowedDomains    []string `yaml:"allowed_domains" env:"ALLOWED_DOMAINS"`
	StrictMode        bool     `yaml:"strict_mode" env:"STRICT_MODE" default:"true"`
	EnableSanitization bool    `yaml:"enable_sanitization" env:"ENABLE_SANITIZATION" default:"true"`
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	IsValid      bool     `json:"is_valid"`
	Errors       []string `json:"errors,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
	SanitizedValue string `json:"sanitized_value,omitempty"`
	ThreatLevel  string   `json:"threat_level,omitempty"`
}

// ThreatLevel represents the severity of detected threats
type ThreatLevel string

const (
	ThreatLevelNone     ThreatLevel = "none"
	ThreatLevelLow      ThreatLevel = "low"
	ThreatLevelMedium   ThreatLevel = "medium"
	ThreatLevelHigh     ThreatLevel = "high"
	ThreatLevelCritical ThreatLevel = "critical"
)

// Common regex patterns for validation
var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex    = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	uuidRegex     = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	alphanumRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	
	// Security patterns
	sqlInjectionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)`),
		regexp.MustCompile(`(?i)(script|javascript|vbscript|onload|onerror|onclick)`),
		regexp.MustCompile(`(?i)(\-\-|\#|\/\*|\*\/)`),
		regexp.MustCompile(`(?i)(char|nchar|varchar|nvarchar|alter|begin|cast|create|cursor|declare|delete|drop|end|exec|execute|fetch|insert|kill|open|select|sys|sysobjects|syscolumns|table|update)`),
	}
	
	xssPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
		regexp.MustCompile(`(?i)<iframe[^>]*>.*?</iframe>`),
		regexp.MustCompile(`(?i)<object[^>]*>.*?</object>`),
		regexp.MustCompile(`(?i)<embed[^>]*>.*?</embed>`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)vbscript:`),
		regexp.MustCompile(`(?i)on\w+\s*=`),
	}
	
	pathTraversalPatterns = []*regexp.Regexp{
		regexp.MustCompile(`\.\.\/`),
		regexp.MustCompile(`\.\.\\`),
		regexp.MustCompile(`%2e%2e%2f`),
		regexp.MustCompile(`%2e%2e%5c`),
	}
)

// NewValidationService creates a new validation service
func NewValidationService(config *Config, logger *logger.Logger) *ValidationService {
	return &ValidationService{
		config: config,
		logger: logger,
	}
}

// ValidateEmail validates email format
func (v *ValidationService) ValidateEmail(email string) *ValidationResult {
	result := &ValidationResult{
		IsValid:     true,
		ThreatLevel: string(ThreatLevelNone),
	}

	if email == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "email is required")
		return result
	}

	if len(email) > 254 {
		result.IsValid = false
		result.Errors = append(result.Errors, "email is too long")
		return result
	}

	if !emailRegex.MatchString(email) {
		result.IsValid = false
		result.Errors = append(result.Errors, "invalid email format")
		return result
	}

	// Check for suspicious patterns
	if v.containsSuspiciousPatterns(email) {
		result.ThreatLevel = string(ThreatLevelMedium)
		result.Warnings = append(result.Warnings, "email contains suspicious patterns")
	}

	if v.config.EnableSanitization {
		result.SanitizedValue = v.sanitizeInput(email)
	}

	return result
}

// ValidatePassword validates password strength
func (v *ValidationService) ValidatePassword(password string) *ValidationResult {
	result := &ValidationResult{
		IsValid:     true,
		ThreatLevel: string(ThreatLevelNone),
	}

	if password == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "password is required")
		return result
	}

	if len(password) < 8 {
		result.IsValid = false
		result.Errors = append(result.Errors, "password must be at least 8 characters long")
	}

	if len(password) > 128 {
		result.IsValid = false
		result.Errors = append(result.Errors, "password is too long")
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

	if !hasUpper {
		result.IsValid = false
		result.Errors = append(result.Errors, "password must contain at least one uppercase letter")
	}

	if !hasLower {
		result.IsValid = false
		result.Errors = append(result.Errors, "password must contain at least one lowercase letter")
	}

	if !hasDigit {
		result.IsValid = false
		result.Errors = append(result.Errors, "password must contain at least one digit")
	}

	if !hasSpecial {
		result.IsValid = false
		result.Errors = append(result.Errors, "password must contain at least one special character")
	}

	// Check for common weak passwords
	weakPasswords := []string{"password", "123456", "qwerty", "admin", "letmein"}
	for _, weak := range weakPasswords {
		if strings.ToLower(password) == weak {
			result.IsValid = false
			result.Errors = append(result.Errors, "password is too common")
			result.ThreatLevel = string(ThreatLevelHigh)
			break
		}
	}

	return result
}

// ValidateInput validates general input for security threats
func (v *ValidationService) ValidateInput(input string) *ValidationResult {
	result := &ValidationResult{
		IsValid:     true,
		ThreatLevel: string(ThreatLevelNone),
	}

	if len(input) > v.config.MaxInputLength {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("input exceeds maximum length of %d", v.config.MaxInputLength))
		return result
	}

	// Check for SQL injection
	if v.detectSQLInjection(input) {
		result.IsValid = false
		result.Errors = append(result.Errors, "potential SQL injection detected")
		result.ThreatLevel = string(ThreatLevelCritical)
	}

	// Check for XSS
	if v.detectXSS(input) {
		result.IsValid = false
		result.Errors = append(result.Errors, "potential XSS attack detected")
		result.ThreatLevel = string(ThreatLevelHigh)
	}

	// Check for path traversal
	if v.detectPathTraversal(input) {
		result.IsValid = false
		result.Errors = append(result.Errors, "potential path traversal attack detected")
		result.ThreatLevel = string(ThreatLevelHigh)
	}

	// Check blocked patterns
	for _, pattern := range v.config.BlockedPatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			result.IsValid = false
			result.Errors = append(result.Errors, "input contains blocked pattern")
			result.ThreatLevel = string(ThreatLevelMedium)
			break
		}
	}

	if v.config.EnableSanitization && result.IsValid {
		result.SanitizedValue = v.sanitizeInput(input)
	}

	return result
}

// ValidateURL validates URL format and domain
func (v *ValidationService) ValidateURL(urlStr string) *ValidationResult {
	result := &ValidationResult{
		IsValid:     true,
		ThreatLevel: string(ThreatLevelNone),
	}

	if urlStr == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "URL is required")
		return result
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, "invalid URL format")
		return result
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		result.IsValid = false
		result.Errors = append(result.Errors, "only HTTP and HTTPS schemes are allowed")
		return result
	}

	// Check allowed domains
	if len(v.config.AllowedDomains) > 0 {
		allowed := false
		for _, domain := range v.config.AllowedDomains {
			if strings.HasSuffix(parsedURL.Host, domain) {
				allowed = true
				break
			}
		}
		if !allowed {
			result.IsValid = false
			result.Errors = append(result.Errors, "domain not allowed")
			result.ThreatLevel = string(ThreatLevelMedium)
		}
	}

	return result
}

// ValidateIP validates IP address
func (v *ValidationService) ValidateIP(ip string) *ValidationResult {
	result := &ValidationResult{
		IsValid:     true,
		ThreatLevel: string(ThreatLevelNone),
	}

	if ip == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "IP address is required")
		return result
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		result.IsValid = false
		result.Errors = append(result.Errors, "invalid IP address format")
		return result
	}

	// Check for private/local IPs in strict mode
	if v.config.StrictMode {
		if parsedIP.IsLoopback() || parsedIP.IsPrivate() {
			result.Warnings = append(result.Warnings, "private or loopback IP address")
			result.ThreatLevel = string(ThreatLevelLow)
		}
	}

	return result
}

// Helper methods

func (v *ValidationService) detectSQLInjection(input string) bool {
	for _, pattern := range sqlInjectionPatterns {
		if pattern.MatchString(input) {
			v.logger.Warn("SQL injection pattern detected", map[string]any{
				"input":   input,
				"pattern": pattern.String(),
			})
			return true
		}
	}
	return false
}

func (v *ValidationService) detectXSS(input string) bool {
	for _, pattern := range xssPatterns {
		if pattern.MatchString(input) {
			v.logger.Warn("XSS pattern detected", map[string]any{
				"input":   input,
				"pattern": pattern.String(),
			})
			return true
		}
	}
	return false
}

func (v *ValidationService) detectPathTraversal(input string) bool {
	for _, pattern := range pathTraversalPatterns {
		if pattern.MatchString(input) {
			v.logger.Warn("Path traversal pattern detected", map[string]any{
				"input":   input,
				"pattern": pattern.String(),
			})
			return true
		}
	}
	return false
}

func (v *ValidationService) containsSuspiciousPatterns(input string) bool {
	// Check for suspicious patterns that might indicate malicious intent
	suspiciousPatterns := []string{
		"admin", "root", "test", "guest", "anonymous",
		"null", "undefined", "eval", "exec",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	return false
}

func (v *ValidationService) sanitizeInput(input string) string {
	// Basic sanitization - remove/escape dangerous characters
	sanitized := strings.ReplaceAll(input, "<", "&lt;")
	sanitized = strings.ReplaceAll(sanitized, ">", "&gt;")
	sanitized = strings.ReplaceAll(sanitized, "\"", "&quot;")
	sanitized = strings.ReplaceAll(sanitized, "'", "&#x27;")
	sanitized = strings.ReplaceAll(sanitized, "&", "&amp;")
	
	return sanitized
}
