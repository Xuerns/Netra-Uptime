package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AlertRuleTemplate struct {
	ID             uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`
	Name           string    `gorm:"not null" json:"name"`
	Description    string    `json:"description"`
	IsDefault      bool      `gorm:"default:false" json:"is_default"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Organization  Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"-"`
	Rules         []AlertRule  `gorm:"foreignKey:TemplateID;constraint:OnDelete:CASCADE;" json:"rules"`
	MonitorsCount int64        `gorm:"-" json:"monitors_count,omitempty"`
}

func (m *AlertRuleTemplate) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
