package monitoring

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// GetDefaultTransactionMonitorConfig returns default transaction monitor configuration
func GetDefaultTransactionMonitorConfig() TransactionMonitorConfig {
	return TransactionMonitorConfig{
		Enabled:                true,
		UpdateInterval:         10 * time.Second,
		MaxTrackedTransactions: 1000,
		HistoryRetentionPeriod: 24 * time.Hour,
		ConfirmationConfig: ConfirmationTrackerConfig{
			Enabled:               true,
			RequiredConfirmations: 3,
			MaxConfirmationTime:   15 * time.Minute,
			ConfirmationTimeout:   1 * time.Hour,
			BlockReorgProtection:  true,
			DeepReorgThreshold:    6,
		},
		FailureConfig: FailureDetectorConfig{
			Enabled: true,
			DetectionMethods: []string{
				"timeout", "gas_limit", "nonce_conflict", "insufficient_funds",
				"revert", "out_of_gas", "invalid_transaction",
			},
			FailureThresholds: map[string]decimal.Decimal{
				"timeout_minutes":     decimal.NewFromFloat(30),
				"gas_usage_ratio":     decimal.NewFromFloat(0.95),
				"confirmation_delay":  decimal.NewFromFloat(20), // minutes
			},
			MonitoringInterval:     30 * time.Second,
			GasLimitAnalysis:       true,
			NonceConflictDetection: true,
		},
		RetryConfig: RetryManagerConfig{
			Enabled:          true,
			MaxRetryAttempts: 3,
			RetryStrategies: []string{
				"increase_gas_price", "replace_transaction", "cancel_and_retry",
			},
			BackoffStrategy:   "exponential",
			InitialRetryDelay: 1 * time.Minute,
			MaxRetryDelay:     10 * time.Minute,
			GasPriceIncrease:  decimal.NewFromFloat(1.2), // 20% increase
			AutoRetryConditions: []string{
				"timeout", "low_gas_price", "network_congestion",
			},
		},
		AlertConfig: AlertManagerConfig{
			Enabled: true,
			AlertChannels: []string{
				"log", "webhook", "email", "slack",
			},
			AlertThresholds: map[string]decimal.Decimal{
				"slow_confirmation_minutes": decimal.NewFromFloat(10),
				"high_gas_usage_ratio":      decimal.NewFromFloat(0.9),
				"failure_rate_threshold":    decimal.NewFromFloat(0.1), // 10%
			},
			NotificationDelay: 5 * time.Minute,
			AlertAggregation:  true,
			SeverityLevels:    []string{"info", "warning", "error", "critical"},
		},
	}
}

// GetHighFrequencyConfig returns configuration optimized for high-frequency monitoring
func GetHighFrequencyConfig() TransactionMonitorConfig {
	config := GetDefaultTransactionMonitorConfig()
	
	// More frequent updates
	config.UpdateInterval = 2 * time.Second
	config.MaxTrackedTransactions = 5000
	config.HistoryRetentionPeriod = 6 * time.Hour
	
	// Faster confirmation tracking
	config.ConfirmationConfig.RequiredConfirmations = 1
	config.ConfirmationConfig.MaxConfirmationTime = 5 * time.Minute
	config.ConfirmationConfig.ConfirmationTimeout = 15 * time.Minute
	
	// More aggressive failure detection
	config.FailureConfig.MonitoringInterval = 10 * time.Second
	config.FailureConfig.FailureThresholds["timeout_minutes"] = decimal.NewFromFloat(10)
	config.FailureConfig.FailureThresholds["confirmation_delay"] = decimal.NewFromFloat(5)
	
	// Faster retry strategy
	config.RetryConfig.InitialRetryDelay = 30 * time.Second
	config.RetryConfig.MaxRetryDelay = 5 * time.Minute
	config.RetryConfig.GasPriceIncrease = decimal.NewFromFloat(1.5) // 50% increase
	
	// More sensitive alerts
	config.AlertConfig.AlertThresholds["slow_confirmation_minutes"] = decimal.NewFromFloat(3)
	config.AlertConfig.NotificationDelay = 1 * time.Minute
	
	return config
}

