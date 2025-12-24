package paper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"market-ingestor/internal/model"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type Order struct {
	ID          int64
	UserID      int64
	Symbol      string
	Side        string
	Type        string
	Price       decimal.Decimal
	Qty         decimal.Decimal
	Status      string
	FilledPrice decimal.Decimal
}

type PaperEngine struct {
	db       *pgxpool.Pool
	js       nats.JetStreamContext
	logger   *zap.Logger
	orders   map[string][]Order // key: symbol
	mu       sync.RWMutex
	fillChan chan Order
}

func NewPaperEngine(db *pgxpool.Pool, js nats.JetStreamContext, logger *zap.Logger) *PaperEngine {
	return &PaperEngine{
		db:       db,
		js:       js,
		logger:   logger,
		orders:   make(map[string][]Order),
		fillChan: make(chan Order, 1000),
	}
}

func (e *PaperEngine) Start(ctx context.Context) error {
	// 1. Load open orders from DB
	if err := e.loadOpenOrders(ctx); err != nil {
		return err
	}

	// 2. Subscribe to price updates
	_, err := e.js.Subscribe("market.kline.1m.*", func(msg *nats.Msg) {
		var candle model.KLine
		if err := json.Unmarshal(msg.Data, &candle); err != nil {
			return
		}
		e.processPriceUpdate(candle)
	})

	if err != nil {
		return err
	}

	go e.batchFlushLoop(ctx)
	e.logger.Info("paper trading engine started with batching")
	return nil
}

func (e *PaperEngine) loadOpenOrders(ctx context.Context) error {
	rows, err := e.db.Query(ctx, "SELECT id, user_id, symbol, side, type, price, qty, status FROM paper_orders WHERE status = 'open'")
	if err != nil {
		return err
	}
	defer rows.Close()

	e.mu.Lock()
	defer e.mu.Unlock()
	e.orders = make(map[string][]Order)

	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Symbol, &o.Side, &o.Type, &o.Price, &o.Qty, &o.Status); err != nil {
			continue
		}
		e.orders[o.Symbol] = append(e.orders[o.Symbol], o)
	}
	return nil
}

func (e *PaperEngine) processPriceUpdate(candle model.KLine) {
	e.mu.Lock()
	orders, ok := e.orders[candle.Symbol]
	if !ok || len(orders) == 0 {
		e.mu.Unlock()
		return
	}

	var remaining []Order
	var toFill []Order

	for _, o := range orders {
		filled := false
		switch o.Type {
		case "market":
			filled = true
			o.FilledPrice = candle.Close
		case "limit":
			if o.Side == "buy" && candle.Low.LessThanOrEqual(o.Price) {
				filled = true
				o.FilledPrice = o.Price
			} else if o.Side == "sell" && candle.High.GreaterThanOrEqual(o.Price) {
				filled = true
				o.FilledPrice = o.Price
			}
		}

		if filled {
			toFill = append(toFill, o)
		} else {
			remaining = append(remaining, o)
		}
	}

	if len(toFill) > 0 {
		e.orders[candle.Symbol] = remaining
	}
	e.mu.Unlock()

	for _, o := range toFill {
		e.fillChan <- o
	}
}

func (e *PaperEngine) batchFlushLoop(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var batch []Order
	for {
		select {
		case <-ctx.Done():
			return
		case o := <-e.fillChan:
			batch = append(batch, o)
			if len(batch) >= 50 {
				e.flushBatch(ctx, batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				e.flushBatch(ctx, batch)
				batch = nil
			}
		}
	}
}

func (e *PaperEngine) flushBatch(ctx context.Context, batch []Order) {
	tx, err := e.db.Begin(ctx)
	if err != nil {
		e.logger.Error("failed to start batch transaction", zap.Error(err))
		return
	}
	defer tx.Rollback(ctx)

	for _, o := range batch {
		// 1. Update order status
		_, err = tx.Exec(ctx, "UPDATE paper_orders SET status = 'filled', filled_price = $1, filled_time = NOW() WHERE id = $2",
			o.FilledPrice, o.ID)
		if err != nil {
			continue
		}

		// 2. Update balance
		amount := o.Qty.Mul(o.FilledPrice)
		if o.Side == "buy" {
			_, err = tx.Exec(ctx, "UPDATE paper_accounts SET balance = balance - $1 WHERE user_id = $2", amount, o.UserID)
		} else {
			_, err = tx.Exec(ctx, "UPDATE paper_accounts SET balance = balance + $1 WHERE user_id = $2", amount, o.UserID)
		}
		if err != nil {
			continue
		}

		// 3. Update positions
		if o.Side == "buy" {
			_, err = tx.Exec(ctx, `
				INSERT INTO paper_positions (user_id, symbol, qty, avg_price) 
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (user_id, symbol) DO UPDATE SET
					avg_price = (paper_positions.qty * paper_positions.avg_price + $3 * $4) / (paper_positions.qty + $3),
					qty = paper_positions.qty + $3`,
				o.UserID, o.Symbol, o.Qty, o.FilledPrice)
		} else {
			_, err = tx.Exec(ctx, "UPDATE paper_positions SET qty = qty - $1 WHERE user_id = $2 AND symbol = $3",
				o.Qty, o.UserID, o.Symbol)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		e.logger.Error("failed to commit batch transaction", zap.Error(err))
		return
	}
	e.logger.Info("paper order batch flushed", zap.Int("count", len(batch)))
}

func (e *PaperEngine) PlaceOrder(ctx context.Context, o Order) (int64, error) {
	// Simple validation: check balance if buy
	if o.Side == "buy" {
		var balance decimal.Decimal
		err := e.db.QueryRow(ctx, "SELECT balance FROM paper_accounts WHERE user_id = $1", o.UserID).Scan(&balance)
		if err != nil {
			return 0, fmt.Errorf("failed to get balance: %w", err)
		}
		cost := o.Qty.Mul(o.Price)
		if o.Type == "market" {
			cost = o.Qty.Mul(decimal.Zero) // Actual cost calculated later, but usually we'd estimate
		}
		if balance.LessThan(cost) {
			return 0, errors.New("insufficient balance")
		}
	}

	var id int64
	err := e.db.QueryRow(ctx,
		"INSERT INTO paper_orders (user_id, symbol, side, type, price, qty) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		o.UserID, o.Symbol, o.Side, o.Type, o.Price, o.Qty).Scan(&id)

	if err != nil {
		return 0, err
	}

	o.ID = id
	e.mu.Lock()
	e.orders[o.Symbol] = append(e.orders[o.Symbol], o)
	e.mu.Unlock()

	return id, nil
}
