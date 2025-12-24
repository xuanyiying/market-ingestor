package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func (h *Handler) GetPaperAccount(c *gin.Context) {
	userID := c.MustGet("userID").(int64)

	var balance decimal.Decimal
	err := h.db.QueryRow(c.Request.Context(), "SELECT balance FROM paper_accounts WHERE user_id = $1", userID).Scan(&balance)
	if err != nil {
		// Initialize with 100,000 if not exists
		balance = decimal.NewFromFloat(100000.0)
		_, _ = h.db.Exec(c.Request.Context(), "INSERT INTO paper_accounts (user_id, balance) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, balance)
	}

	c.JSON(http.StatusOK, gin.H{"balance": balance})
}

func (h *Handler) CreatePaperOrder(c *gin.Context) {
	userID := c.MustGet("userID").(int64)
	var req struct {
		Symbol string          `json:"symbol" binding:"required"`
		Side   string          `json:"side" binding:"required"` // buy, sell
		Type   string          `json:"type" binding:"required"` // market, limit
		Price  decimal.Decimal `json:"price"`
		Qty    decimal.Decimal `json:"qty" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Risk Check
	if err := h.risk.PreTradeCheck(c.Request.Context(), userID, req.Symbol, req.Side, req.Qty, req.Price); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "risk limit reached: " + err.Error()})
		return
	}

	var orderID int64
	err := h.db.QueryRow(c.Request.Context(),
		"INSERT INTO paper_orders (user_id, symbol, side, type, price, qty) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		userID, strings.ToUpper(req.Symbol), req.Side, req.Type, req.Price, req.Qty).Scan(&orderID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to place paper order"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": orderID})
}

func (h *Handler) GetPaperPositions(c *gin.Context) {
	userID := c.MustGet("userID").(int64)

	rows, err := h.db.Query(c.Request.Context(),
		"SELECT symbol, qty, avg_price FROM paper_positions WHERE user_id = $1 AND qty > 0", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch positions"})
		return
	}
	defer rows.Close()

	positions := make([]map[string]interface{}, 0)
	for rows.Next() {
		var (
			symbol   string
			qty      decimal.Decimal
			avgPrice decimal.Decimal
		)
		if err := rows.Scan(&symbol, &qty, &avgPrice); err != nil {
			continue
		}
		positions = append(positions, map[string]interface{}{
			"symbol":    symbol,
			"qty":       qty,
			"avg_price": avgPrice,
		})
	}

	c.JSON(http.StatusOK, positions)
}
