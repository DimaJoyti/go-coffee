package kafka

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/DimaJoyti/go-coffee/accounts-service/internal/events"
	"github.com/DimaJoyti/go-coffee/accounts-service/internal/models"
	"github.com/google/uuid"
)

// EventHandlers contains all the event handlers
type EventHandlers struct {
	// We'll use simple logging for now instead of service dependencies
}

// NewEventHandlers creates a new event handlers instance
func NewEventHandlers() *EventHandlers {
	return &EventHandlers{}
}

// RegisterHandlers registers all event handlers with the consumer
func (h *EventHandlers) RegisterHandlers(consumer Consumer) {
	// Order events
	consumer.RegisterHandler(events.EventTypeOrderCreated, h.HandleOrderCreated)
	consumer.RegisterHandler(events.EventTypeOrderStatusChanged, h.HandleOrderStatusChanged)
	consumer.RegisterHandler(events.EventTypeOrderDeleted, h.HandleOrderDeleted)

	// Product events
	consumer.RegisterHandler(events.EventTypeProductCreated, h.HandleProductCreated)
	consumer.RegisterHandler(events.EventTypeProductUpdated, h.HandleProductUpdated)
	consumer.RegisterHandler(events.EventTypeProductDeleted, h.HandleProductDeleted)

	// Vendor events
	consumer.RegisterHandler(events.EventTypeVendorCreated, h.HandleVendorCreated)
	consumer.RegisterHandler(events.EventTypeVendorUpdated, h.HandleVendorUpdated)
	consumer.RegisterHandler(events.EventTypeVendorDeleted, h.HandleVendorDeleted)

	// Account events
	consumer.RegisterHandler(events.EventTypeAccountCreated, h.HandleAccountCreated)
	consumer.RegisterHandler(events.EventTypeAccountUpdated, h.HandleAccountUpdated)
	consumer.RegisterHandler(events.EventTypeAccountDeleted, h.HandleAccountDeleted)
}

// HandleOrderCreated handles the order.created event
func (h *EventHandlers) HandleOrderCreated(event events.Event) error {
	log.Printf("Handling order.created event: %s", event.ID)

	// Parse the payload
	var order models.Order
	if err := parsePayload(event.Payload, &order); err != nil {
		return err
	}

	// Process the order
	// This is just an example, in a real application you would do something with the order
	log.Printf("Order created: %s for account %s with total amount %f", order.ID, order.AccountID, order.TotalAmount)

	return nil
}

// HandleOrderStatusChanged handles the order.status_changed event
func (h *EventHandlers) HandleOrderStatusChanged(event events.Event) error {
	log.Printf("Handling order.status_changed event: %s", event.ID)

	// Parse the payload
	var payload struct {
		OrderID uuid.UUID          `json:"order_id"`
		Status  models.OrderStatus `json:"status"`
	}
	if err := parsePayload(event.Payload, &payload); err != nil {
		return err
	}

	// Process the order status change
	// This is just an example, in a real application you would do something with the order status
	log.Printf("Order %s status changed to %s", payload.OrderID, payload.Status)

	return nil
}

// HandleOrderDeleted handles the order.deleted event
func (h *EventHandlers) HandleOrderDeleted(event events.Event) error {
	log.Printf("Handling order.deleted event: %s", event.ID)

	// Parse the payload
	var payload struct {
		OrderID uuid.UUID `json:"order_id"`
	}
	if err := parsePayload(event.Payload, &payload); err != nil {
		return err
	}

	// Process the order deletion
	// This is just an example, in a real application you would do something with the order deletion
	log.Printf("Order %s deleted", payload.OrderID)

	return nil
}

// HandleProductCreated handles the product.created event
func (h *EventHandlers) HandleProductCreated(event events.Event) error {
	log.Printf("Handling product.created event: %s", event.ID)

	// Parse the payload
	var product models.Product
	if err := parsePayload(event.Payload, &product); err != nil {
		return err
	}

	// Process the product
	// This is just an example, in a real application you would do something with the product
	log.Printf("Product created: %s for vendor %s with price %f", product.ID, product.VendorID, product.Price)

	return nil
}

