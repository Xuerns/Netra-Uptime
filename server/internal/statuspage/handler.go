package statuspage

import (
	"PulseWatch/pkg/apierror"
	"net/http"

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
