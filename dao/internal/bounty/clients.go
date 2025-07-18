package bounty

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// BountyManagerClientInterface defines the interface for bounty manager client
type BountyManagerClientInterface interface {
	CreateBounty(ctx context.Context, req *BountyRequest) (uint64, string, error)
	AssignBounty(ctx context.Context, bountyID uint64, developer string) (string, error)
	CompleteMilestone(ctx context.Context, bountyID uint64, milestoneIndex uint64) (string, error)
	VerifyPerformance(ctx context.Context, bountyID uint64, tvlImpact, mauImpact int64) (string, error)
}

// BountyManagerClient handles interactions with the BountyManager contract
type BountyManagerClient struct {
	client   *ethclient.Client
	contract common.Address
	logger   *logger.Logger
}

// NewBountyManagerClient creates a new bounty manager client
func NewBountyManagerClient(client *ethclient.Client, contractAddress string, logger *logger.Logger) (BountyManagerClientInterface, error) {
	if !common.IsHexAddress(contractAddress) {
		return nil, fmt.Errorf("invalid contract address: %s", contractAddress)
	}

	return &BountyManagerClient{
		client:   client,
		contract: common.HexToAddress(contractAddress),
		logger:   logger,
	}, nil
}

// CreateBounty creates a new bounty on-chain
func (bmc *BountyManagerClient) CreateBounty(ctx context.Context, req *BountyRequest) (uint64, string, error) {
	bmc.logger.Info("Creating bounty on-chain",
		zap.String("title", req.Title),
		zap.String("category", req.Category.String()))

	// Mock bounty ID and transaction hash - in real implementation, this would interact with the contract
	bountyID := uint64(1)
	txHash := "0xbounty1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	bmc.logger.Info("Bounty created on-chain",
		zap.Uint64("bountyID", bountyID),
		zap.String("txHash", txHash))

	return bountyID, txHash, nil
}

// AssignBounty assigns a bounty to a developer
func (bmc *BountyManagerClient) AssignBounty(ctx context.Context, bountyID uint64, developer string) (string, error) {
	bmc.logger.Info("Assigning bounty on-chain",
		zap.Uint64("bountyID", bountyID),
		zap.String("developer", developer))

	// Mock transaction hash
	txHash := "0xassign1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	return txHash, nil
}

// CompleteMilestone completes a bounty milestone
func (bmc *BountyManagerClient) CompleteMilestone(ctx context.Context, bountyID uint64, milestoneIndex uint64) (string, error) {
	bmc.logger.Info("Completing milestone on-chain",
		zap.Uint64("bountyID", bountyID),
		zap.Uint64("milestoneIndex", milestoneIndex))

	// Mock transaction hash
	txHash := "0xmilestone1234567890abcdef1234567890abcdef1234567890abcdef123456789"

	return txHash, nil
}

// VerifyPerformance verifies bounty performance metrics
func (bmc *BountyManagerClient) VerifyPerformance(ctx context.Context, bountyID uint64, tvlImpact, mauImpact int64) (string, error) {
	bmc.logger.Info("Verifying performance on-chain",
		zap.Uint64("bountyID", bountyID),
		zap.Int64("tvlImpact", tvlImpact),
		zap.Int64("mauImpact", mauImpact))

	// Mock transaction hash
	txHash := "0xperformance1234567890abcdef1234567890abcdef1234567890abcdef12345"

	return txHash, nil
}

// Additional repository implementations

// milestoneRepository implements MilestoneRepository
type milestoneRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewMilestoneRepository creates a new milestone repository
func NewMilestoneRepository(db *sql.DB, logger *logger.Logger) MilestoneRepository {
	return &milestoneRepository{
		db:     db,
		logger: logger,
	}
}

