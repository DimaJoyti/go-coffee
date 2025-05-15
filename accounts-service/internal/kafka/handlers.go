package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/models"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/service"
)

// EventHandlers contains all the event handlers
type EventHandlers struct {
	accountService *service.AccountService
	orderService   *service.OrderService
	productService *service.ProductService
	vendorService  *service.VendorService
}

// NewEventHandlers creates a new event handlers instance
func NewEventHandlers(
	accountService *service.AccountService,
	orderService *service.OrderService,
	productService *service.ProductService,
	vendorService *service.VendorService,
) *EventHandlers {
	return &EventHandlers{
		accountService: accountService,
		orderService:   orderService,
		productService: productService,
		vendorService:  vendorService,
	}
}

// RegisterHandlers registers all event handlers with the consumer
func (h *EventHandlers) RegisterHandlers(consumer Consumer) {
	// Order events
	consumer.RegisterHandler(EventTypeOrderCreated, h.HandleOrderCreated)
	consumer.RegisterHandler(EventTypeOrderStatusChanged, h.HandleOrderStatusChanged)
	consumer.RegisterHandler(EventTypeOrderDeleted, h.HandleOrderDeleted)

	// Product events
	consumer.RegisterHandler(EventTypeProductCreated, h.HandleProductCreated)
	consumer.RegisterHandler(EventTypeProductUpdated, h.HandleProductUpdated)
	consumer.RegisterHandler(EventTypeProductDeleted, h.HandleProductDeleted)

	// Vendor events
	consumer.RegisterHandler(EventTypeVendorCreated, h.HandleVendorCreated)
	consumer.RegisterHandler(EventTypeVendorUpdated, h.HandleVendorUpdated)
	consumer.RegisterHandler(EventTypeVendorDeleted, h.HandleVendorDeleted)

	// Account events
	consumer.RegisterHandler(EventTypeAccountCreated, h.HandleAccountCreated)
	consumer.RegisterHandler(EventTypeAccountUpdated, h.HandleAccountUpdated)
	consumer.RegisterHandler(EventTypeAccountDeleted, h.HandleAccountDeleted)
}

// HandleOrderCreated handles the order.created event
func (h *EventHandlers) HandleOrderCreated(event Event) error {
	log.Printf("Handling order.created event: %s", event.ID)

	// Parse the payload
	var order models.Order
	if err := parsePayload(event.Payload, &order); err != nil {
		return err
	}

	// Process the order
	ctx := context.Background()
	// This is just an example, in a real application you would do something with the order
	log.Printf("Order created: %s for account %s with total amount %f", order.ID, order.AccountID, order.TotalAmount)

	return nil
}

// HandleOrderStatusChanged handles the order.status_changed event
func (h *EventHandlers) HandleOrderStatusChanged(event Event) error {
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
	ctx := context.Background()
	// This is just an example, in a real application you would do something with the order status
	log.Printf("Order %s status changed to %s", payload.OrderID, payload.Status)

	return nil
}

// HandleOrderDeleted handles the order.deleted event
func (h *EventHandlers) HandleOrderDeleted(event Event) error {
	log.Printf("Handling order.deleted event: %s", event.ID)

	// Parse the payload
	var payload struct {
		OrderID uuid.UUID `json:"order_id"`
	}
	if err := parsePayload(event.Payload, &payload); err != nil {
		return err
	}

	// Process the order deletion
	ctx := context.Background()
	// This is just an example, in a real application you would do something with the order deletion
	log.Printf("Order %s deleted", payload.OrderID)

	return nil
}

// HandleProductCreated handles the product.created event
func (h *EventHandlers) HandleProductCreated(event Event) error {
	log.Printf("Handling product.created event: %s", event.ID)

	// Parse the payload
	var product models.Product
	if err := parsePayload(event.Payload, &product); err != nil {
		return err
	}

	// Process the product
	ctx := context.Background()
	// This is just an example, in a real application you would do something with the product
	log.Printf("Product created: %s for vendor %s with price %f", product.ID, product.VendorID, product.Price)

	return nil
}

