package middleware

import (
	"fmt"
	"net/http"
	"netra-uptime/internal/apikey"
	"netra-uptime/pkg/cache"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func APIKeyAuthMiddleware(service apikey.APIKeyService, cacheStore *cache.MemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawKey := c.GetHeader("X-API-Key")
		if rawKey == "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				rawKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if rawKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required in X-API-Key or Authorization Bearer header"})
			c.Abort()
			return
		}

		key, err := service.ValidateKey(rawKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Enforce Rate Limiting (60 req / min per API Key)
		if cacheStore != nil {
			nowWindow := time.Now().Unix() / 60
			rlKey := fmt.Sprintf("rate_limit:api_key:%s:%d", key.ID.String(), nowWindow)

			count := cacheStore.Incr(rlKey, time.Minute)

			limit := int64(60)
			remaining := limit - count
			if remaining < 0 {
				remaining = 0
			}

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", (nowWindow+1)*60))

			if count > limit {
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "API Key rate limit exceeded (60 requests per minute)"})
				c.Abort()
				return
			}
		}

		// Touch last used in background
		go func(keyID string, ip string) {
			_ = service.TouchLastUsed(keyID, ip)
		}(key.ID.String(), c.ClientIP())

		c.Set("org_id", key.OrganizationID.String())
		c.Set("userID", key.CreatedByID.String())
		c.Set("api_key_id", key.ID.String())
		c.Set("api_key_permissions", key.Permissions)
		c.Next()
	}
}

func RequireAPIKeyPermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permVal, exists := c.Get("api_key_permissions")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "API Key permissions not found"})
			c.Abort()
			return
		}
		perm := permVal.(string)

		// 'write' permission includes 'read' capabilities
		if requiredPermission == "write" && perm != "write" {
			c.JSON(http.StatusForbidden, gin.H{"error": "API key requires 'write' permission for this endpoint"})
			c.Abort()
			return
		}

		c.Next()
	}
}
