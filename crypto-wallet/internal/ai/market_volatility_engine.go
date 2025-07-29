package ai

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/redis"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// MarketVolatilityEngine provides real-time market volatility analysis and position sizing recommendations
type MarketVolatilityEngine struct {
	logger *logger.Logger
	cache  redis.Client

	// Configuration
	config VolatilityConfig

	// State tracking
	assetVolatilities map[string]*AssetVolatility
	correlationMatrix map[string]map[string]decimal.Decimal
	marketRegimes     map[string]*MarketRegime
	positionSizes     map[string]*PositionSizeRecommendation
	volatilityAlerts  map[string]*VolatilityAlert
	historicalData    map[string]*HistoricalVolatilityData
	realTimeMetrics   *RealTimeVolatilityMetrics
	mutex             sync.RWMutex
	stopChan          chan struct{}
	isRunning         bool
}

// VolatilityConfig holds configuration for volatility analysis
type VolatilityConfig struct {
	Enabled                   bool            `json:"enabled" yaml:"enabled"`
	AnalysisInterval          time.Duration   `json:"analysis_interval" yaml:"analysis_interval"`
	CorrelationUpdateInterval time.Duration   `json:"correlation_update_interval" yaml:"correlation_update_interval"`
	VolatilityWindow          time.Duration   `json:"volatility_window" yaml:"volatility_window"`
	HighVolatilityThreshold   decimal.Decimal `json:"high_volatility_threshold" yaml:"high_volatility_threshold"`
	LowVolatilityThreshold    decimal.Decimal `json:"low_volatility_threshold" yaml:"low_volatility_threshold"`
	CorrelationThreshold      decimal.Decimal `json:"correlation_threshold" yaml:"correlation_threshold"`
	MaxPositionSize           decimal.Decimal `json:"max_position_size" yaml:"max_position_size"`
	MinPositionSize           decimal.Decimal `json:"min_position_size" yaml:"min_position_size"`
	RiskAdjustmentFactor      decimal.Decimal `json:"risk_adjustment_factor" yaml:"risk_adjustment_factor"`
	EnableRealTimeAlerts      bool            `json:"enable_real_time_alerts" yaml:"enable_real_time_alerts"`
	EnablePositionSizing      bool            `json:"enable_position_sizing" yaml:"enable_position_sizing"`
	MonitoredAssets           []string        `json:"monitored_assets" yaml:"monitored_assets"`
}

// AssetVolatility represents volatility metrics for a specific asset
type AssetVolatility struct {
	Asset                string          `json:"asset"`
	CurrentVolatility    decimal.Decimal `json:"current_volatility"`
	AverageVolatility    decimal.Decimal `json:"average_volatility"`
	VolatilityPercentile decimal.Decimal `json:"volatility_percentile"`
	VolatilityTrend      string          `json:"volatility_trend"` // increasing, decreasing, stable
	TrendStrength        decimal.Decimal `json:"trend_strength"`
	LastUpdate           time.Time       `json:"last_update"`
	HistoricalHigh       decimal.Decimal `json:"historical_high"`
	HistoricalLow        decimal.Decimal `json:"historical_low"`
	VolatilityRegime     string          `json:"volatility_regime"` // low, normal, high, extreme
	PriceMovements       []PriceMovement `json:"price_movements"`
}

// PriceMovement represents a price movement data point
type PriceMovement struct {
	Timestamp     time.Time       `json:"timestamp"`
	Price         decimal.Decimal `json:"price"`
	Return        decimal.Decimal `json:"return"`
	AbsReturn     decimal.Decimal `json:"abs_return"`
	SquaredReturn decimal.Decimal `json:"squared_return"`
}

// MarketRegime represents the current market regime
type MarketRegime struct {
	Regime          string          `json:"regime"` // bull, bear, sideways, volatile
	Confidence      decimal.Decimal `json:"confidence"`
	Duration        time.Duration   `json:"duration"`
	StartTime       time.Time       `json:"start_time"`
	Characteristics []string        `json:"characteristics"`
	VolatilityLevel string          `json:"volatility_level"`
	TrendDirection  string          `json:"trend_direction"`
	MarketStress    decimal.Decimal `json:"market_stress"`
	LiquidityLevel  decimal.Decimal `json:"liquidity_level"`
}

// PositionSizeRecommendation provides position sizing recommendations
type PositionSizeRecommendation struct {
	Asset                 string          `json:"asset"`
	RecommendedSize       decimal.Decimal `json:"recommended_size"`
	MaxSafeSize           decimal.Decimal `json:"max_safe_size"`
	MinEffectiveSize      decimal.Decimal `json:"min_effective_size"`
	RiskAdjustedSize      decimal.Decimal `json:"risk_adjusted_size"`
	VolatilityAdjustment  decimal.Decimal `json:"volatility_adjustment"`
	CorrelationAdjustment decimal.Decimal `json:"correlation_adjustment"`
	Reasoning             []string        `json:"reasoning"`
	Confidence            decimal.Decimal `json:"confidence"`
	LastUpdated           time.Time       `json:"last_updated"`
	ValidUntil            time.Time       `json:"valid_until"`
}

// VolatilityAlert represents a volatility-based alert
type VolatilityAlert struct {
	ID              string          `json:"id"`
	Asset           string          `json:"asset"`
	AlertType       string          `json:"alert_type"` // high_volatility, low_volatility, regime_change, correlation_break
	Severity        string          `json:"severity"`   // low, medium, high, critical
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	CurrentValue    decimal.Decimal `json:"current_value"`
	ThresholdValue  decimal.Decimal `json:"threshold_value"`
	Recommendations []string        `json:"recommendations"`
	CreatedAt       time.Time       `json:"created_at"`
	ResolvedAt      *time.Time      `json:"resolved_at,omitempty"`
	Status          string          `json:"status"` // active, resolved, dismissed
}

// HistoricalVolatilityData stores historical volatility data for analysis
type HistoricalVolatilityData struct {
	Asset               string                `json:"asset"`
	DailyVolatilities   []VolatilityDataPoint `json:"daily_volatilities"`
	WeeklyVolatilities  []VolatilityDataPoint `json:"weekly_volatilities"`
	MonthlyVolatilities []VolatilityDataPoint `json:"monthly_volatilities"`
	VolatilityStats     *VolatilityStatistics `json:"volatility_stats"`
	LastUpdated         time.Time             `json:"last_updated"`
}

// VolatilityDataPoint represents a single volatility measurement
type VolatilityDataPoint struct {
	Timestamp          time.Time       `json:"timestamp"`
	Volatility         decimal.Decimal `json:"volatility"`
	RealizedVolatility decimal.Decimal `json:"realized_volatility"`
	ImpliedVolatility  decimal.Decimal `json:"implied_volatility,omitempty"`
	VolumeWeighted     decimal.Decimal `json:"volume_weighted"`
}

