package defi

import (
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

// Mock implementations for testing and demonstration

// MockProtocolRegistry provides mock protocol registry
type MockProtocolRegistry struct{}

func (m *MockProtocolRegistry) GetProtocol(name string) (*ProtocolInfo, error) {
	// Mock protocol data based on name
	switch strings.ToLower(name) {
	case "uniswap":
		return &ProtocolInfo{
			Name:          "Uniswap",
			Symbol:        "UNI",
			Category:      "dex",
			Description:   "Decentralized exchange protocol for automated token swaps",
			Website:       "https://uniswap.org",
			Documentation: "https://docs.uniswap.org",
			Contracts: map[string]common.Address{
				"router":  common.HexToAddress("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D"),
				"factory": common.HexToAddress("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f"),
			},
			SupportedNetworks: []string{"ethereum", "polygon", "arbitrum"},
			TVL:               decimal.NewFromFloat(5000000000), // $5B
			Volume24h:         decimal.NewFromFloat(1000000000), // $1B
			Fees: map[string]decimal.Decimal{
				"swap": decimal.NewFromFloat(0.003), // 0.3%
			},
			Features:    []string{"AMM", "Liquidity Mining", "Governance"},
			LastUpdated: time.Now(),
		}, nil
	case "aave":
		return &ProtocolInfo{
			Name:          "Aave",
			Symbol:        "AAVE",
			Category:      "lending",
			Description:   "Decentralized lending and borrowing protocol",
			Website:       "https://aave.com",
			Documentation: "https://docs.aave.com",
			Contracts: map[string]common.Address{
				"lending_pool": common.HexToAddress("0x7d2768dE32b0b80b7a3454c06BdAc94A69DDc7A9"),
			},
			SupportedNetworks: []string{"ethereum", "polygon", "avalanche"},
			TVL:               decimal.NewFromFloat(8000000000), // $8B
			Volume24h:         decimal.NewFromFloat(500000000),  // $500M
			Features:          []string{"Lending", "Borrowing", "Flash Loans", "Governance"},
			LastUpdated:       time.Now(),
		}, nil
	case "compound":
		return &ProtocolInfo{
			Name:              "Compound",
			Symbol:            "COMP",
			Category:          "lending",
			Description:       "Algorithmic money market protocol",
			Website:           "https://compound.finance",
			Documentation:     "https://docs.compound.finance",
			SupportedNetworks: []string{"ethereum"},
			TVL:               decimal.NewFromFloat(3000000000), // $3B
			Volume24h:         decimal.NewFromFloat(200000000),  // $200M
			Features:          []string{"Lending", "Borrowing", "Governance"},
			LastUpdated:       time.Now(),
		}, nil
	default:
		return nil, fmt.Errorf("protocol not found: %s", name)
	}
}

func (m *MockProtocolRegistry) ListProtocols(category string) ([]*ProtocolInfo, error) {
	protocols := []*ProtocolInfo{}

	if category == "" || category == "dex" {
		uniswap, _ := m.GetProtocol("uniswap")
		protocols = append(protocols, uniswap)
	}

	if category == "" || category == "lending" {
		aave, _ := m.GetProtocol("aave")
		compound, _ := m.GetProtocol("compound")
		protocols = append(protocols, aave, compound)
	}

	return protocols, nil
}

func (m *MockProtocolRegistry) SearchProtocols(query string) ([]*ProtocolInfo, error) {
	allProtocols, _ := m.ListProtocols("")
	var results []*ProtocolInfo

	queryLower := strings.ToLower(query)
	for _, protocol := range allProtocols {
		if strings.Contains(strings.ToLower(protocol.Name), queryLower) ||
			strings.Contains(strings.ToLower(protocol.Description), queryLower) ||
			strings.Contains(strings.ToLower(protocol.Category), queryLower) {
			results = append(results, protocol)
		}
	}

	return results, nil
}

func (m *MockProtocolRegistry) UpdateProtocol(protocol *ProtocolInfo) error {
	// Mock update
	return nil
}

// MockStrategyAnalyzer provides mock strategy analysis
type MockStrategyAnalyzer struct{}

func (m *MockStrategyAnalyzer) AnalyzeStrategy(strategy *StrategyInfo) (*StrategyAnalysis, error) {
	return &StrategyAnalysis{
		Complexity:         strategy.Complexity,
		RiskLevel:          strategy.RiskLevel,
		ExpectedReturn:     strategy.ExpectedAPY,
		TimeCommitment:     24 * time.Hour, // Daily monitoring
		CapitalRequirement: strategy.MinimumInvestment,
		SkillRequirements:  []string{"DeFi basics", "Wallet management", "Risk assessment"},
		Recommendations:    []string{"Start with small amounts", "Monitor regularly", "Understand risks"},
	}, nil
}

func (m *MockStrategyAnalyzer) CompareStrategies(strategies []*StrategyInfo) (*StrategyComparison, error) {
	if len(strategies) < 2 {
		return nil, fmt.Errorf("need at least 2 strategies for comparison")
	}

	return &StrategyComparison{
		ComparedWith: strategies[1].Name,
		Metrics: map[string]interface{}{
			"apy_difference":  strategies[0].ExpectedAPY.Sub(strategies[1].ExpectedAPY),
			"risk_comparison": fmt.Sprintf("%s vs %s", strategies[0].RiskLevel, strategies[1].RiskLevel),
		},
		Pros:           []string{"Higher yield potential", "More established protocol"},
		Cons:           []string{"Higher complexity", "More gas fees"},
		Recommendation: "Choose based on risk tolerance and experience level",
	}, nil
}

func (m *MockStrategyAnalyzer) OptimizeStrategy(strategy *StrategyInfo, constraints map[string]interface{}) (*StrategyInfo, error) {
	// Mock optimization - return modified strategy
	optimized := *strategy
	optimized.Name = "Optimized " + strategy.Name
	optimized.ExpectedAPY = strategy.ExpectedAPY.Mul(decimal.NewFromFloat(1.1)) // 10% improvement
	return &optimized, nil
}

// MockRiskAssessor provides mock risk assessment
type MockRiskAssessor struct{}

func (m *MockRiskAssessor) AssessProtocolRisk(protocol *ProtocolInfo) (*RiskAnalysis, error) {
	riskLevel := "medium"
	riskScore := decimal.NewFromFloat(6.5)

	// Adjust based on protocol category
	switch protocol.Category {
	case "lending":
		riskLevel = "medium"
		riskScore = decimal.NewFromFloat(7.0)
	case "dex":
		riskLevel = "low"
		riskScore = decimal.NewFromFloat(5.0)
	case "derivatives":
		riskLevel = "high"
		riskScore = decimal.NewFromFloat(8.5)
	}

	return &RiskAnalysis{
		OverallRiskLevel: riskLevel,
		RiskScore:        riskScore,
		RiskFactors: []RiskFactor{
			{
				Type:        "smart_contract",
				Severity:    "medium",
				Description: "Smart contract vulnerabilities could lead to fund loss",
				Mitigation:  "Protocol has been audited by reputable firms",
				Probability: decimal.NewFromFloat(0.1),
				Impact:      decimal.NewFromFloat(0.8),
			},
			{
				Type:        "liquidity",
				Severity:    "low",
				Description: "Liquidity risk during market stress",
				Mitigation:  "Large TVL provides good liquidity buffer",
				Probability: decimal.NewFromFloat(0.2),
				Impact:      decimal.NewFromFloat(0.4),
			},
		},
		MitigationStrategies: []string{
			"Diversify across multiple protocols",
			"Start with small amounts",
			"Monitor protocol updates and audits",
		},
		MonitoringMetrics: []string{"TVL", "Volume", "Active users", "Governance activity"},
	}, nil
}

func (m *MockRiskAssessor) AssessStrategyRisk(strategy *StrategyInfo) (*RiskAnalysis, error) {
	return m.AssessProtocolRisk(&ProtocolInfo{Category: "strategy"})
}

func (m *MockRiskAssessor) CalculateRiskScore(factors []RiskFactor) decimal.Decimal {
	if len(factors) == 0 {
		return decimal.NewFromFloat(5.0)
	}

	totalScore := decimal.Zero
	for _, factor := range factors {
		score := factor.Probability.Mul(factor.Impact).Mul(decimal.NewFromFloat(10))
		totalScore = totalScore.Add(score)
	}

	return totalScore.Div(decimal.NewFromInt(int64(len(factors))))
}

// MockYieldCalculator provides mock yield calculations
type MockYieldCalculator struct{}

func (m *MockYieldCalculator) CalculateAPY(protocol *ProtocolInfo, amount decimal.Decimal) (decimal.Decimal, error) {
	// Mock APY calculation based on protocol category
	baseAPY := decimal.NewFromFloat(5.0) // 5% base

	switch protocol.Category {
	case "lending":
		baseAPY = decimal.NewFromFloat(8.0)
	case "dex":
		baseAPY = decimal.NewFromFloat(12.0)
	case "yield":
		baseAPY = decimal.NewFromFloat(15.0)
	}

	// Adjust based on amount (larger amounts might get slightly lower rates)
	if amount.GreaterThan(decimal.NewFromFloat(100000)) {
		baseAPY = baseAPY.Mul(decimal.NewFromFloat(0.95))
	}

	return baseAPY, nil
}

func (m *MockYieldCalculator) ProjectYield(strategy *StrategyInfo, amount decimal.Decimal, period time.Duration) (*YieldProjections, error) {
	baseAPY := strategy.ExpectedAPY

	return &YieldProjections{
		Conservative: baseAPY.Mul(decimal.NewFromFloat(0.7)),
		Realistic:    baseAPY,
		Optimistic:   baseAPY.Mul(decimal.NewFromFloat(1.3)),
		Timeframes: map[string]decimal.Decimal{
			"1_month":  baseAPY.Div(decimal.NewFromFloat(12)),
			"3_months": baseAPY.Div(decimal.NewFromFloat(4)),
			"6_months": baseAPY.Div(decimal.NewFromFloat(2)),
			"1_year":   baseAPY,
		},
		Assumptions: []string{
			"Market conditions remain stable",
			"Protocol maintains current performance",
			"No major security incidents",
		},
		Scenarios: []YieldScenario{
			{
				Name:          "Bull Market",
				Description:   "Favorable market conditions",
				Probability:   decimal.NewFromFloat(0.3),
				ExpectedYield: baseAPY.Mul(decimal.NewFromFloat(1.5)),
			},
			{
				Name:          "Bear Market",
				Description:   "Unfavorable market conditions",
				Probability:   decimal.NewFromFloat(0.2),
				ExpectedYield: baseAPY.Mul(decimal.NewFromFloat(0.5)),
			},
		},
	}, nil
}

func (m *MockYieldCalculator) CompareYields(protocols []*ProtocolInfo) ([]*YieldComparison, error) {
	var comparisons []*YieldComparison

	for _, protocol := range protocols {
		apy, _ := m.CalculateAPY(protocol, decimal.NewFromFloat(10000))

		risk := "medium"
		if protocol.Category == "dex" {
			risk = "low"
		} else if protocol.Category == "derivatives" {
			risk = "high"
		}

		liquidity := "high"
		if protocol.TVL.LessThan(decimal.NewFromFloat(1000000000)) {
			liquidity = "medium"
		}

		comparisons = append(comparisons, &YieldComparison{
			Protocol:  protocol.Name,
			APY:       apy,
			Risk:      risk,
			Liquidity: liquidity,
		})
	}

	return comparisons, nil
}

// MockExplanationEngine provides mock explanations
type MockExplanationEngine struct{}

func (m *MockExplanationEngine) ExplainProtocol(protocol *ProtocolInfo, level string, language string) (*ProtocolExplanation, error) {
	explanation := &ProtocolExplanation{
		Protocol: protocol,
		Summary:  fmt.Sprintf("%s is a %s protocol that %s", protocol.Name, protocol.Category, protocol.Description),
	}

	// Adjust explanation based on level
	switch level {
	case "beginner":
		explanation.DetailedExplanation = fmt.Sprintf("%s is a user-friendly platform where you can %s. It's designed to be simple and secure for newcomers to DeFi.", protocol.Name, strings.ToLower(protocol.Description))
		explanation.HowItWorks = []string{
			"Connect your wallet to the platform",
			"Choose the service you want to use",
			"Follow the simple step-by-step process",
			"Monitor your positions regularly",
		}
	case "intermediate":
		explanation.DetailedExplanation = fmt.Sprintf("%s operates as a %s protocol with advanced features including %s. Users can leverage various strategies to optimize their returns.", protocol.Name, protocol.Category, strings.Join(protocol.Features, ", "))
		explanation.HowItWorks = []string{
			"Protocol uses smart contracts for automation",
			"Liquidity is provided by users in exchange for fees",
			"Governance token holders can vote on protocol changes",
			"Multiple strategies can be combined for optimization",
		}
	case "advanced":
		explanation.DetailedExplanation = fmt.Sprintf("%s implements sophisticated %s mechanisms with contracts deployed at %v. The protocol architecture enables %s with current TVL of %s.", protocol.Name, protocol.Category, protocol.Contracts, strings.Join(protocol.Features, ", "), protocol.TVL.String())
		explanation.HowItWorks = []string{
			"Smart contract architecture ensures trustless operations",
			"Automated market makers use constant product formulas",
			"Governance mechanisms enable decentralized decision making",
			"Advanced users can compose multiple protocols",
		}
	}

	// Add key features
	for _, feature := range protocol.Features {
		explanation.KeyFeatures = append(explanation.KeyFeatures, FeatureExplanation{
			Name:        feature,
			Description: fmt.Sprintf("%s functionality in %s", feature, protocol.Name),
			Benefits:    []string{"Automated execution", "Transparent operations", "Composable with other protocols"},
			Limitations: []string{"Smart contract risk", "Gas costs", "Complexity for beginners"},
		})
	}

	// Add use cases
	explanation.UseCases = []UseCaseExplanation{
		{
			Title:        "Basic Usage",
			Description:  fmt.Sprintf("Use %s for basic %s operations", protocol.Name, protocol.Category),
			Scenario:     "New user wants to start with DeFi",
			Benefits:     []string{"Easy to use", "Lower risk", "Good learning experience"},
			Requirements: []string{"Wallet setup", "Basic understanding", "Small initial amount"},
		},
	}

	// Add FAQ
	explanation.FAQ = []FAQItem{
		{
			Question: fmt.Sprintf("How safe is %s?", protocol.Name),
			Answer:   fmt.Sprintf("%s has been audited and has a strong security track record, but all DeFi protocols carry some risk.", protocol.Name),
			Category: "security",
		},
		{
			Question: fmt.Sprintf("What are the fees for using %s?", protocol.Name),
			Answer:   "Fees vary depending on the operation and network congestion. Check the current rates on the platform.",
			Category: "fees",
		},
	}

	return explanation, nil
}

func (m *MockExplanationEngine) ExplainStrategy(strategy *StrategyInfo, level string, language string) (*StrategyExplanation, error) {
	explanation := &StrategyExplanation{
		Strategy: strategy,
		Summary:  fmt.Sprintf("%s is a %s strategy that %s", strategy.Name, strategy.Type, strategy.Description),
	}

	// Adjust explanation based on level
	switch level {
	case "beginner":
		explanation.DetailedExplanation = fmt.Sprintf("This strategy involves %s and is suitable for beginners with %s risk tolerance. Expected returns are around %s%% annually.", strings.ToLower(strategy.Description), strategy.RiskLevel, strategy.ExpectedAPY.String())
	case "intermediate":
		explanation.DetailedExplanation = fmt.Sprintf("The %s strategy leverages %s protocols to generate yield through %s. It requires %s minimum investment and %s complexity level.", strategy.Name, strings.Join(strategy.Protocols, ", "), strategy.Type, strategy.MinimumInvestment.String(), strategy.Complexity)
	case "advanced":
		explanation.DetailedExplanation = fmt.Sprintf("Advanced %s implementation across %s with sophisticated risk management. Strategy complexity: %s, Expected APY: %s%%, Time commitment: %s.", strategy.Type, strings.Join(strategy.Protocols, ", "), strategy.Complexity, strategy.ExpectedAPY.String(), strategy.TimeCommitment)
	}

	// Add step-by-step guide
	for i, step := range strategy.Steps {
		explanation.StepByStepGuide = append(explanation.StepByStepGuide, StepExplanation{
			Step:                &step,
			DetailedExplanation: fmt.Sprintf("Step %d involves %s which is necessary for %s", i+1, step.Description, step.Action),
			WhyNecessary:        fmt.Sprintf("This step is required to %s", step.Action),
			Alternatives:        []string{"Alternative approach available", "Can be automated"},
			CommonMistakes:      []string{"Insufficient gas", "Wrong slippage settings"},
			BestPractices:       []string{"Double-check parameters", "Start with small amounts"},
		})
	}

	// Add tips and warnings
	explanation.Tips = []string{
		"Start with small amounts to learn",
		"Monitor positions regularly",
		"Understand the risks involved",
		"Keep some ETH for gas fees",
	}

	explanation.Warnings = []string{
		"DeFi protocols carry smart contract risks",
		"Impermanent loss can affect returns",
		"Gas fees can be significant during network congestion",
	}

	return explanation, nil
}

func (m *MockExplanationEngine) GenerateComparison(items []interface{}, criteria []string) (interface{}, error) {
	// Mock comparison generation
	return map[string]interface{}{
		"comparison_type": "protocol_comparison",
		"criteria":        criteria,
		"items_count":     len(items),
		"generated_at":    time.Now(),
	}, nil
}

// MockTutorialGenerator provides mock tutorials
type MockTutorialGenerator struct{}

func (m *MockTutorialGenerator) GenerateProtocolTutorial(protocol *ProtocolInfo, level string) (*ProtocolTutorial, error) {
	return &ProtocolTutorial{
		Title:         fmt.Sprintf("How to use %s", protocol.Name),
		Difficulty:    level,
		EstimatedTime: 30 * time.Minute,
		Prerequisites: []string{"Wallet setup", "Basic DeFi knowledge"},
		Steps: []TutorialStep{
			{
				Number:  1,
				Title:   "Connect Wallet",
				Content: fmt.Sprintf("Connect your wallet to %s platform", protocol.Name),
				Actions: []string{"Click connect wallet", "Select your wallet", "Approve connection"},
				Tips:    []string{"Make sure you're on the correct website", "Check the URL carefully"},
			},
			{
				Number:  2,
				Title:   "Choose Operation",
				Content: "Select the operation you want to perform",
				Actions: []string{"Browse available options", "Read the details", "Select your choice"},
				Tips:    []string{"Start with smaller amounts", "Understand the risks"},
			},
		},
		Resources: []Resource{
			{
				Type:        "documentation",
				Title:       "Official Documentation",
				URL:         protocol.Documentation,
				Description: "Comprehensive protocol documentation",
			},
		},
	}, nil
}

func (m *MockTutorialGenerator) GenerateStrategyTutorial(strategy *StrategyInfo, level string) (*StrategyTutorial, error) {
	return &StrategyTutorial{
		Title:    fmt.Sprintf("%s Strategy Tutorial", strategy.Name),
		Overview: fmt.Sprintf("Learn how to implement %s strategy step by step", strategy.Name),
		LearningObjectives: []string{
			"Understand the strategy mechanics",
			"Learn risk management",
			"Execute the strategy safely",
		},
		Modules: []TutorialModule{
			{
				Number:        1,
				Title:         "Strategy Overview",
				Objectives:    []string{"Understand basic concepts", "Learn terminology"},
				Content:       fmt.Sprintf("Introduction to %s strategy", strategy.Name),
				Duration:      15 * time.Minute,
				Prerequisites: []string{"Basic DeFi knowledge"},
			},
		},
	}, nil
}

func (m *MockTutorialGenerator) CreateInteractiveTutorial(content interface{}) (interface{}, error) {
	return map[string]interface{}{
		"type":         "interactive_tutorial",
		"content":      content,
		"interactive":  true,
		"generated_at": time.Now(),
	}, nil
}

// MockComparisonEngine provides mock comparisons
type MockComparisonEngine struct{}

func (m *MockComparisonEngine) CompareProtocols(protocols []*ProtocolInfo, metrics []string) ([]*ProtocolComparison, error) {
	var comparisons []*ProtocolComparison

	if len(protocols) < 2 {
		return comparisons, nil
	}

	for i := 1; i < len(protocols); i++ {
		comparison := &ProtocolComparison{
			ComparedWith: protocols[i].Name,
			Similarities: []string{
				fmt.Sprintf("Both are %s protocols", protocols[0].Category),
				"Both support multiple networks",
			},
			Differences: []string{
				fmt.Sprintf("Different TVL: %s vs %s", protocols[0].TVL.String(), protocols[i].TVL.String()),
				"Different feature sets",
			},
			Advantages: []string{
				"Higher liquidity",
				"More established",
			},
			Disadvantages: []string{
				"Higher fees",
				"More complex",
			},
			UseCase: "Better for large transactions",
		}
		comparisons = append(comparisons, comparison)
	}

	return comparisons, nil
}

func (m *MockComparisonEngine) CompareStrategies(strategies []*StrategyInfo, metrics []string) ([]*StrategyComparison, error) {
	var comparisons []*StrategyComparison

	if len(strategies) < 2 {
		return comparisons, nil
	}

	for i := 1; i < len(strategies); i++ {
		comparison := &StrategyComparison{
			ComparedWith: strategies[i].Name,
			Metrics: map[string]interface{}{
				"apy_difference":  strategies[0].ExpectedAPY.Sub(strategies[i].ExpectedAPY),
				"risk_comparison": fmt.Sprintf("%s vs %s", strategies[0].RiskLevel, strategies[i].RiskLevel),
			},
			Pros: []string{
				"Higher expected returns",
				"More diversified",
			},
			Cons: []string{
				"Higher complexity",
				"More time required",
			},
			Recommendation: "Choose based on your risk tolerance and experience",
		}
		comparisons = append(comparisons, comparison)
	}

	return comparisons, nil
}

func (m *MockComparisonEngine) GenerateRecommendations(comparisons interface{}) ([]string, error) {
	return []string{
		"Consider your risk tolerance when choosing",
		"Start with simpler strategies",
		"Diversify across multiple protocols",
		"Monitor your positions regularly",
	}, nil
}

// MockProtocolDatabase provides mock protocol database
type MockProtocolDatabase struct{}

func (m *MockProtocolDatabase) GetProtocolData(name string) (*ProtocolInfo, error) {
	registry := &MockProtocolRegistry{}
	return registry.GetProtocol(name)
}

func (m *MockProtocolDatabase) UpdateProtocolData(protocol *ProtocolInfo) error {
	return nil
}

func (m *MockProtocolDatabase) SearchProtocols(criteria map[string]interface{}) ([]*ProtocolInfo, error) {
	registry := &MockProtocolRegistry{}
	return registry.ListProtocols("")
}

// MockStrategyDatabase provides mock strategy database
type MockStrategyDatabase struct{}

func (m *MockStrategyDatabase) GetStrategyData(name string) (*StrategyInfo, error) {
	// Mock strategy data
	return &StrategyInfo{
		Name:        name,
		Type:        "yield_farming",
		Description: "Provide liquidity to earn trading fees and token rewards",
		Protocols:   []string{"uniswap", "aave"},
		RequiredTokens: []TokenRequirement{
			{Symbol: "ETH", Amount: decimal.NewFromFloat(1), Purpose: "Liquidity provision"},
			{Symbol: "USDC", Amount: decimal.NewFromFloat(2000), Purpose: "Liquidity provision"},
		},
		Steps: []StrategyStep{
			{
				Number:      1,
				Title:       "Prepare tokens",
				Description: "Acquire required tokens",
				Action:      "Buy ETH and USDC",
				Protocol:    "exchange",
			},
			{
				Number:      2,
				Title:       "Provide liquidity",
				Description: "Add liquidity to the pool",
				Action:      "Add liquidity",
				Protocol:    "uniswap",
			},
		},
		ExpectedAPY:       decimal.NewFromFloat(12.5),
		MinimumInvestment: decimal.NewFromFloat(1000),
		Complexity:        "intermediate",
		RiskLevel:         "medium",
		TimeCommitment:    "Daily monitoring recommended",
	}, nil
}

func (m *MockStrategyDatabase) UpdateStrategyData(strategy *StrategyInfo) error {
	return nil
}

func (m *MockStrategyDatabase) SearchStrategies(criteria map[string]interface{}) ([]*StrategyInfo, error) {
	// Mock strategy search
	strategies := []*StrategyInfo{}

	strategy1, _ := m.GetStrategyData("Uniswap Liquidity Farming")
	strategy2, _ := m.GetStrategyData("Aave Lending Strategy")

	strategies = append(strategies, strategy1, strategy2)
	return strategies, nil
}

// MockGlossaryDatabase provides mock glossary
type MockGlossaryDatabase struct{}

func (m *MockGlossaryDatabase) GetTerm(term string) (*GlossaryItem, error) {
	// Mock glossary terms
	terms := map[string]*GlossaryItem{
		"apy": {
			Term:         "APY",
			Definition:   "Annual Percentage Yield - the rate of return earned on an investment over a year",
			Category:     "finance",
			Examples:     []string{"A pool offering 12% APY", "Compare APYs across protocols"},
			RelatedTerms: []string{"APR", "Yield", "Interest"},
		},
		"tvl": {
			Term:         "TVL",
			Definition:   "Total Value Locked - the total amount of assets locked in a DeFi protocol",
			Category:     "defi",
			Examples:     []string{"Protocol has $1B TVL", "TVL indicates protocol popularity"},
			RelatedTerms: []string{"Liquidity", "Assets", "Protocol"},
		},
		"impermanent_loss": {
			Term:         "Impermanent Loss",
			Definition:   "The temporary loss of funds experienced by liquidity providers due to volatility",
			Category:     "defi",
			Examples:     []string{"IL occurs when token prices diverge", "Can be mitigated with stable pairs"},
			RelatedTerms: []string{"Liquidity Provider", "AMM", "Volatility"},
		},
	}

	if glossaryItem, exists := terms[strings.ToLower(term)]; exists {
		return glossaryItem, nil
	}

	return nil, fmt.Errorf("term not found: %s", term)
}

func (m *MockGlossaryDatabase) SearchTerms(query string) ([]*GlossaryItem, error) {
	// Mock search - return all terms for simplicity
	var results []*GlossaryItem

	terms := []string{"apy", "tvl", "impermanent_loss"}
	for _, term := range terms {
		if item, err := m.GetTerm(term); err == nil {
			results = append(results, item)
		}
	}

	return results, nil
}

func (m *MockGlossaryDatabase) AddTerm(item *GlossaryItem) error {
	// Mock add term
	return nil
}
