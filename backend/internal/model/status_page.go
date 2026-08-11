package model

import (
	"time"

	"github.com/google/uuid"
)

type StatusPage struct {
	ID             uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null"`
	Subdomain      string    `gorm:"unique;not null"`
	Title          string    `gorm:"not null"`
	Description    *string   `gorm:"type:text"`
	IsPublic       bool      `gorm:"default:true"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`

	Organization       Organization        `gorm:"foreignKey:OrganizationID"`
	StatusPageMonitors []StatusPageMonitor `gorm:"foreignKey:StatusPageID;constraint:OnDelete:CASCADE"`
}

type StatusPageMonitor struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	StatusPageID uuid.UUID `gorm:"type:uuid;not null"`
	MonitorID    uuid.UUID `gorm:"type:uuid;not null"`
	DisplayOrder int       `gorm:"default:0"`

	StatusPage StatusPage `gorm:"foreignKey:StatusPageID"`
	Monitor    Monitor    `gorm:"foreignKey:MonitorID"`
}
