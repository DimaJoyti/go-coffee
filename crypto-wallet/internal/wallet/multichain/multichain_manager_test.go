package multichain

import (
	"context"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a test logger
func createTestLogger() *logger.Logger {
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	return logger.NewLogger(logConfig)
}

func TestNewMultichainManager(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultMultichainConfig()

	manager := NewMultichainManager(logger, config)

	assert.NotNil(t, manager)
	assert.Equal(t, config.SupportedChains, manager.config.SupportedChains)
	assert.Equal(t, config.DefaultChain, manager.config.DefaultChain)
	assert.False(t, manager.isRunning)
	assert.NotNil(t, manager.chainManagers)
	assert.NotNil(t, manager.unifiedBalances)
	assert.NotNil(t, manager.bridgeManager)
	assert.NotNil(t, manager.gasTracker)
	assert.NotNil(t, manager.priceOracle)
	assert.NotNil(t, manager.portfolioTracker)
}

func TestMultichainManager_StartStop(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultMultichainConfig()
	// Disable actual network connections for testing
	for chainName := range config.ChainConfigs {
		config.ChainConfigs[chainName] = ChainConfig{
			ChainID: config.ChainConfigs[chainName].ChainID,
			Name:    config.ChainConfigs[chainName].Name,
			Enabled: false, // Disable for testing
		}
	}

	manager := NewMultichainManager(logger, config)
	ctx := context.Background()

	err := manager.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, manager.IsRunning())

	err = manager.Stop()
	assert.NoError(t, err)
	assert.False(t, manager.IsRunning())
}

func TestMultichainManager_StartDisabled(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultMultichainConfig()
	config.Enabled = false

	manager := NewMultichainManager(logger, config)
	ctx := context.Background()

	err := manager.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, manager.IsRunning()) // Should remain false when disabled
}

func TestMultichainManager_GetSupportedChains(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultMultichainConfig()

	manager := NewMultichainManager(logger, config)

	chains := manager.GetSupportedChains()
	assert.Equal(t, config.SupportedChains, chains)
	assert.Contains(t, chains, "ethereum")
	assert.Contains(t, chains, "bsc")
	assert.Contains(t, chains, "polygon")
	assert.Contains(t, chains, "arbitrum")
	assert.Contains(t, chains, "optimism")
	assert.Contains(t, chains, "avalanche")
}

func TestMultichainManager_GetChainStatus(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultMultichainConfig()

	manager := NewMultichainManager(logger, config)

	status := manager.GetChainStatus()
	assert.NotNil(t, status)
	
	for _, chainName := range config.SupportedChains {
		chainStatus, exists := status[chainName]
		assert.True(t, exists)
		assert.Equal(t, chainName, chainStatus.Chain)
		assert.Equal(t, config.ChainConfigs[chainName].ChainID, chainStatus.ChainID)
	}
}

func TestMultichainManager_GetNetworkStats(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultMultichainConfig()

	manager := NewMultichainManager(logger, config)

	stats := manager.GetNetworkStats()
	assert.NotNil(t, stats)
	assert.Equal(t, len(config.SupportedChains), stats.TotalChains)
	assert.GreaterOrEqual(t, stats.ActiveChains, 0)
	assert.LessOrEqual(t, stats.ActiveChains, stats.TotalChains)
	assert.NotNil(t, stats.ChainStats)
}

