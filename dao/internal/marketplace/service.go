package marketplace

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/config"
	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/logger"
	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/redis"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// Service provides solution marketplace operations
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
	solutionRegistryClient SolutionRegistryClientInterface

	// Repositories
	solutionRepo     SolutionRepository
	reviewRepo       ReviewRepository
	installationRepo InstallationRepository
	categoryRepo     CategoryRepository

	// Cache
	solutionCache      map[uint64]*Solution
	qualityScoreCache  map[uint64]*QualityScore
	compatibilityCache map[string]*CompatibilityResult
}

// ServiceConfig holds the configuration for the marketplace service
type ServiceConfig struct {
	DB          *sql.DB
	Redis       redis.Client
	Logger      *logger.Logger
	Config      *config.Config
	ServiceName string
}

// NewService creates a new marketplace service instance
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
	solutionRegistryClient, err := NewSolutionRegistryClient(ethClient, cfg.Config.Contracts.SolutionRegistry, cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize solution registry client: %w", err)
	}

	// Initialize repositories
	solutionRepo := NewSolutionRepository(cfg.DB, cfg.Logger)
	reviewRepo := NewReviewRepository(cfg.DB, cfg.Logger)
	installationRepo := NewInstallationRepository(cfg.DB, cfg.Logger)
	categoryRepo := NewCategoryRepository(cfg.DB, cfg.Logger)

	service := &Service{
		db:          cfg.DB,
		redis:       cfg.Redis,
		logger:      cfg.Logger,
		config:      cfg.Config,
		serviceName: cfg.ServiceName,

		ethClient:     ethClient,
		bscClient:     bscClient,
		polygonClient: polygonClient,

		solutionRegistryClient: solutionRegistryClient,

		solutionRepo:     solutionRepo,
		reviewRepo:       reviewRepo,
		installationRepo: installationRepo,
		categoryRepo:     categoryRepo,

		solutionCache:      make(map[uint64]*Solution),
		qualityScoreCache:  make(map[uint64]*QualityScore),
		compatibilityCache: make(map[string]*CompatibilityResult),
	}

	return service, nil
}

