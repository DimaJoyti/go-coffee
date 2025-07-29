package prediction

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// EnhancedMarketPredictor provides advanced AI-powered market prediction capabilities
type EnhancedMarketPredictor struct {
	logger *logger.Logger
	config EnhancedPredictorConfig

	// Enhanced prediction engines
	sentimentEngine AdvancedSentimentEngine
	onChainEngine   AdvancedOnChainEngine
	technicalEngine AdvancedTechnicalEngine
	macroEngine     AdvancedMacroEngine
	mlEngine        MachineLearningEngine
	ensembleEngine  AdvancedEnsembleEngine

	// Data sources
	dataAggregator   DataAggregator
	featureExtractor FeatureExtractor

	// Model management
	modelManager    ModelManager
	predictionCache map[string]*EnhancedPrediction

	// State management
	isRunning    bool
	updateTicker *time.Ticker
	stopChan     chan struct{}
	mutex        sync.RWMutex
	cacheMutex   sync.RWMutex
}

// EnhancedPredictorConfig holds configuration for enhanced prediction
type EnhancedPredictorConfig struct {
	Enabled                bool                    `json:"enabled" yaml:"enabled"`
	UpdateInterval         time.Duration           `json:"update_interval" yaml:"update_interval"`
	PredictionHorizons     []time.Duration         `json:"prediction_horizons" yaml:"prediction_horizons"`
	CacheRetentionPeriod   time.Duration           `json:"cache_retention_period" yaml:"cache_retention_period"`
	SentimentConfig        AdvancedSentimentConfig `json:"sentiment_config" yaml:"sentiment_config"`
	OnChainConfig          AdvancedOnChainConfig   `json:"onchain_config" yaml:"onchain_config"`
	TechnicalConfig        AdvancedTechnicalConfig `json:"technical_config" yaml:"technical_config"`
	MacroConfig            AdvancedMacroConfig     `json:"macro_config" yaml:"macro_config"`
	MLConfig               MachineLearningConfig   `json:"ml_config" yaml:"ml_config"`
	EnsembleConfig         AdvancedEnsembleConfig  `json:"ensemble_config" yaml:"ensemble_config"`
	DataAggregatorConfig   DataAggregatorConfig    `json:"data_aggregator_config" yaml:"data_aggregator_config"`
	FeatureExtractorConfig FeatureExtractorConfig  `json:"feature_extractor_config" yaml:"feature_extractor_config"`
	ModelManagerConfig     ModelManagerConfig      `json:"model_manager_config" yaml:"model_manager_config"`
}

// AdvancedSentimentConfig holds advanced sentiment analysis configuration
type AdvancedSentimentConfig struct {
	Enabled           bool                       `json:"enabled" yaml:"enabled"`
	Sources           []string                   `json:"sources" yaml:"sources"`
	LanguageModels    []string                   `json:"language_models" yaml:"language_models"`
	SentimentWeights  map[string]decimal.Decimal `json:"sentiment_weights" yaml:"sentiment_weights"`
	InfluencerWeights map[string]decimal.Decimal `json:"influencer_weights" yaml:"influencer_weights"`
	UpdateFrequency   time.Duration              `json:"update_frequency" yaml:"update_frequency"`
	HistoryWindow     time.Duration              `json:"history_window" yaml:"history_window"`
	EmotionAnalysis   bool                       `json:"emotion_analysis" yaml:"emotion_analysis"`
	TopicModeling     bool                       `json:"topic_modeling" yaml:"topic_modeling"`
}

// AdvancedOnChainConfig holds advanced on-chain analysis configuration
type AdvancedOnChainConfig struct {
	Enabled         bool                       `json:"enabled" yaml:"enabled"`
	Metrics         []string                   `json:"metrics" yaml:"metrics"`
	Chains          []string                   `json:"chains" yaml:"chains"`
	MetricWeights   map[string]decimal.Decimal `json:"metric_weights" yaml:"metric_weights"`
	UpdateFrequency time.Duration              `json:"update_frequency" yaml:"update_frequency"`
	HistoryWindow   time.Duration              `json:"history_window" yaml:"history_window"`
	WhaleTracking   bool                       `json:"whale_tracking" yaml:"whale_tracking"`
	DeFiIntegration bool                       `json:"defi_integration" yaml:"defi_integration"`
	NFTAnalysis     bool                       `json:"nft_analysis" yaml:"nft_analysis"`
}