// HandleProductUpdated handles the product.updated event
func (h *EventHandlers) HandleProductUpdated(event Event) error {
	log.Printf("Handling product.updated event: %s", event.ID)

	// Parse the payload
	var product models.Product
	if err := parsePayload(event.Payload, &product); err != nil {
		return err
	}

	// Process the product update
	ctx := context.Background()
	// This is just an example, in a real application you would do something with the product update
	log.Printf("Product updated: %s for vendor %s with price %f", product.ID, product.VendorID, product.Price)

	return nil
}

// HandleProductDeleted handles the product.deleted event
func (h *EventHandlers) HandleProductDeleted(event Event) error {
	log.Printf("Handling product.deleted event: %s", event.ID)

	// Parse the payload
	var payload struct {
		ProductID uuid.UUID `json:"product_id"`
	}
	if err := parsePayload(event.Payload, &payload); err != nil {
		return err
	}

	// Process the product deletion
	ctx := context.Background()
	// This is just an example, in a real application you would do something with the product deletion
	log.Printf("Product %s deleted", payload.ProductID)

	return nil
}

// HandleVendorCreated handles the vendor.created event
func (h *EventHandlers) HandleVendorCreated(event Event) error {
	log.Printf("Handling vendor.created event: %s", event.ID)

	// Parse the payload
	var vendor models.Vendor
	if err := parsePayload(event.Payload, &vendor); err != nil {
		return err
	}

	// Process the vendor
	ctx := context.Background()
	// This is just an example, in a real application you would do something with the vendor
	log.Printf("Vendor created: %s with name %s", vendor.ID, vendor.Name)

	return nil
}

// HandleVendorUpdated handles the vendor.updated event
func (h *EventHandlers) HandleVendorUpdated(event Event) error {
	log.Printf("Handling vendor.updated event: %s", event.ID)

	// Parse the payload
	var vendor models.Vendor
	if err := parsePayload(event.Payload, &vendor); err != nil {
		return err
	}

	// Process the vendor update
	ctx := context.Background()
	// This is just an example, in a real application you would do something with the vendor update
	log.Printf("Vendor updated: %s with name %s", vendor.ID, vendor.Name)

	return nil
}

// HandleVendorDeleted handles the vendor.deleted event
func (h *EventHandlers) HandleVendorDeleted(event Event) error {
	log.Printf("Handling vendor.deleted event: %s", event.ID)

	// Parse the payload
	var payload struct {
		VendorID uuid.UUID `json:"vendor_id"`
	}
	if err := parsePayload(event.Payload, &payload); err != nil {
		return err
	}

	// Process the vendor deletion
	ctx := context.Background()
	// This is just an example, in a real application you would do something with the vendor deletion
	log.Printf("Vendor %s deleted", payload.VendorID)

	return nil
}

// HandleAccountCreated handles the account.created event
func (h *EventHandlers) HandleAccountCreated(event Event) error {
	log.Printf("Handling account.created event: %s", event.ID)

	// Parse the payload
	var account models.Account
	if err := parsePayload(event.Payload, &account); err != nil {
		return err
	}

	// Process the account
	ctx := context.Background()
	// This is just an example, in a real application you would do something with the account
	log.Printf("Account created: %s with username %s", account.ID, account.Username)

	return nil
}

// HandleAccountUpdated handles the account.updated event
func (h *EventHandlers) HandleAccountUpdated(event Event) error {
	log.Printf("Handling account.updated event: %s", event.ID)

	// Parse the payload
	var account models.Account
	if err := parsePayload(event.Payload, &account); err != nil {
		return err
	}

	// Process the account update
	ctx := context.Background()
	// This is just an example, in a real application you would do something with the account update
	log.Printf("Account updated: %s with username %s", account.ID, account.Username)

	return nil
}

// HandleAccountDeleted handles the account.deleted event
func (h *EventHandlers) HandleAccountDeleted(event Event) error {
	log.Printf("Handling account.deleted event: %s", event.ID)

	// Parse the payload
	var payload struct {
		AccountID uuid.UUID `json:"account_id"`
	}
	if err := parsePayload(event.Payload, &payload); err != nil {
		return err
	}

	// Process the account deletion
	ctx := context.Background()
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
