package prediction

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// MarketPredictor provides AI-powered market prediction capabilities
type MarketPredictor struct {
	logger *logger.Logger
	config MarketPredictorConfig

	// Prediction engines
	sentimentAnalyzer *SentimentAnalyzer
	onChainAnalyzer   *OnChainAnalyzer
	technicalAnalyzer *TechnicalAnalyzer
	macroAnalyzer     *MacroAnalyzer
	ensembleModel     *EnsembleModel

	// Data management
	predictionCache map[string]*MarketPrediction
	modelCache      map[string]*PredictionModel
	cacheMutex      sync.RWMutex
	modelMutex      sync.RWMutex

	// State management
	isRunning    bool
	updateTicker *time.Ticker
	stopChan     chan struct{}
	mutex        sync.RWMutex
}

// MarketPredictorConfig holds configuration for market prediction
type MarketPredictorConfig struct {
	Enabled               bool                      `json:"enabled" yaml:"enabled"`
	UpdateInterval        time.Duration             `json:"update_interval" yaml:"update_interval"`
	CacheTimeout          time.Duration             `json:"cache_timeout" yaml:"cache_timeout"`
	PredictionHorizons    []time.Duration           `json:"prediction_horizons" yaml:"prediction_horizons"`
	ConfidenceThreshold   decimal.Decimal           `json:"confidence_threshold" yaml:"confidence_threshold"`
	SentimentConfig       SentimentAnalysisConfig   `json:"sentiment_config" yaml:"sentiment_config"`
	OnChainConfig         OnChainAnalysisConfig     `json:"onchain_config" yaml:"onchain_config"`
	TechnicalConfig       TechnicalAnalysisConfig   `json:"technical_config" yaml:"technical_config"`
	MacroConfig           MacroAnalysisConfig       `json:"macro_config" yaml:"macro_config"`
	EnsembleConfig        EnsembleModelConfig       `json:"ensemble_config" yaml:"ensemble_config"`
	DataSources           []string                  `json:"data_sources" yaml:"data_sources"`
	ModelRetrainingConfig ModelRetrainingConfig     `json:"model_retraining_config" yaml:"model_retraining_config"`
	AlertThresholds       PredictionAlertThresholds `json:"alert_thresholds" yaml:"alert_thresholds"`
}

// SentimentAnalysisConfig holds sentiment analysis configuration
type SentimentAnalysisConfig struct {
	Enabled          bool            `json:"enabled" yaml:"enabled"`
	Sources          []string        `json:"sources" yaml:"sources"`
	UpdateInterval   time.Duration   `json:"update_interval" yaml:"update_interval"`
	SentimentWeight  decimal.Decimal `json:"sentiment_weight" yaml:"sentiment_weight"`
	NewsWeight       decimal.Decimal `json:"news_weight" yaml:"news_weight"`
	SocialWeight     decimal.Decimal `json:"social_weight" yaml:"social_weight"`
	InfluencerWeight decimal.Decimal `json:"influencer_weight" yaml:"influencer_weight"`
	LanguageModels   []string        `json:"language_models" yaml:"language_models"`
}

// OnChainAnalysisConfig holds on-chain analysis configuration
type OnChainAnalysisConfig struct {
	Enabled             bool            `json:"enabled" yaml:"enabled"`
	Metrics             []string        `json:"metrics" yaml:"metrics"`
	UpdateInterval      time.Duration   `json:"update_interval" yaml:"update_interval"`
	TransactionWeight   decimal.Decimal `json:"transaction_weight" yaml:"transaction_weight"`
	AddressWeight       decimal.Decimal `json:"address_weight" yaml:"address_weight"`
	VolumeWeight        decimal.Decimal `json:"volume_weight" yaml:"volume_weight"`
	LiquidityWeight     decimal.Decimal `json:"liquidity_weight" yaml:"liquidity_weight"`
	DeFiWeight          decimal.Decimal `json:"defi_weight" yaml:"defi_weight"`
	NetworkHealthWeight decimal.Decimal `json:"network_health_weight" yaml:"network_health_weight"`
}

// TechnicalAnalysisConfig holds technical analysis configuration
type TechnicalAnalysisConfig struct {
	Enabled                 bool            `json:"enabled" yaml:"enabled"`
	Indicators              []string        `json:"indicators" yaml:"indicators"`
	Timeframes              []string        `json:"timeframes" yaml:"timeframes"`
	UpdateInterval          time.Duration   `json:"update_interval" yaml:"update_interval"`
	TrendWeight             decimal.Decimal `json:"trend_weight" yaml:"trend_weight"`
	MomentumWeight          decimal.Decimal `json:"momentum_weight" yaml:"momentum_weight"`
	VolatilityWeight        decimal.Decimal `json:"volatility_weight" yaml:"volatility_weight"`
	VolumeWeight            decimal.Decimal `json:"volume_weight" yaml:"volume_weight"`
	SupportResistanceWeight decimal.Decimal `json:"support_resistance_weight" yaml:"support_resistance_weight"`
}

