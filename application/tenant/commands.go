package tenant

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/DimaJoyti/go-coffee/domain/shared"
	"github.com/DimaJoyti/go-coffee/domain/tenant"
)

// CreateTenantCommand represents a command to create a new tenant
type CreateTenantCommand struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	TenantType  string `json:"tenant_type" validate:"required,oneof=restaurant cafe chain franchise"`
	Email       string `json:"email" validate:"required,email"`
	PhoneNumber string `json:"phone_number" validate:"required"`
	Address     struct {
		Street     string `json:"street" validate:"required"`
		City       string `json:"city" validate:"required"`
		State      string `json:"state"`
		PostalCode string `json:"postal_code"`
		Country    string `json:"country" validate:"required"`
	} `json:"address" validate:"required"`
	OwnerID          string `json:"owner_id" validate:"required"`
	SubscriptionPlan string `json:"subscription_plan" validate:"required,oneof=basic professional enterprise"`
}

// CreateTenantResult represents the result of creating a tenant
type CreateTenantResult struct {
	TenantID   string    `json:"tenant_id"`
	TenantName string    `json:"tenant_name"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// UpdateTenantCommand represents a command to update a tenant
type UpdateTenantCommand struct {
	TenantID    string `json:"tenant_id" validate:"required"`
	Name        string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Email       string `json:"email,omitempty" validate:"omitempty,email"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Address     *struct {
		Street     string `json:"street"`
		City       string `json:"city"`
		State      string `json:"state"`
		PostalCode string `json:"postal_code"`
		Country    string `json:"country"`
	} `json:"address,omitempty"`
}

// ActivateTenantCommand represents a command to activate a tenant
type ActivateTenantCommand struct {
	TenantID string `json:"tenant_id" validate:"required"`
}

// SuspendTenantCommand represents a command to suspend a tenant
type SuspendTenantCommand struct {
	TenantID string `json:"tenant_id" validate:"required"`
	Reason   string `json:"reason" validate:"required,min=10,max=500"`
}

// UpdateSubscriptionCommand represents a command to update tenant subscription
type UpdateSubscriptionCommand struct {
	TenantID         string `json:"tenant_id" validate:"required"`
	SubscriptionPlan string `json:"subscription_plan" validate:"required,oneof=basic professional enterprise"`
	BillingCycle     string `json:"billing_cycle" validate:"required,oneof=monthly yearly"`
	MaxUsers         int    `json:"max_users" validate:"min=1"`
	MaxLocations     int    `json:"max_locations" validate:"min=1"`
	MaxOrdersPerMonth int   `json:"max_orders_per_month" validate:"min=1"`
}

// AddLocationCommand represents a command to add a location to a tenant
type AddLocationCommand struct {
	TenantID    string `json:"tenant_id" validate:"required"`
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Email       string `json:"email" validate:"required,email"`
	PhoneNumber string `json:"phone_number" validate:"required"`
	Address     struct {
		Street     string `json:"street" validate:"required"`
		City       string `json:"city" validate:"required"`
		State      string `json:"state"`
		PostalCode string `json:"postal_code"`
		Country    string `json:"country" validate:"required"`
	} `json:"address" validate:"required"`
}

// TenantCommandHandler handles tenant-related commands
type TenantCommandHandler struct {
	tenantRepository  tenant.TenantRepository
	domainService     *tenant.TenantDomainService
	eventPublisher    shared.DomainEventPublisher
}

// NewTenantCommandHandler creates a new tenant command handler
func NewTenantCommandHandler(
	tenantRepository tenant.TenantRepository,
	domainService *tenant.TenantDomainService,
	eventPublisher shared.DomainEventPublisher,
) *TenantCommandHandler {
	return &TenantCommandHandler{
		tenantRepository: tenantRepository,
		domainService:    domainService,
		eventPublisher:   eventPublisher,
	}
}

