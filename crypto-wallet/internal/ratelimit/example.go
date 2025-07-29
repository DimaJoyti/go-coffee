package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"go.uber.org/zap"
)

// ExampleRateLimitingService demonstrates how to use the rate limiting system
type ExampleRateLimitingService struct {
	rateLimiter *RateLimiter
	middleware  *HTTPMiddleware
	logger      *logger.Logger
}

// NewExampleRateLimitingService creates a new example service
func NewExampleRateLimitingService(logger *logger.Logger) *ExampleRateLimitingService {
	// Create rate limiter with default configuration
	config := GetDefaultRateLimiterConfig()
	rateLimiter := NewRateLimiter(logger, config)

	// Create HTTP middleware
	middlewareConfig := GetDefaultMiddlewareConfig()
	middleware := NewHTTPMiddleware(rateLimiter, logger, middlewareConfig)

	return &ExampleRateLimitingService{
		rateLimiter: rateLimiter,
		middleware:  middleware,
		logger:      logger.Named("example-rate-limiting"),
	}
}

// Start starts the example service
func (s *ExampleRateLimitingService) Start(ctx context.Context) error {
	s.logger.Info("Starting example rate limiting service")

	// Start the rate limiter
	if err := s.rateLimiter.Start(ctx); err != nil {
		return fmt.Errorf("failed to start rate limiter: %w", err)
	}

	// Run examples
	s.runBasicRateLimitingExample(ctx)
	s.runQuotaManagementExample(ctx)
	s.runPolicyManagementExample(ctx)
	s.runMetricsExample(ctx)

	return nil
}

// Stop stops the example service
func (s *ExampleRateLimitingService) Stop() error {
	s.logger.Info("Stopping example rate limiting service")
	return s.rateLimiter.Stop()
}

// runBasicRateLimitingExample demonstrates basic rate limiting
func (s *ExampleRateLimitingService) runBasicRateLimitingExample(ctx context.Context) {
	s.logger.Info("Running basic rate limiting example")

	// Example 1: Token bucket rate limiting
	request := &RateLimitRequest{
		Key:       "user:123",
		Algorithm: "token_bucket",
		Limit:     10,
		Window:    time.Minute,
		Cost:      1,
		Metadata: map[string]interface{}{
			"endpoint": "/api/balance",
			"method":   "GET",
		},
	}

	for i := 0; i < 15; i++ {
		response, err := s.rateLimiter.CheckRateLimit(ctx, request)
		if err != nil {
			s.logger.Error("Rate limit check failed", zap.Error(err))
			continue
		}

		s.logger.Info("Rate limit check result",
			zap.Int("attempt", i+1),
			zap.Bool("allowed", response.Allowed),
			zap.Int64("remaining", response.Remaining),
			zap.String("algorithm", response.Algorithm))

		if !response.Allowed {
			s.logger.Info("Rate limit exceeded, waiting...",
				zap.Duration("retry_after", response.RetryAfter))
			time.Sleep(100 * time.Millisecond) // Short wait for demo
		}
	}

	// Example 2: Sliding window rate limiting
	s.logger.Info("Testing sliding window algorithm")
	request.Algorithm = "sliding_window"
	request.Key = "user:456"

	for i := 0; i < 5; i++ {
		response, err := s.rateLimiter.CheckRateLimit(ctx, request)
		if err != nil {
			s.logger.Error("Rate limit check failed", zap.Error(err))
			continue
		}

		s.logger.Info("Sliding window result",
			zap.Int("attempt", i+1),
			zap.Bool("allowed", response.Allowed),
			zap.Int64("remaining", response.Remaining))

		time.Sleep(100 * time.Millisecond)
	}
}

// runQuotaManagementExample demonstrates quota management
func (s *ExampleRateLimitingService) runQuotaManagementExample(ctx context.Context) {
	s.logger.Info("Running quota management example")

	// Example quota request
	quotaRequest := &QuotaRequest{
		UserID:    "user:789",
		APIKey:    "api_key_123",
		Resource:  "api_calls",
		Operation: "GET /api/portfolio",
		Cost:      1,
		Period:    "hour",
		Metadata: map[string]interface{}{
			"plan": "premium",
		},
	}

	// Make several quota checks
	for i := 0; i < 10; i++ {
		response, err := s.rateLimiter.CheckQuota(ctx, quotaRequest)
		if err != nil {
			s.logger.Error("Quota check failed", zap.Error(err))
			continue
		}

		s.logger.Info("Quota check result",
			zap.Int("attempt", i+1),
			zap.Bool("allowed", response.Allowed),
			zap.Int64("used", response.Used),
			zap.Int64("remaining", response.Remaining),
			zap.Int64("total", response.Total),
			zap.String("period", response.Period))

		if !response.Allowed {
			s.logger.Info("Quota exceeded",
				zap.Time("reset_time", response.ResetTime))
			break
		}
	}

	// Get quota usage summary
	usage, err := s.rateLimiter.GetQuotaUsage(ctx, "user:789", "hour")
	if err != nil {
		s.logger.Error("Failed to get quota usage", zap.Error(err))
	} else {
		s.logger.Info("Quota usage summary", zap.Any("usage", usage))
	}
}

