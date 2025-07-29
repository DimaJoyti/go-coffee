package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// GasOracle provides gas price estimation and optimization
type GasOracle struct {
	logger *logger.Logger
	config SmartContractConfig

	// Gas price cache
	gasPrices map[string]*GasPriceData
	mutex     sync.RWMutex

	// Monitoring
	stopChan  chan struct{}
	isRunning bool
}

// GasPriceData holds gas price information for a chain
type GasPriceData struct {
	Chain            string          `json:"chain"`
	SafeGasPrice     *big.Int        `json:"safe_gas_price"`
	StandardGasPrice *big.Int        `json:"standard_gas_price"`
	FastGasPrice     *big.Int        `json:"fast_gas_price"`
	BaseFee          *big.Int        `json:"base_fee,omitempty"`
	MaxPriorityFee   *big.Int        `json:"max_priority_fee,omitempty"`
	GasUsedRatio     decimal.Decimal `json:"gas_used_ratio"`
	BlockNumber      uint64          `json:"block_number"`
	Timestamp        time.Time       `json:"timestamp"`
	Source           string          `json:"source"`
}

// GasEstimate represents a gas estimation
type GasEstimate struct {
	GasLimit       uint64          `json:"gas_limit"`
	GasPrice       *big.Int        `json:"gas_price"`
	MaxFeePerGas   *big.Int        `json:"max_fee_per_gas,omitempty"`
	MaxPriorityFee *big.Int        `json:"max_priority_fee,omitempty"`
	EstimatedCost  decimal.Decimal `json:"estimated_cost"`
	Confidence     decimal.Decimal `json:"confidence"`
	Source         string          `json:"source"`
}

// NewGasOracle creates a new gas oracle
func NewGasOracle(logger *logger.Logger, config SmartContractConfig) *GasOracle {
	return &GasOracle{
		logger:    logger.Named("gas-oracle"),
		config:    config,
		gasPrices: make(map[string]*GasPriceData),
		stopChan:  make(chan struct{}),
	}
}

// Start starts the gas oracle
func (g *GasOracle) Start(ctx context.Context) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.isRunning {
		return fmt.Errorf("gas oracle is already running")
	}

	g.logger.Info("Starting gas oracle")
	g.isRunning = true

	// Initialize gas prices for supported chains
	for _, chain := range g.config.SupportedChains {
		g.initializeGasPrices(chain)
	}

	// Start monitoring goroutine
	go g.monitorGasPrices(ctx)

	return nil
}

// Stop stops the gas oracle
func (g *GasOracle) Stop() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if !g.isRunning {
		return nil
	}

	g.logger.Info("Stopping gas oracle")
	g.isRunning = false
	close(g.stopChan)

	return nil
}

// GetGasPrice returns the gas price for a chain and priority
func (g *GasOracle) GetGasPrice(ctx context.Context, chain string, priority TransactionPriority) (*big.Int, error) {
	g.mutex.RLock()
	gasData, exists := g.gasPrices[chain]
	g.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("gas price data not available for chain: %s", chain)
	}

	var gasPrice *big.Int
	switch priority {
	case PriorityLow:
		gasPrice = gasData.SafeGasPrice
	case PriorityNormal:
		gasPrice = gasData.StandardGasPrice
	case PriorityHigh, PriorityUrgent:
		gasPrice = gasData.FastGasPrice
	default:
		gasPrice = gasData.StandardGasPrice
	}

	// Apply gas multiplier
	if !g.config.GasMultiplier.IsZero() {
		multiplier := g.config.GasMultiplier
		gasPriceDecimal := decimal.NewFromBigInt(gasPrice, 0)
		adjustedPrice := gasPriceDecimal.Mul(multiplier)
		gasPrice = adjustedPrice.BigInt()
	}

	// Check against max gas price
	if g.config.MaxGasPrice != nil && gasPrice.Cmp(g.config.MaxGasPrice) > 0 {
		gasPrice = g.config.MaxGasPrice
	}

	g.logger.Debug("Retrieved gas price",
		zap.String("chain", chain),
		zap.String("priority", g.getPriorityString(priority)),
		zap.String("gas_price", gasPrice.String()))

	return gasPrice, nil
}

