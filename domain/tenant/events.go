package tenant

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/DimaJoyti/go-coffee/domain/shared"
)

// Event types for tenant domain
const (
	TenantCreatedEventType            = "tenant.created"
	TenantActivatedEventType          = "tenant.activated"
	TenantSuspendedEventType          = "tenant.suspended"
	TenantDeactivatedEventType        = "tenant.deactivated"
	TenantSubscriptionUpdatedEventType = "tenant.subscription_updated"
	TenantSettingsUpdatedEventType    = "tenant.settings_updated"
	LocationAddedEventType            = "tenant.location_added"
	LocationRemovedEventType          = "tenant.location_removed"
	LocationUpdatedEventType          = "tenant.location_updated"
)

// TenantCreatedEvent is raised when a new tenant is created
type TenantCreatedEvent struct {
	*shared.BaseDomainEvent
}

// NewTenantCreatedEvent creates a new TenantCreatedEvent
func NewTenantCreatedEvent(tenant *Tenant) *TenantCreatedEvent {
	data := map[string]interface{}{
		"tenant_name":    tenant.Name(),
		"tenant_type":    tenant.TenantType().String(),
		"email":          tenant.Email().Value(),
		"phone_number":   tenant.PhoneNumber().Value(),
		"address":        tenant.Address().String(),
		"owner_id":       tenant.OwnerID().Value(),
		"status":         tenant.Status().String(),
	}

	baseEvent := shared.NewBaseDomainEvent(
		uuid.New().String(),
		TenantCreatedEventType,
		tenant.ID(),
		tenant.GetTenantID(),
		data,
	)

	return &TenantCreatedEvent{
		BaseDomainEvent: baseEvent,
	}
}

// TenantActivatedEvent is raised when a tenant is activated
type TenantActivatedEvent struct {
	*shared.BaseDomainEvent
}

// NewTenantActivatedEvent creates a new TenantActivatedEvent
func NewTenantActivatedEvent(tenant *Tenant) *TenantActivatedEvent {
	data := map[string]interface{}{
		"tenant_name":     tenant.Name(),
		"previous_status": "pending", // Could be tracked better
		"new_status":      tenant.Status().String(),
	}

	baseEvent := shared.NewBaseDomainEvent(
		uuid.New().String(),
		TenantActivatedEventType,
		tenant.ID(),
		tenant.GetTenantID(),
		data,
	)

	return &TenantActivatedEvent{
		BaseDomainEvent: baseEvent,
	}
}

// TenantSuspendedEvent is raised when a tenant is suspended
type TenantSuspendedEvent struct {
	*shared.BaseDomainEvent
}

// NewTenantSuspendedEvent creates a new TenantSuspendedEvent
func NewTenantSuspendedEvent(tenant *Tenant, reason string) *TenantSuspendedEvent {
	data := map[string]interface{}{
		"tenant_name":     tenant.Name(),
		"previous_status": "active",
		"new_status":      tenant.Status().String(),
		"reason":          reason,
	}

	baseEvent := shared.NewBaseDomainEvent(
		uuid.New().String(),
		TenantSuspendedEventType,
		tenant.ID(),
		tenant.GetTenantID(),
		data,
	)

	return &TenantSuspendedEvent{
		BaseDomainEvent: baseEvent,
	}
}

// TenantDeactivatedEvent is raised when a tenant is deactivated
type TenantDeactivatedEvent struct {
	*shared.BaseDomainEvent
}

// NewTenantDeactivatedEvent creates a new TenantDeactivatedEvent
func NewTenantDeactivatedEvent(tenant *Tenant) *TenantDeactivatedEvent {
	data := map[string]interface{}{
		"tenant_name":     tenant.Name(),
		"previous_status": tenant.Status().String(),
		"new_status":      TenantStatusInactive.String(),
	}

	baseEvent := shared.NewBaseDomainEvent(
		uuid.New().String(),
		TenantDeactivatedEventType,
		tenant.ID(),
		tenant.GetTenantID(),
		data,
	)

	return &TenantDeactivatedEvent{
		BaseDomainEvent: baseEvent,
	}
}

// TenantSubscriptionUpdatedEvent is raised when a tenant's subscription is updated
type TenantSubscriptionUpdatedEvent struct {
	*shared.BaseDomainEvent
}

