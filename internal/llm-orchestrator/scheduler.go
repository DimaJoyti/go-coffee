package llmorchestrator

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// LLMScheduler handles intelligent scheduling of LLM workloads
type LLMScheduler struct {
	logger     *zap.Logger
	kubeClient kubernetes.Interface
	config     *SchedulerConfig
	metrics    *SchedulerMetrics
}

// SchedulerConfig defines scheduler configuration
type SchedulerConfig struct {
	// Scheduling strategy
	Strategy string `yaml:"strategy"` // round-robin, least-loaded, performance-aware, cost-optimized

	// Resource allocation
	ResourceOvercommitRatio float64 `yaml:"resourceOvercommitRatio"` // 1.2 = 20% overcommit
	GPUFragmentationThreshold float64 `yaml:"gpuFragmentationThreshold"` // 0.8 = 80% utilization before fragmentation

	// Performance optimization
	LocalityPreference      bool    `yaml:"localityPreference"`      // prefer nodes with cached models
	ModelAffinityWeight     float64 `yaml:"modelAffinityWeight"`     // weight for model affinity
	LatencyOptimization     bool    `yaml:"latencyOptimization"`     // optimize for low latency
	ThroughputOptimization  bool    `yaml:"throughputOptimization"`  // optimize for high throughput

	// Scaling behavior
	ScaleUpCooldown   time.Duration `yaml:"scaleUpCooldown"`   // minimum time between scale-up operations
	ScaleDownCooldown time.Duration `yaml:"scaleDownCooldown"` // minimum time between scale-down operations
	PredictiveScaling bool          `yaml:"predictiveScaling"` // enable ML-based demand prediction

	// Quality of Service
	QoSClasses map[string]QoSConfig `yaml:"qosClasses"`
}

// QoSConfig defines quality of service parameters
type QoSConfig struct {
	Priority            int32   `yaml:"priority"`            // scheduling priority
	ResourceGuarantee   float64 `yaml:"resourceGuarantee"`   // guaranteed resource percentage
	MaxLatency          int32   `yaml:"maxLatency"`          // maximum acceptable latency in ms
	PreemptionPolicy    string  `yaml:"preemptionPolicy"`    // never, lower-priority, always
	BurstingAllowed     bool    `yaml:"burstingAllowed"`     // allow resource bursting
}

// SchedulerMetrics tracks scheduler performance
type SchedulerMetrics struct {
	SchedulingLatency     time.Duration `json:"schedulingLatency"`
	SuccessfulSchedulings int64         `json:"successfulSchedulings"`
	FailedSchedulings     int64         `json:"failedSchedulings"`
	NodeUtilization       map[string]float64 `json:"nodeUtilization"`
	ModelCacheHitRate     float64       `json:"modelCacheHitRate"`
	AverageQueueTime      time.Duration `json:"averageQueueTime"`
	ResourceEfficiency    float64       `json:"resourceEfficiency"`
}

// NodeScore represents a node's suitability for scheduling
type NodeScore struct {
	NodeName           string
	Score              float64
	ResourceScore      float64
	AffinityScore      float64
	LocalityScore      float64
	PerformanceScore   float64
	AvailableResources ResourceCapacity
	Reasons            []string
}

// ResourceCapacity represents available resources on a node
type ResourceCapacity struct {
	CPU     float64 `json:"cpu"`
	Memory  float64 `json:"memory"`
	GPU     int32   `json:"gpu"`
	Storage float64 `json:"storage"`
}

// SchedulingRequest represents a request to schedule an LLM workload
type SchedulingRequest struct {
	Workload     *LLMWorkload
	Priority     int32
	QoSClass     string
	Constraints  []SchedulingConstraint
	Preferences  []SchedulingPreference
	RequestTime  time.Time
}

// SchedulingConstraint defines hard requirements
type SchedulingConstraint struct {
	Type   string      `json:"type"`   // node-selector, anti-affinity, resource-limit
	Key    string      `json:"key"`
	Value  interface{} `json:"value"`
	Operator string    `json:"operator"` // equals, not-equals, in, not-in, greater-than, less-than
}

// SchedulingPreference defines soft requirements
type SchedulingPreference struct {
	Type   string      `json:"type"`   // node-preference, zone-preference, model-affinity
	Key    string      `json:"key"`
	Value  interface{} `json:"value"`
	Weight float64     `json:"weight"` // 0-100
}

