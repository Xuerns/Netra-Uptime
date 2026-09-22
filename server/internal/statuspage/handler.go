package statuspage

import (
	"PulseWatch/pkg/apierror"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StatusPageHandler struct {
	service StatusPageService
}

func NewStatusPageHandler(service StatusPageService) StatusPageHandler {
	return StatusPageHandler{service: service}
}

func (h *StatusPageHandler) CreateStatusPage(c *gin.Context) {
	var req CreateStatusPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	orgId, exist := c.Get("organizationID")

	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	userId, exist := c.Get("userID")

	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	res, err := h.service.Create(orgId.(uuid.UUID).String(), userId.(string), c.ClientIP(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "message": res})
}

func (h *StatusPageHandler) GetAllStatusPage(c *gin.Context) {
	orgId, exist := c.Get("organizationID")

	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	userId, exist := c.Get("userID")

	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	statusPages, err := h.service.GetAllByOrganizationID(orgId.(uuid.UUID).String(), userId.(string))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": statusPages})
}

func (h *StatusPageHandler) DeleteStatusPage(c *gin.Context) {
	orgId, exist := c.Get("organizationID")

	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	userId, exist := c.Get("userID")

	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	statusPageID := c.Param("statusPageId")

	err := h.service.Delete(statusPageID, orgId.(uuid.UUID).String(), userId.(string), c.ClientIP())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Status page deleted successfully"})
}

func (h *StatusPageHandler) UpdateStatusPage(c *gin.Context) {
	var req UpdateStatusPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	orgId, exist := c.Get("organizationID")

	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	userId, exist := c.Get("userID")

	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	statusPageID := c.Param("statusPageId")

	res, err := h.service.Update(statusPageID, orgId.(uuid.UUID).String(), userId.(string), c.ClientIP(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": res})
}

func (h *StatusPageHandler) GetStatusPageMonitor(c *gin.Context) {
	statusPageID := c.Param("statusPageId")

	statusPageMonitors, err := h.service.GetStatusPageMonitor(statusPageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": statusPageMonitors})
}

func (h *StatusPageHandler) GetById(c *gin.Context) {
	orgId, exist := c.Get("organizationID")
	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	userId, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	statusPageID := c.Param("statusPageId")

	statusPage, err := h.service.GetById(statusPageID, orgId.(uuid.UUID).String(), userId.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": statusPage})
}

func (h *StatusPageHandler) GetPublicBySubdomain(c *gin.Context) {
	subdomain := c.Param("subdomain")

	statusPage, err := h.service.GetBySubdomain(subdomain)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Status page not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": statusPage})
}

func (h *StatusPageHandler) GetPublicSummaryJSON(c *gin.Context) {
	subdomain := c.Param("subdomain")

	statusPage, err := h.service.GetBySubdomain(subdomain)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Status page not found"})
		return
	}

	c.JSON(http.StatusOK, statusPage)
}

func (h *StatusPageHandler) GetPublicStatusJSON(c *gin.Context) {
	subdomain := c.Param("subdomain")

	statusPage, err := h.service.GetBySubdomain(subdomain)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Status page not found"})
		return
	}

	desc := "All Systems Operational"
	switch statusPage.OverallStatus {
	case "degraded":
		desc = "Degraded Performance"
	case "partial_outage":
		desc = "Partial System Outage"
	case "major_outage":
		desc = "Major Service Outage"
	case "maintenance":
		desc = "Under Scheduled Maintenance"
	}

	c.JSON(http.StatusOK, PublicStatusJSONResponse{
		Status:      statusPage.OverallStatus,
		Description: desc,
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *StatusPageHandler) GetStatusBadge(c *gin.Context) {
	subdomain := c.Param("subdomain")

	statusPage, err := h.service.GetBySubdomain(subdomain)

	statusText := "operational"
	bgColor := "#10b981"
	if err != nil {
		statusText = "unknown"
		bgColor = "#6b7280"
	} else {
		switch statusPage.OverallStatus {
		case "operational":
			statusText = "operational"
			bgColor = "#10b981"
		case "degraded":
			statusText = "degraded"
			bgColor = "#f59e0b"
		case "partial_outage":
			statusText = "partial outage"
			bgColor = "#f97316"
		case "major_outage":
			statusText = "major outage"
			bgColor = "#ef4444"
		case "maintenance":
			statusText = "maintenance"
			bgColor = "#3b82f6"
		default:
			statusText = statusPage.OverallStatus
			bgColor = "#10b981"
		}
	}

	svg := generateShieldBadgeSVG("status", statusText, bgColor)
	c.Header("Content-Type", "image/svg+xml")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.String(http.StatusOK, svg)
}

func generateShieldBadgeSVG(label, status, color string) string {
	labelWidth := len(label)*7 + 14
	if labelWidth < 45 {
		labelWidth = 45
	}
	statusWidth := len(status)*7 + 16
	if statusWidth < 50 {
		statusWidth = 50
	}
	totalWidth := labelWidth + statusWidth
	labelCenter := labelWidth / 2
	statusCenter := labelWidth + (statusWidth / 2)

	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20" role="img" aria-label="%s: %s">
	<title>%s: %s</title>
	<linearGradient id="s" x2="0" y2="100%%">
		<stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
		<stop offset="1" stop-opacity=".1"/>
	</linearGradient>
	<clipPath id="r">
		<rect width="%d" height="20" rx="3" fill="#fff"/>
	</clipPath>
	<g clip-path="url(#r)">
		<rect width="%d" height="20" fill="#2d3748"/>
		<rect x="%d" width="%d" height="20" fill="%s"/>
		<rect width="%d" height="20" fill="url(#s)"/>
	</g>
	<g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="11">
		<text aria-hidden="true" x="%d" y="15" fill="#010101" fill-opacity=".3">%s</text>
		<text x="%d" y="14">%s</text>
		<text aria-hidden="true" x="%d" y="15" fill="#010101" fill-opacity=".3">%s</text>
		<text x="%d" y="14">%s</text>
	</g>
</svg>`,
		totalWidth, label, status,
		label, status,
		totalWidth,
		labelWidth,
		labelWidth, statusWidth, color,
		totalWidth,
		labelCenter, label,
		labelCenter, label,
		statusCenter, status,
		statusCenter, status,
	)
}

func (h *StatusPageHandler) GetAllStatusPageByUserId(c *gin.Context) {
	userId, exist := c.Get("userID")

	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	statusPages, err := h.service.GetStatusPageByUserId(userId.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": statusPages})
}

func (h *StatusPageHandler) CreateMaintenance(c *gin.Context) {
	var req CreateMaintenanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	orgId, exist := c.Get("organizationID")
	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	userId, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	statusPageID := c.Param("statusPageId")

	maintenance, err := h.service.CreateMaintenance(orgId.(uuid.UUID).String(), statusPageID, userId.(string), c.ClientIP(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": maintenance})
}

func (h *StatusPageHandler) GetMaintenances(c *gin.Context) {
	orgId, exist := c.Get("organizationID")
	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	userId, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	statusPageID := c.Param("statusPageId")

	maintenances, err := h.service.GetMaintenances(orgId.(uuid.UUID).String(), statusPageID, userId.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": maintenances})
}

func (h *StatusPageHandler) UpdateMaintenance(c *gin.Context) {
	var req UpdateMaintenanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	orgId, exist := c.Get("organizationID")
	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	userId, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	statusPageID := c.Param("statusPageId")
	maintenanceID := c.Param("maintenanceId")

	maintenance, err := h.service.UpdateMaintenance(orgId.(uuid.UUID).String(), statusPageID, maintenanceID, userId.(string), c.ClientIP(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": maintenance})
}

func (h *StatusPageHandler) DeleteMaintenance(c *gin.Context) {
	orgId, exist := c.Get("organizationID")
	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	userId, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	statusPageID := c.Param("statusPageId")
	maintenanceID := c.Param("maintenanceId")

	err := h.service.DeleteMaintenance(orgId.(uuid.UUID).String(), statusPageID, maintenanceID, userId.(string), c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Maintenance deleted successfully"})
}

func (h *StatusPageHandler) CreateComponentGroup(c *gin.Context) {
	var req CreateComponentGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	orgId, exist := c.Get("organizationID")
	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	userId, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	statusPageID := c.Param("statusPageId")

	group, err := h.service.CreateComponentGroup(orgId.(uuid.UUID).String(), statusPageID, userId.(string), c.ClientIP(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": group})
}

func (h *StatusPageHandler) GetComponentGroups(c *gin.Context) {
	orgId, exist := c.Get("organizationID")
	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	userId, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	statusPageID := c.Param("statusPageId")

	groups, err := h.service.GetComponentGroups(orgId.(uuid.UUID).String(), statusPageID, userId.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": groups})
}

func (h *StatusPageHandler) UpdateComponentGroup(c *gin.Context) {
	var req UpdateComponentGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	orgId, exist := c.Get("organizationID")
	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	userId, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	statusPageID := c.Param("statusPageId")
	groupID := c.Param("groupId")

	group, err := h.service.UpdateComponentGroup(orgId.(uuid.UUID).String(), statusPageID, groupID, userId.(string), c.ClientIP(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": group})
}

func (h *StatusPageHandler) DeleteComponentGroup(c *gin.Context) {
	orgId, exist := c.Get("organizationID")
	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	userId, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	statusPageID := c.Param("statusPageId")
	groupID := c.Param("groupId")

	err := h.service.DeleteComponentGroup(orgId.(uuid.UUID).String(), statusPageID, groupID, userId.(string), c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Component group deleted successfully"})
}

func (h *StatusPageHandler) AssignMonitorGroup(c *gin.Context) {
	var req AssignMonitorGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	orgId, exist := c.Get("organizationID")
	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	userId, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	statusPageID := c.Param("statusPageId")
	monitorID := c.Param("monitorId")

	err := h.service.AssignMonitorToGroup(orgId.(uuid.UUID).String(), statusPageID, monitorID, userId.(string), c.ClientIP(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Monitor assigned to group successfully"})
}
