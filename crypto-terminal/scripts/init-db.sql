-- Crypto Terminal Database Schema

-- Create database if not exists
CREATE DATABASE IF NOT EXISTS crypto_terminal;
USE crypto_terminal;

-- Users table (for demo purposes)
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Portfolios table
CREATE TABLE IF NOT EXISTS portfolios (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    total_value DECIMAL(20, 8) DEFAULT 0,
    total_cost DECIMAL(20, 8) DEFAULT 0,
    total_pnl DECIMAL(20, 8) DEFAULT 0,
    total_pnl_percent DECIMAL(10, 4) DEFAULT 0,
    day_change DECIMAL(20, 8) DEFAULT 0,
    day_change_percent DECIMAL(10, 4) DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
);

-- Holdings table
CREATE TABLE IF NOT EXISTS holdings (
    id VARCHAR(36) PRIMARY KEY,
    portfolio_id VARCHAR(36) NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    quantity DECIMAL(20, 8) NOT NULL DEFAULT 0,
    average_price DECIMAL(20, 8) NOT NULL DEFAULT 0,
    current_price DECIMAL(20, 8) NOT NULL DEFAULT 0,
    total_cost DECIMAL(20, 8) NOT NULL DEFAULT 0,
    current_value DECIMAL(20, 8) NOT NULL DEFAULT 0,
    pnl DECIMAL(20, 8) NOT NULL DEFAULT 0,
    pnl_percent DECIMAL(10, 4) NOT NULL DEFAULT 0,
    day_change DECIMAL(20, 8) NOT NULL DEFAULT 0,
    day_change_percent DECIMAL(10, 4) NOT NULL DEFAULT 0,
    allocation_percent DECIMAL(10, 4) NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id) ON DELETE CASCADE,
    UNIQUE KEY unique_portfolio_symbol (portfolio_id, symbol),
    INDEX idx_portfolio_id (portfolio_id),
    INDEX idx_symbol (symbol)
);

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id VARCHAR(36) PRIMARY KEY,
    portfolio_id VARCHAR(36) NOT NULL,
    holding_id VARCHAR(36),
    symbol VARCHAR(20) NOT NULL,
    type ENUM('BUY', 'SELL') NOT NULL,
    quantity DECIMAL(20, 8) NOT NULL,
    price DECIMAL(20, 8) NOT NULL,
    total_amount DECIMAL(20, 8) NOT NULL,
    fee DECIMAL(20, 8) DEFAULT 0,
    exchange VARCHAR(100),
    tx_hash VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id) ON DELETE CASCADE,
    FOREIGN KEY (holding_id) REFERENCES holdings(id) ON DELETE SET NULL,
    INDEX idx_portfolio_id (portfolio_id),
    INDEX idx_symbol (symbol),
    INDEX idx_type (type),
    INDEX idx_created_at (created_at)
);

-- Portfolio historical data table
CREATE TABLE IF NOT EXISTS portfolio_historical_data (
    id VARCHAR(36) PRIMARY KEY,
    portfolio_id VARCHAR(36) NOT NULL,
    date DATE NOT NULL,
    total_value DECIMAL(20, 8) NOT NULL,
    total_cost DECIMAL(20, 8) NOT NULL,
    pnl DECIMAL(20, 8) NOT NULL,
    pnl_percent DECIMAL(10, 4) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id) ON DELETE CASCADE,
    UNIQUE KEY unique_portfolio_date (portfolio_id, date),
    INDEX idx_portfolio_id (portfolio_id),
    INDEX idx_date (date)
);

-- Alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    type ENUM('PRICE', 'VOLUME', 'TECHNICAL', 'NEWS', 'DEFI') NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    condition_operator VARCHAR(20) NOT NULL,
    condition_value DECIMAL(20, 8),
    condition_percentage DECIMAL(10, 4),
    condition_timeframe VARCHAR(10),
    condition_indicator VARCHAR(50),
    condition_parameters JSON,
    is_active BOOLEAN DEFAULT TRUE,
    is_triggered BOOLEAN DEFAULT FALSE,
    trigger_count INT DEFAULT 0,
    max_triggers INT DEFAULT 0,
    cooldown_seconds INT DEFAULT 0,
    last_triggered TIMESTAMP NULL,
    expires_at TIMESTAMP NULL,
    channels JSON,
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_symbol (symbol),
    INDEX idx_type (type),
    INDEX idx_is_active (is_active),
    INDEX idx_is_triggered (is_triggered)
);

-- Alert triggers table
CREATE TABLE IF NOT EXISTS alert_triggers (
    id VARCHAR(36) PRIMARY KEY,
    alert_id VARCHAR(36) NOT NULL,
    trigger_value DECIMAL(20, 8) NOT NULL,
    actual_value DECIMAL(20, 8) NOT NULL,
    message TEXT,
    metadata JSON,
    triggered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    notified_at TIMESTAMP NULL,
    status ENUM('PENDING', 'SENT', 'FAILED') DEFAULT 'PENDING',
    FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE CASCADE,
    INDEX idx_alert_id (alert_id),
    INDEX idx_triggered_at (triggered_at),
    INDEX idx_status (status)
);

