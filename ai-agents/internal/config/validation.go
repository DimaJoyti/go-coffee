package config

import (
	"fmt"
	"net"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// Validator is the global configuration validator
var Validator *validator.Validate

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s (value: %s)", ve.Field, ve.Message, ve.Value)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (ves ValidationErrors) Error() string {
	if len(ves) == 0 {
		return "no validation errors"
	}
	
	var messages []string
	for _, ve := range ves {
		messages = append(messages, ve.Error())
	}
	
	return fmt.Sprintf("validation failed: %s", strings.Join(messages, "; "))
}

// init initializes the validator with custom validation rules
func init() {
	Validator = validator.New()
	
	// Register custom validation functions
	registerCustomValidators()
	
	// Use JSON tag names for field names in validation errors
	Validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// registerCustomValidators registers custom validation functions
func registerCustomValidators() {
	// URL validation
	Validator.RegisterValidation("url", validateURL)
	
	// Host validation (hostname or IP)
	Validator.RegisterValidation("host", validateHost)
	
	// Duration validation
	Validator.RegisterValidation("duration", validateDuration)
	
	// Environment validation
	Validator.RegisterValidation("environment", validateEnvironment)
	
	// Database driver validation
	Validator.RegisterValidation("db_driver", validateDatabaseDriver)
	
	// AI provider validation
	Validator.RegisterValidation("ai_provider", validateAIProvider)
	
	// File path validation
	Validator.RegisterValidation("filepath", validateFilePath)
	
	// Port validation
	Validator.RegisterValidation("port", validatePort)
	
	// Non-empty slice validation
	Validator.RegisterValidation("nonempty", validateNonEmptySlice)
	
	// Kafka topic name validation
	Validator.RegisterValidation("kafka_topic", validateKafkaTopicName)
	
	// JWT algorithm validation
	Validator.RegisterValidation("jwt_algorithm", validateJWTAlgorithm)
	
	// Email validation (enhanced)
	Validator.RegisterValidation("email_enhanced", validateEmailEnhanced)
	
	// Secret validation (non-empty in production)
	Validator.RegisterValidation("secret", validateSecret)
}

// validateURL validates URL format
func validateURL(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty URLs if not required
	}
	
	_, err := url.Parse(value)
	return err == nil
}

// validateHost validates hostname or IP address
func validateHost(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return false
	}
	
	// Check if it's a valid IP address
	if net.ParseIP(value) != nil {
		return true
	}
	
	// Check if it's a valid hostname
	if len(value) > 253 {
		return false
	}
	
	// Hostname validation regex
	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	return hostnameRegex.MatchString(value)
}

// validateDuration validates duration format
func validateDuration(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty durations if not required
	}
	
	_, err := time.ParseDuration(value)
	return err == nil
}

// validateEnvironment validates environment values
func validateEnvironment(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	validEnvs := []string{"development", "staging", "production", "testing"}
	
	for _, env := range validEnvs {
		if value == env {
			return true
		}
	}
	
	return false
}

// validateDatabaseDriver validates database driver names
func validateDatabaseDriver(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	validDrivers := []string{"postgres", "mysql", "sqlite", "sqlserver"}
	
	for _, driver := range validDrivers {
		if value == driver {
			return true
		}
	}
	
	return false
}

// validateAIProvider validates AI provider names
func validateAIProvider(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	validProviders := []string{"openai", "gemini", "claude", "ollama", "huggingface"}
	
	for _, provider := range validProviders {
		if value == provider {
			return true
		}
	}
	
	return false
}

// validateFilePath validates file path format
func validateFilePath(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty paths if not required
	}
	
	// Basic path validation - no null bytes, reasonable length
	if strings.Contains(value, "\x00") {
		return false
	}
	
	if len(value) > 4096 {
		return false
	}
	
	return true
}

// validatePort validates port numbers
func validatePort(fl validator.FieldLevel) bool {
	port := int(fl.Field().Int())
	return port > 0 && port <= 65535
}

// validateNonEmptySlice validates that slice is not empty
func validateNonEmptySlice(fl validator.FieldLevel) bool {
	return fl.Field().Len() > 0
}