func TestMultichainManager_ValidateTransferRequest(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultMultichainConfig()

	manager := NewMultichainManager(logger, config)

	// Valid request
	validReq := &CrossChainTransferRequest{
		SourceChain: "ethereum",
		DestChain:   "polygon",
		Token: TokenConfig{
			Symbol:   "USDC",
			Decimals: 6,
		},
		Amount:      decimal.NewFromFloat(100),
		FromAddress: common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1"),
		ToAddress:   common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E2"),
		Slippage:    decimal.NewFromFloat(0.01),
	}

	err := manager.validateTransferRequest(validReq)
	assert.NoError(t, err)

	// Invalid requests
	invalidReqs := []*CrossChainTransferRequest{
		// Missing source chain
		{
			DestChain:   "polygon",
			Amount:      decimal.NewFromFloat(100),
			FromAddress: common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1"),
			ToAddress:   common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E2"),
			Slippage:    decimal.NewFromFloat(0.01),
		},
		// Same source and dest chain
		{
			SourceChain: "ethereum",
			DestChain:   "ethereum",
			Amount:      decimal.NewFromFloat(100),
			FromAddress: common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1"),
			ToAddress:   common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E2"),
			Slippage:    decimal.NewFromFloat(0.01),
		},
		// Zero amount
		{
			SourceChain: "ethereum",
			DestChain:   "polygon",
			Amount:      decimal.Zero,
			FromAddress: common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1"),
			ToAddress:   common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E2"),
			Slippage:    decimal.NewFromFloat(0.01),
		},
		// Invalid slippage
		{
			SourceChain: "ethereum",
			DestChain:   "polygon",
			Amount:      decimal.NewFromFloat(100),
			FromAddress: common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1"),
			ToAddress:   common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E2"),
			Slippage:    decimal.NewFromFloat(1.5), // > 1
		},
	}

	for i, req := range invalidReqs {
		err := manager.validateTransferRequest(req)
		assert.Error(t, err, "Request %d should be invalid", i)
	}
}

func TestMultichainManager_CalculateRiskMetrics(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultMultichainConfig()

	manager := NewMultichainManager(logger, config)

	// Test with empty balances
	emptyBalances := make(map[string]*UnifiedBalance)
	riskMetrics := manager.calculateRiskMetrics(emptyBalances, decimal.Zero)
	assert.NotNil(t, riskMetrics)
	assert.Equal(t, "low", riskMetrics.OverallRisk)

	// Test with concentrated balances
	concentratedBalances := map[string]*UnifiedBalance{
		"ETH": {
			Token:         TokenConfig{Symbol: "ETH"},
			TotalValueUSD: decimal.NewFromFloat(8000), // 80% of total
			ChainBalances: map[string]*ChainBalance{
				"ethereum": {ValueUSD: decimal.NewFromFloat(8000)},
			},
		},
		"USDC": {
			Token:         TokenConfig{Symbol: "USDC"},
			TotalValueUSD: decimal.NewFromFloat(2000), // 20% of total
			ChainBalances: map[string]*ChainBalance{
				"ethereum": {ValueUSD: decimal.NewFromFloat(2000)},
			},
		},
	}

	totalValue := decimal.NewFromFloat(10000)
	riskMetrics = manager.calculateRiskMetrics(concentratedBalances, totalValue)
	assert.NotNil(t, riskMetrics)
	assert.Equal(t, "high", riskMetrics.OverallRisk)
	assert.True(t, riskMetrics.ConcentrationRisk.GreaterThan(decimal.NewFromFloat(0.7)))
	assert.NotEmpty(t, riskMetrics.Recommendations)
}

