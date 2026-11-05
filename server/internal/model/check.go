package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Check struct {
	ID             uuid.UUID `gorm:"primaryKey;type:uuid"`
	MonitorID      uuid.UUID `gorm:"type:uuid;not null"`
	Status         string    `gorm:"not null"` // up | down
	StatusCode     int
	ResponseTimeMs int
	ErrorMessage   string
	CheckedAt      time.Time `gorm:"autoCreateTime"`

	Monitor Monitor `gorm:"foreignKey:MonitorID"`
}

func (m *Check) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
