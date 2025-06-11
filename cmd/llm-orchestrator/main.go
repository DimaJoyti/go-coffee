package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	llmorchestrator "github.com/DimaJoyti/go-coffee/internal/llm-orchestrator"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(llmorchestrator.AddToScheme(scheme))
}

func main() {
	var (
		configFile           = flag.String("config", "", "Path to configuration file")
		metricsAddr          = flag.String("metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
		probeAddr            = flag.String("health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
		enableLeaderElection = flag.Bool("leader-elect", false, "Enable leader election for controller manager.")
		leaderElectionID     = flag.String("leader-election-id", "llm-orchestrator-leader", "Leader election ID.")
		zapLogLevel          = flag.String("zap-log-level", "info", "Zap log level (debug, info, warn, error)")
		zapEncoder           = flag.String("zap-encoder", "json", "Zap log encoder (json, console)")
		namespace            = flag.String("namespace", "llm-orchestrator", "Namespace to watch for resources")
	)
	flag.Parse()

	// Setup logging
	logger := setupLogging(*zapLogLevel, *zapEncoder)
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Development: false,
		Level:       zapcore.InfoLevel,
	})))

	logger.Info("Starting LLM Orchestrator",
		zap.String("version", "1.0.0"),
		zap.String("namespace", *namespace),
		zap.String("metrics-addr", *metricsAddr),
		zap.String("probe-addr", *probeAddr),
		zap.Bool("leader-election", *enableLeaderElection),
	)

	// Load configuration
	config, err := loadConfig(*configFile)
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Override config with command line flags
	if *namespace != "" {
		config.Namespace = *namespace
	}
	if *metricsAddr != "" {
		config.MetricsPort = parsePort(*metricsAddr)
	}
	if *probeAddr != "" {
		config.HealthPort = parsePort(*probeAddr)
	}
	config.LeaderElection = *enableLeaderElection
	if *leaderElectionID != "" {
		config.LeaderLockName = *leaderElectionID
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Create and start the orchestrator
	orchestrator, err := llmorchestrator.NewLLMOrchestrator(logger, config)
	if err != nil {
		logger.Fatal("Failed to create orchestrator", zap.Error(err))
	}

	// Start the orchestrator
	if err := orchestrator.Start(ctx); err != nil {
		logger.Fatal("Failed to start orchestrator", zap.Error(err))
	}

	logger.Info("LLM Orchestrator started successfully")

	// Start health and metrics servers
	go startHealthServer(logger, *probeAddr, orchestrator)
	go startMetricsServer(logger, *metricsAddr, orchestrator)

	// Wait for shutdown signal
	<-sigCh
	logger.Info("Received shutdown signal, gracefully shutting down...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop the orchestrator
	if err := orchestrator.Stop(); err != nil {
		logger.Error("Error during orchestrator shutdown", zap.Error(err))
	}

	// Cancel the main context
	cancel()

	// Wait for shutdown to complete or timeout
	select {
	case <-shutdownCtx.Done():
		logger.Warn("Shutdown timeout exceeded")
	default:
		logger.Info("LLM Orchestrator shut down successfully")
	}
}

// setupLogging configures the logger
func setupLogging(level, encoder string) *zap.Logger {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)

	if encoder == "console" {
		config.Encoding = "console"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	return logger
}

// loadConfig loads configuration from file
func loadConfig(configFile string) (*llmorchestrator.OrchestratorConfig, error) {
	config := &llmorchestrator.OrchestratorConfig{
		// Set defaults
		Namespace:      "llm-orchestrator",
		MetricsEnabled: true,
		MetricsPort:    8080,
		HealthPort:     8081,
		WorkerThreads:  10,
		SyncInterval:   30 * time.Second,
		LeaderElection: false,
		LeaderLockName: "llm-orchestrator-leader",
		TLSEnabled:     false,

		// Component defaults
		Scheduler: &llmorchestrator.SchedulerConfig{
			Strategy:                  "performance-aware",
			ResourceOvercommitRatio:   1.2,
			GPUFragmentationThreshold: 0.8,
			LocalityPreference:        true,
			ModelAffinityWeight:       0.3,
			LatencyOptimization:       true,
			ThroughputOptimization:    true,
			ScaleUpCooldown:           3 * time.Minute,
			ScaleDownCooldown:         10 * time.Minute,
			PredictiveScaling:         true,
			QoSClasses:                make(map[string]llmorchestrator.QoSConfig),
		},

		ResourceManager: &llmorchestrator.ResourceManagerConfig{
			DefaultCPURequest:         "1000m",
			DefaultMemoryRequest:      "2Gi",
			DefaultGPURequest:         0,
			OptimizationInterval:      5 * time.Minute,
			ResourceUtilizationTarget: 0.8,
			OvercommitRatio:           1.2,
			ScaleUpThreshold:          0.8,
			ScaleDownThreshold:        0.3,
			ScaleUpCooldown:           3 * time.Minute,
			ScaleDownCooldown:         10 * time.Minute,
			MaxCPUPerWorkload:         "8000m",
			MaxMemoryPerWorkload:      "32Gi",
			MaxGPUPerWorkload:         4,
			QoSResourceAllocation:     make(map[string]llmorchestrator.QoSResourceConfig),
		},

		ModelRegistry: &llmorchestrator.ModelRegistryConfig{
			StorageType:          "local",
			StoragePath:          "/var/lib/llm-models",
			CacheSize:            100 * 1024 * 1024 * 1024, // 100GB
			CacheEnabled:         true,
			MaxVersions:          10,
			AutoCleanup:          true,
			CleanupInterval:      "24h",
			MetricsEnabled:       true,
			BenchmarkOnRegister:  true,
			PerformanceThreshold: 0.8,
			ChecksumValidation:   true,
			SignatureValidation:  false,
		},

		Operator: &llmorchestrator.OperatorConfig{
			ReconcileInterval:       30 * time.Second,
			MaxConcurrentReconciles: 10,
			DefaultNamespace:        "llm-workloads",
			DefaultImage:            "llm-server:latest",
			DefaultServiceAccount:   "llm-orchestrator",
			DefaultCPURequest:       "500m",
			DefaultMemoryRequest:    "1Gi",
			DefaultCPULimit:         "2000m",
			DefaultMemoryLimit:      "4Gi",
			MetricsEnabled:          true,
			HealthCheckEnabled:      true,
		},
	}

	if configFile == "" {
		return config, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// parsePort extracts port number from address
func parsePort(addr string) int {
	if addr[0] == ':' {
		addr = addr[1:]
	}

	var port int
	fmt.Sscanf(addr, "%d", &port)
	return port
}

// startHealthServer starts the health check server
func startHealthServer(logger *zap.Logger, addr string, orchestrator *llmorchestrator.LLMOrchestrator) {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     "0", // Disable metrics on this manager
		HealthProbeBindAddress: addr,
		LeaderElection:         false,
	})
	if err != nil {
		logger.Error("Failed to create health manager", zap.Error(err))
		return
	}

	// Add health checks
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		logger.Error("Failed to add health check", zap.Error(err))
		return
	}

	if err := mgr.AddReadyzCheck("readyz", func(req *http.Request) error {
		if !orchestrator.IsRunning() {
			return fmt.Errorf("orchestrator is not running")
		}
		return nil
	}); err != nil {
		logger.Error("Failed to add readiness check", zap.Error(err))
		return
	}

	logger.Info("Starting health server", zap.String("address", addr))
	if err := mgr.Start(context.Background()); err != nil {
		logger.Error("Health server failed", zap.Error(err))
	}
}

// startMetricsServer starts the metrics server
func startMetricsServer(logger *zap.Logger, addr string, orchestrator *llmorchestrator.LLMOrchestrator) {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     addr,
		HealthProbeBindAddress: "0", // Disable health on this manager
		LeaderElection:         false,
	})
	if err != nil {
		logger.Error("Failed to create metrics manager", zap.Error(err))
		return
	}

	logger.Info("Starting metrics server", zap.String("address", addr))
	if err := mgr.Start(context.Background()); err != nil {
		logger.Error("Metrics server failed", zap.Error(err))
	}
}
