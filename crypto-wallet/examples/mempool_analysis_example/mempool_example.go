package main

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/blockchain/mempool"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
)

// Helper function to create test transactions
func createTestTransaction(gasPrice *big.Int, gasLimit uint64, nonce uint64, value *big.Int) *types.Transaction {
	to := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1")
	data := []byte{}
	return types.NewTransaction(nonce, to, value, gasLimit, gasPrice, data)
}

// Helper function to create EIP-1559 transaction
func createEIP1559Transaction(gasFeeCap, gasTipCap *big.Int, gasLimit uint64, nonce uint64, value *big.Int) *types.Transaction {
	to := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1")
	data := []byte{}

	return types.NewTx(&types.DynamicFeeTx{
		ChainID:   big.NewInt(1),
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &to,
		Value:     value,
		Data:      data,
	})
}

// Helper function to create sample transactions
func createSampleTransactions() []*types.Transaction {
	var transactions []*types.Transaction
	
	// Create a mix of legacy and EIP-1559 transactions with varying gas prices
	gasPrices := []*big.Int{
		big.NewInt(10000000000),  // 10 gwei
		big.NewInt(15000000000),  // 15 gwei
		big.NewInt(20000000000),  // 20 gwei
		big.NewInt(25000000000),  // 25 gwei
		big.NewInt(30000000000),  // 30 gwei
		big.NewInt(40000000000),  // 40 gwei
		big.NewInt(50000000000),  // 50 gwei
		big.NewInt(75000000000),  // 75 gwei
		big.NewInt(100000000000), // 100 gwei
		big.NewInt(150000000000), // 150 gwei
	}
	
	values := []*big.Int{
		big.NewInt(100000000000000000),  // 0.1 ETH
		big.NewInt(500000000000000000),  // 0.5 ETH
		big.NewInt(1000000000000000000), // 1 ETH
		big.NewInt(2000000000000000000), // 2 ETH
		big.NewInt(5000000000000000000), // 5 ETH
	}
	
	// Create legacy transactions
	for i, gasPrice := range gasPrices[:5] {
		tx := createTestTransaction(gasPrice, 21000, uint64(i+1), values[i%len(values)])
		transactions = append(transactions, tx)
	}
	
	// Create EIP-1559 transactions
	for i, gasPrice := range gasPrices[5:] {
		tipCap := big.NewInt(2000000000) // 2 gwei tip
		tx := createEIP1559Transaction(gasPrice, tipCap, 21000, uint64(i+6), values[i%len(values)])
		transactions = append(transactions, tx)
	}
	
	return transactions
}

