package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DimaJoyti/go-coffee/producer/config"
	"github.com/DimaJoyti/go-coffee/producer/store"
)

// MockProducer is a mock implementation of the kafka.Producer interface
type MockProducer struct {
	PushToQueueFunc      func(topic string, message []byte) error
	PushToQueueAsyncFunc func(topic string, message []byte) error
	CloseFunc            func() error
	FlushFunc            func() error
}

func (m *MockProducer) PushToQueue(topic string, message []byte) error {
	if m.PushToQueueFunc != nil {
		return m.PushToQueueFunc(topic, message)
	}
	return nil
}

func (m *MockProducer) PushToQueueAsync(topic string, message []byte) error {
	if m.PushToQueueAsyncFunc != nil {
		return m.PushToQueueAsyncFunc(topic, message)
	}
	return nil
}

func (m *MockProducer) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockProducer) Flush() error {
	if m.FlushFunc != nil {
		return m.FlushFunc()
	}
	return nil
}

// MockOrderStore is a mock implementation of the store.OrderStore interface
type MockOrderStore struct {
	orders map[string]*store.Order
}

func NewMockOrderStore() *MockOrderStore {
	return &MockOrderStore{
		orders: make(map[string]*store.Order),
	}
}

func (m *MockOrderStore) Add(order *store.Order) error {
	m.orders[order.ID] = order
	return nil
}

func (m *MockOrderStore) Get(id string) (*store.Order, error) {
	if order, exists := m.orders[id]; exists {
		return order, nil
	}
	return nil, errors.New("order not found")
}

func (m *MockOrderStore) Update(order *store.Order) error {
	m.orders[order.ID] = order
	return nil
}

func (m *MockOrderStore) Delete(id string) error {
	delete(m.orders, id)
	return nil
}

func (m *MockOrderStore) List() ([]*store.Order, error) {
	orders := make([]*store.Order, 0, len(m.orders))
	for _, order := range m.orders {
		orders = append(orders, order)
	}
	return orders, nil
}

func (m *MockOrderStore) ListByStatus(status store.OrderStatus) ([]*store.Order, error) {
	orders := make([]*store.Order, 0)
	for _, order := range m.orders {
		if order.Status == status {
			orders = append(orders, order)
		}
	}
	return orders, nil
}

func (m *MockOrderStore) ListByCustomer(customerName string) ([]*store.Order, error) {
	orders := make([]*store.Order, 0)
	for _, order := range m.orders {
		if order.CustomerName == customerName {
			orders = append(orders, order)
		}
	}
	return orders, nil
}

func TestPlaceOrder(t *testing.T) {
	tests := []struct {
		name          string
		order         OrderRequest
		expectedCode  int
		expectedError bool
	}{
		{
			name: "Valid order",
			order: OrderRequest{
				CustomerName: "Test Customer",
				CoffeeType:   "Test Coffee",
			},
			expectedCode:  http.StatusOK,
			expectedError: false,
		},
		{
			name: "Invalid order - empty customer",
			order: OrderRequest{
				CustomerName: "",
				CoffeeType:   "Test Coffee",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: true,
		},
		{
			name: "Invalid order - empty coffee type",
			order: OrderRequest{
				CustomerName: "Test Customer",
				CoffeeType:   "",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProducer := &MockProducer{
				PushToQueueFunc: func(topic string, message []byte) error {
					return nil
				},
			}

			mockOrderStore := NewMockOrderStore()

			cfg := &config.Config{
				Kafka: config.KafkaConfig{
					Topic: "test_topic",
				},
			}

			h := NewHandler(mockProducer, cfg, mockOrderStore)

			orderJSON, err := json.Marshal(tt.order)
			if err != nil {
				t.Fatalf("Failed to marshal order: %v", err)
			}

			req, err := http.NewRequest("POST", "/order", bytes.NewBuffer(orderJSON))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			h.PlaceOrder(rr, req)

			if status := rr.Code; status != tt.expectedCode {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, tt.expectedCode)
			}

			var response OrderResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if tt.expectedError && response.Success {
				t.Error("Expected error response but got success")
			}
		})
	}
}

func TestHealthCheck(t *testing.T) {
	// Create a mock producer
	mockProducer := &MockProducer{}

	// Create a mock order store
	mockOrderStore := NewMockOrderStore()

	// Create a test configuration
	cfg := &config.Config{}

	// Create a handler with the mock producer
	h := NewHandler(mockProducer, cfg, mockOrderStore)

	// Create a test request
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a test response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	h.HealthCheck(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if status, ok := response["status"]; !ok || status != "ok" {
		t.Errorf("Handler returned wrong status: got %v want %v", status, "ok")
	}
}
