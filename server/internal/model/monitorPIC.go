package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MonitorPIC struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid"`
	MonitorID uuid.UUID `gorm:"not null;type:uuid"`
	UserID    uuid.UUID `gorm:"not null;type:uuid"`
	CreateAt  time.Time `gorm:"autoCreateTime"`

	Monitor Monitor `gorm:"foreignKey:MonitorID"`
	User    User    `gorm:"foreignKey:UserID"`
}

func (m *MonitorPIC) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
