package bounty

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BountyServiceServer defines the gRPC service interface
type BountyServiceServer interface {
	CreateBounty(context.Context, *CreateBountyRequest) (*CreateBountyResponse, error)
	GetBounty(context.Context, *GetBountyRequest) (*BountyDetails, error)
	ApplyForBounty(context.Context, *ApplyBountyRequest) (*ApplyBountyResponse, error)
	AssignBounty(context.Context, *AssignBountyRequest) (*AssignBountyResponse, error)
	CompleteMilestone(context.Context, *CompleteMilestoneRequest) (*CompleteMilestoneResponse, error)
	VerifyPerformance(context.Context, *VerifyPerformanceRequest) (*VerifyPerformanceResponse, error)
}

// GetBountyRequest represents a gRPC request to get a bounty
type GetBountyRequest struct {
	BountyID uint64 `json:"bounty_id"`
}

// RegisterBountyServiceServer registers the bounty service with gRPC server
func RegisterBountyServiceServer(s *grpc.Server, srv BountyServiceServer) {
	// In a real implementation, this would register the generated gRPC service
	// For now, we'll just log that the service is registered
}

// gRPC method implementations

// GetBountyGRPC retrieves a bounty via gRPC
func (s *Service) GetBountyGRPC(ctx context.Context, req *GetBountyRequest) (*BountyDetails, error) {
	if req.BountyID == 0 {
		return nil, status.Error(codes.InvalidArgument, "bounty ID is required")
	}

	bountyDetails, err := s.GetBounty(ctx, req.BountyID)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return bountyDetails, nil
}

// VerifyPerformance verifies bounty performance via gRPC
func (s *Service) VerifyPerformance(ctx context.Context, req *VerifyPerformanceRequest) (*VerifyPerformanceResponse, error) {
	if req.BountyID == 0 {
		return nil, status.Error(codes.InvalidArgument, "bounty ID is required")
	}

	// Get bounty
	bounty, err := s.bountyRepo.GetByID(ctx, req.BountyID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "bounty not found")
	}

	// Update performance metrics
	bounty.TVLImpact = req.TVLImpact
	bounty.MAUImpact = req.MAUImpact
	bounty.PerformanceVerified = true

	if err := s.bountyRepo.Update(ctx, bounty); err != nil {
		return nil, status.Error(codes.Internal, "failed to update bounty performance")
	}

	// Calculate bonus based on performance
	bonusEarned := s.calculatePerformanceBonus(req.TVLImpact, req.MAUImpact)
	reputationBonus := s.calculateReputationBonus(req.TVLImpact, req.MAUImpact)

	response := &VerifyPerformanceResponse{
		BountyID:        req.BountyID,
		TVLImpact:       req.TVLImpact,
		MAUImpact:       req.MAUImpact,
		BonusEarned:     bonusEarned,
		ReputationBonus: reputationBonus,
	}

	return response, nil
}
