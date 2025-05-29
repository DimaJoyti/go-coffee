-- Fintech Platform Sample Data
-- This script inserts sample data for testing and development

-- Set search path
SET search_path TO accounts, payments, yield_farming, trading, cards, public;

-- ============================================================================
-- SAMPLE ACCOUNTS
-- ============================================================================

-- Insert sample accounts
INSERT INTO accounts.accounts (
    id, user_id, email, phone, first_name, last_name, date_of_birth,
    nationality, country, state, city, address, postal_code,
    account_type, account_status, kyc_status, kyc_level, risk_score,
    compliance_flags, two_factor_enabled, two_factor_method,
    account_limits, notification_preferences, metadata
) VALUES 
(
    '550e8400-e29b-41d4-a716-446655440001',
    '550e8400-e29b-41d4-a716-446655440101',
    'demo@fintech.com',
    '+1234567890',
    'Demo',
    'User',
    '1990-01-01',
    'USA',
    'USA',
    'California',
    'San Francisco',
    '123 Main Street',
    '94102',
    'personal',
    'active',
    'approved',
    'enhanced',
    0.1,
    '{}',
    true,
    'totp',
    '{"daily_transaction_limit": "10000.00", "monthly_transaction_limit": "100000.00", "single_transaction_limit": "5000.00", "max_wallets": 10, "max_cards": 5, "withdrawal_limit": "5000.00", "deposit_limit": "50000.00"}',
    '{"email_enabled": true, "sms_enabled": true, "push_enabled": true, "security_alerts": true, "transaction_alerts": true, "marketing_emails": false, "product_updates": true, "weekly_reports": true, "monthly_statements": true}',
    '{"password_hash": "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9qK", "created_by": "system", "test_account": true}'
),
(
    '550e8400-e29b-41d4-a716-446655440002',
    '550e8400-e29b-41d4-a716-446655440102',
    'business@fintech.com',
    '+1234567891',
    'Business',
    'Account',
    '1985-05-15',
    'USA',
    'USA',
    'New York',
    'New York',
    '456 Business Ave',
    '10001',
    'business',
    'active',
    'approved',
    'enhanced',
    0.2,
    '{}',
    true,
    'sms',
    '{"daily_transaction_limit": "50000.00", "monthly_transaction_limit": "500000.00", "single_transaction_limit": "25000.00", "max_wallets": 20, "max_cards": 10, "withdrawal_limit": "25000.00", "deposit_limit": "250000.00"}',
    '{"email_enabled": true, "sms_enabled": true, "push_enabled": true, "security_alerts": true, "transaction_alerts": true, "marketing_emails": true, "product_updates": true, "weekly_reports": true, "monthly_statements": true}',
    '{"password_hash": "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9qK", "created_by": "system", "test_account": true, "business_type": "llc"}'
),
(
    '550e8400-e29b-41d4-a716-446655440003',
    '550e8400-e29b-41d4-a716-446655440103',
    'premium@fintech.com',
    '+1234567892',
    'Premium',
    'Customer',
    '1980-12-25',
    'GBR',
    'GBR',
    'England',
    'London',
    '789 Premium Street',
    'SW1A 1AA',
    'personal',
    'active',
    'approved',
    'enhanced',
    0.05,
    '{}',
    true,
    'totp',
    '{"daily_transaction_limit": "25000.00", "monthly_transaction_limit": "250000.00", "single_transaction_limit": "10000.00", "max_wallets": 15, "max_cards": 8, "withdrawal_limit": "10000.00", "deposit_limit": "100000.00"}',
    '{"email_enabled": true, "sms_enabled": true, "push_enabled": true, "security_alerts": true, "transaction_alerts": true, "marketing_emails": false, "product_updates": true, "weekly_reports": true, "monthly_statements": true}',
    '{"password_hash": "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9qK", "created_by": "system", "test_account": true, "tier": "premium"}'
) ON CONFLICT (email) DO NOTHING;

