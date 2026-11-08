package worker_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"netra-uptime/config"
	"netra-uptime/internal/model"
	"netra-uptime/internal/monitor"
	"netra-uptime/internal/worker"
	"netra-uptime/pkg/cache"
	"netra-uptime/pkg/database"
	"netra-uptime/pkg/utils"

	"gorm.io/gorm"
)

type mockEmailSender struct {
	sentCount atomic.Int32
}

func (m *mockEmailSender) SendEmail(to []string, subject string, htmlContent string) (string, error) {
	m.sentCount.Add(1)
	return "mock-email-id", nil
}

func (m *mockEmailSender) SendAlert(to []string, subject string, htmlContent string) error {
	m.sentCount.Add(1)
	return nil
}

func (m *mockEmailSender) SendExpiryReminder(to []string, monitorName, target, expiryType string, daysRemaining int, expiresAt string) (string, error) {
	m.sentCount.Add(1)
	return "mock-expiry-id", nil
}

func setupWorkerTestDB(t *testing.T) (*gorm.DB, *config.Config) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "worker_pool_test.db")
	cfg := &config.Config{
		DBDriver:          "sqlite",
		DBPath:            dbPath,
		RecentChecksLimit: 60,
	}

	db, err := database.NewDatabase(cfg)
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	err = db.AutoMigrate(
		&model.User{},
		&model.Check{},
		&model.Incident{},
		&model.IncidentUpdate{},
		&model.Monitor{},
		&model.NotificationChannel{},
		&model.NotificationLog{},
		&model.Organization{},
		&model.OrganizationMember{},
		&model.StatusPage{},
		&model.StatusPageMonitor{},
		&model.IncidentAssignee{},
		&model.MonitorPIC{},
		&model.MonitorDailyStat{},
		&model.AlertRule{},
		&model.AlertRuleTemplate{},
		&model.ScheduledMaintenance{},
		&model.MaintenanceMonitor{},
		&model.ComponentGroup{},
	)
	if err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	return db, cfg
}

func startPrivateServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	utils.SetDefaultAllowPrivateNetwork(true)
	t.Cleanup(func() {
		utils.SetDefaultAllowPrivateNetwork(false)
	})

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		t.Fatalf("failed to list interface addresses: %v", err)
	}
	var privateIP string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip4 := ipnet.IP.To4(); ip4 != nil && ip4.IsPrivate() {
				privateIP = ip4.String()
				break
			}
		}
	}
	if privateIP == "" {
		t.Skip("no private network interface available for test server")
	}

	ln, err := net.Listen("tcp", privateIP+":0")
	if err != nil {
		t.Fatalf("failed to listen on %s:0: %v", privateIP, err)
	}
	srv := httptest.NewUnstartedServer(handler)
	_ = srv.Listener.Close()
	srv.Listener = ln
	srv.Start()
	t.Cleanup(srv.Close)
	return srv
}

func TestWorkerPool_ImmediateRetry_PreventsFalsePositive(t *testing.T) {
	db, _ := setupWorkerTestDB(t)

	var reqCount atomic.Int32
	srv := startPrivateServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := reqCount.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"transient blip"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))

	org := &model.Organization{Name: "Retry Org", Slug: "retry-org"}
	if err := db.Create(org).Error; err != nil {
		t.Fatalf("failed to create org: %v", err)
	}

	mon := &model.Monitor{
		OrganizationID:     org.ID,
		Name:               "Flaky Service",
		Type:               "http",
		Target:             srv.URL,
		HTTPMethod:         "GET",
		IntervalSeconds:    60,
		TimeoutSeconds:     5,
		ExpectedStatusCode: 200,
		IsActive:           true,
		Status:             "up",
		Criticality:        "high",
	}
	if err := db.Create(mon).Error; err != nil {
		t.Fatalf("failed to create monitor: %v", err)
	}

	emailMock := &mockEmailSender{}
	wp := worker.NewWorkerPool(db, cache.NewMemoryCache(), 1, emailMock, nil)
	wp.SetRetryDelay(5 * time.Millisecond)

	check, err := wp.ProcessImmediate(mon.ID)
	if err != nil {
		t.Fatalf("ProcessImmediate failed: %v", err)
	}
	wp.WaitNotifications()

	if reqCount.Load() != 2 {
		t.Errorf("expected 2 requests (1 initial + 1 immediate retry), got %d", reqCount.Load())
	}
	if check.Status != "up" {
		t.Errorf("expected check status 'up' after successful retry, got %q", check.Status)
	}

	var dbMon model.Monitor
	db.First(&dbMon, "id = ?", mon.ID)
	if dbMon.FailureCount != 0 {
		t.Errorf("expected FailureCount=0 after successful retry, got %d", dbMon.FailureCount)
	}

	var checkCount int64
	db.Model(&model.Check{}).Where("monitor_id = ?", mon.ID).Count(&checkCount)
	if checkCount != 1 {
		t.Errorf("expected exactly 1 check record in DB, got %d", checkCount)
	}
}

