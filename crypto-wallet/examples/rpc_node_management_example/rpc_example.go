package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/blockchain/rpc"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("🌐 RPC Node Management System Example")
	fmt.Println("=====================================")

	// Initialize logger
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	logger := logger.NewLogger(logConfig)

	// Create node manager configuration
	config := createExampleConfig()

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Enabled: %v\n", config.Enabled)
	fmt.Printf("  Health Check Interval: %v\n", config.HealthCheckInterval)
	fmt.Printf("  Failover Timeout: %v\n", config.FailoverTimeout)
	fmt.Printf("  Max Retries: %d\n", config.MaxRetries)
	fmt.Printf("  Load Balancing Strategy: %s\n", config.LoadBalancingStrategy)
	fmt.Printf("  Total Nodes: %d\n", len(config.Nodes))
	fmt.Printf("  Health Checks: %v\n", config.HealthCheckConfig.Enabled)
	fmt.Printf("  Metrics Collection: %v\n", config.MetricsConfig.Enabled)
	fmt.Println()

	// Display configured nodes
	fmt.Println("📡 Configured RPC Nodes:")
	fmt.Println("========================")
	for i, node := range config.Nodes {
		fmt.Printf("%d. %s (%s)\n", i+1, node.Name, node.ID)
		fmt.Printf("   Provider: %s\n", node.Provider)
		fmt.Printf("   Chain: %s\n", node.Chain)
		fmt.Printf("   URL: %s\n", maskURL(node.URL))
		fmt.Printf("   Priority: %d\n", node.Priority)
		fmt.Printf("   Weight: %s\n", node.Weight.String())
		fmt.Printf("   Max Requests: %d\n", node.MaxRequests)
		fmt.Printf("   Timeout: %v\n", node.Timeout)
		fmt.Printf("   Enabled: %v\n", node.Enabled)
		fmt.Println()
	}

	// Create node manager
	nodeManager := rpc.NewNodeManager(logger, config)

	// Note: In this example, we'll demonstrate the system without actual RPC connections
	// since we don't have real RPC endpoints configured
	fmt.Println("⚠️  Note: This example demonstrates the RPC management system")
	fmt.Println("   without actual RPC connections. Configure real endpoints for production use.")
	fmt.Println()

	// Show node manager metrics before starting
	fmt.Println("📊 Node Manager Metrics (Before Start):")
	fmt.Println("=======================================")
	metrics := nodeManager.GetMetrics()
	displayMetrics(metrics)

	// Demonstrate configuration validation
	fmt.Println("🔍 Configuration Validation:")
	fmt.Println("============================")
	if err := rpc.ValidateNodeManagerConfig(config); err != nil {
		fmt.Printf("❌ Configuration validation failed: %v\n", err)
	} else {
		fmt.Printf("✅ Configuration validation passed\n")
	}
	fmt.Println()

	// Show different configuration profiles
	fmt.Println("🔧 Configuration Profiles:")
	fmt.Println("==========================")

	// Default configuration
	defaultConfig := rpc.GetDefaultNodeManagerConfig()
	fmt.Printf("📊 Default Configuration:\n")
	fmt.Printf("  Health Check Interval: %v\n", defaultConfig.HealthCheckInterval)
	fmt.Printf("  Load Balancing: %s\n", defaultConfig.LoadBalancingStrategy)
	fmt.Printf("  Nodes: %d\n", len(defaultConfig.Nodes))
	fmt.Println()

	// High availability configuration
	haConfig := rpc.GetHighAvailabilityConfig()
	fmt.Printf("🛡️  High Availability Configuration:\n")
	fmt.Printf("  Health Check Interval: %v (more frequent)\n", haConfig.HealthCheckInterval)
	fmt.Printf("  Failover Timeout: %v (faster)\n", haConfig.FailoverTimeout)
	fmt.Printf("  Load Balancing: %s (latency-based)\n", haConfig.LoadBalancingStrategy)
	fmt.Printf("  Sticky Sessions: %v\n", haConfig.LoadBalancerConfig.StickySession)
	fmt.Println()

	// Development configuration
	devConfig := rpc.GetDevelopmentConfig()
	fmt.Printf("🔧 Development Configuration:\n")
	fmt.Printf("  Health Check Interval: %v (less frequent)\n", devConfig.HealthCheckInterval)
	fmt.Printf("  Load Balancing: %s (simple)\n", devConfig.LoadBalancingStrategy)
	fmt.Printf("  Nodes: %d (local only)\n", len(devConfig.Nodes))
	fmt.Println()

	// Multi-chain configuration
	multiChainConfig := rpc.GetMultiChainConfig()
	fmt.Printf("🌐 Multi-Chain Configuration:\n")
	fmt.Printf("  Total Nodes: %d\n", len(multiChainConfig.Nodes))

	// Count nodes by chain
	chainCounts := make(map[string]int)
	for _, node := range multiChainConfig.Nodes {
		chainCounts[node.Chain]++
	}
	fmt.Printf("  Supported Chains:\n")
	for chain, count := range chainCounts {
		fmt.Printf("    %s: %d nodes\n", chain, count)
	}
	fmt.Println()

	// Demonstrate load balancing strategies
	fmt.Println("⚖️  Load Balancing Strategies:")
	fmt.Println("=============================")
	strategies := rpc.GetLoadBalancingStrategies()
	for i, strategy := range strategies {
		description := getStrategyDescription(strategy)
		fmt.Printf("%d. %s: %s\n", i+1, strategy, description)
	}
	fmt.Println()

	// Show supported providers and chains
	fmt.Println("🔌 Supported Providers:")
	fmt.Println("======================")
	providers := rpc.GetSupportedProviders()
	for i, provider := range providers {
		if i > 0 && i%5 == 0 {
			fmt.Println()
		}
		fmt.Printf("%-12s ", provider)
	}
	fmt.Println("\n")

	fmt.Println("⛓️  Supported Chains:")
	fmt.Println("====================")
	chains := rpc.GetSupportedChains()
	for i, chain := range chains {
		if i > 0 && i%4 == 0 {
			fmt.Println()
		}
		fmt.Printf("%-15s ", chain)
	}
	fmt.Println("\n")

	// Demonstrate health check methods
	fmt.Println("🏥 Health Check Methods:")
	fmt.Println("=======================")
	methods := rpc.GetHealthCheckMethods()
	for i, method := range methods {
		description := getHealthCheckDescription(method)
		fmt.Printf("%d. %s: %s\n", i+1, method, description)
	}
	fmt.Println()

	// Simulate RPC client usage (without actual connections)
	fmt.Println("🔄 RPC Client Usage Simulation:")
	fmt.Println("===============================")

	// Create a mock client configuration
	clientConfig := rpc.ClientConfig{
		Chain:     "ethereum",
		SessionID: "example-session-123",
	}

	fmt.Printf("Client Configuration:\n")
	fmt.Printf("  Chain: %s\n", clientConfig.Chain)
	fmt.Printf("  Session ID: %s\n", clientConfig.SessionID)
	fmt.Println()

	// Demonstrate typical RPC operations that would be performed
	fmt.Println("📋 Typical RPC Operations:")
	fmt.Println("==========================")
	operations := []string{
		"ChainID() - Get blockchain chain ID",
		"BlockNumber() - Get latest block number",
		"BalanceAt(address) - Get account balance",
		"TransactionByHash(hash) - Get transaction details",
		"SendTransaction(tx) - Submit transaction",
		"CallContract(msg) - Execute contract call",
		"EstimateGas(msg) - Estimate gas usage",
		"FilterLogs(query) - Query event logs",
	}

	for i, operation := range operations {
		fmt.Printf("%d. %s\n", i+1, operation)
	}
	fmt.Println()

	// Performance considerations
	fmt.Println("⚡ Performance Features:")
	fmt.Println("=======================")
	fmt.Println("✅ Automatic failover and retry logic")
	fmt.Println("✅ Load balancing across multiple nodes")
	fmt.Println("✅ Health monitoring and node recovery")
	fmt.Println("✅ Request metrics and performance tracking")
	fmt.Println("✅ Sticky sessions for WebSocket connections")
	fmt.Println("✅ Configurable timeouts and retry policies")
	fmt.Println("✅ Multi-chain support with chain-specific routing")
	fmt.Println("✅ Real-time node health and performance metrics")
	fmt.Println()

	// Best practices
	fmt.Println("💡 Best Practices:")
	fmt.Println("==================")
	fmt.Println("1. Configure multiple RPC providers for redundancy")
	fmt.Println("2. Use appropriate health check intervals for your use case")
	fmt.Println("3. Monitor node performance and adjust weights accordingly")
	fmt.Println("4. Implement proper error handling in your application")
	fmt.Println("5. Use sticky sessions for WebSocket subscriptions")
	fmt.Println("6. Set reasonable timeouts to avoid hanging requests")
	fmt.Println("7. Monitor and alert on node health and performance")
	fmt.Println("8. Regularly review and update node configurations")
	fmt.Println()

	fmt.Println("🎉 RPC Node Management System example completed!")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("  ✅ Multi-provider RPC node management")
	fmt.Println("  ✅ Automatic failover and load balancing")
	fmt.Println("  ✅ Health monitoring and recovery")
	fmt.Println("  ✅ Performance metrics and monitoring")
	fmt.Println("  ✅ Configurable strategies and policies")
	fmt.Println("  ✅ Multi-chain support")
	fmt.Println("  ✅ Production-ready configuration profiles")
	fmt.Println()
	fmt.Println("Configure real RPC endpoints and API keys for production use.")
}