-- ============================================================================
-- SAMPLE PROTOCOLS (YIELD FARMING)
-- ============================================================================

-- Insert sample DeFi protocols
INSERT INTO yield_farming.protocols (
    id, name, description, website, logo_url, category, network,
    contract_address, tvl, volume_24h, average_apy, risk_score,
    is_active, is_audited, audit_reports, supported_tokens, metadata
) VALUES 
(
    '660e8400-e29b-41d4-a716-446655440001',
    'Uniswap V3',
    'Decentralized exchange with concentrated liquidity',
    'https://uniswap.org',
    'https://uniswap.org/logo.png',
    'dex',
    'ethereum',
    '0x1F98431c8aD98523631AE4a59f267346ea31F984',
    '5000000000.00',
    '1000000000.00',
    8.5,
    0.3,
    true,
    true,
    '{"consensys": "https://consensys.net/diligence/audits/2021/03/uniswap-v3/"}',
    '["ETH", "USDC", "USDT", "DAI", "WBTC"]',
    '{"version": "v3", "fee_tiers": [0.05, 0.3, 1.0], "concentrated_liquidity": true}'
),
(
    '660e8400-e29b-41d4-a716-446655440002',
    'Compound V3',
    'Algorithmic money market protocol',
    'https://compound.finance',
    'https://compound.finance/logo.png',
    'lending',
    'ethereum',
    '0xc3d688B66703497DAA19211EEdff47f25384cdc3',
    '3000000000.00',
    '50000000.00',
    4.2,
    0.2,
    true,
    true,
    '{"openzeppelin": "https://blog.openzeppelin.com/compound-audit/"}',
    '["ETH", "USDC", "DAI", "WBTC"]',
    '{"version": "v3", "collateral_factor": 0.8, "liquidation_threshold": 0.85}'
),
(
    '660e8400-e29b-41d4-a716-446655440003',
    'Aave V3',
    'Open source and non-custodial liquidity protocol',
    'https://aave.com',
    'https://aave.com/logo.png',
    'lending',
    'ethereum',
    '0x87870Bca3F3fD6335C3F4ce8392D69350B4fA4E2',
    '8000000000.00',
    '200000000.00',
    5.8,
    0.25,
    true,
    true,
    '{"consensys": "https://consensys.net/diligence/audits/2022/01/aave-v3/"}',
    '["ETH", "USDC", "USDT", "DAI", "WBTC", "LINK"]',
    '{"version": "v3", "efficiency_mode": true, "isolation_mode": true, "portal": true}'
) ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- SAMPLE POOLS
-- ============================================================================

-- Insert sample yield farming pools
INSERT INTO yield_farming.pools (
    id, protocol_id, name, description, pool_type, token_pair, reward_tokens,
    apy, apr, tvl, volume_24h, min_deposit, max_deposit, lock_period,
    withdrawal_fee, performance_fee, risk_score, is_active, metadata
) VALUES 
(
    '770e8400-e29b-41d4-a716-446655440001',
    '660e8400-e29b-41d4-a716-446655440001',
    'ETH/USDC 0.3%',
    'Ethereum and USDC liquidity pool with 0.3% fee tier',
    'liquidity_mining',
    '["ETH", "USDC"]',
    '["UNI"]',
    12.5,
    11.8,
    '500000000.00',
    '50000000.00',
    '100.00',
    '10000000.00',
    NULL,
    0.0,
    0.0,
    0.3,
    true,
    '{"fee_tier": 0.3, "tick_spacing": 60, "price_range": "active"}'
),
(
    '770e8400-e29b-41d4-a716-446655440002',
    '660e8400-e29b-41d4-a716-446655440002',
    'USDC Supply',
    'Supply USDC to earn interest',
    'lending',
    '["USDC"]',
    '["COMP"]',
    4.2,
    4.0,
    '1000000000.00',
    '10000000.00',
    '10.00',
    '50000000.00',
    NULL,
    0.0,
    0.1,
    0.2,
    true,
    '{"utilization_rate": 0.75, "supply_apy": 4.2, "borrow_apy": 6.8}'
),
(
    '770e8400-e29b-41d4-a716-446655440003',
    '660e8400-e29b-41d4-a716-446655440003',
    'ETH Lending',
    'Supply ETH to earn interest and AAVE rewards',
    'lending',
    '["ETH"]',
    '["AAVE"]',
    5.8,
    5.5,
    '2000000000.00',
    '100000000.00',
    '0.1',
    '10000.00',
    NULL,
    0.0,
    0.0,
    0.25,
    true,
    '{"ltv": 0.8, "liquidation_threshold": 0.825, "liquidation_bonus": 0.05}'
) ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- SAMPLE YIELD POSITIONS
-- ============================================================================

