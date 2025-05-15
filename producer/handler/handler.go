package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"kafka_producer/config"
	"kafka_producer/kafka"
)

// Order struct
type Order struct {
	CustomerName string `json:"customer_name"`
	CoffeeType   string `json:"coffee_type"`
}

// Response struct
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"msg"`
}

// Handler holds dependencies for the handlers
type Handler struct {
	kafkaProducer kafka.Producer
	config        *config.Config
}

// NewHandler creates a new Handler
func NewHandler(kafkaProducer kafka.Producer, config *config.Config) *Handler {
	return &Handler{
		kafkaProducer: kafkaProducer,
		config:        config,
	}
}

// PlaceOrder handles the order placement
func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parse request body into order
	order := new(Order)
	if err := json.NewDecoder(r.Body).Decode(order); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Convert body into bytes
	orderInBytes, err := json.Marshal(order)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Send the bytes to kafka
	err = h.kafkaProducer.PushToQueue(h.config.Kafka.Topic, orderInBytes)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Respond back to the user
	response := Response{
		Success: true,
		Message: "Order for " + order.CustomerName + " placed successfully!",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println(err)
		http.Error(w, "Error placing order", http.StatusInternalServerError)
		return
	}
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
