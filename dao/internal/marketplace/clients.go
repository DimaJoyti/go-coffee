package marketplace

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// SolutionRegistryClientInterface defines the interface for solution registry client
type SolutionRegistryClientInterface interface {
	SubmitSolution(ctx context.Context, req *SolutionSubmissionRequest) (uint64, string, error)
	ApproveSolution(ctx context.Context, solutionID uint64) (string, error)
	GetSolutionInfo(ctx context.Context, solutionID uint64) (*SolutionInfo, error)
}

// SolutionInfo represents solution information from the blockchain
type SolutionInfo struct {
	SolutionID  uint64
	Developer   string
	IsApproved  bool
	QualityScore uint8
	InstallCount uint64
}

// SolutionRegistryClient handles interactions with the SolutionRegistry contract
type SolutionRegistryClient struct {
	client   *ethclient.Client
	contract common.Address
	logger   *logger.Logger
}

// NewSolutionRegistryClient creates a new solution registry client
func NewSolutionRegistryClient(client *ethclient.Client, contractAddress string, logger *logger.Logger) (SolutionRegistryClientInterface, error) {
	if !common.IsHexAddress(contractAddress) {
		return nil, fmt.Errorf("invalid contract address: %s", contractAddress)
	}

	return &SolutionRegistryClient{
		client:   client,
		contract: common.HexToAddress(contractAddress),
		logger:   logger,
	}, nil
}

// SubmitSolution submits a new solution to the registry
func (src *SolutionRegistryClient) SubmitSolution(ctx context.Context, req *SolutionSubmissionRequest) (uint64, string, error) {
	src.logger.Info("Submitting solution to registry",
		zap.String("name", req.Name),
		zap.String("developer", req.Developer))

	// Mock solution ID and transaction hash - in real implementation, this would interact with the contract
	solutionID := uint64(1)
	txHash := "0xsolution1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	src.logger.Info("Solution submitted to registry",
		zap.Uint64("solutionID", solutionID),
		zap.String("txHash", txHash))

	return solutionID, txHash, nil
}

// ApproveSolution approves a solution in the registry
func (src *SolutionRegistryClient) ApproveSolution(ctx context.Context, solutionID uint64) (string, error) {
	src.logger.Info("Approving solution in registry",
		zap.Uint64("solutionID", solutionID))

	// Mock transaction hash
	txHash := "0xapprove1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	return txHash, nil
}

// GetSolutionInfo retrieves solution information from the registry
func (src *SolutionRegistryClient) GetSolutionInfo(ctx context.Context, solutionID uint64) (*SolutionInfo, error) {
	src.logger.Info("Getting solution info from registry",
		zap.Uint64("solutionID", solutionID))

	// Mock solution info
	info := &SolutionInfo{
		SolutionID:   solutionID,
		Developer:    "0x1234567890123456789012345678901234567890",
		IsApproved:   true,
		QualityScore: 85,
		InstallCount: 42,
	}

	return info, nil
}

// Additional repository implementations

// reviewRepository implements ReviewRepository
type reviewRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewReviewRepository creates a new review repository
func NewReviewRepository(db *sql.DB, logger *logger.Logger) ReviewRepository {
	return &reviewRepository{
		db:     db,
		logger: logger,
	}
}

