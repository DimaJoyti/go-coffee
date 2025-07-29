package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// RateLimiter provides comprehensive rate limiting and quota management
type RateLimiter struct {
	logger *logger.Logger
	config RateLimiterConfig

	// Rate limiting engines
	tokenBucket    TokenBucketEngine
	slidingWindow  SlidingWindowEngine
	fixedWindow    FixedWindowEngine
	leakyBucket    LeakyBucketEngine
	
	// Quota management
	quotaManager   QuotaManager
	usageTracker   UsageTracker
	
	// Storage backends
	redisBackend   RedisBackend
	memoryBackend  MemoryBackend
	
	// Policy management
	policyEngine   PolicyEngine
	ruleEngine     RuleEngine
	
	// Monitoring and analytics
	metricsCollector MetricsCollector
	alertManager     AlertManager
	
	// State management
	isRunning        bool
	cleanupTicker    *time.Ticker
	stopChan         chan struct{}
	mutex            sync.RWMutex
}

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	Enabled                bool                      `json:"enabled" yaml:"enabled"`
	DefaultAlgorithm       string                    `json:"default_algorithm" yaml:"default_algorithm"` // "token_bucket", "sliding_window", "fixed_window", "leaky_bucket"
	StorageBackend         string                    `json:"storage_backend" yaml:"storage_backend"` // "redis", "memory", "hybrid"
	CleanupInterval        time.Duration             `json:"cleanup_interval" yaml:"cleanup_interval"`
	TokenBucketConfig      TokenBucketConfig         `json:"token_bucket_config" yaml:"token_bucket_config"`
	SlidingWindowConfig    SlidingWindowConfig       `json:"sliding_window_config" yaml:"sliding_window_config"`
	FixedWindowConfig      FixedWindowConfig         `json:"fixed_window_config" yaml:"fixed_window_config"`
	LeakyBucketConfig      LeakyBucketConfig         `json:"leaky_bucket_config" yaml:"leaky_bucket_config"`
	QuotaManagerConfig     QuotaManagerConfig        `json:"quota_manager_config" yaml:"quota_manager_config"`
	UsageTrackerConfig     UsageTrackerConfig        `json:"usage_tracker_config" yaml:"usage_tracker_config"`
	RedisConfig            RedisBackendConfig        `json:"redis_config" yaml:"redis_config"`
	MemoryConfig           MemoryBackendConfig       `json:"memory_config" yaml:"memory_config"`
	PolicyEngineConfig     PolicyEngineConfig        `json:"policy_engine_config" yaml:"policy_engine_config"`
	RuleEngineConfig       RuleEngineConfig          `json:"rule_engine_config" yaml:"rule_engine_config"`
	MetricsConfig          MetricsCollectorConfig    `json:"metrics_config" yaml:"metrics_config"`
	AlertConfig            AlertManagerConfig        `json:"alert_config" yaml:"alert_config"`
}

// Algorithm-specific configurations
type TokenBucketConfig struct {
	Enabled              bool                        `json:"enabled" yaml:"enabled"`
	DefaultCapacity      int64                       `json:"default_capacity" yaml:"default_capacity"`
	DefaultRefillRate    int64                       `json:"default_refill_rate" yaml:"default_refill_rate"`
	RefillInterval       time.Duration               `json:"refill_interval" yaml:"refill_interval"`
	MaxBurstSize         int64                       `json:"max_burst_size" yaml:"max_burst_size"`
	PrecisionMode        bool                        `json:"precision_mode" yaml:"precision_mode"`
}

type SlidingWindowConfig struct {
	Enabled              bool                        `json:"enabled" yaml:"enabled"`
	DefaultWindowSize    time.Duration               `json:"default_window_size" yaml:"default_window_size"`
	DefaultLimit         int64                       `json:"default_limit" yaml:"default_limit"`
	SubWindowCount       int                         `json:"sub_window_count" yaml:"sub_window_count"`
	PrecisionLevel       string                      `json:"precision_level" yaml:"precision_level"` // "second", "millisecond", "microsecond"
}

type FixedWindowConfig struct {
	Enabled              bool                        `json:"enabled" yaml:"enabled"`
	DefaultWindowSize    time.Duration               `json:"default_window_size" yaml:"default_window_size"`
	DefaultLimit         int64                       `json:"default_limit" yaml:"default_limit"`
	WindowAlignment      string                      `json:"window_alignment" yaml:"window_alignment"` // "start", "current"
}

