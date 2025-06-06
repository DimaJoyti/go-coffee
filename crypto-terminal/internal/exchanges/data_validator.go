package exchanges

import (
	"math"
	"time"

	"github.com/shopspring/decimal"
)

// DataValidatorImpl implements the DataValidator interface
type DataValidatorImpl struct {
	// Configuration
	maxPriceDeviation   float64
	maxVolumeDeviation  float64
	timestampTolerance  time.Duration
	minPrice           decimal.Decimal
	maxPrice           decimal.Decimal
	minVolume          decimal.Decimal
}

// NewDataValidator creates a new data validator
func NewDataValidator() DataValidator {
	return &DataValidatorImpl{
		maxPriceDeviation:  0.1,  // 10% max deviation
		maxVolumeDeviation: 2.0,  // 200% max deviation
		timestampTolerance: 5 * time.Minute,
		minPrice:          decimal.NewFromFloat(0.000001), // Minimum price
		maxPrice:          decimal.NewFromFloat(1000000),  // Maximum price
		minVolume:         decimal.Zero,
	}
}

// ValidatePrice validates if a price is within acceptable range
func (dv *DataValidatorImpl) ValidatePrice(price decimal.Decimal, symbol string) bool {
	// Check if price is positive
	if price.LessThanOrEqual(decimal.Zero) {
		return false
	}
	
	// Check if price is within reasonable bounds
	if price.LessThan(dv.minPrice) || price.GreaterThan(dv.maxPrice) {
		return false
	}
	
	// Check for obvious errors (e.g., prices that are too high/low for known assets)
	return dv.validatePriceBySymbol(price, symbol)
}

// ValidatePriceRange validates if a price is within tolerance of a reference price
func (dv *DataValidatorImpl) ValidatePriceRange(price decimal.Decimal, reference decimal.Decimal, tolerance float64) bool {
	if reference.IsZero() {
		return dv.ValidatePrice(price, "")
	}
	
	deviation := price.Sub(reference).Abs().Div(reference)
	toleranceDecimal := decimal.NewFromFloat(tolerance)
	
	return deviation.LessThanOrEqual(toleranceDecimal)
}

// ValidateVolume validates if volume data is reasonable
func (dv *DataValidatorImpl) ValidateVolume(volume decimal.Decimal) bool {
	// Volume should be non-negative
	if volume.LessThan(decimal.Zero) {
		return false
	}
	
	// Check for unreasonably high volumes (potential data errors)
	maxVolume := decimal.NewFromFloat(1e12) // 1 trillion
	if volume.GreaterThan(maxVolume) {
		return false
	}
	
	return true
}

// ValidateTimestamp validates if timestamp is within acceptable range
func (dv *DataValidatorImpl) ValidateTimestamp(timestamp time.Time, tolerance time.Duration) bool {
	now := time.Now()
	
	// Check if timestamp is too old
	if now.Sub(timestamp) > tolerance {
		return false
	}
	
	// Check if timestamp is in the future (with small tolerance for clock skew)
	if timestamp.Sub(now) > time.Minute {
		return false
	}
	
	return true
}

// ValidateOrderBook validates order book data
func (dv *DataValidatorImpl) ValidateOrderBook(orderBook *OrderBook) bool {
	if orderBook == nil {
		return false
	}
	
	// Validate timestamp
	if !dv.ValidateTimestamp(orderBook.Timestamp, dv.timestampTolerance) {
		return false
	}
	
	// Validate bids (should be in descending order)
	for i := 0; i < len(orderBook.Bids)-1; i++ {
		if orderBook.Bids[i].Price.LessThan(orderBook.Bids[i+1].Price) {
			return false
		}
		if !dv.ValidatePrice(orderBook.Bids[i].Price, orderBook.Symbol) {
			return false
		}
		if !dv.ValidateVolume(orderBook.Bids[i].Quantity) {
			return false
		}
	}
	
	// Validate asks (should be in ascending order)
	for i := 0; i < len(orderBook.Asks)-1; i++ {
		if orderBook.Asks[i].Price.GreaterThan(orderBook.Asks[i+1].Price) {
			return false
		}
		if !dv.ValidatePrice(orderBook.Asks[i].Price, orderBook.Symbol) {
			return false
		}
		if !dv.ValidateVolume(orderBook.Asks[i].Quantity) {
			return false
		}
	}
	
	// Validate spread (best ask should be higher than best bid)
	if len(orderBook.Bids) > 0 && len(orderBook.Asks) > 0 {
		bestBid := orderBook.Bids[0].Price
		bestAsk := orderBook.Asks[0].Price
		
		if bestAsk.LessThanOrEqual(bestBid) {
			return false
		}
		
		// Check for unreasonable spreads
		spread := bestAsk.Sub(bestBid)
		midPrice := bestBid.Add(bestAsk).Div(decimal.NewFromInt(2))
		spreadPercent := spread.Div(midPrice)
		
		// Spread should not be more than 50%
		if spreadPercent.GreaterThan(decimal.NewFromFloat(0.5)) {
			return false
		}
	}
	
	return true
}

