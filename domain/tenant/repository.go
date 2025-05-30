package tenant

import (
	"context"
	"errors"

	"github.com/DimaJoyti/go-coffee/domain/shared"
)

// TenantRepository defines the interface for tenant persistence
type TenantRepository interface {
	// Save saves a tenant aggregate
	Save(ctx context.Context, tenant *Tenant) error

	// FindByID finds a tenant by ID
	FindByID(ctx context.Context, id shared.AggregateID) (*Tenant, error)
	
	// FindByTenantID finds a tenant by tenant ID (for self-reference)
	FindByTenantID(ctx context.Context, tenantID shared.TenantID) (*Tenant, error)
	
	// FindByEmail finds a tenant by email
	FindByEmail(ctx context.Context, email shared.Email) (*Tenant, error)
	
	// FindByOwnerID finds tenants by owner ID
	FindByOwnerID(ctx context.Context, ownerID shared.AggregateID) ([]*Tenant, error)
	
	// FindByStatus finds tenants by status
	FindByStatus(ctx context.Context, status TenantStatus) ([]*Tenant, error)
	
	// FindBySubscriptionPlan finds tenants by subscription plan
	FindBySubscriptionPlan(ctx context.Context, plan shared.SubscriptionPlan) ([]*Tenant, error)
	
	// ExistsByEmail checks if a tenant exists with the given email
	ExistsByEmail(ctx context.Context, email shared.Email) (bool, error)
	
	// Delete removes a tenant
	Delete(ctx context.Context, id shared.AggregateID) error
	
	// FindAll finds all tenants with pagination
	FindAll(ctx context.Context, offset, limit int) ([]*Tenant, int64, error)
	
	// FindActiveTenants finds all active tenants
	FindActiveTenants(ctx context.Context) ([]*Tenant, error)
	
	// CountByStatus counts tenants by status
	CountByStatus(ctx context.Context, status TenantStatus) (int64, error)
	
	// FindExpiredSubscriptions finds tenants with expired subscriptions
	FindExpiredSubscriptions(ctx context.Context) ([]*Tenant, error)
}

// TenantSpecification defines specifications for tenant queries
type TenantSpecification interface {
	shared.Specification
}

// ActiveTenantSpecification checks if a tenant is active
type ActiveTenantSpecification struct {
	*shared.BaseSpecification
}

// NewActiveTenantSpecification creates a new active tenant specification
func NewActiveTenantSpecification() *ActiveTenantSpecification {
	return &ActiveTenantSpecification{
		BaseSpecification: &shared.BaseSpecification{},
	}
}

// IsSatisfiedBy checks if the tenant is active
func (spec *ActiveTenantSpecification) IsSatisfiedBy(candidate interface{}) bool {
	tenant, ok := candidate.(*Tenant)
	if !ok {
		return false
	}
	return tenant.IsActive()
}

// SubscriptionPlanSpecification checks if a tenant has a specific subscription plan
type SubscriptionPlanSpecification struct {
	*shared.BaseSpecification
	plan shared.SubscriptionPlan
}

// NewSubscriptionPlanSpecification creates a new subscription plan specification
func NewSubscriptionPlanSpecification(plan shared.SubscriptionPlan) *SubscriptionPlanSpecification {
	return &SubscriptionPlanSpecification{
		BaseSpecification: &shared.BaseSpecification{},
		plan:              plan,
	}
}

// IsSatisfiedBy checks if the tenant has the specified subscription plan
func (spec *SubscriptionPlanSpecification) IsSatisfiedBy(candidate interface{}) bool {
	tenant, ok := candidate.(*Tenant)
	if !ok {
		return false
	}
	if tenant.subscription == nil {
		return false
	}
	return tenant.subscription.plan == spec.plan
}

// HasFeatureSpecification checks if a tenant has a specific feature
type HasFeatureSpecification struct {
	*shared.BaseSpecification
	feature string
}

// NewHasFeatureSpecification creates a new feature specification
func NewHasFeatureSpecification(feature string) *HasFeatureSpecification {
	return &HasFeatureSpecification{
		BaseSpecification: &shared.BaseSpecification{},
		feature:           feature,
	}
}