// HandleCreateTenant handles the create tenant command
func (h *TenantCommandHandler) HandleCreateTenant(ctx context.Context, cmd *CreateTenantCommand) (*CreateTenantResult, error) {
	// Validate command
	if err := h.validateCreateTenantCommand(cmd); err != nil {
		return nil, err
	}

	// Create value objects
	email, err := shared.NewEmail(cmd.Email)
	if err != nil {
		return nil, err
	}

	phoneNumber, err := shared.NewPhoneNumber(cmd.PhoneNumber)
	if err != nil {
		return nil, err
	}

	address, err := shared.NewAddress(
		cmd.Address.Street,
		cmd.Address.City,
		cmd.Address.State,
		cmd.Address.PostalCode,
		cmd.Address.Country,
	)
	if err != nil {
		return nil, err
	}

	// Validate uniqueness
	if err := h.domainService.ValidateUniqueness(ctx, email, nil); err != nil {
		return nil, err
	}

	// Generate IDs
	tenantAggregateID := shared.NewAggregateID(uuid.New().String())
	tenantID, err := shared.NewTenantID(uuid.New().String())
	if err != nil {
		return nil, err
	}
	ownerID := shared.NewAggregateID(cmd.OwnerID)

	// Parse tenant type
	tenantType := tenant.TenantType(cmd.TenantType)

	// Create tenant aggregate
	tenantAggregate, err := tenant.NewTenant(
		tenantAggregateID,
		tenantID,
		cmd.Name,
		tenantType,
		email,
		phoneNumber,
		address,
		ownerID,
	)
	if err != nil {
		return nil, err
	}

	// Create subscription
	subscription := tenant.NewSubscription(
		shared.SubscriptionPlan(cmd.SubscriptionPlan),
		tenant.SubscriptionStatusTrial,
		time.Now(),
		tenant.BillingCycleMonthly,
		getDefaultMaxUsers(shared.SubscriptionPlan(cmd.SubscriptionPlan)),
		getDefaultMaxLocations(shared.SubscriptionPlan(cmd.SubscriptionPlan)),
		getDefaultMaxOrdersPerMonth(shared.SubscriptionPlan(cmd.SubscriptionPlan)),
	)

	// Update subscription
	if err := tenantAggregate.UpdateSubscription(subscription); err != nil {
		return nil, err
	}

	// Save tenant
	if err := h.tenantRepository.Save(ctx, tenantAggregate); err != nil {
		return nil, err
	}

	// Publish domain events
	if err := h.publishDomainEvents(ctx, tenantAggregate); err != nil {
		// Log error but don't fail the command
		// In a real implementation, you might want to use an outbox pattern
	}

	return &CreateTenantResult{
		TenantID:   tenantID.Value(),
		TenantName: tenantAggregate.Name(),
		Status:     tenantAggregate.Status().String(),
		CreatedAt:  tenantAggregate.CreatedAt(),
	}, nil
}

// HandleActivateTenant handles the activate tenant command
func (h *TenantCommandHandler) HandleActivateTenant(ctx context.Context, cmd *ActivateTenantCommand) error {
	tenantID, err := shared.NewTenantID(cmd.TenantID)
	if err != nil {
		return err
	}

	tenantAggregate, err := h.tenantRepository.FindByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}

	if tenantAggregate == nil {
		return errors.New("tenant not found")
	}

	if err := tenantAggregate.Activate(); err != nil {
		return err
	}

	if err := h.tenantRepository.Save(ctx, tenantAggregate); err != nil {
		return err
	}

	return h.publishDomainEvents(ctx, tenantAggregate)
}

// HandleSuspendTenant handles the suspend tenant command
func (h *TenantCommandHandler) HandleSuspendTenant(ctx context.Context, cmd *SuspendTenantCommand) error {
	tenantID, err := shared.NewTenantID(cmd.TenantID)
	if err != nil {
		return err
	}

	tenantAggregate, err := h.tenantRepository.FindByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}

	if tenantAggregate == nil {
		return errors.New("tenant not found")
	}

	if err := tenantAggregate.Suspend(cmd.Reason); err != nil {
		return err
	}

	if err := h.tenantRepository.Save(ctx, tenantAggregate); err != nil {
		return err
	}

	return h.publishDomainEvents(ctx, tenantAggregate)
}