type LeakyBucketConfig struct {
	Enabled              bool                        `json:"enabled" yaml:"enabled"`
	DefaultCapacity      int64                       `json:"default_capacity" yaml:"default_capacity"`
	DefaultLeakRate      int64                       `json:"default_leak_rate" yaml:"default_leak_rate"`
	LeakInterval         time.Duration               `json:"leak_interval" yaml:"leak_interval"`
	OverflowBehavior     string                      `json:"overflow_behavior" yaml:"overflow_behavior"` // "drop", "queue", "reject"
}

type QuotaManagerConfig struct {
	Enabled              bool                        `json:"enabled" yaml:"enabled"`
	QuotaPeriods         []string                    `json:"quota_periods" yaml:"quota_periods"` // "minute", "hour", "day", "week", "month"
	DefaultQuotas        map[string]int64            `json:"default_quotas" yaml:"default_quotas"`
	QuotaResetBehavior   string                      `json:"quota_reset_behavior" yaml:"quota_reset_behavior"` // "rolling", "fixed"
	OverageHandling      string                      `json:"overage_handling" yaml:"overage_handling"` // "block", "throttle", "charge"
	GracePeriod          time.Duration               `json:"grace_period" yaml:"grace_period"`
}

type UsageTrackerConfig struct {
	Enabled              bool                        `json:"enabled" yaml:"enabled"`
	TrackingGranularity  string                      `json:"tracking_granularity" yaml:"tracking_granularity"` // "request", "endpoint", "user", "api_key"
	RetentionPeriod      time.Duration               `json:"retention_period" yaml:"retention_period"`
	AggregationLevels    []string                    `json:"aggregation_levels" yaml:"aggregation_levels"`
	RealTimeUpdates      bool                        `json:"real_time_updates" yaml:"real_time_updates"`
}

type RedisBackendConfig struct {
	Enabled              bool                        `json:"enabled" yaml:"enabled"`
	Address              string                      `json:"address" yaml:"address"`
	Password             string                      `json:"password" yaml:"password"`
	Database             int                         `json:"database" yaml:"database"`
	PoolSize             int                         `json:"pool_size" yaml:"pool_size"`
	KeyPrefix            string                      `json:"key_prefix" yaml:"key_prefix"`
	KeyTTL               time.Duration               `json:"key_ttl" yaml:"key_ttl"`
	ClusterMode          bool                        `json:"cluster_mode" yaml:"cluster_mode"`
	ClusterAddresses     []string                    `json:"cluster_addresses" yaml:"cluster_addresses"`
}

type MemoryBackendConfig struct {
	Enabled              bool                        `json:"enabled" yaml:"enabled"`
	MaxEntries           int                         `json:"max_entries" yaml:"max_entries"`
	CleanupInterval      time.Duration               `json:"cleanup_interval" yaml:"cleanup_interval"`
	EvictionPolicy       string                      `json:"eviction_policy" yaml:"eviction_policy"` // "lru", "lfu", "ttl"
	ShardCount           int                         `json:"shard_count" yaml:"shard_count"`
}

type PolicyEngineConfig struct {
	Enabled              bool                        `json:"enabled" yaml:"enabled"`
	PolicySources        []string                    `json:"policy_sources" yaml:"policy_sources"` // "file", "database", "api"
	PolicyFormat         string                      `json:"policy_format" yaml:"policy_format"` // "yaml", "json", "rego"
	DynamicPolicies      bool                        `json:"dynamic_policies" yaml:"dynamic_policies"`
	PolicyCaching        bool                        `json:"policy_caching" yaml:"policy_caching"`
	PolicyValidation     bool                        `json:"policy_validation" yaml:"policy_validation"`
}

type RuleEngineConfig struct {
	Enabled              bool                        `json:"enabled" yaml:"enabled"`
	RuleTypes            []string                    `json:"rule_types" yaml:"rule_types"` // "conditional", "time_based", "usage_based", "geographic"
	CustomRules          bool                        `json:"custom_rules" yaml:"custom_rules"`
	RuleChaining         bool                        `json:"rule_chaining" yaml:"rule_chaining"`
	RulePriority         bool                        `json:"rule_priority" yaml:"rule_priority"`
}

type MetricsCollectorConfig struct {
	Enabled              bool                        `json:"enabled" yaml:"enabled"`
	MetricTypes          []string                    `json:"metric_types" yaml:"metric_types"`
	CollectionInterval   time.Duration               `json:"collection_interval" yaml:"collection_interval"`
	ExportFormats        []string                    `json:"export_formats" yaml:"export_formats"` // "prometheus", "statsd", "influxdb"
	RetentionPeriod      time.Duration               `json:"retention_period" yaml:"retention_period"`
}

type AlertManagerConfig struct {
	Enabled              bool                        `json:"enabled" yaml:"enabled"`
	AlertThresholds      map[string]decimal.Decimal  `json:"alert_thresholds" yaml:"alert_thresholds"`
	NotificationChannels []string                    `json:"notification_channels" yaml:"notification_channels"`
	AlertCooldown        time.Duration               `json:"alert_cooldown" yaml:"alert_cooldown"`
	EscalationRules      bool                        `json:"escalation_rules" yaml:"escalation_rules"`
}

