package bounty

import (
	"time"

	"github.com/shopspring/decimal"
)

// BountyCategory represents the category of a bounty
type BountyCategory int

const (
	BountyCategoryTVLGrowth BountyCategory = iota
	BountyCategoryMAUExpansion
	BountyCategoryInnovation
	BountyCategoryMaintenance
	BountyCategorySecurity
	BountyCategoryIntegration
)

func (bc BountyCategory) String() string {
	switch bc {
	case BountyCategoryTVLGrowth:
		return "TVL_GROWTH"
	case BountyCategoryMAUExpansion:
		return "MAU_EXPANSION"
	case BountyCategoryInnovation:
		return "INNOVATION"
	case BountyCategoryMaintenance:
		return "MAINTENANCE"
	case BountyCategorySecurity:
		return "SECURITY"
	case BountyCategoryIntegration:
		return "INTEGRATION"
	default:
		return "UNKNOWN"
	}
}

// BountyStatus represents the status of a bounty
type BountyStatus int

const (
	BountyStatusOpen BountyStatus = iota
	BountyStatusAssigned
	BountyStatusInProgress
	BountyStatusSubmitted
	BountyStatusCompleted
	BountyStatusCancelled
)

func (bs BountyStatus) String() string {
	switch bs {
	case BountyStatusOpen:
		return "OPEN"
	case BountyStatusAssigned:
		return "ASSIGNED"
	case BountyStatusInProgress:
		return "IN_PROGRESS"
	case BountyStatusSubmitted:
		return "SUBMITTED"
	case BountyStatusCompleted:
		return "COMPLETED"
	case BountyStatusCancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

// ApplicationStatus represents the status of a bounty application
type ApplicationStatus int

const (
	ApplicationStatusPending ApplicationStatus = iota
	ApplicationStatusAccepted
	ApplicationStatusRejected
)

func (as ApplicationStatus) String() string {
	switch as {
	case ApplicationStatusPending:
		return "PENDING"
	case ApplicationStatusAccepted:
		return "ACCEPTED"
	case ApplicationStatusRejected:
		return "REJECTED"
	default:
		return "UNKNOWN"
	}
}

// Bounty represents a bounty in the system
type Bounty struct {
	ID                     int64            `json:"id" db:"id"`
	BountyID               uint64           `json:"bounty_id" db:"bounty_id"`
	Title                  string           `json:"title" db:"title"`
	Description            string           `json:"description" db:"description"`
	Category               BountyCategory   `json:"category" db:"category"`
	Status                 BountyStatus     `json:"status" db:"status"`
	CreatorAddress         string           `json:"creator_address" db:"creator_address"`
	AssigneeAddress        *string          `json:"assignee_address" db:"assignee_address"`
	TotalReward            decimal.Decimal  `json:"total_reward" db:"total_reward"`
	Deadline               time.Time        `json:"deadline" db:"deadline"`
	CreatedAt              time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time        `json:"updated_at" db:"updated_at"`
	AssignedAt             *time.Time       `json:"assigned_at" db:"assigned_at"`
	CompletedAt            *time.Time       `json:"completed_at" db:"completed_at"`
	TransactionHash        string           `json:"transaction_hash" db:"transaction_hash"`
	BlockNumber            *int64           `json:"block_number" db:"block_number"`
	TVLImpact              decimal.Decimal  `json:"tvl_impact" db:"tvl_impact"`
	MAUImpact              int              `json:"mau_impact" db:"mau_impact"`
	PerformanceVerified    bool             `json:"performance_verified" db:"performance_verified"`
}

// Milestone represents a bounty milestone
type Milestone struct {
	ID          int64           `json:"id" db:"id"`
	BountyID    uint64          `json:"bounty_id" db:"bounty_id"`
	Index       uint64          `json:"milestone_index" db:"milestone_index"`
	Description string          `json:"description" db:"description"`
	Reward      decimal.Decimal `json:"reward" db:"reward"`
	Deadline    time.Time       `json:"deadline" db:"deadline"`
	Completed   bool            `json:"completed" db:"completed"`
	Paid        bool            `json:"paid" db:"paid"`
	CompletedAt *time.Time      `json:"completed_at" db:"completed_at"`
	TransactionHash string      `json:"transaction_hash" db:"transaction_hash"`
}

// Application represents a bounty application
type Application struct {
	ID                 int64             `json:"id" db:"id"`
	BountyID           uint64            `json:"bounty_id" db:"bounty_id"`
	ApplicantAddress   string            `json:"applicant_address" db:"applicant_address"`
	ApplicationMessage string            `json:"application_message" db:"application_message"`
	ProposedTimeline   int               `json:"proposed_timeline" db:"proposed_timeline"`
	AppliedAt          time.Time         `json:"applied_at" db:"applied_at"`
	Status             ApplicationStatus `json:"status" db:"status"`
}

// DeveloperProfile represents a developer's profile
type DeveloperProfile struct {
	ID                     int64           `json:"id" db:"id"`
	WalletAddress          string          `json:"wallet_address" db:"wallet_address"`
	Username               string          `json:"username" db:"username"`
	ReputationScore        int             `json:"reputation_score" db:"reputation_score"`
	TotalBountiesCompleted int             `json:"total_bounties_completed" db:"total_bounties_completed"`
	TotalEarnings          decimal.Decimal `json:"total_earnings" db:"total_earnings"`
	TVLContributed         decimal.Decimal `json:"tvl_contributed" db:"tvl_contributed"`
	MAUContributed         int             `json:"mau_contributed" db:"mau_contributed"`
	IsActive               bool            `json:"is_active" db:"is_active"`
	CreatedAt              time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at" db:"updated_at"`
	LastActivity           time.Time       `json:"last_activity" db:"last_activity"`
}

// PerformanceMetrics represents performance metrics for a bounty
type PerformanceMetrics struct {
	ID              int64           `json:"id" db:"id"`
	BountyID        uint64          `json:"bounty_id" db:"bounty_id"`
	TVLImpact       decimal.Decimal `json:"tvl_impact" db:"tvl_impact"`
	MAUImpact       int             `json:"mau_impact" db:"mau_impact"`
	RevenueGenerated decimal.Decimal `json:"revenue_generated" db:"revenue_generated"`
	TransactionCount int             `json:"transaction_count" db:"transaction_count"`
	UniqueUsers     int             `json:"unique_users" db:"unique_users"`
	MeasuredAt      time.Time       `json:"measured_at" db:"measured_at"`
}

// BountyDetails represents detailed bounty information
type BountyDetails struct {
	Bounty       *Bounty        `json:"bounty"`
	Milestones   []*Milestone   `json:"milestones"`
	Applications []*Application `json:"applications"`
}

// Request/Response types for API

// CreateBountyRequest represents a request to create a new bounty
type CreateBountyRequest struct {
	Title          string                    `json:"title" binding:"required"`
	Description    string                    `json:"description" binding:"required"`
	Category       BountyCategory            `json:"category" binding:"required"`
	CreatorAddress string                    `json:"creator_address" binding:"required"`
	TotalReward    decimal.Decimal           `json:"total_reward" binding:"required"`
	Deadline       time.Time                 `json:"deadline" binding:"required"`
	Milestones     []CreateMilestoneRequest  `json:"milestones" binding:"required"`
}

// CreateMilestoneRequest represents a milestone in a bounty creation request
type CreateMilestoneRequest struct {
	Description string          `json:"description" binding:"required"`
	Reward      decimal.Decimal `json:"reward" binding:"required"`
	Deadline    time.Time       `json:"deadline" binding:"required"`
}

// CreateBountyResponse represents the response after creating a bounty
type CreateBountyResponse struct {
	BountyID        uint64       `json:"bounty_id"`
	TransactionHash string       `json:"transaction_hash"`
	Status          BountyStatus `json:"status"`
}

// ApplyBountyRequest represents a request to apply for a bounty
type ApplyBountyRequest struct {
	BountyID          uint64 `json:"bounty_id" binding:"required"`
	ApplicantAddress  string `json:"applicant_address" binding:"required"`
	Message           string `json:"message"`
	ProposedTimeline  int    `json:"proposed_timeline"`
}

// ApplyBountyResponse represents the response after applying for a bounty
type ApplyBountyResponse struct {
	ApplicationID int64             `json:"application_id"`
	Status        ApplicationStatus `json:"status"`
}

// AssignBountyRequest represents a request to assign a bounty
type AssignBountyRequest struct {
	BountyID        uint64 `json:"bounty_id" binding:"required"`
	AssigneeAddress string `json:"assignee_address" binding:"required"`
}

// AssignBountyResponse represents the response after assigning a bounty
type AssignBountyResponse struct {
	TransactionHash string       `json:"transaction_hash"`
	Status          BountyStatus `json:"status"`
}

// CompleteMilestoneRequest represents a request to complete a milestone
type CompleteMilestoneRequest struct {
	BountyID       uint64 `json:"bounty_id" binding:"required"`
	MilestoneIndex uint64 `json:"milestone_index" binding:"required"`
}

// CompleteMilestoneResponse represents the response after completing a milestone
type CompleteMilestoneResponse struct {
	TransactionHash string          `json:"transaction_hash"`
	Reward          decimal.Decimal `json:"reward"`
}

// GetBountiesRequest represents a request to get bounties
type GetBountiesRequest struct {
	Status          *BountyStatus   `json:"status"`
	Category        *BountyCategory `json:"category"`
	CreatorAddress  string          `json:"creator_address"`
	AssigneeAddress string          `json:"assignee_address"`
	Limit           int             `json:"limit"`
	Offset          int             `json:"offset"`
}

// GetBountiesResponse represents the response with bounties list
type GetBountiesResponse struct {
	Bounties []*Bounty `json:"bounties"`
	Total    int       `json:"total"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}

// VerifyPerformanceRequest represents a request to verify bounty performance
type VerifyPerformanceRequest struct {
	BountyID         uint64          `json:"bounty_id" binding:"required"`
	TVLImpact        decimal.Decimal `json:"tvl_impact"`
	MAUImpact        int             `json:"mau_impact"`
	RevenueGenerated decimal.Decimal `json:"revenue_generated"`
	TransactionCount int             `json:"transaction_count"`
	UniqueUsers      int             `json:"unique_users"`
}

// VerifyPerformanceResponse represents the response after verifying performance
type VerifyPerformanceResponse struct {
	BountyID         uint64          `json:"bounty_id"`
	TVLImpact        decimal.Decimal `json:"tvl_impact"`
	MAUImpact        int             `json:"mau_impact"`
	BonusEarned      decimal.Decimal `json:"bonus_earned"`
	ReputationBonus  int             `json:"reputation_bonus"`
}

// DeveloperStats represents developer statistics
type DeveloperStats struct {
	Address                string          `json:"address"`
	Username               string          `json:"username"`
	ReputationScore        int             `json:"reputation_score"`
	TotalBountiesCompleted int             `json:"total_bounties_completed"`
	TotalEarnings          decimal.Decimal `json:"total_earnings"`
	TVLContributed         decimal.Decimal `json:"tvl_contributed"`
	MAUContributed         int             `json:"mau_contributed"`
	AverageCompletionTime  int             `json:"average_completion_time"`
	SuccessRate            float64         `json:"success_rate"`
}

// BountyFilter represents filters for querying bounties
type BountyFilter struct {
	Status          *BountyStatus
	Category        *BountyCategory
	CreatorAddress  string
	AssigneeAddress string
	Limit           int
	Offset          int
}

// Blockchain request types
type BountyRequest struct {
	Title       string
	Description string
	Category    BountyCategory
	Reward      decimal.Decimal
	Deadline    int64
}
