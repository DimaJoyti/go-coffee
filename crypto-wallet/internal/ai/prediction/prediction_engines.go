package prediction

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// SentimentAnalyzer analyzes market sentiment from various sources
type SentimentAnalyzer struct {
	logger *logger.Logger
	config SentimentAnalysisConfig
}

// NewSentimentAnalyzer creates a new sentiment analyzer
func NewSentimentAnalyzer(logger *logger.Logger, config SentimentAnalysisConfig) *SentimentAnalyzer {
	return &SentimentAnalyzer{
		logger: logger.Named("sentiment-analyzer"),
		config: config,
	}
}

// Start starts the sentiment analyzer
func (sa *SentimentAnalyzer) Start(ctx context.Context) error {
	if !sa.config.Enabled {
		sa.logger.Info("Sentiment analyzer is disabled")
		return nil
	}
	sa.logger.Info("Starting sentiment analyzer")
	return nil
}

// Stop stops the sentiment analyzer
func (sa *SentimentAnalyzer) Stop() error {
	sa.logger.Info("Stopping sentiment analyzer")
	return nil
}

// AnalyzeSentiment analyzes sentiment for a given asset
func (sa *SentimentAnalyzer) AnalyzeSentiment(ctx context.Context, asset string) (decimal.Decimal, error) {
	sa.logger.Debug("Analyzing sentiment", zap.String("asset", asset))

	// Mock sentiment analysis - in production, integrate with real sentiment APIs
	sentimentScore := sa.calculateMockSentiment(asset)

	sa.logger.Info("Sentiment analysis completed",
		zap.String("asset", asset),
		zap.String("sentiment_score", sentimentScore.String()))

	return sentimentScore, nil
}

// calculateMockSentiment calculates mock sentiment based on asset characteristics
func (sa *SentimentAnalyzer) calculateMockSentiment(asset string) decimal.Decimal {
	// Mock sentiment based on asset name
	asset = strings.ToLower(asset)

	baseScore := decimal.NewFromFloat(0.5) // Neutral base

	// Positive sentiment assets
	if strings.Contains(asset, "btc") || strings.Contains(asset, "bitcoin") {
		baseScore = decimal.NewFromFloat(0.65)
	} else if strings.Contains(asset, "eth") || strings.Contains(asset, "ethereum") {
		baseScore = decimal.NewFromFloat(0.6)
	} else if strings.Contains(asset, "sol") || strings.Contains(asset, "solana") {
		baseScore = decimal.NewFromFloat(0.55)
	}

	// Add some randomness to simulate real sentiment fluctuations
	variation := decimal.NewFromFloat((float64(time.Now().Unix()%100) - 50) / 1000) // ±0.05
	sentimentScore := baseScore.Add(variation)

	// Ensure score is between 0 and 1
	if sentimentScore.LessThan(decimal.Zero) {
		sentimentScore = decimal.Zero
	} else if sentimentScore.GreaterThan(decimal.NewFromFloat(1)) {
		sentimentScore = decimal.NewFromFloat(1)
	}

	return sentimentScore
}

// OnChainAnalyzer analyzes on-chain metrics
type OnChainAnalyzer struct {
	logger *logger.Logger
	config OnChainAnalysisConfig
}

// NewOnChainAnalyzer creates a new on-chain analyzer
func NewOnChainAnalyzer(logger *logger.Logger, config OnChainAnalysisConfig) *OnChainAnalyzer {
	return &OnChainAnalyzer{
		logger: logger.Named("onchain-analyzer"),
		config: config,
	}
}

// Start starts the on-chain analyzer
func (oca *OnChainAnalyzer) Start(ctx context.Context) error {
	if !oca.config.Enabled {
		oca.logger.Info("On-chain analyzer is disabled")
		return nil
	}
	oca.logger.Info("Starting on-chain analyzer")
	return nil
}

// Stop stops the on-chain analyzer
func (oca *OnChainAnalyzer) Stop() error {
	oca.logger.Info("Stopping on-chain analyzer")
	return nil
}