// GetLowLatencyConfig returns configuration optimized for low latency
func GetLowLatencyConfig() TransactionMonitorConfig {
	config := GetDefaultTransactionMonitorConfig()
	
	// Very frequent updates
	config.UpdateInterval = 1 * time.Second
	config.MaxTrackedTransactions = 2000
	
	// Minimal confirmation requirements
	config.ConfirmationConfig.RequiredConfirmations = 1
	config.ConfirmationConfig.MaxConfirmationTime = 2 * time.Minute
	
	// Immediate failure detection
	config.FailureConfig.MonitoringInterval = 5 * time.Second
	config.FailureConfig.FailureThresholds["timeout_minutes"] = decimal.NewFromFloat(5)
	
	// Aggressive retry strategy
	config.RetryConfig.InitialRetryDelay = 15 * time.Second
	config.RetryConfig.GasPriceIncrease = decimal.NewFromFloat(2.0) // 100% increase
	
	// Immediate alerts
	config.AlertConfig.AlertThresholds["slow_confirmation_minutes"] = decimal.NewFromFloat(1)
	config.AlertConfig.NotificationDelay = 30 * time.Second
	
	return config
}

// GetRobustConfig returns configuration optimized for robustness
func GetRobustConfig() TransactionMonitorConfig {
	config := GetDefaultTransactionMonitorConfig()
	
	// Conservative settings
	config.UpdateInterval = 30 * time.Second
	config.MaxTrackedTransactions = 10000
	config.HistoryRetentionPeriod = 7 * 24 * time.Hour
	
	// Higher confirmation requirements
	config.ConfirmationConfig.RequiredConfirmations = 6
	config.ConfirmationConfig.MaxConfirmationTime = 1 * time.Hour
	config.ConfirmationConfig.ConfirmationTimeout = 4 * time.Hour
	config.ConfirmationConfig.DeepReorgThreshold = 12
	
	// Conservative failure detection
	config.FailureConfig.FailureThresholds["timeout_minutes"] = decimal.NewFromFloat(60)
	config.FailureConfig.FailureThresholds["confirmation_delay"] = decimal.NewFromFloat(45)
	
	// Conservative retry strategy
	config.RetryConfig.MaxRetryAttempts = 5
	config.RetryConfig.InitialRetryDelay = 5 * time.Minute
	config.RetryConfig.MaxRetryDelay = 30 * time.Minute
	config.RetryConfig.GasPriceIncrease = decimal.NewFromFloat(1.1) // 10% increase
	
	// Less sensitive alerts
	config.AlertConfig.AlertThresholds["slow_confirmation_minutes"] = decimal.NewFromFloat(30)
	config.AlertConfig.NotificationDelay = 15 * time.Minute
	
	return config
}

// ValidateTransactionMonitorConfig validates transaction monitor configuration
func ValidateTransactionMonitorConfig(config TransactionMonitorConfig) error {
	if !config.Enabled {
		return nil
	}
	
	if config.UpdateInterval <= 0 {
		return fmt.Errorf("update interval must be positive")
	}
	
	if config.MaxTrackedTransactions <= 0 {
		return fmt.Errorf("max tracked transactions must be positive")
	}
	
	if config.HistoryRetentionPeriod <= 0 {
		return fmt.Errorf("history retention period must be positive")
	}
	
	// Validate confirmation config
	if config.ConfirmationConfig.Enabled {
		if config.ConfirmationConfig.RequiredConfirmations <= 0 {
			return fmt.Errorf("required confirmations must be positive")
		}
		if config.ConfirmationConfig.MaxConfirmationTime <= 0 {
			return fmt.Errorf("max confirmation time must be positive")
		}
		if config.ConfirmationConfig.ConfirmationTimeout <= 0 {
			return fmt.Errorf("confirmation timeout must be positive")
		}
		if config.ConfirmationConfig.DeepReorgThreshold <= 0 {
			return fmt.Errorf("deep reorg threshold must be positive")
		}
	}
	
	// Validate failure config
	if config.FailureConfig.Enabled {
		if len(config.FailureConfig.DetectionMethods) == 0 {
			return fmt.Errorf("at least one failure detection method must be specified")
		}
		if config.FailureConfig.MonitoringInterval <= 0 {
			return fmt.Errorf("failure monitoring interval must be positive")
		}
		
		for threshold, value := range config.FailureConfig.FailureThresholds {
			if value.LessThan(decimal.Zero) {
				return fmt.Errorf("failure threshold %s must be non-negative", threshold)
			}
		}
	}
	
	// Validate retry config
	if config.RetryConfig.Enabled {
		if config.RetryConfig.MaxRetryAttempts <= 0 {
			return fmt.Errorf("max retry attempts must be positive")
		}
		if len(config.RetryConfig.RetryStrategies) == 0 {
			return fmt.Errorf("at least one retry strategy must be specified")
		}
		if config.RetryConfig.InitialRetryDelay <= 0 {
			return fmt.Errorf("initial retry delay must be positive")
		}
		if config.RetryConfig.MaxRetryDelay <= config.RetryConfig.InitialRetryDelay {
			return fmt.Errorf("max retry delay must be greater than initial retry delay")
		}
		if config.RetryConfig.GasPriceIncrease.LessThanOrEqual(decimal.NewFromFloat(1)) {
			return fmt.Errorf("gas price increase must be greater than 1.0")
		}
		
		validBackoffStrategies := []string{"linear", "exponential", "fixed"}
		isValidBackoff := false
		for _, strategy := range validBackoffStrategies {
			if config.RetryConfig.BackoffStrategy == strategy {
				isValidBackoff = true
				break
			}
		}
		if !isValidBackoff {
			return fmt.Errorf("invalid backoff strategy: %s", config.RetryConfig.BackoffStrategy)
		}
	}
	
	// Validate alert config
	if config.AlertConfig.Enabled {
		if len(config.AlertConfig.AlertChannels) == 0 {
			return fmt.Errorf("at least one alert channel must be specified")
		}
		if config.AlertConfig.NotificationDelay < 0 {
			return fmt.Errorf("notification delay must be non-negative")
		}
		if len(config.AlertConfig.SeverityLevels) == 0 {
			return fmt.Errorf("at least one severity level must be specified")
		}
		
		for threshold, value := range config.AlertConfig.AlertThresholds {
			if value.LessThan(decimal.Zero) {
				return fmt.Errorf("alert threshold %s must be non-negative", threshold)
			}
		}
	}
	
	return nil
}

