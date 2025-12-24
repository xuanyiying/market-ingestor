package api

import (
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetSubscription(c *gin.Context) {
	userID := c.MustGet("userID").(int64)

	var sub struct {
		TierName   string    `json:"tier_name"`
		MaxSymbols int       `json:"max_symbols"`
		Status     string    `json:"status"`
		ExpiresAt  time.Time `json:"expires_at"`
	}

	err := h.db.QueryRow(c.Request.Context(),
		`SELECT t.name, t.max_symbols, s.status, s.expires_at 
		 FROM user_subscriptions s 
		 JOIN subscription_tiers t ON s.tier_id = t.id 
		 WHERE s.user_id = $1`, userID).Scan(&sub.TierName, &sub.MaxSymbols, &sub.Status, &sub.ExpiresAt)

	if err != nil {
		// Return default Free tier info if no record
		c.JSON(http.StatusOK, gin.H{
			"tier_name":   "Free",
			"max_symbols": 1,
			"status":      "active",
			"expires_at":  time.Now().AddDate(99, 0, 0),
		})
		return
	}

	c.JSON(http.StatusOK, sub)
}

func (h *Handler) CreateCheckoutSession(c *gin.Context) {
	userID := c.MustGet("userID").(int64)
	var req struct {
		PriceID string `json:"price_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	url, err := h.stripe.CreateCheckoutSession(userID, req.PriceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create checkout session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *Handler) HandleStripeWebhook(c *gin.Context) {
	payload, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	sigHeader := c.GetHeader("Stripe-Signature")
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	err = h.stripe.HandleWebhook(payload, sigHeader, endpointSecret)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	c.Status(http.StatusOK)
}