// AnalyzeOnChain analyzes on-chain metrics for a given asset
func (oca *OnChainAnalyzer) AnalyzeOnChain(ctx context.Context, asset string) (decimal.Decimal, error) {
	oca.logger.Debug("Analyzing on-chain metrics", zap.String("asset", asset))

	// Mock on-chain analysis - in production, integrate with blockchain APIs
	onChainScore := oca.calculateMockOnChainScore(asset)

	oca.logger.Info("On-chain analysis completed",
		zap.String("asset", asset),
		zap.String("onchain_score", onChainScore.String()))

	return onChainScore, nil
}

// calculateMockOnChainScore calculates mock on-chain score
func (oca *OnChainAnalyzer) calculateMockOnChainScore(asset string) decimal.Decimal {
	asset = strings.ToLower(asset)

	baseScore := decimal.NewFromFloat(0.5) // Neutral base

	// Different on-chain strength for different assets
	if strings.Contains(asset, "btc") || strings.Contains(asset, "bitcoin") {
		baseScore = decimal.NewFromFloat(0.7) // Strong on-chain metrics
	} else if strings.Contains(asset, "eth") || strings.Contains(asset, "ethereum") {
		baseScore = decimal.NewFromFloat(0.75) // Very strong on-chain metrics
	} else if strings.Contains(asset, "usdc") || strings.Contains(asset, "usdt") {
		baseScore = decimal.NewFromFloat(0.8) // Stable on-chain metrics
	}

	// Add time-based variation
	timeVariation := decimal.NewFromFloat(math.Sin(float64(time.Now().Unix())/86400) * 0.1) // Daily cycle
	onChainScore := baseScore.Add(timeVariation)

	// Ensure score is between 0 and 1
	if onChainScore.LessThan(decimal.Zero) {
		onChainScore = decimal.Zero
	} else if onChainScore.GreaterThan(decimal.NewFromFloat(1)) {
		onChainScore = decimal.NewFromFloat(1)
	}

	return onChainScore
}

// TechnicalAnalyzer analyzes technical indicators
type TechnicalAnalyzer struct {
	logger *logger.Logger
	config TechnicalAnalysisConfig
}

// NewTechnicalAnalyzer creates a new technical analyzer
func NewTechnicalAnalyzer(logger *logger.Logger, config TechnicalAnalysisConfig) *TechnicalAnalyzer {
	return &TechnicalAnalyzer{
		logger: logger.Named("technical-analyzer"),
		config: config,
	}
}

// Start starts the technical analyzer
func (ta *TechnicalAnalyzer) Start(ctx context.Context) error {
	if !ta.config.Enabled {
		ta.logger.Info("Technical analyzer is disabled")
		return nil
	}
	ta.logger.Info("Starting technical analyzer")
	return nil
}

// Stop stops the technical analyzer
func (ta *TechnicalAnalyzer) Stop() error {
	ta.logger.Info("Stopping technical analyzer")
	return nil
}

// AnalyzeTechnical analyzes technical indicators for a given asset
func (ta *TechnicalAnalyzer) AnalyzeTechnical(ctx context.Context, asset string, currentPrice decimal.Decimal) (decimal.Decimal, error) {
	ta.logger.Debug("Analyzing technical indicators",
		zap.String("asset", asset),
		zap.String("current_price", currentPrice.String()))

	// Mock technical analysis - in production, integrate with price data and indicators
	technicalScore := ta.calculateMockTechnicalScore(asset, currentPrice)

	ta.logger.Info("Technical analysis completed",
		zap.String("asset", asset),
		zap.String("technical_score", technicalScore.String()))

	return technicalScore, nil
}

// calculateMockTechnicalScore calculates mock technical score
func (ta *TechnicalAnalyzer) calculateMockTechnicalScore(asset string, currentPrice decimal.Decimal) decimal.Decimal {
	// Mock technical analysis based on price patterns
	priceFloat, _ := currentPrice.Float64()

	// Use price to simulate different technical conditions
	priceHash := int(priceFloat) % 100

	baseScore := decimal.NewFromFloat(0.5)

	// Simulate trend analysis
	if priceHash > 60 {
		baseScore = decimal.NewFromFloat(0.65) // Bullish trend
	} else if priceHash < 40 {
		baseScore = decimal.NewFromFloat(0.35) // Bearish trend
	}

	// Add momentum component
	momentumAdjustment := decimal.NewFromFloat(float64(priceHash-50) / 500) // ±0.1
	technicalScore := baseScore.Add(momentumAdjustment)

	// Ensure score is between 0 and 1
	if technicalScore.LessThan(decimal.Zero) {
		technicalScore = decimal.Zero
	} else if technicalScore.GreaterThan(decimal.NewFromFloat(1)) {
		technicalScore = decimal.NewFromFloat(1)
	}

	return technicalScore
}

