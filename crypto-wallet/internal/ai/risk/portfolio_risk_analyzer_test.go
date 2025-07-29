package risk

import (
	"context"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test logger
func createTestLoggerForPortfolio() *logger.Logger {
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	return logger.NewLogger(logConfig)
}

// Helper function to create a test portfolio
func createTestPortfolio() *Portfolio {
	return &Portfolio{
		ID:         "test_portfolio_1",
		Address:    common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1"),
		TotalValue: decimal.NewFromFloat(100000), // $100k portfolio
		Assets: []*PortfolioAsset{
			{
				Symbol:      "BTC",
				Name:        "Bitcoin",
				Amount:      decimal.NewFromFloat(2.5),
				Value:       decimal.NewFromFloat(50000),
				Weight:      decimal.NewFromFloat(0.5),
				Price:       decimal.NewFromFloat(20000),
				Chain:       "bitcoin",
				Protocol:    "bitcoin",
				Sector:      "layer1",
				AssetType:   "major_crypto",
				LastUpdated: time.Now(),
			},
			{
				Symbol:      "ETH",
				Name:        "Ethereum",
				Amount:      decimal.NewFromFloat(20),
				Value:       decimal.NewFromFloat(30000),
				Weight:      decimal.NewFromFloat(0.3),
				Price:       decimal.NewFromFloat(1500),
				Chain:       "ethereum",
				Protocol:    "ethereum",
				Sector:      "layer1",
				AssetType:   "major_crypto",
				LastUpdated: time.Now(),
			},
			{
				Symbol:      "UNI",
				Name:        "Uniswap",
				Amount:      decimal.NewFromFloat(2000),
				Value:       decimal.NewFromFloat(20000),
				Weight:      decimal.NewFromFloat(0.2),
				Price:       decimal.NewFromFloat(10),
				Chain:       "ethereum",
				Protocol:    "uniswap",
				Sector:      "defi",
				AssetType:   "defi_token",
				LastUpdated: time.Now(),
			},
		},
		LastUpdated: time.Now(),
		Metadata:    make(map[string]interface{}),
	}
}

// Helper function to create a concentrated portfolio
func createConcentratedPortfolio() *Portfolio {
	return &Portfolio{
		ID:         "concentrated_portfolio",
		Address:    common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E2"),
		TotalValue: decimal.NewFromFloat(100000),
		Assets: []*PortfolioAsset{
			{
				Symbol:      "BTC",
				Name:        "Bitcoin",
				Amount:      decimal.NewFromFloat(4.5),
				Value:       decimal.NewFromFloat(90000),
				Weight:      decimal.NewFromFloat(0.9), // 90% concentration
				Price:       decimal.NewFromFloat(20000),
				Chain:       "bitcoin",
				Protocol:    "bitcoin",
				Sector:      "layer1",
				AssetType:   "major_crypto",
				LastUpdated: time.Now(),
			},
			{
				Symbol:      "ETH",
				Name:        "Ethereum",
				Amount:      decimal.NewFromFloat(6.67),
				Value:       decimal.NewFromFloat(10000),
				Weight:      decimal.NewFromFloat(0.1),
				Price:       decimal.NewFromFloat(1500),
				Chain:       "ethereum",
				Protocol:    "ethereum",
				Sector:      "layer1",
				AssetType:   "major_crypto",
				LastUpdated: time.Now(),
			},
		},
		LastUpdated: time.Now(),
		Metadata:    make(map[string]interface{}),
	}
}

func TestNewPortfolioRiskAnalyzer(t *testing.T) {
	logger := createTestLoggerForPortfolio()
	config := GetDefaultPortfolioRiskAnalyzerConfig()

	analyzer := NewPortfolioRiskAnalyzer(logger, config)

	assert.NotNil(t, analyzer)
	assert.Equal(t, config.Enabled, analyzer.config.Enabled)
	assert.Equal(t, config.UpdateInterval, analyzer.config.UpdateInterval)
	assert.False(t, analyzer.IsRunning())
	assert.NotNil(t, analyzer.correlationAnalyzer)
	assert.NotNil(t, analyzer.diversificationEngine)
	assert.NotNil(t, analyzer.varCalculator)
	assert.NotNil(t, analyzer.riskMetricsEngine)
}

func TestPortfolioRiskAnalyzer_StartStop(t *testing.T) {
	logger := createTestLoggerForPortfolio()
	config := GetDefaultPortfolioRiskAnalyzerConfig()

	analyzer := NewPortfolioRiskAnalyzer(logger, config)
	ctx := context.Background()

	err := analyzer.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, analyzer.IsRunning())

	err = analyzer.Stop()
	assert.NoError(t, err)
	assert.False(t, analyzer.IsRunning())
}

