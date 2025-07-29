package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/ai/risk"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

// Sample smart contracts for demonstration
const secureContract = `
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract SecureVault is ReentrancyGuard, Ownable {
    mapping(address => uint256) private balances;
    
    event Deposit(address indexed user, uint256 amount);
    event Withdrawal(address indexed user, uint256 amount);
    
    modifier validAmount(uint256 amount) {
        require(amount > 0, "Amount must be positive");
        _;
    }
    
    function deposit() external payable validAmount(msg.value) {
        balances[msg.sender] += msg.value;
        emit Deposit(msg.sender, msg.value);
    }
    
    function withdraw(uint256 amount) external nonReentrant validAmount(amount) {
        require(balances[msg.sender] >= amount, "Insufficient balance");
        
        balances[msg.sender] -= amount;
        
        (bool success, ) = payable(msg.sender).call{value: amount}("");
        require(success, "Transfer failed");
        
        emit Withdrawal(msg.sender, amount);
    }
    
    function getBalance(address user) external view returns (uint256) {
        return balances[user];
    }
    
    function emergencyWithdraw() external onlyOwner {
        uint256 balance = address(this).balance;
        require(balance > 0, "No funds to withdraw");
        
        (bool success, ) = payable(owner()).call{value: balance}("");
        require(success, "Emergency withdrawal failed");
    }
}
`

const vulnerableContract = `
pragma solidity ^0.7.0;

contract VulnerableBank {
    mapping(address => uint) public balances;
    
    function deposit() public payable {
        balances[msg.sender] += msg.value;
    }
    
    function withdraw(uint amount) public {
        require(balances[msg.sender] >= amount);
        
        msg.sender.call{value: amount}("");
        balances[msg.sender] -= amount;
    }
    
    function getBalance() public view returns (uint) {
        return balances[msg.sender];
    }
}
`

