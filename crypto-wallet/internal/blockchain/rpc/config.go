package rpc

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// GetDefaultNodeManagerConfig returns default node manager configuration
func GetDefaultNodeManagerConfig() NodeManagerConfig {
	return NodeManagerConfig{
		Enabled:               true,
		HealthCheckInterval:   30 * time.Second,
		FailoverTimeout:       5 * time.Second,
		MaxRetries:            3,
		LoadBalancingStrategy: "round_robin",
		Nodes: []NodeConfig{
			{
				ID:          "infura_mainnet",
				Name:        "Infura Mainnet",
				URL:         "https://mainnet.infura.io/v3/YOUR_PROJECT_ID",
				Chain:       "ethereum",
				Provider:    "infura",
				Priority:    1,
				Weight:      decimal.NewFromFloat(1.0),
				MaxRequests: 1000,
				Timeout:     10 * time.Second,
				Enabled:     true,
			},
			{
				ID:          "alchemy_mainnet",
				Name:        "Alchemy Mainnet",
				URL:         "https://eth-mainnet.alchemyapi.io/v2/YOUR_API_KEY",
				Chain:       "ethereum",
				Provider:    "alchemy",
				Priority:    2,
				Weight:      decimal.NewFromFloat(1.0),
				MaxRequests: 1000,
				Timeout:     10 * time.Second,
				Enabled:     true,
			},
			{
				ID:          "quicknode_mainnet",
				Name:        "QuickNode Mainnet",
				URL:         "https://YOUR_ENDPOINT.quiknode.pro/YOUR_API_KEY/",
				Chain:       "ethereum",
				Provider:    "quicknode",
				Priority:    3,
				Weight:      decimal.NewFromFloat(0.8),
				MaxRequests: 800,
				Timeout:     10 * time.Second,
				Enabled:     true,
			},
		},
		HealthCheckConfig: HealthCheckConfig{
			Enabled:            true,
			Interval:           30 * time.Second,
			Timeout:            5 * time.Second,
			MaxLatency:         2 * time.Second,
			MinSuccessRate:     decimal.NewFromFloat(0.95),
			CheckMethods:       []string{"chain_id", "block_number", "eth_syncing"},
			UnhealthyThreshold: 3,
			HealthyThreshold:   2,
		},
		LoadBalancerConfig: LoadBalancerConfig{
			Strategy:           "round_robin",
			WeightedRoundRobin: false,
			StickySession:      false,
			SessionTimeout:     30 * time.Minute,
			MaxRequestsPerNode: 1000,
		},
		MetricsConfig: MetricsConfig{
			Enabled:            true,
			CollectionInterval: 1 * time.Minute,
			RetentionPeriod:    24 * time.Hour,
			ExportEnabled:      false,
		},
	}
}

// GetHighAvailabilityConfig returns high availability configuration
func GetHighAvailabilityConfig() NodeManagerConfig {
	config := GetDefaultNodeManagerConfig()
	
	// More aggressive health checking
	config.HealthCheckInterval = 15 * time.Second
	config.HealthCheckConfig.Interval = 15 * time.Second
	config.HealthCheckConfig.Timeout = 3 * time.Second
	config.HealthCheckConfig.MaxLatency = 1 * time.Second
	config.HealthCheckConfig.MinSuccessRate = decimal.NewFromFloat(0.98)
	config.HealthCheckConfig.UnhealthyThreshold = 2
	
	// Faster failover
	config.FailoverTimeout = 2 * time.Second
	config.MaxRetries = 5
	
	// Load balancing with sticky sessions
	config.LoadBalancingStrategy = "lowest_latency"
	config.LoadBalancerConfig.Strategy = "lowest_latency"
	config.LoadBalancerConfig.StickySession = true
	
	// More frequent metrics collection
	config.MetricsConfig.CollectionInterval = 30 * time.Second
	
	return config
}

