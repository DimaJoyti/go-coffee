package security

import (
	"fmt"
	"html"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// InputValidator provides comprehensive input validation and sanitization
type InputValidator struct {
	config *ValidationConfig
	logger Logger
}

// ValidationConfig contains validation configuration
type ValidationConfig struct {
	MaxStringLength     int      `json:"max_string_length"`
	MaxArrayLength      int      `json:"max_array_length"`
	AllowedFileTypes    []string `json:"allowed_file_types"`
	BlockedPatterns     []string `json:"blocked_patterns"`
	RequireHTTPS        bool     `json:"require_https"`
	AllowedDomains      []string `json:"allowed_domains"`
	MaxUploadSize       int64    `json:"max_upload_size"`
	EnableSQLInjection  bool     `json:"enable_sql_injection_detection"`
	EnableXSSProtection bool     `json:"enable_xss_protection"`
}

// ValidationResult represents the result of input validation
type ValidationResult struct {
	Valid      bool     `json:"valid"`
	Errors     []string `json:"errors"`
	Warnings   []string `json:"warnings"`
	Sanitized  string   `json:"sanitized"`
	Confidence float64  `json:"confidence"`
}

// ValidationRule represents a validation rule
type ValidationRule struct {
	Name        string                                    `json:"name"`
	Required    bool                                      `json:"required"`
	Validator   func(value interface{}) *ValidationResult `json:"-"`
	Sanitizer   func(value string) string                 `json:"-"`
	Description string                                    `json:"description"`
}

// NewInputValidator creates a new input validator
func NewInputValidator(config *ValidationConfig, logger Logger) *InputValidator {
	if config == nil {
		config = DefaultValidationConfig()
	}

	return &InputValidator{
		config: config,
		logger: logger,
	}
}

// DefaultValidationConfig returns default validation configuration
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		MaxStringLength:     10000,
		MaxArrayLength:      1000,
		AllowedFileTypes:    []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".txt", ".csv"},
		BlockedPatterns:     []string{"<script", "javascript:", "vbscript:", "onload=", "onerror="},
		RequireHTTPS:        true,
		AllowedDomains:      []string{},
		MaxUploadSize:       10 * 1024 * 1024, // 10MB
		EnableSQLInjection:  true,
		EnableXSSProtection: true,
	}
}

// ValidateString validates and sanitizes a string input
func (iv *InputValidator) ValidateString(input string, rules ...ValidationRule) *ValidationResult {
	result := &ValidationResult{
		Valid:      true,
		Errors:     make([]string, 0),
		Warnings:   make([]string, 0),
		Sanitized:  input,
		Confidence: 1.0,
	}

	// Check length
	if len(input) > iv.config.MaxStringLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("String length exceeds maximum of %d characters", iv.config.MaxStringLength))
		result.Confidence -= 0.3
	}

	// Check for blocked patterns
	for _, pattern := range iv.config.BlockedPatterns {
		if strings.Contains(strings.ToLower(input), strings.ToLower(pattern)) {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Input contains blocked pattern: %s", pattern))
			result.Confidence -= 0.5
		}
	}

	// SQL injection detection
	if iv.config.EnableSQLInjection && iv.detectSQLInjection(input) {
		result.Valid = false
		result.Errors = append(result.Errors, "Potential SQL injection detected")
		result.Confidence -= 0.7
	}

	// XSS detection
	if iv.config.EnableXSSProtection && iv.detectXSS(input) {
		result.Valid = false
		result.Errors = append(result.Errors, "Potential XSS attack detected")
		result.Confidence -= 0.7
	}

	// Apply custom rules
	for _, rule := range rules {
		if rule.Validator != nil {
			ruleResult := rule.Validator(input)
			if !ruleResult.Valid {
				result.Valid = false
				result.Errors = append(result.Errors, ruleResult.Errors...)
				result.Warnings = append(result.Warnings, ruleResult.Warnings...)
				result.Confidence *= ruleResult.Confidence
			}
		}
	}

	// Sanitize input
	result.Sanitized = iv.sanitizeString(input)

	if !result.Valid && iv.logger != nil {
		iv.logger.Warn("Input validation failed", "input", input, "errors", result.Errors)
	}

	return result
}

