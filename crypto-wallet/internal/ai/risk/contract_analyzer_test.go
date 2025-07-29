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
func createTestLoggerForContract() *logger.Logger {
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	return logger.NewLogger(logConfig)
}

// Sample Solidity contract for testing
const sampleContract = `
pragma solidity ^0.8.0;

contract TestContract {
    mapping(address => uint256) public balances;
    address public owner;
    bool private locked;
    
    modifier onlyOwner() {
        require(msg.sender == owner, "Not owner");
        _;
    }
    
    modifier nonReentrant() {
        require(!locked, "Reentrant call");
        locked = true;
        _;
        locked = false;
    }
    
    constructor() {
        owner = msg.sender;
    }
    
    function deposit() external payable {
        balances[msg.sender] += msg.value;
    }
    
    function withdraw(uint256 amount) external nonReentrant {
        require(balances[msg.sender] >= amount, "Insufficient balance");
        balances[msg.sender] -= amount;
        (bool success, ) = msg.sender.call{value: amount}("");
        require(success, "Transfer failed");
    }
    
    function emergencyWithdraw() external onlyOwner {
        payable(owner).transfer(address(this).balance);
    }
}
`

const vulnerableContract = `
pragma solidity ^0.7.0;

contract VulnerableContract {
    mapping(address => uint256) public balances;
    
    function deposit() external payable {
        balances[msg.sender] += msg.value;
    }
    
    function withdraw(uint256 amount) external {
        require(balances[msg.sender] >= amount);
        msg.sender.call{value: amount}("");
        balances[msg.sender] -= amount;
    }
}
`

func TestNewSmartContractAnalyzer(t *testing.T) {
	logger := createTestLoggerForContract()
	config := GetDefaultContractAnalysisConfig()

	analyzer := NewSmartContractAnalyzer(logger, config)

	assert.NotNil(t, analyzer)
	assert.Equal(t, config.Enabled, analyzer.config.Enabled)
	assert.Equal(t, config.MaxAnalysisTime, analyzer.config.MaxAnalysisTime)
	assert.False(t, analyzer.IsRunning())
	assert.NotNil(t, analyzer.patternAnalyzer)
	assert.NotNil(t, analyzer.vulnerabilityScanner)
	assert.NotNil(t, analyzer.formalVerifier)
	assert.NotNil(t, analyzer.exploitDatabase)
}

func TestSmartContractAnalyzer_StartStop(t *testing.T) {
	logger := createTestLoggerForContract()
	config := GetDefaultContractAnalysisConfig()

	analyzer := NewSmartContractAnalyzer(logger, config)
	ctx := context.Background()

	err := analyzer.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, analyzer.IsRunning())

	err = analyzer.Stop()
	assert.NoError(t, err)
	assert.False(t, analyzer.IsRunning())
}

func TestSmartContractAnalyzer_StartDisabled(t *testing.T) {
	logger := createTestLoggerForContract()
	config := GetDefaultContractAnalysisConfig()
	config.Enabled = false

	analyzer := NewSmartContractAnalyzer(logger, config)
	ctx := context.Background()

	err := analyzer.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, analyzer.IsRunning()) // Should remain false when disabled
}

func TestSmartContractAnalyzer_AnalyzeContract(t *testing.T) {
	logger := createTestLoggerForContract()
	config := GetDefaultContractAnalysisConfig()

	analyzer := NewSmartContractAnalyzer(logger, config)
	ctx := context.Background()

	// Start the analyzer
	err := analyzer.Start(ctx)
	require.NoError(t, err)
	defer analyzer.Stop()

	// Analyze sample contract
	contractAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1")
	result, err := analyzer.AnalyzeContract(ctx, contractAddress, sampleContract)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Validate result
	assert.Equal(t, contractAddress, result.ContractAddress)
	assert.NotEmpty(t, result.AnalysisID)
	assert.True(t, result.OverallScore.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, result.OverallScore.LessThanOrEqual(decimal.NewFromFloat(100)))
	assert.Contains(t, []string{"A+", "A", "A-", "B+", "B", "B-", "C+", "C", "C-", "D", "F"}, result.SecurityGrade)
	assert.Contains(t, []string{"low", "medium", "high", "critical"}, result.RiskLevel)
	assert.True(t, result.Confidence.GreaterThan(decimal.Zero))
	assert.True(t, result.Confidence.LessThanOrEqual(decimal.NewFromFloat(1)))

	// Check analysis components
	assert.NotNil(t, result.SecurityPatterns)
	assert.NotNil(t, result.Vulnerabilities)
	assert.NotNil(t, result.CodeQuality)
	assert.NotNil(t, result.GasOptimization)

	// Check recommendations and warnings
	assert.NotNil(t, result.Recommendations)
	assert.NotNil(t, result.Warnings)
}

