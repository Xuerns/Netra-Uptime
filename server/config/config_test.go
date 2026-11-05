package config

import (
	"net"
	"os"
	"path/filepath"
	"testing"

	"netra-uptime/pkg/utils"
)

func TestGetExecutableDir(t *testing.T) {
	dir := GetExecutableDir()
	if dir == "" {
		t.Fatal("expected non-empty executable directory")
	}
	if !filepath.IsAbs(dir) {
		t.Fatalf("expected absolute path for executable directory, got: %s", dir)
	}
}

func TestHostResolution(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-key-at-least-32-characters-long!")

	// 1. SQLite default -> 127.0.0.1
	t.Run("SQLite default host is 127.0.0.1", func(t *testing.T) {
		t.Setenv("DB_DRIVER", "sqlite")
		t.Setenv("DB_HOST", "")
		t.Setenv("HOST", "")
		t.Setenv("BIND_HOST", "")

		cfg := loadConfig()
		if cfg.Host != "127.0.0.1" {
			t.Fatalf("expected host 127.0.0.1, got %q", cfg.Host)
		}
	})

	// 2. Postgres default -> 0.0.0.0
	t.Run("Postgres default host is 0.0.0.0", func(t *testing.T) {
		t.Setenv("DB_DRIVER", "postgres")
		t.Setenv("DB_HOST", "localhost")
		t.Setenv("HOST", "")
		t.Setenv("BIND_HOST", "")

		cfg := loadConfig()
		if cfg.Host != "0.0.0.0" {
			t.Fatalf("expected host 0.0.0.0, got %q", cfg.Host)
		}
	})

	// 3. Explicit HOST env var
	t.Run("Explicit HOST overrides default", func(t *testing.T) {
		t.Setenv("DB_DRIVER", "sqlite")
		t.Setenv("HOST", "192.168.1.10")

		cfg := loadConfig()
		if cfg.Host != "192.168.1.10" {
			t.Fatalf("expected host 192.168.1.10, got %q", cfg.Host)
		}
	})

	// 4. BIND_HOST fallback
	t.Run("BIND_HOST fallback works", func(t *testing.T) {
		t.Setenv("DB_DRIVER", "sqlite")
		t.Setenv("HOST", "")
		t.Setenv("BIND_HOST", "0.0.0.0")

		cfg := loadConfig()
		if cfg.Host != "0.0.0.0" {
			t.Fatalf("expected host 0.0.0.0, got %q", cfg.Host)
		}
	})
}

func TestSQLitePathResolution(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-key-at-least-32-characters-long!")

	t.Run("Relative DB_PATH is resolved against executable dir", func(t *testing.T) {
		t.Setenv("DB_DRIVER", "sqlite")
		t.Setenv("DB_PATH", "custom_test.db")

		cfg := loadConfig()
		expected := filepath.Join(GetExecutableDir(), "custom_test.db")
		if cfg.DBPath != expected {
			t.Fatalf("expected %q, got %q", expected, cfg.DBPath)
		}
	})

	t.Run("Absolute DB_PATH is preserved", func(t *testing.T) {
		absPath := filepath.Join(os.TempDir(), "abs_test.db")
		t.Setenv("DB_DRIVER", "sqlite")
		t.Setenv("DB_PATH", absPath)

		cfg := loadConfig()
		if cfg.DBPath != absPath {
			t.Fatalf("expected %q, got %q", absPath, cfg.DBPath)
		}
	})
}

