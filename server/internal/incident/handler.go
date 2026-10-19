package incident

import (
	"net/http"
	"netra-uptime/internal/middleware"
	"netra-uptime/pkg/apierror"
	"netra-uptime/pkg/utils"

	"github.com/gin-gonic/gin"
)

type Incidenthandler struct {
	service IncidentService
}

func NewIncidentHandler(service IncidentService) Incidenthandler {
	return Incidenthandler{service: service}
}

func (h *Incidenthandler) GetIncidentDetail(c *gin.Context) {
	orgId, ok := middleware.GetOrganizationIDString(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	incidentId := c.Param("incidentId")

	if incidentId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incidend ID is required"})
		return
	}

	incident, err := h.service.GetIncidentDetail(orgId, incidentId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get incident"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": incident})
}

func (h *Incidenthandler) GetAllIncident(c *gin.Context) {
	orgId, ok := middleware.GetOrganizationIDString(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	monitorId := c.Query("monitor_id")
	incidents, err := h.service.GetAllIncident(orgId, monitorId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get incident"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": incidents})
}

func (h *Incidenthandler) AddIncidentUpdate(c *gin.Context) {
	orgId, ok := middleware.GetOrganizationIDString(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	userId, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	incidentId := c.Param("incidentId")

	if incidentId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incidend ID is required"})
		return
	}

	var req UpdateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	req.Message = utils.SanitizeString(req.Message)

	updatedIncident, err := h.service.AddIncidentUpdate(orgId, incidentId, userId, req.Message)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": updatedIncident})
}

func (h *Incidenthandler) JoinAssignee(c *gin.Context) {
	orgId, ok := middleware.GetOrganizationIDString(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	userId, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	incidentId := c.Param("incidentId")

	if incidentId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incidend ID is required"})
		return
	}

	if err := h.service.JoinAssignee(orgId, incidentId, userId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "You successfully joined the incident"})
}

func (h *Incidenthandler) GetMessageAndUser(c *gin.Context) {
	orgId, ok := middleware.GetOrganizationIDString(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	incidentId := c.Param("incidentId")

	if incidentId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incidend ID is required"})
		return
	}

	IncidentUpdates, err := h.service.GetIncidentUpdate(orgId, incidentId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": IncidentUpdates})
}
