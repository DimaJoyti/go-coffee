package risk

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// GetDefaultContractAnalysisConfig returns default contract analysis configuration
func GetDefaultContractAnalysisConfig() ContractAnalysisConfig {
	return ContractAnalysisConfig{
		Enabled:         true,
		MaxAnalysisTime: 5 * time.Minute,
		MaxCodeSize:     1000000, // 1MB
		CacheTimeout:    1 * time.Hour,
		SecurityPatterns: []SecurityPattern{
			{
				ID:          "checks_effects_interactions",
				Name:        "Checks-Effects-Interactions",
				Description: "Proper ordering of checks, effects, and interactions",
				Pattern:     `require\s*\([^)]+\).*\n.*[^=]=[^=].*\n.*\.call`,
				Weight:      decimal.NewFromFloat(0.3),
				Category:    "reentrancy_protection",
				Severity:    "high",
				Required:    true,
			},
			{
				ID:          "access_control",
				Name:        "Access Control",
				Description: "Proper access control implementation",
				Pattern:     `modifier\s+\w+.*\{.*require\s*\(.*msg\.sender`,
				Weight:      decimal.NewFromFloat(0.25),
				Category:    "access_control",
				Severity:    "high",
				Required:    true,
			},
			{
				ID:          "reentrancy_guard",
				Name:        "Reentrancy Guard",
				Description: "Reentrancy protection mechanism",
				Pattern:     `nonReentrant|ReentrancyGuard|_status\s*=`,
				Weight:      decimal.NewFromFloat(0.2),
				Category:    "reentrancy_protection",
				Severity:    "high",
				Required:    false,
			},
			{
				ID:          "safe_math",
				Name:        "Safe Math",
				Description: "Safe arithmetic operations",
				Pattern:     `SafeMath|pragma\s+solidity\s+\^?0\.[8-9]`,
				Weight:      decimal.NewFromFloat(0.15),
				Category:    "arithmetic_safety",
				Severity:    "medium",
				Required:    false,
			},
			{
				ID:          "input_validation",
				Name:        "Input Validation",
				Description: "Proper input validation",
				Pattern:     `require\s*\([^)]*!=\s*0|require\s*\([^)]*>\s*0`,
				Weight:      decimal.NewFromFloat(0.1),
				Category:    "input_validation",
				Severity:    "medium",
				Required:    false,
			},
		},
		VulnerabilityRules: []VulnerabilityRule{
			{
				ID:          "reentrancy",
				Name:        "Reentrancy Vulnerability",
				Description: "Potential reentrancy attack vector",
				Pattern:     `\.call\{value:\s*\w+\}.*\n.*[^=]=[^=]`,
				Severity:    "critical",
				Impact:      decimal.NewFromFloat(25),
				Confidence:  decimal.NewFromFloat(0.8),
				Category:    "reentrancy",
				CWE:         "CWE-841",
				References:  []string{"https://consensys.github.io/smart-contract-best-practices/attacks/reentrancy/"},
			},
			{
				ID:          "integer_overflow",
				Name:        "Integer Overflow",
				Description: "Potential integer overflow vulnerability",
				Pattern:     `\+\+|\+=|--|-=|\*=|/=.*(?!SafeMath)(?!pragma\s+solidity\s+\^?0\.[8-9])`,
				Severity:    "high",
				Impact:      decimal.NewFromFloat(20),
				Confidence:  decimal.NewFromFloat(0.7),
				Category:    "integer_overflow",
				CWE:         "CWE-190",
				References:  []string{"https://consensys.github.io/smart-contract-best-practices/attacks/insecure-arithmetic/"},
			},
			{
				ID:          "unchecked_call",
				Name:        "Unchecked External Call",
				Description: "External call without checking return value",
				Pattern:     `\.call\(|\.send\(|\.transfer\(.*(?!\s*(require|assert|if))`,
				Severity:    "medium",
				Impact:      decimal.NewFromFloat(15),
				Confidence:  decimal.NewFromFloat(0.6),
				Category:    "unchecked_calls",
				CWE:         "CWE-252",
				References:  []string{"https://consensys.github.io/smart-contract-best-practices/recommendations/#handle-errors-in-external-calls"},
			},
			{
				ID:          "timestamp_dependence",
				Name:        "Timestamp Dependence",
				Description: "Dangerous use of block.timestamp",
				Pattern:     `block\.timestamp|now\s*[<>=]`,
				Severity:    "low",
				Impact:      decimal.NewFromFloat(10),
				Confidence:  decimal.NewFromFloat(0.9),
				Category:    "timestamp_dependence",
				CWE:         "CWE-829",
				References:  []string{"https://consensys.github.io/smart-contract-best-practices/recommendations/#timestamp-dependence"},
			},
			{
				ID:          "gas_limit",
				Name:        "Gas Limit Vulnerability",
				Description: "Potential gas limit issues",
				Pattern:     `for\s*\([^)]*;\s*\w+\s*<\s*\w+\.length`,
				Severity:    "medium",
				Impact:      decimal.NewFromFloat(12),
				Confidence:  decimal.NewFromFloat(0.5),
				Category:    "gas_limit",
				CWE:         "CWE-400",
				References:  []string{"https://consensys.github.io/smart-contract-best-practices/attacks/denial-of-service/"},
			},
			{
				ID:          "uninitialized_storage",
				Name:        "Uninitialized Storage Pointer",
				Description: "Uninitialized storage pointer vulnerability",
				Pattern:     `struct\s+\w+\s+\w+;(?!\s*\w+\s*=)`,
				Severity:    "high",
				Impact:      decimal.NewFromFloat(18),
				Confidence:  decimal.NewFromFloat(0.7),
				Category:    "uninitialized_storage",
				CWE:         "CWE-824",
				References:  []string{"https://consensys.github.io/smart-contract-best-practices/attacks/uninitialized-storage/"},
			},
		},
		ExploitDatabase: ExploitDatabaseConfig{
			Enabled:        false, // Disabled by default
			DatabaseURL:    "https://api.exploitdb.com/v1/",
			APIKey:         "",
			UpdateInterval: 24 * time.Hour,
			CacheSize:      1000,
		},
		FormalVerification: FormalVerificationConfig{
			Enabled:    false, // Disabled by default
			ToolPath:   "/usr/local/bin/solc-verify",
			Timeout:    2 * time.Minute,
			MaxMemory:  1024, // 1GB
			Properties: []string{"overflow", "underflow", "reentrancy", "access_control"},
		},
		ScoringWeights: ScoringWeights{
			SecurityPatterns:   decimal.NewFromFloat(0.25),
			Vulnerabilities:    decimal.NewFromFloat(0.30),
			FormalVerification: decimal.NewFromFloat(0.20),
			ExploitHistory:     decimal.NewFromFloat(0.15),
			CodeQuality:        decimal.NewFromFloat(0.05),
			GasOptimization:    decimal.NewFromFloat(0.05),
		},
		CompilerVersions: []string{
			"0.8.0", "0.8.1", "0.8.2", "0.8.3", "0.8.4", "0.8.5",
			"0.8.6", "0.8.7", "0.8.8", "0.8.9", "0.8.10", "0.8.11",
			"0.8.12", "0.8.13", "0.8.14", "0.8.15", "0.8.16", "0.8.17",
			"0.8.18", "0.8.19", "0.8.20", "0.8.21", "0.8.22", "0.8.23",
		},
		OptimizationChecks: []string{
			"storage_packing",
			"calldata_usage",
			"constant_variables",
			"immutable_variables",
			"short_circuit_evaluation",
			"loop_optimization",
			"function_visibility",
			"modifier_usage",
		},
	}
}

