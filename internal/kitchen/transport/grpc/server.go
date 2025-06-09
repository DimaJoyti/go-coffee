package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/kitchen/application"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	pb "github.com/DimaJoyti/go-coffee/proto/kitchen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server represents the gRPC server for kitchen service
type Server struct {
	pb.UnimplementedKitchenServiceServer
	kitchenService      application.KitchenService
	queueService        application.QueueService
	optimizerService    application.OptimizerService
	notificationService application.NotificationService
	logger              *logger.Logger
	grpcServer          *grpc.Server
	port                string
}

// NewServer creates a new gRPC server instance
func NewServer(
	kitchenService application.KitchenService,
	queueService application.QueueService,
	optimizerService application.OptimizerService,
	notificationService application.NotificationService,
	logger *logger.Logger,
	port string,
) *Server {
	return &Server{
		kitchenService:      kitchenService,
		queueService:        queueService,
		optimizerService:    optimizerService,
		notificationService: notificationService,
		logger:              logger,
		port:                port,
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
		grpc.StreamInterceptor(s.streamInterceptor),
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

// Equipment Management gRPC Methods

// CreateEquipment creates new kitchen equipment
func (s *Server) CreateEquipment(ctx context.Context, req *pb.CreateEquipmentRequest) (*pb.EquipmentResponse, error) {
	s.logger.WithField("equipment_id", req.Id).Info("gRPC CreateEquipment called")

	// Convert protobuf request to application request
	appReq := &application.CreateEquipmentRequest{
		ID:          req.Id,
		Name:        req.Name,
		StationType: convertStationType(req.StationType),
		MaxCapacity: req.MaxCapacity,
	}

	// Call application service
	equipment, err := s.kitchenService.CreateEquipment(ctx, appReq)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create equipment")
		return nil, err
	}

	// Convert to protobuf response
	return convertEquipmentToProto(equipment), nil
}

// GetEquipment retrieves equipment by ID
func (s *Server) GetEquipment(ctx context.Context, req *pb.GetEquipmentRequest) (*pb.EquipmentResponse, error) {
	s.logger.WithField("equipment_id", req.Id).Info("gRPC GetEquipment called")

	equipment, err := s.kitchenService.GetEquipment(ctx, req.Id)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get equipment")
		return nil, err
	}

	return convertEquipmentToProto(equipment), nil
}

// UpdateEquipmentStatus updates equipment status
func (s *Server) UpdateEquipmentStatus(ctx context.Context, req *pb.UpdateEquipmentStatusRequest) (*pb.UpdateEquipmentStatusResponse, error) {
	s.logger.WithFields(map[string]interface{}{
		"equipment_id": req.Id,
		"status":       req.Status,
	}).Info("gRPC UpdateEquipmentStatus called")

	err := s.kitchenService.UpdateEquipmentStatus(ctx, req.Id, convertEquipmentStatus(req.Status))
	if err != nil {
		s.logger.WithError(err).Error("Failed to update equipment status")
		return nil, err
	}

	return &pb.UpdateEquipmentStatusResponse{
		Success: true,
		Message: "Equipment status updated successfully",
	}, nil
}

// ListEquipment lists equipment with optional filtering
func (s *Server) ListEquipment(ctx context.Context, req *pb.ListEquipmentRequest) (*pb.ListEquipmentResponse, error) {
	s.logger.Info("gRPC ListEquipment called")

	// Convert protobuf filter to application filter
	filter := &application.EquipmentFilter{
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	if req.StationType != pb.StationType_STATION_TYPE_UNSPECIFIED {
		stationType := convertStationType(req.StationType)
		filter.StationType = &stationType
	}

	if req.Status != pb.EquipmentStatus_EQUIPMENT_STATUS_UNSPECIFIED {
		status := convertEquipmentStatus(req.Status)
		filter.Status = &status
	}

	if req.Available != nil {
		filter.Available = &req.Available.Value
	}

	equipment, err := s.kitchenService.ListEquipment(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list equipment")
		return nil, err
	}

	// Convert to protobuf response
	response := &pb.ListEquipmentResponse{
		Equipment: make([]*pb.EquipmentResponse, len(equipment)),
		Total:     int32(len(equipment)),
	}

	for i, eq := range equipment {
		response.Equipment[i] = convertEquipmentToProto(eq)
	}

	return response, nil
}

// Staff Management gRPC Methods

// CreateStaff creates new kitchen staff
func (s *Server) CreateStaff(ctx context.Context, req *pb.CreateStaffRequest) (*pb.StaffResponse, error) {
	s.logger.WithField("staff_id", req.Id).Info("gRPC CreateStaff called")

	// Convert protobuf specializations
	specializations := make([]domain.StationType, len(req.Specializations))
	for i, spec := range req.Specializations {
		specializations[i] = convertStationType(spec)
	}

	appReq := &application.CreateStaffRequest{
		ID:                  req.Id,
		Name:                req.Name,
		Specializations:     specializations,
		SkillLevel:          req.SkillLevel,
		MaxConcurrentOrders: req.MaxConcurrentOrders,
	}

	staff, err := s.kitchenService.CreateStaff(ctx, appReq)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create staff")
		return nil, err
	}

	return convertStaffToProto(staff), nil
}

