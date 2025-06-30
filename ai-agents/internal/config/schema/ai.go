package schema

import (
	"fmt"
	"time"
)

// AIConfig holds AI provider configuration
type AIConfig struct {
	// Default provider
	DefaultProvider string `yaml:"default_provider" json:"default_provider" validate:"required,ai_provider" default:"openai"`
	
	// Provider configurations
	Providers map[string]AIProviderConfig `yaml:"providers" json:"providers" validate:"required,nonempty"`
	
	// Global AI settings
	Global AIGlobalConfig `yaml:"global" json:"global"`
	
	// Rate limiting
	RateLimit AIRateLimitConfig `yaml:"rate_limit" json:"rate_limit"`
	
	// Cost tracking
	CostTracking AICostTrackingConfig `yaml:"cost_tracking" json:"cost_tracking"`
	
	// Fallback configuration
	Fallback AIFallbackConfig `yaml:"fallback" json:"fallback"`
	
	// Monitoring
	Monitoring AIMonitoringConfig `yaml:"monitoring" json:"monitoring"`
}

// AIProviderConfig holds configuration for a specific AI provider
type AIProviderConfig struct {
	Name        string `yaml:"name" json:"name" validate:"required"`
	Type        string `yaml:"type" json:"type" validate:"required,ai_provider"`
	Enabled     bool   `yaml:"enabled" json:"enabled" default:"true"`
	
	// Authentication
	APIKey      string `yaml:"api_key" json:"api_key" validate:"secret"`
	APIKeyPath  string `yaml:"api_key_path" json:"api_key_path"` // Path to secret in Vault
	
	// Connection settings
	BaseURL     string        `yaml:"base_url" json:"base_url" validate:"url"`
	Timeout     time.Duration `yaml:"timeout" json:"timeout" default:"60s"`
	RetryAttempts int         `yaml:"retry_attempts" json:"retry_attempts" validate:"min=0,max=10" default:"3"`
	
	// Model configuration
	Models      map[string]AIModelConfig `yaml:"models" json:"models"`
	DefaultModel string                  `yaml:"default_model" json:"default_model"`
	
	// Rate limiting (provider-specific)
	RateLimit   AIProviderRateLimitConfig `yaml:"rate_limit" json:"rate_limit"`
	
	// Cost settings
	CostPerToken float64 `yaml:"cost_per_token" json:"cost_per_token" validate:"min=0"`
	Currency     string  `yaml:"currency" json:"currency" default:"USD"`
	
	// Provider-specific settings
	Settings    map[string]interface{} `yaml:"settings" json:"settings"`
	
	// Health check
	HealthCheck AIProviderHealthCheckConfig `yaml:"health_check" json:"health_check"`
}

// AIModelConfig holds configuration for a specific AI model
type AIModelConfig struct {
	Name            string  `yaml:"name" json:"name" validate:"required"`
	DisplayName     string  `yaml:"display_name" json:"display_name"`
	Description     string  `yaml:"description" json:"description"`
	Enabled         bool    `yaml:"enabled" json:"enabled" default:"true"`
	
	// Model capabilities
	MaxTokens       int     `yaml:"max_tokens" json:"max_tokens" validate:"min=1" default:"4096"`
	ContextWindow   int     `yaml:"context_window" json:"context_window" validate:"min=1" default:"4096"`
	SupportsStreaming bool  `yaml:"supports_streaming" json:"supports_streaming" default:"false"`
	SupportsImages  bool    `yaml:"supports_images" json:"supports_images" default:"false"`
	SupportsFunctions bool  `yaml:"supports_functions" json:"supports_functions" default:"false"`
	
	// Cost settings
	InputCostPer1K  float64 `yaml:"input_cost_per_1k" json:"input_cost_per_1k" validate:"min=0"`
	OutputCostPer1K float64 `yaml:"output_cost_per_1k" json:"output_cost_per_1k" validate:"min=0"`
	
	// Quality settings
	Temperature     float64 `yaml:"temperature" json:"temperature" validate:"min=0,max=2" default:"0.7"`
	TopP            float64 `yaml:"top_p" json:"top_p" validate:"min=0,max=1" default:"1.0"`
	FrequencyPenalty float64 `yaml:"frequency_penalty" json:"frequency_penalty" validate:"min=-2,max=2" default:"0"`
	PresencePenalty float64 `yaml:"presence_penalty" json:"presence_penalty" validate:"min=-2,max=2" default:"0"`
	
	// Usage patterns
	RecommendedFor  []string `yaml:"recommended_for" json:"recommended_for"`
	NotRecommendedFor []string `yaml:"not_recommended_for" json:"not_recommended_for"`
}

