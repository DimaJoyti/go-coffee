package orderflow

import (
	"sort"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

)

// ImbalanceDetector detects order flow imbalances and absorption patterns
type ImbalanceDetector struct {
	config *config.Config
}

// NewImbalanceDetector creates a new imbalance detector
func NewImbalanceDetector(config *config.Config) *ImbalanceDetector {
	return &ImbalanceDetector{
		config: config,
	}
}

// DetectImbalances detects various types of order flow imbalances
func (id *ImbalanceDetector) DetectImbalances(ticks []models.Tick, config models.OrderFlowConfig) ([]models.OrderFlowImbalance, error) {
	if len(ticks) == 0 {
		return []models.OrderFlowImbalance{}, nil
	}

	var imbalances []models.OrderFlowImbalance

	// Detect volume imbalances
	volumeImbalances := id.detectVolumeImbalances(ticks, config)
	imbalances = append(imbalances, volumeImbalances...)

	// Detect bid/ask stack imbalances
	stackImbalances := id.detectStackImbalances(ticks, config)
	imbalances = append(imbalances, stackImbalances...)

	// Detect absorption patterns
	absorptionPatterns := id.detectAbsorptionPatterns(ticks, config)
	imbalances = append(imbalances, absorptionPatterns...)

	return imbalances, nil
}

// detectVolumeImbalances detects volume imbalances at price levels
func (id *ImbalanceDetector) detectVolumeImbalances(ticks []models.Tick, config models.OrderFlowConfig) []models.OrderFlowImbalance {
	var imbalances []models.OrderFlowImbalance

	// Group ticks by price levels
	priceGroups := make(map[string]*volumeData)
	
	for _, tick := range ticks {
		priceLevel := id.roundToTickSize(tick.Price, config.PriceTickSize)
		priceKey := priceLevel.String()
		
		if priceGroups[priceKey] == nil {
			priceGroups[priceKey] = &volumeData{
				price:     priceLevel,
				buyVolume: decimal.Zero,
				sellVolume: decimal.Zero,
				startTime: tick.Timestamp,
			}
		}
		
		data := priceGroups[priceKey]
		if tick.Side == "BUY" {
			data.buyVolume = data.buyVolume.Add(tick.Volume)
		} else if tick.Side == "SELL" {
			data.sellVolume = data.sellVolume.Add(tick.Volume)
		}
		data.endTime = tick.Timestamp
	}

	// Analyze each price level for imbalances
	for _, data := range priceGroups {
		totalVolume := data.buyVolume.Add(data.sellVolume)
		
		// Skip if total volume is below minimum threshold
		if totalVolume.LessThan(config.ImbalanceMinVolume) {
			continue
		}

		// Calculate imbalance ratio
		var imbalanceRatio decimal.Decimal
		var imbalanceType string
		var severity string

		if data.sellVolume.IsZero() {
			if data.buyVolume.GreaterThan(decimal.Zero) {
				imbalanceRatio = decimal.NewFromFloat(100) // 100% buy
				imbalanceType = "BID_STACK"
			} else {
				continue
			}
		} else {
			ratio := data.buyVolume.Div(data.sellVolume)
			imbalanceRatio = ratio.Sub(decimal.NewFromFloat(1)).Mul(decimal.NewFromFloat(100))
			
			if imbalanceRatio.GreaterThan(config.ImbalanceThreshold) {
				imbalanceType = "BID_STACK"
			} else if imbalanceRatio.LessThan(config.ImbalanceThreshold.Neg()) {
				imbalanceType = "ASK_STACK"
				imbalanceRatio = imbalanceRatio.Neg()
			} else {
				continue // No significant imbalance
			}
		}

		// Determine severity
		absRatio := imbalanceRatio
		if absRatio.LessThan(decimal.Zero) {
			absRatio = absRatio.Neg()
		}

		if absRatio.GreaterThan(decimal.NewFromFloat(300)) {
			severity = "EXTREME"
		} else if absRatio.GreaterThan(decimal.NewFromFloat(200)) {
			severity = "HIGH"
		} else if absRatio.GreaterThan(decimal.NewFromFloat(100)) {
			severity = "MEDIUM"
		} else {
			severity = "LOW"
		}

		// Create imbalance record
		imbalance := models.OrderFlowImbalance{
			ID:             uuid.New().String(),
			Symbol:         ticks[0].Symbol,
			Price:          data.price,
			ImbalanceType:  imbalanceType,
			Severity:       severity,
			BuyVolume:      data.buyVolume,
			SellVolume:     data.sellVolume,
			ImbalanceRatio: imbalanceRatio,
			Duration:       data.endTime.Sub(data.startTime),
			IsActive:       true,
			IsResolved:     false,
			DetectedAt:     data.startTime,
			CreatedAt:      time.Now(),
		}

		imbalances = append(imbalances, imbalance)
	}

	return imbalances
}