func TestPortfolioRiskAnalyzer_StartDisabled(t *testing.T) {
	logger := createTestLoggerForPortfolio()
	config := GetDefaultPortfolioRiskAnalyzerConfig()
	config.Enabled = false

	analyzer := NewPortfolioRiskAnalyzer(logger, config)
	ctx := context.Background()

	err := analyzer.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, analyzer.IsRunning()) // Should remain false when disabled
}

func TestPortfolioRiskAnalyzer_AnalyzePortfolioRisk(t *testing.T) {
	logger := createTestLoggerForPortfolio()
	config := GetDefaultPortfolioRiskAnalyzerConfig()

	analyzer := NewPortfolioRiskAnalyzer(logger, config)
	ctx := context.Background()

	// Start the analyzer
	err := analyzer.Start(ctx)
	require.NoError(t, err)
	defer analyzer.Stop()

	// Analyze test portfolio
	portfolio := createTestPortfolio()
	analysis, err := analyzer.AnalyzePortfolioRisk(ctx, portfolio)
	assert.NoError(t, err)
	assert.NotNil(t, analysis)

	// Validate analysis result
	assert.Equal(t, portfolio.ID, analysis.PortfolioID)
	assert.Equal(t, portfolio.Address, analysis.Address)
	assert.NotEmpty(t, analysis.AnalysisID)
	assert.True(t, analysis.OverallRiskScore.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, analysis.OverallRiskScore.LessThanOrEqual(decimal.NewFromFloat(100)))
	assert.Contains(t, []string{"low", "medium", "high", "critical"}, analysis.RiskLevel)
	assert.True(t, analysis.Confidence.GreaterThan(decimal.Zero))
	assert.True(t, analysis.Confidence.LessThanOrEqual(decimal.NewFromFloat(1)))

	// Check analysis components
	assert.NotNil(t, analysis.CorrelationAnalysis)
	assert.NotNil(t, analysis.DiversificationMetrics)
	assert.NotNil(t, analysis.VaRAnalysis)
	assert.NotNil(t, analysis.RiskMetrics)
	assert.NotNil(t, analysis.PerformanceMetrics)
	assert.NotNil(t, analysis.RebalancingAdvice)

	// Check recommendations and alerts
	assert.NotNil(t, analysis.Recommendations)
	assert.NotNil(t, analysis.RiskAlerts)
}

func TestPortfolioRiskAnalyzer_AnalyzeConcentratedPortfolio(t *testing.T) {
	logger := createTestLoggerForPortfolio()
	config := GetDefaultPortfolioRiskAnalyzerConfig()

	analyzer := NewPortfolioRiskAnalyzer(logger, config)
	ctx := context.Background()

	// Start the analyzer
	err := analyzer.Start(ctx)
	require.NoError(t, err)
	defer analyzer.Stop()

	// Analyze concentrated portfolio
	portfolio := createConcentratedPortfolio()
	analysis, err := analyzer.AnalyzePortfolioRisk(ctx, portfolio)
	assert.NoError(t, err)
	assert.NotNil(t, analysis)

	// Should detect high concentration risk
	assert.NotNil(t, analysis.DiversificationMetrics)
	assert.True(t, analysis.DiversificationMetrics.ConcentrationRisk.GreaterThan(decimal.NewFromFloat(0.8)),
		"Should detect high concentration risk")

	// Should have lower diversification score
	assert.True(t, analysis.DiversificationMetrics.DiversificationScore.LessThan(decimal.NewFromFloat(50)),
		"Concentrated portfolio should have lower diversification score")

	// Should generate alerts
	concentrationAlert := false
	for _, alert := range analysis.RiskAlerts {
		if alert.Type == "concentration_risk" {
			concentrationAlert = true
			break
		}
	}
	assert.True(t, concentrationAlert, "Should generate concentration risk alert")

	// Should have recommendations
	assert.NotEmpty(t, analysis.Recommendations, "Should generate recommendations for concentrated portfolio")
}

