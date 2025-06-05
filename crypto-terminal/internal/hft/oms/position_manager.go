package oms

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// PositionManagerImpl implements the PositionManager interface
type PositionManagerImpl struct {
	config *config.Config
	db     *sql.DB
	redis  *redis.Client
}

// NewPositionManager creates a new position manager
func NewPositionManager(cfg *config.Config, db *sql.DB, redis *redis.Client) PositionManager {
	return &PositionManagerImpl{
		config: cfg,
		db:     db,
		redis:  redis,
	}
}

// GetPosition retrieves a position for a specific strategy, symbol, and exchange
func (pm *PositionManagerImpl) GetPosition(ctx context.Context, strategyID, symbol, exchange string) (*models.Position, error) {
	query := `
		SELECT id, strategy_id, symbol, exchange, side, size, entry_price,
		       mark_price, unrealized_pnl, realized_pnl, margin, maintenance_margin,
		       created_at, updated_at
		FROM hft_positions 
		WHERE strategy_id = $1 AND symbol = $2 AND exchange = $3
	`

	position := &models.Position{}
	err := pm.db.QueryRowContext(ctx, query, strategyID, symbol, exchange).Scan(
		&position.ID, &position.StrategyID, &position.Symbol, &position.Exchange,
		&position.Side, &position.Size, &position.EntryPrice, &position.MarkPrice,
		&position.UnrealizedPnL, &position.RealizedPnL, &position.Margin,
		&position.MaintenanceMargin, &position.CreatedAt, &position.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Create new position if it doesn't exist
		return pm.createNewPosition(ctx, strategyID, symbol, exchange)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get position from database: %w", err)
	}

	return position, nil
}

// GetAllPositions retrieves all positions for a strategy
func (pm *PositionManagerImpl) GetAllPositions(ctx context.Context, strategyID string) ([]*models.Position, error) {
	query := `
		SELECT id, strategy_id, symbol, exchange, side, size, entry_price,
		       mark_price, unrealized_pnl, realized_pnl, margin, maintenance_margin,
		       created_at, updated_at
		FROM hft_positions 
		WHERE strategy_id = $1
		ORDER BY symbol, exchange
	`

	rows, err := pm.db.QueryContext(ctx, query, strategyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}
	defer rows.Close()

	var positions []*models.Position
	for rows.Next() {
		position := &models.Position{}
		err := rows.Scan(
			&position.ID, &position.StrategyID, &position.Symbol, &position.Exchange,
			&position.Side, &position.Size, &position.EntryPrice, &position.MarkPrice,
			&position.UnrealizedPnL, &position.RealizedPnL, &position.Margin,
			&position.MaintenanceMargin, &position.CreatedAt, &position.UpdatedAt,
		)
		if err != nil {
			logrus.WithError(err).Error("Failed to scan position row")
			continue
		}
		positions = append(positions, position)
	}

	return positions, nil
}

// UpdatePosition updates an existing position
func (pm *PositionManagerImpl) UpdatePosition(ctx context.Context, position *models.Position) error {
	// Calculate unrealized PnL
	if err := pm.calculateUnrealizedPnL(ctx, position); err != nil {
		logrus.WithError(err).Error("Failed to calculate unrealized PnL")
	}

	position.UpdatedAt = time.Now()

	query := `
		UPDATE hft_positions SET
			side = $2, size = $3, entry_price = $4, mark_price = $5,
			unrealized_pnl = $6, realized_pnl = $7, margin = $8,
			maintenance_margin = $9, updated_at = $10
		WHERE id = $1
	`

	_, err := pm.db.ExecContext(ctx, query,
		position.ID, position.Side, position.Size, position.EntryPrice,
		position.MarkPrice, position.UnrealizedPnL, position.RealizedPnL,
		position.Margin, position.MaintenanceMargin, position.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update position: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"position_id":  position.ID,
		"strategy_id":  position.StrategyID,
		"symbol":       position.Symbol,
		"size":         position.Size,
		"unrealized_pnl": position.UnrealizedPnL,
	}).Debug("Position updated successfully")

	return nil
}

// ClosePosition closes a position by setting size to zero
func (pm *PositionManagerImpl) ClosePosition(ctx context.Context, strategyID, symbol, exchange string) error {
	position, err := pm.GetPosition(ctx, strategyID, symbol, exchange)
	if err != nil {
		return fmt.Errorf("failed to get position: %w", err)
	}

	// Calculate realized PnL before closing
	realizedPnL := position.UnrealizedPnL
	position.RealizedPnL = position.RealizedPnL.Add(realizedPnL)
	position.Size = decimal.Zero
	position.UnrealizedPnL = decimal.Zero
	position.UpdatedAt = time.Now()

	if err := pm.UpdatePosition(ctx, position); err != nil {
		return fmt.Errorf("failed to close position: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"strategy_id":   strategyID,
		"symbol":        symbol,
		"exchange":      exchange,
		"realized_pnl":  realizedPnL,
	}).Info("Position closed successfully")

	return nil
}

