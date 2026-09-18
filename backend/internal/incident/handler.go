package incident

import (
	"PulseWatch/pkg/apierror"
	"PulseWatch/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Incidenthandler struct {
	service IncidentService
}

func NewIncidentHandler(service IncidentService) Incidenthandler {
	return Incidenthandler{service: service}
}

func (h *Incidenthandler) GetIncidentDetail(c *gin.Context) {
	orgId, exist := c.Get("organizationID")

	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	incidentId := c.Param("incidentId")

	if incidentId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incidend ID is required"})
		return
	}

	incident, err := h.service.GetIncidentDetail(orgId.(uuid.UUID).String(), incidentId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get incident"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": incident})
}

func (h *Incidenthandler) GetAllIncident(c *gin.Context) {
	orgId, exist := c.Get("organizationID")

	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	incidents, err := h.service.GetAllIncident(orgId.(uuid.UUID).String())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get incident"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": incidents})
}

func (h *Incidenthandler) AddIncidentUpdate(c *gin.Context) {
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

	updatedIncident, err := h.service.AddIncidentUpdate(orgId.(uuid.UUID).String(), incidentId, userId.(string), req.Message)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": updatedIncident})
}

func (h *Incidenthandler) JoinAssignee(c *gin.Context) {
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

	incidentId := c.Param("incidentId")

	if incidentId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incidend ID is required"})
		return
	}

	if err := h.service.JoinAssignee(orgId.(uuid.UUID).String(), incidentId, userId.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "You successfully joined the incident"})
}

func (h *Incidenthandler) GetMessageAndUser(c *gin.Context) {
	orgId, exist := c.Get("organizationID")

	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	incidentId := c.Param("incidentId")

	if incidentId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incidend ID is required"})
		return
	}

	IncidentUpdates, err := h.service.GetIncidentUpdate(orgId.(uuid.UUID).String(), incidentId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": IncidentUpdates})
}
