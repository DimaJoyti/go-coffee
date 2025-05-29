package yield

import (
	"time"
)

// YieldPosition represents a yield farming or staking position
type YieldPosition struct {
	ID                string                 `json:"id" db:"id"`
	AccountID         string                 `json:"account_id" db:"account_id"`
	WalletID          string                 `json:"wallet_id" db:"wallet_id"`
	ProtocolID        string                 `json:"protocol_id" db:"protocol_id"`
	PoolID            string                 `json:"pool_id" db:"pool_id"`
	PositionType      PositionType           `json:"position_type" db:"position_type"`
	Strategy          string                 `json:"strategy" db:"strategy"`
	TokenAddress      string                 `json:"token_address" db:"token_address"`
	TokenSymbol       string                 `json:"token_symbol" db:"token_symbol"`
	Amount            string                 `json:"amount" db:"amount"`
	USDValue          string                 `json:"usd_value" db:"usd_value"`
	EntryPrice        string                 `json:"entry_price" db:"entry_price"`
	CurrentPrice      string                 `json:"current_price" db:"current_price"`
	APY               float64                `json:"apy" db:"apy"`
	APR               float64                `json:"apr" db:"apr"`
	DailyRewards      string                 `json:"daily_rewards" db:"daily_rewards"`
	TotalRewards      string                 `json:"total_rewards" db:"total_rewards"`
	ClaimedRewards    string                 `json:"claimed_rewards" db:"claimed_rewards"`
	PendingRewards    string                 `json:"pending_rewards" db:"pending_rewards"`
	ImpermanentLoss   string                 `json:"impermanent_loss" db:"impermanent_loss"`
	Status            PositionStatus         `json:"status" db:"status"`
	AutoCompound      bool                   `json:"auto_compound" db:"auto_compound"`
	LockPeriod        *time.Duration         `json:"lock_period" db:"lock_period"`
	UnlockDate        *time.Time             `json:"unlock_date" db:"unlock_date"`
	LastRewardClaim   *time.Time             `json:"last_reward_claim" db:"last_reward_claim"`
	LastCompound      *time.Time             `json:"last_compound" db:"last_compound"`
	RiskScore         float64                `json:"risk_score" db:"risk_score"`
	PerformanceMetrics PerformanceMetrics    `json:"performance_metrics" db:"performance_metrics"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// PositionType represents the type of yield position
type PositionType string

const (
	PositionTypeStaking         PositionType = "staking"
	PositionTypeLiquidityMining PositionType = "liquidity_mining"
	PositionTypeLending         PositionType = "lending"
	PositionTypeFarming         PositionType = "farming"
	PositionTypeVault           PositionType = "vault"
)

// PositionStatus represents the status of a yield position
type PositionStatus string

const (
	PositionStatusActive    PositionStatus = "active"
	PositionStatusInactive  PositionStatus = "inactive"
	PositionStatusLocked    PositionStatus = "locked"
	PositionStatusUnlocking PositionStatus = "unlocking"
	PositionStatusClosed    PositionStatus = "closed"
	PositionStatusError     PositionStatus = "error"
)

// PerformanceMetrics represents performance metrics for a position
type PerformanceMetrics struct {
	TotalReturn       string    `json:"total_return"`
	TotalReturnUSD    string    `json:"total_return_usd"`
	ROI               float64   `json:"roi"`
	DailyROI          float64   `json:"daily_roi"`
	WeeklyROI         float64   `json:"weekly_roi"`
	MonthlyROI        float64   `json:"monthly_roi"`
	YearlyROI         float64   `json:"yearly_roi"`
	Volatility        float64   `json:"volatility"`
	SharpeRatio       float64   `json:"sharpe_ratio"`
	MaxDrawdown       float64   `json:"max_drawdown"`
	DaysActive        int       `json:"days_active"`
	LastUpdated       time.Time `json:"last_updated"`
}

// Protocol represents a DeFi protocol
type Protocol struct {
	ID              string                 `json:"id" db:"id"`
	Name            string                 `json:"name" db:"name"`
	Description     string                 `json:"description" db:"description"`
	Website         string                 `json:"website" db:"website"`
	LogoURL         string                 `json:"logo_url" db:"logo_url"`
	Category        ProtocolCategory       `json:"category" db:"category"`
	Network         string                 `json:"network" db:"network"`
	ContractAddress string                 `json:"contract_address" db:"contract_address"`
	TVL             string                 `json:"tvl" db:"tvl"`
	Volume24h       string                 `json:"volume_24h" db:"volume_24h"`
	AverageAPY      float64                `json:"average_apy" db:"average_apy"`
	RiskScore       float64                `json:"risk_score" db:"risk_score"`
	IsActive        bool                   `json:"is_active" db:"is_active"`
	IsAudited       bool                   `json:"is_audited" db:"is_audited"`
	AuditReports    []string               `json:"audit_reports" db:"audit_reports"`
	SupportedTokens []string               `json:"supported_tokens" db:"supported_tokens"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// ProtocolCategory represents the category of a DeFi protocol
type ProtocolCategory string

const (
	ProtocolCategoryDEX         ProtocolCategory = "dex"
	ProtocolCategoryLending     ProtocolCategory = "lending"
	ProtocolCategoryStaking     ProtocolCategory = "staking"
	ProtocolCategoryYieldFarm   ProtocolCategory = "yield_farm"
	ProtocolCategoryVault       ProtocolCategory = "vault"
	ProtocolCategoryInsurance   ProtocolCategory = "insurance"
	ProtocolCategoryDerivatives ProtocolCategory = "derivatives"
)

// Pool represents a yield farming or staking pool
type Pool struct {
	ID              string                 `json:"id" db:"id"`
	ProtocolID      string                 `json:"protocol_id" db:"protocol_id"`
	Name            string                 `json:"name" db:"name"`
	Description     string                 `json:"description" db:"description"`
	PoolType        PoolType               `json:"pool_type" db:"pool_type"`
	TokenPair       []string               `json:"token_pair" db:"token_pair"`
	RewardTokens    []string               `json:"reward_tokens" db:"reward_tokens"`
	APY             float64                `json:"apy" db:"apy"`
	APR             float64                `json:"apr" db:"apr"`
	TVL             string                 `json:"tvl" db:"tvl"`
	Volume24h       string                 `json:"volume_24h" db:"volume_24h"`
	MinDeposit      string                 `json:"min_deposit" db:"min_deposit"`
	MaxDeposit      string                 `json:"max_deposit" db:"max_deposit"`
	LockPeriod      *time.Duration         `json:"lock_period" db:"lock_period"`
	WithdrawalFee   float64                `json:"withdrawal_fee" db:"withdrawal_fee"`
	PerformanceFee  float64                `json:"performance_fee" db:"performance_fee"`
	RiskScore       float64                `json:"risk_score" db:"risk_score"`
	IsActive        bool                   `json:"is_active" db:"is_active"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// PoolType represents the type of pool
type PoolType string

const (
	PoolTypeStaking         PoolType = "staking"
	PoolTypeLiquidityMining PoolType = "liquidity_mining"
	PoolTypeLending         PoolType = "lending"
	PoolTypeVault           PoolType = "vault"
	PoolTypeFarming         PoolType = "farming"
)

// Strategy represents a yield optimization strategy
type Strategy struct {
	ID              string                 `json:"id" db:"id"`
	Name            string                 `json:"name" db:"name"`
	Description     string                 `json:"description" db:"description"`
	StrategyType    StrategyType           `json:"strategy_type" db:"strategy_type"`
	RiskLevel       RiskLevel              `json:"risk_level" db:"risk_level"`
	TargetAPY       float64                `json:"target_apy" db:"target_apy"`
	MaxDrawdown     float64                `json:"max_drawdown" db:"max_drawdown"`
	MinInvestment   string                 `json:"min_investment" db:"min_investment"`
	MaxInvestment   string                 `json:"max_investment" db:"max_investment"`
	SupportedTokens []string               `json:"supported_tokens" db:"supported_tokens"`
	Protocols       []string               `json:"protocols" db:"protocols"`
	AutoRebalance   bool                   `json:"auto_rebalance" db:"auto_rebalance"`
	AutoCompound    bool                   `json:"auto_compound" db:"auto_compound"`
	IsActive        bool                   `json:"is_active" db:"is_active"`
	Performance     StrategyPerformance    `json:"performance" db:"performance"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// StrategyType represents the type of strategy
type StrategyType string

const (
	StrategyTypeConservative StrategyType = "conservative"
	StrategyTypeModerate     StrategyType = "moderate"
	StrategyTypeAggressive   StrategyType = "aggressive"
	StrategyTypeCustom       StrategyType = "custom"
)

// RiskLevel represents the risk level of a strategy
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
)

// StrategyPerformance represents strategy performance metrics
type StrategyPerformance struct {
	TotalReturn     string    `json:"total_return"`
	AnnualizedReturn float64   `json:"annualized_return"`
	Volatility      float64   `json:"volatility"`
	SharpeRatio     float64   `json:"sharpe_ratio"`
	MaxDrawdown     float64   `json:"max_drawdown"`
	WinRate         float64   `json:"win_rate"`
	TotalPositions  int       `json:"total_positions"`
	ActivePositions int       `json:"active_positions"`
	LastUpdated     time.Time `json:"last_updated"`
}

// Reward represents earned rewards from yield farming
type Reward struct {
	ID            string                 `json:"id" db:"id"`
	PositionID    string                 `json:"position_id" db:"position_id"`
	AccountID     string                 `json:"account_id" db:"account_id"`
	TokenAddress  string                 `json:"token_address" db:"token_address"`
	TokenSymbol   string                 `json:"token_symbol" db:"token_symbol"`
	Amount        string                 `json:"amount" db:"amount"`
	USDValue      string                 `json:"usd_value" db:"usd_value"`
	RewardType    RewardType             `json:"reward_type" db:"reward_type"`
	Status        RewardStatus           `json:"status" db:"status"`
	ClaimedAt     *time.Time             `json:"claimed_at" db:"claimed_at"`
	TransactionHash string               `json:"transaction_hash" db:"transaction_hash"`
	Metadata      map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// RewardType represents the type of reward
type RewardType string

const (
	RewardTypeStaking    RewardType = "staking"
	RewardTypeFarming    RewardType = "farming"
	RewardTypeLiquidity  RewardType = "liquidity"
	RewardTypeGovernance RewardType = "governance"
	RewardTypeBonus      RewardType = "bonus"
)

// RewardStatus represents the status of a reward
type RewardStatus string

const (
	RewardStatusPending   RewardStatus = "pending"
	RewardStatusAvailable RewardStatus = "available"
	RewardStatusClaimed   RewardStatus = "claimed"
	RewardStatusExpired   RewardStatus = "expired"
)

// CreatePositionRequest represents a request to create a yield position
type CreatePositionRequest struct {
	AccountID      string                 `json:"account_id" validate:"required"`
	WalletID       string                 `json:"wallet_id" validate:"required"`
	ProtocolID     string                 `json:"protocol_id" validate:"required"`
	PoolID         string                 `json:"pool_id" validate:"required"`
	PositionType   PositionType           `json:"position_type" validate:"required"`
	Strategy       string                 `json:"strategy,omitempty"`
	TokenAddress   string                 `json:"token_address" validate:"required"`
	Amount         string                 `json:"amount" validate:"required"`
	AutoCompound   bool                   `json:"auto_compound"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// UpdatePositionRequest represents a request to update a yield position
type UpdatePositionRequest struct {
	AutoCompound *bool                  `json:"auto_compound,omitempty"`
	Strategy     *string                `json:"strategy,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ClosePositionRequest represents a request to close a yield position
type ClosePositionRequest struct {
	Amount   string                 `json:"amount,omitempty"` // If empty, close entire position
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ClaimRewardsRequest represents a request to claim rewards
type ClaimRewardsRequest struct {
	PositionIDs []string               `json:"position_ids" validate:"required"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CompoundRewardsRequest represents a request to compound rewards
type CompoundRewardsRequest struct {
	PositionIDs []string               `json:"position_ids" validate:"required"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PositionListRequest represents a request to list yield positions
type PositionListRequest struct {
	Page         int            `json:"page" validate:"min=1"`
	Limit        int            `json:"limit" validate:"min=1,max=100"`
	AccountID    string         `json:"account_id,omitempty"`
	ProtocolID   string         `json:"protocol_id,omitempty"`
	PositionType PositionType   `json:"position_type,omitempty"`
	Status       PositionStatus `json:"status,omitempty"`
	TokenSymbol  string         `json:"token_symbol,omitempty"`
}

// PositionListResponse represents a response to list yield positions
type PositionListResponse struct {
	Positions  []YieldPosition `json:"positions"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"total_pages"`
}

// ProtocolListRequest represents a request to list protocols
type ProtocolListRequest struct {
	Page     int              `json:"page" validate:"min=1"`
	Limit    int              `json:"limit" validate:"min=1,max=100"`
	Category ProtocolCategory `json:"category,omitempty"`
	Network  string           `json:"network,omitempty"`
	IsActive *bool            `json:"is_active,omitempty"`
}

// ProtocolListResponse represents a response to list protocols
type ProtocolListResponse struct {
	Protocols  []Protocol `json:"protocols"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	TotalPages int        `json:"total_pages"`
}