// Core data structures

// RateLimitRequest represents a rate limit check request
type RateLimitRequest struct {
	Key                  string                      `json:"key"`
	Algorithm            string                      `json:"algorithm"`
	Limit                int64                       `json:"limit"`
	Window               time.Duration               `json:"window"`
	Cost                 int64                       `json:"cost"` // Number of tokens/requests to consume
	Metadata             map[string]interface{}      `json:"metadata"`
	Timestamp            time.Time                   `json:"timestamp"`
}

// RateLimitResponse represents the result of a rate limit check
type RateLimitResponse struct {
	Allowed              bool                        `json:"allowed"`
	Remaining            int64                       `json:"remaining"`
	ResetTime            time.Time                   `json:"reset_time"`
	RetryAfter           time.Duration               `json:"retry_after"`
	TotalLimit           int64                       `json:"total_limit"`
	WindowSize           time.Duration               `json:"window_size"`
	Algorithm            string                      `json:"algorithm"`
	Metadata             map[string]interface{}      `json:"metadata"`
	Timestamp            time.Time                   `json:"timestamp"`
}

// QuotaRequest represents a quota check request
type QuotaRequest struct {
	UserID               string                      `json:"user_id"`
	APIKey               string                      `json:"api_key"`
	Resource             string                      `json:"resource"`
	Operation            string                      `json:"operation"`
	Cost                 int64                       `json:"cost"`
	Period               string                      `json:"period"` // "minute", "hour", "day", "week", "month"
	Metadata             map[string]interface{}      `json:"metadata"`
	Timestamp            time.Time                   `json:"timestamp"`
}

// QuotaResponse represents the result of a quota check
type QuotaResponse struct {
	Allowed              bool                        `json:"allowed"`
	Used                 int64                       `json:"used"`
	Remaining            int64                       `json:"remaining"`
	Total                int64                       `json:"total"`
	Period               string                      `json:"period"`
	ResetTime            time.Time                   `json:"reset_time"`
	OverageAllowed       bool                        `json:"overage_allowed"`
	OverageCost          decimal.Decimal             `json:"overage_cost"`
	Metadata             map[string]interface{}      `json:"metadata"`
	Timestamp            time.Time                   `json:"timestamp"`
}

// Policy represents a rate limiting policy
type Policy struct {
	ID                   string                      `json:"id"`
	Name                 string                      `json:"name"`
	Description          string                      `json:"description"`
	Scope                PolicyScope                 `json:"scope"`
	Rules                []PolicyRule                `json:"rules"`
	Priority             int                         `json:"priority"`
	Enabled              bool                        `json:"enabled"`
	ValidFrom            time.Time                   `json:"valid_from"`
	ValidTo              *time.Time                  `json:"valid_to"`
	CreatedAt            time.Time                   `json:"created_at"`
	UpdatedAt            time.Time                   `json:"updated_at"`
	Metadata             map[string]interface{}      `json:"metadata"`
}

// PolicyScope defines the scope of a policy
type PolicyScope struct {
	Type                 string                      `json:"type"` // "global", "user", "api_key", "endpoint", "ip"
	Values               []string                    `json:"values"`
	Conditions           []ScopeCondition            `json:"conditions"`
}

// ScopeCondition defines conditions for policy scope
type ScopeCondition struct {
	Field                string                      `json:"field"`
	Operator             string                      `json:"operator"` // "eq", "ne", "in", "not_in", "regex", "range"
	Value                interface{}                 `json:"value"`
}

// PolicyRule defines a rate limiting rule within a policy
type PolicyRule struct {
	ID                   string                      `json:"id"`
	Type                 string                      `json:"type"` // "rate_limit", "quota", "throttle", "block"
	Algorithm            string                      `json:"algorithm"`
	Limit                int64                       `json:"limit"`
	Window               time.Duration               `json:"window"`
	BurstLimit           *int64                      `json:"burst_limit"`
	Conditions           []RuleCondition             `json:"conditions"`
	Actions              []RuleAction                `json:"actions"`
	Enabled              bool                        `json:"enabled"`
	Priority             int                         `json:"priority"`
}

// RuleCondition defines conditions for rule activation
type RuleCondition struct {
	Type                 string                      `json:"type"` // "time", "usage", "geographic", "custom"
	Field                string                      `json:"field"`
	Operator             string                      `json:"operator"`
	Value                interface{}                 `json:"value"`
	Metadata             map[string]interface{}      `json:"metadata"`
}

