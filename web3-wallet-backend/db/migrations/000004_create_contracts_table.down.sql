-- Drop trigger
DROP TRIGGER IF EXISTS update_contracts_updated_at ON contracts;

-- Drop indexes
DROP INDEX IF EXISTS idx_contracts_chain;
DROP INDEX IF EXISTS idx_contracts_address;
DROP INDEX IF EXISTS idx_contracts_user_id;

-- Drop table
DROP TABLE IF EXISTS contracts;