// detectStackImbalances detects bid/ask stack imbalances
func (id *ImbalanceDetector) detectStackImbalances(ticks []models.Tick, config models.OrderFlowConfig) []models.OrderFlowImbalance {
	var imbalances []models.OrderFlowImbalance

	// This would require order book data, which we don't have from trade ticks alone
	// In a real implementation, you would analyze the order book depth
	// For now, we'll detect stack imbalances based on consecutive trades at the same price

	if len(ticks) < 5 {
		return imbalances
	}

	// Look for consecutive trades of the same side at similar prices
	consecutiveThreshold := 5
	priceTolerancePercent := decimal.NewFromFloat(0.1) // 0.1% price tolerance

	for i := 0; i <= len(ticks)-consecutiveThreshold; i++ {
		// Check if we have consecutive trades of the same side
		baseTick := ticks[i]
		consecutiveCount := 1
		totalVolume := baseTick.Volume
		
		for j := i + 1; j < len(ticks) && j < i+consecutiveThreshold*2; j++ {
			currentTick := ticks[j]
			
			// Check if price is within tolerance
			priceDiff := currentTick.Price.Sub(baseTick.Price).Abs()
			priceThreshold := baseTick.Price.Mul(priceTolerancePercent).Div(decimal.NewFromFloat(100))
			
			if currentTick.Side == baseTick.Side && priceDiff.LessThanOrEqual(priceThreshold) {
				consecutiveCount++
				totalVolume = totalVolume.Add(currentTick.Volume)
			} else {
				break
			}
		}

		// If we found enough consecutive trades, it might indicate a stack imbalance
		if consecutiveCount >= consecutiveThreshold && totalVolume.GreaterThanOrEqual(config.ImbalanceMinVolume) {
			imbalanceType := "BID_STACK"
			if baseTick.Side == "SELL" {
				imbalanceType = "ASK_STACK"
			}

			severity := "MEDIUM"
			if consecutiveCount >= 10 {
				severity = "HIGH"
			} else if consecutiveCount >= 15 {
				severity = "EXTREME"
			}

			imbalance := models.OrderFlowImbalance{
				ID:            uuid.New().String(),
				Symbol:        baseTick.Symbol,
				Price:         baseTick.Price,
				ImbalanceType: imbalanceType,
				Severity:      severity,
				BuyVolume:     totalVolume,
				SellVolume:    decimal.Zero,
				Duration:      ticks[i+consecutiveCount-1].Timestamp.Sub(baseTick.Timestamp),
				IsActive:      true,
				IsResolved:    false,
				DetectedAt:    baseTick.Timestamp,
				CreatedAt:     time.Now(),
			}

			if baseTick.Side == "SELL" {
				imbalance.BuyVolume = decimal.Zero
				imbalance.SellVolume = totalVolume
			}

			imbalances = append(imbalances, imbalance)
		}
	}

	return imbalances
}

