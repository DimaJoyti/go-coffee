package aggregators

import (
	"fmt"
	"sort"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// RoutingEngine handles intelligent routing decisions across DEX aggregators
type RoutingEngine struct {
	logger *logger.Logger
	config RoutingConfig
}

// NewRoutingEngine creates a new routing engine
func NewRoutingEngine(logger *logger.Logger, config RoutingConfig) *RoutingEngine {
	return &RoutingEngine{
		logger: logger.Named("routing-engine"),
		config: config,
	}
}

// SelectBestQuote selects the best quote based on the configured strategy
func (re *RoutingEngine) SelectBestQuote(quotes []*SwapQuote, req *QuoteRequest) *SwapQuote {
	if len(quotes) == 0 {
		return nil
	}

	if len(quotes) == 1 {
		return quotes[0]
	}

	re.logger.Info("Selecting best quote",
		zap.String("strategy", string(re.config.Strategy)),
		zap.Int("quote_count", len(quotes)))

	switch re.config.Strategy {
	case RoutingStrategyBestPrice:
		return re.selectBestPriceQuote(quotes)
	case RoutingStrategyLowestGas:
		return re.selectLowestGasQuote(quotes)
	case RoutingStrategyBestValue:
		return re.selectBestValueQuote(quotes)
	case RoutingStrategyFastest:
		return re.selectFastestQuote(quotes)
	case RoutingStrategyMostLiquid:
		return re.selectMostLiquidQuote(quotes)
	case RoutingStrategyBalanced:
		return re.selectBalancedQuote(quotes)
	default:
		return re.selectBestPriceQuote(quotes)
	}
}

// selectBestPriceQuote selects the quote with the highest output amount
func (re *RoutingEngine) selectBestPriceQuote(quotes []*SwapQuote) *SwapQuote {
	bestQuote := quotes[0]
	for _, quote := range quotes[1:] {
		if quote.AmountOut.GreaterThan(bestQuote.AmountOut) {
			bestQuote = quote
		}
	}
	return bestQuote
}

// selectLowestGasQuote selects the quote with the lowest gas cost
func (re *RoutingEngine) selectLowestGasQuote(quotes []*SwapQuote) *SwapQuote {
	bestQuote := quotes[0]
	for _, quote := range quotes[1:] {
		if quote.GasCost.LessThan(bestQuote.GasCost) {
			bestQuote = quote
		}
	}
	return bestQuote
}

// selectBestValueQuote selects the quote with the best value (output - gas cost)
func (re *RoutingEngine) selectBestValueQuote(quotes []*SwapQuote) *SwapQuote {
	bestQuote := quotes[0]
	bestValue := bestQuote.AmountOut.Sub(bestQuote.GasCost)

	for _, quote := range quotes[1:] {
		value := quote.AmountOut.Sub(quote.GasCost)
		if value.GreaterThan(bestValue) {
			bestQuote = quote
			bestValue = value
		}
	}
	return bestQuote
}

// selectFastestQuote selects the quote from the fastest aggregator
func (re *RoutingEngine) selectFastestQuote(quotes []*SwapQuote) *SwapQuote {
	// Simplified: assume 1inch is fastest, then 0x, then Paraswap, then Matcha
	aggregatorPriority := map[string]int{
		"1inch":    1,
		"0x":       2,
		"paraswap": 3,
		"matcha":   4,
	}

	bestQuote := quotes[0]
	bestPriority := aggregatorPriority[bestQuote.Aggregator]

	for _, quote := range quotes[1:] {
		priority := aggregatorPriority[quote.Aggregator]
		if priority < bestPriority {
			bestQuote = quote
			bestPriority = priority
		}
	}
	return bestQuote
}

// selectMostLiquidQuote selects the quote with the lowest price impact
func (re *RoutingEngine) selectMostLiquidQuote(quotes []*SwapQuote) *SwapQuote {
	bestQuote := quotes[0]
	for _, quote := range quotes[1:] {
		if quote.PriceImpact.LessThan(bestQuote.PriceImpact) {
			bestQuote = quote
		}
	}
	return bestQuote
}

// selectBalancedQuote selects the quote using a balanced scoring algorithm
func (re *RoutingEngine) selectBalancedQuote(quotes []*SwapQuote) *SwapQuote {
	type scoredQuote struct {
		quote *SwapQuote
		score decimal.Decimal
	}

	scoredQuotes := make([]scoredQuote, 0, len(quotes))

	// Calculate scores for each quote
	for _, quote := range quotes {
		score := re.calculateBalancedScore(quote, quotes)
		scoredQuotes = append(scoredQuotes, scoredQuote{
			quote: quote,
			score: score,
		})
	}

	// Sort by score (highest first)
	sort.Slice(scoredQuotes, func(i, j int) bool {
		return scoredQuotes[i].score.GreaterThan(scoredQuotes[j].score)
	})

	return scoredQuotes[0].quote
}

// calculateBalancedScore calculates a balanced score for a quote
func (re *RoutingEngine) calculateBalancedScore(quote *SwapQuote, allQuotes []*SwapQuote) decimal.Decimal {
	// Normalize metrics to 0-1 scale
	priceScore := re.normalizePrice(quote, allQuotes)
	gasScore := re.normalizeGas(quote, allQuotes)
	liquidityScore := re.normalizeLiquidity(quote, allQuotes)
	reliabilityScore := re.getReliabilityScore(quote.Aggregator)

	// Weighted average (adjust weights as needed)
	weights := map[string]decimal.Decimal{
		"price":       decimal.NewFromFloat(0.4),
		"gas":         decimal.NewFromFloat(0.2),
		"liquidity":   decimal.NewFromFloat(0.2),
		"reliability": decimal.NewFromFloat(0.2),
	}

	score := priceScore.Mul(weights["price"]).
		Add(gasScore.Mul(weights["gas"])).
		Add(liquidityScore.Mul(weights["liquidity"])).
		Add(reliabilityScore.Mul(weights["reliability"]))

	return score
}

// normalizePrice normalizes price to 0-1 scale (higher is better)
func (re *RoutingEngine) normalizePrice(quote *SwapQuote, allQuotes []*SwapQuote) decimal.Decimal {
	minAmount := allQuotes[0].AmountOut
	maxAmount := allQuotes[0].AmountOut

	for _, q := range allQuotes {
		if q.AmountOut.LessThan(minAmount) {
			minAmount = q.AmountOut
		}
		if q.AmountOut.GreaterThan(maxAmount) {
			maxAmount = q.AmountOut
		}
	}

	if maxAmount.Equal(minAmount) {
		return decimal.NewFromFloat(1.0)
	}

	return quote.AmountOut.Sub(minAmount).Div(maxAmount.Sub(minAmount))
}

// normalizeGas normalizes gas cost to 0-1 scale (lower is better, so we invert)
func (re *RoutingEngine) normalizeGas(quote *SwapQuote, allQuotes []*SwapQuote) decimal.Decimal {
	minGas := allQuotes[0].GasCost
	maxGas := allQuotes[0].GasCost

	for _, q := range allQuotes {
		if q.GasCost.LessThan(minGas) {
			minGas = q.GasCost
		}
		if q.GasCost.GreaterThan(maxGas) {
			maxGas = q.GasCost
		}
	}

	if maxGas.Equal(minGas) {
		return decimal.NewFromFloat(1.0)
	}

	// Invert so lower gas cost gets higher score
	return decimal.NewFromFloat(1.0).Sub(quote.GasCost.Sub(minGas).Div(maxGas.Sub(minGas)))
}

// normalizeLiquidity normalizes liquidity (price impact) to 0-1 scale
func (re *RoutingEngine) normalizeLiquidity(quote *SwapQuote, allQuotes []*SwapQuote) decimal.Decimal {
	minImpact := allQuotes[0].PriceImpact
	maxImpact := allQuotes[0].PriceImpact

	for _, q := range allQuotes {
		if q.PriceImpact.LessThan(minImpact) {
			minImpact = q.PriceImpact
		}
		if q.PriceImpact.GreaterThan(maxImpact) {
			maxImpact = q.PriceImpact
		}
	}

	if maxImpact.Equal(minImpact) {
		return decimal.NewFromFloat(1.0)
	}

	// Invert so lower price impact gets higher score
	return decimal.NewFromFloat(1.0).Sub(quote.PriceImpact.Sub(minImpact).Div(maxImpact.Sub(minImpact)))
}

// getReliabilityScore returns a reliability score for an aggregator
func (re *RoutingEngine) getReliabilityScore(aggregator string) decimal.Decimal {
	// Simplified reliability scores based on track record
	scores := map[string]decimal.Decimal{
		"1inch":    decimal.NewFromFloat(0.95),
		"0x":       decimal.NewFromFloat(0.90),
		"paraswap": decimal.NewFromFloat(0.85),
		"matcha":   decimal.NewFromFloat(0.80),
	}

	if score, exists := scores[aggregator]; exists {
		return score
	}
	return decimal.NewFromFloat(0.5) // Default score
}

// GenerateRecommendation generates a routing recommendation
func (re *RoutingEngine) GenerateRecommendation(quotes []*SwapQuote, bestQuote *SwapQuote, req *QuoteRequest) *RouteRecommendation {
	// Sort quotes by amount out (descending)
	sortedQuotes := make([]*SwapQuote, len(quotes))
	copy(sortedQuotes, quotes)
	sort.Slice(sortedQuotes, func(i, j int) bool {
		return sortedQuotes[i].AmountOut.GreaterThan(sortedQuotes[j].AmountOut)
	})

	// Generate alternatives (top 3 excluding the best)
	alternatives := make([]*SwapQuote, 0, 3)
	for _, quote := range sortedQuotes {
		if quote.ID != bestQuote.ID && len(alternatives) < 3 {
			alternatives = append(alternatives, quote)
		}
	}

	// Calculate confidence based on price difference
	confidence := re.calculateConfidence(bestQuote, quotes)

	// Determine risk level
	riskLevel := re.assessRiskLevel(bestQuote)

	// Generate reason
	reason := re.generateReason(bestQuote, re.config.Strategy)

	return &RouteRecommendation{
		Strategy:     re.config.Strategy,
		Reason:       reason,
		Confidence:   confidence,
		Alternatives: alternatives,
		RiskLevel:    riskLevel,
		Metadata: map[string]interface{}{
			"quote_count":    len(quotes),
			"price_spread":   re.calculatePriceSpread(quotes),
			"gas_cost_range": re.calculateGasCostRange(quotes),
		},
	}
}

// calculateConfidence calculates confidence in the recommendation
func (re *RoutingEngine) calculateConfidence(bestQuote *SwapQuote, allQuotes []*SwapQuote) decimal.Decimal {
	if len(allQuotes) <= 1 {
		return decimal.NewFromFloat(0.5)
	}

	// Find second best quote
	var secondBest *SwapQuote
	for _, quote := range allQuotes {
		if quote.ID != bestQuote.ID {
			if secondBest == nil || quote.AmountOut.GreaterThan(secondBest.AmountOut) {
				secondBest = quote
			}
		}
	}

	if secondBest == nil {
		return decimal.NewFromFloat(0.5)
	}

	// Calculate percentage difference
	diff := bestQuote.AmountOut.Sub(secondBest.AmountOut).Div(secondBest.AmountOut)

	// Higher difference = higher confidence (capped at 0.95)
	confidence := decimal.NewFromFloat(0.5).Add(diff.Mul(decimal.NewFromFloat(10)))
	if confidence.GreaterThan(decimal.NewFromFloat(0.95)) {
		confidence = decimal.NewFromFloat(0.95)
	}
	if confidence.LessThan(decimal.NewFromFloat(0.1)) {
		confidence = decimal.NewFromFloat(0.1)
	}

	return confidence
}

// assessRiskLevel assesses the risk level of a quote
func (re *RoutingEngine) assessRiskLevel(quote *SwapQuote) string {
	// Assess based on price impact and aggregator reliability
	if quote.PriceImpact.GreaterThan(decimal.NewFromFloat(0.05)) {
		return "high"
	}
	if quote.PriceImpact.GreaterThan(decimal.NewFromFloat(0.02)) {
		return "medium"
	}
	return "low"
}

// generateReason generates a human-readable reason for the recommendation
func (re *RoutingEngine) generateReason(quote *SwapQuote, strategy RoutingStrategy) string {
	switch strategy {
	case RoutingStrategyBestPrice:
		return fmt.Sprintf("Selected %s for best output amount: %s %s",
			quote.Aggregator, quote.AmountOut.String(), quote.TokenOut.Symbol)
	case RoutingStrategyLowestGas:
		return fmt.Sprintf("Selected %s for lowest gas cost: %s ETH",
			quote.Aggregator, quote.GasCost.String())
	case RoutingStrategyBestValue:
		return fmt.Sprintf("Selected %s for best overall value (output minus gas)",
			quote.Aggregator)
	case RoutingStrategyFastest:
		return fmt.Sprintf("Selected %s for fastest execution",
			quote.Aggregator)
	case RoutingStrategyMostLiquid:
		return fmt.Sprintf("Selected %s for lowest price impact: %s%%",
			quote.Aggregator, quote.PriceImpact.Mul(decimal.NewFromInt(100)).String())
	case RoutingStrategyBalanced:
		return fmt.Sprintf("Selected %s based on balanced scoring across all factors",
			quote.Aggregator)
	default:
		return fmt.Sprintf("Selected %s", quote.Aggregator)
	}
}

// calculatePriceSpread calculates the price spread across all quotes
func (re *RoutingEngine) calculatePriceSpread(quotes []*SwapQuote) decimal.Decimal {
	if len(quotes) <= 1 {
		return decimal.Zero
	}

	minAmount := quotes[0].AmountOut
	maxAmount := quotes[0].AmountOut

	for _, quote := range quotes {
		if quote.AmountOut.LessThan(minAmount) {
			minAmount = quote.AmountOut
		}
		if quote.AmountOut.GreaterThan(maxAmount) {
			maxAmount = quote.AmountOut
		}
	}

	if minAmount.IsZero() {
		return decimal.Zero
	}

	return maxAmount.Sub(minAmount).Div(minAmount)
}

// calculateGasCostRange calculates the gas cost range across all quotes
func (re *RoutingEngine) calculateGasCostRange(quotes []*SwapQuote) map[string]decimal.Decimal {
	if len(quotes) == 0 {
		return map[string]decimal.Decimal{
			"min": decimal.Zero,
			"max": decimal.Zero,
		}
	}

	minGas := quotes[0].GasCost
	maxGas := quotes[0].GasCost

	for _, quote := range quotes {
		if quote.GasCost.LessThan(minGas) {
			minGas = quote.GasCost
		}
		if quote.GasCost.GreaterThan(maxGas) {
			maxGas = quote.GasCost
		}
	}

	return map[string]decimal.Decimal{
		"min": minGas,
		"max": maxGas,
	}
}