// MacroAnalysisConfig holds macro analysis configuration
type MacroAnalysisConfig struct {
	Enabled            bool            `json:"enabled" yaml:"enabled"`
	Indicators         []string        `json:"indicators" yaml:"indicators"`
	UpdateInterval     time.Duration   `json:"update_interval" yaml:"update_interval"`
	EconomicWeight     decimal.Decimal `json:"economic_weight" yaml:"economic_weight"`
	MonetaryWeight     decimal.Decimal `json:"monetary_weight" yaml:"monetary_weight"`
	GeopoliticalWeight decimal.Decimal `json:"geopolitical_weight" yaml:"geopolitical_weight"`
	RegulatoryWeight   decimal.Decimal `json:"regulatory_weight" yaml:"regulatory_weight"`
}

// EnsembleModelConfig holds ensemble model configuration
type EnsembleModelConfig struct {
	Enabled           bool                       `json:"enabled" yaml:"enabled"`
	Models            []string                   `json:"models" yaml:"models"`
	WeightingStrategy string                     `json:"weighting_strategy" yaml:"weighting_strategy"`
	ModelWeights      map[string]decimal.Decimal `json:"model_weights" yaml:"model_weights"`
	RebalanceInterval time.Duration              `json:"rebalance_interval" yaml:"rebalance_interval"`
	PerformanceWindow time.Duration              `json:"performance_window" yaml:"performance_window"`
}

// ModelRetrainingConfig holds model retraining configuration
type ModelRetrainingConfig struct {
	Enabled              bool            `json:"enabled" yaml:"enabled"`
	RetrainingInterval   time.Duration   `json:"retraining_interval" yaml:"retraining_interval"`
	PerformanceThreshold decimal.Decimal `json:"performance_threshold" yaml:"performance_threshold"`
	DataWindow           time.Duration   `json:"data_window" yaml:"data_window"`
	ValidationSplit      decimal.Decimal `json:"validation_split" yaml:"validation_split"`
	AutoDeploy           bool            `json:"auto_deploy" yaml:"auto_deploy"`
}

// PredictionAlertThresholds defines alert thresholds for predictions
type PredictionAlertThresholds struct {
	HighConfidenceBull decimal.Decimal `json:"high_confidence_bull" yaml:"high_confidence_bull"`
	HighConfidenceBear decimal.Decimal `json:"high_confidence_bear" yaml:"high_confidence_bear"`
	LowConfidence      decimal.Decimal `json:"low_confidence" yaml:"low_confidence"`
	VolatilitySpike    decimal.Decimal `json:"volatility_spike" yaml:"volatility_spike"`
	TrendReversal      decimal.Decimal `json:"trend_reversal" yaml:"trend_reversal"`
	AnomalyDetection   decimal.Decimal `json:"anomaly_detection" yaml:"anomaly_detection"`
}

