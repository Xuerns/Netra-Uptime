package model

import (
	"time"

	"github.com/google/uuid"
)

type NotificationLog struct {
	ID          uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	MonitorID   *uuid.UUID `gorm:"type:uuid;index"` // nullable: nil for org-level channel dispatches
	IncidentID  *uuid.UUID `gorm:"type:uuid;index"` // nullable: nil for SSL/domain alerts
	ChannelID   *uuid.UUID `gorm:"type:uuid;index"` // nullable: nil for assignee email dispatch
	Channel     string     // "email_assignee" | "email_pic" | "email_channel" | "slack" | "webhook"
	Message     string     `gorm:"type:text"` // human-readable summary of what was sent
	Status      string     // "sent" | "failed"
	ErrorDetail string     `gorm:"type:text"` // error message when Status == "failed"
	SentAt      time.Time  `gorm:"autoCreateTime"`

	Monitor             *Monitor             `gorm:"foreignKey:MonitorID;constraint:OnDelete:CASCADE;"`
	Incident            *Incident            `gorm:"foreignKey:IncidentID;constraint:OnDelete:CASCADE;"`
	NotificationChannel *NotificationChannel `gorm:"foreignKey:ChannelID;constraint:OnDelete:CASCADE;"`
}
