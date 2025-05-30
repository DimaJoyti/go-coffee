package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb_ai_order "github.com/DimaJoyti/go-coffee/api/proto/ai_order"
	pb_kitchen "github.com/DimaJoyti/go-coffee/api/proto/kitchen"
	pb_communication "github.com/DimaJoyti/go-coffee/api/proto/communication"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Handlers contains all HTTP handlers for the User Gateway
type Handlers struct {
	aiOrderClient pb_ai_order.AIOrderServiceClient
	kitchenClient pb_kitchen.KitchenServiceClient
	commClient    pb_communication.CommunicationServiceClient
	logger        *logger.Logger
}

// NewHandlers creates a new handlers instance
func NewHandlers(
	aiOrderConn *grpc.ClientConn,
	kitchenConn *grpc.ClientConn,
	commConn *grpc.ClientConn,
	logger *logger.Logger,
) *Handlers {
	return &Handlers{
		aiOrderClient: pb_ai_order.NewAIOrderServiceClient(aiOrderConn),
		kitchenClient: pb_kitchen.NewKitchenServiceClient(kitchenConn),
		commClient:    pb_communication.NewCommunicationServiceClient(commConn),
		logger:        logger,
	}
}

// HealthCheck returns the health status of the service
func (h *Handlers) HealthCheck(c *gin.Context) {
	h.logger.Info("Health check requested")
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "user-gateway",
		"version": "1.0.0",
		"timestamp": gin.H{
			"unix": gin.H{
				"seconds": 1234567890,
			},
		},
	})
}

