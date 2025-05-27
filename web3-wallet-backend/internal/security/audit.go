package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// SecurityAuditor provides security auditing capabilities for DeFi operations
type SecurityAuditor struct {
	logger           *zap.Logger
	riskThresholds   RiskThresholds
	auditRules       []AuditRule
	suspiciousEvents chan SuspiciousEvent
	alertHandlers    []AlertHandler
}

// RiskThresholds defines security risk thresholds
type RiskThresholds struct {
	MaxTransactionAmount decimal.Decimal `json:"max_transaction_amount"`
	MaxSlippage          decimal.Decimal `json:"max_slippage"`
	MaxGasPrice          decimal.Decimal `json:"max_gas_price"`
	MaxDailyVolume       decimal.Decimal `json:"max_daily_volume"`
	MinLiquidity         decimal.Decimal `json:"min_liquidity"`
	MaxPriceImpact       decimal.Decimal `json:"max_price_impact"`
}

// AuditRule represents a security audit rule
type AuditRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Severity    SeverityLevel          `json:"severity"`
	Category    AuditCategory          `json:"category"`
	Condition   func(context.Context, AuditEvent) bool `json:"-"`
	Action      AuditAction            `json:"action"`
	Enabled     bool                   `json:"enabled"`
}

// SeverityLevel represents the severity of a security issue
type SeverityLevel string

const (
	SeverityLow      SeverityLevel = "low"
	SeverityMedium   SeverityLevel = "medium"
	SeverityHigh     SeverityLevel = "high"
	SeverityCritical SeverityLevel = "critical"
)

// AuditCategory represents the category of audit
type AuditCategory string

const (
	CategoryTransaction AuditCategory = "transaction"
	CategoryContract    AuditCategory = "contract"
	CategoryLiquidity   AuditCategory = "liquidity"
	CategoryArbitrage   AuditCategory = "arbitrage"
	CategoryYield       AuditCategory = "yield"
	CategoryBot         AuditCategory = "bot"
)

// AuditAction represents the action to take when a rule is triggered
type AuditAction string

const (
	ActionLog     AuditAction = "log"
	ActionAlert   AuditAction = "alert"
	ActionBlock   AuditAction = "block"
	ActionPause   AuditAction = "pause"
	ActionReject  AuditAction = "reject"
)

