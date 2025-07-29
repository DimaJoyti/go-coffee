package ratelimit

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// Mock implementations for testing and demonstration

// MockTokenBucketEngine provides mock token bucket rate limiting
type MockTokenBucketEngine struct {
	buckets map[string]*BucketState
	mutex   sync.RWMutex
}

func (m *MockTokenBucketEngine) CheckLimit(ctx context.Context, request *RateLimitRequest) (*RateLimitResponse, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.buckets == nil {
		m.buckets = make(map[string]*BucketState)
	}

	bucket, exists := m.buckets[request.Key]
	if !exists {
		refillRate := request.Limit / int64(request.Window.Seconds())
		if refillRate == 0 {
			refillRate = 1 // Ensure minimum refill rate
		}
		bucket = &BucketState{
			Tokens:     request.Limit,
			Capacity:   request.Limit,
			RefillRate: refillRate,
			LastRefill: time.Now(),
		}
		m.buckets[request.Key] = bucket
	}

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(bucket.LastRefill)
	tokensToAdd := int64(elapsed.Seconds()) * bucket.RefillRate
	bucket.Tokens = min(bucket.Capacity, bucket.Tokens+tokensToAdd)
	bucket.LastRefill = now

	// Check if request can be allowed
	allowed := bucket.Tokens >= request.Cost
	if allowed {
		bucket.Tokens -= request.Cost
	}

	var resetTime time.Time
	if bucket.RefillRate > 0 {
		resetTime = now.Add(time.Duration((bucket.Capacity-bucket.Tokens)/bucket.RefillRate) * time.Second)
	} else {
		resetTime = now.Add(time.Minute) // Default reset time if refill rate is 0
	}

	return &RateLimitResponse{
		Allowed:    allowed,
		Remaining:  bucket.Tokens,
		ResetTime:  resetTime,
		RetryAfter: time.Until(resetTime),
		TotalLimit: bucket.Capacity,
		WindowSize: request.Window,
		Algorithm:  "token_bucket",
		Timestamp:  now,
	}, nil
}

func (m *MockTokenBucketEngine) RefillTokens(ctx context.Context, key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if bucket, exists := m.buckets[key]; exists {
		bucket.Tokens = bucket.Capacity
		bucket.LastRefill = time.Now()
	}
	return nil
}

func (m *MockTokenBucketEngine) GetBucketState(ctx context.Context, key string) (*BucketState, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if bucket, exists := m.buckets[key]; exists {
		return bucket, nil
	}
	return nil, fmt.Errorf("bucket not found: %s", key)
}

// MockSlidingWindowEngine provides mock sliding window rate limiting
type MockSlidingWindowEngine struct {
	windows map[string]*WindowState
	mutex   sync.RWMutex
}

func (m *MockSlidingWindowEngine) CheckLimit(ctx context.Context, request *RateLimitRequest) (*RateLimitResponse, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.windows == nil {
		m.windows = make(map[string]*WindowState)
	}

	now := time.Now()
	windowStart := now.Add(-request.Window)

	window, exists := m.windows[request.Key]
	if !exists {
		window = &WindowState{
			Count:       0,
			WindowStart: windowStart,
			WindowEnd:   now,
			SubWindows:  []SubWindow{},
		}
		m.windows[request.Key] = window
	}

	// Clean up old sub-windows
	var validSubWindows []SubWindow
	for _, subWindow := range window.SubWindows {
		if subWindow.End.After(windowStart) {
			validSubWindows = append(validSubWindows, subWindow)
		}
	}
	window.SubWindows = validSubWindows

	// Calculate current count
	currentCount := int64(0)
	for _, subWindow := range window.SubWindows {
		currentCount += subWindow.Count
	}

	// Check if request can be allowed
	allowed := currentCount+request.Cost <= request.Limit
	if allowed {
		// Add new sub-window or update existing one
		subWindowDuration := request.Window / 10 // Divide window into 10 sub-windows
		subWindowStart := now.Truncate(subWindowDuration)
		subWindowEnd := subWindowStart.Add(subWindowDuration)

		found := false
		for i := range window.SubWindows {
			if window.SubWindows[i].Start.Equal(subWindowStart) {
				window.SubWindows[i].Count += request.Cost
				found = true
				break
			}
		}

		if !found {
			window.SubWindows = append(window.SubWindows, SubWindow{
				Start: subWindowStart,
				End:   subWindowEnd,
				Count: request.Cost,
			})
		}

		currentCount += request.Cost
	}

	window.Count = currentCount
	window.WindowStart = windowStart
	window.WindowEnd = now

	return &RateLimitResponse{
		Allowed:    allowed,
		Remaining:  max(0, request.Limit-currentCount),
		ResetTime:  now.Add(request.Window),
		RetryAfter: request.Window,
		TotalLimit: request.Limit,
		WindowSize: request.Window,
		Algorithm:  "sliding_window",
		Timestamp:  now,
	}, nil
}

