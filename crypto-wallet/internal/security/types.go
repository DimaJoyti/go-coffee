package security

import (
	"time"

	"github.com/shopspring/decimal"
)

// ArbitrageValidationRequest represents a request to validate an arbitrage transaction
type ArbitrageValidationRequest struct {
	UserID         string          `json:"user_id"`
	Token          string          `json:"token"`
	Amount         decimal.Decimal `json:"amount"`
	Chain          string          `json:"chain"`
	Protocol       string          `json:"protocol"`
	SourceExchange string          `json:"source_exchange"`
	TargetExchange string          `json:"target_exchange"`
	ProfitMargin   decimal.Decimal `json:"profit_margin"`
	Slippage       decimal.Decimal `json:"slippage"`
	GasPrice       decimal.Decimal `json:"gas_price"`
	PriceImpact    decimal.Decimal `json:"price_impact"`
	Liquidity      decimal.Decimal `json:"liquidity"`
	Timestamp      time.Time       `json:"timestamp"`
}

// YieldFarmingValidationRequest represents a request to validate a yield farming transaction
type YieldFarmingValidationRequest struct {
	UserID           string          `json:"user_id"`
	Token            string          `json:"token"`
	Amount           decimal.Decimal `json:"amount"`
	Chain            string          `json:"chain"`
	Protocol         string          `json:"protocol"`
	PoolAddress      string          `json:"pool_address"`
	ExpectedAPY      decimal.Decimal `json:"expected_apy"`
	LockPeriod       time.Duration   `json:"lock_period"`
	ImpermanentLoss  decimal.Decimal `json:"impermanent_loss"`
	Slippage         decimal.Decimal `json:"slippage"`
	GasPrice         decimal.Decimal `json:"gas_price"`
	PriceImpact      decimal.Decimal `json:"price_impact"`
	Liquidity        decimal.Decimal `json:"liquidity"`
	Timestamp        time.Time       `json:"timestamp"`
}

// TradingBotValidationRequest represents a request to validate a trading bot operation
type TradingBotValidationRequest struct {
	UserID          string          `json:"user_id"`
	BotID           string          `json:"bot_id"`
	Operation       string          `json:"operation"` // start, stop, pause, resume, update
	Strategy        string          `json:"strategy"`  // arbitrage, yield_farming, dca, grid_trading
	Token           string          `json:"token"`
	Amount          decimal.Decimal `json:"amount"`
	Chain           string          `json:"chain"`
	Protocol        string          `json:"protocol"`
	PositionSize    decimal.Decimal `json:"position_size"`
	RiskLevel       string          `json:"risk_level"` // low, medium, high
	MaxDailyTrades  int             `json:"max_daily_trades"`
	Timestamp       time.Time       `json:"timestamp"`
}

// ValidationResponse represents the response from a security validation
type ValidationResponse struct {
	Valid       bool            `json:"valid"`
	EventID     string          `json:"event_id"`
	Risk        decimal.Decimal `json:"risk"`
	Violations  []RuleViolation `json:"violations,omitempty"`
	Message     string          `json:"message"`
	Timestamp   time.Time       `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SecurityMetricsResponse represents security metrics response
type SecurityMetricsResponse struct {
	TotalEvents       int64                      `json:"total_events"`
	SuspiciousEvents  int64                      `json:"suspicious_events"`
	BlockedEvents     int64                      `json:"blocked_events"`
	AlertsSent        int64                      `json:"alerts_sent"`
	AverageRisk       decimal.Decimal            `json:"average_risk"`
	LastUpdate        time.Time                  `json:"last_update"`
	EventsByCategory  map[AuditCategory]int64    `json:"events_by_category"`
	EventsBySeverity  map[SeverityLevel]int64    `json:"events_by_severity"`
}

// ContractValidationRequest represents a request to validate a smart contract
type ContractValidationRequest struct {
	ContractAddress string            `json:"contract_address"`
	Chain           string            `json:"chain"`
	Protocol        string            `json:"protocol"`
	ByteCode        string            `json:"byte_code,omitempty"`
	ABI             string            `json:"abi,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Timestamp       time.Time         `json:"timestamp"`
}

// ContractValidationResponse represents the response from contract validation
type ContractValidationResponse struct {
	Valid           bool                  `json:"valid"`
	Verified        bool                  `json:"verified"`
	RiskScore       decimal.Decimal       `json:"risk_score"`
	SecurityIssues  []SecurityIssue       `json:"security_issues,omitempty"`
	Recommendations []string              `json:"recommendations,omitempty"`
	Timestamp       time.Time             `json:"timestamp"`
}

// SecurityIssue represents a security issue found in a contract
type SecurityIssue struct {
	Type        string        `json:"type"`
	Severity    SeverityLevel `json:"severity"`
	Description string        `json:"description"`
	Location    string        `json:"location,omitempty"`
	Suggestion  string        `json:"suggestion,omitempty"`
}

// TransactionMonitoringRequest represents a request to monitor a transaction
type TransactionMonitoringRequest struct {
	TransactionHash string            `json:"transaction_hash"`
	Chain           string            `json:"chain"`
	UserID          string            `json:"user_id"`
	BotID           string            `json:"bot_id,omitempty"`
	Type            string            `json:"type"` // arbitrage, yield_farming, trading_bot
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Timestamp       time.Time         `json:"timestamp"`
}

// TransactionMonitoringResponse represents the response from transaction monitoring
type TransactionMonitoringResponse struct {
	Status          string          `json:"status"` // pending, confirmed, failed
	Confirmations   int             `json:"confirmations"`
	GasUsed         decimal.Decimal `json:"gas_used"`
	GasPrice        decimal.Decimal `json:"gas_price"`
	Success         bool            `json:"success"`
	SecurityAlerts  []SecurityAlert `json:"security_alerts,omitempty"`
	Timestamp       time.Time       `json:"timestamp"`
}

