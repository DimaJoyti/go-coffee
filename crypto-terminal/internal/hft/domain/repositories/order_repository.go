package repositories

import (
	"context"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/entities"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/valueobjects"
)

// OrderRepository defines the interface for order persistence
type OrderRepository interface {
	// Basic CRUD operations
	Save(ctx context.Context, order *entities.Order) error
	FindByID(ctx context.Context, id entities.OrderID) (*entities.Order, error)
	Update(ctx context.Context, order *entities.Order) error
	Delete(ctx context.Context, id entities.OrderID) error

	// Query operations
	FindByStrategyID(ctx context.Context, strategyID entities.StrategyID) ([]*entities.Order, error)
	FindBySymbol(ctx context.Context, symbol entities.Symbol) ([]*entities.Order, error)
	FindByExchange(ctx context.Context, exchange entities.Exchange) ([]*entities.Order, error)
	FindByStatus(ctx context.Context, status valueobjects.OrderStatus) ([]*entities.Order, error)
	FindActiveOrders(ctx context.Context) ([]*entities.Order, error)
	FindOrdersInDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entities.Order, error)

	// Complex queries
	FindOrdersByStrategyAndStatus(ctx context.Context, strategyID entities.StrategyID, status valueobjects.OrderStatus) ([]*entities.Order, error)
	FindOrdersBySymbolAndSide(ctx context.Context, symbol entities.Symbol, side valueobjects.OrderSide) ([]*entities.Order, error)
	FindRecentOrders(ctx context.Context, limit int) ([]*entities.Order, error)

	// Aggregation operations
	CountOrdersByStrategy(ctx context.Context, strategyID entities.StrategyID) (int64, error)
	CountOrdersByStatus(ctx context.Context, status valueobjects.OrderStatus) (int64, error)
	GetOrderVolumeByStrategy(ctx context.Context, strategyID entities.StrategyID) (valueobjects.Quantity, error)

	// Event sourcing support
	SaveEvents(ctx context.Context, orderID entities.OrderID, events []valueobjects.OrderEvent) error
	GetEvents(ctx context.Context, orderID entities.OrderID) ([]valueobjects.OrderEvent, error)
	GetEventsSince(ctx context.Context, orderID entities.OrderID, since time.Time) ([]valueobjects.OrderEvent, error)
}

// StrategyRepository defines the interface for strategy persistence
type StrategyRepository interface {
	// Basic CRUD operations
	Save(ctx context.Context, strategy *entities.Strategy) error
	FindByID(ctx context.Context, id entities.StrategyID) (*entities.Strategy, error)
	Update(ctx context.Context, strategy *entities.Strategy) error
	Delete(ctx context.Context, id entities.StrategyID) error

	// Query operations
	FindAll(ctx context.Context) ([]*entities.Strategy, error)
	FindByType(ctx context.Context, strategyType valueobjects.StrategyType) ([]*entities.Strategy, error)
	FindByStatus(ctx context.Context, status valueobjects.StrategyStatus) ([]*entities.Strategy, error)
	FindBySymbol(ctx context.Context, symbol entities.Symbol) ([]*entities.Strategy, error)
	FindByExchange(ctx context.Context, exchange entities.Exchange) ([]*entities.Strategy, error)

	// Complex queries
	FindActiveStrategies(ctx context.Context) ([]*entities.Strategy, error)
	FindStrategiesByTypeAndStatus(ctx context.Context, strategyType valueobjects.StrategyType, status valueobjects.StrategyStatus) ([]*entities.Strategy, error)
	FindStrategiesCreatedAfter(ctx context.Context, after time.Time) ([]*entities.Strategy, error)

	// Performance queries
	FindTopPerformingStrategies(ctx context.Context, limit int) ([]*entities.Strategy, error)
	FindStrategiesByPerformanceMetric(ctx context.Context, metric string, threshold float64) ([]*entities.Strategy, error)

	// Event sourcing support
	SaveEvents(ctx context.Context, strategyID entities.StrategyID, events []valueobjects.StrategyEvent) error
	GetEvents(ctx context.Context, strategyID entities.StrategyID) ([]valueobjects.StrategyEvent, error)
	GetEventsSince(ctx context.Context, strategyID entities.StrategyID, since time.Time) ([]valueobjects.StrategyEvent, error)
}

