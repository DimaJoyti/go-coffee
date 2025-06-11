package shared

import (
	"context"
	"errors"
	"fmt"
)

// TenantID represents a unique tenant identifier
type TenantID struct {
	value string
}

// NewTenantID creates a new TenantID value object
func NewTenantID(value string) (TenantID, error) {
	if value == "" {
		return TenantID{}, errors.New("tenant ID cannot be empty")
	}
	if len(value) < 3 || len(value) > 50 {
		return TenantID{}, errors.New("tenant ID must be between 3 and 50 characters")
	}
	return TenantID{value: value}, nil
}

// Value returns the string value of TenantID
func (t TenantID) Value() string {
	return t.value
}

// String implements the Stringer interface
func (t TenantID) String() string {
	return t.value
}

// Equals checks if two TenantIDs are equal
func (t TenantID) Equals(other TenantID) bool {
	return t.value == other.value
}

// IsEmpty checks if TenantID is empty
func (t TenantID) IsEmpty() bool {
	return t.value == ""
}

// TenantContext holds tenant-specific information for request processing
type TenantContext struct {
	tenantID     TenantID
	tenantName   string
	subscription SubscriptionPlan
	features     map[string]bool
	metadata     map[string]interface{}
}

// NewTenantContext creates a new tenant context
func NewTenantContext(tenantID TenantID, tenantName string, subscription SubscriptionPlan) *TenantContext {
	return &TenantContext{
		tenantID:     tenantID,
		tenantName:   tenantName,
		subscription: subscription,
		features:     make(map[string]bool),
		metadata:     make(map[string]interface{}),
	}
}

// TenantID returns the tenant ID
func (tc *TenantContext) TenantID() TenantID {
	return tc.tenantID
}

// TenantName returns the tenant name
func (tc *TenantContext) TenantName() string {
	return tc.tenantName
}

// Subscription returns the subscription plan
func (tc *TenantContext) Subscription() SubscriptionPlan {
	return tc.subscription
}

// HasFeature checks if a feature is enabled for the tenant
func (tc *TenantContext) HasFeature(feature string) bool {
	enabled, exists := tc.features[feature]
	return exists && enabled
}

// EnableFeature enables a feature for the tenant
func (tc *TenantContext) EnableFeature(feature string) {
	tc.features[feature] = true
}

// DisableFeature disables a feature for the tenant
func (tc *TenantContext) DisableFeature(feature string) {
	tc.features[feature] = false
}

// SetMetadata sets metadata for the tenant context
func (tc *TenantContext) SetMetadata(key string, value interface{}) {
	tc.metadata[key] = value
}

// GetMetadata gets metadata from the tenant context
func (tc *TenantContext) GetMetadata(key string) (interface{}, bool) {
	value, exists := tc.metadata[key]
	return value, exists
}

// SubscriptionPlan represents different subscription tiers
type SubscriptionPlan string

const (
	SubscriptionBasic        SubscriptionPlan = "basic"
	SubscriptionProfessional SubscriptionPlan = "professional"
	SubscriptionEnterprise   SubscriptionPlan = "enterprise"
)

// IsValid checks if the subscription plan is valid
func (sp SubscriptionPlan) IsValid() bool {
	switch sp {
	case SubscriptionBasic, SubscriptionProfessional, SubscriptionEnterprise:
		return true
	default:
		return false
	}
}

// String returns the string representation of the subscription plan
func (sp SubscriptionPlan) String() string {
	return string(sp)
}

// GetFeatures returns the features available for the subscription plan
func (sp SubscriptionPlan) GetFeatures() map[string]bool {
	switch sp {
	case SubscriptionBasic:
		return map[string]bool{
			"basic_orders":       true,
			"ai_recommendations": false,
			"advanced_analytics": false,
			"multi_location":     false,
			"api_access":         false,
		}
	case SubscriptionProfessional:
		return map[string]bool{
			"basic_orders":       true,
			"ai_recommendations": true,
			"advanced_analytics": true,
			"multi_location":     true,
			"api_access":         true,
			"custom_branding":    false,
			"priority_support":   false,
		}
	case SubscriptionEnterprise:
		return map[string]bool{
			"basic_orders":        true,
			"ai_recommendations":  true,
			"advanced_analytics":  true,
			"multi_location":      true,
			"api_access":          true,
			"custom_branding":     true,
			"priority_support":    true,
			"white_label":         true,
			"custom_integrations": true,
		}
	default:
		return make(map[string]bool)
	}
}

// Context keys for tenant context
type contextKey string

const (
	TenantContextKey contextKey = "tenant_context"
)

