package defi

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/redis"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// YieldAggregator finds and optimizes yield farming opportunities
type YieldAggregator struct {
	logger        *logger.Logger
	cache         redis.Client
	uniswapClient *UniswapClient
	aaveClient    *AaveClient

	// Configuration
	minAPY       decimal.Decimal
	maxRisk      RiskLevel
	scanInterval time.Duration

	// State
	opportunities map[string]*YieldFarmingOpportunity
	strategies    map[string]*YieldStrategy
	mutex         sync.RWMutex

	// Channels
	opportunityChan chan *YieldFarmingOpportunity
	stopChan        chan struct{}
}



// NewYieldAggregator creates a new yield aggregator
func NewYieldAggregator(
	logger *logger.Logger,
	cache redis.Client,
	uniswapClient *UniswapClient,
	aaveClient *AaveClient,
) *YieldAggregator {
	return &YieldAggregator{
		logger:          logger.Named("yield-aggregator"),
		cache:           cache,
		uniswapClient:   uniswapClient,
		aaveClient:      aaveClient,
		minAPY:          decimal.NewFromFloat(0.05), // 5% minimum APY
		maxRisk:         RiskLevelMedium,
		scanInterval:    time.Minute * 5, // Scan every 5 minutes
		opportunities:   make(map[string]*YieldFarmingOpportunity),
		strategies:      make(map[string]*YieldStrategy),
		opportunityChan: make(chan *YieldFarmingOpportunity, 50),
		stopChan:        make(chan struct{}),
	}
}

// Start begins the yield aggregation process
func (ya *YieldAggregator) Start(ctx context.Context) error {
	ya.logger.Info("Starting yield aggregator")

	// Start the main scanning loop
	go ya.scanningLoop(ctx)

	// Start the opportunity processor
	go ya.processOpportunities(ctx)

	// Initialize default strategies
	ya.initializeDefaultStrategies()

	return nil
}

// Stop stops the yield aggregation process
func (ya *YieldAggregator) Stop() {
	ya.logger.Info("Stopping yield aggregator")
	close(ya.stopChan)
}

// GetBestOpportunities returns the best yield farming opportunities
func (ya *YieldAggregator) GetBestOpportunities(ctx context.Context, limit int) ([]*YieldFarmingOpportunity, error) {
	ya.mutex.RLock()
	defer ya.mutex.RUnlock()

	// Convert map to slice
	opportunities := make([]*YieldFarmingOpportunity, 0, len(ya.opportunities))
	for _, opp := range ya.opportunities {
		if opp.Active {
			opportunities = append(opportunities, opp)
		}
	}

	// Sort by APY (descending)
	sort.Slice(opportunities, func(i, j int) bool {
		return opportunities[i].APY.GreaterThan(opportunities[j].APY)
	})

	// Apply limit
	if limit > 0 && len(opportunities) > limit {
		opportunities = opportunities[:limit]
	}

	return opportunities, nil
}

// GetOptimalStrategy returns the optimal yield strategy for given parameters
func (ya *YieldAggregator) GetOptimalStrategy(ctx context.Context, req *OptimalStrategyRequest) (*YieldStrategy, error) {
	ya.logger.Info("Getting optimal strategy",
		zap.String("investment", req.InvestmentAmount.String()),
		zap.String("risk_tolerance", string(req.RiskTolerance)))

	// Get available opportunities
	opportunities, err := ya.GetBestOpportunities(ctx, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to get opportunities: %w", err)
	}

	// Filter opportunities based on requirements
	filteredOpps := ya.filterOpportunities(opportunities, req)

	// Create optimal strategy
	strategy := ya.createOptimalStrategy(filteredOpps, req)

	return strategy, nil
}



// scanningLoop runs the main scanning loop
func (ya *YieldAggregator) scanningLoop(ctx context.Context) {
	ticker := time.NewTicker(ya.scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ya.stopChan:
			return
		case <-ticker.C:
			ya.scanForOpportunities(ctx)
		}
	}
}

