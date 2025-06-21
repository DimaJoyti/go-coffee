package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Handlers contains all HTTP handlers for the User Gateway
type Handlers struct {
	aiOrderConn *grpc.ClientConn
	kitchenConn *grpc.ClientConn
	commConn    *grpc.ClientConn
	container   infrastructure.ContainerInterface
	logger      *logger.Logger
}

// NewHandlers creates a new handlers instance
func NewHandlers(
	aiOrderConn *grpc.ClientConn,
	kitchenConn *grpc.ClientConn,
	commConn *grpc.ClientConn,
	container infrastructure.ContainerInterface,
	logger *logger.Logger,
) *Handlers {
	return &Handlers{
		aiOrderConn: aiOrderConn,
		kitchenConn: kitchenConn,
		commConn:    commConn,
		container:   container,
		logger:      logger,
	}
}

// HealthCheck returns the health status of the service (Clean HTTP Handler)
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Health check requested")

	// Get infrastructure health status
	infraHealth, err := h.container.HealthCheck(r.Context())
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Infrastructure health check failed", err)
		return
	}

	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "user-gateway",
		"version":   "1.0.0",
		"timestamp": time.Now().Unix(),
		"connections": map[string]interface{}{
			"ai_order": h.getConnectionStatus(h.aiOrderConn),
			"kitchen":  h.getConnectionStatus(h.kitchenConn),
			"comm":     h.getConnectionStatus(h.commConn),
		},
		"infrastructure": infraHealth,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// CreateOrder creates a new order (Clean HTTP Handler)
func (h *Handlers) CreateOrder(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Creating new order")

	var req map[string]interface{}
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Mock response
	response := map[string]interface{}{
		"order_id":       "order_" + generateOrderID(),
		"status":         "PENDING",
		"total":          15.99,
		"currency":       "USD",
		"created_at":     time.Now().Format(time.RFC3339),
		"estimated_time": 300,
		"items":          req["items"],
	}

	h.respondWithJSON(w, http.StatusCreated, response)
}

