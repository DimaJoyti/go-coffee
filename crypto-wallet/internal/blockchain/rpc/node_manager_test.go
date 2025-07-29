package rpc

import (
	"context"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a test logger
func createTestLoggerForRPC() *logger.Logger {
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	return logger.NewLogger(logConfig)
}

// Helper function to create test node manager config
func createTestNodeManagerConfig() NodeManagerConfig {
	return NodeManagerConfig{
		Enabled:               true,
		HealthCheckInterval:   1 * time.Second,
		FailoverTimeout:       100 * time.Millisecond,
		MaxRetries:            2,
		LoadBalancingStrategy: "round_robin",
		Nodes: []NodeConfig{
			{
				ID:          "test_node_1",
				Name:        "Test Node 1",
				URL:         "http://localhost:8545",
				Chain:       "ethereum",
				Provider:    "test",
				Priority:    1,
				Weight:      decimal.NewFromFloat(1.0),
				MaxRequests: 1000,
				Timeout:     5 * time.Second,
				Enabled:     true,
			},
			{
				ID:          "test_node_2",
				Name:        "Test Node 2",
				URL:         "http://localhost:8546",
				Chain:       "ethereum",
				Provider:    "test",
				Priority:    2,
				Weight:      decimal.NewFromFloat(0.8),
				MaxRequests: 800,
				Timeout:     5 * time.Second,
				Enabled:     true,
			},
		},
		HealthCheckConfig: HealthCheckConfig{
			Enabled:            false, // Disable for testing
			Interval:           1 * time.Second,
			Timeout:            1 * time.Second,
			MaxLatency:         500 * time.Millisecond,
			MinSuccessRate:     decimal.NewFromFloat(0.9),
			CheckMethods:       []string{"chain_id"},
			UnhealthyThreshold: 2,
			HealthyThreshold:   1,
		},
		LoadBalancerConfig: LoadBalancerConfig{
			Strategy:           "round_robin",
			WeightedRoundRobin: false,
			StickySession:      false,
			SessionTimeout:     10 * time.Minute,
			MaxRequestsPerNode: 1000,
		},
		MetricsConfig: MetricsConfig{
			Enabled:            true,
			CollectionInterval: 1 * time.Second,
			RetentionPeriod:    1 * time.Hour,
			ExportEnabled:      false,
		},
	}
}

func TestNewNodeManager(t *testing.T) {
	logger := createTestLoggerForRPC()
	config := createTestNodeManagerConfig()

	nm := NewNodeManager(logger, config)

	assert.NotNil(t, nm)
	assert.Equal(t, config.Enabled, nm.config.Enabled)
	assert.Equal(t, config.LoadBalancingStrategy, nm.config.LoadBalancingStrategy)
	assert.False(t, nm.IsRunning())
	assert.NotNil(t, nm.healthChecker)
	assert.NotNil(t, nm.loadBalancer)
	assert.NotNil(t, nm.metrics)
}

func TestNodeManager_StartStop(t *testing.T) {
	logger := createTestLoggerForRPC()
	config := createTestNodeManagerConfig()
	// Disable actual node connections for testing
	config.Nodes = []NodeConfig{}

	nm := NewNodeManager(logger, config)
	ctx := context.Background()

	// Test start with no nodes (should fail)
	err := nm.Start(ctx)
	assert.Error(t, err)
	assert.False(t, nm.IsRunning())

	// Test stop when not running
	err = nm.Stop()
	assert.NoError(t, err)
}

func TestNodeManager_StartDisabled(t *testing.T) {
	logger := createTestLoggerForRPC()
	config := createTestNodeManagerConfig()
	config.Enabled = false

	nm := NewNodeManager(logger, config)
	ctx := context.Background()

	err := nm.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, nm.IsRunning()) // Should remain false when disabled
}

func TestNodeManager_GetMetrics(t *testing.T) {
	logger := createTestLoggerForRPC()
	config := createTestNodeManagerConfig()
	config.Nodes = []NodeConfig{} // No actual nodes for testing

	nm := NewNodeManager(logger, config)

	metrics := nm.GetMetrics()
	assert.NotNil(t, metrics)

	// Validate metrics structure
	assert.Contains(t, metrics, "total_nodes")
	assert.Contains(t, metrics, "healthy_nodes")
	assert.Contains(t, metrics, "active_nodes")
	assert.Contains(t, metrics, "total_requests")
	assert.Contains(t, metrics, "average_latency")
	assert.Contains(t, metrics, "is_running")
	assert.Contains(t, metrics, "uptime")

	assert.Equal(t, 0, metrics["total_nodes"])
	assert.Equal(t, 0, metrics["healthy_nodes"])
	assert.Equal(t, 0, metrics["active_nodes"])
	assert.Equal(t, false, metrics["is_running"])
}

