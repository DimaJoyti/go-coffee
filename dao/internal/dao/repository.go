package dao

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/logger"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

// ProposalRepository handles proposal data operations
type ProposalRepository interface {
	Create(ctx context.Context, proposal *Proposal) error
	GetByID(ctx context.Context, proposalID string) (*Proposal, error)
	List(ctx context.Context, filter *ProposalFilter) ([]*Proposal, int, error)
	Update(ctx context.Context, proposal *Proposal) error
	Delete(ctx context.Context, proposalID string) error
	GetStats(ctx context.Context) (*GovernanceStats, error)
}

// DeveloperRepository handles developer profile operations
type DeveloperRepository interface {
	Create(ctx context.Context, profile *DeveloperProfile) error
	GetByAddress(ctx context.Context, address string) (*DeveloperProfile, error)
	GetActive(ctx context.Context, limit, offset int) ([]*DeveloperProfile, error)
	Update(ctx context.Context, profile *DeveloperProfile) error
	UpdateActivity(ctx context.Context, address string) error
	GetTopByReputation(ctx context.Context, limit int) ([]*DeveloperProfile, error)
}

// VoteRepository handles vote data operations
type VoteRepository interface {
	Create(ctx context.Context, vote *Vote) error
	GetByProposal(ctx context.Context, proposalID string, limit, offset int) ([]*Vote, error)
	GetByVoter(ctx context.Context, voterAddress string, limit, offset int) ([]*Vote, error)
	GetVoteCount(ctx context.Context, proposalID string) (int, error)
}

// proposalRepository implements ProposalRepository
type proposalRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewProposalRepository creates a new proposal repository
func NewProposalRepository(db *sql.DB, logger *logger.Logger) ProposalRepository {
	return &proposalRepository{
		db:     db,
		logger: logger,
	}
}

func (r *proposalRepository) Create(ctx context.Context, proposal *Proposal) error {
	query := `
		INSERT INTO dao_proposals (
			proposal_id, title, description, category, proposer_address, 
			status, execution_deadline, transaction_hash, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		proposal.ProposalID,
		proposal.Title,
		proposal.Description,
		int(proposal.Category),
		proposal.ProposerAddress,
		int(proposal.Status),
		proposal.ExecutionDeadline,
		proposal.TransactionHash,
		proposal.CreatedAt,
		proposal.UpdatedAt,
	).Scan(&proposal.ID)

	if err != nil {
		r.logger.Error("Failed to create proposal", zap.Error(err))
		return fmt.Errorf("failed to create proposal: %w", err)
	}

	r.logger.Info("Proposal created", zap.String("proposalID", proposal.ProposalID))
	return nil
}

func (r *proposalRepository) GetByID(ctx context.Context, proposalID string) (*Proposal, error) {
	query := `
		SELECT id, proposal_id, title, description, category, proposer_address,
			   status, votes_for, votes_against, votes_abstain, quorum_reached,
			   execution_deadline, created_at, updated_at, executed_at,
			   transaction_hash, block_number, gas_used
		FROM dao_proposals 
		WHERE proposal_id = $1`

	proposal := &Proposal{}
	err := r.db.QueryRowContext(ctx, query, proposalID).Scan(
		&proposal.ID,
		&proposal.ProposalID,
		&proposal.Title,
		&proposal.Description,
		&proposal.Category,
		&proposal.ProposerAddress,
		&proposal.Status,
		&proposal.VotesFor,
		&proposal.VotesAgainst,
		&proposal.VotesAbstain,
		&proposal.QuorumReached,
		&proposal.ExecutionDeadline,
		&proposal.CreatedAt,
		&proposal.UpdatedAt,
		&proposal.ExecutedAt,
		&proposal.TransactionHash,
		&proposal.BlockNumber,
		&proposal.GasUsed,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("proposal not found: %s", proposalID)
		}
		r.logger.Error("Failed to get proposal", zap.Error(err))
		return nil, fmt.Errorf("failed to get proposal: %w", err)
	}

	return proposal, nil
}

func (r *proposalRepository) List(ctx context.Context, filter *ProposalFilter) ([]*Proposal, int, error) {
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

	if filter.Proposer != "" {
		conditions = append(conditions, fmt.Sprintf("proposer_address = $%d", argIndex))
		args = append(args, filter.Proposer)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM dao_proposals %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count proposals: %w", err)
	}

	// Get proposals with pagination
	query := fmt.Sprintf(`
		SELECT id, proposal_id, title, description, category, proposer_address,
			   status, votes_for, votes_against, votes_abstain, quorum_reached,
			   execution_deadline, created_at, updated_at, executed_at,
			   transaction_hash, block_number, gas_used
		FROM dao_proposals %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list proposals: %w", err)
	}
	defer rows.Close()

	var proposals []*Proposal
	for rows.Next() {
		proposal := &Proposal{}
		err := rows.Scan(
			&proposal.ID,
			&proposal.ProposalID,
			&proposal.Title,
			&proposal.Description,
			&proposal.Category,
			&proposal.ProposerAddress,
			&proposal.Status,
			&proposal.VotesFor,
			&proposal.VotesAgainst,
			&proposal.VotesAbstain,
			&proposal.QuorumReached,
			&proposal.ExecutionDeadline,
			&proposal.CreatedAt,
			&proposal.UpdatedAt,
			&proposal.ExecutedAt,
			&proposal.TransactionHash,
			&proposal.BlockNumber,
			&proposal.GasUsed,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan proposal: %w", err)
		}
		proposals = append(proposals, proposal)
	}

	return proposals, total, nil
}

