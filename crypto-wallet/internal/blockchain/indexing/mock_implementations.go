package indexing

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
)

// Mock implementations for testing and demonstration

// MockBlockIndexer provides mock block indexing
type MockBlockIndexer struct{}

func (m *MockBlockIndexer) IndexBlock(ctx context.Context, block *types.Block) (*IndexedBlock, error) {
	return &IndexedBlock{
		Number:           block.NumberU64(),
		Hash:             block.Hash(),
		ParentHash:       block.ParentHash(),
		Timestamp:        time.Unix(int64(block.Time()), 0),
		Miner:            block.Coinbase(),
		Difficulty:       decimal.NewFromBigInt(block.Difficulty(), 0),
		TotalDifficulty:  decimal.NewFromBigInt(block.Difficulty(), 0), // Mock total difficulty
		GasLimit:         block.GasLimit(),
		GasUsed:          block.GasUsed(),
		TransactionCount: len(block.Transactions()),
		Size:             block.Size(),
		ExtraData:        block.Extra(),
		IndexedAt:        time.Now(),
	}, nil
}

func (m *MockBlockIndexer) GetBlock(blockNumber uint64) (*IndexedBlock, error) {
	// Mock block retrieval
	return &IndexedBlock{
		Number:           blockNumber,
		Hash:             common.HexToHash(fmt.Sprintf("0x%064x", blockNumber)),
		Timestamp:        time.Now(),
		GasLimit:         8000000,
		GasUsed:          4000000,
		TransactionCount: 5,
		IndexedAt:        time.Now(),
	}, nil
}

func (m *MockBlockIndexer) GetBlockByHash(hash common.Hash) (*IndexedBlock, error) {
	// Mock block retrieval by hash
	return &IndexedBlock{
		Number:           12345,
		Hash:             hash,
		Timestamp:        time.Now(),
		GasLimit:         8000000,
		GasUsed:          4000000,
		TransactionCount: 5,
		IndexedAt:        time.Now(),
	}, nil
}

// MockTransactionIndexer provides mock transaction indexing
type MockTransactionIndexer struct{}

func (m *MockTransactionIndexer) IndexTransaction(ctx context.Context, tx *types.Transaction, receipt *types.Receipt) (*IndexedTransaction, error) {
	// Extract sender address (mock implementation)
	from := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Mock method ID extraction
	var methodID []byte
	if len(tx.Data()) >= 4 {
		methodID = tx.Data()[:4]
	}

	return &IndexedTransaction{
		Hash:             tx.Hash(),
		BlockNumber:      receipt.BlockNumber.Uint64(),
		BlockHash:        receipt.BlockHash,
		TransactionIndex: receipt.TransactionIndex,
		From:             from,
		To:               tx.To(),
		Value:            decimal.NewFromBigInt(tx.Value(), 0),
		GasLimit:         tx.Gas(),
		GasPrice:         decimal.NewFromBigInt(tx.GasPrice(), 0),
		GasUsed:          receipt.GasUsed,
		Status:           receipt.Status,
		Nonce:            tx.Nonce(),
		Data:             tx.Data(),
		MethodID:         methodID,
		DecodedInput:     make(map[string]interface{}),
		TokenTransfers:   []TokenTransfer{},
		Events:           []IndexedEvent{},
		IndexedAt:        time.Now(),
	}, nil
}

func (m *MockTransactionIndexer) GetTransaction(hash common.Hash) (*IndexedTransaction, error) {
	// Mock transaction retrieval
	return &IndexedTransaction{
		Hash:        hash,
		BlockNumber: 12345,
		From:        common.HexToAddress("0x1234567890123456789012345678901234567890"),
		To:          &common.Address{},
		Value:       decimal.NewFromFloat(1000000000000000000), // 1 ETH
		GasLimit:    21000,
		GasPrice:    decimal.NewFromFloat(20000000000), // 20 gwei
		GasUsed:     21000,
		Status:      1,
		IndexedAt:   time.Now(),
	}, nil
}

