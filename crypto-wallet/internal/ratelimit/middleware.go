package ratelimit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"go.uber.org/zap"
)

// HTTPMiddleware provides HTTP middleware for rate limiting
type HTTPMiddleware struct {
	rateLimiter *RateLimiter
	logger      *logger.Logger
	config      MiddlewareConfig
}

// MiddlewareConfig holds configuration for HTTP middleware
type MiddlewareConfig struct {
	Enabled              bool                                                         `json:"enabled" yaml:"enabled"`
	KeyExtractor         string                                                       `json:"key_extractor" yaml:"key_extractor"` // "ip", "user_id", "api_key", "custom"
	CustomKeyExtractor   func(*http.Request) string                                   `json:"-" yaml:"-"`
	DefaultLimit         int64                                                        `json:"default_limit" yaml:"default_limit"`
	DefaultWindow        time.Duration                                                `json:"default_window" yaml:"default_window"`
	DefaultAlgorithm     string                                                       `json:"default_algorithm" yaml:"default_algorithm"`
	SkipPaths            []string                                                     `json:"skip_paths" yaml:"skip_paths"`
	SkipMethods          []string                                                     `json:"skip_methods" yaml:"skip_methods"`
	HeaderPrefix         string                                                       `json:"header_prefix" yaml:"header_prefix"`
	IncludeHeaders       bool                                                         `json:"include_headers" yaml:"include_headers"`
	OnLimitExceeded      func(http.ResponseWriter, *http.Request, *RateLimitResponse) `json:"-" yaml:"-"`
	TrustedProxies       []string                                                     `json:"trusted_proxies" yaml:"trusted_proxies"`
	EnableQuotaChecking  bool                                                         `json:"enable_quota_checking" yaml:"enable_quota_checking"`
	QuotaKeyExtractor    string                                                       `json:"quota_key_extractor" yaml:"quota_key_extractor"` // "user_id", "api_key"
	DefaultQuotaResource string                                                       `json:"default_quota_resource" yaml:"default_quota_resource"`
	DefaultQuotaPeriod   string                                                       `json:"default_quota_period" yaml:"default_quota_period"`
}

// RateLimitErrorResponse represents an error response for rate limiting
type RateLimitErrorResponse struct {
	Error      string    `json:"error"`
	Message    string    `json:"message"`
	Code       string    `json:"code"`
	Limit      int64     `json:"limit"`
	Remaining  int64     `json:"remaining"`
	ResetTime  time.Time `json:"reset_time"`
	RetryAfter int64     `json:"retry_after_seconds"`
	Timestamp  time.Time `json:"timestamp"`
}

// QuotaErrorResponse represents an error response for quota limits
type QuotaErrorResponse struct {
	Error     string    `json:"error"`
	Message   string    `json:"message"`
	Code      string    `json:"code"`
	Used      int64     `json:"used"`
	Total     int64     `json:"total"`
	Remaining int64     `json:"remaining"`
	Period    string    `json:"period"`
	ResetTime time.Time `json:"reset_time"`
	Timestamp time.Time `json:"timestamp"`
}

// NewHTTPMiddleware creates a new HTTP middleware instance
func NewHTTPMiddleware(rateLimiter *RateLimiter, logger *logger.Logger, config MiddlewareConfig) *HTTPMiddleware {
	return &HTTPMiddleware{
		rateLimiter: rateLimiter,
		logger:      logger.Named("rate-limit-middleware"),
		config:      config,
	}
}