func TestGetDefaultMultichainConfig(t *testing.T) {
	config := GetDefaultMultichainConfig()

	assert.True(t, config.Enabled)
	assert.NotEmpty(t, config.SupportedChains)
	assert.Equal(t, "ethereum", config.DefaultChain)
	assert.Equal(t, 30*time.Second, config.UpdateInterval)
	assert.Equal(t, decimal.NewFromFloat(0.001), config.BalanceThreshold)

	// Check chain configs
	assert.Contains(t, config.ChainConfigs, "ethereum")
	assert.Contains(t, config.ChainConfigs, "bsc")
	assert.Contains(t, config.ChainConfigs, "polygon")
	assert.Contains(t, config.ChainConfigs, "arbitrum")
	assert.Contains(t, config.ChainConfigs, "optimism")
	assert.Contains(t, config.ChainConfigs, "avalanche")

	// Check Ethereum config
	ethConfig := config.ChainConfigs["ethereum"]
	assert.Equal(t, int64(1), ethConfig.ChainID)
	assert.Equal(t, "Ethereum", ethConfig.Name)
	assert.Equal(t, "ETH", ethConfig.Symbol)
	assert.True(t, ethConfig.Enabled)
	assert.NotEmpty(t, ethConfig.RPCEndpoints)
	assert.Equal(t, "ETH", ethConfig.NativeToken.Symbol)
	assert.True(t, ethConfig.NativeToken.IsNative)

	// Check bridge config
	assert.True(t, config.BridgeConfig.Enabled)
	assert.NotEmpty(t, config.BridgeConfig.SupportedBridges)
	assert.Equal(t, "stargate", config.BridgeConfig.DefaultBridge)

	// Check gas config
	assert.True(t, config.GasConfig.Enabled)
	assert.Equal(t, 1*time.Minute, config.GasConfig.UpdateInterval)

	// Check price oracle config
	assert.True(t, config.PriceOracleConfig.Enabled)
	assert.Equal(t, "coingecko", config.PriceOracleConfig.Provider)
	assert.Equal(t, 5*time.Minute, config.PriceOracleConfig.UpdateInterval)

	// Check portfolio config
	assert.True(t, config.PortfolioConfig.Enabled)
	assert.Equal(t, 1*time.Hour, config.PortfolioConfig.UpdateInterval)
	assert.True(t, config.PortfolioConfig.RiskCalculation)
}

func TestValidateMultichainConfig(t *testing.T) {
	// Test valid config
	validConfig := GetDefaultMultichainConfig()
	err := ValidateMultichainConfig(validConfig)
	assert.NoError(t, err)

	// Test disabled config
	disabledConfig := GetDefaultMultichainConfig()
	disabledConfig.Enabled = false
	err = ValidateMultichainConfig(disabledConfig)
	assert.NoError(t, err)

	// Test invalid configs
	invalidConfigs := []MultichainConfig{
		// No supported chains
		{
			Enabled:         true,
			SupportedChains: []string{},
		},
		// No default chain
		{
			Enabled:         true,
			SupportedChains: []string{"ethereum"},
			DefaultChain:    "",
		},
		// Invalid update interval
		{
			Enabled:         true,
			SupportedChains: []string{"ethereum"},
			DefaultChain:    "ethereum",
			UpdateInterval:  0,
		},
		// Negative balance threshold
		{
			Enabled:          true,
			SupportedChains:  []string{"ethereum"},
			DefaultChain:     "ethereum",
			UpdateInterval:   30 * time.Second,
			BalanceThreshold: decimal.NewFromFloat(-1),
		},
	}

	for i, config := range invalidConfigs {
		err := ValidateMultichainConfig(config)
		assert.Error(t, err, "Config %d should be invalid", i)
	}
}

func TestUtilityFunctions(t *testing.T) {
	// Test IsValidChain
	assert.True(t, IsValidChain("ethereum"))
	assert.True(t, IsValidChain("bsc"))
	assert.True(t, IsValidChain("polygon"))
	assert.False(t, IsValidChain("invalid"))

	// Test GetChainID
	assert.Equal(t, int64(1), GetChainID("ethereum"))
	assert.Equal(t, int64(56), GetChainID("bsc"))
	assert.Equal(t, int64(137), GetChainID("polygon"))
	assert.Equal(t, int64(0), GetChainID("invalid"))

	// Test GetNativeToken
	ethToken := GetNativeToken("ethereum")
	assert.Equal(t, "ETH", ethToken.Symbol)
	assert.Equal(t, "Ethereum", ethToken.Name)
	assert.True(t, ethToken.IsNative)

	bnbToken := GetNativeToken("bsc")
	assert.Equal(t, "BNB", bnbToken.Symbol)
	assert.Equal(t, "Binance Coin", bnbToken.Name)
	assert.True(t, bnbToken.IsNative)

	// Test min/max functions
	assert.Equal(t, 1, min(1, 2))
	assert.Equal(t, 1, min(2, 1))
	assert.Equal(t, 2, max(1, 2))
	assert.Equal(t, 2, max(2, 1))
}
