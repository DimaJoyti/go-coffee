package rpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// NodeManager manages RPC nodes with automatic failover and load balancing
type NodeManager struct {
	logger *logger.Logger
	config NodeManagerConfig

	// Node management
	nodes       map[string]*RPCNode
	nodesMutex  sync.RWMutex
	activeNodes []string

	// Health monitoring
	healthChecker *HealthChecker

	// Load balancing
	loadBalancer *LoadBalancer

	// Metrics
	metrics *NodeMetrics

	// State management
	isRunning bool
	stopChan  chan struct{}
	mutex     sync.RWMutex
}

// NodeManagerConfig holds configuration for node management
type NodeManagerConfig struct {
	Enabled               bool               `json:"enabled" yaml:"enabled"`
	HealthCheckInterval   time.Duration      `json:"health_check_interval" yaml:"health_check_interval"`
	FailoverTimeout       time.Duration      `json:"failover_timeout" yaml:"failover_timeout"`
	MaxRetries            int                `json:"max_retries" yaml:"max_retries"`
	LoadBalancingStrategy string             `json:"load_balancing_strategy" yaml:"load_balancing_strategy"`
	Nodes                 []NodeConfig       `json:"nodes" yaml:"nodes"`
	HealthCheckConfig     HealthCheckConfig  `json:"health_check_config" yaml:"health_check_config"`
	LoadBalancerConfig    LoadBalancerConfig `json:"load_balancer_config" yaml:"load_balancer_config"`
	MetricsConfig         MetricsConfig      `json:"metrics_config" yaml:"metrics_config"`
}

// NodeConfig holds configuration for individual RPC nodes
type NodeConfig struct {
	ID          string            `json:"id" yaml:"id"`
	Name        string            `json:"name" yaml:"name"`
	URL         string            `json:"url" yaml:"url"`
	Chain       string            `json:"chain" yaml:"chain"`
	Provider    string            `json:"provider" yaml:"provider"`
	Priority    int               `json:"priority" yaml:"priority"`
	Weight      decimal.Decimal   `json:"weight" yaml:"weight"`
	MaxRequests int               `json:"max_requests" yaml:"max_requests"`
	Timeout     time.Duration     `json:"timeout" yaml:"timeout"`
	Headers     map[string]string `json:"headers" yaml:"headers"`
	Enabled     bool              `json:"enabled" yaml:"enabled"`
}

// HealthCheckConfig holds health check configuration
type HealthCheckConfig struct {
	Enabled            bool            `json:"enabled" yaml:"enabled"`
	Interval           time.Duration   `json:"interval" yaml:"interval"`
	Timeout            time.Duration   `json:"timeout" yaml:"timeout"`
	MaxLatency         time.Duration   `json:"max_latency" yaml:"max_latency"`
	MinSuccessRate     decimal.Decimal `json:"min_success_rate" yaml:"min_success_rate"`
	CheckMethods       []string        `json:"check_methods" yaml:"check_methods"`
	UnhealthyThreshold int             `json:"unhealthy_threshold" yaml:"unhealthy_threshold"`
	HealthyThreshold   int             `json:"healthy_threshold" yaml:"healthy_threshold"`
}

