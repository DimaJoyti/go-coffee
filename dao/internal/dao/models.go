package dao

import (
	"time"

	"github.com/shopspring/decimal"
)

// ProposalCategory represents the category of a proposal
type ProposalCategory int

const (
	ProposalCategoryGeneral ProposalCategory = iota
	ProposalCategoryBounty
	ProposalCategoryTreasury
	ProposalCategoryTechnical
	ProposalCategoryPartnership
)

func (pc ProposalCategory) String() string {
	switch pc {
	case ProposalCategoryGeneral:
		return "GENERAL"
	case ProposalCategoryBounty:
		return "BOUNTY"
	case ProposalCategoryTreasury:
		return "TREASURY"
	case ProposalCategoryTechnical:
		return "TECHNICAL"
	case ProposalCategoryPartnership:
		return "PARTNERSHIP"
	default:
		return "UNKNOWN"
	}
}

// ProposalStatus represents the status of a proposal
type ProposalStatus int

const (
	ProposalStatusPending ProposalStatus = iota
	ProposalStatusActive
	ProposalStatusCanceled
	ProposalStatusDefeated
	ProposalStatusSucceeded
	ProposalStatusQueued
	ProposalStatusExpired
	ProposalStatusExecuted
)

func (ps ProposalStatus) String() string {
	switch ps {
	case ProposalStatusPending:
		return "PENDING"
	case ProposalStatusActive:
		return "ACTIVE"
	case ProposalStatusCanceled:
		return "CANCELED"
	case ProposalStatusDefeated:
		return "DEFEATED"
	case ProposalStatusSucceeded:
		return "SUCCEEDED"
	case ProposalStatusQueued:
		return "QUEUED"
	case ProposalStatusExpired:
		return "EXPIRED"
	case ProposalStatusExecuted:
		return "EXECUTED"
	default:
		return "UNKNOWN"
	}
}

// VoteSupport represents the support type for a vote
type VoteSupport int

const (
	VoteSupportAgainst VoteSupport = iota
	VoteSupportFor
	VoteSupportAbstain
)

func (vs VoteSupport) String() string {
	switch vs {
	case VoteSupportAgainst:
		return "AGAINST"
	case VoteSupportFor:
		return "FOR"
	case VoteSupportAbstain:
		return "ABSTAIN"
	default:
		return "UNKNOWN"
	}
}

// Proposal represents a DAO governance proposal
type Proposal struct {
	ID                int64            `json:"id" db:"id"`
	ProposalID        string           `json:"proposal_id" db:"proposal_id"`
	Title             string           `json:"title" db:"title"`
	Description       string           `json:"description" db:"description"`
	Category          ProposalCategory `json:"category" db:"category"`
	ProposerAddress   string           `json:"proposer_address" db:"proposer_address"`
	Status            ProposalStatus   `json:"status" db:"status"`
	VotesFor          decimal.Decimal  `json:"votes_for" db:"votes_for"`
	VotesAgainst      decimal.Decimal  `json:"votes_against" db:"votes_against"`
	VotesAbstain      decimal.Decimal  `json:"votes_abstain" db:"votes_abstain"`
	QuorumReached     bool             `json:"quorum_reached" db:"quorum_reached"`
	ExecutionDeadline *time.Time       `json:"execution_deadline" db:"execution_deadline"`
	CreatedAt         time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at" db:"updated_at"`
	ExecutedAt        *time.Time       `json:"executed_at" db:"executed_at"`
	TransactionHash   string           `json:"transaction_hash" db:"transaction_hash"`
	BlockNumber       *int64           `json:"block_number" db:"block_number"`
	GasUsed           *int64           `json:"gas_used" db:"gas_used"`
}

// Vote represents a vote on a proposal
type Vote struct {
	ID              int64           `json:"id" db:"id"`
	ProposalID      string          `json:"proposal_id" db:"proposal_id"`
	VoterAddress    string          `json:"voter_address" db:"voter_address"`
	Support         VoteSupport     `json:"support" db:"support"`
	VotingPower     decimal.Decimal `json:"voting_power" db:"voting_power"`
	Reason          string          `json:"reason" db:"reason"`
	VotedAt         time.Time       `json:"voted_at" db:"voted_at"`
	TransactionHash string          `json:"transaction_hash" db:"transaction_hash"`
	BlockNumber     *int64          `json:"block_number" db:"block_number"`
}

