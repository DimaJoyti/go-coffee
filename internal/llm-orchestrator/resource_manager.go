package llmorchestrator

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

// ResourceManager handles intelligent resource allocation and optimization
type ResourceManager struct {
	logger        *zap.Logger
	kubeClient    kubernetes.Interface
	metricsClient versioned.Interface
	config        *ResourceManagerConfig
	cache         *ResourceCache
	optimizer     *ResourceOptimizer
	monitor       *ResourceMonitor
	mutex         sync.RWMutex
}

// ResourceManagerConfig defines resource management configuration
type ResourceManagerConfig struct {
	// Resource allocation
	DefaultCPURequest    string `yaml:"defaultCPURequest"`    // e.g., "1000m"
	DefaultMemoryRequest string `yaml:"defaultMemoryRequest"` // e.g., "2Gi"
	DefaultGPURequest    int32  `yaml:"defaultGPURequest"`    // e.g., 1

	// Optimization settings
	OptimizationInterval      time.Duration `yaml:"optimizationInterval"`      // e.g., 5m
	ResourceUtilizationTarget float64       `yaml:"resourceUtilizationTarget"` // e.g., 0.8 (80%)
	OvercommitRatio           float64       `yaml:"overcommitRatio"`           // e.g., 1.2 (20% overcommit)

	// Auto-scaling
	ScaleUpThreshold   float64       `yaml:"scaleUpThreshold"`   // e.g., 0.8
	ScaleDownThreshold float64       `yaml:"scaleDownThreshold"` // e.g., 0.3
	ScaleUpCooldown    time.Duration `yaml:"scaleUpCooldown"`    // e.g., 3m
	ScaleDownCooldown  time.Duration `yaml:"scaleDownCooldown"`  // e.g., 10m

	// Resource limits
	MaxCPUPerWorkload    string `yaml:"maxCPUPerWorkload"`    // e.g., "8000m"
	MaxMemoryPerWorkload string `yaml:"maxMemoryPerWorkload"` // e.g., "32Gi"
	MaxGPUPerWorkload    int32  `yaml:"maxGPUPerWorkload"`    // e.g., 4

	// Quality of Service
	QoSResourceAllocation map[string]QoSResourceConfig `yaml:"qosResourceAllocation"`
}

// QoSResourceConfig defines resource allocation per QoS class
type QoSResourceConfig struct {
	CPUMultiplier    float64 `yaml:"cpuMultiplier"`    // e.g., 1.5 for premium
	MemoryMultiplier float64 `yaml:"memoryMultiplier"` // e.g., 2.0 for premium
	GPUPriority      int32   `yaml:"gpuPriority"`      // 1-10, higher = more priority
	BurstingAllowed  bool    `yaml:"burstingAllowed"`  // allow resource bursting
}

// ResourceCache stores resource information for fast access
type ResourceCache struct {
	nodes           map[string]*NodeResourceInfo
	workloads       map[string]*WorkloadResourceInfo
	clusterCapacity *ClusterResourceInfo
	lastUpdated     time.Time
	mutex           sync.RWMutex
}

// NodeResourceInfo contains resource information for a node
type NodeResourceInfo struct {
	Name               string
	Capacity           ResourceCapacity
	Allocatable        ResourceCapacity
	Used               ResourceCapacity
	Available          ResourceCapacity
	UtilizationPercent ResourceUtilization
	WorkloadCount      int32
	LastUpdated        time.Time
	PerformanceMetrics NodePerformanceMetrics
}

// WorkloadResourceInfo contains resource information for a workload
type WorkloadResourceInfo struct {
	Name               string
	Namespace          string
	Requested          ResourceCapacity
	Used               ResourceCapacity
	Limits             ResourceCapacity
	UtilizationPercent ResourceUtilization
	QoSClass           string
	LastUpdated        time.Time
	PerformanceMetrics WorkloadPerformanceMetrics
}

