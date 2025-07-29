package risk

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// GetDefaultPortfolioRiskAnalyzerConfig returns default portfolio risk analyzer configuration
func GetDefaultPortfolioRiskAnalyzerConfig() PortfolioRiskAnalyzerConfig {
	return PortfolioRiskAnalyzerConfig{
		Enabled:         true,
		UpdateInterval:  5 * time.Minute,
		CacheTimeout:    30 * time.Minute,
		HistoryWindow:   252, // 1 year of trading days
		ConfidenceLevel: decimal.NewFromFloat(0.95),
		RiskFreeRate:    decimal.NewFromFloat(0.02), // 2% risk-free rate
		CorrelationConfig: CorrelationConfig{
			Enabled:            true,
			WindowSize:         30,
			UpdateInterval:     1 * time.Hour,
			MinDataPoints:      10,
			CorrelationMethods: []string{"pearson", "spearman"},
		},
		DiversificationConfig: DiversificationConfig{
			Enabled:          true,
			MaxConcentration: decimal.NewFromFloat(0.4), // 40% max concentration
			MinAssets:        3,
			SectorLimits: map[string]decimal.Decimal{
				"defi":       decimal.NewFromFloat(0.5),
				"layer1":     decimal.NewFromFloat(0.4),
				"layer2":     decimal.NewFromFloat(0.3),
				"stablecoin": decimal.NewFromFloat(0.3),
				"nft":        decimal.NewFromFloat(0.2),
				"gaming":     decimal.NewFromFloat(0.2),
				"metaverse":  decimal.NewFromFloat(0.2),
			},
			ChainLimits: map[string]decimal.Decimal{
				"ethereum":  decimal.NewFromFloat(0.6),
				"bitcoin":   decimal.NewFromFloat(0.4),
				"polygon":   decimal.NewFromFloat(0.3),
				"arbitrum":  decimal.NewFromFloat(0.3),
				"optimism":  decimal.NewFromFloat(0.3),
				"avalanche": decimal.NewFromFloat(0.3),
				"solana":    decimal.NewFromFloat(0.3),
			},
			ProtocolLimits: map[string]decimal.Decimal{
				"uniswap":   decimal.NewFromFloat(0.3),
				"aave":      decimal.NewFromFloat(0.3),
				"compound":  decimal.NewFromFloat(0.3),
				"curve":     decimal.NewFromFloat(0.3),
				"balancer":  decimal.NewFromFloat(0.2),
				"sushiswap": decimal.NewFromFloat(0.2),
			},
		},
		VaRConfig: VaRConfig{
			Enabled: true,
			Methods: []string{"historical", "parametric", "monte_carlo"},
			ConfidenceLevels: []decimal.Decimal{
				decimal.NewFromFloat(0.90),
				decimal.NewFromFloat(0.95),
				decimal.NewFromFloat(0.99),
			},
			TimeHorizons:   []int{1, 7, 30}, // 1 day, 1 week, 1 month
			MonteCarloSims: 10000,
		},
		RiskMetricsConfig: RiskMetricsConfig{
			Enabled:          true,
			CalculateSharpе:  true,
			CalculateSortino: true,
			CalculateTreynor: true,
			CalculateAlpha:   true,
			CalculateBeta:    true,
			BenchmarkAsset:   "BTC", // Bitcoin as benchmark
		},
		AlertThresholds: PortfolioAlertThresholds{
			MaxConcentration:   decimal.NewFromFloat(0.5),  // 50%
			MaxCorrelation:     decimal.NewFromFloat(0.8),  // 80%
			MaxVaR:             decimal.NewFromFloat(0.15), // 15%
			MinSharpeRatio:     decimal.NewFromFloat(0.5),  // 0.5
			MaxDrawdown:        decimal.NewFromFloat(0.2),  // 20%
			MinDiversification: decimal.NewFromFloat(0.6),  // 60%
		},
		DataSources: []string{
			"coingecko",
			"coinmarketcap",
			"binance",
			"kraken",
		},
		RebalancingThresholds: RebalancingThresholds{
			Enabled:              true,
			DeviationThreshold:   decimal.NewFromFloat(0.05), // 5% deviation
			TimeThreshold:        7 * 24 * time.Hour,         // 1 week
			VolatilityThreshold:  decimal.NewFromFloat(0.3),  // 30% volatility
			CorrelationThreshold: decimal.NewFromFloat(0.8),  // 80% correlation
		},
	}
}