func (r *proposalRepository) Update(ctx context.Context, proposal *Proposal) error {
	query := `
		UPDATE dao_proposals 
		SET status = $1, votes_for = $2, votes_against = $3, votes_abstain = $4,
			quorum_reached = $5, updated_at = $6, executed_at = $7,
			block_number = $8, gas_used = $9
		WHERE proposal_id = $10`

	proposal.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		int(proposal.Status),
		proposal.VotesFor,
		proposal.VotesAgainst,
		proposal.VotesAbstain,
		proposal.QuorumReached,
		proposal.UpdatedAt,
		proposal.ExecutedAt,
		proposal.BlockNumber,
		proposal.GasUsed,
		proposal.ProposalID,
	)

	if err != nil {
		r.logger.Error("Failed to update proposal", zap.Error(err))
		return fmt.Errorf("failed to update proposal: %w", err)
	}

	return nil
}

func (r *proposalRepository) Delete(ctx context.Context, proposalID string) error {
	query := `DELETE FROM dao_proposals WHERE proposal_id = $1`

	result, err := r.db.ExecContext(ctx, query, proposalID)
	if err != nil {
		return fmt.Errorf("failed to delete proposal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("proposal not found: %s", proposalID)
	}

	return nil
}

func (r *proposalRepository) GetStats(ctx context.Context) (*GovernanceStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_proposals,
			COUNT(CASE WHEN status = 1 THEN 1 END) as active_proposals,
			COUNT(CASE WHEN status = 7 THEN 1 END) as executed_proposals,
			COALESCE(
				(SELECT COUNT(*) FROM dao_votes), 0
			) as total_votes,
			COALESCE(
				(SELECT COUNT(DISTINCT voter_address) FROM dao_votes), 0
			) as total_voters
		FROM dao_proposals`

	stats := &GovernanceStats{}
	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalProposals,
		&stats.ActiveProposals,
		&stats.ExecutedProposals,
		&stats.TotalVotes,
		&stats.TotalVoters,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get governance stats: %w", err)
	}

	// Calculate participation rate
	if stats.TotalProposals > 0 {
		stats.ParticipationRate = float64(stats.TotalVotes) / float64(stats.TotalProposals)
	}

	return stats, nil
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
	query := `
		INSERT INTO developer_profiles (
			wallet_address, username, email, github_username, discord_username,
			bio, skills, reputation_score, is_verified, is_active,
			created_at, updated_at, last_activity
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		profile.WalletAddress,
		profile.Username,
		profile.Email,
		profile.GithubUsername,
		profile.DiscordUsername,
		profile.Bio,
		pq.Array(profile.Skills),
		profile.ReputationScore,
		profile.IsVerified,
		profile.IsActive,
		profile.CreatedAt,
		profile.UpdatedAt,
		profile.LastActivity,
	).Scan(&profile.ID)

	if err != nil {
		r.logger.Error("Failed to create developer profile", zap.Error(err))
		return fmt.Errorf("failed to create developer profile: %w", err)
	}

	return nil
}

func (r *developerRepository) GetByAddress(ctx context.Context, address string) (*DeveloperProfile, error) {
	query := `
		SELECT id, wallet_address, username, email, github_username, discord_username,
			   bio, skills, reputation_score, total_bounties_completed, total_earnings,
			   tvl_contributed, mau_contributed, is_verified, is_active,
			   created_at, updated_at, last_activity
		FROM developer_profiles 
		WHERE wallet_address = $1`

	profile := &DeveloperProfile{}
	err := r.db.QueryRowContext(ctx, query, address).Scan(
		&profile.ID,
		&profile.WalletAddress,
		&profile.Username,
		&profile.Email,
		&profile.GithubUsername,
		&profile.DiscordUsername,
		&profile.Bio,
		pq.Array(&profile.Skills),
		&profile.ReputationScore,
		&profile.TotalBountiesCompleted,
		&profile.TotalEarnings,
		&profile.TVLContributed,
		&profile.MAUContributed,
		&profile.IsVerified,
		&profile.IsActive,
		&profile.CreatedAt,
		&profile.UpdatedAt,
		&profile.LastActivity,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("developer profile not found: %s", address)
		}
		return nil, fmt.Errorf("failed to get developer profile: %w", err)
	}

	return profile, nil
}

