package bounty

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/config"
	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/logger"
	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/redis"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// Service provides bounty management operations
type Service struct {
	db          *sql.DB
	redis       redis.Client
	logger      *logger.Logger
	config      *config.Config
	serviceName string

	// Blockchain clients
	ethClient     *ethclient.Client
	bscClient     *ethclient.Client
	polygonClient *ethclient.Client

	// Contract clients
	bountyManagerClient BountyManagerClientInterface

	// Repositories
	bountyRepo      BountyRepository
	milestoneRepo   MilestoneRepository
	applicationRepo ApplicationRepository
	developerRepo   DeveloperRepository

	// Cache
	bountyCache      map[uint64]*Bounty
	reputationCache  map[string]int
	performanceCache map[uint64]*PerformanceMetrics
}

// ServiceConfig holds the configuration for the bounty service
type ServiceConfig struct {
	DB          *sql.DB
	Redis       redis.Client
	Logger      *logger.Logger
	Config      *config.Config
	ServiceName string
}

// NewService creates a new bounty service instance
func NewService(cfg ServiceConfig) (*Service, error) {
	// Initialize blockchain clients
	ethClient, err := ethclient.Dial(cfg.Config.Blockchain.Ethereum.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum: %w", err)
	}

	bscClient, err := ethclient.Dial(cfg.Config.Blockchain.BSC.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to BSC: %w", err)
	}

	polygonClient, err := ethclient.Dial(cfg.Config.Blockchain.Polygon.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Polygon: %w", err)
	}

	// Initialize contract client
	bountyManagerClient, err := NewBountyManagerClient(ethClient, cfg.Config.Contracts.BountyManager, cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize bounty manager client: %w", err)
	}

	// Initialize repositories
	bountyRepo := NewBountyRepository(cfg.DB, cfg.Logger)
	milestoneRepo := NewMilestoneRepository(cfg.DB, cfg.Logger)
	applicationRepo := NewApplicationRepository(cfg.DB, cfg.Logger)
	developerRepo := NewDeveloperRepository(cfg.DB, cfg.Logger)

	service := &Service{
		db:          cfg.DB,
		redis:       cfg.Redis,
		logger:      cfg.Logger,
		config:      cfg.Config,
		serviceName: cfg.ServiceName,

		ethClient:     ethClient,
		bscClient:     bscClient,
		polygonClient: polygonClient,

		bountyManagerClient: bountyManagerClient,

		bountyRepo:      bountyRepo,
		milestoneRepo:   milestoneRepo,
		applicationRepo: applicationRepo,
		developerRepo:   developerRepo,

		bountyCache:      make(map[uint64]*Bounty),
		reputationCache:  make(map[string]int),
		performanceCache: make(map[uint64]*PerformanceMetrics),
	}

	return service, nil
}