// MarketPrediction represents a market prediction result
type MarketPrediction struct {
	ID                 string                 `json:"id"`
	Asset              string                 `json:"asset"`
	Timestamp          time.Time              `json:"timestamp"`
	PredictionHorizon  time.Duration          `json:"prediction_horizon"`
	CurrentPrice       decimal.Decimal        `json:"current_price"`
	PredictedPrice     decimal.Decimal        `json:"predicted_price"`
	PriceChange        decimal.Decimal        `json:"price_change"`
	PriceChangePercent decimal.Decimal        `json:"price_change_percent"`
	Direction          string                 `json:"direction"` // "bullish", "bearish", "neutral"
	Confidence         decimal.Decimal        `json:"confidence"`
	Volatility         decimal.Decimal        `json:"volatility"`
	SentimentScore     decimal.Decimal        `json:"sentiment_score"`
	OnChainScore       decimal.Decimal        `json:"onchain_score"`
	TechnicalScore     decimal.Decimal        `json:"technical_score"`
	MacroScore         decimal.Decimal        `json:"macro_score"`
	EnsembleScore      decimal.Decimal        `json:"ensemble_score"`
	RiskLevel          string                 `json:"risk_level"`
	Signals            []*PredictionSignal    `json:"signals"`
	Factors            []*PredictionFactor    `json:"factors"`
	Scenarios          []*PredictionScenario  `json:"scenarios"`
	Alerts             []*PredictionAlert     `json:"alerts"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// PredictionSignal represents a prediction signal
type PredictionSignal struct {
	Type        string          `json:"type"`
	Source      string          `json:"source"`
	Strength    decimal.Decimal `json:"strength"`
	Direction   string          `json:"direction"`
	Confidence  decimal.Decimal `json:"confidence"`
	Description string          `json:"description"`
	Timestamp   time.Time       `json:"timestamp"`
}

// PredictionFactor represents a factor influencing the prediction
type PredictionFactor struct {
	Name        string          `json:"name"`
	Category    string          `json:"category"`
	Impact      decimal.Decimal `json:"impact"`
	Weight      decimal.Decimal `json:"weight"`
	Value       decimal.Decimal `json:"value"`
	Description string          `json:"description"`
}

// Note: PredictionScenario and PredictionAlert types are defined in enhanced_predictor.go

// PredictionModel represents a prediction model
type PredictionModel struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Type            string                 `json:"type"`
	Version         string                 `json:"version"`
	Accuracy        decimal.Decimal        `json:"accuracy"`
	LastTrained     time.Time              `json:"last_trained"`
	TrainingData    int                    `json:"training_data"`
	Features        []string               `json:"features"`
	Hyperparameters map[string]interface{} `json:"hyperparameters"`
	Performance     *ModelPerformance      `json:"performance"`
	Status          string                 `json:"status"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// Note: ModelPerformance type is defined in enhanced_predictor.go

// NewMarketPredictor creates a new market predictor
func NewMarketPredictor(logger *logger.Logger, config MarketPredictorConfig) *MarketPredictor {
	predictor := &MarketPredictor{
		logger:          logger.Named("market-predictor"),
		config:          config,
		predictionCache: make(map[string]*MarketPrediction),
		modelCache:      make(map[string]*PredictionModel),
		stopChan:        make(chan struct{}),
	}

	// Initialize prediction engines
	predictor.sentimentAnalyzer = NewSentimentAnalyzer(logger, config.SentimentConfig)
	predictor.onChainAnalyzer = NewOnChainAnalyzer(logger, config.OnChainConfig)
	predictor.technicalAnalyzer = NewTechnicalAnalyzer(logger, config.TechnicalConfig)
	predictor.macroAnalyzer = NewMacroAnalyzer(logger, config.MacroConfig)
	predictor.ensembleModel = NewEnsembleModel(logger, config.EnsembleConfig)

	return predictor
}

// Start starts the market predictor
func (mp *MarketPredictor) Start(ctx context.Context) error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	if mp.isRunning {
		return fmt.Errorf("market predictor is already running")
	}

	if !mp.config.Enabled {
		mp.logger.Info("Market predictor is disabled")
		return nil
	}

	mp.logger.Info("Starting market predictor",
		zap.Duration("update_interval", mp.config.UpdateInterval),
		zap.Int("prediction_horizons", len(mp.config.PredictionHorizons)))

	// Start prediction engines
	if err := mp.sentimentAnalyzer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start sentiment analyzer: %w", err)
	}

	if err := mp.onChainAnalyzer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start on-chain analyzer: %w", err)
	}

	if err := mp.technicalAnalyzer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start technical analyzer: %w", err)
	}

	if err := mp.macroAnalyzer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start macro analyzer: %w", err)
	}

	if err := mp.ensembleModel.Start(ctx); err != nil {
		return fmt.Errorf("failed to start ensemble model: %w", err)
	}

	// Start monitoring loop
	mp.updateTicker = time.NewTicker(mp.config.UpdateInterval)
	go mp.monitoringLoop(ctx)

	mp.isRunning = true
	mp.logger.Info("Market predictor started successfully")
	return nil
}

