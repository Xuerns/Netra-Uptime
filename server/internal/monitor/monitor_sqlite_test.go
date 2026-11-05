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

	// Seed recent checks (within 24h) and old checks (2 days ago, older than 24h retention threshold)
	checks := []model.Check{
		{
			MonitorID:      mon.ID,
			Status:         "up",
			StatusCode:     200,
			ResponseTimeMs: 40,
			CheckedAt:      now.Add(-2 * time.Hour),
		},
		{
			MonitorID:      mon.ID,
			Status:         "up",
			StatusCode:     200,
			ResponseTimeMs: 60,
			CheckedAt:      now.Add(-1 * time.Hour),
		},
		{
			MonitorID:      mon.ID,
			Status:         "up",
			StatusCode:     200,
			ResponseTimeMs: 120,
			CheckedAt:      now.Add(-30 * time.Minute),
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
