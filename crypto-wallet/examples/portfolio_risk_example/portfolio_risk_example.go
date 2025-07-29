package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/ai/risk"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

// Sample portfolios for demonstration
func createDiversifiedPortfolio() *risk.Portfolio {
	return &risk.Portfolio{
		ID:         "diversified_portfolio",
		Address:    common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1"),
		TotalValue: decimal.NewFromFloat(500000), // $500k portfolio
		Assets: []*risk.PortfolioAsset{
			{
				Symbol:      "BTC",
				Name:        "Bitcoin",
				Amount:      decimal.NewFromFloat(5),
				Value:       decimal.NewFromFloat(150000),
				Weight:      decimal.NewFromFloat(0.3),
				Price:       decimal.NewFromFloat(30000),
				Chain:       "bitcoin",
				Protocol:    "bitcoin",
				Sector:      "layer1",
				AssetType:   "major_crypto",
				LastUpdated: time.Now(),
			},
			{
				Symbol:      "ETH",
				Name:        "Ethereum",
				Amount:      decimal.NewFromFloat(50),
				Value:       decimal.NewFromFloat(100000),
				Weight:      decimal.NewFromFloat(0.2),
				Price:       decimal.NewFromFloat(2000),
				Chain:       "ethereum",
				Protocol:    "ethereum",
				Sector:      "layer1",
				AssetType:   "major_crypto",
				LastUpdated: time.Now(),
			},
			{
				Symbol:      "USDC",
				Name:        "USD Coin",
				Amount:      decimal.NewFromFloat(75000),
				Value:       decimal.NewFromFloat(75000),
				Weight:      decimal.NewFromFloat(0.15),
				Price:       decimal.NewFromFloat(1),
				Chain:       "ethereum",
				Protocol:    "centre",
				Sector:      "stablecoin",
				AssetType:   "stablecoin",
				LastUpdated: time.Now(),
			},
			{
				Symbol:      "UNI",
				Name:        "Uniswap",
				Amount:      decimal.NewFromFloat(5000),
				Value:       decimal.NewFromFloat(50000),
				Weight:      decimal.NewFromFloat(0.1),
				Price:       decimal.NewFromFloat(10),
				Chain:       "ethereum",
				Protocol:    "uniswap",
				Sector:      "defi",
				AssetType:   "defi_token",
				LastUpdated: time.Now(),
			},
			{
				Symbol:      "AAVE",
				Name:        "Aave",
				Amount:      decimal.NewFromFloat(500),
				Value:       decimal.NewFromFloat(50000),
				Weight:      decimal.NewFromFloat(0.1),
				Price:       decimal.NewFromFloat(100),
				Chain:       "ethereum",
				Protocol:    "aave",
				Sector:      "defi",
				AssetType:   "defi_token",
				LastUpdated: time.Now(),
			},
			{
				Symbol:      "MATIC",
				Name:        "Polygon",
				Amount:      decimal.NewFromFloat(50000),
				Value:       decimal.NewFromFloat(50000),
				Weight:      decimal.NewFromFloat(0.1),
				Price:       decimal.NewFromFloat(1),
				Chain:       "polygon",
				Protocol:    "polygon",
				Sector:      "layer2",
				AssetType:   "altcoin",
				LastUpdated: time.Now(),
			},
			{
				Symbol:      "SOL",
				Name:        "Solana",
				Amount:      decimal.NewFromFloat(1000),
				Value:       decimal.NewFromFloat(25000),
				Weight:      decimal.NewFromFloat(0.05),
				Price:       decimal.NewFromFloat(25),
				Chain:       "solana",
				Protocol:    "solana",
				Sector:      "layer1",
				AssetType:   "altcoin",
				LastUpdated: time.Now(),
			},
		},
		LastUpdated: time.Now(),
		Metadata:    make(map[string]interface{}),
	}
}

