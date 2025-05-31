package user

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Handlers contains all HTTP handlers for the User Gateway
type Handlers struct {
	aiOrderConn *grpc.ClientConn
	kitchenConn *grpc.ClientConn
	commConn    *grpc.ClientConn
	logger      *logger.Logger
}

// NewHandlers creates a new handlers instance
func NewHandlers(
	aiOrderConn *grpc.ClientConn,
	kitchenConn *grpc.ClientConn,
	commConn *grpc.ClientConn,
	logger *logger.Logger,
) *Handlers {
	return &Handlers{
		aiOrderConn: aiOrderConn,
		kitchenConn: kitchenConn,
		commConn:    commConn,
		logger:      logger,
	}
}

// HealthCheck returns the health status of the service
func (h *Handlers) HealthCheck(c *gin.Context) {
	h.logger.Info("Health check requested")
	
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "user-gateway",
		"version":   "1.0.0",
		"timestamp": time.Now().Unix(),
		"connections": gin.H{
			"ai_order": h.getConnectionStatus(h.aiOrderConn),
			"kitchen":  h.getConnectionStatus(h.kitchenConn),
			"comm":     h.getConnectionStatus(h.commConn),
		},
	})
}

// CreateOrder creates a new order
func (h *Handlers) CreateOrder(c *gin.Context) {
	h.logger.Info("Creating new order")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Mock response
	resp := gin.H{
		"order_id":       "order_" + generateOrderID(),
		"status":         "PENDING",
		"total":          15.99,
		"currency":       "USD",
		"created_at":     time.Now().Format(time.RFC3339),
		"estimated_time": 300,
		"items":          req["items"],
	}

	c.JSON(http.StatusCreated, resp)
}

// GetOrder retrieves an order by ID
func (h *Handlers) GetOrder(c *gin.Context) {
	orderID := c.Param("id")
	h.logger.Info("Getting order", map[string]interface{}{"order_id": orderID})

	// Mock response
	resp := gin.H{
		"order_id":     orderID,
		"status":       "PREPARING",
		"total":        15.99,
		"currency":     "USD",
		"created_at":   time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
		"updated_at":   time.Now().Format(time.RFC3339),
		"customer_id":  "customer_123",
		"items": []gin.H{
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

	c.JSON(http.StatusOK, resp)
}

// ListOrders lists orders with optional filtering
func (h *Handlers) ListOrders(c *gin.Context) {
	h.logger.Info("Listing orders")

	customerID := c.Query("customer_id")
	status := c.Query("status")

	h.logger.Info("Listing orders", map[string]interface{}{
		"customer_id": customerID,
		"status":      status,
	})

	// Mock response
	orders := []gin.H{
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

	resp := gin.H{
		"orders":     orders,
		"total":      len(orders),
		"page_size":  10,
		"page":       1,
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateOrderStatus updates the status of an order
func (h *Handlers) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	h.logger.Info("Updating order status", map[string]interface{}{"order_id": orderID})

	var reqBody struct {
		NewStatus      string `json:"new_status" binding:"required"`
		Reason         string `json:"reason"`
		NotifyCustomer bool   `json:"notify_customer"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		h.logger.Error("Failed to bind JSON", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Mock response
	resp := gin.H{
		"order_id":   orderID,
		"status":     reqBody.NewStatus,
		"updated_at": time.Now().Format(time.RFC3339),
		"reason":     reqBody.Reason,
	}

	c.JSON(http.StatusOK, resp)
}

// CancelOrder cancels an order
func (h *Handlers) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")
	h.logger.Info("Cancelling order", map[string]interface{}{"order_id": orderID})

	var reqBody struct {
		Reason         string `json:"reason"`
		RefundRequired bool   `json:"refund_required"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		h.logger.Error("Failed to bind JSON", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Mock response
	resp := gin.H{
		"order_id":        orderID,
		"status":          "CANCELLED",
		"cancelled_at":    time.Now().Format(time.RFC3339),
		"reason":          reqBody.Reason,
		"refund_required": reqBody.RefundRequired,
	}

	c.JSON(http.StatusOK, resp)
}

// GetOrderRecommendations gets AI recommendations for orders
func (h *Handlers) GetOrderRecommendations(c *gin.Context) {
	h.logger.Info("Getting order recommendations")

	customerID := c.Query("customer_id")
	locationID := c.Query("location_id")

	h.logger.Info("Getting recommendations", map[string]interface{}{
		"customer_id": customerID,
		"location_id": locationID,
	})

	// Mock recommendations
	recommendations := []gin.H{
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

	resp := gin.H{
		"recommendations": recommendations,
		"generated_at":    time.Now().Format(time.RFC3339),
		"customer_id":     customerID,
	}

	c.JSON(http.StatusOK, resp)
}

// GetKitchenQueue gets the current kitchen queue
func (h *Handlers) GetKitchenQueue(c *gin.Context) {
	h.logger.Info("Getting kitchen queue")

	locationID := c.Query("location_id")

	// Mock kitchen queue
	queue := []gin.H{
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

	resp := gin.H{
		"queue":       queue,
		"total_items": len(queue),
		"location_id": locationID,
		"updated_at":  time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, resp)
}

// Placeholder handlers for remaining endpoints
func (h *Handlers) AnalyzeOrderPatterns(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Order patterns analysis", "status": "mock"})
}

func (h *Handlers) PredictCompletionTime(c *gin.Context) {
	orderID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"order_id":        orderID,
		"estimated_time":  300,
		"confidence":      0.85,
		"predicted_at":    time.Now().Format(time.RFC3339),
	})
}

func (h *Handlers) AddToKitchenQueue(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "Added to kitchen queue", "status": "mock"})
}

func (h *Handlers) UpdatePreparationStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Preparation status updated", "status": "mock"})
}

func (h *Handlers) CompleteOrder(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Order completed", "status": "mock"})
}

func (h *Handlers) GetKitchenMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Kitchen metrics", "status": "mock"})
}

func (h *Handlers) OptimizeKitchenWorkflow(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Kitchen workflow optimized", "status": "mock"})
}

func (h *Handlers) PredictKitchenCapacity(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Kitchen capacity predicted", "status": "mock"})
}

func (h *Handlers) GetIngredientRequirements(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Ingredient requirements", "status": "mock"})
}

func (h *Handlers) SendMessage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Message sent", "status": "mock"})
}

func (h *Handlers) BroadcastMessage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Message broadcasted", "status": "mock"})
}

