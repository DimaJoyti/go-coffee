package ratelimit

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// GetDefaultRateLimiterConfig returns default rate limiter configuration
func GetDefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		Enabled:          true,
		DefaultAlgorithm: "token_bucket",
		StorageBackend:   "memory",
		CleanupInterval:  5 * time.Minute,
		TokenBucketConfig: TokenBucketConfig{
			Enabled:           true,
			DefaultCapacity:   1000,
			DefaultRefillRate: 100,
			RefillInterval:    time.Second,
			MaxBurstSize:      2000,
			PrecisionMode:     false,
		},
		SlidingWindowConfig: SlidingWindowConfig{
			Enabled:           true,
			DefaultWindowSize: time.Hour,
			DefaultLimit:      1000,
			SubWindowCount:    10,
			PrecisionLevel:    "second",
		},
		FixedWindowConfig: FixedWindowConfig{
			Enabled:           true,
			DefaultWindowSize: time.Hour,
			DefaultLimit:      1000,
			WindowAlignment:   "start",
		},
		LeakyBucketConfig: LeakyBucketConfig{
			Enabled:          true,
			DefaultCapacity:  1000,
			DefaultLeakRate:  100,
			LeakInterval:     time.Second,
			OverflowBehavior: "drop",
		},
		QuotaManagerConfig: QuotaManagerConfig{
			Enabled:        true,
			QuotaPeriods:   []string{"minute", "hour", "day", "week", "month"},
			DefaultQuotas: map[string]int64{
				"api_calls":     10000,
				"data_transfer": 100000000, // 100MB
				"storage":       1000000000, // 1GB
			},
			QuotaResetBehavior: "rolling",
			OverageHandling:    "block",
			GracePeriod:        5 * time.Minute,
		},
		UsageTrackerConfig: UsageTrackerConfig{
			Enabled:             true,
			TrackingGranularity: "request",
			RetentionPeriod:     30 * 24 * time.Hour, // 30 days
			AggregationLevels:   []string{"minute", "hour", "day"},
			RealTimeUpdates:     true,
		},
		RedisConfig: RedisBackendConfig{
			Enabled:          false,
			Address:          "localhost:6379",
			Password:         "",
			Database:         0,
			PoolSize:         10,
			KeyPrefix:        "ratelimit:",
			KeyTTL:           24 * time.Hour,
			ClusterMode:      false,
			ClusterAddresses: []string{},
		},
		MemoryConfig: MemoryBackendConfig{
			Enabled:         true,
			MaxEntries:      10000,
			CleanupInterval: 5 * time.Minute,
			EvictionPolicy:  "lru",
			ShardCount:      16,
		},
		PolicyEngineConfig: PolicyEngineConfig{
			Enabled:          true,
			PolicySources:    []string{"file"},
			PolicyFormat:     "yaml",
			DynamicPolicies:  true,
			PolicyCaching:    true,
			PolicyValidation: true,
		},
		RuleEngineConfig: RuleEngineConfig{
			Enabled:      true,
			RuleTypes:    []string{"conditional", "time_based", "usage_based"},
			CustomRules:  true,
			RuleChaining: true,
			RulePriority: true,
		},
		MetricsConfig: MetricsCollectorConfig{
			Enabled:            true,
			MetricTypes:        []string{"counter", "gauge", "histogram"},
			CollectionInterval: 30 * time.Second,
			ExportFormats:      []string{"prometheus", "json"},
			RetentionPeriod:    7 * 24 * time.Hour, // 7 days
		},
		AlertConfig: AlertManagerConfig{
			Enabled: true,
			AlertThresholds: map[string]decimal.Decimal{
				"high_usage":    decimal.NewFromFloat(0.8),  // 80% of limit
				"error_rate":    decimal.NewFromFloat(0.05), // 5% error rate
				"latency":       decimal.NewFromFloat(1000), // 1000ms
				"quota_usage":   decimal.NewFromFloat(0.9),  // 90% of quota
			},
			NotificationChannels: []string{"email", "slack"},
			AlertCooldown:        15 * time.Minute,
			EscalationRules:      true,
		},
	}
}

