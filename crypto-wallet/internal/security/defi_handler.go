package security

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/defi"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
)

// DeFiSecurityHandler handles security for DeFi operations
type DeFiSecurityHandler struct {
	auditor         *SecurityAuditor
	monitor         *SecurityMonitor
	logger          *zap.Logger
	defiService     *defi.Service
	riskThresholds  RiskThresholds
}

// NewDeFiSecurityHandler creates a new DeFi security handler
func NewDeFiSecurityHandler(
	auditor *SecurityAuditor,
	monitor *SecurityMonitor,
	logger *zap.Logger,
	defiService *defi.Service,
) *DeFiSecurityHandler {
	return &DeFiSecurityHandler{
		auditor:     auditor,
		monitor:     monitor,
		logger:      logger,
		defiService: defiService,
		riskThresholds: RiskThresholds{
			MaxTransactionAmount: decimal.NewFromFloat(100000),  // $100k
			MaxSlippage:          decimal.NewFromFloat(0.05),    // 5%
			MaxGasPrice:          decimal.NewFromFloat(100),     // 100 gwei
			MaxDailyVolume:       decimal.NewFromFloat(1000000), // $1M
			MinLiquidity:         decimal.NewFromFloat(10000),   // $10k
			MaxPriceImpact:       decimal.NewFromFloat(0.03),    // 3%
		},
	}
}

// ValidateArbitrageTransaction validates an arbitrage transaction
func (h *DeFiSecurityHandler) ValidateArbitrageTransaction(ctx context.Context, req *ArbitrageValidationRequest) (*ValidationResponse, error) {
	h.logger.Info("Validating arbitrage transaction",
		zap.String("user_id", req.UserID),
		zap.String("token", req.Token),
		zap.String("amount", req.Amount.String()),
	)

	// Create audit event
	auditEvent := AuditEvent{
		ID:          generateEventID(),
		Type:        "arbitrage_transaction",
		UserID:      req.UserID,
		Amount:      req.Amount,
		Token:       req.Token,
		Chain:       req.Chain,
		Protocol:    req.Protocol,
		Slippage:    req.Slippage,
		GasPrice:    req.GasPrice,
		PriceImpact: req.PriceImpact,
		Liquidity:   req.Liquidity,
		Metadata: map[string]interface{}{
			"source_exchange": req.SourceExchange,
			"target_exchange": req.TargetExchange,
			"profit_margin":   req.ProfitMargin.String(),
		},
		Timestamp: time.Now(),
	}

	// Perform security audit
	auditResult, err := h.auditor.AuditEvent(ctx, auditEvent)
	if err != nil {
		h.logger.Error("Failed to audit arbitrage transaction", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "security audit failed: %v", err)
	}

	// Check if transaction passed security checks
	if !auditResult.Passed {
		h.logger.Warn("Arbitrage transaction failed security audit",
			zap.String("event_id", auditEvent.ID),
			zap.Int("violations", len(auditResult.Violations)),
		)

		return &ValidationResponse{
			Valid:       false,
			EventID:     auditEvent.ID,
			Risk:        auditResult.Risk,
			Violations:  auditResult.Violations,
			Message:     "Transaction blocked due to security violations",
			Timestamp:   time.Now(),
		}, nil
	}

	// Additional DeFi-specific validations
	if err := h.validateArbitrageSpecific(req); err != nil {
		h.logger.Warn("Arbitrage-specific validation failed", zap.Error(err))
		return &ValidationResponse{
			Valid:     false,
			EventID:   auditEvent.ID,
			Risk:      decimal.NewFromFloat(5.0),
			Message:   fmt.Sprintf("Arbitrage validation failed: %v", err),
			Timestamp: time.Now(),
		}, nil
	}

	h.logger.Info("Arbitrage transaction validated successfully",
		zap.String("event_id", auditEvent.ID),
		zap.String("risk", auditResult.Risk.String()),
	)

	return &ValidationResponse{
		Valid:     true,
		EventID:   auditEvent.ID,
		Risk:      auditResult.Risk,
		Message:   "Transaction approved",
		Timestamp: time.Now(),
	}, nil
}