// NewTenantSubscriptionUpdatedEvent creates a new TenantSubscriptionUpdatedEvent
func NewTenantSubscriptionUpdatedEvent(tenant *Tenant, oldSubscription, newSubscription *Subscription) *TenantSubscriptionUpdatedEvent {
	data := map[string]interface{}{
		"tenant_name": tenant.Name(),
	}

	if oldSubscription != nil {
		data["old_plan"] = oldSubscription.plan.String()
		data["old_status"] = oldSubscription.status
	}

	if newSubscription != nil {
		data["new_plan"] = newSubscription.plan.String()
		data["new_status"] = newSubscription.status
		data["max_users"] = newSubscription.maxUsers
		data["max_locations"] = newSubscription.maxLocations
		data["max_orders_per_month"] = newSubscription.maxOrdersPerMonth
	}

	baseEvent := shared.NewBaseDomainEvent(
		uuid.New().String(),
		TenantSubscriptionUpdatedEventType,
		tenant.ID(),
		tenant.GetTenantID(),
		data,
	)

	return &TenantSubscriptionUpdatedEvent{
		BaseDomainEvent: baseEvent,
	}
}

// TenantSettingsUpdatedEvent is raised when tenant settings are updated
type TenantSettingsUpdatedEvent struct {
	*shared.BaseDomainEvent
}

// NewTenantSettingsUpdatedEvent creates a new TenantSettingsUpdatedEvent
func NewTenantSettingsUpdatedEvent(tenant *Tenant, oldSettings, newSettings *TenantSettings) *TenantSettingsUpdatedEvent {
	data := map[string]interface{}{
		"tenant_name": tenant.Name(),
	}

	if oldSettings != nil {
		data["old_timezone"] = oldSettings.timezone
		data["old_currency"] = oldSettings.currency
		data["old_language"] = oldSettings.language
	}

	if newSettings != nil {
		data["new_timezone"] = newSettings.timezone
		data["new_currency"] = newSettings.currency
		data["new_language"] = newSettings.language
		data["ai_enabled"] = newSettings.aiSettings.aiEnabled
		data["recommendations_enabled"] = newSettings.aiSettings.recommendationsEnabled
	}

	baseEvent := shared.NewBaseDomainEvent(
		uuid.New().String(),
		TenantSettingsUpdatedEventType,
		tenant.ID(),
		tenant.GetTenantID(),
		data,
	)

	return &TenantSettingsUpdatedEvent{
		BaseDomainEvent: baseEvent,
	}
}

// LocationAddedEvent is raised when a location is added to a tenant
type LocationAddedEvent struct {
	*shared.BaseDomainEvent
}

// NewLocationAddedEvent creates a new LocationAddedEvent
func NewLocationAddedEvent(tenant *Tenant, location *Location) *LocationAddedEvent {
	data := map[string]interface{}{
		"tenant_name":     tenant.Name(),
		"location_id":     location.ID().Value(),
		"location_name":   location.name,
		"location_address": location.address.String(),
		"location_phone":  location.phoneNumber.Value(),
		"location_email":  location.email.Value(),
		"is_active":       location.isActive,
	}

	baseEvent := shared.NewBaseDomainEvent(
		uuid.New().String(),
		LocationAddedEventType,
		tenant.ID(),
		tenant.GetTenantID(),
		data,
	)

	return &LocationAddedEvent{
		BaseDomainEvent: baseEvent,
	}
}

// LocationRemovedEvent is raised when a location is removed from a tenant
type LocationRemovedEvent struct {
	*shared.BaseDomainEvent
}

// NewLocationRemovedEvent creates a new LocationRemovedEvent
func NewLocationRemovedEvent(tenant *Tenant, location *Location) *LocationRemovedEvent {
	data := map[string]interface{}{
		"tenant_name":     tenant.Name(),
		"location_id":     location.ID().Value(),
		"location_name":   location.name,
		"location_address": location.address.String(),
	}

	baseEvent := shared.NewBaseDomainEvent(
		uuid.New().String(),
		LocationRemovedEventType,
		tenant.ID(),
		tenant.GetTenantID(),
		data,
	)

	return &LocationRemovedEvent{
		BaseDomainEvent: baseEvent,
	}
}

// LocationUpdatedEvent is raised when a location is updated
type LocationUpdatedEvent struct {
	*shared.BaseDomainEvent
}

