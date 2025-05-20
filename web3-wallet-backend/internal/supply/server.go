package supply

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/yourusername/web3-wallet-backend/api/proto/supply"
	"github.com/yourusername/web3-wallet-backend/pkg/logger"
)

// Server represents the gRPC server for the supply service
type Server struct {
	pb.UnimplementedSupplyServiceServer
	service *Service
	logger  *logger.Logger
}

// NewServer creates a new gRPC server for the supply service
func NewServer(service *Service, logger *logger.Logger) *Server {
	return &Server{
		service: service,
		logger:  logger.Named("supply-server"),
	}
}

// GetSupply gets a supply by ID
func (s *Server) GetSupply(ctx context.Context, req *pb.GetSupplyRequest) (*pb.GetSupplyResponse, error) {
	// Validate request
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "supply ID is required")
	}

	// Get supply
	supply, err := s.service.GetSupply(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get supply", "id", req.Id, "error", err)
		return nil, status.Error(codes.Internal, "failed to get supply")
	}

	if supply == nil {
		return nil, status.Error(codes.NotFound, "supply not found")
	}

	// Convert to proto
	pbSupply, err := s.convertToPbSupply(supply)
	if err != nil {
		s.logger.Error("Failed to convert supply to proto", "id", req.Id, "error", err)
		return nil, status.Error(codes.Internal, "failed to convert supply to proto")
	}

	return &pb.GetSupplyResponse{
		Supply: pbSupply,
	}, nil
}

// CreateSupply creates a new supply
func (s *Server) CreateSupply(ctx context.Context, req *pb.CreateSupplyRequest) (*pb.CreateSupplyResponse, error) {
	// Validate request
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}

	if req.Currency == "" {
		return nil, status.Error(codes.InvalidArgument, "currency is required")
	}

	if req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}

	// Create supply
	supply := &Supply{
		ID:        uuid.New().String(),
		UserID:    req.UserId,
		Currency:  req.Currency,
		Amount:    req.Amount,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.service.CreateSupply(ctx, supply); err != nil {
		s.logger.Error("Failed to create supply", "error", err)
		return nil, status.Error(codes.Internal, "failed to create supply")
	}

	// Convert to proto
	pbSupply, err := s.convertToPbSupply(supply)
	if err != nil {
		s.logger.Error("Failed to convert supply to proto", "id", supply.ID, "error", err)
		return nil, status.Error(codes.Internal, "failed to convert supply to proto")
	}

	return &pb.CreateSupplyResponse{
		Supply: pbSupply,
	}, nil
}

// UpdateSupply updates a supply
func (s *Server) UpdateSupply(ctx context.Context, req *pb.UpdateSupplyRequest) (*pb.UpdateSupplyResponse, error) {
	// Validate request
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "supply ID is required")
	}

	// Get supply
	supply, err := s.service.GetSupply(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get supply", "id", req.Id, "error", err)
		return nil, status.Error(codes.Internal, "failed to get supply")
	}

	if supply == nil {
		return nil, status.Error(codes.NotFound, "supply not found")
	}

	// Update supply
	if req.Status != "" {
		supply.Status = req.Status
	}

	if req.Amount > 0 {
		supply.Amount = req.Amount
	}

	supply.UpdatedAt = time.Now()

	if err := s.service.UpdateSupply(ctx, req.Id, supply); err != nil {
		s.logger.Error("Failed to update supply", "id", req.Id, "error", err)
		return nil, status.Error(codes.Internal, "failed to update supply")
	}

	// Convert to proto
	pbSupply, err := s.convertToPbSupply(supply)
	if err != nil {
		s.logger.Error("Failed to convert supply to proto", "id", req.Id, "error", err)
		return nil, status.Error(codes.Internal, "failed to convert supply to proto")
	}

	return &pb.UpdateSupplyResponse{
		Supply: pbSupply,
	}, nil
}

// DeleteSupply deletes a supply
func (s *Server) DeleteSupply(ctx context.Context, req *pb.DeleteSupplyRequest) (*pb.DeleteSupplyResponse, error) {
	// Validate request
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "supply ID is required")
	}

	// Delete supply
	if err := s.service.DeleteSupply(ctx, req.Id); err != nil {
		s.logger.Error("Failed to delete supply", "id", req.Id, "error", err)
		return nil, status.Error(codes.Internal, "failed to delete supply")
	}

	return &pb.DeleteSupplyResponse{
		Success: true,
	}, nil
}

// ListSupplies lists supplies
func (s *Server) ListSupplies(ctx context.Context, req *pb.ListSuppliesRequest) (*pb.ListSuppliesResponse, error) {
	// Get supplies
	supplies, total, err := s.service.ListSupplies(ctx, req.UserId, req.Currency, req.Status, int(req.Page), int(req.PageSize))
	if err != nil {
		s.logger.Error("Failed to list supplies", "error", err)
		return nil, status.Error(codes.Internal, "failed to list supplies")
	}

	// Convert to proto
	pbSupplies := make([]*pb.Supply, 0, len(supplies))
	for _, supply := range supplies {
		pbSupply, err := s.convertToPbSupply(supply)
		if err != nil {
			s.logger.Error("Failed to convert supply to proto", "id", supply.ID, "error", err)
			continue
		}
		pbSupplies = append(pbSupplies, pbSupply)
	}

	return &pb.ListSuppliesResponse{
		Supplies: pbSupplies,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// convertToPbSupply converts a supply to a proto supply
func (s *Server) convertToPbSupply(supply *Supply) (*pb.Supply, error) {
	createdAt, err := ptypes.TimestampProto(supply.CreatedAt)
	if err != nil {
		return nil, err
	}

	updatedAt, err := ptypes.TimestampProto(supply.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &pb.Supply{
		Id:        supply.ID,
		UserId:    supply.UserID,
		Currency:  supply.Currency,
		Amount:    supply.Amount,
		Status:    supply.Status,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
