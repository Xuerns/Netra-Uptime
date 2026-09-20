package model

import (
	"time"

	"github.com/google/uuid"
)

type MonitorDailyStat struct {
	ID               uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	MonitorID        uuid.UUID `gorm:"type:uuid;not null;index:idx_monitor_daily_date,unique"`
	Date             time.Time `gorm:"type:date;not null;index:idx_monitor_daily_date,unique"`
	TotalChecks      int       `gorm:"not null;default:0"`
	UpChecks         int       `gorm:"not null;default:0"`
	DownChecks       int       `gorm:"not null;default:0"`
	UptimePercentage float64   `gorm:"type:numeric(5,2);not null;default:100.00"`
	AvgLatencyMs     int       `gorm:"not null;default:0"`
	MinLatencyMs     int       `gorm:"not null;default:0"`
	MaxLatencyMs     int       `gorm:"not null;default:0"`
	Count2xx         int       `gorm:"not null;default:0"`
	Count3xx         int       `gorm:"not null;default:0"`
	Count4xx         int       `gorm:"not null;default:0"`
	Count5xx         int       `gorm:"not null;default:0"`
	CountError       int       `gorm:"not null;default:0"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`

	Monitor Monitor `gorm:"foreignKey:MonitorID"`
}
