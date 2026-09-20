package middleware

import (
	"PulseWatch/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TenantMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exist := c.Get("userID")

		if !exist {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		orgID := c.Param("orgId")
		if orgID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Organization ID required"})
			c.Abort()
			return
		}

		var member model.OrganizationMember
		if err := db.Where("organization_id = ? AND user_id = ?", orgID, userID).First(&member).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Doesn't have access to this organization'"})
			c.Abort()
			return
		}

		c.Set("organizationID", member.OrganizationID)
		c.Set("organizationRole", member.Role)

		c.Next()
	}
}

func RequireRole(allowedRole ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exist := c.Get("organizationRole")
		if !exist {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access forbidden"})
			c.Abort()
			return
		}

		isAllowed := false
		for _, r := range allowedRole {
			if role.(string) == r {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to perform this action"})
			c.Abort()
			return
		}

		c.Next()
	}
}
