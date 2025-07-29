package indexing

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test logger
func createTestLoggerForIndexing() *logger.Logger {
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	return logger.NewLogger(logConfig)
}

// Helper function to create test block
func createTestBlockForIndexing(blockNumber uint64) *types.Block {
	header := &types.Header{
		Number:     big.NewInt(int64(blockNumber)),
		Time:       uint64(time.Now().Unix()),
		GasLimit:   8000000,
		GasUsed:    4000000,
		Difficulty: big.NewInt(1000000),
	}

	// Create test transactions
	var transactions []*types.Transaction
	for i := 0; i < 3; i++ {
		tx := types.NewTransaction(
			uint64(i),
			common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1"),
			big.NewInt(1000000000000000000), // 1 ETH
			21000,
			big.NewInt(20000000000), // 20 gwei
			nil,
		)
		transactions = append(transactions, tx)
	}

	body := &types.Body{
		Transactions: transactions,
	}
	return types.NewBlock(header, body, nil, trie.NewStackTrie(nil))
}

func TestNewBlockchainDataIndexer(t *testing.T) {
	logger := createTestLoggerForIndexing()
	config := GetDefaultDataIndexerConfig()

	indexer := NewBlockchainDataIndexer(logger, config)

	assert.NotNil(t, indexer)
	assert.Equal(t, config.Enabled, indexer.config.Enabled)
	assert.Equal(t, config.IndexingInterval, indexer.config.IndexingInterval)
	assert.False(t, indexer.IsRunning())
	assert.NotNil(t, indexer.blockIndexer)
	assert.NotNil(t, indexer.transactionIndexer)
	assert.NotNil(t, indexer.addressIndexer)
	assert.NotNil(t, indexer.contractIndexer)
	assert.NotNil(t, indexer.eventIndexer)
	assert.NotNil(t, indexer.storageEngine)
	assert.NotNil(t, indexer.cacheEngine)
	assert.NotNil(t, indexer.dataProcessor)
	assert.NotNil(t, indexer.analyticsEngine)
}

func TestBlockchainDataIndexer_StartStop(t *testing.T) {
	logger := createTestLoggerForIndexing()
	config := GetDefaultDataIndexerConfig()

	indexer := NewBlockchainDataIndexer(logger, config)
	ctx := context.Background()

	err := indexer.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, indexer.IsRunning())

	err = indexer.Stop()
	assert.NoError(t, err)
	assert.False(t, indexer.IsRunning())
}

func TestBlockchainDataIndexer_StartDisabled(t *testing.T) {
	logger := createTestLoggerForIndexing()
	config := GetDefaultDataIndexerConfig()
	config.Enabled = false

	indexer := NewBlockchainDataIndexer(logger, config)
	ctx := context.Background()

	err := indexer.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, indexer.IsRunning()) // Should remain false when disabled
}

func TestBlockchainDataIndexer_IndexBlock(t *testing.T) {
	logger := createTestLoggerForIndexing()
	config := GetDefaultDataIndexerConfig()

	indexer := NewBlockchainDataIndexer(logger, config)
	ctx := context.Background()

	// Start the indexer
	err := indexer.Start(ctx)
	require.NoError(t, err)
	defer indexer.Stop()

	// Create test block
	block := createTestBlockForIndexing(12345)

	// Index the block
	err = indexer.IndexBlock(ctx, block)
	assert.NoError(t, err)

	// Verify indexing progress was updated
	progress := indexer.GetIndexingProgress()
	assert.Equal(t, uint64(12345), progress.CurrentBlock)
	assert.True(t, progress.BlocksProcessed > 0)
}

func TestBlockchainDataIndexer_GetBlock(t *testing.T) {
	logger := createTestLoggerForIndexing()
	config := GetDefaultDataIndexerConfig()

	indexer := NewBlockchainDataIndexer(logger, config)

	// Test getting a block (should use mock implementation)
	block, err := indexer.GetBlock(12345)
	assert.NoError(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, uint64(12345), block.Number)
}

func TestBlockchainDataIndexer_GetTransaction(t *testing.T) {
	logger := createTestLoggerForIndexing()
	config := GetDefaultDataIndexerConfig()

	indexer := NewBlockchainDataIndexer(logger, config)

	// Test getting a transaction (should use mock implementation)
	txHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	tx, err := indexer.GetTransaction(txHash)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, txHash, tx.Hash)
}

