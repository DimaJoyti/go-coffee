package order

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/DimaJoyti/go-coffee/domain/shared"
	"github.com/DimaJoyti/go-coffee/domain/order"
)

// CreateOrderCommand represents a command to create a new order
type CreateOrderCommand struct {
	CustomerID          string `json:"customer_id" validate:"required"`
	LocationID          string `json:"location_id" validate:"required"`
	Items               []OrderItemCommand `json:"items" validate:"required,min=1"`
	SpecialInstructions string `json:"special_instructions,omitempty"`
	Priority            string `json:"priority,omitempty" validate:"omitempty,oneof=low normal high urgent"`
}

// OrderItemCommand represents an item in the order command
type OrderItemCommand struct {
	ProductID      string   `json:"product_id" validate:"required"`
	ProductName    string   `json:"product_name" validate:"required"`
	Quantity       int      `json:"quantity" validate:"required,min=1"`
	UnitPrice      float64  `json:"unit_price" validate:"required,min=0"`
	Currency       string   `json:"currency" validate:"required,len=3"`
	Customizations []string `json:"customizations,omitempty"`
	SpecialNotes   string   `json:"special_notes,omitempty"`
}

// CreateOrderResult represents the result of creating an order
type CreateOrderResult struct {
	OrderID             string    `json:"order_id"`
	OrderNumber         string    `json:"order_number"`
	Status              string    `json:"status"`
	TotalAmount         float64   `json:"total_amount"`
	Currency            string    `json:"currency"`
	EstimatedCompletion time.Time `json:"estimated_completion"`
	CreatedAt           time.Time `json:"created_at"`
}

// UpdateOrderStatusCommand represents a command to update order status
type UpdateOrderStatusCommand struct {
	OrderID   string `json:"order_id" validate:"required"`
	NewStatus string `json:"new_status" validate:"required,oneof=pending confirmed preparing ready completed cancelled"`
	Reason    string `json:"reason,omitempty"`
}

// UpdateOrderPriorityCommand represents a command to update order priority
type UpdateOrderPriorityCommand struct {
	OrderID     string `json:"order_id" validate:"required"`
	NewPriority string `json:"new_priority" validate:"required,oneof=low normal high urgent"`
	Reason      string `json:"reason,omitempty"`
}

// AddOrderItemCommand represents a command to add an item to an order
type AddOrderItemCommand struct {
	OrderID        string `json:"order_id" validate:"required"`
	ProductID      string `json:"product_id" validate:"required"`
	ProductName    string `json:"product_name" validate:"required"`
	Quantity       int    `json:"quantity" validate:"required,min=1"`
	UnitPrice      float64 `json:"unit_price" validate:"required,min=0"`
	Currency       string `json:"currency" validate:"required,len=3"`
	Customizations []string `json:"customizations,omitempty"`
	SpecialNotes   string `json:"special_notes,omitempty"`
}

// RemoveOrderItemCommand represents a command to remove an item from an order
type RemoveOrderItemCommand struct {
	OrderID string `json:"order_id" validate:"required"`
	ItemID  string `json:"item_id" validate:"required"`
}

// ApplyDiscountCommand represents a command to apply a discount to an order
type ApplyDiscountCommand struct {
	OrderID        string  `json:"order_id" validate:"required"`
	DiscountAmount float64 `json:"discount_amount" validate:"required,min=0"`
	Currency       string  `json:"currency" validate:"required,len=3"`
	Reason         string  `json:"reason" validate:"required"`
}

// CancelOrderCommand represents a command to cancel an order
type CancelOrderCommand struct {
	OrderID string `json:"order_id" validate:"required"`
	Reason  string `json:"reason" validate:"required,min=5,max=500"`
}

// OrderCommandHandler handles order-related commands
type OrderCommandHandler struct {
	orderRepository order.OrderRepository
	eventPublisher  shared.DomainEventPublisher
}