// ClusterResourceInfo contains cluster-wide resource information
type ClusterResourceInfo struct {
	TotalCapacity      ResourceCapacity
	TotalAllocatable   ResourceCapacity
	TotalUsed          ResourceCapacity
	TotalAvailable     ResourceCapacity
	UtilizationPercent ResourceUtilization
	NodeCount          int32
	WorkloadCount      int32
	LastUpdated        time.Time
}

// ResourceUtilization represents resource utilization percentages
type ResourceUtilization struct {
	CPU     float64 `json:"cpu"`
	Memory  float64 `json:"memory"`
	GPU     float64 `json:"gpu"`
	Storage float64 `json:"storage"`
}

// NodePerformanceMetrics tracks node performance
type NodePerformanceMetrics struct {
	CPUFrequency       float64 `json:"cpuFrequency"`       // GHz
	MemoryBandwidth    float64 `json:"memoryBandwidth"`    // GB/s
	NetworkBandwidth   float64 `json:"networkBandwidth"`   // Gbps
	StorageIOPS        float64 `json:"storageIOPS"`        // IOPS
	GPUMemoryBandwidth float64 `json:"gpuMemoryBandwidth"` // GB/s
	Temperature        float64 `json:"temperature"`        // Celsius
}

// WorkloadPerformanceMetrics tracks workload performance
type WorkloadPerformanceMetrics struct {
	RequestsPerSecond float64       `json:"requestsPerSecond"`
	AverageLatency    time.Duration `json:"averageLatency"`
	TokensPerSecond   float64       `json:"tokensPerSecond"`
	ErrorRate         float64       `json:"errorRate"`
	QueueLength       int32         `json:"queueLength"`
	ThroughputMBps    float64       `json:"throughputMBps"`
}

// ResourceOptimizer handles resource optimization algorithms
type ResourceOptimizer struct {
	logger *zap.Logger
	config *ResourceManagerConfig
}

// ResourceMonitor handles real-time resource monitoring
type ResourceMonitor struct {
	logger        *zap.Logger
	kubeClient    kubernetes.Interface
	metricsClient versioned.Interface
	cache         *ResourceCache
	stopCh        chan struct{}
}

// NewResourceManager creates a new resource manager instance
func NewResourceManager(logger *zap.Logger, kubeClient kubernetes.Interface, metricsClient versioned.Interface, config *ResourceManagerConfig) *ResourceManager {
	cache := &ResourceCache{
		nodes:           make(map[string]*NodeResourceInfo),
		workloads:       make(map[string]*WorkloadResourceInfo),
		clusterCapacity: &ClusterResourceInfo{},
	}

	optimizer := &ResourceOptimizer{
		logger: logger,
		config: config,
	}

	monitor := &ResourceMonitor{
		logger:        logger,
		kubeClient:    kubeClient,
		metricsClient: metricsClient,
		cache:         cache,
		stopCh:        make(chan struct{}),
	}

	return &ResourceManager{
		logger:        logger,
		kubeClient:    kubeClient,
		metricsClient: metricsClient,
		config:        config,
		cache:         cache,
		optimizer:     optimizer,
		monitor:       monitor,
	}
}

// Start begins resource monitoring and optimization
func (rm *ResourceManager) Start(ctx context.Context) error {
	rm.logger.Info("Starting resource manager")

	// Start resource monitoring
	go rm.monitor.Start(ctx)

	// Start optimization loop
	go rm.startOptimizationLoop(ctx)

	return nil
}

// Stop stops the resource manager
func (rm *ResourceManager) Stop() {
	rm.logger.Info("Stopping resource manager")
	close(rm.monitor.stopCh)
}

