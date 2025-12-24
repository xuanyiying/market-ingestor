package risk

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type RiskManager struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewRiskManager(db *pgxpool.Pool, logger *zap.Logger) *RiskManager {
	return &RiskManager{
		db:     db,
		logger: logger,
	}
}

// PreTradeCheck validates if an order can be placed based on risk limits
func (m *RiskManager) PreTradeCheck(ctx context.Context, userID int64, symbol string, side string, qty decimal.Decimal, price decimal.Decimal) error {
	// 1. Check max position size (example: 10% of balance)
	var balance decimal.Decimal
	err := m.db.QueryRow(ctx, "SELECT balance FROM paper_accounts WHERE user_id = $1", userID).Scan(&balance)
	if err != nil {
		return errors.New("failed to get account balance for risk check")
	}

	orderValue := qty.Mul(price)
	if orderValue.GreaterThan(balance.Mul(decimal.NewFromFloat(0.1))) {
		return errors.New("order size exceeds 10% of account balance (risk limit)")
	}

	// 2. Check for existing exposure
	var totalExposure decimal.Decimal
	err = m.db.QueryRow(ctx, "SELECT COALESCE(SUM(qty * avg_price), 0) FROM paper_positions WHERE user_id = $1", userID).Scan(&totalExposure)
	if err != nil {
		return errors.New("failed to calculate total exposure")
	}

	if totalExposure.Add(orderValue).GreaterThan(balance.Mul(decimal.NewFromFloat(0.5))) {
		return errors.New("total portfolio exposure would exceed 50% limit")
	}

	return nil
}

// CheckStopLoss monitors positions and returns a list of symbols to liquidate
func (m *RiskManager) CheckStopLoss(ctx context.Context, userID int64, currentPrices map[string]decimal.Decimal) ([]string, error) {
	rows, err := m.db.Query(ctx, "SELECT symbol, qty, avg_price FROM paper_positions WHERE user_id = $1 AND qty > 0", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var toLiquidate []string
	for rows.Next() {
		var symbol string
		var qty, avgPrice decimal.Decimal
		if err := rows.Scan(&symbol, &qty, &avgPrice); err != nil {
			continue
		}

		price, ok := currentPrices[symbol]
		if !ok {
			continue
		}

		// Hardcoded 5% stop loss for demonstration
		stopLossPrice := avgPrice.Mul(decimal.NewFromFloat(0.95))
		if price.LessThanOrEqual(stopLossPrice) {
			toLiquidate = append(toLiquidate, symbol)
			m.logger.Warn("STOP LOSS TRIGGERED", zap.String("symbol", symbol), zap.String("price", price.String()))
		}
	}

	return toLiquidate, nil
}
