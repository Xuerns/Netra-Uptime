package subscription

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	service SubscriptionService
}

func NewSubscriptionHandler(service SubscriptionService) SubscriptionHandler {
	return SubscriptionHandler{service: service}
}

func (h *SubscriptionHandler) MidtransWebHookHandler(c *gin.Context) {
	var notification MidtransNotification

	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.service.ProcessSubscriptionWebhook(&notification); err != nil {
		if err.Error() == "Invalid signature" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something wrong, please try again later"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook processed successfully"})
}

func (h *SubscriptionHandler) CreateSubcriptionTransaction(c *gin.Context) {
	_, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	orgId := c.Param("orgId")
	if orgId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization ID is required"})
		return
	}

	var req CreateSubscriptionTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	resp, err := h.service.CreateSubscriptionTransaction(orgId, req.Plan, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something wrong, please try again later"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": resp,
	})
}

func (h *SubscriptionHandler) GetSubByUser(c *gin.Context) {
	userId, exist := c.Get("userID")

	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	data, err := h.service.GetAllSubsByUserId(userId.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Something wrong, please try again later",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": data,
	})
}
