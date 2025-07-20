package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/entities"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/repositories"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/valueobjects"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// PostgresOrderRepository implements OrderRepository using PostgreSQL
type PostgresOrderRepository struct {
	db     *sql.DB
	tracer trace.Tracer
}

// NewPostgresOrderRepository creates a new PostgreSQL order repository
func NewPostgresOrderRepository(db *sql.DB) repositories.OrderRepository {
	return &PostgresOrderRepository{
		db:     db,
		tracer: otel.Tracer("hft.infrastructure.order_repository"),
	}
}

// Save saves an order to the database
func (r *PostgresOrderRepository) Save(ctx context.Context, order *entities.Order) error {
	ctx, span := r.tracer.Start(ctx, "PostgresOrderRepository.Save")
	defer span.End()

	span.SetAttributes(
		attribute.String("order_id", string(order.GetID())),
		attribute.String("strategy_id", string(order.GetStrategyID())),
		attribute.String("symbol", string(order.GetSymbol())),
	)

	query := `
		INSERT INTO hft_orders (
			id, client_order_id, strategy_id, symbol, exchange, side, type,
			quantity, price, stop_price, time_in_force, status, filled_quantity,
			remaining_quantity, avg_fill_price, commission_amount, commission_asset,
			created_at, updated_at, expires_at, exchange_order_id, error_message, latency
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21, $22, $23
		)`

	commission := order.GetCommission()
	var stopPrice *decimal.Decimal
	// Note: This would require adding GetStopPrice method to Order entity
	// if order.GetStopPrice() != nil {
	//     stopPrice = &order.GetStopPrice().Decimal
	// }

	_, err := r.db.ExecContext(ctx, query,
		string(order.GetID()),
		"", // client_order_id - would need to be added to Order entity
		string(order.GetStrategyID()),
		string(order.GetSymbol()),
		string(order.GetExchange()),
		string(order.GetSide()),
		string(order.GetOrderType()),
		order.GetQuantity().Decimal,
		order.GetPrice().Decimal,
		stopPrice,
		string(order.GetOrderType()), // time_in_force - would need to be added to Order entity
		string(order.GetStatus()),
		order.GetFilledQuantity().Decimal,
		order.GetRemainingQuantity().Decimal,
		order.GetAvgFillPrice().Decimal,
		commission.Amount,
		commission.Asset,
		order.GetCreatedAt(),
		order.GetUpdatedAt(),
		nil, // expires_at - would need to be added to Order entity
		"", // exchange_order_id - would need to be added to Order entity
		"", // error_message - would need to be added to Order entity
		int64(order.GetLatency()),
	)

	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to save order: %w", err)
	}

	return nil
}

// FindByID finds an order by ID
func (r *PostgresOrderRepository) FindByID(ctx context.Context, id entities.OrderID) (*entities.Order, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresOrderRepository.FindByID")
	defer span.End()

	span.SetAttributes(attribute.String("order_id", string(id)))

	query := `
		SELECT id, client_order_id, strategy_id, symbol, exchange, side, type,
			   quantity, price, stop_price, time_in_force, status, filled_quantity,
			   remaining_quantity, avg_fill_price, commission_amount, commission_asset,
			   created_at, updated_at, expires_at, exchange_order_id, error_message, latency
		FROM hft_orders
		WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, string(id))

	order, err := r.scanOrder(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found: %s", id)
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to find order: %w", err)
	}

	return order, nil
}

// Update updates an order in the database
func (r *PostgresOrderRepository) Update(ctx context.Context, order *entities.Order) error {
	ctx, span := r.tracer.Start(ctx, "PostgresOrderRepository.Update")
	defer span.End()

	span.SetAttributes(
		attribute.String("order_id", string(order.GetID())),
		attribute.String("status", string(order.GetStatus())),
	)

	query := `
		UPDATE hft_orders SET
			status = $2, filled_quantity = $3, remaining_quantity = $4,
			avg_fill_price = $5, commission_amount = $6, commission_asset = $7,
			updated_at = $8, latency = $9
		WHERE id = $1`

	commission := order.GetCommission()

	result, err := r.db.ExecContext(ctx, query,
		string(order.GetID()),
		string(order.GetStatus()),
		order.GetFilledQuantity().Decimal,
		order.GetRemainingQuantity().Decimal,
		order.GetAvgFillPrice().Decimal,
		commission.Amount,
		commission.Asset,
		order.GetUpdatedAt(),
		int64(order.GetLatency()),
	)

	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found: %s", order.GetID())
	}

	return nil
}

// Delete deletes an order from the database
func (r *PostgresOrderRepository) Delete(ctx context.Context, id entities.OrderID) error {
	ctx, span := r.tracer.Start(ctx, "PostgresOrderRepository.Delete")
	defer span.End()

	span.SetAttributes(attribute.String("order_id", string(id)))

	query := `DELETE FROM hft_orders WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, string(id))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found: %s", id)
	}

	return nil
}

