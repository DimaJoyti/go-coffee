-- Order Service Database Initialization Script

-- Create database if not exists (this is handled by Docker environment variables)
-- CREATE DATABASE IF NOT EXISTS order_db;

-- Use the database
-- \c order_db;

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create enum types
CREATE TYPE order_status AS ENUM (
    'pending',
    'confirmed',
    'preparing',
    'ready',
    'completed',
    'cancelled'
);

CREATE TYPE payment_status AS ENUM (
    'pending',
    'processing',
    'completed',
    'failed',
    'cancelled',
    'refunded'
);

CREATE TYPE payment_method AS ENUM (
    'credit_card',
    'debit_card',
    'crypto',
    'loyalty_token',
    'cash'
);

CREATE TYPE crypto_network AS ENUM (
    'bitcoin',
    'ethereum',
    'solana',
    'polygon'
);

-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id VARCHAR(255) NOT NULL,
    status order_status NOT NULL DEFAULT 'pending',
    total_amount BIGINT NOT NULL, -- Amount in cents
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    estimated_time INTEGER, -- Estimated preparation time in seconds
    special_instructions TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    confirmed_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    cancellation_reason TEXT
);

-- Create order_items table
CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price BIGINT NOT NULL, -- Price in cents
    total_price BIGINT NOT NULL, -- Total price in cents
    customizations JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create payments table
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    customer_id VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL, -- Amount in cents
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    payment_method payment_method NOT NULL,
    status payment_status NOT NULL DEFAULT 'pending',
    processor_id VARCHAR(255),
    processor_ref VARCHAR(255),
    transaction_hash VARCHAR(255), -- For crypto payments
    crypto_network crypto_network, -- For crypto payments
    crypto_address VARCHAR(255), -- For crypto payments
    failure_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    refunded_at TIMESTAMP WITH TIME ZONE
);

-- Create domain_events table for event sourcing
CREATE TABLE IF NOT EXISTS domain_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_id UUID NOT NULL,
    aggregate_type VARCHAR(255) NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    event_data JSONB NOT NULL,
    event_version INTEGER NOT NULL DEFAULT 1,
    occurred_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);
CREATE INDEX IF NOT EXISTS idx_orders_updated_at ON orders(updated_at);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);

CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_customer_id ON payments(customer_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_payment_method ON payments(payment_method);
CREATE INDEX IF NOT EXISTS idx_payments_processor_ref ON payments(processor_ref);
CREATE INDEX IF NOT EXISTS idx_payments_transaction_hash ON payments(transaction_hash);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at);

CREATE INDEX IF NOT EXISTS idx_domain_events_aggregate_id ON domain_events(aggregate_id);
CREATE INDEX IF NOT EXISTS idx_domain_events_aggregate_type ON domain_events(aggregate_type);
CREATE INDEX IF NOT EXISTS idx_domain_events_event_type ON domain_events(event_type);
CREATE INDEX IF NOT EXISTS idx_domain_events_occurred_at ON domain_events(occurred_at);
CREATE INDEX IF NOT EXISTS idx_domain_events_processed_at ON domain_events(processed_at);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_orders_updated_at 
    BEFORE UPDATE ON orders 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payments_updated_at 
    BEFORE UPDATE ON payments 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Insert sample data for testing
INSERT INTO orders (id, customer_id, status, total_amount, currency, estimated_time, special_instructions) VALUES
    ('550e8400-e29b-41d4-a716-446655440001', 'customer-001', 'pending', 1500, 'USD', 300, 'Extra hot, no sugar'),
    ('550e8400-e29b-41d4-a716-446655440002', 'customer-002', 'confirmed', 2000, 'USD', 450, 'Oat milk please'),
    ('550e8400-e29b-41d4-a716-446655440003', 'customer-003', 'completed', 1200, 'USD', 240, NULL)
ON CONFLICT (id) DO NOTHING;

INSERT INTO order_items (order_id, product_id, name, quantity, unit_price, total_price) VALUES
    ('550e8400-e29b-41d4-a716-446655440001', 'coffee-001', 'Espresso', 1, 500, 500),
    ('550e8400-e29b-41d4-a716-446655440001', 'pastry-001', 'Croissant', 2, 500, 1000),
    ('550e8400-e29b-41d4-a716-446655440002', 'coffee-002', 'Latte', 1, 600, 600),
    ('550e8400-e29b-41d4-a716-446655440002', 'coffee-003', 'Cappuccino', 1, 550, 550),
    ('550e8400-e29b-41d4-a716-446655440002', 'pastry-002', 'Muffin', 1, 850, 850),
    ('550e8400-e29b-41d4-a716-446655440003', 'coffee-001', 'Espresso', 2, 500, 1000),
    ('550e8400-e29b-41d4-a716-446655440003', 'addon-001', 'Extra Shot', 1, 200, 200);

INSERT INTO payments (order_id, customer_id, amount, currency, payment_method, status, processor_id) VALUES
    ('550e8400-e29b-41d4-a716-446655440001', 'customer-001', 1500, 'USD', 'credit_card', 'pending', 'stripe'),
    ('550e8400-e29b-41d4-a716-446655440002', 'customer-002', 2000, 'USD', 'crypto', 'processing', 'bitcoin'),
    ('550e8400-e29b-41d4-a716-446655440003', 'customer-003', 1200, 'USD', 'loyalty_token', 'completed', 'loyalty')
ON CONFLICT DO NOTHING;

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO order_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO order_user;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO order_user;
