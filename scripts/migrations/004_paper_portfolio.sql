-- Migration: Paper Trading & Portfolio Management

-- 1. Paper Trading Tables
CREATE TABLE IF NOT EXISTS paper_accounts (
    user_id BIGINT PRIMARY KEY REFERENCES users(id),
    balance NUMERIC(20, 8) DEFAULT 100000.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS paper_orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL, -- buy/sell
    type VARCHAR(20) NOT NULL, -- market/limit
    price NUMERIC(20, 8),
    qty NUMERIC(20, 8) NOT NULL,
    status VARCHAR(20) DEFAULT 'open', -- open, filled, cancelled
    filled_price NUMERIC(20, 8),
    filled_time TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS paper_positions (
    user_id BIGINT REFERENCES users(id),
    symbol VARCHAR(20) NOT NULL,
    qty NUMERIC(20, 8) DEFAULT 0,
    avg_price NUMERIC(20, 8) DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, symbol)
);

-- 2. Portfolio Management Tables
CREATE TABLE IF NOT EXISTS portfolios (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS portfolio_assets (
    portfolio_id BIGINT REFERENCES portfolios(id),
    symbol VARCHAR(20) NOT NULL,
    weight NUMERIC(5, 2), -- Weight percentage
    PRIMARY KEY (portfolio_id, symbol)
);

-- Indices for performance
CREATE INDEX IF NOT EXISTS idx_paper_orders_user ON paper_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_portfolios_user ON portfolios(user_id);