func main() {
	fmt.Println("⛽ Mempool Analysis System Example")
	fmt.Println("=================================")

	// Initialize logger
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	logger := logger.NewLogger(logConfig)

	// Create mempool analyzer configuration
	config := mempool.GetDefaultMempoolAnalyzerConfig()

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Enabled: %v\n", config.Enabled)
	fmt.Printf("  Update Interval: %v\n", config.UpdateInterval)
	fmt.Printf("  Data Retention Period: %v\n", config.DataRetentionPeriod)
	fmt.Printf("  Max Transactions: %d\n", config.MaxTransactions)
	fmt.Printf("  Gas Tracker: %v\n", config.GasTrackerConfig.Enabled)
	fmt.Printf("  Congestion Model: %v\n", config.CongestionModelConfig.Enabled)
	fmt.Printf("  Gas Predictor: %v\n", config.GasPredictorConfig.Enabled)
	fmt.Printf("  Time Estimator: %v\n", config.TimeEstimatorConfig.Enabled)
	fmt.Printf("  Priority Analyzer: %v\n", config.PriorityAnalyzerConfig.Enabled)
	fmt.Println()

	// Create mempool analyzer
	analyzer := mempool.NewMempoolAnalyzer(logger, config)

	// Start the analyzer
	ctx := context.Background()
	if err := analyzer.Start(ctx); err != nil {
		fmt.Printf("Failed to start mempool analyzer: %v\n", err)
		return
	}

	fmt.Println("✅ Mempool analyzer started successfully!")
	fmt.Println()

	// Show analyzer metrics
	fmt.Println("📊 Analyzer Metrics:")
	fmt.Println("===================")
	metrics := analyzer.GetMetrics()
	fmt.Printf("  Total Transactions: %v\n", metrics["total_transactions"])
	fmt.Printf("  Is Running: %v\n", metrics["is_running"])
	fmt.Printf("  Gas Tracker Enabled: %v\n", metrics["gas_tracker_enabled"])
	fmt.Printf("  Congestion Model Enabled: %v\n", metrics["congestion_model_enabled"])
	fmt.Printf("  Gas Predictor Enabled: %v\n", metrics["gas_predictor_enabled"])
	fmt.Printf("  Time Estimator Enabled: %v\n", metrics["time_estimator_enabled"])
	fmt.Printf("  Priority Analyzer Enabled: %v\n", metrics["priority_analyzer_enabled"])
	fmt.Println()

	// Create and add sample transactions
	fmt.Println("📝 Adding Sample Transactions:")
	fmt.Println("=============================")
	
	transactions := createSampleTransactions()
	fmt.Printf("Created %d sample transactions\n", len(transactions))
	
	for i, tx := range transactions {
		err := analyzer.AddTransaction(tx)
		if err != nil {
			fmt.Printf("Failed to add transaction %d: %v\n", i+1, err)
			continue
		}
		
		// Display transaction info
		txType := "Legacy"
		gasInfo := fmt.Sprintf("Gas Price: %s gwei", new(big.Int).Div(tx.GasPrice(), big.NewInt(1000000000)).String())
		
		if tx.Type() == types.DynamicFeeTxType {
			txType = "EIP-1559"
			gasInfo = fmt.Sprintf("Fee Cap: %s gwei, Tip Cap: %s gwei", 
				new(big.Int).Div(tx.GasFeeCap(), big.NewInt(1000000000)).String(),
				new(big.Int).Div(tx.GasTipCap(), big.NewInt(1000000000)).String())
		}
		
		fmt.Printf("  %d. %s - %s, Value: %s ETH\n", 
			i+1, txType, gasInfo, 
			new(big.Int).Div(tx.Value(), big.NewInt(1000000000000000000)).String())
	}
	fmt.Println()

	// Wait a moment for processing
	time.Sleep(1 * time.Second)

	// Perform mempool analysis
	fmt.Println("🔍 Performing Mempool Analysis:")
	fmt.Println("==============================")
	
	analysis, err := analyzer.AnalyzeMempool(ctx)
	if err != nil {
		fmt.Printf("Failed to analyze mempool: %v\n", err)
		return
	}

	fmt.Println("✅ Analysis completed!")
	fmt.Println()

	// Display analysis results
	displayAnalysisResults(analysis)

	// Show different configuration profiles
	fmt.Println("🔧 Configuration Profiles:")
	fmt.Println("==========================")
	
	// High frequency trading config
	hfConfig := mempool.GetHighFrequencyConfig()
	fmt.Printf("📈 High Frequency Trading:\n")
	fmt.Printf("  Update Interval: %v (vs %v default)\n", hfConfig.UpdateInterval, config.UpdateInterval)
	fmt.Printf("  Max Transactions: %d (vs %d default)\n", hfConfig.MaxTransactions, config.MaxTransactions)
	fmt.Printf("  Gas Tracking Window: %v\n", hfConfig.GasTrackerConfig.TrackingWindow)
	fmt.Println()

	// Low latency config
	llConfig := mempool.GetLowLatencyConfig()
	fmt.Printf("⚡ Low Latency:\n")
	fmt.Printf("  Update Interval: %v (optimized for speed)\n", llConfig.UpdateInterval)
	fmt.Printf("  Max Transactions: %d (reduced for performance)\n", llConfig.MaxTransactions)
	fmt.Printf("  Sample Size: %d (minimal for speed)\n", llConfig.GasTrackerConfig.SampleSize)
	fmt.Println()

	// Analytics config
	analyticsConfig := mempool.GetAnalyticsConfig()
	fmt.Printf("📊 Analytics:\n")
	fmt.Printf("  Data Retention: %v (extended for analysis)\n", analyticsConfig.DataRetentionPeriod)
	fmt.Printf("  Max Transactions: %d (comprehensive data)\n", analyticsConfig.MaxTransactions)
	fmt.Printf("  Prediction Methods: %d (all available)\n", len(analyticsConfig.GasPredictorConfig.PredictionMethods))
	fmt.Println()

	// Show supported features
	fmt.Println("🛠️  Supported Features:")
	fmt.Println("======================")
	
	fmt.Println("Prediction Methods:")
	methods := mempool.GetSupportedPredictionMethods()
	for i, method := range methods {
		if i > 0 && i%3 == 0 {
			fmt.Println()
		}
		fmt.Printf("  %-20s", method)
	}
	fmt.Println("\n")

	fmt.Println("Priority Factors:")
	factors := mempool.GetSupportedPriorityFactors()
	for i, factor := range factors {
		if i > 0 && i%4 == 0 {
			fmt.Println()
		}
		fmt.Printf("  %-18s", factor)
	}
	fmt.Println("\n")

	fmt.Println("Estimation Methods:")
	estimationMethods := mempool.GetSupportedEstimationMethods()
	for _, method := range estimationMethods {
		fmt.Printf("  • %s\n", method)
	}
	fmt.Println()

	// Show congestion level descriptions
	fmt.Println("📊 Congestion Levels:")
	fmt.Println("====================")
	descriptions := mempool.GetCongestionLevelDescription()
	for level, desc := range descriptions {
		icon := getCongestionIcon(level)
		fmt.Printf("  %s %s: %s\n", icon, strings.Title(level), desc)
	}
	fmt.Println()

	// Performance demonstration
	fmt.Println("⚡ Performance Demonstration:")
	fmt.Println("============================")
	
	// Measure analysis time
	startTime := time.Now()
	for i := 0; i < 5; i++ {
		_, err := analyzer.AnalyzeMempool(ctx)
		if err != nil {
			fmt.Printf("Analysis %d failed: %v\n", i+1, err)
		}
	}
	avgTime := time.Since(startTime) / 5
	
	fmt.Printf("Average Analysis Time: %v\n", avgTime)
	fmt.Printf("Analyses per Second: %.1f\n", float64(time.Second)/float64(avgTime))
	fmt.Printf("Memory Efficiency: Tracking %d transactions\n", len(transactions))
	fmt.Println()

	// Best practices
	fmt.Println("💡 Best Practices:")
	fmt.Println("==================")
	fmt.Println("1. Monitor gas price trends for optimal transaction timing")
	fmt.Println("2. Use congestion analysis to adjust gas prices dynamically")
	fmt.Println("3. Implement priority-based transaction queuing")
	fmt.Println("4. Set appropriate gas price based on urgency and cost tolerance")
	fmt.Println("5. Monitor mempool for transaction replacement opportunities")
	fmt.Println("6. Use prediction models to anticipate gas price movements")
	fmt.Println("7. Implement automatic gas price adjustment based on network conditions")
	fmt.Println("8. Consider transaction batching during high congestion periods")
	fmt.Println()

	fmt.Println("🎉 Mempool Analysis System example completed!")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("  ✅ Real-time mempool transaction tracking")
	fmt.Println("  ✅ Comprehensive gas price analysis and statistics")
	fmt.Println("  ✅ Network congestion modeling and prediction")
	fmt.Println("  ✅ Transaction confirmation time estimation")
	fmt.Println("  ✅ Priority-based transaction analysis")
	fmt.Println("  ✅ Gas price prediction with multiple models")
	fmt.Println("  ✅ Configurable analysis profiles for different use cases")
	fmt.Println("  ✅ Performance optimization for high-frequency scenarios")
	fmt.Println()
	fmt.Println("Note: This example demonstrates the system with mock data.")
	fmt.Println("Integrate with real mempool data sources for production use.")

	// Stop the analyzer
	if err := analyzer.Stop(); err != nil {
		fmt.Printf("Error stopping mempool analyzer: %v\n", err)
	} else {
		fmt.Println("\n🛑 Mempool analyzer stopped")
	}
}