// AdvancedTechnicalConfig holds advanced technical analysis configuration
type AdvancedTechnicalConfig struct {
	Enabled            bool                       `json:"enabled" yaml:"enabled"`
	Indicators         []string                   `json:"indicators" yaml:"indicators"`
	Timeframes         []string                   `json:"timeframes" yaml:"timeframes"`
	IndicatorWeights   map[string]decimal.Decimal `json:"indicator_weights" yaml:"indicator_weights"`
	UpdateFrequency    time.Duration              `json:"update_frequency" yaml:"update_frequency"`
	PatternRecognition bool                       `json:"pattern_recognition" yaml:"pattern_recognition"`
	VolumeAnalysis     bool                       `json:"volume_analysis" yaml:"volume_analysis"`
	SupportResistance  bool                       `json:"support_resistance" yaml:"support_resistance"`
}

// AdvancedMacroConfig holds advanced macro analysis configuration
type AdvancedMacroConfig struct {
	Enabled            bool                       `json:"enabled" yaml:"enabled"`
	Factors            []string                   `json:"factors" yaml:"factors"`
	FactorWeights      map[string]decimal.Decimal `json:"factor_weights" yaml:"factor_weights"`
	UpdateFrequency    time.Duration              `json:"update_frequency" yaml:"update_frequency"`
	EconomicIndicators bool                       `json:"economic_indicators" yaml:"economic_indicators"`
	GeopoliticalEvents bool                       `json:"geopolitical_events" yaml:"geopolitical_events"`
	MonetaryPolicy     bool                       `json:"monetary_policy" yaml:"monetary_policy"`
}

// MachineLearningConfig holds machine learning configuration
type MachineLearningConfig struct {
	Enabled              bool                       `json:"enabled" yaml:"enabled"`
	Models               []string                   `json:"models" yaml:"models"`
	ModelWeights         map[string]decimal.Decimal `json:"model_weights" yaml:"model_weights"`
	TrainingFrequency    time.Duration              `json:"training_frequency" yaml:"training_frequency"`
	ValidationSplit      decimal.Decimal            `json:"validation_split" yaml:"validation_split"`
	FeatureSelection     bool                       `json:"feature_selection" yaml:"feature_selection"`
	HyperparameterTuning bool                       `json:"hyperparameter_tuning" yaml:"hyperparameter_tuning"`
	EnsembleMethods      []string                   `json:"ensemble_methods" yaml:"ensemble_methods"`
}

// AdvancedEnsembleConfig holds advanced ensemble configuration
type AdvancedEnsembleConfig struct {
	Enabled             bool                       `json:"enabled" yaml:"enabled"`
	CombinationMethod   string                     `json:"combination_method" yaml:"combination_method"`
	WeightingStrategy   string                     `json:"weighting_strategy" yaml:"weighting_strategy"`
	PerformanceWeights  map[string]decimal.Decimal `json:"performance_weights" yaml:"performance_weights"`
	ConfidenceThreshold decimal.Decimal            `json:"confidence_threshold" yaml:"confidence_threshold"`
	DiversityBonus      decimal.Decimal            `json:"diversity_bonus" yaml:"diversity_bonus"`
	AdaptiveLearning    bool                       `json:"adaptive_learning" yaml:"adaptive_learning"`
}

// DataAggregatorConfig holds data aggregator configuration
type DataAggregatorConfig struct {
	Enabled           bool          `json:"enabled" yaml:"enabled"`
	Sources           []string      `json:"sources" yaml:"sources"`
	UpdateFrequency   time.Duration `json:"update_frequency" yaml:"update_frequency"`
	DataRetention     time.Duration `json:"data_retention" yaml:"data_retention"`
	QualityFiltering  bool          `json:"quality_filtering" yaml:"quality_filtering"`
	RealTimeStreaming bool          `json:"real_time_streaming" yaml:"real_time_streaming"`
}

// FeatureExtractorConfig holds feature extractor configuration
type FeatureExtractorConfig struct {
	Enabled                 bool     `json:"enabled" yaml:"enabled"`
	FeatureTypes            []string `json:"feature_types" yaml:"feature_types"`
	ExtractionMethods       []string `json:"extraction_methods" yaml:"extraction_methods"`
	FeatureSelection        bool     `json:"feature_selection" yaml:"feature_selection"`
	DimensionalityReduction bool     `json:"dimensionality_reduction" yaml:"dimensionality_reduction"`
	FeatureEngineering      bool     `json:"feature_engineering" yaml:"feature_engineering"`
}

// ModelManagerConfig holds model manager configuration
type ModelManagerConfig struct {
	Enabled             bool     `json:"enabled" yaml:"enabled"`
	ModelTypes          []string `json:"model_types" yaml:"model_types"`
	AutoRetraining      bool     `json:"auto_retraining" yaml:"auto_retraining"`
	PerformanceTracking bool     `json:"performance_tracking" yaml:"performance_tracking"`
	ModelVersioning     bool     `json:"model_versioning" yaml:"model_versioning"`
	A_BTesting          bool     `json:"ab_testing" yaml:"ab_testing"`
}

