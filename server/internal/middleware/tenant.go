package middleware

import (
	"net/http"
	"netra-uptime/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// GetOrganizationID extracts the organization UUID from the Gin context safely.
func GetOrganizationID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("organizationID")
	if !exists {
		return uuid.Nil, false
	}
	switch v := val.(type) {
	case uuid.UUID:
		return v, true
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return uuid.Nil, false
		}
		return parsed, true
	default:
		return uuid.Nil, false
	}
}

// GetOrganizationIDString extracts the organization ID as a string safely.
func GetOrganizationIDString(c *gin.Context) (string, bool) {
	id, ok := GetOrganizationID(c)
	if !ok {
		return "", false
	}
	return id.String(), true
}

// GetUserID extracts the user ID string from Gin context safely.
func GetUserID(c *gin.Context) (string, bool) {
	val, exists := c.Get("userID")
	if !exists {
		return "", false
	}
	switch v := val.(type) {
	case string:
		return v, true
	case uuid.UUID:
		return v.String(), true
	default:
		return "", false
	}
}
