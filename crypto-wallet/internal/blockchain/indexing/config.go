package indexing

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// GetDefaultDataIndexerConfig returns default data indexer configuration
func GetDefaultDataIndexerConfig() DataIndexerConfig {
	return DataIndexerConfig{
		Enabled:              true,
		IndexingInterval:     5 * time.Second,
		BatchSize:            10,
		MaxConcurrentWorkers: 4,
		StartBlock:           0,
		BlockIndexerConfig: BlockIndexerConfig{
			Enabled:            true,
			IndexHeaders:       true,
			IndexTransactions:  true,
			IndexReceipts:      true,
			IndexUncles:        false,
			CompressionEnabled: true,
			RetentionPeriod:    30 * 24 * time.Hour, // 30 days
		},
		TransactionConfig: TransactionIndexerConfig{
			Enabled:             true,
			IndexByHash:         true,
			IndexByAddress:      true,
			IndexByBlock:        true,
			IndexMethodCalls:    true,
			IndexTokenTransfers: true,
			FilterCriteria: []TransactionFilter{
				{
					Type:     "value_threshold",
					Criteria: map[string]interface{}{"min_value": "1000000000000000000"}, // 1 ETH
					Enabled:  true,
				},
			},
		},
		AddressConfig: AddressIndexerConfig{
			Enabled:           true,
			IndexBalances:     true,
			IndexTransactions: true,
			IndexTokenHoldings: true,
			IndexNFTs:         true,
			TrackingAddresses: []common.Address{},
			UpdateInterval:    1 * time.Minute,
		},
		ContractConfig: ContractIndexerConfig{
			Enabled:             true,
			IndexCreation:       true,
			IndexInteractions:   true,
			IndexEvents:         true,
			IndexABI:            true,
			ContractTypes:       []string{"erc20", "erc721", "erc1155", "uniswap", "aave"},
			VerificationEnabled: true,
		},
		EventConfig: EventIndexerConfig{
			Enabled:        true,
			IndexAllEvents: false,
			EventFilters: []EventFilter{
				{
					ContractAddress: common.Address{},
					EventName:       "Transfer",
					Enabled:         true,
				},
				{
					ContractAddress: common.Address{},
					EventName:       "Approval",
					Enabled:         true,
				},
			},
			DecodeEvents:    true,
			IndexEventData:  true,
		},
		StorageConfig: StorageEngineConfig{
			Type:                "postgres",
			ConnectionString:    "postgres://user:password@localhost/blockchain_data",
			MaxConnections:      20,
			CompressionEnabled:  true,
			PartitioningEnabled: true,
			BackupEnabled:       true,
		},
		CacheConfig: CacheEngineConfig{
			Type:             "redis",
			ConnectionString: "redis://localhost:6379",
			TTL:              1 * time.Hour,
			MaxSize:          10000,
			EvictionPolicy:   "lru",
		},
		DataProcessorConfig: DataProcessorConfig{
			Enabled: true,
			ProcessingPipelines: []string{
				"validation", "enrichment", "transformation", "analytics",
			},
			EnrichmentEnabled: true,
			ValidationEnabled: true,
			TransformationRules: []TransformationRule{
				{
					Name:    "normalize_addresses",
					Type:    "address_normalization",
					Source:  "raw_address",
					Target:  "normalized_address",
					Function: "toLowerCase",
					Enabled: true,
				},
			},
		},
		AnalyticsConfig: AnalyticsEngineConfig{
			Enabled:             true,
			RealTimeAnalytics:   true,
			HistoricalAnalytics: true,
			MetricsCollection: []string{
				"block_metrics", "transaction_metrics", "address_metrics", "contract_metrics",
			},
			AggregationInterval: 5 * time.Minute,
		},
	}
}

// GetHighThroughputConfig returns configuration optimized for high throughput
func GetHighThroughputConfig() DataIndexerConfig {
	config := GetDefaultDataIndexerConfig()
	
	// Optimize for high throughput
	config.IndexingInterval = 1 * time.Second
	config.BatchSize = 50
	config.MaxConcurrentWorkers = 16
	
	// Reduce storage overhead
	config.BlockIndexerConfig.IndexUncles = false
	config.BlockIndexerConfig.CompressionEnabled = true
	config.TransactionConfig.IndexMethodCalls = false
	config.EventConfig.IndexAllEvents = false
	
	// Optimize cache
	config.CacheConfig.MaxSize = 50000
	config.CacheConfig.TTL = 30 * time.Minute
	
	// Reduce analytics overhead
	config.AnalyticsConfig.RealTimeAnalytics = false
	config.AnalyticsConfig.AggregationInterval = 15 * time.Minute
	
	return config
}

