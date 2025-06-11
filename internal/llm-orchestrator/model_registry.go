package llmorchestrator

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// ModelRegistry manages LLM model artifacts, versions, and metadata
type ModelRegistry struct {
	logger      *zap.Logger
	config      *ModelRegistryConfig
	storage     ModelStorage
	cache       *ModelCache
	versioning  *ModelVersioning
	performance *ModelPerformanceTracker
	mutex       sync.RWMutex
}

// ModelRegistryConfig defines model registry configuration
type ModelRegistryConfig struct {
	// Storage configuration
	StorageType     string `yaml:"storageType"`     // local, s3, gcs, azure
	StoragePath     string `yaml:"storagePath"`     // base path for model storage
	CacheSize       int64  `yaml:"cacheSize"`       // cache size in bytes
	CacheEnabled    bool   `yaml:"cacheEnabled"`    // enable model caching
	
	// Versioning configuration
	MaxVersions     int    `yaml:"maxVersions"`     // maximum versions to keep
	AutoCleanup     bool   `yaml:"autoCleanup"`     // auto cleanup old versions
	CleanupInterval string `yaml:"cleanupInterval"` // cleanup interval
	
	// Performance tracking
	MetricsEnabled       bool   `yaml:"metricsEnabled"`       // enable performance metrics
	BenchmarkOnRegister  bool   `yaml:"benchmarkOnRegister"`  // run benchmarks on registration
	PerformanceThreshold float64 `yaml:"performanceThreshold"` // minimum performance threshold
	
	// Security
	ChecksumValidation bool `yaml:"checksumValidation"` // validate model checksums
	SignatureValidation bool `yaml:"signatureValidation"` // validate model signatures
}

