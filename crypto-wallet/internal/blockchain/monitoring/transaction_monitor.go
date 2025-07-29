package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// TransactionMonitor provides comprehensive transaction tracking and monitoring
type TransactionMonitor struct {
	logger *logger.Logger
	config TransactionMonitorConfig

	// Monitoring components
	confirmationTracker ConfirmationTracker
	failureDetector     FailureDetector
	retryManager        RetryManager
	alertManager        AlertManager

	// Data storage
	trackedTransactions map[common.Hash]*TrackedTransaction
	transactionHistory  []TransactionEvent

	// State management
	isRunning    bool
	updateTicker *time.Ticker
	stopChan     chan struct{}
	mutex        sync.RWMutex
	dataMutex    sync.RWMutex
}

// TransactionMonitorConfig holds configuration for transaction monitoring
type TransactionMonitorConfig struct {
	Enabled                bool                      `json:"enabled" yaml:"enabled"`
	UpdateInterval         time.Duration             `json:"update_interval" yaml:"update_interval"`
	MaxTrackedTransactions int                       `json:"max_tracked_transactions" yaml:"max_tracked_transactions"`
	HistoryRetentionPeriod time.Duration             `json:"history_retention_period" yaml:"history_retention_period"`
	ConfirmationConfig     ConfirmationTrackerConfig `json:"confirmation_config" yaml:"confirmation_config"`
	FailureConfig          FailureDetectorConfig     `json:"failure_config" yaml:"failure_config"`
	RetryConfig            RetryManagerConfig        `json:"retry_config" yaml:"retry_config"`
	AlertConfig            AlertManagerConfig        `json:"alert_config" yaml:"alert_config"`
}

// ConfirmationTrackerConfig holds confirmation tracking configuration
type ConfirmationTrackerConfig struct {
	Enabled               bool          `json:"enabled" yaml:"enabled"`
	RequiredConfirmations int           `json:"required_confirmations" yaml:"required_confirmations"`
	MaxConfirmationTime   time.Duration `json:"max_confirmation_time" yaml:"max_confirmation_time"`
	ConfirmationTimeout   time.Duration `json:"confirmation_timeout" yaml:"confirmation_timeout"`
	BlockReorgProtection  bool          `json:"block_reorg_protection" yaml:"block_reorg_protection"`
	DeepReorgThreshold    int           `json:"deep_reorg_threshold" yaml:"deep_reorg_threshold"`
}

// FailureDetectorConfig holds failure detection configuration
type FailureDetectorConfig struct {
	Enabled                bool                       `json:"enabled" yaml:"enabled"`
	DetectionMethods       []string                   `json:"detection_methods" yaml:"detection_methods"`
	FailureThresholds      map[string]decimal.Decimal `json:"failure_thresholds" yaml:"failure_thresholds"`
	MonitoringInterval     time.Duration              `json:"monitoring_interval" yaml:"monitoring_interval"`
	GasLimitAnalysis       bool                       `json:"gas_limit_analysis" yaml:"gas_limit_analysis"`
	NonceConflictDetection bool                       `json:"nonce_conflict_detection" yaml:"nonce_conflict_detection"`
}

// RetryManagerConfig holds retry management configuration
type RetryManagerConfig struct {
	Enabled             bool            `json:"enabled" yaml:"enabled"`
	MaxRetryAttempts    int             `json:"max_retry_attempts" yaml:"max_retry_attempts"`
	RetryStrategies     []string        `json:"retry_strategies" yaml:"retry_strategies"`
	BackoffStrategy     string          `json:"backoff_strategy" yaml:"backoff_strategy"`
	InitialRetryDelay   time.Duration   `json:"initial_retry_delay" yaml:"initial_retry_delay"`
	MaxRetryDelay       time.Duration   `json:"max_retry_delay" yaml:"max_retry_delay"`
	GasPriceIncrease    decimal.Decimal `json:"gas_price_increase" yaml:"gas_price_increase"`
	AutoRetryConditions []string        `json:"auto_retry_conditions" yaml:"auto_retry_conditions"`
}