func (r *reviewRepository) Create(ctx context.Context, review *Review) error {
	query := `
		INSERT INTO solution_reviews (
			solution_id, reviewer_address, rating, comment, security_score,
			performance_score, usability_score, documentation_score, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		review.SolutionID,
		review.ReviewerAddress,
		review.Rating,
		review.Comment,
		review.SecurityScore,
		review.PerformanceScore,
		review.UsabilityScore,
		review.DocumentationScore,
		review.CreatedAt,
		review.UpdatedAt,
	).Scan(&review.ID)

	if err != nil {
		r.logger.Error("Failed to create review", zap.Error(err))
		return fmt.Errorf("failed to create review: %w", err)
	}

	return nil
}

func (r *reviewRepository) GetByID(ctx context.Context, id int64) (*Review, error) {
	query := `
		SELECT id, solution_id, reviewer_address, rating, comment, security_score,
			   performance_score, usability_score, documentation_score, created_at, updated_at
		FROM solution_reviews 
		WHERE id = $1`

	review := &Review{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&review.ID,
		&review.SolutionID,
		&review.ReviewerAddress,
		&review.Rating,
		&review.Comment,
		&review.SecurityScore,
		&review.PerformanceScore,
		&review.UsabilityScore,
		&review.DocumentationScore,
		&review.CreatedAt,
		&review.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("review not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get review: %w", err)
	}

	return review, nil
}

func (r *reviewRepository) GetBySolutionID(ctx context.Context, solutionID uint64) ([]*Review, error) {
	query := `
		SELECT id, solution_id, reviewer_address, rating, comment, security_score,
			   performance_score, usability_score, documentation_score, created_at, updated_at
		FROM solution_reviews 
		WHERE solution_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, solutionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}
	defer rows.Close()

	var reviews []*Review
	for rows.Next() {
		review := &Review{}
		err := rows.Scan(
			&review.ID,
			&review.SolutionID,
			&review.ReviewerAddress,
			&review.Rating,
			&review.Comment,
			&review.SecurityScore,
			&review.PerformanceScore,
			&review.UsabilityScore,
			&review.DocumentationScore,
			&review.CreatedAt,
			&review.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan review: %w", err)
		}
		reviews = append(reviews, review)
	}

	return reviews, nil
}

func (r *reviewRepository) GetByReviewer(ctx context.Context, reviewerAddress string, limit, offset int) ([]*Review, error) {
	// Implementation similar to GetBySolutionID but with reviewer filter
	return []*Review{}, nil // Simplified for now
}

func (r *reviewRepository) HasReviewed(ctx context.Context, solutionID uint64, reviewerAddress string) (bool, error) {
	query := `SELECT COUNT(*) FROM solution_reviews WHERE solution_id = $1 AND reviewer_address = $2`

	var count int
	err := r.db.QueryRowContext(ctx, query, solutionID, reviewerAddress).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check review status: %w", err)
	}

	return count > 0, nil
}

func (r *reviewRepository) Update(ctx context.Context, review *Review) error {
	query := `
		UPDATE solution_reviews 
		SET rating = $1, comment = $2, security_score = $3, performance_score = $4,
			usability_score = $5, documentation_score = $6, updated_at = $7
		WHERE id = $8`

	_, err := r.db.ExecContext(ctx, query,
		review.Rating,
		review.Comment,
		review.SecurityScore,
		review.PerformanceScore,
		review.UsabilityScore,
		review.DocumentationScore,
		review.UpdatedAt,
		review.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update review", zap.Error(err))
		return fmt.Errorf("failed to update review: %w", err)
	}

	return nil
}

func (r *reviewRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM solution_reviews WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("review not found: %d", id)
	}

	return nil
}

// installationRepository implements InstallationRepository
type installationRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewInstallationRepository creates a new installation repository
func NewInstallationRepository(db *sql.DB, logger *logger.Logger) InstallationRepository {
	return &installationRepository{
		db:     db,
		logger: logger,
	}
}

