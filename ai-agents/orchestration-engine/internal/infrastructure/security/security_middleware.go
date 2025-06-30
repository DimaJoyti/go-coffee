package security

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SecurityMiddleware provides comprehensive security middleware
type SecurityMiddleware struct {
	authService     *AuthService
	rateLimiter     *RateLimiter
	inputValidator  *InputValidator
	tokenBlacklist  *TokenBlacklist
	config          *SecurityConfig
	logger          Logger
}

// SecurityConfig contains security middleware configuration
type SecurityConfig struct {
	// Authentication
	EnableAuthentication bool     `json:"enable_authentication"`
	SkipAuthPaths       []string `json:"skip_auth_paths"`
	
	// Rate limiting
	EnableRateLimit     bool          `json:"enable_rate_limit"`
	RateLimit           int           `json:"rate_limit"`
	RateLimitWindow     time.Duration `json:"rate_limit_window"`
	
	// Input validation
	EnableInputValidation bool `json:"enable_input_validation"`
	MaxRequestSize        int64 `json:"max_request_size"`
	
	// Security headers
	EnableSecurityHeaders bool              `json:"enable_security_headers"`
	SecurityHeaders       map[string]string `json:"security_headers"`
	
	// CORS
	EnableCORS      bool     `json:"enable_cors"`
	AllowedOrigins  []string `json:"allowed_origins"`
	AllowedMethods  []string `json:"allowed_methods"`
	AllowedHeaders  []string `json:"allowed_headers"`
	AllowCredentials bool    `json:"allow_credentials"`
	
	// API Key authentication
	EnableAPIKey    bool              `json:"enable_api_key"`
	APIKeys         map[string]string `json:"api_keys"` // key -> description
	APIKeyHeader    string            `json:"api_key_header"`
	
	// Request logging
	EnableRequestLogging bool `json:"enable_request_logging"`
	LogSensitiveData     bool `json:"log_sensitive_data"`
}

// SecurityContext contains security information for a request
type SecurityContext struct {
	UserID          string            `json:"user_id"`
	Username        string            `json:"username"`
	Roles           []string          `json:"roles"`
	Permissions     []string          `json:"permissions"`
	IsAuthenticated bool              `json:"is_authenticated"`
	AuthMethod      string            `json:"auth_method"`
	IPAddress       string            `json:"ip_address"`
	UserAgent       string            `json:"user_agent"`
	RequestID       string            `json:"request_id"`
	Metadata        map[string]string `json:"metadata"`
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(
	authService *AuthService,
	rateLimiter *RateLimiter,
	inputValidator *InputValidator,
	tokenBlacklist *TokenBlacklist,
	config *SecurityConfig,
	logger Logger,
) *SecurityMiddleware {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	return &SecurityMiddleware{
		authService:    authService,
		rateLimiter:    rateLimiter,
		inputValidator: inputValidator,
		tokenBlacklist: tokenBlacklist,
		config:         config,
		logger:         logger,
	}
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		EnableAuthentication:  true,
		SkipAuthPaths:        []string{"/health", "/metrics", "/api/v1/auth/login"},
		EnableRateLimit:      true,
		RateLimit:            100,
		RateLimitWindow:      time.Minute,
		EnableInputValidation: true,
		MaxRequestSize:       10 * 1024 * 1024, // 10MB
		EnableSecurityHeaders: true,
		SecurityHeaders: map[string]string{
			"X-Content-Type-Options":  "nosniff",
			"X-Frame-Options":         "DENY",
			"X-XSS-Protection":        "1; mode=block",
			"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
			"Content-Security-Policy": "default-src 'self'",
			"Referrer-Policy":         "strict-origin-when-cross-origin",
		},
		EnableCORS:           true,
		AllowedOrigins:       []string{"*"},
		AllowedMethods:       []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:       []string{"Content-Type", "Authorization", "X-Requested-With"},
		AllowCredentials:     false,
		EnableAPIKey:         false,
		APIKeyHeader:         "X-API-Key",
		EnableRequestLogging: true,
		LogSensitiveData:     false,
	}
}