func (m *MockSlidingWindowEngine) CleanupExpiredWindows(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	for key, window := range m.windows {
		if window.WindowEnd.Add(time.Hour).Before(now) {
			delete(m.windows, key)
		}
	}
	return nil
}

func (m *MockSlidingWindowEngine) GetWindowState(ctx context.Context, key string) (*WindowState, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if window, exists := m.windows[key]; exists {
		return window, nil
	}
	return nil, fmt.Errorf("window not found: %s", key)
}

// MockFixedWindowEngine provides mock fixed window rate limiting
type MockFixedWindowEngine struct {
	windows map[string]*WindowState
	mutex   sync.RWMutex
}

func (m *MockFixedWindowEngine) CheckLimit(ctx context.Context, request *RateLimitRequest) (*RateLimitResponse, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.windows == nil {
		m.windows = make(map[string]*WindowState)
	}

	now := time.Now()
	windowStart := now.Truncate(request.Window)
	windowEnd := windowStart.Add(request.Window)

	window, exists := m.windows[request.Key]
	if !exists || !window.WindowStart.Equal(windowStart) {
		window = &WindowState{
			Count:       0,
			WindowStart: windowStart,
			WindowEnd:   windowEnd,
		}
		m.windows[request.Key] = window
	}

	// Check if request can be allowed
	allowed := window.Count+request.Cost <= request.Limit
	if allowed {
		window.Count += request.Cost
	}

	return &RateLimitResponse{
		Allowed:    allowed,
		Remaining:  max(0, request.Limit-window.Count),
		ResetTime:  windowEnd,
		RetryAfter: time.Until(windowEnd),
		TotalLimit: request.Limit,
		WindowSize: request.Window,
		Algorithm:  "fixed_window",
		Timestamp:  now,
	}, nil
}

func (m *MockFixedWindowEngine) ResetWindow(ctx context.Context, key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if window, exists := m.windows[key]; exists {
		window.Count = 0
		window.WindowStart = time.Now().Truncate(time.Hour)
		window.WindowEnd = window.WindowStart.Add(time.Hour)
	}
	return nil
}

func (m *MockFixedWindowEngine) GetWindowState(ctx context.Context, key string) (*WindowState, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if window, exists := m.windows[key]; exists {
		return window, nil
	}
	return nil, fmt.Errorf("window not found: %s", key)
}

// MockLeakyBucketEngine provides mock leaky bucket rate limiting
type MockLeakyBucketEngine struct {
	buckets map[string]*BucketState
	mutex   sync.RWMutex
}

func (m *MockLeakyBucketEngine) CheckLimit(ctx context.Context, request *RateLimitRequest) (*RateLimitResponse, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.buckets == nil {
		m.buckets = make(map[string]*BucketState)
	}

	bucket, exists := m.buckets[request.Key]
	if !exists {
		refillRate := request.Limit / int64(request.Window.Seconds())
		if refillRate == 0 {
			refillRate = 1 // Ensure minimum refill rate
		}
		bucket = &BucketState{
			Tokens:     0,
			Capacity:   request.Limit,
			RefillRate: refillRate,
			LastRefill: time.Now(),
		}
		m.buckets[request.Key] = bucket
	}

	// Leak tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(bucket.LastRefill)
	tokensToLeak := int64(elapsed.Seconds()) * bucket.RefillRate
	bucket.Tokens = max(0, bucket.Tokens-tokensToLeak)
	bucket.LastRefill = now

	// Check if request can be allowed
	allowed := bucket.Tokens+request.Cost <= bucket.Capacity
	if allowed {
		bucket.Tokens += request.Cost
	}

	var resetTime time.Time
	var retryAfter time.Duration
	if bucket.RefillRate > 0 {
		resetTime = now.Add(time.Duration(bucket.Tokens/bucket.RefillRate) * time.Second)
		retryAfter = time.Duration(max(0, bucket.Tokens+request.Cost-bucket.Capacity)/bucket.RefillRate) * time.Second
	} else {
		resetTime = now.Add(time.Minute)
		retryAfter = time.Minute
	}

	return &RateLimitResponse{
		Allowed:    allowed,
		Remaining:  max(0, bucket.Capacity-bucket.Tokens),
		ResetTime:  resetTime,
		RetryAfter: retryAfter,
		TotalLimit: bucket.Capacity,
		WindowSize: request.Window,
		Algorithm:  "leaky_bucket",
		Timestamp:  now,
	}, nil
}

