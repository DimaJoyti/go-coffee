package defi

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

// GetDefaultProtocolExplainerConfig returns default protocol explainer configuration
func GetDefaultProtocolExplainerConfig() ProtocolExplainerConfig {
	return ProtocolExplainerConfig{
		Enabled:                  true,
		Language:                 "en",
		ExplanationLevel:         "intermediate",
		IncludeRisks:             true,
		IncludeYieldCalculations: true,
		ProtocolRegistryConfig: ProtocolRegistryConfig{
			Enabled:       true,
			AutoDiscovery: true,
			UpdateInterval: 1 * time.Hour,
			SupportedProtocols: []string{
				"uniswap", "aave", "compound", "makerdao", "curve", "balancer",
				"yearn", "convex", "lido", "rocket_pool", "frax", "olympus",
			},
			CustomProtocols: map[string]ProtocolInfo{},
		},
		StrategyAnalyzerConfig: StrategyAnalyzerConfig{
			Enabled:           true,
			AnalyzeComplexity: true,
			CalculateAPY:      true,
			AssessRisks:       true,
			CompareStrategies: true,
		},
		RiskAssessorConfig: RiskAssessorConfig{
			Enabled: true,
			RiskFactors: []string{
				"smart_contract", "liquidity", "market", "regulatory",
				"operational", "governance", "oracle", "bridge",
			},
			RiskWeights: map[string]decimal.Decimal{
				"smart_contract": decimal.NewFromFloat(0.3),
				"liquidity":      decimal.NewFromFloat(0.2),
				"market":         decimal.NewFromFloat(0.2),
				"regulatory":     decimal.NewFromFloat(0.1),
				"operational":    decimal.NewFromFloat(0.1),
				"governance":     decimal.NewFromFloat(0.05),
				"oracle":         decimal.NewFromFloat(0.03),
				"bridge":         decimal.NewFromFloat(0.02),
			},
			IncludeHistorical:  true,
			MonitorLiquidation: true,
		},
		YieldCalculatorConfig: YieldCalculatorConfig{
			Enabled: true,
			CalculationMethods: []string{
				"simple_interest", "compound_interest", "time_weighted",
				"dollar_weighted", "risk_adjusted",
			},
			IncludeCompounding: true,
			IncludeFees:        true,
			ProjectionPeriods: []string{
				"1_day", "1_week", "1_month", "3_months", "6_months", "1_year",
			},
		},
		ExplanationEngineConfig: ExplanationEngineConfig{
			Enabled:            true,
			UseTemplates:       true,
			IncludeExamples:    true,
			IncludeDiagrams:    false,
			CustomizationLevel: "high",
		},
		TutorialGeneratorConfig: TutorialGeneratorConfig{
			Enabled:            true,
			StepByStep:         true,
			IncludeScreenshots: false,
			InteractiveMode:    false,
			DifficultyLevels:   []string{"beginner", "intermediate", "advanced"},
		},
		ComparisonEngineConfig: ComparisonEngineConfig{
			Enabled: true,
			ComparisonMetrics: []string{
				"apy", "tvl", "volume", "fees", "risk_score", "user_count",
				"governance_score", "audit_score", "liquidity_score",
			},
			IncludeProsCons:   true,
			VisualComparisons: false,
		},
		DatabaseConfig: DatabaseConfig{
			Enabled:         true,
			UpdateFrequency: 6 * time.Hour,
			DataSources: []string{
				"defipulse", "defillama", "coingecko", "dune_analytics",
				"the_graph", "official_apis",
			},
			CacheSize: 1000,
		},
		CacheConfig: ExplanationCacheConfig{
			Enabled:         true,
			MaxSize:         5000,
			TTL:             2 * time.Hour,
			CleanupInterval: 30 * time.Minute,
		},
	}
}

