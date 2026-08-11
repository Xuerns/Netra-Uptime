package organization

import (
	"PulseWatch/config"
	"PulseWatch/internal/audit"
	"PulseWatch/internal/model"
	"errors"
	"fmt"

	"time"

	"github.com/google/uuid"
)

type OrganizationService interface {
	GetAllOrganization(userId string) (*[]GetAllOrganizationResponse, error)
	GetOrganizationAllMember(orgId, memberId string) (*[]MemberResponse, error)
	CreateOrganization(req *CreateOrganizationRequest, ownerMemberId string) (string, error)
	GetInvitation(userId string) (*[]InvitationResponse, error)
	KickMember(orgId, actorId, memberId, ipAddress string) (string, error)
	InvitationAction(req *InvitationActionRequest, userId, ipAddress string) (string, error)
	InviteMember(orgId, ownerId, ipAddress string, invitations BulkInviteMemberRequest) (string, error)
	Disband(orgId, ownerId, ipAddress string) (string, error)
	UpdateOrganization(req *UpdateOrganizationRequest, orgId, actorId, ipAddress string) (string, error)
	SearchPersons(name string, excludeUserId string, page int, orgId string) ([]SearchResponse, error)
	PromoteToAdmin(userId, orgId, actorId, ipAddress string) (string, error)
	DemotedToMember(userId, orgId, actorId, ipAddress string) (string, error)
	TransferOwnership(ownerId, userId, orgId, ipAddress string) (string, error)
	LeaveOrganization(userId, orgId, ipAddress string) (string, error)
	GetOrganizationPlan(orgId string) (*model.Subscription, error)
}

type OrganizationServiceImpl struct {
	repo         OrganizationRepository
	auditService audit.AuditService
}

func NewOrganizationService(repo OrganizationRepository, auditService audit.AuditService) OrganizationService {
	return &OrganizationServiceImpl{repo: repo, auditService: auditService}
}

func (s *OrganizationServiceImpl) GetAllOrganization(userId string) (*[]GetAllOrganizationResponse, error) {
	organizations, err := s.repo.GetAllOrganizationByUserId(userId)
	if err != nil {
		return nil, err
	}

	var response []GetAllOrganizationResponse
	for _, organization := range *organizations {
		var activeMonitor int64
		for _, monitor := range organization.Monitors {
			if monitor.Status != "down" {
				activeMonitor++
			}
		}

		response = append(response, GetAllOrganizationResponse{
			ID:             organization.ID.String(),
			Name:           organization.Name,
			Slug:           organization.Slug,
			CreaAt:         organization.CreatedAt,
			TotalMonitors:  int64(len(organization.Monitors)),
			ActiveMonitors: activeMonitor,
			TotalMember:    int64(len(organization.OrganizationMembers)),
		})
	}
	return &response, nil
}

func (s *OrganizationServiceImpl) GetOrganizationAllMember(orgId, memberId string) (*[]MemberResponse, error) {
	if err := s.repo.GetOrganizationMember(orgId, memberId); err != nil {
		return nil, errors.New("You are not member of this organization")
	}

	members, err := s.repo.GetOrganizationAllMember(orgId)
	if err != nil {
		return nil, err
	}

	var response []MemberResponse
	for _, member := range *members {
		response = append(response, MemberResponse{
			ID:             member.ID.String(),
			OrganizationID: member.OrganizationID.String(),
			UserID:         member.UserID.String(),
			Role:           member.Role,
			User: MemberUser{
				ID:        member.UserID.String(),
				Name:      member.User.Name,
				AvatarURL: &member.User.AvatarURL,
			},
		})
	}

	return &response, nil
}