// AllocateResources allocates resources for a workload
func (rm *ResourceManager) AllocateResources(ctx context.Context, workload *LLMWorkload) (*ResourceAllocation, error) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	rm.logger.Info("Allocating resources for workload",
		zap.String("workload", workload.Name),
		zap.String("model", workload.Spec.ModelName),
	)

	// Calculate resource requirements
	requirements, err := rm.calculateResourceRequirements(workload)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate resource requirements: %w", err)
	}

	// Find suitable nodes
	suitableNodes, err := rm.findSuitableNodes(ctx, requirements)
	if err != nil {
		return nil, fmt.Errorf("failed to find suitable nodes: %w", err)
	}

	if len(suitableNodes) == 0 {
		return nil, fmt.Errorf("no suitable nodes found for workload")
	}

	// Select optimal node
	selectedNode := rm.selectOptimalNode(suitableNodes, requirements)

	// Create resource allocation
	allocation := &ResourceAllocation{
		WorkloadName:       workload.Name,
		WorkloadNamespace:  workload.Namespace,
		NodeName:           selectedNode.Name,
		AllocatedResources: requirements.Requested,
		QoSClass:           workload.Spec.Performance.SLARequirements.ErrorRate < 0.01, // Determine QoS based on SLA
		AllocationTime:     time.Now(),
	}

	// Update cache
	rm.updateAllocationCache(allocation)

	rm.logger.Info("Successfully allocated resources",
		zap.String("workload", workload.Name),
		zap.String("node", selectedNode.Name),
		zap.Float64("cpu", requirements.Requested.CPU),
		zap.Float64("memory", requirements.Requested.Memory),
	)

	return allocation, nil
}

// ResourceAllocation represents an allocated resource set
type ResourceAllocation struct {
	WorkloadName       string
	WorkloadNamespace  string
	NodeName           string
	AllocatedResources ResourceCapacity
	QoSClass           bool
	AllocationTime     time.Time
}

// ResourceRequirements represents calculated resource needs
type ResourceRequirements struct {
	Requested ResourceCapacity
	Limits    ResourceCapacity
	QoSClass  string
}

// calculateResourceRequirements calculates optimal resource allocation
func (rm *ResourceManager) calculateResourceRequirements(workload *LLMWorkload) (*ResourceRequirements, error) {
	spec := workload.Spec

	// Base resource requirements
	baseCPU := rm.parseResourceQuantity(spec.Resources.CPU, rm.config.DefaultCPURequest)
	baseMemory := rm.parseResourceQuantity(spec.Resources.Memory, rm.config.DefaultMemoryRequest)
	baseGPU := float64(rm.config.DefaultGPURequest)

	if spec.Resources.GPU != "" {
		baseGPU = float64(parseGPURequirement(spec.Resources.GPU))
	}

	// Apply model-specific multipliers
	modelMultipliers := rm.getModelResourceMultipliers(spec.ModelName, spec.ModelSize)

	requestedCPU := baseCPU * modelMultipliers.CPU
	requestedMemory := baseMemory * modelMultipliers.Memory
	requestedGPU := baseGPU * modelMultipliers.GPU

	// Apply QoS multipliers
	qosClass := rm.determineQoSClass(workload)
	qosMultipliers := rm.getQoSMultipliers(qosClass)

	requestedCPU *= qosMultipliers.CPUMultiplier
	requestedMemory *= qosMultipliers.MemoryMultiplier

	// Calculate limits (typically 1.5-2x requests)
	limitCPU := requestedCPU * 1.5
	limitMemory := requestedMemory * 1.5
	limitGPU := requestedGPU // GPU limits typically equal requests

	// Apply maximum limits
	maxCPU := rm.parseResourceQuantity(rm.config.MaxCPUPerWorkload, "8000m")
	maxMemory := rm.parseResourceQuantity(rm.config.MaxMemoryPerWorkload, "32Gi")
	maxGPU := float64(rm.config.MaxGPUPerWorkload)

	requestedCPU = math.Min(requestedCPU, maxCPU)
	requestedMemory = math.Min(requestedMemory, maxMemory)
	requestedGPU = math.Min(requestedGPU, maxGPU)
	limitCPU = math.Min(limitCPU, maxCPU)
	limitMemory = math.Min(limitMemory, maxMemory)
	limitGPU = math.Min(limitGPU, maxGPU)

	return &ResourceRequirements{
		Requested: ResourceCapacity{
			CPU:    requestedCPU,
			Memory: requestedMemory,
			GPU:    int32(requestedGPU),
		},
		Limits: ResourceCapacity{
			CPU:    limitCPU,
			Memory: limitMemory,
			GPU:    int32(limitGPU),
		},
		QoSClass: qosClass,
	}, nil
}