// GetBeginnerProtocolExplainerConfig returns configuration optimized for beginners
func GetBeginnerProtocolExplainerConfig() ProtocolExplainerConfig {
	config := GetDefaultProtocolExplainerConfig()
	
	// Simplify for beginners
	config.ExplanationLevel = "beginner"
	config.IncludeRisks = true // Important for beginners
	config.IncludeYieldCalculations = false // Keep it simple
	
	// Focus on major, well-established protocols
	config.ProtocolRegistryConfig.SupportedProtocols = []string{
		"uniswap", "aave", "compound", "makerdao", "curve",
	}
	
	// Simplified risk assessment
	config.RiskAssessorConfig.RiskFactors = []string{
		"smart_contract", "liquidity", "market",
	}
	
	// Basic yield calculations
	config.YieldCalculatorConfig.CalculationMethods = []string{
		"simple_interest", "compound_interest",
	}
	config.YieldCalculatorConfig.ProjectionPeriods = []string{
		"1_month", "3_months", "1_year",
	}
	
	// Enhanced tutorials for beginners
	config.TutorialGeneratorConfig.StepByStep = true
	config.TutorialGeneratorConfig.IncludeScreenshots = true
	config.TutorialGeneratorConfig.DifficultyLevels = []string{"beginner"}
	
	// Simplified comparisons
	config.ComparisonEngineConfig.ComparisonMetrics = []string{
		"apy", "risk_score", "user_count",
	}
	
	return config
}

// GetAdvancedProtocolExplainerConfig returns configuration for advanced users
func GetAdvancedProtocolExplainerConfig() ProtocolExplainerConfig {
	config := GetDefaultProtocolExplainerConfig()
	
	// Advanced settings
	config.ExplanationLevel = "advanced"
	config.IncludeRisks = true
	config.IncludeYieldCalculations = true
	
	// Include more protocols
	config.ProtocolRegistryConfig.SupportedProtocols = append(
		config.ProtocolRegistryConfig.SupportedProtocols,
		"synthetix", "perpetual", "dydx", "gmx", "ribbon", "tokemak",
		"alchemix", "abracadabra", "euler", "notional", "pendle",
	)
	
	// Comprehensive risk assessment
	config.RiskAssessorConfig.IncludeHistorical = true
	config.RiskAssessorConfig.MonitorLiquidation = true
	
	// Advanced yield calculations
	config.YieldCalculatorConfig.CalculationMethods = append(
		config.YieldCalculatorConfig.CalculationMethods,
		"sharpe_ratio", "sortino_ratio", "max_drawdown", "var_calculation",
	)
	
	// Advanced explanations
	config.ExplanationEngineConfig.IncludeDiagrams = true
	config.ExplanationEngineConfig.CustomizationLevel = "maximum"
	
	// Interactive tutorials
	config.TutorialGeneratorConfig.InteractiveMode = true
	
	// Comprehensive comparisons
	config.ComparisonEngineConfig.ComparisonMetrics = append(
		config.ComparisonEngineConfig.ComparisonMetrics,
		"sharpe_ratio", "max_drawdown", "correlation", "beta",
	)
	config.ComparisonEngineConfig.VisualComparisons = true
	
	// More frequent updates
	config.DatabaseConfig.UpdateFrequency = 1 * time.Hour
	
	// Larger cache
	config.CacheConfig.MaxSize = 20000
	config.CacheConfig.TTL = 4 * time.Hour
	
	return config
}

// GetResearcherProtocolExplainerConfig returns configuration for researchers and analysts
func GetResearcherProtocolExplainerConfig() ProtocolExplainerConfig {
	config := GetAdvancedProtocolExplainerConfig()
	
	// Research-focused settings
	config.ExplanationLevel = "advanced"
	
	// Include all available protocols
	config.ProtocolRegistryConfig.AutoDiscovery = true
	config.ProtocolRegistryConfig.UpdateInterval = 30 * time.Minute
	
	// Comprehensive risk analysis
	config.RiskAssessorConfig.RiskFactors = append(
		config.RiskAssessorConfig.RiskFactors,
		"concentration", "correlation", "tail_risk", "black_swan",
	)
	
	// Research-grade yield calculations
	config.YieldCalculatorConfig.CalculationMethods = append(
		config.YieldCalculatorConfig.CalculationMethods,
		"monte_carlo", "stress_testing", "scenario_analysis",
	)
	
	// Detailed data sources
	config.DatabaseConfig.DataSources = append(
		config.DatabaseConfig.DataSources,
		"messari", "nansen", "chainalysis", "elliptic", "glassnode",
	)
	config.DatabaseConfig.UpdateFrequency = 15 * time.Minute
	
	// Maximum cache for research
	config.CacheConfig.MaxSize = 50000
	config.CacheConfig.TTL = 8 * time.Hour
	
	return config
}

