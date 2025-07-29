package gas

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// Optimize optimizes gas using EIP-1559 strategy
func (eip *EIP1559Optimizer) Optimize(ctx context.Context, request *OptimizationRequest, metrics *NetworkMetrics) (*OptimizationResult, error) {
	if !eip.config.Enabled {
		return nil, fmt.Errorf("EIP-1559 optimizer is disabled")
	}

	// Calculate base fee multiplier based on priority
	baseFeeMultiplier := eip.config.BaseFeeMultiplier
	switch request.Priority {
	case "low":
		baseFeeMultiplier = baseFeeMultiplier.Mul(decimal.NewFromFloat(0.8))
	case "medium":
		baseFeeMultiplier = baseFeeMultiplier.Mul(decimal.NewFromFloat(1.0))
	case "high":
		baseFeeMultiplier = baseFeeMultiplier.Mul(decimal.NewFromFloat(1.2))
	case "urgent":
		baseFeeMultiplier = baseFeeMultiplier.Mul(decimal.NewFromFloat(1.5))
	}

	// Calculate max fee per gas
	maxFeePerGas := metrics.CurrentBaseFee.Mul(baseFeeMultiplier)

	// Calculate priority fee based on strategy
	var priorityFee decimal.Decimal
	switch eip.config.PriorityFeeStrategy {
	case "fixed":
		priorityFee = decimal.NewFromFloat(2) // 2 gwei
	case "dynamic":
		priorityFee = metrics.RecommendedPriority
	case "aggressive":
		priorityFee = metrics.RecommendedPriority.Mul(decimal.NewFromFloat(1.5))
	default:
		priorityFee = metrics.RecommendedPriority
	}

	// Adjust priority fee based on aggressiveness
	switch eip.config.AggressivenessLevel {
	case "conservative":
		priorityFee = priorityFee.Mul(decimal.NewFromFloat(0.8))
	case "moderate":
		priorityFee = priorityFee.Mul(decimal.NewFromFloat(1.0))
	case "aggressive":
		priorityFee = priorityFee.Mul(decimal.NewFromFloat(1.3))
	}

	// Ensure max fee cap
	maxFeeCap := maxFeePerGas.Mul(eip.config.MaxFeeCapMultiplier)
	if maxFeePerGas.GreaterThan(maxFeeCap) {
		maxFeePerGas = maxFeeCap
	}

	// Calculate estimated cost
	gasLimit := decimal.NewFromInt(int64(request.GasLimit))
	estimatedCost := maxFeePerGas.Mul(gasLimit)

	// Estimate confirmation time based on priority fee
	var estimatedTime time.Duration
	if priorityFee.GreaterThan(decimal.NewFromFloat(5)) {
		estimatedTime = 1 * time.Minute
	} else if priorityFee.GreaterThan(decimal.NewFromFloat(2)) {
		estimatedTime = 2 * time.Minute
	} else {
		estimatedTime = 5 * time.Minute
	}

	result := &OptimizationResult{
		Strategy:             "eip1559",
		GasPrice:             maxFeePerGas, // For legacy compatibility
		MaxFeePerGas:         maxFeePerGas,
		MaxPriorityFeePerGas: priorityFee,
		EstimatedCost:        estimatedCost,
		EstimatedTime:        estimatedTime,
		Confidence:           decimal.NewFromFloat(0.85),
		Reasoning: []string{
			fmt.Sprintf("Base fee: %s gwei", metrics.CurrentBaseFee.String()),
			fmt.Sprintf("Priority fee: %s gwei", priorityFee.String()),
			fmt.Sprintf("Max fee: %s gwei", maxFeePerGas.String()),
			fmt.Sprintf("Strategy: %s", eip.config.PriorityFeeStrategy),
		},
		Timestamp: time.Now(),
	}

	return result, nil
}

