package domain

import (
	"errors"
	"time"
)

// OrderStatus represents the status of an order
type OrderStatus int32

const (
	OrderStatusUnknown    OrderStatus = 0
	OrderStatusPending    OrderStatus = 1
	OrderStatusConfirmed  OrderStatus = 2
	OrderStatusPreparing  OrderStatus = 3
	OrderStatusReady      OrderStatus = 4
	OrderStatusCompleted  OrderStatus = 5
	OrderStatusCancelled  OrderStatus = 6
	OrderStatusRefunded   OrderStatus = 7
)

// String returns the string representation of OrderStatus
func (s OrderStatus) String() string {
	switch s {
	case OrderStatusPending:
		return "PENDING"
	case OrderStatusConfirmed:
		return "CONFIRMED"
	case OrderStatusPreparing:
		return "PREPARING"
	case OrderStatusReady:
		return "READY"
	case OrderStatusCompleted:
		return "COMPLETED"
	case OrderStatusCancelled:
		return "CANCELLED"
	case OrderStatusRefunded:
		return "REFUNDED"
	default:
		return "UNKNOWN"
	}
}

// OrderPriority represents the priority level of an order
type OrderPriority int32

const (
	OrderPriorityNormal OrderPriority = 0
	OrderPriorityHigh   OrderPriority = 1
	OrderPriorityUrgent OrderPriority = 2
)

// PaymentMethod represents different payment methods
type PaymentMethod int32

const (
	PaymentMethodUnknown     PaymentMethod = 0
	PaymentMethodCreditCard  PaymentMethod = 1
	PaymentMethodDebitCard   PaymentMethod = 2
	PaymentMethodCash        PaymentMethod = 3
	PaymentMethodCrypto      PaymentMethod = 4
	PaymentMethodLoyaltyToken PaymentMethod = 5
)

