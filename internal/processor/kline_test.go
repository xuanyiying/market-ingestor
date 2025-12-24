package processor

import (
	"fmt"
	"market-ingestor/internal/model"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestKlineProcessor_ProcessTrade(t *testing.T) {
	logger := zap.NewNop()
	p := NewKlineProcessor(nil, logger)

	now := time.Now().Truncate(24 * time.Hour) // Use a large window to avoid truncation issues
	symbol := "BTCUSDT"
	exchange := "binance"

	// 1. First trade creates candles for all periods
	trade1 := model.Trade{
		ID:        "1",
		Symbol:    symbol,
		Exchange:  exchange,
		Price:     decimal.NewFromFloat(50000),
		Amount:    decimal.NewFromFloat(1),
		Timestamp: now.Add(10 * time.Second),
	}
	p.processTrade(trade1)

	// Check 1m candle
	key1m := fmt.Sprintf("binance:BTCUSDT:1m:%s", now.Truncate(time.Minute).Format(time.RFC3339))
	candle1m, ok := p.candles[key1m]
	assert.True(t, ok)
	assert.True(t, candle1m.Open.Equal(decimal.NewFromFloat(50000)))

	// Check 1h candle
	key1h := fmt.Sprintf("binance:BTCUSDT:1h:%s", now.Truncate(time.Hour).Format(time.RFC3339))
	candle1h, ok := p.candles[key1h]
	assert.True(t, ok)
	assert.True(t, candle1h.Open.Equal(decimal.NewFromFloat(50000)))

	// 2. Second trade updates all candles
	trade2 := model.Trade{
		ID:        "2",
		Symbol:    symbol,
		Exchange:  exchange,
		Price:     decimal.NewFromFloat(50100),
		Amount:    decimal.NewFromFloat(0.5),
		Timestamp: now.Add(20 * time.Second),
	}
	p.processTrade(trade2)

	assert.True(t, candle1m.High.Equal(decimal.NewFromFloat(50100)))
	assert.True(t, candle1h.High.Equal(decimal.NewFromFloat(50100)))
	assert.True(t, candle1m.Volume.Equal(decimal.NewFromFloat(1.5)))
}