// SchedulingResult represents the outcome of a scheduling decision
type SchedulingResult struct {
	Success       bool
	SelectedNode  string
	Score         float64
	Reason        string
	Alternatives  []NodeScore
	SchedulingTime time.Duration
	Warnings      []string
}

// NewLLMScheduler creates a new LLM scheduler instance
func NewLLMScheduler(logger *zap.Logger, kubeClient kubernetes.Interface, config *SchedulerConfig) *LLMScheduler {
	return &LLMScheduler{
		logger:     logger,
		kubeClient: kubeClient,
		config:     config,
		metrics:    &SchedulerMetrics{
			NodeUtilization: make(map[string]float64),
		},
	}
}

// ScheduleWorkload schedules an LLM workload to the most suitable node
func (s *LLMScheduler) ScheduleWorkload(ctx context.Context, request *SchedulingRequest) (*SchedulingResult, error) {
	startTime := time.Now()
	
	s.logger.Info("Scheduling LLM workload",
		zap.String("workload", request.Workload.Name),
		zap.String("model", request.Workload.Spec.ModelName),
		zap.String("qos", request.QoSClass),
	)

	// Get available nodes
	nodes, err := s.getAvailableNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get available nodes: %w", err)
	}

	if len(nodes) == 0 {
		return &SchedulingResult{
			Success: false,
			Reason:  "no available nodes",
		}, nil
	}

	// Score nodes based on suitability
	nodeScores, err := s.scoreNodes(ctx, request, nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to score nodes: %w", err)
	}

	// Filter nodes based on constraints
	filteredScores := s.filterNodesByConstraints(request, nodeScores)
	if len(filteredScores) == 0 {
		return &SchedulingResult{
			Success: false,
			Reason:  "no nodes satisfy constraints",
		}, nil
	}

	// Sort by score (highest first)
	sort.Slice(filteredScores, func(i, j int) bool {
		return filteredScores[i].Score > filteredScores[j].Score
	})

	selectedNode := filteredScores[0]
	schedulingTime := time.Since(startTime)

	// Update metrics
	s.updateSchedulingMetrics(true, schedulingTime)

	result := &SchedulingResult{
		Success:        true,
		SelectedNode:   selectedNode.NodeName,
		Score:          selectedNode.Score,
		Reason:         fmt.Sprintf("Best score: %.2f", selectedNode.Score),
		Alternatives:   filteredScores[1:],
		SchedulingTime: schedulingTime,
	}

	s.logger.Info("Successfully scheduled workload",
		zap.String("workload", request.Workload.Name),
		zap.String("node", selectedNode.NodeName),
		zap.Float64("score", selectedNode.Score),
		zap.Duration("schedulingTime", schedulingTime),
	)

	return result, nil
}

// getAvailableNodes retrieves all schedulable nodes
func (s *LLMScheduler) getAvailableNodes(ctx context.Context) ([]corev1.Node, error) {
	nodeList, err := s.kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var availableNodes []corev1.Node
	for _, node := range nodeList.Items {
		if s.isNodeSchedulable(&node) {
			availableNodes = append(availableNodes, node)
		}
	}

	return availableNodes, nil
}

// isNodeSchedulable checks if a node can accept new workloads
func (s *LLMScheduler) isNodeSchedulable(node *corev1.Node) bool {
	// Check if node is ready
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady && condition.Status != corev1.ConditionTrue {
			return false
		}
	}

	// Check if node is cordoned
	if node.Spec.Unschedulable {
		return false
	}

	// Check for taints that prevent scheduling
	for _, taint := range node.Spec.Taints {
		if taint.Effect == corev1.TaintEffectNoSchedule {
			return false
		}
	}

	return true
}

// scoreNodes calculates suitability scores for each node
func (s *LLMScheduler) scoreNodes(ctx context.Context, request *SchedulingRequest, nodes []corev1.Node) ([]NodeScore, error) {
	var scores []NodeScore

	for _, node := range nodes {
		score, err := s.calculateNodeScore(ctx, request, &node)
		if err != nil {
			s.logger.Warn("Failed to calculate node score",
				zap.String("node", node.Name),
				zap.Error(err),
			)
			continue
		}
		scores = append(scores, *score)
	}

	return scores, nil
}