func TestLoadBalancer_SelectNode(t *testing.T) {
	logger := createTestLoggerForRPC()
	config := LoadBalancerConfig{
		Strategy:      "round_robin",
		StickySession: false,
	}

	// Create mock nodes
	nodes := map[string]*RPCNode{
		"node1": {
			Config:   NodeConfig{ID: "node1", Weight: decimal.NewFromFloat(1.0)},
			Health:   &NodeHealth{IsHealthy: true},
			IsActive: true,
		},
		"node2": {
			Config:   NodeConfig{ID: "node2", Weight: decimal.NewFromFloat(1.0)},
			Health:   &NodeHealth{IsHealthy: true},
			IsActive: true,
		},
		"node3": {
			Config:   NodeConfig{ID: "node3", Weight: decimal.NewFromFloat(1.0)},
			Health:   &NodeHealth{IsHealthy: false}, // Unhealthy node
			IsActive: true,
		},
	}

	lb := &LoadBalancer{
		logger:   logger.Named("load-balancer"),
		config:   config,
		nodes:    nodes,
		sessions: make(map[string]string),
	}

	// Test round-robin selection
	selected1 := lb.SelectNode("")
	assert.Contains(t, []string{"node1", "node2"}, selected1) // Should not select unhealthy node3

	selected2 := lb.SelectNode("")
	assert.Contains(t, []string{"node1", "node2"}, selected2)
	assert.NotEqual(t, selected1, selected2) // Should be different due to round-robin
}

func TestLoadBalancer_StickySession(t *testing.T) {
	logger := createTestLoggerForRPC()
	config := LoadBalancerConfig{
		Strategy:      "round_robin",
		StickySession: true,
	}

	nodes := map[string]*RPCNode{
		"node1": {
			Config:   NodeConfig{ID: "node1"},
			Health:   &NodeHealth{IsHealthy: true},
			IsActive: true,
		},
		"node2": {
			Config:   NodeConfig{ID: "node2"},
			Health:   &NodeHealth{IsHealthy: true},
			IsActive: true,
		},
	}

	lb := &LoadBalancer{
		logger:   logger.Named("load-balancer"),
		config:   config,
		nodes:    nodes,
		sessions: make(map[string]string),
	}

	sessionID := "test-session"

	// First selection should create session
	selected1 := lb.SelectNode(sessionID)
	assert.Contains(t, []string{"node1", "node2"}, selected1)

	// Second selection with same session should return same node
	selected2 := lb.SelectNode(sessionID)
	assert.Equal(t, selected1, selected2)
}

func TestGetDefaultNodeManagerConfig(t *testing.T) {
	config := GetDefaultNodeManagerConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, 30*time.Second, config.HealthCheckInterval)
	assert.Equal(t, 5*time.Second, config.FailoverTimeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, "round_robin", config.LoadBalancingStrategy)
	assert.NotEmpty(t, config.Nodes)

	// Check health check config
	assert.True(t, config.HealthCheckConfig.Enabled)
	assert.Equal(t, 30*time.Second, config.HealthCheckConfig.Interval)
	assert.NotEmpty(t, config.HealthCheckConfig.CheckMethods)

	// Check load balancer config
	assert.Equal(t, "round_robin", config.LoadBalancerConfig.Strategy)
	assert.False(t, config.LoadBalancerConfig.StickySession)

	// Check metrics config
	assert.True(t, config.MetricsConfig.Enabled)
	assert.Equal(t, 1*time.Minute, config.MetricsConfig.CollectionInterval)
}

func TestValidateNodeManagerConfig(t *testing.T) {
	// Test valid config
	validConfig := GetDefaultNodeManagerConfig()
	err := ValidateNodeManagerConfig(validConfig)
	assert.NoError(t, err)

	// Test disabled config
	disabledConfig := GetDefaultNodeManagerConfig()
	disabledConfig.Enabled = false
	err = ValidateNodeManagerConfig(disabledConfig)
	assert.NoError(t, err)

	// Test invalid configs
	invalidConfigs := []NodeManagerConfig{
		// Invalid health check interval
		{
			Enabled:             true,
			HealthCheckInterval: 0,
		},
		// Invalid failover timeout
		{
			Enabled:             true,
			HealthCheckInterval: 30 * time.Second,
			FailoverTimeout:     0,
		},
		// Negative max retries
		{
			Enabled:             true,
			HealthCheckInterval: 30 * time.Second,
			FailoverTimeout:     5 * time.Second,
			MaxRetries:          -1,
		},
		// No nodes
		{
			Enabled:             true,
			HealthCheckInterval: 30 * time.Second,
			FailoverTimeout:     5 * time.Second,
			MaxRetries:          3,
			Nodes:               []NodeConfig{},
		},
	}

	for i, config := range invalidConfigs {
		err := ValidateNodeManagerConfig(config)
		assert.Error(t, err, "Config %d should be invalid", i)
	}
}

