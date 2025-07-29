package main

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/blockchain/monitoring"
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
	
	// Create a mix of legacy and EIP-1559 transactions
	gasPrices := []*big.Int{
		big.NewInt(15000000000),  // 15 gwei
		big.NewInt(20000000000),  // 20 gwei
		big.NewInt(25000000000),  // 25 gwei
		big.NewInt(30000000000),  // 30 gwei
		big.NewInt(40000000000),  // 40 gwei
	}
	
	values := []*big.Int{
		big.NewInt(100000000000000000),  // 0.1 ETH
		big.NewInt(500000000000000000),  // 0.5 ETH
		big.NewInt(1000000000000000000), // 1 ETH
		big.NewInt(2000000000000000000), // 2 ETH
		big.NewInt(5000000000000000000), // 5 ETH
	}
	
	// Create legacy transactions
	for i, gasPrice := range gasPrices[:3] {
		tx := createTestTransaction(gasPrice, 21000, uint64(i+1), values[i%len(values)])
		transactions = append(transactions, tx)
	}
	
	// Create EIP-1559 transactions
	for i, gasPrice := range gasPrices[3:] {
		tipCap := big.NewInt(2000000000) // 2 gwei tip
		tx := createEIP1559Transaction(gasPrice, tipCap, 21000, uint64(i+4), values[i%len(values)])
		transactions = append(transactions, tx)
	}
	
	return transactions
}

// Helper function to get status icon
func getStatusIcon(status monitoring.TransactionStatus) string {
	switch status {
	case monitoring.StatusPending:
		return "⏳"
	case monitoring.StatusConfirming:
		return "🔄"
	case monitoring.StatusConfirmed:
		return "✅"
	case monitoring.StatusFailed:
		return "❌"
	case monitoring.StatusDropped:
		return "🗑️"
	case monitoring.StatusReplaced:
		return "🔄"
	case monitoring.StatusStuck:
		return "🚫"
	default:
		return "❓"
	}
}

// Helper function to get severity icon
func getSeverityIcon(severity string) string {
	switch severity {
	case "info":
		return "ℹ️"
	case "warning":
		return "⚠️"
	case "error":
		return "🚨"
	case "critical":
		return "🔥"
	default:
		return "📢"
	}
}