func (s *OrganizationServiceImpl) CreateOrganization(req *CreateOrganizationRequest, ownerMemberId string) (string, error) {

	totalOrganizationHave, err := s.repo.TotalOrganizationUser(ownerMemberId)

	if err != nil {
		return "", err
	}

	maxOrgs := 7
	maxMembers := 20
	cfg := config.Load()
	if cfg.AppMode == "selfhosted" {
		if cfg.SelfHostedMaxOrgs > 0 {
			maxOrgs = cfg.SelfHostedMaxOrgs
		} else {
			maxOrgs = 999999
		}
		if cfg.SelfHostedMaxMembers > 0 {
			maxMembers = cfg.SelfHostedMaxMembers
		} else {
			maxMembers = 999999
		}
	}

	if totalOrganizationHave >= int64(maxOrgs) {
		return "", errors.New("You have reach max organization limit")
	}

	if req.Invitations != nil && len(*req.Invitations) >= maxMembers {
		return "", fmt.Errorf("Max organization member is %d", maxMembers)
	}

	exist, err := s.repo.GetAllSlugOrganization(req.Slug)

	if err != nil {
		return "", err
	}

	if exist != nil {
		return "", errors.New("Slug has taken or already exist, please try another slug")
	}

	organization := model.Organization{
		Name: req.Name,
		Slug: req.Slug,
	}

	owner := model.OrganizationMember{
		UserID: uuid.MustParse(ownerMemberId),
		Role:   "owner",
	}

	var invitations []model.OrganizationInvitation
	if req.Invitations != nil {
		for _, member := range *req.Invitations {
			invitations = append(invitations, model.OrganizationInvitation{
				InviteeID: uuid.MustParse(member.InviteeID),
				Status:    "pending",
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			})
		}
	}

	hobbySubs := model.Subscription{
		Plan:   "hobby",
		Status: "active",
	}

	if err := s.repo.CreateOrganization(&organization, &owner, &invitations, &hobbySubs); err != nil {
		return "", err
	}

	return "Organization created successfully", nil
}

func (s *OrganizationServiceImpl) GetInvitation(userId string) (*[]InvitationResponse, error) {
	invitations, err := s.repo.GetInvitation(userId)
	if err != nil {
		return nil, err
	}

	var response []InvitationResponse
	for _, inv := range *invitations {
		response = append(response, InvitationResponse{
			ID:     inv.ID.String(),
			Role:   inv.Role,
			Status: inv.Status,
			Organization: OrganizationInfo{
				Name: inv.Organization.Name,
			},
			Inviter: UserInfo{
				Name:      inv.Inviter.Name,
				AvatarURL: &inv.Inviter.AvatarURL,
			},
		})
	}

	return &response, nil
}

func (s *OrganizationServiceImpl) KickMember(orgId, actorId, memberId, ipAddress string) (string, error) {
	member, err := s.repo.GetMemberById(orgId, memberId)

	if err != nil {
		return "", err
	}

	if member == nil {
		return "", errors.New("Member not found")
	}

	if err := s.repo.KickMember(orgId, memberId); err != nil {
		return "", err
	}

	if s.auditService != nil {
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgId,
			UserID:         actorId,
			Action:         "ORG_MEMBER_KICK",
			EntityType:     "organization_member",
			EntityID:       &memberId,
			Description:    fmt.Sprintf("Kicked member ID %s from organization", memberId),
			IPAddress:      ipAddress,
		})
	}

	return "Member kicked successfully", nil
}

func (s *OrganizationServiceImpl) InvitationAction(req *InvitationActionRequest, userId, ipAddress string) (string, error) {
	invitation, err := s.repo.GetInvitationById(req.InvitationID)

	if err != nil {
		return "", err
	}

	if invitation.InviteeID.String() != userId {
		return "", errors.New("Unauthorized: this is not your invitation")
	}

	if invitation.ExpiresAt.Before(time.Now()) {
		invitation.Status = "expired"
		if err := s.repo.RejectInvitation(invitation); err != nil {
			return "", err
		}
		return "", errors.New("Invitation expired")
	}

	switch req.Action {
	case "accept":
		totalOrganizationHave, err := s.repo.TotalOrganizationUser(userId)
		if err != nil {
			return "", err
		}
		maxOrgs := 7
		maxMembers := 20
		cfg := config.Load()
		if cfg.AppMode == "selfhosted" {
			if cfg.SelfHostedMaxOrgs > 0 {
				maxOrgs = cfg.SelfHostedMaxOrgs
			} else {
				maxOrgs = 999999
			}
			if cfg.SelfHostedMaxMembers > 0 {
				maxMembers = cfg.SelfHostedMaxMembers
			} else {
				maxMembers = 999999
			}
		}

		if totalOrganizationHave >= int64(maxOrgs) {
			return "", errors.New("You have reach max organization limit")
		}

		totalCurrentMember, err := s.repo.TotalOrganizationMember(invitation.OrganizationID.String())
		if err != nil {
			return "", err
		}

		if totalCurrentMember >= int64(maxMembers) {
			return "", fmt.Errorf("Organization has reached its max member limit (%d)", maxMembers)
		}

		invitation.Status = "accepted"
		var member = model.OrganizationMember{
			OrganizationID: invitation.OrganizationID,
			UserID:         invitation.InviteeID,
			Role:           invitation.Role,
		}

		if err := s.repo.AcceptInvitation(invitation, &member); err != nil {
			return "", err
		}

		if s.auditService != nil {
			orgIDStr := invitation.OrganizationID.String()
			_ = s.auditService.Record(audit.RecordAuditLogParams{
				OrganizationID: orgIDStr,
				UserID:         userId,
				Action:         "ORG_MEMBER_JOIN",
				EntityType:     "organization_member",
				Description:    "Accepted invitation and joined organization",
				IPAddress:      ipAddress,
			})
		}

		return "Invitation accepted successfully", nil

	case "reject":
		invitation.Status = "rejected"
		if err := s.repo.RejectInvitation(invitation); err != nil {
			return "", err
		}

		if s.auditService != nil {
			orgIDStr := invitation.OrganizationID.String()
			_ = s.auditService.Record(audit.RecordAuditLogParams{
				OrganizationID: orgIDStr,
				UserID:         userId,
				Action:         "ORG_INVITE_REJECT",
				EntityType:     "organization_invitation",
				Description:    "Rejected organization invitation",
				IPAddress:      ipAddress,
			})
		}

		return "Invitation rejected", nil
	default:
		return "", errors.New("Invalid action")
	}
}