// ValidateYieldFarmingTransaction validates a yield farming transaction
func (h *DeFiSecurityHandler) ValidateYieldFarmingTransaction(ctx context.Context, req *YieldFarmingValidationRequest) (*ValidationResponse, error) {
	h.logger.Info("Validating yield farming transaction",
		zap.String("user_id", req.UserID),
		zap.String("protocol", req.Protocol),
		zap.String("amount", req.Amount.String()),
	)

	// Create audit event
	auditEvent := AuditEvent{
		ID:          generateEventID(),
		Type:        "yield_farming_transaction",
		UserID:      req.UserID,
		Amount:      req.Amount,
		Token:       req.Token,
		Chain:       req.Chain,
		Protocol:    req.Protocol,
		Slippage:    req.Slippage,
		GasPrice:    req.GasPrice,
		PriceImpact: req.PriceImpact,
		Liquidity:   req.Liquidity,
		Metadata: map[string]interface{}{
			"pool_address":      req.PoolAddress,
			"expected_apy":      req.ExpectedAPY.String(),
			"lock_period":       req.LockPeriod.String(),
			"impermanent_loss":  req.ImpermanentLoss.String(),
		},
		Timestamp: time.Now(),
	}

	// Perform security audit
	auditResult, err := h.auditor.AuditEvent(ctx, auditEvent)
	if err != nil {
		h.logger.Error("Failed to audit yield farming transaction", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "security audit failed: %v", err)
	}

	if !auditResult.Passed {
		return &ValidationResponse{
			Valid:       false,
			EventID:     auditEvent.ID,
			Risk:        auditResult.Risk,
			Violations:  auditResult.Violations,
			Message:     "Transaction blocked due to security violations",
			Timestamp:   time.Now(),
		}, nil
	}

	// Additional yield farming validations
	if err := h.validateYieldFarmingSpecific(req); err != nil {
		return &ValidationResponse{
			Valid:     false,
			EventID:   auditEvent.ID,
			Risk:      decimal.NewFromFloat(5.0),
			Message:   fmt.Sprintf("Yield farming validation failed: %v", err),
			Timestamp: time.Now(),
		}, nil
	}

	return &ValidationResponse{
		Valid:     true,
		EventID:   auditEvent.ID,
		Risk:      auditResult.Risk,
		Message:   "Transaction approved",
		Timestamp: time.Now(),
	}, nil
}

// ValidateTradingBotOperation validates trading bot operations
func (h *DeFiSecurityHandler) ValidateTradingBotOperation(ctx context.Context, req *TradingBotValidationRequest) (*ValidationResponse, error) {
	h.logger.Info("Validating trading bot operation",
		zap.String("user_id", req.UserID),
		zap.String("bot_id", req.BotID),
		zap.String("operation", req.Operation),
	)

	// Create audit event
	auditEvent := AuditEvent{
		ID:       generateEventID(),
		Type:     "trading_bot_operation",
		UserID:   req.UserID,
		BotID:    req.BotID,
		Amount:   req.Amount,
		Token:    req.Token,
		Chain:    req.Chain,
		Protocol: req.Protocol,
		Metadata: map[string]interface{}{
			"operation":        req.Operation,
			"strategy":         req.Strategy,
			"position_size":    req.PositionSize.String(),
			"risk_level":       req.RiskLevel,
			"max_daily_trades": req.MaxDailyTrades,
		},
		Timestamp: time.Now(),
	}

	// Perform security audit
	auditResult, err := h.auditor.AuditEvent(ctx, auditEvent)
	if err != nil {
		h.logger.Error("Failed to audit trading bot operation", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "security audit failed: %v", err)
	}

	if !auditResult.Passed {
		return &ValidationResponse{
			Valid:       false,
			EventID:     auditEvent.ID,
			Risk:        auditResult.Risk,
			Violations:  auditResult.Violations,
			Message:     "Operation blocked due to security violations",
			Timestamp:   time.Now(),
		}, nil
	}

	// Additional trading bot validations
	if err := h.validateTradingBotSpecific(req); err != nil {
		return &ValidationResponse{
			Valid:     false,
			EventID:   auditEvent.ID,
			Risk:      decimal.NewFromFloat(5.0),
			Message:   fmt.Sprintf("Trading bot validation failed: %v", err),
			Timestamp: time.Now(),
		}, nil
	}

	return &ValidationResponse{
		Valid:     true,
		EventID:   auditEvent.ID,
		Risk:      auditResult.Risk,
		Message:   "Operation approved",
		Timestamp: time.Now(),
	}, nil
}

