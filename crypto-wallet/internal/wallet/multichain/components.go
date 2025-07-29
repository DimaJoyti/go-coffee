package multichain

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

// GasTracker tracks gas prices across chains
type GasTracker struct {
	logger *logger.Logger
	config GasConfig

	// Gas prices
	gasPrices map[string]*GasPrice
	mutex     sync.RWMutex

	// State management
	isRunning    bool
	updateTicker *time.Ticker
	stopChan     chan struct{}
}

// NewGasTracker creates a new gas tracker
func NewGasTracker(logger *logger.Logger, config GasConfig) *GasTracker {
	return &GasTracker{
		logger:    logger.Named("gas-tracker"),
		config:    config,
		gasPrices: make(map[string]*GasPrice),
		stopChan:  make(chan struct{}),
	}
}

// Start starts the gas tracker
func (gt *GasTracker) Start(ctx context.Context) error {
	if gt.isRunning {
		return fmt.Errorf("gas tracker is already running")
	}

	if !gt.config.Enabled {
		gt.logger.Info("Gas tracker is disabled")
		return nil
	}

	gt.logger.Info("Starting gas tracker")

	// Initial update
	gt.updateGasPrices()

	// Start update ticker
	gt.updateTicker = time.NewTicker(gt.config.UpdateInterval)
	go gt.updateLoop(ctx)

	gt.isRunning = true
	gt.logger.Info("Gas tracker started successfully")
	return nil
}

// Stop stops the gas tracker
func (gt *GasTracker) Stop() error {
	if !gt.isRunning {
		return nil
	}

	gt.logger.Info("Stopping gas tracker")

	if gt.updateTicker != nil {
		gt.updateTicker.Stop()
	}

	close(gt.stopChan)
	gt.isRunning = false
	gt.logger.Info("Gas tracker stopped")
	return nil
}

// GetGasPrice returns gas price for a chain
func (gt *GasTracker) GetGasPrice(chain string) *GasPrice {
	gt.mutex.RLock()
	defer gt.mutex.RUnlock()
	return gt.gasPrices[chain]
}

// updateLoop runs the gas price update loop
func (gt *GasTracker) updateLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-gt.stopChan:
			return
		case <-gt.updateTicker.C:
			gt.updateGasPrices()
		}
	}
}

// updateGasPrices updates gas prices for all chains
func (gt *GasTracker) updateGasPrices() {
	for chain := range gt.config.ChainConfigs {
		gasPrice := gt.fetchGasPrice(chain)
		if gasPrice != nil {
			gt.mutex.Lock()
			gt.gasPrices[chain] = gasPrice
			gt.mutex.Unlock()
		}
	}
}

// fetchGasPrice fetches gas price for a specific chain
func (gt *GasTracker) fetchGasPrice(chain string) *GasPrice {
	// Mock implementation - in production, fetch from gas stations
	return &GasPrice{
		Chain:       chain,
		Slow:        decimal.NewFromFloat(10),
		Normal:      decimal.NewFromFloat(15),
		Fast:        decimal.NewFromFloat(25),
		Instant:     decimal.NewFromFloat(35),
		LastUpdated: time.Now(),
		Source:      "mock",
	}
}

// PriceOracle tracks token prices
type PriceOracle struct {
	logger *logger.Logger
	config PriceOracleConfig

	// Token prices
	tokenPrices map[string]*TokenPrice
	mutex       sync.RWMutex

	// State management
	isRunning    bool
	updateTicker *time.Ticker
	stopChan     chan struct{}
}

// NewPriceOracle creates a new price oracle
func NewPriceOracle(logger *logger.Logger, config PriceOracleConfig) *PriceOracle {
	return &PriceOracle{
		logger:      logger.Named("price-oracle"),
		config:      config,
		tokenPrices: make(map[string]*TokenPrice),
		stopChan:    make(chan struct{}),
	}
}

// Start starts the price oracle
func (po *PriceOracle) Start(ctx context.Context) error {
	if po.isRunning {
		return fmt.Errorf("price oracle is already running")
	}

	if !po.config.Enabled {
		po.logger.Info("Price oracle is disabled")
		return nil
	}

	po.logger.Info("Starting price oracle")

	// Initial update
	po.updatePrices()

	// Start update ticker
	po.updateTicker = time.NewTicker(po.config.UpdateInterval)
	go po.updateLoop(ctx)

	po.isRunning = true
	po.logger.Info("Price oracle started successfully")
	return nil
}

// Stop stops the price oracle
func (po *PriceOracle) Stop() error {
	if !po.isRunning {
		return nil
	}

	po.logger.Info("Stopping price oracle")

	if po.updateTicker != nil {
		po.updateTicker.Stop()
	}

	close(po.stopChan)
	po.isRunning = false
	po.logger.Info("Price oracle stopped")
	return nil
}

// GetPrice returns price for a token
func (po *PriceOracle) GetPrice(tokenID string) decimal.Decimal {
	po.mutex.RLock()
	defer po.mutex.RUnlock()
	
	if price, exists := po.tokenPrices[tokenID]; exists {
		return price.PriceUSD
	}
	return decimal.Zero
}

// GetChange24h returns 24h price change for a token
func (po *PriceOracle) GetChange24h(tokenID string) decimal.Decimal {
	po.mutex.RLock()
	defer po.mutex.RUnlock()
	
	if price, exists := po.tokenPrices[tokenID]; exists {
		return price.Change24h
	}
	return decimal.Zero
}