func (m *MockTransactionIndexer) GetTransactionsByAddress(address common.Address, limit int) ([]*IndexedTransaction, error) {
	// Mock transaction list for address
	var transactions []*IndexedTransaction
	for i := 0; i < limit && i < 5; i++ {
		tx := &IndexedTransaction{
			Hash:        common.HexToHash(fmt.Sprintf("0x%064x", i)),
			BlockNumber: uint64(12345 + i),
			From:        address,
			Value:       decimal.NewFromFloat(1000000000000000000), // 1 ETH
			GasLimit:    21000,
			GasPrice:    decimal.NewFromFloat(20000000000), // 20 gwei
			IndexedAt:   time.Now(),
		}
		transactions = append(transactions, tx)
	}
	return transactions, nil
}

// MockAddressIndexer provides mock address indexing
type MockAddressIndexer struct{}

func (m *MockAddressIndexer) IndexAddress(ctx context.Context, address common.Address) (*IndexedAddress, error) {
	return &IndexedAddress{
		Address:          address,
		Type:             "eoa",                                     // Assume EOA for mock
		Balance:          decimal.NewFromFloat(5000000000000000000), // 5 ETH
		Nonce:            10,
		TransactionCount: 25,
		FirstSeen:        time.Now().Add(-30 * 24 * time.Hour),
		LastActivity:     time.Now().Add(-1 * time.Hour),
		TokenHoldings:    []TokenHolding{},
		NFTHoldings:      []NFTHolding{},
		Tags:             []string{},
		IndexedAt:        time.Now(),
	}, nil
}

func (m *MockAddressIndexer) GetAddress(address common.Address) (*IndexedAddress, error) {
	return m.IndexAddress(context.Background(), address)
}

func (m *MockAddressIndexer) UpdateAddressBalance(address common.Address, balance decimal.Decimal) error {
	// Mock balance update
	return nil
}

// MockContractIndexer provides mock contract indexing
type MockContractIndexer struct{}

func (m *MockContractIndexer) IndexContract(ctx context.Context, address common.Address) (*IndexedContract, error) {
	return &IndexedContract{
		Address:          address,
		Creator:          common.HexToAddress("0x1234567890123456789012345678901234567890"),
		CreationTxHash:   common.HexToHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"),
		CreationBlock:    12000,
		ContractType:     "erc20",
		Name:             "Mock Token",
		Symbol:           "MOCK",
		ABI:              []byte("[]"), // Empty ABI for mock
		SourceCode:       "",
		CompilerVersion:  "0.8.0",
		IsVerified:       false,
		InteractionCount: 100,
		IndexedAt:        time.Now(),
	}, nil
}

func (m *MockContractIndexer) GetContract(address common.Address) (*IndexedContract, error) {
	return m.IndexContract(context.Background(), address)
}

func (m *MockContractIndexer) VerifyContract(address common.Address, sourceCode string) error {
	// Mock contract verification
	return nil
}

// MockEventIndexer provides mock event indexing
type MockEventIndexer struct{}

func (m *MockEventIndexer) IndexEvent(ctx context.Context, log *types.Log) (*IndexedEvent, error) {
	return &IndexedEvent{
		ID:               fmt.Sprintf("%s-%d", log.TxHash.Hex(), log.Index),
		BlockNumber:      log.BlockNumber,
		BlockHash:        log.BlockHash,
		TransactionHash:  log.TxHash,
		TransactionIndex: log.TxIndex,
		LogIndex:         log.Index,
		Address:          log.Address,
		Topics:           log.Topics,
		Data:             log.Data,
		EventName:        "Transfer", // Mock event name
		DecodedData:      make(map[string]interface{}),
		IndexedAt:        time.Now(),
	}, nil
}

func (m *MockEventIndexer) GetEvents(filter EventFilter, limit int) ([]*IndexedEvent, error) {
	// Mock event list
	var events []*IndexedEvent
	for i := 0; i < limit && i < 3; i++ {
		event := &IndexedEvent{
			ID:          fmt.Sprintf("event_%d", i),
			BlockNumber: uint64(12345 + i),
			Address:     filter.ContractAddress,
			EventName:   filter.EventName,
			DecodedData: make(map[string]interface{}),
			IndexedAt:   time.Now(),
		}
		events = append(events, event)
	}
	return events, nil
}

func (m *MockEventIndexer) GetEventsByContract(address common.Address, limit int) ([]*IndexedEvent, error) {
	filter := EventFilter{
		ContractAddress: address,
		EventName:       "Transfer",
	}
	return m.GetEvents(filter, limit)
}

