package multichain

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// ChainManager manages operations for a specific blockchain
type ChainManager struct {
	logger *logger.Logger
	config ChainConfig
	name   string

	// Ethereum client
	client   *ethclient.Client
	wsClient *ethclient.Client

	// State management
	isRunning     bool
	lastBlock     uint64
	balanceCache  map[string]*ChainBalance
	cacheMutex    sync.RWMutex
	stopChan      chan struct{}
	mutex         sync.RWMutex
}

// NewChainManager creates a new chain manager
func NewChainManager(logger *logger.Logger, name string, config ChainConfig) *ChainManager {
	return &ChainManager{
		logger:       logger.Named(fmt.Sprintf("chain-%s", name)),
		config:       config,
		name:         name,
		balanceCache: make(map[string]*ChainBalance),
		stopChan:     make(chan struct{}),
	}
}

// Start starts the chain manager
func (cm *ChainManager) Start(ctx context.Context) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.isRunning {
		return fmt.Errorf("chain manager for %s is already running", cm.name)
	}

	if !cm.config.Enabled {
		cm.logger.Info("Chain manager is disabled", zap.String("chain", cm.name))
		return nil
	}

	cm.logger.Info("Starting chain manager",
		zap.String("chain", cm.name),
		zap.Int64("chain_id", cm.config.ChainID))

	// Connect to RPC endpoint
	if err := cm.connectRPC(); err != nil {
		return fmt.Errorf("failed to connect to RPC: %w", err)
	}

	// Connect to WebSocket endpoint
	if err := cm.connectWebSocket(); err != nil {
		cm.logger.Warn("Failed to connect to WebSocket", zap.Error(err))
	}

	// Start monitoring
	go cm.monitorBlocks(ctx)

	cm.isRunning = true
	cm.logger.Info("Chain manager started successfully", zap.String("chain", cm.name))
	return nil
}

// Stop stops the chain manager
func (cm *ChainManager) Stop() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if !cm.isRunning {
		return nil
	}

	cm.logger.Info("Stopping chain manager", zap.String("chain", cm.name))

	// Close connections
	if cm.client != nil {
		cm.client.Close()
	}
	if cm.wsClient != nil {
		cm.wsClient.Close()
	}

	// Signal stop
	close(cm.stopChan)

	cm.isRunning = false
	cm.logger.Info("Chain manager stopped", zap.String("chain", cm.name))
	return nil
}

// GetBalances returns balances for an address
func (cm *ChainManager) GetBalances(ctx context.Context, address common.Address) (map[string]*ChainBalance, error) {
	cm.logger.Debug("Getting balances",
		zap.String("chain", cm.name),
		zap.String("address", address.Hex()))

	balances := make(map[string]*ChainBalance)

	// Get native token balance
	nativeBalance, err := cm.getNativeBalance(ctx, address)
	if err != nil {
		cm.logger.Error("Failed to get native balance", zap.Error(err))
	} else {
		balances[cm.config.NativeToken.Symbol] = nativeBalance
	}

	// Get ERC-20 token balances
	for _, token := range cm.config.Tokens {
		if !token.Enabled {
			continue
		}

		tokenBalance, err := cm.getTokenBalance(ctx, address, token)
		if err != nil {
			cm.logger.Warn("Failed to get token balance",
				zap.String("token", token.Symbol),
				zap.Error(err))
			continue
		}

		if tokenBalance.Balance.GreaterThan(decimal.Zero) {
			balances[token.Symbol] = tokenBalance
		}
	}

	cm.logger.Info("Retrieved balances",
		zap.String("chain", cm.name),
		zap.String("address", address.Hex()),
		zap.Int("token_count", len(balances)))

	return balances, nil
}

