-- Create contract_events table
CREATE TABLE IF NOT EXISTS contract_events (
    id UUID PRIMARY KEY,
    contract_id UUID NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL,
    event VARCHAR(100) NOT NULL,
    block_number BIGINT NOT NULL,
    log_index BIGINT NOT NULL,
    data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on contract_id
CREATE INDEX IF NOT EXISTS idx_contract_events_contract_id ON contract_events(contract_id);

-- Create index on transaction_id
CREATE INDEX IF NOT EXISTS idx_contract_events_transaction_id ON contract_events(transaction_id);

-- Create index on event
CREATE INDEX IF NOT EXISTS idx_contract_events_event ON contract_events(event);

-- Create index on block_number
CREATE INDEX IF NOT EXISTS idx_contract_events_block_number ON contract_events(block_number);
