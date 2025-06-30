package config

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// ConfigValidator provides comprehensive configuration validation
type ConfigValidator struct {
	errors ValidationErrors
}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		errors: make(ValidationErrors, 0),
	}
}

// Validate validates the entire configuration
func (v *ConfigValidator) Validate(config *Config) error {
	v.errors = make(ValidationErrors, 0)

	// Validate each section
	v.validateService(config.Service)
	v.validateServer(config.Server)
	v.validateDatabase(config.Database)
	v.validateKafka(config.Kafka)
	v.validateAI(config.AI)
	v.validateExternal(config.External)
	v.validateSecurity(config.Security)
	v.validateFeatures(config.Features)
	v.validateEnvironmentSpecific(config)

	if len(v.errors) > 0 {
		return v.errors
	}

	return nil
}

// validateService validates service configuration
func (v *ConfigValidator) validateService(service ServiceConfig) {
	if service.Name == "" {
		v.addError("service.name", service.Name, "service name is required")
	} else if !isValidServiceName(service.Name) {
		v.addError("service.name", service.Name, "service name must contain only lowercase letters, numbers, and hyphens")
	}

	if service.Version == "" {
		v.addError("service.version", service.Version, "service version is required")
	} else if !isValidVersion(service.Version) {
		v.addError("service.version", service.Version, "service version must follow semantic versioning (e.g., 1.0.0)")
	}

	if service.Port <= 0 || service.Port > 65535 {
		v.addError("service.port", service.Port, "port must be between 1 and 65535")
	}

	if service.Host == "" {
		v.addError("service.host", service.Host, "host is required")
	} else if !isValidHost(service.Host) {
		v.addError("service.host", service.Host, "invalid host format")
	}

	if service.BasePath != "" && !strings.HasPrefix(service.BasePath, "/") {
		v.addError("service.base_path", service.BasePath, "base path must start with '/'")
	}
}

// validateServer validates server configuration
func (v *ConfigValidator) validateServer(server ServerConfig) {
	if server.ReadTimeout <= 0 {
		v.addError("server.read_timeout", server.ReadTimeout, "read timeout must be positive")
	}

	if server.WriteTimeout <= 0 {
		v.addError("server.write_timeout", server.WriteTimeout, "write timeout must be positive")
	}

	if server.IdleTimeout <= 0 {
		v.addError("server.idle_timeout", server.IdleTimeout, "idle timeout must be positive")
	}

	if server.ShutdownTimeout <= 0 {
		v.addError("server.shutdown_timeout", server.ShutdownTimeout, "shutdown timeout must be positive")
	}

	if server.MaxHeaderBytes <= 0 {
		v.addError("server.max_header_bytes", server.MaxHeaderBytes, "max header bytes must be positive")
	}

	if server.MetricsPath != "" && !strings.HasPrefix(server.MetricsPath, "/") {
		v.addError("server.metrics_path", server.MetricsPath, "metrics path must start with '/'")
	}

	if server.HealthPath != "" && !strings.HasPrefix(server.HealthPath, "/") {
		v.addError("server.health_path", server.HealthPath, "health path must start with '/'")
	}

	// Validate CORS origins
	for i, origin := range server.CORSOrigins {
		if origin != "*" && !isValidURL(origin) {
			v.addError(fmt.Sprintf("server.cors_origins[%d]", i), origin, "invalid CORS origin URL")
		}
	}
}

