package mempool

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// MempoolAnalyzer analyzes mempool data for transaction optimization
type MempoolAnalyzer struct {
	logger *logger.Logger
	config MempoolAnalyzerConfig

	// Data storage
	transactions    map[common.Hash]*MempoolTransaction
	gasTracker      *GasTracker
	congestionModel *CongestionModel

	// Analysis components
	gasPredictor     *GasPredictor
	timeEstimator    *TimeEstimator
	priorityAnalyzer *PriorityAnalyzer

	// State management
	isRunning    bool
	updateTicker *time.Ticker
	stopChan     chan struct{}
	mutex        sync.RWMutex
	dataMutex    sync.RWMutex
}

// MempoolAnalyzerConfig holds configuration for mempool analysis
type MempoolAnalyzerConfig struct {
	Enabled                bool                   `json:"enabled" yaml:"enabled"`
	UpdateInterval         time.Duration          `json:"update_interval" yaml:"update_interval"`
	DataRetentionPeriod    time.Duration          `json:"data_retention_period" yaml:"data_retention_period"`
	MaxTransactions        int                    `json:"max_transactions" yaml:"max_transactions"`
	GasTrackerConfig       GasTrackerConfig       `json:"gas_tracker_config" yaml:"gas_tracker_config"`
	CongestionModelConfig  CongestionModelConfig  `json:"congestion_model_config" yaml:"congestion_model_config"`
	GasPredictorConfig     GasPredictorConfig     `json:"gas_predictor_config" yaml:"gas_predictor_config"`
	TimeEstimatorConfig    TimeEstimatorConfig    `json:"time_estimator_config" yaml:"time_estimator_config"`
	PriorityAnalyzerConfig PriorityAnalyzerConfig `json:"priority_analyzer_config" yaml:"priority_analyzer_config"`
}

// GasTrackerConfig holds gas tracker configuration
type GasTrackerConfig struct {
	Enabled           bool          `json:"enabled" yaml:"enabled"`
	TrackingWindow    time.Duration `json:"tracking_window" yaml:"tracking_window"`
	SampleSize        int           `json:"sample_size" yaml:"sample_size"`
	PercentileTargets []int         `json:"percentile_targets" yaml:"percentile_targets"`
	UpdateFrequency   time.Duration `json:"update_frequency" yaml:"update_frequency"`
}

// CongestionModelConfig holds congestion model configuration
type CongestionModelConfig struct {
	Enabled              bool                       `json:"enabled" yaml:"enabled"`
	AnalysisWindow       time.Duration              `json:"analysis_window" yaml:"analysis_window"`
	CongestionThresholds map[string]decimal.Decimal `json:"congestion_thresholds" yaml:"congestion_thresholds"`
	PredictionHorizon    time.Duration              `json:"prediction_horizon" yaml:"prediction_horizon"`
}

// GasPredictorConfig holds gas predictor configuration
type GasPredictorConfig struct {
	Enabled           bool              `json:"enabled" yaml:"enabled"`
	PredictionMethods []string          `json:"prediction_methods" yaml:"prediction_methods"`
	ConfidenceLevels  []decimal.Decimal `json:"confidence_levels" yaml:"confidence_levels"`
	TimeHorizons      []time.Duration   `json:"time_horizons" yaml:"time_horizons"`
}

// TimeEstimatorConfig holds time estimator configuration
type TimeEstimatorConfig struct {
	Enabled          bool            `json:"enabled" yaml:"enabled"`
	EstimationMethod string          `json:"estimation_method" yaml:"estimation_method"`
	HistoryWindow    time.Duration   `json:"history_window" yaml:"history_window"`
	AccuracyTarget   decimal.Decimal `json:"accuracy_target" yaml:"accuracy_target"`
}

// PriorityAnalyzerConfig holds priority analyzer configuration
type PriorityAnalyzerConfig struct {
	Enabled         bool     `json:"enabled" yaml:"enabled"`
	PriorityFactors []string `json:"priority_factors" yaml:"priority_factors"`
	WeightingMethod string   `json:"weighting_method" yaml:"weighting_method"`
}

