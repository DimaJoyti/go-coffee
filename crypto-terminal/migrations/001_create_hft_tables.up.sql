-- Create HFT Orders table
CREATE TABLE IF NOT EXISTS hft_orders (
    id UUID PRIMARY KEY,
    client_order_id VARCHAR(255),
    strategy_id VARCHAR(255) NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    side VARCHAR(10) NOT NULL CHECK (side IN ('buy', 'sell')),
    type VARCHAR(20) NOT NULL CHECK (type IN ('market', 'limit', 'stop', 'stop_limit', 'ioc', 'fok', 'post')),
    quantity DECIMAL(20, 8) NOT NULL CHECK (quantity > 0),
    price DECIMAL(20, 8) NOT NULL DEFAULT 0,
    stop_price DECIMAL(20, 8),
    time_in_force VARCHAR(10) NOT NULL CHECK (time_in_force IN ('gtc', 'ioc', 'fok', 'gtd')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'new', 'partially_filled', 'filled', 'canceled', 'rejected', 'expired')),
    filled_quantity DECIMAL(20, 8) NOT NULL DEFAULT 0,
    remaining_quantity DECIMAL(20, 8) NOT NULL,
    avg_fill_price DECIMAL(20, 8) NOT NULL DEFAULT 0,
    commission_amount DECIMAL(20, 8) NOT NULL DEFAULT 0,
    commission_asset VARCHAR(10) NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    exchange_order_id VARCHAR(255),
    error_message TEXT,
    latency BIGINT NOT NULL DEFAULT 0 -- in nanoseconds
);

-- Create indexes for hft_orders
CREATE INDEX IF NOT EXISTS idx_hft_orders_strategy_id ON hft_orders(strategy_id);
CREATE INDEX IF NOT EXISTS idx_hft_orders_symbol ON hft_orders(symbol);
CREATE INDEX IF NOT EXISTS idx_hft_orders_exchange ON hft_orders(exchange);
CREATE INDEX IF NOT EXISTS idx_hft_orders_status ON hft_orders(status);
CREATE INDEX IF NOT EXISTS idx_hft_orders_created_at ON hft_orders(created_at);
CREATE INDEX IF NOT EXISTS idx_hft_orders_updated_at ON hft_orders(updated_at);
CREATE INDEX IF NOT EXISTS idx_hft_orders_strategy_status ON hft_orders(strategy_id, status);
CREATE INDEX IF NOT EXISTS idx_hft_orders_symbol_side ON hft_orders(symbol, side);
CREATE INDEX IF NOT EXISTS idx_hft_orders_exchange_order_id ON hft_orders(exchange_order_id);

-- Create HFT Strategies table
CREATE TABLE IF NOT EXISTS hft_strategies (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('market_making', 'arbitrage', 'momentum', 'mean_revert', 'stat_arb', 'custom')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('stopped', 'running', 'paused', 'error')),
    symbols TEXT[] NOT NULL,
    exchanges TEXT[] NOT NULL,
    parameters JSONB NOT NULL DEFAULT '{}',
    
    -- Risk limits
    max_position_size DECIMAL(20, 8) NOT NULL DEFAULT 0,
    max_daily_loss DECIMAL(20, 8) NOT NULL DEFAULT 0,
    max_drawdown DECIMAL(10, 4) NOT NULL DEFAULT 0 CHECK (max_drawdown >= 0 AND max_drawdown <= 1),
    max_order_size DECIMAL(20, 8) NOT NULL DEFAULT 0,
    max_orders_per_second INTEGER NOT NULL DEFAULT 0,
    max_exposure DECIMAL(20, 8) NOT NULL DEFAULT 0,
    stop_loss_percent DECIMAL(10, 4) NOT NULL DEFAULT 0 CHECK (stop_loss_percent >= 0 AND stop_loss_percent <= 1),
    take_profit_percent DECIMAL(10, 4) NOT NULL DEFAULT 0,
    
    -- Performance metrics
    total_pnl DECIMAL(20, 8) NOT NULL DEFAULT 0,
    daily_pnl DECIMAL(20, 8) NOT NULL DEFAULT 0,
    total_trades BIGINT NOT NULL DEFAULT 0,
    winning_trades BIGINT NOT NULL DEFAULT 0,
    losing_trades BIGINT NOT NULL DEFAULT 0,
    win_rate DECIMAL(10, 4) NOT NULL DEFAULT 0,
    avg_win DECIMAL(20, 8) NOT NULL DEFAULT 0,
    avg_loss DECIMAL(20, 8) NOT NULL DEFAULT 0,
    profit_factor DECIMAL(10, 4) NOT NULL DEFAULT 0,
    sharpe_ratio DECIMAL(10, 4) NOT NULL DEFAULT 0,
    max_drawdown_realized DECIMAL(10, 4) NOT NULL DEFAULT 0,
    volume_traded DECIMAL(20, 8) NOT NULL DEFAULT 0,
    avg_latency BIGINT NOT NULL DEFAULT 0, -- in nanoseconds
    performance_last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    stopped_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for hft_strategies
