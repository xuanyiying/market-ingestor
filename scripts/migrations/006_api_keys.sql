-- Migration: API Key Management

CREATE TABLE IF NOT EXISTS api_keys (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    key_id VARCHAR(50) UNIQUE NOT NULL, -- Public identifier
    key_secret VARCHAR(100) NOT NULL, -- Hashed secret
    name VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for lookup
CREATE INDEX IF NOT EXISTS idx_api_keys_user ON api_keys(user_id);
