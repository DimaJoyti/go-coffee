package ml

import (
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/performance/latency"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// LatencyPredictor provides ML-based latency prediction and optimization
type LatencyPredictor struct {
	// Model configuration
	config        *MLConfig
	isInitialized int32 // atomic bool
	isTraining    int32 // atomic bool

	// Neural network models
	latencyModel      *NeuralNetwork
	optimizationModel *NeuralNetwork
	anomalyModel      *NeuralNetwork

	// Training data
	trainingData   []TrainingExample
	validationData []TrainingExample
	testData       []TrainingExample
	dataBuffer     *CircularBuffer

	// Feature extractors
	featureExtractor *FeatureExtractor
	featureScaler    *FeatureScaler

	// Model performance
	modelAccuracy    float64
	predictionError  float64
	lastTrainingTime int64
	trainingCount    uint64
	predictionCount  uint64

	// Real-time prediction
	predictionCache sync.Map // map[string]*LatencyPrediction
	cacheExpiry     time.Duration

	// Optimization recommendations
	optimizationQueue chan *OptimizationRecommendation

	// Observability
	tracer         trace.Tracer
	latencyTracker *latency.LatencyTracker

	// Worker control
	workers  sync.WaitGroup
	stopChan chan struct{}
	mutex    sync.RWMutex
}

// MLConfig holds machine learning configuration
type MLConfig struct {
	// Model architecture
	HiddenLayers   []int   // Hidden layer sizes
	ActivationFunc string  // Activation function (relu, sigmoid, tanh)
	LearningRate   float64 // Learning rate
	BatchSize      int     // Training batch size
	Epochs         int     // Training epochs

	// Data configuration
	FeatureWindow     int     // Feature window size
	PredictionHorizon int     // Prediction horizon (microseconds)
	TrainingDataSize  int     // Training dataset size
	ValidationSplit   float64 // Validation split ratio

	// Feature engineering
	EnableTimeFeatures    bool // Enable time-based features
	EnableSystemFeatures  bool // Enable system metrics features
	EnableMarketFeatures  bool // Enable market data features
	EnableNetworkFeatures bool // Enable network features

	// Model optimization
	EnableEarlyStopping   bool    // Enable early stopping
	EarlyStoppingPatience int     // Early stopping patience
	EnableRegularization  bool    // Enable L2 regularization
	RegularizationLambda  float64 // Regularization strength

	// Real-time configuration
	PredictionCacheSize int           // Prediction cache size
	CacheExpiry         time.Duration // Cache expiry time
	RetrainingInterval  time.Duration // Model retraining interval
	MinTrainingExamples int           // Minimum examples for training

	// Performance tuning
	WorkerThreads         int  // Number of worker threads
	EnableGPUAcceleration bool // Enable GPU acceleration
	BatchPrediction       bool // Enable batch prediction
}

// TrainingExample represents a training example
type TrainingExample struct {
	Features  []float64              // Input features
	Target    float64                // Target latency (microseconds)
	Timestamp int64                  // Example timestamp
	Weight    float64                // Example weight
	Metadata  map[string]interface{} // Additional metadata
}

// LatencyPrediction represents a latency prediction
type LatencyPrediction struct {
	PredictedLatency float64   // Predicted latency (microseconds)
	Confidence       float64   // Prediction confidence (0-1)
	Features         []float64 // Input features used
	ModelVersion     int       // Model version used
	Timestamp        int64     // Prediction timestamp
	ExpiresAt        int64     // Cache expiry timestamp
}

// OptimizationRecommendation represents an optimization recommendation
type OptimizationRecommendation struct {
	Type        string                 // Recommendation type
	Description string                 // Human-readable description
	Impact      float64                // Expected latency improvement (microseconds)
	Confidence  float64                // Recommendation confidence
	Parameters  map[string]interface{} // Optimization parameters
	Priority    int                    // Priority (1-10)
	Timestamp   int64                  // Recommendation timestamp
}

// NeuralNetwork represents a simple neural network
type NeuralNetwork struct {
	layers       []*Layer
	weights      [][][]float64 // weights[layer][neuron][input]
	biases       [][]float64   // biases[layer][neuron]
	learningRate float64
	activation   ActivationFunc
	loss         LossFunc
}

// Layer represents a neural network layer
type Layer struct {
	size       int
	activation ActivationFunc
}

// ActivationFunc represents an activation function
type ActivationFunc func(x float64) float64

// LossFunc represents a loss function
type LossFunc func(predicted, actual []float64) float64

// FeatureExtractor extracts features for ML models
type FeatureExtractor struct {
	config         *MLConfig
	systemMetrics  *SystemMetricsCollector
	marketMetrics  *MarketMetricsCollector
	networkMetrics *NetworkMetricsCollector
	timeFeatures   *TimeFeatureExtractor
}

// FeatureScaler normalizes features
type FeatureScaler struct {
	means  []float64
	stds   []float64
	mins   []float64
	maxs   []float64
	method string // "standardize" or "normalize"
}

// CircularBuffer implements a circular buffer for training data
type CircularBuffer struct {
	data  []TrainingExample
	size  int
	head  int
	tail  int
	count int
	mutex sync.RWMutex
}

// NewLatencyPredictor creates a new latency predictor
func NewLatencyPredictor(config *MLConfig) (*LatencyPredictor, error) {
	predictor := &LatencyPredictor{
		config:            config,
		cacheExpiry:       config.CacheExpiry,
		optimizationQueue: make(chan *OptimizationRecommendation, 1000),
		tracer:            otel.Tracer("hft.ml.latency_predictor"),
		latencyTracker:    latency.GetGlobalLatencyTracker(),
		stopChan:          make(chan struct{}),
	}

	// Initialize data buffer
	predictor.dataBuffer = NewCircularBuffer(config.TrainingDataSize)

	// Initialize feature extractor
	predictor.featureExtractor = NewFeatureExtractor(config)

	// Initialize feature scaler
	predictor.featureScaler = NewFeatureScaler("standardize")

	return predictor, nil
}

// Initialize initializes the latency predictor
func (lp *LatencyPredictor) Initialize() error {
	_, span := lp.tracer.Start(context.Background(), "LatencyPredictor.Initialize")
	defer span.End()

	// Initialize neural network models
	if err := lp.initializeModels(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to initialize models: %w", err)
	}

	// Load pre-trained models if available
	if err := lp.loadPretrainedModels(); err != nil {
		// Log warning but don't fail initialization
		span.SetAttributes(attribute.String("warning", "failed to load pretrained models"))
	}

	// Initialize feature extractor
	if err := lp.featureExtractor.Initialize(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to initialize feature extractor: %w", err)
	}

	atomic.StoreInt32(&lp.isInitialized, 1)

	span.SetAttributes(
		attribute.Int("hidden_layers", len(lp.config.HiddenLayers)),
		attribute.Float64("learning_rate", lp.config.LearningRate),
		attribute.Int("batch_size", lp.config.BatchSize),
		attribute.Bool("gpu_acceleration", lp.config.EnableGPUAcceleration),
	)

	return nil
}

// Start starts the latency predictor
func (lp *LatencyPredictor) Start() error {
	if atomic.LoadInt32(&lp.isInitialized) == 0 {
		return fmt.Errorf("latency predictor not initialized")
	}

	_, span := lp.tracer.Start(context.Background(), "LatencyPredictor.Start")
	defer span.End()

	// Start data collection worker
	lp.workers.Add(1)
	go lp.dataCollectionWorker()

	// Start training worker
	lp.workers.Add(1)
	go lp.trainingWorker()

	// Start optimization worker
	lp.workers.Add(1)
	go lp.optimizationWorker()

	// Start cache cleanup worker
	lp.workers.Add(1)
	go lp.cacheCleanupWorker()

	span.SetAttributes(attribute.Bool("started", true))
	return nil
}

// Stop stops the latency predictor
func (lp *LatencyPredictor) Stop() error {
	_, span := lp.tracer.Start(context.Background(), "LatencyPredictor.Stop")
	defer span.End()

	// Signal workers to stop
	close(lp.stopChan)

	// Wait for workers to finish
	lp.workers.Wait()

	// Save trained models
	if err := lp.saveModels(); err != nil {
		span.RecordError(err)
		// Log error but don't fail shutdown
	}

	span.SetAttributes(attribute.Bool("stopped", true))
	return nil
}

// PredictLatency predicts latency for given system state
func (lp *LatencyPredictor) PredictLatency(ctx context.Context, systemState *SystemState) (*LatencyPrediction, error) {
	if atomic.LoadInt32(&lp.isInitialized) == 0 {
		return nil, fmt.Errorf("latency predictor not initialized")
	}

	// Generate cache key
	cacheKey := lp.generateCacheKey(systemState)

	// Check cache first
	if cached, exists := lp.predictionCache.Load(cacheKey); exists {
		prediction := cached.(*LatencyPrediction)
		if time.Now().UnixNano() < prediction.ExpiresAt {
			atomic.AddUint64(&lp.predictionCount, 1)
			return prediction, nil
		}
		// Cache expired, remove it
		lp.predictionCache.Delete(cacheKey)
	}

	// Extract features
	features, err := lp.featureExtractor.ExtractFeatures(systemState)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features: %w", err)
	}

	// Scale features
	scaledFeatures := lp.featureScaler.Transform(features)

	// Make prediction
	prediction := lp.latencyModel.Predict(scaledFeatures)
	confidence := lp.calculatePredictionConfidence(scaledFeatures, prediction)

	result := &LatencyPrediction{
		PredictedLatency: prediction[0],
		Confidence:       confidence,
		Features:         scaledFeatures,
		ModelVersion:     1, // TODO: track model versions
		Timestamp:        time.Now().UnixNano(),
		ExpiresAt:        time.Now().Add(lp.cacheExpiry).UnixNano(),
	}

	// Cache the prediction
	lp.predictionCache.Store(cacheKey, result)

	atomic.AddUint64(&lp.predictionCount, 1)
	return result, nil
}