// updateLoop runs the price update loop
func (po *PriceOracle) updateLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-po.stopChan:
			return
		case <-po.updateTicker.C:
			po.updatePrices()
		}
	}
}

// updatePrices updates token prices
func (po *PriceOracle) updatePrices() {
	for _, tokenID := range po.config.SupportedTokens {
		price := po.fetchTokenPrice(tokenID)
		if price != nil {
			po.mutex.Lock()
			po.tokenPrices[tokenID] = price
			po.mutex.Unlock()
		}
	}
}

// fetchTokenPrice fetches price for a specific token
func (po *PriceOracle) fetchTokenPrice(tokenID string) *TokenPrice {
	// Mock implementation - in production, fetch from price APIs
	mockPrices := map[string]decimal.Decimal{
		"ethereum":     decimal.NewFromFloat(2000),
		"bitcoin":      decimal.NewFromFloat(45000),
		"usd-coin":     decimal.NewFromFloat(1),
		"tether":       decimal.NewFromFloat(1),
		"binancecoin":  decimal.NewFromFloat(300),
		"matic-network": decimal.NewFromFloat(0.8),
	}

	price, exists := mockPrices[tokenID]
	if !exists {
		price = decimal.NewFromFloat(1) // Default price
	}

	return &TokenPrice{
		Symbol:      tokenID,
		PriceUSD:    price,
		Change24h:   decimal.NewFromFloat(-2.5), // Mock change
		Volume24h:   decimal.NewFromFloat(1000000),
		MarketCap:   price.Mul(decimal.NewFromFloat(1000000)),
		LastUpdated: time.Now(),
		Source:      "mock",
	}
}

// PortfolioTracker tracks portfolio performance
type PortfolioTracker struct {
	logger *logger.Logger
	config PortfolioConfig

	// Portfolio snapshots
	snapshots map[string][]*PortfolioSnapshot // address -> snapshots
	mutex     sync.RWMutex

	// State management
	isRunning    bool
	updateTicker *time.Ticker
	stopChan     chan struct{}
}

// NewPortfolioTracker creates a new portfolio tracker
func NewPortfolioTracker(logger *logger.Logger, config PortfolioConfig) *PortfolioTracker {
	return &PortfolioTracker{
		logger:    logger.Named("portfolio-tracker"),
		config:    config,
		snapshots: make(map[string][]*PortfolioSnapshot),
		stopChan:  make(chan struct{}),
	}
}

// Start starts the portfolio tracker
func (pt *PortfolioTracker) Start(ctx context.Context) error {
	if pt.isRunning {
		return fmt.Errorf("portfolio tracker is already running")
	}

	if !pt.config.Enabled {
		pt.logger.Info("Portfolio tracker is disabled")
		return nil
	}

	pt.logger.Info("Starting portfolio tracker")

	// Start update ticker
	pt.updateTicker = time.NewTicker(pt.config.UpdateInterval)
	go pt.updateLoop(ctx)

	pt.isRunning = true
	pt.logger.Info("Portfolio tracker started successfully")
	return nil
}

// Stop stops the portfolio tracker
func (pt *PortfolioTracker) Stop() error {
	if !pt.isRunning {
		return nil
	}

	pt.logger.Info("Stopping portfolio tracker")

	if pt.updateTicker != nil {
		pt.updateTicker.Stop()
	}

	close(pt.stopChan)
	pt.isRunning = false
	pt.logger.Info("Portfolio tracker stopped")
	return nil
}

// AddSnapshot adds a portfolio snapshot
func (pt *PortfolioTracker) AddSnapshot(address common.Address, snapshot *PortfolioSnapshot) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	addressStr := address.Hex()
	pt.snapshots[addressStr] = append(pt.snapshots[addressStr], snapshot)

	// Cleanup old snapshots
	pt.cleanupOldSnapshots(addressStr)
}

// GetSnapshots returns portfolio snapshots for an address
func (pt *PortfolioTracker) GetSnapshots(address common.Address, limit int) []*PortfolioSnapshot {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()

	addressStr := address.Hex()
	snapshots := pt.snapshots[addressStr]

	if len(snapshots) <= limit {
		return snapshots
	}

	return snapshots[len(snapshots)-limit:]
}

// updateLoop runs the portfolio update loop
func (pt *PortfolioTracker) updateLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-pt.stopChan:
			return
		case <-pt.updateTicker.C:
			pt.performMaintenance()
		}
	}
}

// performMaintenance performs portfolio maintenance tasks
func (pt *PortfolioTracker) performMaintenance() {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	// Cleanup old snapshots for all addresses
	for address := range pt.snapshots {
		pt.cleanupOldSnapshots(address)
	}
}

// cleanupOldSnapshots removes old snapshots beyond retention period
func (pt *PortfolioTracker) cleanupOldSnapshots(address string) {
	snapshots := pt.snapshots[address]
	if len(snapshots) == 0 {
		return
	}

	cutoff := time.Now().Add(-pt.config.HistoryRetention)
	var validSnapshots []*PortfolioSnapshot

	for _, snapshot := range snapshots {
		if snapshot.Timestamp.After(cutoff) {
			validSnapshots = append(validSnapshots, snapshot)
		}
	}

	pt.snapshots[address] = validSnapshots
}