// ModelMetadata contains comprehensive model information
type ModelMetadata struct {
	// Basic information
	Name        string    `yaml:"name" json:"name"`
	Version     string    `yaml:"version" json:"version"`
	Description string    `yaml:"description" json:"description"`
	Author      string    `yaml:"author" json:"author"`
	License     string    `yaml:"license" json:"license"`
	CreatedAt   time.Time `yaml:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `yaml:"updatedAt" json:"updatedAt"`
	
	// Model characteristics
	Type         string            `yaml:"type" json:"type"`         // text-generation, embedding, classification
	Architecture string            `yaml:"architecture" json:"architecture"` // transformer, lstm, cnn
	Framework    string            `yaml:"framework" json:"framework"`    // pytorch, tensorflow, onnx
	Size         ModelSize         `yaml:"size" json:"size"`
	Parameters   ModelParameters   `yaml:"parameters" json:"parameters"`
	Capabilities []string          `yaml:"capabilities" json:"capabilities"`
	Languages    []string          `yaml:"languages" json:"languages"`
	
	// Technical specifications
	InputFormat  string            `yaml:"inputFormat" json:"inputFormat"`   // text, tokens, embeddings
	OutputFormat string            `yaml:"outputFormat" json:"outputFormat"` // text, tokens, probabilities
	MaxTokens    int32             `yaml:"maxTokens" json:"maxTokens"`
	ContextLength int32            `yaml:"contextLength" json:"contextLength"`
	Vocabulary   int32             `yaml:"vocabulary" json:"vocabulary"`
	
	// Resource requirements
	MinCPU       string            `yaml:"minCPU" json:"minCPU"`
	MinMemory    string            `yaml:"minMemory" json:"minMemory"`
	MinGPU       int32             `yaml:"minGPU" json:"minGPU"`
	RecommendedCPU    string       `yaml:"recommendedCPU" json:"recommendedCPU"`
	RecommendedMemory string       `yaml:"recommendedMemory" json:"recommendedMemory"`
	RecommendedGPU    int32        `yaml:"recommendedGPU" json:"recommendedGPU"`
	
	// Performance metrics
	Benchmarks   map[string]BenchmarkResult `yaml:"benchmarks" json:"benchmarks"`
	Accuracy     float64                    `yaml:"accuracy" json:"accuracy"`
	Latency      time.Duration              `yaml:"latency" json:"latency"`
	Throughput   float64                    `yaml:"throughput" json:"throughput"` // tokens/second
	
	// File information
	Files        []ModelFile       `yaml:"files" json:"files"`
	TotalSize    int64             `yaml:"totalSize" json:"totalSize"`
	Checksum     string            `yaml:"checksum" json:"checksum"`
	Signature    string            `yaml:"signature,omitempty" json:"signature,omitempty"`
	
	// Deployment information
	Status       ModelStatus       `yaml:"status" json:"status"`
	Deployments  []ModelDeployment `yaml:"deployments" json:"deployments"`
	UsageStats   ModelUsageStats   `yaml:"usageStats" json:"usageStats"`
	
	// Tags and labels
	Tags         []string          `yaml:"tags" json:"tags"`
	Labels       map[string]string `yaml:"labels" json:"labels"`
}

// ModelSize represents model size information
type ModelSize struct {
	Category    string `yaml:"category" json:"category"` // small, medium, large, xlarge
	Parameters  int64  `yaml:"parameters" json:"parameters"` // number of parameters
	SizeBytes   int64  `yaml:"sizeBytes" json:"sizeBytes"`   // total size in bytes
	Quantization string `yaml:"quantization,omitempty" json:"quantization,omitempty"` // fp32, fp16, int8, int4
}

// ModelParameters contains model-specific parameters
type ModelParameters struct {
	Temperature     float64           `yaml:"temperature,omitempty" json:"temperature,omitempty"`
	TopP           float64           `yaml:"topP,omitempty" json:"topP,omitempty"`
	TopK           int32             `yaml:"topK,omitempty" json:"topK,omitempty"`
	MaxNewTokens   int32             `yaml:"maxNewTokens,omitempty" json:"maxNewTokens,omitempty"`
	RepetitionPenalty float64        `yaml:"repetitionPenalty,omitempty" json:"repetitionPenalty,omitempty"`
	Custom         map[string]interface{} `yaml:"custom,omitempty" json:"custom,omitempty"`
}

// ModelFile represents a model file
type ModelFile struct {
	Name     string `yaml:"name" json:"name"`
	Path     string `yaml:"path" json:"path"`
	Size     int64  `yaml:"size" json:"size"`
	Checksum string `yaml:"checksum" json:"checksum"`
	Type     string `yaml:"type" json:"type"` // model, config, tokenizer, vocab
}

// ModelStatus represents the current status of a model
type ModelStatus string

const (
	ModelStatusRegistering ModelStatus = "registering"
	ModelStatusAvailable   ModelStatus = "available"
	ModelStatusDeprecated  ModelStatus = "deprecated"
	ModelStatusArchived    ModelStatus = "archived"
	ModelStatusFailed      ModelStatus = "failed"
)

// ModelDeployment represents a model deployment
type ModelDeployment struct {
	ID          string            `yaml:"id" json:"id"`
	Environment string            `yaml:"environment" json:"environment"` // dev, staging, prod
	Namespace   string            `yaml:"namespace" json:"namespace"`
	Replicas    int32             `yaml:"replicas" json:"replicas"`
	Status      string            `yaml:"status" json:"status"`
	CreatedAt   time.Time         `yaml:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time         `yaml:"updatedAt" json:"updatedAt"`
	Endpoints   []string          `yaml:"endpoints" json:"endpoints"`
	Metrics     DeploymentMetrics `yaml:"metrics" json:"metrics"`
}

// DeploymentMetrics contains deployment performance metrics
type DeploymentMetrics struct {
	RequestsPerSecond float64       `yaml:"requestsPerSecond" json:"requestsPerSecond"`
	AverageLatency    time.Duration `yaml:"averageLatency" json:"averageLatency"`
	ErrorRate         float64       `yaml:"errorRate" json:"errorRate"`
	Uptime            time.Duration `yaml:"uptime" json:"uptime"`
	LastUpdated       time.Time     `yaml:"lastUpdated" json:"lastUpdated"`
}

