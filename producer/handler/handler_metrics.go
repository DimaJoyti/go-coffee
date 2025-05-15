package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"kafka_producer/config"
	"kafka_producer/kafka"
	"kafka_producer/metrics"
	"kafka_producer/store"
)

// OrderRequest represents a request to place an order
type OrderRequest struct {
	CustomerName string `json:"customer_name"`
	CoffeeType   string `json:"coffee_type"`
}

// OrderResponse represents a response to an order request
type OrderResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"msg"`
	Order   *store.Order `json:"order,omitempty"`
}

// Handler holds dependencies for the handlers
type Handler struct {
	kafkaProducer kafka.Producer
	config        *config.Config
	orderStore    store.OrderStore
}

// NewHandler creates a new Handler
func NewHandler(kafkaProducer kafka.Producer, config *config.Config, orderStore store.OrderStore) *Handler {
	return &Handler{
		kafkaProducer: kafkaProducer,
		config:        config,
		orderStore:    orderStore,
	}
}

// PlaceOrder handles the order placement
func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	metrics.OrdersTotal.Inc()
	metrics.HttpRequestsTotal.WithLabelValues(r.Method, "/order", "").Inc()
	defer func() {
		metrics.HttpRequestDuration.WithLabelValues(r.Method, "/order").Observe(time.Since(startTime).Seconds())
	}()

	if r.Method != http.MethodPost {
		metrics.HttpRequestsTotal.WithLabelValues(r.Method, "/order", "405").Inc()
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parse request body into order request
	orderReq := new(OrderRequest)
	if err := json.NewDecoder(r.Body).Decode(orderReq); err != nil {
		metrics.OrdersFailedTotal.Inc()
		metrics.HttpRequestsTotal.WithLabelValues(r.Method, "/order", "400").Inc()
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Create a new order
	now := time.Now()
	order := &store.Order{
		ID:           uuid.New().String(),
		CustomerName: orderReq.CustomerName,
		CoffeeType:   orderReq.CoffeeType,
		Status:       store.OrderStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 3. Add the order to the store
	if err := h.orderStore.Add(order); err != nil {
		metrics.OrdersFailedTotal.Inc()
		metrics.HttpRequestsTotal.WithLabelValues(r.Method, "/order", "500").Inc()
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Convert order to JSON
	orderJSON, err := json.Marshal(order)
	if err != nil {
		metrics.OrdersFailedTotal.Inc()
		metrics.HttpRequestsTotal.WithLabelValues(r.Method, "/order", "500").Inc()
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Send the order to Kafka
	err = h.kafkaProducer.PushToQueue(h.config.Kafka.Topic, orderJSON)
	if err != nil {
		metrics.OrdersFailedTotal.Inc()
		metrics.KafkaMessagesFailedTotal.Inc()
		metrics.HttpRequestsTotal.WithLabelValues(r.Method, "/order", "500").Inc()
		log.Println(err)
		// Remove the order from the store if we can't send it to Kafka
		h.orderStore.Delete(order.ID)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	metrics.OrdersSuccessTotal.Inc()
	metrics.KafkaMessagesSentTotal.Inc()
	metrics.OrderProcessingTime.Observe(time.Since(startTime).Seconds())
	metrics.HttpRequestsTotal.WithLabelValues(r.Method, "/order", "200").Inc()

	// 6. Respond back to the user
	response := OrderResponse{
		Success: true,
		Message: "Order for " + order.CustomerName + " placed successfully!",
		Order:   order,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		metrics.HttpRequestsTotal.WithLabelValues(r.Method, "/order", "500").Inc()
		log.Println(err)
		http.Error(w, "Error placing order", http.StatusInternalServerError)
		return
	}
}

// GetOrder handles retrieving an order
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	metrics.HttpRequestsTotal.WithLabelValues(r.Method, "/order/{id}", "").Inc()
	defer func() {
		metrics.HttpRequestDuration.WithLabelValues(r.Method, "/order/{id}").Observe(time.Since(startTime).Seconds())
	}()

	if r.Method != http.MethodGet {
		metrics.HttpRequestsTotal.WithLabelValues(r.Method, "/order/{id}", "405").Inc()
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Extract order ID from URL
	orderID := strings.TrimPrefix(r.URL.Path, "/order/")
	if orderID == "" {
		metrics.HttpRequestsTotal.WithLabelValues(r.Method, "/order/{id}", "400").Inc()
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	// Get the order from the store
	order, err := h.orderStore.Get(orderID)
	if err != nil {
		metrics.HttpRequestsTotal.WithLabelValues(r.Method, "/order/{id}", "404").Inc()
		log.Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	metrics.HttpRequestsTotal.WithLabelValues(r.Method, "/order/{id}", "200").Inc()

	// Respond with the order
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		metrics.HttpRequestsTotal.WithLabelValues(r.Method, "/order/{id}", "500").Inc()
		log.Println(err)
		http.Error(w, "Error retrieving order", http.StatusInternalServerError)
		return
	}
}

// Інші методи обробника також можна оновити з метриками, але для прикладу я показав лише два основних методи
