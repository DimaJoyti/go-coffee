package defi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/redis"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// FrontrunDetector detects frontrunning attacks in the mempool
type FrontrunDetector struct {
	logger *logger.Logger
	cache  redis.Client

	// Detection parameters
	gasPriceThreshold   decimal.Decimal
	timeWindowThreshold time.Duration
	profitThreshold     decimal.Decimal
	similarityThreshold decimal.Decimal

	// State tracking
	pendingTransactions map[string]*FrontrunTransaction
	detectedFrontruns   map[string]*FrontrunPattern
	mutex               sync.RWMutex
	stopChan            chan struct{}
}

// FrontrunTransaction represents a transaction being monitored for frontrunning
type FrontrunTransaction struct {
	Hash            string          `json:"hash"`
	From            string          `json:"from"`
	To              string          `json:"to"`
	Value           *big.Int        `json:"value"`
	GasPrice        *big.Int        `json:"gas_price"`
	GasLimit        uint64          `json:"gas_limit"`
	Data            []byte          `json:"data"`
	MethodSignature string          `json:"method_signature"`
	Timestamp       time.Time       `json:"timestamp"`
	IsHighValue     bool            `json:"is_high_value"`
	IsArbitrage     bool            `json:"is_arbitrage"`
	IsLiquidation   bool            `json:"is_liquidation"`
	EstimatedProfit decimal.Decimal `json:"estimated_profit"`
}

// FrontrunPattern represents a detected frontrunning pattern
type FrontrunPattern struct {
	ID              string               `json:"id"`
	OriginalTx      *FrontrunTransaction `json:"original_tx"`
	FrontrunTx      *FrontrunTransaction `json:"frontrun_tx"`
	Similarity      decimal.Decimal      `json:"similarity"`
	GasPriceDiff    *big.Int             `json:"gas_price_diff"`
	TimeDiff        time.Duration        `json:"time_diff"`
	EstimatedProfit decimal.Decimal      `json:"estimated_profit"`
	Confidence      decimal.Decimal      `json:"confidence"`
	DetectedAt      time.Time            `json:"detected_at"`
}

// NewFrontrunDetector creates a new frontrun detector
func NewFrontrunDetector(logger *logger.Logger, cache redis.Client) *FrontrunDetector {
	return &FrontrunDetector{
		logger:              logger.Named("frontrun-detector"),
		cache:               cache,
		gasPriceThreshold:   decimal.NewFromFloat(1.1), // 10% higher gas price
		timeWindowThreshold: 60 * time.Second,
		profitThreshold:     decimal.NewFromFloat(0.005), // 0.5% minimum profit
		similarityThreshold: decimal.NewFromFloat(0.8),   // 80% similarity
		pendingTransactions: make(map[string]*FrontrunTransaction),
		detectedFrontruns:   make(map[string]*FrontrunPattern),
		stopChan:            make(chan struct{}),
	}
}

// Start starts the frontrun detector
func (fd *FrontrunDetector) Start(ctx context.Context) {
	fd.logger.Info("Starting frontrun detector")

	// Start cleanup routine
	go fd.cleanupLoop(ctx)

	// Start detection routine
	go fd.detectionLoop(ctx)
}

// Stop stops the frontrun detector
func (fd *FrontrunDetector) Stop() {
	fd.logger.Info("Stopping frontrun detector")
	close(fd.stopChan)
}

// DetectFrontrunAttack analyzes a transaction for potential frontrunning
func (fd *FrontrunDetector) DetectFrontrunAttack(ctx context.Context, tx *types.Transaction) *MEVDetection {
	// Parse transaction
	frontrunTx := fd.parseTransaction(tx)

	fd.mutex.Lock()
	fd.pendingTransactions[frontrunTx.Hash] = frontrunTx
	fd.mutex.Unlock()

	// Look for frontrun patterns
	pattern := fd.findFrontrunPattern(frontrunTx)
	if pattern == nil {
		return nil
	}

	// Create MEV detection
	detection := &MEVDetection{
		ID:                fd.generateDetectionID(),
		Type:              MEVAttackFrontrun,
		TargetTransaction: pattern.OriginalTx.Hash,
		AttackerAddress:   pattern.FrontrunTx.From,
		VictimAddress:     pattern.OriginalTx.From,
		TokenAddress:      "", // Will be extracted from transaction data
		EstimatedLoss:     pattern.EstimatedProfit,
		Confidence:        pattern.Confidence,
		BlockNumber:       0, // Will be set when mined
		Timestamp:         time.Now(),
		Prevented:         false,
		PreventionMethod:  "",
	}

	fd.logger.Warn("Frontrun attack detected",
		zap.String("detection_id", detection.ID),
		zap.String("original_tx", pattern.OriginalTx.Hash),
		zap.String("frontrun_tx", pattern.FrontrunTx.Hash),
		zap.String("attacker", pattern.FrontrunTx.From),
		zap.String("confidence", pattern.Confidence.String()))

	return detection
}

