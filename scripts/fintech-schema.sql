-- Fintech Platform Database Schema
-- This script creates the complete database schema for all five modules:
-- Accounts, Payments, Yield, Trading, and Cards

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create schemas for better organization
CREATE SCHEMA IF NOT EXISTS accounts;
CREATE SCHEMA IF NOT EXISTS payments;
CREATE SCHEMA IF NOT EXISTS yield_farming;
CREATE SCHEMA IF NOT EXISTS trading;
CREATE SCHEMA IF NOT EXISTS cards;

-- Set search path
SET search_path TO accounts, payments, yield_farming, trading, cards, public;

-- ============================================================================
-- ACCOUNTS MODULE
-- ============================================================================

-- Accounts table
CREATE TABLE IF NOT EXISTS accounts.accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(50),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    date_of_birth DATE,
    nationality VARCHAR(3), -- ISO 3166-1 alpha-3
    country VARCHAR(3) NOT NULL, -- ISO 3166-1 alpha-3
    state VARCHAR(100),
    city VARCHAR(100),
    address TEXT,
    postal_code VARCHAR(20),
    account_type VARCHAR(20) NOT NULL CHECK (account_type IN ('personal', 'business', 'enterprise')),
    account_status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (account_status IN ('active', 'inactive', 'suspended', 'closed', 'pending')),
    kyc_status VARCHAR(20) NOT NULL DEFAULT 'not_started' CHECK (kyc_status IN ('not_started', 'pending', 'in_review', 'approved', 'rejected', 'expired')),
    kyc_level VARCHAR(20) NOT NULL DEFAULT 'none' CHECK (kyc_level IN ('none', 'basic', 'standard', 'enhanced')),
    risk_score DECIMAL(3,2) DEFAULT 0.00,
    compliance_flags TEXT[],
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    two_factor_method VARCHAR(20),
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_login_ip INET,
    failed_login_count INTEGER DEFAULT 0,
    account_limits JSONB,
    notification_preferences JSONB,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Sessions table
CREATE TABLE IF NOT EXISTS accounts.sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts.accounts(id) ON DELETE CASCADE,
    device_id VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    location VARCHAR(255),
    session_token VARCHAR(255) UNIQUE NOT NULL,
    refresh_token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Security events table
CREATE TABLE IF NOT EXISTS accounts.security_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID REFERENCES accounts.accounts(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    description TEXT NOT NULL,
    ip_address INET,
    user_agent TEXT,
    location VARCHAR(255),
    resolved BOOLEAN DEFAULT FALSE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- KYC documents table
CREATE TABLE IF NOT EXISTS accounts.kyc_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts.accounts(id) ON DELETE CASCADE,
    document_type VARCHAR(50) NOT NULL,
    document_url TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'expired')),
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    verified_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}'
);

-- ============================================================================
-- PAYMENTS MODULE
-- ============================================================================

-- Payments table
CREATE TABLE IF NOT EXISTS payments.payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts.accounts(id),
    wallet_id UUID,
    merchant_id UUID,
    order_id VARCHAR(255),
    payment_method VARCHAR(50) NOT NULL CHECK (payment_method IN ('crypto', 'card', 'bank_transfer', 'wallet', 'stablecoin')),
    currency VARCHAR(10) NOT NULL,
    amount DECIMAL(36,18) NOT NULL,
    fee_amount DECIMAL(36,18) DEFAULT 0,
    net_amount DECIMAL(36,18) NOT NULL,
    exchange_rate DECIMAL(36,18),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled', 'refunded', 'expired', 'on_hold')),
    payment_type VARCHAR(20) NOT NULL CHECK (payment_type IN ('inbound', 'outbound', 'internal', 'refund')),
    description TEXT,
    reference VARCHAR(255),
    transaction_hash VARCHAR(255),
    block_number BIGINT,
    confirmations INTEGER DEFAULT 0,
    network VARCHAR(50),
    from_address VARCHAR(255),
    to_address VARCHAR(255),
    gas_used DECIMAL(36,18),
    gas_price DECIMAL(36,18),
    risk_score DECIMAL(3,2) DEFAULT 0.00,
    fraud_flags TEXT[],
    processed_at TIMESTAMP WITH TIME ZONE,
    settled_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Payment intents table