// ModelResourceMultipliers defines resource multipliers for different models
type ModelResourceMultipliers struct {
	CPU    float64
	Memory float64
	GPU    float64
}

// getModelResourceMultipliers returns resource multipliers based on model characteristics
func (rm *ResourceManager) getModelResourceMultipliers(modelName, modelSize string) ModelResourceMultipliers {
	// Model size multipliers
	sizeMultipliers := map[string]ModelResourceMultipliers{
		"small":  {CPU: 0.5, Memory: 0.5, GPU: 0.5},
		"medium": {CPU: 1.0, Memory: 1.0, GPU: 1.0},
		"large":  {CPU: 2.0, Memory: 2.0, GPU: 1.5},
		"xlarge": {CPU: 4.0, Memory: 4.0, GPU: 2.0},
	}

	// Model type multipliers
	typeMultipliers := map[string]ModelResourceMultipliers{
		"llama":  {CPU: 1.2, Memory: 1.5, GPU: 1.0},
		"gpt":    {CPU: 1.0, Memory: 1.2, GPU: 1.0},
		"bert":   {CPU: 0.8, Memory: 0.8, GPU: 0.8},
		"t5":     {CPU: 1.1, Memory: 1.3, GPU: 1.0},
		"gemini": {CPU: 1.3, Memory: 1.4, GPU: 1.1},
	}

	// Get base multipliers
	baseMultiplier := sizeMultipliers["medium"] // Default
	if multiplier, exists := sizeMultipliers[modelSize]; exists {
		baseMultiplier = multiplier
	}

	// Apply model type adjustments
	for modelType, typeMultiplier := range typeMultipliers {
		if contains(modelName, modelType) {
			baseMultiplier.CPU *= typeMultiplier.CPU
			baseMultiplier.Memory *= typeMultiplier.Memory
			baseMultiplier.GPU *= typeMultiplier.GPU
			break
		}
	}

	return baseMultiplier
}

// determineQoSClass determines the QoS class for a workload
func (rm *ResourceManager) determineQoSClass(workload *LLMWorkload) string {
	sla := workload.Spec.Performance.SLARequirements

	// High-performance requirements
	if sla.Availability >= 99.9 && sla.ResponseTimeP95 <= 100 {
		return "premium"
	}

	// Standard performance requirements
	if sla.Availability >= 99.5 && sla.ResponseTimeP95 <= 500 {
		return "standard"
	}

	// Basic performance requirements
	return "basic"
}

// getQoSMultipliers returns resource multipliers for QoS classes
func (rm *ResourceManager) getQoSMultipliers(qosClass string) QoSResourceConfig {
	if config, exists := rm.config.QoSResourceAllocation[qosClass]; exists {
		return config
	}

	// Default multipliers
	return QoSResourceConfig{
		CPUMultiplier:    1.0,
		MemoryMultiplier: 1.0,
		GPUPriority:      5,
		BurstingAllowed:  true,
	}
}

// findSuitableNodes finds nodes that can accommodate the resource requirements
func (rm *ResourceManager) findSuitableNodes(ctx context.Context, requirements *ResourceRequirements) ([]*NodeResourceInfo, error) {
	rm.cache.mutex.RLock()
	defer rm.cache.mutex.RUnlock()

	var suitableNodes []*NodeResourceInfo

	for _, node := range rm.cache.nodes {
		if rm.canNodeAccommodate(node, requirements) {
			suitableNodes = append(suitableNodes, node)
		}
	}

	return suitableNodes, nil
}