func displayAnalysisResults(analysis *mempool.MempoolAnalysis) {
	fmt.Printf("📋 Analysis Results:\n")
	fmt.Printf("  Timestamp: %v\n", analysis.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Total Transactions: %d\n", analysis.TotalTransactions)
	fmt.Printf("  Pending Transactions: %d\n", analysis.PendingTransactions)
	fmt.Printf("  Congestion Level: %s %s\n", getCongestionIcon(analysis.CongestionLevel), analysis.CongestionLevel)
	fmt.Printf("  Congestion Score: %.2f\n", analysis.CongestionScore.InexactFloat64())
	fmt.Printf("  Optimal Gas Price: %s gwei\n", analysis.OptimalGasPrice.Div(decimal.NewFromInt(1000000000)).StringFixed(1))
	fmt.Printf("  Estimated Wait Time: %v\n", analysis.EstimatedWaitTime)
	fmt.Println()

	// Gas statistics
	if analysis.GasStatistics != nil {
		fmt.Printf("⛽ Gas Statistics:\n")
		fmt.Printf("  Mean: %s gwei\n", analysis.GasStatistics.Mean.Div(decimal.NewFromInt(1000000000)).StringFixed(1))
		fmt.Printf("  Median: %s gwei\n", analysis.GasStatistics.Median.Div(decimal.NewFromInt(1000000000)).StringFixed(1))
		fmt.Printf("  Standard Deviation: %s gwei\n", analysis.GasStatistics.StandardDev.Div(decimal.NewFromInt(1000000000)).StringFixed(1))
		fmt.Printf("  Trend: %s\n", getTrendIcon(analysis.GasStatistics.Trend))
		fmt.Printf("  Volatility: %.1f%%\n", analysis.GasStatistics.Volatility.Mul(decimal.NewFromFloat(100)).InexactFloat64())
		
		if len(analysis.GasStatistics.Percentiles) > 0 {
			fmt.Printf("  Percentiles:\n")
			percentiles := []int{25, 50, 75, 90, 95}
			for _, p := range percentiles {
				if value, exists := analysis.GasStatistics.Percentiles[p]; exists {
					fmt.Printf("    %d%%: %s gwei\n", p, value.Div(decimal.NewFromInt(1000000000)).StringFixed(1))
				}
			}
		}
		fmt.Println()
	}

	// Gas predictions
	if len(analysis.GasPredictions) > 0 {
		fmt.Printf("🔮 Gas Predictions:\n")
		for horizon, prediction := range analysis.GasPredictions {
			fmt.Printf("  %s (%v):\n", strings.Title(strings.Replace(horizon, "_", " ", -1)), prediction.TimeHorizon)
			fmt.Printf("    Predicted Price: %s gwei\n", prediction.PredictedPrice.Div(decimal.NewFromInt(1000000000)).StringFixed(1))
			fmt.Printf("    Confidence: %.1f%%\n", prediction.Confidence.Mul(decimal.NewFromFloat(100)).InexactFloat64())
			fmt.Printf("    Range: %s - %s gwei\n", 
				prediction.Range.Low.Div(decimal.NewFromInt(1000000000)).StringFixed(1),
				prediction.Range.High.Div(decimal.NewFromInt(1000000000)).StringFixed(1))
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

	// Top transactions
	if len(analysis.TopTransactions) > 0 {
		fmt.Printf("🏆 Top Priority Transactions:\n")
		for i, tx := range analysis.TopTransactions[:min(5, len(analysis.TopTransactions))] {
			gasPrice := "N/A"
			if !tx.GasPrice.IsZero() {
				gasPrice = tx.GasPrice.Div(decimal.NewFromInt(1000000000)).StringFixed(1) + " gwei"
			} else if !tx.GasFeeCap.IsZero() {
				gasPrice = tx.GasFeeCap.Div(decimal.NewFromInt(1000000000)).StringFixed(1) + " gwei (cap)"
			}
			
			fmt.Printf("  %d. %s - %s, Priority: %.2f\n", 
				i+1, tx.Hash.Hex()[:10]+"...", gasPrice, tx.Priority.InexactFloat64())
		}
		fmt.Println()
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

func getTrendIcon(trend string) string {
	switch trend {
	case "increasing":
		return "📈 " + trend
	case "decreasing":
		return "📉 " + trend
	case "stable":
		return "➡️ " + trend
	default:
		return "❓ " + trend
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
