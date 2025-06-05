package risk

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// RiskCheckerImpl implements the RiskChecker interface
type RiskCheckerImpl struct {
	config *config.Config
	db     *sql.DB
	redis  *redis.Client
}

// NewRiskChecker creates a new risk checker
func NewRiskChecker(cfg *config.Config, db *sql.DB, redis *redis.Client) RiskChecker {
	return &RiskCheckerImpl{
		config: cfg,
		db:     db,
		redis:  redis,
	}
}

// ValidateOrder validates an order against risk limits
func (rc *RiskCheckerImpl) ValidateOrder(ctx context.Context, order *models.Order) error {
	// Check order size limits
	if err := rc.checkOrderSize(order); err != nil {
		return fmt.Errorf("order size check failed: %w", err)
	}

	// Check position limits
	if err := rc.checkPositionLimits(ctx, order); err != nil {
		return fmt.Errorf("position limit check failed: %w", err)
	}

	// Check exposure limits
	if err := rc.checkExposureLimits(ctx, order); err != nil {
		return fmt.Errorf("exposure limit check failed: %w", err)
	}

	// Check order rate limits
	if err := rc.checkOrderRateLimits(ctx, order); err != nil {
		return fmt.Errorf("order rate limit check failed: %w", err)
	}

	// Check market hours
	if err := rc.checkMarketHours(order); err != nil {
		return fmt.Errorf("market hours check failed: %w", err)
	}

	return nil
}

// ValidatePosition validates a position against risk limits
func (rc *RiskCheckerImpl) ValidatePosition(ctx context.Context, position *models.Position) error {
	// Check position size limits
	if err := rc.checkPositionSize(position); err != nil {
		return fmt.Errorf("position size check failed: %w", err)
	}

	// Check PnL limits
	if err := rc.checkPnLLimits(position); err != nil {
		return fmt.Errorf("PnL limit check failed: %w", err)
	}

	// Check margin requirements
	if err := rc.checkMarginRequirements(position); err != nil {
		return fmt.Errorf("margin requirement check failed: %w", err)
	}

	return nil
}

// CheckExposure checks total exposure for a strategy
func (rc *RiskCheckerImpl) CheckExposure(ctx context.Context, strategyID string) error {
	// Get strategy risk limits
	limits, err := rc.getStrategyRiskLimits(ctx, strategyID)
	if err != nil {
		return fmt.Errorf("failed to get strategy risk limits: %w", err)
	}

	// Calculate current exposure
	exposure, err := rc.calculateExposure(ctx, strategyID)
	if err != nil {
		return fmt.Errorf("failed to calculate exposure: %w", err)
	}

	// Check exposure limit
	if exposure.GreaterThan(limits.MaxExposure) {
		return fmt.Errorf("exposure %.2f exceeds limit %.2f", exposure, limits.MaxExposure)
	}

	return nil
}

// CheckDrawdown checks drawdown limits for a strategy
func (rc *RiskCheckerImpl) CheckDrawdown(ctx context.Context, strategyID string) error {
	// Get strategy risk limits
	limits, err := rc.getStrategyRiskLimits(ctx, strategyID)
	if err != nil {
		return fmt.Errorf("failed to get strategy risk limits: %w", err)
	}

	// Calculate current drawdown
	drawdown, err := rc.calculateDrawdown(ctx, strategyID)
	if err != nil {
		return fmt.Errorf("failed to calculate drawdown: %w", err)
	}

	// Check drawdown limit
	if drawdown.GreaterThan(limits.MaxDrawdown) {
		return fmt.Errorf("drawdown %.2f%% exceeds limit %.2f%%", drawdown, limits.MaxDrawdown)
	}

	return nil
}

// checkOrderSize validates order size against limits
func (rc *RiskCheckerImpl) checkOrderSize(order *models.Order) error {
	// Get strategy risk limits
	limits, err := rc.getStrategyRiskLimits(context.Background(), order.StrategyID)
	if err != nil {
		// Use default limits if strategy limits not found
		limits = rc.getDefaultRiskLimits()
	}

	if order.Quantity.GreaterThan(limits.MaxOrderSize) {
		return fmt.Errorf("order size %.8f exceeds limit %.8f", order.Quantity, limits.MaxOrderSize)
	}

	return nil
}

// checkPositionLimits validates position limits
func (rc *RiskCheckerImpl) checkPositionLimits(ctx context.Context, order *models.Order) error {
	// Get current position
	position, err := rc.getCurrentPosition(ctx, order.StrategyID, order.Symbol, order.Exchange)
	if err != nil {
		// No existing position, order is allowed
		return nil
	}

	// Get strategy risk limits
	limits, err := rc.getStrategyRiskLimits(ctx, order.StrategyID)
	if err != nil {
		limits = rc.getDefaultRiskLimits()
	}

	// Calculate new position size after order
	newSize := position.Size
	if position.Side == order.Side {
		newSize = newSize.Add(order.Quantity)
	} else {
		newSize = newSize.Sub(order.Quantity)
		if newSize.IsNegative() {
			newSize = newSize.Abs()
		}
	}

	if newSize.GreaterThan(limits.MaxPositionSize) {
		return fmt.Errorf("new position size %.8f would exceed limit %.8f", newSize, limits.MaxPositionSize)
	}

	return nil
}

