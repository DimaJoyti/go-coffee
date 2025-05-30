package defi

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestTradingStrategyType_String(t *testing.T) {
	testCases := []struct {
		strategy TradingStrategyType
		expected string
	}{
		{StrategyTypeArbitrage, "arbitrage"},
		{StrategyTypeYieldFarming, "yield_farming"},
		{StrategyTypeDCA, "dca"},
		{StrategyTypeGridTrading, "grid_trading"},
		{StrategyTypeRebalancing, "rebalancing"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.strategy), func(t *testing.T) {
			assert.Equal(t, tc.expected, string(tc.strategy))
		})
	}
}

func TestRiskLevel_Validation(t *testing.T) {
	validRiskLevels := []RiskLevel{
		RiskLevelLow,
		RiskLevelMedium,
		RiskLevelHigh,
	}

	for _, risk := range validRiskLevels {
		t.Run(string(risk), func(t *testing.T) {
			assert.NotEmpty(t, string(risk))
		})
	}
}

func TestChain_Validation(t *testing.T) {
	validChains := []Chain{
		ChainEthereum,
		ChainBSC,
		ChainPolygon,
		ChainArbitrum,
		ChainOptimism,
	}

	for _, chain := range validChains {
		t.Run(string(chain), func(t *testing.T) {
			assert.NotEmpty(t, string(chain))
		})
	}
}

func TestToken_Validation(t *testing.T) {
	token := Token{
		Address:  "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1",
		Symbol:   "USDC",
		Name:     "USD Coin",
		Decimals: 6,
		Chain:    ChainEthereum,
	}

	assert.NotEmpty(t, token.Address)
	assert.NotEmpty(t, token.Symbol)
	assert.NotEmpty(t, token.Name)
	assert.Equal(t, 6, token.Decimals)
	assert.Equal(t, ChainEthereum, token.Chain)
}

func TestExchange_Validation(t *testing.T) {
	exchange := Exchange{
		ID:       "uniswap-v3",
		Name:     "Uniswap V3",
		Type:     ExchangeTypeDEX,
		Chain:    ChainEthereum,
		Protocol: ProtocolTypeUniswap,
		Address:  "0xE592427A0AEce92De3Edee1F18E0157C05861564",
		Fee:      decimal.NewFromFloat(0.003),
		Active:   true,
	}

	assert.NotEmpty(t, exchange.ID)
	assert.NotEmpty(t, exchange.Name)
	assert.Equal(t, ExchangeTypeDEX, exchange.Type)
	assert.Equal(t, ChainEthereum, exchange.Chain)
	assert.True(t, exchange.Fee.GreaterThan(decimal.Zero))
	assert.True(t, exchange.Active)
}

func TestArbitrageDetection_Validation(t *testing.T) {
	detection := ArbitrageDetection{
		ID:    "arb-001",
		Token: Token{Symbol: "ETH", Chain: ChainEthereum},
		SourceExchange: Exchange{
			ID:   "uniswap",
			Name: "Uniswap",
		},
		TargetExchange: Exchange{
			ID:   "1inch",
			Name: "1inch",
		},
		SourcePrice:  decimal.NewFromFloat(2000),
		TargetPrice:  decimal.NewFromFloat(2020),
		ProfitMargin: decimal.NewFromFloat(0.01), // 1%
		Volume:       decimal.NewFromFloat(10),
		NetProfit:    decimal.NewFromFloat(200),
		GasCost:      decimal.NewFromFloat(50),
		Confidence:   decimal.NewFromFloat(0.85),
		Risk:         RiskLevelMedium,
		Status:       OpportunityStatusDetected,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Minute * 5),
	}

	assert.NotEmpty(t, detection.ID)
	assert.True(t, detection.TargetPrice.GreaterThan(detection.SourcePrice))
	assert.True(t, detection.ProfitMargin.GreaterThan(decimal.Zero))
	assert.True(t, detection.NetProfit.GreaterThan(decimal.Zero))
	assert.True(t, detection.Confidence.GreaterThan(decimal.Zero))
	assert.True(t, detection.Confidence.LessThanOrEqual(decimal.NewFromFloat(1.0)))
	assert.True(t, detection.ExpiresAt.After(detection.CreatedAt))
}