// NewOrderCommandHandler creates a new order command handler
func NewOrderCommandHandler(
	orderRepository order.OrderRepository,
	eventPublisher shared.DomainEventPublisher,
) *OrderCommandHandler {
	return &OrderCommandHandler{
		orderRepository: orderRepository,
		eventPublisher:  eventPublisher,
	}
}

// HandleCreateOrder handles the create order command
func (h *OrderCommandHandler) HandleCreateOrder(ctx context.Context, cmd *CreateOrderCommand) (*CreateOrderResult, error) {
	// Extract tenant context
	tenantCtx, err := shared.FromContext(ctx)
	if err != nil {
		return nil, errors.New("tenant context is required")
	}

	// Validate command
	if err := h.validateCreateOrderCommand(cmd); err != nil {
		return nil, err
	}

	// Generate IDs
	orderID := shared.NewAggregateID(uuid.New().String())
	customerID := shared.NewAggregateID(cmd.CustomerID)
	locationID := shared.NewAggregateID(cmd.LocationID)

	// Generate order number (in real implementation, this would be more sophisticated)
	orderNumber := generateOrderNumber()

	// Create customer (in real implementation, this would be fetched from customer service)
	customer, err := h.createCustomerFromCommand(customerID, tenantCtx.TenantID())
	if err != nil {
		return nil, err
	}

	// Create order aggregate
	orderAggregate, err := order.NewOrder(
		orderID,
		tenantCtx.TenantID(),
		orderNumber,
		customer,
		locationID,
	)
	if err != nil {
		return nil, err
	}

	// Set priority if specified
	if cmd.Priority != "" {
		priority := order.OrderPriority(cmd.Priority)
		if err := orderAggregate.UpdatePriority(priority, "Set during order creation"); err != nil {
			return nil, err
		}
	}

	// Add items to order
	for _, itemCmd := range cmd.Items {
		item, err := h.createOrderItemFromCommand(itemCmd, tenantCtx.TenantID())
		if err != nil {
			return nil, err
		}

		if err := orderAggregate.AddItem(item); err != nil {
			return nil, err
		}
	}

	// Set special instructions
	if cmd.SpecialInstructions != "" {
		orderAggregate.SetMetadata("special_instructions", cmd.SpecialInstructions)
	}

	// Save order
	if err := h.orderRepository.Save(ctx, orderAggregate); err != nil {
		return nil, err
	}

	// Publish domain events
	if err := h.publishDomainEvents(ctx, orderAggregate); err != nil {
		// Log error but don't fail the command
	}

	return &CreateOrderResult{
		OrderID:             orderID.Value(),
		OrderNumber:         orderAggregate.OrderNumber(),
		Status:              orderAggregate.Status().String(),
		TotalAmount:         orderAggregate.TotalAmount().ToFloat(),
		Currency:            orderAggregate.TotalAmount().Currency(),
		EstimatedCompletion: orderAggregate.EstimatedCompletion(),
		CreatedAt:           orderAggregate.CreatedAt(),
	}, nil
}

// HandleUpdateOrderStatus handles the update order status command
func (h *OrderCommandHandler) HandleUpdateOrderStatus(ctx context.Context, cmd *UpdateOrderStatusCommand) error {
	// Extract tenant context
	tenantCtx, err := shared.FromContext(ctx)
	if err != nil {
		return errors.New("tenant context is required")
	}

	orderID := shared.NewAggregateID(cmd.OrderID)
	
	orderAggregate, err := h.orderRepository.FindByID(ctx, orderID, tenantCtx.TenantID())
	if err != nil {
		return err
	}

	if orderAggregate == nil {
		return errors.New("order not found")
	}

	newStatus := order.OrderStatus(cmd.NewStatus)
	if err := orderAggregate.UpdateStatus(newStatus, cmd.Reason); err != nil {
		return err
	}

	if err := h.orderRepository.Save(ctx, orderAggregate); err != nil {
		return err
	}

	return h.publishDomainEvents(ctx, orderAggregate)
}

