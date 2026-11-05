package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditLog struct {
	ID             uuid.UUID  `gorm:"primaryKey;type:uuid" json:"id"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;index;not null" json:"organization_id"`
	UserID         uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	Action         string     `gorm:"type:varchar(50);not null" json:"action"`
	EntityType     string     `gorm:"type:varchar(50);not null" json:"entity_type"`
	EntityID       *uuid.UUID `gorm:"type:uuid" json:"entity_id,omitempty"`
	Description    string     `gorm:"type:text" json:"description"`
	Details        string     `gorm:"type:text" json:"details,omitempty"`
	IPAddress      string     `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`

	Organization Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE" json:"organization,omitempty"`
	User         User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (m *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
