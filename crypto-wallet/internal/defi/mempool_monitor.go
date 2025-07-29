package defi

import (
	"context"
	"encoding/json"
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

// MempoolMonitor monitors the mempool for MEV opportunities and threats
type MempoolMonitor struct {
	logger *logger.Logger
	cache  redis.Client

	// Monitoring configuration
	maxTransactions    int
	gasThreshold       *big.Int
	valueThreshold     *big.Int
	monitoringInterval time.Duration

	// State tracking
	pendingTransactions map[string]*MempoolTransaction
	gasStatistics       *GasStatistics
	mempoolMetrics      *MempoolMetrics
	mutex               sync.RWMutex
	stopChan            chan struct{}
	isRunning           bool
}

// MempoolTransaction represents a transaction in the mempool
type MempoolTransaction struct {
	Hash            string          `json:"hash"`
	From            string          `json:"from"`
	To              string          `json:"to"`
	Value           *big.Int        `json:"value"`
	GasPrice        *big.Int        `json:"gas_price"`
	GasLimit        uint64          `json:"gas_limit"`
	Data            []byte          `json:"data"`
	Nonce           uint64          `json:"nonce"`
	Timestamp       time.Time       `json:"timestamp"`
	Priority        int             `json:"priority"`
	MEVRisk         MEVRiskLevel    `json:"mev_risk"`
	EstimatedProfit decimal.Decimal `json:"estimated_profit"`
}

// MEVRiskLevel represents the MEV risk level of a transaction
type MEVRiskLevel string

const (
	MEVRiskLow      MEVRiskLevel = "low"
	MEVRiskMedium   MEVRiskLevel = "medium"
	MEVRiskHigh     MEVRiskLevel = "high"
	MEVRiskCritical MEVRiskLevel = "critical"
)

// GasStatistics holds gas price statistics
type GasStatistics struct {
	MinGasPrice      *big.Int  `json:"min_gas_price"`
	MaxGasPrice      *big.Int  `json:"max_gas_price"`
	AvgGasPrice      *big.Int  `json:"avg_gas_price"`
	MedianGasPrice   *big.Int  `json:"median_gas_price"`
	GasPriceP95      *big.Int  `json:"gas_price_p95"`
	TotalTxCount     int64     `json:"total_tx_count"`
	HighValueTxCount int64     `json:"high_value_tx_count"`
	LastUpdate       time.Time `json:"last_update"`
}

// MempoolMetrics holds mempool monitoring metrics
type MempoolMetrics struct {
	TotalTransactions    int64           `json:"total_transactions"`
	HighRiskTransactions int64           `json:"high_risk_transactions"`
	AverageGasPrice      decimal.Decimal `json:"average_gas_price"`
	MempoolSize          int             `json:"mempool_size"`
	ProcessingRate       decimal.Decimal `json:"processing_rate"`
	LastUpdate           time.Time       `json:"last_update"`
}

// NewMempoolMonitor creates a new mempool monitor
func NewMempoolMonitor(logger *logger.Logger, cache redis.Client) *MempoolMonitor {
	return &MempoolMonitor{
		logger:              logger.Named("mempool-monitor"),
		cache:               cache,
		maxTransactions:     10000,
		gasThreshold:        big.NewInt(20000000000),         // 20 gwei
		valueThreshold:      big.NewInt(1000000000000000000), // 1 ETH
		monitoringInterval:  1 * time.Second,
		pendingTransactions: make(map[string]*MempoolTransaction),
		gasStatistics:       &GasStatistics{},
		mempoolMetrics:      &MempoolMetrics{},
		stopChan:            make(chan struct{}),
	}
}

// Start starts the mempool monitor
func (mm *MempoolMonitor) Start(ctx context.Context) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if mm.isRunning {
		return fmt.Errorf("mempool monitor is already running")
	}

	mm.logger.Info("Starting mempool monitor")
	mm.isRunning = true

	// Start monitoring routines
	go mm.monitoringLoop(ctx)
	go mm.statisticsLoop(ctx)
	go mm.cleanupLoop(ctx)

	mm.logger.Info("Mempool monitor started successfully")
	return nil
}