// Handler returns an HTTP middleware handler
func (m *HTTPMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Skip rate limiting for certain paths or methods
		if m.shouldSkip(r) {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()

		// Extract rate limiting key
		key := m.extractKey(r)
		if key == "" {
			m.logger.Warn("Failed to extract rate limiting key", zap.String("path", r.URL.Path))
			next.ServeHTTP(w, r)
			return
		}

		// Check rate limit
		rateLimitRequest := &RateLimitRequest{
			Key:       key,
			Algorithm: m.config.DefaultAlgorithm,
			Limit:     m.config.DefaultLimit,
			Window:    m.config.DefaultWindow,
			Cost:      1,
			Metadata: map[string]interface{}{
				"method":     r.Method,
				"path":       r.URL.Path,
				"user_agent": r.UserAgent(),
				"ip":         m.getClientIP(r),
			},
			Timestamp: time.Now(),
		}

		rateLimitResponse, err := m.rateLimiter.CheckRateLimit(ctx, rateLimitRequest)
		if err != nil {
			m.logger.Error("Rate limit check failed", zap.Error(err), zap.String("key", key))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Add rate limit headers
		if m.config.IncludeHeaders {
			m.addRateLimitHeaders(w, rateLimitResponse)
		}

		// Check if rate limit exceeded
		if !rateLimitResponse.Allowed {
			m.handleRateLimitExceeded(w, r, rateLimitResponse)
			return
		}

		// Check quota if enabled
		if m.config.EnableQuotaChecking {
			quotaKey := m.extractQuotaKey(r)
			if quotaKey != "" {
				quotaRequest := &QuotaRequest{
					UserID:    quotaKey,
					APIKey:    m.extractAPIKey(r),
					Resource:  m.config.DefaultQuotaResource,
					Operation: fmt.Sprintf("%s %s", r.Method, r.URL.Path),
					Cost:      1,
					Period:    m.config.DefaultQuotaPeriod,
					Metadata: map[string]interface{}{
						"method": r.Method,
						"path":   r.URL.Path,
					},
					Timestamp: time.Now(),
				}

				quotaResponse, err := m.rateLimiter.CheckQuota(ctx, quotaRequest)
				if err != nil {
					m.logger.Error("Quota check failed", zap.Error(err), zap.String("user_id", quotaKey))
				} else {
					// Add quota headers
					if m.config.IncludeHeaders {
						m.addQuotaHeaders(w, quotaResponse)
					}

					// Check if quota exceeded
					if !quotaResponse.Allowed {
						m.handleQuotaExceeded(w, r, quotaResponse)
						return
					}
				}
			}
		}

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// HandlerFunc returns an HTTP middleware handler function
func (m *HTTPMiddleware) HandlerFunc(next http.HandlerFunc) http.HandlerFunc {
	return m.Handler(next).ServeHTTP
}

// shouldSkip determines if rate limiting should be skipped for this request
func (m *HTTPMiddleware) shouldSkip(r *http.Request) bool {
	// Check skip paths
	for _, path := range m.config.SkipPaths {
		if strings.HasPrefix(r.URL.Path, path) {
			return true
		}
	}

	// Check skip methods
	for _, method := range m.config.SkipMethods {
		if r.Method == method {
			return true
		}
	}

	return false
}

// extractKey extracts the rate limiting key from the request
func (m *HTTPMiddleware) extractKey(r *http.Request) string {
	switch m.config.KeyExtractor {
	case "ip":
		return m.getClientIP(r)
	case "user_id":
		return m.extractUserID(r)
	case "api_key":
		return m.extractAPIKey(r)
	case "custom":
		if m.config.CustomKeyExtractor != nil {
			return m.config.CustomKeyExtractor(r)
		}
		return m.getClientIP(r)
	default:
		return m.getClientIP(r)
	}
}

// extractQuotaKey extracts the quota key from the request
func (m *HTTPMiddleware) extractQuotaKey(r *http.Request) string {
	switch m.config.QuotaKeyExtractor {
	case "user_id":
		return m.extractUserID(r)
	case "api_key":
		return m.extractAPIKey(r)
	default:
		return m.extractUserID(r)
	}
}

// getClientIP extracts the client IP address from the request
func (m *HTTPMiddleware) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if m.isTrustedProxy(r.RemoteAddr) {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if m.isTrustedProxy(r.RemoteAddr) {
			return xri
		}
	}

	// Use remote address
	ip := r.RemoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}

	return ip
}

// extractUserID extracts the user ID from the request
func (m *HTTPMiddleware) extractUserID(r *http.Request) string {
	// Try to get from header
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}

	// Try to get from context (set by authentication middleware)
	if userID := r.Context().Value("user_id"); userID != nil {
		if uid, ok := userID.(string); ok {
			return uid
		}
	}

	// Try to get from query parameter
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		return userID
	}

	return ""
}

// extractAPIKey extracts the API key from the request
func (m *HTTPMiddleware) extractAPIKey(r *http.Request) string {
	// Try Authorization header
	if auth := r.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
		if strings.HasPrefix(auth, "ApiKey ") {
			return strings.TrimPrefix(auth, "ApiKey ")
		}
	}

	// Try X-API-Key header
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return apiKey
	}

	// Try query parameter
	if apiKey := r.URL.Query().Get("api_key"); apiKey != "" {
		return apiKey
	}

	return ""
}

// isTrustedProxy checks if the remote address is a trusted proxy
func (m *HTTPMiddleware) isTrustedProxy(remoteAddr string) bool {
	if len(m.config.TrustedProxies) == 0 {
		return true // Trust all if no specific proxies configured
	}

	ip := remoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}

	for _, proxy := range m.config.TrustedProxies {
		if ip == proxy {
			return true
		}
	}

	return false
}

// addRateLimitHeaders adds rate limit headers to the response
func (m *HTTPMiddleware) addRateLimitHeaders(w http.ResponseWriter, response *RateLimitResponse) {
	prefix := m.config.HeaderPrefix
	if prefix == "" {
		prefix = "X-RateLimit"
	}

	w.Header().Set(prefix+"-Limit", strconv.FormatInt(response.TotalLimit, 10))
	w.Header().Set(prefix+"-Remaining", strconv.FormatInt(response.Remaining, 10))
	w.Header().Set(prefix+"-Reset", strconv.FormatInt(response.ResetTime.Unix(), 10))
	w.Header().Set(prefix+"-Window", response.WindowSize.String())
	w.Header().Set(prefix+"-Algorithm", response.Algorithm)

	if !response.Allowed {
		w.Header().Set("Retry-After", strconv.FormatInt(int64(response.RetryAfter.Seconds()), 10))
	}
}