// PredictMarket generates market predictions for a given asset
func (mp *MarketPredictor) PredictMarket(ctx context.Context, asset string, currentPrice decimal.Decimal) (*MarketPrediction, error) {
	startTime := time.Now()
	mp.logger.Info("Starting market prediction",
		zap.String("asset", asset),
		zap.String("current_price", currentPrice.String()))

	// Check cache first
	cacheKey := mp.generateCacheKey(asset)
	if cached := mp.getCachedPrediction(cacheKey); cached != nil {
		mp.logger.Debug("Returning cached market prediction")
		return cached, nil
	}

	// Initialize prediction result
	prediction := &MarketPrediction{
		ID:                mp.generatePredictionID(),
		Asset:             asset,
		Timestamp:         time.Now(),
		PredictionHorizon: mp.config.PredictionHorizons[0], // Use first horizon as default
		CurrentPrice:      currentPrice,
		Signals:           []*PredictionSignal{},
		Factors:           []*PredictionFactor{},
		Scenarios:         []*PredictionScenario{},
		Alerts:            []*PredictionAlert{},
		Metadata:          make(map[string]interface{}),
	}

	// Run prediction engines in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make([]error, 0)

	// Sentiment analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		sentimentScore, err := mp.sentimentAnalyzer.AnalyzeSentiment(ctx, asset)
		mu.Lock()
		if err != nil {
			errors = append(errors, fmt.Errorf("sentiment analysis: %w", err))
		} else {
			prediction.SentimentScore = sentimentScore
		}
		mu.Unlock()
	}()

	// On-chain analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		onChainScore, err := mp.onChainAnalyzer.AnalyzeOnChain(ctx, asset)
		mu.Lock()
		if err != nil {
			errors = append(errors, fmt.Errorf("on-chain analysis: %w", err))
		} else {
			prediction.OnChainScore = onChainScore
		}
		mu.Unlock()
	}()

	// Technical analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		technicalScore, err := mp.technicalAnalyzer.AnalyzeTechnical(ctx, asset, currentPrice)
		mu.Lock()
		if err != nil {
			errors = append(errors, fmt.Errorf("technical analysis: %w", err))
		} else {
			prediction.TechnicalScore = technicalScore
		}
		mu.Unlock()
	}()

	// Macro analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		macroScore, err := mp.macroAnalyzer.AnalyzeMacro(ctx, asset)
		mu.Lock()
		if err != nil {
			errors = append(errors, fmt.Errorf("macro analysis: %w", err))
		} else {
			prediction.MacroScore = macroScore
		}
		mu.Unlock()
	}()

	// Wait for all analyses to complete
	wg.Wait()

	// Check for errors
	if len(errors) > 0 {
		mp.logger.Warn("Some prediction analyses failed", zap.Int("error_count", len(errors)))
		for _, err := range errors {
			mp.logger.Warn("Prediction analysis error", zap.Error(err))
		}
	}

	// Generate ensemble prediction
	ensembleResult, err := mp.ensembleModel.GeneratePrediction(ctx, prediction)
	if err != nil {
		mp.logger.Error("Ensemble prediction failed", zap.Error(err))
		// Continue with individual scores
	} else {
		prediction.EnsembleScore = ensembleResult.Score
		prediction.PredictedPrice = ensembleResult.PredictedPrice
		prediction.Confidence = ensembleResult.Confidence
	}

	// Calculate derived metrics
	mp.calculateDerivedMetrics(prediction)

	// Generate prediction signals
	prediction.Signals = mp.generatePredictionSignals(prediction)

	// Generate prediction factors
	prediction.Factors = mp.generatePredictionFactors(prediction)

	// Generate prediction scenarios
	prediction.Scenarios = mp.generatePredictionScenarios(prediction)

	// Generate prediction alerts
	prediction.Alerts = mp.generatePredictionAlerts(prediction)

	// Cache the prediction
	mp.cachePrediction(cacheKey, prediction)

	mp.logger.Info("Market prediction completed",
		zap.String("asset", asset),
		zap.String("predicted_price", prediction.PredictedPrice.String()),
		zap.String("direction", prediction.Direction),
		zap.String("confidence", prediction.Confidence.String()),
		zap.Duration("duration", time.Since(startTime)))

	return prediction, nil
}