// VolatilityStatistics provides statistical analysis of volatility
type VolatilityStatistics struct {
	Mean              decimal.Decimal `json:"mean"`
	Median            decimal.Decimal `json:"median"`
	StandardDeviation decimal.Decimal `json:"standard_deviation"`
	Skewness          decimal.Decimal `json:"skewness"`
	Kurtosis          decimal.Decimal `json:"kurtosis"`
	Percentile25      decimal.Decimal `json:"percentile_25"`
	Percentile75      decimal.Decimal `json:"percentile_75"`
	Percentile95      decimal.Decimal `json:"percentile_95"`
	Percentile99      decimal.Decimal `json:"percentile_99"`
	AutoCorrelation   decimal.Decimal `json:"auto_correlation"`
}

// RealTimeVolatilityMetrics provides real-time market volatility metrics
type RealTimeVolatilityMetrics struct {
	OverallMarketVolatility decimal.Decimal            `json:"overall_market_volatility"`
	VolatilityIndex         decimal.Decimal            `json:"volatility_index"`
	MarketStressIndex       decimal.Decimal            `json:"market_stress_index"`
	CorrelationBreakdowns   []CorrelationBreakdown     `json:"correlation_breakdowns"`
	RegimeChanges           []RegimeChange             `json:"regime_changes"`
	VolatilitySpikes        []VolatilitySpike          `json:"volatility_spikes"`
	CrossAssetVolatility    map[string]decimal.Decimal `json:"cross_asset_volatility"`
	LastUpdated             time.Time                  `json:"last_updated"`
}

// CorrelationBreakdown represents a breakdown in asset correlation
type CorrelationBreakdown struct {
	Asset1             string          `json:"asset1"`
	Asset2             string          `json:"asset2"`
	PreviousCorr       decimal.Decimal `json:"previous_correlation"`
	CurrentCorr        decimal.Decimal `json:"current_correlation"`
	BreakdownMagnitude decimal.Decimal `json:"breakdown_magnitude"`
	DetectedAt         time.Time       `json:"detected_at"`
	Significance       string          `json:"significance"`
}

// RegimeChange represents a detected market regime change
type RegimeChange struct {
	FromRegime   string          `json:"from_regime"`
	ToRegime     string          `json:"to_regime"`
	Confidence   decimal.Decimal `json:"confidence"`
	DetectedAt   time.Time       `json:"detected_at"`
	TriggerEvent string          `json:"trigger_event"`
	Impact       string          `json:"impact"`
}

// VolatilitySpike represents a detected volatility spike
type VolatilitySpike struct {
	Asset          string          `json:"asset"`
	SpikeIntensity decimal.Decimal `json:"spike_intensity"`
	Duration       time.Duration   `json:"duration"`
	StartTime      time.Time       `json:"start_time"`
	PeakTime       time.Time       `json:"peak_time"`
	TriggerFactor  string          `json:"trigger_factor"`
	MarketImpact   decimal.Decimal `json:"market_impact"`
}

// NewMarketVolatilityEngine creates a new market volatility engine
func NewMarketVolatilityEngine(logger *logger.Logger, cache redis.Client, config VolatilityConfig) *MarketVolatilityEngine {
	return &MarketVolatilityEngine{
		logger:            logger.Named("market-volatility-engine"),
		cache:             cache,
		config:            config,
		assetVolatilities: make(map[string]*AssetVolatility),
		correlationMatrix: make(map[string]map[string]decimal.Decimal),
		marketRegimes:     make(map[string]*MarketRegime),
		positionSizes:     make(map[string]*PositionSizeRecommendation),
		volatilityAlerts:  make(map[string]*VolatilityAlert),
		historicalData:    make(map[string]*HistoricalVolatilityData),
		realTimeMetrics:   &RealTimeVolatilityMetrics{},
		stopChan:          make(chan struct{}),
	}
}

// Start starts the market volatility engine
func (mve *MarketVolatilityEngine) Start(ctx context.Context) error {
	mve.mutex.Lock()
	defer mve.mutex.Unlock()

	if mve.isRunning {
		return fmt.Errorf("market volatility engine is already running")
	}

	if !mve.config.Enabled {
		mve.logger.Info("Market volatility analysis is disabled")
		return nil
	}

	mve.logger.Info("Starting market volatility engine",
		zap.Duration("analysis_interval", mve.config.AnalysisInterval),
		zap.Duration("volatility_window", mve.config.VolatilityWindow),
		zap.Strings("monitored_assets", mve.config.MonitoredAssets))

	// Initialize historical data
	if err := mve.initializeHistoricalData(ctx); err != nil {
		return fmt.Errorf("failed to initialize historical data: %w", err)
	}

	mve.isRunning = true

	// Start analysis loops
	go mve.volatilityAnalysisLoop(ctx)
	go mve.correlationAnalysisLoop(ctx)
	go mve.regimeDetectionLoop(ctx)
	go mve.positionSizingLoop(ctx)
	go mve.alertManagementLoop(ctx)

	mve.logger.Info("Market volatility engine started successfully")
	return nil
}

// Stop stops the market volatility engine
func (mve *MarketVolatilityEngine) Stop() error {
	mve.mutex.Lock()
	defer mve.mutex.Unlock()

	if !mve.isRunning {
		return nil
	}

	mve.logger.Info("Stopping market volatility engine")
	mve.isRunning = false
	close(mve.stopChan)

	mve.logger.Info("Market volatility engine stopped")
	return nil
}

// AnalyzeAssetVolatility analyzes volatility for a specific asset
func (mve *MarketVolatilityEngine) AnalyzeAssetVolatility(ctx context.Context, asset string, priceData []PriceMovement) (*AssetVolatility, error) {
	mve.logger.Debug("Analyzing asset volatility", zap.String("asset", asset))

	if len(priceData) < 2 {
		return nil, fmt.Errorf("insufficient price data for volatility analysis")
	}

	// Calculate returns
	returns := mve.calculateReturns(priceData)

	// Calculate current volatility (annualized)
	currentVolatility := mve.calculateVolatility(returns)

	// Get historical volatility for comparison
	historicalVol := mve.getHistoricalVolatility(asset)

	// Calculate volatility percentile
	percentile := mve.calculateVolatilityPercentile(asset, currentVolatility)

	// Determine volatility trend
	trend, strength := mve.analyzeVolatilityTrend(asset, currentVolatility)

	// Determine volatility regime
	regime := mve.determineVolatilityRegime(currentVolatility, historicalVol)

	volatility := &AssetVolatility{
		Asset:                asset,
		CurrentVolatility:    currentVolatility,
		AverageVolatility:    historicalVol,
		VolatilityPercentile: percentile,
		VolatilityTrend:      trend,
		TrendStrength:        strength,
		LastUpdate:           time.Now(),
		HistoricalHigh:       mve.getHistoricalHigh(asset),
		HistoricalLow:        mve.getHistoricalLow(asset),
		VolatilityRegime:     regime,
		PriceMovements:       priceData,
	}

	// Store volatility data
	mve.mutex.Lock()
	mve.assetVolatilities[asset] = volatility
	mve.mutex.Unlock()

	// Check for alerts
	mve.checkVolatilityAlerts(volatility)

	mve.logger.Info("Asset volatility analyzed",
		zap.String("asset", asset),
		zap.String("current_volatility", currentVolatility.String()),
		zap.String("regime", regime),
		zap.String("trend", trend))

	return volatility, nil
}

