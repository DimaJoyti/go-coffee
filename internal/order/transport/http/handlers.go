package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/DimaJoyti/go-coffee/internal/order/application"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Handler represents the HTTP handler for order service
type Handler struct {
	orderService   *application.OrderService
	paymentService *application.PaymentService
	logger         *logger.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(orderService *application.OrderService, paymentService *application.PaymentService, logger *logger.Logger) *Handler {
	return &Handler{
		orderService:   orderService,
		paymentService: paymentService,
		logger:         logger,
	}
}

// SetupRoutes configures the HTTP routes for the order service
func (h *Handler) SetupRoutes(mux *http.ServeMux) {
	// Health check endpoint
	mux.HandleFunc("/health", h.methodHandler("GET", h.handleHealthCheck))

	// Order endpoints
	mux.HandleFunc("/api/v1/orders", h.methodHandler("POST", h.handleCreateOrder))
	mux.HandleFunc("/api/v1/orders/", h.handleOrderWithID)

	// Payment endpoints
	mux.HandleFunc("/api/v1/payments", h.methodHandler("POST", h.handleCreatePayment))
	mux.HandleFunc("/api/v1/payments/", h.handlePaymentWithID)
}

// methodHandler is a middleware that checks HTTP method
func (h *Handler) methodHandler(allowedMethod string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != allowedMethod {
			h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		handler(w, r)
	}
}

// handleHealthCheck handles health check requests
func (h *Handler) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "order-service",
		"timestamp": "2024-01-01T00:00:00Z", // This would be time.Now().UTC() in real implementation
	}
	h.writeSuccessResponse(w, response)
}

// handleCreateOrder handles order creation requests
func (h *Handler) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var req application.CreateOrderRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	resp, err := h.orderService.CreateOrder(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create order")
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeSuccessResponseWithStatus(w, http.StatusCreated, resp)
}

// handleOrderWithID handles order operations with ID parameter
func (h *Handler) handleOrderWithID(w http.ResponseWriter, r *http.Request) {
	// Extract order ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Order ID is required")
		return
	}

	orderID := parts[0]

	switch r.Method {
	case http.MethodGet:
		h.handleGetOrder(w, r, orderID)
	case http.MethodPut:
		if len(parts) >= 2 && parts[1] == "status" {
			h.handleUpdateOrderStatus(w, r, orderID)
		} else {
			h.writeErrorResponse(w, http.StatusNotFound, "Endpoint not found")
		}
	case http.MethodPost:
		if len(parts) >= 2 && parts[1] == "confirm" {
			h.handleConfirmOrder(w, r, orderID)
		} else {
			h.writeErrorResponse(w, http.StatusNotFound, "Endpoint not found")
		}
	case http.MethodDelete:
		h.handleCancelOrder(w, r, orderID)
	default:
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleGetOrder handles get order requests
func (h *Handler) handleGetOrder(w http.ResponseWriter, r *http.Request, orderID string) {
	customerID := r.URL.Query().Get("customer_id")

	req := &application.GetOrderRequest{
		OrderID:    orderID,
		CustomerID: customerID,
	}

	resp, err := h.orderService.GetOrder(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get order")
		h.writeErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	h.writeSuccessResponse(w, resp)
}

// handleConfirmOrder handles order confirmation requests
func (h *Handler) handleConfirmOrder(w http.ResponseWriter, r *http.Request, orderID string) {
	var req application.ConfirmOrderRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}
	req.OrderID = orderID

	resp, err := h.orderService.ConfirmOrder(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to confirm order")
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeSuccessResponse(w, resp)
}

// handleUpdateOrderStatus handles order status update requests
func (h *Handler) handleUpdateOrderStatus(w http.ResponseWriter, r *http.Request, orderID string) {
	var req application.UpdateOrderStatusRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}
	req.OrderID = orderID

	resp, err := h.orderService.UpdateOrderStatus(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update order status")
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeSuccessResponse(w, resp)
}

// handleCancelOrder handles order cancellation requests
func (h *Handler) handleCancelOrder(w http.ResponseWriter, r *http.Request, orderID string) {
	customerID := r.URL.Query().Get("customer_id")
	reason := r.URL.Query().Get("reason")

	req := &application.CancelOrderRequest{
		OrderID:    orderID,
		CustomerID: customerID,
		Reason:     reason,
	}

	resp, err := h.orderService.CancelOrder(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to cancel order")
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeSuccessResponse(w, resp)
}

// handleCreatePayment handles payment creation requests
func (h *Handler) handleCreatePayment(w http.ResponseWriter, r *http.Request) {
	var req application.CreatePaymentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	resp, err := h.paymentService.CreatePayment(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create payment")
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeSuccessResponseWithStatus(w, http.StatusCreated, resp)
}

// handlePaymentWithID handles payment operations with ID parameter
func (h *Handler) handlePaymentWithID(w http.ResponseWriter, r *http.Request) {
	// Extract payment ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/payments/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Payment ID is required")
		return
	}

	paymentID := parts[0]

	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if len(parts) < 2 {
		h.writeErrorResponse(w, http.StatusNotFound, "Action is required")
		return
	}

	action := parts[1]
	switch action {
	case "process":
		h.handleProcessPayment(w, r, paymentID)
	case "refund":
		h.handleRefundPayment(w, r, paymentID)
	default:
		h.writeErrorResponse(w, http.StatusNotFound, "Action not found")
	}
}

// handleProcessPayment handles payment processing requests
func (h *Handler) handleProcessPayment(w http.ResponseWriter, r *http.Request, paymentID string) {
	var req application.ProcessPaymentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}
	req.PaymentID = paymentID

	resp, err := h.paymentService.ProcessPayment(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to process payment")
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeSuccessResponse(w, resp)
}

// handleRefundPayment handles payment refund requests
func (h *Handler) handleRefundPayment(w http.ResponseWriter, r *http.Request, paymentID string) {
	var req application.RefundPaymentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}
	req.PaymentID = paymentID

	resp, err := h.paymentService.RefundPayment(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to refund payment")
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeSuccessResponse(w, resp)
}

// Helper methods for response handling
func (h *Handler) writeSuccessResponse(w http.ResponseWriter, data interface{}) {
	h.writeSuccessResponseWithStatus(w, http.StatusOK, data)
}

func (h *Handler) writeSuccessResponseWithStatus(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
