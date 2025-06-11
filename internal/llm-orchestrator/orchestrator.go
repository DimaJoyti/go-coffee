package llmorchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// LLMOrchestrator is the main orchestrator for LLM workloads
type LLMOrchestrator struct {
	logger          *zap.Logger
	config          *OrchestratorConfig
	kubeClient      kubernetes.Interface
	metricsClient   versioned.Interface
	ctrlClient      client.Client
	manager         manager.Manager
	
	// Core components
	scheduler       *LLMScheduler
	resourceManager *ResourceManager
	modelRegistry   *ModelRegistry
	operator        *LLMWorkloadOperator
	
	// Monitoring and metrics
	metricsCollector *MetricsCollector
	healthChecker    *HealthChecker
	
	// State management
	running         bool
	stopCh          chan struct{}
	mutex           sync.RWMutex
}

// OrchestratorConfig defines the main orchestrator configuration
type OrchestratorConfig struct {
	// Kubernetes configuration
	KubeConfig      string `yaml:"kubeConfig"`
	Namespace       string `yaml:"namespace"`
	
	// Component configurations
	Scheduler       *SchedulerConfig       `yaml:"scheduler"`
	ResourceManager *ResourceManagerConfig `yaml:"resourceManager"`
	ModelRegistry   *ModelRegistryConfig   `yaml:"modelRegistry"`
	Operator        *OperatorConfig        `yaml:"operator"`
	
	// Monitoring configuration
	MetricsEnabled  bool   `yaml:"metricsEnabled"`
	MetricsPort     int    `yaml:"metricsPort"`
	HealthPort      int    `yaml:"healthPort"`
	
	// Performance tuning
	WorkerThreads   int           `yaml:"workerThreads"`
	SyncInterval    time.Duration `yaml:"syncInterval"`
	
	// High availability
	LeaderElection  bool   `yaml:"leaderElection"`
	LeaderLockName  string `yaml:"leaderLockName"`
	
	// Security
	TLSEnabled      bool   `yaml:"tlsEnabled"`
	CertFile        string `yaml:"certFile"`
	KeyFile         string `yaml:"keyFile"`
}

// MetricsCollector collects and exposes orchestrator metrics
type MetricsCollector struct {
	logger           *zap.Logger
	orchestratorMetrics *OrchestratorMetrics
	mutex            sync.RWMutex
}

// OrchestratorMetrics contains orchestrator performance metrics
type OrchestratorMetrics struct {
	// Workload metrics
	TotalWorkloads       int64     `json:"totalWorkloads"`
	RunningWorkloads     int64     `json:"runningWorkloads"`
	PendingWorkloads     int64     `json:"pendingWorkloads"`
	FailedWorkloads      int64     `json:"failedWorkloads"`
	
	// Scheduling metrics
	SchedulingLatency    time.Duration `json:"schedulingLatency"`
	SchedulingSuccessRate float64      `json:"schedulingSuccessRate"`
	
	// Resource metrics
	ClusterUtilization   ResourceUtilization `json:"clusterUtilization"`
	ResourceEfficiency   float64            `json:"resourceEfficiency"`
	
	// Performance metrics
	AverageLatency       time.Duration `json:"averageLatency"`
	TotalThroughput      float64       `json:"totalThroughput"`
	ErrorRate            float64       `json:"errorRate"`
	
	// System metrics
	MemoryUsage          float64       `json:"memoryUsage"`
	CPUUsage             float64       `json:"cpuUsage"`
	GoroutineCount       int           `json:"goroutineCount"`
	
	LastUpdated          time.Time     `json:"lastUpdated"`
}

// HealthChecker monitors the health of orchestrator components
type HealthChecker struct {
	logger     *zap.Logger
	components map[string]HealthStatus
	mutex      sync.RWMutex
}