// calculateDerivedMetrics calculates derived prediction metrics
func (mp *MarketPredictor) calculateDerivedMetrics(prediction *MarketPrediction) {
	// Calculate price change
	if !prediction.PredictedPrice.IsZero() && !prediction.CurrentPrice.IsZero() {
		prediction.PriceChange = prediction.PredictedPrice.Sub(prediction.CurrentPrice)
		prediction.PriceChangePercent = prediction.PriceChange.Div(prediction.CurrentPrice).Mul(decimal.NewFromFloat(100))
	}

	// Determine direction
	if prediction.PriceChangePercent.GreaterThan(decimal.NewFromFloat(2)) {
		prediction.Direction = "bullish"
	} else if prediction.PriceChangePercent.LessThan(decimal.NewFromFloat(-2)) {
		prediction.Direction = "bearish"
	} else {
		prediction.Direction = "neutral"
	}

	// Calculate overall confidence if not set by ensemble
	if prediction.Confidence.IsZero() {
		scores := []decimal.Decimal{}
		if !prediction.SentimentScore.IsZero() {
			scores = append(scores, prediction.SentimentScore)
		}
		if !prediction.OnChainScore.IsZero() {
			scores = append(scores, prediction.OnChainScore)
		}
		if !prediction.TechnicalScore.IsZero() {
			scores = append(scores, prediction.TechnicalScore)
		}
		if !prediction.MacroScore.IsZero() {
			scores = append(scores, prediction.MacroScore)
		}

		if len(scores) > 0 {
			total := decimal.Zero
			for _, score := range scores {
				total = total.Add(score)
			}
			prediction.Confidence = total.Div(decimal.NewFromInt(int64(len(scores))))
		} else {
			prediction.Confidence = decimal.NewFromFloat(0.5) // Default neutral confidence
		}
	}

	// Determine risk level
	prediction.RiskLevel = mp.determineRiskLevel(prediction)

	// Calculate volatility estimate
	prediction.Volatility = mp.calculateVolatilityEstimate(prediction)
}

// determineRiskLevel determines the risk level based on prediction metrics
func (mp *MarketPredictor) determineRiskLevel(prediction *MarketPrediction) string {
	volatilityThreshold := decimal.NewFromFloat(0.3) // 30%
	confidenceThreshold := decimal.NewFromFloat(0.7) // 70%

	if prediction.Volatility.GreaterThan(volatilityThreshold) {
		return "high"
	}

	if prediction.Confidence.LessThan(confidenceThreshold) {
		return "medium"
	}

	if prediction.PriceChangePercent.Abs().GreaterThan(decimal.NewFromFloat(10)) {
		return "medium"
	}

	return "low"
}

// calculateVolatilityEstimate calculates volatility estimate
func (mp *MarketPredictor) calculateVolatilityEstimate(prediction *MarketPrediction) decimal.Decimal {
	// Simple volatility estimate based on prediction uncertainty
	baseVolatility := decimal.NewFromFloat(0.15) // 15% base volatility

	// Adjust based on confidence (lower confidence = higher volatility)
	confidenceAdjustment := decimal.NewFromFloat(1).Sub(prediction.Confidence)
	volatility := baseVolatility.Add(confidenceAdjustment.Mul(decimal.NewFromFloat(0.2)))

	// Adjust based on predicted price change magnitude
	priceChangeAdjustment := prediction.PriceChangePercent.Abs().Div(decimal.NewFromFloat(100))
	volatility = volatility.Add(priceChangeAdjustment.Mul(decimal.NewFromFloat(0.1)))

	return volatility
}

// generatePredictionSignals generates prediction signals
func (mp *MarketPredictor) generatePredictionSignals(prediction *MarketPrediction) []*PredictionSignal {
	var signals []*PredictionSignal

	// Sentiment signals
	if !prediction.SentimentScore.IsZero() {
		direction := "neutral"
		if prediction.SentimentScore.GreaterThan(decimal.NewFromFloat(0.6)) {
			direction = "bullish"
		} else if prediction.SentimentScore.LessThan(decimal.NewFromFloat(0.4)) {
			direction = "bearish"
		}

		signals = append(signals, &PredictionSignal{
			Type:        "sentiment",
			Source:      "sentiment_analyzer",
			Strength:    prediction.SentimentScore,
			Direction:   direction,
			Confidence:  prediction.SentimentScore,
			Description: fmt.Sprintf("Market sentiment analysis indicates %s sentiment", direction),
			Timestamp:   time.Now(),
		})
	}

	// Technical signals
	if !prediction.TechnicalScore.IsZero() {
		direction := "neutral"
		if prediction.TechnicalScore.GreaterThan(decimal.NewFromFloat(0.6)) {
			direction = "bullish"
		} else if prediction.TechnicalScore.LessThan(decimal.NewFromFloat(0.4)) {
			direction = "bearish"
		}

		signals = append(signals, &PredictionSignal{
			Type:        "technical",
			Source:      "technical_analyzer",
			Strength:    prediction.TechnicalScore,
			Direction:   direction,
			Confidence:  prediction.TechnicalScore,
			Description: fmt.Sprintf("Technical analysis shows %s signals", direction),
			Timestamp:   time.Now(),
		})
	}

	// On-chain signals
	if !prediction.OnChainScore.IsZero() {
		direction := "neutral"
		if prediction.OnChainScore.GreaterThan(decimal.NewFromFloat(0.6)) {
			direction = "bullish"
		} else if prediction.OnChainScore.LessThan(decimal.NewFromFloat(0.4)) {
			direction = "bearish"
		}

		signals = append(signals, &PredictionSignal{
			Type:        "onchain",
			Source:      "onchain_analyzer",
			Strength:    prediction.OnChainScore,
			Direction:   direction,
			Confidence:  prediction.OnChainScore,
			Description: fmt.Sprintf("On-chain metrics indicate %s fundamentals", direction),
			Timestamp:   time.Now(),
		})
	}

	return signals
}