// parseTransaction parses a transaction to extract frontrun-relevant information
func (fd *FrontrunDetector) parseTransaction(tx *types.Transaction) *FrontrunTransaction {
	frontrunTx := &FrontrunTransaction{
		Hash:      tx.Hash().Hex(),
		From:      "", // Will be recovered from signature
		To:        "",
		Value:     tx.Value(),
		GasPrice:  tx.GasPrice(),
		GasLimit:  tx.Gas(),
		Data:      tx.Data(),
		Timestamp: time.Now(),
	}

	if tx.To() != nil {
		frontrunTx.To = tx.To().Hex()
	}

	// Extract method signature
	if len(tx.Data()) >= 4 {
		frontrunTx.MethodSignature = hex.EncodeToString(tx.Data()[:4])
	}

	// Classify transaction type
	fd.classifyTransaction(frontrunTx)

	return frontrunTx
}

// classifyTransaction classifies the transaction type
func (fd *FrontrunDetector) classifyTransaction(tx *FrontrunTransaction) {
	// Check if high value transaction
	valueThreshold := big.NewInt(1000000000000000000) // 1 ETH
	tx.IsHighValue = tx.Value.Cmp(valueThreshold) > 0

	// Check for common arbitrage patterns
	tx.IsArbitrage = fd.isArbitrageTransaction(tx)

	// Check for liquidation patterns
	tx.IsLiquidation = fd.isLiquidationTransaction(tx)

	// Estimate potential profit
	tx.EstimatedProfit = fd.estimateTransactionProfit(tx)
}

// isArbitrageTransaction checks if transaction is likely an arbitrage
func (fd *FrontrunDetector) isArbitrageTransaction(tx *FrontrunTransaction) bool {
	// Check for common arbitrage method signatures
	arbitrageSigs := []string{
		"38ed1739", // swapExactTokensForTokens
		"8803dbee", // swapTokensForExactTokens
		"7ff36ab5", // swapExactETHForTokens
		"18cbafe5", // swapExactTokensForETH
		"fb3bdb41", // swapETHForExactTokens
		"4a25d94a", // swapTokensForExactETH
	}

	for _, sig := range arbitrageSigs {
		if tx.MethodSignature == sig {
			return true
		}
	}

	return false
}

// isLiquidationTransaction checks if transaction is likely a liquidation
func (fd *FrontrunDetector) isLiquidationTransaction(tx *FrontrunTransaction) bool {
	// Check for common liquidation method signatures
	liquidationSigs := []string{
		"96cd4ddb", // liquidateBorrow (Compound)
		"00a718a9", // liquidationCall (Aave)
		"f5e3c462", // liquidate (various protocols)
	}

	for _, sig := range liquidationSigs {
		if tx.MethodSignature == sig {
			return true
		}
	}

	return false
}

// estimateTransactionProfit estimates potential profit from a transaction
func (fd *FrontrunDetector) estimateTransactionProfit(tx *FrontrunTransaction) decimal.Decimal {
	// Simplified profit estimation based on transaction type and value
	baseProfit := decimal.NewFromBigInt(tx.Value, -18) // Convert to ETH

	if tx.IsArbitrage {
		// Arbitrage typically has 1-5% profit margins
		return baseProfit.Mul(decimal.NewFromFloat(0.03))
	}

	if tx.IsLiquidation {
		// Liquidations typically have 5-15% profit margins
		return baseProfit.Mul(decimal.NewFromFloat(0.10))
	}

	if tx.IsHighValue {
		// High value transactions might have smaller percentage but higher absolute profit
		return baseProfit.Mul(decimal.NewFromFloat(0.01))
	}

	return decimal.Zero
}