// HandleUpdateOrderPriority handles the update order priority command
func (h *OrderCommandHandler) HandleUpdateOrderPriority(ctx context.Context, cmd *UpdateOrderPriorityCommand) error {
	// Extract tenant context
	tenantCtx, err := shared.FromContext(ctx)
	if err != nil {
		return errors.New("tenant context is required")
	}

	orderID := shared.NewAggregateID(cmd.OrderID)
	
	orderAggregate, err := h.orderRepository.FindByID(ctx, orderID, tenantCtx.TenantID())
	if err != nil {
		return err
	}

	if orderAggregate == nil {
		return errors.New("order not found")
	}

	newPriority := order.OrderPriority(cmd.NewPriority)
	if err := orderAggregate.UpdatePriority(newPriority, cmd.Reason); err != nil {
		return err
	}

	if err := h.orderRepository.Save(ctx, orderAggregate); err != nil {
		return err
	}

	return h.publishDomainEvents(ctx, orderAggregate)
}

// HandleAddOrderItem handles the add order item command
func (h *OrderCommandHandler) HandleAddOrderItem(ctx context.Context, cmd *AddOrderItemCommand) error {
	// Extract tenant context
	tenantCtx, err := shared.FromContext(ctx)
	if err != nil {
		return errors.New("tenant context is required")
	}

	orderID := shared.NewAggregateID(cmd.OrderID)
	
	orderAggregate, err := h.orderRepository.FindByID(ctx, orderID, tenantCtx.TenantID())
	if err != nil {
		return err
	}

	if orderAggregate == nil {
		return errors.New("order not found")
	}

	if !orderAggregate.CanBeModified() {
		return errors.New("order cannot be modified in current status")
	}

	// Create order item
	item, err := h.createOrderItemFromAddCommand(cmd, tenantCtx.TenantID())
	if err != nil {
		return err
	}

	if err := orderAggregate.AddItem(item); err != nil {
		return err
	}

	if err := h.orderRepository.Save(ctx, orderAggregate); err != nil {
		return err
	}

	return h.publishDomainEvents(ctx, orderAggregate)
}

// HandleRemoveOrderItem handles the remove order item command
func (h *OrderCommandHandler) HandleRemoveOrderItem(ctx context.Context, cmd *RemoveOrderItemCommand) error {
	// Extract tenant context
	tenantCtx, err := shared.FromContext(ctx)
	if err != nil {
		return errors.New("tenant context is required")
	}

	orderID := shared.NewAggregateID(cmd.OrderID)
	itemID := shared.NewAggregateID(cmd.ItemID)
	
	orderAggregate, err := h.orderRepository.FindByID(ctx, orderID, tenantCtx.TenantID())
	if err != nil {
		return err
	}

	if orderAggregate == nil {
		return errors.New("order not found")
	}

	if !orderAggregate.CanBeModified() {
		return errors.New("order cannot be modified in current status")
	}

	if err := orderAggregate.RemoveItem(itemID); err != nil {
		return err
	}

	if err := h.orderRepository.Save(ctx, orderAggregate); err != nil {
		return err
	}

	return h.publishDomainEvents(ctx, orderAggregate)
}

// HandleApplyDiscount handles the apply discount command
func (h *OrderCommandHandler) HandleApplyDiscount(ctx context.Context, cmd *ApplyDiscountCommand) error {
	// Extract tenant context
	tenantCtx, err := shared.FromContext(ctx)
	if err != nil {
		return errors.New("tenant context is required")
	}

	orderID := shared.NewAggregateID(cmd.OrderID)
	
	orderAggregate, err := h.orderRepository.FindByID(ctx, orderID, tenantCtx.TenantID())
	if err != nil {
		return err
	}

	if orderAggregate == nil {
		return errors.New("order not found")
	}

	// Create discount money
	discountMoney, err := shared.NewMoneyFromFloat(cmd.DiscountAmount, cmd.Currency)
	if err != nil {
		return err
	}

	if err := orderAggregate.ApplyDiscount(discountMoney, cmd.Reason); err != nil {
		return err
	}

	if err := h.orderRepository.Save(ctx, orderAggregate); err != nil {
		return err
	}

	return h.publishDomainEvents(ctx, orderAggregate)
}