// Optimize optimizes gas using historical analysis
func (ha *HistoricalAnalyzer) Optimize(ctx context.Context, request *OptimizationRequest, history []GasDataPoint) (*OptimizationResult, error) {
	if !ha.config.Enabled {
		return nil, fmt.Errorf("historical analyzer is disabled")
	}

	if len(history) == 0 {
		return nil, fmt.Errorf("no historical data available")
	}

	// Filter recent history
	cutoff := time.Now().Add(-ha.config.AnalysisWindow)
	var recentHistory []GasDataPoint
	for _, point := range history {
		if point.Timestamp.After(cutoff) {
			recentHistory = append(recentHistory, point)
		}
	}

	if len(recentHistory) == 0 {
		return nil, fmt.Errorf("no recent historical data available")
	}

	// Limit sample size
	if len(recentHistory) > ha.config.SampleSize {
		recentHistory = recentHistory[len(recentHistory)-ha.config.SampleSize:]
	}

	// Calculate statistics
	var totalGasPrice decimal.Decimal
	var totalConfirmTime time.Duration
	for _, point := range recentHistory {
		totalGasPrice = totalGasPrice.Add(point.GasPrice)
		totalConfirmTime += point.ConfirmationTime
	}

	avgGasPrice := totalGasPrice.Div(decimal.NewFromInt(int64(len(recentHistory))))
	avgConfirmTime := totalConfirmTime / time.Duration(len(recentHistory))

	// Apply weighting strategy
	var recommendedGasPrice decimal.Decimal
	switch ha.config.WeightingStrategy {
	case "simple_average":
		recommendedGasPrice = avgGasPrice
	case "weighted_recent":
		// Give more weight to recent data
		weightedSum := decimal.Zero
		totalWeight := decimal.Zero
		for i, point := range recentHistory {
			weight := decimal.NewFromFloat(float64(i+1)) // Linear weighting
			weightedSum = weightedSum.Add(point.GasPrice.Mul(weight))
			totalWeight = totalWeight.Add(weight)
		}
		recommendedGasPrice = weightedSum.Div(totalWeight)
	default:
		recommendedGasPrice = avgGasPrice
	}

	// Adjust based on priority
	switch request.Priority {
	case "low":
		recommendedGasPrice = recommendedGasPrice.Mul(decimal.NewFromFloat(0.9))
	case "medium":
		recommendedGasPrice = recommendedGasPrice.Mul(decimal.NewFromFloat(1.0))
	case "high":
		recommendedGasPrice = recommendedGasPrice.Mul(decimal.NewFromFloat(1.1))
	case "urgent":
		recommendedGasPrice = recommendedGasPrice.Mul(decimal.NewFromFloat(1.3))
	}

	// Calculate estimated cost
	gasLimit := decimal.NewFromInt(int64(request.GasLimit))
	estimatedCost := recommendedGasPrice.Mul(gasLimit)

	result := &OptimizationResult{
		Strategy:             "historical",
		GasPrice:             recommendedGasPrice,
		MaxFeePerGas:         recommendedGasPrice,
		MaxPriorityFeePerGas: decimal.NewFromFloat(2), // Default priority fee
		EstimatedCost:        estimatedCost,
		EstimatedTime:        avgConfirmTime,
		Confidence:           decimal.NewFromFloat(0.75),
		Reasoning: []string{
			fmt.Sprintf("Analyzed %d historical data points", len(recentHistory)),
			fmt.Sprintf("Average gas price: %s gwei", avgGasPrice.String()),
			fmt.Sprintf("Weighting strategy: %s", ha.config.WeightingStrategy),
			fmt.Sprintf("Priority adjustment: %s", request.Priority),
		},
		Timestamp: time.Now(),
	}

	return result, nil
}

