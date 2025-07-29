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
func createTestLogger() *logger.Logger {
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	return logger.NewLogger(logConfig)
}

func TestNewRiskManager(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultRiskManagerConfig()

	manager := NewRiskManager(logger, config)

	assert.NotNil(t, manager)
	assert.Equal(t, config.Enabled, manager.config.Enabled)
	assert.Equal(t, config.UpdateInterval, manager.config.UpdateInterval)
	assert.False(t, manager.IsRunning())
	assert.NotNil(t, manager.transactionRiskScorer)
	assert.NotNil(t, manager.volatilityAnalyzer)
	assert.NotNil(t, manager.contractAuditor)
	assert.NotNil(t, manager.portfolioAssessor)
	assert.NotNil(t, manager.marketPredictor)
}

func TestRiskManager_StartStop(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultRiskManagerConfig()

	manager := NewRiskManager(logger, config)
	ctx := context.Background()

	err := manager.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, manager.IsRunning())

	err = manager.Stop()
	assert.NoError(t, err)
	assert.False(t, manager.IsRunning())
}

func TestRiskManager_StartDisabled(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultRiskManagerConfig()
	config.Enabled = false

	manager := NewRiskManager(logger, config)
	ctx := context.Background()

	err := manager.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, manager.IsRunning()) // Should remain false when disabled
}

func TestRiskManager_AssessRisk(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultRiskManagerConfig()

	manager := NewRiskManager(logger, config)
	ctx := context.Background()

	// Start the manager
	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop()

	// Create assessment request
	req := &RiskAssessmentRequest{
		Address:                common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1"),
		AssessmentType:         "comprehensive",
		TimeFrame:              24 * time.Hour,
		IncludeTransactionRisk: true,
		IncludePortfolioRisk:   true,
		IncludeVolatilityRisk:  true,
		IncludeContractRisk:    false,
		IncludeMarketRisk:      true,
		Assets:                 []string{"ETH", "BTC", "USDC"},
		Portfolio: &PortfolioData{
			TotalValue: decimal.NewFromFloat(10000),
			Assets: map[string]*AssetHolding{
				"ETH": {
					Symbol:     "ETH",
					Amount:     decimal.NewFromFloat(5),
					Value:      decimal.NewFromFloat(8000),
					Percentage: decimal.NewFromFloat(0.8),
				},
				"USDC": {
					Symbol:     "USDC",
					Amount:     decimal.NewFromFloat(2000),
					Value:      decimal.NewFromFloat(2000),
					Percentage: decimal.NewFromFloat(0.2),
				},
			},
		},
	}

	// Perform risk assessment
	assessment, err := manager.AssessRisk(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, assessment)

	// Validate assessment
	assert.Equal(t, req.Address, assessment.Address)
	assert.NotEmpty(t, assessment.ID)
	assert.True(t, assessment.OverallRiskScore.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, assessment.OverallRiskScore.LessThanOrEqual(decimal.NewFromFloat(100)))
	assert.Contains(t, []string{"low", "medium", "high"}, assessment.RiskLevel)
	assert.True(t, assessment.Confidence.GreaterThan(decimal.Zero))
	assert.True(t, assessment.Confidence.LessThanOrEqual(decimal.NewFromFloat(1)))

	// Check individual risk assessments
	assert.NotNil(t, assessment.TransactionRisk)
	assert.NotNil(t, assessment.PortfolioRisk)
	assert.NotNil(t, assessment.VolatilityRisk)
	assert.Nil(t, assessment.ContractRisk) // Not requested
	assert.NotNil(t, assessment.MarketRisk)

	// Check recommendations and alerts
	assert.NotNil(t, assessment.Recommendations)
	assert.NotNil(t, assessment.Alerts)
}

func TestRiskManager_GetRiskMetrics(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultRiskManagerConfig()

	manager := NewRiskManager(logger, config)
	ctx := context.Background()

	// Start the manager
	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop()

	// Get risk metrics
	metrics, err := manager.GetRiskMetrics(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)

	// Validate metrics
	assert.GreaterOrEqual(t, metrics.TotalAssessments, 0)
	assert.GreaterOrEqual(t, metrics.HighRiskCount, 0)
	assert.GreaterOrEqual(t, metrics.MediumRiskCount, 0)
	assert.GreaterOrEqual(t, metrics.LowRiskCount, 0)
	assert.True(t, metrics.AverageRiskScore.GreaterThanOrEqual(decimal.Zero))
	assert.NotNil(t, metrics.AlertCounts)
	assert.NotNil(t, metrics.RiskDistribution)
	assert.NotNil(t, metrics.TrendAnalysis)
}

func TestRiskManager_GetActiveAlerts(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultRiskManagerConfig()

	manager := NewRiskManager(logger, config)
	ctx := context.Background()

	// Start the manager
	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop()

	// Get active alerts
	alerts, err := manager.GetActiveAlerts(ctx, nil)
	assert.NoError(t, err)
	assert.NotNil(t, alerts)

	// Get alerts for specific address
	address := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1")
	addressAlerts, err := manager.GetActiveAlerts(ctx, &address)
	assert.NoError(t, err)
	assert.NotNil(t, addressAlerts)
}

