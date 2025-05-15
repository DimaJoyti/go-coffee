# Kafka Integration

The Accounts Service integrates with Kafka for event-driven communication with other services in the system.

## Event Types

The service handles the following event types:

### Account Events

- `account.created`: Triggered when a new account is created
- `account.updated`: Triggered when an account is updated
- `account.deleted`: Triggered when an account is deleted

### Vendor Events

- `vendor.created`: Triggered when a new vendor is created
- `vendor.updated`: Triggered when a vendor is updated
- `vendor.deleted`: Triggered when a vendor is deleted

### Product Events

- `product.created`: Triggered when a new product is created
- `product.updated`: Triggered when a product is updated
- `product.deleted`: Triggered when a product is deleted

### Order Events

- `order.created`: Triggered when a new order is created
- `order.status_changed`: Triggered when an order's status changes
- `order.deleted`: Triggered when an order is deleted

## Event Format

Events are published to Kafka in the following format:

```json
{
  "id": "unique-event-id",
  "type": "event.type",
  "timestamp": "2023-05-15T12:34:56Z",
  "payload": {
    // Event-specific data
  }
}
```

## Publishing Events

The service publishes events to Kafka when changes occur to accounts, vendors, products, or orders. For example, when a new account is created, an `account.created` event is published with the account details in the payload.

Example of publishing an event:

```go
// Create a new account
account, err := accountService.Create(ctx, input)
if err != nil {
    return nil, err
}

// Publish an event
err = kafkaProducer.Publish(kafka.EventTypeAccountCreated, account)
if err != nil {
    log.Printf("Failed to publish account.created event: %v", err)
}
```

## Consuming Events

The service consumes events from Kafka to react to changes in other services. For example, when a product is updated in another service, the Accounts Service receives an `product.updated` event and updates its local copy of the product.

Event handlers are registered for each event type:

```go
// Create event handlers
eventHandlers := kafka.NewEventHandlers(accountService, orderService, productService, vendorService)

// Register event handlers
eventHandlers.RegisterHandlers(kafkaConsumer)

// Start the Kafka consumer
go func() {
    if err := kafkaConsumer.Start(ctx); err != nil {
        log.Fatalf("Failed to start Kafka consumer: %v", err)
    }
}()
```

## Configuration

Kafka integration can be configured using environment variables or a configuration file. See the [configuration documentation](configuration.md) for details.

## Event Handlers

Event handlers are implemented in the `internal/kafka/handlers.go` file. Each handler processes a specific event type and updates the service's state accordingly.

Example of an event handler:

```go
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
```

## Testing

Event handlers can be tested using the `MockKafkaConsumer` in the `internal/kafka/consumer_test.go` file. This mock allows you to simulate receiving events and verify that they are handled correctly.

Example of testing an event handler:

```go
func TestHandleProductCreated(t *testing.T) {
    // Create mock consumer
    mockConsumer := NewMockKafkaConsumer()

    // Create mock services
    mockAccountService := &MockAccountService{}
    mockOrderService := &MockOrderService{}
    mockProductService := &MockProductService{}
    mockVendorService := &MockVendorService{}

    // Create event handlers
    eventHandlers := NewEventHandlers(mockAccountService, mockOrderService, mockProductService, mockVendorService)

    // Register handlers
    eventHandlers.RegisterHandlers(mockConsumer)

    // Create an event
    event := Event{
        ID:        "test-id",
        Type:      EventTypeProductCreated,
        Timestamp: time.Now(),
        Payload:   map[string]interface{}{
            "id": "123",
            "vendor_id": "456",
            "name": "Test Product",
            "price": 9.99,
        },
    }

    // Simulate receiving the event
    err := mockConsumer.SimulateEvent(event)
    assert.NoError(t, err)

    // Verify that the product was processed
    // In a real test, you would verify that the product was added to the database
}
```
