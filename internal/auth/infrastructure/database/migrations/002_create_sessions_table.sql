-- Create sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    refresh_token_hash VARCHAR(255) NOT NULL,
    device_info JSONB,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    refresh_expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_activity TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Foreign key constraint
    CONSTRAINT fk_sessions_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token_hash ON sessions(refresh_token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_refresh_expires_at ON sessions(refresh_expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_is_active ON sessions(is_active);
CREATE INDEX IF NOT EXISTS idx_sessions_last_activity ON sessions(last_activity);
CREATE INDEX IF NOT EXISTS idx_sessions_ip_address ON sessions(ip_address);

-- Create partial indexes for active sessions
CREATE INDEX IF NOT EXISTS idx_sessions_active_user ON sessions(user_id) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_sessions_active_token ON sessions(token_hash) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_sessions_active_refresh ON sessions(refresh_token_hash) WHERE is_active = TRUE;

-- Create composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_sessions_user_active ON sessions(user_id, is_active);
CREATE INDEX IF NOT EXISTS idx_sessions_user_last_activity ON sessions(user_id, last_activity DESC);

-- Add trigger to update updated_at timestamp
CREATE TRIGGER update_sessions_updated_at 
    BEFORE UPDATE ON sessions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE sessions IS 'User authentication sessions';
COMMENT ON COLUMN sessions.id IS 'Unique session identifier (UUID)';
COMMENT ON COLUMN sessions.user_id IS 'Reference to the user who owns this session';
COMMENT ON COLUMN sessions.token_hash IS 'Hash of the access token';
COMMENT ON COLUMN sessions.refresh_token_hash IS 'Hash of the refresh token';
COMMENT ON COLUMN sessions.device_info IS 'JSON object containing device information';
COMMENT ON COLUMN sessions.ip_address IS 'IP address from which the session was created';
COMMENT ON COLUMN sessions.user_agent IS 'User agent string from the client';
COMMENT ON COLUMN sessions.expires_at IS 'When the access token expires';
COMMENT ON COLUMN sessions.refresh_expires_at IS 'When the refresh token expires';
COMMENT ON COLUMN sessions.is_active IS 'Whether the session is currently active';
COMMENT ON COLUMN sessions.last_activity IS 'Timestamp of the last activity in this session';