func (s *OrganizationServiceImpl) InviteMember(orgId, ownerId, ipAddress string, req BulkInviteMemberRequest) (string, error) {

	var dbInvitations []model.OrganizationInvitation

	totalCurrentMember, err := s.repo.TotalOrganizationMember(orgId)

	if err != nil {
		return "", err
	}

	maxMembers := 20
	cfg := config.Load()
	if cfg.AppMode == "selfhosted" {
		if cfg.SelfHostedMaxMembers > 0 {
			maxMembers = cfg.SelfHostedMaxMembers
		} else {
			maxMembers = 999999
		}
	}

	if totalCurrentMember+int64(len(req.Invitations)) > int64(maxMembers) {
		return "", fmt.Errorf("Exceeded total organization member limit (max %d)", maxMembers)
	}

	for _, invite := range req.Invitations {

		member, err := s.repo.GetMemberById(orgId, invite.InviteeID)

		if err == nil && member != nil {
			return "", errors.New("User is already a member of this organization")
		}

		pendingInvites, err := s.repo.GetInvitation(invite.InviteeID)

		if err == nil && pendingInvites != nil {
			for _, pending := range *pendingInvites {
				if pending.OrganizationID.String() == orgId {
					return "", errors.New("User already has a pending invitation for this organization")
				}
			}
		}

		role := invite.Role
		if role == "" {
			role = "member"
		}
		dbInvitations = append(dbInvitations, model.OrganizationInvitation{
			OrganizationID: uuid.MustParse(orgId),
			InviteeID:      uuid.MustParse(invite.InviteeID),
			Status:         "pending",
			ExpiresAt:      time.Now().Add(7 * 24 * time.Hour),
			InviterID:      uuid.MustParse(ownerId),
		})
	}

	if len(dbInvitations) == 0 {
		return "", errors.New("No invitations found")
	}

	if err := s.repo.InviteMember(&dbInvitations); err != nil {
		return "", err
	}

	if s.auditService != nil {
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgId,
			UserID:         ownerId,
			Action:         "ORG_MEMBER_INVITE",
			EntityType:     "organization_invitation",
			Description:    fmt.Sprintf("Sent %d member invitation(s)", len(dbInvitations)),
			IPAddress:      ipAddress,
		})
	}

	return "Invitations sent successfully", nil
}

func (s *OrganizationServiceImpl) Disband(orgId, ownerId, ipAddress string) (string, error) {

	if s.auditService != nil {
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgId,
			UserID:         ownerId,
			Action:         "ORG_DISBAND",
			EntityType:     "organization",
			Description:    "Disbanded organization",
			IPAddress:      ipAddress,
		})
	}

	if err := s.repo.Disband(orgId, ownerId); err != nil {
		return "", err
	}

	return "Organization disbanded", nil
}

func (s *OrganizationServiceImpl) UpdateOrganization(req *UpdateOrganizationRequest, orgId, actorId, ipAddress string) (string, error) {

	updateData := make(map[string]interface{})
	if req.Name != nil {
		updateData["name"] = *req.Name
	}
	if req.Slug != nil {
		updateData["slug"] = *req.Slug
	}
	if len(updateData) > 0 {
		if err := s.repo.UpdateOrganization(&updateData, orgId); err != nil {
			return "", err
		}
	}

	if s.auditService != nil {
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgId,
			UserID:         actorId,
			Action:         "ORG_UPDATE",
			EntityType:     "organization",
			Description:    "Updated organization settings",
			IPAddress:      ipAddress,
		})
	}

	return "Organization updated successfully", nil
}