// HandleCancelOrder handles the cancel order command
func (h *OrderCommandHandler) HandleCancelOrder(ctx context.Context, cmd *CancelOrderCommand) error {
	// Extract tenant context
	tenantCtx, err := shared.FromContext(ctx)
	if err != nil {
		return errors.New("tenant context is required")
	}

	orderID := shared.NewAggregateID(cmd.OrderID)
	
	orderAggregate, err := h.orderRepository.FindByID(ctx, orderID, tenantCtx.TenantID())
	if err != nil {
		return err
	}

	if orderAggregate == nil {
		return errors.New("order not found")
	}

	if err := orderAggregate.Cancel(cmd.Reason); err != nil {
		return err
	}

	if err := h.orderRepository.Save(ctx, orderAggregate); err != nil {
		return err
	}

	return h.publishDomainEvents(ctx, orderAggregate)
}

// Helper methods

func (h *OrderCommandHandler) validateCreateOrderCommand(cmd *CreateOrderCommand) error {
	if cmd.CustomerID == "" {
		return errors.New("customer ID is required")
	}
	if cmd.LocationID == "" {
		return errors.New("location ID is required")
	}
	if len(cmd.Items) == 0 {
		return errors.New("at least one item is required")
	}
	return nil
}

func (h *OrderCommandHandler) createCustomerFromCommand(customerID shared.AggregateID, tenantID shared.TenantID) (*order.Customer, error) {
	// In a real implementation, this would fetch customer data from customer service
	// For now, we'll create a mock customer
	email, _ := shared.NewEmail("customer@example.com")
	phoneNumber, _ := shared.NewPhoneNumber("+1234567890")
	
	return order.NewCustomer(
		customerID,
		tenantID,
		"John Doe",
		email,
		phoneNumber,
	)
}

func (h *OrderCommandHandler) createOrderItemFromCommand(cmd OrderItemCommand, tenantID shared.TenantID) (*order.OrderItem, error) {
	itemID := shared.NewAggregateID(uuid.New().String())
	productID := shared.NewAggregateID(cmd.ProductID)
	
	unitPrice, err := shared.NewMoneyFromFloat(cmd.UnitPrice, cmd.Currency)
	if err != nil {
		return nil, err
	}

	return order.NewOrderItem(
		itemID,
		tenantID,
		productID,
		cmd.ProductName,
		cmd.Quantity,
		unitPrice,
	)
}

func (h *OrderCommandHandler) createOrderItemFromAddCommand(cmd *AddOrderItemCommand, tenantID shared.TenantID) (*order.OrderItem, error) {
	itemID := shared.NewAggregateID(uuid.New().String())
	productID := shared.NewAggregateID(cmd.ProductID)
	
	unitPrice, err := shared.NewMoneyFromFloat(cmd.UnitPrice, cmd.Currency)
	if err != nil {
		return nil, err
	}

	return order.NewOrderItem(
		itemID,
		tenantID,
		productID,
		cmd.ProductName,
		cmd.Quantity,
		unitPrice,
	)
}

func (h *OrderCommandHandler) publishDomainEvents(ctx context.Context, aggregate *order.Order) error {
	events := aggregate.GetDomainEvents()
	for _, event := range events {
		if err := h.eventPublisher.Publish(ctx, event); err != nil {
			return err
		}
	}
	aggregate.ClearDomainEvents()
	return nil
}

func generateOrderNumber() string {
	// Simple order number generation - in production, this would be more sophisticated
	return "ORD-" + uuid.New().String()[:8]
}