// IsSatisfiedBy checks if the tenant has the specified feature
func (spec *HasFeatureSpecification) IsSatisfiedBy(candidate interface{}) bool {
	tenant, ok := candidate.(*Tenant)
	if !ok {
		return false
	}
	return tenant.HasFeature(spec.feature)
}

// TenantTypeSpecification checks if a tenant has a specific type
type TenantTypeSpecification struct {
	*shared.BaseSpecification
	tenantType TenantType
}

// NewTenantTypeSpecification creates a new tenant type specification
func NewTenantTypeSpecification(tenantType TenantType) *TenantTypeSpecification {
	return &TenantTypeSpecification{
		BaseSpecification: &shared.BaseSpecification{},
		tenantType:        tenantType,
	}
}

// IsSatisfiedBy checks if the tenant has the specified type
func (spec *TenantTypeSpecification) IsSatisfiedBy(candidate interface{}) bool {
	tenant, ok := candidate.(*Tenant)
	if !ok {
		return false
	}
	return tenant.tenantType == spec.tenantType
}

// LocationCountSpecification checks if a tenant has a specific number of locations
type LocationCountSpecification struct {
	*shared.BaseSpecification
	minCount int
	maxCount int
}

// NewLocationCountSpecification creates a new location count specification
func NewLocationCountSpecification(minCount, maxCount int) *LocationCountSpecification {
	return &LocationCountSpecification{
		BaseSpecification: &shared.BaseSpecification{},
		minCount:          minCount,
		maxCount:          maxCount,
	}
}

// IsSatisfiedBy checks if the tenant has the specified number of locations
func (spec *LocationCountSpecification) IsSatisfiedBy(candidate interface{}) bool {
	tenant, ok := candidate.(*Tenant)
	if !ok {
		return false
	}
	
	locationCount := len(tenant.locations)
	return locationCount >= spec.minCount && (spec.maxCount == -1 || locationCount <= spec.maxCount)
}

// TenantQueryBuilder helps build complex tenant queries
type TenantQueryBuilder struct {
	specifications []TenantSpecification
}

// NewTenantQueryBuilder creates a new tenant query builder
func NewTenantQueryBuilder() *TenantQueryBuilder {
	return &TenantQueryBuilder{
		specifications: make([]TenantSpecification, 0),
	}
}

// WithActiveStatus adds active status specification
func (qb *TenantQueryBuilder) WithActiveStatus() *TenantQueryBuilder {
	qb.specifications = append(qb.specifications, NewActiveTenantSpecification())
	return qb
}

// WithSubscriptionPlan adds subscription plan specification
func (qb *TenantQueryBuilder) WithSubscriptionPlan(plan shared.SubscriptionPlan) *TenantQueryBuilder {
	qb.specifications = append(qb.specifications, NewSubscriptionPlanSpecification(plan))
	return qb
}

// WithFeature adds feature specification
func (qb *TenantQueryBuilder) WithFeature(feature string) *TenantQueryBuilder {
	qb.specifications = append(qb.specifications, NewHasFeatureSpecification(feature))
	return qb
}

// WithTenantType adds tenant type specification
func (qb *TenantQueryBuilder) WithTenantType(tenantType TenantType) *TenantQueryBuilder {
	qb.specifications = append(qb.specifications, NewTenantTypeSpecification(tenantType))
	return qb
}

// WithLocationCount adds location count specification
func (qb *TenantQueryBuilder) WithLocationCount(minCount, maxCount int) *TenantQueryBuilder {
	qb.specifications = append(qb.specifications, NewLocationCountSpecification(minCount, maxCount))
	return qb
}

// Build builds the final specification
func (qb *TenantQueryBuilder) Build() TenantSpecification {
	if len(qb.specifications) == 0 {
		return &AlwaysTrueSpecification{}
	}
	
	if len(qb.specifications) == 1 {
		return qb.specifications[0]
	}
	
	// Combine all specifications with AND logic
	result := qb.specifications[0]
	for i := 1; i < len(qb.specifications); i++ {
		result = result.And(qb.specifications[i])
	}
	
	return result
}

// AlwaysTrueSpecification always returns true
type AlwaysTrueSpecification struct {
	*shared.BaseSpecification
}