// FindByStrategyID finds orders by strategy ID
func (r *PostgresOrderRepository) FindByStrategyID(ctx context.Context, strategyID entities.StrategyID) ([]*entities.Order, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresOrderRepository.FindByStrategyID")
	defer span.End()

	span.SetAttributes(attribute.String("strategy_id", string(strategyID)))

	query := `
		SELECT id, client_order_id, strategy_id, symbol, exchange, side, type,
			   quantity, price, stop_price, time_in_force, status, filled_quantity,
			   remaining_quantity, avg_fill_price, commission_amount, commission_asset,
			   created_at, updated_at, expires_at, exchange_order_id, error_message, latency
		FROM hft_orders
		WHERE strategy_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, string(strategyID))
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query orders by strategy: %w", err)
	}
	defer rows.Close()

	return r.scanOrders(rows)
}

// FindBySymbol finds orders by symbol
func (r *PostgresOrderRepository) FindBySymbol(ctx context.Context, symbol entities.Symbol) ([]*entities.Order, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresOrderRepository.FindBySymbol")
	defer span.End()

	span.SetAttributes(attribute.String("symbol", string(symbol)))

	query := `
		SELECT id, client_order_id, strategy_id, symbol, exchange, side, type,
			   quantity, price, stop_price, time_in_force, status, filled_quantity,
			   remaining_quantity, avg_fill_price, commission_amount, commission_asset,
			   created_at, updated_at, expires_at, exchange_order_id, error_message, latency
		FROM hft_orders
		WHERE symbol = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, string(symbol))
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query orders by symbol: %w", err)
	}
	defer rows.Close()

	return r.scanOrders(rows)
}

// FindByExchange finds orders by exchange
func (r *PostgresOrderRepository) FindByExchange(ctx context.Context, exchange entities.Exchange) ([]*entities.Order, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresOrderRepository.FindByExchange")
	defer span.End()

	span.SetAttributes(attribute.String("exchange", string(exchange)))

	query := `
		SELECT id, client_order_id, strategy_id, symbol, exchange, side, type,
			   quantity, price, stop_price, time_in_force, status, filled_quantity,
			   remaining_quantity, avg_fill_price, commission_amount, commission_asset,
			   created_at, updated_at, expires_at, exchange_order_id, error_message, latency
		FROM hft_orders
		WHERE exchange = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, string(exchange))
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query orders by exchange: %w", err)
	}
	defer rows.Close()

	return r.scanOrders(rows)
}

// FindByStatus finds orders by status
func (r *PostgresOrderRepository) FindByStatus(ctx context.Context, status valueobjects.OrderStatus) ([]*entities.Order, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresOrderRepository.FindByStatus")
	defer span.End()

	span.SetAttributes(attribute.String("status", string(status)))

	query := `
		SELECT id, client_order_id, strategy_id, symbol, exchange, side, type,
			   quantity, price, stop_price, time_in_force, status, filled_quantity,
			   remaining_quantity, avg_fill_price, commission_amount, commission_asset,
			   created_at, updated_at, expires_at, exchange_order_id, error_message, latency
		FROM hft_orders
		WHERE status = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, string(status))
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query orders by status: %w", err)
	}
	defer rows.Close()

	return r.scanOrders(rows)
}

