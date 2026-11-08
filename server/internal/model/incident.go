package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Incident struct {
	ID              uuid.UUID `gorm:"primaryKey;type:uuid"`
	MonitorID       uuid.UUID `gorm:"type:uuid;not null"`
	Status          string    `gorm:"not null"`                  // ongoing | resolved
	Severity        string    `gorm:"not null;default:'medium'"` // critical | high | medium | low
	StartedAt       time.Time
	ResolvedAt      *time.Time
	Summary         string
	TriggeredRule   string     `gorm:"default:''"`
	ResponseSnippet string     `gorm:"type:text;default:''" json:"response_snippet,omitempty"`
	AlertRuleID     *uuid.UUID `gorm:"type:uuid;default:null"`

	Monitor          Monitor            `gorm:"foreignKey:MonitorID"`
	IncidentUpdates  []IncidentUpdate   `gorm:"foreignKey:IncidentID;constraint:OnDelete:CASCADE;"`
	NotificationLogs []NotificationLog  `gorm:"foreignKey:IncidentID;constraint:OnDelete:CASCADE;"`
	Assignees        []IncidentAssignee `gorm:"foreignKey:IncidentID;constraint:OnDelete:CASCADE;"`
}

func (m *Incident) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

type IncidentUpdate struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid"`
	IncidentID uuid.UUID `gorm:"type:uuid;not null"`
	CreatedBy  uuid.UUID `gorm:"type:uuid;not null"`
	Message    string
	CreatedAt  time.Time `gorm:"autoCreateTime"`

	Incident Incident `gorm:"foreignKey:IncidentID"`
	User     User     `gorm:"foreignKey:CreatedBy"`
}

func (m *IncidentUpdate) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

type IncidentAssignee struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid"`
	IncidentID uuid.UUID `gorm:"type:uuid;not null"`
	UserID     uuid.UUID `gorm:"type:uuid;not null"`
	Role       string    `gorm:"not null;default:'responder'"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`

	Incident Incident `gorm:"foreignKey:IncidentID"`
	User     User     `gorm:"foreignKey:UserID"`
}

func (m *IncidentAssignee) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
