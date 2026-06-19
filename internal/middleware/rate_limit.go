package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"rbac-platform/internal/httpx"
)

// RateLimit is a fixed-window counter per client IP. Like the rest of the
// Redis-backed middleware, it fails OPEN: a rate limiter that takes the
// whole API down when its own backing store hiccups has made things
// worse, not better.
func RateLimit(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rdb == nil {
			c.Next()
			return
		}
		ctx := c.Request.Context()
		windowID := time.Now().Unix() / int64(window.Seconds())
		key := fmt.Sprintf("ratelimit:%s:%d", c.ClientIP(), windowID)

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}
		if count == 1 {
			rdb.Expire(ctx, key, window)
		}
		if count > int64(limit) {
			httpx.Error(c, http.StatusTooManyRequests, "rate_limited", "too many requests, please try again shortly")
			c.Abort()
			return
		}
		c.Next()
	}
}
