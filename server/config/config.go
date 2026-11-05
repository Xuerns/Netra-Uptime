package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"

	"netra-uptime/pkg/utils"
)

var (
	cfgInstance *Config
	cfgOnce     sync.Once
)

type Config struct {
	AppEnv string
	Host   string
	Port   string

	RecentChecksLimit     int
	SelfHostedMaxMonitors int
	SelfHostedMaxOrgs     int
	SelfHostedMaxMembers  int
	SelfHostedMaxPIC      int
	SelfHostedMinInterval int
	SelfHostedMaxTimeout  int

	MinMonitorIntervalSeconds     int
	DefaultMonitorIntervalSeconds int

	MaxMonitorsPerOrg      int
	MaxStatusPagesPerOrg   int
	MaxOrgsPerUser         int
	MaxMembersPerOrg       int
	MaxAssigneesPerMonitor int
	AllowRegistration      bool
	AuthRateLimit          int

	AllowPrivateNetworkMonitoring       bool
	AllowPrivateNetworkMonitoringReason string

	DBDriver   string
	DBPath     string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JwtSecret string
}

func parseIntEnv(key string, defaultVal int) int {
	if valStr := os.Getenv(key); valStr != "" {
		if val, err := strconv.Atoi(strings.TrimSpace(valStr)); err == nil {
			return val
		}
	}
	return defaultVal
}

func parseBoolEnv(key string, defaultVal bool) bool {
	if valStr := os.Getenv(key); valStr != "" {
		if val, err := strconv.ParseBool(strings.TrimSpace(valStr)); err == nil {
			return val
		}
	}
	return defaultVal
}

func isLoopbackHost(host string) bool {
	h := strings.Trim(strings.ToLower(strings.TrimSpace(host)), "[]")
	if h == "localhost" || h == "127.0.0.1" || h == "::1" {
		return true
	}
	if ip := net.ParseIP(h); ip != nil {
		return ip.IsLoopback()
	}
	return false
}

// GetExecutableDir returns the directory containing the currently running executable.
// If os.Executable fails, it falls back to ".".
func GetExecutableDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "."
	}
	if realPath, err := filepath.EvalSymlinks(exePath); err == nil {
		exePath = realPath
	}
	return filepath.Dir(exePath)
}

func Load() *Config {
	cfgOnce.Do(func() {
		cfgInstance = loadConfig()
	})
	return cfgInstance
}

var knownPlaceholderSecrets = map[string]bool{
	"change_me_to_a_random_secret_of_at_least_32_characters":  true,
	"netra-uptime-dev-secret-must-be-at-least-32-chars-long!": true,
}

func loadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			if err := godotenv.Load(filepath.Join(GetExecutableDir(), ".env")); err != nil {
				if os.Getenv("DB_HOST") == "" && os.Getenv("APP_ENV") == "" {
					log.Println("No .env file found")
				}
			}
		}
	}

	dbDriver := strings.ToLower(strings.TrimSpace(os.Getenv("DB_DRIVER")))
	if dbDriver == "" {
		if os.Getenv("DB_HOST") != "" {
			dbDriver = "postgres"
		} else {
			dbDriver = "sqlite"
		}
	}

	host := strings.TrimSpace(os.Getenv("HOST"))
	if host == "" {
		host = strings.TrimSpace(os.Getenv("BIND_HOST"))
	}
	if host == "" {
		if dbDriver == "sqlite" {
			host = "127.0.0.1"
		} else {
			host = "0.0.0.0"
		}
	}

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "27346"
	}

	dbPath := strings.TrimSpace(os.Getenv("DB_PATH"))
	if dbPath == "" {
		dbPath = "netra.db"
	}
	if dbDriver == "sqlite" && !filepath.IsAbs(dbPath) {
		dbPath = filepath.Join(GetExecutableDir(), dbPath)
	}

	dbSSLMode := os.Getenv("DB_SSLMODE")
	if dbSSLMode == "" {
		dbSSLMode = "disable"
	}

	minInterval := parseIntEnv("SELFHOSTED_MIN_INTERVAL", 30)
	if minInterval < 30 {
		minInterval = 30
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if dbDriver == "sqlite" && jwtSecret == "" {
		secretFile := filepath.Clean(filepath.Join(GetExecutableDir(), ".netra_secret"))
		// #nosec G304 -- .netra_secret is an internal secret file path in the executable directory
		if data, err := os.ReadFile(secretFile); err == nil {
			if trimmed := strings.TrimSpace(string(data)); len(trimmed) >= 32 {
				jwtSecret = trimmed
				_ = os.Setenv("JWT_SECRET", jwtSecret)
			}
		}
		if jwtSecret == "" {
			buf := make([]byte, 32)
			if _, err := rand.Read(buf); err != nil {
				log.Fatalf("FATAL: Failed to generate random JWT_SECRET: %v", err)
			}
			secret := hex.EncodeToString(buf)
			// #nosec G304 -- .netra_secret is an internal secret file path in the executable directory
			if err := os.WriteFile(secretFile, []byte(secret), 0600); err != nil {
				log.Printf("WARNING: Failed to persist .netra_secret at %s: %v", secretFile, err)
			} else {
				log.Printf("[Config] Generated persistent JWT_SECRET in %s for standalone SQLite mode", secretFile)
			}
			jwtSecret = secret
			_ = os.Setenv("JWT_SECRET", secret)
		}
	}

	var allowPrivate bool
	var allowPrivateReason string
	if envVal := strings.TrimSpace(os.Getenv("ALLOW_PRIVATE_NETWORK_MONITORING")); envVal != "" {
		allowPrivate = parseBoolEnv("ALLOW_PRIVATE_NETWORK_MONITORING", false)
		allowPrivateReason = "env override"
	} else {
		if dbDriver == "sqlite" && isLoopbackHost(host) {
			allowPrivate = true
			allowPrivateReason = "desktop loopback default"
		} else {
			allowPrivate = false
			allowPrivateReason = "production default"
		}
	}
	utils.SetDefaultAllowPrivateNetwork(allowPrivate)

	cfg := &Config{
		AppEnv: os.Getenv("APP_ENV"),
		Host:   host,
		Port:   port,

		RecentChecksLimit:     parseIntEnv("RECENT_CHECKS_LIMIT", 60),
		SelfHostedMaxMonitors: parseIntEnv("SELFHOSTED_MAX_MONITORS", 0),
		SelfHostedMaxOrgs:     parseIntEnv("SELFHOSTED_MAX_ORGS", 0),
		SelfHostedMaxMembers:  parseIntEnv("SELFHOSTED_MAX_MEMBERS", 0),
		SelfHostedMaxPIC:      parseIntEnv("SELFHOSTED_MAX_PIC", 10),
		SelfHostedMinInterval: minInterval,
		SelfHostedMaxTimeout:  parseIntEnv("SELFHOSTED_MAX_TIMEOUT", 30),

		MinMonitorIntervalSeconds:     parseIntEnv("MIN_MONITOR_INTERVAL_SECONDS", 30),
		DefaultMonitorIntervalSeconds: parseIntEnv("DEFAULT_MONITOR_INTERVAL_SECONDS", 60),

		MaxMonitorsPerOrg:      parseIntEnv("MAX_MONITORS_PER_ORG", 200),
		MaxStatusPagesPerOrg:   parseIntEnv("MAX_STATUS_PAGES_PER_ORG", 10),
		MaxOrgsPerUser:         parseIntEnv("MAX_ORGS_PER_USER", 10),
		MaxMembersPerOrg:       parseIntEnv("MAX_MEMBERS_PER_ORG", 100),
		MaxAssigneesPerMonitor: parseIntEnv("MAX_ASSIGNEES_PER_MONITOR", 50),
		AllowRegistration:      parseBoolEnv("ALLOW_REGISTRATION", true),
		AuthRateLimit:          parseIntEnv("AUTH_RATE_LIMIT", 60),

		AllowPrivateNetworkMonitoring:       allowPrivate,
		AllowPrivateNetworkMonitoringReason: allowPrivateReason,

		DBDriver:   dbDriver,
		DBPath:     dbPath,
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBSSLMode:  dbSSLMode,

		JwtSecret: jwtSecret,
	}

	const AbsoluteMinimumInterval = 30
	if cfg.AuthRateLimit <= 0 {
		cfg.AuthRateLimit = 60
	}
	if cfg.MinMonitorIntervalSeconds < AbsoluteMinimumInterval {
		log.Printf("WARNING: MIN_MONITOR_INTERVAL_SECONDS (%d) below safety floor (%d). Enforcing %d seconds.",
			cfg.MinMonitorIntervalSeconds, AbsoluteMinimumInterval, AbsoluteMinimumInterval)
		cfg.MinMonitorIntervalSeconds = AbsoluteMinimumInterval
	}
	if cfg.DefaultMonitorIntervalSeconds < cfg.MinMonitorIntervalSeconds {
		log.Printf("NOTICE: DEFAULT_MONITOR_INTERVAL_SECONDS (%d) is lower than min interval (%d). Adjusting default to %d.",
			cfg.DefaultMonitorIntervalSeconds, cfg.MinMonitorIntervalSeconds, cfg.MinMonitorIntervalSeconds)
		cfg.DefaultMonitorIntervalSeconds = cfg.MinMonitorIntervalSeconds
	}

	if strings.EqualFold(cfg.AppEnv, "production") {
		if knownPlaceholderSecrets[cfg.JwtSecret] {
			log.Fatal("FATAL: Insecure placeholder JWT_SECRET detected. Please provide a secure random secret.")
		}
		if len(cfg.JwtSecret) < 32 {
			log.Fatal("FATAL SECURITY CONFIGURATION: JWT_SECRET must be at least 32 characters in production")
		}
		if cfg.DBDriver == "postgres" && cfg.DBSSLMode == "disable" {
			log.Println("SECURITY WARNING: DB_SSLMODE is set to 'disable' in production. Configure 'require' or 'verify-full' to prevent unencrypted database traffic.")
		}
	} else if len(cfg.JwtSecret) == 0 {
		log.Println("WARNING: JWT_SECRET is not set in development. Using development fallback.")
		cfg.JwtSecret = "netra-uptime-dev-secret-must-be-at-least-32-chars-long!"
	}

	return cfg
}

func (c *Config) PostgresDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta", c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode)
}
