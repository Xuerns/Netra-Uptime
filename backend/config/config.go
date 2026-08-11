package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv string
	Port   string

	AppMode               string
	RecentChecksLimit     int
	SelfHostedMaxMonitors int
	SelfHostedMaxOrgs     int
	SelfHostedMaxMembers  int
	SelfHostedMaxPIC      int
	SelfHostedMinInterval int
	SelfHostedMaxTimeout  int

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
		if val, err := strconv.Atoi(valStr); err == nil {
			return val
		}
	}
	return defaultVal
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	appMode := os.Getenv("APP_MODE")
	if appMode == "" {
		appMode = "selfhosted"
	}

	return &Config{
		AppEnv: os.Getenv("APP_ENV"),
		Port:   os.Getenv("PORT"),

		AppMode:               appMode,
		RecentChecksLimit:     parseIntEnv("RECENT_CHECKS_LIMIT", 60),
		SelfHostedMaxMonitors: parseIntEnv("SELFHOSTED_MAX_MONITORS", 0),
		SelfHostedMaxOrgs:     parseIntEnv("SELFHOSTED_MAX_ORGS", 0),
		SelfHostedMaxMembers:  parseIntEnv("SELFHOSTED_MAX_MEMBERS", 0),
		SelfHostedMaxPIC:      parseIntEnv("SELFHOSTED_MAX_PIC", 10),
		SelfHostedMinInterval: parseIntEnv("SELFHOSTED_MIN_INTERVAL", 10),
		SelfHostedMaxTimeout:  parseIntEnv("SELFHOSTED_MAX_TIMEOUT", 60),

		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBSSLMode:  os.Getenv("DB_SSLMODE"),

		RedisHost:     os.Getenv("REDIS_HOST"),
		RedisPort:     os.Getenv("REDIS_PORT"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),

		JwtSecret: os.Getenv("JWT_SECRET"),
	}
}

func (c *Config) PostgresDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta", c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode)
}
