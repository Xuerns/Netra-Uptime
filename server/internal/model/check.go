package model

import (
	"time"

	"github.com/google/uuid"
)

type Check struct {
	ID             uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	MonitorID      uuid.UUID `gorm:"type:uuid;not null"`
	Status         string    `gorm:"not null"` // up | down
	StatusCode     int
	ResponseTimeMs int
	ErrorMessage   string
	CheckedAt      time.Time `gorm:"autoCreateTime"`

	Monitor Monitor `gorm:"foreignKey:MonitorID"`
}