-- Insert sample yield positions
INSERT INTO yield_farming.positions (
    id, account_id, protocol_id, pool_id, position_type, strategy,
    token_address, token_symbol, amount, usd_value, entry_price, current_price,
    apy, apr, daily_rewards, total_rewards, claimed_rewards, pending_rewards,
    status, auto_compound, risk_score, performance_metrics, metadata
) VALUES 
(
    '880e8400-e29b-41d4-a716-446655440001',
    '550e8400-e29b-41d4-a716-446655440001',
    '660e8400-e29b-41d4-a716-446655440001',
    '770e8400-e29b-41d4-a716-446655440001',
    'liquidity_mining',
    'conservative',
    '0xA0b86a33E6441E6C8D3C8C8C8C8C8C8C8C8C8C8C',
    'ETH',
    '5.0',
    '10000.00',
    '2000.00',
    '2100.00',
    12.5,
    11.8,
    '3.42',
    '125.50',
    '100.00',
    '25.50',
    'active',
    true,
    0.3,
    '{"total_return": "625.50", "total_return_usd": "625.50", "roi": 0.0625, "daily_roi": 0.000342, "days_active": 30}',
    '{"pool_share": 0.001, "impermanent_loss": "15.25", "fees_earned": "45.75"}'
),
(
    '880e8400-e29b-41d4-a716-446655440002',
    '550e8400-e29b-41d4-a716-446655440002',
    '660e8400-e29b-41d4-a716-446655440002',
    '770e8400-e29b-41d4-a716-446655440002',
    'lending',
    'moderate',
    '0xA0b86a33E6441E6C8D3C8C8C8C8C8C8C8C8C8C8D',
    'USDC',
    '50000.0',
    '50000.00',
    '1.00',
    '1.00',
    4.2,
    4.0,
    '5.75',
    '210.00',
    '180.00',
    '30.00',
    'active',
    true,
    0.2,
    '{"total_return": "210.00", "total_return_usd": "210.00", "roi": 0.0042, "daily_roi": 0.000115, "days_active": 60}',
    '{"utilization_rate": 0.75, "health_factor": 2.5}'
) ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- SAMPLE PAYMENTS
-- ============================================================================