func TestAllowPrivateNetworkMonitoring(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-key-at-least-32-characters-long!")

	// 1. sqlite + loopback (127.0.0.1) -> true (desktop loopback default)
	t.Run("sqlite + 127.0.0.1 defaults to true", func(t *testing.T) {
		t.Setenv("DB_DRIVER", "sqlite")
		t.Setenv("DB_HOST", "")
		t.Setenv("HOST", "127.0.0.1")
		t.Setenv("ALLOW_PRIVATE_NETWORK_MONITORING", "")

		cfg := loadConfig()
		if !cfg.AllowPrivateNetworkMonitoring {
			t.Fatalf("expected AllowPrivateNetworkMonitoring=true, got %v", cfg.AllowPrivateNetworkMonitoring)
		}
		if cfg.AllowPrivateNetworkMonitoringReason != "desktop loopback default" {
			t.Fatalf("expected reason 'desktop loopback default', got %q", cfg.AllowPrivateNetworkMonitoringReason)
		}
		if os.Getenv("ALLOW_PRIVATE_NETWORK_MONITORING") != "" {
			t.Fatalf("expected ALLOW_PRIVATE_NETWORK_MONITORING env var to remain unset, got %q", os.Getenv("ALLOW_PRIVATE_NETWORK_MONITORING"))
		}
		if err := utils.IsRestrictedIP(net.ParseIP("192.168.1.1")); err != nil {
			t.Fatalf("expected private IP 192.168.1.1 to be allowed in desktop loopback mode, got err: %v", err)
		}
	})

	// 2. sqlite + loopback (localhost) -> true
	t.Run("sqlite + localhost defaults to true", func(t *testing.T) {
		t.Setenv("DB_DRIVER", "sqlite")
		t.Setenv("DB_HOST", "")
		t.Setenv("HOST", "localhost")
		t.Setenv("ALLOW_PRIVATE_NETWORK_MONITORING", "")

		cfg := loadConfig()
		if !cfg.AllowPrivateNetworkMonitoring {
			t.Fatalf("expected AllowPrivateNetworkMonitoring=true, got %v", cfg.AllowPrivateNetworkMonitoring)
		}
		if cfg.AllowPrivateNetworkMonitoringReason != "desktop loopback default" {
			t.Fatalf("expected reason 'desktop loopback default', got %q", cfg.AllowPrivateNetworkMonitoringReason)
		}
	})

	// 3. sqlite + loopback (::1) -> true
	t.Run("sqlite + ::1 defaults to true", func(t *testing.T) {
		t.Setenv("DB_DRIVER", "sqlite")
		t.Setenv("DB_HOST", "")
		t.Setenv("HOST", "::1")
		t.Setenv("ALLOW_PRIVATE_NETWORK_MONITORING", "")

		cfg := loadConfig()
		if !cfg.AllowPrivateNetworkMonitoring {
			t.Fatalf("expected AllowPrivateNetworkMonitoring=true, got %v", cfg.AllowPrivateNetworkMonitoring)
		}
		if cfg.AllowPrivateNetworkMonitoringReason != "desktop loopback default" {
			t.Fatalf("expected reason 'desktop loopback default', got %q", cfg.AllowPrivateNetworkMonitoringReason)
		}
	})

	// 4. sqlite + 0.0.0.0 -> false (production default, e.g. docker run)
	t.Run("sqlite + 0.0.0.0 defaults to false", func(t *testing.T) {
		t.Setenv("DB_DRIVER", "sqlite")
		t.Setenv("DB_HOST", "")
		t.Setenv("HOST", "0.0.0.0")
		t.Setenv("ALLOW_PRIVATE_NETWORK_MONITORING", "")

		cfg := loadConfig()
		if cfg.AllowPrivateNetworkMonitoring {
			t.Fatalf("expected AllowPrivateNetworkMonitoring=false, got %v", cfg.AllowPrivateNetworkMonitoring)
		}
		if cfg.AllowPrivateNetworkMonitoringReason != "production default" {
			t.Fatalf("expected reason 'production default', got %q", cfg.AllowPrivateNetworkMonitoringReason)
		}
	})

	// 5. postgres -> false (production default, regardless of host)
	t.Run("postgres defaults to false", func(t *testing.T) {
		t.Setenv("DB_DRIVER", "postgres")
		t.Setenv("DB_HOST", "localhost")
		t.Setenv("HOST", "127.0.0.1")
		t.Setenv("ALLOW_PRIVATE_NETWORK_MONITORING", "")

		cfg := loadConfig()
		if cfg.AllowPrivateNetworkMonitoring {
			t.Fatalf("expected AllowPrivateNetworkMonitoring=false, got %v", cfg.AllowPrivateNetworkMonitoring)
		}
		if cfg.AllowPrivateNetworkMonitoringReason != "production default" {
			t.Fatalf("expected reason 'production default', got %q", cfg.AllowPrivateNetworkMonitoringReason)
		}
	})

	// 6. Explicit env override to false when sqlite + loopback
	t.Run("explicit false override on sqlite loopback", func(t *testing.T) {
		t.Setenv("DB_DRIVER", "sqlite")
		t.Setenv("DB_HOST", "")
		t.Setenv("HOST", "127.0.0.1")
		t.Setenv("ALLOW_PRIVATE_NETWORK_MONITORING", "false")

		cfg := loadConfig()
		if cfg.AllowPrivateNetworkMonitoring {
			t.Fatalf("expected AllowPrivateNetworkMonitoring=false, got %v", cfg.AllowPrivateNetworkMonitoring)
		}
		if cfg.AllowPrivateNetworkMonitoringReason != "env override" {
			t.Fatalf("expected reason 'env override', got %q", cfg.AllowPrivateNetworkMonitoringReason)
		}
	})

	// 7. Explicit env override to true when sqlite + 0.0.0.0
	t.Run("explicit true override on sqlite 0.0.0.0", func(t *testing.T) {
		t.Setenv("DB_DRIVER", "sqlite")
		t.Setenv("DB_HOST", "")
		t.Setenv("HOST", "0.0.0.0")
		t.Setenv("ALLOW_PRIVATE_NETWORK_MONITORING", "true")

		cfg := loadConfig()
		if !cfg.AllowPrivateNetworkMonitoring {
			t.Fatalf("expected AllowPrivateNetworkMonitoring=true, got %v", cfg.AllowPrivateNetworkMonitoring)
		}
		if cfg.AllowPrivateNetworkMonitoringReason != "env override" {
			t.Fatalf("expected reason 'env override', got %q", cfg.AllowPrivateNetworkMonitoringReason)
		}
	})

	// 8. Explicit env override to true on postgres
	t.Run("explicit true override on postgres", func(t *testing.T) {
		t.Setenv("DB_DRIVER", "postgres")
		t.Setenv("DB_HOST", "postgres")
		t.Setenv("HOST", "0.0.0.0")
		t.Setenv("ALLOW_PRIVATE_NETWORK_MONITORING", "true")

		cfg := loadConfig()
		if !cfg.AllowPrivateNetworkMonitoring {
			t.Fatalf("expected AllowPrivateNetworkMonitoring=true, got %v", cfg.AllowPrivateNetworkMonitoring)
		}
		if cfg.AllowPrivateNetworkMonitoringReason != "env override" {
			t.Fatalf("expected reason 'env override', got %q", cfg.AllowPrivateNetworkMonitoringReason)
		}
	})
}
