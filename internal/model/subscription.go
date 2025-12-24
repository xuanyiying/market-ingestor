package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type SubscriptionTier struct {
	ID              int             `json:"id" db:"id"`
	Name            string          `json:"name" db:"name"`
	MaxSymbols      int             `json:"max_symbols" db:"max_symbols"`
	RealtimeEnabled bool            `json:"realtime_enabled" db:"realtime_enabled"`
	PriceMonthly    decimal.Decimal `json:"price_monthly" db:"price_monthly"`
}

type UserSubscription struct {
	UserID    int64     `json:"user_id" db:"user_id"`
	TierID    int       `json:"tier_id" db:"tier_id"`
	Status    string    `json:"status" db:"status"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
