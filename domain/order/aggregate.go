package order

import (
	"errors"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/domain/shared"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusPreparing  OrderStatus = "preparing"
	OrderStatusReady      OrderStatus = "ready"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

// IsValid checks if the order status is valid
func (os OrderStatus) IsValid() bool {
	switch os {
	case OrderStatusPending, OrderStatusConfirmed, OrderStatusPreparing,
		OrderStatusReady, OrderStatusCompleted, OrderStatusCancelled:
		return true
	default:
		return false
	}
}

// String returns the string representation of order status
func (os OrderStatus) String() string {
	return string(os)
}

// OrderPriority represents the priority of an order
type OrderPriority string

const (
	OrderPriorityLow    OrderPriority = "low"
	OrderPriorityNormal OrderPriority = "normal"
	OrderPriorityHigh   OrderPriority = "high"
	OrderPriorityUrgent OrderPriority = "urgent"
)

// IsValid checks if the order priority is valid
func (op OrderPriority) IsValid() bool {
	switch op {
	case OrderPriorityLow, OrderPriorityNormal, OrderPriorityHigh, OrderPriorityUrgent:
		return true
	default:
		return false
	}
}

// String returns the string representation of order priority
func (op OrderPriority) String() string {
	return string(op)
}

// Customer represents a customer within a tenant
type Customer struct {
	*shared.Entity
	name            string
	email           shared.Email
	phoneNumber     shared.PhoneNumber
	loyaltyPoints   int
	preferences     map[string]string
	aiProfile       *CustomerAIProfile
}

// CustomerAIProfile holds AI-generated insights about a customer
type CustomerAIProfile struct {
	favoriteItems        []string
	averageOrderValue    shared.Money
	orderFrequency       time.Duration
	preferredOrderTime   time.Time
	dietaryRestrictions  []string
	satisfactionScore    shared.Rating
	churnRisk           shared.Percentage
	lifetimeValue       shared.Money
}

// OrderItem represents an item in an order
type OrderItem struct {
	*shared.Entity
	productID      shared.AggregateID
	productName    string
	quantity       int
	unitPrice      shared.Money
	totalPrice     shared.Money
	customizations []string
	specialNotes   string
	aiInsights     *ItemAIInsights
}

// ItemAIInsights holds AI insights for order items
type ItemAIInsights struct {
	popularityScore      shared.Percentage
	preparationTime      time.Duration
	profitMargin         shared.Percentage
	suggestedPairings    []string
	seasonalDemand       shared.Percentage
	customerSatisfaction shared.Rating
}

// OrderAIInsights holds AI insights for the entire order
type OrderAIInsights struct {
	complexityScore         shared.Rating
	estimatedPrepTime       time.Duration
	revenueImpact          shared.Money
	customerSatisfactionPrediction shared.Rating
	upsellOpportunities    []string
	crossSellSuggestions   []string
	riskFactors            []string
	optimizationSuggestions []string
}

// Order represents the order aggregate root
type Order struct {
	*shared.BaseAggregate
	orderNumber         string
	customer            *Customer
	items               []*OrderItem
	status              OrderStatus
	priority            OrderPriority
	totalAmount         shared.Money
	discountAmount      shared.Money
	taxAmount           shared.Money
	finalAmount         shared.Money
	locationID          shared.AggregateID
	estimatedCompletion time.Time
	actualCompletion    *time.Time
	specialInstructions string
	aiInsights          *OrderAIInsights
	metadata            map[string]interface{}
}

// NewOrder creates a new order aggregate
func NewOrder(
	id shared.AggregateID,
	tenantID shared.TenantID,
	orderNumber string,
	customer *Customer,
	locationID shared.AggregateID,
) (*Order, error) {
	if orderNumber == "" {
		return nil, errors.New("order number cannot be empty")
	}
	
	if customer == nil {
		return nil, errors.New("customer cannot be nil")
	}
	
	if locationID.IsEmpty() {
		return nil, errors.New("location ID cannot be empty")
	}
	
	// Initialize with zero money in USD (this should be configurable per tenant)
	zeroMoney, _ := shared.NewMoney(0, "USD")
	
	order := &Order{
		BaseAggregate:       shared.NewBaseAggregate(id, tenantID),
		orderNumber:         orderNumber,
		customer:            customer,
		items:               make([]*OrderItem, 0),
		status:              OrderStatusPending,
		priority:            OrderPriorityNormal,
		totalAmount:         zeroMoney,
		discountAmount:      zeroMoney,
		taxAmount:           zeroMoney,
		finalAmount:         zeroMoney,
		locationID:          locationID,
		estimatedCompletion: time.Now().Add(15 * time.Minute), // Default 15 minutes
		metadata:            make(map[string]interface{}),
	}
	
	// Add domain event
	event := NewOrderCreatedEvent(order)
	order.AddDomainEvent(event)
	
	return order, nil
}

// OrderNumber returns the order number
func (o *Order) OrderNumber() string {
	return o.orderNumber
}

// Customer returns the customer
func (o *Order) Customer() *Customer {
	return o.customer
}

// Items returns all order items
func (o *Order) Items() []*OrderItem {
	return o.items
}

// Status returns the order status
func (o *Order) Status() OrderStatus {
	return o.status
}

// Priority returns the order priority
func (o *Order) Priority() OrderPriority {
	return o.priority
}

// TotalAmount returns the total amount
func (o *Order) TotalAmount() shared.Money {
	return o.totalAmount
}

// FinalAmount returns the final amount after discounts and taxes
func (o *Order) FinalAmount() shared.Money {
	return o.finalAmount
}

// LocationID returns the location ID
func (o *Order) LocationID() shared.AggregateID {
	return o.locationID
}

// EstimatedCompletion returns the estimated completion time
func (o *Order) EstimatedCompletion() time.Time {
	return o.estimatedCompletion
}

// ActualCompletion returns the actual completion time
func (o *Order) ActualCompletion() *time.Time {
	return o.actualCompletion
}

// AIInsights returns the AI insights
func (o *Order) AIInsights() *OrderAIInsights {
	return o.aiInsights
}

// AddItem adds an item to the order
func (o *Order) AddItem(item *OrderItem) error {
	if item == nil {
		return errors.New("item cannot be nil")
	}
	
	// Ensure item belongs to the same tenant
	if !item.GetTenantID().Equals(o.GetTenantID()) {
		return errors.New("item must belong to the same tenant")
	}
	
	o.items = append(o.items, item)
	o.recalculateAmounts()
	o.IncrementVersion()
	
	// Add domain event
	event := NewOrderItemAddedEvent(o, item)
	o.AddDomainEvent(event)
	
	return nil
}

// RemoveItem removes an item from the order
func (o *Order) RemoveItem(itemID shared.AggregateID) error {
	for i, item := range o.items {
		if item.ID().Equals(itemID) {
			// Remove item from slice
			o.items = append(o.items[:i], o.items[i+1:]...)
			o.recalculateAmounts()
			o.IncrementVersion()
			
			// Add domain event
			event := NewOrderItemRemovedEvent(o, item)
			o.AddDomainEvent(event)
			
			return nil
		}
	}
	
	return errors.New("item not found")
}

// UpdateStatus updates the order status
func (o *Order) UpdateStatus(newStatus OrderStatus, reason string) error {
	if !newStatus.IsValid() {
		return errors.New("invalid order status")
	}
	
	// Validate status transition
	if !o.isValidStatusTransition(o.status, newStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", o.status, newStatus)
	}
	
	oldStatus := o.status
	o.status = newStatus
	o.IncrementVersion()
	
	// Set completion time if order is completed
	if newStatus == OrderStatusCompleted && o.actualCompletion == nil {
		now := time.Now()
		o.actualCompletion = &now
	}
	
	// Add domain event
	event := NewOrderStatusChangedEvent(o, oldStatus, newStatus, reason)
	o.AddDomainEvent(event)
	
	return nil
}

// UpdatePriority updates the order priority
func (o *Order) UpdatePriority(newPriority OrderPriority, reason string) error {
	if !newPriority.IsValid() {
		return errors.New("invalid order priority")
	}
	
	oldPriority := o.priority
	o.priority = newPriority
	o.IncrementVersion()
	
	// Add domain event
	event := NewOrderPriorityChangedEvent(o, oldPriority, newPriority, reason)
	o.AddDomainEvent(event)
	
	return nil
}

// ApplyDiscount applies a discount to the order
func (o *Order) ApplyDiscount(discountAmount shared.Money, reason string) error {
	if discountAmount.Currency() != o.totalAmount.Currency() {
		return errors.New("discount currency must match order currency")
	}
	
	if discountAmount.IsNegative() {
		return errors.New("discount amount cannot be negative")
	}
	
	if discountAmount.Amount() > o.totalAmount.Amount() {
		return errors.New("discount cannot exceed total amount")
	}
	
	o.discountAmount = discountAmount
	o.recalculateAmounts()
	o.IncrementVersion()
	
	// Add domain event
	event := NewOrderDiscountAppliedEvent(o, discountAmount, reason)
	o.AddDomainEvent(event)
	
	return nil
}

// UpdateEstimatedCompletion updates the estimated completion time
func (o *Order) UpdateEstimatedCompletion(newTime time.Time) error {
	if newTime.Before(time.Now()) {
		return errors.New("estimated completion cannot be in the past")
	}
	
	oldTime := o.estimatedCompletion
	o.estimatedCompletion = newTime
	o.IncrementVersion()
	
	// Add domain event
	event := NewOrderEstimatedCompletionUpdatedEvent(o, oldTime, newTime)
	o.AddDomainEvent(event)
	
	return nil
}

// SetAIInsights sets AI insights for the order
func (o *Order) SetAIInsights(insights *OrderAIInsights) {
	o.aiInsights = insights
	o.IncrementVersion()
	
	// Add domain event
	event := NewOrderAIInsightsUpdatedEvent(o, insights)
	o.AddDomainEvent(event)
}

// Cancel cancels the order
func (o *Order) Cancel(reason string) error {
	if o.status == OrderStatusCancelled {
		return errors.New("order is already cancelled")
	}
	
	if o.status == OrderStatusCompleted {
		return errors.New("completed orders cannot be cancelled")
	}
	
	return o.UpdateStatus(OrderStatusCancelled, reason)
}

// IsActive checks if the order is active (not cancelled or completed)
func (o *Order) IsActive() bool {
	return o.status != OrderStatusCancelled && o.status != OrderStatusCompleted
}

// CanBeModified checks if the order can be modified
func (o *Order) CanBeModified() bool {
	return o.status == OrderStatusPending || o.status == OrderStatusConfirmed
}

// GetItemByID returns an item by ID
func (o *Order) GetItemByID(itemID shared.AggregateID) (*OrderItem, error) {
	for _, item := range o.items {
		if item.ID().Equals(itemID) {
			return item, nil
		}
	}
	return nil, errors.New("item not found")
}

// GetItemCount returns the total number of items
func (o *Order) GetItemCount() int {
	totalCount := 0
	for _, item := range o.items {
		totalCount += item.quantity
	}
	return totalCount
}

// SetMetadata sets metadata for the order
func (o *Order) SetMetadata(key string, value interface{}) {
	o.metadata[key] = value
	o.IncrementVersion()
}

// GetMetadata gets metadata from the order
func (o *Order) GetMetadata(key string) (interface{}, bool) {
	value, exists := o.metadata[key]
	return value, exists
}

// Private methods

// recalculateAmounts recalculates all monetary amounts
func (o *Order) recalculateAmounts() {
	// Calculate total from items
	total := int64(0)
	currency := "USD" // Default currency, should be tenant-specific
	
	for _, item := range o.items {
		total += item.totalPrice.Amount()
		currency = item.totalPrice.Currency()
	}
	
	totalMoney, _ := shared.NewMoney(total, currency)
	o.totalAmount = totalMoney
	
	// Calculate final amount (total - discount + tax)
	finalAmount := total - o.discountAmount.Amount() + o.taxAmount.Amount()
	finalMoney, _ := shared.NewMoney(finalAmount, currency)
	o.finalAmount = finalMoney
}

// isValidStatusTransition checks if a status transition is valid
func (o *Order) isValidStatusTransition(from, to OrderStatus) bool {
	validTransitions := map[OrderStatus][]OrderStatus{
		OrderStatusPending: {
			OrderStatusConfirmed,
			OrderStatusCancelled,
		},
		OrderStatusConfirmed: {
			OrderStatusPreparing,
			OrderStatusCancelled,
		},
		OrderStatusPreparing: {
			OrderStatusReady,
			OrderStatusCancelled,
		},
		OrderStatusReady: {
			OrderStatusCompleted,
		},
		OrderStatusCompleted: {}, // Terminal state
		OrderStatusCancelled: {}, // Terminal state
	}
	
	allowedTransitions, exists := validTransitions[from]
	if !exists {
		return false
	}
	
	for _, allowed := range allowedTransitions {
		if allowed == to {
			return true
		}
	}
	
	return false
}

// NewOrderItem creates a new order item
func NewOrderItem(
	id shared.AggregateID,
	tenantID shared.TenantID,
	productID shared.AggregateID,
	productName string,
	quantity int,
	unitPrice shared.Money,
) (*OrderItem, error) {
	if productName == "" {
		return nil, errors.New("product name cannot be empty")
	}
	
	if quantity <= 0 {
		return nil, errors.New("quantity must be positive")
	}
	
	if unitPrice.IsNegative() {
		return nil, errors.New("unit price cannot be negative")
	}
	
	totalPrice := unitPrice.Multiply(float64(quantity))
	
	return &OrderItem{
		Entity:         shared.NewEntity(id, tenantID),
		productID:      productID,
		productName:    productName,
		quantity:       quantity,
		unitPrice:      unitPrice,
		totalPrice:     totalPrice,
		customizations: make([]string, 0),
	}, nil
}

// NewCustomer creates a new customer
func NewCustomer(
	id shared.AggregateID,
	tenantID shared.TenantID,
	name string,
	email shared.Email,
	phoneNumber shared.PhoneNumber,
) (*Customer, error) {
	if name == "" {
		return nil, errors.New("customer name cannot be empty")
	}
	
	return &Customer{
		Entity:        shared.NewEntity(id, tenantID),
		name:          name,
		email:         email,
		phoneNumber:   phoneNumber,
		loyaltyPoints: 0,
		preferences:   make(map[string]string),
	}, nil
}
