package aiorder

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/DimaJoyti/go-coffee/api/proto/ai_order"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/google/uuid"
)

// Service implements the AI Order Service
type Service struct {
	pb.UnimplementedAIOrderServiceServer
	repository  Repository
	aiProcessor AIProcessor
	logger      *logger.Logger
}

// NewService creates a new AI Order Service instance
func NewService(repo Repository, aiProcessor AIProcessor, logger *logger.Logger) *Service {
	return &Service{
		repository:  repo,
		aiProcessor: aiProcessor,
		logger:      logger,
	}
}

// CreateOrder creates a new order with AI analysis
func (s *Service) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	s.logger.Info("Creating new order with AI analysis",
		zap.String("customer_id", req.Customer.Id),
		zap.String("location_id", req.LocationId),
		zap.Int("items_count", len(req.Items)),
	)

	// Validate request
	if err := s.validateCreateOrderRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	// Generate order ID
	orderID := uuid.New().String()

	// Create order object
	order := &pb.Order{
		Id:                  orderID,
		Customer:            req.Customer,
		Items:               req.Items,
		Status:              pb.OrderStatus_ORDER_STATUS_PENDING,
		Priority:            pb.OrderPriority_ORDER_PRIORITY_NORMAL,
		TotalAmount:         s.calculateTotalAmount(req.Items),
		CreatedAt:           timestamppb.Now(),
		UpdatedAt:           timestamppb.Now(),
		LocationId:          req.LocationId,
		SpecialInstructions: req.SpecialInstructions,
	}

	// Apply AI analysis if enabled
	if req.EnableAiOptimization {
		aiInsights, err := s.aiProcessor.AnalyzeOrder(ctx, order)
		if err != nil {
			s.logger.Warn("Failed to get AI insights for order", zap.Error(err))
		} else {
			order.AiInsights = aiInsights
			order.Priority = s.determinePriorityFromAI(aiInsights)
			order.EstimatedCompletion = s.calculateEstimatedCompletion(aiInsights)
		}
	}

	// Save order to repository
	if err := s.repository.CreateOrder(ctx, order); err != nil {
		s.logger.Error("Failed to create order", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create order: %v", err)
	}

	// Get AI recommendations
	recommendations, err := s.aiProcessor.GetRecommendations(ctx, req.Customer, req.Items)
	if err != nil {
		s.logger.Warn("Failed to get AI recommendations", zap.Error(err))
		recommendations = []string{} // Empty recommendations on error
	}

	// Calculate estimated wait time
	estimatedWaitTime := s.calculateEstimatedWaitTime(order)

	s.logger.Info("Order created successfully",
		zap.String("order_id", orderID),
		zap.Float64("total_amount", order.TotalAmount),
		zap.Float64("estimated_wait_time", estimatedWaitTime),
	)

	return &pb.CreateOrderResponse{
		Order:                    order,
		AiRecommendations:        recommendations,
		EstimatedWaitTimeMinutes: estimatedWaitTime,
		Success:                  true,
		Message:                  "Order created successfully",
	}, nil
}

// GetOrder retrieves an order by ID with optional AI insights
func (s *Service) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	s.logger.Info("Getting order", zap.String("order_id", req.OrderId))

	if req.OrderId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "order_id is required")
	}

	order, err := s.repository.GetOrder(ctx, req.OrderId)
	if err != nil {
		s.logger.Error("Failed to get order", zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "order not found: %v", err)
	}

	// Refresh AI insights if requested
	if req.IncludeAiInsights && order.AiInsights == nil {
		aiInsights, err := s.aiProcessor.AnalyzeOrder(ctx, order)
		if err != nil {
			s.logger.Warn("Failed to get fresh AI insights", zap.Error(err))
		} else {
			order.AiInsights = aiInsights
		}
	}

	return &pb.GetOrderResponse{
		Order:   order,
		Success: true,
		Message: "Order retrieved successfully",
	}, nil
}