// AlertManagerConfig holds alert management configuration
type AlertManagerConfig struct {
	Enabled           bool                       `json:"enabled" yaml:"enabled"`
	AlertChannels     []string                   `json:"alert_channels" yaml:"alert_channels"`
	AlertThresholds   map[string]decimal.Decimal `json:"alert_thresholds" yaml:"alert_thresholds"`
	NotificationDelay time.Duration              `json:"notification_delay" yaml:"notification_delay"`
	AlertAggregation  bool                       `json:"alert_aggregation" yaml:"alert_aggregation"`
	SeverityLevels    []string                   `json:"severity_levels" yaml:"severity_levels"`
}

// TrackedTransaction represents a transaction being monitored
type TrackedTransaction struct {
	Hash              common.Hash            `json:"hash"`
	From              common.Address         `json:"from"`
	To                *common.Address        `json:"to"`
	Value             decimal.Decimal        `json:"value"`
	GasLimit          uint64                 `json:"gas_limit"`
	GasPrice          decimal.Decimal        `json:"gas_price"`
	GasTipCap         decimal.Decimal        `json:"gas_tip_cap"`
	GasFeeCap         decimal.Decimal        `json:"gas_fee_cap"`
	Nonce             uint64                 `json:"nonce"`
	Data              []byte                 `json:"data"`
	Status            TransactionStatus      `json:"status"`
	SubmittedAt       time.Time              `json:"submitted_at"`
	ConfirmedAt       *time.Time             `json:"confirmed_at"`
	FailedAt          *time.Time             `json:"failed_at"`
	Confirmations     int                    `json:"confirmations"`
	BlockNumber       *uint64                `json:"block_number"`
	BlockHash         *common.Hash           `json:"block_hash"`
	TransactionIndex  *uint                  `json:"transaction_index"`
	GasUsed           *uint64                `json:"gas_used"`
	EffectiveGasPrice *decimal.Decimal       `json:"effective_gas_price"`
	Receipt           *types.Receipt         `json:"receipt"`
	RetryAttempts     int                    `json:"retry_attempts"`
	LastRetryAt       *time.Time             `json:"last_retry_at"`
	FailureReason     string                 `json:"failure_reason"`
	Alerts            []*TransactionAlert    `json:"alerts"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// TransactionStatus represents the status of a tracked transaction
type TransactionStatus string

const (
	StatusPending    TransactionStatus = "pending"
	StatusConfirming TransactionStatus = "confirming"
	StatusConfirmed  TransactionStatus = "confirmed"
	StatusFailed     TransactionStatus = "failed"
	StatusDropped    TransactionStatus = "dropped"
	StatusReplaced   TransactionStatus = "replaced"
	StatusStuck      TransactionStatus = "stuck"
)

// TransactionEvent represents a transaction monitoring event
type TransactionEvent struct {
	TransactionHash common.Hash            `json:"transaction_hash"`
	EventType       string                 `json:"event_type"`
	Timestamp       time.Time              `json:"timestamp"`
	BlockNumber     *uint64                `json:"block_number"`
	Data            map[string]interface{} `json:"data"`
	Severity        string                 `json:"severity"`
}

// TransactionAlert represents a transaction alert
type TransactionAlert struct {
	ID                 string                 `json:"id"`
	TransactionHash    common.Hash            `json:"transaction_hash"`
	Type               string                 `json:"type"`
	Severity           string                 `json:"severity"`
	Title              string                 `json:"title"`
	Message            string                 `json:"message"`
	Timestamp          time.Time              `json:"timestamp"`
	Acknowledged       bool                   `json:"acknowledged"`
	ActionRequired     bool                   `json:"action_required"`
	RecommendedActions []string               `json:"recommended_actions"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// MonitoringResult represents the result of transaction monitoring
type MonitoringResult struct {
	TotalTracked       int                 `json:"total_tracked"`
	PendingCount       int                 `json:"pending_count"`
	ConfirmingCount    int                 `json:"confirming_count"`
	ConfirmedCount     int                 `json:"confirmed_count"`
	FailedCount        int                 `json:"failed_count"`
	StuckCount         int                 `json:"stuck_count"`
	AverageConfirmTime time.Duration       `json:"average_confirm_time"`
	SuccessRate        decimal.Decimal     `json:"success_rate"`
	ActiveAlerts       int                 `json:"active_alerts"`
	RecentEvents       []TransactionEvent  `json:"recent_events"`
	PerformanceMetrics *PerformanceMetrics `json:"performance_metrics"`
}

// PerformanceMetrics represents monitoring performance metrics
type PerformanceMetrics struct {
	AverageGasUsed    decimal.Decimal  `json:"average_gas_used"`
	AverageGasPrice   decimal.Decimal  `json:"average_gas_price"`
	TotalGasCost      decimal.Decimal  `json:"total_gas_cost"`
	ConfirmationTimes []time.Duration  `json:"confirmation_times"`
	FailureReasons    map[string]int   `json:"failure_reasons"`
	RetryStatistics   *RetryStatistics `json:"retry_statistics"`
}

// RetryStatistics represents retry statistics
type RetryStatistics struct {
	TotalRetries      int             `json:"total_retries"`
	SuccessfulRetries int             `json:"successful_retries"`
	FailedRetries     int             `json:"failed_retries"`
	AverageRetryTime  time.Duration   `json:"average_retry_time"`
	RetrySuccessRate  decimal.Decimal `json:"retry_success_rate"`
}

// Component interfaces
type ConfirmationTracker interface {
	TrackConfirmations(ctx context.Context, tx *TrackedTransaction) error
	GetConfirmationStatus(hash common.Hash) (int, error)
	IsConfirmed(hash common.Hash, requiredConfirmations int) (bool, error)
}

type FailureDetector interface {
	DetectFailures(ctx context.Context, tx *TrackedTransaction) (*FailureAnalysis, error)
	AnalyzeFailureReason(receipt *types.Receipt) string
	IsTransactionStuck(tx *TrackedTransaction) bool
}

type RetryManager interface {
	ShouldRetry(tx *TrackedTransaction, failure *FailureAnalysis) bool
	CreateRetryTransaction(tx *TrackedTransaction) (*types.Transaction, error)
	ScheduleRetry(tx *TrackedTransaction) error
}

type AlertManager interface {
	CreateAlert(tx *TrackedTransaction, alertType string, severity string, message string) *TransactionAlert
	SendAlert(alert *TransactionAlert) error
	AggregateAlerts(alerts []*TransactionAlert) []*TransactionAlert
}

// Supporting types
type FailureAnalysis struct {
	IsFailed          bool             `json:"is_failed"`
	FailureType       string           `json:"failure_type"`
	FailureReason     string           `json:"failure_reason"`
	IsRetryable       bool             `json:"is_retryable"`
	RecommendedAction string           `json:"recommended_action"`
	GasEstimate       *uint64          `json:"gas_estimate"`
	SuggestedGasPrice *decimal.Decimal `json:"suggested_gas_price"`
}

// NewTransactionMonitor creates a new transaction monitor
func NewTransactionMonitor(logger *logger.Logger, config TransactionMonitorConfig) *TransactionMonitor {
	tm := &TransactionMonitor{
		logger:              logger.Named("transaction-monitor"),
		config:              config,
		trackedTransactions: make(map[common.Hash]*TrackedTransaction),
		transactionHistory:  make([]TransactionEvent, 0),
		stopChan:            make(chan struct{}),
	}

	// Initialize components
	tm.confirmationTracker = &MockConfirmationTracker{}
	tm.failureDetector = &MockFailureDetector{}
	tm.retryManager = &MockRetryManager{}
	tm.alertManager = &MockAlertManager{}

	return tm
}

// Start starts the transaction monitor
func (tm *TransactionMonitor) Start(ctx context.Context) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if tm.isRunning {
		return fmt.Errorf("transaction monitor is already running")
	}

	if !tm.config.Enabled {
		tm.logger.Info("Transaction monitor is disabled")
		return nil
	}

	tm.logger.Info("Starting transaction monitor",
		zap.Duration("update_interval", tm.config.UpdateInterval),
		zap.Int("max_tracked_transactions", tm.config.MaxTrackedTransactions))

	// Start monitoring loop
	tm.updateTicker = time.NewTicker(tm.config.UpdateInterval)
	go tm.monitoringLoop(ctx)

	// Start cleanup routine
	go tm.cleanupLoop(ctx)

	tm.isRunning = true
	tm.logger.Info("Transaction monitor started successfully")
	return nil
}