// ValidateTicker validates ticker data
func (dv *DataValidatorImpl) ValidateTicker(ticker *Ticker) bool {
	if ticker == nil {
		return false
	}
	
	// Validate timestamp
	if !dv.ValidateTimestamp(ticker.Timestamp, dv.timestampTolerance) {
		return false
	}
	
	// Validate prices
	if !dv.ValidatePrice(ticker.LastPrice, ticker.Symbol) {
		return false
	}
	if !dv.ValidatePrice(ticker.BidPrice, ticker.Symbol) {
		return false
	}
	if !dv.ValidatePrice(ticker.AskPrice, ticker.Symbol) {
		return false
	}
	
	// Validate price relationships
	if ticker.AskPrice.LessThanOrEqual(ticker.BidPrice) {
		return false
	}
	
	// Last price should be between bid and ask (with some tolerance)
	if ticker.LastPrice.LessThan(ticker.BidPrice.Mul(decimal.NewFromFloat(0.95))) ||
		ticker.LastPrice.GreaterThan(ticker.AskPrice.Mul(decimal.NewFromFloat(1.05))) {
		return false
	}
	
	// Validate volumes
	if !dv.ValidateVolume(ticker.Volume24h) {
		return false
	}
	if !dv.ValidateVolume(ticker.VolumeQuote24h) {
		return false
	}
	
	// Validate 24h high/low
	if !ticker.High24h.IsZero() && !dv.ValidatePrice(ticker.High24h, ticker.Symbol) {
		return false
	}
	if !ticker.Low24h.IsZero() && !dv.ValidatePrice(ticker.Low24h, ticker.Symbol) {
		return false
	}
	
	// High should be >= Low
	if !ticker.High24h.IsZero() && !ticker.Low24h.IsZero() {
		if ticker.High24h.LessThan(ticker.Low24h) {
			return false
		}
	}
	
	// Current price should be within 24h range (with tolerance)
	if !ticker.High24h.IsZero() && ticker.LastPrice.GreaterThan(ticker.High24h.Mul(decimal.NewFromFloat(1.1))) {
		return false
	}
	if !ticker.Low24h.IsZero() && ticker.LastPrice.LessThan(ticker.Low24h.Mul(decimal.NewFromFloat(0.9))) {
		return false
	}
	
	return true
}

// CalculateDataQuality calculates overall data quality score
func (dv *DataValidatorImpl) CalculateDataQuality(metrics *DataQualityMetrics) float64 {
	if metrics == nil {
		return 0.0
	}
	
	// Weight different factors
	availabilityWeight := 0.4
	latencyWeight := 0.3
	errorRateWeight := 0.2
	freshnessWeight := 0.1
	
	// Normalize availability (already 0-1)
	availabilityScore := metrics.Availability
	
	// Normalize latency (lower is better)
	latencyMs := float64(metrics.Latency.Milliseconds())
	latencyScore := 1.0 - math.Min(latencyMs/5000.0, 1.0) // 5 seconds = 0 score
	
	// Normalize error rate (lower is better)
	errorScore := 1.0 - math.Min(metrics.ErrorRate, 1.0)
	
	// Normalize freshness (more recent is better)
	age := time.Since(metrics.LastUpdate)
	freshnessScore := 1.0 - math.Min(age.Minutes()/60.0, 1.0) // 1 hour = 0 score
	
	// Calculate weighted score
	qualityScore := availabilityWeight*availabilityScore +
		latencyWeight*latencyScore +
		errorRateWeight*errorScore +
		freshnessWeight*freshnessScore
	
	// Ensure score is between 0 and 1
	if qualityScore < 0 {
		qualityScore = 0
	}
	if qualityScore > 1 {
		qualityScore = 1
	}
	
	return qualityScore
}

// UpdateQualityMetrics updates quality metrics for an exchange and symbol
func (dv *DataValidatorImpl) UpdateQualityMetrics(exchange ExchangeType, symbol string, success bool, latency time.Duration) {
	// This would typically update a metrics store
	// For now, this is a placeholder implementation
}

// validatePriceBySymbol validates price based on known symbol characteristics
func (dv *DataValidatorImpl) validatePriceBySymbol(price decimal.Decimal, symbol string) bool {
	// Simple validation based on common symbols
	// In a real implementation, this would use historical data and market knowledge
	
	priceFloat, _ := price.Float64()
	
	switch {
	case containsString([]string{"BTC", "BITCOIN"}, symbol):
		// Bitcoin should be between $1,000 and $500,000
		return priceFloat >= 1000 && priceFloat <= 500000
	case containsString([]string{"ETH", "ETHEREUM"}, symbol):
		// Ethereum should be between $10 and $50,000
		return priceFloat >= 10 && priceFloat <= 50000
	case containsString([]string{"USDT", "USDC", "DAI", "BUSD"}, symbol):
		// Stablecoins should be close to $1
		return priceFloat >= 0.95 && priceFloat <= 1.05
	default:
		// For unknown symbols, use general bounds
		return priceFloat >= 0.000001 && priceFloat <= 1000000
	}
}

// containsString checks if a slice contains a string
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
