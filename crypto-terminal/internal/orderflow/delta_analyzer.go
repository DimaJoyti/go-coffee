package orderflow

import (
	"fmt"
	"sort"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/google/uuid"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// DeltaAnalyzer analyzes order flow delta and buying/selling pressure
type DeltaAnalyzer struct {
	config *config.Config
}

// NewDeltaAnalyzer creates a new delta analyzer
func NewDeltaAnalyzer(config *config.Config) *DeltaAnalyzer {
	return &DeltaAnalyzer{
		config: config,
	}
}

// AnalyzeDelta performs comprehensive delta analysis on tick data
func (da *DeltaAnalyzer) AnalyzeDelta(ticks []models.Tick, config models.OrderFlowConfig) (*models.DeltaProfile, error) {
	if len(ticks) == 0 {
		return nil, fmt.Errorf("no ticks provided for delta analysis")
	}

	// Sort ticks by timestamp
	sort.Slice(ticks, func(i, j int) bool {
		return ticks[i].Timestamp.Before(ticks[j].Timestamp)
	})

	profile := &models.DeltaProfile{
		ID:        uuid.New().String(),
		Symbol:    ticks[0].Symbol,
		Timeframe: config.TickAggregationMethod,
		StartTime: ticks[0].Timestamp,
		EndTime:   ticks[len(ticks)-1].Timestamp,
		CreatedAt: time.Now(),
	}

	// Calculate cumulative delta
	da.calculateCumulativeDelta(profile, ticks)

	// Calculate delta momentum and acceleration
	da.calculateDeltaMomentum(profile, ticks, config)

	// Calculate buying and selling pressure
	da.calculatePressureMetrics(profile, ticks)

	// Detect delta divergences
	da.detectDeltaDivergence(profile, ticks)

	// Detect delta exhaustion
	da.detectDeltaExhaustion(profile, ticks)

	return profile, nil
}

// calculateCumulativeDelta calculates cumulative delta from ticks
func (da *DeltaAnalyzer) calculateCumulativeDelta(profile *models.DeltaProfile, ticks []models.Tick) {
	cumulativeDelta := decimal.Zero
	deltaHigh := decimal.Zero
	deltaLow := decimal.Zero
	
	for i, tick := range ticks {
		var tickDelta decimal.Decimal
		
		if tick.Side == "BUY" {
			tickDelta = tick.Volume
		} else if tick.Side == "SELL" {
			tickDelta = tick.Volume.Neg()
		}
		
		cumulativeDelta = cumulativeDelta.Add(tickDelta)
		
		// Track delta high and low
		if i == 0 {
			deltaHigh = cumulativeDelta
			deltaLow = cumulativeDelta
		} else {
			if cumulativeDelta.GreaterThan(deltaHigh) {
				deltaHigh = cumulativeDelta
			}
			if cumulativeDelta.LessThan(deltaLow) {
				deltaLow = cumulativeDelta
			}
		}
	}
	
	profile.CumulativeDelta = cumulativeDelta
	profile.DeltaHigh = deltaHigh
	profile.DeltaLow = deltaLow
	profile.DeltaRange = deltaHigh.Sub(deltaLow)
}

// calculateDeltaMomentum calculates delta momentum and acceleration
func (da *DeltaAnalyzer) calculateDeltaMomentum(profile *models.DeltaProfile, ticks []models.Tick, config models.OrderFlowConfig) {
	if len(ticks) < 2 {
		return
	}

	// Calculate delta momentum using smoothing period
	smoothingPeriod := config.DeltaSmoothingPeriod
	if smoothingPeriod <= 0 {
		smoothingPeriod = 10 // Default smoothing period
	}

	// Group ticks into periods for momentum calculation
	periodSize := len(ticks) / smoothingPeriod
	if periodSize < 1 {
		periodSize = 1
	}

	var deltaValues []decimal.Decimal
	
	for i := 0; i < len(ticks); i += periodSize {
		endIdx := i + periodSize
		if endIdx > len(ticks) {
			endIdx = len(ticks)
		}
		
		periodDelta := decimal.Zero
		for j := i; j < endIdx; j++ {
			if ticks[j].Side == "BUY" {
				periodDelta = periodDelta.Add(ticks[j].Volume)
			} else if ticks[j].Side == "SELL" {
				periodDelta = periodDelta.Sub(ticks[j].Volume)
			}
		}
		
		deltaValues = append(deltaValues, periodDelta)
	}

	// Calculate momentum (rate of change)
	if len(deltaValues) >= 2 {
		recent := deltaValues[len(deltaValues)-1]
		previous := deltaValues[len(deltaValues)-2]
		profile.DeltaMomentum = recent.Sub(previous)
	}

	// Calculate acceleration (rate of change of momentum)
	if len(deltaValues) >= 3 {
		current := deltaValues[len(deltaValues)-1]
		previous := deltaValues[len(deltaValues)-2]
		beforePrevious := deltaValues[len(deltaValues)-3]
		
		currentMomentum := current.Sub(previous)
		previousMomentum := previous.Sub(beforePrevious)
		profile.DeltaAcceleration = currentMomentum.Sub(previousMomentum)
	}
}

// calculatePressureMetrics calculates buying and selling pressure metrics
func (da *DeltaAnalyzer) calculatePressureMetrics(profile *models.DeltaProfile, ticks []models.Tick) {
	buyVolume := decimal.Zero
	sellVolume := decimal.Zero
	totalVolume := decimal.Zero
	
	for _, tick := range ticks {
		totalVolume = totalVolume.Add(tick.Volume)
		
		if tick.Side == "BUY" {
			buyVolume = buyVolume.Add(tick.Volume)
		} else if tick.Side == "SELL" {
			sellVolume = sellVolume.Add(tick.Volume)
		}
	}
	
	// Calculate pressure percentages
	if totalVolume.GreaterThan(decimal.Zero) {
		profile.BuyPressure = buyVolume.Div(totalVolume).Mul(decimal.NewFromFloat(100))
		profile.SellPressure = sellVolume.Div(totalVolume).Mul(decimal.NewFromFloat(100))
	}
	
	// Calculate net pressure
	profile.NetPressure = profile.BuyPressure.Sub(profile.SellPressure)
	
	// Calculate delta strength (absolute value of net pressure)
	profile.DeltaStrength = profile.NetPressure
	if profile.DeltaStrength.LessThan(decimal.Zero) {
		profile.DeltaStrength = profile.DeltaStrength.Neg()
	}
}

// detectDeltaDivergence detects delta divergence patterns
func (da *DeltaAnalyzer) detectDeltaDivergence(profile *models.DeltaProfile, ticks []models.Tick) {
	// This is a simplified divergence detection
	// In a real implementation, you would compare price action with delta
	
	// For now, we'll detect divergence based on delta momentum vs delta strength
	if profile.DeltaMomentum.LessThan(decimal.Zero) && profile.DeltaStrength.GreaterThan(decimal.NewFromFloat(60)) {
		profile.IsDivergent = true
		logrus.Debugf("Delta divergence detected for %s: momentum=%s, strength=%s", 
			profile.Symbol, profile.DeltaMomentum.String(), profile.DeltaStrength.String())
	}
}

// detectDeltaExhaustion detects delta exhaustion patterns
func (da *DeltaAnalyzer) detectDeltaExhaustion(profile *models.DeltaProfile, ticks []models.Tick) {
	// Detect exhaustion when delta strength is very high but momentum is declining
	if profile.DeltaStrength.GreaterThan(decimal.NewFromFloat(80)) && 
	   profile.DeltaMomentum.LessThan(decimal.NewFromFloat(-10)) {
		profile.IsExhausted = true
		logrus.Debugf("Delta exhaustion detected for %s: strength=%s, momentum=%s", 
			profile.Symbol, profile.DeltaStrength.String(), profile.DeltaMomentum.String())
	}
}

// AnalyzeDeltaHistory analyzes historical delta patterns
func (da *DeltaAnalyzer) AnalyzeDeltaHistory(historicalProfiles []models.DeltaProfile) map[string]interface{} {
	if len(historicalProfiles) == 0 {
		return map[string]interface{}{
			"total_profiles": 0,
		}
	}

	// Sort by timestamp
	sort.Slice(historicalProfiles, func(i, j int) bool {
		return historicalProfiles[i].StartTime.Before(historicalProfiles[j].StartTime)
	})

	// Calculate statistics
	totalProfiles := len(historicalProfiles)
	divergentCount := 0
	exhaustedCount := 0
	
	var avgDeltaStrength decimal.Decimal
	var avgBuyPressure decimal.Decimal
	var avgSellPressure decimal.Decimal
	
	for _, profile := range historicalProfiles {
		if profile.IsDivergent {
			divergentCount++
		}
		if profile.IsExhausted {
			exhaustedCount++
		}
		
		avgDeltaStrength = avgDeltaStrength.Add(profile.DeltaStrength)
		avgBuyPressure = avgBuyPressure.Add(profile.BuyPressure)
		avgSellPressure = avgSellPressure.Add(profile.SellPressure)
	}
	
	// Calculate averages
	totalProfilesDecimal := decimal.NewFromInt(int64(totalProfiles))
	avgDeltaStrength = avgDeltaStrength.Div(totalProfilesDecimal)
	avgBuyPressure = avgBuyPressure.Div(totalProfilesDecimal)
	avgSellPressure = avgSellPressure.Div(totalProfilesDecimal)
	
	// Calculate trend
	trend := "NEUTRAL"
	if len(historicalProfiles) >= 2 {
		recent := historicalProfiles[len(historicalProfiles)-1]
		previous := historicalProfiles[len(historicalProfiles)-2]
		
		if recent.CumulativeDelta.GreaterThan(previous.CumulativeDelta) {
			trend = "BULLISH"
		} else if recent.CumulativeDelta.LessThan(previous.CumulativeDelta) {
			trend = "BEARISH"
		}
	}
	
	return map[string]interface{}{
		"total_profiles":       totalProfiles,
		"divergent_count":      divergentCount,
		"exhausted_count":      exhaustedCount,
		"divergence_rate":      float64(divergentCount) / float64(totalProfiles) * 100,
		"exhaustion_rate":      float64(exhaustedCount) / float64(totalProfiles) * 100,
		"avg_delta_strength":   avgDeltaStrength.String(),
		"avg_buy_pressure":     avgBuyPressure.String(),
		"avg_sell_pressure":    avgSellPressure.String(),
		"current_trend":        trend,
	}
}

// CalculateRealTimeDelta calculates real-time delta metrics
func (da *DeltaAnalyzer) CalculateRealTimeDelta(recentTicks []models.Tick, windowSize int) map[string]interface{} {
	if len(recentTicks) == 0 {
		return map[string]interface{}{
			"cumulative_delta": "0",
			"buy_pressure": "0",
			"sell_pressure": "0",
		}
	}

	// Use only the most recent ticks within the window
	startIdx := 0
	if len(recentTicks) > windowSize {
		startIdx = len(recentTicks) - windowSize
	}
	
	windowTicks := recentTicks[startIdx:]
	
	cumulativeDelta := decimal.Zero
	buyVolume := decimal.Zero
	sellVolume := decimal.Zero
	totalVolume := decimal.Zero
	
	for _, tick := range windowTicks {
		totalVolume = totalVolume.Add(tick.Volume)
		
		if tick.Side == "BUY" {
			buyVolume = buyVolume.Add(tick.Volume)
			cumulativeDelta = cumulativeDelta.Add(tick.Volume)
		} else if tick.Side == "SELL" {
			sellVolume = sellVolume.Add(tick.Volume)
			cumulativeDelta = cumulativeDelta.Sub(tick.Volume)
		}
	}
	
	buyPressure := decimal.Zero
	sellPressure := decimal.Zero
	
	if totalVolume.GreaterThan(decimal.Zero) {
		buyPressure = buyVolume.Div(totalVolume).Mul(decimal.NewFromFloat(100))
		sellPressure = sellVolume.Div(totalVolume).Mul(decimal.NewFromFloat(100))
	}
	
	return map[string]interface{}{
		"cumulative_delta": cumulativeDelta.String(),
		"buy_pressure":     buyPressure.String(),
		"sell_pressure":    sellPressure.String(),
		"net_pressure":     buyPressure.Sub(sellPressure).String(),
		"total_volume":     totalVolume.String(),
		"window_size":      len(windowTicks),
	}
}

// DetectDeltaSignals detects trading signals based on delta analysis
func (da *DeltaAnalyzer) DetectDeltaSignals(profile *models.DeltaProfile, priceData []decimal.Decimal) []map[string]interface{} {
	var signals []map[string]interface{}
	
	// Signal 1: Strong buying pressure with positive momentum
	if profile.BuyPressure.GreaterThan(decimal.NewFromFloat(70)) && 
	   profile.DeltaMomentum.GreaterThan(decimal.Zero) {
		signals = append(signals, map[string]interface{}{
			"type":        "BUY_PRESSURE",
			"signal":      "BULLISH",
			"strength":    "HIGH",
			"description": "Strong buying pressure with positive momentum",
			"confidence":  85,
		})
	}
	
	// Signal 2: Strong selling pressure with negative momentum
	if profile.SellPressure.GreaterThan(decimal.NewFromFloat(70)) && 
	   profile.DeltaMomentum.LessThan(decimal.Zero) {
		signals = append(signals, map[string]interface{}{
			"type":        "SELL_PRESSURE",
			"signal":      "BEARISH",
			"strength":    "HIGH",
			"description": "Strong selling pressure with negative momentum",
			"confidence":  85,
		})
	}
	
	// Signal 3: Delta divergence
	if profile.IsDivergent {
		signals = append(signals, map[string]interface{}{
			"type":        "DIVERGENCE",
			"signal":      "REVERSAL",
			"strength":    "MEDIUM",
			"description": "Delta divergence detected - potential reversal",
			"confidence":  70,
		})
	}
	
	// Signal 4: Delta exhaustion
	if profile.IsExhausted {
		signals = append(signals, map[string]interface{}{
			"type":        "EXHAUSTION",
			"signal":      "REVERSAL",
			"strength":    "HIGH",
			"description": "Delta exhaustion detected - trend may reverse",
			"confidence":  80,
		})
	}
	
	return signals
}

// GetDeltaSummary returns a summary of delta analysis
func (da *DeltaAnalyzer) GetDeltaSummary(profile *models.DeltaProfile) map[string]interface{} {
	if profile == nil {
		return map[string]interface{}{
			"cumulative_delta": "0",
			"delta_strength": "0",
		}
	}

	// Determine overall sentiment
	sentiment := "NEUTRAL"
	if profile.NetPressure.GreaterThan(decimal.NewFromFloat(20)) {
		sentiment = "BULLISH"
	} else if profile.NetPressure.LessThan(decimal.NewFromFloat(-20)) {
		sentiment = "BEARISH"
	}
	
	// Determine momentum direction
	momentumDirection := "NEUTRAL"
	if profile.DeltaMomentum.GreaterThan(decimal.Zero) {
		momentumDirection = "POSITIVE"
	} else if profile.DeltaMomentum.LessThan(decimal.Zero) {
		momentumDirection = "NEGATIVE"
	}
	
	return map[string]interface{}{
		"cumulative_delta":    profile.CumulativeDelta.String(),
		"delta_high":          profile.DeltaHigh.String(),
		"delta_low":           profile.DeltaLow.String(),
		"delta_range":         profile.DeltaRange.String(),
		"delta_momentum":      profile.DeltaMomentum.String(),
		"delta_acceleration":  profile.DeltaAcceleration.String(),
		"buy_pressure":        profile.BuyPressure.String(),
		"sell_pressure":       profile.SellPressure.String(),
		"net_pressure":        profile.NetPressure.String(),
		"delta_strength":      profile.DeltaStrength.String(),
		"is_divergent":        profile.IsDivergent,
		"is_exhausted":        profile.IsExhausted,
		"sentiment":           sentiment,
		"momentum_direction":  momentumDirection,
	}
}
