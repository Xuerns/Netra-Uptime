package middleware

import (
	"fmt"
	"net/http"
	"netra-uptime/pkg/cache"
	"time"

	"github.com/gin-gonic/gin"
)

func RateLimiter(cacheStore *cache.MemoryCache, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cacheStore == nil {
			c.Next()
			return
		}

		ip := c.ClientIP()
		key := fmt.Sprintf("rate:%s:%s", c.FullPath(), ip)

		currentCount := cacheStore.Incr(key, window)

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