// LoadBalancerConfig holds load balancer configuration
type LoadBalancerConfig struct {
	Strategy           string        `json:"strategy" yaml:"strategy"`
	WeightedRoundRobin bool          `json:"weighted_round_robin" yaml:"weighted_round_robin"`
	StickySession      bool          `json:"sticky_session" yaml:"sticky_session"`
	SessionTimeout     time.Duration `json:"session_timeout" yaml:"session_timeout"`
	MaxRequestsPerNode int           `json:"max_requests_per_node" yaml:"max_requests_per_node"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled            bool          `json:"enabled" yaml:"enabled"`
	CollectionInterval time.Duration `json:"collection_interval" yaml:"collection_interval"`
	RetentionPeriod    time.Duration `json:"retention_period" yaml:"retention_period"`
	ExportEnabled      bool          `json:"export_enabled" yaml:"export_enabled"`
}

// RPCNode represents an RPC node with health and performance metrics
type RPCNode struct {
	Config    NodeConfig        `json:"config"`
	Client    *ethclient.Client `json:"-"`
	RawClient *rpc.Client       `json:"-"`
	Health    *NodeHealth       `json:"health"`
	Metrics   *NodeMetrics      `json:"metrics"`
	LastUsed  time.Time         `json:"last_used"`
	IsActive  bool              `json:"is_active"`
	mutex     sync.RWMutex
}

// NodeHealth represents node health status
type NodeHealth struct {
	IsHealthy        bool            `json:"is_healthy"`
	LastCheck        time.Time       `json:"last_check"`
	Latency          time.Duration   `json:"latency"`
	SuccessRate      decimal.Decimal `json:"success_rate"`
	ConsecutiveFails int             `json:"consecutive_fails"`
	TotalChecks      int             `json:"total_checks"`
	SuccessfulChecks int             `json:"successful_checks"`
	LastError        string          `json:"last_error"`
	BlockHeight      uint64          `json:"block_height"`
	PeerCount        int             `json:"peer_count"`
	Syncing          bool            `json:"syncing"`
}

// NodeMetrics represents node performance metrics
type NodeMetrics struct {
	TotalRequests      int64           `json:"total_requests"`
	SuccessfulRequests int64           `json:"successful_requests"`
	FailedRequests     int64           `json:"failed_requests"`
	AverageLatency     time.Duration   `json:"average_latency"`
	RequestsPerSecond  decimal.Decimal `json:"requests_per_second"`
	ErrorRate          decimal.Decimal `json:"error_rate"`
	LastRequestTime    time.Time       `json:"last_request_time"`
	Uptime             time.Duration   `json:"uptime"`
	StartTime          time.Time       `json:"start_time"`
}

// HealthChecker monitors node health
type HealthChecker struct {
	logger   *logger.Logger
	config   HealthCheckConfig
	nodes    map[string]*RPCNode
	mutex    sync.RWMutex
	stopChan chan struct{}
}

// LoadBalancer handles request distribution
type LoadBalancer struct {
	logger   *logger.Logger
	config   LoadBalancerConfig
	nodes    map[string]*RPCNode
	current  int
	sessions map[string]string // session ID -> node ID
	mutex    sync.RWMutex
}

// NewNodeManager creates a new RPC node manager
func NewNodeManager(logger *logger.Logger, config NodeManagerConfig) *NodeManager {
	nm := &NodeManager{
		logger:   logger.Named("node-manager"),
		config:   config,
		nodes:    make(map[string]*RPCNode),
		stopChan: make(chan struct{}),
		metrics:  &NodeMetrics{StartTime: time.Now()},
	}

	// Initialize health checker
	nm.healthChecker = &HealthChecker{
		logger:   logger.Named("health-checker"),
		config:   config.HealthCheckConfig,
		nodes:    nm.nodes,
		stopChan: make(chan struct{}),
	}

	// Initialize load balancer
	nm.loadBalancer = &LoadBalancer{
		logger:   logger.Named("load-balancer"),
		config:   config.LoadBalancerConfig,
		nodes:    nm.nodes,
		sessions: make(map[string]string),
	}

	return nm
}

// Start starts the node manager
func (nm *NodeManager) Start(ctx context.Context) error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	if nm.isRunning {
		return fmt.Errorf("node manager is already running")
	}

	if !nm.config.Enabled {
		nm.logger.Info("Node manager is disabled")
		return nil
	}

	nm.logger.Info("Starting RPC node manager",
		zap.Int("node_count", len(nm.config.Nodes)),
		zap.String("load_balancing_strategy", nm.config.LoadBalancingStrategy))

	// Initialize nodes
	if err := nm.initializeNodes(ctx); err != nil {
		return fmt.Errorf("failed to initialize nodes: %w", err)
	}

	// Start health checker
	if nm.config.HealthCheckConfig.Enabled {
		go nm.healthChecker.Start(ctx)
	}

	// Start metrics collection
	if nm.config.MetricsConfig.Enabled {
		go nm.startMetricsCollection(ctx)
	}

	nm.isRunning = true
	nm.logger.Info("RPC node manager started successfully",
		zap.Int("active_nodes", len(nm.activeNodes)))

	return nil
}

// Stop stops the node manager
func (nm *NodeManager) Stop() error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	if !nm.isRunning {
		return nil
	}

	nm.logger.Info("Stopping RPC node manager")

	// Stop health checker
	close(nm.healthChecker.stopChan)

	// Stop metrics collection
	close(nm.stopChan)

	// Close all node connections
	nm.nodesMutex.Lock()
	for _, node := range nm.nodes {
		if node.Client != nil {
			node.Client.Close()
		}
		if node.RawClient != nil {
			node.RawClient.Close()
		}
	}
	nm.nodesMutex.Unlock()

	nm.isRunning = false
	nm.logger.Info("RPC node manager stopped")
	return nil
}

// initializeNodes initializes all configured nodes
func (nm *NodeManager) initializeNodes(ctx context.Context) error {
	nm.nodesMutex.Lock()
	defer nm.nodesMutex.Unlock()

	for _, nodeConfig := range nm.config.Nodes {
		if !nodeConfig.Enabled {
			nm.logger.Debug("Skipping disabled node", zap.String("node_id", nodeConfig.ID))
			continue
		}

		node, err := nm.createNode(ctx, nodeConfig)
		if err != nil {
			nm.logger.Warn("Failed to create node",
				zap.String("node_id", nodeConfig.ID),
				zap.Error(err))
			continue
		}

		nm.nodes[nodeConfig.ID] = node
		nm.activeNodes = append(nm.activeNodes, nodeConfig.ID)

		nm.logger.Info("Node initialized successfully",
			zap.String("node_id", nodeConfig.ID),
			zap.String("provider", nodeConfig.Provider),
			zap.String("chain", nodeConfig.Chain))
	}

	if len(nm.activeNodes) == 0 {
		return fmt.Errorf("no active nodes available")
	}

	return nil
}

// createNode creates a new RPC node
func (nm *NodeManager) createNode(ctx context.Context, config NodeConfig) (*RPCNode, error) {
	// Create raw RPC client
	rawClient, err := rpc.DialContext(ctx, config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to dial RPC: %w", err)
	}

	// Create eth client
	ethClient := ethclient.NewClient(rawClient)

	// Test connection
	_, err = ethClient.ChainID(ctx)
	if err != nil {
		rawClient.Close()
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	node := &RPCNode{
		Config:    config,
		Client:    ethClient,
		RawClient: rawClient,
		Health: &NodeHealth{
			IsHealthy:   true,
			LastCheck:   time.Now(),
			SuccessRate: decimal.NewFromFloat(1.0),
		},
		Metrics: &NodeMetrics{
			StartTime: time.Now(),
		},
		LastUsed: time.Now(),
		IsActive: true,
	}

	return node, nil
}

// GetHealthyNode returns a healthy node using load balancing
func (nm *NodeManager) GetHealthyNode(sessionID string) (*RPCNode, error) {
	nm.nodesMutex.RLock()
	defer nm.nodesMutex.RUnlock()

	if len(nm.activeNodes) == 0 {
		return nil, fmt.Errorf("no active nodes available")
	}

	// Use load balancer to select node
	nodeID := nm.loadBalancer.SelectNode(sessionID)
	if nodeID == "" {
		return nil, fmt.Errorf("load balancer failed to select node")
	}

	node, exists := nm.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("selected node not found: %s", nodeID)
	}

	if !node.Health.IsHealthy {
		// Try to find another healthy node
		for _, activeNodeID := range nm.activeNodes {
			if activeNode, exists := nm.nodes[activeNodeID]; exists && activeNode.Health.IsHealthy {
				return activeNode, nil
			}
		}
		return nil, fmt.Errorf("no healthy nodes available")
	}

	// Update last used time
	node.mutex.Lock()
	node.LastUsed = time.Now()
	node.mutex.Unlock()

	return node, nil
}

// GetNodeByID returns a specific node by ID
func (nm *NodeManager) GetNodeByID(nodeID string) (*RPCNode, error) {
	nm.nodesMutex.RLock()
	defer nm.nodesMutex.RUnlock()

	node, exists := nm.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}

	return node, nil
}

// GetAllNodes returns all nodes
func (nm *NodeManager) GetAllNodes() map[string]*RPCNode {
	nm.nodesMutex.RLock()
	defer nm.nodesMutex.RUnlock()

	nodes := make(map[string]*RPCNode)
	for id, node := range nm.nodes {
		nodes[id] = node
	}

	return nodes
}

// GetHealthyNodes returns all healthy nodes
func (nm *NodeManager) GetHealthyNodes() []*RPCNode {
	nm.nodesMutex.RLock()
	defer nm.nodesMutex.RUnlock()

	var healthyNodes []*RPCNode
	for _, node := range nm.nodes {
		if node.Health.IsHealthy {
			healthyNodes = append(healthyNodes, node)
		}
	}

	return healthyNodes
}

// IsRunning returns whether the node manager is running
func (nm *NodeManager) IsRunning() bool {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()
	return nm.isRunning
}

// GetMetrics returns node manager metrics
func (nm *NodeManager) GetMetrics() map[string]interface{} {
	nm.nodesMutex.RLock()
	defer nm.nodesMutex.RUnlock()

	healthyCount := 0
	totalRequests := int64(0)
	totalLatency := time.Duration(0)

	for _, node := range nm.nodes {
		if node.Health.IsHealthy {
			healthyCount++
		}
		totalRequests += node.Metrics.TotalRequests
		totalLatency += node.Metrics.AverageLatency
	}

	avgLatency := time.Duration(0)
	if len(nm.nodes) > 0 {
		avgLatency = totalLatency / time.Duration(len(nm.nodes))
	}

	return map[string]interface{}{
		"total_nodes":     len(nm.nodes),
		"healthy_nodes":   healthyCount,
		"active_nodes":    len(nm.activeNodes),
		"total_requests":  totalRequests,
		"average_latency": avgLatency,
		"is_running":      nm.isRunning,
		"uptime":          time.Since(nm.metrics.StartTime),
	}
}

// startMetricsCollection starts metrics collection
func (nm *NodeManager) startMetricsCollection(ctx context.Context) {
	ticker := time.NewTicker(nm.config.MetricsConfig.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-nm.stopChan:
			return
		case <-ticker.C:
			nm.collectMetrics()
		}
	}
}

// collectMetrics collects metrics from all nodes
func (nm *NodeManager) collectMetrics() {
	nm.nodesMutex.RLock()
	defer nm.nodesMutex.RUnlock()

	for _, node := range nm.nodes {
		node.mutex.Lock()
		// Update uptime
		node.Metrics.Uptime = time.Since(node.Metrics.StartTime)

		// Calculate requests per second
		if node.Metrics.Uptime > 0 {
			node.Metrics.RequestsPerSecond = decimal.NewFromInt(node.Metrics.TotalRequests).
				Div(decimal.NewFromFloat(node.Metrics.Uptime.Seconds()))
		}

		// Calculate error rate
		if node.Metrics.TotalRequests > 0 {
			node.Metrics.ErrorRate = decimal.NewFromInt(node.Metrics.FailedRequests).
				Div(decimal.NewFromInt(node.Metrics.TotalRequests))
		}
		node.mutex.Unlock()
	}
}

// Start starts the health checker
func (hc *HealthChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(hc.config.Interval)
	defer ticker.Stop()

	hc.logger.Info("Starting health checker",
		zap.Duration("interval", hc.config.Interval),
		zap.Duration("timeout", hc.config.Timeout))

	for {
		select {
		case <-ctx.Done():
			return
		case <-hc.stopChan:
			return
		case <-ticker.C:
			hc.checkAllNodes(ctx)
		}
	}
}

// checkAllNodes performs health checks on all nodes
func (hc *HealthChecker) checkAllNodes(ctx context.Context) {
	hc.mutex.RLock()
	nodes := make(map[string]*RPCNode)
	for id, node := range hc.nodes {
		nodes[id] = node
	}
	hc.mutex.RUnlock()

	for nodeID, node := range nodes {
		go hc.checkNode(ctx, nodeID, node)
	}
}

// checkNode performs health check on a single node
func (hc *HealthChecker) checkNode(ctx context.Context, nodeID string, node *RPCNode) {
	checkCtx, cancel := context.WithTimeout(ctx, hc.config.Timeout)
	defer cancel()

	startTime := time.Now()
	isHealthy := true
	var lastError string

	// Perform health checks
	for _, method := range hc.config.CheckMethods {
		if err := hc.performCheck(checkCtx, node, method); err != nil {
			isHealthy = false
			lastError = err.Error()
			hc.logger.Debug("Health check failed",
				zap.String("node_id", nodeID),
				zap.String("method", method),
				zap.Error(err))
			break
		}
	}

	latency := time.Since(startTime)

	// Update node health
	node.mutex.Lock()
	node.Health.LastCheck = time.Now()
	node.Health.Latency = latency
	node.Health.TotalChecks++

	if isHealthy {
		node.Health.SuccessfulChecks++
		node.Health.ConsecutiveFails = 0
	} else {
		node.Health.ConsecutiveFails++
		node.Health.LastError = lastError
	}

	// Calculate success rate
	node.Health.SuccessRate = decimal.NewFromInt(int64(node.Health.SuccessfulChecks)).
		Div(decimal.NewFromInt(int64(node.Health.TotalChecks)))

	// Determine if node is healthy
	wasHealthy := node.Health.IsHealthy
	if isHealthy && latency <= hc.config.MaxLatency &&
		node.Health.SuccessRate.GreaterThanOrEqual(hc.config.MinSuccessRate) {
		if !wasHealthy && node.Health.ConsecutiveFails == 0 {
			// Node recovered
			node.Health.IsHealthy = true
			hc.logger.Info("Node recovered",
				zap.String("node_id", nodeID),
				zap.Duration("latency", latency))
		} else if wasHealthy {
			node.Health.IsHealthy = true
		}
	} else {
		if node.Health.ConsecutiveFails >= hc.config.UnhealthyThreshold {
			node.Health.IsHealthy = false
			if wasHealthy {
				hc.logger.Warn("Node marked as unhealthy",
					zap.String("node_id", nodeID),
					zap.Int("consecutive_fails", node.Health.ConsecutiveFails),
					zap.String("error", lastError))
			}
		}
	}
	node.mutex.Unlock()
}

// performCheck performs a specific health check
func (hc *HealthChecker) performCheck(ctx context.Context, node *RPCNode, method string) error {
	switch method {
	case "chain_id":
		_, err := node.Client.ChainID(ctx)
		return err
	case "block_number":
		blockNumber, err := node.Client.BlockNumber(ctx)
		if err != nil {
			return err
		}
		node.mutex.Lock()
		node.Health.BlockHeight = blockNumber
		node.mutex.Unlock()
		return nil
	case "net_peer_count":
		var peerCount string
		err := node.RawClient.CallContext(ctx, &peerCount, "net_peerCount")
		if err != nil {
			return err
		}
		// Parse peer count and update health
		return nil
	case "eth_syncing":
		syncing, err := node.Client.SyncProgress(ctx)
		if err != nil {
			return err
		}
		node.mutex.Lock()
		node.Health.Syncing = syncing != nil
		node.mutex.Unlock()
		return nil
	default:
		return fmt.Errorf("unknown health check method: %s", method)
	}
}

// SelectNode selects a node using the configured load balancing strategy
func (lb *LoadBalancer) SelectNode(sessionID string) string {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	// Check for sticky session
	if lb.config.StickySession && sessionID != "" {
		if nodeID, exists := lb.sessions[sessionID]; exists {
			if node, exists := lb.nodes[nodeID]; exists && node.Health.IsHealthy {
				return nodeID
			}
			// Remove invalid session
			delete(lb.sessions, sessionID)
		}
	}

	// Get healthy nodes
	var healthyNodes []*RPCNode
	for _, node := range lb.nodes {
		if node.Health.IsHealthy && node.IsActive {
			healthyNodes = append(healthyNodes, node)
		}
	}

	if len(healthyNodes) == 0 {
		return ""
	}

	var selectedNode *RPCNode
	switch lb.config.Strategy {
	case "round_robin":
		selectedNode = lb.roundRobin(healthyNodes)
	case "weighted_round_robin":
		selectedNode = lb.weightedRoundRobin(healthyNodes)
	case "least_connections":
		selectedNode = lb.leastConnections(healthyNodes)
	case "lowest_latency":
		selectedNode = lb.lowestLatency(healthyNodes)
	default:
		selectedNode = lb.roundRobin(healthyNodes)
	}

	if selectedNode == nil {
		return ""
	}

	// Store session if sticky sessions enabled
	if lb.config.StickySession && sessionID != "" {
		lb.sessions[sessionID] = selectedNode.Config.ID
	}

	return selectedNode.Config.ID
}

// roundRobin implements round-robin load balancing
func (lb *LoadBalancer) roundRobin(nodes []*RPCNode) *RPCNode {
	if len(nodes) == 0 {
		return nil
	}

	lb.current = (lb.current + 1) % len(nodes)
	return nodes[lb.current]
}

// weightedRoundRobin implements weighted round-robin load balancing
func (lb *LoadBalancer) weightedRoundRobin(nodes []*RPCNode) *RPCNode {
	if len(nodes) == 0 {
		return nil
	}

	// Simple weighted selection based on node weights
	totalWeight := decimal.Zero
	for _, node := range nodes {
		totalWeight = totalWeight.Add(node.Config.Weight)
	}

	if totalWeight.IsZero() {
		return lb.roundRobin(nodes)
	}

	// Select based on weight
	target := decimal.NewFromFloat(float64(lb.current) / float64(len(nodes))).Mul(totalWeight)
	currentWeight := decimal.Zero

	for _, node := range nodes {
		currentWeight = currentWeight.Add(node.Config.Weight)
		if currentWeight.GreaterThanOrEqual(target) {
			lb.current = (lb.current + 1) % len(nodes)
			return node
		}
	}

	return nodes[0]
}

// leastConnections implements least connections load balancing
func (lb *LoadBalancer) leastConnections(nodes []*RPCNode) *RPCNode {
	if len(nodes) == 0 {
		return nil
	}

	var selectedNode *RPCNode
	minRequests := int64(-1)

	for _, node := range nodes {
		if minRequests == -1 || node.Metrics.TotalRequests < minRequests {
			minRequests = node.Metrics.TotalRequests
			selectedNode = node
		}
	}

	return selectedNode
}

// lowestLatency implements lowest latency load balancing
func (lb *LoadBalancer) lowestLatency(nodes []*RPCNode) *RPCNode {
	if len(nodes) == 0 {
		return nil
	}

	var selectedNode *RPCNode
	minLatency := time.Duration(-1)

	for _, node := range nodes {
		if minLatency == -1 || node.Health.Latency < minLatency {
			minLatency = node.Health.Latency
			selectedNode = node
		}
	}

	return selectedNode
}