func (r *milestoneRepository) Create(ctx context.Context, milestone *Milestone) error {
	query := `
		INSERT INTO bounty_milestones (
			bounty_id, milestone_index, description, reward, deadline, completed, paid
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		milestone.BountyID,
		milestone.Index,
		milestone.Description,
		milestone.Reward,
		milestone.Deadline,
		milestone.Completed,
		milestone.Paid,
	).Scan(&milestone.ID)

	if err != nil {
		r.logger.Error("Failed to create milestone", zap.Error(err))
		return fmt.Errorf("failed to create milestone: %w", err)
	}

	return nil
}

func (r *milestoneRepository) GetByID(ctx context.Context, id int64) (*Milestone, error) {
	query := `
		SELECT id, bounty_id, milestone_index, description, reward, deadline,
			   completed, paid, completed_at, transaction_hash
		FROM bounty_milestones 
		WHERE id = $1`

	milestone := &Milestone{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&milestone.ID,
		&milestone.BountyID,
		&milestone.Index,
		&milestone.Description,
		&milestone.Reward,
		&milestone.Deadline,
		&milestone.Completed,
		&milestone.Paid,
		&milestone.CompletedAt,
		&milestone.TransactionHash,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("milestone not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get milestone: %w", err)
	}

	return milestone, nil
}

func (r *milestoneRepository) GetByBountyID(ctx context.Context, bountyID uint64) ([]*Milestone, error) {
	query := `
		SELECT id, bounty_id, milestone_index, description, reward, deadline,
			   completed, paid, completed_at, transaction_hash
		FROM bounty_milestones 
		WHERE bounty_id = $1
		ORDER BY milestone_index`

	rows, err := r.db.QueryContext(ctx, query, bountyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestones: %w", err)
	}
	defer rows.Close()

	var milestones []*Milestone
	for rows.Next() {
		milestone := &Milestone{}
		err := rows.Scan(
			&milestone.ID,
			&milestone.BountyID,
			&milestone.Index,
			&milestone.Description,
			&milestone.Reward,
			&milestone.Deadline,
			&milestone.Completed,
			&milestone.Paid,
			&milestone.CompletedAt,
			&milestone.TransactionHash,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan milestone: %w", err)
		}
		milestones = append(milestones, milestone)
	}

	return milestones, nil
}

func (r *milestoneRepository) GetByBountyAndIndex(ctx context.Context, bountyID uint64, index uint64) (*Milestone, error) {
	query := `
		SELECT id, bounty_id, milestone_index, description, reward, deadline,
			   completed, paid, completed_at, transaction_hash
		FROM bounty_milestones 
		WHERE bounty_id = $1 AND milestone_index = $2`

	milestone := &Milestone{}
	err := r.db.QueryRowContext(ctx, query, bountyID, index).Scan(
		&milestone.ID,
		&milestone.BountyID,
		&milestone.Index,
		&milestone.Description,
		&milestone.Reward,
		&milestone.Deadline,
		&milestone.Completed,
		&milestone.Paid,
		&milestone.CompletedAt,
		&milestone.TransactionHash,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("milestone not found: bounty %d, index %d", bountyID, index)
		}
		return nil, fmt.Errorf("failed to get milestone: %w", err)
	}

	return milestone, nil
}

func (r *milestoneRepository) Update(ctx context.Context, milestone *Milestone) error {
	query := `
		UPDATE bounty_milestones 
		SET completed = $1, paid = $2, completed_at = $3, transaction_hash = $4
		WHERE id = $5`

	_, err := r.db.ExecContext(ctx, query,
		milestone.Completed,
		milestone.Paid,
		milestone.CompletedAt,
		milestone.TransactionHash,
		milestone.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update milestone", zap.Error(err))
		return fmt.Errorf("failed to update milestone: %w", err)
	}

	return nil
}

func (r *milestoneRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM bounty_milestones WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete milestone: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("milestone not found: %d", id)
	}

	return nil
}

// applicationRepository implements ApplicationRepository
type applicationRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewApplicationRepository creates a new application repository
func NewApplicationRepository(db *sql.DB, logger *logger.Logger) ApplicationRepository {
	return &applicationRepository{
		db:     db,
		logger: logger,
	}
}

func (r *applicationRepository) Create(ctx context.Context, application *Application) error {
	query := `
		INSERT INTO bounty_applications (
			bounty_id, applicant_address, application_message, proposed_timeline,
			applied_at, status
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		application.BountyID,
		application.ApplicantAddress,
		application.ApplicationMessage,
		application.ProposedTimeline,
		application.AppliedAt,
		int(application.Status),
	).Scan(&application.ID)

	if err != nil {
		r.logger.Error("Failed to create application", zap.Error(err))
		return fmt.Errorf("failed to create application: %w", err)
	}

	return nil
}