// findFrontrunPattern looks for frontrunning patterns
func (fd *FrontrunDetector) findFrontrunPattern(candidateTx *FrontrunTransaction) *FrontrunPattern {
	fd.mutex.RLock()
	defer fd.mutex.RUnlock()

	// Look for similar transactions with lower gas prices
	for _, originalTx := range fd.pendingTransactions {
		if fd.isPotentialFrontrun(candidateTx, originalTx) {
			pattern := &FrontrunPattern{
				ID:         fd.generatePatternID(),
				OriginalTx: originalTx,
				FrontrunTx: candidateTx,
				DetectedAt: time.Now(),
			}

			// Calculate pattern metrics
			pattern.Similarity = fd.calculateSimilarity(candidateTx, originalTx)
			pattern.GasPriceDiff = new(big.Int).Sub(candidateTx.GasPrice, originalTx.GasPrice)
			pattern.TimeDiff = candidateTx.Timestamp.Sub(originalTx.Timestamp)
			pattern.EstimatedProfit = fd.calculateFrontrunProfit(pattern)
			pattern.Confidence = fd.calculateFrontrunConfidence(pattern)

			// Check if pattern meets thresholds
			if pattern.Confidence.GreaterThan(decimal.NewFromFloat(0.7)) &&
				pattern.Similarity.GreaterThan(fd.similarityThreshold) {
				fd.detectedFrontruns[pattern.ID] = pattern
				return pattern
			}
		}
	}

	return nil
}

// isPotentialFrontrun checks if candidateTx could be frontrunning originalTx
func (fd *FrontrunDetector) isPotentialFrontrun(candidateTx, originalTx *FrontrunTransaction) bool {
	// Skip if same transaction
	if candidateTx.Hash == originalTx.Hash {
		return false
	}

	// Skip if from same address (not frontrunning)
	if candidateTx.From == originalTx.From {
		return false
	}

	// Check if candidate has higher gas price
	gasPriceRatio := decimal.NewFromBigInt(candidateTx.GasPrice, 0).Div(decimal.NewFromBigInt(originalTx.GasPrice, 0))
	if gasPriceRatio.LessThan(fd.gasPriceThreshold) {
		return false
	}

	// Check timing (candidate should be after original)
	if candidateTx.Timestamp.Before(originalTx.Timestamp) {
		return false
	}

	// Check time window
	timeDiff := candidateTx.Timestamp.Sub(originalTx.Timestamp)
	if timeDiff > fd.timeWindowThreshold {
		return false
	}

	// Check if transactions are similar enough
	similarity := fd.calculateSimilarity(candidateTx, originalTx)
	if similarity.LessThan(fd.similarityThreshold) {
		return false
	}

	return true
}

// calculateSimilarity calculates similarity between two transactions
func (fd *FrontrunDetector) calculateSimilarity(tx1, tx2 *FrontrunTransaction) decimal.Decimal {
	similarity := decimal.Zero

	// Method signature similarity (40% weight)
	if tx1.MethodSignature == tx2.MethodSignature {
		similarity = similarity.Add(decimal.NewFromFloat(0.4))
	}

	// Target contract similarity (30% weight)
	if tx1.To == tx2.To {
		similarity = similarity.Add(decimal.NewFromFloat(0.3))
	}

	// Transaction type similarity (20% weight)
	if tx1.IsArbitrage == tx2.IsArbitrage && tx1.IsLiquidation == tx2.IsLiquidation {
		similarity = similarity.Add(decimal.NewFromFloat(0.2))
	}

	// Data similarity (10% weight) - simplified check
	if len(tx1.Data) == len(tx2.Data) {
		similarity = similarity.Add(decimal.NewFromFloat(0.1))
	}

	return similarity
}

