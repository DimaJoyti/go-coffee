package domain

import (
	"encoding/json"
	"time"
)

// EventType represents different types of domain events
type EventType string

const (
	// Order events
	EventTypeOrderCreated   EventType = "order.created"
	EventTypeOrderConfirmed EventType = "order.confirmed"
	EventTypeOrderPreparing EventType = "order.preparing"
	EventTypeOrderReady     EventType = "order.ready"
	EventTypeOrderCompleted EventType = "order.completed"
	EventTypeOrderCancelled EventType = "order.cancelled"
	EventTypeOrderRefunded  EventType = "order.refunded"
	
	// Payment events
	EventTypePaymentCreated   EventType = "payment.created"
	EventTypePaymentProcessing EventType = "payment.processing"
	EventTypePaymentCompleted EventType = "payment.completed"
	EventTypePaymentFailed    EventType = "payment.failed"
	EventTypePaymentRefunded  EventType = "payment.refunded"
	
	// Loyalty events
	EventTypeLoyaltyEarned EventType = "loyalty.earned"
	EventTypeLoyaltyRedeemed EventType = "loyalty.redeemed"
	
	// Crypto events
	EventTypeCryptoPaymentReceived EventType = "crypto.payment.received"
	EventTypeCryptoPaymentConfirmed EventType = "crypto.payment.confirmed"
	EventTypeTokensTransferred EventType = "tokens.transferred"
)

