package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"market-ingestor/internal/infrastructure"
	"market-ingestor/internal/model"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type KlineProcessor struct {
	js         nats.JetStreamContext
	logger     *zap.Logger
	candles    map[string]*model.KLine
	mu         sync.Mutex
	jobs       chan model.Trade
	numWorkers int
}

func NewKlineProcessor(js nats.JetStreamContext, logger *zap.Logger) *KlineProcessor {
	return &KlineProcessor{
		js:         js,
		logger:     logger,
		candles:    make(map[string]*model.KLine),
		jobs:       make(chan model.Trade, 1000),
		numWorkers: 4, // Configurable based on CPU cores
	}
}

func (p *KlineProcessor) Run(ctx context.Context) error {
	// Start workers
	for i := 0; i < p.numWorkers; i++ {
		go p.worker(ctx)
	}

	_, err := p.js.Subscribe("market.raw.*.*", func(msg *nats.Msg) {
		var trade model.Trade
		if err := json.Unmarshal(msg.Data, &trade); err != nil {
			p.logger.Error("failed to unmarshal trade in processor", zap.Error(err))
			return
		}
		infrastructure.TradeProcessRate.WithLabelValues(trade.Symbol).Inc()

		// Send to worker pool instead of direct processing
		select {
		case p.jobs <- trade:
		default:
			p.logger.Warn("processor job queue full, trade dropped", zap.String("symbol", trade.Symbol))
		}

		msg.Ack()
	}, nats.Durable("kline-processor"), nats.ManualAck())

	if err != nil {
		return err
	}

	go p.flushLoop(ctx)
	p.logger.Info("kline processor started")
	return nil
}

func (p *KlineProcessor) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case trade, ok := <-p.jobs:
			if !ok {
				return
			}
			p.processTrade(trade)
		}
	}
}

func (p *KlineProcessor) processTrade(trade model.Trade) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, period := range model.SupportedPeriods {
		duration := model.PeriodToDuration(period)
		window := trade.Timestamp.Truncate(duration)
		key := fmt.Sprintf("%s:%s:%s:%s", trade.Exchange, trade.Symbol, period, window.Format(time.RFC3339))

		candle, ok := p.candles[key]
		if !ok {
			candle = &model.KLine{
				Symbol:    trade.Symbol,
				Exchange:  trade.Exchange,
				Period:    period,
				Open:      trade.Price,
				High:      trade.Price,
				Low:       trade.Price,
				Close:     trade.Price,
				Volume:    trade.Amount,
				Timestamp: window,
			}
			p.candles[key] = candle
		} else {
			if trade.Price.GreaterThan(candle.High) {
				candle.High = trade.Price
			}
			if trade.Price.LessThan(candle.Low) {
				candle.Low = trade.Price
			}
			candle.Close = trade.Price
			candle.Volume = candle.Volume.Add(trade.Amount)
		}
	}
}

func (p *KlineProcessor) flushLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second) // Flush more frequently to handle multiple periods
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.flush()
		}
	}
}

func (p *KlineProcessor) flush() {
	p.mu.Lock()
	now := time.Now()
	toFlush := make([]*model.KLine, 0)

	for key, candle := range p.candles {
		duration := model.PeriodToDuration(candle.Period)
		// If current time is after the end of this candle window
		if now.After(candle.Timestamp.Add(duration)) {
			toFlush = append(toFlush, candle)
			delete(p.candles, key)
		}
	}
	p.mu.Unlock()

	for _, candle := range toFlush {
		subject := fmt.Sprintf("market.kline.%s.%s", candle.Period, candle.Symbol)
		data, _ := json.Marshal(candle)
		_, err := p.js.Publish(subject, data)
		if err != nil {
			p.logger.Error("failed to publish kline", zap.String("period", candle.Period), zap.Error(err))
		}
	}
}