// MempoolTransaction represents a transaction in the mempool
type MempoolTransaction struct {
	Hash             common.Hash     `json:"hash"`
	From             common.Address  `json:"from"`
	To               *common.Address `json:"to"`
	Value            decimal.Decimal `json:"value"`
	GasLimit         uint64          `json:"gas_limit"`
	GasPrice         decimal.Decimal `json:"gas_price"`
	GasTipCap        decimal.Decimal `json:"gas_tip_cap"`
	GasFeeCap        decimal.Decimal `json:"gas_fee_cap"`
	Nonce            uint64          `json:"nonce"`
	Data             []byte          `json:"data"`
	Size             int             `json:"size"`
	FirstSeen        time.Time       `json:"first_seen"`
	LastSeen         time.Time       `json:"last_seen"`
	Priority         decimal.Decimal `json:"priority"`
	EstimatedTime    time.Duration   `json:"estimated_time"`
	CongestionLevel  string          `json:"congestion_level"`
	TransactionType  string          `json:"transaction_type"`
	IsReplacement    bool            `json:"is_replacement"`
	ReplacementCount int             `json:"replacement_count"`
}

// GasTracker tracks gas price trends and statistics
type GasTracker struct {
	logger        *logger.Logger
	config        GasTrackerConfig
	gasPrices     []GasPricePoint
	gasStatistics *GasStatistics
	mutex         sync.RWMutex
}

// GasPricePoint represents a gas price data point
type GasPricePoint struct {
	Timestamp   time.Time       `json:"timestamp"`
	GasPrice    decimal.Decimal `json:"gas_price"`
	BaseFee     decimal.Decimal `json:"base_fee"`
	TipCap      decimal.Decimal `json:"tip_cap"`
	BlockNumber uint64          `json:"block_number"`
	Utilization decimal.Decimal `json:"utilization"`
}

// GasStatistics represents gas price statistics
type GasStatistics struct {
	Mean        decimal.Decimal         `json:"mean"`
	Median      decimal.Decimal         `json:"median"`
	StandardDev decimal.Decimal         `json:"standard_dev"`
	Percentiles map[int]decimal.Decimal `json:"percentiles"`
	Trend       string                  `json:"trend"`
	Volatility  decimal.Decimal         `json:"volatility"`
	LastUpdated time.Time               `json:"last_updated"`
}

// CongestionModel models network congestion
type CongestionModel struct {
	logger           *logger.Logger
	config           CongestionModelConfig
	congestionLevel  string
	congestionScore  decimal.Decimal
	pendingTxCount   int
	blockUtilization decimal.Decimal
	avgWaitTime      time.Duration
	mutex            sync.RWMutex
}

// GasPredictor predicts optimal gas prices
type GasPredictor struct {
	logger *logger.Logger
	config GasPredictorConfig
	models map[string]*PredictionModel
	mutex  sync.RWMutex
}

// PredictionModel represents a gas prediction model
type PredictionModel struct {
	Name        string                   `json:"name"`
	Accuracy    decimal.Decimal          `json:"accuracy"`
	Predictions map[string]GasPrediction `json:"predictions"`
	LastUpdate  time.Time                `json:"last_update"`
}

// GasPrediction represents a gas price prediction
type GasPrediction struct {
	TimeHorizon    time.Duration   `json:"time_horizon"`
	PredictedPrice decimal.Decimal `json:"predicted_price"`
	Confidence     decimal.Decimal `json:"confidence"`
	Range          PriceRange      `json:"range"`
}

// PriceRange represents a price range
type PriceRange struct {
	Low  decimal.Decimal `json:"low"`
	High decimal.Decimal `json:"high"`
}

// TimeEstimator estimates transaction confirmation times
type TimeEstimator struct {
	logger           *logger.Logger
	config           TimeEstimatorConfig
	confirmationData []ConfirmationDataPoint
	mutex            sync.RWMutex
}

// ConfirmationDataPoint represents a confirmation time data point
type ConfirmationDataPoint struct {
	GasPrice         decimal.Decimal `json:"gas_price"`
	ConfirmationTime time.Duration   `json:"confirmation_time"`
	BlockNumber      uint64          `json:"block_number"`
	Timestamp        time.Time       `json:"timestamp"`
}

// PriorityAnalyzer analyzes transaction priority
type PriorityAnalyzer struct {
	logger  *logger.Logger
	config  PriorityAnalyzerConfig
	weights map[string]decimal.Decimal
	mutex   sync.RWMutex
}

