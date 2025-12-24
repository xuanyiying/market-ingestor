package indicators

import (
	"github.com/shopspring/decimal"
)

// SMA (Simple Moving Average)
func CalculateSMA(data []decimal.Decimal, period int) []decimal.Decimal {
	if len(data) < period {
		return make([]decimal.Decimal, len(data))
	}

	sma := make([]decimal.Decimal, len(data))
	sum := decimal.Zero

	for i := 0; i < len(data); i++ {
		sum = sum.Add(data[i])
		if i >= period {
			sum = sum.Sub(data[i-period])
		}

		if i >= period-1 {
			sma[i] = sum.Div(decimal.NewFromInt(int64(period)))
		} else {
			sma[i] = decimal.Zero
		}
	}

	return sma
}

// EMA (Exponential Moving Average)
func CalculateEMA(data []decimal.Decimal, period int) []decimal.Decimal {
	if len(data) == 0 {
		return nil
	}

	ema := make([]decimal.Decimal, len(data))
	multiplier := decimal.NewFromFloat(2.0).Div(decimal.NewFromInt(int64(period + 1)))

	// Initial EMA is SMA or just the first value
	ema[0] = data[0]

	for i := 1; i < len(data); i++ {
		// EMA = (Close - EMA_prev) * Multiplier + EMA_prev
		ema[i] = data[i].Sub(ema[i-1]).Mul(multiplier).Add(ema[i-1])
	}

	return ema
}
