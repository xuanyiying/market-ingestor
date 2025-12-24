-- Migration: Add subscription system

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