// ValidateProtocolExplainerConfig validates protocol explainer configuration
func ValidateProtocolExplainerConfig(config ProtocolExplainerConfig) error {
	if !config.Enabled {
		return nil
	}
	
	// Validate language
	supportedLanguages := GetSupportedLanguages()
	isValidLanguage := false
	for _, lang := range supportedLanguages {
		if config.Language == lang {
			isValidLanguage = true
			break
		}
	}
	if !isValidLanguage {
		return fmt.Errorf("unsupported language: %s", config.Language)
	}
	
	// Validate explanation level
	supportedLevels := GetSupportedExplanationLevels()
	isValidLevel := false
	for _, level := range supportedLevels {
		if config.ExplanationLevel == level {
			isValidLevel = true
			break
		}
	}
	if !isValidLevel {
		return fmt.Errorf("unsupported explanation level: %s", config.ExplanationLevel)
	}
	
	// Validate protocol registry config
	if config.ProtocolRegistryConfig.Enabled {
		if config.ProtocolRegistryConfig.UpdateInterval <= 0 {
			return fmt.Errorf("protocol registry update interval must be positive")
		}
		
		if len(config.ProtocolRegistryConfig.SupportedProtocols) == 0 {
			return fmt.Errorf("at least one supported protocol must be specified")
		}
	}
	
	// Validate risk assessor config
	if config.RiskAssessorConfig.Enabled {
		if len(config.RiskAssessorConfig.RiskFactors) == 0 {
			return fmt.Errorf("at least one risk factor must be specified")
		}
		
		// Validate risk weights sum to approximately 1.0
		totalWeight := decimal.Zero
		for _, weight := range config.RiskAssessorConfig.RiskWeights {
			totalWeight = totalWeight.Add(weight)
		}
		if totalWeight.LessThan(decimal.NewFromFloat(0.9)) || totalWeight.GreaterThan(decimal.NewFromFloat(1.1)) {
			return fmt.Errorf("risk weights should sum to approximately 1.0, got %s", totalWeight.String())
		}
	}
	
	// Validate yield calculator config
	if config.YieldCalculatorConfig.Enabled {
		if len(config.YieldCalculatorConfig.CalculationMethods) == 0 {
			return fmt.Errorf("at least one calculation method must be specified")
		}
		
		if len(config.YieldCalculatorConfig.ProjectionPeriods) == 0 {
			return fmt.Errorf("at least one projection period must be specified")
		}
	}
	
	// Validate database config
	if config.DatabaseConfig.Enabled {
		if config.DatabaseConfig.UpdateFrequency <= 0 {
			return fmt.Errorf("database update frequency must be positive")
		}
		
		if config.DatabaseConfig.CacheSize <= 0 {
			return fmt.Errorf("database cache size must be positive")
		}
	}
	
	// Validate cache config
	if config.CacheConfig.Enabled {
		if config.CacheConfig.MaxSize <= 0 {
			return fmt.Errorf("cache max size must be positive")
		}
		
		if config.CacheConfig.TTL <= 0 {
			return fmt.Errorf("cache TTL must be positive")
		}
		
		if config.CacheConfig.CleanupInterval <= 0 {
			return fmt.Errorf("cache cleanup interval must be positive")
		}
	}
	
	return nil
}

// GetSupportedLanguages returns supported languages for explanations
func GetSupportedLanguages() []string {
	return []string{
		"en", "es", "fr", "de", "it", "pt", "ru", "ja", "ko", "zh",
	}
}

// GetSupportedExplanationLevels returns supported explanation levels
func GetSupportedExplanationLevels() []string {
	return []string{
		"beginner", "intermediate", "advanced",
	}
}