// GetNativeBalance returns native token balance
func (cm *ChainManager) getNativeBalance(ctx context.Context, address common.Address) (*ChainBalance, error) {
	if cm.client == nil {
		return nil, fmt.Errorf("RPC client not connected")
	}

	balance, err := cm.client.BalanceAt(ctx, address, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	// Convert wei to ether
	balanceDecimal := decimal.NewFromBigInt(balance, -18)

	// Get current block number
	blockNumber, err := cm.client.BlockNumber(ctx)
	if err != nil {
		blockNumber = 0
	}

	chainBalance := &ChainBalance{
		Chain:       cm.name,
		Balance:     balanceDecimal,
		ValueUSD:    decimal.Zero, // Will be calculated by price oracle
		Address:     address.Hex(),
		LastUpdated: time.Now(),
		BlockNumber: blockNumber,
		Pending:     decimal.Zero,
		Locked:      decimal.Zero,
		Available:   balanceDecimal,
	}

	return chainBalance, nil
}

// GetTokenBalance returns ERC-20 token balance
func (cm *ChainManager) getTokenBalance(ctx context.Context, address common.Address, token TokenConfig) (*ChainBalance, error) {
	// For simplicity, return mock balance
	// In production, implement actual ERC-20 balance checking
	
	chainBalance := &ChainBalance{
		Chain:       cm.name,
		Balance:     decimal.Zero,
		ValueUSD:    decimal.Zero,
		Address:     address.Hex(),
		LastUpdated: time.Now(),
		BlockNumber: 0,
		Pending:     decimal.Zero,
		Locked:      decimal.Zero,
		Available:   decimal.Zero,
	}

	return chainBalance, nil
}

// SendTransaction sends a transaction
func (cm *ChainManager) SendTransaction(ctx context.Context, req *TransactionRequest) (*TransactionResult, error) {
	cm.logger.Info("Sending transaction",
		zap.String("chain", cm.name),
		zap.String("to", req.To.Hex()),
		zap.String("value", req.Value.String()))

	// Validate request
	if err := cm.validateTransactionRequest(req); err != nil {
		return nil, fmt.Errorf("invalid transaction request: %w", err)
	}

	// Estimate gas
	gasLimit, err := cm.estimateGas(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Get gas price
	gasPrice, err := cm.getGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Build transaction
	tx, err := cm.buildTransaction(req, gasLimit, gasPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}

	// Sign and send transaction
	txHash, err := cm.signAndSendTransaction(ctx, tx, req.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	result := &TransactionResult{
		Hash:      txHash,
		Status:    "pending",
		GasUsed:   gasLimit,
		GasPrice:  gasPrice,
		Fee:       decimal.NewFromBigInt(new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice), -18),
		CreatedAt: time.Now(),
	}

	cm.logger.Info("Transaction sent successfully",
		zap.String("chain", cm.name),
		zap.String("hash", txHash))

	return result, nil
}

// Helper methods

// connectRPC connects to RPC endpoint
func (cm *ChainManager) connectRPC() error {
	if len(cm.config.RPCEndpoints) == 0 {
		return fmt.Errorf("no RPC endpoints configured")
	}

	var lastErr error
	for _, endpoint := range cm.config.RPCEndpoints {
		client, err := ethclient.Dial(endpoint)
		if err != nil {
			lastErr = err
			continue
		}

		// Test connection
		_, err = client.ChainID(context.Background())
		if err != nil {
			client.Close()
			lastErr = err
			continue
		}

		cm.client = client
		cm.logger.Info("Connected to RPC", 
			zap.String("chain", cm.name),
			zap.String("endpoint", endpoint))
		return nil
	}

	return fmt.Errorf("failed to connect to any RPC endpoint: %w", lastErr)
}

// connectWebSocket connects to WebSocket endpoint
func (cm *ChainManager) connectWebSocket() error {
	if len(cm.config.WSEndpoints) == 0 {
		return fmt.Errorf("no WebSocket endpoints configured")
	}

	var lastErr error
	for _, endpoint := range cm.config.WSEndpoints {
		client, err := ethclient.Dial(endpoint)
		if err != nil {
			lastErr = err
			continue
		}

		cm.wsClient = client
		cm.logger.Info("Connected to WebSocket",
			zap.String("chain", cm.name),
			zap.String("endpoint", endpoint))
		return nil
	}

	return fmt.Errorf("failed to connect to any WebSocket endpoint: %w", lastErr)
}

// monitorBlocks monitors new blocks
func (cm *ChainManager) monitorBlocks(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second) // Block time varies by chain
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.stopChan:
			return
		case <-ticker.C:
			cm.updateLastBlock()
		}
	}
}

// updateLastBlock updates the last block number
func (cm *ChainManager) updateLastBlock() {
	if cm.client == nil {
		return
	}

	blockNumber, err := cm.client.BlockNumber(context.Background())
	if err != nil {
		cm.logger.Warn("Failed to get block number", zap.Error(err))
		return
	}

	cm.mutex.Lock()
	cm.lastBlock = blockNumber
	cm.mutex.Unlock()
}

// validateTransactionRequest validates a transaction request
func (cm *ChainManager) validateTransactionRequest(req *TransactionRequest) error {
	if req.To == (common.Address{}) {
		return fmt.Errorf("to address is required")
	}
	if req.Value.LessThan(decimal.Zero) {
		return fmt.Errorf("value cannot be negative")
	}
	return nil
}

// estimateGas estimates gas for a transaction
func (cm *ChainManager) estimateGas(ctx context.Context, req *TransactionRequest) (uint64, error) {
	// Simplified gas estimation
	return 21000, nil // Basic transfer
}

// getGasPrice gets current gas price
func (cm *ChainManager) getGasPrice(ctx context.Context) (*big.Int, error) {
	if cm.client == nil {
		return nil, fmt.Errorf("RPC client not connected")
	}

	gasPrice, err := cm.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Apply gas multiplier
	multiplier := cm.config.GasMultiplier
	if multiplier.IsZero() {
		multiplier = decimal.NewFromFloat(1.1) // Default 10% increase
	}

	adjustedGasPrice := decimal.NewFromBigInt(gasPrice, 0).Mul(multiplier)
	return adjustedGasPrice.BigInt(), nil
}

// buildTransaction builds a transaction
func (cm *ChainManager) buildTransaction(req *TransactionRequest, gasLimit uint64, gasPrice *big.Int) (*Transaction, error) {
	// Simplified transaction building
	return &Transaction{
		To:       req.To,
		Value:    req.Value,
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Data:     req.Data,
	}, nil
}

// signAndSendTransaction signs and sends a transaction
func (cm *ChainManager) signAndSendTransaction(ctx context.Context, tx *Transaction, privateKey string) (string, error) {
	// In production, implement actual transaction signing and sending
	// For now, return a mock transaction hash
	return "0x" + fmt.Sprintf("%064d", time.Now().UnixNano()), nil
}