CREATE TABLE IF NOT EXISTS payments.payment_intents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts.accounts(id),
    merchant_id UUID,
    amount DECIMAL(36,18) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    payment_methods TEXT[] NOT NULL,
    description TEXT,
    return_url TEXT,
    cancel_url TEXT,
    webhook_url TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'created' CHECK (status IN ('created', 'processing', 'succeeded', 'failed', 'cancelled', 'expired')),
    client_secret VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Refunds table
CREATE TABLE IF NOT EXISTS payments.refunds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_id UUID NOT NULL REFERENCES payments.payments(id),
    account_id UUID NOT NULL REFERENCES accounts.accounts(id),
    amount DECIMAL(36,18) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    reason VARCHAR(50) NOT NULL CHECK (reason IN ('requested', 'fraud', 'duplicate', 'error', 'chargeback', 'cancellation')),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    description TEXT,
    processed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Fee structures table
CREATE TABLE IF NOT EXISTS payments.fee_structures (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    payment_method VARCHAR(50) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    fee_type VARCHAR(20) NOT NULL CHECK (fee_type IN ('fixed', 'percentage', 'tiered', 'hybrid')),
    fixed_fee DECIMAL(36,18) DEFAULT 0,
    percentage_fee DECIMAL(5,4) DEFAULT 0,
    min_fee DECIMAL(36,18) DEFAULT 0,
    max_fee DECIMAL(36,18),
    tier_rules JSONB,
    is_active BOOLEAN DEFAULT TRUE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================================
-- YIELD FARMING MODULE
-- ============================================================================

-- Protocols table
CREATE TABLE IF NOT EXISTS yield_farming.protocols (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    website VARCHAR(255),
    logo_url VARCHAR(255),
    category VARCHAR(50) NOT NULL CHECK (category IN ('dex', 'lending', 'staking', 'yield_farm', 'vault', 'insurance', 'derivatives')),
    network VARCHAR(50) NOT NULL,
    contract_address VARCHAR(255),
    tvl DECIMAL(36,18) DEFAULT 0,
    volume_24h DECIMAL(36,18) DEFAULT 0,
    average_apy DECIMAL(8,4) DEFAULT 0,
    risk_score DECIMAL(3,2) DEFAULT 0.00,
    is_active BOOLEAN DEFAULT TRUE,
    is_audited BOOLEAN DEFAULT FALSE,
    audit_reports TEXT[],
    supported_tokens TEXT[],
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Pools table
CREATE TABLE IF NOT EXISTS yield_farming.pools (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    protocol_id UUID NOT NULL REFERENCES yield_farming.protocols(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    pool_type VARCHAR(50) NOT NULL CHECK (pool_type IN ('staking', 'liquidity_mining', 'lending', 'vault', 'farming')),
    token_pair TEXT[],
    reward_tokens TEXT[],
    apy DECIMAL(8,4) DEFAULT 0,
    apr DECIMAL(8,4) DEFAULT 0,
    tvl DECIMAL(36,18) DEFAULT 0,
    volume_24h DECIMAL(36,18) DEFAULT 0,
    min_deposit DECIMAL(36,18) DEFAULT 0,
    max_deposit DECIMAL(36,18),
    lock_period INTERVAL,
    withdrawal_fee DECIMAL(5,4) DEFAULT 0,
    performance_fee DECIMAL(5,4) DEFAULT 0,
    risk_score DECIMAL(3,2) DEFAULT 0.00,
    is_active BOOLEAN DEFAULT TRUE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Yield positions table
CREATE TABLE IF NOT EXISTS yield_farming.positions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts.accounts(id),
    wallet_id UUID,
    protocol_id UUID NOT NULL REFERENCES yield_farming.protocols(id),
    pool_id UUID NOT NULL REFERENCES yield_farming.pools(id),
    position_type VARCHAR(50) NOT NULL CHECK (position_type IN ('staking', 'liquidity_mining', 'lending', 'farming', 'vault')),
    strategy VARCHAR(255),
    token_address VARCHAR(255) NOT NULL,
    token_symbol VARCHAR(20) NOT NULL,
    amount DECIMAL(36,18) NOT NULL,
    usd_value DECIMAL(36,18),
    entry_price DECIMAL(36,18),
    current_price DECIMAL(36,18),
    apy DECIMAL(8,4) DEFAULT 0,
    apr DECIMAL(8,4) DEFAULT 0,
    daily_rewards DECIMAL(36,18) DEFAULT 0,
    total_rewards DECIMAL(36,18) DEFAULT 0,
    claimed_rewards DECIMAL(36,18) DEFAULT 0,
    pending_rewards DECIMAL(36,18) DEFAULT 0,
    impermanent_loss DECIMAL(36,18) DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'locked', 'unlocking', 'closed', 'error')),
    auto_compound BOOLEAN DEFAULT FALSE,
    lock_period INTERVAL,
    unlock_date TIMESTAMP WITH TIME ZONE,
    last_reward_claim TIMESTAMP WITH TIME ZONE,
    last_compound TIMESTAMP WITH TIME ZONE,
    risk_score DECIMAL(3,2) DEFAULT 0.00,
    performance_metrics JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Rewards table
CREATE TABLE IF NOT EXISTS yield_farming.rewards (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    position_id UUID NOT NULL REFERENCES yield_farming.positions(id),
    account_id UUID NOT NULL REFERENCES accounts.accounts(id),
    token_address VARCHAR(255) NOT NULL,
    token_symbol VARCHAR(20) NOT NULL,
    amount DECIMAL(36,18) NOT NULL,
    usd_value DECIMAL(36,18),
    reward_type VARCHAR(50) NOT NULL CHECK (reward_type IN ('staking', 'farming', 'liquidity', 'governance', 'bonus')),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'available', 'claimed', 'expired')),
    claimed_at TIMESTAMP WITH TIME ZONE,
    transaction_hash VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================================
-- TRADING MODULE
-- ============================================================================

-- Orders table
CREATE TABLE IF NOT EXISTS trading.orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts.accounts(id),
    wallet_id UUID,
    exchange_id UUID,
    symbol VARCHAR(50) NOT NULL,
    base_asset VARCHAR(20) NOT NULL,
    quote_asset VARCHAR(20) NOT NULL,
    order_type VARCHAR(20) NOT NULL CHECK (order_type IN ('market', 'limit', 'stop_loss', 'stop_limit', 'take_profit', 'trailing_stop')),
    side VARCHAR(10) NOT NULL CHECK (side IN ('buy', 'sell')),
    quantity DECIMAL(36,18) NOT NULL,
    price DECIMAL(36,18),
    stop_price DECIMAL(36,18),
    filled_quantity DECIMAL(36,18) DEFAULT 0,
    remaining_quantity DECIMAL(36,18),
    average_price DECIMAL(36,18),
    total_value DECIMAL(36,18),
    fee DECIMAL(36,18) DEFAULT 0,
    fee_asset VARCHAR(20),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'open', 'partially_filled', 'filled', 'cancelled', 'rejected', 'expired')),
    time_in_force VARCHAR(10) DEFAULT 'gtc' CHECK (time_in_force IN ('gtc', 'ioc', 'fok', 'gtd')),
    expires_at TIMESTAMP WITH TIME ZONE,
    executed_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    exchange_order_id VARCHAR(255),
    client_order_id VARCHAR(255),
    strategy_id UUID,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Trades table