// GetHighThroughputConfig returns configuration optimized for high throughput
func GetHighThroughputConfig() RateLimiterConfig {
	config := GetDefaultRateLimiterConfig()
	
	// Optimize for high throughput
	config.DefaultAlgorithm = "sliding_window"
	config.StorageBackend = "redis"
	config.CleanupInterval = 1 * time.Minute
	
	// Higher limits
	config.TokenBucketConfig.DefaultCapacity = 10000
	config.TokenBucketConfig.DefaultRefillRate = 1000
	config.TokenBucketConfig.MaxBurstSize = 20000
	
	config.SlidingWindowConfig.DefaultLimit = 10000
	config.SlidingWindowConfig.SubWindowCount = 20
	config.SlidingWindowConfig.PrecisionLevel = "millisecond"
	
	config.FixedWindowConfig.DefaultLimit = 10000
	
	config.LeakyBucketConfig.DefaultCapacity = 10000
	config.LeakyBucketConfig.DefaultLeakRate = 1000
	
	// Higher quotas
	config.QuotaManagerConfig.DefaultQuotas = map[string]int64{
		"api_calls":     100000,
		"data_transfer": 1000000000, // 1GB
		"storage":       10000000000, // 10GB
	}
	
	// Redis backend for better performance
	config.RedisConfig.Enabled = true
	config.RedisConfig.PoolSize = 50
	config.MemoryConfig.Enabled = false
	
	// More frequent metrics collection
	config.MetricsConfig.CollectionInterval = 10 * time.Second
	
	return config
}

// GetStrictSecurityConfig returns configuration with strict security settings
func GetStrictSecurityConfig() RateLimiterConfig {
	config := GetDefaultRateLimiterConfig()
	
	// Strict security settings
	config.DefaultAlgorithm = "leaky_bucket"
	config.CleanupInterval = 30 * time.Second
	
	// Lower limits
	config.TokenBucketConfig.DefaultCapacity = 100
	config.TokenBucketConfig.DefaultRefillRate = 10
	config.TokenBucketConfig.MaxBurstSize = 200
	
	config.SlidingWindowConfig.DefaultLimit = 100
	config.SlidingWindowConfig.DefaultWindowSize = 15 * time.Minute
	
	config.FixedWindowConfig.DefaultLimit = 100
	config.FixedWindowConfig.DefaultWindowSize = 15 * time.Minute
	
	config.LeakyBucketConfig.DefaultCapacity = 100
	config.LeakyBucketConfig.DefaultLeakRate = 10
	config.LeakyBucketConfig.OverflowBehavior = "reject"
	
	// Stricter quotas
	config.QuotaManagerConfig.DefaultQuotas = map[string]int64{
		"api_calls":     1000,
		"data_transfer": 10000000, // 10MB
		"storage":       100000000, // 100MB
	}
	config.QuotaManagerConfig.OverageHandling = "block"
	config.QuotaManagerConfig.GracePeriod = 0
	
	// Enhanced monitoring
	config.AlertConfig.AlertThresholds = map[string]decimal.Decimal{
		"high_usage":  decimal.NewFromFloat(0.7),  // 70% of limit
		"error_rate":  decimal.NewFromFloat(0.02), // 2% error rate
		"latency":     decimal.NewFromFloat(500),  // 500ms
		"quota_usage": decimal.NewFromFloat(0.8),  // 80% of quota
	}
	config.AlertConfig.AlertCooldown = 5 * time.Minute
	
	return config
}

// GetDevelopmentConfig returns configuration suitable for development
func GetDevelopmentConfig() RateLimiterConfig {
	config := GetDefaultRateLimiterConfig()
	
	// Development-friendly settings
	config.CleanupInterval = 10 * time.Second
	
	// Relaxed limits for development
	config.TokenBucketConfig.DefaultCapacity = 10000
	config.TokenBucketConfig.DefaultRefillRate = 1000
	
	config.SlidingWindowConfig.DefaultLimit = 10000
	config.FixedWindowConfig.DefaultLimit = 10000
	config.LeakyBucketConfig.DefaultCapacity = 10000
	config.LeakyBucketConfig.DefaultLeakRate = 1000
	
	// Higher quotas for development
	config.QuotaManagerConfig.DefaultQuotas = map[string]int64{
		"api_calls":     100000,
		"data_transfer": 1000000000, // 1GB
		"storage":       10000000000, // 10GB
	}
	config.QuotaManagerConfig.OverageHandling = "throttle"
	
	// Memory backend for simplicity
	config.MemoryConfig.MaxEntries = 1000
	config.RedisConfig.Enabled = false
	
	// Frequent metrics for debugging
	config.MetricsConfig.CollectionInterval = 5 * time.Second
	config.MetricsConfig.RetentionPeriod = 24 * time.Hour
	
	// Lower alert thresholds for testing
	config.AlertConfig.AlertThresholds = map[string]decimal.Decimal{
		"high_usage":  decimal.NewFromFloat(0.9),  // 90% of limit
		"error_rate":  decimal.NewFromFloat(0.1),  // 10% error rate
		"latency":     decimal.NewFromFloat(2000), // 2000ms
		"quota_usage": decimal.NewFromFloat(0.95), // 95% of quota
	}
	
	return config
}

