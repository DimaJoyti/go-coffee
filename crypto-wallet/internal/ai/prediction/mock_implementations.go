package prediction

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// Mock implementations for testing and demonstration

// MockAdvancedSentimentEngine provides mock sentiment analysis
type MockAdvancedSentimentEngine struct{}

func (m *MockAdvancedSentimentEngine) AnalyzeSentiment(ctx context.Context, asset string) (decimal.Decimal, error) {
	// Mock sentiment score between 0 and 1
	return decimal.NewFromFloat(0.65), nil
}

func (m *MockAdvancedSentimentEngine) GetSentimentTrends(asset string, window time.Duration) ([]SentimentDataPoint, error) {
	return []SentimentDataPoint{
		{
			Timestamp: time.Now().Add(-1 * time.Hour),
			Score:     decimal.NewFromFloat(0.6),
			Volume:    1000,
			Source:    "twitter",
		},
		{
			Timestamp: time.Now(),
			Score:     decimal.NewFromFloat(0.7),
			Volume:    1200,
			Source:    "reddit",
		},
	}, nil
}

// MockAdvancedOnChainEngine provides mock on-chain analysis
type MockAdvancedOnChainEngine struct{}

func (m *MockAdvancedOnChainEngine) AnalyzeOnChain(ctx context.Context, asset string) (decimal.Decimal, error) {
	// Mock on-chain score between 0 and 1
	return decimal.NewFromFloat(0.72), nil
}

func (m *MockAdvancedOnChainEngine) GetOnChainMetrics(asset string) (*OnChainMetrics, error) {
	return &OnChainMetrics{
		TransactionVolume:  decimal.NewFromFloat(1000000),
		ActiveAddresses:    50000,
		NetworkUtilization: decimal.NewFromFloat(0.75),
		WhaleActivity:      decimal.NewFromFloat(0.3),
		DeFiTVL:           decimal.NewFromFloat(5000000),
	}, nil
}

// MockAdvancedTechnicalEngine provides mock technical analysis
type MockAdvancedTechnicalEngine struct{}

func (m *MockAdvancedTechnicalEngine) AnalyzeTechnical(ctx context.Context, asset string) (decimal.Decimal, error) {
	// Mock technical score between 0 and 1
	return decimal.NewFromFloat(0.58), nil
}

func (m *MockAdvancedTechnicalEngine) GetTechnicalIndicators(asset string) (*TechnicalIndicators, error) {
	return &TechnicalIndicators{
		RSI:  decimal.NewFromFloat(45),
		MACD: decimal.NewFromFloat(0.02),
		BollingerBands: PriceRange{
			Low:  decimal.NewFromFloat(45000),
			High: decimal.NewFromFloat(55000),
			Mean: decimal.NewFromFloat(50000),
		},
		SupportLevels:    []decimal.Decimal{decimal.NewFromFloat(48000), decimal.NewFromFloat(46000)},
		ResistanceLevels: []decimal.Decimal{decimal.NewFromFloat(52000), decimal.NewFromFloat(54000)},
		VolumeProfile:    decimal.NewFromFloat(0.8),
	}, nil
}

// MockAdvancedMacroEngine provides mock macro analysis
type MockAdvancedMacroEngine struct{}

func (m *MockAdvancedMacroEngine) AnalyzeMacro(ctx context.Context, asset string) (decimal.Decimal, error) {
	// Mock macro score between 0 and 1
	return decimal.NewFromFloat(0.55), nil
}

func (m *MockAdvancedMacroEngine) GetMacroFactors() (*MacroFactors, error) {
	return &MacroFactors{
		InterestRates:        decimal.NewFromFloat(0.05),
		InflationRate:        decimal.NewFromFloat(0.03),
		USDIndex:             decimal.NewFromFloat(102.5),
		StockMarketSentiment: decimal.NewFromFloat(0.6),
		GeopoliticalRisk:     decimal.NewFromFloat(0.4),
	}, nil
}

// MockMachineLearningEngine provides mock ML predictions
type MockMachineLearningEngine struct{}

