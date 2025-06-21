package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/DimaJoyti/go-coffee/domain/shared"
	"github.com/DimaJoyti/go-coffee/domain/tenant"
	"github.com/gin-gonic/gin"
)

// TenantContextMiddleware extracts and validates tenant context from requests
type TenantContextMiddleware struct {
	tenantRepository tenant.TenantRepository
	validator        *TenantContextValidator
	jwtExtractor     *TenantJWTExtractor
}

// TenantContextConfig holds configuration for tenant context middleware
type TenantContextConfig struct {
	TenantRepository tenant.TenantRepository
	JWTExtractor     *TenantJWTExtractor
}

// NewTenantContextMiddleware creates a new tenant context middleware
func NewTenantContextMiddleware(tenantRepository tenant.TenantRepository) *TenantContextMiddleware {
	return &TenantContextMiddleware{
		tenantRepository: tenantRepository,
		validator:        NewTenantContextValidator(),
		jwtExtractor:     nil, // Optional JWT extractor
	}
}

// NewTenantContextMiddlewareWithConfig creates a new tenant context middleware with configuration
func NewTenantContextMiddlewareWithConfig(config TenantContextConfig) *TenantContextMiddleware {
	return &TenantContextMiddleware{
		tenantRepository: config.TenantRepository,
		validator:        NewTenantContextValidator(),
		jwtExtractor:     config.JWTExtractor,
	}
}

// TenantContextValidator validates tenant context
type TenantContextValidator struct{}

// NewTenantContextValidator creates a new tenant context validator
func NewTenantContextValidator() *TenantContextValidator {
	return &TenantContextValidator{}
}

// GinTenantContext returns a Gin middleware that extracts tenant context
func (m *TenantContextMiddleware) GinTenantContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantCtx, err := m.extractTenantContext(c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Add tenant context to Gin context
		c.Set("tenant_context", tenantCtx)

		// Add tenant context to request context
		ctx := shared.WithTenantContext(c.Request.Context(), tenantCtx)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// HTTPTenantContext returns an HTTP middleware that extracts tenant context
// Deprecated: Use HTTPTenantMiddleware.TenantContext instead for better error handling
func (m *TenantContextMiddleware) HTTPTenantContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantCtx, err := m.extractTenantContext(r)
		if err != nil {
			writeJSONErrorResponse(w, http.StatusUnauthorized, "Unauthorized", err.Error())
			return
		}

		// Add tenant context to request context
		ctx := shared.WithTenantContext(r.Context(), tenantCtx)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// extractTenantContext extracts tenant context from HTTP request
func (m *TenantContextMiddleware) extractTenantContext(r *http.Request) (*shared.TenantContext, error) {
	// Try different methods to extract tenant ID
	tenantID, err := m.extractTenantID(r)
	if err != nil {
		return nil, err
	}

	// Load tenant information from repository
	tenant, err := m.tenantRepository.FindByTenantID(r.Context(), tenantID)
	if err != nil {
		return nil, errors.New("tenant not found")
	}

	if tenant == nil {
		return nil, errors.New("tenant not found")
	}

	// Check if tenant is active
	if !tenant.IsActive() {
		return nil, errors.New("tenant is not active")
	}

	// Create tenant context
	var plan shared.SubscriptionPlan
	if tenant.Subscription() != nil {
		plan = tenant.Subscription().Plan()
	} else {
		plan = shared.SubscriptionBasic // Default plan
	}

	tenantCtx := shared.NewTenantContext(
		tenantID,
		tenant.Name(),
		plan,
	)

	// Set features based on subscription
	if tenant.Subscription() != nil {
		for feature, enabled := range tenant.Subscription().Features() {
			if enabled {
				tenantCtx.EnableFeature(feature)
			}
		}
	}

	// Add metadata
	tenantCtx.SetMetadata("tenant_type", tenant.TenantType().String())
	tenantCtx.SetMetadata("location_count", len(tenant.Locations()))

	return tenantCtx, nil
}

// extractTenantID extracts tenant ID from various sources in the request
func (m *TenantContextMiddleware) extractTenantID(r *http.Request) (shared.TenantID, error) {
	// Method 1: From subdomain (e.g., tenant1.api.example.com)
	if tenantID := m.extractFromSubdomain(r); tenantID != "" {
		return shared.NewTenantID(tenantID)
	}

	// Method 2: From custom header
	if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
		return shared.NewTenantID(tenantID)
	}

	// Method 3: From Authorization header (JWT token)
	if tenantID := m.extractFromJWT(r); tenantID != "" {
		return shared.NewTenantID(tenantID)
	}

	// Method 4: From query parameter
	if tenantID := r.URL.Query().Get("tenant_id"); tenantID != "" {
		return shared.NewTenantID(tenantID)
	}

	// Method 5: From path parameter (e.g., /api/v1/tenants/{tenant_id}/orders)
	if tenantID := m.extractFromPath(r); tenantID != "" {
		return shared.NewTenantID(tenantID)
	}

	return shared.TenantID{}, errors.New("tenant ID not found in request")
}

