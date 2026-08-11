package model

import (
	"time"

	"github.com/google/uuid"
)

type NotificationLog struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	IncidentID uuid.UUID `gorm:"type:uuid;not null"`
	ChannelID  uuid.UUID `gorm:"type:uuid;not null"`
	Status     string    // sent | failed
	SentAt     time.Time `gorm:"autoCreateTime"`

	Incident            Incident            `gorm:"foreignKey:IncidentID"`
	NotificationChannel NotificationChannel `gorm:"foreignKey:ChannelID"`
}