func TestSmartContractAnalyzer_AnalyzeVulnerableContract(t *testing.T) {
	logger := createTestLoggerForContract()
	config := GetDefaultContractAnalysisConfig()

	analyzer := NewSmartContractAnalyzer(logger, config)
	ctx := context.Background()

	// Start the analyzer
	err := analyzer.Start(ctx)
	require.NoError(t, err)
	defer analyzer.Stop()

	// Analyze vulnerable contract
	contractAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E2")
	result, err := analyzer.AnalyzeContract(ctx, contractAddress, vulnerableContract)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Should detect vulnerabilities
	vulnerabilityFound := false
	for _, vuln := range result.Vulnerabilities {
		if vuln.Found && vuln.Severity == "critical" {
			vulnerabilityFound = true
			break
		}
	}
	assert.True(t, vulnerabilityFound, "Should detect critical vulnerabilities in vulnerable contract")

	// Should have lower security score
	assert.True(t, result.OverallScore.LessThan(decimal.NewFromFloat(80)),
		"Vulnerable contract should have lower security score")

	// Should have warnings
	assert.NotEmpty(t, result.Warnings, "Vulnerable contract should generate warnings")
}

func TestSmartContractAnalyzer_GetAnalysisMetrics(t *testing.T) {
	logger := createTestLoggerForContract()
	config := GetDefaultContractAnalysisConfig()

	analyzer := NewSmartContractAnalyzer(logger, config)
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
	assert.Contains(t, metrics, "pattern_analyzer")
	assert.Contains(t, metrics, "vulnerability_scanner")
	assert.Contains(t, metrics, "formal_verifier")
	assert.Contains(t, metrics, "exploit_database")

	assert.Equal(t, true, metrics["is_running"])
	assert.Equal(t, true, metrics["pattern_analyzer"])
	assert.Equal(t, true, metrics["vulnerability_scanner"])
	assert.Equal(t, true, metrics["formal_verifier"])
	assert.Equal(t, true, metrics["exploit_database"])
}

func TestSmartContractAnalyzer_Caching(t *testing.T) {
	logger := createTestLoggerForContract()
	config := GetDefaultContractAnalysisConfig()
	config.CacheTimeout = 1 * time.Hour // Long cache timeout

	analyzer := NewSmartContractAnalyzer(logger, config)
	ctx := context.Background()

	// Start the analyzer
	err := analyzer.Start(ctx)
	require.NoError(t, err)
	defer analyzer.Stop()

	contractAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1")

	// First analysis
	result1, err := analyzer.AnalyzeContract(ctx, contractAddress, sampleContract)
	assert.NoError(t, err)
	assert.NotNil(t, result1)

	// Second analysis should return cached result
	result2, err := analyzer.AnalyzeContract(ctx, contractAddress, sampleContract)
	assert.NoError(t, err)
	assert.NotNil(t, result2)

	// Should be the same result (cached)
	assert.Equal(t, result1.AnalysisID, result2.AnalysisID)
	assert.Equal(t, result1.Timestamp, result2.Timestamp)
}