// GetConservativePortfolioConfig returns conservative portfolio configuration
func GetConservativePortfolioConfig() PortfolioRiskAnalyzerConfig {
	config := GetDefaultPortfolioRiskAnalyzerConfig()

	// More conservative thresholds
	config.DiversificationConfig.MaxConcentration = decimal.NewFromFloat(0.25) // 25%
	config.DiversificationConfig.MinAssets = 5

	// Lower sector limits
	for sector := range config.DiversificationConfig.SectorLimits {
		config.DiversificationConfig.SectorLimits[sector] =
			config.DiversificationConfig.SectorLimits[sector].Mul(decimal.NewFromFloat(0.8))
	}

	// More conservative alert thresholds
	config.AlertThresholds.MaxConcentration = decimal.NewFromFloat(0.3)   // 30%
	config.AlertThresholds.MaxCorrelation = decimal.NewFromFloat(0.7)     // 70%
	config.AlertThresholds.MaxVaR = decimal.NewFromFloat(0.1)             // 10%
	config.AlertThresholds.MinSharpeRatio = decimal.NewFromFloat(0.8)     // 0.8
	config.AlertThresholds.MaxDrawdown = decimal.NewFromFloat(0.15)       // 15%
	config.AlertThresholds.MinDiversification = decimal.NewFromFloat(0.7) // 70%

	return config
}

// GetAggressivePortfolioConfig returns aggressive portfolio configuration
func GetAggressivePortfolioConfig() PortfolioRiskAnalyzerConfig {
	config := GetDefaultPortfolioRiskAnalyzerConfig()

	// More aggressive thresholds
	config.DiversificationConfig.MaxConcentration = decimal.NewFromFloat(0.6) // 60%
	config.DiversificationConfig.MinAssets = 2

	// Higher sector limits
	for sector := range config.DiversificationConfig.SectorLimits {
		config.DiversificationConfig.SectorLimits[sector] =
			config.DiversificationConfig.SectorLimits[sector].Mul(decimal.NewFromFloat(1.2))
	}

	// More aggressive alert thresholds
	config.AlertThresholds.MaxConcentration = decimal.NewFromFloat(0.7)   // 70%
	config.AlertThresholds.MaxCorrelation = decimal.NewFromFloat(0.9)     // 90%
	config.AlertThresholds.MaxVaR = decimal.NewFromFloat(0.25)            // 25%
	config.AlertThresholds.MinSharpeRatio = decimal.NewFromFloat(0.3)     // 0.3
	config.AlertThresholds.MaxDrawdown = decimal.NewFromFloat(0.3)        // 30%
	config.AlertThresholds.MinDiversification = decimal.NewFromFloat(0.4) // 40%

	return config
}