func createExampleConfig() rpc.NodeManagerConfig {
	return rpc.NodeManagerConfig{
		Enabled:               true,
		HealthCheckInterval:   30 * time.Second,
		FailoverTimeout:       5 * time.Second,
		MaxRetries:            3,
		LoadBalancingStrategy: "round_robin",
		Nodes: []rpc.NodeConfig{
			{
				ID:          "infura_mainnet",
				Name:        "Infura Ethereum Mainnet",
				URL:         "https://mainnet.infura.io/v3/YOUR_PROJECT_ID",
				Chain:       "ethereum",
				Provider:    "infura",
				Priority:    1,
				Weight:      decimal.NewFromFloat(1.0),
				MaxRequests: 1000,
				Timeout:     10 * time.Second,
				Headers:     map[string]string{"User-Agent": "CryptoWallet/1.0"},
				Enabled:     true,
			},
			{
				ID:          "alchemy_mainnet",
				Name:        "Alchemy Ethereum Mainnet",
				URL:         "https://eth-mainnet.alchemyapi.io/v2/YOUR_API_KEY",
				Chain:       "ethereum",
				Provider:    "alchemy",
				Priority:    2,
				Weight:      decimal.NewFromFloat(1.0),
				MaxRequests: 1000,
				Timeout:     10 * time.Second,
				Headers:     map[string]string{"User-Agent": "CryptoWallet/1.0"},
				Enabled:     true,
			},
			{
				ID:          "quicknode_mainnet",
				Name:        "QuickNode Ethereum Mainnet",
				URL:         "https://YOUR_ENDPOINT.quiknode.pro/YOUR_API_KEY/",
				Chain:       "ethereum",
				Provider:    "quicknode",
				Priority:    3,
				Weight:      decimal.NewFromFloat(0.8),
				MaxRequests: 800,
				Timeout:     10 * time.Second,
				Headers:     map[string]string{"User-Agent": "CryptoWallet/1.0"},
				Enabled:     true,
			},
		},
		HealthCheckConfig: rpc.HealthCheckConfig{
			Enabled:            true,
			Interval:           30 * time.Second,
			Timeout:            5 * time.Second,
			MaxLatency:         2 * time.Second,
			MinSuccessRate:     decimal.NewFromFloat(0.95),
			CheckMethods:       []string{"chain_id", "block_number", "eth_syncing"},
			UnhealthyThreshold: 3,
			HealthyThreshold:   2,
		},
		LoadBalancerConfig: rpc.LoadBalancerConfig{
			Strategy:           "round_robin",
			WeightedRoundRobin: false,
			StickySession:      false,
			SessionTimeout:     30 * time.Minute,
			MaxRequestsPerNode: 1000,
		},
		MetricsConfig: rpc.MetricsConfig{
			Enabled:            true,
			CollectionInterval: 1 * time.Minute,
			RetentionPeriod:    24 * time.Hour,
			ExportEnabled:      false,
		},
	}
}

