package kafka

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/accounts-service/internal/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockKafkaConsumer is a mock implementation of the Consumer interface
type MockKafkaConsumer struct {
	mock.Mock
	handlers map[events.EventType]events.EventHandler
	mu       sync.RWMutex
}

func NewMockKafkaConsumer() *MockKafkaConsumer {
	return &MockKafkaConsumer{
		handlers: make(map[events.EventType]events.EventHandler),
	}
}

func (m *MockKafkaConsumer) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockKafkaConsumer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockKafkaConsumer) RegisterHandler(eventType events.EventType, handler events.EventHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[eventType] = handler
	m.Called(eventType, handler)
}

// SimulateEvent simulates receiving an event
func (m *MockKafkaConsumer) SimulateEvent(event events.Event) error {
	m.mu.RLock()
	handler, ok := m.handlers[event.Type]
	m.mu.RUnlock()

	if !ok {
		return nil
	}

	return handler(event)
}

func TestKafkaConsumer_RegisterHandler(t *testing.T) {
	// Create mock consumer
	mockConsumer := NewMockKafkaConsumer()

	// Set up expectations
	mockConsumer.On("RegisterHandler", events.EventTypeAccountCreated, mock.AnythingOfType("func(events.Event) error")).Return()

	// Create a handler
	handler := func(event events.Event) error {
		return nil
	}

	// Register the handler
	mockConsumer.RegisterHandler(events.EventTypeAccountCreated, handler)

	// Verify expectations
	mockConsumer.AssertExpectations(t)
}

func TestKafkaConsumer_SimulateEvent(t *testing.T) {
	// Create mock consumer
	mockConsumer := NewMockKafkaConsumer()

	// Create a channel to signal that the handler was called
	handlerCalled := make(chan bool, 1)

	// Create a handler
	handler := func(event events.Event) error {
		// Check that the event is correct
		assert.Equal(t, events.EventTypeAccountCreated, event.Type)
		assert.Equal(t, "test-id", event.ID)

		// Signal that the handler was called
		handlerCalled <- true
		return nil
	}

	// Set up expectations
	mockConsumer.On("RegisterHandler", events.EventTypeAccountCreated, mock.AnythingOfType("func(events.Event) error")).Return()

	// Register the handler
	mockConsumer.RegisterHandler(events.EventTypeAccountCreated, handler)

	// Create an event
	event := events.Event{
		ID:        "test-id",
		Type:      events.EventTypeAccountCreated,
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{"id": "123", "username": "testuser"},
	}

	// Simulate receiving the event
	err := mockConsumer.SimulateEvent(event)
	assert.NoError(t, err)

	// Wait for the handler to be called
	select {
	case <-handlerCalled:
		// Handler was called, test passes
	case <-time.After(time.Second):
		t.Fatal("Handler was not called within timeout")
	}

	// Verify expectations
	mockConsumer.AssertExpectations(t)
}

func TestEventHandlers_RegisterHandlers(t *testing.T) {
	// Create mock consumer
	mockConsumer := NewMockKafkaConsumer()

	// Set up expectations for all event types
	mockConsumer.On("RegisterHandler", events.EventTypeOrderCreated, mock.AnythingOfType("func(events.Event) error")).Return()
	mockConsumer.On("RegisterHandler", events.EventTypeOrderStatusChanged, mock.AnythingOfType("func(events.Event) error")).Return()
	mockConsumer.On("RegisterHandler", events.EventTypeOrderDeleted, mock.AnythingOfType("func(events.Event) error")).Return()
	mockConsumer.On("RegisterHandler", events.EventTypeProductCreated, mock.AnythingOfType("func(events.Event) error")).Return()
	mockConsumer.On("RegisterHandler", events.EventTypeProductUpdated, mock.AnythingOfType("func(events.Event) error")).Return()
	mockConsumer.On("RegisterHandler", events.EventTypeProductDeleted, mock.AnythingOfType("func(events.Event) error")).Return()
	mockConsumer.On("RegisterHandler", events.EventTypeVendorCreated, mock.AnythingOfType("func(events.Event) error")).Return()
	mockConsumer.On("RegisterHandler", events.EventTypeVendorUpdated, mock.AnythingOfType("func(events.Event) error")).Return()
	mockConsumer.On("RegisterHandler", events.EventTypeVendorDeleted, mock.AnythingOfType("func(events.Event) error")).Return()
	mockConsumer.On("RegisterHandler", events.EventTypeAccountCreated, mock.AnythingOfType("func(events.Event) error")).Return()
	mockConsumer.On("RegisterHandler", events.EventTypeAccountUpdated, mock.AnythingOfType("func(events.Event) error")).Return()
	mockConsumer.On("RegisterHandler", events.EventTypeAccountDeleted, mock.AnythingOfType("func(events.Event) error")).Return()

	// Create event handlers (simplified, no service dependencies)
	eventHandlers := NewEventHandlers()

	// Register handlers
	eventHandlers.RegisterHandlers(mockConsumer)

	// Verify expectations
	mockConsumer.AssertExpectations(t)
}