// AuditEvent represents an event to be audited
type AuditEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	UserID      string                 `json:"user_id"`
	BotID       string                 `json:"bot_id,omitempty"`
	Amount      decimal.Decimal        `json:"amount"`
	Token       string                 `json:"token"`
	Chain       string                 `json:"chain"`
	Protocol    string                 `json:"protocol"`
	Slippage    decimal.Decimal        `json:"slippage"`
	GasPrice    decimal.Decimal        `json:"gas_price"`
	PriceImpact decimal.Decimal        `json:"price_impact"`
	Liquidity   decimal.Decimal        `json:"liquidity"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
}

// SuspiciousEvent represents a suspicious activity
type SuspiciousEvent struct {
	ID          string        `json:"id"`
	RuleID      string        `json:"rule_id"`
	Event       AuditEvent    `json:"event"`
	Severity    SeverityLevel `json:"severity"`
	Description string        `json:"description"`
	Risk        decimal.Decimal `json:"risk"`
	Timestamp   time.Time     `json:"timestamp"`
}

// AlertHandler handles security alerts
type AlertHandler interface {
	HandleAlert(ctx context.Context, event SuspiciousEvent) error
}

// NewSecurityAuditor creates a new security auditor
func NewSecurityAuditor(logger *zap.Logger) *SecurityAuditor {
	auditor := &SecurityAuditor{
		logger: logger,
		riskThresholds: RiskThresholds{
			MaxTransactionAmount: decimal.NewFromFloat(100000),  // $100k
			MaxSlippage:          decimal.NewFromFloat(0.05),    // 5%
			MaxGasPrice:          decimal.NewFromFloat(100),     // 100 gwei
			MaxDailyVolume:       decimal.NewFromFloat(1000000), // $1M
			MinLiquidity:         decimal.NewFromFloat(10000),   // $10k
			MaxPriceImpact:       decimal.NewFromFloat(0.03),    // 3%
		},
		auditRules:       []AuditRule{},
		suspiciousEvents: make(chan SuspiciousEvent, 1000),
		alertHandlers:    []AlertHandler{},
	}

	auditor.initializeDefaultRules()
	return auditor
}

// initializeDefaultRules sets up default security rules
func (sa *SecurityAuditor) initializeDefaultRules() {
	// High transaction amount rule
	sa.auditRules = append(sa.auditRules, AuditRule{
		ID:          "high_transaction_amount",
		Name:        "High Transaction Amount",
		Description: "Detects transactions above threshold",
		Severity:    SeverityHigh,
		Category:    CategoryTransaction,
		Condition: func(ctx context.Context, event AuditEvent) bool {
			return event.Amount.GreaterThan(sa.riskThresholds.MaxTransactionAmount)
		},
		Action:  ActionAlert,
		Enabled: true,
	})

	// High slippage rule
	sa.auditRules = append(sa.auditRules, AuditRule{
		ID:          "high_slippage",
		Name:        "High Slippage",
		Description: "Detects transactions with high slippage",
		Severity:    SeverityMedium,
		Category:    CategoryTransaction,
		Condition: func(ctx context.Context, event AuditEvent) bool {
			return event.Slippage.GreaterThan(sa.riskThresholds.MaxSlippage)
		},
		Action:  ActionAlert,
		Enabled: true,
	})

	// Low liquidity rule
	sa.auditRules = append(sa.auditRules, AuditRule{
		ID:          "low_liquidity",
		Name:        "Low Liquidity",
		Description: "Detects operations in low liquidity pools",
		Severity:    SeverityHigh,
		Category:    CategoryLiquidity,
		Condition: func(ctx context.Context, event AuditEvent) bool {
			return event.Liquidity.LessThan(sa.riskThresholds.MinLiquidity)
		},
		Action:  ActionBlock,
		Enabled: true,
	})

	// Suspicious contract interaction
	sa.auditRules = append(sa.auditRules, AuditRule{
		ID:          "suspicious_contract",
		Name:        "Suspicious Contract",
		Description: "Detects interactions with unverified contracts",
		Severity:    SeverityCritical,
		Category:    CategoryContract,
		Condition: func(ctx context.Context, event AuditEvent) bool {
			return sa.isSuspiciousContract(event.Protocol)
		},
		Action:  ActionReject,
		Enabled: true,
	})

	// MEV attack detection
	sa.auditRules = append(sa.auditRules, AuditRule{
		ID:          "mev_attack",
		Name:        "MEV Attack",
		Description: "Detects potential MEV attacks",
		Severity:    SeverityCritical,
		Category:    CategoryArbitrage,
		Condition: func(ctx context.Context, event AuditEvent) bool {
			return sa.detectMEVAttack(event)
		},
		Action:  ActionBlock,
		Enabled: true,
	})
}

// AuditEvent audits a single event
func (sa *SecurityAuditor) AuditEvent(ctx context.Context, event AuditEvent) (*AuditResult, error) {
	result := &AuditResult{
		EventID:   event.ID,
		Passed:    true,
		Violations: []RuleViolation{},
		Risk:      decimal.Zero,
		Timestamp: time.Now(),
	}

	// Apply all enabled rules
	for _, rule := range sa.auditRules {
		if !rule.Enabled {
			continue
		}

		if rule.Condition(ctx, event) {
			violation := RuleViolation{
				RuleID:      rule.ID,
				RuleName:    rule.Name,
				Severity:    rule.Severity,
				Description: rule.Description,
				Action:      rule.Action,
			}

			result.Violations = append(result.Violations, violation)
			result.Risk = result.Risk.Add(sa.calculateRiskScore(rule.Severity))

			// Handle critical violations
			if rule.Severity == SeverityCritical || rule.Action == ActionReject || rule.Action == ActionBlock {
				result.Passed = false
			}

			// Create suspicious event
			suspiciousEvent := SuspiciousEvent{
				ID:          sa.generateEventID(),
				RuleID:      rule.ID,
				Event:       event,
				Severity:    rule.Severity,
				Description: fmt.Sprintf("Rule '%s' triggered: %s", rule.Name, rule.Description),
				Risk:        sa.calculateRiskScore(rule.Severity),
				Timestamp:   time.Now(),
			}

			// Send to monitoring channel
			select {
			case sa.suspiciousEvents <- suspiciousEvent:
			default:
				sa.logger.Warn("Suspicious events channel full, dropping event")
			}

			// Handle alerts
			if rule.Action == ActionAlert || rule.Severity == SeverityHigh || rule.Severity == SeverityCritical {
				sa.handleAlert(ctx, suspiciousEvent)
			}
		}
	}

	sa.logger.Info("Security audit completed",
		zap.String("event_id", event.ID),
		zap.Bool("passed", result.Passed),
		zap.Int("violations", len(result.Violations)),
		zap.String("risk", result.Risk.String()),
	)

	return result, nil
}

// AuditResult represents the result of a security audit
type AuditResult struct {
	EventID    string          `json:"event_id"`
	Passed     bool            `json:"passed"`
	Violations []RuleViolation `json:"violations"`
	Risk       decimal.Decimal `json:"risk"`
	Timestamp  time.Time       `json:"timestamp"`
}

// RuleViolation represents a violated security rule
type RuleViolation struct {
	RuleID      string        `json:"rule_id"`
	RuleName    string        `json:"rule_name"`
	Severity    SeverityLevel `json:"severity"`
	Description string        `json:"description"`
	Action      AuditAction   `json:"action"`
}

// calculateRiskScore calculates risk score based on severity
func (sa *SecurityAuditor) calculateRiskScore(severity SeverityLevel) decimal.Decimal {
	switch severity {
	case SeverityLow:
		return decimal.NewFromFloat(1.0)
	case SeverityMedium:
		return decimal.NewFromFloat(3.0)
	case SeverityHigh:
		return decimal.NewFromFloat(7.0)
	case SeverityCritical:
		return decimal.NewFromFloat(10.0)
	default:
		return decimal.NewFromFloat(1.0)
	}
}

// isSuspiciousContract checks if a contract is suspicious
func (sa *SecurityAuditor) isSuspiciousContract(protocol string) bool {
	// List of known malicious or unverified contracts
	suspiciousContracts := []string{
		"unknown_protocol",
		"unverified_contract",
		"suspicious_dex",
	}

	protocol = strings.ToLower(protocol)
	for _, suspicious := range suspiciousContracts {
		if strings.Contains(protocol, suspicious) {
			return true
		}
	}

	return false
}

// detectMEVAttack detects potential MEV attacks
func (sa *SecurityAuditor) detectMEVAttack(event AuditEvent) bool {
	// Check for sandwich attack patterns
	if event.PriceImpact.GreaterThan(decimal.NewFromFloat(0.05)) && // High price impact
		event.Amount.GreaterThan(decimal.NewFromFloat(50000)) { // Large amount
		return true
	}

	// Check for front-running patterns
	if event.GasPrice.GreaterThan(sa.riskThresholds.MaxGasPrice) {
		return true
	}

	return false
}

// generateEventID generates a unique event ID
func (sa *SecurityAuditor) generateEventID() string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return hex.EncodeToString(hash[:])[:16]
}

// handleAlert handles security alerts
func (sa *SecurityAuditor) handleAlert(ctx context.Context, event SuspiciousEvent) {
	for _, handler := range sa.alertHandlers {
		if err := handler.HandleAlert(ctx, event); err != nil {
			sa.logger.Error("Failed to handle alert",
				zap.String("event_id", event.ID),
				zap.Error(err),
			)
		}
	}
}

// AddAlertHandler adds an alert handler
func (sa *SecurityAuditor) AddAlertHandler(handler AlertHandler) {
	sa.alertHandlers = append(sa.alertHandlers, handler)
}

// GetSuspiciousEvents returns the channel for suspicious events
func (sa *SecurityAuditor) GetSuspiciousEvents() <-chan SuspiciousEvent {
	return sa.suspiciousEvents
}

// UpdateRiskThresholds updates risk thresholds
func (sa *SecurityAuditor) UpdateRiskThresholds(thresholds RiskThresholds) {
	sa.riskThresholds = thresholds
	sa.logger.Info("Risk thresholds updated",
		zap.String("max_transaction", thresholds.MaxTransactionAmount.String()),
		zap.String("max_slippage", thresholds.MaxSlippage.String()),
	)
}

// AddCustomRule adds a custom audit rule
func (sa *SecurityAuditor) AddCustomRule(rule AuditRule) {
	sa.auditRules = append(sa.auditRules, rule)
	sa.logger.Info("Custom audit rule added",
		zap.String("rule_id", rule.ID),
		zap.String("rule_name", rule.Name),
	)
}

// GetAuditRules returns all audit rules
func (sa *SecurityAuditor) GetAuditRules() []AuditRule {
	return sa.auditRules
}

// EnableRule enables/disables a specific rule
func (sa *SecurityAuditor) EnableRule(ruleID string, enabled bool) error {
	for i, rule := range sa.auditRules {
		if rule.ID == ruleID {
			sa.auditRules[i].Enabled = enabled
			sa.logger.Info("Audit rule updated",
				zap.String("rule_id", ruleID),
				zap.Bool("enabled", enabled),
			)
			return nil
		}
	}
	return fmt.Errorf("rule not found: %s", ruleID)
}
