package exchanges

import (
	"context"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

// PriceAggregatorImpl implements the PriceAggregator interface
type PriceAggregatorImpl struct{}

// NewPriceAggregator creates a new price aggregator
func NewPriceAggregator() PriceAggregator {
	return &PriceAggregatorImpl{}
}

// CalculateWeightedPrice calculates volume-weighted average price
func (pa *PriceAggregatorImpl) CalculateWeightedPrice(tickers []*Ticker) decimal.Decimal {
	if len(tickers) == 0 {
		return decimal.Zero
	}
	
	totalWeightedPrice := decimal.Zero
	totalVolume := decimal.Zero
	
	for _, ticker := range tickers {
		if ticker.Volume24h.IsZero() {
			continue
		}
		
		weightedPrice := ticker.LastPrice.Mul(ticker.Volume24h)
		totalWeightedPrice = totalWeightedPrice.Add(weightedPrice)
		totalVolume = totalVolume.Add(ticker.Volume24h)
	}
	
	if totalVolume.IsZero() {
		// Fallback to simple average if no volume data
		return pa.calculateSimpleAverage(tickers)
	}
	
	return totalWeightedPrice.Div(totalVolume)
}

// CalculateVWAP calculates Volume Weighted Average Price
func (pa *PriceAggregatorImpl) CalculateVWAP(tickers []*Ticker) decimal.Decimal {
	return pa.CalculateWeightedPrice(tickers) // Same as weighted price
}

// CalculateMedianPrice calculates median price across exchanges
func (pa *PriceAggregatorImpl) CalculateMedianPrice(tickers []*Ticker) decimal.Decimal {
	if len(tickers) == 0 {
		return decimal.Zero
	}
	
	prices := make([]decimal.Decimal, len(tickers))
	for i, ticker := range tickers {
		prices[i] = ticker.LastPrice
	}
	
	// Sort prices
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].LessThan(prices[j])
	})
	
	n := len(prices)
	if n%2 == 0 {
		// Even number of prices - average of middle two
		mid1 := prices[n/2-1]
		mid2 := prices[n/2]
		return mid1.Add(mid2).Div(decimal.NewFromInt(2))
	} else {
		// Odd number of prices - middle value
		return prices[n/2]
	}
}

// CalculateBestBidAsk finds the best bid and ask prices across exchanges
func (pa *PriceAggregatorImpl) CalculateBestBidAsk(tickers []*Ticker) (*ExchangePrice, *ExchangePrice) {
	if len(tickers) == 0 {
		return nil, nil
	}
	
	var bestBid *ExchangePrice
	var bestAsk *ExchangePrice
	
	for _, ticker := range tickers {
		// Find highest bid
		if bestBid == nil || ticker.BidPrice.GreaterThan(bestBid.Price) {
			bestBid = &ExchangePrice{
				Exchange:  ticker.Exchange,
				Price:     ticker.BidPrice,
				Volume:    ticker.Volume24h,
				Timestamp: ticker.Timestamp,
			}
		}
		
		// Find lowest ask
		if bestAsk == nil || ticker.AskPrice.LessThan(bestAsk.Price) {
			bestAsk = &ExchangePrice{
				Exchange:  ticker.Exchange,
				Price:     ticker.AskPrice,
				Volume:    ticker.Volume24h,
				Timestamp: ticker.Timestamp,
			}
		}
	}
	
	return bestBid, bestAsk
}

// CalculateSpread calculates the spread between bid and ask
func (pa *PriceAggregatorImpl) CalculateSpread(bid, ask decimal.Decimal) decimal.Decimal {
	if bid.IsZero() || ask.IsZero() {
		return decimal.Zero
	}
	return ask.Sub(bid)
}

// CalculateSpreadPercent calculates the spread as a percentage
func (pa *PriceAggregatorImpl) CalculateSpreadPercent(bid, ask decimal.Decimal) decimal.Decimal {
	if bid.IsZero() || ask.IsZero() {
		return decimal.Zero
	}
	
	spread := ask.Sub(bid)
	midPrice := bid.Add(ask).Div(decimal.NewFromInt(2))
	
	if midPrice.IsZero() {
		return decimal.Zero
	}
	
	return spread.Div(midPrice).Mul(decimal.NewFromInt(100))
}