func TestBlockchainDataIndexer_GetAddress(t *testing.T) {
	logger := createTestLoggerForIndexing()
	config := GetDefaultDataIndexerConfig()

	indexer := NewBlockchainDataIndexer(logger, config)

	// Test getting an address (should use mock implementation)
	address := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1")
	addr, err := indexer.GetAddress(address)
	assert.NoError(t, err)
	assert.NotNil(t, addr)
	assert.Equal(t, address, addr.Address)
}

func TestBlockchainDataIndexer_GetTransactionsByAddress(t *testing.T) {
	logger := createTestLoggerForIndexing()
	config := GetDefaultDataIndexerConfig()

	indexer := NewBlockchainDataIndexer(logger, config)

	// Test getting transactions by address
	address := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1")
	txs, err := indexer.GetTransactionsByAddress(address, 10)
	assert.NoError(t, err)
	assert.NotNil(t, txs)
	assert.True(t, len(txs) <= 10)
}

func TestBlockchainDataIndexer_GetAnalytics(t *testing.T) {
	logger := createTestLoggerForIndexing()
	config := GetDefaultDataIndexerConfig()

	indexer := NewBlockchainDataIndexer(logger, config)
	ctx := context.Background()

	// Test getting analytics
	timeRange := TimeRange{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}

	analytics, err := indexer.GetAnalytics(ctx, timeRange)
	assert.NoError(t, err)
	assert.NotNil(t, analytics)
	assert.NotNil(t, analytics.BlockMetrics)
	assert.NotNil(t, analytics.TransactionMetrics)
	assert.NotNil(t, analytics.AddressMetrics)
	assert.NotNil(t, analytics.ContractMetrics)
}

func TestBlockchainDataIndexer_GetIndexingProgress(t *testing.T) {
	logger := createTestLoggerForIndexing()
	config := GetDefaultDataIndexerConfig()

	indexer := NewBlockchainDataIndexer(logger, config)

	progress := indexer.GetIndexingProgress()
	assert.NotNil(t, progress)
	assert.Equal(t, config.StartBlock, progress.StartBlock)
	assert.False(t, progress.StartTime.IsZero())
}

func TestBlockchainDataIndexer_GetIndexingStatistics(t *testing.T) {
	logger := createTestLoggerForIndexing()
	config := GetDefaultDataIndexerConfig()

	indexer := NewBlockchainDataIndexer(logger, config)

	stats := indexer.GetIndexingStatistics()
	assert.NotNil(t, stats)
	assert.False(t, stats.LastUpdated.IsZero())
}

func TestBlockchainDataIndexer_GetMetrics(t *testing.T) {
	logger := createTestLoggerForIndexing()
	config := GetDefaultDataIndexerConfig()

	indexer := NewBlockchainDataIndexer(logger, config)

	metrics := indexer.GetMetrics()
	assert.NotNil(t, metrics)

	// Validate metrics structure
	assert.Contains(t, metrics, "is_running")
	assert.Contains(t, metrics, "current_block")
	assert.Contains(t, metrics, "blocks_processed")
	assert.Contains(t, metrics, "indexing_rate")
	assert.Contains(t, metrics, "total_blocks")
	assert.Contains(t, metrics, "block_indexer_enabled")
	assert.Contains(t, metrics, "transaction_indexer_enabled")
	assert.Contains(t, metrics, "address_indexer_enabled")
	assert.Contains(t, metrics, "contract_indexer_enabled")
	assert.Contains(t, metrics, "event_indexer_enabled")
	assert.Contains(t, metrics, "analytics_enabled")

	assert.Equal(t, false, metrics["is_running"])
	assert.Equal(t, true, metrics["block_indexer_enabled"])
	assert.Equal(t, true, metrics["transaction_indexer_enabled"])
	assert.Equal(t, true, metrics["address_indexer_enabled"])
	assert.Equal(t, true, metrics["contract_indexer_enabled"])
	assert.Equal(t, true, metrics["event_indexer_enabled"])
	assert.Equal(t, true, metrics["analytics_enabled"])
}

