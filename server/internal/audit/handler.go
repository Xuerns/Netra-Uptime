package audit

import (
	"netra-uptime/pkg/apierror"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuditHandler interface {
	GetOrgAuditLogs(c *gin.Context)
}

type AuditHandlerImpl struct {
	service AuditService
}

func NewAuditHandler(service AuditService) AuditHandler {
	return &AuditHandlerImpl{service: service}
}

func (h *AuditHandlerImpl) GetOrgAuditLogs(c *gin.Context) {
	orgId := c.Param("orgId")
	if orgId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization ID is required"})
		return
	}

	var query GetAuditLogsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	resp, err := h.service.GetLogs(orgId, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "audit logs retrieved successfully",
		"data":    resp,
	})
}
