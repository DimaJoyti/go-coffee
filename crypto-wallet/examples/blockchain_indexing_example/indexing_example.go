package main

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/blockchain/indexing"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/shopspring/decimal"
)

// Helper function to create sample blocks
func createSampleBlocks(count int, startBlock uint64) []*types.Block {
	var blocks []*types.Block

	for i := 0; i < count; i++ {
		blockNumber := startBlock + uint64(i)

		// Create block header
		header := &types.Header{
			Number:     big.NewInt(int64(blockNumber)),
			Time:       uint64(time.Now().Add(time.Duration(i) * time.Second).Unix()),
			GasLimit:   8000000,
			GasUsed:    uint64(4000000 + i*100000), // Varying gas usage
			Difficulty: big.NewInt(int64(1000000 + i*10000)),
		}

		// Create sample transactions
		var transactions []*types.Transaction
		txCount := 3 + i%5 // Varying transaction count

		for j := 0; j < txCount; j++ {
			to := common.HexToAddress(fmt.Sprintf("0x%040x", j+1))
			value := big.NewInt(int64((j + 1) * 1000000000000000000)) // Varying values
			gasPrice := big.NewInt(int64(15000000000 + j*5000000000)) // Varying gas prices

			tx := types.NewTransaction(
				uint64(j),
				to,
				value,
				21000,
				gasPrice,
				nil,
			)
			transactions = append(transactions, tx)
		}

		// Create block body with transactions
		body := &types.Body{
			Transactions: transactions,
			Uncles:       []*types.Header{},
		}

		// Use the stack trie hasher for compatibility
		hasher := trie.NewStackTrie(nil)
		block := types.NewBlock(header, body, nil, hasher)
		blocks = append(blocks, block)
	}

	return blocks
}

// Helper function to format duration
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", d.Seconds()*1000)
	} else if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
}