// GetDevelopmentConfig returns development configuration
func GetDevelopmentConfig() NodeManagerConfig {
	config := GetDefaultNodeManagerConfig()
	
	// Less aggressive health checking for development
	config.HealthCheckInterval = 1 * time.Minute
	config.HealthCheckConfig.Interval = 1 * time.Minute
	config.HealthCheckConfig.Timeout = 10 * time.Second
	config.HealthCheckConfig.MaxLatency = 5 * time.Second
	config.HealthCheckConfig.MinSuccessRate = decimal.NewFromFloat(0.8)
	config.HealthCheckConfig.UnhealthyThreshold = 5
	
	// Simple round-robin
	config.LoadBalancingStrategy = "round_robin"
	config.LoadBalancerConfig.Strategy = "round_robin"
	
	// Less frequent metrics collection
	config.MetricsConfig.CollectionInterval = 5 * time.Minute
	
	// Only use one node for development
	config.Nodes = []NodeConfig{
		{
			ID:          "local_ganache",
			Name:        "Local Ganache",
			URL:         "http://localhost:8545",
			Chain:       "ethereum",
			Provider:    "ganache",
			Priority:    1,
			Weight:      decimal.NewFromFloat(1.0),
			MaxRequests: 10000,
			Timeout:     30 * time.Second,
			Enabled:     true,
		},
	}
	
	return config
}

// GetMultiChainConfig returns multi-chain configuration
func GetMultiChainConfig() NodeManagerConfig {
	config := GetDefaultNodeManagerConfig()
	
	// Add nodes for multiple chains
	config.Nodes = []NodeConfig{
		// Ethereum Mainnet
		{
			ID:          "infura_ethereum",
			Name:        "Infura Ethereum",
			URL:         "https://mainnet.infura.io/v3/YOUR_PROJECT_ID",
			Chain:       "ethereum",
			Provider:    "infura",
			Priority:    1,
			Weight:      decimal.NewFromFloat(1.0),
			MaxRequests: 1000,
			Timeout:     10 * time.Second,
			Enabled:     true,
		},
		{
			ID:          "alchemy_ethereum",
			Name:        "Alchemy Ethereum",
			URL:         "https://eth-mainnet.alchemyapi.io/v2/YOUR_API_KEY",
			Chain:       "ethereum",
			Provider:    "alchemy",
			Priority:    2,
			Weight:      decimal.NewFromFloat(1.0),
			MaxRequests: 1000,
			Timeout:     10 * time.Second,
			Enabled:     true,
		},
		// Polygon
		{
			ID:          "infura_polygon",
			Name:        "Infura Polygon",
			URL:         "https://polygon-mainnet.infura.io/v3/YOUR_PROJECT_ID",
			Chain:       "polygon",
			Provider:    "infura",
			Priority:    1,
			Weight:      decimal.NewFromFloat(1.0),
			MaxRequests: 1000,
			Timeout:     10 * time.Second,
			Enabled:     true,
		},
		{
			ID:          "alchemy_polygon",
			Name:        "Alchemy Polygon",
			URL:         "https://polygon-mainnet.g.alchemy.com/v2/YOUR_API_KEY",
			Chain:       "polygon",
			Provider:    "alchemy",
			Priority:    2,
			Weight:      decimal.NewFromFloat(1.0),
			MaxRequests: 1000,
			Timeout:     10 * time.Second,
			Enabled:     true,
		},
		// Arbitrum
		{
			ID:          "infura_arbitrum",
			Name:        "Infura Arbitrum",
			URL:         "https://arbitrum-mainnet.infura.io/v3/YOUR_PROJECT_ID",
			Chain:       "arbitrum",
			Provider:    "infura",
			Priority:    1,
			Weight:      decimal.NewFromFloat(1.0),
			MaxRequests: 1000,
			Timeout:     10 * time.Second,
			Enabled:     true,
		},
		{
			ID:          "alchemy_arbitrum",
			Name:        "Alchemy Arbitrum",
			URL:         "https://arb-mainnet.g.alchemy.com/v2/YOUR_API_KEY",
			Chain:       "arbitrum",
			Provider:    "alchemy",
			Priority:    2,
			Weight:      decimal.NewFromFloat(1.0),
			MaxRequests: 1000,
			Timeout:     10 * time.Second,
			Enabled:     true,
		},
		// Optimism
		{
			ID:          "infura_optimism",
			Name:        "Infura Optimism",
			URL:         "https://optimism-mainnet.infura.io/v3/YOUR_PROJECT_ID",
			Chain:       "optimism",
			Provider:    "infura",
			Priority:    1,
			Weight:      decimal.NewFromFloat(1.0),
			MaxRequests: 1000,
			Timeout:     10 * time.Second,
			Enabled:     true,
		},
		{
			ID:          "alchemy_optimism",
			Name:        "Alchemy Optimism",
			URL:         "https://opt-mainnet.g.alchemy.com/v2/YOUR_API_KEY",
			Chain:       "optimism",
			Provider:    "alchemy",
			Priority:    2,
			Weight:      decimal.NewFromFloat(1.0),
			MaxRequests: 1000,
			Timeout:     10 * time.Second,
			Enabled:     true,
		},
	}
	
	return config
}