// AddTrainingExample adds a training example
func (lp *LatencyPredictor) AddTrainingExample(features []float64, actualLatency float64, metadata map[string]interface{}) {
	example := TrainingExample{
		Features:  features,
		Target:    actualLatency,
		Timestamp: time.Now().UnixNano(),
		Weight:    1.0,
		Metadata:  metadata,
	}

	lp.dataBuffer.Add(example)
}

// GetOptimizationRecommendations returns optimization recommendations
func (lp *LatencyPredictor) GetOptimizationRecommendations() []*OptimizationRecommendation {
	recommendations := make([]*OptimizationRecommendation, 0)

	// Drain the optimization queue
	for {
		select {
		case rec := <-lp.optimizationQueue:
			recommendations = append(recommendations, rec)
		default:
			return recommendations
		}
	}
}

// GetStatistics returns predictor statistics
func (lp *LatencyPredictor) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"model_accuracy":     lp.modelAccuracy,
		"prediction_error":   lp.predictionError,
		"training_count":     atomic.LoadUint64(&lp.trainingCount),
		"prediction_count":   atomic.LoadUint64(&lp.predictionCount),
		"last_training_time": atomic.LoadInt64(&lp.lastTrainingTime),
		"training_data_size": lp.dataBuffer.Count(),
		"cache_size":         lp.getCacheSize(),
		"is_training":        atomic.LoadInt32(&lp.isTraining) == 1,
		"is_initialized":     atomic.LoadInt32(&lp.isInitialized) == 1,
	}
}

