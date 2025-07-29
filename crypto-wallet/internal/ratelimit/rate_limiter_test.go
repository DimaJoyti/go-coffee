package ratelimit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test logger
func createTestLogger(level string) *logger.Logger {
	logConfig := config.LoggingConfig{
		Level:  level,
		Format: "console",
		Output: "stdout",
	}
	return logger.NewLogger(logConfig)
}

func TestRateLimiter_BasicFunctionality(t *testing.T) {
	logger := createTestLogger("debug")
	config := GetDefaultRateLimiterConfig()
	rateLimiter := NewRateLimiter(logger, config)

	ctx := context.Background()
	err := rateLimiter.Start(ctx)
	require.NoError(t, err)
	defer rateLimiter.Stop()

	// Test basic rate limiting
	request := &RateLimitRequest{
		Key:       "test:user",
		Algorithm: "token_bucket",
		Limit:     5,
		Window:    time.Minute,
		Cost:      1,
	}

	// First 5 requests should be allowed
	for i := 0; i < 5; i++ {
		response, err := rateLimiter.CheckRateLimit(ctx, request)
		require.NoError(t, err)
		assert.True(t, response.Allowed, "Request %d should be allowed", i+1)
		assert.Equal(t, int64(4-i), response.Remaining, "Remaining tokens should decrease")
	}

	// 6th request should be blocked
	response, err := rateLimiter.CheckRateLimit(ctx, request)
	require.NoError(t, err)
	assert.False(t, response.Allowed, "6th request should be blocked")
	assert.Equal(t, int64(0), response.Remaining, "No tokens should remain")
}

func TestRateLimiter_DifferentAlgorithms(t *testing.T) {
	logger := createTestLogger("debug")
	config := GetDefaultRateLimiterConfig()
	rateLimiter := NewRateLimiter(logger, config)

	ctx := context.Background()
	err := rateLimiter.Start(ctx)
	require.NoError(t, err)
	defer rateLimiter.Stop()

	algorithms := []string{"token_bucket", "sliding_window", "fixed_window", "leaky_bucket"}

	for _, algorithm := range algorithms {
		t.Run(algorithm, func(t *testing.T) {
			request := &RateLimitRequest{
				Key:       "test:" + algorithm,
				Algorithm: algorithm,
				Limit:     3,
				Window:    10 * time.Second,
				Cost:      1,
			}

			// Test that algorithm works
			response, err := rateLimiter.CheckRateLimit(ctx, request)
			require.NoError(t, err)
			assert.True(t, response.Allowed)
			assert.Equal(t, algorithm, response.Algorithm)
		})
	}
}

func TestRateLimiter_QuotaManagement(t *testing.T) {
	logger := createTestLogger("debug")
	config := GetDefaultRateLimiterConfig()
	rateLimiter := NewRateLimiter(logger, config)

	ctx := context.Background()
	err := rateLimiter.Start(ctx)
	require.NoError(t, err)
	defer rateLimiter.Stop()

	// Test quota checking
	quotaRequest := &QuotaRequest{
		UserID:    "test:user",
		Resource:  "api_calls",
		Operation: "GET /api/test",
		Cost:      1,
		Period:    "hour",
	}

	// First request should be allowed
	response, err := rateLimiter.CheckQuota(ctx, quotaRequest)
	require.NoError(t, err)
	assert.True(t, response.Allowed)
	assert.Equal(t, int64(1), response.Used)
	assert.Equal(t, "hour", response.Period)

	// Get quota usage
	usage, err := rateLimiter.GetQuotaUsage(ctx, "test:user", "hour")
	require.NoError(t, err)
	assert.NotEmpty(t, usage)
}

