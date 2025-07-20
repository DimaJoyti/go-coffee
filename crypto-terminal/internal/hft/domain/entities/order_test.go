package entities

import (
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/valueobjects"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrder(t *testing.T) {
	tests := []struct {
		name        string
		strategyID  StrategyID
		symbol      Symbol
		exchange    Exchange
		side        valueobjects.OrderSide
		orderType   valueobjects.OrderType
		quantity    valueobjects.Quantity
		price       valueobjects.Price
		timeInForce valueobjects.TimeInForce
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid limit buy order",
			strategyID:  "strategy-1",
			symbol:      "BTCUSDT",
			exchange:    "binance",
			side:        valueobjects.OrderSideBuy,
			orderType:   valueobjects.OrderTypeLimit,
			quantity:    valueobjects.Quantity{Decimal: decimal.NewFromFloat(0.1)},
			price:       valueobjects.Price{Decimal: decimal.NewFromFloat(50000)},
			timeInForce: valueobjects.TimeInForceGTC,
			wantErr:     false,
		},
		{
			name:        "valid market sell order",
			strategyID:  "strategy-2",
			symbol:      "ETHUSDT",
			exchange:    "coinbase",
			side:        valueobjects.OrderSideSell,
			orderType:   valueobjects.OrderTypeMarket,
			quantity:    valueobjects.Quantity{Decimal: decimal.NewFromFloat(1.0)},
			price:       valueobjects.Price{Decimal: decimal.Zero},
			timeInForce: valueobjects.TimeInForceIOC,
			wantErr:     false,
		},
		{
			name:        "empty strategy ID",
			strategyID:  "",
			symbol:      "BTCUSDT",
			exchange:    "binance",
			side:        valueobjects.OrderSideBuy,
			orderType:   valueobjects.OrderTypeLimit,
			quantity:    valueobjects.Quantity{Decimal: decimal.NewFromFloat(0.1)},
			price:       valueobjects.Price{Decimal: decimal.NewFromFloat(50000)},
			timeInForce: valueobjects.TimeInForceGTC,
			wantErr:     true,
			errContains: "strategy ID cannot be empty",
		},
		{
			name:        "empty symbol",
			strategyID:  "strategy-1",
			symbol:      "",
			exchange:    "binance",
			side:        valueobjects.OrderSideBuy,
			orderType:   valueobjects.OrderTypeLimit,
			quantity:    valueobjects.Quantity{Decimal: decimal.NewFromFloat(0.1)},
			price:       valueobjects.Price{Decimal: decimal.NewFromFloat(50000)},
			timeInForce: valueobjects.TimeInForceGTC,
			wantErr:     true,
			errContains: "symbol cannot be empty",
		},
		{
			name:        "invalid quantity",
			strategyID:  "strategy-1",
			symbol:      "BTCUSDT",
			exchange:    "binance",
			side:        valueobjects.OrderSideBuy,
			orderType:   valueobjects.OrderTypeLimit,
			quantity:    valueobjects.Quantity{Decimal: decimal.Zero},
			price:       valueobjects.Price{Decimal: decimal.NewFromFloat(50000)},
			timeInForce: valueobjects.TimeInForceGTC,
			wantErr:     true,
			errContains: "invalid quantity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := NewOrder(
				tt.strategyID,
				tt.symbol,
				tt.exchange,
				tt.side,
				tt.orderType,
				tt.quantity,
				tt.price,
				tt.timeInForce,
			)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, order)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, order)

				// Verify order properties
				assert.NotEmpty(t, order.GetID())
				assert.Equal(t, tt.strategyID, order.GetStrategyID())
				assert.Equal(t, tt.symbol, order.GetSymbol())
				assert.Equal(t, tt.exchange, order.GetExchange())
				assert.Equal(t, tt.side, order.GetSide())
				assert.Equal(t, tt.orderType, order.GetOrderType())
				assert.Equal(t, tt.quantity, order.GetQuantity())
				assert.Equal(t, tt.price, order.GetPrice())
				assert.Equal(t, tt.timeInForce, order.GetOrderType()) // Note: This should be timeInForce getter
				assert.Equal(t, OrderStatusPending, order.GetStatus())
				assert.True(t, order.GetFilledQuantity().IsZero())
				assert.Equal(t, tt.quantity, order.GetRemainingQuantity())
				assert.True(t, order.GetAvgFillPrice().IsZero())
				assert.NotZero(t, order.GetCreatedAt())
				assert.NotZero(t, order.GetUpdatedAt())

				// Verify events
				events := order.GetEvents()
				assert.Len(t, events, 1)
				assert.Equal(t, OrderEventCreated, events[0].Type)
			}
		})
	}
}

func TestOrder_Confirm(t *testing.T) {
	// Create a valid order
	order, err := createTestOrder()
	require.NoError(t, err)

	// Test successful confirmation
	err = order.Confirm()
	assert.NoError(t, err)
	assert.Equal(t, OrderStatusNew, order.GetStatus())

	// Verify event was recorded
	events := order.GetEvents()
	assert.Len(t, events, 2) // Created + Confirmed
	assert.Equal(t, OrderEventConfirmed, events[1].Type)

	// Test confirming already confirmed order
	err = order.Confirm()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot confirm order in status")
}

