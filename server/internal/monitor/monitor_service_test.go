package monitor

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"netra-uptime/config"
	"netra-uptime/internal/model"
	"netra-uptime/pkg/cache"
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

func newMockMemoryCache(t *testing.T) (*cache.MemoryCache, func()) {
	t.Helper()
	return cache.NewMemoryCache(), func() {}
}

// Mock Monitor Repository
type mockMonitorRepo struct {
	createFunc                   func(monitor *model.Monitor) error
	getByIdFunc                  func(orgId, monitorId string) (*model.Monitor, error)
	getAllByOrgFunc              func(orgId string) ([]model.Monitor, error)
	updateFunc                   func(monitor *model.Monitor) error
	deleteFunc                   func(orgId, monitorId string) error
	getActiveMonitorFunc         func() ([]model.Monitor, error)
	validateUsersInOrgFunc       func(orgId string, userIds []string) bool
	createMonitorPICFunc         func(monitorPIC *model.MonitorPIC) error
	replaceMonitorPICsFunc       func(monitor *model.Monitor, newPICs []model.MonitorPIC) error
	getNumberOfPICFunc           func(orgId, monitorId string) (int, error)
	totalOrganizationMonitorFunc func(orgId string) (int64, error)
	getDailyStatsFunc            func(monitorId string, days int) ([]model.MonitorDailyStat, error)
	get30DayUptimeMapFunc        func(monitorIDs []string) (map[string]float64, error)
	getFleetOverviewFunc         func(orgId string, windowDays int) (*FleetOverviewResponse, error)
}

func (m *mockMonitorRepo) Create(monitor *model.Monitor) error {
	if m.createFunc != nil {
		return m.createFunc(monitor)
	}
	monitor.ID = uuid.New()
	return nil
}
func (m *mockMonitorRepo) GetById(orgId, monitorId string) (*model.Monitor, error) {
	if m.getByIdFunc != nil {
		return m.getByIdFunc(orgId, monitorId)
	}
	return nil, errors.New("monitor not found")
}
func (m *mockMonitorRepo) GetAllByOrg(orgId string) ([]model.Monitor, error) {
	if m.getAllByOrgFunc != nil {
		return m.getAllByOrgFunc(orgId)
	}
	return nil, nil
}
func (m *mockMonitorRepo) Update(monitor *model.Monitor) error {
	if m.updateFunc != nil {
		return m.updateFunc(monitor)
	}
	return nil
}
func (m *mockMonitorRepo) Delete(orgId, monitorId string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(orgId, monitorId)
	}
	return nil
}
func (m *mockMonitorRepo) GetActiveMonitor() ([]model.Monitor, error) {
	if m.getActiveMonitorFunc != nil {
		return m.getActiveMonitorFunc()
	}
	return nil, nil
}
func (m *mockMonitorRepo) ValidateUsersInOrg(orgId string, userIds []string) bool {
	if m.validateUsersInOrgFunc != nil {
		return m.validateUsersInOrgFunc(orgId, userIds)
	}
	return true
}
func (m *mockMonitorRepo) CreateMonitorPIC(monitorPIC *model.MonitorPIC) error {
	if m.createMonitorPICFunc != nil {
		return m.createMonitorPICFunc(monitorPIC)
	}
	return nil
}
func (m *mockMonitorRepo) ReplaceMonitorPICs(monitor *model.Monitor, newPICs []model.MonitorPIC) error {
	if m.replaceMonitorPICsFunc != nil {
		return m.replaceMonitorPICsFunc(monitor, newPICs)
	}
	return nil
}
func (m *mockMonitorRepo) GetNumberOfPIC(orgId, monitorId string) (int, error) {
	if m.getNumberOfPICFunc != nil {
		return m.getNumberOfPICFunc(orgId, monitorId)
	}
	return 0, nil
}
func (m *mockMonitorRepo) TotalOrganizationMonitor(orgId string) (int64, error) {
	if m.totalOrganizationMonitorFunc != nil {
		return m.totalOrganizationMonitorFunc(orgId)
	}
	return 0, nil
}
func (m *mockMonitorRepo) GetDailyStats(monitorId string, days int) ([]model.MonitorDailyStat, error) {
	if m.getDailyStatsFunc != nil {
		return m.getDailyStatsFunc(monitorId, days)
	}
	return nil, nil
}
func (m *mockMonitorRepo) Get30DayUptimeMap(monitorIDs []string) (map[string]float64, error) {
	if m.get30DayUptimeMapFunc != nil {
		return m.get30DayUptimeMapFunc(monitorIDs)
	}
	return nil, nil
}
func (m *mockMonitorRepo) GetFleetOverview(orgId string, windowDays int) (*FleetOverviewResponse, error) {
	if m.getFleetOverviewFunc != nil {
		return m.getFleetOverviewFunc(orgId, windowDays)
	}
	return &FleetOverviewResponse{WindowDays: windowDays}, nil
}

// ====================== SERVICE TESTS ======================

func TestCreate_Success(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	repo := &mockMonitorRepo{
		totalOrganizationMonitorFunc: func(orgId string) (int64, error) {
			return 2, nil
		},
	}
	svc := NewMonitorService(repo, rdb, nil)

	mon, err := svc.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "My API",
		Type:            "http",
		Target:          "https://1.1.1.1/health",
		IntervalSeconds: 300,
		TimeoutSeconds:  10,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mon.Name != "My API" {
		t.Errorf("expected name My API, got %s", mon.Name)
	}
}

