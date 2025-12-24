package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) ListAPIKeys(c *gin.Context) {
	userID := c.MustGet("userID").(int64)

	rows, err := h.db.Query(c.Request.Context(),
		"SELECT id, key_id, name, is_active, created_at FROM api_keys WHERE user_id = $1", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch API keys"})
		return
	}
	defer rows.Close()

	var keys []map[string]interface{}
	for rows.Next() {
		var (
			id       int64
			keyID    string
			name     string
			isActive bool
			cat      time.Time
		)
		if err := rows.Scan(&id, &keyID, &name, &isActive, &cat); err != nil {
			continue
		}
		keys = append(keys, map[string]interface{}{
			"id":         id,
			"key_id":     keyID,
			"name":       name,
			"is_active":  isActive,
			"created_at": cat,
		})
	}

	c.JSON(http.StatusOK, keys)
}

func (h *Handler) CreateAPIKey(c *gin.Context) {
	userID := c.MustGet("userID").(int64)
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	keyID := fmt.Sprintf("qt_%d_%d", userID, time.Now().Unix())
	rawSecret := fmt.Sprintf("sec_%d_%d", userID, time.Now().UnixNano())
	hash, _ := bcrypt.GenerateFromPassword([]byte(rawSecret), bcrypt.DefaultCost)

	_, err := h.db.Exec(c.Request.Context(),
		"INSERT INTO api_keys (user_id, key_id, key_secret, name) VALUES ($1, $2, $3, $4)",
		userID, keyID, string(hash), req.Name)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create API key"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"key_id":     keyID,
		"key_secret": rawSecret,
		"message":    "Store the secret safely, it will not be shown again",
	})
}