// CalculateCorrelationMatrix calculates correlation matrix for monitored assets
func (mve *MarketVolatilityEngine) CalculateCorrelationMatrix(ctx context.Context) (map[string]map[string]decimal.Decimal, error) {
	mve.logger.Debug("Calculating correlation matrix")

	assets := mve.config.MonitoredAssets
	correlations := make(map[string]map[string]decimal.Decimal)

	for _, asset1 := range assets {
		correlations[asset1] = make(map[string]decimal.Decimal)

		for _, asset2 := range assets {
			if asset1 == asset2 {
				correlations[asset1][asset2] = decimal.NewFromFloat(1.0)
			} else {
				corr, err := mve.calculatePairwiseCorrelation(asset1, asset2)
				if err != nil {
					mve.logger.Warn("Failed to calculate correlation",
						zap.String("asset1", asset1),
						zap.String("asset2", asset2),
						zap.Error(err))
					correlations[asset1][asset2] = decimal.Zero
				} else {
					correlations[asset1][asset2] = corr
				}
			}
		}
	}

	// Store correlation matrix
	mve.mutex.Lock()
	mve.correlationMatrix = correlations
	mve.mutex.Unlock()

	// Check for correlation breakdowns
	mve.detectCorrelationBreakdowns(correlations)

	mve.logger.Info("Correlation matrix calculated",
		zap.Int("asset_count", len(assets)))

	return correlations, nil
}

// GeneratePositionSizeRecommendation generates position sizing recommendations
func (mve *MarketVolatilityEngine) GeneratePositionSizeRecommendation(ctx context.Context, asset string, portfolioValue decimal.Decimal) (*PositionSizeRecommendation, error) {
	mve.logger.Debug("Generating position size recommendation",
		zap.String("asset", asset),
		zap.String("portfolio_value", portfolioValue.String()))

	// Get asset volatility
	mve.mutex.RLock()
	volatility, exists := mve.assetVolatilities[asset]
	mve.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("volatility data not available for asset: %s", asset)
	}

	// Calculate base position size using Kelly criterion
	baseSize := mve.calculateKellyPositionSize(asset, portfolioValue)

	// Apply volatility adjustment
	volAdjustment := mve.calculateVolatilityAdjustment(volatility)
	volatilityAdjustedSize := baseSize.Mul(volAdjustment)

	// Apply correlation adjustment
	corrAdjustment := mve.calculateCorrelationAdjustment(asset)
	finalSize := volatilityAdjustedSize.Mul(corrAdjustment)

	// Apply risk limits
	maxSafe := portfolioValue.Mul(mve.config.MaxPositionSize)
	minEffective := portfolioValue.Mul(mve.config.MinPositionSize)

	if finalSize.GreaterThan(maxSafe) {
		finalSize = maxSafe
	}
	if finalSize.LessThan(minEffective) {
		finalSize = minEffective
	}

	// Generate reasoning
	reasoning := mve.generatePositionSizeReasoning(volatility, volAdjustment, corrAdjustment)

	// Calculate confidence based on data quality
	confidence := mve.calculatePositionSizeConfidence(volatility)

	recommendation := &PositionSizeRecommendation{
		Asset:                 asset,
		RecommendedSize:       finalSize,
		MaxSafeSize:           maxSafe,
		MinEffectiveSize:      minEffective,
		RiskAdjustedSize:      finalSize,
		VolatilityAdjustment:  volAdjustment,
		CorrelationAdjustment: corrAdjustment,
		Reasoning:             reasoning,
		Confidence:            confidence,
		LastUpdated:           time.Now(),
		ValidUntil:            time.Now().Add(1 * time.Hour),
	}

	// Store recommendation
	mve.mutex.Lock()
	mve.positionSizes[asset] = recommendation
	mve.mutex.Unlock()

	mve.logger.Info("Position size recommendation generated",
		zap.String("asset", asset),
		zap.String("recommended_size", finalSize.String()),
		zap.String("confidence", confidence.String()))

	return recommendation, nil
}

// Core calculation methods

// calculateReturns calculates returns from price movements
func (mve *MarketVolatilityEngine) calculateReturns(priceData []PriceMovement) []decimal.Decimal {
	if len(priceData) < 2 {
		return []decimal.Decimal{}
	}

	returns := make([]decimal.Decimal, len(priceData)-1)
	for i := 1; i < len(priceData); i++ {
		if priceData[i-1].Price.IsZero() {
			returns[i-1] = decimal.Zero
			continue
		}

		// Calculate log return: ln(P_t / P_{t-1})
		ratio := priceData[i].Price.Div(priceData[i-1].Price)
		if ratio.LessThanOrEqual(decimal.Zero) {
			returns[i-1] = decimal.Zero
			continue
		}

		// Approximate log return for small changes
		logReturn := ratio.Sub(decimal.NewFromFloat(1.0))
		returns[i-1] = logReturn
	}

	return returns
}

// calculateVolatility calculates annualized volatility from returns
func (mve *MarketVolatilityEngine) calculateVolatility(returns []decimal.Decimal) decimal.Decimal {
	if len(returns) < 2 {
		return decimal.Zero
	}

	// Calculate mean return
	sum := decimal.Zero
	for _, ret := range returns {
		sum = sum.Add(ret)
	}
	mean := sum.Div(decimal.NewFromInt(int64(len(returns))))

	// Calculate variance
	variance := decimal.Zero
	for _, ret := range returns {
		diff := ret.Sub(mean)
		variance = variance.Add(diff.Mul(diff))
	}
	variance = variance.Div(decimal.NewFromInt(int64(len(returns) - 1)))

	// Calculate standard deviation
	stdDev := mve.sqrt(variance)

	// Annualize (assuming daily returns, multiply by sqrt(365))
	annualizedVol := stdDev.Mul(decimal.NewFromFloat(math.Sqrt(365)))

	return annualizedVol
}

// sqrt calculates square root using Newton's method
func (mve *MarketVolatilityEngine) sqrt(x decimal.Decimal) decimal.Decimal {
	if x.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero
	}

	// Convert to float for sqrt calculation
	f, _ := x.Float64()
	result := math.Sqrt(f)
	return decimal.NewFromFloat(result)
}

// calculatePairwiseCorrelation calculates correlation between two assets
func (mve *MarketVolatilityEngine) calculatePairwiseCorrelation(asset1, asset2 string) (decimal.Decimal, error) {
	// Get price data for both assets
	data1 := mve.getPriceData(asset1)
	data2 := mve.getPriceData(asset2)

	if len(data1) < 2 || len(data2) < 2 {
		return decimal.Zero, fmt.Errorf("insufficient data for correlation calculation")
	}

	// Calculate returns
	returns1 := mve.calculateReturns(data1)
	returns2 := mve.calculateReturns(data2)

	// Align data (use minimum length)
	minLen := len(returns1)
	if len(returns2) < minLen {
		minLen = len(returns2)
	}

	if minLen < 2 {
		return decimal.Zero, fmt.Errorf("insufficient aligned data for correlation")
	}

	returns1 = returns1[:minLen]
	returns2 = returns2[:minLen]

	// Calculate correlation coefficient
	return mve.calculateCorrelationCoefficient(returns1, returns2), nil
}