// detectAbsorptionPatterns detects absorption patterns where large volume is absorbed without price movement
func (id *ImbalanceDetector) detectAbsorptionPatterns(ticks []models.Tick, config models.OrderFlowConfig) []models.OrderFlowImbalance {
	var imbalances []models.OrderFlowImbalance

	if len(ticks) < 10 {
		return imbalances
	}

	// Look for periods of high volume with minimal price movement
	windowSize := 20
	maxPriceMovementPercent := decimal.NewFromFloat(0.2) // 0.2% max price movement

	for i := 0; i <= len(ticks)-windowSize; i++ {
		windowTicks := ticks[i : i+windowSize]
		
		// Calculate price range and total volume in window
		minPrice := windowTicks[0].Price
		maxPrice := windowTicks[0].Price
		totalVolume := decimal.Zero
		buyVolume := decimal.Zero
		sellVolume := decimal.Zero

		for _, tick := range windowTicks {
			if tick.Price.LessThan(minPrice) {
				minPrice = tick.Price
			}
			if tick.Price.GreaterThan(maxPrice) {
				maxPrice = tick.Price
			}
			
			totalVolume = totalVolume.Add(tick.Volume)
			if tick.Side == "BUY" {
				buyVolume = buyVolume.Add(tick.Volume)
			} else if tick.Side == "SELL" {
				sellVolume = sellVolume.Add(tick.Volume)
			}
		}

		// Calculate price movement percentage
		priceRange := maxPrice.Sub(minPrice)
		avgPrice := minPrice.Add(maxPrice).Div(decimal.NewFromFloat(2))
		priceMovementPercent := priceRange.Div(avgPrice).Mul(decimal.NewFromFloat(100))

		// Check if this qualifies as absorption
		if priceMovementPercent.LessThanOrEqual(maxPriceMovementPercent) && 
		   totalVolume.GreaterThanOrEqual(config.ImbalanceMinVolume.Mul(decimal.NewFromFloat(2))) {
			
			// Determine which side is being absorbed
			var absorptionType string
			var dominantVolume decimal.Decimal
			
			if buyVolume.GreaterThan(sellVolume) {
				absorptionType = "BUY_ABSORPTION"
				dominantVolume = buyVolume
			} else {
				absorptionType = "SELL_ABSORPTION"
				dominantVolume = sellVolume
			}

			// Determine severity based on volume
			severity := "MEDIUM"
			if totalVolume.GreaterThan(config.ImbalanceMinVolume.Mul(decimal.NewFromFloat(5))) {
				severity = "HIGH"
			} else if totalVolume.GreaterThan(config.ImbalanceMinVolume.Mul(decimal.NewFromFloat(10))) {
				severity = "EXTREME"
			}

			imbalance := models.OrderFlowImbalance{
				ID:             uuid.New().String(),
				Symbol:         windowTicks[0].Symbol,
				Price:          avgPrice,
				ImbalanceType:  absorptionType,
				Severity:       severity,
				BuyVolume:      buyVolume,
				SellVolume:     sellVolume,
				ImbalanceRatio: dominantVolume.Div(totalVolume).Mul(decimal.NewFromFloat(100)),
				Duration:       windowTicks[windowSize-1].Timestamp.Sub(windowTicks[0].Timestamp),
				IsActive:       true,
				IsResolved:     false,
				DetectedAt:     windowTicks[0].Timestamp,
				CreatedAt:      time.Now(),
			}

			imbalances = append(imbalances, imbalance)
		}
	}

	return imbalances
}

// DetectDeltaDivergences detects delta divergences
func (id *ImbalanceDetector) DetectDeltaDivergences(ticks []models.Tick, deltaProfile models.DeltaProfile, config models.OrderFlowConfig) ([]models.OrderFlowImbalance, error) {
	var divergences []models.OrderFlowImbalance

	// This is a simplified divergence detection
	// In a real implementation, you would compare price action with delta patterns

	if deltaProfile.IsDivergent {
		divergence := models.OrderFlowImbalance{
			ID:            uuid.New().String(),
			Symbol:        deltaProfile.Symbol,
			Price:         decimal.Zero, // Would need current price
			ImbalanceType: "DELTA_DIVERGENCE",
			Severity:      "MEDIUM",
			BuyVolume:     decimal.Zero,
			SellVolume:    decimal.Zero,
			Duration:      deltaProfile.EndTime.Sub(deltaProfile.StartTime),
			IsActive:      true,
			IsResolved:    false,
			DetectedAt:    deltaProfile.StartTime,
			CreatedAt:     time.Now(),
		}

		// Determine severity based on delta strength
		if deltaProfile.DeltaStrength.GreaterThan(decimal.NewFromFloat(80)) {
			divergence.Severity = "HIGH"
		} else if deltaProfile.DeltaStrength.GreaterThan(decimal.NewFromFloat(90)) {
			divergence.Severity = "EXTREME"
		}

		divergences = append(divergences, divergence)
	}

	return divergences, nil
}

// ResolveImbalance marks an imbalance as resolved
func (id *ImbalanceDetector) ResolveImbalance(imbalance *models.OrderFlowImbalance, resolutionType string) {
	imbalance.IsResolved = true
	imbalance.IsActive = false
	imbalance.ResolutionType = resolutionType
	now := time.Now()
	imbalance.ResolvedAt = &now
}

