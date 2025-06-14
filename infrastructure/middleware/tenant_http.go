package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/DimaJoyti/go-coffee/domain/shared"
	"github.com/DimaJoyti/go-coffee/domain/tenant"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	tenantMetricsKey contextKey = "tenant_metrics"
)

// HTTPTenantMiddleware provides HTTP-specific tenant middleware functions
type HTTPTenantMiddleware struct {
	tenantMiddleware *TenantContextMiddleware
	jwtExtractor     *TenantJWTExtractor
}

// NewHTTPTenantMiddleware creates a new HTTP tenant middleware
func NewHTTPTenantMiddleware(
	tenantRepository tenant.TenantRepository,
	jwtExtractor *TenantJWTExtractor,
) *HTTPTenantMiddleware {
	return &HTTPTenantMiddleware{
		tenantMiddleware: NewTenantContextMiddleware(tenantRepository),
		jwtExtractor:     jwtExtractor,
	}
}

// TenantContext returns an HTTP middleware that extracts and validates tenant context
func (m *HTTPTenantMiddleware) TenantContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantCtx, err := m.tenantMiddleware.extractTenantContext(r)
		if err != nil {
			m.writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized", err.Error())
			return
		}

		// Add tenant context to request context
		ctx := shared.WithTenantContext(r.Context(), tenantCtx)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// RequireFeature returns an HTTP middleware that requires a specific feature
func (m *HTTPTenantMiddleware) RequireFeature(feature string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantCtx, err := shared.FromContext(r.Context())
			if err != nil {
				m.writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized", "Tenant context not found")
				return
			}

			if !tenantCtx.HasFeature(feature) {
				response := map[string]string{
					"error":   "Forbidden",
					"message": "Feature not available for current subscription",
					"feature": feature,
				}
				m.writeJSONResponse(w, http.StatusForbidden, response)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireSubscription returns an HTTP middleware that requires a specific subscription plan
func (m *HTTPTenantMiddleware) RequireSubscription(plan shared.SubscriptionPlan) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantCtx, err := shared.FromContext(r.Context())
			if err != nil {
				m.writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized", "Tenant context not found")
				return
			}

			if tenantCtx.Subscription() != plan {
				response := map[string]string{
					"error":         "Forbidden",
					"message":       "Subscription plan required",
					"required_plan": plan.String(),
					"current_plan":  tenantCtx.Subscription().String(),
				}
				m.writeJSONResponse(w, http.StatusForbidden, response)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ValidateTenantAccess validates that the request has access to the specified tenant resource
func (m *HTTPTenantMiddleware) ValidateTenantAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantCtx, err := shared.FromContext(r.Context())
		if err != nil {
			m.writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized", "Tenant context not found")
			return
		}

		// Extract tenant ID from URL path
		urlTenantID := m.extractTenantIDFromURL(r)
		if urlTenantID != "" && urlTenantID != tenantCtx.TenantID().Value() {
			m.writeErrorResponse(w, http.StatusForbidden, "Forbidden", "Access denied to tenant resource")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// JWTTenantContext returns an HTTP middleware that extracts tenant context from JWT
func (m *HTTPTenantMiddleware) JWTTenantContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.jwtExtractor == nil {
			m.writeErrorResponse(w, http.StatusInternalServerError, "Internal Server Error", "JWT extractor not configured")
			return
		}

		// Extract tenant ID from JWT
		tenantID, err := m.jwtExtractor.ExtractTenantFromJWT(r)
		if err != nil {
			m.writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized", err.Error())
			return
		}

		// Load tenant information
		tenant, err := m.tenantMiddleware.tenantRepository.FindByTenantID(r.Context(), tenantID)
		if err != nil || tenant == nil {
			m.writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized", "Tenant not found")
			return
		}

		// Check if tenant is active
		if !tenant.IsActive() {
			m.writeErrorResponse(w, http.StatusForbidden, "Forbidden", "Tenant is not active")
			return
		}

		// Create tenant context
		tenantCtx := shared.NewTenantContext(
			tenantID,
			tenant.Name(),
			tenant.Subscription().Plan(),
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

		// Add tenant context to request context
		ctx := shared.WithTenantContext(r.Context(), tenantCtx)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// GetHTTPTenantContext extracts tenant context from HTTP request context
func GetHTTPTenantContext(r *http.Request) (*shared.TenantContext, error) {
	return shared.FromContext(r.Context())
}

// GetHTTPTenantID extracts tenant ID from HTTP request context
func GetHTTPTenantID(r *http.Request) (shared.TenantID, error) {
	tenantCtx, err := GetHTTPTenantContext(r)
	if err != nil {
		return shared.TenantID{}, err
	}

	return tenantCtx.TenantID(), nil
}

// extractTenantIDFromURL extracts tenant ID from URL path parameters
func (m *HTTPTenantMiddleware) extractTenantIDFromURL(r *http.Request) string {
	// This is a simple implementation - in a real application,
	// you might use a router that provides path parameters
	// Look for patterns like /api/v1/tenants/{tenant_id}/...
	// or /tenants/{tenant_id}/...
	return m.tenantMiddleware.extractFromPath(r)
}

// writeErrorResponse writes a JSON error response
func (m *HTTPTenantMiddleware) writeErrorResponse(w http.ResponseWriter, statusCode int, error, message string) {
	response := map[string]string{
		"error":   error,
		"message": message,
	}
	m.writeJSONResponse(w, statusCode, response)
}

// writeJSONResponse writes a JSON response
func (m *HTTPTenantMiddleware) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// TenantMetricsHandler provides HTTP handlers for tenant metrics
type TenantMetricsHandler struct{}

// NewTenantMetricsHandler creates a new tenant metrics handler
func NewTenantMetricsHandler() *TenantMetricsHandler {
	return &TenantMetricsHandler{}
}

// CollectMetrics returns an HTTP middleware that collects tenant-specific metrics
func (h *TenantMetricsHandler) CollectMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantCtx, err := GetHTTPTenantContext(r)
		if err != nil {
			// Continue without metrics if tenant context is not available
			next.ServeHTTP(w, r)
			return
		}

		// Record request metrics per tenant
		// In a real implementation, you would send these to a metrics system
		metrics := map[string]interface{}{
			"tenant_id":    tenantCtx.TenantID().Value(),
			"tenant_name":  tenantCtx.TenantName(),
			"subscription": tenantCtx.Subscription().String(),
			"endpoint":     r.URL.Path,
			"method":       r.Method,
		}

		// Add metrics to request context for potential use by handlers
		ctx := r.Context()
		ctx = context.WithValue(ctx, tenantMetricsKey, metrics)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
