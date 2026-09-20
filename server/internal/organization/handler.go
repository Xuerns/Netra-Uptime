package organization

import (
	"PulseWatch/pkg/apierror"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrganizationHandler struct {
	service OrganizationService
}

func NewOrganizationHandler(service OrganizationService) OrganizationHandler {
	return OrganizationHandler{service: service}
}

func (h OrganizationHandler) GetAllOrganization(c *gin.Context) {
	userId, exist := c.Get("userID")

	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	orgs, err := h.service.GetAllOrganization(userId.(string))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": orgs})
}

func (h *OrganizationHandler) GetAllMember(c *gin.Context) {
	orgId := c.Param("orgId")

	memberId, exist := c.Get("userID")

	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	members, err := h.service.GetOrganizationAllMember(orgId, memberId.(string))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": members})
}

func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	var req CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	ownerId, exist := c.Get("userID")

	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	res, err := h.service.CreateOrganization(&req, ownerId.(string))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "message": res})
}

func (h *OrganizationHandler) GetInvitation(c *gin.Context) {
	userId, exist := c.Get("userID")

	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	invitations, err := h.service.GetInvitation(userId.(string))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": invitations})
}

func (h *OrganizationHandler) InvitationAction(c *gin.Context) {
	userId, exist := c.Get("userID")

	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	var req InvitationActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	res, err := h.service.InvitationAction(&req, userId.(string), c.ClientIP())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": res})
}

func (h *OrganizationHandler) InviteMember(c *gin.Context) {

	var req BulkInviteMemberRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

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

	res, err := h.service.InviteMember(orgId.(uuid.UUID).String(), userId.(string), c.ClientIP(), req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": res})
}

func (h *OrganizationHandler) KickMember(c *gin.Context) {
	userId, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	var memberId string
	if err := c.ShouldBindJSON(&memberId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	orgId, exist := c.Get("organizationID")

	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	res, err := h.service.KickMember(orgId.(uuid.UUID).String(), userId.(string), memberId, c.ClientIP())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": res})
}

func (h *OrganizationHandler) Disband(c *gin.Context) {
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

	res, err := h.service.Disband(orgId.(uuid.UUID).String(), userId.(string), c.ClientIP())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": res})

}

func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	userId, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		c.Abort()
		return
	}

	orgId, exist := c.Get("organizationID")

	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		c.Abort()
		return
	}

	var req UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		c.Abort()
		return
	}

	res, err := h.service.UpdateOrganization(&req, orgId.(uuid.UUID).String(), userId.(string), c.ClientIP())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": res})
}

func (h *OrganizationHandler) SearchPersons(c *gin.Context) {

	userId, exist := c.Get("userID")

	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	name := c.Query("name")
	pageStr := c.Query("page")
	orgid := c.Query("orgId")

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query 'name' is required"})
		c.Abort()
		return
	}

	users, err := h.service.SearchPersons(name, userId.(string), page, orgid)
	if err != nil {
		if err.Error() == "No users found" {
			c.JSON(http.StatusOK, gin.H{"status": "success", "data": []interface{}{}})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		}
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": users})
}

func (h *OrganizationHandler) PromoteToAdmin(c *gin.Context) {
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

	var req RoleChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	res, err := h.service.PromoteToAdmin(req.UserID, orgId.(uuid.UUID).String(), userId.(string), c.ClientIP())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": res})
}

func (h *OrganizationHandler) DemoteToMember(c *gin.Context) {
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

	var req RoleChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	res, err := h.service.DemotedToMember(req.UserID, orgId.(uuid.UUID).String(), userId.(string), c.ClientIP())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": res})
}

func (h *OrganizationHandler) TransferOwnership(c *gin.Context) {
	owner, exist := c.Get("userID")

	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
		return
	}

	orgId, exist := c.Get("organizationID")

	if !exist {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization context missing"})
		return
	}

	var req RoleChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apierror.FormatValidationError(err)})
		return
	}

	res, err := h.service.TransferOwnership(owner.(string), req.UserID, orgId.(uuid.UUID).String(), c.ClientIP())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": res})
}

func (h *OrganizationHandler) LeaveOrganization(c *gin.Context) {
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

	res, err := h.service.LeaveOrganization(userId.(string), orgId.(uuid.UUID).String(), c.ClientIP())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierror.SanitizeServiceError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": res})
}