package statuspage

import (
	"PulseWatch/internal/model"
	"PulseWatch/pkg/utils"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func init() {
	gin.SetMode(gin.TestMode)
	os.Setenv("APP_MODE", "saas")
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("nohtml", utils.NoHTMLValidator)
	}
}

// Mock Repository
type mockStatusPageRepo struct {
	createFunc                      func(statuspage *model.StatusPage) (*model.StatusPage, error)
	getAllByOrganizationIDFunc     func(orgID string) ([]model.StatusPage, error)
	deleteFunc                      func(statusPageID, orgID string) error
	updateFunc                      func(statuspageID, orgID string, statuspage *model.StatusPage) (*model.StatusPage, error)
	getByIdFunc                     func(statuspageID, orgID string) (*model.StatusPage, error)
	getBySubdomainFunc              func(subdomain string) (*model.StatusPage, error)
	getStatusPageMonitorFunc        func(statusPageID string) ([]model.StatusPageMonitor, error)
	getStatusPageByUserIdFunc       func(userId string) ([]model.StatusPage, error)
	totalOrganizationStatusPageFunc func(orgId string) (int64, error)
}

func (m *mockStatusPageRepo) Create(statuspage *model.StatusPage) (*model.StatusPage, error) {
	if m.createFunc != nil {
		return m.createFunc(statuspage)
	}
	statuspage.ID = uuid.New()
	statuspage.CreatedAt = time.Now()
	return statuspage, nil
}
func (m *mockStatusPageRepo) GetAllByOrganizationID(orgID string) ([]model.StatusPage, error) {
	if m.getAllByOrganizationIDFunc != nil {
		return m.getAllByOrganizationIDFunc(orgID)
	}
	return nil, nil
}
func (m *mockStatusPageRepo) Delete(statusPageID, orgID string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(statusPageID, orgID)
	}
	return nil
}
func (m *mockStatusPageRepo) Update(statuspageID, orgID string, statuspage *model.StatusPage) (*model.StatusPage, error) {
	if m.updateFunc != nil {
		return m.updateFunc(statuspageID, orgID, statuspage)
	}
	return statuspage, nil
}
func (m *mockStatusPageRepo) GetById(statuspageID, orgID string) (*model.StatusPage, error) {
	if m.getByIdFunc != nil {
		return m.getByIdFunc(statuspageID, orgID)
	}
	return nil, errors.New("not found")
}
func (m *mockStatusPageRepo) GetBySubdomain(subdomain string) (*model.StatusPage, error) {
	if m.getBySubdomainFunc != nil {
		return m.getBySubdomainFunc(subdomain)
	}
	return nil, errors.New("status page not found")
}
func (m *mockStatusPageRepo) GetStatusPageMonitor(statusPageID string) ([]model.StatusPageMonitor, error) {
	if m.getStatusPageMonitorFunc != nil {
		return m.getStatusPageMonitorFunc(statusPageID)
	}
	return nil, nil
}
func (m *mockStatusPageRepo) GetStatusPageByUserId(userId string) ([]model.StatusPage, error) {
	if m.getStatusPageByUserIdFunc != nil {
		return m.getStatusPageByUserIdFunc(userId)
	}
	return nil, nil
}
func (m *mockStatusPageRepo) TotalOrganizationStatusPage(orgId string) (int64, error) {
	if m.totalOrganizationStatusPageFunc != nil {
		return m.totalOrganizationStatusPageFunc(orgId)
	}
	return 0, nil
}

// Mock SubscriptionChecker
type mockStatusPageChecker struct {
	plan      string
	member    *model.OrganizationMember
	notMember bool
}

func (m *mockStatusPageChecker) GetOrganizationPlan(orgId string) (*model.Subscription, error) {
	plan := m.plan
	if plan == "" {
		plan = "hobby"
	}
	return &model.Subscription{Plan: plan, Status: "active"}, nil
}

func (m *mockStatusPageChecker) GetMemberById(orgId, userId string) (*model.OrganizationMember, error) {
	if m.notMember {
		return nil, nil
	}
	if m.member != nil {
		return m.member, nil
	}
	return &model.OrganizationMember{Role: "member"}, nil
}

// ====================== SERVICE TESTS ======================