// CreateSolution creates a new solution in the marketplace
func (s *Service) CreateSolution(ctx context.Context, req *CreateSolutionRequest) (*CreateSolutionResponse, error) {
	s.logger.Info("Creating new solution",
		zap.String("developer", req.DeveloperAddress),
		zap.String("name", req.Name),
		zap.String("category", req.Category.String()))

	// Validate request
	if err := s.validateSolutionRequest(req); err != nil {
		return nil, fmt.Errorf("invalid solution request: %w", err)
	}

	// Register solution on-chain
	solutionID, txHash, err := s.solutionRegistryClient.SubmitSolution(ctx, &SolutionSubmissionRequest{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Version:     req.Version,
		Developer:   req.DeveloperAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register solution on-chain: %w", err)
	}

	// Store solution in database
	solution := &Solution{
		SolutionID:       solutionID,
		Name:             req.Name,
		Description:      req.Description,
		Category:         req.Category,
		Version:          req.Version,
		DeveloperAddress: req.DeveloperAddress,
		Status:           SolutionStatusPending,
		RepositoryURL:    req.RepositoryURL,
		DocumentationURL: req.DocumentationURL,
		DemoURL:          req.DemoURL,
		Tags:             req.Tags,
		CreatedAt:        time.Now(),
		TransactionHash:  txHash,
	}

	if err := s.solutionRepo.Create(ctx, solution); err != nil {
		return nil, fmt.Errorf("failed to store solution: %w", err)
	}

	// Calculate initial quality score
	qualityScore, err := s.calculateQualityScore(ctx, solution)
	if err != nil {
		s.logger.Error("Failed to calculate quality score", zap.Error(err))
		// Continue without quality score for now
	} else {
		s.qualityScoreCache[solutionID] = qualityScore
	}

	// Cache the solution
	s.solutionCache[solutionID] = solution

	s.logger.Info("Solution created successfully",
		zap.Uint64("solutionID", solutionID),
		zap.String("txHash", txHash))

	return &CreateSolutionResponse{
		SolutionID:      solutionID,
		TransactionHash: txHash,
		Status:          solution.Status,
		QualityScore:    qualityScore,
	}, nil
}

// GetSolution retrieves a solution by ID
func (s *Service) GetSolution(ctx context.Context, solutionID uint64) (*SolutionDetails, error) {
	// Check cache first
	if solution, exists := s.solutionCache[solutionID]; exists {
		return s.buildSolutionDetails(ctx, solution)
	}

	// Get from database
	solution, err := s.solutionRepo.GetByID(ctx, solutionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get solution: %w", err)
	}

	// Update cache
	s.solutionCache[solutionID] = solution

	return s.buildSolutionDetails(ctx, solution)
}

// ReviewSolution adds a review for a solution
func (s *Service) ReviewSolution(ctx context.Context, req *ReviewSolutionRequest) (*ReviewSolutionResponse, error) {
	s.logger.Info("Adding solution review",
		zap.Uint64("solutionID", req.SolutionID),
		zap.String("reviewer", req.ReviewerAddress),
		zap.Int("rating", req.Rating))

	// Validate request
	if err := s.validateReviewRequest(req); err != nil {
		return nil, fmt.Errorf("invalid review request: %w", err)
	}

	// Check if solution exists
	solution, err := s.solutionRepo.GetByID(ctx, req.SolutionID)
	if err != nil {
		return nil, fmt.Errorf("solution not found: %w", err)
	}

	// Check if user already reviewed
	hasReviewed, err := s.reviewRepo.HasReviewed(ctx, req.SolutionID, req.ReviewerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to check review status: %w", err)
	}

	if hasReviewed {
		return nil, fmt.Errorf("user has already reviewed this solution")
	}

	// Create review
	review := &Review{
		SolutionID:         req.SolutionID,
		ReviewerAddress:    req.ReviewerAddress,
		Rating:             req.Rating,
		Comment:            req.Comment,
		SecurityScore:      req.SecurityScore,
		PerformanceScore:   req.PerformanceScore,
		UsabilityScore:     req.UsabilityScore,
		DocumentationScore: req.DocumentationScore,
		CreatedAt:          time.Now(),
	}

	if err := s.reviewRepo.Create(ctx, review); err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	// Update solution quality score
	qualityScore, err := s.calculateQualityScore(ctx, solution)
	if err != nil {
		s.logger.Error("Failed to update quality score", zap.Error(err))
	} else {
		s.qualityScoreCache[req.SolutionID] = qualityScore
	}

	s.logger.Info("Review added successfully",
		zap.Uint64("solutionID", req.SolutionID),
		zap.Int64("reviewID", review.ID))

	return &ReviewSolutionResponse{
		ReviewID:     review.ID,
		QualityScore: qualityScore,
	}, nil
}

// ApproveSolution approves a solution for marketplace listing
func (s *Service) ApproveSolution(ctx context.Context, req *ApproveSolutionRequest) (*ApproveSolutionResponse, error) {
	s.logger.Info("Approving solution",
		zap.Uint64("solutionID", req.SolutionID),
		zap.String("approver", req.ApproverAddress))

	// Get solution
	solution, err := s.solutionRepo.GetByID(ctx, req.SolutionID)
	if err != nil {
		return nil, fmt.Errorf("solution not found: %w", err)
	}

	// Validate approval
	if solution.Status != SolutionStatusPending {
		return nil, fmt.Errorf("solution is not pending approval")
	}

	// Approve solution on-chain
	txHash, err := s.solutionRegistryClient.ApproveSolution(ctx, req.SolutionID)
	if err != nil {
		return nil, fmt.Errorf("failed to approve solution on-chain: %w", err)
	}

	// Update solution in database
	solution.Status = SolutionStatusApproved
	solution.ApproverAddress = &req.ApproverAddress
	solution.ApprovedAt = &time.Time{}
	*solution.ApprovedAt = time.Now()
	solution.UpdatedAt = time.Now()

	if err := s.solutionRepo.Update(ctx, solution); err != nil {
		return nil, fmt.Errorf("failed to update solution: %w", err)
	}

	// Update cache
	s.solutionCache[req.SolutionID] = solution

	s.logger.Info("Solution approved successfully",
		zap.Uint64("solutionID", req.SolutionID),
		zap.String("txHash", txHash))

	return &ApproveSolutionResponse{
		TransactionHash: txHash,
		Status:          solution.Status,
	}, nil
}

// InstallSolution records a solution installation
func (s *Service) InstallSolution(ctx context.Context, req *InstallSolutionRequest) (*InstallSolutionResponse, error) {
	s.logger.Info("Installing solution",
		zap.Uint64("solutionID", req.SolutionID),
		zap.String("installer", req.InstallerAddress))

	// Check if solution exists and is approved
	solution, err := s.solutionRepo.GetByID(ctx, req.SolutionID)
	if err != nil {
		return nil, fmt.Errorf("solution not found: %w", err)
	}

	if solution.Status != SolutionStatusApproved {
		return nil, fmt.Errorf("solution is not approved for installation")
	}

	// Check compatibility
	compatible, err := s.checkCompatibility(ctx, req.SolutionID, req.Environment)
	if err != nil {
		s.logger.Error("Failed to check compatibility", zap.Error(err))
		// Continue with installation but log warning
	} else if !compatible.IsCompatible {
		return nil, fmt.Errorf("solution is not compatible with environment: %s", compatible.Issues)
	}

	// Record installation
	installation := &Installation{
		SolutionID:       req.SolutionID,
		InstallerAddress: req.InstallerAddress,
		Environment:      req.Environment,
		Version:          solution.Version,
		InstalledAt:      time.Now(),
		Status:           InstallationStatusActive,
	}

	if err := s.installationRepo.Create(ctx, installation); err != nil {
		return nil, fmt.Errorf("failed to record installation: %w", err)
	}

	// Update solution installation count
	if err := s.solutionRepo.IncrementInstallCount(ctx, req.SolutionID); err != nil {
		s.logger.Error("Failed to update install count", zap.Error(err))
	}

	s.logger.Info("Solution installed successfully",
		zap.Uint64("solutionID", req.SolutionID),
		zap.Int64("installationID", installation.ID))

	return &InstallSolutionResponse{
		InstallationID: installation.ID,
		Status:         installation.Status,
		Version:        installation.Version,
	}, nil
}

// StartQualityMonitoring starts the quality monitoring background service
func (s *Service) StartQualityMonitoring(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	s.logger.Info("Starting quality monitoring service")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping quality monitoring service")
			return ctx.Err()
		case <-ticker.C:
			if err := s.updateQualityScores(ctx); err != nil {
				s.logger.Error("Failed to update quality scores", zap.Error(err))
			}
		}
	}
}