// addQuotaHeaders adds quota headers to the response
func (m *HTTPMiddleware) addQuotaHeaders(w http.ResponseWriter, response *QuotaResponse) {
	prefix := m.config.HeaderPrefix
	if prefix == "" {
		prefix = "X-Quota"
	}

	w.Header().Set(prefix+"-Used", strconv.FormatInt(response.Used, 10))
	w.Header().Set(prefix+"-Remaining", strconv.FormatInt(response.Remaining, 10))
	w.Header().Set(prefix+"-Total", strconv.FormatInt(response.Total, 10))
	w.Header().Set(prefix+"-Period", response.Period)
	w.Header().Set(prefix+"-Reset", strconv.FormatInt(response.ResetTime.Unix(), 10))
}

// handleRateLimitExceeded handles rate limit exceeded scenarios
func (m *HTTPMiddleware) handleRateLimitExceeded(w http.ResponseWriter, r *http.Request, response *RateLimitResponse) {
	if m.config.OnLimitExceeded != nil {
		m.config.OnLimitExceeded(w, r, response)
		return
	}

	// Default rate limit exceeded response
	errorResponse := RateLimitErrorResponse{
		Error:      "Rate limit exceeded",
		Message:    "Too many requests. Please try again later.",
		Code:       "RATE_LIMIT_EXCEEDED",
		Limit:      response.TotalLimit,
		Remaining:  response.Remaining,
		ResetTime:  response.ResetTime,
		RetryAfter: int64(response.RetryAfter.Seconds()),
		Timestamp:  time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	json.NewEncoder(w).Encode(errorResponse)
}

// handleQuotaExceeded handles quota exceeded scenarios
func (m *HTTPMiddleware) handleQuotaExceeded(w http.ResponseWriter, r *http.Request, response *QuotaResponse) {
	errorResponse := QuotaErrorResponse{
		Error:     "Quota exceeded",
		Message:   "Usage quota exceeded for this period. Please upgrade your plan or wait for quota reset.",
		Code:      "QUOTA_EXCEEDED",
		Used:      response.Used,
		Total:     response.Total,
		Remaining: response.Remaining,
		Period:    response.Period,
		ResetTime: response.ResetTime,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusPaymentRequired) // 402 Payment Required
	json.NewEncoder(w).Encode(errorResponse)
}

// GetDefaultMiddlewareConfig returns default middleware configuration
func GetDefaultMiddlewareConfig() MiddlewareConfig {
	return MiddlewareConfig{
		Enabled:              true,
		KeyExtractor:         "ip",
		DefaultLimit:         1000,
		DefaultWindow:        time.Hour,
		DefaultAlgorithm:     "token_bucket",
		SkipPaths:            []string{"/health", "/metrics", "/favicon.ico"},
		SkipMethods:          []string{"OPTIONS"},
		HeaderPrefix:         "X-RateLimit",
		IncludeHeaders:       true,
		TrustedProxies:       []string{"127.0.0.1", "::1"},
		EnableQuotaChecking:  false,
		QuotaKeyExtractor:    "user_id",
		DefaultQuotaResource: "api_calls",
		DefaultQuotaPeriod:   "hour",
	}
}

// GetAPIKeyMiddlewareConfig returns configuration for API key-based rate limiting
func GetAPIKeyMiddlewareConfig() MiddlewareConfig {
	config := GetDefaultMiddlewareConfig()
	config.KeyExtractor = "api_key"
	config.EnableQuotaChecking = true
	config.QuotaKeyExtractor = "api_key"
	config.DefaultLimit = 10000
	config.DefaultQuotaResource = "api_calls"
	config.DefaultQuotaPeriod = "day"
	return config
}

// GetUserBasedMiddlewareConfig returns configuration for user-based rate limiting
func GetUserBasedMiddlewareConfig() MiddlewareConfig {
	config := GetDefaultMiddlewareConfig()
	config.KeyExtractor = "user_id"
	config.EnableQuotaChecking = true
	config.QuotaKeyExtractor = "user_id"
	config.DefaultLimit = 5000
	config.DefaultQuotaResource = "api_calls"
	config.DefaultQuotaPeriod = "hour"
	return config
}

// GetStrictMiddlewareConfig returns configuration with strict rate limiting
func GetStrictMiddlewareConfig() MiddlewareConfig {
	config := GetDefaultMiddlewareConfig()
	config.DefaultLimit = 100
	config.DefaultWindow = 15 * time.Minute
	config.DefaultAlgorithm = "leaky_bucket"
	config.SkipPaths = []string{"/health"}
	config.EnableQuotaChecking = true
	config.DefaultQuotaResource = "api_calls"
	config.DefaultQuotaPeriod = "hour"
	return config
}
