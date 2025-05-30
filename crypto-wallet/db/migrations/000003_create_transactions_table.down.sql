-- Drop trigger
DROP TRIGGER IF EXISTS update_transactions_updated_at ON transactions;

-- Drop indexes
DROP INDEX IF EXISTS idx_transactions_chain;
DROP INDEX IF EXISTS idx_transactions_status;
DROP INDEX IF EXISTS idx_transactions_hash;
DROP INDEX IF EXISTS idx_transactions_wallet_id;
DROP INDEX IF EXISTS idx_transactions_user_id;

-- Drop table
DROP TABLE IF EXISTS transactions;