// MacroAnalyzer analyzes macroeconomic factors
type MacroAnalyzer struct {
	logger *logger.Logger
	config MacroAnalysisConfig
}

// NewMacroAnalyzer creates a new macro analyzer
func NewMacroAnalyzer(logger *logger.Logger, config MacroAnalysisConfig) *MacroAnalyzer {
	return &MacroAnalyzer{
		logger: logger.Named("macro-analyzer"),
		config: config,
	}
}

// Start starts the macro analyzer
func (ma *MacroAnalyzer) Start(ctx context.Context) error {
	if !ma.config.Enabled {
		ma.logger.Info("Macro analyzer is disabled")
		return nil
	}
	ma.logger.Info("Starting macro analyzer")
	return nil
}

// Stop stops the macro analyzer
func (ma *MacroAnalyzer) Stop() error {
	ma.logger.Info("Stopping macro analyzer")
	return nil
}

// AnalyzeMacro analyzes macroeconomic factors for a given asset
func (ma *MacroAnalyzer) AnalyzeMacro(ctx context.Context, asset string) (decimal.Decimal, error) {
	ma.logger.Debug("Analyzing macro factors", zap.String("asset", asset))

	// Mock macro analysis - in production, integrate with economic data APIs
	macroScore := ma.calculateMockMacroScore(asset)

	ma.logger.Info("Macro analysis completed",
		zap.String("asset", asset),
		zap.String("macro_score", macroScore.String()))

	return macroScore, nil
}

// calculateMockMacroScore calculates mock macro score
func (ma *MacroAnalyzer) calculateMockMacroScore(asset string) decimal.Decimal {
	// Mock macro analysis - generally favorable for crypto
	baseScore := decimal.NewFromFloat(0.6) // Slightly positive macro environment

	// Add weekly cycle to simulate changing macro conditions
	weekCycle := math.Sin(float64(time.Now().Unix())/604800) * 0.15 // Weekly cycle ±0.15
	macroScore := baseScore.Add(decimal.NewFromFloat(weekCycle))

	// Ensure score is between 0 and 1
	if macroScore.LessThan(decimal.Zero) {
		macroScore = decimal.Zero
	} else if macroScore.GreaterThan(decimal.NewFromFloat(1)) {
		macroScore = decimal.NewFromFloat(1)
	}

	return macroScore
}

// EnsembleModel combines predictions from multiple models
type EnsembleModel struct {
	logger *logger.Logger
	config EnsembleModelConfig
}

// EnsembleResult represents ensemble model result
type EnsembleResult struct {
	Score          decimal.Decimal            `json:"score"`
	PredictedPrice decimal.Decimal            `json:"predicted_price"`
	Confidence     decimal.Decimal            `json:"confidence"`
	ModelWeights   map[string]decimal.Decimal `json:"model_weights"`
}

// NewEnsembleModel creates a new ensemble model
func NewEnsembleModel(logger *logger.Logger, config EnsembleModelConfig) *EnsembleModel {
	return &EnsembleModel{
		logger: logger.Named("ensemble-model"),
		config: config,
	}
}

// Start starts the ensemble model
func (em *EnsembleModel) Start(ctx context.Context) error {
	if !em.config.Enabled {
		em.logger.Info("Ensemble model is disabled")
		return nil
	}
	em.logger.Info("Starting ensemble model")
	return nil
}

// Stop stops the ensemble model
func (em *EnsembleModel) Stop() error {
	em.logger.Info("Stopping ensemble model")
	return nil
}