// IsSatisfiedBy always returns true
func (spec *AlwaysTrueSpecification) IsSatisfiedBy(candidate interface{}) bool {
	return true
}

// TenantDomainService provides domain services for tenant operations
type TenantDomainService struct {
	repository TenantRepository
}

// NewTenantDomainService creates a new tenant domain service
func NewTenantDomainService(repository TenantRepository) *TenantDomainService {
	return &TenantDomainService{
		repository: repository,
	}
}

// ValidateUniqueness validates that tenant email is unique
func (s *TenantDomainService) ValidateUniqueness(ctx context.Context, email shared.Email, excludeID *shared.AggregateID) error {
	exists, err := s.repository.ExistsByEmail(ctx, email)
	if err != nil {
		return err
	}
	
	if exists {
		// If we're updating an existing tenant, check if it's the same tenant
		if excludeID != nil {
			existingTenant, err := s.repository.FindByEmail(ctx, email)
			if err != nil {
				return err
			}
			if existingTenant != nil && existingTenant.ID().Equals(*excludeID) {
				return nil // Same tenant, uniqueness is maintained
			}
		}
		return errors.New("tenant with this email already exists")
	}
	
	return nil
}

// CanUpgradeSubscription checks if a tenant can upgrade to a specific plan
func (s *TenantDomainService) CanUpgradeSubscription(ctx context.Context, tenantID shared.TenantID, newPlan shared.SubscriptionPlan) error {
	tenant, err := s.repository.FindByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}
	
	if tenant == nil {
		return errors.New("tenant not found")
	}
	
	if !tenant.IsActive() {
		return errors.New("only active tenants can upgrade subscription")
	}
	
	if tenant.subscription != nil {
		currentPlan := tenant.subscription.plan
		
		// Define upgrade paths
		upgradeMatrix := map[shared.SubscriptionPlan][]shared.SubscriptionPlan{
			shared.SubscriptionBasic: {
				shared.SubscriptionProfessional,
				shared.SubscriptionEnterprise,
			},
			shared.SubscriptionProfessional: {
				shared.SubscriptionEnterprise,
			},
			shared.SubscriptionEnterprise: {}, // No upgrades from enterprise
		}
		
		allowedUpgrades, exists := upgradeMatrix[currentPlan]
		if !exists {
			return errors.New("invalid current subscription plan")
		}
		
		for _, allowedPlan := range allowedUpgrades {
			if allowedPlan == newPlan {
				return nil // Upgrade is allowed
			}
		}
		
		return errors.New("upgrade to specified plan is not allowed")
	}
	
	return nil // No current subscription, any plan is allowed
}

// GetTenantsBySpecification finds tenants matching a specification
func (s *TenantDomainService) GetTenantsBySpecification(ctx context.Context, spec TenantSpecification) ([]*Tenant, error) {
	// This would typically be implemented with database queries
	// For now, we'll get all tenants and filter (not efficient for production)
	allTenants, _, err := s.repository.FindAll(ctx, 0, 1000)
	if err != nil {
		return nil, err
	}
	
	filteredTenants := make([]*Tenant, 0)
	for _, tenant := range allTenants {
		if spec.IsSatisfiedBy(tenant) {
			filteredTenants = append(filteredTenants, tenant)
		}
	}
	
	return filteredTenants, nil
}

// CalculateSubscriptionMetrics calculates metrics for subscription analysis
func (s *TenantDomainService) CalculateSubscriptionMetrics(ctx context.Context) (*SubscriptionMetrics, error) {
	basicCount, err := s.repository.CountByStatus(ctx, TenantStatusActive)
	if err != nil {
		return nil, err
	}
	
	// This is simplified - in reality, you'd have separate counts per plan
	metrics := &SubscriptionMetrics{
		TotalActiveTenants: basicCount,
		BasicPlanCount:     0,
		ProfessionalCount:  0,
		EnterpriseCount:    0,
	}
	
	return metrics, nil
}

// SubscriptionMetrics holds subscription-related metrics
type SubscriptionMetrics struct {
	TotalActiveTenants int64
	BasicPlanCount     int64
	ProfessionalCount  int64
	EnterpriseCount    int64
	ChurnRate          float64
	AverageRevenue     float64
}