// MockStorageEngine provides mock storage
type MockStorageEngine struct{}

func (m *MockStorageEngine) Store(ctx context.Context, data interface{}) error {
	// Mock storage operation
	return nil
}

func (m *MockStorageEngine) Retrieve(ctx context.Context, key string, result interface{}) error {
	// Mock retrieval operation
	return fmt.Errorf("not found")
}

func (m *MockStorageEngine) Query(ctx context.Context, query interface{}) ([]interface{}, error) {
	// Mock query operation
	return []interface{}{}, nil
}

func (m *MockStorageEngine) Delete(ctx context.Context, key string) error {
	// Mock deletion operation
	return nil
}

// MockCacheEngine provides mock caching
type MockCacheEngine struct{}

func (m *MockCacheEngine) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Mock cache set operation
	return nil
}

func (m *MockCacheEngine) Get(ctx context.Context, key string, result interface{}) error {
	// Mock cache get operation - always return not found for simplicity
	return fmt.Errorf("not found in cache")
}

func (m *MockCacheEngine) Delete(ctx context.Context, key string) error {
	// Mock cache delete operation
	return nil
}

func (m *MockCacheEngine) Clear(ctx context.Context) error {
	// Mock cache clear operation
	return nil
}

// MockDataProcessor provides mock data processing
type MockDataProcessor struct{}

func (m *MockDataProcessor) Process(ctx context.Context, data interface{}) (interface{}, error) {
	// Mock data processing
	return data, nil
}

func (m *MockDataProcessor) Enrich(ctx context.Context, data interface{}) (interface{}, error) {
	// Mock data enrichment
	return data, nil
}

func (m *MockDataProcessor) Validate(ctx context.Context, data interface{}) error {
	// Mock data validation
	return nil
}

// MockAnalyticsEngine provides mock analytics
type MockAnalyticsEngine struct{}

func (m *MockAnalyticsEngine) Analyze(ctx context.Context, data interface{}) (*AnalyticsResult, error) {
	return &AnalyticsResult{
		Metrics: map[string]decimal.Decimal{
			"transaction_count": decimal.NewFromFloat(1000),
			"average_gas_price": decimal.NewFromFloat(20000000000),
			"total_volume":      decimal.NewFromFloat(100000000000000000000), // 100 ETH
		},
		Insights: []string{
			"Transaction volume increased by 15%",
			"Average gas price decreased by 5%",
			"New addresses increased by 8%",
		},
		Timestamp: time.Now(),
	}, nil
}

func (m *MockAnalyticsEngine) GetMetrics(ctx context.Context, timeRange TimeRange) (*MetricsResult, error) {
	return &MetricsResult{
		BlockMetrics: &BlockMetrics{
			TotalBlocks:      1000,
			AverageBlockTime: decimal.NewFromFloat(13.5),
			AverageGasUsed:   decimal.NewFromFloat(4000000),
			TotalGasUsed:     4000000000,
		},
		TransactionMetrics: &TransactionMetrics{
			TotalTransactions: 5000,
			AverageGasPrice:   decimal.NewFromFloat(20000000000),
			AverageValue:      decimal.NewFromFloat(1000000000000000000),
			SuccessRate:       decimal.NewFromFloat(0.98),
		},
		AddressMetrics: &AddressMetrics{
			TotalAddresses:  10000,
			ActiveAddresses: 2500,
			NewAddresses:    150,
		},
		ContractMetrics: &ContractMetrics{
			TotalContracts:    500,
			NewContracts:      25,
			VerifiedContracts: 300,
			PopularContracts: []PopularContract{
				{
					Address:          common.HexToAddress("0x1234567890123456789012345678901234567890"),
					Name:             "Popular Contract",
					InteractionCount: 1000,
				},
			},
		},
		TimeRange: timeRange,
	}, nil
}

func (m *MockAnalyticsEngine) GenerateReport(ctx context.Context, reportType string) (*ReportResult, error) {
	return &ReportResult{
		ReportType: reportType,
		Data: map[string]interface{}{
			"summary":     "Mock report data",
			"total_items": 100,
			"period":      "24h",
		},
		GeneratedAt: time.Now(),
	}, nil
}