// canNodeAccommodate checks if a node can accommodate the resource requirements
func (rm *ResourceManager) canNodeAccommodate(node *NodeResourceInfo, requirements *ResourceRequirements) bool {
	// Check CPU availability
	if node.Available.CPU < requirements.Requested.CPU {
		return false
	}

	// Check memory availability
	if node.Available.Memory < requirements.Requested.Memory {
		return false
	}

	// Check GPU availability
	if node.Available.GPU < requirements.Requested.GPU {
		return false
	}

	// Check utilization thresholds
	if node.UtilizationPercent.CPU > rm.config.ResourceUtilizationTarget*100 {
		return false
	}

	if node.UtilizationPercent.Memory > rm.config.ResourceUtilizationTarget*100 {
		return false
	}

	return true
}

// selectOptimalNode selects the best node from suitable candidates
func (rm *ResourceManager) selectOptimalNode(nodes []*NodeResourceInfo, requirements *ResourceRequirements) *NodeResourceInfo {
	if len(nodes) == 0 {
		return nil
	}

	// Score nodes based on multiple factors
	bestNode := nodes[0]
	bestScore := rm.calculateNodeScore(bestNode, requirements)

	for _, node := range nodes[1:] {
		score := rm.calculateNodeScore(node, requirements)
		if score > bestScore {
			bestScore = score
			bestNode = node
		}
	}

	return bestNode
}

// calculateNodeScore calculates a score for node selection
func (rm *ResourceManager) calculateNodeScore(node *NodeResourceInfo, requirements *ResourceRequirements) float64 {
	// Resource availability score (0-100)
	cpuScore := (node.Available.CPU / node.Capacity.CPU) * 100
	memoryScore := (node.Available.Memory / node.Capacity.Memory) * 100
	gpuScore := float64(node.Available.GPU) / float64(node.Capacity.GPU) * 100

	// Utilization score (prefer balanced utilization)
	utilizationScore := 100 - math.Abs(node.UtilizationPercent.CPU-50) - math.Abs(node.UtilizationPercent.Memory-50)

	// Performance score
	performanceScore := rm.calculatePerformanceScore(node)

	// Weighted average
	return (cpuScore*0.3 + memoryScore*0.3 + gpuScore*0.2 + utilizationScore*0.1 + performanceScore*0.1)
}

// calculatePerformanceScore calculates performance score for a node
func (rm *ResourceManager) calculatePerformanceScore(node *NodeResourceInfo) float64 {
	// Normalize performance metrics to 0-100 scale
	cpuScore := math.Min(100, node.PerformanceMetrics.CPUFrequency/4.0*100)           // Assume 4GHz is max
	memoryScore := math.Min(100, node.PerformanceMetrics.MemoryBandwidth/100.0*100)   // Assume 100GB/s is max
	networkScore := math.Min(100, node.PerformanceMetrics.NetworkBandwidth/100.0*100) // Assume 100Gbps is max

	return (cpuScore + memoryScore + networkScore) / 3.0
}

