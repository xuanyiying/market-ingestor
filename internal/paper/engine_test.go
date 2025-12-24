package paper

import (
	"market-ingestor/internal/model"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestProcessPriceUpdate(t *testing.T) {
	engine := NewPaperEngine(nil, nil, nil)

	symbol := "BTCUSDT"
	order1 := Order{
		ID:     1,
		Symbol: symbol,
		Type:   "limit",
		Side:   "buy",
		Price:  decimal.NewFromFloat(50000.0),
		Qty:    decimal.NewFromFloat(1.0),
	}
	order2 := Order{
		ID:     2,
		Symbol: symbol,
		Type:   "limit",
		Side:   "sell",
		Price:  decimal.NewFromFloat(60000.0),
		Qty:    decimal.NewFromFloat(1.0),
	}
	order3 := Order{
		ID:     3,
		Symbol: symbol,
		Type:   "market",
		Side:   "buy",
		Qty:    decimal.NewFromFloat(1.0),
	}

	engine.orders[symbol] = []Order{order1, order2, order3}

	// Case 1: Price is 55000 (between buy and sell limit)
	candle1 := model.KLine{
		Symbol: symbol,
		Open:   decimal.NewFromFloat(55000.0),
		High:   decimal.NewFromFloat(55500.0),
		Low:    decimal.NewFromFloat(54500.0),
		Close:  decimal.NewFromFloat(55000.0),
	}

	engine.processPriceUpdate(candle1)

	// Market order (3) should be filled.
	// Limit orders (1, 2) should NOT be filled.
	assert.Len(t, engine.orders[symbol], 2)
	assert.Equal(t, int64(1), engine.orders[symbol][0].ID)
	assert.Equal(t, int64(2), engine.orders[symbol][1].ID)

	filled := <-engine.fillChan
	assert.Equal(t, int64(3), filled.ID)
	assert.Equal(t, decimal.NewFromFloat(55000.0), filled.FilledPrice)

	// Case 2: Price drops to 49000 (triggers buy limit)
	candle2 := model.KLine{
		Symbol: symbol,
		Open:   decimal.NewFromFloat(50000.0),
		High:   decimal.NewFromFloat(50500.0),
		Low:    decimal.NewFromFloat(49000.0),
		Close:  decimal.NewFromFloat(49500.0),
	}

	engine.processPriceUpdate(candle2)

	assert.Len(t, engine.orders[symbol], 1)
	assert.Equal(t, int64(2), engine.orders[symbol][0].ID)

	filled = <-engine.fillChan
	assert.Equal(t, int64(1), filled.ID)
	assert.Equal(t, decimal.NewFromFloat(50000.0), filled.FilledPrice)

	// Case 3: Price rises to 61000 (triggers sell limit)
	candle3 := model.KLine{
		Symbol: symbol,
		Open:   decimal.NewFromFloat(55000.0),
		High:   decimal.NewFromFloat(61000.0),
		Low:    decimal.NewFromFloat(55000.0),
		Close:  decimal.NewFromFloat(60500.0),
	}

	engine.processPriceUpdate(candle3)

	assert.Len(t, engine.orders[symbol], 0)

	filled = <-engine.fillChan
	assert.Equal(t, int64(2), filled.ID)
	assert.Equal(t, decimal.NewFromFloat(60000.0), filled.FilledPrice)
}
