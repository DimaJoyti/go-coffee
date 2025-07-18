package dao

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

// Service provides DAO governance operations
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
	governorClient         GovernorClientInterface
	bountyManagerClient    BountyManagerClientInterface
	revenueSharingClient   RevenueSharingClientInterface
	solutionRegistryClient SolutionRegistryClientInterface

	// Repositories
	proposalRepo  ProposalRepository
	developerRepo DeveloperRepository
	voteRepo      VoteRepository

	// Cache
	votingPowerCache map[string]decimal.Decimal
	proposalCache    map[string]*Proposal
}

// ServiceConfig holds the configuration for the DAO service
type ServiceConfig struct {
	DB          *sql.DB
	Redis       redis.Client
	Logger      *logger.Logger
	Config      *config.Config
	ServiceName string
}

// NewService creates a new DAO service instance
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

	// Initialize contract clients
	governorClient, err := NewGovernorClient(ethClient, cfg.Config.Contracts.DAOGovernor, cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize governor client: %w", err)
	}

	bountyManagerClient, err := NewBountyManagerClient(ethClient, cfg.Config.Contracts.BountyManager, cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize bounty manager client: %w", err)
	}

	revenueSharingClient, err := NewRevenueSharingClient(ethClient, cfg.Config.Contracts.RevenueSharing, cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize revenue sharing client: %w", err)
	}

	solutionRegistryClient, err := NewSolutionRegistryClient(ethClient, cfg.Config.Contracts.SolutionRegistry, cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize solution registry client: %w", err)
	}

	// Initialize repositories
	proposalRepo := NewProposalRepository(cfg.DB, cfg.Logger)
	developerRepo := NewDeveloperRepository(cfg.DB, cfg.Logger)
	voteRepo := NewVoteRepository(cfg.DB, cfg.Logger)

	service := &Service{
		db:          cfg.DB,
		redis:       cfg.Redis,
		logger:      cfg.Logger,
		config:      cfg.Config,
		serviceName: cfg.ServiceName,

		ethClient:     ethClient,
		bscClient:     bscClient,
		polygonClient: polygonClient,

		governorClient:         governorClient,
		bountyManagerClient:    bountyManagerClient,
		revenueSharingClient:   revenueSharingClient,
		solutionRegistryClient: solutionRegistryClient,

		proposalRepo:  proposalRepo,
		developerRepo: developerRepo,
		voteRepo:      voteRepo,

		votingPowerCache: make(map[string]decimal.Decimal),
		proposalCache:    make(map[string]*Proposal),
	}

	return service, nil
}

