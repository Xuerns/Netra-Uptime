package organization

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"netra-uptime/config"
	"netra-uptime/internal/model"
	"netra-uptime/pkg/utils"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func init() {
	gin.SetMode(gin.TestMode)
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("nohtml", utils.NoHTMLValidator)
	}
}

// Mock Repository
type mockOrgRepo struct {
	getAllSlugOrgFunc         func(slug string) (*[]model.Organization, error)
	getMemberByIdFunc         func(orgId, memberId string) (*model.OrganizationMember, error)
	getOrgMemberFunc          func(orgId, memberId string) error
	getOrgAllMemberFunc       func(orgId string) (*[]model.OrganizationMember, error)
	getAllOrgByUserIdFunc     func(userId string) (*[]model.Organization, error)
	createOrgFunc             func(org *model.Organization, ownerMember *model.OrganizationMember, invitations *[]model.OrganizationInvitation) error
	acceptInvitationFunc      func(invitation *model.OrganizationInvitation, member *model.OrganizationMember) error
	rejectInvitationFunc      func(invitation *model.OrganizationInvitation) error
	kickMemberFunc            func(orgId, memberId string) error
	inviteMemberFunc          func(invitations *[]model.OrganizationInvitation) error
	getInvitationFunc         func(userId string) (*[]model.OrganizationInvitation, error)
	getInvitationByIdFunc     func(invitationId string) (*model.OrganizationInvitation, error)
	disbandFunc               func(orgId, ownerId string) error
	updateOrgFunc             func(updateData *map[string]interface{}, orgId string) error
	searchPersonsFunc         func(name string, excludeUserId string, limit int, offset int, orgId string) ([]model.User, error)
	promoteToAdminFunc        func(userId, orgId string) error
	demotedToMemberFunc       func(userId, orgId string) error
	transferOwnershipFunc     func(ownerId, newOwnerId, orgId string) error
	leaveOrgFunc              func(userId, orgId string) error
	totalOrgUserFunc          func(userId string) (int64, error)
	totalOrgMemberFunc        func(orgId string) (int64, error)
	revokeAPIKeysByMemberFunc func(orgId, memberId string) error
}