// AnalyzeImbalanceResolution analyzes how imbalances are typically resolved
func (id *ImbalanceDetector) AnalyzeImbalanceResolution(imbalances []models.OrderFlowImbalance) map[string]interface{} {
	if len(imbalances) == 0 {
		return map[string]interface{}{
			"total_imbalances": 0,
		}
	}

	totalImbalances := len(imbalances)
	resolvedCount := 0
	absorptionCount := 0
	continuationCount := 0
	reversalCount := 0

	var avgDuration time.Duration
	totalDuration := time.Duration(0)

	for _, imbalance := range imbalances {
		if imbalance.IsResolved {
			resolvedCount++
			totalDuration += imbalance.Duration

			switch imbalance.ResolutionType {
			case "ABSORPTION":
				absorptionCount++
			case "CONTINUATION":
				continuationCount++
			case "REVERSAL":
				reversalCount++
			}
		}
	}

	if resolvedCount > 0 {
		avgDuration = totalDuration / time.Duration(resolvedCount)
	}

	resolutionRate := float64(resolvedCount) / float64(totalImbalances) * 100

	return map[string]interface{}{
		"total_imbalances":    totalImbalances,
		"resolved_count":      resolvedCount,
		"resolution_rate":     resolutionRate,
		"absorption_count":    absorptionCount,
		"continuation_count":  continuationCount,
		"reversal_count":      reversalCount,
		"avg_duration_ms":     avgDuration.Milliseconds(),
		"absorption_rate":     float64(absorptionCount) / float64(resolvedCount) * 100,
		"continuation_rate":   float64(continuationCount) / float64(resolvedCount) * 100,
		"reversal_rate":       float64(reversalCount) / float64(resolvedCount) * 100,
	}
}

// GetActiveImbalances returns currently active imbalances
func (id *ImbalanceDetector) GetActiveImbalances(imbalances []models.OrderFlowImbalance) []models.OrderFlowImbalance {
	var activeImbalances []models.OrderFlowImbalance

	for _, imbalance := range imbalances {
		if imbalance.IsActive && !imbalance.IsResolved {
			activeImbalances = append(activeImbalances, imbalance)
		}
	}

	// Sort by detection time (most recent first)
	sort.Slice(activeImbalances, func(i, j int) bool {
		return activeImbalances[i].DetectedAt.After(activeImbalances[j].DetectedAt)
	})

	return activeImbalances
}

// Helper types and methods

type volumeData struct {
	price      decimal.Decimal
	buyVolume  decimal.Decimal
	sellVolume decimal.Decimal
	startTime  time.Time
	endTime    time.Time
}

func (id *ImbalanceDetector) roundToTickSize(price, tickSize decimal.Decimal) decimal.Decimal {
	if tickSize.IsZero() {
		return price
	}
	
	divided := price.Div(tickSize)
	rounded := divided.Round(0)
	return rounded.Mul(tickSize)
}

// GetImbalanceSummary returns a summary of detected imbalances
func (id *ImbalanceDetector) GetImbalanceSummary(imbalances []models.OrderFlowImbalance) map[string]interface{} {
	if len(imbalances) == 0 {
		return map[string]interface{}{
			"total_imbalances": 0,
		}
	}

	totalImbalances := len(imbalances)
	activeCount := 0
	
	severityCounts := map[string]int{
		"LOW":     0,
		"MEDIUM":  0,
		"HIGH":    0,
		"EXTREME": 0,
	}
	
	typeCounts := map[string]int{
		"BID_STACK":        0,
		"ASK_STACK":        0,
		"VOLUME_IMBALANCE": 0,
		"BUY_ABSORPTION":   0,
		"SELL_ABSORPTION":  0,
		"DELTA_DIVERGENCE": 0,
	}

	for _, imbalance := range imbalances {
		if imbalance.IsActive {
			activeCount++
		}
		
		severityCounts[imbalance.Severity]++
		typeCounts[imbalance.ImbalanceType]++
	}

	return map[string]interface{}{
		"total_imbalances": totalImbalances,
		"active_count":     activeCount,
		"severity_counts":  severityCounts,
		"type_counts":      typeCounts,
		"active_rate":      float64(activeCount) / float64(totalImbalances) * 100,
	}
}
