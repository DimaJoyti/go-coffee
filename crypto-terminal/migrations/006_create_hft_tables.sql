-- Create HFT Orders table
CREATE TABLE IF NOT EXISTS hft_orders (
    id VARCHAR(255) PRIMARY KEY,
    client_order_id VARCHAR(255),
    strategy_id VARCHAR(255) NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    side VARCHAR(10) NOT NULL CHECK (side IN ('buy', 'sell')),
    type VARCHAR(20) NOT NULL CHECK (type IN ('market', 'limit', 'stop', 'stop_limit', 'ioc', 'fok', 'post')),
    quantity DECIMAL(20, 8) NOT NULL,
    price DECIMAL(20, 8),
    stop_price DECIMAL(20, 8),
    time_in_force VARCHAR(10) NOT NULL CHECK (time_in_force IN ('gtc', 'ioc', 'fok', 'gtd')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'new', 'partially_filled', 'filled', 'canceled', 'rejected', 'expired')),
    filled_quantity DECIMAL(20, 8) DEFAULT 0,
    remaining_quantity DECIMAL(20, 8),
    avg_fill_price DECIMAL(20, 8) DEFAULT 0,
    commission DECIMAL(20, 8) DEFAULT 0,
    commission_asset VARCHAR(10),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    exchange_order_id VARCHAR(255),
    error_message TEXT,
    latency BIGINT DEFAULT 0
);

-- Create indexes for HFT Orders
CREATE INDEX IF NOT EXISTS idx_hft_orders_strategy_id ON hft_orders(strategy_id);
CREATE INDEX IF NOT EXISTS idx_hft_orders_symbol ON hft_orders(symbol);
CREATE INDEX IF NOT EXISTS idx_hft_orders_status ON hft_orders(status);
CREATE INDEX IF NOT EXISTS idx_hft_orders_created_at ON hft_orders(created_at);
CREATE INDEX IF NOT EXISTS idx_hft_orders_exchange ON hft_orders(exchange);

-- Create HFT Fills table
CREATE TABLE IF NOT EXISTS hft_fills (
    id VARCHAR(255) PRIMARY KEY,
    order_id VARCHAR(255) NOT NULL REFERENCES hft_orders(id),
    trade_id VARCHAR(255),
    symbol VARCHAR(50) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    side VARCHAR(10) NOT NULL CHECK (side IN ('buy', 'sell')),
    quantity DECIMAL(20, 8) NOT NULL,
    price DECIMAL(20, 8) NOT NULL,
    commission DECIMAL(20, 8) DEFAULT 0,
    commission_asset VARCHAR(10),
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_maker BOOLEAN DEFAULT false
);

-- Create indexes for HFT Fills
CREATE INDEX IF NOT EXISTS idx_hft_fills_order_id ON hft_fills(order_id);
CREATE INDEX IF NOT EXISTS idx_hft_fills_symbol ON hft_fills(symbol);
CREATE INDEX IF NOT EXISTS idx_hft_fills_timestamp ON hft_fills(timestamp);

-- Create HFT Positions table
CREATE TABLE IF NOT EXISTS hft_positions (
    id VARCHAR(255) PRIMARY KEY,
    strategy_id VARCHAR(255) NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    side VARCHAR(10) NOT NULL CHECK (side IN ('buy', 'sell')),
    size DECIMAL(20, 8) NOT NULL DEFAULT 0,
    entry_price DECIMAL(20, 8) DEFAULT 0,
    mark_price DECIMAL(20, 8) DEFAULT 0,
    unrealized_pnl DECIMAL(20, 8) DEFAULT 0,
    realized_pnl DECIMAL(20, 8) DEFAULT 0,
    margin DECIMAL(20, 8) DEFAULT 0,
    maintenance_margin DECIMAL(20, 8) DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(strategy_id, symbol, exchange)
);

-- Create indexes for HFT Positions
CREATE INDEX IF NOT EXISTS idx_hft_positions_strategy_id ON hft_positions(strategy_id);
CREATE INDEX IF NOT EXISTS idx_hft_positions_symbol ON hft_positions(symbol);
CREATE INDEX IF NOT EXISTS idx_hft_positions_exchange ON hft_positions(exchange);