// generatePredictionFactors generates prediction factors
func (mp *MarketPredictor) generatePredictionFactors(prediction *MarketPrediction) []*PredictionFactor {
	var factors []*PredictionFactor

	// Sentiment factor
	if !prediction.SentimentScore.IsZero() {
		factors = append(factors, &PredictionFactor{
			Name:        "Market Sentiment",
			Category:    "sentiment",
			Impact:      prediction.SentimentScore.Sub(decimal.NewFromFloat(0.5)).Mul(decimal.NewFromFloat(2)),
			Weight:      mp.config.SentimentConfig.SentimentWeight,
			Value:       prediction.SentimentScore,
			Description: "Overall market sentiment from news, social media, and expert opinions",
		})
	}

	// Technical factor
	if !prediction.TechnicalScore.IsZero() {
		factors = append(factors, &PredictionFactor{
			Name:        "Technical Indicators",
			Category:    "technical",
			Impact:      prediction.TechnicalScore.Sub(decimal.NewFromFloat(0.5)).Mul(decimal.NewFromFloat(2)),
			Weight:      mp.config.TechnicalConfig.TrendWeight,
			Value:       prediction.TechnicalScore,
			Description: "Technical analysis indicators including trend, momentum, and volume",
		})
	}

	// On-chain factor
	if !prediction.OnChainScore.IsZero() {
		factors = append(factors, &PredictionFactor{
			Name:        "On-Chain Metrics",
			Category:    "onchain",
			Impact:      prediction.OnChainScore.Sub(decimal.NewFromFloat(0.5)).Mul(decimal.NewFromFloat(2)),
			Weight:      mp.config.OnChainConfig.TransactionWeight,
			Value:       prediction.OnChainScore,
			Description: "Blockchain metrics including transaction volume, active addresses, and network health",
		})
	}

	// Macro factor
	if !prediction.MacroScore.IsZero() {
		factors = append(factors, &PredictionFactor{
			Name:        "Macro Environment",
			Category:    "macro",
			Impact:      prediction.MacroScore.Sub(decimal.NewFromFloat(0.5)).Mul(decimal.NewFromFloat(2)),
			Weight:      mp.config.MacroConfig.EconomicWeight,
			Value:       prediction.MacroScore,
			Description: "Macroeconomic factors including monetary policy, regulation, and global events",
		})
	}

	return factors
}

// generatePredictionScenarios generates prediction scenarios
func (mp *MarketPredictor) generatePredictionScenarios(prediction *MarketPrediction) []*PredictionScenario {
	var scenarios []*PredictionScenario

	// Bull scenario
	bullProbability := prediction.Confidence.Mul(decimal.NewFromFloat(0.4))
	if prediction.Direction == "bullish" {
		bullProbability = prediction.Confidence.Mul(decimal.NewFromFloat(0.7))
	}

	bullTarget := prediction.CurrentPrice.Mul(decimal.NewFromFloat(1.15)) // 15% upside
	scenarios = append(scenarios, &PredictionScenario{
		Name:              "Bull Case",
		Probability:       bullProbability,
		PriceTarget:       bullTarget,
		TimeToTarget:      prediction.PredictionHorizon,
		TriggerConditions: []string{"Positive sentiment", "Strong on-chain metrics", "Favorable macro environment"},
		RiskFactors:       []string{"Market correction", "Regulatory uncertainty"},
	})

	// Bear scenario
	bearProbability := prediction.Confidence.Mul(decimal.NewFromFloat(0.4))
	if prediction.Direction == "bearish" {
		bearProbability = prediction.Confidence.Mul(decimal.NewFromFloat(0.7))
	}

	bearTarget := prediction.CurrentPrice.Mul(decimal.NewFromFloat(0.85)) // 15% downside
	scenarios = append(scenarios, &PredictionScenario{
		Name:              "Bear Case",
		Probability:       bearProbability,
		PriceTarget:       bearTarget,
		TimeToTarget:      prediction.PredictionHorizon,
		TriggerConditions: []string{"Negative sentiment", "Weak on-chain metrics", "Unfavorable macro environment"},
		RiskFactors:       []string{"Economic downturn", "Technical breakdown"},
	})

	// Base scenario
	baseProbability := decimal.NewFromFloat(1).Sub(bullProbability).Sub(bearProbability)
	if baseProbability.LessThan(decimal.NewFromFloat(0.1)) {
		baseProbability = decimal.NewFromFloat(0.1)
	}

	scenarios = append(scenarios, &PredictionScenario{
		Name:              "Base Case",
		Probability:       baseProbability,
		PriceTarget:       prediction.PredictedPrice,
		TimeToTarget:      prediction.PredictionHorizon,
		TriggerConditions: []string{"Current trends continue", "No major market disruptions"},
		RiskFactors:       []string{"Market volatility", "External events"},
	})

	return scenarios
}