// calculateCorrelationCoefficient calculates Pearson correlation coefficient
func (mve *MarketVolatilityEngine) calculateCorrelationCoefficient(x, y []decimal.Decimal) decimal.Decimal {
	if len(x) != len(y) || len(x) < 2 {
		return decimal.Zero
	}

	n := decimal.NewFromInt(int64(len(x)))

	// Calculate means
	sumX, sumY := decimal.Zero, decimal.Zero
	for i := 0; i < len(x); i++ {
		sumX = sumX.Add(x[i])
		sumY = sumY.Add(y[i])
	}
	meanX := sumX.Div(n)
	meanY := sumY.Div(n)

	// Calculate correlation components
	numerator := decimal.Zero
	sumXX := decimal.Zero
	sumYY := decimal.Zero

	for i := 0; i < len(x); i++ {
		diffX := x[i].Sub(meanX)
		diffY := y[i].Sub(meanY)

		numerator = numerator.Add(diffX.Mul(diffY))
		sumXX = sumXX.Add(diffX.Mul(diffX))
		sumYY = sumYY.Add(diffY.Mul(diffY))
	}

	denominator := mve.sqrt(sumXX.Mul(sumYY))
	if denominator.IsZero() {
		return decimal.Zero
	}

	return numerator.Div(denominator)
}

// Helper methods

// getHistoricalVolatility gets historical average volatility for an asset
func (mve *MarketVolatilityEngine) getHistoricalVolatility(asset string) decimal.Decimal {
	mve.mutex.RLock()
	defer mve.mutex.RUnlock()

	if data, exists := mve.historicalData[asset]; exists && data.VolatilityStats != nil {
		return data.VolatilityStats.Mean
	}

	// Default volatility if no historical data
	return decimal.NewFromFloat(0.3) // 30% default
}

// calculateVolatilityPercentile calculates where current volatility ranks historically
func (mve *MarketVolatilityEngine) calculateVolatilityPercentile(asset string, currentVol decimal.Decimal) decimal.Decimal {
	mve.mutex.RLock()
	defer mve.mutex.RUnlock()

	data, exists := mve.historicalData[asset]
	if !exists || len(data.DailyVolatilities) < 10 {
		return decimal.NewFromFloat(0.5) // 50th percentile default
	}

	// Count how many historical values are below current
	below := 0
	total := len(data.DailyVolatilities)

	for _, point := range data.DailyVolatilities {
		if point.Volatility.LessThan(currentVol) {
			below++
		}
	}

	percentile := decimal.NewFromInt(int64(below)).Div(decimal.NewFromInt(int64(total)))
	return percentile
}

// analyzeVolatilityTrend analyzes the trend in volatility
func (mve *MarketVolatilityEngine) analyzeVolatilityTrend(asset string, currentVol decimal.Decimal) (string, decimal.Decimal) {
	mve.mutex.RLock()
	defer mve.mutex.RUnlock()

	data, exists := mve.historicalData[asset]
	if !exists || len(data.DailyVolatilities) < 5 {
		return "stable", decimal.NewFromFloat(0.5)
	}

	// Get recent volatilities (last 5 days)
	recent := data.DailyVolatilities
	if len(recent) > 5 {
		recent = recent[len(recent)-5:]
	}

	// Calculate trend
	increasing := 0
	decreasing := 0

	for i := 1; i < len(recent); i++ {
		if recent[i].Volatility.GreaterThan(recent[i-1].Volatility) {
			increasing++
		} else if recent[i].Volatility.LessThan(recent[i-1].Volatility) {
			decreasing++
		}
	}

	if increasing > decreasing {
		strength := decimal.NewFromInt(int64(increasing)).Div(decimal.NewFromInt(int64(len(recent) - 1)))
		return "increasing", strength
	} else if decreasing > increasing {
		strength := decimal.NewFromInt(int64(decreasing)).Div(decimal.NewFromInt(int64(len(recent) - 1)))
		return "decreasing", strength
	}

	return "stable", decimal.NewFromFloat(0.5)
}

// determineVolatilityRegime determines the current volatility regime
func (mve *MarketVolatilityEngine) determineVolatilityRegime(currentVol, historicalVol decimal.Decimal) string {
	ratio := currentVol.Div(historicalVol)

	if ratio.LessThan(decimal.NewFromFloat(0.5)) {
		return "low"
	} else if ratio.LessThan(decimal.NewFromFloat(1.5)) {
		return "normal"
	} else if ratio.LessThan(decimal.NewFromFloat(2.5)) {
		return "high"
	}
	return "extreme"
}

// getHistoricalHigh gets historical high volatility
func (mve *MarketVolatilityEngine) getHistoricalHigh(asset string) decimal.Decimal {
	mve.mutex.RLock()
	defer mve.mutex.RUnlock()

	if data, exists := mve.historicalData[asset]; exists && data.VolatilityStats != nil {
		return data.VolatilityStats.Percentile99
	}
	return decimal.NewFromFloat(1.0) // 100% default high
}

// getHistoricalLow gets historical low volatility
func (mve *MarketVolatilityEngine) getHistoricalLow(asset string) decimal.Decimal {
	mve.mutex.RLock()
	defer mve.mutex.RUnlock()

	if data, exists := mve.historicalData[asset]; exists && data.VolatilityStats != nil {
		return data.VolatilityStats.Percentile25
	}
	return decimal.NewFromFloat(0.05) // 5% default low
}

// getPriceData gets price data for an asset
func (mve *MarketVolatilityEngine) getPriceData(asset string) []PriceMovement {
	mve.mutex.RLock()
	defer mve.mutex.RUnlock()

	if volatility, exists := mve.assetVolatilities[asset]; exists {
		return volatility.PriceMovements
	}
	return []PriceMovement{}
}

// Position sizing methods

// calculateKellyPositionSize calculates position size using Kelly criterion
func (mve *MarketVolatilityEngine) calculateKellyPositionSize(asset string, portfolioValue decimal.Decimal) decimal.Decimal {
	// Simplified Kelly criterion: f = (bp - q) / b
	// where f = fraction to bet, b = odds, p = probability of win, q = probability of loss

	// For crypto, we'll use a conservative approach
	// Assume 55% win probability and 1:1 odds for simplicity
	winProb := decimal.NewFromFloat(0.55)
	lossProb := decimal.NewFromFloat(0.45)
	odds := decimal.NewFromFloat(1.0)

	// Kelly fraction
	kellyFraction := winProb.Sub(lossProb.Div(odds))

	// Apply conservative cap (max 25% of portfolio)
	maxKelly := decimal.NewFromFloat(0.25)
	if kellyFraction.GreaterThan(maxKelly) {
		kellyFraction = maxKelly
	}

	// Ensure minimum position
	minKelly := decimal.NewFromFloat(0.01) // 1%
	if kellyFraction.LessThan(minKelly) {
		kellyFraction = minKelly
	}

	return portfolioValue.Mul(kellyFraction)
}