// GetStaff retrieves staff by ID
func (s *Server) GetStaff(ctx context.Context, req *pb.GetStaffRequest) (*pb.StaffResponse, error) {
	s.logger.WithField("staff_id", req.Id).Info("gRPC GetStaff called")

	staff, err := s.kitchenService.GetStaff(ctx, req.Id)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get staff")
		return nil, err
	}

	return convertStaffToProto(staff), nil
}

// UpdateStaffAvailability updates staff availability
func (s *Server) UpdateStaffAvailability(ctx context.Context, req *pb.UpdateStaffAvailabilityRequest) (*pb.UpdateStaffAvailabilityResponse, error) {
	s.logger.WithFields(map[string]interface{}{
		"staff_id":  req.Id,
		"available": req.Available,
	}).Info("gRPC UpdateStaffAvailability called")

	err := s.kitchenService.UpdateStaffAvailability(ctx, req.Id, req.Available)
	if err != nil {
		s.logger.WithError(err).Error("Failed to update staff availability")
		return nil, err
	}

	return &pb.UpdateStaffAvailabilityResponse{
		Success: true,
		Message: "Staff availability updated successfully",
	}, nil
}

// Order Management gRPC Methods

// AddOrderToQueue adds an order to the kitchen queue
func (s *Server) AddOrderToQueue(ctx context.Context, req *pb.AddOrderRequest) (*pb.OrderResponse, error) {
	s.logger.WithField("order_id", req.Id).Info("gRPC AddOrderToQueue called")

	// Convert protobuf items to application items
	items := make([]*application.OrderItemRequest, len(req.Items))
	for i, item := range req.Items {
		requirements := make([]domain.StationType, len(item.Requirements))
		for j, req := range item.Requirements {
			requirements[j] = convertStationType(req)
		}

		items[i] = &application.OrderItemRequest{
			ID:           item.Id,
			Name:         item.Name,
			Quantity:     item.Quantity,
			Instructions: item.Instructions,
			Requirements: requirements,
			Metadata:     item.Metadata,
		}
	}

	appReq := &application.AddOrderRequest{
		ID:                  req.Id,
		CustomerID:          req.CustomerId,
		Items:               items,
		Priority:            convertOrderPriority(req.Priority),
		SpecialInstructions: req.SpecialInstructions,
	}

	order, err := s.kitchenService.AddOrderToQueue(ctx, appReq)
	if err != nil {
		s.logger.WithError(err).Error("Failed to add order to queue")
		return nil, err
	}

	return convertOrderToProto(order), nil
}

// GetOrder retrieves an order by ID
func (s *Server) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderResponse, error) {
	s.logger.WithField("order_id", req.Id).Info("gRPC GetOrder called")

	order, err := s.kitchenService.GetOrder(ctx, req.Id)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get order")
		return nil, err
	}

	return convertOrderToProto(order), nil
}

// UpdateOrderStatus updates order status
func (s *Server) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.UpdateOrderStatusResponse, error) {
	s.logger.WithFields(map[string]interface{}{
		"order_id": req.Id,
		"status":   req.Status,
	}).Info("gRPC UpdateOrderStatus called")

	err := s.kitchenService.UpdateOrderStatus(ctx, req.Id, convertOrderStatus(req.Status))
	if err != nil {
		s.logger.WithError(err).Error("Failed to update order status")
		return nil, err
	}

	return &pb.UpdateOrderStatusResponse{
		Success: true,
		Message: "Order status updated successfully",
	}, nil
}

// Queue Management gRPC Methods

// GetQueueStatus returns current queue status
func (s *Server) GetQueueStatus(ctx context.Context, req *pb.GetQueueStatusRequest) (*pb.QueueStatusResponse, error) {
	s.logger.Info("gRPC GetQueueStatus called")

	status, err := s.kitchenService.GetQueueStatus(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get queue status")
		return nil, err
	}

	return convertQueueStatusToProto(status), nil
}

// GetNextOrder returns the next order to be processed
func (s *Server) GetNextOrder(ctx context.Context, req *pb.GetNextOrderRequest) (*pb.GetNextOrderResponse, error) {
	s.logger.Info("gRPC GetNextOrder called")

	order, err := s.kitchenService.GetNextOrder(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get next order")
		return nil, err
	}

	response := &pb.GetNextOrderResponse{}
	if order != nil {
		response.Order = convertOrderToProto(order)
		response.HasOrder = true
	}

	return response, nil
}

// OptimizeQueue optimizes the current queue
func (s *Server) OptimizeQueue(ctx context.Context, req *pb.OptimizeQueueRequest) (*pb.OptimizeQueueResponse, error) {
	s.logger.Info("gRPC OptimizeQueue called")

	optimization, err := s.kitchenService.OptimizeQueue(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to optimize queue")
		return nil, err
	}

	return convertOptimizationToProto(optimization), nil
}

// Middleware

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

// streamInterceptor provides logging for streaming calls
func (s *Server) streamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	start := time.Now()

	s.logger.WithField("method", info.FullMethod).Info("gRPC stream started")

	err := handler(srv, ss)

	duration := time.Since(start)

	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"method":   info.FullMethod,
			"duration": duration,
			"error":    err.Error(),
		}).Error("gRPC stream failed")
	} else {
		s.logger.WithFields(map[string]interface{}{
			"method":   info.FullMethod,
			"duration": duration,
		}).Info("gRPC stream completed")
	}

	return err
}