// Stop stops the transaction monitor
func (tm *TransactionMonitor) Stop() error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if !tm.isRunning {
		return nil
	}

	tm.logger.Info("Stopping transaction monitor")

	// Stop monitoring
	if tm.updateTicker != nil {
		tm.updateTicker.Stop()
	}
	close(tm.stopChan)

	tm.isRunning = false
	tm.logger.Info("Transaction monitor stopped")
	return nil
}

// TrackTransaction starts tracking a transaction
func (tm *TransactionMonitor) TrackTransaction(tx *types.Transaction, metadata map[string]interface{}) error {
	tm.dataMutex.Lock()
	defer tm.dataMutex.Unlock()

	// Check if already tracking
	if _, exists := tm.trackedTransactions[tx.Hash()]; exists {
		return fmt.Errorf("transaction %s is already being tracked", tx.Hash().Hex())
	}

	// Create tracked transaction
	trackedTx := &TrackedTransaction{
		Hash:        tx.Hash(),
		From:        tm.getSenderAddress(tx),
		To:          tx.To(),
		Value:       decimal.NewFromBigInt(tx.Value(), 0),
		GasLimit:    tx.Gas(),
		Nonce:       tx.Nonce(),
		Data:        tx.Data(),
		Status:      StatusPending,
		SubmittedAt: time.Now(),
		Alerts:      make([]*TransactionAlert, 0),
		Metadata:    metadata,
	}

	// Handle different transaction types
	switch tx.Type() {
	case types.LegacyTxType:
		trackedTx.GasPrice = decimal.NewFromBigInt(tx.GasPrice(), 0)
	case types.DynamicFeeTxType:
		trackedTx.GasTipCap = decimal.NewFromBigInt(tx.GasTipCap(), 0)
		trackedTx.GasFeeCap = decimal.NewFromBigInt(tx.GasFeeCap(), 0)
	}

	// Add to tracking
	tm.trackedTransactions[tx.Hash()] = trackedTx

	// Limit tracked transactions
	if len(tm.trackedTransactions) > tm.config.MaxTrackedTransactions {
		tm.removeOldestTransactions()
	}

	// Log event
	tm.addEvent(TransactionEvent{
		TransactionHash: tx.Hash(),
		EventType:       "transaction_submitted",
		Timestamp:       time.Now(),
		Data: map[string]interface{}{
			"gas_limit": tx.Gas(),
			"gas_price": tx.GasPrice().String(),
			"value":     tx.Value().String(),
		},
		Severity: "info",
	})

	tm.logger.Info("Started tracking transaction",
		zap.String("hash", tx.Hash().Hex()),
		zap.String("status", string(trackedTx.Status)))

	return nil
}

