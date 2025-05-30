package tenant

import (
	"errors"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/domain/shared"
)

// TenantStatus represents the status of a tenant
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusInactive  TenantStatus = "inactive"
	TenantStatusPending   TenantStatus = "pending"
)

// IsValid checks if the tenant status is valid
func (ts TenantStatus) IsValid() bool {
	switch ts {
	case TenantStatusActive, TenantStatusSuspended, TenantStatusInactive, TenantStatusPending:
		return true
	default:
		return false
	}
}

// String returns the string representation of tenant status
func (ts TenantStatus) String() string {
	return string(ts)
}

// TenantType represents the type of tenant
type TenantType string

const (
	TenantTypeRestaurant TenantType = "restaurant"
	TenantTypeCafe       TenantType = "cafe"
	TenantTypeChain      TenantType = "chain"
	TenantTypeFranchise  TenantType = "franchise"
)

// IsValid checks if the tenant type is valid
func (tt TenantType) IsValid() bool {
	switch tt {
	case TenantTypeRestaurant, TenantTypeCafe, TenantTypeChain, TenantTypeFranchise:
		return true
	default:
		return false
	}
}

// String returns the string representation of tenant type
func (tt TenantType) String() string {
	return string(tt)
}

// TenantSettings holds tenant-specific settings
type TenantSettings struct {
	timezone             string
	currency             string
	language             string
	dateFormat           string
	timeFormat           string
	businessHours        map[string]BusinessHours
	notificationSettings NotificationSettings
	aiSettings           AISettings
}

// NewTenantSettings creates new tenant settings with defaults
func NewTenantSettings() *TenantSettings {
	return &TenantSettings{
		timezone:             "UTC",
		currency:             "USD",
		language:             "en",
		dateFormat:           "2006-01-02",
		timeFormat:           "15:04",
		businessHours:        make(map[string]BusinessHours),
		notificationSettings: NewDefaultNotificationSettings(),
		aiSettings:           NewDefaultAISettings(),
	}
}

// BusinessHours represents business hours for a day
type BusinessHours struct {
	isOpen    bool
	openTime  time.Time
	closeTime time.Time
	breaks    []TimeRange
}

// TimeRange represents a time range
type TimeRange struct {
	start time.Time
	end   time.Time
}

// NotificationSettings holds notification preferences
type NotificationSettings struct {
	emailEnabled  bool
	smsEnabled    bool
	pushEnabled   bool
	webhookURL    string
	slackWebhook  string
}

// NewDefaultNotificationSettings creates default notification settings
func NewDefaultNotificationSettings() NotificationSettings {
	return NotificationSettings{
		emailEnabled: true,
		smsEnabled:   false,
		pushEnabled:  true,
	}
}

// AISettings holds AI-specific settings for the tenant
type AISettings struct {
	aiEnabled                bool
	recommendationsEnabled   bool
	predictiveAnalytics      bool
	customModelID           string
	aiModelParameters       map[string]interface{}
	dataRetentionDays       int
	privacyLevel            string
}

// NewDefaultAISettings creates default AI settings
func NewDefaultAISettings() AISettings {
	return AISettings{
		aiEnabled:              true,
		recommendationsEnabled: true,
		predictiveAnalytics:    false,
		aiModelParameters:      make(map[string]interface{}),
		dataRetentionDays:      365,
		privacyLevel:           "standard",
	}
}

// NewSubscription creates a new subscription
func NewSubscription(
	plan shared.SubscriptionPlan,
	status SubscriptionStatus,
	startDate time.Time,
	billingCycle BillingCycle,
	maxUsers int,
	maxLocations int,
	maxOrdersPerMonth int,
) *Subscription {
	return &Subscription{
		plan:              plan,
		status:            status,
		startDate:         startDate,
		billingCycle:      billingCycle,
		maxUsers:          maxUsers,
		maxLocations:      maxLocations,
		maxOrdersPerMonth: maxOrdersPerMonth,
		features:          plan.GetFeatures(),
	}
}

// Plan returns the subscription plan
func (s *Subscription) Plan() shared.SubscriptionPlan {
	return s.plan
}

// Status returns the subscription status
func (s *Subscription) Status() SubscriptionStatus {
	return s.status
}

// StartDate returns the subscription start date
func (s *Subscription) StartDate() time.Time {
	return s.startDate
}

// BillingCycle returns the billing cycle
func (s *Subscription) BillingCycle() BillingCycle {
	return s.billingCycle
}

// MaxUsers returns the maximum number of users
func (s *Subscription) MaxUsers() int {
	return s.maxUsers
}