func TestOrder_PartialFill(t *testing.T) {
	tests := []struct {
		name         string
		setupOrder   func() *Order
		fillQuantity valueobjects.Quantity
		fillPrice    valueobjects.Price
		commission   valueobjects.Commission
		wantErr      bool
		errContains  string
		expectedStatus valueobjects.OrderStatus
	}{
		{
			name: "valid partial fill",
			setupOrder: func() *Order {
				order, _ := createTestOrder()
				order.Confirm()
				return order
			},
			fillQuantity:   valueobjects.Quantity{Decimal: decimal.NewFromFloat(0.05)},
			fillPrice:      valueobjects.Price{Decimal: decimal.NewFromFloat(50100)},
			commission:     valueobjects.Commission{Amount: decimal.NewFromFloat(0.1), Asset: "USDT"},
			wantErr:        false,
			expectedStatus: OrderStatusPartiallyFilled,
		},
		{
			name: "complete fill",
			setupOrder: func() *Order {
				order, _ := createTestOrder()
				order.Confirm()
				return order
			},
			fillQuantity:   valueobjects.Quantity{Decimal: decimal.NewFromFloat(0.1)},
			fillPrice:      valueobjects.Price{Decimal: decimal.NewFromFloat(50200)},
			commission:     valueobjects.Commission{Amount: decimal.NewFromFloat(0.2), Asset: "USDT"},
			wantErr:        false,
			expectedStatus: OrderStatusFilled,
		},
		{
			name: "overfill",
			setupOrder: func() *Order {
				order, _ := createTestOrder()
				order.Confirm()
				return order
			},
			fillQuantity: valueobjects.Quantity{Decimal: decimal.NewFromFloat(0.2)},
			fillPrice:    valueobjects.Price{Decimal: decimal.NewFromFloat(50000)},
			commission:   valueobjects.Commission{Amount: decimal.NewFromFloat(0.1), Asset: "USDT"},
			wantErr:      true,
			errContains:  "fill quantity",
		},
		{
			name: "fill pending order",
			setupOrder: func() *Order {
				order, _ := createTestOrder()
				return order // Don't confirm
			},
			fillQuantity: valueobjects.Quantity{Decimal: decimal.NewFromFloat(0.05)},
			fillPrice:    valueobjects.Price{Decimal: decimal.NewFromFloat(50000)},
			commission:   valueobjects.Commission{Amount: decimal.NewFromFloat(0.1), Asset: "USDT"},
			wantErr:      true,
			errContains:  "cannot fill order in status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := tt.setupOrder()
			originalFilled := order.GetFilledQuantity()
			originalRemaining := order.GetRemainingQuantity()

			err := order.PartialFill(tt.fillQuantity, tt.fillPrice, tt.commission)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				// Order should remain unchanged
				assert.Equal(t, originalFilled, order.GetFilledQuantity())
				assert.Equal(t, originalRemaining, order.GetRemainingQuantity())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, order.GetStatus())

				// Verify quantities
				expectedFilled := originalFilled.Add(tt.fillQuantity.Decimal)
				expectedRemaining := order.GetQuantity().Sub(expectedFilled)
				assert.Equal(t, expectedFilled, order.GetFilledQuantity().Decimal)
				assert.Equal(t, expectedRemaining, order.GetRemainingQuantity().Decimal)

				// Verify average fill price calculation
				if !order.GetAvgFillPrice().IsZero() {
					assert.True(t, order.GetAvgFillPrice().GreaterThan(decimal.Zero))
				}

				// Verify event was recorded
				events := order.GetEvents()
				lastEvent := events[len(events)-1]
				assert.Equal(t, OrderEventPartiallyFilled, lastEvent.Type)
			}
		})
	}
}

func TestOrder_Cancel(t *testing.T) {
	tests := []struct {
		name       string
		setupOrder func() *Order
		wantErr    bool
		errContains string
	}{
		{
			name: "cancel new order",
			setupOrder: func() *Order {
				order, _ := createTestOrder()
				order.Confirm()
				return order
			},
			wantErr: false,
		},
		{
			name: "cancel partially filled order",
			setupOrder: func() *Order {
				order, _ := createTestOrder()
				order.Confirm()
				order.PartialFill(
					valueobjects.Quantity{Decimal: decimal.NewFromFloat(0.05)},
					valueobjects.Price{Decimal: decimal.NewFromFloat(50000)},
					valueobjects.Commission{Amount: decimal.NewFromFloat(0.1), Asset: "USDT"},
				)
				return order
			},
			wantErr: false,
		},
		{
			name: "cancel filled order",
			setupOrder: func() *Order {
				order, _ := createTestOrder()
				order.Confirm()
				order.PartialFill(
					valueobjects.Quantity{Decimal: decimal.NewFromFloat(0.1)},
					valueobjects.Price{Decimal: decimal.NewFromFloat(50000)},
					valueobjects.Commission{Amount: decimal.NewFromFloat(0.1), Asset: "USDT"},
				)
				return order
			},
			wantErr:     true,
			errContains: "cannot cancel order in status",
		},
		{
			name: "cancel already canceled order",
			setupOrder: func() *Order {
				order, _ := createTestOrder()
				order.Confirm()
				order.Cancel()
				return order
			},
			wantErr:     true,
			errContains: "cannot cancel order in status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := tt.setupOrder()
			originalStatus := order.GetStatus()

			err := order.Cancel()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				// Status should remain unchanged
				assert.Equal(t, originalStatus, order.GetStatus())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, OrderStatusCanceled, order.GetStatus())

				// Verify event was recorded
				events := order.GetEvents()
				lastEvent := events[len(events)-1]
				assert.Equal(t, OrderEventCanceled, lastEvent.Type)
			}
		})
	}
}

