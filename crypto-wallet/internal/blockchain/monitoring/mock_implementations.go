package monitoring

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
)

// Mock implementations for testing and demonstration

// MockConfirmationTracker provides mock confirmation tracking
type MockConfirmationTracker struct{}

func (m *MockConfirmationTracker) TrackConfirmations(ctx context.Context, tx *TrackedTransaction) error {
	// Mock confirmation tracking
	if tx.Status == StatusPending {
		// Simulate confirmation progress
		tx.Status = StatusConfirming
		tx.Confirmations = 1

		// Mock block information
		blockNumber := uint64(12345678)
		blockHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
		txIndex := uint(0)
		gasUsed := uint64(21000)
		effectiveGasPrice := decimal.NewFromFloat(20000000000) // 20 gwei

		tx.BlockNumber = &blockNumber
		tx.BlockHash = &blockHash
		tx.TransactionIndex = &txIndex
		tx.GasUsed = &gasUsed
		tx.EffectiveGasPrice = &effectiveGasPrice

		// Mock receipt
		tx.Receipt = &types.Receipt{
			Status:            types.ReceiptStatusSuccessful,
			CumulativeGasUsed: gasUsed,
			BlockNumber:       big.NewInt(int64(blockNumber)),
			BlockHash:         blockHash,
			TxHash:            tx.Hash,
			GasUsed:           gasUsed,
		}
	} else if tx.Status == StatusConfirming {
		// Increment confirmations
		tx.Confirmations++

		// Mark as confirmed after 3 confirmations
		if tx.Confirmations >= 3 {
			tx.Status = StatusConfirmed
			now := time.Now()
			tx.ConfirmedAt = &now
		}
	}

	return nil
}

func (m *MockConfirmationTracker) GetConfirmationStatus(hash common.Hash) (int, error) {
	// Mock confirmation count
	return 3, nil
}

func (m *MockConfirmationTracker) IsConfirmed(hash common.Hash, requiredConfirmations int) (bool, error) {
	// Mock confirmation status
	return true, nil
}

// MockFailureDetector provides mock failure detection
type MockFailureDetector struct{}

func (m *MockFailureDetector) DetectFailures(ctx context.Context, tx *TrackedTransaction) (*FailureAnalysis, error) {
	// Mock failure detection
	analysis := &FailureAnalysis{
		IsFailed:    false,
		IsRetryable: false,
	}

	// Simulate failure detection based on time
	if time.Since(tx.SubmittedAt) > 1*time.Hour && tx.Status == StatusPending {
		analysis.IsFailed = true
		analysis.FailureType = "timeout"
		analysis.FailureReason = "Transaction timed out"
		analysis.IsRetryable = true
		analysis.RecommendedAction = "Increase gas price and retry"

		// Mock gas estimates
		gasEstimate := uint64(25000)
		suggestedGasPrice := decimal.NewFromFloat(25000000000) // 25 gwei
		analysis.GasEstimate = &gasEstimate
		analysis.SuggestedGasPrice = &suggestedGasPrice
	}

	return analysis, nil
}

func (m *MockFailureDetector) AnalyzeFailureReason(receipt *types.Receipt) string {
	if receipt.Status == types.ReceiptStatusFailed {
		return "Transaction execution failed"
	}
	return ""
}

func (m *MockFailureDetector) IsTransactionStuck(tx *TrackedTransaction) bool {
	return time.Since(tx.SubmittedAt) > 30*time.Minute && tx.Status == StatusPending
}

// MockRetryManager provides mock retry management
type MockRetryManager struct{}

func (m *MockRetryManager) ShouldRetry(tx *TrackedTransaction, failure *FailureAnalysis) bool {
	// Mock retry decision
	return failure.IsRetryable && tx.RetryAttempts < 3
}

func (m *MockRetryManager) CreateRetryTransaction(tx *TrackedTransaction) (*types.Transaction, error) {
	// Mock retry transaction creation
	// In a real implementation, this would create a new transaction with higher gas price
	return nil, fmt.Errorf("mock implementation - retry transaction creation not implemented")
}

func (m *MockRetryManager) ScheduleRetry(tx *TrackedTransaction) error {
	// Mock retry scheduling
	tx.RetryAttempts++
	now := time.Now()
	tx.LastRetryAt = &now

	// Reset status to pending for retry
	tx.Status = StatusPending

	return nil
}

// MockAlertManager provides mock alert management
type MockAlertManager struct{}

func (m *MockAlertManager) CreateAlert(tx *TrackedTransaction, alertType string, severity string, message string) *TransactionAlert {
	return &TransactionAlert{
		ID:              fmt.Sprintf("alert_%d", time.Now().Unix()),
		TransactionHash: tx.Hash,
		Type:            alertType,
		Severity:        severity,
		Title:           fmt.Sprintf("%s Alert", alertType),
		Message:         message,
		Timestamp:       time.Now(),
		Acknowledged:    false,
		ActionRequired:  severity == "error" || severity == "warning",
		RecommendedActions: []string{
			"Monitor transaction status",
			"Consider increasing gas price if stuck",
		},
		Metadata: make(map[string]interface{}),
	}
}

func (m *MockAlertManager) SendAlert(alert *TransactionAlert) error {
	// Mock alert sending
	// In a real implementation, this would send alerts via configured channels
	return nil
}

func (m *MockAlertManager) AggregateAlerts(alerts []*TransactionAlert) []*TransactionAlert {
	// Mock alert aggregation
	// In a real implementation, this would group similar alerts
	return alerts
}