// Optimize optimizes gas based on network congestion
func (cm *CongestionMonitor) Optimize(ctx context.Context, request *OptimizationRequest) (*OptimizationResult, error) {
	if !cm.config.Enabled {
		return nil, fmt.Errorf("congestion monitor is disabled")
	}

	cm.mutex.RLock()
	metrics := *cm.metrics
	cm.mutex.RUnlock()

	// Base gas price from current metrics
	baseGasPrice := metrics.CurrentBaseFee.Add(metrics.RecommendedPriority)

	// Get congestion adjustment factor
	adjustmentFactor := decimal.NewFromFloat(1.0)
	if factor, exists := cm.config.AdjustmentFactors[metrics.CongestionLevel]; exists {
		adjustmentFactor = factor
	}

	// Apply congestion adjustment
	adjustedGasPrice := baseGasPrice.Mul(adjustmentFactor)

	// Further adjust based on priority
	switch request.Priority {
	case "low":
		adjustedGasPrice = adjustedGasPrice.Mul(decimal.NewFromFloat(0.8))
	case "medium":
		adjustedGasPrice = adjustedGasPrice.Mul(decimal.NewFromFloat(1.0))
	case "high":
		adjustedGasPrice = adjustedGasPrice.Mul(decimal.NewFromFloat(1.2))
	case "urgent":
		adjustedGasPrice = adjustedGasPrice.Mul(decimal.NewFromFloat(1.5))
	}

	// Calculate estimated cost
	gasLimit := decimal.NewFromInt(int64(request.GasLimit))
	estimatedCost := adjustedGasPrice.Mul(gasLimit)

	// Estimate confirmation time based on congestion
	var estimatedTime time.Duration
	switch metrics.CongestionLevel {
	case "low":
		estimatedTime = 1 * time.Minute
	case "medium":
		estimatedTime = 3 * time.Minute
	case "high":
		estimatedTime = 8 * time.Minute
	default:
		estimatedTime = 5 * time.Minute
	}

	result := &OptimizationResult{
		Strategy:             "congestion_based",
		GasPrice:             adjustedGasPrice,
		MaxFeePerGas:         adjustedGasPrice,
		MaxPriorityFeePerGas: metrics.RecommendedPriority,
		EstimatedCost:        estimatedCost,
		EstimatedTime:        estimatedTime,
		Confidence:           decimal.NewFromFloat(0.8),
		Reasoning: []string{
			fmt.Sprintf("Network congestion: %s", metrics.CongestionLevel),
			fmt.Sprintf("Adjustment factor: %s", adjustmentFactor.String()),
			fmt.Sprintf("Network utilization: %s%%", metrics.NetworkUtilization.Mul(decimal.NewFromFloat(100)).String()),
			fmt.Sprintf("Pending transactions: %d", metrics.PendingTransactions),
		},
		Timestamp: time.Now(),
	}

	return result, nil
}

// Optimize optimizes gas using prediction models
func (pe *PredictionEngine) Optimize(ctx context.Context, request *OptimizationRequest) (*OptimizationResult, error) {
	if !pe.config.Enabled {
		return nil, fmt.Errorf("prediction engine is disabled")
	}

	pe.mutex.RLock()
	defer pe.mutex.RUnlock()

	// Find best prediction for target time horizon
	var bestPrediction *GasPrediction
	var bestModel string
	targetHorizon := request.TargetConfirmTime
	if targetHorizon == 0 {
		targetHorizon = 5 * time.Minute // Default
	}

	for modelName, model := range pe.models {
		for _, prediction := range model.Predictions {
			if prediction.TimeHorizon <= targetHorizon && 
			   prediction.Confidence.GreaterThan(pe.config.ConfidenceThreshold) {
				if bestPrediction == nil || 
				   prediction.Confidence.GreaterThan(bestPrediction.Confidence) {
					bestPrediction = &prediction
					bestModel = modelName
				}
			}
		}
	}

	if bestPrediction == nil {
		return nil, fmt.Errorf("no suitable predictions available")
	}

	// Use predicted price as base
	recommendedGasPrice := bestPrediction.PredictedPrice

	// Adjust based on priority
	switch request.Priority {
	case "low":
		recommendedGasPrice = recommendedGasPrice.Mul(decimal.NewFromFloat(0.9))
	case "medium":
		recommendedGasPrice = recommendedGasPrice.Mul(decimal.NewFromFloat(1.0))
	case "high":
		recommendedGasPrice = recommendedGasPrice.Mul(decimal.NewFromFloat(1.1))
	case "urgent":
		recommendedGasPrice = recommendedGasPrice.Mul(decimal.NewFromFloat(1.2))
	}

	// Calculate estimated cost
	gasLimit := decimal.NewFromInt(int64(request.GasLimit))
	estimatedCost := recommendedGasPrice.Mul(gasLimit)

	result := &OptimizationResult{
		Strategy:             "prediction_based",
		GasPrice:             recommendedGasPrice,
		MaxFeePerGas:         recommendedGasPrice,
		MaxPriorityFeePerGas: decimal.NewFromFloat(2), // Default priority fee
		EstimatedCost:        estimatedCost,
		EstimatedTime:        bestPrediction.TimeHorizon,
		Confidence:           bestPrediction.Confidence,
		Reasoning: []string{
			fmt.Sprintf("Using prediction model: %s", bestModel),
			fmt.Sprintf("Predicted price: %s gwei", bestPrediction.PredictedPrice.String()),
			fmt.Sprintf("Prediction confidence: %s%%", bestPrediction.Confidence.Mul(decimal.NewFromFloat(100)).String()),
			fmt.Sprintf("Time horizon: %v", bestPrediction.TimeHorizon),
		},
		Timestamp: time.Now(),
	}

	return result, nil
}
