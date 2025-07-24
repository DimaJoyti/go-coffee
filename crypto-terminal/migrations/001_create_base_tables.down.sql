-- Drop base tables for crypto terminal

-- Drop triggers
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_portfolios_updated_at ON portfolios;
DROP TRIGGER IF EXISTS update_portfolio_holdings_updated_at ON portfolio_holdings;
DROP TRIGGER IF EXISTS update_alerts_updated_at ON alerts;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_portfolios_user_id;
DROP INDEX IF EXISTS idx_portfolio_holdings_portfolio_id;
DROP INDEX IF EXISTS idx_portfolio_holdings_symbol;
DROP INDEX IF EXISTS idx_transactions_portfolio_id;
DROP INDEX IF EXISTS idx_transactions_symbol;
DROP INDEX IF EXISTS idx_transactions_date;
DROP INDEX IF EXISTS idx_alerts_user_id;
DROP INDEX IF EXISTS idx_alerts_symbol;
DROP INDEX IF EXISTS idx_alerts_active;
DROP INDEX IF EXISTS idx_alert_triggers_alert_id;
DROP INDEX IF EXISTS idx_market_data_cache_symbol;
DROP INDEX IF EXISTS idx_market_data_cache_expires;
DROP INDEX IF EXISTS idx_price_history_symbol;
DROP INDEX IF EXISTS idx_price_history_timestamp;
DROP INDEX IF EXISTS idx_trading_signals_symbol;
DROP INDEX IF EXISTS idx_trading_signals_created;
DROP INDEX IF EXISTS idx_news_articles_published;
DROP INDEX IF EXISTS idx_news_articles_symbols;

-- Drop tables in reverse order of creation
DROP TABLE IF EXISTS news_articles;
DROP TABLE IF EXISTS trading_signals;
DROP TABLE IF EXISTS price_history;
DROP TABLE IF EXISTS market_data_cache;
DROP TABLE IF EXISTS alert_triggers;
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS portfolio_holdings;
DROP TABLE IF EXISTS portfolios;
DROP TABLE IF EXISTS users;