// GetOrder retrieves an order by ID (Clean HTTP Handler)
func (h *Handlers) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := h.getPathParam(r, "id")
	h.logger.WithFields(map[string]interface{}{"order_id": orderID}).Info("Getting order")

	// Mock response
	response := map[string]interface{}{
		"order_id":    orderID,
		"status":      "PREPARING",
		"total":       15.99,
		"currency":    "USD",
		"created_at":  time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
		"updated_at":  time.Now().Format(time.RFC3339),
		"customer_id": "customer_123",
		"items": []map[string]interface{}{
			{
				"id":       "item_1",
				"name":     "Espresso",
				"quantity": 1,
				"price":    3.50,
			},
			{
				"id":       "item_2",
				"name":     "Croissant",
				"quantity": 1,
				"price":    2.50,
			},
		},
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// ListOrders lists orders with optional filtering (Clean HTTP Handler)
func (h *Handlers) ListOrders(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Listing orders")

	customerID := h.getQueryParam(r, "customer_id")
	status := h.getQueryParam(r, "status")
	page := h.getQueryParamInt(r, "page", 1)
	pageSize := h.getQueryParamInt(r, "page_size", 10)

	h.logger.Info("Listing orders", map[string]interface{}{
		"customer_id": customerID,
		"status":      status,
		"page":        page,
		"page_size":   pageSize,
	})

	// Mock response
	orders := []map[string]interface{}{
		{
			"order_id":   "order_001",
			"status":     "COMPLETED",
			"total":      12.50,
			"created_at": time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
		},
		{
			"order_id":   "order_002",
			"status":     "PREPARING",
			"total":      18.75,
			"created_at": time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
		},
	}

	response := map[string]interface{}{
		"orders":    orders,
		"total":     len(orders),
		"page_size": pageSize,
		"page":      page,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// UpdateOrderStatus updates the status of an order (Clean HTTP Handler)
func (h *Handlers) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	orderID := h.getPathParam(r, "id")
	h.logger.Info("Updating order status", map[string]interface{}{"order_id": orderID})

	var reqBody struct {
		NewStatus      string `json:"new_status"`
		Reason         string `json:"reason"`
		NotifyCustomer bool   `json:"notify_customer"`
	}

	if err := h.decodeJSON(r, &reqBody); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	if reqBody.NewStatus == "" {
		h.respondWithError(w, http.StatusBadRequest, "new_status is required", nil)
		return
	}

	// Mock response
	response := map[string]interface{}{
		"order_id":   orderID,
		"status":     reqBody.NewStatus,
		"updated_at": time.Now().Format(time.RFC3339),
		"reason":     reqBody.Reason,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// CancelOrder cancels an order (Clean HTTP Handler)
func (h *Handlers) CancelOrder(w http.ResponseWriter, r *http.Request) {
	orderID := h.getPathParam(r, "id")
	h.logger.Info("Cancelling order", map[string]interface{}{"order_id": orderID})

	var reqBody struct {
		Reason         string `json:"reason"`
		RefundRequired bool   `json:"refund_required"`
	}

	if err := h.decodeJSON(r, &reqBody); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Mock response
	response := map[string]interface{}{
		"order_id":        orderID,
		"status":          "CANCELLED",
		"cancelled_at":    time.Now().Format(time.RFC3339),
		"reason":          reqBody.Reason,
		"refund_required": reqBody.RefundRequired,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// GetOrderRecommendations gets AI recommendations for orders (Clean HTTP Handler)
func (h *Handlers) GetOrderRecommendations(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting order recommendations")

	customerID := h.getQueryParam(r, "customer_id")
	locationID := h.getQueryParam(r, "location_id")

	h.logger.Info("Getting recommendations", map[string]interface{}{
		"customer_id": customerID,
		"location_id": locationID,
	})

	// Mock recommendations
	recommendations := []map[string]interface{}{
		{
			"item_id":     "coffee_001",
			"name":        "Cappuccino",
			"description": "Based on your previous orders",
			"price":       4.50,
			"confidence":  0.85,
		},
		{
			"item_id":     "pastry_001",
			"name":        "Blueberry Muffin",
			"description": "Popular with cappuccino",
			"price":       3.25,
			"confidence":  0.72,
		},
	}

	response := map[string]interface{}{
		"recommendations": recommendations,
		"generated_at":    time.Now().Format(time.RFC3339),
		"customer_id":     customerID,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// GetKitchenQueue gets the current kitchen queue (Clean HTTP Handler)
func (h *Handlers) GetKitchenQueue(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting kitchen queue")

	locationID := h.getQueryParam(r, "location_id")

	// Mock kitchen queue
	queue := []map[string]interface{}{
		{
			"order_id":       "order_001",
			"position":       1,
			"estimated_time": 180,
			"status":         "PREPARING",
			"items_count":    2,
		},
		{
			"order_id":       "order_002",
			"position":       2,
			"estimated_time": 240,
			"status":         "PENDING",
			"items_count":    3,
		},
	}

	response := map[string]interface{}{
		"queue":       queue,
		"total_items": len(queue),
		"location_id": locationID,
		"updated_at":  time.Now().Format(time.RFC3339),
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// Clean HTTP handlers for remaining endpoints
func (h *Handlers) AnalyzeOrderPatterns(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Order patterns analysis",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) PredictCompletionTime(w http.ResponseWriter, r *http.Request) {
	orderID := h.getPathParam(r, "id")
	response := map[string]interface{}{
		"order_id":       orderID,
		"estimated_time": 300,
		"confidence":     0.85,
		"predicted_at":   time.Now().Format(time.RFC3339),
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) AddToKitchenQueue(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Added to kitchen queue",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusCreated, response)
}

func (h *Handlers) UpdatePreparationStatus(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Preparation status updated",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) CompleteOrder(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Order completed",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) GetKitchenMetrics(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Kitchen metrics",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) OptimizeKitchenWorkflow(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Kitchen workflow optimized",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) PredictKitchenCapacity(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Kitchen capacity predicted",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) GetIngredientRequirements(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Ingredient requirements",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) SendMessage(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Message sent",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) BroadcastMessage(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Message broadcasted",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) GetMessageHistory(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Message history",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) GetActiveServices(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"services": []map[string]interface{}{
			{"name": "ai-search", "status": "healthy", "port": 8092},
			{"name": "auth-service", "status": "healthy", "port": 8080},
			{"name": "kitchen-service", "status": "healthy", "port": 50052},
			{"name": "communication-hub", "status": "healthy", "port": 50053},
		},
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) SendNotification(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Notification sent",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) GetCommunicationAnalytics(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Communication analytics",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) GetCustomerProfile(w http.ResponseWriter, r *http.Request) {
	customerID := h.getPathParam(r, "id")
	response := map[string]interface{}{
		"customer_id": customerID,
		"name":        "John Doe",
		"email":       "john@example.com",
		"status":      "active",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) UpdateCustomerProfile(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Customer profile updated",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) GetCustomerOrders(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Customer orders",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) GetCustomerRecommendations(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Customer recommendations",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) GetOrderAnalytics(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Order analytics",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) GetKitchenAnalytics(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Kitchen analytics",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) GetPerformanceAnalytics(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Performance analytics",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) GetAIInsights(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "AI insights",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "WebSocket endpoint",
		"status":  "mock",
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) GetAPIDocumentation(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"title":       "Go Coffee User Gateway API",
		"version":     "1.0.0",
		"description": "User gateway for Go Coffee microservices",
		"endpoints": map[string]interface{}{
			"orders":          "/api/v1/orders",
			"recommendations": "/api/v1/recommendations",
			"kitchen":         "/api/v1/kitchen",
			"communication":   "/api/v1/communication",
			"customers":       "/api/v1/customers",
			"analytics":       "/api/v1/analytics",
		},
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

// Helper functions

func (h *Handlers) getConnectionStatus(conn *grpc.ClientConn) string {
	if conn == nil {
		return "disconnected"
	}
	return conn.GetState().String()
}

func generateOrderID() string {
	return time.Now().Format("20060102150405")
}

// HTTP Helper Functions for Clean Architecture

// decodeJSON decodes JSON request body into the provided struct
func (h *Handlers) decodeJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}

// respondWithJSON sends a JSON response
func (h *Handlers) respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// respondWithError sends an error response
func (h *Handlers) respondWithError(w http.ResponseWriter, statusCode int, message string, err error) {
	h.logger.WithError(err).WithField("status_code", statusCode).Error("%s", message)

	response := map[string]interface{}{
		"error":     message,
		"timestamp": time.Now().Unix(),
	}

	if err != nil {
		response["details"] = err.Error()
	}

	h.respondWithJSON(w, statusCode, response)
}

// getPathParam extracts a path parameter from the URL using gorilla/mux
func (h *Handlers) getPathParam(r *http.Request, key string) string {
	vars := mux.Vars(r)
	return vars[key]
}

// getQueryParam extracts a query parameter from the URL
func (h *Handlers) getQueryParam(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

// getQueryParamInt extracts an integer query parameter from the URL
func (h *Handlers) getQueryParamInt(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// Session Management Examples (Clean HTTP Handlers)

// GetUserProfile retrieves user profile with session context
func (h *Handlers) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	// Extract user information from session context (set by session middleware)
	userID := r.Context().Value("user_id")
	userEmail := r.Context().Value("user_email")
	userRole := r.Context().Value("user_role")
	sessionID := r.Context().Value("session_id")

	h.logger.Info("Getting user profile", map[string]interface{}{
		"user_id":    userID,
		"session_id": sessionID,
	})

	// Mock user profile response
	response := map[string]interface{}{
		"user_id":    userID,
		"email":      userEmail,
		"role":       userRole,
		"session_id": sessionID,
		"profile": map[string]interface{}{
			"name":  "John Doe",
			"phone": "+1-555-0123",
			"preferences": map[string]interface{}{
				"favorite_drink": "Espresso",
				"milk_type":      "Oat",
				"sugar_level":    "Medium",
			},
			"loyalty_points": 150,
			"member_since":   "2023-01-15",
		},
		"last_login": time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// LoginExample demonstrates session creation
func (h *Handlers) LoginExample(w http.ResponseWriter, r *http.Request) {
	var loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := h.decodeJSON(r, &loginRequest); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Mock authentication (in real implementation, validate credentials)
	if loginRequest.Email == "" || loginRequest.Password == "" {
		h.respondWithError(w, http.StatusBadRequest, "Email and password required", nil)
		return
	}

	// Mock successful authentication
	userID := "user_123"
	email := loginRequest.Email
	role := "customer"

	// In a real implementation, you would:
	// 1. Get session manager from container: sessionManager := h.container.GetSessionManager()
	// 2. Create session: session, err := sessionManager.CreateSession(ctx, userID, email, role, r)
	// 3. Set session cookie: sessionManager.SetSessionCookie(w, session.ID)

	response := map[string]interface{}{
		"message":    "Login successful",
		"user_id":    userID,
		"email":      email,
		"role":       role,
		"session_id": "session_" + userID + "_" + fmt.Sprintf("%d", time.Now().Unix()),
		"expires_at": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// LogoutExample demonstrates session cleanup
func (h *Handlers) LogoutExample(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Context().Value("session_id")
	userID := r.Context().Value("user_id")

	h.logger.Info("User logout", map[string]interface{}{
		"user_id":    userID,
		"session_id": sessionID,
	})

	// In a real implementation, you would:
	// 1. Get session manager from container: sessionManager := h.container.GetSessionManager()
	// 2. Revoke session: err := sessionManager.RevokeSession(ctx, sessionID.(string))
	// 3. Clear session cookie: sessionManager.ClearSessionCookie(w)

	response := map[string]interface{}{
		"message":    "Logout successful",
		"session_id": sessionID,
		"logged_out": true,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}