// MaxLocations returns the maximum number of locations
func (s *Subscription) MaxLocations() int {
	return s.maxLocations
}

// MaxOrdersPerMonth returns the maximum orders per month
func (s *Subscription) MaxOrdersPerMonth() int {
	return s.maxOrdersPerMonth
}

// Features returns the subscription features
func (s *Subscription) Features() map[string]bool {
	return s.features
}

// Subscription represents a tenant's subscription
type Subscription struct {
	plan              shared.SubscriptionPlan
	status            SubscriptionStatus
	startDate         time.Time
	endDate           *time.Time
	billingCycle      BillingCycle
	maxUsers          int
	maxLocations      int
	maxOrdersPerMonth int
	features          map[string]bool
}

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusExpired  SubscriptionStatus = "expired"
	SubscriptionStatusCanceled SubscriptionStatus = "canceled"
	SubscriptionStatusTrial    SubscriptionStatus = "trial"
)

// BillingCycle represents the billing cycle
type BillingCycle string

const (
	BillingCycleMonthly BillingCycle = "monthly"
	BillingCycleYearly  BillingCycle = "yearly"
)

// Location represents a tenant location
type Location struct {
	*shared.Entity
	name        string
	address     shared.Address
	phoneNumber shared.PhoneNumber
	email       shared.Email
	isActive    bool
	settings    *LocationSettings
}

// LocationSettings holds location-specific settings
type LocationSettings struct {
	kitchenCapacity    int
	maxConcurrentOrders int
	averageServiceTime  time.Duration
	specializations    []string
}

// Tenant represents the tenant aggregate root
type Tenant struct {
	*shared.BaseAggregate
	name         string
	tenantType   TenantType
	status       TenantStatus
	email        shared.Email
	phoneNumber  shared.PhoneNumber
	address      shared.Address
	settings     *TenantSettings
	subscription *Subscription
	locations    map[shared.AggregateID]*Location
	ownerID      shared.AggregateID
}

// NewTenant creates a new tenant aggregate
func NewTenant(
	id shared.AggregateID,
	tenantID shared.TenantID,
	name string,
	tenantType TenantType,
	email shared.Email,
	phoneNumber shared.PhoneNumber,
	address shared.Address,
	ownerID shared.AggregateID,
) (*Tenant, error) {
	if name == "" {
		return nil, errors.New("tenant name cannot be empty")
	}
	
	if !tenantType.IsValid() {
		return nil, errors.New("invalid tenant type")
	}
	
	tenant := &Tenant{
		BaseAggregate: shared.NewBaseAggregate(id, tenantID),
		name:          name,
		tenantType:    tenantType,
		status:        TenantStatusPending,
		email:         email,
		phoneNumber:   phoneNumber,
		address:       address,
		settings:      NewTenantSettings(),
		locations:     make(map[shared.AggregateID]*Location),
		ownerID:       ownerID,
	}
	
	// Add domain event
	event := NewTenantCreatedEvent(tenant)
	tenant.AddDomainEvent(event)
	
	return tenant, nil
}

// Name returns the tenant name
func (t *Tenant) Name() string {
	return t.name
}

// TenantType returns the tenant type
func (t *Tenant) TenantType() TenantType {
	return t.tenantType
}

// Status returns the tenant status
func (t *Tenant) Status() TenantStatus {
	return t.status
}

// Email returns the tenant email
func (t *Tenant) Email() shared.Email {
	return t.email
}

// PhoneNumber returns the tenant phone number
func (t *Tenant) PhoneNumber() shared.PhoneNumber {
	return t.phoneNumber
}

// Address returns the tenant address
func (t *Tenant) Address() shared.Address {
	return t.address
}

// Settings returns the tenant settings
func (t *Tenant) Settings() *TenantSettings {
	return t.settings
}

// Subscription returns the tenant subscription
func (t *Tenant) Subscription() *Subscription {
	return t.subscription
}

// Locations returns all tenant locations
func (t *Tenant) Locations() map[shared.AggregateID]*Location {
	return t.locations
}

// OwnerID returns the owner ID
func (t *Tenant) OwnerID() shared.AggregateID {
	return t.ownerID
}

// Activate activates the tenant
func (t *Tenant) Activate() error {
	if t.status == TenantStatusActive {
		return errors.New("tenant is already active")
	}
	
	t.status = TenantStatusActive
	t.IncrementVersion()
	
	// Add domain event
	event := NewTenantActivatedEvent(t)
	t.AddDomainEvent(event)
	
	return nil
}