func TestPortfolioRiskAnalyzer_GetAnalysisMetrics(t *testing.T) {
	logger := createTestLoggerForPortfolio()
	config := GetDefaultPortfolioRiskAnalyzerConfig()

	analyzer := NewPortfolioRiskAnalyzer(logger, config)
	ctx := context.Background()

	// Start the analyzer
	err := analyzer.Start(ctx)
	require.NoError(t, err)
	defer analyzer.Stop()

	// Get analysis metrics
	metrics := analyzer.GetAnalysisMetrics()
	assert.NotNil(t, metrics)

	// Validate metrics
	assert.Contains(t, metrics, "cached_analyses")
	assert.Contains(t, metrics, "is_running")
	assert.Contains(t, metrics, "correlation_analyzer")
	assert.Contains(t, metrics, "diversification_engine")
	assert.Contains(t, metrics, "var_calculator")
	assert.Contains(t, metrics, "risk_metrics_engine")
	assert.Contains(t, metrics, "price_history_assets")

	assert.Equal(t, true, metrics["is_running"])
	assert.Equal(t, true, metrics["correlation_analyzer"])
	assert.Equal(t, true, metrics["diversification_engine"])
	assert.Equal(t, true, metrics["var_calculator"])
	assert.Equal(t, true, metrics["risk_metrics_engine"])
}

func TestPortfolioRiskAnalyzer_GetPortfolioSummary(t *testing.T) {
	logger := createTestLoggerForPortfolio()
	config := GetDefaultPortfolioRiskAnalyzerConfig()

	analyzer := NewPortfolioRiskAnalyzer(logger, config)
	ctx := context.Background()

	// Start the analyzer
	err := analyzer.Start(ctx)
	require.NoError(t, err)
	defer analyzer.Stop()

	// Get portfolio summary
	portfolio := createTestPortfolio()
	summary, err := analyzer.GetPortfolioSummary(ctx, portfolio)
	assert.NoError(t, err)
	assert.NotNil(t, summary)

	// Validate summary
	assert.Equal(t, portfolio.ID, summary.PortfolioID)
	assert.Equal(t, portfolio.TotalValue, summary.TotalValue)
	assert.Equal(t, len(portfolio.Assets), summary.AssetCount)
	assert.True(t, summary.OverallRiskScore.GreaterThanOrEqual(decimal.Zero))
	assert.Contains(t, []string{"low", "medium", "high", "critical"}, summary.RiskLevel)
	assert.True(t, summary.Confidence.GreaterThan(decimal.Zero))
}

func TestPortfolioRiskAnalyzer_Caching(t *testing.T) {
	logger := createTestLoggerForPortfolio()
	config := GetDefaultPortfolioRiskAnalyzerConfig()
	config.CacheTimeout = 1 * time.Hour // Long cache timeout

	analyzer := NewPortfolioRiskAnalyzer(logger, config)
	ctx := context.Background()

	// Start the analyzer
	err := analyzer.Start(ctx)
	require.NoError(t, err)
	defer analyzer.Stop()

	portfolio := createTestPortfolio()

	// First analysis
	analysis1, err := analyzer.AnalyzePortfolioRisk(ctx, portfolio)
	assert.NoError(t, err)
	assert.NotNil(t, analysis1)

	// Second analysis should return cached result
	analysis2, err := analyzer.AnalyzePortfolioRisk(ctx, portfolio)
	assert.NoError(t, err)
	assert.NotNil(t, analysis2)

	// Should be the same result (cached)
	assert.Equal(t, analysis1.AnalysisID, analysis2.AnalysisID)
	assert.Equal(t, analysis1.Timestamp, analysis2.Timestamp)
}

func TestGetDefaultPortfolioRiskAnalyzerConfig(t *testing.T) {
	config := GetDefaultPortfolioRiskAnalyzerConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, 5*time.Minute, config.UpdateInterval)
	assert.Equal(t, 30*time.Minute, config.CacheTimeout)
	assert.Equal(t, 252, config.HistoryWindow)

	// Check correlation config
	assert.True(t, config.CorrelationConfig.Enabled)
	assert.Equal(t, 30, config.CorrelationConfig.WindowSize)
	assert.NotEmpty(t, config.CorrelationConfig.CorrelationMethods)

	// Check diversification config
	assert.True(t, config.DiversificationConfig.Enabled)
	assert.True(t, config.DiversificationConfig.MaxConcentration.GreaterThan(decimal.Zero))
	assert.True(t, config.DiversificationConfig.MinAssets > 0)
	assert.NotEmpty(t, config.DiversificationConfig.SectorLimits)

	// Check VaR config
	assert.True(t, config.VaRConfig.Enabled)
	assert.NotEmpty(t, config.VaRConfig.Methods)
	assert.NotEmpty(t, config.VaRConfig.ConfidenceLevels)
	assert.True(t, config.VaRConfig.MonteCarloSims > 0)

	// Check risk metrics config
	assert.True(t, config.RiskMetricsConfig.Enabled)
	assert.True(t, config.RiskMetricsConfig.CalculateSharpе)
	assert.NotEmpty(t, config.RiskMetricsConfig.BenchmarkAsset)

	// Check alert thresholds
	assert.True(t, config.AlertThresholds.MaxConcentration.GreaterThan(decimal.Zero))
	assert.True(t, config.AlertThresholds.MaxCorrelation.GreaterThan(decimal.Zero))

	// Check rebalancing thresholds
	assert.True(t, config.RebalancingThresholds.Enabled)
	assert.True(t, config.RebalancingThresholds.DeviationThreshold.GreaterThan(decimal.Zero))
}

