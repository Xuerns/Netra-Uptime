package publicapi

import (
	"PulseWatch/internal/incident"
	"PulseWatch/internal/monitor"
	"PulseWatch/internal/statuspage"
	"PulseWatch/pkg/apierror"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PublicAPIHandler struct {
	monitorService    monitor.MonitorService
	incidentService   incident.IncidentService
	statusPageService statuspage.StatusPageService
}

func NewPublicAPIHandler(
	monitorService monitor.MonitorService,
	incidentService incident.IncidentService,
	statusPageService statuspage.StatusPageService,
) *PublicAPIHandler {
	return &PublicAPIHandler{
		monitorService:    monitorService,
		incidentService:   incidentService,
		statusPageService: statusPageService,
	}
}

// ── MONITORS ──────────────────────────────────────────────────────────────────

func (h *PublicAPIHandler) ListMonitors(c *gin.Context) {
	orgID := c.GetString("org_id")
	monitors, err := h.monitorService.GetAllMonitor(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   monitors,
	})
}

func (h *PublicAPIHandler) GetMonitor(c *gin.Context) {
	orgID := c.GetString("org_id")
	monitorID := c.Param("id")

	mon, err := h.monitorService.GetById(orgID, monitorID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Monitor not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   mon,
	})
}

func (h *PublicAPIHandler) CreateMonitor(c *gin.Context) {
	orgID := c.GetString("org_id")
	userID := c.GetString("userID")

	var req monitor.CreateMonitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	created, err := h.monitorService.Create(orgID, userID, c.ClientIP(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Monitor created successfully",
		"data":    created,
	})
}

func (h *PublicAPIHandler) UpdateMonitor(c *gin.Context) {
	orgID := c.GetString("org_id")
	userID := c.GetString("userID")
	monitorID := c.Param("id")

	var req monitor.UpdatedMonitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	updated, err := h.monitorService.UpdateMonitor(orgID, userID, c.ClientIP(), monitorID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Monitor updated successfully",
		"data":    updated,
	})
}

func (h *PublicAPIHandler) DeleteMonitor(c *gin.Context) {
	orgID := c.GetString("org_id")
	userID := c.GetString("userID")
	monitorID := c.Param("id")

	if err := h.monitorService.DeleteMonitor(orgID, userID, c.ClientIP(), monitorID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Monitor deleted successfully",
	})
}

// ── INCIDENTS ─────────────────────────────────────────────────────────────────

func (h *PublicAPIHandler) ListIncidents(c *gin.Context) {
	orgID := c.GetString("org_id")

	incidents, err := h.incidentService.GetAllIncident(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   incidents,
	})
}

type CreateIncidentNoteRequest struct {
	IncidentID string `json:"incident_id" binding:"required"`
	Message    string `json:"message" binding:"required,min=2,max=1000"`
}

func (h *PublicAPIHandler) CreateIncidentNote(c *gin.Context) {
	orgID := c.GetString("org_id")
	userID := c.GetString("userID")

	var req CreateIncidentNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	note, err := h.incidentService.AddIncidentUpdate(orgID, req.IncidentID, userID, req.Message)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Incident update note added successfully",
		"data":    note,
	})
}

// ── STATUS PAGES ──────────────────────────────────────────────────────────────

func (h *PublicAPIHandler) ListStatusPages(c *gin.Context) {
	orgID := c.GetString("org_id")

	pages, err := h.statusPageService.GetAllByOrganizationID(orgID, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   pages,
	})
}
