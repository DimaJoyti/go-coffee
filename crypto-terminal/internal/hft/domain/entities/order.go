package entities

import (
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/valueobjects"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Import value objects for cleaner code
type (
	OrderSide      = valueobjects.OrderSide
	OrderType      = valueobjects.OrderType
	OrderStatus    = valueobjects.OrderStatus
	TimeInForce    = valueobjects.TimeInForce
	Quantity       = valueobjects.Quantity
	Price          = valueobjects.Price
	Commission     = valueobjects.Commission
	OrderEvent     = valueobjects.OrderEvent
	OrderEventType = valueobjects.OrderEventType
)

// Import constants
const (
	OrderStatusPending         = valueobjects.OrderStatusPending
	OrderStatusNew             = valueobjects.OrderStatusNew
	OrderStatusPartiallyFilled = valueobjects.OrderStatusPartiallyFilled
	OrderStatusFilled          = valueobjects.OrderStatusFilled
	OrderStatusCanceled        = valueobjects.OrderStatusCanceled
	OrderStatusRejected        = valueobjects.OrderStatusRejected
	OrderEventCreated          = valueobjects.OrderEventCreated
	OrderEventConfirmed        = valueobjects.OrderEventConfirmed
	OrderEventPartiallyFilled  = valueobjects.OrderEventPartiallyFilled
	OrderEventCanceled         = valueobjects.OrderEventCanceled
	OrderEventRejected         = valueobjects.OrderEventRejected
	OrderEventExchangeIDSet    = valueobjects.OrderEventExchangeIDSet
)

// Order represents a trading order entity in the domain
type Order struct {
	id              OrderID
	clientOrderID   string
	strategyID      StrategyID
	symbol          Symbol
	exchange        Exchange
	side            OrderSide
	orderType       OrderType
	quantity        Quantity
	price           Price
	stopPrice       *Price
	timeInForce     TimeInForce
	status          OrderStatus
	filledQuantity  Quantity
	remainingQty    Quantity
	avgFillPrice    Price
	commission      Commission
	createdAt       time.Time
	updatedAt       time.Time
	expiresAt       *time.Time
	exchangeOrderID string
	errorMessage    string
	latency         time.Duration
	events          []OrderEvent
}

// OrderID represents a unique order identifier
type OrderID string

// StrategyID represents a strategy identifier
type StrategyID string

// Symbol represents a trading symbol
type Symbol string

// Exchange represents an exchange identifier
type Exchange string

// NewOrder creates a new order entity
func NewOrder(
	strategyID StrategyID,
	symbol Symbol,
	exchange Exchange,
	side OrderSide,
	orderType OrderType,
	quantity Quantity,
	price Price,
	timeInForce TimeInForce,
) (*Order, error) {
	if err := validateOrderParameters(strategyID, symbol, exchange, side, orderType, quantity, price); err != nil {
		return nil, fmt.Errorf("invalid order parameters: %w", err)
	}

	now := time.Now()
	order := &Order{
		id:             OrderID(generateOrderID()),
		strategyID:     strategyID,
		symbol:         symbol,
		exchange:       exchange,
		side:           side,
		orderType:      orderType,
		quantity:       quantity,
		price:          price,
		timeInForce:    timeInForce,
		status:         OrderStatusPending,
		filledQuantity: Quantity{Decimal: decimal.Zero},
		remainingQty:   quantity,
		avgFillPrice:   Price{Decimal: decimal.Zero},
		commission:     Commission{Amount: decimal.Zero, Asset: ""},
		createdAt:      now,
		updatedAt:      now,
		events:         make([]OrderEvent, 0),
	}

	// Record order creation event
	order.recordEvent(OrderEventCreated, map[string]interface{}{
		"order_id":    order.id,
		"strategy_id": order.strategyID,
		"symbol":      order.symbol,
		"side":        order.side,
		"quantity":    order.quantity,
		"price":       order.price,
	})

	return order, nil
}

// GetID returns the order ID
func (o *Order) GetID() OrderID {
	return o.id
}

// GetStrategyID returns the strategy ID
func (o *Order) GetStrategyID() StrategyID {
	return o.strategyID
}

// GetSymbol returns the trading symbol
func (o *Order) GetSymbol() Symbol {
	return o.symbol
}

// GetExchange returns the exchange
func (o *Order) GetExchange() Exchange {
	return o.exchange
}

// GetSide returns the order side
func (o *Order) GetSide() OrderSide {
	return o.side
}

// GetOrderType returns the order type
func (o *Order) GetOrderType() OrderType {
	return o.orderType
}

// GetQuantity returns the order quantity
func (o *Order) GetQuantity() Quantity {
	return o.quantity
}

// GetPrice returns the order price
func (o *Order) GetPrice() Price {
	return o.price
}

// GetStatus returns the order status
func (o *Order) GetStatus() OrderStatus {
	return o.status
}

// GetFilledQuantity returns the filled quantity
func (o *Order) GetFilledQuantity() Quantity {
	return o.filledQuantity
}

// GetRemainingQuantity returns the remaining quantity
func (o *Order) GetRemainingQuantity() Quantity {
	return o.remainingQty
}

// GetAvgFillPrice returns the average fill price
func (o *Order) GetAvgFillPrice() Price {
	return o.avgFillPrice
}

// GetCommission returns the commission
func (o *Order) GetCommission() Commission {
	return o.commission
}

// GetCreatedAt returns the creation timestamp
func (o *Order) GetCreatedAt() time.Time {
	return o.createdAt
}

// GetUpdatedAt returns the last update timestamp
func (o *Order) GetUpdatedAt() time.Time {
	return o.updatedAt
}

// GetLatency returns the order latency
func (o *Order) GetLatency() time.Duration {
	return o.latency
}

// GetEvents returns the order events
func (o *Order) GetEvents() []OrderEvent {
	return o.events
}

// SetExchangeOrderID sets the exchange order ID
func (o *Order) SetExchangeOrderID(exchangeOrderID string) {
	o.exchangeOrderID = exchangeOrderID
	o.updatedAt = time.Now()

	o.recordEvent(OrderEventExchangeIDSet, map[string]interface{}{
		"exchange_order_id": exchangeOrderID,
	})
}

// Confirm confirms the order placement
func (o *Order) Confirm() error {
	if o.status != OrderStatusPending {
		return fmt.Errorf("cannot confirm order in status %s", o.status)
	}

	o.status = OrderStatusNew
	o.updatedAt = time.Now()

	o.recordEvent(OrderEventConfirmed, map[string]interface{}{
		"status": o.status,
	})

	return nil
}

// PartialFill processes a partial fill
func (o *Order) PartialFill(fillQuantity Quantity, fillPrice Price, commission Commission) error {
	if o.status != OrderStatusNew && o.status != OrderStatusPartiallyFilled {
		return fmt.Errorf("cannot fill order in status %s", o.status)
	}

	if fillQuantity.GreaterThan(o.remainingQty.Decimal) {
		return fmt.Errorf("fill quantity %s exceeds remaining quantity %s", fillQuantity, o.remainingQty)
	}

	// Update fill information
	totalFilled := o.filledQuantity.Add(fillQuantity.Decimal)
	totalValue := o.avgFillPrice.Mul(o.filledQuantity.Decimal).Add(fillPrice.Mul(fillQuantity.Decimal))

	o.filledQuantity = Quantity{Decimal: totalFilled}
	o.remainingQty = Quantity{Decimal: o.quantity.Sub(totalFilled)}
	o.avgFillPrice = Price{Decimal: totalValue.Div(totalFilled)}
	o.commission = Commission{
		Amount: o.commission.Amount.Add(commission.Amount),
		Asset:  commission.Asset,
	}

	// Update status
	if o.remainingQty.IsZero() {
		o.status = OrderStatusFilled
	} else {
		o.status = OrderStatusPartiallyFilled
	}

	o.updatedAt = time.Now()

	o.recordEvent(OrderEventPartiallyFilled, map[string]interface{}{
		"fill_quantity": fillQuantity,
		"fill_price":    fillPrice,
		"filled_total":  o.filledQuantity,
		"remaining":     o.remainingQty,
		"status":        o.status,
	})

	return nil
}

// Cancel cancels the order
func (o *Order) Cancel() error {
	if o.status == OrderStatusFilled || o.status == OrderStatusCanceled {
		return fmt.Errorf("cannot cancel order in status %s", o.status)
	}

	o.status = OrderStatusCanceled
	o.updatedAt = time.Now()

	o.recordEvent(OrderEventCanceled, map[string]interface{}{
		"status": o.status,
	})

	return nil
}

// Reject rejects the order
func (o *Order) Reject(reason string) {
	o.status = OrderStatusRejected
	o.errorMessage = reason
	o.updatedAt = time.Now()

	o.recordEvent(OrderEventRejected, map[string]interface{}{
		"status": o.status,
		"reason": reason,
	})
}

// SetLatency sets the order processing latency
func (o *Order) SetLatency(latency time.Duration) {
	o.latency = latency
	o.updatedAt = time.Now()
}

// IsActive returns true if the order is active (can be filled or canceled)
func (o *Order) IsActive() bool {
	return o.status == OrderStatusNew || o.status == OrderStatusPartiallyFilled
}

// IsFilled returns true if the order is completely filled
func (o *Order) IsFilled() bool {
	return o.status == OrderStatusFilled
}

// IsCanceled returns true if the order is canceled
func (o *Order) IsCanceled() bool {
	return o.status == OrderStatusCanceled
}

// IsRejected returns true if the order is rejected
func (o *Order) IsRejected() bool {
	return o.status == OrderStatusRejected
}

// recordEvent records a domain event
func (o *Order) recordEvent(eventType OrderEventType, data map[string]interface{}) {
	event := OrderEvent{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}
	o.events = append(o.events, event)
}

// generateOrderID generates a unique order ID
func generateOrderID() string {
	return uuid.New().String()
}

// validateOrderParameters validates order creation parameters
func validateOrderParameters(
	strategyID StrategyID,
	symbol Symbol,
	exchange Exchange,
	side OrderSide,
	orderType OrderType,
	quantity Quantity,
	price Price,
) error {
	if strategyID == "" {
		return fmt.Errorf("strategy ID cannot be empty")
	}
	if symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}
	if exchange == "" {
		return fmt.Errorf("exchange cannot be empty")
	}
	if !side.IsValid() {
		return fmt.Errorf("invalid order side: %s", side)
	}
	if !orderType.IsValid() {
		return fmt.Errorf("invalid order type: %s", orderType)
	}
	if !quantity.IsValid() {
		return fmt.Errorf("invalid quantity: %s", quantity)
	}
	if orderType.RequiresPrice() && !price.IsValid() {
		return fmt.Errorf("invalid price for order type %s: %s", orderType, price)
	}
	return nil
}
