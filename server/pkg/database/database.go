package database

import (
	"fmt"
	"log"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"netra-uptime/config"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	logLevel := logger.Silent
	if cfg.AppEnv == "development" {
		logLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	switch cfg.DBDriver {
	case "sqlite":
		log.Printf("[Database] Initializing SQLite engine (%s) in WAL mode...", cfg.DBPath)
		dsn := fmt.Sprintf(
			"%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)&_pragma=cache_size(-64000)&_pragma=temp_store(MEMORY)",
			cfg.DBPath,
		)
		db, err := gorm.Open(sqlite.Open(dsn), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to open sqlite database: %w", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			return nil, err
		}

		// Serialize connections to prevent SQLITE_BUSY under concurrent worker writes
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
		sqlDB.SetConnMaxLifetime(time.Hour)
		return db, nil

	case "postgres":
		fallthrough
	default:
		log.Printf("[Database] Initializing PostgreSQL engine (%s:%s/%s)...", cfg.DBHost, cfg.DBPort, cfg.DBName)
		db, err := gorm.Open(postgres.Open(cfg.PostgresDSN()), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to open postgresql database: %w", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			return nil, err
		}

		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetConnMaxLifetime(time.Hour)
		return db, nil
	}
}