func (m *mockOrgRepo) GetAllSlugOrganization(slug string) (*[]model.Organization, error) {
	if m.getAllSlugOrgFunc != nil {
		return m.getAllSlugOrgFunc(slug)
	}
	return nil, nil
}
func (m *mockOrgRepo) GetMemberById(orgId, memberId string) (*model.OrganizationMember, error) {
	if m.getMemberByIdFunc != nil {
		return m.getMemberByIdFunc(orgId, memberId)
	}
	return nil, errors.New("member not found")
}
func (m *mockOrgRepo) GetOrganizationMember(orgId, memberId string) error {
	if m.getOrgMemberFunc != nil {
		return m.getOrgMemberFunc(orgId, memberId)
	}
	return nil
}
func (m *mockOrgRepo) GetOrganizationAllMember(orgId string) (*[]model.OrganizationMember, error) {
	if m.getOrgAllMemberFunc != nil {
		return m.getOrgAllMemberFunc(orgId)
	}
	return nil, nil
}
func (m *mockOrgRepo) GetAllOrganizationByUserId(userId string) (*[]model.Organization, error) {
	if m.getAllOrgByUserIdFunc != nil {
		return m.getAllOrgByUserIdFunc(userId)
	}
	return nil, nil
}
func (m *mockOrgRepo) CreateOrganization(org *model.Organization, ownerMember *model.OrganizationMember, invitations *[]model.OrganizationInvitation) error {
	if m.createOrgFunc != nil {
		return m.createOrgFunc(org, ownerMember, invitations)
	}
	return nil
}
func (m *mockOrgRepo) AcceptInvitation(invitation *model.OrganizationInvitation, member *model.OrganizationMember) error {
	if m.acceptInvitationFunc != nil {
		return m.acceptInvitationFunc(invitation, member)
	}
	return nil
}
func (m *mockOrgRepo) RejectInvitation(invitation *model.OrganizationInvitation) error {
	if m.rejectInvitationFunc != nil {
		return m.rejectInvitationFunc(invitation)
	}
	return nil
}
func (m *mockOrgRepo) KickMember(orgId, memberId string) error {
	if m.kickMemberFunc != nil {
		return m.kickMemberFunc(orgId, memberId)
	}
	return nil
}
func (m *mockOrgRepo) InviteMember(invitations *[]model.OrganizationInvitation) error {
	if m.inviteMemberFunc != nil {
		return m.inviteMemberFunc(invitations)
	}
	return nil
}
func (m *mockOrgRepo) GetInvitation(userId string) (*[]model.OrganizationInvitation, error) {
	if m.getInvitationFunc != nil {
		return m.getInvitationFunc(userId)
	}
	return nil, nil
}
func (m *mockOrgRepo) GetInvitationById(invitationId string) (*model.OrganizationInvitation, error) {
	if m.getInvitationByIdFunc != nil {
		return m.getInvitationByIdFunc(invitationId)
	}
	return nil, errors.New("invitation not found")
}
func (m *mockOrgRepo) Disband(orgId, ownerId string) error {
	if m.disbandFunc != nil {
		return m.disbandFunc(orgId, ownerId)
	}
	return nil
}
func (m *mockOrgRepo) UpdateOrganization(updateData *map[string]interface{}, orgId string) error {
	if m.updateOrgFunc != nil {
		return m.updateOrgFunc(updateData, orgId)
	}
	return nil
}
func (m *mockOrgRepo) SearchPersons(name string, excludeUserId string, limit int, offset int, orgId string) ([]model.User, error) {
	if m.searchPersonsFunc != nil {
		return m.searchPersonsFunc(name, excludeUserId, limit, offset, orgId)
	}
	return nil, nil
}
func (m *mockOrgRepo) PromoteToAdmin(userId, orgId string) error {
	if m.promoteToAdminFunc != nil {
		return m.promoteToAdminFunc(userId, orgId)
	}
	return nil
}
func (m *mockOrgRepo) DemotedToMember(userId, orgId string) error {
	if m.demotedToMemberFunc != nil {
		return m.demotedToMemberFunc(userId, orgId)
	}
	return nil
}
func (m *mockOrgRepo) TransferOwnership(ownerId, newOwnerId, orgId string) error {
	if m.transferOwnershipFunc != nil {
		return m.transferOwnershipFunc(ownerId, newOwnerId, orgId)
	}
	return nil
}
func (m *mockOrgRepo) LeaveOrganization(userId, orgId string) error {
	if m.leaveOrgFunc != nil {
		return m.leaveOrgFunc(userId, orgId)
	}
	return nil
}
func (m *mockOrgRepo) TotalOrganizationUser(userId string) (int64, error) {
	if m.totalOrgUserFunc != nil {
		return m.totalOrgUserFunc(userId)
	}
	return 0, nil
}
func (m *mockOrgRepo) TotalOrganizationMember(orgId string) (int64, error) {
	if m.totalOrgMemberFunc != nil {
		return m.totalOrgMemberFunc(orgId)
	}
	return 0, nil
}
func (m *mockOrgRepo) RevokeAPIKeysByMember(orgId, memberId string) error {
	if m.revokeAPIKeysByMemberFunc != nil {
		return m.revokeAPIKeysByMemberFunc(orgId, memberId)
	}
	return nil
}

// ====================== SERVICE TESTS ======================

