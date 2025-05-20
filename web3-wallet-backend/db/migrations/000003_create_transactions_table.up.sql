-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    hash VARCHAR(255),
    from_address VARCHAR(255) NOT NULL,
    to_address VARCHAR(255) NOT NULL,
    value VARCHAR(100) NOT NULL,
    gas BIGINT NOT NULL,
    gas_price VARCHAR(100) NOT NULL,
    nonce BIGINT NOT NULL,
    data TEXT,
    chain VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    block_number BIGINT,
    block_hash VARCHAR(255),
    confirmations BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on user_id
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);

-- Create index on wallet_id
CREATE INDEX IF NOT EXISTS idx_transactions_wallet_id ON transactions(wallet_id);

-- Create index on hash
CREATE INDEX IF NOT EXISTS idx_transactions_hash ON transactions(hash);

-- Create index on status
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);

-- Create index on chain
CREATE INDEX IF NOT EXISTS idx_transactions_chain ON transactions(chain);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER update_transactions_updated_at
BEFORE UPDATE ON transactions
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