// EnhancedPrediction represents an enhanced market prediction
type EnhancedPrediction struct {
	Asset             string                               `json:"asset"`
	Timestamp         time.Time                            `json:"timestamp"`
	Horizons          map[time.Duration]*HorizonPrediction `json:"horizons"`
	OverallDirection  string                               `json:"overall_direction"`
	OverallConfidence decimal.Decimal                      `json:"overall_confidence"`
	RiskLevel         string                               `json:"risk_level"`
	Volatility        decimal.Decimal                      `json:"volatility"`
	SentimentScore    decimal.Decimal                      `json:"sentiment_score"`
	OnChainScore      decimal.Decimal                      `json:"onchain_score"`
	TechnicalScore    decimal.Decimal                      `json:"technical_score"`
	MacroScore        decimal.Decimal                      `json:"macro_score"`
	MLScore           decimal.Decimal                      `json:"ml_score"`
	EnsembleScore     decimal.Decimal                      `json:"ensemble_score"`
	FeatureImportance map[string]decimal.Decimal           `json:"feature_importance"`
	Scenarios         []*PredictionScenario                `json:"scenarios"`
	Alerts            []*PredictionAlert                   `json:"alerts"`
	Metadata          map[string]interface{}               `json:"metadata"`
}

// HorizonPrediction represents prediction for a specific time horizon
type HorizonPrediction struct {
	Horizon            time.Duration   `json:"horizon"`
	PredictedPrice     decimal.Decimal `json:"predicted_price"`
	PriceRange         PriceRange      `json:"price_range"`
	Direction          string          `json:"direction"`
	Confidence         decimal.Decimal `json:"confidence"`
	ExpectedReturn     decimal.Decimal `json:"expected_return"`
	RiskAdjustedReturn decimal.Decimal `json:"risk_adjusted_return"`
	Probability        decimal.Decimal `json:"probability"`
}

// PredictionScenario represents a prediction scenario
type PredictionScenario struct {
	Name              string          `json:"name"`
	Probability       decimal.Decimal `json:"probability"`
	PriceTarget       decimal.Decimal `json:"price_target"`
	TimeToTarget      time.Duration   `json:"time_to_target"`
	TriggerConditions []string        `json:"trigger_conditions"`
	RiskFactors       []string        `json:"risk_factors"`
}

// PredictionAlert represents a prediction alert
type PredictionAlert struct {
	Type                 string          `json:"type"`
	Severity             string          `json:"severity"`
	Title                string          `json:"title"`
	Message              string          `json:"message"`
	Confidence           decimal.Decimal `json:"confidence"`
	ActionRecommendation string          `json:"action_recommendation"`
	ExpiresAt            time.Time       `json:"expires_at"`
}

// PriceRange represents a price range
type PriceRange struct {
	Low               decimal.Decimal `json:"low"`
	High              decimal.Decimal `json:"high"`
	Mean              decimal.Decimal `json:"mean"`
	StandardDeviation decimal.Decimal `json:"standard_deviation"`
}

// Component interfaces
type AdvancedSentimentEngine interface {
	AnalyzeSentiment(ctx context.Context, asset string) (decimal.Decimal, error)
	GetSentimentTrends(asset string, window time.Duration) ([]SentimentDataPoint, error)
}

type AdvancedOnChainEngine interface {
	AnalyzeOnChain(ctx context.Context, asset string) (decimal.Decimal, error)
	GetOnChainMetrics(asset string) (*OnChainMetrics, error)
}

type AdvancedTechnicalEngine interface {
	AnalyzeTechnical(ctx context.Context, asset string) (decimal.Decimal, error)
	GetTechnicalIndicators(asset string) (*TechnicalIndicators, error)
}

type AdvancedMacroEngine interface {
	AnalyzeMacro(ctx context.Context, asset string) (decimal.Decimal, error)
	GetMacroFactors() (*MacroFactors, error)
}

type MachineLearningEngine interface {
	Predict(ctx context.Context, features []decimal.Decimal) (*MLPrediction, error)
	TrainModel(ctx context.Context, data *TrainingData) error
	GetModelPerformance() (*ModelPerformance, error)
}

type AdvancedEnsembleEngine interface {
	CombinePredictions(predictions map[string]decimal.Decimal) (*EnsemblePrediction, error)
	UpdateWeights(performance map[string]decimal.Decimal) error
}

type DataAggregator interface {
	AggregateData(ctx context.Context, asset string) (*AggregatedData, error)
	GetDataQuality(source string) decimal.Decimal
}