func (m *MockLeakyBucketEngine) LeakTokens(ctx context.Context, key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if bucket, exists := m.buckets[key]; exists {
		bucket.Tokens = 0
		bucket.LastRefill = time.Now()
	}
	return nil
}

func (m *MockLeakyBucketEngine) GetBucketState(ctx context.Context, key string) (*BucketState, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if bucket, exists := m.buckets[key]; exists {
		return bucket, nil
	}
	return nil, fmt.Errorf("bucket not found: %s", key)
}

// MockQuotaManager provides mock quota management
type MockQuotaManager struct {
	quotas map[string]*QuotaUsage
	mutex  sync.RWMutex
}

func (m *MockQuotaManager) CheckQuota(ctx context.Context, request *QuotaRequest) (*QuotaResponse, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.quotas == nil {
		m.quotas = make(map[string]*QuotaUsage)
	}

	key := fmt.Sprintf("%s:%s:%s", request.UserID, request.Resource, request.Period)
	quota, exists := m.quotas[key]

	// Default quota limits
	defaultLimits := map[string]int64{
		"api_calls":     1000,
		"data_transfer": 1000000,  // 1MB
		"storage":       10000000, // 10MB
	}

	totalLimit := defaultLimits[request.Resource]
	if totalLimit == 0 {
		totalLimit = 100 // Default fallback
	}

	if !exists {
		quota = &QuotaUsage{
			Resource:  request.Resource,
			Used:      0,
			Total:     totalLimit,
			Period:    request.Period,
			ResetTime: getNextResetTime(request.Period),
		}
		m.quotas[key] = quota
	}

	// Check if quota period has reset
	if time.Now().After(quota.ResetTime) {
		quota.Used = 0
		quota.ResetTime = getNextResetTime(request.Period)
	}

	// Check if request can be allowed
	allowed := quota.Used+request.Cost <= quota.Total
	if allowed {
		quota.Used += request.Cost
	}

	return &QuotaResponse{
		Allowed:        allowed,
		Used:           quota.Used,
		Remaining:      max(0, quota.Total-quota.Used),
		Total:          quota.Total,
		Period:         quota.Period,
		ResetTime:      quota.ResetTime,
		OverageAllowed: false,
		OverageCost:    decimal.Zero,
		Timestamp:      time.Now(),
	}, nil
}

func (m *MockQuotaManager) UpdateQuota(ctx context.Context, userID string, resource string, usage int64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := fmt.Sprintf("%s:%s:hour", userID, resource)
	if quota, exists := m.quotas[key]; exists {
		quota.Used += usage
	}
	return nil
}

func (m *MockQuotaManager) ResetQuota(ctx context.Context, userID string, resource string, period string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := fmt.Sprintf("%s:%s:%s", userID, resource, period)
	if quota, exists := m.quotas[key]; exists {
		quota.Used = 0
		quota.ResetTime = getNextResetTime(period)
	}
	return nil
}

func (m *MockQuotaManager) GetQuotaUsage(ctx context.Context, userID string, period string) (map[string]*QuotaUsage, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]*QuotaUsage)
	prefix := fmt.Sprintf("%s:", userID)

	for key, quota := range m.quotas {
		if strings.HasPrefix(key, prefix) && strings.HasSuffix(key, ":"+period) {
			parts := strings.Split(key, ":")
			if len(parts) >= 2 {
				resource := parts[1]
				result[resource] = quota
			}
		}
	}

	return result, nil
}

// Helper functions
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func getNextResetTime(period string) time.Time {
	now := time.Now()
	switch period {
	case "minute":
		return now.Truncate(time.Minute).Add(time.Minute)
	case "hour":
		return now.Truncate(time.Hour).Add(time.Hour)
	case "day":
		return now.Truncate(24 * time.Hour).Add(24 * time.Hour)
	case "week":
		return now.Truncate(7 * 24 * time.Hour).Add(7 * 24 * time.Hour)
	case "month":
		return time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	default:
		return now.Add(time.Hour)
	}
}