// CreateBounty creates a new bounty
func (s *Service) CreateBounty(ctx context.Context, req *CreateBountyRequest) (*CreateBountyResponse, error) {
	s.logger.Info("Creating new bounty",
		zap.String("creator", req.CreatorAddress),
		zap.String("title", req.Title),
		zap.String("category", req.Category.String()))

	// Validate request
	if err := s.validateBountyRequest(req); err != nil {
		return nil, fmt.Errorf("invalid bounty request: %w", err)
	}

	// Create bounty on-chain
	bountyID, txHash, err := s.bountyManagerClient.CreateBounty(ctx, &BountyRequest{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Reward:      req.TotalReward,
		Deadline:    req.Deadline.Unix(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create bounty on-chain: %w", err)
	}

	// Store bounty in database
	bounty := &Bounty{
		BountyID:        bountyID,
		Title:           req.Title,
		Description:     req.Description,
		Category:        req.Category,
		Status:          BountyStatusOpen,
		CreatorAddress:  req.CreatorAddress,
		TotalReward:     req.TotalReward,
		Deadline:        req.Deadline,
		CreatedAt:       time.Now(),
		TransactionHash: txHash,
	}

	if err := s.bountyRepo.Create(ctx, bounty); err != nil {
		return nil, fmt.Errorf("failed to store bounty: %w", err)
	}

	// Create milestones
	for i, milestone := range req.Milestones {
		ms := &Milestone{
			BountyID:    bountyID,
			Index:       uint64(i),
			Description: milestone.Description,
			Reward:      milestone.Reward,
			Deadline:    milestone.Deadline,
			Completed:   false,
			Paid:        false,
		}

		if err := s.milestoneRepo.Create(ctx, ms); err != nil {
			s.logger.Error("Failed to create milestone", zap.Error(err))
			// Continue with other milestones
		}
	}

	// Cache the bounty
	s.bountyCache[bountyID] = bounty

	s.logger.Info("Bounty created successfully",
		zap.Uint64("bountyID", bountyID),
		zap.String("txHash", txHash))

	return &CreateBountyResponse{
		BountyID:        bountyID,
		TransactionHash: txHash,
		Status:          bounty.Status,
	}, nil
}

// GetBounty retrieves a bounty by ID
func (s *Service) GetBounty(ctx context.Context, bountyID uint64) (*BountyDetails, error) {
	// Check cache first
	if bounty, exists := s.bountyCache[bountyID]; exists {
		return s.buildBountyDetails(ctx, bounty)
	}

	// Get from database
	bounty, err := s.bountyRepo.GetByID(ctx, bountyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bounty: %w", err)
	}

	// Update cache
	s.bountyCache[bountyID] = bounty

	return s.buildBountyDetails(ctx, bounty)
}

// ApplyForBounty allows a developer to apply for a bounty
func (s *Service) ApplyForBounty(ctx context.Context, req *ApplyBountyRequest) (*ApplyBountyResponse, error) {
	s.logger.Info("Developer applying for bounty",
		zap.Uint64("bountyID", req.BountyID),
		zap.String("applicant", req.ApplicantAddress))

	// Check if bounty exists and is open
	bounty, err := s.bountyRepo.GetByID(ctx, req.BountyID)
	if err != nil {
		return nil, fmt.Errorf("bounty not found: %w", err)
	}

	if bounty.Status != BountyStatusOpen {
		return nil, fmt.Errorf("bounty is not open for applications")
	}

	// Check if already applied
	exists, err := s.applicationRepo.HasApplied(ctx, req.BountyID, req.ApplicantAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to check application status: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("already applied for this bounty")
	}

	// Create application
	application := &Application{
		BountyID:          req.BountyID,
		ApplicantAddress:  req.ApplicantAddress,
		ApplicationMessage: req.Message,
		ProposedTimeline:  req.ProposedTimeline,
		AppliedAt:         time.Now(),
		Status:            ApplicationStatusPending,
	}

	if err := s.applicationRepo.Create(ctx, application); err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	s.logger.Info("Application submitted successfully",
		zap.Uint64("bountyID", req.BountyID),
		zap.String("applicant", req.ApplicantAddress))

	return &ApplyBountyResponse{
		ApplicationID: application.ID,
		Status:        application.Status,
	}, nil
}

// AssignBounty assigns a bounty to a developer
func (s *Service) AssignBounty(ctx context.Context, req *AssignBountyRequest) (*AssignBountyResponse, error) {
	s.logger.Info("Assigning bounty",
		zap.Uint64("bountyID", req.BountyID),
		zap.String("assignee", req.AssigneeAddress))

	// Get bounty
	bounty, err := s.bountyRepo.GetByID(ctx, req.BountyID)
	if err != nil {
		return nil, fmt.Errorf("bounty not found: %w", err)
	}

	// Validate assignment
	if bounty.Status != BountyStatusOpen {
		return nil, fmt.Errorf("bounty is not open for assignment")
	}

	// Check if applicant applied
	hasApplied, err := s.applicationRepo.HasApplied(ctx, req.BountyID, req.AssigneeAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to check application: %w", err)
	}

	if !hasApplied {
		return nil, fmt.Errorf("developer has not applied for this bounty")
	}

	// Assign bounty on-chain
	txHash, err := s.bountyManagerClient.AssignBounty(ctx, req.BountyID, req.AssigneeAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to assign bounty on-chain: %w", err)
	}

	// Update bounty in database
	bounty.Status = BountyStatusAssigned
	bounty.AssigneeAddress = &req.AssigneeAddress
	bounty.AssignedAt = &time.Time{}
	*bounty.AssignedAt = time.Now()
	bounty.UpdatedAt = time.Now()

	if err := s.bountyRepo.Update(ctx, bounty); err != nil {
		return nil, fmt.Errorf("failed to update bounty: %w", err)
	}

	// Update application status
	if err := s.applicationRepo.UpdateStatus(ctx, req.BountyID, req.AssigneeAddress, ApplicationStatusAccepted); err != nil {
		s.logger.Error("Failed to update application status", zap.Error(err))
	}

	// Update cache
	s.bountyCache[req.BountyID] = bounty

	s.logger.Info("Bounty assigned successfully",
		zap.Uint64("bountyID", req.BountyID),
		zap.String("txHash", txHash))

	return &AssignBountyResponse{
		TransactionHash: txHash,
		Status:          bounty.Status,
	}, nil
}

// CompleteMilestone marks a milestone as completed
func (s *Service) CompleteMilestone(ctx context.Context, req *CompleteMilestoneRequest) (*CompleteMilestoneResponse, error) {
	s.logger.Info("Completing milestone",
		zap.Uint64("bountyID", req.BountyID),
		zap.Uint64("milestoneIndex", req.MilestoneIndex))

	// Get milestone
	milestone, err := s.milestoneRepo.GetByBountyAndIndex(ctx, req.BountyID, req.MilestoneIndex)
	if err != nil {
		return nil, fmt.Errorf("milestone not found: %w", err)
	}

	if milestone.Completed {
		return nil, fmt.Errorf("milestone already completed")
	}

	// Complete milestone on-chain
	txHash, err := s.bountyManagerClient.CompleteMilestone(ctx, req.BountyID, req.MilestoneIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to complete milestone on-chain: %w", err)
	}

	// Update milestone in database
	milestone.Completed = true
	milestone.Paid = true
	milestone.CompletedAt = &time.Time{}
	*milestone.CompletedAt = time.Now()

	if err := s.milestoneRepo.Update(ctx, milestone); err != nil {
		return nil, fmt.Errorf("failed to update milestone: %w", err)
	}

	// Update developer reputation
	bounty, err := s.bountyRepo.GetByID(ctx, req.BountyID)
	if err == nil && bounty.AssigneeAddress != nil {
		if err := s.updateDeveloperReputation(ctx, *bounty.AssigneeAddress, 10); err != nil {
			s.logger.Error("Failed to update developer reputation", zap.Error(err))
		}
	}

	s.logger.Info("Milestone completed successfully",
		zap.Uint64("bountyID", req.BountyID),
		zap.Uint64("milestoneIndex", req.MilestoneIndex),
		zap.String("txHash", txHash))

	return &CompleteMilestoneResponse{
		TransactionHash: txHash,
		Reward:          milestone.Reward,
	}, nil
}

// StartBountyMonitoring starts the bounty monitoring background service
func (s *Service) StartBountyMonitoring(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	s.logger.Info("Starting bounty monitoring service")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping bounty monitoring service")
			return ctx.Err()
		case <-ticker.C:
			if err := s.syncBounties(ctx); err != nil {
				s.logger.Error("Failed to sync bounties", zap.Error(err))
			}
		}
	}
}

