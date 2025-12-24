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

type BatchSaver struct {
	pool      *pgxpool.Pool
	logger    *zap.Logger
	buffer    []model.Trade
	mu        sync.Mutex
	flushIntv time.Duration
	batchSize int
}

func NewBatchSaver(pool *pgxpool.Pool, logger *zap.Logger, flushIntv time.Duration, batchSize int) *BatchSaver {
	saver := &BatchSaver{
		pool:      pool,
		logger:    logger,
		buffer:    make([]model.Trade, 0, batchSize),
		flushIntv: flushIntv,
		batchSize: batchSize,
	}
	go saver.run()
	return saver
}

func (s *BatchSaver) Add(trade model.Trade) {
	s.mu.Lock()
	s.buffer = append(s.buffer, trade)
	s.mu.Unlock()

	if len(s.buffer) >= s.batchSize {
		s.Flush()
	}
}

func (s *BatchSaver) run() {
	ticker := time.NewTicker(s.flushIntv)
	defer ticker.Stop()

	for range ticker.C {
		s.Flush()
	}
}

func (s *BatchSaver) Flush() {
	s.mu.Lock()
	if len(s.buffer) == 0 {
		s.mu.Unlock()
		return
	}
	trades := s.buffer
	s.buffer = make([]model.Trade, 0, s.batchSize)
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	batch := &pgx.Batch{}
	for _, t := range trades {
		batch.Queue(`INSERT INTO trades (time, symbol, exchange, price, amount, side, trade_id) 
                     VALUES ($1, $2, $3, $4, $5, $6, $7)
                     ON CONFLICT (symbol, exchange, trade_id, time) DO NOTHING`,
			t.Timestamp, t.Symbol, t.Exchange, t.Price, t.Amount, t.Side, t.ID)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(trades); i++ {
		_, err := br.Exec()
		if err != nil {
			s.logger.Error("failed to execute batch insert", zap.Error(err))
		}
	}
	infrastructure.DBInsertRate.WithLabelValues("trades").Add(float64(len(trades)))
}