// MockUsageTracker provides mock usage tracking
type MockUsageTracker struct {
	usage map[string]*UsageMetrics
	mutex sync.RWMutex
}

func (m *MockUsageTracker) RecordUsage(ctx context.Context, key string, cost int64, metadata map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.usage == nil {
		m.usage = make(map[string]*UsageMetrics)
	}

	metrics, exists := m.usage[key]
	if !exists {
		metrics = &UsageMetrics{
			Key:                    key,
			Period:                 "hour",
			TotalRequests:          0,
			AllowedRequests:        0,
			BlockedRequests:        0,
			ThrottledRequests:      0,
			AverageLatency:         0,
			PeakRPS:                0,
			ErrorRate:              decimal.Zero,
			QuotaUsage:             make(map[string]int64),
			TopEndpoints:           []EndpointMetric{},
			GeographicDistribution: make(map[string]int64),
			Timestamp:              time.Now(),
		}
		m.usage[key] = metrics
	}

	metrics.TotalRequests += cost
	if allowed, ok := metadata["allowed"].(bool); ok && allowed {
		metrics.AllowedRequests += cost
	} else {
		metrics.BlockedRequests += cost
	}

	if latency, ok := metadata["latency"].(time.Duration); ok {
		// Simple moving average
		metrics.AverageLatency = (metrics.AverageLatency + latency) / 2
	}

	metrics.Timestamp = time.Now()
	return nil
}

func (m *MockUsageTracker) GetUsageMetrics(ctx context.Context, key string, period string) (*UsageMetrics, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if metrics, exists := m.usage[key]; exists {
		return metrics, nil
	}

	// Return empty metrics if not found
	return &UsageMetrics{
		Key:                    key,
		Period:                 period,
		TotalRequests:          0,
		AllowedRequests:        0,
		BlockedRequests:        0,
		ThrottledRequests:      0,
		AverageLatency:         0,
		PeakRPS:                0,
		ErrorRate:              decimal.Zero,
		QuotaUsage:             make(map[string]int64),
		TopEndpoints:           []EndpointMetric{},
		GeographicDistribution: make(map[string]int64),
		Timestamp:              time.Now(),
	}, nil
}

func (m *MockUsageTracker) GetTopUsers(ctx context.Context, period string, limit int) ([]UserUsage, error) {
	// Mock top users
	return []UserUsage{
		{
			UserID:        "user_1",
			TotalRequests: 1000,
			QuotaUsage: map[string]int64{
				"api_calls":     800,
				"data_transfer": 500000,
			},
			LastActivity: time.Now().Add(-1 * time.Hour),
		},
		{
			UserID:        "user_2",
			TotalRequests: 750,
			QuotaUsage: map[string]int64{
				"api_calls":     600,
				"data_transfer": 300000,
			},
			LastActivity: time.Now().Add(-2 * time.Hour),
		},
	}, nil
}

func (m *MockUsageTracker) ExportMetrics(ctx context.Context, format string, period string) ([]byte, error) {
	switch format {
	case "json":
		return []byte(`{"metrics": "mock_json_data"}`), nil
	case "csv":
		return []byte("key,requests,allowed,blocked\nmock_key,100,90,10"), nil
	case "prometheus":
		return []byte("# HELP rate_limit_requests_total Total requests\nrate_limit_requests_total 100"), nil
	default:
		return []byte("mock metrics data"), nil
	}
}

// MockRedisBackend provides mock Redis backend
type MockRedisBackend struct {
	data  map[string]interface{}
	mutex sync.RWMutex
}

func (m *MockRedisBackend) Get(ctx context.Context, key string) (interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.data == nil {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	if value, exists := m.data[key]; exists {
		return value, nil
	}
	return nil, fmt.Errorf("key not found: %s", key)
}

func (m *MockRedisBackend) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.data == nil {
		m.data = make(map[string]interface{})
	}

	m.data[key] = value
	return nil
}

func (m *MockRedisBackend) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.data == nil {
		m.data = make(map[string]interface{})
	}

	current, exists := m.data[key]
	if !exists {
		m.data[key] = delta
		return delta, nil
	}

	if val, ok := current.(int64); ok {
		newVal := val + delta
		m.data[key] = newVal
		return newVal, nil
	}

	return 0, fmt.Errorf("value is not an integer")
}