func (m *MockMachineLearningEngine) Predict(ctx context.Context, features []decimal.Decimal) (*MLPrediction, error) {
	return &MLPrediction{
		Prediction: decimal.NewFromFloat(0.68),
		Confidence: decimal.NewFromFloat(0.85),
		FeatureImportance: map[string]decimal.Decimal{
			"price_momentum":    decimal.NewFromFloat(0.3),
			"volume_trend":      decimal.NewFromFloat(0.25),
			"sentiment_score":   decimal.NewFromFloat(0.2),
			"onchain_activity":  decimal.NewFromFloat(0.15),
			"macro_indicators":  decimal.NewFromFloat(0.1),
		},
	}, nil
}

func (m *MockMachineLearningEngine) TrainModel(ctx context.Context, data *TrainingData) error {
	// Mock training - would implement actual training logic
	return nil
}

func (m *MockMachineLearningEngine) GetModelPerformance() (*ModelPerformance, error) {
	return &ModelPerformance{
		Accuracy:  decimal.NewFromFloat(0.82),
		Precision: decimal.NewFromFloat(0.78),
		Recall:    decimal.NewFromFloat(0.85),
		F1Score:   decimal.NewFromFloat(0.81),
		MAE:       decimal.NewFromFloat(0.05),
		RMSE:      decimal.NewFromFloat(0.07),
	}, nil
}

// MockAdvancedEnsembleEngine provides mock ensemble predictions
type MockAdvancedEnsembleEngine struct{}

func (m *MockAdvancedEnsembleEngine) CombinePredictions(predictions map[string]decimal.Decimal) (*EnsemblePrediction, error) {
	// Simple weighted average for mock
	var weightedSum decimal.Decimal
	var totalWeight decimal.Decimal
	
	weights := map[string]decimal.Decimal{
		"sentiment": decimal.NewFromFloat(0.2),
		"onchain":   decimal.NewFromFloat(0.25),
		"technical": decimal.NewFromFloat(0.2),
		"macro":     decimal.NewFromFloat(0.15),
		"ml":        decimal.NewFromFloat(0.2),
	}
	
	for name, prediction := range predictions {
		if weight, exists := weights[name]; exists {
			weightedSum = weightedSum.Add(prediction.Mul(weight))
			totalWeight = totalWeight.Add(weight)
		}
	}
	
	if totalWeight.IsZero() {
		totalWeight = decimal.NewFromFloat(1)
	}
	
	finalPrediction := weightedSum.Div(totalWeight)
	
	return &EnsemblePrediction{
		Prediction:       finalPrediction,
		Confidence:       decimal.NewFromFloat(0.8),
		ComponentWeights: weights,
	}, nil
}

func (m *MockAdvancedEnsembleEngine) UpdateWeights(performance map[string]decimal.Decimal) error {
	// Mock weight update - would implement actual weight adjustment logic
	return nil
}

// MockDataAggregator provides mock data aggregation
type MockDataAggregator struct{}

func (m *MockDataAggregator) AggregateData(ctx context.Context, asset string) (*AggregatedData, error) {
	now := time.Now()
	
	// Mock price data
	priceData := []PricePoint{
		{
			Timestamp: now.Add(-2 * time.Hour),
			Open:      decimal.NewFromFloat(49000),
			High:      decimal.NewFromFloat(49500),
			Low:       decimal.NewFromFloat(48800),
			Close:     decimal.NewFromFloat(49200),
		},
		{
			Timestamp: now.Add(-1 * time.Hour),
			Open:      decimal.NewFromFloat(49200),
			High:      decimal.NewFromFloat(50000),
			Low:       decimal.NewFromFloat(49000),
			Close:     decimal.NewFromFloat(49800),
		},
		{
			Timestamp: now,
			Open:      decimal.NewFromFloat(49800),
			High:      decimal.NewFromFloat(50200),
			Low:       decimal.NewFromFloat(49600),
			Close:     decimal.NewFromFloat(50000),
		},
	}
	
	// Mock volume data
	volumeData := []VolumePoint{
		{
			Timestamp:  now.Add(-2 * time.Hour),
			Volume:     decimal.NewFromFloat(1000000),
			BuyVolume:  decimal.NewFromFloat(600000),
			SellVolume: decimal.NewFromFloat(400000),
		},
		{
			Timestamp:  now.Add(-1 * time.Hour),
			Volume:     decimal.NewFromFloat(1200000),
			BuyVolume:  decimal.NewFromFloat(700000),
			SellVolume: decimal.NewFromFloat(500000),
		},
		{
			Timestamp:  now,
			Volume:     decimal.NewFromFloat(800000),
			BuyVolume:  decimal.NewFromFloat(450000),
			SellVolume: decimal.NewFromFloat(350000),
		},
	}
	
	// Mock sentiment data
	sentimentData := []SentimentDataPoint{
		{
			Timestamp: now.Add(-1 * time.Hour),
			Score:     decimal.NewFromFloat(0.6),
			Volume:    1000,
			Source:    "twitter",
		},
		{
			Timestamp: now,
			Score:     decimal.NewFromFloat(0.7),
			Volume:    1200,
			Source:    "reddit",
		},
	}
	
	// Mock on-chain data
	onChainData := &OnChainMetrics{
		TransactionVolume:  decimal.NewFromFloat(1000000),
		ActiveAddresses:    50000,
		NetworkUtilization: decimal.NewFromFloat(0.75),
		WhaleActivity:      decimal.NewFromFloat(0.3),
		DeFiTVL:           decimal.NewFromFloat(5000000),
	}
	
	// Mock macro data
	macroData := &MacroFactors{
		InterestRates:        decimal.NewFromFloat(0.05),
		InflationRate:        decimal.NewFromFloat(0.03),
		USDIndex:             decimal.NewFromFloat(102.5),
		StockMarketSentiment: decimal.NewFromFloat(0.6),
		GeopoliticalRisk:     decimal.NewFromFloat(0.4),
	}
	
	return &AggregatedData{
		PriceData:     priceData,
		VolumeData:    volumeData,
		SentimentData: sentimentData,
		OnChainData:   onChainData,
		MacroData:     macroData,
	}, nil
}

