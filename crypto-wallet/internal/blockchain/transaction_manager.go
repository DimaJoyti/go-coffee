package blockchain

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

// TransactionManager handles transaction execution and monitoring
type TransactionManager struct {
	logger *logger.Logger
	config SmartContractConfig

	// Transaction tracking
	pendingTransactions   map[string]*TransactionRequest
	completedTransactions map[string]*TransactionResult
	mutex                 sync.RWMutex

	// Monitoring
	stopChan  chan struct{}
	isRunning bool
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(logger *logger.Logger, config SmartContractConfig) *TransactionManager {
	return &TransactionManager{
		logger:                logger.Named("transaction-manager"),
		config:                config,
		pendingTransactions:   make(map[string]*TransactionRequest),
		completedTransactions: make(map[string]*TransactionResult),
		stopChan:              make(chan struct{}),
	}
}

// Start starts the transaction manager
func (tm *TransactionManager) Start(ctx context.Context) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if tm.isRunning {
		return fmt.Errorf("transaction manager is already running")
	}

	tm.logger.Info("Starting transaction manager")
	tm.isRunning = true

	// Start monitoring goroutine
	go tm.monitorTransactions(ctx)

	return nil
}

// Stop stops the transaction manager
func (tm *TransactionManager) Stop() error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if !tm.isRunning {
		return nil
	}

	tm.logger.Info("Stopping transaction manager")
	tm.isRunning = false
	close(tm.stopChan)

	return nil
}

// ExecuteTransaction executes a transaction
func (tm *TransactionManager) ExecuteTransaction(ctx context.Context, request *TransactionRequest) (*TransactionResult, error) {
	tm.logger.Info("Executing transaction",
		zap.String("id", request.ID),
		zap.String("function", request.FunctionName))

	// Track pending transaction
	tm.mutex.Lock()
	tm.pendingTransactions[request.ID] = request
	tm.mutex.Unlock()

	// Create mock transaction result for demonstration
	result := &TransactionResult{
		ID:              request.ID,
		TransactionHash: common.HexToHash(fmt.Sprintf("0x%s%d", request.ID, time.Now().Unix())),
		BlockNumber:     uint64(time.Now().Unix()),
		BlockHash:       common.HexToHash(fmt.Sprintf("0xblock%d", time.Now().Unix())),
		GasUsed:         request.GasLimit * 80 / 100, // Simulate 80% gas usage
		GasPrice:        request.GasPrice,
		Status:          StatusConfirmed,
		Logs:            []types.Log{},
		Events:          []ParsedEvent{},
		ExecutedAt:      time.Now(),
		ConfirmedAt:     &[]time.Time{time.Now().Add(15 * time.Second)}[0],
		Metadata:        request.Metadata,
	}

	// Move to completed transactions
	tm.mutex.Lock()
	delete(tm.pendingTransactions, request.ID)
	tm.completedTransactions[request.ID] = result
	tm.mutex.Unlock()

	tm.logger.Info("Transaction executed successfully",
		zap.String("id", request.ID),
		zap.String("tx_hash", result.TransactionHash.Hex()))

	return result, nil
}

// GetTransactionStatus returns the status of a transaction
func (tm *TransactionManager) GetTransactionStatus(txID string) (*TransactionResult, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	if result, exists := tm.completedTransactions[txID]; exists {
		return result, nil
	}

	if _, exists := tm.pendingTransactions[txID]; exists {
		return &TransactionResult{
			ID:     txID,
			Status: StatusPending,
		}, nil
	}

	return nil, fmt.Errorf("transaction not found: %s", txID)
}

// GetPendingTransactions returns all pending transactions
func (tm *TransactionManager) GetPendingTransactions() []*TransactionRequest {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	transactions := make([]*TransactionRequest, 0, len(tm.pendingTransactions))
	for _, tx := range tm.pendingTransactions {
		transactions = append(transactions, tx)
	}

	return transactions
}

// GetCompletedTransactions returns all completed transactions
func (tm *TransactionManager) GetCompletedTransactions() []*TransactionResult {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	transactions := make([]*TransactionResult, 0, len(tm.completedTransactions))
	for _, tx := range tm.completedTransactions {
		transactions = append(transactions, tx)
	}

	return transactions
}

// monitorTransactions monitors pending transactions
func (tm *TransactionManager) monitorTransactions(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tm.stopChan:
			return
		case <-ticker.C:
			tm.checkPendingTransactions()
		}
	}
}

// checkPendingTransactions checks the status of pending transactions
func (tm *TransactionManager) checkPendingTransactions() {
	tm.mutex.RLock()
	pendingCount := len(tm.pendingTransactions)
	tm.mutex.RUnlock()

	if pendingCount > 0 {
		tm.logger.Debug("Monitoring pending transactions", zap.Int("count", pendingCount))
	}

	// In a real implementation, this would check transaction status on-chain
	// For now, we'll just log the monitoring activity
}

// CancelTransaction cancels a pending transaction
func (tm *TransactionManager) CancelTransaction(txID string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if _, exists := tm.pendingTransactions[txID]; !exists {
		return fmt.Errorf("pending transaction not found: %s", txID)
	}

	// Create cancelled result
	result := &TransactionResult{
		ID:         txID,
		Status:     StatusCancelled,
		ExecutedAt: time.Now(),
		Error:      "Transaction cancelled by user",
	}

	delete(tm.pendingTransactions, txID)
	tm.completedTransactions[txID] = result

	tm.logger.Info("Transaction cancelled", zap.String("id", txID))
	return nil
}

// RetryTransaction retries a failed transaction
func (tm *TransactionManager) RetryTransaction(ctx context.Context, txID string) (*TransactionResult, error) {
	tm.mutex.RLock()
	result, exists := tm.completedTransactions[txID]
	tm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("transaction not found: %s", txID)
	}

	if result.Status != StatusFailed {
		return nil, fmt.Errorf("transaction is not in failed state: %s", txID)
	}

	// Create new transaction request based on the failed one
	// In a real implementation, this would reconstruct the original request
	newRequest := &TransactionRequest{
		ID:              fmt.Sprintf("%s_retry_%d", txID, time.Now().Unix()),
		Chain:           "ethereum",       // Would be extracted from original
		ContractAddress: common.Address{}, // Would be extracted from original
		FunctionName:    "retry",
		CreatedAt:       time.Now(),
	}

	return tm.ExecuteTransaction(ctx, newRequest)
}

// GetTransactionMetrics returns transaction metrics
func (tm *TransactionManager) GetTransactionMetrics() map[string]interface{} {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	pendingCount := len(tm.pendingTransactions)
	completedCount := len(tm.completedTransactions)

	// Count by status
	statusCounts := make(map[TransactionStatus]int)
	for _, result := range tm.completedTransactions {
		statusCounts[result.Status]++
	}

	return map[string]interface{}{
		"pending_count":   pendingCount,
		"completed_count": completedCount,
		"total_count":     pendingCount + completedCount,
		"status_counts":   statusCounts,
		"last_updated":    time.Now(),
	}
}

// CleanupOldTransactions removes old completed transactions
func (tm *TransactionManager) CleanupOldTransactions(maxAge time.Duration) int {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	cutoff := time.Now().Add(-maxAge)
	cleaned := 0

	for id, result := range tm.completedTransactions {
		if result.ExecutedAt.Before(cutoff) {
			delete(tm.completedTransactions, id)
			cleaned++
		}
	}

	if cleaned > 0 {
		tm.logger.Info("Cleaned up old transactions", zap.Int("count", cleaned))
	}

	return cleaned
}