func TestRateLimiter_UsageMetrics(t *testing.T) {
	logger := createTestLogger("debug")
	config := GetDefaultRateLimiterConfig()
	rateLimiter := NewRateLimiter(logger, config)

	ctx := context.Background()
	err := rateLimiter.Start(ctx)
	require.NoError(t, err)
	defer rateLimiter.Stop()

	// Make some requests to generate metrics
	request := &RateLimitRequest{
		Key:       "metrics:test",
		Algorithm: "token_bucket",
		Limit:     10,
		Window:    time.Minute,
		Cost:      1,
	}

	for i := 0; i < 5; i++ {
		_, err := rateLimiter.CheckRateLimit(ctx, request)
		require.NoError(t, err)
	}

	// Get usage metrics
	metrics, err := rateLimiter.GetUsageMetrics(ctx, "metrics:test", "hour")
	require.NoError(t, err)
	assert.Equal(t, "metrics:test", metrics.Key)
	assert.Equal(t, int64(5), metrics.TotalRequests)

	// Get top users
	topUsers, err := rateLimiter.GetTopUsers(ctx, "hour", 5)
	require.NoError(t, err)
	assert.NotEmpty(t, topUsers)

	// Export metrics
	data, err := rateLimiter.ExportMetrics(ctx, "json", "hour")
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestHTTPMiddleware_BasicFunctionality(t *testing.T) {
	logger := createTestLogger("debug")
	config := GetDefaultRateLimiterConfig()
	rateLimiter := NewRateLimiter(logger, config)

	ctx := context.Background()
	err := rateLimiter.Start(ctx)
	require.NoError(t, err)
	defer rateLimiter.Stop()

	// Create middleware with strict limits for testing
	middlewareConfig := GetDefaultMiddlewareConfig()
	middlewareConfig.DefaultLimit = 3
	middlewareConfig.DefaultWindow = time.Minute
	middleware := NewHTTPMiddleware(rateLimiter, logger, middlewareConfig)

	// Create test handler
	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Test requests from same IP
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if i < 3 {
			assert.Equal(t, http.StatusOK, w.Code, "Request %d should be allowed", i+1)
			assert.Contains(t, w.Header().Get("X-RateLimit-Remaining"), "", "Should have rate limit headers")
		} else {
			assert.Equal(t, http.StatusTooManyRequests, w.Code, "Request %d should be rate limited", i+1)
		}
	}
}

func TestHTTPMiddleware_SkipPaths(t *testing.T) {
	logger := createTestLogger("debug")
	config := GetDefaultRateLimiterConfig()
	rateLimiter := NewRateLimiter(logger, config)

	ctx := context.Background()
	err := rateLimiter.Start(ctx)
	require.NoError(t, err)
	defer rateLimiter.Stop()

	middlewareConfig := GetDefaultMiddlewareConfig()
	middlewareConfig.DefaultLimit = 1 // Very strict limit
	middlewareConfig.SkipPaths = []string{"/health", "/metrics"}
	middleware := NewHTTPMiddleware(rateLimiter, logger, middlewareConfig)

	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Test that health endpoint is not rate limited
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Health endpoint should not be rate limited")
	}

	// Test that regular endpoint is rate limited
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "First request should be allowed")

	// Second request should be blocked
	req = httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code, "Second request should be rate limited")
}

func TestHTTPMiddleware_DifferentKeyExtractors(t *testing.T) {
	logger := createTestLogger("debug")
	config := GetDefaultRateLimiterConfig()
	rateLimiter := NewRateLimiter(logger, config)

	ctx := context.Background()
	err := rateLimiter.Start(ctx)
	require.NoError(t, err)
	defer rateLimiter.Stop()

	testCases := []struct {
		name         string
		keyExtractor string
		setupRequest func(*http.Request)
		expectedKey  string
	}{
		{
			name:         "IP-based",
			keyExtractor: "ip",
			setupRequest: func(req *http.Request) {
				req.RemoteAddr = "192.168.1.100:12345"
			},
			expectedKey: "192.168.1.100",
		},
		{
			name:         "User ID-based",
			keyExtractor: "user_id",
			setupRequest: func(req *http.Request) {
				req.Header.Set("X-User-ID", "user123")
			},
			expectedKey: "user123",
		},
		{
			name:         "API Key-based",
			keyExtractor: "api_key",
			setupRequest: func(req *http.Request) {
				req.Header.Set("X-API-Key", "api_key_456")
			},
			expectedKey: "api_key_456",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			middlewareConfig := GetDefaultMiddlewareConfig()
			middlewareConfig.KeyExtractor = tc.keyExtractor
			middlewareConfig.DefaultLimit = 1
			middleware := NewHTTPMiddleware(rateLimiter, logger, middlewareConfig)

			handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			// First request should be allowed
			req := httptest.NewRequest("GET", "/api/test", nil)
			tc.setupRequest(req)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code, "First request should be allowed")

			// Second request should be blocked (same key)
			req = httptest.NewRequest("GET", "/api/test", nil)
			tc.setupRequest(req)
			w = httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, http.StatusTooManyRequests, w.Code, "Second request should be rate limited")
		})
	}
}

