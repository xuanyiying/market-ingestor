package indicators

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCalculateSMA(t *testing.T) {
	data := []decimal.Decimal{
		decimal.NewFromInt(10),
		decimal.NewFromInt(20),
		decimal.NewFromInt(30),
		decimal.NewFromInt(40),
		decimal.NewFromInt(50),
	}
	period := 3
	expected := []decimal.Decimal{
		decimal.Zero,
		decimal.Zero,
		decimal.NewFromInt(20), // (10+20+30)/3
		decimal.NewFromInt(30), // (20+30+40)/3
		decimal.NewFromInt(40), // (30+40+50)/3
	}

	result := CalculateSMA(data, period)
	for i := range expected {
		assert.True(t, expected[i].Equal(result[i]), "Mismatch at index %d: expected %v, got %v", i, expected[i], result[i])
	}
}

func TestCalculateEMA(t *testing.T) {
	data := []decimal.Decimal{
		decimal.NewFromInt(10),
		decimal.NewFromInt(20),
		decimal.NewFromInt(30),
	}
	// Multiplier = 2 / (2+1) = 0.6666...
	// EMA[0] = 10
	// EMA[1] = (20 - 10) * 0.6666 + 10 = 16.666...
	// EMA[2] = (30 - 16.666) * 0.6666 + 16.666 = 25.555...

	result := CalculateEMA(data, 2)
	assert.Equal(t, 3, len(result))
	assert.True(t, decimal.NewFromInt(10).Equal(result[0]))
}
