-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- 1. Trades Table
CREATE TABLE IF NOT EXISTS trades (
    trade_id TEXT NOT NULL,
    time TIMESTAMPTZ NOT NULL,
    symbol TEXT NOT NULL,
    exchange TEXT NOT NULL,
    price NUMERIC NOT NULL,
    amount NUMERIC NOT NULL,
    side TEXT, -- 'buy' or 'sell'
    PRIMARY KEY (symbol, exchange, trade_id, time)
);

-- Convert to hypertable (with exception handling for "already exists")
DO $$
BEGIN
    BEGIN
        PERFORM create_hypertable('trades', 'time');
    EXCEPTION
        WHEN others THEN
            RAISE NOTICE 'Table trades is already a hypertable or failed to convert: %', SQLERRM;
    END;
END $$;

CREATE INDEX IF NOT EXISTS idx_trades_symbol_time ON trades (symbol, time DESC);

-- 2. K-Line Table
CREATE TABLE IF NOT EXISTS klines (
    time TIMESTAMPTZ NOT NULL,
    symbol TEXT NOT NULL,
    exchange TEXT NOT NULL,
    period TEXT NOT NULL, -- '1m', '5m', '1h'
    open NUMERIC NOT NULL,
    high NUMERIC NOT NULL,
    low NUMERIC NOT NULL,
    close NUMERIC NOT NULL,
    volume NUMERIC NOT NULL,
    PRIMARY KEY (symbol, exchange, period, time)
);

-- Convert to hypertable (with exception handling for "already exists")
DO $$
BEGIN
    BEGIN
        PERFORM create_hypertable('klines', 'time');
    EXCEPTION
        WHEN others THEN
            RAISE NOTICE 'Table klines is already a hypertable or failed to convert: %', SQLERRM;
    END;
END $$;

CREATE INDEX IF NOT EXISTS idx_klines_symbol_period_time ON klines (symbol, period, time DESC);

-- 3. Business Tables
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW (),
    updated_at TIMESTAMPTZ DEFAULT NOW ()
);

CREATE TABLE IF NOT EXISTS user_exchange_keys (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users (id) ON DELETE CASCADE,
    exchange VARCHAR(50) NOT NULL,
    api_key VARCHAR(255) NOT NULL,
    api_secret VARCHAR(255) NOT NULL,
    label VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW ()
);

CREATE TABLE IF NOT EXISTS strategies (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users (id),
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    config JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW ()
);

CREATE TABLE IF NOT EXISTS backtest_runs (
    id BIGSERIAL PRIMARY KEY,
    strategy_id BIGINT REFERENCES strategies (id),
    symbol VARCHAR(50) NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    report JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW ()
);

-- 4. Subscription System
CREATE TABLE IF NOT EXISTS subscription_tiers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE, -- Free, Pro, Enterprise
    max_symbols INT NOT NULL,
    realtime_enabled BOOLEAN DEFAULT FALSE,
    price_monthly NUMERIC(10, 2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_subscriptions (
    user_id BIGINT PRIMARY KEY,
    tier_id INT REFERENCES subscription_tiers(id),
    status VARCHAR(20) DEFAULT 'active', -- active, expired, canceled
    expires_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Seed basic tiers
INSERT INTO subscription_tiers (name, max_symbols, realtime_enabled, price_monthly) 
VALUES ('Free', 1, FALSE, 0.00) ON CONFLICT DO NOTHING;
INSERT INTO subscription_tiers (name, max_symbols, realtime_enabled, price_monthly) 
VALUES ('Pro', 10, TRUE, 29.00) ON CONFLICT DO NOTHING;
INSERT INTO subscription_tiers (name, max_symbols, realtime_enabled, price_monthly) 
VALUES ('Enterprise', 1000, TRUE, 99.00) ON CONFLICT DO NOTHING;

-- 5. Alert System
CREATE TABLE IF NOT EXISTS alerts (
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    symbol VARCHAR(20) NOT NULL,
    condition_type VARCHAR(50) NOT NULL, -- price_above, price_below
    target_value NUMERIC(20, 8),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alerts_user_id ON alerts(user_id);
CREATE INDEX IF NOT EXISTS idx_alerts_symbol_active ON alerts(symbol) WHERE is_active = TRUE;
