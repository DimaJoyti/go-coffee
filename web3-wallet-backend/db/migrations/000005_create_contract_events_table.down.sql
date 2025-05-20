-- Drop indexes
DROP INDEX IF EXISTS idx_contract_events_block_number;
DROP INDEX IF EXISTS idx_contract_events_event;
DROP INDEX IF EXISTS idx_contract_events_transaction_id;
DROP INDEX IF EXISTS idx_contract_events_contract_id;

-- Drop table
DROP TABLE IF EXISTS contract_events;
