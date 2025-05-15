package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/yourusername/coffee-order-system/pkg/errors"
	"github.com/yourusername/coffee-order-system/pkg/logger"
	"github.com/yourusername/coffee-order-system/pkg/models"

	"kafka_producer/internal/service"
)

// OrderRequest представляє запит на створення замовлення
type OrderRequest struct {
	CustomerName string `json:"customer_name"`
	CoffeeType   string `json:"coffee_type"`
}

// OrderResponse представляє відповідь на запит
type OrderResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message,omitempty"`
	Order   *models.Order `json:"order,omitempty"`
}

// OrdersResponse представляє відповідь зі списком замовлень
type OrdersResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Orders  []*models.Order `json:"orders,omitempty"`
}

// Handler представляє HTTP обробник
type Handler struct {
	orderService *service.OrderService
	logger       logger.Logger
}

// NewHandler створює новий HTTP обробник
func NewHandler(orderService *service.OrderService, logger logger.Logger) *Handler {
	return &Handler{
		orderService: orderService,
		logger:       logger,
	}
}

// PlaceOrder обробляє запит на створення замовлення
func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Розбір запиту
	var req OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request: %v", err)
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Валідація запиту
	if req.CustomerName == "" {
		h.respondWithError(w, http.StatusBadRequest, "Customer name is required")
		return
	}
	if req.CoffeeType == "" {
		h.respondWithError(w, http.StatusBadRequest, "Coffee type is required")
		return
	}

	// Створення замовлення
	order, err := h.orderService.CreateOrder(req.CustomerName, req.CoffeeType)
	if err != nil {
		h.logger.Error("Failed to create order: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create order")
		return
	}

	// Відповідь
	h.respondWithJSON(w, http.StatusCreated, OrderResponse{
		Success: true,
		Message: "Order placed successfully",
		Order:   order,
	})
}

// GetOrder обробляє запит на отримання замовлення
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Отримання ID замовлення з URL
	id := strings.TrimPrefix(r.URL.Path, "/order/")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Order ID is required")
		return
	}

	// Отримання замовлення
	order, err := h.orderService.GetOrder(id)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok && appErr.Code == "ORDER_NOT_FOUND" {
			h.respondWithError(w, http.StatusNotFound, "Order not found")
			return
		}
		h.logger.Error("Failed to get order: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get order")
		return
	}

	// Відповідь
	h.respondWithJSON(w, http.StatusOK, OrderResponse{
		Success: true,
		Order:   order,
	})
}

// CancelOrder обробляє запит на скасування замовлення
func (h *Handler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Отримання ID замовлення з URL
	path := strings.TrimPrefix(r.URL.Path, "/order/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "cancel" {
		h.respondWithError(w, http.StatusBadRequest, "Invalid URL")
		return
	}
	id := parts[0]
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Order ID is required")
		return
	}

	// Скасування замовлення
	order, err := h.orderService.CancelOrder(id)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok && appErr.Code == "ORDER_NOT_FOUND" {
			h.respondWithError(w, http.StatusNotFound, "Order not found")
			return
		}
		h.logger.Error("Failed to cancel order: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to cancel order")
		return
	}

	// Відповідь
	h.respondWithJSON(w, http.StatusOK, OrderResponse{
		Success: true,
		Message: "Order cancelled successfully",
		Order:   order,
	})
}

// ListOrders обробляє запит на отримання списку замовлень
func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Отримання списку замовлень
	orders, err := h.orderService.ListOrders()
	if err != nil {
		h.logger.Error("Failed to list orders: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list orders")
		return
	}

	// Відповідь
	h.respondWithJSON(w, http.StatusOK, OrdersResponse{
		Success: true,
		Orders:  orders,
	})
}

// HealthCheck обробляє запит на перевірку стану сервісу
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondWithJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// respondWithJSON відправляє JSON відповідь
func (h *Handler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		h.logger.Error("Failed to marshal response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// respondWithError відправляє помилку
func (h *Handler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, OrderResponse{
		Success: false,
		Message: message,
	})
}