// calculateVolatilityAdjustment calculates volatility-based position adjustment
func (mve *MarketVolatilityEngine) calculateVolatilityAdjustment(volatility *AssetVolatility) decimal.Decimal {
	// Higher volatility = smaller position
	// Use inverse relationship with volatility

	baseAdjustment := decimal.NewFromFloat(1.0)

	switch volatility.VolatilityRegime {
	case "low":
		return baseAdjustment.Mul(decimal.NewFromFloat(1.2)) // 20% larger position
	case "normal":
		return baseAdjustment
	case "high":
		return baseAdjustment.Mul(decimal.NewFromFloat(0.7)) // 30% smaller position
	case "extreme":
		return baseAdjustment.Mul(decimal.NewFromFloat(0.4)) // 60% smaller position
	default:
		return baseAdjustment
	}
}

// calculateCorrelationAdjustment calculates correlation-based position adjustment
func (mve *MarketVolatilityEngine) calculateCorrelationAdjustment(asset string) decimal.Decimal {
	mve.mutex.RLock()
	defer mve.mutex.RUnlock()

	// Calculate average correlation with other assets
	correlations, exists := mve.correlationMatrix[asset]
	if !exists || len(correlations) <= 1 {
		return decimal.NewFromFloat(1.0) // No adjustment if no correlation data
	}

	totalCorr := decimal.Zero
	count := 0

	for otherAsset, corr := range correlations {
		if otherAsset != asset {
			totalCorr = totalCorr.Add(corr.Abs()) // Use absolute correlation
			count++
		}
	}

	if count == 0 {
		return decimal.NewFromFloat(1.0)
	}

	avgCorr := totalCorr.Div(decimal.NewFromInt(int64(count)))

	// Higher correlation = smaller position (less diversification benefit)
	// Adjustment factor: 1 - (avgCorr * 0.5)
	adjustment := decimal.NewFromFloat(1.0).Sub(avgCorr.Mul(decimal.NewFromFloat(0.5)))

	// Ensure adjustment is between 0.3 and 1.2
	if adjustment.LessThan(decimal.NewFromFloat(0.3)) {
		adjustment = decimal.NewFromFloat(0.3)
	}
	if adjustment.GreaterThan(decimal.NewFromFloat(1.2)) {
		adjustment = decimal.NewFromFloat(1.2)
	}

	return adjustment
}

// generatePositionSizeReasoning generates reasoning for position size recommendation
func (mve *MarketVolatilityEngine) generatePositionSizeReasoning(volatility *AssetVolatility, volAdjustment, corrAdjustment decimal.Decimal) []string {
	var reasoning []string

	// Volatility reasoning
	switch volatility.VolatilityRegime {
	case "low":
		reasoning = append(reasoning, "Low volatility environment allows for larger position size")
	case "normal":
		reasoning = append(reasoning, "Normal volatility conditions support standard position sizing")
	case "high":
		reasoning = append(reasoning, "High volatility requires reduced position size for risk management")
	case "extreme":
		reasoning = append(reasoning, "Extreme volatility necessitates significantly reduced position size")
	}

	// Trend reasoning
	if volatility.VolatilityTrend == "increasing" {
		reasoning = append(reasoning, "Increasing volatility trend suggests caution in position sizing")
	} else if volatility.VolatilityTrend == "decreasing" {
		reasoning = append(reasoning, "Decreasing volatility trend supports more confident position sizing")
	}

	// Correlation reasoning
	if corrAdjustment.LessThan(decimal.NewFromFloat(0.8)) {
		reasoning = append(reasoning, "High correlation with other assets reduces diversification benefit")
	} else if corrAdjustment.GreaterThan(decimal.NewFromFloat(1.1)) {
		reasoning = append(reasoning, "Low correlation with other assets provides diversification benefit")
	}

	// Percentile reasoning
	if volatility.VolatilityPercentile.GreaterThan(decimal.NewFromFloat(0.8)) {
		reasoning = append(reasoning, "Current volatility is in the top 20% historically")
	} else if volatility.VolatilityPercentile.LessThan(decimal.NewFromFloat(0.2)) {
		reasoning = append(reasoning, "Current volatility is in the bottom 20% historically")
	}

	return reasoning
}

// calculatePositionSizeConfidence calculates confidence in position size recommendation
func (mve *MarketVolatilityEngine) calculatePositionSizeConfidence(volatility *AssetVolatility) decimal.Decimal {
	confidence := decimal.NewFromFloat(0.5) // Base confidence

	// Data quality factors
	if len(volatility.PriceMovements) >= 30 {
		confidence = confidence.Add(decimal.NewFromFloat(0.2)) // Good data sample
	}

	// Volatility stability
	if volatility.VolatilityTrend == "stable" {
		confidence = confidence.Add(decimal.NewFromFloat(0.1))
	}

	// Regime clarity
	if volatility.VolatilityRegime == "normal" {
		confidence = confidence.Add(decimal.NewFromFloat(0.1))
	} else if volatility.VolatilityRegime == "extreme" {
		confidence = confidence.Sub(decimal.NewFromFloat(0.1)) // Less confident in extreme conditions
	}

	// Recent update
	if time.Since(volatility.LastUpdate) < 1*time.Hour {
		confidence = confidence.Add(decimal.NewFromFloat(0.1))
	}

	// Cap confidence between 0.1 and 1.0
	if confidence.LessThan(decimal.NewFromFloat(0.1)) {
		confidence = decimal.NewFromFloat(0.1)
	}
	if confidence.GreaterThan(decimal.NewFromFloat(1.0)) {
		confidence = decimal.NewFromFloat(1.0)
	}

	return confidence
}

// Alert and monitoring methods

// checkVolatilityAlerts checks for volatility-based alerts
func (mve *MarketVolatilityEngine) checkVolatilityAlerts(volatility *AssetVolatility) {
	// High volatility alert
	if volatility.CurrentVolatility.GreaterThan(mve.config.HighVolatilityThreshold) {
		alert := &VolatilityAlert{
			ID:              mve.generateAlertID(),
			Asset:           volatility.Asset,
			AlertType:       "high_volatility",
			Severity:        "high",
			Title:           fmt.Sprintf("High Volatility Alert: %s", volatility.Asset),
			Description:     fmt.Sprintf("Asset %s volatility (%.2f%%) exceeds threshold (%.2f%%)", volatility.Asset, volatility.CurrentVolatility.Mul(decimal.NewFromFloat(100)).InexactFloat64(), mve.config.HighVolatilityThreshold.Mul(decimal.NewFromFloat(100)).InexactFloat64()),
			CurrentValue:    volatility.CurrentVolatility,
			ThresholdValue:  mve.config.HighVolatilityThreshold,
			Recommendations: []string{"Consider reducing position size", "Implement tighter stop losses", "Monitor for regime change"},
			CreatedAt:       time.Now(),
			Status:          "active",
		}

		mve.mutex.Lock()
		mve.volatilityAlerts[alert.ID] = alert
		mve.mutex.Unlock()

		mve.logger.Warn("High volatility alert created",
			zap.String("alert_id", alert.ID),
			zap.String("asset", volatility.Asset))
	}

	// Low volatility alert
	if volatility.CurrentVolatility.LessThan(mve.config.LowVolatilityThreshold) {
		alert := &VolatilityAlert{
			ID:              mve.generateAlertID(),
			Asset:           volatility.Asset,
			AlertType:       "low_volatility",
			Severity:        "medium",
			Title:           fmt.Sprintf("Low Volatility Alert: %s", volatility.Asset),
			Description:     fmt.Sprintf("Asset %s volatility (%.2f%%) below threshold (%.2f%%)", volatility.Asset, volatility.CurrentVolatility.Mul(decimal.NewFromFloat(100)).InexactFloat64(), mve.config.LowVolatilityThreshold.Mul(decimal.NewFromFloat(100)).InexactFloat64()),
			CurrentValue:    volatility.CurrentVolatility,
			ThresholdValue:  mve.config.LowVolatilityThreshold,
			Recommendations: []string{"Consider increasing position size", "Look for breakout opportunities", "Monitor for regime change"},
			CreatedAt:       time.Now(),
			Status:          "active",
		}

		mve.mutex.Lock()
		mve.volatilityAlerts[alert.ID] = alert
		mve.mutex.Unlock()

		mve.logger.Info("Low volatility alert created",
			zap.String("alert_id", alert.ID),
			zap.String("asset", volatility.Asset))
	}
}