func TestRiskManager_AcknowledgeAlert(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultRiskManagerConfig()

	manager := NewRiskManager(logger, config)
	ctx := context.Background()

	// Start the manager
	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop()

	// Try to acknowledge non-existent alert
	err = manager.AcknowledgeAlert(ctx, "non-existent-alert")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "alert not found")
}

func TestRiskManager_GetSystemHealth(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultRiskManagerConfig()

	manager := NewRiskManager(logger, config)
	ctx := context.Background()

	// Start the manager
	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop()

	// Get system health
	health := manager.GetSystemHealth()
	assert.NotNil(t, health)
	assert.Contains(t, []string{"healthy", "degraded", "unhealthy"}, health.OverallStatus)
	assert.NotNil(t, health.ComponentStatuses)
	assert.NotNil(t, health.PerformanceMetrics)
	assert.True(t, health.ErrorRate.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, health.Uptime >= 0)

	// Check component statuses
	expectedComponents := []string{
		"transaction_risk_scorer",
		"volatility_analyzer",
		"contract_auditor",
		"portfolio_assessor",
		"market_predictor",
	}

	for _, component := range expectedComponents {
		status, exists := health.ComponentStatuses[component]
		assert.True(t, exists, "Component %s should exist in health status", component)
		assert.Contains(t, []string{"healthy", "unhealthy"}, status)
	}
}

func TestGetDefaultRiskManagerConfig(t *testing.T) {
	config := GetDefaultRiskManagerConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, 5*time.Minute, config.UpdateInterval)
	assert.Equal(t, 30*time.Minute, config.CacheTimeout)

	// Check alert thresholds
	assert.Equal(t, decimal.NewFromFloat(70), config.AlertThresholds.TransactionRisk)
	assert.Equal(t, decimal.NewFromFloat(65), config.AlertThresholds.PortfolioRisk)
	assert.Equal(t, decimal.NewFromFloat(60), config.AlertThresholds.VolatilityRisk)

	// Check component configs
	assert.True(t, config.TransactionRiskConfig.Enabled)
	assert.True(t, config.VolatilityConfig.Enabled)
	assert.True(t, config.ContractAuditConfig.Enabled)
	assert.True(t, config.PortfolioConfig.Enabled)
	assert.True(t, config.MarketPredictionConfig.Enabled)

	// Check data sources
	assert.Contains(t, config.DataSources, "coingecko")
	assert.Contains(t, config.DataSources, "etherscan")
	assert.True(t, config.DataSources["coingecko"].Enabled)

	// Check ML model paths
	assert.Contains(t, config.MLModelPaths, "transaction_risk")
	assert.Contains(t, config.MLModelPaths, "volatility")
	assert.Contains(t, config.MLModelPaths, "market_prediction")
}

func TestValidateRiskManagerConfig(t *testing.T) {
	// Test valid config
	validConfig := GetDefaultRiskManagerConfig()
	err := ValidateRiskManagerConfig(validConfig)
	assert.NoError(t, err)

	// Test disabled config
	disabledConfig := GetDefaultRiskManagerConfig()
	disabledConfig.Enabled = false
	err = ValidateRiskManagerConfig(disabledConfig)
	assert.NoError(t, err)

	// Test invalid configs
	invalidConfigs := []RiskManagerConfig{
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
		// Invalid alert thresholds
		{
			Enabled:        true,
			UpdateInterval: 5 * time.Minute,
			CacheTimeout:   30 * time.Minute,
			AlertThresholds: AlertThresholds{
				TransactionRisk: decimal.NewFromFloat(150), // > 100
			},
		},
	}

	for i, config := range invalidConfigs {
		err := ValidateRiskManagerConfig(config)
		assert.Error(t, err, "Config %d should be invalid", i)
	}
}

func TestUtilityFunctions(t *testing.T) {
	// Test risk level descriptions
	descriptions := GetRiskLevelDescription()
	assert.NotEmpty(t, descriptions)
	assert.Contains(t, descriptions, "low")
	assert.Contains(t, descriptions, "medium")
	assert.Contains(t, descriptions, "high")
	assert.Contains(t, descriptions, "critical")

	// Test alert severity descriptions
	severities := GetAlertSeverityDescription()
	assert.NotEmpty(t, severities)
	assert.Contains(t, severities, "low")
	assert.Contains(t, severities, "medium")
	assert.Contains(t, severities, "high")
	assert.Contains(t, severities, "critical")

	// Test supported risk factors
	riskFactors := GetSupportedRiskFactors()
	assert.NotEmpty(t, riskFactors)
	assert.Contains(t, riskFactors, "concentration")
	assert.Contains(t, riskFactors, "correlation")
	assert.Contains(t, riskFactors, "liquidity")
	assert.Contains(t, riskFactors, "volatility")

	// Test supported ML models
	mlModels := GetSupportedMLModels()
	assert.NotEmpty(t, mlModels)
	assert.Contains(t, mlModels, "linear_regression")
	assert.Contains(t, mlModels, "random_forest")
	assert.Contains(t, mlModels, "neural_network")
	assert.Contains(t, mlModels, "lstm")
}
