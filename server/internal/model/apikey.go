package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type APIKey struct {
	ID             uuid.UUID  `gorm:"primaryKey;type:uuid"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null;index"`
	CreatedByID    uuid.UUID  `gorm:"type:uuid;not null"`
	Name           string     `gorm:"not null"`
	KeyHash        string     `gorm:"not null;uniqueIndex"`
	KeyPrefix      string     `gorm:"not null"`
	Permissions    string     `gorm:"not null;default:'read'"` // read | write
	LastUsedAt     *time.Time `gorm:"default:null"`
	LastUsedIP     string     `gorm:"default:''"`
	ExpiresAt      *time.Time `gorm:"default:null"`
	CreatedAt      time.Time  `gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime"`

	Organization Organization `gorm:"foreignKey:OrganizationID"`
	CreatedBy    User         `gorm:"foreignKey:CreatedByID"`
}

func (m *APIKey) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
