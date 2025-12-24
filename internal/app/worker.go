package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"market-ingestor/internal/connector"
	"market-ingestor/internal/engine"
	"market-ingestor/internal/infrastructure"
	"market-ingestor/internal/model"
	"market-ingestor/internal/storage"
	"market-ingestor/internal/strategy"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// NormalizeSymbol unifies different exchange symbol formats into a standard one (e.g. BTCUSDT)
func NormalizeSymbol(s string) string {
	s = strings.ToUpper(s)
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}

// startIngestionWorker initializes and starts the ingestion connectors
func (a *App) startIngestionWorker(ctx context.Context) {
	targets := []struct {
		Exchange string
		Symbol   string
	}{
		{"binance", "btcusdt"},
		{"okx", "BTC-USDT"},
		{"bybit", "BTCUSDT"},
		{"coinbase", "BTC-USD"},
		{"kraken", "XBT/USD"},
	}

	for _, target := range targets {
		t := target
		go func() {
			tradeChan := make(chan model.Trade, 1000)
			var c interface {
				Run(context.Context, chan<- model.Trade)
			}

			switch t.Exchange {
			case "binance":
				c = connector.NewBinanceConnector(a.Logger, t.Symbol)
			case "okx":
				c = connector.NewOKXConnector(a.Logger, t.Symbol)
			case "bybit":
				c = connector.NewBybitConnector(a.Logger, t.Symbol)
			case "coinbase":
				c = connector.NewCoinbaseConnector(a.Logger, t.Symbol)
			case "kraken":
				c = connector.NewKrakenConnector(a.Logger, t.Symbol)
			default:
				a.Logger.Warn("unknown exchange", zap.String("exchange", t.Exchange))
				return
			}

			go c.Run(ctx, tradeChan)

			for {
				select {
				case <-ctx.Done():
					return
				case trade := <-tradeChan:
					trade.Symbol = NormalizeSymbol(trade.Symbol)

					subject := fmt.Sprintf("market.raw.%s.%s", trade.Exchange, trade.Symbol)
					data, err := json.Marshal(trade)
					if err != nil {
						a.Logger.Error("failed to marshal trade", zap.Error(err))
						continue
					}
					_, err = a.JS.Publish(subject, data)
					if err != nil {
						a.Logger.Error("failed to publish to NATS", zap.Error(err))
					}
					infrastructure.TradeProcessRate.WithLabelValues(trade.Symbol).Inc()
				}
			}
		}()
	}
}

// startPersistenceService subscribes to NATS and saves trades and klines to the database
func (a *App) startPersistenceService(tradeSaver *storage.BatchSaver, klineSaver *storage.KlineSaver) {
	// 1. Subscribe to raw trades
	_, err := a.JS.Subscribe("market.raw.*.*", func(m *nats.Msg) {
		var trade model.Trade
		if err := json.Unmarshal(m.Data, &trade); err != nil {
			a.Logger.Error("failed to unmarshal trade", zap.Error(err))
			return
		}
		tradeSaver.Add(trade)
	}, nats.Durable("trade_saver"), nats.ManualAck())
	if err != nil {
		a.Logger.Fatal("failed to subscribe to trades", zap.Error(err))
	}

	// 2. Subscribe to K-lines
	_, err = a.JS.Subscribe("market.kline.*.*", func(m *nats.Msg) {
		var kline model.KLine
		if err := json.Unmarshal(m.Data, &kline); err != nil {
			a.Logger.Error("failed to unmarshal kline", zap.Error(err))
			return
		}
		klineSaver.Add(kline)
	}, nats.Durable("kline_saver"), nats.ManualAck())
	if err != nil {
		a.Logger.Fatal("failed to subscribe to klines", zap.Error(err))
	}
}

// startStrategyRunner initializes and starts the live strategy runner
func (a *App) startStrategyRunner(ctx context.Context) {
	runner := engine.NewStrategyRunner(a.JS, a.Logger)

	// Add default strategy for testing
	maCross, err := strategy.NewStrategy("ma_cross_v2", map[string]interface{}{
		"short_period": float64(5),
		"long_period":  float64(20),
	})
	if err != nil {
		a.Logger.Error("failed to create strategy", zap.Error(err))
		return
	}
	runner.AddStrategy(maCross)

	if err := runner.Run(ctx); err != nil {
		a.Logger.Error("failed to start strategy runner", zap.Error(err))
	}
}