func TestCreate_ExceedsMonitorLimit(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	repo := &mockMonitorRepo{
		totalOrganizationMonitorFunc: func(orgId string) (int64, error) {
			return 5, nil
		},
	}
	svc := NewMonitorService(repo, rdb, nil, &config.Config{SelfHostedMaxMonitors: 5})

	_, err := svc.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "My API",
		Type:            "http",
		Target:          "https://1.1.1.1/health",
		IntervalSeconds: 300,
		TimeoutSeconds:  10,
	})

	if err == nil {
		t.Fatal("expected error for exceeding monitor limit, got nil")
	}
}

func TestCreate_ExceedsPICLimit(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	svc := NewMonitorService(&mockMonitorRepo{}, rdb, nil, &config.Config{SelfHostedMaxPIC: 5})

	_, err := svc.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "My API",
		Type:            "http",
		Target:          "https://1.1.1.1",
		IntervalSeconds: 300,
		TimeoutSeconds:  10,
		Assignees:       []string{uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String()}, // 6 Assignees > 5 max
	})

	if err == nil {
		t.Fatal("expected error for exceeding Assignee limit, got nil")
	}
}

func TestCreate_ZeroAssignees_Success(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	svc := NewMonitorService(&mockMonitorRepo{}, rdb, nil)

	_, err := svc.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "My API",
		Type:            "http",
		Target:          "https://1.1.1.1",
		IntervalSeconds: 300,
		TimeoutSeconds:  10,
		Assignees:       []string{},
	})

	if err != nil {
		t.Fatalf("expected nil error for 0 assignees, got: %v", err)
	}
}

func TestCreate_InvalidIntervalBelowMinimum(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	svc := NewMonitorService(&mockMonitorRepo{}, rdb, nil)

	// Check interval below default minimum (30 seconds)
	_, err := svc.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "My API",
		Type:            "http",
		Target:          "https://1.1.1.1",
		IntervalSeconds: 20, // Min is 30
		TimeoutSeconds:  10,
	})

	if err == nil || err.Error() != "Check interval is below the minimum allowed frequency (minimum 30 seconds)" {
		t.Fatalf("expected 'Check interval is below the minimum allowed frequency (minimum 30 seconds)', got: %v", err)
	}

	// Check interval below custom SelfHostedMinInterval
	svcCustom := NewMonitorService(&mockMonitorRepo{}, rdb, nil, &config.Config{SelfHostedMinInterval: 120})
	_, err = svcCustom.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "My API",
		Type:            "http",
		Target:          "https://1.1.1.1",
		IntervalSeconds: 60, // Min is 120
		TimeoutSeconds:  10,
	})

	if err == nil || err.Error() != "Check interval is below the minimum allowed frequency (minimum 30 seconds)" {
		t.Fatalf("expected error for interval below custom minimum, got: %v", err)
	}
}

func TestCreate_TimeoutTooHigh(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	svc := NewMonitorService(&mockMonitorRepo{}, rdb, nil)

	// Timeout exceeds default maximum (30 seconds)
	_, err := svc.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "My API",
		Type:            "http",
		Target:          "https://1.1.1.1",
		IntervalSeconds: 300,
		TimeoutSeconds:  35, // Max timeout is 30
	})

	if err == nil || err.Error() != "Request timeout exceeds the maximum allowed timeout (maximum 30 seconds)" {
		t.Fatalf("expected 'Request timeout exceeds the maximum allowed timeout (maximum 30 seconds)', got: %v", err)
	}

	// Timeout exceeds custom SelfHostedMaxTimeout
	svcCustom := NewMonitorService(&mockMonitorRepo{}, rdb, nil, &config.Config{SelfHostedMaxTimeout: 20})
	_, err = svcCustom.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "My API",
		Type:            "http",
		Target:          "https://1.1.1.1",
		IntervalSeconds: 300,
		TimeoutSeconds:  25, // Max timeout is 20
	})

	if err == nil || err.Error() != "Request timeout exceeds the maximum allowed timeout (maximum 30 seconds)" {
		t.Fatalf("expected error for timeout exceeding custom limit, got: %v", err)
	}
}

func TestCreate_TimeoutTooLow(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	svc := NewMonitorService(&mockMonitorRepo{}, rdb, nil)

	_, err := svc.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "My API",
		Type:            "http",
		Target:          "https://1.1.1.1",
		IntervalSeconds: 300,
		TimeoutSeconds:  0, // < 1
	})

	if err == nil {
		t.Fatal("expected error for timeout < 1, got nil")
	}
}

func TestCreate_TimeoutGreaterOrEqualInterval(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	svc := NewMonitorService(&mockMonitorRepo{}, rdb, nil)

	// Equal: 30s timeout >= 30s interval
	_, err := svc.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "My API",
		Type:            "http",
		Target:          "https://1.1.1.1",
		IntervalSeconds: 30,
		TimeoutSeconds:  30,
	})
	if err == nil || err.Error() != "monitor timeout must be strictly less than check interval to prevent overlapping checks" {
		t.Fatalf("expected overlapping checks error for timeout == interval, got: %v", err)
	}
}