// MempoolAnalysis represents the result of mempool analysis
type MempoolAnalysis struct {
	Timestamp           time.Time                `json:"timestamp"`
	TotalTransactions   int                      `json:"total_transactions"`
	PendingTransactions int                      `json:"pending_transactions"`
	GasStatistics       *GasStatistics           `json:"gas_statistics"`
	CongestionLevel     string                   `json:"congestion_level"`
	CongestionScore     decimal.Decimal          `json:"congestion_score"`
	GasPredictions      map[string]GasPrediction `json:"gas_predictions"`
	OptimalGasPrice     decimal.Decimal          `json:"optimal_gas_price"`
	EstimatedWaitTime   time.Duration            `json:"estimated_wait_time"`
	Recommendations     []string                 `json:"recommendations"`
	TopTransactions     []*MempoolTransaction    `json:"top_transactions"`
	Metadata            map[string]interface{}   `json:"metadata"`
}

// NewMempoolAnalyzer creates a new mempool analyzer
func NewMempoolAnalyzer(logger *logger.Logger, config MempoolAnalyzerConfig) *MempoolAnalyzer {
	ma := &MempoolAnalyzer{
		logger:       logger.Named("mempool-analyzer"),
		config:       config,
		transactions: make(map[common.Hash]*MempoolTransaction),
		stopChan:     make(chan struct{}),
	}

	// Initialize components
	ma.gasTracker = &GasTracker{
		logger:        logger.Named("gas-tracker"),
		config:        config.GasTrackerConfig,
		gasPrices:     make([]GasPricePoint, 0),
		gasStatistics: &GasStatistics{Percentiles: make(map[int]decimal.Decimal)},
	}

	ma.congestionModel = &CongestionModel{
		logger: logger.Named("congestion-model"),
		config: config.CongestionModelConfig,
	}

	ma.gasPredictor = &GasPredictor{
		logger: logger.Named("gas-predictor"),
		config: config.GasPredictorConfig,
		models: make(map[string]*PredictionModel),
	}

	ma.timeEstimator = &TimeEstimator{
		logger:           logger.Named("time-estimator"),
		config:           config.TimeEstimatorConfig,
		confirmationData: make([]ConfirmationDataPoint, 0),
	}

	ma.priorityAnalyzer = &PriorityAnalyzer{
		logger:  logger.Named("priority-analyzer"),
		config:  config.PriorityAnalyzerConfig,
		weights: make(map[string]decimal.Decimal),
	}

	return ma
}

// Start starts the mempool analyzer
func (ma *MempoolAnalyzer) Start(ctx context.Context) error {
	ma.mutex.Lock()
	defer ma.mutex.Unlock()

	if ma.isRunning {
		return fmt.Errorf("mempool analyzer is already running")
	}

	if !ma.config.Enabled {
		ma.logger.Info("Mempool analyzer is disabled")
		return nil
	}

	ma.logger.Info("Starting mempool analyzer",
		zap.Duration("update_interval", ma.config.UpdateInterval),
		zap.Int("max_transactions", ma.config.MaxTransactions))

	// Initialize priority weights
	ma.initializePriorityWeights()

	// Start monitoring loop
	ma.updateTicker = time.NewTicker(ma.config.UpdateInterval)
	go ma.monitoringLoop(ctx)

	// Start data cleanup routine
	go ma.dataCleanupLoop(ctx)

	ma.isRunning = true
	ma.logger.Info("Mempool analyzer started successfully")
	return nil
}

// Stop stops the mempool analyzer
func (ma *MempoolAnalyzer) Stop() error {
	ma.mutex.Lock()
	defer ma.mutex.Unlock()

	if !ma.isRunning {
		return nil
	}

	ma.logger.Info("Stopping mempool analyzer")

	// Stop monitoring
	if ma.updateTicker != nil {
		ma.updateTicker.Stop()
	}
	close(ma.stopChan)

	ma.isRunning = false
	ma.logger.Info("Mempool analyzer stopped")
	return nil
}

// AddTransaction adds a transaction to the mempool analysis
func (ma *MempoolAnalyzer) AddTransaction(tx *types.Transaction) error {
	ma.dataMutex.Lock()
	defer ma.dataMutex.Unlock()

	// Convert to mempool transaction
	mempoolTx := ma.convertToMempoolTransaction(tx)

	// Check if transaction already exists
	if existingTx, exists := ma.transactions[mempoolTx.Hash]; exists {
		// Update existing transaction
		existingTx.LastSeen = time.Now()
		if mempoolTx.GasPrice.GreaterThan(existingTx.GasPrice) {
			existingTx.IsReplacement = true
			existingTx.ReplacementCount++
			existingTx.GasPrice = mempoolTx.GasPrice
			existingTx.GasTipCap = mempoolTx.GasTipCap
			existingTx.GasFeeCap = mempoolTx.GasFeeCap
		}
	} else {
		// Add new transaction
		ma.transactions[mempoolTx.Hash] = mempoolTx
	}

	// Limit transaction count
	if len(ma.transactions) > ma.config.MaxTransactions {
		ma.removeOldestTransactions()
	}

	return nil
}