// checkExposureLimits validates exposure limits
func (rc *RiskCheckerImpl) checkExposureLimits(ctx context.Context, order *models.Order) error {
	// Get strategy risk limits
	limits, err := rc.getStrategyRiskLimits(ctx, order.StrategyID)
	if err != nil {
		limits = rc.getDefaultRiskLimits()
	}

	// Calculate current exposure
	exposure, err := rc.calculateExposure(ctx, order.StrategyID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to calculate exposure, allowing order")
		return nil
	}

	// Calculate additional exposure from this order
	orderExposure := order.Quantity.Mul(order.Price)
	newExposure := exposure.Add(orderExposure)

	if newExposure.GreaterThan(limits.MaxExposure) {
		return fmt.Errorf("new exposure %.2f would exceed limit %.2f", newExposure, limits.MaxExposure)
	}

	return nil
}

// checkOrderRateLimits validates order rate limits
func (rc *RiskCheckerImpl) checkOrderRateLimits(ctx context.Context, order *models.Order) error {
	// Get strategy risk limits
	limits, err := rc.getStrategyRiskLimits(ctx, order.StrategyID)
	if err != nil {
		limits = rc.getDefaultRiskLimits()
	}

	// Check orders per second limit
	orderCount, err := rc.getRecentOrderCount(ctx, order.StrategyID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get recent order count, allowing order")
		return nil
	}

	if orderCount >= limits.MaxOrdersPerSecond {
		return fmt.Errorf("order rate %d exceeds limit %d orders per second", orderCount, limits.MaxOrdersPerSecond)
	}

	return nil
}

// checkMarketHours validates market hours
func (rc *RiskCheckerImpl) checkMarketHours(order *models.Order) error {
	// Crypto markets are 24/7, so always allow
	// In traditional markets, would check trading hours
	return nil
}

// checkPositionSize validates position size
func (rc *RiskCheckerImpl) checkPositionSize(position *models.Position) error {
	limits, err := rc.getStrategyRiskLimits(context.Background(), position.StrategyID)
	if err != nil {
		limits = rc.getDefaultRiskLimits()
	}

	if position.Size.GreaterThan(limits.MaxPositionSize) {
		return fmt.Errorf("position size %.8f exceeds limit %.8f", position.Size, limits.MaxPositionSize)
	}

	return nil
}

// checkPnLLimits validates PnL limits
func (rc *RiskCheckerImpl) checkPnLLimits(position *models.Position) error {
	limits, err := rc.getStrategyRiskLimits(context.Background(), position.StrategyID)
	if err != nil {
		limits = rc.getDefaultRiskLimits()
	}

	// Check daily loss limit
	if position.UnrealizedPnL.IsNegative() && position.UnrealizedPnL.Abs().GreaterThan(limits.MaxDailyLoss) {
		return fmt.Errorf("unrealized loss %.2f exceeds daily limit %.2f", position.UnrealizedPnL.Abs(), limits.MaxDailyLoss)
	}

	return nil
}

// checkMarginRequirements validates margin requirements
func (rc *RiskCheckerImpl) checkMarginRequirements(position *models.Position) error {
	if position.Margin.GreaterThan(decimal.Zero) && position.Margin.LessThan(position.MaintenanceMargin) {
		return fmt.Errorf("margin %.2f below maintenance margin %.2f", position.Margin, position.MaintenanceMargin)
	}

	return nil
}

// getStrategyRiskLimits gets risk limits for a strategy
func (rc *RiskCheckerImpl) getStrategyRiskLimits(ctx context.Context, strategyID string) (*models.RiskLimits, error) {
	// Placeholder implementation - would query database
	return rc.getDefaultRiskLimits(), nil
}

// getDefaultRiskLimits returns default risk limits
func (rc *RiskCheckerImpl) getDefaultRiskLimits() *models.RiskLimits {
	return &models.RiskLimits{
		MaxPositionSize:    decimal.NewFromFloat(10.0),
		MaxDailyLoss:       decimal.NewFromFloat(1000.0),
		MaxDrawdown:        decimal.NewFromFloat(5.0), // 5%
		MaxOrderSize:       decimal.NewFromFloat(1.0),
		MaxOrdersPerSecond: 10,
		MaxExposure:        decimal.NewFromFloat(50000.0),
		StopLossPercent:    decimal.NewFromFloat(2.0),
		TakeProfitPercent:  decimal.NewFromFloat(5.0),
	}
}

// getCurrentPosition gets current position for a strategy/symbol/exchange
func (rc *RiskCheckerImpl) getCurrentPosition(ctx context.Context, strategyID, symbol, exchange string) (*models.Position, error) {
	// Placeholder implementation - would query database
	return nil, fmt.Errorf("position not found")
}

// calculateExposure calculates total exposure for a strategy
func (rc *RiskCheckerImpl) calculateExposure(ctx context.Context, strategyID string) (decimal.Decimal, error) {
	// Placeholder implementation - would calculate from positions
	return decimal.NewFromFloat(10000.0), nil
}

// calculateDrawdown calculates current drawdown for a strategy
func (rc *RiskCheckerImpl) calculateDrawdown(ctx context.Context, strategyID string) (decimal.Decimal, error) {
	// Placeholder implementation - would calculate from performance history
	return decimal.NewFromFloat(2.5), nil
}

// getRecentOrderCount gets recent order count for rate limiting
func (rc *RiskCheckerImpl) getRecentOrderCount(ctx context.Context, strategyID string) (int, error) {
	// Placeholder implementation - would query Redis for recent orders
	return 5, nil
}
