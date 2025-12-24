-- Migration: Strategy Marketplace

CREATE TABLE IF NOT EXISTS strategy_market (
    id BIGSERIAL PRIMARY KEY,
    owner_id BIGINT REFERENCES users(id),
    strategy_id BIGINT REFERENCES strategies(id),
    price NUMERIC(10, 2) DEFAULT 0, -- 0 for free
    description TEXT,
    performance_metrics JSONB, -- Sharpe, Drawdown, etc.
    is_public BOOLEAN DEFAULT FALSE,
    subscriber_count INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(strategy_id)
);

CREATE TABLE IF NOT EXISTS strategy_purchases (
    user_id BIGINT REFERENCES users(id),
    market_item_id BIGINT REFERENCES strategy_market(id),
    purchased_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'active',
    PRIMARY KEY (user_id, market_item_id)
);

-- Index for discovery
CREATE INDEX IF NOT EXISTS idx_market_public ON strategy_market(is_public) WHERE is_public = TRUE;