// SecurityAlert represents a security alert
type SecurityAlert struct {
	ID          string        `json:"id"`
	Type        string        `json:"type"`
	Severity    SeverityLevel `json:"severity"`
	Message     string        `json:"message"`
	Action      string        `json:"action"`
	Timestamp   time.Time     `json:"timestamp"`
}

// RiskAssessmentRequest represents a request for risk assessment
type RiskAssessmentRequest struct {
	UserID      string                 `json:"user_id"`
	Operation   string                 `json:"operation"`
	Amount      decimal.Decimal        `json:"amount"`
	Token       string                 `json:"token"`
	Chain       string                 `json:"chain"`
	Protocol    string                 `json:"protocol"`
	Parameters  map[string]interface{} `json:"parameters"`
	Timestamp   time.Time              `json:"timestamp"`
}

// RiskAssessmentResponse represents the response from risk assessment
type RiskAssessmentResponse struct {
	RiskScore       decimal.Decimal   `json:"risk_score"`
	RiskLevel       string            `json:"risk_level"` // low, medium, high, critical
	Recommendation  string            `json:"recommendation"`
	Factors         []RiskFactor      `json:"factors"`
	MaxAmount       decimal.Decimal   `json:"max_amount,omitempty"`
	Warnings        []string          `json:"warnings,omitempty"`
	Timestamp       time.Time         `json:"timestamp"`
}

// RiskFactor represents a factor contributing to risk
type RiskFactor struct {
	Name        string          `json:"name"`
	Impact      decimal.Decimal `json:"impact"`
	Description string          `json:"description"`
	Weight      decimal.Decimal `json:"weight"`
}

// ComplianceCheckRequest represents a request for compliance checking
type ComplianceCheckRequest struct {
	UserID        string                 `json:"user_id"`
	Operation     string                 `json:"operation"`
	Amount        decimal.Decimal        `json:"amount"`
	Token         string                 `json:"token"`
	Chain         string                 `json:"chain"`
	Country       string                 `json:"country,omitempty"`
	KYCLevel      string                 `json:"kyc_level,omitempty"`
	Parameters    map[string]interface{} `json:"parameters"`
	Timestamp     time.Time              `json:"timestamp"`
}

// ComplianceCheckResponse represents the response from compliance checking
type ComplianceCheckResponse struct {
	Compliant       bool              `json:"compliant"`
	Violations      []ComplianceViolation `json:"violations,omitempty"`
	RequiredActions []string          `json:"required_actions,omitempty"`
	Restrictions    []string          `json:"restrictions,omitempty"`
	Timestamp       time.Time         `json:"timestamp"`
}

// ComplianceViolation represents a compliance violation
type ComplianceViolation struct {
	Rule        string        `json:"rule"`
	Severity    SeverityLevel `json:"severity"`
	Description string        `json:"description"`
	Action      string        `json:"action"`
}

// AuditLogRequest represents a request to retrieve audit logs
type AuditLogRequest struct {
	UserID      string    `json:"user_id,omitempty"`
	BotID       string    `json:"bot_id,omitempty"`
	EventType   string    `json:"event_type,omitempty"`
	Severity    string    `json:"severity,omitempty"`
	StartTime   time.Time `json:"start_time,omitempty"`
	EndTime     time.Time `json:"end_time,omitempty"`
	Limit       int       `json:"limit,omitempty"`
	Offset      int       `json:"offset,omitempty"`
}

// AuditLogResponse represents the response containing audit logs
type AuditLogResponse struct {
	Events     []AuditLogEntry `json:"events"`
	Total      int             `json:"total"`
	Limit      int             `json:"limit"`
	Offset     int             `json:"offset"`
	Timestamp  time.Time       `json:"timestamp"`
}

// AuditLogEntry represents an audit log entry
type AuditLogEntry struct {
	ID          string                 `json:"id"`
	EventType   string                 `json:"event_type"`
	UserID      string                 `json:"user_id"`
	BotID       string                 `json:"bot_id,omitempty"`
	Severity    SeverityLevel          `json:"severity"`
	Description string                 `json:"description"`
	Result      string                 `json:"result"`
	Risk        decimal.Decimal        `json:"risk"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
}

// SecurityConfigRequest represents a request to update security configuration
type SecurityConfigRequest struct {
	RiskThresholds  *RiskThresholds       `json:"risk_thresholds,omitempty"`
	AuditRules      []AuditRuleConfig     `json:"audit_rules,omitempty"`
	AlertSettings   *AlertSettings        `json:"alert_settings,omitempty"`
	Timestamp       time.Time             `json:"timestamp"`
}

// AuditRuleConfig represents audit rule configuration
type AuditRuleConfig struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Enabled     bool          `json:"enabled"`
	Severity    SeverityLevel `json:"severity"`
	Action      AuditAction   `json:"action"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// AlertSettings represents alert configuration settings
type AlertSettings struct {
	EmailEnabled    bool     `json:"email_enabled"`
	SlackEnabled    bool     `json:"slack_enabled"`
	PagerDutyEnabled bool    `json:"pagerduty_enabled"`
	Recipients      []string `json:"recipients,omitempty"`
	Channels        []string `json:"channels,omitempty"`
	Thresholds      map[SeverityLevel]bool `json:"thresholds,omitempty"`
}

// SecurityConfigResponse represents the response from security configuration update
type SecurityConfigResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