// ValidateNodeManagerConfig validates node manager configuration
func ValidateNodeManagerConfig(config NodeManagerConfig) error {
	if !config.Enabled {
		return nil
	}
	
	if config.HealthCheckInterval <= 0 {
		return fmt.Errorf("health check interval must be positive")
	}
	
	if config.FailoverTimeout <= 0 {
		return fmt.Errorf("failover timeout must be positive")
	}
	
	if config.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}
	
	if len(config.Nodes) == 0 {
		return fmt.Errorf("at least one node must be configured")
	}
	
	// Validate individual nodes
	nodeIDs := make(map[string]bool)
	enabledNodes := 0
	
	for i, node := range config.Nodes {
		if err := validateNodeConfig(node); err != nil {
			return fmt.Errorf("node %d validation failed: %w", i, err)
		}
		
		if nodeIDs[node.ID] {
			return fmt.Errorf("duplicate node ID: %s", node.ID)
		}
		nodeIDs[node.ID] = true
		
		if node.Enabled {
			enabledNodes++
		}
	}
	
	if enabledNodes == 0 {
		return fmt.Errorf("at least one node must be enabled")
	}
	
	// Validate health check config
	if config.HealthCheckConfig.Enabled {
		if config.HealthCheckConfig.Interval <= 0 {
			return fmt.Errorf("health check interval must be positive")
		}
		if config.HealthCheckConfig.Timeout <= 0 {
			return fmt.Errorf("health check timeout must be positive")
		}
		if config.HealthCheckConfig.MaxLatency <= 0 {
			return fmt.Errorf("max latency must be positive")
		}
		if config.HealthCheckConfig.MinSuccessRate.LessThan(decimal.Zero) || 
		   config.HealthCheckConfig.MinSuccessRate.GreaterThan(decimal.NewFromFloat(1)) {
			return fmt.Errorf("min success rate must be between 0 and 1")
		}
		if len(config.HealthCheckConfig.CheckMethods) == 0 {
			return fmt.Errorf("at least one health check method must be specified")
		}
	}
	
	// Validate load balancer config
	validStrategies := []string{"round_robin", "weighted_round_robin", "least_connections", "lowest_latency"}
	isValidStrategy := false
	for _, strategy := range validStrategies {
		if config.LoadBalancerConfig.Strategy == strategy {
			isValidStrategy = true
			break
		}
	}
	if !isValidStrategy {
		return fmt.Errorf("invalid load balancing strategy: %s", config.LoadBalancerConfig.Strategy)
	}
	
	return nil
}

// validateNodeConfig validates individual node configuration
func validateNodeConfig(config NodeConfig) error {
	if config.ID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	
	if config.URL == "" {
		return fmt.Errorf("node URL cannot be empty")
	}
	
	if config.Chain == "" {
		return fmt.Errorf("node chain cannot be empty")
	}
	
	if config.Provider == "" {
		return fmt.Errorf("node provider cannot be empty")
	}
	
	if config.Priority < 0 {
		return fmt.Errorf("node priority cannot be negative")
	}
	
	if config.Weight.LessThan(decimal.Zero) {
		return fmt.Errorf("node weight cannot be negative")
	}
	
	if config.MaxRequests < 0 {
		return fmt.Errorf("max requests cannot be negative")
	}
	
	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	
	return nil
}

// GetSupportedProviders returns list of supported RPC providers
func GetSupportedProviders() []string {
	return []string{
		"infura",
		"alchemy",
		"quicknode",
		"moralis",
		"ankr",
		"chainstack",
		"getblock",
		"nodereal",
		"pokt",
		"local",
		"ganache",
		"hardhat",
	}
}

// GetSupportedChains returns list of supported blockchain networks
func GetSupportedChains() []string {
	return []string{
		"ethereum",
		"polygon",
		"arbitrum",
		"optimism",
		"avalanche",
		"bsc",
		"fantom",
		"gnosis",
		"celo",
		"moonbeam",
		"aurora",
		"harmony",
	}
}

// GetLoadBalancingStrategies returns available load balancing strategies
func GetLoadBalancingStrategies() []string {
	return []string{
		"round_robin",
		"weighted_round_robin",
		"least_connections",
		"lowest_latency",
	}
}

// GetHealthCheckMethods returns available health check methods
func GetHealthCheckMethods() []string {
	return []string{
		"chain_id",
		"block_number",
		"net_peer_count",
		"eth_syncing",
		"eth_gas_price",
		"net_version",
	}
}