// Worker functions

func (lp *LatencyPredictor) dataCollectionWorker() {
	defer lp.workers.Done()

	ticker := time.NewTicker(100 * time.Millisecond) // Collect data every 100ms
	defer ticker.Stop()

	for {
		select {
		case <-lp.stopChan:
			return
		case <-ticker.C:
			lp.collectTrainingData()
		}
	}
}

func (lp *LatencyPredictor) trainingWorker() {
	defer lp.workers.Done()

	ticker := time.NewTicker(lp.config.RetrainingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-lp.stopChan:
			return
		case <-ticker.C:
			if lp.dataBuffer.Count() >= lp.config.MinTrainingExamples {
				lp.trainModels()
			}
		}
	}
}

func (lp *LatencyPredictor) optimizationWorker() {
	defer lp.workers.Done()

	ticker := time.NewTicker(10 * time.Second) // Generate recommendations every 10s
	defer ticker.Stop()

	for {
		select {
		case <-lp.stopChan:
			return
		case <-ticker.C:
			lp.generateOptimizationRecommendations()
		}
	}
}

func (lp *LatencyPredictor) cacheCleanupWorker() {
	defer lp.workers.Done()

	ticker := time.NewTicker(1 * time.Minute) // Cleanup every minute
	defer ticker.Stop()

	for {
		select {
		case <-lp.stopChan:
			return
		case <-ticker.C:
			lp.cleanupExpiredCache()
		}
	}
}

