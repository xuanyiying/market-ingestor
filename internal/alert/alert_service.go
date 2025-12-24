package alert

import (
	"context"
	"encoding/json"
	"fmt"
	"market-ingestor/internal/indicators"
	"market-ingestor/internal/model"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type AlertCondition struct {
	ID            int64
	UserID        int64
	Symbol        string
	ConditionType string // "price_above", "price_below"
	TargetValue   decimal.Decimal
}

type AlertService struct {
	db      *pgxpool.Pool
	js      nats.JetStreamContext
	logger  *zap.Logger
	alerts  map[string][]AlertCondition // key: symbol
	candles map[string][]model.KLine    // key: symbol, buffer for indicators
	mu      sync.RWMutex
}

func NewAlertService(db *pgxpool.Pool, js nats.JetStreamContext, logger *zap.Logger) *AlertService {
	return &AlertService{
		db:      db,
		js:      js,
		logger:  logger,
		alerts:  make(map[string][]AlertCondition),
		candles: make(map[string][]model.KLine),
	}
}

func (s *AlertService) Start(ctx context.Context) error {
	// 1. Load active alerts from DB
	if err := s.loadAlerts(ctx); err != nil {
		return err
	}

	// 2. Subscribe to K-line updates
	_, err := s.js.Subscribe("market.kline.1m.*", func(msg *nats.Msg) {
		var candle model.KLine
		if err := json.Unmarshal(msg.Data, &candle); err != nil {
			return
		}
		s.checkAlerts(candle)
	})

	if err != nil {
		return err
	}

	s.logger.Info("alert service started")
	return nil
}

func (s *AlertService) loadAlerts(ctx context.Context) error {
	rows, err := s.db.Query(ctx, "SELECT id, user_id, symbol, condition_type, target_value FROM alerts WHERE is_active = TRUE")
	if err != nil {
		return err
	}
	defer rows.Close()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.alerts = make(map[string][]AlertCondition)

	for rows.Next() {
		var a AlertCondition
		if err := rows.Scan(&a.ID, &a.UserID, &a.Symbol, &a.ConditionType, &a.TargetValue); err != nil {
			continue
		}
		s.alerts[a.Symbol] = append(s.alerts[a.Symbol], a)
	}
	return nil
}

func (s *AlertService) checkAlerts(candle model.KLine) {
	s.mu.Lock()
	// Update candle buffer (keep last 100)
	s.candles[candle.Symbol] = append(s.candles[candle.Symbol], candle)
	if len(s.candles[candle.Symbol]) > 100 {
		s.candles[candle.Symbol] = s.candles[candle.Symbol][1:]
	}
	buffer := s.candles[candle.Symbol]
	s.mu.Unlock()

	s.mu.RLock()
	alerts, ok := s.alerts[candle.Symbol]
	s.mu.RUnlock()

	if !ok {
		return
	}

	for _, a := range alerts {
		triggered := false
		switch a.ConditionType {
		case "price_above":
			if candle.Close.GreaterThanOrEqual(a.TargetValue) {
				triggered = true
			}
		case "price_below":
			if candle.Close.LessThanOrEqual(a.TargetValue) {
				triggered = true
			}
		case "rsi_overbought":
			if len(buffer) > 14 {
				closes := make([]decimal.Decimal, len(buffer))
				for i, c := range buffer {
					closes[i] = c.Close
				}
				rsi := indicators.CalculateRSI(closes, 14)
				if rsi[len(rsi)-1].GreaterThanOrEqual(a.TargetValue) {
					triggered = true
				}
			}
		case "rsi_oversold":
			if len(buffer) > 14 {
				closes := make([]decimal.Decimal, len(buffer))
				for i, c := range buffer {
					closes[i] = c.Close
				}
				rsi := indicators.CalculateRSI(closes, 14)
				if rsi[len(rsi)-1].LessThanOrEqual(a.TargetValue) {
					triggered = true
				}
			}
		}

		if triggered {
			s.triggerAlert(a, candle)
		}
	}
}

func (s *AlertService) triggerAlert(a AlertCondition, candle model.KLine) {
	s.logger.Info("ALERT TRIGGERED",
		zap.Int64("user_id", a.UserID),
		zap.String("symbol", candle.Symbol),
		zap.String("type", a.ConditionType),
		zap.String("price", candle.Close.String()))

	// Publish to a notification topic
	subject := fmt.Sprintf("notification.user.%d", a.UserID)
	msg := map[string]interface{}{
		"type":    "alert",
		"symbol":  candle.Symbol,
		"message": fmt.Sprintf("%s triggered at %s", a.ConditionType, candle.Close.String()),
		"time":    candle.Timestamp,
	}
	data, _ := json.Marshal(msg)
	s.js.Publish(subject, data)

	// Telegram notification (Enterprise feature)
	go s.SendTelegramNotification(a.UserID, fmt.Sprintf("🚨 ALERT: %s %s triggered at %s", candle.Symbol, a.ConditionType, candle.Close.String()))

	// Mark alert as inactive or handle re-trigger logic
	// For now just keep it simple
}

// SendTelegramNotification sends a message via Telegram bot (Mock/Enterprise)
func (s *AlertService) SendTelegramNotification(userID int64, message string) {
	// In a real implementation, we'd fetch the user's telegram_chat_id from DB
	// and use a bot token to call Telegram API.
	s.logger.Info("TELEGRAM NOTIFICATION SENT", zap.Int64("user_id", userID), zap.String("message", message))
}