func TestCreate_Success(t *testing.T) {
	repo := &mockStatusPageRepo{}
	checker := &mockStatusPageChecker{plan: "hobby"}
	svc := NewStatusPageService(repo, checker, nil, nil)

	orgID := uuid.New().String()
	userID := uuid.New().String()
	monID := uuid.New().String()

	desc1 := "Status of all services"
	sp, err := svc.Create(orgID, userID, "127.0.0.1", &CreateStatusPageRequest{
		Title:       "Public Status",
		Description: &desc1,
		Subdomain:   "status-corp",
		IsPublic:    true,
		MonitorID:   []string{monID},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sp.Title != "Public Status" {
		t.Errorf("expected title Public Status, got %s", sp.Title)
	}
}

func TestCreate_ExceedsLimit(t *testing.T) {
	repo := &mockStatusPageRepo{
		totalOrganizationStatusPageFunc: func(orgId string) (int64, error) {
			return 1, nil // Hobby allows max 1 status page
		},
	}
	checker := &mockStatusPageChecker{plan: "hobby"}
	svc := NewStatusPageService(repo, checker, nil, nil)

	_, err := svc.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateStatusPageRequest{
		Title:     "Second Page",
		Subdomain: "second",
		IsPublic:  true,
	})

	if err == nil {
		t.Fatal("expected error for exceeding status page limit, got nil")
	}
}

func TestCreate_InvalidMonitorUUID(t *testing.T) {
	repo := &mockStatusPageRepo{}
	checker := &mockStatusPageChecker{plan: "hobby"}
	svc := NewStatusPageService(repo, checker, nil, nil)

	_, err := svc.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateStatusPageRequest{
		Title:     "Page",
		Subdomain: "page",
		MonitorID: []string{"not-a-valid-uuid"},
	})

	if err == nil {
		t.Fatal("expected error for invalid monitor UUID, got nil")
	}
}

func TestGetBySubdomain_FiltersInactiveMonitors(t *testing.T) {
	activeMonID := uuid.New()
	inactiveMonID := uuid.New()

	repo := &mockStatusPageRepo{
		getBySubdomainFunc: func(subdomain string) (*model.StatusPage, error) {
			desc2 := "All systems"
			return &model.StatusPage{
				Title:       "Acme Status",
				Description: &desc2,
				Subdomain:   subdomain,
				StatusPageMonitors: []model.StatusPageMonitor{
					{
						MonitorID: activeMonID,
						Monitor: model.Monitor{
							ID:           activeMonID,
							Name:         "Active Web",
							Status:       "up",
							IsActive:     true,
							RecentChecks: []model.Check{{ResponseTimeMs: 45}},
						},
					},
					{
						MonitorID: inactiveMonID,
						Monitor: model.Monitor{
							ID:       inactiveMonID,
							Name:     "Paused Web",
							Status:   "paused",
							IsActive: false, // Inactive! Should be filtered out
						},
					},
				},
			}, nil
		},
	}

	svc := NewStatusPageService(repo, &mockStatusPageChecker{}, nil, nil)
	resp, err := svc.GetBySubdomain("acme")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Monitors) != 1 {
		t.Fatalf("expected exactly 1 active monitor, got %d", len(resp.Monitors))
	}
	if resp.Monitors[0].Name != "Active Web" {
		t.Errorf("expected Active Web, got %s", resp.Monitors[0].Name)
	}
}

func TestDelete_NotMember(t *testing.T) {
	checker := &mockStatusPageChecker{
		notMember: true, // Not a member
	}
	repo := &mockStatusPageRepo{}
	svc := NewStatusPageService(repo, checker, nil, nil)

	// Since GetMemberById returns nil member:
	err := svc.Delete(uuid.New().String(), uuid.New().String(), uuid.New().String(), "127.0.0.1")
	if err == nil {
		t.Fatal("expected error when non-member tries to delete status page, got nil")
	}
}