// GetDeFiPortfolioConfig returns DeFi-focused portfolio configuration
func GetDeFiPortfolioConfig() PortfolioRiskAnalyzerConfig {
	config := GetDefaultPortfolioRiskAnalyzerConfig()

	// DeFi-specific sector limits
	config.DiversificationConfig.SectorLimits = map[string]decimal.Decimal{
		"defi":       decimal.NewFromFloat(0.8), // Higher DeFi allocation
		"layer1":     decimal.NewFromFloat(0.3),
		"layer2":     decimal.NewFromFloat(0.4),
		"stablecoin": decimal.NewFromFloat(0.4),
		"governance": decimal.NewFromFloat(0.3),
		"yield":      decimal.NewFromFloat(0.5),
		"lending":    decimal.NewFromFloat(0.4),
		"dex":        decimal.NewFromFloat(0.4),
	}

	// DeFi-specific protocol limits
	config.DiversificationConfig.ProtocolLimits = map[string]decimal.Decimal{
		"uniswap":   decimal.NewFromFloat(0.4),
		"aave":      decimal.NewFromFloat(0.4),
		"compound":  decimal.NewFromFloat(0.4),
		"curve":     decimal.NewFromFloat(0.4),
		"balancer":  decimal.NewFromFloat(0.3),
		"sushiswap": decimal.NewFromFloat(0.3),
		"yearn":     decimal.NewFromFloat(0.3),
		"convex":    decimal.NewFromFloat(0.3),
		"makerdao":  decimal.NewFromFloat(0.3),
		"synthetix": decimal.NewFromFloat(0.2),
	}

	return config
}

// ValidatePortfolioRiskAnalyzerConfig validates portfolio risk analyzer configuration
func ValidatePortfolioRiskAnalyzerConfig(config PortfolioRiskAnalyzerConfig) error {
	if !config.Enabled {
		return nil
	}

	if config.UpdateInterval <= 0 {
		return fmt.Errorf("update interval must be positive")
	}

	if config.CacheTimeout <= 0 {
		return fmt.Errorf("cache timeout must be positive")
	}

	if config.HistoryWindow <= 0 {
		return fmt.Errorf("history window must be positive")
	}

	if config.ConfidenceLevel.LessThan(decimal.NewFromFloat(0.5)) ||
		config.ConfidenceLevel.GreaterThan(decimal.NewFromFloat(0.99)) {
		return fmt.Errorf("confidence level must be between 0.5 and 0.99")
	}

	// Validate correlation config
	if config.CorrelationConfig.Enabled {
		if config.CorrelationConfig.WindowSize <= 0 {
			return fmt.Errorf("correlation window size must be positive")
		}
		if config.CorrelationConfig.MinDataPoints <= 0 {
			return fmt.Errorf("correlation min data points must be positive")
		}
	}

	// Validate diversification config
	if config.DiversificationConfig.Enabled {
		if config.DiversificationConfig.MaxConcentration.LessThan(decimal.Zero) ||
			config.DiversificationConfig.MaxConcentration.GreaterThan(decimal.NewFromFloat(1.0)) {
			return fmt.Errorf("max concentration must be between 0 and 1")
		}
		if config.DiversificationConfig.MinAssets <= 0 {
			return fmt.Errorf("min assets must be positive")
		}
	}

	// Validate VaR config
	if config.VaRConfig.Enabled {
		for _, confidence := range config.VaRConfig.ConfidenceLevels {
			if confidence.LessThan(decimal.NewFromFloat(0.5)) ||
				confidence.GreaterThan(decimal.NewFromFloat(0.99)) {
				return fmt.Errorf("VaR confidence levels must be between 0.5 and 0.99")
			}
		}
		if config.VaRConfig.MonteCarloSims <= 0 {
			return fmt.Errorf("Monte Carlo simulations must be positive")
		}
	}

	// Validate alert thresholds
	thresholds := config.AlertThresholds
	if thresholds.MaxConcentration.LessThan(decimal.Zero) ||
		thresholds.MaxConcentration.GreaterThan(decimal.NewFromFloat(1.0)) {
		return fmt.Errorf("max concentration threshold must be between 0 and 1")
	}

	if thresholds.MaxCorrelation.LessThan(decimal.Zero) ||
		thresholds.MaxCorrelation.GreaterThan(decimal.NewFromFloat(1.0)) {
		return fmt.Errorf("max correlation threshold must be between 0 and 1")
	}

	return nil
}

// Note: GetRiskLevelDescription function is defined in config.go