// GetEIP1559GasPrice returns EIP-1559 gas prices
func (g *GasOracle) GetEIP1559GasPrice(ctx context.Context, chain string, priority TransactionPriority) (*big.Int, *big.Int, error) {
	g.mutex.RLock()
	gasData, exists := g.gasPrices[chain]
	g.mutex.RUnlock()

	if !exists {
		return nil, nil, fmt.Errorf("gas price data not available for chain: %s", chain)
	}

	if gasData.BaseFee == nil {
		return nil, nil, fmt.Errorf("EIP-1559 not supported for chain: %s", chain)
	}

	// Calculate max fee per gas based on priority
	var priorityFee *big.Int
	switch priority {
	case PriorityLow:
		priorityFee = big.NewInt(1000000000) // 1 gwei
	case PriorityNormal:
		priorityFee = big.NewInt(2000000000) // 2 gwei
	case PriorityHigh:
		priorityFee = big.NewInt(3000000000) // 3 gwei
	case PriorityUrgent:
		priorityFee = big.NewInt(5000000000) // 5 gwei
	default:
		priorityFee = big.NewInt(2000000000) // 2 gwei
	}

	// Max fee = (base fee * 2) + priority fee
	maxFeePerGas := new(big.Int).Mul(gasData.BaseFee, big.NewInt(2))
	maxFeePerGas.Add(maxFeePerGas, priorityFee)

	// Apply gas multiplier
	if !g.config.GasMultiplier.IsZero() {
		multiplier := g.config.GasMultiplier
		maxFeeDecimal := decimal.NewFromBigInt(maxFeePerGas, 0)
		priorityFeeDecimal := decimal.NewFromBigInt(priorityFee, 0)

		adjustedMaxFee := maxFeeDecimal.Mul(multiplier)
		adjustedPriorityFee := priorityFeeDecimal.Mul(multiplier)

		maxFeePerGas = adjustedMaxFee.BigInt()
		priorityFee = adjustedPriorityFee.BigInt()
	}

	// Check against max gas price
	if g.config.MaxGasPrice != nil && maxFeePerGas.Cmp(g.config.MaxGasPrice) > 0 {
		maxFeePerGas = g.config.MaxGasPrice
	}

	g.logger.Debug("Retrieved EIP-1559 gas prices",
		zap.String("chain", chain),
		zap.String("priority", g.getPriorityString(priority)),
		zap.String("max_fee_per_gas", maxFeePerGas.String()),
		zap.String("max_priority_fee", priorityFee.String()))

	return maxFeePerGas, priorityFee, nil
}

// EstimateGasCost estimates the total cost of a transaction
func (g *GasOracle) EstimateGasCost(ctx context.Context, chain string, gasLimit uint64, priority TransactionPriority) (*GasEstimate, error) {
	gasPrice, err := g.GetGasPrice(ctx, chain, priority)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Calculate estimated cost
	gasCost := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))
	estimatedCost := decimal.NewFromBigInt(gasCost, 0)

	// Convert to ETH (assuming 18 decimals)
	estimatedCostETH := estimatedCost.Div(decimal.NewFromFloat(1e18))

	return &GasEstimate{
		GasLimit:      gasLimit,
		GasPrice:      gasPrice,
		EstimatedCost: estimatedCostETH,
		Confidence:    decimal.NewFromFloat(0.85), // 85% confidence
		Source:        "internal_oracle",
	}, nil
}

// GetGasPriceHistory returns historical gas price data
func (g *GasOracle) GetGasPriceHistory(chain string, duration time.Duration) ([]*GasPriceData, error) {
	// In a real implementation, this would return historical data
	// For now, return current data as a single point
	g.mutex.RLock()
	gasData, exists := g.gasPrices[chain]
	g.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("gas price data not available for chain: %s", chain)
	}

	return []*GasPriceData{gasData}, nil
}

