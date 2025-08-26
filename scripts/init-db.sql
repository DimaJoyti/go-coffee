-- Initialize Go Coffee Database
-- This script sets up the basic database structure for the Go Coffee platform

-- Create database if not exists (handled by Docker environment)
-- CREATE DATABASE IF NOT EXISTS go_coffee;

-- Use the database
-- \c go_coffee;

-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_name VARCHAR(255) NOT NULL,
    coffee_type VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP WITH TIME ZONE,
    total_amount DECIMAL(10,2),
    payment_method VARCHAR(50),
    notes TEXT
);

-- Create index on status for faster queries
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);

-- Create coffee_types table for menu management
CREATE TABLE IF NOT EXISTS coffee_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    available BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert default coffee types
INSERT INTO coffee_types (name, description, price) VALUES
    ('Espresso', 'Strong black coffee', 2.50),
    ('Americano', 'Espresso with hot water', 3.00),
    ('Latte', 'Espresso with steamed milk', 4.50),
    ('Cappuccino', 'Espresso with steamed milk and foam', 4.00),
    ('Macchiato', 'Espresso with a dollop of foam', 3.50),
    ('Mocha', 'Espresso with chocolate and steamed milk', 5.00),
    ('Flat White', 'Espresso with microfoam milk', 4.25),
    ('Cold Brew', 'Cold extracted coffee', 3.75)
ON CONFLICT (name) DO NOTHING;

-- Create customers table
CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(50),
    loyalty_points INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create order_items table for detailed order tracking
CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
    coffee_type_id INTEGER REFERENCES coffee_types(id),
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(10,2) NOT NULL,
    customizations JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create inventory table
CREATE TABLE IF NOT EXISTS inventory (
    id SERIAL PRIMARY KEY,
    item_name VARCHAR(255) NOT NULL UNIQUE,
    current_stock INTEGER NOT NULL DEFAULT 0,
    min_stock INTEGER NOT NULL DEFAULT 10,
    max_stock INTEGER NOT NULL DEFAULT 100,
    unit VARCHAR(50) NOT NULL DEFAULT 'units',
    cost_per_unit DECIMAL(10,2),
    supplier VARCHAR(255),
    last_restocked TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert basic inventory items
INSERT INTO inventory (item_name, current_stock, min_stock, max_stock, unit, cost_per_unit) VALUES
    ('Coffee Beans - Arabica', 50, 10, 100, 'kg', 15.00),
    ('Coffee Beans - Robusta', 30, 10, 100, 'kg', 12.00),
    ('Milk', 20, 5, 50, 'liters', 1.50),
    ('Sugar', 25, 5, 50, 'kg', 2.00),
    ('Paper Cups - Small', 500, 100, 1000, 'pieces', 0.05),
    ('Paper Cups - Medium', 400, 100, 1000, 'pieces', 0.07),
    ('Paper Cups - Large', 300, 100, 1000, 'pieces', 0.10),
    ('Coffee Lids', 800, 200, 2000, 'pieces', 0.03)
ON CONFLICT (item_name) DO NOTHING;

-- Create staff table
CREATE TABLE IF NOT EXISTS staff (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    role VARCHAR(100) NOT NULL,
    shift_start TIME,
    shift_end TIME,
    hourly_rate DECIMAL(10,2),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create audit log table for tracking changes
CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name VARCHAR(100) NOT NULL,
    record_id VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL, -- INSERT, UPDATE, DELETE
    old_values JSONB,
    new_values JSONB,
    changed_by VARCHAR(255),
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_customers_updated_at BEFORE UPDATE ON customers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_inventory_updated_at BEFORE UPDATE ON inventory
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create views for common queries
CREATE OR REPLACE VIEW order_summary AS
SELECT 
    o.id,
    o.customer_name,
    o.status,
    o.created_at,
    COUNT(oi.id) as item_count,
    SUM(oi.quantity * oi.unit_price) as total_amount
FROM orders o
LEFT JOIN order_items oi ON o.id = oi.order_id
GROUP BY o.id, o.customer_name, o.status, o.created_at;

-- Create view for low stock items
CREATE OR REPLACE VIEW low_stock_items AS
SELECT 
    item_name,
    current_stock,
    min_stock,
    (min_stock - current_stock) as shortage
FROM inventory
WHERE current_stock <= min_stock;

-- Grant permissions (adjust as needed for your security requirements)
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO go_coffee_user;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO go_coffee_user;

-- Insert some sample data for testing
INSERT INTO customers (name, email, phone, loyalty_points) VALUES
    ('John Doe', 'john.doe@example.com', '+1234567890', 150),
    ('Jane Smith', 'jane.smith@example.com', '+1234567891', 75),
    ('Bob Johnson', 'bob.johnson@example.com', '+1234567892', 200)
ON CONFLICT (email) DO NOTHING;

-- Insert sample staff
INSERT INTO staff (name, email, role, shift_start, shift_end, hourly_rate) VALUES
    ('Alice Manager', 'alice@gocoffee.com', 'Manager', '06:00', '14:00', 25.00),
    ('Bob Barista', 'bob@gocoffee.com', 'Barista', '07:00', '15:00', 18.00),
    ('Carol Cashier', 'carol@gocoffee.com', 'Cashier', '08:00', '16:00', 16.00)
ON CONFLICT (email) DO NOTHING;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_table_record ON audit_log(table_name, record_id);
CREATE INDEX IF NOT EXISTS idx_inventory_stock_level ON inventory(current_stock, min_stock);

-- Success message
SELECT 'Go Coffee database initialized successfully!' as message;