func createConcentratedPortfolio() *risk.Portfolio {
	return &risk.Portfolio{
		ID:         "concentrated_portfolio",
		Address:    common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E2"),
		TotalValue: decimal.NewFromFloat(200000), // $200k portfolio
		Assets: []*risk.PortfolioAsset{
			{
				Symbol:      "BTC",
				Name:        "Bitcoin",
				Amount:      decimal.NewFromFloat(5.33),
				Value:       decimal.NewFromFloat(160000),
				Weight:      decimal.NewFromFloat(0.8), // 80% concentration
				Price:       decimal.NewFromFloat(30000),
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
				Value:       decimal.NewFromFloat(40000),
				Weight:      decimal.NewFromFloat(0.2),
				Price:       decimal.NewFromFloat(2000),
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

func createDeFiPortfolio() *risk.Portfolio {
	return &risk.Portfolio{
		ID:         "defi_portfolio",
		Address:    common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E3"),
		TotalValue: decimal.NewFromFloat(300000), // $300k DeFi portfolio
		Assets: []*risk.PortfolioAsset{
			{
				Symbol:      "UNI",
				Name:        "Uniswap",
				Amount:      decimal.NewFromFloat(6000),
				Value:       decimal.NewFromFloat(90000),
				Weight:      decimal.NewFromFloat(0.3),
				Price:       decimal.NewFromFloat(15),
				Chain:       "ethereum",
				Protocol:    "uniswap",
				Sector:      "defi",
				AssetType:   "defi_token",
				LastUpdated: time.Now(),
			},
			{
				Symbol:      "AAVE",
				Name:        "Aave",
				Amount:      decimal.NewFromFloat(600),
				Value:       decimal.NewFromFloat(60000),
				Weight:      decimal.NewFromFloat(0.2),
				Price:       decimal.NewFromFloat(100),
				Chain:       "ethereum",
				Protocol:    "aave",
				Sector:      "defi",
				AssetType:   "defi_token",
				LastUpdated: time.Now(),
			},
			{
				Symbol:      "CRV",
				Name:        "Curve DAO Token",
				Amount:      decimal.NewFromFloat(60000),
				Value:       decimal.NewFromFloat(60000),
				Weight:      decimal.NewFromFloat(0.2),
				Price:       decimal.NewFromFloat(1),
				Chain:       "ethereum",
				Protocol:    "curve",
				Sector:      "defi",
				AssetType:   "defi_token",
				LastUpdated: time.Now(),
			},
			{
				Symbol:      "COMP",
				Name:        "Compound",
				Amount:      decimal.NewFromFloat(1000),
				Value:       decimal.NewFromFloat(60000),
				Weight:      decimal.NewFromFloat(0.2),
				Price:       decimal.NewFromFloat(60),
				Chain:       "ethereum",
				Protocol:    "compound",
				Sector:      "defi",
				AssetType:   "defi_token",
				LastUpdated: time.Now(),
			},
			{
				Symbol:      "SUSHI",
				Name:        "SushiSwap",
				Amount:      decimal.NewFromFloat(30000),
				Value:       decimal.NewFromFloat(30000),
				Weight:      decimal.NewFromFloat(0.1),
				Price:       decimal.NewFromFloat(1),
				Chain:       "ethereum",
				Protocol:    "sushiswap",
				Sector:      "defi",
				AssetType:   "defi_token",
				LastUpdated: time.Now(),
			},
		},
		LastUpdated: time.Now(),
		Metadata:    make(map[string]interface{}),
	}
}

func main() {
	fmt.Println("📊 Portfolio Risk Assessment System Example")
	fmt.Println("==========================================")

	// Initialize logger
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	logger := logger.NewLogger(logConfig)

	// Create portfolio risk analyzer configuration
	config := risk.GetDefaultPortfolioRiskAnalyzerConfig()

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Enabled: %v\n", config.Enabled)
	fmt.Printf("  Update Interval: %v\n", config.UpdateInterval)
	fmt.Printf("  Cache Timeout: %v\n", config.CacheTimeout)
	fmt.Printf("  History Window: %d days\n", config.HistoryWindow)
	fmt.Printf("  Confidence Level: %s\n", config.ConfidenceLevel.String())
	fmt.Printf("  Risk-Free Rate: %s%%\n", config.RiskFreeRate.Mul(decimal.NewFromFloat(100)).String())
	fmt.Printf("  Correlation Analysis: %v\n", config.CorrelationConfig.Enabled)
	fmt.Printf("  Diversification Analysis: %v\n", config.DiversificationConfig.Enabled)
	fmt.Printf("  VaR Analysis: %v\n", config.VaRConfig.Enabled)
	fmt.Printf("  Risk Metrics: %v\n", config.RiskMetricsConfig.Enabled)
	fmt.Println()

	// Create portfolio risk analyzer
	analyzer := risk.NewPortfolioRiskAnalyzer(logger, config)

	// Start the analyzer
	ctx := context.Background()
	if err := analyzer.Start(ctx); err != nil {
		fmt.Printf("Failed to start portfolio risk analyzer: %v\n", err)
		return
	}

	fmt.Println("✅ Portfolio risk analyzer started successfully!")
	fmt.Println()

	// Show analysis metrics
	fmt.Println("📊 Analysis Engine Status:")
	metrics := analyzer.GetAnalysisMetrics()
	fmt.Printf("  Running: %v\n", metrics["is_running"])
	fmt.Printf("  Correlation Analyzer: %v\n", metrics["correlation_analyzer"])
	fmt.Printf("  Diversification Engine: %v\n", metrics["diversification_engine"])
	fmt.Printf("  VaR Calculator: %v\n", metrics["var_calculator"])
	fmt.Printf("  Risk Metrics Engine: %v\n", metrics["risk_metrics_engine"])
	fmt.Printf("  Cached Analyses: %v\n", metrics["cached_analyses"])
	fmt.Printf("  Price History Assets: %v\n", metrics["price_history_assets"])
	fmt.Println()

	// Analyze different portfolio types
	portfolios := []*risk.Portfolio{
		createDiversifiedPortfolio(),
		createConcentratedPortfolio(),
		createDeFiPortfolio(),
	}

	portfolioNames := []string{
		"Well-Diversified Portfolio",
		"Concentrated Portfolio",
		"DeFi-Focused Portfolio",
	}

	var analyses []*risk.PortfolioRiskAnalysis

	for i, portfolio := range portfolios {
		fmt.Printf("🔍 Analyzing %s:\n", portfolioNames[i])
		fmt.Println(strings.Repeat("=", 50))
		
		fmt.Printf("Portfolio ID: %s\n", portfolio.ID)
		fmt.Printf("Address: %s\n", portfolio.Address.Hex())
		fmt.Printf("Total Value: $%s\n", portfolio.TotalValue.String())
		fmt.Printf("Asset Count: %d\n", len(portfolio.Assets))
		fmt.Println()

		fmt.Println("Assets:")
		for _, asset := range portfolio.Assets {
			fmt.Printf("  %s (%s): $%s (%.1f%%) - %s/%s\n",
				asset.Symbol, asset.Name,
				asset.Value.String(),
				asset.Weight.Mul(decimal.NewFromFloat(100)).InexactFloat64(),
				asset.Sector, asset.Chain)
		}
		fmt.Println()

		fmt.Println("🔄 Performing comprehensive risk analysis...")
		analysis, err := analyzer.AnalyzePortfolioRisk(ctx, portfolio)
		if err != nil {
			fmt.Printf("Failed to analyze portfolio: %v\n", err)
			continue
		}

		analyses = append(analyses, analysis)

		fmt.Println("✅ Analysis completed!")
		fmt.Println()

		// Display analysis results
		displayPortfolioAnalysis(portfolioNames[i], analysis)
		fmt.Println("\n" + strings.Repeat("=", 80) + "\n")
	}

	// Comparative analysis
	if len(analyses) > 1 {
		fmt.Println("📊 Comparative Portfolio Analysis:")
		fmt.Println("=================================")
		
		fmt.Printf("%-25s %-15s %-15s %-15s %-10s %-10s\n",
			"Portfolio", "Risk Score", "Risk Level", "Concentration", "Alerts", "Confidence")
		fmt.Println(strings.Repeat("-", 100))
		
		for i, analysis := range analyses {
			concentration := "N/A"
			if analysis.DiversificationMetrics != nil {
				concentration = fmt.Sprintf("%.1f%%", 
					analysis.DiversificationMetrics.ConcentrationRisk.Mul(decimal.NewFromFloat(100)).InexactFloat64())
			}
			
			fmt.Printf("%-25s %-15s %-15s %-15s %-10d %-10s\n",
				portfolioNames[i],
				analysis.OverallRiskScore.StringFixed(1),
				analysis.RiskLevel,
				concentration,
				len(analysis.RiskAlerts),
				analysis.Confidence.Mul(decimal.NewFromFloat(100)).StringFixed(1)+"%")
		}
		fmt.Println()

		// Risk ranking
		fmt.Println("🏆 Risk Ranking (Best to Worst):")
		type portfolioRank struct {
			name  string
			score decimal.Decimal
		}
		
		var rankings []portfolioRank
		for i, analysis := range analyses {
			rankings = append(rankings, portfolioRank{
				name:  portfolioNames[i],
				score: analysis.OverallRiskScore,
			})
		}
		
		// Sort by score (higher is better)
		for i := 0; i < len(rankings)-1; i++ {
			for j := i + 1; j < len(rankings); j++ {
				if rankings[j].score.GreaterThan(rankings[i].score) {
					rankings[i], rankings[j] = rankings[j], rankings[i]
				}
			}
		}
		
		for i, rank := range rankings {
			medal := "🥉"
			if i == 0 {
				medal = "🥇"
			} else if i == 1 {
				medal = "🥈"
			}
			fmt.Printf("  %s %d. %s (Score: %s)\n", medal, i+1, rank.name, rank.score.StringFixed(1))
		}
		fmt.Println()
	}

	// Performance metrics summary
	fmt.Println("⚡ Performance Summary:")
	fmt.Println("======================")
	totalAnalysisTime := time.Duration(0)
	for i, analysis := range analyses {
		fmt.Printf("%s Analysis Time: %v\n", portfolioNames[i], analysis.AnalysisDuration)
		totalAnalysisTime += analysis.AnalysisDuration
	}
	if len(analyses) > 0 {
		avgTime := totalAnalysisTime / time.Duration(len(analyses))
		fmt.Printf("Average Analysis Time: %v\n", avgTime)
	}
	fmt.Println()

	// Risk management insights
	fmt.Println("💡 Risk Management Insights:")
	fmt.Println("============================")
	fmt.Println("1. Diversification Benefits:")
	fmt.Println("   • Well-diversified portfolios typically show lower concentration risk")
	fmt.Println("   • Multi-chain exposure reduces blockchain-specific risks")
	fmt.Println("   • Sector diversification helps mitigate sector-specific downturns")
	fmt.Println()
	
	fmt.Println("2. Correlation Analysis:")
	fmt.Println("   • High correlation between assets reduces diversification benefits")
	fmt.Println("   • DeFi tokens often show high correlation during market stress")
	fmt.Println("   • Stablecoins provide low correlation and stability")
	fmt.Println()
	
	fmt.Println("3. Risk Metrics:")
	fmt.Println("   • Sharpe ratio measures risk-adjusted returns")
	fmt.Println("   • VaR quantifies potential losses at specific confidence levels")
	fmt.Println("   • Maximum drawdown shows worst-case historical performance")
	fmt.Println()

	fmt.Println("🎉 Portfolio Risk Assessment example completed!")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("  ✅ Comprehensive portfolio risk analysis")
	fmt.Println("  ✅ Multi-dimensional risk scoring")
	fmt.Println("  ✅ Correlation and diversification analysis")
	fmt.Println("  ✅ Value at Risk (VaR) calculations")
	fmt.Println("  ✅ Performance metrics and risk-adjusted returns")
	fmt.Println("  ✅ Intelligent rebalancing recommendations")
	fmt.Println("  ✅ Real-time risk alerts and warnings")
	fmt.Println("  ✅ Comparative portfolio analysis")
	fmt.Println("  ✅ Sector and chain diversification metrics")
	fmt.Println()
	fmt.Println("Note: This example demonstrates the system with mock data.")
	fmt.Println("Configure real price feeds and historical data for production use.")

	// Stop the analyzer
	if err := analyzer.Stop(); err != nil {
		fmt.Printf("Error stopping portfolio risk analyzer: %v\n", err)
	} else {
		fmt.Println("\n🛑 Portfolio risk analyzer stopped")
	}
}

func displayPortfolioAnalysis(title string, analysis *risk.PortfolioRiskAnalysis) {
	fmt.Printf("📋 %s Analysis Results:\n", title)
	fmt.Printf("  Analysis ID: %s\n", analysis.AnalysisID)
	fmt.Printf("  Overall Risk Score: %s/100\n", analysis.OverallRiskScore.StringFixed(1))
	fmt.Printf("  Risk Level: %s\n", analysis.RiskLevel)
	fmt.Printf("  Confidence: %s%%\n", analysis.Confidence.Mul(decimal.NewFromFloat(100)).StringFixed(1))
	fmt.Printf("  Analysis Duration: %v\n", analysis.AnalysisDuration)
	fmt.Println()

	// Diversification metrics
	if analysis.DiversificationMetrics != nil {
		fmt.Printf("🎯 Diversification Analysis:\n")
		fmt.Printf("  Concentration Risk: %s%%\n", 
			analysis.DiversificationMetrics.ConcentrationRisk.Mul(decimal.NewFromFloat(100)).StringFixed(1))
		fmt.Printf("  Diversification Score: %s/100\n", 
			analysis.DiversificationMetrics.DiversificationScore.StringFixed(1))
		fmt.Printf("  Effective Asset Count: %s\n", 
			analysis.DiversificationMetrics.EffectiveAssetCount.StringFixed(1))
		fmt.Printf("  Sector Diversification: %s%%\n", 
			analysis.DiversificationMetrics.SectorDiversification.Mul(decimal.NewFromFloat(100)).StringFixed(1))
		fmt.Printf("  Chain Diversification: %s%%\n", 
			analysis.DiversificationMetrics.ChainDiversification.Mul(decimal.NewFromFloat(100)).StringFixed(1))
		fmt.Println()
	}

	// Correlation analysis
	if analysis.CorrelationAnalysis != nil {
		fmt.Printf("🔗 Correlation Analysis:\n")
		fmt.Printf("  Average Correlation: %s\n", 
			analysis.CorrelationAnalysis.AverageCorrelation.StringFixed(3))
		fmt.Printf("  Max Correlation: %s\n", 
			analysis.CorrelationAnalysis.MaxCorrelation.StringFixed(3))
		fmt.Printf("  Correlation Risk: %s%%\n", 
			analysis.CorrelationAnalysis.CorrelationRisk.Mul(decimal.NewFromFloat(100)).StringFixed(1))
		fmt.Printf("  Highly Correlated Pairs: %d\n", 
			len(analysis.CorrelationAnalysis.HighlyCorrelatedPairs))
		fmt.Println()
	}

	// VaR analysis
	if analysis.VaRAnalysis != nil {
		fmt.Printf("📉 Value at Risk (VaR):\n")
		if var95, exists := analysis.VaRAnalysis.HistoricalVaR["95"]; exists {
			fmt.Printf("  95%% VaR: %s%%\n", var95.Abs().Mul(decimal.NewFromFloat(100)).StringFixed(2))
		}
		if var99, exists := analysis.VaRAnalysis.HistoricalVaR["99"]; exists {
			fmt.Printf("  99%% VaR: %s%%\n", var99.Abs().Mul(decimal.NewFromFloat(100)).StringFixed(2))
		}
		fmt.Printf("  Max Drawdown: %s%%\n", 
			analysis.VaRAnalysis.MaxDrawdown.Mul(decimal.NewFromFloat(100)).StringFixed(1))
		fmt.Println()
	}

	// Risk metrics
	if analysis.RiskMetrics != nil {
		fmt.Printf("📊 Risk Metrics:\n")
		fmt.Printf("  Sharpe Ratio: %s\n", analysis.RiskMetrics.SharpeRatio.StringFixed(2))
		fmt.Printf("  Sortino Ratio: %s\n", analysis.RiskMetrics.SortinoRatio.StringFixed(2))
		fmt.Printf("  Volatility: %s%%\n", 
			analysis.RiskMetrics.Volatility.Mul(decimal.NewFromFloat(100)).StringFixed(1))
		fmt.Printf("  Beta: %s\n", analysis.RiskMetrics.Beta.StringFixed(2))
		fmt.Println()
	}

	// Rebalancing advice
	if analysis.RebalancingAdvice != nil {
		fmt.Printf("⚖️ Rebalancing Advice:\n")
		fmt.Printf("  Should Rebalance: %v\n", analysis.RebalancingAdvice.ShouldRebalance)
		if analysis.RebalancingAdvice.ShouldRebalance {
			fmt.Printf("  Reason: %s\n", analysis.RebalancingAdvice.RebalanceReason)
			fmt.Printf("  Expected Improvement: %s%%\n", 
				analysis.RebalancingAdvice.ExpectedImprovement.Mul(decimal.NewFromFloat(100)).StringFixed(1))
			fmt.Printf("  Rebalance Actions: %d\n", len(analysis.RebalancingAdvice.RebalanceActions))
		}
		fmt.Println()
	}

	// Risk alerts
	if len(analysis.RiskAlerts) > 0 {
		fmt.Printf("🚨 Risk Alerts (%d):\n", len(analysis.RiskAlerts))
		for _, alert := range analysis.RiskAlerts {
			severityIcon := getSeverityIconForPortfolio(alert.Severity)
			fmt.Printf("  %s %s: %s\n", severityIcon, alert.Title, alert.Message)
		}
		fmt.Println()
	}

	// Recommendations
	if len(analysis.Recommendations) > 0 {
		fmt.Printf("💡 Recommendations:\n")
		for i, rec := range analysis.Recommendations {
			fmt.Printf("  %d. %s\n", i+1, rec)
		}
		fmt.Println()
	}
}

func getSeverityIconForPortfolio(severity string) string {
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