// CreateOrder creates a new order
func (h *Handlers) CreateOrder(c *gin.Context) {
	h.logger.Info("Creating new order")

	var req pb_ai_order.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	resp, err := h.aiOrderClient.CreateOrder(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create order", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetOrder retrieves an order by ID
func (h *Handlers) GetOrder(c *gin.Context) {
	orderID := c.Param("id")
	h.logger.Info("Getting order", zap.String("order_id", orderID))

	includeAI := c.Query("include_ai") == "true"

	req := &pb_ai_order.GetOrderRequest{
		OrderId:          orderID,
		IncludeAiInsights: includeAI,
	}

	resp, err := h.aiOrderClient.GetOrder(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get order", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListOrders lists orders with optional filtering
func (h *Handlers) ListOrders(c *gin.Context) {
	h.logger.Info("Listing orders")

	req := &pb_ai_order.ListOrdersRequest{
		CustomerId:        c.Query("customer_id"),
		LocationId:        c.Query("location_id"),
		EnableAiFiltering: c.Query("enable_ai") == "true",
	}

	// Parse page size
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			req.PageSize = int32(pageSize)
		}
	}

	// Parse status
	if statusStr := c.Query("status"); statusStr != "" {
		if status, ok := pb_ai_order.OrderStatus_value[statusStr]; ok {
			req.Status = pb_ai_order.OrderStatus(status)
		}
	}

	resp, err := h.aiOrderClient.ListOrders(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list orders", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list orders"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateOrderStatus updates the status of an order
func (h *Handlers) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	h.logger.Info("Updating order status", zap.String("order_id", orderID))

	var reqBody struct {
		NewStatus      string `json:"new_status" binding:"required"`
		Reason         string `json:"reason"`
		NotifyCustomer bool   `json:"notify_customer"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Parse status
	status, ok := pb_ai_order.OrderStatus_value[reqBody.NewStatus]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	req := &pb_ai_order.UpdateOrderStatusRequest{
		OrderId:        orderID,
		NewStatus:      pb_ai_order.OrderStatus(status),
		Reason:         reqBody.Reason,
		NotifyCustomer: reqBody.NotifyCustomer,
	}

	resp, err := h.aiOrderClient.UpdateOrderStatus(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to update order status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CancelOrder cancels an order
func (h *Handlers) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")
	h.logger.Info("Cancelling order", zap.String("order_id", orderID))

	var reqBody struct {
		Reason         string `json:"reason"`
		RefundRequired bool   `json:"refund_required"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	req := &pb_ai_order.CancelOrderRequest{
		OrderId:        orderID,
		Reason:         reqBody.Reason,
		RefundRequired: reqBody.RefundRequired,
	}

	resp, err := h.aiOrderClient.CancelOrder(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to cancel order", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetOrderRecommendations gets AI recommendations for orders
func (h *Handlers) GetOrderRecommendations(c *gin.Context) {
	h.logger.Info("Getting order recommendations")

	req := &pb_ai_order.GetOrderRecommendationsRequest{
		CustomerId: c.Query("customer_id"),
		LocationId: c.Query("location_id"),
		TimeOfDay:  c.Query("time_of_day"),
	}

	// Parse current items
	if items := c.QueryArray("current_items"); len(items) > 0 {
		req.CurrentItems = items
	}

	resp, err := h.aiOrderClient.GetOrderRecommendations(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get recommendations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recommendations"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AnalyzeOrderPatterns analyzes order patterns
func (h *Handlers) AnalyzeOrderPatterns(c *gin.Context) {
	h.logger.Info("Analyzing order patterns")

	req := &pb_ai_order.AnalyzeOrderPatternsRequest{
		LocationId:   c.Query("location_id"),
		AnalysisType: c.Query("analysis_type"),
	}

	resp, err := h.aiOrderClient.AnalyzeOrderPatterns(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to analyze patterns", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze patterns"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// PredictCompletionTime predicts order completion time
func (h *Handlers) PredictCompletionTime(c *gin.Context) {
	orderID := c.Param("id")
	h.logger.Info("Predicting completion time", zap.String("order_id", orderID))

	req := &pb_ai_order.PredictCompletionTimeRequest{
		OrderId:    orderID,
		LocationId: c.Query("location_id"),
	}

	// Parse queue size
	if queueSizeStr := c.Query("queue_size"); queueSizeStr != "" {
		if queueSize, err := strconv.Atoi(queueSizeStr); err == nil {
			req.CurrentQueueSize = int32(queueSize)
		}
	}

	resp, err := h.aiOrderClient.PredictCompletionTime(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to predict completion time", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to predict completion time"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Kitchen-related handlers

// GetKitchenQueue gets the current kitchen queue
func (h *Handlers) GetKitchenQueue(c *gin.Context) {
	h.logger.Info("Getting kitchen queue")

	req := &pb_kitchen.GetQueueRequest{
		LocationId:        c.Query("location_id"),
		IncludeAiInsights: c.Query("include_ai") == "true",
	}

	resp, err := h.kitchenClient.GetQueue(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get kitchen queue", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get kitchen queue"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AddToKitchenQueue adds an order to the kitchen queue
func (h *Handlers) AddToKitchenQueue(c *gin.Context) {
	h.logger.Info("Adding order to kitchen queue")

	var req pb_kitchen.AddToQueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	resp, err := h.kitchenClient.AddToQueue(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to add to kitchen queue", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to kitchen queue"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Placeholder handlers for remaining endpoints
func (h *Handlers) UpdatePreparationStatus(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handlers) CompleteOrder(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handlers) GetKitchenMetrics(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handlers) OptimizeKitchenWorkflow(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handlers) PredictKitchenCapacity(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handlers) GetIngredientRequirements(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handlers) SendMessage(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handlers) BroadcastMessage(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handlers) GetMessageHistory(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handlers) GetActiveServices(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handlers) SendNotification(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handlers) GetCommunicationAnalytics(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handlers) GetCustomerProfile(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handlers) UpdateCustomerProfile(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handlers) GetCustomerOrders(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handlers) GetCustomerRecommendations(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handlers) GetOrderAnalytics(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handlers) GetKitchenAnalytics(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handlers) GetPerformanceAnalytics(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handlers) GetAIInsights(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handlers) HandleWebSocket(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "WebSocket not implemented yet"})
}

func (h *Handlers) GetAPIDocumentation(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"title":       "AI Order Management API",
		"version":     "1.0.0",
		"description": "Intelligent order management system with AI analytics",
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