// RemoveTransaction removes a transaction from the mempool analysis
func (ma *MempoolAnalyzer) RemoveTransaction(hash common.Hash) {
	ma.dataMutex.Lock()
	defer ma.dataMutex.Unlock()
	delete(ma.transactions, hash)
}

// AnalyzeMempool performs comprehensive mempool analysis
func (ma *MempoolAnalyzer) AnalyzeMempool(ctx context.Context) (*MempoolAnalysis, error) {
	ma.dataMutex.RLock()
	transactions := make([]*MempoolTransaction, 0, len(ma.transactions))
	for _, tx := range ma.transactions {
		transactions = append(transactions, tx)
	}
	ma.dataMutex.RUnlock()

	ma.logger.Debug("Analyzing mempool", zap.Int("transaction_count", len(transactions)))

	// Update gas statistics
	gasStats := ma.gasTracker.CalculateStatistics(transactions)

	// Update congestion model
	congestionLevel, congestionScore := ma.congestionModel.AnalyzeCongestion(transactions)

	// Generate gas predictions
	gasPredictions := ma.gasPredictor.PredictGasPrices(gasStats, congestionScore)

	// Calculate optimal gas price
	optimalGasPrice := ma.calculateOptimalGasPrice(gasStats, gasPredictions)

	// Estimate wait time
	estimatedWaitTime := ma.timeEstimator.EstimateConfirmationTime(optimalGasPrice)

	// Generate recommendations
	recommendations := ma.generateRecommendations(gasStats, congestionLevel, gasPredictions)

	// Get top priority transactions
	topTransactions := ma.getTopPriorityTransactions(transactions, 10)

	analysis := &MempoolAnalysis{
		Timestamp:           time.Now(),
		TotalTransactions:   len(transactions),
		PendingTransactions: len(transactions), // All tracked transactions are pending
		GasStatistics:       gasStats,
		CongestionLevel:     congestionLevel,
		CongestionScore:     congestionScore,
		GasPredictions:      gasPredictions,
		OptimalGasPrice:     optimalGasPrice,
		EstimatedWaitTime:   estimatedWaitTime,
		Recommendations:     recommendations,
		TopTransactions:     topTransactions,
		Metadata:            make(map[string]interface{}),
	}

	ma.logger.Info("Mempool analysis completed",
		zap.Int("total_transactions", analysis.TotalTransactions),
		zap.String("congestion_level", analysis.CongestionLevel),
		zap.String("optimal_gas_price", analysis.OptimalGasPrice.String()),
		zap.Duration("estimated_wait_time", analysis.EstimatedWaitTime))

	return analysis, nil
}

// Helper methods

// convertToMempoolTransaction converts a transaction to mempool transaction
func (ma *MempoolAnalyzer) convertToMempoolTransaction(tx *types.Transaction) *MempoolTransaction {
	now := time.Now()

	mempoolTx := &MempoolTransaction{
		Hash:      tx.Hash(),
		From:      ma.getSenderAddress(tx),
		To:        tx.To(),
		Value:     decimal.NewFromBigInt(tx.Value(), 0),
		GasLimit:  tx.Gas(),
		Nonce:     tx.Nonce(),
		Data:      tx.Data(),
		Size:      int(tx.Size()),
		FirstSeen: now,
		LastSeen:  now,
	}

	// Handle different transaction types
	switch tx.Type() {
	case types.LegacyTxType:
		mempoolTx.GasPrice = decimal.NewFromBigInt(tx.GasPrice(), 0)
		mempoolTx.TransactionType = "legacy"
	case types.DynamicFeeTxType:
		mempoolTx.GasTipCap = decimal.NewFromBigInt(tx.GasTipCap(), 0)
		mempoolTx.GasFeeCap = decimal.NewFromBigInt(tx.GasFeeCap(), 0)
		mempoolTx.TransactionType = "eip1559"
	default:
		mempoolTx.TransactionType = "unknown"
	}

	// Calculate priority
	mempoolTx.Priority = ma.priorityAnalyzer.CalculatePriority(mempoolTx)

	return mempoolTx
}