// DeveloperProfile represents a developer's profile in the DAO
type DeveloperProfile struct {
	ID                     int64           `json:"id" db:"id"`
	WalletAddress          string          `json:"wallet_address" db:"wallet_address"`
	Username               string          `json:"username" db:"username"`
	Email                  string          `json:"email" db:"email"`
	GithubUsername         string          `json:"github_username" db:"github_username"`
	DiscordUsername        string          `json:"discord_username" db:"discord_username"`
	Bio                    string          `json:"bio" db:"bio"`
	Skills                 []string        `json:"skills" db:"skills"`
	ReputationScore        int             `json:"reputation_score" db:"reputation_score"`
	TotalBountiesCompleted int             `json:"total_bounties_completed" db:"total_bounties_completed"`
	TotalEarnings          decimal.Decimal `json:"total_earnings" db:"total_earnings"`
	TVLContributed         decimal.Decimal `json:"tvl_contributed" db:"tvl_contributed"`
	MAUContributed         int             `json:"mau_contributed" db:"mau_contributed"`
	IsVerified             bool            `json:"is_verified" db:"is_verified"`
	IsActive               bool            `json:"is_active" db:"is_active"`
	CreatedAt              time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at" db:"updated_at"`
	LastActivity           time.Time       `json:"last_activity" db:"last_activity"`
}

// GovernanceStats represents governance statistics
type GovernanceStats struct {
	TotalProposals     int             `json:"total_proposals"`
	ActiveProposals    int             `json:"active_proposals"`
	ExecutedProposals  int             `json:"executed_proposals"`
	TotalVotes         int             `json:"total_votes"`
	TotalVoters        int             `json:"total_voters"`
	TotalSupply        decimal.Decimal `json:"total_supply"`
	QuorumPercentage   int             `json:"quorum_percentage"`
	AverageVotingPower decimal.Decimal `json:"average_voting_power"`
	ParticipationRate  float64         `json:"participation_rate"`
}

// Request/Response types for API

// CreateProposalRequest represents a request to create a new proposal
type CreateProposalRequest struct {
	Title             string           `json:"title" binding:"required"`
	Description       string           `json:"description" binding:"required"`
	Category          ProposalCategory `json:"category" binding:"required"`
	ProposerAddress   string           `json:"proposer_address" binding:"required"`
	Targets           []string         `json:"targets"`
	Values            []string         `json:"values"`
	Calldatas         []string         `json:"calldatas"`
	ExecutionDeadline *time.Time       `json:"execution_deadline"`
}

// CreateProposalResponse represents the response after creating a proposal
type CreateProposalResponse struct {
	ProposalID      string         `json:"proposal_id"`
	TransactionHash string         `json:"transaction_hash"`
	Status          ProposalStatus `json:"status"`
}

// GetProposalsRequest represents a request to get proposals
type GetProposalsRequest struct {
	Status          *ProposalStatus   `json:"status"`
	Category        *ProposalCategory `json:"category"`
	ProposerAddress string            `json:"proposer_address"`
	Limit           int               `json:"limit"`
	Offset          int               `json:"offset"`
}

// GetProposalsResponse represents the response with proposals list
type GetProposalsResponse struct {
	Proposals []*Proposal `json:"proposals"`
	Total     int         `json:"total"`
	Limit     int         `json:"limit"`
	Offset    int         `json:"offset"`
}

// VoteRequest represents a request to vote on a proposal
type VoteRequest struct {
	ProposalID   string      `json:"proposal_id" binding:"required"`
	VoterAddress string      `json:"voter_address" binding:"required"`
	Support      VoteSupport `json:"support" binding:"required"`
	Reason       string      `json:"reason"`
}

// VoteResponse represents the response after voting
type VoteResponse struct {
	TransactionHash string          `json:"transaction_hash"`
	VotingPower     decimal.Decimal `json:"voting_power"`
}

// ProposalFilter represents filters for querying proposals
type ProposalFilter struct {
	Status   *ProposalStatus
	Category *ProposalCategory
	Proposer string
	Limit    int
	Offset   int
}

// GovernorProposalRequest represents a request to create a proposal on the Governor contract
type GovernorProposalRequest struct {
	Targets     []string
	Values      []string
	Calldatas   []string
	Description string
	Category    ProposalCategory
	Title       string
	Deadline    *time.Time
}

// GovernorVoteRequest represents a request to vote on the Governor contract
type GovernorVoteRequest struct {
	ProposalID string
	Support    VoteSupport
	Reason     string
}