-- Create HFT Strategies table
CREATE TABLE IF NOT EXISTS hft_strategies (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('market_making', 'arbitrage', 'momentum', 'mean_revert', 'stat_arb', 'custom')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('stopped', 'running', 'paused', 'error')),
    symbols TEXT[], -- Array of symbols
    exchanges TEXT[], -- Array of exchanges
    parameters JSONB,
    risk_limits JSONB,
    performance JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    stopped_at TIMESTAMP
);

-- Create indexes for HFT Strategies
CREATE INDEX IF NOT EXISTS idx_hft_strategies_type ON hft_strategies(type);
CREATE INDEX IF NOT EXISTS idx_hft_strategies_status ON hft_strategies(status);
CREATE INDEX IF NOT EXISTS idx_hft_strategies_created_at ON hft_strategies(created_at);

-- Create HFT Signals table
CREATE TABLE IF NOT EXISTS hft_signals (
    id VARCHAR(255) PRIMARY KEY,
    strategy_id VARCHAR(255) NOT NULL REFERENCES hft_strategies(id),
    symbol VARCHAR(50) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    side VARCHAR(10) NOT NULL CHECK (side IN ('buy', 'sell')),
    strength DECIMAL(10, 4) NOT NULL,
    price DECIMAL(20, 8) NOT NULL,
    quantity DECIMAL(20, 8) NOT NULL,
    confidence DECIMAL(5, 4) NOT NULL,
    reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    executed BOOLEAN DEFAULT false
);

-- Create indexes for HFT Signals
CREATE INDEX IF NOT EXISTS idx_hft_signals_strategy_id ON hft_signals(strategy_id);
CREATE INDEX IF NOT EXISTS idx_hft_signals_symbol ON hft_signals(symbol);
CREATE INDEX IF NOT EXISTS idx_hft_signals_created_at ON hft_signals(created_at);
CREATE INDEX IF NOT EXISTS idx_hft_signals_executed ON hft_signals(executed);

-- Create HFT Risk Events table
CREATE TABLE IF NOT EXISTS hft_risk_events (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    strategy_id VARCHAR(255),
    symbol VARCHAR(50),
    description TEXT NOT NULL,
    data JSONB,
    action VARCHAR(50) NOT NULL,
    resolved BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP
);

-- Create indexes for HFT Risk Events
CREATE INDEX IF NOT EXISTS idx_hft_risk_events_type ON hft_risk_events(type);
CREATE INDEX IF NOT EXISTS idx_hft_risk_events_severity ON hft_risk_events(severity);
CREATE INDEX IF NOT EXISTS idx_hft_risk_events_strategy_id ON hft_risk_events(strategy_id);
CREATE INDEX IF NOT EXISTS idx_hft_risk_events_created_at ON hft_risk_events(created_at);
CREATE INDEX IF NOT EXISTS idx_hft_risk_events_resolved ON hft_risk_events(resolved);

-- Create HFT Market Data Ticks table (for historical analysis)
CREATE TABLE IF NOT EXISTS hft_market_ticks (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    price DECIMAL(20, 8) NOT NULL,
    quantity DECIMAL(20, 8) NOT NULL,
    side VARCHAR(10) NOT NULL CHECK (side IN ('buy', 'sell')),
    bid_price DECIMAL(20, 8),
    bid_quantity DECIMAL(20, 8),
    ask_price DECIMAL(20, 8),
    ask_quantity DECIMAL(20, 8),
    timestamp TIMESTAMP NOT NULL,
    receive_time TIMESTAMP NOT NULL,
    process_time TIMESTAMP NOT NULL,
    latency BIGINT NOT NULL,
    sequence_num BIGINT NOT NULL
);

-- Create indexes for HFT Market Data Ticks
CREATE INDEX IF NOT EXISTS idx_hft_market_ticks_symbol_exchange ON hft_market_ticks(symbol, exchange);
CREATE INDEX IF NOT EXISTS idx_hft_market_ticks_timestamp ON hft_market_ticks(timestamp);
CREATE INDEX IF NOT EXISTS idx_hft_market_ticks_sequence_num ON hft_market_ticks(sequence_num);

-- Create partitioning for market ticks by date (optional, for high volume)
-- This would be implemented based on actual data volume requirements

-- Create triggers to update updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_hft_orders_updated_at BEFORE UPDATE ON hft_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_hft_positions_updated_at BEFORE UPDATE ON hft_positions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_hft_strategies_updated_at BEFORE UPDATE ON hft_strategies FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