// validateKafkaTopicName validates Kafka topic name format
func validateKafkaTopicName(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return false
	}
	
	// Kafka topic name rules
	if len(value) > 249 {
		return false
	}
	
	// Must contain only alphanumeric, dots, underscores, and hyphens
	topicRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !topicRegex.MatchString(value) {
		return false
	}
	
	// Cannot be "." or ".."
	if value == "." || value == ".." {
		return false
	}
	
	return true
}

// validateJWTAlgorithm validates JWT algorithm names
func validateJWTAlgorithm(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	validAlgorithms := []string{"HS256", "HS384", "HS512", "RS256", "RS384", "RS512", "ES256", "ES384", "ES512"}
	
	for _, alg := range validAlgorithms {
		if value == alg {
			return true
		}
	}
	
	return false
}

// validateEmailEnhanced validates email format with enhanced rules
func validateEmailEnhanced(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty emails if not required
	}
	
	// Enhanced email regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(value)
}

// validateSecret validates secrets (must be non-empty in production)
func validateSecret(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	
	// Get the parent struct to check environment
	parent := fl.Parent()
	if parent.Kind() == reflect.Ptr {
		parent = parent.Elem()
	}
	
	// Try to find environment field
	envField := parent.FieldByName("Environment")
	if !envField.IsValid() {
		// If no environment field, require non-empty secret
		return value != ""
	}
	
	env := envField.String()
	if env == "production" {
		return value != ""
	}
	
	// In non-production environments, allow empty secrets
	return true
}

// ValidateConfig validates the entire configuration
func ValidateConfig(config *Config) error {
	err := Validator.Struct(config)
	if err == nil {
		return nil
	}
	
	var validationErrors ValidationErrors
	
	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validatorErrors {
			validationError := ValidationError{
				Field: fieldError.Field(),
				Tag:   fieldError.Tag(),
				Value: fmt.Sprintf("%v", fieldError.Value()),
			}
			
			// Generate human-readable error messages
			validationError.Message = generateErrorMessage(fieldError)
			validationErrors = append(validationErrors, validationError)
		}
	}
	
	// Additional custom validations
	if customErrors := validateCustomRules(config); len(customErrors) > 0 {
		validationErrors = append(validationErrors, customErrors...)
	}
	
	if len(validationErrors) > 0 {
		return validationErrors
	}
	
	return err
}

// generateErrorMessage generates human-readable error messages
func generateErrorMessage(fieldError validator.FieldError) string {
	field := fieldError.Field()
	tag := fieldError.Tag()
	param := fieldError.Param()
	value := fmt.Sprintf("%v", fieldError.Value())
	
	switch tag {
	case "required":
		return fmt.Sprintf("field '%s' is required", field)
	case "min":
		return fmt.Sprintf("field '%s' must be at least %s", field, param)
	case "max":
		return fmt.Sprintf("field '%s' must be at most %s", field, param)
	case "email":
		return fmt.Sprintf("field '%s' must be a valid email address", field)
	case "url":
		return fmt.Sprintf("field '%s' must be a valid URL", field)
	case "host":
		return fmt.Sprintf("field '%s' must be a valid hostname or IP address", field)
	case "port":
		return fmt.Sprintf("field '%s' must be a valid port number (1-65535)", field)
	case "duration":
		return fmt.Sprintf("field '%s' must be a valid duration (e.g., '30s', '5m')", field)
	case "environment":
		return fmt.Sprintf("field '%s' must be one of: development, staging, production, testing", field)
	case "db_driver":
		return fmt.Sprintf("field '%s' must be one of: postgres, mysql, sqlite, sqlserver", field)
	case "ai_provider":
		return fmt.Sprintf("field '%s' must be one of: openai, gemini, claude, ollama, huggingface", field)
	case "jwt_algorithm":
		return fmt.Sprintf("field '%s' must be a valid JWT algorithm", field)
	case "kafka_topic":
		return fmt.Sprintf("field '%s' must be a valid Kafka topic name", field)
	case "nonempty":
		return fmt.Sprintf("field '%s' cannot be empty", field)
	case "secret":
		return fmt.Sprintf("field '%s' is required in production environment", field)
	case "oneof":
		return fmt.Sprintf("field '%s' must be one of: %s", field, param)
	case "gte":
		return fmt.Sprintf("field '%s' must be greater than or equal to %s", field, param)
	case "lte":
		return fmt.Sprintf("field '%s' must be less than or equal to %s", field, param)
	case "len":
		return fmt.Sprintf("field '%s' must be exactly %s characters long", field, param)
	default:
		return fmt.Sprintf("field '%s' failed validation '%s' with value '%s'", field, tag, value)
	}
}