func (r *developerRepository) GetActive(ctx context.Context, limit, offset int) ([]*DeveloperProfile, error) {
	query := `
		SELECT id, wallet_address, username, email, github_username, discord_username,
			   bio, skills, reputation_score, total_bounties_completed, total_earnings,
			   tvl_contributed, mau_contributed, is_verified, is_active,
			   created_at, updated_at, last_activity
		FROM developer_profiles 
		WHERE is_active = true
		ORDER BY reputation_score DESC, last_activity DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get active developers: %w", err)
	}
	defer rows.Close()

	var profiles []*DeveloperProfile
	for rows.Next() {
		profile := &DeveloperProfile{}
		err := rows.Scan(
			&profile.ID,
			&profile.WalletAddress,
			&profile.Username,
			&profile.Email,
			&profile.GithubUsername,
			&profile.DiscordUsername,
			&profile.Bio,
			pq.Array(&profile.Skills),
			&profile.ReputationScore,
			&profile.TotalBountiesCompleted,
			&profile.TotalEarnings,
			&profile.TVLContributed,
			&profile.MAUContributed,
			&profile.IsVerified,
			&profile.IsActive,
			&profile.CreatedAt,
			&profile.UpdatedAt,
			&profile.LastActivity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan developer profile: %w", err)
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

func (r *developerRepository) Update(ctx context.Context, profile *DeveloperProfile) error {
	query := `
		UPDATE developer_profiles 
		SET username = $1, email = $2, github_username = $3, discord_username = $4,
			bio = $5, skills = $6, reputation_score = $7, total_bounties_completed = $8,
			total_earnings = $9, tvl_contributed = $10, mau_contributed = $11,
			is_verified = $12, is_active = $13, updated_at = $14, last_activity = $15
		WHERE wallet_address = $16`

	profile.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		profile.Username,
		profile.Email,
		profile.GithubUsername,
		profile.DiscordUsername,
		profile.Bio,
		pq.Array(profile.Skills),
		profile.ReputationScore,
		profile.TotalBountiesCompleted,
		profile.TotalEarnings,
		profile.TVLContributed,
		profile.MAUContributed,
		profile.IsVerified,
		profile.IsActive,
		profile.UpdatedAt,
		profile.LastActivity,
		profile.WalletAddress,
	)

	if err != nil {
		return fmt.Errorf("failed to update developer profile: %w", err)
	}

	return nil
}

func (r *developerRepository) UpdateActivity(ctx context.Context, address string) error {
	query := `UPDATE developer_profiles SET last_activity = $1 WHERE wallet_address = $2`

	_, err := r.db.ExecContext(ctx, query, time.Now(), address)
	if err != nil {
		return fmt.Errorf("failed to update developer activity: %w", err)
	}

	return nil
}

func (r *developerRepository) GetTopByReputation(ctx context.Context, limit int) ([]*DeveloperProfile, error) {
	query := `
		SELECT id, wallet_address, username, email, github_username, discord_username,
			   bio, skills, reputation_score, total_bounties_completed, total_earnings,
			   tvl_contributed, mau_contributed, is_verified, is_active,
			   created_at, updated_at, last_activity
		FROM developer_profiles 
		WHERE is_active = true
		ORDER BY reputation_score DESC
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top developers: %w", err)
	}
	defer rows.Close()

	var profiles []*DeveloperProfile
	for rows.Next() {
		profile := &DeveloperProfile{}
		err := rows.Scan(
			&profile.ID,
			&profile.WalletAddress,
			&profile.Username,
			&profile.Email,
			&profile.GithubUsername,
			&profile.DiscordUsername,
			&profile.Bio,
			pq.Array(&profile.Skills),
			&profile.ReputationScore,
			&profile.TotalBountiesCompleted,
			&profile.TotalEarnings,
			&profile.TVLContributed,
			&profile.MAUContributed,
			&profile.IsVerified,
			&profile.IsActive,
			&profile.CreatedAt,
			&profile.UpdatedAt,
			&profile.LastActivity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan developer profile: %w", err)
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// voteRepository implements VoteRepository
type voteRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewVoteRepository creates a new vote repository
func NewVoteRepository(db *sql.DB, logger *logger.Logger) VoteRepository {
	return &voteRepository{
		db:     db,
		logger: logger,
	}
}

func (r *voteRepository) Create(ctx context.Context, vote *Vote) error {
	query := `
		INSERT INTO dao_votes (
			proposal_id, voter_address, support, voting_power, reason,
			voted_at, transaction_hash, block_number
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		vote.ProposalID,
		vote.VoterAddress,
		int(vote.Support),
		vote.VotingPower,
		vote.Reason,
		vote.VotedAt,
		vote.TransactionHash,
		vote.BlockNumber,
	).Scan(&vote.ID)

	if err != nil {
		r.logger.Error("Failed to create vote", zap.Error(err))
		return fmt.Errorf("failed to create vote: %w", err)
	}

	return nil
}

func (r *voteRepository) GetByProposal(ctx context.Context, proposalID string, limit, offset int) ([]*Vote, error) {
	query := `
		SELECT id, proposal_id, voter_address, support, voting_power, reason,
			   voted_at, transaction_hash, block_number
		FROM dao_votes 
		WHERE proposal_id = $1
		ORDER BY voted_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, proposalID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get votes by proposal: %w", err)
	}
	defer rows.Close()

	var votes []*Vote
	for rows.Next() {
		vote := &Vote{}
		err := rows.Scan(
			&vote.ID,
			&vote.ProposalID,
			&vote.VoterAddress,
			&vote.Support,
			&vote.VotingPower,
			&vote.Reason,
			&vote.VotedAt,
			&vote.TransactionHash,
			&vote.BlockNumber,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vote: %w", err)
		}
		votes = append(votes, vote)
	}

	return votes, nil
}

func (r *voteRepository) GetByVoter(ctx context.Context, voterAddress string, limit, offset int) ([]*Vote, error) {
	query := `
		SELECT id, proposal_id, voter_address, support, voting_power, reason,
			   voted_at, transaction_hash, block_number
		FROM dao_votes 
		WHERE voter_address = $1
		ORDER BY voted_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, voterAddress, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get votes by voter: %w", err)
	}
	defer rows.Close()

	var votes []*Vote
	for rows.Next() {
		vote := &Vote{}
		err := rows.Scan(
			&vote.ID,
			&vote.ProposalID,
			&vote.VoterAddress,
			&vote.Support,
			&vote.VotingPower,
			&vote.Reason,
			&vote.VotedAt,
			&vote.TransactionHash,
			&vote.BlockNumber,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vote: %w", err)
		}
		votes = append(votes, vote)
	}

	return votes, nil
}

func (r *voteRepository) GetVoteCount(ctx context.Context, proposalID string) (int, error) {
	query := `SELECT COUNT(*) FROM dao_votes WHERE proposal_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, proposalID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get vote count: %w", err)
	}

	return count, nil
}