// RuleAction defines actions to take when rule is triggered
type RuleAction struct {
	Type                 string                      `json:"type"` // "block", "throttle", "alert", "log", "redirect"
	Parameters           map[string]interface{}      `json:"parameters"`
	Priority             int                         `json:"priority"`
}

// UsageMetrics represents usage statistics
type UsageMetrics struct {
	Key                  string                      `json:"key"`
	Period               string                      `json:"period"`
	TotalRequests        int64                       `json:"total_requests"`
	AllowedRequests      int64                       `json:"allowed_requests"`
	BlockedRequests      int64                       `json:"blocked_requests"`
	ThrottledRequests    int64                       `json:"throttled_requests"`
	AverageLatency       time.Duration               `json:"average_latency"`
	PeakRPS              int64                       `json:"peak_rps"`
	ErrorRate            decimal.Decimal             `json:"error_rate"`
	QuotaUsage           map[string]int64            `json:"quota_usage"`
	TopEndpoints         []EndpointMetric            `json:"top_endpoints"`
	GeographicDistribution map[string]int64          `json:"geographic_distribution"`
	Timestamp            time.Time                   `json:"timestamp"`
}

// EndpointMetric represents metrics for a specific endpoint
type EndpointMetric struct {
	Endpoint             string                      `json:"endpoint"`
	Method               string                      `json:"method"`
	RequestCount         int64                       `json:"request_count"`
	AverageLatency       time.Duration               `json:"average_latency"`
	ErrorRate            decimal.Decimal             `json:"error_rate"`
	RateLimitHits        int64                       `json:"rate_limit_hits"`
}

// Component interfaces
type TokenBucketEngine interface {
	CheckLimit(ctx context.Context, request *RateLimitRequest) (*RateLimitResponse, error)
	RefillTokens(ctx context.Context, key string) error
	GetBucketState(ctx context.Context, key string) (*BucketState, error)
}

type SlidingWindowEngine interface {
	CheckLimit(ctx context.Context, request *RateLimitRequest) (*RateLimitResponse, error)
	CleanupExpiredWindows(ctx context.Context) error
	GetWindowState(ctx context.Context, key string) (*WindowState, error)
}

type FixedWindowEngine interface {
	CheckLimit(ctx context.Context, request *RateLimitRequest) (*RateLimitResponse, error)
	ResetWindow(ctx context.Context, key string) error
	GetWindowState(ctx context.Context, key string) (*WindowState, error)
}

type LeakyBucketEngine interface {
	CheckLimit(ctx context.Context, request *RateLimitRequest) (*RateLimitResponse, error)
	LeakTokens(ctx context.Context, key string) error
	GetBucketState(ctx context.Context, key string) (*BucketState, error)
}

type QuotaManager interface {
	CheckQuota(ctx context.Context, request *QuotaRequest) (*QuotaResponse, error)
	UpdateQuota(ctx context.Context, userID string, resource string, usage int64) error
	ResetQuota(ctx context.Context, userID string, resource string, period string) error
	GetQuotaUsage(ctx context.Context, userID string, period string) (map[string]*QuotaUsage, error)
}

type UsageTracker interface {
	RecordUsage(ctx context.Context, key string, cost int64, metadata map[string]interface{}) error
	GetUsageMetrics(ctx context.Context, key string, period string) (*UsageMetrics, error)
	GetTopUsers(ctx context.Context, period string, limit int) ([]UserUsage, error)
	ExportMetrics(ctx context.Context, format string, period string) ([]byte, error)
}

type RedisBackend interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Increment(ctx context.Context, key string, delta int64) (int64, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Pipeline(ctx context.Context, commands []RedisCommand) ([]interface{}, error)
}

type MemoryBackend interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Increment(ctx context.Context, key string, delta int64) (int64, error)
	Delete(ctx context.Context, key string) error
	Cleanup(ctx context.Context) error
	GetStats() *MemoryStats
}

type PolicyEngine interface {
	LoadPolicies(ctx context.Context) error
	GetPolicy(ctx context.Context, scope PolicyScope) (*Policy, error)
	EvaluatePolicy(ctx context.Context, request interface{}) (*PolicyDecision, error)
	UpdatePolicy(ctx context.Context, policy *Policy) error
	DeletePolicy(ctx context.Context, policyID string) error
}

type RuleEngine interface {
	EvaluateRules(ctx context.Context, rules []PolicyRule, context map[string]interface{}) (*RuleDecision, error)
	AddRule(ctx context.Context, rule *PolicyRule) error
	UpdateRule(ctx context.Context, rule *PolicyRule) error
	DeleteRule(ctx context.Context, ruleID string) error
}