// createNewPosition creates a new position with zero size
func (pm *PositionManagerImpl) createNewPosition(ctx context.Context, strategyID, symbol, exchange string) (*models.Position, error) {
	position := &models.Position{
		ID:                fmt.Sprintf("pos_%d", time.Now().UnixNano()),
		StrategyID:        strategyID,
		Symbol:            symbol,
		Exchange:          exchange,
		Side:              models.OrderSideBuy, // Default side
		Size:              decimal.Zero,
		EntryPrice:        decimal.Zero,
		MarkPrice:         decimal.Zero,
		UnrealizedPnL:     decimal.Zero,
		RealizedPnL:       decimal.Zero,
		Margin:            decimal.Zero,
		MaintenanceMargin: decimal.Zero,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	query := `
		INSERT INTO hft_positions (
			id, strategy_id, symbol, exchange, side, size, entry_price,
			mark_price, unrealized_pnl, realized_pnl, margin, maintenance_margin,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	_, err := pm.db.ExecContext(ctx, query,
		position.ID, position.StrategyID, position.Symbol, position.Exchange,
		position.Side, position.Size, position.EntryPrice, position.MarkPrice,
		position.UnrealizedPnL, position.RealizedPnL, position.Margin,
		position.MaintenanceMargin, position.CreatedAt, position.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create new position: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"position_id": position.ID,
		"strategy_id": strategyID,
		"symbol":      symbol,
		"exchange":    exchange,
	}).Info("New position created")

	return position, nil
}

// calculateUnrealizedPnL calculates unrealized PnL for a position
func (pm *PositionManagerImpl) calculateUnrealizedPnL(ctx context.Context, position *models.Position) error {
	if position.Size.IsZero() {
		position.UnrealizedPnL = decimal.Zero
		return nil
	}

	// Get current market price (mark price)
	markPrice, err := pm.getCurrentPrice(ctx, position.Symbol, position.Exchange)
	if err != nil {
		return fmt.Errorf("failed to get current price: %w", err)
	}

	position.MarkPrice = markPrice

	// Calculate unrealized PnL
	if position.Side == models.OrderSideBuy {
		// Long position: PnL = (mark_price - entry_price) * size
		priceDiff := markPrice.Sub(position.EntryPrice)
		position.UnrealizedPnL = priceDiff.Mul(position.Size)
	} else {
		// Short position: PnL = (entry_price - mark_price) * size
		priceDiff := position.EntryPrice.Sub(markPrice)
		position.UnrealizedPnL = priceDiff.Mul(position.Size)
	}

	return nil
}

// getCurrentPrice gets the current market price for a symbol
func (pm *PositionManagerImpl) getCurrentPrice(ctx context.Context, symbol, exchange string) (decimal.Decimal, error) {
	// Try to get from Redis cache first
	cacheKey := fmt.Sprintf("hft:price:%s:%s", exchange, symbol)
	priceStr, err := pm.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		price, err := decimal.NewFromString(priceStr)
		if err == nil {
			return price, nil
		}
	}

	// Fallback to a default price (in real implementation, would fetch from market data)
	logrus.WithFields(logrus.Fields{
		"symbol":   symbol,
		"exchange": exchange,
	}).Warn("Using fallback price for position calculation")

	return decimal.NewFromFloat(50000.0), nil // Default BTC price
}

// UpdatePositionFromFill updates position based on a trade fill
func (pm *PositionManagerImpl) UpdatePositionFromFill(ctx context.Context, fill *models.Fill) error {
	position, err := pm.GetPosition(ctx, "", fill.Symbol, fill.Exchange) // Strategy ID would come from order
	if err != nil {
		return fmt.Errorf("failed to get position: %w", err)
	}

	// Update position based on fill
	if position.Size.IsZero() {
		// New position
		position.Side = fill.Side
		position.Size = fill.Quantity
		position.EntryPrice = fill.Price
	} else {
		// Existing position
		if position.Side == fill.Side {
			// Adding to position
			totalValue := position.EntryPrice.Mul(position.Size).Add(fill.Price.Mul(fill.Quantity))
			position.Size = position.Size.Add(fill.Quantity)
			position.EntryPrice = totalValue.Div(position.Size)
		} else {
			// Reducing or reversing position
			if fill.Quantity.GreaterThanOrEqual(position.Size) {
				// Position reversal
				remainingQty := fill.Quantity.Sub(position.Size)
				if remainingQty.IsPositive() {
					position.Side = fill.Side
					position.Size = remainingQty
					position.EntryPrice = fill.Price
				} else {
					position.Size = decimal.Zero
				}
			} else {
				// Partial reduction
				position.Size = position.Size.Sub(fill.Quantity)
			}
		}
	}

	return pm.UpdatePosition(ctx, position)
}