// Handler returns the security middleware handler
func (sm *SecurityMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		// Generate request ID
		requestID := generateRequestID()
		ctx = context.WithValue(ctx, "request_id", requestID)
		r = r.WithContext(ctx)

		// Log request
		if sm.config.EnableRequestLogging {
			sm.logRequest(r)
		}

		// Apply security headers
		if sm.config.EnableSecurityHeaders {
			sm.applySecurityHeaders(w)
		}

		// Handle CORS
		if sm.config.EnableCORS {
			if sm.handleCORS(w, r) {
				return // Preflight request handled
			}
		}

		// Rate limiting
		if sm.config.EnableRateLimit {
			if !sm.checkRateLimit(w, r) {
				return // Rate limit exceeded
			}
		}

		// Input validation
		if sm.config.EnableInputValidation {
			if !sm.validateInput(w, r) {
				return // Input validation failed
			}
		}

		// Authentication
		securityCtx := &SecurityContext{
			IPAddress: getClientIP(r),
			UserAgent: r.UserAgent(),
			RequestID: requestID,
			Metadata:  make(map[string]string),
		}

		if sm.config.EnableAuthentication && !sm.isSkipAuthPath(r.URL.Path) {
			if !sm.authenticate(w, r, securityCtx) {
				return // Authentication failed
			}
		}

		// Add security context to request context
		ctx = context.WithValue(ctx, "security_context", securityCtx)
		r = r.WithContext(ctx)

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// authenticate performs authentication
func (sm *SecurityMiddleware) authenticate(w http.ResponseWriter, r *http.Request, securityCtx *SecurityContext) bool {
	// Try API key authentication first
	if sm.config.EnableAPIKey {
		if sm.authenticateAPIKey(r, securityCtx) {
			return true
		}
	}

	// Try JWT authentication
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		sm.writeErrorResponse(w, http.StatusUnauthorized, "Authorization header required")
		return false
	}

	token, err := sm.authService.jwtManager.ExtractTokenFromHeader(authHeader)
	if err != nil {
		sm.writeErrorResponse(w, http.StatusUnauthorized, "Invalid authorization header")
		return false
	}

	// Check token blacklist
	if sm.tokenBlacklist != nil {
		claims, _ := sm.authService.jwtManager.ValidateToken(token)
		if claims != nil && sm.tokenBlacklist.IsBlacklisted(claims.ID) {
			sm.writeErrorResponse(w, http.StatusUnauthorized, "Token has been revoked")
			return false
		}
	}

	// Validate token
	user, err := sm.authService.ValidateToken(r.Context(), token)
	if err != nil {
		sm.writeErrorResponse(w, http.StatusUnauthorized, "Invalid token")
		return false
	}

	// Update security context
	securityCtx.UserID = user.ID
	securityCtx.Username = user.Username
	securityCtx.Roles = user.Roles
	securityCtx.Permissions = user.Permissions
	securityCtx.IsAuthenticated = true
	securityCtx.AuthMethod = "jwt"

	return true
}

// authenticateAPIKey performs API key authentication
func (sm *SecurityMiddleware) authenticateAPIKey(r *http.Request, securityCtx *SecurityContext) bool {
	apiKey := r.Header.Get(sm.config.APIKeyHeader)
	if apiKey == "" {
		return false
	}

	// Check if API key is valid
	for validKey, description := range sm.config.APIKeys {
		if subtle.ConstantTimeCompare([]byte(apiKey), []byte(validKey)) == 1 {
			securityCtx.IsAuthenticated = true
			securityCtx.AuthMethod = "api_key"
			securityCtx.Metadata["api_key_description"] = description
			return true
		}
	}

	return false
}

// checkRateLimit checks rate limiting
func (sm *SecurityMiddleware) checkRateLimit(w http.ResponseWriter, r *http.Request) bool {
	if sm.rateLimiter == nil {
		return true
	}

	clientIP := getClientIP(r)
	if !sm.rateLimiter.Allow(clientIP) {
		sm.writeErrorResponse(w, http.StatusTooManyRequests, "Rate limit exceeded")
		return false
	}

	return true
}