// scanForOpportunities scans for yield farming opportunities
func (ya *YieldAggregator) scanForOpportunities(ctx context.Context) {
	ya.logger.Debug("Scanning for yield opportunities")

	// Scan Uniswap V3 pools
	uniswapOpps, err := ya.scanUniswapOpportunities(ctx)
	if err != nil {
		ya.logger.Error("Failed to scan Uniswap opportunities", zap.Error(err))
	} else {
		for _, opp := range uniswapOpps {
			select {
			case ya.opportunityChan <- opp:
			default:
				ya.logger.Warn("Opportunity channel full")
			}
		}
	}

	// Scan Aave lending opportunities
	aaveOpps, err := ya.scanAaveOpportunities(ctx)
	if err != nil {
		ya.logger.Error("Failed to scan Aave opportunities", zap.Error(err))
	} else {
		for _, opp := range aaveOpps {
			select {
			case ya.opportunityChan <- opp:
			default:
				ya.logger.Warn("Opportunity channel full")
			}
		}
	}

	// Scan Coffee Token staking
	coffeeOpps, err := ya.scanCoffeeStakingOpportunities(ctx)
	if err != nil {
		ya.logger.Error("Failed to scan Coffee staking opportunities", zap.Error(err))
	} else {
		for _, opp := range coffeeOpps {
			select {
			case ya.opportunityChan <- opp:
			default:
				ya.logger.Warn("Opportunity channel full")
			}
		}
	}
}

// processOpportunities processes detected opportunities
func (ya *YieldAggregator) processOpportunities(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ya.stopChan:
			return
		case opp := <-ya.opportunityChan:
			ya.handleOpportunity(ctx, opp)
		}
	}
}

// handleOpportunity handles a detected yield opportunity
func (ya *YieldAggregator) handleOpportunity(ctx context.Context, opp *YieldFarmingOpportunity) {
	ya.mutex.Lock()
	defer ya.mutex.Unlock()

	// Store opportunity
	ya.opportunities[opp.ID] = opp

	// Cache opportunity
	cacheKey := fmt.Sprintf("yield:opportunity:%s", opp.ID)
	if err := ya.cache.Set(ctx, cacheKey, opp, time.Hour); err != nil {
		ya.logger.Error("Failed to cache opportunity", zap.Error(err))
	}

	ya.logger.Info("Yield opportunity detected",
		zap.String("id", opp.ID),
		zap.String("protocol", string(opp.Protocol)),
		zap.String("apy", opp.APY.String()),
		zap.String("tvl", opp.TVL.String()),
		zap.String("risk", string(opp.Risk)))
}

// scanUniswapOpportunities scans Uniswap for yield opportunities
func (ya *YieldAggregator) scanUniswapOpportunities(ctx context.Context) ([]*YieldFarmingOpportunity, error) {
	var opportunities []*YieldFarmingOpportunity

	// Get popular liquidity pools
	pools, err := ya.uniswapClient.GetLiquidityPools(ctx, &GetLiquidityPoolsRequest{
		Chain:  ChainEthereum,
		MinTVL: decimal.NewFromFloat(1000000), // $1M minimum TVL
		Limit:  10,
	})
	if err != nil {
		return nil, err
	}

	for _, pool := range pools {
		// Calculate estimated APY based on fees and volume
		estimatedAPY := ya.calculatePoolAPY(pool)

		if estimatedAPY.GreaterThan(ya.minAPY) {
			opp := &YieldFarmingOpportunity{
				ID:              uuid.New().String(),
				Protocol:        ProtocolTypeUniswap,
				Chain:           pool.Chain,
				Pool:            pool,
				Strategy:        "liquidity_provision",
				APY:             estimatedAPY,
				APR:             estimatedAPY, // Simplified
				TVL:             pool.TVL,
				MinDeposit:      decimal.NewFromFloat(100),   // $100 minimum
				MaxDeposit:      decimal.NewFromFloat(50000), // $50k maximum
				LockPeriod:      0,                           // No lock period for Uniswap
				RewardTokens:    []Token{},                   // Fee rewards in pool tokens
				Risk:            ya.calculatePoolRisk(pool),
				ImpermanentLoss: ya.calculateImpermanentLoss(pool),
				Active:          true,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}

			opportunities = append(opportunities, opp)
		}
	}

	return opportunities, nil
}

