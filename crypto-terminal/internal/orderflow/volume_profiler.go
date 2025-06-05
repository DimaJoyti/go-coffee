package orderflow

import (
	"fmt"
	"sort"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// VolumeProfiler generates volume profiles from tick data
type VolumeProfiler struct {
	config *config.Config
}

// NewVolumeProfiler creates a new volume profiler
func NewVolumeProfiler(config *config.Config) *VolumeProfiler {
	return &VolumeProfiler{
		config: config,
	}
}

// GenerateVolumeProfile generates a volume profile from tick data
func (vp *VolumeProfiler) GenerateVolumeProfile(ticks []models.Tick, profileType string, config models.OrderFlowConfig) (*models.VolumeProfile, error) {
	if len(ticks) == 0 {
		return nil, fmt.Errorf("no ticks provided")
	}

	// Sort ticks by timestamp
	sort.Slice(ticks, func(i, j int) bool {
		return ticks[i].Timestamp.Before(ticks[j].Timestamp)
	})

	profile := &models.VolumeProfile{
		ID:          uuid.New().String(),
		Symbol:      ticks[0].Symbol,
		ProfileType: profileType,
		StartTime:   ticks[0].Timestamp,
		EndTime:     ticks[len(ticks)-1].Timestamp,
		CreatedAt:   time.Now(),
	}

	// Calculate price range
	highPrice := decimal.Zero
	lowPrice := decimal.Zero
	
	for i, tick := range ticks {
		if i == 0 {
			highPrice = tick.Price
			lowPrice = tick.Price
		} else {
			if tick.Price.GreaterThan(highPrice) {
				highPrice = tick.Price
			}
			if tick.Price.LessThan(lowPrice) {
				lowPrice = tick.Price
			}
		}
	}
	
	profile.HighPrice = highPrice
	profile.LowPrice = lowPrice

	// Generate price levels
	priceLevels, err := vp.generatePriceLevels(ticks, config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate price levels: %w", err)
	}

	profile.PriceLevels = priceLevels

	// Calculate total volume
	totalVolume := decimal.Zero
	for _, level := range priceLevels {
		totalVolume = totalVolume.Add(level.Volume)
	}
	profile.TotalVolume = totalVolume

	// Calculate Point of Control (POC)
	vp.calculatePointOfControl(profile)

	// Calculate Value Area
	vp.calculateValueArea(profile, config)

	// Identify High/Low Volume Nodes
	vp.identifyVolumeNodes(profile, config)

	return profile, nil
}

// generatePriceLevels generates volume profile price levels
func (vp *VolumeProfiler) generatePriceLevels(ticks []models.Tick, config models.OrderFlowConfig) ([]models.VolumeProfileLevel, error) {
	// Group ticks by price levels
	priceGroups := make(map[string]*models.VolumeProfileLevel)
	
	for _, tick := range ticks {
		// Round price to tick size
		priceLevel := vp.roundToTickSize(tick.Price, config.PriceTickSize)
		priceKey := priceLevel.String()
		
		if priceGroups[priceKey] == nil {
			priceGroups[priceKey] = &models.VolumeProfileLevel{
				Price:      priceLevel,
				Volume:     decimal.Zero,
				BuyVolume:  decimal.Zero,
				SellVolume: decimal.Zero,
				TradeCount: 0,
			}
		}
		
		level := priceGroups[priceKey]
		level.Volume = level.Volume.Add(tick.Volume)
		level.TradeCount++
		
		if tick.Side == "BUY" {
			level.BuyVolume = level.BuyVolume.Add(tick.Volume)
		} else if tick.Side == "SELL" {
			level.SellVolume = level.SellVolume.Add(tick.Volume)
		}
		
		// Calculate delta
		level.Delta = level.BuyVolume.Sub(level.SellVolume)
	}
	
	// Convert map to slice and sort by price
	var levels []models.VolumeProfileLevel
	for _, level := range priceGroups {
		levels = append(levels, *level)
	}
	
	sort.Slice(levels, func(i, j int) bool {
		return levels[i].Price.LessThan(levels[j].Price)
	})
	
	// Calculate percentages
	totalVolume := decimal.Zero
	for _, level := range levels {
		totalVolume = totalVolume.Add(level.Volume)
	}
	
	for i := range levels {
		if totalVolume.GreaterThan(decimal.Zero) {
			levels[i].Percentage = levels[i].Volume.Div(totalVolume).Mul(decimal.NewFromFloat(100))
		}
	}
	
	return levels, nil
}

// calculatePointOfControl finds the price level with highest volume
func (vp *VolumeProfiler) calculatePointOfControl(profile *models.VolumeProfile) {
	if len(profile.PriceLevels) == 0 {
		return
	}

	maxVolume := decimal.Zero
	pocPrice := decimal.Zero
	pocIndex := -1
	
	for i, level := range profile.PriceLevels {
		if level.Volume.GreaterThan(maxVolume) {
			maxVolume = level.Volume
			pocPrice = level.Price
			pocIndex = i
		}
	}
	
	profile.PointOfControl = pocPrice
	
	// Mark the POC level
	if pocIndex >= 0 {
		profile.PriceLevels[pocIndex].IsPOC = true
	}
}

// calculateValueArea calculates the Value Area (70% of volume by default)
func (vp *VolumeProfiler) calculateValueArea(profile *models.VolumeProfile, config models.OrderFlowConfig) {
	if len(profile.PriceLevels) == 0 {
		return
	}

	// Find POC index
	pocIndex := -1
	for i, level := range profile.PriceLevels {
		if level.IsPOC {
			pocIndex = i
			break
		}
	}
	
	if pocIndex == -1 {
		return
	}

	// Calculate target volume (70% by default)
	targetPercentage := config.ValueAreaPercentage
	if targetPercentage.IsZero() {
		targetPercentage = decimal.NewFromFloat(70)
	}
	
	targetVolume := profile.TotalVolume.Mul(targetPercentage).Div(decimal.NewFromFloat(100))
	
	// Start from POC and expand up and down
	currentVolume := profile.PriceLevels[pocIndex].Volume
	profile.ValueAreaVolume = currentVolume
	
	valueAreaHigh := profile.PriceLevels[pocIndex].Price
	valueAreaLow := profile.PriceLevels[pocIndex].Price
	
	upIndex := pocIndex + 1
	downIndex := pocIndex - 1
	
	// Mark POC as part of value area
	profile.PriceLevels[pocIndex].IsValueArea = true
	
	// Expand value area until we reach target volume
	for currentVolume.LessThan(targetVolume) && (upIndex < len(profile.PriceLevels) || downIndex >= 0) {
		var nextUpVolume, nextDownVolume decimal.Decimal
		
		// Check volumes at next levels
		if upIndex < len(profile.PriceLevels) {
			nextUpVolume = profile.PriceLevels[upIndex].Volume
		}
		if downIndex >= 0 {
			nextDownVolume = profile.PriceLevels[downIndex].Volume
		}
		
		// Choose the level with higher volume
		if upIndex < len(profile.PriceLevels) && (downIndex < 0 || nextUpVolume.GreaterThanOrEqual(nextDownVolume)) {
			// Expand up
			currentVolume = currentVolume.Add(nextUpVolume)
			valueAreaHigh = profile.PriceLevels[upIndex].Price
			profile.PriceLevels[upIndex].IsValueArea = true
			upIndex++
		} else if downIndex >= 0 {
			// Expand down
			currentVolume = currentVolume.Add(nextDownVolume)
			valueAreaLow = profile.PriceLevels[downIndex].Price
			profile.PriceLevels[downIndex].IsValueArea = true
			downIndex--
		} else {
			break
		}
	}
	
	profile.ValueAreaHigh = valueAreaHigh
	profile.ValueAreaLow = valueAreaLow
	profile.ValueAreaVolume = currentVolume
}

// identifyVolumeNodes identifies High Volume Nodes (HVN) and Low Volume Nodes (LVN)
func (vp *VolumeProfiler) identifyVolumeNodes(profile *models.VolumeProfile, config models.OrderFlowConfig) {
	if len(profile.PriceLevels) == 0 {
		return
	}

	// Calculate average volume
	totalVolume := decimal.Zero
	for _, level := range profile.PriceLevels {
		totalVolume = totalVolume.Add(level.Volume)
	}
	avgVolume := totalVolume.Div(decimal.NewFromInt(int64(len(profile.PriceLevels))))
	
	// Set thresholds
	hvnThreshold := config.HVNThreshold
	if hvnThreshold.IsZero() {
		hvnThreshold = decimal.NewFromFloat(1.5) // 150% of average
	}
	
	lvnThreshold := config.LVNThreshold
	if lvnThreshold.IsZero() {
		lvnThreshold = decimal.NewFromFloat(0.5) // 50% of average
	}
	
	hvnVolumeThreshold := avgVolume.Mul(hvnThreshold)
	lvnVolumeThreshold := avgVolume.Mul(lvnThreshold)
	
	// Identify nodes
	for i := range profile.PriceLevels {
		level := &profile.PriceLevels[i]
		
		if level.Volume.GreaterThanOrEqual(hvnVolumeThreshold) {
			level.IsHVN = true
		} else if level.Volume.LessThanOrEqual(lvnVolumeThreshold) {
			level.IsLVN = true
		}
	}
}

// GenerateSessionProfile generates a volume profile for a specific trading session
func (vp *VolumeProfiler) GenerateSessionProfile(ticks []models.Tick, sessionType string, config models.OrderFlowConfig) (*models.VolumeProfile, error) {
	// Filter ticks by session time
	sessionTicks := vp.filterTicksBySession(ticks, sessionType)
	
	if len(sessionTicks) == 0 {
		return nil, fmt.Errorf("no ticks found for session %s", sessionType)
	}
	
	return vp.GenerateVolumeProfile(sessionTicks, "VPSV", config)
}

// filterTicksBySession filters ticks by trading session
func (vp *VolumeProfiler) filterTicksBySession(ticks []models.Tick, sessionType string) []models.Tick {
	var sessionTicks []models.Tick
	
	for _, tick := range ticks {
		if vp.isTickInSession(tick, sessionType) {
			sessionTicks = append(sessionTicks, tick)
		}
	}
	
	return sessionTicks
}

// isTickInSession checks if a tick belongs to a specific trading session
func (vp *VolumeProfiler) isTickInSession(tick models.Tick, sessionType string) bool {
	// Convert to UTC for session calculation
	utcTime := tick.Timestamp.UTC()
	hour := utcTime.Hour()
	
	switch sessionType {
	case "ASIAN":
		// Asian session: 00:00 - 09:00 UTC
		return hour >= 0 && hour < 9
	case "LONDON":
		// London session: 08:00 - 17:00 UTC
		return hour >= 8 && hour < 17
	case "NEW_YORK":
		// New York session: 13:00 - 22:00 UTC
		return hour >= 13 && hour < 22
	case "LONDON_NY_OVERLAP":
		// London/NY overlap: 13:00 - 17:00 UTC
		return hour >= 13 && hour < 17
	default:
		return true // Include all ticks if session type is unknown
	}
}

// GetVolumeProfileSummary returns a summary of volume profile data
func (vp *VolumeProfiler) GetVolumeProfileSummary(profile *models.VolumeProfile) map[string]interface{} {
	if profile == nil || len(profile.PriceLevels) == 0 {
		return map[string]interface{}{
			"total_levels": 0,
			"total_volume": "0",
		}
	}

	hvnCount := 0
	lvnCount := 0
	valueAreaLevels := 0
	
	for _, level := range profile.PriceLevels {
		if level.IsHVN {
			hvnCount++
		}
		if level.IsLVN {
			lvnCount++
		}
		if level.IsValueArea {
			valueAreaLevels++
		}
	}
	
	priceRange := profile.HighPrice.Sub(profile.LowPrice)
	valueAreaRange := profile.ValueAreaHigh.Sub(profile.ValueAreaLow)
	valueAreaPercentage := decimal.Zero
	
	if profile.TotalVolume.GreaterThan(decimal.Zero) {
		valueAreaPercentage = profile.ValueAreaVolume.Div(profile.TotalVolume).Mul(decimal.NewFromFloat(100))
	}
	
	return map[string]interface{}{
		"total_levels":          len(profile.PriceLevels),
		"total_volume":          profile.TotalVolume.String(),
		"price_range":           priceRange.String(),
		"point_of_control":      profile.PointOfControl.String(),
		"value_area_high":       profile.ValueAreaHigh.String(),
		"value_area_low":        profile.ValueAreaLow.String(),
		"value_area_range":      valueAreaRange.String(),
		"value_area_percentage": valueAreaPercentage.String(),
		"value_area_levels":     valueAreaLevels,
		"hvn_count":            hvnCount,
		"lvn_count":            lvnCount,
	}
}

// Helper methods

func (vp *VolumeProfiler) roundToTickSize(price, tickSize decimal.Decimal) decimal.Decimal {
	if tickSize.IsZero() {
		return price
	}
	
	// Round to nearest tick size
	divided := price.Div(tickSize)
	rounded := divided.Round(0)
	return rounded.Mul(tickSize)
}

// FindSupportResistanceLevels identifies potential support and resistance levels from volume profile
func (vp *VolumeProfiler) FindSupportResistanceLevels(profile *models.VolumeProfile) ([]decimal.Decimal, []decimal.Decimal) {
	var supportLevels, resistanceLevels []decimal.Decimal
	
	if len(profile.PriceLevels) < 3 {
		return supportLevels, resistanceLevels
	}
	
	// Look for local volume maxima (potential support/resistance)
	for i := 1; i < len(profile.PriceLevels)-1; i++ {
		current := profile.PriceLevels[i]
		prev := profile.PriceLevels[i-1]
		next := profile.PriceLevels[i+1]
		
		// Check if current level has higher volume than neighbors
		if current.Volume.GreaterThan(prev.Volume) && current.Volume.GreaterThan(next.Volume) {
			// This is a local volume maximum
			if current.IsHVN {
				// High volume nodes can act as both support and resistance
				supportLevels = append(supportLevels, current.Price)
				resistanceLevels = append(resistanceLevels, current.Price)
			}
		}
	}
	
	// Add POC as a key level
	if !profile.PointOfControl.IsZero() {
		supportLevels = append(supportLevels, profile.PointOfControl)
		resistanceLevels = append(resistanceLevels, profile.PointOfControl)
	}
	
	// Add Value Area boundaries
	if !profile.ValueAreaHigh.IsZero() {
		resistanceLevels = append(resistanceLevels, profile.ValueAreaHigh)
	}
	if !profile.ValueAreaLow.IsZero() {
		supportLevels = append(supportLevels, profile.ValueAreaLow)
	}
	
	return supportLevels, resistanceLevels
}
