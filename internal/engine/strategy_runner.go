package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"market-ingestor/internal/model"
	"market-ingestor/internal/strategy"
	"strings"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// StrategyRunner 负责在实盘数据流上运行策略
type StrategyRunner struct {
	js         nats.JetStreamContext
	logger     *zap.Logger
	strategies []strategy.Strategy
}

func NewStrategyRunner(js nats.JetStreamContext, logger *zap.Logger) *StrategyRunner {
	return &StrategyRunner{
		js:     js,
		logger: logger,
	}
}

// AddStrategy 添加要运行的策略
func (r *StrategyRunner) AddStrategy(s strategy.Strategy) {
	r.strategies = append(r.strategies, s)
}

// Run 启动策略运行引擎
func (r *StrategyRunner) Run(ctx context.Context) error {
	// 订阅所有周期的 K 线数据
	_, err := r.js.Subscribe("market.kline.*.*", func(msg *nats.Msg) {
		var candle model.KLine
		if err := json.Unmarshal(msg.Data, &candle); err != nil {
			r.logger.Error("failed to unmarshal kline in strategy runner", zap.Error(err))
			return
		}

		r.executeStrategies(candle)
		msg.Ack()
	}, nats.Durable("strategy-runner"), nats.ManualAck())

	if err != nil {
		return err
	}

	r.logger.Info("strategy runner started", zap.Int("strategy_count", len(r.strategies)))
	return nil
}

func (r *StrategyRunner) executeStrategies(candle model.KLine) {
	for _, s := range r.strategies {
		action := s.OnCandle(candle)
		if action != strategy.ActionHold {
			r.logger.Info("STRATEGY SIGNAL",
				zap.String("strategy", s.Name()),
				zap.String("symbol", candle.Symbol),
				zap.String("period", candle.Period),
				zap.String("action", string(action)),
				zap.String("price", candle.Close.String()),
				zap.Time("time", candle.Timestamp),
			)

			// 将信号推送到 NATS，以便 UI 实时展示
			signalSubject := fmt.Sprintf("strategy.signal.%s.%s", s.Name(), candle.Symbol)
			signalData := map[string]interface{}{
				"strategy": s.Name(),
				"symbol":   candle.Symbol,
				"period":   candle.Period,
				"action":   strings.ToLower(string(action)),
				"price":    candle.Close.String(),
				"time":     candle.Timestamp,
			}
			data, _ := json.Marshal(signalData)
			r.js.Publish(signalSubject, data)
		}
	}
}