// ValidateRateLimiterConfig validates rate limiter configuration
func ValidateRateLimiterConfig(config RateLimiterConfig) error {
	if !config.Enabled {
		return nil
	}
	
	// Validate default algorithm
	supportedAlgorithms := GetSupportedAlgorithms()
	isValidAlgorithm := false
	for _, algo := range supportedAlgorithms {
		if config.DefaultAlgorithm == algo {
			isValidAlgorithm = true
			break
		}
	}
	if !isValidAlgorithm {
		return fmt.Errorf("unsupported default algorithm: %s", config.DefaultAlgorithm)
	}
	
	// Validate storage backend
	supportedBackends := GetSupportedStorageBackends()
	isValidBackend := false
	for _, backend := range supportedBackends {
		if config.StorageBackend == backend {
			isValidBackend = true
			break
		}
	}
	if !isValidBackend {
		return fmt.Errorf("unsupported storage backend: %s", config.StorageBackend)
	}
	
	// Validate cleanup interval
	if config.CleanupInterval <= 0 {
		return fmt.Errorf("cleanup interval must be positive")
	}
	
	// Validate token bucket config
	if config.TokenBucketConfig.Enabled {
		if config.TokenBucketConfig.DefaultCapacity <= 0 {
			return fmt.Errorf("token bucket default capacity must be positive")
		}
		if config.TokenBucketConfig.DefaultRefillRate <= 0 {
			return fmt.Errorf("token bucket default refill rate must be positive")
		}
		if config.TokenBucketConfig.RefillInterval <= 0 {
			return fmt.Errorf("token bucket refill interval must be positive")
		}
	}
	
	// Validate sliding window config
	if config.SlidingWindowConfig.Enabled {
		if config.SlidingWindowConfig.DefaultWindowSize <= 0 {
			return fmt.Errorf("sliding window default window size must be positive")
		}
		if config.SlidingWindowConfig.DefaultLimit <= 0 {
			return fmt.Errorf("sliding window default limit must be positive")
		}
		if config.SlidingWindowConfig.SubWindowCount <= 0 {
			return fmt.Errorf("sliding window sub-window count must be positive")
		}
	}
	
	// Validate quota manager config
	if config.QuotaManagerConfig.Enabled {
		if len(config.QuotaManagerConfig.QuotaPeriods) == 0 {
			return fmt.Errorf("at least one quota period must be specified")
		}
		if config.QuotaManagerConfig.GracePeriod < 0 {
			return fmt.Errorf("quota grace period cannot be negative")
		}
	}
	
	// Validate usage tracker config
	if config.UsageTrackerConfig.Enabled {
		if config.UsageTrackerConfig.RetentionPeriod <= 0 {
			return fmt.Errorf("usage tracker retention period must be positive")
		}
	}
	
	// Validate Redis config
	if config.RedisConfig.Enabled {
		if config.RedisConfig.Address == "" {
			return fmt.Errorf("Redis address must be specified when Redis is enabled")
		}
		if config.RedisConfig.PoolSize <= 0 {
			return fmt.Errorf("Redis pool size must be positive")
		}
		if config.RedisConfig.KeyTTL <= 0 {
			return fmt.Errorf("Redis key TTL must be positive")
		}
	}
	
	// Validate memory config
	if config.MemoryConfig.Enabled {
		if config.MemoryConfig.MaxEntries <= 0 {
			return fmt.Errorf("memory backend max entries must be positive")
		}
		if config.MemoryConfig.CleanupInterval <= 0 {
			return fmt.Errorf("memory backend cleanup interval must be positive")
		}
		if config.MemoryConfig.ShardCount <= 0 {
			return fmt.Errorf("memory backend shard count must be positive")
		}
	}
	
	// Validate metrics config
	if config.MetricsConfig.Enabled {
		if config.MetricsConfig.CollectionInterval <= 0 {
			return fmt.Errorf("metrics collection interval must be positive")
		}
		if config.MetricsConfig.RetentionPeriod <= 0 {
			return fmt.Errorf("metrics retention period must be positive")
		}
	}
	
	return nil
}