// CalculateTotalVolume calculates total volume across all exchanges
func (pa *PriceAggregatorImpl) CalculateTotalVolume(tickers []*Ticker) decimal.Decimal {
	totalVolume := decimal.Zero
	
	for _, ticker := range tickers {
		totalVolume = totalVolume.Add(ticker.Volume24h)
	}
	
	return totalVolume
}

// CalculateVolumeWeights calculates volume weights for each exchange
func (pa *PriceAggregatorImpl) CalculateVolumeWeights(tickers []*Ticker) map[ExchangeType]decimal.Decimal {
	weights := make(map[ExchangeType]decimal.Decimal)
	totalVolume := pa.CalculateTotalVolume(tickers)
	
	if totalVolume.IsZero() {
		// Equal weights if no volume data
		equalWeight := decimal.NewFromInt(1).Div(decimal.NewFromInt(int64(len(tickers))))
		for _, ticker := range tickers {
			weights[ticker.Exchange] = equalWeight
		}
		return weights
	}
	
	for _, ticker := range tickers {
		weight := ticker.Volume24h.Div(totalVolume)
		weights[ticker.Exchange] = weight
	}
	
	return weights
}

// calculateSimpleAverage calculates simple arithmetic average
func (pa *PriceAggregatorImpl) calculateSimpleAverage(tickers []*Ticker) decimal.Decimal {
	if len(tickers) == 0 {
		return decimal.Zero
	}
	
	total := decimal.Zero
	for _, ticker := range tickers {
		total = total.Add(ticker.LastPrice)
	}
	
	return total.Div(decimal.NewFromInt(int64(len(tickers))))
}

// ArbitrageDetectorImpl implements the ArbitrageDetector interface
type ArbitrageDetectorImpl struct {
	minProfitThreshold decimal.Decimal
}

// NewArbitrageDetector creates a new arbitrage detector
func NewArbitrageDetector(minProfitThreshold float64) ArbitrageDetector {
	return &ArbitrageDetectorImpl{
		minProfitThreshold: decimal.NewFromFloat(minProfitThreshold),
	}
}

// DetectOpportunities detects arbitrage opportunities between exchanges
func (ad *ArbitrageDetectorImpl) DetectOpportunities(ctx context.Context, tickers []*Ticker) ([]*ArbitrageOpportunity, error) {
	if len(tickers) < 2 {
		return nil, nil
	}
	
	var opportunities []*ArbitrageOpportunity
	
	// Compare each pair of exchanges
	for i := 0; i < len(tickers); i++ {
		for j := i + 1; j < len(tickers); j++ {
			ticker1 := tickers[i]
			ticker2 := tickers[j]
			
			// Check opportunity: buy from ticker1, sell to ticker2
			if ticker1.AskPrice.LessThan(ticker2.BidPrice) {
				profit := ad.CalculateProfitability(ticker1.AskPrice, ticker2.BidPrice, ticker1.Volume24h)
				if profit.GreaterThan(ad.minProfitThreshold) {
					opportunity := &ArbitrageOpportunity{
						Symbol:          ticker1.Symbol,
						BuyExchange:     ticker1.Exchange,
						SellExchange:    ticker2.Exchange,
						BuyPrice:        ticker1.AskPrice,
						SellPrice:       ticker2.BidPrice,
						PriceDifference: ticker2.BidPrice.Sub(ticker1.AskPrice),
						ProfitPercent:   profit,
						Volume:          minDecimal(ticker1.Volume24h, ticker2.Volume24h),
						Timestamp:       ticker1.Timestamp,
						Confidence:      ad.CalculateConfidence(&ArbitrageOpportunity{
							BuyPrice:  ticker1.AskPrice,
							SellPrice: ticker2.BidPrice,
							Volume:    minDecimal(ticker1.Volume24h, ticker2.Volume24h),
						}),
					}
					opportunities = append(opportunities, opportunity)
				}
			}
			
			// Check opportunity: buy from ticker2, sell to ticker1
			if ticker2.AskPrice.LessThan(ticker1.BidPrice) {
				profit := ad.CalculateProfitability(ticker2.AskPrice, ticker1.BidPrice, ticker2.Volume24h)
				if profit.GreaterThan(ad.minProfitThreshold) {
					opportunity := &ArbitrageOpportunity{
						Symbol:          ticker2.Symbol,
						BuyExchange:     ticker2.Exchange,
						SellExchange:    ticker1.Exchange,
						BuyPrice:        ticker2.AskPrice,
						SellPrice:       ticker1.BidPrice,
						PriceDifference: ticker1.BidPrice.Sub(ticker2.AskPrice),
						ProfitPercent:   profit,
						Volume:          minDecimal(ticker1.Volume24h, ticker2.Volume24h),
						Timestamp:       ticker2.Timestamp,
						Confidence:      ad.CalculateConfidence(&ArbitrageOpportunity{
							BuyPrice:  ticker2.AskPrice,
							SellPrice: ticker1.BidPrice,
							Volume:    minDecimal(ticker1.Volume24h, ticker2.Volume24h),
						}),
					}
					opportunities = append(opportunities, opportunity)
				}
			}
		}
	}
	
	// Sort by profit percentage
	sort.Slice(opportunities, func(i, j int) bool {
		return opportunities[i].ProfitPercent.GreaterThan(opportunities[j].ProfitPercent)
	})
	
	return opportunities, nil
}

