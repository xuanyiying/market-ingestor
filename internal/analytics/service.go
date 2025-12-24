package analytics

import (
	"context"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type AnalyticsService struct {
	db *pgxpool.Pool
}

func NewAnalyticsService(db *pgxpool.Pool) *AnalyticsService {
	return &AnalyticsService{db: db}
}

type PerformanceReport struct {
	TotalReturn decimal.Decimal `json:"total_return"`
	MaxDrawdown decimal.Decimal `json:"max_drawdown"`
	SharpeRatio float64         `json:"sharpe_ratio"`
	WinRate     float64         `json:"win_rate"`
}

func (s *AnalyticsService) GetPortfolioReport(ctx context.Context, userID int64) (*PerformanceReport, error) {
	// 1. Get current balance
	var balance decimal.Decimal
	err := s.db.QueryRow(ctx, "SELECT balance FROM paper_accounts WHERE user_id = $1", userID).Scan(&balance)
	if err != nil {
		return nil, err
	}

	// 2. Calculate position market value
	rows, err := s.db.Query(ctx, "SELECT symbol, qty, avg_price FROM paper_positions WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	totalMarketValue := balance
	var initialInvestment decimal.Decimal = decimal.NewFromFloat(100000.0) // Assume default start

	for rows.Next() {
		var symbol string
		var qty, avgPrice decimal.Decimal
		if err := rows.Scan(&symbol, &qty, &avgPrice); err != nil {
			continue
		}

		// In a real app, we'd fetch current price from a cache or NATS.
		// Here we'll use avg_price as a baseline or fetch last closed kline.
		var currentPrice decimal.Decimal
		err = s.db.QueryRow(ctx, "SELECT close FROM klines WHERE symbol = $1 ORDER BY time DESC LIMIT 1", symbol).Scan(&currentPrice)
		if err != nil {
			currentPrice = avgPrice // Fallback
		}

		totalMarketValue = totalMarketValue.Add(qty.Mul(currentPrice))
	}

	totalReturn := totalMarketValue.Sub(initialInvestment).Div(initialInvestment).Mul(decimal.NewFromInt(100))

	return &PerformanceReport{
		TotalReturn: totalReturn,
		MaxDrawdown: decimal.NewFromFloat(0), // Placeholder for historical drawdown calc
		SharpeRatio: 1.25,                    // Placeholder
		WinRate:     0.55,                    // Placeholder
	}, nil
}

// MonteCarloSimulation runs a simple simulation based on historical returns
func (s *AnalyticsService) MonteCarloSimulation(ctx context.Context, returns []float64, iterations int, days int) [][]float64 {
	results := make([][]float64, iterations)
	for i := 0; i < iterations; i++ {
		path := make([]float64, days)
		price := 1.0 // Normalized start price
		for d := 0; d < days; d++ {
			// Randomly pick a return from the history
			r := returns[int(uint64(time.Now().UnixNano())%uint64(len(returns)))]
			price = price * (1 + r)
			path[d] = price
		}
		results[i] = path
	}
	return results
}

func CalculateSharpeRatio(returns []float64, riskFreeRate float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	sum := 0.0
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))

	variance := 0.0
	for _, r := range returns {
		variance += math.Pow(r-mean, 2)
	}
	stdDev := math.Sqrt(variance / float64(len(returns)))

	if stdDev == 0 {
		return 0
	}

	return (mean - riskFreeRate) / stdDev
}