type FeatureExtractor interface {
	ExtractFeatures(ctx context.Context, data *AggregatedData) ([]decimal.Decimal, error)
	GetFeatureImportance() map[string]decimal.Decimal
}

type ModelManager interface {
	GetBestModel(modelType string) (Model, error)
	UpdateModelPerformance(modelID string, performance decimal.Decimal) error
	TriggerRetraining(modelType string) error
}

// Supporting types
type SentimentDataPoint struct {
	Timestamp time.Time       `json:"timestamp"`
	Score     decimal.Decimal `json:"score"`
	Volume    int             `json:"volume"`
	Source    string          `json:"source"`
}

type OnChainMetrics struct {
	TransactionVolume  decimal.Decimal `json:"transaction_volume"`
	ActiveAddresses    int             `json:"active_addresses"`
	NetworkUtilization decimal.Decimal `json:"network_utilization"`
	WhaleActivity      decimal.Decimal `json:"whale_activity"`
	DeFiTVL            decimal.Decimal `json:"defi_tvl"`
}

type TechnicalIndicators struct {
	RSI              decimal.Decimal   `json:"rsi"`
	MACD             decimal.Decimal   `json:"macd"`
	BollingerBands   PriceRange        `json:"bollinger_bands"`
	SupportLevels    []decimal.Decimal `json:"support_levels"`
	ResistanceLevels []decimal.Decimal `json:"resistance_levels"`
	VolumeProfile    decimal.Decimal   `json:"volume_profile"`
}

type MacroFactors struct {
	InterestRates        decimal.Decimal `json:"interest_rates"`
	InflationRate        decimal.Decimal `json:"inflation_rate"`
	USDIndex             decimal.Decimal `json:"usd_index"`
	StockMarketSentiment decimal.Decimal `json:"stock_market_sentiment"`
	GeopoliticalRisk     decimal.Decimal `json:"geopolitical_risk"`
}

type MLPrediction struct {
	Prediction        decimal.Decimal            `json:"prediction"`
	Confidence        decimal.Decimal            `json:"confidence"`
	FeatureImportance map[string]decimal.Decimal `json:"feature_importance"`
}

type EnsemblePrediction struct {
	Prediction       decimal.Decimal            `json:"prediction"`
	Confidence       decimal.Decimal            `json:"confidence"`
	ComponentWeights map[string]decimal.Decimal `json:"component_weights"`
}

type AggregatedData struct {
	PriceData     []PricePoint         `json:"price_data"`
	VolumeData    []VolumePoint        `json:"volume_data"`
	SentimentData []SentimentDataPoint `json:"sentiment_data"`
	OnChainData   *OnChainMetrics      `json:"onchain_data"`
	MacroData     *MacroFactors        `json:"macro_data"`
}

type PricePoint struct {
	Timestamp time.Time       `json:"timestamp"`
	Open      decimal.Decimal `json:"open"`
	High      decimal.Decimal `json:"high"`
	Low       decimal.Decimal `json:"low"`
	Close     decimal.Decimal `json:"close"`
}

type VolumePoint struct {
	Timestamp  time.Time       `json:"timestamp"`
	Volume     decimal.Decimal `json:"volume"`
	BuyVolume  decimal.Decimal `json:"buy_volume"`
	SellVolume decimal.Decimal `json:"sell_volume"`
}

type TrainingData struct {
	Features   [][]decimal.Decimal `json:"features"`
	Labels     []decimal.Decimal   `json:"labels"`
	Timestamps []time.Time         `json:"timestamps"`
}

type ModelPerformance struct {
	Accuracy  decimal.Decimal `json:"accuracy"`
	Precision decimal.Decimal `json:"precision"`
	Recall    decimal.Decimal `json:"recall"`
	F1Score   decimal.Decimal `json:"f1_score"`
	MAE       decimal.Decimal `json:"mae"`
	RMSE      decimal.Decimal `json:"rmse"`
}

type Model interface {
	Predict(features []decimal.Decimal) (decimal.Decimal, error)
	Train(data *TrainingData) error
	GetPerformance() *ModelPerformance
}

// NewEnhancedMarketPredictor creates a new enhanced market predictor
func NewEnhancedMarketPredictor(logger *logger.Logger, config EnhancedPredictorConfig) *EnhancedMarketPredictor {
	emp := &EnhancedMarketPredictor{
		logger:          logger.Named("enhanced-market-predictor"),
		config:          config,
		predictionCache: make(map[string]*EnhancedPrediction),
		stopChan:        make(chan struct{}),
	}

	// Initialize components (mock implementations for this example)
	emp.initializeComponents()

	return emp
}

