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

// Helper function to format duration for display
func formatDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	} else {
		return fmt.Sprintf("%.1fd", d.Hours()/24)
	}
}

// Helper function to format percentage
func formatPercentage(d decimal.Decimal) string {
	return fmt.Sprintf("%.1f%%", d.Mul(decimal.NewFromFloat(100)).InexactFloat64())
}

// Helper function to format price
func formatPrice(d decimal.Decimal) string {
	return fmt.Sprintf("$%.2f", d.InexactFloat64())
}

// Helper function to get direction icon
func getDirectionIcon(direction string) string {
	switch direction {
	case "bullish":
		return "📈"
	case "bearish":
		return "📉"
	case "neutral":
		return "➡️"
	default:
		return "❓"
	}
}

// Helper function to get risk level icon
func getRiskLevelIcon(level string) string {
	switch level {
	case "low":
		return "🟢"
	case "medium":
		return "🟡"
	case "high":
		return "🔴"
	default:
		return "⚪"
	}
}

func main() {
	fmt.Println("🤖 Enhanced AI-Powered Market Prediction System")
	fmt.Println("===============================================")

	// Initialize logger
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	logger := logger.NewLogger(logConfig)

	// Create enhanced predictor configuration
	config := prediction.GetDefaultEnhancedPredictorConfig()

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Enabled: %v\n", config.Enabled)
	fmt.Printf("  Update Interval: %v\n", config.UpdateInterval)
	fmt.Printf("  Prediction Horizons: %d\n", len(config.PredictionHorizons))
	fmt.Printf("  Cache Retention: %v\n", config.CacheRetentionPeriod)
	fmt.Printf("  Sentiment Analysis: %v\n", config.SentimentConfig.Enabled)
	fmt.Printf("  On-Chain Analysis: %v\n", config.OnChainConfig.Enabled)
	fmt.Printf("  Technical Analysis: %v\n", config.TechnicalConfig.Enabled)
	fmt.Printf("  Macro Analysis: %v\n", config.MacroConfig.Enabled)
	fmt.Printf("  Machine Learning: %v\n", config.MLConfig.Enabled)
	fmt.Printf("  Ensemble Methods: %v\n", config.EnsembleConfig.Enabled)
	fmt.Println()

	// Create enhanced market predictor
	predictor := prediction.NewEnhancedMarketPredictor(logger, config)

	// Start the predictor
	ctx := context.Background()
	if err := predictor.Start(ctx); err != nil {
		fmt.Printf("Failed to start enhanced predictor: %v\n", err)
		return
	}

	fmt.Println("✅ Enhanced market predictor started successfully!")
	fmt.Println()

	// Show predictor metrics
	fmt.Println("📊 Predictor Metrics:")
	fmt.Println("====================")
	metrics := predictor.GetMetrics()
	fmt.Printf("  Is Running: %v\n", metrics["is_running"])
	fmt.Printf("  Cache Size: %v\n", metrics["cache_size"])
	fmt.Printf("  Sentiment Engine: %v\n", metrics["sentiment_enabled"])
	fmt.Printf("  On-Chain Engine: %v\n", metrics["onchain_enabled"])
	fmt.Printf("  Technical Engine: %v\n", metrics["technical_enabled"])
	fmt.Printf("  Macro Engine: %v\n", metrics["macro_enabled"])
	fmt.Printf("  ML Engine: %v\n", metrics["ml_enabled"])
	fmt.Printf("  Ensemble Engine: %v\n", metrics["ensemble_enabled"])
	fmt.Printf("  Prediction Horizons: %v\n", metrics["prediction_horizons"])
	fmt.Println()

	// Demonstrate enhanced predictions for different assets
	fmt.Println("🔮 Enhanced Market Predictions:")
	fmt.Println("==============================")

	assets := []struct {
		symbol string
		name   string
		price  decimal.Decimal
	}{
		{"BTC", "Bitcoin", decimal.NewFromFloat(50000)},
		{"ETH", "Ethereum", decimal.NewFromFloat(3000)},
		{"SOL", "Solana", decimal.NewFromFloat(100)},
	}

	for i, asset := range assets {
		fmt.Printf("%d. %s (%s) - Current Price: %s\n", i+1, asset.name, asset.symbol, formatPrice(asset.price))
		
		// Generate enhanced prediction
		startTime := time.Now()
		enhancedPrediction, err := predictor.PredictMarketEnhanced(ctx, asset.symbol, asset.price)
		predictionTime := time.Since(startTime)

		if err != nil {
			fmt.Printf("   ❌ Prediction failed: %v\n", err)
			continue
		}

		fmt.Printf("   %s Overall Direction: %s (Confidence: %s)\n", 
			getDirectionIcon(enhancedPrediction.OverallDirection),
			strings.Title(enhancedPrediction.OverallDirection),
			formatPercentage(enhancedPrediction.OverallConfidence))
		
		fmt.Printf("   %s Risk Level: %s\n", 
			getRiskLevelIcon(enhancedPrediction.RiskLevel),
			strings.Title(enhancedPrediction.RiskLevel))
		
		fmt.Printf("   📊 Volatility: %s\n", formatPercentage(enhancedPrediction.Volatility))
		fmt.Printf("   ⚡ Prediction Time: %v\n", predictionTime)
		
		// Show component scores
		fmt.Printf("   🧠 Component Scores:\n")
		if !enhancedPrediction.SentimentScore.IsZero() {
			fmt.Printf("      Sentiment: %s\n", formatPercentage(enhancedPrediction.SentimentScore))
		}
		if !enhancedPrediction.OnChainScore.IsZero() {
			fmt.Printf("      On-Chain: %s\n", formatPercentage(enhancedPrediction.OnChainScore))
		}
		if !enhancedPrediction.TechnicalScore.IsZero() {
			fmt.Printf("      Technical: %s\n", formatPercentage(enhancedPrediction.TechnicalScore))
		}
		if !enhancedPrediction.MacroScore.IsZero() {
			fmt.Printf("      Macro: %s\n", formatPercentage(enhancedPrediction.MacroScore))
		}
		if !enhancedPrediction.MLScore.IsZero() {
			fmt.Printf("      ML Model: %s\n", formatPercentage(enhancedPrediction.MLScore))
		}
		if !enhancedPrediction.EnsembleScore.IsZero() {
			fmt.Printf("      Ensemble: %s\n", formatPercentage(enhancedPrediction.EnsembleScore))
		}
		
		// Show horizon predictions
		fmt.Printf("   ⏰ Time Horizon Predictions:\n")
		for _, horizon := range config.PredictionHorizons {
			if horizonPred, exists := enhancedPrediction.Horizons[horizon]; exists {
				priceChange := horizonPred.PredictedPrice.Sub(asset.price).Div(asset.price)
				fmt.Printf("      %s: %s %s (%s change, %s confidence)\n",
					formatDuration(horizon),
					getDirectionIcon(horizonPred.Direction),
					formatPrice(horizonPred.PredictedPrice),
					formatPercentage(priceChange),
					formatPercentage(horizonPred.Confidence))
			}
		}
		
		// Show feature importance
		if len(enhancedPrediction.FeatureImportance) > 0 {
			fmt.Printf("   🎯 Top Feature Importance:\n")
			count := 0
			for feature, importance := range enhancedPrediction.FeatureImportance {
				if count >= 3 { // Show top 3
					break
				}
				fmt.Printf("      %s: %s\n", strings.Title(strings.Replace(feature, "_", " ", -1)), formatPercentage(importance))
				count++
			}
		}
		
		// Show scenarios
		if len(enhancedPrediction.Scenarios) > 0 {
			fmt.Printf("   📋 Market Scenarios:\n")
			for _, scenario := range enhancedPrediction.Scenarios {
				priceChange := scenario.PriceTarget.Sub(asset.price).Div(asset.price)
				fmt.Printf("      %s: %s (%s probability, %s target)\n",
					scenario.Name,
					formatPercentage(priceChange),
					formatPercentage(scenario.Probability),
					formatDuration(scenario.TimeToTarget))
			}
		}
		
		// Show alerts
		if len(enhancedPrediction.Alerts) > 0 {
			fmt.Printf("   🚨 Alerts:\n")
			for _, alert := range enhancedPrediction.Alerts {
				severityIcon := "ℹ️"
				if alert.Severity == "warning" {
					severityIcon = "⚠️"
				} else if alert.Severity == "error" {
					severityIcon = "🚨"
				}
				fmt.Printf("      %s %s: %s\n", severityIcon, alert.Title, alert.Message)
				if alert.ActionRecommendation != "" {
					fmt.Printf("         💡 Recommendation: %s\n", alert.ActionRecommendation)
				}
			}
		}
		
		fmt.Println()
	}

	// Show different configuration profiles
	fmt.Println("🔧 Configuration Profiles:")
	fmt.Println("==========================")
	
	// High frequency config
	hfConfig := prediction.GetHighFrequencyConfig()
	fmt.Printf("⚡ High Frequency Trading:\n")
	fmt.Printf("  Update Interval: %v (vs %v default)\n", hfConfig.UpdateInterval, config.UpdateInterval)
	fmt.Printf("  Cache Retention: %v (vs %v default)\n", hfConfig.CacheRetentionPeriod, config.CacheRetentionPeriod)
	fmt.Printf("  Prediction Horizons: %d (vs %d default)\n", len(hfConfig.PredictionHorizons), len(config.PredictionHorizons))
	fmt.Printf("  Sentiment Update: %v (vs %v default)\n", hfConfig.SentimentConfig.UpdateFrequency, config.SentimentConfig.UpdateFrequency)
	fmt.Println()

	// Long term config
	ltConfig := prediction.GetLongTermConfig()
	fmt.Printf("📈 Long-Term Analysis:\n")
	fmt.Printf("  Update Interval: %v (extended for stability)\n", ltConfig.UpdateInterval)
	fmt.Printf("  Prediction Horizons: %d (extended timeframes)\n", len(ltConfig.PredictionHorizons))
	fmt.Printf("  Data Retention: %v (comprehensive history)\n", ltConfig.DataAggregatorConfig.DataRetention)
	fmt.Printf("  ML Training: %v (less frequent)\n", ltConfig.MLConfig.TrainingFrequency)
	fmt.Println()

	// Research config
	researchConfig := prediction.GetResearchConfig()
	fmt.Printf("🔬 Research & Analysis:\n")
	fmt.Printf("  Prediction Horizons: %d (comprehensive coverage)\n", len(researchConfig.PredictionHorizons))
	fmt.Printf("  Data Retention: %v (maximum history)\n", researchConfig.DataAggregatorConfig.DataRetention)
	fmt.Printf("  Technical Timeframes: %d (all available)\n", len(researchConfig.TechnicalConfig.Timeframes))
	fmt.Printf("  ML Models: %d (all available)\n", len(researchConfig.MLConfig.Models))
	fmt.Println()

	// Show supported features
	fmt.Println("🛠️  Supported Features:")
	fmt.Println("======================")
	
	fmt.Println("Sentiment Sources:")
	sources := prediction.GetSupportedSentimentSources()
	for i, source := range sources {
		if i > 0 && i%5 == 0 {
			fmt.Println()
		}
		fmt.Printf("  %-12s", source)
	}
	fmt.Println("\n")

	fmt.Println("On-Chain Metrics:")
	metrics_list := prediction.GetSupportedOnChainMetrics()
	for i, metric := range metrics_list {
		if i > 0 && i%3 == 0 {
			fmt.Println()
		}
		fmt.Printf("  %-20s", strings.Replace(metric, "_", " ", -1))
	}
	fmt.Println("\n")

	fmt.Println("Technical Indicators:")
	indicators := prediction.GetSupportedTechnicalIndicators()
	for i, indicator := range indicators {
		if i > 0 && i%4 == 0 {
			fmt.Println()
		}
		fmt.Printf("  %-15s", strings.ToUpper(indicator))
	}
	fmt.Println("\n")

	fmt.Println("ML Models:")
	models := prediction.GetSupportedMLModels()
	for i, model := range models {
		if i > 0 && i%4 == 0 {
			fmt.Println()
		}
		fmt.Printf("  %-15s", strings.ToUpper(model))
	}
	fmt.Println("\n")

	// Performance demonstration
	fmt.Println("⚡ Performance Demonstration:")
	fmt.Println("============================")
	
	// Measure prediction performance
	testAsset := "BTC"
	testPrice := decimal.NewFromFloat(50000)
	iterations := 5
	
	startTime := time.Now()
	for i := 0; i < iterations; i++ {
		_, err := predictor.PredictMarketEnhanced(ctx, testAsset, testPrice)
		if err != nil {
			fmt.Printf("Prediction %d failed: %v\n", i+1, err)
		}
	}
	totalTime := time.Since(startTime)
	avgTime := totalTime / time.Duration(iterations)
	
	fmt.Printf("Predictions per Second: %.1f\n", float64(iterations)/totalTime.Seconds())
	fmt.Printf("Average Prediction Time: %v\n", avgTime)
	fmt.Printf("Cache Efficiency: High (subsequent requests use cached results)\n")
	fmt.Printf("Multi-Engine Processing: Parallel execution of all analysis engines\n")
	fmt.Println()

	// Best practices
	fmt.Println("💡 Best Practices:")
	fmt.Println("==================")
	fmt.Println("1. Use appropriate prediction horizons for your trading strategy")
	fmt.Println("2. Consider ensemble confidence levels when making decisions")
	fmt.Println("3. Monitor feature importance to understand prediction drivers")
	fmt.Println("4. Review scenario analysis for risk management")
	fmt.Println("5. Pay attention to alerts for significant market changes")
	fmt.Println("6. Combine multiple prediction engines for better accuracy")
	fmt.Println("7. Regularly retrain ML models with fresh data")
	fmt.Println("8. Validate predictions against actual market outcomes")
	fmt.Println()

	fmt.Println("🎉 Enhanced AI Prediction System example completed!")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("  ✅ Multi-engine prediction system (Sentiment, On-Chain, Technical, Macro, ML)")
	fmt.Println("  ✅ Advanced ensemble methods for improved accuracy")
	fmt.Println("  ✅ Multiple time horizon predictions")
	fmt.Println("  ✅ Feature importance analysis")
	fmt.Println("  ✅ Scenario-based risk analysis")
	fmt.Println("  ✅ Real-time alerts and recommendations")
	fmt.Println("  ✅ Configurable prediction profiles")
	fmt.Println("  ✅ Performance optimization with caching")
	fmt.Println()
	fmt.Println("Note: This example demonstrates the system with mock data and models.")
	fmt.Println("Integrate with real data sources and trained models for production use.")

	// Stop the predictor
	if err := predictor.Stop(); err != nil {
		fmt.Printf("Error stopping enhanced predictor: %v\n", err)
	} else {
		fmt.Println("\n🛑 Enhanced market predictor stopped")
	}
}