// ModelUsageStats contains model usage statistics
type ModelUsageStats struct {
	TotalRequests     int64     `yaml:"totalRequests" json:"totalRequests"`
	TotalTokens       int64     `yaml:"totalTokens" json:"totalTokens"`
	AverageLatency    time.Duration `yaml:"averageLatency" json:"averageLatency"`
	ErrorRate         float64   `yaml:"errorRate" json:"errorRate"`
	PopularityScore   float64   `yaml:"popularityScore" json:"popularityScore"`
	LastUsed          time.Time `yaml:"lastUsed" json:"lastUsed"`
	UsageByDay        map[string]int64 `yaml:"usageByDay" json:"usageByDay"`
}

// BenchmarkResult contains benchmark test results
type BenchmarkResult struct {
	TestName    string        `yaml:"testName" json:"testName"`
	Score       float64       `yaml:"score" json:"score"`
	Unit        string        `yaml:"unit" json:"unit"`
	Latency     time.Duration `yaml:"latency" json:"latency"`
	Throughput  float64       `yaml:"throughput" json:"throughput"`
	Accuracy    float64       `yaml:"accuracy" json:"accuracy"`
	RunAt       time.Time     `yaml:"runAt" json:"runAt"`
	Environment string        `yaml:"environment" json:"environment"`
}

// ModelStorage interface for different storage backends
type ModelStorage interface {
	Store(ctx context.Context, modelID string, files []ModelFile) error
	Retrieve(ctx context.Context, modelID string) ([]ModelFile, error)
	Delete(ctx context.Context, modelID string) error
	Exists(ctx context.Context, modelID string) (bool, error)
	GetSize(ctx context.Context, modelID string) (int64, error)
}

// ModelCache handles model caching
type ModelCache struct {
	cache     map[string]*ModelMetadata
	usage     map[string]time.Time
	maxSize   int64
	currentSize int64
	mutex     sync.RWMutex
}

// ModelVersioning handles model version management
type ModelVersioning struct {
	versions map[string][]string // modelName -> versions
	latest   map[string]string   // modelName -> latest version
	mutex    sync.RWMutex
}

// ModelPerformanceTracker tracks model performance across deployments
type ModelPerformanceTracker struct {
	metrics map[string]*PerformanceHistory
	mutex   sync.RWMutex
}

// PerformanceHistory contains historical performance data
type PerformanceHistory struct {
	ModelID     string
	Metrics     []PerformanceDataPoint
	Trends      PerformanceTrends
	LastUpdated time.Time
}

// PerformanceDataPoint represents a single performance measurement
type PerformanceDataPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	Latency     float64   `json:"latency"`     // milliseconds
	Throughput  float64   `json:"throughput"`  // tokens/second
	Accuracy    float64   `json:"accuracy"`    // 0-1
	ErrorRate   float64   `json:"errorRate"`   // 0-1
	CPUUsage    float64   `json:"cpuUsage"`    // 0-1
	MemoryUsage float64   `json:"memoryUsage"` // 0-1
	GPUUsage    float64   `json:"gpuUsage"`    // 0-1
}

// PerformanceTrends contains performance trend analysis
type PerformanceTrends struct {
	LatencyTrend    string  `json:"latencyTrend"`    // improving, stable, degrading
	ThroughputTrend string  `json:"throughputTrend"` // improving, stable, degrading
	AccuracyTrend   string  `json:"accuracyTrend"`   // improving, stable, degrading
	OverallScore    float64 `json:"overallScore"`    // 0-100
}