// GeneratePrediction generates ensemble prediction
func (em *EnsembleModel) GeneratePrediction(ctx context.Context, prediction *MarketPrediction) (*EnsembleResult, error) {
	em.logger.Debug("Generating ensemble prediction", zap.String("asset", prediction.Asset))

	// Calculate weighted ensemble score
	weights := em.config.ModelWeights
	if len(weights) == 0 {
		// Default equal weights
		weights = map[string]decimal.Decimal{
			"sentiment": decimal.NewFromFloat(0.25),
			"onchain":   decimal.NewFromFloat(0.25),
			"technical": decimal.NewFromFloat(0.25),
			"macro":     decimal.NewFromFloat(0.25),
		}
	}

	totalScore := decimal.Zero
	totalWeight := decimal.Zero

	// Sentiment component
	if !prediction.SentimentScore.IsZero() {
		if weight, exists := weights["sentiment"]; exists {
			totalScore = totalScore.Add(prediction.SentimentScore.Mul(weight))
			totalWeight = totalWeight.Add(weight)
		}
	}

	// On-chain component
	if !prediction.OnChainScore.IsZero() {
		if weight, exists := weights["onchain"]; exists {
			totalScore = totalScore.Add(prediction.OnChainScore.Mul(weight))
			totalWeight = totalWeight.Add(weight)
		}
	}

	// Technical component
	if !prediction.TechnicalScore.IsZero() {
		if weight, exists := weights["technical"]; exists {
			totalScore = totalScore.Add(prediction.TechnicalScore.Mul(weight))
			totalWeight = totalWeight.Add(weight)
		}
	}

	// Macro component
	if !prediction.MacroScore.IsZero() {
		if weight, exists := weights["macro"]; exists {
			totalScore = totalScore.Add(prediction.MacroScore.Mul(weight))
			totalWeight = totalWeight.Add(weight)
		}
	}

	// Calculate final ensemble score
	ensembleScore := decimal.NewFromFloat(0.5) // Default neutral
	if !totalWeight.IsZero() {
		ensembleScore = totalScore.Div(totalWeight)
	}

	// Generate price prediction based on ensemble score
	priceMultiplier := decimal.NewFromFloat(1.0)
	if ensembleScore.GreaterThan(decimal.NewFromFloat(0.6)) {
		// Bullish prediction
		priceMultiplier = decimal.NewFromFloat(1.0).Add(ensembleScore.Sub(decimal.NewFromFloat(0.5)).Mul(decimal.NewFromFloat(0.3)))
	} else if ensembleScore.LessThan(decimal.NewFromFloat(0.4)) {
		// Bearish prediction
		priceMultiplier = decimal.NewFromFloat(1.0).Sub(decimal.NewFromFloat(0.5).Sub(ensembleScore).Mul(decimal.NewFromFloat(0.3)))
	}

	predictedPrice := prediction.CurrentPrice.Mul(priceMultiplier)

	// Calculate confidence based on score consistency
	confidence := em.calculateConfidence(prediction, ensembleScore)

	result := &EnsembleResult{
		Score:          ensembleScore,
		PredictedPrice: predictedPrice,
		Confidence:     confidence,
		ModelWeights:   weights,
	}

	em.logger.Info("Ensemble prediction completed",
		zap.String("asset", prediction.Asset),
		zap.String("ensemble_score", ensembleScore.String()),
		zap.String("predicted_price", predictedPrice.String()),
		zap.String("confidence", confidence.String()))

	return result, nil
}

// calculateConfidence calculates prediction confidence
func (em *EnsembleModel) calculateConfidence(prediction *MarketPrediction, ensembleScore decimal.Decimal) decimal.Decimal {
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

	if len(scores) == 0 {
		return decimal.NewFromFloat(0.5)
	}

	// Calculate variance from ensemble score
	variance := decimal.Zero
	for _, score := range scores {
		diff := score.Sub(ensembleScore)
		variance = variance.Add(diff.Mul(diff))
	}
	variance = variance.Div(decimal.NewFromInt(int64(len(scores))))

	// Convert variance to confidence (lower variance = higher confidence)
	confidence := decimal.NewFromFloat(1.0).Sub(variance.Mul(decimal.NewFromFloat(2)))

	// Ensure confidence is between 0.1 and 1.0
	if confidence.LessThan(decimal.NewFromFloat(0.1)) {
		confidence = decimal.NewFromFloat(0.1)
	} else if confidence.GreaterThan(decimal.NewFromFloat(1.0)) {
		confidence = decimal.NewFromFloat(1.0)
	}

	return confidence
}