// initializeComponents initializes all prediction components
func (emp *EnhancedMarketPredictor) initializeComponents() {
	// Initialize components with mock implementations
	// In production, these would be real implementations
	emp.sentimentEngine = &MockAdvancedSentimentEngine{}
	emp.onChainEngine = &MockAdvancedOnChainEngine{}
	emp.technicalEngine = &MockAdvancedTechnicalEngine{}
	emp.macroEngine = &MockAdvancedMacroEngine{}
	emp.mlEngine = &MockMachineLearningEngine{}
	emp.ensembleEngine = &MockAdvancedEnsembleEngine{}
	emp.dataAggregator = &MockDataAggregator{}
	emp.featureExtractor = &MockFeatureExtractor{}
	emp.modelManager = &MockModelManager{}
}

// Start starts the enhanced market predictor
func (emp *EnhancedMarketPredictor) Start(ctx context.Context) error {
	emp.mutex.Lock()
	defer emp.mutex.Unlock()

	if emp.isRunning {
		return fmt.Errorf("enhanced market predictor is already running")
	}

	if !emp.config.Enabled {
		emp.logger.Info("Enhanced market predictor is disabled")
		return nil
	}

	emp.logger.Info("Starting enhanced market predictor",
		zap.Duration("update_interval", emp.config.UpdateInterval),
		zap.Int("prediction_horizons", len(emp.config.PredictionHorizons)))

	// Start monitoring loop
	emp.updateTicker = time.NewTicker(emp.config.UpdateInterval)
	go emp.monitoringLoop(ctx)

	// Start cache cleanup routine
	go emp.cacheCleanupLoop(ctx)

	emp.isRunning = true
	emp.logger.Info("Enhanced market predictor started successfully")
	return nil
}

// Stop stops the enhanced market predictor
func (emp *EnhancedMarketPredictor) Stop() error {
	emp.mutex.Lock()
	defer emp.mutex.Unlock()

	if !emp.isRunning {
		return nil
	}

	emp.logger.Info("Stopping enhanced market predictor")

	// Stop monitoring
	if emp.updateTicker != nil {
		emp.updateTicker.Stop()
	}
	close(emp.stopChan)

	emp.isRunning = false
	emp.logger.Info("Enhanced market predictor stopped")
	return nil
}