// FillRepository defines the interface for fill/execution persistence
type FillRepository interface {
	// Basic CRUD operations
	Save(ctx context.Context, fill *Fill) error
	FindByID(ctx context.Context, id string) (*Fill, error)
	Update(ctx context.Context, fill *Fill) error
	Delete(ctx context.Context, id string) error

	// Query operations
	FindByOrderID(ctx context.Context, orderID entities.OrderID) ([]*Fill, error)
	FindByStrategyID(ctx context.Context, strategyID entities.StrategyID) ([]*Fill, error)
	FindBySymbol(ctx context.Context, symbol entities.Symbol) ([]*Fill, error)
	FindByExchange(ctx context.Context, exchange entities.Exchange) ([]*Fill, error)
	FindInDateRange(ctx context.Context, startDate, endDate time.Time) ([]*Fill, error)

	// Aggregation operations
	GetTotalVolumeByStrategy(ctx context.Context, strategyID entities.StrategyID) (valueobjects.Quantity, error)
	GetTotalPnLByStrategy(ctx context.Context, strategyID entities.StrategyID) (valueobjects.Price, error)
	GetFillCountByStrategy(ctx context.Context, strategyID entities.StrategyID) (int64, error)
}

// PositionRepository defines the interface for position persistence
type PositionRepository interface {
	// Basic CRUD operations
	Save(ctx context.Context, position *Position) error
	FindByID(ctx context.Context, id string) (*Position, error)
	Update(ctx context.Context, position *Position) error
	Delete(ctx context.Context, id string) error

	// Query operations
	FindByStrategyID(ctx context.Context, strategyID entities.StrategyID) ([]*Position, error)
	FindBySymbol(ctx context.Context, symbol entities.Symbol) ([]*Position, error)
	FindByExchange(ctx context.Context, exchange entities.Exchange) ([]*Position, error)
	FindActivePositions(ctx context.Context) ([]*Position, error)

	// Complex queries
	FindPositionsByStrategyAndSymbol(ctx context.Context, strategyID entities.StrategyID, symbol entities.Symbol) (*Position, error)
	FindLargestPositions(ctx context.Context, limit int) ([]*Position, error)
	FindPositionsWithUnrealizedPnL(ctx context.Context, threshold valueobjects.Price) ([]*Position, error)

	// Aggregation operations
	GetTotalExposureByStrategy(ctx context.Context, strategyID entities.StrategyID) (valueobjects.Price, error)
	GetNetPositionBySymbol(ctx context.Context, symbol entities.Symbol) (valueobjects.Quantity, error)
}

// MarketDataRepository defines the interface for market data persistence
type MarketDataRepository interface {
	// Tick data operations
	SaveTick(ctx context.Context, tick *MarketDataTick) error
	GetLatestTick(ctx context.Context, symbol entities.Symbol, exchange entities.Exchange) (*MarketDataTick, error)
	GetTicksInRange(ctx context.Context, symbol entities.Symbol, exchange entities.Exchange, startTime, endTime time.Time) ([]*MarketDataTick, error)
	GetRecentTicks(ctx context.Context, symbol entities.Symbol, exchange entities.Exchange, limit int) ([]*MarketDataTick, error)

	// Order book operations
	SaveOrderBook(ctx context.Context, orderBook *OrderBook) error
	GetLatestOrderBook(ctx context.Context, symbol entities.Symbol, exchange entities.Exchange) (*OrderBook, error)
	GetOrderBookHistory(ctx context.Context, symbol entities.Symbol, exchange entities.Exchange, limit int) ([]*OrderBook, error)

	// Aggregated data operations
	GetOHLCV(ctx context.Context, symbol entities.Symbol, exchange entities.Exchange, interval string, startTime, endTime time.Time) ([]*OHLCV, error)
	GetVWAP(ctx context.Context, symbol entities.Symbol, exchange entities.Exchange, startTime, endTime time.Time) (valueobjects.Price, error)
}

// EventStore defines the interface for event sourcing
type EventStore interface {
	// Event operations
	SaveEvents(ctx context.Context, aggregateID string, events []DomainEvent) error
	GetEvents(ctx context.Context, aggregateID string) ([]DomainEvent, error)
	GetEventsSince(ctx context.Context, aggregateID string, since time.Time) ([]DomainEvent, error)
	GetEventsFromVersion(ctx context.Context, aggregateID string, fromVersion int) ([]DomainEvent, error)

	// Snapshot operations
	SaveSnapshot(ctx context.Context, snapshot Snapshot) error
	GetLatestSnapshot(ctx context.Context, aggregateID string) (*Snapshot, error)

	// Stream operations
	GetEventStream(ctx context.Context, streamName string) (<-chan DomainEvent, error)
	PublishEvent(ctx context.Context, event DomainEvent) error
}