type MetricsCollector interface {
	RecordMetric(ctx context.Context, metric *Metric) error
	GetMetrics(ctx context.Context, query *MetricQuery) ([]*Metric, error)
	ExportMetrics(ctx context.Context, format string) ([]byte, error)
	StartCollection(ctx context.Context) error
	StopCollection(ctx context.Context) error
}

type AlertManager interface {
	CheckThresholds(ctx context.Context, metrics *UsageMetrics) error
	SendAlert(ctx context.Context, alert *Alert) error
	GetActiveAlerts(ctx context.Context) ([]*Alert, error)
	AcknowledgeAlert(ctx context.Context, alertID string) error
}

// Supporting types
type BucketState struct {
	Tokens               int64                       `json:"tokens"`
	Capacity             int64                       `json:"capacity"`
	RefillRate           int64                       `json:"refill_rate"`
	LastRefill           time.Time                   `json:"last_refill"`
}

type WindowState struct {
	Count                int64                       `json:"count"`
	WindowStart          time.Time                   `json:"window_start"`
	WindowEnd            time.Time                   `json:"window_end"`
	SubWindows           []SubWindow                 `json:"sub_windows"`
}

type SubWindow struct {
	Start                time.Time                   `json:"start"`
	End                  time.Time                   `json:"end"`
	Count                int64                       `json:"count"`
}

type QuotaUsage struct {
	Resource             string                      `json:"resource"`
	Used                 int64                       `json:"used"`
	Total                int64                       `json:"total"`
	Period               string                      `json:"period"`
	ResetTime            time.Time                   `json:"reset_time"`
}

type UserUsage struct {
	UserID               string                      `json:"user_id"`
	TotalRequests        int64                       `json:"total_requests"`
	QuotaUsage           map[string]int64            `json:"quota_usage"`
	LastActivity         time.Time                   `json:"last_activity"`
}

type RedisCommand struct {
	Command              string                      `json:"command"`
	Args                 []interface{}               `json:"args"`
}

type MemoryStats struct {
	EntryCount           int                         `json:"entry_count"`
	MemoryUsage          int64                       `json:"memory_usage"`
	HitRate              decimal.Decimal             `json:"hit_rate"`
	EvictionCount        int64                       `json:"eviction_count"`
}

type PolicyDecision struct {
	Allowed              bool                        `json:"allowed"`
	Policy               *Policy                     `json:"policy"`
	AppliedRules         []PolicyRule                `json:"applied_rules"`
	Actions              []RuleAction                `json:"actions"`
	Reason               string                      `json:"reason"`
}

type RuleDecision struct {
	Allowed              bool                        `json:"allowed"`
	MatchedRules         []PolicyRule                `json:"matched_rules"`
	Actions              []RuleAction                `json:"actions"`
	Score                decimal.Decimal             `json:"score"`
}

type Metric struct {
	Name                 string                      `json:"name"`
	Value                decimal.Decimal             `json:"value"`
	Tags                 map[string]string           `json:"tags"`
	Timestamp            time.Time                   `json:"timestamp"`
	Type                 string                      `json:"type"` // "counter", "gauge", "histogram", "summary"
}

type MetricQuery struct {
	Name                 string                      `json:"name"`
	Tags                 map[string]string           `json:"tags"`
	StartTime            time.Time                   `json:"start_time"`
	EndTime              time.Time                   `json:"end_time"`
	Aggregation          string                      `json:"aggregation"` // "sum", "avg", "min", "max", "count"
}

