package database

import (
	"netra-uptime/config"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgres(c *config.Config) (*gorm.DB, error) {
	logLevel := logger.Silent
	if c.AppEnv == "development" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(c.PostgresDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})

	if err != nil {
		return nil, err
	}

	postDB, err := db.DB()

	if err != nil {
		return nil, err
	}

	postDB.SetConnMaxIdleTime(10)
	postDB.SetMaxIdleConns(100)
	postDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
