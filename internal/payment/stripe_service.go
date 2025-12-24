package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
	"github.com/stripe/stripe-go/v74/webhook"
	"go.uber.org/zap"
)

type StripeService struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewStripeService(db *pgxpool.Pool, logger *zap.Logger, apiKey string) *StripeService {
	stripe.Key = apiKey
	return &StripeService{
		db:     db,
		logger: logger,
	}
}

func (s *StripeService) CreateCheckoutSession(userID int64, priceID string) (string, error) {
	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String("https://quant-trader.com/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String("https://quant-trader.com/canceled"),
	}
	params.AddMetadata("user_id", fmt.Sprintf("%d", userID))

	sess, err := session.New(params)
	if err != nil {
		return "", err
	}

	return sess.URL, nil
}

func (s *StripeService) HandleWebhook(payload []byte, sigHeader string, endpointSecret string) error {
	event, err := webhook.ConstructEvent(payload, sigHeader, endpointSecret)
	if err != nil {
		s.logger.Error("webhook signature verification failed", zap.Error(err))
		return err
	}

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			return err
		}

		userIDStr := session.Metadata["user_id"]
		userID, _ := strconv.ParseInt(userIDStr, 10, 64)

		// In a real app, we'd map Stripe Price ID to our Tier ID
		// For now, assume everything from Stripe is 'Pro'
		return s.UpdateUserTier(context.Background(), userID, "Pro")
	}

	return nil
}

func (s *StripeService) UpdateUserTier(ctx context.Context, userID int64, tierName string) error {
	// 1. Get tier ID
	var tierID int64
	err := s.db.QueryRow(ctx, "SELECT id FROM subscription_tiers WHERE name = $1", tierName).Scan(&tierID)
	if err != nil {
		return fmt.Errorf("tier not found: %w", err)
	}

	// 2. Upsert subscription
	_, err = s.db.Exec(ctx, `
		INSERT INTO user_subscriptions (user_id, tier_id, status, expires_at)
		VALUES ($1, $2, 'active', NOW() + INTERVAL '30 days')
		ON CONFLICT (user_id) DO UPDATE SET
			tier_id = $2,
			status = 'active',
			expires_at = NOW() + INTERVAL '30 days'`,
		userID, tierID)

	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	s.logger.Info("user tier updated", zap.Int64("user_id", userID), zap.String("tier", tierName))
	return nil
}