// getSenderAddress gets the sender address from transaction
func (ma *MempoolAnalyzer) getSenderAddress(tx *types.Transaction) common.Address {
	// In a real implementation, you would recover the sender from the transaction signature
	// For this example, we'll return a zero address
	return common.Address{}
}

// removeOldestTransactions removes oldest transactions to maintain limit
func (ma *MempoolAnalyzer) removeOldestTransactions() {
	// Convert to slice for sorting
	transactions := make([]*MempoolTransaction, 0, len(ma.transactions))
	for _, tx := range ma.transactions {
		transactions = append(transactions, tx)
	}

	// Sort by first seen time
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].FirstSeen.Before(transactions[j].FirstSeen)
	})

	// Remove oldest transactions
	removeCount := len(transactions) - ma.config.MaxTransactions + 100 // Remove extra to avoid frequent cleanup
	for i := 0; i < removeCount && i < len(transactions); i++ {
		delete(ma.transactions, transactions[i].Hash)
	}
}

// calculateOptimalGasPrice calculates optimal gas price
func (ma *MempoolAnalyzer) calculateOptimalGasPrice(gasStats *GasStatistics, predictions map[string]GasPrediction) decimal.Decimal {
	// Use median as base
	optimalPrice := gasStats.Median

	// Adjust based on congestion
	if ma.congestionModel.congestionLevel == "high" {
		optimalPrice = optimalPrice.Mul(decimal.NewFromFloat(1.2))
	} else if ma.congestionModel.congestionLevel == "low" {
		optimalPrice = optimalPrice.Mul(decimal.NewFromFloat(0.9))
	}

	// Consider predictions
	if prediction, exists := predictions["short_term"]; exists {
		if prediction.Confidence.GreaterThan(decimal.NewFromFloat(0.8)) {
			// High confidence prediction, adjust towards it
			optimalPrice = optimalPrice.Add(prediction.PredictedPrice).Div(decimal.NewFromFloat(2))
		}
	}

	return optimalPrice
}

// generateRecommendations generates optimization recommendations
func (ma *MempoolAnalyzer) generateRecommendations(gasStats *GasStatistics, congestionLevel string, predictions map[string]GasPrediction) []string {
	var recommendations []string

	// Congestion-based recommendations
	switch congestionLevel {
	case "low":
		recommendations = append(recommendations, "Network congestion is low - good time for transactions")
		recommendations = append(recommendations, "Consider using lower gas prices for non-urgent transactions")
	case "medium":
		recommendations = append(recommendations, "Moderate network congestion - use standard gas prices")
		recommendations = append(recommendations, "Monitor gas prices for potential increases")
	case "high":
		recommendations = append(recommendations, "High network congestion - expect delays and higher fees")
		recommendations = append(recommendations, "Consider delaying non-urgent transactions")
		recommendations = append(recommendations, "Use higher gas prices for time-sensitive transactions")
	}

	// Gas price trend recommendations
	if gasStats.Trend == "increasing" {
		recommendations = append(recommendations, "Gas prices are trending upward - consider submitting soon")
	} else if gasStats.Trend == "decreasing" {
		recommendations = append(recommendations, "Gas prices are trending downward - consider waiting if not urgent")
	}

	// Volatility recommendations
	if gasStats.Volatility.GreaterThan(decimal.NewFromFloat(0.3)) {
		recommendations = append(recommendations, "High gas price volatility - monitor closely")
	}

	return recommendations
}

// getTopPriorityTransactions returns top priority transactions
func (ma *MempoolAnalyzer) getTopPriorityTransactions(transactions []*MempoolTransaction, limit int) []*MempoolTransaction {
	// Sort by priority
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Priority.GreaterThan(transactions[j].Priority)
	})

	// Return top transactions
	if len(transactions) > limit {
		return transactions[:limit]
	}
	return transactions
}

// initializePriorityWeights initializes priority calculation weights
func (ma *MempoolAnalyzer) initializePriorityWeights() {
	ma.priorityAnalyzer.weights = map[string]decimal.Decimal{
		"gas_price":        decimal.NewFromFloat(0.4),
		"gas_tip":          decimal.NewFromFloat(0.3),
		"transaction_size": decimal.NewFromFloat(0.1),
		"age":              decimal.NewFromFloat(0.1),
		"replacement":      decimal.NewFromFloat(0.1),
	}
}

// monitoringLoop runs the main monitoring loop
func (ma *MempoolAnalyzer) monitoringLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ma.stopChan:
			return
		case <-ma.updateTicker.C:
			ma.performPeriodicAnalysis()
		}
	}
}