// GetSecurityPatternsByCategory returns security patterns grouped by category
func GetSecurityPatternsByCategory() map[string][]SecurityPattern {
	config := GetDefaultContractAnalysisConfig()
	categories := make(map[string][]SecurityPattern)

	for _, pattern := range config.SecurityPatterns {
		categories[pattern.Category] = append(categories[pattern.Category], pattern)
	}

	return categories
}

// GetVulnerabilityRulesBySeverity returns vulnerability rules grouped by severity
func GetVulnerabilityRulesBySeverity() map[string][]VulnerabilityRule {
	config := GetDefaultContractAnalysisConfig()
	severities := make(map[string][]VulnerabilityRule)

	for _, rule := range config.VulnerabilityRules {
		severities[rule.Severity] = append(severities[rule.Severity], rule)
	}

	return severities
}

// GetCriticalVulnerabilityRules returns only critical vulnerability rules
func GetCriticalVulnerabilityRules() []VulnerabilityRule {
	config := GetDefaultContractAnalysisConfig()
	var critical []VulnerabilityRule

	for _, rule := range config.VulnerabilityRules {
		if rule.Severity == "critical" {
			critical = append(critical, rule)
		}
	}

	return critical
}

// GetRequiredSecurityPatterns returns only required security patterns
func GetRequiredSecurityPatterns() []SecurityPattern {
	config := GetDefaultContractAnalysisConfig()
	var required []SecurityPattern

	for _, pattern := range config.SecurityPatterns {
		if pattern.Required {
			required = append(required, pattern)
		}
	}

	return required
}