// GetLowLatencyConfig returns configuration optimized for low latency
func GetLowLatencyConfig() DataIndexerConfig {
	config := GetDefaultDataIndexerConfig()
	
	// Optimize for low latency
	config.IndexingInterval = 500 * time.Millisecond
	config.BatchSize = 5
	config.MaxConcurrentWorkers = 8
	
	// Aggressive caching
	config.CacheConfig.MaxSize = 100000
	config.CacheConfig.TTL = 2 * time.Hour
	
	// Real-time analytics
	config.AnalyticsConfig.RealTimeAnalytics = true
	config.AnalyticsConfig.AggregationInterval = 1 * time.Minute
	
	// Minimal storage operations
	config.StorageConfig.CompressionEnabled = false
	config.DataProcessorConfig.ValidationEnabled = false
	
	return config
}

// GetArchivalConfig returns configuration optimized for archival/historical data
func GetArchivalConfig() DataIndexerConfig {
	config := GetDefaultDataIndexerConfig()
	
	// Optimize for completeness
	config.IndexingInterval = 10 * time.Second
	config.BatchSize = 100
	config.MaxConcurrentWorkers = 2
	
	// Index everything
	config.BlockIndexerConfig.IndexUncles = true
	config.TransactionConfig.IndexMethodCalls = true
	config.EventConfig.IndexAllEvents = true
	config.ContractConfig.VerificationEnabled = true
	
	// Long retention
	config.BlockIndexerConfig.RetentionPeriod = 365 * 24 * time.Hour // 1 year
	config.CacheConfig.TTL = 24 * time.Hour
	
	// Comprehensive analytics
	config.AnalyticsConfig.HistoricalAnalytics = true
	config.AnalyticsConfig.AggregationInterval = 1 * time.Hour
	
	// Storage optimization
	config.StorageConfig.CompressionEnabled = true
	config.StorageConfig.PartitioningEnabled = true
	config.StorageConfig.BackupEnabled = true
	
	return config
}

// ValidateDataIndexerConfig validates data indexer configuration
func ValidateDataIndexerConfig(config DataIndexerConfig) error {
	if !config.Enabled {
		return nil
	}
	
	if config.IndexingInterval <= 0 {
		return fmt.Errorf("indexing interval must be positive")
	}
	
	if config.BatchSize <= 0 {
		return fmt.Errorf("batch size must be positive")
	}
	
	if config.MaxConcurrentWorkers <= 0 {
		return fmt.Errorf("max concurrent workers must be positive")
	}
	
	// Validate block indexer config
	if config.BlockIndexerConfig.Enabled {
		if config.BlockIndexerConfig.RetentionPeriod <= 0 {
			return fmt.Errorf("block retention period must be positive")
		}
	}
	
	// Validate transaction config
	if config.TransactionConfig.Enabled {
		if !config.TransactionConfig.IndexByHash && 
		   !config.TransactionConfig.IndexByAddress && 
		   !config.TransactionConfig.IndexByBlock {
			return fmt.Errorf("at least one transaction indexing method must be enabled")
		}
	}
	
	// Validate address config
	if config.AddressConfig.Enabled {
		if config.AddressConfig.UpdateInterval <= 0 {
			return fmt.Errorf("address update interval must be positive")
		}
	}
	
	// Validate storage config
	if config.StorageConfig.Type == "" {
		return fmt.Errorf("storage type must be specified")
	}
	
	validStorageTypes := []string{"postgres", "mongodb", "elasticsearch", "mysql"}
	isValidStorage := false
	for _, validType := range validStorageTypes {
		if config.StorageConfig.Type == validType {
			isValidStorage = true
			break
		}
	}
	if !isValidStorage {
		return fmt.Errorf("invalid storage type: %s", config.StorageConfig.Type)
	}
	
	if config.StorageConfig.MaxConnections <= 0 {
		return fmt.Errorf("max connections must be positive")
	}
	
	// Validate cache config
	if config.CacheConfig.Type == "" {
		return fmt.Errorf("cache type must be specified")
	}
	
	validCacheTypes := []string{"redis", "memcached", "in-memory"}
	isValidCache := false
	for _, validType := range validCacheTypes {
		if config.CacheConfig.Type == validType {
			isValidCache = true
			break
		}
	}
	if !isValidCache {
		return fmt.Errorf("invalid cache type: %s", config.CacheConfig.Type)
	}
	
	if config.CacheConfig.TTL <= 0 {
		return fmt.Errorf("cache TTL must be positive")
	}
	
	if config.CacheConfig.MaxSize <= 0 {
		return fmt.Errorf("cache max size must be positive")
	}
	
	// Validate analytics config
	if config.AnalyticsConfig.Enabled {
		if config.AnalyticsConfig.AggregationInterval <= 0 {
			return fmt.Errorf("analytics aggregation interval must be positive")
		}
		
		if len(config.AnalyticsConfig.MetricsCollection) == 0 {
			return fmt.Errorf("at least one metrics collection must be specified")
		}
	}
	
	return nil
}

