package model

import (
	"time"

	"github.com/google/uuid"
)

type NotificationChannel struct {
	ID             uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null"`
	Name           string    `gorm:"not null;default:''"`
	Type           string    `gorm:"not null"` // email | slack | webhook
	Config         string    // stored as JSON string
	IsActive       bool      `gorm:"default:true"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`

	Organization     Organization      `gorm:"foreignKey:OrganizationID"`
	NotificationLogs []NotificationLog `gorm:"foreignKey:ChannelID;constraint:OnDelete:CASCADE;"`
}
