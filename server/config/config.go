package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"sync"
)

var (
	cfgInstance *Config
	cfgOnce     sync.Once
)

type Config struct {
	AppEnv string
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

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	RedisHost     string
	RedisPort     string
	RedisPassword string

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
			if os.Getenv("DB_HOST") == "" && os.Getenv("APP_ENV") == "" {
				log.Println("No .env file found")
			}
		}
	}

	dbSSLMode := os.Getenv("DB_SSLMODE")
	if dbSSLMode == "" {
		dbSSLMode = "disable"
	}

	minInterval := parseIntEnv("SELFHOSTED_MIN_INTERVAL", 30)
	if minInterval < 30 {
		minInterval = 30
	}

	cfg := &Config{
		AppEnv: os.Getenv("APP_ENV"),
		Port:   os.Getenv("PORT"),

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

		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBSSLMode:  dbSSLMode,

		RedisHost:     os.Getenv("REDIS_HOST"),
		RedisPort:     os.Getenv("REDIS_PORT"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),

		JwtSecret: os.Getenv("JWT_SECRET"),
	}

	const AbsoluteMinimumInterval = 30
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
		if cfg.DBSSLMode == "disable" {
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