// PredictMarketEnhanced generates enhanced market predictions
func (emp *EnhancedMarketPredictor) PredictMarketEnhanced(ctx context.Context, asset string, currentPrice decimal.Decimal) (*EnhancedPrediction, error) {
	startTime := time.Now()
	emp.logger.Debug("Generating enhanced market prediction",
		zap.String("asset", asset),
		zap.String("current_price", currentPrice.String()))

	// Check cache first
	cacheKey := fmt.Sprintf("%s_%s", asset, currentPrice.String())
	if cached := emp.getCachedPrediction(cacheKey); cached != nil {
		emp.logger.Debug("Returning cached enhanced prediction")
		return cached, nil
	}

	// Aggregate data from all sources
	aggregatedData, err := emp.dataAggregator.AggregateData(ctx, asset)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate data: %w", err)
	}

	// Extract features
	features, err := emp.featureExtractor.ExtractFeatures(ctx, aggregatedData)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features: %w", err)
	}

	// Run individual prediction engines
	predictions := make(map[string]decimal.Decimal)

	// Sentiment analysis
	if emp.config.SentimentConfig.Enabled {
		sentimentScore, err := emp.sentimentEngine.AnalyzeSentiment(ctx, asset)
		if err != nil {
			emp.logger.Warn("Sentiment analysis failed", zap.Error(err))
		} else {
			predictions["sentiment"] = sentimentScore
		}
	}

	// On-chain analysis
	if emp.config.OnChainConfig.Enabled {
		onChainScore, err := emp.onChainEngine.AnalyzeOnChain(ctx, asset)
		if err != nil {
			emp.logger.Warn("On-chain analysis failed", zap.Error(err))
		} else {
			predictions["onchain"] = onChainScore
		}
	}

	// Technical analysis
	if emp.config.TechnicalConfig.Enabled {
		technicalScore, err := emp.technicalEngine.AnalyzeTechnical(ctx, asset)
		if err != nil {
			emp.logger.Warn("Technical analysis failed", zap.Error(err))
		} else {
			predictions["technical"] = technicalScore
		}
	}

	// Macro analysis
	if emp.config.MacroConfig.Enabled {
		macroScore, err := emp.macroEngine.AnalyzeMacro(ctx, asset)
		if err != nil {
			emp.logger.Warn("Macro analysis failed", zap.Error(err))
		} else {
			predictions["macro"] = macroScore
		}
	}

	// Machine learning prediction
	if emp.config.MLConfig.Enabled {
		mlPrediction, err := emp.mlEngine.Predict(ctx, features)
		if err != nil {
			emp.logger.Warn("ML prediction failed", zap.Error(err))
		} else {
			predictions["ml"] = mlPrediction.Prediction
		}
	}

	// Ensemble prediction
	ensemblePrediction, err := emp.ensembleEngine.CombinePredictions(predictions)
	if err != nil {
		return nil, fmt.Errorf("failed to combine predictions: %w", err)
	}

	// Generate horizon-specific predictions
	horizonPredictions := make(map[time.Duration]*HorizonPrediction)
	for _, horizon := range emp.config.PredictionHorizons {
		horizonPred := emp.generateHorizonPrediction(horizon, currentPrice, ensemblePrediction)
		horizonPredictions[horizon] = horizonPred
	}

	// Create enhanced prediction
	prediction := &EnhancedPrediction{
		Asset:             asset,
		Timestamp:         time.Now(),
		Horizons:          horizonPredictions,
		OverallDirection:  emp.determineOverallDirection(ensemblePrediction.Prediction, currentPrice),
		OverallConfidence: ensemblePrediction.Confidence,
		RiskLevel:         emp.calculateRiskLevel(ensemblePrediction),
		Volatility:        emp.calculateVolatility(aggregatedData),
		SentimentScore:    predictions["sentiment"],
		OnChainScore:      predictions["onchain"],
		TechnicalScore:    predictions["technical"],
		MacroScore:        predictions["macro"],
		MLScore:           predictions["ml"],
		EnsembleScore:     ensemblePrediction.Prediction,
		FeatureImportance: emp.featureExtractor.GetFeatureImportance(),
		Scenarios:         emp.generateScenarios(currentPrice, ensemblePrediction),
		Alerts:            emp.generateAlerts(asset, currentPrice, ensemblePrediction),
		Metadata:          make(map[string]interface{}),
	}

	// Cache the prediction
	emp.cachePrediction(cacheKey, prediction)

	emp.logger.Info("Enhanced market prediction completed",
		zap.String("asset", asset),
		zap.String("overall_direction", prediction.OverallDirection),
		zap.String("overall_confidence", prediction.OverallConfidence.String()),
		zap.String("risk_level", prediction.RiskLevel),
		zap.Duration("processing_time", time.Since(startTime)))

	return prediction, nil
}

// generateHorizonPrediction generates prediction for a specific time horizon
func (emp *EnhancedMarketPredictor) generateHorizonPrediction(horizon time.Duration, currentPrice decimal.Decimal, ensemble *EnsemblePrediction) *HorizonPrediction {
	// Apply time decay to prediction confidence
	timeDecayFactor := emp.calculateTimeDecayFactor(horizon)
	adjustedConfidence := ensemble.Confidence.Mul(timeDecayFactor)

	// Calculate predicted price based on ensemble prediction and horizon
	priceChange := ensemble.Prediction.Sub(decimal.NewFromFloat(0.5)).Mul(decimal.NewFromFloat(0.2)) // Scale to ±10%
	horizonMultiplier := decimal.NewFromFloat(float64(horizon) / float64(time.Hour))
	adjustedPriceChange := priceChange.Mul(horizonMultiplier)
	predictedPrice := currentPrice.Mul(decimal.NewFromFloat(1).Add(adjustedPriceChange))

	// Calculate price range
	volatility := emp.calculateHorizonVolatility(horizon)
	priceRange := PriceRange{
		Low:               predictedPrice.Mul(decimal.NewFromFloat(1).Sub(volatility)),
		High:              predictedPrice.Mul(decimal.NewFromFloat(1).Add(volatility)),
		Mean:              predictedPrice,
		StandardDeviation: predictedPrice.Mul(volatility),
	}

	// Determine direction
	direction := "neutral"
	if predictedPrice.GreaterThan(currentPrice.Mul(decimal.NewFromFloat(1.01))) {
		direction = "bullish"
	} else if predictedPrice.LessThan(currentPrice.Mul(decimal.NewFromFloat(0.99))) {
		direction = "bearish"
	}

	// Calculate expected return
	expectedReturn := predictedPrice.Sub(currentPrice).Div(currentPrice)
	riskAdjustedReturn := expectedReturn.Div(volatility.Add(decimal.NewFromFloat(0.01))) // Add small epsilon

	return &HorizonPrediction{
		Horizon:            horizon,
		PredictedPrice:     predictedPrice,
		PriceRange:         priceRange,
		Direction:          direction,
		Confidence:         adjustedConfidence,
		ExpectedReturn:     expectedReturn,
		RiskAdjustedReturn: riskAdjustedReturn,
		Probability:        adjustedConfidence,
	}
}

