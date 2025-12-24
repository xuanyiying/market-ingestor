-- Migration: Add alert system

CREATE TABLE IF NOT EXISTS alerts (
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    symbol VARCHAR(20) NOT NULL,
    condition_type VARCHAR(50) NOT NULL, -- price_above, price_below
    target_value NUMERIC(20, 8),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for performance
CREATE INDEX idx_alerts_user_id ON alerts(user_id);
CREATE INDEX idx_alerts_symbol_active ON alerts(symbol) WHERE is_active = TRUE;
