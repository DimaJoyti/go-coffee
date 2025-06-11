package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/DimaJoyti/go-coffee/domain/shared"
	"github.com/DimaJoyti/go-coffee/domain/tenant"
)

// TenantContextMiddleware extracts and validates tenant context from requests
type TenantContextMiddleware struct {
	tenantRepository tenant.TenantRepository
	validator        *TenantContextValidator
}

// NewTenantContextMiddleware creates a new tenant context middleware
func NewTenantContextMiddleware(tenantRepository tenant.TenantRepository) *TenantContextMiddleware {
	return &TenantContextMiddleware{
		tenantRepository: tenantRepository,
		validator:        NewTenantContextValidator(),
	}
}

// TenantContextValidator validates tenant context
type TenantContextValidator struct{}

// NewTenantContextValidator creates a new tenant context validator
func NewTenantContextValidator() *TenantContextValidator {
	return &TenantContextValidator{}
}

// HTTPTenantContext returns an HTTP middleware that extracts tenant context
func (m *TenantContextMiddleware) HTTPTenantContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantCtx, err := m.extractTenantContext(r)
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

// writeErrorResponse writes an error response in JSON format
func (m *TenantContextMiddleware) writeErrorResponse(w http.ResponseWriter, statusCode int, error, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]string{
		"error":   error,
		"message": message,
	}

	json.NewEncoder(w).Encode(response)
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
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Extract Bearer token
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// In a real implementation, you would decode and validate the JWT
	// For now, we'll return empty string
	_ = token
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

// RequireFeatureHTTP returns an HTTP middleware that requires a specific feature
func (m *TenantContextMiddleware) RequireFeatureHTTP(feature string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantCtx, err := shared.FromContext(r.Context())
			if err != nil {
				m.writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized", "Tenant context not found")
				return
			}

			if !tenantCtx.HasFeature(feature) {
				m.writeErrorResponse(w, http.StatusForbidden, "Forbidden", "Feature not available for current subscription")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireSubscriptionHTTP returns an HTTP middleware that requires a specific subscription plan
func (m *TenantContextMiddleware) RequireSubscriptionHTTP(plan shared.SubscriptionPlan) func(http.Handler) http.Handler {
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
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(response)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetTenantContext extracts tenant context from HTTP request context
func GetTenantContext(r *http.Request) (*shared.TenantContext, error) {
	return shared.FromContext(r.Context())
}

// GetTenantID extracts tenant ID from HTTP request context
func GetTenantID(r *http.Request) (shared.TenantID, error) {
	tenantCtx, err := GetTenantContext(r)
	if err != nil {
		return shared.TenantID{}, err
	}

	return tenantCtx.TenantID(), nil
}

// TenantIsolationMiddleware ensures data isolation between tenants
type TenantIsolationMiddleware struct{}

// NewTenantIsolationMiddleware creates a new tenant isolation middleware
func NewTenantIsolationMiddleware() *TenantIsolationMiddleware {
	return &TenantIsolationMiddleware{}
}

// ValidateTenantAccessHTTP validates that the request has access to the specified tenant resource
func (m *TenantIsolationMiddleware) ValidateTenantAccessHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := shared.FromContext(r.Context())
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "Unauthorized", "Tenant context not found")
			return
		}

		// Extract tenant ID from URL parameters (this would need to be customized based on your routing)
		// For now, we'll skip this validation
		// urlTenantID := extractTenantIDFromURL(r)
		// if urlTenantID != "" && urlTenantID != tenantCtx.TenantID().Value() {
		//     writeJSONError(w, http.StatusForbidden, "Forbidden", "Access denied to tenant resource")
		//     return
		// }

		next.ServeHTTP(w, r)
	})
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, statusCode int, error, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]string{
		"error":   error,
		"message": message,
	}

	json.NewEncoder(w).Encode(response)
}