type Alert struct {
	ID                   string                      `json:"id"`
	Type                 string                      `json:"type"`
	Severity             string                      `json:"severity"` // "low", "medium", "high", "critical"
	Message              string                      `json:"message"`
	Threshold            decimal.Decimal             `json:"threshold"`
	CurrentValue         decimal.Decimal             `json:"current_value"`
	Metadata             map[string]interface{}      `json:"metadata"`
	CreatedAt            time.Time                   `json:"created_at"`
	AcknowledgedAt       *time.Time                  `json:"acknowledged_at"`
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(logger *logger.Logger, config RateLimiterConfig) *RateLimiter {
	rl := &RateLimiter{
		logger:   logger.Named("rate-limiter"),
		config:   config,
		stopChan: make(chan struct{}),
	}

	// Initialize components (mock implementations for this example)
	rl.initializeComponents()

	return rl
}

// initializeComponents initializes all rate limiter components
func (rl *RateLimiter) initializeComponents() {
	// Initialize components with mock implementations
	// In production, these would be real implementations
	rl.tokenBucket = &MockTokenBucketEngine{}
	rl.slidingWindow = &MockSlidingWindowEngine{}
	rl.fixedWindow = &MockFixedWindowEngine{}
	rl.leakyBucket = &MockLeakyBucketEngine{}
	rl.quotaManager = &MockQuotaManager{}
	rl.usageTracker = &MockUsageTracker{}
	rl.redisBackend = &MockRedisBackend{}
	rl.memoryBackend = &MockMemoryBackend{}
	rl.policyEngine = &MockPolicyEngine{}
	rl.ruleEngine = &MockRuleEngine{}
	rl.metricsCollector = &MockMetricsCollector{}
	rl.alertManager = &MockAlertManager{}
}

// Start starts the rate limiter
func (rl *RateLimiter) Start(ctx context.Context) error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	if rl.isRunning {
		return fmt.Errorf("rate limiter is already running")
	}

	if !rl.config.Enabled {
		rl.logger.Info("Rate limiter is disabled")
		return nil
	}

	rl.logger.Info("Starting rate limiter",
		zap.String("default_algorithm", rl.config.DefaultAlgorithm),
		zap.String("storage_backend", rl.config.StorageBackend))

	// Load policies
	if err := rl.policyEngine.LoadPolicies(ctx); err != nil {
		rl.logger.Warn("Failed to load policies", zap.Error(err))
	}

	// Start metrics collection
	if rl.config.MetricsConfig.Enabled {
		if err := rl.metricsCollector.StartCollection(ctx); err != nil {
			rl.logger.Warn("Failed to start metrics collection", zap.Error(err))
		}
	}

	// Start cleanup routine
	rl.cleanupTicker = time.NewTicker(rl.config.CleanupInterval)
	go rl.cleanupLoop(ctx)

	rl.isRunning = true
	rl.logger.Info("Rate limiter started successfully")
	return nil
}

// Stop stops the rate limiter
func (rl *RateLimiter) Stop() error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	if !rl.isRunning {
		return nil
	}

	rl.logger.Info("Stopping rate limiter")

	// Stop cleanup routine
	if rl.cleanupTicker != nil {
		rl.cleanupTicker.Stop()
	}
	close(rl.stopChan)

	// Stop metrics collection
	if rl.config.MetricsConfig.Enabled {
		if err := rl.metricsCollector.StopCollection(context.Background()); err != nil {
			rl.logger.Warn("Failed to stop metrics collection", zap.Error(err))
		}
	}

	rl.isRunning = false
	rl.logger.Info("Rate limiter stopped")
	return nil
}