func displayMetrics(metrics map[string]interface{}) {
	fmt.Printf("  Total Nodes: %v\n", metrics["total_nodes"])
	fmt.Printf("  Healthy Nodes: %v\n", metrics["healthy_nodes"])
	fmt.Printf("  Active Nodes: %v\n", metrics["active_nodes"])
	fmt.Printf("  Total Requests: %v\n", metrics["total_requests"])
	fmt.Printf("  Average Latency: %v\n", metrics["average_latency"])
	fmt.Printf("  Is Running: %v\n", metrics["is_running"])
	fmt.Printf("  Uptime: %v\n", metrics["uptime"])
	fmt.Println()
}

func maskURL(url string) string {
	// Mask API keys in URLs for display
	if strings.Contains(url, "/v3/") {
		parts := strings.Split(url, "/v3/")
		if len(parts) == 2 {
			return parts[0] + "/v3/***"
		}
	}
	if strings.Contains(url, "/v2/") {
		parts := strings.Split(url, "/v2/")
		if len(parts) == 2 {
			return parts[0] + "/v2/***"
		}
	}
	if strings.Contains(url, ".pro/") {
		parts := strings.Split(url, ".pro/")
		if len(parts) == 2 {
			return parts[0] + ".pro/***"
		}
	}
	return url
}

func getStrategyDescription(strategy string) string {
	descriptions := map[string]string{
		"round_robin":          "Distributes requests evenly across all healthy nodes",
		"weighted_round_robin": "Distributes requests based on node weights",
		"least_connections":    "Routes to the node with the fewest active requests",
		"lowest_latency":       "Routes to the node with the lowest response latency",
	}
	if desc, exists := descriptions[strategy]; exists {
		return desc
	}
	return "Unknown strategy"
}

func getHealthCheckDescription(method string) string {
	descriptions := map[string]string{
		"chain_id":       "Verifies the blockchain chain ID",
		"block_number":   "Checks the latest block number",
		"net_peer_count": "Verifies network peer connections",
		"eth_syncing":    "Checks if the node is syncing",
		"eth_gas_price":  "Verifies gas price estimation",
		"net_version":    "Checks the network version",
	}
	if desc, exists := descriptions[method]; exists {
		return desc
	}
	return "Unknown health check method"
}