// AIGlobalConfig holds global AI settings
type AIGlobalConfig struct {
	// Default parameters
	DefaultTemperature    float64 `yaml:"default_temperature" json:"default_temperature" validate:"min=0,max=2" default:"0.7"`
	DefaultMaxTokens      int     `yaml:"default_max_tokens" json:"default_max_tokens" validate:"min=1" default:"1000"`
	DefaultTimeout        time.Duration `yaml:"default_timeout" json:"default_timeout" default:"60s"`
	
	// Safety settings
	ContentFiltering      bool     `yaml:"content_filtering" json:"content_filtering" default:"true"`
	SafetyLevel          string   `yaml:"safety_level" json:"safety_level" validate:"oneof=low medium high" default:"medium"`
	BlockedWords         []string `yaml:"blocked_words" json:"blocked_words"`
	
	// Logging and auditing
	LogRequests          bool `yaml:"log_requests" json:"log_requests" default:"true"`
	LogResponses         bool `yaml:"log_responses" json:"log_responses" default:"false"`
	AuditTrail           bool `yaml:"audit_trail" json:"audit_trail" default:"true"`
	
	// Caching
	EnableCaching        bool          `yaml:"enable_caching" json:"enable_caching" default:"true"`
	CacheTTL             time.Duration `yaml:"cache_ttl" json:"cache_ttl" default:"1h"`
	CacheKeyPrefix       string        `yaml:"cache_key_prefix" json:"cache_key_prefix" default:"ai_cache"`
	
	// Retry settings
	RetryEnabled         bool          `yaml:"retry_enabled" json:"retry_enabled" default:"true"`
	RetryMaxAttempts     int           `yaml:"retry_max_attempts" json:"retry_max_attempts" validate:"min=1,max=10" default:"3"`
	RetryBaseDelay       time.Duration `yaml:"retry_base_delay" json:"retry_base_delay" default:"1s"`
	RetryMaxDelay        time.Duration `yaml:"retry_max_delay" json:"retry_max_delay" default:"30s"`
}

// AIRateLimitConfig holds global AI rate limiting configuration
type AIRateLimitConfig struct {
	Enabled             bool          `yaml:"enabled" json:"enabled" default:"true"`
	RequestsPerMinute   int           `yaml:"requests_per_minute" json:"requests_per_minute" validate:"min=1" default:"60"`
	TokensPerMinute     int           `yaml:"tokens_per_minute" json:"tokens_per_minute" validate:"min=1" default:"100000"`
	ConcurrentRequests  int           `yaml:"concurrent_requests" json:"concurrent_requests" validate:"min=1" default:"10"`
	
	// Burst settings
	BurstRequests       int           `yaml:"burst_requests" json:"burst_requests" validate:"min=1" default:"10"`
	BurstTokens         int           `yaml:"burst_tokens" json:"burst_tokens" validate:"min=1" default:"10000"`
	
	// Window settings
	WindowSize          time.Duration `yaml:"window_size" json:"window_size" default:"1m"`
	CleanupInterval     time.Duration `yaml:"cleanup_interval" json:"cleanup_interval" default:"5m"`
	
	// Per-user limits
	PerUserLimits       bool          `yaml:"per_user_limits" json:"per_user_limits" default:"true"`
	UserRequestsPerMinute int         `yaml:"user_requests_per_minute" json:"user_requests_per_minute" validate:"min=1" default:"10"`
	UserTokensPerMinute int           `yaml:"user_tokens_per_minute" json:"user_tokens_per_minute" validate:"min=1" default:"10000"`
}

// AIProviderRateLimitConfig holds provider-specific rate limiting
type AIProviderRateLimitConfig struct {
	Enabled             bool          `yaml:"enabled" json:"enabled" default:"true"`
	RequestsPerMinute   int           `yaml:"requests_per_minute" json:"requests_per_minute" validate:"min=1"`
	TokensPerMinute     int           `yaml:"tokens_per_minute" json:"tokens_per_minute" validate:"min=1"`
	ConcurrentRequests  int           `yaml:"concurrent_requests" json:"concurrent_requests" validate:"min=1"`
	
	// Provider-specific limits
	DailyLimit          int           `yaml:"daily_limit" json:"daily_limit" validate:"min=0"`
	MonthlyLimit        int           `yaml:"monthly_limit" json:"monthly_limit" validate:"min=0"`
	
	// Cost limits
	DailyCostLimit      float64       `yaml:"daily_cost_limit" json:"daily_cost_limit" validate:"min=0"`
	MonthlyCostLimit    float64       `yaml:"monthly_cost_limit" json:"monthly_cost_limit" validate:"min=0"`
}