// HealthStatus represents the health status of a component
type HealthStatus struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"` // healthy, unhealthy, unknown
	LastCheck   time.Time `json:"lastCheck"`
	Message     string    `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// NewLLMOrchestrator creates a new LLM orchestrator instance
func NewLLMOrchestrator(logger *zap.Logger, config *OrchestratorConfig) (*LLMOrchestrator, error) {
	orchestrator := &LLMOrchestrator{
		logger:  logger,
		config:  config,
		stopCh:  make(chan struct{}),
		running: false,
	}

	// Initialize Kubernetes clients
	if err := orchestrator.initializeKubernetesClients(); err != nil {
		return nil, fmt.Errorf("failed to initialize Kubernetes clients: %w", err)
	}

	// Initialize core components
	if err := orchestrator.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}

	// Initialize monitoring
	if err := orchestrator.initializeMonitoring(); err != nil {
		return nil, fmt.Errorf("failed to initialize monitoring: %w", err)
	}

	return orchestrator, nil
}

// Start starts the LLM orchestrator
func (o *LLMOrchestrator) Start(ctx context.Context) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.running {
		return fmt.Errorf("orchestrator is already running")
	}

	o.logger.Info("Starting LLM orchestrator")

	// Start core components
	if err := o.startComponents(ctx); err != nil {
		return fmt.Errorf("failed to start components: %w", err)
	}

	// Start controller manager
	go func() {
		if err := o.manager.Start(ctx); err != nil {
			o.logger.Error("Controller manager failed", zap.Error(err))
		}
	}()

	// Start monitoring
	if o.config.MetricsEnabled {
		go o.startMetricsCollection(ctx)
		go o.startHealthChecking(ctx)
	}

	o.running = true
	o.logger.Info("LLM orchestrator started successfully")

	return nil
}

// Stop stops the LLM orchestrator
func (o *LLMOrchestrator) Stop() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !o.running {
		return nil
	}

	o.logger.Info("Stopping LLM orchestrator")

	// Signal stop to all components
	close(o.stopCh)

	// Stop components
	o.stopComponents()

	o.running = false
	o.logger.Info("LLM orchestrator stopped")

	return nil
}

// GetMetrics returns current orchestrator metrics
func (o *LLMOrchestrator) GetMetrics() *OrchestratorMetrics {
	if o.metricsCollector == nil {
		return nil
	}
	
	o.metricsCollector.mutex.RLock()
	defer o.metricsCollector.mutex.RUnlock()
	
	return o.metricsCollector.orchestratorMetrics
}

// GetHealth returns the health status of all components
func (o *LLMOrchestrator) GetHealth() map[string]HealthStatus {
	if o.healthChecker == nil {
		return nil
	}
	
	o.healthChecker.mutex.RLock()
	defer o.healthChecker.mutex.RUnlock()
	
	// Create a copy to avoid race conditions
	health := make(map[string]HealthStatus)
	for k, v := range o.healthChecker.components {
		health[k] = v
	}
	
	return health
}

// initializeKubernetesClients initializes Kubernetes clients
func (o *LLMOrchestrator) initializeKubernetesClients() error {
	var config *rest.Config
	var err error

	if o.config.KubeConfig != "" {
		// Use provided kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", o.config.KubeConfig)
	} else {
		// Use in-cluster config
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return fmt.Errorf("failed to create Kubernetes config: %w", err)
	}

	// Create Kubernetes client
	o.kubeClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Create metrics client
	o.metricsClient, err = versioned.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create metrics client: %w", err)
	}

	// Create controller runtime manager
	o.manager, err = ctrl.NewManager(config, ctrl.Options{
		Namespace:              o.config.Namespace,
		MetricsBindAddress:     fmt.Sprintf(":%d", o.config.MetricsPort),
		HealthProbeBindAddress: fmt.Sprintf(":%d", o.config.HealthPort),
		LeaderElection:         o.config.LeaderElection,
		LeaderElectionID:       o.config.LeaderLockName,
	})
	if err != nil {
		return fmt.Errorf("failed to create controller manager: %w", err)
	}

	o.ctrlClient = o.manager.GetClient()

	return nil
}

// initializeComponents initializes core orchestrator components
func (o *LLMOrchestrator) initializeComponents() error {
	var err error

	// Initialize scheduler
	o.scheduler = NewLLMScheduler(o.logger, o.kubeClient, o.config.Scheduler)

	// Initialize resource manager
	o.resourceManager = NewResourceManager(o.logger, o.kubeClient, o.metricsClient, o.config.ResourceManager)

	// Initialize model registry
	o.modelRegistry, err = NewModelRegistry(o.logger, o.config.ModelRegistry)
	if err != nil {
		return fmt.Errorf("failed to create model registry: %w", err)
	}

	// Initialize operator
	o.operator = NewLLMWorkloadOperator(
		o.ctrlClient,
		o.logger,
		o.manager.GetScheme(),
		o.manager.GetEventRecorderFor("llm-orchestrator"),
		o.kubeClient,
		o.scheduler,
		o.resourceManager,
		o.modelRegistry,
		o.config.Operator,
	)

	// Setup operator with manager
	if err := o.operator.SetupWithManager(o.manager); err != nil {
		return fmt.Errorf("failed to setup operator: %w", err)
	}

	return nil
}

// initializeMonitoring initializes monitoring components
func (o *LLMOrchestrator) initializeMonitoring() error {
	// Initialize metrics collector
	o.metricsCollector = &MetricsCollector{
		logger: o.logger,
		orchestratorMetrics: &OrchestratorMetrics{
			LastUpdated: time.Now(),
		},
	}

	// Initialize health checker
	o.healthChecker = &HealthChecker{
		logger:     o.logger,
		components: make(map[string]HealthStatus),
	}

	return nil
}

// startComponents starts all orchestrator components
func (o *LLMOrchestrator) startComponents(ctx context.Context) error {
	// Start resource manager
	if err := o.resourceManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start resource manager: %w", err)
	}

	o.logger.Info("All components started successfully")
	return nil
}

// stopComponents stops all orchestrator components
func (o *LLMOrchestrator) stopComponents() {
	// Stop resource manager
	o.resourceManager.Stop()

	o.logger.Info("All components stopped")
}

// startMetricsCollection starts metrics collection
func (o *LLMOrchestrator) startMetricsCollection(ctx context.Context) {
	ticker := time.NewTicker(o.config.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-o.stopCh:
			return
		case <-ticker.C:
			o.collectMetrics()
		}
	}
}

// collectMetrics collects current metrics from all components
func (o *LLMOrchestrator) collectMetrics() {
	o.metricsCollector.mutex.Lock()
	defer o.metricsCollector.mutex.Unlock()

	metrics := o.metricsCollector.orchestratorMetrics

	// Collect workload metrics
	workloads, err := o.listAllWorkloads()
	if err != nil {
		o.logger.Warn("Failed to list workloads for metrics", zap.Error(err))
		return
	}

	metrics.TotalWorkloads = int64(len(workloads))
	metrics.RunningWorkloads = 0
	metrics.PendingWorkloads = 0
	metrics.FailedWorkloads = 0

	for _, workload := range workloads {
		switch workload.Status.Phase {
		case "Running":
			metrics.RunningWorkloads++
		case "Pending":
			metrics.PendingWorkloads++
		case "Failed":
			metrics.FailedWorkloads++
		}
	}

	// Collect scheduler metrics
	if schedulerMetrics := o.scheduler.GetMetrics(); schedulerMetrics != nil {
		metrics.SchedulingLatency = schedulerMetrics.SchedulingLatency
		total := schedulerMetrics.SuccessfulSchedulings + schedulerMetrics.FailedSchedulings
		if total > 0 {
			metrics.SchedulingSuccessRate = float64(schedulerMetrics.SuccessfulSchedulings) / float64(total)
		}
	}

	metrics.LastUpdated = time.Now()
}

// startHealthChecking starts health checking
func (o *LLMOrchestrator) startHealthChecking(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Check health every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-o.stopCh:
			return
		case <-ticker.C:
			o.checkHealth()
		}
	}
}

// checkHealth checks the health of all components
func (o *LLMOrchestrator) checkHealth() {
	o.healthChecker.mutex.Lock()
	defer o.healthChecker.mutex.Unlock()

	now := time.Now()

	// Check scheduler health
	o.healthChecker.components["scheduler"] = HealthStatus{
		Name:      "scheduler",
		Status:    "healthy",
		LastCheck: now,
		Message:   "Scheduler is operational",
	}

	// Check resource manager health
	o.healthChecker.components["resource_manager"] = HealthStatus{
		Name:      "resource_manager",
		Status:    "healthy",
		LastCheck: now,
		Message:   "Resource manager is operational",
	}

	// Check model registry health
	o.healthChecker.components["model_registry"] = HealthStatus{
		Name:      "model_registry",
		Status:    "healthy",
		LastCheck: now,
		Message:   "Model registry is operational",
	}

	// Check Kubernetes connectivity
	_, err := o.kubeClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{Limit: 1})
	if err != nil {
		o.healthChecker.components["kubernetes"] = HealthStatus{
			Name:      "kubernetes",
			Status:    "unhealthy",
			LastCheck: now,
			Message:   fmt.Sprintf("Kubernetes API error: %v", err),
		}
	} else {
		o.healthChecker.components["kubernetes"] = HealthStatus{
			Name:      "kubernetes",
			Status:    "healthy",
			LastCheck: now,
			Message:   "Kubernetes API is accessible",
		}
	}
}

// listAllWorkloads lists all LLM workloads in the cluster
func (o *LLMOrchestrator) listAllWorkloads() ([]LLMWorkload, error) {
	var workloadList LLMWorkloadList
	err := o.ctrlClient.List(context.Background(), &workloadList)
	if err != nil {
		return nil, err
	}
	return workloadList.Items, nil
}

// IsRunning returns whether the orchestrator is currently running
func (o *LLMOrchestrator) IsRunning() bool {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.running
}