// GetSupportedCompilerVersions returns supported Solidity compiler versions
func GetSupportedCompilerVersions() []string {
	config := GetDefaultContractAnalysisConfig()
	return config.CompilerVersions
}

// GetGasOptimizationChecks returns available gas optimization checks
func GetGasOptimizationChecks() []string {
	config := GetDefaultContractAnalysisConfig()
	return config.OptimizationChecks
}

// ValidateContractAnalysisConfig validates contract analysis configuration
func ValidateContractAnalysisConfig(config ContractAnalysisConfig) error {
	if !config.Enabled {
		return nil
	}

	if config.MaxAnalysisTime <= 0 {
		return fmt.Errorf("max analysis time must be positive")
	}

	if config.MaxCodeSize <= 0 {
		return fmt.Errorf("max code size must be positive")
	}

	if config.CacheTimeout <= 0 {
		return fmt.Errorf("cache timeout must be positive")
	}

	// Validate scoring weights sum to 1.0
	weights := config.ScoringWeights
	totalWeight := weights.SecurityPatterns.Add(weights.Vulnerabilities).
		Add(weights.FormalVerification).Add(weights.ExploitHistory).
		Add(weights.CodeQuality).Add(weights.GasOptimization)

	if !totalWeight.Equal(decimal.NewFromFloat(1.0)) {
		return fmt.Errorf("scoring weights must sum to 1.0, got %s", totalWeight.String())
	}

	// Validate security patterns
	for _, pattern := range config.SecurityPatterns {
		if pattern.ID == "" {
			return fmt.Errorf("security pattern ID cannot be empty")
		}
		if pattern.Weight.LessThan(decimal.Zero) || pattern.Weight.GreaterThan(decimal.NewFromFloat(1.0)) {
			return fmt.Errorf("security pattern weight must be between 0 and 1")
		}
	}

	// Validate vulnerability rules
	for _, rule := range config.VulnerabilityRules {
		if rule.ID == "" {
			return fmt.Errorf("vulnerability rule ID cannot be empty")
		}
		if rule.Impact.LessThan(decimal.Zero) || rule.Impact.GreaterThan(decimal.NewFromFloat(100)) {
			return fmt.Errorf("vulnerability impact must be between 0 and 100")
		}
		if rule.Confidence.LessThan(decimal.Zero) || rule.Confidence.GreaterThan(decimal.NewFromFloat(1.0)) {
			return fmt.Errorf("vulnerability confidence must be between 0 and 1")
		}
	}

	return nil
}

// GetSecurityGradeDescription returns description for security grades
func GetSecurityGradeDescription() map[string]string {
	return map[string]string{
		"A+": "Excellent security - Best practices implemented",
		"A":  "Very good security - Minor improvements possible",
		"A-": "Good security - Some areas for improvement",
		"B+": "Above average security - Several improvements recommended",
		"B":  "Average security - Multiple improvements needed",
		"B-": "Below average security - Significant improvements required",
		"C+": "Poor security - Major security issues present",
		"C":  "Very poor security - Critical issues need immediate attention",
		"C-": "Extremely poor security - High risk of exploitation",
		"D":  "Dangerous - Multiple critical vulnerabilities",
		"F":  "Fail - Severe security flaws, do not deploy",
	}
}

// GetVulnerabilitySeverityDescription returns description for vulnerability severities
func GetVulnerabilitySeverityDescription() map[string]string {
	return map[string]string{
		"critical": "Critical - Immediate exploitation possible, high impact",
		"high":     "High - Exploitation likely, significant impact",
		"medium":   "Medium - Exploitation possible under certain conditions",
		"low":      "Low - Limited exploitation potential, minimal impact",
		"info":     "Informational - No direct security impact",
	}
}

// GetSecurityPatternCategories returns available security pattern categories
func GetSecurityPatternCategories() []string {
	return []string{
		"reentrancy_protection",
		"access_control",
		"arithmetic_safety",
		"input_validation",
		"state_management",
		"external_calls",
		"gas_optimization",
		"upgrade_safety",
	}
}

// GetVulnerabilityCategories returns available vulnerability categories
func GetVulnerabilityCategories() []string {
	return []string{
		"reentrancy",
		"integer_overflow",
		"access_control",
		"unchecked_calls",
		"timestamp_dependence",
		"gas_limit",
		"uninitialized_storage",
		"front_running",
		"denial_of_service",
		"logic_errors",
	}
}