// ValidateEmail validates an email address
func (iv *InputValidator) ValidateEmail(email string) *ValidationResult {
	result := &ValidationResult{
		Valid:      true,
		Errors:     make([]string, 0),
		Sanitized:  strings.TrimSpace(strings.ToLower(email)),
		Confidence: 1.0,
	}

	// Basic email validation
	if _, err := mail.ParseAddress(email); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, "Invalid email format")
		result.Confidence = 0.0
		return result
	}

	// Check domain restrictions
	if len(iv.config.AllowedDomains) > 0 {
		parts := strings.Split(email, "@")
		if len(parts) == 2 {
			domain := parts[1]
			allowed := false
			for _, allowedDomain := range iv.config.AllowedDomains {
				if domain == allowedDomain {
					allowed = true
					break
				}
			}
			if !allowed {
				result.Valid = false
				result.Errors = append(result.Errors, "Email domain not allowed")
				result.Confidence = 0.2
			}
		}
	}

	return result
}

// ValidateURL validates a URL
func (iv *InputValidator) ValidateURL(urlStr string) *ValidationResult {
	result := &ValidationResult{
		Valid:      true,
		Errors:     make([]string, 0),
		Sanitized:  strings.TrimSpace(urlStr),
		Confidence: 1.0,
	}

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, "Invalid URL format")
		result.Confidence = 0.0
		return result
	}

	// Check scheme
	if parsedURL.Scheme == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "URL scheme is required")
		result.Confidence = 0.3
	} else if iv.config.RequireHTTPS && parsedURL.Scheme != "https" {
		result.Valid = false
		result.Errors = append(result.Errors, "HTTPS is required")
		result.Confidence = 0.5
	}

	// Check for suspicious patterns
	if strings.Contains(strings.ToLower(urlStr), "javascript:") ||
		strings.Contains(strings.ToLower(urlStr), "data:") ||
		strings.Contains(strings.ToLower(urlStr), "vbscript:") {
		result.Valid = false
		result.Errors = append(result.Errors, "Suspicious URL scheme detected")
		result.Confidence = 0.0
	}

	return result
}

// ValidateJSON validates JSON structure and content
func (iv *InputValidator) ValidateJSON(jsonStr string) *ValidationResult {
	result := &ValidationResult{
		Valid:      true,
		Errors:     make([]string, 0),
		Sanitized:  jsonStr,
		Confidence: 1.0,
	}

	// Check length
	if len(jsonStr) > iv.config.MaxStringLength {
		result.Valid = false
		result.Errors = append(result.Errors, "JSON string too long")
		result.Confidence = 0.3
	}

	// Check for malicious patterns
	if iv.config.EnableXSSProtection && iv.detectXSS(jsonStr) {
		result.Valid = false
		result.Errors = append(result.Errors, "Potential XSS in JSON detected")
		result.Confidence = 0.0
	}

	return result
}

// ValidateInteger validates an integer value
func (iv *InputValidator) ValidateInteger(value string, min, max int64) *ValidationResult {
	result := &ValidationResult{
		Valid:      true,
		Errors:     make([]string, 0),
		Sanitized:  strings.TrimSpace(value),
		Confidence: 1.0,
	}

	// Parse integer
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, "Invalid integer format")
		result.Confidence = 0.0
		return result
	}

	// Check range
	if intValue < min {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Value %d is below minimum %d", intValue, min))
		result.Confidence = 0.5
	}

	if intValue > max {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Value %d is above maximum %d", intValue, max))
		result.Confidence = 0.5
	}

	return result
}