// CalculateProfitability calculates the profitability percentage
func (ad *ArbitrageDetectorImpl) CalculateProfitability(buyPrice, sellPrice, volume decimal.Decimal) decimal.Decimal {
	if buyPrice.IsZero() {
		return decimal.Zero
	}
	
	profit := sellPrice.Sub(buyPrice)
	return profit.Div(buyPrice).Mul(decimal.NewFromInt(100))
}

// EstimateExecutionCost estimates the cost of executing the arbitrage
func (ad *ArbitrageDetectorImpl) EstimateExecutionCost(exchange ExchangeType, volume decimal.Decimal) decimal.Decimal {
	// Simple fee estimation - in reality this would be more complex
	feeRate := decimal.NewFromFloat(0.001) // 0.1% fee
	return volume.Mul(feeRate)
}

// AssessRisk assesses the risk level of an arbitrage opportunity
func (ad *ArbitrageDetectorImpl) AssessRisk(opportunity *ArbitrageOpportunity) float64 {
	// Simple risk assessment based on profit margin and volume
	profitFloat, _ := opportunity.ProfitPercent.Float64()
	volumeFloat, _ := opportunity.Volume.Float64()
	
	// Higher profit = lower risk, higher volume = lower risk
	riskScore := 1.0 - (profitFloat/10.0 + volumeFloat/1000000.0)
	
	if riskScore < 0 {
		riskScore = 0
	}
	if riskScore > 1 {
		riskScore = 1
	}
	
	return riskScore
}

// CalculateConfidence calculates confidence score for an opportunity
func (ad *ArbitrageDetectorImpl) CalculateConfidence(opportunity *ArbitrageOpportunity) float64 {
	// Simple confidence calculation
	profitFloat, _ := opportunity.ProfitPercent.Float64()
	volumeFloat, _ := opportunity.Volume.Float64()
	
	// Higher profit and volume = higher confidence
	confidence := (profitFloat/5.0 + volumeFloat/500000.0) / 2.0
	
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}
	
	return confidence
}

// ValidateOpportunity validates if an arbitrage opportunity is still valid
func (ad *ArbitrageDetectorImpl) ValidateOpportunity(ctx context.Context, opportunity *ArbitrageOpportunity) bool {
	// In a real implementation, this would check current prices
	// For now, just check if the opportunity is recent
	return opportunity.Timestamp.After(opportunity.Timestamp.Add(-time.Minute))
}

// Helper functions

// minDecimal returns the smaller of two decimal values
func minDecimal(a, b decimal.Decimal) decimal.Decimal {
	if a.LessThan(b) {
		return a
	}
	return b
}