// Supporting types for repositories

// Fill represents a trade execution
type Fill struct {
	ID              string                    `json:"id"`
	OrderID         entities.OrderID         `json:"order_id"`
	TradeID         string                    `json:"trade_id"`
	Symbol          entities.Symbol          `json:"symbol"`
	Exchange        entities.Exchange        `json:"exchange"`
	Side            valueobjects.OrderSide   `json:"side"`
	Quantity        valueobjects.Quantity    `json:"quantity"`
	Price           valueobjects.Price       `json:"price"`
	Commission      valueobjects.Commission  `json:"commission"`
	Timestamp       time.Time                `json:"timestamp"`
	IsMaker         bool                     `json:"is_maker"`
}

// Position represents a trading position
type Position struct {
	ID               string                   `json:"id"`
	StrategyID       entities.StrategyID     `json:"strategy_id"`
	Symbol           entities.Symbol         `json:"symbol"`
	Exchange         entities.Exchange       `json:"exchange"`
	Side             valueobjects.OrderSide  `json:"side"`
	Size             valueobjects.Quantity   `json:"size"`
	EntryPrice       valueobjects.Price      `json:"entry_price"`
	MarkPrice        valueobjects.Price      `json:"mark_price"`
	UnrealizedPnL    valueobjects.Price      `json:"unrealized_pnl"`
	RealizedPnL      valueobjects.Price      `json:"realized_pnl"`
	Margin           valueobjects.Price      `json:"margin"`
	MaintenanceMargin valueobjects.Price     `json:"maintenance_margin"`
	CreatedAt        time.Time               `json:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
}

// MarketDataTick represents ultra-low latency market data
type MarketDataTick struct {
	Symbol       entities.Symbol         `json:"symbol"`
	Exchange     entities.Exchange       `json:"exchange"`
	Price        valueobjects.Price      `json:"price"`
	Quantity     valueobjects.Quantity   `json:"quantity"`
	Side         valueobjects.OrderSide  `json:"side"`
	BidPrice     valueobjects.Price      `json:"bid_price"`
	BidQuantity  valueobjects.Quantity   `json:"bid_quantity"`
	AskPrice     valueobjects.Price      `json:"ask_price"`
	AskQuantity  valueobjects.Quantity   `json:"ask_quantity"`
	Timestamp    time.Time               `json:"timestamp"`
	ReceiveTime  time.Time               `json:"receive_time"`
	ProcessTime  time.Time               `json:"process_time"`
	Latency      time.Duration           `json:"latency"`
	SequenceNum  uint64                  `json:"sequence_num"`
}

// OrderBook represents the current order book state
type OrderBook struct {
	Symbol      entities.Symbol    `json:"symbol"`
	Exchange    entities.Exchange  `json:"exchange"`
	Bids        []OrderBookLevel   `json:"bids"`
	Asks        []OrderBookLevel   `json:"asks"`
	Timestamp   time.Time          `json:"timestamp"`
	ReceiveTime time.Time          `json:"receive_time"`
	SequenceNum uint64             `json:"sequence_num"`
}

// OrderBookLevel represents a level in the order book
type OrderBookLevel struct {
	Price    valueobjects.Price    `json:"price"`
	Quantity valueobjects.Quantity `json:"quantity"`
	Count    int                   `json:"count"`
}

// OHLCV represents OHLCV candlestick data
type OHLCV struct {
	Symbol    entities.Symbol       `json:"symbol"`
	Exchange  entities.Exchange     `json:"exchange"`
	Timestamp time.Time             `json:"timestamp"`
	Open      valueobjects.Price    `json:"open"`
	High      valueobjects.Price    `json:"high"`
	Low       valueobjects.Price    `json:"low"`
	Close     valueobjects.Price    `json:"close"`
	Volume    valueobjects.Quantity `json:"volume"`
}

// DomainEvent represents a domain event for event sourcing
type DomainEvent struct {
	ID           string                 `json:"id"`
	AggregateID  string                 `json:"aggregate_id"`
	EventType    string                 `json:"event_type"`
	EventData    map[string]interface{} `json:"event_data"`
	Timestamp    time.Time              `json:"timestamp"`
	Version      int                    `json:"version"`
}

// Snapshot represents an aggregate snapshot for event sourcing
type Snapshot struct {
	AggregateID string                 `json:"aggregate_id"`
	Data        map[string]interface{} `json:"data"`
	Version     int                    `json:"version"`
	Timestamp   time.Time              `json:"timestamp"`
}