// scanAaveOpportunities scans Aave for lending opportunities
func (ya *YieldAggregator) scanAaveOpportunities(ctx context.Context) ([]*YieldFarmingOpportunity, error) {
	var opportunities []*YieldFarmingOpportunity

	// Mock Aave lending rates (in real implementation, fetch from Aave API)
	lendingRates := map[string]decimal.Decimal{
		"USDC": decimal.NewFromFloat(0.045), // 4.5% APY
		"USDT": decimal.NewFromFloat(0.042), // 4.2% APY
		"DAI":  decimal.NewFromFloat(0.038), // 3.8% APY
		"WETH": decimal.NewFromFloat(0.025), // 2.5% APY
	}

	for tokenSymbol, apy := range lendingRates {
		if apy.GreaterThan(ya.minAPY) {
			opp := &YieldFarmingOpportunity{
				ID:         uuid.New().String(),
				Protocol:   ProtocolTypeAave,
				Chain:      ChainEthereum,
				Strategy:   fmt.Sprintf("lending_%s", tokenSymbol),
				APY:        apy,
				APR:        apy,
				TVL:        decimal.NewFromFloat(100000000), // $100M TVL
				MinDeposit: decimal.NewFromFloat(10),        // $10 minimum
				MaxDeposit: decimal.NewFromFloat(1000000),   // $1M maximum
				LockPeriod: 0,                               // No lock period
				Risk:       RiskLevelLow,                    // Aave is considered low risk
				Active:     true,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			opportunities = append(opportunities, opp)
		}
	}

	return opportunities, nil
}

// scanCoffeeStakingOpportunities scans Coffee Token staking opportunities
func (ya *YieldAggregator) scanCoffeeStakingOpportunities(ctx context.Context) ([]*YieldFarmingOpportunity, error) {
	var opportunities []*YieldFarmingOpportunity

	// Coffee Token staking opportunity
	opp := &YieldFarmingOpportunity{
		ID:         uuid.New().String(),
		Protocol:   ProtocolType("coffee"),
		Chain:      ChainEthereum,
		Strategy:   "staking",
		APY:        decimal.NewFromFloat(0.12), // 12% APY
		APR:        decimal.NewFromFloat(0.12),
		TVL:        decimal.NewFromFloat(5000000), // $5M TVL
		MinDeposit: decimal.NewFromFloat(1),       // 1 COFFEE minimum
		MaxDeposit: decimal.NewFromFloat(1000000), // 1M COFFEE maximum
		LockPeriod: 0,                             // No lock period
		Risk:       RiskLevelMedium,               // Medium risk for new token
		Active:     true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	opportunities = append(opportunities, opp)

	return opportunities, nil
}

// calculatePoolAPY calculates estimated APY for a liquidity pool
func (ya *YieldAggregator) calculatePoolAPY(pool LiquidityPool) decimal.Decimal {
	// Simplified APY calculation based on fee tier and TVL
	// In real implementation, use 24h volume and fee data

	baseFeeAPY := pool.Fee.Mul(decimal.NewFromFloat(365)) // Annualized fee

	// Adjust based on TVL (higher TVL = more stable but lower APY)
	tvlFactor := decimal.NewFromFloat(1.0)
	if pool.TVL.GreaterThan(decimal.NewFromFloat(10000000)) { // $10M+
		tvlFactor = decimal.NewFromFloat(0.8)
	} else if pool.TVL.GreaterThan(decimal.NewFromFloat(1000000)) { // $1M+
		tvlFactor = decimal.NewFromFloat(0.9)
	}

	estimatedAPY := baseFeeAPY.Mul(tvlFactor)

	// Cap at reasonable maximum
	maxAPY := decimal.NewFromFloat(2.0) // 200% max
	if estimatedAPY.GreaterThan(maxAPY) {
		estimatedAPY = maxAPY
	}

	return estimatedAPY
}

// calculatePoolRisk calculates risk level for a liquidity pool
func (ya *YieldAggregator) calculatePoolRisk(pool LiquidityPool) RiskLevel {
	// Risk assessment based on various factors

	// TVL-based risk (higher TVL = lower risk)
	if pool.TVL.LessThan(decimal.NewFromFloat(100000)) { // < $100k
		return RiskLevelHigh
	} else if pool.TVL.LessThan(decimal.NewFromFloat(1000000)) { // < $1M
		return RiskLevelMedium
	}

	// Token pair risk assessment
	stableTokens := map[string]bool{
		"USDC": true,
		"USDT": true,
		"DAI":  true,
	}

	token0Stable := stableTokens[pool.Token0.Symbol]
	token1Stable := stableTokens[pool.Token1.Symbol]

	if token0Stable && token1Stable {
		return RiskLevelLow // Stable-stable pairs
	} else if token0Stable || token1Stable {
		return RiskLevelMedium // Stable-volatile pairs
	}

	return RiskLevelHigh // Volatile-volatile pairs
}

// calculateImpermanentLoss estimates impermanent loss for a pool
func (ya *YieldAggregator) calculateImpermanentLoss(pool LiquidityPool) decimal.Decimal {
	// Simplified impermanent loss calculation
	// In real implementation, use historical price data and volatility

	risk := ya.calculatePoolRisk(pool)

	switch risk {
	case RiskLevelLow:
		return decimal.NewFromFloat(0.01) // 1% estimated IL
	case RiskLevelMedium:
		return decimal.NewFromFloat(0.05) // 5% estimated IL
	case RiskLevelHigh:
		return decimal.NewFromFloat(0.15) // 15% estimated IL
	default:
		return decimal.NewFromFloat(0.05)
	}
}

// filterOpportunities filters opportunities based on requirements
func (ya *YieldAggregator) filterOpportunities(opportunities []*YieldFarmingOpportunity, req *OptimalStrategyRequest) []*YieldFarmingOpportunity {
	var filtered []*YieldFarmingOpportunity

	for _, opp := range opportunities {
		// Check risk tolerance
		if ya.riskLevelToInt(opp.Risk) > ya.riskLevelToInt(req.RiskTolerance) {
			continue
		}

		// Check minimum APY
		if opp.APY.LessThan(req.MinAPY) {
			continue
		}

		// Check lock period
		if req.MaxLockPeriod > 0 && opp.LockPeriod > req.MaxLockPeriod {
			continue
		}

		// Check minimum investment
		if opp.MinDeposit.GreaterThan(req.InvestmentAmount) {
			continue
		}

		filtered = append(filtered, opp)
	}

	return filtered
}

// createOptimalStrategy creates an optimal strategy from filtered opportunities
func (ya *YieldAggregator) createOptimalStrategy(opportunities []*YieldFarmingOpportunity, req *OptimalStrategyRequest) *YieldStrategy {
	if len(opportunities) == 0 {
		return nil
	}

	strategy := &YieldStrategy{
		ID:            uuid.New().String(),
		Name:          fmt.Sprintf("Optimal Strategy - %s", req.RiskTolerance),
		Type:          ya.getStrategyType(req.RiskTolerance),
		AutoCompound:  req.AutoCompound,
		RebalanceFreq: time.Hour * 24, // Daily rebalancing
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if req.Diversification && len(opportunities) > 1 {
		// Diversified strategy - allocate across multiple opportunities
		strategy.Opportunities = ya.createDiversifiedAllocation(opportunities, req)
	} else {
		// Single best opportunity
		strategy.Opportunities = []*YieldFarmingOpportunity{opportunities[0]}
	}

	// Calculate strategy metrics
	strategy.TotalAPY = ya.calculateStrategyAPY(strategy.Opportunities)
	strategy.Risk = ya.calculateStrategyRisk(strategy.Opportunities)
	strategy.MinInvestment = ya.calculateMinInvestment(strategy.Opportunities)
	strategy.MaxInvestment = ya.calculateMaxInvestment(strategy.Opportunities)

	return strategy
}

// initializeDefaultStrategies initializes default yield strategies
func (ya *YieldAggregator) initializeDefaultStrategies() {
	ya.mutex.Lock()
	defer ya.mutex.Unlock()

	// Conservative strategy
	conservative := &YieldStrategy{
		ID:            "conservative-default",
		Name:          "Conservative Yield Strategy",
		Type:          YieldStrategyTypeConservative,
		AutoCompound:  true,
		RebalanceFreq: time.Hour * 24 * 7, // Weekly rebalancing
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	ya.strategies[conservative.ID] = conservative

	// Balanced strategy
	balanced := &YieldStrategy{
		ID:            "balanced-default",
		Name:          "Balanced Yield Strategy",
		Type:          YieldStrategyTypeBalanced,
		AutoCompound:  true,
		RebalanceFreq: time.Hour * 24 * 3, // Every 3 days
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	ya.strategies[balanced.ID] = balanced

	// Aggressive strategy
	aggressive := &YieldStrategy{
		ID:            "aggressive-default",
		Name:          "Aggressive Yield Strategy",
		Type:          YieldStrategyTypeAggressive,
		AutoCompound:  true,
		RebalanceFreq: time.Hour * 24, // Daily rebalancing
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	ya.strategies[aggressive.ID] = aggressive
}

// Helper methods

func (ya *YieldAggregator) riskLevelToInt(risk RiskLevel) int {
	switch risk {
	case RiskLevelLow:
		return 1
	case RiskLevelMedium:
		return 2
	case RiskLevelHigh:
		return 3
	default:
		return 2
	}
}

func (ya *YieldAggregator) getStrategyType(risk RiskLevel) YieldStrategyType {
	switch risk {
	case RiskLevelLow:
		return YieldStrategyTypeConservative
	case RiskLevelMedium:
		return YieldStrategyTypeBalanced
	case RiskLevelHigh:
		return YieldStrategyTypeAggressive
	default:
		return YieldStrategyTypeBalanced
	}
}

func (ya *YieldAggregator) createDiversifiedAllocation(opportunities []*YieldFarmingOpportunity, req *OptimalStrategyRequest) []*YieldFarmingOpportunity {
	// Simple diversification - take top 3 opportunities
	maxOpps := 3
	if len(opportunities) < maxOpps {
		maxOpps = len(opportunities)
	}

	return opportunities[:maxOpps]
}

func (ya *YieldAggregator) calculateStrategyAPY(opportunities []*YieldFarmingOpportunity) decimal.Decimal {
	if len(opportunities) == 0 {
		return decimal.Zero
	}

	// Simple average APY (in real implementation, weight by allocation)
	totalAPY := decimal.Zero
	for _, opp := range opportunities {
		totalAPY = totalAPY.Add(opp.APY)
	}

	return totalAPY.Div(decimal.NewFromInt(int64(len(opportunities))))
}

func (ya *YieldAggregator) calculateStrategyRisk(opportunities []*YieldFarmingOpportunity) RiskLevel {
	if len(opportunities) == 0 {
		return RiskLevelMedium
	}

	// Take the highest risk level among opportunities
	maxRisk := RiskLevelLow
	for _, opp := range opportunities {
		if ya.riskLevelToInt(opp.Risk) > ya.riskLevelToInt(maxRisk) {
			maxRisk = opp.Risk
		}
	}

	return maxRisk
}

func (ya *YieldAggregator) calculateMinInvestment(opportunities []*YieldFarmingOpportunity) decimal.Decimal {
	if len(opportunities) == 0 {
		return decimal.Zero
	}

	// Sum of minimum deposits
	total := decimal.Zero
	for _, opp := range opportunities {
		total = total.Add(opp.MinDeposit)
	}

	return total
}

func (ya *YieldAggregator) calculateMaxInvestment(opportunities []*YieldFarmingOpportunity) decimal.Decimal {
	if len(opportunities) == 0 {
		return decimal.Zero
	}

	// Sum of maximum deposits
	total := decimal.Zero
	for _, opp := range opportunities {
		total = total.Add(opp.MaxDeposit)
	}

	return total
}