func TestUpdateMonitor_TimeoutGreaterOrEqualInterval(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	orgID := uuid.New().String()
	monitorID := uuid.New().String()

	repo := &mockMonitorRepo{
		getByIdFunc: func(o, m string) (*model.Monitor, error) {
			return &model.Monitor{
				ID:              uuid.MustParse(monitorID),
				OrganizationID:  uuid.MustParse(orgID),
				IntervalSeconds: 30,
				TimeoutSeconds:  10,
			}, nil
		},
	}

	svc := NewMonitorService(repo, rdb, nil)

	// Update timeout to 30 (equal to current interval 30)
	t30 := 30
	_, err := svc.UpdateMonitor(orgID, uuid.New().String(), "127.0.0.1", monitorID, UpdatedMonitorRequest{
		TimeoutSeconds: &t30,
	})
	if err == nil || err.Error() != "monitor timeout must be strictly less than check interval to prevent overlapping checks" {
		t.Fatalf("expected overlapping checks error on update, got: %v", err)
	}

	// Update interval to 30 when current timeout is 30 (equal to timeout 30)
	repo30 := &mockMonitorRepo{
		getByIdFunc: func(o, m string) (*model.Monitor, error) {
			return &model.Monitor{
				ID:              uuid.MustParse(monitorID),
				OrganizationID:  uuid.MustParse(orgID),
				IntervalSeconds: 60,
				TimeoutSeconds:  30,
			}, nil
		},
	}
	svc30 := NewMonitorService(repo30, rdb, nil)
	i30 := 30
	_, err = svc30.UpdateMonitor(orgID, uuid.New().String(), "127.0.0.1", monitorID, UpdatedMonitorRequest{
		IntervalSeconds: &i30,
	})
	if err == nil || err.Error() != "monitor timeout must be strictly less than check interval to prevent overlapping checks" {
		t.Fatalf("expected overlapping checks error on update interval <= timeout, got: %v", err)
	}
}

func TestUpdateMonitor_IntervalBelowMinimum(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	orgID := uuid.New().String()
	monitorID := uuid.New().String()

	repo := &mockMonitorRepo{
		getByIdFunc: func(o, m string) (*model.Monitor, error) {
			return &model.Monitor{
				ID:              uuid.MustParse(monitorID),
				OrganizationID:  uuid.MustParse(orgID),
				IntervalSeconds: 60,
				TimeoutSeconds:  10,
			}, nil
		},
	}

	svc := NewMonitorService(repo, rdb, nil)

	i20 := 20
	_, err := svc.UpdateMonitor(orgID, uuid.New().String(), "127.0.0.1", monitorID, UpdatedMonitorRequest{
		IntervalSeconds: &i20,
	})
	if err == nil || err.Error() != "Check interval is below the minimum allowed frequency (minimum 30 seconds)" {
		t.Fatalf("expected interval below minimum error on update, got: %v", err)
	}
}

func TestUpdateMonitor_TimeoutTooHigh(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	orgID := uuid.New().String()
	monitorID := uuid.New().String()

	repo := &mockMonitorRepo{
		getByIdFunc: func(o, m string) (*model.Monitor, error) {
			return &model.Monitor{
				ID:              uuid.MustParse(monitorID),
				OrganizationID:  uuid.MustParse(orgID),
				IntervalSeconds: 300,
				TimeoutSeconds:  10,
			}, nil
		},
	}

	svc := NewMonitorService(repo, rdb, nil)

	t35 := 35
	_, err := svc.UpdateMonitor(orgID, uuid.New().String(), "127.0.0.1", monitorID, UpdatedMonitorRequest{
		TimeoutSeconds: &t35,
	})
	if err == nil || err.Error() != "Request timeout exceeds the maximum allowed timeout (maximum 30 seconds)" {
		t.Fatalf("expected timeout exceeds maximum error on update, got: %v", err)
	}
}

func TestCreate_MaxMonitorsPerOrg(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	// Case 1: Reaching custom MaxMonitorsPerOrg
	repo := &mockMonitorRepo{
		totalOrganizationMonitorFunc: func(orgId string) (int64, error) {
			return 5, nil
		},
	}
	svc := NewMonitorService(repo, rdb, nil, &config.Config{MaxMonitorsPerOrg: 5})

	_, err := svc.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "My API",
		Type:            "http",
		Target:          "https://1.1.1.1/health",
		IntervalSeconds: 300,
		TimeoutSeconds:  10,
	})
	if err == nil || err.Error() != "Maximum monitor limit reached for this organization" {
		t.Fatalf("expected 'Maximum monitor limit reached for this organization', got: %v", err)
	}

	// Case 2: Reaching default MaxMonitorsPerOrg fallback (200)
	repoDefault := &mockMonitorRepo{
		totalOrganizationMonitorFunc: func(orgId string) (int64, error) {
			return 200, nil
		},
	}
	svcDefault := NewMonitorService(repoDefault, rdb, nil, &config.Config{MaxMonitorsPerOrg: 0})
	_, err = svcDefault.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "My API",
		Type:            "http",
		Target:          "https://1.1.1.1/health",
		IntervalSeconds: 300,
		TimeoutSeconds:  10,
	})
	if err == nil || err.Error() != "Maximum monitor limit reached for this organization" {
		t.Fatalf("expected 'Maximum monitor limit reached for this organization' on fallback 200, got: %v", err)
	}
}

func TestCreate_CustomConfigLimits(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	repo := &mockMonitorRepo{
		totalOrganizationMonitorFunc: func(orgId string) (int64, error) {
			return 7, nil
		},
	}
	svc := NewMonitorService(repo, rdb, nil, &config.Config{
		SelfHostedMaxMonitors: 20,
		SelfHostedMinInterval: 60,
		SelfHostedMaxTimeout:  60,
	})

	mon, err := svc.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "Custom Monitor",
		Type:            "http",
		Target:          "https://1.1.1.1",
		IntervalSeconds: 180,
		TimeoutSeconds:  15,
	})

	if err != nil {
		t.Fatalf("unexpected error with custom config: %v", err)
	}
	if mon == nil {
		t.Fatal("expected monitor, got nil")
	}
}