// GetTransactionStatus returns the status of a tracked transaction
func (tm *TransactionMonitor) GetTransactionStatus(hash common.Hash) (*TrackedTransaction, error) {
	tm.dataMutex.RLock()
	defer tm.dataMutex.RUnlock()

	if tx, exists := tm.trackedTransactions[hash]; exists {
		// Return a copy
		txCopy := *tx
		return &txCopy, nil
	}

	return nil, fmt.Errorf("transaction %s is not being tracked", hash.Hex())
}

// StopTracking stops tracking a transaction
func (tm *TransactionMonitor) StopTracking(hash common.Hash) error {
	tm.dataMutex.Lock()
	defer tm.dataMutex.Unlock()

	if _, exists := tm.trackedTransactions[hash]; !exists {
		return fmt.Errorf("transaction %s is not being tracked", hash.Hex())
	}

	delete(tm.trackedTransactions, hash)

	tm.logger.Info("Stopped tracking transaction", zap.String("hash", hash.Hex()))
	return nil
}

// GetMonitoringResult returns comprehensive monitoring results
func (tm *TransactionMonitor) GetMonitoringResult() *MonitoringResult {
	tm.dataMutex.RLock()
	defer tm.dataMutex.RUnlock()

	result := &MonitoringResult{
		TotalTracked:       len(tm.trackedTransactions),
		RecentEvents:       tm.getRecentEvents(10),
		PerformanceMetrics: tm.calculatePerformanceMetrics(),
	}

	// Count by status
	for _, tx := range tm.trackedTransactions {
		switch tx.Status {
		case StatusPending:
			result.PendingCount++
		case StatusConfirming:
			result.ConfirmingCount++
		case StatusConfirmed:
			result.ConfirmedCount++
		case StatusFailed:
			result.FailedCount++
		case StatusStuck:
			result.StuckCount++
		}

		// Count active alerts
		for _, alert := range tx.Alerts {
			if !alert.Acknowledged {
				result.ActiveAlerts++
			}
		}
	}

	// Calculate success rate
	total := result.ConfirmedCount + result.FailedCount
	if total > 0 {
		result.SuccessRate = decimal.NewFromInt(int64(result.ConfirmedCount)).Div(decimal.NewFromInt(int64(total)))
	}

	// Calculate average confirmation time
	result.AverageConfirmTime = tm.calculateAverageConfirmationTime()

	return result
}

