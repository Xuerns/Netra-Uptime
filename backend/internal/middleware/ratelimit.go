package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimiter(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		ip := c.ClientIP()
		key := fmt.Sprintf("rate:%s:%s", c.FullPath(), ip)

		pipe := rdb.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window)

		_, err := pipe.Exec(ctx)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server error"})
			return
		}

		currentCount := incr.Val()

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))

		remaining := int64(limit) - currentCount
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if currentCount > int64(limit) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "To many request, please try again later"})
			return
		}

		c.Next()
	}
}
