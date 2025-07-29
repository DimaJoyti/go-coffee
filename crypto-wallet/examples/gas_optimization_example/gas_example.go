package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/blockchain/gas"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/shopspring/decimal"
)

// Helper function to create sample optimization requests
func createSampleRequests() []*gas.OptimizationRequest {
	return []*gas.OptimizationRequest{
		{
			TransactionType:   "transfer",
			Priority:          "low",
			TargetConfirmTime: 10 * time.Minute,
			MaxGasPrice:       decimal.NewFromFloat(50),
			GasLimit:          21000,
			Value:             decimal.NewFromFloat(0.1),
			IsReplacement:     false,
			UserPreferences: gas.UserPreferences{
				CostOptimization:  true,
				SpeedOptimization: false,
				MaxCostTolerance:  decimal.NewFromFloat(0.005),
				RiskTolerance:     "low",
			},
		},
		{
			TransactionType:   "contract_call",
			Priority:          "medium",
			TargetConfirmTime: 5 * time.Minute,
			MaxGasPrice:       decimal.NewFromFloat(100),
			GasLimit:          150000,
			Value:             decimal.NewFromFloat(1.0),
			IsReplacement:     false,
			UserPreferences: gas.UserPreferences{
				CostOptimization:  true,
				SpeedOptimization: true,
				MaxCostTolerance:  decimal.NewFromFloat(0.02),
				RiskTolerance:     "medium",
			},
		},
		{
			TransactionType:   "defi_swap",
			Priority:          "high",
			TargetConfirmTime: 2 * time.Minute,
			MaxGasPrice:       decimal.NewFromFloat(200),
			GasLimit:          300000,
			Value:             decimal.NewFromFloat(5.0),
			IsReplacement:     false,
			UserPreferences: gas.UserPreferences{
				CostOptimization:  false,
				SpeedOptimization: true,
				MaxCostTolerance:  decimal.NewFromFloat(0.1),
				RiskTolerance:     "high",
			},
		},
		{
			TransactionType:   "nft_mint",
			Priority:          "urgent",
			TargetConfirmTime: 30 * time.Second,
			MaxGasPrice:       decimal.NewFromFloat(500),
			GasLimit:          200000,
			Value:             decimal.NewFromFloat(0.5),
			IsReplacement:     true,
			CurrentGasPrice:   decimal.NewFromFloat(80),
			UserPreferences: gas.UserPreferences{
				CostOptimization:  false,
				SpeedOptimization: true,
				MaxCostTolerance:  decimal.NewFromFloat(0.2),
				RiskTolerance:     "high",
			},
		},
	}
}