func TestMockBlockIndexer(t *testing.T) {
	indexer := &MockBlockIndexer{}
	ctx := context.Background()

	// Test block indexing
	block := createTestBlockForIndexing(12345)
	indexedBlock, err := indexer.IndexBlock(ctx, block)
	assert.NoError(t, err)
	assert.NotNil(t, indexedBlock)
	assert.Equal(t, block.NumberU64(), indexedBlock.Number)
	assert.Equal(t, block.Hash(), indexedBlock.Hash)

	// Test block retrieval
	retrievedBlock, err := indexer.GetBlock(12345)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedBlock)
	assert.Equal(t, uint64(12345), retrievedBlock.Number)
}

func TestMockTransactionIndexer(t *testing.T) {
	indexer := &MockTransactionIndexer{}
	ctx := context.Background()

	// Create test transaction
	tx := types.NewTransaction(
		1,
		common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1"),
		big.NewInt(1000000000000000000), // 1 ETH
		21000,
		big.NewInt(20000000000), // 20 gwei
		nil,
	)

	// Create mock receipt
	receipt := &types.Receipt{
		Status:            types.ReceiptStatusSuccessful,
		CumulativeGasUsed: 21000,
		BlockNumber:       big.NewInt(12345),
		BlockHash:         common.HexToHash("0xabcd"),
		TxHash:            tx.Hash(),
		GasUsed:           21000,
	}

	// Test transaction indexing
	indexedTx, err := indexer.IndexTransaction(ctx, tx, receipt)
	assert.NoError(t, err)
	assert.NotNil(t, indexedTx)
	assert.Equal(t, tx.Hash(), indexedTx.Hash)
	assert.Equal(t, receipt.BlockNumber.Uint64(), indexedTx.BlockNumber)

	// Test transaction retrieval
	retrievedTx, err := indexer.GetTransaction(tx.Hash())
	assert.NoError(t, err)
	assert.NotNil(t, retrievedTx)
	assert.Equal(t, tx.Hash(), retrievedTx.Hash)

	// Test transactions by address
	address := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1")
	txs, err := indexer.GetTransactionsByAddress(address, 5)
	assert.NoError(t, err)
	assert.NotNil(t, txs)
	assert.True(t, len(txs) <= 5)
}

func TestMockAddressIndexer(t *testing.T) {
	indexer := &MockAddressIndexer{}
	ctx := context.Background()

	address := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1")

	// Test address indexing
	indexedAddr, err := indexer.IndexAddress(ctx, address)
	assert.NoError(t, err)
	assert.NotNil(t, indexedAddr)
	assert.Equal(t, address, indexedAddr.Address)
	assert.Equal(t, "eoa", indexedAddr.Type)

	// Test address retrieval
	retrievedAddr, err := indexer.GetAddress(address)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedAddr)
	assert.Equal(t, address, retrievedAddr.Address)
}

func TestMockAnalyticsEngine(t *testing.T) {
	engine := &MockAnalyticsEngine{}
	ctx := context.Background()

	// Test analytics
	result, err := engine.Analyze(ctx, "test_data")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Metrics)
	assert.NotEmpty(t, result.Insights)

	// Test metrics
	timeRange := TimeRange{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}
	metrics, err := engine.GetMetrics(ctx, timeRange)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.BlockMetrics)
	assert.NotNil(t, metrics.TransactionMetrics)

	// Test report generation
	report, err := engine.GenerateReport(ctx, "daily_summary")
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "daily_summary", report.ReportType)
}

func TestGetDefaultDataIndexerConfig(t *testing.T) {
	config := GetDefaultDataIndexerConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, 5*time.Second, config.IndexingInterval)
	assert.Equal(t, 10, config.BatchSize)
	assert.Equal(t, 4, config.MaxConcurrentWorkers)

	// Check block indexer config
	assert.True(t, config.BlockIndexerConfig.Enabled)
	assert.True(t, config.BlockIndexerConfig.IndexHeaders)
	assert.True(t, config.BlockIndexerConfig.IndexTransactions)

	// Check transaction config
	assert.True(t, config.TransactionConfig.Enabled)
	assert.True(t, config.TransactionConfig.IndexByHash)
	assert.True(t, config.TransactionConfig.IndexByAddress)

	// Check storage config
	assert.Equal(t, "postgres", config.StorageConfig.Type)
	assert.Equal(t, 20, config.StorageConfig.MaxConnections)

	// Check cache config
	assert.Equal(t, "redis", config.CacheConfig.Type)
	assert.Equal(t, 1*time.Hour, config.CacheConfig.TTL)
}

