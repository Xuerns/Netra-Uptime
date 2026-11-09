package monitor_test

import (
	"path/filepath"
	"testing"
	"time"

	"netra-uptime/config"
	"netra-uptime/internal/model"
	"netra-uptime/internal/monitor"
	"netra-uptime/internal/worker"
	"netra-uptime/pkg/database"
)

func TestDualModeSQLite_QueriesAndRetention(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "netra_test.db")
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
		&model.RefreshToken{},
		&model.StatusPage{},
		&model.StatusPageMonitor{},
		&model.OrganizationInvitation{},
		&model.PasswordResetToken{},
		&model.IncidentAssignee{},
		&model.MonitorPIC{},
		&model.AuditLog{},
		&model.APIKey{},
		&model.MonitorDailyStat{},
		&model.AlertRule{},
		&model.AlertRuleTemplate{},
		&model.ScheduledMaintenance{},
		&model.MaintenanceMonitor{},
		&model.ComponentGroup{},
	)
	if err != nil {
		t.Fatalf("AutoMigrate failed on SQLite: %v", err)
	}

	org := &model.Organization{
		Name: "SQLite Fleet Org",
		Slug: "sqlite-fleet-org",
	}
	if err := db.Create(org).Error; err != nil {
		t.Fatalf("Failed to seed Organization: %v", err)
	}

	mon := &model.Monitor{
		OrganizationID:       org.ID,
		Name:                 "API Gateway",
		Type:                 "http",
		Target:               "https://api.example.com/health",
		TargetDomain:         "api.example.com",
		IntervalSeconds:      60,
		TimeoutSeconds:       30,
		ExpectedStatusCode:   200,
		IsActive:             true,
		Status:               "up",
		Criticality:          "high",
		EnableSSLDomainCheck: true,
		SSLStatus:            "valid",
		DomainStatus:         "valid",
	}
	if err := db.Create(mon).Error; err != nil {
		t.Fatalf("Failed to seed Monitor: %v", err)
	}

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.UTC)
	twoDaysAgo := time.Date(now.Year(), now.Month(), now.Day()-2, 0, 0, 0, 0, time.UTC)

	// Seed an existing daily stat for 2 days ago to test both reading and ON CONFLICT upsert merging
	dailyStat := &model.MonitorDailyStat{
		MonitorID:        mon.ID,
		Date:             twoDaysAgo,
		TotalChecks:      10,
		UpChecks:         9,
		DownChecks:       1,
		UptimePercentage: 90.0,
		AvgLatencyMs:     100,
		MinLatencyMs:     80,
		MaxLatencyMs:     150,
		Count2xx:         9,
		Count5xx:         1,
	}
	if err := db.Create(dailyStat).Error; err != nil {
		t.Fatalf("Failed to seed MonitorDailyStat: %v", err)
	}

	// Seed recent checks on today (within 24h) and old checks (2 days ago, older than 24h retention threshold)
	checks := []model.Check{
		{
			MonitorID:      mon.ID,
			Status:         "up",
			StatusCode:     200,
			ResponseTimeMs: 40,
			CheckedAt:      today.Add(1 * time.Hour),
		},
		{
			MonitorID:      mon.ID,
			Status:         "up",
			StatusCode:     200,
			ResponseTimeMs: 60,
			CheckedAt:      today.Add(2 * time.Hour),
		},
		{
			MonitorID:      mon.ID,
			Status:         "up",
			StatusCode:     200,
			ResponseTimeMs: 120,
			CheckedAt:      today.Add(3 * time.Hour),
		},
		// Old check (>24h ago on twoDaysAgo) to be aggregated by RetentionJob.CleanOldChecks()
		{
			MonitorID:      mon.ID,
			Status:         "up",
			StatusCode:     200,
			ResponseTimeMs: 50,
			CheckedAt:      twoDaysAgo.Add(12 * time.Hour),
		},
	}
	for i := range checks {
		if err := db.Create(&checks[i]).Error; err != nil {
			t.Fatalf("Failed to seed Check %d: %v", i, err)
		}
	}

	monitorRepo := monitor.NewMonitorRepository(db, cfg)

	// 1. Test GetFleetOverview (CTE uptimeQuery + Go percentileCont latencyQuery)
	overview, err := monitorRepo.GetFleetOverview(org.ID.String(), 30)
	if err != nil {
		t.Fatalf("GetFleetOverview failed on SQLite: %v", err)
	}
	if overview.Uptime.CurrentPercent == nil {
		t.Fatalf("Expected non-nil CurrentPercent in GetFleetOverview")
	}
	if overview.Latency.AvgMs == nil || overview.Latency.MedianMs == nil || overview.Latency.P95Ms == nil || overview.Latency.FastestMs == nil {
		t.Fatalf("Expected non-nil latency metrics in GetFleetOverview, got %+v", overview.Latency)
	}
	if *overview.Latency.FastestMs != 40 {
		t.Errorf("Expected FastestMs=40, got %d", *overview.Latency.FastestMs)
	}
	if *overview.Latency.MedianMs != 60 {
		t.Errorf("Expected MedianMs=60, got %d", *overview.Latency.MedianMs)
	}

	// 2. Test GetDailyStats (DATE(checked_at) string scanning on SQLite)
	stats, err := monitorRepo.GetDailyStats(mon.ID.String(), 30)
	if err != nil {
		t.Fatalf("GetDailyStats failed on SQLite: %v", err)
	}
	if len(stats) < 2 {
		t.Fatalf("Expected at least 2 daily stat entries (twoDaysAgo + today), got %d", len(stats))
	}
	todayStat := stats[len(stats)-1]
	if todayStat.AvgLatencyMs != 73 {
		t.Errorf("Expected todayStat.AvgLatencyMs=73, got %d", todayStat.AvgLatencyMs)
	}

	// 3. Test Get30DayUptimeMap, GetById, and GetAllByOrg
	uptimeMap, err := monitorRepo.Get30DayUptimeMap([]string{mon.ID.String()})
	if err != nil {
		t.Fatalf("Get30DayUptimeMap failed on SQLite: %v", err)
	}
	pct, ok := uptimeMap[mon.ID.String()]
	if !ok || pct <= 0 {
		t.Fatalf("Expected valid uptime percentage in Get30DayUptimeMap, got %v (ok=%v)", pct, ok)
	}

	fetchedMon, err := monitorRepo.GetById(org.ID.String(), mon.ID.String())
	if err != nil || fetchedMon == nil {
		t.Fatalf("GetById failed on SQLite: %v", err)
	}
	allMons, err := monitorRepo.GetAllByOrg(org.ID.String())
	if err != nil || len(allMons) != 1 {
		t.Fatalf("GetAllByOrg failed on SQLite: err=%v, len=%d", err, len(allMons))
	}

	// 4. Test RetentionJob.CleanOldChecks (clause.OnConflict upsert without LEAST/GREATEST/::numeric/NOW() on SQLite)
	retentionJob := worker.NewRetentionJob(db)
	retentionJob.CleanOldChecks()

	// Verify old check was deleted and merged into MonitorDailyStat
	var remainingOldChecks int64
	db.Model(&model.Check{}).Where("checked_at < ?", now.Add(-24*time.Hour)).Count(&remainingOldChecks)
	if remainingOldChecks != 0 {
		t.Errorf("Expected 0 old checks remaining after CleanOldChecks, got %d", remainingOldChecks)
	}

	var updatedStat model.MonitorDailyStat
	if err := db.Where("monitor_id = ? AND date = ?", mon.ID, twoDaysAgo).First(&updatedStat).Error; err != nil {
		t.Fatalf("Failed to fetch updated MonitorDailyStat: %v", err)
	}
	if updatedStat.TotalChecks != 11 {
		t.Errorf("Expected merged TotalChecks=11 after upsert, got %d", updatedStat.TotalChecks)
	}
	if updatedStat.UpChecks != 10 {
		t.Errorf("Expected merged UpChecks=10 after upsert, got %d", updatedStat.UpChecks)
	}
	if updatedStat.MinLatencyMs != 50 {
		t.Errorf("Expected merged MinLatencyMs=50 after upsert, got %d", updatedStat.MinLatencyMs)
	}
	if updatedStat.MaxLatencyMs != 150 {
		t.Errorf("Expected merged MaxLatencyMs=150 after upsert, got %d", updatedStat.MaxLatencyMs)
	}
}

