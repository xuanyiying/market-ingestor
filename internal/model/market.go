package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Trade 代表一笔实时成交
type Trade struct {
	ID        string          `json:"id" db:"trade_id"`
	Symbol    string          `json:"symbol" db:"symbol"`
	Exchange  string          `json:"exchange" db:"exchange"`
	Price     decimal.Decimal `json:"price" db:"price"`
	Amount    decimal.Decimal `json:"amount" db:"amount"`
	Side      string          `json:"side" db:"side"` // "buy" or "sell"
	Timestamp time.Time       `json:"ts" db:"time"`
}

// KLine (Candle) 代表一根K线
type KLine struct {
	Symbol    string          `json:"symbol" db:"symbol"`
	Exchange  string          `json:"exchange" db:"exchange"`
	Period    string          `json:"period" db:"period"` // "1m", "5m", "15m", "1h", "4h", "1d"
	Open      decimal.Decimal `json:"o" db:"open"`
	High      decimal.Decimal `json:"h" db:"high"`
	Low       decimal.Decimal `json:"l" db:"low"`
	Close     decimal.Decimal `json:"c" db:"close"`
	Volume    decimal.Decimal `json:"v" db:"volume"`
	Timestamp time.Time       `json:"t" db:"time"`
}

// SupportedPeriods 定义系统支持的 K 线周期
var SupportedPeriods = []string{"1m", "5m", "15m", "1h", "4h", "1d"}

// PeriodToDuration 将字符串周期转换为 time.Duration
func PeriodToDuration(period string) time.Duration {
	switch period {
	case "1m":
		return time.Minute
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "1h":
		return time.Hour
	case "4h":
		return 4 * time.Hour
	case "1d":
		return 24 * time.Hour
	default:
		return time.Minute
	}
}

// OrderBook 代表深度快照 (用于回测时的高精度模拟)
type OrderBook struct {
	Symbol    string      `json:"s"`
	Timestamp time.Time   `json:"t"`
	Bids      [][2]string `json:"b"` // 使用 string 防止精度丢失，[Price, Amount]
	Asks      [][2]string `json:"a"`
}
