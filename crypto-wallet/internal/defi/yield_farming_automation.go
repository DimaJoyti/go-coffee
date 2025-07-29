package defi

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/redis"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// YieldFarmingAutomation provides advanced yield farming automation with auto-compounding and optimization
type YieldFarmingAutomation struct {
	logger *logger.Logger
	cache  redis.Client

	// Configuration
	config YieldFarmingConfig

	// Protocol clients
	protocolClients map[string]YieldProtocolClient

	// State tracking
	activeFarms        map[string]*ActiveFarm
	farmingStrategies  map[string]*FarmingStrategy
	yieldOpportunities map[string]*YieldOpportunity
	compoundingTasks   map[string]*CompoundingTask
	migrationTasks     map[string]*MigrationTask
	performanceMetrics *YieldFarmingMetrics
	ilProtection       *ImpermanentLossProtection
	mutex              sync.RWMutex
	stopChan           chan struct{}
	isRunning          bool
}

// YieldFarmingConfig holds configuration for yield farming automation
type YieldFarmingConfig struct {
	Enabled                   bool            `json:"enabled" yaml:"enabled"`
	AutoCompoundingEnabled    bool            `json:"auto_compounding_enabled" yaml:"auto_compounding_enabled"`
	PoolMigrationEnabled      bool            `json:"pool_migration_enabled" yaml:"pool_migration_enabled"`
	ImpermanentLossProtection bool            `json:"impermanent_loss_protection" yaml:"impermanent_loss_protection"`
	MinYieldThreshold         decimal.Decimal `json:"min_yield_threshold" yaml:"min_yield_threshold"`
	MaxSlippageTolerance      decimal.Decimal `json:"max_slippage_tolerance" yaml:"max_slippage_tolerance"`
	CompoundingInterval       time.Duration   `json:"compounding_interval" yaml:"compounding_interval"`
	OpportunityCheckInterval  time.Duration   `json:"opportunity_check_interval" yaml:"opportunity_check_interval"`
	MaxPositionSize           decimal.Decimal `json:"max_position_size" yaml:"max_position_size"`
	MinCompoundingAmount      decimal.Decimal `json:"min_compounding_amount" yaml:"min_compounding_amount"`
	GasOptimizationEnabled    bool            `json:"gas_optimization_enabled" yaml:"gas_optimization_enabled"`
	MaxGasPriceGwei           decimal.Decimal `json:"max_gas_price_gwei" yaml:"max_gas_price_gwei"`
	SupportedProtocols        []string        `json:"supported_protocols" yaml:"supported_protocols"`
	SupportedChains           []string        `json:"supported_chains" yaml:"supported_chains"`
	RiskLevel                 string          `json:"risk_level" yaml:"risk_level"` // conservative, moderate, aggressive
}

// ActiveFarm represents an active yield farming position
type ActiveFarm struct {
	ID              string          `json:"id"`
	Protocol        string          `json:"protocol"`
	Chain           string          `json:"chain"`
	PoolAddress     string          `json:"pool_address"`
	PoolName        string          `json:"pool_name"`
	Token0          string          `json:"token0"`
	Token1          string          `json:"token1"`
	LiquidityAmount decimal.Decimal `json:"liquidity_amount"`
	Token0Amount    decimal.Decimal `json:"token0_amount"`
	Token1Amount    decimal.Decimal `json:"token1_amount"`
	CurrentAPY      decimal.Decimal `json:"current_apy"`
	RewardsEarned   decimal.Decimal `json:"rewards_earned"`
	RewardTokens    []string        `json:"reward_tokens"`
	EntryPrice0     decimal.Decimal `json:"entry_price0"`
	EntryPrice1     decimal.Decimal `json:"entry_price1"`
	CurrentPrice0   decimal.Decimal `json:"current_price0"`
	CurrentPrice1   decimal.Decimal `json:"current_price1"`
	ImpermanentLoss decimal.Decimal `json:"impermanent_loss"`
	TotalValue      decimal.Decimal `json:"total_value"`
	ProfitLoss      decimal.Decimal `json:"profit_loss"`
	LastCompounded  time.Time       `json:"last_compounded"`
	CreatedAt       time.Time       `json:"created_at"`
	Status          string          `json:"status"` // active, migrating, withdrawing, paused
	AutoCompounding bool            `json:"auto_compounding"`
	Strategy        string          `json:"strategy"`
}