// GetSupportedAlgorithms returns supported rate limiting algorithms
func GetSupportedAlgorithms() []string {
	return []string{
		"token_bucket", "sliding_window", "fixed_window", "leaky_bucket",
	}
}

// GetSupportedStorageBackends returns supported storage backends
func GetSupportedStorageBackends() []string {
	return []string{
		"memory", "redis", "hybrid",
	}
}

// GetSupportedQuotaPeriods returns supported quota periods
func GetSupportedQuotaPeriods() []string {
	return []string{
		"minute", "hour", "day", "week", "month",
	}
}

// GetSupportedMetricTypes returns supported metric types
func GetSupportedMetricTypes() []string {
	return []string{
		"counter", "gauge", "histogram", "summary",
	}
}

// GetSupportedExportFormats returns supported export formats
func GetSupportedExportFormats() []string {
	return []string{
		"json", "csv", "prometheus", "influxdb", "statsd",
	}
}

// GetOptimalConfigForUseCase returns optimal configuration for specific use cases
func GetOptimalConfigForUseCase(useCase string) (RateLimiterConfig, error) {
	switch useCase {
	case "high_throughput":
		return GetHighThroughputConfig(), nil
	case "strict_security":
		return GetStrictSecurityConfig(), nil
	case "development":
		return GetDevelopmentConfig(), nil
	case "default":
		return GetDefaultRateLimiterConfig(), nil
	default:
		return RateLimiterConfig{}, fmt.Errorf("unsupported use case: %s", useCase)
	}
}

// GetUseCaseDescriptions returns descriptions for use cases
func GetUseCaseDescriptions() map[string]string {
	return map[string]string{
		"high_throughput": "Optimized for high-volume API traffic with Redis backend and relaxed limits",
		"strict_security": "Enhanced security with strict limits, leaky bucket algorithm, and aggressive monitoring",
		"development":     "Development-friendly configuration with relaxed limits and memory backend",
		"default":         "Balanced configuration suitable for most production environments",
	}
}

// GetAlgorithmDescriptions returns descriptions for rate limiting algorithms
func GetAlgorithmDescriptions() map[string]string {
	return map[string]string{
		"token_bucket":   "Allows bursts up to bucket capacity, refills at constant rate",
		"sliding_window": "Maintains precise rate over sliding time window, prevents burst at window boundaries",
		"fixed_window":   "Simple counter reset at fixed intervals, allows bursts at window start",
		"leaky_bucket":   "Smooths out bursts by processing requests at constant rate",
	}
}

// GetDefaultPolicyRules returns default policy rules
func GetDefaultPolicyRules() []PolicyRule {
	return []PolicyRule{
		{
			ID:        "default_api_limit",
			Type:      "rate_limit",
			Algorithm: "token_bucket",
			Limit:     1000,
			Window:    time.Hour,
			Enabled:   true,
			Priority:  1,
			Conditions: []RuleCondition{
				{
					Type:     "endpoint",
					Field:    "path",
					Operator: "starts_with",
					Value:    "/api/",
				},
			},
		},
		{
			ID:        "auth_endpoint_limit",
			Type:      "rate_limit",
			Algorithm: "leaky_bucket",
			Limit:     10,
			Window:    time.Minute,
			Enabled:   true,
			Priority:  2,
			Conditions: []RuleCondition{
				{
					Type:     "endpoint",
					Field:    "path",
					Operator: "eq",
					Value:    "/auth/login",
				},
			},
		},
		{
			ID:        "high_value_operation_limit",
			Type:      "rate_limit",
			Algorithm: "sliding_window",
			Limit:     5,
			Window:    time.Minute,
			Enabled:   true,
			Priority:  3,
			Conditions: []RuleCondition{
				{
					Type:     "endpoint",
					Field:    "path",
					Operator: "in",
					Value:    []string{"/api/transfer", "/api/withdraw", "/api/trade"},
				},
			},
		},
	}
}
