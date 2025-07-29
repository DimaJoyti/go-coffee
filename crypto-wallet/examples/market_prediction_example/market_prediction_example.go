package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/ai/prediction"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/shopspring/decimal"
)

// Sample assets for demonstration
func getSampleAssets() map[string]decimal.Decimal {
	return map[string]decimal.Decimal{
		"BTC":  decimal.NewFromFloat(43250.75),  // Bitcoin
		"ETH":  decimal.NewFromFloat(2580.30),   // Ethereum
		"SOL":  decimal.NewFromFloat(98.45),     // Solana
		"ADA":  decimal.NewFromFloat(0.52),      // Cardano
		"DOT":  decimal.NewFromFloat(7.23),      // Polkadot
		"LINK": decimal.NewFromFloat(14.67),     // Chainlink
		"UNI":  decimal.NewFromFloat(6.89),      // Uniswap
		"AAVE": decimal.NewFromFloat(95.12),     // Aave
		"USDC": decimal.NewFromFloat(1.0001),    // USD Coin
		"MATIC": decimal.NewFromFloat(0.89),     // Polygon
	}
}

func main() {
	fmt.Println("🔮 AI-Powered Market Prediction System Example")
	fmt.Println("==============================================")

	// Initialize logger
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	logger := logger.NewLogger(logConfig)

	// Create market predictor configuration
	config := prediction.GetDefaultMarketPredictorConfig()

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Enabled: %v\n", config.Enabled)
	fmt.Printf("  Update Interval: %v\n", config.UpdateInterval)
	fmt.Printf("  Cache Timeout: %v\n", config.CacheTimeout)
	fmt.Printf("  Prediction Horizons: %v\n", config.PredictionHorizons)
	fmt.Printf("  Confidence Threshold: %s\n", config.ConfidenceThreshold.String())
	fmt.Printf("  Sentiment Analysis: %v\n", config.SentimentConfig.Enabled)
	fmt.Printf("  On-Chain Analysis: %v\n", config.OnChainConfig.Enabled)
	fmt.Printf("  Technical Analysis: %v\n", config.TechnicalConfig.Enabled)
	fmt.Printf("  Macro Analysis: %v\n", config.MacroConfig.Enabled)
	fmt.Printf("  Ensemble Model: %v\n", config.EnsembleConfig.Enabled)
	fmt.Printf("  Data Sources: %v\n", config.DataSources)
	fmt.Println()

	// Create market predictor
	predictor := prediction.NewMarketPredictor(logger, config)

	// Start the predictor
	ctx := context.Background()
	if err := predictor.Start(ctx); err != nil {
		fmt.Printf("Failed to start market predictor: %v\n", err)
		return
	}

	fmt.Println("✅ Market predictor started successfully!")
	fmt.Println()

	// Show prediction metrics
	fmt.Println("📊 Prediction Engine Status:")
	metrics := predictor.GetPredictionMetrics()
	fmt.Printf("  Running: %v\n", metrics["is_running"])
	fmt.Printf("  Sentiment Analyzer: %v\n", metrics["sentiment_analyzer"])
	fmt.Printf("  On-Chain Analyzer: %v\n", metrics["onchain_analyzer"])
	fmt.Printf("  Technical Analyzer: %v\n", metrics["technical_analyzer"])
	fmt.Printf("  Macro Analyzer: %v\n", metrics["macro_analyzer"])
	fmt.Printf("  Ensemble Model: %v\n", metrics["ensemble_model"])
	fmt.Printf("  Cached Predictions: %v\n", metrics["cached_predictions"])
	fmt.Printf("  Cached Models: %v\n", metrics["cached_models"])
	fmt.Println()

	// Get sample assets
	assets := getSampleAssets()

	// Single asset prediction example
	fmt.Println("🎯 Single Asset Prediction Example:")
	fmt.Println("===================================")
	
	btcPrice := assets["BTC"]
	fmt.Printf("Analyzing Bitcoin (BTC) at $%s\n", btcPrice.String())
	fmt.Println()

	fmt.Println("🔄 Performing comprehensive market prediction...")
	btcPrediction, err := predictor.PredictMarket(ctx, "BTC", btcPrice)
	if err != nil {
		fmt.Printf("Failed to predict BTC: %v\n", err)
		return
	}

	fmt.Println("✅ Prediction completed!")
	fmt.Println()

	// Display detailed prediction results
	displayDetailedPrediction("Bitcoin (BTC)", btcPrediction)

	fmt.Println("\n" + strings.Repeat("=", 80) + "\n")

	// Multiple asset predictions
	fmt.Println("📈 Multiple Asset Predictions:")
	fmt.Println("=============================")
	
	fmt.Printf("Analyzing %d assets simultaneously...\n", len(assets))
	fmt.Println()

	predictions, err := predictor.PredictMultipleAssets(ctx, assets)
	if err != nil {
		fmt.Printf("Failed to predict multiple assets: %v\n", err)
		return
	}

	fmt.Printf("✅ Generated predictions for %d assets!\n", len(predictions))
	fmt.Println()

	// Display prediction summary table
	fmt.Println("📊 Prediction Summary:")
	fmt.Printf("%-8s %-12s %-12s %-12s %-10s %-10s %-10s\n",
		"Asset", "Current", "Predicted", "Change %", "Direction", "Confidence", "Risk")
	fmt.Println(strings.Repeat("-", 85))

	// Sort assets for consistent display
	assetNames := []string{"BTC", "ETH", "SOL", "ADA", "DOT", "LINK", "UNI", "AAVE", "USDC", "MATIC"}
	
	for _, asset := range assetNames {
		if pred, exists := predictions[asset]; exists {
			changePercent := "N/A"
			if !pred.PriceChangePercent.IsZero() {
				changePercent = pred.PriceChangePercent.StringFixed(2) + "%"
			}
			
			fmt.Printf("%-8s $%-11s $%-11s %-12s %-10s %-10s %-10s\n",
				asset,
				pred.CurrentPrice.StringFixed(2),
				pred.PredictedPrice.StringFixed(2),
				changePercent,
				pred.Direction,
				pred.Confidence.Mul(decimal.NewFromFloat(100)).StringFixed(1)+"%",
				pred.RiskLevel)
		}
	}
	fmt.Println()

	// Analysis by direction
	fmt.Println("📊 Market Sentiment Analysis:")
	fmt.Println("=============================")
	
	bullishCount := 0
	bearishCount := 0
	neutralCount := 0
	totalConfidence := decimal.Zero
	
	for _, pred := range predictions {
		switch pred.Direction {
		case "bullish":
			bullishCount++
		case "bearish":
			bearishCount++
		case "neutral":
			neutralCount++
		}
		totalConfidence = totalConfidence.Add(pred.Confidence)
	}
	
	avgConfidence := totalConfidence.Div(decimal.NewFromInt(int64(len(predictions))))
	
	fmt.Printf("Market Sentiment Distribution:\n")
	fmt.Printf("  🟢 Bullish: %d assets (%.1f%%)\n", bullishCount, float64(bullishCount)/float64(len(predictions))*100)
	fmt.Printf("  🔴 Bearish: %d assets (%.1f%%)\n", bearishCount, float64(bearishCount)/float64(len(predictions))*100)
	fmt.Printf("  🟡 Neutral: %d assets (%.1f%%)\n", neutralCount, float64(neutralCount)/float64(len(predictions))*100)
	fmt.Printf("  📊 Average Confidence: %.1f%%\n", avgConfidence.Mul(decimal.NewFromFloat(100)).InexactFloat64())
	fmt.Println()

	// Top predictions by confidence
	fmt.Println("🏆 Top Predictions by Confidence:")
	fmt.Println("=================================")
	
	// Create slice for sorting
	type predictionPair struct {
		asset      string
		prediction *prediction.MarketPrediction
	}
	
	var predPairs []predictionPair
	for asset, pred := range predictions {
		predPairs = append(predPairs, predictionPair{asset, pred})
	}
	
	// Sort by confidence (descending)
	for i := 0; i < len(predPairs)-1; i++ {
		for j := i + 1; j < len(predPairs); j++ {
			if predPairs[j].prediction.Confidence.GreaterThan(predPairs[i].prediction.Confidence) {
				predPairs[i], predPairs[j] = predPairs[j], predPairs[i]
			}
		}
	}
	
	// Display top 5
	for i := 0; i < 5 && i < len(predPairs); i++ {
		pair := predPairs[i]
		pred := pair.prediction
		
		directionIcon := getDirectionIcon(pred.Direction)
		fmt.Printf("%d. %s %s - %s direction, %.1f%% confidence, %s change\n",
			i+1, directionIcon, pair.asset, pred.Direction,
			pred.Confidence.Mul(decimal.NewFromFloat(100)).InexactFloat64(),
			pred.PriceChangePercent.StringFixed(2)+"%")
	}
	fmt.Println()

	// Risk analysis
	fmt.Println("⚠️ Risk Analysis:")
	fmt.Println("=================")
	
	highRiskCount := 0
	mediumRiskCount := 0
	lowRiskCount := 0
	
	for _, pred := range predictions {
		switch pred.RiskLevel {
		case "high":
			highRiskCount++
		case "medium":
			mediumRiskCount++
		case "low":
			lowRiskCount++
		}
	}
	
	fmt.Printf("Risk Distribution:\n")
	fmt.Printf("  🔴 High Risk: %d assets\n", highRiskCount)
	fmt.Printf("  🟡 Medium Risk: %d assets\n", mediumRiskCount)
	fmt.Printf("  🟢 Low Risk: %d assets\n", lowRiskCount)
	fmt.Println()

	// Alerts summary
	fmt.Println("🚨 Prediction Alerts:")
	fmt.Println("=====================")
	
	totalAlerts := 0
	alertTypes := make(map[string]int)
	
	for asset, pred := range predictions {
		if len(pred.Alerts) > 0 {
			fmt.Printf("%s: %d alerts\n", asset, len(pred.Alerts))
			totalAlerts += len(pred.Alerts)
			
			for _, alert := range pred.Alerts {
				alertTypes[alert.Type]++
			}
		}
	}
	
	if totalAlerts > 0 {
		fmt.Printf("\nAlert Types:\n")
		for alertType, count := range alertTypes {
			fmt.Printf("  %s: %d\n", alertType, count)
		}
	} else {
		fmt.Printf("No alerts generated\n")
	}
	fmt.Println()

	// Performance metrics
	fmt.Println("⚡ Performance Metrics:")
	fmt.Println("======================")
	
	// Calculate total prediction time (mock)
	totalPredictionTime := time.Duration(len(predictions)) * 150 * time.Millisecond
	avgPredictionTime := totalPredictionTime / time.Duration(len(predictions))
	
	fmt.Printf("Total Assets Analyzed: %d\n", len(predictions))
	fmt.Printf("Total Analysis Time: %v\n", totalPredictionTime)
	fmt.Printf("Average Time per Asset: %v\n", avgPredictionTime)
	fmt.Printf("Predictions per Second: %.1f\n", float64(len(predictions))/totalPredictionTime.Seconds())
	fmt.Printf("Cache Hit Rate: %.1f%%\n", 15.0) // Mock cache hit rate
	fmt.Println()

	// Configuration showcase
	fmt.Println("🔧 Configuration Showcase:")
	fmt.Println("==========================")
	
	// Show different configuration profiles
	conservativeConfig := prediction.GetConservativePredictorConfig()
	aggressiveConfig := prediction.GetAggressivePredictorConfig()
	dayTradingConfig := prediction.GetDayTradingPredictorConfig()
	
	fmt.Printf("Available Configuration Profiles:\n")
	fmt.Printf("  📊 Default: Balanced approach with %.0f%% confidence threshold\n", 
		config.ConfidenceThreshold.Mul(decimal.NewFromFloat(100)).InexactFloat64())
	fmt.Printf("  🛡️ Conservative: Higher confidence (%.0f%%) with fundamental focus\n", 
		conservativeConfig.ConfidenceThreshold.Mul(decimal.NewFromFloat(100)).InexactFloat64())
	fmt.Printf("  🚀 Aggressive: Lower confidence (%.0f%%) with sentiment/technical focus\n", 
		aggressiveConfig.ConfidenceThreshold.Mul(decimal.NewFromFloat(100)).InexactFloat64())
	fmt.Printf("  ⚡ Day Trading: High frequency (%v updates) with technical focus\n", 
		dayTradingConfig.UpdateInterval)
	fmt.Println()

	fmt.Println("🎉 AI-Powered Market Prediction example completed!")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("  ✅ Multi-dimensional market analysis")
	fmt.Println("  ✅ Sentiment analysis from multiple sources")
	fmt.Println("  ✅ On-chain metrics and network health analysis")
	fmt.Println("  ✅ Advanced technical indicator analysis")
	fmt.Println("  ✅ Macroeconomic factor integration")
	fmt.Println("  ✅ Ensemble model predictions")
	fmt.Println("  ✅ Confidence scoring and risk assessment")
	fmt.Println("  ✅ Real-time prediction alerts")
	fmt.Println("  ✅ Multiple asset batch predictions")
	fmt.Println("  ✅ Configurable prediction horizons")
	fmt.Println("  ✅ Performance optimization with caching")
	fmt.Println()
	fmt.Println("Note: This example demonstrates the system with mock data.")
	fmt.Println("Configure real data sources and ML models for production use.")

	// Stop the predictor
	if err := predictor.Stop(); err != nil {
		fmt.Printf("Error stopping market predictor: %v\n", err)
	} else {
		fmt.Println("\n🛑 Market predictor stopped")
	}
}