// Helper functions

func (lp *LatencyPredictor) initializeModels() error {
	// Initialize latency prediction model
	lp.latencyModel = NewNeuralNetwork(
		lp.config.HiddenLayers,
		lp.config.LearningRate,
		ReLU,
		MeanSquaredError,
	)

	// Initialize optimization model
	lp.optimizationModel = NewNeuralNetwork(
		[]int{64, 32, 16}, // Smaller network for optimization
		lp.config.LearningRate,
		ReLU,
		MeanSquaredError,
	)

	// Initialize anomaly detection model
	lp.anomalyModel = NewNeuralNetwork(
		[]int{32, 16, 8}, // Even smaller for anomaly detection
		lp.config.LearningRate,
		Sigmoid,
		BinaryCrossEntropy,
	)

	return nil
}

func (lp *LatencyPredictor) loadPretrainedModels() error {
	// This would load pre-trained models from disk
	// For now, this is a placeholder
	return nil
}

func (lp *LatencyPredictor) saveModels() error {
	// This would save trained models to disk
	// For now, this is a placeholder
	return nil
}

func (lp *LatencyPredictor) collectTrainingData() {
	// Collect current system state and recent latency measurements
	systemState := lp.featureExtractor.GetCurrentSystemState()
	features, err := lp.featureExtractor.ExtractFeatures(systemState)
	if err != nil {
		return
	}

	// Get recent latency measurements
	recentLatencies := lp.latencyTracker.GetAllStats()
	if len(recentLatencies) == 0 {
		return
	}

	// Create training examples from recent data
	for point, stats := range recentLatencies {
		if stats.Count > 0 {
			avgLatency := float64(stats.Sum) / float64(stats.Count) / 1000.0 // Convert to microseconds

			metadata := map[string]interface{}{
				"latency_point": string(point),
				"sample_count":  stats.Count,
			}

			lp.AddTrainingExample(features, avgLatency, metadata)
		}
	}
}

func (lp *LatencyPredictor) trainModels() {
	if atomic.CompareAndSwapInt32(&lp.isTraining, 0, 1) {
		defer atomic.StoreInt32(&lp.isTraining, 0)

		// Get training data
		trainingData := lp.dataBuffer.GetAll()
		if len(trainingData) < lp.config.MinTrainingExamples {
			return
		}

		// Split data
		lp.splitTrainingData(trainingData)

		// Train latency model
		lp.trainLatencyModel()

		// Train optimization model
		lp.trainOptimizationModel()

		// Update statistics
		atomic.StoreInt64(&lp.lastTrainingTime, time.Now().UnixNano())
		atomic.AddUint64(&lp.trainingCount, 1)
	}
}

func (lp *LatencyPredictor) trainLatencyModel() {
	// Implement neural network training
	// This is a simplified placeholder
	lp.modelAccuracy = 0.85   // Placeholder accuracy
	lp.predictionError = 0.15 // Placeholder error
}

func (lp *LatencyPredictor) trainOptimizationModel() {
	// Train model for optimization recommendations
	// This is a placeholder
}

func (lp *LatencyPredictor) splitTrainingData(data []TrainingExample) {
	// Split data into training, validation, and test sets
	validationSize := int(float64(len(data)) * lp.config.ValidationSplit)
	testSize := validationSize / 2

	lp.trainingData = data[:len(data)-validationSize-testSize]
	lp.validationData = data[len(data)-validationSize-testSize : len(data)-testSize]
	lp.testData = data[len(data)-testSize:]
}