func TestGetById_CacheMiss_ThenCacheSet(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	monID := uuid.New()
	orgID := uuid.New()
	repoCalls := 0
	repo := &mockMonitorRepo{
		getByIdFunc: func(o, m string) (*model.Monitor, error) {
			repoCalls++
			return &model.Monitor{
				ID:             monID,
				OrganizationID: orgID,
				Name:           "Cached Monitor",
				Status:         "up",
				IsActive:       true,
			}, nil
		},
	}
	svc := NewMonitorService(repo, rdb, nil)

	// First call: cache miss
	resp1, err := svc.GetById(orgID.String(), monID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp1.Name != "Cached Monitor" {
		t.Errorf("expected Cached Monitor, got %s", resp1.Name)
	}
	if repoCalls != 1 {
		t.Fatalf("expected 1 repo call, got %d", repoCalls)
	}

	// Verify key in cache is scoped by orgId
	expectedKey := "monitor:org:" + orgID.String() + ":" + monID.String()
	val, ok := rdb.Get(expectedKey)
	if !ok || val == "" {
		t.Errorf("expected key %s in cache", expectedKey)
	}

	// Verify legacy global key without orgId does NOT exist
	legacyKey := "monitor:" + monID.String()
	_, ok = rdb.Get(legacyKey)
	if ok {
		t.Errorf("expected legacy key %s to NOT exist in cache", legacyKey)
	}

	// Second call: cache hit! repoCalls should still be 1
	resp2, err := svc.GetById(orgID.String(), monID.String())
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if resp2.Name != "Cached Monitor" {
		t.Errorf("expected Cached Monitor, got %s", resp2.Name)
	}
	if repoCalls != 1 {
		t.Errorf("expected cache hit to not call repo, got %d calls", repoCalls)
	}
}

func TestGetById_CrossTenantCacheIsolation(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	orgA := uuid.New().String()
	orgB := uuid.New().String()
	monID := uuid.New().String()

	repo := &mockMonitorRepo{
		getByIdFunc: func(o, m string) (*model.Monitor, error) {
			if o == orgA && m == monID {
				parsedMonID, _ := uuid.Parse(monID)
				parsedOrgA, _ := uuid.Parse(orgA)
				return &model.Monitor{
					ID:             parsedMonID,
					OrganizationID: parsedOrgA,
					Name:           "Tenant A Private Monitor",
					Status:         "up",
					IsActive:       true,
				}, nil
			}
			return nil, errors.New("monitor not found")
		},
	}
	svc := NewMonitorService(repo, rdb, nil)

	// Tenant A reads their monitor -> populates cache under "monitor:org:orgA:monID"
	respA, err := svc.GetById(orgA, monID)
	if err != nil {
		t.Fatalf("unexpected error for Tenant A: %v", err)
	}
	if respA.Name != "Tenant A Private Monitor" {
		t.Errorf("expected Tenant A Private Monitor, got %s", respA.Name)
	}

	// Tenant B attempts to read Tenant A's monitor using Tenant A's monitorId
	// Because cache is scoped to org, Tenant B hits "monitor:org:orgB:monID" (cache-miss)
	// and repo query for (orgB, monID) fails.
	respB, err := svc.GetById(orgB, monID)
	if err == nil {
		t.Fatalf("expected error for Tenant B cross-tenant read, got nil (leaked data: %+v)", respB)
	}

	// Defensive check: Even if a poisoned/malformed cache entry exists with Org B's key containing Org A's ID
	poisonedResp := MonitorResponse{
		ID:             monID,
		OrganizationID: orgA, // Mismatched organization
		Name:           "Leaked Monitor Data",
	}
	poisonedData, _ := json.Marshal(poisonedResp)
	rdb.Set("monitor:org:"+orgB+":"+monID, string(poisonedData), 60*time.Second)

	respPoisoned, err := svc.GetById(orgB, monID)
	if err == nil {
		t.Fatalf("expected error when cache has mismatched OrganizationID, got response: %+v", respPoisoned)
	}
}

func TestGetAllMonitor_Success(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	orgID := uuid.New()
	repo := &mockMonitorRepo{
		getAllByOrgFunc: func(org string) ([]model.Monitor, error) {
			return []model.Monitor{
				{ID: uuid.New(), OrganizationID: orgID, Name: "Mon 1", Status: "up", IsActive: true},
				{ID: uuid.New(), OrganizationID: orgID, Name: "Mon 2", Status: "down", IsActive: true},
			}, nil
		},
	}
	svc := NewMonitorService(repo, rdb, nil)

	monitors, err := svc.GetAllMonitor(orgID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(monitors) != 2 {
		t.Fatalf("expected 2 monitors, got %d", len(monitors))
	}
}

func TestUpdateMonitor_PauseAndResume(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	monID := uuid.New()
	orgID := uuid.New()
	existing := &model.Monitor{
		ID:              monID,
		OrganizationID:  orgID,
		Name:            "Active Mon",
		Status:          "up",
		IsActive:        true,
		IntervalSeconds: 300,
		TimeoutSeconds:  10,
	}

	repo := &mockMonitorRepo{
		getByIdFunc: func(o, m string) (*model.Monitor, error) {
			return existing, nil
		},
		updateFunc: func(m *model.Monitor) error {
			return nil
		},
	}
	svc := NewMonitorService(repo, rdb, nil)

	// Pause monitor
	pause := false
	mon, err := svc.UpdateMonitor(orgID.String(), "user-1", "127.0.0.1", monID.String(), UpdatedMonitorRequest{
		IsActive: &pause,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mon.Status != "paused" {
		t.Errorf("expected status paused, got %s", mon.Status)
	}

	// Resume monitor
	resume := true
	mon2, err := svc.UpdateMonitor(orgID.String(), "user-1", "127.0.0.1", monID.String(), UpdatedMonitorRequest{
		IsActive: &resume,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mon2.Status != "up" {
		t.Errorf("expected status up after resume, got %s", mon2.Status)
	}
}

func TestDeleteMonitor_Success(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	deleted := false
	repo := &mockMonitorRepo{
		deleteFunc: func(orgId, monitorId string) error {
			deleted = true
			return nil
		},
	}
	svc := NewMonitorService(repo, rdb, nil)

	err := svc.DeleteMonitor(uuid.New().String(), "user-1", "127.0.0.1", uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Fatal("expected Delete to be called")
	}
}

func TestUpdateMonitor_InvalidatesOrgScopedCache(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	monID := uuid.New()
	orgID := uuid.New()
	existing := &model.Monitor{
		ID:              monID,
		OrganizationID:  orgID,
		Name:            "Active Mon",
		Status:          "up",
		IsActive:        true,
		IntervalSeconds: 300,
		TimeoutSeconds:  10,
	}

	repo := &mockMonitorRepo{
		getByIdFunc: func(o, m string) (*model.Monitor, error) {
			return existing, nil
		},
		updateFunc: func(m *model.Monitor) error {
			return nil
		},
	}
	svc := NewMonitorService(repo, rdb, nil)

	monitorKey := "monitor:org:" + orgID.String() + ":" + monID.String()
	listKey := "monitors:org:" + orgID.String()

	rdb.Set(monitorKey, "cached-data", 60*time.Second)
	rdb.Set(listKey, "cached-list", 60*time.Second)

	newName := "Renamed Mon"
	_, err := svc.UpdateMonitor(orgID.String(), "user-1", "127.0.0.1", monID.String(), UpdatedMonitorRequest{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify both cache keys are invalidated
	if _, ok := rdb.Get(monitorKey); ok {
		t.Errorf("expected monitor cache key %s to be invalidated", monitorKey)
	}
	if _, ok := rdb.Get(listKey); ok {
		t.Errorf("expected monitors list cache key %s to be invalidated", listKey)
	}
}

func TestDeleteMonitor_InvalidatesOrgScopedCache(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	monID := uuid.New().String()
	orgID := uuid.New().String()

	repo := &mockMonitorRepo{
		deleteFunc: func(o, m string) error {
			return nil
		},
	}
	svc := NewMonitorService(repo, rdb, nil)

	monitorKey := "monitor:org:" + orgID + ":" + monID
	listKey := "monitors:org:" + orgID

	rdb.Set(monitorKey, "cached-data", 60*time.Second)
	rdb.Set(listKey, "cached-list", 60*time.Second)

	err := svc.DeleteMonitor(orgID, "user-1", "127.0.0.1", monID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify both cache keys are invalidated
	if _, ok := rdb.Get(monitorKey); ok {
		t.Errorf("expected monitor cache key %s to be invalidated", monitorKey)
	}
	if _, ok := rdb.Get(listKey); ok {
		t.Errorf("expected monitors list cache key %s to be invalidated", listKey)
	}
}

func TestGetDailyStats_DefaultDays(t *testing.T) {
	var requestedDays int
	repo := &mockMonitorRepo{
		getByIdFunc: func(orgId, monitorId string) (*model.Monitor, error) {
			return &model.Monitor{ID: uuid.New()}, nil
		},
		getDailyStatsFunc: func(monitorId string, days int) ([]model.MonitorDailyStat, error) {
			requestedDays = days
			return []model.MonitorDailyStat{}, nil
		},
	}
	svc := NewMonitorService(repo, nil, nil)

	_, err := svc.GetDailyStats("org-1", "mon-1", 0) // <= 0 should default to 30
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestedDays != 30 {
		t.Errorf("expected default 30 days, got %d", requestedDays)
	}
}

func TestMapMonitorToResponse_UptimeFromRecentChecks(t *testing.T) {
	mon := model.Monitor{
		ID:       uuid.New(),
		Name:     "Test",
		Status:   "up",
		IsActive: true,
		RecentChecks: []model.Check{
			{Status: "up", CheckedAt: time.Now()},
			{Status: "down", CheckedAt: time.Now()},
			{Status: "up", CheckedAt: time.Now()},
			{Status: "up", CheckedAt: time.Now()},
		},
	}

	resp := mapMonitorToResponse(mon)
	if resp.UptimePercent == nil {
		t.Fatal("expected uptime percent to be calculated")
	}
	// 3 up out of 4 = 75%
	if *resp.UptimePercent != 75.0 {
		t.Errorf("expected 75.0 uptime percent, got %f", *resp.UptimePercent)
	}
}

// ====================== HANDLER TESTS ======================

func TestMonitorHandler_Create_Success(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	orgID := uuid.New()
	userID := uuid.New()
	repo := &mockMonitorRepo{
		totalOrganizationMonitorFunc: func(orgId string) (int64, error) {
			return 1, nil
		},
	}
	svc := NewMonitorService(repo, rdb, nil)
	handler := NewMonitorHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("organizationID", orgID)
		c.Set("userID", userID.String())
		c.Next()
	})
	r.POST("/organization/:orgId/monitors", handler.CreateMonitor)

	body := CreateMonitorRequest{
		Name:            "Dashboard Web",
		Type:            "http",
		Target:          "https://1.1.1.1",
		IntervalSeconds: 300,
		TimeoutSeconds:  10,
		Criticality:     "high",
		MonitorPICs:     []string{uuid.New().String()},
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/organization/"+orgID.String()+"/monitors", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestMonitorHandler_GetById(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	orgID := uuid.New()
	monID := uuid.New()
	repo := &mockMonitorRepo{
		getByIdFunc: func(o, m string) (*model.Monitor, error) {
			return &model.Monitor{
				ID:             monID,
				OrganizationID: orgID,
				Name:           "Mon 1",
				Status:         "up",
				IsActive:       true,
			}, nil
		},
	}
	svc := NewMonitorService(repo, rdb, nil)
	handler := NewMonitorHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("organizationID", orgID)
		c.Next()
	})
	r.GET("/organization/:orgId/monitors/:monitorId", handler.GetById)

	req, _ := http.NewRequest(http.MethodGet, "/organization/"+orgID.String()+"/monitors/"+monID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetFleetOverview_ColdStart(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	orgID := uuid.New().String()
	repo := &mockMonitorRepo{
		getFleetOverviewFunc: func(orgId string, windowDays int) (*FleetOverviewResponse, error) {
			resp := &FleetOverviewResponse{
				WindowDays: windowDays,
			}
			resp.Fleet.Intervals.IsUniform = true
			return resp, nil
		},
	}

	svc := NewMonitorService(repo, rdb, nil)
	overview, err := svc.GetFleetOverview(orgID, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if overview.WindowDays != 30 {
		t.Errorf("expected WindowDays 30, got %d", overview.WindowDays)
	}
	if overview.Uptime.CurrentPercent != nil {
		t.Errorf("expected CurrentPercent nil, got %v", *overview.Uptime.CurrentPercent)
	}
	if overview.Uptime.PreviousPercent != nil {
		t.Errorf("expected PreviousPercent nil, got %v", *overview.Uptime.PreviousPercent)
	}
	if overview.Uptime.TrendDelta != nil {
		t.Errorf("expected TrendDelta nil, got %v", *overview.Uptime.TrendDelta)
	}
	if overview.Latency.AvgMs != nil {
		t.Errorf("expected AvgMs nil, got %v", *overview.Latency.AvgMs)
	}
	if overview.Latency.MedianMs != nil {
		t.Errorf("expected MedianMs nil, got %v", *overview.Latency.MedianMs)
	}
	if overview.Latency.P95Ms != nil {
		t.Errorf("expected P95Ms nil, got %v", *overview.Latency.P95Ms)
	}
	if overview.Latency.FastestMs != nil {
		t.Errorf("expected FastestMs nil, got %v", *overview.Latency.FastestMs)
	}
	if overview.Fleet.TotalTargets != 0 {
		t.Errorf("expected TotalTargets 0, got %d", overview.Fleet.TotalTargets)
	}
}

func TestGetFleetOverview_NewMonitor_NoPriorData(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	orgID := uuid.New().String()
	currPct := 99.5
	repo := &mockMonitorRepo{
		getFleetOverviewFunc: func(orgId string, windowDays int) (*FleetOverviewResponse, error) {
			resp := &FleetOverviewResponse{
				WindowDays: windowDays,
			}
			resp.Uptime.CurrentPercent = &currPct
			resp.Uptime.PreviousPercent = nil
			return resp, nil
		},
	}

	svc := NewMonitorService(repo, rdb, nil)
	overview, err := svc.GetFleetOverview(orgID, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if overview.Uptime.CurrentPercent == nil || *overview.Uptime.CurrentPercent != 99.5 {
		t.Errorf("expected CurrentPercent 99.5, got %v", overview.Uptime.CurrentPercent)
	}
	if overview.Uptime.PreviousPercent != nil {
		t.Errorf("expected PreviousPercent nil, got %v", *overview.Uptime.PreviousPercent)
	}
	if overview.Uptime.TrendDelta != nil {
		t.Errorf("expected TrendDelta nil when prior data is missing, got %v", *overview.Uptime.TrendDelta)
	}
}

func TestGetFleetOverview_EstablishedMonitor_PositiveDelta(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	orgID := uuid.New().String()
	curr := 99.85
	prev := 98.42
	repo := &mockMonitorRepo{
		getFleetOverviewFunc: func(orgId string, windowDays int) (*FleetOverviewResponse, error) {
			resp := &FleetOverviewResponse{
				WindowDays: windowDays,
			}
			resp.Uptime.CurrentPercent = &curr
			resp.Uptime.PreviousPercent = &prev
			return resp, nil
		},
	}

	svc := NewMonitorService(repo, rdb, nil)
	overview, err := svc.GetFleetOverview(orgID, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if overview.Uptime.TrendDelta == nil {
		t.Fatalf("expected TrendDelta not nil")
	}
	if *overview.Uptime.TrendDelta != 1.43 {
		t.Errorf("expected TrendDelta 1.43, got %f", *overview.Uptime.TrendDelta)
	}
}

func TestGetFleetOverview_EstablishedMonitor_NegativeDelta(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	orgID := uuid.New().String()
	curr := 95.50
	prev := 99.00
	repo := &mockMonitorRepo{
		getFleetOverviewFunc: func(orgId string, windowDays int) (*FleetOverviewResponse, error) {
			resp := &FleetOverviewResponse{
				WindowDays: windowDays,
			}
			resp.Uptime.CurrentPercent = &curr
			resp.Uptime.PreviousPercent = &prev
			return resp, nil
		},
	}

	svc := NewMonitorService(repo, rdb, nil)
	overview, err := svc.GetFleetOverview(orgID, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if overview.Uptime.TrendDelta == nil {
		t.Fatalf("expected TrendDelta not nil")
	}
	if *overview.Uptime.TrendDelta != -3.50 {
		t.Errorf("expected TrendDelta -3.50, got %f", *overview.Uptime.TrendDelta)
	}
}

func TestGetFleetOverview_FloatPrecisionRounding(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	orgID := uuid.New().String()
	curr := 99.99
	prev := 99.96
	repo := &mockMonitorRepo{
		getFleetOverviewFunc: func(orgId string, windowDays int) (*FleetOverviewResponse, error) {
			resp := &FleetOverviewResponse{
				WindowDays: windowDays,
			}
			resp.Uptime.CurrentPercent = &curr
			resp.Uptime.PreviousPercent = &prev
			return resp, nil
		},
	}

	svc := NewMonitorService(repo, rdb, nil)
	overview, err := svc.GetFleetOverview(orgID, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if overview.Uptime.TrendDelta == nil {
		t.Fatalf("expected TrendDelta not nil")
	}
	if *overview.Uptime.TrendDelta != 0.03 {
		t.Errorf("expected TrendDelta 0.03, got %f", *overview.Uptime.TrendDelta)
	}
}

func TestGetFleetOverview_MemoryCache(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	orgID := uuid.New().String()
	curr := 99.0
	callCount := 0
	repo := &mockMonitorRepo{
		getFleetOverviewFunc: func(orgId string, windowDays int) (*FleetOverviewResponse, error) {
			callCount++
			resp := &FleetOverviewResponse{
				WindowDays: windowDays,
			}
			resp.Uptime.CurrentPercent = &curr
			return resp, nil
		},
	}

	svc := NewMonitorService(repo, rdb, nil)
	res1, err := svc.GetFleetOverview(orgID, 30)
	if err != nil || res1 == nil {
		t.Fatalf("first call failed: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 repo call, got %d", callCount)
	}

	res2, err := svc.GetFleetOverview(orgID, 30)
	if err != nil || res2 == nil {
		t.Fatalf("second call failed: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected cached response with 1 repo call, got %d", callCount)
	}
}

func TestMonitorHandler_GetFleetOverview(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	orgID := uuid.New()
	curr := 98.5
	repo := &mockMonitorRepo{
		getFleetOverviewFunc: func(orgId string, windowDays int) (*FleetOverviewResponse, error) {
			resp := &FleetOverviewResponse{
				WindowDays: windowDays,
			}
			resp.Uptime.CurrentPercent = &curr
			return resp, nil
		},
	}

	svc := NewMonitorService(repo, rdb, nil)
	handler := NewMonitorHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("organizationID", orgID)
		c.Next()
	})
	r.GET("/organization/:orgId/monitors/overview", handler.GetFleetOverview)

	// Valid with default window
	req, _ := http.NewRequest(http.MethodGet, "/organization/"+orgID.String()+"/monitors/overview", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Valid with custom window
	reqCustom, _ := http.NewRequest(http.MethodGet, "/organization/"+orgID.String()+"/monitors/overview?window=7", nil)
	wCustom := httptest.NewRecorder()
	r.ServeHTTP(wCustom, reqCustom)
	if wCustom.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", wCustom.Code, wCustom.Body.String())
	}

	// Invalid window param (negative)
	reqBad, _ := http.NewRequest(http.MethodGet, "/organization/"+orgID.String()+"/monitors/overview?window=-5", nil)
	wBad := httptest.NewRecorder()
	r.ServeHTTP(wBad, reqBad)
	if wBad.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for negative window, got %d", wBad.Code)
	}

	// Invalid window param (non-integer)
	reqBad2, _ := http.NewRequest(http.MethodGet, "/organization/"+orgID.String()+"/monitors/overview?window=abc", nil)
	wBad2 := httptest.NewRecorder()
	r.ServeHTTP(wBad2, reqBad2)
	if wBad2.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for non-numeric window, got %d", wBad2.Code)
	}
}

func TestUpdateMonitor_DisableSSLDomainCheck_ResetsFieldsAndInvalidatesOverviewCache(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	monID := uuid.New()
	orgID := uuid.New()
	daysLeft := 14
	now := time.Now()
	reminderDay := 14
	existing := &model.Monitor{
		ID:                        monID,
		OrganizationID:            orgID,
		Name:                      "Portofolio",
		Type:                      "http",
		Target:                    "https://1.1.1.1",
		Status:                    "up",
		IsActive:                  true,
		IntervalSeconds:           300,
		TimeoutSeconds:            10,
		EnableSSLDomainCheck:      true,
		SSLStatus:                 "warning",
		DomainStatus:              "error",
		SSLDaysLeft:               &daysLeft,
		DomainDaysLeft:            &daysLeft,
		SSLExpiresAt:              &now,
		DomainExpiresAt:           &now,
		SSLError:                  "cert expiring soon",
		DomainError:               "whois lookup failed",
		LastSSLCheckedAt:          &now,
		LastDomainCheckedAt:       &now,
		LastSSLSentReminderDay:    &reminderDay,
		LastDomainSentReminderDay: &reminderDay,
	}

	repo := &mockMonitorRepo{
		getByIdFunc: func(o, m string) (*model.Monitor, error) {
			return existing, nil
		},
		updateFunc: func(m *model.Monitor) error {
			return nil
		},
	}
	svc := NewMonitorService(repo, rdb, nil)

	overviewKey := "org:" + orgID.String() + ":monitor_overview:30d"
	rdb.Set(overviewKey, "stale-overview", 60*time.Second)

	disableCheck := false
	updated, err := svc.UpdateMonitor(orgID.String(), "user-1", "127.0.0.1", monID.String(), UpdatedMonitorRequest{
		EnableSSLDomainCheck: &disableCheck,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.EnableSSLDomainCheck {
		t.Errorf("expected EnableSSLDomainCheck false, got true")
	}
	if updated.SSLStatus != "none" || updated.DomainStatus != "none" {
		t.Errorf("expected SSLStatus and DomainStatus 'none', got %s / %s", updated.SSLStatus, updated.DomainStatus)
	}
	if updated.SSLDaysLeft != nil || updated.DomainDaysLeft != nil {
		t.Errorf("expected SSLDaysLeft and DomainDaysLeft nil")
	}
	if updated.SSLExpiresAt != nil || updated.DomainExpiresAt != nil {
		t.Errorf("expected SSLExpiresAt and DomainExpiresAt nil")
	}
	if updated.SSLError != "" || updated.DomainError != "" {
		t.Errorf("expected SSLError and DomainError empty, got %q / %q", updated.SSLError, updated.DomainError)
	}
	if updated.LastSSLCheckedAt != nil || updated.LastDomainCheckedAt != nil {
		t.Errorf("expected LastSSLCheckedAt and LastDomainCheckedAt nil")
	}
	if updated.LastSSLSentReminderDay != nil || updated.LastDomainSentReminderDay != nil {
		t.Errorf("expected LastSSLSentReminderDay and LastDomainSentReminderDay nil")
	}

	if _, ok := rdb.Get(overviewKey); ok {
		t.Errorf("expected overview cache key %s to be invalidated on update", overviewKey)
	}
}

func TestCreateAndUpdateMonitor_HTTPMethodHeadersBodyValidation(t *testing.T) {
	rdb, cleanup := newMockMemoryCache(t)
	defer cleanup()

	orgID := uuid.New().String()
	userID := uuid.New().String()

	var savedMonitor *model.Monitor
	repo := &mockMonitorRepo{
		totalOrganizationMonitorFunc: func(orgId string) (int64, error) {
			return 1, nil
		},
		createFunc: func(m *model.Monitor) error {
			m.ID = uuid.New()
			savedMonitor = m
			return nil
		},
		getByIdFunc: func(o, m string) (*model.Monitor, error) {
			if savedMonitor != nil {
				return savedMonitor, nil
			}
			return nil, errors.New("not found")
		},
		updateFunc: func(m *model.Monitor) error {
			savedMonitor = m
			return nil
		},
	}

	svc := NewMonitorService(repo, rdb, nil)

	// 1. Create POST monitor with valid headers and body
	mon, err := svc.Create(orgID, userID, "127.0.0.1", &CreateMonitorRequest{
		Name:            "Webhook Endpoint",
		Type:            "http",
		Target:          "https://1.1.1.1/webhook",
		IntervalSeconds: 60,
		TimeoutSeconds:  10,
		HTTPMethod:      "post",
		HTTPHeaders:     `{"Authorization": "Bearer token123", "X-Custom": "val"}`,
		HTTPBody:        `{"event":"ping"}`,
		Criticality:     "high",
	})
	if err != nil {
		t.Fatalf("expected valid POST monitor creation to succeed, got: %v", err)
	}
	if mon.HTTPMethod != "POST" {
		t.Errorf("expected normalized HTTPMethod 'POST', got %q", mon.HTTPMethod)
	}
	if mon.HTTPBody != `{"event":"ping"}` {
		t.Errorf("expected HTTPBody preserved for POST, got %q", mon.HTTPBody)
	}

	// 2. Reject invalid JSON in HTTPHeaders
	_, err = svc.Create(orgID, userID, "127.0.0.1", &CreateMonitorRequest{
		Name:            "Bad Headers",
		Type:            "http",
		Target:          "https://1.1.1.1/bad",
		IntervalSeconds: 60,
		TimeoutSeconds:  10,
		HTTPMethod:      "GET",
		HTTPHeaders:     `not-a-json`,
		Criticality:     "medium",
	})
	if err == nil {
		t.Fatal("expected error for invalid JSON in http_headers, got nil")
	}

	// 3. Switching method to GET on update should clear HTTPBody
	getMethod := "GET"
	updated, err := svc.UpdateMonitor(orgID, userID, "127.0.0.1", mon.ID.String(), UpdatedMonitorRequest{
		HTTPMethod: &getMethod,
	})
	if err != nil {
		t.Fatalf("unexpected error updating HTTPMethod to GET: %v", err)
	}
	if updated.HTTPMethod != "GET" {
		t.Errorf("expected HTTPMethod 'GET', got %q", updated.HTTPMethod)
	}
	if updated.HTTPBody != "" {
		t.Errorf("expected HTTPBody to be cleared when method is GET, got %q", updated.HTTPBody)
	}
}

func TestMapMonitorToResponse_AvailabilityVsPerformanceAndMaintenance(t *testing.T) {
	now := time.Now()
	mon := model.Monitor{
		ID:         uuid.New(),
		Name:       "API",
		Type:       "http",
		HTTPMethod: "POST",
		Status:     "up",
		IsActive:   true,
		RecentChecks: []model.Check{
			{Status: "up", StatusCode: 200, ResponseTimeMs: 80, CheckedAt: now},
			{Status: "degraded", StatusCode: 200, ResponseTimeMs: 1800, ErrorMessage: "Response time 1800ms exceeded 1000", CheckedAt: now},
			{Status: "degraded", StatusCode: 0, ResponseTimeMs: 10000, ErrorMessage: "Request timed out after 10s", CheckedAt: now},
			{Status: "down", StatusCode: 502, ResponseTimeMs: 50, ResponseSnippet: `{"error":"bad gateway"}`, CheckedAt: now},
			{Status: "maintenance", StatusCode: 503, ResponseTimeMs: 20, CheckedAt: now},
		},
	}

	resp := mapMonitorToResponse(mon)
	if resp.UptimePercent == nil {
		t.Fatal("expected UptimePercent to be non-nil")
	}
	// Active checks = 4 (maintenance ignored). Available (Up) = 2 (up + degraded 200 OK). Unavailable = 2 (degraded timeout + down 502).
	if *resp.UptimePercent != 50.0 {
		t.Errorf("expected UptimePercent=50.0, got %f", *resp.UptimePercent)
	}
	if resp.HTTPMethod != "POST" {
		t.Errorf("expected HTTPMethod='POST', got %q", resp.HTTPMethod)
	}
	if resp.RecentChecks[3].ResponseSnippet == nil || *resp.RecentChecks[3].ResponseSnippet != `{"error":"bad gateway"}` {
		t.Errorf("expected ResponseSnippet on 4th recent check, got %v", resp.RecentChecks[3].ResponseSnippet)
	}
}
