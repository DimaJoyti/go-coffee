-- Drop trigger
DROP TRIGGER IF EXISTS update_wallets_updated_at ON wallets;

-- Drop indexes
DROP INDEX IF EXISTS idx_wallets_chain;
DROP INDEX IF EXISTS idx_wallets_address;
DROP INDEX IF EXISTS idx_wallets_user_id;

-- Drop table
DROP TABLE IF EXISTS wallets;