// WithTenantContext adds tenant context to the Go context
func WithTenantContext(ctx context.Context, tenantCtx *TenantContext) context.Context {
	return context.WithValue(ctx, TenantContextKey, tenantCtx)
}

// FromContext extracts tenant context from Go context
func FromContext(ctx context.Context) (*TenantContext, error) {
	tenantCtx, ok := ctx.Value(TenantContextKey).(*TenantContext)
	if !ok {
		return nil, errors.New("tenant context not found in request context")
	}
	return tenantCtx, nil
}

// MustFromContext extracts tenant context from Go context or panics
func MustFromContext(ctx context.Context) *TenantContext {
	tenantCtx, err := FromContext(ctx)
	if err != nil {
		panic(fmt.Sprintf("tenant context is required: %v", err))
	}
	return tenantCtx
}

// TenantIsolationLevel defines the level of tenant isolation
type TenantIsolationLevel int

const (
	// SharedDatabase - all tenants share the same database with tenant_id column
	SharedDatabase TenantIsolationLevel = iota
	// SchemaPerTenant - each tenant has its own database schema
	SchemaPerTenant
	// DatabasePerTenant - each tenant has its own database
	DatabasePerTenant
)

// String returns the string representation of isolation level
func (til TenantIsolationLevel) String() string {
	switch til {
	case SharedDatabase:
		return "shared_database"
	case SchemaPerTenant:
		return "schema_per_tenant"
	case DatabasePerTenant:
		return "database_per_tenant"
	default:
		return "unknown"
	}
}

// IsValid checks if the isolation level is valid
func (til TenantIsolationLevel) IsValid() bool {
	switch til {
	case SharedDatabase, SchemaPerTenant, DatabasePerTenant:
		return true
	default:
		return false
	}
}

// TenantAware interface for entities that belong to a tenant
type TenantAware interface {
	GetTenantID() TenantID
	SetTenantID(TenantID)
}

// TenantConfiguration holds tenant-specific configuration
type TenantConfiguration struct {
	tenantID       TenantID
	isolationLevel TenantIsolationLevel
	databaseConfig map[string]string
	aiModelConfig  map[string]interface{}
	businessRules  map[string]interface{}
	integrations   map[string]bool
}

// NewTenantConfiguration creates a new tenant configuration
func NewTenantConfiguration(tenantID TenantID, isolationLevel TenantIsolationLevel) *TenantConfiguration {
	return &TenantConfiguration{
		tenantID:       tenantID,
		isolationLevel: isolationLevel,
		databaseConfig: make(map[string]string),
		aiModelConfig:  make(map[string]interface{}),
		businessRules:  make(map[string]interface{}),
		integrations:   make(map[string]bool),
	}
}

// TenantID returns the tenant ID
func (tc *TenantConfiguration) TenantID() TenantID {
	return tc.tenantID
}

// IsolationLevel returns the isolation level
func (tc *TenantConfiguration) IsolationLevel() TenantIsolationLevel {
	return tc.isolationLevel
}

// SetDatabaseConfig sets database configuration
func (tc *TenantConfiguration) SetDatabaseConfig(key, value string) {
	tc.databaseConfig[key] = value
}

// GetDatabaseConfig gets database configuration
func (tc *TenantConfiguration) GetDatabaseConfig(key string) (string, bool) {
	value, exists := tc.databaseConfig[key]
	return value, exists
}

// SetAIModelConfig sets AI model configuration
func (tc *TenantConfiguration) SetAIModelConfig(key string, value interface{}) {
	tc.aiModelConfig[key] = value
}

// GetAIModelConfig gets AI model configuration
func (tc *TenantConfiguration) GetAIModelConfig(key string) (interface{}, bool) {
	value, exists := tc.aiModelConfig[key]
	return value, exists
}

// SetBusinessRule sets a business rule
func (tc *TenantConfiguration) SetBusinessRule(key string, value interface{}) {
	tc.businessRules[key] = value
}

// GetBusinessRule gets a business rule
func (tc *TenantConfiguration) GetBusinessRule(key string) (interface{}, bool) {
	value, exists := tc.businessRules[key]
	return value, exists
}

// EnableIntegration enables an integration
func (tc *TenantConfiguration) EnableIntegration(integration string) {
	tc.integrations[integration] = true
}

// DisableIntegration disables an integration
func (tc *TenantConfiguration) DisableIntegration(integration string) {
	tc.integrations[integration] = false
}

// IsIntegrationEnabled checks if an integration is enabled
func (tc *TenantConfiguration) IsIntegrationEnabled(integration string) bool {
	enabled, exists := tc.integrations[integration]
	return exists && enabled
}