// validateDatabase validates database configuration
func (v *ConfigValidator) validateDatabase(db DatabaseConfig) {
	if db.Driver == "" {
		v.addError("database.driver", db.Driver, "database driver is required")
	} else if !isValidDatabaseDriver(db.Driver) {
		v.addError("database.driver", db.Driver, "unsupported database driver")
	}

	if db.Host == "" {
		v.addError("database.host", db.Host, "database host is required")
	}

	if db.Port <= 0 || db.Port > 65535 {
		v.addError("database.port", db.Port, "database port must be between 1 and 65535")
	}

	if db.Database == "" {
		v.addError("database.database", db.Database, "database name is required")
	}

	if db.Username == "" {
		v.addError("database.username", db.Username, "database username is required")
	}

	if db.MaxOpenConns <= 0 {
		v.addError("database.max_open_conns", db.MaxOpenConns, "max open connections must be positive")
	}

	if db.MaxIdleConns < 0 {
		v.addError("database.max_idle_conns", db.MaxIdleConns, "max idle connections cannot be negative")
	}

	if db.MaxIdleConns > db.MaxOpenConns {
		v.addError("database.max_idle_conns", db.MaxIdleConns, "max idle connections cannot exceed max open connections")
	}

	if db.ConnMaxLifetime <= 0 {
		v.addError("database.conn_max_lifetime", db.ConnMaxLifetime, "connection max lifetime must be positive")
	}
}

// validateKafka validates Kafka configuration
func (v *ConfigValidator) validateKafka(kafka KafkaConfig) {
	if len(kafka.Brokers) == 0 {
		v.addError("kafka.brokers", kafka.Brokers, "at least one Kafka broker is required")
	}

	for i, broker := range kafka.Brokers {
		if !isValidBrokerAddress(broker) {
			v.addError(fmt.Sprintf("kafka.brokers[%d]", i), broker, "invalid broker address format")
		}
	}

	if kafka.ClientID == "" {
		v.addError("kafka.client_id", kafka.ClientID, "Kafka client ID is required")
	}

	if kafka.ConnectTimeout <= 0 {
		v.addError("kafka.connect_timeout", kafka.ConnectTimeout, "connect timeout must be positive")
	}

	if kafka.ReadTimeout <= 0 {
		v.addError("kafka.read_timeout", kafka.ReadTimeout, "read timeout must be positive")
	}

	if kafka.WriteTimeout <= 0 {
		v.addError("kafka.write_timeout", kafka.WriteTimeout, "write timeout must be positive")
	}

	if kafka.BatchSize <= 0 {
		v.addError("kafka.batch_size", kafka.BatchSize, "batch size must be positive")
	}

	if kafka.RetryMax < 0 {
		v.addError("kafka.retry_max", kafka.RetryMax, "retry max cannot be negative")
	}

	// Validate topics
	v.validateKafkaTopics(kafka.Topics)
}

// validateKafkaTopics validates Kafka topic configuration
func (v *ConfigValidator) validateKafkaTopics(topics TopicsConfig) {
	topicFields := map[string]string{
		"beverage_created":     topics.BeverageCreated,
		"beverage_updated":     topics.BeverageUpdated,
		"task_created":         topics.TaskCreated,
		"task_updated":         topics.TaskUpdated,
		"notification_sent":    topics.NotificationSent,
		"ai_request_completed": topics.AIRequestCompleted,
		"system_event":         topics.SystemEvent,
	}

	for field, topic := range topicFields {
		if topic == "" {
			v.addError(fmt.Sprintf("kafka.topics.%s", field), topic, "topic name is required")
		} else if !isValidTopicName(topic) {
			v.addError(fmt.Sprintf("kafka.topics.%s", field), topic, "invalid topic name format")
		}
	}
}

// validateAI validates AI configuration
func (v *ConfigValidator) validateAI(ai AIConfig) {
	if ai.DefaultProvider == "" {
		v.addError("ai.default_provider", ai.DefaultProvider, "default AI provider is required")
	}

	if len(ai.Providers) == 0 {
		v.addError("ai.providers", ai.Providers, "at least one AI provider must be configured")
	}

	// Check if default provider exists
	if ai.DefaultProvider != "" {
		if _, exists := ai.Providers[ai.DefaultProvider]; !exists {
			v.addError("ai.default_provider", ai.DefaultProvider, "default provider not found in providers list")
		}
	}

	// Validate each provider
	for name, provider := range ai.Providers {
		v.validateAIProvider(name, provider)
	}

	// Validate rate limits
	if ai.RateLimits.RequestsPerMinute <= 0 {
		v.addError("ai.rate_limits.requests_per_minute", ai.RateLimits.RequestsPerMinute, "requests per minute must be positive")
	}

	if ai.RateLimits.BurstSize <= 0 {
		v.addError("ai.rate_limits.burst_size", ai.RateLimits.BurstSize, "burst size must be positive")
	}
}