// runPolicyManagementExample demonstrates policy management
func (s *ExampleRateLimitingService) runPolicyManagementExample(ctx context.Context) {
	s.logger.Info("Running policy management example")

	// Create a custom policy
	policy := &Policy{
		ID:          "premium_user_policy",
		Name:        "Premium User Rate Limiting",
		Description: "Enhanced rate limits for premium users",
		Scope: PolicyScope{
			Type:   "user",
			Values: []string{"premium_users"},
		},
		Rules: []PolicyRule{
			{
				ID:        "premium_api_limit",
				Type:      "rate_limit",
				Algorithm: "token_bucket",
				Limit:     5000,
				Window:    time.Hour,
				Enabled:   true,
				Priority:  1,
				Conditions: []RuleCondition{
					{
						Type:     "user_plan",
						Field:    "plan",
						Operator: "eq",
						Value:    "premium",
					},
				},
			},
		},
		Priority:  1,
		Enabled:   true,
		ValidFrom: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Update the policy
	if err := s.rateLimiter.UpdatePolicy(ctx, policy); err != nil {
		s.logger.Error("Failed to update policy", zap.Error(err))
	} else {
		s.logger.Info("Policy updated successfully", zap.String("policy_id", policy.ID))
	}
}

// runMetricsExample demonstrates metrics collection
func (s *ExampleRateLimitingService) runMetricsExample(ctx context.Context) {
	s.logger.Info("Running metrics example")

	// Get usage metrics
	metrics, err := s.rateLimiter.GetUsageMetrics(ctx, "user:123", "hour")
	if err != nil {
		s.logger.Error("Failed to get usage metrics", zap.Error(err))
	} else {
		s.logger.Info("Usage metrics",
			zap.String("key", metrics.Key),
			zap.Int64("total_requests", metrics.TotalRequests),
			zap.Int64("allowed_requests", metrics.AllowedRequests),
			zap.Int64("blocked_requests", metrics.BlockedRequests),
			zap.Duration("average_latency", metrics.AverageLatency))
	}

	// Get top users
	topUsers, err := s.rateLimiter.GetTopUsers(ctx, "hour", 5)
	if err != nil {
		s.logger.Error("Failed to get top users", zap.Error(err))
	} else {
		s.logger.Info("Top users by usage")
		for i, user := range topUsers {
			s.logger.Info("Top user",
				zap.Int("rank", i+1),
				zap.String("user_id", user.UserID),
				zap.Int64("total_requests", user.TotalRequests),
				zap.Time("last_activity", user.LastActivity))
		}
	}

	// Export metrics
	data, err := s.rateLimiter.ExportMetrics(ctx, "json", "hour")
	if err != nil {
		s.logger.Error("Failed to export metrics", zap.Error(err))
	} else {
		s.logger.Info("Exported metrics", zap.Int("size_bytes", len(data)))
	}
}

// SetupHTTPServer demonstrates how to set up an HTTP server with rate limiting
func (s *ExampleRateLimitingService) SetupHTTPServer() *http.ServeMux {
	mux := http.NewServeMux()

	// Apply rate limiting middleware to all routes
	rateLimitedMux := s.middleware.Handler(mux)

	// Define API endpoints
	mux.HandleFunc("/api/balance", s.handleBalance)
	mux.HandleFunc("/api/portfolio", s.handlePortfolio)
	mux.HandleFunc("/api/transfer", s.handleTransfer)
	mux.HandleFunc("/auth/login", s.handleLogin)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/metrics", s.handleMetrics)

	// Return the rate-limited mux
	return rateLimitedMux.(*http.ServeMux)
}

// Example HTTP handlers
func (s *ExampleRateLimitingService) handleBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"balance": "1000.00", "currency": "USD"}`))
}

func (s *ExampleRateLimitingService) handlePortfolio(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"assets": [{"symbol": "BTC", "amount": "0.5"}]}`))
}

func (s *ExampleRateLimitingService) handleTransfer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"transaction_id": "tx_123", "status": "pending"}`))
}

func (s *ExampleRateLimitingService) handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"token": "jwt_token_123", "expires_in": 3600}`))
}