CREATE TABLE IF NOT EXISTS trading.trades (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES trading.orders(id),
    account_id UUID NOT NULL REFERENCES accounts.accounts(id),
    exchange_id UUID,
    symbol VARCHAR(50) NOT NULL,
    side VARCHAR(10) NOT NULL CHECK (side IN ('buy', 'sell')),
    quantity DECIMAL(36,18) NOT NULL,
    price DECIMAL(36,18) NOT NULL,
    value DECIMAL(36,18) NOT NULL,
    fee DECIMAL(36,18) DEFAULT 0,
    fee_asset VARCHAR(20),
    is_maker BOOLEAN DEFAULT FALSE,
    exchange_trade_id VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    executed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Portfolios table
CREATE TABLE IF NOT EXISTS trading.portfolios (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts.accounts(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    base_currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    total_value DECIMAL(36,18) DEFAULT 0,
    total_pnl DECIMAL(36,18) DEFAULT 0,
    total_pnl_percent DECIMAL(8,4) DEFAULT 0,
    day_pnl DECIMAL(36,18) DEFAULT 0,
    day_pnl_percent DECIMAL(8,4) DEFAULT 0,
    holdings JSONB DEFAULT '[]',
    performance JSONB DEFAULT '{}',
    risk_metrics JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Trading strategies table
CREATE TABLE IF NOT EXISTS trading.strategies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts.accounts(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    strategy_type VARCHAR(50) NOT NULL CHECK (strategy_type IN ('manual', 'algorithmic', 'copy_trading', 'grid_trading', 'dca', 'arbitrage')),
    status VARCHAR(20) NOT NULL DEFAULT 'inactive' CHECK (status IN ('active', 'inactive', 'paused', 'stopped', 'error')),
    parameters JSONB DEFAULT '{}',
    performance JSONB DEFAULT '{}',
    risk_limits JSONB DEFAULT '{}',
    allocation DECIMAL(36,18) DEFAULT 0,
    max_allocation DECIMAL(36,18) DEFAULT 0,
    is_backtested BOOLEAN DEFAULT FALSE,
    backtest_results JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================================
-- CARDS MODULE
-- ============================================================================

-- Cards table
CREATE TABLE IF NOT EXISTS cards.cards (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts.accounts(id),
    wallet_id UUID,
    card_number VARCHAR(255) NOT NULL, -- Encrypted
    masked_number VARCHAR(20) NOT NULL,
    card_type VARCHAR(20) NOT NULL CHECK (card_type IN ('virtual', 'physical')),
    card_brand VARCHAR(20) NOT NULL CHECK (card_brand IN ('visa', 'mastercard', 'amex', 'discover', 'unionpay')),
    card_network VARCHAR(20) NOT NULL CHECK (card_network IN ('visa', 'mastercard', 'amex', 'discover')),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'inactive', 'blocked', 'suspended', 'expired', 'cancelled')),
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    balance DECIMAL(36,18) DEFAULT 0,
    available_balance DECIMAL(36,18) DEFAULT 0,
    spending_limits JSONB DEFAULT '{}',
    security_settings JSONB DEFAULT '{}',
    expiry_month INTEGER NOT NULL,
    expiry_year INTEGER NOT NULL,
    cvv VARCHAR(255), -- Encrypted
    pin VARCHAR(255), -- Encrypted
    holder_name VARCHAR(255) NOT NULL,
    billing_address JSONB NOT NULL,
    shipping_address JSONB,
    design_id VARCHAR(255),
    issued_at TIMESTAMP WITH TIME ZONE,
    activated_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    blocked_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    rewards_program JSONB,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Card transactions table
CREATE TABLE IF NOT EXISTS cards.transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    card_id UUID NOT NULL REFERENCES cards.cards(id),
    account_id UUID NOT NULL REFERENCES accounts.accounts(id),
    merchant_name VARCHAR(255),
    merchant_category VARCHAR(100),
    merchant_id VARCHAR(255),
    amount DECIMAL(36,18) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    original_amount DECIMAL(36,18),
    original_currency VARCHAR(10),
    exchange_rate DECIMAL(36,18),
    fee DECIMAL(36,18) DEFAULT 0,
    transaction_type VARCHAR(20) NOT NULL CHECK (transaction_type IN ('purchase', 'withdrawal', 'refund', 'reversal', 'adjustment', 'fee', 'reward')),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'authorized', 'settled', 'declined', 'reversed', 'refunded')),
    authorization_code VARCHAR(255),
    processor_response VARCHAR(255),
    decline_reason VARCHAR(255),
    location JSONB,
    payment_method VARCHAR(20) CHECK (payment_method IN ('chip', 'contactless', 'mag_stripe', 'online', 'atm', 'mobile')),
    is_recurring BOOLEAN DEFAULT FALSE,
    risk_score DECIMAL(3,2) DEFAULT 0.00,
    fraud_flags TEXT[],
    rewards JSONB,
    authorized_at TIMESTAMP WITH TIME ZONE NOT NULL,
    settled_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================================
-- INDEXES FOR PERFORMANCE
-- ============================================================================

-- Accounts indexes
CREATE INDEX IF NOT EXISTS idx_accounts_email ON accounts.accounts(email);
CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts.accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_accounts_status ON accounts.accounts(account_status);
CREATE INDEX IF NOT EXISTS idx_accounts_kyc_status ON accounts.accounts(kyc_status);
CREATE INDEX IF NOT EXISTS idx_accounts_created_at ON accounts.accounts(created_at);

-- Sessions indexes
CREATE INDEX IF NOT EXISTS idx_sessions_account_id ON accounts.sessions(account_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON accounts.sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON accounts.sessions(expires_at);

-- Payments indexes
CREATE INDEX IF NOT EXISTS idx_payments_account_id ON payments.payments(account_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments.payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments.payments(created_at);
CREATE INDEX IF NOT EXISTS idx_payments_transaction_hash ON payments.payments(transaction_hash);

-- Yield positions indexes
CREATE INDEX IF NOT EXISTS idx_positions_account_id ON yield_farming.positions(account_id);
CREATE INDEX IF NOT EXISTS idx_positions_protocol_id ON yield_farming.positions(protocol_id);
CREATE INDEX IF NOT EXISTS idx_positions_status ON yield_farming.positions(status);

-- Trading orders indexes
CREATE INDEX IF NOT EXISTS idx_orders_account_id ON trading.orders(account_id);
CREATE INDEX IF NOT EXISTS idx_orders_symbol ON trading.orders(symbol);
CREATE INDEX IF NOT EXISTS idx_orders_status ON trading.orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON trading.orders(created_at);

-- Cards indexes
CREATE INDEX IF NOT EXISTS idx_cards_account_id ON cards.cards(account_id);
CREATE INDEX IF NOT EXISTS idx_cards_status ON cards.cards(status);
CREATE INDEX IF NOT EXISTS idx_cards_masked_number ON cards.cards(masked_number);

-- Card transactions indexes
CREATE INDEX IF NOT EXISTS idx_card_transactions_card_id ON cards.transactions(card_id);
CREATE INDEX IF NOT EXISTS idx_card_transactions_account_id ON cards.transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_card_transactions_status ON cards.transactions(status);
CREATE INDEX IF NOT EXISTS idx_card_transactions_authorized_at ON cards.transactions(authorized_at);

-- ============================================================================
-- TRIGGERS FOR UPDATED_AT
-- ============================================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers to all tables with updated_at column
CREATE TRIGGER update_accounts_updated_at BEFORE UPDATE ON accounts.accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sessions_updated_at BEFORE UPDATE ON accounts.sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON payments.payments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payment_intents_updated_at BEFORE UPDATE ON payments.payment_intents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_refunds_updated_at BEFORE UPDATE ON payments.refunds
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_protocols_updated_at BEFORE UPDATE ON yield_farming.protocols
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_pools_updated_at BEFORE UPDATE ON yield_farming.pools
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_positions_updated_at BEFORE UPDATE ON yield_farming.positions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_rewards_updated_at BEFORE UPDATE ON yield_farming.rewards
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON trading.orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_portfolios_updated_at BEFORE UPDATE ON trading.portfolios
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_strategies_updated_at BEFORE UPDATE ON trading.strategies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cards_updated_at BEFORE UPDATE ON cards.cards
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_card_transactions_updated_at BEFORE UPDATE ON cards.transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- SAMPLE DATA FOR TESTING
-- ============================================================================

-- Insert sample account
INSERT INTO accounts.accounts (
    email, phone, first_name, last_name, country, account_type, account_status, kyc_status
) VALUES (
    'demo@fintech.com', '+1234567890', 'Demo', 'User', 'USA', 'personal', 'active', 'approved'
) ON CONFLICT (email) DO NOTHING;

-- Grant permissions
GRANT USAGE ON SCHEMA accounts TO postgres;
GRANT USAGE ON SCHEMA payments TO postgres;
GRANT USAGE ON SCHEMA yield_farming TO postgres;
GRANT USAGE ON SCHEMA trading TO postgres;
GRANT USAGE ON SCHEMA cards TO postgres;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA accounts TO postgres;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA payments TO postgres;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA yield_farming TO postgres;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA trading TO postgres;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA cards TO postgres;

GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA accounts TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA payments TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA yield_farming TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA trading TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA cards TO postgres;

COMMIT;
