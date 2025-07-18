package bounty

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/logger"
	"go.uber.org/zap"
)

// Repository interfaces

// BountyRepository handles bounty data operations
type BountyRepository interface {
	Create(ctx context.Context, bounty *Bounty) error
	GetByID(ctx context.Context, bountyID uint64) (*Bounty, error)
	List(ctx context.Context, filter *BountyFilter) ([]*Bounty, int, error)
	Update(ctx context.Context, bounty *Bounty) error
	Delete(ctx context.Context, bountyID uint64) error
	GetActive(ctx context.Context, limit, offset int) ([]*Bounty, error)
	GetCompleted(ctx context.Context, limit, offset int) ([]*Bounty, error)
	GetByCreator(ctx context.Context, creatorAddress string, limit, offset int) ([]*Bounty, error)
	GetByAssignee(ctx context.Context, assigneeAddress string, limit, offset int) ([]*Bounty, error)
}

// MilestoneRepository handles milestone data operations
type MilestoneRepository interface {
	Create(ctx context.Context, milestone *Milestone) error
	GetByID(ctx context.Context, id int64) (*Milestone, error)
	GetByBountyID(ctx context.Context, bountyID uint64) ([]*Milestone, error)
	GetByBountyAndIndex(ctx context.Context, bountyID uint64, index uint64) (*Milestone, error)
	Update(ctx context.Context, milestone *Milestone) error
	Delete(ctx context.Context, id int64) error
}

// ApplicationRepository handles application data operations
type ApplicationRepository interface {
	Create(ctx context.Context, application *Application) error
	GetByID(ctx context.Context, id int64) (*Application, error)
	GetByBountyID(ctx context.Context, bountyID uint64) ([]*Application, error)
	GetByApplicant(ctx context.Context, applicantAddress string, limit, offset int) ([]*Application, error)
	HasApplied(ctx context.Context, bountyID uint64, applicantAddress string) (bool, error)
	UpdateStatus(ctx context.Context, bountyID uint64, applicantAddress string, status ApplicationStatus) error
}

// DeveloperRepository handles developer profile operations
type DeveloperRepository interface {
	Create(ctx context.Context, profile *DeveloperProfile) error
	GetByAddress(ctx context.Context, address string) (*DeveloperProfile, error)
	Update(ctx context.Context, profile *DeveloperProfile) error
	GetTopByReputation(ctx context.Context, limit int) ([]*DeveloperProfile, error)
	UpdateReputation(ctx context.Context, address string, points int) error
}

// bountyRepository implements BountyRepository
type bountyRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewBountyRepository creates a new bounty repository
func NewBountyRepository(db *sql.DB, logger *logger.Logger) BountyRepository {
	return &bountyRepository{
		db:     db,
		logger: logger,
	}
}

func (r *bountyRepository) Create(ctx context.Context, bounty *Bounty) error {
	query := `
		INSERT INTO bounties (
			bounty_id, title, description, category, status, creator_address,
			total_reward, deadline, created_at, updated_at, transaction_hash
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		bounty.BountyID,
		bounty.Title,
		bounty.Description,
		int(bounty.Category),
		int(bounty.Status),
		bounty.CreatorAddress,
		bounty.TotalReward,
		bounty.Deadline,
		bounty.CreatedAt,
		bounty.UpdatedAt,
		bounty.TransactionHash,
	).Scan(&bounty.ID)

	if err != nil {
		r.logger.Error("Failed to create bounty", zap.Error(err))
		return fmt.Errorf("failed to create bounty: %w", err)
	}

	r.logger.Info("Bounty created", zap.Uint64("bountyID", bounty.BountyID))
	return nil
}

func (r *bountyRepository) GetByID(ctx context.Context, bountyID uint64) (*Bounty, error) {
	query := `
		SELECT id, bounty_id, title, description, category, status, creator_address,
			   assignee_address, total_reward, deadline, created_at, updated_at,
			   assigned_at, completed_at, transaction_hash, block_number,
			   tvl_impact, mau_impact, performance_verified
		FROM bounties 
		WHERE bounty_id = $1`

	bounty := &Bounty{}
	err := r.db.QueryRowContext(ctx, query, bountyID).Scan(
		&bounty.ID,
		&bounty.BountyID,
		&bounty.Title,
		&bounty.Description,
		&bounty.Category,
		&bounty.Status,
		&bounty.CreatorAddress,
		&bounty.AssigneeAddress,
		&bounty.TotalReward,
		&bounty.Deadline,
		&bounty.CreatedAt,
		&bounty.UpdatedAt,
		&bounty.AssignedAt,
		&bounty.CompletedAt,
		&bounty.TransactionHash,
		&bounty.BlockNumber,
		&bounty.TVLImpact,
		&bounty.MAUImpact,
		&bounty.PerformanceVerified,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("bounty not found: %d", bountyID)
		}
		r.logger.Error("Failed to get bounty", zap.Error(err))
		return nil, fmt.Errorf("failed to get bounty: %w", err)
	}

	return bounty, nil
}

func (r *bountyRepository) List(ctx context.Context, filter *BountyFilter) ([]*Bounty, int, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Build WHERE conditions
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, int(*filter.Status))
		argIndex++
	}

	if filter.Category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, int(*filter.Category))
		argIndex++
	}

	if filter.CreatorAddress != "" {
		conditions = append(conditions, fmt.Sprintf("creator_address = $%d", argIndex))
		args = append(args, filter.CreatorAddress)
		argIndex++
	}

	if filter.AssigneeAddress != "" {
		conditions = append(conditions, fmt.Sprintf("assignee_address = $%d", argIndex))
		args = append(args, filter.AssigneeAddress)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM bounties %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count bounties: %w", err)
	}

	// Get bounties with pagination
	query := fmt.Sprintf(`
		SELECT id, bounty_id, title, description, category, status, creator_address,
			   assignee_address, total_reward, deadline, created_at, updated_at,
			   assigned_at, completed_at, transaction_hash, block_number,
			   tvl_impact, mau_impact, performance_verified
		FROM bounties %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list bounties: %w", err)
	}
	defer rows.Close()

	var bounties []*Bounty
	for rows.Next() {
		bounty := &Bounty{}
		err := rows.Scan(
			&bounty.ID,
			&bounty.BountyID,
			&bounty.Title,
			&bounty.Description,
			&bounty.Category,
			&bounty.Status,
			&bounty.CreatorAddress,
			&bounty.AssigneeAddress,
			&bounty.TotalReward,
			&bounty.Deadline,
			&bounty.CreatedAt,
			&bounty.UpdatedAt,
			&bounty.AssignedAt,
			&bounty.CompletedAt,
			&bounty.TransactionHash,
			&bounty.BlockNumber,
			&bounty.TVLImpact,
			&bounty.MAUImpact,
			&bounty.PerformanceVerified,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan bounty: %w", err)
		}
		bounties = append(bounties, bounty)
	}

	return bounties, total, nil
}

