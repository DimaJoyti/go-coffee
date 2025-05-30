-- Rollback initial database schema migration

-- Drop triggers
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
DROP TRIGGER IF EXISTS update_vendors_updated_at ON vendors;
DROP TRIGGER IF EXISTS update_accounts_updated_at ON accounts;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_order_items_product_id;
DROP INDEX IF EXISTS idx_order_items_order_id;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_account_id;
DROP INDEX IF EXISTS idx_products_name;
DROP INDEX IF EXISTS idx_products_vendor_id;
DROP INDEX IF EXISTS idx_vendors_name;
DROP INDEX IF EXISTS idx_accounts_email;
DROP INDEX IF EXISTS idx_accounts_username;

-- Drop tables
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS vendors;
DROP TABLE IF EXISTS accounts;