func main() {
	fmt.Println("📊 Transaction Monitoring System Example")
	fmt.Println("========================================")

	// Initialize logger
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	logger := logger.NewLogger(logConfig)

	// Create transaction monitor configuration
	config := monitoring.GetDefaultTransactionMonitorConfig()

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Enabled: %v\n", config.Enabled)
	fmt.Printf("  Update Interval: %v\n", config.UpdateInterval)
	fmt.Printf("  Max Tracked Transactions: %d\n", config.MaxTrackedTransactions)
	fmt.Printf("  History Retention: %v\n", config.HistoryRetentionPeriod)
	fmt.Printf("  Confirmation Tracking: %v\n", config.ConfirmationConfig.Enabled)
	fmt.Printf("  Failure Detection: %v\n", config.FailureConfig.Enabled)
	fmt.Printf("  Retry Management: %v\n", config.RetryConfig.Enabled)
	fmt.Printf("  Alert Management: %v\n", config.AlertConfig.Enabled)
	fmt.Println()

	// Create transaction monitor
	monitor := monitoring.NewTransactionMonitor(logger, config)

	// Start the monitor
	ctx := context.Background()
	if err := monitor.Start(ctx); err != nil {
		fmt.Printf("Failed to start transaction monitor: %v\n", err)
		return
	}

	fmt.Println("✅ Transaction monitor started successfully!")
	fmt.Println()

	// Show monitor metrics
	fmt.Println("📊 Monitor Metrics:")
	fmt.Println("==================")
	metrics := monitor.GetMetrics()
	fmt.Printf("  Is Running: %v\n", metrics["is_running"])
	fmt.Printf("  Total Tracked: %v\n", metrics["total_tracked"])
	fmt.Printf("  Confirmation Enabled: %v\n", metrics["confirmation_enabled"])
	fmt.Printf("  Failure Detection Enabled: %v\n", metrics["failure_detection_enabled"])
	fmt.Printf("  Retry Enabled: %v\n", metrics["retry_enabled"])
	fmt.Printf("  Alert Enabled: %v\n", metrics["alert_enabled"])
	fmt.Println()

	// Create and track sample transactions
	fmt.Println("📝 Tracking Sample Transactions:")
	fmt.Println("================================")
	
	transactions := createSampleTransactions()
	fmt.Printf("Created %d sample transactions\n", len(transactions))
	
	for i, tx := range transactions {
		// Create metadata for each transaction
		metadata := map[string]interface{}{
			"user_id":     fmt.Sprintf("user_%d", i+1),
			"purpose":     "example_transfer",
			"priority":    []string{"low", "medium", "high"}[i%3],
			"created_at":  time.Now(),
		}
		
		err := monitor.TrackTransaction(tx, metadata)
		if err != nil {
			fmt.Printf("Failed to track transaction %d: %v\n", i+1, err)
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
		
		fmt.Printf("  %d. %s - %s, Value: %s ETH, Priority: %s\n", 
			i+1, txType, gasInfo, 
			new(big.Int).Div(tx.Value(), big.NewInt(1000000000000000000)).String(),
			metadata["priority"])
	}
	fmt.Println()

	// Wait for some monitoring updates
	fmt.Println("⏳ Waiting for monitoring updates...")
	time.Sleep(3 * time.Second)

	// Show transaction statuses
	fmt.Println("📋 Transaction Status Overview:")
	fmt.Println("==============================")
	
	trackedTransactions := monitor.GetTrackedTransactions()
	for i, tx := range transactions {
		if trackedTx, exists := trackedTransactions[tx.Hash()]; exists {
			fmt.Printf("  %d. %s %s - %s (%d confirmations)\n",
				i+1,
				getStatusIcon(trackedTx.Status),
				tx.Hash().Hex()[:10]+"...",
				strings.Title(string(trackedTx.Status)),
				trackedTx.Confirmations)
			
			if trackedTx.BlockNumber != nil {
				fmt.Printf("      Block: %d", *trackedTx.BlockNumber)
				if trackedTx.TransactionIndex != nil {
					fmt.Printf(", Index: %d", *trackedTx.TransactionIndex)
				}
				fmt.Println()
			}
			
			if trackedTx.GasUsed != nil && trackedTx.EffectiveGasPrice != nil {
				cost := trackedTx.EffectiveGasPrice.Mul(decimal.NewFromInt(int64(*trackedTx.GasUsed)))
				fmt.Printf("      Gas Used: %d, Cost: %s ETH\n", 
					*trackedTx.GasUsed, 
					cost.Div(decimal.NewFromInt(1000000000000000000)).StringFixed(6))
			}
			
			if len(trackedTx.Alerts) > 0 {
				fmt.Printf("      Alerts: %d active\n", len(trackedTx.Alerts))
				for _, alert := range trackedTx.Alerts {
					fmt.Printf("        %s %s: %s\n", 
						getSeverityIcon(alert.Severity), 
						alert.Title, 
						alert.Message)
				}
			}
		}
	}
	fmt.Println()

	// Show comprehensive monitoring results
	fmt.Println("📊 Comprehensive Monitoring Results:")
	fmt.Println("===================================")
	
	result := monitor.GetMonitoringResult()
	fmt.Printf("  Total Tracked: %d\n", result.TotalTracked)
	fmt.Printf("  Pending: %d\n", result.PendingCount)
	fmt.Printf("  Confirming: %d\n", result.ConfirmingCount)
	fmt.Printf("  Confirmed: %d\n", result.ConfirmedCount)
	fmt.Printf("  Failed: %d\n", result.FailedCount)
	fmt.Printf("  Stuck: %d\n", result.StuckCount)
	fmt.Printf("  Success Rate: %s%%\n", result.SuccessRate.Mul(decimal.NewFromFloat(100)).StringFixed(1))
	fmt.Printf("  Active Alerts: %d\n", result.ActiveAlerts)
	fmt.Printf("  Average Confirmation Time: %v\n", result.AverageConfirmTime)
	fmt.Println()

	// Show performance metrics
	if result.PerformanceMetrics != nil {
		fmt.Println("⚡ Performance Metrics:")
		fmt.Println("======================")
		metrics := result.PerformanceMetrics
		
		if !metrics.AverageGasUsed.IsZero() {
			fmt.Printf("  Average Gas Used: %s\n", metrics.AverageGasUsed.StringFixed(0))
		}
		if !metrics.AverageGasPrice.IsZero() {
			fmt.Printf("  Average Gas Price: %s gwei\n", 
				metrics.AverageGasPrice.Div(decimal.NewFromInt(1000000000)).StringFixed(1))
		}
		if !metrics.TotalGasCost.IsZero() {
			fmt.Printf("  Total Gas Cost: %s ETH\n", 
				metrics.TotalGasCost.Div(decimal.NewFromInt(1000000000000000000)).StringFixed(6))
		}
		
		if len(metrics.ConfirmationTimes) > 0 {
			fmt.Printf("  Confirmation Times: %d samples\n", len(metrics.ConfirmationTimes))
		}
		
		if len(metrics.FailureReasons) > 0 {
			fmt.Printf("  Failure Reasons:\n")
			for reason, count := range metrics.FailureReasons {
				fmt.Printf("    %s: %d\n", reason, count)
			}
		}
		
		if metrics.RetryStatistics != nil {
			retry := metrics.RetryStatistics
			fmt.Printf("  Retry Statistics:\n")
			fmt.Printf("    Total Retries: %d\n", retry.TotalRetries)
			fmt.Printf("    Successful Retries: %d\n", retry.SuccessfulRetries)
			fmt.Printf("    Retry Success Rate: %s%%\n", 
				retry.RetrySuccessRate.Mul(decimal.NewFromFloat(100)).StringFixed(1))
		}
		fmt.Println()
	}

	// Show recent events
	if len(result.RecentEvents) > 0 {
		fmt.Println("📅 Recent Events:")
		fmt.Println("================")
		for i, event := range result.RecentEvents {
			if i >= 5 { // Show only last 5 events
				break
			}
			fmt.Printf("  %s %s - %s (%s)\n",
				getSeverityIcon(event.Severity),
				event.Timestamp.Format("15:04:05"),
				strings.Replace(event.EventType, "_", " ", -1),
				event.TransactionHash.Hex()[:10]+"...")
		}
		fmt.Println()
	}

	// Show different configuration profiles
	fmt.Println("🔧 Configuration Profiles:")
	fmt.Println("==========================")
	
	// High frequency config
	hfConfig := monitoring.GetHighFrequencyConfig()
	fmt.Printf("⚡ High Frequency:\n")
	fmt.Printf("  Update Interval: %v (vs %v default)\n", hfConfig.UpdateInterval, config.UpdateInterval)
	fmt.Printf("  Max Tracked: %d (vs %d default)\n", hfConfig.MaxTrackedTransactions, config.MaxTrackedTransactions)
	fmt.Printf("  Required Confirmations: %d (vs %d default)\n", 
		hfConfig.ConfirmationConfig.RequiredConfirmations, 
		config.ConfirmationConfig.RequiredConfirmations)
	fmt.Println()

	// Low latency config
	llConfig := monitoring.GetLowLatencyConfig()
	fmt.Printf("🚀 Low Latency:\n")
	fmt.Printf("  Update Interval: %v (optimized for speed)\n", llConfig.UpdateInterval)
	fmt.Printf("  Required Confirmations: %d (minimal)\n", llConfig.ConfirmationConfig.RequiredConfirmations)
	fmt.Printf("  Max Confirmation Time: %v (aggressive)\n", llConfig.ConfirmationConfig.MaxConfirmationTime)
	fmt.Println()

	// Robust config
	robustConfig := monitoring.GetRobustConfig()
	fmt.Printf("🛡️  Robust:\n")
	fmt.Printf("  Update Interval: %v (conservative)\n", robustConfig.UpdateInterval)
	fmt.Printf("  Required Confirmations: %d (secure)\n", robustConfig.ConfirmationConfig.RequiredConfirmations)
	fmt.Printf("  History Retention: %v (extended)\n", robustConfig.HistoryRetentionPeriod)
	fmt.Println()

	// Show supported features
	fmt.Println("🛠️  Supported Features:")
	fmt.Println("======================")
	
	fmt.Println("Failure Detection Methods:")
	methods := monitoring.GetSupportedFailureDetectionMethods()
	for i, method := range methods {
		if i > 0 && i%3 == 0 {
			fmt.Println()
		}
		fmt.Printf("  %-20s", strings.Replace(method, "_", " ", -1))
	}
	fmt.Println("\n")

	fmt.Println("Retry Strategies:")
	strategies := monitoring.GetSupportedRetryStrategies()
	for i, strategy := range strategies {
		if i > 0 && i%3 == 0 {
			fmt.Println()
		}
		fmt.Printf("  %-20s", strings.Replace(strategy, "_", " ", -1))
	}
	fmt.Println("\n")

	fmt.Println("Alert Channels:")
	channels := monitoring.GetSupportedAlertChannels()
	for i, channel := range channels {
		if i > 0 && i%4 == 0 {
			fmt.Println()
		}
		fmt.Printf("  %-15s", channel)
	}
	fmt.Println("\n")

	// Show status descriptions
	fmt.Println("📊 Transaction Status Guide:")
	fmt.Println("===========================")
	descriptions := monitoring.GetTransactionStatusDescription()
	for status, desc := range descriptions {
		fmt.Printf("  %s %s: %s\n", getStatusIcon(status), strings.Title(string(status)), desc)
	}
	fmt.Println()

	// Best practices
	fmt.Println("💡 Best Practices:")
	fmt.Println("==================")
	fmt.Println("1. Monitor transaction status regularly for timely intervention")
	fmt.Println("2. Set appropriate confirmation requirements based on transaction value")
	fmt.Println("3. Configure retry strategies for automatic failure recovery")
	fmt.Println("4. Use alerts to stay informed about transaction issues")
	fmt.Println("5. Analyze performance metrics to optimize gas usage")
	fmt.Println("6. Implement proper error handling for failed transactions")
	fmt.Println("7. Consider network congestion when setting gas prices")
	fmt.Println("8. Keep transaction history for audit and analysis purposes")
	fmt.Println()

	fmt.Println("🎉 Transaction Monitoring System example completed!")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("  ✅ Real-time transaction tracking and status monitoring")
	fmt.Println("  ✅ Confirmation tracking with block reorganization protection")
	fmt.Println("  ✅ Intelligent failure detection and analysis")
	fmt.Println("  ✅ Automatic retry management with configurable strategies")
	fmt.Println("  ✅ Comprehensive alerting system")
	fmt.Println("  ✅ Performance metrics and analytics")
	fmt.Println("  ✅ Configurable monitoring profiles")
	fmt.Println("  ✅ Event history and audit trail")
	fmt.Println()
	fmt.Println("Note: This example demonstrates the system with mock data.")
	fmt.Println("Integrate with real blockchain nodes for production use.")

	// Stop the monitor
	if err := monitor.Stop(); err != nil {
		fmt.Printf("Error stopping transaction monitor: %v\n", err)
	} else {
		fmt.Println("\n🛑 Transaction monitor stopped")
	}
}
