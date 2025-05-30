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

// FootprintEngine generates footprint charts from tick data
type FootprintEngine struct {
	config *config.Config
}

// NewFootprintEngine creates a new footprint engine
func NewFootprintEngine(config *config.Config) *FootprintEngine {
	return &FootprintEngine{
		config: config,
	}
}

// GenerateFootprintBars generates footprint bars from tick data
func (fe *FootprintEngine) GenerateFootprintBars(ticks []models.Tick, config models.OrderFlowConfig) ([]models.FootprintBar, error) {
	if len(ticks) == 0 {
		return []models.FootprintBar{}, nil
	}

	// Sort ticks by timestamp
	sort.Slice(ticks, func(i, j int) bool {
		return ticks[i].Timestamp.Before(ticks[j].Timestamp)
	})

	var bars []models.FootprintBar
	
	switch config.TickAggregationMethod {
	case "TIME":
		bars = fe.aggregateByTime(ticks, config)
	case "VOLUME":
		bars = fe.aggregateByVolume(ticks, config)
	case "TICK_COUNT":
		bars = fe.aggregateByTickCount(ticks, config)
	default:
		return nil, fmt.Errorf("unsupported aggregation method: %s", config.TickAggregationMethod)
	}

	// Calculate additional metrics for each bar
	for i := range bars {
		fe.calculateBarMetrics(&bars[i], config)
	}

	// Identify Point of Control
	fe.identifyPointOfControl(bars)

	// Detect imbalances
	fe.detectVolumeImbalances(bars, config)

	return bars, nil
}

// aggregateByTime aggregates ticks by time periods
func (fe *FootprintEngine) aggregateByTime(ticks []models.Tick, config models.OrderFlowConfig) []models.FootprintBar {
	var bars []models.FootprintBar
	
	if len(ticks) == 0 {
		return bars
	}

	// Group ticks by time periods and price levels
	timeGroups := make(map[time.Time]map[string]*models.FootprintBar)
	
	for _, tick := range ticks {
		// Round timestamp to the configured time period
		periodStart := fe.roundToTimePeriod(tick.Timestamp, config.TimePerRow)
		
		// Round price to tick size
		priceLevel := fe.roundToTickSize(tick.Price, config.PriceTickSize)
		priceKey := priceLevel.String()
		
		if timeGroups[periodStart] == nil {
			timeGroups[periodStart] = make(map[string]*models.FootprintBar)
		}
		
		if timeGroups[periodStart][priceKey] == nil {
			timeGroups[periodStart][priceKey] = &models.FootprintBar{
				ID:         uuid.New().String(),
				Symbol:     tick.Symbol,
				Timeframe:  config.TickAggregationMethod,
				PriceLevel: priceLevel,
				TickSize:   config.PriceTickSize,
				StartTime:  periodStart,
				EndTime:    periodStart.Add(config.TimePerRow),
				CreatedAt:  time.Now(),
			}
		}
		
		bar := timeGroups[periodStart][priceKey]
		fe.addTickToBar(bar, tick)
	}
	
	// Convert map to slice
	for _, priceGroups := range timeGroups {
		for _, bar := range priceGroups {
			bars = append(bars, *bar)
		}
	}
	
	// Sort bars by time and price
	sort.Slice(bars, func(i, j int) bool {
		if bars[i].StartTime.Equal(bars[j].StartTime) {
			return bars[i].PriceLevel.LessThan(bars[j].PriceLevel)
		}
		return bars[i].StartTime.Before(bars[j].StartTime)
	})
	
	return bars
}

