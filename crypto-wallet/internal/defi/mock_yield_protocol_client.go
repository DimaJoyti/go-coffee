package defi

import (
	"context"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// MockYieldProtocolClient implements YieldProtocolClient for testing and development
type MockYieldProtocolClient struct {
	protocolName string
	logger       *logger.Logger
}

// NewMockYieldProtocolClient creates a new mock yield protocol client
func NewMockYieldProtocolClient(protocolName string, logger *logger.Logger) *MockYieldProtocolClient {
	return &MockYieldProtocolClient{
		protocolName: protocolName,
		logger:       logger.Named(fmt.Sprintf("mock-%s-client", protocolName)),
	}
}

// GetPoolInfo returns mock pool information
func (m *MockYieldProtocolClient) GetPoolInfo(ctx context.Context, poolAddress string) (*PoolInfo, error) {
	m.logger.Debug("Getting pool info", zap.String("pool", poolAddress))

	// Generate mock pool data based on protocol
	var apy decimal.Decimal
	var tvl decimal.Decimal
	var volume decimal.Decimal

	switch m.protocolName {
	case "uniswap":
		apy = decimal.NewFromFloat(0.12)       // 12% APY
		tvl = decimal.NewFromFloat(5000000)    // $5M TVL
		volume = decimal.NewFromFloat(1000000) // $1M daily volume
	case "sushiswap":
		apy = decimal.NewFromFloat(0.18)      // 18% APY
		tvl = decimal.NewFromFloat(3000000)   // $3M TVL
		volume = decimal.NewFromFloat(800000) // $800k daily volume
	case "curve":
		apy = decimal.NewFromFloat(0.08)       // 8% APY
		tvl = decimal.NewFromFloat(10000000)   // $10M TVL
		volume = decimal.NewFromFloat(2000000) // $2M daily volume
	case "balancer":
		apy = decimal.NewFromFloat(0.15)      // 15% APY
		tvl = decimal.NewFromFloat(4000000)   // $4M TVL
		volume = decimal.NewFromFloat(600000) // $600k daily volume
	default:
		apy = decimal.NewFromFloat(0.10)      // 10% APY
		tvl = decimal.NewFromFloat(1000000)   // $1M TVL
		volume = decimal.NewFromFloat(500000) // $500k daily volume
	}

	return &PoolInfo{
		Address:        poolAddress,
		Name:           fmt.Sprintf("%s Pool", m.protocolName),
		Protocol:       m.protocolName,
		Chain:          "ethereum",
		Token0:         "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1", // USDC
		Token1:         "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // WETH
		Token0Symbol:   "USDC",
		Token1Symbol:   "WETH",
		TVL:            tvl,
		Volume24h:      volume,
		FeeAPR:         apy.Mul(decimal.NewFromFloat(0.6)), // 60% from fees
		RewardAPR:      apy.Mul(decimal.NewFromFloat(0.4)), // 40% from rewards
		TotalAPY:       apy,
		RewardTokens:   []string{fmt.Sprintf("%s_TOKEN", m.protocolName)},
		LiquidityDepth: tvl.Mul(decimal.NewFromFloat(0.8)), // 80% of TVL
		IsActive:       true,
		RiskLevel:      "moderate",
		LastUpdated:    time.Now(),
	}, nil
}

// GetUserPosition returns mock user position
func (m *MockYieldProtocolClient) GetUserPosition(ctx context.Context, userAddress, poolAddress string) (*UserPosition, error) {
	m.logger.Debug("Getting user position",
		zap.String("user", userAddress),
		zap.String("pool", poolAddress))

	return &UserPosition{
		PoolAddress:     poolAddress,
		LiquidityAmount: decimal.NewFromFloat(1000), // $1000 liquidity
		Token0Amount:    decimal.NewFromFloat(500),  // $500 in token0
		Token1Amount:    decimal.NewFromFloat(0.25), // 0.25 ETH
		RewardsEarned:   decimal.NewFromFloat(50),   // $50 in rewards
		RewardTokens:    []string{fmt.Sprintf("%s_TOKEN", m.protocolName)},
		LastUpdated:     time.Now(),
	}, nil
}

// DepositLiquidity simulates depositing liquidity
func (m *MockYieldProtocolClient) DepositLiquidity(ctx context.Context, params *DepositParams) (*YieldTransactionResult, error) {
	m.logger.Info("Depositing liquidity",
		zap.String("pool", params.PoolAddress),
		zap.String("token0_amount", params.Token0Amount.String()),
		zap.String("token1_amount", params.Token1Amount.String()))

	// Simulate transaction
	return &YieldTransactionResult{
		TransactionHash: fmt.Sprintf("0x%s_deposit_%d", m.protocolName, time.Now().Unix()),
		Success:         true,
		GasUsed:         decimal.NewFromFloat(150000),      // 150k gas
		GasPrice:        decimal.NewFromFloat(20000000000), // 20 gwei
		BlockNumber:     uint64(time.Now().Unix()),
		Timestamp:       time.Now(),
	}, nil
}

// WithdrawLiquidity simulates withdrawing liquidity
func (m *MockYieldProtocolClient) WithdrawLiquidity(ctx context.Context, params *WithdrawParams) (*YieldTransactionResult, error) {
	m.logger.Info("Withdrawing liquidity",
		zap.String("pool", params.PoolAddress),
		zap.String("liquidity_amount", params.LiquidityAmount.String()))

	// Simulate transaction
	return &YieldTransactionResult{
		TransactionHash: fmt.Sprintf("0x%s_withdraw_%d", m.protocolName, time.Now().Unix()),
		Success:         true,
		GasUsed:         decimal.NewFromFloat(120000),      // 120k gas
		GasPrice:        decimal.NewFromFloat(20000000000), // 20 gwei
		BlockNumber:     uint64(time.Now().Unix()),
		Timestamp:       time.Now(),
	}, nil
}

// ClaimRewards simulates claiming rewards
func (m *MockYieldProtocolClient) ClaimRewards(ctx context.Context, params *ClaimParams) (*YieldTransactionResult, error) {
	m.logger.Info("Claiming rewards",
		zap.String("pool", params.PoolAddress),
		zap.Strings("reward_tokens", params.RewardTokens))

	// Simulate transaction
	return &YieldTransactionResult{
		TransactionHash: fmt.Sprintf("0x%s_claim_%d", m.protocolName, time.Now().Unix()),
		Success:         true,
		GasUsed:         decimal.NewFromFloat(80000),       // 80k gas
		GasPrice:        decimal.NewFromFloat(20000000000), // 20 gwei
		BlockNumber:     uint64(time.Now().Unix()),
		Timestamp:       time.Now(),
	}, nil
}

// CompoundRewards simulates compounding rewards
func (m *MockYieldProtocolClient) CompoundRewards(ctx context.Context, params *CompoundParams) (*YieldTransactionResult, error) {
	m.logger.Info("Compounding rewards",
		zap.String("pool", params.PoolAddress),
		zap.String("reward_amount", params.RewardAmount.String()))

	// Simulate transaction
	return &YieldTransactionResult{
		TransactionHash: fmt.Sprintf("0x%s_compound_%d", m.protocolName, time.Now().Unix()),
		Success:         true,
		GasUsed:         decimal.NewFromFloat(200000),      // 200k gas
		GasPrice:        decimal.NewFromFloat(20000000000), // 20 gwei
		BlockNumber:     uint64(time.Now().Unix()),
		Timestamp:       time.Now(),
	}, nil
}

// GetAvailablePools returns mock available pools
func (m *MockYieldProtocolClient) GetAvailablePools(ctx context.Context) ([]*PoolInfo, error) {
	m.logger.Debug("Getting available pools")

	pools := []*PoolInfo{}

	// Generate mock pools based on protocol
	poolConfigs := []struct {
		name    string
		token0  string
		token1  string
		symbol0 string
		symbol1 string
		apy     float64
		tvl     float64
		volume  float64
	}{
		{"USDC/WETH", "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1", "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", "USDC", "WETH", 0.12, 5000000, 1000000},
		{"DAI/USDC", "0x6B175474E89094C44Da98b954EedeAC495271d0F", "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1", "DAI", "USDC", 0.08, 8000000, 1500000},
		{"WBTC/WETH", "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599", "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", "WBTC", "WETH", 0.15, 3000000, 800000},
	}

	for i, config := range poolConfigs {
		poolAddress := fmt.Sprintf("0x%s_pool_%d", m.protocolName, i+1)

		pool := &PoolInfo{
			Address:        poolAddress,
			Name:           fmt.Sprintf("%s %s Pool", m.protocolName, config.name),
			Protocol:       m.protocolName,
			Chain:          "ethereum",
			Token0:         config.token0,
			Token1:         config.token1,
			Token0Symbol:   config.symbol0,
			Token1Symbol:   config.symbol1,
			TVL:            decimal.NewFromFloat(config.tvl),
			Volume24h:      decimal.NewFromFloat(config.volume),
			FeeAPR:         decimal.NewFromFloat(config.apy * 0.6),
			RewardAPR:      decimal.NewFromFloat(config.apy * 0.4),
			TotalAPY:       decimal.NewFromFloat(config.apy),
			RewardTokens:   []string{fmt.Sprintf("%s_TOKEN", m.protocolName)},
			LiquidityDepth: decimal.NewFromFloat(config.tvl * 0.8),
			IsActive:       true,
			RiskLevel:      "moderate",
			LastUpdated:    time.Now(),
		}

		pools = append(pools, pool)
	}

	return pools, nil
}

// EstimateGas estimates gas for an operation
func (m *MockYieldProtocolClient) EstimateGas(ctx context.Context, operation string, params interface{}) (decimal.Decimal, error) {
	m.logger.Debug("Estimating gas", zap.String("operation", operation))

	// Mock gas estimates based on operation
	gasEstimates := map[string]decimal.Decimal{
		"deposit":  decimal.NewFromFloat(150000),
		"withdraw": decimal.NewFromFloat(120000),
		"claim":    decimal.NewFromFloat(80000),
		"compound": decimal.NewFromFloat(200000),
	}

	if estimate, exists := gasEstimates[operation]; exists {
		return estimate, nil
	}

	return decimal.NewFromFloat(100000), nil // Default estimate
}

// GetSupportedTokens returns supported tokens
func (m *MockYieldProtocolClient) GetSupportedTokens() []string {
	return []string{
		"0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1", // USDC
		"0x6B175474E89094C44Da98b954EedeAC495271d0F", // DAI
		"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // WETH
		"0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599", // WBTC
		"0xdAC17F958D2ee523a2206206994597C13D831ec7", // USDT
	}
}

// GetProtocolName returns the protocol name
func (m *MockYieldProtocolClient) GetProtocolName() string {
	return m.protocolName
}
