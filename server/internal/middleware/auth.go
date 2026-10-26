package middleware

import (
	"net/http"
	"netra-uptime/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("access_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication token missing"})
			c.Abort()
			return
		}

		claims, err := jwt.ValidateToken(tokenString)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		if claims.TokenType != jwt.TokenTypeAccess {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token type for access"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}