// Helper methods

func (emp *EnhancedMarketPredictor) calculateTimeDecayFactor(horizon time.Duration) decimal.Decimal {
	// Confidence decreases with longer time horizons
	hours := float64(horizon) / float64(time.Hour)
	decayFactor := 1.0 / (1.0 + hours*0.1) // 10% decay per hour
	return decimal.NewFromFloat(decayFactor)
}

func (emp *EnhancedMarketPredictor) calculateHorizonVolatility(horizon time.Duration) decimal.Decimal {
	// Volatility increases with longer time horizons
	hours := float64(horizon) / float64(time.Hour)
	baseVolatility := 0.05                                   // 5% base volatility
	horizonVolatility := baseVolatility * (1.0 + hours*0.02) // 2% increase per hour
	return decimal.NewFromFloat(horizonVolatility)
}

func (emp *EnhancedMarketPredictor) determineOverallDirection(prediction, currentPrice decimal.Decimal) string {
	threshold := decimal.NewFromFloat(0.02) // 2% threshold

	if prediction.GreaterThan(decimal.NewFromFloat(0.5).Add(threshold)) {
		return "bullish"
	} else if prediction.LessThan(decimal.NewFromFloat(0.5).Sub(threshold)) {
		return "bearish"
	}
	return "neutral"
}

func (emp *EnhancedMarketPredictor) calculateRiskLevel(ensemble *EnsemblePrediction) string {
	confidence := ensemble.Confidence

	if confidence.GreaterThan(decimal.NewFromFloat(0.8)) {
		return "low"
	} else if confidence.GreaterThan(decimal.NewFromFloat(0.6)) {
		return "medium"
	}
	return "high"
}

func (emp *EnhancedMarketPredictor) calculateVolatility(data *AggregatedData) decimal.Decimal {
	// Calculate volatility from price data
	if len(data.PriceData) < 2 {
		return decimal.NewFromFloat(0.05) // Default 5%
	}

	var returns []decimal.Decimal
	for i := 1; i < len(data.PriceData); i++ {
		prevPrice := data.PriceData[i-1].Close
		currPrice := data.PriceData[i].Close
		if !prevPrice.IsZero() {
			ret := currPrice.Sub(prevPrice).Div(prevPrice)
			returns = append(returns, ret)
		}
	}

	if len(returns) == 0 {
		return decimal.NewFromFloat(0.05)
	}

	// Calculate standard deviation of returns
	var sum decimal.Decimal
	for _, ret := range returns {
		sum = sum.Add(ret)
	}
	mean := sum.Div(decimal.NewFromInt(int64(len(returns))))

	var variance decimal.Decimal
	for _, ret := range returns {
		diff := ret.Sub(mean)
		variance = variance.Add(diff.Mul(diff))
	}
	variance = variance.Div(decimal.NewFromInt(int64(len(returns))))

	// Return standard deviation as volatility
	return decimal.NewFromFloat(variance.InexactFloat64()).Pow(decimal.NewFromFloat(0.5))
}

func (emp *EnhancedMarketPredictor) generateScenarios(currentPrice decimal.Decimal, ensemble *EnsemblePrediction) []*PredictionScenario {
	scenarios := []*PredictionScenario{
		{
			Name:              "Bull Case",
			Probability:       decimal.NewFromFloat(0.3),
			PriceTarget:       currentPrice.Mul(decimal.NewFromFloat(1.2)),
			TimeToTarget:      7 * 24 * time.Hour,
			TriggerConditions: []string{"Strong positive sentiment", "Increased institutional adoption"},
			RiskFactors:       []string{"Regulatory uncertainty", "Market correction"},
		},
		{
			Name:              "Base Case",
			Probability:       decimal.NewFromFloat(0.4),
			PriceTarget:       currentPrice.Mul(decimal.NewFromFloat(1.05)),
			TimeToTarget:      30 * 24 * time.Hour,
			TriggerConditions: []string{"Stable market conditions", "Gradual adoption"},
			RiskFactors:       []string{"Economic uncertainty", "Competition"},
		},
		{
			Name:              "Bear Case",
			Probability:       decimal.NewFromFloat(0.3),
			PriceTarget:       currentPrice.Mul(decimal.NewFromFloat(0.8)),
			TimeToTarget:      14 * 24 * time.Hour,
			TriggerConditions: []string{"Negative regulatory news", "Market downturn"},
			RiskFactors:       []string{"Liquidity crisis", "Technical issues"},
		},
	}

	return scenarios
}

