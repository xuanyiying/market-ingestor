package analytics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateSharpeRatio(t *testing.T) {
	// Case 1: Simple positive returns
	returns := []float64{0.01, 0.02, -0.01, 0.03, 0.01}
	riskFreeRate := 0.001

	sharpe := CalculateSharpeRatio(returns, riskFreeRate)
	assert.True(t, sharpe > 0)

	// Case 2: Zero returns
	returnsZero := []float64{0, 0, 0}
	assert.Equal(t, 0.0, CalculateSharpeRatio(returnsZero, riskFreeRate))

	// Case 3: Empty returns
	assert.Equal(t, 0.0, CalculateSharpeRatio([]float64{}, riskFreeRate))
}