// FindActiveOrders finds all active orders
func (r *PostgresOrderRepository) FindActiveOrders(ctx context.Context) ([]*entities.Order, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresOrderRepository.FindActiveOrders")
	defer span.End()

	query := `
		SELECT id, client_order_id, strategy_id, symbol, exchange, side, type,
			   quantity, price, stop_price, time_in_force, status, filled_quantity,
			   remaining_quantity, avg_fill_price, commission_amount, commission_asset,
			   created_at, updated_at, expires_at, exchange_order_id, error_message, latency
		FROM hft_orders
		WHERE status IN ('new', 'partially_filled')
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query active orders: %w", err)
	}
	defer rows.Close()

	return r.scanOrders(rows)
}

// FindOrdersInDateRange finds orders in a date range
func (r *PostgresOrderRepository) FindOrdersInDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entities.Order, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresOrderRepository.FindOrdersInDateRange")
	defer span.End()

	span.SetAttributes(
		attribute.String("start_date", startDate.Format(time.RFC3339)),
		attribute.String("end_date", endDate.Format(time.RFC3339)),
	)

	query := `
		SELECT id, client_order_id, strategy_id, symbol, exchange, side, type,
			   quantity, price, stop_price, time_in_force, status, filled_quantity,
			   remaining_quantity, avg_fill_price, commission_amount, commission_asset,
			   created_at, updated_at, expires_at, exchange_order_id, error_message, latency
		FROM hft_orders
		WHERE created_at >= $1 AND created_at <= $2
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query orders in date range: %w", err)
	}
	defer rows.Close()

	return r.scanOrders(rows)
}

// FindOrdersByStrategyAndStatus finds orders by strategy and status
func (r *PostgresOrderRepository) FindOrdersByStrategyAndStatus(ctx context.Context, strategyID entities.StrategyID, status valueobjects.OrderStatus) ([]*entities.Order, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresOrderRepository.FindOrdersByStrategyAndStatus")
	defer span.End()

	span.SetAttributes(
		attribute.String("strategy_id", string(strategyID)),
		attribute.String("status", string(status)),
	)

	query := `
		SELECT id, client_order_id, strategy_id, symbol, exchange, side, type,
			   quantity, price, stop_price, time_in_force, status, filled_quantity,
			   remaining_quantity, avg_fill_price, commission_amount, commission_asset,
			   created_at, updated_at, expires_at, exchange_order_id, error_message, latency
		FROM hft_orders
		WHERE strategy_id = $1 AND status = $2
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, string(strategyID), string(status))
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query orders by strategy and status: %w", err)
	}
	defer rows.Close()

	return r.scanOrders(rows)
}

// FindOrdersBySymbolAndSide finds orders by symbol and side
func (r *PostgresOrderRepository) FindOrdersBySymbolAndSide(ctx context.Context, symbol entities.Symbol, side valueobjects.OrderSide) ([]*entities.Order, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresOrderRepository.FindOrdersBySymbolAndSide")
	defer span.End()

	span.SetAttributes(
		attribute.String("symbol", string(symbol)),
		attribute.String("side", string(side)),
	)

	query := `
		SELECT id, client_order_id, strategy_id, symbol, exchange, side, type,
			   quantity, price, stop_price, time_in_force, status, filled_quantity,
			   remaining_quantity, avg_fill_price, commission_amount, commission_asset,
			   created_at, updated_at, expires_at, exchange_order_id, error_message, latency
		FROM hft_orders
		WHERE symbol = $1 AND side = $2
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, string(symbol), string(side))
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query orders by symbol and side: %w", err)
	}
	defer rows.Close()

	return r.scanOrders(rows)
}