// dataCleanupLoop runs the data cleanup loop
func (ma *MempoolAnalyzer) dataCleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(ma.config.DataRetentionPeriod / 10) // Cleanup 10 times per retention period
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ma.stopChan:
			return
		case <-ticker.C:
			ma.cleanupOldData()
		}
	}
}

// performPeriodicAnalysis performs periodic analysis
func (ma *MempoolAnalyzer) performPeriodicAnalysis() {
	ma.logger.Debug("Performing periodic mempool analysis")

	// Update gas tracker
	ma.gasTracker.UpdateStatistics()

	// Update congestion model
	ma.congestionModel.UpdateCongestion()

	// Update gas predictions
	ma.gasPredictor.UpdatePredictions()
}

// cleanupOldData removes old data beyond retention period
func (ma *MempoolAnalyzer) cleanupOldData() {
	ma.dataMutex.Lock()
	defer ma.dataMutex.Unlock()

	cutoff := time.Now().Add(-ma.config.DataRetentionPeriod)

	for hash, tx := range ma.transactions {
		if tx.FirstSeen.Before(cutoff) {
			delete(ma.transactions, hash)
		}
	}

	// Cleanup gas tracker data
	ma.gasTracker.mutex.Lock()
	var validPrices []GasPricePoint
	for _, price := range ma.gasTracker.gasPrices {
		if price.Timestamp.After(cutoff) {
			validPrices = append(validPrices, price)
		}
	}
	ma.gasTracker.gasPrices = validPrices
	ma.gasTracker.mutex.Unlock()

	// Cleanup time estimator data
	ma.timeEstimator.mutex.Lock()
	var validConfirmations []ConfirmationDataPoint
	for _, confirmation := range ma.timeEstimator.confirmationData {
		if confirmation.Timestamp.After(cutoff) {
			validConfirmations = append(validConfirmations, confirmation)
		}
	}
	ma.timeEstimator.confirmationData = validConfirmations
	ma.timeEstimator.mutex.Unlock()
}

// IsRunning returns whether the analyzer is running
func (ma *MempoolAnalyzer) IsRunning() bool {
	ma.mutex.RLock()
	defer ma.mutex.RUnlock()
	return ma.isRunning
}

// GetMetrics returns analyzer metrics
func (ma *MempoolAnalyzer) GetMetrics() map[string]interface{} {
	ma.dataMutex.RLock()
	defer ma.dataMutex.RUnlock()

	return map[string]interface{}{
		"total_transactions":        len(ma.transactions),
		"is_running":                ma.IsRunning(),
		"gas_tracker_enabled":       ma.config.GasTrackerConfig.Enabled,
		"congestion_model_enabled":  ma.config.CongestionModelConfig.Enabled,
		"gas_predictor_enabled":     ma.config.GasPredictorConfig.Enabled,
		"time_estimator_enabled":    ma.config.TimeEstimatorConfig.Enabled,
		"priority_analyzer_enabled": ma.config.PriorityAnalyzerConfig.Enabled,
	}
}

// Component methods