// Helper methods

// getSenderAddress gets the sender address from transaction
func (tm *TransactionMonitor) getSenderAddress(tx *types.Transaction) common.Address {
	// In a real implementation, you would recover the sender from the transaction signature
	// For this example, we'll return a zero address
	return common.Address{}
}

// removeOldestTransactions removes oldest transactions to maintain limit
func (tm *TransactionMonitor) removeOldestTransactions() {
	// Find oldest transaction
	var oldestHash common.Hash
	var oldestTime time.Time = time.Now()

	for hash, tx := range tm.trackedTransactions {
		if tx.SubmittedAt.Before(oldestTime) {
			oldestTime = tx.SubmittedAt
			oldestHash = hash
		}
	}

	if oldestHash != (common.Hash{}) {
		delete(tm.trackedTransactions, oldestHash)
	}
}

// addEvent adds a transaction event to history
func (tm *TransactionMonitor) addEvent(event TransactionEvent) {
	tm.transactionHistory = append(tm.transactionHistory, event)

	// Limit history size
	maxHistory := 1000
	if len(tm.transactionHistory) > maxHistory {
		tm.transactionHistory = tm.transactionHistory[len(tm.transactionHistory)-maxHistory:]
	}
}

// getRecentEvents returns recent transaction events
func (tm *TransactionMonitor) getRecentEvents(limit int) []TransactionEvent {
	if len(tm.transactionHistory) == 0 {
		return []TransactionEvent{}
	}

	start := len(tm.transactionHistory) - limit
	if start < 0 {
		start = 0
	}

	events := make([]TransactionEvent, len(tm.transactionHistory[start:]))
	copy(events, tm.transactionHistory[start:])
	return events
}

// calculatePerformanceMetrics calculates performance metrics
func (tm *TransactionMonitor) calculatePerformanceMetrics() *PerformanceMetrics {
	metrics := &PerformanceMetrics{
		FailureReasons:  make(map[string]int),
		RetryStatistics: &RetryStatistics{},
	}

	var totalGasUsed, totalGasPrice, totalGasCost decimal.Decimal
	var confirmationTimes []time.Duration
	var totalRetries, successfulRetries int

	for _, tx := range tm.trackedTransactions {
		// Gas metrics
		if tx.GasUsed != nil {
			totalGasUsed = totalGasUsed.Add(decimal.NewFromInt(int64(*tx.GasUsed)))
		}
		if !tx.GasPrice.IsZero() {
			totalGasPrice = totalGasPrice.Add(tx.GasPrice)
		}
		if tx.EffectiveGasPrice != nil && tx.GasUsed != nil {
			cost := tx.EffectiveGasPrice.Mul(decimal.NewFromInt(int64(*tx.GasUsed)))
			totalGasCost = totalGasCost.Add(cost)
		}

		// Confirmation times
		if tx.Status == StatusConfirmed && tx.ConfirmedAt != nil {
			confirmTime := tx.ConfirmedAt.Sub(tx.SubmittedAt)
			confirmationTimes = append(confirmationTimes, confirmTime)
		}

		// Failure reasons
		if tx.Status == StatusFailed && tx.FailureReason != "" {
			metrics.FailureReasons[tx.FailureReason]++
		}

		// Retry statistics
		totalRetries += tx.RetryAttempts
		if tx.Status == StatusConfirmed && tx.RetryAttempts > 0 {
			successfulRetries++
		}
	}

	// Calculate averages
	txCount := len(tm.trackedTransactions)
	if txCount > 0 {
		metrics.AverageGasUsed = totalGasUsed.Div(decimal.NewFromInt(int64(txCount)))
		metrics.AverageGasPrice = totalGasPrice.Div(decimal.NewFromInt(int64(txCount)))
	}
	metrics.TotalGasCost = totalGasCost
	metrics.ConfirmationTimes = confirmationTimes

	// Retry statistics
	metrics.RetryStatistics.TotalRetries = totalRetries
	metrics.RetryStatistics.SuccessfulRetries = successfulRetries
	metrics.RetryStatistics.FailedRetries = totalRetries - successfulRetries
	if totalRetries > 0 {
		metrics.RetryStatistics.RetrySuccessRate = decimal.NewFromInt(int64(successfulRetries)).Div(decimal.NewFromInt(int64(totalRetries)))
	}

	return metrics
}