// calculateNodeScore computes a comprehensive score for a node
func (s *LLMScheduler) calculateNodeScore(ctx context.Context, request *SchedulingRequest, node *corev1.Node) (*NodeScore, error) {
	score := &NodeScore{
		NodeName: node.Name,
		Reasons:  []string{},
	}

	// Calculate resource score (0-100)
	resourceScore, err := s.calculateResourceScore(request, node)
	if err != nil {
		return nil, err
	}
	score.ResourceScore = resourceScore

	// Calculate affinity score (0-100)
	affinityScore := s.calculateAffinityScore(request, node)
	score.AffinityScore = affinityScore

	// Calculate locality score (0-100)
	localityScore := s.calculateLocalityScore(request, node)
	score.LocalityScore = localityScore

	// Calculate performance score (0-100)
	performanceScore := s.calculatePerformanceScore(request, node)
	score.PerformanceScore = performanceScore

	// Weighted final score
	weights := s.getScoreWeights(request.QoSClass)
	score.Score = (resourceScore*weights.Resource +
		affinityScore*weights.Affinity +
		localityScore*weights.Locality +
		performanceScore*weights.Performance) / 100.0

	return score, nil
}

// ScoreWeights defines weights for different scoring factors
type ScoreWeights struct {
	Resource    float64
	Affinity    float64
	Locality    float64
	Performance float64
}

// getScoreWeights returns scoring weights based on QoS class
func (s *LLMScheduler) getScoreWeights(qosClass string) ScoreWeights {
	if qos, exists := s.config.QoSClasses[qosClass]; exists {
		// Customize weights based on QoS requirements
		if qos.MaxLatency < 100 { // Low latency requirements
			return ScoreWeights{
				Resource:    20.0,
				Affinity:    20.0,
				Locality:    30.0,
				Performance: 30.0,
			}
		}
	}

	// Default weights
	return ScoreWeights{
		Resource:    40.0,
		Affinity:    20.0,
		Locality:    20.0,
		Performance: 20.0,
	}
}

// calculateResourceScore evaluates resource availability
func (s *LLMScheduler) calculateResourceScore(request *SchedulingRequest, node *corev1.Node) (float64, error) {
	// Get node capacity and allocatable resources
	capacity := node.Status.Capacity
	allocatable := node.Status.Allocatable

	// Calculate resource requirements
	reqResources := request.Workload.Spec.Resources

	// CPU score
	cpuCapacity := allocatable.Cpu().AsApproximateFloat64()
	cpuRequired := parseCPURequirement(reqResources.CPU)
	cpuScore := math.Max(0, (cpuCapacity-cpuRequired)/cpuCapacity*100)

	// Memory score
	memCapacity := allocatable.Memory().AsApproximateFloat64()
	memRequired := parseMemoryRequirement(reqResources.Memory)
	memScore := math.Max(0, (memCapacity-memRequired)/memCapacity*100)

	// GPU score (if required)
	gpuScore := 100.0
	if reqResources.GPU != "" {
		gpuRequired := parseGPURequirement(reqResources.GPU)
		gpuCapacity := getGPUCapacity(node)
		if gpuCapacity > 0 {
			gpuScore = math.Max(0, float64(gpuCapacity-gpuRequired)/float64(gpuCapacity)*100)
		} else if gpuRequired > 0 {
			gpuScore = 0 // No GPU available but required
		}
	}

	// Weighted average of resource scores
	return (cpuScore + memScore + gpuScore) / 3.0, nil
}

// calculateAffinityScore evaluates node affinity preferences
func (s *LLMScheduler) calculateAffinityScore(request *SchedulingRequest, node *corev1.Node) float64 {
	score := 50.0 // Base score

	// Check node labels for affinity
	for _, preference := range request.Preferences {
		if preference.Type == "node-preference" {
			if value, exists := node.Labels[preference.Key]; exists {
				if value == preference.Value {
					score += preference.Weight
				}
			}
		}
	}

	return math.Min(100.0, score)
}

// calculateLocalityScore evaluates data locality benefits
func (s *LLMScheduler) calculateLocalityScore(request *SchedulingRequest, node *corev1.Node) float64 {
	if !s.config.LocalityPreference {
		return 50.0 // Neutral score if locality is not preferred
	}

	// Check if the model is already cached on this node
	modelName := request.Workload.Spec.ModelName
	if s.isModelCachedOnNode(modelName, node.Name) {
		return 100.0 // Perfect score for cached models
	}

	// Check zone locality
	zone := node.Labels["topology.kubernetes.io/zone"]
	if s.hasModelInZone(modelName, zone) {
		return 75.0 // Good score for same-zone models
	}

	return 25.0 // Lower score for remote models
}