// detectCorrelationBreakdowns detects significant changes in asset correlations
func (mve *MarketVolatilityEngine) detectCorrelationBreakdowns(currentCorrelations map[string]map[string]decimal.Decimal) {
	// This would compare with previous correlations to detect breakdowns
	// For now, we'll implement a simplified version

	for asset1, correlations := range currentCorrelations {
		for asset2, corr := range correlations {
			if asset1 >= asset2 { // Avoid duplicate pairs
				continue
			}

			// Check if correlation is unusually high or low
			if corr.Abs().GreaterThan(mve.config.CorrelationThreshold) {
				breakdown := &CorrelationBreakdown{
					Asset1:             asset1,
					Asset2:             asset2,
					PreviousCorr:       decimal.NewFromFloat(0.5), // Mock previous correlation
					CurrentCorr:        corr,
					BreakdownMagnitude: corr.Abs().Sub(decimal.NewFromFloat(0.5)).Abs(),
					DetectedAt:         time.Now(),
					Significance:       "high",
				}

				// Add to real-time metrics
				mve.mutex.Lock()
				mve.realTimeMetrics.CorrelationBreakdowns = append(mve.realTimeMetrics.CorrelationBreakdowns, *breakdown)
				mve.mutex.Unlock()

				mve.logger.Info("Correlation breakdown detected",
					zap.String("asset1", asset1),
					zap.String("asset2", asset2),
					zap.String("correlation", corr.String()))
			}
		}
	}
}

// generateAlertID generates a unique alert ID
func (mve *MarketVolatilityEngine) generateAlertID() string {
	return fmt.Sprintf("vol_alert_%d", time.Now().UnixNano())
}

// Loop methods

// initializeHistoricalData initializes historical volatility data
func (mve *MarketVolatilityEngine) initializeHistoricalData(ctx context.Context) error {
	mve.logger.Info("Initializing historical volatility data")

	for _, asset := range mve.config.MonitoredAssets {
		// Initialize with mock historical data
		historicalData := &HistoricalVolatilityData{
			Asset:               asset,
			DailyVolatilities:   mve.generateMockVolatilityData(30), // 30 days
			WeeklyVolatilities:  mve.generateMockVolatilityData(12), // 12 weeks
			MonthlyVolatilities: mve.generateMockVolatilityData(6),  // 6 months
			VolatilityStats:     mve.calculateVolatilityStats(asset),
			LastUpdated:         time.Now(),
		}

		mve.mutex.Lock()
		mve.historicalData[asset] = historicalData
		mve.mutex.Unlock()
	}

	mve.logger.Info("Historical volatility data initialized",
		zap.Int("asset_count", len(mve.config.MonitoredAssets)))

	return nil
}

// generateMockVolatilityData generates mock volatility data for testing
func (mve *MarketVolatilityEngine) generateMockVolatilityData(count int) []VolatilityDataPoint {
	data := make([]VolatilityDataPoint, count)
	baseVol := decimal.NewFromFloat(0.3) // 30% base volatility

	for i := 0; i < count; i++ {
		// Add some randomness to volatility
		variation := decimal.NewFromFloat(float64(i%10-5) * 0.02) // ±10% variation
		volatility := baseVol.Add(variation)

		if volatility.LessThan(decimal.NewFromFloat(0.05)) {
			volatility = decimal.NewFromFloat(0.05) // Minimum 5%
		}

		data[i] = VolatilityDataPoint{
			Timestamp:          time.Now().Add(-time.Duration(count-i) * 24 * time.Hour),
			Volatility:         volatility,
			RealizedVolatility: volatility.Mul(decimal.NewFromFloat(0.9)),
			VolumeWeighted:     volatility.Mul(decimal.NewFromFloat(1.1)),
		}
	}

	return data
}

// calculateVolatilityStats calculates statistical measures for volatility
func (mve *MarketVolatilityEngine) calculateVolatilityStats(asset string) *VolatilityStatistics {
	// Mock statistics - in production would calculate from real data
	return &VolatilityStatistics{
		Mean:              decimal.NewFromFloat(0.3),
		Median:            decimal.NewFromFloat(0.28),
		StandardDeviation: decimal.NewFromFloat(0.15),
		Skewness:          decimal.NewFromFloat(0.5),
		Kurtosis:          decimal.NewFromFloat(3.2),
		Percentile25:      decimal.NewFromFloat(0.2),
		Percentile75:      decimal.NewFromFloat(0.4),
		Percentile95:      decimal.NewFromFloat(0.6),
		Percentile99:      decimal.NewFromFloat(0.8),
		AutoCorrelation:   decimal.NewFromFloat(0.1),
	}
}

// volatilityAnalysisLoop performs periodic volatility analysis
func (mve *MarketVolatilityEngine) volatilityAnalysisLoop(ctx context.Context) {
	ticker := time.NewTicker(mve.config.AnalysisInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-mve.stopChan:
			return
		case <-ticker.C:
			mve.performVolatilityAnalysis(ctx)
		}
	}
}

// correlationAnalysisLoop performs periodic correlation analysis
func (mve *MarketVolatilityEngine) correlationAnalysisLoop(ctx context.Context) {
	ticker := time.NewTicker(mve.config.CorrelationUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-mve.stopChan:
			return
		case <-ticker.C:
			_, err := mve.CalculateCorrelationMatrix(ctx)
			if err != nil {
				mve.logger.Error("Failed to update correlation matrix", zap.Error(err))
			}
		}
	}
}

// regimeDetectionLoop performs periodic market regime detection
func (mve *MarketVolatilityEngine) regimeDetectionLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-mve.stopChan:
			return
		case <-ticker.C:
			mve.detectMarketRegimeChanges(ctx)
		}
	}
}