// validateCustomRules performs additional custom validation rules
func validateCustomRules(config *Config) ValidationErrors {
	var errors ValidationErrors
	
	// Validate AI provider configuration
	if config.AI.DefaultProvider != "" {
		if _, exists := config.AI.Providers[config.AI.DefaultProvider]; !exists {
			errors = append(errors, ValidationError{
				Field:   "ai.default_provider",
				Tag:     "custom",
				Value:   config.AI.DefaultProvider,
				Message: fmt.Sprintf("default AI provider '%s' not found in providers configuration", config.AI.DefaultProvider),
			})
		}
	}
	
	// Validate database configuration consistency
	if config.Database.Driver == "sqlite" && config.Database.Host != "" {
		errors = append(errors, ValidationError{
			Field:   "database.host",
			Tag:     "custom",
			Value:   config.Database.Host,
			Message: "host should not be specified for SQLite driver",
		})
	}
	
	// Validate Kafka topic names
	topics := []struct {
		name  string
		value string
	}{
		{"kafka.topics.beverage_created", config.Kafka.Topics.BeverageCreated},
		{"kafka.topics.beverage_updated", config.Kafka.Topics.BeverageUpdated},
		{"kafka.topics.task_created", config.Kafka.Topics.TaskCreated},
		{"kafka.topics.task_updated", config.Kafka.Topics.TaskUpdated},
		{"kafka.topics.notification_sent", config.Kafka.Topics.NotificationSent},
		{"kafka.topics.ai_request_completed", config.Kafka.Topics.AIRequestCompleted},
		{"kafka.topics.system_event", config.Kafka.Topics.SystemEvent},
	}
	
	for _, topic := range topics {
		if topic.value != "" && !validateKafkaTopicNameString(topic.value) {
			errors = append(errors, ValidationError{
				Field:   topic.name,
				Tag:     "kafka_topic",
				Value:   topic.value,
				Message: "invalid Kafka topic name format",
			})
		}
	}
	
	// Validate security configuration in production
	if config.Environment == Production {
		if config.Security.JWT.SecretKey == "" {
			errors = append(errors, ValidationError{
				Field:   "security.jwt.secret_key",
				Tag:     "required_in_production",
				Value:   "",
				Message: "JWT secret key is required in production environment",
			})
		}
		
		if config.Security.Encryption.SecretKey == "" {
			errors = append(errors, ValidationError{
				Field:   "security.encryption.secret_key",
				Tag:     "required_in_production",
				Value:   "",
				Message: "encryption secret key is required in production environment",
			})
		}
	}
	
	// Validate timeout configurations
	if config.Server.ReadTimeout > 0 && config.Server.WriteTimeout > 0 {
		if config.Server.WriteTimeout <= config.Server.ReadTimeout {
			errors = append(errors, ValidationError{
				Field:   "server.write_timeout",
				Tag:     "custom",
				Value:   config.Server.WriteTimeout.String(),
				Message: "write timeout should be greater than read timeout",
			})
		}
	}
	
	return errors
}

// validateKafkaTopicNameString validates a Kafka topic name string
func validateKafkaTopicNameString(topic string) bool {
	if topic == "" || len(topic) > 249 {
		return false
	}
	
	topicRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !topicRegex.MatchString(topic) {
		return false
	}
	
	return topic != "." && topic != ".."
}

// ValidatePartialConfig validates a partial configuration (useful for updates)
func ValidatePartialConfig(config interface{}) error {
	return Validator.Struct(config)
}

// ValidateField validates a single field value
func ValidateField(value interface{}, tag string) error {
	return Validator.Var(value, tag)
}