// calculatePerformanceScore evaluates expected performance
func (s *LLMScheduler) calculatePerformanceScore(request *SchedulingRequest, node *corev1.Node) float64 {
	score := 50.0 // Base score

	// Check node performance characteristics
	if nodeType, exists := node.Labels["node.kubernetes.io/instance-type"]; exists {
		score += s.getNodeTypePerformanceBonus(nodeType)
	}

	// Check for high-performance storage
	if _, exists := node.Labels["storage.kubernetes.io/ssd"]; exists {
		score += 10.0
	}

	// Check for high-speed networking
	if bandwidth, exists := node.Labels["networking.kubernetes.io/bandwidth"]; exists {
		score += s.getNetworkPerformanceBonus(bandwidth)
	}

	return math.Min(100.0, score)
}

// Helper functions for resource parsing and node evaluation
func parseCPURequirement(cpu string) float64 {
	// Parse CPU requirement (e.g., "2000m" -> 2.0)
	// Implementation would parse the CPU string format
	return 2.0 // Placeholder
}

func parseMemoryRequirement(memory string) float64 {
	// Parse memory requirement (e.g., "8Gi" -> bytes)
	// Implementation would parse the memory string format
	return 8 * 1024 * 1024 * 1024 // Placeholder: 8GB in bytes
}

func parseGPURequirement(gpu string) int32 {
	// Parse GPU requirement
	return 1 // Placeholder
}

func getGPUCapacity(node *corev1.Node) int32 {
	// Get GPU capacity from node resources
	return 1 // Placeholder
}

func (s *LLMScheduler) isModelCachedOnNode(modelName, nodeName string) bool {
	// Check if model is cached on the specified node
	return false // Placeholder
}

func (s *LLMScheduler) hasModelInZone(modelName, zone string) bool {
	// Check if model is available in the specified zone
	return false // Placeholder
}

func (s *LLMScheduler) getNodeTypePerformanceBonus(nodeType string) float64 {
	// Return performance bonus based on node type
	performanceBonuses := map[string]float64{
		"c5.xlarge":   10.0,
		"c5.2xlarge":  15.0,
		"c5.4xlarge":  20.0,
		"p3.2xlarge":  30.0,
		"p3.8xlarge":  40.0,
		"p4d.24xlarge": 50.0,
	}
	
	if bonus, exists := performanceBonuses[nodeType]; exists {
		return bonus
	}
	return 0.0
}

func (s *LLMScheduler) getNetworkPerformanceBonus(bandwidth string) float64 {
	// Return performance bonus based on network bandwidth
	return 5.0 // Placeholder
}

// filterNodesByConstraints filters nodes based on hard constraints
func (s *LLMScheduler) filterNodesByConstraints(request *SchedulingRequest, scores []NodeScore) []NodeScore {
	var filtered []NodeScore

	for _, score := range scores {
		if s.satisfiesConstraints(request, score.NodeName) {
			filtered = append(filtered, score)
		}
	}

	return filtered
}

// satisfiesConstraints checks if a node satisfies all constraints
func (s *LLMScheduler) satisfiesConstraints(request *SchedulingRequest, nodeName string) bool {
	// Check all constraints
	for _, constraint := range request.Constraints {
		if !s.evaluateConstraint(constraint, nodeName) {
			return false
		}
	}
	return true
}

// evaluateConstraint evaluates a single constraint
func (s *LLMScheduler) evaluateConstraint(constraint SchedulingConstraint, nodeName string) bool {
	// Implementation would evaluate the specific constraint
	return true // Placeholder
}

// updateSchedulingMetrics updates scheduler performance metrics
func (s *LLMScheduler) updateSchedulingMetrics(success bool, duration time.Duration) {
	s.metrics.SchedulingLatency = duration
	
	if success {
		s.metrics.SuccessfulSchedulings++
	} else {
		s.metrics.FailedSchedulings++
	}
}

// GetMetrics returns current scheduler metrics
func (s *LLMScheduler) GetMetrics() *SchedulerMetrics {
	return s.metrics
}