func TestWorkerPool_PersistentFailure_RecordsSnippetAndIncident(t *testing.T) {
	db, _ := setupWorkerTestDB(t)

	var reqCount atomic.Int32
	srv := startPrivateServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCount.Add(1)
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"upstream database connection failed"}`))
	}))

	org := &model.Organization{Name: "Fail Org", Slug: "fail-org"}
	db.Create(org)

	mon := &model.Monitor{
		OrganizationID:     org.ID,
		Name:               "Broken Service",
		Type:               "http",
		Target:             srv.URL,
		HTTPMethod:         "POST",
		HTTPHeaders:        `{"X-Api-Version":"2"}`,
		HTTPBody:           `{"check":true}`,
		IntervalSeconds:    60,
		TimeoutSeconds:     5,
		ExpectedStatusCode: 200,
		IsActive:           true,
		Status:             "up",
		FailureCount:       2, // Next breach (3rd) triggers incident
		Criticality:        "high",
	}
	db.Create(mon)

	emailMock := &mockEmailSender{}
	wp := worker.NewWorkerPool(db, cache.NewMemoryCache(), 1, emailMock, nil)
	wp.SetRetryDelay(5 * time.Millisecond)

	check, err := wp.ProcessImmediate(mon.ID)
	if err != nil {
		t.Fatalf("ProcessImmediate failed: %v", err)
	}
	wp.WaitNotifications()

	if reqCount.Load() != 2 {
		t.Errorf("expected 2 requests (initial + 1 retry max), got %d", reqCount.Load())
	}
	if check.Status != "down" {
		t.Errorf("expected check status 'down', got %q", check.Status)
	}
	if !strings.Contains(check.ResponseSnippet, "upstream database connection failed") {
		t.Errorf("expected check ResponseSnippet to contain error body, got %q", check.ResponseSnippet)
	}
	if !strings.Contains(check.ErrorMessage, "Response:") {
		t.Errorf("expected check ErrorMessage to include response snippet, got %q", check.ErrorMessage)
	}

	var inc model.Incident
	if err := db.Where("monitor_id = ? AND status = ?", mon.ID, "ongoing").First(&inc).Error; err != nil {
		t.Fatalf("expected ongoing incident to be created, got err: %v", err)
	}
	if !strings.Contains(inc.ResponseSnippet, "upstream database connection failed") {
		t.Errorf("expected incident ResponseSnippet to contain error body, got %q", inc.ResponseSnippet)
	}
}

func TestWorkerPool_ScheduledMaintenance_SuppressesIncidentAndProtectsUptime(t *testing.T) {
	db, cfg := setupWorkerTestDB(t)

	srv := startPrivateServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"maintenance_reboot"}`))
	}))

	org := &model.Organization{Name: "Maint Org", Slug: "maint-org"}
	db.Create(org)

	user := &model.User{Email: "oncall@example.com", Name: "OnCall"}
	db.Create(user)

	sp := &model.StatusPage{
		OrganizationID: org.ID,
		Subdomain:      "maint-status",
		Title:          "Maint Status",
		IsPublic:       true,
	}
	db.Create(sp)

	mon := &model.Monitor{
		OrganizationID:     org.ID,
		Name:               "Database Primary",
		Type:               "http",
		Target:             srv.URL,
		IntervalSeconds:    60,
		TimeoutSeconds:     5,
		ExpectedStatusCode: 200,
		IsActive:           true,
		Status:             "up",
		FailureCount:       2, // Even on threshold edge, maintenance must suppress incident
		Criticality:        "extreme",
	}
	db.Create(mon)
	db.Create(&model.MonitorPIC{MonitorID: mon.ID, UserID: user.ID})

	// Seed 1 healthy check before maintenance
	now := time.Now().UTC()
	db.Create(&model.Check{
		MonitorID:      mon.ID,
		Status:         "up",
		StatusCode:     200,
		ResponseTimeMs: 35,
		CheckedAt:      now.Add(-15 * time.Minute),
	})

	// Create active ScheduledMaintenance (started 5m ago, ends in 30m, status initially 'scheduled')
	activeMaint := &model.ScheduledMaintenance{
		OrganizationID: org.ID,
		StatusPageID:   sp.ID,
		Title:          "PostgreSQL 17 Upgrade",
		Description:    "Planned DB upgrade",
		Status:         "scheduled",
		ScheduledStart: now.Add(-5 * time.Minute),
		ScheduledEnd:   now.Add(30 * time.Minute),
	}
	db.Create(activeMaint)
	db.Create(&model.MaintenanceMonitor{
		MaintenanceID: activeMaint.ID,
		MonitorID:     mon.ID,
	})

	// Also create a past ScheduledMaintenance that should auto-transition to 'completed'
	expiredMaint := &model.ScheduledMaintenance{
		OrganizationID: org.ID,
		StatusPageID:   sp.ID,
		Title:          "Old Patching",
		Status:         "in_progress",
		ScheduledStart: now.Add(-2 * time.Hour),
		ScheduledEnd:   now.Add(-1 * time.Hour),
	}
	db.Create(expiredMaint)

	emailMock := &mockEmailSender{}
	wp := worker.NewWorkerPool(db, cache.NewMemoryCache(), 1, emailMock, nil)
	wp.SetRetryDelay(5 * time.Millisecond)

	check, err := wp.ProcessImmediate(mon.ID)
	if err != nil {
		t.Fatalf("ProcessImmediate failed: %v", err)
	}
	wp.WaitNotifications()

	// 1. Check status must be 'maintenance' and contain maintenance title
	if check.Status != "maintenance" {
		t.Errorf("expected check status 'maintenance', got %q", check.Status)
	}
	if !strings.Contains(check.ErrorMessage, "[Maintenance: PostgreSQL 17 Upgrade]") {
		t.Errorf("expected check ErrorMessage to include maintenance prefix, got %q", check.ErrorMessage)
	}

	// 2. ScheduledMaintenance status must transition to 'in_progress', expired to 'completed'
	var updatedActive model.ScheduledMaintenance
	db.First(&updatedActive, "id = ?", activeMaint.ID)
	if updatedActive.Status != "in_progress" {
		t.Errorf("expected active maintenance status 'in_progress', got %q", updatedActive.Status)
	}

	var updatedExpired model.ScheduledMaintenance
	db.First(&updatedExpired, "id = ?", expiredMaint.ID)
	if updatedExpired.Status != "completed" {
		t.Errorf("expected expired maintenance status 'completed', got %q", updatedExpired.Status)
	}

	// 3. No Incident created & no email alert sent & monitor status remains 'up'
	var incidentCount int64
	db.Model(&model.Incident{}).Where("monitor_id = ?", mon.ID).Count(&incidentCount)
	if incidentCount != 0 {
		t.Errorf("expected 0 incidents during maintenance, got %d", incidentCount)
	}
	if emailMock.sentCount.Load() != 0 {
		t.Errorf("expected 0 email notifications during maintenance, got %d", emailMock.sentCount.Load())
	}

	var dbMon model.Monitor
	db.First(&dbMon, "id = ?", mon.ID)
	if dbMon.Status != "up" {
		t.Errorf("expected monitor status to remain 'up' during maintenance, got %q", dbMon.Status)
	}

	// 4. 30-day Uptime SLA must remain 100% (maintenance check excluded)
	monitorRepo := monitor.NewMonitorRepository(db, cfg)
	uptimeMap, err := monitorRepo.Get30DayUptimeMap([]string{mon.ID.String()})
	if err != nil {
		t.Fatalf("Get30DayUptimeMap failed: %v", err)
	}
	if uptimeMap[mon.ID.String()] != 100.0 {
		t.Errorf("expected 100.0%% uptime when only failure happened during maintenance, got %f%%", uptimeMap[mon.ID.String()])
	}
}