func TestRateLimiter_Configuration(t *testing.T) {
	// Test default configuration
	defaultConfig := GetDefaultRateLimiterConfig()
	err := ValidateRateLimiterConfig(defaultConfig)
	assert.NoError(t, err, "Default configuration should be valid")

	// Test high throughput configuration
	highThroughputConfig := GetHighThroughputConfig()
	err = ValidateRateLimiterConfig(highThroughputConfig)
	assert.NoError(t, err, "High throughput configuration should be valid")

	// Test strict security configuration
	strictConfig := GetStrictSecurityConfig()
	err = ValidateRateLimiterConfig(strictConfig)
	assert.NoError(t, err, "Strict security configuration should be valid")

	// Test development configuration
	devConfig := GetDevelopmentConfig()
	err = ValidateRateLimiterConfig(devConfig)
	assert.NoError(t, err, "Development configuration should be valid")

	// Test invalid configuration
	invalidConfig := GetDefaultRateLimiterConfig()
	invalidConfig.DefaultAlgorithm = "invalid_algorithm"
	err = ValidateRateLimiterConfig(invalidConfig)
	assert.Error(t, err, "Invalid algorithm should cause validation error")
}

func TestRateLimiter_PolicyManagement(t *testing.T) {
	logger := createTestLogger("debug")
	config := GetDefaultRateLimiterConfig()
	rateLimiter := NewRateLimiter(logger, config)

	ctx := context.Background()
	err := rateLimiter.Start(ctx)
	require.NoError(t, err)
	defer rateLimiter.Stop()

	// Create a test policy
	policy := &Policy{
		ID:          "test_policy",
		Name:        "Test Policy",
		Description: "Test policy for unit tests",
		Scope: PolicyScope{
			Type:   "user",
			Values: []string{"test_user"},
		},
		Rules: []PolicyRule{
			{
				ID:        "test_rule",
				Type:      "rate_limit",
				Algorithm: "token_bucket",
				Limit:     100,
				Window:    time.Hour,
				Enabled:   true,
				Priority:  1,
			},
		},
		Priority:  1,
		Enabled:   true,
		ValidFrom: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Update policy
	err = rateLimiter.UpdatePolicy(ctx, policy)
	assert.NoError(t, err, "Policy update should succeed")
}

func BenchmarkRateLimiter_CheckRateLimit(b *testing.B) {
	logger := createTestLogger("error") // Reduce log noise
	config := GetDefaultRateLimiterConfig()
	rateLimiter := NewRateLimiter(logger, config)

	ctx := context.Background()
	rateLimiter.Start(ctx)
	defer rateLimiter.Stop()

	request := &RateLimitRequest{
		Key:       "benchmark:test",
		Algorithm: "token_bucket",
		Limit:     1000000, // High limit to avoid blocking
		Window:    time.Hour,
		Cost:      1,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := rateLimiter.CheckRateLimit(ctx, request)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkHTTPMiddleware_Handler(b *testing.B) {
	logger := createTestLogger("error")
	config := GetDefaultRateLimiterConfig()
	rateLimiter := NewRateLimiter(logger, config)

	ctx := context.Background()
	rateLimiter.Start(ctx)
	defer rateLimiter.Stop()

	middlewareConfig := GetDefaultMiddlewareConfig()
	middlewareConfig.DefaultLimit = 1000000 // High limit
	middleware := NewHTTPMiddleware(rateLimiter, logger, middlewareConfig)

	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}