// validateInput validates request input
func (sm *SecurityMiddleware) validateInput(w http.ResponseWriter, r *http.Request) bool {
	if sm.inputValidator == nil {
		return true
	}

	// Check request size
	if r.ContentLength > sm.config.MaxRequestSize {
		sm.writeErrorResponse(w, http.StatusRequestEntityTooLarge, "Request too large")
		return false
	}

	// Validate URL parameters
	for key, values := range r.URL.Query() {
		for _, value := range values {
			result := sm.inputValidator.ValidateString(value)
			if !result.Valid {
				sm.logger.Warn("Invalid URL parameter", "key", key, "value", value, "errors", result.Errors)
				sm.writeErrorResponse(w, http.StatusBadRequest, "Invalid URL parameter")
				return false
			}
		}
	}

	// Validate headers
	for key, values := range r.Header {
		// Skip certain headers
		if strings.ToLower(key) == "authorization" || strings.ToLower(key) == "cookie" {
			continue
		}

		for _, value := range values {
			result := sm.inputValidator.ValidateString(value)
			if !result.Valid {
				sm.logger.Warn("Invalid header", "key", key, "errors", result.Errors)
				sm.writeErrorResponse(w, http.StatusBadRequest, "Invalid header value")
				return false
			}
		}
	}

	return true
}

// handleCORS handles CORS requests
func (sm *SecurityMiddleware) handleCORS(w http.ResponseWriter, r *http.Request) bool {
	origin := r.Header.Get("Origin")
	
	// Check if origin is allowed
	if !sm.isOriginAllowed(origin) {
		sm.writeErrorResponse(w, http.StatusForbidden, "Origin not allowed")
		return true
	}

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(sm.config.AllowedMethods, ", "))
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(sm.config.AllowedHeaders, ", "))
	
	if sm.config.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	// Handle preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return true
	}

	return false
}

// applySecurityHeaders applies security headers
func (sm *SecurityMiddleware) applySecurityHeaders(w http.ResponseWriter) {
	for header, value := range sm.config.SecurityHeaders {
		w.Header().Set(header, value)
	}
}

// isSkipAuthPath checks if path should skip authentication
func (sm *SecurityMiddleware) isSkipAuthPath(path string) bool {
	for _, skipPath := range sm.config.SkipAuthPaths {
		if path == skipPath || strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// isOriginAllowed checks if origin is allowed for CORS
func (sm *SecurityMiddleware) isOriginAllowed(origin string) bool {
	if len(sm.config.AllowedOrigins) == 0 {
		return true
	}

	for _, allowedOrigin := range sm.config.AllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
	}

	return false
}

// logRequest logs the incoming request
func (sm *SecurityMiddleware) logRequest(r *http.Request) {
	logData := map[string]interface{}{
		"method":     r.Method,
		"path":       r.URL.Path,
		"ip":         getClientIP(r),
		"user_agent": r.UserAgent(),
	}

	if sm.config.LogSensitiveData {
		logData["headers"] = r.Header
		logData["query"] = r.URL.Query()
	}

	sm.logger.Info("HTTP request", logData)
}

// writeErrorResponse writes an error response
func (sm *SecurityMiddleware) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	// Write JSON response directly
	fmt.Fprintf(w, `{"error": true, "message": "%s", "code": %d}`, message, statusCode)
}

// getClientIP extracts the client IP address
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}

	return ip
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// GetSecurityContext extracts security context from request context
func GetSecurityContext(ctx context.Context) (*SecurityContext, bool) {
	securityCtx, ok := ctx.Value("security_context").(*SecurityContext)
	return securityCtx, ok
}

// RequireRole middleware that requires specific roles
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			securityCtx, ok := GetSecurityContext(r.Context())
			if !ok || !securityCtx.IsAuthenticated {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if user has any of the required roles
			hasRole := false
			for _, requiredRole := range roles {
				for _, userRole := range securityCtx.Roles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission middleware that requires specific permissions
func RequirePermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			securityCtx, ok := GetSecurityContext(r.Context())
			if !ok || !securityCtx.IsAuthenticated {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if user has any of the required permissions
			hasPermission := false
			for _, requiredPerm := range permissions {
				for _, userPerm := range securityCtx.Permissions {
					if userPerm == requiredPerm {
						hasPermission = true
						break
					}
				}
				if hasPermission {
					break
				}
			}

			if !hasPermission {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