-- Insert sample payments
INSERT INTO payments.payments (
    id, account_id, payment_method, currency, amount, fee_amount, net_amount,
    status, payment_type, description, reference, network, from_address, to_address,
    risk_score, processed_at, metadata
) VALUES 
(
    '990e8400-e29b-41d4-a716-446655440001',
    '550e8400-e29b-41d4-a716-446655440001',
    'crypto',
    'ETH',
    '1.0',
    '0.005',
    '0.995',
    'completed',
    'inbound',
    'Deposit from external wallet',
    'DEP-001',
    'ethereum',
    '0x742d35Cc6634C0532925a3b8D4C8C8C8C8C8C8C8',
    '0x8ba1f109551bD432803012645Hac136c8C8C8C8C',
    0.1,
    NOW() - INTERVAL '1 day',
    '{"transaction_hash": "0x1234567890abcdef", "block_number": 18500000, "confirmations": 12}'
),
(
    '990e8400-e29b-41d4-a716-446655440002',
    '550e8400-e29b-41d4-a716-446655440001',
    'crypto',
    'USDC',
    '5000.0',
    '2.50',
    '4997.50',
    'completed',
    'outbound',
    'Payment to merchant',
    'PAY-002',
    'ethereum',
    '0x8ba1f109551bD432803012645Hac136c8C8C8C8C',
    '0x742d35Cc6634C0532925a3b8D4C8C8C8C8C8C8C8',
    0.2,
    NOW() - INTERVAL '2 hours',
    '{"transaction_hash": "0xabcdef1234567890", "block_number": 18500100, "confirmations": 8}'
) ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- SAMPLE TRADING ORDERS
-- ============================================================================

-- Insert sample trading orders
INSERT INTO trading.orders (
    id, account_id, symbol, base_asset, quote_asset, order_type, side,
    quantity, price, filled_quantity, remaining_quantity, average_price,
    total_value, fee, fee_asset, status, time_in_force, metadata
) VALUES 
(
    'aa0e8400-e29b-41d4-a716-446655440001',
    '550e8400-e29b-41d4-a716-446655440001',
    'ETH/USDC',
    'ETH',
    'USDC',
    'limit',
    'buy',
    '2.0',
    '2000.00',
    '2.0',
    '0.0',
    '2000.00',
    '4000.00',
    '4.00',
    'USDC',
    'filled',
    'gtc',
    '{"exchange": "uniswap", "slippage": 0.5, "gas_fee": "15.50"}'
),
(
    'aa0e8400-e29b-41d4-a716-446655440002',
    '550e8400-e29b-41d4-a716-446655440002',
    'BTC/USDT',
    'BTC',
    'USDT',
    'market',
    'sell',
    '0.1',
    '45000.00',
    '0.1',
    '0.0',
    '45000.00',
    '4500.00',
    '4.50',
    'USDT',
    'filled',
    'ioc',
    '{"exchange": "binance", "execution_time": "2.5s"}'
) ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- SAMPLE CARDS
-- ============================================================================

-- Insert sample cards
INSERT INTO cards.cards (
    id, account_id, card_number, masked_number, card_type, card_brand, card_network,
    status, currency, balance, available_balance, expiry_month, expiry_year,
    cvv, holder_name, billing_address, spending_limits, security_settings,
    issued_at, activated_at, metadata
) VALUES 
(
    'bb0e8400-e29b-41d4-a716-446655440001',
    '550e8400-e29b-41d4-a716-446655440001',
    'encrypted_card_number_1',
    '**** **** **** 1234',
    'virtual',
    'visa',
    'visa',
    'active',
    'USD',
    '5000.00',
    '4500.00',
    12,
    2027,
    'encrypted_cvv_1',
    'Demo User',
    '{"first_name": "Demo", "last_name": "User", "address_line1": "123 Main Street", "city": "San Francisco", "state": "CA", "postal_code": "94102", "country": "US"}',
    '{"daily_limit": "1000.00", "monthly_limit": "10000.00", "transaction_limit": "500.00", "atm_limit": "500.00", "online_limit": "2000.00"}',
    '{"pin_required": true, "cvv_required": true, "fraud_detection": true, "velocity_checks": true, "notifications_enabled": true}',
    NOW() - INTERVAL '30 days',
    NOW() - INTERVAL '29 days',
    '{"card_design": "default", "delivery_method": "instant", "issuer": "marqeta"}'
),
(
    'bb0e8400-e29b-41d4-a716-446655440002',
    '550e8400-e29b-41d4-a716-446655440002',
    'encrypted_card_number_2',
    '**** **** **** 5678',
    'physical',
    'mastercard',
    'mastercard',
    'active',
    'USD',
    '25000.00',
    '24000.00',
    6,
    2028,
    'encrypted_cvv_2',
    'Business Account',
    '{"first_name": "Business", "last_name": "Account", "company": "Business Corp", "address_line1": "456 Business Ave", "city": "New York", "state": "NY", "postal_code": "10001", "country": "US"}',
    '{"daily_limit": "5000.00", "monthly_limit": "50000.00", "transaction_limit": "2500.00", "atm_limit": "2000.00", "online_limit": "10000.00"}',
    '{"pin_required": true, "cvv_required": true, "biometric_required": true, "fraud_detection": true, "velocity_checks": true, "geofencing_enabled": true}',
    NOW() - INTERVAL '60 days',
    NOW() - INTERVAL '55 days',
    '{"card_design": "business", "delivery_method": "express", "issuer": "galileo", "physical_shipped": true}'
) ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- SAMPLE CARD TRANSACTIONS
-- ============================================================================