// CalculateStatistics calculates gas price statistics
func (gt *GasTracker) CalculateStatistics(transactions []*MempoolTransaction) *GasStatistics {
	gt.mutex.Lock()
	defer gt.mutex.Unlock()

	if len(transactions) == 0 {
		return gt.gasStatistics
	}

	// Extract gas prices
	var gasPrices []decimal.Decimal
	for _, tx := range transactions {
		if !tx.GasPrice.IsZero() {
			gasPrices = append(gasPrices, tx.GasPrice)
		} else if !tx.GasFeeCap.IsZero() {
			gasPrices = append(gasPrices, tx.GasFeeCap)
		}
	}

	if len(gasPrices) == 0 {
		return gt.gasStatistics
	}

	// Sort prices
	sort.Slice(gasPrices, func(i, j int) bool {
		return gasPrices[i].LessThan(gasPrices[j])
	})

	// Calculate statistics
	stats := &GasStatistics{
		Percentiles: make(map[int]decimal.Decimal),
		LastUpdated: time.Now(),
	}

	// Mean
	total := decimal.Zero
	for _, price := range gasPrices {
		total = total.Add(price)
	}
	stats.Mean = total.Div(decimal.NewFromInt(int64(len(gasPrices))))

	// Median
	mid := len(gasPrices) / 2
	if len(gasPrices)%2 == 0 {
		stats.Median = gasPrices[mid-1].Add(gasPrices[mid]).Div(decimal.NewFromFloat(2))
	} else {
		stats.Median = gasPrices[mid]
	}

	// Percentiles
	for _, p := range gt.config.PercentileTargets {
		index := int(float64(len(gasPrices)) * float64(p) / 100.0)
		if index >= len(gasPrices) {
			index = len(gasPrices) - 1
		}
		stats.Percentiles[p] = gasPrices[index]
	}

	// Standard deviation
	variance := decimal.Zero
	for _, price := range gasPrices {
		diff := price.Sub(stats.Mean)
		variance = variance.Add(diff.Mul(diff))
	}
	variance = variance.Div(decimal.NewFromInt(int64(len(gasPrices))))
	stats.StandardDev = decimal.NewFromFloat(variance.InexactFloat64()).Pow(decimal.NewFromFloat(0.5))

	// Trend analysis (simplified)
	if len(gt.gasPrices) > 1 {
		recent := gt.gasPrices[len(gt.gasPrices)-1].GasPrice
		older := gt.gasPrices[0].GasPrice
		if recent.GreaterThan(older.Mul(decimal.NewFromFloat(1.05))) {
			stats.Trend = "increasing"
		} else if recent.LessThan(older.Mul(decimal.NewFromFloat(0.95))) {
			stats.Trend = "decreasing"
		} else {
			stats.Trend = "stable"
		}
	} else {
		stats.Trend = "stable"
	}

	// Volatility (coefficient of variation)
	if !stats.Mean.IsZero() {
		stats.Volatility = stats.StandardDev.Div(stats.Mean)
	}

	gt.gasStatistics = stats
	return stats
}

// UpdateStatistics updates gas tracker statistics
func (gt *GasTracker) UpdateStatistics() {
	gt.mutex.Lock()
	defer gt.mutex.Unlock()

	// Add current gas price point (mock data)
	now := time.Now()
	gasPricePoint := GasPricePoint{
		Timestamp:   now,
		GasPrice:    decimal.NewFromFloat(20),  // Mock gas price
		BaseFee:     decimal.NewFromFloat(15),  // Mock base fee
		TipCap:      decimal.NewFromFloat(2),   // Mock tip
		BlockNumber: uint64(now.Unix()),        // Mock block number
		Utilization: decimal.NewFromFloat(0.7), // Mock utilization
	}

	gt.gasPrices = append(gt.gasPrices, gasPricePoint)

	// Limit data points
	if len(gt.gasPrices) > gt.config.SampleSize {
		gt.gasPrices = gt.gasPrices[len(gt.gasPrices)-gt.config.SampleSize:]
	}
}

// AnalyzeCongestion analyzes network congestion
func (cm *CongestionModel) AnalyzeCongestion(transactions []*MempoolTransaction) (string, decimal.Decimal) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.pendingTxCount = len(transactions)

	// Calculate congestion score based on transaction count and gas prices
	var totalGasPrice decimal.Decimal
	for _, tx := range transactions {
		if !tx.GasPrice.IsZero() {
			totalGasPrice = totalGasPrice.Add(tx.GasPrice)
		} else if !tx.GasFeeCap.IsZero() {
			totalGasPrice = totalGasPrice.Add(tx.GasFeeCap)
		}
	}

	avgGasPrice := decimal.Zero
	if len(transactions) > 0 {
		avgGasPrice = totalGasPrice.Div(decimal.NewFromInt(int64(len(transactions))))
	}

	// Simple congestion scoring
	score := decimal.Zero
	if len(transactions) > 1000 {
		score = score.Add(decimal.NewFromFloat(0.4))
	} else if len(transactions) > 500 {
		score = score.Add(decimal.NewFromFloat(0.2))
	}

	if avgGasPrice.GreaterThan(decimal.NewFromFloat(50)) {
		score = score.Add(decimal.NewFromFloat(0.4))
	} else if avgGasPrice.GreaterThan(decimal.NewFromFloat(25)) {
		score = score.Add(decimal.NewFromFloat(0.2))
	}

	cm.congestionScore = score

	// Determine congestion level
	if score.GreaterThan(decimal.NewFromFloat(0.7)) {
		cm.congestionLevel = "high"
	} else if score.GreaterThan(decimal.NewFromFloat(0.3)) {
		cm.congestionLevel = "medium"
	} else {
		cm.congestionLevel = "low"
	}

	return cm.congestionLevel, cm.congestionScore
}

