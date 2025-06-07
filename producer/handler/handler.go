package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/DimaJoyti/go-coffee/producer/config"
	"github.com/DimaJoyti/go-coffee/producer/kafka"
	"github.com/DimaJoyti/go-coffee/producer/store"
)

// OrderRequest represents a request to place an order
type OrderRequest struct {
	CustomerName string `json:"customer_name"`
	CoffeeType   string `json:"coffee_type"`
}

// OrderResponse represents a response to an order request
type OrderResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
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
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parse request body into order request
	orderReq := new(OrderRequest)
	if err := json.NewDecoder(r.Body).Decode(orderReq); err != nil {
		log.Println(err)
		response := OrderResponse{
			Success: false,
			Message: "Invalid JSON format",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 1.1. Validate input
	if strings.TrimSpace(orderReq.CustomerName) == "" {
		response := OrderResponse{
			Success: false,
			Message: "Customer name is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if strings.TrimSpace(orderReq.CoffeeType) == "" {
		response := OrderResponse{
			Success: false,
			Message: "Coffee type is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
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
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Convert order to JSON
	orderJSON, err := json.Marshal(order)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Send the order to Kafka
	err = h.kafkaProducer.PushToQueue(h.config.Kafka.Topic, orderJSON)
	if err != nil {
		log.Println(err)
		// Remove the order from the store if we can't send it to Kafka
		h.orderStore.Delete(order.ID)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 6. Respond back to the user
	response := OrderResponse{
		Success: true,
		Message: "Order for " + order.CustomerName + " placed successfully!",
		Order:   order,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println(err)
		http.Error(w, "Error placing order", http.StatusInternalServerError)
		return
	}
}

// GetOrder handles retrieving an order
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Extract order ID from URL
	orderID := strings.TrimPrefix(r.URL.Path, "/order/")
	if orderID == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	// Get the order from the store
	order, err := h.orderStore.Get(orderID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Respond with the order
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		log.Println(err)
		http.Error(w, "Error retrieving order", http.StatusInternalServerError)
		return
	}
}

// CancelOrder handles cancelling an order
func (h *Handler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Extract order ID from URL
	orderID := strings.TrimPrefix(r.URL.Path, "/order/")
	orderID = strings.TrimSuffix(orderID, "/cancel")
	if orderID == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	// Get the order from the store
	order, err := h.orderStore.Get(orderID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Check if the order can be cancelled
	if order.Status != store.OrderStatusPending && order.Status != store.OrderStatusProcessing {
		http.Error(w, "Order cannot be cancelled", http.StatusBadRequest)
		return
	}

	// Update the order status
	order.Status = store.OrderStatusCancelled
	order.UpdatedAt = time.Now()
	if err := h.orderStore.Update(order); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert order to JSON
	orderJSON, err := json.Marshal(order)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send the cancelled order to Kafka
	err = h.kafkaProducer.PushToQueue(h.config.Kafka.Topic, orderJSON)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with success
	response := OrderResponse{
		Success: true,
		Message: "Order cancelled successfully!",
		Order:   order,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println(err)
		http.Error(w, "Error cancelling order", http.StatusInternalServerError)
		return
	}
}

// ListOrders handles listing all orders
func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	status := r.URL.Query().Get("status")
	customer := r.URL.Query().Get("customer")

	var orders []*store.Order
	var err error

	// Filter orders based on query parameters
	if status != "" {
		orders, err = h.orderStore.ListByStatus(store.OrderStatus(status))
	} else if customer != "" {
		orders, err = h.orderStore.ListByCustomer(customer)
	} else {
		orders, err = h.orderStore.List()
	}

	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with the orders
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		log.Println(err)
		http.Error(w, "Error listing orders", http.StatusInternalServerError)
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
