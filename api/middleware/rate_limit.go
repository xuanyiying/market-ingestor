package middleware

import (
	"net/http"
	"time"

	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var (
	limiters  = make(map[int64]*rate.Limiter)
	limiterMu sync.Mutex
)

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get("userID")
		if !ok {
			c.Next()
			return
		}

		uid := userID.(int64)
		tier, _ := c.Get("tier")
		tierName := tier.(string)

		// Define limits: (rate, burst)
		var r rate.Limit
		var b int
		switch tierName {
		case "Enterprise":
			r = rate.Every(time.Minute / 1000)
			b = 1000
		case "Pro":
			r = rate.Every(time.Minute / 100)
			b = 100
		default: // Free
			r = rate.Every(time.Minute / 10)
			b = 10
		}

		limiterMu.Lock()
		limiter, exists := limiters[uid]
		if !exists {
			limiter = rate.NewLimiter(r, b)
			limiters[uid] = limiter
		}
		limiterMu.Unlock()

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded for your tier"})
			c.Abort()
			return
		}

		c.Next()
	}
}