func (s *ExampleRateLimitingService) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`))
}

func (s *ExampleRateLimitingService) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := s.rateLimiter.GetMetrics()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	// Simple JSON encoding of metrics
	response := fmt.Sprintf(`{
		"is_running": %t,
		"default_algorithm": "%s",
		"storage_backend": "%s",
		"token_bucket_enabled": %t,
		"quota_manager_enabled": %t,
		"metrics_enabled": %t
	}`, 
		metrics["is_running"], 
		metrics["default_algorithm"], 
		metrics["storage_backend"],
		metrics["token_bucket_enabled"],
		metrics["quota_manager_enabled"],
		metrics["metrics_enabled"])
	
	w.Write([]byte(response))
}

// DemonstrateAdvancedFeatures shows advanced rate limiting features
func (s *ExampleRateLimitingService) DemonstrateAdvancedFeatures(ctx context.Context) {
	s.logger.Info("Demonstrating advanced rate limiting features")

	// 1. Custom key extraction
	customConfig := GetDefaultMiddlewareConfig()
	customConfig.CustomKeyExtractor = func(r *http.Request) string {
		// Custom logic to extract key based on multiple factors
		userID := r.Header.Get("X-User-ID")
		apiKey := r.Header.Get("X-API-Key")
		if userID != "" && apiKey != "" {
			return fmt.Sprintf("user:%s:key:%s", userID, apiKey[:8])
		}
		return r.RemoteAddr
	}

	// 2. Different algorithms for different endpoints
	s.demonstrateAlgorithmComparison(ctx)

	// 3. Dynamic policy updates
	s.demonstrateDynamicPolicies(ctx)

	// 4. Burst handling
	s.demonstrateBurstHandling(ctx)
}

func (s *ExampleRateLimitingService) demonstrateAlgorithmComparison(ctx context.Context) {
	s.logger.Info("Comparing rate limiting algorithms")

	algorithms := []string{"token_bucket", "sliding_window", "fixed_window", "leaky_bucket"}
	
	for _, algorithm := range algorithms {
		s.logger.Info("Testing algorithm", zap.String("algorithm", algorithm))
		
		request := &RateLimitRequest{
			Key:       fmt.Sprintf("test:%s", algorithm),
			Algorithm: algorithm,
			Limit:     5,
			Window:    10 * time.Second,
			Cost:      1,
		}

		// Make rapid requests to test burst behavior
		for i := 0; i < 8; i++ {
			response, err := s.rateLimiter.CheckRateLimit(ctx, request)
			if err != nil {
				s.logger.Error("Rate limit check failed", zap.Error(err))
				continue
			}

			s.logger.Info("Algorithm test result",
				zap.String("algorithm", algorithm),
				zap.Int("request", i+1),
				zap.Bool("allowed", response.Allowed),
				zap.Int64("remaining", response.Remaining))

			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (s *ExampleRateLimitingService) demonstrateDynamicPolicies(ctx context.Context) {
	s.logger.Info("Demonstrating dynamic policy updates")

	// Create time-based policy
	timeBasedPolicy := &Policy{
		ID:          "business_hours_policy",
		Name:        "Business Hours Rate Limiting",
		Description: "Different limits during business hours",
		Scope: PolicyScope{
			Type:   "global",
			Values: []string{"*"},
		},
		Rules: []PolicyRule{
			{
				ID:        "business_hours_rule",
				Type:      "rate_limit",
				Algorithm: "token_bucket",
				Limit:     2000, // Higher limit during business hours
				Window:    time.Hour,
				Enabled:   true,
				Priority:  1,
				Conditions: []RuleCondition{
					{
						Type:     "time_based",
						Field:    "hour",
						Operator: "between",
						Value:    []int{9, 17}, // 9 AM to 5 PM
					},
				},
			},
		},
		Priority:  2,
		Enabled:   true,
		ValidFrom: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.rateLimiter.UpdatePolicy(ctx, timeBasedPolicy); err != nil {
		s.logger.Error("Failed to update time-based policy", zap.Error(err))
	} else {
		s.logger.Info("Time-based policy updated successfully")
	}
}

func (s *ExampleRateLimitingService) demonstrateBurstHandling(ctx context.Context) {
	s.logger.Info("Demonstrating burst handling")

	// Test burst with token bucket
	request := &RateLimitRequest{
		Key:       "burst:test",
		Algorithm: "token_bucket",
		Limit:     10,
		Window:    time.Minute,
		Cost:      5, // Higher cost per request
	}

	s.logger.Info("Testing burst with high-cost requests")
	for i := 0; i < 5; i++ {
		response, err := s.rateLimiter.CheckRateLimit(ctx, request)
		if err != nil {
			s.logger.Error("Burst test failed", zap.Error(err))
			continue
		}

		s.logger.Info("Burst test result",
			zap.Int("request", i+1),
			zap.Int64("cost", request.Cost),
			zap.Bool("allowed", response.Allowed),
			zap.Int64("remaining", response.Remaining))
	}
}