// positionSizingLoop performs periodic position sizing updates
func (mve *MarketVolatilityEngine) positionSizingLoop(ctx context.Context) {
	if !mve.config.EnablePositionSizing {
		return
	}

	ticker := time.NewTicker(15 * time.Minute) // Update every 15 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-mve.stopChan:
			return
		case <-ticker.C:
			mve.updatePositionSizeRecommendations(ctx)
		}
	}
}

// alertManagementLoop manages volatility alerts
func (mve *MarketVolatilityEngine) alertManagementLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-mve.stopChan:
			return
		case <-ticker.C:
			mve.manageVolatilityAlerts(ctx)
		}
	}
}

// performVolatilityAnalysis performs comprehensive volatility analysis
func (mve *MarketVolatilityEngine) performVolatilityAnalysis(ctx context.Context) {
	mve.logger.Debug("Performing volatility analysis")

	for _, asset := range mve.config.MonitoredAssets {
		// Generate mock price data for analysis
		priceData := mve.generateMockPriceData(asset, 30) // 30 data points

		_, err := mve.AnalyzeAssetVolatility(ctx, asset, priceData)
		if err != nil {
			mve.logger.Error("Failed to analyze asset volatility",
				zap.String("asset", asset),
				zap.Error(err))
		}
	}

	// Update real-time metrics
	mve.updateRealTimeMetrics()
}

// generateMockPriceData generates mock price data for testing
func (mve *MarketVolatilityEngine) generateMockPriceData(asset string, count int) []PriceMovement {
	data := make([]PriceMovement, count)
	basePrice := decimal.NewFromFloat(1000) // $1000 base price

	for i := 0; i < count; i++ {
		// Simulate price movements with some volatility
		change := decimal.NewFromFloat(float64(i%20-10) * 0.02) // ±20% variation
		price := basePrice.Mul(decimal.NewFromFloat(1.0).Add(change))

		var returnVal decimal.Decimal
		if i > 0 {
			returnVal = price.Div(data[i-1].Price).Sub(decimal.NewFromFloat(1.0))
		}

		data[i] = PriceMovement{
			Timestamp:     time.Now().Add(-time.Duration(count-i) * time.Hour),
			Price:         price,
			Return:        returnVal,
			AbsReturn:     returnVal.Abs(),
			SquaredReturn: returnVal.Mul(returnVal),
		}
	}

	return data
}

// detectMarketRegimeChanges detects changes in market regimes
func (mve *MarketVolatilityEngine) detectMarketRegimeChanges(ctx context.Context) {
	mve.logger.Debug("Detecting market regime changes")

	// Analyze overall market conditions
	overallVolatility := mve.calculateOverallMarketVolatility()

	// Determine current regime
	var currentRegime string
	if overallVolatility.LessThan(decimal.NewFromFloat(0.2)) {
		currentRegime = "low_volatility"
	} else if overallVolatility.LessThan(decimal.NewFromFloat(0.4)) {
		currentRegime = "normal"
	} else if overallVolatility.LessThan(decimal.NewFromFloat(0.6)) {
		currentRegime = "high_volatility"
	} else {
		currentRegime = "extreme_volatility"
	}

	// Check for regime change
	mve.mutex.RLock()
	previousRegime := "normal" // Default previous regime
	if len(mve.realTimeMetrics.RegimeChanges) > 0 {
		previousRegime = mve.realTimeMetrics.RegimeChanges[len(mve.realTimeMetrics.RegimeChanges)-1].ToRegime
	}
	mve.mutex.RUnlock()

	if currentRegime != previousRegime {
		regimeChange := RegimeChange{
			FromRegime:   previousRegime,
			ToRegime:     currentRegime,
			Confidence:   decimal.NewFromFloat(0.8),
			DetectedAt:   time.Now(),
			TriggerEvent: "volatility_threshold",
			Impact:       "position_sizing_adjustment",
		}

		mve.mutex.Lock()
		mve.realTimeMetrics.RegimeChanges = append(mve.realTimeMetrics.RegimeChanges, regimeChange)
		mve.mutex.Unlock()

		mve.logger.Info("Market regime change detected",
			zap.String("from", previousRegime),
			zap.String("to", currentRegime))
	}
}

// calculateOverallMarketVolatility calculates overall market volatility
func (mve *MarketVolatilityEngine) calculateOverallMarketVolatility() decimal.Decimal {
	mve.mutex.RLock()
	defer mve.mutex.RUnlock()

	if len(mve.assetVolatilities) == 0 {
		return decimal.NewFromFloat(0.3) // Default 30%
	}

	totalVol := decimal.Zero
	count := 0

	for _, volatility := range mve.assetVolatilities {
		totalVol = totalVol.Add(volatility.CurrentVolatility)
		count++
	}

	return totalVol.Div(decimal.NewFromInt(int64(count)))
}

// updatePositionSizeRecommendations updates position size recommendations for all assets
func (mve *MarketVolatilityEngine) updatePositionSizeRecommendations(ctx context.Context) {
	mve.logger.Debug("Updating position size recommendations")

	portfolioValue := decimal.NewFromFloat(100000) // Mock $100k portfolio

	for _, asset := range mve.config.MonitoredAssets {
		_, err := mve.GeneratePositionSizeRecommendation(ctx, asset, portfolioValue)
		if err != nil {
			mve.logger.Error("Failed to update position size recommendation",
				zap.String("asset", asset),
				zap.Error(err))
		}
	}
}

// updateRealTimeMetrics updates real-time volatility metrics
func (mve *MarketVolatilityEngine) updateRealTimeMetrics() {
	mve.mutex.Lock()
	defer mve.mutex.Unlock()

	mve.realTimeMetrics.OverallMarketVolatility = mve.calculateOverallMarketVolatility()
	mve.realTimeMetrics.VolatilityIndex = mve.realTimeMetrics.OverallMarketVolatility.Mul(decimal.NewFromFloat(100))
	mve.realTimeMetrics.MarketStressIndex = mve.calculateMarketStressIndex()
	mve.realTimeMetrics.LastUpdated = time.Now()

	// Update cross-asset volatility
	mve.realTimeMetrics.CrossAssetVolatility = make(map[string]decimal.Decimal)
	for asset, volatility := range mve.assetVolatilities {
		mve.realTimeMetrics.CrossAssetVolatility[asset] = volatility.CurrentVolatility
	}
}

// calculateMarketStressIndex calculates market stress index
func (mve *MarketVolatilityEngine) calculateMarketStressIndex() decimal.Decimal {
	// Simplified stress index based on volatility levels
	overallVol := mve.calculateOverallMarketVolatility()

	// Normalize to 0-100 scale
	stressIndex := overallVol.Mul(decimal.NewFromFloat(200)) // Scale factor

	if stressIndex.GreaterThan(decimal.NewFromFloat(100)) {
		stressIndex = decimal.NewFromFloat(100)
	}

	return stressIndex
}

// manageVolatilityAlerts manages active volatility alerts
func (mve *MarketVolatilityEngine) manageVolatilityAlerts(ctx context.Context) {
	mve.mutex.Lock()
	defer mve.mutex.Unlock()

	// Clean up old resolved alerts
	for id, alert := range mve.volatilityAlerts {
		if alert.Status == "resolved" && time.Since(alert.CreatedAt) > 24*time.Hour {
			delete(mve.volatilityAlerts, id)
		}
	}
}

