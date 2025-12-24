package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func (h *Handler) ListMarketStrategies(c *gin.Context) {
	rows, err := h.db.Query(c.Request.Context(),
		`SELECT m.id, m.price, m.description, m.performance_metrics, s.name, u.email as author
		 FROM strategy_market m
		 JOIN strategies s ON m.strategy_id = s.id
		 JOIN users u ON m.owner_id = u.id
		 WHERE m.is_public = TRUE`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch marketplace"})
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var (
			id      int64
			price   decimal.Decimal
			desc    string
			metrics []byte
			sname   string
			author  string
		)
		if err := rows.Scan(&id, &price, &desc, &metrics, &sname, &author); err != nil {
			continue
		}
		items = append(items, map[string]interface{}{
			"id":          id,
			"price":       price,
			"description": desc,
			"metrics":     json.RawMessage(metrics),
			"name":        sname,
			"author":      author,
		})
	}

	c.JSON(http.StatusOK, items)
}

func (h *Handler) PurchaseStrategy(c *gin.Context) {
	userID := c.MustGet("userID").(int64)
	marketItemID := c.Param("id")

	// Verify payment would happen here. For now, we just grant access.
	_, err := h.db.Exec(c.Request.Context(),
		"INSERT INTO strategy_purchases (user_id, market_item_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		userID, marketItemID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to subscribe to strategy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "strategy subscribed successfully"})
}
