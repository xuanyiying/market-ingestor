package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"market-ingestor/internal/model"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// HistoryBackfiller 负责从交易所拉取历史 K 线数据并存入数据库
type HistoryBackfiller struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewHistoryBackfiller(db *pgxpool.Pool, logger *zap.Logger) *HistoryBackfiller {
	return &HistoryBackfiller{
		db:     db,
		logger: logger,
	}
}

// BackfillBinance 从 Binance 拉取历史 1m K 线
func (b *HistoryBackfiller) BackfillBinance(ctx context.Context, symbol string, startTime, endTime time.Time) error {
	b.logger.Info("starting backfill for binance", zap.String("symbol", symbol), zap.Time("start", startTime), zap.Time("end", endTime))

	currentStart := startTime
	for currentStart.Before(endTime) {
		// Binance limit is 1000 per request
		url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=1m&startTime=%d&limit=1000",
			symbol, currentStart.UnixMilli())

		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to fetch from binance: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("binance api returned status: %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var rawData [][]interface{}
		if err := json.Unmarshal(body, &rawData); err != nil {
			return fmt.Errorf("failed to unmarshal binance response: %w", err)
		}

		if len(rawData) == 0 {
			break
		}

		klines := make([]model.KLine, 0, len(rawData))
		var lastTs int64
		for _, r := range rawData {
			// Binance K-line format: [Open time, Open, High, Low, Close, Volume, Close time, ...]
			openTime := int64(r[0].(float64))
			lastTs = openTime
			k := model.KLine{
				Symbol:    symbol,
				Exchange:  "binance",
				Period:    "1m",
				Open:      parseDecimal(r[1].(string)),
				High:      parseDecimal(r[2].(string)),
				Low:       parseDecimal(r[3].(string)),
				Close:     parseDecimal(r[4].(string)),
				Volume:    parseDecimal(r[5].(string)),
				Timestamp: time.UnixMilli(openTime),
			}
			klines = append(klines, k)
		}

		// Save to DB
		if err := b.saveKLines(ctx, klines); err != nil {
			return fmt.Errorf("failed to save klines: %w", err)
		}

		b.logger.Debug("backfilled chunk", zap.String("symbol", symbol), zap.Int("count", len(klines)), zap.Time("last_ts", time.UnixMilli(lastTs)))

		// Update start time for next iteration (lastTs + 1 minute)
		currentStart = time.UnixMilli(lastTs).Add(time.Minute)

		// Respect rate limits
		time.Sleep(200 * time.Millisecond)
	}

	b.logger.Info("backfill completed", zap.String("symbol", symbol))
	return nil
}

func (b *HistoryBackfiller) saveKLines(ctx context.Context, klines []model.KLine) error {
	tx, err := b.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, k := range klines {
		_, err := tx.Exec(ctx,
			`INSERT INTO klines (symbol, exchange, period, open, high, low, close, volume, time)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			 ON CONFLICT (symbol, exchange, period, time) DO UPDATE SET
			 open = EXCLUDED.open, high = EXCLUDED.high, low = EXCLUDED.low, close = EXCLUDED.close, volume = EXCLUDED.volume`,
			k.Symbol, k.Exchange, k.Period, k.Open, k.High, k.Low, k.Close, k.Volume, k.Timestamp)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func parseDecimal(s string) decimal.Decimal {
	d, _ := decimal.NewFromString(s)
	return d
}