// GetSupportedStorageTypes returns supported storage engine types
func GetSupportedStorageTypes() []string {
	return []string{
		"postgres",
		"mongodb", 
		"elasticsearch",
		"mysql",
		"sqlite",
		"cassandra",
	}
}

// GetSupportedCacheTypes returns supported cache engine types
func GetSupportedCacheTypes() []string {
	return []string{
		"redis",
		"memcached",
		"in-memory",
		"hazelcast",
	}
}

// GetSupportedContractTypes returns supported contract types
func GetSupportedContractTypes() []string {
	return []string{
		"erc20",
		"erc721",
		"erc1155",
		"uniswap",
		"aave",
		"compound",
		"makerdao",
		"chainlink",
		"multisig",
		"proxy",
	}
}

// GetSupportedEventNames returns commonly indexed event names
func GetSupportedEventNames() []string {
	return []string{
		"Transfer",
		"Approval",
		"ApprovalForAll",
		"Deposit",
		"Withdrawal",
		"Swap",
		"Mint",
		"Burn",
		"Sync",
		"PairCreated",
	}
}

// GetSupportedProcessingPipelines returns supported processing pipelines
func GetSupportedProcessingPipelines() []string {
	return []string{
		"validation",
		"enrichment",
		"transformation",
		"analytics",
		"deduplication",
		"normalization",
		"aggregation",
	}
}

// GetSupportedMetricsCollections returns supported metrics collections
func GetSupportedMetricsCollections() []string {
	return []string{
		"block_metrics",
		"transaction_metrics",
		"address_metrics",
		"contract_metrics",
		"event_metrics",
		"gas_metrics",
		"volume_metrics",
		"performance_metrics",
	}
}

// GetOptimalConfigForUseCase returns optimal configuration for specific use cases
func GetOptimalConfigForUseCase(useCase string) (DataIndexerConfig, error) {
	switch useCase {
	case "high_throughput":
		return GetHighThroughputConfig(), nil
	case "low_latency":
		return GetLowLatencyConfig(), nil
	case "archival":
		return GetArchivalConfig(), nil
	case "default":
		return GetDefaultDataIndexerConfig(), nil
	default:
		return DataIndexerConfig{}, fmt.Errorf("unsupported use case: %s", useCase)
	}
}

// GetIndexingStrategyDescription returns descriptions for indexing strategies
func GetIndexingStrategyDescription() map[string]string {
	return map[string]string{
		"high_throughput": "Optimized for maximum indexing speed with reduced storage overhead",
		"low_latency":     "Optimized for minimal query response time with aggressive caching",
		"archival":        "Optimized for complete data retention and historical analysis",
		"default":         "Balanced configuration suitable for most use cases",
	}
}

// GetStorageEngineDescription returns descriptions for storage engines
func GetStorageEngineDescription() map[string]string {
	return map[string]string{
		"postgres":      "Relational database with excellent ACID properties and complex queries",
		"mongodb":       "Document database with flexible schema and horizontal scaling",
		"elasticsearch": "Search engine with powerful full-text search and analytics capabilities",
		"mysql":         "Popular relational database with good performance and reliability",
		"sqlite":        "Lightweight embedded database suitable for development and testing",
		"cassandra":     "Distributed NoSQL database optimized for high write throughput",
	}
}

// GetCacheEngineDescription returns descriptions for cache engines
func GetCacheEngineDescription() map[string]string {
	return map[string]string{
		"redis":      "In-memory data structure store with persistence and clustering support",
		"memcached":  "High-performance distributed memory caching system",
		"in-memory":  "Simple in-process cache suitable for single-instance deployments",
		"hazelcast":  "Distributed in-memory computing platform with advanced features",
	}
}