// Helper functions
func (rm *ResourceManager) parseResourceQuantity(value, defaultValue string) float64 {
	if value == "" {
		value = defaultValue
	}

	quantity, err := resource.ParseQuantity(value)
	if err != nil {
		rm.logger.Warn("Failed to parse resource quantity", zap.String("value", value), zap.Error(err))
		defaultQuantity, _ := resource.ParseQuantity(defaultValue)
		return defaultQuantity.AsApproximateFloat64()
	}

	return quantity.AsApproximateFloat64()
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

func (rm *ResourceManager) updateAllocationCache(allocation *ResourceAllocation) {
	rm.cache.mutex.Lock()
	defer rm.cache.mutex.Unlock()

	// Update workload cache
	workloadKey := fmt.Sprintf("%s/%s", allocation.WorkloadNamespace, allocation.WorkloadName)
	if workloadInfo, exists := rm.cache.workloads[workloadKey]; exists {
		workloadInfo.Requested = allocation.AllocatedResources
		workloadInfo.LastUpdated = time.Now()
	}

	// Update node cache
	if nodeInfo, exists := rm.cache.nodes[allocation.NodeName]; exists {
		nodeInfo.Used.CPU += allocation.AllocatedResources.CPU
		nodeInfo.Used.Memory += allocation.AllocatedResources.Memory
		nodeInfo.Used.GPU += allocation.AllocatedResources.GPU

		nodeInfo.Available.CPU -= allocation.AllocatedResources.CPU
		nodeInfo.Available.Memory -= allocation.AllocatedResources.Memory
		nodeInfo.Available.GPU -= allocation.AllocatedResources.GPU

		nodeInfo.WorkloadCount++
		nodeInfo.LastUpdated = time.Now()

		// Recalculate utilization
		nodeInfo.UtilizationPercent.CPU = (nodeInfo.Used.CPU / nodeInfo.Capacity.CPU) * 100
		nodeInfo.UtilizationPercent.Memory = (nodeInfo.Used.Memory / nodeInfo.Capacity.Memory) * 100
		nodeInfo.UtilizationPercent.GPU = (float64(nodeInfo.Used.GPU) / float64(nodeInfo.Capacity.GPU)) * 100
	}
}

func (rm *ResourceManager) startOptimizationLoop(ctx context.Context) {
	ticker := time.NewTicker(rm.config.OptimizationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := rm.optimizeResourceAllocation(ctx); err != nil {
				rm.logger.Error("Resource optimization failed", zap.Error(err))
			}
		}
	}
}

func (rm *ResourceManager) optimizeResourceAllocation(ctx context.Context) error {
	rm.logger.Debug("Running resource optimization")

	rm.cache.mutex.RLock()
	defer rm.cache.mutex.RUnlock()

	// Identify underutilized nodes
	underutilizedNodes := rm.findUnderutilizedNodes()

	// Identify overutilized nodes
	overutilizedNodes := rm.findOverutilizedNodes()

	// Suggest workload migrations
	migrations := rm.suggestWorkloadMigrations(underutilizedNodes, overutilizedNodes)

	if len(migrations) > 0 {
		rm.logger.Info("Resource optimization suggestions",
			zap.Int("migrations", len(migrations)),
			zap.Int("underutilized_nodes", len(underutilizedNodes)),
			zap.Int("overutilized_nodes", len(overutilizedNodes)),
		)
	}

	return nil
}

func (rm *ResourceManager) findUnderutilizedNodes() []*NodeResourceInfo {
	var underutilized []*NodeResourceInfo

	for _, node := range rm.cache.nodes {
		avgUtilization := (node.UtilizationPercent.CPU + node.UtilizationPercent.Memory) / 2
		if avgUtilization < rm.config.ScaleDownThreshold*100 {
			underutilized = append(underutilized, node)
		}
	}

	return underutilized
}

func (rm *ResourceManager) findOverutilizedNodes() []*NodeResourceInfo {
	var overutilized []*NodeResourceInfo

	for _, node := range rm.cache.nodes {
		if node.UtilizationPercent.CPU > rm.config.ScaleUpThreshold*100 ||
			node.UtilizationPercent.Memory > rm.config.ScaleUpThreshold*100 {
			overutilized = append(overutilized, node)
		}
	}

	return overutilized
}

func (rm *ResourceManager) suggestWorkloadMigrations(underutilized, overutilized []*NodeResourceInfo) []WorkloadMigration {
	var migrations []WorkloadMigration

	// Simple migration strategy: move workloads from overutilized to underutilized nodes
	for _, overNode := range overutilized {
		for _, underNode := range underutilized {
			// Find workloads that can be migrated
			for workloadKey, workload := range rm.cache.workloads {
				if rm.canMigrateWorkload(workload, overNode.Name, underNode) {
					migrations = append(migrations, WorkloadMigration{
						WorkloadName: workloadKey,
						FromNode:     overNode.Name,
						ToNode:       underNode.Name,
						Reason:       "Load balancing",
					})
				}
			}
		}
	}

	return migrations
}