// Stop stops the market predictor
func (mp *MarketPredictor) Stop() error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	if !mp.isRunning {
		return nil
	}

	mp.logger.Info("Stopping market predictor")

	// Stop monitoring
	if mp.updateTicker != nil {
		mp.updateTicker.Stop()
	}
	close(mp.stopChan)

	// Stop prediction engines
	if mp.ensembleModel != nil {
		mp.ensembleModel.Stop()
	}
	if mp.macroAnalyzer != nil {
		mp.macroAnalyzer.Stop()
	}
	if mp.technicalAnalyzer != nil {
		mp.technicalAnalyzer.Stop()
	}
	if mp.onChainAnalyzer != nil {
		mp.onChainAnalyzer.Stop()
	}
	if mp.sentimentAnalyzer != nil {
		mp.sentimentAnalyzer.Stop()
	}

	mp.isRunning = false
	mp.logger.Info("Market predictor stopped")
	return nil
}

// generatePredictionAlerts generates prediction alerts
func (mp *MarketPredictor) generatePredictionAlerts(prediction *MarketPrediction) []*PredictionAlert {
	alerts := make([]*PredictionAlert, 0)

	// High confidence bullish alert
	if prediction.Direction == "bullish" && prediction.Confidence.GreaterThan(mp.config.AlertThresholds.HighConfidenceBull) {
		alerts = append(alerts, &PredictionAlert{
			Type:                 "high_confidence_bull",
			Severity:             "info",
			Title:                "High Confidence Bullish Signal",
			Message:              fmt.Sprintf("Strong bullish prediction for %s with %.1f%% confidence", prediction.Asset, prediction.Confidence.Mul(decimal.NewFromFloat(100)).InexactFloat64()),
			Confidence:           prediction.Confidence,
			ActionRecommendation: "Consider increasing position and monitor for entry opportunities",
			ExpiresAt:            time.Now().Add(24 * time.Hour),
		})
	}

	// High confidence bearish alert
	if prediction.Direction == "bearish" && prediction.Confidence.GreaterThan(mp.config.AlertThresholds.HighConfidenceBear) {
		alerts = append(alerts, &PredictionAlert{
			Type:                 "high_confidence_bear",
			Severity:             "warning",
			Title:                "High Confidence Bearish Signal",
			Message:              fmt.Sprintf("Strong bearish prediction for %s with %.1f%% confidence", prediction.Asset, prediction.Confidence.Mul(decimal.NewFromFloat(100)).InexactFloat64()),
			Confidence:           prediction.Confidence,
			ActionRecommendation: "Consider reducing position and implement stop losses",
			ExpiresAt:            time.Now().Add(24 * time.Hour),
		})
	}

	// Low confidence alert
	if prediction.Confidence.LessThan(mp.config.AlertThresholds.LowConfidence) {
		alerts = append(alerts, &PredictionAlert{
			Type:                 "low_confidence",
			Severity:             "info",
			Title:                "Low Confidence Prediction",
			Message:              fmt.Sprintf("Prediction confidence for %s is low (%.1f%%)", prediction.Asset, prediction.Confidence.Mul(decimal.NewFromFloat(100)).InexactFloat64()),
			Confidence:           prediction.Confidence,
			ActionRecommendation: "Wait for clearer signals and reduce position size",
			ExpiresAt:            time.Now().Add(12 * time.Hour),
		})
	}

	// High volatility alert
	if prediction.Volatility.GreaterThan(mp.config.AlertThresholds.VolatilitySpike) {
		alerts = append(alerts, &PredictionAlert{
			Type:                 "volatility_spike",
			Severity:             "warning",
			Title:                "High Volatility Expected",
			Message:              fmt.Sprintf("High volatility predicted for %s (%.1f%%)", prediction.Asset, prediction.Volatility.Mul(decimal.NewFromFloat(100)).InexactFloat64()),
			Confidence:           prediction.Confidence,
			ActionRecommendation: "Adjust position sizing and implement risk management",
			ExpiresAt:            time.Now().Add(6 * time.Hour),
		})
	}

	return alerts
}

