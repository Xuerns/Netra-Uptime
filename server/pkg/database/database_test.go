package database

import (
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"netra-uptime/config"
	"netra-uptime/internal/model"
)

func TestNewDatabase_SQLiteAutoMigrateAndInsert(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	cfg := &config.Config{
		DBDriver: "sqlite",
		DBPath:   dbPath,
	}

	db, err := NewDatabase(cfg)
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

	user := &model.User{
		Email: "test@example.com",
		Name:  "Test User",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to insert User: %v", err)
	}
	if user.ID == uuid.Nil {
		t.Fatalf("Expected BeforeCreate to generate non-nil UUID for User, got %v", user.ID)
	}

	org := &model.Organization{
		Name: "Test Org",
		Slug: "test-org",
	}
	if err := db.Create(org).Error; err != nil {
		t.Fatalf("Failed to insert Organization: %v", err)
	}
	if org.ID == uuid.Nil {
		t.Fatalf("Expected BeforeCreate to generate non-nil UUID for Organization, got %v", org.ID)
	}
}