// calculateAverageConfirmationTime calculates average confirmation time
func (tm *TransactionMonitor) calculateAverageConfirmationTime() time.Duration {
	var totalTime time.Duration
	var count int

	for _, tx := range tm.trackedTransactions {
		if tx.Status == StatusConfirmed && tx.ConfirmedAt != nil {
			totalTime += tx.ConfirmedAt.Sub(tx.SubmittedAt)
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return totalTime / time.Duration(count)
}

// monitoringLoop runs the main monitoring loop
func (tm *TransactionMonitor) monitoringLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-tm.stopChan:
			return
		case <-tm.updateTicker.C:
			tm.performMonitoringUpdate()
		}
	}
}

// cleanupLoop runs the cleanup loop
func (tm *TransactionMonitor) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(tm.config.HistoryRetentionPeriod / 10)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tm.stopChan:
			return
		case <-ticker.C:
			tm.cleanupOldData()
		}
	}
}

// performMonitoringUpdate performs periodic monitoring updates
func (tm *TransactionMonitor) performMonitoringUpdate() {
	tm.dataMutex.Lock()
	defer tm.dataMutex.Unlock()

	tm.logger.Debug("Performing transaction monitoring update",
		zap.Int("tracked_transactions", len(tm.trackedTransactions)))

	for hash, tx := range tm.trackedTransactions {
		// Update confirmation status
		if tm.config.ConfirmationConfig.Enabled {
			if err := tm.confirmationTracker.TrackConfirmations(context.Background(), tx); err != nil {
				tm.logger.Warn("Failed to track confirmations",
					zap.String("hash", hash.Hex()),
					zap.Error(err))
			}
		}

		// Detect failures
		if tm.config.FailureConfig.Enabled {
			if failure, err := tm.failureDetector.DetectFailures(context.Background(), tx); err == nil && failure.IsFailed {
				tm.handleTransactionFailure(tx, failure)
			}
		}

		// Check for stuck transactions
		if tm.isTransactionStuck(tx) {
			tm.handleStuckTransaction(tx)
		}

		// Generate alerts
		if tm.config.AlertConfig.Enabled {
			tm.generateAlertsForTransaction(tx)
		}
	}
}

// cleanupOldData removes old data beyond retention period
func (tm *TransactionMonitor) cleanupOldData() {
	tm.dataMutex.Lock()
	defer tm.dataMutex.Unlock()

	cutoff := time.Now().Add(-tm.config.HistoryRetentionPeriod)

	// Cleanup completed transactions
	for hash, tx := range tm.trackedTransactions {
		if (tx.Status == StatusConfirmed || tx.Status == StatusFailed) &&
			tx.SubmittedAt.Before(cutoff) {
			delete(tm.trackedTransactions, hash)
		}
	}

	// Cleanup transaction history
	var validHistory []TransactionEvent
	for _, event := range tm.transactionHistory {
		if event.Timestamp.After(cutoff) {
			validHistory = append(validHistory, event)
		}
	}
	tm.transactionHistory = validHistory
}

// handleTransactionFailure handles a failed transaction
func (tm *TransactionMonitor) handleTransactionFailure(tx *TrackedTransaction, failure *FailureAnalysis) {
	tx.Status = StatusFailed
	now := time.Now()
	tx.FailedAt = &now
	tx.FailureReason = failure.FailureReason

	// Log event
	tm.addEvent(TransactionEvent{
		TransactionHash: tx.Hash,
		EventType:       "transaction_failed",
		Timestamp:       time.Now(),
		Data: map[string]interface{}{
			"failure_reason": failure.FailureReason,
			"failure_type":   failure.FailureType,
			"is_retryable":   failure.IsRetryable,
		},
		Severity: "error",
	})

	// Check if retry is needed
	if tm.config.RetryConfig.Enabled && failure.IsRetryable {
		if tm.retryManager.ShouldRetry(tx, failure) {
			if err := tm.retryManager.ScheduleRetry(tx); err != nil {
				tm.logger.Error("Failed to schedule retry",
					zap.String("hash", tx.Hash.Hex()),
					zap.Error(err))
			}
		}
	}

	tm.logger.Warn("Transaction failed",
		zap.String("hash", tx.Hash.Hex()),
		zap.String("reason", failure.FailureReason))
}

