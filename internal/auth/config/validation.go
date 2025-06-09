package config

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateConfig validates the configuration
func ValidateConfig(cfg *Config) error {
	var errors []string

	// Validate server configuration
	if err := validateServerConfig(&cfg.Server); err != nil {
		errors = append(errors, fmt.Sprintf("server config: %v", err))
	}

	// Validate Redis configuration
	if err := validateRedisConfig(&cfg.Redis); err != nil {
		errors = append(errors, fmt.Sprintf("redis config: %v", err))
	}

	// Validate security configuration
	if err := validateSecurityConfig(&cfg.Security); err != nil {
		errors = append(errors, fmt.Sprintf("security config: %v", err))
	}

	// Validate rate limiting configuration
	if err := validateRateLimitingConfig(&cfg.RateLimiting); err != nil {
		errors = append(errors, fmt.Sprintf("rate limiting config: %v", err))
	}

	// Validate logging configuration
	if err := validateLoggingConfig(&cfg.Logging); err != nil {
		errors = append(errors, fmt.Sprintf("logging config: %v", err))
	}

	// Validate monitoring configuration
	if err := validateMonitoringConfig(&cfg.Monitoring); err != nil {
		errors = append(errors, fmt.Sprintf("monitoring config: %v", err))
	}

	// Validate TLS configuration
	if err := validateTLSConfig(&cfg.TLS); err != nil {
		errors = append(errors, fmt.Sprintf("tls config: %v", err))
	}

	// Validate environment
	if err := validateEnvironment(cfg.Environment); err != nil {
		errors = append(errors, fmt.Sprintf("environment: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// validateServerConfig validates server configuration
func validateServerConfig(cfg *ServerConfig) error {
	var errors []string

	if cfg.HTTPPort < 1 || cfg.HTTPPort > 65535 {
		errors = append(errors, "http_port must be between 1 and 65535")
	}

	if cfg.GRPCPort < 1 || cfg.GRPCPort > 65535 {
		errors = append(errors, "grpc_port must be between 1 and 65535")
	}

	if cfg.HTTPPort == cfg.GRPCPort {
		errors = append(errors, "http_port and grpc_port cannot be the same")
	}

	if cfg.Host == "" {
		errors = append(errors, "host is required")
	}

	if cfg.ReadTimeout <= 0 {
		errors = append(errors, "read_timeout must be positive")
	}

	if cfg.WriteTimeout <= 0 {
		errors = append(errors, "write_timeout must be positive")
	}

	if cfg.IdleTimeout <= 0 {
		errors = append(errors, "idle_timeout must be positive")
	}

	if cfg.ShutdownTimeout <= 0 {
		errors = append(errors, "shutdown_timeout must be positive")
	}

	if cfg.MaxHeaderBytes <= 0 {
		errors = append(errors, "max_header_bytes must be positive")
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}

	return nil
}

// validateRedisConfig validates Redis configuration
func validateRedisConfig(cfg *RedisConfig) error {
	var errors []string

	if cfg.URL == "" {
		errors = append(errors, "url is required")
	} else {
		if _, err := url.Parse(cfg.URL); err != nil {
			errors = append(errors, fmt.Sprintf("invalid url: %v", err))
		}
	}

	if cfg.DB < 0 || cfg.DB > 15 {
		errors = append(errors, "db must be between 0 and 15")
	}

	if cfg.MaxRetries < 0 {
		errors = append(errors, "max_retries must be non-negative")
	}

	if cfg.PoolSize <= 0 {
		errors = append(errors, "pool_size must be positive")
	}

	if cfg.MinIdleConns < 0 {
		errors = append(errors, "min_idle_conns must be non-negative")
	}

	if cfg.MinIdleConns > cfg.PoolSize {
		errors = append(errors, "min_idle_conns cannot be greater than pool_size")
	}

	if cfg.DialTimeout <= 0 {
		errors = append(errors, "dial_timeout must be positive")
	}

	if cfg.ReadTimeout <= 0 {
		errors = append(errors, "read_timeout must be positive")
	}

	if cfg.WriteTimeout <= 0 {
		errors = append(errors, "write_timeout must be positive")
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}

	return nil
}

// validateSecurityConfig validates security configuration
func validateSecurityConfig(cfg *SecurityConfig) error {
	var errors []string

	// Validate JWT configuration
	if cfg.JWT.Secret == "" {
		errors = append(errors, "jwt.secret is required")
	} else if len(cfg.JWT.Secret) < 32 {
		errors = append(errors, "jwt.secret must be at least 32 characters")
	}

	if cfg.JWT.AccessTokenTTL <= 0 {
		errors = append(errors, "jwt.access_token_ttl must be positive")
	}

	if cfg.JWT.RefreshTokenTTL <= 0 {
		errors = append(errors, "jwt.refresh_token_ttl must be positive")
	}

	if cfg.JWT.AccessTokenTTL >= cfg.JWT.RefreshTokenTTL {
		errors = append(errors, "jwt.refresh_token_ttl must be greater than access_token_ttl")
	}

	if cfg.JWT.Issuer == "" {
		errors = append(errors, "jwt.issuer is required")
	}

	if cfg.JWT.Audience == "" {
		errors = append(errors, "jwt.audience is required")
	}

	// Validate password configuration
	if cfg.Password.BcryptCost < 10 || cfg.Password.BcryptCost > 15 {
		errors = append(errors, "password.bcrypt_cost must be between 10 and 15")
	}

	if cfg.Password.Policy.MinLength < 1 {
		errors = append(errors, "password.policy.min_length must be at least 1")
	}

	if cfg.Password.Policy.MaxLength < cfg.Password.Policy.MinLength {
		errors = append(errors, "password.policy.max_length must be greater than or equal to min_length")
	}

	// Validate account configuration
	if cfg.Account.MaxLoginAttempts <= 0 {
		errors = append(errors, "account.max_login_attempts must be positive")
	}

	if cfg.Account.LockoutDuration <= 0 {
		errors = append(errors, "account.lockout_duration must be positive")
	}

	if cfg.Account.SessionTimeout <= 0 {
		errors = append(errors, "account.session_timeout must be positive")
	}

	if cfg.Account.MaxSessionsPerUser <= 0 {
		errors = append(errors, "account.max_sessions_per_user must be positive")
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}

	return nil
}

// validateRateLimitingConfig validates rate limiting configuration
func validateRateLimitingConfig(cfg *RateLimitingConfig) error {
	if !cfg.Enabled {
		return nil
	}

	var errors []string

	if cfg.RequestsPerMinute <= 0 {
		errors = append(errors, "requests_per_minute must be positive when rate limiting is enabled")
	}

	if cfg.BurstSize <= 0 {
		errors = append(errors, "burst_size must be positive when rate limiting is enabled")
	}

	if cfg.CleanupInterval <= 0 {
		errors = append(errors, "cleanup_interval must be positive when rate limiting is enabled")
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}

	return nil
}

// validateLoggingConfig validates logging configuration
func validateLoggingConfig(cfg *LoggingConfig) error {
	var errors []string

	validLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
	if !contains(validLevels, cfg.Level) {
		errors = append(errors, fmt.Sprintf("level must be one of: %s", strings.Join(validLevels, ", ")))
	}

	validFormats := []string{"json", "text"}
	if !contains(validFormats, cfg.Format) {
		errors = append(errors, fmt.Sprintf("format must be one of: %s", strings.Join(validFormats, ", ")))
	}

	validOutputs := []string{"stdout", "stderr", "file"}
	if !contains(validOutputs, cfg.Output) {
		errors = append(errors, fmt.Sprintf("output must be one of: %s", strings.Join(validOutputs, ", ")))
	}

	if cfg.Output == "file" && cfg.FilePath == "" {
		errors = append(errors, "file_path is required when output is 'file'")
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}

	return nil
}

// validateMonitoringConfig validates monitoring configuration
func validateMonitoringConfig(cfg *MonitoringConfig) error {
	if !cfg.Enabled {
		return nil
	}

	var errors []string

	if cfg.Port < 1 || cfg.Port > 65535 {
		errors = append(errors, "port must be between 1 and 65535 when monitoring is enabled")
	}

	if cfg.Path == "" {
		errors = append(errors, "path is required when monitoring is enabled")
	}

	if cfg.Tracing.Enabled {
		if cfg.Tracing.ServiceName == "" {
			errors = append(errors, "tracing.service_name is required when tracing is enabled")
		}

		if cfg.Tracing.SampleRate < 0 || cfg.Tracing.SampleRate > 1 {
			errors = append(errors, "tracing.sample_rate must be between 0 and 1")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}

	return nil
}

// validateTLSConfig validates TLS configuration
func validateTLSConfig(cfg *TLSConfig) error {
	if !cfg.Enabled {
		return nil
	}

	var errors []string

	if cfg.CertFile == "" {
		errors = append(errors, "cert_file is required when TLS is enabled")
	}

	if cfg.KeyFile == "" {
		errors = append(errors, "key_file is required when TLS is enabled")
	}

	validVersions := []string{"1.0", "1.1", "1.2", "1.3"}
	if !contains(validVersions, cfg.MinVersion) {
		errors = append(errors, fmt.Sprintf("min_version must be one of: %s", strings.Join(validVersions, ", ")))
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}

	return nil
}

// validateEnvironment validates environment configuration
func validateEnvironment(env string) error {
	validEnvironments := []string{"development", "staging", "production"}
	if !contains(validEnvironments, env) {
		return fmt.Errorf("must be one of: %s", strings.Join(validEnvironments, ", "))
	}
	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