// Stop stops the mempool monitor
func (mm *MempoolMonitor) Stop() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if !mm.isRunning {
		return
	}

	mm.logger.Info("Stopping mempool monitor")
	mm.isRunning = false
	close(mm.stopChan)
}

// AddTransaction adds a transaction to mempool monitoring
func (mm *MempoolMonitor) AddTransaction(tx *types.Transaction) {
	mempoolTx := mm.parseTransaction(tx)

	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// Check if we're at capacity
	if len(mm.pendingTransactions) >= mm.maxTransactions {
		// Remove oldest transaction
		mm.removeOldestTransaction()
	}

	mm.pendingTransactions[mempoolTx.Hash] = mempoolTx
	mm.mempoolMetrics.TotalTransactions++

	// Update high risk counter
	if mempoolTx.MEVRisk == MEVRiskHigh || mempoolTx.MEVRisk == MEVRiskCritical {
		mm.mempoolMetrics.HighRiskTransactions++
	}

	mm.logger.Debug("Added transaction to mempool monitoring",
		zap.String("hash", mempoolTx.Hash),
		zap.String("mev_risk", string(mempoolTx.MEVRisk)),
		zap.String("gas_price", mempoolTx.GasPrice.String()))
}

// RemoveTransaction removes a transaction from mempool monitoring
func (mm *MempoolMonitor) RemoveTransaction(hash string) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	delete(mm.pendingTransactions, hash)
}

// parseTransaction parses a transaction for mempool monitoring
func (mm *MempoolMonitor) parseTransaction(tx *types.Transaction) *MempoolTransaction {
	mempoolTx := &MempoolTransaction{
		Hash:      tx.Hash().Hex(),
		From:      "", // Will be recovered from signature
		Value:     tx.Value(),
		GasPrice:  tx.GasPrice(),
		GasLimit:  tx.Gas(),
		Data:      tx.Data(),
		Nonce:     tx.Nonce(),
		Timestamp: time.Now(),
		Priority:  mm.calculatePriority(tx),
	}

	if tx.To() != nil {
		mempoolTx.To = tx.To().Hex()
	}

	// Assess MEV risk
	mempoolTx.MEVRisk = mm.assessMEVRisk(mempoolTx)

	// Estimate potential profit
	mempoolTx.EstimatedProfit = mm.estimateProfit(mempoolTx)

	return mempoolTx
}

// calculatePriority calculates transaction priority
func (mm *MempoolMonitor) calculatePriority(tx *types.Transaction) int {
	priority := 0

	// High gas price increases priority
	if tx.GasPrice().Cmp(mm.gasThreshold) > 0 {
		priority += 10
	}

	// High value increases priority
	if tx.Value().Cmp(mm.valueThreshold) > 0 {
		priority += 20
	}

	// Contract interaction increases priority
	if tx.To() != nil && len(tx.Data()) > 0 {
		priority += 5
	}

	return priority
}

// assessMEVRisk assesses the MEV risk level of a transaction
func (mm *MempoolMonitor) assessMEVRisk(tx *MempoolTransaction) MEVRiskLevel {
	risk := MEVRiskLow

	// High gas price indicates potential MEV
	gasRatio := decimal.NewFromBigInt(tx.GasPrice, 0).Div(decimal.NewFromBigInt(mm.gasThreshold, 0))
	if gasRatio.GreaterThan(decimal.NewFromFloat(2.0)) {
		risk = MEVRiskMedium
	}
	if gasRatio.GreaterThan(decimal.NewFromFloat(5.0)) {
		risk = MEVRiskHigh
	}

	// High value transactions are attractive to MEV
	if tx.Value.Cmp(mm.valueThreshold) > 0 {
		if risk == MEVRiskLow {
			risk = MEVRiskMedium
		} else if risk == MEVRiskMedium {
			risk = MEVRiskHigh
		}
	}

	// Check for known MEV-prone patterns
	if mm.isArbitrageTransaction(tx) || mm.isLiquidationTransaction(tx) {
		risk = MEVRiskHigh
	}

	// DEX interactions are high risk
	if mm.isDEXTransaction(tx) {
		if risk == MEVRiskLow {
			risk = MEVRiskMedium
		} else if risk == MEVRiskMedium {
			risk = MEVRiskHigh
		} else if risk == MEVRiskHigh {
			risk = MEVRiskCritical
		}
	}

	return risk
}

