package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetPortfolios(c *gin.Context) {
	userID := c.MustGet("userID").(int64)

	rows, err := h.db.Query(c.Request.Context(),
		"SELECT id, name, created_at FROM portfolios WHERE user_id = $1", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch portfolios"})
		return
	}
	defer rows.Close()

	var portfolios []map[string]interface{}
	for rows.Next() {
		var (
			id   int64
			name string
			cat  time.Time
		)
		if err := rows.Scan(&id, &name, &cat); err != nil {
			continue
		}
		portfolios = append(portfolios, map[string]interface{}{
			"id":         id,
			"name":       name,
			"created_at": cat,
		})
	}

	c.JSON(http.StatusOK, portfolios)
}

func (h *Handler) CreatePortfolio(c *gin.Context) {
	userID := c.MustGet("userID").(int64)
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var portfolioID int64
	err := h.db.QueryRow(c.Request.Context(),
		"INSERT INTO portfolios (user_id, name) VALUES ($1, $2) RETURNING id",
		userID, req.Name).Scan(&portfolioID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create portfolio"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": portfolioID})
}

func (h *Handler) GetPortfolioReport(c *gin.Context) {
	userID := c.MustGet("userID").(int64)

	report, err := h.analytics.GetPortfolioReport(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate report"})
		return
	}

	c.JSON(http.StatusOK, report)
}