func (s *OrganizationServiceImpl) SearchPersons(name string, excludeUserId string, page int, orgId string) ([]SearchResponse, error) {
	limit := 10
	offset := (page - 1) * limit
	users, err := s.repo.SearchPersons(name, excludeUserId, limit, offset, orgId)
	if err != nil {
		return nil, err
	}

	if users == nil {
		return nil, errors.New("No users found")
	}

	var response []SearchResponse
	for _, user := range users {
		response = append(response, SearchResponse{
			ID:        user.ID.String(),
			Name:      user.Name,
			AvatarURL: &user.AvatarURL,
		})
	}

	return response, nil
}

func (s *OrganizationServiceImpl) PromoteToAdmin(userId, orgId, actorId, ipAddress string) (string, error) {
	member, err := s.repo.GetMemberById(orgId, userId)
	if err != nil {
		return "", errors.New("You are not member of this organization")
	}

	if member.Role == "owner" {
		return "", errors.New("You can't demoted your self")
	}

	if member.Role == "admin" {
		return "", errors.New("You're already admin")
	}

	if err := s.repo.PromoteToAdmin(userId, orgId); err != nil {
		return "", err
	}

	if s.auditService != nil {
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgId,
			UserID:         actorId,
			Action:         "ORG_MEMBER_PROMOTE",
			EntityType:     "organization_member",
			EntityID:       &userId,
			Description:    fmt.Sprintf("Promoted user ID %s to Admin", userId),
			IPAddress:      ipAddress,
		})
	}

	return "User promoted to admin successfully", nil
}

func (s *OrganizationServiceImpl) DemotedToMember(userId, orgId, actorId, ipAddress string) (string, error) {
	member, err := s.repo.GetMemberById(orgId, userId)
	if err != nil {
		return "", errors.New("You are not member of this organization")
	}

	if member.Role == "owner" {
		return "", errors.New("You can't demoted your self")
	}

	if member.Role == "member" {
		return "", errors.New("You're already member")
	}

	if err := s.repo.DemotedToMember(userId, orgId); err != nil {
		return "", err
	}

	if s.auditService != nil {
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgId,
			UserID:         actorId,
			Action:         "ORG_MEMBER_DEMOTE",
			EntityType:     "organization_member",
			EntityID:       &userId,
			Description:    fmt.Sprintf("Demoted user ID %s to Member", userId),
			IPAddress:      ipAddress,
		})
	}

	return "User demoted to member successfully", nil
}

func (s *OrganizationServiceImpl) TransferOwnership(ownerId, userId, orgId, ipAddress string) (string, error) {
	member, err := s.repo.GetMemberById(orgId, userId)
	if err != nil {
		return "", errors.New("Target user is not a member of this organization")
	}

	owner, err := s.repo.GetMemberById(orgId, ownerId)
	if err != nil {
		return "", errors.New("You are not a member of this organization")
	}

	if owner.Role != "owner" {
		return "", errors.New("You are not owner of this organization")
	}

	if err := s.repo.TransferOwnership(owner.UserID.String(), member.UserID.String(), orgId); err != nil {
		return "", err
	}

	if s.auditService != nil {
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgId,
			UserID:         ownerId,
			Action:         "ORG_TRANSFER_OWNERSHIP",
			EntityType:     "organization",
			EntityID:       &userId,
			Description:    fmt.Sprintf("Transferred ownership to user ID %s", userId),
			IPAddress:      ipAddress,
		})
	}

	return "Ownership transferred successfully", nil
}

func (s *OrganizationServiceImpl) LeaveOrganization(userId, orgId, ipAddress string) (string, error) {
	member, err := s.repo.GetMemberById(orgId, userId)
	if err != nil {
		return "", errors.New("You are not member of this organization")
	}

	if member.Role == "owner" {
		return "", errors.New("You can't leave your own organization")
	}

	if err := s.repo.LeaveOrganization(userId, orgId); err != nil {
		return "", err
	}

	if s.auditService != nil {
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgId,
			UserID:         userId,
			Action:         "ORG_MEMBER_LEAVE",
			EntityType:     "organization_member",
			Description:    "Left organization",
			IPAddress:      ipAddress,
		})
	}

	return "User left organization successfully", nil
}

func (s *OrganizationServiceImpl) GetOrganizationPlan(orgId string) (*model.Subscription, error) {
	return s.repo.GetOrganizationPlan(orgId)
}