// HandleUpdateSubscription handles the update subscription command
func (h *TenantCommandHandler) HandleUpdateSubscription(ctx context.Context, cmd *UpdateSubscriptionCommand) error {
	tenantID, err := shared.NewTenantID(cmd.TenantID)
	if err != nil {
		return err
	}

	// Validate upgrade path
	newPlan := shared.SubscriptionPlan(cmd.SubscriptionPlan)
	if err := h.domainService.CanUpgradeSubscription(ctx, tenantID, newPlan); err != nil {
		return err
	}

	tenantAggregate, err := h.tenantRepository.FindByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}

	if tenantAggregate == nil {
		return errors.New("tenant not found")
	}

	// Create new subscription
	subscription := tenant.NewSubscription(
		newPlan,
		tenant.SubscriptionStatusActive,
		time.Now(),
		tenant.BillingCycle(cmd.BillingCycle),
		cmd.MaxUsers,
		cmd.MaxLocations,
		cmd.MaxOrdersPerMonth,
	)

	if err := tenantAggregate.UpdateSubscription(subscription); err != nil {
		return err
	}

	if err := h.tenantRepository.Save(ctx, tenantAggregate); err != nil {
		return err
	}

	return h.publishDomainEvents(ctx, tenantAggregate)
}

// HandleAddLocation handles the add location command
func (h *TenantCommandHandler) HandleAddLocation(ctx context.Context, cmd *AddLocationCommand) error {
	tenantID, err := shared.NewTenantID(cmd.TenantID)
	if err != nil {
		return err
	}

	tenantAggregate, err := h.tenantRepository.FindByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}

	if tenantAggregate == nil {
		return errors.New("tenant not found")
	}

	// Create value objects
	email, err := shared.NewEmail(cmd.Email)
	if err != nil {
		return err
	}

	phoneNumber, err := shared.NewPhoneNumber(cmd.PhoneNumber)
	if err != nil {
		return err
	}

	address, err := shared.NewAddress(
		cmd.Address.Street,
		cmd.Address.City,
		cmd.Address.State,
		cmd.Address.PostalCode,
		cmd.Address.Country,
	)
	if err != nil {
		return err
	}

	locationID := shared.NewAggregateID(uuid.New().String())

	if err := tenantAggregate.AddLocation(locationID, cmd.Name, address, phoneNumber, email); err != nil {
		return err
	}

	if err := h.tenantRepository.Save(ctx, tenantAggregate); err != nil {
		return err
	}

	return h.publishDomainEvents(ctx, tenantAggregate)
}

// Helper methods

func (h *TenantCommandHandler) validateCreateTenantCommand(cmd *CreateTenantCommand) error {
	if cmd.Name == "" {
		return errors.New("tenant name is required")
	}
	if cmd.Email == "" {
		return errors.New("email is required")
	}
	if cmd.OwnerID == "" {
		return errors.New("owner ID is required")
	}
	return nil
}

func (h *TenantCommandHandler) publishDomainEvents(ctx context.Context, aggregate *tenant.Tenant) error {
	events := aggregate.GetDomainEvents()
	for _, event := range events {
		if err := h.eventPublisher.Publish(ctx, event); err != nil {
			return err
		}
	}
	aggregate.ClearDomainEvents()
	return nil
}

func getDefaultMaxUsers(plan shared.SubscriptionPlan) int {
	switch plan {
	case shared.SubscriptionBasic:
		return 5
	case shared.SubscriptionProfessional:
		return 25
	case shared.SubscriptionEnterprise:
		return 100
	default:
		return 1
	}
}

func getDefaultMaxLocations(plan shared.SubscriptionPlan) int {
	switch plan {
	case shared.SubscriptionBasic:
		return 1
	case shared.SubscriptionProfessional:
		return 5
	case shared.SubscriptionEnterprise:
		return 50
	default:
		return 1
	}
}

func getDefaultMaxOrdersPerMonth(plan shared.SubscriptionPlan) int {
	switch plan {
	case shared.SubscriptionBasic:
		return 1000
	case shared.SubscriptionProfessional:
		return 10000
	case shared.SubscriptionEnterprise:
		return 100000
	default:
		return 100
	}
}