func (m *MockDataAggregator) GetDataQuality(source string) decimal.Decimal {
	// Mock data quality score
	return decimal.NewFromFloat(0.9)
}

// MockFeatureExtractor provides mock feature extraction
type MockFeatureExtractor struct{}

func (m *MockFeatureExtractor) ExtractFeatures(ctx context.Context, data *AggregatedData) ([]decimal.Decimal, error) {
	// Mock feature vector
	features := []decimal.Decimal{
		decimal.NewFromFloat(0.1),  // price momentum
		decimal.NewFromFloat(0.2),  // volume trend
		decimal.NewFromFloat(0.65), // sentiment score
		decimal.NewFromFloat(0.72), // on-chain activity
		decimal.NewFromFloat(0.55), // macro indicators
		decimal.NewFromFloat(0.8),  // volatility
		decimal.NewFromFloat(0.3),  // correlation
		decimal.NewFromFloat(0.45), // technical indicators
	}
	
	return features, nil
}

func (m *MockFeatureExtractor) GetFeatureImportance() map[string]decimal.Decimal {
	return map[string]decimal.Decimal{
		"price_momentum":    decimal.NewFromFloat(0.3),
		"volume_trend":      decimal.NewFromFloat(0.25),
		"sentiment_score":   decimal.NewFromFloat(0.2),
		"onchain_activity":  decimal.NewFromFloat(0.15),
		"macro_indicators":  decimal.NewFromFloat(0.1),
	}
}

// MockModelManager provides mock model management
type MockModelManager struct{}

func (m *MockModelManager) GetBestModel(modelType string) (Model, error) {
	return &MockModel{}, nil
}

func (m *MockModelManager) UpdateModelPerformance(modelID string, performance decimal.Decimal) error {
	// Mock performance update
	return nil
}

func (m *MockModelManager) TriggerRetraining(modelType string) error {
	// Mock retraining trigger
	return nil
}

// MockModel provides mock model implementation
type MockModel struct{}

func (m *MockModel) Predict(features []decimal.Decimal) (decimal.Decimal, error) {
	// Mock prediction based on features
	return decimal.NewFromFloat(0.68), nil
}

func (m *MockModel) Train(data *TrainingData) error {
	// Mock training
	return nil
}

func (m *MockModel) GetPerformance() *ModelPerformance {
	return &ModelPerformance{
		Accuracy:  decimal.NewFromFloat(0.82),
		Precision: decimal.NewFromFloat(0.78),
		Recall:    decimal.NewFromFloat(0.85),
		F1Score:   decimal.NewFromFloat(0.81),
		MAE:       decimal.NewFromFloat(0.05),
		RMSE:      decimal.NewFromFloat(0.07),
	}
}