// GetSupportedFailureDetectionMethods returns supported failure detection methods
func GetSupportedFailureDetectionMethods() []string {
	return []string{
		"timeout",
		"gas_limit",
		"nonce_conflict",
		"insufficient_funds",
		"revert",
		"out_of_gas",
		"invalid_transaction",
		"network_error",
		"rpc_error",
	}
}

// GetSupportedRetryStrategies returns supported retry strategies
func GetSupportedRetryStrategies() []string {
	return []string{
		"increase_gas_price",
		"replace_transaction",
		"cancel_and_retry",
		"speed_up",
		"resubmit",
	}
}

// GetSupportedBackoffStrategies returns supported backoff strategies
func GetSupportedBackoffStrategies() []string {
	return []string{
		"linear",
		"exponential",
		"fixed",
	}
}

// GetSupportedAlertChannels returns supported alert channels
func GetSupportedAlertChannels() []string {
	return []string{
		"log",
		"webhook",
		"email",
		"slack",
		"discord",
		"telegram",
		"sms",
		"push_notification",
	}
}

// GetSupportedSeverityLevels returns supported severity levels
func GetSupportedSeverityLevels() []string {
	return []string{
		"info",
		"warning",
		"error",
		"critical",
	}
}

// GetOptimalConfigForUseCase returns optimal configuration for specific use cases
func GetOptimalConfigForUseCase(useCase string) (TransactionMonitorConfig, error) {
	switch useCase {
	case "high_frequency":
		return GetHighFrequencyConfig(), nil
	case "low_latency":
		return GetLowLatencyConfig(), nil
	case "robust":
		return GetRobustConfig(), nil
	case "default":
		return GetDefaultTransactionMonitorConfig(), nil
	default:
		return TransactionMonitorConfig{}, fmt.Errorf("unsupported use case: %s", useCase)
	}
}

// GetTransactionStatusDescription returns descriptions for transaction statuses
func GetTransactionStatusDescription() map[TransactionStatus]string {
	return map[TransactionStatus]string{
		StatusPending:    "Transaction has been submitted and is waiting to be included in a block",
		StatusConfirming: "Transaction has been included in a block and is accumulating confirmations",
		StatusConfirmed:  "Transaction has received the required number of confirmations",
		StatusFailed:     "Transaction execution failed or was rejected by the network",
		StatusDropped:    "Transaction was dropped from the mempool without being mined",
		StatusReplaced:   "Transaction was replaced by another transaction with the same nonce",
		StatusStuck:      "Transaction appears to be stuck and may need intervention",
	}
}

// GetFailureTypeDescription returns descriptions for failure types
func GetFailureTypeDescription() map[string]string {
	return map[string]string{
		"timeout":            "Transaction timed out waiting for confirmation",
		"gas_limit":          "Transaction ran out of gas during execution",
		"nonce_conflict":     "Transaction nonce conflicts with another transaction",
		"insufficient_funds": "Insufficient funds to pay for transaction",
		"revert":             "Transaction was reverted during execution",
		"out_of_gas":         "Transaction consumed all allocated gas",
		"invalid_transaction": "Transaction is invalid or malformed",
		"network_error":      "Network error occurred during transaction processing",
		"rpc_error":          "RPC error occurred while submitting transaction",
	}
}
