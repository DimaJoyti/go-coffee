# Testing Guide

This guide describes how to test the Coffee Order System.

## Testing Approach

The Coffee Order System uses the following testing approaches:

1. **Unit Testing**: Testing individual components in isolation.
2. **Integration Testing**: Testing the interaction between components.
3. **End-to-End Testing**: Testing the entire system.

## Unit Testing

Unit tests are written using Go's built-in testing package. They test individual components in isolation, using mocks for dependencies.

### Running Unit Tests

To run all unit tests:

```bash
go test ./...
```

To run tests for a specific package:

```bash
go test ./handler
```

To run a specific test:

```bash
go test -run TestPlaceOrder ./handler
```

### Writing Unit Tests

Unit tests are located in files with the `_test.go` suffix. Here's an example of a unit test for the `PlaceOrder` handler:

```go
func TestPlaceOrder(t *testing.T) {
    // Create a mock producer
    mockProducer := &MockProducer{
        PushToQueueFunc: func(topic string, message []byte) error {
            return nil
        },
    }

    // Create a test configuration
    cfg := &config.Config{
        Kafka: config.KafkaConfig{
            Topic: "test_topic",
        },
    }

    // Create a handler with the mock producer
    h := NewHandler(mockProducer, cfg)

    // Create a test order
    order := Order{
        CustomerName: "Test Customer",
        CoffeeType:   "Test Coffee",
    }

    // Convert the order to JSON
    orderJSON, err := json.Marshal(order)
    if err != nil {
        t.Fatalf("Failed to marshal order: %v", err)
    }

    // Create a test request
    req, err := http.NewRequest("POST", "/order", bytes.NewBuffer(orderJSON))
    if err != nil {
        t.Fatalf("Failed to create request: %v", err)
    }
    req.Header.Set("Content-Type", "application/json")

    // Create a test response recorder
    rr := httptest.NewRecorder()

    // Call the handler
    h.PlaceOrder(rr, req)

    // Check the status code
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    // Check the response body
    var response Response
    if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
        t.Fatalf("Failed to unmarshal response: %v", err)
    }

    if !response.Success {
        t.Errorf("Handler returned wrong success value: got %v want %v", response.Success, true)
    }

    expectedMessage := "Order for Test Customer placed successfully!"
    if response.Message != expectedMessage {
        t.Errorf("Handler returned wrong message: got %v want %v", response.Message, expectedMessage)
    }
}
```

## Integration Testing

Integration tests test the interaction between components. For example, testing the interaction between the Producer service and Kafka.

### Running Integration Tests

Integration tests are not currently implemented in the Coffee Order System. However, they could be implemented using Docker Compose to set up a test environment with Kafka.

## End-to-End Testing

End-to-end tests test the entire system, from the API to the Consumer service.

### Running End-to-End Tests

End-to-end tests are not currently implemented in the Coffee Order System. However, they could be implemented using tools like Postman or custom scripts.

## Test Coverage

To check test coverage:

```bash
go test -cover ./...
```

To generate a detailed coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Mocking

The Coffee Order System uses interface-based design to make mocking easier. For example, the `kafka.Producer` interface can be mocked for testing:

```go
// MockProducer is a mock implementation of the kafka.Producer interface
type MockProducer struct {
    PushToQueueFunc func(topic string, message []byte) error
    CloseFunc       func() error
}

func (m *MockProducer) PushToQueue(topic string, message []byte) error {
    if m.PushToQueueFunc != nil {
        return m.PushToQueueFunc(topic, message)
    }
    return nil
}

func (m *MockProducer) Close() error {
    if m.CloseFunc != nil {
        return m.CloseFunc()
    }
    return nil
}
```

## Continuous Integration

The project does not currently have a CI/CD pipeline set up. However, you can run the following commands locally to ensure your changes meet the project standards:

```bash
# Format code
go fmt ./...

# Run tests
go test ./...

# Check for common mistakes
go vet ./...
```

## Next Steps

- [Development Guide](development-guide.md): Learn how to develop the system.
- [Troubleshooting](troubleshooting.md): Troubleshoot testing issues.
- [API Reference](api-reference.md): Explore the API endpoints.