func TestMetricAndSLAAccuracy_NoBiasFromTimeoutsOrDownChecks(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "netra_sla_accuracy.db")
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
		&model.Organization{},
		&model.Monitor{},
		&model.Check{},
		&model.MonitorDailyStat{},
		&model.MonitorPIC{},
		&model.MaintenanceMonitor{},
	)
	if err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	org := &model.Organization{Name: "SLA Org", Slug: "sla-org"}
	db.Create(org)

	mon := &model.Monitor{
		OrganizationID:     org.ID,
		Name:               "Payment API",
		Type:               "http",
		Target:             "https://pay.example.com/health",
		TargetDomain:       "pay.example.com",
		IntervalSeconds:    60,
		TimeoutSeconds:     10,
		ExpectedStatusCode: 200,
		IsActive:           true,
		Status:             "up",
		Criticality:        "extreme",
	}
	db.Create(mon)

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.UTC)
	checks := []model.Check{
		// 1. Healthy check: 100ms
		{
			MonitorID:      mon.ID,
			Status:         "up",
			StatusCode:     200,
			ResponseTimeMs: 100,
			CheckedAt:      today.Add(10 * time.Minute),
		},
		// 2. High-latency degraded check (200 OK, 500ms): Available (Up) + valid latency
		{
			MonitorID:      mon.ID,
			Status:         "degraded",
			StatusCode:     200,
			ResponseTimeMs: 500,
			ErrorMessage:   "Response time 500ms exceeded 300",
			CheckedAt:      today.Add(20 * time.Minute),
		},
		// 3. Timeout degraded check (status_code=0, 10000ms): Unavailable (Down) + excluded from latency
		{
			MonitorID:      mon.ID,
			Status:         "degraded",
			StatusCode:     0,
			ResponseTimeMs: 10000,
			ErrorMessage:   "Request timed out after 10s (Service did not respond in time)",
			CheckedAt:      today.Add(30 * time.Minute),
		},
		// 4. Connection refused down check (status_code=0, 1ms): Unavailable (Down) + excluded from latency
		{
			MonitorID:      mon.ID,
			Status:         "down",
			StatusCode:     0,
			ResponseTimeMs: 1,
			ErrorMessage:   "Connection refused: Host is offline or port is closed",
			CheckedAt:      today.Add(40 * time.Minute),
		},
		// 5. Maintenance check (status_code=503, 50ms): Excluded from SLA total/up/down and latency
		{
			MonitorID:      mon.ID,
			Status:         "maintenance",
			StatusCode:     503,
			ResponseTimeMs: 50,
			ErrorMessage:   "[Maintenance: Planned Window] Status code mismatch",
			CheckedAt:      today.Add(50 * time.Minute),
		},
	}
	for i := range checks {
		if err := db.Create(&checks[i]).Error; err != nil {
			t.Fatalf("failed to seed check %d: %v", i, err)
		}
	}

	monitorRepo := monitor.NewMonitorRepository(db, cfg)

	// Verify GetDailyStats
	stats, err := monitorRepo.GetDailyStats(mon.ID.String(), 7)
	if err != nil {
		t.Fatalf("GetDailyStats failed: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 daily stat row, got %d", len(stats))
	}
	s := stats[0]
	// Total active checks = 4 (maintenance excluded). Up = 2 (100ms up + 500ms degraded 200 OK). Down = 2 (timeout + down).
	if s.TotalChecks != 4 {
		t.Errorf("expected TotalChecks=4 (maintenance excluded), got %d", s.TotalChecks)
	}
	if s.UpChecks != 2 {
		t.Errorf("expected UpChecks=2 (including degraded 200 OK), got %d", s.UpChecks)
	}
	if s.DownChecks != 2 {
		t.Errorf("expected DownChecks=2 (including timeout), got %d", s.DownChecks)
	}
	if s.UptimePercentage != 50.0 {
		t.Errorf("expected UptimePercentage=50.0, got %f", s.UptimePercentage)
	}
	// Latency must only include 100ms and 500ms -> Avg=300ms, Min=100ms, Max=500ms (NOT 10000ms timeout or 1ms down!)
	if s.AvgLatencyMs != 300 {
		t.Errorf("expected AvgLatencyMs=300 (excluding 10000ms timeout and 1ms down), got %d", s.AvgLatencyMs)
	}
	if s.MinLatencyMs != 100 {
		t.Errorf("expected MinLatencyMs=100, got %d", s.MinLatencyMs)
	}
	if s.MaxLatencyMs != 500 {
		t.Errorf("expected MaxLatencyMs=500 (excluding 10000ms timeout), got %d", s.MaxLatencyMs)
	}

	// Verify Get30DayUptimeMap
	uptimeMap, err := monitorRepo.Get30DayUptimeMap([]string{mon.ID.String()})
	if err != nil {
		t.Fatalf("Get30DayUptimeMap failed: %v", err)
	}
	if uptimeMap[mon.ID.String()] != 50.0 {
		t.Errorf("expected 30d uptime 50.0%%, got %f%%", uptimeMap[mon.ID.String()])
	}

	// Verify GetFleetOverview
	overview, err := monitorRepo.GetFleetOverview(org.ID.String(), 30)
	if err != nil {
		t.Fatalf("GetFleetOverview failed: %v", err)
	}
	if overview.Uptime.CurrentPercent == nil || *overview.Uptime.CurrentPercent != 50.0 {
		t.Errorf("expected fleet CurrentPercent=50.0, got %v", overview.Uptime.CurrentPercent)
	}
	if overview.Latency.AvgMs == nil || *overview.Latency.AvgMs != 300 {
		t.Errorf("expected fleet Latency.AvgMs=300, got %v", overview.Latency.AvgMs)
	}
	if overview.Latency.FastestMs == nil || *overview.Latency.FastestMs != 100 {
		t.Errorf("expected fleet Latency.FastestMs=100, got %v", overview.Latency.FastestMs)
	}
}