func displayDetailedPrediction(title string, pred *prediction.MarketPrediction) {
	fmt.Printf("📋 %s Detailed Analysis:\n", title)
	fmt.Printf("  Prediction ID: %s\n", pred.ID)
	fmt.Printf("  Current Price: $%s\n", pred.CurrentPrice.String())
	fmt.Printf("  Predicted Price: $%s\n", pred.PredictedPrice.String())
	fmt.Printf("  Price Change: %s%%\n", pred.PriceChangePercent.StringFixed(2))
	fmt.Printf("  Direction: %s %s\n", getDirectionIcon(pred.Direction), pred.Direction)
	fmt.Printf("  Confidence: %s%%\n", pred.Confidence.Mul(decimal.NewFromFloat(100)).StringFixed(1))
	fmt.Printf("  Risk Level: %s\n", pred.RiskLevel)
	fmt.Printf("  Volatility: %s%%\n", pred.Volatility.Mul(decimal.NewFromFloat(100)).StringFixed(1))
	fmt.Printf("  Prediction Horizon: %v\n", pred.PredictionHorizon)
	fmt.Println()

	// Analysis scores
	fmt.Printf("🔍 Analysis Scores:\n")
	fmt.Printf("  Sentiment Score: %s%%\n", pred.SentimentScore.Mul(decimal.NewFromFloat(100)).StringFixed(1))
	fmt.Printf("  On-Chain Score: %s%%\n", pred.OnChainScore.Mul(decimal.NewFromFloat(100)).StringFixed(1))
	fmt.Printf("  Technical Score: %s%%\n", pred.TechnicalScore.Mul(decimal.NewFromFloat(100)).StringFixed(1))
	fmt.Printf("  Macro Score: %s%%\n", pred.MacroScore.Mul(decimal.NewFromFloat(100)).StringFixed(1))
	fmt.Printf("  Ensemble Score: %s%%\n", pred.EnsembleScore.Mul(decimal.NewFromFloat(100)).StringFixed(1))
	fmt.Println()

	// Prediction signals
	if len(pred.Signals) > 0 {
		fmt.Printf("📡 Prediction Signals (%d):\n", len(pred.Signals))
		for _, signal := range pred.Signals {
			fmt.Printf("  %s %s: %s (%.1f%% strength)\n",
				getDirectionIcon(signal.Direction), signal.Type, signal.Description,
				signal.Strength.Mul(decimal.NewFromFloat(100)).InexactFloat64())
		}
		fmt.Println()
	}

	// Prediction factors
	if len(pred.Factors) > 0 {
		fmt.Printf("⚖️ Key Factors (%d):\n", len(pred.Factors))
		for _, factor := range pred.Factors {
			impact := "neutral"
			if factor.Impact.GreaterThan(decimal.NewFromFloat(0.1)) {
				impact = "positive"
			} else if factor.Impact.LessThan(decimal.NewFromFloat(-0.1)) {
				impact = "negative"
			}
			fmt.Printf("  %s: %s impact (weight: %.1f%%)\n",
				factor.Name, impact,
				factor.Weight.Mul(decimal.NewFromFloat(100)).InexactFloat64())
		}
		fmt.Println()
	}

	// Prediction scenarios
	if len(pred.Scenarios) > 0 {
		fmt.Printf("🎭 Scenarios (%d):\n", len(pred.Scenarios))
		for _, scenario := range pred.Scenarios {
			fmt.Printf("  %s: $%s (%.1f%% probability)\n",
				scenario.Name, scenario.PriceTarget.StringFixed(2),
				scenario.Probability.Mul(decimal.NewFromFloat(100)).InexactFloat64())
		}
		fmt.Println()
	}

	// Alerts
	if len(pred.Alerts) > 0 {
		fmt.Printf("🚨 Alerts (%d):\n", len(pred.Alerts))
		for _, alert := range pred.Alerts {
			severityIcon := getSeverityIconForPrediction(alert.Severity)
			fmt.Printf("  %s %s: %s\n", severityIcon, alert.Title, alert.Message)
		}
		fmt.Println()
	}
}

func getDirectionIcon(direction string) string {
	switch direction {
	case "bullish":
		return "🟢"
	case "bearish":
		return "🔴"
	case "neutral":
		return "🟡"
	default:
		return "⚪"
	}
}

func getSeverityIconForPrediction(severity string) string {
	switch severity {
	case "critical":
		return "🚨"
	case "warning":
		return "⚠️"
	case "info":
		return "ℹ️"
	default:
		return "📢"
	}
}