func (r *applicationRepository) GetByID(ctx context.Context, id int64) (*Application, error) {
	query := `
		SELECT id, bounty_id, applicant_address, application_message,
			   proposed_timeline, applied_at, status
		FROM bounty_applications 
		WHERE id = $1`

	application := &Application{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&application.ID,
		&application.BountyID,
		&application.ApplicantAddress,
		&application.ApplicationMessage,
		&application.ProposedTimeline,
		&application.AppliedAt,
		&application.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("application not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	return application, nil
}

func (r *applicationRepository) GetByBountyID(ctx context.Context, bountyID uint64) ([]*Application, error) {
	query := `
		SELECT id, bounty_id, applicant_address, application_message,
			   proposed_timeline, applied_at, status
		FROM bounty_applications 
		WHERE bounty_id = $1
		ORDER BY applied_at DESC`

	rows, err := r.db.QueryContext(ctx, query, bountyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get applications: %w", err)
	}
	defer rows.Close()

	var applications []*Application
	for rows.Next() {
		application := &Application{}
		err := rows.Scan(
			&application.ID,
			&application.BountyID,
			&application.ApplicantAddress,
			&application.ApplicationMessage,
			&application.ProposedTimeline,
			&application.AppliedAt,
			&application.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan application: %w", err)
		}
		applications = append(applications, application)
	}

	return applications, nil
}

func (r *applicationRepository) GetByApplicant(ctx context.Context, applicantAddress string, limit, offset int) ([]*Application, error) {
	// Implementation similar to GetByBountyID but with applicant filter
	return []*Application{}, nil // Simplified for now
}

func (r *applicationRepository) HasApplied(ctx context.Context, bountyID uint64, applicantAddress string) (bool, error) {
	query := `SELECT COUNT(*) FROM bounty_applications WHERE bounty_id = $1 AND applicant_address = $2`

	var count int
	err := r.db.QueryRowContext(ctx, query, bountyID, applicantAddress).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check application: %w", err)
	}

	return count > 0, nil
}

func (r *applicationRepository) UpdateStatus(ctx context.Context, bountyID uint64, applicantAddress string, status ApplicationStatus) error {
	query := `UPDATE bounty_applications SET status = $1 WHERE bounty_id = $2 AND applicant_address = $3`

	_, err := r.db.ExecContext(ctx, query, int(status), bountyID, applicantAddress)
	if err != nil {
		return fmt.Errorf("failed to update application status: %w", err)
	}

	return nil
}

// developerRepository implements DeveloperRepository
type developerRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewDeveloperRepository creates a new developer repository
func NewDeveloperRepository(db *sql.DB, logger *logger.Logger) DeveloperRepository {
	return &developerRepository{
		db:     db,
		logger: logger,
	}
}

func (r *developerRepository) Create(ctx context.Context, profile *DeveloperProfile) error {
	// Implementation would create developer profile
	return nil
}

func (r *developerRepository) GetByAddress(ctx context.Context, address string) (*DeveloperProfile, error) {
	// Implementation would get developer by address
	return &DeveloperProfile{}, nil
}

func (r *developerRepository) Update(ctx context.Context, profile *DeveloperProfile) error {
	// Implementation would update developer profile
	return nil
}

func (r *developerRepository) GetTopByReputation(ctx context.Context, limit int) ([]*DeveloperProfile, error) {
	// Implementation would get top developers by reputation
	return []*DeveloperProfile{}, nil
}

func (r *developerRepository) UpdateReputation(ctx context.Context, address string, points int) error {
	// Implementation would update developer reputation
	return nil
}
