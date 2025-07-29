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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// SandwichDetector detects sandwich attacks in the mempool
type SandwichDetector struct {
	logger *logger.Logger
	cache  redis.Client

	// Detection parameters
	maxSlippageThreshold decimal.Decimal
	minProfitThreshold   decimal.Decimal
	timeWindow           time.Duration

	// State tracking
	pendingTransactions map[string]*PendingTransaction
	detectedSandwiches  map[string]*SandwichPattern
	mutex               sync.RWMutex
	stopChan            chan struct{}
}

// PendingTransaction represents a transaction in the mempool
type PendingTransaction struct {
	Hash      string          `json:"hash"`
	From      string          `json:"from"`
	To        string          `json:"to"`
	Value     *big.Int        `json:"value"`
	GasPrice  *big.Int        `json:"gas_price"`
	GasLimit  uint64          `json:"gas_limit"`
	Data      []byte          `json:"data"`
	TokenIn   string          `json:"token_in,omitempty"`
	TokenOut  string          `json:"token_out,omitempty"`
	AmountIn  decimal.Decimal `json:"amount_in,omitempty"`
	AmountOut decimal.Decimal `json:"amount_out,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	IsSwap    bool            `json:"is_swap"`
	DEX       string          `json:"dex,omitempty"`
}

// SandwichPattern represents a detected sandwich attack pattern
type SandwichPattern struct {
	ID              string              `json:"id"`
	FrontrunTx      *PendingTransaction `json:"frontrun_tx"`
	VictimTx        *PendingTransaction `json:"victim_tx"`
	BackrunTx       *PendingTransaction `json:"backrun_tx"`
	EstimatedProfit decimal.Decimal     `json:"estimated_profit"`
	Confidence      decimal.Decimal     `json:"confidence"`
	DetectedAt      time.Time           `json:"detected_at"`
}

// NewSandwichDetector creates a new sandwich attack detector
func NewSandwichDetector(logger *logger.Logger, cache redis.Client) *SandwichDetector {
	return &SandwichDetector{
		logger:               logger.Named("sandwich-detector"),
		cache:                cache,
		maxSlippageThreshold: decimal.NewFromFloat(0.05), // 5% max slippage
		minProfitThreshold:   decimal.NewFromFloat(0.01), // 1% min profit
		timeWindow:           30 * time.Second,
		pendingTransactions:  make(map[string]*PendingTransaction),
		detectedSandwiches:   make(map[string]*SandwichPattern),
		stopChan:             make(chan struct{}),
	}
}

// Start starts the sandwich detector
func (sd *SandwichDetector) Start(ctx context.Context) {
	sd.logger.Info("Starting sandwich attack detector")

	// Start cleanup routine
	go sd.cleanupLoop(ctx)

	// Start detection routine
	go sd.detectionLoop(ctx)
}

// Stop stops the sandwich detector
func (sd *SandwichDetector) Stop() {
	sd.logger.Info("Stopping sandwich attack detector")
	close(sd.stopChan)
}

// DetectSandwichAttack analyzes a transaction for potential sandwich attacks
func (sd *SandwichDetector) DetectSandwichAttack(ctx context.Context, tx *types.Transaction) *MEVDetection {
	// Parse transaction to extract swap information
	pendingTx := sd.parseTransaction(tx)
	if !pendingTx.IsSwap {
		return nil // Not a swap transaction
	}

	sd.mutex.Lock()
	sd.pendingTransactions[pendingTx.Hash] = pendingTx
	sd.mutex.Unlock()

	// Look for sandwich patterns
	pattern := sd.findSandwichPattern(pendingTx)
	if pattern == nil {
		return nil
	}

	// Create MEV detection
	detection := &MEVDetection{
		ID:                sd.generateDetectionID(),
		Type:              MEVAttackSandwich,
		TargetTransaction: pendingTx.Hash,
		AttackerAddress:   pattern.FrontrunTx.From,
		VictimAddress:     pendingTx.From,
		TokenAddress:      pendingTx.TokenIn,
		EstimatedLoss:     pattern.EstimatedProfit,
		Confidence:        pattern.Confidence,
		BlockNumber:       0, // Will be set when mined
		Timestamp:         time.Now(),
		Prevented:         false,
		PreventionMethod:  "",
	}

	sd.logger.Warn("Sandwich attack detected",
		zap.String("detection_id", detection.ID),
		zap.String("victim_tx", pendingTx.Hash),
		zap.String("attacker", pattern.FrontrunTx.From),
		zap.String("estimated_profit", pattern.EstimatedProfit.String()))

	return detection
}

// parseTransaction parses a transaction to extract relevant information
func (sd *SandwichDetector) parseTransaction(tx *types.Transaction) *PendingTransaction {
	pendingTx := &PendingTransaction{
		Hash:      tx.Hash().Hex(),
		From:      "", // Will be recovered from signature
		To:        tx.To().Hex(),
		Value:     tx.Value(),
		GasPrice:  tx.GasPrice(),
		GasLimit:  tx.Gas(),
		Data:      tx.Data(),
		Timestamp: time.Now(),
		IsSwap:    false,
	}

	// Try to parse as Uniswap V2/V3 swap
	if sd.isUniswapSwap(tx) {
		pendingTx.IsSwap = true
		pendingTx.DEX = "uniswap"
		sd.parseUniswapSwap(tx, pendingTx)
	}

	// Try to parse as other DEX swaps
	if !pendingTx.IsSwap {
		if sd.isSushiSwap(tx) {
			pendingTx.IsSwap = true
			pendingTx.DEX = "sushiswap"
			sd.parseSushiSwap(tx, pendingTx)
		}
	}

	return pendingTx
}

// isUniswapSwap checks if transaction is a Uniswap swap
func (sd *SandwichDetector) isUniswapSwap(tx *types.Transaction) bool {
	if tx.To() == nil {
		return false
	}

	// Check for Uniswap V2 Router
	uniswapV2Router := common.HexToAddress("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
	if *tx.To() == uniswapV2Router {
		return true
	}

	// Check for Uniswap V3 Router
	uniswapV3Router := common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564")
	if *tx.To() == uniswapV3Router {
		return true
	}

	return false
}

// isSushiSwap checks if transaction is a SushiSwap swap
func (sd *SandwichDetector) isSushiSwap(tx *types.Transaction) bool {
	if tx.To() == nil {
		return false
	}

	// Check for SushiSwap Router
	sushiRouter := common.HexToAddress("0xd9e1cE17f2641f24aE83637ab66a2cca9C378B9F")
	return *tx.To() == sushiRouter
}

// parseUniswapSwap parses Uniswap swap transaction
func (sd *SandwichDetector) parseUniswapSwap(tx *types.Transaction, pendingTx *PendingTransaction) {
	// Simplified parsing - in production, use ABI decoding
	data := tx.Data()
	if len(data) < 4 {
		return
	}

	// Extract method signature
	methodSig := hex.EncodeToString(data[:4])

	switch methodSig {
	case "38ed1739": // swapExactTokensForTokens
		sd.parseSwapExactTokensForTokens(data, pendingTx)
	case "8803dbee": // swapTokensForExactTokens
		sd.parseSwapTokensForExactTokens(data, pendingTx)
	case "7ff36ab5": // swapExactETHForTokens
		sd.parseSwapExactETHForTokens(data, pendingTx)
	case "18cbafe5": // swapExactTokensForETH
		sd.parseSwapExactTokensForETH(data, pendingTx)
	}
}

// parseSushiSwap parses SushiSwap transaction (similar to Uniswap)
func (sd *SandwichDetector) parseSushiSwap(tx *types.Transaction, pendingTx *PendingTransaction) {
	// Similar to Uniswap parsing
	sd.parseUniswapSwap(tx, pendingTx)
}

// parseSwapExactTokensForTokens parses swapExactTokensForTokens call
func (sd *SandwichDetector) parseSwapExactTokensForTokens(data []byte, pendingTx *PendingTransaction) {
	// Simplified parsing - in production, use proper ABI decoding
	if len(data) < 164 { // 4 + 32*5 bytes minimum
		return
	}

	// Extract amountIn (first parameter)
	amountInBytes := data[4:36]
	amountIn := new(big.Int).SetBytes(amountInBytes)
	pendingTx.AmountIn = decimal.NewFromBigInt(amountIn, 0)

	// Extract path (tokens involved in swap)
	// This is a simplified extraction - proper ABI decoding needed
	if len(data) >= 196 {
		tokenInBytes := data[164:196]
		pendingTx.TokenIn = common.BytesToAddress(tokenInBytes[12:]).Hex()
	}
}

// parseSwapTokensForExactTokens parses swapTokensForExactTokens call
func (sd *SandwichDetector) parseSwapTokensForExactTokens(data []byte, pendingTx *PendingTransaction) {
	// Similar to parseSwapExactTokensForTokens but for exact output swaps
	if len(data) < 164 {
		return
	}

	amountOutBytes := data[4:36]
	amountOut := new(big.Int).SetBytes(amountOutBytes)
	pendingTx.AmountOut = decimal.NewFromBigInt(amountOut, 0)
}

// parseSwapExactETHForTokens parses swapExactETHForTokens call
func (sd *SandwichDetector) parseSwapExactETHForTokens(data []byte, pendingTx *PendingTransaction) {
	pendingTx.TokenIn = "0x0000000000000000000000000000000000000000" // ETH
	pendingTx.AmountIn = decimal.NewFromBigInt(pendingTx.Value, 0)
}

// parseSwapExactTokensForETH parses swapExactTokensForETH call
func (sd *SandwichDetector) parseSwapExactTokensForETH(data []byte, pendingTx *PendingTransaction) {
	pendingTx.TokenOut = "0x0000000000000000000000000000000000000000" // ETH

	if len(data) >= 36 {
		amountInBytes := data[4:36]
		amountIn := new(big.Int).SetBytes(amountInBytes)
		pendingTx.AmountIn = decimal.NewFromBigInt(amountIn, 0)
	}
}

// findSandwichPattern looks for sandwich attack patterns
func (sd *SandwichDetector) findSandwichPattern(victimTx *PendingTransaction) *SandwichPattern {
	sd.mutex.RLock()
	defer sd.mutex.RUnlock()

	// Look for potential frontrun transactions
	for _, tx := range sd.pendingTransactions {
		if sd.isPotentialFrontrun(tx, victimTx) {
			// Look for corresponding backrun
			if backrunTx := sd.findBackrunTransaction(tx, victimTx); backrunTx != nil {
				pattern := &SandwichPattern{
					ID:         sd.generatePatternID(),
					FrontrunTx: tx,
					VictimTx:   victimTx,
					BackrunTx:  backrunTx,
					DetectedAt: time.Now(),
				}

				// Calculate estimated profit and confidence
				pattern.EstimatedProfit = sd.calculateEstimatedProfit(pattern)
				pattern.Confidence = sd.calculateConfidence(pattern)

				if pattern.Confidence.GreaterThan(decimal.NewFromFloat(0.7)) {
					sd.detectedSandwiches[pattern.ID] = pattern
					return pattern
				}
			}
		}
	}

	return nil
}

// isPotentialFrontrun checks if a transaction could be a frontrun
func (sd *SandwichDetector) isPotentialFrontrun(frontrunTx, victimTx *PendingTransaction) bool {
	// Check if both are swaps on the same DEX
	if !frontrunTx.IsSwap || !victimTx.IsSwap {
		return false
	}

	// Check if they involve the same token pair
	if frontrunTx.TokenIn != victimTx.TokenIn || frontrunTx.TokenOut != victimTx.TokenOut {
		return false
	}

	// Check if frontrun has higher gas price
	if frontrunTx.GasPrice.Cmp(victimTx.GasPrice) <= 0 {
		return false
	}

	// Check timing (frontrun should be slightly before victim)
	timeDiff := victimTx.Timestamp.Sub(frontrunTx.Timestamp)
	if timeDiff < 0 || timeDiff > 10*time.Second {
		return false
	}

	return true
}

// findBackrunTransaction finds a potential backrun transaction
func (sd *SandwichDetector) findBackrunTransaction(frontrunTx, victimTx *PendingTransaction) *PendingTransaction {
	for _, tx := range sd.pendingTransactions {
		if sd.isPotentialBackrun(tx, frontrunTx, victimTx) {
			return tx
		}
	}
	return nil
}

// isPotentialBackrun checks if a transaction could be a backrun
func (sd *SandwichDetector) isPotentialBackrun(backrunTx, frontrunTx, victimTx *PendingTransaction) bool {
	// Check if it's a swap
	if !backrunTx.IsSwap {
		return false
	}

	// Check if it's from the same attacker
	if backrunTx.From != frontrunTx.From {
		return false
	}

	// Check if it reverses the frontrun trade
	if backrunTx.TokenIn != frontrunTx.TokenOut || backrunTx.TokenOut != frontrunTx.TokenIn {
		return false
	}

	// Check timing (backrun should be after victim)
	if backrunTx.Timestamp.Before(victimTx.Timestamp) {
		return false
	}

	return true
}

// calculateEstimatedProfit calculates the estimated profit from a sandwich attack
func (sd *SandwichDetector) calculateEstimatedProfit(pattern *SandwichPattern) decimal.Decimal {
	// Simplified calculation - in production, use proper price impact modeling
	frontrunAmount := pattern.FrontrunTx.AmountIn
	backrunAmount := pattern.BackrunTx.AmountOut

	if frontrunAmount.IsZero() || backrunAmount.IsZero() {
		return decimal.Zero
	}

	profit := backrunAmount.Sub(frontrunAmount)
	return profit.Div(frontrunAmount) // Return as percentage
}

// calculateConfidence calculates confidence score for sandwich detection
func (sd *SandwichDetector) calculateConfidence(pattern *SandwichPattern) decimal.Decimal {
	confidence := decimal.NewFromFloat(0.5) // Base confidence

	// Increase confidence based on various factors

	// Same attacker for frontrun and backrun
	if pattern.FrontrunTx.From == pattern.BackrunTx.From {
		confidence = confidence.Add(decimal.NewFromFloat(0.3))
	}

	// High gas price difference
	gasPriceDiff := new(big.Int).Sub(pattern.FrontrunTx.GasPrice, pattern.VictimTx.GasPrice)
	if gasPriceDiff.Cmp(big.NewInt(1000000000)) > 0 { // > 1 gwei difference
		confidence = confidence.Add(decimal.NewFromFloat(0.2))
	}

	// Timing pattern
	frontrunTime := pattern.FrontrunTx.Timestamp
	victimTime := pattern.VictimTx.Timestamp
	backrunTime := pattern.BackrunTx.Timestamp

	if frontrunTime.Before(victimTime) && victimTime.Before(backrunTime) {
		confidence = confidence.Add(decimal.NewFromFloat(0.2))
	}

	// Cap at 1.0
	if confidence.GreaterThan(decimal.NewFromFloat(1.0)) {
		confidence = decimal.NewFromFloat(1.0)
	}

	return confidence
}

// generateDetectionID generates a unique detection ID
func (sd *SandwichDetector) generateDetectionID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("sandwich_%s", hex.EncodeToString(bytes))
}

// generatePatternID generates a unique pattern ID
func (sd *SandwichDetector) generatePatternID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("pattern_%s", hex.EncodeToString(bytes))
}

// cleanupLoop periodically cleans up old transactions
func (sd *SandwichDetector) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sd.stopChan:
			return
		case <-ticker.C:
			sd.cleanup()
		}
	}
}

// cleanup removes old transactions from memory
func (sd *SandwichDetector) cleanup() {
	sd.mutex.Lock()
	defer sd.mutex.Unlock()

	cutoff := time.Now().Add(-sd.timeWindow)

	for hash, tx := range sd.pendingTransactions {
		if tx.Timestamp.Before(cutoff) {
			delete(sd.pendingTransactions, hash)
		}
	}

	// Also cleanup old patterns
	for id, pattern := range sd.detectedSandwiches {
		if pattern.DetectedAt.Before(cutoff) {
			delete(sd.detectedSandwiches, id)
		}
	}
}

// detectionLoop runs the main detection logic
func (sd *SandwichDetector) detectionLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sd.stopChan:
			return
		case <-ticker.C:
			// Periodic analysis of pending transactions
			sd.analyzePendingTransactions()
		}
	}
}

// analyzePendingTransactions analyzes all pending transactions for patterns
func (sd *SandwichDetector) analyzePendingTransactions() {
	sd.mutex.RLock()
	transactions := make([]*PendingTransaction, 0, len(sd.pendingTransactions))
	for _, tx := range sd.pendingTransactions {
		transactions = append(transactions, tx)
	}
	sd.mutex.RUnlock()

	// Look for new patterns among recent transactions
	for _, tx := range transactions {
		if tx.IsSwap && time.Since(tx.Timestamp) < 30*time.Second {
			sd.findSandwichPattern(tx)
		}
	}
}