// StartAnalyticsService starts the analytics background service
func (s *Service) StartAnalyticsService(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	s.logger.Info("Starting analytics service")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping analytics service")
			return ctx.Err()
		case <-ticker.C:
			if err := s.updateAnalytics(ctx); err != nil {
				s.logger.Error("Failed to update analytics", zap.Error(err))
			}
		}
	}
}

// StartCompatibilityService starts the compatibility checking background service
func (s *Service) StartCompatibilityService(ctx context.Context) error {
	ticker := time.NewTicker(2 * time.Hour)
	defer ticker.Stop()

	s.logger.Info("Starting compatibility service")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping compatibility service")
			return ctx.Err()
		case <-ticker.C:
			if err := s.updateCompatibilityMatrix(ctx); err != nil {
				s.logger.Error("Failed to update compatibility matrix", zap.Error(err))
			}
		}
	}
}

// Helper methods

func (s *Service) validateSolutionRequest(req *CreateSolutionRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Description == "" {
		return fmt.Errorf("description is required")
	}
	if req.DeveloperAddress == "" {
		return fmt.Errorf("developer address is required")
	}
	if req.Version == "" {
		return fmt.Errorf("version is required")
	}
	return nil
}

func (s *Service) validateReviewRequest(req *ReviewSolutionRequest) error {
	if req.SolutionID == 0 {
		return fmt.Errorf("solution ID is required")
	}
	if req.ReviewerAddress == "" {
		return fmt.Errorf("reviewer address is required")
	}
	if req.Rating < 1 || req.Rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}
	return nil
}