// AICostTrackingConfig holds AI cost tracking configuration
type AICostTrackingConfig struct {
	Enabled             bool    `yaml:"enabled" json:"enabled" default:"true"`
	Currency            string  `yaml:"currency" json:"currency" default:"USD"`
	
	// Budget settings
	DailyBudget         float64 `yaml:"daily_budget" json:"daily_budget" validate:"min=0"`
	MonthlyBudget       float64 `yaml:"monthly_budget" json:"monthly_budget" validate:"min=0"`
	
	// Alert thresholds
	AlertThresholds     []float64 `yaml:"alert_thresholds" json:"alert_thresholds"` // e.g., [0.5, 0.8, 0.9] for 50%, 80%, 90%
	
	// Reporting
	ReportingEnabled    bool      `yaml:"reporting_enabled" json:"reporting_enabled" default:"true"`
	ReportingInterval   time.Duration `yaml:"reporting_interval" json:"reporting_interval" default:"24h"`
	
	// Storage
	StorageRetention    time.Duration `yaml:"storage_retention" json:"storage_retention" default:"90d"`
}

// AIFallbackConfig holds AI fallback configuration
type AIFallbackConfig struct {
	Enabled             bool     `yaml:"enabled" json:"enabled" default:"true"`
	FallbackProviders   []string `yaml:"fallback_providers" json:"fallback_providers"`
	FallbackStrategy    string   `yaml:"fallback_strategy" json:"fallback_strategy" validate:"oneof=round_robin priority cost_optimized" default:"priority"`
	
	// Fallback triggers
	TriggerOnError      bool     `yaml:"trigger_on_error" json:"trigger_on_error" default:"true"`
	TriggerOnTimeout    bool     `yaml:"trigger_on_timeout" json:"trigger_on_timeout" default:"true"`
	TriggerOnRateLimit  bool     `yaml:"trigger_on_rate_limit" json:"trigger_on_rate_limit" default:"true"`
	
	// Fallback behavior
	MaxFallbackAttempts int      `yaml:"max_fallback_attempts" json:"max_fallback_attempts" validate:"min=1" default:"2"`
	FallbackDelay       time.Duration `yaml:"fallback_delay" json:"fallback_delay" default:"1s"`
}

// AIProviderHealthCheckConfig holds provider health check configuration
type AIProviderHealthCheckConfig struct {
	Enabled             bool          `yaml:"enabled" json:"enabled" default:"true"`
	Interval            time.Duration `yaml:"interval" json:"interval" default:"5m"`
	Timeout             time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	FailureThreshold    int           `yaml:"failure_threshold" json:"failure_threshold" validate:"min=1" default:"3"`
	SuccessThreshold    int           `yaml:"success_threshold" json:"success_threshold" validate:"min=1" default:"1"`
	
	// Health check method
	Method              string        `yaml:"method" json:"method" validate:"oneof=ping simple_request model_list" default:"ping"`
	TestPrompt          string        `yaml:"test_prompt" json:"test_prompt" default:"Hello"`
	ExpectedResponse    string        `yaml:"expected_response" json:"expected_response"`
}

// AIMonitoringConfig holds AI monitoring configuration
type AIMonitoringConfig struct {
	Enabled             bool          `yaml:"enabled" json:"enabled" default:"true"`
	MetricsInterval     time.Duration `yaml:"metrics_interval" json:"metrics_interval" default:"1m"`
	
	// Performance monitoring
	TrackLatency        bool          `yaml:"track_latency" json:"track_latency" default:"true"`
	TrackTokenUsage     bool          `yaml:"track_token_usage" json:"track_token_usage" default:"true"`
	TrackCosts          bool          `yaml:"track_costs" json:"track_costs" default:"true"`
	TrackErrors         bool          `yaml:"track_errors" json:"track_errors" default:"true"`
	
	// Quality monitoring
	TrackResponseQuality bool         `yaml:"track_response_quality" json:"track_response_quality" default:"false"`
	QualityMetrics      []string      `yaml:"quality_metrics" json:"quality_metrics"`
	
	// Alerting
	Alerting            AIAlertingConfig `yaml:"alerting" json:"alerting"`
}

// AIAlertingConfig holds AI alerting configuration
type AIAlertingConfig struct {
	Enabled             bool          `yaml:"enabled" json:"enabled" default:"false"`
	Channels            []string      `yaml:"channels" json:"channels"`
	
	// Alert thresholds
	ErrorRateThreshold  float64       `yaml:"error_rate_threshold" json:"error_rate_threshold" default:"0.1"` // 10%
	LatencyThreshold    time.Duration `yaml:"latency_threshold" json:"latency_threshold" default:"30s"`
	CostThreshold       float64       `yaml:"cost_threshold" json:"cost_threshold" default:"100.0"`
	
	// Alert settings
	CooldownPeriod      time.Duration `yaml:"cooldown_period" json:"cooldown_period" default:"15m"`
	MaxAlerts           int           `yaml:"max_alerts" json:"max_alerts" validate:"min=1" default:"10"`
}