CREATE INDEX IF NOT EXISTS idx_hft_strategies_type ON hft_strategies(type);
CREATE INDEX IF NOT EXISTS idx_hft_strategies_status ON hft_strategies(status);
CREATE INDEX IF NOT EXISTS idx_hft_strategies_created_at ON hft_strategies(created_at);
CREATE INDEX IF NOT EXISTS idx_hft_strategies_type_status ON hft_strategies(type, status);
CREATE INDEX IF NOT EXISTS idx_hft_strategies_symbols ON hft_strategies USING GIN(symbols);
CREATE INDEX IF NOT EXISTS idx_hft_strategies_exchanges ON hft_strategies USING GIN(exchanges);
CREATE INDEX IF NOT EXISTS idx_hft_strategies_total_pnl ON hft_strategies(total_pnl DESC);
CREATE INDEX IF NOT EXISTS idx_hft_strategies_sharpe_ratio ON hft_strategies(sharpe_ratio DESC);

-- Create HFT Fills table
CREATE TABLE IF NOT EXISTS hft_fills (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES hft_orders(id) ON DELETE CASCADE,
    trade_id VARCHAR(255) NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    side VARCHAR(10) NOT NULL CHECK (side IN ('buy', 'sell')),
    quantity DECIMAL(20, 8) NOT NULL CHECK (quantity > 0),
    price DECIMAL(20, 8) NOT NULL CHECK (price > 0),
    commission_amount DECIMAL(20, 8) NOT NULL DEFAULT 0,
    commission_asset VARCHAR(10) NOT NULL DEFAULT '',
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    is_maker BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for hft_fills
CREATE INDEX IF NOT EXISTS idx_hft_fills_order_id ON hft_fills(order_id);
CREATE INDEX IF NOT EXISTS idx_hft_fills_symbol ON hft_fills(symbol);
CREATE INDEX IF NOT EXISTS idx_hft_fills_exchange ON hft_fills(exchange);
CREATE INDEX IF NOT EXISTS idx_hft_fills_timestamp ON hft_fills(timestamp);
CREATE INDEX IF NOT EXISTS idx_hft_fills_trade_id ON hft_fills(trade_id);

-- Create HFT Positions table
CREATE TABLE IF NOT EXISTS hft_positions (
    id UUID PRIMARY KEY,
    strategy_id VARCHAR(255) NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    side VARCHAR(10) NOT NULL CHECK (side IN ('buy', 'sell', 'long', 'short')),
    size DECIMAL(20, 8) NOT NULL,
    entry_price DECIMAL(20, 8) NOT NULL,
    mark_price DECIMAL(20, 8) NOT NULL DEFAULT 0,
    unrealized_pnl DECIMAL(20, 8) NOT NULL DEFAULT 0,
    realized_pnl DECIMAL(20, 8) NOT NULL DEFAULT 0,
    margin DECIMAL(20, 8) NOT NULL DEFAULT 0,
    maintenance_margin DECIMAL(20, 8) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(strategy_id, symbol, exchange)
);

-- Create indexes for hft_positions
CREATE INDEX IF NOT EXISTS idx_hft_positions_strategy_id ON hft_positions(strategy_id);
CREATE INDEX IF NOT EXISTS idx_hft_positions_symbol ON hft_positions(symbol);
CREATE INDEX IF NOT EXISTS idx_hft_positions_exchange ON hft_positions(exchange);
CREATE INDEX IF NOT EXISTS idx_hft_positions_strategy_symbol ON hft_positions(strategy_id, symbol);
CREATE INDEX IF NOT EXISTS idx_hft_positions_unrealized_pnl ON hft_positions(unrealized_pnl);

-- Create HFT Market Data table
CREATE TABLE IF NOT EXISTS hft_market_data (
    id UUID PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    price DECIMAL(20, 8) NOT NULL,
    quantity DECIMAL(20, 8) NOT NULL,
    side VARCHAR(10) CHECK (side IN ('buy', 'sell')),
    bid_price DECIMAL(20, 8),
    bid_quantity DECIMAL(20, 8),
    ask_price DECIMAL(20, 8),
    ask_quantity DECIMAL(20, 8),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    receive_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    process_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    latency BIGINT NOT NULL DEFAULT 0, -- in nanoseconds
    sequence_num BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for hft_market_data
CREATE INDEX IF NOT EXISTS idx_hft_market_data_symbol ON hft_market_data(symbol);
CREATE INDEX IF NOT EXISTS idx_hft_market_data_exchange ON hft_market_data(exchange);
CREATE INDEX IF NOT EXISTS idx_hft_market_data_timestamp ON hft_market_data(timestamp);
CREATE INDEX IF NOT EXISTS idx_hft_market_data_symbol_exchange ON hft_market_data(symbol, exchange);
CREATE INDEX IF NOT EXISTS idx_hft_market_data_symbol_timestamp ON hft_market_data(symbol, timestamp);

-- Create HFT Order Events table for event sourcing
CREATE TABLE IF NOT EXISTS hft_order_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for hft_order_events
CREATE INDEX IF NOT EXISTS idx_hft_order_events_order_id ON hft_order_events(order_id);
CREATE INDEX IF NOT EXISTS idx_hft_order_events_timestamp ON hft_order_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_hft_order_events_event_type ON hft_order_events(event_type);

-- Create HFT Strategy Events table for event sourcing
CREATE TABLE IF NOT EXISTS hft_strategy_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    strategy_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for hft_strategy_events
CREATE INDEX IF NOT EXISTS idx_hft_strategy_events_strategy_id ON hft_strategy_events(strategy_id);
CREATE INDEX IF NOT EXISTS idx_hft_strategy_events_timestamp ON hft_strategy_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_hft_strategy_events_event_type ON hft_strategy_events(event_type);

-- Create HFT Events table for general event sourcing
CREATE TABLE IF NOT EXISTS hft_events (
    id UUID PRIMARY KEY,
    aggregate_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    event_data JSONB NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    version INTEGER NOT NULL,
    stream_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(aggregate_id, version)
);

-- Create indexes for hft_events
CREATE INDEX IF NOT EXISTS idx_hft_events_aggregate_id ON hft_events(aggregate_id);
CREATE INDEX IF NOT EXISTS idx_hft_events_stream_name ON hft_events(stream_name);
CREATE INDEX IF NOT EXISTS idx_hft_events_timestamp ON hft_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_hft_events_event_type ON hft_events(event_type);

-- Create HFT Snapshots table for event sourcing
CREATE TABLE IF NOT EXISTS hft_snapshots (
    aggregate_id VARCHAR(255) PRIMARY KEY,
    data JSONB NOT NULL,
    version INTEGER NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for hft_snapshots
CREATE INDEX IF NOT EXISTS idx_hft_snapshots_version ON hft_snapshots(version);
CREATE INDEX IF NOT EXISTS idx_hft_snapshots_timestamp ON hft_snapshots(timestamp);

-- Create HFT Risk Events table
CREATE TABLE IF NOT EXISTS hft_risk_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    strategy_id VARCHAR(255) NOT NULL,
    risk_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('info', 'warning', 'violation', 'critical')),
    message TEXT NOT NULL,
    data JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for hft_risk_events
CREATE INDEX IF NOT EXISTS idx_hft_risk_events_strategy_id ON hft_risk_events(strategy_id);
CREATE INDEX IF NOT EXISTS idx_hft_risk_events_risk_type ON hft_risk_events(risk_type);
CREATE INDEX IF NOT EXISTS idx_hft_risk_events_severity ON hft_risk_events(severity);
CREATE INDEX IF NOT EXISTS idx_hft_risk_events_timestamp ON hft_risk_events(timestamp);

-- Create triggers for updated_at columns
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_hft_orders_updated_at BEFORE UPDATE ON hft_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_hft_strategies_updated_at BEFORE UPDATE ON hft_strategies FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_hft_positions_updated_at BEFORE UPDATE ON hft_positions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create views for common queries
CREATE OR REPLACE VIEW hft_active_orders AS
SELECT * FROM hft_orders 
WHERE status IN ('new', 'partially_filled')
ORDER BY created_at DESC;

CREATE OR REPLACE VIEW hft_active_strategies AS
SELECT * FROM hft_strategies 
WHERE status IN ('running', 'paused')
ORDER BY created_at DESC;

CREATE OR REPLACE VIEW hft_strategy_performance AS
SELECT 
    id,
    name,
    type,
    status,
    total_pnl,
    daily_pnl,
    total_trades,
    win_rate,
    sharpe_ratio,
    max_drawdown_realized,
    volume_traded,
    avg_latency,
    performance_last_updated
FROM hft_strategies
ORDER BY total_pnl DESC;

-- Add comments for documentation
COMMENT ON TABLE hft_orders IS 'High-frequency trading orders with ultra-low latency tracking';
COMMENT ON TABLE hft_strategies IS 'Trading strategies with performance metrics and risk limits';
COMMENT ON TABLE hft_fills IS 'Order execution fills with detailed trade information';
COMMENT ON TABLE hft_positions IS 'Current trading positions by strategy and symbol';
COMMENT ON TABLE hft_market_data IS 'Real-time market data with latency measurements';
COMMENT ON TABLE hft_events IS 'Event sourcing store for domain events';
COMMENT ON TABLE hft_snapshots IS 'Aggregate snapshots for event sourcing optimization';
COMMENT ON TABLE hft_risk_events IS 'Risk management events and violations';

COMMENT ON COLUMN hft_orders.latency IS 'Order processing latency in nanoseconds';
COMMENT ON COLUMN hft_strategies.avg_latency IS 'Average strategy execution latency in nanoseconds';
COMMENT ON COLUMN hft_market_data.latency IS 'Market data processing latency in nanoseconds';
