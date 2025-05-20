-- Create contracts table
CREATE TABLE IF NOT EXISTS contracts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    address VARCHAR(255) NOT NULL,
    chain VARCHAR(20) NOT NULL,
    abi TEXT NOT NULL,
    bytecode TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on user_id
CREATE INDEX IF NOT EXISTS idx_contracts_user_id ON contracts(user_id);

-- Create index on address
CREATE INDEX IF NOT EXISTS idx_contracts_address ON contracts(address);

-- Create index on chain
CREATE INDEX IF NOT EXISTS idx_contracts_chain ON contracts(chain);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER update_contracts_updated_at
BEFORE UPDATE ON contracts
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
