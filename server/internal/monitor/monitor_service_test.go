package monitor

import (
	"netra-uptime/internal/model"
	"netra-uptime/pkg/utils"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func init() {
	gin.SetMode(gin.TestMode)
	os.Setenv("APP_MODE", "saas")
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("nohtml", utils.NoHTMLValidator)
	}
}

// In-process RESP mock redis server
func newMockRedisClient(t *testing.T) (*redis.Client, func()) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on random port: %v", err)
	}

	store := make(map[string]string)
	var mu sync.Mutex

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				r := bufio.NewReader(c)
				for {
					line, err := r.ReadString('\n')
					if err != nil {
						return
					}
					line = strings.TrimSpace(line)
					if len(line) == 0 {
						continue
					}

					var args []string
					if strings.HasPrefix(line, "*") {
						numArgs, _ := strconv.Atoi(line[1:])
						for i := 0; i < numArgs; i++ {
							lenLine, err := r.ReadString('\n')
							if err != nil {
								return
							}
							argLen, _ := strconv.Atoi(strings.TrimSpace(lenLine[1:]))
							buf := make([]byte, argLen+2)
							if _, err := io.ReadFull(r, buf); err != nil {
								return
							}
							args = append(args, string(buf[:argLen]))
						}
					} else {
						args = strings.Fields(line)
					}

					if len(args) == 0 {
						continue
					}

					cmd := strings.ToUpper(args[0])
					switch cmd {
					case "PING":
						c.Write([]byte("+PONG\r\n"))
					case "GET":
						mu.Lock()
						val, ok := store[args[1]]
						mu.Unlock()
						if !ok {
							c.Write([]byte("$-1\r\n"))
						} else {
							fmt.Fprintf(c, "$%d\r\n%s\r\n", len(val), val)
						}
					case "SET":
						mu.Lock()
						store[args[1]] = args[2]
						mu.Unlock()
						c.Write([]byte("+OK\r\n"))
					case "DEL":
						mu.Lock()
						for _, k := range args[1:] {
							delete(store, k)
						}
						mu.Unlock()
						c.Write([]byte(":1\r\n"))
					default:
						c.Write([]byte("-ERR unknown command\r\n"))
					}
				}
			}(conn)
		}
	}()

	client := redis.NewClient(&redis.Options{
		Addr:     l.Addr().String(),
		Protocol: 2,
	})

	cleanup := func() {
		client.Close()
		l.Close()
	}

	return client, cleanup
}

// Mock Monitor Repository
type mockMonitorRepo struct {
	createFunc                   func(monitor *model.Monitor) error
	getByIdFunc                  func(orgId, monitorId string) (*model.Monitor, error)
	getAllByOrgFunc              func(orgId string) ([]model.Monitor, error)
	updateFunc                   func(monitor *model.Monitor) error
	deleteFunc                   func(orgId, monitorId string) error
	getActiveMonitorFunc         func() ([]model.Monitor, error)
	getAllMonitorByUserIdFunc    func(userId string) ([]model.Monitor, error)
	validateUsersInOrgFunc       func(orgId string, userIds []string) bool
	createMonitorPICFunc         func(monitorPIC *model.MonitorPIC) error
	replaceMonitorPICsFunc       func(monitor *model.Monitor, newPICs []model.MonitorPIC) error
	getNumberOfPICFunc           func(orgId, monitorId string) (int, error)
	totalOrganizationMonitorFunc func(orgId string) (int64, error)
	getDailyStatsFunc            func(monitorId string, days int) ([]model.MonitorDailyStat, error)
	get30DayUptimeMapFunc        func(monitorIDs []string) (map[string]float64, error)
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
func (m *mockMonitorRepo) GetAllMonitorByUserId(userId string) ([]model.Monitor, error) {
	if m.getAllMonitorByUserIdFunc != nil {
		return m.getAllMonitorByUserIdFunc(userId)
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

// Mock SubscriptionChecker
type mockSubsChecker struct {
	plan string
}

func (m *mockSubsChecker) GetOrganizationPlan(orgId string) (*model.Subscription, error) {
	plan := m.plan
	if plan == "" {
		plan = "hobby"
	}
	return &model.Subscription{Plan: plan, Status: "active"}, nil
}

// ====================== SERVICE TESTS ======================

func TestCreate_Success_HobbyPlan(t *testing.T) {
	rdb, cleanup := newMockRedisClient(t)
	defer cleanup()

	repo := &mockMonitorRepo{
		totalOrganizationMonitorFunc: func(orgId string) (int64, error) {
			return 2, nil
		},
	}
	svc := NewMonitorService(repo, rdb, &mockSubsChecker{plan: "hobby"}, nil)

	mon, err := svc.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "My API",
		Type:            "http",
		Target:          "https://api.example.com/health",
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
	rdb, cleanup := newMockRedisClient(t)
	defer cleanup()

	repo := &mockMonitorRepo{
		totalOrganizationMonitorFunc: func(orgId string) (int64, error) {
			return 5, nil // Hobby max is 5
		},
	}
	svc := NewMonitorService(repo, rdb, &mockSubsChecker{plan: "hobby"}, nil)

	_, err := svc.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "My API",
		Type:            "http",
		Target:          "https://api.example.com/health",
		IntervalSeconds: 300,
		TimeoutSeconds:  10,
	})

	if err == nil {
		t.Fatal("expected error for exceeding monitor limit, got nil")
	}
}

func TestCreate_ExceedsPICLimit(t *testing.T) {
	rdb, cleanup := newMockRedisClient(t)
	defer cleanup()

	svc := NewMonitorService(&mockMonitorRepo{}, rdb, &mockSubsChecker{plan: "hobby"}, nil)

	_, err := svc.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "My API",
		Type:            "http",
		Target:          "https://api.example.com",
		IntervalSeconds: 300,
		TimeoutSeconds:  10,
		Assignees:       []string{uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String()}, // 6 Assignees > 5 max
	})

	if err == nil {
		t.Fatal("expected error for exceeding Assignee limit, got nil")
	}
}