func (r *bountyRepository) Update(ctx context.Context, bounty *Bounty) error {
	query := `
		UPDATE bounties 
		SET status = $1, assignee_address = $2, updated_at = $3, assigned_at = $4,
			completed_at = $5, block_number = $6, tvl_impact = $7, mau_impact = $8,
			performance_verified = $9
		WHERE bounty_id = $10`

	bounty.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		int(bounty.Status),
		bounty.AssigneeAddress,
		bounty.UpdatedAt,
		bounty.AssignedAt,
		bounty.CompletedAt,
		bounty.BlockNumber,
		bounty.TVLImpact,
		bounty.MAUImpact,
		bounty.PerformanceVerified,
		bounty.BountyID,
	)

	if err != nil {
		r.logger.Error("Failed to update bounty", zap.Error(err))
		return fmt.Errorf("failed to update bounty: %w", err)
	}

	return nil
}

func (r *bountyRepository) Delete(ctx context.Context, bountyID uint64) error {
	query := `DELETE FROM bounties WHERE bounty_id = $1`

	result, err := r.db.ExecContext(ctx, query, bountyID)
	if err != nil {
		return fmt.Errorf("failed to delete bounty: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("bounty not found: %d", bountyID)
	}

	return nil
}

func (r *bountyRepository) GetActive(ctx context.Context, limit, offset int) ([]*Bounty, error) {
	query := `
		SELECT id, bounty_id, title, description, category, status, creator_address,
			   assignee_address, total_reward, deadline, created_at, updated_at,
			   assigned_at, completed_at, transaction_hash, block_number,
			   tvl_impact, mau_impact, performance_verified
		FROM bounties 
		WHERE status IN (0, 1, 2, 3) -- OPEN, ASSIGNED, IN_PROGRESS, SUBMITTED
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get active bounties: %w", err)
	}
	defer rows.Close()

	var bounties []*Bounty
	for rows.Next() {
		bounty := &Bounty{}
		err := rows.Scan(
			&bounty.ID,
			&bounty.BountyID,
			&bounty.Title,
			&bounty.Description,
			&bounty.Category,
			&bounty.Status,
			&bounty.CreatorAddress,
			&bounty.AssigneeAddress,
			&bounty.TotalReward,
			&bounty.Deadline,
			&bounty.CreatedAt,
			&bounty.UpdatedAt,
			&bounty.AssignedAt,
			&bounty.CompletedAt,
			&bounty.TransactionHash,
			&bounty.BlockNumber,
			&bounty.TVLImpact,
			&bounty.MAUImpact,
			&bounty.PerformanceVerified,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bounty: %w", err)
		}
		bounties = append(bounties, bounty)
	}

	return bounties, nil
}

func (r *bountyRepository) GetCompleted(ctx context.Context, limit, offset int) ([]*Bounty, error) {
	query := `
		SELECT id, bounty_id, title, description, category, status, creator_address,
			   assignee_address, total_reward, deadline, created_at, updated_at,
			   assigned_at, completed_at, transaction_hash, block_number,
			   tvl_impact, mau_impact, performance_verified
		FROM bounties 
		WHERE status = 4 -- COMPLETED
		ORDER BY completed_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed bounties: %w", err)
	}
	defer rows.Close()

	var bounties []*Bounty
	for rows.Next() {
		bounty := &Bounty{}
		err := rows.Scan(
			&bounty.ID,
			&bounty.BountyID,
			&bounty.Title,
			&bounty.Description,
			&bounty.Category,
			&bounty.Status,
			&bounty.CreatorAddress,
			&bounty.AssigneeAddress,
			&bounty.TotalReward,
			&bounty.Deadline,
			&bounty.CreatedAt,
			&bounty.UpdatedAt,
			&bounty.AssignedAt,
			&bounty.CompletedAt,
			&bounty.TransactionHash,
			&bounty.BlockNumber,
			&bounty.TVLImpact,
			&bounty.MAUImpact,
			&bounty.PerformanceVerified,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bounty: %w", err)
		}
		bounties = append(bounties, bounty)
	}

	return bounties, nil
}

func (r *bountyRepository) GetByCreator(ctx context.Context, creatorAddress string, limit, offset int) ([]*Bounty, error) {
	// Implementation similar to GetActive but with creator filter
	return []*Bounty{}, nil // Simplified for now
}

func (r *bountyRepository) GetByAssignee(ctx context.Context, assigneeAddress string, limit, offset int) ([]*Bounty, error) {
	// Implementation similar to GetActive but with assignee filter
	return []*Bounty{}, nil // Simplified for now
}