// Utility methods

// monitoringLoop runs the main monitoring loop
func (mp *MarketPredictor) monitoringLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-mp.stopChan:
			return
		case <-mp.updateTicker.C:
			mp.performMaintenance()
		}
	}
}

// performMaintenance performs periodic maintenance tasks
func (mp *MarketPredictor) performMaintenance() {
	mp.logger.Debug("Performing market predictor maintenance")

	// Clean up expired cache entries
	mp.cleanupExpiredCache()

	// Update model performance metrics
	mp.updateModelMetrics()
}

// cleanupExpiredCache removes expired cache entries
func (mp *MarketPredictor) cleanupExpiredCache() {
	mp.cacheMutex.Lock()
	defer mp.cacheMutex.Unlock()

	now := time.Now()
	for key, prediction := range mp.predictionCache {
		if now.Sub(prediction.Timestamp) > mp.config.CacheTimeout {
			delete(mp.predictionCache, key)
		}
	}
}

// updateModelMetrics updates model performance metrics
func (mp *MarketPredictor) updateModelMetrics() {
	mp.modelMutex.Lock()
	defer mp.modelMutex.Unlock()

	// Mock model metrics update - in production, calculate actual performance
	mp.logger.Debug("Updating model performance metrics")
}

// generateCacheKey generates a cache key for predictions
func (mp *MarketPredictor) generateCacheKey(asset string) string {
	return fmt.Sprintf("%s_%d", asset, time.Now().Truncate(mp.config.UpdateInterval).Unix())
}

// generatePredictionID generates a unique prediction ID
func (mp *MarketPredictor) generatePredictionID() string {
	return fmt.Sprintf("pred_%d", time.Now().UnixNano())
}

// getCachedPrediction retrieves cached prediction
func (mp *MarketPredictor) getCachedPrediction(key string) *MarketPrediction {
	mp.cacheMutex.RLock()
	defer mp.cacheMutex.RUnlock()

	prediction, exists := mp.predictionCache[key]
	if !exists {
		return nil
	}

	// Check if cache entry is still valid
	if time.Since(prediction.Timestamp) > mp.config.CacheTimeout {
		delete(mp.predictionCache, key)
		return nil
	}

	return prediction
}

// cachePrediction caches prediction
func (mp *MarketPredictor) cachePrediction(key string, prediction *MarketPrediction) {
	mp.cacheMutex.Lock()
	defer mp.cacheMutex.Unlock()
	mp.predictionCache[key] = prediction
}

// IsRunning returns whether the predictor is running
func (mp *MarketPredictor) IsRunning() bool {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	return mp.isRunning
}

// GetPredictionMetrics returns prediction metrics
func (mp *MarketPredictor) GetPredictionMetrics() map[string]interface{} {
	mp.cacheMutex.RLock()
	defer mp.cacheMutex.RUnlock()

	return map[string]interface{}{
		"cached_predictions": len(mp.predictionCache),
		"is_running":         mp.IsRunning(),
		"sentiment_analyzer": mp.sentimentAnalyzer != nil,
		"onchain_analyzer":   mp.onChainAnalyzer != nil,
		"technical_analyzer": mp.technicalAnalyzer != nil,
		"macro_analyzer":     mp.macroAnalyzer != nil,
		"ensemble_model":     mp.ensembleModel != nil,
		"cached_models":      len(mp.modelCache),
	}
}

// GetModelPerformance returns model performance metrics
func (mp *MarketPredictor) GetModelPerformance(modelID string) (*ModelPerformance, error) {
	mp.modelMutex.RLock()
	defer mp.modelMutex.RUnlock()

	model, exists := mp.modelCache[modelID]
	if !exists {
		return nil, fmt.Errorf("model not found: %s", modelID)
	}

	return model.Performance, nil
}

// PredictMultipleAssets generates predictions for multiple assets
func (mp *MarketPredictor) PredictMultipleAssets(ctx context.Context, assets map[string]decimal.Decimal) (map[string]*MarketPrediction, error) {
	predictions := make(map[string]*MarketPrediction)

	for asset, price := range assets {
		prediction, err := mp.PredictMarket(ctx, asset, price)
		if err != nil {
			mp.logger.Warn("Failed to predict asset", zap.String("asset", asset), zap.Error(err))
			continue
		}
		predictions[asset] = prediction
	}

	return predictions, nil
}