// CreateProposal creates a new governance proposal
func (s *Service) CreateProposal(ctx context.Context, req *CreateProposalRequest) (*CreateProposalResponse, error) {
	s.logger.Info("Creating new proposal",
		zap.String("proposer", req.ProposerAddress),
		zap.String("title", req.Title),
		zap.String("category", req.Category.String()))

	// Validate proposer has enough voting power
	votingPower, err := s.GetVotingPower(ctx, req.ProposerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get voting power: %w", err)
	}

	minThreshold := decimal.NewFromFloat(float64(s.config.DAO.ProposalThreshold))
	if votingPower.LessThan(minThreshold) {
		return nil, fmt.Errorf("insufficient voting power: have %s, need %s",
			votingPower.String(), minThreshold.String())
	}

	// Create proposal on-chain
	proposalID, txHash, err := s.governorClient.CreateProposal(ctx, &GovernorProposalRequest{
		Targets:     req.Targets,
		Values:      req.Values,
		Calldatas:   req.Calldatas,
		Description: req.Description,
		Category:    req.Category,
		Title:       req.Title,
		Deadline:    req.ExecutionDeadline,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal on-chain: %w", err)
	}

	// Store proposal in database
	proposal := &Proposal{
		ProposalID:        proposalID,
		Title:             req.Title,
		Description:       req.Description,
		Category:          req.Category,
		ProposerAddress:   req.ProposerAddress,
		Status:            ProposalStatusPending,
		ExecutionDeadline: req.ExecutionDeadline,
		TransactionHash:   txHash,
		CreatedAt:         time.Now(),
	}

	if err := s.proposalRepo.Create(ctx, proposal); err != nil {
		return nil, fmt.Errorf("failed to store proposal: %w", err)
	}

	// Cache the proposal
	s.proposalCache[proposalID] = proposal

	s.logger.Info("Proposal created successfully",
		zap.String("proposalID", proposalID),
		zap.String("txHash", txHash))

	return &CreateProposalResponse{
		ProposalID:      proposalID,
		TransactionHash: txHash,
		Status:          proposal.Status,
	}, nil
}

// GetProposal retrieves a proposal by ID
func (s *Service) GetProposal(ctx context.Context, proposalID string) (*Proposal, error) {
	// Check cache first
	if proposal, exists := s.proposalCache[proposalID]; exists {
		return proposal, nil
	}

	// Get from database
	proposal, err := s.proposalRepo.GetByID(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal: %w", err)
	}

	// Update cache
	s.proposalCache[proposalID] = proposal

	return proposal, nil
}

// GetProposals retrieves proposals with pagination and filtering
func (s *Service) GetProposals(ctx context.Context, req *GetProposalsRequest) (*GetProposalsResponse, error) {
	proposals, total, err := s.proposalRepo.List(ctx, &ProposalFilter{
		Status:   req.Status,
		Category: req.Category,
		Proposer: req.ProposerAddress,
		Limit:    req.Limit,
		Offset:   req.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get proposals: %w", err)
	}

	return &GetProposalsResponse{
		Proposals: proposals,
		Total:     total,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}, nil
}

// VoteOnProposal casts a vote on a proposal
func (s *Service) VoteOnProposal(ctx context.Context, req *VoteRequest) (*VoteResponse, error) {
	s.logger.Info("Casting vote on proposal",
		zap.String("proposalID", req.ProposalID),
		zap.String("voter", req.VoterAddress),
		zap.String("support", req.Support.String()))

	// Get voting power
	votingPower, err := s.GetVotingPower(ctx, req.VoterAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get voting power: %w", err)
	}

	if votingPower.IsZero() {
		return nil, fmt.Errorf("no voting power")
	}

	// Cast vote on-chain
	txHash, err := s.governorClient.CastVote(ctx, &GovernorVoteRequest{
		ProposalID: req.ProposalID,
		Support:    req.Support,
		Reason:     req.Reason,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to cast vote on-chain: %w", err)
	}

	// Store vote in database
	vote := &Vote{
		ProposalID:      req.ProposalID,
		VoterAddress:    req.VoterAddress,
		Support:         req.Support,
		VotingPower:     votingPower,
		Reason:          req.Reason,
		TransactionHash: txHash,
		VotedAt:         time.Now(),
	}

	if err := s.voteRepo.Create(ctx, vote); err != nil {
		return nil, fmt.Errorf("failed to store vote: %w", err)
	}

	s.logger.Info("Vote cast successfully",
		zap.String("proposalID", req.ProposalID),
		zap.String("txHash", txHash))

	return &VoteResponse{
		TransactionHash: txHash,
		VotingPower:     votingPower,
	}, nil
}

// GetVotingPower retrieves the voting power for an address
func (s *Service) GetVotingPower(ctx context.Context, address string) (decimal.Decimal, error) {
	// Check cache first
	if power, exists := s.votingPowerCache[address]; exists {
		return power, nil
	}

	// Get from contract
	power, err := s.governorClient.GetVotingPower(ctx, address)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get voting power: %w", err)
	}

	// Cache the result
	s.votingPowerCache[address] = power

	return power, nil
}

// GetGovernanceStats retrieves governance statistics
func (s *Service) GetGovernanceStats(ctx context.Context) (*GovernanceStats, error) {
	stats, err := s.proposalRepo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get governance stats: %w", err)
	}

	// Get additional on-chain stats
	totalSupply, err := s.governorClient.GetTotalSupply(ctx)
	if err != nil {
		s.logger.Warn("Failed to get total supply", zap.Error(err))
	}

	stats.TotalSupply = totalSupply
	stats.QuorumPercentage = s.config.DAO.QuorumPercentage

	return stats, nil
}

// StartProposalMonitoring starts the proposal monitoring background service
func (s *Service) StartProposalMonitoring(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	s.logger.Info("Starting proposal monitoring service")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping proposal monitoring service")
			return ctx.Err()
		case <-ticker.C:
			if err := s.syncProposals(ctx); err != nil {
				s.logger.Error("Failed to sync proposals", zap.Error(err))
			}
		}
	}
}

// StartVotingPowerUpdater starts the voting power cache updater
func (s *Service) StartVotingPowerUpdater(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	s.logger.Info("Starting voting power updater")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping voting power updater")
			return ctx.Err()
		case <-ticker.C:
			if err := s.updateVotingPowerCache(ctx); err != nil {
				s.logger.Error("Failed to update voting power cache", zap.Error(err))
			}
		}
	}
}

// syncProposals synchronizes proposals from the blockchain
func (s *Service) syncProposals(ctx context.Context) error {
	// Get active proposals from database
	activeStatus := ProposalStatusActive
	proposals, _, err := s.proposalRepo.List(ctx, &ProposalFilter{
		Status: &activeStatus,
		Limit:  100,
	})
	if err != nil {
		return fmt.Errorf("failed to get active proposals: %w", err)
	}

	// Check status of each proposal on-chain
	for _, proposal := range proposals {
		status, err := s.governorClient.GetProposalStatus(ctx, proposal.ProposalID)
		if err != nil {
			s.logger.Error("Failed to get proposal status",
				zap.String("proposalID", proposal.ProposalID),
				zap.Error(err))
			continue
		}

		// Update if status changed
		if status != proposal.Status {
			proposal.Status = status
			proposal.UpdatedAt = time.Now()

			if err := s.proposalRepo.Update(ctx, proposal); err != nil {
				s.logger.Error("Failed to update proposal status",
					zap.String("proposalID", proposal.ProposalID),
					zap.Error(err))
			} else {
				s.logger.Info("Updated proposal status",
					zap.String("proposalID", proposal.ProposalID),
					zap.String("status", status.String()))
			}

			// Update cache
			s.proposalCache[proposal.ProposalID] = proposal
		}
	}

	return nil
}

// updateVotingPowerCache updates the voting power cache
func (s *Service) updateVotingPowerCache(ctx context.Context) error {
	// Get active developers
	developers, err := s.developerRepo.GetActive(ctx, 1000, 0)
	if err != nil {
		return fmt.Errorf("failed to get active developers: %w", err)
	}

	// Update voting power for each developer
	for _, developer := range developers {
		power, err := s.governorClient.GetVotingPower(ctx, developer.WalletAddress)
		if err != nil {
			s.logger.Error("Failed to get voting power",
				zap.String("address", developer.WalletAddress),
				zap.Error(err))
			continue
		}

		s.votingPowerCache[developer.WalletAddress] = power
	}

	s.logger.Debug("Updated voting power cache",
		zap.Int("developers", len(developers)))

	return nil
}