func (r *installationRepository) Create(ctx context.Context, installation *Installation) error {
	query := `
		INSERT INTO solution_installations (
			solution_id, installer_address, environment, version, status, installed_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		installation.SolutionID,
		installation.InstallerAddress,
		installation.Environment,
		installation.Version,
		int(installation.Status),
		installation.InstalledAt,
	).Scan(&installation.ID)

	if err != nil {
		r.logger.Error("Failed to create installation", zap.Error(err))
		return fmt.Errorf("failed to create installation: %w", err)
	}

	return nil
}

func (r *installationRepository) GetByID(ctx context.Context, id int64) (*Installation, error) {
	query := `
		SELECT id, solution_id, installer_address, environment, version, status, installed_at, last_used
		FROM solution_installations 
		WHERE id = $1`

	installation := &Installation{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&installation.ID,
		&installation.SolutionID,
		&installation.InstallerAddress,
		&installation.Environment,
		&installation.Version,
		&installation.Status,
		&installation.InstalledAt,
		&installation.LastUsed,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("installation not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get installation: %w", err)
	}

	return installation, nil
}

func (r *installationRepository) GetBySolutionID(ctx context.Context, solutionID uint64) ([]*Installation, error) {
	query := `
		SELECT id, solution_id, installer_address, environment, version, status, installed_at, last_used
		FROM solution_installations 
		WHERE solution_id = $1
		ORDER BY installed_at DESC`

	rows, err := r.db.QueryContext(ctx, query, solutionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get installations: %w", err)
	}
	defer rows.Close()

	var installations []*Installation
	for rows.Next() {
		installation := &Installation{}
		err := rows.Scan(
			&installation.ID,
			&installation.SolutionID,
			&installation.InstallerAddress,
			&installation.Environment,
			&installation.Version,
			&installation.Status,
			&installation.InstalledAt,
			&installation.LastUsed,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan installation: %w", err)
		}
		installations = append(installations, installation)
	}

	return installations, nil
}

func (r *installationRepository) GetByInstaller(ctx context.Context, installerAddress string, limit, offset int) ([]*Installation, error) {
	// Implementation similar to GetBySolutionID but with installer filter
	return []*Installation{}, nil // Simplified for now
}

func (r *installationRepository) Update(ctx context.Context, installation *Installation) error {
	query := `
		UPDATE solution_installations 
		SET status = $1, last_used = $2
		WHERE id = $3`

	_, err := r.db.ExecContext(ctx, query,
		int(installation.Status),
		installation.LastUsed,
		installation.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update installation", zap.Error(err))
		return fmt.Errorf("failed to update installation: %w", err)
	}

	return nil
}

func (r *installationRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM solution_installations WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete installation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("installation not found: %d", id)
	}

	return nil
}

// categoryRepository implements CategoryRepository
type categoryRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *sql.DB, logger *logger.Logger) CategoryRepository {
	return &categoryRepository{
		db:     db,
		logger: logger,
	}
}

func (r *categoryRepository) GetAll(ctx context.Context) ([]*Category, error) {
	// Mock categories for now
	categories := []*Category{
		{ID: 1, Category: SolutionCategoryDeFi, Name: "DeFi", Description: "Decentralized Finance solutions", SolutionCount: 15},
		{ID: 2, Category: SolutionCategoryNFT, Name: "NFT", Description: "Non-Fungible Token solutions", SolutionCount: 8},
		{ID: 3, Category: SolutionCategoryDAO, Name: "DAO", Description: "Decentralized Autonomous Organization tools", SolutionCount: 12},
		{ID: 4, Category: SolutionCategoryAnalytics, Name: "Analytics", Description: "Data analytics and reporting", SolutionCount: 6},
		{ID: 5, Category: SolutionCategoryInfrastructure, Name: "Infrastructure", Description: "Platform infrastructure components", SolutionCount: 10},
		{ID: 6, Category: SolutionCategorySecurity, Name: "Security", Description: "Security and audit tools", SolutionCount: 4},
		{ID: 7, Category: SolutionCategoryUI, Name: "UI/UX", Description: "User interface components", SolutionCount: 9},
		{ID: 8, Category: SolutionCategoryIntegration, Name: "Integration", Description: "Third-party integrations", SolutionCount: 7},
	}

	return categories, nil
}

func (r *categoryRepository) GetByCategory(ctx context.Context, category SolutionCategory) (*Category, error) {
	// Mock category lookup
	categories, _ := r.GetAll(ctx)
	for _, cat := range categories {
		if cat.Category == category {
			return cat, nil
		}
	}
	return nil, fmt.Errorf("category not found: %d", int(category))
}

func (r *categoryRepository) UpdateSolutionCount(ctx context.Context, category SolutionCategory, count int) error {
	// Implementation would update solution count for category
	r.logger.Debug("Updated solution count for category",
		zap.String("category", category.String()),
		zap.Int("count", count))
	return nil
}