func TestUpdate_PartialFields(t *testing.T) {
	spID := uuid.New()
	orgID := uuid.New()
	existing := &model.StatusPage{
		ID:             spID,
		OrganizationID: orgID,
		Title:          "Old Title",
		Subdomain:      "old-sub",
	}

	repo := &mockStatusPageRepo{
		getByIdFunc: func(sid, oid string) (*model.StatusPage, error) {
			return existing, nil
		},
		updateFunc: func(sid, oid string, sp *model.StatusPage) (*model.StatusPage, error) {
			return sp, nil
		},
	}
	checker := &mockStatusPageChecker{}
	svc := NewStatusPageService(repo, checker, nil, nil)

	newTitle := "Updated Title"
	res, err := svc.Update(spID.String(), orgID.String(), uuid.New().String(), "127.0.0.1", &UpdateStatusPageRequest{
		Title: &newTitle,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Title != "Updated Title" {
		t.Errorf("expected Updated Title, got %s", res.Title)
	}
	if res.Subdomain != "old-sub" {
		t.Errorf("expected subdomain to remain old-sub, got %s", res.Subdomain)
	}
}

// ====================== HANDLER TESTS ======================

func TestStatusPageHandler_Create_Success(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	repo := &mockStatusPageRepo{}
	checker := &mockStatusPageChecker{}
	svc := NewStatusPageService(repo, checker, nil, nil)
	handler := NewStatusPageHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("organizationID", orgID)
		c.Set("userID", userID.String())
		c.Next()
	})
	r.POST("/organization/:orgId/statuspage", handler.CreateStatusPage)

	desc3 := "Production dashboard"
	body := CreateStatusPageRequest{
		Title:       "PulseWatch Status",
		Description: &desc3,
		Subdomain:   "pulsewatch-test",
		IsPublic:    true,
		MonitorID:   []string{uuid.New().String()},
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/organization/"+orgID.String()+"/statuspage", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestStatusPageService_GetAllByOrganizationID(t *testing.T) {
	orgID := uuid.New().String()
	userID := uuid.New().String()
	repo := &mockStatusPageRepo{
		getAllByOrganizationIDFunc: func(o string) ([]model.StatusPage, error) {
			return []model.StatusPage{
				{ID: uuid.New(), Title: "Page 1", Subdomain: "sub1"},
				{ID: uuid.New(), Title: "Page 2", Subdomain: "sub2"},
			}, nil
		},
	}
	checker := &mockStatusPageChecker{}
	svc := NewStatusPageService(repo, checker, nil, nil)

	pages, err := svc.GetAllByOrganizationID(orgID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pages) != 2 {
		t.Fatalf("expected 2 status pages, got %d", len(pages))
	}
}

func TestStatusPageService_GetStatusPageMonitor(t *testing.T) {
	spID := uuid.New()
	repo := &mockStatusPageRepo{
		getStatusPageMonitorFunc: func(id string) ([]model.StatusPageMonitor, error) {
			return []model.StatusPageMonitor{
				{ID: uuid.New(), StatusPageID: spID},
			}, nil
		},
	}
	checker := &mockStatusPageChecker{}
	svc := NewStatusPageService(repo, checker, nil, nil)

	monitors, err := svc.GetStatusPageMonitor(spID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(monitors) != 1 {
		t.Fatalf("expected 1 monitor, got %d", len(monitors))
	}
}

func TestStatusPageHandler_GetAll(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	repo := &mockStatusPageRepo{
		getAllByOrganizationIDFunc: func(o string) ([]model.StatusPage, error) {
			return []model.StatusPage{}, nil
		},
	}
	checker := &mockStatusPageChecker{}
	svc := NewStatusPageService(repo, checker, nil, nil)
	handler := NewStatusPageHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("organizationID", orgID)
		c.Set("userID", userID.String())
		c.Next()
	})
	r.GET("/organization/:orgId/statuspage", handler.GetAllStatusPage)

	req, _ := http.NewRequest(http.MethodGet, "/organization/"+orgID.String()+"/statuspage", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestStatusPageHandler_Delete(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	spID := uuid.New()

	repo := &mockStatusPageRepo{
		deleteFunc: func(sid, oid string) error { return nil },
	}
	checker := &mockStatusPageChecker{}
	svc := NewStatusPageService(repo, checker, nil, nil)
	handler := NewStatusPageHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("organizationID", orgID)
		c.Set("userID", userID.String())
		c.Next()
	})
	r.DELETE("/organization/:orgId/statuspage/:statuspageId", handler.DeleteStatusPage)

	req, _ := http.NewRequest(http.MethodDelete, "/organization/"+orgID.String()+"/statuspage/"+spID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestStatusPageHandler_GetPublicBySubdomain(t *testing.T) {
	repo := &mockStatusPageRepo{
		getBySubdomainFunc: func(sub string) (*model.StatusPage, error) {
			return &model.StatusPage{
				Title:     "Public Status",
				Subdomain: sub,
			}, nil
		},
	}
	checker := &mockStatusPageChecker{}
	svc := NewStatusPageService(repo, checker, nil, nil)
	handler := NewStatusPageHandler(svc)

	r := gin.New()
	r.GET("/public/statuspage/:subdomain", handler.GetPublicBySubdomain)

	req, _ := http.NewRequest(http.MethodGet, "/public/statuspage/pulsewatch", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
