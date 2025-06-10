package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	pb "github.com/DimaJoyti/go-coffee/api/proto"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/application"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server represents the gRPC server for kitchen service
type Server struct {
	pb.UnimplementedKitchenServiceServer
	kitchenService application.KitchenService
	logger         *logger.Logger
	grpcServer     *grpc.Server
	port           string
}

// NewServer creates a new gRPC server instance
func NewServer(
	kitchenService application.KitchenService,
	logger *logger.Logger,
	port string,
) *Server {
	return &Server{
		kitchenService: kitchenService,
		logger:         logger,
		port:           port,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		s.logger.WithError(err).WithField("port", s.port).Error("Failed to listen on port")
		return fmt.Errorf("failed to listen on port %s: %w", s.port, err)
	}

	// Create gRPC server with middleware
	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(s.unaryInterceptor),
	)

	// Register service
	pb.RegisterKitchenServiceServer(s.grpcServer, s)

	// Enable reflection for development
	reflection.Register(s.grpcServer)

	s.logger.WithField("port", s.port).Info("Starting gRPC server")

	if err := s.grpcServer.Serve(lis); err != nil {
		s.logger.WithError(err).Error("gRPC server failed")
		return fmt.Errorf("gRPC server failed: %w", err)
	}

	return nil
}

// Stop stops the gRPC server gracefully
func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.logger.Info("Stopping gRPC server")
		s.grpcServer.GracefulStop()
	}
}

// AddToQueue adds an order to kitchen queue with AI prioritization
func (s *Server) AddToQueue(ctx context.Context, req *pb.AddToQueueRequest) (*pb.AddToQueueResponse, error) {
	s.logger.WithField("order_id", req.OrderId).Info("gRPC AddToQueue called")

	// Create queue item from request
	queueItem := &pb.QueueItem{
		OrderId:      req.OrderId,
		CustomerName: req.CustomerName,
		Items:        req.Items,
		Status:       pb.PreparationStatus_PREPARATION_STATUS_QUEUED,
		QueuedAt:     timestamppb.Now(),
		PriorityScore: int32(req.CustomerPriority),
	}

	// For now, return a simple response
	// In a real implementation, you'd call the kitchen service
	response := &pb.AddToQueueResponse{
		QueueItem:         queueItem,
		PositionInQueue:   1,
		EstimatedWaitTime: 15.0, // 15 minutes
		Success:           true,
		Message:           "Order added to queue successfully",
	}

	return response, nil
}

// GetQueue gets current kitchen queue with AI insights
func (s *Server) GetQueue(ctx context.Context, req *pb.GetQueueRequest) (*pb.GetQueueResponse, error) {
	s.logger.WithField("location_id", req.LocationId).Info("gRPC GetQueue called")

	// For now, return an empty queue
	// In a real implementation, you'd call the kitchen service
	response := &pb.GetQueueResponse{
		QueueItems:      []*pb.QueueItem{},
		TotalItems:      0,
		AverageWaitTime: 0.0,
		Success:         true,
		Message:         "Queue retrieved successfully",
	}

	return response, nil
}

// UpdatePreparationStatus updates order preparation status
func (s *Server) UpdatePreparationStatus(ctx context.Context, req *pb.UpdatePreparationStatusRequest) (*pb.UpdatePreparationStatusResponse, error) {
	s.logger.WithFields(map[string]interface{}{
		"order_id": req.OrderId,
		"status":   req.NewStatus,
		"staff_id": req.StaffId,
	}).Info("gRPC UpdatePreparationStatus called")

	// For now, return a simple response
	// In a real implementation, you'd call the kitchen service
	response := &pb.UpdatePreparationStatusResponse{
		UpdatedItem: &pb.QueueItem{
			OrderId: req.OrderId,
			Status:  req.NewStatus,
		},
		Success: true,
		Message: "Preparation status updated successfully",
	}

	return response, nil
}

// CompleteOrder completes order preparation
func (s *Server) CompleteOrder(ctx context.Context, req *pb.CompleteOrderRequest) (*pb.CompleteOrderResponse, error) {
	s.logger.WithFields(map[string]interface{}{
		"order_id":               req.OrderId,
		"staff_id":               req.StaffId,
		"actual_preparation_time": req.ActualPreparationTime,
		"quality_rating":         req.QualityRating,
	}).Info("gRPC CompleteOrder called")

	// For now, return a simple response
	// In a real implementation, you'd call the kitchen service
	response := &pb.CompleteOrderResponse{
		CompletedItem: &pb.QueueItem{
			OrderId: req.OrderId,
			Status:  pb.PreparationStatus_PREPARATION_STATUS_READY,
		},
		Success: true,
		Message: "Order completed successfully",
	}

	return response, nil
}