// calculateFrontrunProfit calculates estimated profit from frontrunning
func (fd *FrontrunDetector) calculateFrontrunProfit(pattern *FrontrunPattern) decimal.Decimal {
	// Estimate profit based on the original transaction's potential profit
	originalProfit := pattern.OriginalTx.EstimatedProfit

	// Frontrunner typically captures 50-90% of the original profit
	captureRate := decimal.NewFromFloat(0.7)

	return originalProfit.Mul(captureRate)
}

// calculateFrontrunConfidence calculates confidence score for frontrun detection
func (fd *FrontrunDetector) calculateFrontrunConfidence(pattern *FrontrunPattern) decimal.Decimal {
	confidence := decimal.NewFromFloat(0.3) // Base confidence

	// High similarity increases confidence
	confidence = confidence.Add(pattern.Similarity.Mul(decimal.NewFromFloat(0.4)))

	// Significant gas price difference increases confidence
	gasPriceRatio := decimal.NewFromBigInt(pattern.GasPriceDiff, 0).Div(decimal.NewFromBigInt(pattern.OriginalTx.GasPrice, 0))
	if gasPriceRatio.GreaterThan(decimal.NewFromFloat(0.2)) { // 20% higher
		confidence = confidence.Add(decimal.NewFromFloat(0.2))
	}

	// Quick timing increases confidence
	if pattern.TimeDiff < 10*time.Second {
		confidence = confidence.Add(decimal.NewFromFloat(0.1))
	}

	// High value transactions increase confidence
	if pattern.OriginalTx.IsHighValue {
		confidence = confidence.Add(decimal.NewFromFloat(0.1))
	}

	// Profitable transactions increase confidence
	if pattern.EstimatedProfit.GreaterThan(fd.profitThreshold) {
		confidence = confidence.Add(decimal.NewFromFloat(0.1))
	}

	// Cap at 1.0
	if confidence.GreaterThan(decimal.NewFromFloat(1.0)) {
		confidence = decimal.NewFromFloat(1.0)
	}

	return confidence
}

// generateDetectionID generates a unique detection ID
func (fd *FrontrunDetector) generateDetectionID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("frontrun_%s", hex.EncodeToString(bytes))
}

// generatePatternID generates a unique pattern ID
func (fd *FrontrunDetector) generatePatternID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("pattern_%s", hex.EncodeToString(bytes))
}

// cleanupLoop periodically cleans up old transactions
func (fd *FrontrunDetector) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-fd.stopChan:
			return
		case <-ticker.C:
			fd.cleanup()
		}
	}
}

// cleanup removes old transactions from memory
func (fd *FrontrunDetector) cleanup() {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	cutoff := time.Now().Add(-fd.timeWindowThreshold)

	for hash, tx := range fd.pendingTransactions {
		if tx.Timestamp.Before(cutoff) {
			delete(fd.pendingTransactions, hash)
		}
	}

	// Also cleanup old patterns
	for id, pattern := range fd.detectedFrontruns {
		if pattern.DetectedAt.Before(cutoff) {
			delete(fd.detectedFrontruns, id)
		}
	}
}

// detectionLoop runs the main detection logic
func (fd *FrontrunDetector) detectionLoop(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-fd.stopChan:
			return
		case <-ticker.C:
			// Periodic analysis of pending transactions
			fd.analyzePendingTransactions()
		}
	}
}

// analyzePendingTransactions analyzes all pending transactions for patterns
func (fd *FrontrunDetector) analyzePendingTransactions() {
	fd.mutex.RLock()
	transactions := make([]*FrontrunTransaction, 0, len(fd.pendingTransactions))
	for _, tx := range fd.pendingTransactions {
		transactions = append(transactions, tx)
	}
	fd.mutex.RUnlock()

	// Look for new patterns among recent transactions
	for _, tx := range transactions {
		if time.Since(tx.Timestamp) < 30*time.Second {
			fd.findFrontrunPattern(tx)
		}
	}
}

// GetDetectedPatterns returns all detected frontrun patterns
func (fd *FrontrunDetector) GetDetectedPatterns() map[string]*FrontrunPattern {
	fd.mutex.RLock()
	defer fd.mutex.RUnlock()

	patterns := make(map[string]*FrontrunPattern)
	for k, v := range fd.detectedFrontruns {
		patterns[k] = v
	}
	return patterns
}