// NewModelRegistry creates a new model registry instance
func NewModelRegistry(logger *zap.Logger, config *ModelRegistryConfig) (*ModelRegistry, error) {
	// Initialize storage backend
	storage, err := createStorageBackend(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage backend: %w", err)
	}

	// Initialize cache
	cache := &ModelCache{
		cache:   make(map[string]*ModelMetadata),
		usage:   make(map[string]time.Time),
		maxSize: config.CacheSize,
	}

	// Initialize versioning
	versioning := &ModelVersioning{
		versions: make(map[string][]string),
		latest:   make(map[string]string),
	}

	// Initialize performance tracker
	performance := &ModelPerformanceTracker{
		metrics: make(map[string]*PerformanceHistory),
	}

	return &ModelRegistry{
		logger:      logger,
		config:      config,
		storage:     storage,
		cache:       cache,
		versioning:  versioning,
		performance: performance,
	}, nil
}

// RegisterModel registers a new model in the registry
func (mr *ModelRegistry) RegisterModel(ctx context.Context, metadata *ModelMetadata, modelFiles []ModelFile) error {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	mr.logger.Info("Registering model",
		zap.String("name", metadata.Name),
		zap.String("version", metadata.Version),
		zap.String("type", metadata.Type),
	)

	// Validate model metadata
	if err := mr.validateMetadata(metadata); err != nil {
		return fmt.Errorf("invalid model metadata: %w", err)
	}

	// Calculate checksums if validation is enabled
	if mr.config.ChecksumValidation {
		if err := mr.calculateChecksums(modelFiles); err != nil {
			return fmt.Errorf("failed to calculate checksums: %w", err)
		}
	}

	// Store model files
	modelID := mr.generateModelID(metadata.Name, metadata.Version)
	if err := mr.storage.Store(ctx, modelID, modelFiles); err != nil {
		return fmt.Errorf("failed to store model files: %w", err)
	}

	// Update metadata with file information
	metadata.Files = modelFiles
	metadata.TotalSize = mr.calculateTotalSize(modelFiles)
	metadata.Status = ModelStatusRegistering
	metadata.CreatedAt = time.Now()
	metadata.UpdatedAt = time.Now()

	// Run benchmarks if enabled
	if mr.config.BenchmarkOnRegister {
		benchmarks, err := mr.runBenchmarks(ctx, metadata)
		if err != nil {
			mr.logger.Warn("Failed to run benchmarks", zap.Error(err))
		} else {
			metadata.Benchmarks = benchmarks
		}
	}

	// Update versioning
	mr.versioning.addVersion(metadata.Name, metadata.Version)

	// Cache the model
	if mr.config.CacheEnabled {
		mr.cache.add(modelID, metadata)
	}

	// Update status to available
	metadata.Status = ModelStatusAvailable

	mr.logger.Info("Successfully registered model",
		zap.String("modelID", modelID),
		zap.Int64("totalSize", metadata.TotalSize),
	)

	return nil
}

// GetModel retrieves model metadata
func (mr *ModelRegistry) GetModel(ctx context.Context, name, version string) (*ModelMetadata, error) {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	modelID := mr.generateModelID(name, version)

	// Check cache first
	if mr.config.CacheEnabled {
		if metadata := mr.cache.get(modelID); metadata != nil {
			return metadata, nil
		}
	}

	// Load from storage
	metadata, err := mr.loadMetadata(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to load model metadata: %w", err)
	}

	// Update cache
	if mr.config.CacheEnabled {
		mr.cache.add(modelID, metadata)
	}

	return metadata, nil
}

// ListModels returns a list of all registered models
func (mr *ModelRegistry) ListModels(ctx context.Context) ([]*ModelMetadata, error) {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	var models []*ModelMetadata

	// Get all models from versioning
	for modelName, versions := range mr.versioning.versions {
		for _, version := range versions {
			metadata, err := mr.GetModel(ctx, modelName, version)
			if err != nil {
				mr.logger.Warn("Failed to load model",
					zap.String("name", modelName),
					zap.String("version", version),
					zap.Error(err),
				)
				continue
			}
			models = append(models, metadata)
		}
	}

	return models, nil
}

// Helper methods
func (mr *ModelRegistry) generateModelID(name, version string) string {
	return fmt.Sprintf("%s:%s", name, version)
}