// CheckRateLimit checks if a request is allowed based on rate limiting rules
func (rl *RateLimiter) CheckRateLimit(ctx context.Context, request *RateLimitRequest) (*RateLimitResponse, error) {
	startTime := time.Now()

	rl.logger.Debug("Checking rate limit",
		zap.String("key", request.Key),
		zap.String("algorithm", request.Algorithm),
		zap.Int64("limit", request.Limit),
		zap.Int64("cost", request.Cost))

	// Set default algorithm if not specified
	if request.Algorithm == "" {
		request.Algorithm = rl.config.DefaultAlgorithm
	}

	// Set default timestamp if not specified
	if request.Timestamp.IsZero() {
		request.Timestamp = time.Now()
	}

	// Check policy first
	policyDecision, err := rl.evaluatePolicy(ctx, request)
	if err != nil {
		rl.logger.Warn("Failed to evaluate policy", zap.Error(err))
	} else if !policyDecision.Allowed {
		return &RateLimitResponse{
			Allowed:    false,
			Remaining:  0,
			ResetTime:  time.Now().Add(request.Window),
			RetryAfter: request.Window,
			TotalLimit: request.Limit,
			WindowSize: request.Window,
			Algorithm:  request.Algorithm,
			Metadata:   map[string]interface{}{"policy_blocked": true, "reason": policyDecision.Reason},
			Timestamp:  time.Now(),
		}, nil
	}

	// Apply rate limiting algorithm
	var response *RateLimitResponse
	switch request.Algorithm {
	case "token_bucket":
		response, err = rl.tokenBucket.CheckLimit(ctx, request)
	case "sliding_window":
		response, err = rl.slidingWindow.CheckLimit(ctx, request)
	case "fixed_window":
		response, err = rl.fixedWindow.CheckLimit(ctx, request)
	case "leaky_bucket":
		response, err = rl.leakyBucket.CheckLimit(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", request.Algorithm)
	}

	if err != nil {
		return nil, fmt.Errorf("rate limit check failed: %w", err)
	}

	// Record usage metrics
	if rl.config.UsageTrackerConfig.Enabled {
		usageMetadata := map[string]interface{}{
			"algorithm": request.Algorithm,
			"allowed":   response.Allowed,
			"latency":   time.Since(startTime),
		}
		for k, v := range request.Metadata {
			usageMetadata[k] = v
		}

		if err := rl.usageTracker.RecordUsage(ctx, request.Key, request.Cost, usageMetadata); err != nil {
			rl.logger.Warn("Failed to record usage", zap.Error(err))
		}
	}

	// Record metrics
	if rl.config.MetricsConfig.Enabled {
		metric := &Metric{
			Name:  "rate_limit_check",
			Value: decimal.NewFromFloat(1),
			Tags: map[string]string{
				"algorithm": request.Algorithm,
				"allowed":   fmt.Sprintf("%t", response.Allowed),
				"key":       request.Key,
			},
			Timestamp: time.Now(),
			Type:      "counter",
		}
		if err := rl.metricsCollector.RecordMetric(ctx, metric); err != nil {
			rl.logger.Warn("Failed to record metric", zap.Error(err))
		}
	}

	rl.logger.Debug("Rate limit check completed",
		zap.String("key", request.Key),
		zap.Bool("allowed", response.Allowed),
		zap.Int64("remaining", response.Remaining),
		zap.Duration("latency", time.Since(startTime)))

	return response, nil
}

// CheckQuota checks if a request is allowed based on quota limits
func (rl *RateLimiter) CheckQuota(ctx context.Context, request *QuotaRequest) (*QuotaResponse, error) {
	startTime := time.Now()

	rl.logger.Debug("Checking quota",
		zap.String("user_id", request.UserID),
		zap.String("resource", request.Resource),
		zap.String("operation", request.Operation),
		zap.Int64("cost", request.Cost))

	// Set default timestamp if not specified
	if request.Timestamp.IsZero() {
		request.Timestamp = time.Now()
	}

	// Check quota
	response, err := rl.quotaManager.CheckQuota(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("quota check failed: %w", err)
	}

	// Record usage metrics
	if rl.config.UsageTrackerConfig.Enabled {
		usageMetadata := map[string]interface{}{
			"user_id":   request.UserID,
			"resource":  request.Resource,
			"operation": request.Operation,
			"allowed":   response.Allowed,
			"latency":   time.Since(startTime),
		}
		for k, v := range request.Metadata {
			usageMetadata[k] = v
		}

		key := fmt.Sprintf("quota:%s:%s", request.UserID, request.Resource)
		if err := rl.usageTracker.RecordUsage(ctx, key, request.Cost, usageMetadata); err != nil {
			rl.logger.Warn("Failed to record quota usage", zap.Error(err))
		}
	}

	// Record metrics
	if rl.config.MetricsConfig.Enabled {
		metric := &Metric{
			Name:  "quota_check",
			Value: decimal.NewFromFloat(1),
			Tags: map[string]string{
				"user_id":   request.UserID,
				"resource":  request.Resource,
				"operation": request.Operation,
				"allowed":   fmt.Sprintf("%t", response.Allowed),
			},
			Timestamp: time.Now(),
			Type:      "counter",
		}
		if err := rl.metricsCollector.RecordMetric(ctx, metric); err != nil {
			rl.logger.Warn("Failed to record quota metric", zap.Error(err))
		}
	}

	rl.logger.Debug("Quota check completed",
		zap.String("user_id", request.UserID),
		zap.String("resource", request.Resource),
		zap.Bool("allowed", response.Allowed),
		zap.Int64("remaining", response.Remaining),
		zap.Duration("latency", time.Since(startTime)))

	return response, nil
}

// GetUsageMetrics retrieves usage metrics for a key
func (rl *RateLimiter) GetUsageMetrics(ctx context.Context, key string, period string) (*UsageMetrics, error) {
	rl.logger.Debug("Getting usage metrics",
		zap.String("key", key),
		zap.String("period", period))

	metrics, err := rl.usageTracker.GetUsageMetrics(ctx, key, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage metrics: %w", err)
	}

	return metrics, nil
}

// GetQuotaUsage retrieves quota usage for a user
func (rl *RateLimiter) GetQuotaUsage(ctx context.Context, userID string, period string) (map[string]*QuotaUsage, error) {
	rl.logger.Debug("Getting quota usage",
		zap.String("user_id", userID),
		zap.String("period", period))

	usage, err := rl.quotaManager.GetQuotaUsage(ctx, userID, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get quota usage: %w", err)
	}

	return usage, nil
}

// UpdatePolicy updates a rate limiting policy
func (rl *RateLimiter) UpdatePolicy(ctx context.Context, policy *Policy) error {
	rl.logger.Debug("Updating policy",
		zap.String("policy_id", policy.ID),
		zap.String("policy_name", policy.Name))

	if err := rl.policyEngine.UpdatePolicy(ctx, policy); err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	rl.logger.Info("Policy updated successfully",
		zap.String("policy_id", policy.ID),
		zap.String("policy_name", policy.Name))

	return nil
}

// GetTopUsers retrieves top users by usage
func (rl *RateLimiter) GetTopUsers(ctx context.Context, period string, limit int) ([]UserUsage, error) {
	rl.logger.Debug("Getting top users",
		zap.String("period", period),
		zap.Int("limit", limit))

	users, err := rl.usageTracker.GetTopUsers(ctx, period, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top users: %w", err)
	}

	return users, nil
}

// ExportMetrics exports metrics in the specified format
func (rl *RateLimiter) ExportMetrics(ctx context.Context, format string, period string) ([]byte, error) {
	rl.logger.Debug("Exporting metrics",
		zap.String("format", format),
		zap.String("period", period))

	data, err := rl.usageTracker.ExportMetrics(ctx, format, period)
	if err != nil {
		return nil, fmt.Errorf("failed to export metrics: %w", err)
	}

	return data, nil
}

// Helper methods

func (rl *RateLimiter) evaluatePolicy(ctx context.Context, request *RateLimitRequest) (*PolicyDecision, error) {
	if !rl.config.PolicyEngineConfig.Enabled {
		return &PolicyDecision{Allowed: true}, nil
	}

	// Create policy scope based on request
	scope := PolicyScope{
		Type:   "key",
		Values: []string{request.Key},
	}

	// Get applicable policy
	policy, err := rl.policyEngine.GetPolicy(ctx, scope)
	if err != nil {
		return nil, err
	}

	if policy == nil {
		return &PolicyDecision{Allowed: true}, nil
	}

	// Evaluate policy
	decision, err := rl.policyEngine.EvaluatePolicy(ctx, request)
	if err != nil {
		return nil, err
	}

	return decision, nil
}

func (rl *RateLimiter) cleanupLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-rl.stopChan:
			return
		case <-rl.cleanupTicker.C:
			if err := rl.performCleanup(ctx); err != nil {
				rl.logger.Error("Error during cleanup", zap.Error(err))
			}
		}
	}
}

