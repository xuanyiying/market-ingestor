package storage

import (
	"context"
	"market-ingestor/internal/infrastructure"
	"market-ingestor/internal/model"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type KlineSaver struct {
	pool      *pgxpool.Pool
	logger    *zap.Logger
	buffer    []model.KLine
	mu        sync.Mutex
	flushIntv time.Duration
	batchSize int
}

func NewKlineSaver(pool *pgxpool.Pool, logger *zap.Logger, flushIntv time.Duration, batchSize int) *KlineSaver {
	saver := &KlineSaver{
		pool:      pool,
		logger:    logger,
		buffer:    make([]model.KLine, 0, batchSize),
		flushIntv: flushIntv,
		batchSize: batchSize,
	}
	go saver.run()
	return saver
}

func (s *KlineSaver) Add(kline model.KLine) {
	s.mu.Lock()
	s.buffer = append(s.buffer, kline)
	s.mu.Unlock()

	if len(s.buffer) >= s.batchSize {
		s.Flush()
	}
}

func (s *KlineSaver) run() {
	ticker := time.NewTicker(s.flushIntv)
	defer ticker.Stop()

	for range ticker.C {
		s.Flush()
	}
}

func (s *KlineSaver) Flush() {
	s.mu.Lock()
	if len(s.buffer) == 0 {
		s.mu.Unlock()
		return
	}
	klines := s.buffer
	s.buffer = make([]model.KLine, 0, s.batchSize)
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	batch := &pgx.Batch{}
	for _, k := range klines {
		batch.Queue(`INSERT INTO klines (time, symbol, exchange, period, open, high, low, close, volume) 
                     VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
                     ON CONFLICT (symbol, exchange, period, time) DO UPDATE SET
                     open = EXCLUDED.open,
                     high = EXCLUDED.high,
                     low = EXCLUDED.low,
                     close = EXCLUDED.close,
                     volume = EXCLUDED.volume`,
			k.Timestamp, k.Symbol, k.Exchange, k.Period, k.Open, k.High, k.Low, k.Close, k.Volume)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(klines); i++ {
		_, err := br.Exec()
		if err != nil {
			s.logger.Error("failed to execute kline batch insert", zap.Error(err))
		}
	}
	infrastructure.DBInsertRate.WithLabelValues("klines").Add(float64(len(klines)))
}