// GetPortfolioMetricsDescription returns description for portfolio metrics
func GetPortfolioMetricsDescription() map[string]string {
	return map[string]string{
		"sharpe_ratio":    "Risk-adjusted return measure (higher is better)",
		"sortino_ratio":   "Downside risk-adjusted return measure",
		"treynor_ratio":   "Systematic risk-adjusted return measure",
		"alpha":           "Excess return over benchmark",
		"beta":            "Systematic risk relative to benchmark",
		"var":             "Value at Risk - potential loss at given confidence level",
		"max_drawdown":    "Maximum peak-to-trough decline",
		"correlation":     "Measure of asset price movement relationships",
		"concentration":   "Measure of portfolio concentration in single assets",
		"diversification": "Measure of portfolio diversification effectiveness",
	}
}

// GetRecommendedAssetAllocation returns recommended asset allocation by risk profile
func GetRecommendedAssetAllocation() map[string]map[string]decimal.Decimal {
	return map[string]map[string]decimal.Decimal{
		"conservative": {
			"stablecoin":   decimal.NewFromFloat(0.4),
			"major_crypto": decimal.NewFromFloat(0.4),
			"defi":         decimal.NewFromFloat(0.15),
			"altcoin":      decimal.NewFromFloat(0.05),
		},
		"moderate": {
			"stablecoin":   decimal.NewFromFloat(0.2),
			"major_crypto": decimal.NewFromFloat(0.5),
			"defi":         decimal.NewFromFloat(0.2),
			"altcoin":      decimal.NewFromFloat(0.1),
		},
		"aggressive": {
			"stablecoin":   decimal.NewFromFloat(0.1),
			"major_crypto": decimal.NewFromFloat(0.4),
			"defi":         decimal.NewFromFloat(0.3),
			"altcoin":      decimal.NewFromFloat(0.2),
		},
		"defi_focused": {
			"stablecoin":   decimal.NewFromFloat(0.15),
			"major_crypto": decimal.NewFromFloat(0.25),
			"defi":         decimal.NewFromFloat(0.5),
			"altcoin":      decimal.NewFromFloat(0.1),
		},
	}
}

// GetCorrelationInterpretation returns correlation interpretation guidelines
func GetCorrelationInterpretation() map[string]string {
	return map[string]string{
		"very_high": "0.8 - 1.0: Very high correlation - assets move together",
		"high":      "0.6 - 0.8: High correlation - strong relationship",
		"moderate":  "0.4 - 0.6: Moderate correlation - some relationship",
		"low":       "0.2 - 0.4: Low correlation - weak relationship",
		"very_low":  "0.0 - 0.2: Very low correlation - little relationship",
		"negative":  "-1.0 - 0.0: Negative correlation - assets move opposite",
	}
}

// GetVaRInterpretation returns VaR interpretation guidelines
func GetVaRInterpretation() map[string]string {
	return map[string]string{
		"95_confidence":  "95% confidence: Expected to be exceeded 5% of the time (1 in 20 days)",
		"99_confidence":  "99% confidence: Expected to be exceeded 1% of the time (1 in 100 days)",
		"interpretation": "VaR represents the maximum expected loss over a given time period at a specific confidence level",
		"limitations":    "VaR does not capture tail risk beyond the confidence level",
		"complement":     "Use Expected Shortfall (Conditional VaR) to understand tail risk",
	}
}

// GetDiversificationBenefits returns diversification benefits explanation
func GetDiversificationBenefits() map[string]string {
	return map[string]string{
		"risk_reduction":           "Reduces portfolio volatility without proportional return reduction",
		"correlation_benefit":      "Low correlation between assets provides diversification benefit",
		"sector_diversification":   "Spreading across sectors reduces sector-specific risks",
		"chain_diversification":    "Multi-chain exposure reduces blockchain-specific risks",
		"protocol_diversification": "Multiple protocols reduce smart contract risks",
		"rebalancing_benefit":      "Regular rebalancing maintains target allocations and captures returns",
	}
}