func TestValidateNodeConfig(t *testing.T) {
	// Test valid node config
	validNode := NodeConfig{
		ID:          "test_node",
		Name:        "Test Node",
		URL:         "http://localhost:8545",
		Chain:       "ethereum",
		Provider:    "test",
		Priority:    1,
		Weight:      decimal.NewFromFloat(1.0),
		MaxRequests: 1000,
		Timeout:     5 * time.Second,
		Enabled:     true,
	}
	err := validateNodeConfig(validNode)
	assert.NoError(t, err)

	// Test invalid node configs
	invalidNodes := []NodeConfig{
		// Empty ID
		{
			URL:      "http://localhost:8545",
			Chain:    "ethereum",
			Provider: "test",
			Timeout:  5 * time.Second,
		},
		// Empty URL
		{
			ID:       "test",
			Chain:    "ethereum",
			Provider: "test",
			Timeout:  5 * time.Second,
		},
		// Empty chain
		{
			ID:       "test",
			URL:      "http://localhost:8545",
			Provider: "test",
			Timeout:  5 * time.Second,
		},
		// Invalid timeout
		{
			ID:       "test",
			URL:      "http://localhost:8545",
			Chain:    "ethereum",
			Provider: "test",
			Timeout:  0,
		},
	}

	for i, node := range invalidNodes {
		err := validateNodeConfig(node)
		assert.Error(t, err, "Node config %d should be invalid", i)
	}
}

func TestConfigVariants(t *testing.T) {
	// Test high availability config
	haConfig := GetHighAvailabilityConfig()
	assert.True(t, haConfig.HealthCheckInterval < GetDefaultNodeManagerConfig().HealthCheckInterval)
	assert.True(t, haConfig.FailoverTimeout < GetDefaultNodeManagerConfig().FailoverTimeout)
	assert.Equal(t, "lowest_latency", haConfig.LoadBalancingStrategy)

	// Test development config
	devConfig := GetDevelopmentConfig()
	assert.True(t, devConfig.HealthCheckInterval > GetDefaultNodeManagerConfig().HealthCheckInterval)
	assert.Equal(t, "round_robin", devConfig.LoadBalancingStrategy)
	assert.Len(t, devConfig.Nodes, 1) // Only one node for development

	// Test multi-chain config
	multiChainConfig := GetMultiChainConfig()
	assert.True(t, len(multiChainConfig.Nodes) > len(GetDefaultNodeManagerConfig().Nodes))

	// Check that multiple chains are represented
	chains := make(map[string]bool)
	for _, node := range multiChainConfig.Nodes {
		chains[node.Chain] = true
	}
	assert.True(t, len(chains) > 1, "Multi-chain config should have multiple chains")

	// Validate all configs
	assert.NoError(t, ValidateNodeManagerConfig(haConfig))
	assert.NoError(t, ValidateNodeManagerConfig(devConfig))
	assert.NoError(t, ValidateNodeManagerConfig(multiChainConfig))
}

func TestUtilityFunctions(t *testing.T) {
	// Test supported providers
	providers := GetSupportedProviders()
	assert.NotEmpty(t, providers)
	assert.Contains(t, providers, "infura")
	assert.Contains(t, providers, "alchemy")

	// Test supported chains
	chains := GetSupportedChains()
	assert.NotEmpty(t, chains)
	assert.Contains(t, chains, "ethereum")
	assert.Contains(t, chains, "polygon")

	// Test load balancing strategies
	strategies := GetLoadBalancingStrategies()
	assert.NotEmpty(t, strategies)
	assert.Contains(t, strategies, "round_robin")
	assert.Contains(t, strategies, "lowest_latency")

	// Test health check methods
	methods := GetHealthCheckMethods()
	assert.NotEmpty(t, methods)
	assert.Contains(t, methods, "chain_id")
	assert.Contains(t, methods, "block_number")
}