// UpdateCongestion updates congestion model
func (cm *CongestionModel) UpdateCongestion() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Mock congestion update
	// In production, this would analyze recent blocks and mempool data
}

// PredictGasPrices predicts gas prices
func (gp *GasPredictor) PredictGasPrices(gasStats *GasStatistics, congestionScore decimal.Decimal) map[string]GasPrediction {
	gp.mutex.Lock()
	defer gp.mutex.Unlock()

	predictions := make(map[string]GasPrediction)

	// Short-term prediction (next 5 minutes)
	shortTermPrice := gasStats.Mean
	if congestionScore.GreaterThan(decimal.NewFromFloat(0.5)) {
		shortTermPrice = shortTermPrice.Mul(decimal.NewFromFloat(1.1))
	}

	predictions["short_term"] = GasPrediction{
		TimeHorizon:    5 * time.Minute,
		PredictedPrice: shortTermPrice,
		Confidence:     decimal.NewFromFloat(0.8),
		Range: PriceRange{
			Low:  shortTermPrice.Mul(decimal.NewFromFloat(0.9)),
			High: shortTermPrice.Mul(decimal.NewFromFloat(1.1)),
		},
	}

	// Medium-term prediction (next 30 minutes)
	mediumTermPrice := gasStats.Median
	if gasStats.Trend == "increasing" {
		mediumTermPrice = mediumTermPrice.Mul(decimal.NewFromFloat(1.05))
	} else if gasStats.Trend == "decreasing" {
		mediumTermPrice = mediumTermPrice.Mul(decimal.NewFromFloat(0.95))
	}

	predictions["medium_term"] = GasPrediction{
		TimeHorizon:    30 * time.Minute,
		PredictedPrice: mediumTermPrice,
		Confidence:     decimal.NewFromFloat(0.6),
		Range: PriceRange{
			Low:  mediumTermPrice.Mul(decimal.NewFromFloat(0.8)),
			High: mediumTermPrice.Mul(decimal.NewFromFloat(1.2)),
		},
	}

	return predictions
}

// UpdatePredictions updates gas predictions
func (gp *GasPredictor) UpdatePredictions() {
	gp.mutex.Lock()
	defer gp.mutex.Unlock()

	// Mock prediction update
	// In production, this would retrain prediction models
}

// EstimateConfirmationTime estimates transaction confirmation time
func (te *TimeEstimator) EstimateConfirmationTime(gasPrice decimal.Decimal) time.Duration {
	te.mutex.RLock()
	defer te.mutex.RUnlock()

	// Simple estimation based on gas price
	if gasPrice.GreaterThan(decimal.NewFromFloat(50)) {
		return 1 * time.Minute
	} else if gasPrice.GreaterThan(decimal.NewFromFloat(25)) {
		return 3 * time.Minute
	} else if gasPrice.GreaterThan(decimal.NewFromFloat(10)) {
		return 5 * time.Minute
	} else {
		return 10 * time.Minute
	}
}

// CalculatePriority calculates transaction priority
func (pa *PriorityAnalyzer) CalculatePriority(tx *MempoolTransaction) decimal.Decimal {
	pa.mutex.RLock()
	defer pa.mutex.RUnlock()

	priority := decimal.Zero

	// Gas price factor
	if !tx.GasPrice.IsZero() {
		gasFactor := tx.GasPrice.Div(decimal.NewFromFloat(100)) // Normalize
		priority = priority.Add(gasFactor.Mul(pa.weights["gas_price"]))
	}

	// Gas tip factor
	if !tx.GasTipCap.IsZero() {
		tipFactor := tx.GasTipCap.Div(decimal.NewFromFloat(10)) // Normalize
		priority = priority.Add(tipFactor.Mul(pa.weights["gas_tip"]))
	}

	// Size factor (smaller transactions have higher priority)
	sizeFactor := decimal.NewFromFloat(1000).Div(decimal.NewFromInt(int64(tx.Size)))
	priority = priority.Add(sizeFactor.Mul(pa.weights["transaction_size"]))

	// Age factor (older transactions have higher priority)
	age := time.Since(tx.FirstSeen)
	ageFactor := decimal.NewFromFloat(age.Minutes())
	priority = priority.Add(ageFactor.Mul(pa.weights["age"]))

	// Replacement factor
	if tx.IsReplacement {
		replacementFactor := decimal.NewFromInt(int64(tx.ReplacementCount))
		priority = priority.Add(replacementFactor.Mul(pa.weights["replacement"]))
	}

	return priority
}