// FindRecentOrders finds recent orders with limit
func (r *PostgresOrderRepository) FindRecentOrders(ctx context.Context, limit int) ([]*entities.Order, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresOrderRepository.FindRecentOrders")
	defer span.End()

	span.SetAttributes(attribute.Int("limit", limit))

	query := `
		SELECT id, client_order_id, strategy_id, symbol, exchange, side, type,
			   quantity, price, stop_price, time_in_force, status, filled_quantity,
			   remaining_quantity, avg_fill_price, commission_amount, commission_asset,
			   created_at, updated_at, expires_at, exchange_order_id, error_message, latency
		FROM hft_orders
		ORDER BY created_at DESC
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query recent orders: %w", err)
	}
	defer rows.Close()

	return r.scanOrders(rows)
}

// CountOrdersByStrategy counts orders by strategy
func (r *PostgresOrderRepository) CountOrdersByStrategy(ctx context.Context, strategyID entities.StrategyID) (int64, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresOrderRepository.CountOrdersByStrategy")
	defer span.End()

	span.SetAttributes(attribute.String("strategy_id", string(strategyID)))

	query := `SELECT COUNT(*) FROM hft_orders WHERE strategy_id = $1`

	var count int64
	err := r.db.QueryRowContext(ctx, query, string(strategyID)).Scan(&count)
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to count orders by strategy: %w", err)
	}

	span.SetAttributes(attribute.Int64("count", count))
	return count, nil
}

// CountOrdersByStatus counts orders by status
func (r *PostgresOrderRepository) CountOrdersByStatus(ctx context.Context, status valueobjects.OrderStatus) (int64, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresOrderRepository.CountOrdersByStatus")
	defer span.End()

	span.SetAttributes(attribute.String("status", string(status)))

	query := `SELECT COUNT(*) FROM hft_orders WHERE status = $1`

	var count int64
	err := r.db.QueryRowContext(ctx, query, string(status)).Scan(&count)
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to count orders by status: %w", err)
	}

	span.SetAttributes(attribute.Int64("count", count))
	return count, nil
}

// GetOrderVolumeByStrategy gets total order volume by strategy
func (r *PostgresOrderRepository) GetOrderVolumeByStrategy(ctx context.Context, strategyID entities.StrategyID) (valueobjects.Quantity, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresOrderRepository.GetOrderVolumeByStrategy")
	defer span.End()

	span.SetAttributes(attribute.String("strategy_id", string(strategyID)))

	query := `SELECT COALESCE(SUM(quantity), 0) FROM hft_orders WHERE strategy_id = $1`

	var volume decimal.Decimal
	err := r.db.QueryRowContext(ctx, query, string(strategyID)).Scan(&volume)
	if err != nil {
		span.RecordError(err)
		return valueobjects.Quantity{}, fmt.Errorf("failed to get order volume by strategy: %w", err)
	}

	return valueobjects.Quantity{Decimal: volume}, nil
}

// SaveEvents saves order events to the database
func (r *PostgresOrderRepository) SaveEvents(ctx context.Context, orderID entities.OrderID, events []valueobjects.OrderEvent) error {
	ctx, span := r.tracer.Start(ctx, "PostgresOrderRepository.SaveEvents")
	defer span.End()

	span.SetAttributes(
		attribute.String("order_id", string(orderID)),
		attribute.Int("events_count", len(events)),
	)

	if len(events) == 0 {
		return nil
	}

	query := `
		INSERT INTO hft_order_events (order_id, event_type, event_data, timestamp)
		VALUES ($1, $2, $3, $4)`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, event := range events {
		eventData, err := json.Marshal(event.Data)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to marshal event data: %w", err)
		}

		_, err = tx.ExecContext(ctx, query,
			string(orderID),
			string(event.Type),
			eventData,
			event.Timestamp,
		)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to save event: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetEvents gets order events from the database
func (r *PostgresOrderRepository) GetEvents(ctx context.Context, orderID entities.OrderID) ([]valueobjects.OrderEvent, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresOrderRepository.GetEvents")
	defer span.End()

	span.SetAttributes(attribute.String("order_id", string(orderID)))

	query := `
		SELECT event_type, event_data, timestamp
		FROM hft_order_events
		WHERE order_id = $1
		ORDER BY timestamp ASC`

	rows, err := r.db.QueryContext(ctx, query, string(orderID))
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query order events: %w", err)
	}
	defer rows.Close()

	var events []valueobjects.OrderEvent
	for rows.Next() {
		var eventType string
		var eventDataJSON []byte
		var timestamp time.Time

		if err := rows.Scan(&eventType, &eventDataJSON, &timestamp); err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}

		var eventData map[string]interface{}
		if err := json.Unmarshal(eventDataJSON, &eventData); err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
		}

		events = append(events, valueobjects.OrderEvent{
			Type:      valueobjects.OrderEventType(eventType),
			Data:      eventData,
			Timestamp: timestamp,
		})
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to iterate over event rows: %w", err)
	}

	span.SetAttributes(attribute.Int("events_count", len(events)))
	return events, nil
}

// GetEventsSince gets order events since a specific time
func (r *PostgresOrderRepository) GetEventsSince(ctx context.Context, orderID entities.OrderID, since time.Time) ([]valueobjects.OrderEvent, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresOrderRepository.GetEventsSince")
	defer span.End()

	span.SetAttributes(
		attribute.String("order_id", string(orderID)),
		attribute.String("since", since.Format(time.RFC3339)),
	)

	query := `
		SELECT event_type, event_data, timestamp
		FROM hft_order_events
		WHERE order_id = $1 AND timestamp > $2
		ORDER BY timestamp ASC`

	rows, err := r.db.QueryContext(ctx, query, string(orderID), since)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query order events since: %w", err)
	}
	defer rows.Close()

	var events []valueobjects.OrderEvent
	for rows.Next() {
		var eventType string
		var eventDataJSON []byte
		var timestamp time.Time

		if err := rows.Scan(&eventType, &eventDataJSON, &timestamp); err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}

		var eventData map[string]interface{}
		if err := json.Unmarshal(eventDataJSON, &eventData); err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
		}

		events = append(events, valueobjects.OrderEvent{
			Type:      valueobjects.OrderEventType(eventType),
			Data:      eventData,
			Timestamp: timestamp,
		})
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to iterate over event rows: %w", err)
	}

	span.SetAttributes(attribute.Int("events_count", len(events)))
	return events, nil
}

// Helper methods

// scanOrder scans a single order from a database row
func (r *PostgresOrderRepository) scanOrder(row *sql.Row) (*entities.Order, error) {
	var id, clientOrderID, strategyID, symbol, exchange, side, orderType string
	var quantity, price, filledQuantity, remainingQuantity, avgFillPrice decimal.Decimal
	var stopPrice *decimal.Decimal
	var timeInForce, status string
	var commissionAmount decimal.Decimal
	var commissionAsset string
	var createdAt, updatedAt time.Time
	var expiresAt *time.Time
	var exchangeOrderID, errorMessage string
	var latency int64

	err := row.Scan(
		&id, &clientOrderID, &strategyID, &symbol, &exchange, &side, &orderType,
		&quantity, &price, &stopPrice, &timeInForce, &status, &filledQuantity,
		&remainingQuantity, &avgFillPrice, &commissionAmount, &commissionAsset,
		&createdAt, &updatedAt, &expiresAt, &exchangeOrderID, &errorMessage, &latency,
	)
	if err != nil {
		return nil, err
	}

	// Create order entity - this is a simplified version
	// In a real implementation, you would need to reconstruct the order properly
	// with all its domain logic and events
	
	// Note: This is a placeholder implementation
	// The actual implementation would need to properly reconstruct the order entity
	// with all its business logic, validation, and events
	
	return nil, fmt.Errorf("order reconstruction not implemented")
}

// scanOrders scans multiple orders from database rows
func (r *PostgresOrderRepository) scanOrders(rows *sql.Rows) ([]*entities.Order, error) {
	var orders []*entities.Order

	for rows.Next() {
		// Similar scanning logic as scanOrder but for multiple rows
		// This would need to be implemented properly
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over order rows: %w", err)
	}

	return orders, nil
}
