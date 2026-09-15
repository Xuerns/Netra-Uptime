package model

import (
	"time"

	"github.com/google/uuid"
)

type Incident struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	MonitorID  uuid.UUID `gorm:"type:uuid;not null"`
	Status     string    `gorm:"not null"` // ongoing | resolved
	Severity   string    `gorm:"not null;default:'medium'"`
	StartedAt  time.Time
	ResolvedAt *time.Time 
	Summary       string
	TriggeredRule string     `gorm:"default:''"`
	AlertRuleID   *uuid.UUID `gorm:"type:uuid;default:null"`

	Monitor          Monitor            `gorm:"foreignKey:MonitorID"`
	IncidentUpdates  []IncidentUpdate   `gorm:"foreignKey:IncidentID;constraint:OnDelete:CASCADE;"`
	NotificationLogs []NotificationLog  `gorm:"foreignKey:IncidentID;constraint:OnDelete:CASCADE;"`
	Assignees        []IncidentAssignee `gorm:"foreignKey:IncidentID;constraint:OnDelete:CASCADE;"`
}

type IncidentUpdate struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	IncidentID uuid.UUID `gorm:"type:uuid;not null"`
	CreatedBy  uuid.UUID `gorm:"type:uuid;not null"` 
	Message    string
	CreatedAt  time.Time `gorm:"autoCreateTime"`

	Incident Incident `gorm:"foreignKey:IncidentID"`
	User     User     `gorm:"foreignKey:CreatedBy"`
}

type IncidentAssignee struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	IncidentID uuid.UUID `gorm:"type:uuid;not null"`
	UserID     uuid.UUID `gorm:"type:uuid;not null"`
	Role       string    `gorm:"not null;default:'responder'"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`

	Incident Incident `gorm:"foreignKey:IncidentID"`
	User     User     `gorm:"foreignKey:UserID"`
}