func (rm *ResourceManager) canMigrateWorkload(workload *WorkloadResourceInfo, fromNode string, toNode *NodeResourceInfo) bool {
	// Check if workload can fit on target node
	return toNode.Available.CPU >= workload.Requested.CPU &&
		toNode.Available.Memory >= workload.Requested.Memory &&
		toNode.Available.GPU >= workload.Requested.GPU
}

// WorkloadMigration represents a suggested workload migration
type WorkloadMigration struct {
	WorkloadName string
	FromNode     string
	ToNode       string
	Reason       string
}

// Start monitoring resources
func (monitor *ResourceMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Update every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-monitor.stopCh:
			return
		case <-ticker.C:
			if err := monitor.updateResourceCache(ctx); err != nil {
				monitor.logger.Error("Failed to update resource cache", zap.Error(err))
			}
		}
	}
}

func (monitor *ResourceMonitor) updateResourceCache(ctx context.Context) error {
	monitor.logger.Debug("Updating resource cache")

	// Get current nodes
	nodes, err := monitor.kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	monitor.cache.mutex.Lock()
	defer monitor.cache.mutex.Unlock()

	// Update node information
	for _, node := range nodes.Items {
		nodeInfo := &NodeResourceInfo{
			Name:        node.Name,
			LastUpdated: time.Now(),
		}

		// Extract capacity
		nodeInfo.Capacity = ResourceCapacity{
			CPU:    node.Status.Capacity.Cpu().AsApproximateFloat64(),
			Memory: node.Status.Capacity.Memory().AsApproximateFloat64(),
			GPU:    getGPUCapacity(&node),
		}

		// Extract allocatable
		nodeInfo.Allocatable = ResourceCapacity{
			CPU:    node.Status.Allocatable.Cpu().AsApproximateFloat64(),
			Memory: node.Status.Allocatable.Memory().AsApproximateFloat64(),
			GPU:    getGPUCapacity(&node),
		}

		// Calculate available (this would be updated with actual usage)
		nodeInfo.Available = nodeInfo.Allocatable

		monitor.cache.nodes[node.Name] = nodeInfo
	}

	// Update cluster capacity
	monitor.updateClusterCapacity()

	return nil
}

func (monitor *ResourceMonitor) updateClusterCapacity() {
	var totalCapacity, totalAllocatable, totalUsed ResourceCapacity
	nodeCount := int32(len(monitor.cache.nodes))

	for _, node := range monitor.cache.nodes {
		totalCapacity.CPU += node.Capacity.CPU
		totalCapacity.Memory += node.Capacity.Memory
		totalCapacity.GPU += node.Capacity.GPU

		totalAllocatable.CPU += node.Allocatable.CPU
		totalAllocatable.Memory += node.Allocatable.Memory
		totalAllocatable.GPU += node.Allocatable.GPU

		totalUsed.CPU += node.Used.CPU
		totalUsed.Memory += node.Used.Memory
		totalUsed.GPU += node.Used.GPU
	}

	monitor.cache.clusterCapacity = &ClusterResourceInfo{
		TotalCapacity:    totalCapacity,
		TotalAllocatable: totalAllocatable,
		TotalUsed:        totalUsed,
		TotalAvailable: ResourceCapacity{
			CPU:    totalAllocatable.CPU - totalUsed.CPU,
			Memory: totalAllocatable.Memory - totalUsed.Memory,
			GPU:    totalAllocatable.GPU - totalUsed.GPU,
		},
		UtilizationPercent: ResourceUtilization{
			CPU:    (totalUsed.CPU / totalAllocatable.CPU) * 100,
			Memory: (totalUsed.Memory / totalAllocatable.Memory) * 100,
			GPU:    (float64(totalUsed.GPU) / float64(totalAllocatable.GPU)) * 100,
		},
		NodeCount:     nodeCount,
		WorkloadCount: int32(len(monitor.cache.workloads)),
		LastUpdated:   time.Now(),
	}
}