func TestValidateDataIndexerConfig(t *testing.T) {
	// Test valid config
	validConfig := GetDefaultDataIndexerConfig()
	err := ValidateDataIndexerConfig(validConfig)
	assert.NoError(t, err)

	// Test disabled config
	disabledConfig := GetDefaultDataIndexerConfig()
	disabledConfig.Enabled = false
	err = ValidateDataIndexerConfig(disabledConfig)
	assert.NoError(t, err)

	// Test invalid configs
	invalidConfigs := []DataIndexerConfig{
		// Invalid indexing interval
		{
			Enabled:          true,
			IndexingInterval: 0,
		},
		// Invalid batch size
		{
			Enabled:          true,
			IndexingInterval: 5 * time.Second,
			BatchSize:        0,
		},
		// Invalid storage type
		{
			Enabled:          true,
			IndexingInterval: 5 * time.Second,
			BatchSize:        10,
			StorageConfig: StorageEngineConfig{
				Type:           "invalid_type",
				MaxConnections: 10,
			},
			CacheConfig: CacheEngineConfig{
				Type:    "redis",
				TTL:     1 * time.Hour,
				MaxSize: 1000,
			},
		},
	}

	for i, config := range invalidConfigs {
		err := ValidateDataIndexerConfig(config)
		assert.Error(t, err, "Config %d should be invalid", i)
	}
}

func TestConfigVariants(t *testing.T) {
	// Test high throughput config
	htConfig := GetHighThroughputConfig()
	assert.True(t, htConfig.IndexingInterval < GetDefaultDataIndexerConfig().IndexingInterval)
	assert.True(t, htConfig.BatchSize > GetDefaultDataIndexerConfig().BatchSize)

	// Test low latency config
	llConfig := GetLowLatencyConfig()
	assert.True(t, llConfig.IndexingInterval < GetDefaultDataIndexerConfig().IndexingInterval)
	assert.True(t, llConfig.AnalyticsConfig.RealTimeAnalytics)

	// Test archival config
	archivalConfig := GetArchivalConfig()
	assert.True(t, archivalConfig.BlockIndexerConfig.IndexUncles)
	assert.True(t, archivalConfig.EventConfig.IndexAllEvents)

	// Validate all configs
	assert.NoError(t, ValidateDataIndexerConfig(htConfig))
	assert.NoError(t, ValidateDataIndexerConfig(llConfig))
	assert.NoError(t, ValidateDataIndexerConfig(archivalConfig))
}

func TestUtilityFunctions(t *testing.T) {
	// Test supported storage types
	storageTypes := GetSupportedStorageTypes()
	assert.NotEmpty(t, storageTypes)
	assert.Contains(t, storageTypes, "postgres")
	assert.Contains(t, storageTypes, "mongodb")

	// Test supported cache types
	cacheTypes := GetSupportedCacheTypes()
	assert.NotEmpty(t, cacheTypes)
	assert.Contains(t, cacheTypes, "redis")
	assert.Contains(t, cacheTypes, "memcached")

	// Test supported contract types
	contractTypes := GetSupportedContractTypes()
	assert.NotEmpty(t, contractTypes)
	assert.Contains(t, contractTypes, "erc20")
	assert.Contains(t, contractTypes, "erc721")

	// Test optimal config for use case
	config, err := GetOptimalConfigForUseCase("high_throughput")
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Test invalid use case
	_, err = GetOptimalConfigForUseCase("invalid_use_case")
	assert.Error(t, err)

	// Test descriptions
	strategyDesc := GetIndexingStrategyDescription()
	assert.NotEmpty(t, strategyDesc)
	assert.Contains(t, strategyDesc, "high_throughput")

	storageDesc := GetStorageEngineDescription()
	assert.NotEmpty(t, storageDesc)
	assert.Contains(t, storageDesc, "postgres")

	cacheDesc := GetCacheEngineDescription()
	assert.NotEmpty(t, cacheDesc)
	assert.Contains(t, cacheDesc, "redis")
}