// GetPerformanceMetrics gets kitchen performance metrics with AI analysis
func (s *Server) GetPerformanceMetrics(ctx context.Context, req *pb.GetPerformanceMetricsRequest) (*pb.GetPerformanceMetricsResponse, error) {
	s.logger.WithField("location_id", req.LocationId).Info("gRPC GetPerformanceMetrics called")

	// For now, return mock metrics
	// In a real implementation, you'd call the kitchen service
	metrics := &pb.PerformanceMetrics{
		AveragePreparationTime: 12.5,
		OrdersCompleted:        150,
		OrdersInQueue:          5,
		EfficiencyRate:         0.85,
		CustomerSatisfaction:   4.2,
	}

	response := &pb.GetPerformanceMetricsResponse{
		Metrics: metrics,
		Success: true,
		Message: "Performance metrics retrieved successfully",
	}

	return response, nil
}

// OptimizeWorkflow optimizes kitchen workflow with AI
func (s *Server) OptimizeWorkflow(ctx context.Context, req *pb.OptimizeWorkflowRequest) (*pb.OptimizeWorkflowResponse, error) {
	s.logger.WithField("location_id", req.LocationId).Info("gRPC OptimizeWorkflow called")

	// For now, return a simple optimization response
	// In a real implementation, you'd call the AI optimization service
	response := &pb.OptimizeWorkflowResponse{
		Optimizations:        []*pb.WorkflowOptimization{},
		EstimatedTimeSavings: 5.0,
		EfficiencyImprovement: 0.1,
		Success:              true,
		Message:              "Workflow optimization completed",
	}

	return response, nil
}

// PredictCapacity predicts kitchen capacity
func (s *Server) PredictCapacity(ctx context.Context, req *pb.PredictCapacityRequest) (*pb.PredictCapacityResponse, error) {
	s.logger.WithFields(map[string]interface{}{
		"location_id":     req.LocationId,
		"expected_orders": req.ExpectedOrders,
	}).Info("gRPC PredictCapacity called")

	// For now, return a simple capacity prediction
	// In a real implementation, you'd call the AI prediction service
	response := &pb.PredictCapacityResponse{
		CapacityUtilization: 0.75,
		CanHandleLoad:       true,
		Success:             true,
		Message:             "Capacity prediction completed",
	}

	return response, nil
}

// GetIngredientRequirements gets ingredient requirements with AI forecasting
func (s *Server) GetIngredientRequirements(ctx context.Context, req *pb.GetIngredientRequirementsRequest) (*pb.GetIngredientRequirementsResponse, error) {
	s.logger.WithField("location_id", req.LocationId).Info("gRPC GetIngredientRequirements called")

	// For now, return empty requirements
	// In a real implementation, you'd call the inventory service
	response := &pb.GetIngredientRequirementsResponse{
		Requirements: []*pb.IngredientRequirement{},
		Success:      true,
		Message:      "Ingredient requirements retrieved successfully",
	}

	return response, nil
}

// AlertKitchenIssue alerts for kitchen issues
func (s *Server) AlertKitchenIssue(ctx context.Context, req *pb.AlertKitchenIssueRequest) (*pb.AlertKitchenIssueResponse, error) {
	s.logger.WithFields(map[string]interface{}{
		"location_id":  req.LocationId,
		"issue_type":   req.IssueType,
		"description":  req.Description,
		"severity":     req.Severity,
		"equipment_id": req.EquipmentId,
	}).Info("gRPC AlertKitchenIssue called")

	// For now, return a simple alert response
	// In a real implementation, you'd call the alert service
	response := &pb.AlertKitchenIssueResponse{
		AlertId:                   fmt.Sprintf("alert-%d", time.Now().Unix()),
		EstimatedResolutionTime:   "30 minutes",
		Success:                   true,
		Message:                   "Kitchen issue alert created successfully",
	}

	return response, nil
}

// unaryInterceptor provides logging and error handling for unary calls
func (s *Server) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	s.logger.WithField("method", info.FullMethod).Info("gRPC call started")

	resp, err := handler(ctx, req)

	duration := time.Since(start)

	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"method":   info.FullMethod,
			"duration": duration,
			"error":    err.Error(),
		}).Error("gRPC call failed")
	} else {
		s.logger.WithFields(map[string]interface{}{
			"method":   info.FullMethod,
			"duration": duration,
		}).Info("gRPC call completed")
	}

	return resp, err
}