// GetProvider returns a specific AI provider configuration
func (ac *AIConfig) GetProvider(name string) (AIProviderConfig, bool) {
	provider, exists := ac.Providers[name]
	return provider, exists
}

// GetDefaultProvider returns the default AI provider configuration
func (ac *AIConfig) GetDefaultProvider() (AIProviderConfig, error) {
	if ac.DefaultProvider == "" {
		return AIProviderConfig{}, fmt.Errorf("no default provider configured")
	}
	
	provider, exists := ac.Providers[ac.DefaultProvider]
	if !exists {
		return AIProviderConfig{}, fmt.Errorf("default provider '%s' not found", ac.DefaultProvider)
	}
	
	return provider, nil
}

// GetEnabledProviders returns all enabled AI providers
func (ac *AIConfig) GetEnabledProviders() map[string]AIProviderConfig {
	enabled := make(map[string]AIProviderConfig)
	
	for name, provider := range ac.Providers {
		if provider.Enabled {
			enabled[name] = provider
		}
	}
	
	return enabled
}

// GetModel returns a specific model configuration from a provider
func (apc *AIProviderConfig) GetModel(name string) (AIModelConfig, bool) {
	model, exists := apc.Models[name]
	return model, exists
}

// GetDefaultModel returns the default model configuration
func (apc *AIProviderConfig) GetDefaultModel() (AIModelConfig, error) {
	if apc.DefaultModel == "" {
		return AIModelConfig{}, fmt.Errorf("no default model configured for provider %s", apc.Name)
	}
	
	model, exists := apc.Models[apc.DefaultModel]
	if !exists {
		return AIModelConfig{}, fmt.Errorf("default model '%s' not found for provider %s", apc.DefaultModel, apc.Name)
	}
	
	return model, nil
}

// GetEnabledModels returns all enabled models for a provider
func (apc *AIProviderConfig) GetEnabledModels() map[string]AIModelConfig {
	enabled := make(map[string]AIModelConfig)
	
	for name, model := range apc.Models {
		if model.Enabled {
			enabled[name] = model
		}
	}
	
	return enabled
}

// CalculateCost calculates the cost for a given number of input and output tokens
func (amc *AIModelConfig) CalculateCost(inputTokens, outputTokens int) float64 {
	inputCost := float64(inputTokens) / 1000.0 * amc.InputCostPer1K
	outputCost := float64(outputTokens) / 1000.0 * amc.OutputCostPer1K
	return inputCost + outputCost
}

// Validate validates the AI configuration
func (ac *AIConfig) Validate() error {
	if ac.DefaultProvider == "" {
		return fmt.Errorf("default AI provider is required")
	}
	
	if len(ac.Providers) == 0 {
		return fmt.Errorf("at least one AI provider must be configured")
	}
	
	// Check if default provider exists
	if _, exists := ac.Providers[ac.DefaultProvider]; !exists {
		return fmt.Errorf("default provider '%s' not found in providers", ac.DefaultProvider)
	}
	
	// Validate each provider
	for name, provider := range ac.Providers {
		if err := provider.Validate(); err != nil {
			return fmt.Errorf("provider '%s': %w", name, err)
		}
	}
	
	return nil
}

// Validate validates an AI provider configuration
func (apc *AIProviderConfig) Validate() error {
	if apc.Name == "" {
		return fmt.Errorf("provider name is required")
	}
	
	if apc.Type == "" {
		return fmt.Errorf("provider type is required")
	}
	
	if apc.APIKey == "" && apc.APIKeyPath == "" {
		return fmt.Errorf("either api_key or api_key_path must be provided")
	}
	
	// Validate models
	for name, model := range apc.Models {
		if err := model.Validate(); err != nil {
			return fmt.Errorf("model '%s': %w", name, err)
		}
	}
	
	// Check if default model exists
	if apc.DefaultModel != "" {
		if _, exists := apc.Models[apc.DefaultModel]; !exists {
			return fmt.Errorf("default model '%s' not found in models", apc.DefaultModel)
		}
	}
	
	return nil
}

// Validate validates an AI model configuration
func (amc *AIModelConfig) Validate() error {
	if amc.Name == "" {
		return fmt.Errorf("model name is required")
	}
	
	if amc.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be greater than 0")
	}
	
	if amc.ContextWindow <= 0 {
		return fmt.Errorf("context_window must be greater than 0")
	}
	
	if amc.Temperature < 0 || amc.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}
	
	if amc.TopP < 0 || amc.TopP > 1 {
		return fmt.Errorf("top_p must be between 0 and 1")
	}
	
	return nil
}