// FarmingStrategy defines a yield farming strategy
type FarmingStrategy struct {
	ID                   string          `json:"id"`
	Name                 string          `json:"name"`
	Description          string          `json:"description"`
	TargetAPY            decimal.Decimal `json:"target_apy"`
	MaxImpermanentLoss   decimal.Decimal `json:"max_impermanent_loss"`
	PreferredProtocols   []string        `json:"preferred_protocols"`
	PreferredChains      []string        `json:"preferred_chains"`
	RiskLevel            string          `json:"risk_level"`
	CompoundingFrequency time.Duration   `json:"compounding_frequency"`
	RebalanceThreshold   decimal.Decimal `json:"rebalance_threshold"`
	ExitConditions       []ExitCondition `json:"exit_conditions"`
	AllocationPercentage decimal.Decimal `json:"allocation_percentage"`
	IsActive             bool            `json:"is_active"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

// ExitCondition defines conditions for exiting a farming position
type ExitCondition struct {
	Type      string          `json:"type"` // apy_drop, impermanent_loss, time_based, profit_target
	Threshold decimal.Decimal `json:"threshold"`
	Duration  time.Duration   `json:"duration,omitempty"`
	Enabled   bool            `json:"enabled"`
}

// YieldOpportunity represents a detected yield farming opportunity
type YieldOpportunity struct {
	ID                  string          `json:"id"`
	Protocol            string          `json:"protocol"`
	Chain               string          `json:"chain"`
	PoolAddress         string          `json:"pool_address"`
	PoolName            string          `json:"pool_name"`
	Token0              string          `json:"token0"`
	Token1              string          `json:"token1"`
	CurrentAPY          decimal.Decimal `json:"current_apy"`
	ProjectedAPY        decimal.Decimal `json:"projected_apy"`
	TVL                 decimal.Decimal `json:"tvl"`
	Volume24h           decimal.Decimal `json:"volume_24h"`
	FeeAPR              decimal.Decimal `json:"fee_apr"`
	RewardAPR           decimal.Decimal `json:"reward_apr"`
	RiskScore           decimal.Decimal `json:"risk_score"`
	ImpermanentLossRisk decimal.Decimal `json:"impermanent_loss_risk"`
	LiquidityDepth      decimal.Decimal `json:"liquidity_depth"`
	EntrySlippage       decimal.Decimal `json:"entry_slippage"`
	ExitSlippage        decimal.Decimal `json:"exit_slippage"`
	RecommendedAmount   decimal.Decimal `json:"recommended_amount"`
	Confidence          decimal.Decimal `json:"confidence"`
	DetectedAt          time.Time       `json:"detected_at"`
	ExpiresAt           time.Time       `json:"expires_at"`
	Status              string          `json:"status"` // detected, analyzing, recommended, executed, expired
}

// CompoundingTask represents an auto-compounding task
type CompoundingTask struct {
	ID              string          `json:"id"`
	FarmID          string          `json:"farm_id"`
	Protocol        string          `json:"protocol"`
	PoolAddress     string          `json:"pool_address"`
	RewardAmount    decimal.Decimal `json:"reward_amount"`
	RewardTokens    []string        `json:"reward_tokens"`
	EstimatedGas    decimal.Decimal `json:"estimated_gas"`
	EstimatedProfit decimal.Decimal `json:"estimated_profit"`
	ScheduledAt     time.Time       `json:"scheduled_at"`
	ExecutedAt      *time.Time      `json:"executed_at,omitempty"`
	Status          string          `json:"status"` // scheduled, executing, completed, failed, cancelled
	TransactionHash string          `json:"transaction_hash,omitempty"`
	GasUsed         decimal.Decimal `json:"gas_used,omitempty"`
	ActualProfit    decimal.Decimal `json:"actual_profit,omitempty"`
	ErrorMessage    string          `json:"error_message,omitempty"`
}

// MigrationTask represents a pool migration task
type MigrationTask struct {
	ID                string          `json:"id"`
	FromFarmID        string          `json:"from_farm_id"`
	ToOpportunityID   string          `json:"to_opportunity_id"`
	FromProtocol      string          `json:"from_protocol"`
	ToProtocol        string          `json:"to_protocol"`
	FromPool          string          `json:"from_pool"`
	ToPool            string          `json:"to_pool"`
	Amount            decimal.Decimal `json:"amount"`
	CurrentAPY        decimal.Decimal `json:"current_apy"`
	TargetAPY         decimal.Decimal `json:"target_apy"`
	ExpectedGain      decimal.Decimal `json:"expected_gain"`
	EstimatedCost     decimal.Decimal `json:"estimated_cost"`
	NetBenefit        decimal.Decimal `json:"net_benefit"`
	ScheduledAt       time.Time       `json:"scheduled_at"`
	ExecutedAt        *time.Time      `json:"executed_at,omitempty"`
	Status            string          `json:"status"` // scheduled, executing, completed, failed, cancelled
	TransactionHashes []string        `json:"transaction_hashes,omitempty"`
	ActualCost        decimal.Decimal `json:"actual_cost,omitempty"`
	ActualGain        decimal.Decimal `json:"actual_gain,omitempty"`
	ErrorMessage      string          `json:"error_message,omitempty"`
}

// ImpermanentLossProtection provides impermanent loss protection mechanisms
type ImpermanentLossProtection struct {
	Enabled              bool            `json:"enabled"`
	MaxAcceptableLoss    decimal.Decimal `json:"max_acceptable_loss"`
	HedgingEnabled       bool            `json:"hedging_enabled"`
	HedgingThreshold     decimal.Decimal `json:"hedging_threshold"`
	AutoExitEnabled      bool            `json:"auto_exit_enabled"`
	AutoExitThreshold    decimal.Decimal `json:"auto_exit_threshold"`
	RebalancingEnabled   bool            `json:"rebalancing_enabled"`
	RebalancingThreshold decimal.Decimal `json:"rebalancing_threshold"`
	InsuranceEnabled     bool            `json:"insurance_enabled"`
	InsuranceProvider    string          `json:"insurance_provider"`
	MonitoringInterval   time.Duration   `json:"monitoring_interval"`
}

// YieldFarmingMetrics holds performance metrics for yield farming
type YieldFarmingMetrics struct {
	TotalValueLocked     decimal.Decimal            `json:"total_value_locked"`
	TotalRewardsEarned   decimal.Decimal            `json:"total_rewards_earned"`
	TotalFeesEarned      decimal.Decimal            `json:"total_fees_earned"`
	TotalImpermanentLoss decimal.Decimal            `json:"total_impermanent_loss"`
	NetProfit            decimal.Decimal            `json:"net_profit"`
	AverageAPY           decimal.Decimal            `json:"average_apy"`
	ActiveFarmsCount     int64                      `json:"active_farms_count"`
	CompletedCompounds   int64                      `json:"completed_compounds"`
	SuccessfulMigrations int64                      `json:"successful_migrations"`
	FailedTransactions   int64                      `json:"failed_transactions"`
	TotalGasSpent        decimal.Decimal            `json:"total_gas_spent"`
	ProtocolDistribution map[string]decimal.Decimal `json:"protocol_distribution"`
	ChainDistribution    map[string]decimal.Decimal `json:"chain_distribution"`
	LastUpdated          time.Time                  `json:"last_updated"`
}

// YieldProtocolClient interface for interacting with yield farming protocols
type YieldProtocolClient interface {
	GetPoolInfo(ctx context.Context, poolAddress string) (*PoolInfo, error)
	GetUserPosition(ctx context.Context, userAddress, poolAddress string) (*UserPosition, error)
	DepositLiquidity(ctx context.Context, params *DepositParams) (*YieldTransactionResult, error)
	WithdrawLiquidity(ctx context.Context, params *WithdrawParams) (*YieldTransactionResult, error)
	ClaimRewards(ctx context.Context, params *ClaimParams) (*YieldTransactionResult, error)
	CompoundRewards(ctx context.Context, params *CompoundParams) (*YieldTransactionResult, error)
	GetAvailablePools(ctx context.Context) ([]*PoolInfo, error)
	EstimateGas(ctx context.Context, operation string, params interface{}) (decimal.Decimal, error)
	GetSupportedTokens() []string
	GetProtocolName() string
}

// Data structures for protocol interactions
type PoolInfo struct {
	Address        string          `json:"address"`
	Name           string          `json:"name"`
	Protocol       string          `json:"protocol"`
	Chain          string          `json:"chain"`
	Token0         string          `json:"token0"`
	Token1         string          `json:"token1"`
	Token0Symbol   string          `json:"token0_symbol"`
	Token1Symbol   string          `json:"token1_symbol"`
	TVL            decimal.Decimal `json:"tvl"`
	Volume24h      decimal.Decimal `json:"volume_24h"`
	FeeAPR         decimal.Decimal `json:"fee_apr"`
	RewardAPR      decimal.Decimal `json:"reward_apr"`
	TotalAPY       decimal.Decimal `json:"total_apy"`
	RewardTokens   []string        `json:"reward_tokens"`
	LiquidityDepth decimal.Decimal `json:"liquidity_depth"`
	IsActive       bool            `json:"is_active"`
	RiskLevel      string          `json:"risk_level"`
	LastUpdated    time.Time       `json:"last_updated"`
}

type UserPosition struct {
	PoolAddress     string          `json:"pool_address"`
	LiquidityAmount decimal.Decimal `json:"liquidity_amount"`
	Token0Amount    decimal.Decimal `json:"token0_amount"`
	Token1Amount    decimal.Decimal `json:"token1_amount"`
	RewardsEarned   decimal.Decimal `json:"rewards_earned"`
	RewardTokens    []string        `json:"reward_tokens"`
	LastUpdated     time.Time       `json:"last_updated"`
}

type DepositParams struct {
	PoolAddress  string          `json:"pool_address"`
	Token0Amount decimal.Decimal `json:"token0_amount"`
	Token1Amount decimal.Decimal `json:"token1_amount"`
	MinLiquidity decimal.Decimal `json:"min_liquidity"`
	Deadline     time.Time       `json:"deadline"`
	Slippage     decimal.Decimal `json:"slippage"`
}

type WithdrawParams struct {
	PoolAddress     string          `json:"pool_address"`
	LiquidityAmount decimal.Decimal `json:"liquidity_amount"`
	MinToken0       decimal.Decimal `json:"min_token0"`
	MinToken1       decimal.Decimal `json:"min_token1"`
	Deadline        time.Time       `json:"deadline"`
}

type ClaimParams struct {
	PoolAddress  string   `json:"pool_address"`
	RewardTokens []string `json:"reward_tokens"`
}

type CompoundParams struct {
	PoolAddress  string          `json:"pool_address"`
	RewardAmount decimal.Decimal `json:"reward_amount"`
	MinLiquidity decimal.Decimal `json:"min_liquidity"`
	Slippage     decimal.Decimal `json:"slippage"`
}

type YieldTransactionResult struct {
	TransactionHash string          `json:"transaction_hash"`
	Success         bool            `json:"success"`
	GasUsed         decimal.Decimal `json:"gas_used"`
	GasPrice        decimal.Decimal `json:"gas_price"`
	BlockNumber     uint64          `json:"block_number"`
	Timestamp       time.Time       `json:"timestamp"`
	ErrorMessage    string          `json:"error_message,omitempty"`
}

// NewYieldFarmingAutomation creates a new yield farming automation engine
func NewYieldFarmingAutomation(logger *logger.Logger, cache redis.Client, config YieldFarmingConfig) *YieldFarmingAutomation {
	return &YieldFarmingAutomation{
		logger:             logger.Named("yield-farming-automation"),
		cache:              cache,
		config:             config,
		protocolClients:    make(map[string]YieldProtocolClient),
		activeFarms:        make(map[string]*ActiveFarm),
		farmingStrategies:  make(map[string]*FarmingStrategy),
		yieldOpportunities: make(map[string]*YieldOpportunity),
		compoundingTasks:   make(map[string]*CompoundingTask),
		migrationTasks:     make(map[string]*MigrationTask),
		performanceMetrics: &YieldFarmingMetrics{
			ProtocolDistribution: make(map[string]decimal.Decimal),
			ChainDistribution:    make(map[string]decimal.Decimal),
		},
		ilProtection: &ImpermanentLossProtection{
			Enabled:              config.ImpermanentLossProtection,
			MaxAcceptableLoss:    decimal.NewFromFloat(0.05), // 5%
			HedgingEnabled:       false,
			HedgingThreshold:     decimal.NewFromFloat(0.02), // 2%
			AutoExitEnabled:      true,
			AutoExitThreshold:    decimal.NewFromFloat(0.1), // 10%
			RebalancingEnabled:   true,
			RebalancingThreshold: decimal.NewFromFloat(0.03), // 3%
			MonitoringInterval:   5 * time.Minute,
		},
		stopChan: make(chan struct{}),
	}
}

// Start starts the yield farming automation engine
func (yfa *YieldFarmingAutomation) Start(ctx context.Context) error {
	yfa.mutex.Lock()
	defer yfa.mutex.Unlock()

	if yfa.isRunning {
		return fmt.Errorf("yield farming automation is already running")
	}

	if !yfa.config.Enabled {
		yfa.logger.Info("Yield farming automation is disabled")
		return nil
	}

	yfa.logger.Info("Starting yield farming automation",
		zap.Bool("auto_compounding", yfa.config.AutoCompoundingEnabled),
		zap.Bool("pool_migration", yfa.config.PoolMigrationEnabled),
		zap.Bool("il_protection", yfa.config.ImpermanentLossProtection),
		zap.Strings("protocols", yfa.config.SupportedProtocols),
		zap.Strings("chains", yfa.config.SupportedChains))

	// Initialize protocol clients
	if err := yfa.initializeProtocolClients(); err != nil {
		return fmt.Errorf("failed to initialize protocol clients: %w", err)
	}

	// Load existing farms and strategies
	if err := yfa.loadExistingData(ctx); err != nil {
		return fmt.Errorf("failed to load existing data: %w", err)
	}

	yfa.isRunning = true

	// Start automation loops
	go yfa.opportunityDetectionLoop(ctx)
	go yfa.autoCompoundingLoop(ctx)
	go yfa.poolMigrationLoop(ctx)
	go yfa.impermanentLossMonitoringLoop(ctx)
	go yfa.performanceTrackingLoop(ctx)
	go yfa.strategyExecutionLoop(ctx)

	yfa.logger.Info("Yield farming automation started successfully")
	return nil
}

// Stop stops the yield farming automation engine
func (yfa *YieldFarmingAutomation) Stop() error {
	yfa.mutex.Lock()
	defer yfa.mutex.Unlock()

	if !yfa.isRunning {
		return nil
	}

	yfa.logger.Info("Stopping yield farming automation")
	yfa.isRunning = false
	close(yfa.stopChan)

	yfa.logger.Info("Yield farming automation stopped")
	return nil
}

// Core functionality methods

// ScanForYieldOpportunities scans for new yield farming opportunities
func (yfa *YieldFarmingAutomation) ScanForYieldOpportunities(ctx context.Context) ([]*YieldOpportunity, error) {
	yfa.logger.Debug("Scanning for yield opportunities")

	var allOpportunities []*YieldOpportunity

	// Scan each supported protocol
	for _, protocolName := range yfa.config.SupportedProtocols {
		client, exists := yfa.protocolClients[protocolName]
		if !exists {
			yfa.logger.Warn("Protocol client not found", zap.String("protocol", protocolName))
			continue
		}

		pools, err := client.GetAvailablePools(ctx)
		if err != nil {
			yfa.logger.Error("Failed to get pools for protocol",
				zap.String("protocol", protocolName),
				zap.Error(err))
			continue
		}

		for _, pool := range pools {
			opportunity := yfa.analyzePoolOpportunity(pool)
			if opportunity != nil && yfa.isOpportunityViable(opportunity) {
				allOpportunities = append(allOpportunities, opportunity)
			}
		}
	}

	// Filter and rank opportunities
	filteredOpportunities := yfa.filterOpportunities(allOpportunities)
	rankedOpportunities := yfa.rankOpportunities(filteredOpportunities)

	// Store opportunities
	yfa.mutex.Lock()
	for _, opp := range rankedOpportunities {
		yfa.yieldOpportunities[opp.ID] = opp
	}
	yfa.mutex.Unlock()

	yfa.logger.Info("Yield opportunity scan completed",
		zap.Int("total_found", len(allOpportunities)),
		zap.Int("viable_opportunities", len(rankedOpportunities)))

	return rankedOpportunities, nil
}

// EnterFarm enters a new yield farming position
func (yfa *YieldFarmingAutomation) EnterFarm(ctx context.Context, opportunityID string, amount decimal.Decimal, strategyID string) (*ActiveFarm, error) {
	yfa.logger.Info("Entering yield farm",
		zap.String("opportunity_id", opportunityID),
		zap.String("amount", amount.String()),
		zap.String("strategy_id", strategyID))

	// Get opportunity
	yfa.mutex.RLock()
	opportunity, exists := yfa.yieldOpportunities[opportunityID]
	yfa.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("opportunity not found: %s", opportunityID)
	}

	// Get strategy
	yfa.mutex.RLock()
	strategy, strategyExists := yfa.farmingStrategies[strategyID]
	yfa.mutex.RUnlock()

	if !strategyExists {
		return nil, fmt.Errorf("strategy not found: %s", strategyID)
	}

	// Validate amount
	if amount.GreaterThan(yfa.config.MaxPositionSize) {
		return nil, fmt.Errorf("amount exceeds maximum position size")
	}

	// Get protocol client
	client, exists := yfa.protocolClients[opportunity.Protocol]
	if !exists {
		return nil, fmt.Errorf("protocol client not found: %s", opportunity.Protocol)
	}

	// Calculate token amounts based on current pool ratio
	token0Amount, token1Amount, err := yfa.calculateOptimalTokenAmounts(opportunity, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate token amounts: %w", err)
	}

	// Prepare deposit parameters
	depositParams := &DepositParams{
		PoolAddress:  opportunity.PoolAddress,
		Token0Amount: token0Amount,
		Token1Amount: token1Amount,
		MinLiquidity: amount.Mul(decimal.NewFromFloat(0.95)), // 5% slippage tolerance
		Deadline:     time.Now().Add(10 * time.Minute),
		Slippage:     yfa.config.MaxSlippageTolerance,
	}

	// Execute deposit
	result, err := client.DepositLiquidity(ctx, depositParams)
	if err != nil {
		return nil, fmt.Errorf("failed to deposit liquidity: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("deposit transaction failed: %s", result.ErrorMessage)
	}

	// Create active farm
	farm := &ActiveFarm{
		ID:              yfa.generateFarmID(),
		Protocol:        opportunity.Protocol,
		Chain:           opportunity.Chain,
		PoolAddress:     opportunity.PoolAddress,
		PoolName:        opportunity.PoolName,
		Token0:          opportunity.Token0,
		Token1:          opportunity.Token1,
		LiquidityAmount: amount,
		Token0Amount:    token0Amount,
		Token1Amount:    token1Amount,
		CurrentAPY:      opportunity.CurrentAPY,
		RewardsEarned:   decimal.Zero,
		RewardTokens:    []string{}, // Will be populated when rewards are earned
		EntryPrice0:     yfa.getCurrentPrice(opportunity.Token0),
		EntryPrice1:     yfa.getCurrentPrice(opportunity.Token1),
		CurrentPrice0:   yfa.getCurrentPrice(opportunity.Token0),
		CurrentPrice1:   yfa.getCurrentPrice(opportunity.Token1),
		ImpermanentLoss: decimal.Zero,
		TotalValue:      amount,
		ProfitLoss:      decimal.Zero,
		LastCompounded:  time.Now(),
		CreatedAt:       time.Now(),
		Status:          "active",
		AutoCompounding: strategy.CompoundingFrequency > 0,
		Strategy:        strategyID,
	}

	// Store active farm
	yfa.mutex.Lock()
	yfa.activeFarms[farm.ID] = farm
	yfa.mutex.Unlock()

	// Update metrics
	yfa.updateMetricsAfterEntry(farm)

	yfa.logger.Info("Successfully entered yield farm",
		zap.String("farm_id", farm.ID),
		zap.String("transaction_hash", result.TransactionHash),
		zap.String("gas_used", result.GasUsed.String()))

	return farm, nil
}

// ExitFarm exits a yield farming position
func (yfa *YieldFarmingAutomation) ExitFarm(ctx context.Context, farmID string, percentage decimal.Decimal) (*YieldTransactionResult, error) {
	yfa.logger.Info("Exiting yield farm",
		zap.String("farm_id", farmID),
		zap.String("percentage", percentage.String()))

	// Get active farm
	yfa.mutex.RLock()
	farm, exists := yfa.activeFarms[farmID]
	yfa.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("active farm not found: %s", farmID)
	}

	// Validate percentage
	if percentage.LessThanOrEqual(decimal.Zero) || percentage.GreaterThan(decimal.NewFromFloat(1.0)) {
		return nil, fmt.Errorf("invalid percentage: must be between 0 and 1")
	}

	// Get protocol client
	client, exists := yfa.protocolClients[farm.Protocol]
	if !exists {
		return nil, fmt.Errorf("protocol client not found: %s", farm.Protocol)
	}

	// First claim any pending rewards
	if farm.RewardsEarned.GreaterThan(decimal.Zero) {
		claimParams := &ClaimParams{
			PoolAddress:  farm.PoolAddress,
			RewardTokens: farm.RewardTokens,
		}

		_, err := client.ClaimRewards(ctx, claimParams)
		if err != nil {
			yfa.logger.Warn("Failed to claim rewards before exit", zap.Error(err))
		}
	}

	// Calculate withdrawal amount
	withdrawAmount := farm.LiquidityAmount.Mul(percentage)
	minToken0 := farm.Token0Amount.Mul(percentage).Mul(decimal.NewFromFloat(0.95)) // 5% slippage
	minToken1 := farm.Token1Amount.Mul(percentage).Mul(decimal.NewFromFloat(0.95))

	// Prepare withdrawal parameters
	withdrawParams := &WithdrawParams{
		PoolAddress:     farm.PoolAddress,
		LiquidityAmount: withdrawAmount,
		MinToken0:       minToken0,
		MinToken1:       minToken1,
		Deadline:        time.Now().Add(10 * time.Minute),
	}

	// Execute withdrawal
	result, err := client.WithdrawLiquidity(ctx, withdrawParams)
	if err != nil {
		return nil, fmt.Errorf("failed to withdraw liquidity: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("withdrawal transaction failed: %s", result.ErrorMessage)
	}

	// Update farm state
	yfa.mutex.Lock()
	if percentage.Equal(decimal.NewFromFloat(1.0)) {
		// Complete exit
		farm.Status = "exited"
		delete(yfa.activeFarms, farmID)
	} else {
		// Partial exit
		farm.LiquidityAmount = farm.LiquidityAmount.Mul(decimal.NewFromFloat(1.0).Sub(percentage))
		farm.Token0Amount = farm.Token0Amount.Mul(decimal.NewFromFloat(1.0).Sub(percentage))
		farm.Token1Amount = farm.Token1Amount.Mul(decimal.NewFromFloat(1.0).Sub(percentage))
		farm.TotalValue = farm.TotalValue.Mul(decimal.NewFromFloat(1.0).Sub(percentage))
	}
	yfa.mutex.Unlock()

	// Update metrics
	yfa.updateMetricsAfterExit(farm, percentage)

	yfa.logger.Info("Successfully exited yield farm",
		zap.String("farm_id", farmID),
		zap.String("transaction_hash", result.TransactionHash),
		zap.String("gas_used", result.GasUsed.String()))

	return result, nil
}

// CompoundRewards compounds rewards for a specific farm
func (yfa *YieldFarmingAutomation) CompoundRewards(ctx context.Context, farmID string) (*YieldTransactionResult, error) {
	yfa.logger.Debug("Compounding rewards", zap.String("farm_id", farmID))

	// Get active farm
	yfa.mutex.RLock()
	farm, exists := yfa.activeFarms[farmID]
	yfa.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("active farm not found: %s", farmID)
	}

	// Check if compounding is worthwhile
	if farm.RewardsEarned.LessThan(yfa.config.MinCompoundingAmount) {
		return nil, fmt.Errorf("reward amount too small for compounding: %s", farm.RewardsEarned.String())
	}

	// Get protocol client
	client, exists := yfa.protocolClients[farm.Protocol]
	if !exists {
		return nil, fmt.Errorf("protocol client not found: %s", farm.Protocol)
	}

	// Estimate gas cost
	gasEstimate, err := client.EstimateGas(ctx, "compound", &CompoundParams{
		PoolAddress:  farm.PoolAddress,
		RewardAmount: farm.RewardsEarned,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Check if compounding is profitable after gas costs
	gasCost := gasEstimate.Mul(yfa.getCurrentGasPrice())
	if gasCost.GreaterThanOrEqual(farm.RewardsEarned.Mul(decimal.NewFromFloat(0.1))) {
		return nil, fmt.Errorf("gas cost too high for compounding: %s vs reward %s", gasCost.String(), farm.RewardsEarned.String())
	}

	// Prepare compound parameters
	compoundParams := &CompoundParams{
		PoolAddress:  farm.PoolAddress,
		RewardAmount: farm.RewardsEarned,
		MinLiquidity: farm.RewardsEarned.Mul(decimal.NewFromFloat(0.95)), // 5% slippage
		Slippage:     yfa.config.MaxSlippageTolerance,
	}

	// Execute compounding
	result, err := client.CompoundRewards(ctx, compoundParams)
	if err != nil {
		return nil, fmt.Errorf("failed to compound rewards: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("compound transaction failed: %s", result.ErrorMessage)
	}

	// Update farm state
	yfa.mutex.Lock()
	farm.LiquidityAmount = farm.LiquidityAmount.Add(farm.RewardsEarned)
	farm.TotalValue = farm.TotalValue.Add(farm.RewardsEarned)
	farm.RewardsEarned = decimal.Zero
	farm.LastCompounded = time.Now()
	yfa.mutex.Unlock()

	// Update metrics
	yfa.updateMetricsAfterCompounding(farm, farm.RewardsEarned)

	yfa.logger.Info("Successfully compounded rewards",
		zap.String("farm_id", farmID),
		zap.String("transaction_hash", result.TransactionHash),
		zap.String("compounded_amount", farm.RewardsEarned.String()))

	return result, nil
}

// Helper methods

// analyzePoolOpportunity analyzes a pool to create a yield opportunity
func (yfa *YieldFarmingAutomation) analyzePoolOpportunity(pool *PoolInfo) *YieldOpportunity {
	// Calculate risk score based on various factors
	riskScore := yfa.calculatePoolRiskScore(pool)

	// Calculate impermanent loss risk
	ilRisk := yfa.calculateImpermanentLossRisk(pool)

	// Calculate confidence based on TVL, volume, and other factors
	confidence := yfa.calculateOpportunityConfidence(pool)

	// Generate opportunity ID
	opportunityID := fmt.Sprintf("opp_%s_%s_%d", pool.Protocol, pool.Address, time.Now().Unix())

	return &YieldOpportunity{
		ID:                  opportunityID,
		Protocol:            pool.Protocol,
		Chain:               pool.Chain,
		PoolAddress:         pool.Address,
		PoolName:            pool.Name,
		Token0:              pool.Token0,
		Token1:              pool.Token1,
		CurrentAPY:          pool.TotalAPY,
		ProjectedAPY:        pool.TotalAPY.Mul(decimal.NewFromFloat(0.95)), // Conservative estimate
		TVL:                 pool.TVL,
		Volume24h:           pool.Volume24h,
		FeeAPR:              pool.FeeAPR,
		RewardAPR:           pool.RewardAPR,
		RiskScore:           riskScore,
		ImpermanentLossRisk: ilRisk,
		LiquidityDepth:      pool.LiquidityDepth,
		EntrySlippage:       yfa.estimateSlippage(pool, decimal.NewFromFloat(1000)), // Estimate for $1000
		ExitSlippage:        yfa.estimateSlippage(pool, decimal.NewFromFloat(1000)),
		RecommendedAmount:   yfa.calculateRecommendedAmount(pool),
		Confidence:          confidence,
		DetectedAt:          time.Now(),
		ExpiresAt:           time.Now().Add(1 * time.Hour),
		Status:              "detected",
	}
}

// isOpportunityViable checks if an opportunity meets minimum criteria
func (yfa *YieldFarmingAutomation) isOpportunityViable(opportunity *YieldOpportunity) bool {
	// Check minimum APY threshold
	if opportunity.CurrentAPY.LessThan(yfa.config.MinYieldThreshold) {
		return false
	}

	// Check risk level compatibility
	switch yfa.config.RiskLevel {
	case "conservative":
		if opportunity.RiskScore.GreaterThan(decimal.NewFromFloat(0.3)) {
			return false
		}
	case "moderate":
		if opportunity.RiskScore.GreaterThan(decimal.NewFromFloat(0.6)) {
			return false
		}
	case "aggressive":
		// Accept higher risk
	}

	// Check minimum TVL
	if opportunity.TVL.LessThan(decimal.NewFromFloat(100000)) { // $100k minimum TVL
		return false
	}

	// Check slippage tolerance
	if opportunity.EntrySlippage.GreaterThan(yfa.config.MaxSlippageTolerance) {
		return false
	}

	return true
}

// filterOpportunities filters opportunities based on various criteria
func (yfa *YieldFarmingAutomation) filterOpportunities(opportunities []*YieldOpportunity) []*YieldOpportunity {
	var filtered []*YieldOpportunity

	for _, opp := range opportunities {
		// Skip if already have position in this pool
		if yfa.hasActivePosition(opp.PoolAddress) {
			continue
		}

		// Skip if confidence is too low
		if opp.Confidence.LessThan(decimal.NewFromFloat(0.7)) {
			continue
		}

		// Skip if impermanent loss risk is too high
		if opp.ImpermanentLossRisk.GreaterThan(decimal.NewFromFloat(0.1)) && yfa.config.RiskLevel == "conservative" {
			continue
		}

		filtered = append(filtered, opp)
	}

	return filtered
}

// rankOpportunities ranks opportunities by attractiveness
func (yfa *YieldFarmingAutomation) rankOpportunities(opportunities []*YieldOpportunity) []*YieldOpportunity {
	// Sort by a composite score: APY * Confidence / Risk Score
	for i := 0; i < len(opportunities)-1; i++ {
		for j := i + 1; j < len(opportunities); j++ {
			scoreI := yfa.calculateOpportunityScore(opportunities[i])
			scoreJ := yfa.calculateOpportunityScore(opportunities[j])

			if scoreJ.GreaterThan(scoreI) {
				opportunities[i], opportunities[j] = opportunities[j], opportunities[i]
			}
		}
	}

	return opportunities
}

// calculateOptimalTokenAmounts calculates optimal token amounts for deposit
func (yfa *YieldFarmingAutomation) calculateOptimalTokenAmounts(opportunity *YieldOpportunity, totalAmount decimal.Decimal) (decimal.Decimal, decimal.Decimal, error) {
	// Get current prices
	price0 := yfa.getCurrentPrice(opportunity.Token0)
	price1 := yfa.getCurrentPrice(opportunity.Token1)

	if price0.IsZero() || price1.IsZero() {
		return decimal.Zero, decimal.Zero, fmt.Errorf("unable to get current prices")
	}

	// For simplicity, assume 50/50 split (in practice, would use pool ratio)
	value0 := totalAmount.Div(decimal.NewFromFloat(2))
	value1 := totalAmount.Div(decimal.NewFromFloat(2))

	token0Amount := value0.Div(price0)
	token1Amount := value1.Div(price1)

	return token0Amount, token1Amount, nil
}

// generateFarmID generates a unique farm ID
func (yfa *YieldFarmingAutomation) generateFarmID() string {
	return fmt.Sprintf("farm_%d", time.Now().UnixNano())
}

// getCurrentPrice gets current price for a token (mock implementation)
func (yfa *YieldFarmingAutomation) getCurrentPrice(tokenAddress string) decimal.Decimal {
	// Mock prices - in production would fetch from price oracle
	mockPrices := map[string]decimal.Decimal{
		"0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1": decimal.NewFromFloat(1.0),     // USDC
		"0x6B175474E89094C44Da98b954EedeAC495271d0F": decimal.NewFromFloat(1.0),     // DAI
		"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2": decimal.NewFromFloat(2000.0),  // WETH
		"0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599": decimal.NewFromFloat(30000.0), // WBTC
	}

	if price, exists := mockPrices[tokenAddress]; exists {
		return price
	}

	return decimal.NewFromFloat(1.0) // Default price
}

// getCurrentGasPrice gets current gas price (mock implementation)
func (yfa *YieldFarmingAutomation) getCurrentGasPrice() decimal.Decimal {
	// Mock gas price - in production would fetch from gas oracle
	return decimal.NewFromFloat(20000000000) // 20 gwei
}

// calculatePoolRiskScore calculates risk score for a pool
func (yfa *YieldFarmingAutomation) calculatePoolRiskScore(pool *PoolInfo) decimal.Decimal {
	riskScore := decimal.NewFromFloat(0.3) // Base risk

	// Higher risk for lower TVL
	if pool.TVL.LessThan(decimal.NewFromFloat(1000000)) { // < $1M
		riskScore = riskScore.Add(decimal.NewFromFloat(0.2))
	}

	// Higher risk for newer protocols (simplified check)
	if pool.Protocol == "new_protocol" {
		riskScore = riskScore.Add(decimal.NewFromFloat(0.3))
	}

	// Cap at 1.0
	if riskScore.GreaterThan(decimal.NewFromFloat(1.0)) {
		riskScore = decimal.NewFromFloat(1.0)
	}

	return riskScore
}

// calculateImpermanentLossRisk calculates impermanent loss risk
func (yfa *YieldFarmingAutomation) calculateImpermanentLossRisk(pool *PoolInfo) decimal.Decimal {
	// Simplified IL risk calculation based on token volatility
	// In practice, would use historical price correlation and volatility

	// Stablecoin pairs have low IL risk
	stablecoins := map[string]bool{
		"0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1": true, // USDC
		"0x6B175474E89094C44Da98b954EedeAC495271d0F": true, // DAI
		"0xdAC17F958D2ee523a2206206994597C13D831ec7": true, // USDT
	}

	if stablecoins[pool.Token0] && stablecoins[pool.Token1] {
		return decimal.NewFromFloat(0.01) // 1% IL risk for stablecoin pairs
	}

	if stablecoins[pool.Token0] || stablecoins[pool.Token1] {
		return decimal.NewFromFloat(0.05) // 5% IL risk for stable-volatile pairs
	}

	return decimal.NewFromFloat(0.15) // 15% IL risk for volatile-volatile pairs
}

// calculateOpportunityConfidence calculates confidence score for an opportunity
func (yfa *YieldFarmingAutomation) calculateOpportunityConfidence(pool *PoolInfo) decimal.Decimal {
	confidence := decimal.NewFromFloat(0.5) // Base confidence

	// Higher confidence for higher TVL
	if pool.TVL.GreaterThan(decimal.NewFromFloat(10000000)) { // > $10M
		confidence = confidence.Add(decimal.NewFromFloat(0.2))
	} else if pool.TVL.GreaterThan(decimal.NewFromFloat(1000000)) { // > $1M
		confidence = confidence.Add(decimal.NewFromFloat(0.1))
	}

	// Higher confidence for higher volume
	if pool.Volume24h.GreaterThan(decimal.NewFromFloat(1000000)) { // > $1M daily volume
		confidence = confidence.Add(decimal.NewFromFloat(0.1))
	}

	// Higher confidence for established protocols
	establishedProtocols := map[string]bool{
		"uniswap":   true,
		"sushiswap": true,
		"curve":     true,
		"balancer":  true,
	}

	if establishedProtocols[pool.Protocol] {
		confidence = confidence.Add(decimal.NewFromFloat(0.2))
	}

	// Cap at 1.0
	if confidence.GreaterThan(decimal.NewFromFloat(1.0)) {
		confidence = decimal.NewFromFloat(1.0)
	}

	return confidence
}

// estimateSlippage estimates slippage for a given trade amount
func (yfa *YieldFarmingAutomation) estimateSlippage(pool *PoolInfo, amount decimal.Decimal) decimal.Decimal {
	// Simplified slippage calculation based on liquidity depth
	if pool.LiquidityDepth.IsZero() {
		return decimal.NewFromFloat(0.05) // 5% default slippage
	}

	// Slippage roughly proportional to amount / liquidity
	slippage := amount.Div(pool.LiquidityDepth).Mul(decimal.NewFromFloat(0.1))

	// Minimum 0.1% slippage
	if slippage.LessThan(decimal.NewFromFloat(0.001)) {
		slippage = decimal.NewFromFloat(0.001)
	}

	// Maximum 10% slippage
	if slippage.GreaterThan(decimal.NewFromFloat(0.1)) {
		slippage = decimal.NewFromFloat(0.1)
	}

	return slippage
}

// calculateRecommendedAmount calculates recommended investment amount
func (yfa *YieldFarmingAutomation) calculateRecommendedAmount(pool *PoolInfo) decimal.Decimal {
	// Base recommendation on TVL and risk
	baseAmount := decimal.NewFromFloat(1000) // $1000 base

	// Adjust based on TVL
	if pool.TVL.GreaterThan(decimal.NewFromFloat(10000000)) { // > $10M TVL
		baseAmount = baseAmount.Mul(decimal.NewFromFloat(5)) // Up to $5000
	} else if pool.TVL.GreaterThan(decimal.NewFromFloat(1000000)) { // > $1M TVL
		baseAmount = baseAmount.Mul(decimal.NewFromFloat(2)) // Up to $2000
	}

	// Adjust based on APY
	if pool.TotalAPY.GreaterThan(decimal.NewFromFloat(0.5)) { // > 50% APY
		baseAmount = baseAmount.Mul(decimal.NewFromFloat(1.5))
	}

	// Cap at max position size
	if baseAmount.GreaterThan(yfa.config.MaxPositionSize) {
		baseAmount = yfa.config.MaxPositionSize
	}

	return baseAmount
}

// hasActivePosition checks if there's an active position in a pool
func (yfa *YieldFarmingAutomation) hasActivePosition(poolAddress string) bool {
	yfa.mutex.RLock()
	defer yfa.mutex.RUnlock()

	for _, farm := range yfa.activeFarms {
		if farm.PoolAddress == poolAddress && farm.Status == "active" {
			return true
		}
	}

	return false
}

// calculateOpportunityScore calculates composite score for ranking
func (yfa *YieldFarmingAutomation) calculateOpportunityScore(opportunity *YieldOpportunity) decimal.Decimal {
	// Score = APY * Confidence / (Risk Score + 0.1)
	score := opportunity.CurrentAPY.Mul(opportunity.Confidence).Div(
		opportunity.RiskScore.Add(decimal.NewFromFloat(0.1)))

	return score
}

// updateMetricsAfterEntry updates metrics after entering a farm
func (yfa *YieldFarmingAutomation) updateMetricsAfterEntry(farm *ActiveFarm) {
	yfa.mutex.Lock()
	defer yfa.mutex.Unlock()

	yfa.performanceMetrics.TotalValueLocked = yfa.performanceMetrics.TotalValueLocked.Add(farm.TotalValue)
	yfa.performanceMetrics.ActiveFarmsCount++

	// Update protocol distribution
	if yfa.performanceMetrics.ProtocolDistribution == nil {
		yfa.performanceMetrics.ProtocolDistribution = make(map[string]decimal.Decimal)
	}
	yfa.performanceMetrics.ProtocolDistribution[farm.Protocol] =
		yfa.performanceMetrics.ProtocolDistribution[farm.Protocol].Add(farm.TotalValue)

	// Update chain distribution
	if yfa.performanceMetrics.ChainDistribution == nil {
		yfa.performanceMetrics.ChainDistribution = make(map[string]decimal.Decimal)
	}
	yfa.performanceMetrics.ChainDistribution[farm.Chain] =
		yfa.performanceMetrics.ChainDistribution[farm.Chain].Add(farm.TotalValue)

	yfa.performanceMetrics.LastUpdated = time.Now()
}

// updateMetricsAfterExit updates metrics after exiting a farm
func (yfa *YieldFarmingAutomation) updateMetricsAfterExit(farm *ActiveFarm, percentage decimal.Decimal) {
	yfa.mutex.Lock()
	defer yfa.mutex.Unlock()

	exitValue := farm.TotalValue.Mul(percentage)
	yfa.performanceMetrics.TotalValueLocked = yfa.performanceMetrics.TotalValueLocked.Sub(exitValue)

	if percentage.Equal(decimal.NewFromFloat(1.0)) {
		yfa.performanceMetrics.ActiveFarmsCount--
	}

	// Update protocol distribution
	yfa.performanceMetrics.ProtocolDistribution[farm.Protocol] =
		yfa.performanceMetrics.ProtocolDistribution[farm.Protocol].Sub(exitValue)

	// Update chain distribution
	yfa.performanceMetrics.ChainDistribution[farm.Chain] =
		yfa.performanceMetrics.ChainDistribution[farm.Chain].Sub(exitValue)

	yfa.performanceMetrics.LastUpdated = time.Now()
}

// updateMetricsAfterCompounding updates metrics after compounding
func (yfa *YieldFarmingAutomation) updateMetricsAfterCompounding(farm *ActiveFarm, compoundedAmount decimal.Decimal) {
	yfa.mutex.Lock()
	defer yfa.mutex.Unlock()

	yfa.performanceMetrics.TotalRewardsEarned = yfa.performanceMetrics.TotalRewardsEarned.Add(compoundedAmount)
	yfa.performanceMetrics.CompletedCompounds++
	yfa.performanceMetrics.LastUpdated = time.Now()
}

// Automation loop methods

// initializeProtocolClients initializes protocol clients
func (yfa *YieldFarmingAutomation) initializeProtocolClients() error {
	yfa.logger.Info("Initializing protocol clients")

	// Initialize mock protocol clients for supported protocols
	for _, protocol := range yfa.config.SupportedProtocols {
		client := NewMockYieldProtocolClient(protocol, yfa.logger)
		yfa.protocolClients[protocol] = client
	}

	yfa.logger.Info("Protocol clients initialized",
		zap.Int("client_count", len(yfa.protocolClients)))

	return nil
}

// loadExistingData loads existing farms and strategies
func (yfa *YieldFarmingAutomation) loadExistingData(ctx context.Context) error {
	yfa.logger.Info("Loading existing yield farming data")

	// In production, would load from database
	// For now, create default strategies
	yfa.createDefaultStrategies()

	yfa.logger.Info("Existing data loaded")
	return nil
}

// createDefaultStrategies creates default farming strategies
func (yfa *YieldFarmingAutomation) createDefaultStrategies() {
	strategies := []*FarmingStrategy{
		{
			ID:                   "conservative_stable",
			Name:                 "Conservative Stablecoin Strategy",
			Description:          "Low-risk strategy focusing on stablecoin pairs",
			TargetAPY:            decimal.NewFromFloat(0.08), // 8%
			MaxImpermanentLoss:   decimal.NewFromFloat(0.02), // 2%
			PreferredProtocols:   []string{"curve", "aave"},
			PreferredChains:      []string{"ethereum", "polygon"},
			RiskLevel:            "conservative",
			CompoundingFrequency: 24 * time.Hour,
			RebalanceThreshold:   decimal.NewFromFloat(0.05),
			ExitConditions: []ExitCondition{
				{Type: "apy_drop", Threshold: decimal.NewFromFloat(0.05), Enabled: true},
				{Type: "impermanent_loss", Threshold: decimal.NewFromFloat(0.03), Enabled: true},
			},
			AllocationPercentage: decimal.NewFromFloat(0.4), // 40%
			IsActive:             true,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		},
		{
			ID:                   "moderate_mixed",
			Name:                 "Moderate Mixed Strategy",
			Description:          "Balanced strategy with mixed asset pairs",
			TargetAPY:            decimal.NewFromFloat(0.15), // 15%
			MaxImpermanentLoss:   decimal.NewFromFloat(0.05), // 5%
			PreferredProtocols:   []string{"uniswap", "sushiswap"},
			PreferredChains:      []string{"ethereum", "arbitrum"},
			RiskLevel:            "moderate",
			CompoundingFrequency: 12 * time.Hour,
			RebalanceThreshold:   decimal.NewFromFloat(0.1),
			ExitConditions: []ExitCondition{
				{Type: "apy_drop", Threshold: decimal.NewFromFloat(0.08), Enabled: true},
				{Type: "impermanent_loss", Threshold: decimal.NewFromFloat(0.08), Enabled: true},
			},
			AllocationPercentage: decimal.NewFromFloat(0.4), // 40%
			IsActive:             true,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		},
		{
			ID:                   "aggressive_high_yield",
			Name:                 "Aggressive High-Yield Strategy",
			Description:          "High-risk, high-reward strategy for maximum yields",
			TargetAPY:            decimal.NewFromFloat(0.3),  // 30%
			MaxImpermanentLoss:   decimal.NewFromFloat(0.15), // 15%
			PreferredProtocols:   []string{"pancakeswap", "quickswap"},
			PreferredChains:      []string{"bsc", "polygon"},
			RiskLevel:            "aggressive",
			CompoundingFrequency: 6 * time.Hour,
			RebalanceThreshold:   decimal.NewFromFloat(0.15),
			ExitConditions: []ExitCondition{
				{Type: "apy_drop", Threshold: decimal.NewFromFloat(0.15), Enabled: true},
				{Type: "impermanent_loss", Threshold: decimal.NewFromFloat(0.2), Enabled: true},
			},
			AllocationPercentage: decimal.NewFromFloat(0.2), // 20%
			IsActive:             true,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		},
	}

	yfa.mutex.Lock()
	for _, strategy := range strategies {
		yfa.farmingStrategies[strategy.ID] = strategy
	}
	yfa.mutex.Unlock()
}

// opportunityDetectionLoop continuously scans for yield opportunities
func (yfa *YieldFarmingAutomation) opportunityDetectionLoop(ctx context.Context) {
	ticker := time.NewTicker(yfa.config.OpportunityCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-yfa.stopChan:
			return
		case <-ticker.C:
			_, err := yfa.ScanForYieldOpportunities(ctx)
			if err != nil {
				yfa.logger.Error("Failed to scan for opportunities", zap.Error(err))
			}
		}
	}
}

// autoCompoundingLoop handles automatic compounding of rewards
func (yfa *YieldFarmingAutomation) autoCompoundingLoop(ctx context.Context) {
	if !yfa.config.AutoCompoundingEnabled {
		return
	}

	ticker := time.NewTicker(yfa.config.CompoundingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-yfa.stopChan:
			return
		case <-ticker.C:
			yfa.performAutoCompounding(ctx)
		}
	}
}

// poolMigrationLoop handles automatic pool migration
func (yfa *YieldFarmingAutomation) poolMigrationLoop(ctx context.Context) {
	if !yfa.config.PoolMigrationEnabled {
		return
	}

	ticker := time.NewTicker(1 * time.Hour) // Check every hour
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-yfa.stopChan:
			return
		case <-ticker.C:
			yfa.evaluatePoolMigrations(ctx)
		}
	}
}

// impermanentLossMonitoringLoop monitors impermanent loss
func (yfa *YieldFarmingAutomation) impermanentLossMonitoringLoop(ctx context.Context) {
	if !yfa.config.ImpermanentLossProtection {
		return
	}

	ticker := time.NewTicker(yfa.ilProtection.MonitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-yfa.stopChan:
			return
		case <-ticker.C:
			yfa.monitorImpermanentLoss(ctx)
		}
	}
}

// performanceTrackingLoop tracks and updates performance metrics
func (yfa *YieldFarmingAutomation) performanceTrackingLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-yfa.stopChan:
			return
		case <-ticker.C:
			yfa.updatePerformanceMetrics(ctx)
		}
	}
}

// strategyExecutionLoop executes farming strategies
func (yfa *YieldFarmingAutomation) strategyExecutionLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-yfa.stopChan:
			return
		case <-ticker.C:
			yfa.executeStrategies(ctx)
		}
	}
}

// Loop implementation methods

// performAutoCompounding performs automatic compounding for eligible farms
func (yfa *YieldFarmingAutomation) performAutoCompounding(ctx context.Context) {
	yfa.logger.Debug("Performing auto-compounding")

	yfa.mutex.RLock()
	farms := make([]*ActiveFarm, 0, len(yfa.activeFarms))
	for _, farm := range yfa.activeFarms {
		if farm.AutoCompounding && farm.Status == "active" {
			farms = append(farms, farm)
		}
	}
	yfa.mutex.RUnlock()

	for _, farm := range farms {
		// Check if enough time has passed since last compound
		strategy, exists := yfa.farmingStrategies[farm.Strategy]
		if !exists {
			continue
		}

		if time.Since(farm.LastCompounded) < strategy.CompoundingFrequency {
			continue
		}

		// Check if rewards are sufficient for compounding
		if farm.RewardsEarned.LessThan(yfa.config.MinCompoundingAmount) {
			continue
		}

		// Perform compounding
		_, err := yfa.CompoundRewards(ctx, farm.ID)
		if err != nil {
			yfa.logger.Error("Failed to compound rewards",
				zap.String("farm_id", farm.ID),
				zap.Error(err))
		}
	}
}

// evaluatePoolMigrations evaluates and executes pool migrations
func (yfa *YieldFarmingAutomation) evaluatePoolMigrations(ctx context.Context) {
	yfa.logger.Debug("Evaluating pool migrations")

	yfa.mutex.RLock()
	farms := make([]*ActiveFarm, 0, len(yfa.activeFarms))
	for _, farm := range yfa.activeFarms {
		if farm.Status == "active" {
			farms = append(farms, farm)
		}
	}
	opportunities := make([]*YieldOpportunity, 0, len(yfa.yieldOpportunities))
	for _, opp := range yfa.yieldOpportunities {
		if opp.Status == "recommended" {
			opportunities = append(opportunities, opp)
		}
	}
	yfa.mutex.RUnlock()

	for _, farm := range farms {
		// Find better opportunities
		for _, opp := range opportunities {
			// Skip if same pool
			if farm.PoolAddress == opp.PoolAddress {
				continue
			}

			// Check if migration is beneficial
			apyImprovement := opp.CurrentAPY.Sub(farm.CurrentAPY)
			if apyImprovement.LessThan(decimal.NewFromFloat(0.02)) { // Minimum 2% improvement
				continue
			}

			// Estimate migration cost
			migrationCost := yfa.estimateMigrationCost(farm, opp)

			// Calculate expected benefit
			expectedBenefit := farm.TotalValue.Mul(apyImprovement).Div(decimal.NewFromFloat(365)) // Daily benefit

			// Check if migration is profitable (payback period < 30 days)
			if migrationCost.GreaterThan(expectedBenefit.Mul(decimal.NewFromFloat(30))) {
				continue
			}

			// Create migration task
			migrationTask := &MigrationTask{
				ID:              yfa.generateMigrationID(),
				FromFarmID:      farm.ID,
				ToOpportunityID: opp.ID,
				FromProtocol:    farm.Protocol,
				ToProtocol:      opp.Protocol,
				FromPool:        farm.PoolAddress,
				ToPool:          opp.PoolAddress,
				Amount:          farm.TotalValue,
				CurrentAPY:      farm.CurrentAPY,
				TargetAPY:       opp.CurrentAPY,
				ExpectedGain:    expectedBenefit.Mul(decimal.NewFromFloat(365)), // Annual gain
				EstimatedCost:   migrationCost,
				NetBenefit:      expectedBenefit.Mul(decimal.NewFromFloat(365)).Sub(migrationCost),
				ScheduledAt:     time.Now().Add(5 * time.Minute), // Schedule for 5 minutes
				Status:          "scheduled",
			}

			yfa.mutex.Lock()
			yfa.migrationTasks[migrationTask.ID] = migrationTask
			yfa.mutex.Unlock()

			yfa.logger.Info("Migration task scheduled",
				zap.String("task_id", migrationTask.ID),
				zap.String("from_pool", farm.PoolAddress),
				zap.String("to_pool", opp.PoolAddress),
				zap.String("expected_gain", migrationTask.ExpectedGain.String()))

			break // Only migrate to one better opportunity per farm
		}
	}
}

// monitorImpermanentLoss monitors and manages impermanent loss
func (yfa *YieldFarmingAutomation) monitorImpermanentLoss(ctx context.Context) {
	yfa.logger.Debug("Monitoring impermanent loss")

	yfa.mutex.RLock()
	farms := make([]*ActiveFarm, 0, len(yfa.activeFarms))
	for _, farm := range yfa.activeFarms {
		if farm.Status == "active" {
			farms = append(farms, farm)
		}
	}
	yfa.mutex.RUnlock()

	for _, farm := range farms {
		// Calculate current impermanent loss
		currentIL := yfa.calculateCurrentImpermanentLoss(farm)

		yfa.mutex.Lock()
		farm.ImpermanentLoss = currentIL
		yfa.mutex.Unlock()

		// Check if IL exceeds threshold
		if currentIL.GreaterThan(yfa.ilProtection.MaxAcceptableLoss) {
			yfa.logger.Warn("High impermanent loss detected",
				zap.String("farm_id", farm.ID),
				zap.String("il_percentage", currentIL.Mul(decimal.NewFromFloat(100)).String()))

			// Take protective action based on configuration
			if yfa.ilProtection.AutoExitEnabled && currentIL.GreaterThan(yfa.ilProtection.AutoExitThreshold) {
				// Auto-exit the position
				_, err := yfa.ExitFarm(ctx, farm.ID, decimal.NewFromFloat(1.0))
				if err != nil {
					yfa.logger.Error("Failed to auto-exit farm due to IL",
						zap.String("farm_id", farm.ID),
						zap.Error(err))
				} else {
					yfa.logger.Info("Auto-exited farm due to high impermanent loss",
						zap.String("farm_id", farm.ID))
				}
			} else if yfa.ilProtection.RebalancingEnabled && currentIL.GreaterThan(yfa.ilProtection.RebalancingThreshold) {
				// Rebalance the position (simplified implementation)
				yfa.logger.Info("Rebalancing recommended due to impermanent loss",
					zap.String("farm_id", farm.ID))
			}
		}
	}
}

// updatePerformanceMetrics updates overall performance metrics
func (yfa *YieldFarmingAutomation) updatePerformanceMetrics(ctx context.Context) {
	yfa.logger.Debug("Updating performance metrics")

	yfa.mutex.Lock()
	defer yfa.mutex.Unlock()

	// Reset metrics
	yfa.performanceMetrics.TotalValueLocked = decimal.Zero
	yfa.performanceMetrics.TotalRewardsEarned = decimal.Zero
	yfa.performanceMetrics.TotalImpermanentLoss = decimal.Zero
	yfa.performanceMetrics.NetProfit = decimal.Zero
	yfa.performanceMetrics.ActiveFarmsCount = 0

	totalAPY := decimal.Zero
	farmCount := 0

	// Calculate metrics from active farms
	for _, farm := range yfa.activeFarms {
		if farm.Status == "active" {
			yfa.performanceMetrics.TotalValueLocked = yfa.performanceMetrics.TotalValueLocked.Add(farm.TotalValue)
			yfa.performanceMetrics.TotalRewardsEarned = yfa.performanceMetrics.TotalRewardsEarned.Add(farm.RewardsEarned)
			yfa.performanceMetrics.TotalImpermanentLoss = yfa.performanceMetrics.TotalImpermanentLoss.Add(farm.ImpermanentLoss)
			yfa.performanceMetrics.NetProfit = yfa.performanceMetrics.NetProfit.Add(farm.ProfitLoss)
			yfa.performanceMetrics.ActiveFarmsCount++

			totalAPY = totalAPY.Add(farm.CurrentAPY)
			farmCount++
		}
	}

	// Calculate average APY
	if farmCount > 0 {
		yfa.performanceMetrics.AverageAPY = totalAPY.Div(decimal.NewFromInt(int64(farmCount)))
	}

	yfa.performanceMetrics.LastUpdated = time.Now()
}

// executeStrategies executes farming strategies
func (yfa *YieldFarmingAutomation) executeStrategies(ctx context.Context) {
	yfa.logger.Debug("Executing farming strategies")

	yfa.mutex.RLock()
	strategies := make([]*FarmingStrategy, 0, len(yfa.farmingStrategies))
	for _, strategy := range yfa.farmingStrategies {
		if strategy.IsActive {
			strategies = append(strategies, strategy)
		}
	}
	opportunities := make([]*YieldOpportunity, 0, len(yfa.yieldOpportunities))
	for _, opp := range yfa.yieldOpportunities {
		if opp.Status == "recommended" {
			opportunities = append(opportunities, opp)
		}
	}
	yfa.mutex.RUnlock()

	for _, strategy := range strategies {
		// Find opportunities that match strategy criteria
		matchingOpportunities := yfa.findMatchingOpportunities(strategy, opportunities)

		if len(matchingOpportunities) == 0 {
			continue
		}

		// Calculate available allocation for this strategy
		availableAllocation := yfa.calculateAvailableAllocation(strategy)

		if availableAllocation.LessThan(decimal.NewFromFloat(100)) { // Minimum $100
			continue
		}

		// Select best opportunity
		bestOpportunity := matchingOpportunities[0] // Already ranked

		// Calculate position size
		positionSize := availableAllocation.Mul(strategy.AllocationPercentage)
		if positionSize.GreaterThan(bestOpportunity.RecommendedAmount) {
			positionSize = bestOpportunity.RecommendedAmount
		}

		// Enter farm
		_, err := yfa.EnterFarm(ctx, bestOpportunity.ID, positionSize, strategy.ID)
		if err != nil {
			yfa.logger.Error("Failed to enter farm via strategy",
				zap.String("strategy_id", strategy.ID),
				zap.String("opportunity_id", bestOpportunity.ID),
				zap.Error(err))
		} else {
			yfa.logger.Info("Entered farm via strategy",
				zap.String("strategy_id", strategy.ID),
				zap.String("opportunity_id", bestOpportunity.ID),
				zap.String("position_size", positionSize.String()))
		}
	}
}

// Additional helper methods

// estimateMigrationCost estimates the cost of migrating between pools
func (yfa *YieldFarmingAutomation) estimateMigrationCost(farm *ActiveFarm, opportunity *YieldOpportunity) decimal.Decimal {
	// Estimate gas costs for exit + entry
	exitGasCost := decimal.NewFromFloat(120000).Mul(yfa.getCurrentGasPrice())  // Exit gas
	entryGasCost := decimal.NewFromFloat(150000).Mul(yfa.getCurrentGasPrice()) // Entry gas

	// Add slippage costs
	exitSlippage := farm.TotalValue.Mul(decimal.NewFromFloat(0.005)) // 0.5% exit slippage
	entrySlippage := farm.TotalValue.Mul(opportunity.EntrySlippage)

	totalCost := exitGasCost.Add(entryGasCost).Add(exitSlippage).Add(entrySlippage)
	return totalCost
}

// generateMigrationID generates a unique migration task ID
func (yfa *YieldFarmingAutomation) generateMigrationID() string {
	return fmt.Sprintf("migration_%d", time.Now().UnixNano())
}

// calculateCurrentImpermanentLoss calculates current impermanent loss for a farm
func (yfa *YieldFarmingAutomation) calculateCurrentImpermanentLoss(farm *ActiveFarm) decimal.Decimal {
	// Get current prices
	currentPrice0 := yfa.getCurrentPrice(farm.Token0)
	currentPrice1 := yfa.getCurrentPrice(farm.Token1)

	if currentPrice0.IsZero() || currentPrice1.IsZero() || farm.EntryPrice0.IsZero() || farm.EntryPrice1.IsZero() {
		return decimal.Zero
	}

	// Calculate price ratio change
	entryRatio := farm.EntryPrice0.Div(farm.EntryPrice1)
	currentRatio := currentPrice0.Div(currentPrice1)

	// Simplified IL calculation: IL = 2 * sqrt(ratio) / (1 + ratio) - 1
	// where ratio = current_price_ratio / entry_price_ratio
	ratio := currentRatio.Div(entryRatio)

	// For simplicity, use approximation: IL ≈ (ratio - 1)² / (4 * ratio)
	if ratio.Equal(decimal.NewFromFloat(1.0)) {
		return decimal.Zero
	}

	ratioMinusOne := ratio.Sub(decimal.NewFromFloat(1.0))
	il := ratioMinusOne.Mul(ratioMinusOne).Div(ratio.Mul(decimal.NewFromFloat(4)))

	// Cap IL at 50% (extreme case)
	if il.GreaterThan(decimal.NewFromFloat(0.5)) {
		il = decimal.NewFromFloat(0.5)
	}

	return il
}

// findMatchingOpportunities finds opportunities that match strategy criteria
func (yfa *YieldFarmingAutomation) findMatchingOpportunities(strategy *FarmingStrategy, opportunities []*YieldOpportunity) []*YieldOpportunity {
	var matching []*YieldOpportunity

	for _, opp := range opportunities {
		// Check APY threshold
		if opp.CurrentAPY.LessThan(strategy.TargetAPY) {
			continue
		}

		// Check impermanent loss threshold
		if opp.ImpermanentLossRisk.GreaterThan(strategy.MaxImpermanentLoss) {
			continue
		}

		// Check preferred protocols
		if len(strategy.PreferredProtocols) > 0 {
			protocolMatch := false
			for _, protocol := range strategy.PreferredProtocols {
				if opp.Protocol == protocol {
					protocolMatch = true
					break
				}
			}
			if !protocolMatch {
				continue
			}
		}

		// Check preferred chains
		if len(strategy.PreferredChains) > 0 {
			chainMatch := false
			for _, chain := range strategy.PreferredChains {
				if opp.Chain == chain {
					chainMatch = true
					break
				}
			}
			if !chainMatch {
				continue
			}
		}

		// Check risk level compatibility
		if !yfa.isRiskLevelCompatible(strategy.RiskLevel, opp.RiskScore) {
			continue
		}

		matching = append(matching, opp)
	}

	// Sort by opportunity score (already implemented)
	for i := 0; i < len(matching)-1; i++ {
		for j := i + 1; j < len(matching); j++ {
			scoreI := yfa.calculateOpportunityScore(matching[i])
			scoreJ := yfa.calculateOpportunityScore(matching[j])

			if scoreJ.GreaterThan(scoreI) {
				matching[i], matching[j] = matching[j], matching[i]
			}
		}
	}

	return matching
}

// calculateAvailableAllocation calculates available allocation for a strategy
func (yfa *YieldFarmingAutomation) calculateAvailableAllocation(strategy *FarmingStrategy) decimal.Decimal {
	// Mock available capital - in production would check actual wallet balance
	totalAvailableCapital := decimal.NewFromFloat(10000) // $10k available

	// Calculate current allocation for this strategy
	currentAllocation := decimal.Zero
	for _, farm := range yfa.activeFarms {
		if farm.Strategy == strategy.ID && farm.Status == "active" {
			currentAllocation = currentAllocation.Add(farm.TotalValue)
		}
	}

	// Calculate target allocation
	targetAllocation := totalAvailableCapital.Mul(strategy.AllocationPercentage)

	// Available allocation is target minus current
	availableAllocation := targetAllocation.Sub(currentAllocation)

	if availableAllocation.LessThan(decimal.Zero) {
		availableAllocation = decimal.Zero
	}

	return availableAllocation
}

// isRiskLevelCompatible checks if risk level is compatible with strategy
func (yfa *YieldFarmingAutomation) isRiskLevelCompatible(strategyRiskLevel string, opportunityRiskScore decimal.Decimal) bool {
	switch strategyRiskLevel {
	case "conservative":
		return opportunityRiskScore.LessThanOrEqual(decimal.NewFromFloat(0.3))
	case "moderate":
		return opportunityRiskScore.LessThanOrEqual(decimal.NewFromFloat(0.6))
	case "aggressive":
		return opportunityRiskScore.LessThanOrEqual(decimal.NewFromFloat(1.0))
	default:
		return true
	}
}

// Public interface methods

// GetActiveFarms returns all active farms
func (yfa *YieldFarmingAutomation) GetActiveFarms() []*ActiveFarm {
	yfa.mutex.RLock()
	defer yfa.mutex.RUnlock()

	farms := make([]*ActiveFarm, 0, len(yfa.activeFarms))
	for _, farm := range yfa.activeFarms {
		farms = append(farms, farm)
	}

	return farms
}

// GetActiveFarm returns a specific active farm
func (yfa *YieldFarmingAutomation) GetActiveFarm(farmID string) (*ActiveFarm, error) {
	yfa.mutex.RLock()
	defer yfa.mutex.RUnlock()

	farm, exists := yfa.activeFarms[farmID]
	if !exists {
		return nil, fmt.Errorf("active farm not found: %s", farmID)
	}

	return farm, nil
}

// GetYieldOpportunities returns all detected yield opportunities
func (yfa *YieldFarmingAutomation) GetYieldOpportunities() []*YieldOpportunity {
	yfa.mutex.RLock()
	defer yfa.mutex.RUnlock()

	opportunities := make([]*YieldOpportunity, 0, len(yfa.yieldOpportunities))
	for _, opp := range yfa.yieldOpportunities {
		opportunities = append(opportunities, opp)
	}

	return opportunities
}

// GetFarmingStrategies returns all farming strategies
func (yfa *YieldFarmingAutomation) GetFarmingStrategies() []*FarmingStrategy {
	yfa.mutex.RLock()
	defer yfa.mutex.RUnlock()

	strategies := make([]*FarmingStrategy, 0, len(yfa.farmingStrategies))
	for _, strategy := range yfa.farmingStrategies {
		strategies = append(strategies, strategy)
	}

	return strategies
}

// GetFarmingStrategy returns a specific farming strategy
func (yfa *YieldFarmingAutomation) GetFarmingStrategy(strategyID string) (*FarmingStrategy, error) {
	yfa.mutex.RLock()
	defer yfa.mutex.RUnlock()

	strategy, exists := yfa.farmingStrategies[strategyID]
	if !exists {
		return nil, fmt.Errorf("farming strategy not found: %s", strategyID)
	}

	return strategy, nil
}

// GetCompoundingTasks returns all compounding tasks
func (yfa *YieldFarmingAutomation) GetCompoundingTasks() []*CompoundingTask {
	yfa.mutex.RLock()
	defer yfa.mutex.RUnlock()

	tasks := make([]*CompoundingTask, 0, len(yfa.compoundingTasks))
	for _, task := range yfa.compoundingTasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// GetMigrationTasks returns all migration tasks
func (yfa *YieldFarmingAutomation) GetMigrationTasks() []*MigrationTask {
	yfa.mutex.RLock()
	defer yfa.mutex.RUnlock()

	tasks := make([]*MigrationTask, 0, len(yfa.migrationTasks))
	for _, task := range yfa.migrationTasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// GetPerformanceMetrics returns current performance metrics
func (yfa *YieldFarmingAutomation) GetPerformanceMetrics() *YieldFarmingMetrics {
	yfa.mutex.RLock()
	defer yfa.mutex.RUnlock()

	return yfa.performanceMetrics
}

// GetImpermanentLossProtection returns IL protection settings
func (yfa *YieldFarmingAutomation) GetImpermanentLossProtection() *ImpermanentLossProtection {
	yfa.mutex.RLock()
	defer yfa.mutex.RUnlock()

	return yfa.ilProtection
}

// UpdateConfig updates the yield farming configuration
func (yfa *YieldFarmingAutomation) UpdateConfig(config YieldFarmingConfig) error {
	yfa.mutex.Lock()
	defer yfa.mutex.Unlock()

	yfa.config = config
	yfa.logger.Info("Yield farming configuration updated",
		zap.Bool("auto_compounding", config.AutoCompoundingEnabled),
		zap.Bool("pool_migration", config.PoolMigrationEnabled),
		zap.Bool("il_protection", config.ImpermanentLossProtection),
		zap.String("min_yield_threshold", config.MinYieldThreshold.String()),
		zap.Strings("supported_protocols", config.SupportedProtocols))

	return nil
}

// GetConfig returns the current configuration
func (yfa *YieldFarmingAutomation) GetConfig() YieldFarmingConfig {
	yfa.mutex.RLock()
	defer yfa.mutex.RUnlock()

	return yfa.config
}

// IsRunning returns whether the yield farming automation is running
func (yfa *YieldFarmingAutomation) IsRunning() bool {
	yfa.mutex.RLock()
	defer yfa.mutex.RUnlock()

	return yfa.isRunning
}

// CreateFarmingStrategy creates a new farming strategy
func (yfa *YieldFarmingAutomation) CreateFarmingStrategy(strategy *FarmingStrategy) error {
	yfa.mutex.Lock()
	defer yfa.mutex.Unlock()

	if strategy.ID == "" {
		strategy.ID = fmt.Sprintf("strategy_%d", time.Now().UnixNano())
	}

	strategy.CreatedAt = time.Now()
	strategy.UpdatedAt = time.Now()

	yfa.farmingStrategies[strategy.ID] = strategy

	yfa.logger.Info("Farming strategy created",
		zap.String("strategy_id", strategy.ID),
		zap.String("name", strategy.Name),
		zap.String("risk_level", strategy.RiskLevel))

	return nil
}

// UpdateFarmingStrategy updates an existing farming strategy
func (yfa *YieldFarmingAutomation) UpdateFarmingStrategy(strategyID string, updates *FarmingStrategy) error {
	yfa.mutex.Lock()
	defer yfa.mutex.Unlock()

	strategy, exists := yfa.farmingStrategies[strategyID]
	if !exists {
		return fmt.Errorf("farming strategy not found: %s", strategyID)
	}

	// Update fields
	if updates.Name != "" {
		strategy.Name = updates.Name
	}
	if updates.Description != "" {
		strategy.Description = updates.Description
	}
	if !updates.TargetAPY.IsZero() {
		strategy.TargetAPY = updates.TargetAPY
	}
	if !updates.MaxImpermanentLoss.IsZero() {
		strategy.MaxImpermanentLoss = updates.MaxImpermanentLoss
	}
	if len(updates.PreferredProtocols) > 0 {
		strategy.PreferredProtocols = updates.PreferredProtocols
	}
	if len(updates.PreferredChains) > 0 {
		strategy.PreferredChains = updates.PreferredChains
	}
	if updates.RiskLevel != "" {
		strategy.RiskLevel = updates.RiskLevel
	}
	if updates.CompoundingFrequency > 0 {
		strategy.CompoundingFrequency = updates.CompoundingFrequency
	}
	if !updates.AllocationPercentage.IsZero() {
		strategy.AllocationPercentage = updates.AllocationPercentage
	}

	strategy.UpdatedAt = time.Now()

	yfa.logger.Info("Farming strategy updated",
		zap.String("strategy_id", strategyID))

	return nil
}

// DeleteFarmingStrategy deletes a farming strategy
func (yfa *YieldFarmingAutomation) DeleteFarmingStrategy(strategyID string) error {
	yfa.mutex.Lock()
	defer yfa.mutex.Unlock()

	if _, exists := yfa.farmingStrategies[strategyID]; !exists {
		return fmt.Errorf("farming strategy not found: %s", strategyID)
	}

	delete(yfa.farmingStrategies, strategyID)

	yfa.logger.Info("Farming strategy deleted",
		zap.String("strategy_id", strategyID))

	return nil
}

// GetDefaultYieldFarmingConfig returns default yield farming configuration
func GetDefaultYieldFarmingConfig() YieldFarmingConfig {
	return YieldFarmingConfig{
		Enabled:                   true,
		AutoCompoundingEnabled:    true,
		PoolMigrationEnabled:      true,
		ImpermanentLossProtection: true,
		MinYieldThreshold:         decimal.NewFromFloat(0.05),  // 5% minimum APY
		MaxSlippageTolerance:      decimal.NewFromFloat(0.01),  // 1% max slippage
		CompoundingInterval:       12 * time.Hour,              // Compound every 12 hours
		OpportunityCheckInterval:  5 * time.Minute,             // Check opportunities every 5 minutes
		MaxPositionSize:           decimal.NewFromFloat(10000), // $10k max position
		MinCompoundingAmount:      decimal.NewFromFloat(10),    // $10 minimum to compound
		GasOptimizationEnabled:    true,
		MaxGasPriceGwei:           decimal.NewFromFloat(100), // 100 gwei max gas price
		SupportedProtocols:        []string{"uniswap", "sushiswap", "curve", "balancer"},
		SupportedChains:           []string{"ethereum", "polygon", "arbitrum", "optimism"},
		RiskLevel:                 "moderate",
	}
}