func main() {
	fmt.Println("🔍 Smart Contract Security Audit System Example")
	fmt.Println("===============================================")

	// Initialize logger
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	logger := logger.NewLogger(logConfig)

	// Create contract analysis configuration
	config := risk.GetDefaultContractAnalysisConfig()

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Enabled: %v\n", config.Enabled)
	fmt.Printf("  Max Analysis Time: %v\n", config.MaxAnalysisTime)
	fmt.Printf("  Max Code Size: %d bytes\n", config.MaxCodeSize)
	fmt.Printf("  Cache Timeout: %v\n", config.CacheTimeout)
	fmt.Printf("  Security Patterns: %d\n", len(config.SecurityPatterns))
	fmt.Printf("  Vulnerability Rules: %d\n", len(config.VulnerabilityRules))
	fmt.Printf("  Formal Verification: %v\n", config.FormalVerification.Enabled)
	fmt.Printf("  Exploit Database: %v\n", config.ExploitDatabase.Enabled)
	fmt.Println()

	// Create smart contract analyzer
	analyzer := risk.NewSmartContractAnalyzer(logger, config)

	// Start the analyzer
	ctx := context.Background()
	if err := analyzer.Start(ctx); err != nil {
		fmt.Printf("Failed to start contract analyzer: %v\n", err)
		return
	}

	fmt.Println("✅ Smart contract analyzer started successfully!")
	fmt.Println()

	// Show analysis metrics
	fmt.Println("📊 Analysis Engine Status:")
	metrics := analyzer.GetAnalysisMetrics()
	fmt.Printf("  Running: %v\n", metrics["is_running"])
	fmt.Printf("  Pattern Analyzer: %v\n", metrics["pattern_analyzer"])
	fmt.Printf("  Vulnerability Scanner: %v\n", metrics["vulnerability_scanner"])
	fmt.Printf("  Formal Verifier: %v\n", metrics["formal_verifier"])
	fmt.Printf("  Exploit Database: %v\n", metrics["exploit_database"])
	fmt.Printf("  Cached Analyses: %v\n", metrics["cached_analyses"])
	fmt.Println()

	// Analyze secure contract
	fmt.Println("🔒 Analyzing Secure Contract:")
	fmt.Println("=============================")

	secureAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1")
	fmt.Printf("Contract Address: %s\n", secureAddress.Hex())
	fmt.Printf("Contract Size: %d bytes\n", len(secureContract))
	fmt.Println()

	fmt.Println("🔄 Performing comprehensive security analysis...")
	secureResult, err := analyzer.AnalyzeContract(ctx, secureAddress, secureContract)
	if err != nil {
		fmt.Printf("Failed to analyze secure contract: %v\n", err)
		return
	}

	fmt.Println("✅ Analysis completed!")
	fmt.Println()

	// Display secure contract results
	displayAnalysisResults("Secure Contract", secureResult)

	fmt.Println("\n" + strings.Repeat("=", 80) + "\n")

	// Analyze vulnerable contract
	fmt.Println("⚠️  Analyzing Vulnerable Contract:")
	fmt.Println("=================================")

	vulnerableAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E2")
	fmt.Printf("Contract Address: %s\n", vulnerableAddress.Hex())
	fmt.Printf("Contract Size: %d bytes\n", len(vulnerableContract))
	fmt.Println()

	fmt.Println("🔄 Performing comprehensive security analysis...")
	vulnerableResult, err := analyzer.AnalyzeContract(ctx, vulnerableAddress, vulnerableContract)
	if err != nil {
		fmt.Printf("Failed to analyze vulnerable contract: %v\n", err)
		return
	}

	fmt.Println("✅ Analysis completed!")
	fmt.Println()

	// Display vulnerable contract results
	displayAnalysisResults("Vulnerable Contract", vulnerableResult)

	// Comparison
	fmt.Println("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Println("📊 Security Comparison:")
	fmt.Println("======================")

	fmt.Printf("Secure Contract:\n")
	fmt.Printf("  Overall Score: %s/100\n", secureResult.OverallScore.String())
	fmt.Printf("  Security Grade: %s\n", secureResult.SecurityGrade)
	fmt.Printf("  Risk Level: %s\n", secureResult.RiskLevel)
	fmt.Printf("  Vulnerabilities Found: %d\n", countFoundVulnerabilities(secureResult.Vulnerabilities))
	fmt.Printf("  Critical Issues: %d\n", countCriticalVulnerabilities(secureResult.Vulnerabilities))
	fmt.Println()

	fmt.Printf("Vulnerable Contract:\n")
	fmt.Printf("  Overall Score: %s/100\n", vulnerableResult.OverallScore.String())
	fmt.Printf("  Security Grade: %s\n", vulnerableResult.SecurityGrade)
	fmt.Printf("  Risk Level: %s\n", vulnerableResult.RiskLevel)
	fmt.Printf("  Vulnerabilities Found: %d\n", countFoundVulnerabilities(vulnerableResult.Vulnerabilities))
	fmt.Printf("  Critical Issues: %d\n", countCriticalVulnerabilities(vulnerableResult.Vulnerabilities))
	fmt.Println()

	scoreDiff := secureResult.OverallScore.Sub(vulnerableResult.OverallScore)
	fmt.Printf("Security Score Difference: %s points\n", scoreDiff.String())
	fmt.Println()

	// Show security patterns analysis
	fmt.Println("🛡️ Security Patterns Analysis:")
	fmt.Println("==============================")

	securePatterns := countFoundPatterns(secureResult.SecurityPatterns)
	vulnerablePatterns := countFoundPatterns(vulnerableResult.SecurityPatterns)

	fmt.Printf("Secure Contract - Security Patterns Implemented: %d/%d\n",
		securePatterns, len(secureResult.SecurityPatterns))
	fmt.Printf("Vulnerable Contract - Security Patterns Implemented: %d/%d\n",
		vulnerablePatterns, len(vulnerableResult.SecurityPatterns))
	fmt.Println()

	// Show detailed pattern analysis
	fmt.Println("Pattern Implementation Details:")
	for _, pattern := range secureResult.SecurityPatterns {
		status := "❌"
		if pattern.Found {
			status = "✅"
		}
		fmt.Printf("  %s %s (%s)\n", status, pattern.Pattern.Name, pattern.Pattern.Category)
	}
	fmt.Println()

	// Performance metrics
	fmt.Println("⚡ Performance Metrics:")
	fmt.Println("======================")
	fmt.Printf("Secure Contract Analysis Time: %v\n", secureResult.AnalysisDuration)
	fmt.Printf("Vulnerable Contract Analysis Time: %v\n", vulnerableResult.AnalysisDuration)
	fmt.Printf("Average Analysis Time: %v\n",
		(secureResult.AnalysisDuration+vulnerableResult.AnalysisDuration)/2)
	fmt.Printf("Analysis Confidence (Secure): %s%%\n",
		secureResult.Confidence.Mul(decimal.NewFromInt(100)).String())
	fmt.Printf("Analysis Confidence (Vulnerable): %s%%\n",
		vulnerableResult.Confidence.Mul(decimal.NewFromInt(100)).String())
	fmt.Println()

	// Security recommendations
	fmt.Println("💡 Security Recommendations:")
	fmt.Println("============================")

	if len(vulnerableResult.Recommendations) > 0 {
		fmt.Println("For Vulnerable Contract:")
		for i, rec := range vulnerableResult.Recommendations {
			fmt.Printf("  %d. %s\n", i+1, rec)
		}
	} else {
		fmt.Println("For Vulnerable Contract: No specific recommendations (contract needs complete rewrite)")
	}
	fmt.Println()

	if len(secureResult.Recommendations) > 0 {
		fmt.Println("For Secure Contract:")
		for i, rec := range secureResult.Recommendations {
			fmt.Printf("  %d. %s\n", i+1, rec)
		}
	} else {
		fmt.Println("For Secure Contract: No additional recommendations - excellent security implementation!")
	}
	fmt.Println()

	fmt.Println("🎉 Smart Contract Security Audit example completed!")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("  ✅ Comprehensive security pattern analysis")
	fmt.Println("  ✅ Advanced vulnerability detection")
	fmt.Println("  ✅ Security scoring and grading")
	fmt.Println("  ✅ Code quality assessment")
	fmt.Println("  ✅ Gas optimization analysis")
	fmt.Println("  ✅ Risk level determination")
	fmt.Println("  ✅ Intelligent security recommendations")
	fmt.Println("  ✅ Performance metrics and confidence scoring")
	fmt.Println("  ✅ Comparative security analysis")
	fmt.Println()
	fmt.Println("Note: This example demonstrates the system with pattern-based analysis.")
	fmt.Println("Configure formal verification tools and exploit databases for enhanced analysis.")

	// Stop the analyzer
	if err := analyzer.Stop(); err != nil {
		fmt.Printf("Error stopping contract analyzer: %v\n", err)
	} else {
		fmt.Println("\n🛑 Smart contract analyzer stopped")
	}
}

func displayAnalysisResults(title string, result *risk.ContractAnalysisResult) {
	fmt.Printf("📋 %s Analysis Results:\n", title)
	fmt.Printf("  Analysis ID: %s\n", result.AnalysisID)
	fmt.Printf("  Overall Score: %s/100\n", result.OverallScore.String())
	fmt.Printf("  Security Grade: %s\n", result.SecurityGrade)
	fmt.Printf("  Risk Level: %s\n", result.RiskLevel)
	fmt.Printf("  Confidence: %s%%\n", result.Confidence.Mul(decimal.NewFromInt(100)).String())
	fmt.Printf("  Analysis Duration: %v\n", result.AnalysisDuration)
	fmt.Println()

	// Security patterns
	if len(result.SecurityPatterns) > 0 {
		fmt.Printf("🛡️ Security Patterns (%d found):\n", countFoundPatterns(result.SecurityPatterns))
		for _, pattern := range result.SecurityPatterns {
			if pattern.Found {
				fmt.Printf("  ✅ %s - %s\n", pattern.Pattern.Name, pattern.Pattern.Category)
			}
		}
		fmt.Println()
	}

	// Vulnerabilities
	if len(result.Vulnerabilities) > 0 {
		vulnCount := countFoundVulnerabilities(result.Vulnerabilities)
		if vulnCount > 0 {
			fmt.Printf("🚨 Vulnerabilities Found (%d):\n", vulnCount)
			for _, vuln := range result.Vulnerabilities {
				if vuln.Found {
					severityIcon := getSeverityIcon(vuln.Severity)
					fmt.Printf("  %s %s (%s) - Impact: %s\n",
						severityIcon, vuln.Rule.Name, vuln.Severity, vuln.Impact.String())
					if len(vuln.Locations) > 0 {
						fmt.Printf("    Locations: %v\n", vuln.Locations)
					}
				}
			}
		} else {
			fmt.Printf("✅ No vulnerabilities detected\n")
		}
		fmt.Println()
	}

	// Code quality
	if result.CodeQuality != nil {
		fmt.Printf("📊 Code Quality:\n")
		fmt.Printf("  Score: %s/100\n", result.CodeQuality.Score.String())
		fmt.Printf("  Lines of Code: %d\n", result.CodeQuality.LinesOfCode)
		fmt.Printf("  Complexity: %d\n", result.CodeQuality.Complexity)
		fmt.Printf("  Documentation: %s%%\n", result.CodeQuality.Documentation.String())
		fmt.Printf("  Compiler Version: %s\n", result.CodeQuality.CompilerVersion)
		fmt.Println()
	}

	// Gas optimization
	if result.GasOptimization != nil {
		fmt.Printf("⛽ Gas Optimization:\n")
		fmt.Printf("  Score: %s/100\n", result.GasOptimization.Score.String())
		fmt.Printf("  Efficiency Rating: %s\n", result.GasOptimization.EfficiencyRating)
		fmt.Printf("  Potential Savings: %s gas\n", result.GasOptimization.PotentialSavings.String())
		if len(result.GasOptimization.Optimizations) > 0 {
			fmt.Printf("  Optimizations Available: %d\n", len(result.GasOptimization.Optimizations))
		}
		fmt.Println()
	}

	// Warnings
	if len(result.Warnings) > 0 {
		fmt.Printf("⚠️ Warnings:\n")
		for _, warning := range result.Warnings {
			fmt.Printf("  • %s\n", warning)
		}
		fmt.Println()
	}
}

func countFoundPatterns(patterns []risk.SecurityPatternResult) int {
	count := 0
	for _, pattern := range patterns {
		if pattern.Found {
			count++
		}
	}
	return count
}

func countFoundVulnerabilities(vulnerabilities []risk.VulnerabilityResult) int {
	count := 0
	for _, vuln := range vulnerabilities {
		if vuln.Found {
			count++
		}
	}
	return count
}

func countCriticalVulnerabilities(vulnerabilities []risk.VulnerabilityResult) int {
	count := 0
	for _, vuln := range vulnerabilities {
		if vuln.Found && vuln.Severity == "critical" {
			count++
		}
	}
	return count
}

func getSeverityIcon(severity string) string {
	switch severity {
	case "critical":
		return "🚨"
	case "high":
		return "🔴"
	case "medium":
		return "🟡"
	case "low":
		return "🟢"
	default:
		return "ℹ️"
	}
}