// Order represents a customer order
type Order struct {
	ID              string         `json:"id"`
	CustomerID      string         `json:"customer_id"`
	Items           []*OrderItem   `json:"items"`
	Status          OrderStatus    `json:"status"`
	Priority        OrderPriority  `json:"priority"`
	TotalAmount     int64          `json:"total_amount"` // in cents
	Currency        string         `json:"currency"`
	PaymentMethod   PaymentMethod  `json:"payment_method"`
	PaymentID       string         `json:"payment_id,omitempty"`
	EstimatedTime   int32          `json:"estimated_time"` // in seconds
	ActualTime      int32          `json:"actual_time"`    // in seconds
	SpecialInstructions string     `json:"special_instructions,omitempty"`
	DeliveryAddress *Address       `json:"delivery_address,omitempty"`
	IsDelivery      bool           `json:"is_delivery"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	ConfirmedAt     *time.Time     `json:"confirmed_at,omitempty"`
	CompletedAt     *time.Time     `json:"completed_at,omitempty"`
	CancelledAt     *time.Time     `json:"cancelled_at,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID           string            `json:"id"`
	ProductID    string            `json:"product_id"`
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	Quantity     int32             `json:"quantity"`
	UnitPrice    int64             `json:"unit_price"` // in cents
	TotalPrice   int64             `json:"total_price"` // in cents
	Customizations []*Customization `json:"customizations,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Customization represents a customization for an order item
type Customization struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	ExtraPrice  int64  `json:"extra_price"` // in cents
}

// Address represents a delivery address
type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
	Latitude   float64 `json:"latitude,omitempty"`
	Longitude  float64 `json:"longitude,omitempty"`
}

// Business Rules and Validation

// NewOrder creates a new order with validation
func NewOrder(customerID string, items []*OrderItem) (*Order, error) {
	if customerID == "" {
		return nil, errors.New("customer ID is required")
	}
	
	if len(items) == 0 {
		return nil, errors.New("order must have at least one item")
	}

	order := &Order{
		ID:         generateOrderID(),
		CustomerID: customerID,
		Items:      items,
		Status:     OrderStatusPending,
		Priority:   OrderPriorityNormal,
		Currency:   "USD",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Metadata:   make(map[string]string),
	}

	// Calculate total amount
	if err := order.CalculateTotal(); err != nil {
		return nil, err
	}

	return order, nil
}

// CalculateTotal calculates the total amount for the order
func (o *Order) CalculateTotal() error {
	var total int64
	
	for _, item := range o.Items {
		if item.Quantity <= 0 {
			return errors.New("item quantity must be positive")
		}
		
		if item.UnitPrice < 0 {
			return errors.New("item unit price cannot be negative")
		}
		
		// Calculate item total including customizations
		itemTotal := item.UnitPrice * int64(item.Quantity)
		
		for _, customization := range item.Customizations {
			itemTotal += customization.ExtraPrice * int64(item.Quantity)
		}
		
		item.TotalPrice = itemTotal
		total += itemTotal
	}
	
	o.TotalAmount = total
	o.UpdatedAt = time.Now()
	
	return nil
}

// CanTransitionTo checks if the order can transition to the given status
func (o *Order) CanTransitionTo(newStatus OrderStatus) bool {
	switch o.Status {
	case OrderStatusPending:
		return newStatus == OrderStatusConfirmed || newStatus == OrderStatusCancelled
	case OrderStatusConfirmed:
		return newStatus == OrderStatusPreparing || newStatus == OrderStatusCancelled
	case OrderStatusPreparing:
		return newStatus == OrderStatusReady || newStatus == OrderStatusCancelled
	case OrderStatusReady:
		return newStatus == OrderStatusCompleted || newStatus == OrderStatusCancelled
	case OrderStatusCompleted:
		return newStatus == OrderStatusRefunded
	case OrderStatusCancelled:
		return newStatus == OrderStatusRefunded
	case OrderStatusRefunded:
		return false // Terminal state
	default:
		return false
	}
}

// UpdateStatus updates the order status with validation
func (o *Order) UpdateStatus(newStatus OrderStatus) error {
	if !o.CanTransitionTo(newStatus) {
		return errors.New("invalid status transition")
	}
	
	o.Status = newStatus
	o.UpdatedAt = time.Now()
	
	// Set timestamps for specific status changes
	now := time.Now()
	switch newStatus {
	case OrderStatusConfirmed:
		o.ConfirmedAt = &now
	case OrderStatusCompleted:
		o.CompletedAt = &now
	case OrderStatusCancelled:
		o.CancelledAt = &now
	}
	
	return nil
}

// SetPriority sets the order priority
func (o *Order) SetPriority(priority OrderPriority) {
	o.Priority = priority
	o.UpdatedAt = time.Now()
}

// AddItem adds an item to the order
func (o *Order) AddItem(item *OrderItem) error {
	if item == nil {
		return errors.New("item cannot be nil")
	}
	
	if item.Quantity <= 0 {
		return errors.New("item quantity must be positive")
	}
	
	o.Items = append(o.Items, item)
	return o.CalculateTotal()
}

// RemoveItem removes an item from the order
func (o *Order) RemoveItem(itemID string) error {
	for i, item := range o.Items {
		if item.ID == itemID {
			o.Items = append(o.Items[:i], o.Items[i+1:]...)
			return o.CalculateTotal()
		}
	}
	return errors.New("item not found")
}

// IsExpired checks if the order has expired (pending for too long)
func (o *Order) IsExpired(timeout time.Duration) bool {
	if o.Status != OrderStatusPending {
		return false
	}
	return time.Since(o.CreatedAt) > timeout
}

// GetEstimatedCompletionTime returns the estimated completion time
func (o *Order) GetEstimatedCompletionTime() time.Time {
	if o.ConfirmedAt != nil {
		return o.ConfirmedAt.Add(time.Duration(o.EstimatedTime) * time.Second)
	}
	return o.CreatedAt.Add(time.Duration(o.EstimatedTime) * time.Second)
}

// Helper functions

// generateOrderID generates a unique order ID
func generateOrderID() string {
	// In a real implementation, this would use a proper ID generation strategy
	// For now, using timestamp-based ID
	return "order_" + time.Now().Format("20060102150405") + "_" + generateRandomString(6)
}

// generateRandomString generates a random string of given length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}
