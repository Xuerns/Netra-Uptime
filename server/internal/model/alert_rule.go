package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AlertRule struct {
	ID             uuid.UUID  `gorm:"primaryKey;type:uuid" json:"id"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null;index" json:"organization_id"`
	MonitorID      *uuid.UUID `gorm:"type:uuid;index" json:"monitor_id"`  // If nil, could belong to a template or apply org-wide
	TemplateID     *uuid.UUID `gorm:"type:uuid;index" json:"template_id"` // Optional link to an AlertRuleTemplate
	Name           string     `gorm:"not null" json:"name"`               // e.g. "5xx Server Error", "High Latency Warning"
	Metric         string     `gorm:"not null" json:"metric"`             // "status_code" | "response_time_ms" | "timeout"
	Operator       string     `gorm:"not null" json:"operator"`           // "!=" | "==" | ">" | ">=" | "<" | "<=" | "in" | "not_in"
	TargetValue    string     `gorm:"not null" json:"target_value"`       // "200", "2000", "500,502,503", "5xx", "4xx"
	Consecutive    int        `gorm:"not null;default:2" json:"consecutive"`
	Severity       string     `gorm:"not null;default:'critical'" json:"severity"` // "critical" | "warning"
	IsActive       bool       `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	Organization Organization       `gorm:"foreignKey:OrganizationID" json:"-"`
	Monitor      *Monitor           `gorm:"foreignKey:MonitorID;constraint:OnDelete:CASCADE;" json:"monitor,omitempty"`
	Template     *AlertRuleTemplate `gorm:"foreignKey:TemplateID;constraint:OnDelete:CASCADE;" json:"-"`
}

func (m *AlertRule) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
