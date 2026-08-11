package organization

import "time"

type CreateOrganizationRequest struct {
	Name        string                 `json:"name" binding:"required,min=3,max=24,nohtml"`
	Slug        string                 `json:"slug" binding:"required,min=6,max=24,printascii,nohtml"`
	Invitations *[]InviteMemberRequest `json:"invitations" binding:"omitempty,dive"`
}

type UpdateOrganizationRequest struct {
	OrganizationID string  `json:"organization_id" binding:"required"`
	Name           *string `json:"name" binding:"omitempty,min=3,max=24,nohtml"`
	Slug           *string `json:"slug" binding:"omitempty,min=6,max=24,printascii,nohtml"`
}

type InvitationActionRequest struct {
	InvitationID string `json:"invitation_id" binding:"required"`
	Action       string `json:"action" binding:"required,oneof=accept reject"`
}

type KickMemberRequest struct {
	MemberID string `json:"member_id" binding:"required"`
}

type InviteMemberRequest struct {
	InviteeID string `json:"invitee_id" binding:"required"`
	Role      string `json:"role" binding:"omitempty,oneof=admin member"`
}

type BulkInviteMemberRequest struct {
	Invitations []InviteMemberRequest `json:"invitations" binding:"required,dive"`
}

type RoleChangeRequest struct {
	UserID string `json:"user_id"`
}

type GetAllOrganizationResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	CreaAt         time.Time `json:"created_at"`
	TotalMonitors  int64     `json:"total_monitors"`
	ActiveMonitors int64     `json:"active_monitors"`
	TotalMember    int64     `json:"total_member"`
}

type SearchResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatar_url"`
}

type MemberResponse struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	UserID         string     `json:"user_id"`
	Role           string     `json:"role"`
	User           MemberUser `json:"user"`
}

type MemberUser struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatar_url"`
}

type InvitationResponse struct {
	ID           string           `json:"id"`
	Role         string           `json:"role"`
	Status       string           `json:"status"`
	Organization OrganizationInfo `json:"organization"`
	Inviter      UserInfo         `json:"inviter"`
}

type OrganizationInfo struct {
	Name string `json:"name"`
}

type UserInfo struct {
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatar_url"`
}
