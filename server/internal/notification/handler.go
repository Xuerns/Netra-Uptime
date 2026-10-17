package notification

import (
	"netra-uptime/pkg/apierror"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChannelHandler struct {
	service ChannelService
}

func NewChannelHandler(service ChannelService) *ChannelHandler {
	return &ChannelHandler{service: service}
}

// ─── POST /organization/:orgId/notifications/channels ───────────────────────
// Role: owner, admin only (enforced at router level)
func (h *ChannelHandler) Create(c *gin.Context) {
	orgID, ok := getOrgID(c)
	if !ok {
		return
	}

	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	resp, err := h.service.Create(orgID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": resp})
}

// ─── GET /organization/:orgId/notifications/channels ────────────────────────
func (h *ChannelHandler) GetAll(c *gin.Context) {
	orgID, ok := getOrgID(c)
	if !ok {
		return
	}

	resp, err := h.service.GetAll(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": resp})
}

// ─── GET /organization/:orgId/notifications/channels/:channelId ─────────────
func (h *ChannelHandler) GetByID(c *gin.Context) {
	orgID, ok := getOrgID(c)
	if !ok {
		return
	}

	channelID := c.Param("channelId")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel ID is required"})
		return
	}

	resp, err := h.service.GetByID(orgID, channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": resp})
}

// ─── PUT /organization/:orgId/notifications/channels/:channelId ─────────────
// Role: owner, admin only (enforced at router level)
func (h *ChannelHandler) Update(c *gin.Context) {
	orgID, ok := getOrgID(c)
	if !ok {
		return
	}

	channelID := c.Param("channelId")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel ID is required"})
		return
	}

	var req UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	resp, err := h.service.Update(orgID, channelID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": resp})
}

// ─── DELETE /organization/:orgId/notifications/channels/:channelId ──────────
// Role: owner, admin only (enforced at router level)
func (h *ChannelHandler) Delete(c *gin.Context) {
	orgID, ok := getOrgID(c)
	if !ok {
		return
	}

	channelID := c.Param("channelId")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel ID is required"})
		return
	}

	if err := h.service.Delete(orgID, channelID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "notification channel deleted"})
}

// ─── GET /organization/:orgId/notifications/logs ────────────────────────────
// Query params: limit (default 50, max 100), offset (default 0)
func (h *ChannelHandler) GetLogs(c *gin.Context) {
	orgID, ok := getOrgID(c)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	resp, err := h.service.GetLogs(orgID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": resp})
}

// ─── Helper ──────────────────────────────────────────────────────────────────

// getOrgID extracts the organization UUID set by TenantMiddleware.
func getOrgID(c *gin.Context) (string, bool) {
	orgID, exists := c.Get("organizationID")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "organization context missing"})
		return "", false
	}
	return orgID.(uuid.UUID).String(), true
}