// ListOrders lists orders with AI filtering and analytics
func (s *Service) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	s.logger.Info("Listing orders",
		zap.String("customer_id", req.CustomerId),
		zap.String("location_id", req.LocationId),
		zap.String("status", req.Status.String()),
	)

	orders, totalCount, err := s.repository.ListOrders(ctx, &ListOrdersFilter{
		CustomerID: req.CustomerId,
		LocationID: req.LocationId,
		Status:     req.Status,
		FromDate:   req.FromDate,
		ToDate:     req.ToDate,
		PageSize:   req.PageSize,
		PageToken:  req.PageToken,
	})
	if err != nil {
		s.logger.Error("Failed to list orders", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to list orders: %v", err)
	}

	// Generate AI analytics if enabled
	var aiAnalytics *pb.AIAnalytics
	if req.EnableAiFiltering {
		aiAnalytics, err = s.aiProcessor.GenerateAnalytics(ctx, orders)
		if err != nil {
			s.logger.Warn("Failed to generate AI analytics", zap.Error(err))
		}
	}

	return &pb.ListOrdersResponse{
		Orders:        orders,
		NextPageToken: s.generateNextPageToken(req.PageSize, totalCount),
		TotalCount:    totalCount,
		AiAnalytics:   aiAnalytics,
		Success:       true,
		Message:       "Orders retrieved successfully",
	}, nil
}

// UpdateOrderStatus updates the status of an order with AI validation
func (s *Service) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.UpdateOrderStatusResponse, error) {
	s.logger.Info("Updating order status",
		zap.String("order_id", req.OrderId),
		zap.String("new_status", req.NewStatus.String()),
	)

	if req.OrderId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "order_id is required")
	}

	// Get current order
	order, err := s.repository.GetOrder(ctx, req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "order not found: %v", err)
	}

	// Validate status transition with AI
	if err := s.aiProcessor.ValidateStatusTransition(ctx, order.Status, req.NewStatus); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid status transition: %v", err)
	}

	// Update order status
	order.Status = req.NewStatus
	order.UpdatedAt = timestamppb.Now()

	if err := s.repository.UpdateOrder(ctx, order); err != nil {
		s.logger.Error("Failed to update order", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update order: %v", err)
	}

	// Generate AI notifications
	notifications, err := s.aiProcessor.GenerateStatusNotifications(ctx, order, req.NotifyCustomer)
	if err != nil {
		s.logger.Warn("Failed to generate AI notifications", zap.Error(err))
		notifications = []string{}
	}

	return &pb.UpdateOrderStatusResponse{
		Order:           order,
		AiNotifications: notifications,
		Success:         true,
		Message:         "Order status updated successfully",
	}, nil
}

// CancelOrder cancels an order with AI impact analysis
func (s *Service) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	s.logger.Info("Cancelling order",
		zap.String("order_id", req.OrderId),
		zap.String("reason", req.Reason),
	)

	if req.OrderId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "order_id is required")
	}

	// Get current order
	order, err := s.repository.GetOrder(ctx, req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "order not found: %v", err)
	}

	// Check if order can be cancelled
	if !s.canCancelOrder(order.Status) {
		return nil, status.Errorf(codes.FailedPrecondition, "order cannot be cancelled in current status: %v", order.Status)
	}

	// Generate AI impact analysis
	impactAnalysis, err := s.aiProcessor.AnalyzeCancellationImpact(ctx, order)
	if err != nil {
		s.logger.Warn("Failed to generate impact analysis", zap.Error(err))
		impactAnalysis = &pb.AIImpactAnalysis{
			RevenueImpact:               order.TotalAmount,
			CustomerSatisfactionImpact:  -0.5,
			MitigationSuggestions:       []string{"Offer discount for next order"},
			KitchenEfficiencyImpact:     0.1,
		}
	}

	// Update order status to cancelled
	order.Status = pb.OrderStatus_ORDER_STATUS_CANCELLED
	order.UpdatedAt = timestamppb.Now()

	if err := s.repository.UpdateOrder(ctx, order); err != nil {
		s.logger.Error("Failed to cancel order", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to cancel order: %v", err)
	}

	return &pb.CancelOrderResponse{
		Order:           order,
		ImpactAnalysis:  impactAnalysis,
		Success:         true,
		Message:         "Order cancelled successfully",
	}, nil
}

// Helper methods

func (s *Service) validateCreateOrderRequest(req *pb.CreateOrderRequest) error {
	if req.Customer == nil {
		return fmt.Errorf("customer is required")
	}
	if len(req.Items) == 0 {
		return fmt.Errorf("at least one item is required")
	}
	if req.LocationId == "" {
		return fmt.Errorf("location_id is required")
	}
	return nil
}

func (s *Service) calculateTotalAmount(items []*pb.OrderItem) float64 {
	total := 0.0
	for _, item := range items {
		total += item.TotalPrice
	}
	return total
}