// Suspend suspends the tenant
func (t *Tenant) Suspend(reason string) error {
	if t.status != TenantStatusActive {
		return errors.New("only active tenants can be suspended")
	}
	
	t.status = TenantStatusSuspended
	t.IncrementVersion()
	
	// Add domain event
	event := NewTenantSuspendedEvent(t, reason)
	t.AddDomainEvent(event)
	
	return nil
}

// Deactivate deactivates the tenant
func (t *Tenant) Deactivate() error {
	if t.status == TenantStatusInactive {
		return errors.New("tenant is already inactive")
	}
	
	t.status = TenantStatusInactive
	t.IncrementVersion()
	
	// Add domain event
	event := NewTenantDeactivatedEvent(t)
	t.AddDomainEvent(event)
	
	return nil
}

// UpdateSubscription updates the tenant subscription
func (t *Tenant) UpdateSubscription(subscription *Subscription) error {
	if subscription == nil {
		return errors.New("subscription cannot be nil")
	}
	
	oldSubscription := t.subscription
	t.subscription = subscription
	t.IncrementVersion()
	
	// Update features based on subscription plan
	t.updateFeaturesFromSubscription()
	
	// Add domain event
	event := NewTenantSubscriptionUpdatedEvent(t, oldSubscription, subscription)
	t.AddDomainEvent(event)
	
	return nil
}

// AddLocation adds a new location to the tenant
func (t *Tenant) AddLocation(
	locationID shared.AggregateID,
	name string,
	address shared.Address,
	phoneNumber shared.PhoneNumber,
	email shared.Email,
) error {
	if _, exists := t.locations[locationID]; exists {
		return errors.New("location already exists")
	}
	
	// Check subscription limits
	if t.subscription != nil && len(t.locations) >= t.subscription.maxLocations {
		return errors.New("maximum number of locations reached for current subscription")
	}
	
	location := &Location{
		Entity:      shared.NewEntity(locationID, t.GetTenantID()),
		name:        name,
		address:     address,
		phoneNumber: phoneNumber,
		email:       email,
		isActive:    true,
		settings:    &LocationSettings{},
	}
	
	t.locations[locationID] = location
	t.IncrementVersion()
	
	// Add domain event
	event := NewLocationAddedEvent(t, location)
	t.AddDomainEvent(event)
	
	return nil
}

// RemoveLocation removes a location from the tenant
func (t *Tenant) RemoveLocation(locationID shared.AggregateID) error {
	location, exists := t.locations[locationID]
	if !exists {
		return errors.New("location not found")
	}
	
	delete(t.locations, locationID)
	t.IncrementVersion()
	
	// Add domain event
	event := NewLocationRemovedEvent(t, location)
	t.AddDomainEvent(event)
	
	return nil
}

// UpdateSettings updates tenant settings
func (t *Tenant) UpdateSettings(settings *TenantSettings) error {
	if settings == nil {
		return errors.New("settings cannot be nil")
	}
	
	oldSettings := t.settings
	t.settings = settings
	t.IncrementVersion()
	
	// Add domain event
	event := NewTenantSettingsUpdatedEvent(t, oldSettings, settings)
	t.AddDomainEvent(event)
	
	return nil
}

// IsActive checks if the tenant is active
func (t *Tenant) IsActive() bool {
	return t.status == TenantStatusActive
}

// HasFeature checks if the tenant has a specific feature
func (t *Tenant) HasFeature(feature string) bool {
	if t.subscription == nil {
		return false
	}
	
	enabled, exists := t.subscription.features[feature]
	return exists && enabled
}

// CanCreateOrder checks if the tenant can create orders
func (t *Tenant) CanCreateOrder() bool {
	return t.IsActive() && t.HasFeature("basic_orders")
}

// updateFeaturesFromSubscription updates features based on subscription plan
func (t *Tenant) updateFeaturesFromSubscription() {
	if t.subscription == nil {
		return
	}
	
	t.subscription.features = t.subscription.plan.GetFeatures()
}

// GetLocationByID returns a location by ID
func (t *Tenant) GetLocationByID(locationID shared.AggregateID) (*Location, error) {
	location, exists := t.locations[locationID]
	if !exists {
		return nil, fmt.Errorf("location with ID %s not found", locationID.Value())
	}
	return location, nil
}

// GetActiveLocations returns all active locations
func (t *Tenant) GetActiveLocations() []*Location {
	activeLocations := make([]*Location, 0)
	for _, location := range t.locations {
		if location.isActive {
			activeLocations = append(activeLocations, location)
		}
	}
	return activeLocations
}
