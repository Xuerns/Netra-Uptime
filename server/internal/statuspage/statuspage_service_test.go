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
	"strings"
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
	getStatusPageMonitorFunc        func(statusPageID, orgID string) ([]model.StatusPageMonitor, error)
	getStatusPageByUserIdFunc       func(userId string) ([]model.StatusPage, error)
	totalOrganizationStatusPageFunc           func(orgId string) (int64, error)
	getActiveIncidentsByMonitorIDsFunc        func(monitorIDs []string) ([]model.Incident, error)
	getRecentResolvedIncidentsByMonitorIDsFunc func(monitorIDs []string, days int) ([]model.Incident, error)
	get30DayUptimeMapFunc                     func(monitorIDs []string) (map[string]float64, error)
	getMonitorDailyStatsFunc                  func(monitorIDs []string, days int) (map[string][]model.MonitorDailyStat, error)
	getActiveMaintenancesFunc                 func(statusPageID string) ([]model.ScheduledMaintenance, []model.ScheduledMaintenance, error)
	createMaintenanceFunc                     func(maintenance *model.ScheduledMaintenance, monitorIDs []string) (*model.ScheduledMaintenance, error)
	getMaintenancesByStatusPageIDFunc         func(statusPageID, orgID string) ([]model.ScheduledMaintenance, error)
	getMaintenanceByIDFunc                    func(maintenanceID, orgID string) (*model.ScheduledMaintenance, error)
	updateMaintenanceFunc                     func(maintenance *model.ScheduledMaintenance, monitorIDs *[]string) (*model.ScheduledMaintenance, error)
	deleteMaintenanceFunc                     func(maintenanceID, orgID string) error
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
func (m *mockStatusPageRepo) GetStatusPageMonitor(statusPageID, orgID string) ([]model.StatusPageMonitor, error) {
	if m.getStatusPageMonitorFunc != nil {
		return m.getStatusPageMonitorFunc(statusPageID, orgID)
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
func (m *mockStatusPageRepo) GetActiveIncidentsByMonitorIDs(monitorIDs []string) ([]model.Incident, error) {
	if m.getActiveIncidentsByMonitorIDsFunc != nil {
		return m.getActiveIncidentsByMonitorIDsFunc(monitorIDs)
	}
	return nil, nil
}
func (m *mockStatusPageRepo) GetRecentResolvedIncidentsByMonitorIDs(monitorIDs []string, days int) ([]model.Incident, error) {
	if m.getRecentResolvedIncidentsByMonitorIDsFunc != nil {
		return m.getRecentResolvedIncidentsByMonitorIDsFunc(monitorIDs, days)
	}
	return nil, nil
}
func (m *mockStatusPageRepo) Get30DayUptimeMap(monitorIDs []string) (map[string]float64, error) {
	if m.get30DayUptimeMapFunc != nil {
		return m.get30DayUptimeMapFunc(monitorIDs)
	}
	return make(map[string]float64), nil
}
func (m *mockStatusPageRepo) GetMonitorDailyStats(monitorIDs []string, days int) (map[string][]model.MonitorDailyStat, error) {
	if m.getMonitorDailyStatsFunc != nil {
		return m.getMonitorDailyStatsFunc(monitorIDs, days)
	}
	return make(map[string][]model.MonitorDailyStat), nil
}
func (m *mockStatusPageRepo) GetActiveMaintenances(statusPageID string) ([]model.ScheduledMaintenance, []model.ScheduledMaintenance, error) {
	if m.getActiveMaintenancesFunc != nil {
		return m.getActiveMaintenancesFunc(statusPageID)
	}
	return nil, nil, nil
}
func (m *mockStatusPageRepo) CreateMaintenance(maintenance *model.ScheduledMaintenance, monitorIDs []string) (*model.ScheduledMaintenance, error) {
	if m.createMaintenanceFunc != nil {
		return m.createMaintenanceFunc(maintenance, monitorIDs)
	}
	maintenance.ID = uuid.New()
	return maintenance, nil
}
func (m *mockStatusPageRepo) GetMaintenancesByStatusPageID(statusPageID, orgID string) ([]model.ScheduledMaintenance, error) {
	if m.getMaintenancesByStatusPageIDFunc != nil {
		return m.getMaintenancesByStatusPageIDFunc(statusPageID, orgID)
	}
	return nil, nil
}
func (m *mockStatusPageRepo) GetMaintenanceByID(maintenanceID, orgID string) (*model.ScheduledMaintenance, error) {
	if m.getMaintenanceByIDFunc != nil {
		return m.getMaintenanceByIDFunc(maintenanceID, orgID)
	}
	return nil, errors.New("not found")
}
func (m *mockStatusPageRepo) UpdateMaintenance(maintenance *model.ScheduledMaintenance, monitorIDs *[]string) (*model.ScheduledMaintenance, error) {
	if m.updateMaintenanceFunc != nil {
		return m.updateMaintenanceFunc(maintenance, monitorIDs)
	}
	return maintenance, nil
}
func (m *mockStatusPageRepo) DeleteMaintenance(maintenanceID, orgID string) error {
	if m.deleteMaintenanceFunc != nil {
		return m.deleteMaintenanceFunc(maintenanceID, orgID)
	}
	return nil
}
func (m *mockStatusPageRepo) GetIncidentHistory(monitorIDs []string, days int) ([]model.Incident, error) {
	return nil, nil
}
func (m *mockStatusPageRepo) CreateComponentGroup(group *model.ComponentGroup) (*model.ComponentGroup, error) {
	group.ID = uuid.New()
	return group, nil
}
func (m *mockStatusPageRepo) GetComponentGroupsByStatusPageID(statusPageID string) ([]model.ComponentGroup, error) {
	return nil, nil
}
func (m *mockStatusPageRepo) GetComponentGroupByID(groupID, statusPageID string) (*model.ComponentGroup, error) {
	return &model.ComponentGroup{Name: "Mock Group"}, nil
}
func (m *mockStatusPageRepo) UpdateComponentGroup(group *model.ComponentGroup) (*model.ComponentGroup, error) {
	return group, nil
}
func (m *mockStatusPageRepo) DeleteComponentGroup(groupID, statusPageID string) error {
	return nil
}
func (m *mockStatusPageRepo) AssignMonitorToGroup(statusPageID, monitorID string, groupID *string) error {
	return nil
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
	orgID := uuid.New()
	repo := &mockStatusPageRepo{
		getStatusPageMonitorFunc: func(id, oId string) ([]model.StatusPageMonitor, error) {
			if oId != orgID.String() {
				return nil, errors.New("unauthorized org")
			}
			return []model.StatusPageMonitor{
				{ID: uuid.New(), StatusPageID: spID},
			}, nil
		},
	}
	checker := &mockStatusPageChecker{}
	svc := NewStatusPageService(repo, checker, nil, nil)

	monitors, err := svc.GetStatusPageMonitor(spID.String(), orgID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(monitors) != 1 {
		t.Fatalf("expected 1 monitor, got %d", len(monitors))
	}

	// Test IDOR rejection with wrong org
	_, err = svc.GetStatusPageMonitor(spID.String(), uuid.New().String())
	if err == nil {
		t.Fatal("expected error for unauthorized org accessing status page monitors, got nil")
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

func TestStatusPageService_GetBySubdomain_OverallStatus(t *testing.T) {
	mon1 := uuid.New()
	mon2 := uuid.New()

	tests := []struct {
		name           string
		monitors       []model.StatusPageMonitor
		incidents      []model.Incident
		expectedStatus string
	}{
		{
			name: "all up => operational",
			monitors: []model.StatusPageMonitor{
				{Monitor: model.Monitor{ID: mon1, Status: "up", IsActive: true}},
				{Monitor: model.Monitor{ID: mon2, Status: "up", IsActive: true}},
			},
			expectedStatus: "operational",
		},
		{
			name: "one down => partial_outage",
			monitors: []model.StatusPageMonitor{
				{Monitor: model.Monitor{ID: mon1, Status: "up", IsActive: true}},
				{Monitor: model.Monitor{ID: mon2, Status: "down", IsActive: true}},
			},
			expectedStatus: "partial_outage",
		},
		{
			name: "all down => major_outage",
			monitors: []model.StatusPageMonitor{
				{Monitor: model.Monitor{ID: mon1, Status: "down", IsActive: true}},
				{Monitor: model.Monitor{ID: mon2, Status: "down", IsActive: true}},
			},
			expectedStatus: "major_outage",
		},
		{
			name: "one degraded => degraded",
			monitors: []model.StatusPageMonitor{
				{Monitor: model.Monitor{ID: mon1, Status: "up", IsActive: true}},
				{Monitor: model.Monitor{ID: mon2, Status: "degraded", IsActive: true}},
			},
			expectedStatus: "degraded",
		},
		{
			name: "active critical incident => degraded",
			monitors: []model.StatusPageMonitor{
				{Monitor: model.Monitor{ID: mon1, Status: "up", IsActive: true}},
				{Monitor: model.Monitor{ID: mon2, Status: "up", IsActive: true}},
			},
			incidents: []model.Incident{
				{ID: uuid.New(), Severity: "critical", Status: "ongoing"},
			},
			expectedStatus: "degraded",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockStatusPageRepo{
				getBySubdomainFunc: func(sub string) (*model.StatusPage, error) {
					return &model.StatusPage{
						Title:              "PulseWatch Status",
						Subdomain:          sub,
						StatusPageMonitors: tc.monitors,
					}, nil
				},
				getActiveIncidentsByMonitorIDsFunc: func(ids []string) ([]model.Incident, error) {
					return tc.incidents, nil
				},
			}
			svc := NewStatusPageService(repo, nil, nil, nil)
			resp, err := svc.GetBySubdomain("pulsewatch")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.OverallStatus != tc.expectedStatus {
				t.Errorf("expected overall status %s, got %s", tc.expectedStatus, resp.OverallStatus)
			}
		})
	}
}

func TestStatusPageService_GetBySubdomain_IncidentsAndUptime(t *testing.T) {
	monID := uuid.New()
	incID := uuid.New()
	resolvedTime := time.Now().Add(-1 * time.Hour)

	repo := &mockStatusPageRepo{
		getBySubdomainFunc: func(sub string) (*model.StatusPage, error) {
			return &model.StatusPage{
				Title:     "PulseWatch Status",
				Subdomain: sub,
				StatusPageMonitors: []model.StatusPageMonitor{
					{
						Monitor: model.Monitor{
							ID:       monID,
							Name:     "API Server",
							Type:     "http",
							Status:   "up",
							IsActive: true,
							RecentChecks: []model.Check{
								{ResponseTimeMs: 120},
								{ResponseTimeMs: 95},
							},
						},
					},
				},
			}, nil
		},
		getActiveIncidentsByMonitorIDsFunc: func(ids []string) ([]model.Incident, error) {
			return []model.Incident{
				{
					ID:        incID,
					MonitorID: monID,
					Status:    "ongoing",
					Severity:  "high",
					Summary:   "High latency detected",
					StartedAt: time.Now().Add(-30 * time.Minute),
					Monitor:   model.Monitor{Name: "API Server"},
					IncidentUpdates: []model.IncidentUpdate{
						{
							Message:   "Investigating database queries",
							CreatedAt: time.Now().Add(-15 * time.Minute),
						},
					},
				},
			}, nil
		},
		getRecentResolvedIncidentsByMonitorIDsFunc: func(ids []string, days int) ([]model.Incident, error) {
			return []model.Incident{
				{
					ID:         uuid.New(),
					MonitorID:  monID,
					Status:     "resolved",
					Severity:   "medium",
					Summary:    "DNS lookup timeout",
					StartedAt:  time.Now().Add(-2 * time.Hour),
					ResolvedAt: &resolvedTime,
					Monitor:    model.Monitor{Name: "API Server"},
				},
			}, nil
		},
		get30DayUptimeMapFunc: func(ids []string) (map[string]float64, error) {
			return map[string]float64{
				monID.String(): 99.85,
			}, nil
		},
	}

	svc := NewStatusPageService(repo, nil, nil, nil)
	resp, err := svc.GetBySubdomain("pulsewatch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Monitors) != 1 {
		t.Fatalf("expected 1 monitor, got %d", len(resp.Monitors))
	}
	if resp.Monitors[0].UptimePercent == nil || *resp.Monitors[0].UptimePercent != 99.85 {
		t.Errorf("expected uptime 99.85, got %v", resp.Monitors[0].UptimePercent)
	}

	if len(resp.ActiveIncidents) != 1 {
		t.Fatalf("expected 1 active incident, got %d", len(resp.ActiveIncidents))
	}
	if resp.ActiveIncidents[0].Summary != "High latency detected" {
		t.Errorf("unexpected summary: %s", resp.ActiveIncidents[0].Summary)
	}
	if resp.ActiveIncidents[0].MonitorName != "API Server" {
		t.Errorf("expected monitor name 'API Server', got '%s'", resp.ActiveIncidents[0].MonitorName)
	}
	if len(resp.ActiveIncidents[0].Updates) != 1 {
		t.Fatalf("expected 1 incident update, got %d", len(resp.ActiveIncidents[0].Updates))
	}

	if len(resp.RecentIncidents) != 1 {
		t.Fatalf("expected 1 recent incident, got %d", len(resp.RecentIncidents))
	}
	if resp.RecentIncidents[0].ResolvedAt == nil {
		t.Errorf("expected resolved_at to be set")
	}
}

func TestStatusPageService_GetBySubdomain_90DayStatsAndMaintenance(t *testing.T) {
	spID := uuid.New()
	monID := uuid.New()
	now := time.Now()

	repo := &mockStatusPageRepo{
		getBySubdomainFunc: func(sub string) (*model.StatusPage, error) {
			return &model.StatusPage{
				ID:        spID,
				Title:     "PulseWatch Test",
				Subdomain: sub,
				StatusPageMonitors: []model.StatusPageMonitor{
					{
						Monitor: model.Monitor{
							ID:       monID,
							Name:     "Database Cluster",
							Type:     "tcp",
							Status:   "up",
							IsActive: true,
						},
					},
				},
			}, nil
		},
		getMonitorDailyStatsFunc: func(ids []string, days int) (map[string][]model.MonitorDailyStat, error) {
			return map[string][]model.MonitorDailyStat{
				monID.String(): {
					{
						MonitorID:        monID,
						Date:             now.AddDate(0, 0, -1),
						TotalChecks:      1440,
						UpChecks:         1438,
						DownChecks:       2,
						UptimePercentage: 99.86,
					},
				},
			}, nil
		},
		getActiveMaintenancesFunc: func(statusPageID string) ([]model.ScheduledMaintenance, []model.ScheduledMaintenance, error) {
			ongoing := []model.ScheduledMaintenance{
				{
					ID:             uuid.New(),
					Title:          "Database Migration",
					Description:    "Upgrading to PostgreSQL 17",
					Status:         "in_progress",
					ScheduledStart: now.Add(-30 * time.Minute),
					ScheduledEnd:   now.Add(1 * time.Hour),
					AffectedMonitors: []model.MaintenanceMonitor{
						{Monitor: model.Monitor{Name: "Database Cluster"}},
					},
				},
			}
			upcoming := []model.ScheduledMaintenance{
				{
					ID:             uuid.New(),
					Title:          "Network Switch Upgrade",
					Description:    "Core switch firmware update",
					Status:         "scheduled",
					ScheduledStart: now.Add(24 * time.Hour),
					ScheduledEnd:   now.Add(26 * time.Hour),
				},
			}
			return ongoing, upcoming, nil
		},
	}

	svc := NewStatusPageService(repo, nil, nil, nil)
	resp, err := svc.GetBySubdomain("pulsewatch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Monitors) != 1 {
		t.Fatalf("expected 1 monitor, got %d", len(resp.Monitors))
	}
	// Verify exactly 90 days of stats
	if len(resp.Monitors[0].DailyStats) != 90 {
		t.Fatalf("expected 90 daily stats, got %d", len(resp.Monitors[0].DailyStats))
	}

	// Verify the day with data has has_data = true
	yesterdayStr := now.AddDate(0, 0, -1).Format("2006-01-02")
	foundYesterday := false
	for _, day := range resp.Monitors[0].DailyStats {
		if day.Date == yesterdayStr {
			foundYesterday = true
			if !day.HasData {
				t.Errorf("expected yesterday to have has_data=true")
			}
			if day.UptimePercentage != 99.86 {
				t.Errorf("expected uptime 99.86, got %f", day.UptimePercentage)
			}
		}
	}
	if !foundYesterday {
		t.Errorf("expected yesterday's date %s in daily stats", yesterdayStr)
	}

	// Verify ongoing and upcoming maintenances
	if len(resp.OngoingMaintenances) != 1 {
		t.Fatalf("expected 1 ongoing maintenance, got %d", len(resp.OngoingMaintenances))
	}
	if resp.OngoingMaintenances[0].Title != "Database Migration" {
		t.Errorf("unexpected title: %s", resp.OngoingMaintenances[0].Title)
	}
	if len(resp.OngoingMaintenances[0].AffectedMonitors) != 1 || resp.OngoingMaintenances[0].AffectedMonitors[0] != "Database Cluster" {
		t.Errorf("unexpected affected monitors: %v", resp.OngoingMaintenances[0].AffectedMonitors)
	}

	if len(resp.UpcomingMaintenances) != 1 {
		t.Fatalf("expected 1 upcoming maintenance, got %d", len(resp.UpcomingMaintenances))
	}
	if resp.UpcomingMaintenances[0].Title != "Network Switch Upgrade" {
		t.Errorf("unexpected title: %s", resp.UpcomingMaintenances[0].Title)
	}
}

func TestStatusPageService_MaintenanceCRUD(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	spID := uuid.New()
	mID := uuid.New()

	checker := &mockStatusPageChecker{
		member: &model.OrganizationMember{Role: "admin"},
	}

	repo := &mockStatusPageRepo{
		getByIdFunc: func(sid, oid string) (*model.StatusPage, error) {
			return &model.StatusPage{ID: spID, OrganizationID: orgID}, nil
		},
		createMaintenanceFunc: func(m *model.ScheduledMaintenance, ids []string) (*model.ScheduledMaintenance, error) {
			m.ID = mID
			return m, nil
		},
		getMaintenancesByStatusPageIDFunc: func(sid, oid string) ([]model.ScheduledMaintenance, error) {
			return []model.ScheduledMaintenance{
				{
					ID:           mID,
					StatusPageID: spID,
					Title:        "Routine Reboot",
					Status:       "scheduled",
				},
			}, nil
		},
		getMaintenanceByIDFunc: func(mid, oid string) (*model.ScheduledMaintenance, error) {
			return &model.ScheduledMaintenance{
				ID:           mID,
				StatusPageID: spID,
				Title:        "Routine Reboot",
				Status:       "scheduled",
			}, nil
		},
		updateMaintenanceFunc: func(m *model.ScheduledMaintenance, ids *[]string) (*model.ScheduledMaintenance, error) {
			return m, nil
		},
		deleteMaintenanceFunc: func(mid, oid string) error {
			return nil
		},
	}

	svc := NewStatusPageService(repo, checker, nil, nil)

	// Create
	createReq := &CreateMaintenanceRequest{
		Title:          "Routine Reboot",
		ScheduledStart: time.Now(),
		ScheduledEnd:   time.Now().Add(1 * time.Hour),
	}
	created, err := svc.CreateMaintenance(orgID.String(), spID.String(), userID.String(), "127.0.0.1", createReq)
	if err != nil {
		t.Fatalf("failed to create maintenance: %v", err)
	}
	if created.Title != "Routine Reboot" {
		t.Errorf("expected Routine Reboot, got %s", created.Title)
	}

	// List
	list, err := svc.GetMaintenances(orgID.String(), spID.String(), userID.String())
	if err != nil {
		t.Fatalf("failed to get maintenances: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 maintenance, got %d", len(list))
	}

	// Update
	newTitle := "Emergency Reboot"
	newStatus := "in_progress"
	updated, err := svc.UpdateMaintenance(orgID.String(), spID.String(), mID.String(), userID.String(), "127.0.0.1", &UpdateMaintenanceRequest{
		Title:  &newTitle,
		Status: &newStatus,
	})
	if err != nil {
		t.Fatalf("failed to update maintenance: %v", err)
	}
	if updated.Title != "Emergency Reboot" {
		t.Errorf("expected Emergency Reboot, got %s", updated.Title)
	}
	if updated.ActualStart == nil {
		t.Errorf("expected ActualStart to be set when moving to in_progress")
	}

	// Delete
	err = svc.DeleteMaintenance(orgID.String(), spID.String(), mID.String(), userID.String(), "127.0.0.1")
	if err != nil {
		t.Fatalf("failed to delete maintenance: %v", err)
	}
}

func TestComponentGroupCRUDAndIncidentHistory(t *testing.T) {
	repo := &mockStatusPageRepo{}
	checker := &mockStatusPageChecker{plan: "pro"}
	svc := NewStatusPageService(repo, checker, nil, nil)

	orgID := uuid.New()
	userID := uuid.New()
	spID := uuid.New()
	mID := uuid.New()

	repo.getByIdFunc = func(statuspageID, oID string) (*model.StatusPage, error) {
		return &model.StatusPage{
			ID:             spID,
			OrganizationID: orgID,
			Title:          "Production Status",
			Subdomain:      "prod",
			IsPublic:       true,
			StatusPageMonitors: []model.StatusPageMonitor{
				{
					ID:           uuid.New(),
					StatusPageID: spID,
					MonitorID:    mID,
				},
			},
		}, nil
	}

	// Create Group
	created, err := svc.CreateComponentGroup(orgID.String(), spID.String(), userID.String(), "127.0.0.1", &CreateComponentGroupRequest{
		Name:         "Core Infrastructure",
		DisplayOrder: 1,
	})
	if err != nil {
		t.Fatalf("failed to create component group: %v", err)
	}
	if created.Name != "Core Infrastructure" {
		t.Errorf("expected Core Infrastructure, got %s", created.Name)
	}

	// Get Groups
	groups, err := svc.GetComponentGroups(orgID.String(), spID.String(), userID.String())
	if err != nil {
		t.Fatalf("failed to get component groups: %v", err)
	}
	_ = groups

	// Update Group
	newName := "Core Services"
	updated, err := svc.UpdateComponentGroup(orgID.String(), spID.String(), created.ID.String(), userID.String(), "127.0.0.1", &UpdateComponentGroupRequest{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("failed to update component group: %v", err)
	}
	if updated.Name != "Core Services" {
		t.Errorf("expected Core Services, got %s", updated.Name)
	}

	// Assign Monitor to Group
	gIDStr := created.ID.String()
	err = svc.AssignMonitorToGroup(orgID.String(), spID.String(), mID.String(), userID.String(), "127.0.0.1", &AssignMonitorGroupRequest{
		GroupID: &gIDStr,
	})
	if err != nil {
		t.Fatalf("failed to assign monitor to group: %v", err)
	}

	// Delete Group
	err = svc.DeleteComponentGroup(orgID.String(), spID.String(), created.ID.String(), userID.String(), "127.0.0.1")
	if err != nil {
		t.Fatalf("failed to delete component group: %v", err)
	}

	// Test groupIncidentsByDay
	now := time.Now()
	resAt := now.Add(-10 * time.Minute)
	incidents := []model.Incident{
		{
			ID:         uuid.New(),
			Summary:    "API Timeout Spike",
			Severity:   "high",
			Status:     "resolved",
			StartedAt:  now.Add(-30 * time.Minute),
			ResolvedAt: &resAt,
		},
	}
	dailyGroups := groupIncidentsByDay(incidents, 7)
	if len(dailyGroups) < 7 {
		t.Errorf("expected at least 7 day groups, got %d", len(dailyGroups))
	}
	if len(dailyGroups[0].Incidents) != 1 {
		t.Errorf("expected 1 incident today, got %d", len(dailyGroups[0].Incidents))
	}
}

func TestStatusPageBrandingAndBadgeEndpoints(t *testing.T) {
	subdomain := "acme-test"
	repo := &mockStatusPageRepo{
		getBySubdomainFunc: func(sub string) (*model.StatusPage, error) {
			desc := "Acme status page"
			return &model.StatusPage{
				ID:          uuid.New(),
				Title:       "Acme Corp",
				Subdomain:   subdomain,
				Description: &desc,
				IsPublic:    true,
				LogoURL:     "https://example.com/logo.png",
				FaviconURL:  "https://example.com/favicon.ico",
				ThemeColor:  "#3b82f6",
				CustomCSS:   "body { font-family: sans-serif; }",
			}, nil
		},
	}
	checker := &mockStatusPageChecker{}
	svc := NewStatusPageService(repo, checker, nil, nil)
	handler := NewStatusPageHandler(svc)

	r := gin.New()
	r.GET("/public/statuspage/:subdomain/summary.json", handler.GetPublicSummaryJSON)
	r.GET("/public/statuspage/:subdomain/status.json", handler.GetPublicStatusJSON)
	r.GET("/public/statuspage/:subdomain/badge.svg", handler.GetStatusBadge)

	// Test 1: summary.json
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/public/statuspage/"+subdomain+"/summary.json", nil)
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Errorf("expected 200 for summary.json, got %d", w1.Code)
	}
	var summary PublicStatusPageResponse
	if err := json.Unmarshal(w1.Body.Bytes(), &summary); err != nil {
		t.Fatalf("failed to parse summary.json: %v", err)
	}
	if summary.LogoURL != "https://example.com/logo.png" || summary.ThemeColor != "#3b82f6" {
		t.Errorf("unexpected branding values in summary.json: %+v", summary)
	}

	// Test 2: status.json
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/public/statuspage/"+subdomain+"/status.json", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("expected 200 for status.json, got %d", w2.Code)
	}
	var statResp PublicStatusJSONResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &statResp); err != nil {
		t.Fatalf("failed to parse status.json: %v", err)
	}
	if statResp.Status != "operational" {
		t.Errorf("expected status 'operational', got '%s'", statResp.Status)
	}

	// Test 3: badge.svg
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/public/statuspage/"+subdomain+"/badge.svg", nil)
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Errorf("expected 200 for badge.svg, got %d", w3.Code)
	}
	if w3.Header().Get("Content-Type") != "image/svg+xml" {
		t.Errorf("expected Content-Type image/svg+xml, got %s", w3.Header().Get("Content-Type"))
	}
	svgContent := w3.Body.String()
	if !strings.Contains(svgContent, "<svg") || !strings.Contains(svgContent, "operational") {
		t.Errorf("expected valid SVG with operational text, got: %s", svgContent)
	}
}