// Helper function to format bytes
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func main() {
	fmt.Println("🗂️  Blockchain Data Indexing System Example")
	fmt.Println("===========================================")

	// Initialize logger
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	logger := logger.NewLogger(logConfig)

	// Create data indexer configuration
	config := indexing.GetDefaultDataIndexerConfig()

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Enabled: %v\n", config.Enabled)
	fmt.Printf("  Indexing Interval: %v\n", config.IndexingInterval)
	fmt.Printf("  Batch Size: %d\n", config.BatchSize)
	fmt.Printf("  Max Concurrent Workers: %d\n", config.MaxConcurrentWorkers)
	fmt.Printf("  Start Block: %d\n", config.StartBlock)
	fmt.Printf("  Storage Engine: %s\n", config.StorageConfig.Type)
	fmt.Printf("  Cache Engine: %s\n", config.CacheConfig.Type)
	fmt.Printf("  Block Indexing: %v\n", config.BlockIndexerConfig.Enabled)
	fmt.Printf("  Transaction Indexing: %v\n", config.TransactionConfig.Enabled)
	fmt.Printf("  Address Indexing: %v\n", config.AddressConfig.Enabled)
	fmt.Printf("  Contract Indexing: %v\n", config.ContractConfig.Enabled)
	fmt.Printf("  Event Indexing: %v\n", config.EventConfig.Enabled)
	fmt.Printf("  Analytics: %v\n", config.AnalyticsConfig.Enabled)
	fmt.Println()

	// Create blockchain data indexer
	indexer := indexing.NewBlockchainDataIndexer(logger, config)

	// Start the indexer
	ctx := context.Background()
	if err := indexer.Start(ctx); err != nil {
		fmt.Printf("Failed to start blockchain data indexer: %v\n", err)
		return
	}

	fmt.Println("✅ Blockchain data indexer started successfully!")
	fmt.Println()

	// Show indexer metrics
	fmt.Println("📊 Indexer Metrics:")
	fmt.Println("==================")
	metrics := indexer.GetMetrics()
	fmt.Printf("  Is Running: %v\n", metrics["is_running"])
	fmt.Printf("  Current Block: %v\n", metrics["current_block"])
	fmt.Printf("  Blocks Processed: %v\n", metrics["blocks_processed"])
	fmt.Printf("  Indexing Rate: %s blocks/sec\n", metrics["indexing_rate"])
	fmt.Printf("  Total Blocks: %v\n", metrics["total_blocks"])
	fmt.Printf("  Total Transactions: %v\n", metrics["total_transactions"])
	fmt.Printf("  Total Addresses: %v\n", metrics["total_addresses"])
	fmt.Printf("  Total Contracts: %v\n", metrics["total_contracts"])
	fmt.Printf("  Total Events: %v\n", metrics["total_events"])
	fmt.Printf("  Error Count: %v\n", metrics["error_count"])
	fmt.Printf("  Cache Hit Rate: %s\n", metrics["cache_hit_rate"])
	fmt.Printf("  Storage Size: %s\n", formatBytes(metrics["storage_size"].(uint64)))
	fmt.Println()

	// Demonstrate indexing sample blocks
	fmt.Println("📝 Indexing Sample Blocks:")
	fmt.Println("==========================")

	sampleBlocks := createSampleBlocks(5, 12345)
	fmt.Printf("Created %d sample blocks starting from block %d\n", len(sampleBlocks), 12345)

	for i, block := range sampleBlocks {
		startTime := time.Now()
		err := indexer.IndexBlock(ctx, block)
		indexingTime := time.Since(startTime)

		if err != nil {
			fmt.Printf("  %d. ❌ Block %d - Failed: %v\n", i+1, block.NumberU64(), err)
			continue
		}

		fmt.Printf("  %d. ✅ Block %d - %d txs, %s gas used, indexed in %s\n",
			i+1, block.NumberU64(), len(block.Transactions()),
			formatBytes(block.GasUsed()), formatDuration(indexingTime))
	}
	fmt.Println()

	// Wait for indexing to process
	fmt.Println("⏳ Waiting for indexing to process...")
	time.Sleep(2 * time.Second)

	// Show indexing progress
	fmt.Println("📈 Indexing Progress:")
	fmt.Println("====================")
	progress := indexer.GetIndexingProgress()
	fmt.Printf("  Start Block: %d\n", progress.StartBlock)
	fmt.Printf("  Current Block: %d\n", progress.CurrentBlock)
	fmt.Printf("  Latest Block: %d\n", progress.LatestBlock)
	fmt.Printf("  Blocks Processed: %d\n", progress.BlocksProcessed)
	fmt.Printf("  Blocks Remaining: %d\n", progress.BlocksRemaining)
	fmt.Printf("  Progress: %s%%\n", progress.ProgressPercent.StringFixed(1))
	fmt.Printf("  Estimated Time Remaining: %s\n", formatDuration(progress.EstimatedTime))
	fmt.Printf("  Start Time: %s\n", progress.StartTime.Format("15:04:05"))
	fmt.Printf("  Last Update: %s\n", progress.LastUpdate.Format("15:04:05"))
	fmt.Println()

	// Show indexing statistics
	fmt.Println("📊 Indexing Statistics:")
	fmt.Println("======================")
	stats := indexer.GetIndexingStatistics()
	fmt.Printf("  Total Blocks: %d\n", stats.TotalBlocks)
	fmt.Printf("  Total Transactions: %d\n", stats.TotalTransactions)
	fmt.Printf("  Total Addresses: %d\n", stats.TotalAddresses)
	fmt.Printf("  Total Contracts: %d\n", stats.TotalContracts)
	fmt.Printf("  Total Events: %d\n", stats.TotalEvents)
	fmt.Printf("  Indexing Rate: %s blocks/sec\n", stats.IndexingRate.StringFixed(2))
	fmt.Printf("  Processing Time: %s\n", formatDuration(stats.ProcessingTime))
	fmt.Printf("  Storage Size: %s\n", formatBytes(stats.StorageSize))
	fmt.Printf("  Cache Hit Rate: %s%%\n", stats.CacheHitRate.Mul(decimal.NewFromFloat(100)).StringFixed(1))
	fmt.Printf("  Error Count: %d\n", stats.ErrorCount)
	fmt.Printf("  Last Updated: %s\n", stats.LastUpdated.Format("15:04:05"))
	fmt.Println()

	// Demonstrate data retrieval
	fmt.Println("🔍 Data Retrieval Examples:")
	fmt.Println("===========================")

	// Retrieve a block
	blockNumber := uint64(12345)
	block, err := indexer.GetBlock(blockNumber)
	if err != nil {
		fmt.Printf("Failed to retrieve block %d: %v\n", blockNumber, err)
	} else {
		fmt.Printf("📦 Block %d:\n", blockNumber)
		fmt.Printf("    Hash: %s\n", block.Hash.Hex()[:10]+"...")
		fmt.Printf("    Timestamp: %s\n", block.Timestamp.Format("15:04:05"))
		fmt.Printf("    Gas Limit: %s\n", formatBytes(block.GasLimit))
		fmt.Printf("    Gas Used: %s\n", formatBytes(block.GasUsed))
		fmt.Printf("    Transactions: %d\n", block.TransactionCount)
		fmt.Printf("    Size: %s\n", formatBytes(block.Size))
	}
	fmt.Println()

	// Retrieve a transaction
	txHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	tx, err := indexer.GetTransaction(txHash)
	if err != nil {
		fmt.Printf("Failed to retrieve transaction: %v\n", err)
	} else {
		fmt.Printf("💸 Transaction %s:\n", txHash.Hex()[:10]+"...")
		fmt.Printf("    Block: %d\n", tx.BlockNumber)
		fmt.Printf("    From: %s\n", tx.From.Hex()[:10]+"...")
		if tx.To != nil {
			fmt.Printf("    To: %s\n", tx.To.Hex()[:10]+"...")
		}
		fmt.Printf("    Value: %s ETH\n", tx.Value.Div(decimal.NewFromInt(1000000000000000000)).StringFixed(4))
		fmt.Printf("    Gas Used: %d\n", tx.GasUsed)
		fmt.Printf("    Status: %s\n", map[uint64]string{0: "Failed", 1: "Success"}[tx.Status])
	}
	fmt.Println()

	// Retrieve an address
	address := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1")
	addr, err := indexer.GetAddress(address)
	if err != nil {
		fmt.Printf("Failed to retrieve address: %v\n", err)
	} else {
		fmt.Printf("👤 Address %s:\n", address.Hex()[:10]+"...")
		fmt.Printf("    Type: %s\n", strings.ToUpper(string(addr.Type[0]))+addr.Type[1:])
		fmt.Printf("    Balance: %s ETH\n", addr.Balance.Div(decimal.NewFromInt(1000000000000000000)).StringFixed(4))
		fmt.Printf("    Nonce: %d\n", addr.Nonce)
		fmt.Printf("    Transactions: %d\n", addr.TransactionCount)
		fmt.Printf("    First Seen: %s\n", addr.FirstSeen.Format("2006-01-02 15:04:05"))
		fmt.Printf("    Last Activity: %s\n", addr.LastActivity.Format("2006-01-02 15:04:05"))
		fmt.Printf("    Token Holdings: %d\n", len(addr.TokenHoldings))
		fmt.Printf("    NFT Holdings: %d\n", len(addr.NFTHoldings))
	}
	fmt.Println()

	// Demonstrate analytics
	fmt.Println("📈 Analytics Example:")
	fmt.Println("====================")
	timeRange := indexing.TimeRange{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}

	analytics, err := indexer.GetAnalytics(ctx, timeRange)
	if err != nil {
		fmt.Printf("Failed to get analytics: %v\n", err)
	} else {
		fmt.Printf("📊 Block Metrics (24h):\n")
		fmt.Printf("    Total Blocks: %d\n", analytics.BlockMetrics.TotalBlocks)
		fmt.Printf("    Average Block Time: %ss\n", analytics.BlockMetrics.AverageBlockTime.StringFixed(1))
		fmt.Printf("    Average Gas Used: %s\n", formatBytes(uint64(analytics.BlockMetrics.AverageGasUsed.IntPart())))
		fmt.Printf("    Total Gas Used: %s\n", formatBytes(analytics.BlockMetrics.TotalGasUsed))
		fmt.Println()

		fmt.Printf("💸 Transaction Metrics (24h):\n")
		fmt.Printf("    Total Transactions: %d\n", analytics.TransactionMetrics.TotalTransactions)
		fmt.Printf("    Average Gas Price: %s gwei\n",
			analytics.TransactionMetrics.AverageGasPrice.Div(decimal.NewFromInt(1000000000)).StringFixed(1))
		fmt.Printf("    Average Value: %s ETH\n",
			analytics.TransactionMetrics.AverageValue.Div(decimal.NewFromInt(1000000000000000000)).StringFixed(4))
		fmt.Printf("    Success Rate: %s%%\n",
			analytics.TransactionMetrics.SuccessRate.Mul(decimal.NewFromFloat(100)).StringFixed(1))
		fmt.Println()

		fmt.Printf("👥 Address Metrics (24h):\n")
		fmt.Printf("    Total Addresses: %d\n", analytics.AddressMetrics.TotalAddresses)
		fmt.Printf("    Active Addresses: %d\n", analytics.AddressMetrics.ActiveAddresses)
		fmt.Printf("    New Addresses: %d\n", analytics.AddressMetrics.NewAddresses)
		fmt.Println()

		fmt.Printf("📋 Contract Metrics (24h):\n")
		fmt.Printf("    Total Contracts: %d\n", analytics.ContractMetrics.TotalContracts)
		fmt.Printf("    New Contracts: %d\n", analytics.ContractMetrics.NewContracts)
		fmt.Printf("    Verified Contracts: %d\n", analytics.ContractMetrics.VerifiedContracts)
		fmt.Printf("    Popular Contracts: %d\n", len(analytics.ContractMetrics.PopularContracts))
	}
	fmt.Println()

	// Show different configuration profiles
	fmt.Println("🔧 Configuration Profiles:")
	fmt.Println("==========================")

	// High throughput config
	htConfig := indexing.GetHighThroughputConfig()
	fmt.Printf("⚡ High Throughput:\n")
	fmt.Printf("  Indexing Interval: %v (vs %v default)\n", htConfig.IndexingInterval, config.IndexingInterval)
	fmt.Printf("  Batch Size: %d (vs %d default)\n", htConfig.BatchSize, config.BatchSize)
	fmt.Printf("  Max Workers: %d (vs %d default)\n", htConfig.MaxConcurrentWorkers, config.MaxConcurrentWorkers)
	fmt.Printf("  Cache Size: %d (vs %d default)\n", htConfig.CacheConfig.MaxSize, config.CacheConfig.MaxSize)
	fmt.Printf("  Real-time Analytics: %v (optimized for speed)\n", htConfig.AnalyticsConfig.RealTimeAnalytics)
	fmt.Println()

	// Low latency config
	llConfig := indexing.GetLowLatencyConfig()
	fmt.Printf("🚀 Low Latency:\n")
	fmt.Printf("  Indexing Interval: %v (optimized for speed)\n", llConfig.IndexingInterval)
	fmt.Printf("  Batch Size: %d (smaller batches)\n", llConfig.BatchSize)
	fmt.Printf("  Cache Size: %d (aggressive caching)\n", llConfig.CacheConfig.MaxSize)
	fmt.Printf("  Cache TTL: %v (longer retention)\n", llConfig.CacheConfig.TTL)
	fmt.Printf("  Real-time Analytics: %v\n", llConfig.AnalyticsConfig.RealTimeAnalytics)
	fmt.Println()

	// Archival config
	archivalConfig := indexing.GetArchivalConfig()
	fmt.Printf("🗄️  Archival:\n")
	fmt.Printf("  Indexing Interval: %v (comprehensive)\n", archivalConfig.IndexingInterval)
	fmt.Printf("  Batch Size: %d (larger batches)\n", archivalConfig.BatchSize)
	fmt.Printf("  Index Uncles: %v (complete data)\n", archivalConfig.BlockIndexerConfig.IndexUncles)
	fmt.Printf("  Index All Events: %v (comprehensive)\n", archivalConfig.EventConfig.IndexAllEvents)
	fmt.Printf("  Retention Period: %v (long-term)\n", archivalConfig.BlockIndexerConfig.RetentionPeriod)
	fmt.Printf("  Compression: %v (storage optimization)\n", archivalConfig.StorageConfig.CompressionEnabled)
	fmt.Println()

	// Show supported features
	fmt.Println("🛠️  Supported Features:")
	fmt.Println("======================")

	fmt.Println("Storage Engines:")
	storageTypes := indexing.GetSupportedStorageTypes()
	for i, storageType := range storageTypes {
		if i > 0 && i%3 == 0 {
			fmt.Println()
		}
		fmt.Printf("  %-12s", storageType)
	}
	fmt.Println("\n")

	fmt.Println("Cache Engines:")
	cacheTypes := indexing.GetSupportedCacheTypes()
	for i, cacheType := range cacheTypes {
		if i > 0 && i%4 == 0 {
			fmt.Println()
		}
		fmt.Printf("  %-12s", cacheType)
	}
	fmt.Println("\n")

	fmt.Println("Contract Types:")
	contractTypes := indexing.GetSupportedContractTypes()
	for i, contractType := range contractTypes {
		if i > 0 && i%5 == 0 {
			fmt.Println()
		}
		fmt.Printf("  %-10s", contractType)
	}
	fmt.Println("\n")

	fmt.Println("Event Names:")
	eventNames := indexing.GetSupportedEventNames()
	for i, eventName := range eventNames {
		if i > 0 && i%5 == 0 {
			fmt.Println()
		}
		fmt.Printf("  %-12s", eventName)
	}
	fmt.Println("\n")

	// Best practices
	fmt.Println("💡 Best Practices:")
	fmt.Println("==================")
	fmt.Println("1. Choose appropriate configuration profile for your use case")
	fmt.Println("2. Monitor indexing progress and performance metrics regularly")
	fmt.Println("3. Implement proper error handling and retry mechanisms")
	fmt.Println("4. Use caching effectively to improve query performance")
	fmt.Println("5. Partition data by time or block ranges for better performance")
	fmt.Println("6. Implement data retention policies to manage storage costs")
	fmt.Println("7. Use analytics to gain insights into blockchain activity")
	fmt.Println("8. Regularly backup indexed data for disaster recovery")
	fmt.Println()

	fmt.Println("🎉 Blockchain Data Indexing System example completed!")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("  ✅ Comprehensive blockchain data indexing (blocks, transactions, addresses, contracts, events)")
	fmt.Println("  ✅ Multi-tier storage architecture with caching")
	fmt.Println("  ✅ Real-time indexing progress tracking")
	fmt.Println("  ✅ Performance metrics and analytics")
	fmt.Println("  ✅ Configurable indexing strategies")
	fmt.Println("  ✅ Data retrieval and query capabilities")
	fmt.Println("  ✅ Scalable architecture with concurrent processing")
	fmt.Println("  ✅ Data validation and enrichment pipelines")
	fmt.Println()
	fmt.Println("Note: This example demonstrates the system with mock data and storage.")
	fmt.Println("Integrate with real blockchain nodes and databases for production use.")

	// Stop the indexer
	if err := indexer.Stop(); err != nil {
		fmt.Printf("Error stopping blockchain data indexer: %v\n", err)
	} else {
		fmt.Println("\n🛑 Blockchain data indexer stopped")
	}
}