// NewLocationUpdatedEvent creates a new LocationUpdatedEvent
func NewLocationUpdatedEvent(tenant *Tenant, location *Location, changes map[string]interface{}) *LocationUpdatedEvent {
	data := map[string]interface{}{
		"tenant_name":   tenant.Name(),
		"location_id":   location.ID().Value(),
		"location_name": location.name,
		"changes":       changes,
	}

	baseEvent := shared.NewBaseDomainEvent(
		uuid.New().String(),
		LocationUpdatedEventType,
		tenant.ID(),
		tenant.GetTenantID(),
		data,
	)

	return &LocationUpdatedEvent{
		BaseDomainEvent: baseEvent,
	}
}

// TenantEventHandler handles tenant domain events
type TenantEventHandler struct {
	name string
}

// NewTenantEventHandler creates a new tenant event handler
func NewTenantEventHandler() *TenantEventHandler {
	return &TenantEventHandler{
		name: "TenantEventHandler",
	}
}

// HandlerName returns the handler name
func (h *TenantEventHandler) HandlerName() string {
	return h.name
}

// CanHandle checks if the handler can process the event type
func (h *TenantEventHandler) CanHandle(eventType string) bool {
	switch eventType {
	case TenantCreatedEventType,
		TenantActivatedEventType,
		TenantSuspendedEventType,
		TenantDeactivatedEventType,
		TenantSubscriptionUpdatedEventType,
		TenantSettingsUpdatedEventType,
		LocationAddedEventType,
		LocationRemovedEventType,
		LocationUpdatedEventType:
		return true
	default:
		return false
	}
}

// Handle processes tenant domain events
func (h *TenantEventHandler) Handle(ctx context.Context, event shared.DomainEvent) error {
	switch event.EventType() {
	case TenantCreatedEventType:
		return h.handleTenantCreated(ctx, event)
	case TenantActivatedEventType:
		return h.handleTenantActivated(ctx, event)
	case TenantSuspendedEventType:
		return h.handleTenantSuspended(ctx, event)
	case TenantDeactivatedEventType:
		return h.handleTenantDeactivated(ctx, event)
	case TenantSubscriptionUpdatedEventType:
		return h.handleTenantSubscriptionUpdated(ctx, event)
	case TenantSettingsUpdatedEventType:
		return h.handleTenantSettingsUpdated(ctx, event)
	case LocationAddedEventType:
		return h.handleLocationAdded(ctx, event)
	case LocationRemovedEventType:
		return h.handleLocationRemoved(ctx, event)
	case LocationUpdatedEventType:
		return h.handleLocationUpdated(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.EventType())
	}
}

func (h *TenantEventHandler) handleTenantCreated(ctx context.Context, event shared.DomainEvent) error {
	// Handle tenant creation logic (e.g., send welcome email, setup default data)
	return nil
}

func (h *TenantEventHandler) handleTenantActivated(ctx context.Context, event shared.DomainEvent) error {
	// Handle tenant activation logic (e.g., enable services, send notification)
	return nil
}

func (h *TenantEventHandler) handleTenantSuspended(ctx context.Context, event shared.DomainEvent) error {
	// Handle tenant suspension logic (e.g., disable services, send notification)
	return nil
}

func (h *TenantEventHandler) handleTenantDeactivated(ctx context.Context, event shared.DomainEvent) error {
	// Handle tenant deactivation logic (e.g., cleanup data, send notification)
	return nil
}

func (h *TenantEventHandler) handleTenantSubscriptionUpdated(ctx context.Context, event shared.DomainEvent) error {
	// Handle subscription update logic (e.g., update feature flags, billing)
	return nil
}

func (h *TenantEventHandler) handleTenantSettingsUpdated(ctx context.Context, event shared.DomainEvent) error {
	// Handle settings update logic (e.g., update configurations, restart services)
	return nil
}

func (h *TenantEventHandler) handleLocationAdded(ctx context.Context, event shared.DomainEvent) error {
	// Handle location addition logic (e.g., setup location data, configure services)
	return nil
}

func (h *TenantEventHandler) handleLocationRemoved(ctx context.Context, event shared.DomainEvent) error {
	// Handle location removal logic (e.g., cleanup location data, disable services)
	return nil
}

func (h *TenantEventHandler) handleLocationUpdated(ctx context.Context, event shared.DomainEvent) error {
	// Handle location update logic (e.g., update configurations, sync data)
	return nil
}
