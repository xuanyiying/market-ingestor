package api

import (
	"market-ingestor/internal/analytics"
	"market-ingestor/internal/payment"
	"market-ingestor/internal/risk"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Handler struct {
	db        *pgxpool.Pool
	logger    *zap.Logger
	risk      *risk.RiskManager
	analytics *analytics.AnalyticsService
	stripe    *payment.StripeService
}

func NewHandler(db *pgxpool.Pool, logger *zap.Logger) *Handler {
	stripeKey := os.Getenv("STRIPE_API_KEY")
	return &Handler{
		db:        db,
		logger:    logger,
		risk:      risk.NewRiskManager(db, logger),
		analytics: analytics.NewAnalyticsService(db),
		stripe:    payment.NewStripeService(db, logger, stripeKey),
	}
}
