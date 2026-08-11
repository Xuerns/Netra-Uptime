package statuspage

import (
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orgId, exist := c.Get("organizationID")

	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Something wrong, please try again"})
		return
	}

	userId, exist := c.Get("userID")

	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Something wrong, please try again"})
		return
	}

	res, err := h.service.Create(orgId.(uuid.UUID).String(), userId.(string), c.ClientIP(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": res})
}

func (h *StatusPageHandler) GetAllStatusPage(c *gin.Context) {
	orgId, exist := c.Get("organizationID")

	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Something wrong, please try again"})
		return
	}

	userId, exist := c.Get("userID")

	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Something wrong, please try again"})
		return
	}

	statusPages, err := h.service.GetAllByOrganizationID(orgId.(uuid.UUID).String(), userId.(string))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": statusPages})
}

func (h *StatusPageHandler) DeleteStatusPage(c *gin.Context) {
	orgId, exist := c.Get("organizationID")

	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Something wrong, please try again"})
		return
	}

	userId, exist := c.Get("userID")

	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Something wrong, please try again"})
		return
	}

	statusPageID := c.Param("statusPageId")

	err := h.service.Delete(statusPageID, orgId.(uuid.UUID).String(), userId.(string), c.ClientIP())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Status page deleted successfully"})
}

func (h *StatusPageHandler) UpdateStatusPage(c *gin.Context) {
	var req UpdateStatusPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orgId, exist := c.Get("organizationID")

	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Something wrong, please try again"})
		return
	}

	userId, exist := c.Get("userID")

	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Something wrong, please try again"})
		return
	}

	statusPageID := c.Param("statusPageId")

	res, err := h.service.Update(statusPageID, orgId.(uuid.UUID).String(), userId.(string), c.ClientIP(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": res})
}

func (h *StatusPageHandler) GetStatusPageMonitor(c *gin.Context) {
	statusPageID := c.Param("statusPageId")

	statusPageMonitors, err := h.service.GetStatusPageMonitor(statusPageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": statusPageMonitors})
}

func (h *StatusPageHandler) GetById(c *gin.Context) {
	orgId, exist := c.Get("organizationID")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Something wrong, please try again"})
		return
	}

	userId, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Something wrong, please try again"})
		return
	}

	statusPageID := c.Param("statusPageId")

	statusPage, err := h.service.GetById(statusPageID, orgId.(uuid.UUID).String(), userId.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	statusPages, err := h.service.GetStatusPageByUserId(userId.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": statusPages})
}