func main() {
	fmt.Println("⛽ Gas Optimization System Example")
	fmt.Println("==================================")

	// Initialize logger
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	logger := logger.NewLogger(logConfig)

	// Create gas optimizer configuration
	config := gas.GetDefaultGasOptimizerConfig()

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Enabled: %v\n", config.Enabled)
	fmt.Printf("  Update Interval: %v\n", config.UpdateInterval)
	fmt.Printf("  History Retention: %v\n", config.HistoryRetentionPeriod)
	fmt.Printf("  Max History Size: %d\n", config.MaxHistorySize)
	fmt.Printf("  Optimization Strategies: %v\n", config.OptimizationStrategies)
	fmt.Printf("  EIP-1559 Enabled: %v\n", config.EIP1559Config.Enabled)
	fmt.Printf("  Historical Analysis: %v\n", config.HistoricalConfig.Enabled)
	fmt.Printf("  Congestion Monitoring: %v\n", config.CongestionConfig.Enabled)
	fmt.Printf("  Prediction Engine: %v\n", config.PredictionConfig.Enabled)
	fmt.Println()

	// Create gas optimizer
	optimizer := gas.NewGasOptimizer(logger, config)

	// Start the optimizer
	ctx := context.Background()
	if err := optimizer.Start(ctx); err != nil {
		fmt.Printf("Failed to start gas optimizer: %v\n", err)
		return
	}

	fmt.Println("✅ Gas optimizer started successfully!")
	fmt.Println()

	// Wait for initial network metrics
	time.Sleep(1 * time.Second)

	// Show network metrics
	fmt.Println("📊 Current Network Metrics:")
	fmt.Println("===========================")
	metrics := optimizer.GetNetworkMetrics()
	fmt.Printf("  Base Fee: %s gwei\n", metrics.CurrentBaseFee.String())
	fmt.Printf("  Recommended Priority: %s gwei\n", metrics.RecommendedPriority.String())
	fmt.Printf("  Network Utilization: %.1f%%\n", metrics.NetworkUtilization.Mul(decimal.NewFromFloat(100)).InexactFloat64())
	fmt.Printf("  Congestion Level: %s %s\n", getCongestionIcon(metrics.CongestionLevel), metrics.CongestionLevel)
	fmt.Printf("  Average Confirmation Time: %v\n", metrics.AverageConfirmTime)
	fmt.Printf("  Pending Transactions: %d\n", metrics.PendingTransactions)
	fmt.Printf("  Last Updated: %v\n", metrics.LastUpdated.Format("15:04:05"))
	fmt.Println()

	// Show optimizer metrics
	fmt.Println("🔧 Optimizer Metrics:")
	fmt.Println("=====================")
	optimizerMetrics := optimizer.GetMetrics()
	fmt.Printf("  Is Running: %v\n", optimizerMetrics["is_running"])
	fmt.Printf("  Gas History Size: %v\n", optimizerMetrics["gas_history_size"])
	fmt.Printf("  Cache Size: %v\n", optimizerMetrics["cache_size"])
	fmt.Printf("  EIP-1559 Enabled: %v\n", optimizerMetrics["eip1559_enabled"])
	fmt.Printf("  Historical Enabled: %v\n", optimizerMetrics["historical_enabled"])
	fmt.Printf("  Congestion Enabled: %v\n", optimizerMetrics["congestion_enabled"])
	fmt.Printf("  Prediction Enabled: %v\n", optimizerMetrics["prediction_enabled"])
	fmt.Println()

	// Demonstrate gas optimization for different scenarios
	fmt.Println("🎯 Gas Optimization Examples:")
	fmt.Println("=============================")

	requests := createSampleRequests()
	for i, request := range requests {
		fmt.Printf("%d. %s Transaction (%s priority)\n", i+1, 
			strings.Title(strings.Replace(request.TransactionType, "_", " ", -1)), 
			request.Priority)
		
		fmt.Printf("   Target Time: %v, Max Gas: %s gwei, Gas Limit: %d\n",
			request.TargetConfirmTime,
			request.MaxGasPrice.String(),
			request.GasLimit)

		// Optimize gas price
		startTime := time.Now()
		result, err := optimizer.OptimizeGasPrice(ctx, request)
		optimizationTime := time.Since(startTime)

		if err != nil {
			fmt.Printf("   ❌ Optimization failed: %v\n", err)
			continue
		}

		fmt.Printf("   ✅ Strategy: %s\n", result.Strategy)
		fmt.Printf("   💰 Gas Price: %s gwei\n", result.GasPrice.String())
		
		if !result.MaxFeePerGas.IsZero() {
			fmt.Printf("   🔝 Max Fee: %s gwei\n", result.MaxFeePerGas.String())
		}
		if !result.MaxPriorityFeePerGas.IsZero() {
			fmt.Printf("   ⚡ Priority Fee: %s gwei\n", result.MaxPriorityFeePerGas.String())
		}
		
		fmt.Printf("   💸 Estimated Cost: %s ETH\n", result.EstimatedCost.String())
		fmt.Printf("   ⏱️  Estimated Time: %v\n", result.EstimatedTime)
		fmt.Printf("   🎯 Confidence: %.1f%%\n", result.Confidence.Mul(decimal.NewFromFloat(100)).InexactFloat64())
		fmt.Printf("   ⚡ Optimization Time: %v\n", optimizationTime)
		
		if len(result.Reasoning) > 0 {
			fmt.Printf("   📝 Reasoning:\n")
			for _, reason := range result.Reasoning {
				fmt.Printf("      • %s\n", reason)
			}
		}
		
		if len(result.Alternatives) > 0 {
			fmt.Printf("   🔄 Alternatives:\n")
			for j, alt := range result.Alternatives[:min(3, len(result.Alternatives))] {
				fmt.Printf("      %d. %s: %s gwei (%.1f%% confidence)\n", 
					j+1, alt.Name, alt.GasPrice.String(), 
					alt.Confidence.Mul(decimal.NewFromFloat(100)).InexactFloat64())
			}
		}
		fmt.Println()
	}

	// Show different configuration profiles
	fmt.Println("🔧 Configuration Profiles:")
	fmt.Println("==========================")
	
	// High frequency trading config
	hfConfig := gas.GetHighFrequencyConfig()
	fmt.Printf("📈 High Frequency Trading:\n")
	fmt.Printf("  Update Interval: %v (vs %v default)\n", hfConfig.UpdateInterval, config.UpdateInterval)
	fmt.Printf("  Base Fee Multiplier: %s (vs %s default)\n", 
		hfConfig.EIP1559Config.BaseFeeMultiplier.String(),
		config.EIP1559Config.BaseFeeMultiplier.String())
	fmt.Printf("  Aggressiveness: %s (vs %s default)\n", 
		hfConfig.EIP1559Config.AggressivenessLevel,
		config.EIP1559Config.AggressivenessLevel)
	fmt.Println()

	// Cost optimized config
	costConfig := gas.GetCostOptimizedConfig()
	fmt.Printf("💰 Cost Optimized:\n")
	fmt.Printf("  Update Interval: %v (longer for cost efficiency)\n", costConfig.UpdateInterval)
	fmt.Printf("  Base Fee Multiplier: %s (conservative)\n", costConfig.EIP1559Config.BaseFeeMultiplier.String())
	fmt.Printf("  Safety Multiplier: %s (vs %s default)\n", 
		costConfig.SafetyMargins.SafetyMultiplier.String(),
		config.SafetyMargins.SafetyMultiplier.String())
	fmt.Println()

	// Balanced config
	balancedConfig := gas.GetBalancedConfig()
	fmt.Printf("⚖️  Balanced:\n")
	fmt.Printf("  Base Fee Multiplier: %s\n", balancedConfig.EIP1559Config.BaseFeeMultiplier.String())
	fmt.Printf("  Target Confirmation: %v\n", balancedConfig.EIP1559Config.TargetConfirmationTime)
	fmt.Printf("  Analysis Window: %v\n", balancedConfig.HistoricalConfig.AnalysisWindow)
	fmt.Println()

	// Show supported features
	fmt.Println("🛠️  Supported Features:")
	fmt.Println("======================")
	
	fmt.Println("Optimization Strategies:")
	strategies := gas.GetSupportedOptimizationStrategies()
	descriptions := gas.GetStrategyDescription()
	for _, strategy := range strategies {
		if desc, exists := descriptions[strategy]; exists {
			fmt.Printf("  • %s: %s\n", strategy, desc)
		} else {
			fmt.Printf("  • %s\n", strategy)
		}
	}
	fmt.Println()

	fmt.Println("Priority Fee Strategies:")
	priorityStrategies := gas.GetSupportedPriorityFeeStrategies()
	for _, strategy := range priorityStrategies {
		fmt.Printf("  • %s\n", strategy)
	}
	fmt.Println()

	fmt.Println("Aggressiveness Levels:")
	levels := gas.GetSupportedAggressivenessLevels()
	for _, level := range levels {
		fmt.Printf("  • %s\n", level)
	}
	fmt.Println()

	// Show gas history
	fmt.Println("📈 Recent Gas History:")
	fmt.Println("=====================")
	history := optimizer.GetGasHistory(5)
	if len(history) > 0 {
		fmt.Printf("  %-12s %-10s %-12s %-15s %-10s\n", "Time", "Base Fee", "Priority", "Total", "Util%")
		fmt.Printf("  %-12s %-10s %-12s %-15s %-10s\n", "----", "--------", "--------", "-----", "-----")
		for _, entry := range history {
			fmt.Printf("  %-12s %-10s %-12s %-15s %-10.1f\n",
				entry.Timestamp.Format("15:04:05"),
				entry.BaseFee.StringFixed(1),
				entry.PriorityFee.StringFixed(1),
				entry.GasPrice.StringFixed(1),
				entry.BlockUtilization.Mul(decimal.NewFromFloat(100)).InexactFloat64())
		}
	} else {
		fmt.Println("  No history available yet")
	}
	fmt.Println()

	// Performance demonstration
	fmt.Println("⚡ Performance Demonstration:")
	fmt.Println("============================")
	
	// Measure optimization performance
	testRequest := requests[1] // Use medium priority request
	iterations := 10
	
	startTime := time.Now()
	for i := 0; i < iterations; i++ {
		_, err := optimizer.OptimizeGasPrice(ctx, testRequest)
		if err != nil {
			fmt.Printf("Optimization %d failed: %v\n", i+1, err)
		}
	}
	totalTime := time.Since(startTime)
	avgTime := totalTime / time.Duration(iterations)
	
	fmt.Printf("Optimizations per Second: %.1f\n", float64(iterations)/totalTime.Seconds())
	fmt.Printf("Average Optimization Time: %v\n", avgTime)
	fmt.Printf("Cache Hit Rate: High (subsequent requests use cached results)\n")
	fmt.Println()

	// Best practices
	fmt.Println("💡 Best Practices:")
	fmt.Println("==================")
	fmt.Println("1. Use appropriate priority levels based on transaction urgency")
	fmt.Println("2. Set realistic target confirmation times")
	fmt.Println("3. Configure max gas price limits to prevent overpaying")
	fmt.Println("4. Monitor network congestion and adjust strategies accordingly")
	fmt.Println("5. Use cost optimization for non-urgent transactions")
	fmt.Println("6. Enable multiple optimization strategies for better results")
	fmt.Println("7. Regularly review and update safety margins")
	fmt.Println("8. Consider user preferences in optimization decisions")
	fmt.Println()

	fmt.Println("🎉 Gas Optimization System example completed!")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("  ✅ Multi-strategy gas price optimization")
	fmt.Println("  ✅ EIP-1559 transaction support")
	fmt.Println("  ✅ Historical data analysis")
	fmt.Println("  ✅ Network congestion monitoring")
	fmt.Println("  ✅ Prediction-based optimization")
	fmt.Println("  ✅ Configurable safety margins")
	fmt.Println("  ✅ Performance optimization with caching")
	fmt.Println("  ✅ Multiple configuration profiles")
	fmt.Println()
	fmt.Println("Note: This example demonstrates the system with mock data.")
	fmt.Println("Integrate with real blockchain data sources for production use.")

	// Stop the optimizer
	if err := optimizer.Stop(); err != nil {
		fmt.Printf("Error stopping gas optimizer: %v\n", err)
	} else {
		fmt.Println("\n🛑 Gas optimizer stopped")
	}
}

func getCongestionIcon(level string) string {
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