// HandleProductUpdated handles the product.updated event
func (h *EventHandlers) HandleProductUpdated(event events.Event) error {
	log.Printf("Handling product.updated event: %s", event.ID)

	// Parse the payload
	var product models.Product
	if err := parsePayload(event.Payload, &product); err != nil {
		return err
	}

	// Process the product update
	// This is just an example, in a real application you would do something with the product update
	log.Printf("Product updated: %s for vendor %s with price %f", product.ID, product.VendorID, product.Price)

	return nil
}

// HandleProductDeleted handles the product.deleted event
func (h *EventHandlers) HandleProductDeleted(event events.Event) error {
	log.Printf("Handling product.deleted event: %s", event.ID)

	// Parse the payload
	var payload struct {
		ProductID uuid.UUID `json:"product_id"`
	}
	if err := parsePayload(event.Payload, &payload); err != nil {
		return err
	}

	// Process the product deletion
	// This is just an example, in a real application you would do something with the product deletion
	log.Printf("Product %s deleted", payload.ProductID)

	return nil
}

// HandleVendorCreated handles the vendor.created event
func (h *EventHandlers) HandleVendorCreated(event events.Event) error {
	log.Printf("Handling vendor.created event: %s", event.ID)

	// Parse the payload
	var vendor models.Vendor
	if err := parsePayload(event.Payload, &vendor); err != nil {
		return err
	}

	// Process the vendor
	// This is just an example, in a real application you would do something with the vendor
	log.Printf("Vendor created: %s with name %s", vendor.ID, vendor.Name)

	return nil
}

// HandleVendorUpdated handles the vendor.updated event
func (h *EventHandlers) HandleVendorUpdated(event events.Event) error {
	log.Printf("Handling vendor.updated event: %s", event.ID)

	// Parse the payload
	var vendor models.Vendor
	if err := parsePayload(event.Payload, &vendor); err != nil {
		return err
	}

	// Process the vendor update
	// This is just an example, in a real application you would do something with the vendor update
	log.Printf("Vendor updated: %s with name %s", vendor.ID, vendor.Name)

	return nil
}

// HandleVendorDeleted handles the vendor.deleted event
func (h *EventHandlers) HandleVendorDeleted(event events.Event) error {
	log.Printf("Handling vendor.deleted event: %s", event.ID)

	// Parse the payload
	var payload struct {
		VendorID uuid.UUID `json:"vendor_id"`
	}
	if err := parsePayload(event.Payload, &payload); err != nil {
		return err
	}

	// Process the vendor deletion
	// This is just an example, in a real application you would do something with the vendor deletion
	log.Printf("Vendor %s deleted", payload.VendorID)

	return nil
}

// HandleAccountCreated handles the account.created event
func (h *EventHandlers) HandleAccountCreated(event events.Event) error {
	log.Printf("Handling account.created event: %s", event.ID)

	// Parse the payload
	var account models.Account
	if err := parsePayload(event.Payload, &account); err != nil {
		return err
	}

	// Process the account
	// This is just an example, in a real application you would do something with the account
	log.Printf("Account created: %s with username %s", account.ID, account.Username)

	return nil
}

// HandleAccountUpdated handles the account.updated event
func (h *EventHandlers) HandleAccountUpdated(event events.Event) error {
	log.Printf("Handling account.updated event: %s", event.ID)

	// Parse the payload
	var account models.Account
	if err := parsePayload(event.Payload, &account); err != nil {
		return err
	}

	// Process the account update
	// This is just an example, in a real application you would do something with the account update
	log.Printf("Account updated: %s with username %s", account.ID, account.Username)

	return nil
}

// HandleAccountDeleted handles the account.deleted event
func (h *EventHandlers) HandleAccountDeleted(event events.Event) error {
	log.Printf("Handling account.deleted event: %s", event.ID)

	// Parse the payload
	var payload struct {
		AccountID uuid.UUID `json:"account_id"`
	}
	if err := parsePayload(event.Payload, &payload); err != nil {
		return err
	}

	// Process the account deletion
	// This is just an example, in a real application you would do something with the account deletion
	log.Printf("Account %s deleted", payload.AccountID)

	return nil
}

// Helper function to parse the payload
func parsePayload(payload interface{}, target interface{}) error {
	// Convert the payload to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Unmarshal the JSON into the target
	if err := json.Unmarshal(payloadJSON, target); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return nil
}