// isArbitrageTransaction checks if transaction is likely an arbitrage
func (mm *MempoolMonitor) isArbitrageTransaction(tx *MempoolTransaction) bool {
	if len(tx.Data) < 4 {
		return false
	}

	// Check for common arbitrage method signatures
	methodSig := fmt.Sprintf("%x", tx.Data[:4])
	arbitrageSigs := []string{
		"38ed1739", // swapExactTokensForTokens
		"8803dbee", // swapTokensForExactTokens
		"7ff36ab5", // swapExactETHForTokens
		"18cbafe5", // swapExactTokensForETH
	}

	for _, sig := range arbitrageSigs {
		if methodSig == sig {
			return true
		}
	}

	return false
}

// isLiquidationTransaction checks if transaction is likely a liquidation
func (mm *MempoolMonitor) isLiquidationTransaction(tx *MempoolTransaction) bool {
	if len(tx.Data) < 4 {
		return false
	}

	methodSig := fmt.Sprintf("%x", tx.Data[:4])
	liquidationSigs := []string{
		"96cd4ddb", // liquidateBorrow (Compound)
		"00a718a9", // liquidationCall (Aave)
		"f5e3c462", // liquidate (various protocols)
	}

	for _, sig := range liquidationSigs {
		if methodSig == sig {
			return true
		}
	}

	return false
}

// isDEXTransaction checks if transaction is a DEX interaction
func (mm *MempoolMonitor) isDEXTransaction(tx *MempoolTransaction) bool {
	// Check for known DEX router addresses
	dexRouters := []string{
		"0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D", // Uniswap V2
		"0xE592427A0AEce92De3Edee1F18E0157C05861564", // Uniswap V3
		"0xd9e1cE17f2641f24aE83637ab66a2cca9C378B9F", // SushiSwap
		"0x1111111254fb6c44bAC0beD2854e76F90643097d", // 1inch V4
	}

	for _, router := range dexRouters {
		if tx.To == router {
			return true
		}
	}

	return false
}

// estimateProfit estimates potential profit from a transaction
func (mm *MempoolMonitor) estimateProfit(tx *MempoolTransaction) decimal.Decimal {
	baseValue := decimal.NewFromBigInt(tx.Value, -18) // Convert to ETH

	switch tx.MEVRisk {
	case MEVRiskCritical:
		return baseValue.Mul(decimal.NewFromFloat(0.05)) // 5%
	case MEVRiskHigh:
		return baseValue.Mul(decimal.NewFromFloat(0.03)) // 3%
	case MEVRiskMedium:
		return baseValue.Mul(decimal.NewFromFloat(0.01)) // 1%
	default:
		return decimal.Zero
	}
}

// removeOldestTransaction removes the oldest transaction from monitoring
func (mm *MempoolMonitor) removeOldestTransaction() {
	var oldestHash string
	var oldestTime time.Time

	for hash, tx := range mm.pendingTransactions {
		if oldestHash == "" || tx.Timestamp.Before(oldestTime) {
			oldestHash = hash
			oldestTime = tx.Timestamp
		}
	}

	if oldestHash != "" {
		delete(mm.pendingTransactions, oldestHash)
	}
}

// monitoringLoop runs the main monitoring logic
func (mm *MempoolMonitor) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(mm.monitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-mm.stopChan:
			return
		case <-ticker.C:
			mm.processMempool()
		}
	}
}