// validateArbitrageSpecific performs arbitrage-specific validations
func (h *DeFiSecurityHandler) validateArbitrageSpecific(req *ArbitrageValidationRequest) error {
	// Check minimum profit margin
	if req.ProfitMargin.LessThan(decimal.NewFromFloat(0.005)) { // 0.5%
		return fmt.Errorf("profit margin too low: %s", req.ProfitMargin.String())
	}

	// Check maximum slippage
	if req.Slippage.GreaterThan(h.riskThresholds.MaxSlippage) {
		return fmt.Errorf("slippage too high: %s", req.Slippage.String())
	}

	// Check liquidity
	if req.Liquidity.LessThan(h.riskThresholds.MinLiquidity) {
		return fmt.Errorf("liquidity too low: %s", req.Liquidity.String())
	}

	// Check price impact
	if req.PriceImpact.GreaterThan(h.riskThresholds.MaxPriceImpact) {
		return fmt.Errorf("price impact too high: %s", req.PriceImpact.String())
	}

	return nil
}

// validateYieldFarmingSpecific performs yield farming-specific validations
func (h *DeFiSecurityHandler) validateYieldFarmingSpecific(req *YieldFarmingValidationRequest) error {
	// Check minimum APY
	if req.ExpectedAPY.LessThan(decimal.NewFromFloat(0.01)) { // 1%
		return fmt.Errorf("APY too low: %s", req.ExpectedAPY.String())
	}

	// Check maximum APY (suspicious if too high)
	if req.ExpectedAPY.GreaterThan(decimal.NewFromFloat(2.0)) { // 200%
		return fmt.Errorf("APY suspiciously high: %s", req.ExpectedAPY.String())
	}

	// Check impermanent loss
	if req.ImpermanentLoss.GreaterThan(decimal.NewFromFloat(0.20)) { // 20%
		return fmt.Errorf("impermanent loss too high: %s", req.ImpermanentLoss.String())
	}

	// Check lock period (max 1 year)
	if req.LockPeriod > time.Hour*24*365 {
		return fmt.Errorf("lock period too long: %s", req.LockPeriod.String())
	}

	return nil
}

// validateTradingBotSpecific performs trading bot-specific validations
func (h *DeFiSecurityHandler) validateTradingBotSpecific(req *TradingBotValidationRequest) error {
	// Check position size
	if req.PositionSize.GreaterThan(h.riskThresholds.MaxTransactionAmount) {
		return fmt.Errorf("position size too large: %s", req.PositionSize.String())
	}

	// Check daily trades limit
	if req.MaxDailyTrades > 100 {
		return fmt.Errorf("daily trades limit too high: %d", req.MaxDailyTrades)
	}

	// Check risk level
	validRiskLevels := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
	}
	if !validRiskLevels[req.RiskLevel] {
		return fmt.Errorf("invalid risk level: %s", req.RiskLevel)
	}

	return nil
}

// GetSecurityMetrics returns current security metrics
func (h *DeFiSecurityHandler) GetSecurityMetrics(ctx context.Context) (*SecurityMetricsResponse, error) {
	metrics := h.monitor.GetMetrics()

	return &SecurityMetricsResponse{
		TotalEvents:      metrics.TotalEvents,
		SuspiciousEvents: metrics.SuspiciousEvents,
		BlockedEvents:    metrics.BlockedEvents,
		AlertsSent:       metrics.AlertsSent,
		AverageRisk:      metrics.AverageRisk,
		LastUpdate:       metrics.LastUpdate,
		EventsByCategory: metrics.EventsByCategory,
		EventsBySeverity: metrics.EventsBySeverity,
	}, nil
}

// UpdateRiskThresholds updates security risk thresholds
func (h *DeFiSecurityHandler) UpdateRiskThresholds(ctx context.Context, thresholds RiskThresholds) error {
	h.riskThresholds = thresholds
	h.auditor.UpdateRiskThresholds(thresholds)

	h.logger.Info("Risk thresholds updated",
		zap.String("max_transaction", thresholds.MaxTransactionAmount.String()),
		zap.String("max_slippage", thresholds.MaxSlippage.String()),
		zap.String("min_liquidity", thresholds.MinLiquidity.String()),
	)

	return nil
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("event_%d", time.Now().UnixNano())
}
