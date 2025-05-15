package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kafka_producer/config"
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