// initializeGasPrices initializes gas prices for a chain
func (g *GasOracle) initializeGasPrices(chain string) {
	// Mock gas prices for different chains
	var gasData *GasPriceData

	switch chain {
	case "ethereum":
		gasData = &GasPriceData{
			Chain:            chain,
			SafeGasPrice:     big.NewInt(20000000000), // 20 gwei
			StandardGasPrice: big.NewInt(25000000000), // 25 gwei
			FastGasPrice:     big.NewInt(35000000000), // 35 gwei
			BaseFee:          big.NewInt(15000000000), // 15 gwei
			MaxPriorityFee:   big.NewInt(2000000000),  // 2 gwei
			GasUsedRatio:     decimal.NewFromFloat(0.7),
			BlockNumber:      uint64(time.Now().Unix()),
			Timestamp:        time.Now(),
			Source:           "mock_oracle",
		}
	case "polygon":
		gasData = &GasPriceData{
			Chain:            chain,
			SafeGasPrice:     big.NewInt(30000000000), // 30 gwei
			StandardGasPrice: big.NewInt(35000000000), // 35 gwei
			FastGasPrice:     big.NewInt(45000000000), // 45 gwei
			BaseFee:          big.NewInt(25000000000), // 25 gwei
			MaxPriorityFee:   big.NewInt(1000000000),  // 1 gwei
			GasUsedRatio:     decimal.NewFromFloat(0.6),
			BlockNumber:      uint64(time.Now().Unix()),
			Timestamp:        time.Now(),
			Source:           "mock_oracle",
		}
	case "arbitrum":
		gasData = &GasPriceData{
			Chain:            chain,
			SafeGasPrice:     big.NewInt(100000000), // 0.1 gwei
			StandardGasPrice: big.NewInt(200000000), // 0.2 gwei
			FastGasPrice:     big.NewInt(500000000), // 0.5 gwei
			BaseFee:          big.NewInt(100000000), // 0.1 gwei
			MaxPriorityFee:   big.NewInt(10000000),  // 0.01 gwei
			GasUsedRatio:     decimal.NewFromFloat(0.5),
			BlockNumber:      uint64(time.Now().Unix()),
			Timestamp:        time.Now(),
			Source:           "mock_oracle",
		}
	default:
		gasData = &GasPriceData{
			Chain:            chain,
			SafeGasPrice:     big.NewInt(1000000000), // 1 gwei
			StandardGasPrice: big.NewInt(2000000000), // 2 gwei
			FastGasPrice:     big.NewInt(5000000000), // 5 gwei
			GasUsedRatio:     decimal.NewFromFloat(0.5),
			BlockNumber:      uint64(time.Now().Unix()),
			Timestamp:        time.Now(),
			Source:           "mock_oracle",
		}
	}

	g.mutex.Lock()
	g.gasPrices[chain] = gasData
	g.mutex.Unlock()

	g.logger.Info("Initialized gas prices for chain",
		zap.String("chain", chain),
		zap.String("standard_price", gasData.StandardGasPrice.String()))
}

// monitorGasPrices monitors and updates gas prices
func (g *GasOracle) monitorGasPrices(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Update every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-g.stopChan:
			return
		case <-ticker.C:
			g.updateGasPrices()
		}
	}
}

// updateGasPrices updates gas prices for all chains
func (g *GasOracle) updateGasPrices() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	for chain := range g.gasPrices {
		// Simulate gas price fluctuation
		gasData := g.gasPrices[chain]

		// Add some randomness to simulate real gas price changes
		variation := decimal.NewFromFloat(0.9 + (0.2 * float64(time.Now().UnixNano()%100) / 100))

		standardPrice := decimal.NewFromBigInt(gasData.StandardGasPrice, 0)
		newStandardPrice := standardPrice.Mul(variation)

		gasData.StandardGasPrice = newStandardPrice.BigInt()
		gasData.SafeGasPrice = new(big.Int).Mul(gasData.StandardGasPrice, big.NewInt(8))
		gasData.SafeGasPrice.Div(gasData.SafeGasPrice, big.NewInt(10)) // 80% of standard
		gasData.FastGasPrice = new(big.Int).Mul(gasData.StandardGasPrice, big.NewInt(14))
		gasData.FastGasPrice.Div(gasData.FastGasPrice, big.NewInt(10)) // 140% of standard

		gasData.Timestamp = time.Now()
		gasData.BlockNumber++
	}
}

// getPriorityString converts priority enum to string
func (g *GasOracle) getPriorityString(priority TransactionPriority) string {
	switch priority {
	case PriorityLow:
		return "low"
	case PriorityNormal:
		return "normal"
	case PriorityHigh:
		return "high"
	case PriorityUrgent:
		return "urgent"
	default:
		return "unknown"
	}
}