-- Insert sample card transactions
INSERT INTO cards.transactions (
    id, card_id, account_id, merchant_name, merchant_category, amount, currency,
    transaction_type, status, payment_method, location, authorized_at, settled_at,
    risk_score, metadata
) VALUES 
(
    'cc0e8400-e29b-41d4-a716-446655440001',
    'bb0e8400-e29b-41d4-a716-446655440001',
    '550e8400-e29b-41d4-a716-446655440001',
    'Coffee Shop',
    'dining',
    '15.50',
    'USD',
    'purchase',
    'settled',
    'contactless',
    '{"country": "US", "city": "San Francisco", "latitude": 37.7749, "longitude": -122.4194}',
    NOW() - INTERVAL '2 hours',
    NOW() - INTERVAL '1 hour',
    0.1,
    '{"merchant_id": "COFFEE123", "terminal_id": "T001", "authorization_code": "123456"}'
),
(
    'cc0e8400-e29b-41d4-a716-446655440002',
    'bb0e8400-e29b-41d4-a716-446655440001',
    '550e8400-e29b-41d4-a716-446655440001',
    'Amazon',
    'online',
    '89.99',
    'USD',
    'purchase',
    'settled',
    'online',
    '{"country": "US", "city": "Seattle", "latitude": 47.6062, "longitude": -122.3321}',
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '23 hours',
    0.05,
    '{"merchant_id": "AMAZON", "order_id": "AMZ123456789", "authorization_code": "789012"}'
) ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- UPDATE SEQUENCES AND CONSTRAINTS
-- ============================================================================

-- Refresh materialized views if any exist
-- REFRESH MATERIALIZED VIEW IF EXISTS account_summary;

-- Update statistics
ANALYZE accounts.accounts;
ANALYZE payments.payments;
ANALYZE yield_farming.protocols;
ANALYZE yield_farming.pools;
ANALYZE yield_farming.positions;
ANALYZE trading.orders;
ANALYZE cards.cards;
ANALYZE cards.transactions;

-- Success message
DO $$
BEGIN
    RAISE NOTICE 'Sample data inserted successfully!';
    RAISE NOTICE 'Created:';
    RAISE NOTICE '- 3 sample accounts (demo@fintech.com, business@fintech.com, premium@fintech.com)';
    RAISE NOTICE '- 3 DeFi protocols (Uniswap, Compound, Aave)';
    RAISE NOTICE '- 3 yield farming pools';
    RAISE NOTICE '- 2 yield positions';
    RAISE NOTICE '- 2 payment transactions';
    RAISE NOTICE '- 2 trading orders';
    RAISE NOTICE '- 2 cards (1 virtual, 1 physical)';
    RAISE NOTICE '- 2 card transactions';
    RAISE NOTICE '';
    RAISE NOTICE 'Test credentials:';
    RAISE NOTICE 'Email: demo@fintech.com';
    RAISE NOTICE 'Password: password123';
END $$;