// ValidateFloat validates a float value
func (iv *InputValidator) ValidateFloat(value string, min, max float64) *ValidationResult {
	result := &ValidationResult{
		Valid:      true,
		Errors:     make([]string, 0),
		Sanitized:  strings.TrimSpace(value),
		Confidence: 1.0,
	}

	// Parse float
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, "Invalid float format")
		result.Confidence = 0.0
		return result
	}

	// Check range
	if floatValue < min {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Value %f is below minimum %f", floatValue, min))
		result.Confidence = 0.5
	}

	if floatValue > max {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Value %f is above maximum %f", floatValue, max))
		result.Confidence = 0.5
	}

	return result
}

// ValidateDate validates a date string
func (iv *InputValidator) ValidateDate(dateStr, format string) *ValidationResult {
	result := &ValidationResult{
		Valid:      true,
		Errors:     make([]string, 0),
		Sanitized:  strings.TrimSpace(dateStr),
		Confidence: 1.0,
	}

	// Parse date
	_, err := time.Parse(format, dateStr)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid date format, expected: %s", format))
		result.Confidence = 0.0
	}

	return result
}

// ValidateFilename validates a filename
func (iv *InputValidator) ValidateFilename(filename string) *ValidationResult {
	result := &ValidationResult{
		Valid:      true,
		Errors:     make([]string, 0),
		Sanitized:  iv.sanitizeFilename(filename),
		Confidence: 1.0,
	}

	// Check for dangerous characters
	dangerousChars := []string{"..", "/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range dangerousChars {
		if strings.Contains(filename, char) {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Filename contains dangerous character: %s", char))
			result.Confidence = 0.0
		}
	}

	// Check file extension
	if len(iv.config.AllowedFileTypes) > 0 {
		allowed := false
		for _, ext := range iv.config.AllowedFileTypes {
			if strings.HasSuffix(strings.ToLower(filename), ext) {
				allowed = true
				break
			}
		}
		if !allowed {
			result.Valid = false
			result.Errors = append(result.Errors, "File type not allowed")
			result.Confidence = 0.2
		}
	}

	return result
}

// detectSQLInjection detects potential SQL injection attempts
func (iv *InputValidator) detectSQLInjection(input string) bool {
	sqlPatterns := []string{
		`(?i)(union\s+select)`,
		`(?i)(select\s+.*\s+from)`,
		`(?i)(insert\s+into)`,
		`(?i)(delete\s+from)`,
		`(?i)(update\s+.*\s+set)`,
		`(?i)(drop\s+table)`,
		`(?i)(create\s+table)`,
		`(?i)(alter\s+table)`,
		`(?i)(\'\s*or\s*\'\s*=\s*\')`,
		`(?i)(\'\s*or\s*1\s*=\s*1)`,
		`(?i)(--\s*$)`,
		`(?i)(/\*.*\*/)`,
		`(?i)(;\s*drop)`,
		`(?i)(;\s*delete)`,
		`(?i)(;\s*update)`,
		`(?i)(;\s*insert)`,
	}

	for _, pattern := range sqlPatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return true
		}
	}

	return false
}

// detectXSS detects potential XSS attempts
func (iv *InputValidator) detectXSS(input string) bool {
	xssPatterns := []string{
		`(?i)<script[^>]*>`,
		`(?i)</script>`,
		`(?i)javascript:`,
		`(?i)vbscript:`,
		`(?i)onload\s*=`,
		`(?i)onerror\s*=`,
		`(?i)onclick\s*=`,
		`(?i)onmouseover\s*=`,
		`(?i)onfocus\s*=`,
		`(?i)onblur\s*=`,
		`(?i)<iframe[^>]*>`,
		`(?i)<object[^>]*>`,
		`(?i)<embed[^>]*>`,
		`(?i)<link[^>]*>`,
		`(?i)<meta[^>]*>`,
		`(?i)expression\s*\(`,
		`(?i)@import`,
		`(?i)document\.cookie`,
		`(?i)document\.write`,
		`(?i)eval\s*\(`,
	}

	for _, pattern := range xssPatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return true
		}
	}

	return false
}

