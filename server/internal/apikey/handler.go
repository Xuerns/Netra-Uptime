package apikey

import (
	"net/http"
	"netra-uptime/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type APIKeyHandler struct {
	service APIKeyService
}

func NewAPIKeyHandler(service APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{service: service}
}

func (h *APIKeyHandler) Create(c *gin.Context) {
	orgID := c.Param("orgId")
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}
	userID := userIDVal.(string)

	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	result, err := h.service.CreateKey(orgID, userID, req, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "API key created successfully",
		"data":    result,
	})
}

func (h *APIKeyHandler) List(c *gin.Context) {
	orgID := c.Param("orgId")

	keys, err := h.service.ListKeys(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   keys,
	})
}

func (h *APIKeyHandler) Revoke(c *gin.Context) {
	orgID := c.Param("orgId")
	keyID := c.Param("keyId")
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}
	userID := userIDVal.(string)

	if err := h.service.RevokeKey(orgID, userID, keyID, c.ClientIP()); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "API key revoked successfully",
	})
}