// DomainEvent represents a domain event
type DomainEvent struct {
	ID          string                 `json:"id"`
	Type        EventType              `json:"type"`
	AggregateID string                 `json:"aggregate_id"`
	Version     int64                  `json:"version"`
	Data        map[string]interface{} `json:"data"`
	Metadata    map[string]string      `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NewDomainEvent creates a new domain event
func NewDomainEvent(eventType EventType, aggregateID string, data map[string]interface{}) *DomainEvent {
	return &DomainEvent{
		ID:          generateEventID(),
		Type:        eventType,
		AggregateID: aggregateID,
		Version:     1,
		Data:        data,
		Metadata:    make(map[string]string),
		Timestamp:   time.Now(),
	}
}

// ToJSON converts the event to JSON
func (e *DomainEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON creates an event from JSON
func FromJSON(data []byte) (*DomainEvent, error) {
	var event DomainEvent
	err := json.Unmarshal(data, &event)
	return &event, err
}

// Order Event Factories

// NewOrderCreatedEvent creates an order created event
func NewOrderCreatedEvent(order *Order) *DomainEvent {
	data := map[string]interface{}{
		"order_id":       order.ID,
		"customer_id":    order.CustomerID,
		"total_amount":   order.TotalAmount,
		"currency":       order.Currency,
		"payment_method": order.PaymentMethod,
		"item_count":     len(order.Items),
		"is_delivery":    order.IsDelivery,
	}
	
	event := NewDomainEvent(EventTypeOrderCreated, order.ID, data)
	event.Metadata["service"] = "order-service"
	event.Metadata["customer_id"] = order.CustomerID
	
	return event
}

// NewOrderConfirmedEvent creates an order confirmed event
func NewOrderConfirmedEvent(order *Order) *DomainEvent {
	data := map[string]interface{}{
		"order_id":       order.ID,
		"customer_id":    order.CustomerID,
		"estimated_time": order.EstimatedTime,
		"priority":       order.Priority,
	}
	
	event := NewDomainEvent(EventTypeOrderConfirmed, order.ID, data)
	event.Metadata["service"] = "order-service"
	event.Metadata["customer_id"] = order.CustomerID
	
	return event
}

// NewOrderStatusChangedEvent creates an order status changed event
func NewOrderStatusChangedEvent(order *Order, previousStatus OrderStatus) *DomainEvent {
	var eventType EventType
	switch order.Status {
	case OrderStatusPreparing:
		eventType = EventTypeOrderPreparing
	case OrderStatusReady:
		eventType = EventTypeOrderReady
	case OrderStatusCompleted:
		eventType = EventTypeOrderCompleted
	case OrderStatusCancelled:
		eventType = EventTypeOrderCancelled
	case OrderStatusRefunded:
		eventType = EventTypeOrderRefunded
	default:
		eventType = EventTypeOrderConfirmed
	}
	
	data := map[string]interface{}{
		"order_id":        order.ID,
		"customer_id":     order.CustomerID,
		"new_status":      order.Status.String(),
		"previous_status": previousStatus.String(),
		"updated_at":      order.UpdatedAt,
	}
	
	event := NewDomainEvent(eventType, order.ID, data)
	event.Metadata["service"] = "order-service"
	event.Metadata["customer_id"] = order.CustomerID
	
	return event
}

// Payment Event Factories

// NewPaymentCreatedEvent creates a payment created event
func NewPaymentCreatedEvent(payment *Payment) *DomainEvent {
	data := map[string]interface{}{
		"payment_id":     payment.ID,
		"order_id":       payment.OrderID,
		"customer_id":    payment.CustomerID,
		"amount":         payment.Amount,
		"currency":       payment.Currency,
		"payment_method": payment.PaymentMethod,
	}
	
	event := NewDomainEvent(EventTypePaymentCreated, payment.ID, data)
	event.Metadata["service"] = "order-service"
	event.Metadata["customer_id"] = payment.CustomerID
	event.Metadata["order_id"] = payment.OrderID
	
	return event
}

// NewPaymentStatusChangedEvent creates a payment status changed event
func NewPaymentStatusChangedEvent(payment *Payment, previousStatus PaymentStatus) *DomainEvent {
	var eventType EventType
	switch payment.Status {
	case PaymentStatusProcessing:
		eventType = EventTypePaymentProcessing
	case PaymentStatusCompleted:
		eventType = EventTypePaymentCompleted
	case PaymentStatusFailed:
		eventType = EventTypePaymentFailed
	case PaymentStatusRefunded:
		eventType = EventTypePaymentRefunded
	default:
		eventType = EventTypePaymentCreated
	}
	
	data := map[string]interface{}{
		"payment_id":      payment.ID,
		"order_id":        payment.OrderID,
		"customer_id":     payment.CustomerID,
		"new_status":      payment.Status.String(),
		"previous_status": previousStatus.String(),
		"amount":          payment.Amount,
		"currency":        payment.Currency,
	}
	
	// Add crypto-specific data if applicable
	if payment.IsCryptoPayment() {
		data["transaction_hash"] = payment.TransactionHash
		data["crypto_network"] = payment.CryptoNetwork
		data["crypto_token"] = payment.CryptoToken
		data["wallet_address"] = payment.WalletAddress
	}
	
	event := NewDomainEvent(eventType, payment.ID, data)
	event.Metadata["service"] = "order-service"
	event.Metadata["customer_id"] = payment.CustomerID
	event.Metadata["order_id"] = payment.OrderID
	
	return event
}

// Loyalty Event Factories

// NewLoyaltyEarnedEvent creates a loyalty earned event
func NewLoyaltyEarnedEvent(customerID, orderID string, tokensEarned int64, reason string) *DomainEvent {
	data := map[string]interface{}{
		"customer_id":    customerID,
		"order_id":       orderID,
		"tokens_earned":  tokensEarned,
		"reason":         reason,
	}
	
	event := NewDomainEvent(EventTypeLoyaltyEarned, customerID, data)
	event.Metadata["service"] = "order-service"
	event.Metadata["customer_id"] = customerID
	event.Metadata["order_id"] = orderID
	
	return event
}

// NewLoyaltyRedeemedEvent creates a loyalty redeemed event
func NewLoyaltyRedeemedEvent(customerID, orderID string, tokensRedeemed int64, value int64) *DomainEvent {
	data := map[string]interface{}{
		"customer_id":     customerID,
		"order_id":        orderID,
		"tokens_redeemed": tokensRedeemed,
		"value":           value,
	}
	
	event := NewDomainEvent(EventTypeLoyaltyRedeemed, customerID, data)
	event.Metadata["service"] = "order-service"
	event.Metadata["customer_id"] = customerID
	event.Metadata["order_id"] = orderID
	
	return event
}

// Crypto Event Factories

// NewCryptoPaymentReceivedEvent creates a crypto payment received event
func NewCryptoPaymentReceivedEvent(payment *Payment) *DomainEvent {
	data := map[string]interface{}{
		"payment_id":       payment.ID,
		"order_id":         payment.OrderID,
		"customer_id":      payment.CustomerID,
		"amount":           payment.Amount,
		"currency":         payment.Currency,
		"transaction_hash": payment.TransactionHash,
		"crypto_network":   payment.CryptoNetwork,
		"crypto_token":     payment.CryptoToken,
		"wallet_address":   payment.WalletAddress,
		"block_number":     payment.BlockNumber,
	}
	
	event := NewDomainEvent(EventTypeCryptoPaymentReceived, payment.ID, data)
	event.Metadata["service"] = "order-service"
	event.Metadata["customer_id"] = payment.CustomerID
	event.Metadata["order_id"] = payment.OrderID
	event.Metadata["network"] = string(rune(payment.CryptoNetwork))
	
	return event
}

// NewTokensTransferredEvent creates a tokens transferred event
func NewTokensTransferredEvent(fromAddress, toAddress string, amount int64, tokenContract, txHash string) *DomainEvent {
	data := map[string]interface{}{
		"from_address":    fromAddress,
		"to_address":      toAddress,
		"amount":          amount,
		"token_contract":  tokenContract,
		"transaction_hash": txHash,
	}
	
	event := NewDomainEvent(EventTypeTokensTransferred, txHash, data)
	event.Metadata["service"] = "web3-service"
	event.Metadata["token_contract"] = tokenContract
	
	return event
}

// Helper functions

// generateEventID generates a unique event ID
func generateEventID() string {
	return "evt_" + time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}