// GetSupportedProtocols returns list of supported DeFi protocols
func GetSupportedProtocols() []string {
	return []string{
		"uniswap", "aave", "compound", "makerdao", "curve", "balancer",
		"yearn", "convex", "lido", "rocket_pool", "frax", "olympus",
		"synthetix", "perpetual", "dydx", "gmx", "ribbon", "tokemak",
		"alchemix", "abracadabra", "euler", "notional", "pendle",
		"chainlink", "uma", "nexus_mutual", "bancor", "kyber",
	}
}

// GetSupportedStrategyTypes returns supported strategy types
func GetSupportedStrategyTypes() []string {
	return []string{
		"yield_farming", "liquidity_provision", "lending", "borrowing",
		"arbitrage", "delta_neutral", "leveraged_farming", "staking",
		"governance_mining", "options_strategies", "perpetual_trading",
	}
}

// GetSupportedRiskFactors returns supported risk factors
func GetSupportedRiskFactors() []string {
	return []string{
		"smart_contract", "liquidity", "market", "regulatory",
		"operational", "governance", "oracle", "bridge",
		"concentration", "correlation", "tail_risk", "black_swan",
	}
}

// GetSupportedCalculationMethods returns supported yield calculation methods
func GetSupportedCalculationMethods() []string {
	return []string{
		"simple_interest", "compound_interest", "time_weighted",
		"dollar_weighted", "risk_adjusted", "sharpe_ratio",
		"sortino_ratio", "max_drawdown", "var_calculation",
		"monte_carlo", "stress_testing", "scenario_analysis",
	}
}

// GetSupportedComparisonMetrics returns supported comparison metrics
func GetSupportedComparisonMetrics() []string {
	return []string{
		"apy", "tvl", "volume", "fees", "risk_score", "user_count",
		"governance_score", "audit_score", "liquidity_score",
		"sharpe_ratio", "max_drawdown", "correlation", "beta",
	}
}

// GetSupportedDataSources returns supported data sources
func GetSupportedDataSources() []string {
	return []string{
		"defipulse", "defillama", "coingecko", "dune_analytics",
		"the_graph", "official_apis", "messari", "nansen",
		"chainalysis", "elliptic", "glassnode",
	}
}

// GetOptimalConfigForUseCase returns optimal configuration for specific use cases
func GetOptimalConfigForUseCase(useCase string) (ProtocolExplainerConfig, error) {
	switch useCase {
	case "beginner":
		return GetBeginnerProtocolExplainerConfig(), nil
	case "advanced":
		return GetAdvancedProtocolExplainerConfig(), nil
	case "researcher":
		return GetResearcherProtocolExplainerConfig(), nil
	case "default":
		return GetDefaultProtocolExplainerConfig(), nil
	default:
		return ProtocolExplainerConfig{}, fmt.Errorf("unsupported use case: %s", useCase)
	}
}

// GetUseCaseDescription returns descriptions for use cases
func GetUseCaseDescription() map[string]string {
	return map[string]string{
		"beginner":   "Simplified explanations for DeFi newcomers with focus on safety and education",
		"advanced":   "Comprehensive explanations for experienced users with detailed analysis",
		"researcher": "In-depth explanations for researchers and analysts with maximum data coverage",
		"default":    "Balanced explanations suitable for most intermediate users",
	}
}

// GetProtocolCategoryDescription returns descriptions for protocol categories
func GetProtocolCategoryDescription() map[string]string {
	return map[string]string{
		"dex":         "Decentralized exchanges for token swapping and liquidity provision",
		"lending":     "Lending and borrowing protocols for earning interest and accessing credit",
		"yield":       "Yield farming and optimization protocols for maximizing returns",
		"derivatives": "Derivatives and synthetic asset protocols for advanced trading",
		"insurance":   "Decentralized insurance protocols for risk management",
		"staking":     "Staking protocols for earning rewards on proof-of-stake networks",
		"bridge":      "Cross-chain bridge protocols for asset transfers between networks",
		"oracle":      "Oracle protocols for providing external data to smart contracts",
	}
}

// GetStrategyComplexityDescription returns descriptions for strategy complexity levels
func GetStrategyComplexityDescription() map[string]string {
	return map[string]string{
		"beginner":     "Simple strategies suitable for DeFi newcomers with minimal risk",
		"intermediate": "Moderate complexity strategies requiring basic DeFi knowledge",
		"advanced":     "Complex strategies requiring deep understanding and active management",
	}
}