// sanitizeString sanitizes a string by removing/escaping dangerous content
func (iv *InputValidator) sanitizeString(input string) string {
	// HTML escape
	sanitized := html.EscapeString(input)

	// Remove null bytes
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")

	// Remove control characters except tab, newline, and carriage return
	sanitized = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\t' && r != '\n' && r != '\r' {
			return -1
		}
		return r
	}, sanitized)

	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

// sanitizeFilename sanitizes a filename
func (iv *InputValidator) sanitizeFilename(filename string) string {
	// Remove path separators and dangerous characters
	sanitized := strings.ReplaceAll(filename, "..", "")
	sanitized = strings.ReplaceAll(sanitized, "/", "")
	sanitized = strings.ReplaceAll(sanitized, "\\", "")
	sanitized = strings.ReplaceAll(sanitized, ":", "")
	sanitized = strings.ReplaceAll(sanitized, "*", "")
	sanitized = strings.ReplaceAll(sanitized, "?", "")
	sanitized = strings.ReplaceAll(sanitized, "\"", "")
	sanitized = strings.ReplaceAll(sanitized, "<", "")
	sanitized = strings.ReplaceAll(sanitized, ">", "")
	sanitized = strings.ReplaceAll(sanitized, "|", "")

	// Remove control characters
	sanitized = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, sanitized)

	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	// Ensure filename is not empty
	if sanitized == "" {
		sanitized = "sanitized_file"
	}

	return sanitized
}

// CreateValidationRule creates a custom validation rule
func CreateValidationRule(name string, required bool, validator func(interface{}) *ValidationResult) ValidationRule {
	return ValidationRule{
		Name:      name,
		Required:  required,
		Validator: validator,
	}
}

// Common validation rules

// RequiredStringRule validates that a string is not empty
func RequiredStringRule() ValidationRule {
	return CreateValidationRule("required_string", true, func(value interface{}) *ValidationResult {
		str, ok := value.(string)
		if !ok {
			return &ValidationResult{
				Valid:      false,
				Errors:     []string{"Value must be a string"},
				Confidence: 0.0,
			}
		}

		if strings.TrimSpace(str) == "" {
			return &ValidationResult{
				Valid:      false,
				Errors:     []string{"String cannot be empty"},
				Confidence: 0.0,
			}
		}

		return &ValidationResult{Valid: true, Confidence: 1.0}
	})
}

// AlphanumericRule validates that a string contains only alphanumeric characters
func AlphanumericRule() ValidationRule {
	return CreateValidationRule("alphanumeric", false, func(value interface{}) *ValidationResult {
		str, ok := value.(string)
		if !ok {
			return &ValidationResult{
				Valid:      false,
				Errors:     []string{"Value must be a string"},
				Confidence: 0.0,
			}
		}

		for _, r := range str {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
				return &ValidationResult{
					Valid:      false,
					Errors:     []string{"String must contain only alphanumeric characters"},
					Confidence: 0.5,
				}
			}
		}

		return &ValidationResult{Valid: true, Confidence: 1.0}
	})
}

// LengthRule validates string length
func LengthRule(min, max int) ValidationRule {
	return CreateValidationRule("length", false, func(value interface{}) *ValidationResult {
		str, ok := value.(string)
		if !ok {
			return &ValidationResult{
				Valid:      false,
				Errors:     []string{"Value must be a string"},
				Confidence: 0.0,
			}
		}

		length := len(str)
		if length < min {
			return &ValidationResult{
				Valid:      false,
				Errors:     []string{fmt.Sprintf("String length %d is below minimum %d", length, min)},
				Confidence: 0.5,
			}
		}

		if length > max {
			return &ValidationResult{
				Valid:      false,
				Errors:     []string{fmt.Sprintf("String length %d exceeds maximum %d", length, max)},
				Confidence: 0.5,
			}
		}

		return &ValidationResult{Valid: true, Confidence: 1.0}
	})
}