// aggregateByVolume aggregates ticks by volume thresholds
func (fe *FootprintEngine) aggregateByVolume(ticks []models.Tick, config models.OrderFlowConfig) []models.FootprintBar {
	var bars []models.FootprintBar
	
	if len(ticks) == 0 {
		return bars
	}

	// Group ticks by accumulated volume and price levels
	currentVolume := decimal.Zero
	volumeGroups := make(map[string]*models.FootprintBar)
	groupIndex := 0
	
	for _, tick := range ticks {
		priceLevel := fe.roundToTickSize(tick.Price, config.PriceTickSize)
		priceKey := priceLevel.String()
		
		if volumeGroups[priceKey] == nil {
			volumeGroups[priceKey] = &models.FootprintBar{
				ID:         uuid.New().String(),
				Symbol:     tick.Symbol,
				Timeframe:  config.TickAggregationMethod,
				PriceLevel: priceLevel,
				TickSize:   config.PriceTickSize,
				StartTime:  tick.Timestamp,
				CreatedAt:  time.Now(),
			}
		}
		
		bar := volumeGroups[priceKey]
		fe.addTickToBar(bar, tick)
		
		currentVolume = currentVolume.Add(tick.Volume)
		
		// Check if we've reached the volume threshold
		if currentVolume.GreaterThanOrEqual(config.VolumePerRow) {
			// Finalize current bars
			for _, bar := range volumeGroups {
				bar.EndTime = tick.Timestamp
				bars = append(bars, *bar)
			}
			
			// Reset for next group
			volumeGroups = make(map[string]*models.FootprintBar)
			currentVolume = decimal.Zero
			groupIndex++
		}
	}
	
	// Add remaining bars
	for _, bar := range volumeGroups {
		if !bar.StartTime.IsZero() {
			bars = append(bars, *bar)
		}
	}
	
	return bars
}

// aggregateByTickCount aggregates ticks by tick count
func (fe *FootprintEngine) aggregateByTickCount(ticks []models.Tick, config models.OrderFlowConfig) []models.FootprintBar {
	var bars []models.FootprintBar
	
	if len(ticks) == 0 {
		return bars
	}

	// Group ticks by count and price levels
	tickCount := 0
	tickGroups := make(map[string]*models.FootprintBar)
	
	for _, tick := range ticks {
		priceLevel := fe.roundToTickSize(tick.Price, config.PriceTickSize)
		priceKey := priceLevel.String()
		
		if tickGroups[priceKey] == nil {
			tickGroups[priceKey] = &models.FootprintBar{
				ID:         uuid.New().String(),
				Symbol:     tick.Symbol,
				Timeframe:  config.TickAggregationMethod,
				PriceLevel: priceLevel,
				TickSize:   config.PriceTickSize,
				StartTime:  tick.Timestamp,
				CreatedAt:  time.Now(),
			}
		}
		
		bar := tickGroups[priceKey]
		fe.addTickToBar(bar, tick)
		
		tickCount++
		
		// Check if we've reached the tick count threshold
		if tickCount >= config.TicksPerRow {
			// Finalize current bars
			for _, bar := range tickGroups {
				bar.EndTime = tick.Timestamp
				bars = append(bars, *bar)
			}
			
			// Reset for next group
			tickGroups = make(map[string]*models.FootprintBar)
			tickCount = 0
		}
	}
	
	// Add remaining bars
	for _, bar := range tickGroups {
		if !bar.StartTime.IsZero() {
			bars = append(bars, *bar)
		}
	}
	
	return bars
}

// addTickToBar adds a tick to a footprint bar
func (fe *FootprintEngine) addTickToBar(bar *models.FootprintBar, tick models.Tick) {
	bar.TotalVolume = bar.TotalVolume.Add(tick.Volume)
	bar.TotalTrades++
	
	if tick.Side == "BUY" {
		bar.BuyVolume = bar.BuyVolume.Add(tick.Volume)
		bar.BuyTrades++
		if tick.Volume.GreaterThan(bar.MaxBuyVolume) {
			bar.MaxBuyVolume = tick.Volume
		}
	} else if tick.Side == "SELL" {
		bar.SellVolume = bar.SellVolume.Add(tick.Volume)
		bar.SellTrades++
		if tick.Volume.GreaterThan(bar.MaxSellVolume) {
			bar.MaxSellVolume = tick.Volume
		}
	}
	
	// Update time range
	if bar.StartTime.IsZero() || tick.Timestamp.Before(bar.StartTime) {
		bar.StartTime = tick.Timestamp
	}
	if bar.EndTime.IsZero() || tick.Timestamp.After(bar.EndTime) {
		bar.EndTime = tick.Timestamp
	}
}