// extractFromSubdomain extracts tenant ID from subdomain
func (m *TenantContextMiddleware) extractFromSubdomain(r *http.Request) string {
	host := r.Host
	if host == "" {
		return ""
	}

	// Remove port if present
	if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	// Split by dots
	parts := strings.Split(host, ".")
	if len(parts) >= 3 {
		// Assume first part is tenant ID (e.g., tenant1.api.example.com)
		return parts[0]
	}

	return ""
}

// extractFromJWT extracts tenant ID from JWT token
func (m *TenantContextMiddleware) extractFromJWT(r *http.Request) string {
	// If JWT extractor is configured, use it
	if m.jwtExtractor != nil {
		tenantID, err := m.jwtExtractor.ExtractTenantFromJWT(r)
		if err != nil {
			return ""
		}
		return tenantID.Value()
	}

	// Fallback to basic token extraction (for backward compatibility)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Extract Bearer token
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}

	// In a real implementation without JWT extractor,
	// you would decode and validate the JWT here
	// For now, we'll return empty string to indicate no tenant found
	return ""
}

// extractFromPath extracts tenant ID from URL path
func (m *TenantContextMiddleware) extractFromPath(r *http.Request) string {
	path := r.URL.Path

	// Look for patterns like /api/v1/tenants/{tenant_id}/...
	if strings.Contains(path, "/tenants/") {
		parts := strings.Split(path, "/")
		for i, part := range parts {
			if part == "tenants" && i+1 < len(parts) {
				return parts[i+1]
			}
		}
	}

	return ""
}

// RequireFeature returns a middleware that requires a specific feature
func (m *TenantContextMiddleware) RequireFeature(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantCtx, exists := c.Get("tenant_context")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Tenant context not found",
			})
			c.Abort()
			return
		}

		ctx, ok := tenantCtx.(*shared.TenantContext)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Invalid tenant context",
			})
			c.Abort()
			return
		}

		if !ctx.HasFeature(feature) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Feature not available for current subscription",
				"feature": feature,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSubscription returns a middleware that requires a specific subscription plan
func (m *TenantContextMiddleware) RequireSubscription(plan shared.SubscriptionPlan) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantCtx, exists := c.Get("tenant_context")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Tenant context not found",
			})
			c.Abort()
			return
		}

		ctx, ok := tenantCtx.(*shared.TenantContext)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Invalid tenant context",
			})
			c.Abort()
			return
		}

		if ctx.Subscription() != plan {
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Forbidden",
				"message":       "Subscription plan required",
				"required_plan": plan.String(),
				"current_plan":  ctx.Subscription().String(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TenantIsolationMiddleware ensures data isolation between tenants
type TenantIsolationMiddleware struct{}

// NewTenantIsolationMiddleware creates a new tenant isolation middleware
func NewTenantIsolationMiddleware() *TenantIsolationMiddleware {
	return &TenantIsolationMiddleware{}
}

// ValidateTenantAccess validates that the request has access to the specified tenant resource
func (m *TenantIsolationMiddleware) ValidateTenantAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantCtx, exists := c.Get("tenant_context")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Tenant context not found",
			})
			c.Abort()
			return
		}

		ctx, ok := tenantCtx.(*shared.TenantContext)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Invalid tenant context",
			})
			c.Abort()
			return
		}

		// Extract tenant ID from URL parameters
		urlTenantID := c.Param("tenant_id")
		if urlTenantID != "" {
			if urlTenantID != ctx.TenantID().Value() {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "Forbidden",
					"message": "Access denied to tenant resource",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// GetTenantContext extracts tenant context from Gin context
func GetTenantContext(c *gin.Context) (*shared.TenantContext, error) {
	tenantCtx, exists := c.Get("tenant_context")
	if !exists {
		return nil, errors.New("tenant context not found")
	}

	ctx, ok := tenantCtx.(*shared.TenantContext)
	if !ok {
		return nil, errors.New("invalid tenant context")
	}

	return ctx, nil
}

// GetTenantID extracts tenant ID from Gin context
func GetTenantID(c *gin.Context) (shared.TenantID, error) {
	tenantCtx, err := GetTenantContext(c)
	if err != nil {
		return shared.TenantID{}, err
	}

	return tenantCtx.TenantID(), nil
}

// TenantMetricsMiddleware collects tenant-specific metrics
type TenantMetricsMiddleware struct{}

// NewTenantMetricsMiddleware creates a new tenant metrics middleware
func NewTenantMetricsMiddleware() *TenantMetricsMiddleware {
	return &TenantMetricsMiddleware{}
}

// CollectMetrics collects tenant-specific request metrics
func (m *TenantMetricsMiddleware) CollectMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantCtx, err := GetTenantContext(c)
		if err != nil {
			c.Next()
			return
		}

		// Record request metrics per tenant
		// In a real implementation, you would send these to a metrics system
		c.Set("tenant_metrics", map[string]interface{}{
			"tenant_id":    tenantCtx.TenantID().Value(),
			"tenant_name":  tenantCtx.TenantName(),
			"subscription": tenantCtx.Subscription().String(),
			"endpoint":     c.Request.URL.Path,
			"method":       c.Request.Method,
		})

		c.Next()
	}
}

// writeJSONErrorResponse writes a JSON error response for HTTP handlers
func writeJSONErrorResponse(w http.ResponseWriter, statusCode int, error, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]string{
		"error":   error,
		"message": message,
	}

	json.NewEncoder(w).Encode(response)
}