// processMempool processes the current mempool state
func (mm *MempoolMonitor) processMempool() {
	mm.mutex.RLock()
	txCount := len(mm.pendingTransactions)
	mm.mutex.RUnlock()

	mm.logger.Debug("Processing mempool",
		zap.Int("transaction_count", txCount))

	// Update mempool size metric
	mm.mutex.Lock()
	mm.mempoolMetrics.MempoolSize = txCount
	mm.mempoolMetrics.LastUpdate = time.Now()
	mm.mutex.Unlock()
}

// statisticsLoop calculates and updates gas statistics
func (mm *MempoolMonitor) statisticsLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-mm.stopChan:
			return
		case <-ticker.C:
			mm.updateGasStatistics()
		}
	}
}

// updateGasStatistics updates gas price statistics
func (mm *MempoolMonitor) updateGasStatistics() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if len(mm.pendingTransactions) == 0 {
		return
	}

	gasPrices := make([]*big.Int, 0, len(mm.pendingTransactions))
	totalGas := big.NewInt(0)
	highValueCount := int64(0)

	for _, tx := range mm.pendingTransactions {
		gasPrices = append(gasPrices, tx.GasPrice)
		totalGas.Add(totalGas, tx.GasPrice)

		if tx.Value.Cmp(mm.valueThreshold) > 0 {
			highValueCount++
		}
	}

	// Calculate statistics
	mm.gasStatistics.TotalTxCount = int64(len(mm.pendingTransactions))
	mm.gasStatistics.HighValueTxCount = highValueCount
	mm.gasStatistics.LastUpdate = time.Now()

	if len(gasPrices) > 0 {
		// Min and Max
		mm.gasStatistics.MinGasPrice = new(big.Int).Set(gasPrices[0])
		mm.gasStatistics.MaxGasPrice = new(big.Int).Set(gasPrices[0])

		for _, price := range gasPrices {
			if price.Cmp(mm.gasStatistics.MinGasPrice) < 0 {
				mm.gasStatistics.MinGasPrice.Set(price)
			}
			if price.Cmp(mm.gasStatistics.MaxGasPrice) > 0 {
				mm.gasStatistics.MaxGasPrice.Set(price)
			}
		}

		// Average
		mm.gasStatistics.AvgGasPrice = new(big.Int).Div(totalGas, big.NewInt(int64(len(gasPrices))))

		// Cache statistics
		statsJSON, _ := json.Marshal(mm.gasStatistics)
		mm.cache.Set(context.Background(), "mempool:gas_stats", string(statsJSON), 1*time.Minute)
	}
}

// cleanupLoop periodically cleans up old transactions
func (mm *MempoolMonitor) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-mm.stopChan:
			return
		case <-ticker.C:
			mm.cleanup()
		}
	}
}

// cleanup removes old transactions from monitoring
func (mm *MempoolMonitor) cleanup() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	cutoff := time.Now().Add(-10 * time.Minute)

	for hash, tx := range mm.pendingTransactions {
		if tx.Timestamp.Before(cutoff) {
			delete(mm.pendingTransactions, hash)
		}
	}

	mm.logger.Debug("Cleaned up old transactions",
		zap.Int("remaining_count", len(mm.pendingTransactions)))
}

// GetGasStatistics returns current gas statistics
func (mm *MempoolMonitor) GetGasStatistics() *GasStatistics {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	// Return a copy
	stats := *mm.gasStatistics
	return &stats
}

// GetMempoolMetrics returns current mempool metrics
func (mm *MempoolMonitor) GetMempoolMetrics() *MempoolMetrics {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	// Return a copy
	metrics := *mm.mempoolMetrics
	return &metrics
}

// GetHighRiskTransactions returns transactions with high MEV risk
func (mm *MempoolMonitor) GetHighRiskTransactions() []*MempoolTransaction {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	var highRiskTxs []*MempoolTransaction
	for _, tx := range mm.pendingTransactions {
		if tx.MEVRisk == MEVRiskHigh || tx.MEVRisk == MEVRiskCritical {
			highRiskTxs = append(highRiskTxs, tx)
		}
	}

	return highRiskTxs
}