// calculateBarMetrics calculates additional metrics for a bar
func (fe *FootprintEngine) calculateBarMetrics(bar *models.FootprintBar, config models.OrderFlowConfig) {
	// Calculate delta
	bar.Delta = bar.BuyVolume.Sub(bar.SellVolume)
	
	// Calculate volume imbalance ratio
	if bar.SellVolume.IsZero() {
		if bar.BuyVolume.IsZero() {
			bar.VolumeImbalance = decimal.Zero
		} else {
			bar.VolumeImbalance = decimal.NewFromFloat(100) // 100% buy
		}
	} else {
		ratio := bar.BuyVolume.Div(bar.SellVolume)
		bar.VolumeImbalance = ratio.Sub(decimal.NewFromFloat(1)).Mul(decimal.NewFromFloat(100))
	}
}

// identifyPointOfControl identifies the Point of Control (highest volume price level)
func (fe *FootprintEngine) identifyPointOfControl(bars []models.FootprintBar) {
	if len(bars) == 0 {
		return
	}

	// Group bars by time period to find POC for each period
	timeGroups := make(map[time.Time][]int)
	
	for i, bar := range bars {
		timeGroups[bar.StartTime] = append(timeGroups[bar.StartTime], i)
	}
	
	// Find POC for each time period
	for _, indices := range timeGroups {
		maxVolume := decimal.Zero
		pocIndex := -1
		
		for _, i := range indices {
			if bars[i].TotalVolume.GreaterThan(maxVolume) {
				maxVolume = bars[i].TotalVolume
				pocIndex = i
			}
		}
		
		if pocIndex >= 0 {
			bars[pocIndex].IsPointOfControl = true
		}
	}
}

// detectVolumeImbalances detects volume imbalances in bars
func (fe *FootprintEngine) detectVolumeImbalances(bars []models.FootprintBar, config models.OrderFlowConfig) {
	for i := range bars {
		bar := &bars[i]
		
		// Check if volume imbalance exceeds threshold
		absImbalance := bar.VolumeImbalance
		if absImbalance.LessThan(decimal.Zero) {
			absImbalance = absImbalance.Neg()
		}
		
		// Check if total volume meets minimum threshold
		if bar.TotalVolume.GreaterThanOrEqual(config.ImbalanceMinVolume) &&
		   absImbalance.GreaterThanOrEqual(config.ImbalanceThreshold) {
			bar.IsImbalanced = true
		}
	}
}

// Helper methods

func (fe *FootprintEngine) roundToTimePeriod(timestamp time.Time, period time.Duration) time.Time {
	return timestamp.Truncate(period)
}

func (fe *FootprintEngine) roundToTickSize(price, tickSize decimal.Decimal) decimal.Decimal {
	if tickSize.IsZero() {
		return price
	}
	
	// Round to nearest tick size
	divided := price.Div(tickSize)
	rounded := divided.Round(0)
	return rounded.Mul(tickSize)
}

// GetFootprintSummary returns a summary of footprint data
func (fe *FootprintEngine) GetFootprintSummary(bars []models.FootprintBar) map[string]interface{} {
	if len(bars) == 0 {
		return map[string]interface{}{
			"total_bars": 0,
			"total_volume": "0",
			"total_delta": "0",
		}
	}

	totalVolume := decimal.Zero
	totalDelta := decimal.Zero
	imbalancedBars := 0
	pocBars := 0
	
	for _, bar := range bars {
		totalVolume = totalVolume.Add(bar.TotalVolume)
		totalDelta = totalDelta.Add(bar.Delta)
		
		if bar.IsImbalanced {
			imbalancedBars++
		}
		if bar.IsPointOfControl {
			pocBars++
		}
	}
	
	return map[string]interface{}{
		"total_bars":      len(bars),
		"total_volume":    totalVolume.String(),
		"total_delta":     totalDelta.String(),
		"imbalanced_bars": imbalancedBars,
		"poc_bars":        pocBars,
		"imbalance_rate":  float64(imbalancedBars) / float64(len(bars)) * 100,
	}
}
