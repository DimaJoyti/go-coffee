package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dimasudakov/go-coffee/producer/config"
	"github.com/dimasudakov/go-coffee/producer/model"
)

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

func TestPlaceOrder(t *testing.T) {
	tests := []struct {
		name           string
		order         model.Order
		expectedCode   int
		expectedError  bool
	}{
		{
			name: "Valid order",
			order: model.Order{
				CustomerName: "Test Customer",
				CoffeeType:   "Test Coffee",
				Quantity:     1,
			},
			expectedCode:  http.StatusOK,
			expectedError: false,
		},
		{
			name: "Invalid order - empty customer",
			order: model.Order{
				CustomerName: "",
				CoffeeType:   "Test Coffee",
				Quantity:     1,
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: true,
		},
		{
			name: "Invalid order - empty coffee type",
			order: model.Order{
				CustomerName: "Test Customer",
				CoffeeType:   "",
				Quantity:     1,
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

			cfg := &config.Config{
				Kafka: config.KafkaConfig{
					Topic: "test_topic",
				},
			}

			h := NewHandler(mockProducer, cfg)

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

			var response Response
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

	// Create a test configuration
	cfg := &config.Config{}

	// Create a handler with the mock producer
	h := NewHandler(mockProducer, cfg)

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
