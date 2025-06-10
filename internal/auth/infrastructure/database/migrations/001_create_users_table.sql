-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone_number VARCHAR(20),
    role VARCHAR(20) NOT NULL DEFAULT 'customer',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_phone_verified BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- MFA fields
    mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    mfa_method VARCHAR(20) DEFAULT 'none',
    mfa_secret VARCHAR(255),
    mfa_backup_codes JSONB,
    
    -- Security fields
    failed_login_attempts INTEGER NOT NULL DEFAULT 0,
    last_failed_login TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_password_change TIMESTAMP WITH TIME ZONE,
    security_level VARCHAR(20) NOT NULL DEFAULT 'low',
    risk_score DECIMAL(3,2) NOT NULL DEFAULT 0.0,
    
    -- Device tracking
    device_fingerprints JSONB,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_mfa_enabled ON users(mfa_enabled);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_users_last_login ON users(last_login_at);

-- Create partial index for active users
CREATE INDEX IF NOT EXISTS idx_users_active ON users(id) WHERE status = 'active';

-- Add constraints
ALTER TABLE users ADD CONSTRAINT chk_users_role 
    CHECK (role IN ('admin', 'customer', 'staff', 'manager'));

ALTER TABLE users ADD CONSTRAINT chk_users_status 
    CHECK (status IN ('active', 'inactive', 'suspended', 'deleted'));

ALTER TABLE users ADD CONSTRAINT chk_users_security_level 
    CHECK (security_level IN ('low', 'medium', 'high'));

ALTER TABLE users ADD CONSTRAINT chk_users_mfa_method 
    CHECK (mfa_method IN ('none', 'totp', 'sms', 'email'));

ALTER TABLE users ADD CONSTRAINT chk_users_risk_score 
    CHECK (risk_score >= 0.0 AND risk_score <= 1.0);

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE users IS 'User accounts and authentication information';
COMMENT ON COLUMN users.id IS 'Unique user identifier (UUID)';
COMMENT ON COLUMN users.email IS 'User email address (unique)';
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hashed password';
COMMENT ON COLUMN users.role IS 'User role (admin, customer, staff, manager)';
COMMENT ON COLUMN users.status IS 'Account status (active, inactive, suspended, deleted)';
COMMENT ON COLUMN users.mfa_enabled IS 'Whether multi-factor authentication is enabled';
COMMENT ON COLUMN users.mfa_method IS 'MFA method (none, totp, sms, email)';
COMMENT ON COLUMN users.mfa_secret IS 'MFA secret key for TOTP';
COMMENT ON COLUMN users.mfa_backup_codes IS 'JSON array of backup codes for MFA';
COMMENT ON COLUMN users.failed_login_attempts IS 'Number of consecutive failed login attempts';
COMMENT ON COLUMN users.security_level IS 'Security level based on risk assessment';
COMMENT ON COLUMN users.risk_score IS 'Risk score from 0.0 to 1.0';
COMMENT ON COLUMN users.device_fingerprints IS 'JSON array of trusted device fingerprints';
