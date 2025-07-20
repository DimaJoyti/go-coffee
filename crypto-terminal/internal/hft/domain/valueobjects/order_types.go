package valueobjects

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// OrderSide represents the side of an order
type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

// String returns the string representation
func (os OrderSide) String() string {
	return string(os)
}

// IsValid validates the order side
func (os OrderSide) IsValid() bool {
	return os == OrderSideBuy || os == OrderSideSell
}

// OrderType represents the type of an order
type OrderType string

const (
	OrderTypeMarket    OrderType = "market"
	OrderTypeLimit     OrderType = "limit"
	OrderTypeStop      OrderType = "stop"
	OrderTypeStopLimit OrderType = "stop_limit"
	OrderTypeIOC       OrderType = "ioc"      // Immediate or Cancel
	OrderTypeFOK       OrderType = "fok"      // Fill or Kill
	OrderTypePost      OrderType = "post"     // Post Only
)

// String returns the string representation
func (ot OrderType) String() string {
	return string(ot)
}

// IsValid validates the order type
func (ot OrderType) IsValid() bool {
	validTypes := []OrderType{
		OrderTypeMarket, OrderTypeLimit, OrderTypeStop,
		OrderTypeStopLimit, OrderTypeIOC, OrderTypeFOK, OrderTypePost,
	}
	
	for _, validType := range validTypes {
		if ot == validType {
			return true
		}
	}
	return false
}

// RequiresPrice returns true if the order type requires a price
func (ot OrderType) RequiresPrice() bool {
	return ot != OrderTypeMarket
}

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending          OrderStatus = "pending"
	OrderStatusNew              OrderStatus = "new"
	OrderStatusPartiallyFilled  OrderStatus = "partially_filled"
	OrderStatusFilled           OrderStatus = "filled"
	OrderStatusCanceled         OrderStatus = "canceled"
	OrderStatusRejected         OrderStatus = "rejected"
	OrderStatusExpired          OrderStatus = "expired"
)

// String returns the string representation
func (os OrderStatus) String() string {
	return string(os)
}

// IsValid validates the order status
func (os OrderStatus) IsValid() bool {
	validStatuses := []OrderStatus{
		OrderStatusPending, OrderStatusNew, OrderStatusPartiallyFilled,
		OrderStatusFilled, OrderStatusCanceled, OrderStatusRejected, OrderStatusExpired,
	}
	
	for _, validStatus := range validStatuses {
		if os == validStatus {
			return true
		}
	}
	return false
}

// IsActive returns true if the order is in an active state
func (os OrderStatus) IsActive() bool {
	return os == OrderStatusNew || os == OrderStatusPartiallyFilled
}

// IsFinal returns true if the order is in a final state
func (os OrderStatus) IsFinal() bool {
	return os == OrderStatusFilled || os == OrderStatusCanceled || 
		   os == OrderStatusRejected || os == OrderStatusExpired
}

// TimeInForce represents how long an order remains active
type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "gtc" // Good Till Canceled
	TimeInForceIOC TimeInForce = "ioc" // Immediate or Cancel
	TimeInForceFOK TimeInForce = "fok" // Fill or Kill
	TimeInForceGTD TimeInForce = "gtd" // Good Till Date
)

// String returns the string representation
func (tif TimeInForce) String() string {
	return string(tif)
}

// IsValid validates the time in force
func (tif TimeInForce) IsValid() bool {
	validTIFs := []TimeInForce{
		TimeInForceGTC, TimeInForceIOC, TimeInForceFOK, TimeInForceGTD,
	}
	
	for _, validTIF := range validTIFs {
		if tif == validTIF {
			return true
		}
	}
	return false
}

// Quantity represents an order quantity
type Quantity struct {
	decimal.Decimal
}

// NewQuantity creates a new quantity
func NewQuantity(value decimal.Decimal) (Quantity, error) {
	if value.IsNegative() || value.IsZero() {
		return Quantity{}, fmt.Errorf("quantity must be positive, got %s", value)
	}
	return Quantity{Decimal: value}, nil
}

// NewQuantityFromString creates a new quantity from string
func NewQuantityFromString(value string) (Quantity, error) {
	d, err := decimal.NewFromString(value)
	if err != nil {
		return Quantity{}, fmt.Errorf("invalid quantity string %s: %w", value, err)
	}
	return NewQuantity(d)
}

// NewQuantityFromFloat creates a new quantity from float64
func NewQuantityFromFloat(value float64) (Quantity, error) {
	d := decimal.NewFromFloat(value)
	return NewQuantity(d)
}

// IsValid validates the quantity
func (q Quantity) IsValid() bool {
	return q.IsPositive()
}

// Price represents an order price
type Price struct {
	decimal.Decimal
}

// NewPrice creates a new price
func NewPrice(value decimal.Decimal) (Price, error) {
	if value.IsNegative() {
		return Price{}, fmt.Errorf("price cannot be negative, got %s", value)
	}
	return Price{Decimal: value}, nil
}

// NewPriceFromString creates a new price from string
func NewPriceFromString(value string) (Price, error) {
	d, err := decimal.NewFromString(value)
	if err != nil {
		return Price{}, fmt.Errorf("invalid price string %s: %w", value, err)
	}
	return NewPrice(d)
}

// NewPriceFromFloat creates a new price from float64
func NewPriceFromFloat(value float64) (Price, error) {
	d := decimal.NewFromFloat(value)
	return NewPrice(d)
}

// IsValid validates the price
func (p Price) IsValid() bool {
	return !p.IsNegative()
}

// IsZero returns true if price is zero
func (p Price) IsZero() bool {
	return p.Decimal.IsZero()
}

// Commission represents trading commission
type Commission struct {
	Amount decimal.Decimal `json:"amount"`
	Asset  string          `json:"asset"`
}

// NewCommission creates a new commission
func NewCommission(amount decimal.Decimal, asset string) (Commission, error) {
	if amount.IsNegative() {
		return Commission{}, fmt.Errorf("commission amount cannot be negative, got %s", amount)
	}
	if asset == "" {
		return Commission{}, fmt.Errorf("commission asset cannot be empty")
	}
	return Commission{Amount: amount, Asset: asset}, nil
}

// IsValid validates the commission
func (c Commission) IsValid() bool {
	return !c.Amount.IsNegative() && c.Asset != ""
}

// OrderEvent represents a domain event for orders
type OrderEvent struct {
	Type      OrderEventType         `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// OrderEventType represents the type of order event
type OrderEventType string

const (
	OrderEventCreated          OrderEventType = "order_created"
	OrderEventConfirmed        OrderEventType = "order_confirmed"
	OrderEventPartiallyFilled  OrderEventType = "order_partially_filled"
	OrderEventFilled           OrderEventType = "order_filled"
	OrderEventCanceled         OrderEventType = "order_canceled"
	OrderEventRejected         OrderEventType = "order_rejected"
	OrderEventExpired          OrderEventType = "order_expired"
	OrderEventExchangeIDSet    OrderEventType = "order_exchange_id_set"
)

// String returns the string representation
func (oet OrderEventType) String() string {
	return string(oet)
}

// IsValid validates the order event type
func (oet OrderEventType) IsValid() bool {
	validTypes := []OrderEventType{
		OrderEventCreated, OrderEventConfirmed, OrderEventPartiallyFilled,
		OrderEventFilled, OrderEventCanceled, OrderEventRejected,
		OrderEventExpired, OrderEventExchangeIDSet,
	}
	
	for _, validType := range validTypes {
		if oet == validType {
			return true
		}
	}
	return false
}
