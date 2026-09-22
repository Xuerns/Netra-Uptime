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

	// Branding
	LogoURL    string `gorm:"default:''"`
	FaviconURL string `gorm:"default:''"`
	ThemeColor string `gorm:"default:'#4edea3'"`
	CustomCSS  string `gorm:"type:text;default:''"`

	Organization       Organization        `gorm:"foreignKey:OrganizationID"`
	StatusPageMonitors []StatusPageMonitor `gorm:"foreignKey:StatusPageID;constraint:OnDelete:CASCADE"`
	ComponentGroups    []ComponentGroup    `gorm:"foreignKey:StatusPageID;constraint:OnDelete:CASCADE"`
}

type StatusPageMonitor struct {
	ID           uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	StatusPageID uuid.UUID  `gorm:"type:uuid;not null;index"`
	MonitorID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	DisplayOrder int        `gorm:"default:0"`
	GroupID      *uuid.UUID `gorm:"type:uuid;default:null"`

	StatusPage     StatusPage      `gorm:"foreignKey:StatusPageID"`
	Monitor        Monitor         `gorm:"foreignKey:MonitorID"`
	ComponentGroup *ComponentGroup `gorm:"foreignKey:GroupID;constraint:OnDelete:SET NULL"`
}
