package alerttemplate

import (
	"net/http"
	"netra-uptime/internal/alertrule"
	"netra-uptime/pkg/apierror"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *gin.Context) {
	orgID, ok := getOrgID(c)
	if !ok {
		return
	}

	templates, err := h.service.GetAll(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": templates})
}

func (h *Handler) Create(c *gin.Context) {
	orgID, ok := getOrgID(c)
	if !ok {
		return
	}

	var req CreateAlertTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	template, err := h.service.Create(orgID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": template})
}

func (h *Handler) GetByID(c *gin.Context) {
	orgID, ok := getOrgID(c)
	if !ok {
		return
	}

	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	template, err := h.service.GetByID(orgID, templateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": template})
}

func (h *Handler) Update(c *gin.Context) {
	orgID, ok := getOrgID(c)
	if !ok {
		return
	}

	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var req UpdateAlertTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	template, err := h.service.Update(orgID, templateID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": template})
}

func (h *Handler) Delete(c *gin.Context) {
	orgID, ok := getOrgID(c)
	if !ok {
		return
	}

	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	if err := h.service.Delete(orgID, templateID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Alert template deleted successfully"})
}

func (h *Handler) CreateRule(c *gin.Context) {
	orgID, ok := getOrgID(c)
	if !ok {
		return
	}

	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var req alertrule.CreateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	rule, err := h.service.CreateRule(orgID, templateID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": rule})
}

func (h *Handler) UpdateRule(c *gin.Context) {
	orgID, ok := getOrgID(c)
	if !ok {
		return
	}

	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	ruleID, err := uuid.Parse(c.Param("ruleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}

	var req alertrule.UpdateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	rule, err := h.service.UpdateRule(orgID, templateID, ruleID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": rule})
}

func (h *Handler) DeleteRule(c *gin.Context) {
	orgID, ok := getOrgID(c)
	if !ok {
		return
	}

	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	ruleID, err := uuid.Parse(c.Param("ruleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}

	if err := h.service.DeleteRule(orgID, templateID, ruleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Alert rule deleted from template"})
}

func (h *Handler) AssignToMonitor(c *gin.Context) {
	orgID, ok := getOrgID(c)
	if !ok {
		return
	}

	monitorID, err := uuid.Parse(c.Param("monitorId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor ID"})
		return
	}

	var req ApplyTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	if err := h.service.AssignToMonitor(orgID, monitorID, req.TemplateID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Monitor template updated successfully"})
}

func getOrgID(c *gin.Context) (uuid.UUID, bool) {
	orgIDVal, exists := c.Get("organizationID")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return uuid.Nil, false
	}
	orgID, ok := orgIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return uuid.Nil, false
	}
	return orgID, true
}
