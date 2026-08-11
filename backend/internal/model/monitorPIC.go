package model

import (
	"time"

	"github.com/google/uuid"
)

type MonitorPIC struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	MonitorID uuid.UUID `gorm:"not null;type:uuid"`
	UserID    uuid.UUID `gorm:"not null;type:uuid"`
	CreateAt  time.Time `gorm:"autoCreateTime"`

	Monitor Monitor `gorm:"foreignKey:MonitorID"`
	User    User    `gorm:"foreignKey:UserID"`
}