func (emp *EnhancedMarketPredictor) generateAlerts(asset string, currentPrice decimal.Decimal, ensemble *EnsemblePrediction) []*PredictionAlert {
	alerts := make([]*PredictionAlert, 0)

	// High confidence prediction alert
	if ensemble.Confidence.GreaterThan(decimal.NewFromFloat(0.9)) {
		direction := "neutral"
		if ensemble.Prediction.GreaterThan(decimal.NewFromFloat(0.6)) {
			direction = "bullish"
		} else if ensemble.Prediction.LessThan(decimal.NewFromFloat(0.4)) {
			direction = "bearish"
		}

		if direction != "neutral" {
			alerts = append(alerts, &PredictionAlert{
				Type:                 "high_confidence_prediction",
				Severity:             "info",
				Title:                fmt.Sprintf("High Confidence %s Signal", strings.Title(direction)),
				Message:              fmt.Sprintf("Strong %s prediction for %s with %.1f%% confidence", direction, asset, ensemble.Confidence.Mul(decimal.NewFromFloat(100)).InexactFloat64()),
				Confidence:           ensemble.Confidence,
				ActionRecommendation: fmt.Sprintf("Consider %s position", direction),
				ExpiresAt:            time.Now().Add(24 * time.Hour),
			})
		}
	}

	return alerts
}

// Cache management methods

func (emp *EnhancedMarketPredictor) getCachedPrediction(key string) *EnhancedPrediction {
	emp.cacheMutex.RLock()
	defer emp.cacheMutex.RUnlock()

	if prediction, exists := emp.predictionCache[key]; exists {
		// Check if prediction is still fresh
		if time.Since(prediction.Timestamp) < emp.config.CacheRetentionPeriod {
			return prediction
		}
		// Remove stale prediction
		delete(emp.predictionCache, key)
	}
	return nil
}

func (emp *EnhancedMarketPredictor) cachePrediction(key string, prediction *EnhancedPrediction) {
	emp.cacheMutex.Lock()
	defer emp.cacheMutex.Unlock()

	emp.predictionCache[key] = prediction

	// Limit cache size
	if len(emp.predictionCache) > 1000 {
		// Remove oldest entries
		oldestTime := time.Now()
		var oldestKey string
		for k, v := range emp.predictionCache {
			if v.Timestamp.Before(oldestTime) {
				oldestTime = v.Timestamp
				oldestKey = k
			}
		}
		if oldestKey != "" {
			delete(emp.predictionCache, oldestKey)
		}
	}
}

// Monitoring and cleanup loops

func (emp *EnhancedMarketPredictor) monitoringLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-emp.stopChan:
			return
		case <-emp.updateTicker.C:
			emp.performPeriodicUpdate()
		}
	}
}

func (emp *EnhancedMarketPredictor) cacheCleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(emp.config.CacheRetentionPeriod / 10)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-emp.stopChan:
			return
		case <-ticker.C:
			emp.cleanupCache()
		}
	}
}

func (emp *EnhancedMarketPredictor) performPeriodicUpdate() {
	emp.logger.Debug("Performing periodic enhanced prediction update")

	// Update model performance tracking
	// Retrain models if needed
	// Update ensemble weights
	// This would be implemented based on specific requirements
}

func (emp *EnhancedMarketPredictor) cleanupCache() {
	emp.cacheMutex.Lock()
	defer emp.cacheMutex.Unlock()

	cutoff := time.Now().Add(-emp.config.CacheRetentionPeriod)
	for key, prediction := range emp.predictionCache {
		if prediction.Timestamp.Before(cutoff) {
			delete(emp.predictionCache, key)
		}
	}
}

// IsRunning returns whether the predictor is running
func (emp *EnhancedMarketPredictor) IsRunning() bool {
	emp.mutex.RLock()
	defer emp.mutex.RUnlock()
	return emp.isRunning
}

// GetMetrics returns predictor metrics
func (emp *EnhancedMarketPredictor) GetMetrics() map[string]interface{} {
	emp.cacheMutex.RLock()
	defer emp.cacheMutex.RUnlock()

	return map[string]interface{}{
		"is_running":          emp.IsRunning(),
		"cache_size":          len(emp.predictionCache),
		"sentiment_enabled":   emp.config.SentimentConfig.Enabled,
		"onchain_enabled":     emp.config.OnChainConfig.Enabled,
		"technical_enabled":   emp.config.TechnicalConfig.Enabled,
		"macro_enabled":       emp.config.MacroConfig.Enabled,
		"ml_enabled":          emp.config.MLConfig.Enabled,
		"ensemble_enabled":    emp.config.EnsembleConfig.Enabled,
		"prediction_horizons": len(emp.config.PredictionHorizons),
	}
}
