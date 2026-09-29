package monitor

import (
	"PulseWatch/pkg/apierror"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MonitorHandler struct {
	service MonitorService
}

func NewMonitorHandler(service MonitorService) MonitorHandler {
	return MonitorHandler{service: service}
}

func (h *MonitorHandler) CreateMonitor(c *gin.Context) {
	userId, exist := c.Get("userID")

	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	orgId, exist := c.Get("organizationID")

	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	var req CreateMonitorRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	monitor, err := h.service.Create(orgId.(uuid.UUID).String(), userId.(string), c.ClientIP(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": monitor})
}

func (h *MonitorHandler) GetById(c *gin.Context) {
	orgId, exist := c.Get("organizationID")

	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	monitorId := c.Param("monitorId")

	if monitorId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Monitor ID is required"})
		return
	}

	monitor, err := h.service.GetById(orgId.(uuid.UUID).String(), monitorId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": monitor})
}

func (h *MonitorHandler) GetAllMonitor(c *gin.Context) {
	orgId, exist := c.Get("organizationID")

	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	monitors, err := h.service.GetAllMonitor(orgId.(uuid.UUID).String())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": monitors})
}

func (h *MonitorHandler) UpdateMonitor(c *gin.Context) {
	userId, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	orgId, exist := c.Get("organizationID")

	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	monitorId := c.Param("monitorId")

	if monitorId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Monitor ID is required"})
		return
	}

	var req UpdatedMonitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	monitor, err := h.service.UpdateMonitor(orgId.(uuid.UUID).String(), userId.(string), c.ClientIP(), monitorId, req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": monitor})
}

func (h *MonitorHandler) DeleteMonitor(c *gin.Context) {
	userId, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	orgId, exist := c.Get("organizationID")

	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	monitorId := c.Param("monitorId")

	if monitorId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Monitor ID is required"})
		return
	}

	if err := h.service.DeleteMonitor(orgId.(uuid.UUID).String(), userId.(string), c.ClientIP(), monitorId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Monitor deleted successfully"})
}

func (h *MonitorHandler) GetAllMonitorByUserId(c *gin.Context) {
	userId, exist := c.Get("userID")

	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	monitors, err := h.service.GetAllMonitorByUserId(userId.(string))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": monitors})
}

func (h *MonitorHandler) GetDailyStats(c *gin.Context) {
	orgId, exist := c.Get("organizationID")
	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	monitorId := c.Param("monitorId")
	if monitorId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Monitor ID is required"})
		return
	}

	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30
	}

	stats, err := h.service.GetDailyStats(orgId.(uuid.UUID).String(), monitorId, days)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": stats})
}

func (h *MonitorHandler) TriggerCheck(c *gin.Context) {
	orgId, exist := c.Get("organizationID")
	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	monitorId := c.Param("monitorId")
	if monitorId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Monitor ID is required"})
		return
	}

	result, err := h.service.TriggerCheck(orgId.(uuid.UUID).String(), monitorId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": result})
}