// handleStuckTransaction handles a stuck transaction
func (tm *TransactionMonitor) handleStuckTransaction(tx *TrackedTransaction) {
	if tx.Status != StatusStuck {
		tx.Status = StatusStuck

		// Log event
		tm.addEvent(TransactionEvent{
			TransactionHash: tx.Hash,
			EventType:       "transaction_stuck",
			Timestamp:       time.Now(),
			Data: map[string]interface{}{
				"stuck_duration": time.Since(tx.SubmittedAt).String(),
			},
			Severity: "warning",
		})

		tm.logger.Warn("Transaction appears to be stuck",
			zap.String("hash", tx.Hash.Hex()),
			zap.Duration("duration", time.Since(tx.SubmittedAt)))
	}
}

// isTransactionStuck checks if a transaction is stuck
func (tm *TransactionMonitor) isTransactionStuck(tx *TrackedTransaction) bool {
	if tx.Status != StatusPending {
		return false
	}

	stuckThreshold := 30 * time.Minute // Configurable threshold
	return time.Since(tx.SubmittedAt) > stuckThreshold
}

// generateAlertsForTransaction generates alerts for a transaction
func (tm *TransactionMonitor) generateAlertsForTransaction(tx *TrackedTransaction) {
	// Check for long confirmation time
	if tx.Status == StatusPending || tx.Status == StatusConfirming {
		duration := time.Since(tx.SubmittedAt)
		if duration > tm.config.ConfirmationConfig.MaxConfirmationTime {
			alert := tm.alertManager.CreateAlert(tx, "slow_confirmation", "warning",
				fmt.Sprintf("Transaction taking longer than expected to confirm (%v)", duration))
			tx.Alerts = append(tx.Alerts, alert)
		}
	}

	// Check for high gas usage
	if tx.GasUsed != nil && tx.GasLimit > 0 {
		gasUsageRatio := decimal.NewFromInt(int64(*tx.GasUsed)).Div(decimal.NewFromInt(int64(tx.GasLimit)))
		if gasUsageRatio.GreaterThan(decimal.NewFromFloat(0.9)) {
			alert := tm.alertManager.CreateAlert(tx, "high_gas_usage", "info",
				fmt.Sprintf("Transaction used %.1f%% of gas limit", gasUsageRatio.Mul(decimal.NewFromFloat(100)).InexactFloat64()))
			tx.Alerts = append(tx.Alerts, alert)
		}
	}
}

// IsRunning returns whether the monitor is running
func (tm *TransactionMonitor) IsRunning() bool {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return tm.isRunning
}

// GetTrackedTransactions returns all tracked transactions
func (tm *TransactionMonitor) GetTrackedTransactions() map[common.Hash]*TrackedTransaction {
	tm.dataMutex.RLock()
	defer tm.dataMutex.RUnlock()

	// Return a copy
	result := make(map[common.Hash]*TrackedTransaction)
	for hash, tx := range tm.trackedTransactions {
		txCopy := *tx
		result[hash] = &txCopy
	}
	return result
}

// GetMetrics returns monitor metrics
func (tm *TransactionMonitor) GetMetrics() map[string]interface{} {
	tm.dataMutex.RLock()
	defer tm.dataMutex.RUnlock()

	result := tm.GetMonitoringResult()

	return map[string]interface{}{
		"is_running":                tm.IsRunning(),
		"total_tracked":             result.TotalTracked,
		"pending_count":             result.PendingCount,
		"confirming_count":          result.ConfirmingCount,
		"confirmed_count":           result.ConfirmedCount,
		"failed_count":              result.FailedCount,
		"stuck_count":               result.StuckCount,
		"success_rate":              result.SuccessRate.String(),
		"active_alerts":             result.ActiveAlerts,
		"average_confirm_time":      result.AverageConfirmTime.String(),
		"confirmation_enabled":      tm.config.ConfirmationConfig.Enabled,
		"failure_detection_enabled": tm.config.FailureConfig.Enabled,
		"retry_enabled":             tm.config.RetryConfig.Enabled,
		"alert_enabled":             tm.config.AlertConfig.Enabled,
	}
}
