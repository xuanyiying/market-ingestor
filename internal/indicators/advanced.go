package indicators

import (
	"market-ingestor/internal/model"
	"math"

	"github.com/shopspring/decimal"
)

// RSI (Relative Strength Index)
func CalculateRSI(data []decimal.Decimal, period int) []decimal.Decimal {
	if len(data) < period+1 {
		return make([]decimal.Decimal, len(data))
	}

	rsi := make([]decimal.Decimal, len(data))
	avgGain := decimal.Zero
	avgLoss := decimal.Zero

	// Initial Avg Gain/Loss
	for i := 1; i <= period; i++ {
		change := data[i].Sub(data[i-1])
		if change.GreaterThan(decimal.Zero) {
			avgGain = avgGain.Add(change)
		} else {
			avgLoss = avgLoss.Add(change.Abs())
		}
	}

	avgGain = avgGain.Div(decimal.NewFromInt(int64(period)))
	avgLoss = avgLoss.Div(decimal.NewFromInt(int64(period)))

	rsi[period] = calculateRSIValue(avgGain, avgLoss)

	for i := period + 1; i < len(data); i++ {
		change := data[i].Sub(data[i-1])
		gain := decimal.Zero
		loss := decimal.Zero
		if change.GreaterThan(decimal.Zero) {
			gain = change
		} else {
			loss = change.Abs()
		}

		// Smoothing
		avgGain = avgGain.Mul(decimal.NewFromInt(int64(period - 1))).Add(gain).Div(decimal.NewFromInt(int64(period)))
		avgLoss = avgLoss.Mul(decimal.NewFromInt(int64(period - 1))).Add(loss).Div(decimal.NewFromInt(int64(period)))

		rsi[i] = calculateRSIValue(avgGain, avgLoss)
	}

	return rsi
}

func calculateRSIValue(avgGain, avgLoss decimal.Decimal) decimal.Decimal {
	if avgLoss.IsZero() {
		return decimal.NewFromInt(100)
	}
	rs := avgGain.Div(avgLoss)
	return decimal.NewFromInt(100).Sub(decimal.NewFromInt(100).Div(decimal.NewFromInt(1).Add(rs)))
}

// MACD (Moving Average Convergence Divergence)
// Returns: MACD Line, Signal Line, Histogram
func CalculateMACD(data []decimal.Decimal, fastPeriod, slowPeriod, signalPeriod int) ([]decimal.Decimal, []decimal.Decimal, []decimal.Decimal) {
	fastEMA := CalculateEMA(data, fastPeriod)
	slowEMA := CalculateEMA(data, slowPeriod)

	macdLine := make([]decimal.Decimal, len(data))
	for i := 0; i < len(data); i++ {
		macdLine[i] = fastEMA[i].Sub(slowEMA[i])
	}

	signalLine := CalculateEMA(macdLine, signalPeriod)

	histogram := make([]decimal.Decimal, len(data))
	for i := 0; i < len(data); i++ {
		histogram[i] = macdLine[i].Sub(signalLine[i])
	}

	return macdLine, signalLine, histogram
}

// CalculateBollingerBands calculates middle, upper, and lower bands
func CalculateBollingerBands(candles []model.KLine, period int, stdDev float64) ([]float64, []float64, []float64) {
	if len(candles) < period {
		return nil, nil, nil
	}

	data := make([]decimal.Decimal, len(candles))
	for i, c := range candles {
		data[i] = c.Close
	}

	sma := CalculateSMA(data, period)
	middleBand := make([]float64, len(candles))
	upperBand := make([]float64, len(candles))
	lowerBand := make([]float64, len(candles))

	for i := period - 1; i < len(candles); i++ {
		sumSqDiff := 0.0
		mean, _ := sma[i].Float64()
		for j := i - period + 1; j <= i; j++ {
			val, _ := candles[j].Close.Float64()
			sumSqDiff += math.Pow(val-mean, 2)
		}
		variance := sumSqDiff / float64(period)
		sd := math.Sqrt(variance)

		middleBand[i] = mean
		upperBand[i] = mean + (stdDev * sd)
		lowerBand[i] = mean - (stdDev * sd)
	}

	return middleBand, upperBand, lowerBand
}

// CalculateATR calculates the Average True Range
func CalculateATR(candles []model.KLine, period int) []float64 {
	if len(candles) < period+1 {
		return nil
	}

	tr := make([]float64, len(candles))
	for i := 1; i < len(candles); i++ {
		high, _ := candles[i].High.Float64()
		low, _ := candles[i].Low.Float64()
		prevClose, _ := candles[i-1].Close.Float64()

		v1 := high - low
		v2 := math.Abs(high - prevClose)
		v3 := math.Abs(low - prevClose)
		tr[i] = math.Max(v1, math.Max(v2, v3))
	}

	atr := make([]float64, len(candles))
	// Initial ATR is an SMA of TR
	sum := 0.0
	for i := 1; i <= period; i++ {
		sum += tr[i]
	}
	atr[period] = sum / float64(period)

	for i := period + 1; i < len(candles); i++ {
		atr[i] = (atr[i-1]*float64(period-1) + tr[i]) / float64(period)
	}

	return atr
}
