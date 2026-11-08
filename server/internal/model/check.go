package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Check struct {
	ID              uuid.UUID `gorm:"primaryKey;type:uuid"`
	MonitorID       uuid.UUID `gorm:"type:uuid;not null"`
	Status          string    `gorm:"not null"` // up | down | degraded | maintenance
	StatusCode      int
	ResponseTimeMs  int
	ErrorMessage    string
	ResponseSnippet string    `gorm:"type:text;default:''" json:"response_snippet,omitempty"`
	CheckedAt       time.Time `gorm:"autoCreateTime"`

	Monitor Monitor `gorm:"foreignKey:MonitorID"`
}

func (m *Check) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
