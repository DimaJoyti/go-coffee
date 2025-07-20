-- Drop views
DROP VIEW IF EXISTS hft_strategy_performance;
DROP VIEW IF EXISTS hft_active_strategies;
DROP VIEW IF EXISTS hft_active_orders;

-- Drop triggers
DROP TRIGGER IF EXISTS update_hft_positions_updated_at ON hft_positions;
DROP TRIGGER IF EXISTS update_hft_strategies_updated_at ON hft_strategies;
DROP TRIGGER IF EXISTS update_hft_orders_updated_at ON hft_orders;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS hft_risk_events;
DROP TABLE IF EXISTS hft_snapshots;
DROP TABLE IF EXISTS hft_events;
DROP TABLE IF EXISTS hft_strategy_events;
DROP TABLE IF EXISTS hft_order_events;
DROP TABLE IF EXISTS hft_market_data;
DROP TABLE IF EXISTS hft_positions;
DROP TABLE IF EXISTS hft_fills;
DROP TABLE IF EXISTS hft_strategies;
DROP TABLE IF EXISTS hft_orders;