func (rl *RateLimiter) performCleanup(ctx context.Context) error {
	rl.logger.Debug("Performing cleanup")

	// Cleanup sliding windows
	if rl.config.SlidingWindowConfig.Enabled {
		if err := rl.slidingWindow.CleanupExpiredWindows(ctx); err != nil {
			rl.logger.Warn("Failed to cleanup sliding windows", zap.Error(err))
		}
	}

	// Cleanup memory backend
	if rl.config.MemoryConfig.Enabled {
		if err := rl.memoryBackend.Cleanup(ctx); err != nil {
			rl.logger.Warn("Failed to cleanup memory backend", zap.Error(err))
		}
	}

	// Check alert thresholds
	if rl.config.AlertConfig.Enabled {
		// Get overall usage metrics
		metrics, err := rl.usageTracker.GetUsageMetrics(ctx, "global", "hour")
		if err != nil {
			rl.logger.Warn("Failed to get metrics for alerting", zap.Error(err))
		} else {
			if err := rl.alertManager.CheckThresholds(ctx, metrics); err != nil {
				rl.logger.Warn("Failed to check alert thresholds", zap.Error(err))
			}
		}
	}

	return nil
}

// IsRunning returns whether the rate limiter is running
func (rl *RateLimiter) IsRunning() bool {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()
	return rl.isRunning
}

// GetMetrics returns rate limiter metrics
func (rl *RateLimiter) GetMetrics() map[string]interface{} {
	metrics := map[string]interface{}{
		"is_running":         rl.IsRunning(),
		"default_algorithm":  rl.config.DefaultAlgorithm,
		"storage_backend":    rl.config.StorageBackend,
		"cleanup_interval":   rl.config.CleanupInterval.String(),
		"token_bucket_enabled": rl.config.TokenBucketConfig.Enabled,
		"sliding_window_enabled": rl.config.SlidingWindowConfig.Enabled,
		"fixed_window_enabled": rl.config.FixedWindowConfig.Enabled,
		"leaky_bucket_enabled": rl.config.LeakyBucketConfig.Enabled,
		"quota_manager_enabled": rl.config.QuotaManagerConfig.Enabled,
		"usage_tracker_enabled": rl.config.UsageTrackerConfig.Enabled,
		"policy_engine_enabled": rl.config.PolicyEngineConfig.Enabled,
		"metrics_enabled":    rl.config.MetricsConfig.Enabled,
		"alerts_enabled":     rl.config.AlertConfig.Enabled,
	}

	// Add memory backend stats if available
	if rl.config.MemoryConfig.Enabled {
		if stats := rl.memoryBackend.GetStats(); stats != nil {
			metrics["memory_stats"] = stats
		}
	}

	return metrics
}