func (lp *LatencyPredictor) generateOptimizationRecommendations() {
	// Generate optimization recommendations based on current system state
	// This is a placeholder implementation

	recommendation := &OptimizationRecommendation{
		Type:        "cpu_affinity",
		Description: "Optimize CPU affinity for better cache locality",
		Impact:      5.0, // 5 microseconds improvement
		Confidence:  0.8,
		Parameters: map[string]interface{}{
			"core_id": 2,
			"thread":  "market_data_processor",
		},
		Priority:  7,
		Timestamp: time.Now().UnixNano(),
	}

	select {
	case lp.optimizationQueue <- recommendation:
	default:
		// Queue full, drop recommendation
	}
}

func (lp *LatencyPredictor) generateCacheKey(systemState *SystemState) string {
	// Generate a cache key based on system state
	// This is a simplified implementation
	return fmt.Sprintf("state_%d_%d_%d",
		systemState.CPUUsage,
		systemState.MemoryUsage,
		systemState.NetworkLoad)
}

func (lp *LatencyPredictor) calculatePredictionConfidence(features []float64, prediction []float64) float64 {
	// Calculate prediction confidence based on model uncertainty
	// This is a simplified implementation
	return 0.9 // Placeholder confidence
}

func (lp *LatencyPredictor) getCacheSize() int {
	count := 0
	lp.predictionCache.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

func (lp *LatencyPredictor) cleanupExpiredCache() {
	now := time.Now().UnixNano()

	lp.predictionCache.Range(func(key, value interface{}) bool {
		prediction := value.(*LatencyPrediction)
		if now > prediction.ExpiresAt {
			lp.predictionCache.Delete(key)
		}
		return true
	})
}

// SystemState represents the current system state
type SystemState struct {
	CPUUsage      int64
	MemoryUsage   int64
	NetworkLoad   int64
	DiskIO        int64
	QueueLengths  map[string]int
	ActiveThreads int
	Timestamp     int64
}

// Placeholder implementations for neural network components
// In a real implementation, these would be more sophisticated

func NewNeuralNetwork(hiddenLayers []int, learningRate float64, activation ActivationFunc, loss LossFunc) *NeuralNetwork {
	return &NeuralNetwork{
		learningRate: learningRate,
		activation:   activation,
		loss:         loss,
	}
}

func (nn *NeuralNetwork) Predict(input []float64) []float64 {
	// Placeholder prediction
	return []float64{100.0} // 100 microseconds
}

func ReLU(x float64) float64 {
	return math.Max(0, x)
}

func Sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

func MeanSquaredError(predicted, actual []float64) float64 {
	sum := 0.0
	for i := range predicted {
		diff := predicted[i] - actual[i]
		sum += diff * diff
	}
	return sum / float64(len(predicted))
}

func BinaryCrossEntropy(predicted, actual []float64) float64 {
	// Placeholder implementation
	return 0.0
}

// SystemMetricsCollector collects system performance metrics
type SystemMetricsCollector struct {
	cpuUsage    float64
	memoryUsage float64
	diskIO      float64
	networkIO   float64
	mutex       sync.RWMutex
}

// MarketMetricsCollector collects market data metrics
type MarketMetricsCollector struct {
	orderFlow  float64
	volatility float64
	spread     float64
	volume     float64
	mutex      sync.RWMutex
}

// NetworkMetricsCollector collects network performance metrics
type NetworkMetricsCollector struct {
	latency    float64
	jitter     float64
	packetLoss float64
	bandwidth  float64
	mutex      sync.RWMutex
}

// TimeFeatureExtractor extracts time-based features
type TimeFeatureExtractor struct {
	timezone *time.Location
}

// NewFeatureExtractor creates a new feature extractor
func NewFeatureExtractor(config *MLConfig) *FeatureExtractor {
	return &FeatureExtractor{
		config:         config,
		systemMetrics:  &SystemMetricsCollector{},
		marketMetrics:  &MarketMetricsCollector{},
		networkMetrics: &NetworkMetricsCollector{},
		timeFeatures:   &TimeFeatureExtractor{timezone: time.UTC},
	}
}

// Initialize initializes the feature extractor
func (fe *FeatureExtractor) Initialize() error {
	// Initialize metrics collectors
	return nil
}

// ExtractFeatures extracts features from system state
func (fe *FeatureExtractor) ExtractFeatures(state *SystemState) ([]float64, error) {
	features := make([]float64, 0, 20)

	// System features
	if fe.config.EnableSystemFeatures {
		features = append(features,
			float64(state.CPUUsage)/100.0,
			float64(state.MemoryUsage)/100.0,
			float64(state.NetworkLoad)/100.0,
			float64(state.DiskIO)/100.0,
			float64(state.ActiveThreads)/1000.0,
		)
	}

	// Time features
	if fe.config.EnableTimeFeatures {
		now := time.Now()
		features = append(features,
			float64(now.Hour())/24.0,
			float64(now.Minute())/60.0,
			float64(now.Second())/60.0,
			float64(now.Weekday())/7.0,
		)
	}

	// Market features
	if fe.config.EnableMarketFeatures {
		fe.marketMetrics.mutex.RLock()
		features = append(features,
			fe.marketMetrics.orderFlow,
			fe.marketMetrics.volatility,
			fe.marketMetrics.spread,
			fe.marketMetrics.volume,
		)
		fe.marketMetrics.mutex.RUnlock()
	}

	// Network features
	if fe.config.EnableNetworkFeatures {
		fe.networkMetrics.mutex.RLock()
		features = append(features,
			fe.networkMetrics.latency,
			fe.networkMetrics.jitter,
			fe.networkMetrics.packetLoss,
			fe.networkMetrics.bandwidth,
		)
		fe.networkMetrics.mutex.RUnlock()
	}

	return features, nil
}

// GetCurrentSystemState returns current system state
func (fe *FeatureExtractor) GetCurrentSystemState() *SystemState {
	return &SystemState{
		CPUUsage:      int64(fe.systemMetrics.cpuUsage * 100),
		MemoryUsage:   int64(fe.systemMetrics.memoryUsage * 100),
		NetworkLoad:   int64(fe.networkMetrics.bandwidth * 100),
		DiskIO:        int64(fe.systemMetrics.diskIO * 100),
		QueueLengths:  make(map[string]int),
		ActiveThreads: 10, // Placeholder
		Timestamp:     time.Now().UnixNano(),
	}
}

// NewFeatureScaler creates a new feature scaler
func NewFeatureScaler(method string) *FeatureScaler {
	return &FeatureScaler{
		method: method,
	}
}

// Transform scales features using the configured method
func (fs *FeatureScaler) Transform(features []float64) []float64 {
	if len(fs.means) == 0 {
		// Initialize with identity transformation if not fitted
		return features
	}

	scaled := make([]float64, len(features))
	for i, feature := range features {
		if i < len(fs.means) {
			if fs.method == "standardize" {
				scaled[i] = (feature - fs.means[i]) / fs.stds[i]
			} else {
				scaled[i] = (feature - fs.mins[i]) / (fs.maxs[i] - fs.mins[i])
			}
		} else {
			scaled[i] = feature
		}
	}

	return scaled
}

// NewCircularBuffer creates a new circular buffer
func NewCircularBuffer(size int) *CircularBuffer {
	return &CircularBuffer{
		data: make([]TrainingExample, size),
		size: size,
	}
}

// Add adds an example to the circular buffer
func (cb *CircularBuffer) Add(example TrainingExample) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.data[cb.head] = example
	cb.head = (cb.head + 1) % cb.size

	if cb.count < cb.size {
		cb.count++
	} else {
		cb.tail = (cb.tail + 1) % cb.size
	}
}

// GetAll returns all examples in the buffer
func (cb *CircularBuffer) GetAll() []TrainingExample {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	if cb.count == 0 {
		return nil
	}

	result := make([]TrainingExample, cb.count)
	for i := 0; i < cb.count; i++ {
		idx := (cb.tail + i) % cb.size
		result[i] = cb.data[idx]
	}

	return result
}

// Count returns the number of examples in the buffer
func (cb *CircularBuffer) Count() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.count
}