// GetRiskLevelDescription returns descriptions for risk levels
func GetRiskLevelDescription() map[string]string {
	return map[string]string{
		"low":    "Conservative strategies with minimal risk of capital loss",
		"medium": "Balanced strategies with moderate risk and return potential",
		"high":   "Aggressive strategies with high risk and high return potential",
	}
}

// GetCommonProtocolAddresses returns common protocol contract addresses
func GetCommonProtocolAddresses() map[string]map[string]common.Address {
	return map[string]map[string]common.Address{
		"uniswap": {
			"router_v2": common.HexToAddress("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D"),
			"factory_v2": common.HexToAddress("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f"),
			"router_v3": common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564"),
			"factory_v3": common.HexToAddress("0x1F98431c8aD98523631AE4a59f267346ea31F984"),
		},
		"aave": {
			"lending_pool": common.HexToAddress("0x7d2768dE32b0b80b7a3454c06BdAc94A69DDc7A9"),
			"data_provider": common.HexToAddress("0x057835Ad21a177dbdd3090bB1CAE03EaCF78Fc6d"),
		},
		"compound": {
			"comptroller": common.HexToAddress("0x3d9819210A31b4961b30EF54bE2aeD79B9c9Cd3B"),
			"ceth": common.HexToAddress("0x4Ddc2D193948926D02f9B1fE9e1daa0718270ED5"),
		},
	}
}

// GetDefaultRiskWeights returns default risk factor weights
func GetDefaultRiskWeights() map[string]decimal.Decimal {
	return map[string]decimal.Decimal{
		"smart_contract": decimal.NewFromFloat(0.30),
		"liquidity":      decimal.NewFromFloat(0.20),
		"market":         decimal.NewFromFloat(0.20),
		"regulatory":     decimal.NewFromFloat(0.10),
		"operational":    decimal.NewFromFloat(0.10),
		"governance":     decimal.NewFromFloat(0.05),
		"oracle":         decimal.NewFromFloat(0.03),
		"bridge":         decimal.NewFromFloat(0.02),
	}
}

// GetExplanationTemplates returns explanation templates for different levels
func GetExplanationTemplates() map[string]map[string]string {
	return map[string]map[string]string{
		"beginner": {
			"protocol_summary": "{{.Name}} is a {{.Category}} platform that {{.Description}}. It's designed to be user-friendly and secure.",
			"strategy_summary": "{{.Name}} is a {{.Type}} strategy that {{.Description}}. It's suitable for beginners with {{.RiskLevel}} risk.",
		},
		"intermediate": {
			"protocol_summary": "{{.Name}} operates as a {{.Category}} protocol with features including {{.Features}}. Current TVL: {{.TVL}}.",
			"strategy_summary": "{{.Name}} leverages {{.Protocols}} to generate {{.ExpectedAPY}}% APY through {{.Type}} with {{.Complexity}} complexity.",
		},
		"advanced": {
			"protocol_summary": "{{.Name}} implements {{.Category}} mechanisms with contracts at {{.Contracts}}. Architecture enables {{.Features}} with {{.TVL}} TVL.",
			"strategy_summary": "Advanced {{.Type}} implementation across {{.Protocols}} with {{.ExpectedAPY}}% expected APY, requiring {{.MinimumInvestment}} minimum investment.",
		},
	}
}

// GetCommonQuestions returns frequently asked questions about DeFi
func GetCommonQuestions() map[string][]string {
	return map[string][]string{
		"general": {
			"What is DeFi?",
			"How does DeFi work?",
			"Is DeFi safe?",
			"What are the risks of DeFi?",
			"How do I get started with DeFi?",
		},
		"protocols": {
			"What is a liquidity pool?",
			"How do automated market makers work?",
			"What is yield farming?",
			"What are governance tokens?",
			"How do flash loans work?",
		},
		"strategies": {
			"What is impermanent loss?",
			"How do I calculate APY?",
			"What is the difference between APR and APY?",
			"How do I manage risk in DeFi?",
			"What are the best DeFi strategies for beginners?",
		},
	}
}