func (s *Service) determinePriorityFromAI(insights *pb.AIOrderInsights) pb.OrderPriority {
	if insights == nil {
		return pb.OrderPriority_ORDER_PRIORITY_NORMAL
	}

	// Simple AI-based priority logic
	if insights.ComplexityScore > 8.0 {
		return pb.OrderPriority_ORDER_PRIORITY_HIGH
	}
	if insights.RevenueImpact > 50.0 {
		return pb.OrderPriority_ORDER_PRIORITY_HIGH
	}
	if insights.CustomerSatisfactionPrediction < 7.0 {
		return pb.OrderPriority_ORDER_PRIORITY_LOW
	}

	return pb.OrderPriority_ORDER_PRIORITY_NORMAL
}

func (s *Service) calculateEstimatedCompletion(insights *pb.AIOrderInsights) *timestamppb.Timestamp {
	if insights == nil || insights.EstimatedPrepTimeMinutes == 0 {
		// Default estimation: 10 minutes
		return timestamppb.New(time.Now().Add(10 * time.Minute))
	}

	estimatedTime := time.Now().Add(time.Duration(insights.EstimatedPrepTimeMinutes) * time.Minute)
	return timestamppb.New(estimatedTime)
}

func (s *Service) calculateEstimatedWaitTime(order *pb.Order) float64 {
	if order.AiInsights != nil && order.AiInsights.EstimatedPrepTimeMinutes > 0 {
		return order.AiInsights.EstimatedPrepTimeMinutes
	}
	// Default estimation based on number of items
	return float64(len(order.Items)) * 2.5 // 2.5 minutes per item
}

func (s *Service) generateNextPageToken(pageSize int32, totalCount int32) string {
	// Simple implementation - in production, use proper pagination tokens
	if pageSize >= totalCount {
		return ""
	}
	return fmt.Sprintf("page_%d", pageSize)
}

func (s *Service) canCancelOrder(status pb.OrderStatus) bool {
	switch status {
	case pb.OrderStatus_ORDER_STATUS_PENDING,
		pb.OrderStatus_ORDER_STATUS_CONFIRMED:
		return true
	case pb.OrderStatus_ORDER_STATUS_PREPARING:
		return true // Can cancel but with impact
	default:
		return false
	}
}

// GetOrderRecommendations gets AI recommendations for order
func (s *Service) GetOrderRecommendations(ctx context.Context, req *pb.GetOrderRecommendationsRequest) (*pb.GetOrderRecommendationsResponse, error) {
	s.logger.Info("Getting AI order recommendations",
		zap.String("customer_id", req.CustomerId),
		zap.String("location_id", req.LocationId),
	)

	if req.CustomerId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "customer_id is required")
	}

	recommendations, reasoning, err := s.aiProcessor.GetOrderRecommendations(ctx, req)
	if err != nil {
		s.logger.Error("Failed to get AI recommendations", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get recommendations: %v", err)
	}

	return &pb.GetOrderRecommendationsResponse{
		Recommendations: recommendations,
		Reasoning:       reasoning,
		Success:         true,
		Message:         "Recommendations generated successfully",
	}, nil
}

// AnalyzeOrderPatterns analyzes order patterns with AI
func (s *Service) AnalyzeOrderPatterns(ctx context.Context, req *pb.AnalyzeOrderPatternsRequest) (*pb.AnalyzeOrderPatternsResponse, error) {
	s.logger.Info("Analyzing order patterns",
		zap.String("location_id", req.LocationId),
		zap.String("analysis_type", req.AnalysisType),
	)

	if req.LocationId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "location_id is required")
	}

	insights, recommendations, err := s.aiProcessor.AnalyzeOrderPatterns(ctx, req)
	if err != nil {
		s.logger.Error("Failed to analyze order patterns", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to analyze patterns: %v", err)
	}

	return &pb.AnalyzeOrderPatternsResponse{
		Insights:        insights,
		Recommendations: recommendations,
		Success:         true,
		Message:         "Pattern analysis completed successfully",
	}, nil
}

// PredictCompletionTime predicts order completion time
func (s *Service) PredictCompletionTime(ctx context.Context, req *pb.PredictCompletionTimeRequest) (*pb.PredictCompletionTimeResponse, error) {
	s.logger.Info("Predicting order completion time",
		zap.String("order_id", req.OrderId),
		zap.String("location_id", req.LocationId),
	)

	if req.OrderId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "order_id is required")
	}

	prediction, err := s.aiProcessor.PredictCompletionTime(ctx, req)
	if err != nil {
		s.logger.Error("Failed to predict completion time", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to predict completion time: %v", err)
	}

	return prediction, nil
}