// Public interface methods

// GetAssetVolatilities returns volatility data for all monitored assets
func (mve *MarketVolatilityEngine) GetAssetVolatilities() map[string]*AssetVolatility {
	mve.mutex.RLock()
	defer mve.mutex.RUnlock()

	result := make(map[string]*AssetVolatility)
	for asset, volatility := range mve.assetVolatilities {
		result[asset] = volatility
	}

	return result
}

// GetAssetVolatility returns volatility data for a specific asset
func (mve *MarketVolatilityEngine) GetAssetVolatility(asset string) (*AssetVolatility, error) {
	mve.mutex.RLock()
	defer mve.mutex.RUnlock()

	volatility, exists := mve.assetVolatilities[asset]
	if !exists {
		return nil, fmt.Errorf("volatility data not found for asset: %s", asset)
	}

	return volatility, nil
}

// GetCorrelationMatrix returns the current correlation matrix
func (mve *MarketVolatilityEngine) GetCorrelationMatrix() map[string]map[string]decimal.Decimal {
	mve.mutex.RLock()
	defer mve.mutex.RUnlock()

	result := make(map[string]map[string]decimal.Decimal)
	for asset1, correlations := range mve.correlationMatrix {
		result[asset1] = make(map[string]decimal.Decimal)
		for asset2, corr := range correlations {
			result[asset1][asset2] = corr
		}
	}

	return result
}

// GetPositionSizeRecommendations returns position size recommendations for all assets
func (mve *MarketVolatilityEngine) GetPositionSizeRecommendations() map[string]*PositionSizeRecommendation {
	mve.mutex.RLock()
	defer mve.mutex.RUnlock()

	result := make(map[string]*PositionSizeRecommendation)
	for asset, recommendation := range mve.positionSizes {
		result[asset] = recommendation
	}

	return result
}

// GetPositionSizeRecommendation returns position size recommendation for a specific asset
func (mve *MarketVolatilityEngine) GetPositionSizeRecommendation(asset string) (*PositionSizeRecommendation, error) {
	mve.mutex.RLock()
	defer mve.mutex.RUnlock()

	recommendation, exists := mve.positionSizes[asset]
	if !exists {
		return nil, fmt.Errorf("position size recommendation not found for asset: %s", asset)
	}

	return recommendation, nil
}

// GetVolatilityAlerts returns all volatility alerts
func (mve *MarketVolatilityEngine) GetVolatilityAlerts() []*VolatilityAlert {
	mve.mutex.RLock()
	defer mve.mutex.RUnlock()

	alerts := make([]*VolatilityAlert, 0, len(mve.volatilityAlerts))
	for _, alert := range mve.volatilityAlerts {
		alerts = append(alerts, alert)
	}

	return alerts
}

// GetActiveVolatilityAlerts returns only active volatility alerts
func (mve *MarketVolatilityEngine) GetActiveVolatilityAlerts() []*VolatilityAlert {
	mve.mutex.RLock()
	defer mve.mutex.RUnlock()

	alerts := make([]*VolatilityAlert, 0)
	for _, alert := range mve.volatilityAlerts {
		if alert.Status == "active" {
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// GetRealTimeMetrics returns real-time volatility metrics
func (mve *MarketVolatilityEngine) GetRealTimeMetrics() *RealTimeVolatilityMetrics {
	mve.mutex.RLock()
	defer mve.mutex.RUnlock()

	return mve.realTimeMetrics
}

// GetHistoricalData returns historical volatility data for an asset
func (mve *MarketVolatilityEngine) GetHistoricalData(asset string) (*HistoricalVolatilityData, error) {
	mve.mutex.RLock()
	defer mve.mutex.RUnlock()

	data, exists := mve.historicalData[asset]
	if !exists {
		return nil, fmt.Errorf("historical data not found for asset: %s", asset)
	}

	return data, nil
}

// ResolveVolatilityAlert resolves a volatility alert
func (mve *MarketVolatilityEngine) ResolveVolatilityAlert(alertID string) error {
	mve.mutex.Lock()
	defer mve.mutex.Unlock()

	alert, exists := mve.volatilityAlerts[alertID]
	if !exists {
		return fmt.Errorf("volatility alert not found: %s", alertID)
	}

	alert.Status = "resolved"
	now := time.Now()
	alert.ResolvedAt = &now

	mve.logger.Info("Volatility alert resolved", zap.String("alert_id", alertID))
	return nil
}

// DismissVolatilityAlert dismisses a volatility alert
func (mve *MarketVolatilityEngine) DismissVolatilityAlert(alertID string) error {
	mve.mutex.Lock()
	defer mve.mutex.Unlock()

	alert, exists := mve.volatilityAlerts[alertID]
	if !exists {
		return fmt.Errorf("volatility alert not found: %s", alertID)
	}

	alert.Status = "dismissed"
	now := time.Now()
	alert.ResolvedAt = &now

	mve.logger.Info("Volatility alert dismissed", zap.String("alert_id", alertID))
	return nil
}

// UpdateConfig updates the volatility engine configuration
func (mve *MarketVolatilityEngine) UpdateConfig(config VolatilityConfig) error {
	mve.mutex.Lock()
	defer mve.mutex.Unlock()

	mve.config = config
	mve.logger.Info("Market volatility engine configuration updated",
		zap.Duration("analysis_interval", config.AnalysisInterval),
		zap.Duration("volatility_window", config.VolatilityWindow),
		zap.Bool("enable_position_sizing", config.EnablePositionSizing),
		zap.Strings("monitored_assets", config.MonitoredAssets))

	return nil
}

// GetConfig returns the current configuration
func (mve *MarketVolatilityEngine) GetConfig() VolatilityConfig {
	mve.mutex.RLock()
	defer mve.mutex.RUnlock()

	return mve.config
}

// IsRunning returns whether the volatility engine is running
func (mve *MarketVolatilityEngine) IsRunning() bool {
	mve.mutex.RLock()
	defer mve.mutex.RUnlock()

	return mve.isRunning
}

// GetDefaultVolatilityConfig returns default volatility engine configuration
func GetDefaultVolatilityConfig() VolatilityConfig {
	return VolatilityConfig{
		Enabled:                   true,
		AnalysisInterval:          1 * time.Minute,
		CorrelationUpdateInterval: 5 * time.Minute,
		VolatilityWindow:          24 * time.Hour,
		HighVolatilityThreshold:   decimal.NewFromFloat(0.5),  // 50%
		LowVolatilityThreshold:    decimal.NewFromFloat(0.1),  // 10%
		CorrelationThreshold:      decimal.NewFromFloat(0.8),  // 80%
		MaxPositionSize:           decimal.NewFromFloat(0.2),  // 20%
		MinPositionSize:           decimal.NewFromFloat(0.01), // 1%
		RiskAdjustmentFactor:      decimal.NewFromFloat(1.0),
		EnableRealTimeAlerts:      true,
		EnablePositionSizing:      true,
		MonitoredAssets:           []string{"BTC", "ETH", "USDC", "USDT"},
	}
}
