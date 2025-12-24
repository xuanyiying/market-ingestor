package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func (h *Handler) GetAlerts(c *gin.Context) {
	userID := c.MustGet("userID").(int64)

	rows, err := h.db.Query(c.Request.Context(),
		"SELECT id, symbol, condition_type, target_value, is_active, created_at FROM alerts WHERE user_id = $1",
		userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch alerts"})
		return
	}
	defer rows.Close()

	alerts := make([]map[string]interface{}, 0)
	for rows.Next() {
		var (
			id            int64
			symbol        string
			conditionType string
			targetValue   decimal.Decimal
			isActive      bool
			createdAt     time.Time
		)
		if err := rows.Scan(&id, &symbol, &conditionType, &targetValue, &isActive, &createdAt); err != nil {
			continue
		}
		alerts = append(alerts, map[string]interface{}{
			"id":             id,
			"symbol":         symbol,
			"condition_type": conditionType,
			"target_value":   targetValue,
			"is_active":      isActive,
			"created_at":     createdAt,
		})
	}

	c.JSON(http.StatusOK, alerts)
}

func (h *Handler) CreateAlert(c *gin.Context) {
	userID := c.MustGet("userID").(int64)
	var req struct {
		Symbol        string          `json:"symbol" binding:"required"`
		ConditionType string          `json:"condition_type" binding:"required"` // price_above, price_below
		TargetValue   decimal.Decimal `json:"target_value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var alertID int64
	err := h.db.QueryRow(c.Request.Context(),
		"INSERT INTO alerts (user_id, symbol, condition_type, target_value) VALUES ($1, $2, $3, $4) RETURNING id",
		userID, strings.ToUpper(req.Symbol), req.ConditionType, req.TargetValue).Scan(&alertID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create alert"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": alertID})
}

func (h *Handler) DeleteAlert(c *gin.Context) {
	userID := c.MustGet("userID").(int64)
	alertID := c.Param("id")

	result, err := h.db.Exec(c.Request.Context(),
		"DELETE FROM alerts WHERE id = $1 AND user_id = $2", alertID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete alert"})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "alert deleted"})
}