-- Alert notifications table
CREATE TABLE IF NOT EXISTS alert_notifications (
    id VARCHAR(36) PRIMARY KEY,
    alert_id VARCHAR(36) NOT NULL,
    trigger_id VARCHAR(36) NOT NULL,
    channel ENUM('EMAIL', 'SMS', 'PUSH', 'WEBHOOK') NOT NULL,
    recipient VARCHAR(255) NOT NULL,
    subject VARCHAR(255),
    message TEXT NOT NULL,
    status ENUM('PENDING', 'SENT', 'DELIVERED', 'FAILED') DEFAULT 'PENDING',
    sent_at TIMESTAMP NULL,
    delivered_at TIMESTAMP NULL,
    error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE CASCADE,
    FOREIGN KEY (trigger_id) REFERENCES alert_triggers(id) ON DELETE CASCADE,
    INDEX idx_alert_id (alert_id),
    INDEX idx_trigger_id (trigger_id),
    INDEX idx_channel (channel),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);

-- Market data cache table
CREATE TABLE IF NOT EXISTS market_data_cache (
    id VARCHAR(36) PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    data_type ENUM('PRICE', 'OHLCV', 'MARKET_DATA', 'INDICATOR') NOT NULL,
    timeframe VARCHAR(10),
    data JSON NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_symbol_type_timeframe (symbol, data_type, timeframe),
    INDEX idx_symbol (symbol),
    INDEX idx_data_type (data_type),
    INDEX idx_expires_at (expires_at)
);

-- Wallet connections table
CREATE TABLE IF NOT EXISTS wallet_connections (
    id VARCHAR(36) PRIMARY KEY,
    portfolio_id VARCHAR(36) NOT NULL,
    wallet_type ENUM('METAMASK', 'COINBASE', 'HARDWARE', 'EXCHANGE') NOT NULL,
    address VARCHAR(255) NOT NULL,
    network VARCHAR(50) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    last_synced TIMESTAMP NULL,
    sync_status ENUM('PENDING', 'SYNCING', 'COMPLETED', 'FAILED') DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id) ON DELETE CASCADE,
    INDEX idx_portfolio_id (portfolio_id),
    INDEX idx_address (address),
    INDEX idx_network (network),
    INDEX idx_is_active (is_active)
);

-- Insert demo user
INSERT IGNORE INTO users (id, email, username, password_hash, first_name, last_name) VALUES
('default-user', 'demo@cryptoterminal.com', 'demo', '$2a$10$dummy.hash.for.demo.purposes', 'Demo', 'User');

-- Insert demo portfolio
INSERT IGNORE INTO portfolios (id, user_id, name, description, total_value, total_cost, total_pnl, total_pnl_percent) VALUES
('portfolio-1', 'default-user', 'Main Portfolio', 'Primary cryptocurrency portfolio', 50000.00000000, 45000.00000000, 5000.00000000, 11.1100);

-- Insert demo holdings
INSERT IGNORE INTO holdings (id, portfolio_id, symbol, name, quantity, average_price, current_price, total_cost, current_value, pnl, pnl_percent, allocation_percent) VALUES
('holding-1', 'portfolio-1', 'BTC', 'Bitcoin', 0.50000000, 45000.00000000, 65000.00000000, 22500.00000000, 32500.00000000, 10000.00000000, 44.4400, 65.0000),
('holding-2', 'portfolio-1', 'ETH', 'Ethereum', 5.50000000, 2500.00000000, 3200.00000000, 13750.00000000, 17600.00000000, 3850.00000000, 28.0000, 35.2000);

-- Insert demo transactions
INSERT IGNORE INTO transactions (id, portfolio_id, holding_id, symbol, type, quantity, price, total_amount, fee, exchange) VALUES
('tx-1', 'portfolio-1', 'holding-1', 'BTC', 'BUY', 0.50000000, 45000.00000000, 22500.00000000, 22.50000000, 'Binance'),
('tx-2', 'portfolio-1', 'holding-2', 'ETH', 'BUY', 5.50000000, 2500.00000000, 13750.00000000, 13.75000000, 'Coinbase');

-- Insert demo alerts
INSERT IGNORE INTO alerts (id, user_id, type, symbol, name, description, condition_operator, condition_value, is_active, channels) VALUES
('alert-1', 'default-user', 'PRICE', 'BTC', 'Bitcoin Price Alert', 'Alert when Bitcoin reaches $70,000', 'ABOVE', 70000.00000000, TRUE, '["EMAIL", "PUSH"]'),
('alert-2', 'default-user', 'TECHNICAL', 'ETH', 'Ethereum RSI Alert', 'Alert when Ethereum RSI goes below 30', 'BELOW', 30.00000000, TRUE, '["PUSH", "WEBHOOK"]');