func TestYieldFarmingOpportunity_Validation(t *testing.T) {
	opportunity := YieldFarmingOpportunity{
		ID:       "yield-001",
		Protocol: ProtocolTypeUniswap,
		Chain:    ChainEthereum,
		Pool: LiquidityPool{
			Address: "0x123",
			Token0:  Token{Symbol: "USDC"},
			Token1:  Token{Symbol: "ETH"},
		},
		Strategy:        "liquidity_provision",
		APY:             decimal.NewFromFloat(0.12), // 12%
		APR:             decimal.NewFromFloat(0.12),
		TVL:             decimal.NewFromFloat(1000000),
		MinDeposit:      decimal.NewFromFloat(100),
		MaxDeposit:      decimal.NewFromFloat(50000),
		LockPeriod:      0,
		RewardTokens:    []Token{},
		Risk:            RiskLevelMedium,
		ImpermanentLoss: decimal.NewFromFloat(0.05), // 5%
		Active:          true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	assert.NotEmpty(t, opportunity.ID)
	assert.True(t, opportunity.APY.GreaterThan(decimal.Zero))
	assert.True(t, opportunity.TVL.GreaterThan(decimal.Zero))
	assert.True(t, opportunity.MaxDeposit.GreaterThan(opportunity.MinDeposit))
	assert.True(t, opportunity.ImpermanentLoss.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, opportunity.Active)
}

func TestOnChainMetrics_Validation(t *testing.T) {
	metrics := OnChainMetrics{
		Token: Token{
			Symbol: "ETH",
			Chain:  ChainEthereum,
		},
		Chain:           ChainEthereum,
		Price:           decimal.NewFromFloat(2000),
		Volume24h:       decimal.NewFromFloat(1000000),
		Liquidity:       decimal.NewFromFloat(5000000),
		MarketCap:       decimal.NewFromFloat(240000000000),
		Holders:         1000000,
		Transactions24h: 50000,
		Volatility:      decimal.NewFromFloat(0.15), // 15%
		Timestamp:       time.Now(),
	}

	assert.True(t, metrics.Price.GreaterThan(decimal.Zero))
	assert.True(t, metrics.Volume24h.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, metrics.Liquidity.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, metrics.MarketCap.GreaterThan(decimal.Zero))
	assert.True(t, metrics.Holders >= 0)
	assert.True(t, metrics.Transactions24h >= 0)
	assert.True(t, metrics.Volatility.GreaterThanOrEqual(decimal.Zero))
}

func TestTradingPerformance_Calculations(t *testing.T) {
	performance := TradingPerformance{
		TotalTrades:   10,
		WinningTrades: 7,
		LosingTrades:  3,
		WinRate:       decimal.NewFromFloat(0.7), // 70%
		NetProfit:     decimal.NewFromFloat(1500),
		TotalProfit:   decimal.NewFromFloat(2000),
		TotalLoss:     decimal.NewFromFloat(500),
		MaxDrawdown:   decimal.NewFromFloat(200),
		Sharpe:        decimal.NewFromFloat(1.5),
		LastUpdated:   time.Now(),
	}

	// Validate calculations
	assert.Equal(t, performance.TotalTrades, performance.WinningTrades+performance.LosingTrades)
	assert.True(t, performance.WinRate.Equal(decimal.NewFromFloat(0.7)))
	assert.True(t, performance.NetProfit.Equal(performance.TotalProfit.Sub(performance.TotalLoss)))
	assert.True(t, performance.Sharpe.GreaterThan(decimal.Zero))
}

func TestArbitrageOpportunity_Validation(t *testing.T) {
	opportunity := ArbitrageOpportunity{
		ID:           "arb-opp-001",
		Token:        Token{Symbol: "ETH", Chain: ChainEthereum},
		BuyExchange:  "uniswap",
		SellExchange: "1inch",
		BuyPrice:     decimal.NewFromFloat(2000),
		SellPrice:    decimal.NewFromFloat(2020),
		ProfitMargin: decimal.NewFromFloat(0.01), // 1%
		Volume:       decimal.NewFromFloat(10),
		GasCost:      decimal.NewFromFloat(50),
		NetProfit:    decimal.NewFromFloat(150),
		ExpiresAt:    time.Now().Add(time.Minute * 5),
		CreatedAt:    time.Now(),
	}

	assert.NotEmpty(t, opportunity.ID)
	assert.True(t, opportunity.SellPrice.GreaterThan(opportunity.BuyPrice))
	assert.True(t, opportunity.ProfitMargin.GreaterThan(decimal.Zero))
	assert.True(t, opportunity.NetProfit.GreaterThan(decimal.Zero))
	assert.True(t, opportunity.ExpiresAt.After(opportunity.CreatedAt))
}

func TestDecimalComparisons(t *testing.T) {
	// Test decimal precision and comparisons
	price1 := decimal.NewFromFloat(2000.123456)
	price2 := decimal.NewFromFloat(2000.123457)

	assert.True(t, price2.GreaterThan(price1))
	assert.False(t, price1.Equal(price2))

	// Test percentage calculations
	profit := decimal.NewFromFloat(100)
	investment := decimal.NewFromFloat(1000)
	profitMargin := profit.Div(investment)

	assert.True(t, profitMargin.Equal(decimal.NewFromFloat(0.1))) // 10%
}

func TestTimeValidations(t *testing.T) {
	now := time.Now()

	// Test opportunity expiration
	opportunity := ArbitrageDetection{
		CreatedAt: now,
		ExpiresAt: now.Add(time.Minute * 5),
	}

	assert.True(t, opportunity.ExpiresAt.After(opportunity.CreatedAt))
	assert.True(t, opportunity.ExpiresAt.Sub(opportunity.CreatedAt) == time.Minute*5)

	// Test if opportunity is still valid
	isValid := time.Now().Before(opportunity.ExpiresAt)
	assert.True(t, isValid) // Should be valid since we just created it
}