// StartPerformanceTracking starts the performance tracking background service
func (s *Service) StartPerformanceTracking(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	s.logger.Info("Starting performance tracking service")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping performance tracking service")
			return ctx.Err()
		case <-ticker.C:
			if err := s.trackPerformance(ctx); err != nil {
				s.logger.Error("Failed to track performance", zap.Error(err))
			}
		}
	}
}

// Helper methods

func (s *Service) validateBountyRequest(req *CreateBountyRequest) error {
	if req.Title == "" {
		return fmt.Errorf("title is required")
	}
	if req.Description == "" {
		return fmt.Errorf("description is required")
	}
	if req.TotalReward.LessThan(decimal.NewFromFloat(100)) {
		return fmt.Errorf("reward must be at least 100 tokens")
	}
	if req.Deadline.Before(time.Now()) {
		return fmt.Errorf("deadline must be in the future")
	}
	return nil
}

func (s *Service) buildBountyDetails(ctx context.Context, bounty *Bounty) (*BountyDetails, error) {
	// Get milestones
	milestones, err := s.milestoneRepo.GetByBountyID(ctx, bounty.BountyID)
	if err != nil {
		s.logger.Error("Failed to get milestones", zap.Error(err))
		milestones = []*Milestone{} // Continue with empty milestones
	}

	// Get applications
	applications, err := s.applicationRepo.GetByBountyID(ctx, bounty.BountyID)
	if err != nil {
		s.logger.Error("Failed to get applications", zap.Error(err))
		applications = []*Application{} // Continue with empty applications
	}

	return &BountyDetails{
		Bounty:       bounty,
		Milestones:   milestones,
		Applications: applications,
	}, nil
}

func (s *Service) syncBounties(ctx context.Context) error {
	// Get active bounties from database
	bounties, err := s.bountyRepo.GetActive(ctx, 100, 0)
	if err != nil {
		return fmt.Errorf("failed to get active bounties: %w", err)
	}

	// Check status of each bounty on-chain
	for _, bounty := range bounties {
		// In a real implementation, we would check the contract status
		// For now, we'll just log the sync operation
		s.logger.Debug("Syncing bounty",
			zap.Uint64("bountyID", bounty.BountyID),
			zap.String("status", bounty.Status.String()))
	}

	return nil
}

func (s *Service) trackPerformance(ctx context.Context) error {
	// Get completed bounties
	bounties, err := s.bountyRepo.GetCompleted(ctx, 50, 0)
	if err != nil {
		return fmt.Errorf("failed to get completed bounties: %w", err)
	}

	// Track performance for each bounty
	for _, bounty := range bounties {
		if !bounty.PerformanceVerified {
			// In a real implementation, we would measure actual TVL/MAU impact
			// For now, we'll simulate performance tracking
			s.logger.Debug("Tracking performance for bounty",
				zap.Uint64("bountyID", bounty.BountyID))
		}
	}

	return nil
}

func (s *Service) updateDeveloperReputation(ctx context.Context, address string, points int) error {
	// Update reputation in cache
	if current, exists := s.reputationCache[address]; exists {
		s.reputationCache[address] = current + points
	} else {
		s.reputationCache[address] = points
	}

	// Update in database (would be implemented in a real system)
	s.logger.Debug("Updated developer reputation",
		zap.String("address", address),
		zap.Int("points", points))

	return nil
}