func TestGetAllOrganization_CountActiveMonitors(t *testing.T) {
	repo := &mockOrgRepo{
		getAllOrgByUserIdFunc: func(userId string) (*[]model.Organization, error) {
			return &[]model.Organization{
				{
					ID:   uuid.New(),
					Name: "Acme Corp",
					Slug: "acme-corp",
					Monitors: []model.Monitor{
						{Status: "up"},
						{Status: "degraded"},
						{Status: "down"},
					},
					OrganizationMembers: []model.OrganizationMember{{}},
				},
			}, nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	res, err := svc.GetAllOrganization(uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(*res) != 1 {
		t.Fatalf("expected 1 org, got %d", len(*res))
	}
	// "up" and "degraded" are != "down", so 2 active monitors
	if (*res)[0].ActiveMonitors != 2 {
		t.Errorf("expected 2 active monitors, got %d", (*res)[0].ActiveMonitors)
	}
	if (*res)[0].TotalMonitors != 3 {
		t.Errorf("expected 3 total monitors, got %d", (*res)[0].TotalMonitors)
	}
}

func TestCreateOrganization_Success(t *testing.T) {
	created := false
	repo := &mockOrgRepo{
		totalOrgUserFunc: func(userId string) (int64, error) {
			return 1, nil
		},
		getAllSlugOrgFunc: func(slug string) (*[]model.Organization, error) {
			return nil, nil
		},
		createOrgFunc: func(org *model.Organization, owner *model.OrganizationMember, inv *[]model.OrganizationInvitation) error {
			created = true
			return nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	msg, err := svc.CreateOrganization(&CreateOrganizationRequest{
		Name: "New Org",
		Slug: "new-org-123",
	}, uuid.New().String())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Fatal("expected CreateOrganization to be called")
	}
	if msg != "Organization created successfully" {
		t.Errorf("unexpected message: %s", msg)
	}
}

func TestCreateOrganization_SlugAlreadyTaken(t *testing.T) {
	repo := &mockOrgRepo{
		totalOrgUserFunc: func(userId string) (int64, error) {
			return 0, nil
		},
		getAllSlugOrgFunc: func(slug string) (*[]model.Organization, error) {
			return &[]model.Organization{{Slug: slug}}, nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.CreateOrganization(&CreateOrganizationRequest{
		Name: "Duplicate",
		Slug: "duplicate-slug",
	}, uuid.New().String())

	if err == nil {
		t.Fatal("expected error for duplicate slug, got nil")
	}
}

func TestCreateOrganization_MaxOrgsReached(t *testing.T) {
	// Case 1: Configured MaxOrgsPerUser
	repo := &mockOrgRepo{
		totalOrgUserFunc: func(userId string) (int64, error) {
			return 5, nil
		},
	}
	svc := NewOrganizationService(repo, nil, &config.Config{MaxOrgsPerUser: 5})

	_, err := svc.CreateOrganization(&CreateOrganizationRequest{
		Name: "Limit Org",
		Slug: "limit-org",
	}, uuid.New().String())

	if err == nil || err.Error() != "You have reach max organization limit" {
		t.Fatalf("expected error 'You have reach max organization limit', got: %v", err)
	}

	// Case 2: Default MaxOrgsPerUser fallback (10)
	repoDefault := &mockOrgRepo{
		totalOrgUserFunc: func(userId string) (int64, error) {
			return 10, nil
		},
	}
	svcDefault := NewOrganizationService(repoDefault, nil, &config.Config{MaxOrgsPerUser: 0})
	_, err = svcDefault.CreateOrganization(&CreateOrganizationRequest{
		Name: "Limit Org 2",
		Slug: "limit-org-2",
	}, uuid.New().String())

	if err == nil || err.Error() != "You have reach max organization limit" {
		t.Fatalf("expected error on fallback 10 orgs, got: %v", err)
	}
}

type mockWsEvicter struct {
	evictedOrgID  string
	evictedUserID string
}

func (m *mockWsEvicter) EvictUserFromOrg(orgID, userID string) {
	m.evictedOrgID = orgID
	m.evictedUserID = userID
}

type mockAPIKeyRevoker struct {
	revokedOrgID  string
	revokedUserID string
}

func (m *mockAPIKeyRevoker) RevokeKeysByOrgAndUser(orgID, userID string) error {
	m.revokedOrgID = orgID
	m.revokedUserID = userID
	return nil
}

func TestKickMember_Success(t *testing.T) {
	kicked := false
	repoRevoked := false
	repo := &mockOrgRepo{
		getMemberByIdFunc: func(orgId, memberId string) (*model.OrganizationMember, error) {
			return &model.OrganizationMember{}, nil
		},
		kickMemberFunc: func(orgId, memberId string) error {
			kicked = true
			return nil
		},
		revokeAPIKeysByMemberFunc: func(orgId, memberId string) error {
			repoRevoked = true
			return nil
		},
	}
	svc := NewOrganizationService(repo, nil)
	wsEvicter := &mockWsEvicter{}
	apiKeyRevoker := &mockAPIKeyRevoker{}
	svc.SetWebSocketEvicter(wsEvicter)
	svc.SetAPIKeyRevoker(apiKeyRevoker)

	_, err := svc.KickMember("org-1", "actor-1", "member-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !kicked {
		t.Fatal("expected KickMember to be called")
	}
	if !repoRevoked {
		t.Fatal("expected repo.RevokeAPIKeysByMember to be called")
	}
	if apiKeyRevoker.revokedOrgID != "org-1" || apiKeyRevoker.revokedUserID != "member-1" {
		t.Fatalf("expected apiKeyRevoker to be called with org-1, member-1, got %s, %s", apiKeyRevoker.revokedOrgID, apiKeyRevoker.revokedUserID)
	}
	if wsEvicter.evictedOrgID != "org-1" || wsEvicter.evictedUserID != "member-1" {
		t.Fatalf("expected wsEvicter to be called with org-1, member-1, got %s, %s", wsEvicter.evictedOrgID, wsEvicter.evictedUserID)
	}
}

func TestKickMember_MemberNotFound(t *testing.T) {
	repo := &mockOrgRepo{
		getMemberByIdFunc: func(orgId, memberId string) (*model.OrganizationMember, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.KickMember("org-1", "actor-1", "member-1", "127.0.0.1")
	if err == nil {
		t.Fatal("expected error when member not found, got nil")
	}
}

func TestInvitationAction_Accept_Success(t *testing.T) {
	accepted := false
	inviteeID := uuid.New()
	orgID := uuid.New()
	repo := &mockOrgRepo{
		getInvitationByIdFunc: func(id string) (*model.OrganizationInvitation, error) {
			return &model.OrganizationInvitation{
				InviteeID:      inviteeID,
				OrganizationID: orgID,
				ExpiresAt:      time.Now().Add(24 * time.Hour),
				Role:           "member",
				Status:         "pending",
			}, nil
		},
		totalOrgUserFunc: func(userId string) (int64, error) {
			return 1, nil
		},
		totalOrgMemberFunc: func(orgId string) (int64, error) {
			return 5, nil
		},
		acceptInvitationFunc: func(inv *model.OrganizationInvitation, member *model.OrganizationMember) error {
			accepted = true
			return nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.InvitationAction(&InvitationActionRequest{
		InvitationID: "inv-1",
		Action:       "accept",
	}, inviteeID.String(), "127.0.0.1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !accepted {
		t.Fatal("expected invitation to be accepted")
	}
}

func TestInvitationAction_Reject_Success(t *testing.T) {
	rejected := false
	inviteeID := uuid.New()
	repo := &mockOrgRepo{
		getInvitationByIdFunc: func(id string) (*model.OrganizationInvitation, error) {
			return &model.OrganizationInvitation{
				InviteeID: inviteeID,
				ExpiresAt: time.Now().Add(24 * time.Hour),
				Status:    "pending",
			}, nil
		},
		rejectInvitationFunc: func(inv *model.OrganizationInvitation) error {
			rejected = true
			return nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.InvitationAction(&InvitationActionRequest{
		InvitationID: "inv-1",
		Action:       "reject",
	}, inviteeID.String(), "127.0.0.1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rejected {
		t.Fatal("expected invitation to be rejected")
	}
}

func TestInvitationAction_Expired(t *testing.T) {
	inviteeID := uuid.New()
	repo := &mockOrgRepo{
		getInvitationByIdFunc: func(id string) (*model.OrganizationInvitation, error) {
			return &model.OrganizationInvitation{
				InviteeID: inviteeID,
				ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
				Status:    "pending",
			}, nil
		},
		rejectInvitationFunc: func(inv *model.OrganizationInvitation) error {
			return nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.InvitationAction(&InvitationActionRequest{
		InvitationID: "inv-1",
		Action:       "accept",
	}, inviteeID.String(), "127.0.0.1")

	if err == nil {
		t.Fatal("expected error for expired invitation, got nil")
	}
}

func TestInvitationAction_NotYourInvitation(t *testing.T) {
	repo := &mockOrgRepo{
		getInvitationByIdFunc: func(id string) (*model.OrganizationInvitation, error) {
			return &model.OrganizationInvitation{
				InviteeID: uuid.New(), // different user
				ExpiresAt: time.Now().Add(24 * time.Hour),
				Status:    "pending",
			}, nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.InvitationAction(&InvitationActionRequest{
		InvitationID: "inv-1",
		Action:       "accept",
	}, uuid.New().String(), "127.0.0.1")

	if err == nil {
		t.Fatal("expected unauthorized error for non-matching user, got nil")
	}
}

func TestInvitationAction_AlreadyAccepted_ReplayBlocked(t *testing.T) {
	inviteeID := uuid.New()
	repo := &mockOrgRepo{
		getInvitationByIdFunc: func(id string) (*model.OrganizationInvitation, error) {
			return &model.OrganizationInvitation{
				InviteeID: inviteeID,
				ExpiresAt: time.Now().Add(24 * time.Hour),
				Status:    "accepted", // Already processed
			}, nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.InvitationAction(&InvitationActionRequest{
		InvitationID: "inv-1",
		Action:       "accept",
	}, inviteeID.String(), "127.0.0.1")

	if err == nil {
		t.Fatal("expected error when replaying already accepted invitation, got nil")
	}
}

func TestInvitationAction_InvalidAction(t *testing.T) {
	inviteeID := uuid.New()
	repo := &mockOrgRepo{
		getInvitationByIdFunc: func(id string) (*model.OrganizationInvitation, error) {
			return &model.OrganizationInvitation{
				InviteeID: inviteeID,
				ExpiresAt: time.Now().Add(24 * time.Hour),
				Status:    "pending",
			}, nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.InvitationAction(&InvitationActionRequest{
		InvitationID: "inv-1",
		Action:       "unknown-action",
	}, inviteeID.String(), "127.0.0.1")

	if err == nil {
		t.Fatal("expected error for invalid action, got nil")
	}
}

func TestInviteMember_AlreadyMember(t *testing.T) {
	orgID := uuid.New()
	inviteeID := uuid.New()
	repo := &mockOrgRepo{
		totalOrgMemberFunc: func(orgId string) (int64, error) {
			return 2, nil
		},
		getMemberByIdFunc: func(o, m string) (*model.OrganizationMember, error) {
			return &model.OrganizationMember{}, nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.InviteMember(orgID.String(), uuid.New().String(), "127.0.0.1", BulkInviteMemberRequest{
		Invitations: []InviteMemberRequest{{InviteeID: inviteeID.String(), Role: "member"}},
	})

	if err == nil {
		t.Fatal("expected error when inviting existing member, got nil")
	}
}

func TestInviteMember_PendingInviteExists(t *testing.T) {
	orgID := uuid.New()
	inviteeID := uuid.New()
	repo := &mockOrgRepo{
		totalOrgMemberFunc: func(orgId string) (int64, error) {
			return 2, nil
		},
		getMemberByIdFunc: func(o, m string) (*model.OrganizationMember, error) {
			return nil, errors.New("not found")
		},
		getInvitationFunc: func(userId string) (*[]model.OrganizationInvitation, error) {
			return &[]model.OrganizationInvitation{
				{OrganizationID: orgID, InviteeID: inviteeID, Status: "pending"},
			}, nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.InviteMember(orgID.String(), uuid.New().String(), "127.0.0.1", BulkInviteMemberRequest{
		Invitations: []InviteMemberRequest{{InviteeID: inviteeID.String()}},
	})

	if err == nil {
		t.Fatal("expected error when pending invite exists, got nil")
	}
}

func TestInviteMember_ExceedsMemberLimit(t *testing.T) {
	orgID := uuid.New()

	// Case 1: Configured MaxMembersPerOrg
	repo := &mockOrgRepo{
		totalOrgMemberFunc: func(orgId string) (int64, error) {
			return 10, nil
		},
	}
	svc := NewOrganizationService(repo, nil, &config.Config{MaxMembersPerOrg: 10})

	_, err := svc.InviteMember(orgID.String(), uuid.New().String(), "127.0.0.1", BulkInviteMemberRequest{
		Invitations: []InviteMemberRequest{{InviteeID: uuid.New().String()}},
	})

	if err == nil || err.Error() != "Exceeded total organization member limit (max 10)" {
		t.Fatalf("expected 'Exceeded total organization member limit (max 10)', got: %v", err)
	}

	// Case 2: Default MaxMembersPerOrg fallback (100)
	repoDefault := &mockOrgRepo{
		totalOrgMemberFunc: func(orgId string) (int64, error) {
			return 100, nil
		},
	}
	svcDefault := NewOrganizationService(repoDefault, nil, &config.Config{MaxMembersPerOrg: 0})
	_, err = svcDefault.InviteMember(orgID.String(), uuid.New().String(), "127.0.0.1", BulkInviteMemberRequest{
		Invitations: []InviteMemberRequest{{InviteeID: uuid.New().String()}},
	})
	if err == nil || err.Error() != "Exceeded total organization member limit (max 100)" {
		t.Fatalf("expected 'Exceeded total organization member limit (max 100)', got: %v", err)
	}
}

func TestInviteMember_AdminRoleSuccess(t *testing.T) {
	orgID := uuid.New()
	inviteeID := uuid.New()
	ownerID := uuid.New()
	var capturedInvitations *[]model.OrganizationInvitation

	repo := &mockOrgRepo{
		totalOrgMemberFunc: func(orgId string) (int64, error) {
			return 2, nil
		},
		getMemberByIdFunc: func(o, m string) (*model.OrganizationMember, error) {
			if m == ownerID.String() {
				return &model.OrganizationMember{Role: "owner"}, nil
			}
			return nil, errors.New("not found")
		},
		getInvitationFunc: func(userId string) (*[]model.OrganizationInvitation, error) {
			return &[]model.OrganizationInvitation{}, nil
		},
		inviteMemberFunc: func(invitations *[]model.OrganizationInvitation) error {
			capturedInvitations = invitations
			return nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.InviteMember(orgID.String(), ownerID.String(), "127.0.0.1", BulkInviteMemberRequest{
		Invitations: []InviteMemberRequest{{InviteeID: inviteeID.String(), Role: "admin"}},
	})

	if err != nil {
		t.Fatalf("unexpected error inviting member: %v", err)
	}
	if capturedInvitations == nil || len(*capturedInvitations) == 0 {
		t.Fatal("expected invitations to be saved, got nil or empty")
	}
	if (*capturedInvitations)[0].Role != "admin" {
		t.Fatalf("expected invitation role to be 'admin', got '%s'", (*capturedInvitations)[0].Role)
	}
}

func TestInviteMember_NonOwnerCannotInviteAdmin(t *testing.T) {
	orgID := uuid.New()
	inviteeID := uuid.New()
	adminID := uuid.New()

	repo := &mockOrgRepo{
		totalOrgMemberFunc: func(orgId string) (int64, error) {
			return 2, nil
		},
		getMemberByIdFunc: func(o, m string) (*model.OrganizationMember, error) {
			if m == adminID.String() {
				return &model.OrganizationMember{Role: "admin"}, nil
			}
			return nil, errors.New("not found")
		},
		getInvitationFunc: func(userId string) (*[]model.OrganizationInvitation, error) {
			return &[]model.OrganizationInvitation{}, nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.InviteMember(orgID.String(), adminID.String(), "127.0.0.1", BulkInviteMemberRequest{
		Invitations: []InviteMemberRequest{{InviteeID: inviteeID.String(), Role: "admin"}},
	})

	if err == nil {
		t.Fatal("expected error when admin tries to invite another admin, got nil")
	}
}

func TestPromoteToAdmin_AlreadyAdmin(t *testing.T) {
	repo := &mockOrgRepo{
		getMemberByIdFunc: func(orgId, memberId string) (*model.OrganizationMember, error) {
			return &model.OrganizationMember{Role: "admin"}, nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.PromoteToAdmin("user-1", "org-1", "actor-1", "127.0.0.1")
	if err == nil {
		t.Fatal("expected error when promoting existing admin, got nil")
	}
}

func TestPromoteToAdmin_OwnerBlocked(t *testing.T) {
	repo := &mockOrgRepo{
		getMemberByIdFunc: func(orgId, memberId string) (*model.OrganizationMember, error) {
			return &model.OrganizationMember{Role: "owner"}, nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.PromoteToAdmin("user-1", "org-1", "actor-1", "127.0.0.1")
	if err == nil {
		t.Fatal("expected error when promoting owner, got nil")
	}
}

func TestDemotedToMember_AlreadyMember(t *testing.T) {
	repo := &mockOrgRepo{
		getMemberByIdFunc: func(orgId, memberId string) (*model.OrganizationMember, error) {
			return &model.OrganizationMember{Role: "member"}, nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.DemotedToMember("user-1", "org-1", "actor-1", "127.0.0.1")
	if err == nil {
		t.Fatal("expected error when demoting existing member, got nil")
	}
}

func TestDemotedToMember_Success_RevokesAPIKey(t *testing.T) {
	demoted := false
	repoRevoked := false
	repo := &mockOrgRepo{
		getMemberByIdFunc: func(orgId, memberId string) (*model.OrganizationMember, error) {
			return &model.OrganizationMember{Role: "admin"}, nil
		},
		demotedToMemberFunc: func(userId, orgId string) error {
			demoted = true
			return nil
		},
		revokeAPIKeysByMemberFunc: func(orgId, memberId string) error {
			repoRevoked = true
			return nil
		},
	}
	svc := NewOrganizationService(repo, nil)
	apiKeyRevoker := &mockAPIKeyRevoker{}
	svc.SetAPIKeyRevoker(apiKeyRevoker)

	_, err := svc.DemotedToMember("user-1", "org-1", "actor-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !demoted {
		t.Fatal("expected DemotedToMember to be called")
	}
	if !repoRevoked {
		t.Fatal("expected repo.RevokeAPIKeysByMember to be called")
	}
	if apiKeyRevoker.revokedOrgID != "org-1" || apiKeyRevoker.revokedUserID != "user-1" {
		t.Fatalf("expected apiKeyRevoker to be called with org-1, user-1, got %s, %s", apiKeyRevoker.revokedOrgID, apiKeyRevoker.revokedUserID)
	}
}

func TestTransferOwnership_NotOwner(t *testing.T) {
	targetID := uuid.New()
	callerID := uuid.New()
	repo := &mockOrgRepo{
		getMemberByIdFunc: func(orgId, memberId string) (*model.OrganizationMember, error) {
			if memberId == targetID.String() {
				return &model.OrganizationMember{UserID: targetID, Role: "member"}, nil
			}
			return &model.OrganizationMember{UserID: callerID, Role: "admin"}, nil // not owner!
		},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.TransferOwnership(callerID.String(), targetID.String(), "org-1", "127.0.0.1")
	if err == nil {
		t.Fatal("expected error when non-owner transfers ownership, got nil")
	}
}

func TestLeaveOrganization_OwnerBlocked(t *testing.T) {
	repo := &mockOrgRepo{
		getMemberByIdFunc: func(orgId, memberId string) (*model.OrganizationMember, error) {
			return &model.OrganizationMember{Role: "owner"}, nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.LeaveOrganization("user-1", "org-1", "127.0.0.1")
	if err == nil {
		t.Fatal("expected error when owner attempts to leave, got nil")
	}
}

func TestLeaveOrganization_Success_EvictsWebSocketAndRevokesAPIKey(t *testing.T) {
	left := false
	repoRevoked := false
	repo := &mockOrgRepo{
		getMemberByIdFunc: func(orgId, memberId string) (*model.OrganizationMember, error) {
			return &model.OrganizationMember{Role: "admin"}, nil
		},
		leaveOrgFunc: func(userId, orgId string) error {
			left = true
			return nil
		},
		revokeAPIKeysByMemberFunc: func(orgId, memberId string) error {
			repoRevoked = true
			return nil
		},
	}
	svc := NewOrganizationService(repo, nil)
	wsEvicter := &mockWsEvicter{}
	apiKeyRevoker := &mockAPIKeyRevoker{}
	svc.SetWebSocketEvicter(wsEvicter)
	svc.SetAPIKeyRevoker(apiKeyRevoker)

	_, err := svc.LeaveOrganization("user-1", "org-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !left {
		t.Fatal("expected LeaveOrganization to be called")
	}
	if !repoRevoked {
		t.Fatal("expected repo.RevokeAPIKeysByMember to be called")
	}
	if apiKeyRevoker.revokedOrgID != "org-1" || apiKeyRevoker.revokedUserID != "user-1" {
		t.Fatalf("expected apiKeyRevoker to be called with org-1, user-1, got %s, %s", apiKeyRevoker.revokedOrgID, apiKeyRevoker.revokedUserID)
	}
	if wsEvicter.evictedOrgID != "org-1" || wsEvicter.evictedUserID != "user-1" {
		t.Fatalf("expected wsEvicter to be called with org-1, user-1, got %s, %s", wsEvicter.evictedOrgID, wsEvicter.evictedUserID)
	}
}

func TestGetOrganizationAllMember(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	repo := &mockOrgRepo{
		getOrgMemberFunc: func(o, m string) error {
			return nil
		},
		getOrgAllMemberFunc: func(o string) (*[]model.OrganizationMember, error) {
			return &[]model.OrganizationMember{
				{ID: uuid.New(), OrganizationID: orgID, UserID: userID, Role: "owner", User: model.User{Name: "Boss"}},
			}, nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	members, err := svc.GetOrganizationAllMember(orgID.String(), userID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(*members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(*members))
	}
}

func TestGetInvitation_Success(t *testing.T) {
	userID := uuid.New()
	repo := &mockOrgRepo{
		getInvitationFunc: func(u string) (*[]model.OrganizationInvitation, error) {
			return &[]model.OrganizationInvitation{
				{
					ID:           uuid.New(),
					Role:         "member",
					Status:       "pending",
					Organization: model.Organization{Name: "Pulse"},
					Inviter:      model.User{Name: "Inviter"},
				},
			}, nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	invs, err := svc.GetInvitation(userID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(*invs) != 1 {
		t.Fatalf("expected 1 invitation, got %d", len(*invs))
	}
}

func TestDisband_Success(t *testing.T) {
	disbanded := false
	repo := &mockOrgRepo{
		disbandFunc: func(o, u string) error {
			disbanded = true
			return nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.Disband("org-1", "user-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !disbanded {
		t.Fatal("expected Disband to be called")
	}
}

func TestUpdateOrganization_Success(t *testing.T) {
	updated := false
	repo := &mockOrgRepo{
		updateOrgFunc: func(data *map[string]interface{}, o string) error {
			updated = true
			return nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	name := "Renamed Org"
	_, err := svc.UpdateOrganization(&UpdateOrganizationRequest{Name: &name}, "org-1", "user-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated {
		t.Fatal("expected Update to be called")
	}
}

func TestSearchPersons_Success(t *testing.T) {
	repo := &mockOrgRepo{
		searchPersonsFunc: func(name, exclude string, limit, offset int, orgId string) ([]model.User, error) {
			return []model.User{{ID: uuid.New(), Name: "Alice"}}, nil
		},
	}
	svc := NewOrganizationService(repo, nil)

	users, err := svc.SearchPersons("Ali", "user-1", 1, "org-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
}

// ====================== HANDLER TESTS ======================

func TestOrganizationHandler_Create(t *testing.T) {
	repo := &mockOrgRepo{
		totalOrgUserFunc: func(userId string) (int64, error) {
			return 0, nil
		},
		getAllSlugOrgFunc: func(slug string) (*[]model.Organization, error) {
			return nil, nil
		},
		createOrgFunc: func(org *model.Organization, owner *model.OrganizationMember, inv *[]model.OrganizationInvitation) error {
			return nil
		},
	}
	svc := NewOrganizationService(repo, nil)
	handler := NewOrganizationHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		c.Next()
	})
	r.POST("/organization", handler.CreateOrganization)

	body := CreateOrganizationRequest{
		Name: "Super Team",
		Slug: "super-team",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/organization", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestOrganizationHandler_GetAllOrganization(t *testing.T) {
	repo := &mockOrgRepo{
		getAllOrgByUserIdFunc: func(userId string) (*[]model.Organization, error) {
			return &[]model.Organization{}, nil
		},
	}
	svc := NewOrganizationService(repo, nil)
	handler := NewOrganizationHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		c.Next()
	})
	r.GET("/organization/all", handler.GetAllOrganization)

	req, _ := http.NewRequest(http.MethodGet, "/organization/all", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestOrganizationHandler_GetAllMember(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	repo := &mockOrgRepo{
		getOrgMemberFunc: func(o, m string) error { return nil },
		getOrgAllMemberFunc: func(o string) (*[]model.OrganizationMember, error) {
			return &[]model.OrganizationMember{}, nil
		},
	}
	svc := NewOrganizationService(repo, nil)
	handler := NewOrganizationHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("organizationID", orgID)
		c.Set("userID", userID.String())
		c.Next()
	})
	r.GET("/organization/:orgId/members", handler.GetAllMember)

	req, _ := http.NewRequest(http.MethodGet, "/organization/"+orgID.String()+"/members", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