func (s *Service) buildSolutionDetails(ctx context.Context, solution *Solution) (*SolutionDetails, error) {
	// Get reviews
	reviews, err := s.reviewRepo.GetBySolutionID(ctx, solution.SolutionID)
	if err != nil {
		s.logger.Error("Failed to get reviews", zap.Error(err))
		reviews = []*Review{} // Continue with empty reviews
	}

	// Get installations
	installations, err := s.installationRepo.GetBySolutionID(ctx, solution.SolutionID)
	if err != nil {
		s.logger.Error("Failed to get installations", zap.Error(err))
		installations = []*Installation{} // Continue with empty installations
	}

	// Get quality score
	qualityScore := s.qualityScoreCache[solution.SolutionID]

	return &SolutionDetails{
		Solution:      solution,
		Reviews:       reviews,
		Installations: installations,
		QualityScore:  qualityScore,
	}, nil
}

func (s *Service) calculateQualityScore(ctx context.Context, solution *Solution) (*QualityScore, error) {
	// Get all reviews for the solution
	reviews, err := s.reviewRepo.GetBySolutionID(ctx, solution.SolutionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}

	if len(reviews) == 0 {
		// Default quality score for new solutions
		return &QualityScore{
			SolutionID:         solution.SolutionID,
			OverallScore:       3.0, // Default neutral score
			SecurityScore:      3.0,
			PerformanceScore:   3.0,
			UsabilityScore:     3.0,
			DocumentationScore: 3.0,
			ReviewCount:        0,
			LastUpdated:        time.Now(),
		}, nil
	}

	// Calculate average scores
	var totalRating, totalSecurity, totalPerformance, totalUsability, totalDocumentation float64
	for _, review := range reviews {
		totalRating += float64(review.Rating)
		totalSecurity += float64(review.SecurityScore)
		totalPerformance += float64(review.PerformanceScore)
		totalUsability += float64(review.UsabilityScore)
		totalDocumentation += float64(review.DocumentationScore)
	}

	count := float64(len(reviews))
	qualityScore := &QualityScore{
		SolutionID:         solution.SolutionID,
		OverallScore:       totalRating / count,
		SecurityScore:      totalSecurity / count,
		PerformanceScore:   totalPerformance / count,
		UsabilityScore:     totalUsability / count,
		DocumentationScore: totalDocumentation / count,
		ReviewCount:        len(reviews),
		LastUpdated:        time.Now(),
	}

	return qualityScore, nil
}

func (s *Service) checkCompatibility(ctx context.Context, solutionID uint64, environment string) (*CompatibilityResult, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("%d:%s", solutionID, environment)
	if result, exists := s.compatibilityCache[cacheKey]; exists {
		return result, nil
	}

	// In a real implementation, this would check actual compatibility
	// For now, we'll simulate compatibility checking
	result := &CompatibilityResult{
		SolutionID:   solutionID,
		Environment:  environment,
		IsCompatible: true,
		Issues:       []string{},
		CheckedAt:    time.Now(),
	}

	// Cache the result
	s.compatibilityCache[cacheKey] = result

	return result, nil
}

func (s *Service) updateQualityScores(ctx context.Context) error {
	// Get all solutions that need quality score updates
	solutions, err := s.solutionRepo.GetRecentlyUpdated(ctx, 24*time.Hour, 100, 0)
	if err != nil {
		return fmt.Errorf("failed to get recently updated solutions: %w", err)
	}

	for _, solution := range solutions {
		qualityScore, err := s.calculateQualityScore(ctx, solution)
		if err != nil {
			s.logger.Error("Failed to calculate quality score",
				zap.Uint64("solutionID", solution.SolutionID),
				zap.Error(err))
			continue
		}

		s.qualityScoreCache[solution.SolutionID] = qualityScore
	}

	s.logger.Debug("Updated quality scores", zap.Int("count", len(solutions)))
	return nil
}

func (s *Service) updateAnalytics(ctx context.Context) error {
	// Update analytics data (implementation would track trends, popular solutions, etc.)
	s.logger.Debug("Updated analytics data")
	return nil
}

func (s *Service) updateCompatibilityMatrix(ctx context.Context) error {
	// Update compatibility matrix (implementation would check solution compatibility)
	s.logger.Debug("Updated compatibility matrix")
	return nil
}