// validateAIProvider validates a specific AI provider configuration
func (v *ConfigValidator) validateAIProvider(name string, provider AIProviderConfig) {
	prefix := fmt.Sprintf("ai.providers.%s", name)

	if provider.Enabled && provider.APIKey == "" {
		v.addError(prefix+".api_key", provider.APIKey, "API key is required for enabled provider")
	}

	if provider.BaseURL != "" && !isValidURL(provider.BaseURL) {
		v.addError(prefix+".base_url", provider.BaseURL, "invalid base URL format")
	}

	if provider.MaxTokens <= 0 {
		v.addError(prefix+".max_tokens", provider.MaxTokens, "max tokens must be positive")
	}

	if provider.Temperature < 0 || provider.Temperature > 2 {
		v.addError(prefix+".temperature", provider.Temperature, "temperature must be between 0 and 2")
	}

	if provider.Timeout <= 0 {
		v.addError(prefix+".timeout", provider.Timeout, "timeout must be positive")
	}
}

// validateExternal validates external service configuration
func (v *ConfigValidator) validateExternal(external ExternalConfig) {
	v.validateClickUp(external.ClickUp)
	v.validateSlack(external.Slack)
	v.validateGoogleSheets(external.GoogleSheets)
	v.validateEmail(external.Email)
}

// validateClickUp validates ClickUp configuration
func (v *ConfigValidator) validateClickUp(clickup ClickUpConfig) {
	if clickup.Enabled {
		if clickup.APIKey == "" {
			v.addError("external.clickup.api_key", clickup.APIKey, "ClickUp API key is required when enabled")
		}

		if clickup.BaseURL != "" && !isValidURL(clickup.BaseURL) {
			v.addError("external.clickup.base_url", clickup.BaseURL, "invalid ClickUp base URL")
		}

		if clickup.Timeout <= 0 {
			v.addError("external.clickup.timeout", clickup.Timeout, "ClickUp timeout must be positive")
		}

		if clickup.RetryCount < 0 {
			v.addError("external.clickup.retry_count", clickup.RetryCount, "ClickUp retry count cannot be negative")
		}
	}
}

// validateSlack validates Slack configuration
func (v *ConfigValidator) validateSlack(slack SlackConfig) {
	if slack.Enabled {
		if slack.BotToken == "" {
			v.addError("external.slack.bot_token", slack.BotToken, "Slack bot token is required when enabled")
		}

		if slack.Timeout <= 0 {
			v.addError("external.slack.timeout", slack.Timeout, "Slack timeout must be positive")
		}
	}
}

// validateGoogleSheets validates Google Sheets configuration
func (v *ConfigValidator) validateGoogleSheets(sheets GoogleSheetsConfig) {
	if sheets.Enabled {
		if sheets.CredentialsPath == "" {
			v.addError("external.google_sheets.credentials_path", sheets.CredentialsPath, "Google Sheets credentials path is required when enabled")
		}

		if sheets.Timeout <= 0 {
			v.addError("external.google_sheets.timeout", sheets.Timeout, "Google Sheets timeout must be positive")
		}
	}
}

// validateEmail validates email configuration
func (v *ConfigValidator) validateEmail(email EmailConfig) {
	if email.Enabled {
		if email.Provider == "" {
			v.addError("external.email.provider", email.Provider, "email provider is required when enabled")
		}

		if email.FromEmail == "" {
			v.addError("external.email.from_email", email.FromEmail, "from email is required when enabled")
		} else if !isValidEmail(email.FromEmail) {
			v.addError("external.email.from_email", email.FromEmail, "invalid from email format")
		}

		if email.Provider == "smtp" {
			if email.SMTPHost == "" {
				v.addError("external.email.smtp_host", email.SMTPHost, "SMTP host is required for SMTP provider")
			}

			if email.SMTPPort <= 0 || email.SMTPPort > 65535 {
				v.addError("external.email.smtp_port", email.SMTPPort, "SMTP port must be between 1 and 65535")
			}
		}
	}
}

