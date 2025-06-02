package events

import (
	"time"
)

// EventType represents the type of event
type EventType string

const (
	// Account events
	EventTypeAccountCreated EventType = "account.created"
	EventTypeAccountUpdated EventType = "account.updated"
	EventTypeAccountDeleted EventType = "account.deleted"

	// Vendor events
	EventTypeVendorCreated EventType = "vendor.created"
	EventTypeVendorUpdated EventType = "vendor.updated"
	EventTypeVendorDeleted EventType = "vendor.deleted"

	// Product events
	EventTypeProductCreated EventType = "product.created"
	EventTypeProductUpdated EventType = "product.updated"
	EventTypeProductDeleted EventType = "product.deleted"

	// Order events
	EventTypeOrderCreated       EventType = "order.created"
	EventTypeOrderStatusChanged EventType = "order.status_changed"
	EventTypeOrderDeleted       EventType = "order.deleted"
)

// Event represents a Kafka event
type Event struct {
	ID        string      `json:"id"`
	Type      EventType   `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

// EventHandler is a function that handles an event
type EventHandler func(event Event) error