func (h *Handlers) GetMessageHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Message history", "status": "mock"})
}

func (h *Handlers) GetActiveServices(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"services": []gin.H{
			{"name": "ai-search", "status": "healthy", "port": 8092},
			{"name": "auth-service", "status": "healthy", "port": 8080},
			{"name": "kitchen-service", "status": "healthy", "port": 50052},
			{"name": "communication-hub", "status": "healthy", "port": 50053},
		},
	})
}

func (h *Handlers) SendNotification(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Notification sent", "status": "mock"})
}

func (h *Handlers) GetCommunicationAnalytics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Communication analytics", "status": "mock"})
}

func (h *Handlers) GetCustomerProfile(c *gin.Context) {
	customerID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"customer_id": customerID,
		"name":        "John Doe",
		"email":       "john@example.com",
		"status":      "active",
	})
}

func (h *Handlers) UpdateCustomerProfile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Customer profile updated", "status": "mock"})
}

func (h *Handlers) GetCustomerOrders(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Customer orders", "status": "mock"})
}

func (h *Handlers) GetCustomerRecommendations(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Customer recommendations", "status": "mock"})
}

func (h *Handlers) GetOrderAnalytics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Order analytics", "status": "mock"})
}

func (h *Handlers) GetKitchenAnalytics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Kitchen analytics", "status": "mock"})
}

func (h *Handlers) GetPerformanceAnalytics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Performance analytics", "status": "mock"})
}

func (h *Handlers) GetAIInsights(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "AI insights", "status": "mock"})
}

func (h *Handlers) HandleWebSocket(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "WebSocket endpoint", "status": "mock"})
}

func (h *Handlers) GetAPIDocumentation(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"title":       "Go Coffee User Gateway API",
		"version":     "1.0.0",
		"description": "User gateway for Go Coffee microservices",
		"endpoints": gin.H{
			"orders":         "/api/v1/orders",
			"recommendations": "/api/v1/recommendations",
			"kitchen":        "/api/v1/kitchen",
			"communication":  "/api/v1/communication",
			"customers":      "/api/v1/customers",
			"analytics":      "/api/v1/analytics",
		},
	})
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
