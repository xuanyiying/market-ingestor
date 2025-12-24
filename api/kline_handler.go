package api

import (
	"context"
	"market-ingestor/internal/engine"
	"market-ingestor/internal/model"
	"market-ingestor/internal/storage"
	"market-ingestor/internal/strategy"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func (h *Handler) GetHistoryKLines(c *gin.Context) {
	symbol := strings.ReplaceAll(strings.ToUpper(c.Param("symbol")), "-", "")
	symbol = strings.ReplaceAll(symbol, "/", "")
	period := c.DefaultQuery("period", "1m")

	rows, err := h.db.Query(c.Request.Context(),
		"SELECT symbol, exchange, open, high, low, close, volume, time FROM klines WHERE symbol = $1 AND period = $2 ORDER BY time DESC LIMIT 100",
		symbol, period)
	if err != nil {
		h.logger.Error("failed to query klines", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	klines := make([]model.KLine, 0)
	for rows.Next() {
		var k model.KLine
		if err := rows.Scan(&k.Symbol, &k.Exchange, &k.Open, &k.High, &k.Low, &k.Close, &k.Volume, &k.Timestamp); err != nil {
			h.logger.Error("failed to scan kline", zap.Error(err))
			continue
		}
		k.Period = period
		klines = append(klines, k)
	}

	c.JSON(http.StatusOK, klines)
}

func (h *Handler) RunBacktest(c *gin.Context) {
	var req struct {
		Symbol         string                 `json:"symbol" binding:"required"`
		StrategyType   string                 `json:"strategy_type" binding:"required"`
		Config         map[string]interface{} `json:"config"`
		InitialBalance decimal.Decimal        `json:"initial_balance"`
		StartTime      time.Time              `json:"start_time" binding:"required"`
		EndTime        time.Time              `json:"end_time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	symbol := strings.ReplaceAll(strings.ToUpper(req.Symbol), "-", "")
	symbol = strings.ReplaceAll(symbol, "/", "")

	// 1. Fetch history data for backtest
	rows, err := h.db.Query(c.Request.Context(),
		"SELECT symbol, exchange, open, high, low, close, volume, time FROM klines WHERE symbol = $1 AND time BETWEEN $2 AND $3 ORDER BY time ASC",
		symbol, req.StartTime, req.EndTime)
	if err != nil {
		h.logger.Error("failed to fetch history for backtest", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch data"})
		return
	}
	defer rows.Close()

	klines := make([]model.KLine, 0)
	for rows.Next() {
		var k model.KLine
		if err := rows.Scan(&k.Symbol, &k.Exchange, &k.Open, &k.High, &k.Low, &k.Close, &k.Volume, &k.Timestamp); err != nil {
			continue
		}
		klines = append(klines, k)
	}

	// 2. Setup Strategy
	strat, err := strategy.NewStrategy(req.StrategyType, req.Config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. Run Backtest
	tester := engine.NewBacktester(strat, req.InitialBalance)
	report := tester.Run(klines)

	c.JSON(http.StatusOK, report)
}

func (h *Handler) TriggerBackfill(c *gin.Context) {
	var req struct {
		Exchange  string    `json:"exchange" binding:"required"`
		Symbol    string    `json:"symbol" binding:"required"`
		StartTime time.Time `json:"start_time" binding:"required"`
		EndTime   time.Time `json:"end_time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	backfiller := storage.NewHistoryBackfiller(h.db, h.logger)

	// Async run backfill
	go func() {
		ctx := context.Background()
		var err error
		switch strings.ToLower(req.Exchange) {
		case "binance":
			err = backfiller.BackfillBinance(ctx, strings.ToUpper(req.Symbol), req.StartTime, req.EndTime)
		default:
			h.logger.Warn("unsupported exchange for backfill", zap.String("exchange", req.Exchange))
			return
		}

		if err != nil {
			h.logger.Error("backfill failed", zap.Error(err), zap.String("symbol", req.Symbol))
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{"message": "backfill task started"})
}