func (mr *ModelRegistry) validateMetadata(metadata *ModelMetadata) error {
	if metadata.Name == "" {
		return fmt.Errorf("model name is required")
	}
	if metadata.Version == "" {
		return fmt.Errorf("model version is required")
	}
	if metadata.Type == "" {
		return fmt.Errorf("model type is required")
	}
	return nil
}

func (mr *ModelRegistry) calculateChecksums(files []ModelFile) error {
	for i := range files {
		checksum, err := calculateFileChecksum(files[i].Path)
		if err != nil {
			return fmt.Errorf("failed to calculate checksum for %s: %w", files[i].Name, err)
		}
		files[i].Checksum = checksum
	}
	return nil
}

func (mr *ModelRegistry) calculateTotalSize(files []ModelFile) int64 {
	var total int64
	for _, file := range files {
		total += file.Size
	}
	return total
}

func (mr *ModelRegistry) runBenchmarks(ctx context.Context, metadata *ModelMetadata) (map[string]BenchmarkResult, error) {
	// Placeholder for benchmark implementation
	benchmarks := make(map[string]BenchmarkResult)
	
	// Example benchmark
	benchmarks["latency_test"] = BenchmarkResult{
		TestName:   "Latency Test",
		Score:      100.0,
		Unit:       "ms",
		Latency:    100 * time.Millisecond,
		Throughput: 50.0,
		Accuracy:   0.95,
		RunAt:      time.Now(),
		Environment: "test",
	}
	
	return benchmarks, nil
}

func (mr *ModelRegistry) loadMetadata(ctx context.Context, modelID string) (*ModelMetadata, error) {
	// Placeholder for loading metadata from storage
	return nil, fmt.Errorf("not implemented")
}

// Versioning methods
func (mv *ModelVersioning) addVersion(modelName, version string) {
	mv.mutex.Lock()
	defer mv.mutex.Unlock()

	if mv.versions[modelName] == nil {
		mv.versions[modelName] = []string{}
	}
	
	mv.versions[modelName] = append(mv.versions[modelName], version)
	mv.latest[modelName] = version
}

// Cache methods
func (mc *ModelCache) add(modelID string, metadata *ModelMetadata) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.cache[modelID] = metadata
	mc.usage[modelID] = time.Now()
}

func (mc *ModelCache) get(modelID string) *ModelMetadata {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	if metadata, exists := mc.cache[modelID]; exists {
		mc.usage[modelID] = time.Now()
		return metadata
	}
	return nil
}

// Utility functions
func calculateFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func createStorageBackend(config *ModelRegistryConfig) (ModelStorage, error) {
	switch config.StorageType {
	case "local":
		return &LocalStorage{basePath: config.StoragePath}, nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.StorageType)
	}
}

// LocalStorage implements ModelStorage for local filesystem
type LocalStorage struct {
	basePath string
}

func (ls *LocalStorage) Store(ctx context.Context, modelID string, files []ModelFile) error {
	modelDir := filepath.Join(ls.basePath, modelID)
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return err
	}
	
	// Copy files to model directory
	for _, file := range files {
		destPath := filepath.Join(modelDir, file.Name)
		if err := copyFile(file.Path, destPath); err != nil {
			return fmt.Errorf("failed to copy file %s: %w", file.Name, err)
		}
	}
	
	return nil
}

func (ls *LocalStorage) Retrieve(ctx context.Context, modelID string) ([]ModelFile, error) {
	// Implementation for retrieving files
	return nil, fmt.Errorf("not implemented")
}

func (ls *LocalStorage) Delete(ctx context.Context, modelID string) error {
	modelDir := filepath.Join(ls.basePath, modelID)
	return os.RemoveAll(modelDir)
}

func (ls *LocalStorage) Exists(ctx context.Context, modelID string) (bool, error) {
	modelDir := filepath.Join(ls.basePath, modelID)
	_, err := os.Stat(modelDir)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func (ls *LocalStorage) GetSize(ctx context.Context, modelID string) (int64, error) {
	modelDir := filepath.Join(ls.basePath, modelID)
	var size int64
	
	err := filepath.Walk(modelDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	
	return size, err
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