// validateSecurity validates security configuration
func (v *ConfigValidator) validateSecurity(security SecurityConfig) {
	// Validate JWT
	if security.JWT.SecretKey == "" {
		v.addError("security.jwt.secret_key", security.JWT.SecretKey, "JWT secret key is required")
	} else if len(security.JWT.SecretKey) < 32 {
		v.addError("security.jwt.secret_key", "***", "JWT secret key must be at least 32 characters")
	}

	if security.JWT.ExpirationTime <= 0 {
		v.addError("security.jwt.expiration_time", security.JWT.ExpirationTime, "JWT expiration time must be positive")
	}

	// Validate rate limiting
	if security.RateLimit.Enabled {
		if security.RateLimit.RequestsPerMin <= 0 {
			v.addError("security.rate_limit.requests_per_min", security.RateLimit.RequestsPerMin, "requests per minute must be positive")
		}

		if security.RateLimit.BurstSize <= 0 {
			v.addError("security.rate_limit.burst_size", security.RateLimit.BurstSize, "burst size must be positive")
		}
	}
}

// validateFeatures validates feature configuration
func (v *ConfigValidator) validateFeatures(features FeatureConfig) {
	// No specific validation needed for feature flags currently
	// They are all boolean values with sensible defaults
}

// validateEnvironmentSpecific validates environment-specific requirements
func (v *ConfigValidator) validateEnvironmentSpecific(config *Config) {
	switch config.Environment {
	case Production:
		v.validateProductionRequirements(config)
	case Staging:
		v.validateStagingRequirements(config)
	}
}

// validateProductionRequirements validates production-specific requirements
func (v *ConfigValidator) validateProductionRequirements(config *Config) {
	if config.Service.Debug {
		v.addError("service.debug", config.Service.Debug, "debug mode should be disabled in production")
	}

	if config.Database.SSLMode == "disable" {
		v.addError("database.ssl_mode", config.Database.SSLMode, "SSL should be enabled in production")
	}

	if config.Security.JWT.SecretKey == "dev-secret-key-change-in-production" {
		v.addError("security.jwt.secret_key", "***", "default development secret key should not be used in production")
	}

	if !config.Security.API.RequireHTTPS {
		v.addError("security.api.require_https", config.Security.API.RequireHTTPS, "HTTPS should be required in production")
	}
}

// validateStagingRequirements validates staging-specific requirements
func (v *ConfigValidator) validateStagingRequirements(config *Config) {
	if config.Service.Debug {
		v.addError("service.debug", config.Service.Debug, "debug mode should be disabled in staging")
	}
}

// addError adds a validation error
func (v *ConfigValidator) addError(field string, value interface{}, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Tag:     "custom",
		Value:   fmt.Sprintf("%v", value),
		Message: message,
	})
}

// Helper validation functions

func isValidServiceName(name string) bool {
	matched, _ := regexp.MatchString(`^[a-z0-9-]+$`, name)
	return matched
}

func isValidVersion(version string) bool {
	matched, _ := regexp.MatchString(`^\d+\.\d+\.\d+`, version)
	return matched
}

func isValidHost(host string) bool {
	if host == "localhost" || host == "0.0.0.0" {
		return true
	}
	return net.ParseIP(host) != nil
}

func isValidURL(urlStr string) bool {
	_, err := url.Parse(urlStr)
	return err == nil
}

func isValidDatabaseDriver(driver string) bool {
	validDrivers := []string{"postgres", "mysql", "sqlite3"}
	for _, valid := range validDrivers {
		if driver == valid {
			return true
		}
	}
	return false
}

func isValidBrokerAddress(broker string) bool {
	parts := strings.Split(broker, ":")
	if len(parts) != 2 {
		return false
	}

	host, portStr := parts[0], parts[1]
	if host == "" {
		return false
	}

	port, err := strconv.Atoi(portStr)
	return err == nil && port > 0 && port <= 65535
}

func isValidTopicName(topic string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._-]+$`, topic)
	return matched
}

func isValidEmail(email string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, email)
	return matched
}