func TestValidatePortfolioRiskAnalyzerConfig(t *testing.T) {
	// Test valid config
	validConfig := GetDefaultPortfolioRiskAnalyzerConfig()
	err := ValidatePortfolioRiskAnalyzerConfig(validConfig)
	assert.NoError(t, err)

	// Test disabled config
	disabledConfig := GetDefaultPortfolioRiskAnalyzerConfig()
	disabledConfig.Enabled = false
	err = ValidatePortfolioRiskAnalyzerConfig(disabledConfig)
	assert.NoError(t, err)

	// Test invalid configs
	invalidConfigs := []PortfolioRiskAnalyzerConfig{
		// Invalid update interval
		{
			Enabled:        true,
			UpdateInterval: 0,
		},
		// Invalid cache timeout
		{
			Enabled:        true,
			UpdateInterval: 5 * time.Minute,
			CacheTimeout:   0,
		},
		// Invalid history window
		{
			Enabled:        true,
			UpdateInterval: 5 * time.Minute,
			CacheTimeout:   30 * time.Minute,
			HistoryWindow:  0,
		},
	}

	for i, config := range invalidConfigs {
		err := ValidatePortfolioRiskAnalyzerConfig(config)
		assert.Error(t, err, "Config %d should be invalid", i)
	}
}

func TestPortfolioConfigVariants(t *testing.T) {
	// Test conservative config
	conservativeConfig := GetConservativePortfolioConfig()
	assert.True(t, conservativeConfig.DiversificationConfig.MaxConcentration.LessThan(
		GetDefaultPortfolioRiskAnalyzerConfig().DiversificationConfig.MaxConcentration))

	// Test aggressive config
	aggressiveConfig := GetAggressivePortfolioConfig()
	assert.True(t, aggressiveConfig.DiversificationConfig.MaxConcentration.GreaterThan(
		GetDefaultPortfolioRiskAnalyzerConfig().DiversificationConfig.MaxConcentration))

	// Test DeFi config
	defiConfig := GetDeFiPortfolioConfig()
	assert.Contains(t, defiConfig.DiversificationConfig.SectorLimits, "defi")
	assert.Contains(t, defiConfig.DiversificationConfig.ProtocolLimits, "uniswap")

	// Validate all configs
	assert.NoError(t, ValidatePortfolioRiskAnalyzerConfig(conservativeConfig))
	assert.NoError(t, ValidatePortfolioRiskAnalyzerConfig(aggressiveConfig))
	assert.NoError(t, ValidatePortfolioRiskAnalyzerConfig(defiConfig))
}

func TestPortfolioUtilityFunctions(t *testing.T) {
	// Test risk level descriptions
	riskDescriptions := GetRiskLevelDescription()
	assert.NotEmpty(t, riskDescriptions)
	assert.Contains(t, riskDescriptions, "low")
	assert.Contains(t, riskDescriptions, "critical")

	// Test portfolio metrics descriptions
	metricsDescriptions := GetPortfolioMetricsDescription()
	assert.NotEmpty(t, metricsDescriptions)
	assert.Contains(t, metricsDescriptions, "sharpe_ratio")
	assert.Contains(t, metricsDescriptions, "var")

	// Test recommended asset allocation
	allocations := GetRecommendedAssetAllocation()
	assert.NotEmpty(t, allocations)
	assert.Contains(t, allocations, "conservative")
	assert.Contains(t, allocations, "aggressive")

	// Test correlation interpretation
	correlationInterpretation := GetCorrelationInterpretation()
	assert.NotEmpty(t, correlationInterpretation)
	assert.Contains(t, correlationInterpretation, "very_high")
	assert.Contains(t, correlationInterpretation, "negative")

	// Test VaR interpretation
	varInterpretation := GetVaRInterpretation()
	assert.NotEmpty(t, varInterpretation)
	assert.Contains(t, varInterpretation, "95_confidence")
	assert.Contains(t, varInterpretation, "interpretation")

	// Test diversification benefits
	diversificationBenefits := GetDiversificationBenefits()
	assert.NotEmpty(t, diversificationBenefits)
	assert.Contains(t, diversificationBenefits, "risk_reduction")
	assert.Contains(t, diversificationBenefits, "rebalancing_benefit")
}