func TestGetDefaultContractAnalysisConfig(t *testing.T) {
	config := GetDefaultContractAnalysisConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, 5*time.Minute, config.MaxAnalysisTime)
	assert.Equal(t, 1000000, config.MaxCodeSize)
	assert.Equal(t, 1*time.Hour, config.CacheTimeout)

	// Check security patterns
	assert.NotEmpty(t, config.SecurityPatterns)
	assert.GreaterOrEqual(t, len(config.SecurityPatterns), 3)

	// Check vulnerability rules
	assert.NotEmpty(t, config.VulnerabilityRules)
	assert.GreaterOrEqual(t, len(config.VulnerabilityRules), 5)

	// Check scoring weights sum to 1.0
	weights := config.ScoringWeights
	totalWeight := weights.SecurityPatterns.Add(weights.Vulnerabilities).
		Add(weights.FormalVerification).Add(weights.ExploitHistory).
		Add(weights.CodeQuality).Add(weights.GasOptimization)
	assert.True(t, totalWeight.Equal(decimal.NewFromFloat(1.0)))

	// Check compiler versions
	assert.NotEmpty(t, config.CompilerVersions)
	assert.Contains(t, config.CompilerVersions, "0.8.0")

	// Check optimization checks
	assert.NotEmpty(t, config.OptimizationChecks)
	assert.Contains(t, config.OptimizationChecks, "storage_packing")
}

func TestValidateContractAnalysisConfig(t *testing.T) {
	// Test valid config
	validConfig := GetDefaultContractAnalysisConfig()
	err := ValidateContractAnalysisConfig(validConfig)
	assert.NoError(t, err)

	// Test disabled config
	disabledConfig := GetDefaultContractAnalysisConfig()
	disabledConfig.Enabled = false
	err = ValidateContractAnalysisConfig(disabledConfig)
	assert.NoError(t, err)

	// Test invalid configs
	invalidConfigs := []ContractAnalysisConfig{
		// Invalid max analysis time
		{
			Enabled:         true,
			MaxAnalysisTime: 0,
		},
		// Invalid max code size
		{
			Enabled:         true,
			MaxAnalysisTime: 5 * time.Minute,
			MaxCodeSize:     0,
		},
		// Invalid cache timeout
		{
			Enabled:         true,
			MaxAnalysisTime: 5 * time.Minute,
			MaxCodeSize:     1000000,
			CacheTimeout:    0,
		},
	}

	for i, config := range invalidConfigs {
		err := ValidateContractAnalysisConfig(config)
		assert.Error(t, err, "Config %d should be invalid", i)
	}
}

func TestContractUtilityFunctions(t *testing.T) {
	// Test security pattern categories
	categories := GetSecurityPatternsByCategory()
	assert.NotEmpty(t, categories)
	assert.Contains(t, categories, "reentrancy_protection")
	assert.Contains(t, categories, "access_control")

	// Test vulnerability rules by severity
	severities := GetVulnerabilityRulesBySeverity()
	assert.NotEmpty(t, severities)
	assert.Contains(t, severities, "critical")
	assert.Contains(t, severities, "high")

	// Test critical vulnerability rules
	critical := GetCriticalVulnerabilityRules()
	assert.NotEmpty(t, critical)
	for _, rule := range critical {
		assert.Equal(t, "critical", rule.Severity)
	}

	// Test required security patterns
	required := GetRequiredSecurityPatterns()
	assert.NotEmpty(t, required)
	for _, pattern := range required {
		assert.True(t, pattern.Required)
	}

	// Test security grade descriptions
	gradeDescriptions := GetSecurityGradeDescription()
	assert.NotEmpty(t, gradeDescriptions)
	assert.Contains(t, gradeDescriptions, "A+")
	assert.Contains(t, gradeDescriptions, "F")

	// Test vulnerability severity descriptions
	severityDescriptions := GetVulnerabilitySeverityDescription()
	assert.NotEmpty(t, severityDescriptions)
	assert.Contains(t, severityDescriptions, "critical")
	assert.Contains(t, severityDescriptions, "low")

	// Test supported compiler versions
	compilerVersions := GetSupportedCompilerVersions()
	assert.NotEmpty(t, compilerVersions)
	assert.Contains(t, compilerVersions, "0.8.0")

	// Test gas optimization checks
	optimizationChecks := GetGasOptimizationChecks()
	assert.NotEmpty(t, optimizationChecks)
	assert.Contains(t, optimizationChecks, "storage_packing")
}