func (m *MockRedisBackend) Delete(ctx context.Context, key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.data != nil {
		delete(m.data, key)
	}
	return nil
}

func (m *MockRedisBackend) Exists(ctx context.Context, key string) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.data == nil {
		return false, nil
	}

	_, exists := m.data[key]
	return exists, nil
}

func (m *MockRedisBackend) Pipeline(ctx context.Context, commands []RedisCommand) ([]interface{}, error) {
	results := make([]interface{}, len(commands))
	for i, cmd := range commands {
		switch cmd.Command {
		case "GET":
			if len(cmd.Args) > 0 {
				if key, ok := cmd.Args[0].(string); ok {
					val, _ := m.Get(ctx, key)
					results[i] = val
				}
			}
		case "SET":
			if len(cmd.Args) >= 2 {
				if key, ok := cmd.Args[0].(string); ok {
					m.Set(ctx, key, cmd.Args[1], 0)
					results[i] = "OK"
				}
			}
		case "INCR":
			if len(cmd.Args) > 0 {
				if key, ok := cmd.Args[0].(string); ok {
					val, _ := m.Increment(ctx, key, 1)
					results[i] = val
				}
			}
		default:
			results[i] = nil
		}
	}
	return results, nil
}

// MockMemoryBackend provides mock memory backend
type MockMemoryBackend struct {
	data  map[string]interface{}
	stats *MemoryStats
	mutex sync.RWMutex
}

func (m *MockMemoryBackend) Get(ctx context.Context, key string) (interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.data == nil {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	if value, exists := m.data[key]; exists {
		return value, nil
	}
	return nil, fmt.Errorf("key not found: %s", key)
}

func (m *MockMemoryBackend) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.data == nil {
		m.data = make(map[string]interface{})
	}

	m.data[key] = value
	return nil
}

func (m *MockMemoryBackend) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.data == nil {
		m.data = make(map[string]interface{})
	}

	current, exists := m.data[key]
	if !exists {
		m.data[key] = delta
		return delta, nil
	}

	if val, ok := current.(int64); ok {
		newVal := val + delta
		m.data[key] = newVal
		return newVal, nil
	}

	return 0, fmt.Errorf("value is not an integer")
}

func (m *MockMemoryBackend) Delete(ctx context.Context, key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.data != nil {
		delete(m.data, key)
	}
	return nil
}

func (m *MockMemoryBackend) Cleanup(ctx context.Context) error {
	// Mock cleanup - remove some entries
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.data != nil && len(m.data) > 100 {
		// Remove oldest entries (simplified)
		count := 0
		for key := range m.data {
			delete(m.data, key)
			count++
			if count >= 10 {
				break
			}
		}
	}
	return nil
}

func (m *MockMemoryBackend) GetStats() *MemoryStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	entryCount := 0
	if m.data != nil {
		entryCount = len(m.data)
	}

	return &MemoryStats{
		EntryCount:    entryCount,
		MemoryUsage:   int64(entryCount * 100), // Mock memory usage
		HitRate:       decimal.NewFromFloat(0.85),
		EvictionCount: 10,
	}
}

// MockPolicyEngine provides mock policy engine
type MockPolicyEngine struct {
	policies map[string]*Policy
	mutex    sync.RWMutex
}

func (m *MockPolicyEngine) LoadPolicies(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.policies == nil {
		m.policies = make(map[string]*Policy)
	}

	// Load default policies
	defaultPolicy := &Policy{
		ID:          "default",
		Name:        "Default Rate Limiting Policy",
		Description: "Default policy for all users",
		Scope: PolicyScope{
			Type:   "global",
			Values: []string{"*"},
		},
		Rules: []PolicyRule{
			{
				ID:        "default_rule",
				Type:      "rate_limit",
				Algorithm: "token_bucket",
				Limit:     1000,
				Window:    time.Hour,
				Enabled:   true,
				Priority:  1,
			},
		},
		Priority:  1,
		Enabled:   true,
		ValidFrom: time.Now().Add(-24 * time.Hour),
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
	}

	m.policies["default"] = defaultPolicy
	return nil
}

func (m *MockPolicyEngine) GetPolicy(ctx context.Context, scope PolicyScope) (*Policy, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return default policy for simplicity
	if policy, exists := m.policies["default"]; exists {
		return policy, nil
	}

	return nil, nil
}