func TestCreate_ZeroAssignees_Success(t *testing.T) {
	rdb, cleanup := newMockRedisClient(t)
	defer cleanup()

	svc := NewMonitorService(&mockMonitorRepo{}, rdb, &mockSubsChecker{plan: "hobby"}, nil)

	_, err := svc.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "My API",
		Type:            "http",
		Target:          "https://api.example.com",
		IntervalSeconds: 300,
		TimeoutSeconds:  10,
		Assignees:       []string{},
	})

	if err != nil {
		t.Fatalf("expected nil error for 0 assignees, got: %v", err)
	}
}

func TestCreate_InvalidIntervalForPlan(t *testing.T) {
	rdb, cleanup := newMockRedisClient(t)
	defer cleanup()

	svc := NewMonitorService(&mockMonitorRepo{}, rdb, &mockSubsChecker{plan: "hobby"}, nil)

	_, err := svc.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "My API",
		Type:            "http",
		Target:          "https://api.example.com",
		IntervalSeconds: 60, // Hobby min is 300
		TimeoutSeconds:  10,
	})

	if err == nil {
		t.Fatal("expected error for interval below plan minimum, got nil")
	}
}

func TestCreate_TimeoutTooHigh(t *testing.T) {
	rdb, cleanup := newMockRedisClient(t)
	defer cleanup()

	svc := NewMonitorService(&mockMonitorRepo{}, rdb, &mockSubsChecker{plan: "hobby"}, nil)

	_, err := svc.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "My API",
		Type:            "http",
		Target:          "https://api.example.com",
		IntervalSeconds: 300,
		TimeoutSeconds:  25, // Hobby max timeout is 10
	})

	if err == nil {
		t.Fatal("expected error for timeout too high, got nil")
	}
}

func TestCreate_TimeoutTooLow(t *testing.T) {
	rdb, cleanup := newMockRedisClient(t)
	defer cleanup()

	svc := NewMonitorService(&mockMonitorRepo{}, rdb, &mockSubsChecker{plan: "hobby"}, nil)

	_, err := svc.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "My API",
		Type:            "http",
		Target:          "https://api.example.com",
		IntervalSeconds: 300,
		TimeoutSeconds:  5, // < 10
	})

	if err == nil {
		t.Fatal("expected error for timeout < 10, got nil")
	}
}

func TestCreate_ProPlan_HigherLimits(t *testing.T) {
	rdb, cleanup := newMockRedisClient(t)
	defer cleanup()

	repo := &mockMonitorRepo{
		totalOrganizationMonitorFunc: func(orgId string) (int64, error) {
			return 7, nil // 7 < 10 pro limit
		},
	}
	svc := NewMonitorService(repo, rdb, &mockSubsChecker{plan: "pro"}, nil)

	mon, err := svc.Create(uuid.New().String(), uuid.New().String(), "127.0.0.1", &CreateMonitorRequest{
		Name:            "Pro Monitor",
		Type:            "http",
		Target:          "https://api.example.com",
		IntervalSeconds: 180, // Pro allowed 180
		TimeoutSeconds:  15,  // Pro allowed 15
	})

	if err != nil {
		t.Fatalf("unexpected error on pro plan: %v", err)
	}
	if mon == nil {
		t.Fatal("expected monitor, got nil")
	}
}

func TestGetById_CacheMiss_ThenCacheSet(t *testing.T) {
	rdb, cleanup := newMockRedisClient(t)
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
	svc := NewMonitorService(repo, rdb, &mockSubsChecker{}, nil)

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

func TestGetAllMonitor_Success(t *testing.T) {
	rdb, cleanup := newMockRedisClient(t)
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
	svc := NewMonitorService(repo, rdb, &mockSubsChecker{}, nil)

	monitors, err := svc.GetAllMonitor(orgID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(monitors) != 2 {
		t.Fatalf("expected 2 monitors, got %d", len(monitors))
	}
}

func TestUpdateMonitor_PauseAndResume(t *testing.T) {
	rdb, cleanup := newMockRedisClient(t)
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
	svc := NewMonitorService(repo, rdb, &mockSubsChecker{}, nil)

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
	rdb, cleanup := newMockRedisClient(t)
	defer cleanup()

	deleted := false
	repo := &mockMonitorRepo{
		deleteFunc: func(orgId, monitorId string) error {
			deleted = true
			return nil
		},
	}
	svc := NewMonitorService(repo, rdb, &mockSubsChecker{}, nil)

	err := svc.DeleteMonitor(uuid.New().String(), "user-1", "127.0.0.1", uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Fatal("expected Delete to be called")
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
	svc := NewMonitorService(repo, nil, &mockSubsChecker{}, nil)

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
	rdb, cleanup := newMockRedisClient(t)
	defer cleanup()

	orgID := uuid.New()
	userID := uuid.New()
	repo := &mockMonitorRepo{
		totalOrganizationMonitorFunc: func(orgId string) (int64, error) {
			return 1, nil
		},
	}
	svc := NewMonitorService(repo, rdb, &mockSubsChecker{}, nil)
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
		Target:          "https://example.com",
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
	rdb, cleanup := newMockRedisClient(t)
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
	svc := NewMonitorService(repo, rdb, &mockSubsChecker{}, nil)
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