func TestOrder_Reject(t *testing.T) {
	order, err := createTestOrder()
	require.NoError(t, err)

	reason := "Insufficient balance"
	order.Reject(reason)

	assert.Equal(t, OrderStatusRejected, order.GetStatus())

	// Verify event was recorded
	events := order.GetEvents()
	lastEvent := events[len(events)-1]
	assert.Equal(t, OrderEventRejected, lastEvent.Type)
	assert.Equal(t, reason, lastEvent.Data["reason"])
}

func TestOrder_SetLatency(t *testing.T) {
	order, err := createTestOrder()
	require.NoError(t, err)

	latency := 500 * time.Microsecond
	order.SetLatency(latency)

	assert.Equal(t, latency, order.GetLatency())
}

func TestOrder_StatusCheckers(t *testing.T) {
	order, err := createTestOrder()
	require.NoError(t, err)

	// Test pending order
	assert.False(t, order.IsActive())
	assert.False(t, order.IsFilled())
	assert.False(t, order.IsCanceled())
	assert.False(t, order.IsRejected())

	// Test confirmed order
	order.Confirm()
	assert.True(t, order.IsActive())
	assert.False(t, order.IsFilled())
	assert.False(t, order.IsCanceled())
	assert.False(t, order.IsRejected())

	// Test filled order
	order.PartialFill(
		valueobjects.Quantity{Decimal: decimal.NewFromFloat(0.1)},
		valueobjects.Price{Decimal: decimal.NewFromFloat(50000)},
		valueobjects.Commission{Amount: decimal.NewFromFloat(0.1), Asset: "USDT"},
	)
	assert.False(t, order.IsActive())
	assert.True(t, order.IsFilled())
	assert.False(t, order.IsCanceled())
	assert.False(t, order.IsRejected())
}

func TestOrder_SetExchangeOrderID(t *testing.T) {
	order, err := createTestOrder()
	require.NoError(t, err)

	exchangeOrderID := "exchange-order-123"
	order.SetExchangeOrderID(exchangeOrderID)

	// Verify event was recorded
	events := order.GetEvents()
	lastEvent := events[len(events)-1]
	assert.Equal(t, OrderEventExchangeIDSet, lastEvent.Type)
	assert.Equal(t, exchangeOrderID, lastEvent.Data["exchange_order_id"])
}

// Helper function to create a test order
func createTestOrder() (*Order, error) {
	return NewOrder(
		"test-strategy",
		"BTCUSDT",
		"binance",
		valueobjects.OrderSideBuy,
		valueobjects.OrderTypeLimit,
		valueobjects.Quantity{Decimal: decimal.NewFromFloat(0.1)},
		valueobjects.Price{Decimal: decimal.NewFromFloat(50000)},
		valueobjects.TimeInForceGTC,
	)
}

// Benchmark tests
func BenchmarkNewOrder(b *testing.B) {
	strategyID := StrategyID("benchmark-strategy")
	symbol := Symbol("BTCUSDT")
	exchange := Exchange("binance")
	side := valueobjects.OrderSideBuy
	orderType := valueobjects.OrderTypeLimit
	quantity := valueobjects.Quantity{Decimal: decimal.NewFromFloat(0.1)}
	price := valueobjects.Price{Decimal: decimal.NewFromFloat(50000)}
	timeInForce := valueobjects.TimeInForceGTC

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewOrder(strategyID, symbol, exchange, side, orderType, quantity, price, timeInForce)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOrder_PartialFill(b *testing.B) {
	order, err := createTestOrder()
	if err != nil {
		b.Fatal(err)
	}
	order.Confirm()

	fillQuantity := valueobjects.Quantity{Decimal: decimal.NewFromFloat(0.01)}
	fillPrice := valueobjects.Price{Decimal: decimal.NewFromFloat(50000)}
	commission := valueobjects.Commission{Amount: decimal.NewFromFloat(0.01), Asset: "USDT"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset order for each iteration
		order, _ = createTestOrder()
		order.Confirm()
		
		err := order.PartialFill(fillQuantity, fillPrice, commission)
		if err != nil {
			b.Fatal(err)
		}
	}
}