func (m *MockPolicyEngine) EvaluatePolicy(ctx context.Context, request interface{}) (*PolicyDecision, error) {
	// Mock policy evaluation - always allow for simplicity
	return &PolicyDecision{
		Allowed: true,
		Policy:  m.policies["default"],
		Reason:  "Mock evaluation - allowed",
	}, nil
}

func (m *MockPolicyEngine) UpdatePolicy(ctx context.Context, policy *Policy) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.policies == nil {
		m.policies = make(map[string]*Policy)
	}

	policy.UpdatedAt = time.Now()
	m.policies[policy.ID] = policy
	return nil
}

func (m *MockPolicyEngine) DeletePolicy(ctx context.Context, policyID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.policies != nil {
		delete(m.policies, policyID)
	}
	return nil
}

// MockRuleEngine provides mock rule engine
type MockRuleEngine struct{}

func (m *MockRuleEngine) EvaluateRules(ctx context.Context, rules []PolicyRule, context map[string]interface{}) (*RuleDecision, error) {
	// Mock rule evaluation - always allow
	return &RuleDecision{
		Allowed:      true,
		MatchedRules: rules,
		Actions:      []RuleAction{},
		Score:        decimal.NewFromFloat(1.0),
	}, nil
}

func (m *MockRuleEngine) AddRule(ctx context.Context, rule *PolicyRule) error {
	return nil
}

func (m *MockRuleEngine) UpdateRule(ctx context.Context, rule *PolicyRule) error {
	return nil
}

func (m *MockRuleEngine) DeleteRule(ctx context.Context, ruleID string) error {
	return nil
}

// MockMetricsCollector provides mock metrics collection
type MockMetricsCollector struct {
	metrics []Metric
	mutex   sync.RWMutex
}

func (m *MockMetricsCollector) RecordMetric(ctx context.Context, metric *Metric) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.metrics = append(m.metrics, *metric)
	return nil
}

func (m *MockMetricsCollector) GetMetrics(ctx context.Context, query *MetricQuery) ([]*Metric, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var result []*Metric
	for _, metric := range m.metrics {
		if query.Name == "" || metric.Name == query.Name {
			if metric.Timestamp.After(query.StartTime) && metric.Timestamp.Before(query.EndTime) {
				result = append(result, &metric)
			}
		}
	}

	return result, nil
}

func (m *MockMetricsCollector) ExportMetrics(ctx context.Context, format string) ([]byte, error) {
	switch format {
	case "prometheus":
		return []byte("# HELP rate_limit_requests_total Total requests\nrate_limit_requests_total 100"), nil
	case "json":
		return []byte(`{"metrics": []}`), nil
	default:
		return []byte("mock metrics"), nil
	}
}

func (m *MockMetricsCollector) StartCollection(ctx context.Context) error {
	return nil
}

func (m *MockMetricsCollector) StopCollection(ctx context.Context) error {
	return nil
}

// MockAlertManager provides mock alert management
type MockAlertManager struct {
	alerts []Alert
	mutex  sync.RWMutex
}

func (m *MockAlertManager) CheckThresholds(ctx context.Context, metrics *UsageMetrics) error {
	// Mock threshold checking
	if metrics.TotalRequests > 1000 {
		alert := Alert{
			ID:           fmt.Sprintf("alert_%d", time.Now().Unix()),
			Type:         "high_usage",
			Severity:     "medium",
			Message:      "High request volume detected",
			Threshold:    decimal.NewFromFloat(1000),
			CurrentValue: decimal.NewFromInt(metrics.TotalRequests),
			CreatedAt:    time.Now(),
		}

		m.mutex.Lock()
		m.alerts = append(m.alerts, alert)
		m.mutex.Unlock()
	}

	return nil
}

func (m *MockAlertManager) SendAlert(ctx context.Context, alert *Alert) error {
	// Mock alert sending
	return nil
}

func (m *MockAlertManager) GetActiveAlerts(ctx context.Context) ([]*Alert, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var activeAlerts []*Alert
	for _, alert := range m.alerts {
		if alert.AcknowledgedAt == nil {
			activeAlerts = append(activeAlerts, &alert)
		}
	}

	return activeAlerts, nil
}

func (m *MockAlertManager) AcknowledgeAlert(ctx context.Context, alertID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for i := range m.alerts {
		if m.alerts[i].ID == alertID {
			now := time.Now()
			m.alerts[i].AcknowledgedAt = &now
			break
		}
	}

	return nil
}
